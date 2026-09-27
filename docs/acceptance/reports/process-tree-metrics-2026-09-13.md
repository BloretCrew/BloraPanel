# 原生运行进程树指标

F10/E01增量，完整平台和负载验收未完成。

Daemon使用运行适配器确认的成员集合，逐个按PID、出生时间及RunID采样，合计CPU和RSS。原生结果增加scope=process-tree和processCount；界面说明RSS为合计，共享页可能重复计算，不等同去重物理占用。缺少任一成员的CPU基线或内存数据时，对应总量标不可用，不展示部分成员之和为完整结果。

采样后再次观察运行成员，拒绝重复PID、空出生身份、成员数/身份变化或已退出运行；最多三次重新观察和采样，受原请求上下文约束。一次最多1024成员，内存累计不超过JavaScript精确整数范围。进程基线表达到上限后，已有PID采样不会无故清空全部基线。

Windows Job观察列表现在逐项打开句柄，核对仍属于原Job并记录出生时间，供指标采样二次核验；不是仅凭Job枚举时的PID进行采样。Windows实际运行仍未验证。

验证：

- 真实TLS Master/Daemon实例运行父shell及两个子进程，其中子进程持续消耗CPU；TestMetricsAggregateOwnedDescendants要求scope、至少三成员、正RSS和正CPU，定向race2.380s通过。
- TestNativeMetricTreeRejectsChangedMembershipAndRetries使用自有FIFO控制真实运行新增子进程，旧两成员快照被拒绝，再采样获得三成员结果；测试最后停止并清理全部运行资源。
- 最终`go test -race ./internal/daemon ./internal/monitor ./internal/master -run 'Test(NativeMetricTree|OwnedProcessCPU|Metrics|InstanceMetric)' -count=1`：1.324s/1.270s/5.963s退出0，包含真实CPU、聚合、成员变化及断开/重连缓存回归。
- 最终check/build/windows及前端构建退出0。界面的计数/RSS说明尚未单独新增浏览器断言。

Go使用既有`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false`。无活动fixture或测试进程。

边界：此结果聚合观测时仍存在的运行成员，不是累计已退出子进程的历史CPU账本；短命成员持续变动时可返回不可用/旧采样，不能假称完整无损采样。原生网络和磁盘归属仍标不可用。下一步是E08真实八窗口、两终端、万条目录与传输并行负载，及Windows Job真机验证。
