# 备份归档写入中进程强制终止恢复（2026-09-21）

关联 E03/F12。新增 `internal/backup/archive_crash_test.go`，使用当前平台的真实文件系统服务、备份仓库和 64 MiB 稀疏源文件。独立测试子进程以 store 模式执行实际 `Manager.Create`；父测试观察私有 `pending/<id>.zip` 增长到 1 MiB 以上、仍小于源文件后立即 SIGKILL，确保进程在 `packing` 阶段、ZIP 尾部目录/摘要/快照发布前退出。

父进程用原仓库和状态目录重开 Manager，并以相同意图 ID 调用 Create。系统明确返回 `ErrInterrupted`，状态为 `interrupted`、`unknown=true`；带对象身份的部分归档被清除，没有可 Inspect 的快照。随后用同一源目录的新备份 ID 创建小型备份，结果 `succeeded` 且 Inspect 验证通过，表明仓库完成中断清理后可继续工作。

验证命令与结果：

```sh
env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-backup-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup -run '^TestBackupCreateProcessKillDuringPacking$' -count=1 -v -timeout 60s
# PASS；进程在 ZIP 已写入 1,048,627 字节（64 MiB 源文件）时被终止，用例 2.26s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-backup-crash GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup -count=1 -timeout 120s
# PASS，6.524s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-backup-crash GOFLAGS=-buildvcs=false \
  go vet ./internal/backup
# PASS
```

这覆盖 Linux 上真实备份进程在临时归档写入期间被杀后的恢复，不模拟物理断电、设备写缓存丢失、Windows 文件语义或 Master/Daemon 的在线任务回执重建；这些范围仍需单独验收。
