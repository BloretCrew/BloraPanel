# 持久化变更请求键边界复验

记录日期：2026-09-14。为避免丢响应后重复副作用，本轮把扩展包安装/升级/回滚/启停/卸载、目录安装、防火墙租约确认、用户创建/改密、节点登记/身份轮换、云端工作区写入和通用任务/上传取消统一接入 `requireRequestID`；OpenAPI 同步声明 `Idempotency-Key`。只读的通知、预览和查询接口不伪装成持久化任务，因此保持原有语义。

调度计划的更新和删除随后接入同一 SQLite 请求回执事务：`SaveMutation` 在版本 CAS 与计划记录写入同一事务内保存原始响应，`DeleteMutation` 在删除与回执写入同一事务内完成；同键更新重试不再因版本已推进而冲突，同键删除重试仍返回成功，不同载荷返回 `IDEMPOTENCY_CONFLICT`。Master 调度集成覆盖更新/删除回放与冲突，备份/调度 race 及提权全仓普通、竞态回归均通过；前端删除动作把稳定键保存在窗口恢复状态中。

云工作区布局/引用同步现在也把请求键写入窗口恢复状态：首次 PUT 在网络响应丢失后保留同一键，刷新页面再点“同步布局/引用”会复用该键；成功后才清理键，内容同步仍需单独显式操作。真实 Chromium 场景 `cloud workspace sync` 1/1 通过（8.0s），路由器记录两次 PUT 使用同一键，第一次中止、第二次成功且只计一次服务端写入；测试结束后 Vite/浏览器进程已退出。

节点管理界面的登记、换钥和撤销动作现在分别保存待处理请求键：刷新或网络失败后重试沿用原键，成功下载票据或完成撤销才清理；换钥目标的请求键与节点身份一起恢复。浏览器节点场景继续覆盖票据只经显式下载交付，节点 TLS 管理集成与请求键冲突回归通过。

可恢复上传的每个分片键按原始上传键和偏移派生，完成与取消分别使用固定后缀；Master 在读取分片 JSON 或检查完成状态前统一要求请求键。上传重连/取消定向集成（含空键 400、稳定分片键和完成键）提权运行退出 0（2.180s），未确认的节点检查点仍按已有来源指纹和偏移校验恢复。

同一上传场景的定向竞态回归 `go test -race ./internal/master -run 'Test(FileAPIRealNodesTasksAndConditionalSaves|BrowserUploadCheckpointReconnectAndCancellation|CancelAcceptedUploadBeforeNodePreparation)' -count=1` 提权退出 0（5.144s）。

上传浏览器恢复场景增加了分片键与完成键断言：`npm run test:e2e -- --grep 'upload reload uses node checkpoint'` 提权运行 1/1 通过（5.6s），三次分片请求分别复用原上传键加偏移后缀，完成请求复用固定 `:complete` 键；既有响应中止后刷新仍只创建一个上传任务。Vite 代理到未启动 Master 的 `ECONNREFUSED 127.0.0.1:8443` 日志属于该替身场景的预期输出。

新增节点与上传请求键后的串行全仓回归 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-full-final54 GOFLAGS=-buildvcs=false go test -p 1 ./... -count=1` 提权退出 0；Master 125.232s、terminal 16.885s、extensions 16.389s，其余 Go 包均通过。该结果与定向 race 一起覆盖本轮服务端入口改动，仍不替代缺失平台和物理故障验收。

备份包新增的 `TestSchedulerMutationReceiptsSurviveStoreRestart` 关闭并重新打开同一 SQLite 文件后复用更新、删除请求键，原始结果仍可回放；与调度 CAS 回归的 race 定向命令退出 0（1.163s）。

OpenAPI 同步将调度删除声明为带 JSON 修订体并返回 200 JSON（含 400/404/409 错误），创建也补充 409 幂等冲突响应；文档解析仍通过 3.1.0、103 路径、120 操作断言。

默认系统/监控应用也补齐请求键现场：服务和计划任务动作、防火墙应用/确认按节点与目标保存待处理键，进程终止确认框保存键，成功后清理，失败或丢响应时重试沿用原键。提权 Chromium 浏览器系统场景 3/3（14.6s）通过；Vite 代理在无 Master 的替身测试中出现预期 `ECONNREFUSED 127.0.0.1:8443` 日志，不影响三项断言。

默认备份清理、用户创建、改密和扩展管理界面随后也保存待处理请求键：备份清理键写入窗口恢复状态，用户创建键与名称/管理员草稿一起恢复，密码表单在组件生命周期内复用键并在成功后轮换，扩展安装/升级/启停/回滚/卸载及目录安装按扩展和版本保留键。失败时不清理键，成功后清理；密码和扩展包正文仍不写入工作区恢复。提权 Chromium 浏览器备份、账户/密码和扩展管理场景 4/4（9.4s）通过，Vite 代理的 `ECONNREFUSED 127.0.0.1:8443` 仍是无 Master 替身测试的预期日志。

当前源码串行全仓复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-full-serial GOFLAGS=-buildvcs=false go test -p 1 ./... -count=1` 明确退出码 0；Master 126.349s，runlog、runtime、storage、terminal、extensions 及其余 Go 包均通过。该结果不替代 Windows 真机、远程节点和物理故障验收。

