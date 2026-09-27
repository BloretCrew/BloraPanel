# A09 慢日志消费者持续背压验收（2026-09-21）

在 `TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop` 上新增可选 `BLORA_A09_SOAK_SECONDS`，默认 0 保持常规回归速度，允许 0–30 秒。测试用真实 TLS Master/双 Daemon、持续日志进程和跨节点文件传输；WSS 客户端读取首帧后一直不归还日志字节信用，逐秒检查连接未消费字节和接收队列上限，随后启动真实文件传输并在同一慢日志场景下停止实例，验证停止期限、归档重读和目标文件摘要/正文。

当前源码命令：

```sh
env BLORA_A09_SOAK_SECONDS=30 go test -race ./internal/master -run '^TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop$' -count=1 -v -timeout 90s
```

退出 0。30 个逐秒样本中，未消费字节峰值为 **33,048 / 262,144 B**，接收队列峰值为 **28,320 / 2,097,152 B**；整个集成测试 44.71 秒、包 45.776 秒。后续并行传输、5 秒内实例停止、历史日志读取及传输目标核对全部通过。`go vet ./internal/master` 退出 0。

一次 60 秒探索运行保持队列有界，但结束时旧 WSS 会话已关闭，不能再向旧连接发送 ACK；这是超过日志服务45秒慢消费者期限后的预期边界，不计作 soak 通过。现有 `TestLogSlowConsumerDeadlineAndCursorReattach` 已单独验证45秒时收到 `SLOW_CONSUMER` 诊断并可按归档游标重连。当前有效 soak 因此限制在30秒以内。

这是持续背压下的30秒当前源码证据，不代表长时全连接内存上界或跨平台通过；Master/Daemon进程RSS和归档滚动的其他压力证据见[混合流报告](mixed-stream-control-2026-09-13.md)。测试实现见[日志集成测试](../../../internal/master/logs_integration_test.go)。
