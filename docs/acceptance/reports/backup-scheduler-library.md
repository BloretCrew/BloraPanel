# B08 本地备份与平台调度模块验证

记录日期：2026-09-09。范围为 `internal/backup/**` 及必要的安全文件库原语；未修改根验收矩阵或把模块测试视为 B08 产品链路完成。

## 已实现的模块合同

`backup.New(administratorRepositoryRoot, Options{StateDir,...})` 只打开 Daemon 固定的私有仓库/状态根。实例文件来源由调用者借用 `filesFor` 返回的授权 `*filesystem.Service`；整项工作必须 `defer release()`，不能按浏览器主机路径新开实例根。通过实际根/对象关系拒绝实例来源或恢复目标与仓库、状态目录的包含/别名关系。

公开数据服务：`Create`、`Inspect`、`List`、`PlanRestore`、`GetRestorePlan`、`Restore`、`CreateStatus`、`RestoreStatus`、`CleanupCreate`、`Prune`。共享 `TaskPayload` 包含冻结的 `model.InstanceConfig` 及 Create/Restore/Retention 之一。接口类型与 JSON 字段以 `internal/backup/types.go` 为准；`SnapshotInfo`/`Result` 为有界概要，完整清单、计划由上层分页，不能塞进任务控制消息。

备份格式为 `blora-backup-v1`：单磁盘、非 ZIP64 的标准 ZIP，固定 `manifest.json` 与 `data/` 内容。清单包含相对路径、文件 SHA-256、大小、普通 mode、mtime、来源资源/版本与根/对象身份。支持 Store/Deflate。打包流每次最多 64 KiB，从文件服务读取时逐片核实版本/长度/hash；写入私有 pending 包，文件/检查点 sync 后才记录进度。最终复核来源，记录包整体 hash/长度/对象身份的 committing 意图，再 rename 发布。重启仅凭实际对象身份与包 hash 确认已完成提交，不把同内容的其他对象当作本次提交。受控 save/stop hook 可使应用原子替换来源文件；先验证原始基线与固定实例根，随后明确记录 hook 完成后的 CapturedVersion/SourceObjectID，并在整个打包期间固定该状态。

恢复先完整验证包整体 hash、实际对象、中央目录预算、嵌入清单一致性、每个成员 hash/长度/CRC，然后生成固定计划：包 hash/对象身份、目标根/对象、每项目标版本/对象和最终路径均参与 PlanHash。读取已有计划也重新校验 PlanHash，不能修改磁盘计划后保留旧确认。

已有文件覆盖必须显式提供逐路径 `Overwrite[path]=plannedVersion`，或提供 `OverwritePlanHash=PlanHash` 且映射为空。后者仅确认这一份不可变计划内全部已有文件，保持 4096 项恢复任务正文有界；HTTP 层必须来自用户明确确认，不能自动填充为全局 force。恢复通过 `Mkdir`、分片上传/提交及最后的目录元信息操作执行，保留已有目录的其他内容与元信息。每个分片/项前复核读取和目标写入权限；最终再次检查所有输出对象与文件 hash，以及原包 hash。

所有恢复副作用前持久化操作意图。中断后只观察当前上传的真实提交回执或确认清理本次 staging，保留已完成文件和部分计数；不能因重复同一 Restore 请求自动继续后续覆盖。未知 mkdir/metadata 回执明确保留 unknown。继续工作需要重新检查实际目录并创建新的确认计划。

保留策略限定 `OwnerID + Source resourceRef + Path`，支持 KeepLast 下限、MaxAge、MaxBytes；先记录 pruning 意图，确认实际对象/hash 后删除。保留正在被恢复/检查持有的包，不删除未登记文件。删除后保留 pruned 回执以避免同 ID 再次执行备份。默认保留 1000 个活动快照，元记录枚举也有 4000 条硬上限；达到元记录上限会背压，尚无无限历史元记录归档/迁移能力。

## 一致性与授权边界

`files` 表示所读取文件集合在前后版本检查中一致；不等同于运行中应用事务一致性。`save/pause/stop/hooks` 需要真实 `HookRunner`，未在 Daemon 配置 `backupHookCommands` 时返回 Unsupported。Daemon 现在提供可选的固定 argv 适配器：命令由管理员配置、无 shell 拼接，按稳定 hookID 先写入运行回执，再启动进程，并将结构化资源身份作为环境变量传入。Run 必须先持久接受稳定 hookID，Reconcile 只能观察原操作，不能重发未知副作用。

hook `succeeded` 仅证明所选择机制完成或命令已交付。尤其 stdin 写入成功不能证明应用已完成保存或建立应用一致性。`backupHookCommands` 的 key 可使用 `backup.<mode>.<phase>` 默认引用，或由请求明确选择已配置的 Before/After 引用；未知引用保持不可用。Daemon 没有通用运行时 pause 能力时，pause 仍需管理员提供专用命令。不得将 Before/After 字符串当任意宿主 shell。

