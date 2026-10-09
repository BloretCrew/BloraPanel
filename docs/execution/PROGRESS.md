# Blora Panel 当前执行进度

更新：2026-10-09。全部此前逐轮记录保存在[完整历史归档](archive/PROGRESS-before-closeout-2026-10-08.md)。历史结果保留各自源码、环境和失败记录。

## 最新实现任务：本地监听与可信反代来源

按用户要求，Master 默认仍仅监听 `127.0.0.1:8443`，取消必填固定域名来源；自动模式按本地回环请求的 HTTPS 主机/端口校验。新增显式 `--trusted-proxies`（默认不信任任何代理），从可信直连代理读取原始 HTTPS 主机/端口及客户端 IP，API、事件/终端/日志/容器日志 WSS 使用同一来源，登录限流采用解析后的客户端 IP。XFF 右到左校验，未经信任的头忽略，可信头歧义/错误拒绝；保留真正 TLS 与可选固定 origin 的保护。中英 README、Master unit、Nginx 示例已改为本地后端方式。

实际结果：Master 全包竞态 272.171s 退出 0；最终默认 443 端口规范化及信任默认关闭守卫、真实 HTTPS 代理登录/WSS/跨来源拒绝/限流来源复验 1.373s 退出 0；最终 Go vet 和 Linux/Windows 构建通过。完整包回归先于最后端口规范化，最终差异由定向守卫覆盖，不将两次运行混称同一源码全包重跑。98 个 README/运维链接、JSON、命令引号及 diff 检查通过。详见[本地来源与代理验收](../acceptance/reports/local-origin-proxy-2026-10-09.md)。Go 测试自有回环监听已清理；没有运行 Nginx、部署系统服务或开启生产入口，没有待接管会话。

版本边界：这些行为晚于已发行 beta.1，当前按源码构建交付；不修改 beta 标签、附件或重写其验证证据。Linux/Windows 编译不等于新 Windows 真机运行；E08、前端与旧功能范围保持原记录。

## 最新文档任务：三个组件的生产启动与职责

中英 README 已增加独立可见的生产启动章节，分别说明 Master 一次初始化/常驻命令、Daemon 私有配置/一次登记/身份复用、前端由 Master 托管或 Nginx 同源独立部署；包含 Linux systemd 示例、Windows PowerShell 命令及平台前置条件。前端不是第三个必须常驻的 Node 进程，SDK 也不是运行服务。Nginx 示例保留 HTTPS 后端证书校验、全部 `/api/` WSS 与 API 转发和对应内容安全策略；主服务的外部 origin、Daemon 的入口 URL 随部署方式一致变更。

核对当前二进制 `--help`、真实 JSON 配置字段及同源 API/证书/身份代码；87 个中英 README 本地链接、JSON/命令引号和 `git diff --check` 通过。systemd 原样示例校验因尚未安装 `/opt/blora/...` 二进制返回非零；仅将两个示例的可执行文件/工作路径替换为现有本地 `dist` 后，`systemd-analyze verify --man=no` 退出 0，未安装或启动服务。本机没有 Nginx，配置对照官方文档与实际路由审阅，未声称生产代理链路已运行。本次只改文档/示例，beta 标签及附件保持不变；没有部署服务或新增运行会话。

## 最新授权：首个 beta 预发布

用户已选择 GPL-3.0，并委托版本/发行类型，明确暂不正式发行、本人尚未试用。**[v0.1.0-beta.1 已预发布](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1)**，GPL 标识为 `GPL-3.0-only`，固定源码 `b5d0287b49b53cafaec94ee04bdd8b115cf767b4`。六包清单/源码/许可证、三份 Web、完整源码包和七包摘要均核对通过；重复打包字节一致。新包初始化、双节点、SDK 构建签名、beta 自身停机恢复、RC34 → beta 升级及兼容 RC34 快照回退通过。九个远端附件大小/摘要一致，发布状态为非草稿、预发布，未标记正式最新版。详见[beta 验收](../acceptance/reports/beta-0.1.0-2026-10-08.md)。新 beta Windows 包未在真机重跑，原设备证据保留原源码范围。

发行说明见[beta 说明](../releases/v0.1.0-beta.1.md)。新增 `scripts/release-verify.py` 验证六包清单、来源、三份 Web 与完整源码；`package-smoke.py --upgrade-from` 真正由旧程序创建状态后升级，而非用新包创建状态假称升级。发行过程中更新了两项已报告漏洞的锁定依赖，未改变应用功能或重开 E08。

## 本轮边界与处置

用户冻结 UI，接受 E08 极端混合负载约 100ms 的有界收尾，要求完成三项非发行工作：当前源码完整回归、代码/文档整理与 GitHub 交接、Chromium 圆角静态绘制问题修复。**这三项已收尾**。此前非发行阶段未创建发行包或标签；最新授权下 beta 已完成，正式稳定版仍等待本人体验。没有公开部署或触碰生产节点。

