# Blora Panel 执行进度与续接记录

2026-09-29 本地完成度复核：原先写作“运行中”的独立24h长测已从私有原始证据重新校验为 **PASSED_24H**，且[最终报告](../acceptance/reports/local-endurance-24h-2026-09-27.md)与原始终态完全一致；实际86,415.480秒、17,262次采样、1,440个分钟槽、两次故障恢复、末尾真实备份恢复与所属资源清理通过。已同步[验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)及[非UI剩余工作](NON_UI_REMAINING_2026-09-26.md)，旧“RUNNING”条目仅为当时过程记录。本轮未改产品源码、未重跑已通过的功能或发行检查，当前仓库起点为 `6f48131`；无长测/fixture会话需接管。Linux本地功能51/51、mock56/56及RC24发行证据保留；严格E08当前UI三引擎p95 1240.1/1230/739ms，均高于≤50ms目标，本机无`/dev/dri`，不将旧UI或软件合成对照视为通过。下一具体动作：仅在能区分绘制热点且保持用户选定UI的条件下继续源码级性能优化并定向复验；取得可用硬件加速及Windows/systemd/跨主机/物理故障等隔离环境后执行对应真实验证。全范围仍未完成。

2026-09-28 README双语与首页排版：按用户要求将默认 `README.md` 改为英文，新增 `README.zh-CN.md` 双向语言跳转；使用现有品牌几何制作静态标志，增加技术徽章、随深浅色切换的真实界面截图、功能表、快速开始、架构与文档导航，高级配置/验收折叠保留。中文版保留原运行指南；其余工程文档仍为中文，英文页已明确说明。仅将两张既有演示PNG纳入Git，其他历史图库继续忽略，未使用图片生成或修改产品UI。全部本地链接存在，17个Shell示例与4个JSON示例语法校验、GitHub原生Markdown表格/picture/details渲染通过；实际GitHub匿名Chromium验证英文浅/深、中文版、390px手机四场景，3200px原图加载、语言链接及折叠比较均通过。首次深色图片等待超时、手机页面水合期间节点替换与整页load等待超时均保留为检查入口失败，改为等待实际README内容/图片后定向复核，未改变产品或验收阈值。首页更新已推送 `main`，提交 `43d80ea`；本轮是文档交付，不提升任何功能/性能/平台验收状态。

2026-09-27 GitHub 源码交付：按用户明确授权，将项目上传至公开仓库 https://github.com/BloretCrew/BloraPanel 的 `main` 分支，首次提交 `212a5ddc7ada7195c1b96d5206bca63eaf4c49d9`。共1,137个源码、测试、文档、锁文件及代码使用的图标资产，远端逐项核对路径、文件模式与blob摘要一致。完善 `.gitignore`，排除凭据、运行数据库、依赖、发行归档、生成的参考扩展包和历史截图；这些本地文件没有删除。README补齐首次克隆后的 `make sdk` 步骤，三个参考包可从源码生成。仓库Description由用户自行填写。产品UI及验收状态不变，独立24h长测和只读收取器继续运行；下一动作仍为收取实际24h终态并保留性能/外部平台验证缺口。

2026-09-27 11:14 长测18h阶段核对通过：第2次故障03:13:25 UTC完成；所属Daemon SIGSTOP后明确OFFLINE，Master重启前03:13:20/重启后03:13:21均核对原120点stale历史，同一Daemon恢复后原runId `7d43fbe2017b5c6b1aa1e017a24eb45c`仍RUNNING。最新Master为PID1372175/startTicks16367598、Daemon598122/startTicks9882575、worker597961/startTicks9882482，同bootId；任何接管仍应读最新checkpoint确认身份。当前实际64,869s/12,956采样/1,081成功slot，两个声明故障阶段完成，RSS Master/Daemon34.04/34.01MB、FD14/19、归档16,776,441B、状态/证据峰值205,456,312B、普通采样gap6.067s，继续使用原预算，不提前标24h通过。独立收取器1366399仍运行，将自动更新[终态报告](../acceptance/reports/local-endurance-24h-2026-09-27.md)；预计17:12后由长测worker执行实际恢复与所属资源清理，收取器核验后结束。RC24与真实51项/完整mock56项功能证据已闭合；严格E08本机完整材质仍未达标，软件后端比较无足够收益，Windows等外部环境仍未验证。新收取入口不改变本次不可变二进制或产品UI。下一动作：收取24h实际终态及边界，保留性能/外部平台缺口；无需为结果收取重新启动服务或要求用户回复“继续”。

2026-09-27 11:08 结果自动收取：新增只读 `scripts/endurance-report.py`，当前真实短测终态正确识别为PASSED_SHORT；旧短测缺新清理字段明确拒绝，短测伪标24h/缺实际恢复/遗留服务/私有字段不导出四项拒绝边界验证通过。独立收取器PID1366399/startTicks16319189、bootId同长测，03:05:16 UTC启动，持续更新[24h动态报告](../acceptance/reports/local-endurance-24h-2026-09-27.md)并在实际终态结束，不需要用户回复“继续”；日志和身份 `.local/evidence/rc24/endurance-collector.*`。报告不是预先通过声明，超时/检查点停更也以未完成退出且不改服务。最新长测17h54m、12,883采样、1,074成功slot，状态RUNNING，18h阶段约11:12，24h终态17:12后；实际disk峰值204,685,379B仍在原256MiB预算，SQLite任务/事件累积继续真实计入预算，未调整上限或删除历史。新无终端/存储活动的八玻璃微探针：SwiftShader p95 400ms、明确Mesa llvmpipe 433.3ms、Vulkan仍SwiftShader383.3ms，非E08验收，未采纳对照；Xvfb包装器未进入探针，直接带所属显示认证后的诊断37.830s完成；唯一容器已停止，包装器137不算通过。产品UI和RC24产物未变化。下一动作：核对18h故障后原runId与120点历史，持续保留24h真实终态及严格E08未达标；Windows等外部环境仍未验证。

2026-09-27 10:57 RC24发行闭合：六包构建退出0（49.479s），外部SHA256和2,191条包内路径/模式/长度/SHA256记录全部通过，两个Master及独立Web包精确匹配当前浏览器构建。独立SDK/参考包签名、Master/TLS/双Daemon、停机快照恢复及兼容RC23实际回退 **退出0（15.359s）**，所属进程已停止，见[RC24报告](../acceptance/reports/release-rc24-2026-09-27.md)。初次打包因headless包缺LICENSE退出2，现严格绑定同源commit/MIT/原文hash补齐通知；失败记录保留。当前56项完整mock、另Worker终止1项、完整真实51项及分段Go race通过，不以这些功能结果覆盖E08。完整材质的禁用GPU启动参数诊断仍p95 1005ms、退出1，最终WebGL仍SwiftShader，不算得到独立CPU后端；产品未改。接下来仅私有隔离容器比较已有图形后端是否能改善合成；probe exec77336，不访问宿主设备/网络/凭据。独立24h `.local/endurance-zs_dzpzj` 最新elapsed63,438s/12,679采样/1,057成功slot，RUNNING；下一18h故障约11:12，24h终态约17:12，勿停止。正式服务和生产节点未触碰，Windows/systemd等外部平台仍未验证。

2026-09-27 10:43 本地完整功能回归闭合：当前56项全mock浏览器通过；最终原生鼠标hover命中探针后的两条扩展链路各重复3次共6/6通过，真实完整51项 **全部通过（809.066s，退出0，无skip）**，含121点监控/Master重启、两个真实分钟调度slot、真实Docker/Compose/扩展WASI、PTY全屏/刷新/撤权/移窗与16MiB原上传续接。新增真实终止运行中的检查点Worker后等待原生产30秒截止、中文UTF-8分块/屏幕保护/重挂不重放输入 **1/1通过（35.553s）**。详见[本地功能报告](../acceptance/reports/local-functional-2026-09-27.md)。同材质独立绘制伪元素性能诊断仍p95 1003.8ms、退出1，无足够收益，不进入产品；当前三引擎E08失败继续保留。真实fixture9444 exec36317已Ctrl+C退出0，Vite5173 exec50786已停止，新Worker测试自管Vite也已退出；性能容器已`--rm`清理。独立24h监督 `.local/endurance-zs_dzpzj` 不停止，最新status仍RUNNING。下一动作：RC24六包构建、全部外部/包内摘要校验及独立停机恢复/兼容RC23回退，更新发行台账；Windows/systemd等实际运行仍未验证。

2026-09-27 10:30 功能复验进展：当前56项完整mock浏览器回归全部通过（365.482s、退出0）；绘制就绪两次RAF后仍使用可信鼠标，参考包实际WASI/丢202回执同幂等键恢复/快捷入口/移窗关闭及成员权限2/2（31.694s、退出0）通过。临时指针事件打印、body状态诊断与样式对照已删除，未合成click或重试副作用。首轮真实参考包未包裹绘制等待的开窗点击实际复现失败后修正；登录前一次ERR_CERT_VERIFIER_CHANGED单独保留。两轮短暂启动的完整套件分别因需独立性能入口及fixture已有备份导致隔离不足而中断，退出130，不能计通过；旧fixture均Ctrl+C退出0。现在唯一完整真实功能套件为exec52431（51项，性能严格另验），fixture9444 exec36317，私有凭据 `/tmp/fixture-1296441527/browser-credentials.json` 勿输出内容；Vite仍5173 exec50786。独立24h长测保持运行。见[本地收尾结果与失败记录](../acceptance/reports/local-functional-2026-09-27.md)。下一动作：收取51项终态，继续非视觉合成成本对照（仅测试侧），通过后停止所属fixture并重新打包RC24、验证独立恢复与RC23回退；E08已知失败不被mock17ms覆盖。

2026-09-27 10:20 本地收尾更新：保持 UI 冻结，修复 Compose 保存流程先解析参数后授权导致无 requestId 的未授权请求返回400，现先查权限、执行时重查；真实 Docker 分阶段保存/Daemon重启/显式应用及未授权回归通过。修复按焦点顺序重排窗口 DOM 会重新加载扩展 iframe、丢本地输入的问题，改为稳定创建顺序和 CSS 层叠；新增实际输入/三次焦点切换回归通过。修复旧测试主题选择器、保存回执对账替身、异步重新读取正文确认后刷新时序，以及监控 fixture `--listen` PID 边界识别，未放宽产品合同。全 Go race 首轮其余包通过、Master两项失败；独立真实PTY重复10次通过，修复后完整Master race **通过（435.868s）**。实际浏览器51项首轮42通过9失败，修复后的12项定向复验 **10通过2失败**：备份、容器指标、扩展取消/通知/名称保存/迁移/目录、监控121点历史与Master重启、实例设置和编辑器均通过；扩展参考包鼠标点击及其清理失败引起的成员409仍在定位，不能据定向通过声称完整回归通过。新版前端mock25项22通过3失败，焦点新增项通过；现用系统和缓存浏览器对照确认剩余 iframe 真实点击未进入子文档，移除诊断前需取得复验。当前服务：Vite5173 exec50786；真实fixture9444 exec66125，私有凭据路径 `/tmp/fixture-2906434352/browser-credentials.json`（勿输出内容），只能清理所属fixture。独立24h长测 `.local/endurance-zs_dzpzj` 已17h04m、1025成功分钟slot、12287采样，仍RUNNING；预计17:12后收取，勿随构建/fixture清理停止。各本轮命令实际终态保存在私有 `.local/evidence/rc24/*.log.result.json`。下一动作：定位 iframe 命中/点击问题，完成全套前端与真实浏览器回归，再RC24打包和独立恢复/RC23回退；严格E08三引擎失败及Windows等平台未验证继续保留。

2026-09-27 09:43 续接本地收尾：独立长测已连续16h24m、跨UTC日期，984个分钟备份slot、11805采样，6h失联/同库Master重启后保持原runId；仍RUNNING，预计17:12后收取，不能提前算通过。当前Master已由长测监督器重启为1020809，身份必须使用status读取。昨日tool会话失效，未取得完整回归终态，不据文件时间戳计作通过。今日加入私有磁盘日志/真实退出码并重跑：`make check build windows`退出0（4.90s）、前端63/63、干净SDK与三个参考包退出0。前端55项首轮受5173服务中断影响已取消，持有独立Vite后复验；发现旧`主题`测试定位匹配两个控件，修正选择器，结束后定向复验。Go首轮缺少显式exec-helper测试配置，取消后以新构建helper重跑全套race（exec48147）。真实51项串行浏览器已启用监控121点soak（exec46948），fixture为9444、`/tmp/fixture-172569861`、exec16295；Vite exec50786、前端exec26075。昨日性能当前完整材质基线三引擎均失败：Chromium1240.1ms、Firefox1230ms、WebKit739ms；测试侧去模糊/分层对照不进入产品，见[当前性能报告](../acceptance/reports/performance-current-2026-09-26.md)。下一动作：修复回归发现并复验，完成RC24打包与独立恢复/回退冒烟，继续保留严格E08失败和24h进行中。

2026-09-26 本地收尾更新：Firefox/WebKit 恢复故障各 3/3 通过；新增真实墙钟监督入口，两轮完整五分钟短测及 stop/resume/监督中断安全清理通过，见[长测入口报告](../acceptance/reports/local-endurance-2026-09-26.md)。独立 24h 运行 `.local/endurance-zs_dzpzj` 已从北京时间 9/26 17:12:32 开始，预计 9/27 17:12 后收取终态；必须用 `python3 scripts/local-endurance.py status .local/endurance-zs_dzpzj` 读取最新身份，不能在运行中记为通过。终端序列化 Worker 已实现，真实 xterm 一致性/刷新/移窗/不可用回退浏览器 4/4 通过；单独迁移的 Firefox 混合负载 p95 1227ms，仍失败。新增断言实际复现终端日志数组嵌套 Vue 代理造成 `DataCloneError`，修正为校验后原位追加，63/63 前端单测及生产构建通过；严格负载正在复验。UI、负载、恢复 ACK 顺序及 ≤50ms 阈值保持不变。性能 fixture 仍为 exec 48529、9443、`/tmp/fixture-872806179`；Vite 测试 exec 34771、5173；只 Ctrl+C 清理本任务服务，独立长测保留。下一动作：收取性能结果，必要时继续定位，再完成当前全套回归、RC24 构建及独立发行恢复/回退冒烟。

2026-09-26 本地收尾执行中：用户已授权完成本地可推进工作，保持 UI 冻结。Firefox 155.0 和 WebKit 26.6 的恢复故障三项各 3/3 通过，见[跨引擎报告](../acceptance/reports/recovery-engines-2026-09-26.md)，不代表真实物理掉电或浏览器实际配额阈值通过。当前冻结 UI 的 Firefox 严格双 PTY 60 秒基线 p95 1519ms，失败证据位于 `/tmp/blora-perf-firefox-baseline-20260926`；开始仅测试侧窗口合成对照，输出负载、输入时序及 ≤50ms 阈值不变。根代理临时性能 fixture：127.0.0.1:9443，exec 会话 48529，凭据文件路径 `/tmp/fixture-872806179/browser-credentials.json`（不得输出内容），仅 Ctrl+C 清理所属 fixture。另有独立监督长测入口正在准备，先短时真实墙钟验证，再启动 24 小时跨日运行；未跑完前不得计作通过。下一动作：比较玻璃合成/终端隔离绘制开销，完成必要修复，然后当前版本真实回归、重新打包及独立发行冒烟。

2026-09-26 UI 暂停与非 UI 剩余工作盘点：用户明确要求 UI 先优化到这里，本轮只读审计源码/台账/最新报告，未启动测试或实施功能修改。主要功能已有代码与 Linux 本地真实证据；明确仍未解决的质量问题是 WebKit 严格双 PTY p95 78ms、Firefox 144ms，超过 ≤50ms。本地仍可推进性能热点定位、冻结版本真实全套回归与重新发行：最近完整回归为 RC15，最近发行 RC23 的包内 index.html 与当前 web/dist 摘要不同。其余严格验证缺口包括 24 小时/更长期边界、Windows 真机、独立 systemd、跨主机 Engine/注册源/网络、设备断电与系统通知中心真实呈现，详见[剩余工作核对](NON_UI_REMAINING_2026-09-26.md)。没有把已有 ENOSPC、进程崩溃、权限撤销或恢复通过项重新列为功能缺失；整体验收状态保持原台账。本轮无运行会话需接管。下一动作：停止视觉发散，非 UI 工作优先定位非 Chromium 绘制性能，再做冻结版本回归与发行收尾。

2026-09-26 全配色深浅色截图交付：按用户要求重新拍摄当前六套配色（冰蓝、青灰、暖砂、苔绿、雾紫、烟粉），每套覆盖浅/深色、通透开/关和桌面/启动器/实例中心三个场景，共 **72 张 3200×2000 原图**。使用真实 Vue 设置选择外观、24 个隔离浏览器 context 和本地 GET-only 四实例/双节点数据；未改产品源码。资源 SHA256 核对当前暖砂深色壁纸及新版浅/深备份图标，材质开关保持壁纸/图标一致，页面错误、API 写入和外网请求均为 0，见[验证记录](../../web/ui-screenshots/all-appearance-20260926/verification.json)。生成每配色/材质一张深浅并排的三场景图板，向远程用户直接嵌入 PNG；[72 张原图 ZIP](../../web/ui-screenshots/all-appearance-20260926/blora-all-appearance-originals-20260926.zip)已校验 72 项及 CRC。捕获脚本为 `web/.local/capture-all-appearance-20260926.mjs`，临时 Vite/Chromium 已退出。本轮只提供视觉复查，不改变产品/外部环境验收状态。下一动作：依据用户对完整配色与场景的反馈进行定向调整。

2026-09-25 暖砂深色壁纸中段留白修订：用户指出浅色版顺眼，但深色版上方约 70% 是一整块近黑底。定位到 `sand-dark.svg` 的 `#1b1b1a` 全幅底色和仅从 y≈700 开始的下方曲面，无额外 CSS 蒙层。本轮只调整深色暖砂 SVG：底色改为清晰暖炭色，加入低对比的上方宽曲面，并让下方沙丘提前进入画面，保持纯色几何、不加渐变；浅色壁纸、主题/材质切换及应用图标未改。真实 Vue 以隔离 GET 数据重拍浅色/深色和深色通透 `on/off` 共 **6 张**完整桌面/启动器，加一张[对照总览](../../web/ui-screenshots/sand-wallpaper-rework/00-overview.png)；[验证记录](../../web/ui-screenshots/sand-wallpaper-rework/verification.json)确认浅/深资源不同、材质开关不改壁纸/图标、零页面错误/API 写入/外网请求。`cd web && npm run build`（含类型检查）通过，暖砂深色/自选壁纸聚焦 Playwright **1/1** 通过；此 UI 微调不提升全范围验收状态。下一动作：根据用户对新版深色实景的观感继续定向调整。

2026-09-25 备份与计划图标收口：用户指出旧深色图标左侧回转缺口切断外圈、浅色钟面越过内径，观感像白色溢出。保持现有 01A 底板和浅/深配色体系，前景改为闭合圆形表盘、内收钟面与简洁指针；浅色外环单独拉开明度，更新共享矢量几何、默认/正式浅色/03 深色三份预渲染资源。`cd web && npm run build`（含类型检查）通过；主题/通透与深色图标切换 Playwright **2/2** 通过。真实 Vue 在暖砂浅色/深色、通透 `on` 下拍摄桌面/启动器与桌面 58px、Dock 44px、启动器 55px 实际尺寸共 **11 张**，确认三入口均引用更新资源、零页面错误/API 写入/外网请求；见[实际尺寸对照](../../web/ui-screenshots/backup-icon-final/00-comparison.png)、[验证记录](../../web/ui-screenshots/backup-icon-final/verification.json)。本轮仅改图标，不改变其他主题、壁纸或全范围验收状态。下一动作：根据用户对新图标实景的反馈继续微调。

2026-09-24 暖砂深色壁纸再次修订与远程截图交付：用户明确否定上轮青蓝壁纸，并指出远程对话无法打开工作机本地 HTML。比较真实桌面中的石墨、莓红、赤陶、香槟砂候选后，将正式 `sand-dark.svg` 改成近黑底与浅米金纯色曲面，保留既有几何和细轮廓，不用渐变或伪光影；其他五套配色、03 深色图标、浅色壁纸和通透开关未改。`cd web && npm run build` 通过；主题、图标、壁纸独立性 Playwright **3/3** 通过。六套深色完整桌面/启动器实景均已重拍，16:10、2× DPR、本地 GET-only fixture；[暖砂桌面](../../web/ui-screenshots/dark-wallpaper-final/sand-dark-desktop.png)、[暖砂启动器](../../web/ui-screenshots/dark-wallpaper-final/sand-dark-launcher.png)、[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)。六壁纸互异、通透开关不改变壁纸，零页面错误/API 写入/外网请求。向用户交付时直接在对话嵌入六张完整 PNG，不再仅给本地 HTML 链接。下一步根据用户对米金版及六配色的视觉反馈继续定向改进；全范围验收状态不因本轮 UI 截图而改变。

2026-09-24 暖砂深色壁纸配色修订：用户认为上轮可见度增强后的灰棕色块浑浊难看。用真实 Vue 桌面分别对照冷蓝、冷灰蓝、深蓝绿三种保留原曲线的方案，选深蓝绿替换正式 `sand-dark.svg`；去掉灰棕色及半透明灰叠色，底部曲面采用明确的青绿层次，保留一条细轮廓线。没有改动布局、03 默认深色图标、浅色壁纸、主题配色或通透开关。`cd web && npm run build` 通过；主题/图标/壁纸独立性 Playwright **3/3** 通过。真实 Vue 只读演示数据重拍[暖砂深色桌面](../../web/ui-screenshots/dark-wallpaper-final/sand-dark-desktop.png)、[启动器](../../web/ui-screenshots/dark-wallpaper-final/sand-dark-launcher.png)及[六配色总览](../../web/ui-screenshots/dark-wallpaper-final/index.html)，通透开关不改变壁纸、六壁纸互异、全部加载 03 图标、零页面错误/API 写入/外网请求；[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)。图标方向对照图库也已按新正式壁纸重拍。下一步依据用户对新实景的观感继续定向微调；本次仅属 UI 改善，不改变全范围验收状态。

2026-09-24 深色壁纸可见度修复：用户指出暖砂深色桌面几乎纯色；定位到原浅色 SVG 前的 `#11171ce6` 遮罩（约 90% 不透明）把色块差压到约 1–4 RGB。保留六套主题原有几何，为冰蓝、青灰、暖砂、苔绿、雾紫、烟粉各新增深色专用 SVG，去掉这层遮罩；浅色壁纸、03 默认暗色图标和通透开关不变。青灰主题的暗色默认壁纸规则同时限定在“随配色”，不再盖掉用户自选壁纸。`cd web && npm run build`（含类型检查）通过；主题、壁纸选择、03 图标与通透独立性相关 Playwright **4/4** 通过。真实 Vue 在本地只读 fixture 下重拍六配色暗色桌面/启动器及暖砂浅色/实色共 14 张原图，六壁纸互异、通透 on/off 背景完全一致、六组均加载 14 枚 03 图标、零页面错误/API 写入/外网请求；见[总览](../../web/ui-screenshots/dark-wallpaper-final/00-overview.png)、[全图索引](../../web/ui-screenshots/dark-wallpaper-final/index.html)与[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)。既有 03 图标七向对照图库也已重拍成新壁纸状态。下一步继续根据用户对暗色壁纸亮度和配色的实际观感定向微调；本 UI 修复不扩大全范围验收结论。

2026-09-24 深色图标方向定稿：用户从六套预览中选定 **03「清晰造型」**。深色模式默认资源已切为 `h04-spectrum-night-clear-forms` 的 14 枚图标；浅色资源、六套主题配色与通透开关未变。旧深色版作为仅开发对照继续可复现；24 场景截图脚本的默认深色资源断言已同步。`cd web && npm run build`（含类型检查）通过，主题/通透独立性与深色图标切换 Playwright **2/2** 通过；七向真实截图重拍并确认 7×14 图标加载、相同布局、零页面错误与 API 写入，见[预览图库](../../web/ui-screenshots/dark-icon-directions/index.html)和[验证记录](../../web/ui-screenshots/dark-icon-directions/verification.json)。下一步在后续 UI 反馈中继续检查其他深色组件；当前所选图标已为正式默认。

2026-09-24 深色图标六方向预览：针对用户指出的 Dock 图标偏暗、辨识度不足，保留现状作对照，另备色彩提亮、柔亮底板、清晰造型、双色强调、暖色材质、亮色中心六套开发预览；**默认图标未切换**。[七向对比页](../../web/ui-screenshots/dark-icon-directions/index.html)、[Dock 55px 总览](../../web/ui-screenshots/dark-icon-directions/00-comparison-dock.png)与[启动器总览](../../web/ui-screenshots/dark-icon-directions/00-comparison-panel.png)包含每向完整桌面、Dock 局部和启动器面板实拍。真实 Vue 深色/暖砂/通透开场景使用隔离 GET 数据：每向 14 张 288px 图标均加载，七向布局一致，零页面错误、零 API 写入和外网请求；详见[验证记录](../../web/ui-screenshots/dark-icon-directions/verification.json)。`cd web && npm run build` 通过，主题配色与通透独立性、深色图标切换两项 Playwright 测试 **2/2** 通过。下一步由用户选择偏好的方向，再据此细化并决定是否切换默认方案。

2026-09-24 深色通透与亮边复核：用户所示启动器截图确为通透模式 `on`，但此前深色面材质遮盖约 96–97%，因此视觉上接近实色。本轮将深色 `on` 的启动器/菜单与顶栏调整为约 80% 遮盖、窗口约 89%、Dock 约 78%，缩小模糊半径并收减内侧反射与投影；深色窗口、启动器和 Dock 外圈使用暗壳边，启动器底部分隔线同步压暗。14 枚深色应用图标继续逐个调色，并降低仅深色 h04 导出中的反射与壳层亮度，浅色资源及六套主题配色保持独立。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[六配色×深浅×六场景通透 `on` 图库](../../web/ui-screenshots/dark-mode-review/index.html)有 **72 张**实景，另拍暖砂 `off` **12 张**对照，可直接查看[暖砂深色启动器 `on`](../../web/ui-screenshots/dark-mode-review/sand-on/08-dark-launcher.png)与[`off`](../../web/ui-screenshots/dark-mode-review/sand-off/08-dark-launcher.png)。本地隔离 fixture 记录零页面错误、零 API 写入；这些视觉证据不替代真实节点、后端或外部平台验收。下一动作：根据用户对通透度、暗边与图标亮度的实景反馈继续定向微调，保留全范围验收矩阵中的未验证项。

2026-09-24 深色窗口边界与图标亮度收口：按用户指出的实例中心截图，深色窗口外框由带主题色的明显描边改为中性细线（普通/焦点分别为 5%/7% 前景色），保留阴影区分窗口；实例中心内容区原有四角圆角在左下角露出壳体，现仅在深色模式将两个底角归零，消除侧栏底部的月牙形接缝。14 枚深色应用图标进一步按暗背景调低高亮、调整底板与主体图层对比，保留各应用造型和辨识色；浅色模式、六套主题配色与壁纸选择未改。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过。[六配色×深浅×六场景图库](../../web/ui-screenshots/dark-mode-review/index.html)共 72 张实景；[暖砂实例中心左下角局部](../../web/ui-screenshots/dark-mode-review/sand-on/09-dark-instances-corner.png)用于复核接缝。图库的本地隔离 fixture 记录为零页面错误、零 API 写入；它不替代真实节点、后端或外部平台验收。下一动作：按用户对本轮暗色边界和图标亮度的实际反馈继续精修，不恢复已暂停的壁纸探索。

2026-09-24 深色模式去浑浊修订：用户指出六套深色截图整体发脏、图标难辨。本轮将暗色大面积窗口、菜单、Dock 和表单统一到中性炭灰层级，各主题的强调色保留在按钮、焦点和选中态；壁纸暗色显示层不再分别染棕/绿/紫，壁纸资源、选择项和浅色呈现未改。暗色材质面提高到约94–97%不透明并减轻阴影、卡片内描边；终端暗色画布由绿黑改为中性黑灰，已打开终端随主题即时切换。14枚选定应用图标重新以清晰浅彩底板和更强主体轮廓导出，保留原有01A造型与应用颜色。`npm run build`（含类型检查）通过，聚焦桌面 Playwright **4/4** 和终端恢复回归 **1/1** 通过；[重新拍摄的六配色×六场景深浅对照图库](../../web/ui-screenshots/dark-mode-review/index.html)共72张实景及六张概览，核验每套14枚浅色/深色图标、零页面错误和零API写入。截图使用隔离演示数据，不代表真实节点/后端验收；下一动作按用户对新暗色观感的反馈继续定向微调，暂不探索新壁纸。

2026-09-24 深色模式视觉适配：按用户最新反馈暂停壁纸探索，本轮没有修改壁纸资源、壁纸选项或六套主题配色。为选定的 01A 应用图标新增 14 枚逐个调整的深色资源；沿用原图形与应用辨识色，主题切换时立即替换，配色和通透开关不改变图标造型。修正深色顶栏、Dock、窗口、菜单、表单、列表、状态标签、应用卡片和日志表面的层次与边界；Monaco 编辑器在已打开时也随主题切换。冰蓝深色状态色与禁用文字对比单独调整。`npm run build`（含类型检查）及聚焦 Playwright **4/4** 通过，覆盖深色图标/编辑器实时切换、六配色和通透材质独立及刷新恢复；截图脚本语法检查通过。[六配色深浅色对照图库](../../web/ui-screenshots/dark-mode-review/index.html)共 **72 张**真实 Vue 截图，覆盖桌面、启动器、实例中心、设置、应用市场、监控与终端六种场景，另有六张概览。截图核验每套均加载 14 枚浅色与 14 枚深色图标，页面错误、API 写入均为 0。截图使用隔离演示数据，不扩大后端、节点或外部平台验收结论。下一动作：依据用户对暗色图标和各场景的反馈做定向微调；壁纸方向保持暂停。

