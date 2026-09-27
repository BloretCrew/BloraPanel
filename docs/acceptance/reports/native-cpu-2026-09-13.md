# 原生进程 CPU 与 fixture 清理

F10/E01及测试基础设施增量，整项未完成。

Linux进程采样读取`/proc/PID/stat`用户态/内核态CPU时间，与两次主机累计CPU时间差比较，得到主机总CPU容量占比。首次采样、CPU时间倒退、系统计数无增量或运行代次改变时标为不可用；基线按PID、出生时间及RunID核对，最多1024项。采样前后核对出生时间，Daemon传入运行记录的StartTicks，避免将复用PID的进程数据归入原实例。Windows保留原有进程句柄采样，补相同预期出生身份及RunID基线检查，尚未真机验证。

主机累计时间不再重复加guest/guest_nice（已包含在user/nice中），并补短记录和计数溢出拒绝。原生CPU使用host-total口径，Docker统计使用one-core口径，API新增cpuBasis且界面标明基准。字段依据[Linux进程统计手册](https://man7.org/linux/man-pages/man5/proc_pid_stat.5.html)。当前原生指标仍是主进程，整个运行所属进程树的聚合需要继续核对实现，不能将其当作所有后代的总消耗。

验证：

- `go test -race ./internal/monitor ./internal/master -run 'Test(OwnedProcessCPU|MetricsUseLiveNode|InstanceMetric)' -count=1`：monitor1.269s、Master4.667s退出0。测试在自有Go进程内实际消耗250ms CPU，要求第二次采样非零；错误出生身份拒绝、RunID改变重置基线，以及既有真实节点指标/断链缓存场景通过。
- 最终check/build/windows和前端生产构建退出0。最终cpuBasis为随后补充的展示元数据，构建已覆盖；Windows运行和完整压力场景没有据此计通过。
- 直接编译并运行`dist/blora-devfixture`（fixture-404818613），真实API启动自有原生实例后发送Ctrl+C。监督进程退出0；其Master SQLite显示一次start SUCCEEDED、两次kill SUCCEEDED、两实例均STOPPED。与旧行为仅留start不同，Linux独立子进程会话使实例清理先于服务退出。无活动fixture。`make fixture`改为编译并直接执行监督二进制，避免go run中间包装。

测试启动脚本位于本次临时检查文件`/tmp/blora_fixture_signal_test.py`，只从私有文件读取账号并通过HTTPS实际API启动自有实例，不输出凭据。Go命令使用既有/tmp模块缓存与GOFLAGS=-buildvcs=false。

下一步：原生运行进程树的CPU/RSS聚合范围、采样过程中的成员变化，以及Windows真机和完整E08负载。第二账号界面登录等待问题仍按此前报告保留。
