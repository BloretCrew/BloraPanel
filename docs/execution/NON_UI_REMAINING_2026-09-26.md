# UI 暂停后的剩余工作核对

更新：2026-10-01，补充用户 Windows 回传报告及本地修复，当前发行仍为 RC34（当时完整功能回归51/51通过）。保留原路径供已有链接使用，以下为当前状态；早期过程和失败记录见[执行进度](PROGRESS.md)、[验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)和 Git 历史。

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
| Windows运行 | [第四轮真机报告](../acceptance/reports/windows-fourth-report-2026-10-01.md)：27/27定向Go、11/11必需原生项、Chromium/Firefox各32/32通过，便携GCC及全race实际执行成功；全race仍有2失败/14跳过，WebKit29通过/3失败。本地修复和最终跨引擎定向回归已通过，新修复仍待真机验证 | 用户Windows设备运行新版[补测脚本](../operations/WINDOWS_VALIDATION.md)的 -Followup 并回传ZIP；原生系统通知、服务/任务变更及防火墙回滚等仍须对应隔离验收。Linux容器不能替代Windows |
| 跨主机运维 | 本机Docker/Compose、loopback TLS、私有netem、注册源两版本/CA轮换已有证据 | 独立主机/Engine/网络/注册源，补跨主机长时故障与证书运维；loopback不是远端验证 |
| 设备故障 | ENOSPC、SIGKILL、SQLite重开、归档/恢复/调度不重放已有证据 | 可回滚的一次性环境，补写入中断电、缓存丢失、Engine主机存储耗尽；不能触碰生产磁盘或以进程退出替代掉电 |
| 其他桌面与长期组合 | Linux X11和本地24h已覆盖，其他平台/更长期组合未覆盖 | 对应隔离环境及实际持续时间；模拟时钟/短测不能替代 |

## 继续执行条件

本轮已复现并修复旧终端连接解析失败晚返回禁用新连接的问题，三引擎定向组合各4/4通过；不能将它认定为原Vim偶发或撤权超时根因。完整回归session90325已退出0，无活动测试句柄，日志`.local/evidence/rc34/real-full.log`。玻璃遮挡裁剪的像素等价检查未通过，不纳入产品。旧表中的RC23发行、24h仍RUNNING及Linux systemd/通知完全缺环境已不再是当前待办。

取得新复现、可区分性能假设或缺失环境后继续对应实现和定向验证。不要重复已通过测试、重复打包、无限扩展VT功能或把环境缺口记作通过。完整验收仍未关闭。
