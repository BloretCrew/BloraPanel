# 扩展任务丢响应重试

真实取消丢回执组合已补验：私有 HTTPS fixture-1780818324、真实 Master/双 Daemon、Chromium。`npm run test:e2e:real -- --grep 'accepted extension cancellation'` 1/1（4.6s）通过。测试核对 `/proc` 命令行中的本轮配置路径与进程出生标识后暂停目标 Daemon，提交真实扩展任务；浏览器取消经 `route.fetch()` 实际送达服务端、核对 202 和持久取消键后中止响应。刷新后取消按钮禁用且没有重发取消，恢复 Daemon 后真实任务为 CANCELLED，界面显示已取消，原取消键保持不变。测试不要求任务已进入 WASI 执行阶段，取消可能在派发前完成；运行中 WASI 强制取消仍属于其他场景。finally 恢复任务自有进程并卸载扩展，夹具随后停止。

任务状态查询恢复：参考扩展增加“刷新任务状态”按钮，在读取失败后允许用户重新查询；请求期间禁用按钮并避免并发查询，切换到新任务后忽略旧查询结果。独立包重新构建后 Chromium 定向场景 1/1（8.5s）通过：模拟 GET 断网、显示错误并禁用取消，恢复网络后手动读取重新显示 RUNNING，创建请求保持一次且没有取消请求；随后取消重试场景仍通过。此项是模拟 HTTP 后端的真实浏览器交互证据，代码晚于 development-20260919-rc1。

增量完成后的完整浏览器回归：`npm run test:e2e -- tests/browser/extensions.spec.ts` 提权退出 0，6/6（20.1s），覆盖沙箱视图操作、迁移失败保留、跨设备资源校准、独立包取消恢复、管理安装/启停/回滚、opaque origin 能力桥接。此套件的 HTTP 接口均为测试路由替身，不替代下述真实 HTTPS 证据。

浏览器取消恢复补验：重新构建独立参考扩展包后，`cd web && npm run test:e2e -- --grep 'independently packaged reference'` 提权运行 1/1，通过（8.7s）。首次取消在路由层中止，刷新只查询原任务，不自动重发；第二次显式取消复用原键，显示 CANCELLED，创建任务总次数保持 1。该场景使用真实 Chromium/沙箱/恢复存储和模拟 HTTP 后端；首次取消未传到真实节点，不能称为真实节点已接受取消后的丢回执验证。无 Master 的 Vite WSS 代理拒绝连接是该测试环境预期日志。

后续取消增量：`cancelTask(taskId, requestId)` 现在显式传递稳定取消键，参考扩展在发送前将键与任务身份一起保存。SDK→宿主联测模拟取消回执丢失、重建宿主和同键重试，12/12 定向单测通过；SDK 和参考扩展构建通过。此项属于传输联测，未作为真实节点取消故障的证据。

随后补充真实 TLS Master/双 Daemon 的取消回执集成：`env GOCACHE=/tmp/blora-go-full-final54 GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test ./internal/master -run '^TestExtensionAdminAPIInstallsAndServesPackage$' -count=1` 提权退出 0（7.018s）。首次取消和同键回放返回相同任务及原 `cancellationRequestId`；轮询确认 CANCELLED 后再次回放保持修订号不变。既有扩展安装、WASI 执行、应用/节点权限和撤权断言一并通过。这是服务端真实链路证据，浏览器取消回执丢失与刷新组合仍未验证。

范围：F14/E06 的 SDK → 沙箱桥接 → Master → Daemon WASI 任务身份。SDK 曾声明 `createTask(payload, requestId)` 但实现丢弃 requestId，导致宿主每次自动生成新键。现保留请求键并将任务正文独立传递；参考扩展在发送前保存待提交内容和键，收到回执后清理。

验证环境：本机 Linux Chromium，私有 HTTPS fixture `fixture-950580611`，真实 Master 与双 Daemon。参考扩展从独立 SDK 重新构建、打包后安装。

- `cd web && npm run check && npm test -- --run`：类型检查通过，10 个文件 47/47。新增测试跨 SDK 与真实宿主传输函数，中止第一次响应后重建宿主，同键同正文重试只产生一份模拟回执；该项不是后端运行证据。
- `cd web && npm run build`、参考扩展 `npm run package`：通过。
- `BLORA_E2E_CREDENTIALS=<私有凭据路径> npm run test:e2e:real -- --grep 'independent package installs'`：1/1，通过，9.2s。浏览器拦截首次请求，先 `route.fetch()` 让真实后端接受，再中止响应。刷新后提交数仍为 1；显式重试带相同键，返回原任务 ID。真实 WASI 任务完成并核对字数结果，移窗、刷新和关闭视图未产生第三次提交。

首次浏览器执行已通过重试和任务结果断言，但末尾旧断言仍期望一次请求，实际为包含重试的两次；修正计数后复跑通过。未删除或跳过断言。

测试 finally 卸载本轮扩展，fixture 会话收到 Ctrl+C 后退出 0。该增量尚未进入 final54 归档。Windows、远程高延迟、完整扩展权限及长期故障组合不因本场景通过而完成。
