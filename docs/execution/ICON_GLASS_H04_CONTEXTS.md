# h04 透明包边：实际界面场景评审

日期：2026-09-22。

用户选定第八轮 h04「透明包边」，要求将该方向扩展到多种实际情况。本轮只扩展使用场景，保持 h04 的原图形、配色、材质参数和包边厚度。禁止图片生成、斜向反光带、扫光条和面内大块方向光影的约束继续有效。

## 实现与入口

- `AppIcon.vue` 在已有开发评审参数基础上增加 `icon-surface=bare`，将用户截图中的独立轮廓用于真实桌面、启动器、Dock、窗口标题栏和应用市场。保留底板时仍使用原有 h04。
- `Desktop.vue` 增加仅开发评审可用的 `icon-background=graphite|warm`；只在 `icon-study=h04` 下生效，分别提供石墨色、暖灰色纯色背景。未更改默认壁纸或保存的工作区偏好。
- 浅色独立轮廓：`/?icon-study=h04&icon-surface=bare`。
- 深底独立轮廓：`/?icon-study=h04&icon-surface=bare&icon-background=graphite`，在设置中切换深色主题。
- 暖灰独立轮廓：`/?icon-study=h04&icon-surface=bare&icon-background=warm`。
- 保留底板对照：`/?icon-study=h04`。
- 默认生产图标未切换；本轮没有调整玻璃渲染器或引入新的材质方向。

## 截图

[24 张截图浏览页](../../web/ui-screenshots/glass-h04-contexts/index.html) · [捕获记录](../../web/ui-screenshots/glass-h04-contexts/captures.json)

也可在开发服务访问 `/ui-screenshots/glass-h04-contexts/index.html`。浏览页支持按场景筛选，点击图片打开原始分辨率。

| 序号 | 场景 | 图片 |
|---|---|---|
| 01 | 原壁纸、独立轮廓桌面 | [查看](../../web/ui-screenshots/glass-h04-contexts/01-desktop.png) |
| 02 | 完整桌面与应用启动器 | [查看](../../web/ui-screenshots/glass-h04-contexts/02-launcher.png) |
| 03 | 13 个应用的启动器细节 | [查看](../../web/ui-screenshots/glass-h04-contexts/03-launcher-detail.png) |
| 04 | 实例中心列表、小尺寸标题栏 | [查看](../../web/ui-screenshots/glass-h04-contexts/04-instances-list.png) |
| 05 | 实例卡片视图 | [查看](../../web/ui-screenshots/glass-h04-contexts/05-instances-cards.png) |
| 06 | 文件管理器图标视图、叠放窗口 | [查看](../../web/ui-screenshots/glass-h04-contexts/06-files.png) |
| 07 | 文件管理器与编辑器工作现场 | [查看](../../web/ui-screenshots/glass-h04-contexts/07-editor-workspace.png) |
| 08 | 多窗口上展开启动器 | [查看](../../web/ui-screenshots/glass-h04-contexts/08-launcher-over-windows.png) |
| 09 | 应用市场演示目录 | [查看](../../web/ui-screenshots/glass-h04-contexts/09-marketplace.png) |
| 10 | 节点管理与任务中心 | [查看](../../web/ui-screenshots/glass-h04-contexts/10-nodes-tasks.png) |
| 11 | 多应用运行时 Dock 局部 | [查看](../../web/ui-screenshots/glass-h04-contexts/11-running-dock.png) |
| 12 | 工作区设置与账号菜单 | [查看](../../web/ui-screenshots/glass-h04-contexts/12-settings-account.png) |
| 13 | 深色桌面、石墨背景 | [查看](../../web/ui-screenshots/glass-h04-contexts/13-dark-desktop.png) |
| 14 | 深色启动器完整桌面 | [查看](../../web/ui-screenshots/glass-h04-contexts/14-dark-launcher.png) |
| 15 | 深色启动器局部 | [查看](../../web/ui-screenshots/glass-h04-contexts/15-dark-launcher-detail.png) |
| 16 | 深色实例中心与应用市场 | [查看](../../web/ui-screenshots/glass-h04-contexts/16-dark-workspace.png) |
| 17 | 暖灰背景启动器 | [查看](../../web/ui-screenshots/glass-h04-contexts/17-warm-desktop.png) |
| 18 | 1366 × 768 笔记本上的实例窗口 | [查看](../../web/ui-screenshots/glass-h04-contexts/18-laptop-instances.png) |
| 19 | 1024 × 768 较窄桌面 | [查看](../../web/ui-screenshots/glass-h04-contexts/19-compact-launcher.png) |
| 20 | 较大字号、图文比例 | [查看](../../web/ui-screenshots/glass-h04-contexts/20-large-type.png) |
| 21 | 保留底板、浅色桌面 | [查看](../../web/ui-screenshots/glass-h04-contexts/21-tiled-light.png) |
| 22 | 保留底板、浅色启动器局部 | [查看](../../web/ui-screenshots/glass-h04-contexts/22-tiled-launcher-detail.png) |
| 23 | 保留底板、深色桌面 | [查看](../../web/ui-screenshots/glass-h04-contexts/23-tiled-dark.png) |
| 24 | 保留底板、深色启动器局部 | [查看](../../web/ui-screenshots/glass-h04-contexts/24-tiled-dark-detail.png) |

## 验证与边界

- `web/npm run build` 通过，包含 Vue/TypeScript 类型检查；仍有既有大块产物提醒。
- `node .local/capture-glass-contexts.mjs` 最终完整运行通过，24 张、2 倍像素、零 `pageerror`；每张捕获前检查所有光学图标 `data-ready=true`。
- 24 张 PNG 文件头/尺寸及浏览页对应条目已核实。主视口为 1600 × 1000，另外包含 1366 × 768、1024 × 768 和原生元素局部截图。
- 首轮在完成前 12 张后因脚本精确匹配“主题”标签失败；改为匹配设置面板对应标签后完整重跑通过。不是产品主题切换失败。
- 截图均由实际 Vue 组件、窗口系统、编辑器和原有交互渲染；实例、文件、任务、应用市场目录是本地拦截的演示数据。所有 API 写请求被脚本拒绝，WebSocket 关闭，没有连接或操作远程节点，不构成后端或扩展安装验收。
- 应用市场三个演示应用没有自有图标，展示宿主默认通用应用图标；不是新绘制的三套应用图标。
- 暖灰、石墨是用于辨识度对照的纯色评审背景；桌面布局与应用内部布局保持现有实现。
- 包边在 Dock 和启动器尺寸较清楚，在很小的标题栏图标中较弱；深底更突出亮边。本轮将这些真实差异保留给用户评审，没有擅自加强光影。

后续按用户对这些实际场景的反馈细调 h04，不恢复被否定的其他材质方案。
