# 本地真实墙钟长期验收入口（2026-09-26）

关联 E01/E04，并为 A09 提供服务进程、独立日志 helper 和归档资源长期采样。**24 小时验收目前尚未通过；运行中不能计为通过。** 本次保持 UI 冻结，只新增 `scripts/local-endurance.py`，不修改生产服务、监控/备份实现或已有性能验收标准。

## 验收内容

入口创建唯一 `0700 .local/endurance-*` 私有目录、随机 `127.0.0.1` HTTPS 端口和单节点真实 Master/Daemon；将本轮源码编译的两个二进制复制到目录内并记录 SHA256，后续发行构建不会替换运行中的版本。使用真实 TLS 校验与随机私有管理员凭据。凭据仅存 `0600` 文件，不输出密码、登记 token 或 Cookie。

默认连续实际墙钟 **86,400 秒**，按 **5 秒**节奏调用真实节点与实例指标 API，不修改系统时间，也不注入模拟时间。每次核对在线/非 stale、采样时间严格递增、实例 runId 不变；节点历史必须精确等于最近 120 次真实采样。末尾以只读私有 SQLite 查询核对实例历史也只有 120 条有序记录，执行 integrity_check。

创建生产 `* * * * *` / `Asia/Shanghai` / `run_once` / `wait_one` 文件备份计划，由真实分钟 ticker 触发。每个成功槽检查 taskId/requestId、真实分钟边界、来源版本与 ready 归档，随后通过真实文件任务修改来源正文，并按保留策略只保留最新三份 ready 快照。最终分页读取全部所属定时任务，拒绝重复或未结束任务；每个普通墙钟分钟均须有成功 occurrence，只有明确记录的故障窗口及下一分钟对账区间按 run_once 语义排除。最终真实恢复最新快照到唯一新文件，并核对入账版本对应的中文正文。

默认第 **6 小时、18 小时**只暂停本 fixture 的 Daemon：等待节点判定离线，确认节点 120 点及同 runId 实例缓存为 stale；保持 Daemon 暂停并重启同私有数据库的 Master，重新认证后核对相同缓存与明确诊断；恢复相同 Daemon，确认原实例 runId 仍为 RUNNING。计划在这些窗口内采用实际策略处理，不伪造网络结果。普通采样间隔最大不得超过 15 秒，24 小时普通采样数量须达到配置节奏的 99%；声明的故障窗口单独记账。最终还要求跨至少两个真实 UTC 日期、实际墙钟与单调时钟均达到 86,400 秒、至少 121 个真实点与两个分钟槽，以及生产日志归档已实际轮转。

资源上限在启动前固定并随 run.json 记录：每个 Master/Daemon/日志 helper RSS ≤256 MiB 且增长 ≤128 MiB；Master/Daemon 文件描述符增长 ≤128；每份实例归档严格使用生产 **16 MiB** 上限；所属 ready 快照最多三份；私有运行状态与证据磁盘总量 ≤256 MiB（不含固定复制二进制）。进程、helper 和每段归档在每次采样时读取实际 `/proc`/文件事实。采样和阶段日志分别有文件大小及份数轮换上限，服务输出由监督线程持续读取并轮换，不依赖浏览器重复开窗。

这些 RSS/FD 数值是本场景的预先声明上限，并不把其他设备、WebKit 性能失败或跨平台完整 A09 验收改判为通过。此次低频日志负载为实例约 1 KiB/s；慢 WSS 消费者、混合双终端/大文件压力由既有专项与并行性能验收承担。

## 监督、停止与恢复

`start` 通过独立 OS session、DEVNULL stdin 和私有日志启动 worker，没有工具会话/PTY 管道；启动命令退出后 worker 仍运行。私有文件锁阻止两个监督者同时接管同一目录，checkpoint 以 fsync+原子替换写入。`status` 展示 worker 身份、检查点时间、阶段/计数/峰值，不输出凭据。默认沙箱不能打开监听 socket，启动已按授权在沙箱外执行；若沙箱不能读取沙箱外 `/proc/<pid>/exe`，状态明确显示 UNKNOWN，而不会据此认定进程已死。

信号使用 pidfd，并同时核对 PID、startTicks、exe、bootId。停止只删除记录的 scheduleId，并通过真实 API 确认指定 instanceId 停止，然后退出记录的服务进程；保留私有证据。若监督 worker 遭强制结束，`stop` 启动独立恢复清理者，使用相同身份/目录恢复管理连接并停止原实例。`resume` 仅在原 worker 已退出且无清理错误时接管，继续同私有数据库/节点/资源身份；**重新开始完整连续覆盖区间**，不将无人监督的空档累计为 24 小时通过，不重放结果不明的远程写入。

