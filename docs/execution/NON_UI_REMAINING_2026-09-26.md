# UI 暂停后的剩余工作核对

2026-10-06 最新结论：R4普通Windows真机真实联调已于固定e546c8f0 **14/14通过**，报告完整性和八阶段exit0核验通过，见[报告](../acceptance/reports/windows-device-passed-2026-10-06.md)。R2、R3和R4均已关闭；**当前非发行剩余只有R1 E08严格延迟**。R5发行仍由用户延期，无需再运行任何已关闭Windows补测组。以下历史R4待执行措辞不再代表当前任务。

更新：2026-10-05，按用户最新要求由三个子代理只读审查并收缩难提供环境的必测范围，见[范围调整与审查报告](../acceptance/reports/scope-review-2026-10-05.md)。保留原路径供已有链接使用；本页“尚未完成”的 R1～R5 是当前阻塞清单，下面历史报告中的旧补测指令不再代表下一动作。已验证的发行基线为 RC34，后续源码修复尚需最终发行收尾，不把历史51/51冒充最新全平台验收。

物理掉电、缓存丢失、宿主/NTFS实际磁盘耗尽、浏览器真实配额阈值/驱逐、独立公网Engine/注册源/证书运维、多平台全组合及超过既有24h的长测，改为本次不强制的环境兼容性项。Windows管理员服务/任务实际变更、通知中心/系统IME/原生对话框、额外headless/自然周期/冷启动/指针/绘制与容器组合也不要求逐项提供环境。已有安全ENOSPC、SIGKILL、权限和恢复测试保留；审查不记作对应实测PASS。Windows防火墙仅支持读取状态，规则应用/回滚不是当前Windows支持子集。

最新第十六轮：[固定8d6b14e报告](../acceptance/reports/windows-sixteenth-report-2026-10-05.md)确认Windows实际headed剩余取消导航用例3/3，正式构建及全部阶段PASS，完整checksum/ZIP/来源及阶段检查点匹配，原45/5秒和恢复断言保留。结合[第十五轮](../acceptance/reports/windows-fifteenth-report-2026-10-05.md)刷新/queued/日志/Worker终端各3/3，用户要求的visible WebKit补测组结束，无需再跑同组脚本。两份报告按原版本独立记录，不导入旧PASS或伪称新版单批15/15，旧14/15 FAILED保留。未测环境仍不是PASS，但是否阻塞本次交付以本页最新范围为准；E08和普通Windows真实联调/发行验收仍保留。

## 历史增量（保留结果，不作为当前补测要求）

第十四轮 Windows 增量：[固定63cacc0的报告](../acceptance/reports/windows-fourteenth-report-2026-10-04.md)确认17/17摘要及完整ZIP匹配，构建PASS、OS端口3570 fresh preview成功，WebKit12/15。无帧/真实queued/日志各3/3；剩filter恢复5秒、held read准备迟到、初次导航慢后接管45秒耗尽，Windows原因未定。脚本新增可选-Headed同五项3次环境对照和逐执行实际模式证据，原45/5秒及既有断言保持。初次Linux headed12/15及两次Worker探针0/1保留，发现reactive controls跨Worker克隆失败；现controls/links转为纯数据，真实刷新后验证Worker保持，UI/协议/ACK/不重放不变。新类型构建、90单测、脚本harness/实际JSON/finally通过；Linux headed/headless精确各15/15，WebKit六文件终端恢复62/62、Firefox/Chromium正式preview各7/7，无skip/retry/flaky/全局错误且引擎依次。回传仍保留FAILED，不用本地修复假称Windows三处超时已解决。下一次固定新提交运行`-Followup -BrowsersOnly -WebKitOnly -Headed`，保持桌面解锁且允许自动窗口运行，正式构建PASS及五项3次15/15回传；可见模式结果不自动关闭默认无界面可靠性，完整原生/E08/外部缺口保持。

