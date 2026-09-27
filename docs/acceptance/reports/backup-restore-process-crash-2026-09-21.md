# 备份恢复上传中进程强制终止（2026-09-21）

关联 E03/F12。`internal/backup/restore_crash_test.go` 用真实备份仓库、文件服务和 `Manager.Restore` 建立 8 MiB 目录快照。独立测试子进程将归档文件写入目标上传暂存区时，父进程观察到 1,048,576 字节已暂存后立即强制终止它；该字节数仍小于完整文件，目标文件尚未发布。

父进程使用原备份状态、目标状态和仓库重新打开服务，并用原 Restore ID 对账。系统返回 `interrupted`，已完成目录保留为明确的部分进度（`Completed=1`、`Partial=true`），上传状态为 `cancelled` 且偏移大于零；暂存文件已清除，目标文件不存在，结果既非 `unknown` 也无待清理状态。同一 ID 重放仍保持中断结果。重新生成恢复计划后，以新 ID 恢复成功，8 MiB 正文长度及 SHA-256 与归档清单一致。

验证命令与结果：

```sh
env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-build \
  go test -race -run '^TestBackupRestoreProcessKillDuringUpload$' -count=1 -v -timeout 90s ./internal/backup
# PASS；1,048,576 字节暂存时杀死恢复子进程，测试 3.301s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-build \
  go test -race -run '^TestBackupRestoreProcessKillDuringUpload$' -count=5 -timeout 2m ./internal/backup
# PASS，连续5次，12.408s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-build \
  go test -race -count=1 -timeout 5m ./internal/backup
# PASS，6.864s

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-build \
  go vet ./internal/backup
# PASS

env GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-build \
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go test -c -o /tmp/blora-backup-restore-crash.test.exe ./internal/backup
# PASS；仅交叉编译，不代表 Windows 运行验收
```

这是 Linux 进程强制终止测试，操作系统和文件系统仍在运行；没有模拟断电、设备写缓存丢失、Windows 运行或恢复期间的 Master/Daemon 整体崩溃，E03/F12 仍保持进行中。
