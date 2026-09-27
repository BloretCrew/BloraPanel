# 平台验收运行手册

本文只描述在相应测试环境中执行真实平台验收的入口。交叉编译、Linux 替身和 loopback 代理可以证明代码或传输合同，但不能替代下面的运行结果。

## Windows Job、ConPTY、监控和系统适配

在 Windows 10/Server 2016 或更新版本的隔离测试主机上，从仓库根目录执行：

```powershell
go test -race -v ./internal/runtime -run 'TestWindows'
go test -race -v ./internal/terminal -run 'TestWindowsConPTY'
go test -race -v ./internal/runlog -run 'TestWindows'
go test -race -v ./internal/monitor -run 'TestWindowsNative'
go test -race -v ./internal/systeminfo -run 'TestWindowsNativeServiceEnumeration|TestWindowsTaskSchedulerEnumeration'
```

这些用例会创建并清理自己的进程、Job、ConPTY 和只读系统查询；不要把生产 Daemon、现有服务或计划任务作为测试目标。Job keeper 需要允许独立子进程和 `CREATE_BREAKAWAY_FROM_JOB`；权限不足应记录为环境失败，不应改成跳过或成功。

Windows 交付包先在 Linux 生成：

```sh
make windows
```

这只验证 `blora-master.exe` 和 `blora-daemon.exe` 可构建，不能计作 Windows 行为验收。

## systemd service/timer

在一次性虚拟机或明确隔离的 systemd 容器中，先确认实际 user/system manager 可连接，再运行 `internal/systeminfo` 的服务和计划任务集成测试。测试环境必须允许创建临时 `.service`/`.timer` 单元、启动/停止/启用/禁用并读取状态；不得连接当前开发宿主机的生产服务或用户总线。当前 runner 没有可操作的 manager，相关结果保持“环境缺失”。

## 远程 Docker Engine 与网络故障

在独立远端 Engine 上准备短期 TLS CA/客户端证书和测试账户，将 endpoint 通过私有环境变量传入，不把证书或地址写入仓库：

```sh
BLORA_CONTAINER_E2E=1 \
BLORA_CONTAINER_E2E_ENDPOINT=https://engine.example.invalid:2376 \
BLORA_CONTAINER_E2E_TLS_DIR=/private/test/docker-tls \
go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1 -v -timeout 5m
```

远端主机必须只包含本轮带有 `blora-e2e-*` 标签的对象；测试结束后按标签清理容器、网络、卷和临时镜像。高 RTT/丢包需在测试链路或远端测试 namespace 使用受控 `tc/netem`，分别记录 RTT、丢包率、控制请求截止时间、归档和任务最终状态。当前 `BLORA_CONTAINER_E2E_REMOTE_TLS=1` 的 loopback TLS 代理只验证 HTTPS 传输，不是远程 Engine 证据。

若没有远端主机，但需要先验证浏览器/监控在受控网络条件下的行为，可让 task-owned fixture、浏览器和 `tc/netem` 共享一个 Docker network namespace；不要对宿主 `lo` 或生产接口加规则。2026-09-22 的完整可复现结果见[netem 监控历史报告](../acceptance/reports/netem-monitor-history-2026-09-22.md)。这种单机 namespace 场景只能补充受控延迟/丢包证据，仍不替代远端 Engine 验收。

## WebKit 浏览器

项目的 `playwright.real.config.ts` 支持 `BLORA_BROWSER=webkit`。若宿主 Linux 的 Playwright WebKit 缺少匹配 ABI 库，可用与项目 `@playwright/test` 同版本的官方容器运行；容器只连接本轮私有 fixture。监控历史场景需要读取并暂停 fixture Daemon，因此将 fixture 目录以原路径只读挂入，并使用 host PID namespace；测试源码先按该私有配置路径精确确认 PID，不会选择其他进程。

```sh
./dist/blora-devfixture -listen 127.0.0.1:9443 -state-parent /tmp
# 将下面的 FIXTURE_DIR 设为上一命令输出的私有目录。
docker run --rm --network host --pid host \
  -e BLORA_E2E_CREDENTIALS="$FIXTURE_DIR/browser-credentials.json" \
  -e BLORA_BROWSER=webkit \
  -e BLORA_E01_HISTORY_SOAK=1 \
  -e BLORA_E01_SAMPLE_INTERVAL_MS=1000 \
  -v /data/instances/blora-panel:/work \
  -v "$FIXTURE_DIR:$FIXTURE_DIR:ro" \
  -w /work/web mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test --config playwright.real.config.ts \
  tests/real/monitor-history-soak.spec.ts --workers=1 --trace=off --reporter=line
```

测试结束后向 fixture 发送 Ctrl+C，确认其进程和目录均已清理。该路径验证 Linux 上的 WebKit 浏览器链路，不代替 Windows 浏览器、远端网络或一小时性能验收。

## 物理故障和长期场景

掉电、设备写缓存丢失和 Engine 主机存储耗尽只能在可恢复的快照/虚拟机上执行。每次运行前记录基线和测试对象身份，运行后核对 SQLite 完整性、任务回执、归档对象和源数据；不要在开发宿主机上填满磁盘或切断生产服务。WebKit、Windows 浏览器和系统通知中心也要分别记录真实引擎结果，不能用 Chromium 或 Notification API 替代。

本地 Linux 的真实 24 小时监控/分钟调度可使用独立私有 fixture，无需切断宿主网络或改变系统时间。构建当前 Master/Daemon 后启动受监督入口：

```sh
python3 scripts/local-endurance.py start
python3 scripts/local-endurance.py status .local/endurance-<启动返回的标识>
# 只有需要停止时执行；入口按完整进程身份停止所属测试资源。
python3 scripts/local-endurance.py stop .local/endurance-<启动返回的标识>
```

默认复制当前二进制、使用随机 loopback TLS 端口，每 5 秒采样，真实分钟备份并保持三份归档，第 6/18 小时验证所属 Daemon 失联与同库 Master 重启缓存。末尾要求实际墙钟和单调时钟均满 24 小时、跨 UTC 日期、采样覆盖及进程/归档预算正常，实际恢复最新正文并清理测试服务。监督中断后的 `resume` 重新开始完整连续覆盖，不累计无监督空档。只有最终 `PASSED_24H` 且清理成功才计作通过；`RUNNING` 与短测不算。参数、身份保护、当前证据与接管见[长测入口报告](../acceptance/reports/local-endurance-2026-09-26.md)。
