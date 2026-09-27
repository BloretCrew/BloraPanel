# 站内通知与扩展 SDK（2026-09-12）

关联 F09/F14/E06，整项仍进行中。

SDK 新增 notification.publish / host.notify，独立参考包通过真实 transport 使用。Master POST /extensions/{id}/notifications 重新校验启用能力与 app.use，只接受标题/正文并返回绑定 appId；不能冒充其他来源。内容为纯文本，标题256 UTF-8字节、正文4096字节。桌面调用者身份绑定原视图，异步回执遇到注销/工作区更换/视图销毁拒绝写入。

公共通知服务与任务栏通知列表落地：当前账号工作区恢复日志立即记录，刷新不重发；最多100条、每应用每秒5次、可选key替换旧项（重复key也计发送频率）。通知不抢焦点，用户点击后复用原标签或打开原应用资源；显式移除也持久化。SDK 不申请浏览器通知权限、不发给其他用户。云端仅布局同步在前后端均排除通知内容，显式工作内容同步保留。

验证与结果：

- `make sdk`：SDK检查/构建及独立三个版本包干净构建通过。
- `go test -race ./internal/master -run TestExtensionNotificationAuthorizationAndBounds -count=1`：真实TLS权限/禁用/冒充来源/长度边界通过（1.756s）。
- `go test -race ./internal/master -run 'TestCloudWorkspaceSnapshotsOwnerRevisionAndContentPolicy|TestExtensionNotificationAuthorizationAndBounds' -count=1`：最终通知与云端选择策略通过（2.195s）。
- `cd web && npm test && npm run build`：37/37、类型检查与生产构建通过；含transport11/11、来源绑定/同key限流/100条边界2/2。
- 本地真实HTTPS双节点 fixture-3796281576：`npm run test:e2e:real -- tests/real/extensions.spec.ts --grep 'notification retains' --workers=1 --reporter=line` 1/1（4.7s）。独立包安装、纯文本、来源标识、立即刷新不重发、原标签复用、移除后刷新均通过。fixture已Ctrl+C停止，未触碰生产。
- `make check build windows` 在通知接口接入后通过；云端过滤最后改动后再次执行，结果见PROGRESS。OpenAPI3.1解析通过，共101路径。

后续后台任务接入：`/events?notifications=1` 复用持久 task_events 和 WSS/Protobuf，首次从当前持久边界开始，after续传；Snapshot推进游标，TaskEvent只包含获准根任务的终态最小摘要。原通用事件流不变。默认桌面持续订阅并重连、心跳，不依赖任务窗口；通知与游标同一恢复日志提交。系统通知只有用户在任务中心显式启用且浏览器授权后发送，支持关闭，点击可打开任务详情；保存游标后才尝试系统提示，不重放已确认项。

定向真实WSS测试覆盖首次边界、用户授权过滤、去除payload/error、重连不重放，通过race（2.321s）。真实双节点文件mkdir任务完成通知、未打开任务中心仍接收、点击精确任务详情、移除后刷新不重复，浏览器1/1（3.6s）通过。前端38/38、类型/构建、make check/build/windows通过。恢复日志错误只记状态的行为已在扩展restore/通知回执处显式检查，避免误报保存成功。

最终三场景组合首轮2/3通过，独立样例笔记恢复失败：输入框在服务器数据读取完成前已可输入，恢复逻辑随后覆盖新输入。修复为初始控件禁用，现场恢复和事件绑定完成才启用；三个样例包已重建，并增加显式阻塞数据读取的浏览器复现保护，复验结果见PROGRESS。此失败不计通过。

最终三场景复验3/3（9.8s）通过，含初始化数据读取确定性阻塞、恢复日志故障时禁止远程PATCH、修复存储后同请求重试、扩展通知与真实文件任务通知。fixture-1098943595已Ctrl+C停止。新增通知日志失败单元2/2通过；全仓race最终退出0，extensions164.323s、Master175.982s通过。其后任务历史与完整统计的增量另见 [历史报告](task-history-2026-09-12.md)。

真实操作系统的系统通知展示/点击仍未验证；完整E06能力组合、Windows真机及远程故障验证仍保留。
