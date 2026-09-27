# 备份与调度真实集成竞态复验

2026-09-19 Master重开增量：`TestScheduledBackupMasterReopenKeepsAcceptedOccurrence` 接受定时备份后关闭真实Master和SQLite，重新打开数据库、服务并让两个Daemon重连；原任务ID/requestId保持，任务SUCCEEDED、数据库只一项任务、备份目录只一份归档，同触发时刻再次Tick不提交任务。定向race退出0，3.198s。这是服务/数据库真实重开，不称SIGKILL或物理掉电验证。

上述2026-09-19新增用例的合并复验：`BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'Test(Backup|Schedules|TransferTargetENOSPC)' -count=1`，退出0，20.324s；包含真实私有挂载ENOSPC，未把默认跳过算通过。无残留测试服务。

2026-09-19 固定命令钩子全链路：`TestBackupFixedSaveHooksCapturePublishedStateAndRejectFailure` 在真实TLS Master/双Daemon内运行独立固定argv保存程序。before原子发布最新正文、after记录资源身份，生成归档后实际恢复得到最新正文；故意返回非零时任务FAILED且after补偿执行。同键回放原任务，两个钩子均只执行一次。race6.174s，两个子场景通过。此证据证明示例保存程序的实际save语义，不泛化为任意数据库/业务程序的pause/stop一致性；其operator仍必须配置适合该应用的命令。

2026-09-19 增量：`TestSchedulesRealNodeOfflineCoalescesAndRechecksPermission` 使用真实TLS双Daemon，实际关闭目标Daemon，受控时钟推进五个触发时刻；离线不新增任务，恢复后仅补跑一次并实际完成备份，使用恢复时最新源版本和原创建者身份。另一子场景在离线期间撤销创建者备份权限，节点重连后明确authorization_denied且账本无新任务。`go test -race ./internal/master -run '^TestSchedulesRealNodeOfflineCoalescesAndRechecksPermission$' -count=1 -v`，退出0，2.869s（两个子场景）。这是实际节点失联/重连与受控时间组合，不是五分钟或更长期墙钟运行，也不代替Master进程崩溃验收。Go环境沿用 `/tmp/blora-go-full-final54` 缓存。

Linux备份创建与恢复的真实ENOSPC现已独立通过，见[空间不足报告](enospc-2026-09-19.md)；历史段落中的该缺口由新证据补齐，物理掉电和Windows仍未验证。

记录日期：2026-09-14。验证只使用本地临时目录、SQLite 和回环 TLS Master/Daemon，不接触生产资源。

本轮在提权测试环境执行：

```sh
GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-sched2 GOFLAGS=-buildvcs=false \
  go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1
```

结果：退出码 0；`internal/backup` 1.856s，`internal/master` 5.816s。覆盖真实 Master API 的备份归档、条件恢复、撤权、取消、定时计划创建/删除/触发，以及调度时区、错过执行、重叠、节点离线、权限变化和 SQLite 持久回执场景。该命令是在受限沙箱 IPv6 回环监听失败后，以隔离提权方式重跑；未修改服务配置。

固定 argv 一致性 Hook 适配器另执行：

```sh
GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-hooks GOFLAGS=-buildvcs=false \
  go test -race ./internal/daemon -run 'TestBackupHook' -count=1
```

结果：退出码 0，2.039s。测试确认 hook 命令参数不经 shell 拼接，环境中的资源身份正确绑定；重复稳定 hook ID 返回持久回执；`running`/未知回执在重启后只转为 `unknown`，不会再次执行可能已产生副作用的命令。具体应用的 save/pause/stop 语义、物理 ENOSPC/掉电和长期调度仍按矩阵保留未验证。

随后补齐计划配置写入的丢响应边界：更新和删除分别通过 `SaveMutation`/`DeleteMutation` 在 SQLite 事务中保存请求回执；同键重试保持原始结果，不同载荷返回 `IDEMPOTENCY_CONFLICT`。调度定向竞态退出 0；包含该改动的提权串行全仓普通回归退出 0（Master 128.030s），竞态回归退出 0（Master 267.872s、extensions 149.546s）。
