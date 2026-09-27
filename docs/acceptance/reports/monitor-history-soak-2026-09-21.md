# E01 实时监控历史持续采样（2026-09-21）

为补充“真实节点采样”和“120 点裁剪”原本分开验证的缺口，新增可选真实浏览器验收 `web/tests/real/monitor-history-soak.spec.ts`。默认跳过，不延长常规浏览器回归；显式设置 `BLORA_E01_HISTORY_SOAK=1` 才执行。

在本机隔离 HTTPS 双节点 fixture 上，通过真实 Master/Daemon 指标 API 按 5 秒间隔读取同一在线节点 121 次。121 个响应均为在线采样，`observedAt` 严格递增。随后查询真实历史接口，返回 `retention=120`、恰好 120 个点；首点对应第 2 次读取，末点对应第 121 次读取，所有保留点严格递增。Playwright **1/1 passed**，耗时 **10.1 分钟**。新增测试的 TypeScript 检查 `npm run check` 退出 0。

随后扩展同一验收：在121次真实采样及在线环形历史检查后，只对 fixture 所属 Daemon 执行 `SIGSTOP`，等待 Master 将节点判为 `OFFLINE`，再读取历史接口。快速组合复验以1秒间隔收集121点，耗时 **2.8分钟**、Playwright **1/1 passed**；离线接口返回的120点全部标为 `stale`，诊断明确，时间戳恰与在线采样第2至第121点一致。

最终场景再让 fixture supervisor 收到测试专用 `SIGUSR1`，由它优雅停止并以相同 Master state directory 重启 Master，保持所属 Daemon 暂停。测试核对 Master PID 已变化且健康检查成功，重新登录后再次读取历史；重启后的接口仍返回相同120个 `stale` 点及相同时间戳，之后 `SIGCONT` 原 Daemon 并等待其重新上线。当前完整在线/失联/Master重启/Daemon恢复组合 Playwright **1/1 passed**、耗时 **2.7分钟**。supervisor 仍负责重启后的 Master 子进程，并在测试结束时一并清理。

2026-09-21 在 Firefox 155.0（Playwright Firefox v1543）上执行同一1秒组合，完整采样、失联 stale、Master重启缓存恢复及 Daemon 上线场景 **1/1 passed**，耗时 **2.9分钟**。`playwright.real.config.ts` 新增 `BLORA_BROWSER=chromium|firefox|webkit` 选择；默认仍是 Chromium，Firefox/WebKit 不使用 Chromium 的自定义可执行路径。第一次 Firefox 场景启动因漏传 `-listen` 无法按测试约定定位 fixture supervisor，该次不计通过；补上显式监听地址后整套测试通过。Firefox 在普通沙箱中无法使用临时 profile，在获准的沙箱外命令运行成功。fixture 收到 SIGINT 并退出0，进程扫描无残留，9443/9444健康端点均拒绝连接。

运行命令：

```sh
env BLORA_E2E_CREDENTIALS=<隔离 fixture 的私有凭据文件> BLORA_E01_HISTORY_SOAK=1 npm run test:e2e:real -- tests/real/monitor-history-soak.spec.ts --workers=1 --trace=off --reporter=line
# Chromium 默认 5 秒间隔，121 点约需 10 分钟

env BLORA_BROWSER=firefox BLORA_E01_HISTORY_SOAK=1 BLORA_E01_SAMPLE_INTERVAL_MS=1000 BLORA_E2E_CREDENTIALS=<隔离 fixture 的私有凭据文件> npm run test:e2e:real -- tests/real/monitor-history-soak.spec.ts --workers=1 --trace=off
# Firefox 155.0 / Playwright v1543：PASS，完整组合 2.9 分钟
```

普通沙箱无法启动 Chromium（crashpad `setsockopt: Operation not permitted`）；Chromium 测试按权限流程在沙箱外运行后通过。Firefox 也需沙箱外运行以创建临时 profile。测试结束后向 fixture 发送 SIGINT；进程检查确认 Master、Daemon、fixture、Playwright 均已退出，9443/9444 端口无监听。

夹具重启支持仅用于 Linux 测试二进制：`cmd/devfixture` Linux 编译与测试通过，Windows amd64 交叉构建通过（Windows 实现不注册该信号）。该组证据覆盖约10分钟 Linux 实时采样、真实节点失联后的 Master/SQLite 缓存回读，以及 Master 重启后的缓存恢复；完整1秒组合已在 Chromium 和 Firefox 各通过一次。尚不覆盖24小时或更长期实时运行、Windows 采样、WebKit 或远程网络。