2026-09-24 壁纸艺术媒介重做：用户指出此前六张只是同一套低饱和扁平插画的构图变化。本轮新增五个真正跨媒介方向：国家公园真实摄影、瑞士套印海报、深色数字光学、水墨纸张、压纹玻璃材质；保留此前壁纸与“随配色”默认值，不强制切换用户现有选择。真实摄影取自[NPS 公有领域原图](https://npgallery.nps.gov/AssetDetail/204B23EC-1DD8-B71B-0B21B310C2A73B98)，裁切及镜像等处理见[来源记录](../../web/src/appearance/wallpapers/media/PHOTO_SOURCE.md)；其余四张为代码绘制的 SVG，无 AI 图片生成。照片的桌面图标文字采用局部半透明底衬，夜光图用浅色文字以保持可读。壁纸独立选择仍不改变主题、通透材质或统一的 01A 彩色应用图标。`npm run build`（含类型检查）通过，聚焦 Playwright **3/3** 覆盖全部11种自选壁纸、配色/材质独立及刷新恢复。实拍[五种艺术媒介的四场景图库](../../web/ui-screenshots/wallpaper-art/index.html)，含桌面、启动器、实例中心、深色桌面共20张原图和4张对照图；五种浅/深色壁纸均各异、14应用图标资源及配色/材质固定，页面错误和API写入均为0。截图使用隔离演示数据，不扩大后端验收范围；临时服务与浏览器已停止。下一动作：根据用户对艺术方向的反馈选择和细化；不再把单纯换构图/配色描述为不同画风。

2026-09-24 六种独立壁纸风格探索：针对用户指出原六套壁纸仅是同一弧线构图换色，新增柔和色场、建筑平面、等高线、编辑网格、抽象雕塑、静谧地平线六张不同构图的 1600×1000 代码 SVG。设置新增独立“壁纸风格”选择，“随配色”为默认且保留原壁纸；指定风格时更换壁纸不改变主题主色、浅/深色、通透材质或选定的 01A 彩色应用图标。聚焦 Playwright **3/3** 覆盖旧配色/通透独立性和六张壁纸唯一性、配色/材质独立、刷新持久化；`npm run build`（含类型检查）通过。用真实 Vue 桌面和隔离演示数据拍摄[六种壁纸的四场景图库](../../web/ui-screenshots/wallpaper-styles/index.html)，含桌面、启动器、实例中心、深色桌面共24张原图与4张对照图；核验六种浅/深色壁纸各异、配色/材质/14图标资源一致，0 页面错误、0 API 写入。截图不扩大后端验收范围；临时服务已停止。下一动作：按用户对壁纸方向的选择收敛，而非继续仅更换弧线配色。

2026-09-24 六套主题配色实景对照：在冰蓝、青灰之外新增暖砂、苔绿、雾紫、烟粉；每套具有独立的浅/深色语义色和代码绘制壁纸，保留用户选定的 01A 彩色应用图标。主题配色、浅/深色和通透材质仍分别切换，通透开关不改变当前配色、壁纸或图标。`npm run build`（含 `vue-tsc --noEmit`）通过；聚焦 Playwright **2/2** 验证四套新增配色及旧两套的材质独立性与刷新持久化。[六套主题图库](../../web/ui-screenshots/palette-showcase/index.html)包含六配色×通透开/关×桌面、启动器、实例中心、设置、深色桌面、深色启动器，共 **72 张真实 Vue 截图**及六张概览图；截图脚本核对六种主色/壁纸各异、同配色两材质主色/壁纸一致、12 组合的 14 个应用图标资源相同、材质确实变化，页面错误和远端写入均为 0。截图使用隔离演示数据，不扩大后端验收结论；本地截图服务和浏览器已停止。下一动作：根据用户对六套配色的选择，细调选中方案的壁纸与界面色。

2026-09-24 通透材质与主题配色解耦：按用户新要求，工作区设置新增独立“主题配色”（冰蓝 / 青灰）选择；通透模式现在只控制顶部栏、窗口、菜单与 Dock 的通透材质，切换时保留配色、壁纸和统一的 01A 应用图标。配色选择负责色彩令牌与配套壁纸，不会改变通透开关；原有“浅色 / 深色”主题仍单独保留。旧工作区尚无配色字段时，按原通透开关推断并在首次切换前固化旧视觉选择，避免迁移后突然换色。样式将冰蓝色值/壁纸与调色板无关的通透层分离，通透层改用当前配色色值混合，不再硬编码冰蓝。聚焦 Playwright 覆盖四组合、主色/壁纸/图标不随材质变化、两设置独立持久化 **1/1**；`npm run build`（含 `vue-tsc --noEmit`）通过。用真实设置控件与隔离演示数据实拍[四组合×25场景图库](../../web/ui-screenshots/palette-matrix/index.html)共100张及7张四宫格，包含深色设置页；校验同配色开/关主色、壁纸与14图标资源一致，顶部栏材质有差异；四组无页面错误或远端写操作。截图不扩大后端验收结论。下一动作：按用户对四组合效果的反馈微调材质强度与配色；不再将通透开关和配色合并。

2026-09-24 双模式应用图标统一：按用户反馈，实色模式不再切换到较简化的旧 h04 图标；通透与实色模式均使用选定的 01A（`h04-spectrum-silver`）彩色应用图标，开关只改变界面材质与壁纸。设置说明同步更新。用真实设置开关重拍[通透24场景](../../web/ui-screenshots/translucency-on/index.html)、[实色24场景](../../web/ui-screenshots/translucency-off/index.html)及[24组对照](../../web/ui-screenshots/translucency-comparison/index.html)；实际渲染的14个图标资源路径在两组中完全相同，零页面错误、零远端写入。`npm run build`（含 `vue-tsc --noEmit`）通过，Playwright 通透开关/刷新后图标一致性测试 **1/1** 通过。截图仍使用隔离演示数据，不扩大后端验收结论；临时截图服务已停止。下一动作：按用户对双模式统一图标的实景反馈微调。

2026-09-23 实例中心控件回退与通透模式开关：按用户图1撤回上两轮放大搜索条/白底下拉框的改动，恢复直接搜索输入和紧凑填充控件；下排三筛选框回到28px/11px并保持等宽。实际截图的实例窗口区域与既有图1/图3候选截图近乎逐像素一致，控件区域三通道平均绝对差均小于0.35/255。工作区设置新增可访问的“通透模式”开关，偏好随工作现场持久化，默认开启；开启切换冰蓝通透材质、曲面壁纸与已选01A彩色图标，关闭切换用户图3的青灰实色表面、默认壁纸与默认图标。补齐两种模式深色桌面的壁纸压暗与可读文字，开启时深色窗口/Dock仅保留克制通透。用真实设置开关和隔离演示数据分别实拍[开启24场景](../../web/ui-screenshots/translucency-on/index.html)、[关闭24场景](../../web/ui-screenshots/translucency-off/index.html)，并生成[24组并排对照](../../web/ui-screenshots/translucency-comparison/index.html)；两组均14图标、零pageerror/远端写操作，图库筛选和48张原始PNG完整性核验通过。`npm run build`（含`vue-tsc --noEmit`）、`npm run test:e2e -- tests/browser/desktop.spec.ts tests/browser/instance-batch.spec.ts` **8/8**、截图脚本语法检查通过。截图为本地隔离演示数据，不扩大后端验收结论；Vite和浏览器会话已结束。下一动作：根据用户对两种模式实景对照的反馈微调；不要再自动把实例筛选放大。

2026-09-23 实例中心节点筛选圆角修正：用户指出“全部节点”下拉框看起来变成直角；前一轮照搬4px小圆角到56px高控件，实际视觉比例不合适。现将该下拉框调为16px圆角、下排40px高的三个筛选框调为12px圆角，保留原生选择语义、高度和布局。检查无其他CSS覆盖形状；重新实拍[实例列表](../../web/ui-screenshots/01a-flow-contexts/05-instances.png)与[1366宽列表](../../web/ui-screenshots/01a-flow-contexts/23-laptop-instances.png)，并重拍完整24场景图库（14张选定图标、零pageerror/远端写操作）；`npm run build`（含`vue-tsc --noEmit`）通过。截图使用隔离演示数据，不计后端验收；临时Vite与浏览器已停止，无运行会话。下一动作：按用户对圆角比例的反馈微调，其余视觉继续沿用选定原版。

2026-09-23 实例中心表单形状核对与修正：用户质疑截图中的输入框是否符合 Material Design。核对 Google 官方 Material Web [文本框](https://github.com/material-components/material-web/blob/main/docs/components/text-field.md)、[选择框](https://github.com/material-components/material-web/blob/main/docs/components/select.md)和 Android [搜索](https://github.com/material-components/material-components-android/blob/master/docs/components/Search.md)文档后确认，原搜索框约28px、第二排筛选约24px且视觉角色混同，不能称为按 M3 默认组件实现。现将实例搜索改为带搜索图标的56px搜索条，节点下拉改为56px/4px角的选择框，第二排改为40px/8px角的桌面密度选择框；保留真实原生选择语义、筛选状态和布局。当前是基于官方角色与形状令牌的桌面适配，不宣称引入官方 Material Web 组件库或每个像素都与 Android 默认控件相同。已重拍[实例列表](../../web/ui-screenshots/01a-flow-contexts/05-instances.png)、[卡片](../../web/ui-screenshots/01a-flow-contexts/06-instance-cards.png)和[1366宽列表](../../web/ui-screenshots/01a-flow-contexts/23-laptop-instances.png)，并修正窄屏截图误拍资源详情的问题；全24张图库零pageerror/写API。Playwright `desktop.spec.ts`＋`instance-batch.spec.ts` **7/7**通过，`npm run build`（含 `vue-tsc --noEmit`）与截图脚本语法检查通过。截图仍是隔离演示数据，不计后端验收；临时Vite与浏览器已停止，无运行会话。下一动作：依据用户对新比例的视觉反馈继续微调，不把其他页面自动认定为已完成 Material 3 审计。

2026-09-23 选定原版的多场景截图：按用户要求继续展示 **01A (`spectrum-silver`) 图标＋W4青灰流线 (`flow`) 壁纸＋现有Material布局**，复用真实 Vue 组件截图库并新增 `--01a --flow` 参数，生成[24张可筛选图库](../../web/ui-screenshots/01a-flow-contexts/index.html)：桌面/启动器/Dock、实例列表与卡片、文件和编辑器多窗口、终端/监控、应用市场、节点/任务、设置/账号、深色主题、1366×768与1024×768。新浅色启动器与用户选定的原图同为3200×2000，RGB逐像素完全一致。抽查时发现原深色预览因浅壁纸导致桌面标签几乎不可读，已在预览专用 `ice.css` 中为同一壁纸加均匀暗色遮罩；深色图库明确标记为适配预览。14张选定图标均加载，24张PNG/27张图库图片与分类筛选核实；零pageerror、零API写请求，终端只读样本协议连接1次。`node --check .local/capture-04a-contexts.mjs`、`npm run check`、`npm run build`通过；见[验证结果](../../web/ui-screenshots/01a-flow-contexts/verification.json)。截图使用隔离本地演示数据，不增加后端验收结论，也没有将预览图标/壁纸切换为产品默认。临时Vite和浏览器已停止，无运行会话。下一动作：按用户对这些实景截图的反馈细调选定视觉基线，不恢复已被否定的其他图标方向。

2026-09-23 用户选回原版作为设计基线：用户在看过F1～F4后表示“感觉还是这个原版好看”。附图与 `harmony-12/spectrum-silver--flow--launcher.png` 的3200×2000 RGB像素完全一致，明确对应 **01A (`spectrum-silver`) 原版图标＋W4青灰流线 (`flow`) 壁纸＋冷白冰蓝界面/现有Material布局**。后续以此为基线，保留各应用独立颜色、图标底板与h04轻透材质；F1～F4及其他配色保留为探索档案，不作为当前选定方向。已将[对照页](../../web/ui-screenshots/harmony-12/index.html)默认组合改为该选择，并标出基线及[原图](../../web/ui-screenshots/harmony-12/spectrum-silver--flow--launcher.png)。本轮仅记录选择并修改图库默认展示，未把偏好反馈扩展为正式产品发布或默认主题切换；无需重跑产品构建。没有启动服务，无运行会话。下一动作：围绕此基线接受具体细调要求，不继续无目标扩散图标风格。

2026-09-23 图标主体四方向发散：用户明确要求图标本身产生多个方向。新增F1透明叠片、F2饱满大形、F3精致器物、F4彩色折面，每套14个主体全部重画；固定上轮A浅底板、暖白壁纸与h04材质。渲染器补齐旋转和evenodd填充，导出56张288×288资产共814,685字节。最终类型检查/构建通过，正式产物仍仅14张默认图标；33张Chromium实拍含全套/尺寸/桌面/启动器/Dock/实例窗口。去除颜色后比较确认56个几何全部改变、方向间无同应用重复，底板一致；布局、25种图库切换及390px图库验证通过，零pageerror/写API。人工修正市场类似卡通脸、设置类似风扇及分色齿轮重复边缘后重拍对应截图。见[造型与验证记录](ICON_FORMS_13.md)、[四套对照页](../../web/ui-screenshots/forms-13/index.html)。仅开发预览启用；使用隔离本地演示数据，不计后端验收。全部本轮浏览器、Vite与构建会话已结束。下一动作：按用户对主体方向的反馈继续细化，再与已认可配色/壁纸组合，不自动切换默认。

2026-09-23 01A 图标/壁纸协调方案：针对实色底与浅色底割裂，新增浅瓷底、适中彩底、全彩底、深底彩芯四套完整配色，主体和内部图层同步调整；保留01A图形、h04材质和Material布局。另有暖白、纸层、雾紫、青灰流线、浅灰纹理五张代码壁纸，可与原版一起独立切换30种组合。56张288×288图标共749,685字节，仅开发预览启用。类型检查/构建通过，正式产物仅14张默认PNG；实际Chromium实拍57张，30组启动器/5组窗口几何一致，零pageerror/写API，图库切换与390px宽度验证通过。审图后修顺青灰壁纸尖角并重拍7张实景和2张对照。见[实现与验证记录](ICON_HARMONY_12.md)、[完整对照页](../../web/ui-screenshots/harmony-12/index.html)。截图为隔离本地演示数据，不增加后端验收结论；默认外观未切换，临时Vite与浏览器已停止，无运行会话。下一动作：依据用户对图标和壁纸的独立选择继续收敛。

2026-09-23 改用01A预览：用户否定04A实景方向并要求查看01A。沿用原始 `spectrum-silver` 资产，在与上一轮相同的24个场景、固定时间和窗口内容下重新截图；未重绘图标或修改产品布局/主题。新启动器特写与最初01A特写逐RGB像素一致（1240×904）；14张01A资产全部加载，零pageerror、零API写请求，图库27张图片和3/3/4/24分类验证通过。见[01A完整图库](../../web/ui-screenshots/01a-contexts/index.html)、[验证结果](../../web/ui-screenshots/01a-contexts/verification.json)。复用命令：`cd web && node .local/capture-04a-contexts.mjs --01a`（需先启动本机Vite）。截图继续使用隔离演示数据，不计后端验收；仅截图脚本新增01A参数，无需重新构建产品。临时Vite和浏览器已停止，无运行会话。下一动作：依据用户对01A的反馈继续，04A已被用户否定，不能作为新的默认方向。

2026-09-23 04A 多场景视觉评审：按用户选中的 `pastel-silver` 扩展到桌面/启动器、实例列表与卡片、文件编辑多窗口、终端与监控、市场、节点、任务、设置、账号菜单、深色主题及两档小屏，共24张2倍像素截图。保留图标与产品源码，未切换默认主题。全部渲染图标为04A，14张资产加载；实际Vue/Monaco/xterm组件使用隔离只读视觉样本，零pageerror、零API写请求，图库27张图片加载及四类筛选通过。记录现有深色主题配浅壁纸造成桌面文字对比度不足。见[截图与验证记录](ICON_04A_CONTEXTS_11.md)、[24张图库](../../web/ui-screenshots/04a-contexts/index.html)。临时Vite和浏览器已停止，无本轮运行会话。下一动作：依据用户对04A实际场景的反馈继续收敛，不自行恢复其他环境验收。

2026-09-23 终端与监控六组细化：用户认可01/04整体配色，要求改终端造型及两应用颜色。基于两套基础分别新增A银白窗口＋绿白监控、B紫色叠片＋蓝白监控、C浅色命令页＋暖橙监控，共六组。新代码绘制三种终端几何，继续沿用h04材质和底板；每组另外12个图标与对应基础版PNG逐字节一致。84张288×288候选资产共1,095,755字节，仅开发参数启用。最终类型检查/构建通过，正式产物仅14张默认图标；Chromium禁用WebGL完成30张实景/局部＋2张对照，六组13应用图标加载、三类布局比较与对照页切换均通过，零pageerror。见[细化与验证记录](ICON_UTILITIES_10.md)、[六组对照页](../../web/ui-screenshots/icon-utilities/index.html)。截图使用只读本地演示数据，不计真实终端或后端验收；所有本轮进程已停止。下一动作：按用户对01A/B/C和04A/B/C的选择继续收敛。

2026-09-23 图标独立配色四版：用户认可 h04 图标质感但要求摆脱统一冰蓝色。参考 Apple App icons 与 Liquid Glass 官方文档，保留图标几何、材质、底板和冷白冰蓝桌面，逐应用/逐图层制作「应用原色」「浅底彩芯」「明暗交错」「彩色薄片」四组；56张代码渲染PNG共734,796字节。开发参数 `?appearance=ice&icon-colorway=...` 可切换，默认版不变。禁用WebGL的Chromium实拍四版各桌面/启动器/特写/Dock/实例中心及并排总览21张，13个已安装应用图标逐版加载，三组场景几何一致，零pageerror。首次构建发现预览PNG误入正式产物，已改为仅开发服务器加载；复验 `npm run build` 通过，正式产物只含14张默认图标PNG。对照页切换和原冰蓝预览资产加载复核通过。见[设计、截图与验证记录](ICON_COLORWAYS_09.md)、[可切换对照页](../../web/ui-screenshots/icon-colorways/index.html)。截图仅用只读本地演示数据，不计后端验收；Chromium及临时Vite已停止，无运行会话。下一动作是根据用户具体偏好收敛各应用颜色。

2026-09-23 冷白冰蓝单版预览：按用户新参考图保留当前 Material 形状、圆角、布局与 h04 图标底板，新增开发参数 `?appearance=ice`；用冷白/蓝灰表面、蓝/青绿/紫/深灰图标、柔和阴影和代码曲面壁纸复现参考氛围，未替换默认主题。14个图标仅改原图层颜色，保留轮廓与材质参数，导出288×288资产共182,848字节；正常显示不依赖WebGL，没有图片生成或斜向反光条。最终类型检查/构建通过；禁用WebGL的Chromium实拍12张新版＋3张默认对照，七组场景的几何/字号/间距/圆角/Dock路径一致、零pageerror。已检查菜单透出与文字对比度并更新截图。见[实现与验证记录](MATERIAL_ICE_PREVIEW.md)和[交互对照页](../../web/ui-screenshots/material-ice/index.html)。截图使用本地演示数据，不计后端验收；浏览器和临时Vite均已停止。下一动作：依据用户对这一个版本的反馈收敛外观，继续保留Material形状与图标底板。

2026-09-22 h04 保留底板设为默认：用户选择“保留底版好一些”。已将「透明包边＋保留底板」统一用于桌面、启动器、Dock、窗口标题栏和应用市场。原材质/图形/配色未改；从原代码渲染器导出14个288×288资产（185,447字节），默认显示不再依赖实时WebGL，保留原viewBox与Dock比例。类型检查/正式构建通过；本地正式构建在禁用WebGL的Chromium中验证14类资产全部加载、零WebGL请求、零pageerror，桌面→启动器→实例中心和生产忽略开发参数检查通过，另有5张实际截图。见[默认图标记录](ICON_GLASS_H04_DEFAULT.md)。测试使用本地演示数据，无远程节点操作或公开部署；临时开发/预览服务及浏览器已停止。下一动作：后续视觉细化以本版为基线，保留底板，不恢复斜向反光或其他被否定方向。

2026-09-22 h04「透明包边」实际场景扩展：用户明确选定 h04，要求多种实际界面截图。本轮保持 h04 图形/配色/材质参数，不新增风格；开发预览增加 `icon-surface=bare` 对齐所选裸主体截图，`icon-background=graphite|warm` 提供深浅背景对照。24 张真实组件实拍涵盖完整桌面、13应用启动器、实例列表/卡片、文件/编辑器、多窗口、应用市场、节点/任务、设置/账号、深色、暖灰、1366×768和1024×768、字号及保留底板对照。最终脚本24张零pageerror、所有光学图标ready，PNG与浏览页链接核实；生产构建/类型检查通过。首轮主题选择器不匹配已修正并完整重拍。API使用本地演示数据且拒绝写请求，未操作真实节点，不新增后端验收结论。默认图标未切换；禁止斜向反光和图片生成继续有效。见[场景记录](ICON_GLASS_H04_CONTEXTS.md)及[24图浏览页](../../web/ui-screenshots/glass-h04-contexts/index.html)。本轮浏览器已退出，临时Vite已停止，无运行会话需接管。下一动作：依据用户对实际场景的反馈收敛h04；不擅自加强光影或恢复被否定的其他方案。

2026-09-22 图标玻璃材质第八轮：用户明确禁止斜向反光带。已从shader和配置删除该效果，旧g系列后续渲染也移除；今后不再引入斜向反光、扫光条或面内大块方向光影。新增h01～h08八种透色/吸色/折射/包边/彩边/夹层/乳白/平衡壳层候选，保留原图形与底板；细线壳宽限制为笔画20%，避免备份箭头变淡。构建通过；八版前景内部像素验证确认实际透色，平整中央无方向光色差。并发构建/浏览器超时后改为顺序执行复核。46张代码实拍涵盖八版全套、小尺寸、桌面/启动器/Dock与测试衬底；默认图标未切换。见[玻璃通透记录](ICON_GLASS_08.md)。下一动作：按用户选定的玻璃方向收敛，严格保留禁止斜向反光带和禁止图片生成要求。

2026-09-22 图标光学材质第七轮：用户指出薄边缘不可见，要求更多反射玻璃方向。查阅Apple Liquid Glass官方说明，新增共享WebGL渲染器：保留原14应用/63部件，轮廓生成法线/距离纹理，逐层折射采样与移动光源反射，6版主体不使用Gaussian blur。开发 `?icon-study=g01`～`g06`，默认图标不变；36张实际Chromium截图含裸主体、深浅底、完整图标、32/48/64px、三光源和桌面/启动器/Dock。类型检查/构建通过；canvas渲染状态与六版左右光源图像变化均核实，无pageerror。修复透明采样暗点及光源轴向；一次Vite被SIGTERM后的连接失败已恢复重拍。属于网页近似材质候选，非Apple原生或跨平台性能验收。见[光学候选记录](ICON_OPTICS_07.md)。下一动作：按用户选择收敛，继续直接代码渲染，不用图片生成。

2026-09-22 图标光影第六轮：用户认可主体模糊透叠，但要求优化夸张渐变及提供无光影版。以f06固定几何/颜色/透明度/blur制作A1微弱面光、A2仅薄边缘、B无光影；B的SVG不生成任何渐变/光斑/亮边，仅保留图层模糊透色。修复填色与描边重复合成的暗边。默认图标未切换，开发参数 `?icon-study=l01`～`l03`；15张真实Chromium截图覆盖放大、全套、小尺寸与桌面/启动器/Dock。类型检查/构建通过，截图脚本验证B零光影定义、模糊仍存在、底板一致且无pageerror。见[光影对照记录](ICON_LIGHTING_06.md)。下一动作：按用户选择收敛；不恢复无关阻塞验收。

2026-09-22 图标主体材质第五轮：用户指出上一轮只改底板。本轮新增原版14应用/63部件的逐层材质合成，Gaussian diffusion裁剪于主体面板，文件夹前板实际透出后方纸张，笔身和店铺等分别处理，6版共用原版不透明底板。备份等粗线使用alpha mask，终端符号保留清晰实色。全部SVG/Vue代码，无图片生成；仅开发 `?icon-study=f01`～`f06` 启用。类型检查/生产构建及Chromium截图验证通过，28张包含去底板主体对比、放大、20/32/48px和实际桌面/启动器/Dock；中途线条裁剪问题已修复并重拍。见[主体材质记录](ICON_FOREGROUND_05.md)。下一动作：依据用户对主体材质的反馈收敛，默认图标尚未切换，不恢复无关环境阻塞验收。

2026-09-22 图标材质第四轮：用户否定第三轮拟物化，明确以最初Material图标为基础，仅增加适量通透与高斯模糊材质。新 `ClassicIconArtwork.vue` 提取原版图形，源码比较确认几何不变；`MaterialAppIcon.vue` 实现12种 Gaussian blur底层透色/CSS backdrop-filter/微弱层次，非整图模糊。全部手写SVG/CSS，无图片生成。仅开发 `?icon-study=m01` 至 `m12` 启用；`icon-materials.html` 包含原版对照、深浅底、20/32/48像素和分组。修正SVG属性类型后类型检查与构建通过；Chromium12套实际桌面/启动器无pageerror，53张2倍像素截图核实齐全。见[材质方案与官方资料](ICON_MATERIALS_04.md)。下一动作：依据用户选择的材质方向细化；不自行改变默认风格或恢复环境阻塞验收。

2026-09-22 应用图标第三轮：用户指出 D–L 偏工具栏功能符号，并明确要求停止图片生成、直接编写代码后截图。已遵从：新 `NativeAppIcon.vue` 14图标全部使用手写 SVG，未引入生成图片资产；文件夹/文档笔/安装袋/服务器托盘/包装箱采用独立轮廓，终端与监控使用完整面板，有限局部渐变只表现材质面。仅开发 `?icon-study=native` 启用，默认方案保持评审前状态。`native-icons.html` 展示四个代表图标、全套及 Dock；截图脚本 `web/.local/capture-native-icons.mjs` 输出8张2倍像素图，目录 `web/ui-screenshots/native-icons-03/`。`npm run check`、`npm run build` 通过；Chromium 桌面→启动器→实例中心正常，无 pageerror。固定 API 演示数据不构成后端验收。下一动作：依据用户对完整应用造型的反馈细化；后续图标工作直接使用 SVG/CSS，不使用图片生成。

2026-09-22 图标设计第二轮：用户否定 B 渐变并要求更多方向。移除 B 主体/底色渐变；新增 D–L 九套14图标候选（双色线条、全幅纯色、无底板、折面几何、石墨面板、几何标记、透明浅底、侧边索引和墨色主体）。四套线条方向统一 Lucide，其余为本地双色矢量；默认图标与桌面布局保留。开发预览 `?icon-study=d` 至 `l`，总览 `/icon-explorations.html`。类型检查和生产构建通过；Chromium生成37张截图，包含九套真实桌面/启动器及20/32/48像素和深底检查。见[设计与截图索引](ICON_EXPLORATIONS_02.md)。下一动作：等待用户评价具体方向再细化，未恢复环境阻塞的其他任务。

2026-09-22 图标三方向设计评审：按用户最新要求重启图标视觉探索，保留桌面布局，新增 `IconStudy.vue` 三套14图标候选：A双色圆形、B浅色层叠、C深石墨双色。仅开发模式 `?icon-study=a|b|c` 使用候选，默认图标未切换；`icon-study.html` 提供并排比较。参考 Android 官方 adaptive icons 和 Apple HIG app-icons 的遮罩、分层与清晰轮廓原则，非系统原生材质复刻。`npm run check`、`npm run build` 通过，Chromium 已生成10张2倍像素截图，目录 `web/ui-screenshots/icon-study/`，包括三套实际桌面/启动器及图标对比。截图使用固定登录测试数据，不作为后端验收证据。下一动作：根据用户选定方向细化，不自行恢复其他环境阻塞任务。

2026-09-22 E08 Firefox 双 PTY 实测：官方 Firefox 155.0 在强化后的真实8窗口/双PTY/万项目录/16MiB循环传输60秒组合中，两条PTY分别116,029.8/116,120.3B/s、四轮传输、队列和实例状态均正常，但1,020个指针样本 p95 **144ms**，故按≤50ms阈值失败；Firefox无可用WebGL，两个终端使用默认渲染器。该结果与WebKit78ms、Chromium42ms并列记录，不能用Chromium覆盖非Chromium失败；见[Firefox报告](../acceptance/reports/firefox-performance-dual-pty-2026-09-22.md)。下一具体动作：本机已补齐三种可用浏览器的严格双PTY短时组合，停止无信息增益的浏览器重跑，仅在获得可区分绘制热点或外部平台时继续。

2026-09-22 E08 双 PTY 负载复核：修正真实性能测试不稳定地按前两个终端 DOM 容器输入的问题，改为在每条新会话仍处于前台时向其唯一可写 PTY 注入有界输出；每分钟严格要求两个 WebSocket 都有新增字节。官方 WebKit 26.6 的60秒真实8窗口/万项目录/16MiB循环传输组合中，两条 PTY 分别约115.8KB/s、四轮传输与队列均正常，但3,072个指针样本 p95仍为**78ms**，故按≤50ms阈值失败；同一修正场景官方 Chromium 双PTY p95 **42ms**通过。后台终端画面暂停无收益，动画帧合批又会破坏 Vim 后紧接输入的时序，均已撤回；回退后的真实 Vim/刷新/不重放输入1/1（9.1s）通过。UI未改；见[双PTY复核](../acceptance/reports/webkit-performance-dual-pty-short-2026-09-22.md)。下一具体动作：不重跑已知失败的一小时WebKit负载；本机只剩需要新的可区分性能热点或外部平台的严格验收。

2026-09-22 E02 rootless Engine 安全替代未启动：官方 `docker:29-dind-rootless` 未使用 `--privileged`；精确把任务证书目录交给其 UID/GID 1000、模式700后，mTLS证书成功生成，但 rootlesskit 受当前 seccomp 的 `/proc/self/exe` 执行限制拒绝（`operation not permitted`）。未尝试 `seccomp=unconfined` 或其他安全策略放宽；已删除失败容器、私有证书/密钥目录和拉取镜像。独立 Engine 仍需明确授权高权限/放宽安全策略，或使用外部测试主机。

2026-09-22 E02 独立 TCP Engine 探索未执行：为补足现有 loopback TLS 代理之外的独立 Engine 协议链路，计划启动任务专用 mTLS Docker-in-Docker 并仅映射临时 loopback 端口、导入现有测试镜像；平台自动审批拒绝 `--privileged` Docker-in-Docker，理由是其扩大到宿主内核与容器边界的高权限范围。没有执行特权容器、端口映射、镜像导出或证书写入；复核无残留。当前安全替代不能提供独立 Engine，因此 E02 的跨主机/独立 Engine 缺口继续保留，等待明确授权或外部测试主机。

2026-09-22 真实 Chromium 容器回退修正：`playwright.real.config.ts` 在未设置 `BLORA_CHROMIUM` 时改由锁定 Playwright 解析自带 Chromium，显式变量仍优先系统路径；官方容器此前因硬编码 `/usr/bin/chromium-browser` 无法启动。类型检查通过；在80ms/10ms抖动/1%丢包的隔离 namespace 中，官方 Chromium E01完整组合 **1/1通过（3.6m）**，与既有WebKit证据合并见[netem监控历史报告](../acceptance/reports/netem-monitor-history-2026-09-22.md)。fixture、规则和状态均已清理。下一具体动作：不重复已覆盖的本机netem浏览器组合，保留剩余跨主机、Windows、systemd、物理故障与24小时边界。

2026-09-22 E07 受控网络补验：80ms单向延迟、10ms抖动、1%丢包的任务专用Docker namespace 中，WebKit 本地注册源两版本浏览、安装、升级、禁用/启用与 sandbox bundle 执行 **1/1通过（13.6s）**；见[netem注册源报告](../acceptance/reports/netem-extension-catalog-2026-09-22.md)。fixture/规则/状态目录已清理。下一具体动作：停止重复单机netem组合，回到仍需外部环境的跨主机、Windows、systemd、物理故障与24小时验收缺口。

2026-09-22 E06 受控网络补验：同一任务专用 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 独立扩展包安装、多窗口、真实WASI节点任务、服务端接受后浏览器回执丢失的同键恢复、快捷方式和刷新现场 **1/1通过（49.4s）**；扩展测试改为断言明确 `Error:` 状态，兼容 WebKit 的 `Load failed` 与 Chromium 的错误文本，仍严格核对同请求键恢复。报告见[netem扩展任务报告](../acceptance/reports/netem-extension-task-2026-09-22.md)，测试容器和状态已清理。下一具体动作：继续只运行不同的可隔离验收组合，不用网络场景替代Windows、跨主机或物理故障。

2026-09-22 E04 受控网络补验：任务专用 Docker namespace 的 `tc/netem` 80ms单向延迟、10ms抖动、1%丢包下，官方WebKit的真实分钟计划场景 **1/1通过（1.5m）**；相邻两个槽各唯一接受一次，第二槽绑定更新后的源正文，两份归档均恢复核对。fixture、namespace规则和状态目录已清理；见[netem分钟调度报告](../acceptance/reports/netem-schedule-wall-clock-2026-09-22.md)。不替代跨主机、24小时或物理掉电验证。下一具体动作：继续审计是否还有可在隔离环境中补充且不重复的严格验收组合。

2026-09-22 E01 受控网络补验：任务专用 Docker network/PID namespace 的 `lo` 用 `tc/netem` 实际施加80ms单向延迟、10ms抖动和1%丢包；官方 WebKit 与私有HTTPS Master/双Daemon fixture 上，1秒间隔121点→Daemon离线stale→Master重启缓存→Daemon恢复完整场景 **1/1通过（3.6m）**。容器销毁即删除私有netem规则，fixture状态目录也已清理；详见[netem监控历史报告](../acceptance/reports/netem-monitor-history-2026-09-22.md)。这增加受控网络证据，不替代跨主机远端Engine、Windows或24小时墙钟。下一具体动作：继续只选择能在隔离环境实际运行的缺口；不重跑已覆盖的普通浏览器与本地Docker场景。

2026-09-22 E08 WebKit 失败后诊断：官方容器空白页双 `requestAnimationFrame` 180次为 p50 32ms/p95 33ms；60秒真实混合诊断中恢复日志写入p95 1ms、没有 long task，慢响应集中在事件处理后的绘制阶段。尝试将 WebKit 的xterm从WebGL改为默认渲染器，真实p95恶化至130ms，已撤销，源码不保留无收益改动；详见[WebKit小时报告](../acceptance/reports/webkit-performance-hour-2026-09-22.md)。下一具体动作：只有在具备可区分WebKit绘制热点的独立profile时才继续性能代码优化；其余严格验收继续等待Windows、独立systemd、远端高RTT/丢包、物理故障和24小时环境，避免重复长测。

2026-09-22 E08 WebKit 一小时真实复验：官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 容器中的 WebKit 26.6 以私有 HTTPS 双 Daemon `--performance` fixture 运行 3,603.841 秒，8窗口/双PTY/10,000项目录/16MiB循环传输均持续完成；539轮传输、末轮目标、两路终端和实例状态均正常。但 71,496 个指针样本 p95 **78ms**，超过≤50ms阈值，Playwright 退出1；详见[WebKit小时报告](../acceptance/reports/webkit-performance-hour-2026-09-22.md)。已删除容器、停止fixture并清理目录。该失败保留为非Chromium实际证据，不覆盖 Chromium 的48.1ms通过；下一具体动作：不重复本地长测，整理当前严格验收中只能在 Windows、独立systemd、远端高RTT/丢包、物理故障和24小时环境完成的入口与缺口。

2026-09-22 E01/F10 WebKit 真实补验：Fedora 宿主无法直接运行 Ubuntu fallback WebKit（缺 ABI 依赖），因此使用版本匹配的官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 隔离容器；私有 HTTPS Master/双 Daemon fixture 上，以1秒间隔运行121点真实采样→Daemon离线 stale→Master重启缓存→Daemon恢复链路，WebKit **1/1通过**。容器按 `--rm` 删除，fixture Ctrl+C 后进程、9443监听和临时目录均清理；报告见[WebKit监控历史](../acceptance/reports/webkit-monitor-history-2026-09-22.md)。这补齐 E01/F10 的 WebKit 浏览器证据，不替代 Windows、24小时墙钟、远端网络或高 RTT/丢包。

2026-09-22 B10 当前源码缺口复审：重新对照 `ACTION_GUIDE.md` 的 B00～B10 完成标准、两份规划和验收矩阵；生产代码扫描未发现 TODO、固定成功/假数据或未实现占位，`CAPABILITY_UNAVAILABLE` 仅出现在明确的能力边界（未开放通道、无容器运行时和由所属应用处理的重试）。A01～A08、A10～A12、A14～A17 的实现列为已实现，A09/A13 以及 F/E 整项的剩余条件均指向 Windows、systemd、远端网络、物理故障、长时或非 Chromium 环境。当前没有新的安全本地生产改动可补；继续重复现有短测不会改变验收状态。

2026-09-22 E05 私有 systemd 探针收口：在 `/tmp` 临时运行目录中尝试启动隔离 user manager；`unshare --user` 的 uid 映射和 `/proc` 挂载均被当前沙箱拒绝（`Operation not permitted`），直接启动私有 `systemd --user` 后也无法通过本地 user bus 连接（同样为 `Operation not permitted`）。未接触宿主服务、cgroup 或 D-Bus；探针目录、进程和 socket 已清理，复核无残留。E05 的真实 service/timer 生命周期仍需可操作的 systemd 环境，本机没有可再推进的安全路径。

2026-09-22 B10 测试夹具清理：复核无 `blora-devfixture`、Master、Daemon、Vite、Playwright 活动进程后，删除本任务生成的 120 个 `.local/fixture-*` 停止夹具目录；`.local` 从约 3.3GiB 可读残留降为 0，源码、`web/dist` 和 RC23 发行包未触碰。当前没有遗留 socket 或夹具会话。

2026-09-22 A09/E02/E08 网络边界复核：审查现有慢消费者、Docker HTTPS 代理和扩展远程目录测试后，确认本地可复用的证据已经覆盖有界日志/归档、控制请求不被阻塞、HTTPS Engine/Compose 传输，以及目录源的 80ms 延迟和证书轮换；没有把 loopback 代理伪称远端高 RTT/丢包。当前源码 `TestExtensionRemoteCatalogAdminAPI` 在沙箱外以 `go test -race` 复验通过（1.08s，退出0）；沙箱内失败仅因 `[::1]` loopback 绑定受限。由于本机没有 `tc/netem`、远端 Engine 或更高层网络故障环境，本地没有新的安全生产改动可补，剩余项转为相应平台验收手册入口。

2026-09-22 B10 环境复探与文档链审计：重新检查 `wine`/Windows 工具、systemd user bus、Docker endpoint、Chromium/Firefox/WebKit 和 `tc`；仍无 Windows、可操作 systemd manager、沙箱内 Docker socket 或 WebKit，Firefox 已有真实监控证据。验收矩阵、执行进度、README 和两份运维文档共 **5** 份 Markdown 的相对链接审计 **0** 条断链。没有因环境未变重复长时间测试。

2026-09-22 B10 本地收口审计：用户要求暂停视觉重构后继续核对可独立完成的功能。源码扫描未发现生产路径 TODO、固定成功响应或空实现；B07 的实例列表/卡片视图、批量操作、任务中心/站内通知，B08 的 Docker/Compose、监控、备份调度和有限系统管理，B09 的独立扩展包、沙箱、任务/资源桥接与迁移均已有实际入口和对应证据。本轮复跑 `cd web && npm run check`、`npm test -- --run` **63/63**、`npm run build` 和 `make check` 均退出0；当前源码 `make windows` 交叉构建也退出0。进度表中的旧“前端配置和模板正在接线”等描述已按当前源码修正；新增[平台验收运行手册](../operations/PLATFORM_VALIDATION.md)，把剩余 Windows、systemd、远程 Engine 和长期故障入口固定下来。剩余工作主要是矩阵严格验收仍缺的真实环境证据，不能由本机替代测试冒充完成。

2026-09-22 RC23 当前源码发行复核：`BLORA_VERSION=development-20260922-rc23 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`sha256sum -c SHA256SUMS` 六项通过，SHA256SUMS SHA-256 为 `5208cc267914f11144df11586ec4f3a7c936d92ff0a02c590fe745cf00dd5f82`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/static/login、双 Daemon、停机快照恢复和 RC22 兼容回退；见[RC23发行报告](../acceptance/reports/release-rc23-2026-09-22.md)。

2026-09-22 RC22 当前源码发行复核：`BLORA_VERSION=development-20260922-rc22 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`sha256sum -c SHA256SUMS` 六项通过，SHA256SUMS SHA-256 为 `eb6556276146b764795666924321d881b7c7dbdd0c8ad2e7f718459749b277cf`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/static/login、双 Daemon、停机快照恢复和 RC21 兼容回退；见[RC22发行报告](../acceptance/reports/release-rc22-2026-09-22.md)。

2026-09-22 Docker 日志归档入口增量：后端已有的有界 `GET /api/v1/nodes/{id}/docker/containers/{containerId}/logs/history` 现在接入 `DockerLogs.vue`，实时窗口与归档窗口明确分离，归档显示保留上限、观察时间、截断/可能缺口标记，并可返回实时日志；读取失败不伪造历史。`npm run check`、前端单测 **63/63**、`npm run build`、`make check` 均通过；Docker 浏览器链路 **2/2（11.1s）**通过，覆盖实时日志入口、归档内容、缺口提示和返回实时状态；真实隔离 Docker Engine 的 Master/Daemon 生命周期测试 **2.79s，退出0**，新增验证归档由 Master API 读取、未授权403和 `limit=101` 400；真实 Chromium Docker 日志 UI **1/1（7.6s）**通过，覆盖真实容器启动、实时正文、归档正文、缺口提示、返回实时及停止/删除清理；原有真实 Compose 长流程 **1/1（45.6s）**复验通过，配置草稿、显式应用、阶段输出和清理链路无回归。详见[Docker日志归档报告](../acceptance/reports/docker-log-history-2026-09-22.md)。这补齐了当前 Linux/浏览器/Docker 可推进的 F11 子项；远端 Engine、Windows 容器、主机掉电/存储耗尽和长时故障仍需相应环境。

2026-09-22 RC21 当前源码发行复核：预览模式控件禁用修正后的 `BLORA_VERSION=development-20260922-rc21 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `0c6c2a2d02c44c953d38c62aa1da9ed65b9759700c44be061b7a57eabaf86d9c`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC20 回退；见[RC21发行报告](../acceptance/reports/release-rc21-2026-09-22.md)。本地可推进的文件/桌面增量已继续完成，剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合等环境证据。

2026-09-22 F08 大文件只读分段预览增量：新增 `GET /api/v1/instances/{id}/files/preview`，按节点确认的文件版本读取不超过60KiB的 UTF-8 分段，带边界重叠、分片校验、偏移/总量/下一段状态，不把完整大文件装入 Master 或浏览器；二进制/NUL/非法 UTF-8 仍明确拒绝并保留下载入口。编辑器在超过4MiB时进入只读预览，支持上一段/下一段，保存/撤销/编辑控件不会误作用于预览。真实 TLS 文件集成包含大文件首段/第二段与二进制拒绝，定向浏览器边界 **3/3（14.9s）**通过；`make check`、前端 **63/63** 单测、`npm run check`、`npm run build` 和 OpenAPI 解析 **105 paths / 122 operations**通过；改动已纳入 RC21。详见[编辑器能力增量报告](../acceptance/reports/editor-capabilities-2026-09-22.md)。F08仍因完整跨应用/故障交叉及外部平台证据保持进行中。

2026-09-22 RC19 当前源码发行复核：设置页窄屏两列修正后的 `BLORA_VERSION=development-20260922-rc19 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `94590b0a100b2e1b60af204894fc8526752ff34d0729bf7264dd544ef65440fe`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC18 回退；见[RC19发行报告](../acceptance/reports/release-rc19-2026-09-22.md)。当前本地可推进的功能与回归已经继续完成，剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合等环境证据。

2026-09-22 RC18 当前源码发行复核：加入快捷键偏好后的 `BLORA_VERSION=development-20260922-rc18 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`npm run check`、前端单测 **63/63**、`npm run build` 和控制栏/设置浏览器回归 **3/3（9.9s）**通过；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `02f61fe468ba62d784df1d76f38cb79927cbd2685a4af3ba5c18624676d4c21c`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC17 回退；见[RC18发行报告](../acceptance/reports/release-rc18-2026-09-22.md)。下一步剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合；这些不是本地代码停滞，而是当前环境没有对应运行条件。

2026-09-22 工作区偏好与布局保护增量：设置页补齐主题、字号、界面密度和减少动态效果的工作区持久偏好；新增“恢复默认布局”，先保存完整布局备份工作区，再重置窗口几何、吸附状态、最小化状态和桌面入口坐标，草稿/终端检查点/资源身份仍保留在备份副本。`npm run check`、前端单测 **63/63**、`npm run build` 和设置/布局浏览器回归 **3/3（9.3s）**通过。当前源码重新打包为 `development-20260922-rc17`，六包 SHA 校验及独立包冒烟通过，见[RC17发行报告](../acceptance/reports/release-rc17-2026-09-22.md)。下一步剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合；这些不是本地代码停滞，而是当前环境没有对应运行条件。

2026-09-22 RC16 当前源码发行复核：`BLORA_VERSION=development-20260922-rc16 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `c3199c01de67de1e8da40b677caefad8e7b1960d4a2f879fe875dc76e97906ca`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC15 回退；见[RC16发行报告](../acceptance/reports/release-rc16-2026-09-22.md)。这轮没有停在 UI：当前源码中可本地闭环的文件目录树、列表/图标视图、列表框选、实例列表/卡片视图均已实现并有定向浏览器证据。下一步只继续处理尚无本地证据的 B07/B08 小缺口或整理全范围外部验证清单；Windows 真机、systemd、远程网络、物理掉电和长时跨平台仍需相应环境，不能用本地包冒烟替代。

2026-09-22 B07 实例中心视图增量：补齐规划要求的列表/卡片视图。两种排布共用实例筛选、选择、权限和幂等批量操作，视图选择写入标签现场并在刷新后恢复。新增报告[实例视图增量](../acceptance/reports/instances-view-2026-09-22.md)。`npm run check`通过；`npm run test:e2e -- tests/browser/instance-batch.spec.ts` **1/1通过（13.8s）**，含切换、刷新恢复和部分接受批量请求回归。下一步继续审计 B07/B08 尚无真实证据的本地功能，不重复已通过套件。

2026-09-22 B06 文件管理器视图增量：补齐规划要求的图标/列表视图，并为列表空白区域加入框选。列表保持按滚动窗口读取的虚拟化路径，图标视图按最多100项分页并提供前后页；视图和页码写入现场，刷新后恢复。新增报告[文件视图增量](../acceptance/reports/files-view-2026-09-22.md)。`npm run check`通过；前端单测 **63/63通过**；`npm run build`通过；`npm run test:e2e -- tests/browser/files.spec.ts` **5/5通过（19.0s）**，同时回归目录树、10,000项虚拟列表、框选、上传续传/取消和目录快捷方式。F07仍进行中：跨节点/物理故障/Windows场景按矩阵跟踪。下一步继续审计剩余 B07/B08 本地功能，不重复已通过文件套件。

2026-09-22 B06 文件管理器目录树增量：补齐规划要求的目录树入口。根目录及展开分支按需读取首分页块，未展开目录不递归请求；展开路径写入视图现场，点击目录复用路径历史，刷新后恢复树分支和当前路径。新增报告[文件目录树增量](../acceptance/reports/files-tree-2026-09-22.md)。`npm run check`通过；`npm run test:e2e -- tests/browser/files.spec.ts` **5/5通过（16.7s）**，同时回归10,000项虚拟列表、上传续传/取消和目录快捷方式。F07仍进行中：完整跨节点/物理故障/Windows场景仍按矩阵跟踪。下一步继续按B07审计任务中心、托盘通知和后台任务关闭网页后的本地恢复证据，不重复已通过的文件套件。

2026-09-22 F03/F08 编辑器授权、边界与恢复增量（分段预览加入前的记录）：新增文件访问能力查询与 fail-closed 只读编辑器，补充 UTF-8/BOM、LF/CRLF、最大字节数元数据往返、扩展名语言映射、共享正文分栏与多光标刷新恢复；超 4 MiB/非 UTF-8 错误页提供原始下载。编辑器撤销日志现按最多 8 MiB/文件上限两倍执行；活动输入组在操作结束前不被切开，单组超预算时完整落为新基线，避免只撤销部分粘贴。修复前的实测确曾暴露“撤销后正文仍是粘贴尾部”的缺陷，修复后恢复单测 **14/14**、编辑器 Playwright 合组 **8/8（36.1s）**、`npm run check` 与 `npm run build` 通过；后端真实 TLS Master/双 Daemon 文件 API race **11.302s**通过，覆盖精确 4 MiB 可读、超限/二进制拒绝和源字节下载。B06 本机文件系统 race 套件 **1.332s**通过（符号链接/FIFO/目录交换隔离），Windows amd64 `internal/filesystem` 测试二进制交叉构建成功（5.9 MiB，未在 Windows 运行）。OpenAPI 3.1.0 当时为 **104 paths / 121 operations**。详细证据和限制见[编辑器能力增量报告](../acceptance/reports/editor-capabilities-2026-09-22.md)。当时 F03/F08仍进行中：分栏固定50/50、编码只支持UTF-8/BOM、分段预览尚未实现；随后已由上方 F08 增量补齐，完整跨应用/故障组合和外部平台仍待验；视觉改动按用户要求冻结。

2026-09-21 E01 Firefox真实浏览器补验：`web/playwright.real.config.ts` 新增 `BLORA_BROWSER` 引擎选择（默认Chromium），`npm run check`通过。安装Playwright Firefox后，在隔离真实HTTPS双节点 fixture 上以1秒间隔跑完整121点采样→失联/stale→Master重启缓存→Daemon恢复链路 **1/1通过（2.9分钟）**。首轮 fixture 少显式 `-listen` 导致 supervisor 不能定位，已补参数复验；Firefox普通沙箱 profile错误后经批准沙箱外运行通过。Ctrl+C清理fixture，退出0，进程无残留，9443/9444不再响应。Windows、WebKit与24小时墙钟采样仍未验证；UI保持冻结。证据见[监控历史报告](../acceptance/reports/monitor-history-soak-2026-09-21.md)。

2026-09-21 E01跨UTC日边界补验：扩展现有节点/实例历史测试，让125个带时间戳样本跨越午夜；最新120点从次日00:00:01 UTC开始并有序保留至00:02:00 UTC。定向 `go test -race` 两项测试通过（2.718s）。这验证时间戳 key 跨日排序和裁剪，不代表运行24小时以上的实时采样；后者、Windows和非Chromium仍缺。沙箱拒绝 HTTPS fixture 所需 IPv6 loopback 后，获准在沙箱外执行相同测试并通过。UI保持冻结。

2026-09-21 E03/F12 备份恢复进程中断：新增 `internal/backup/restore_crash_test.go` 和[恢复进程崩溃报告](../acceptance/reports/backup-restore-process-crash-2026-09-21.md)。独立恢复子进程在真实 8 MiB 文件已有1,048,576B写入目标暂存区后被强制终止；重开原备份/目标状态后，原ID明确停在 `interrupted`，已完成目录作为部分进度保留、上传取消、暂存清除、正文未发布；同ID重放不变，生成新计划后重新恢复并核对清单SHA成功。定向 race 连续5次12.408s、备份全包 race 6.864s、`go vet ./internal/backup`与Windows amd64测试二进制交叉编译通过。只证明Linux进程级恢复，不覆盖物理掉电和Windows运行；UI保持冻结。下一步按矩阵继续找当前环境可真实执行、且不重复既有套件的验收缺口。

2026-09-21 E01 节点失联及 Master 重启缓存组合补验：fixture Daemon 在121点真实采样后按 `/proc` 身份精确 `SIGSTOP`，Master判为 `OFFLINE` 后仍读到120个 `stale` 点；测试再经fixture supervisor专用`SIGUSR1`请求优雅重启 Master，核实进程 PID 更换、健康检查/重新登录成功，原 Daemon 仍暂停时缓存的120个 stale 点及全部采样时间戳保持一致。随后 `SIGCONT` 原 Daemon并等待重新上线。1秒间隔组合真实Chromium 1/1通过（2.7分钟）；`cmd/devfixture` Linux编译测试、Windows amd64交叉构建、`web/npm run check`通过。SIGINT清理 fixture 后进程及9443/9444端口无残留。证据见[长时监控报告](../acceptance/reports/monitor-history-soak-2026-09-21.md)，F10/E01矩阵已更新。此项不覆盖Windows运行或跨日采样；UI保持冻结。

2026-09-21 E01 本机真实监控历史长采样：新增默认跳过、需 `BLORA_E01_HISTORY_SOAK=1` 显式开启的 `web/tests/real/monitor-history-soak.spec.ts`。在隔离 HTTPS 双节点 fixture 上按5秒间隔调用真实节点指标121次；每点非stale且时间严格递增，最终历史恰保留最新120点，首点等于第2样本、末点等于第121样本。沙箱外真实Chromium 1/1通过，10.1分钟；`web/npm run check`退出0。证据见[报告](../acceptance/reports/monitor-history-soak-2026-09-21.md)，验收矩阵F10/E01已更新。SIGINT停止专用fixture，进程/9443/9444端口复核无残留。此项只补约10分钟Linux在线采样，不覆盖跨日、Windows、非Chromium或Master重启后持久缓存读取；UI保持冻结。下一步继续处理剩余跨平台与外部环境验收，不重复RC15已通过套件。

2026-09-21 E04/F12 实际 Scheduler.Tick 阶段崩溃：新增 `internal/backup/scheduler_stage_crash_test.go` 与[阶段崩溃报告](../acceptance/reports/scheduler-staged-process-crash-2026-09-21.md)。独立子进程分别在 pending slot 已持久化但尚未 BuildTask、prepared fire/taskId/requestId 已持久化但尚未 Accept 时被 SIGKILL；SQLite/WAL 重开后，前者恰好构造并接受一个任务，后者复用原 taskId/requestId且不重建，重复 Tick 均不重放。定向 race 两场景通过（包1.224s），backup 全包race 8.017s、`go vet ./internal/backup`通过；Windows amd64 测试二进制交叉编译成功。它们覆盖实际调度代码的提交阶段之间，不是设备写入/SQLite事务内部的掉电测试。

2026-09-21 E09 Windows 交叉编译补核：HTTPS Docker E2E 测试辅助代码连同 `internal/containers` 测试包通过 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c`，产物 `/tmp/blora-containers-remote-tls.test.exe`。storage/backup 测试包本轮也均已交叉编译成功。上述只证明测试代码可构建，不是 Windows Docker/Job/ConPTY 真机验收。

2026-09-21 E02/F11 HTTPS Engine 传输子项：E2E 新增 `BLORA_CONTAINER_E2E_REMOTE_TLS=1`，自动生成短期测试CA，在 loopback TLS 代理与本机 Docker Unix socket 间转发；Master Engine 客户端与 Compose CLI 均走真实 HTTPS/TLS。Docker Engine 29.4.1 上生命周期 race E2E 30.49s、包31.52s通过，覆盖容器/镜像/卷/网络、双流日志、Compose更新删除、健康失败、数据保留及tmpfs ENOSPC；随机标签资源通过测试 defer 清理。`go vet ./internal/containers`通过。此项证实 HTTPS 传输与 Compose TLS，不等同远端主机或公网测试；见[HTTPS传输报告](../acceptance/reports/docker-https-transport-2026-09-21.md)。

2026-09-21 E09 Windows 测试构建增量：新增 crash test binaries 均成功交叉编译：`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ... ./internal/storage` 与 `./internal/backup`，产物位于 `/tmp/blora-storage-crash.test.exe`、`/tmp/blora-backup-archive-crash.test.exe`。只证明新增测试代码可编译到 Windows，未执行真机行为。`tc` 不在当前环境，暂不能用 `netem` 做真实内核 RTT/丢包；后续核对可用隔离 Engine/代理是否能提供不重复的真实远端组合。UI保持冻结。

2026-09-21 E03/F12 备份归档中断恢复：新增 [归档进程崩溃报告](../acceptance/reports/backup-archive-process-crash-2026-09-21.md) 与 `internal/backup/archive_crash_test.go`。子进程以真实 `Manager.Create` 打包 64 MiB 源文件，父进程观察 pending ZIP 写入1,048,627B时 SIGKILL。用原 ID 重试后返回 `interrupted/unknown`、未发布快照并清除部分 ZIP；再用新 ID 完成备份且 Inspect 验证通过。定向race通过（2.26s），备份全包race 6.524s及 `go vet ./internal/backup`通过。未模拟掉电/设备缓存丢失或 Windows。下一步继续核查本机可运行的远程/系统环境模拟价值，避免重跑已有浏览器套件；UI保持冻结。

2026-09-21 E04/F12 SQLite 事务进程崩溃边界：新增 [元数据事务崩溃报告](../acceptance/reports/metadata-transaction-crash-2026-09-21.md) 与 `internal/storage/metadata_transaction_crash_test.go`。独立子进程经真实 `Store.MetadataMutation` 在同一 WAL 事务中写 schedule/fire 两项后停在 commit 前，父进程 `SIGKILL`，重开数据库确认业务记录、幂等回执、审计均未部分落盘，既有管理员记录保留，`integrity_check=ok`。定向race通过（用例0.09s），storage全包race 20.570s、`go vet ./internal/storage`通过。该证据只覆盖共用存储事务实现，不覆盖实际 Scheduler.Tick 回调/设备掉电；这些与长期墙钟仍待验。下一步继续挑选可在当前Linux环境真实执行的剩余故障组合；UI保持冻结。

2026-09-21 A09/E01 验收续接：A09 当前源码真实 TLS 慢日志消费者 30 秒逐秒采样通过；未消费字节峰值33,048B/262,144B，接收队列峰值28,320B/2,097,152B，随后跨节点传输、5秒内停止、归档重读与目标校验均通过，`go vet ./internal/master`通过。60秒探索超过协议45秒慢消费者截止期，连接按预期关闭，不计通过；已有测试验证诊断与游标重连，证据见[30秒报告](../acceptance/reports/slow-consumer-soak-2026-09-21.md)。E01 源码与现有真实浏览器证据已核对：节点/实例指标、进程树、120点有界历史、失联 stale、重连采样和搜索/排序/授权终止均已有覆盖；短时本机重复采样无新增价值，跨平台及长期监控专项仍待补。下一步优先在可用平台/远端测试环境补验；UI继续冻结。

2026-09-21 E04 发行包进程级崩溃复验：扩展 `scripts/package-smoke.py --scheduled-backup-crash`，用 RC15 Master/双 Daemon 和真实分钟计划创建定时备份；任务接受后同时 `SIGKILL` Master 与所属 Daemon，再用原数据目录重启。目标 Daemon startupId 更新，原 schedule/taskId/requestId 保持，备份最终SUCCEEDED且归档唯一1份，脚本退出0。初始 harness 的等待函数作用域及 PUT 方法问题均已修正，最终证据见[Master/Daemon崩溃报告](../acceptance/reports/scheduled-backup-master-daemon-crash-2026-09-21.md)。这补足进程级全服务崩溃后的 accepted occurrence 恢复，不覆盖掉电或事务中途断电；长期墙钟和Windows/systemd仍待补，UI继续冻结。

2026-09-21 E02 真实 Docker ENOSPC 子场景：在 `TestRealDockerComposeLifecycle` 中为随机标签的 Compose 服务配置 1 MiB tmpfs，容器实际写入触发内核 `ENOSPC`；当前源码真实 Engine/Compose race E2E 29.17s 通过。确认 apply 返回 `FAILED/health`、exit code 1、`Unknown=false`、保留容器事实与 “No space left on device” 日志，然后通过 Compose delete 清理。`go test -race ./internal/containers -count=1` 在获准 loopback 环境8.248s通过，`go vet ./internal/containers`通过；普通沙箱首轮因 httptest IPv6 loopback 权限失败，不是测试断言失败。报告见[Docker ENOSPC 验收](../acceptance/reports/docker-enospc-2026-09-21.md)，矩阵E02/F11已更新。此项只覆盖容器 tmpfs，不代表 Engine 主机存储耗尽、远程 Engine、掉电或 Windows 容器已通过；UI继续冻结。

2026-09-21 E08 CPU热点优化与一小时复验完成：60秒 CPU profile 将 `recovery/state.ts` 的 JSON 深拷贝定位为主要应用热点；内部恢复值复制改为 `structuredClone`，Vue代理/不可克隆值仍回退 JSON，`json()` 保留原清洗语义。`web/npm run check`、前端单测 **50/50**、生产构建通过；桌面/文件/PTY/编辑器/恢复/云工作区定向真实浏览器 **27/27**。无 profiler 短时真实混合场景 64.954秒 **1/1**，p95 **45.2ms**、max71.9ms、3/3565长帧、双PTY约112kB/s、万条目录/16MiB校验传输完成、未确认峰值79,006B。完整一小时真实混合场景 **1/1通过**：3,604.492秒、48,840个响应样本、p95 **48.1ms**、max160.4ms、731/197,137长帧，8窗口/双PTY/万条目录、165轮传输均完成；队列峰值86,180B，堆35.8–124.6MB。62组RSS/归档采样中57组在浏览器负载期；Master/Daemon RSS峰值39.4/32.0MB，单会话归档最大16,774,920B（低于16MiB预算），无采样竞态。报告与原始指标见[恢复克隆优化小时报告](../acceptance/reports/performance-hour-recovery-copy-2026-09-21.md)和[JSON](../acceptance/reports/performance-hour-recovery-copy-2026-09-21.json)。夹具/浏览器/采样器退出并清理，9443/9444空闲。本机E08一小时p95目标已达标；远端高RTT/丢包、Windows及非Chromium尚未覆盖。此前RC14发行包不含本轮克隆优化；UI按用户要求冻结。

2026-09-21 RC15 交付复核：`BLORA_VERSION=development-20260921-rc15 make package` 退出 0，Linux/Windows Master、Daemon、SDK、Web 六包 `sha256sum -c SHA256SUMS` 全部通过，SHA256SUMS SHA-256 `b5aa3b0c1ec273010b8439521678aa880a1ce133d181320861a670b7f57ada2d`。独立包冒烟在沙箱内首轮受 `esbuild` `EPERM` 阻止；同一测试在沙箱外重跑退出0，SDK/参考扩展签名、Master初始化/TLS/登录、双Daemon、停机快照恢复与兼容RC14回退均通过。见[RC15发行报告](../acceptance/reports/release-rc15-2026-09-21.md)。私有冒烟目录 `/tmp/blora-release-smoke-pmravw8_` 保留，所有服务/测试进程已停止。Windows真机/systemd等外部验收仍未覆盖。

2026-09-21 RC15 当前源码普通真实浏览器全套：`fixture-2757846034`（真实 HTTPS Master/双Daemon/本机 Docker Engine）串行运行49项，排除单独的一小时场景，Playwright **49/49 passed**，退出0，11.5分钟。账号/备份/调度/Docker/Compose/桌面/实例/扩展/文件/PTY/通知/恢复/云工作区均通过；续传从1,179,648B确认偏移恢复16MiB上传且10,000项解压完成。fixture收到Ctrl+C并清理，Master/Daemon/Playwright及9444端口无残留；完整命令与边界见[RC15浏览器回归报告](../acceptance/reports/browser-regression-rc15-2026-09-21.md)。独立E08小时测试另见上条。

2026-09-21 A09 RC15 当前源码定向复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop$' -count=1` 退出0（13.918s）。普通沙箱因 HTTPS 测试监听无法绑定 IPv6 loopback 而失败，获准在沙箱外重跑同一命令后通过。真实TLS慢日志消费者/8.4MB跨节点传输/停止及目标与归档校验子场景当前源码未回归；长时全连接和跨平台资源验证仍缺，见[混合流报告](../acceptance/reports/mixed-stream-control-2026-09-13.md)。

2026-09-21 E04 定时调度强制终止子项：新增独立 Go 测试进程，在计划 occurrence/task/receipt 持久提交后由父进程 `Process.Kill`；另一个进程重开原 SQLite，核对 accepted fire、taskId/requestId、schedule LastTaskID/slot 不变、旧 occurrence 不重放，并确认撤权后下一触发不创建任务。定向 race 1.130s、备份全包 race 2.351s、Windows amd64 测试程序交叉编译通过；最终 `make check`（`go vet ./...`）退出0。全包首轮露出测试输入受当前 `umask 0077` 影响（期望0640、实际0600），已在该测试显式Chmod固定，不改生产行为。证据见[E04进程崩溃报告](../acceptance/reports/scheduler-process-crash-2026-09-21.md)。这不是事务中途掉电或 Master+Daemon 整体硬崩溃验证；物理掉电、长时墙钟、Windows/systemd/远程环境仍缺。UI保持冻结。

2026-09-21 可执行环境与验收矩阵对照完成：当前 runner 为 Linux x86_64；`wine`、`firefox` 不存在，Playwright 缓存只有 Chromium，因此 Windows 真机及非 Chromium 验证不能在此环境执行。`systemctl` 命令存在，但可操作 systemd manager 的缺失已由[环境报告](../acceptance/reports/systemd-environment-2026-09-20.md)记录。RC15 49/49、E08 一小时、A09 定向、E04 scheduler crash，以及本地 Docker、备份/ENOSPC、firewalld、扩展目录等可执行子项均有当前源码证据；未发现应为形式而重复的本机短测。仍缺 A09/E01/E04 的长期与跨平台证据、E02 远端 Engine/长时故障、E03/E04 物理掉电、E05 systemd/Windows 管理器、E06/E07 真实远端运营和 Windows、E08 高 RTT/丢包/Windows/非 Chromium、E09 Windows 真机与 systemd。下一步应直接补入可用的 Windows、systemd 和独立远端测试环境，再针对未覆盖组合验证；若环境仍未提供，则继续做有明确收益的长期故障场景，不重复已通过套件。当前无测试子进程或 fixture 在运行，UI继续冻结。

2026-09-21 RC14 当前源码常规真实浏览器套件：隔离 `--performance --docker-endpoint=unix:///var/run/docker.sock` fixture 上串行运行49项（排除单独的一小时性能测试），Playwright **49/49 passed**，退出0，11.4分钟。覆盖账号/权限、备份/真实分钟调度、Docker/Compose、桌面与实例控制、扩展安装/WASI、文件/PTY/日志/传输、监控/通知、恢复故障注入和工作区等；完整命令及边界见[浏览器回归报告](../acceptance/reports/browser-regression-rc14-2026-09-21.md)。fixture、容器任务和浏览器均已停止；`.local/fixture-1611252974`已清理，定向进程检查无 Master/Daemon/Playwright 残留。此次不覆盖 E08 单独一小时门槛，也不验证 Windows 真机、systemd、远程高延迟或物理故障。

2026-09-21 E08 拖窗优化尝试已回退：13项真实窗口几何测试通过，但真实60秒混合场景在两种直接DOM样式写入变体下 p95 分别为57.0ms和54.9ms，均超50ms；因此未保留该改动。恢复原RAF响应式更新后 `npm run build` 退出0，最终源码继续以RC14 49/49回归所覆盖的窗口实现为准。失败实验和环境见[拖窗优化尝试报告](../acceptance/reports/performance-drag-experiment-rc14-2026-09-21.md)。性能fixture与Playwright均已停止。

2026-09-21 当前源码交付复核：`BLORA_VERSION=development-20260920-rc14 make package` 退出0，Linux/Windows Master 和 Daemon、SDK、Web 六个归档均通过 `sha256sum -c SHA256SUMS`。包级独立冒烟 `python3 scripts/package-smoke.py dist/releases/development-20260920-rc14 --state-restore --rollback-release dist/releases/development-20260920-rc13` 退出0：SDK示例签名、Master初始化/TLS/前端/登录、双Daemon上线、停机快照中身份/权限/任务回执/资源正文恢复以及兼容rc13包回退均通过。服务已停止，无测试进程残留；私有诊断目录保留于`/tmp/blora-release-smoke-h9uvvo0v`。证据见[rc14报告](../acceptance/reports/release-rc14-2026-09-21.md)，SHA256SUMS哈希`46cb8a9d566f0158708e9f68741131aad5c2c96d7731a80ea2b18da7b5c20768`。此包已包含rc13之后的恢复日志/终端快照性能修复，但未代表Windows运行验收或全范围完成。UI按用户要求保持冻结。

2026-09-21 E08 当前源码一小时复验已结束但未通过：3606.506秒、45,072次响应采样，p95 **51.7ms（目标≤50ms）**、max159.4ms、1312/190831长帧；8窗口、双PTY、万条目录和16MiB循环传输运行完成，162轮传输及末轮目标校验通过。未确认队列峰值86,193B，浏览器堆样本36.9–135.9MB，归档最高观测16,775,268B后持续轮转；RSS/归档为定期样本而非瞬时上界。唯一失败为p95断言；不能用旧源码通过结果覆盖。完整环境与数值见[小时报告](../acceptance/reports/performance-hour-rc14-2026-09-21.md)及[原始JSON](../acceptance/reports/performance-hour-rc14-2026-09-21.json)。当前源码两次60秒诊断1/1通过但不计性能验收：p95均47.4ms，分别14/3468、11/3634长帧；慢输入同时来自事件前排队和后续渲染，恢复日志同步`setItem` p95低于1ms，profile热点分散于IndexedDB、xterm、列表更新和GC，详见[诊断报告](../acceptance/reports/performance-diagnostic-rc14-2026-09-21.md)。直接DOM拖窗改动实测p95为57.0/54.9ms，未达到目标并已回退，详见[实验报告](../acceptance/reports/performance-drag-experiment-rc14-2026-09-21.md)。所有浏览器和fixture均已停止。E07 CA根切换定向race已通过（测试1.20秒/包2.240秒）：旧根信任新CA时失败关闭，显式换根后安装新版本；详见[CA根轮换报告](../acceptance/reports/remote-catalog-root-rotation-2026-09-21.md)。下一步继续定位可重复、可量化的非视觉优化；未取得短时达标证据前不重跑小时场景。UI按用户要求保持冻结。

2026-09-20 E08恢复写放大优化：修复 `flush()` 在旧快照后无预算校验写回超256KiB恢复尾日志的竞态；IndexedDB快照追赶期间遇到超额增量时保持未保护提示并立即提交最新状态。新增 `terminal-output` 增量恢复操作，旧 `outputJournal`检查点仍兼容，避免每个终端块复制完整累计输出数组。真实混合负载profiling指出xterm全屏序列化仍是主要热点后，将快照日志预算从64KiB调整为128KiB，仍限定128条事件，逐条输出依旧同步提交，ACK顺序不变。`npm test` 49/49、`npm run check`、`npm run build`通过；fake-indexeddb并发超预算/旧检查点连续追加测试通过；真实WSS PTY刷新/移窗1/1（18.7s），无输入重放。当前干净8窗口/双PTY/万条目录/16MiB传输样本64.63s，交互p95 **45.1ms**、max72.9ms、2/3636长帧、双PTY约114KB/s，性能测试1/1通过。带诊断的上一样本日志`setItem`累计字符由约1.57亿降至2,421万，时间2.96s降至0.57s；诊断扰动不算验收。fixture `fixture-2507872921`、`fixture-1933717844`、`fixture-3066140007`均已停止并清理，当前无测试/fixture进程。UI保持冻结。下一步继续处理其它可完成的功能块与外部验收缺口；E08需在最终源码上补长期复验，Windows/独立远程环境及掉电仍不能由本机短测替代。详细数据见[持续性能报告](../acceptance/reports/performance-continuous-2026-09-19.md)。

2026-09-20 E08 CPU采样跟进：带 profiler/长任务观察器的60秒诊断p95 53.2ms、max127.5ms、3个55/66/59ms长任务、7/3506长帧，API RTT p95 17.5ms；慢样本同时包含事件处理前排队与处理后等待渲染。采样热点为同步 `setItem` 3.09s、xterm `_nextCell` 2.04s/`serialize` 1.89s、IndexedDB `put` 1.70s、`_diffStyle` 1.39s及虚拟列表 `replaceChildren`/`createRow`。它们对应恢复日志、终端缓冲/快照和列表渲染路径，但工具本身扰动测量，不能断定唯一根因或作验收。双PTY约114KB/s、4次核验传输，队列有界且无pageerror；fixture已退出清理。未改生产逻辑或UI。下一步在保留同步可恢复语义及存储失败停ACK保证的前提下，量化RecoveryService每次提交日志写入成本并寻找可安全降复杂度的点；E08仍进行中。见[持续性能报告](../acceptance/reports/performance-continuous-2026-09-19.md)。

2026-09-20 当前源码真实浏览器全套增量：排除独立性能场景后 Playwright 实际收集49项（矩阵旧计数48已修正）。本轮同夹具完整串行47/49通过；两失败分别是监控失联用例未等待首个实时进程列表，以及上传续传用例缺少 `--performance` 数据。监控用例添加“先成功读取实时列表/最近成功时间/进程行”的前置断言后，在真实SIGSTOP/CONT双节点夹具复验1/1（54.0s）；上传关闭续传改用性能夹具后复验1/1（1.2m），1万文件解压和16MiB上传从655,360B确认偏移续传均通过。因此每个常规场景都有一次通过证据，但未在一次全套运行中达到49/49；不为形式目标重复整套测试。两夹具和所属测试实例均已停止/清理。报告索引见验收矩阵；监控证据更新见[E01](../acceptance/reports/monitor-offline-2026-09-19.md)。未改 UI 或生产逻辑。

2026-09-20 E08短时性能诊断复测：将主线程长任务/慢输入采集改为 `BLORA_PERF_DIAGNOSTICS=1` 显式启用，默认保持原低开销路径；`web/npm run check` 通过。带诊断样本p95 54.8ms、最大87.3ms，3个54–73ms长任务，最慢输入主要延迟在事件处理器之前；不作为验收基线。随后全新夹具未启用诊断复跑真实8窗口/双PTY/万条目录/传输60秒混合负载：65.522s、p95 54.4ms（未达50ms目标）、最大80.6ms、14/3478长帧、双PTY约112KB/s、队列峰值67,680B、两次16MiB目标校验通过。唯一失败为p95阈值；无pageerror。真实浏览器/夹具会话均已退出并清理。本轮只改性能测试诊断及证据，不动生产逻辑或UI。见[持续性能报告](../acceptance/reports/performance-continuous-2026-09-19.md)。下一步按矩阵归纳剩余验收，并优先继续当前环境可完成的 Linux 真实组合；E08需定位输入排队来源，原一小时46.5ms证据与本次短时失败并存。

2026-09-20 F09 系统通知真实浏览器补验：新增真实 TLS 双节点任务通知用例；headed Chromium/Xvfb 在授予来源通知权限后收到真实 `Notification` 对象，核对标题/正文/tag，并派发 click 事件后打开精确 taskId 资源窗口，最终1/1（3.1s）通过。首次只因两个“任务中心”标题使通用定位严格模式冲突，收窄至 `data-window-mode="resource"` 后复验通过。报告：[系统通知 API 验收](../acceptance/reports/system-notifications-2026-09-20.md)。虚拟显示无法验证桌面通知中心的实际绘制或物理点击，故F09仍进行中；用户要求暂停的视觉样式未改。fixture进程已退出，两个本轮创建的临时夹具目录已清理。

2026-09-20 E07远程注册源受控故障补验：增强 `TestExtensionRemoteCatalogAdminAPI`，Master 从 loopback HTTPS 源读取清单并安装参考包；每次注册源响应延迟80ms，浏览与安装之间续期同一受信CA签发的服务端叶证书，并关闭空闲连接强制再次握手。清单读取和安装均断言经过受控延迟，成员读取被拒绝；`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestExtensionRemoteCatalogAdminAPI$' -count=1 -v` 退出0，race测试总耗时1.986s。详细范围和限制见[远程注册源验收](../acceptance/reports/remote-catalog-rotation-2026-09-20.md)。这不是公网远端、高RTT、CA根更换或 Master 二进制参数的验收；E07继续进行中。未改生产逻辑或 UI。

2026-09-20 当前源码 rc13 回归：按用户要求暂停 UI 样式工作，只给新版应用市场卡片/已安装项和窗口增加非视觉 `data-app-id`/`data-window-mode` 识别属性，并将真实测试同步到现有标题、版本和账号菜单。`npm run build` 通过；真实双节点常规浏览器套件（排除独立一小时性能场景）48项首轮43项通过、5项因旧断言/测试数据隔离失败，修正后受影响6项定向复验6/6通过，因此48个常规场景均至少各有一次通过证据，但没有单次48/48的全绿套件运行。`development-20260920-rc13` 六包构建、SHA校验、停机状态恢复及兼容rc12回退冒烟退出0；Docker浏览器2/2通过。fixture和测试Docker对象已清理。详见[rc13报告](../acceptance/reports/rc13-current-source-2026-09-20.md)。矩阵整体仍未完成；下一步只能在真实 Windows、独立 systemd、远端故障及物理掉电环境补证，不能以本地替代测试覆盖。

2026-09-20 最新现代 Material 3 Expressive 整理：壁纸替换为青灰抽象曲面，主题改为中性雾灰/矿物青绿并收敛图标色彩；统一菜单、弹窗、表单、卡片和编辑/终端外框，清理 Docker 局部旧色。修复 JS 启动器未遵循减少动态效果的问题，新增双来源偏好实测。完整浏览器38/38通过（1.9分钟），最后焦点样式修正另3/3通过（10.8秒），最终构建通过；真实截图34张，13组主要布局无差异，Dock四边14px。见[现代视觉报告](../acceptance/reports/material-modern-2026-09-20.md)。源码/web/dist最新；所有进程结束，fixture 83928退出0；旧rc11未重打包。后续以2026-09-20-material-modern为视觉基准，保留布局；外部平台验收缺口及第三方应用内容边界见报告。

2026-09-20 最新 Material 3 细化：联网核对官方 Shape/List Expressive/Checkbox 文档，补齐内容区、表头、分组列表与复选框的圆角层级，调整备份/终端/设置/市场图标配色。构建通过，相关浏览器17/17通过（53.6秒，API替身），真实本地截图24张及空格选择/取消验证通过；13组主要布局无差异，Dock四边14px。见[细化报告](../acceptance/reports/material-refine-2026-09-20.md)。源码及web/dist最新；所有构建/测试/截图会话结束，fixture 85791退出0，旧rc11未重打包。下一步视觉调整以2026-09-20-material-refine为基准；保留布局、多色扁平图标及既有外部验收边界。

2026-09-20 最新 Material You 重构：按用户“布局不变、整套改扁平、多色图标”的要求，新增统一颜色角色，改写桌面/窗口/侧栏/菜单/表单样式及整套图标，删除玻璃材质文件，替换平面几何壁纸。最终构建退出0，完整浏览器37/37通过（1.8分钟，API替身），真实Master/双Daemon截图22张。13组桌面/手机区域稳定态布局比较无差异，Dock四边14px、曲线误差小于0.01px；首次布局采样的入场动画差异已保留并修正采样后复验。见[本轮报告](../acceptance/reports/material-you-2026-09-20.md)。源码和web/dist最新，构建/测试/截图全部结束，fixture 78272退出0，旧rc11未重打包。下一步视觉调整以2026-09-20-material-you为基准，保留用户要求的布局；全范围外部验证仍按已有环境缺口处理。

2026-09-20 最新薄片图标修订：按用户新提供的照片/App Store/邮件近景，去除此前被否定的追光、光晕和厚重高光，重新绘制薄片交叠前景、平缓底色及轻微接触阴影。最终构建退出0，控制栏/桌面浏览器7/7通过（44.6秒，API替身）；真实隔离Master/双Daemon截图22张，7种场景的Dock四边均14px、曲线误差小于0.01px。见[本轮报告](../acceptance/reports/thin-glass-icons-2026-09-20.md)。源码和web/dist已更新；所有本轮构建/测试/截图会话结束，fixture 94030退出0，旧rc11未重打包。下一步若继续视觉调整，以2026-09-20-thin-glass截图与源码为基准；全范围外部验证仍按既有环境缺口处理。

2026-09-20 最新圆角对齐：依据 Apple 连续曲率/同心嵌套文档，应用图标和 Dock 共用连续轮廓，外框沿图标轮廓法线等距外扩14px；移除透明SVG留边对Dock布局的干扰，上下左右可见间距统一，悬停和窄屏保持关系。最终构建通过，控制栏/桌面浏览器7/7通过（38.5秒）；真实fixture直接测量7种场景的渲染轮廓，四边均14px、曲线偏差小于0.01px，无畸变或越界。截图5张与数值见[圆角报告](../acceptance/reports/concentric-corners-2026-09-20.md)。源码及web/dist已更新，截图和fixture会话70529已结束，旧rc11未重打包。下一步如继续视觉调整，以此共享轮廓与截图为基准；全范围外部验证仍按原环境缺口处理。

2026-09-20 最新轻拟物修订：按用户最新图例重写应用图标为简洁的玻璃叠层，线条工具图标改用精确锁定的 @lucide/vue 1.47.0；账号菜单文字增加1px、宽度240→232px、行高37→35px，内容间距小幅收紧。最终生产构建通过，选定控制栏/桌面/文件浏览器11/11通过（52.9秒，API替身）；真实本地Master/双Daemon截图17张，见[本轮报告](../acceptance/reports/glass-icons-2026-09-20.md)。源码与web/dist已更新，旧rc11归档未重打包；截图浏览器和fixture会话44040已退出0，没有待接管测试进程。下一步若继续调视觉，以本轮截图和源码为准；全范围剩余验证仍需先前记录的专用外部环境。

2026-09-20 最新桌面视觉修订：按用户对顶栏/旧macOS仿制感的反馈，新增独立DesktopBar组件和三列排版，统一工具图标尺度，账号菜单收纳退出/设置；移除三色窗口控制，重写窗口与Dock光影、背景及应用材质，调整共用窗口工作区域。build和47单测通过；本轮选定16浏览器场景经15+1修复复验通过。真实fixture截图14张，见[本轮报告](../acceptance/reports/modern-desktop-2026-09-20.md)。新源码与web/dist已更新，截图fixture交付前停止，旧归档未含本轮UI。

2026-09-20 原生桌面 UI 重写：依据用户后续反馈移除旧 style.css 内容及 desktop-light.css 补丁层，重写桌面/Dock/窗口/应用样式；具象SVG应用图标、左侧三色窗口控制、无常驻标题Dock、资源侧栏与紧凑列表已落地。构建与47单测通过，浏览器首轮35通过+同名入口定位修复后1复验通过，36个场景全部验证。真实隔离Master/双Daemon截图11张及证据见[UI报告](../acceptance/reports/native-ui-2026-09-20.md)。本轮测试进程已结束，截图fixture在交付前停止；源码与web/dist最新，rc11归档仍是旧UI。下一步如用户要求打包，基于当前源码生成新版归档；原环境阻塞项未改变。

2026-09-20 UI 重构：按用户反馈改为浅色桌面、蓝色操作重点、白色窗口与居中浮动任务栏；移除桌面营销文案，统一桌面/启动器/标题栏 SVG 图标，Monaco 使用浅色主题。原桌面 slice(0,6) 隐藏了扩展入口，现显式固定七个入口并将扩展管理更名为应用市场；市场展示真实注册源版本、已安装应用与本地安装，保留管理员权限和安装生命周期。当前源码及 web/dist 已更新，旧 rc11 归档未重新打包。

本轮验证：`cd web && npm run build` 退出0；单测47/47；`npm run test:e2e -- tests/browser/desktop.spec.ts tests/browser/extensions.spec.ts --workers=1` 最后整组11通过、1因上传输入框名称变化失败，恢复 aria-label 后以 `--grep 'independently packaged'` 单独复验1/1通过（合计12个场景通过，并非一次整组12/12）。初轮应用按钮模糊定位冲突已改为启动器专用定位。Chromium连接真实本地Master/Daemon fixture，在无API模拟的情况下保存[九张截图](../acceptance/screenshots/2026-09-20-redesign/)，脚本为 `web/capture-ui.mjs`；市场中的条目为fixture参考扩展，不代表公共应用生态。最后截图后只新增上传框可访问名称，不影响画面。原Windows/systemd等环境验证缺口保持不变；后续如交付安装包，需基于此源码重新打包。

最新交付rc11：包含当前生产源码、前端回归修复、Docker 真实复验及同步后的验收台账；六包外部SHA/归档内MANIFEST、独立SDK签名/Master初始化TLS/双节点冒烟全部通过，包级冒烟退出0。SHA256SUMS 哈希与六个归档逐项结果见[rc11报告](../acceptance/reports/release-rc11-2026-09-20.md)。

2026-09-20 续接审计：当前树重新执行 `make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及 47/47 单测退出0。源码扫描未发现生产路径空实现或遗留 TODO；当前无活动测试、服务或包冒烟进程。

2026-09-20 文档证据审计：检查 112 份 Markdown 的相对链接，缺失链接为 0；修正一份历史报告对已清理截图的失实链接并明确其证据边界。

2026-09-20 当前源码真实 Docker 复验：确认 Engine 29.4.1/overlayfs 与隔离 fixture 可用，`BLORA_CONTAINER_E2E=1 go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1` 退出0（27.681s）；真实容器、镜像、卷、网络、双流日志、Compose 更新/删除、健康失败、卷保留和普通用户拒权均通过，带标签对象已清理。见[当前 Docker 验收](../acceptance/reports/docker-current-2026-09-20.md)。

2026-09-20 systemd 环境复核：宿主 systemd user bus 返回 `Operation not permitted`；不挂载宿主 cgroup 的隔离 Debian systemd 容器退出255，特权宿主 cgroup 方案被安全审查拒绝，未执行。未修改宿主服务；真实 systemd 生命周期继续记录为环境阻塞，见[systemd 环境报告](../acceptance/reports/systemd-environment-2026-09-20.md)。

2026-09-20 浏览器回归收口：修复三个真实失败后，`web/npm run test:e2e -- --workers=1` 完整 36/36（2.5分钟）通过；实例批量、跨实例传输和扩展跨设备场景均复验，见[默认应用浏览器回归](../acceptance/reports/browser-regression-2026-09-20.md)。

2026-09-20 当前源码完整竞态回归：提权环境执行 `make test`（`go test -race ./...`）退出0；containers 8.485s、extensions 152.757s、master 277.018s、protocol 1.230s、runlog 2.279s、runtime 3.097s及其余Go包全部通过。受限沙箱首轮仅因IPv6回环监听和helper子进程限制失败，未计作代码失败。见[全仓竞态报告](../acceptance/reports/full-race-2026-09-20.md)。

2026-09-20 续接复验：`make check` 退出0；`make windows` 重新生成 Linux/Windows 交叉构建产物并退出0；`web/npm run check && npm test -- --run` 的类型检查和 47/47 单测退出0。未启动常驻服务或留下测试进程。

2026-09-20 定时备份一致性补齐：Master 调度校验现接受 `save`、`pause`、`stop`、`hooks`，校验固定钩子引用格式，保持 `files` 模式禁止钩子引用；真实调度触发的保存程序已验证前后钩子、资源身份、最新正文归档和任务回执，定向集成测试及 race 均退出0（7.625s）。

默认备份应用的定时任务表单已同步暴露一致性策略及前后钩子引用，编辑时回填已有参数；类型检查、47/47 单测和备份界面浏览器 1/1（5.0s，实际断言 save 引用进入请求）通过，增量进入 rc7。

扩展用户数据迁移复验：真实 WASI guest 双用户迁移、失败保留旧包、回滚迁移、事务中断恢复和迁移期间并发写入保护通过，`go test -race ./internal/extensions -run 'TestUserDataIsolationCASMigrationAndRecovery|TestMigrationDoesNotBlockRegistryAndRejectsConcurrentWrite' -count=1` 退出0（85.404s），见[扩展数据迁移报告](../acceptance/reports/extensions-data-2026-09-20.md)。

最终会话检查点：实际分钟调度完整复跑 **44800退出0，1/1（1.6分钟）通过**，每槽唯一任务、新来源版本和两份归档恢复正文均核对成功。首次78155仅因新增测试误期望201为200失败，已保留记录。所属计划已删除，fixture **98737退出0**。当前无活动测试或服务会话待接管；小时35594和冷构建已完成，不重复运行。见[实际分钟调度](../acceptance/reports/schedule-wall-clock-2026-09-19.md)。

补充 F02/A01：扩展 `workspaces.spec.ts` 真实终端检查点内容策略并加入双浏览器设备冲突场景，fixture **1902302221** 会话 **27729** 2/2（6.9s）通过。默认布局同步剥离 `terminals`，显式工作内容同步携带真实终端检查点；第二设备先提交 revision 2，第一设备旧 revision 同步被拒绝并显示冲突，本地未同步 Monaco 正文保留。fixture 已停止。见[多设备冲突](../acceptance/reports/workspaces-2026-09-19.md)。

续接后的本地回归：`make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及47/47单测通过；提权隔离 `go test ./internal/systeminfo ./internal/runtime ./internal/terminal ./internal/runlog` 退出0；工作区后端 `go test -race ./internal/master -run 'TestCloudWorkspace' -count=1` 退出0。受限沙箱首次运行的回环绑定失败未计作代码失败，已用隔离权限复跑并通过。当前无活动测试、fixture或构建会话。

E09 平台适配修复：发现并修正 `internal/runtime/native_windows.go` 多核 Job Object CPU 配额被错误覆盖的问题，新增 `internal/runtime/cpu_quota_test.go` 锁定一核/八核/小配额/超容量边界。`go test -race ./internal/runtime`、`make check`、`make windows` 与 Windows runtime 测试二进制交叉编译均退出0。Windows 真机仍未提供，不能将该证据称为运行验证；修复随后已进入 rc5，rc4 保留为修复前的历史包。

最新交付已升级为 **rc5**：`BLORA_VERSION=development-20260919-rc5 make package` 退出0；六包 SHA256SUMS 逐项校验通过（文件哈希 `6a683cfbd4433036e2f088cf4e64673f6d5139e0231d7ea9697fce6cdabe73d5`），独立 `package-smoke.py` 会话 **60112退出0**，SDK/参考包签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均通过。见[rc5报告](../acceptance/reports/release-rc5-2026-09-20.md)。当前生产修复已进入 rc5；Windows 真机与 systemd 等外部环境缺口仍未关闭。

全范围仍不能宣布完成：Windows真机、独立systemd服务/计时器生命周期、独立远程环境与物理故障等矩阵未验证项仍保留。下一条具体动作是取得用户提供的专用测试环境访问配置后，运行对应平台验收入口并补证；已异步询问，尚无回复，不触碰当前生产宿主。以下为按时间倒序保留的历史过程记录，旧“当前”句不代表活动会话。

最新终态与当前任务：第二小时 **35594退出0**，p95 **46.5ms**、60528次交互、220次轮换与末轮校验通过，双PTY全程持续；945.7ms最大延迟保留，100/207818长帧。优化后完整JSON/61次资源样本/心跳已写`performance-hour-optimized*-2026-09-19.json`，10351采样自然退出0，31985夹具退出0。rc3对应小时证据通过，不覆盖其他平台。

后续E01监控真实失联首轮 **22756退出1（54.8s）**：指标stale/时间/诊断正确，但缓存进程页仍显示“实时列表”。已修正`MonitorApp.vue`：旧进程列表明确提示、显示最近成功读取时间、旧列表禁用终止入口和翻页。新增测试要求非空旧列表禁用按钮且恢复后转实时。build通过。当前复验 **17778**，fixture **17481**（`.local/fixture-3280126874/browser-credentials.json`），继续原复验；此前finally已恢复Daemon并停止实例。此新增生产修复不在rc3，复验/监控浏览器回归通过后需统一rc4并独立冒烟，不为仅报告结果循环重打包。小时性能路径未再更改，无需对监控文案单独再跑一小时。

第二轮小时当前心跳2064秒、128次轮换正常，主测试 **35594** /采样 **10351** /fixture **31985** 不变，预计15:17 UTC结束。33次已读样本内归档未超预算，最新浏览器堆44.6MB；不是最终通过。继续poll同一35594并将JSON追加functions.store相应2180714798键。生产源码/静态构建未再变化；新增monitor-offline真实测试仍未执行，必须等此小时终态后在新夹具运行。若最终p95仍超50ms，保留原数据并继续定位，不放宽阈值。

小时等待期间仅补测试源码：新增`web/tests/real/monitor-offline.spec.ts`，尚未运行。使用私有fixture精确配置路径+PID出生身份定位所属Daemon，实际SIGSTOP/CONT验证节点与实例缓存时间/诊断/原runId、另一节点仍采样、恢复后清除旧数据标记，finally恢复并停止所属实例。另断言失联时缓存进程列表不能继续显示“实时列表”；源码MonitorApp目前无条件显示该标签，可能需真实复现后修复，不能把新测试算通过。须在35594小时终态并停止旧夹具后，开新fixture执行此测试，不能暂停当前性能节点。当前生产代码/静态构建/rc3均未因这段测试而改变。

当前最新会话（2026-09-19 14:17 UTC）：优化后小时复验 **35594**，fixture **31985**（`.local/fixture-2180714798/browser-credentials.json`），RSS/归档只读采样 **10351**。原普通权限采样61436看不到/proc进程，已停止130，改为提权只读采样；主测试未重启。命令`BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，无CPU profiler、默认0.01秒输出、原50ms断言。预计15:17 UTC终态，必须继续原35594，不构建web/dist、不并行重负载。functions.store采样键`blora-perf-progress-2180714798`和`blora-perf-resources-2180714798`；终态保存第二轮独立JSON，再停10351/31985。

当前发行已经完成：`development-20260919-rc3`包含终端合批/关闭码/提示、Linuxunit边界、原生快照复制和依赖通知。六包外部hash/内部每文件manifest/许可文本逐字节全部通过；SHA256SUMS哈希`263567cb00344b3c6686fc1e08517bae1811341d916ea35e4ed657869cbfab38`。独立解包SDK/签名/双节点以及停机恢复分别使用rc3自身和兼容rc2均退出0，5671/18622均已终止并清理所属进程，详见`release-rc3-2026-09-19.md`。不用为了补写该包报告循环重新打包。

快照优化回归：check、47项单测、build通过；真实恢复配额/schema/指针事务中止3/3（7.8s），桌面编辑器终端组合8/9后末项额外导航导致故障注入失效（疑似同时另一测试报告触发开发服务器监听，保留失败），隔离终端2/2（17.6s）通过。优化前60秒CPU JSON复制7.291s，优化后2.059s、p95 45.4ms、堆36.3～79.4MB，短测不覆盖小时失败。旧23641/98648夹具均已停止，53406/28190诊断已结束。第一小时失败50.3ms的完整JSON及60次样本已保存，不能覆盖或改称通过。当前仅35594/10351/31985活动。

小时86432已退出1：全程3600秒/180次轮换、双PTY持续输出与末轮目标校验完成，唯一失败p95 **50.3ms > 50ms**。不得降低阈值或计通过。完整JSON、60次资源样本及已保留心跳已写`docs/acceptance/reports/performance-hour*-2026-09-19.json`；旧fixture96517退出0、采样86738退出130。新增终端双存储失败不ACK断言已实际2/2（18.0s）通过，57397退出0。当前定位性能：新fixture **23641**（`.local/fixture-2836489518`），CPU采样测试 **53406**（60秒，`BLORA_PERF_PROFILE_SOAK=1`），继续原会话取CPU摘要，结束后停止夹具。新开关仅测试诊断，非生产改动。发行rc3/通知验证/停机恢复与兼容rc2回退仍待执行；暂先定位50ms未达标原因。

当前有效小时测试仍为 **86432**，夹具 **96517**（`.local/fixture-2713429467`），只读采样 **86738**。最新2435秒/126次传输轮换正常，双PTY各约6.5MB/分钟，41次资源样本内归档未超16MiB。尚未到终态，不能计通过；继续原会话，不重建web/dist或另起重负载。终态后保存完整JSON及functions.store中的progress/resources数组，再停采样与夹具。随后运行最新terminal.spec.ts双存储失败不ACK断言，统一构建rc3并校验六包与通知，运行`package-smoke.py --state-restore --rollback-release dist/releases/development-20260919-rc2`。恢复脚本语法和参数拒绝检查通过，但实际恢复尚未运行；运维手册已补入口。异步询问的Windows/systemd专用环境尚无回复。

E09后续入口已写、尚未执行：`scripts/package-smoke.py --state-restore [--rollback-release <兼容旧包目录>]`，只管理自身私有目录，停止全部所属进程后复制Master/双Daemon及配置；重开后制造用户/文件变更，再停止、保留变更后状态、恢复旧快照，检查TLS/节点身份、用户登录、只读授权/控制403和文件正文。禁止与crash参数混用。三份Python脚本语法检查通过，**未冒充真实恢复演练通过**。小时86432在1156秒仍正常（63次轮换），86738采样/96517夹具保持。先完成小时与新增存储失败浏览器断言，再统一新包并运行该恢复入口（可用rc2作为兼容回退目标）。

E09独立进展：新增离线`third-party-notices.py`并接入Makefile/package.py，真实收集104条依赖通知（17Go模块+Go安装包+86npm包）、约490KB，缺少依赖/许可证即失败；README和`dependency-notices-2026-09-19.md`记录来源与边界。未运行重打包，不影响小时性能测试。当前86432在911秒仍正常、53次轮换；96517夹具和86738采样保持。小时结束后还需运行新增双存储失败不ACK浏览器断言，并统一打包核验通知与本轮生产修复；rc2均未包含这些增量。

最新427秒心跳：28次传输轮换，双PTY各约6.9MB/分钟，暂无错误；归档首段约20900，16MiB预算内。小时会话86432/采样86738/fixture96517不变。另已向`web/tests/browser/terminal.spec.ts`追加双存储写入失败时不ACK新批次的断言，**尚未运行这段最新断言**；小时结束并停止夹具后，运行`npm run test:e2e -- tests/browser/terminal.spec.ts --workers=1`，再统一验证/发行本轮生产增量。此前2/2（18.3s）仅覆盖追加前的版本，不冒充新断言已通过。

小时复验246秒心跳正常：17次已验证轮换，两路每分钟约7MB；四分钟资源采样归档各16.2MB且首段推进，Master约34.5MB RSS，两Daemon约27MB。主测试86432、采样86738、fixture96517继续运行。`functions.store`内累计记录键为 `blora-perf-progress-2713429467` 和 `blora-perf-resources-2713429467`；后续poll输出可按JSON追加，最终汇总实际采样峰值，不能把过程记录当通过。已异步询问用户是否有专用Windows/systemd测试环境，尚无答复，本机工作继续。

小时复验附属采样会话 **86738**：`python3 scripts/performance-resources.py .local/fixture-2713429467 --samples 62 --interval 60`；仅只读RSS/归档，每分钟输出JSON。主测试 **86432**、夹具 **96517**。终态后先保留结果，再停止86738与96517。README已补可复现长时入口；不为仅文档变化重新构建。

当前唯一小时复验：Playwright **86432**，夹具 **96517**，目录 `.local/fixture-2713429467`，命令 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`（默认0.01秒输出间隔）。包含终端合批、合法关闭码、断流状态提示与逐轮错误检查。上一高输出51451仍在3.1分钟OUTPUT_GAP失败，吞吐约三倍但不能计通过；fixture73328已退出0。最终前端build和终端浏览器2/2（18.3s）通过。继续原86432至终态，不重建静态资源、不重复启动、不因观察超时重启。成功后保存完整JSON/资源采样，再停止96517；失败则按诊断修复。源码增量尚未发行。

终端消费修复当前检查点：高输出10205在1.7分钟明确失败`OUTPUT_GAP`，旧fixture12256已退出0。现TerminalApp合并相邻DATA（64KiB载荷），TerminalModel合并相邻输出解析/保护并保留resize顺序与原序号，累计ACK仅在保护后发送；原缓冲/归档预算不变。关闭码4002修复同在源码。前端build通过，突发输出/resize/刷新序号90/精确ACK浏览器2/2（17.0s）通过。当前真实高输出复验 **51451**：180秒、输出间隔0.001；fixture **73328**，`.local/fixture-3865775071/browser-credentials.json`。继续原51451，终态后记录并停止73328；不要重建静态资源。报告 `terminal-batching-2026-09-19.md`。成功后仍需默认速率长时复验；本轮前端生产增量和之前unit边界均未进入rc2。

最新诊断（2026-09-19）：600秒复现20194退出1（10.4分钟），33次传输轮换、p95 42.8ms，但两次浏览器异常明确指出 `WebSocket.close(1002)` 为脚本不允许的关闭码。已将TerminalApp/InstanceConsole错误关闭改为4002，终端浏览器回归新增服务端OUTPUT_GAP后检查错误显示/只读/无pageerror，2/2（14.8s）通过；生产前端build通过。该修复尚未定位原始断流原因。旧fixture11505已退出0。当前新fixture **12256**，`.local/fixture-2178575800`；诊断测试 **10205**，`BLORA_PERF_SOAK_SECONDS=180 BLORA_PERF_OUTPUT_DELAY_SECONDS=0.001`（提高输出速率加速积压，不能代替默认速率小时验收）。新增逐轮检查终端错误和有限JSON附件，防止末分钟漏检；下一步接管10205查看原始错误，再修复。不要重建静态资源或启动重复负载。当前源码两个WebSocket关闭码修复也未纳入rc2。

600秒故障复现已启动：Playwright **20194**，fixture **11505**（`.local/fixture-30796204`）。65秒心跳正常，两路各约3.89MB/分钟，3次传输轮换；等待同一20194的终态与`BLORA_PERF_TERMINAL_FAILURE`诊断，不重复启动。原小时6442确已失败并停止对应夹具。新测试不修改生产代码、不放宽持续输出断言。

最新失败与接管（2026-09-19）：小时测试6442在9.3分钟退出1，双PTY连续一分钟接收增量为0，持续输出断言失败；最后成功心跳492秒、27次传输轮换。节点session仍running、归档首段继续前进，不能归因于负载生成器已结束，也不能计通过。已停止旧夹具78200（退出0）及采样20793（退出130）；旧句柄不要接管。新增失败时仅采集终端状态/错误和pageerror的诊断，不保存凭据或完整终端正文。新夹具 **11505**，目录 `.local/fixture-30796204`；正在启动600秒定向复现，凭据文件路径该目录下`browser-credentials.json`。下一步读取复现终态诊断、定位浏览器/流/恢复保护故障并修复；不是忽略失败直接重跑小时。只读采样脚本已实际运行并验证归档轮转。全范围仍未完成。

当前资源观测（2026-09-19 12:35 UTC）：新增只读入口 `python3 scripts/performance-resources.py .local/fixture-4166028878 --samples 60 --interval 60`，采样会话 **20793**，与原性能会话 **6442**、夹具 **78200** 并行。前两次磁盘采样显示两路归档各15.8–16.3MB、首段序号从约14544前进至20415，均小于16MiB；Master RSS约33–35MB、两Daemon约26–28MB。性能最新372秒、19次已验证轮换、双PTY持续新增输出。以上均为过程证据，不计小时通过。下一步继续原6442和20793，不重建前端、不启动重复负载；6442终态后保存结果，再停止20793及78200。单元边界生产修复尚未纳入rc2。

修正后的小时级首个心跳：会话6442在67秒仍运行，已完成2次传输轮换；两路PTY该观测区间分别新增6,224,580/6,150,304B，持续输出断言通过，堆53,849,655B、未确认峰值87,404B。尚未结束，继续原句柄；该进度不计小时级通过。

当前有效小时级运行：新fixture-4166028878已就绪，夹具会话 **78200**；修正后Playwright会话 **6442**，命令 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，私有凭据路径 `.local/fixture-4166028878/browser-credentials.json`。继续轮询6442，每分钟进度必须有两路terminalIntervalBytes；不要启动第三份测试、不重建静态资源。终态后保留最终JSON/失败信息并停止78200。旧91740/63313已明确终止，不再接管。当前尚无小时级结果。

纠正小时级负载：发现PTY生成器固定12000次输出会提前结束。已主动中止91740（退出130）并停止63313（退出0），不是因观察超时重启；旧fixture-950062738的约249秒不能计小时混合负载通过。测试现按soak时长设置有限输出次数，并每分钟断言两路PTY都有新字节。正在启动全新私有fixture重跑；旧句柄不再轮询。只读旧夹具内存样本约Master33.6MB、Daemon26.6/28.3MB，非最终小时证据。

小时级验收最新已观测进度：会话91740在66秒输出3次已完成传输轮换、44次活跃检查、当前堆53,724,785B、未确认峰值87,400B；进程仍在运行，非终态。保留原会话继续等待，不算一小时通过，也不重启负载。

小时级测试接管句柄：Playwright会话 **91740**，夹具会话 **63313**。继续轮询91740；测试终态后先保留JSON结果与失败诊断，再停止63313。不要重复启动测试或在其运行期间重建/替换前端静态资源。尚不能计小时级通过。

当前活动E08验收：持续传输入口已改为每轮成功并核验目标后，以目标当前版本覆盖同一测试文件，避免单次传输提前完成后继续空载采样；soak上限3600秒，每分钟打印无凭据进度。30秒真实Chromium通过（59.4s整测），两次传输轮换、p95 42.8ms、队列未确认峰值87,400B、堆38.3–92.0MB。旧fixture-816325812已退出0。新fixture-950062738已就绪，夹具会话63313，私有凭据 `.local/fixture-950062738/browser-credentials.json`；已启动 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，一小时结果尚未产生，必须继续轮询原测试句柄，不能将观察超时当失败或另起重复负载。生产代码无新增变化，上一轮单元边界修复仍未纳入rc2。

最新生产修复（2026-09-19）：Linux服务动作此前可传入`.target/.mount/.socket`，timer动作可带文件路径；现将执行适配器限制为明确`.service`和无路径`.timer`，Windows名称规则不变。修改 `internal/systeminfo/service_actions.go`、`tasks.go`，新增Linux适配器与真实TLS双Daemon边界测试，运维说明同步完整unit名。systeminfo全模块race1.147s、原API权限1.759s、新Daemon拒绝边界2.688s通过，所有测试会话已退出。测试空PATH不调用宿主systemctl；不能把该证据代替真实systemd生命周期。本轮为实质进展；修复尚未纳入rc2，后续统一发行时须包含，勿称rc2已有此边界。下一步继续E05独立systemd验收环境/E08长期负载，或汇总生产增量后统一发行验证。全范围目标仍未完成。

最新续接结果：私有firewalld验收新增显式确认与实际Daemon重开组合，通过race32.103s；确认后两套规则保留，普通用户403，同键回放200、异键409且终态修订不变，最终恢复原快照。会话99317已退出0，测试所属进程/挂载/网络已清理。改动仅 `firewall_namespace_linux_test.go` 与验收记录；生产代码及rc2不变，无需重复打包。上一轮及本轮均为实质进展，目标保持未完成。下一步继续E05独立systemd服务/计划任务环境或E08持续混合负载；Windows、物理掉电、远程及小时级负载缺口仍保留。

最新检查点：定时备份接受后真实Master/SQLite关闭重开、双Daemon重连、任务身份保持和单一归档race3.198s通过（`TestScheduledBackupMasterReopenKeepsAcceptedOccurrence`）。会话89276已退出0，无活动验收会话。矩阵与README已加入真实ENOSPC、钩子、离线调度和私有firewalld入口，`go vet ./internal/master`通过。未确认可用独立systemd环境，不访问宿主服务/计划任务；Windows、物理掉电、远程和小时级负载仍未验证。后续从E05独立系统管理环境或E08持续混合负载入口继续，保留rc2及本轮源码测试，不重复初始化、不为文档记录循环打包。

当前总检查点（2026-09-19）：全范围仍未完成。生产修复（控制接收队列饱和背压、此前SDK请求键与查询恢复）已纳入rc2且独立解包验证通过；后续新增真实ENOSPC、save钩子、节点离线调度、私有firewalld断链回滚测试均通过并留在源码。矩阵A12更新为已实现并明确Linux证据；E03/E04/E05继续按子项记录。Windows真机、系统服务/计划任务实际主机组合、物理掉电、远程与长期负载仍有缺口。无活动会话待接管，不为仅同步本条记录重打包。下一步针对E05系统服务/计划任务检查可用独立systemd环境；如无法提供真实管理器，保持未验证并继续E08/E09可用场景，不调用当前生产宿主服务。

最新E05验收完成：私有user/mount/net namespace + 独立D-Bus/firewalld + 真实TLS Master/双Daemon，防火墙runtime/permanent差异、未托管范围保留、内核规则核对、实际阻断Master端口后期限回滚及解除后FAILED诊断均通过；race25.998s。普通用户403。首两次环境启动失败已修正并记录，见[真实防火墙报告](../acceptance/reports/firewall-private-2026-09-19.md)。会话11284已退出0，所属服务/挂载/规则已清理，无活动测试会话。当前rc2生产代码未再变化；新增ENOSPC、钩子、离线调度、防火墙测试在源码。E05系统服务/计划任务和Windows、物理故障、长期远程组合仍待核对，下一步完善这些增量在矩阵与运行文档中的状态。

当前活动验收：E05私有防火墙环境首次运行，会话81768；`BLORA_TEST_FIREWALL_NAMESPACE=1 go test ./internal/master -run '^TestPrivateFirewalldApplyAndRestore$' -count=1 -v`。新增 `firewall_namespace_linux_test.go` 在显式启用后自启动unshare user/mount/net子进程，核对父子namespace不同、私有/run隐藏宿主总线，再启动任务所属dbus-broker/firewalld。测试真实端口应用、内核nft规则、runtime/permanent区别和回滚；结果尚未确定，不能计通过。工具仅探测私有namespace能力成功，没有调用宿主firewall-cmd。

合并复验完成：启用 `BLORA_TEST_ENOSPC=1` 的备份/调度/传输空间不足race组合退出0，20.324s，覆盖新增钩子成功/失败、真实目标空间不足、真实节点离线补跑/撤权及原有接口。没有新增生产代码改动；rc2仍是当前发行产物。下一步评估E05能否用私有net/user/mount namespace + 独立D-Bus/firewalld实现安全的真实防火墙演练；现有宿主防火墙绝不调用，尚未启动该环境。

E03钩子增量：固定argv独立保存程序接入真实Master/双Daemon，before原子发布新正文、after确认；归档恢复得到保存后的正文，失败分支明确FAILED并补偿，同键不重复钩子。race6.174s通过，新增 `backup_hooks_integration_test.go`。正在运行启用真实ENOSPC的备份/调度合并race回归，命令 `BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'Test(Backup|Schedules|TransferTargetENOSPC)' -count=1`；完毕记录结果。此前发布rc2生产代码未再变化，新增测试单独留在源码。

E04增量已完成：新增真实Daemon离线/重连调度组合，受控时钟五次到期不积累任务，恢复只补跑一次且实际备份最新源版本；离线撤权则无任务并报告authorization_denied。加强账本零任务断言后race2.869s通过，证据已写入备份调度报告。当前没有运行会话；下一步继续E03固定命令save前后钩子的真实备份/恢复全链路，已有runner和库测试不替代该组合。

最新验收（2026-09-19）：A12/E03真实ENOSPC已补齐Linux子项。私有user/mount namespace内的1MiB tmpfs：跨节点移动失败保留源并清理暂存，备份仓库满不发布归档，恢复目标满保留当前正文/其他文件/原备份。传输+备份创建race组合2/2、6.505s；恢复race3.678s通过。新入口 `BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'ENOSPC' -count=1 -v`，代码为两个 `*_enospc_linux_test.go`；未启用时明确skip，不算通过。报告见[真实空间不足](../acceptance/reports/enospc-2026-09-19.md)。当前无活动测试服务或挂载，生产代码未再变化，rc2不为仅增加这些测试重复打包。下一步继续E03/E04应用钩子及调度故障组合的证据核对；物理掉电/Windows/远程长期缺口保留。

rc2交付完成：`dist/releases/development-20260919-rc2` 构建、独立SDK构建签名、Master初始化/TLS/登录、双Daemon ONLINE和六个归档摘要全部通过；摘要文件SHA-256 `9cf78d121d1c28a58ba700b2e0ee5210c7a7670996ab397a7068115c8795dc62`。冒烟 `/tmp/blora-release-smoke-r1q90jz1` 所属进程已停止。继续A12真实ENOSPC：私有user/mount命名空间探测通过，新增 `transfers_enospc_linux_test.go`，仅显式 `BLORA_TEST_ENOSPC=1`运行，自启动隔离子进程并核对挂载namespace分离；1MiB tmpfs装在测试夹具目录，检查移动失败源保留/目标不发布/原数据保留/清理。该测试正在首次运行，尚无结果；新增测试未纳入rc2，不为仅补文档/测试再次打包。

当前交付检查点（2026-09-19）：协议背压修复后全仓Go普通回归退出0（Master129.494s、terminal16.944s），`make check`通过；协议race与真实101任务浏览器均通过。正在统一构建 `development-20260919-rc2`，包含扩展查询恢复、真实取消证据、节点测试入口修复和控制队列背压。下一步独立解包冒烟与归档摘要校验；不为装入包自身验证记录再重复打包。当前无活动fixture。

最新（2026-09-19）：控制通道有界背压修复后的协议race 1.192s通过；重编译Master/Daemon后，新fixture-1287262529的通知/任务分页2/2（8.2s）通过，加强为全部101任务SUCCEEDED后再次2/2（7.9s）。两个本轮fixture均已停止，退出0。完整 `go test -p 1 ./... -count=1` 会话45734正在运行；完成后记录结果并统一生成包含协议修复及扩展查询恢复增量的发行包。详见[控制背压报告](../acceptance/reports/control-backpressure-2026-09-19.md)。剩余全范围缺口不变，不把局部修复等同于全部完成。

当前检查点（2026-09-19）：私有夹具支持的40项真实浏览器回归38通过/2失败（6.3m）。快捷入口测试错误地要求历史启动任务为零，已改为检查本次新增任务，定向1/1（3.6s）通过。任务分页测试101次mkdir中6项为 `INTERRUPTED/node_acceptance_missing`，不是环境缺失；发现控制连接接收队列饱和即断链，正在修复为控制通道有界背压并添加真实TLS慢消费者回归。`internal/protocol/conn.go` 和 `protocol_test.go` 已改，race会话52739；正在重建双二进制用于新fixture复验。旧fixture-1890745423已Ctrl+C退出0。下一步确认协议race、启动新fixture，重跑任务历史与节点/控制场景，再更新证据；此增量尚未发行。

最新故障修复：2026-09-19 节点维护测试误用未显示在桌面上的应用图标，已改为从实际应用菜单打开；添加动作超时与主错误诊断。私有 fixture-1890745423 定向 1/1（4.7s）及完整 nodes.spec.ts 2/2（53.1s）通过，包括真实 Daemon 暂停/恢复和维护/票据下载。详见[全套回归报告](../acceptance/reports/full-browser-regression-2026-09-13.md)。当前继续同夹具支持的真实浏览器回归；夹具会话 83987，凭据路径 `.local/fixture-1890745423/browser-credentials.json`（禁止输出内容），完成后停止所属进程。

## 2026-09-19 续接：扩展任务请求身份

最新：真实取消丢回执/刷新组合已补验，Chromium 1/1（4.6s）通过；暂停已核对身份的私有 Daemon，真实 Master 接受取消后丢弃 HTTP 响应，刷新无重发，恢复节点后任务/界面确认 CANCELLED，取消键保持。fixture-1780818324 已恢复并停止。详细命令与范围见扩展任务重试报告；本项不证明运行中 WASI 中断，也不清除 Windows/远程长期缺口。下一步核对矩阵其余可执行场景，合并增量后统一发行验证。

参考扩展任务查询失败恢复入口已补齐：“刷新任务状态”只发起读取，查询期间去重，旧任务结果不覆盖新任务；定向 Chromium 1/1（8.5s）通过，断网后恢复查询不重发创建/取消。独立包构建通过，增量尚未进入 rc1。下一步继续真实取消丢回执组合，再统一纳入后续发行包；本轮测试已结束，无活动会话待接管。

发行增量已落地 `dist/releases/development-20260919-rc1`：Linux/Windows 构建、前端生产构建、SDK 与参考扩展三版本构建通过；独立提权冒烟通过 SDK 构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE；六个归档摘要全部 OK。冒烟目录 `/tmp/blora-release-smoke-_xl3psca` 所属进程已停止。本条为产物生成后的外部验证记录，不为把该条自身装入包再次打包。下一步继续矩阵尚缺的真实取消故障组合和平台验收；当前无测试会话待接管。

独立参考扩展浏览器取消恢复场景 1/1（8.7s）通过：路由中止首次取消，刷新不自动重发，显式重试复用原取消键，任务创建总数仍为 1。使用真实 Chromium 与模拟 HTTP 后端；真实节点接受取消后丢回执仍未单独演练。参考包重新生成。下一步按验收矩阵继续真实故障组合及统一发行验证，无运行测试会话待接管。

取消服务端补验完成：真实 TLS 双节点 `TestExtensionAdminAPIInstallsAndServesPackage` 7.018s 通过，同键取消保留原回执，CANCELLED 后重试修订不变，原扩展授权与撤权断言仍通过。浏览器取消回执丢失/刷新组合仍待验证；不能将服务端集成当作该 UI 场景完成。

取消链路也已接入显式稳定键：SDK `cancelTask(taskId, requestId)` 传递调用者键，宿主拒绝缺失/空白/超长键，参考扩展先持久化按任务保存的取消键再发送。SDK→宿主传输回归覆盖响应丢失后重建宿主并同键重试，定向 12/12 通过；前端类型检查、47/47 单测及 SDK/参考扩展构建通过。SDK 文档同步新签名；取消的真实浏览器故障场景尚未验证，下一步补此场景及最终发行包。

补验完成：SDK 到宿主的丢响应回归通过，前端 47/47；真实 HTTPS 双 Daemon Chromium 场景 1/1（9.2s）通过，后端接受后中止首次响应，刷新不重发，显式同键重试返回原任务 ID，WASI 结果正确。首轮仅末尾旧请求计数断言失败，修正为两次后通过。fixture-950580611 已停止（退出 0）。详细证据见[扩展任务重试](../acceptance/reports/extensions-task-retry-2026-09-19.md)。下一步核对其他扩展任务调用入口和取消请求身份，完成该能力块后统一生成发行包。

核对当前源码发现 SDK `createTask(payload, requestId)` 实现丢弃第二参数，宿主随后自动生成新请求键。已改为 SDK 传递 `{payload, requestId}`，宿主校验请求键并放入 HTTP 头，任务正文保持原值；参考扩展在发送前持久化待提交内容和键，确认收到任务后清理。前端类型检查与 46/46 单测、SDK 和参考扩展 TypeScript 构建通过。定向传输测试覆盖正文与请求键分离、缺失键拒绝；尚需补充 SDK 到宿主的丢响应重试及真实扩展浏览器验证。此次源码尚未进入 final54 发行包；后续先完成这些验证，再统一打包，避免只为补写包自身报告重复生成归档。全范围验收仍未完成，原矩阵缺口继续有效。

记录日期：2026-09-14
当前阶段：B00～B09 的主要代码链路已接入；B10 正在进行全范围回归、故障边界和交付证据整理。  
当前目标：连续完成全部明确范围；验收矩阵仍按缺失环境和未覆盖组合保留“进行中”，不把局部通过写成整项完成。

云工作区布局/引用 PUT 的请求键已写入窗口恢复状态，网络响应丢失后刷新并重试仍复用同一键，成功后清理；正文同步保持显式选择。`cloud workspace sync` 提权 Chromium 场景 1/1 通过（8.0s），路由中止首次写入后第二次同键只产生一次成功写入。节点登记、换钥、撤销动作也保存待处理键；上传分片按偏移派生稳定键，Master 对分片/完成入口在副作用前拒绝缺失键，上传重连/取消定向集成提权通过（2.180s）。`npm run check`、前端 46/46 单测通过。final53 已完成打包、提权独立冒烟和六个归档校验；冒烟目录 `/tmp/blora-release-smoke-zsjybjny` 已停止所属进程，`SHA256SUMS` SHA-256 为 `6a52012c8afb1335d97cd9792f73e7ff8793574547f38825b91084db824c7484`。Windows、远程节点/Engine、物理故障与长期组合仍按矩阵保留进行中。

上传浏览器恢复场景追加分片/完成键断言，提权 Chromium 1/1（5.6s）通过；三次分片请求按偏移复用同一上传键，完成请求使用固定后缀，响应中止后刷新仍只创建一个任务。final54 已完成打包、提权独立冒烟和六个归档校验；冒烟目录 `/tmp/blora-release-smoke-sgmac81e` 已停止所属进程，`SHA256SUMS` SHA-256 为 `265c1a8b4863d73212e8c94001bce95f4d83b4ec738d466a25bf5a47b33867fc`。其余平台和长期组合缺口不变。

final54 之后的提权串行全仓 Go 回归退出 0（Master 125.232s、extensions 16.389s、terminal 16.885s，其余包通过），节点/上传请求键改动未引入其他包失败。资源复核无 Blora/Vite/Playwright 进程和 `blora.test=1` Docker 对象；`/tmp` 可用约18GiB。最新请求键证据已写入工作区，发行包中的文档按 final54 记录。

调度更新/删除请求回执收口：`backup.Scheduler.SaveMutation` 与 `DeleteMutation` 将配置版本 CAS、记录删除和 `UserMetadataMutation` 回执放入同一 SQLite 事务；同键更新/删除重试稳定回放，不同载荷返回 `IDEMPOTENCY_CONFLICT`。前端定时任务删除将稳定请求键保存在窗口恢复状态中。定向调度 race 通过；随后提权全仓普通回归 Master 128.030s、竞态回归 Master 267.872s，扩展 149.546s，存储 13.581s，终端 18.646s，均退出 0；`make check`、前端构建及 final46/final47 发行包复验均完成。

final46 交付完成：调度请求回执修正与全仓回归证据同步后生成 `BLORA_VERSION=development-20260914-final46`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-94mpn3me` 已停止进程，`SHA256SUMS` SHA-256 为 `18cc8b4e1013c968924f4159eecfeb1925fff78b14e57906461b928989c3b537`。

final47 交付完成：纳入调度重启级回执测试与 OpenAPI 删除契约后的 `development-20260914-final47` 打包、独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-jzdy74x5` 已停止进程，`SHA256SUMS` SHA-256 为 `24e852ccf0fe2a6cba417ddb45217ca25cae98a75c087e2a1b7107a814af4ca2`。

默认系统/监控应用请求键现场收口：服务/计划任务动作、防火墙应用与确认按目标保存稳定键，进程终止确认框保存键，成功后清理，失败或丢响应时重试复用。提权 Chromium 系统管理浏览器场景 3/3（14.6s）通过；前端单测 46/46、类型检查和生产构建通过，修正已纳入 final48 发行包。

final48 交付完成：默认系统/监控应用请求键现场修正、调度回执与重启测试证据同步后生成 `BLORA_VERSION=development-20260914-final48`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-b1wqtbhw` 已停止进程，`SHA256SUMS` SHA-256 为 `580112f9bd3cc8738162a91fe800fb151bf8ac9c3b5f027e4c890ac63b48aee4`。

默认备份清理、用户创建、改密和扩展管理界面请求键现场收口：待处理键分别保存在备份窗口恢复状态、用户创建草稿和密码/扩展组件生命周期中，成功后清理，失败或丢响应时重试复用；密码和扩展包正文仍不写入工作区恢复。提权 Chromium 备份、账户/密码和扩展管理浏览器场景 4/4（9.4s）通过，前端 `npm run check`、46/46 单测和生产构建通过。下一步将该增量和证据纳入新的发行包并复验，Windows、远程节点/Engine、物理故障与长期组合仍按矩阵保留进行中。

final49 交付完成：上述界面请求键修正后的 `development-20260914-final49` 打包、提权独立冒烟和六个归档校验均退出 0；冒烟目录 `/tmp/blora-release-smoke-yb1y0em8` 已停止所属进程，`SHA256SUMS` SHA-256 为 `1a2ad9de3ce7951df30aa0aece4e51703a7cb20f9f7155d14dc99d1c3bc87aca`。受限冒烟首轮的 esbuild `EPERM` 已按环境限制单独记录；当前无活动 Blora、Vite/Playwright 进程或测试 Docker 对象。

定时任务删除成功后清理旧请求键的修正纳入 final50：前端检查、46/46 单测、构建、提权独立发行包冒烟和六个归档校验均通过；冒烟目录 `/tmp/blora-release-smoke-rptovpg1` 已停止所属进程，`SHA256SUMS` SHA-256 为 `8cd89bb3825e86bcac3f348ef7e443ab3e3d6bb767a43dc3ae60064b73c22f3d`。当前无活动 Blora、Vite/Playwright 进程或测试 Docker 对象，验收矩阵仍如实保留 Windows、远程、物理故障和长期组合缺口。

扩展管理请求键现已写入窗口恢复状态，刷新后仍可重试启停、回滚、卸载和目录安装；扩展管理 Chromium 场景 1/1、前端检查、46/46 单测和构建通过。final51 独立冒烟及六个归档校验通过，冒烟目录 `/tmp/blora-release-smoke-yratxpey` 已停止所属进程，`SHA256SUMS` SHA-256 为 `f87459b1eb291633fd6d31c156dfe7e72611dfc5312287ac39852cc50baaa1c1`。Windows、远程、物理故障和长期组合仍按矩阵保留进行中。

任务取消收据幂等收口：`model.Task` 持久化首次取消请求键，普通任务、上传和传输取消均在 SQLite 事务中记录；并发取消只推进一次修订，任务进入 `CANCELLED` 后重试仍返回原收据。任务中心为每个任务保存稳定取消键。存储 `Test(Cancellation|ConcurrentCancellation)` 与真实 TLS Master 集成 `TestTaskCancellationReplayAfterTerminalKeepsReceipt` 通过（0.139s），OpenAPI Task schema 已声明 `cancellationRequestId`。改动后提权全仓普通回归 124.958s、竞态复跑 255.904s 均退出 0（首次竞态仅 runlog `/proc` 时序波动，隔离复跑 2.345s 通过），`make check` 与前端 46/46 通过；修正已纳入 final36 发行包。

节点登记票据幂等收口：`EnrollmentMutation` 将管理员、请求键和名称写入受保护的元数据回执，同键重试复用原票据，不同名称返回 `ErrRequestMismatch`/`IDEMPOTENCY_CONFLICT`；原有无请求键的内部测试入口保持随机票据语义。存储 `TestEnrollmentMutationReplaysTicketByRequest`、真实 TLS 双 Daemon 管理回归和 `make check` 通过。登记修正后的提权全仓普通回归 126.994s、竞态回归 261.687s 均退出 0，前端 46/46 通过。

final38 交付完成：将登记幂等修正、全仓普通/竞态证据和矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final38`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-rrefkg26` 已停止进程，`SHA256SUMS` SHA-256 为 `ce76c1e52dfeb38627f2181c8751cb15048c538f0a025f05c568d7c2ad240628`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

节点换钥票据幂等收口：`NewRotationTicketMutation` 将管理员、节点、请求键和 `reactivate` 写入受保护元数据回执，同键重试复用原票据，不同载荷返回 `IDEMPOTENCY_CONFLICT` 且不覆盖待确认记录；真实 TLS 换钥集成覆盖回放、冲突、身份保持和旧连接栅栏。改动后提权全仓普通回归 128.965s、竞态回归 262.770s 均退出 0；下一步生成包含换钥修正和证据的发行包。

final39 交付完成：将换钥票据幂等修正、全仓普通/竞态证据和矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final39`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-wsor4qlc` 已停止进程，`SHA256SUMS` SHA-256 为 `e63dd2e0ae2c9fbd25e7fba2abf245fb8027984f23d5c07f13939eedc8647728`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final40 交付完成：将 final39 的换钥证据、进度与矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final40`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-ghsj5ufc` 已停止进程，`SHA256SUMS` SHA-256 为 `1739590e081a8fc043f914e82f5adda7c998e759df1e9fb3fb50a01c787d9fce`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

请求键边界再收口：实例创建、实例控制、终端创建和授权写入均在读取请求体前拒绝缺失/非法键；定向回归 6.241s、最新 Master 完整普通回归 127.825s、竞态回归 265.970s 均退出 0。下一步生成包含该边界证据的发行包。

final41 交付完成：将请求键边界证据、进度与矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final41`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-rwgm48yh` 已停止进程，`SHA256SUMS` SHA-256 为 `37996802de1cc4510380974f8d60cbab69ca9513e9371c952be80dd631c7c286`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

Compose 项目保存 prepare/chunk 入口补齐请求键，前端按准备键和分片偏移发送稳定键；定向容器管理回归 0.377s（Compose 真实镜像用例因缺少显式镜像而按规则跳过），前端 46/46 通过。最新 Master 普通回归 129.598s、竞态回归 264.964s 均退出 0；下一步生成包含 Compose 修正和证据的发行包。

final42 交付完成：将 Compose 请求键修正、前端分片键和最新 Master 回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final42`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-k7gfk8j5` 已停止进程，`SHA256SUMS` SHA-256 为 `b6a8cc499522d141cbe05e11fdd7b1878b9ddc5e17d6f465ff4be9ca7eddf73e`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final43 交付完成：将 Compose OpenAPI 请求键声明同步进包后生成 `BLORA_VERSION=development-20260914-final43`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-zg38nbet` 已停止进程，`SHA256SUMS` SHA-256 为 `78520382738ab277e285a3cd084fd22b9ea8214d2f1aa802a9e828652fcee24d`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

换钥票据同键载荷冲突错误码明确为 `IDEMPOTENCY_CONFLICT`（409），真实 TLS 集成回归 0.209s 通过；下一步生成包含该错误码修正的发行包。

final44 交付完成：将换钥冲突错误码修正和回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final44`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-nn8w4hzn` 已停止进程，`SHA256SUMS` SHA-256 为 `15f56cc3151cc38b24ab0d2a916bff6342a9be27055f008bb259231a68deed70`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

管理写入口的请求键冲突错误码统一为 `IDEMPOTENCY_CONFLICT`；定向用户/角色/节点回归 1.198s，通过。最新 Master 完整普通回归 127.672s、竞态回归 263.528s 均退出 0；首次并行回归因 `/tmp` 空间耗尽，清理本轮缓存后重跑通过。下一步生成包含该统一错误码修正的发行包。

final45 交付完成：将管理写入口统一幂等冲突错误码及最新回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final45`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-8pjxh2wo` 已停止进程，`SHA256SUMS` SHA-256 为 `e332a19b9c80ea6fdecdfc1638eaca3583dd547ac3e8aa787203439dec120195`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

2026-09-14 持久化请求键扩展：用户创建/改密、节点登记/身份轮换、云端工作区 PUT/DELETE 及通用任务/上传取消在授权后统一检查 `Idempotency-Key`；用户与工作区写入进入 SQLite 事务幂等回执，工作区 CAS、用户创建重放、改密和传输取消边界已加入集成测试。`go test -race` 定向 Master/Storage、传输取消回归、`make check` 和 OpenAPI 3.1 解析均通过。详见[请求键报告](../acceptance/reports/request-keys-2026-09-14.md)。

防火墙二阶段确认幂等收口：确认请求键随 RPC 传到 Daemon，并写入租约记录和成功任务结果；同键在租约仍存或任务已完成时复用确认结果，其他键返回冲突且不重复修改主机规则。Daemon 定向 race 29.354s、Master 定向 race 2.096s 通过；改动后串行全仓普通测试（Master 126.258s）与 race（Master 256.038s、extensions 136.535s、terminal 18.627s）均退出 0，`make check`、前端46/46亦通过。随后补齐重启后无内存租约的双键 CAS 并发确认：只允许一个键成功，获胜键可回放，另一键返回 `FIREWALL_LEASE_CONFLICT`；定向 race 1.137s，最新串行全仓普通测试 Master 126.926s、race Master 256.432s 均退出 0。下一步生成包含该修正和证据的 final35 发行包并复验。

final34 交付完成：防火墙确认请求键修正后的 `BLORA_VERSION=development-20260914-final34 make package`、独立 `package-smoke.py` 和六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-ukbhphef` 已停止进程，`SHA256SUMS` SHA-256 为 `83dcf17fc84253c0e8731056497a7b0f83b4e2eb5f1768ede6d033260dfccd15`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final35 交付完成：重启后无内存租约的确认 CAS 并发修正纳入 `BLORA_VERSION=development-20260914-final35 make package`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-okvsdiwf` 已停止进程，`SHA256SUMS` SHA-256 为 `de8c519f6af02db48e41129150d41cef7805bd4bdeaa14b08db99ec3bcb1dd4d`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final36 交付完成：任务取消收据幂等修正（首次 `cancellationRequestId` 持久化、并发取消单次提交、终态重试复用）纳入 `BLORA_VERSION=development-20260914-final36 make package`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-7o0b93de` 已停止进程，`SHA256SUMS` SHA-256 为 `50fb039b87dc00bef14d1006f0e6ebcf929f407a1790efe51426d0d77ff80913`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final37 交付完成：将 final36 进度与矩阵记录同步进包后生成 `BLORA_VERSION=development-20260914-final37`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-jna5wn75` 已停止进程，`SHA256SUMS` SHA-256 为 `4fbae56157ee6ed406d6f8deb91e584be61fe44b130f05d16669a42e691a59a3`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

当前源码串行全仓复验已明确退出 0：`go test -p 1 ./... -count=1`，Master 126.349s，runlog、runtime、storage、terminal、extensions 及其余包全部通过；该结果与此前并行 runlog 时序失败记录并存，仍不覆盖 Windows、远程和物理故障环境。

随后以独立缓存执行 `go test -race -p 1 ./... -count=1`，明确退出 0；Master、extensions、terminal、runlog、runtime、storage 及其余 Go 包全部通过。该竞态证据覆盖当前源码，不覆盖 Windows 真机、远程高延迟和物理故障场景。

随后收口文件入口：文件保存、文件动作和新建上传在实例/写权限确认后立即拒绝空白或超过 128 字节的请求键，避免读取正文、生成上传身份或联系节点；新增空白保存键与 129 字节动作键回归 2.693s、竞态回归 5.406s 均通过，完整 Master 126.836s、`make check` 均退出 0。上传分片/完成仍按已确认检查点与任务状态推进，前端 `api()` 会为其自动附带请求键。
随后发现并修复两个上传取消历史用例未携带请求键的问题；定向上传取消回归 1.562s、完整 Master 套件 125.538s 均退出 0。全仓其余包此前已通过，最终交付前继续保持串行全仓证据与发行包校验。

同日全仓与交付复验：提权环境 `go test ./...` 首轮仅因 Master 备份用例出现一次 `terminal access denied` 而失败；该备份用例单独复验通过，随后完整 `go test ./internal/master -count=1` 通过（125.544s），其余 Go 包也通过。受限环境的本机监听/辅助进程失败不计产品失败。前端 `npm test -- --run` 10 文件 46/46、`make check` 均通过。`BLORA_VERSION=development-20260914-final22 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS` 全部通过；`SHA256SUMS` SHA-256 为 `d9f80088796e1655df5fd9c3e4a46b92904a5675b2b1a895aa51d021fa60ca5c`，临时冒烟目录 `/tmp/blora-release-smoke-6svj66pc` 已停止进程但按约定保留私有诊断目录。

同日 final23 文档同步交付：将请求键 race 最终耗时同步后执行 `BLORA_VERSION=development-20260914-final23 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-he8gr2sa` 已停止进程，`SHA256SUMS` SHA-256 为 `7bab5ab5c69dda7cc0fef737ee7885af9b7caaebec3c9fc883ee5164d2eed2cd`。

同日 final24 记录同步交付：修正全仓测试一次环境波动说明后执行 `BLORA_VERSION=development-20260914-final24 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-uf0cgu2x` 已停止进程，`SHA256SUMS` SHA-256 为 `ed434701601b269098053b261a69ce0e785aea7d6021f971367839424996cefc`。

同日 final25 取消请求键交付：通用任务与上传取消新增授权后请求键检查，传输取消回归通过；执行 `BLORA_VERSION=development-20260914-final25 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-luefmwwc` 已停止进程，`SHA256SUMS` SHA-256 为 `466f6d3e6a2da7c3ab37beb389ca45e3c4dae5f87785cbf0788b738a8311a5fa`。

同日 final26 收口：修正两个上传取消历史测试未携带请求键的问题；上传取消定向回归 1.562s、完整 Master 套件 125.538s 退出 0。`BLORA_VERSION=development-20260914-final26 make package`、独立 `package-smoke.py` 与六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-ola9vpox` 已停止进程，`SHA256SUMS` SHA-256 为 `272c3f3bef10d4a380394ba23d2f99692b2328af4f63a12378ae79958f8a6e55`。

同日 final27 文档归档：将 final26 记录纳入包内后生成 `development-20260914-final27`，打包、独立冒烟与六个归档校验退出 0；冒烟临时目录 `/tmp/blora-release-smoke-bqwy03_2` 已停止进程，`SHA256SUMS` SHA-256 为 `a89541bb7657d7174d454828a7cea77477c9aaa73649b37ff85d55fe10f5faa`。

同日改密幂等收口：将自助当前密码校验移入首次事务回调，撤销会话后同键重试可稳定复用回执；定向改密 1.044s、完整 Master 127.375s 退出 0。下一步生成包含本修正与证据的 final28 并复验。

同日 final28 交付：改密幂等修正后的 `development-20260914-final28` 打包、独立冒烟与六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-_y6cvl_e` 已停止进程，`SHA256SUMS` SHA-256 为 `efc9e4dda49e234b47d2037990331f726af7b674094b62667aa941c4c0c4d030`。

收尾复验：存储包单测 0.579s、`make check`（go vet）通过；调试残留、Blora 测试进程及测试标签容器均无活动对象。

同日 final29 交付：将最终收尾证据纳入归档后生成 `development-20260914-final29`；打包、独立冒烟和六个归档校验退出 0，冒烟临时目录 `/tmp/blora-release-smoke-k43e8lci` 已停止进程，`SHA256SUMS` SHA-256 为 `12171c6ed4aa430da966bb63e1b6ca3b20767985df2871667f17b36fba63af2a`。

同日 final30 交付：文件保存、文件动作、新建上传入口请求键边界收口后生成 `development-20260914-final30`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-dl1fg2bz` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `f7630947960072aded20d64f7a64397c8737bf125599f28fd8f661d143a5802d`。

随后将文件场景竞态回归证据纳入归档，生成 `development-20260914-final31`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-m3ccimab` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `18ff79189be48c49c93bb8c0a5330827472726f8f242b54eb7a44c6c28db3e80`。

随后将串行全仓 Go 回归证据纳入归档，生成 `development-20260914-final32`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-4osq9cd9` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `06e336ee2ec13240f2d7b86f6c83b638d5ec42db8cf2d9b0de35289aaae25c96`。

随后将串行全仓 Go 竞态回归证据纳入归档，生成 `development-20260914-final33`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-9tnfnent` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `84931b88163ef6182ca4504b9d26105b427c4c16fe569f783c2807223abbe34e`。

当前收尾检查：工作区无 Blora/fixture/Vite/Playwright 活动进程，Docker `blora.test=1` 标签对象为空；final33 归档校验已复核。下一步仅推进验收矩阵中仍可获得环境的缺口（Windows 真机、远程高延迟/Engine、物理 ENOSPC/掉电、小时级调度与完整跨平台组合），未因本地串行测试和 Linux 冒烟通过而宣称 F01～F14/A01～A17/E01～E09 全部完成。

最新普通 `go test ./... -count=1`：Master 127.270s 及其余包通过，runlog 输入时序用例首轮空输入失败；隔离重跑 0.052s 通过。该非确定性结果已保留，不能宣称本轮全仓无条件通过；生产代码与发行包不受此测试时序波动影响。

2026-09-14 备份/调度竞态复验：提权隔离环境执行 `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（backup 1.856s、Master 5.816s），覆盖真实 TLS API 归档/条件恢复/撤权/取消、计划触发与 SQLite 回执恢复；`go test -race ./internal/daemon -run 'TestBackupHook' -count=1` 退出0（2.039s），固定 argv Hook 的环境绑定、稳定回执及未知副作用不重放通过。具体应用 hook 语义、长期离线调度、物理 ENOSPC/掉电仍按 E03/E04 保留。

同日真实备份浏览器验收：`backups.spec.ts` 在 HTTPS 双节点 fixture-4134305623 中 1/1 通过（3.5s）；真实文件写入/版本核对、备份归档任务、默认备份窗口恢复计划与覆盖任务、中文正文核对，以及 Asia/Shanghai 定时计划保存/修订删除均通过。fixture 已停止并清理。证据见[备份浏览器报告](../acceptance/reports/backup-browser-2026-09-14.md)。物理 ENOSPC/掉电、具体应用 hook 语义和长期离线调度仍未验证。

同日 final20 交付复验：因真实备份浏览器报告更新执行 `BLORA_VERSION=development-20260914-final20 make package`，`python3 scripts/package-smoke.py dist/releases/development-20260914-final20` 与 `cd dist/releases/development-20260914-final20 && sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-p00oohac`，`SHA256SUMS` SHA-256 为 `4e25104368473ca06daaae78870b9ee8b015e5e0ebbda665ad26be116a45c60b`。

同日 final21 交付收口：将最新备份/调度报告同步进包内后执行 `BLORA_VERSION=development-20260914-final21 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final21` 和六个归档 `sha256sum -c SHA256SUMS`，均退出0；冒烟临时目录 `/tmp/blora-release-smoke-klzy0dtj`，`SHA256SUMS` SHA-256 为 `9cb7f53ce64e6703003054bdcce2af95a7b5c392b59b0de9debde1e9b2e5282d`。

2026-09-14 请求键边界扩展：扩展包安装/升级/回滚/启停/卸载、目录安装和防火墙租约确认统一在授权后检查 `Idempotency-Key`；新增定向 race 77.219s 和 OpenAPI 3.1 解析断言通过，见[请求键报告](../acceptance/reports/request-keys-2026-09-14.md)。前端非 GET 请求继续由 `api()` 自动生成合法键。

同日 final19 交付复验：因上述证据更新执行 `BLORA_VERSION=development-20260914-final19 make package`，`python3 scripts/package-smoke.py dist/releases/development-20260914-final19` 和 `cd dist/releases/development-20260914-final19 && sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-urc6k8iz`，`SHA256SUMS` SHA-256 为 `7e365582aad6a03cc84e1d7b7342ce795da2f903036f0db059f37875385bc9b2`。包内含最新 docs/acceptance/reports/backup-scheduler-race-2026-09-14.md。

2026-09-14 512MiB 传输批量步进复验：重编译夹具纳入 `transferChunksPerStep=16` 后启动私有 fixture-1569035732，真实 Chromium 混合场景 1/1 通过（约2.0分钟）。八窗口、双PTY、10,000项目录与512MiB跨节点传输全程并行，目标校验成功；536,870,912B 用时110.868s（约4.62MiB/s），反馈p95 39.0ms、0/285长帧、JS heap53,729,723B、API RTT p95 43.7ms，未确认字节峰值86,306B。fixture已停止且无测试残留。首次旧夹具结果仅保留为索引基线；长期内存/远程网络/归档轮转、Windows及物理故障仍按矩阵保留。证据见[大文件传输报告](../acceptance/reports/transfer-proof-2026-09-13.md)和[真实性能报告](../acceptance/reports/performance-real-2026-09-13.md)。

同日长时性能采样：性能测试新增可选 `BLORA_PERF_SOAK_SECONDS`（每250ms堆采样并持续窗口交互），重编译夹具 fixture-2822501719 设置60秒后真实场景1/1通过（约2.0分钟）；采样65.697s、p95 43.3ms、1/3770长帧、堆34,569,872–97,894,433B、未确认峰值86,867B，512MiB传输109.245s。该模式默认关闭，测试超时预算已包含soak时长；小时级稳定性仍未验证。详情见[真实性能报告](../acceptance/reports/performance-real-2026-09-13.md)。

同日传输改动后的全仓竞态复验：`go test -race -p 1 ./...` 退出0，Master 248.686s、extensions 136.133s、terminal 18.695s，其余 Go 包通过；使用独立 `/tmp/blora-go-race-final14` 缓存，未出现空间或监听权限错误。

同日 final14 交付复验：`BLORA_VERSION=development-20260914-final14 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final14` 均退出0。独立解包完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE；`SHA256SUMS` SHA-256 为 `63a3318876d538a4e91ea52ea7318485e424931f23b5162285e4ede38da81099`，临时冒烟目录 `/tmp/blora-release-smoke-y8dorqs1`。

由于发行脚本包含 README/docs，证据更新后生成 final15：`BLORA_VERSION=development-20260914-final15 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final15` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-7wbln31e`，`SHA256SUMS` SHA-256 为 `746ed95e6d047933374c35b1d3b119902d96a9a8ac90904acbe1abb3b9f1c183`。

加入长时采样开关及60秒证据后生成 final16：`BLORA_VERSION=development-20260914-final16 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final16` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-be_8yxox`，`SHA256SUMS` SHA-256 为 `d2acfcb7b76896028bf701bc1e96778e2de9c29fc0909b4212b707af384d96c7`。

长时采样测试超时预算修正后生成 final17：`BLORA_VERSION=development-20260914-final17 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final17` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-y3ghbju2`，`SHA256SUMS` SHA-256 为 `de8908dc5ea53aca45f154d5413aab5950eddeb04c9bbf7f394928f1dc4df081`。

同日 Windows 适配编译复验：在 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` 下，`go test -c` 分别生成 runtime、runlog、systeminfo 测试入口（`/tmp/blora-runtime-windows.test.exe`、`/tmp/blora-runlog-windows.test.exe`、`/tmp/blora-systeminfo-windows.test.exe`），全部退出0；无 Wine/Windows 真机，仍不计运行验收。

同日 final18 交付复验：`BLORA_VERSION=development-20260914-final18 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final18` 与六个归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-rb_wc5my`，`SHA256SUMS` SHA-256 为 `adb76c7bac48b4c47728c43bd25b91fbb1d81eec930de17eac0aa583a0a30928`。

同日交付检查：`make check`（`go vet ./...`）与 `cd web && npm test -- --run`（10个文件、46/46）退出0；final14 构建已同时完成前端类型检查和生产构建。

同日传输定向 race：`go test -race ./internal/master -run '^TestTransfer' -count=1` 退出0（64.894s），覆盖当前批量步进实现的跨节点传输集成回归。

2026-09-14 E02 真实 Docker 复验：Docker Engine 29.4.1/API 1.45/Compose 5.1.3 上，`TestRealDockerComposeLifecycle` 使用随机标签测试对象完整执行 25.063s 退出 0，覆盖容器/镜像/卷/网络、双流日志、真实拉取及失败、Compose 健康失败/更新/删除和卷保留；清理后无测试对象。远程 Engine、真实 ENOSPC/掉电、Windows 容器与长时网络故障仍未验证。

同日 E02 浏览器补验：任务自有双节点 HTTPS fixture 连接本机 Docker Engine，`tests/real/docker.spec.ts` 1/1 通过（44.5s）；覆盖容器中心草稿刷新、Compose 保存与显式部署、中文输出/阶段分页、任务详情刷新、删除部署和卷保留。fixture 与随机 Docker 对象均清理，远程 Engine、ENOSPC/掉电、Windows 容器和长时故障仍未验证。

同日 A09 内存采样补验：真实 Linux PTY 无消费者滚动场景延长至约17.5秒（32批×2MiB、0.5秒间隔），每5ms采样归档和 `/proc/<pid>/status` RSS；定向 race 退出0，3435次采样，归档峰值65,484B、磁盘54,729B，PTY RSS 741,376B→3,764,224B，未超过基线+32MiB，最终Gap与末尾标识通过。证据见[混合流报告](../acceptance/reports/mixed-stream-control-2026-09-13.md)；跨平台、全连接长期及更长生产负载仍未验证。

同日 A13/E06 未授权浏览器边界：真实 HTTPS fixture 的 `extensions.spec.ts` 成员独立浏览器上下文 1/1 通过（3.5s）。无 `app.use`、仅有 `app.use` 无 `node.read` 时资源摘要/任务均 403；补齐 `node.read` 后资源摘要 200，撤销 `app.use` 后立即回到403；finally 撤销授权、卸载扩展并清理 fixture。证据见[扩展未授权浏览器报告](../acceptance/reports/extensions-unauthorized-browser-2026-09-14.md)。Windows、远程高延迟和长期组合仍未验证。

同日扩展真实套件复验：新增场景并入 `extensions.spec.ts` 后完整真实 HTTPS 套件 6/6 通过（22.6s），fixture Ctrl+C 停止且无残留进程/测试 Docker 对象。

2026-09-14 请求键边界收口：`requireRequestID` 现在统一拒绝空白及超过128字节的 `Idempotency-Key`；实例控制在动作解析、资源授权和节点门禁后、任务接纳前检查。后端定向集成回归、前端类型检查与 46/46 单测通过。`BLORA_VERSION=development-20260914-final10 make package` 和 `python3 scripts/package-smoke.py dist/releases/development-20260914-final10` 均退出 0，完成独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE。

同日最终复验：清理任务缓存后在提权环境以 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-race-final13 GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 退出 0，Master 251.165s、extensions 136.422s、terminal 18.729s，其余 Go 包通过。此前空间耗尽的复跑失败（`no space left on device`）保留为环境记录，未计作代码失败。

幂等键边界回归新增 `TestRequireRequestIDBounds`（空白/缺失/129 字节拒绝，128 字节接受）并通过；随后以 `BLORA_VERSION=development-20260914-final11 make package` 重新生成发行包，独立冒烟 `python3 scripts/package-smoke.py dist/releases/development-20260914-final11` 退出 0。

恢复实例控制的资源授权先后顺序后，以 `BLORA_VERSION=development-20260914-final12 make package` 重新生成发行包；`python3 scripts/package-smoke.py dist/releases/development-20260914-final12` 退出 0，独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 均通过。

A09 长时采样测试改动后，以 `BLORA_VERSION=development-20260914-final13 make package` 重新生成发行包；`python3 scripts/package-smoke.py dist/releases/development-20260914-final13` 退出 0（临时目录 `/tmp/blora-release-smoke-60hxav7x`），独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 均通过。

最终检查：`go vet ./...` 退出0；`cd web && npm run check && npm test -- --run` 退出0（10个文件、46/46）；项目测试/fixture及带 `blora.test=1` 标签的 Docker 对象均无活动残留。发行包 `dist/releases/development-20260914-final13/SHA256SUMS` SHA-256 为 `09865538d7e776eba64424d8f73766a0734c1235ae0a46015de2d2f764ed8cfc`。

## 2026-09-14 扩展跨设备现场与资源身份复验

- AppHost 恢复校准现在可以返回规范化 `resourceRef`：默认实例应用按稳定实例 ID 读取服务端节点身份，具备 `resource.read` 的扩展通过受控资源摘要接口校准实例引用；节点或实例被删除/撤权时保留原窗口和草稿，不把可恢复的离线现场标成恢复存储失败。涉及 [应用类型](../../web/src/app-host/types.ts)、[扩展注册表](../../web/src/app-host/registry.ts)、[桌面恢复](../../web/src/desktop/store.ts)、[默认应用注册](../../web/src/apps/register.ts)。
- 新增跨设备浏览器副本场景：两个不同 `deviceId` 打开同一云端布局/引用副本，扩展视图从状态版本 1 迁移到 2，旧节点提示校准到 `node-new`，恢复后保持两个扩展窗口，1/1 通过（4.7s）。证据：[扩展跨设备报告](../acceptance/reports/extensions-cross-device-2026-09-14.md)。
- 本轮 `npm run check`、`npm test`（46/46）、`npm run build` 和定向 Playwright 场景均退出 0。Windows 真机、远程高延迟、物理 ENOSPC/掉电及完整未授权组合仍按矩阵保留进行中。

## 2026-09-14 最新回归结果

- 统一请求键边界扩展到已持久化的实例/节点设置、用户/角色变更、节点撤销、扩展数据与资源写入、备份计划/恢复和传输取消；相关 Master 集成测试全部通过。OpenAPI 的对应修改操作现在声明 `IdempotencyKey`。
- 提权串行全仓 race 在上述入口改动后再次通过：`go test -race -p 1 ./...` 退出 0，Master 250.611s，其余包通过；受限沙箱重跑仅因本地 TLS 监听权限失败，未计作代码失败。
- A09 慢日志/批量传输/停止并行测试在当前源码下重新通过：`go test -race ./internal/master -run TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop -count=1 -v` 退出 0（13.07s）；真实 TLS Master/双 Daemon、8.4MB 跨节点传输和未确认日志接收端同时运行，未消费峰值 82,635B、接收队列 79,507B，停止和归档/目标校验均成功。全进程长期内存采样仍未覆盖，A09 不提升为整项完成。
- 扩展浏览器套件 6/6 通过（21.1s，含新增跨设备副本场景）；调度 API 删除缺少请求键时返回 400，带键删除及 race 定向测试通过。一次完整 `make test` 在并行资源压力下有 4 个既有时序失败（Daemon 防火墙命令、Master 审计登录、runtime 后代条件、PTY 输出）；随后逐项低并发重跑均通过，其中 Master TLS 场景在提权环境通过。该次全仓失败原样保留，不改写为全仓通过。
- 随后以 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-cache GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 串行完成全仓 race，退出 0：Master 259.481s，其余包通过。并行命令的失败事实仍保留，串行结果作为最新可复核全仓证据。

## 2026-09-13 扩展全链路真实复验

- E06/E07/A13 独立扩展成功链路完成新 fixture 复验：Linux Chromium HTTPS、真实 Master + 双 Daemon，`extensions.spec.ts` 5/5、24.2s，覆盖通知去重、元数据草稿/真实实例写入、用户数据迁移与失败回滚、WASI 任务与资源窗口/快捷入口/移窗/关闭、本地注册源双版本安装升级及禁用门禁。证据：[扩展全链路报告](../acceptance/reports/extensions-real-2026-09-13.md)。fixture `fixture-2953159556` 已停止。
- 仍按矩阵保留 Windows 真机、远程高延迟注册源、物理掉电/ENOSPC 及完整未授权组合，不能把本次成功链路当作 E06/E07 整项完成。

## 2026-09-13 全套浏览器回归

- 最终 Linux 构建在基础双节点 fixture `fixture-2323261112` 执行 43 项真实 Chromium 场景，37 项通过。6 项按环境条件保留失败：容器指标/Docker 需要可用 Docker Engine，性能与上传关闭需要 `--performance` fixture，节点维护长套件一次连接超时造成污染；实例设置单独新 fixture 重跑 1/1（29.8s）通过。详见[全套浏览器回归](../acceptance/reports/full-browser-regression-2026-09-13.md)。fixture 已停止。

## 2026-09-13 性能与关闭网页复验

- 带 `--performance` 的最终构建 fixture `fixture-2936034595`：性能 1/1（27.8s），p95 36.2ms、最大44.4ms、0/298长帧、8窗口/双PTY/10,000目录/16MiB传输全程并行；关闭网页上传续传 1/1（1.4m），10,000项解压、16MiB uploadId 从917,504B续传并完成摘要核验。证据：[真实性能](../acceptance/reports/performance-real-2026-09-13.md)、[关闭网页](../acceptance/reports/upload-close-2026-09-13.md)。fixture 已停止；长期稳定性、512MiB吞吐、远程网络和Windows仍未验证。

## 2026-09-13 最终 race 与扩展边界修复

- `make test`（`go test -race ./...`）退出 0：Master 267.898s，extensions 157.362s，其余包通过。新增软删除与任务接纳竞态保护、扩展迁移资源身份不变式均已通过定向验证；最终 Master/Daemon/devfixture 构建、Windows 交叉构建、前端 check/test/build 与 `make sdk` 均通过。

控制台输入请求键边界加入后再次运行 `make test`（`go test -race ./...`）退出 0：Master 257.810s，其余包通过。

## 2026-09-13 最终发行包冒烟

- `BLORA_VERSION=development-20260913-final make package` 退出 0，生成 Linux/Windows Master、Daemon、SDK 和 web 六个归档；随后 `python3 scripts/package-smoke.py dist/releases/development-20260913-final` 在受限提升环境退出 0，完成解包后的 SDK/参考扩展离线安装构建、临时签名、Master 初始化/TLS/登录及两个 Daemon ONLINE 检查。测试进程已停止，证据已补入[发行包报告](../acceptance/reports/package-2026-09-13.md)。Windows 真机、真实业务运行和物理故障仍按矩阵保留未验证。

随后在任务入口幂等键、慢消费者诊断和云端冲突提示改动后，`BLORA_VERSION=development-20260913-final2 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final2` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，详见[发行包报告](../acceptance/reports/package-2026-09-13.md)。

监控 stale 诊断和前端提示更新后，`BLORA_VERSION=development-20260913-final3 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final3` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../acceptance/reports/package-2026-09-13.md)。

系统管理按钮加入能力禁用状态后，`BLORA_VERSION=development-20260913-final4 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final4` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../acceptance/reports/package-2026-09-13.md)。

