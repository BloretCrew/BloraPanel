# 受控高延迟/丢包分钟调度验收（2026-09-22）

## 环境

任务专用 Docker network/PID namespace 内的 `lo` 使用 `tc/netem` 实际配置为 `delay 80ms 10ms distribution normal loss 1%`。私有 HTTPS Master、双 Daemon 与官方 Playwright WebKit 同处该 namespace；没有发布宿主端口或连接宿主服务。

## 结果

`tests/real/schedule-wall-clock.spec.ts` **1/1 通过（1.5 分钟）**，Playwright `.last-run.json` 为 `{"status":"passed","failedTests":[]}`。

真实 `* * * * *` 计划在相邻 UTC 槽 `05:19:00Z` 与 `05:20:00Z` 各接受一次。第一个快照保存初始正文；更新正文并刷新浏览器后，第二个快照绑定更新后的源版本。计划删除后，两个 task/request 身份与已接受集合一致；两个快照均按受控恢复计划恢复，正文分别精确匹配各自版本。

## 边界与清理

这是单机隔离 namespace 的受控延迟/丢包证据，不是跨主机远端网络、设备写缓存丢失、物理掉电、Windows 或长期墙钟验收。结束后 fixture 容器及其 netem network namespace 已删除，私有状态目录已清理。
