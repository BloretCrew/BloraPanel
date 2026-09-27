# 防火墙预览与任务显示（2026-09-12）

关联 F02/F13、E05；仍进行中。

预览原先把 `--list-all` 的每一行与 PORT/protocol 输入比较，无法反映实际应用的端口子集。现通过 CurrentFirewallPorts 读取相同受管端口，排序去重并计算增删；端口范围等非本适配器管理的规则不混入删除项。预览按应用的同一验证规则拒绝无效端口，返回 apply=true 表示存在对应端口适配器，不能表示操作已经执行。

系统管理界面新增任务查询：接受回执后显示实际排队/执行阶段，只有 RUNNING/awaiting_confirmation 才允许二次确认。任务与原节点身份一同保存，切换筛选及刷新不改变确认目标。防火墙草稿和节点选择立即持久化；旧预览响应在节点或草稿变化后丢弃。活动任务期间禁止覆盖待确认任务入口，最终成功以任务终态为准。

验证：systeminfo 全包 race 退出0（1.074s），新增受管端口差异、重复/非受管规则及无效规则提前拒绝测试。Chromium API路由替身 `npm run test:e2e -- tests/browser/system-management.spec.ts --workers=1 --reporter=line` 1/1（5.9s）通过，覆盖排队不允许确认、切换节点、刷新后确认原目标。make check/build/windows 退出0，前端类型检查退出0；前端最终生产构建记录在进度。

尚未修复：ApplyFirewall 仍将运行态读取结果用于永久规则修改并调用 reload，必须改为明确区域的运行/永久双快照、受控分别应用和分别恢复。官方 [firewall-cmd 合同](https://firewalld.org/documentation/man-pages/firewall-cmd.html) 明确运行与永久配置独立，reload 会以永久配置替换运行配置，因此现有逻辑可能影响无关运行态规则。此缺口未计通过，本轮没有运行真实宿主防火墙变更。
