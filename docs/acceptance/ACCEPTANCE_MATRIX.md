# Blora Panel 全范围验收矩阵

2026-10-03 Windows第九轮增量：见[回传报告](reports/windows-ninth-report-2026-10-03.md)，固定`3e6c31c`，摘要14/14匹配，WebKit14/15，无skip/retry/flaky；四项导航/日志用例各3次通过，前轮两项读取恢复/建立超时已获得真机通过证据。仅第一次fallback/worker终端达到45秒总时限，之后18.810/18.762秒通过，诊断无页面/请求错误；首次刷新约32.1秒才开始，具体慢动作/Windows因果仍未确立。补测现独立类型构建正式bundle，fresh preview禁止端口回退/服务器复用，缺失或失败构建阻止成功；实时终端阶段及有界静态资源耗时加入失败诊断，原UI动作/45秒时限/5秒恢复/worker/ACK/输入不重放断言保持，产品和UI未改。本地正式WebKit36/36、Firefox12/12、Chromium12/12、类型构建及最终脚本harness/实际JSON/严格15项选择核验通过；隔离冷启动单例case21.625秒→14.766秒，不冒充Windows因果/E08。故意失败探针确认有界诊断实际进入报告，其初次cwd错误独立保留，不计产品PASS。Windows实效仍待固定新版`-Followup -BrowsersOnly -WebKitOnly`五项各3次回传。Go/SDK/其他浏览器本轮Windows未选择，不导入旧PASS；原生/E08/外部环境整体验收保持未关闭。

2026-10-03 Windows第八轮增量：见[回传报告](reports/windows-eighth-report-2026-10-03.md)，固定`57d4fb6`，摘要15/15匹配，WebKit13/15，无skip/retry/flaky；原fallback/worker终端各3次通过，剩一次汇总恢复5秒超时及一次日志待取消读取建立超时，后者诊断visible/focused但paint=false。可控无帧基线0/1在原5秒断言失败，共享只读API门禁新增250ms独立恢复兜底，pagehide取消并失效旧帧/旧截止，新导航重置代次。日志测试通过实际interval回调建立待取消读取，queued观察不再依赖paint结束；原错误/限时/ACK/输入不重放断言保持。89单测、类型构建、本地WebKit36/36、Firefox12/12、Chromium19/19加独立补验4/4、PowerShell harness/实际JSON与精确15次选择通过；Windows修复实效仍待新版`-Followup -BrowsersOnly -WebKitOnly`（五项各3次严格15/15）回传。Go/SDK/其他浏览器未在本轮Windows选择中执行，旧PASS不导入；有界兜底不证明所有慢导航/native dialog/bfcache序列，完整平台/E08及外部环境整体验收保持未关闭。

2026-10-03 Windows第七轮增量：见[回传报告](reports/windows-seventh-report-2026-10-03.md)，固定`8ad386a`、摘要14/14匹配，WebKit11/12，无skip/retry/flaky；三项任务查询导航各3次获得Windows通过证据。剩余fallback/worker终端错误来自InstanceConsole独立日志轮询，在beforeunload后9ms、pagehide前1ms。可控日志轮询基线0/1，已增加共享安全API读取导航取消/等待、response body取消检查、pagehide旧帧失效和组件销毁取消，原终端/ACK/不重放断言保持；87单测/类型构建、本地Chromium23/23、Firefox12/12、WebKit36/36通过，最终PowerShell harness/真实JSON解析及精确选择清单核验通过。下一次仅5项各3次严格15/15的`-Followup -BrowsersOnly -WebKitOnly`；Go/SDK/其他浏览器不导入旧PASS，修复实效仍待Windows回传；本条不提升完整平台/E08或外部环境整体验收。

2026-10-01 第六轮 Windows 实机证据：固定 `f98b078`、followup-browser，20文件校验值全部匹配；Chromium/Firefox各8/8，WebKit23/24，仅一次fallback/worker终端任务列表请求在beforeunload后8ms、pagehide前8ms报访问控制诊断。前轮两项查询导航检查在Windows WebKit各3次通过，Go/native/SDK本轮未选，不能导入旧PASS。可控真实QueryObserver排队回调复现取消后仍启动新读取，基线0/1；新增只读任务查询门禁与存活页面恢复，组合本地Chromium9/9、Firefox9/9、WebKit27/27，无skip/retry/flaky，80单测/类型构建及脚本harness通过。新增 `-Followup -BrowsersOnly -WebKitOnly`，只跑剩余终端配置与三项导航，各3次严格12/12，不重复已验证Go/其他浏览器；见[第六轮报告](reports/windows-sixth-report-2026-10-01.md)。本地复现关闭的是取消后排队新读取缺口，Windows原诊断实效及完整平台/E08/外部环境仍待实际验证。

2026-10-01 第五轮 Windows 实机证据：固定 `3c7abb3` 的11必需原生检查、3项Go回归各3次通过；Master/runlog完整race87pass/0fail/4skip，前轮两处Go竞争修复已有真机PASS，WebKit账号场景也全部三轮通过。Chromium/Firefox各6/6，WebKit16/18，仅第三轮两种终端配置的后台任务汇总访问控制诊断仍失败。复现页面离开前读取未取消，任务查询及共享列表接入AbortSignal，在beforeunload/pagehide取消只读查询；最终本地Chromium8/8（79.608秒）、Firefox8/8（115.302秒）、WebKit24/24（263.222秒）通过，无skip/retry/flaky；80单测/类型检查/构建及PowerShell harness通过。新增 `-Followup -BrowsersOnly`，严格要求8/8/24执行，不重复已验证Go/GCC/SDK也不导入旧PASS；见[第五轮报告](reports/windows-fifth-report-2026-10-01.md)。Windows剩余诊断及完整平台/E08/外部环境缺口仍待验证，不能把Linux通过当真机修复通过。

2026-10-01 第四轮 Windows 实机证据：固定 `08db099` 的 27 项定向 Go 与 11 项必需原生全部通过，目录 metadata/ConPTY Unicode 实际执行、Windows 可移植生命周期与 source restart 已获得真机证据；Chromium/Firefox 各32/32，WebKit29/32。全race事件356pass/2fail/14skip，新增取消上传的调度竞争与日志收尾记录竞争已在本地修正，Master/runlog 完整race退出0；三项修正定向race连续8轮通过。最终浏览器夹具本地Chromium/Firefox各6/6、WebKit18/18通过，原始失败和首轮本地失败独立保留，不能因 Linux 通过提升 Windows 整项。`-Followup` 限定两个 Go 模块及两份浏览器文件，校验实际次数并带失败诊断；详见[第四轮报告](reports/windows-fourth-report-2026-10-01.md)。E08、完整 Windows/特权生命周期和外部环境缺口不提升为完成。

2026-09-30 Windows首次用户实机证据：Job/keeper/daemon退出恢复、日志管道、监控、SCM/Task Scheduler查询等九项关键原生通过，ConPTY关键项失败；全Go仍有产品兼容缺陷和Linux夹具不兼容，浏览器尚未运行、race缺GCC。Web构建与76单测通过。首批修正及准确边界见[Windows首轮报告](reports/windows-first-report-2026-09-30.md)，本地Linux回归/Windows交叉编译不是修复后Windows运行证据，不提升Windows整体验收。

2026-09-29 F06/RC34：复现旧连接异步解析失败污染新连接，增加连接代次/销毁保护；当前连接错误仍停止输入。Chromium/Firefox/WebKit定向组合各4/4通过；六包2,251条记录/三份Web、独立停机恢复及兼容RC33回退通过，见[报告](reports/terminal-generation-2026-09-29.md)。RC33完整复验49/51，两项扩展失败保留；补原生鼠标hover命中前置后RC34完整真实功能51/51通过897.040秒、无skip。玻璃遮挡裁剪像素等价检查失败未采用，UI未变；功能套件不含独立E08负载门禁，严格E08及原偶发/外部平台缺口保留。

2026-09-29 F06/RC33：实际复现并修复OSC 8链接跨检查点丢失，保留主/备用/历史屏幕范围与当前输出属性，并使用原生标记管理生命周期。Chromium61项、Firefox/WebKit各50项、76单测、真实双节点PTY7项通过；实际鼠标点击和两次持久恢复目标一致，原生URI限制保留。RC33六包2,243条记录/三份Web及独立恢复/兼容RC32回退通过，见[报告](reports/terminal-links-recovery-2026-09-29.md)。UI未改；原Vim偶发、严格E08和外部平台缺口未关闭。

2026-09-29 F06/RC32：修复主题切换后检查点/原始日志中的旧调色板复活，记录本地重置在输出流中的位置，直接重置原生调色板而不向半段CSI/OSC插入字节。Chromium48项组合、最终交错六项及非法标记1项、Firefox/WebKit各40项、73单测与真实双节点PTY7项通过；RC32六包2,239条记录/三份Web及独立恢复/兼容RC31回退通过，见[报告](reports/terminal-theme-order-2026-09-29.md)。UI未改，链接恢复/原Vim偶发/严格E08/外部平台缺口保留。

2026-09-29 F06/RC31：实际复现并修复OSC索引色/前景/背景/光标颜色恢复丢失，采用最多259槽的数值差异记录与固定指令恢复，保留主题默认重置语义。三浏览器各33项、73单测、真实双节点PTY7项、RC31六包2,235条记录及三份Web一致性、独立停机恢复与兼容RC30回退通过，见[报告](reports/terminal-colour-state-2026-09-29.md)。其他终端恢复边界、原Vim偶发、严格E08与外部平台未据此关闭。

2026-09-29 F06/RC30：修复鼠标SGR/像素编码、光标显隐/闪烁/样式及换行转换跨检查点丢失。真实查询首轮四项失败，修复后Chromium/Firefox/WebKit各27项（含实际鼠标按下/释放字节对照）、71单测通过；真实双节点PTY首轮6/7超时，保持原断言和60秒限时串行复核7/7通过。RC30六包2,231条记录及兼容RC29恢复回退通过，见[报告](reports/terminal-protocol-state-2026-09-29.md)。不据此关闭原Vim偶发、严格E08及外部平台缺口。

2026-09-29 F06/RC29：实际复现并修复字符集、滚动区域、保存光标和制表位跨检查点丢失，新增两缓冲区及字符集版本化状态并同步Worker。Chromium30项组合、持久存储重开19项、Firefox19项、71单测和真实双节点PTY7项通过；RC29六包2,227条记录及兼容RC28恢复/回退通过。WebKit首轮18/19（一项模块加载失败），原断言全组复核19/19通过，首轮失败保留，见[报告](reports/terminal-control-state-2026-09-29.md)。UI未改，E08/原Vim偶发/外部平台缺口不关闭。

2026-09-29 F06/RC28：修复未完成VT控制序列跨检查点丢失，按真实解析器边界压缩，保留有界输出/resize日志并拒绝Worker过时快照。Chromium组合24项与晚返回1项、Firefox/WebKit各21项、正式配置各2项、69单测和真实双节点PTY7项通过；六包2,223条记录及恢复/兼容RC27回退通过，见[报告](reports/terminal-parser-checkpoint-2026-09-29.md)。原Vim偶发、已完成控制指令的其他持久状态、严格E08与外部平台缺口未据此关闭。

2026-09-29 F06/恢复增量：真实复现并修复终端未完成UTF-8码点跨检查点丢失，保留最多3字节解码续接状态；六个分割位置/Worker/查询9项、检查点+日志混合1项、68单测、真实双节点文件/PTY7项通过。RC27六包、2,219条记录/三份Web、独立启动/停机恢复/兼容RC26回退通过，见[报告](reports/terminal-utf8-checkpoint-2026-09-29.md)。这是新的确定恢复缺陷，不是Vim原偶发的根因证明；尚未据此证明VT控制序列全部边界，严格E08与外部平台缺口保持。

2026-09-29 RC26最终补验：完整HTTPS/双节点/Docker功能 **51/51、905.725秒通过**；新增专属子组OOM及PID上限实际触发，与原systemd/cgroup场景合计5项通过、9.652秒，见[平台报告](reports/systemd-cgroup-2026-09-29.md)。Linux X11/DBus/Dunst原生弹窗、文字、系统鼠标点击、刷新去重及浏览器退出后遗留弹窗可关闭通过，最终9.666秒，见[通知报告](reports/native-notifications-2026-09-29.md)。退出即自动消失的早期预期失败保留，未声称后台投递或退出后重启跳转。当前UI与RC26产品未再变更；严格E08、原Vim偶发来源、Windows、跨主机和物理故障等缺口保留。

2026-09-29 F13/E05/A03/E09平台增量：独立私有systemd容器真实运行3项通过，覆盖服务启动/重启/停止及规范名/别名列表、timer可见性/enable/disable/实际触发、cgroup进程出生归属/整组终止/三项配额文件/CPU实际限流。修复原停止服务与已安装未加载timer列表遗漏。runtime/systeminfo race及Master/Daemon相关7项通过；细节和失败保留见[平台报告](reports/systemd-cgroup-2026-09-29.md)。不再将Linux对应子项写作完全缺环境；未计Windows、OOM/进程耗尽、生产部署、跨主机和物理故障通过。RC26六包/2,215条内部记录/独立启动、停机恢复和兼容RC25回退已通过，原E08/Vim缺口不变。

2026-09-29 终端/合成诊断增量：真实Vim/top10/10与最终版本3/3通过但未复现原偶发；新增失败字节附件，不放宽输入不重放断言。真实xterm/Worker/IDB查询回放/权限边界与现有终端定向6/6、最终类型检查通过。跟踪明确主要耗时位于SoftwareRenderer合成；显式SwiftShader合成更慢未采用。60秒真实循环传输诊断240样本p95 1062.5ms，在原50ms断言失败，不能替代RC25正式三引擎E08结果；详见[补验报告](reports/local-diagnostics-2026-09-29.md)。无产品/UI/发行包变更，不将未复现或缺环境视为完成。

2026-09-29 RC25非视觉收尾：恢复快照去除重复复制、重复聚焦不再写恢复日志，66项单测、59/59完整mock、三引擎原生IDB边界及Firefox/WebKit真实恢复故障各3/3通过。真实HTTPS/Docker完整套件为50/51（Vim刷新额外字节一次失败）；带诊断和移除诊断后的原用例各连续3/3通过，保留偶发未定位记录，不能写成首轮全过。RC25六包构建、外部摘要与2,207条包内记录、三份当前Web一致性、独立启动/停机恢复/兼容RC24回退通过，见[完整本轮报告](reports/local-closeout-2026-09-29.md)。当前严格E08 Chromium/Firefox/WebKit分别1003.9/1108/545ms，均大于50ms；快照专项35%～56%的中位改善不代表整页达标。UI和通透效果全程保留。24h终态证据沿用下条已通过结果；Windows/systemd等环境缺口及整体验收状态不变。

2026-09-29 E01/E04/A09 本地24小时长测终态复核：私有原始检查点与事件经独立只读收取器重新校验，状态 **PASSED_24H**、证据缺口0，生成的[最终报告](reports/local-endurance-24h-2026-09-27.md)与现有文件完全一致。实际单调时长86,415.480秒、跨2个UTC日期、17,262次真实采样、1,440个成功分钟槽，6h/18h两次失联和同库重启后保留原runId；终态SQLite完整性、最新真实备份正文恢复、所属计划与服务清理均通过。Master/Daemon RSS峰值34,041,856/34,013,184字节，FD峰值14/20，归档峰值16,776,441字节，状态与证据峰值265,899,979字节，均在该次声明预算内。下方9月26～27日的“运行中”是历史阶段记录，现以本条终态为准。此结果只关闭本机这一次24小时场景；当前E08三引擎严格性能仍失败，Windows真机、独立systemd、跨主机、物理掉电和系统通知中心仍缺实际验证，不将全范围标为完成。

2026-09-27 11:14 长测阶段证据：独立24h运行的6h与18h两次实际故障阶段均完成，管理端同库重启前后120点stale历史精确一致，Daemon恢复后保持原实例runId；当前18h01m/12,956采样/1,081个成功分钟slot，仍RUNNING，不计24h通过。新只读收取器已独立启动，会在实际终态核对恢复、SQLite完整性、所属资源清理及不可变二进制，并更新[动态报告](reports/local-endurance-24h-2026-09-27.md)；其短测与拒绝错误证据测试通过，详见[监督/收取入口](reports/local-endurance-2026-09-26.md)。整项与外部平台仍不提升为完成。

2026-09-27 当前冻结UI复核：完整mock浏览器56/56、另新增运行中Worker实际终止1/1、原失败扩展链路重复6/6、完整真实HTTPS/Docker功能 **51/51**（无skip）通过，见[本地功能报告](reports/local-functional-2026-09-27.md)；Go race各包与修复后的完整Master分段复验通过，未将首轮失败改写为通过。RC24六包构建/外部SHA及2,191条包内记录核验、独立SDK/参考包签名、TLS/双Daemon、停机恢复及兼容RC23回退全部退出0，见[最新发行报告](reports/release-rc24-2026-09-27.md)。E08严格八窗口/双真实PTY/万项目录/16MiB并行传输三引擎仍未达标（Chromium1240.1ms、Firefox1230ms、WebKit739ms），新合成对照亦失败，见[当前性能报告](reports/performance-current-2026-09-26.md)；9月21～22日通过不覆盖本版本。Firefox/WebKit恢复故障各3/3通过，独立24h跨日监督仍RUNNING，不能提前计作通过；Windows等外部平台仍未验证，整体验收不提升为完成。

2026-09-26 F02/F08/A11 非 Chromium 恢复故障补验：Firefox 155.0 **3/3（37.5s）**、WebKit 26.6 **3/3（50.8s）**、退出0，真实 HTTPS/Monaco/IndexedDB 验证最近中文正文导出/恢复、schema 0 迁移与不支持版本原样保留，以及真实事务 abort 后快照/指针原子回滚；详见[跨引擎恢复报告](reports/recovery-engines-2026-09-26.md)。配额异常是定向故障注入，不是实际磁盘/浏览器配额耗尽；不代表物理掉电通过。

2026-09-26 E01/E04/A09 长测入口：新增独立受监督的 `scripts/local-endurance.py`，两个完整五分钟真实短测通过，包含真实分钟调度唯一性、120点历史、Daemon失联、同库Master重启、原runId恢复、最新备份正文恢复和准确清理；stop/resume 与监督 worker 中断后安全接管也通过。24小时运行 `.local/endurance-zs_dzpzj` 已启动但**尚未通过**，需跨真实UTC日期并满足墙钟/单调时钟、采样覆盖、故障阶段、RSS/FD/归档上限及末尾恢复核对，详见[长测报告](reports/local-endurance-2026-09-26.md)。整项与外部平台状态不提升。

