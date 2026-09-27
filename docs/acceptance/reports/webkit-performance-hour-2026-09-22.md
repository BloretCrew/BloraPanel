# WebKit 一小时混合性能验收（2026-09-22）

## 环境与命令

- 服务：本轮私有 HTTPS Master、双 Daemon、`--performance` fixture；测试结束后已停止并删除该 fixture。
- 浏览器：官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 容器中的 WebKit 26.6；宿主为 Fedora Linux 43、AMD Ryzen 7 5800H、12 逻辑 CPU、约 47.0GiB 内存。
- 执行：`BLORA_BROWSER=webkit BLORA_PERF_SOAK_SECONDS=3600 npx playwright test --config playwright.real.config.ts tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line`；容器使用 host network，仅挂入本轮 fixture 目录只读。

## 结果

完整一小时场景实际运行 3,603.841 秒：8 个窗口、2 个真实 PTY、10,000 项目录虚拟列表与 16MiB 跨节点传输并行。传输完成 539 次，最后一轮目标校验通过；两个终端平均速率分别为 111,004 和 111,001 B/s，未确认输出峰值 86,056B，最终实例状态为 `RUNNING`。

测试**未通过**：71,496 个指针响应样本的 p95 为 **78ms**，超过 E08 的 ≤50ms 阈值；最大值 298ms，API RTT p95 64ms，长帧 3,519/175,522。Playwright 退出 1，`real-test-results/.last-run.json` 为 `failed`，失败断言为 `expect(sample.p95).toBeLessThanOrEqual(50)`。

这是一项真实的非 Chromium 失败证据，不能用此前 Chromium 48.1ms 通过结果覆盖。它说明当前 E08 尚未在 WebKit 达标；远端高 RTT/丢包、Windows 和物理故障条件也仍未验证。

## 后续诊断

为避免盲目改动，另以同一官方容器测量了 180 次空白页面双 `requestAnimationFrame`：p50 32ms、p95 33ms，说明容器本身的基础帧节奏可满足阈值。60 秒完整场景诊断显示恢复日志写入 p95 1ms、没有浏览器 long task，慢响应主要在事件处理后的绘制阶段；因此存储恢复不是此次热点。

将 WebKit 终端从现有 WebGL 渲染器切换到 xterm 默认渲染器的真实复验反而由原始 WebGL 小时结果 78ms 恶化至 130ms p95，已撤销，没有把该试验带入产品。当前最有效的后续路径是针对 WebKit 的高频终端绘制与多窗口合成做独立 profiling；不得以降低终端输出、放宽 50ms 阈值或把 Chromium 成绩替代 WebKit 验收。

## 清理

已删除任务专用 `blora-e08-webkit-20260922` 容器；fixture 已收到 Ctrl+C，确认无 `blora-devfixture` 或 9443 监听后删除 `/tmp/fixture-3952471481`。
