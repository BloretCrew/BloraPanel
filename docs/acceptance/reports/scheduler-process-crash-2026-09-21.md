# 定时调度进程强制终止恢复（2026-09-21）

关联 E04、F12。新增 `internal/backup/scheduler_process_crash_test.go`：独立 Go 测试子进程打开持久 SQLite、保存分钟计划并实际触发一次 occurrence；任务接纳及 schedule/fire 回执全部提交后，父测试对该进程调用 `Process.Kill`，不运行关闭回调。父进程随后重新打开同一 SQLite 并启动新的 Scheduler。

定向命令：

```sh
GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup -run '^TestSchedulerProcessCrashKeepsAcceptedOccurrence$' -count=1 -v
```

定向场景通过（race 测试进程 1.130s）。恢复后 Fire 保持 `accepted`，taskId、requestId、scheduleId、slot 与 schedule 的 LastTaskID/LastSlot 完全一致；旧 occurrence 未被再次接受，任务记录仍只有一份。父进程同时撤销授权并标记节点离线，下一分钟触发没有构造或创建新任务，证明新 occurrence 仍重新授权。

首次执行备份包全量 race 时，旧的元数据测试因当前环境 umask `0077` 把 `os.WriteFile(..., 0640)` 实际创建为 `0600`，误把源权限差异报成恢复失败。测试现在在该场景显式 `Chmod(0640)` 固定输入；修复后 `go test -race ./internal/backup -count=1` 全包退出 0（2.351s）。新增测试和修正见[进程崩溃用例](../../../internal/backup/scheduler_process_crash_test.go)与[权限输入固定](../../../internal/backup/backup_test.go)。

Windows amd64 测试程序交叉编译也通过：`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/blora-backup-scheduler-crash.test.exe ./internal/backup`。该结果只证明测试程序可构建，Windows 运行行为仍未验证。

该证据覆盖调度模块在已接受任务提交后的真实进程强制终止及 SQLite 重开；它不模拟事务中途掉电，也不替代 Master+Daemon 整体硬崩溃、物理掉电、长时墙钟或 Windows 环境验证。E04/F12 仍保持进行中。