```sh
# 二进制先从当前源码构建；长测复制它们并记录哈希。
env GOCACHE=/tmp/blora-endurance-go-build GOFLAGS=-buildvcs=false go build -trimpath -o /tmp/blora-endurance-master-20260926 ./cmd/master
env GOCACHE=/tmp/blora-endurance-go-build GOFLAGS=-buildvcs=false go build -trimpath -o /tmp/blora-endurance-daemon-20260926 ./cmd/daemon

# 5 分钟真实入口验证：121点、两个以上真实分钟、130s故障检查点。
python3 scripts/local-endurance.py start --master /tmp/blora-endurance-master-20260926 --daemon /tmp/blora-endurance-daemon-20260926 --seconds 300 --interval 1 --fault-after 130

# 默认24小时真实验收；打印的runDirectory用于所有接管命令。
python3 scripts/local-endurance.py start --master /tmp/blora-endurance-master-20260926 --daemon /tmp/blora-endurance-daemon-20260926
python3 scripts/local-endurance.py status .local/endurance-<本次标识>
python3 scripts/local-endurance.py stop .local/endurance-<本次标识>
python3 scripts/local-endurance.py resume .local/endurance-<本次标识>
```

## 当前证据

Python 编译检查退出 0；两个当前源码 Linux 二进制私有构建退出 0。首次默认沙箱启动因本地 socket `EPERM` 失败，不计作通过，且未启动服务；后续按授权在沙箱外启动。

5 分钟短测目录 `.local/endurance-gm12uhtu`，端口 `48241`，开始 `2026-09-26T09:05:00.890892Z`，于 `09:10:05.089801Z` **PASSED_SHORT**。实际总时长 **304.199 秒**，**253 次**真实采样、**5 个**唯一成功分钟槽（09:06～09:10），来源版本随真实写入改变。09:07:11 暂停所属 Daemon，09:07:56 离线缓存 120 点验证通过；Master 从 PID 549459 重启为 580377，09:07:57 重启后相同 120 个 stale 点验证通过；09:07:59 原 Daemon/原实例 runId 恢复。末尾恢复最新归档 `7b27c3c5edc4e021e460a06c4f9b9b44` 后中文正文匹配入账来源版本、SQLite integrity_check=ok、实例缓存恰好120条；最后 `servicesAlive=[]`。RSS峰值 Master **29,339,648B**、Daemon **28,172,288B**、独立日志 helper **15,761,408B**，Master/Daemon FD峰值14/19，归档峰值321,167B、私有状态/证据峰值7,508,444B。本短测不代表24小时通过。

停止/恢复/监督中断验证目录 `.local/endurance-vff89bpp`，端口 `55993`。正常停止于 `09:06:56Z` 写入 STOPPED，服务清理 `servicesAlive=[]`。随后 resume 使用相同 nodeId/instanceId 重开实际服务，采样重新从第一个点开始，新受控启动具有新 runId，并在 09:08 产生真实成功分钟备份。以错误 startTicks 的身份发信号被拒绝；只 SIGKILL 本测试所属 worker 后，stop 启动独立清理者，`09:08:32Z` 再次 STOPPED，`servicesAlive=[]`。这证明监督进程中断后的接管清理，不作为完整 24 小时通过。

最终入口另复验 `.local/endurance-m0bh00qf`（端口35923、worker602030）：**PASSED_SHORT**，从 `09:12:54.961049Z` 至 `09:17:57.154820Z`，实际 **302.194 秒**，**251 个**真实点、**5 个**唯一成功分钟槽。普通采样最大间隔 **1.829 秒**；45秒真实失联、120点 stale 及同库 Master 重启缓存回读、原 runId恢复通过。新增监督等待心跳和先关闭所属计划再以100条分页做终态审计也实际执行；最终新文件恢复正文匹配、SQLite integrity=ok，`scheduleRemoved=true`、`servicesAlive=[]`，无清理错误。RSS峰值 Master **28,839,936B**、Daemon **28,352,512B**、日志helper **15,691,776B**；归档峰值 **320,100B**、私有状态/证据峰值 **7,950,633B**。三个已结束短测试目录（包含正常停止/恢复测试）所属进程另做精确目录扫描，零残留。

## 24小时运行接管

