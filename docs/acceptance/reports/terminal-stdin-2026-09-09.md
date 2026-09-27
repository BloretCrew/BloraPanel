# B05 独立 stdin 持有与单次输入

日期：2026-09-09。范围是 runlog 的 HoldInput 扩展及 runtime.NativeInput 回调；继承 terminal-runlog-2026-09-09.md 的 stdout 归档、身份、预算和缺口语义。Daemon/console.input 的资源权限与任务去重由主代理集成验证。

## 行为与接口

`Start(Options{HoldInput:true})` 令独立 helper 创建并持有业务 stdin 的真实写端。Start 校验 helper 的出生身份并准备只读副本，`Capture.Stdin()` 传给 runtime.IO.Stdin；业务启动后 `ReleaseInputReader()` 关闭 Daemon 本地只读副本。Daemon 从不持有唯一写端，所以 Daemon 退出不会令业务 stdin 意外 EOF。默认 HoldInput=false 保持既有纯输出捕获行为。

Linux 读取副本通过 `/proc/helperPID/fd/readFD` 打开，核对 helper 前后出生身份、boot ID 和真实 FIFO device/inode。Windows 从出生时间匹配的精确 helper 进程句柄 DuplicateHandle，只接受管道只读句柄。Open 历史捕获不复制 stdin，因此已结束的捕获器仍可读日志。

`Capture.WriteInput(ctx, data)` 向私有 loopback IPC 发一次认证 POST，单次最大 32 KiB、写等待最多 2 秒，返回实际接受字节数。Linux 仅在 EAGAIN/EINTR 未接受字节时等待；发生部分写入立即返回，不补写。Windows 使用本机、拒绝远程客户端的 overlapped named pipe，取消精确 I/O 并确认完成；若取消结果仍不明，保持关联内存至内核结束，禁用后续输入并返回不完整，避免重用仍在工作的缓冲区。任何超时、部分写入或响应丢失都不能自动重放。

stdout EOF 仅进入 output_complete/output_failed；只要还持有 stdin，helper 和输入接口继续存活。只有 runtime 已确认整次运行退出后，`Finish` 的认证内部请求才关闭 stdin 并允许 helper 完成。`Status.InputAvailable` 仅在活跃 HoldInput 且输入通道可用时为 true；旧 stdout-only、完成、失效及历史归档为 false。输入正文不进入输出归档。

`runtime.Options.NativeInput func(context.Context, Record, []byte)(int,error)` 是管理员注入的可信回调。WriteInput 先观察并确认正确 run 正在运行，再调用该回调；未知或已退出的 run 不发送。nil 保留已有直接 stdin 测试/兼容入口，Daemon 成品应配置 Capture 回调。StopInput 同样经此通路。

## 真实验证

统一环境：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。测试进程和私有 loopback listener 经自动权限审查；未操作生产进程。

`go test -race -v ./internal/runlog ./internal/runtime ./internal/containerterm/... ./cmd/log-helper ./cmd/exec-helper` 通过。runlog 2.206 秒、runtime 2.641 秒、Adapter 1.013 秒；容器 helper 命中已通过缓存。真实 Docker/委派 cgroup 未配置的 opt-in 项明确跳过，本轮不算新增 Docker实测。

- 测试 Daemon 是真实子进程，业务关闭 stdout/stderr 后 Daemon 退出；业务仍等待 stdin，helper 状态 output_complete/inputAvailable=true。重新 Open 后发送一次中文命令，业务实际收到并写入结果文件，真实确认业务退出后 Finish 成功。
- 输入未出现在归档事件中；Open 未增加输入读端；结束后 inputAvailable=false。
- 交付给业务的 stdin 是只读副本，写它失败；超过 32 KiB 请求被拒绝。
- 无读取者时填满实际 OS 管道，用 250 ms 调用截止时间验证背压返回有界错误；没有自动补写/重放。测试结束仅收尾本测试 helper。
- stdout 原有完整回归仍通过：Daemon 退出后的中文 ANSI、1.2 MB 输出有界轮转、存储失败继续排空、身份错误拒绝和失效历史读取。

`go vet ./internal/runlog ./internal/runtime ./internal/containerterm/... ./cmd/log-helper ./cmd/exec-helper` 通过。

Windows 构建/静态检查通过：`GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-runlog-input-windows.test.exe ./internal/runlog`、runtime keeper/input 交编及 `GOOS=windows GOARCH=amd64 go vet ./internal/runlog ./internal/runtime`。同一 stdin 生存与背压测试支持 Windows，真机入口 `go test -race -v ./internal/runlog`；本次没有 Windows 真机，运行未验证。

Windows I/O 依据：连接前已打开客户端时 ERROR_PIPE_CONNECTED 表示有效连接；overlapped 使用独立事件。[ConnectNamedPipe](https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-connectnamedpipe) 取消请求本身不等于完成，也不允许立即回收 OVERLAPPED；代码保留内存直到确认内核完成。[CancelIoEx](https://learn.microsoft.com/en-us/windows/win32/api/ioapiset/nf-ioapiset-cancelioex)
