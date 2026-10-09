# Blora Panel 验收矩阵

当前判定：2026-10-08。实现与验证分别记录；以下“本次范围已验证”仅指用户确认的交付范围，**不是所有原环境通过**。历史全部结果及失败保存在[完整归档](archive/ACCEPTANCE_MATRIX-before-closeout-2026-10-08.md)，原规划合同保留。

2026-10-09 文档补充：E09 的生产启动入口和三组件职责已写入 [English README](../../README.md#production-startup) / [中文 README](../../README.zh-CN.md#生产环境启动)，含 systemd/Nginx 示例。命令与配置按现有程序核对；systemd 使用现有本地二进制路径进行解析验证通过，未安装服务；Nginx 配置未在本机运行。本次文档更新不增加生产部署或新 Windows 运行的 PASS，也不修改已发行 beta 身份。

最新发行范围：用户选择 GPL-3.0，委托版本及预发布类型；**[v0.1.0-beta.1 已预发布](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1)**，`GPL-3.0-only`。固定源码、六包/源码包对应、三份 Web、SDK、清单/摘要及重复打包通过；实际新包初始化/恢复、RC34 升级及兼容快照回退通过；九个远端附件核对一致。详见[独立 beta 验收](reports/beta-0.1.0-2026-10-08.md)。正式版等待本人试用，新 beta Windows 包未在真机重跑，不将旧 RC 或源码运行代替新包验收。

## 当前范围与总状态

- F01～F14、A01～A17：已实现，当前普通功能与保护链路已验证（完整运行与修复后独立复验合并覆盖，不声称单轮全绿）；逐项证据与限制见下表及[本轮整合报告](reports/non-release-closeout-2026-10-08.md)。
- E01～E07：声明的平台子集与本次交付范围已有验证；不覆盖全部设备/网络/物理故障组合。
- E08：按用户接受约 100ms、有限尝试后收尾；原 50ms 目标的未达标记录仍保留。
- E09：构建、实际运行、开发交接、首个 beta 包验证及预发布完成；正式稳定版等待本人试用，Windows 新 beta 真机结果未验证。

[10-05 范围调整](reports/scope-review-2026-10-05.md)免除本次重复提供物理掉电/设备缓存损失、实际主机/NTFS/Engine 磁盘填满、自然配额驱逐、额外远程部署/高 RTT/证书运维、超过已完成 24h 的全部组合、额外 Windows 桌面/权限/容器/硬件组合。**免测不等于通过**：安全的已有 ENOSPC、事务中止、进程崩溃、权限/恢复守卫仍保留。

[Windows 普通设备 14 项](reports/windows-device-passed-2026-10-06.md)及此前 Job/ConPTY/visible WebKit 证据已完成；不再将它们列为待用户重新测试。Windows 防火墙仅支持 netsh 状态读取，变更/回滚不在声明支持子集；Linux firewalld 变更有私有环境证据，不能据此操作生产主机。

## 1. 范围与使用方法

本矩阵补充而不替代[技术架构](../plan/Blora-01-技术架构.md)和[功能与交互](../plan/Blora-02-功能与交互.md)的正文。表内描述是追踪入口，不是删减详细要求的依据。

F01～F14 追踪功能交付；A01～A17 沿用原规划的关键场景；E01～E09 补足运维、扩展和最终交付验证。实施块见 [ACTION_GUIDE.md](../../ACTION_GUIDE.md)。

### 1.1 本次必须完成

| 范围 | 解释 |
| --- | --- |
| M0～M3 | 同一开发任务内的依赖顺序；不是只做 MVP 的分期授权 |
| 默认应用 | 桌面宿主统一注册，主要功能以应用提供；实例中心、文件、编辑器、终端、任务、监控、节点、用户、设置和运维应用均有真实入口 |
| 多窗口与多标签 | 主要应用可多开；每个窗口可管理多个节点资源；标签可移窗；同资源可以有多个视图 |
| 实例快捷方式 | 独立资源入口；默认打开或激活专用窗口；可显式再开窗口 |
| Linux 与 Windows | 实际平台运行适配、文件、PTY 和停止控制；平台差异明确，不用空实现冒充支持 |
| 运维增强 | Docker/Compose、完整主机监控与进程视图、跨节点移动、备份恢复、调度、有限系统管理 |
| 扩展生态 | 公共 SDK、沙箱与受控能力、注册源/目录、可安装包、权限、升级迁移、禁用卸载和真实示例 |
| 刷新恢复 | 普通 F5/工具栏刷新保留完整已登记工作现场；最近输入和撤销重做属于核心能力 |

### 1.2 保留原规划边界

| 事项 | 本次处理方式 |
| --- | --- |
| 被管理服务的业务端口、穿透、VPN、游戏协议网关 | 用户独立配置；可以成为未来扩展应用，核心不实现业务中继 |
| 1Panel 的全部产品功能 | 只实现两份规划明确列出的管理能力，不无限扩张为网站托管、邮件等完整产品集合 |
| 公网商业应用市场运营 | 不要求运营公网服务、支付、审核团队；要求可工作的目录/注册源和包分发生命周期，本地测试源可验收 |
| 多 Master、高可用、外部数据库及额外部署形态 | 原规划中按需求增加的事项，当前保留合理模块边界，不强行纳入 |
| Windows 上所有发行版/Engine/系统功能 | 按平台能力声明和独立适配，不能把 Linux 命令视为 Windows 支持；缺少某 Engine 不应阻止其他功能 |
| 跨设备工作现场 | 恢复已同步副本；默认同步布局和引用，工作内容同步由用户开启；不保证未上传状态凭空出现 |
| 操作系统文件选择授权、密码和私钥字段 | 不伪造重新授权、不保存敏感字段；提供真实续传或重新选择流程 |
| 浏览器数据被清除、存储拒绝、磁盘或配额耗尽 | 明示保护失败并保留可导出数据，不声称任何条件下绝对不丢失 |
| 任意 TUI 无检查点重建、浏览器完整离线启动、操作系统级全局快捷键 | 不作不受浏览器与会话条件限制的承诺；正常会话恢复和桌面内快捷操作仍必须实现 |
| 视觉细节后续优化 | 可调颜色、间距、动画参数和文案；窗口模型、拖放、焦点、即时反馈不能留作细节待办 |

## 2. 状态和证据规则

实现状态与当前验证分开记录。“已实现”不等于未运行环境通过；源码回归不等于最终发行包已验证，交叉编译不等于 Windows 运行。每份报告保留时间、源码/产物身份、命令、环境、实际退出码和限制，当前结果不可回填到旧报告。

详细历史子项见归档及原独立报告；当前整合报告链接新的真实检查。未采用的渲染原型位于独立 opt-in 实验目录，其失败不删除、不移为成品通过。普通 Chromium 圆角守卫仍保留原八次稳定条件、整屏 RGBA0、全部场景和节点身份。

## 3. 功能交付 F01～F14

| 编号 | 必须交付的用户能力和边界 | 主要验收 | 实现 | 当前验证 / 证据 |
| --- | --- | --- | --- | --- |
| F01 | 桌面、任务栏、窗口移动缩放吸附/层级/焦点；应用多开、跨节点资源标签、标签移窗、快捷方式、菜单和键盘操作；可复制文本区域保留选择能力 | A01、A13、A15～A17 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[桌面与窗口](reports/rc13-current-source-2026-09-20.md)、[圆角修复](reports/chromium-corner-closeout-2026-10-08.md) |
| F02 | 布局、窗口/标签归属、正文、撤销重做、表单与视图现场、终端检查点；本地日志与快照、云端副本和版本迁移；账号/设备/浏览器标签隔离 | A01、A02、A06、A08、A11、A17 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[工作区](reports/workspaces-2026-09-10.md)、[同步冲突](reports/workspaces-2026-09-19.md)、[Worker 保护与回退](reports/non-release-reconciliation-review-2026-10-08.md) |
| F03 | 登录退出、多用户、角色与资源授权；查看/创建/控制/文件/终端输入/主机管理分权；服务端列表、统计、请求及流均授权 | A08、A14、E04、E07 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[审计与撤权](reports/audit-2026-09-10.md)、[授权拒绝](reports/browser-revoke-2026-09-13.md) |
| F04 | 节点登记、身份轮换吊销、管理连接、心跳、版本/能力协商、在线/异常/离线/维护状态、入口；只管理通信 | A05、A08～A10 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[管理与登记](reports/administration.md)、[节点失联](reports/node-offline-browser-2026-09-13.md) |
| F05 | 聚合全部获准实例、获准节点创建向导、搜索筛选统计/批量操作；多窗口多标签和专用资源窗口；真实启动停止重启与阶段诊断 | A03～A05、A10、A14～A17 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[实例视图](reports/instances-view-2026-09-22.md)、[实例权限](reports/instance-permissions-2026-09-13.md) |
| F06 | 实例控制台、PTY、只读观察/受权输入、多窗口标签/分栏、会话重新挂载、输出流控与检查点；主机 Shell 独立高权限 | A06、A08、A09、A17 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[真实 PTY](reports/terminal-programs-2026-09-13.md)、[恢复](reports/performance-continuous-2026-09-19.md) |
| F07 | 多选、重命名、新建、复制移动、回收/删除、压缩解压、上传下载、跨节点复制与移动；真实根目录隔离、冲突和传输续接 | A07、A08、A09、A12 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[传输](reports/transfers.md)、[关闭后续传](reports/upload-close-2026-09-13.md)、[文件视图](reports/files-view-2026-09-22.md) |
| F08 | 真实编辑器、语法能力、共享正文与独立视图、草稿/撤销重做恢复、显式保存、原子写入/版本冲突；大文件二进制有明确边界 | A01、A02、A08、A11、A17 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[编辑能力](reports/editor-capabilities-2026-09-22.md)、[共享编辑](reports/shared-editor-conflict-2026-09-13.md) |
| F09 | 持久化后台任务、阶段/进度/日志、等待、取消、重试和结果；站内通知与系统通知授权；关闭网页不取消服务端任务 | A04、A05、A07、A09、A10 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[任务阶段](reports/task-stages-2026-09-12.md)、[原生 Linux 通知](reports/native-notifications-2026-09-29.md) |
| F10 | 节点/实例指标、历史采样策略与旧数据标识、完整系统进程视图、诊断和授权终止；窗口最小化降低绘制开销 | E01、E08 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[监控](reports/monitoring-2026-09-09.md)、[离线历史](reports/monitor-history-soak-2026-09-21.md) |
| F11 | 容器、镜像、卷、网络和 Compose 项目；日志/终端、配置草稿、显式应用和分阶段结果；数据删除与容器删除区分 | E02 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[实际 Docker](reports/docker-current-2026-09-20.md)、[HTTPS](reports/docker-https-transport-2026-09-21.md)、[日志归档](reports/docker-log-history-2026-09-22.md) |
| F12 | 可恢复备份、目标/范围、保留策略与一致性钩子；平台调度、时区、错过执行/重叠策略、触发重授权和任务结果 | E03、E04 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[备份调度](reports/backup-scheduler-library.md)、[24h](reports/local-endurance-24h-2026-09-27.md) |
| F13 | 按平台能力提供有限系统服务、防火墙和主机计划任务管理；管理员专用、变更差异、管理连接保护及回滚 | E05 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[systemd/cgroup](reports/systemd-cgroup-2026-09-29.md)、[私有防火墙](reports/firewall-private-2026-09-19.md)、[能力门禁](reports/local-remaining-2026-10-05.md) |
| F14 | 默认与扩展应用共同宿主；SDK、状态合同、受控能力、沙箱、注册源和真实安装包；安装授权、版本兼容、升级、禁用、卸载 | A13、E06、E07 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[SDK](reports/sdk-2026-09-09.md)、[扩展迁移](reports/extensions-data-2026-09-20.md)、[受控网络](reports/netem-extension-task-2026-09-22.md) |

## 4. 原规划关键场景 A01～A17

| 编号 | 原验收场景合同 | 实现 | 当前验证 / 证据 |
| --- | --- | --- | --- |
| A01 | 输入、中文输入提交、粘贴、连续编辑和拖窗后立即 F5/工具栏刷新；正文、窗口布局和应用状态一致，撤销/重做可沿刷新前历史继续；分别验证活跃与后台窗口 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[立即刷新报告](reports/immediate-refresh-2026-09-13.md) |
| A02 | 同用户两窗口打开同文件，随后外部用户修改服务器版本；共享正文策略正确、局部光标/滚动独立，保存时检测版本冲突并保留各版本，不用旧响应覆盖新输入 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[共享编辑报告](reports/shared-editor-conflict-2026-09-13.md) |
| A03 | 真实实例忽略正常停止、派生后代、主 PID 提前退出或后代持有管道；有限等待后升级或明确失败，全局仍可用，归属运行未退出不启动新代次 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[平台报告](reports/systemd-cgroup-2026-09-29.md)、[停止边界报告](reports/stop-boundary-2026-09-13.md) |
| A04 | 重复重启请求、服务端已接受但客户端丢响应、Master 中途重启；同 requestId 对账复用结果，同资源不同请求按策略串行，不重复产生运行 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[重启报告](reports/restart-disconnect-2026-09-13.md) |
| A05 | 重启和传输途中断开节点连接并恢复；标为失联/待确认，不伪造停止或成功；按实际执行记录校准，不因重连重复副作用 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[失联与恢复证据](reports/node-offline-browser-2026-09-13.md)、[重启断线](reports/restart-disconnect-2026-09-13.md) |
| A06 | 在实际 PTY 中运行 vim/top 等全屏程序、中文/ANSI/光标模式后刷新；检查点和输出序号衔接正确，继续使用原会话，回放绝不重发历史输入 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[终端程序报告](reports/terminal-programs-2026-09-13.md) |
| A07 | 同时发起节点端解压和来自浏览器本地文件的上传，再关闭网页；解压独立完成，上传保留已确认检查点并在需要时等待重新选择源文件，状态与真实结果一致 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[完整场景证据](reports/upload-close-2026-09-13.md) |
| A08 | 使用过程中撤销查看、写入或终端能力；新请求拒绝、现有流及时终止/降权；切换账号不能访问前账号草稿；旧快照不能恢复授权 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[撤权报告](reports/browser-revoke-2026-09-13.md) |
| A09 | 大日志、慢接收端与批量文件传输并行；控制请求持续取得服务，队列、内存、归档磁盘有可测上限，背压/丢弃策略可见，无无限累积 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[30秒慢消费者报告](reports/slow-consumer-soak-2026-09-21.md)、[最终报告](reports/local-endurance-24h-2026-09-27.md)、[混合流报告](reports/mixed-stream-control-2026-09-13.md) |
| A10 | Daemon 崩溃后重启，包含已启动进程与中途任务；按运行身份而非仅 PID 对账，既有进程/任务结局明确，不可恢复的操作标为中断 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[Daemon崩溃报告](reports/daemon-crash-2026-09-13.md) |
| A11 | 注入浏览器配额写入失败并迁移旧恢复记录；不清空草稿、不显示虚假保护成功；快照完整切换，迁移失败可保留/导出旧记录 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[恢复故障报告](reports/recovery-quota-2026-09-13.md)、[跨引擎恢复报告](reports/recovery-engines-2026-09-26.md) |
| A12 | 跨节点复制/移动遇到目标磁盘满、断线、来源变化；记录可诊断阶段和清理对象，未验证目标前不删除源；源改变后不错误删除新内容 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[空间不足报告](reports/enospc-2026-09-19.md)、[传输报告](reports/transfers.md) |
| A13 | 默认应用和参考扩展通过同一 AppHost 能力合同注册；无需改桌面核心可开窗口/标签、注册任务并恢复现场；扩展不能获得未授予能力 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[扩展全链路报告](reports/extensions-real-2026-09-13.md)、[跨设备报告](reports/extensions-cross-device-2026-09-14.md)、[未授权浏览器报告](reports/extensions-unauthorized-browser-2026-09-14.md) |
| A14 | 两节点多实例、查看与创建权限不同；实例中心聚合全部可见资源，筛选/统计不泄漏；看见节点不等于可创建，提交再次校验权限/配额 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[实例权限报告](reports/instance-permissions-2026-09-13.md) |
| A15 | 两个实例中心窗口包含多个跨节点实例标签，重复查看同一实例；真实运行状态同步，视图导航独立；两视图并发控制仍由后端统一去重/串行 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[多视图报告](reports/instance-multiview-2026-09-13.md)、[运行适配](reports/runtime-2026-09-09.md) |
| A16 | 添加实例到桌面，单击启动入口、显式新开、改入口名称、删除入口、目标改名/删除/离线/权限撤销；默认打开专用管理窗口且可复用已有专用窗口，不创建/启动/删除实例，不用同名资源替代 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[入口报告](reports/shortcut-lifecycle-2026-09-13.md) |
| A17 | 把含未保存草稿或活跃终端的标签拖出新窗口，再移入兼容窗口并立即刷新；原视图 ID、归属、顺序、正文、撤销历史与会话引用保留，不重复视图/任务/输入 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[拖拽报告](reports/drag-recovery-2026-09-13.md) |

## 5. 补充验收 E01～E09

| 编号 | 原验收场景合同 | 实现 | 当前验证 / 证据 |
| --- | --- | --- | --- |
| E01 | 主机与实例指标来自真实节点；系统进程搜索/排序/详情及授权操作可用，终止只命中指定测试进程；失联标明采样时间，不把旧值当实时；采样和历史保留有边界 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[监控](reports/monitoring-2026-09-09.md)、[离线历史](reports/monitor-history-soak-2026-09-21.md) |
| E02 | 在真实 Docker/Compose 测试环境创建项目、应用配置、查看日志/终端、更新和删除；镜像/卷/网络操作可追踪；镜像拉取失败/依赖缺失/部分成功明确；删容器默认不误删卷数据，配置草稿刷新保留 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[实际 Docker](reports/docker-current-2026-09-20.md)、[HTTPS](reports/docker-https-transport-2026-09-21.md)、[日志归档](reports/docker-log-history-2026-09-22.md) |
| E03 | 创建有可核对内容的备份并恢复到受控目标，验证文件内容/元信息和权限；保留策略只清理自身范围；一致性钩子成功/失败分别呈现；中断、空间不足、恢复覆盖均有明确处理 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[备份调度](reports/backup-scheduler-library.md)、[24h](reports/local-endurance-24h-2026-09-27.md)、[ENOSPC](reports/enospc-2026-09-19.md) |
| E04 | 时区含夏令时、节点离线、Master 重启、错过触发和同资源重叠；按声明策略执行且不重复补跑；创建者失去权限后新触发拒绝；节点离线仅继续已接受任务，不从陈旧计划无限产生特权工作 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[调度进程中断](reports/scheduler-staged-process-crash-2026-09-21.md)、[24h](reports/local-endurance-24h-2026-09-27.md) |
| E05 | 在独立且获准的测试环境操作支持的服务/计划任务/防火墙子集；只修改目标对象，展示差异、保留既有规则；模拟管理链路失联后按期限自动回滚，再连接可诊断；普通用户不能调用主机特权 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[systemd/cgroup](reports/systemd-cgroup-2026-09-29.md)、[私有防火墙](reports/firewall-private-2026-09-19.md)、[能力门禁](reports/local-remaining-2026-10-05.md) |
| E06 | 按 SDK 文档从独立示例目录构建新扩展；不修改桌面核心即可安装、开多窗口/标签、注册资源入口和后台任务、保存迁移现场；前后端沙箱与能力桥接实际拒绝未授权访问 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[SDK](reports/sdk-2026-09-09.md)、[扩展迁移](reports/extensions-data-2026-09-20.md)、[受控网络](reports/netem-extension-task-2026-09-22.md) |
| E07 | 真实测试注册源提供至少两个版本扩展；浏览/获取/验证/授权/安装/启用/禁用/升级/失败回退/卸载跑通；篡改包、版本不兼容、能力提升均按策略阻止；禁用不再接受新任务，卸载默认保留用户数据并给出清理选项 | 已实现 | 本次范围已验证；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[根证书轮换](reports/remote-catalog-root-rotation-2026-09-21.md)、[扩展迁移](reports/extensions-data-2026-09-20.md) |
| E08 | 记录参考设备、浏览器、RTT、日志速率和带宽；8个窗口、2个活跃终端、万条目录与后台传输并行；原规划p95 ≤50ms独立保留，未达标如实记录；2026-10-08用户接受约100ms并明确要求有界尝试后收尾 | 按用户调整范围收尾 | Chromium 原 50ms PASS；Firefox 66ms、WebKit 89–132ms 原断言 FAIL，最后 95ms。不保证每轮 ≤100ms。[本轮整合](reports/non-release-closeout-2026-10-08.md)；[原性能结果](reports/e08-bounded-closeout-2026-10-08.md) |
| E09 | 从干净依赖环境按说明构建并初始化 Master/Daemon/前端；至少两个节点和不同权限用户跑通；交付 Linux 与 Windows 产物/构建入口，分别记录 Linux 运行、Windows Job/ConPTY 真机结果；验证配置/数据迁移、备份恢复、升级回退和示例扩展开发说明可复现 | beta 交付完成 / 正式版待试用 | 六包及完整源码、清单/摘要和三份 Web 一致；新包初始化/双节点/权限、SDK 构建签名、自身停机恢复、RC34 升级和兼容快照回退通过，预发布九附件核对一致。新 beta Windows 包真机未重跑。[beta 验收](reports/beta-0.1.0-2026-10-08.md)；[本轮整合](reports/non-release-closeout-2026-10-08.md)；[原 Windows 14 项](reports/windows-device-passed-2026-10-06.md) |

## 6. 横向合同

- 默认应用与扩展都服从账号/节点/资源授权；查看、创建、输入、主机管理分别检查，已开流可撤销。
- appId、windowId、viewTabId、resourceRef 等身份独立；关闭窗口不停止后端资源，同资源状态共享而视图现场独立。
- 输入与拖动及时反馈；远程结果未确定不显示成功，恢复不能重放远程副作用。
- 正文、撤销重做、布局、标签、表单和终端检查点即时保护；真正保护失败须可诊断/导出，ACK 不能绕过保护。
- 任务按资源串行、去重并持久化；旧运行未确认退出不重复启动；移动未确认目标前不删源。
- 管理连接不接管服务业务网络；第三方扩展不默认获得主机执行权。
- 凭据、真实正文、数据库、私有日志和浏览器快照不公开；只清理本次自建且可识别的测试对象。

执行顺序和接管见[当前进度](../execution/PROGRESS.md)，真实复现见[本地回归指南](../operations/LOCAL_REGRESSION.md)和[平台指南](../operations/PLATFORM_VALIDATION.md)。beta 已发布，正式稳定版仍等待用户实际体验；源码提交、交叉编译及旧设备结果不能改写未验证环境。