资源身份校准、撤权现场保留和调度删除请求键契约更新后，`BLORA_VERSION=development-20260914-final6 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final6` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../acceptance/reports/package-2026-09-13.md)。

调度创建/更新入口统一使用非空白 `Idempotency-Key` 校验；`BLORA_VERSION=development-20260914-final7 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final7` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过。

调度请求键校验提前到请求体解析前，避免无键请求触发解析或资源访问；定向 race 通过。`BLORA_VERSION=development-20260914-final8 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final8` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过。

统一持久化修改请求键后的 `BLORA_VERSION=development-20260914-final9 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final9` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过；发行包包含实例/节点/用户/角色/扩展/备份/传输入口的最终校验。

控制台输入入口补齐请求键校验和 OpenAPI 声明后，`BLORA_VERSION=development-20260913-final5 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final5` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../acceptance/reports/package-2026-09-13.md)。

## 2026-09-13 实例删除用户入口

- 实例中心和实例专用窗口新增管理员“删除实例”入口，使用现有确认对话框、CSRF/幂等请求和刷新反馈；服务端仍在持久化接纳阶段检查 STOPPED、活动任务和墓碑身份。前端 `npm test` 46/46、`npm run check`、生产构建均通过；当前源码下原有 A16 目标改名/软删除/同名替换回归在 fixture-1914707629 1/1（7.3s）通过。普通成员不呈现特权入口，删除失败保留原视图与错误。
- 实例中心和实例专用窗口新增管理员“删除实例”入口，使用现有确认对话框、CSRF/幂等请求和刷新反馈；服务端仍在持久化接纳阶段检查 STOPPED、活动任务和墓碑身份。前端 `npm test` 46/46、`npm run check`、生产构建均通过；当前源码下 A16 目标改名/软删除/同名替换回归改为真实点击管理员删除入口，在 fixture-115753733 中 1/1（7.0s）通过。普通成员不呈现特权入口，删除失败保留原视图与错误。