防火墙二阶段确认的幂等边界随后收口：确认请求键随 RPC 传到 Daemon，并持久化在租约/终态任务结果；同键在租约仍存或任务已完成时均返回确认成功，其他请求键返回 `FIREWALL_LEASE_CONFLICT`，不再次修改主机规则。Daemon 防火墙定向 race 29.354s、Master 系统确认回放定向 race 2.096s 通过；改动后的串行全仓普通测试退出 0（Master 126.258s），串行全仓 race 退出 0（Master 256.038s、extensions 136.535s、terminal 18.627s，其余 Go 包通过），`make check` 与前端 46/46 单测均通过。

随后补齐 Daemon 重启后的确认 CAS 边界：无内存租约时两个不同请求键并发确认只允许一个成功，获胜键可再次回放，另一键明确返回 `FIREWALL_LEASE_CONFLICT`。`go test -race ./internal/daemon -run 'Test(DurableFirewallConfirmationCASIsIdempotent|FirewallLeaseRequiresExplicitConfirmation)' -count=1` 退出 0（1.137s）；该修正后的串行全仓普通测试退出 0（Master 126.926s），串行全仓 race 退出 0（Master 256.432s、extensions 136.361s、terminal 18.604s，其余 Go 包通过）。

随后以独立缓存执行 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-race-final32 GOFLAGS=-buildvcs=false go test -race -p 1 ./... -count=1`，明确退出码 0；Master、extensions、terminal、runlog、runtime、storage 及其余 Go 包均通过。串行竞态结果覆盖当前文件请求键改动，但不替代缺失平台和故障环境。

代码入口：`internal/master/extensions.go`、`internal/master/system.go`、`internal/master/administration.go`、`internal/master/server.go`、`internal/master/node_rotation.go`、`internal/master/workspaces.go`、`internal/storage/administration.go`、`internal/storage/identity.go`、`docs/api/openapi.yaml`。请求键检查位于管理员/资源授权之后、请求体解析和副作用之前；空白、缺失或超过 128 字节的键由统一边界返回 400。用户创建、改密和工作区写入使用 SQLite 事务回执，合法重试复用首次结果；节点登记/轮换的密钥/票据仍由专用一次性流程保护。

验证命令：

```sh
GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-extkey GOFLAGS=-buildvcs=false \
  go test -race ./internal/master ./internal/storage -run 'Test(RolesRemainScopedAndAccountChangesRevokeSessions|CloudWorkspace|NodeKeyRotationKeepsIdentityAndFencesPreviousConnections|RequireRequestID|UserChangesAtomicReceiptsAndLastAdministrator)' -count=1