第十三轮 Windows 增量：[固定c63bcb4的报告](../acceptance/reports/windows-thirteenth-report-2026-10-04.md)确认14/14摘要及完整ZIP集合匹配，正式构建24.04秒PASS；5173端口占用使测试启动失败，五项各3次均未执行。原监听归属未知，不杀/不复用；补测脚本现每个引擎启动前选择OS空闲loopback端口，同步测试/server地址，strictPort/fresh保持，报告新增有界全局错误和实际URL/端口。原任务恢复缺口未因零执行关闭，产品/UI/用例/45及5秒不改。本地端口占用/抢占失败关闭/6种非法配置及PS7 harness/类型通过；保持5173占用时WebKit精确15/15、97.036秒，Firefox正式preview7/7、66.612秒及Chromium开发server7/7、78.294秒通过。真实JSON/finally保留回传零执行FAILED及启动错误，1,145文档链接/差异检查通过，所属容器和会话结束。下一次仍固定新版`-Followup -BrowsersOnly -WebKitOnly`，正式构建PASS且五项各3次严格15/15回传，未选择项不导入旧PASS。

第十二轮 Windows 增量：[固定90658d0的报告](../acceptance/reports/windows-twelfth-report-2026-10-04.md)确认17/17摘要和完整ZIP集合匹配，正式构建PASS、WebKit12/15；终端/日志/无帧恢复各3/3。剩任务窗口准备耗尽45秒、首列表未就绪及导航后原5秒计数失败，后者具体Windows原因未确立。测试现原生Enter打开可见可用任务应用，确认首列表HTTP/body和筛选控件；queued数据只在浏览器实际departure后切换，防止旧普通轮询提前满足恢复断言。诊断补真实JSON消费/DOM计数/浏览器时间的数字及静态事件白名单，非绘制验证。产品/UI不改，原45/5秒/ACK/不重放/错误断言保持；最终WebKit21/21、Firefox7/7、Chromium7/7及冷视图三次真实轮询后释放1/1、类型构建/harness/实际JSON和诊断复制通过，严格15次CLI选择仅核验、不计另一次运行；故意失败0/1诊断探针独立保留。交付固定新版`-Followup -BrowsersOnly -WebKitOnly`，正式构建且同五项各3次严格15/15；Windows实效/完整原生/E08/外部缺口保持，未选择Go/SDK/其他Windows引擎不导入旧PASS。第十一轮结果和修订保留于[历史报告](../acceptance/reports/windows-eleventh-report-2026-10-04.md)。

第十轮 Windows 增量：[固定3ef289a的报告](../acceptance/reports/windows-tenth-report-2026-10-04.md)确认正式构建通过、WebKit12/15，终端3次通过；剩两次控制台点击稳定等待45秒超时（尚无日志读取），及一次无帧汇总5秒断言失败（之后有HTTP完成）。Windows具体绘制/数据发布原因仍未确立。安全读取恢复后现立即重查活动任务查询，保留确认缓存，不重放写入；日志生命周期测试以原生Enter键打开同一实际UI，明确不计这三步指针稳定覆盖。所有查询/日志补实时阶段和有界数字计数/按钮几何诊断。90单测/类型构建、WebKit36/36、Firefox12/12、Chromium12/12及脚本harness/真实JSON/严格15次选择核验通过；探索暂停时钟未复现指针停顿，独立保留，不冒充修复因果。下一次仍`-Followup -BrowsersOnly -WebKitOnly`五项各3次且构建PASS、15/15，不重复Go/GCC/SDK/其他引擎，不导入旧PASS；UI冻结，Windows实效和原生/E08/外部缺口保持。

第九轮 Windows 增量：[固定3e6c31c的报告](../acceptance/reports/windows-ninth-report-2026-10-03.md)确认WebKit14/15，四项导航/日志用例各3次通过，前轮两项超时已获真机通过。仅第一次终端耗尽原45秒总时限，之后两次约19秒通过，无页面/请求错误；首次刷新约32秒才开始，不能从上传确认具体慢动作/Windows因果。补测改为先单独构建正式frontend，再在fresh preview运行，新增实时阶段耗时和有界静态资源诊断，保留原断言/时限，UI不改。下一次仍是`-Followup -BrowsersOnly -WebKitOnly`五项各3次，必须构建通过且15/15；不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS。本地验证详情见报告，Windows实效和完整原生/E08/外部环境缺口保持。