## 2026-09-13 任务入口幂等键边界

- 为主机终端、容器操作、Compose 应用、系统服务/计划任务/防火墙应用、进程终止、实例终端创建/关闭、扩展任务和跨节点传输补充统一的 `REQUEST_ID_REQUIRED` 校验。此前这些入口虽会在存储层拒绝空键，但错误不稳定且可能先触发远端检查；现在在授权后、任务接纳前明确失败，前端 `api()` 自动生成的请求键保持兼容。
- `gofmt` 后以提权本地 TLS 监听运行 `go test ./internal/master -run 'Test(System|Container|Host|Transfer|Extension|Process|Terminal)' -count=1`，退出 0（39.554s）。沙箱内首次运行因禁止监听端口失败，未计作代码失败。

控制台输入入口随后补齐同一 `REQUEST_ID_REQUIRED` 早期校验，OpenAPI 为实例终端、主机终端、Docker/系统任务、Compose 提交、跨节点传输和控制台输入补充必填请求键声明；`TestConsoleInputHasOneDeliveryAndGracefulStopUsesIndependentStdin` 定向运行退出 0（1.620s），覆盖空键 400、合法请求单次交付和独立停止输入。

OpenAPI 3.1 文档重新解析通过（103 paths），并逐项核对上述任务入口的 `IdempotencyKey` 引用。

