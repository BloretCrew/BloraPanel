# 应用图标独立配色：四版实际界面预览

2026-09-23。用户认可上轮 h04 图标造型和质感，但指出冷白冰蓝版的应用图标仍接近统一蓝色主题，要求参考 iOS 系统应用的独立识别色并提供多个实装版本。

## 设计依据和选择

- [Apple App icons](https://developer.apple.com/design/human-interface-guidelines/app-icons)强调图标有清晰的背景与前景层次，并在尺寸变化时保持辨识度；[Adopting Liquid Glass](https://developer.apple.com/documentation/technologyoverviews/adopting-liquid-glass)建议用简化、填充、可叠合的图形，材质效果交给层处理。
- 本轮沿用已获认可的 h04 路径、透明包边和底板轮廓，把配色作为应用身份处理：实例蓝、备份绿、文件橙、编辑紫红、市场珊瑚红，同时保留暗色终端和冷白桌面。不是复制 Apple 的具体图标或应用商标。
- 四版：`spectrum` 应用原色，饱和度较高；`porcelain` 浅底彩芯，多数底板较浅而主体更鲜明；`contrast` 明暗交错，多个深色与浅色底板交替；`pastel` 彩色薄片，底板柔和，前景仍分别着色。四版均按14个应用逐个指定背景和每层主体颜色，并非全图叠同一个滤镜。无图片生成、额外渐变扫光或新几何。
- 与现有冷白冰蓝版共用同一桌面、窗口、Dock 配色和 Material 形状，因此对照只呈现图标颜色差异。本轮未替换默认主题或默认 h04 图标。

## 实现与入口

- [`icon-colorways.ts`](../../web/src/app-host/icon-colorways.ts)定义四组逐应用的颜色映射。原图形数据仍来自 `classic-icon-layers.ts`，h04 材质参数不变。
- `.local/export-selected-icons.mjs --colorway=<名称>`从本地代码渲染器导出14个288×288 PNG。四组56张资产共734,796字节。运行时预览使用静态图片，不需要WebGL。
- 仅开发服务下使用 `/?appearance=ice&icon-colorway=spectrum|porcelain|contrast|pastel`。当前默认版和之前的 `/?appearance=ice` 保留；色彩候选由本地开发服务器直接加载，正式构建只打包默认的14张图标。
- [交互对照页](../../web/ui-screenshots/icon-colorways/index.html)可切换旧冰蓝＋四个方向，以及桌面、启动器、图标特写、Dock、实例中心。并排总览是[00-all-icons.png](../../web/ui-screenshots/icon-colorways/00-all-icons.png)。

| 方案 | 桌面 | 启动器 | 图标特写 | Dock | 实例中心 |
| --- | --- | --- | --- | --- | --- |
| 应用原色 | [查看](../../web/ui-screenshots/icon-colorways/spectrum-01-desktop.png) | [查看](../../web/ui-screenshots/icon-colorways/spectrum-02-launcher.png) | [查看](../../web/ui-screenshots/icon-colorways/spectrum-03-icons.png) | [查看](../../web/ui-screenshots/icon-colorways/spectrum-04-dock.png) | [查看](../../web/ui-screenshots/icon-colorways/spectrum-05-instances.png) |
| 浅底彩芯 | [查看](../../web/ui-screenshots/icon-colorways/porcelain-01-desktop.png) | [查看](../../web/ui-screenshots/icon-colorways/porcelain-02-launcher.png) | [查看](../../web/ui-screenshots/icon-colorways/porcelain-03-icons.png) | [查看](../../web/ui-screenshots/icon-colorways/porcelain-04-dock.png) | [查看](../../web/ui-screenshots/icon-colorways/porcelain-05-instances.png) |
| 明暗交错 | [查看](../../web/ui-screenshots/icon-colorways/contrast-01-desktop.png) | [查看](../../web/ui-screenshots/icon-colorways/contrast-02-launcher.png) | [查看](../../web/ui-screenshots/icon-colorways/contrast-03-icons.png) | [查看](../../web/ui-screenshots/icon-colorways/contrast-04-dock.png) | [查看](../../web/ui-screenshots/icon-colorways/contrast-05-instances.png) |
| 彩色薄片 | [查看](../../web/ui-screenshots/icon-colorways/pastel-01-desktop.png) | [查看](../../web/ui-screenshots/icon-colorways/pastel-02-launcher.png) | [查看](../../web/ui-screenshots/icon-colorways/pastel-03-icons.png) | [查看](../../web/ui-screenshots/icon-colorways/pastel-04-dock.png) | [查看](../../web/ui-screenshots/icon-colorways/pastel-05-instances.png) |

## 验证与边界

- `cd web && npm run build`退出0，包含Vue/TypeScript检查；仅有既存大块产物提醒。构建产物 `dist/assets` 中仅有14张默认图标PNG，四组候选和上一轮冰蓝候选未随正式应用打包。
- `cd web && node .local/capture-icon-colorways.mjs`退出0：禁用WebGL的Chromium实拍21张，包括四版各5个实际Vue界面/局部和一张全套并排图。启动器每版核对13个不同应用图标和288×288资产加载；三类场景的几何测量在四版间一致，零pageerror。原始结果为[verification.json](../../web/ui-screenshots/icon-colorways/verification.json)。
- 对照页实际切换到「明暗交错 / Dock」后成功加载1160×144截图，零pageerror；已有 `?appearance=ice` 预览的288×288图标也可正常加载。首次切换检查只有测试选择器语法错误，修正选择器后通过，产品页面没有失败。
- 截图使用本地只读演示数据；所有写请求拒绝、WebSocket关闭。这只验证视觉呈现和场景布局，不构成后端、第三方安装或远程节点验收。

下一动作：根据用户选中的方向细化颜色，必要时以单个应用为单位调整，不统一改动全部图标。

本轮Chromium和临时Vite开发服务均已停止，无运行会话需要接管。