第八轮 Windows 增量：[固定57d4fb6的报告](../acceptance/reports/windows-eighth-report-2026-10-03.md)确认WebKit13/15，原终端配置3次通过，剩两次读取恢复/测试建立超时。可控无帧基线0/1复现汇总恢复超时，API只读门禁现增加250ms独立恢复兜底，pagehide阻止旧帧/旧截止恢复，新导航重置代次；日志测试用真实interval回调建立待取消读取，queued观察不依赖绘制，原限时/错误/ACK/输入不重放保持。89单测、类型构建及本地WebKit36/36、Firefox12/12、Chromium19/19加独立补验4/4、脚本harness/真实JSON和精确15次选择通过。下一次仍仅`-Followup -BrowsersOnly -WebKitOnly`五项各3次，不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS；Windows实效、原生/E08/外部环境缺口保持。

第七轮 Windows 增量：[固定8ad386a的报告](../acceptance/reports/windows-seventh-report-2026-10-03.md)确认WebKit11/12，三项任务查询导航各3次通过；剩一次fallback/worker终端的错误来自独立控制台日志轮询。可控基线0/1复现页面离开后新日志fetch，安全API读取现统一取消/等待，仅导航取消的只读请求可在存活页面恢复，body取消不能变成空成功，控制台销毁取消所属读取；输入/写入不重放。87单测、类型构建及本地Chromium23/23、Firefox12/12、WebKit36/36通过，最终PowerShell harness/真实JSON解析和精确选择核验通过。下一次仅`-Followup -BrowsersOnly -WebKitOnly`，五项各3次严格15/15；不重复已通过Go/GCC/SDK/其他浏览器，不计未选择项PASS；修复实效、完整平台/E08及外部环境缺口仍待验证。

第六轮 Windows 增量：[固定f98b078的报告](../acceptance/reports/windows-sixth-report-2026-10-01.md)确认Chromium/Firefox各8/8、WebKit23/24，两项导航查询检查也各3次通过；只剩一次fallback/worker终端的任务列表请求在离开窗口期失败。可控排队轮询基线0/1，新增读取门禁后组合本地Chromium/Firefox各9/9、WebKit27/27，原错误断言保持。下一次仅运行 `-Followup -BrowsersOnly -WebKitOnly`，剩余终端配置加三项导航各3次（严格12/12）；不重复Go、GCC、SDK及其他浏览器，不把本轮未选择项当作新PASS。Windows修复实效仍待实际回传，完整平台/E08和外部环境缺口保留。

第五轮 Windows 增量：[固定3c7abb3的报告](../acceptance/reports/windows-fifth-report-2026-10-01.md)确认11必需原生、3项Go回归各3次通过，Master/runlog race87pass/0fail/4skip；前轮两处Go竞争及账号恢复已获得真机通过证据。Chromium/Firefox各6/6，WebKit16/18，剩两次任务汇总请求诊断。任务读取与共享列表已补离开前取消与AbortSignal；最终本地Chromium/Firefox各8/8、WebKit24/24及80单测/类型检查/构建通过，Windows实效仍待验证。下一次仅运行 `-Followup -BrowsersOnly` 的8/8/24浏览器补测，不再重复Go/GCC/SDK；不代表Windows全范围通过。

第四轮 Windows 增量：[固定08db099的报告](../acceptance/reports/windows-fourth-report-2026-10-01.md)已确认27定向Go/11原生项全部通过、Chromium/Firefox各32/32，原来的目录metadata和ConPTY已有真机PASS。全race两项新竞争已在本地修正并通过受影响模块完整race；浏览器夹具修正后本地Chromium/Firefox各6/6、WebKit18/18通过。新修复仍待 -Followup 真机复测，不代表Windows全范围通过。

UI、完整通透材质及 p95 ≤ 50ms 的目标保持不变。证据只覆盖实际验证范围；用户本次排除的特殊环境不再阻塞交付，不宣称原全环境认证通过。

## 已完成的本地增量