## 2026-09-13 系统能力不可用反馈

- SystemApp 根据节点 `system/capabilities` 结果停止对不可用服务/计划任务轮询；对应操作按钮随能力禁用，能力缺失时仍给出明确原因，防火墙预览/应用同样不会提交注定失败的请求。可用能力路径保持原确认、幂等和二阶段租约逻辑。
- `npm run check`、`npm test`（46/46）、`npm run build` 和系统管理浏览器场景 2/2（54.2s）通过；浏览器场景分别覆盖 `systemd/firewalld` 可用和 Windows 能力不可用，使用受控 API，未改变真实主机。

## 2026-09-13 慢消费者诊断交付竞态

- Master 日志流在发送 `SLOW_CONSUMER` 诊断后立即关闭 WSS，客户端偶尔只能观察到 EOF。`internal/master/logs.go` 现在在错误帧成功写入后等待最多 100ms 或连接结束，再释放流，保持有界关闭语义并确保诊断可见。
- `go test ./internal/master -run TestLogSlowConsumerDeadlineAndCursorReattach -count=2` 在提权本地监听环境连续通过（91.364s）；修复前同测试 3 次中 2 次出现 45s 后 EOF。
- 同一场景 `go test -race ./internal/master -run TestLogSlowConsumerDeadlineAndCursorReattach -count=1` 通过（47.369s）。

