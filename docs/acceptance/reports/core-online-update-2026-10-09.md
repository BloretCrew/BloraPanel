# 预构建 Release 在线更新验收

日期：2026-10-09。范围：本轮工作树新增核心更新能力，晚于公开 `v0.1.0-beta.1`。未提交、推送、部署或发行；没有修改此前平台与 E08 验收结论。

## 实现合同

- 管理员设置更新 Master，节点管理更新 Daemon；采用完整预构建 Release，生产运行机器无需 Git/Go/Node 或编译。
- 默认本项目 beta 通道，支持 stable 及 HTTPS GitHub 兼容 API 镜像。标签固定提交、SHA256SUMS、CORE-UPDATE 兼容元数据、包内 MANIFEST 与二进制信息共同核对。
- 下载、校验和切换状态持久化；请求幂等，后台检查不占用节点 RPC 的 20 秒请求窗口。
- 稳定启动器持有排他锁，串行切换一个管理子进程；候选未就绪保留旧版本。私有控制文件兼容 Windows，无运行中二进制覆盖。
- Daemon 排空已接受操作，不取消实例任务；核对运行身份与独立输入/日志助手。管理重连不把已接受操作绑定到旧连接的取消信号。
- 切换前重新检查管理员身份及协议；仅 `host.manage` 不具有核心更新权限。存储迁移指纹变化、对端不兼容或原实例归属不明会拒绝。
- 原安装目录的配置、身份、数据库和实例数据保持原位置；默认捆绑前端随 Master 更新，自定义静态目录不覆盖。

## 实际运行结果

Go 命令使用私有可写缓存：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go`。本地 IPC 测试需获准的回环监听权限。

| 验证 | 命令 / 证据 | 实际结果 |
| --- | --- | --- |
| 双程序真实在线更新 | `python3 scripts/core-update-smoke.py --report .local/core-update-smoke-2026-10-09.json` | 3 个真实场景通过：Master 更新、Daemon 更新、篡改下载拒绝。两次更新保留相同 PID 出生身份、runId 与业务心跳；两端回执重放均完成且不重复更新，稳定启动器 PID 保持不变，默认 Web 切换到新版本 |
| 原生输入/日志独立恢复 | `go test -race ./internal/daemon -run 'Test(PrepareCoreRestart\|CoreRestart\|DaemonCoreUpdateLogHelper)' -count=1` | 1.841s 通过；实际 Close/New 后原运行、独立 stdin 和日志继续可用；缺失助手阻止更新，已有操作排空、超时重新开放及未知已执行任务守卫通过 |
| Release 引擎、启动器、启动恢复 | `go test -race ./internal/coreupdate ./internal/bootstrap ./cmd/master -count=1` | 分别 17.561s、4.524s、9.448s 通过；最后的启动器目录 fsync 补充另复验 4.511s 通过。HTTPS 下载真实预构建测试程序，清空 PATH 后仍可校验执行；包括 tar/zip 路径、链接、冲突、摘要、版本、存储指纹与候选失败回退 |
| 受影响包完整竞态 | `go test -race ./cmd/master ./cmd/daemon ./internal/bootstrap ./internal/coreupdate ./internal/daemon ./internal/master ./internal/storage -count=1` | Master 290.768s、Daemon 33.898s、Storage 17.750s、Daemon 命令 1.026s 等通过；总命令因下述旧启动测试入口失败退出 1，之后修正入口并复验命令包通过，不冒充该次总命令全绿 |
| 授权与幂等 API | `go test ./internal/master -run TestCoreUpdate -count=1` | 4 项定向测试 0.405s 通过；含 8 个入口、离线/旧节点兼容、管理员撤销和核心权限不可由 host.manage 代替；也被上述完整 Master 竞态覆盖 |
| 网页更新交互 | `BLORA_E2E_FRESH_SERVER=1 BLORA_E2E_PORT=5189 npx playwright test tests/browser/core-updates.spec.ts --workers=1 --project=chromium`（web 目录） | 2/2 通过，11.4s；固定目标、未知响应后相同回执重试、真实 progress 元素和非管理员入口保护。仅客户端边界替身，不冒充下载/实例测试 |
| 类型与前端构建 | `npm run check && npm run build`（web 目录） | 最终通过，Vite 4.25s；保留已有大 chunk 提示，未新开性能优化范围 |
| 静态与平台构建 | `go vet ./...`；`make build windows` | 通过；Master/Daemon 两平台可编译，Windows 新增测试程序也交叉编译通过 |
| 本地包格式 | `python3 scripts/package.py --version development --output /tmp/blora-online-update-packaging-final`；独立读取六个归档及 SHA256SUMS | 通过，6 归档共 2472 个载荷逐一摘要核验；四份组件归档的 CORE-UPDATE 与外部资产一致。开发包的工作树修改标记为 true，不能当在线发行包；没有新公开 Release |

## 失败及修复记录

1. 初次受限沙箱禁止回环 socket：独立 log helper 就绪 EOF、httptest 监听权限错误。保留失败，允许临时本地回环后原场景竞态通过，不跳过或改变断言。
2. 旧 Master 启动测试把自身 Go 测试程序当服务入口；新启动器重新执行时，测试框架先解析私有管理参数，导致无法启动。增加仅该测试助手使用的 TestMain 私有入口派发，实际 TLS 健康、重复初始化拒绝、身份状态及两次优雅重启断言保持原样，修正后 9.665s，最终 9.448s 通过。
3. 候选切换前后的状态确认需要等待启动器持久提交，加入 ReadyAndWait；失败记录在新尝试前清除，避免同一目标的成功重试被旧失败标记覆盖。版本目录每次尝试独立，私有指针路径拒绝越界/目录链接。Unix 指针替换同步父目录，Windows 使用 MoveFileEx WRITE_THROUGH。

## 边界与剩余外部验证

- 新 Windows 启动器、真实 Release 切换与实例 Job keeper 连续性尚未设备运行；交叉编译和原 Windows 14 项报告不升级成新 PASS。
- 当前公开 beta 没有新更新元数据/入口，检查会明确要求维护安装。首次采用须人工安装更新能力；新 Release 发布不在本轮授权范围内。
- 只允许相同存储迁移指纹及适配协议的在线维护更新；未实现跨不兼容数据库的自动补丁链或降级。大架构/稳定启动器本身变更仍需维护。
- 管理连接可短暂断开，交互终端会话可能结束，上传可能需要续传；通过证据是运行实例保持，不是全部管理连接不间断。
- 完整真实演练使用私有 HTTPS GitHub 兼容测试源和真实二进制，没有接触生产实例或公共 Release 发布。原始临时凭据/日志不公开，测试结束清理自有进程和目录；机器参考及旧 E08 结论不改写。