2026-09-26 F02 全外观截图证据：当前六配色 × 浅/深色 × 通透开/关 × 桌面/启动器/实例中心共 **72 张**真实 Vue 原图，原图分辨率 3200×2000，正式图标和暖砂深色壁纸 SHA256 与当前资源一致；本地 GET-only 演示数据核验材质开关不改壁纸/图标、零页面错误/API 写入/外网请求，见[验证记录](../../web/ui-screenshots/all-appearance-20260926/verification.json)及[原图 ZIP](../../web/ui-screenshots/all-appearance-20260926/blora-all-appearance-originals-20260926.zip)。本轮未改产品源码，截图只用于远程视觉复查，不提升 F02/A01/A02 或后端/外部平台整体验收状态。

2026-09-25 F02 暖砂深色壁纸视觉增量：旧 `sand-dark.svg` 的近黑底覆盖约 70% 画面，现改为暖炭色并增加上方曲面、上移沙丘层次，浅色壁纸与图标、主题/通透逻辑保持独立。真实 Vue [深浅场景对照](../../web/ui-screenshots/sand-wallpaper-rework/00-overview.png)和[验证记录](../../web/ui-screenshots/sand-wallpaper-rework/verification.json)覆盖浅色、深色通透/实色共六张桌面与启动器，零页面错误/API 写入。`cd web && npm run build`（含类型检查）通过，暖砂深色/自选壁纸聚焦 Playwright **1/1** 通过；此视觉改善不改变 F02/A01/A02 整体验收状态。

2026-09-25 F02 备份与计划应用图标视觉增量：去掉旧回转圆环左侧断口和外溢钟面，使用闭合表盘及内收钟面；同步正式浅/深色 01A 图标。生产构建通过，聚焦 Playwright **2/2** 通过；[真实界面实际尺寸对照](../../web/ui-screenshots/backup-icon-final/00-comparison.png)与[验证记录](../../web/ui-screenshots/backup-icon-final/verification.json)覆盖桌面、Dock、启动器，核实正式资源加载及零页面错误/API 写入。此视觉证据不提升 F02/A01/A02 整体验收状态。

2026-09-24 F02 暖砂深色壁纸二次配色修订：用户否定青蓝版后，正式深色暖砂改为近黑底与浅米金曲面，原几何、03 图标及通透/配色独立性保持。生产构建通过，相关 Playwright **3/3** 通过；[六配色完整截图目录](../../web/ui-screenshots/dark-wallpaper-final/index.html)和[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)证实六壁纸互异、通透开关不改壁纸、零页面错误/API 写入。远程用户须直接收到 PNG 嵌图；此 UI 证据不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 暖砂深色壁纸配色修订：灰棕旧版改为深蓝绿曲面，保留原几何以及已选 03 图标，浅色、配色选择与通透材质逻辑不变。`cd web && npm run build` 通过，相关 Playwright **3/3** 通过；[真实桌面与启动器图库](../../web/ui-screenshots/dark-wallpaper-final/index.html)和[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)证实六套壁纸互异、材质开关不改壁纸、零页面错误/API 写入。该 UI 修订不改变 F02/A01/A02 整体验收状态。

2026-09-24 F02 深色壁纸可见度增量：六套默认主题改用保留原几何的深色专用 SVG，修复约 90% 遮罩令壁纸近乎纯色的问题；青灰深色自选壁纸不再被默认规则覆盖。浅色主题、通透开关与已选 03 深色图标独立。`cd web && npm run build` 通过，相关桌面浏览器测试 **4/4** 通过；[六配色实景图库](../../web/ui-screenshots/dark-wallpaper-final/index.html)与[验证记录](../../web/ui-screenshots/dark-wallpaper-final/verification.json)核验暗色壁纸互异、通透 on/off 背景一致、零页面错误/API 写入。此 UI 证据不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 深色应用图标定稿：用户选择 03「清晰造型」作为默认深色图标；14 枚正式资源已切换，浅色资源、配色和通透开关独立。生产构建及桌面主题/材质聚焦 Playwright **2/2** 通过；[七向实景与验证记录](../../web/ui-screenshots/dark-icon-directions/index.html)可复查。此 UI 决定不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 深色通透材质与亮边增量：用户截图对应通透 `on`，此前暗色面约 96–97% 遮盖；现启动器/顶栏约 80%、窗口约 89%、Dock 约 78%，内侧反射、模糊和投影收减，深色窗口/启动器/Dock 外圈改用暗壳边，启动器分隔线压暗。14 枚深色图标改色，并降低仅深色 h04 导出中的反射与壳层亮度；浅色资源及六配色独立性保留。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[通透 `on` 的六配色×深浅×六场景图库](../../web/ui-screenshots/dark-mode-review/index.html)为 **72 张**，另有暖砂 `off` **12 张**，可对照[深色启动器 `on`](../../web/ui-screenshots/dark-mode-review/sand-on/08-dark-launcher.png)和[`off`](../../web/ui-screenshots/dark-mode-review/sand-off/08-dark-launcher.png)。本地隔离 fixture 为零页面错误、零 API 写入；此 UI 增量不提升 F02/A01/A02 的完整状态，亦不覆盖真实节点、后端或外部平台验收。

2026-09-24 F02 深色窗口与图标增量：深色窗口外框改为弱中性边缘，实例中心内容区底角不再露出壳体形成接缝；14 枚深色应用图标针对暗背景调整高亮、底板和主体对比，浅色界面及六套配色、壁纸选择不变。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[六配色×深浅×六场景图库](../../web/ui-screenshots/dark-mode-review/index.html)有 72 张实景及[实例中心左下角局部](../../web/ui-screenshots/dark-mode-review/sand-on/09-dark-instances-corner.png)，本地隔离 fixture 核验零页面错误、零 API 写入。该 UI 证据不提升 F02/A01/A02 的完整状态，也不覆盖真实节点、后端或外部平台验收。

2026-09-24 F02 深色外观清晰度修订：六套配色共用中性暗色表面，强调色仍独立；调整暗色壁纸呈现、近不透明窗口/Dock、列表边界及终端画布，重绘14枚更易辨认的01A深色图标。原壁纸资源、选择项与浅色模式不变。生产构建（含类型检查）、聚焦桌面浏览器 **4/4** 与终端恢复回归 **1/1** 通过；[六配色×六场景深浅对照](../../web/ui-screenshots/dark-mode-review/index.html)重拍72张原图与六张概览，逐套核验14枚浅/深色图标，0页面错误、0 API写入。截图为本地隔离演示数据，不提升F02/A01/A02或外部平台的完整验收状态。

2026-09-24 F02 工作区深色外观增量：保持现有六套配色和壁纸选项不变，选定 01A 应用图标新增 14 枚随浅/深色切换的暗色资源；深色窗口、Dock、菜单、输入/选择控件、资源列表、状态、应用卡片、日志及已打开的 Monaco 编辑器同步适配。生产构建（含类型检查）、聚焦浏览器 **4/4** 通过；[六配色×六场景×深浅两版实景图库](../../web/ui-screenshots/dark-mode-review/index.html)包含 72 张原图、六张概览，每套分别核验 14 枚浅色/深色图标、0 页面错误和 0 API 写入。截图为本地隔离演示数据，不提升 F02/A01/A02 或外部平台的完整验收状态；壁纸未在本轮改动。

