# Scheduler.Tick 持久阶段进程崩溃恢复（2026-09-21）

关联 E04/F12。新增 `internal/backup/scheduler_stage_crash_test.go`，使用两个真实测试子进程和同一 SQLite/WAL 持久化实现，在 `Scheduler.Tick` 实际执行路径的两个独立持久阶段强制终止进程。

在 `pending` 场景中，调度器已持久化 occurrence slot/NextRun，随后进入 `BuildTask` 前等待；父进程杀死它。重开数据库确认 schedule 的 `Pending` 尚在、无 fire/task。新调度器在同一分钟补做该 occurrence，构造并接受恰好一个任务，随后同槽重放返回空。

在 `prepared` 场景中，调度器已持久化 `Fire{State:"prepared"}` 及 taskId/requestId，但尚未调用 `Store.Accept`；测试在实际接受前的二次授权回调处等待并被父进程杀死。重开后确认 task 表仍为空。新调度器接纳原 Fire 中同一 taskId/requestId，完全不重跑 `BuildTask`；随后同槽重放返回空。两个阶段恢复后 schedule 的 Pending 被清除、LastTaskID/LastSlot 与唯一任务一致。

验证命令与结果：

```sh
env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched-stage-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup -run '^TestSchedulerProcessCrashReconcilesDurableOccurrenceStages$' -count=1 -v -timeout 90s
# PASS，pending 与 prepared 子场景均通过，测试包 1.224s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched-stage-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup -count=1 -timeout 120s
# PASS，8.017s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched-stage-crash GOFLAGS=-buildvcs=false \
  go vet ./internal/backup
# PASS

env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched-stage-crash GOFLAGS=-buildvcs=false \
  go test -c -o /tmp/blora-backup-scheduler-stages.test.exe ./internal/backup
# PASS；只证明 Windows 测试程序可编译，未执行 Windows 行为
```

这是进程级 `SIGKILL`，发生在已持久化阶段之间；不模拟 SQLite 写入/提交调用内部的设备掉电、文件系统缓存丢失、长周期墙钟或 Windows/systemd 运行。
