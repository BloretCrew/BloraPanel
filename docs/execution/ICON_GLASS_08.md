# 玻璃通透材质候选 08

日期：2026-09-22。

用户明确否定斜向反光带，并要求从多个材质方向继续比较 Liquid Glass 的通透效果。**持续视觉约束：不再使用斜向反光带、扫光条或面内大块方向光影。**

## 修改

- 从 WebGL shader 删除整段斜向环境反射带及配置参数；旧 g01–g06 的后续渲染同样不再产生该效果。g06 改为“清晰色片”，历史截图留作过程记录。
- 边界反射乘以边缘权重，图形平整中央没有方向性亮斑。新八版仍使用原 14 应用 / 63 个部件。
- 从透色、颜色吸收、边缘折射、透明壳、彩色边缘、夹层位移、乳白混合与平衡壳层八个角度比较。
- 透明包边的色芯使用较清楚的交界，避免宽幅渐变产生模糊观感；细线部件的壳层宽度限制为笔画宽度的 20%，保留备份箭头等图形的可辨识度。
- 增加原底板 / 网格 / 色块观察页。网格和色块是材质测试衬底，正常桌面仍使用原底板。
- 仅开发预览参数生效，默认桌面图标未切换。全部代码渲染，无图片生成。

## 八版与实拍

| 版本 | 方向 | 放大 / 全套 / 小尺寸 | 桌面 |
| --- | --- | --- | --- |
| h01 | 清透色片 | [对照](../../web/ui-screenshots/icon-glass-08/h01-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h01-desktop.png) |
| h02 | 有色玻璃 | [对照](../../web/ui-screenshots/icon-glass-08/h02-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h02-desktop.png) |
| h03 | 边缘透镜 | [对照](../../web/ui-screenshots/icon-glass-08/h03-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h03-desktop.png) |
| h04 | 透明包边 | [对照](../../web/ui-screenshots/icon-glass-08/h04-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h04-desktop.png) |
| h05 | 彩边清玻 | [对照](../../web/ui-screenshots/icon-glass-08/h05-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h05-desktop.png) |
| h06 | 夹层玻璃 | [对照](../../web/ui-screenshots/icon-glass-08/h06-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h06-desktop.png) |
| h07 | 雾白透色 | [对照](../../web/ui-screenshots/icon-glass-08/h07-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h07-desktop.png) |
| h08 | 平衡薄壳 | [对照](../../web/ui-screenshots/icon-glass-08/h08-detail.png) | [桌面](../../web/ui-screenshots/icon-glass-08/h08-desktop.png) |

[01–04 主体](../../web/ui-screenshots/icon-glass-08/group-1-foreground.png) · [05–08 主体](../../web/ui-screenshots/icon-glass-08/group-2-foreground.png) · [03 透色与折射观察](../../web/ui-screenshots/icon-glass-08/h03-transmission.png)。

## 验证

- `npm run build`：类型检查与生产构建通过，保留已有 chunk 大小提示。
- `node .local/verify-icon-glass.mjs`：八版实例面板内部区域在原底板 / 网格 / 色块下实际发生像素变化，RGB 最大差值为 11～63；八版终端的两处平整内部采样点 RGB 差值均为 0。验证的是主体内部，不是底板外围差异。
- 最初单点采样在夹层版恰好避开位移后的细网格线，改为面板内部 17 × 4 视图单位区域；没有降低透色条件。
- 构建与两个浏览器作业并行时出现超时；改为顺序执行后内部像素检查通过。宿主当时负载较高，不作为性能通过证据。
- `node .local/capture-icon-glass.mjs`：46 张 2× 截图，覆盖八版桌面 / 启动器 / Dock、全套图标、32 / 48 / 64 px、去底板、深浅底、三种材质观察衬底。确认所有 canvas 渲染成功且无 pageerror。
- 视觉评审使用本地 session / 空列表，非后端验收。WebGL 材质仍是定制网页候选，未做跨平台或低端设备性能验收。

开发入口：`/icon-glass.html?group=1` / `group=2`；单版 `?variant=h03` 等；实际桌面 `/?icon-study=h01` 至 `h08`。下一动作：依据用户选择收敛材质，持续遵守禁止斜向反光带的要求。

