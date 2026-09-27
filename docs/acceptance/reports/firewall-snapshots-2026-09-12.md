# firewalld 双配置快照（2026-09-12）

关联 F13/E05，整项仍进行中。

新建生产租约在变更前读取默认区域名称并将区域、运行端口、永久端口共同持久化。后续修改/恢复都显式绑定该区域，不随默认区域变化。分别修改运行和永久端口，保留各自非受管规则，读回两套配置核对结果；新租约生产路径不调用 reload 或 runtime-to-permanent。规则原本不一致时，回滚恢复两份不同原值。

应用中途失败会立即尝试有界快照恢复，并保留租约记账证据。恢复允许本次端口增删产生的部分进度；发现其他端口增删时拒绝覆盖。它不能区分外部操作者对本次相同端口作出的相同修改，这仍需要独占管理约定或更强的外部版本支持。

依据 [firewall-cmd 官方合同](https://firewalld.org/documentation/man-pages/firewall-cmd.html)，运行和永久配置独立，分别调用修改命令可避免全局重载替换无关运行规则。本轮仅运行命令替身，未触碰真实宿主防火墙。

验证：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/systeminfo ./internal/daemon -run 'TestFirewall|TestApplyFirewall|TestRollbackFirewall|TestRecover.*Firewall|TestConfirmedFirewall|TestRollbackCompletion' -count=1` 退出0（systeminfo1.216s、Daemon2.074s）。新增状态化命令替身验证区域参数、不同配置、非受管端口范围保留、永久修改失败后的部分恢复、外部规则冲突，并拒绝任何全局reload调用。make check/build/windows退出0。

后续双份预览：区域、运行/永久增删均展示；planHash绑定区域、两份完整端口快照和目标规则（集合排序去重），Master要求提交摘要，Daemon重新采集并在写租约/变更前核对。目标规则、区域或任一配置变化均须重新预览。界面无预览不能提交，修改草稿或节点会清除预览；OpenAPI已更新。

新增验证：systeminfo/Daemon定向race退出0（1.124s/1.913s）；摘要绑定、顺序去重等价与四类字段变化测试race1.032s。Chromium替身浏览器双份预览、摘要提交、刷新与原节点确认1/1（7.5s）；前端check/build、make check/build/windows通过，OpenAPI3.1解析100路径。Master预览必填/期限/权限TLS定向验证已运行，最终结果见PROGRESS。

Daemon命令链路补充：新增 `TestFirewallDaemonCommandSnapshotAndStalePreview`，临时PATH的firewall-cmd包装器启动独立Go测试子进程，在测试目录保存两套规则，实际执行生产Snapshot路径。覆盖错误planHash在租约写入前拒绝、运行/永久原值持久化、应用后的取消、两份原规则恢复及CANCELLED终态。命令 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/daemon -run TestFirewallDaemonCommandSnapshotAndStalePreview -count=1` 退出0，28.946s；`GOCACHE=/tmp/blora-go-grant-idem make check`退出0。测试对象由t.TempDir清理，无残留会话。

兼容清理完成：旧ApplyFirewall/RollbackFirewall及其全局reload代码已删除，缺少双快照的调用明确拒绝。原命令失败恢复、恢复失败诊断、父context取消后恢复及外部规则冲突测试已迁移到Snapshot接口，保留原场景含义。最终 `go test -race ./internal/systeminfo ./internal/daemon -run 'TestApplyFirewall|TestRollbackFirewall|TestFirewall|TestRecover.*Firewall|TestConfirmedFirewall|TestRollbackCompletion' -count=1` 退出0（systeminfo1.168s、Daemon29.595s），包含上面的独立命令子进程测试；make check/build/windows退出0。源码检索确认无旧函数调用和reload参数。

此证据包含真实子进程和生产调用链，但firewalld响应仍为测试替身，不能证明真实防火墙变更或失联保护。旧租约没有区域/永久快照，不能可靠重建原永久规则，生产恢复保留诊断并要求人工核对，禁止猜测恢复。真实失联与危险环境、Windows防火墙仍未验证或未完成。
