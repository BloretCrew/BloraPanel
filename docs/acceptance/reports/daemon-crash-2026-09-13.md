# 独立Daemon强制中断与原运行恢复（2026-09-13）

关联A10。`scripts/package-smoke.py --daemon-crash` 使用development-20260913独立归档和新临时状态，创建实际原生有限运行；第一次启动在私有目录追加x。根据实际所属节点名称映射到脚本创建的Daemon子进程，SIGKILL并核验退出码-9，以同配置/身份目录启动新Daemon。必须观察同nodeId、新startupId、ONLINE，实例RUNNING及原runId，启动标记仍x。

随后仍对该原运行执行真实restart，并在任务RUNNING期间SIGKILL Master；恢复后原任务成功，标记精确xx，完成后重试不产生新运行。最后通过任务停止实例并关闭所有脚本创建的进程。

命令：`python3 scripts/package-smoke.py dist/releases/development-20260913 --daemon-crash`，44796退出0；临时目录 `/tmp/blora-release-smoke-zoh3sha_`，独立SDK构建/签名、初始化、双节点上线、Daemon硬中断和Master硬中断均PASS。未修改生产代码，私有诊断目录保留。

此场景直接覆盖Daemon崩溃时已启动进程的身份恢复，并验证恢复后可以控制原运行；Daemon恰在任务中途崩溃的结局仍须补直接场景，A10保持进行中。Windows Job真机另列环境限制。

任务中途后续通过：新增 `--daemon-task-crash`，实际restart任务必须报告STOPPING后才SIGKILL所属Daemon，核验返回-9。原身份目录启动后任务必须为INTERRUPTED、阶段daemon_restarted_reconcile_required；同requestId重试仍原taskId及INTERRUPTED，原runId未变、启动计数仍x。finally显式执行stop并等待SUCCEEDED，再停止全部自建进程。

`python3 scripts/package-smoke.py dist/releases/development-20260913 --daemon-task-crash`，27806退出0；临时目录 `/tmp/blora-release-smoke-fnx8gqb6`，独立初始化、双节点及任务STOPPING硬中断均PASS。这是明确Daemon执行阶段的硬中断，不是仅Master任务RUNNING推断。未修改生产代码；不能把“中断后不重放”表述为任务成功完成。

运行身份边界补证：新增 `TestExistingPIDWithDifferentBirthIdentityIsNeverAdoptedOrSignalled`，对实际仍存活的测试PID注入不匹配StartTicks记录；Observe与强制Stop均必须ErrUnknown，正确记录仍能观察存活原运行。与 `TestUnknownOwnershipNeverSignalsProcess`、`TestRecoveredIntentFindsRunMarker`、`TestRecoverUsesPersistentIdentityAfterManagerReplacement` 一起执行race，67725退出0（1.058s）。该测试模拟过期出生标识，不声称操作系统已实际复用PID。

结合独立Daemon稳定运行SIGKILL、STOPPING中途SIGKILL、原身份恢复与明确INTERRUPTED，以及上述出生标识/运行标记边界，A10的Linux原生场景有直接证据。Windows Job运行及物理掉电仍独立记录环境限制。
