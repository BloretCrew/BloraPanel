# 私有网络中的真实防火墙应用与失联回滚

确认与节点重启增量：同一隔离场景在失联回滚后重新应用规则，普通账号确认403；管理员显式确认后任务SUCCEEDED/confirmed。实际重开该Daemon，runtime/permanent规则均保留；同请求键确认200，另一键409，原终态任务修订号不变。最后恢复测试前快照。完整增强场景race退出0，32.103s（私有子场景30.04s）。本次仅新增验收断言，未修改生产代码或重新打包rc2。

范围：F13/E05 Linux firewalld受控端口子集。没有访问宿主D-Bus或改变宿主防火墙。

实际软件：firewalld 2.3.2、nftables 1.1.3、dbus-broker 37（以本机RPM包元数据核对）。

`internal/master/firewall_namespace_linux_test.go` 显式启用后启动独立user/mount/net namespace子进程；先核对父子网络和挂载命名空间不同，将挂载设为private，再用私有tmpfs覆盖子进程的`/run`以隐藏宿主总线。D-Bus broker、日志接收端、firewalld配置、TLS Master及两个Daemon全部属于测试。退出关闭所属进程、私有socket和挂载；网络命名空间销毁清除所有规则。

```sh
GOCACHE=/tmp/blora-go-full-final54 GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go \
GOFLAGS=-buildvcs=false BLORA_TEST_FIREWALL_NAMESPACE=1 \
go test -race ./internal/master -run '^TestPrivateFirewalldApplyAndRestore$' -count=1 -v
```

结果：退出0，25.998s（子场景23.94s）。明确覆盖：

- 真实firewall-cmd配置runtime与permanent不同的已有端口和端口范围。
- 产品适配器捕获快照、应用3333/tcp，并从内核nftables规则读取实际端口；既有服务配置不变。
- 回滚精确保留两套原配置和未托管端口范围。
- 真实Master/Daemon预览与租约执行；普通用户预览403。
- 应用后在私有nftables独立表中丢弃发往测试Master端口的流量，TCP探测确认连接不可达。
- 不发送确认；在管理连接不可达期间直接核对firewalld已按期限恢复原配置。解除阻断后原任务为FAILED，阶段`confirmation_expired_rolled_back`。

独立适配器初步验证8.712s通过。环境搭建首两次失败分别来自私有journal套接字缺失、broker会话总线地址未指向私有socket；已修正，未改产品规则、租约或权限逻辑，失败未计通过。日志接收端仅用于broker诊断，规则后端为真实firewalld/nftables。

默认不启用此测试时明确skip，不计通过。该场景不证明Windows防火墙、真实生产网卡拓扑或主机systemd服务/计划任务；没有据此提升整个E05为全通过。新增验收晚于rc2打包，测试入口保留在源码，生产代码未再修改。
