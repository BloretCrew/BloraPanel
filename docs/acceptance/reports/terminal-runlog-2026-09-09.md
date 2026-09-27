# B05 native 持久日志捕获

日期：2026-09-09。范围：internal/runlog、cmd/log-helper；Daemon API、业务运行状态与浏览器集成由主代理另记。

后续已增加可选的独立 stdin 持有，详见 terminal-stdin-2026-09-09.md。下文 EOF 自动结束描述适用于默认 HoldInput=false；HoldInput=true 会等 runtime 确认 Finish。

## 接口与恢复合同

`runlog.Start(ctx, Options{Root, RunID, Command, Archive, StartupTimeout})` 在业务运行启动前创建独立辅助进程。`Capture.Stdout()` 同时传给 runtime.IO 的 Stdout/Stderr；runtime.Start 返回后立即 `ReleaseWriter()` 释放 Daemon 自己的写端副本。`Command` 仅由管理员/Daemon 配置，可为当前 Daemon 的隐藏 `--log-helper` 模式；`RunHelper(args)` 接收 Start 追加的 `--runlog-root PATH`。单独构建入口是 `make -f cmd/log-helper/Makefile`。

helper 持有业务 stdout 管道的真实读端，独占 terminal.Archive；它不依赖 Daemon 内部 io.Copy goroutine、管理连接或浏览器。Linux 创建独立 session、无 Pdeathsig，持久化 PID、/proc start ticks 和 boot ID；结束使用先取得再复核出生身份的 pidfd。Windows 创建 detached 独立进程组，必要时显式 Job breakaway，并在 helper 内确认不属于任何父 Job；禁止 breakaway 的环境明确启动失败。Windows 身份为进程创建时间及精确进程 handle。

`Open(root, runID)` 只验证并重连原 helper，不创建替代进程或 pipe。活跃期间，读取通过 127.0.0.1 临时端口、随机 256-bit token 的私有 HTTP IPC；token 只在 0600 记录和内部请求，不传给浏览器。最大 8 个并发读取、批次最大 1 MiB、HTTP 时限防止慢读者积压。业务输出直接进入有界落盘归档，不保留浏览器队列。

helper EOF 后完成同步、关闭归档、原子写完成记录，后续直接读取磁盘。活跃归档不允许第二个进程 OpenArchive，以免恢复扫描截断正在写的尾部。若原 helper 已确认死亡且无完整结束记录，可回看已保存历史，Status 为 invalid 并明确日志可能不完整；不假称能恢复实时捕获。

`Read(ctx, after, maxBytes)` 返回 terminal.Batch。after 是排他事件游标，独立于 WSS 字节 seq；输出保留 UTF-8 碎片、ANSI、NUL。轮转返回 Earliest/Latest/Next/Gap，不隐瞒缺口；输入不入归档。默认归档为 16 MiB / 1 MiB 段，可以收紧预算。

只有 runtime 已确认本次整个运行退出，才调用 `Finish(ctx)`。它先等真实 EOF、同步，最长 5 秒；写端仍被继承则仅结束出生身份已确认的日志 helper，返回 ErrIncomplete。Daemon shutdown 只 ReleaseWriter，不能对仍在运行的业务调用 Finish。存储失败时 helper 状态变为 degraded，继续丢弃并排空后续输出，避免主动关闭读端导致业务 SIGPIPE；不能把丢失的日志报告成功。

## 实证

统一环境：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。

`go test -race -v ./internal/runlog ./cmd/log-helper` 通过（最终 1.876 秒），需要经自动权限审查允许测试自己的 loopback listener 和 helper 子进程：

- 独立 launcher 模拟 Daemon，真实退出后业务 PID 1983009 继续写中文与 ANSI，完整回看 86 原始字节。测试主进程未继承该业务管道，也没有后台 goroutine 替 Daemon 读 stdout。
- 无浏览器写入 1,200,000 字节，64 KiB 预算实际保留 53,376 字节，事件 105..110；从 0 读取明确 Gap。
- 活跃 helper 重连、已完成归档恢复、未来游标错误、重复 run 拒绝、无 token 的 IPC 请求 401。
- 错误出生时间不终止 helper；泄漏写端使 Finish 在指定截止时间后结束本次 helper且报告不完整，历史仍可读并标 invalid。
- 主动将测试归档路径替换为普通文件以触发真实磁盘路径错误：业务继续写完整大块输出，不因捕获器关闭 pipe 而失败，日志状态为 degraded/failed。

`go vet ./internal/runlog ./internal/containerterm/... ./cmd/log-helper ./cmd/exec-helper` 通过。

`GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-runlog-windows.test.exe ./internal/runlog` 通过。Windows 真机可执行 `go test -race -v ./internal/runlog`，包含真实父进程退出后 helper 与业务继续工作的专属测试。本次没有 Windows 真机，运行标为未验证。

## 后续集成注意

Docker 既有 runtime.ContainerLogs 提供所有权校验后的 Engine 流，多路复用协议、秒级 since 与 tail 10000 不能当精确事件游标；Engine 已固定 local driver 10m×3。若转成公共日志 API，需要显式解帧、缺口和重连去重语义。

审查还发现既有 Windows native runtime 的 KILL_ON_JOB_CLOSE 直接由 Daemon 持有，Daemon 崩溃会停止业务。这不属于日志 helper 交编已满足的能力，已交接并继续实现独立 Job keeper。不得据本文声称 Windows 整体运行恢复已经验证。
