# UI 暂停后的剩余工作核对

2026-09-29 最新平台/发行状态：真实systemd服务与timer、委派cgroup归属/整组终止/限额文件/CPU实际限流3项通过，并修复停止服务和未加载timer列表遗漏。RC26六包、2,215条内部记录/三份Web、独立双节点启动、停机恢复与兼容RC25回退均通过，见[平台与发行报告](../acceptance/reports/systemd-cgroup-2026-09-29.md)。下方“无可运行systemd环境”仅是历史结论。Windows、生产部署、跨主机、设备故障及原生通知中心未覆盖项继续分别验证；原Vim偶发/E08仍未关闭。

2026-09-29 最新补验：终端定向6/6、真实Vim/top10/10及最终3/3、类型检查通过；原偶发额外字节仍未复现，已补失败时的精确字节/连接阶段附件。绘制线程跟踪确认软件合成耗时占主导，显式SwiftShader合成微探针更慢未采用；完整60秒循环传输诊断p95 1062.5ms仍失败，不能替代正式E08结果。见[诊断报告](../acceptance/reports/local-diagnostics-2026-09-29.md)。产品、UI和RC25产物保持不变。后续需实际捕获偶发来源、可用硬件合成设备及既有外部平台验证，未把本地或全范围标为全部完成。

2026-09-29 当前结论：本轮非视觉恢复/聚焦优化、66项单测、59/59完整mock、三引擎原生恢复与RC25六包/独立恢复/兼容RC24回退已完成，详见[本轮报告](../acceptance/reports/local-closeout-2026-09-29.md)。真实套件50/51，Vim刷新额外字节一次失败，诊断与原用例各3/3复核未再现；尚未定位来源，不能把首轮写为全部通过。独立24h长测为 **PASSED_24H**，见[最终报告](../acceptance/reports/local-endurance-24h-2026-09-27.md)。当前冻结UI严格E08 Chromium/Firefox/WebKit p95为1003.9/1108/545ms，均高于50ms目标；本机无`/dev/dri`，快照专项优化未解决完整材质合成瓶颈，UI及通透效果未降级。下一步仍需定位偶发Vim应答及性能瓶颈，并在可用硬件加速、Windows真机、独立systemd、跨主机Engine/网络/注册源、物理掉电及系统通知中心等隔离环境补验。下面旧表与旧更新仅保留历史时点，整体验收仍未完成，最新动作见[执行进度](PROGRESS.md)。

日期：2026-09-26。用户要求 UI 优化暂到这里，盘点其余未完成工作。本轮仅只读核对源码、最新进度、验收矩阵及报告，没有运行回归、启动产品服务或实施功能改动。

2026-09-27 执行更新（以下旧表仅保留9月26日审计时点）：Firefox/WebKit恢复故障各3项、真实xterm检查点Worker与屏幕等价/初始化回退/运行中实际终止回退已补验证。63项前端单测、干净SDK、56项完整mock浏览器与另新增Worker终止1项、实际HTTPS/Docker **51/51**完整功能回归通过；稳定窗口DOM/扩展原生鼠标命中及Compose授权问题已闭合，Go race首轮失败后的PTY重复10次及完整Master分段复验通过，见[功能报告](../acceptance/reports/local-functional-2026-09-27.md)。**RC24已完成六包构建、2,191条包内记录/三份当前Web校验、独立启动/双Daemon/停机恢复/兼容RC23回退**，见[发行报告](../acceptance/reports/release-rc24-2026-09-27.md)。独立24h监督已超过17h、跨UTC日期及1,000个实际分钟slot，但仍RUNNING，预计今天17:12后收取终态。当前完整材质严格混合负载三引擎未达标（Chromium1240.1ms、Firefox1230ms、WebKit739ms），新分层/启动参数对照亦失败，旧UI的42ms/一小时48.1ms不覆盖本版本，见[当前性能报告](../acceptance/reports/performance-current-2026-09-26.md)。本地仍需性能收敛及长测终态；Windows/systemd/跨主机/设备掉电/系统通知中心依旧缺实际环境证据，没有标记全量完成。最新会话及下一动作见[执行进度](PROGRESS.md)。

主要功能链路已有代码与 Linux 本地真实证据；本次审计没有找到新的明确空实现，但不据此宣布全部产品合同通过。F01～F14、E01～E09 的整体验收仍未关闭。剩余重点如下：

