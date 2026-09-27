# Windows 独立 Job keeper 与恢复合同

日期：2026-09-09。实现范围：runtime.Options.JobKeeperCommand、native_windows.go、keeper_windows.go、非 Windows 入口及专属测试。主代理负责 Daemon/集成测试程序的隐藏入口。

## 问题和实现

旧代码只由 Daemon 持有 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE 的 Job，Daemon 退出会停止实例。现在保留该限制，由独立 keeper 在业务创建之前取得同一个命名 Job 的句柄；keeper 不依赖 Daemon 连接和进程。它查询真实 Job 进程计数，所有关联进程结束后才退出。尚未关联任何业务的失败启动有 10 秒上限；已关联业务没有管理连接超时。正常启动失败使用本次创建的 Job/keeper 精确句柄撤销并有界等待 keeper 退出，清理失败合并到错误。

`Options.JobKeeperCommand` 是管理员选择的可执行程序前缀，默认 `os.Executable()` 加 `--job-keeper`；实例配置不能指定。启动追加 `--state-root PATH --run-id ID`。隐藏入口调用 `runtime.RunJobKeeper(args)` 后立即退出，不能初始化 Daemon 服务。keeper 校验持久运行记录和环境内所有权标记，持久化其 PID、创建时间、Job 句柄值；令牌、句柄仅为内部字段。

keeper 使用 DETACHED_PROCESS、独立进程组，必要时从父 Job breakaway，并在内部确认不属于任何父 Job。业务也必须脱离 Daemon 外层 Job，再通过 JOB_LIST 原子关联自身受限 Job；若外层禁止 breakaway，启动明确失败。业务自身 Job 仍禁止 BREAKAWAY/SILENT_BREAKAWAY，不削弱原 CPU/内存/进程数限制。

恢复并非只凭 PID 或名字：OpenJobObject 重开持久名称，OpenProcess 取得 keeper 的精确进程句柄并核对创建时间，再 DuplicateHandle 取得 keeper 持有的 Job，通过 CompareObjectHandles 确认与重开的 Job 是同一内核对象。已有业务父进程仍在时还检查出生时间与 IsProcessInJob；父进程退出后以同一 Job 的成员和计数继续管理后代。身份不符返回 UNKNOWN，拒绝 Stop 信号。启动在 suspended 持久屏障附近中断时呈现 UNKNOWN，允许明确停止已确认 Job，不再次 ResumeThread 或创建运行。

## 官方依据

- 微软说明 Job 在最后句柄关闭且所有关联进程结束后销毁；KILL_ON_JOB_CLOSE 则在最后句柄关闭时结束关联进程。命名 Job 支持重开。这里以独立 keeper 保留句柄，而不是仅删关闭限制。[Job Objects](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects)
- OpenJobObject 按名称、命名空间和安全描述符返回已有对象句柄。[OpenJobObjectW](https://learn.microsoft.com/en-us/windows/win32/api/jobapi2/nf-jobapi2-openjobobjectw)
- DuplicateHandle 需要源进程的 PROCESS_DUP_HANDLE 权限，复制句柄指向同一内核对象；CompareObjectHandles 直接验证两个句柄是否指向同一对象，最低 Windows 10 / Server 2016。[DuplicateHandle](https://learn.microsoft.com/en-us/windows/win32/api/handleapi/nf-handleapi-duplicatehandle)、[CompareObjectHandles](https://learn.microsoft.com/en-us/windows/win32/api/handleapi/nf-handleapi-compareobjecthandles)
- 创建标志明确规定 breakaway 依赖父 Job 许可；业务及 keeper 无法脱离时不能声称能跨 Daemon 生存。[Process Creation Flags](https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags)

## 构建与验证状态

统一环境：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。

- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-runtime-keeper-windows.test.exe ./internal/runtime` 通过。
- `GOOS=windows GOARCH=amd64 go vet ./internal/runtime ./internal/runlog` 通过。
- Linux `go test -race ./internal/runtime ./internal/runlog ./internal/containerterm/... ./cmd/log-helper ./cmd/exec-helper` 回归通过：runtime 2.638 秒、runlog 1.899 秒、Adapter 1.018 秒、helper 3.138 秒。未设置真实 Docker测试变量，不把被跳过的 Docker 测试计为本轮实测。

本次没有 Windows 真机，以下仅已实现并交叉编译，**运行未验证**。真机命令：`go test -race -v ./internal/runtime -run '^TestWindows'`。

- 真正的 Daemon 子进程启动业务后 os.Exit；新 Manager 读取记录、重开和验证 Job，确认业务父进程结束后后代仍在；Stop 确认原 Job 退出，keeper 最终退出，另一独立控制进程保持存活。
- 丢失原 Daemon 管理句柄后运行保持、再次管理与 Stop。
- keeper 创建时间被改错后 Observe/Stop 拒绝，不对错误身份发信号。
- 在 keeper 已启动后，用不存在的业务工作目录触发 CreateProcess 失败，确认本次 keeper 及空 Job 消失。

这些 Windows 场景必须在符合 Windows 10 / Server 2016+ API 能力和服务账户访问权限的真机执行，不能把交编当作运行证据。实例 stdin 的独立持有和控制随后已实现，见 terminal-stdin-2026-09-09.md；本报告不代表 Windows 全链路完成。
