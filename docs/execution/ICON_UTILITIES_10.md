# 01 / 04 图标细化：终端与监控六组预览

2026-09-23。用户接受上一轮01「应用原色」和04「彩色薄片」的方向，要求改善终端、监控的配色，并重新设计终端主体图形。本轮在这两套基础上各提供三种方向。

## 设计与实现

| 方向 | 终端造型与配色 | 监控配色 |
| --- | --- | --- |
| A 银白窗口 | 银白底板、灰色窗口、顶栏与输出行、蓝色光标 | 绿白面板、暖色采样点 |
| B 紫色叠片 | 两片错位窗口、紫色主体、米橙色光标 | 蓝白面板、橙色采样点 |
| C 浅色命令页 | 白色命令页、蓝色顶栏、块状光标与两行输出 | 米白面板、暖橙折线 |

- 每个方向分别沿用01的鲜明颜色与04的柔和颜色，形成01A/01B/01C/04A/04B/04C六组。终端三种主体全部是新代码绘制路径；监控保持原图形并更换各层颜色。
- 继续使用用户已接受的 h04 透明包边、底板和冷白冰蓝桌面。没有图片生成或新增面内斜向光条。其余12个图标在每组中与对应01/04基础资产逐字节一致。
- 新 [`icon-utility-variants.ts`](../../web/src/app-host/icon-utility-variants.ts)集中定义三种终端几何和两套颜色。`IconPalette`增加可选 `artwork`，共享渲染器只在指定应用上替换局部路径，不修改原图形集合。
- 六组各14张288×288预渲染PNG，共84张、1,095,755字节；正常预览不依赖WebGL。默认图标、已有候选及正式构建行为保留。
- 开发入口：`/?appearance=ice&icon-colorway=spectrum-silver|spectrum-layers|spectrum-paper|pastel-silver|pastel-layers|pastel-paper`。仍只在开发预览启用。

## 实际效果

[对照页](../../web/ui-screenshots/icon-utilities/index.html)提供两组「上一版＋A/B/C」放大图和实际Dock并排，也可切换全部桌面、启动器、图标特写、Dock、实例中心截图。

- [01三组对照](../../web/ui-screenshots/icon-utilities/00-spectrum-comparison.png)
- [04三组对照](../../web/ui-screenshots/icon-utilities/00-pastel-comparison.png)

| 版本 | 启动器实景 | 图标特写 | Dock |
| --- | --- | --- | --- |
| 01A | [查看](../../web/ui-screenshots/icon-utilities/spectrum-silver-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-silver-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-silver-04-dock.png) |
| 01B | [查看](../../web/ui-screenshots/icon-utilities/spectrum-layers-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-layers-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-layers-04-dock.png) |
| 01C | [查看](../../web/ui-screenshots/icon-utilities/spectrum-paper-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-paper-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/spectrum-paper-04-dock.png) |
| 04A | [查看](../../web/ui-screenshots/icon-utilities/pastel-silver-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-silver-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-silver-04-dock.png) |
| 04B | [查看](../../web/ui-screenshots/icon-utilities/pastel-layers-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-layers-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-layers-04-dock.png) |
| 04C | [查看](../../web/ui-screenshots/icon-utilities/pastel-paper-02-launcher.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-paper-03-icons.png) | [查看](../../web/ui-screenshots/icon-utilities/pastel-paper-04-dock.png) |

## 验证与边界

- `cd web && npm run build`最终退出0，包含Vue/TypeScript检查，仍有既有大块产物提示。正式构建只输出14张默认图标PNG；预览几何数据也可由构建裁剪，未增加运行时渲染要求。
- `node .local/export-selected-icons.mjs --colorway=<版本>`六组导出均通过。PNG逐字节比较确认每组仅 `terminal.png` 和 `monitor.png` 与基础版不同。
- `node .local/capture-icon-colorways.mjs --utilities`退出0，禁用WebGL的Chromium产生30张真实界面/局部截图和2张对照图。每组核对13个不同已安装应用图标加载，桌面、启动器、实例中心三类场景的布局测量在六组间一致，零pageerror。对照页切换到04C的Dock并成功解码图片。原始结果见[verification.json](../../web/ui-screenshots/icon-utilities/verification.json)。
- 截图账号与窗口内容使用本地演示数据，拒绝写API且关闭WebSocket；不计后端或真实PTY验收。本轮没有远程资源操作或公开部署。
- 导出、截图、构建与临时Vite会话均已结束，无运行服务需接管。

下一动作：依据用户选择细化终端造型和监控颜色；01与04基础版继续保留。