2026-09-27 11:14补充：长测18h第二次失联/管理端重启/同runId恢复已经通过，当前两个声明故障阶段完整；只读结果收取器在后台运行，实际终态会自动写入[24h结果报告](../acceptance/reports/local-endurance-24h-2026-09-27.md)，不需要用户为收取再回复“继续”。仍在执行的测试及未达标性能不能写成“本地全部完成”。下表是旧审计快照，当前已闭合项目以本页执行更新和最新报告为准。

| 工作 | 当前证据与缺口 | 可推进条件 |
| --- | --- | --- |
| 非 Chromium 性能 | 严格双 PTY 混合负载 WebKit p95 78ms、Firefox 144ms，超过 ≤50ms。Chromium 同负载短时 42ms、既有一小时 48.1ms 已通过；无收益/破坏输入时序的试验已撤回 | 本地可继续定位绘制热点、修复后针对性复验；不能只重复已知失败长测 |
| 冻结版本回归和发行 | 最近完整真实浏览器回归为 RC15 的 49/49；后来有定向增量验证。最近发行仍 RC23（9/22），包内主 index.html 与当前 web/dist 的摘要不同，未包含后续完整外观版本 | 本地完成当前源码真实全链路回归、新包构建、独立启动/恢复/兼容回退冒烟与摘要校验 |
| 长时间稳定性 | 24 小时跨日真实监控、长期墙钟调度、更长期混合负载和全连接内存/归档边界尚缺；已有约十分钟历史采样和一小时 Chromium 证据 | 准备专用隔离环境并保留持续运行时间；本地场景可以补，跨主机/跨平台组合另验 |
| 长测入口与跨引擎恢复故障 | 现有监控入口固定 121 点、采样间隔最多 30 秒，混合性能 soak 最多 3600 秒，需要补有监督的跨日/长期入口。配额失败、旧 schema 与 IndexedDB 事务中止已经在 Chromium 通过，Firefox/WebKit 的对应组合仍缺 | 本地可完善测试入口并利用已可用浏览器补验，不涉及新视觉方案 |
| Windows 运行 | Job、ConPTY、监控、文件/备份和系统适配代码与交叉构建存在，真机运行结果缺失 | 可操作的隔离 Windows 主机或虚拟机 |
| 系统服务与计划任务 | Linux/Windows 生产链路已有实现；真实 service/timer 与 Windows 服务/计划任务生命周期缺证据。私有 firewalld 失联回滚已通过 | 独立可操作 systemd manager 及 Windows 测试环境 |
| 跨主机运维 | 本机 Docker/Compose、loopback HTTPS、隔离 netem、注册源双版本/信任轮换已覆盖；远端 Engine、跨主机长时网络和注册源证书运维组合缺证据 | 独立测试主机、Engine、测试网络与短期证书/注册源配置 |
| 设备故障 | tmpfs ENOSPC、SIGKILL、SQLite 重开及备份/调度不重放已有证据；设备写入中断电、缓存丢失、Engine 主机存储耗尽仍缺 | 可恢复快照的一次性 VM/专用故障环境 |
| 系统通知 | 浏览器原生 Notification 对象和绑定 taskId 的点击路由已验；操作系统通知中心实际绘制、真实用户点击及浏览器退出后的行为未验 | 有可观察原生通知中心的图形桌面环境 |

依据：[验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)、[Firefox 性能](../acceptance/reports/firefox-performance-dual-pty-2026-09-22.md)、[WebKit 性能](../acceptance/reports/webkit-performance-hour-2026-09-22.md)、[双 PTY 复核](../acceptance/reports/webkit-performance-dual-pty-short-2026-09-22.md)、[RC15 全套浏览器](../acceptance/reports/browser-regression-rc15-2026-09-21.md)、[RC23 发行](../acceptance/reports/release-rc23-2026-09-22.md)、[平台验收手册](../operations/PLATFORM_VALIDATION.md)、[系统通知边界](../acceptance/reports/system-notifications-2026-09-20.md)。

本地优先顺序：先定位并处理非 Chromium 绘制性能，再对冻结版本回归与重新发行；长期测试独立安排。外部平台证据保持未验证，不要求先具备所有外部环境才能推进本地工作。
