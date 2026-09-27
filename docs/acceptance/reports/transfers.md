# B06 跨节点传输真实链路报告

日期：2026-09-09。此报告覆盖独立传输协调器，不代表 F07/A12 或全范围验收完成。

## 已实现合同

- `internal/master/transfers.go`：Master 持久化 `transfer.copy` / `transfer.move` 父任务；普通节点 dispatcher 与失联处理跳过本地任务。默认 `New` 注册接口并启动两个有界工作槽，`Close` 在获取连接锁之前停止、等待协调器；关闭不取消已受理的节点任务。
- `POST /api/v1/transfers` 接受 `{source:{instanceId,path,version},target:{instanceId,path,version},move}`，以 `Idempotency-Key` 去重并返回 202 `{task}`。实例配置和资源身份由服务端绑定，正文不进入控制任务。
- `GET /transfers/{id}` 返回任务与有界汇总；`GET /transfers/{id}/entries?offset=&limit=` 分页返回目录清单和逐项状态；取消可走 `POST /transfers/{id}/cancel` 或通用任务取消接口。
- 源版本、目标基线、目录清单在任何目标写入之前核对；目录清单上限 4096 项/4 MiB，单文件 1 GiB，总内容 4 GiB。清单和进度分开存储；单次调度最多中转八片，每片 64 KiB。未缓存整份文件正文。
- 同节点不同实例、不同节点均采用真实对象关系校验。库以打开的对象 handle 获取 Linux filesystem ID/inode 或 Windows volume/file ID，核对固定根的 canonical 名称并记录实际祖先对象；返回 machine/root/object/ancestor/path 指纹，不返回宿主明文路径。清单持久化初始源对象/源根/目标根，每批重核，回收删除额外绑定初始源对象身份。相同对象、硬链接别名及目录包含关系在写入前拒绝；不以节点或实例 ID 推断物理独立。
- 分片绑定源节点、资源、路径、长度、内容 hash、修改时间、权限位和上传 ID；实际节点 checkpoint 决定恢复偏移，每片仅在节点同步数据和检查点后推进。每次 RPC 与子任务派发重新验证源读/目标写权限，移动额外验证源写权限和固定根配置。
- 目标 mkdir、上传提交、目录元信息、源回收删除均是稳定 requestId 的持久化 Daemon 子任务。仅上传提交使用已有幂等 CommitUpload 原语，在 Daemon 中断后允许有界新受理提交认证；结果不明的 mkdir/metadata/delete 不自动重放。
- 已有子任务回执必须同时匹配 actor、action、resource、完整规范化 payload 和 `transferParentId`；浏览器猜中的 requestId 不能借用另一项删除的成功结果，清理也不会取消或清除其他任务对象。
- 移动阶段为 copy → verify_destination → metadata → verify_source → delete_source。必须收到真实源删除子任务成功才设置 `sourceDeleted`。源改变、源消失或删除结果未知不会冒充成功。取消等待子任务实际结果、精确清理本次 upload ID，保留已提交目标及 `partial` 事实；撤权后仅允许自己的 staging 清理与结果查询。
- 已存在目录采用显式版本 merge，逐文件覆盖条件固定，保留目标额外文件；不把源目录元信息套到已有目标目录。新文件及新目录保留普通权限位、修改时间，不提供 owner/ACL/特殊位权限。

## 已运行证据

环境：Go 1.25.9，Linux amd64，本任务临时文件目录、随机测试账号、两个真实 Daemon、loopback TLS Master。测试使用默认生产注册，无注册 fallback。为允许 loopback 监听使用了已批准的提权测试执行。

缓存环境：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。

`go test -race ./internal/master -run '^TestTransfer' -count=1 -v -timeout=180s`：退出 0，48.555s，十一项测试通过：

1. 真实目录 copy/move、空目录、二进制内容、权限位和修改时间、请求去重、分页清单、独立源回收子任务。
2. 分片后改变来源，保留新来源；取消目录传输后保留已经提交文件，仅清理本次临时上传。
3. 源/目标授权、目标已存在冲突、4096 项预算在写入前拒绝。
4. 真正关闭 Master 和 SQLite 后重新 Open/New，两个节点重连、源目标字节一致，稳定键只产生一个最终上传提交子任务。
5. 真实关闭目标 bulk 连接后重连续传；中途撤销目标写权限后拒绝继续、清理 staging；原作者可查看清理确认，不能继续列清单。
6. 目标确认后修改源内容或移除源路径；移动失败且目标保留，绝不把源消失解释为成功。
7. 既有目录 merge、文件条件覆盖、目标额外文件保留、已有目录权限位不被源覆盖。
8. 在真实 TLS 流程中注入持久化的“源删除回执丢失”Interrupted 子任务，再重启 Master；父任务保持来源结果不明，源和目标都保留，没有第二个删除任务。此项是日志故障注入，不是实际 Daemon 崩溃实验。
9. 同节点不同实例的独立根目录移动成功；两个实际 Daemon 在同一宿主访问重叠 native 根时，源目标同目录、目标是源后代、硬链接别名均拒绝，真实来源保留。
10. 目标复制确认后，把来源替换为内容、权限位、修改时间相同但 inode 不同的文件，移动失败并保留新来源。
11. 通过真实文件 API 用猜中的子任务幂等键先完成另一份文件的获准删除，然后重启 Master：其成功回执不能认证移动的源删除，原来源和已复制目标保留，本次传输明确失败。

`go test -race ./internal/filesystem`：退出 0，1.287s，包含新增元信息版本冲突、取消、链接、特殊权限位拒绝、上传元信息持久身份、跨根别名/硬链接/目录包含关系、同内容对象替换防删、授权根 canonical 名称被替换后拒绝的测试。

`go vet ./internal/master ./internal/filesystem`：退出 0。

2026-09-09 追加验证：`GOCACHE=/tmp/blora-go-transfer-cache go test ./internal/master -run 'TestTransfer' -count=1` 退出 0，11 项跨节点传输集成测试通过（25.830s），覆盖持久清单/检查点、断线续接、Master 重启、权限撤销、源变化、目录 merge、同节点隔离和未知删除回执防重放。

`GOOS=windows GOARCH=amd64 go test -c ./internal/filesystem -o /tmp/blora-filesystem-windows.test.exe`：退出 0；只证明 Windows 编译通过，未运行 Windows 真机测试。

## 平台核实与边界

Windows 通过原对象 handle 的 [ReOpenFile](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-reopenfile) 获取元信息访问权，再使用 [SetFileTime](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-setfiletime)，不重新按用户路径打开对象。该 API 要求 `FILE_WRITE_ATTRIBUTES`；[FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers) 要求可写文件句柄。目录元信息延续库既有的进程崩溃恢复保证，不声称 Windows 断电目录持久性；Unix 时间按微秒精度保存，Windows 权限位服从 Windows 文件系统的只读语义。

尚未验证：真实目标磁盘 ENOSPC、实际 Daemon 在提交中崩溃、Windows 真机、跨机器文件系统组合和性能/内存 p95。现有失败和限额测试不代替磁盘满实验。

关系检查使用已打开对象的实际文件系统身份。Windows 的 [FILE_ID_INFO](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_id_info) 提供 volume serial 与 128 位 file ID；Linux 通过 `Fstat`/`Fstatfs` 获取 inode、filesystem ID。未在真实共享 NFS/SMB、重挂载或 mount namespace 别名环境运行测试；相关组合继续保留未验证状态。外部程序仍可能并行改命名空间，系统在操作前重核版本/身份，出现未知回收记录时报告结果不明，不自动重放删除。
