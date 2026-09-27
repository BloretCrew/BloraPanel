# Firefox 双 PTY 短时性能复核（2026-09-22）

## 环境与命令

- 私有 HTTPS Master/双 Daemon `--performance` fixture，结束后已停止。
- 官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 中的 Firefox 155.0。
- 命令：`BLORA_BROWSER=firefox BLORA_PERF_SOAK_SECONDS=60 npx playwright test --config playwright.real.config.ts tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line`。

测试使用已强化的性能夹具：每条新会话在仍处于前台时向唯一可写 PTY 注入有界输出，并按分钟检查两个 WebSocket 都持续有新增字节。

## 结果

真实 60 秒组合中，8 个窗口、10,000 项目录、16MiB 循环跨节点传输和两条独立 PTY 都正常运行。两条 PTY 平均为 116,029.8B/s 与 116,120.3B/s；四轮目标校验、实例 `RUNNING`、队列峰值76,161B均通过。

测试按性能阈值失败：1,020 个指针响应样本 p95 为 **144ms**，超过 E08 的 **≤50ms** 目标（最大221ms、458/1,734长帧）。Firefox 没有可用 WebGL 渲染器，两个终端均使用 xterm 默认渲染器。这不能由 Chromium 42ms 结果覆盖；与 WebKit 78ms 一起说明当前非 Chromium 浏览器未达到本机 E08 门槛。

## 清理

Firefox 容器使用 `--rm` 删除；fixture 通过 Ctrl+C 停止，测试目录已删除。
