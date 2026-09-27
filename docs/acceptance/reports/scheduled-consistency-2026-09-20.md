# 定时备份一致性验证（2026-09-20）

Master 的 `backup.create` 调度入口此前只接受 `files`，导致已经由 Daemon HookRunner 支持的应用保存、暂停、停止和自定义钩子策略不能进入定时任务。本次将调度校验对齐到备份执行合同：`files` 禁止钩子引用，其他四种策略允许可验证的固定引用（`[A-Za-z0-9_.:-]{1,128}`），命令仍只来自 Daemon 管理员配置，不由请求拼接。

验证命令：

```text
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test ./internal/master -run 'TestScheduledBackupSaveHookCapturesPublishedState|TestBackupFixedSaveHooksCapturePublishedStateAndRejectFailure|TestSchedules' -count=1
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test -race ./internal/master -run 'TestScheduledBackupSaveHookCapturesPublishedState|TestSchedules' -count=1
```

两条命令均退出 0；第二条耗时 7.625s。真实测试 Daemon 由固定 argv 保存程序在 `before` 原子发布新正文，在 `after` 写入回执；定时触发后检查任务成功、`save` 策略保留在快照、资源身份绑定、`before|resourceId` 与 `after|resourceId` 各一次，以及归档内容为保存后的正文。

默认备份应用同步提供定时策略选择和钩子引用编辑，创建/编辑时会回填并发送 `mode/before/after`。`cd web && npm run check && npm test -- --run` 的类型检查与 47/47 单测通过；`cd web && npm run test:e2e -- tests/browser/backups.spec.ts --workers=1` 退出 0，备份、恢复计划和定时任务保存 UI 1/1（5.0s）通过，并核对 `save` 策略及前后引用进入请求。

Windows 真机、独立 systemd 和物理掉电仍不由本机测试覆盖。