python3 - <<'PY'
import yaml
d=yaml.safe_load(open('docs/api/openapi.yaml'))
assert d['openapi']=='3.1.0' and len(d['paths'])==103 and sum(len(v) for v in d['paths'].values())==120
PY
```

结果：Master/Storage 定向 race 退出码 0（Master 12.868s、Storage 1.079s）；传输取消回归退出 0（1.329s）；`make check` 退出码 0；OpenAPI 解析断言通过（3.1.0、103 路径、120 操作）。集成场景覆盖用户创建稳定回放、改密无键拒绝、工作区 CAS 与稳定回放、节点轮换无键拒绝、通用任务取消无键拒绝，以及既有扩展/系统边界；无键请求不会读取包正文、查询任务或触发主机变更。前端 `api()` 对所有非 GET 请求自动生成合法键。

随后同步上传取消测试的请求键契约：`go test ./internal/master -run 'TestBrowserUploadCheckpointReconnectAndCancellation|TestCancelAcceptedUploadBeforeNodePreparation' -count=1` 退出码 0（1.562s）；完整 `go test ./internal/master -count=1` 退出码 0（125.538s）。

改密回放边界随后收口：自助改密的当前密码校验移入首次事务回调，已存在回执在会话撤销后重新登录仍可用同键复用，当前密码不进入请求摘要。`go test ./internal/master -run '^TestRolesRemainScopedAndAccountChangesRevokeSessions$' -count=1` 退出码 0（1.044s）；完整 Master 套件退出码 0（127.375s）。

改动后的存储包回归与 `make check`（`go vet ./...`）再次退出码 0；未留下调试测试或活动测试进程/容器。

文件保存、文件动作和新建上传入口随后也在实例/写权限确认后立即调用 `requireRequestID`，因此空白或超长键会在读取请求体、生成上传身份和联系 Daemon 前返回 400；原有上传分片/完成接口继续依靠已确认检查点和任务状态推进。新增文件集成回归覆盖空白保存键与 129 字节动作键，定向 `go test ./internal/master -run '^TestFile(APIRealNodesTasksAndConditionalSaves|ActionsArchiveAndRecycleThroughDurableTasks)$' -count=1` 退出码 0（2.693s）；随后完整 Master 套件退出码 0（126.836s），`make check` 退出码 0。

任务取消随后补齐了请求键的持久收据：`model.Task` 记录首次 `cancellationRequestId`，SQLite 取消事务在并发下只提交一次；任务已进入 `CANCELLED` 后再次 POST 仍返回原取消收据，不重新推进修订或重复派发。普通任务、上传和跨节点传输路径均传递请求键，任务中心为每个任务保存稳定键以支持丢响应重试。存储定向测试 `Test(Cancellation|ConcurrentCancellation)` 通过；真实 TLS Master/双 Daemon 集成 `TestTaskCancellationReplayAfterTerminalKeepsReceipt` 通过（0.139s）。

节点登记随后接入同一持久化请求回执：`POST /nodes/enrollments` 现在以管理员、请求键和节点名称绑定票据生成；同键重试返回原票据，不会再创建第二个有效登记票据，改用同键提交不同名称返回 `IDEMPOTENCY_CONFLICT`。存储 `TestEnrollmentMutationReplaysTicketByRequest` 与真实 TLS 双 Daemon 管理集成 `TestRolesRemainScopedAndAccountChangesRevokeSessions` 均通过；登记票据仍只在受保护状态库的回执中保存，审计不记录票据正文。

登记回执修正后的提权全仓验证：`go test -p 1 ./... -count=1` 退出 0（Master 126.994s、其余 Go 包通过）；`go test -race -p 1 ./... -count=1` 退出 0（Master 261.687s、extensions 142.374s、terminal 18.656s、runlog 2.265s、storage 13.649s，其余包通过）。`make check` 和前端 46/46 仍通过，未留下测试服务或容器。

节点换钥票据随后接入同一持久化回执：`POST /nodes/{id}/rotation` 将管理员、节点、请求键和 `reactivate` 绑定，丢响应重试复用原票据；同键提交不同载荷返回 `IDEMPOTENCY_CONFLICT`，不会覆盖待确认票据。真实 TLS 换钥集成覆盖同键回放和冲突，换钥后的节点身份、代次栅栏与旧连接断开仍通过。

换钥修正后的提权全仓验证：`go test -p 1 ./... -count=1` 退出 0（Master 128.965s、其余 Go 包通过）；`go test -race -p 1 ./... -count=1` 退出 0（Master 262.770s、extensions 139.824s、terminal 18.609s、runlog 2.241s、runtime 2.698s、storage 13.613s，其余包通过）。

请求键边界随后提前到实例创建、实例控制、终端创建和授权写入的请求体解析之前；定向回归退出 0（6.241s）。最新 Master 完整普通回归退出 0（127.825s），完整竞态回归退出 0（265.970s）。

Compose 项目保存的 prepare/chunk 写入入口也补齐请求键边界；前端按准备请求键和分片偏移发送稳定键，节点端已有 `saveId`/偏移检查点保证重复分片不重写已确认内容。容器管理定向回归退出 0（0.377s；需要显式 Docker 镜像的 Compose 端到端用例按环境规则跳过），前端 46/46 通过。

Compose 修正后的最新 Master 完整普通回归退出 0（129.598s），完整竞态回归退出 0（264.964s）。

换钥票据同键载荷冲突的 HTTP 错误码进一步明确为 `IDEMPOTENCY_CONFLICT`（仍为 409），真实 TLS 换钥集成回归退出 0（0.209s）。

该修正后的全仓复验：提权环境 `go test -p 1 ./... -count=1` 退出 0（Master 124.958s，其余 Go 包通过）；`go test -race -p 1 ./... -count=1` 首次仅 runlog 进程时序读 `/proc` 失败，隔离 runlog race 2.345s 通过，随后完整竞态复跑退出 0（Master 255.904s、extensions 137.279s、terminal 18.616s、runlog 2.136s，其余包通过）。`make check` 与前端 46/46 单测通过。

同一文件场景以 `go test -race ./internal/master -run '^TestFile(APIRealNodesTasksAndConditionalSaves|ActionsArchiveAndRecycleThroughDurableTasks)$' -count=1` 再跑，退出码 0（5.406s）。

管理写入口的请求键冲突错误码统一为 `IDEMPOTENCY_CONFLICT`；定向用户、角色、节点回归退出 0（1.198s）。随后完整 Master 普通回归退出 0（127.672s），竞态回归退出 0（263.528s）。首次并行回归因 `/tmp` 空间耗尽失败，清理本轮构建缓存后重跑通过。
