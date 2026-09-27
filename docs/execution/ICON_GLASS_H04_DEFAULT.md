# h04 保留底板：设为默认图标

2026-09-22，用户在实际场景对照后选择“保留底版好一些”。将 h04「透明包边＋保留底板」作为默认应用图标，继续保留禁止图片生成、斜向反光带及夸张光影的要求。

## 变更

- 默认覆盖 13 个应用及启动器：桌面入口、Dock、启动器、窗口标题栏、应用市场共用 `AppIcon.vue`。
- 原 h04 图形、色彩、底板与材质参数不变；从已有代码渲染器导出 14 个 288 × 288 图标，总计 185,447 字节。
- 默认界面使用这些预先渲染的资产，避免每个图标初始化或调用实时 WebGL。原有 SVG viewBox 保持图标边界、留白和 Dock 内的比例。
- 未知应用继续使用通用应用图标。开发参数仍可查看既有候选，原平面版本可用 `?icon-study=classic` 对照；正式构建忽略评审参数。
- 默认布局、壁纸、文字样式未调整。本轮没有公开部署或修改后端。

## 验证

- `npm run build` 通过，包含 Vue/TypeScript 类型检查；仅有既有大块产物提醒。
- `node .local/verify-selected-icons.mjs` 针对本地正式构建验证通过。在禁用 WebGL 的 Chromium 中，启动器展开后的 29 个图标引用完整 14 种资产，均成功加载为 288 × 288，零 WebGL 请求、零实时光学 canvas、零 pageerror。
- 验证桌面→启动器→实例中心交互，并确认正式构建忽略 `icon-study=h03&icon-surface=bare`，仍显示带底板 h04。
- 五张真实界面截图及[验证数据](../../web/ui-screenshots/h04-default/verification.json)保存在 `web/ui-screenshots/h04-default/`。实例数据为本地演示数据，此项不增加后端验收证据。
- [桌面](../../web/ui-screenshots/h04-default/01-desktop.png)、[启动器](../../web/ui-screenshots/h04-default/02-launcher.png)、[启动器细节](../../web/ui-screenshots/h04-default/03-launcher-detail.png)、[Dock](../../web/ui-screenshots/h04-default/04-dock.png)、[实例中心](../../web/ui-screenshots/h04-default/05-instances.png)。

后续视觉细化以本版为基线。
