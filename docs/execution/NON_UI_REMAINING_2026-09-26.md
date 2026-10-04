# UI 暂停后的剩余工作核对

更新：2026-10-04，补充用户 Windows 回传报告及本地修复，当前发行仍为 RC34（当时完整功能回归51/51通过）。保留原路径供已有链接使用，以下为当前状态；早期过程和失败记录见[执行进度](PROGRESS.md)、[验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)和 Git 历史。

第十轮 Windows 增量：[固定3ef289a的报告](../acceptance/reports/windows-tenth-report-2026-10-04.md)确认正式构建通过、WebKit12/15，终端3次通过；剩两次控制台点击稳定等待45秒超时（尚无日志读取），及一次无帧汇总5秒断言失败（之后有HTTP完成）。Windows具体绘制/数据发布原因仍未确立。安全读取恢复后现立即重查活动任务查询，保留确认缓存，不重放写入；日志生命周期测试以原生Enter键打开同一实际UI，明确不计这三步指针稳定覆盖。所有查询/日志补实时阶段和有界数字计数/按钮几何诊断。90单测/类型构建、WebKit36/36、Firefox12/12、Chromium12/12及脚本harness/真实JSON/严格15次选择核验通过；探索暂停时钟未复现指针停顿，独立保留，不冒充修复因果。下一次仍`-Followup -BrowsersOnly -WebKitOnly`五项各3次且构建PASS、15/15，不重复Go/GCC/SDK/其他引擎，不导入旧PASS；UI冻结，Windows实效和原生/E08/外部缺口保持。

第九轮 Windows 增量：[固定3e6c31c的报告](../acceptance/reports/windows-ninth-report-2026-10-03.md)确认WebKit14/15，四项导航/日志用例各3次通过，前轮两项超时已获真机通过。仅第一次终端耗尽原45秒总时限，之后两次约19秒通过，无页面/请求错误；首次刷新约32秒才开始，不能从上传确认具体慢动作/Windows因果。补测改为先单独构建正式frontend，再在fresh preview运行，新增实时阶段耗时和有界静态资源诊断，保留原断言/时限，UI不改。下一次仍是`-Followup -BrowsersOnly -WebKitOnly`五项各3次，必须构建通过且15/15；不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS。本地验证详情见报告，Windows实效和完整原生/E08/外部环境缺口保持。

第八轮 Windows 增量：[固定57d4fb6的报告](../acceptance/reports/windows-eighth-report-2026-10-03.md)确认WebKit13/15，原终端配置3次通过，剩两次读取恢复/测试建立超时。可控无帧基线0/1复现汇总恢复超时，API只读门禁现增加250ms独立恢复兜底，pagehide阻止旧帧/旧截止恢复，新导航重置代次；日志测试用真实interval回调建立待取消读取，queued观察不依赖绘制，原限时/错误/ACK/输入不重放保持。89单测、类型构建及本地WebKit36/36、Firefox12/12、Chromium19/19加独立补验4/4、脚本harness/真实JSON和精确15次选择通过。下一次仍仅`-Followup -BrowsersOnly -WebKitOnly`五项各3次，不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS；Windows实效、原生/E08/外部环境缺口保持。

第七轮 Windows 增量：[固定8ad386a的报告](../acceptance/reports/windows-seventh-report-2026-10-03.md)确认WebKit11/12，三项任务查询导航各3次通过；剩一次fallback/worker终端的错误来自独立控制台日志轮询。可控基线0/1复现页面离开后新日志fetch，安全API读取现统一取消/等待，仅导航取消的只读请求可在存活页面恢复，body取消不能变成空成功，控制台销毁取消所属读取；输入/写入不重放。87单测、类型构建及本地Chromium23/23、Firefox12/12、WebKit36/36通过，最终PowerShell harness/真实JSON解析和精确选择核验通过。下一次仅`-Followup -BrowsersOnly -WebKitOnly`，五项各3次严格15/15；不重复已通过Go/GCC/SDK/其他浏览器，不计未选择项PASS；修复实效、完整平台/E08及外部环境缺口仍待验证。

第六轮 Windows 增量：[固定f98b078的报告](../acceptance/reports/windows-sixth-report-2026-10-01.md)确认Chromium/Firefox各8/8、WebKit23/24，两项导航查询检查也各3次通过；只剩一次fallback/worker终端的任务列表请求在离开窗口期失败。可控排队轮询基线0/1，新增读取门禁后组合本地Chromium/Firefox各9/9、WebKit27/27，原错误断言保持。下一次仅运行 `-Followup -BrowsersOnly -WebKitOnly`，剩余终端配置加三项导航各3次（严格12/12）；不重复Go、GCC、SDK及其他浏览器，不把本轮未选择项当作新PASS。Windows修复实效仍待实际回传，完整平台/E08和外部环境缺口保留。

第五轮 Windows 增量：[固定3c7abb3的报告](../acceptance/reports/windows-fifth-report-2026-10-01.md)确认11必需原生、3项Go回归各3次通过，Master/runlog race87pass/0fail/4skip；前轮两处Go竞争及账号恢复已获得真机通过证据。Chromium/Firefox各6/6，WebKit16/18，剩两次任务汇总请求诊断。任务读取与共享列表已补离开前取消与AbortSignal；最终本地Chromium/Firefox各8/8、WebKit24/24及80单测/类型检查/构建通过，Windows实效仍待验证。下一次仅运行 `-Followup -BrowsersOnly` 的8/8/24浏览器补测，不再重复Go/GCC/SDK；不代表Windows全范围通过。