| 项目 | 当前证据及范围 |
| --- | --- |
| 功能回归与发行 | RC34真实完整功能51/51、897.040秒通过；RC33首轮49/51两处扩展点击失败保留，测试补既有原生hover就绪前置条件，业务断言不变。RC34六包2,251条记录/三份Web、独立启动/停机恢复/兼容RC33回退通过，见[本轮报告](../acceptance/reports/terminal-generation-2026-09-29.md)。Windows包不等于Windows运行通过 |
| 24小时稳定性 | [真实长测](../acceptance/reports/local-endurance-24h-2026-09-27.md)86,415.480秒、17,262次采样、1,440分钟槽、两次故障恢复与末尾备份恢复通过；不代替更长期或跨主机/跨平台组合 |
| 跨引擎存储故障 | [Firefox/WebKit](../acceptance/reports/recovery-engines-2026-09-26.md)各3项通过；不代替设备掉电或实际浏览器配额阈值 |
| 终端恢复 | RC27～33补齐UTF-8续接、未完成VT序列、控制状态、鼠标/光标协议、颜色/主题顺序和OSC 8链接。RC33单测76/76、Chromium61/61、Firefox/WebKit各50/50；不声称所有VT扩展已验证 |
| Linux系统管理与限制 | [独立systemd/cgroup](../acceptance/reports/systemd-cgroup-2026-09-29.md)服务/定时器、归属/整组结束、CPU限流、32MiB OOM与8进程上限实际触发通过；停止服务和未加载timer遗漏已修复 |
| Linux原生通知 | [X11/DBus/Dunst](../acceptance/reports/native-notifications-2026-09-29.md)实际绘制/点击/刷新去重/退出后遗留弹窗关闭通过；不覆盖其他桌面或退出后后台投递 |
| R2：防火墙能力门禁 | 状态读取与仅Linux firewalld支持的规则变更已分开；类型检查及定向浏览器5/5通过，见[本地报告](../acceptance/reports/local-remaining-2026-10-05.md) |
| R3：普通Windows真实联调入口 | 新脚本使用真实Master/Daemon/HTTPS/WSS及Vue/Monaco，临时数据与确认清理、独立报告；当前Linux最新源码实跑14/14、安全单测2/2、Windows交叉构建和PowerShell报告harness通过。Windows实际执行不计PASS，见[入口说明](../operations/WINDOWS_VALIDATION.md) |

## 尚未完成

| 项目 | 当前事实 | 下一步动作 |
| --- | --- | --- |
| R1：E08严格延迟 | 完整材质正式Chromium/Firefox/WebKit最近p95为1003.9/1108/545ms，均失败。[诊断](../acceptance/reports/local-diagnostics-2026-09-29.md)证明软件合成瓶颈；审查不能改成PASS | 本地保留UI/负载/50ms目标继续有针对性的优化与测量；获取GPU或所有远端/冷启动组合不是用户前置条件 |
| R5：最终发行（用户延期） | RC34历史发行与51/51保留，后续修复不自动获得同一发行/全范围结论；用户明确要求本轮不做发行 | 发行打包、兼容回退和发布留到用户要求时执行；本轮仍整理当前源码的普通回归和验收证据 |

## 非阻塞历史观察与环境限制

Vim刷新偶发额外字节和[RC30一次撤权超时](../acceptance/reports/terminal-protocol-state-2026-09-29.md)保留原失败与未定位根因。随后原断言多次通过，三子代理终端代次/回放/撤权审查未确认新缺陷，因此不再作为需要无限补测的交付阻塞；若出现新失败附件则重新调查，不能称已确定根因修复。

物理设备、实际宿主磁盘、公网多主机、所有通知/IME/桌面组合及更长期运行未实测的事实继续保留在[范围报告](../acceptance/reports/scope-review-2026-10-05.md)，状态为本次不强制/部署兼容性待抽验，不是PASS。

## 继续执行条件

本轮已复现并修复旧终端连接解析失败晚返回禁用新连接的问题，三引擎定向组合各4/4通过；不能将它认定为原Vim偶发或撤权超时根因。完整回归session90325已退出0，无活动测试句柄，日志`.local/evidence/rc34/real-full.log`。玻璃遮挡裁剪的像素等价检查未通过，不纳入产品。旧表中的RC23发行、24h仍RUNNING及Linux systemd/通知完全缺环境已不再是当前待办。

本轮R2修复及R3新真实入口已完成本地验证，相关证据见[本轮报告](../acceptance/reports/local-remaining-2026-10-05.md)。R4已由普通Windows实际14/14报告关闭，见[真机报告](../acceptance/reports/windows-device-passed-2026-10-06.md)。R1新增布局隔离探针截图一致但仍约923～953ms，收益不足，未采用，见[性能诊断](../acceptance/reports/render-layout-diagnostic-2026-10-05.md)；不降低50ms或改变UI。下一具体工作是R1有针对性的延迟诊断优化；R5用户明确延期。无活动测试/服务/容器和会话接管。无需危险环境，不重复已关闭补测组，不将Linux或静态审查冒充Windows实跑，不称本次交付全部完成。
