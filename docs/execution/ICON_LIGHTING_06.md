# 图标光影对照 06

日期：2026-09-22。

用户认可模糊通透的改进，但认为主体的渐变光影夸张。本轮保留 f06 的几何、配色、透明度（主体 .78 / 纸张 .87）及层间模糊（1.9），单独比较光影处理。

| 候选 | 处理 | 放大 | 桌面 |
| --- | --- | --- | --- |
| A1 / l01 | 去除光斑与底部压暗；顶部混白从 .23 降至 .065，亮边减弱 | [对照](../../web/ui-screenshots/icon-lighting-06/l01-detail.png) | [桌面](../../web/ui-screenshots/icon-lighting-06/l01-desktop.png) |
| A2 / l02 | 去除面内渐变和光斑，仅保留 0.7 单位宽的薄亮边 | [对照](../../web/ui-screenshots/icon-lighting-06/l02-detail.png) | [桌面](../../web/ui-screenshots/icon-lighting-06/l02-desktop.png) |
| B / l03 | 不生成任何明暗渐变、光斑、亮边或笔身高光；保留前序图层模糊与透色 | [对照](../../web/ui-screenshots/icon-lighting-06/l03-detail.png) | [桌面](../../web/ui-screenshots/icon-lighting-06/l03-desktop.png) |

[重点放大](../../web/ui-screenshots/icon-lighting-06/focus.png) · [主体对比](../../web/ui-screenshots/icon-lighting-06/comparison-foreground.png)。

填色与描边的透明度改为部件整体合成，消除二者重叠产生的额外暗边。B 版仍可能呈现下层纸张模糊后的颜色过渡，这是保留的透叠内容。

## 复现与验证

- 本机开发页面：`/icon-lighting.html`；`?focus=1` 查看实例中心、文件夹与终端；`?variant=l01` / `l02` / `l03` 查看完整图标和小尺寸。
- 实际桌面使用开发参数 `?icon-study=l01` / `l02` / `l03`。默认图标不切换。
- `cd web && npm run build`：类型检查及生产构建通过，保留已有的大 chunk 提示。
- `node .local/capture-icon-lighting.mjs`：15 张 2× PNG；3 版桌面、启动器、Dock 均渲染，无 pageerror。
- 截图脚本核实 B 版 SVG 渐变数为 0、光斑及亮边数为 0，Gaussian blur 仍存在，全部候选底板颜色相同。
- 视觉查看重点放大及完整启动器，确认纸张透色保留、实例面板方向光已去除或减弱。
- 全部为本地 Vue / SVG 代码及浏览器截图，无图片生成。视觉评审使用本地模拟 session / 空列表，不作为后端验收。

下一步依据用户对 A1 / A2 / B 的选择收敛主体光影。