2026-09-24 F02 工作区壁纸艺术媒介增量：针对旧方案画风相同的问题，新添真实摄影、印刷海报、数字夜光、水墨纸张及压纹玻璃五个方向；摄影使用[NPS 公有领域素材](https://npgallery.nps.gov/AssetDetail/204B23EC-1DD8-B71B-0B21B310C2A73B98)，来源和本地处理见[记录](../../web/src/appearance/wallpapers/media/PHOTO_SOURCE.md)。原壁纸与默认值保留；当前11种自选壁纸不改配色、通透材质或图标。生产构建及浏览器聚焦 **3/3** 通过；[五方向真实界面图库](../../web/ui-screenshots/wallpaper-art/index.html)共20张原图/4张对照图，核验页面错误和API写入均为0。截图为隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区壁纸偏好增量：新增六种不同构图的 SVG 壁纸，并与主题配色、浅/深色和通透材质分开选择；“随配色”为兼容原有外观的默认值。聚焦浏览器测试 **3/3**、生产构建通过；[六种壁纸实景图库](../../web/ui-screenshots/wallpaper-styles/index.html)含24张真实界面截图及4张对照图，核验六张各异、配色与14个应用图标不变、零页面错误或 API 写入。截图使用隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区外观配色增量：主题配色由冰蓝/青灰扩展为冰蓝、青灰、暖砂、苔绿、雾紫、烟粉六套，各有浅/深色令牌及配套壁纸；材质开关、配色和浅/深色保持独立，应用图标一致。当前聚焦浏览器测试 **2/2**、生产构建通过；[六套配色实景图库](../../web/ui-screenshots/palette-showcase/index.html)共72张，覆盖两种材质和六类场景，核验六套各异主色/壁纸、两材质同配色/图标及零页面错误/远端写入。截图为本地隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区外观偏好增量：通透材质开关与冰蓝/青灰主题配色已独立持久化；原有浅/深色主题独立保留。当前浏览器聚焦场景覆盖四组合和刷新 **1/1**，生产构建通过；[25场景四组合图库](../../web/ui-screenshots/palette-matrix/index.html)共100张本地截图，验证同配色开/关主色、壁纸和14图标资源一致，材质不同。截图使用隔离演示数据，不提升 F02/A01/A02 等工作现场恢复或外部平台验收状态。

2026-09-22 RC23 当前源码发行复核：`development-20260922-rc23` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC22 兼容回退均通过，见[RC23发行报告](reports/release-rc23-2026-09-22.md)。包内含 Docker 日志归档实时/历史入口、F08 只读分段预览、预览模式控件保护和最新验收台账；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 RC22 当前源码发行复核：`development-20260922-rc22` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC21 兼容回退均通过，见[RC22发行报告](reports/release-rc22-2026-09-22.md)。包内含 Docker 日志归档实时/历史入口和 RC21 的 F08 只读分段预览；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 Docker 日志归档入口增量：`DockerLogs.vue` 已接入现有有界历史接口，支持实时/归档切换、归档窗口观察时间、截断与可能缺口标记；Docker 浏览器 **2/2（11.1s）**、真实隔离 Docker Engine Master/Daemon 归档 API **2.79s**及 `npm run check`、63/63 单测、生产构建、`make check` 通过，真实链路还覆盖未授权403和 `limit=101` 400，详见[Docker日志归档报告](reports/docker-log-history-2026-09-22.md)。这只补齐 Linux/浏览器/Docker 可验证子项，不改变 F11/E02 对远程 Engine、Windows 容器、物理故障和长时组合的进行中状态。

2026-09-22 RC21 当前源码发行复核：`development-20260922-rc21` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC20 兼容回退均通过，见[RC21发行报告](reports/release-rc21-2026-09-22.md)。包内含 F08 只读分段预览、预览模式控件保护与 OpenAPI `previewFile`；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 F08 大文件只读分段预览：新增受文件版本绑定的 `/files/preview` 有界 UTF-8 分片与浏览器上一段/下一段导航；真实文件 API 覆盖超限文本首段、第二段及二进制拒绝，编辑器边界浏览器 **3/3** 通过。完整 F08 交叉场景和外部平台证据仍按下表保持进行中，详见[编辑器能力增量](reports/editor-capabilities-2026-09-22.md)。

2026-09-22 RC19 当前源码发行复核：`development-20260922-rc19` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC18 兼容回退，见[RC19发行报告](reports/release-rc19-2026-09-22.md)。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC18 当前源码发行复核：加入快捷键偏好后的 `development-20260922-rc18` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC17 兼容回退，见[RC18发行报告](reports/release-rc18-2026-09-22.md)。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC17 当前源码发行复核：`development-20260922-rc17` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC16 兼容回退，见[RC17发行报告](reports/release-rc17-2026-09-22.md)。本轮还验证工作区主题/字号/密度偏好与默认布局备份恢复浏览器场景3/3。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC16 当前源码发行复核：`development-20260922-rc16` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC15 兼容回退，见[RC16发行报告](reports/release-rc16-2026-09-22.md)。包内含当前文件目录树/列表与图标视图/框选、实例列表与卡片视图改动。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-21 最新当前源码复核：`development-20260921-rc15` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建、SHA256 校验和独立包启动/TLS/登录/双 Daemon/停机快照恢复/兼容 RC14 回退冒烟均通过，见[RC15发行报告](reports/release-rc15-2026-09-21.md)。RC15 当前源码常规真实浏览器套件 **49/49**（11.5分钟）通过，见[RC15浏览器回归报告](reports/browser-regression-rc15-2026-09-21.md)；独立一小时 E08 也已通过，本机 p95 为48.1ms，见[恢复克隆优化小时报告](reports/performance-hour-recovery-copy-2026-09-21.md)。全范围各项继续按下表保留进行中：这些本机结果不覆盖 Windows 真机、systemd、远程网络或物理故障。历史 RC14 失败与 RC13 证据仍保留在下方及[RC13报告](reports/rc13-current-source-2026-09-20.md)。

2026-09-21 A09 RC15 当前源码定向 race 复验：真实TLS慢日志消费者、8.4MB跨节点传输、限时停止及归档/目标核对通过（13.918s），见[混合流报告](reports/mixed-stream-control-2026-09-13.md)。长时全连接内存边界和跨平台负载仍未验证。

2026-09-21 E04 Linux进程崩溃子项：独立测试进程接受定时 occurrence/task 后遭强制终止，另一个进程重开 SQLite，accepted receipt 与任务/计划身份不变，旧 occurrence 不重放，撤权后的下一次触发被阻止；定向 race 1.130s、备份全包 race 2.351s通过，Windows amd64 测试程序交叉编译通过。见[定时调度进程强制终止报告](reports/scheduler-process-crash-2026-09-21.md)。交叉编译不代表 Windows 运行；本测试也不覆盖事务中途掉电或 Master+Daemon 整体进程崩溃，E04/F12仍进行中。

2026-09-21 E03/F12 Linux恢复进程崩溃子项：真实 `Manager.Restore` 在8 MiB文件已有1 MiB写入暂存区时遭SIGKILL；重开服务后原ID为明确部分进度 `interrupted`，部分上传取消且清理、目标正文未发布，同ID不重放，新恢复计划重新恢复并核对SHA成功。定向race连续5次12.408s、备份全包race6.864s、`go vet ./internal/backup`及Windows测试程序交叉编译通过，见[恢复进程崩溃报告](reports/backup-restore-process-crash-2026-09-21.md)。Windows运行、设备写入中断及物理掉电仍未验证。

2026-09-21 E04 时区边界证据复核：当前源码 `TestSchedulerTimezoneDSTAndFiveFieldContract` 覆盖纽约春季不存在的 02:30 不补跑、秋季重复的 01:30 对应两个不同 UTC slot；`TestSchedulerSkipPoliciesAndBoundedLongDowntime` 覆盖40天停机后的有限跳过策略。它们包含在本轮通过的 `go test -race ./internal/backup` 中，但不替代多日真实墙钟运行或设备掉电验收。

2026-09-21 E01 UTC跨日key顺序补验：节点和实例历史现有125点测试跨过UTC午夜，核实裁剪后的120点时间范围从次日00:00:01到00:02:00且最新实例样本未被错序；`go test -race` 定向两项通过（2.718s），见[历史测试源码](../../internal/master/monitor_integration_test.go)。这是确定性时间戳边界验证，不是24小时墙钟采样；Windows和非Chromium仍未验证。

2026-09-21 E01 Firefox真实浏览器：Firefox 155.0/Playwright v1543 上，真实节点121点采样、离线 stale、Master重启后持久历史读取和 Daemon 恢复组合1/1通过（2.9分钟）；`BLORA_BROWSER=firefox` 经真实 HTTPS 双节点 fixture 验证。Chromium仍为默认引擎，Firefox临时profile需在沙箱外启动。见[长时监控报告](reports/monitor-history-soak-2026-09-21.md)；Windows、WebKit、24小时实时运行仍缺。

2026-09-21 rc14 当前源码一小时混合负载已实际完成，但 E08 未通过：3606.506秒/45,072次响应采样，p95 **51.7ms > 50ms**、max159.4ms、1312/190831长帧；8窗口、双PTY、10,000项目录、16MiB循环复制均运行，162轮传输和末轮目标均通过。队列峰值86,193B，heap采样36.9–135.9MB；过程RSS/归档采样见[本轮报告](reports/performance-hour-rc14-2026-09-21.md)。随后两次诊断工具开启的60秒样本均p95 47.4ms，但有测量扰动，不能覆盖小时失败；CPU热点分散，见[诊断报告](reports/performance-diagnostic-rc14-2026-09-21.md)。直接写窗口几何的非视觉优化尝试，短测p95 57.0/54.9ms且仍超标，已回退；细节见[拖窗优化尝试报告](reports/performance-drag-experiment-rc14-2026-09-21.md)。当前源码49项常规真实浏览器套件一次49/49通过，E08仍需找到可重复优化再决定小时复验。

2026-09-21 后续当前源码复验已达成本机 E08 一小时目标：恢复值深拷贝热点改用原生 `structuredClone` 并保留代理及 JSON 清洗回退；类型检查、50/50前端单测、生产构建、27/27相关真实浏览器测试通过。无 profiler 一小时场景 3,604.492秒、48,840响应样本，p95 **48.1ms**（≤50ms）、max160.4ms、731/197,137长帧，165轮跨节点传输均校验成功；62组资源采样含57组负载中样本，RSS峰值Master 39.4MB/Daemon 32.0MB，单终端归档16,774,920B、队列峰值86,180B，无采样/轮转竞态。详见[恢复克隆优化小时报告](reports/performance-hour-recovery-copy-2026-09-21.md)及[原始指标](reports/performance-hour-recovery-copy-2026-09-21.json)。这次通过取代RC14的本机p95失败，但不覆盖公网远端、高RTT/丢包、Windows或非Chromium场景。

2026-09-21 当前源码已重新打包为 `development-20260921-rc15`：Linux/Windows Master、Daemon、SDK、Web 六个归档 SHA 校验通过；独立包级 smoke 验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、状态备份恢复及兼容 RC14 包回退通过。详见[RC15发行报告](reports/release-rc15-2026-09-21.md)。打包/快照证据不替代 Windows 真机、systemd、远端网络和物理故障验证。

2026-09-21 RC15 当前源码常规真实浏览器回归：Docker-enabled HTTPS 双节点隔离 fixture 上49项串行运行 **49/49 passed**（11.5分钟），含账号/备份/真实分钟调度/Docker/Compose/扩展/桌面/文件/PTY/通知/恢复/工作区；一小时 E08 单独执行，见[RC15浏览器报告](reports/browser-regression-rc15-2026-09-21.md)。这次不验证 Windows 真机、独立systemd、远程高延迟、物理故障或E08以外的平台组合。

2026-09-21 E08优化阶段记录（历史过程态，已由后续完整复验更新）：CPU profile 将高成本定位到恢复模块 JSON 深拷贝，`copy()` 改为原生 `structuredClone` 并保留 JSON/代理回退语义；类型检查、50/50前端单测、构建及27/27恢复相关真实浏览器测试通过。无 profiler 的64.954秒真实混合样本1/1通过，p95 **45.2ms**、3/3565长帧、两个PTY及16MiB校验传输完成。当时准备继续完整一小时测试；最终小时结果为3,604.492秒、p95 **48.1ms**并通过，详见上方[恢复克隆优化小时报告](reports/performance-hour-recovery-copy-2026-09-21.md)，E08不再处于等待复验状态。

2026-09-20 最新现代 Material 3 Expressive 整理：更换抽象壁纸及中性青灰配色，统一公共形状/颜色/状态，整理13个内置应用样式并修复减少动态效果对JS启动器的约束。完整浏览器38/38通过（1.9分钟），最后输入框修正复验3/3通过（10.8秒）；最终构建通过，真实截图34张，13组主要布局无差异，Dock四边14px。见[现代视觉报告](reports/material-modern-2026-09-20.md)。本轮不提升后端和外部平台验收状态，旧rc11未包含最新UI。

2026-09-20 最新 Material 3 细化：依据官方形状/Expressive列表/复选框资料，统一内容区及表头圆角、列表状态、复选框外观，优化四个应用图标配色。构建通过，相关浏览器17/17通过（53.6秒，API替身）；真实本地截图24张及键盘勾选验证通过，13组主要布局无差异，Dock四边14px。见[细化报告](reports/material-refine-2026-09-20.md)。源码/web/dist已更新，本轮未提升后端或外部验证状态，旧rc11未包含最新UI。

2026-09-20 最新 Material You 重构：整套扁平色面、胶囊控件、多色应用图标与几何壁纸落地，保留原布局。最终构建及完整浏览器37/37通过（1.8分钟，API替身）；真实本地截图22张，13组稳定态布局比较无差异，7种Dock场景四边14px、曲线误差小于0.01px。初轮启动器动画采样差异保留并修正后复验，详见[本轮报告](reports/material-you-2026-09-20.md)。源码与web/dist已更新，本轮未提升后端/外部环境验证状态，旧rc11未包含本轮UI。

2026-09-20 最新薄片图标修订：按用户最新近景参考简化图标材质、重绘薄片前景，移除追光/光晕及厚重高光；构建退出0，选定浏览器7/7通过（44.6秒，API替身），真实本地截图22张。7种场景Dock四边均14px，曲线误差小于0.01px。见[薄片图标报告](reports/thin-glass-icons-2026-09-20.md)。源码及web/dist已更新，本轮未提升后端或外部环境验证状态，旧rc11归档未含本轮UI。

2026-09-20 最新圆角对齐：图标和Dock共用连续轮廓，外框等距外扩14px；构建及选定浏览器7/7通过，真实渲染7种场景四边均14px、曲线距离误差小于0.01px，截图5张。见[圆角报告](reports/concentric-corners-2026-09-20.md)。原外部环境验证边界及旧rc11归档状态不变。

2026-09-20 最新轻拟物修订：应用图标改为简洁玻璃叠层，工具图标使用 Lucide，账号菜单及内容字号/间距微调；最终构建及选定浏览器11/11通过，真实本地截图17张。见[图标与密度报告](reports/glass-icons-2026-09-20.md)。本轮未重新验证全部后端/外部环境，旧rc11归档未含本轮UI。

2026-09-20 最新桌面视觉修订：独立三列顶栏、统一工具图标、账号菜单、单色窗口控制及光影/工作区域更新；构建和47单测通过，选定16浏览器场景经15+1修复复验通过，真实本地截图14张。见[本轮报告](reports/modern-desktop-2026-09-20.md)。此轮未重新验证全部后端/外部环境，旧发行归档未含本轮UI。

2026-09-20 原生桌面UI重写（上一版记录）：旧主题与补丁层已替换，图标Dock、具象SVG图标、统一三色窗口控制、实例/文件侧栏列表完成；build、47单测通过，36浏览器场景经35+1修复复验通过，真实本地Master/双Daemon截图11张。见[UI报告](reports/native-ui-2026-09-20.md)。未改变下述外部环境验证边界，旧rc11归档未含此UI。

2026-09-20 当前源码回归：提权环境执行 `make test`（`go test -race ./...`）退出0，所有Go包通过；受限沙箱的IPv6回环/helper失败不计作代码失败。见[全仓竞态报告](reports/full-race-2026-09-20.md)。该证据仍不覆盖Windows真机、独立systemd、远程高延迟和物理掉电。

2026-09-20 续接审计：当前树重新执行 `make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及 47/47 单测退出0。源码扫描未发现生产路径空实现或遗留 TODO；未改变矩阵中外部环境缺口。

2026-09-20 文档证据审计：检查 112 份 Markdown 的相对链接，缺失链接为 0；历史截图已清理的报告改为明确记录“截图未保留”，不再提供失效证据链接。

2026-09-20 当前源码真实 Docker 复验：Engine 29.4.1/overlayfs 与隔离 fixture 可用，`TestRealDockerComposeLifecycle` race 退出0（27.681s）；真实容器/镜像/卷/网络、双流日志、Compose 更新删除、健康失败、卷保留和普通用户拒权均通过，随机带标签对象已清理。见[当前 Docker 验收](reports/docker-current-2026-09-20.md)。远程 Engine、Windows 容器、物理掉电和长期故障组合仍未验证。

2026-09-20 systemd 环境复核：宿主 user bus 无法连接；不挂载宿主 cgroup 的隔离 systemd 容器退出255，特权宿主 cgroup 方案被安全审查拒绝且未执行。未修改宿主服务，真实 systemd service/timer 生命周期仍为环境阻塞，见[systemd 环境报告](reports/systemd-environment-2026-09-20.md)。

2026-09-20 rc8浏览器与发行：前端完整 Chromium 替身回归 36/36（2.5分钟）通过，包含实例批量、传输详情和扩展跨设备修复；六包构建、外部 SHA、归档 MANIFEST 与独立冒烟退出0。见[浏览器回归](reports/browser-regression-2026-09-20.md)和[rc8报告](reports/release-rc8-2026-09-20.md)。

2026-09-20 rc9发行：同步最新验收台账后的六包构建、外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc9报告](reports/release-rc9-2026-09-20.md)。

2026-09-20 rc10发行：续接审计后的当前源码和验收台账重新生成六包；外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc10报告](reports/release-rc10-2026-09-20.md)。

2026-09-20 rc11发行：纳入当前源码 Docker 真实复验和最新台账；六包构建、外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc11报告](reports/release-rc11-2026-09-20.md)。

2026-09-20 构建复验：`make check`、`make windows`、前端类型检查及 47/47 单测均退出0；交叉编译只证明产物可构建，不替代 Windows 真机运行证据。

2026-09-19 F12/E04实际分钟调度：不注入时钟，两个相邻分钟每槽唯一备份任务、源版本更新、两归档分别恢复正文完整通过，1/1（1.6分钟），44800退出0，所属计划与服务已停止。见[实际分钟证据](reports/schedule-wall-clock-2026-09-19.md)。多天墙钟和物理故障仍未验证，不将该子项当整项通过。

2026-09-19 F02/A01多设备与终端现场：真实 HTTPS 两浏览器上下文以不同设备身份竞争同一工作区，旧 revision 被拒绝且本地正文保留；同场景确认布局同步剥离终端检查点、显式内容同步携带终端检查点，`workspaces.spec.ts` 2/2（6.9s，27729）通过。见[多设备冲突证据](reports/workspaces-2026-09-19.md)。断电迁移、存储耗尽和高延迟组合仍未验证。

2026-09-19 E09 Windows运行代码修复：修正多核 Windows Job Object CPU 配额换算，新增跨平台换算边界单测；Linux race、Windows交叉编译和静态检查通过，Windows Job/ConPTY 真机仍受环境阻塞。见[系统适配报告](reports/system-adapters-2026-09-12.md)。

2026-09-20 E09发行验证：包含上述修复的 `development-20260919-rc5` 六包重新打包，SHA256SUMS 与独立包级冒烟（60112退出0）通过；Windows运行仍未验证。见[rc5发行报告](reports/release-rc5-2026-09-20.md)。

2026-09-20 E09发行验证：包含定时备份一致性修复的 `development-20260920-rc6` 六包构建、SHA256SUMS/355项MANIFEST校验和独立包级冒烟（24537退出0）通过；Windows运行仍未验证。见[rc6发行报告](reports/release-rc6-2026-09-20.md)。

2026-09-20 E09发行验证：包含默认备份应用定时一致性 UI 的 `development-20260920-rc7` 六包构建、SHA256SUMS/852项MANIFEST校验和独立包级冒烟（36781退出0）通过；Windows运行仍未验证。见[rc7发行报告](reports/release-rc7-2026-09-20.md)。

2026-09-20 F12/E04定时一致性：Master 调度入口现接受 `save`、`pause`、`stop`、`hooks` 策略并限制固定钩子引用，`files` 仍拒绝钩子字段；真实固定 argv 保存程序由定时触发，前后钩子、资源身份、最新正文归档与任务结果均通过，定向集成 race 退出0（7.625s）。

2026-09-19 E08优化后小时复验：rc3默认速率/无profiler实际通过，8窗口、双PTY、万条目录与循环传输，p95 **46.5ms**、60,528次交互、220轮已验证传输及末轮校验；61次RSS/归档样本有界。保留945.7ms最大延迟及100/207818长帧，不称无卡顿；见[完整小时证据](reports/performance-continuous-2026-09-19.md)。此前50.3ms失败仍保留；Windows和独立远程网络环境缺口不由本机结果覆盖。

2026-09-19 E09发行与恢复：rc3六包外部SHA-256、内部MANIFEST全部文件和依赖许可文本校验通过；独立SDK/示例构建签名、Master初始化/TLS/双Daemon、停机快照恢复分别在rc3和兼容rc2回退路径实际通过，身份/授权/任务回执/文件正文保留，备份后变更消失，详见[rc3报告](reports/release-rc3-2026-09-19.md)与[依赖通知](reports/dependency-notices-2026-09-19.md)。Windows仍仅构建，任意schema降级和物理故障未由此证明。

2026-09-19 E08小时终态：双PTY持续输出、180次传输轮换和末轮目标校验通过，但交互p95 **50.3ms** 超过50ms，整体退出1；完整数据与60次RSS/归档采样见[持续负载报告](reports/performance-continuous-2026-09-19.md)。随后依据CPU采样优化完整恢复快照复制，保持同步日志和ACK时机；类型检查、47项单测、实际恢复故障3/3及隔离终端2/2通过，60秒诊断p95 45.4ms，JSON复制耗时由7.29秒降至2.06秒。见[快照优化](reports/recovery-clone-2026-09-19.md)。仍需优化后的小时复验，不提升E08整体状态。

2026-09-19 F06/A06/A09/E08诊断修复：真实长时断流确认OUTPUT_GAP，前端逐条消费不足；相邻输出现有界合批解析/保护并累计ACK，保留resize顺序和原事件游标。TerminalApp/InstanceConsole非法浏览器关闭码及断流状态提示已修复。构建和突发输出/精确ACK/刷新/缺口提示浏览器2/2（18.3s）通过；高输出压力仍超出消费能力，默认输出速率小时复验正在运行，不提升整体状态。见[终端合批报告](reports/terminal-batching-2026-09-19.md)。

2026-09-19 E08持续负载：30秒循环传输场景通过（p95 42.8ms），修正PTY负载持续时间后一小时首轮在9.3分钟失败：两路浏览器接收连续一分钟无增量，节点PTY和归档仍运行。后续600秒复现及提高输出速率的诊断已结束并保留失败，原因与修复见上条；当前重新运行默认速率小时场景，不能沿用短时结果声称小时通过。只读RSS/归档采样验证观察点内归档轮转且各低于16MiB，完整峰值仍未定。见[持续负载报告](reports/performance-continuous-2026-09-19.md)。

2026-09-19 F13/E05执行目标边界：Linux服务动作仅接受`.service`，计划任务动作仅接受无路径`.timer`；拒绝通过有限服务接口操作其他unit类型或安装外部timer路径。系统适配race1.147s、真实TLS权限1.759s、Daemon拒绝非法目标回执2.688s通过，见[系统适配报告](reports/system-adapters-2026-09-12.md)。真实systemd生命周期仍未验，此生产改动尚未进入rc2。

2026-09-19 F13/E05确认路径增量：真实私有firewalld应用后显式确认、实际Daemon重开仍保留runtime/permanent规则；普通用户403，同键回放200、异键409、终态修订不变。包含原失联回滚的完整场景race32.103s通过，见[私有防火墙报告](reports/firewall-private-2026-09-19.md)。系统服务/计划任务、Windows及生产拓扑仍未由此证明。

2026-09-19 E03/E04补验：固定argv实际保存程序、保存失败补偿及归档恢复最新正文race6.174s；真实Daemon离线/重连、单次补跑及撤权race2.869s；定时备份接受后真实Master/SQLite重开、双节点重连且任务/归档唯一race3.198s。详见[备份调度证据](reports/backup-scheduler-race-2026-09-14.md)。受控时钟不代替长期墙钟，服务重开不代替进程硬崩溃或掉电。

2026-09-19 F13/E05 真实防火墙：私有网络/挂载/用户namespace、独立D-Bus与firewalld中，真实TLS双Daemon应用后阻断测试Master端口，未确认租约按期限恢复runtime/permanent原配置，连接恢复后任务明确FAILED/confirmation_expired_rolled_back；普通账号403，原服务/端口范围保留。race25.998s通过，详见[私有防火墙报告](reports/firewall-private-2026-09-19.md)。未接触宿主防火墙，Windows及系统服务/计划任务真机组合不由此证明。

2026-09-19 A12/E03 Linux真实ENOSPC：私有1MiB tmpfs、真实TLS双Daemon竞态验证跨节点移动、备份创建、备份恢复的空间不足；源和既有目标保留、未完成数据不发布、暂存清理均通过。传输+备份创建6.505s、恢复3.678s，详见[空间不足报告](reports/enospc-2026-09-19.md)。不以tmpfs替代物理掉电、Windows或远程文件系统证据。

2026-09-19 F04/F09/A09 突发任务修复：控制接收队列满时改为有界等待，避免短时存储消费落后导致断链及任务接受记录缺失；协议真实TLS race 1.192s通过，重编译双节点后真实Chromium通知/分页2/2（7.9s），明确核对101任务全部SUCCEEDED。见[控制背压证据](reports/control-backpressure-2026-09-19.md)。长期及跨平台边界仍未验证。

2026-09-19 F04/A05 节点真实浏览器复验：修正测试的应用启动入口后，维护、配额、配置草稿和票据下载定向 1/1（4.7s）；完整节点文件 2/2（53.1s）通过，包含真实 Daemon 暂停后的失联/WAITING_NODE及恢复。详见[全套回归报告](reports/full-browser-regression-2026-09-13.md)。不改变其他平台和长期组合的未验证状态。

2026-09-19 F14/E06 真实取消恢复：私有双节点 HTTPS + Chromium 1/1（4.6s）通过，服务端接受取消后中止浏览器响应，刷新不重放，恢复暂停节点后任务与界面为 CANCELLED，原取消键保留。详见[扩展任务重试](reports/extensions-task-retry-2026-09-19.md)。该项补齐此前取消丢回执/刷新缺口；未要求 WASI 已开始执行，Windows、远程长期与运行中中断组合仍不由此证明。

2026-09-19 F14/E06 查询恢复增量：参考扩展任务 GET 失败后可显式刷新状态，读取请求去重且不重放创建/取消；独立包 Chromium 1/1（8.5s）通过，使用模拟 HTTP 后端。证据见[扩展任务重试](reports/extensions-task-retry-2026-09-19.md)，不改变真实故障组合的未验证状态。

2026-09-19 F14/A13/E06 请求身份增量：SDK 创建/取消任务均显式传递稳定请求键，参考扩展在发送前保存待提交内容和键。创建任务真实 HTTPS 双节点浏览器丢响应、刷新、同键重试 1/1（9.2s）通过；取消回执真实 TLS 集成 7.018s 通过，终态回放不推进修订；取消浏览器恢复使用模拟 HTTP 后端，单场景 1/1（8.7s）。更新后的完整扩展浏览器套件 6/6（20.1s）、前端 47/47、SDK/参考扩展构建通过。证据见[扩展任务重试](reports/extensions-task-retry-2026-09-19.md)。真实节点取消丢回执/刷新组合、Windows 与远程长期场景仍未验证，整项状态不提升。

2026-09-14 F12/E04 调度写入请求回执收口：计划更新与删除的版本 CAS、记录变更和 SQLite 请求回执在同一事务内提交；同键更新/删除可回放，不同载荷返回 `IDEMPOTENCY_CONFLICT`，前端删除保留稳定请求键。调度定向 race、提权全仓普通回归（Master 128.030s）和竞态回归（Master 267.872s、extensions 149.546s）均退出 0，详见[请求键报告](reports/request-keys-2026-09-14.md)。Windows、长期离线与物理故障场景仍按 F12/E04 行保留未验证。

2026-09-14 final46 交付复验：调度回执事务修正及全仓回归证据纳入 `dist/releases/development-20260914-final46`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `18cc8b4e1013c968924f4159eecfeb1925fff78b14e57906461b928989c3b537`，冒烟临时目录 `/tmp/blora-release-smoke-94mpn3me` 已停止进程。

2026-09-14 final47 交付复验：重启级调度回执测试与 OpenAPI 删除契约纳入 `dist/releases/development-20260914-final47`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `24e852ccf0fe2a6cba417ddb45217ca25cae98a75c087e2a1b7107a814af4ca2`，冒烟临时目录 `/tmp/blora-release-smoke-jzdy74x5` 已停止进程。

2026-09-14 F13/E05 默认应用请求键现场：系统服务/计划任务、防火墙应用/确认及进程终止前端均保存稳定请求键，提权 Chromium 系统管理浏览器场景 3/3（14.6s）、前端 46/46、类型检查和生产构建通过；真实危险主机和 Windows 行为仍按矩阵保留未验证。

2026-09-14 final48 交付复验：默认应用请求键修正及调度回执证据纳入 `dist/releases/development-20260914-final48`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `580112f9bd3cc8738162a91fe800fb151bf8ac9c3b5f027e4c890ac63b48aee4`，冒烟临时目录 `/tmp/blora-release-smoke-b1wqtbhw` 已停止进程。

2026-09-14 F09/A09 取消请求幂等收口：任务收据持久化首次 `cancellationRequestId`，普通任务、上传和跨节点传输取消均在事务中记录；并发取消只提交一次，终态重试返回原收据。存储定向取消测试与真实 TLS Master/双 Daemon 集成回归通过，任务中心为每个任务保留稳定请求键，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 F03/F04 节点登记幂等收口：登记票据生成绑定管理员、请求键和节点名称，同键重试只返回原票据，不同名称明确返回 `IDEMPOTENCY_CONFLICT`；存储和真实 TLS 双 Daemon 管理回归通过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 登记回执修正后的全仓复验：提权 `go test -p 1 ./... -count=1` 与 `go test -race -p 1 ./... -count=1` 均退出 0（Master 普通 126.994s、竞态 261.687s），`make check` 与前端 46/46 通过；未留下测试资源。

2026-09-14 final38 交付复验：登记票据幂等修正及全仓证据同步后生成 `dist/releases/development-20260914-final38`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `ce76c1e52dfeb38627f2181c8751cb15048c538f0a025f05c568d7c2ad240628`，冒烟临时目录 `/tmp/blora-release-smoke-rrefkg26` 已停止进程。

2026-09-14 F04 请求键继续收口：节点换钥票据以管理员、节点、请求键和 `reactivate` 绑定持久化回执，同键回放原票据，不同载荷返回 `IDEMPOTENCY_CONFLICT`；真实 TLS 换钥集成及换钥后的代次/旧连接栅栏通过。改动后全仓普通 128.965s、竞态 262.770s 均退出 0，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 final39 交付复验：换钥票据幂等修正及全仓证据同步后生成 `dist/releases/development-20260914-final39`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `e63dd2e0ae2c9fbd25e7fba2abf245fb8027984f23d5c07f13939eedc8647728`，冒烟临时目录 `/tmp/blora-release-smoke-wsor4qlc` 已停止进程。

2026-09-14 final40 交付复验：将 final39 换钥证据与文档同步后生成 `dist/releases/development-20260914-final40`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `1739590e081a8fc043f914e82f5adda7c998e759df1e9fb3fb50a01c787d9fce`，冒烟临时目录 `/tmp/blora-release-smoke-ghsj5ufc` 已停止进程。

2026-09-14 F03/F05/F06/E09 请求键边界再收口：实例创建、实例控制、终端创建和授权写入均在读取请求体前拒绝缺失/非法键；定向回归 6.241s，最新 Master 普通 127.825s、竞态 265.970s 均退出 0，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 final41 交付复验：请求键边界证据及文档同步后生成 `dist/releases/development-20260914-final41`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `37996802de1cc4510380974f8d60cbab69ca9513e9371c952be80dd631c7c286`，冒烟临时目录 `/tmp/blora-release-smoke-rwgm48yh` 已停止进程。

2026-09-14 E02/F11 请求键边界补齐：Compose 项目保存 prepare/chunk 写入在解析正文前要求稳定键，前端按准备键/分片偏移发送；容器管理定向回归 0.377s，前端 46/46，最新 Master 普通 129.598s、竞态 264.964s 均退出 0。需要显式 Docker 镜像的 Compose 真实端到端用例因环境缺失跳过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 final42 交付复验：Compose 请求键修正及最新回归证据同步后生成 `dist/releases/development-20260914-final42`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `b6a8cc499522d141cbe05e11fdd7b1878b9ddc5e17d6f465ff4be9ca7eddf73e`，冒烟临时目录 `/tmp/blora-release-smoke-k7gfk8j5` 已停止进程。

2026-09-14 final43 交付复验：Compose OpenAPI 请求键声明同步后生成 `dist/releases/development-20260914-final43`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `78520382738ab277e285a3cd084fd22b9ea8214d2f1aa802a9e828652fcee24d`，冒烟临时目录 `/tmp/blora-release-smoke-zg38nbet` 已停止进程。

2026-09-14 F04 换钥冲突错误码修正：同键不同载荷明确返回 `IDEMPOTENCY_CONFLICT`（409），真实 TLS 换钥集成回归 0.209s 通过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 final44 交付复验：换钥冲突错误码修正同步后生成 `dist/releases/development-20260914-final44`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `15f56cc3151cc38b24ab0d2a916bff6342a9be27055f008bb259231a68deed70`，冒烟临时目录 `/tmp/blora-release-smoke-nn8w4hzn` 已停止进程。

2026-09-14 F03/F04 管理写入错误码统一：用户、角色、节点设置和换钥等持久化请求键冲突统一返回 `IDEMPOTENCY_CONFLICT`（409）；定向回归 1.198s，最新 Master 普通 127.672s、竞态 263.528s 均退出 0。首次并行回归的 `/tmp` 空间耗尽已清理后复跑通过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 final45 交付复验：管理写入口幂等冲突错误码统一后生成 `dist/releases/development-20260914-final45`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `e332a19b9c80ea6fdecdfc1638eaca3583dd547ac3e8aa787203439dec120195`，冒烟临时目录 `/tmp/blora-release-smoke-8pjxh2wo` 已停止进程。

2026-09-14 取消收据改动后全仓复验：提权 `go test -p 1 ./... -count=1` 退出 0（Master 124.958s）；完整 `go test -race -p 1 ./... -count=1` 复跑退出 0（Master 255.904s、extensions 137.279s、terminal 18.616s、runlog 2.136s），`make check` 与前端 46/46 通过。首轮竞态的 runlog `/proc` 时序失败已隔离重跑通过并如实保留。

2026-09-14 F03/F04/F06/E09 请求键补齐：用户创建/改密、节点登记/身份轮换、云端工作区 PUT/DELETE 及通用任务/上传取消在授权后统一拒绝缺失或非法 `Idempotency-Key`；用户与工作区事务写入支持稳定回放。Master/Storage 定向 race、传输取消回归、`make check` 与 OpenAPI 3.1（103 路径、120 操作）解析通过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 防火墙确认幂等收口：确认请求键传入并持久化于 Daemon 租约和成功任务结果；同键在租约清理前后均可安全回放，其他键明确冲突且不重复触碰主机规则。Daemon 定向 race 29.354s、Master 定向 race 2.096s、改动后串行全仓普通与 race 均退出 0；详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 防火墙确认 CAS 并发补验：Daemon 重启后无内存租约的双键并发确认仅一个成功，获胜键可回放，另一键返回 `FIREWALL_LEASE_CONFLICT`；定向 race 1.137s、改动后串行全仓普通与 race 均退出 0，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 当前源码串行全仓复验：`go test -p 1 ./... -count=1` 明确退出 0，Master 126.349s，runlog、runtime、storage、terminal、extensions 及其余 Go 包全部通过；不把此前并行时序失败记录改写为无条件全绿。

2026-09-14 当前源码串行全仓竞态复验：`go test -race -p 1 ./... -count=1` 明确退出 0，Master、extensions、terminal、runlog、runtime、storage 及其余 Go 包全部通过；Windows、远程节点与物理故障仍按矩阵保留未验证状态。

2026-09-14 文件请求键边界收口：文件保存、文件动作、新建上传均在实例写权限确认后、请求体解析前统一检查 `Idempotency-Key`；空白保存键与 129 字节动作键集成回归及竞态回归（5.406s）通过，完整 Master 126.836s 与 `make check` 退出 0，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 上传取消测试契约同步：两个历史集成用例补充稳定请求键；上传取消定向回归 1.562s、完整 `go test ./internal/master -count=1` 125.538s 退出 0。

2026-09-14 改密同键回放修正：当前密码只在首次事务回调校验，撤销会话后重新登录仍能复用原回执；改密定向回归 1.044s、完整 Master 套件 127.375s 退出 0。

2026-09-14 final22 交付复验：提权环境完整 Master suite（125.544s）及其余 Go 包通过；全仓首轮仅因备份用例一次 `terminal access denied` 失败，单独复验和后续 Master 重跑通过。前端 46/46 单测与 `make check` 通过；`development-20260914-final22` 打包、独立冒烟及六个归档校验通过，见[发行包报告](reports/package-2026-09-13.md)。

2026-09-14 final23 文档同步交付：请求键 race 最终时间写入证据后重新生成 `development-20260914-final23`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `7bab5ab5c69dda7cc0fef737ee7885af9b7caaebec3c9fc883ee5164d2eed2cd`）。

2026-09-14 final24 记录同步交付：修正全仓测试一次环境波动的说明后重新生成 `development-20260914-final24`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `ed434701601b269098053b261a69ce0e785aea7d6021f971367839424996cefc`）。

2026-09-14 final25 取消请求键交付：通用任务与上传取消新增授权后请求键检查，传输取消回归通过；重新生成 `development-20260914-final25`，打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `466f6d3e6a2da7c3ab37beb389ca45e3c4dae5f87785cbf0788b738a8311a5fa`）。

2026-09-14 final26 交付复验：同步上传取消历史用例请求键后，完整 Master 套件 125.538s 通过；重新生成 `development-20260914-final26`，打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `272c3f3bef10d4a380394ba23d2f99692b2328af4f63a12378ae79958f8a6e55`）。

2026-09-14 final27 文档归档复验：将 final26 记录纳入发行包后重新生成 `development-20260914-final27`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `a89541bb7657d7174d454828a7cea77477c9aaa73649b37ff85d55fe10f5faa`）。

2026-09-14 final28 改密幂等交付：自助改密回执在会话撤销后可同键重放；改密定向与完整 Master 套件通过。`development-20260914-final28` 打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `efc9e4dda49e234b47d2037990331f726af7b674094b62667aa941c4c0c4d030`）。

同日收尾复验：`go test ./internal/storage -count=1` 0.579s、`make check`（`go vet ./...`）均退出 0；调试残留及活动测试资源检查为空。

2026-09-14 final29 归档：将最终改密收尾证据纳入包内后生成 `development-20260914-final29`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `12171c6ed4aa430da966bb63e1b6ca3b20767985df2871667f17b36fba63af2a`）。

2026-09-14 final30 归档：文件保存、文件动作、新建上传入口的请求键边界收口后生成 `development-20260914-final30`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `f7630947960072aded20d64f7a64397c8737bf125599f28fd8f661d143a5802d`）。

2026-09-14 final31 归档：将文件请求键竞态回归证据纳入包内后生成 `development-20260914-final31`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `18ff79189be48c49c93bb8c0a5330827472726f8f242b54eb7a44c6c28db3e80`）。

2026-09-14 final32 归档：将串行全仓 Go 回归证据纳入包内后生成 `development-20260914-final32`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `06e336ee2ec13240f2d7b86f6c83b638d5ec42db8cf2d9b0de35289aaae25c96`）。

2026-09-14 final33 归档：将串行全仓 Go 竞态回归证据纳入包内后生成 `development-20260914-final33`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `84931b88163ef6182ca4504b9d26105b427c4c16fe569f783c2807223abbe34e`）。

2026-09-14 final34 归档：防火墙确认请求键持久化与终态回放修正后生成 `development-20260914-final34`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `83dcf17fc84253c0e8731056497a7b0f83b4e2eb5f1768ede6d033260dfccd15`）。

2026-09-14 final35 归档：重启后无内存租约的确认 CAS 并发修正及回归证据纳入 `development-20260914-final35`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `de8c519f6af02db48e41129150d41cef7805bd4bdeaa14b08db99ec3bcb1dd4d`）。

2026-09-14 final36 归档：任务取消收据幂等修正及最新全仓普通/竞态证据纳入 `development-20260914-final36`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `50fb039b87dc00bef14d1006f0e6ebcf929f407a1790efe51426d0d77ff80913`）。

2026-09-14 final37 归档：将 final36 进度/矩阵记录同步进包后生成 `development-20260914-final37`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `4fbae56157ee6ed406d6f8deb91e584be61fe44b130f05d16669a42e691a59a3`）。

2026-09-14 收尾资源检查：无 Blora/fixture/Vite/Playwright 活动进程，Docker `blora.test=1` 标签对象为空；这只证明本轮测试清理完成，不改变 Windows 真机、远程节点/Engine、物理故障、长期调度和跨平台组合的未验证状态。

2026-09-14 最新普通全仓回归：Master（127.270s）及其余 Go 包通过；`internal/runlog/TestStdinSurvivesDaemonAndOutputEOF` 首次并行时序出现空输入，单独重跑 0.052s 通过，未改写为全仓无条件通过。

2026-09-14 请求键边界收口：统一入口拒绝空白及超过128字节的 `Idempotency-Key`，实例控制在动作解析、资源授权和节点门禁后、任务接纳前检查；后端定向回归、前端类型检查与 46/46 单测通过。`development-20260914-final10` 发行包构建及独立冒烟通过（SDK/参考扩展离线构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE）。

2026-09-14 最终 race 复验：清理任务缓存后提权环境 `go test -race -p 1 ./...` 退出 0（Master 251.165s、extensions 136.422s、terminal 18.729s，其余 Go 包通过）；此前一次 `/tmp` 空间耗尽的复跑保留为环境失败记录。

2026-09-14 传输步进优化后 race 复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-race-final14 GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 退出 0（Master 248.686s、extensions 136.133s、terminal 18.695s，其余 Go 包通过）。

2026-09-14 幂等键边界回归：`TestRequireRequestIDBounds` 覆盖缺失、空白、129 字节拒绝和128字节上限接受；`development-20260914-final11` 发行包构建及独立冒烟通过。

2026-09-14 授权顺序修正后的 E09 复验：`development-20260914-final12` 发行包构建及独立冒烟通过；实例控制保持动作解析和资源授权优先，再执行请求键与任务接纳检查。
2026-09-14 A09 长时采样后的 E09 发行包复验：`development-20260914-final13` 构建及独立冒烟通过，独立 SDK/参考扩展、Master 初始化/TLS/登录和双 Daemon ONLINE 均完成；发行包校验见 `dist/releases/development-20260914-final13/SHA256SUMS`。
2026-09-14 A13/E06 浏览器权限复验：真实 `extensions.spec.ts` 完整套件 6/6（22.6s）通过，新增成员上下文授权边界场景同时通过；fixture 已停止并完成清理。

2026-09-14 E09 传输优化后发行包复验：`development-20260914-final14` 构建及独立冒烟通过；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 完成，校验见 `dist/releases/development-20260914-final14/SHA256SUMS`。

2026-09-14 E09 文档同步发行包复验：因 README/docs 更新重新生成 `development-20260914-final15`；独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c` 均通过，校验见 `dist/releases/development-20260914-final15/SHA256SUMS`。

2026-09-14 E09 长时性能证据发行包复验：加入 `BLORA_PERF_SOAK_SECONDS` 说明与60秒采样证据后生成 `development-20260914-final16`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final16/SHA256SUMS`。

2026-09-14 E09 采样超时预算修正发行包复验：`development-20260914-final17` 构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final17/SHA256SUMS`。

2026-09-14 E09 Windows 适配编译复验：`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c` 分别编译 `internal/runtime`、`internal/runlog`、`internal/systeminfo` 测试入口，均退出0；当前环境无 Wine/Windows 真机，运行行为仍未验证。

2026-09-14 E09 交叉编译证据发行包复验：`development-20260914-final18` 构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final18/SHA256SUMS`。

2026-09-14 E03/E04 备份与调度竞态复验：提权隔离环境 `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（backup 1.856s、Master 5.816s），固定 argv Hook `go test -race ./internal/daemon -run 'TestBackupHook' -count=1` 退出0（2.039s）；真实 Chromium 备份/恢复/调度入口 1/1、3.5s 通过；长期离线、具体应用 hook 语义和物理 ENOSPC/掉电仍保留，见[浏览器证据](reports/backup-browser-2026-09-14.md)。

2026-09-14 E09 最新交付包：因备份/调度证据更新生成 `development-20260914-final19`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `7e365582aad6a03cc84e1d7b7342ce795da2f903036f0db059f37875385bc9b2`。

2026-09-14 E09 最新交付包：因真实备份浏览器证据更新生成 `development-20260914-final20`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `4e25104368473ca06daaae78870b9ee8b015e5e0ebbda665ad26be116a45c60b`。

2026-09-14 E09 交付包收口：将备份/调度报告纳入包内后生成 `development-20260914-final21`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `9cb7f53ce64e6703003054bdcce2af95a7b5c392b59b0de9debde1e9b2e5282d`。

2026-09-14 F03/F13/F14 请求键边界修复：扩展安装/升级/回滚/启停/卸载、目录安装和防火墙租约确认均在授权后统一拒绝无效 `Idempotency-Key`；Master 定向 race 77.219s、OpenAPI 3.1 解析通过，详见[请求键报告](reports/request-keys-2026-09-14.md)。

2026-09-14 E08 批量步进复验：重编译夹具纳入 `transferChunksPerStep=16` 后，真实 Chromium 混合场景 1/1 通过（fixture-1569035732，约2.0分钟），512MiB 用时110.868s（约4.62MiB/s），反馈p95 39.0ms、0/285长帧、目标校验成功；此前旧夹具133.246s仅作索引基线；长期内存/远程网络/Windows及物理故障仍保留。

2026-09-14 E08 长时采样：`BLORA_PERF_SOAK_SECONDS=60` 的真实 Chromium 混合场景 1/1 通过（fixture-2822501719），采样65.697s、p95 43.3ms、1/3770长帧、堆34.6–97.9MB、未确认队列峰值86,867B，512MiB传输109.245s；小时级稳定性、远程网络、Windows及物理故障仍保留。

2026-09-14 增量：扩展跨设备副本场景 1/1 通过（Linux Chromium，状态版本 1→2）；恢复校准返回规范资源身份，删除/撤权资源保留窗口与草稿。详见[扩展跨设备报告](reports/extensions-cross-device-2026-09-14.md)。该证据不覆盖 Windows、远程高延迟和物理故障；未授权浏览器组合另见[权限报告](reports/extensions-unauthorized-browser-2026-09-14.md)。

2026-09-14 回归记录：最新一次完整 `make test` 在并行资源压力下未通过（Daemon 防火墙命令、Master 审计登录、runtime 后代条件、PTY 输出各一项时序失败）；四项随后低并发定向 race 均通过，不能将该次全仓命令记为通过。此前全仓通过证据仍保留其对应代码时点。

2026-09-14 race 复验：改用 `go test -race -p 1 ./...` 串行执行，退出 0（Master 259.481s，其余包通过）；该结果与上条并行压力失败记录同时保留。

2026-09-14 发行包与请求键收口：调度创建/更新入口统一拒绝空白 `Idempotency-Key`；`development-20260914-final7` 打包及独立冒烟均通过，包含 SDK/参考扩展离线构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE。

2026-09-14 调度请求边界：请求键校验提前到请求体解析前，定向 race 通过；`development-20260914-final8` 打包及独立冒烟均通过，产物包含该最终源码。

2026-09-14 持久化修改请求键复验：实例/节点设置、用户/角色变更、节点撤销、扩展数据/资源写入、备份计划/恢复和传输取消均在授权后统一拒绝空白 `Idempotency-Key`，相关 Master 集成和提权串行全仓 race（Master 252.584s）通过；OpenAPI 同步声明对应参数。

2026-09-14 E09 增量：统一请求键改动后的 `development-20260914-final9` 发行包构建与独立冒烟通过；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功。

2026-09-14 A09 并行流复验：`TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop` 在真实 TLS Master/双 Daemon 下 race 通过（13.07s）；慢日志未消费峰值 82,635B、接收队列 79,507B，8.4MB 跨节点传输与停止并行，停止和归档核验成功。长时内存采样仍未覆盖，A09 继续保留进行中。

2026-09-13 A16完整入口身份复验：真实目标改名、管理员停止资源墓碑删除/同名替换后旧resourceRef显示不可访问且不指向新资源1/1（7.0s），双账号撤销instance.read后旧入口失效1/1（3.3s），暂停Daemon时入口显示节点失联并恢复原runId1/1（50.4s）；结合专用窗口复用/显式新开和入口改名删除1/1（3.6s），Linux Chromium覆盖全部A16，见[入口报告](reports/shortcut-lifecycle-2026-09-13.md)。

2026-09-13 E06/E07/A13真实扩展复验：Linux Chromium HTTPS fixture 上独立扩展全套 5/5（24.2s）通过，覆盖通知来源与刷新去重、元数据草稿与真实实例写入、v1→v2 数据迁移/不兼容拒绝/失败保留/回滚、真实节点 WASI 任务与资源窗口/快捷入口/移窗/关闭、本地注册源 v1.0.0/v1.1.0 安装升级及禁用门禁；见[扩展全链路报告](reports/extensions-real-2026-09-13.md)。Windows 真机、远程高延迟源和物理故障仍保留。

2026-09-13 E09最终发行包复验：加入实例删除入口后使用不可变标签 `development-20260913-final` 重新生成六个归档；独立解包冒烟在受限提升环境 1/1 通过，SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，见[发行包报告](reports/package-2026-09-13.md)。Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-13 E09发行包最终复验：监控 stale 诊断和前端提示更新后使用不可变标签 `development-20260913-final3` 重新生成六个归档；独立解包冒烟在受限提升环境 1/1 通过，SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，见[发行包报告](reports/package-2026-09-13.md)。Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-13 系统管理能力门禁复验：系统服务、计划任务和防火墙操作按钮现在随节点能力探测禁用，能力不可用时不再允许提交；前端 check/test/build、系统管理浏览器 2/2（54.2s，分别覆盖能力可用与不可用）及 `development-20260913-final4` 独立发行包冒烟通过。危险主机、Windows 真机及远程故障组合仍保留。

2026-09-13 任务接纳边界：主机终端、容器/Compose、系统管理、进程终止、实例终端、扩展任务和跨节点传输入口统一在授权后要求 `Idempotency-Key`，缺失时返回 `REQUEST_ID_REQUIRED`，防止空键请求先触发远端检查或产生不可对账任务；定向 Master 集成回归退出 0。

2026-09-13 控制台输入幂等键补齐：实例日志控制台输入也在授权后、读取运行代次前要求 `Idempotency-Key`，并同步 OpenAPI 必填参数；定向测试覆盖空键 400、合法请求单次交付和独立停止输入，退出 0（1.620s）。

2026-09-13 E09发行包收口：控制台输入请求键边界加入后，以不可变标签 `development-20260913-final5` 重新生成六个归档并完成独立解包冒烟；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-14 全仓竞态复验：控制台输入边界加入后，提权 `make test`（`go test -race ./...`）退出 0，Master 257.810s，其余 Go 包通过；Windows 真机、物理 ENOSPC/掉电、远程高延迟和长期负载仍按各项保留。

2026-09-13 系统能力反馈：SystemApp 按节点能力探测停止不可用服务/计划任务轮询，并对防火墙不可用状态给出原因、阻止预览/应用提交；可用能力路径及系统管理浏览器 1/1 保持通过。

2026-09-13 日志流边界修复：慢消费者超过 45 秒时，Master 在 `SLOW_CONSUMER` 错误帧成功写入后保留 100ms 有界交付窗口，避免客户端只收到 EOF；慢消费者游标重连测试连续 2 次通过，race 复验亦通过。

2026-09-13 云端工作区冲突提示：同步遇到 `WORKSPACE_CONFLICT` 时前端保留本地现场并提示用户刷新列表后选择副本，不覆盖草稿或误报旧云端版本成功。

2026-09-13 A05任务等待增量：真实离线节点提交stop后原任务WAITING_NODE，生产任务卡显示等待节点；恢复后同taskId成功且实例实际停止，浏览器1/1（50.7s）通过，见[失联报告](reports/node-offline-browser-2026-09-13.md)。

2026-09-13 A05浏览器增量：暂停精确识别的fixture Daemon至真实心跳失联，页面显示等待确认而非实例停止，恢复上线runId不变，1/1（50.5s），见[失联浏览器报告](reports/node-offline-browser-2026-09-13.md)。

2026-09-13 A10任务中途增量：独立Daemon实际STOPPING阶段SIGKILL后，原任务INTERRUPTED并保留诊断阶段，同key重试不重放，原runId/启动计数保留，脚本退出0，见[Daemon崩溃报告](reports/daemon-crash-2026-09-13.md)。

2026-09-13 A10增量：独立Daemon SIGKILL后同身份目录启动，新startupId恢复原runId、启动计数不增加，随后可重启原运行；独立包脚本退出0，任务中途Daemon崩溃仍待补，见[Daemon崩溃报告](reports/daemon-crash-2026-09-13.md)。

2026-09-13 A04硬中断增量：独立解包Master在restart RUNNING时SIGKILL，原状态重启后同requestId复用taskId、任务成功且实际启动计数仅初次+一次重启；脚本退出0，见[重启恢复报告](reports/restart-disconnect-2026-09-13.md)。

2026-09-13 A04/A05增量：实际重启任务RUNNING时断开源节点管理连接，新generation重连后同requestId仍复用原taskId，原任务成功、后续重连不改变runId；定向race7.392s通过，见[重启断线报告](reports/restart-disconnect-2026-09-13.md)。Master进程中断等组合仍保留。

2026-09-13 A02增量：真实双编辑窗口共享正文、另一账号修改服务器版本、冲突拒绝/刷新保留/显式采用基线保存1/1（7.7s）通过；独立光标滚动及迟到响应组合仍待补，见[共享编辑报告](reports/shared-editor-conflict-2026-09-13.md)。

2026-09-13 A08增量：真实成员活跃PTY撤销input/read，输入租约失效、真实文件未增加副作用、新请求403、刷新旧快照仍拒权，1/1（6.1s）通过；见[浏览器撤权报告](reports/browser-revoke-2026-09-13.md)。

2026-09-13 A09并行增量：真实慢WSS日志读取端、8.4MB跨节点传输及停止控制同时运行，停止在5秒预算内完成、归档可重读且传输目标逐字节验证通过；定向race14.120s，长期内存/归档预算仍待量化，见[混合流报告](reports/mixed-stream-control-2026-09-13.md)。

2026-09-13 A17编辑器增量：实际鼠标拖出/合并并立即刷新，视图ID、归属、顺序、中文未保存正文及撤销重做保持，1/1（3.8s）通过；活跃PTY同组合仍待补，见[拖拽报告](reports/drag-recovery-2026-09-13.md)。

2026-09-13 A11 增量：真实桌面配额写入失败时最新草稿导出、存储恢复后刷新、旧schema0迁移及未知schema999原记录跨刷新保留，2/2通过（4.6s）；快照事务中断组合仍待补，见[恢复故障报告](reports/recovery-quota-2026-09-13.md)。

2026-09-13 A06 增量：真实 vim 中文未保存缓冲区刷新后保存准确、top 刷新与退出、WSS 输入不重放 1/1 通过（9.0s），见[交互程序报告](reports/terminal-programs-2026-09-13.md)。E09 解包 SDK 离线干净安装/构建/签名及 Master/双 Daemon 独立上线也通过，见发行包报告；Windows 真机仍未验证。

2026-09-13 E09 交付增量：六个本地发行包构建、重复打包逐字节一致及归档清单/摘要核验通过；解包独立运行和 Windows 真机仍待验证，见[发行包报告](reports/package-2026-09-13.md)。

2026-09-13 A07 Linux真实场景通过：解压与16MiB本机上传并行关闭标签，解压10,000文件独立完成；新标签从任务中心恢复原上传，于983,040B继续并验证成功。新增账号/请求/资源/文件身份保护，前端46单元与构建通过，见[关闭网页证据](reports/upload-close-2026-09-13.md)。

2026-09-13 A07/E08最终构建复验：带 `--performance` fixture 的真实关闭网页/上传续传 1/1（1.4m），原 uploadId 从917,504B续传、16MiB摘要和10,000项解压均核对；混合负载 1/1（27.8s），p95 36.2ms、0/298长帧、八窗口/双PTY/万条目录/16MiB传输全程并行，见[关闭网页证据](reports/upload-close-2026-09-13.md)与[真实性能报告](reports/performance-real-2026-09-13.md)。

2026-09-13 F07/A12/E08大文件增量：512MiB真实跨节点传输和独立摘要验证通过，约0.942MiB/s；同一后台任务跨浏览器中断续接，采样p95 42.4ms。源前缀改写并恢复mtime在发布前拒绝，实际源Daemon重启重建索引通过，见[传输证据](reports/transfer-proof-2026-09-13.md)。新增[数据库版本保护与运维文档](reports/upgrade-storage-2026-09-13.md)；完整存储race通过，物理掉电和Windows真机仍保留。

2026-09-13 E08短时真实场景通过：最终p95 38.8ms、0/302长帧，八窗口/双PTY/万条目录/16MiB跨节点传输全程并行；终端增量恢复和软件GPU降级真实复验通过。大文件吞吐及长期/远程网络仍待补，整项保持进行中，见[最终证据与失败历史](reports/performance-real-2026-09-13.md)。

2026-09-13 E08真实负载未通过：8窗口/2PTY/万条目录/16MiB真实传输已运行，p95 278.2ms、303/319长帧，传输成功但未覆盖全采样区间；恢复写入合并及9项恢复测试通过，仍需终端输出增量化与大文件性能优化，见[失败证据](reports/performance-real-2026-09-13.md)。

2026-09-13 F10/E01增量：原生实例CPU/RSS聚合当前运行进程树，出生身份核对和成员变化重采样已接入；真实忙子进程、FIFO新增成员、TLS链路回归与构建通过。Windows真机、短命成员历史累计和E08全负载仍保留，见[进程树证据](reports/process-tree-metrics-2026-09-13.md)。

2026-09-13 F10/E01增量：Linux原生主进程CPU真实双采样、出生身份与代次基线保护、CPU计算口径标识已接入并通过真实耗时测试和节点回归。fixture信号清理实际通过。整个运行进程树聚合、Windows真机和E08完整负载仍待补，见[证据](reports/native-cpu-2026-09-13.md)。

2026-09-13 F10/E01/E08增量：网络累计量展示、原生未采集项不可用标识、后台降频/可见恢复已实现；真实Docker浏览器29.9s和构建通过。Linux原生CPU尚待采集，完整并发负载、fixture新清理流程和一次第二账号界面登录超时仍未解决或未验证，见[可见性证据](reports/monitor-visibility-2026-09-13.md)。

2026-09-13 F10/E01缓存增量：指标采样后复核身份/权限，旧值仅匹配当前RunID。真实TLS断开/重建Daemon、断链撤权和新代次无缓存拒绝通过，监控race5.043s及构建通过；物理断网/采样中精确竞态与跨平台负载仍保留，见[指标证据](reports/container-metrics-2026-09-13.md)。

2026-09-13 F10/E01端到端：实例监控标签已接真实指标，实例授权范围不请求主机信息；修复新节点首次容器启动目录初始化。实际双Daemon/Docker浏览器1/1验证读权限、撤权、CPU/内存、刷新与停止后旧值，runtime/daemon回归及构建通过。真实断链/跨代次缓存仍待验证，见[容器指标证据](reports/container-metrics-2026-09-13.md)，整项仍进行中。

2026-09-13 F10/E01增量：容器实例指标从固定不可用改为实际Docker采样，身份/启动时间复核、CPU与内存及有界网络计数已实现；HTTP边界race和真实隔离容器指标通过。Master/Daemon与浏览器新增组合仍待验证，见[容器指标证据](reports/container-metrics-2026-09-13.md)，整项保持进行中。

2026-09-13 F09/F11/A10/E02增量：Linux Compose CLI持久归属、遗留进程恢复及未确认时禁止新变更已实现；执行器SIGKILL、启动拒绝、容器/Daemon race及真实Docker生命周期通过。完整Daemon重连组合和Windows真机仍未验证，见[恢复证据](reports/compose-cli-recovery-2026-09-13.md)，不计整项完成。

2026-09-12 F09/F11/A10/E02增量：容器执行器SIGKILL后输出检查点保留、任务不明且不重放，真实进程race通过；Docker浏览器输出和命令退出标识1/1通过。Linux遗留Compose CLI的Daemon启动恢复仍有代码缺口，完整A10未通过，见[输出证据](reports/compose-output-2026-09-12.md)。

2026-09-12 F09/F11/E02增量：Compose运行中每250ms有变化才持久化有界输出，退出标识与任务资源结果分开；实际CLI未退出时跨管理器读取、取消及目录写入故障测试通过，容器race与构建通过。Daemon硬崩溃、Windows真机及完整组合仍未验证，见[输出证据](reports/compose-output-2026-09-12.md)，整项保持进行中。

2026-09-12 扩展resource.write实例元数据分支已接SDK/API/CAS幂等和独立授权，transport与构建通过；真实独立示例浏览器、通知和其余资源能力仍待补，E06保持进行中。见 [元数据证据](reports/extensions-metadata-2026-09-12.md)。

2026-09-12 计划任务分页已贯通Linux/Windows、RPC/API和界面；205项三页、参数TLS与浏览器恢复子项通过，Windows COM分页真机仍未验证，见 [系统报告](reports/system-adapters-2026-09-12.md)。

2026-09-12 系统服务名称分页已贯通Linux/Windows适配、RPC/API与界面，游标刷新恢复；模块race、替身浏览器和构建通过。[系统报告](reports/system-adapters-2026-09-12.md)保留计划任务分页、Windows真机与完整性能场景缺口。

2026-09-12 Windows服务操作已改原生SCM并确认STOPPED/RUNNING，30秒截止与取消检查；名称边界race和构建通过，真机状态组合仍未验证。[系统报告](reports/system-adapters-2026-09-12.md)记录细项与分页缺口。

2026-09-12 Windows计划任务已换为固定脚本/COM结构化读写，支持完整Unicode路径并回读Enabled；路径/输出边界race与交叉编译通过，真机COM执行及完整分页未验证或未完成。见 [系统适配](reports/system-adapters-2026-09-12.md)。

2026-09-12 系统适配：Windows服务SCM查询及只读真机入口、Linux.timer身份与操作边界已接入；Linux race及Windows交叉编译通过，Windows真机/计划任务结构化适配/完整分页仍保留。见 [系统报告](reports/system-adapters-2026-09-12.md)。

2026-09-12 旧单份防火墙Apply/Rollback及reload路径已删除，错误/取消/恢复失败测试迁移到Snapshot并通过race；含Daemon命令子进程回归29.595s。[快照证据](reports/firewall-snapshots-2026-09-12.md)已更新，真实危险环境和Windows仍保留。

2026-09-12 Daemon防火墙命令进程链路：独立子进程测试覆盖生产Snapshot路径的摘要拒绝、持久快照、取消恢复与终态，race28.946s通过；firewalld状态仍为测试替身，真实规则和失联环境不计通过。见 [快照证据](reports/firewall-snapshots-2026-09-12.md)。

2026-09-12 双份预览补充：运行/永久差异与区域已显示，提交摘要绑定双快照和目标，执行前重新核对。定向race、浏览器和构建证据见 [快照报告](reports/firewall-snapshots-2026-09-12.md)。真实命令/防火墙环境与旧兼容清理仍待补，E05不计整项完成。

2026-09-12 新防火墙租约已接入固定区域、运行/永久双快照与分别应用/恢复，新生产路径无全局reload；不同原配置、部分失败恢复和外部冲突的状态化命令替身race通过。[快照证据](reports/firewall-snapshots-2026-09-12.md)说明旧租约、预览与真实环境缺口，E05仍进行中。

2026-09-12 防火墙预览/界面增量：实际受管端口差异、任务排队与确认阶段、原节点绑定及刷新恢复已有代码和定向race/浏览器证据；运行/永久规则混用仍待修复，E05保持进行中。见 [证据](reports/firewall-preview-2026-09-12.md)。

2026-09-12 回滚补充：规则恢复完成持久记录、任务终态后清理、重启不重放已恢复规则以及应用失败保留证据已接入；SQLite任务更新故障注入与租约race通过。[防火墙证据](reports/firewall-confirm-2026-09-12.md)已补充，真实规则合同与环境缺口仍保留。

2026-09-12 防火墙确认成功路径改为先写持久任务终态、再清理恢复记录；SQLite故障注入与租约定向race通过。真实防火墙环境、回滚终态故障组合仍未完成，E05保持进行中，见 [证据](reports/firewall-confirm-2026-09-12.md)。

2026-09-12 监控增量：Windows 原生采集/进程及真机入口、Linux pidfd核对并确认退出、完整枚举搜索与PID分页、确认目标和视图恢复已接入；Linux race、Windows构建和浏览器子项通过，真机与跨处理器组CPU等缺口保留。见 [监控证据](reports/monitoring-2026-09-12.md)。扩展迁移已移到锁外并按包/数据/令牌条件提交，授权不阻塞及并发写冲突定向race通过；E07仍不计整项通过。

版本：v1.0 · 2026-09-09  
当前基线：完整实施已开始；部分核心与桌面子场景有证据，整项验收尚未完成。

2026-09-12 复核更正：B08/B09 尚有代码缺口，不能全部归因为环境缺失。主机任务分发和扩展任务查看已修复并通过真实 TLS 子项；独立 SDK 示例正在补实际浏览器包和界面，后台扩展任务仍仅回传输入，完整受限后台执行、资源桥接和升级状态迁移未完成。此前 fixture 演示包不证明独立 SDK 示例验收。见 [复核报告](reports/resume-2026-09-12.md)。

同日增量：独立包浏览器界面与资源摘要桥接已实现，真实 TLS 独立包安装/资源授权和浏览器资源窗口/笔记独立分别验证通过；有界 WASI 后台执行器已通过真实 guest 测试但尚未接入生产任务，不能计作后台扩展全链路通过。完整资源写入/任务能力、状态迁移及原环境缺口继续保留。

后续 WASI 链路增量：组合包、包/模块摘要绑定、批量通道分块模块获取已接入生产任务，原输入回传代码已替换。独立包真实 TLS 计算与浏览器完整包上传分别有证据；首次全仓 race 暴露冷编译限时与测试等待问题，修复后正在复验。真实浏览器后台任务结果、完整迁移/升级故障和其余能力边界仍未完成，不能提升 E06/E07 整项状态。

上述复验已结束：独立包真实浏览器上传、节点窗口和 WASI 计算 1/1 通过（9.4s）；最终全仓 race 通过（Master 191.796s）。E06/E07 仍保留完整状态迁移、生命周期并发/失败恢复及其余能力缺口；详细命令与失败修复过程见同日复核报告。

后续生命周期增量：公共读写锁、一致包快照、损坏回滚前置校验、禁用状态保持和依赖门禁已通过定向 race；视图迁移在 opaque iframe 执行，宿主条件提交状态/版本，浏览器成功/失败保留场景已通过。尚缺包跨文件写入的掉电恢复、后端用户数据迁移、真实两版本部署及完整能力组合，E06/E07 继续进行中。见 [生命周期报告](reports/extensions-lifecycle-2026-09-12.md)。

注册表恢复增量：跨文件撤销记录、持久提交令牌和启动恢复已实现，实际子进程在多阶段退出的恢复测试通过；Windows 写穿替换/删除及长路径代码交叉构建通过。物理断电、ENOSPC、Windows 真机耐久性、后端数据迁移及其余 E06/E07 组合仍未验证或未完成。见 [恢复报告](reports/extensions-recovery-2026-09-12.md)。

## 1. 范围与使用方法

2026-09-12 用户数据增量：按用户隔离的 JSON 读写/CAS/持久回执、data.read/write SDK、WASI 数据版本迁移和包/数据共同事务恢复已实现。真实独立三版本浏览器迁移/失败保留/回滚 1/1（11.7s）、双用户 TLS、定向 race、实际子进程中断恢复及并发/配额保护通过。物理故障、Windows 真机、其余 SDK 能力及完整组合仍保留，E06/E07 不提升为整项通过。见 [用户数据报告](reports/extensions-data-2026-09-12.md)。

桌面 SDK 增量：受控自身标签移窗/关闭、当前资源快捷入口和独立包 0.2.0 已实现，真实双节点浏览器组合通过（9.6s），前端 32/32。首轮替身浏览器入口缺失未复现，完整资源/通知能力与后端数据迁移仍继续。见 [桌面 SDK 报告](reports/extensions-desktop-2026-09-12.md)。

任务审计增量：真实 Daemon 终态回读及派发前取消补充事务审计，故障注入/去重与真实 TLS 验证通过，F03/F09 整项仍进行中。见 [审计报告](reports/task-audit-2026-09-12.md)。

清单签名增量：发布者认证现覆盖清单全部受支持字段与包体摘要，旧的包体单独签名拒绝；签名工具、重启后验证、存储清单/能力篡改拒绝及真实 TLS 回归通过。E07 仍待后端数据迁移和完整组合，不提升整项状态。见 [签名报告](reports/extensions-signature-2026-09-12.md)。

2026-09-12 扩展任务增量：SDK 查询/取消及刷新恢复已实现；真实双节点 TLS 授权/取消 race 和真实独立包浏览器 SDK 结果回读通过。禁用历史查询、撤权拒绝和刷新不重提交有子项证据；清单签名绑定、后端用户数据迁移和其余能力仍待补，整项保持进行中。见 [任务报告](reports/extensions-tasks-2026-09-12.md)。

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

实现状态：未开始 / 进行中 / 已实现 / 阻塞。  
验证状态：未运行 / 通过 / 失败 / 环境缺失 / 不适用（必须写明平台能力依据）。

“已实现”不等于“通过”。只交叉编译 Windows 不等于验证 Windows 运行；只有模拟后端的浏览器测试不等于真实链路通过。环境缺失不计入通过率，不得为了全绿改成不适用。

每条证据至少记录：编号、覆盖子场景、代码位置、可重复命令、测试环境和时间、实际结果/退出码、日志或截图路径、尚未覆盖项。可在本文件追加证据索引，详细报告存放在 docs/acceptance/reports/，但必须是真实生成的报告。

更新表格的实现/验证列并附证据链接。若只验证部分条件，写“进行中”或明确子项结果，不把整行标为通过。功能入口、后端权限、失败行为和持久化缺一时不能完成该 F 项。

## 3. 功能交付 F01～F14

2026-09-12 F09/F11/E02 Compose输出：成功/失败阶段输出有界持久化、截断与单阶段受控查询、任务详情浏览/刷新接入。实际CLI进程race、真实Docker/Compose浏览器1/1（45.9s）、Master权限测试及构建通过，见 [输出证据](reports/compose-output-2026-09-12.md)。运行中流式输出、崩溃窗口及完整平台/故障组合继续。

2026-09-12 F09/F11拉取边界：保存EOF前末条观测，超32MiB流不再因限流EOF进入成功检查，定向真实HTTP协议替身race3.743s及平台构建通过，见 [进度证据](reports/container-progress-2026-09-12.md)。真实下载、Compose输出持久化和完整组合继续。

2026-09-12 F09/F11镜像进度增量：保留Engine分层身份与Current/Total并写入节点日志/任务修订，HTTP协议替身定向race与构建通过，见 [分层进度证据](reports/container-progress-2026-09-12.md)。真实拉取及完整输出日志未验，不提升整项。

2026-09-12 F09阶段详情：发起者/时间/耗时、专用任务详情、持久状态阶段分页与当前授权接入。真实TLS/race、前端41/41、真实任务浏览器1/1（3.8s）通过，见 [阶段证据](reports/task-stages-2026-09-12.md)。各执行器完整输出/进度与长期故障组合仍保留，整项不提升。

2026-09-12 F09统计/筛选增量：全部获准未终态根任务统计与数据库状态筛选、刷新恢复已接入；超过1000条新历史遮蔽旧等待任务的权限TLS/race、真实浏览器组合2/2（9.8s）、前端39/39通过，见 [任务历史证据](reports/task-history-2026-09-12.md)。完整任务元信息/阶段日志、系统通知真桌面和故障组合继续。

2026-09-12 F09历史增量：后端稳定游标读取突破最近1000条，服务端当前权限过滤、桌面翻页/刷新及专用详情直接查询接入；1005项race、真实TLS及101个真实任务浏览器组合2/2通过，见 [任务历史证据](reports/task-history-2026-09-12.md)。后台数量聚合、全历史条件筛选和完整故障组合继续。

2026-09-12 F09终态通知续接：持久事件WSS游标、首次边界与重连、授权过滤、默认桌面通知及精确任务跳转、显式系统通知启用/关闭接入。最终浏览器组合3/3（9.8s）、WSS race2.321s、恢复日志故障保护通过，见 [通知证据](reports/notifications-2026-09-12.md)。操作系统原生通知展示/点击未验，整项继续进行中。

2026-09-12 F09/F14/E06通知增量：受控扩展通知API/SDK、持久站内托盘、来源绑定与点击跳转、立即刷新不重发接入。独立包浏览器1/1、TLS通知/云端选择策略race、前端37/37通过，见 [通知证据](reports/notifications-2026-09-12.md)。后台任务自动完成通知与系统通知发送继续，整项不提升。

2026-09-12 F14/E06元数据增量：独立包实例编辑、立即刷新草稿恢复、真实后端名称保存与刷新不重放，浏览器1/1（4.7s）通过；独立包资源桥接race30.708s通过。见 [元数据证据](reports/extensions-metadata-2026-09-12.md)。其余SDK能力及完整组合仍保留，整项状态不提升。

| 编号 | 必须交付的用户能力和边界 | 主要验收 | 实现 | 验证/证据 |
| --- | --- | --- | --- | --- |
| F01 | 桌面、任务栏、窗口移动缩放吸附/层级/焦点；应用多开、跨节点资源标签、标签移窗、快捷方式、菜单和键盘操作；可复制文本区域保留选择能力 | A01、A13、A15～A17 | 进行中 | 完整场景未通过；本轮专用实例窗口、真实跨节点双中心、账号菜单退出/现场隔离定向场景通过；跨窗口测试修正吸附区重叠后复验通过，见[rc13报告](reports/rc13-current-source-2026-09-20.md)；[当前核心证据](reports/core-2026-09-09.md)、[运行适配子项](reports/runtime-2026-09-09.md) |
| F02 | 布局、窗口/标签归属、正文、撤销重做、表单与视图现场、终端检查点；本地日志与快照、云端副本和版本迁移；账号/设备/浏览器标签隔离 | A01、A02、A06、A08、A11、A17 | 进行中 | 云端布局/引用默认同步、正文显式同步和打开恢复已有真实 1/1；同步 PUT 在中止后刷新重试复用同一请求键并只成功写入一次，真实 Chromium `cloud workspace sync` 1/1（8.0s）通过；配额失败、schema迁移和事务中止实际浏览器已通过（A11）；2026-09-19真实双设备 stale revision 与终端检查点内容策略 2/2（6.9s）通过且本地正文保留；2026-09-20并发快照超256KiB尾日志防护、终端输出增量日志及旧检查点兼容单测通过；RC17补齐主题/字号/密度偏好与默认布局备份恢复，设置浏览器3/3通过，见[RC17发行报告](reports/release-rc17-2026-09-22.md)、[多设备冲突证据](reports/workspaces-2026-09-19.md)与[性能恢复报告](reports/performance-continuous-2026-09-19.md)；断电和存储耗尽仍待补，见[工作区报告](reports/workspaces-2026-09-10.md)、[请求键报告](reports/request-keys-2026-09-14.md) |
| F03 | 登录退出、多用户、角色与资源授权；查看/创建/控制/文件/终端输入/主机管理分权；服务端列表、统计、请求及流均授权 | A08、A14、E04、E07 | 进行中 | 新增管理员审计检索（用户/节点/资源/动作/时间/分页）及普通用户 403 集成证据；用户创建/改密界面保存待处理请求键；本轮 `/files/access` 返回当前用户文件读写能力，真实 TLS 集成确认管理员可写、普通用户只读和撤权后拒绝，浏览器只读界面 1/1 通过；整项场景仍未通过，见[编辑器能力增量](reports/editor-capabilities-2026-09-22.md)、[审计与撤销证据](reports/audit-2026-09-10.md)、[请求键报告](reports/request-keys-2026-09-14.md)及[rc13报告](reports/rc13-current-source-2026-09-20.md) |
| F04 | 节点登记、身份轮换吊销、管理连接、心跳、版本/能力协商、在线/异常/离线/维护状态、入口；只管理通信 | A05、A08～A10 | 进行中 | 节点管理新增持久撤销、连接代次递增、幂等重试、分组/标签设置和界面入口；登记/换钥/撤销界面待处理请求键随窗口恢复，节点 TLS 管理集成通过；完整失联/平台组合仍未通过，见[审计与撤销证据](reports/audit-2026-09-10.md)、[请求键报告](reports/request-keys-2026-09-14.md) |
| F05 | 聚合全部获准实例、获准节点创建向导、搜索筛选统计/批量操作；多窗口多标签和专用资源窗口；真实启动停止重启与阶段诊断 | A03～A05、A10、A14～A17 | 进行中 | 完整场景未通过；实例中心现补齐可恢复的列表/卡片视图，筛选、选择和批量请求身份不受切换影响，`instance-batch.spec.ts` 1/1通过，见[实例视图报告](reports/instances-view-2026-09-22.md)；本轮真实跨节点双中心/专用实例窗口和节点范围创建复验通过；测试按夹具实例身份核验，不依赖临时测试实例的总数，见[rc13报告](reports/rc13-current-source-2026-09-20.md)；[当前核心证据](reports/core-2026-09-09.md)、[运行适配子项](reports/runtime-2026-09-09.md) |
| F06 | 实例控制台、PTY、只读观察/受权输入、多窗口标签/分栏、会话重新挂载、输出流控与检查点；主机 Shell 独立高权限 | A06、A08、A09、A17 | 进行中 | 完整场景未通过；2026-09-20真实WSS终端刷新/移窗后保持原会话、无输入重放1/1（17.6s）；增量输出检查点兼容旧格式、ACK仍等待本地保护，见[性能恢复报告](reports/performance-continuous-2026-09-19.md)；[当前核心证据](reports/core-2026-09-09.md)、[运行适配子项](reports/runtime-2026-09-09.md) |
| F07 | 多选、重命名、新建、复制移动、回收/删除、压缩解压、上传下载、跨节点复制与移动；真实根目录隔离、冲突和传输续接 | A07、A08、A09、A12 | 进行中 | 实例中心新增管理员软删除入口并沿用停止/活动任务/幂等约束；跨节点持久协调、权限撤销和源对象保护通过；浏览器上传分片按偏移使用稳定请求键，Master 在分片/完成前拒绝缺失键；重编译夹具后的512MiB分块校验与批量步进传输真实完成（110.868s、目标校验成功），并有[并行关闭网页续传](reports/upload-close-2026-09-13.md)证据。文件管理器本轮补齐按需目录树、列表空白区框选、可恢复的图标/列表视图和目录快捷方式，列表虚拟滚动/图标分页、根目录→子目录→孙目录刷新恢复、上传续传/取消均在文件浏览器套件5/5通过，见[目录树报告](reports/files-tree-2026-09-22.md)与[视图报告](reports/files-view-2026-09-22.md)。Linux目标tmpfs真实ENOSPC已通过（A12）；本轮 Linux `internal/filesystem` race 套件1.332s通过，包含符号链接/FIFO/目录交换隔离；Windows amd64 测试包交叉构建通过但未运行。物理掉电及其余故障组合仍单列待补 |
| F08 | 真实编辑器、语法能力、共享正文与独立视图、草稿/撤销重做恢复、显式保存、原子写入/版本冲突；大文件二进制有明确边界 | A01、A02、A08、A11、A17 | 进行中 | 完整场景未通过；UTF-8/BOM 与 LF/CRLF 元数据往返、4 MiB 上限显示、扩展名语言映射、共享正文的独立分栏视图、多光标编辑刷新后撤销重做，以及只读用户禁止编辑/保存已补；撤销历史按最多 8 MiB/文件上限两倍执行，过大活动编辑组整体落为基线以避免部分撤销；真实节点 API 验证 4 MiB 可读、超限文本/二进制拒绝且原字节可下载；新增受版本绑定的有限 UTF-8 只读分段预览，真实文件 API覆盖首段/第二段/二进制拒绝，浏览器边界套件3/3通过。完整跨应用/故障交叉场景仍未通过，详见[编辑器能力增量](reports/editor-capabilities-2026-09-22.md)；[当前核心证据](reports/core-2026-09-09.md)、[运行适配子项](reports/runtime-2026-09-09.md) |
| F09 | 持久化后台任务、阶段/进度/日志、等待、取消、重试和结果；站内通知与系统通知授权；关闭网页不取消服务端任务 | A04、A05、A07、A09、A10 | 进行中 | 任务中心显示持久结果、有界分页及失败/中断生命周期任务关联重试；真实TLS覆盖幂等、失败和授权边界。2026-09-29独立X11/DBus/Dunst中真实节点任务触发原生弹窗，系统鼠标点击正确taskId，标题/正文、刷新去重及浏览器退出后用户关闭均通过，见[原生通知报告](reports/native-notifications-2026-09-29.md)。不承诺退出后后台投递或重新启动；其他桌面平台与完整故障组合未由Linux结果覆盖。历史证据见[通知API](reports/system-notifications-2026-09-20.md)、[重试](reports/tasks-retry-2026-09-10.md)、[核心](reports/core-2026-09-09.md) |
| F10 | 节点/实例指标、历史采样策略与旧数据标识、完整系统进程视图、诊断和授权终止；窗口最小化降低绘制开销 | E01、E08 | 进行中 | [监控报告](reports/monitoring-2026-09-09.md)；节点/实例实时指标、历史、进程搜索排序和出生标识终止子项通过，2026-09-19真实Daemon失联/重连、旧指标时间/诊断和旧进程禁用入口1/1（53.4s）通过，见[失联监控报告](reports/monitor-offline-2026-09-19.md)；本轮 Linux 真实5秒间隔采样121点、Master缓存离线回读与Master重启后120个stale点恢复均通过，详见[长时监控报告](reports/monitor-history-soak-2026-09-21.md)；Windows采样仍待补 |
| F11 | 容器、镜像、卷、网络和 Compose 项目；日志/终端、配置草稿、显式应用和分阶段结果；数据删除与容器删除区分 | E02 | 进行中 | 当前源码在真实 Docker Engine 29.4.1/overlayfs 上的 Compose 生命周期 race 27.681s 通过，覆盖容器/镜像/卷/网络、双流日志、更新/删除、健康失败、卷保留和普通用户拒权；本轮容器指标及 Compose 浏览器场景2/2通过；宿主 PTY 与 Docker 浏览器边界也已通过；已有后端有界日志归档现已接入前端实时/归档切换，原有真实 Compose 长流程1/1（45.6s）复验通过，Docker 浏览器链路2/2、真实隔离 Docker Engine 历史 API 2.79s及真实 Chromium 日志 UI 1/1（7.6s）通过，真实链路覆盖归档持久化、未授权403、`limit=101` 400、实时/归档正文和缺口提示，详见[Docker日志归档报告](reports/docker-log-history-2026-09-22.md)；2026-09-21容器内1MiB tmpfs实际触发内核ENOSPC，Compose错误状态/exit code/日志及部分资源事实保持通过，见[ENOSPC报告](reports/docker-enospc-2026-09-21.md)；当前源码另经 loopback HTTPS TLS 代理对真实 Engine/Compose 完成同一生命周期 E2E 30.49s，通过已验证 HTTPS endpoint，但不是远端主机证据，详见[HTTPS传输报告](reports/docker-https-transport-2026-09-21.md)；远程 Engine、Engine主机存储耗尽/掉电及Windows容器仍待补，见[当前 Docker 验收](reports/docker-current-2026-09-20.md) |
| F12 | 可恢复备份、目标/范围、保留策略与一致性钩子；平台调度、时区、错过执行/重叠策略、触发重授权和任务结果 | E03、E04 | 进行中 | [备份与调度模块](reports/backup-scheduler-library.md)、[备份界面增量](reports/backup-ui-2026-09-10.md)、[真实备份浏览器](reports/backup-browser-2026-09-14.md)、[定时一致性验证](reports/scheduled-consistency-2026-09-20.md)；默认桌面已接入真实备份/恢复计划/定时任务入口，清理动作也保存稳定请求键，产品 HookRunner 已提供固定 argv 配置入口，实际固定argv保存程序成功/失败补偿、定时保存钩子、真实ENOSPC及节点离线/服务重开调度已通过；当前源码已验证备份 ZIP 写入中断后的归档清理/新任务恢复（[归档崩溃报告](reports/backup-archive-process-crash-2026-09-21.md)），以及恢复上传中途强制终止后的暂存清理、原ID中断对账和新计划成功恢复（[恢复崩溃报告](reports/backup-restore-process-crash-2026-09-21.md)）；scheduler独立进程及真实发行包 Master+所属Daemon 在已接受 occurrence 后强制终止、原 task/request 回执恢复且归档唯一，分别见[调度器进程崩溃报告](reports/scheduler-process-crash-2026-09-21.md)和[Master/Daemon 崩溃报告](reports/scheduled-backup-master-daemon-crash-2026-09-21.md)；真实 `Scheduler.Tick` 的 pending/prepared 两个已提交阶段中断后均会以同一 occurrence 恢复且无重复，见[调度阶段崩溃报告](reports/scheduler-staged-process-crash-2026-09-21.md)。设备写入中断、长期墙钟和物理掉电仍待补 |
| F13 | 按平台能力提供有限系统服务、防火墙和主机计划任务管理；管理员专用、变更差异、管理连接保护及回滚 | E05 | 进行中 | 能力探测、服务/计划任务任务链路、防火墙预览及 Linux firewalld 端口 apply 子集已实现；两阶段 60 秒租约、确认、取消/过期回滚和 Daemon 重启恢复测试通过；真实私有网络firewalld失联回滚已通过，见[私有环境报告](reports/firewall-private-2026-09-19.md)；不操作生产宿主，系统服务/计划任务与Windows环境仍待补 |
| F14 | 默认与扩展应用共同宿主；SDK、状态合同、受控能力、沙箱、注册源和真实安装包；安装授权、版本兼容、升级、禁用、卸载 | A13、E06、E07 | 进行中 | [SDK报告](reports/sdk-2026-09-09.md)；独立 `sdk/`、参考包、签名校验、安装/升级/回退/禁用/卸载、动态 bundle sandbox 加载和受控任务桥接已有 API/浏览器及真实 15/15 fixture 证据；本轮扩展真实浏览器7个独立场景最终均通过，涵盖通知、任务取消、草稿/迁移、WASI、目录两版本升级和权限隔离；`data-app-id` 仅供稳定定位，不改可见 UI；扩展管理界面现按扩展/版本保存生命周期请求键；HTTPS 远程目录源已有 TLS 单元/集成覆盖，资源标签/跨设备状态迁移已有A13证据；真实远程部署、高延迟证书轮换仍待补，见[rc13报告](reports/rc13-current-source-2026-09-20.md) |

## 4. 原规划关键场景 A01～A17

以下是可执行验证要求，不预设任何测试已经通过。立即刷新测试不得人为等待自动保存完成。

| 编号 | 场景与通过标准 | 实现 | 验证/证据 |
| --- | --- | --- | --- |
| A01 | 输入、中文输入提交、粘贴、连续编辑和拖窗后立即 F5/工具栏刷新；正文、窗口布局和应用状态一致，撤销/重做可沿刷新前历史继续；分别验证活跃与后台窗口 | 已实现 | Linux Chromium真实IME提交/剪贴板粘贴/连续编辑/拖窗/活跃与后台窗口reload两场景2/2（6.0s），末次输入无等待刷新、精确布局/焦点及undo redo通过，见[立即刷新报告](reports/immediate-refresh-2026-09-13.md) |
| A02 | 同用户两窗口打开同文件，随后外部用户修改服务器版本；共享正文策略正确、局部光标/滚动独立，保存时检测版本冲突并保留各版本，不用旧响应覆盖新输入 | 已实现 | Linux Chromium真实双账号文件冲突及迟到响应1/1（9.6s），共享模型独立光标/滚动/刷新后准确插入1/1（12.7s）；旧版本不覆盖新输入、两版本保留与显式解决通过，见[共享编辑报告](reports/shared-editor-conflict-2026-09-13.md) |
| A03 | 真实实例忽略正常停止、派生后代、主 PID 提前退出或后代持有管道；有限等待后升级或明确失败，全局仍可用，归属运行未退出不启动新代次 | 已实现 | 2026-09-29 本机私有委派cgroup出生归属、整组kill确认与CPU实际限流通过，见[平台报告](reports/systemd-cgroup-2026-09-29.md)。Linux真实双节点停止失败/拒绝新运行/另一节点可控race4.375s，真实主PID提前退出及管道后代/强制升级/截止失败三项race1.302s通过，见[停止边界报告](reports/stop-boundary-2026-09-13.md)。Windows/cgroup平台运行限制见E09 |
| A04 | 重复重启请求、服务端已接受但客户端丢响应、Master 中途重启；同 requestId 对账复用结果，同资源不同请求按策略串行，不重复产生运行 | 已实现 | 同key重启/管理断线race通过，服务与SQLite重开通过，独立Master SIGKILL后原任务及精确启动计数通过；真实浏览器202响应丢失后刷新重试1/1（4.4s），见[重启报告](reports/restart-disconnect-2026-09-13.md)。不同请求资源串行见A15，Windows运行见E09 |
| A05 | 重启和传输途中断开节点连接并恢复；标为失联/待确认，不伪造停止或成功；按实际执行记录校准，不因重连重复副作用 | 已实现 | 重启RUNNING管理断线原请求复用、传输中途bulk断线目标验证/提交唯一race12.490s、真实浏览器OFFLINE与WAITING_NODE原任务恢复1/1（50.7s），见[失联与恢复证据](reports/node-offline-browser-2026-09-13.md)及[重启断线](reports/restart-disconnect-2026-09-13.md)。其他平台见E09 |
| A06 | 在实际 PTY 中运行 vim/top 等全屏程序、中文/ANSI/光标模式后刷新；检查点和输出序号衔接正确，继续使用原会话，回放绝不重发历史输入 | 已实现 | Linux真实vim/top及WSS ANSI/备用屏幕/精确恢复sequence/原sessionId/输入不重放组合2/2（26.7s），真实文件副作用单次，见[终端程序报告](reports/terminal-programs-2026-09-13.md)。Windows运行见E09 |
| A07 | 同时发起节点端解压和来自浏览器本地文件的上传，再关闭网页；解压独立完成，上传保留已确认检查点并在需要时等待重新选择源文件，状态与真实结果一致 | 已实现 | Linux真实场景通过：10,000文件解压独立完成，16MiB上传在新browserTabId从983,040B复用原任务续传并核验摘要，见[完整场景证据](reports/upload-close-2026-09-13.md)。Windows平台由E09另行验证 |
| A08 | 使用过程中撤销查看、写入或终端能力；新请求拒绝、现有流及时终止/降权；切换账号不能访问前账号草稿；旧快照不能恢复授权 | 已实现 | Linux真实浏览器PTY撤权1/1（6.1s）、文件查看/写入撤权与旧草稿恢复1/1（6.7s）、账号退出现场隔离1/1（3.5s）通过；新请求403、输入租约失效、服务器正文不变、旧现场不能恢复授权，见[撤权报告](reports/browser-revoke-2026-09-13.md) |
| A09 | 大日志、慢接收端与批量文件传输并行；控制请求持续取得服务，队列、内存、归档磁盘有可测上限，背压/丢弃策略可见，无无限累积 | 进行中 | 真实 TLS Master/双 Daemon 并行复验通过（13.07s）：8.4MB 传输、慢日志未消费峰值82,635B、接收队列79,507B，停止≤5s且归档/目标校验成功；2026-09-21 RC15当前源码同一 race 组合复验13.918s通过；本轮当前源码真实TLS慢读者30秒逐秒采样：未消费峰值33,048B、接收队列峰值28,320B，随后文件传输、停止、归档重读和目标校验通过，见[30秒慢消费者报告](reports/slow-consumer-soak-2026-09-21.md)；真实PTY约17.5s、64MiB无消费者滚动采样归档峰值65,484B，PTY进程RSS基线741,376B/峰值3,764,224B（基线+32MiB内）；2026-09-27本机独立24h真实采样与归档峰值在声明预算内，详见[最终报告](reports/local-endurance-24h-2026-09-27.md)；跨平台、全连接更长期混合负载仍待补，见[混合流报告](reports/mixed-stream-control-2026-09-13.md) |
| A10 | Daemon 崩溃后重启，包含已启动进程与中途任务；按运行身份而非仅 PID 对账，既有进程/任务结局明确，不可恢复的操作标为中断 | 已实现 | 独立Daemon运行中及STOPPING中SIGKILL后恢复原运行/任务明确INTERRUPTED不重放；真实PID错误出生标识拒认拒信号及运行标记恢复4项race1.058s，见[Daemon崩溃报告](reports/daemon-crash-2026-09-13.md)。Windows真机见E09 |
| A11 | 注入浏览器配额写入失败并迁移旧恢复记录；不清空草稿、不显示虚假保护成功；快照完整切换，迁移失败可保留/导出旧记录 | 已实现 | Linux Chromium实际桌面故障注入3/3通过：配额失败导出最新正文、schema0迁移、schema999原始记录保留、指针事务中止前后数据库一致及同步日志刷新恢复；修复未捕获AbortError，见[恢复故障报告](reports/recovery-quota-2026-09-13.md)。2026-09-26 Firefox 155.0和WebKit 26.6对应真实浏览器各3/3、退出0，见[跨引擎恢复报告](reports/recovery-engines-2026-09-26.md)。实际配额阈值、物理掉电及浏览器清除站点数据仍不由故障注入结果覆盖 |
| A12 | 跨节点复制/移动遇到目标磁盘满、断线、来源变化；记录可诊断阶段和清理对象，未验证目标前不删除源；源改变后不错误删除新内容 | 已实现 | 传输断线、来源改变及删除前再次核对已有真实TLS证据；2026-09-19私有1MiB目标tmpfs真实ENOSPC竞态验收4.371s通过：FAILED/诊断明确，源完整保留、目标不发布、既有文件保留及暂存清理，见[空间不足报告](reports/enospc-2026-09-19.md)和[传输报告](reports/transfers.md)。Windows/远程文件系统及物理掉电另行未验 |
| A13 | 默认应用和参考扩展通过同一 AppHost 能力合同注册；无需改桌面核心可开窗口/标签、注册任务并恢复现场；扩展不能获得未授予能力 | 进行中 | 独立包真实浏览器已覆盖通知、元数据、迁移、WASI 任务、资源窗口/快捷入口/移窗/关闭及双版本目录安装；跨设备副本状态迁移与资源身份校准1/1；新增真实成员浏览器上下文未授权组合1/1：无 `app.use`、仅 `app.use` 无 `node.read` 均403，补齐节点读取后资源摘要200，撤销后立即403；本轮七个扩展场景在真实 HTTPS 双节点 fixture 上最终全通过；Windows/远程长期组合仍待补，见[扩展全链路报告](reports/extensions-real-2026-09-13.md)、[跨设备报告](reports/extensions-cross-device-2026-09-14.md)、[未授权浏览器报告](reports/extensions-unauthorized-browser-2026-09-14.md)及[rc13报告](reports/rc13-current-source-2026-09-20.md) |
| A14 | 两节点多实例、查看与创建权限不同；实例中心聚合全部可见资源，筛选/统计不泄漏；看见节点不等于可创建，提交再次校验权限/配额 | 已实现 | Linux Chromium真实双节点双账号：member 查看两实例但创建/第二实例控制受限；仅节点2临时授予创建与所需host.manage后向导仅列节点2，确认前撤权返回403且无实例，恢复授权后创建STOPPED实例并按ID清理，1/1（3.5s）；见[实例权限报告](reports/instance-permissions-2026-09-13.md)。Windows/远程环境见E09 |
| A15 | 两个实例中心窗口包含多个跨节点实例标签，重复查看同一实例；真实运行状态同步，视图导航独立；两视图并发控制仍由后端统一去重/串行 | 已实现 | 双节点双实例中心真实浏览器1/1（7.9s），并发启动仅一成功、另一拒绝且runId一致，两窗口状态同步、独立导航刷新保留，见[多视图报告](reports/instance-multiview-2026-09-13.md)；请求去重与故障归属另见[运行适配](reports/runtime-2026-09-09.md) |
| A16 | 添加实例到桌面，单击启动入口、显式新开、改入口名称、删除入口、目标改名/删除/离线/权限撤销；默认打开专用管理窗口且可复用已有专用窗口，不创建/启动/删除实例，不用同名资源替代 | 已实现 | Linux Chromium真实双节点：专用窗口复用/显式新开/入口改名删除1/1（3.6s）、管理员从实例中心点击删除确认后目标改名/软删除/同名替换固定旧resourceRef1/1（7.0s）、成员撤权入口失效1/1（3.3s）、Daemon离线入口状态与恢复1/1（50.4s）；软删除仅管理员且停止/无活动任务，见[入口报告](reports/shortcut-lifecycle-2026-09-13.md)。Windows真机见E09 |
| A17 | 把含未保存草稿或活跃终端的标签拖出新窗口，再移入兼容窗口并立即刷新；原视图 ID、归属、顺序、正文、撤销历史与会话引用保留，不重复视图/任务/输入 | 已实现 | Linux Chromium真实编辑器拖拽1/1（3.8s）、实际PTY拖拽1/1（8.6s）通过；逐次立即刷新核对身份/归属/顺序/撤销重做/会话和输入帧，真实文件副作用仅一次，见[拖拽报告](reports/drag-recovery-2026-09-13.md)。其他平台见E09 |

## 5. 补充验收 E01～E09

2026-09-22 E08 非 Chromium 双 PTY 复核：强化夹具后，官方 Firefox 155.0 在真实8窗口/双PTY/10,000项目录/16MiB循环传输的60秒组合中两条PTY均约116KB/s、传输和队列正常，但1,020个指针样本 p95 **144ms**，超过≤50ms；此前官方WebKit同组合为78ms，Chromium为42ms。见[Firefox报告](reports/firefox-performance-dual-pty-2026-09-22.md)与[WebKit/Chromium复核](reports/webkit-performance-dual-pty-short-2026-09-22.md)。E08保持进行中。

| 编号 | 必须通过的完整场景 | 实现 | 验证/证据 |
| --- | --- | --- | --- |
| E01 | 主机与实例指标来自真实节点；系统进程搜索/排序/详情及授权操作可用，终止只命中指定测试进程；失联标明采样时间，不把旧值当实时；采样和历史保留有边界 | 进行中 | 实时/历史/身份终止子项通过；节点与实例指标均有 120 点有界持久缓存，节点失联回退标记 `stale`；浏览器已验证进程搜索、排序和 `startTicks` 提交 | [监控报告](reports/monitoring-2026-09-09.md)；真实节点SIGSTOP/CONT的采样时间保持、旧数据诊断、缓存进程旧标记/禁用终止和重连更新1/1（2026-09-20，54.0s）通过；2026-09-21 Linux隔离HTTPS真实Master/Daemon按5秒间隔采样121次，在线历史精确保留120点（10.1m, 1/1）；1秒间隔完整组合中离线 stale、同库 Master 重启后的120点持久回读和 Daemon 恢复在 Chromium 1/1（2.7m）、Firefox 1/1（2.9m）和官方隔离容器 WebKit 1/1（2026-09-22）均通过，见[长时监控报告](reports/monitor-history-soak-2026-09-21.md)与[WebKit报告](reports/webkit-monitor-history-2026-09-22.md)；2026-09-22 受控 Docker namespace 的 `tc/netem` 80ms单向延迟、10ms抖动、1%丢包下，WebKit 与官方Playwright Chromium 同一完整组合均1/1通过（各3.6m），见[netem报告](reports/netem-monitor-history-2026-09-22.md)；节点/实例缓存单测另以125个UTC时间戳跨午夜验证裁剪后仍严格保序（定向race 2.718s，见[测试源码](../../internal/master/monitor_integration_test.go)）。2026-09-27本机独立24h真实采样、两次失联/重启与120点stale历史保持已通过，详见[最终报告](reports/local-endurance-24h-2026-09-27.md)；仍不覆盖跨主机远端网络或Windows运行 |
| E02 | 在真实 Docker/Compose 测试环境创建项目、应用配置、查看日志/终端、更新和删除；镜像/卷/网络操作可追踪；镜像拉取失败/依赖缺失/部分成功明确；删容器默认不误删卷数据，配置草稿刷新保留 | 进行中 | 当前源码在 Docker Engine 29.4.1/overlayfs 上的 `TestRealDockerComposeLifecycle` race 27.681s 退出0，覆盖真实容器、镜像、卷、网络、双流日志、Compose创建/更新/删除、健康失败、卷保留和普通用户拒权；真实 Master/Daemon 容器测试另验证归档窗口持久化后的历史 API、403 权限边界和 limit 边界（2.79s）；真实 Chromium 日志 UI 1/1（7.6s）验证启动容器、实时/归档正文、缺口提示和清理；2026-09-21扩展后当前源码真实E2E race 29.17s通过，其中1MiB容器tmpfs实际触发内核ENOSPC，应用保留FAILED/health、exit code 1、容器事实及日志，见[ENOSPC报告](reports/docker-enospc-2026-09-21.md)；另以私有证书、loopback HTTPS 代理连接实际 Docker Engine，Engine 客户端和 Compose TLS 生命周期 race E2E 30.49s通过，详见[HTTPS传输报告](reports/docker-https-transport-2026-09-21.md)。这不是远端主机验证；远程 Engine、Engine主机存储耗尽/掉电、Windows容器和长时故障仍待补 | [当前 Docker 验收](reports/docker-current-2026-09-20.md)、[HTTPS传输场景](reports/docker-https-transport-2026-09-21.md)、[ENOSPC场景](reports/docker-enospc-2026-09-21.md)、[Docker/Compose API](reports/containers-api.md)、[真实 Docker 生命周期报告](reports/containers-e2e-2026-09-09.md)、[Docker 浏览器报告](reports/docker-browser-2026-09-09.md)、[Docker日志归档报告](reports/docker-log-history-2026-09-22.md) |
| E03 | 创建有可核对内容的备份并恢复到受控目标，验证文件内容/元信息和权限；保留策略只清理自身范围；一致性钩子成功/失败分别呈现；中断、空间不足、恢复覆盖均有明确处理 | 进行中 | 子项通过 | [备份与调度模块](reports/backup-scheduler-library.md)、[备份界面增量](reports/backup-ui-2026-09-10.md)、[竞态复验](reports/backup-scheduler-race-2026-09-14.md)、[真实浏览器](reports/backup-browser-2026-09-14.md)；默认界面已接入版本核对、恢复计划和显式覆盖确认，真实 Master/Daemon/Chromium 已核对归档、恢复正文和任务回执，Linux真实备份仓库及恢复目标ENOSPC竞态已通过，见[空间不足报告](reports/enospc-2026-09-19.md)；实际固定argv保存程序、保存失败补偿和恢复最新正文race6.174s已通过；新增 Linux 真实归档写入中 SIGKILL 场景：原任务显式转为 `interrupted/unknown`，未发布 snapshot，部分 ZIP 按对象身份清理，新 ID 备份成功，见[归档进程崩溃报告](reports/backup-archive-process-crash-2026-09-21.md)；新增 Linux 真实恢复上传中 SIGKILL 场景：分片取消并清理、目标文件不发布、原ID保持部分中断结果，同ID不重放，新计划完成恢复并校验正文SHA，见[恢复进程崩溃报告](reports/backup-restore-process-crash-2026-09-21.md)。物理掉电、设备缓存丢失及 Windows 运行仍未验证 |
| E04 | 时区含夏令时、节点离线、Master 重启、错过触发和同资源重叠；按声明策略执行且不重复补跑；创建者失去权限后新触发拒绝；节点离线仅继续已接受任务，不从陈旧计划无限产生特权工作 | 进行中 | 子项通过 | `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（备份1.856s，Master 5.816s）；真实 Chromium 保存/删除 `Asia/Shanghai` 计划 1/1、3.5s 通过；2026-09-19真实Daemon关闭/重连、受控时钟五次到期合并补跑及离线撤权race2.869s通过；2026-09-19定时任务接受后Master/SQLite实际关闭重开、双节点重连且任务/归档唯一race3.198s通过；2026-09-21 scheduler独立进程在accepted occurrence提交后强制终止、SQLite重开与无重复/新触发撤权子项 race 1.130s通过；同日真实 RC15发行包按真实分钟触发计划后同时 SIGKILL Master 与所属 Daemon，重启后 startupId变化、taskId/requestId不变、原任务SUCCEEDED且仅一份归档，退出0，详见[Master/Daemon 崩溃报告](reports/scheduled-backup-master-daemon-crash-2026-09-21.md)；当前源码实际 `Scheduler.Tick` 的 pending slot 已保存但尚未 BuildTask、prepared fire/task 身份已保存但尚未 Accept 两个阶段均以 SIGKILL 中断；重启后分别只构造并接受一次、复用相同 taskId/requestId，重复 Tick 不重放，完整验证见[调度阶段崩溃报告](reports/scheduler-staged-process-crash-2026-09-21.md)；另以独立子进程在共用 `metadataMutation` 事务中写入 schedule/fire 记录后、回执/审计/提交前 SIGKILL，SQLite重开未见部分记录且完整性检查通过，见[元数据事务崩溃报告](reports/metadata-transaction-crash-2026-09-21.md)；2026-09-22 同一受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 两个真实分钟槽各只接受一次、第二槽绑定新源版本且两份快照恢复校验通过（1.5m），见[netem分钟调度报告](reports/netem-schedule-wall-clock-2026-09-22.md)。上述是进程级/受控网络边界，不是 SQLite 设备写入中的断电；2026-09-27本机独立24h真实墙钟、1,440个成功分钟槽与末尾最新备份恢复已通过，详见[最终报告](reports/local-endurance-24h-2026-09-27.md)；更长期、物理掉电、跨主机网络及Windows仍未验证 |
| E05 | 在独立且获准的测试环境操作支持的服务/计划任务/防火墙子集；只修改目标对象，展示差异、保留既有规则；模拟管理链路失联后按期限自动回滚，再连接可诊断；普通用户不能调用主机特权 | 进行中 | 2026-09-29独立systemd容器服务/定时器生命周期与列表真实验证通过，修复停止服务及未加载timer遗漏，见[平台报告](reports/systemd-cgroup-2026-09-29.md)。跨平台能力探测、有界service控制、timer启停、firewalld及host.manage API已有实现；持久租约、确认、取消/过期回滚、失败诊断与重启恢复通过。2026-09-19私有网络真实firewalld管理端口阻断和期限回滚race25.998s通过，见[防火墙报告](reports/firewall-private-2026-09-19.md)。9月20～22日无法启动manager的记录仅为历史环境失败；当前Linux对应子项已有证据。Windows及其他实际部署环境仍待验证 |
| E06 | 按 SDK 文档从独立示例目录构建新扩展；不修改桌面核心即可安装、开多窗口/标签、注册资源入口和后台任务、保存迁移现场；前后端沙箱与能力桥接实际拒绝未授权访问 | 进行中 | SDK 构建通过；`POST /extensions/{id}/tasks` 受 `app.use`（扩展资源）+ `node.read` 最小权限保护，bundle 在 sandbox iframe opaque origin 执行，Master/Daemon 集成和真实浏览器覆盖成功/拒权；真实 WASI guest 用户数据迁移、失败保留、回滚和事务恢复 race 通过，见[扩展数据迁移报告](reports/extensions-data-2026-09-20.md)；真实成员浏览器上下文未授权组合1/1通过，跨设备状态迁移/资源身份校准1/1；2026-09-22 受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 独立包安装、多窗口、WASI节点任务、已接受回执丢失后的稳定请求键恢复、快捷方式与刷新现场均通过（49.4s），见[netem扩展报告](reports/netem-extension-task-2026-09-22.md)；Windows 与长期跨主机远程场景仍待补，见[扩展跨设备报告](reports/extensions-cross-device-2026-09-14.md)、[未授权浏览器报告](reports/extensions-unauthorized-browser-2026-09-14.md) |
| E07 | 真实测试注册源提供至少两个版本扩展；浏览/获取/验证/授权/安装/启用/禁用/升级/失败回退/卸载跑通；篡改包、版本不兼容、能力提升均按策略阻止；禁用不再接受新任务，卸载默认保留用户数据并给出清理选项 | 进行中 | fixture 本地注册源提供 `v1.0.0`/`v1.1.0`；签名/完整性/回退/能力提升拒绝、启用/禁用/卸载数据保留测试及真实浏览器两版安装和 bundle 执行通过；后端用户数据迁移、失败保留、回滚和事务恢复 race 通过，见[扩展数据迁移报告](reports/extensions-data-2026-09-20.md)；`TestExtensionRemoteCatalogAdminAPI` 验证 Master API 经 loopback HTTPS 浏览/安装、普通成员拒权、同一 CA 的叶证书续期，以及切换新 CA 后旧信任失败关闭、显式更新信任后安装新版本，race 通过，见[原注册源报告](reports/remote-catalog-rotation-2026-09-20.md)和[CA 根轮换报告](reports/remote-catalog-root-rotation-2026-09-21.md)；2026-09-22 受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 本地注册源两版本浏览、安装、升级、禁用/启用和沙箱bundle执行通过（13.6s），见[netem注册源报告](reports/netem-extension-catalog-2026-09-22.md)；公网远端部署、跨主机高 RTT、证书信任更新的实际运维流程和 Windows 仍待补 |
| E08 | 记录参考设备、浏览器、RTT、日志速率和带宽；8 个窗口、2 个活跃终端、万条目录虚拟列表与后台传输并行，实测本地反馈 p95 与长帧、内存/队列/归档上限；以 p95 ≤ 50 ms 为原规划目标，未达标如实记录并优化，不伪造数据 | 进行中 | **2026-09-29 RC25仍未通过**：完整通透材质Chromium/Firefox/WebKit p95 **1003.9/1108/545ms**；快照专项改善不能替代E08，见[本轮报告](reports/local-closeout-2026-09-29.md)。历史2026-09-26当前冻结UI未通过：官方容器同一当前源码/完整材质下Chromium153 p95 **1240.1ms**、Firefox155 **1230ms**、WebKit26.6 **739ms**；每项保留八窗口、双真实PTY、万项目录、16MiB循环真实跨节点传输、两次RAF指针测量，双流/目标校验/有界队列正常，≤50ms明确失败。终端检查点Worker与raw日志修复、63项单测及4项真实parser/恢复/移窗/回退验证通过，不以功能结果代替性能。测试侧去模糊/分层对照未进入产品且不能作为正式通过，详见[当前性能及对照报告](reports/performance-current-2026-09-26.md)。历史证据：9月21日当时源码/旧UI Chromium一小时3,604.492秒、48,840样本、p95 48.1ms，见[当时小时报告](reports/performance-hour-recovery-copy-2026-09-21.md)；9月22日WebKit一小时p95 78ms失败，见[WebKit小时报告](reports/webkit-performance-hour-2026-09-22.md)，同日双PTY短时WebKit78ms/Chromium42ms见[当时双PTY复核](reports/webkit-performance-dual-pty-short-2026-09-22.md)。这些历史数据不覆盖本次冻结UI；远端公网、高RTT/丢包、硬件加速设备、Windows及更长期跨平台组合仍需补验 |
| E09 | 从干净依赖环境按说明构建并初始化 Master/Daemon/前端；至少两个节点和不同权限用户跑通；交付 Linux 与 Windows 产物/构建入口，分别记录 Linux 运行、Windows Job/ConPTY 真机结果；验证配置/数据迁移、备份恢复、升级回退和示例扩展开发说明可复现 | 进行中 | 2026-09-29 RC26六包/2,215条内部记录/三份Web与独立启动、停机恢复及兼容RC25回退通过；Linux systemd与委派cgroup（含CPU实际限流）已有本机隔离证据，见[平台与发行报告](reports/systemd-cgroup-2026-09-29.md)。Windows及其他未覆盖组合仍待验证。Linux 构建/初始化与多节点权限子项已有证据；`make sdk`、`make web` clean 依赖安装和构建退出 0，`make windows` 及 runtime/runlog/systeminfo Windows 测试入口交叉编译通过，OpenAPI YAML 当前为 105 路径 / 122 操作；storage、backup（含本轮 scheduler stage/restore crash）、containers HTTPS E2E 和本轮 `internal/filesystem` 测试程序均以 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c` 编译成功（仅构建证据，不代表Windows运行）；rc3独立发行包的停机快照恢复及兼容rc2回退实际通过，见[rc3报告](reports/release-rc3-2026-09-19.md)；2026-09-19从全新Go模块/构建/npm缓存完成Linux/Windows、前端、SDK及参考包构建，产物与rc4一致，见[冷构建报告](reports/cold-build-2026-09-19.md)；rc5含多核 Windows CPU 配额修复、rc7含定时备份一致性 UI 且包冒烟通过；rc13六包SHA校验、停机恢复与兼容rc12回退通过，见[rc13报告](reports/rc13-current-source-2026-09-20.md)；rc14包证据见[RC14报告](reports/release-rc14-2026-09-21.md)；RC23证据见[历史报告](reports/release-rc23-2026-09-22.md)。2026-09-29 **RC25** 六包/2,207条内部记录/三份Web一致性及独立启动/停机恢复/兼容RC24回退通过，见[本轮报告](reports/local-closeout-2026-09-29.md)。历史2026-09-27 **RC24** 已包含冻结UI、检查点Worker、Compose授权与稳定窗口DOM修复：六包构建/外部SHA及2,191条包内记录核验、全部三份Web与当前构建一致、独立SDK/参考包签名、TLS/双Daemon、停机快照与兼容RC23回退通过，见[RC24报告](reports/release-rc24-2026-09-27.md)；完整真实功能51/51见[功能报告](reports/local-functional-2026-09-27.md)。Windows Job/ConPTY真机、systemd及其他系统组合仍缺环境，任意schema降级不作承诺 |

E05 的真实防火墙验证不授权改动当前生产宿主机。E09 的 Windows 真机缺失必须写“环境缺失”，不影响继续实现代码，但不能计作全部验证完成。E08 的设计目标不代表已测性能。

Windows Job/ConPTY/监控/系统适配、独立 systemd manager、远程 Engine/网络故障和长期物理故障的可复现实验入口见[平台验收运行手册](../operations/PLATFORM_VALIDATION.md)。该手册只提供真实环境运行步骤；当前 runner 缺少对应环境，不能把命令存在或交叉编译结果计作通过。

## 6. 横向核对

- 默认应用和扩展应用均服从账号/节点/资源权限及状态版本协议；窗口标题、确认框与任务记录显示实际目标。
- 服务端成功结果到达前，界面只能显示已提交/执行中；本地拖动、输入和导航立即反馈。
- 非文本桌面区域使用应用式选择、框选和拖放；编辑器、终端、日志等需要复制的区域保留受控文本选择。不得全局封禁浏览器全部快捷键。
- 节点后台任务、PTY 会话、真实实例、前端窗口和浏览器标签的生命周期分别处理。
- 同一用户多个浏览器标签/设备不能静默覆盖彼此较新的状态；显式退出处理未同步草稿与本地数据隔离。
- 不将私钥、令牌、密码、真实用户草稿或不必要的文件正文写进验收日志。
- 模拟后端仅可作为测试替身；默认产品流程不得使用固定成功响应、虚构节点/实例或假进度。

## 7. 执行台账与完成判定

当前统计：14 个功能项、17 个原规划场景、9 个补充场景均尚未整项通过。已开始项及子场景证据见上表；部分测试通过不计为整项完成。

2026-09-09增量证据：F03/F04见[用户角色及节点维护换钥](reports/administration.md)；F05及A04/A14配置与真实自动启动子项见[实例设置](reports/instance-settings.md)；F06/F07/F08及A01/A02/A06/A07/A17见[真实数据流集成](reports/stream-integration.md)、[文件API](reports/files-api.md)、[桌面浏览器](reports/desktop-2026-09-09.md)；A12跨节点/同宿主关系保护与11项真实TLS测试见[传输报告](reports/transfers.md)；F11/E02见[Docker/Compose](reports/containers-api.md)，宿主 docker-host PTY 双主体测试见[容器报告](reports/containers-2026-09-09.md)；F12/E03/E04见[备份与调度](reports/backup-scheduler-library.md)及 Master API 集成测试；A13/E06见[SDK报告](reports/sdk-2026-09-09.md)；F13/E05 的系统管理浏览器边界见[系统管理报告](reports/system-management-2026-09-09.md)。当时真实磁盘满、完整故障组合、容器实例日志归档、监控、系统管理危险变更、扩展安装生命周期和 Windows 真机仍未验证；其中 Docker 日志归档已在 2026-09-22 补上前端入口和浏览器证据，其他边界继续按各项记录。均为已明确列出的子场景，未将交叉编译、替身测试或尚未运行的组合计为整项通过。

2026-09-10最新真实复验：在隔离 fixture `.local/fixture-1866434427` 以真实 HTTPS Master、两个 Daemon 和 Docker 29.4.1 运行，`files-terminal.spec.ts` 2/2 通过（27.1s），`npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` **15/15 通过（2.2m，退出码 0）**。本轮覆盖真实登录、节点/实例、文件、PTY/终端、日志/传输、设置/编辑器、模板创建、账号权限、Docker/Compose、扩展两版本 sandbox 和云工作区；证据详见[桌面真实复验](reports/desktop-2026-09-09.md)、[工作区报告](reports/workspaces-2026-09-10.md)和[SDK报告](reports/sdk-2026-09-09.md)。Docker stop 的 300 秒服务端等待已由 Engine transport 5 分钟 header timeout 覆盖，容器回归与真实 Compose 浏览器 1/1 通过，详见[真实 Docker 报告](reports/containers-e2e-2026-09-09.md)。Master 数据链接重连窗口也已修复并由完整真实套件验证。该证据不改变矩阵中 Windows、真实 ENOSPC/掉电、远程/长时负载、危险主机变更、扩展网络注册源分发和完整资源迁移等未覆盖项。

2026-09-10交付与安全增量：`docs/api/openapi.yaml` 通过 Python `yaml.safe_load`（OpenAPI 3.1，94 路径/108 操作；审计与节点维护增量后更新为 95 路径/109 操作，本轮任务重试后为 96 路径/110 操作）；`Manager.List` 跳过升级回滚快照并拒绝清单身份错配；防火墙租约回滚开始前持久化 `rollback_in_progress`，过期/取消直接执行路径写入终态，普通/race 租约测试通过。`make check`、提权 `make test`（`go test -race ./...`）、前端 `npm run check && npm test`（26/26）、`make windows` 和 Linux 三个二进制构建均退出 0。

2026-09-10审计与节点维护增量：新增 `003_audit_node.sql` 将审计记录统一为毫秒时间并回填节点身份；`Store.Audits` 提供管理员可分页筛选，UsersApp 增加用户/节点/资源/动作/时间检索。节点撤销改为事务幂等回执并保留实例资源，NodesApp 增加管理员确认入口。`go test` 存储/ Master 定向集成、`make check`、`cd web && npm run check` 通过；OpenAPI 更新为 95 路径，证据见[审计与撤销报告](reports/audit-2026-09-10.md)。

实现阶段请在此追加证据索引，并在 [PROGRESS.md](../execution/PROGRESS.md) 同步当前状态。建议每条记录格式：

- 编号及子场景：
- 代码/入口：
- 运行命令与环境：
- 结果、退出码和时间：
- 日志/截图/产物：
- 未覆盖范围及下一步：

最终报告必须明确区分：已实现且已验证、已实现但未验证、尚未实现、环境阻塞。全部必需功能与关键场景具有真实证据且缺口清零，才可以宣布全范围完成；后续仅剩视觉和非关键体验细化时也要具体列出。
# Windows returned evidence update (2026-09-30)

See [second Windows report](reports/windows-second-report-2026-09-30.md): three
browser suites executed with 118/119, 115/119 and 111/119 passes respectively;
Go unavailable meant native and full Go checks did not run. Windows acceptance
remains incomplete; prior native results are not new validation.

## Third report corrections (2026-10-01)

The [third Windows report](reports/windows-third-report-2026-09-30.md) ran
Chromium 120/120, Firefox 116 pass / 3 fail / 1 skip, WebKit 112 pass / 7 fail /
1 skip; Go events were 330 pass / 22 fail / 15 skip. Nine of ten required native
checks passed, while ConPTY failed and race was blocked by a missing compiler.
See [corrections and local regression evidence](reports/windows-report-fixes-2026-10-01.md)
for directory handle access, ConPTY stdio, portable real-process fixtures,
editor history, browser input and network/transfer outcome fixes. Local Go race,
frontend units and Windows cross-compilation pass; Windows runtime retest is
still pending. The new runner checks 27 named regressions, 11 native cases and
affected browser scenarios, and can supply a verified portable race compiler.
No entire matrix item is promoted to complete from these local results.
