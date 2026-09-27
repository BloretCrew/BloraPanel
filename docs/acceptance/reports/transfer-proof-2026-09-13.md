# 大文件传输分块校验索引（本机512MiB验证通过）

原跨节点传输每64KiB分块对整源文件哈希两次，Master每8块又读取完整Stat；512MiB试跑数分钟仅推进数MiB。新增仅供传输路径使用的内存分块校验索引，原ReadChunk的每调用全文件检查保持不变。

首次建立索引时，完整读取文件并核对请求的SHA-256，同时保存每64KiB块的SHA-256。后续读取核对完整覆盖块（包括非对齐区间）、文件身份/大小/模式/mtime；恢复mtime的篡改也无法通过被改块的哈希。每个文件服务最多64个索引、8MiB哈希内容，无源文件副本；节点重启或缓存淘汰后重新建立完整校验。

跨节点传输在发布目的文件前新增完整源Stat/哈希及元信息复核，原目标摘要验证、最终源验证、移动前源对象关系核对、授权及取消检查保留。中间索引Stat不能用作最终源版本证明：已读取的前缀之后被更改并还原mtime，也必须在发布前拒绝。

已通过：

- `go test -race ./internal/filesystem -run 'Test(TransferProof|UploadQuotaAndReadChunks)' -count=1`：1.045s。非对齐块、恢复mtime后内容改变、空文件索引条目上限、同内容对象替换、未知摘要拒绝及24个并发读者。
- `go test -race ./internal/master -run '^TestTransfer' -count=1`：46.307s。原真实TLS传输集成回归，包括源变化、取消、权限撤销、Master恢复和目标验证。
- `go test -race ./internal/master -run '^TestTransfer(ChangedReadPrefixCannotPublish|SourceDaemonRestartRebuildsProof)$' -count=1`：25.963s。真实已读源前缀改写且恢复mtime后，明确由发布前全源验证拒绝，目标未发布；实际源Daemon重启后重建索引并完成验证。

完整filesystem模块race通过（1.300s），校验索引版本的make check/build/windows与fixture构建通过。Windows只有交叉构建，不能计真机验证。

fixture新增 `--performance-transfer-mib 1..1024`，默认16；与`--performance`一起显式启用。只写本次私有测试资源。

## 512MiB实际结果

fixture-4036895406，真实8窗口、双PTY、万条目录及512MiB源，未使用API替身。首次交互采样p95 40.9ms通过目标，但原小文件测试等待预算不足；在3.7分钟手动中断浏览器测试，**未取消后台任务**。测试支持按源/目标实例、源版本和测试目标路径重新核对后，用`BLORA_PERF_TRANSFER_TASK`继续观察同一任务；不重复创建传输。大文件观察期限随规模增加，交互目标仍为50ms。

同一任务 `659de8466bcbe9ef91aed784ce047b90` 最终SUCCEEDED、destinationVerified=true。续接测试通过（约5.0分钟），从原任务创建到观察成功为543.324s，536,870,912B / 543.324s，约**0.942MiB/s**。这包含本机磁盘同步、调度、校验和观察延迟，不是链路带宽上限；仍可优化吞吐，但不再每分块重读整源文件。

续接样本：p95 42.4ms、最大57.4ms、2/357长帧，heap70,985,771B，API RTT p95 71.6ms，实际两终端载荷40,749/40,418B/s；未确认字节峰值87,404B。采样结束后台任务仍RUNNING。图形设备为SwiftShader，两个终端自动使用默认渲染器。

独立`sha256sum`读取实际源和目的文件，均为：

```text
9acca8e8c22201155389f65abbf6bc9723edc7384ead80503839f49dcc56d767
```

两个浏览器运行共创建4个测试PTY，结束时各归档目录占用约13MiB（默认16MiB预算）；原输出循环有界，不能据此宣称无限时长稳定。fixture已Ctrl+C停止。长期内存/完整归档轮转/远程网络及Windows真实运行仍需补充。

## 2026-09-14 批量步进优化后的真实512MiB复验

首轮同日运行使用了改动前编译的旧 `dist/blora-devfixture`（fixture-4144551140），其133.246s结果仅作为校验索引基线，不能归因于16步批处理。随后重编译夹具并新建 fixture-1569035732，在 `transferChunksPerStep=16`（每步最多16个64KiB分块、每块仍执行检查点和取消检查）源码下，真实 HTTPS Master/双 Daemon、八窗口、双 PTY、10,000项目录和512MiB跨节点传输一次完成，未使用API替身。Playwright 性能场景 1/1 通过（约2.0分钟），传输目标逐字节校验成功。

有效结果：536,870,912B 用时 **110.868s**，约 **4.62MiB/s**；采样4.817s期间反馈p95 **39.0ms**、最大56.3ms、285帧且0个长帧，API RTT p95 43.7ms，JS heap 53,729,723B，未确认字节峰值86,306B，两个终端实际载荷速率71,450.7/53,623.8B/s。浏览器151.0.7922.71，AMD Ryzen 7 5800H、12逻辑CPU，SwiftShader环境自动使用default/default终端渲染器。

相对旧夹具基线133.246s（约3.84MiB/s），本次耗时减少约16.8%；相对此前每步8块的索引续接543.324s（约0.942MiB/s），减少约79.6%。两次本地场景均包含磁盘同步、调度、校验和轮询观察，不代表网络带宽上限。有效场景只覆盖一次约2分钟运行；长期内存、完整归档轮转、远程网络、物理故障和Windows真机仍需补充。fixture已Ctrl+C停止，未留下测试进程或带 `blora.test=1` 的Docker对象。

传输当前源码定向回归：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-transfer-final14 GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestTransfer' -count=1` 退出0（64.894s）。

同一批处理源码在 `BLORA_PERF_SOAK_SECONDS=60` 的长时混合场景中再次完成512MiB传输（fixture-2822501719，109.245s、约4.69MiB/s），目标校验成功；该次完整交互和堆/队列采样详见[真实性能报告](performance-real-2026-09-13.md)。
