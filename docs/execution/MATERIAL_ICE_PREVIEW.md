# Material 形状＋冷白冰蓝：单版外观预览

2026-09-23。用户提供早期冷白/浅蓝桌面截图，要求当前 Material Design 形状不变，配色和光影参考该图，先做一版查看效果。

## 设计与实现

- 现有形状、布局、密度、字体尺寸、Material 圆角层级保持原实现。新样式只设置颜色、透明度、阴影、边缘光和背景材质。
- 灰绿主题换为冷白、蓝灰和蓝色操作重点，图标参考图中的蓝、青绿、紫、深灰配色。保留用户已选 h04 图标底板、轮廓和透明包边。
- 顶栏、Dock 和菜单增加克制的半透明表面与细亮边；窗口使用柔和投影，内容区域保持高不透明度。顶栏内的菜单使用更密实的表面，避免桌面图标透出干扰文字。
- 使用代码绘制冷白/冰蓝曲面壁纸。没有图片生成、外部图片资产或面内斜向反光条。
- `icon-optics-renderer.ts` 增加可选颜色表，仅映射原图层的 fill/stroke 和底板色，几何与材质参数不变。输出 14 个288×288预渲染资产，共182,848字节；正常预览显示不依赖实时WebGL。
- `DockSurface.vue` 的填色、边线、阴影变为颜色角色；背景模糊层沿原连续轮廓精确裁剪。默认变量保留原纯色 Dock。
- 本版通过开发参数 `/?appearance=ice` 独立启用；当前默认 h04＋底板不替换，工作区偏好不写入这次试验。

## 效果与入口

[交互对照页](../../web/ui-screenshots/material-ice/index.html)支持在桌面、启动器、实例中心三个场景中切换当前版/冷白冰蓝版，并浏览12张新版截图。

开发服务：`/ui-screenshots/material-ice/index.html`；实际桌面：`/?appearance=ice`。

| 场景 | 截图 |
|---|---|
| 完整桌面 | [查看](../../web/ui-screenshots/material-ice/ice-01-desktop.png) |
| 启动器 | [查看](../../web/ui-screenshots/material-ice/ice-02-launcher.png) |
| 图标细节 | [查看](../../web/ui-screenshots/material-ice/ice-03-icons.png) |
| Dock 材质 | [查看](../../web/ui-screenshots/material-ice/ice-04-dock.png) |
| 实例列表 | [查看](../../web/ui-screenshots/material-ice/ice-05-instances.png) |
| 实例卡片 | [查看](../../web/ui-screenshots/material-ice/ice-06-instance-cards.png) |
| 文件管理器 | [查看](../../web/ui-screenshots/material-ice/ice-07-files.png) |
| 文件与编辑器多窗口 | [查看](../../web/ui-screenshots/material-ice/ice-08-multiwindow.png) |
| 应用市场 | [查看](../../web/ui-screenshots/material-ice/ice-09-marketplace.png) |
| 设置和账号菜单 | [查看](../../web/ui-screenshots/material-ice/ice-10-settings-menu.png) |
| 菜单细节 | [查看](../../web/ui-screenshots/material-ice/ice-11-menu-detail.png) |
| 1366×768笔记本 | [查看](../../web/ui-screenshots/material-ice/ice-12-laptop.png) |

## 验证边界

- 最终 `npm run build` 通过，包含 Vue/TypeScript 类型检查；仍有既有产物体积提醒。
- 截图脚本 `.local/capture-ice-appearance.mjs`：12张新版＋3张当前版对照，全部由实际 Vue 桌面和应用组件渲染，2倍像素，禁用WebGL仍正常显示。
- 桌面、启动器、实例列表、卡片、多窗口、菜单、笔记本7组场景的边界/圆角/字号/间距/留白及Dock路径比较全部一致，浏览器零pageerror。原始结果见[验证数据](../../web/ui-screenshots/material-ice/verification.json)。
- 普通辅助文字色 `#5d6f86` 对三层浅色底的计算对比度为5.05、4.83、4.57；主按钮白字对 `#4473b7` 为4.79。该检查仅覆盖这些角色，不声明全应用无障碍验收完成。
- 菜单不透明度与文字对比度调整后已更新截图。没有改变尺寸或通过截图后处理改变产品画面。
- 账号、节点、文件和市场目录使用本地演示数据，API写请求被拒绝，WebSocket关闭。不是后端、安装生命周期或远程节点验收。

后续根据用户对这一版的反馈决定是否采纳配色与材质。

截图浏览器和临时开发服务均已停止，无运行会话需接管。
