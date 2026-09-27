# E08 一小时复测：恢复日志克隆优化

日期：2026-09-21。当前源码在恢复值克隆优化后完成了本机真实一小时 E08 场景，浏览器响应 p95 **48.1 ms**，达到原规划的 ≤50 ms 目标。8 个窗口、2 个真实 PTY、10,000 项目录和循环 16 MiB 跨节点传输同时运行；Playwright 退出 0，165 次传输轮换及最后目标摘要校验通过。

## 变更与定位

改动前的 RC14 一小时场景 p95 为 51.7 ms。之后带 profiler 的 60 秒定位样本中，`recovery/state.ts` 的 JSON 深克隆函数占据较多 CPU；启用 profiler 的样本 p95 为 52.7 ms，该诊断值受 profiler 扰动，不作为验收结果。内部恢复数据复制现在优先使用 `structuredClone`，Vue 代理或不可克隆值继续回退到 JSON 克隆；`json()` 仍按 JSON 规则清洗数据，例如移除 `undefined` 属性。实现见 [state.ts](../../../web/src/recovery/state.ts#L7)，代理回退和 JSON 边界语义的测试见 [recovery.test.ts](../../../web/tests/recovery.test.ts#L20)。

改动后的无 profiler 短时真实场景 1/1 通过：64.954 秒、p95 **45.2 ms**、max 71.9 ms、3/3,565 长帧，双 PTY 约 111.6/111.7 kB/s、队列峰值 79,006 B。随后恢复、桌面、文件/PTY、编辑器和云工作区相关的真实浏览器测试 27/27 通过（3.6 分钟）。`npm run check`、前端 50/50 单测及 `npm run build` 均通过。

## 一小时结果

测试以真实 loopback HTTPS Master 和双 Daemon 夹具运行，不使用 API 替身；前端不启用 profiler/长任务观察器。

| 指标 | 结果 |
| --- | ---: |
| PTY 响应样本 | 48,840 |
| 持续时间 | 3,604.492 秒；soak 配置 3,600 秒 |
| 响应 p95 / 最大值 | **48.1 ms / 160.4 ms** |
| 长帧 | 731 / 197,137 帧（约 0.37%） |
| 窗口 / 活跃 PTY / 目录项 | 8 / 2 / 10,000 |
| 已完成且校验的传输轮换 | 165 |
| 终端输出速率 | 106,036 / 106,300 B/s |
| 未确认终端队列峰值 | 86,180 B |
| JS 堆采样范围 | 35,845,693–124,648,991 B；2,030 个样本 |
| 结束时任务状态 | `RUNNING`；页面错误、终端错误和传输校验错误均为 0 |

浏览器为 Chromium 151.0.7922.71，CPU 为 AMD Ryzen 7 5800H with Radeon Graphics，12 个逻辑 CPU，主机内存约 50.5 GB。图形层为 SwiftShader，两个终端均使用 xterm 默认 renderer；这不是硬件加速测量。测试期间 API RTT p95 为 82.6 ms。

另用 `scripts/performance-resources.py` 每 60 秒只读采样 RSS 和终端归档，共 62 组，其中 57 组在浏览器负载期间、5 组在测试结束后的夹具清理尾段；无进程漏采或归档轮转读取竞态。Master RSS 峰值 39,448,576 B，两个 Daemon 中较高者 32,038,912 B。每个终端归档目录最大观测 16,774,920 B（低于 16 MiB 上限 2,296 B），最多 17 段。RSS 与归档值是 60 秒间隔样本，不代表瞬时绝对上界。

复现时先运行 `./dist/blora-devfixture --performance --listen 127.0.0.1:9444`，再将夹具输出的私有凭据路径传入以下命令；完成后对夹具发送 Ctrl+C：

```sh
BLORA_E2E_CREDENTIALS=<fixture>/browser-credentials.json \
BLORA_PERF_SOAK_SECONDS=3600 \
npm run test:e2e:real -- tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line

python3 scripts/performance-resources.py .local/<fixture-id> --samples 62 --interval 60
```

本次验收判据满足本机 loopback 场景的 p95 目标。E08 在全范围矩阵仍保持进行中：本轮没有覆盖公网远端、高 RTT/丢包、Windows 真机或非 Chromium 浏览器。新克隆实现也尚未进入之前的 RC14 发行包。
