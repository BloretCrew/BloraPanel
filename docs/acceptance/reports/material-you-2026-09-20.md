# Material You 扁平视觉重构 — 2026-09-20

用户明确要求保留当前布局，将整套样式改为 Material You 扁平风；应用图标要多色，允许参考 Hyprland 常见美化。本轮已更新源码和 `web/dist`，原来的玻璃图标、材质滤镜和背景不再用于产品界面。

## 设计与实现

采用鼠尾草绿、暖白色面和少量暖色点缀。窗口、侧栏、菜单、输入框和选中项通过颜色层级区分；按钮及导航选中项使用胶囊轮廓。参考 [Google Material 3 的色彩、字形和形状系统](https://developer.android.com/develop/ui/compose/designsystems/material3) 与 [end-4 的 Hyprland Material 桌面配置](https://github.com/end-4/dots-hyprland)。使用独立绘制的 SVG，没有复制这些项目的代码或图像。

- 新增 `web/src/theme.css`，集中定义主色、容器色、正文、描边、错误及交互状态角色。当前是固定的浅色主题，未实现壁纸取色或自动深色主题。
- 直接改写 `style.css`、`applications.css`、`DesktopBar.vue` 与 `NotificationTray.vue` 的现有规则，移除玻璃背景、渐变、折射与内高光；窗口和浮层仅保留少量普通投影。没有叠加一层旧主题覆盖补丁。
- `AppIcon.vue` 重绘全部 13 个默认应用、启动器及扩展兜底图标，使用不透明的多色色块；市场使用店铺形象。删除已无用途的 `IconMaterial.vue` 和 `icon-shapes.ts`。
- `DockSurface.vue` 简化为单一不透明色面，保留既有连续曲线与 14px 间距。桌面右侧入口、顶栏三列、底部 Dock、窗口尺寸和侧栏布局不变。
- `wallpaper.svg` 改为配套的平面几何图形；线条工具图标继续使用既有 Lucide。编辑器、终端、文件列表等功能和恢复逻辑不变。

## 验证结果

| 检查 | 结果及边界 |
| --- | --- |
| `cd web && npm run build` | 最终构建退出 0，含 Vue 类型检查与生产构建；原有 Monaco 体积提示仍存在。 |
| `cd web && npm run test:e2e -- --workers=2` | 完整浏览器套件 37/37 通过，1.8 分钟，会话 41153 退出 0。此套件使用 API 替身；不能代替全部真实后端或平台验收。 |
| 本地真实截图 | 隔离 HTTPS Master/双 Daemon 与 Chromium；`capture-ui.mjs`、`capture-dock.mjs` 均退出 0，共 22 张 PNG，没有通过 API 模拟生成截图。市场条目来自本地测试注册源。 |
| 布局比较 | 桌面与手机共 13 组区域，比较位置、宽高与项目数量；0.5 CSS px 容差内无差异。包含顶栏、Dock、桌面图标、启动器、窗口、标签、侧栏及正文区。 |
| Dock 渲染轮廓 | 桌面、悬停、历史项、1280/760/390/320px 视口共 7 种场景；四边均为 14px，曲线距离 13.99647–14.00049 CSS px，宽高畸变与横向越界均为 0。 |
| 色彩配对 | 以最终颜色计算：正文/背景 12.43:1、辅助文字/容器 4.78:1、主按钮 6.64:1、选中项 8.86:1、错误提示 5.01:1。这是列出色对的对比度核对，不代表全界面无障碍认证。 |

初次布局测量在启动器 150ms 入场动画期间抓取了不同帧，只有启动器纵坐标出现 1.78–5.14px 差异。保留该轮的 `layout-*-animated.json`，未将其算作通过。随后修正采样等待：保留改版前的原始样式为 `layout-baseline.css`，在独立浏览器页面重放旧样式，在动画稳定后测量旧/新布局，最终无差异。该基线只用于验收，不进入生产样式；产品无需因采样问题回退。

## 证据与复现

- [桌面整体](../screenshots/2026-09-20-material-you/02-desktop.png)
- [Dock 图标近景](../screenshots/2026-09-20-material-you/01-dock.png)
- [全部应用图标](../screenshots/2026-09-20-material-you/02c-app-library-detail.png)
- [实例中心窗口](../screenshots/2026-09-20-material-you/03b-instance-window.png)
- [应用市场](../screenshots/2026-09-20-material-you/09-extensions.png)
- [移动端市场](../screenshots/2026-09-20-material-you/10-mobile-market.png)
- [完整截图目录](../screenshots/2026-09-20-material-you/)、[布局比较](../screenshots/2026-09-20-material-you/layout-comparison.json)、[Dock 几何](../screenshots/2026-09-20-material-you/geometry.json)

运行 `./dist/blora-devfixture --listen 127.0.0.1:9444`，将 `BLORA_E2E_CREDENTIALS` 指向该 fixture 输出的私有凭据文件，然后运行 `node web/capture-ui.mjs` 与 `node web/capture-dock.mjs`。凭据不进入公开报告。布局入口为 `node web/capture-layout.mjs`：先设 `BLORA_LAYOUT_STAGE=before` 并以 `BLORA_LAYOUT_BASELINE_CSS` 引用本轮保存的基线样式，再设 `BLORA_LAYOUT_STAGE=after` 测量当前界面。比较会等待启动器动画结束。

构建、测试和所有截图浏览器已结束；fixture 会话 78272 已停止并退出 0。旧 rc11 安装归档未重新打包。全范围 Windows 真机、独立 systemd、远程及物理故障等未验证项保持原状态，本轮未提升其验收结论。后续视觉调整以本轮扁平主题和截图为准；如需发行安装包，应基于当前源码重新打包并验证。