失败或取消后，以原稳定 After hookID 在有截止时间的清理 context 中补偿。`CleanupCreate` 只核对既有提交、清理本次 pending、观察/补偿原 hook；不重新 capture，也不发布仍在 pending 的包。它可按原任务所有权在读取权限撤销后进行清理。原运行状态恢复必须由真实 adapter 证明，不能凭模块返回值猜测已经恢复。

`Options.AuthorizeSnapshotRead(ctx, actor, sourceRef)` 非 nil 时按真实来源权限支持跨 owner 恢复，调用者/审计 OwnerID 始终实际 actor；nil 默认仅原 owner 可读。它覆盖旧计划、重复回执、每项与每个实际分片；Create 读来源时同样调用。`AuthorizeRestoreWrite(ctx,actor,targetRef)` 供上层逐批复核目标 file.write + backup.restore。调用者还须绑定当前 Instance 配置/根和运行状态策略。Prune 继续仅处理指定 owner 范围，未加入任意用户清理他人备份的隐式权限。

## 已实现的持久调度器

`NewScheduler(store, SchedulerOptions{Authorize, BuildTask, NodeOnline})` 使用现有 SQLite Store 的专属记录命名空间，不引入第二份任务数据库。Master 负责启动单个 Tick 循环并在 Close 前取消、join；调度器本身不隐式创建后台 goroutine。`Save/Get/List/Delete/Tick/PruneFires` 可供注册层调用。

默认最多 256 个计划、4096 条触发记录；每次最多计算 4096 个历史时点，超过时标记 `MissedCountTruncated`，推进 NextRun，避免长期离线补跑风暴。标准五字段 cron + 显式 IANA 时区；内置时区数据，拒绝秒字段、@every 与表达式内另嵌时区。

离线/错过策略为 skip 或 run_once，后者最多保留一份已持久化待补时点；后续时点合并计入 Skipped。重叠策略为 skip 或 wait_one，按真实 ResourceRef + Action 检查所有用户的未终止任务，也在 BuildTask 之后重新检查。配置修改/停用与最终接受串行，准备阶段读到的旧配置不能再产生新任务。

先持久化 NextRun/选定 slot，再持久化冻结的待接受任务，最后使用 `schedule:<stableFireID>` 调用已有 Store.Accept。FireID 绑定 scheduleID、配置修订与 UTC slot。重启优先核对已接受任务的真实 Actor/RequestID/Action/Resource/Payload 回执，再处理新授权；节点离线或原 actor 已撤权不会导致已接受任务重复创建。新触发每次重新授权，BuildTask 必须只读并重新绑定实际实例配置。Tick 仅返回已被 Store 接受的 fire，prepared/discarded 记录不能派发。

PruneFires 仅删除本模块完成触发的元记录，或确认从未接受且已不再待触发的旧配置意图；不删除任务日志、不清除活动/未知任务。接受过的任务回执缺失时明确 Interrupted，不自动重放。

## 本次实际验证

环境：Linux amd64，内核 `6.19.12-200.fc43.x86_64`，Go 1.25.9，本地真实临时目录与 SQLite。未访问生产资源或输出凭据。

命令统一环境：

```sh
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test -race ./internal/backup ./internal/filesystem -count=1
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go vet ./internal/backup ./internal/filesystem
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false GOOS=windows GOARCH=amd64 go test -c ./internal/backup -o /tmp/blora-backup-windows.test.exe
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false GOOS=windows GOARCH=amd64 go test -c ./internal/filesystem -o /tmp/blora-filesystem-windows.test.exe
```

22 个 backup/scheduler 顶层用例与文件库 race 均通过。最近一次完整模块测试：backup `2.051s`、filesystem `1.304s`，退出 0（增加 `-timeout=180s`）。本轮 Vet 退出 0；先前两个 Windows amd64 测试二进制交叉构建退出 0，Windows 二进制未在 Windows 运行。

续接审查修复恢复预览的目标授权检查间隔：PlanRestore 现在每项读取目标之前、最终写入计划之前重新检查目标写入权限，并在发布计划前重新检查来源读取权限。新增真实目录用例在入口授权成功后撤销目标写入，确认返回拒绝且没有留下可提交计划。Inspect/List/GetRestorePlan/CreateStatus/RestoreStatus 为受信服务接口，没有 actor 参数；网络层必须自行过滤范围并重新授权，不能直接公开整个私有仓库。

覆盖真实 ZIP 的目录/大文件/空文件往返、mode/mtime、版本冲突、同 hash 对象替换、包替换、显式覆盖/现有目录合并、计划篡改、取消/读写撤权、来源在读取期间改变、独立上传不被清理、仓库/来源重叠、符号链接、遍历/Windows 替代路径/重复成员/炸弹大小/内容 hash、所有预算与保留读取 pin。

恢复/保留 rename 或删除后回执滞后的场景，先执行真实文件操作，再注入缺失最终回执的持久状态并实际关闭/重新打开根句柄，以验证核对与不重放。另有真正启动测试子进程，在 before hook 效果文件已 sync、最终 hook 回执未写入时 `os.Exit(73)`，父进程重开 Manager 后只 Reconcile 原 hook并补偿，没有再次运行未知 before hook、没有生成成功快照。失败任务的 After 回执缺失时，即使旧 CleanupPending 字段尚未置位，也根据实际 running 回执重新观察原补偿；测试确认没有重复执行。另有实际原子保存替换文件的 hook 测试，恢复得到保存后的内容与绑定版本。