第四轮 Windows 增量：[固定08db099的报告](../acceptance/reports/windows-fourth-report-2026-10-01.md)已确认27定向Go/11原生项全部通过、Chromium/Firefox各32/32，原来的目录metadata和ConPTY已有真机PASS。全race两项新竞争已在本地修正并通过受影响模块完整race；浏览器夹具修正后本地Chromium/Firefox各6/6、WebKit18/18通过。新修复仍待 -Followup 真机复测，不代表Windows全范围通过。

UI、完整通透材质及 p95 ≤ 50ms 的目标保持不变。证据只覆盖实际验证范围，不代表全范围完成。

## 已完成的本地增量

| 项目 | 当前证据及范围 |
| --- | --- |
| 功能回归与发行 | RC34真实完整功能51/51、897.040秒通过；RC33首轮49/51两处扩展点击失败保留，测试补既有原生hover就绪前置条件，业务断言不变。RC34六包2,251条记录/三份Web、独立启动/停机恢复/兼容RC33回退通过，见[本轮报告](../acceptance/reports/terminal-generation-2026-09-29.md)。Windows包不等于Windows运行通过 |
| 24小时稳定性 | [真实长测](../acceptance/reports/local-endurance-24h-2026-09-27.md)86,415.480秒、17,262次采样、1,440分钟槽、两次故障恢复与末尾备份恢复通过；不代替更长期或跨主机/跨平台组合 |
| 跨引擎存储故障 | [Firefox/WebKit](../acceptance/reports/recovery-engines-2026-09-26.md)各3项通过；不代替设备掉电或实际浏览器配额阈值 |
| 终端恢复 | RC27～33补齐UTF-8续接、未完成VT序列、控制状态、鼠标/光标协议、颜色/主题顺序和OSC 8链接。RC33单测76/76、Chromium61/61、Firefox/WebKit各50/50；不声称所有VT扩展已验证 |
| Linux系统管理与限制 | [独立systemd/cgroup](../acceptance/reports/systemd-cgroup-2026-09-29.md)服务/定时器、归属/整组结束、CPU限流、32MiB OOM与8进程上限实际触发通过；停止服务和未加载timer遗漏已修复 |
| Linux原生通知 | [X11/DBus/Dunst](../acceptance/reports/native-notifications-2026-09-29.md)实际绘制/点击/刷新去重/退出后遗留弹窗关闭通过；不覆盖其他桌面或退出后后台投递 |

## 尚未完成

| 项目 | 当前事实 | 下一步条件 |
| --- | --- | --- |
| E08严格延迟 | 完整材质正式Chromium/Firefox/WebKit p95为1003.9/1108/545ms，均失败。[诊断](../acceptance/reports/local-diagnostics-2026-09-29.md)表明软件合成为主要瓶颈，显式SwiftShader更慢未采用 | 保持原UI/负载/阈值，验证可区分热点的新方案，或取得硬件合成设备补测。再次检查本机无`/dev/dri`；不保证硬件环境一定达标 |
| Vim刷新偶发额外字节 | RC25曾失败，严格逐字节复验及后续套件未再现；已补连接/阶段/字节附件 | 取得新失败附件或确定复现，辨明旧连接、回放或恢复后原生查询来源；不放宽断言或过滤合法live应答 |
| 单次权限撤销超时 | [RC30](../acceptance/reports/terminal-protocol-state-2026-09-29.md)首轮6/7，原限时串行复核7/7，后续通过；首次原因未定位 | 再次失败或新可区分证据后定向排查；不把连续通过等同根因修复 |
| Windows运行 | [第十轮真机报告](../acceptance/reports/windows-tenth-report-2026-10-04.md)：正式构建PASS、WebKit12/15；终端3次通过，剩两次控制台指针稳定等待超时和一次无帧任务汇总断言失败，具体Windows原因未确立。立即任务重查/原生键盘日志现场/新增阶段与计数几何诊断已通过本地相关回归，真机实效仍待验证。Chromium/Firefox保持第六轮各8/8；[第五轮](../acceptance/reports/windows-fifth-report-2026-10-01.md)原生/Go/race只作该版本证据；本轮未选择Go/SDK/其他引擎，不导入旧PASS | 用户运行新版[补测脚本](../operations/WINDOWS_VALIDATION.md) `-Followup -BrowsersOnly -WebKitOnly`（正式构建，5项各3次严格15/15）回传ZIP；原生系统通知、服务/任务变更、防火墙回滚及指针/绘制性能仍须隔离验收。Linux容器不能替代Windows |
| 跨主机运维 | 本机Docker/Compose、loopback TLS、私有netem、注册源两版本/CA轮换已有证据 | 独立主机/Engine/网络/注册源，补跨主机长时故障与证书运维；loopback不是远端验证 |
| 设备故障 | ENOSPC、SIGKILL、SQLite重开、归档/恢复/调度不重放已有证据 | 可回滚的一次性环境，补写入中断电、缓存丢失、Engine主机存储耗尽；不能触碰生产磁盘或以进程退出替代掉电 |
| 其他桌面与长期组合 | Linux X11和本地24h已覆盖，其他平台/更长期组合未覆盖 | 对应隔离环境及实际持续时间；模拟时钟/短测不能替代 |

## 继续执行条件

本轮已复现并修复旧终端连接解析失败晚返回禁用新连接的问题，三引擎定向组合各4/4通过；不能将它认定为原Vim偶发或撤权超时根因。完整回归session90325已退出0，无活动测试句柄，日志`.local/evidence/rc34/real-full.log`。玻璃遮挡裁剪的像素等价检查未通过，不纳入产品。旧表中的RC23发行、24h仍RUNNING及Linux systemd/通知完全缺环境已不再是当前待办。

取得新复现、可区分性能假设或缺失环境后继续对应实现和定向验证。不要重复已通过测试、重复打包、无限扩展VT功能或把环境缺口记作通过。完整验收仍未关闭。
