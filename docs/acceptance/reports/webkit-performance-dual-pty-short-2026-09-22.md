# WebKit 双 PTY 短时性能复核（2026-09-22）

## 目的

复核 E08 的性能夹具确实驱动两条不同的可写 PTY，而不是同一会话的只读副本。此前的场景按 DOM 顺序选择前两个终端容器；复制同一会话后该顺序可能选择只读副本，不能稳定证明两条独立 PTY 都有持续输出。

## 改动

`web/tests/real/performance.spec.ts` 现在在每条会话创建并处于前台时，向其唯一可写终端注入有界 Shell 输出循环；随后仍核对两个不同 `sessionId`，并在每分钟要求两个 WebSocket 都有新增字节。此改动不降低输出速率、窗口数、传输规模或 p95 阈值。

## 环境和命令

- 私有 HTTPS Master/双 Daemon `--performance` fixture，结束后已停止。
- 官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 中的 WebKit 26.6。
- 命令：`BLORA_BROWSER=webkit BLORA_PERF_SOAK_SECONDS=60 npx playwright test --config playwright.real.config.ts tests/real/performance.spec.ts --workers=1 --trace=off --reporter=line`。

## 结果

真实 60 秒组合运行到最终断言：8 个窗口、10,000 项目录、16MiB 循环跨节点传输和两条独立 PTY 均运行。61 秒进度中两条 PTY 分别新增 6,953,314B 和 6,951,491B；全程平均分别为 115,825.5B/s 和 115,828.5B/s。四轮已验证传输、实例 `RUNNING`、队列峰值 86,010B 和两条终端均正常。

测试按设计失败：3,072 个指针响应样本 p95 为 **78ms**，超过 E08 的 **≤50ms** 目标（最大 226ms，47/3,568 长帧）。因此这项修复使负载覆盖更严格，但没有掩盖 WebKit 的性能缺口；此前无收益的后台终端画面暂停试验已撤回，产品 UI 未变。

同一修正后的 60 秒场景也在官方容器的 Chromium 153.0.8010.12 中通过：两条 PTY 分别为 116,335.0B/s 和 116,100.3B/s，p95 **42ms**、最大197.6ms、2/3,690 长帧，4轮传输与实例状态正常。这说明双 PTY 强化没有把 Chromium 的本机反馈推过阈值，WebKit 78ms 是引擎特定的未达标结果。

随后尝试把连续远端输出延后到下一动画帧再写入 xterm，目的是合并小帧；真实 Vim 交互回归暴露该策略会让 `:wq` 后紧接的 `top` 首字符落在旧前台程序，命令成为 `op`。该实现在发现后立即撤回，回退后的真实 Vim/刷新/不重放输入场景 **1/1（9.1s）**通过。不能以终端交互时序换取绘制指标。

## 清理

WebKit 容器使用 `--rm` 删除；fixture 按 Ctrl+C 停止，未保留服务、端口或凭据目录。