调度测试使用真实 SQLite：夏令时春季不存在时刻、秋季两个不同 UTC 时刻、五字段限制、错过跳过/单份补跑、跨用户同资源重叠、准备期间出现手工任务、配置停用竞争、撤权、元记录清理。恢复测试实际 Close SQLite 然后 Open 同一文件，保留真实已接受任务并注入滞后的 schedule/fire 回执；只核对原任务，没有产生重复任务。这些模块测试没有声称真实 Daemon 已执行定时任务。

## 文件库伴随修复与 API 依据

上传新增可选 ExpectedObjectID，Begin/Commit 在路径锁内复核原目标对象。提交意图新增 CommitObjectID，记录本次 staging 对象；重启阶段即使目标内容 hash 一样，也不能用其他对象冒充本次 rename。旧 committing 记录缺 CommitObjectID 时返回冲突/未知，需要核查；已 committed 回执保持兼容。

Linux 普通元信息使用打开的对象句柄调用 `utimensat(AT_EMPTY_PATH)`，保持支持文件系统上的纳秒精度；Linux 5.8 以前不支持该参数时回退安全 fd-based Futimes，精度为微秒。Windows 保留 ReOpenFile/SetFileTime 的真实句柄实现，粒度由 Windows 文件系统决定，目录不宣称断电持久保证。API依据：[Linux man-pages utimensat](https://man7.org/linux/man-pages/man2/utimensat.2.html)。

cron 锁定 `github.com/robfig/cron/v3 v3.0.1`，已读取本地该版本 parser/spec 源码并核对官方合同：[v3.0.1 源码](https://github.com/robfig/cron/tree/v3.0.1)、[官方 Go API](https://pkg.go.dev/github.com/robfig/cron/v3)。

伴随上传身份强化后，额外执行原有真实 TLS 回归 `go test -race ./internal/master -run '^TestTransfer' -count=1 -v -timeout=180s`，11 项通过，`47.291s`。该命令使用同一 Go 环境，因沙箱限制回环监听而经自动审批运行；仅使用自己的临时 SQLite、目录和回环 TLS 节点。它证明原文件传输链路回归，不能代替尚未接线的备份 TLS 验证。

## Root 接管与端到端验证

Master/Daemon 已注册 backup/scheduler 路由、bulk RPC、冻结配置任务适配及可跟踪 Tick 生命周期；`TestBackupAPIRealArchiveConditionalRestoreAndRevocation` 真实双 Daemon 通过（3.563s），调度 API/触发策略集成通过（3.922s）。这些证据覆盖实际服务链路的归档、计划、恢复撤权和定时任务接受，但不能据此称 B08 端到端全部完成。

1. Daemon 固定私有 BackupRoot/StateDir、持有单个 Manager；整项借用 filesFor + release，使用独立 backupSlots，关闭前 join。将 TaskID/Actor/Resource 固定给 Create；Restore Owner 固定真实 actor，校验 PlanID/Hash/目标配置。
2. Master 默认注册备份列表/inspect/计划分页/create/restore/retention/cleanup 与 schedule API；逐资源权限与配置修订检查必须在 Store.Accept 事务和执行时成立。控制通道仅概要，完整清单/计划分页不能作为单帧正文。OverwritePlanHash 只可接用户明确确认。
3. 绑定真实 SnapshotRead、RestoreWrite 与可选的 Daemon `backupHookCommands` HookRunner 回调，hook 通过稳定回执完成；stop/save/配置能力引用由管理员显式配置。未配置 pause 能力时 UI/API 明确不支持，stdin 成功只报告交付。
4. Master 启动/关闭 Tick 及 Fire 元记录保留；BuildTask 只读当前节点、配置、版本，任务仍由既有 durable dispatcher 执行，不能直接调用操作绕过任务记录。
5. 完成真实 TLS Master↔Daemon 备份/恢复/调度/离线重启/撤权/取消链路及浏览器确认、进度、恢复现场验证，再更新根验收矩阵。

调度接收边界：模块在准备前后检查同资源同 Action 未完成任务，但外部手动接收不持有 Scheduler 的锁。公共 Store.Accept 还需对含 scheduleId 的任务在同一接收事务内检查重叠，避免最后一次查询与实际插入之间的竞争；模块测试中的“准备期间出现手工任务”只验证已覆盖的准备窗口。

尚未验证：Windows 真机、真实磁盘 ENOSPC/掉电、真实备份 Daemon 在目标提交途中进程崩溃、配置命令对具体应用 stop/save/pause 语义的真实演练、NFS/SMB/重挂载对象身份、GB 规模/4096 项性能与定时任务长期运行。没有增加默认远端云供应商。未用配额错误、交叉构建或故障状态注入替代这些环境验收。
