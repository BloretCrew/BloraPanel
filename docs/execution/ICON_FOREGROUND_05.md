# 图标主体材质评审 05

日期：2026-09-22。

用户指出上一轮主要变化在底板，主体仍是平涂。本轮使用原版 14 个应用的 63 个图形部件，重做主体材质；外部底板、应用构图和桌面布局保持一致。所有图片来自 SVG / Vue 在 Chromium 中的实际渲染，未使用图片生成。

## 实现

- `classic-icon-layers.ts` 保留原版几何，把面板、纸张、笔身、店铺等分别建模。
- `ForegroundAppIcon.vue` 对每个面单独合成前序图层、Gaussian blur 漫射、带色透明层与细边缘。文件夹的实际纸张进入前板的模糊采样；没有把整个图标模糊。
- 柔光位置依据部件调整；笔身添加很轻的内缘。备份箭头等粗线使用 alpha mask，避免 SVG clipPath 忽略 stroke。
- 终端命令符号、细线和小标记保持清晰。所有方案底板颜色及不透明度相同。
- 开发模式用 `?icon-study=f01` 至 `f06` 查看。默认图标保持现状，未把未经用户选择的候选设为默认。

## 六个方向

| 版本 | 重点 | 放大对照 | 桌面实拍 |
| --- | --- | --- | --- |
| f01 | 细磨砂 | [对照](../../web/ui-screenshots/icon-foreground-05/f01-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f01-desktop.png) |
| f02 | 清透色片 | [对照](../../web/ui-screenshots/icon-foreground-05/f02-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f02-desktop.png) |
| f03 | 柔光雾面 | [对照](../../web/ui-screenshots/icon-foreground-05/f03-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f03-desktop.png) |
| f04 | 克制缎面 | [对照](../../web/ui-screenshots/icon-foreground-05/f04-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f04-desktop.png) |
| f05 | 冷调透光 | [对照](../../web/ui-screenshots/icon-foreground-05/f05-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f05-desktop.png) |
| f06 | 明快叠色 | [对照](../../web/ui-screenshots/icon-foreground-05/f06-detail.png) | [桌面](../../web/ui-screenshots/icon-foreground-05/f06-desktop.png) |

[01–03 主体对比](../../web/ui-screenshots/icon-foreground-05/group-1-foreground.png) · [04–06 主体对比](../../web/ui-screenshots/icon-foreground-05/group-2-foreground.png)。

## 验证

- `cd web && npm run build`：类型检查与生产构建通过；保留已有的大 chunk 提示。
- `node .local/capture-icon-foreground.mjs`：6 个候选实际桌面 / 启动器 / Dock，28 张 2× 截图，无 pageerror。
- 脚本比对候选的底板填色一致；人工查看带底板、去底板、完整 14 应用及 20 / 32 / 48 px 结果。
- 中途发现粗线使用 clipPath 导致线条丢失，已改用 alpha mask，并保留终端符号的清晰原始笔画，重新生成所有截图。
- 桌面截图使用仅供视觉评审的本地 session / 空列表数据，不构成后端功能、跨平台或性能验收。

下一步依据用户对主体材质的反馈收敛，不以本轮视觉预览宣称产品全量完成。

