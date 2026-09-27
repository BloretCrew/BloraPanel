# 浏览器混合负载探针

运行环境：本地 Chromium（Playwright，单 worker，1440×960），Vite 测试服务器，API 请求由浏览器边界 fixture 响应。测试打开 8 个应用窗口，在 60 个 `requestAnimationFrame` 周期中派发本地指针事件并测量端到端帧间隔。

命令：

```sh
npm run test:e2e -- --workers=1 --grep 'mixed desktop load probe' --reporter=line
```

结果（2026-09-09）：单独运行一次为 `p95=50.8 ms`、`max=153.7 ms`、4/60 长帧、Chromium heap 68,020,709 字节；纳入完整 22 项套件时为 `p95=75.4 ms`、`max=89.3 ms`、5/60 长帧、heap 67,856,994 字节；加入第二终端窗口路径后为 `p95=69.5 ms`、`max=73.9 ms`、5/60 长帧、heap 63,964,420 字节。结果均高于规划的 p95 ≤ 50 ms 目标且有运行负载波动；只测本地交互帧，第二终端为窗口级现场，未伪造活跃后端会话，不代表网络/后端混合负载性能。

尚未覆盖：两条真实活跃终端、万条真实节点目录、后台文件传输并行、RTT/日志速率/带宽记录、不同设备和长时间内存/队列/归档增长。E08 仍不能标记完成。
追加复跑（关闭 Playwright trace 以避免本机 trace 资源竞争）：完整 23/23 项串行套件通过（约 1.2 分钟）；同一性能场景记录 p95 16.9ms、最大 71.7ms、1/60 长帧、JS heap 约 64.7MB。该结果改善了本地反馈观测，但仍不覆盖真实两条活跃终端、后台传输和网络 RTT 组合。
最新复跑纳入扩展管理与监控场景后，完整 25/25 项串行套件通过（约 1.2 分钟）；性能场景 p95 17.1ms、最大 54.3ms、1/60 长帧、JS heap 约 64.1MB。仍只代表本地浏览器 fixture 负载。
性能探针现额外执行 5 次 `/api/v1/session`，最新单独运行记录 API RTT p95/max 8.7ms，交互 p95 16.8ms、最大 61.1ms、1/60 长帧、heap 约65.1MB。该 RTT 是本地 fixture HTTP 边界测量，不代表远程链路。
扩展 sandbox 与 Docker 浏览器场景加入后，完整串行套件 `npm run test:e2e -- --workers=1 --trace=off --reporter=dot` 退出 0，27/27 通过（约 1.3 分钟）；本次性能探针输出交互 p95 16.8ms、最大 60.3ms、1/60 长帧、heap 64,318,483 字节，API RTT p95/max 32.7ms。仍为本地 fixture，未覆盖真实终端、目录、传输和远程链路组合。
最新完整前端回归纳入备份、任务重试和传输列表场景，`npm run check && npm test -- --run && npm run test:e2e -- --workers=1 --trace=off --reporter=line` 全部退出 0：Vitest 28/28、Playwright 30/30；性能探针交互 p95 16.9ms、最大 71.9ms、1/60 长帧、JS heap 63,210,217 字节，API RTT p95/max 6.6ms。该数据仍来自本地浏览器 fixture，不覆盖真实终端、万条目录、后台传输和远程网络组合，E08 保持进行中。
