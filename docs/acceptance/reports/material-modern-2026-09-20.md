# Material 3 Expressive 现代视觉与全局样式整理

日期：2026-09-20。用户要求降低卡通感，更换现代壁纸，保持 Material 3 Expressive 风格及现有布局。本轮已更新源码与 `web/dist`。

## 设计依据与选择

联网查阅 [Google 对 Material 3 Expressive 的研究与设计说明](https://design.google/library/expressive-material-design-google-research)及 [MaterialExpressiveTheme 官方参考](https://developer.android.com/reference/kotlin/androidx/compose/material3/MaterialExpressiveTheme.composable)。官方强调色彩、形状、尺寸、动效和容器的组合，也强调按产品情境调整表现强度。本轮将这一原则适配到管理桌面：采用更中性的色面，把强调色集中在关键操作和选中状态。

保留[上一轮已核对的形状、列表与复选框规范依据](material-refine-2026-09-20.md)。这是自有 Vue/Web 实现与项目配色，不宣称采用 Android 原生组件、官方认证或精确复刻全部移动端尺寸。

## 实施与覆盖

| 区域 | 本轮结果 |
| --- | --- |
| 壁纸、桌面、登录 | 替换太阳/叶片式几何拼图，使用大尺度青灰曲面矢量构图。浅灰背景保持桌面标签可读；不含外部图片、滤镜或模糊。 |
| 色彩与图标 | 主色由鼠尾草绿调整为矿物青绿，正文和容器采用中性灰。图标降低底板及部分前景饱和度，保留蓝、赭、紫、绿等辨识色；品牌图形改用主题角色。 |
| 窗口、标签、侧栏、列表 | 保持位置、尺寸与主要分区，统一既有形状令牌。窗口使用低对比描边，正文、列表及选中项延续上一轮圆角层级。 |
| 节点、任务、容器、应用市场 | 统一圆角色面卡片、局部内边距和分组间隙，清理 Docker 局部遗留的半透明白色样式。 |
| 表单与反馈 | 输入、选择、文件选择按钮、进度条、禁用及焦点状态共用语义色。修正弹窗内输入框和容器同色的问题，焦点显示为单一轮廓。 |
| 菜单与弹窗 | 工作区、账号、通知、重命名及应用内对话框共用形状层级；通知项使用圆角色面，退出项轮廓与其他菜单动作一致。 |
| 编辑器、终端 | 统一宿主外框、工具栏和状态栏。Monaco 正文与 xterm 终端仍保留专业编辑/终端呈现和颜色语义。 |
| 动效 | 增加共享时长与缓动角色，列表和主按钮支持轻微形状反馈；启动器采用短距离弹簧入场。系统减少动画偏好及工作区设置同时约束 JS 动画，CSS 规则覆盖伪元素。 |

截图入口覆盖 13 个默认应用：实例、备份、文件、编辑器、终端、任务、节点、监控、容器、用户权限、设置、市场、系统管理；同时覆盖登录、账号、工作区、通知、创建表单、桌面及移动端。部分截图为无资源或未选择节点的正常入口状态，不代表各功能所有数据状态均逐一做了视觉验收。

第三方沙箱内部由扩展自身渲染，本轮统一其宿主窗口，不改写第三方应用。固定浅色主题、原生下拉选项、Monaco/xterm 的专业内容区属于实现边界；未新增动态壁纸取色或深色主题。

## 验证与证据

| 检查 | 结果 |
| --- | --- |
| 最终构建 | `cd web && npm run build`，会话 29081 退出 0，含类型检查；原有 Monaco 大包提示保留。 |
| 完整浏览器回归 | `npm run test:e2e -- --workers=2`：38/38 通过，1.9 分钟，会话 69214 退出 0。包含新增的系统/工作区减少动态效果用例。 |
| 最终焦点与填充修正复验 | `npm run test:e2e -- tests/browser/control-bar.spec.ts tests/browser/instance-settings.spec.ts --workers=2`：3/3 通过，10.8 秒，会话 76913 退出 0。完整套件在这一最后 CSS 修正前执行，未把它伪报为修正后重跑。 |
| 真实页面截图 | 隔离 HTTPS Master/双 Daemon 与 Chromium；最终 `capture-ui.mjs` 会话 96095 退出 0，加上 Dock 截图共 34 张 PNG。未使用 API 替身生成截图。 |
| 布局对照 | 构建前/后 13 组桌面及手机主要区域，0.5 CSS px 容差内无差异；卡片内边距、圆角与分组间隙的主动微调不属于宏观布局不变的结论。 |
| Dock 轮廓 | 7 场景四边 14px，曲线距离误差小于 0.01px，无横向越界与图标变形。最后输入焦点修正没有改动 Dock。 |
| 色对对比度 | 正文/背景 12.27:1，辅助文字/容器 5.07:1，主按钮 7.12:1，主色选中项 8.82:1，次色选中项 7.55:1；只表示列出色对的计算结果，不是全界面无障碍认证。 |

浏览器回归使用 API 替身；测试中的默认 WSS 代理连接拒绝不作为真实后端验证。真实截图使用独立 9444 服务，市场条目来自隔离的本地测试注册源。

- [桌面及新壁纸](../screenshots/2026-09-20-material-modern/02-desktop.png)
- [实例窗口](../screenshots/2026-09-20-material-modern/03b-instance-window.png)
- [创建表单与焦点](../screenshots/2026-09-20-material-modern/03d-create-dialog.png)
- [图标全览](../screenshots/2026-09-20-material-modern/02c-app-library-detail.png)
- [节点卡片](../screenshots/2026-09-20-material-modern/06-nodes.png)
- [应用市场](../screenshots/2026-09-20-material-modern/09-extensions.png)
- [账号菜单](../screenshots/2026-09-20-material-modern/12-account-detail.png)
- [手机市场](../screenshots/2026-09-20-material-modern/10-mobile-market.png)
- [全部截图](../screenshots/2026-09-20-material-modern/)、[布局结果](../screenshots/2026-09-20-material-modern/layout-comparison.json)、[Dock 结果](../screenshots/2026-09-20-material-modern/geometry.json)

复现：启动 `./dist/blora-devfixture --listen 127.0.0.1:9444`，设置 `BLORA_E2E_CREDENTIALS` 为该会话输出的私有文件，然后运行 `node web/capture-ui.mjs`、`node web/capture-dock.mjs`。布局入口为 `capture-layout.mjs`，由 `BLORA_LAYOUT_STAGE=before/after` 选择两侧采样；应在各自版本构建后执行。

本轮所有构建、浏览器、截图进程已结束；fixture 83928 已停止并退出 0。旧 rc11 安装归档未重打包。Windows 真机、独立 systemd、远程和物理故障等既有外部验收缺口保持原状态，本轮不提升其验证结论。
