# 容器模式实例指标

关联F10/E01，整项仍进行中。

修复Daemon对Docker实例只返回`container-cgroup-metrics`不可用的代码缺口。运行适配器沿用已有Docker API 1.45客户端，调用`stats?stream=false`取得两次CPU采样；调用前后核对完整容器ID、运行/实例/归属标签及StartedAt，拒绝采样期间重启、错误ID或早于本次启动的采样。10秒上下文限制采样等待。

CPU取容器和系统CPU时间差、在线CPU数（旧接口回退percpu长度）；缺失、倒退或早于启动的前序采样标为不可用。内存返回Engine原始usage/limit，含文件缓存，不伪称RSS。网络计数聚合接口，溢出或超过JavaScript精确整数范围不输出错误累计值。磁盘容量及RSS未从这些字段推断，明确列入unavailable。对应版本见[Docker API 1.45](https://docs.docker.com/reference/api/engine/version/v1.45/)，统计参数说明亦见[官方API参考](https://docs.docker.com/reference/api/engine/version/v1.46/)的ContainerStats。

现有Master实例指标授权、RunID核对、120点持久历史与失联旧数据标记继续使用。监控窗口优先显示memoryUsedBytes，原生进程保留RSS回退；容器显示文件缓存口径和内存上限。

证据：

- 本地HTTP协议测试覆盖统计解析、CPU计数/旧接口回退、容器ID不符、采样期间重启、过早采样、缺失/倒退计数和网络精度上限。`go test -race ./internal/runtime -run 'Test(ContainerMetrics|DockerMetrics)' -count=1`最终1.040s退出0。这是HTTP替身测试，不算真实Engine。
- 显式真实隔离容器：`BLORA_TEST_DOCKER_ENDPOINT=unix:///var/run/docker.sock BLORA_TEST_DOCKER_IMAGE=blora-isolated-e2e:20260909 go test -race ./internal/runtime -run '^TestDockerRealIsolatedLifecycle$' -count=1`，2.747s退出0；实际CPU双采样、正内存用量/上限和RunID成功，随后停止并清理自有容器。此后新增前序采样早于启动保护和计数溢出测试，普通有效路径未变。
- 最终`GOCACHE=/tmp/blora-go-grant-idem make check build windows`退出0；前端`npm run build`退出0。

Go命令共用`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false`。

后续真实链路增量：实例窗口“监控”标签原来只显示实例JSON，现嵌入MonitorApp；实例上下文只请求实例指标，不发起节点指标/历史/主机进程查询，也不呈现相同query key下的主机缓存。实例采样错误时不继续显示缓存为实时；服务端返回stale时说明“当前采样不可用”，停止实例也可能产生旧值，不能直接称节点失联。

全新节点首次容器启动暴露InstanceRoot目录未创建的产品缺口。运行管理器New现在初始化已配置的实例根目录，保持既有目录权限；真实运行测试改用尚不存在的嵌套根目录。该真实Docker测试2.950s退出0，最终runtime/daemon全包race2.982s/30.584s通过；条件Docker入口不混入全包通过声明。check/build/windows和最终前端构建通过，前端41/41通过。

`web/tests/real/container-metrics.spec.ts`在实际HTTPS Master/双Daemon、Docker fixture-3062496594下1/1（15.4s）通过：创建非root且128MiB限制容器，实际CPU/内存/API RunID、未授权403、授予instance.read后200、节点指标仍403、实例监控标签和刷新、无主机监控请求、停止后旧采样及原RunID、撤权后403。测试最后kill并清理运行资源，fixture已Ctrl+C停止。命令为`BLORA_E2E_CREDENTIALS=<私有fixture文件> npm run test:e2e:real -- tests/real/container-metrics.spec.ts --workers=1 --reporter=line`。

失败记录保留：先前fixture-1599939518三次创建测试失败来自测试使用isolated而非container模式、未带非root UID/GID；补齐API合同后启动失败，实际任务错误为新节点instances目录不存在，按上述修复后新fixture通过。未修改生产权限或校验规则。

缓存代次后续修复：Master在采样返回后重读实例并再次授权；节点/RunID在采样期间变化或节点返回其他RunID均返回409 RUN_CHANGED。失败回退只选择当前非空RunID的历史，不因其他代次时间戳更晚而混用。旧值保留原observedAt并标stale，不重写成当前时间。

真实TLS节点断开/重连追加测试`TestInstanceMetricDisconnectAndRunReplacement`：实际原生实例运行并采样，关闭测试Daemon并等待控制/批量链路均消失；同代次返回旧时间戳，断链期间撤销instance.read后403；重新启动Daemon后同代次恢复实时采样；实际重启实例生成新RunID，未采样时再断开返回503，不拿旧代次补位。清理会重连并kill自有实例。最终监控定向race5.043s通过，包含实时节点/进程、120点保留、另一代次较晚时间戳及未采样代次边界；check/build/windows退出0，OpenAPI3.1解析103路径。

此断链证据使用真实Daemon.Close和重新启动，非物理网络故障，也未强制注入采样处理中恰好更换代次/撤权的时序。Windows宿主连接Linux Engine、远程Engine和长时负载仍未由本次证明。Windows原生指标仍按其既有真机缺口记录。
