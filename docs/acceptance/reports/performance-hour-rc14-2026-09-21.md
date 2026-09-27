# 当前源码一小时混合负载复验 — 2026-09-21

本次运行基于 `development-20260920-rc14` 所含的当前恢复日志/终端快照源码，复验 E08 一小时场景。真实 HTTPS Master 与双 Daemon、Chromium 151.0.7922.71、AMD Ryzen 7 5800H、12 逻辑 CPU、SwiftShader 软件渲染。命令为：

```sh
BLORA_E2E_CREDENTIALS=<private fixture credentials file> BLORA_PERF_SOAK_SECONDS=3600 \
  npm run test:e2e:real -- tests/real/performance.spec.ts --workers=1 --trace=off
```

## 结果

一小时负载阶段完整运行 3,600 秒，端到端采样 3,606.506 秒、45,072 次交互。**未通过规划的 p95 ≤ 50 ms 门槛**：p95 为 **51.700000286 ms**，最大响应 **159.400000095 ms**；190,831 帧中 1,312 帧超过 50 ms。Playwright 退出码 1，唯一失败断言是 p95 超标；没有调整阈值或四舍五入成通过。

其余混合链路保持工作：8 个窗口、双 PTY 分别约 105,127 与 105,011 B/s、10,000 条目录、16 MiB 跨节点复制；负载期间 162 轮传输均完成目标校验，结束后的最终传输也完成校验。未确认字节峰值 86,193 B。浏览器 heap 在采样中为 36,908,345–135,925,802 B，结束时 77,635,705 B。测试结束时五次 API 请求的 RTT p95 为 85.8 ms。未发现浏览器 pageerror 或终端错误。

每分钟只读 RSS/归档采样显示 Master 约 34.5–36.5 MB、Daemon 约 25–30 MB；单终端归档最大采样值为 16,775,268 B，低于 16 MiB 上限 1,948 B。随后采样观察到归档段轮转并回落。定时采样不保证捕获瞬时 RSS 峰值。原始性能对象见[结果 JSON](performance-hour-rc14-2026-09-21.json)。

测试唯一产品验收失败为交互 p95。此前 rc3 版本的一小时 p95 46.5 ms 是旧源码证据，不能覆盖本次当前源码的失败。E08 继续进行中；下一步用短时诊断定位输入排队/主线程长任务，再评估有证据支撑的修复，并在修改后决定是否需要新的小时验收。

测试退出后，Playwright、fixture 和采样均已停止。fixture 会话退出码 0；停止后的定向只读检查未发现属于 `.local/fixture-129414522` 的 Master/Daemon 进程。
