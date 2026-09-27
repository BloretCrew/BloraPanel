# 任务时间与持久阶段记录（2026-09-12）

关联F09，整项仍进行中。

任务列表现在显示发起者、接受/派发/最近更新时间和自接受起的耗时（包含等待）。终态使用已记录更新时间固定耗时；缺少或相互矛盾的时间明确提示，不用当前时钟补造完成时间。列表可打开专用任务详情。

GET /tasks/{id}/events读取持久task_events中的状态和阶段诊断，按任务修订号倒序分页，当前任务授权逐请求复核。输出只有revision、recordedAt、state、phase、error及截断标识；不读取返回执行payload或result。phase最多512字符，error最多4096字符；每页最多100条。详情默认50条，支持更早/最新/刷新，分页游标保存在该视图现场，历史页不自动轮询；读取失败显示错误，隐藏缓存记录。

阶段记录是已经持久化的任务接受、状态变化和诊断，不等同于被执行程序的stdout。现有实例运行日志与终端仍由各自真实服务承载。不能从此子项推断所有执行器已有完整输出日志或进度。

验证：

- `go test -race ./internal/master -run TestTaskStagesPaginationProjectionAndCurrentAuthorization -count=1`：真实TLS，57个修订跨页无遗漏/重复，诊断长度与截断、payload排除、无效分页拒绝、撤销读取授权后404，1.973s通过。
- `cd web && npm test`：9个文件、41/41通过，新增终态耗时冻结、等待时间、缺失/逆序时间2/2。
- `npm run build`与`make check build windows`通过；Windows仅构建，不是真机运行证据。
- 本地真实HTTPS双节点fixture-186544590：`npm run test:e2e:real -- tests/real/notifications.spec.ts --grep 'task completion notifies' --workers=1 --reporter=line` 1/1（3.8s）。真实mkdir完成后通知定位对应任务，发起者与会话身份一致，耗时呈现，从QUEUED到SUCCEEDED阶段均可读取，通知移除后刷新不重放。fixture已Ctrl+C停止。
- OpenAPI3.1解析通过，共103路径。

全范围仍保留执行器输出/进度覆盖、任务长期保留与故障组合、其余SDK资源能力和平台验收。此前全仓race证据不冒充本次阶段接口的全仓回归。
