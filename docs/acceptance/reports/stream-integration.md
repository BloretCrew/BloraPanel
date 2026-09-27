# B05/B06 真实管理数据链路

2026-09-09；仅记录已观察的子场景，不代表完整 F/A/E 验收通过。

增量：独立stdin holder已接 `runtime.NativeInput`；POST实例日志运行的`input`使用固定runId、terminal.input和持久化console.input任务，回执仅确认交付stdin，绝不宣称程序业务处理成功。真实测试`TestConsoleInputHasOneDeliveryAndGracefulStopUsesIndependentStdin`与慢日志测试合计3.944s通过；重复同键仅一次输入、无授权403、旧run409、通过quit换行实际正常退出。新增真实浏览器日志+跨节点传输2/2通过16.0s，控制台中文、输入回显、刷新只恢复草稿/游标且无新输入POST、停止后归档complete；详情见桌面报告。

## 实现合同

节点在签名控制连接之外主动建立 interactive/bulk 两条独立 WSS。数据连接绑定当前控制代次；有界 RPC 单帧128KiB、并发8、处理器4，不将文件正文放入控制任务。文件分片确认在节点持久化后返回，上传未完整时保持 WAITING_CLIENT，节点断线不能将其推进提交队列。

PTY 创建/结束通过持久化任务。浏览器从事件游标重挂同一 sessionId，逐条接收 Protobuf 封装的输出/尺寸事件；ACK 的字节位置和事件游标分别计算。Master 只有在浏览器确认已解析本批输出后，才请求下一批归档。回放完成前禁止输入及终端自动回应。原生 Shell 另需 host.manage；Daemon 在每次读取/输入/租约操作时向 Master 查询当前授权。刷新重叠连接串行交接同一 viewTabId，避免旧连接释放新连接的租约。

Daemon 停机先取消管理循环/任务、等待数据处理器与后台工作退出，再关闭PTY、文件根句柄和数据库。管理连接重连不取消已接受任务。文件回收与终态上传记录按保留期清理，维护错误记录结构化诊断。

## 已运行

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test -race ./internal/master -run 'TestTerminalReal|TestTwoNodes|TestBrowserUpload' -v -count=1 -timeout=60s
```

退出0，8.528s。真实 TLS、SQLite、两节点WSS、Linux进程和PTY。覆盖上传断线/续传/取消、实例重复重启和管理重连、PTY创建去重、中文ANSI、刷新无输入重放、只读输入拒绝、撤权关闭已有流、结束任务后真实会话退出。测试命令最初将实际ESC作为输入按键，导致ANSI断言失败；修正为shell printf的转义序列后复验通过，没有放宽断言。

文件API完整子集另见[文件API报告](files-api.md)；真实三进程浏览器文件/PTY交互另见[桌面报告](desktop-2026-09-09.md)。浏览器测试没有拦截后端API或WSS，两个真实场景13.3s通过。

Docker专用镜像构建实际成功，真实隔离PTY测试通过1.901s：非root、只读根文件系统、固定helper与pidfd身份确认，关闭exec后主实例仍RUNNING。最初Engine探测错误地读取OSType，实际/version字段为Os；实现与替身均修正后复验。镜像/命令及平台证据见容器终端报告（terminal模块维护）。未操作已有容器。

## 仍需推进

实例追加日志与独立捕获器、日志控制台、终端分栏、多会话接管完整交互、全屏vim/top组合、长期慢浏览器/日志负载、跨节点编排、真实磁盘满与Master/Daemon进程崩溃组合仍需实现或验证。Windows只有真实适配代码与交叉编译，缺少真机运行证据。B07～B10及扩展、备份、系统管理仍属于当前完整任务。

## 后续增量：原生日志与整合回归

独立原生日志捕获器已接入runtime启动，不再直接追加无界单文件。每run由独立helper持有输出reader、16MiB分段归档及私有回环读取IPC；Daemon管理进程退出不关闭业务stdout读取端。仅在运行整体退出确认后调用有界Finish，归档失败单独记录诊断。原生日志WSS按事件游标与字节ACK分离，超出保留时发送明确OUTPUT_GAP，历史读不重建进程。独立stdin holder、Docker日志精确事件适配和完整日志控制台仍在继续。

`go test -race ./internal/master -run '^TestNativeLogArchive' -v -count=1 -timeout=60s`真实通过2.207s：浏览器故意不确认第一条输出，实例stop仍在5s测试上限内确认，停止后从节点归档回读真实中文输出。生命周期回归6.560s通过。

`go test -race ./internal/storage ./internal/protocol ./internal/master -count=1 -timeout=180s`退出0（1.304s/1.221s/48.253s），包含当前已加入的跨节点传输测试，细项由transfers报告记录；`go vet ./...`退出0；`make windows`双二进制构建退出0。Windows Job keeper/ConPTY/日志真实运行仍缺真机证据。