`make build` 与 `make windows` 在本次后端改动后均退出 0，Linux Master/Daemon 及 Windows amd64 交叉产物已重新生成。

## 2026-09-13 云端工作区冲突提示

- Desktop 在云端工作区同步收到 `WORKSPACE_CONFLICT` 时明确提示“本地现场保留”，并引导刷新列表后选择副本；不会覆盖本地草稿或把旧云端版本误报为成功。`web/src/desktop/Desktop.vue` 类型检查通过。

## 1. 已知事实

- 已通过 DevSpace 打开 /data/instances/blora-panel；放入本次文件之前，该目录为空。
- 用户已经执行完整启动提示；保留原规划，现已有 Go Master/Daemon、运行适配、协议/存储和 Vue 桌面源码。
- Go/前端依赖已锁；Master/Daemon 与前端生产构建通过。真实 HTTPS fixture 已覆盖 43 项浏览器场景中的大部分核心链路（基础 fixture 37/43；扩展、性能和关闭网页负载另有独立通过证据），整项 F/A/E 仍受 Windows、危险主机、ENOSPC/掉电、远程负载等条件约束。
- 两个签名 Daemon、两个权限账号的 TLS API 集成已运行；任务去重、实际进程重启、管理重连、权限撤销流有子项证据。
- 最近两套本机浏览器验收 fixture 均已 Ctrl+C 停止；没有公开部署或操作生产节点。运行测试会话以第 7 节最新检查点为准。
- DevSpace 工作区标识可能随会话变化；接手时使用当前工具返回的实际工作区，不硬编码旧标识。

入口：[行动指导](../../ACTION_GUIDE.md) · [技术架构](../plan/Blora-01-技术架构.md) · [功能交互](../plan/Blora-02-功能与交互.md) · [验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)

## 2. 范围锁定

M0～M3 为本次完整任务的内部实施顺序。F01～F14、A01～A17、E01～E09 均需跟踪，不能只完成 M1 后将其他功能列为下一阶段。

保留的核心要求：管理通信与业务网络分离；默认/扩展应用统一宿主；跨节点全部获准实例与单独创建授权；应用多窗口多标签/移窗；实例桌面入口与专用窗口；完整刷新恢复；有界停止重启、真实退出确认；前后端分离、多用户多节点。

条件项和排除项以验收矩阵第 1 节为准。不得用“首版”“原型”“时间有限”自动删减明确范围。

## 3. 实施块台账

实现：未开始 / 进行中 / 已实现 / 阻塞。验证：未运行 / 通过 / 失败 / 环境缺失。每块可能覆盖多个 F/A/E，最终以验收矩阵细项为准。

| 块 | 内容 | 实现 | 验证 | 证据/下一动作 |
| --- | --- | --- | --- | --- |
| B00 | 工程、依赖、环境与最小构建 | 已实现 | 基线构建通过 | [核心报告](../acceptance/reports/core-2026-09-09.md)，生产运行文档在继续补全 |
| B01 | 共享协议、存储、迁移、持久化任务 | 已实现 | 子项通过 | 接受/派发持久化、去重、状态CAS、取消竞态与事件；故障组合仍在 B10 验证 |
| B02 | 桌面、AppHost、多窗口标签、恢复 | 已实现 | 子项通过 | Chromium Monaco中文/撤销重做/立即刷新/移窗、云端布局/内容选择和真实后端浏览器均有证据；完整跨设备/长时恢复仍在补 |
| B03 | 身份权限、节点通信与基础默认应用 | 进行中 | 用户/角色/维护/换钥子项通过 | [管理报告](../acceptance/reports/administration.md)；Users/Nodes真实浏览器通过，完整审计查看/保留等继续 |
| B04 | Linux/Windows/隔离运行与生命周期 | 进行中 | Linux子项通过；部分环境缺失 | [runtime报告](../acceptance/reports/runtime-2026-09-09.md)，Linux真实两节点重启；Docker/cgroup/Windows真运行未验 |
| B05 | 日志、PTY、终端及会话恢复 | 进行中 | 真实PTY/浏览器/Docker/原生日志与stdin子项通过 | 独立helper+16MiB归档、固定run输入任务；容器日志事件适配及完整分栏继续 |
| B06 | 文件、编辑、传输与草稿恢复 | 进行中 | 真实文件/上传/跨节点与浏览器子项通过 | [传输报告](../acceptance/reports/transfers.md)，含同宿主根别名/硬链接/源出生检查；ENOSPC等继续 |
| B07 | 实例中心、默认应用与任务体验 | 进行中 | 实例配置、列表/卡片视图、批量操作、任务历史/通知和自动启动子项通过 | [实例视图报告](../acceptance/reports/instances-view-2026-09-22.md)、[任务/通知证据](../acceptance/reports/notifications-2026-09-12.md)、[实例设置报告](../acceptance/reports/instance-settings.md)；整项仍受跨平台与故障组合验收条件约束 |
| B08 | Docker/Compose、监控、备份调度、系统管理 | 进行中 | Linux/Docker/Compose、监控历史、备份调度/恢复、受控系统管理子项通过 | [Docker日志归档报告](../acceptance/reports/docker-log-history-2026-09-22.md)、[长时监控报告](../acceptance/reports/monitor-history-soak-2026-09-21.md)、[调度阶段崩溃报告](../acceptance/reports/scheduler-staged-process-crash-2026-09-21.md)；Windows 主机适配、独立 systemd、远程/物理危险故障仍待相应环境 |
| B09 | 扩展 SDK、沙箱、包与安装生命周期 | 进行中 | 独立包安装/升级/回滚/卸载、沙箱能力门禁、WASI 任务/资源桥接、用户数据迁移子项通过 | [扩展全链路报告](../acceptance/reports/extensions-real-2026-09-13.md)、[扩展数据迁移报告](../acceptance/reports/extensions-data-2026-09-20.md)；远程目录运维、Windows 和长期跨平台组合仍待补 |
| B10 | 故障、性能、全范围集成和交付 | 进行中 | 真实套件通过，条件项未清零 | 真实 HTTPS + Docker fixture 15/15；继续整理 OpenAPI、故障/性能、Windows 与受控系统变更证据 |

当前产品统计：按“整项所有条件均有证据”的严格口径，功能 0/14、原规划场景 0/17、补充场景 0/9；大量子场景已通过，具体证据和未覆盖条件见验收矩阵，不把文档或测试替身计入通过。

## 4. 接手后的第一个具体动作

当用户明确发起实现任务时：

1. 阅读根目录 AGENTS.md、ACTION_GUIDE.md、两份规划和验收矩阵全文。
2. 检查项目目录的现状，尊重用户在本记录之后新增的文件和更改。
3. 检查 Go、Node、包管理器、系统、Docker/Compose 与浏览器测试能力；分别记录可用、缺失、未检查。
4. 根据固定技术基线建立 B00 工程和验证入口，完成后继续 B01～B10。
5. 按每个块的真实结果更新此文件和验收矩阵，不等待用户逐块确认。

若用户当前只要求修改规划，则只执行该文档任务，不因本文件而启动开发。

## 5. 环境、命令与证据

以下均为待填写槽位，不是已经运行的命令或成功结果。

| 项目 | 实际值 |
| --- | --- |
| 工作目录 | /data/instances/blora-panel |
| 主机系统/架构 | Fedora Linux 内核 6.19.12-200.fc43.x86_64；cgroup v2 当前路径不可写 |
| Go / Node / 包管理器版本 | go1.25.9 linux/amd64；Node v24.15.0；npm 11.12.1（选定npm） |
| 浏览器及 E2E 驱动 | Chromium 151.0.7922.71，Playwright CLI 1.60.0；真实HTTPS文件与PTY测试2/2通过13.3s，更多桌面证据见desktop报告 |
| Docker / Compose | Docker29.4.1、API1.54/min1.40、Compose5.1.3；容器、镜像、卷、网络、Compose 生命周期、日志/PTY 与桌面入口均已接入并有真实/隔离测试；远程 Engine、真实 ENOSPC 和危险主机组合仍未验证 |
| Windows 构建工具及真机运行环境 | Go交叉编译入口将提供；未发现Windows真机，运行验收仍环境缺失 |
| 后端构建命令 | make build；根Makefile设缓存及GOFLAGS=-buildvcs=false |
| 前端类型检查与构建命令 | cd web && npm run build；npm run check |
| 本地联合启动/初始化方式 | make fixture（先构建web）；仅本机9443，私有随机凭据见fixture目录 |
| 静态/单元/集成/浏览器测试命令 | make test/check；cd web && npm test / npm run test:e2e，需本机监听和Chromium权限 |
| 平台测试和发布命令 | `make windows`（Windows amd64 交叉编译）；Windows Job/ConPTY 真机待环境 |
| 扩展构建/安装测试命令 | `make sdk`；`go test ./internal/extensions ./internal/master -run 'Extension' -count=1`；真实 fixture `tests/real/extensions.spec.ts` |
| 证据目录 | `docs/acceptance/reports/` 下核心、桌面、Docker、SDK、系统、工作区和 OpenAPI 报告；真实凭据仅保留在 `.local/fixture-*` |

## 6. 活跃工作检查点

每完成一块、遇到重要失败或接近上下文边界时更新：

- 当前块和验收编号：B10 全范围验收与交付证据，关联 F01～F14、A01～A17、E01～E09；B00～B09 主链路已接入。
- 当前具体任务：B10 全范围回归与证据收口；UI 样式按用户要求冻结。Docker 日志归档入口的本地实现、浏览器替身、真实 Master/Daemon API 与真实 Chromium/Docker UI 已闭环；原 Compose 长流程也已复验通过。本轮完成 B07～B09 本地入口审计、前端/Go/Windows 构建复核；下一步转为维护外部环境验收清单，不重复已通过的 UI 视觉套件。
- 本轮修改/证据文件：`web/src/apps/DockerLogs.vue`、`web/tests/browser/docker.spec.ts`、`web/tests/real/docker.spec.ts`、`internal/master/containers_integration_test.go`、`docs/acceptance/reports/docker-log-history-2026-09-22.md`；验收矩阵与本进度文件。真实 fixture 已 Ctrl+C 停止，遗留的本轮 Compose 容器/卷按标签清理，当前没有活动测试或 fixture 进程。
- 最新成功验证：真实 Docker 日志归档 UI 1/1（7.6s）；原真实 Compose 长流程 1/1（45.6s）；真实 Master/Daemon 归档 API race 2.79s；前端 `vue-tsc`、63/63 单测、构建、`make check` 仍通过。
- 最新失败验证：曾有一次把 Compose 容器选取/日志换行/容器启动条件写错的验收脚本失败，均已修正并在最终真实场景通过；不记录为产品失败。矩阵仍保留真实远程 Engine、Windows 容器和物理故障等未验证边界。
- 尚未解决的缺口：F01～F14均未整项通过；A09/A13仍进行中；E01～E09均未整项通过。跨平台缺Windows Job/ConPTY真机和容器；E05缺可操作独立systemd manager；还缺远程 Engine/真实高延迟、物理掉电/Engine主机存储耗尽、长期墙钟故障及非Chromium平台组合。E08 当前源码一小时本地目标已达到 p95 48.1ms，但公网/高 RTT/丢包、Windows 和 WebKit 仍未验证。详见验收矩阵，不以构建或替身覆盖环境缺口。
- 下一条具体动作：当前环境中未发现还应补写生产代码的本地缺口；若提供 Windows 真机、独立 systemd manager 或远程受控节点，按矩阵执行对应验收。Windows Job/ConPTY、独立 systemd、远程高 RTT/丢包、物理掉电和长期跨平台场景继续维持环境缺失记录，不用 Linux 本地结果替代。
- 运行中服务/测试会话及接管方式：当前无活动 fixture；真实回归使用 `.local/fixture-*` 私有凭据并在结束后 Ctrl+C 停止。宿主 PTY 测试需显式 `BLORA_HOST_EXEC_HELPER_BINARY` 与隔离镜像；防火墙真实改动不在本机执行。
- 2026-09-09 扩展升级安全性：升级前保存旧清单与包体，写入失败恢复旧包体；版本限定为 `v?数字.数字.数字` 并拒绝降级；管理员新增 `POST /api/v1/extensions/{id}/rollback`，卸载/重新安装会清理旧回滚状态，生命周期回滚测试及 Master TLS 集成测试通过。验证：`go test ./internal/extensions -count=1`、`go test ./internal/master -run TestExtensionAdminAPIInstallsAndServesPackage -count=1`（提权后通过）。
- 同日扩展契约校正：Master 注册源允许 SDK 使用的 `resource.write`/`task.create` 能力，并对入口、资源处理器、窗口/标签策略做边界校验；全仓编译与 vet 复验通过。
- 同日系统管理增量：计划任务支持跨平台 enable/disable 命令、任务名路径穿越防护、Master `POST /api/v1/nodes/{id}/system/tasks/actions` 和 Daemon 持久任务执行器；普通用户仍受 `host.manage` 门禁。边界及相关模块回归通过，真实主机变更和失联回滚未执行。
- 同日扩展禁用门禁：`Manager.AuthorizeCapability` 每次从持久清单重新读取启用状态和能力，禁用后新能力请求立即拒绝；生命周期测试覆盖该语义。真实扩展任务桥接仍待接入。
- 同日独立包交付：参考扩展新增 `npm run package`，生成 `reference.blora-extension.json`（清单、SHA-256、base64 包体）；Master 新增受限 `POST /api/v1/extensions/install-package` 原始包安装入口，TLS 集成测试验证独立包安装与完整性校验。
- 同日外部清单兼容：注册源持久化 icon/color/permissions、入口和窗口策略字段并拒绝 protected 外部清单，避免 AppHost 加载时因缺少权限字段崩溃；参考包已重新生成并通过安装集成测试。
- 同日参考扩展能力示例：示例源码加入 `task.create` 能力声明和 `enqueueExampleTask` 门禁调用，重新执行 `npm run package` 成功；该示例仍等待真实宿主任务桥接，未将 transport 替身计为 E06 完成。
- 同日容器日志生命周期：Daemon 仅在 `container.delete` 得到成功状态后删除对应节点归档，失败/未知结果保留历史；归档清理测试与相关集成回归通过。
- 同日浏览器回归：`npm run test:e2e -- --workers=1 --reporter=line` 完整 21/21 通过（约 1.5m）；此前中断的输出已由明确终态补证。该套件覆盖刷新、窗口/标签、编辑、文件、日志、终端、传输和权限交互，但不代替真实 Master/Daemon 混合负载性能测试。
- 同日全仓 Go 回归：`go test ./...` 全部通过（Master 48.213s，含现有集成场景）；未将测试替身或交叉编译当作生产环境验收。
- 同日 E08 探针：新增 Playwright 8 窗口本地负载测试；独立运行 p95 50.8ms，完整 22 项套件内 p95 75.4ms（heap 约68MB，4～5/60 长帧），显示负载波动并高于 50ms 目标。完整浏览器套件 22/22 通过；报告明确不覆盖真实终端/万条节点目录/后台传输与 RTT。E08 仍进行中。
- 同日 E08 探针增量：通过任务栏真实右键“新建窗口”建立第二个终端窗口，探针测得 p95 69.5ms、最大73.9ms、5/60 长帧、heap 约64MB；窗口未绑定后端会话，不能替代两条活跃终端验收。
- 同日防火墙受控应用：Linux firewalld 端口规则支持 256 条以内、60 秒确认期限和逐条失败反向回滚；Master/Daemon 任务权限与过期确认测试通过，全仓编译通过。未在当前宿主执行危险改动，租约级失联自动回滚仍待实现。
- 同日防火墙取消语义：应用 reload 成功后若任务上下文已取消，仍执行反向回滚并报告取消；系统信息模块、Master 集成与全仓编译复验通过。
- 同日防火墙回滚测试增强：引入仅测试可替换的命令执行器，隔离模拟 firewalld 在删除规则中途失败，验证已完成新增/删除操作按逆序恢复；生产默认仍使用 `exec.CommandContext`。
- 同日持久化边界：容器日志归档及扩展清单/包体 rename 后同步父目录（Windows 跳过不支持的目录 Sync），相关模块测试和 `make windows` 通过。
- 同日系统信息 race 回归：`go test -race ./internal/systeminfo -count=1` 通过，覆盖防火墙输入与隔离命令回滚测试。
- 同日防火墙诊断：回滚命令失败不再静默忽略，错误同时保留原始操作失败和 `firewall rollback failed` 诊断；隔离测试覆盖该分支。
- 同日平台回归：`make windows` 在新增 firewalld 分支后仍成功生成 Master/Daemon Windows 产物；Windows 实机行为仍未验证。
- 同日系统模块完整回归：`go test ./internal/systeminfo ./internal/master ./internal/daemon -count=1` 与 `go vet ./...` 通过；未执行当前宿主的危险防火墙变更。
- 同日系统管理前端：SystemApp 新增计划任务启用/停用和防火墙受控应用确认按钮，提交 60 秒确认期限；`npm run check && npm run build` 通过。真实主机变更仍未执行。
- 同日系统管理浏览器证据：新增 `web/tests/browser/system-management.spec.ts`，用完整 API 合同验证节点选择、计划任务启用和防火墙确认提交；Playwright 定向场景 1/1 通过。证据记录于 `docs/acceptance/reports/system-management-2026-09-09.md`；真实危险主机变更与租约失联回滚仍未验证。
- 同日全量浏览器复跑：新增场景后尝试 23 项串行套件时，测试开发服务器在首个用例后连接被拒（23 项均报 `ERR_CONNECTION_REFUSED`/trace 资源错误），因此不计为通过；定向系统管理场景仍稳定 1/1，通过既有 22/22 基线保持有效。后续需在稳定浏览器服务环境重跑全套。
- 同日全量浏览器复跑修复：清理本轮 trace 产物并以 `--trace=off` 串行执行，完整 23/23 通过（约 1.2 分钟）；性能场景 p95 16.9ms、最大 71.7ms、1/60 长帧、heap 约64.7MB。前次连接拒绝归因于本机 trace/开发服务器环境竞争，E08 仍受真实后端负载场景限制。
- 同日 Go 回归复核：首次 `go test ./...` 出现两个瞬态集成竞态（进程 `/proc` 已退出、备份节点通道尚未连接）；对应单独测试及完整 `go test ./internal/master -count=1` 随后通过（48.062s）。未把首次失败隐藏，后续应继续降低测试启动同步敏感性。
- 同日测试同步修复：`fileFixture.restartNode` 现在同时等待新 `StartupID` 和 bulk 通道协商完成，避免重启后首个备份/文件请求偶发 503；`TestBackupAPIRealArchiveConditionalRestoreAndRevocation` 定向通过。
- 同日扩展管理浏览器证据：新增 `web/tests/browser/extensions.spec.ts`，管理员生命周期操作禁用、回滚、卸载均走真实 API 边界并断言禁用请求体；定向 Playwright 1/1 通过。真实扩展代码加载与任务桥接仍待实现。
- 同日扩展任务桥接：新增 `POST /api/v1/extensions/{id}/tasks`，服务端复核启用状态与 `task.create`、节点 `host.manage`，payload 限制 64 KiB；Daemon 仅返回不执行的扩展数据结果并拒绝越界。扩展 API 集成测试已通过，SDK README 记录调用合同。
- 同日扩展拒权回归：集成测试增加“禁用扩展后新 task.create 返回 409”断言，验证禁用门禁覆盖任务入口；测试通过。
- 同日前端宿主 transport：新增 `web/src/app-host/extension-runtime.ts`，将 SDK `task.create` 限定映射到 Master 扩展任务 API，并复用 CSRF/幂等 API 客户端；未知能力和缺失 app/node 在浏览器侧拒绝。`npm run check && npm test`（20/20）通过。
- 同日扩展桥接静态审计：`go vet ./...` 通过；清理不可达的通用 `app.use` 映射，`extension.task` 仅由专用授权分支校验扩展能力和节点管理权限，集成测试复验通过。
- 同日扩展边界回归：集成测试新增普通用户无 `host.manage` 时 403、payload 超过 64 KiB 时 413 的断言；扩展任务桥接测试通过。
- 同日交付说明校准：README 更新为当前已接入监控、系统管理和扩展生命周期的实际状态，并保留动态加载、完整沙箱恢复及网络注册源等明确缺口。
- 同日监控浏览器证据：新增 `web/tests/browser/monitor.spec.ts`，验证节点选择、进程搜索、内存排序及带 `startTicks` 的终止提交；定向 Playwright 1/1 通过，报告已更新。
- 同日监控实例视图：MonitorApp 增加实例资源引用下的 `/instances/{id}/metrics` 查询，展示实例 CPU/内存和 stale 标记；`npm run check` 与监控浏览器场景均通过。
- 同日监控前端构建回归：`npm run build` 成功（Vite 1755 modules）且 `npm test` 5 个文件 22/22 通过；仅保留既有大 chunk 警告，不影响构建结果。
- 同日监控缓存修复：实例指标 queryKey 改为响应式，窗口内切换实例不再复用旧实例采样；`npm run check` 与监控浏览器场景 1/1 通过。
- 同日标准检查入口：`make check`（`go vet ./...`）通过。
- 同日浏览器全量回归：纳入扩展管理、监控和系统管理新增场景后，`npm run test:e2e -- --workers=1 --trace=off --reporter=line` 完整 25/25 通过（约 1.2 分钟）；性能探针 p95 17.1ms、最大54.3ms、1/60长帧、heap约64.1MB。
- 同日性能测量增量：探针增加 5 次 API 请求 RTT 采样；最新独立运行 API RTT p95/max 8.7ms，交互 p95 16.8ms、最大61.1ms、1/60长帧、heap约65.1MB。明确标注为本地 fixture 边界，不替代远程 RTT/带宽负载。
- 同日 Master race 全量回归：`go test -race ./internal/master -count=1` 通过（74.668s），覆盖扩展任务、节点重启、权限、文件/终端/备份/容器集成场景。
- 同日进程消失竞态加固：增加 `processGone`，兼容 `/proc` 的 ENOENT/ESRCH 和系统错误文本，避免日志停止时快速退出被误报；runtime 定向测试及三项 Master 生命周期场景重复 3 次均通过。
- 同日回归确认：此前全仓回归中失败的 `TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop` 在修复后重复 3 次通过（2.101s）；后续全仓测试仍需再跑一次以更新总状态。
- 同日全仓 Go 回归完成：`go test ./...` 通过（Master 46.624s，其余包通过/缓存），确认进程竞态和扩展最小权限改动未引入总回归。
- 同日 Linux 进程竞态修复：运行时在弱隔离进程扫描、cgroup 成员读取和 pidfd 信号前统一使用 `os.IsNotExist` 识别进程瞬时退出，避免 `/proc/<pid>/stat` 消失被误报为 restart 失败；runtime 全测试和双节点生命周期集成测试均通过。
- 同日竞态 race 复验：`go test -race ./internal/runtime -count=1` 与 `go test -race ./internal/master -run TestTwoNodesAuthorizationAndDurableLifecycle -count=1` 均通过，分别 2.603s/5.868s。
- 同日 Windows 产物复建：`make windows` 成功生成 `dist/blora-master.exe` 与 `dist/blora-daemon.exe`；Windows Job/ConPTY 真机运行仍属环境缺失，交叉编译不计作运行验收。
- 同日宿主 transport 单测：新增 `web/tests/extension-runtime.test.ts`，覆盖 `task.create` API 路径、CSRF、payload 保真及未知能力/缺失 scope 的本地拒绝；`npm test` 5 个文件 22/22 通过。
- 同日宿主合同扩展：`createExtensionTransport` 支持宿主显式注入 `window.open`、`state.capture`、`state.restore` hooks，未注入能力仍本地拒绝；`npm run check && npm test` 通过（5 文件 23/23）。
- 同日扩展最小权限修正：扩展任务不再要求 `host.manage`；新增扩展资源 `app.use` 授权类型并结合目标节点 `node.read`，集成测试验证获准普通用户可提交、未获准用户仍 403。
- 同日扩展包加载边界：新增认证 `GET /extensions/{id}/bundle`，仅已启用扩展可读取已校验包体，禁用返回 409；前端新增 `fetchExtensionBundle`，为 sandbox loader 保留 no-store API 边界。Master 集成与前端 `npm run check && npm test`（24/24）通过。
- 同日扩展 sandbox loader：新增 `ExternalSandboxApp.vue`/`loadExternalSandboxApp`，在 `sandbox="allow-scripts"` opaque-origin iframe 执行 bundle，以 `postMessage` 接入受限 transport；前端检查和 24/24 单测通过。资源标签/状态恢复端到端仍待验证。
- 同日 sandbox 构建修复：避免 SFC script 中直接出现 `</script>` 导致模板解析截断，改用运行时拼接闭合标签；`npm run build` 与 `npm test`（24/24）通过。
- 同日 SDK 安全文档：README 明确 bundle 只能经认证接口读取并放入 `sandbox="allow-scripts"` iframe，通过 `postMessage` transport，禁止注入宿主主页面。
- 同日编辑器恢复稳定性复验：IME、多光标、查找替换及即时刷新场景单独重复 5 次全部通过；此前完整套件中一次撤销时序失败未能复现。扩展 sandbox loader 改动后的完整套件已有关闭 trace 的 25/25 通过记录，后续若资源允许再做一次完整串行复跑。
- 同日最终 Go 回归：使用临时 GOCACHE 并允许 httptest/WSS 本地监听后，`go test ./...` 全部通过（Master 47.150s）；无权限的默认沙箱运行仅受监听策略阻断，非代码失败。
- 同日扩展恢复链路：工作区初始化扫描未知外部 appId 并经认证 manifest 动态注册 sandbox host；sandbox 组件按 viewTab resourceRef 动态解析 nodeId，支持跨节点标签复用。`npm run check`、`npm test`（24/24）与 `npm run build` 通过。
- 同日扩展浏览器回归：`npm run test:e2e -- --workers=1 --trace=off --grep 'extension management installs' --reporter=line` 1/1 通过。
- 同日扩展入口补齐：扩展管理窗口新增已启用扩展“打开”动作，复用动态 sandbox 注册并避免重复注册；前端检查与 24/24 单测通过。
- 同日扩展能力门禁：sandbox transport 根据 manifest 在调用宿主 hook/API 前检查 `window.open` 与 `task.create` 声明；新增拒绝测试，前端 `npm test` 25/25 通过。
- 同日防火墙回滚修复：ApplyFirewall 在父 context 取消后使用独立 10 秒回滚 context，避免断线取消让恢复命令全部提前失败；`go test ./internal/systeminfo -count=1` 通过。
- 同日 clean 交付构建：`make sdk` 完成 SDK 与 reference app 的 clean `npm ci`、类型检查和构建；`make web` 完成前端 clean `npm ci` 与生产构建，均退出 0。
- 同日真实 Docker/Compose E2E：Docker 29.4.1/API 1.45、Compose 5.1.3 下运行 `TestRealDockerComposeLifecycle`，25.651s 退出 0；覆盖容器/镜像/卷/网络、Compose 应用更新删除、日志、失败诊断与卷保留，报告见 `containers-e2e-2026-09-09.md`。
- 同日宿主 Docker PTY E2E：使用隔离镜像与 exec-helper 运行 `TestDockerExecProtocolOwnershipPTYAndClose`，退出 0；验证 PTY 协议所有权与关闭路径，结果已写入 runtime 报告。
- 同日实例监控失联回退：新增每实例 120 点持久指标缓存；节点断连时 `/instances/{id}/metrics` 返回最近点并标记 `stale`，避免旧值被误认为实时。监控定向集成测试通过（0.924s），矩阵和报告已更新。
- 同日 Docker 浏览器边界：新增节点选择、容器查询、显式启动确认、幂等任务提交和完成回读场景；`npm run test:e2e -- --workers=1 --trace=off --grep 'Docker app queries' --reporter=line` 1/1 通过（5.4s），报告见 `docker-browser-2026-09-09.md`。
- 同日监控 race 复验：`go test -race ./internal/master -run 'TestMetrics|TestMetricHistory|TestInstanceMetric' -count=1` 通过（3.239s）。
- 同日跨节点传输回归：`GOCACHE=/tmp/blora-go-transfer-cache go test ./internal/master -run 'TestTransfer' -count=1` 11 项通过（25.830s）；更新 F07/A12 证据，真实 ENOSPC 与浏览器关闭组合仍待验证。
- 同日标准 race 入口：`make test`（`go test -race ./...`）在允许本地 TLS/WSS 与辅助进程环境中完成，backup、containers、daemon、extensions、filesystem 等包均通过；未将环境缺失的 Windows/危险主机场景计入通过。
- 同日扩展现场能力接线：sandbox 将 `window.open`、`state.capture`、`state.restore` 连接真实桌面 store，返回实际窗口/标签身份并持久化 view 状态；总览扩展无节点时仍允许开窗，任务调用明确拒绝缺少 nodeId。前端检查、25/25 单测和构建通过。
- 同日扩展总览边界测试：新增无 nodeId 场景，验证总览扩展可开窗/恢复但任务调用被拒绝；前端 `npm test` 26/26 通过。
- 同日扩展节点身份修复：sandbox 每次 bridge 调用重新读取当前 viewTab resourceRef，避免跨节点切换/移窗后复用旧 nodeId；`npm run check && npm test`（26/26）与构建通过。
- 同日扩展 sandbox E2E：新增真实 bundle/opaque iframe/`postMessage` 场景，发现并修复 Vue 响应式状态无法结构化克隆导致的 bridge 中断；浏览器测试 1/1 通过（3.5s），SDK 报告已更新。
- 同日扩展浏览器组合回归：管理生命周期与 sandbox bundle 两项场景串行 2/2 通过（5.0s）。
- 同日 B10 浏览器全量回归：新增 Docker 与 sandbox 场景后完整串行套件 27/27 通过（约 1.3 分钟，退出码 0）；性能探针交互 p95 16.8ms、最大 60.3ms、1/60 长帧、heap 64,318,483 字节，API RTT p95/max 32.7ms，性能报告已更新。
- 同日资源恢复复核：桌面恢复/工作区切换现在为每个带 resourceRef 的 view 调用 `reconcileResource`，复核失败保留原现场并记录异常；改动后完整浏览器套件再次 27/27 通过，性能 p95 16.8ms、最大 62ms、1/60 长帧、API RTT p95/max 4.5ms。
- 同日最终后端 race 复验：`GOCACHE=/tmp/blora-go-race-final go test -race ./...` 退出 0，Master 75.886s，其余所有包通过；确认实例指标缓存和资源恢复相关改动无数据竞争。
- 同日监控改动后全仓 Go 回归：`GOCACHE=/tmp/blora-go-final-cache-2 go test ./...` 退出 0，Master 47.206s，所有包通过。
- 本轮创建的测试数据、进程、容器、网络规则及清理方式：fixture三个二进制与自身两个实例；停止用Ctrl+C，仅清理自身实例进程。新增专用镜像blora-isolated-e2e:20260909，manifest sha256:6486725488fcea62b3d2a2da0a395434a359a7642b536098a6cacfa04b0526c9；真实测试容器按随机run标签清理，未报清理错误。未改已有生产容器或防火墙。镜像与fixture目录保留作复现产物。
- 外部阻塞及影响范围：本机网络监听/浏览器/宿主测试需要工具提权，由auto_review按授权处理；Windows真机与可写cgroup委派仍缺失。Docker只读版本可用，真实容器待获准测试镜像环境验证。
- 2026-09-10 真实浏览器全链路复验：`cmd/devfixture` 在隔离 `.local/fixture-3172017811` 启动 Master 与两个 Daemon，并接入 Docker 29.4.1；`files-terminal.spec.ts` 定向 2/2（27.1s），`npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` 完整 13/13 通过（2.2m，退出码 0）。覆盖真实登录、节点/实例、文件、终端/PTY、日志、设置、模板创建、账号权限、Docker 浏览器边界及扩展链路；fixture 已用 Ctrl+C 停止，未留下活动会话。
- 2026-09-10 Docker 停止边界修复：Engine transport 的 `ResponseHeaderTimeout` 从 10 秒上调至 5 分钟，覆盖 Docker stop 的 300 秒服务端等待上限，单请求仍由 context 截止；`go test ./internal/containers -count=1` 通过。真实 Docker 浏览器 Compose 用例改为先从启动器打开 Docker 应用并提交 inline Compose 草稿，真实 apply/update/delete 1/1 通过。
- 2026-09-10 WSS 重连竞态修复：Master `nodeCall` 在控制连接已恢复但 bulk/interactive 数据链路尚未登记时等待最多 2 秒；控制连接消失立即返回 `NODE_UNREACHABLE`，请求取消仍尊重调用方 context。`go test ./internal/master -count=1` 与完整真实 13/13 套件通过；远程高延迟链路仍需单独测量。
- 2026-09-10 扩展回滚元数据修复：`Manager.List` 跳过 `.previous.json` 回滚快照，`read` 拒绝清单内 `appId` 与文件名不一致，避免升级回滚文件被误报为第二个已安装扩展；`go test ./internal/extensions -count=1` 通过。
- 2026-09-10 本地注册源与真实扩展回归：`cmd/devfixture` 预置 `fixture.reference` 的 `v1.0.0`/`v1.1.0` 包；真实 HTTPS 场景安装两版、读取启用 bundle、禁用返回 409、重新启用、opaque sandbox 执行并卸载，`extensions.spec.ts` 1/1 通过。静态 bootstrap 改为从同源 `/blora-extension-bootstrap.js` 加载，再在 opaque iframe 内创建 bundle URL；Master CSP 保持禁止 inline script，仅允许 `self` 与 `blob`。
- 2026-09-10 云端工作区回归：新增布局/引用默认同步、显式正文同步和云端副本打开的 mock 与真实 HTTPS 场景；`workspaces.spec.ts` 真实 1/1 通过，默认同步剥离草稿，显式同步后 revision=2 并恢复正文。
- 2026-09-10 OpenAPI 交付描述：新增 `docs/api/openapi.yaml`，覆盖 94 个路径、108 个操作以及 session/CSRF/idempotency、任务、扩展、系统和工作区合同；Python `yaml.safe_load` 校验通过，README 与验收矩阵已链接。
- 2026-09-10 防火墙租约终态修复：直接执行 `runFirewallApply` 的过期/取消回滚现在持久化终态阶段；回滚开始前先写入 `rollback_in_progress`，重启可继续识别并恢复。普通与 race 定向租约测试均通过。
- 2026-09-10 B10 真实全量复验：在 `.local/fixture-1866434427` 启动真实 HTTPS Master、两个 Daemon 和 Docker 29.4.1，`BLORA_E2E_CREDENTIALS=... npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` **15/15 通过，2.2 分钟，退出码 0**；随后 Ctrl+C 停止 fixture。场景包含账号、桌面/权限、文件/PTY、日志/传输、节点维护、设置/编辑器、Docker/Compose、扩展两版本 sandbox 和云工作区。
- 2026-09-10 全仓验证：提权允许本地监听后 `make test` 的 `go test -race ./...` 通过（Master 79.757s），`make check`、前端 `npm run check && npm test`（28/28）、`make windows` 和 Linux 三个二进制构建均通过。默认沙箱运行 race 测试时的 `httptest` IPv6 监听失败属于权限限制，未计作代码失败。
- 2026-09-10 审计与节点维护增量：新增 `003_audit_node.sql`，把审计时间统一为毫秒并回填节点身份；`Store.Audits` 和 `GET /api/v1/audit` 支持管理员按用户、节点、资源、动作、RFC3339 时间及分页检索，UsersApp 增加筛选界面。节点撤销改为管理员确认后的事务幂等回执，递增代次、保留实例资源并写入节点审计，NodesApp 增加“撤销管理”入口；节点设置增加分组/标签且心跳保留管理员字段；新增存储和 Master TLS 集成测试均退出 0。新增问题修复：扩展 sandbox 不再把实例 ID 当作 nodeId；OpenAPI 更新为 95 路径、109 操作。证据见 `docs/acceptance/reports/audit-2026-09-10.md`。
- 2026-09-10 审计横向回归：任务接受和终态写入审计事件但不写入任务正文；节点/实例资源保留节点维度。`make test` 提权 race 全仓通过（Master 77.778s），`make check`、前端 28/28 与最终生产构建均通过；审计证据报告已补充。
- 2026-09-10 接口合同校正：OpenAPI 将节点撤销改为实际同步 `200` JSON 响应并补充 `400/404/409` 错误，登记者票使用 `201`、退出使用 `200`；YAML 重复键检查及路径/操作计数复验为 OpenAPI 3.1、95 路径、109 操作。
- 2026-09-10 备份与定时任务界面增量：默认注册 `blora.backups`，实例窗口提供真实入口；创建备份先读取源版本，恢复先生成不可变计划并要求 `overwritePlanHash` 显式覆盖确认，保留策略与定时任务使用真实 API、任务回执、幂等键和修订号。新增浏览器场景 1/1 通过；前端 28/28、check 和生产构建通过。Daemon 新增可选固定 argv `backupHookCommands` HookRunner，未配置时 save/pause/stop/hooks 仍明确返回能力不可用；真实 ENOSPC/掉电/长时调度和具体应用 hook 语义继续未验证。证据见 `docs/acceptance/reports/backup-ui-2026-09-10.md`。
- 2026-09-10 远程扩展目录源增量：新增 `extensions.CatalogSource` 与可配置 `RemoteCatalog`，只接受无凭据/无查询的 HTTPS 基址，固定读取 `index.json` 和版本包，限制响应大小、拒绝重定向、校验目录身份/摘要/重复项；Master `Options.ExtensionCatalogURL` 与 `cmd/master --extensions-catalog-url` 接入，目录与 URL 互斥且仍由管理员 API 调用，安装再次经 Manager 的完整性/签名校验。`GOCACHE=/tmp/blora-go-catalog2 go test ./internal/extensions ./internal/master -run 'TestRemoteCatalog|TestExtensionRemoteCatalogAdminAPI|TestExtensionCatalogBrowseAcquireAndUpgrade' -count=1` 通过；完整 `go test -race ./...` 通过，前端完整浏览器套件本轮更新为 30/30 通过。真实远程部署、高延迟证书轮换仍未验证。
- 2026-09-10 任务重试增量：任务中心新增失败/中断实例生命周期操作的关联重试 API。Master 仅重建当前实例配置并重新授权 `start/stop/restart/kill`，拒绝文件、终端、传输、备份和主机任务的通用重放；重试使用 `RetryOf`、独立幂等键和持久任务回执，权限边界与实例资源身份再次核对。新增 TLS Master 集成测试验证真实失败、关联任务、重复请求复用和未授权隐藏；任务中心 UI 显示已确认结果、关联重试入口和刷新后保留的请求身份。`GOCACHE=/tmp/blora-go-retry2 go test ./internal/master -run TestTaskRetryCreatesRelatedInstanceAttemptAndIsIdempotent -count=1` 通过；OpenAPI 更新为 96 路径、110 操作。远程节点故障下的长期重试组合仍待真实环境验证。
- 2026-09-10 任务列表/传输索引增量：任务中心 GET 支持 `offset`/`limit`（输出上限 1000 与 `nextOffset`），新增 `/transfers` 根任务列表并逐项复核传输源读与目标写权限，保留终态用于对账；真实传输集成测试覆盖新列表路由。列表不会绕过任务中心既有授权，也不会把内部传输子任务泄露为根任务。
- 2026-09-10 最终回归：备份列表/恢复计划分页拒绝负偏移，创建/清理/恢复明确要求 `Idempotency-Key`；真实备份归档/恢复定向测试通过。随后 `GOCACHE=/tmp/blora-go-race-final2 make test` 的全仓 race 通过（Master 79.536s），`make check` 与 OpenAPI 3.1 解析复核（96 路径、110 操作）通过；前端最新完整回归为 Vitest 28/28、Playwright 30/30，性能探针交互 p95 16.9ms、最大71.9ms、1/60 长帧、heap 63,210,217 字节、API RTT p95/max 6.6ms。进度文档同步修正 Docker/Compose 已接入的历史误记；Windows 真机、远程/故障负载和整项验收缺口仍按矩阵保留。

