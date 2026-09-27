# 控制接收队列突发与任务完整性

范围：F04/F09、A09。Linux 本机 HTTPS Master、两个真实 Daemon、Chromium；不代表 Windows、远程长期负载或物理故障验证。

## 失败与修改

私有 `fixture-1890745423` 的40项浏览器组合为38通过、2失败（6.3m）。其中任务分页场景连续提交101个真实 mkdir，6项记录为 `INTERRUPTED` / `node_acceptance_missing`，最后任务也未成功。Master账本明确显示节点没有原接受记录；未把该结果归为环境问题或改断言接受中断。

协议接收器原先在64项队列饱和时直接关闭连接。发送方已写入 socket 的后续命令可能尚未被节点持久接受；重连对账因此保守标为中断。`internal/protocol/conn.go` 现在仅对控制物理通道等待有界接收队列释放空间；暂停 socket 读取期间额外一帧计入既有字节预算，并释放统计锁供消费者排空。取消可解除等待。交互/批量通道仍保持原额度处理和超限策略，字节上限未提高。

`TestControlBurstWaitsForSlowConsumer` 使用真实TLS、容量2的队列、20条命令，先等待队列及待交付帧饱和，再确认全部身份顺序一致且字节有界。原批量通道不能饿死控制通道测试保留。

## 验证

- `go test -race ./internal/protocol -count=1`：退出0，1.192s。
- `make build`：Master与Daemon重编译退出0。
- 新 `fixture-1287262529`：`npm run test:e2e:real -- tests/real/notifications.spec.ts`，2/2，8.2s，退出0。
- 将分页用例加强为逐一查询全部101个任务并要求全为 SUCCEEDED；同命令再次2/2，7.9s，退出0。未只检查最后一项。

Go命令环境：`GOCACHE=/tmp/blora-go-full-final54 GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。浏览器凭据通过 `BLORA_E2E_CREDENTIALS` 指向上述夹具私有JSON，不复制其内容。完整 `go test -p 1 ./... -count=1` 退出0（Master 129.494s，terminal 16.944s），`make check` 退出0。两个本轮fixture均已停止，退出0。本增量尚未纳入rc1，下一发行版本统一交付。
