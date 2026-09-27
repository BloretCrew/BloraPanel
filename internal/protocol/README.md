# 管理协议 v1

合同在 [`api/management.proto`](../../api/management.proto)。Go 编解码使用 Protobuf 官方 `protowire`；无需在普通构建环境安装 protoc。互操作测试以官方 `dynamicpb`/`proto` 独立解析并编码，再校验 Go 包络。字段号与 `.proto` 同时核对。新增消息类型须协商新版本；当前版本接受未知非 group 字段，拒绝重复已知字段、整数溢出、非 UTF-8 标识、错误 wire type 和畸形长度。

一条已通过 TLS、身份与 Origin 校验的 WebSocket 包装为一个 `Conn`，只属于 control / interactive / bulk 其中之一。节点对三个通道分别建立连接。`Options.CurrentGeneration` 用于发送前、实际写入前、接收后及消费队列时再次检查 Master 连接代次；调用方更新代次后还应立即关闭旧连接。`Conn` 本身不代替握手授权，调用者不得将未经身份验证的连接交给业务处理器。

```go
conn, err := protocol.NewConn(ctx, ws, protocol.Options{
    Generation: generation,
    Channel: protocol.ChannelControl,
    CurrentGeneration: currentGeneration,
})
// 注册成功后持续 Read；Send 会填充省略的协议版本、代次与通道。
err = conn.Send(ctx, protocol.Envelope{
    Type: protocol.TypeOpen, RequestID: requestID, Payload: versionedJSON,
})
```

`Send` 返回 nil 只表示 socket 写入成功。请求是否接受、实际任务结果必须通过 requestId 对账；取消或连接丢失后的未知结果不能自动重放。终端历史输入永远不由此模块重发。修改请求在服务的持久化任务账本内先接受，再执行副作用。

默认限制：控制消息含包络最多 64 KiB，其余消息最多 1 MiB；每个优先级队列 64 项；发送与接收各 2 MiB 队列预算；单流双向各 256 KiB、连接合计窗口 8 MiB、64 个活跃流。发送入队超限返回明确背压。控制接收队列满时暂停 socket 读取，额外待交付的一帧仍计入字节预算；不会仅因短时消费落后断开管理连接。交互/批量接收仍及时处理额度，无法交付时关闭对应连接；所有通道违反额度或字节预算仍拒绝。`Stats` 可采集实际排队字节、队列项数和未确认字节。单连接之外的节点总连接数、日志归档磁盘配额仍由服务负责。

所有控制消息只在控制物理连接。交互/批量内 ACK、CREDIT、PING、PONG、RESET、ERROR 优先于普通队列，DATA、RESIZE、CLOSE 等保持普通队列顺序。文件通道默认 8 MiB/s，允许最多一个 1 MiB 消息的初始突发；`BulkBytesPerSecond` 可配置。批量速率等待和实际写入共用默认 5 秒截止期限，其他物理连接不共享其等待或预算。

流服务在授权完成且交换 OPEN/OPEN_ACK 后调用：

```go
err = conn.Flow().Open(streamID, sendWindow, receiveWindow, resumeSend, resumeReceive)
err = conn.Send(ctx, protocol.Envelope{
    Type: protocol.TypeData, StreamID: streamID,
    Sequence: previousEnd + uint64(len(data)), Payload: data,
})
// 仅在 xterm 解析完成或节点写入约定持久化检查点之后：
err = conn.Consume(ctx, streamID, processedEnd)
```

DATA 的 sequence 是本方向累计字节的末尾偏移；ACK 是实际处理完毕的累计偏移，credit 必须等于本次新确认字节数；ACK 重复不能重复增加额度。CREDIT 表示绝对累计允许发送到的偏移，重复发送幂等。新流与重连检查点必须由业务服务核实，不能从任意浏览器输入恢复权限或发送预算。`Flow().Close` 在流解除授权/结束后释放预算；服务应发送 RESET 并取消相关生产者。

快照与资源事件使用每资源独立 `Cursor`。先 `Snapshot(revision)` 后应用连续 `Event(revision)`；缺口出现后持续要求新快照，不能用稍后到达的单个旧事件假装修复。`ResumeAvailable` 对已裁剪的字节归档返回明确缺口。

Ed25519 `Challenge` 绑定节点、32 字节随机 nonce、具体 connectionId、代次、通道、协议版本和期限，并加入签名域前缀。Master 必须原子消费 nonce、校验身份吊销/轮换状态；单纯 `VerifyChallenge` 不具有持久化防重放能力。

依赖 API 已按锁定源码核对：[coder/websocket](https://github.com/coder/websocket/tree/v1.8.14)、[Protobuf protowire](https://pkg.go.dev/google.golang.org/protobuf/encoding/protowire)。版本由根 `go.mod/go.sum` 统一管理，不维护第二套 gRPC 通道。
