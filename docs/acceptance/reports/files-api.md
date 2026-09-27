# B06 文件 API 与真实节点集成验证

2026-09-09。实现位置：`internal/master/files.go`、`internal/filesystem/**`；公共 Daemon/任务/错误映射由主代理集成。测试：`internal/master/files_integration_test.go`。测试只使用临时文件及回环 TLS Master/Daemon，没有生产节点。

## 可调用入口

前缀：`/api/v1/instances/{id}/files`。实例与根目录配置由 Master/Daemon 从实际资源解析，不接受客户端宿主路径。每个请求经过会话/CSRF（修改请求）、资源权限检查；活跃请求参与授权撤销，分片再次检查权限。

| 方法/路径 | 合同 |
| --- | --- |
| GET 根路径 | `path,offset,limit,version,search,sort,order`；返回 `items,version,total,nextOffset`。排序/搜索在节点完整有界目录上执行后分页。sort 为 name/size/modified/kind；order 为 asc/desc。每行含 kind 与 isDir。 |
| GET `/stat?path=...` | 普通文件或递归目录版本、大小、修改时间、基本权限。 |
| GET `/content?path=...` | `{text,version,encoding,newline,maxBytes}`，最大 4 MiB UTF-8；二进制/超限明确拒绝。 |
| PUT `/content` | `{path,text,version}` + Idempotency-Key；先持久记录 WAITING_CLIENT，再通过独立 bulk 通道逐片写入节点。节点确认全部片后只排队任务，返回 `202 {task,upload?}`；Daemon 执行原任务真正 CommitUpload。只有 Task.SUCCEEDED 且 `task.result.version` 有实际结果后，前端才能更新服务器保存基线。 |
| GET `/download?path=...&version=...` | 有界流式下载，版本可省略以使用当前版本；每片核验路径/总长/版本/offset/hash。失败或撤销时中断响应，Content-Length 使部分下载不能伪装完整。 |
| POST `/actions` | `{action,path,target,version,targetVersion,trashId,overwrite?}` + Idempotency-Key；mkdir/copy/move/delete/restore/compress/extract，返回 202 任务。源读取与目标写入分别授权，执行前仍重新检查。 |
| GET `/trash?offset=...&limit=...` | 真实回收清单分页；恢复使用列表中的 id。 |
| POST `/uploads` | `{path,total,hash,version,sourceName,sourceModified,sourceFingerprint}` + Idempotency-Key -> `202 {task,upload}`。上传 ID 独立于 taskId，以账号及请求 ID 固定绑定。 |
| GET `/uploads/{taskId}` | 当前账号、实例与来源绑定的节点检查点。 |
| PUT `/uploads/{taskId}/chunks` | `{offset,data:base64,hash}`，每片最多 64 KiB，返回 `200 {task,upload}`。只有节点数据 sync 和检查点 sync 完成后推进 offset。 |
| POST `/uploads/{taskId}/complete` | 确认检查点全长后排队已有任务，返回 202；不直接绕过任务提交文件。 |
| DELETE `/uploads/{taskId}` | 请求取消并由节点确认清理。只清理绑定暂存；已经提交的目标不可通过此入口删除。 |

不存在目标的基线为 `missing`；已有文件/目录版本为 `sha256:<hex>`。上传返回的 `spec` 是完整来源绑定；重选来源必须用同一请求 ID 和同一名称/大小/时间/指纹/hash，再从节点 offset 续传。关闭浏览器不把未上传数据标成完成。

目录和回收 page limit 会按单帧预算下调，前端须依据 nextOffset 加载。Master 每实例文件 API 统一共享每 Server 四个 HTTP 文件操作预算；文本请求正文有解码大小上限，控制任务 payload 只保存元数据/hash，不包含正文。批量物理通道单帧 128 KiB，文件数据片 64 KiB。

## 真实验证证据

命令：

```sh
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false \
  go test -race ./internal/master \
  -run 'TestFileAPI|TestFileActions|TestBrowserUpload|TestDownloadRevocation|TestCancelAcceptedUpload' -count=1 -v
```

最终运行退出 0；包结果 `ok blora.dev/panel/internal/master 7.018s`：

- `TestFileAPIRealNodesTasksAndConditionalSaves`，1.56s：真实 TLS Master + 两个签名 Daemon；不同资源根隔离、跨多片 UTF-8 正文读写、等待实际任务提交、完成后重试复用 requestId、同 ID 不同内容拒绝、外部修改冲突保留文件、只读账号写入拒绝、路径越界拒绝、真实流式下载。任务 payload 未夹带正文。全目录搜索与 size-desc 排序在分页前生效。
- `TestFileActionsArchiveAndRecycleThroughDurableTasks`，2.18s：默认生产注册入口，经节点持久 worker 实际 mkdir/copy/move/ZIP/extract/delete/restore，GET 回收列表与恢复 ID 一致；客户端 config 宿主路径注入被请求解码拒绝。
- `TestBrowserUploadCheckpointReconnectAndCancellation`，1.73s：一个 64 KiB 分片确认后主动断开管理连接，重连更高连接代次仍保留 WAITING_CLIENT，未发布部分文件；新浏览器会话获取同一检查点、改变指纹拒绝、只读他人访问拒绝、后续片续传、最终任务提交及仅清理被取消上传。
- `TestDownloadRevocationTerminatesExistingStream`，0.30s：8 MiB 实际稀疏文件，读取中撤销 file.read，当前 HTTPS 下载以截断错误结束，后续请求 403，管理员仍正常访问。
- `TestCancelAcceptedUploadBeforeNodePreparation`，0.22s：持久任务已接受而节点尚未创建任何暂存，节点确认 NOT_FOUND 后任务可明确取消，没有生成目标文件。

fixture 中没有额外注册文件 API 的 fallback；全部入口来自生产 `New -> registerFiles`。不使用模拟文件后端。

本轮先前失败及修复：默认沙箱禁止监听套接字（`socket: operation not permitted`），按本机临时验收范围获批运行；第一次真实上传测试发现 FILE_TRANSFER_MISMATCH 错误错误映射为 HTTP 502，公共桥接已修复为 409；同时修复断线不能把 WAITING_CLIENT 自动排队执行的状态规则，再运行上述完整集合通过。

## 仍需区分的范围

- 本报告支持 F07/F08、A02/A05/A07/A08/A12 的后端子场景，不代表浏览器编辑历史/F5/多窗移签、跨节点复制/移动编排或整个验收项通过。前端与跨节点编排由对应模块继续集成。
- Windows 文件库代码及测试已交叉编译，Windows 真机未验证；真实跨文件系统回收和磁盘写满环境未运行。详细库限制见 [filesystem-library.md](filesystem-library.md)。
- 目录/元数据 hash 与每片完整源版本核验有实际 I/O 成本；还未完成 E08 混合负载吞吐/p95 测量。
- 回收与终结检查点的维护入口为 `PurgeExpired`、`PruneRecords`，Daemon 的最终维护调度需要持续接入。拷贝/解压在进程崩溃时由任务系统报告中断；尚未提交的非上传暂存对象仍需节点维护记录与清理策略。
- 文本正文未持久保存在 Master；浏览器丢失原上传来源时要等待重选。失败/未知节点结果不被当成远程保存成功；前端需按任务结果保留较新的草稿与输入。
