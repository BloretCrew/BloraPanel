# 扩展用户数据迁移复验（2026-09-20）

针对当前扩展注册表实现运行真实 WASI guest 的数据迁移、并发写入保护和恢复场景：

```text
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test -race ./internal/extensions -run 'TestUserDataIsolationCASMigrationAndRecovery|TestMigrationDoesNotBlockRegistryAndRejectsConcurrentWrite' -count=1
```

退出 0，耗时 85.404s。测试覆盖双用户数据隔离、升级迁移、迁移失败保留旧包、回滚迁移、事务中断恢复，以及迁移期间其他扩展和用户写入不被锁住、同一迁移不重复并发运行。

这证明后端数据迁移链路已实现并通过 Linux race/WASI guest 验证；包跨文件物理掉电、Windows 真机、远程长期部署和完整 E06/E07 组合仍未覆盖。
