# 系统适配增量（2026-09-12）

关联 F13/E05/E09；整项继续进行中。

2026-09-19 Linux单元类型边界修复：`ServiceAction`仅接受完整`.service`单元名，拒绝`.target`、`.mount`、`.socket`及隐式无后缀名称；`TaskAction`要求不含路径字符的`.timer`名称，避免将外部单元文件路径传给`systemctl enable`。Windows通用名称校验不变。界面来自systemctl列表的完整单元名仍有效。

验证：`go test -race ./internal/systeminfo -count=1`退出0（1.147s）；真实TLS管理权限用例1.759s通过；新增`TestSystemUnitTasksRejectPathsAndWrongTypesAtDaemon`覆盖Master接受后的Daemon失败回执，2.688s通过。测试将PATH设为私有空目录，断言错误来自类型校验，既不调用宿主服务管理器，也不把命令不存在算作正确拒绝。首轮纯适配器测试曾误认为带前导连字符的服务名不合法，删除该错误假设；生产命令原有`--`参数分隔保持。此项不证明真实systemd服务/计划任务生命周期，新的生产边界修复尚未纳入rc2发行包。

Windows 服务查询替换本地化 sc.exe 行解析，使用原生SCM枚举服务名并QueryServiceStatusEx读取状态，名称按排序返回。SCM只请求CONNECT与ENUMERATE_SERVICE，服务句柄只请求QUERY_STATUS；每个句柄关闭，访问受限项显示unavailable，不把每行文本当作独立服务。实现核对锁定x/sys v0.44.0源码及 [OpenServiceW官方合同](https://learn.microsoft.com/en-us/windows/win32/api/winsvc/nf-winsvc-openservicew)。现有每次最多200项限制保留，完整分页仍待补。

Linux计划任务修复list-timers的身份解析：倒数第二列UNIT作为.timer目标，最后一列ACTIVATES服务不再被误当定时器。可变长度时间字段保留为计划观察；任务启停只允许.timer并在systemctl参数加入--，防止把服务或选项当计划任务操作。

验证：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/systeminfo -count=1`退出0，覆盖定时器身份、未调度行、拒绝服务目标及原模块回归。`GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-systeminfo-windows.test.exe ./internal/systeminfo`退出0。make check/build/windows退出0。

新增Windows只读真机入口：`go test ./internal/systeminfo -run 'TestWindowsNativeServiceEnumeration|TestWindowsTaskSchedulerEnumeration' -count=1 -v`，要求确有非空服务及计划任务，不启动/停止/安装服务或计划任务。此环境没有Windows真机，入口尚未运行，交叉编译不能算运行验证。

Windows计划任务续接：固定PowerShell脚本使用Task Scheduler COM递归GetTasks/GetFolders，读取完整路径、数值状态和下次运行时间，输出UTF-8 JSON。完整任务路径支持Unicode和空格，拒绝空段/控制字符/路径穿越；名称通过环境作为数据传入，不拼接脚本。启停只设置Enabled并重新读取确认，不执行任务。依据 [GetTasks](https://learn.microsoft.com/en-us/windows/win32/taskschd/taskfolder-gettasks) 与 [Enabled](https://learn.microsoft.com/en-us/windows/win32/taskschd/registeredtask-enabled) 官方合同。子进程20秒截止，输出与诊断各1MiB，文件夹访问及队列上限10000；访问拒绝明确失败。依赖探测改为实际使用的powershell.exe，Linux计划任务改探测systemctl。

最终systeminfo全包race退出0（1.200s），覆盖Unicode/空格身份、非法路径及输出限额；Windows测试程序交叉编译退出0，make check/build/windows通过。PowerShell/COM脚本在本机无法实际执行，计划任务查询和Enabled变更真机结果仍未验证。完整分页、具体应用计划内容及其余E05场景仍未完成。

Windows服务操作续接：删除sc.exe启停/立即重启路径，改SCM原生Query/Control/Start。重启先等待STOPPED再启动，启动等待RUNNING，运行前退出返回失败及退出码；整体30秒截止，发送控制前检查取消。服务句柄按动作只请求QUERY_STATUS与所需START/STOP权限。服务名支持Unicode/空格并限制UTF-16长度，拒绝路径分隔符和控制字符；Linux后端继续严格单位名验证并以--分隔参数。原生SCM能力不再依赖sc.exe是否存在。

本轮systeminfo全包race、Go vet与Linux/Windows构建通过，服务名边界测试覆盖Unicode/空格/非法字符。真实Windows服务启动、停止、挂起状态及超时组合尚未运行，不能将交叉编译计为完成确认的真机证据。

服务分页增量：ListServicePage、Daemon RPC和Master GET接入after名称游标/nextAfter，响应最多200项，列表不再只能查看首批。Linux完整解析后按名称排序，Windows在SCM枚举名称排序后选择游标之后的服务再查询状态。各页是实时查询，回到首批可发现新增的较早名称。SystemApp新增首批/下一批，游标随视图立即持久化，切换节点复位。OpenAPI已记录合同。

验证：systeminfo全包race1.187s通过，临时systemctl子进程返回乱序服务，验证首/末页、无遗漏及非法游标；Chromium替身浏览器服务翻页/立即刷新/返回首批与原防火墙组合1/1（6.2s）通过。make check/build/windows、前端check通过；生产构建见PROGRESS。计划任务分页、Windows真机分页及大规模列表负载仍未验证或未完成。

计划任务分页续接：ListScheduledTaskPage与Daemon/Master传递limit/after/nextAfter，SystemApp提供首批/下一批并恢复游标，切换节点复位。Linux解析完整定时器输出后按单位名排序，Windows按Ordinal完整路径遍历文件夹，保留当前页所需最小路径集合；保留20秒与10000文件夹上限，超过上限明确失败，不返回伪完整页。每页最多200项，实时页不提供跨请求快照一致性。

新增205个乱序定时器经真实临时systemctl子进程返回，三页无遗漏/重复的race1.036s通过；Master分页参数TLS验证1.963s通过；Chromium替身浏览器计划任务翻页/立即刷新/返回首批与原系统组合1/1（6.6s）通过。前端check与make check/build/windows通过，最终模块全包race/前端构建见PROGRESS。Windows COM分页真机及大规模负载仍未验证。

2026-09-19 Windows Job Object CPU 配额修复：`native_windows.go` 原先在计算多核归一化 `CPURate` 后又用未除以逻辑 CPU 数的值覆盖，导致多核主机上的原生实例可能获得高于配置的 CPU 配额。新增 `windowsCPUQuotaRate`，统一按 `quota/(10*logicalCPUs)` 换算并保留容量上限；一核满配额、八核半配额、小配额及超容量均有单测。

验证：`go test ./internal/runtime -run TestWindowsCPUQuotaRate -count=1`、提权 `go test -race ./internal/runtime` 均退出0；`make check`、`make windows` 以及 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ./internal/runtime` 均退出0。Windows 真机执行仍缺环境，不能由交叉编译替代。
