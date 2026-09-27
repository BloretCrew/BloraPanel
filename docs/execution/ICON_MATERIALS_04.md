# 原版图标材质评审（第四轮）

用户明确以最初“卡通化”Material图标为基础，要求适度通透、高斯模糊材质，而不是功能线条图标或过度拟物。沿用禁止图片生成的要求，全部以 SVG/CSS 实现。前三轮候选未设为默认；本轮也仅为开发预览。

## 实现

`ClassicIconArtwork.vue` 从默认 `AppIcon.vue` 提取原有14类应用的图形路径，形状、位置与原始填色保持不变，单次源码比较确认两者几何完全一致。`MaterialAppIcon.vue` 在原图形上增加可配置的 SVG Gaussian blur 色彩漫射底层、CSS backdrop-filter、透明度、极浅色阶和边缘分离。所有前景图形保持清晰，不对整个应用图标高斯模糊。12种配置见 `icon-materials.ts`，含亮、暗和不同透明度对照；不是原生 iOS/Windows 材质实现。

访问开发页 `/icon-materials.html`，支持壁纸、浅底、深底切换；`?group=1|2|3` 每组四套；`?variant=m01` 到 `m12` 查看详细对照。实际桌面使用 `/?icon-study=m01` 到 `m12`。默认样式和布局不变。

## 官方资料

- [Apple App icons](https://developer.apple.com/design/human-interface-guidelines/app-icons?country=63)：采用前景/背景分层和受控透明度；系统材质由原生系统施加，本项目仅以Web技术探索相应视觉效果。
- [Microsoft Windows app icon design](https://learn.microsoft.com/en-us/windows/apps/design/iconography/app-icon-design)：正面平面叠层，限制隐喻数量，克制渐变和细节。
- [Android adaptive icons](https://developer.android.com/develop/ui/compose/system/icon_design_adaptive)：前景/背景层分离、统一遮罩和清晰轮廓。

## 验证与截图

修正 `feColorMatrix.values` SVG属性的字符串类型后，`npm run build`（含 vue-tsc）通过。Chromium截图脚本 `web/.local/capture-icon-materials.mjs` 使用固定会话数据和关闭的测试WebSocket，未连接真实节点。12套实际桌面与启动器均无 pageerror；截图53张，2倍像素。该检查不是完整性能或跨浏览器验收。

总览：[浅色总览](../../web/ui-screenshots/icon-materials-04/overview.png)、[深底总览](../../web/ui-screenshots/icon-materials-04/overview-dark.png)。

|方案|原版对照 / 大小 / 深浅底|桌面|
|---|---|---|
|01 原色柔雾|[查看](../../web/ui-screenshots/icon-materials-04/m01-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m01-desktop.png)|
|02 清玻璃|[查看](../../web/ui-screenshots/icon-materials-04/m02-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m02-desktop.png)|
|03 有色磨砂|[查看](../../web/ui-screenshots/icon-materials-04/m03-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m03-desktop.png)|
|04 轻薄叠层|[查看](../../web/ui-screenshots/icon-materials-04/m04-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m04-desktop.png)|
|05 缎面白|[查看](../../web/ui-screenshots/icon-materials-04/m05-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m05-desktop.png)|
|06 乳白漫射|[查看](../../web/ui-screenshots/icon-materials-04/m06-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m06-desktop.png)|
|07 清晰色片|[查看](../../web/ui-screenshots/icon-materials-04/m07-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m07-desktop.png)|
|08 冷暖微光|[查看](../../web/ui-screenshots/icon-materials-04/m08-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m08-desktop.png)|
|09 烟色磨砂|[查看](../../web/ui-screenshots/icon-materials-04/m09-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m09-desktop.png)|
|10 薄烟玻璃|[查看](../../web/ui-screenshots/icon-materials-04/m10-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m10-desktop.png)|
|11 半透明主体|[查看](../../web/ui-screenshots/icon-materials-04/m11-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m11-desktop.png)|
|12 平衡磨砂|[查看](../../web/ui-screenshots/icon-materials-04/m12-detail.png)|[查看](../../web/ui-screenshots/icon-materials-04/m12-desktop.png)|

每套还生成同目录下的 `mNN-launcher.png` 和 `mNN-dock.png`；`group-1.png`、`group-2.png`、`group-3.png` 为四套一组的对照。
