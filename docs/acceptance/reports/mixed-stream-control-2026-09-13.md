# 慢日志端、批量传输与控制并行（2026-09-13）

关联A09/E08。加强 `internal/master/logs_integration_test.go` 的 `TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop`，实际TLS Master与两个Daemon、真实进程日志和跨节点文件传输。

日志进程每轮输出4097字节并sleep .001；真实WSS客户端读到首条输出后停止归还字节信用。并行在相同源节点传输8,400,000字节文件，等待任务RUNNING且节点确认CurrentOffset大于0，才发出日志实例停止任务。必须5秒内完成停止且实际实例STOPPED；结束后的归档仍能读取。随后等待传输SUCCEEDED、DestinationVerified，并直接读取目标文件与全部源字节比较。

命令：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop$' -count=1`，本机TLS监听授权下运行。最终69124退出0（14.120s）。首次57388因新增测试误用TransferResult字段名编译失败，改为实际CurrentOffset/DestinationVerified后复验通过。生产代码无修改，自建临时资源由fixture清理。

这是三类实际流量同时存在的停止响应与内容正确性证据，不证明长期内存上限或归档持续滚动上限；后者须结合并补充独立预算/长时采样。日志速率是生成命令设置，未将其当作实测速率。A09仍保留进行中，不以这一次通过覆盖全部量化要求。

后续归档滚动验证：新增 `internal/terminal/native_linux_test.go` 的 `TestLinuxPTYRepeatedArchiveRotationWithoutConsumer`。真实PTY无消费者输出32批、共64MiB，再追加完成标识；归档预算64KiB/段16KiB，每5ms请求采样，同时最终独立读取目录文件总大小。迟到挂载必须明确Gap且能读到末尾标识。

连接预算增量：真实日志客户端保持不确认，20次/50ms采样 `Conn.Stats`，检查未消费字节不超过256KiB、接收队列不超过2MiB/64帧，再执行原批量传输与停止。最终29591退出0（15.959s），未消费峰值82,590B，队列峰值79,302B。此前70252也通过，但最终版另将日志进程停止注册到正确测试清理中，确保断言失败时Daemon关闭前终止自身无限输出进程。

首次7241失败于错误假设日志必须填满信用窗口：生产日志按每个最多64KiB原始输出批次等待确认，JSON/base64传输通常约83KiB即停止增长。已按此实际背压机制调整验证；当时还出现临时目录清理与日志写入竞态，因此补上述显式停止。没有把失败运行计为通过。此次未采集全进程内存，也未覆盖45秒慢端超时提示的完整周期，仍须继续。

后续完整慢端期限通过：新增 `TestLogSlowConsumerDeadlineAndCursorReattach`，实际实例先输出第一标识，2秒后输出第二标识；真实WSS客户端读取但不返信用。保持生产45秒期限不变，必须收到SLOW_CONSUMER错误帧（连接关闭不能代替），随后以已读归档cursor重新连接原runId，必须收到严格更大序号及第二标识。

运行 `go test -race ./internal/master -run '^TestLogSlowConsumerDeadlineAndCursorReattach$' -count=1 -v`（同上述Go缓存环境及本机TLS授权），51543退出0，包47.329s；错误送达并重挂载耗时45.03165s，原游标1。自建实例由Cleanup实际停止，临时fixture关闭。`InstanceConsole.vue`已有错误码/消息显示与重新连接日志入口，本轮未新增浏览器强制背压场景，因此只将此记为实际协议/资源恢复证据。生产代码未修改，全进程长期内存采样仍未完成。

命令：上述Go环境下 `go test -race ./internal/terminal -run '^TestLinuxPTYRepeatedArchiveRotationWithoutConsumer$' -count=1 -v`。10470退出0（包3.321s，实际输出2.300s），380次采样，峰值65,358B，最终实际磁盘63,317B与内部计数一致，earliest15083/latest15099，Gap=true。测试创建的PTY/临时目录自动清理。该场景覆盖远超预算的反复滚动，不将2.3秒压力运行称为长时内存稳定性测试；全连接队列和进程内存采样仍待补。生产代码未改。

2026-09-14 长时进程内存增量：将上述真实PTY无消费者场景延长为32批、每批2MiB且间隔0.5秒（约17.5秒），每5ms采样归档目录和实际 `/proc/<pid>/status` 的 PTY shell RSS。`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-a09-mem GOFLAGS=-buildvcs=false go test -race ./internal/terminal -run '^TestLinuxPTYRepeatedArchiveRotationWithoutConsumer$' -count=1 -v -timeout 90s` 退出0；3435次采样，归档峰值65,484B、磁盘54,729B，PTY进程 RSS 从741,376B升至峰值3,764,224B，未超过基线+32MiB，earliest15157/latest15169且Gap=true。测试资源自动清理。该结果补充了短时全进程内存有界证据，仍不代表跨平台或更长时间生产负载。

2026-09-21 RC15 当前源码复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop$' -count=1` 退出0，internal/master 包耗时13.918s。默认沙箱首先因 `httptest` 无法绑定 IPv6 loopback（`operation not permitted`）退出；同一命令在获准的沙箱外执行后通过。该集成场景重新验证真实 TLS Master/Daemon 链路中的慢日志消费者、8.4MB 跨节点传输、实例停止截止时间与最终归档/目标内容核对。此项重跑不改变历史测试，也不覆盖长时全连接内存边界、Windows或更长生产负载。
