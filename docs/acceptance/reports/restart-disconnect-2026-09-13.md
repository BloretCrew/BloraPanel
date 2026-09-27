# 重启任务运行期间管理断线（2026-09-13）

关联A04/A05。加强 `internal/master/integration_test.go` 的真实双节点生命周期测试，旧进程忽略TERM并派生sleep后代，正常停止1秒、强制确认2秒。提交同requestId两次restart，先验证同taskId；等待Master任务RUNNING，关闭源Daemon管理连接，等待更高连接generation重连，再用相同requestId提交，必须仍为原taskId。随后等待原任务SUCCEEDED并取得实际runId；完成后重试仍原任务，再次管理重连不改变runId，最终实际停止。

命令：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestTwoNodesAuthorizationAndDurableLifecycle$' -count=1 -v`。6548退出0（7.392s），同时复验原有跨节点授权、普通用户原生隔离拒绝和事件撤权终止。自建进程及TLS节点在测试结束清理，生产代码无修改。

此处故障注入以Master任务RUNNING为边界，不伪称已精确截在Daemon某个停止系统调用内部。Master进程中途重启、浏览器失联提示及实际Daemon崩溃仍需各自直接证据。A04/A05保持进行中。

后续Master服务/SQLite重开：新增 `TestRestartTaskSurvivesMasterAndDatabaseReopen`，复用实际双Daemon和稳定TLS地址的重开fixture。真实实例每次启动向专有文件追加x，初次启动后为x；restart进入RUNNING时关闭Master服务、关闭SQLite并重开，节点保留独立执行记录并重连。同requestId仍原taskId，原任务SUCCEEDED后文件精确xx。再次重开Master/SQLite、重试已完成请求，仍原taskId和runId，文件仍xx。测试Cleanup实际停止自建运行。

命令同Go环境，`go test -race ./internal/master -run '^TestRestartTaskSurvivesMasterAndDatabaseReopen$' -count=1 -v`，69010通过，测试8.12s。这是真实数据库关闭/重开和服务恢复，但测试宿主进程仍存在；不能称为独立Master进程SIGKILL验证。生产代码无修改，A04仍需该硬中断直接证据。

独立Master硬中断后续通过：扩展 `scripts/package-smoke.py --master-crash`，使用实际发行归档 `dist/releases/development-20260913`，全新 `/tmp/blora-release-smoke-wkzbiotd` 状态、Master及两Daemon独立子进程。创建有限20秒测试运行（每次启动追加x、忽略TERM），初次启动x；restart任务RUNNING时对子进程Master发送SIGKILL并核验返回码-9。以原TLS/SQLite目录重新启动Master，等待节点上线，再提交相同requestId，必须原taskId；任务成功后启动标记恰为xx，完成后再重试仍相同taskId/runId且xx。

运行 `python3 scripts/package-smoke.py dist/releases/development-20260913 --master-crash`，97578退出0，独立初始化/SDK构建签名/双节点上线及硬中断恢复均PASS。finally通过真实停止任务结束实例，再停止全部自建进程。私有诊断目录保留。此证据对应标明的发行包版本；后续最终发行包仍需重新打包，不改写旧归档。

浏览器响应丢失后续通过：`desktop.spec.ts`新增accepted restart with dropped，在实际实例启动后点击重启；拦截器先 `route.fetch()`让真实Master接受并确认202，再向页面 `route.abort('connectionreset')`。界面出现错误、确认现场保留；刷新后显式再次确认。所有实际再次发出的请求key必须与初次一致，后端重启任务只有一个，taskId/requestId与初次接受记录相同，实际runId与成功任务结果相同。finally实际停止该实例。

fixture `.local/fixture-1671058723`，web目录设置其私有凭据后运行 `npm run test:e2e:real -- --grep 'accepted restart with dropped' --reporter=line`，73971退出0，1/1（4.4s）。故障注入只丢交付响应，不伪造接受或任务结果；页面可从任务索引复用已接受任务，因此不要求它一定再次POST。fixture已Ctrl+C停止。结合同key集成、A15不同请求串行和独立Master SIGKILL证据，A04各明确分支已有Linux直接验证；Windows运行仍见E09。
