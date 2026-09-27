# 监控增量（2026-09-12）

关联 F02/F10、E01/E09，整项继续进行中。

Windows 新增原生 CPU、物理内存、磁盘、网卡累计字节、进程工作集与 CPU 采集；Toolhelp32 枚举映像、SID、出生时间。访问受限字段标记 unavailable，缺少身份时禁止终止；出生令牌为 Unix 微秒，保持 JavaScript 整数精度。终止在同一系统句柄上核对身份并等待最多10秒确认退出。

实现核对了 [GetSystemTimes](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getsystemtimes)、[GetProcessMemoryInfo](https://learn.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-getprocessmemoryinfo)、[GetIfTable2Ex](https://learn.microsoft.com/en-us/windows/win32/api/netioapi/nf-netioapi-getiftable2ex) 官方合同。CPU 首次采样不可用，多处理器组主机的节点 CPU 尚未实现聚合，明确返回不可用。网络是非回环接口累计和，虚拟接口可能重复计入流量。

Linux 先获取 pidfd、再核对出生标识，通过该 fd 发送 SIGKILL 并等待退出；拒绝 PID1 与自身。无 pidfd 支持时明确失败，不使用存在 PID 复用竞态的旧路径。内核合同见 [pidfd_open](https://man7.org/linux/man-pages/man2/pidfd_open.2.html) 和 [pidfd_send_signal](https://man7.org/linux/man-pages/man2/pidfd_send_signal.2.html)。

进程列表新增 limit、afterPid、search，搜索完整枚举，响应最多100个 PID，提供 nextAfterPid。各页为实时观察，刷新首批可发现新建低 PID 进程。界面保存节点、搜索、排序和游标；内存排序明确限本页。确认框捕获节点身份，切换节点不改变目标。实例内存改读 rssBytes，CPU零值与不可用分开显示。

验证命令的 Go 环境为 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false`：

- `go test -race ./internal/monitor ./internal/master ./internal/daemon -run 'TestProcess|TestTerminate|TestNodeHistory|TestMonitoring|TestHost' -count=1` 退出0：monitor1.291s、Master2.321s。Daemon没有匹配测试，不算覆盖。使用自建 sleep 验证分页、错误身份拒绝、退出确认，已清理。
- 最终 `go test -race ./internal/master -run TestHostProcessTaskDispatchAndIdentity -count=1` 退出0，2.246s；真实 TLS 搜索/授权/任务去重与终止通过。测试首次错误解码 helper 返回类型导致编译失败，修正后复验通过。
- `GOCACHE=/tmp/blora-go-grant-idem make check windows build` 退出0；Linux/Windows Master与Daemon产物在dist。
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-monitor-windows.test.exe ./internal/monitor` 退出0。Windows 真机入口 `go test ./internal/monitor -run TestWindowsNativeMetricsAndProcessIdentity -count=1 -v` 仅终止测试自身子进程，尚未运行。
- `cd web && npm run check`、`npm run build` 退出0；构建仍有Monaco大块提示。
- `npm run test:e2e -- tests/browser/monitor.spec.ts --workers=1 --reporter=line` 最终1/1通过，5.4s。Chromium API路由替身覆盖确认目标、节点/搜索/排序/分页立即刷新恢复，不能代替真实双节点浏览器全链路。

仍未覆盖 Windows 真机、跨处理器组CPU、超大进程数性能和完整故障负载。本轮没有重跑全仓race。
