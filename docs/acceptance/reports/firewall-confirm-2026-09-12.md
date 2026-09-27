# 防火墙确认恢复顺序（2026-09-12）

关联 F13/E05，整项继续进行中。

原路径收到持久确认后先删除租约，再由外层执行器写任务成功；两步之间崩溃会丢失恢复证据。现在 runFirewallApply 先写成功终态，再清理确认记录，启动恢复也遵循同一顺序。清理失败仅遗留可幂等清理的记录，不重放主机规则。

终态记账失败时保留租约与可恢复任务，执行器记录 confirmation_commit_pending；启动恢复失败明确返回错误，防止后续通用恢复把任务覆盖为 INTERRUPTED。确认已提交但任务存在冲突终态时保留记录并报错。

验证：本机 SQLite、替身防火墙，不变更真实宿主机。新增 TestConfirmedFirewallRetainsRecoveryUntilTaskCommit 使用 SQLite trigger 注入任务更新失败，验证无假成功、确认记录保留、启动不覆盖任务，移除故障后恢复成功且不重放规则。现有显式确认、超时、取消和 prepared 恢复场景一起运行。

命令：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/daemon -run 'TestFirewall|TestRecover.*Firewall|TestConfirmedFirewall' -count=1`，退出0。最终具体时长及构建结果见 PROGRESS 最新检查点。

续接增量：回滚成功先持久化 rolled_back 记录，任务终态写入成功后才清理；重启看到该状态只补记账，不重放规则。回滚执行前发现当前规则已等于原规则时直接记完成，覆盖规则恢复后、完成记录前的中断窗口。prepared但未实际变更的恢复也先记任务失败再清理。应用命令失败可能含部分修改，现保留prepared记录供恢复与诊断，不直接删除。

新增 TestRollbackCompletionSurvivesTaskWriteFailure：注入任务更新失败，检查 rolled_back 记录保留、启动报错阻止通用中断覆盖；解除故障后恢复失败终态，规则恢复函数只调用一次。与现有租约测试一起以 `-run 'TestFirewall|TestRecover.*Firewall|TestConfirmedFirewall|TestRollbackCompletion'` 执行race，退出0。

firewalld 永久/运行规则合同、真实网络失联与危险环境仍须继续实现和验证，不能计为 E05 整项通过。
