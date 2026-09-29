# 未完成VT控制序列恢复修复（2026-09-29）

起点`76f8a55`。在RC27的UTF-8修复后独立检查控制序列，真实复现`ESC [ 31`到达后生成检查点、重建终端，再收到`mR`时显示普通文字`mR`，而不是颜色控制后的`R`。原失败退出1、6.313秒，日志`.local/evidence/rc28/terminal-csi-before.log`保留；没有把UTF-8通过推及全部解析器状态。

## 修改与边界

xterm屏幕序列化不包含尚未完成的CSI/OSC/DCS解析状态。当前只在实际解析器GROUND边界压缩屏幕，未完成指令继续保留在已有有界输出日志中；恢复时从完整屏幕基线按序回放。日志增加受校验的尺寸事件，从而保留未完成指令期间的真实resize，不依赖重新解析猜测或重新实现VT解析器。

- 初次挂载先建立安全基线；旧的纯输出日志仍可读取。
- 尺寸、序号、事件种类和字节/条目预算先全部验证，再修改持久恢复状态；尺寸日志计入预算。
- 原128KiB/128事件预算不放宽。无法在完整边界压缩且超过预算时，明确停止并保留最后受保护记录，不确认未落盘输出。
- Worker屏幕快照返回前，若解析进行中、序号改变或边界改变，就拒绝旧结果，避免覆盖期间保护的新输出。
- 回放仍禁止远端输入应答；仅恢复完成后的live可写输出允许正常终端查询应答。UI、材质和性能阈值未改。

实际边界读取封装在`terminal-parser-state.ts`，使用锁定xterm 6的内部`currentState`。公开序列化API没有这个状态；这里仅检查边界，不复制内部解析器。内部布局或状态值不兼容时明确失败，不能静默生成有损快照。依赖升级需要重新执行真实解析器测试，不能只依赖类型检查。

## 验证证据

除修复前记录外，以下日志均在`.local/evidence/rc28`，包装器保存UTC起止、单调耗时及退出码。

| 场景 | 实际结果 |
| --- | --- |
| 原CSI失败定向复核 | 1/1通过，9.850秒，`terminal-csi-fixed.log` |
| CSI颜色/resize、内嵌换行、光标定位、C1 CSI、OSC标题/链接、DCS查询、字符集半指令及超预算保留 | 9/9通过，21.077秒，`terminal-parser-cases.log` |
| Chromium终端/UTF-8/Worker/原生IDB恢复组合 | 24/24通过，90.286秒，`browser-terminal-regression.log` |
| 实际Worker旧结果延迟返回期间写入新输出，再重建核对正文 | 1/1通过，6.487秒，`terminal-stale-snapshot.log` |
| 前端单测，含resize日志落盘/重开、非法尺寸拒绝及旧格式 | 69/69通过，2.650秒，`unit.log` |
| 生产构建、类型检查 | 通过，构建包装器46.946秒，`web-build.log` |
| 真实HTTPS双节点文件/PTY权限、刷新、移窗、Vim/top不重放 | 7/7通过，74.241秒，`real-terminal.log` |
| Linux Firefox，控制序列/UTF-8/Worker全部21项 | 21/21通过，92.575秒，`firefox-terminal.log` |
| Linux WebKit，同组21项 | 21/21通过，79.382秒，`webkit-terminal.log` |
| 仓库正式配置的Firefox入口，resize与旧快照两项 | 2/2通过，11.590秒，`firefox-public-config.log` |
| 仓库正式配置的WebKit入口，同两项 | 2/2通过，12.498秒，`webkit-public-config.log` |

浏览器组合使用实际xterm、Worker和恢复存储，API请求为隔离数据；真实后端证据单列为7项。三引擎比较完整屏幕序列化结果、尺寸、查询应答和标题事件，与未中断的同一输出流对照，不只检查可见文本。Firefox/WebKit的21项先通过既有临时引擎配置；现已将`BLORA_BROWSER`接入仓库正式配置，并补正式入口定向运行，复现命令见[平台指南](../../operations/PLATFORM_VALIDATION.md#webkit-浏览器)。

## 发行和剩余工作

RC28六包构建通过（47.211秒）；外部摘要、2,223条包内文件记录及三份Web与当前构建一致性核验通过（4.332秒）。`dist/releases/development-20260929-rc28/SHA256SUMS`自身SHA256为`1e7a006ed639d04ad94e33b4ea0e5dc70dd1dc68b012478a6cc9ee4f56467661`。

独立SDK构建/签名、Master初始化/TLS/静态资源/登录、两份发行Daemon上线、停机状态恢复及兼容RC27回退全部通过（10.055秒）。恢复核对TLS、节点身份、只读授权、任务回执和资源正文，快照后的改动不存在；所属进程已停止。命令为`python3 scripts/package-smoke.py dist/releases/development-20260929-rc28 --state-restore --rollback-release dist/releases/development-20260929-rc27`。原始日志及退出0回执为同目录`package-build.log`、`package-verify.log`、`package-smoke.log`和各自`.result.json`。

此修复解决有确定复现的未完成指令丢失，不证明原Vim偶发额外`$y`字节的根因已找到。已完成指令所设置的其他持久终端状态也不能仅凭本组未完成序列测试推定全部覆盖。严格E08仍未达标；Windows、其他桌面平台、跨主机、宿主级资源耗尽和物理故障缺口分别保留。