## 7. 决策记录

2026-09-13 A16入口基础组合：desktop.spec.ts新增真实入口专用窗口复用/显式新开，改名刷新保留但实际实例名/状态/runId不变，删除入口不删资源且无启动重启删除任务。36210退出0，1/1（3.6s），fixture1929604743会话9013已发送Ctrl+C，收取退出；无生产改动。下一动作：新建专用测试实例用于目标改名/删除/同名替换，验证旧resourceRef不会指向新资源，并补成员撤权/离线入口显示；不删除fixture基础实例影响其他场景。A16尚未全项完成。

2026-09-13 A16完整入口身份闭环：新增受控管理员 `DELETE /instances/{id}` 软删除墓碑，停止且无活动任务才可执行；列表/读取隐藏墓碑，Daemon重连快照忽略已删身份，配额统计排除墓碑，审计保留 instance.delete。存储删除幂等/运行中及活动任务冲突测试通过。真实 Chromium fixture-4074818222：目标改名/删除/同名替换旧resourceRef失效1/1（7.0s），双账号撤权入口失效1/1（3.3s），暂停精确Daemon入口显示节点失联并原runId恢复1/1（50.4s）；与基础专用窗口复用/显式新开/入口改名删除1/1（3.6s）合计覆盖A16全部条件，矩阵更新Linux已实现。fixture会话61153已Ctrl+C停止，无活动服务。下一具体动作：按矩阵审计A13/A14/A09/A12及E项剩余证据，优先补可推进的真实权限/资源迁移与长期负载；重新生成最终发行包，Windows真机和物理故障仍如实保留。
2026-09-13 A14创建权限闭环：真实 fixture-3663732060 双节点双账号浏览器1/1（3.5s）验证 member 可见两实例但创建入口禁用；仅节点2授予 `instance.create` 与 native 所需 `host.manage` 后 `/nodes/creatable`/创建向导仅列节点2，确认前撤权真实403且无实例，恢复授权后创建STOPPED实例并管理员按ID清理。报告与矩阵更新A14已实现（Linux Chromium）；fixture会话89877已Ctrl+C停止，无活动服务。下一具体动作：继续A13扩展完整能力/迁移、A09长期内存与A12 ENOSPC/物理故障等可推进验证，再更新最终打包与全量回归；Windows/远程环境仍未验证。

2026-09-13 A01立即刷新闭环：desktop.spec.ts新增活跃/后台参数场景，真实CDP IME提交、系统剪贴板CtrlV、连续输入撤销、鼠标拖窗、切后台或保持活跃后直接reload，正文/精确geometry/焦点/redo保留，再末次输入直接reload通过。16486退出0，2/2（6.0s）；fixture4013429189会话65859已发送Ctrl+C，收取退出。A01矩阵更新Linux覆盖，无生产改动。下一动作：A06实际vim/top与ANSI序号旧报告核对后更新覆盖；A13独立SDK资源任务恢复/未授权能力及A14/A16入口权限完整组合逐项审核。A09/A12和E项仍有明确未验证条件。

2026-09-13 A05传输边界闭环：收紧TestTransferDisconnectAndMidstreamPermissionRevocation为RUNNING且0<offset<total才断真实bulk连接，恢复后Verified/全字节一致/固定上传提交子任务恰一，96813 race退出0（12.490s），同时中途撤权未发布目标通过。结合restart断线原请求及实际OFFLINE/WAITING_NODE浏览器，A05矩阵更新Linux完整分支。无生产改动/活动测试。下一动作：核对A01活跃/后台窗口最近输入立即刷新、A06实际模式/序号以及A13/A14/A16已有证据缺口，按真实当前实现补齐；A09长时内存、A12物理ENOSPC及E项仍保留，不能宣布全量完成。

2026-09-13 A05任务卡等待：扩展actual paused Daemon，真实OFFLINE期间提交stop202，原任务WAITING_NODE并生产任务中心完整taskId卡显示等待节点；finallySIGCONT后原任务SUCCEEDED/卡成功/实际STOPPED，runId保持。30032退出0，1/1（50.7s），fixture3110999777会话42086已发送Ctrl+C，收取退出。无生产改动。下一动作：汇总A05重启/传输实际断线证据与新浏览器呈现，确认各阶段无重复副作用，再处理A01/A06/A13/A14/A16未归档完整条件；E项和全量最终回归/重新打包仍未完成。

2026-09-13 A05真实浏览器失联：nodes.spec.ts新增actual paused Daemon，完整配置路径/出生标识唯一识别自身进程，SIGSTOP后真实OFFLINE与页面等待提示，不伪造STOPPED；finallySIGCONT后ONLINE/原runId，再实际stop。32702退出0，1/1（50.5s）；fixture3470920766会话72955已发送Ctrl+C，收取退出。无生产修改。下一动作：任务中心在节点失联期间的排队/等待确认可见状态及重连结果，不仅验证资源页告警；结合restart/transfer后端断线证据审计A05。A01/A06/A13/A14/A16与E项仍须继续。

2026-09-13 A10身份边界闭环：新增真实存活PID错误StartTicks记录，Observe/ForceStop均ErrUnknown且正确身份原运行存活；结合未知token拒绝、丢PID的marker恢复、manager重建四项67725 race退出0（1.058s）。该项明确为过期身份注入，不声称实际PID复用。结合独立Daemon稳定运行/STOPPING SIGKILL，A10矩阵更新Linux覆盖；无生产改动，无活动测试。下一动作：A05实际浏览器节点断开/恢复的失联待确认呈现，结合已有restart/transfer断线避免重复副作用；A01/A06/A13/A14/A16及E项仍需核对，不宣布全量完成。

2026-09-13 A10任务中途SIGKILL：package-smoke新增--daemon-task-crash，27806退出0；/tmp/blora-release-smoke-fnx8gqb6，实际STOPPING阶段kill所属Daemon并确认-9，原目录恢复后原任务INTERRUPTED/daemon_restarted_reconcile_required，同key仍原任务且不重放，原runId/计数x保持；finally显式stop SUCCEEDED并退出所有自建进程，无活动会话。下一动作：核对A10运行身份防PID误认直接测试与恢复状态UI证据，之后A05浏览器失联/等待而非假停止组合；其他范围与最终重打包仍继续。

2026-09-13 A10独立Daemon SIGKILL：package-smoke新增--daemon-crash，44796退出0；/tmp/blora-release-smoke-zoh3sha_，原运行计数x，所属Daemon kill确认-9，同配置重启新startupId/同nodeId上线，原runId RUNNING及x保留。随后真实restart+Master SIGKILL恢复原任务，计数xx；最终停止实例及全部自身进程。无活动测试，无生产改动。下一动作：Daemon在restart/文件任务中途SIGKILL后的明确终态及不重复副作用，不能把当前稳定运行崩溃证据当作中途任务也已验；A05浏览器等待状态及其他范围继续保留。

2026-09-13 A03停止边界：新增真实双节点TestFailedStopBlocksReplacementButOtherNodeRemainsUsable，忽略TERM实例1秒停止失败、再启动失败且原runId/计数x不变，另一节点期间可启动停止；90516 race退出0（4.375s）。底层主PID先退/后代持管道、升级确认全后代退出、不升级有限失败三项41872 race通过（1.302s）。A03矩阵更新Linux真实覆盖，Windows/cgroup仍环境限制；无生产修改/活动测试。下一动作：A10独立Daemon进程SIGKILL后重启并识别原运行、活动任务结局，复用package-smoke自建进程管理但不要把Daemon.Close重建冒称崩溃。A05浏览器节点失联提示及其他范围仍待继续。

2026-09-13 A04响应丢失闭环：desktop.spec.ts新增真实restart202后route.abort丢浏览器响应，刷新保留确认并显式重试，原requestId/taskId、仅一个restart任务、实际runId一致；73971退出0，1/1（4.4s）。结合同key集成/管理断线/数据库重开/独立SIGKILL/A15资源串行证据，A04矩阵更新。finally实际STOPPED，fixture1671058723会话93296已发送Ctrl+C，收取退出。无生产修改。下一动作：A03完整异常后代/停止超时全局可用组合及A05浏览器节点断线等待状态、A10实际Daemon崩溃对账；同时E09最终重新打包/全套回归仍未完成。

2026-09-13 A04独立Master SIGKILL：scripts/package-smoke.py新增--master-crash，97578退出0；解包development-20260913归档、全新/tmp/blora-release-smoke-wkzbiotd，真实Master子进程在restart RUNNING时kill并确认-9，原目录启动后同key复用任务，成功后计数xx，完成重试不改runId。finally实际停止有限测试实例并停止所有自身进程，无活动会话。生产代码未改。下一具体动作：补客户端已经接受但响应真正丢失的重启重试验证（当前明确测试管理断线与Master硬中断，不冒称浏览器响应丢失）；再审计A04全部条件并推进A03/A05/A10。全范围仍未完成。

2026-09-13 A04服务数据库重开：新增TestRestartTaskSurvivesMasterAndDatabaseReopen，实际双Daemon+真实进程，restart RUNNING时关闭Master服务和SQLite后重开，同key复用任务，原任务成功；启动标记从x到xx，第二次重开已完成重试仍xx且runId一致。69010 race测试通过（8.12s），已收取最终状态，无活动测试。生产无改动。下一动作：补独立Master二进制进程SIGKILL中断而不是同进程服务对象重开，使用自建fixture/临时状态和启动计数标记，继续A04硬中断证据；其余完整范围仍未满足。

2026-09-13 A04/A05任务中断线：加强TestTwoNodesAuthorizationAndDurableLifecycle，restart任务RUNNING时关闭真实Daemon管理连接，等新generation后同key重试仍原taskId，原任务成功及runId在后续重连不变、最终停止；6548 race退出0（7.392s），无活动测试。未将Master RUNNING边界称为精确Daemon停止调用断点，生产无改动。下一动作：实现/验证Master进程中途重启后的同requestId对账入口，优先检查已有可重开store/atomic handler fixture与任务恢复逻辑，再补A04直接证据；其他完整范围仍待继续。

2026-09-13 A02独立视图闭环：新增真实Monaco200行共享模型双窗口，键盘首尾定位、IDB同draftId不同cursor/viewState、刷新后首行精确1/末尾200与原光标、继续输入落在原位置且另一窗口滚动不变。最终81566退出0，1/1（12.7s）；A02结合真实双账号冲突/迟到响应证据更新已实现。fixture3227456034会话87506已发送Ctrl+C，收取退出；无生产改动。下一动作：重新核对A01最近输入/后台窗口、A03/A04/A05/A10运行故障和A13/A14/A16旧矩阵证据，将缺少的真实组合逐项补齐；全范围及长期内存仍未完成。

2026-09-13 A02迟到响应增量：加强实际shared file windows，route.fetch拿到真实Master响应后gate暂缓交付，期间追加中文，释放后只出现重新读取确认而不改正文，两窗口/刷新都保留最新输入。33078退出0，1/1（9.6s），前序双账号冲突一并通过。fixture136228777会话80655已发送Ctrl+C，收取退出；无生产改动。下一具体动作：补共享Monaco模型的两窗口独立光标/滚动位置及刷新恢复，以真实键盘/滚动和持久editorView观测，不通过测试专用模型实现代替。A02仍进行中，长期内存与其他范围仍保留。

2026-09-13 A02真实外部修改增量：files-terminal.spec.ts新增双编辑窗口共享未保存中文、另一成员账号真实修改节点文件、旧基线冲突/服务器新版本保留、刷新保留两视图正文与比较版本、显式采用基线后保存。97420退出0，1/1（7.7s），fixture446296949会话59363已发送Ctrl+C，收取最终退出；生产代码未改。下一动作：补同模型两视图独立光标/滚动的实际浏览器断言，以及迟到服务器读取响应与新输入竞态，A02尚未全项完成。见shared-editor-conflict报告。

2026-09-13 A08文件与账号闭环：新增文件查看/写入/实例查看撤权真实浏览器，66268退出0，1/1（6.7s），保存拒绝/原服务器正文保留，刷新本地草稿undo redo/导出完整，新读取403及实例列表过滤；84359退出账号隔离复验1/1（3.5s）。结合前述实际PTY撤权，A08矩阵更新。测试恢复fixture原授权，fixture708836550会话22573已发送Ctrl+C，收取最终退出；无生产修改。下一具体动作：A02同用户双窗口同文件与另一账号外部修改的完整共享正文/独立光标滚动/版本冲突组合，检查现有editor-copy/settings-editor测试后补真实缺口。长期内存及其他A/E项仍未完成。

2026-09-13 A08活跃PTY：新增真实双账号独立上下文撤权浏览器测试，81846退出0，1/1（6.1s）；成员原生PTY单次文件追加，撤input后租约false/再输入无副作用，撤read后新请求403，刷新旧快照显示权限错误且不能输入。临时host.manage等授权finally撤销，原会话关闭；fixture1366325914会话41642已发送Ctrl+C，收取退出。无生产修改。下一动作：核对A08资源查看撤权期间编辑器草稿保留/账号切换证据，再推进其余仍进行中场景；长期内存采样仍未完成。见browser-revoke报告。

2026-09-13 A09慢端期限：新增TestLogSlowConsumerDeadlineAndCursorReattach，保持真实45秒期限，实际WSS读取不返信用，45.03165s收到SLOW_CONSUMER而非仅断线，原runId/cursor1重挂载收到后续第二标识。51543 race退出0（47.329s），自建实例Cleanup停止，无活动测试。生产Send确认等待socket写入，未发现错误帧仅排队就关闭的问题，无生产改动。下一具体动作：量化长时间混合流全进程内存，或按现有报告推进仍缺直接证据的A08运行中撤权浏览器组合；不要将协议错误送达自动升级为浏览器可见提示已测。证据在mixed-stream-control-2026-09-13.md。

2026-09-13 A09连接预算增量：混合流测试新增20次/50ms Conn.Stats采样，未消费<=256KiB、接收队列<=2MiB/64帧；最终29591 race退出0（15.959s），峰值82590B/79302B，批量传输与5秒停止也通过。7241首次错误假设填满信用窗口，实际日志按批次ACK更早背压；同时暴露失败时临时目录与无限日志进程清理竞态，已将显式stop注册到本测试Cleanup，最终复验通过。生产代码未改，无活动测试。下一动作：覆盖45秒慢消费者期限与可见SLOW_CONSUMER错误、重挂载游标恢复，并记录较长运行内存采样；目前短采样不等于全范围A09完成。见混合流报告。

2026-09-13 A09归档滚动增量：新增真实PTY无读取端64MiB输出、64KiB预算/16KiB分段测试，10470 race退出0（3.321s）；实际2.300s，380次采样峰值65358B，磁盘63317B与计数一致，迟到读Gap及末尾标识通过。见[混合流报告](../acceptance/reports/mixed-stream-control-2026-09-13.md)。无活动测试/fixture。下一具体动作：给真实混合流测试采集Conn.Stats的队列/信用峰值并覆盖慢端饱和，结合持续运行进程内存观测补A09，不把短时64MiB归档测试称为长期稳定性。其余未满足验收继续保留。

2026-09-13 A09真实混合流增量：加强logs_integration_test.go慢日志测试，4097B循环输出、不归还首帧信用、同源节点8.4MB跨节点传输确认进度后发stop，5秒内真实STOPPED，完成归档可读、传输Verified且目标全字节一致。69124定向race退出0（14.120s）；57388首次新增字段名编译错误已修正复验。生产代码未改，无活动测试/fixture。见[混合流报告](../acceptance/reports/mixed-stream-control-2026-09-13.md)。下一具体动作：补A09长期有界队列/归档滚动与慢端可见背压的实测，核对现有ArchiveRetention测试和协议队列指标入口；同时保留A01/A02/A03/A04/A05/A08/A10/A13/A14/A16及E项未完成组合，不能因新子项通过结束全范围任务。

2026-09-13 A15多视图闭环：`desktop.spec.ts` 新增双中心跨节点多标签/重复资源、独立导航、同浏览器任务提交两启动确认；最终31449退出0，1/1（7.9s），仅一个成功、失败限定旧运行未退出、结果runId一致，自动状态同步及刷新导航保持。早期36852被窗口遮挡桌面图标、59146匹配隐藏导航均已修正测试定位。实例最后STOPPED，fixture4012957923会话11603已发送Ctrl+C；收取最终状态。A15矩阵已更新，生产代码未改。下一具体动作：重新审计A01～A17剩余进行中行与最新专项报告，按明确缺口补真实故障组合，尤其A08/A09断线慢流和A03/A04运行归属；不要以当前几个绿色场景宣布全量完成。

2026-09-13 A17活跃终端闭环：新增actual PTY pointer真实鼠标拖出/合并立即刷新，原视图/会话/标签顺序保留，WSS Data输入不增加，真实文件副作用一次，最终99231退出0，1/1（8.6s）。首轮59340因空终端容器同时匹配失败，修正定位后通过。A17矩阵更新为已实现、Linux真实组合通过；生产代码未改。fixture1635242649会话4227已发送Ctrl+C，收取最终退出后继续A15。下一具体动作：检查InstancesApp资源标签开窗/导航/控制接口，新增两个实例中心、跨节点标签、同实例重复视图与并发控制的实际浏览器组合。全范围仍未完成。

2026-09-13 A17真实鼠标增量：`desktop.spec.ts` 新增双编辑标签实际pointer拖出/合并，每次松开立即刷新，原ID唯一/原窗口归属/标签顺序/中文正文/undo redo/另一正文独立均通过；86685退出0，1/1（3.8s）。首轮54242失败为隐藏Monaco也被定位的strict mode，修正可见选择器后通过，生产代码未改。fixture3185892968会话46036已Ctrl+C退出0，无活动测试；见[拖拽报告](../acceptance/reports/drag-recovery-2026-09-13.md)。下一动作：补实际PTY拖出/合并立即刷新与WSS输入不重放，再实现A15真实跨节点双实例中心多视图并发控制组合，不能以菜单移窗旧证据代替新组合。

2026-09-13 A11事务中断闭环：真实Chromium在snapshots写入后、pointers请求时abort事务，数据库集合保持旧值，最新正文导出/刷新日志恢复通过；首轮82235发现未捕获tx.done AbortError，已修复RecoveryService立即观察done并在失败时abort/等待，最终75942为3/3（6.7s），79602前端生产构建及46/46单元通过。A11矩阵已更新实际浏览器覆盖，物理掉电不在本证据中。fixture2619400585会话72052已Ctrl+C退出0，无活动测试。下一条具体动作：A15/A17真实跨节点实例多窗口共享状态、独立视图、标签拖出/合并立即刷新组合；先核对现有桌面/实例实现和真实测试避免重复。全范围仍未完成。

2026-09-13 A11浏览器故障增量：新增 `recovery-failure.spec.ts`；实际HTTPS/Monaco中Storage同步日志与IDB snapshots.put同时QuotaExceededError，最新正文仍可真实下载导出，恢复写入后刷新保留；schema0恢复与schema999原始记录两次刷新可导出均通过。会话37838退出0，2/2（4.6s）；fixture1827070879会话73966已Ctrl+C退出0，无活动测试。见[恢复报告](../acceptance/reports/recovery-quota-2026-09-13.md)。下一条具体动作：补IDB事务在快照写入后/指针切换前中断的浏览器验证，确认当前/旧快照不混合，再继续A15/A17真实跨节点多窗口移窗立即刷新；全范围尚未完成。

2026-09-13 A06 实际程序增量：`files-terminal.spec.ts` 新增 vim 未保存中文/刷新/真实保存及 top 刷新/退出后真实写文件，最终会话65248退出0，1/1（9.0s）。首轮97839因对原始增量ANSI日志断言连续整行失败，修正测试观测方式后通过，生产代码无改动。fixture4180777667会话36955已Ctrl+C停止；独立发行包smoke39175也退出0且进程停止。见[终端报告](../acceptance/reports/terminal-programs-2026-09-13.md)。下一条具体动作：检查并补A11真实浏览器存储配额故障/恢复迁移验证，再推进A15/A17跨节点多窗口移窗立即刷新组合；全范围仍未完成。

2026-09-13 解包独立链路：新增 `scripts/package-smoke.py`，会话39175退出0；临时目录 `/tmp/blora-release-smoke-m3u7kpav`，SDK/参考扩展离线干净依赖安装、源码构建、归档签名工具运行、Master初始化/TLS/HTML/登录及两个解包Daemon ONLINE均通过，三个进程已停止。下一步正在真实vim/top刷新恢复，fixture `.local/fixture-4180777667` 监听9444，会话36955；浏览器完成后必须Ctrl+C停止fixture。报告见发行包报告，Windows真机仍缺失。

2026-09-13 本地发行包检查点：`scripts/package.py` 与 `make package` 已接入，六个归档生成于 `dist/releases/development-20260913`。会话45917构建退出0，5282重复打包及独立清单/摘要校验退出0；Master每平台183文件、Daemon78、SDK21、web115。同输入归档逐字节一致，版本不同内容拒绝覆盖，Windows仍仅交叉编译。README与[发行包报告](../acceptance/reports/package-2026-09-13.md)已补充；这些后续文档不在旧归档里，最终需新版本再打包。无活动测试或fixture。下一条具体动作：将这组Linux Master/Daemon和SDK解包至本任务临时目录，验证独立初始化/健康检查及SDK离开仓库重新安装依赖和构建，然后继续A06实际vim/top、A11浏览器配额、A15/A17真实组合；全范围未完成。

2026-09-13 A07最终通过：TasksApp接回原本机上传，复核账号/任务/动作/requestId/资源及原文件身份；前端46/46单元、类型/生产构建通过。真实关闭标签后万条文件解压完成，16MiB上传WAITING_CLIENT；新browserTabId从983,040B原偏移复用原taskId续传成功，目标摘要正确，约1.1分钟通过。fixture-3124651407已停止，无活动会话。A07标已实现并明确Linux通过，Windows留E09。最终storage版本保护后的check/build/windows也已通过。下一具体动作：产出独立Master/Daemon与SDK发行包及校验清单，再按矩阵补A06实际全屏程序、A11浏览器配额、A15/A17等真实组合；仍不称全量完成。见[上传关闭证据](../acceptance/reports/upload-close-2026-09-13.md)。

2026-09-13 A07继续：发现新浏览器标签丢失本地上传入口而服务端检查点仍在，新增TasksApp“继续本机上传”及restoreUploadFromTask，GET再次核对当前账号/原任务/资源/请求身份和文件指纹，重建本地引用后复用原幂等请求；不创建新上传任务。3项定向单元通过（430ms），前端生产构建及类型检查通过；最后追加action/requestId核对尚待最终复验。真实upload-close.spec已编写：万条目录压缩→解压与16MiB本机上传并行→关闭实际标签→新标签从任务中心选择原文件续传。当前fixture-3124651407监督会话70696，浏览器测试20890；需接收结果。上轮最终make check/build/windows已通过，无其他构建会话。下一动作：修复真实A07失败并记录结果，随后发行包与其余完整场景。

2026-09-13 512MiB实际传输通过：fixture-4036895406同一任务659de8466bcbe9ef91aed784ce047b90完成，543.324s（约0.942MiB/s），源目标独立SHA256一致。首轮小文件观察预算不足手动中断，后台未取消；新测试验证资源/源版本后续接原任务，约5分钟通过，p95 42.4ms、2/357长帧、未确认峰值87,404B，8窗口/2PTY/万条目录/512MiB传输全程采样并行。filesystem完整race1.300s；fixture已停止，无活动浏览器。另补storage未知/不连续迁移拒绝启动，完整存储race17.216s通过，新增运维手册。最终check/build/windows会话52400待接收。下一动作：记录最终构建结果，然后按矩阵继续A07并行解压/本机上传关闭网页场景，以及发行包与长期/平台验证；不要把现有大量旧矩阵行当作已全覆盖。见[传输证据](../acceptance/reports/transfer-proof-2026-09-13.md)、[升级保护](../acceptance/reports/upgrade-storage-2026-09-13.md)。

2026-09-13 大文件优化执行中：新增filesystem/read_proof.go的64KiB块SHA索引（每服务64条/8MiB上限），原ReadChunk不变；Daemon新增file.transfer.stat/chunk，Master中间读取使用索引并在发布前完整核对源版本。文件定向race1.045s、原传输TLS回归46.307s、新增“已读前缀改写并恢复mtime不能发布”与实际源Daemon重启重建索引25.963s通过。fixture增加performance-transfer-mib（1..1024）。当前无fixture，make check/build/windows和fixture二进制正在构建；下一动作：接收构建结果，启动--performance --performance-transfer-mib 512的私有fixture，运行真实performance.spec记录吞吐和本地响应，再补完整filesystem回归。见[校验索引报告](../acceptance/reports/transfer-proof-2026-09-13.md)。