正式长测 **RUNNING，尚未通过**。目录 `/data/instances/blora-panel/.local/endurance-zs_dzpzj`；独立 HTTPS `https://127.0.0.1:58787`。实际覆盖从 **2026-09-26T09:12:32.216331Z（北京时间17:12:32）** 开始，要求至少运行至 **2026-09-27T09:12:32Z（北京时间次日17:12:32）**，随后执行真实恢复与最终检查。故障检查约在北京时间本日23:12、次日11:12；Master PID会按设计变化，接管必须读最新checkpoint而不能只依赖这里的初始PID。

| 所属对象 | 初始 PID | Linux startTicks | executable |
| --- | --- | --- | --- |
| worker | 597961 | 9882482 | `/usr/bin/python3.14` |
| Master | 598058 | 9882540 | `.../endurance-zs_dzpzj/bin/blora-master` |
| Daemon | 598122 | 9882575 | `.../endurance-zs_dzpzj/bin/blora-daemon` |

bootId为 `cce4bc4d-465e-406f-8036-c2e67912d160`。nodeId `60750e26a5ed554e414ad1b740d4f5e2`、instanceId `0221babaf8e349ca12edf7249ccad13a`、runId `7d43fbe2017b5c6b1aa1e017a24eb45c`、scheduleId `13c1fbd78805e518926058a46e583b5a`。

复制二进制SHA256：Master `f8b798e13802306ef8ff806c3abf55a45a055ac9616c6f75cdbabe048e9fa3f8`；Daemon `5c0965d65441c6b65a4c18f1a1305ba3c191523d156283a8db6ed26200e98bed`；入口SHA256 `c3e2488eaab23f4f4b69e656e56473b5e66b4957ce6b9bc130d57b5194e09f3b`。身份、预算与哈希见私有 `run.json`/`worker.json`，动态事实见 `checkpoint.json`，真实采样见有轮换的 `samples.jsonl*`，分钟槽和故障阶段见 `events.jsonl*`，服务诊断为 `master.log*`、`daemon.log*`。仅凭日志存在或 RUNNING 不能提升验收状态；最终须为 PASSED_24H 且清理成功。

```sh
python3 scripts/local-endurance.py status .local/endurance-zs_dzpzj
# 需要停止时才运行；不直接按PID猜测杀进程。
python3 scripts/local-endurance.py stop .local/endurance-zs_dzpzj
# 只有意外/明确停止后才恢复；新的完整24h覆盖会重新开始。
python3 scripts/local-endurance.py resume .local/endurance-zs_dzpzj
```

Windows 真机、远端主机网络、物理掉电与完整跨平台长测仍不由本地结果覆盖。后续接管需收取实际24小时终态并补证，不能把此启动记录改写成完整通过。

## 2026-09-27 自动收取入口

新增 `scripts/endurance-report.py` 仅只读现有私有run/checkpoint/事件及固定二进制，原测试worker继续运行；收取器不认证、不发远程请求、不暂停/重启/停止服务。报告只提取白名单统计，不输出凭据、Cookie、命令、正文或原始错误字符串。运行中至少每分钟更新报告，终态交叉核对本覆盖区间的实际归档恢复/SQLite完整性、计划删除、所属服务清理和固定二进制摘要；24h状态另要求实际时长及跨UTC日期。检查点超过120秒未更新或等待截止时，以未完成报告退出，不擅自接管。

```sh
python3 scripts/endurance-report.py .local/endurance-zs_dzpzj \
  --wait --timeout-seconds 30000 \
  --report docs/acceptance/reports/local-endurance-24h-2026-09-27.md
```

已启动独立OS session收取器，PID1366399/startTicks16319189/同bootId，03:05:16 UTC开始；私有身份与日志 `.local/evidence/rc24/endurance-collector.json`、`endurance-collector.log`。实际结果写入[动态长测报告](local-endurance-24h-2026-09-27.md)，本记录仍是运行时点，不能提前提升为通过。预计本地17:12后执行恢复/清理并获得最终状态；失败也按真实状态记录，报告不会自称全部功能验收完成。

当前入口的已完成真实短测 `.local/endurance-m0bh00qf` 被正确收取为PASSED_SHORT、退出0。旧第一轮短测缺少后续加入的 `scheduleRemoved` 字段，收取器明确返回EVIDENCE_INCOMPLETE，不能默默补造字段。四项额外边界验证（短测伪标24h、缺最终恢复、清理仍有服务、私有字段不导出）均通过，命令0.323s、退出0，记录 `.local/evidence/rc24/endurance-report-check.log`；这些是报告拒绝错误证据的测试，不构成长测通过证据。
