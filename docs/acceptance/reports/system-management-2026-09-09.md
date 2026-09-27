# 有限系统管理能力探测

新增 `internal/systeminfo`，按运行平台探测可用的服务、防火墙和计划任务后端（Linux systemd/firewalld、Windows SCM/netsh/Task Scheduler）。Daemon 通过 `system.capabilities` 上报，并提供有界 `system.services`、`system.tasks` 列表、`system.firewall` 状态、规则差异预览及受限 `ServiceAction`/`TaskAction`；Master 以 `host.manage` 权限提供查询、规则预览、服务/计划任务任务接口和 `POST /api/v1/nodes/{id}/system/firewall/apply` 受控端口变更接口。默认系统管理应用已接入节点选择、能力卡片、服务列表、计划任务、防火墙状态、规则预览和逐项确认操作按钮。

前端 Playwright 场景 `web/tests/browser/system-management.spec.ts` 使用完整 API 响应形状验证节点选择、计划任务“启用”和防火墙“确认并应用”提交，断言请求路径及防火墙 desired 内容；命令：`npm run test:e2e -- --workers=1 --grep 'system management submits' --reporter=line`，1/1 通过（2026-09-09）。该场景只验证浏览器到 Master 边界，未在当前宿主执行危险系统变更。

`internal/systeminfo` 单元测试（含服务名/动作、规则边界拒绝及注入命令验证）、Master TLS 集成测试（普通用户拒绝、非法动作和路径输入拦截）、Go 编译、vet 和 `make windows` 交叉编译已通过。服务任务的三个具体动作及计划任务 enable/disable 已加入 Master/Daemon authority 白名单。防火墙预览使用平台只读规则列举命令，Linux firewalld 端口规则新增受控 apply 任务，要求 60 秒确认期限并在中途失败反向回滚；其他平台或规则形式明确返回不支持。所有危险操作仍需在独立获准测试环境验证。

## 2026-09-10 持久确认租约

防火墙 apply 现在先把前一组规则、目标规则和 60 秒 `confirmUntil` 写入 Daemon 数据库，再执行主机变更；只有管理员显式 `POST /api/v1/nodes/{id}/system/firewall/confirm` 后才保留。取消、确认超时或管理连接上下文消失会使用独立的 10 秒回滚上下文，回滚开始前持久化 `rollback_in_progress`，成功删除租约，失败保留 `rollback_failed` 和诊断。Daemon 重启会读取这些记录并按预期当前规则校验后回滚，不会盲目覆盖外部新规则。

`internal/daemon/firewall_test.go` 覆盖显式确认不回滚、取消/过期恢复、回滚失败诊断和重启恢复；普通及 `-race` 定向测试均通过。真实宿主防火墙变更与断链演练没有执行，以免修改工作环境。