2026-09-13 E08短时真实场景最终通过：识别SwiftShader软件WebGL并降级默认终端渲染器；不透明窗口和拖动轻效果、64KiB/128条输出增量及大检查点持久化确认已接入。最终fixture-746611301串行PTY17.9s+性能35.7s，2/2通过（54.7s）；p95 38.8ms/max48.5ms，0/302长帧，8窗口/2PTY/万条目录/16MiB传输全程并行，未确认峰值87,353B、目标校验成功。浏览器大屏幕恢复/降级2/2（13.2s），前端43单元及构建通过。make check/windows最终通过；所有fixture已停止，无活动会话。下一具体动作：ReadChunk每块两次整文件哈希及Master每8块再次Stat导致512MiB吞吐极低；设计不削弱版本/源变更/目的校验合同的有界读取优化并做真实大文件复验，之后继续长期E08及其他未验证组合。完整证据见[性能报告](../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 E08绘制诊断续接：CPU诊断主要落在原生program；窗口整体不透明后p95降至132.4ms但仍失败。增加硬件WebGL能力探测（实测仍选WebGL，降级未验证）及未确认字节峰值观测，下一样本169.1ms、峰值87,310B，仍失败。全部前端单元43/43通过；128条/64KiB终端增量上限、类型和生产构建通过。新改动拖动期间轻阴影/关闭背景模糊，正在fixture-1196973665（监督会话67716）运行性能测试39784；需接收结果及图形设备字段。之前4028490118/719092740/2394455507均已停止。下一动作：判定绘制改动效果、验证渲染降级分支及最终PTY回归，再继续大文件哈希性能与E08未达标项。

2026-09-13 E08第二次优化：终端恢复增加64KiB输出增量日志，完整屏幕低频合并，旧记录兼容；浏览器7.7s、真实PTY19.4s、生产构建/类型检查通过。混合负载42.5s仍失败：p95 219.6ms、272/288长帧，heap降至43.9MB；双PTY、万条目录及16MiB传输全程并行已确认。fixture-1188777270已停止，新增可选BLORA_PERF_PROFILE只汇总登录后CPU函数时间。新诊断fixture会话47116正在启动；下一动作：读取其凭据路径，带BLORA_PERF_PROFILE=1运行performance.spec，区分脚本与合成绘制瓶颈后继续优化。见[更新证据](../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 E08实测未通过：真实8窗口、2PTY、万条目录和16MiB跨节点传输已运行；反馈p95 278.2ms，303/319长帧，传输验证成功但先于采样结束，不满足全程并行。RecoveryService合并在途重复快照，恢复测试9/9、前端构建及PTY/协议定向race通过；终端逐事件完整屏幕序列化仍需优化，512MiB反复全源哈希导致吞吐严重偏低。所有本轮测试结束，fixture-2336190821/3188205192已Ctrl+C停止，无活动会话。下一具体动作：终端恢复输出增量化，保留立即刷新及不重发输入，复验真实PTY刷新后重新测量全程负载；继续大文件传输性能。见[真实性能失败证据](../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 原生进程树指标：Daemon按运行成员出生身份聚合CPU/RSS，最多1024成员，成员变化拒绝旧快照并最多三次重采样；Windows Job枚举补句柄归属和出生身份，界面显示范围/成员数及RSS共享页口径。真实忙子进程TLS聚合2.380s、最终daemon/monitor/Master指标race1.324s/1.270s/5.963s、check/build/windows及前端构建通过。无活动测试/fixture。见[进程树证据](../acceptance/reports/process-tree-metrics-2026-09-13.md)。下一具体动作：检查现有performance.spec与fixture，补E08真实两终端、万条目录、后台传输和八窗口并行测量；保留短命成员历史账本、Windows真机及第二账号登录等待问题边界。

2026-09-13 fixture清理/原生CPU：编译监督二进制直接运行fixture-404818613，启动实例后Ctrl+C退出0，数据库证明start及两次kill成功、两实例STOPPED；make fixture改为直接执行编译二进制。Linux真实CPU双次采样、出生身份与RunID基线已实现，主机统计不重复guest，CPU基准在API/界面区分。monitor/Master定向race1.269s/4.667s、最终check/build/windows和前端构建通过。无活动会话/fixture。下一动作：原生实例所属进程树CPU/RSS聚合与成员变化边界（当前是主进程），随后E08实际并行负载及既有登录超时诊断。见[CPU/清理证据](../acceptance/reports/native-cpu-2026-09-13.md)，不把主进程采样称为整个运行总量。

2026-09-13 网络/可见性收尾：真实Docker浏览器最终1/1（29.9s）验证网络值、七秒最小化无原五秒轮询及恢复即时采样；修正单窗口直接恢复/多窗口选择器测试。第二账号界面登录曾超时，专项使用真实API登录，问题保留。最终构建和指标定向race通过。fixture-75425925/515580840/2134529472均停止；四个失败轮次容器按记录RunID核对后停止/删除，成功轮次任务已清理，无活动测试会话。Linux fixture子服务新增Setsid避免Ctrl+C过早停止服务，尚需验证信号清理。下一具体动作：先实测新fixture信号清理，随后补Linux原生进程CPU采样（当前明确不可用）；继续E08完整负载和登录超时诊断。见[可见性报告](../acceptance/reports/monitor-visibility-2026-09-13.md)。

2026-09-13 监控网络/可见性执行中：节点与实例显示网络收发累计MiB，缺失或不可用不显示零；Point保留真实零网络计数，原生进程未采集的网络/磁盘标不可用。MonitorApp接visible，后台指标30s/历史60s，恢复立即读取对应范围；实例嵌入传递visible。发现Linux原生进程CPU未采集且先前默认为0，先改为不可用，下一项补真实采样。check/build/windows、前端构建、monitor/runtime定向race1.208s/1.189s通过。首次浏览器最小化七秒无请求成立，恢复步骤因测试错误任务栏定位而失败；旧fixture-75425925已停止并清理。最终fixture-515580840运行会话26367，浏览器复验会话34435仍待收取，需先完成验证并停止fixture再继续。

2026-09-13 监控代次缓存修复：采样后重读实例和授权，拒绝变化或不匹配RunID；缓存按当前代次选取，不跨代次回退。真实TLS原生实例断开/重建Daemon链路、同代次旧值时间戳、断链撤权403、重连实时恢复、实例重启后无新采样503均通过；最终监控定向race5.043s、check/build/windows、OpenAPI103路径通过。无活动测试/fixture。见[监控证据](../acceptance/reports/container-metrics-2026-09-13.md)。下一具体动作：按F10/E08核对节点/实例网络指标的界面呈现，以及最小化窗口的绘制/采样行为和真实负载证据；继续全范围验收，不将Daemon.Close测试等同物理断网或Windows真机。

2026-09-13 F10真实指标端到端：实例监控标签接MonitorApp、实例范围不发主机请求；修复全新节点首次容器启动缺少InstanceRoot初始化。真实双Daemon/Docker浏览器1/1（15.4s）验证授予/撤销权限、CPU/128MiB内存、刷新、停止后stale和无主机请求；全新根目录真实Docker2.950s、runtime/daemon race2.982s/30.584s、check/build/windows、前端构建及41/41通过。旧fixture-1599939518与最终fixture-3062496594均已停止，自有运行资源清理，无活动会话。失败与修复详见[容器指标证据](../acceptance/reports/container-metrics-2026-09-13.md)。下一动作：监控跨运行代次缓存及实际节点断开/重连的语义和可见诊断，随后继续全范围缺口；不把实例停止后的缓存当节点失联。

2026-09-13 F10容器实例指标增量：发现并替换Daemon固定不可用分支，运行适配器采集Docker双样本CPU、内存usage/limit和有界网络计数，采样前后核对归属与StartedAt；界面显示真实内存口径。HTTP身份/重启/计数边界race最终1.040s、真实隔离Docker生命周期含指标2.747s、最终check/build/windows和前端构建退出0。无活动测试/fixture。下一条具体动作：真实Master/双Daemon fixture创建隔离实例，验证监控API权限、浏览器CPU/内存口径和历史/失联标记；然后继续全范围缺口审查。见[容器指标报告](../acceptance/reports/container-metrics-2026-09-13.md)，F10/E01不计整项完成。

2026-09-13 Linux Compose CLI恢复完成增量：启动前持久令牌、启动后会话号，运行结束/启动恢复按身份与pidfd清理，未确认保留证据并阻止新Engine变更；Daemon连接前恢复门禁已接。中断后核对追加测试已落盘、无遗留测试命令，最终containers/daemon race7.900s/30.518s、check/build/windows退出0；真实Docker生命周期27.047s通过。无活动会话/fixture。见[CLI恢复证据](../acceptance/reports/compose-cli-recovery-2026-09-13.md)。下一动作：核对F10容器模式实例指标是否完整接入真实Docker统计，补明确代码缺口；其余SDK按规划公共能力语义审查，不凭旧“其余能力”笼统描述无限扩张。完整Daemon硬崩溃/平台及故障组合仍继续。

2026-09-12 Compose检查点硬崩溃/浏览器复验：Linux独立执行器SIGKILL后日志逐字节保留、查询INTERRUPTED/Unknown、相同任务禁止重放，race最终1.381s退出0。实际Docker浏览器输出/退出标识及刷新1/1（45.9s）通过；fixture-4124677712已Ctrl+C停止，测试部署及卷清理，无活动测试。OpenAPI3.1解析103路径。下一条具体动作：补Linux Compose CLI运行身份持久记录和Daemon启动遗留进程恢复，防止旧CLI仍执行时新任务并行修改同一Engine；当前测试按专属随机令牌清理不能替代产品恢复。继续其余SDK能力和全范围平台/故障验收，见[输出报告](../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose运行中输出检查点：每250ms仅保存变化、每流32KiB与截断标识；命令启动前记录未完成，退出后completed，界面不把缺失退出记录误称仍在运行。落盘失败取消并等待CLI，保留Unknown。受控真实CLI/跨管理器磁盘读取及取消race2.470s、containers全包6.520s、追加目录故障定向race1.952s通过；make check/build/windows、前端构建退出0。无活动会话或fixture。下一动作：Daemon/执行器硬崩溃的日志保留与遗留CLI处理审查、真实浏览器复验，然后其余SDK资源能力与全范围平台/故障验收。详情见[输出报告](../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose输出最终构建收取：会话9398前端生产构建退出0，源码/产物包含最终说明文案。无活动测试、fixture或测试资源待清理。下一动作保持为运行中有界输出检查点与崩溃窗口验证，随后继续完整范围；详见 [Compose输出报告](../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose阶段输出闭环：成功/失败stdout/stderr前32KiB、独立截断标识、最多16阶段写入有8MiB保护的操作日志；operation-output单阶段base64查询保持96KiB，Master/Daemon主机权限与负页码校验，任务详情分页/刷新接入。实际CLI进程双满流/中文/NUL/磁盘重开定向race2.120s、containers全包6.004s、Master TLS1.917s，真实Docker/Compose浏览器1/1（45.9s）、前端41/41、check/build/windows与OpenAPI103路径通过，见 [输出证据](../acceptance/reports/compose-output-2026-09-12.md)。fixture-2005453535已Ctrl+C停止，测试部署及卷已清理；最终文案前端重建尚待收取，此外无活动测试/fixture。下一动作：运行期间有界流式输出与崩溃前检查点，保持命令结果不明语义；然后其余SDK资源能力与全范围平台/故障验收。

2026-09-12 拉取EOF/末条观测闭环：正常EOF刷新最后一条采样，32MiB+1探针辨别超限，不继续旧镜像检查并保留Unknown。真实HTTP协议替身32MiB边界/迟到错误、末条计数与持久化定向race3.743s，最终check/build/windows通过，见 [容器进度证据](../acceptance/reports/container-progress-2026-09-12.md)。无活动测试或fixture。下一动作：Compose阶段成功/失败stdout、stderr有界持久记录与截断标识、受host.manage保护的读取/浏览入口；查询保持96KiB边界，不把整个操作输出塞入控制消息。随后继续流式诊断、真实CLI/Engine与全范围SDK/平台/故障验收。

2026-09-12 容器进度回归收取：会话42746的containers全包race退出0（2.553s），真实Docker条件入口未启用，不将跳过计作真实拉取。无活动测试/fixture。下一动作仍为镜像采样末条计数与限流EOF边界，随后Compose输出持久化/受控读取及完整任务诊断；见 [进度报告](../acceptance/reports/container-progress-2026-09-12.md)。全范围目标不变。

2026-09-12 镜像进度增量：修复Engine progressDetail被丢弃，emitProgress/节点日志/Daemon任务修订保留Layer与Current/Total，任务界面显示分层计数。协议替身首次漏/version协商失败已修正，定向race1.024s、check/build/windows、前端生产构建通过，见 [容器进度报告](../acceptance/reports/container-progress-2026-09-12.md)。正在容器全包race回归，无活动fixture或生产操作。下一动作：收取全包结果，修复采样末条进度保留与流量上限EOF语义，再补Compose实际输出的有界持久记录/浏览入口及真实验证；全范围继续。

2026-09-12 任务阶段详情闭环：发起者/时间/含等待耗时、终态冻结与缺失时间边界、专用详情和持久阶段分页接入。真实TLS57修订/诊断限长/负载排除/撤权race1.973s，前端41/41，真实文件任务详情浏览器1/1（3.8s），check/build/windows及OpenAPI3.1/103路径通过，见 [阶段记录证据](../acceptance/reports/task-stages-2026-09-12.md)。fixture-186544590会话65750已Ctrl+C停止，无活动测试或fixture。下一动作：审查各任务执行器的实际进度与输出记录，优先补Docker镜像拉取/Compose等长任务的有界持久诊断与实时阶段，继续其余SDK受控资源能力及全范围平台/故障验收。阶段记录不等于程序stdout，不能因此提升F09整项通过。

2026-09-12 任务阶段记录进行中：TaskStagePage投影持久task_events，按任务修订号倒序分页，phase/error限512/4096字符并标截断，不返回payload/result；GET /tasks/{id}/events复核当前canTask。任务列表增加详情入口，专用详情显示发起者、接受/派发/更新时间、含等待耗时和阶段页；终态耗时冻结、无效时间明确不可用。真实TLS57修订跨页/诊断/撤权race1.973s、时间边界2/2、check/build/windows与前端生产构建通过。fixture-186544590会话65750活动，正在真实文件任务详情浏览器验证；前端全单元与OpenAPI结果待收取。下一动作收取、修复失败、补报告；阶段记录只代表真实任务状态历史，不冒充执行程序stdout。

2026-09-12 任务统计/筛选闭环：定向TLS race3.499s、真实101任务/后端状态筛选/刷新/任务栏一致/通知详情组合2/2（9.8s）、前端39/39、生产构建、make check/build/windows、OpenAPI3.1/102路径通过，见 [任务历史报告](../acceptance/reports/task-history-2026-09-12.md)。WAITING_NODE测试样本自动派发造成的两次失败已保留；WAITING_CLIENT稳定样本复验通过。fixture-4241791769会话99611已Ctrl+C停止，无活动测试或fixture。下一动作：默认任务显示actorId、创建/更新时间与实际耗时；通过已持久task_events提供获准任务阶段记录分页和真实详情入口，不把阶段日志伪称进程stdout；然后继续完整SDK资源能力、平台/故障及其余验收缺口。全范围仍进行中，系统通知真桌面及Windows真机未验证。

2026-09-12 任务统计/全历史筛选进行中：GET /tasks/summary按500条活动根任务页扫描全部当前授权记录，5秒超时返回错误不返回部分总数；任务栏读取该API且错误显示待确认。before历史模式增加服务端state条件，UI筛选重置游标并恢复筛选。check/build/windows、前端生产构建、OpenAPI3.1/102路径通过。新增TLS测试前两次因WAITING_NODE样本被真实派发器改变状态失败（管理员计数1非2），改WAITING_CLIENT稳定样本复验中会话97929；新fixture会话99611待就绪。下一动作收取定向race与fixture路径，运行notifications.spec.ts含101真实任务/筛选刷新/完整任务栏计数，再更新证据。无生产操作；整项范围不缩减。

2026-09-12 任务历史闭环：最终storage游标race12.905s、Master授权1.883s、真实101任务分页/刷新与通知详情组合2/2（10.9s）、前端39/39、check/build/windows通过，见 [历史报告](../acceptance/reports/task-history-2026-09-12.md)。已记录模板语法、通知定位和测试异步队列期限三个失败/修复；fixture-501473755会话67184已Ctrl+C停止，无活动测试/fixture。下一动作：任务栏完整获准未终态统计（目前旧1000条列表会漏旧未完成任务）、任务状态筛选传到历史查询；随后补默认任务元信息/日志及其余全范围缺口。上一全仓race通过，当前增量定向通过，不将Windows构建称真机验收。

2026-09-12 任务历史进行中：RootTaskPage稳定before游标、权限过滤、真实界面分页与现场、专用详情直接GET接入；1005任务插入稳定性定向race最终12.905s，Master分页授权1.883s通过。101个真实mkdir任务浏览器分页/立即刷新通过；同套件后续通知测试因前一测试尚在完成的同类任务产生40条通知而定位歧义失败，现通知显示资源身份，测试按本次taskId精确定位（不取第一条）。最终前端与两端重建进行中；fixture-501473755会话67184仍活动。上一全仓race78624已退出0，Master175.982s/extensions164.323s通过；本历史分页在全仓之后修改，单独证据不冒充全仓覆盖。下一动作收取构建并复验真实两场景，更新报告/矩阵；其余任务统计/迁移与全范围缺口继续。

2026-09-12 通知最终组合收取：修复初始化竞态后3/3（9.8s）通过；含真实元数据存储故障禁止PATCH/显式重试、独立通知来源和刷新、真实文件任务完成通知。fixture-1098943595会话52328已Ctrl+C停止，无活动fixture。全仓race78624仍运行，extensions164.323s通过，其余已输出包通过；待收取Master。详见 [通知证据](../acceptance/reports/notifications-2026-09-12.md)。下一动作：任务中心历史读取突破RootTasks固定最近1000条，加入稳定游标与权限过滤，专用任务详情直接查询；实现真实界面分页/刷新后再验证。系统通知真桌面展示、全范围平台与故障组合仍保留。

2026-09-12 通知组合失败修复：首轮三场景2/3通过（含注入恢复日志失败时禁止元数据PATCH、恢复后同请求显式重试）；独立样例输入在初始化readData完成前可用，随后被恢复覆盖导致笔记丢失。现将初始输入/按钮禁用到恢复和监听绑定结束，三个包重建通过；通知测试用受控未完成数据请求确定性覆盖禁用时机。最终三场景复验进行中，fixture-1098943595会话52328仍活动，全仓race会话78624仍运行。新增通知日志错误单元2/2通过。无生产操作；继续收取并修复失败，不能写成全量完成。

2026-09-12 任务通知进行中：WSS通知模式、首次当前快照、续传游标、服务端授权/内部子任务过滤/终态最小摘要接入；默认桌面持续订阅，通知和游标同日志提交，显式系统通知启用/关闭及点击详情代码接入。真实WSS游标/权限/race2.321s、真实文件任务浏览器1/1（3.6s）、前端38/38、check/build/windows通过。发现RecoveryService.commit失败仅记录状态，扩展restore/通知桥接现已检查protected并拒绝成功回执；最终前端构建通过。正在fixture-37470389后继fixture会话52328验证本地日志故障时禁止元数据PATCH与恢复后显式重试；旧fixture均已停止。下一动作收取新fixture私有路径/定向浏览器组合与新增错误测试，更新证据，然后全仓race及其余F/A/E缺口。系统通知在真实桌面操作系统的展示仍未验证。

2026-09-12 通知子项最终构建收取：会话77057的make check/build/windows退出0，无活动测试或fixture。后台任务通知下一步复用已有持久task_events和WSS/Protobuf事件通道，增加通知订阅的起始快照/持久游标，避免只轮询最近1000任务漏掉快速完成与旧任务；通知及游标同一恢复日志提交，刷新/断线不重放已读历史。尚未开始该代码，上一通知报告中的范围限制仍有效。

2026-09-12 通知SDK子项：notification.publish/host.notify、服务端实时能力/app.use校验、工作区持久通知托盘与来源跳转接入。独立包浏览器1/1（4.7s）、最终TLS通知/云端策略race2.195s、前端37/37及构建、SDK三个包、OpenAPI101路径通过，见 [通知报告](../acceptance/reports/notifications-2026-09-12.md)。fixture-3796281576与早前fixture-115163169均已Ctrl+C停止；最终make check/build/windows会话77057待收取。下一动作：接入后台任务终态通知（关闭任务窗口仍工作），按已授权真实任务变化触发，持久去重、点击跳任务详情；系统通知必须用户显式启用且不重放历史完成项。其他全范围缺口继续保留。

2026-09-12 独立扩展元数据浏览器闭环：实例窗口、名称草稿、稳定请求现场与显式重试接入，提交前等待本地现场持久化。真实双节点浏览器1/1（4.7s）、独立包Master定向race30.708s通过，SDK与三个版本样例构建通过，见 [元数据报告](../acceptance/reports/extensions-metadata-2026-09-12.md)。fixture-115163169 会话93883暂保留供后续验证，无生产操作；下一动作：通知SDK与带应用来源的站内通知能力，然后继续其余受控资源与全范围验收。

2026-09-12 SDK元数据写入增量：updateResource/resource.write与PATCH扩展资源API接入，实例name/group/tags、显式configRevision、事务幂等、app.use及instance.configure复核，输出不含执行配置。transport10/10、SDK与前端检查/构建、最终check/build/windows及OpenAPI解析通过。TLS新增测试的包序列化和启用路由错误已修正，最终race1.909s通过（含修订号缺失拒绝）。见 [元数据报告](../acceptance/reports/extensions-metadata-2026-09-12.md)。下一动作：独立扩展示例增加实例元数据编辑与稳定请求恢复，跑真实浏览器链路；随后通知SDK及其他受控资源能力。无活动测试、fixture或生产资源操作。

2026-09-12 计划任务分页贯通：Linux.timer完整排序、Windows COM完整路径Ordinal保留后续最小集合，limit/after经API/RPC传递，界面分页和立即刷新恢复。205项跨三页race1.036s、Master参数TLS1.963s、浏览器1/1（6.6s）、最终systeminfo全包race1.246s、check/build/windows、前端生产构建与OpenAPI解析通过，见 [系统报告](../acceptance/reports/system-adapters-2026-09-12.md)。Windows COM运行仍未验证。下一动作：复核剩余SDK受控资源写入/通知能力及生产任务语义，补真实安装示例，再继续全范围故障和平台验收。无活动测试、fixture或主机变更。

2026-09-12 服务分页：Linux/Windows名称游标、Daemon/Master传递与SystemApp首批/下一批/刷新恢复接入，解决只返回前100/200项。systeminfo race1.187s、浏览器1/1（6.2s）、check/build/windows、前端check/build及OpenAPI解析通过，见 [系统报告](../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：计划任务分页，Linux.timer和Windows完整COM路径分别排序并稳定游标；补服务器参数验证和浏览器恢复，然后继续全范围剩余项。无活动测试、fixture或主机变更。

2026-09-12 Windows服务操作增量：原生SCM启停与状态轮询取代sc.exe，重启等STOPPED后才启动、启动等RUNNING，整体30秒并检查取消；合法Unicode/空格服务身份与Linux命令参数边界已补。systeminfo race1.169s及最终make check/build/windows通过，真机启停/超时尚未验证，见 [系统报告](../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：补服务/计划任务完整分页（目前返回前100/200项）、视图查询现场及状态错误呈现，再继续全范围权限/扩展/故障组合。无活动测试、fixture或主机变更。

2026-09-12 Windows计划任务增量：固定PowerShell/Task Scheduler COM脚本替换schtasks本地化LIST，完整路径/数值状态/下次时间JSON；Enabled变更后回读确认。Unicode/空格身份与数据参数隔离、20秒/1MiB边界、实际依赖探测接入，systeminfo race1.200s、Windows测试程序交叉编译及最终make check/build/windows通过。真机COM脚本尚未运行，不能算真实Windows验收。见 [系统报告](../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：系统服务/计划任务完整分页及Windows服务操作合法名称/完成确认，随后继续权限/扩展/故障组合缺口。无活动测试、fixture或主机变更。

2026-09-12 系统适配增量：Windows服务查询改原生SCM最小权限枚举/状态，补只读真机测试入口；Linux定时器返回.timer身份，禁止通过计划任务入口启停.service并隔离命令参数。systeminfo全包race1.172s、Windows测试程序交叉编译及最终make check/build/windows通过，见 [系统报告](../acceptance/reports/system-adapters-2026-09-12.md)。Windows真机未验证。下一动作：Windows计划任务以结构化原生/COM数据替换schtasks本地化LIST解析，补合法Unicode/空格任务身份和操作验证；随后完善系统列表分页及其余全范围缺口。没有活动测试、fixture或主机变更。

2026-09-12 旧防火墙接口清理完成：删除旧ApplyFirewall/RollbackFirewall、全局reload及无用单份比较函数；无双快照的变更明确拒绝。原错误/取消/恢复失败/外部冲突测试迁移到Snapshot接口，最终systeminfo/Daemon定向race1.168s/29.595s通过（含独立命令进程），make check/build/windows通过。见 [快照证据](../acceptance/reports/firewall-snapshots-2026-09-12.md)。无活动测试/fixture，无主机规则操作。下一动作：Windows系统服务查询改用SCM或结构化原生数据，计划任务替换本地化文本解析；修复Linux定时器名称误取service列，补平台构建和测试入口，再继续全范围剩余项。

2026-09-12 Daemon命令进程证据：新增Linux临时firewall-cmd包装器/独立测试子进程，走真实Daemon Snapshot生产路径；过期摘要不落变更租约、双快照持久化、取消恢复与终态race通过（28.946s），make check通过。详见 [快照报告](../acceptance/reports/firewall-snapshots-2026-09-12.md)。这是命令进程链路，firewalld仍为测试替身，不计真实防火墙环境验收。无活动测试/fixture，未触碰主机规则。下一动作：迁移删除旧systeminfo ApplyFirewall/RollbackFirewall函数及对应旧测试，保留取消/补偿失败的验证含义；再推进Windows系统服务/计划任务真实适配和全范围剩余能力。

2026-09-12 双份预览/提交条件：UI展示区域及运行/永久差异，planHash绑定完整双快照与目标集合；Master要求摘要，Daemon变更前重新采集核对。systeminfo/Daemon定向race1.124s/1.913s、摘要绑定race1.032s、浏览器1/1（7.5s）、check/build/windows、前端生产构建和OpenAPI100路径解析通过。Master预览必填/期限/权限定向TLS也通过。见 [快照证据](../acceptance/reports/firewall-snapshots-2026-09-12.md)。无活动测试或fixture。下一动作：删除/迁移旧systeminfo Apply/Rollback兼容实现，补Daemon真实命令进程链路与快照冲突拒绝；随后继续Windows系统服务/计划任务等全范围缺口。

2026-09-12 双配置快照增量：新生产防火墙租约固定区域并持久运行/永久快照，分别应用和恢复、读回核对，不使用reload；部分进度可恢复，外部不同规则修改拒绝覆盖。定向race systeminfo1.216s/Daemon2.074s及check/build/windows通过，见 [快照证据](../acceptance/reports/firewall-snapshots-2026-09-12.md)。旧租约缺少永久快照改为保留诊断，不猜测恢复；最后保护改动的Daemon定向race及重建也通过。下一步：预览显示区域与双份差异、提交条件核对，清理旧兼容Apply/Rollback函数，补Daemon真实命令测试入口；再继续Windows系统适配。无活动测试、fixture或主机变更。

2026-09-12 预览/任务显示增量：端口预览与实际受管集合统一，系统界面按实际任务阶段显示排队/等待确认，确认绑定原节点并恢复草稿；最终定向浏览器1/1（6.3s）、systeminfo全包race1.074s、check/build/windows及前端生产构建通过。构建首次因新增按钮模板实体表达式解析失败，改为computed后退出0。见 [预览证据](../acceptance/reports/firewall-preview-2026-09-12.md)。下一动作仍是firewalld区域及运行/永久双快照，分别应用与恢复，移除reload对无关运行规则的影响；该后端缺口尚未修复。无活动测试、fixture或主机变更。

2026-09-12 回滚耐久顺序续接：增加rolled_back持久阶段，规则恢复后先记任务终态再清理；已恢复原规则不重复变更，应用错误保留prepared证据。任务更新故障注入验证恢复记录、启动失败与不重放规则；定向race退出0（1.618s），随后追加规则恢复后/回执前中断场景也通过。详见 [防火墙证据](../acceptance/reports/firewall-confirm-2026-09-12.md)。最终make check/build/windows退出0，无活动测试、fixture或宿主规则变更。下一动作：修复firewalld预览与apply使用不同规则集、运行/永久状态混用和reload影响其他规则的合同；仍须完整实现Windows系统服务/计划任务及其余范围。

2026-09-12 防火墙确认增量：成功终态先持久化、随后删除确认租约；失败保留恢复记录，启动失败不覆盖成普通中断。SQLite故障注入与确认/取消/超时/重启定向race退出0（1.532s），最终make check/build/windows退出0，未变更宿主防火墙。见 [确认恢复证据](../acceptance/reports/firewall-confirm-2026-09-12.md)。下一具体动作：继续回滚终态的耐久顺序及 firewalld运行/永久规则合同，再补Windows系统服务/计划任务解析；其余SDK受控能力仍保留。没有活动测试或fixture。

2026-09-12 最新检查点：扩展数据转换已移到注册表锁外，重获锁后比对包/数据/提交令牌并重新验证签名信任与依赖。单 Manager 同时一批迁移，冲突保留当前包和最新数据。`go test -race ./internal/extensions ./internal/master -run 'TestMigrationDoesNotBlock|TestUserData|TestRegistry|TestExtensionData|TestIndependentReference' -count=1` 退出0（extensions101.070s、Master36.693s）。迁移 API 仍同步，原全局授权阻塞缺口已修复。

同轮 Windows 原生监控/进程、Linux pidfd 安全终止及完整进程分页已接入，见 [监控证据](../acceptance/reports/monitoring-2026-09-12.md)。monitor race1.291s、最终真实TLS Master2.246s、浏览器确认/分页/恢复1/1（5.4s）、check/windows/build及前端生产构建通过；Windows测试程序交叉编译通过，真机未运行。首次新增TLS测试响应解码编译错误已修正复验。没有活动测试或fixture。下一动作：继续 Windows 系统服务/计划任务本地化解析、防火墙持久化边界及其余SDK受控能力；全范围仍未完成。

2026-09-12 数据迁移最新检查点：全仓 `GOCACHE=/tmp/blora-go-grant-idem make test` 退出0（Master194.482s、extensions144.750s），随后新增的数据进程中断1.396s及并发/配额2.661s定向race通过。make sdk默认沙箱EPERM后提权重跑通过，干净安装及三个样例包生成完成。真实迁移浏览器1/1（11.7s）、前端33/33、OpenAPI100路径、Linux/Windows构建均已有证据；当前正复建最后校验改动后的二进制，除此无活动测试/fixture。下一步Windows监控实际适配，不能把交叉编译当真机验收。详见 [数据迁移报告](../acceptance/reports/extensions-data-2026-09-12.md)。

2026-09-12 数据迁移故障证据：编译复用/双用户 WASI 迁移定向 race 138.095s、Master 数据授权 2.166s；实际子进程退出恢复（含数据、提交令牌）1.396s；并发 CAS 与近 4 MiB 配额文件失败保留2.661s通过。报告见 [用户数据](../acceptance/reports/extensions-data-2026-09-12.md)。全仓 race 会话90976仍运行，extensions已通过144.750s，待Master最终结果。make sdk默认沙箱在esbuild安装验证spawnSync遇EPERM退出2，已提权重跑；不能把首次失败计为通过。fixture均停止。当前下一动作：收取这两个运行会话，更新构建证据，然后补Windows主机监控真实代码（现有!linux文件仍是不可用返回）。

2026-09-12 用户数据迁移进行中：Manager 增加按登录用户隔离的有界 JSON 文档、CAS/稳定请求回执；data.read/data.write SDK 与 API 已接入。升级/回滚/重新安装保留数据时，由目标 WASI 后台逐用户迁移；编译一次、实例内存独立，整批 60 秒。数据与包共同事务恢复，旧事务缺少数据字段时不删现有数据。独立 schema1/2/故意失败3 包已生成；真实 fixture-1932133697 三版本服务器数据与窗口现场迁移/失败保留/回滚 1/1（11.7s）通过并已 Ctrl+C 停止。API 用户隔离/撤权 TLS 0.162s、transport 9/9、前端 33/33、OpenAPI 100 路径、Linux/Windows 构建通过。编译复用后的多用户定向 race 会话 42661 尚在运行，随后已启动全仓 race；收取会话结果后补报告。下一步补实际进程中断包含用户数据的恢复证据、迁移并发/配额边界；不能把早期仅包四文件的证据代替新增数据恢复。

2026-09-12 最新收尾检查点：修正后的独立包升级定向 race 通过（33.921s），原全仓 race 唯一失败项已在相同条件复验通过；原 make test 退出 2 的事实保留，不改写为成功。其余包均在该全仓运行中通过；最终扩展浏览器 5/5（17.2s）、真实独立包组合 1/1（9.6s）、前端 32/32、SDK/包及 Linux/Windows 构建已记录。无活动测试或 fixture。下一条具体动作：扩展按用户隔离的有界 JSON 数据读写、修订/CAS/幂等，升级以受限迁移模块转换数据并与包事务共同提交，失败保留旧包和旧数据；随后补真实两版本迁移。不要将既有 UserDataDir 空目录当成此功能已实现。

2026-09-12 全仓 race 收取：34837 已退出 2，所有其他包通过，Master（177.795s）仅失败于编译时尚未修复的独立包升级版本冲突。修复后普通定向 4.299s 已通过，正在同条件定向 race 复验；不把失败的 make test 写成通过。最终扩展浏览器 5/5（17.2s）通过，fixture 均停止。下一步完成本项 race 后继续按用户隔离的数据读写与升级迁移。

2026-09-12 SDK 桌面真实验证：独立包 0.2.0 的计算/结果/移窗/笔记与任务刷新恢复/快捷入口/关闭组合 1/1（9.6s）通过，fixture-2811626890 已 Ctrl+C 停止。前端 32/32，新增桌面场景连续 5/5；首轮未复现的快捷入口缺失保留记录。升级测试固定版本冲突已改为生成下一主版本，定向复验中。全仓 race 会话 34837 在修复前已启动，需收取最终结果，若只失败旧测试版本则复验该项并更新总状态；下一步继续后端用户数据迁移。见 [桌面 SDK 报告](../acceptance/reports/extensions-desktop-2026-09-12.md)。

2026-09-12 SDK 桌面能力进行中：新增 window.move/window.close/shortcut.create，沙箱只操作调用者自身标签，移入目标限同应用兼容窗口；快捷入口限当前资源。独立包更新为 0.2.0 并增加按钮，SDK/包、Linux/Windows 与前端构建通过，前端 32/32；新增浏览器首轮快捷入口未出现，尚未确定原因，复跑一次及连续五次通过（4.5s/21.3s）。正在真实 fixture-2811626890（127.0.0.1:9444，会话 54633）验证组合；结束后需 Ctrl+C 停止。尚不能把定向替身浏览器当作真实桌面 SDK 全链路证据。

2026-09-12 终态审计补齐：Master ReconcileTask 与派发前取消现在同事务写入终态审计；storage 全包 race 2.404s、故障注入原子性/重放去重及真实节点集成 4.402s 通过，make check/build 通过。无活动会话。下一条具体动作：SDK 自身标签移窗/关闭、快捷入口能力及沙箱身份限制，之后继续后端用户数据迁移。见 [审计报告](../acceptance/reports/task-audit-2026-09-12.md)。

2026-09-12 清单签名绑定：修复仅认证 payload 摘要的缺口，改为带版本域的完整 Manifest + payload SHA-256 签名；新增 extension-sign 工具，旧签名要求发布者重签。签名工具/安装/重开/落盘篡改拒绝与清单字段变更的定向 race 通过（1.027s/1.322s），真实 TLS 扩展回归通过（7.559s），make check/windows 通过。OpenAPI 为 99 路径/113 操作。无活动测试/fixture；下一条具体动作：补 Master ReconcileTask 终态审计的事务写入与重放去重证据，然后继续扩展后端用户数据迁移及完整 SDK 桌面能力。见 [签名报告](../acceptance/reports/extensions-signature-2026-09-12.md)。

2026-09-12 扩展任务闭环：SDK create/read/cancel、应用范围任务授权、执行前/模块获取取消分类、示例任务 ID 恢复已接入。transport 8/8、扩展浏览器 4/4（12.9s）、定向 Master/Daemon race（55.528s/1.331s）通过；真实双节点 fixture 的独立包计算、SDK 结果与刷新不重提交 1/1（8.6s）通过。Linux/前端构建通过，fixture-3028559533 已 Ctrl+C 停止。首次路由冲突已修复并复验；下一步修复包签名未认证 manifest 的缺口，然后继续后端用户数据迁移/完整 SDK 能力。见 [任务报告](../acceptance/reports/extensions-tasks-2026-09-12.md)。

2026-09-12 注册表事务恢复：安装/升级/回滚/卸载增加四文件旧快照及完整性摘要的持久撤销记录，随机提交令牌区分已提交与待恢复事务，启动前恢复未提交事务，读者拒绝待恢复包；损坏元数据不再被当作未安装覆盖。最终实际子进程退出/提交中断恢复定向 race 通过（1.385s），完整扩展模块此前通过（7.169s），最终 TLS 独立包安装/升级/计算通过（4.453s）。Windows 新增 MoveFileEx 写穿替换/删除及长路径处理；`make check build windows`、前端类型检查和 SDK 构建通过。新增 `.previous` 保留后缀避免 ID 与回滚快照冲突。当前无活动测试/fixture；本轮未重跑全仓 race，物理掉电、ENOSPC、Windows 真机耐久性仍未验证。证据见 [恢复报告](../acceptance/reports/extensions-recovery-2026-09-12.md)。

2026-09-12 生命周期/迁移增量：Manager 公共读写锁、一致清单/包体 Snapshot、回滚覆盖前校验、禁用状态保持与依赖门禁已接入；定向 race 通过（1.094s），真实 TLS 扩展后台回归通过（10.771s）。AppHost 响应式更新清单/能力/组件，SDK 增加沙箱 `migrate`，宿主对旧现场/资源身份做条件检查并一次提交新状态和版本，失败保留旧现场。前端单元 30/30、扩展浏览器 4/4（20.1s）、最终迁移定向 1/1（6.9s）、类型检查/SDK 和前端生产构建通过；`make check` 通过。当前无活动测试/fixture；本轮未重跑全仓 race，上一轮全仓结果不得替代本轮未覆盖场景。下一步补包写入的掉电事务恢复、后端数据迁移及真实两版本迁移证据。证据见 [生命周期报告](../acceptance/reports/extensions-lifecycle-2026-09-12.md)。

2026-09-12 WASI 链路收尾：完整独立包已在真实浏览器上传到 HTTPS Master，打开节点资源窗口并执行实际 WASI 笔记计算，1/1 通过（9.4s，`.local/fixture-1553746562`）；fixture 已 Ctrl+C 停止，按精确 fixture 标识检查无残留进程。最终 `GOCACHE=/tmp/blora-go-grant-idem make test` 全仓 race 通过（Master 191.796s，extensions 87.402s），Linux/Windows 二进制及前端生产构建通过，OpenAPI 3.1 解析通过（97 路径）。当前无活动测试/fixture。WASI 获取/编译/执行分别限时 30/30/5 秒，输入输出各 32 KiB；旧输入回传实现已删除。下一步按上方具体动作继续扩展能力和生命周期，不标记全量完成。证据见 [复核报告](../acceptance/reports/resume-2026-09-12.md)。

2026-09-12 WASI 生产链路：`Bundle` 前后端包格式、Master 版本绑定、Daemon 32 KiB 模块分块获取与沙箱执行已接入，独立示例后台计算笔记统计及摘要。真实 TLS 定向测试已通过（21.585s），扩展管理完整包上传/资源窗口浏览器 3/3 通过（12.3s）。全仓 race 首次因冷编译预算和测试等待不匹配失败，已分离编译/执行期限并加入单模块解码缓存，正在定向 race 复验；不能计作全仓通过。原回传生产代码已删除；下一步完成 race 修复、真实浏览器后台任务结果链路和完整升级/状态迁移。

2026-09-12 资源桥接增量：SDK `readResource()` 与 Master 扩展资源摘要 API 已接入，独立生成包在真实 TLS Master 安装和权限边界测试通过（0.156s）；transport 7/7、浏览器资源窗口与笔记独立 3/3、类型和包构建通过。WASI JSON 后台执行器锁定 wazero v1.12.0（间接升级 x/sys v0.44.0）；真实 Go WASI guest 的计算/文件隔离/空环境/输出上限/循环中止测试通过（11.649s），扩展模块完整回归通过（提权本机 TLS，7.373s），Windows 交叉构建通过，OpenAPI 3.1 解析为 97 路径。执行器尚未接入生产任务，不能替代现有回传实现或计作 E06 通过。当前无活动测试；下一步增加受完整性校验的前后端包格式、Daemon 分块制品获取与任务版本绑定，然后替换回传实现并验证独立包真实后台计算。

2026-09-12 续接修复：授权变更事务幂等、主机任务节点分发与普通用户扩展任务可见性已修复，三项真实 TLS 集成测试通过（1.482s）；证据见 [续接报告](../acceptance/reports/resume-2026-09-12.md)。独立参考扩展新增 esbuild 0.25.12 浏览器打包及真实 DOM 界面，`npm run package` 已通过；读取实际生成包的浏览器加载/立即刷新及原扩展场景 3/3 通过（12.4s，后端路由替身），相关 Go 包 vet 通过。无活动测试会话；下一步用独立包接入真实安装测试并补资源桥接、后台受限执行和迁移，不能用此局部证据结束 B09。

| 日期 | 决定 | 原因/影响 |
| --- | --- | --- |
| 2026-09-09 | 两份 v0.3 规划完整保留，另加全范围执行解释 | 用户要求把完整规划放入项目并供后续模型实施，不删改已确认产品行为 |
| 2026-09-09 | M0～M3 作为同一任务实施顺序 | 用户希望一次持续任务覆盖明确功能，之后再调细节 |
| 2026-09-09 | 真实实现与真实验证分开计数 | 缺少 Windows/Docker 等环境不能用桩函数或虚假“通过”掩盖 |
| 2026-09-09 | 本次只准备文档 | 用户本轮要求放置规划与行动指导，尚未要求运行完整开发或部署 |
| 2026-09-09 | 工作位置更正为 /data/instances/blora-panel | 按用户更正同步文档路径；新目录作为后续开发入口 |
| 2026-09-09 | 用户明确执行完整启动提示，开始产品实现 | 保留此前准备记录作为历史；全范围持续任务已启动 |
| 2026-09-09 | 按AGENTS条件分配desktop/runtime/protocol模块 | 当前技能列表无匹配的通用编码模块技能；主代理控制公共合同、依赖及整体验收 |

追加技术决定时注明影响、替代方案和关联验收，不覆盖历史记录。

## 8. 交接规则

后续模型先读当前记录再看必要源码和失败证据，继续未完成内容，不重新初始化已有工程，不要求用户重复已经确认的需求。

如果平台强制中断，交接必须能回答：做到哪里、真实通过什么、什么失败、当前代码可否运行、下一步执行什么。不能只留下“后续完善”。如果全范围尚有缺口，最终回复如实列出；不能为了结束任务把缺口改写成视觉细节优化。
