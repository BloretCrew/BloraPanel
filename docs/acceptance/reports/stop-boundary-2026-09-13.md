# 停止失败边界与其他节点可用性（2026-09-13）

关联A03。新增真实TLS双Daemon `TestFailedStopBlocksReplacementButOtherNodeRemainsUsable`：实例启动标记一次x，设置忽略TERM、派生sleep后代、停止1秒且不升级。发出stop后立即在另一节点启动实例，原stop必须5秒预算内FAILED；再对未退出实例start必须FAILED，runId不变、启动标记仍x。另一节点仍可实际停止。Cleanup对测试实例显式kill并等待任务完成。

命令（缓存环境沿用根Makefile）：`go test -race ./internal/master -run '^TestFailedStopBlocksReplacementButOtherNodeRemainsUsable$' -count=1 -v`，本机TLS授权运行，90516退出0（4.375s）。

同时复验真实Linux运行适配三项：`go test -race ./internal/runtime -run '^(TestParentExitAndInheritedOutputDoesNotMeanRunExit|TestStopEscalatesAndConfirmsAllDescendants|TestStopDeadlineWithoutEscalation)$' -count=1 -v`，41872退出0（1.302s）。分别验证主PID已退出但持有输出管道的后代仍存在不能报整组退出；忽略TERM后按STOPPING/KILLING/VERIFYING_EXIT升级并确认全部后代退出；不升级时按有限截止报告STOP_FAILED。

以上为Linux真实进程证据，无替身成功响应。平台使用显式PGID fallback（本机cgroup不可写）；Windows Job及可委派cgroup真机仍按E09环境限制记录。生产代码未改，无活动测试进程。