原 50ms 合同仍保留：Chromium 38.8ms，Firefox 66ms，WebKit 89～132ms、最后 95ms。不保证每轮 ≤100ms，见[E08 收尾](../acceptance/reports/e08-bounded-closeout-2026-10-08.md)。[10-05 范围调整](../acceptance/reports/scope-review-2026-10-05.md)免除难以提供的物理故障/额外设备组合；免测不记为 PASS。[Windows 普通设备 14 项](../acceptance/reports/windows-device-passed-2026-10-06.md)已有固定源码真机证据，不再请求重复上传。

## 已实现与验证

B00～B09 覆盖 F01～F14、A01～A17 和 E01～E08 本次确认范围：Go Master/Daemon、权限/流撤销、实例生命周期、桌面多窗口多标签、即时工作区恢复、文件/编辑器/终端、任务/监控、备份调度、Docker/Compose、有限主机管理与扩展 SDK/安装生命周期。B10 的构建、运行、开发交接与首个 beta 包验收/预发布完成；正式稳定版等待用户试用。完整编号合同及证据见[验收矩阵](../acceptance/ACCEPTANCE_MATRIX.md)。

本轮修复：

- Chromium 壁纸材质平面与祖先圆角的组合不稳定：只取消该材质伪元素的独立保留，不改窗口形状、材质样本或正文。
- 终端遮挡变化引起工具栏按钮圆角三个像素变化：只在 Chromium 下独立保留终端工具栏绘制，终端内容/尺寸/节点不变。稳定截图前置单独不能修复；正式修复两 DPR × 两轮共 4/4，通过原六场景整屏 RGBA0。
- 恢复 Worker 运行期故障回退：运输/生命周期异常可采用完整前台原子事务；真正 IDB Abort/Quota 仍失败，停止/退出不重新发布旧账号。
- Compose 同步嵌套提交的重复序号：所有所属变更同步串行保护，立即刷新恢复项目，不重放远程操作。
- 共享 Monaco 输入热循环：冷状态观察器在完整 apply 后同步重新绑定，所属已提交历史不可变；保留正文、撤销重做及独立光标/滚动。200 行整块输入约 10～11 秒仍是限制，不称普通交互普遍 100ms。

## 本轮结果与证据

- Go 全包竞态：18 个有测试包通过，退出 0；缺少环境的 opt-in 跳过不算环境通过。
- Linux/Windows 构建、最终前端构建及 TypeScript 检查通过；123/123 单测、SDK 和三个独立参考版本构建通过。
- 完整真实回归第二轮 **49 PASS / 2 FAIL，0 skip / 0 flaky**；迁移/不支持 schema 的原生离线注入修正后 1/1 通过，原万条解压/关闭浏览器/16MiB 原文件续传完整单项 1/1 通过。**51 个场景合并覆盖，不声称一次完整 51/51 全绿**。首轮 46/51 及第二轮失败均保留。
- Chromium 原完整 mock 运行 **175 PASS / 6 FAIL，181 项**。其中 8 项为明确安装未采用适配器的实验，四个文件完整保留并归入显式实验配置；成品默认列表为 173 项，未移动成品守卫。原四个成品失败经修复/独立原场景复验通过，网络与原生 AA 的偶发性仍如实记录；没有放宽像素容差或删减场景。
- 圆角首次 3/3、独立重复光学 24/24；最终完整运行两 DPR 圆角守卫通过。Dock 与局部阴影独立 5/5；最后工具栏修复后的相关光学复验见整合报告。
- Worker 六项守卫三个引擎各 6/6；完整独立编辑器场景 Chromium/Firefox/WebKit 均通过，原 200 行、恢复、独立光标/滚动及零错误断言保留。

详见[本轮整合报告](../acceptance/reports/non-release-closeout-2026-10-08.md)、[圆角报告](../acceptance/reports/chromium-corner-closeout-2026-10-08.md)、[恢复审查](../acceptance/reports/non-release-reconciliation-review-2026-10-08.md)、[编辑器报告](../acceptance/reports/editor-closeout-regression-2026-10-08.md)、[最终光学复核](../acceptance/reports/optics-final-regression-2026-10-08.md)。[本地回归指南](../operations/LOCAL_REGRESSION.md)与自动脚本可复现，失败/skip/flaky 返回非零。

## 接管与下一动作

本轮自有真实 HTTPS fixture、浏览器/测试容器已停止，真实 fixture 清理退出 0；没有触碰其他服务。证据保存在忽略的 `.local` / `web/.local`，凭据、数据库、原始真实日志及浏览器快照不进入 Git。

当前没有待接管的本轮运行会话，自有包验收进程均已停止。首个 beta 已交付，之后由用户实际试用、反馈问题，再决定后续 beta 或正式版；不自动要求重跑已有 Windows 设备验收。后续文档提交不会改写 beta 标签或已验证附件。没有重新开始 E08 优化；新的性能/视觉改进须作为后续任务。
