# 当前源码 E08 短时诊断 — 2026-09-21

## 场景与命令

两次均使用真实 loopback HTTPS Master/双 Daemon、Chromium 151.0.7922.71、AMD Ryzen 7 5800H/12 逻辑 CPU 和 SwiftShader；负载为 8 个窗口、双 PTY、10,000 项目录、16 MiB 跨节点复制。每次为独立 `--performance` 夹具，完成后均停止；只读进程检查未发现相关 Master、Daemon 或 Playwright 残留。

基础慢输入/长任务诊断：

```sh
BLORA_E2E_CREDENTIALS=<private fixture credentials file> \
BLORA_PERF_DIAGNOSTICS=1 BLORA_PERF_SOAK_SECONDS=60 \
  npm run test:e2e:real -- tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line
```

第二次在同场景加浏览器 CPU profiler：

```sh
BLORA_E2E_CREDENTIALS=<private fixture credentials file> \
BLORA_PERF_DIAGNOSTICS=1 BLORA_PERF_PROFILE_SOAK=1 BLORA_PERF_SOAK_SECONDS=60 \
  npm run test:e2e:real -- tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line
```

两次 Playwright 诊断均退出 0（各 1/1，约 1.4 分钟）。每次有 3 轮 16 MiB 传输目标校验、8 窗口及两个实际 PTY 持续输出；队列峰值分别 86,202 B 和 66,672 B，没有 pageerror。

## 观测

| 指标 | 基础诊断 | CPU profile 诊断 |
| --- | ---: | ---: |
| 采样时长 / 输入样本 | 65.62s / 1,224 | 67.60s / 1,344 |
| 输入 p95 / 最大值 | 47.4 / 76.5 ms | 47.4 / 80.2 ms |
| 超 50ms 帧 | 14 / 3,468 | 11 / 3,634 |
| 长任务 | 59、54 ms | 63、57 ms |
| API RTT p95 | 45.8 ms | 50.6 ms |
| 恢复日志同步 `setItem` p95 / 最大值 | 1.0 / 6.7 ms | 0.9 / 8.2 ms |

最慢输入同时出现事件处理开始前排队和处理后帧渲染延迟：两个诊断里都有约 45–60 ms 的输入延迟样本，也有约 48–69 ms 的处理后延迟样本。恢复日志写入的单次 p95 很低，profile 中热点分散在 IndexedDB `put`、xterm `_nextCell`/`serialize`、虚拟列表 `replaceChildren`/`createRow`、`_diffStyle`、`setItem` 和 GC；没有证据指向单一主因。profile 里未映射的 minified frame 名称 `du` 不做源码归因。

## 解释边界与后续

PerformanceObserver 和 CPU profiler 会改变线程调度；本报告的 p95 低于 50 ms **不能**覆盖当前源码一小时实测 51.7 ms 的失败，也不算 E08 通过。这组数据用于说明失败仍可能由终端渲染、列表更新和事件排队共同造成。下一步先完成当前源码常规真实浏览器套件的一次串行回归，再针对有证据的非视觉热点做小范围优化；之后才决定是否重跑一小时场景。当前源码一小时结果见[小时复验报告](performance-hour-rc14-2026-09-21.md)。
