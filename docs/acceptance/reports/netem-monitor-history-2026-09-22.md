# 受控高延迟/丢包监控历史验收（2026-09-22）

## 隔离环境

- Fixture、WebKit 浏览器和 `tc/netem` 共享任务专用 Docker network/PID namespace，没有发布宿主端口，也没有连接宿主 systemd、Docker API 或生产服务。
- 通过一次性 Ubuntu sidecar 的 `iproute2`，仅在该 namespace 的 `lo` 上施加 `netem delay 80ms 10ms distribution normal loss 1%`；qdisc 回显为 `delay 80ms 10ms loss 1%`。
- 浏览器为官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 中的 WebKit 和 Chromium；fixture 使用私有 HTTPS Master 和双 Daemon。

## 结果

`tests/real/monitor-history-soak.spec.ts` 在 `BLORA_E01_HISTORY_SOAK=1` 与 1 秒采样间隔下，WebKit 与 Chromium 均 **1/1 通过（各 3.6 分钟）**，Playwright 的 `.last-run.json` 为 `{"status":"passed","failedTests":[]}`。

该场景在被延迟和随机丢包的管理链路中依次验证：121 个真实样本精确裁剪为最新 120 点且顺序递增、仅暂停 Daemon 后完整缓存标为 `stale`、重启 Master 后持久缓存仍可读取、Daemon 恢复后重新上线。规则同时影响 fixture 内 loopback 控制链路，因此不是无网络条件下的替身。

Chromium 复验还验证了真实浏览器配置：未指定 `BLORA_CHROMIUM` 时，`playwright.real.config.ts` 不再硬编码仅存在于宿主的 `/usr/bin/chromium-browser`，而是使用锁定 Playwright 版本自带的 Chromium；显式设置 `BLORA_CHROMIUM` 仍可使用本地系统二进制。

## 边界与清理

这证明了隔离 Linux namespace 上的受控 RTT/丢包行为，不是跨主机远端 Engine 或 Windows 网络证据。结束后停止并删除 fixture 容器，连同其 network namespace 和 netem qdisc 一并销毁；私有状态目录也已删除。
