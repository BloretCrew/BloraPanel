# SQLite 元数据事务中进程强制终止（2026-09-21）

关联 E04/F12。`internal/storage/metadata_transaction_crash_test.go` 使用项目真实 `storage.Open` 初始化的 WAL SQLite 文件。独立子进程调用真实 `Store.MetadataMutation`，在其数据库事务回调内先写入计划记录和 occurrence 记录；两次 INSERT 成功后发出就绪信号并停在回调内。父测试收到信号后立即 `Process.Kill`，因此幂等请求回执、审计记录与提交尚未执行。测试随后用 `storage.Open` 重开同一数据库。

重开后两条未提交业务记录、请求回执和对应审计记录均不存在；此前已提交的管理员记录仍在，且 `PRAGMA integrity_check` 返回 `ok`。这验证了 Blora 共用 `metadataMutation` SQLite 事务在真实进程强制终止后没有留下部分变更。

验证命令与结果：

```sh
env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-storage-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/storage -run '^TestMetadataMutationRollsBackOnProcessKillBeforeCommit$' -count=1 -v -timeout 60s
# PASS，测试本身 0.09s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-storage-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/storage -count=1 -timeout 120s
# PASS，20.570s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-storage-crash GOFLAGS=-buildvcs=false \
  go vet ./internal/storage
# PASS
```

这是操作系统进程级 SIGKILL，内核和文件系统仍在运行；测试回调写入的是调度记录形状，但没有直接杀死 `Scheduler.Tick` 的实际回调，也不模拟突然断电、设备写缓存丢失或介质损坏。因此 E04/F12 的事务中断专用集成、物理掉电和长期调度仍保持未完成。
