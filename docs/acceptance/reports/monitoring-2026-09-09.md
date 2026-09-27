# F10 指标与实例资源观察

新增 `monitor` collector 和 Master/Daemon bulk RPC：Linux 节点指标从 `/proc/stat`、`/proc/meminfo`、`/proc/net/dev` 和真实 `statfs` 读取，实例指标通过 Daemon 当前运行记录的 PID 和出生身份读取 `/proc/<pid>/status`。默认 MonitorApp 展示实时指标、历史保留量和受保护的进程终止确认。响应始终带 `observedAt`；不支持的容器 cgroup 指标明确列入 `unavailable`，不会把节点总量冒充实例用量。

接口：`GET /api/v1/nodes/{id}/metrics` 需要 `node.read`，`GET /api/v1/instances/{id}/metrics` 需要 `instance.read`；节点进程列表 `GET /api/v1/nodes/{id}/processes` 与终止任务 `POST /api/v1/nodes/{id}/processes/{pid}/terminate` 需要 `host.manage`。终止请求必须提交真实 `startTicks`，Daemon 在发送信号前再次读取并匹配出生标识。每次 RPC 重新核对资源归属、RunID 和运行状态；停止或代次变化返回不可用。

验证命令：

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test ./internal/master -run '^TestMetricsUseLiveNodeAndOwnedProcessScope$' -count=1 -timeout=60s -v
```

结果：真实双 Daemon 集成测试通过（0.693s），验证节点内存/磁盘实际值、节点权限隔离、受管进程 PID/RSS 归属和实例读取权限。Collector 单元测试验证历史环形缓冲固定 120 点且带时间戳；Master SQLite 持久历史测试通过，节点失联时返回 `stale` 历史采样并保留 120 点上限；Linux 进程身份/终止单元测试通过，历史 HTTP 页面和 MonitorApp 已提供有界快照、历史采样展示、进程搜索和排序。失联诊断细节和 Windows/容器 cgroup 指标仍未完成。

浏览器场景 `web/tests/browser/monitor.spec.ts` 进一步验证节点选择、进程搜索、内存排序和终止请求携带 `startTicks`；命令 `npm run test:e2e -- --workers=1 --trace=off --grep 'monitoring searches' --reporter=line`，1/1 通过（2026-09-09）。
MonitorApp 同时按实例资源引用请求 `/instances/{id}/metrics` 并展示实例 CPU/内存及旧数据标记；节点资源继续展示节点指标与历史采样。
实例指标现也写入 `monitor_instance_points_v1` 的每实例 120 点缓存；节点失联时 API 返回最近实例点并设置 `stale=true`。`go test ./internal/master -run 'TestMetrics|TestMetricHistory|TestInstanceMetric' -count=1` 通过（0.924s），覆盖缓存裁剪和旧数据标志。
