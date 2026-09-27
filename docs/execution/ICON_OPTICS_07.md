# 图标反射与折射候选 07

日期：2026-09-22。

用户指出上轮薄边缘不可见，希望比较更接近 Liquid Glass 的反射玻璃材质。本轮仍使用原有 14 应用 / 63 部件的图形，改为自定义 WebGL 图层合成；全部由本地代码渲染并由浏览器截图，没有图片生成。

## 实现边界

- 原始矢量仅用于轮廓与部件遮罩；在 288 × 288 分辨率生成覆盖率、向内距离和法线纹理。
- 光学预览逐层采样下层内容，按边缘法线偏移采样，形成边缘折射；法线与可移动面积光源计算局部反射。
- 本轮六版主体不采用 Gaussian blur；第 06 版增加随光照位置变化的环境反射带。对照行仍使用上一轮 B 的模糊透叠。
- 仅共享一个 WebGL context，图标本身显示在 2D canvas 中。指针移动或参数变化时按需渲染，无空闲动画循环；减少动态效果偏好会关闭指针追光。
- 底板及图形构图保持原有样式。材质是网页中的定制近似实现，不是 Apple 原生系统材质，也不是完整物理光线追踪。
- 仅开发参数启用；未切换默认图标。尚未做跨浏览器 / 低端设备性能验收，候选需要 WebGL。

## 官方参考

[Apple：Meet Liquid Glass](https://developer.apple.com/videos/play/wwdc2025/219/) 说明随几何和光源变化的高光、lensing 和下层内容折射。

[Apple：Adopting Liquid Glass](https://developer.apple.com/documentation/technologyoverviews/adopting-liquid-glass) 说明分层应用图标及系统施加的反射、折射等效果。本项目仅借鉴这些原则。

## 六版与截图

| 版本 | 方向 | 放大 / 小尺寸 | 桌面 |
| --- | --- | --- | --- |
| g01 | 清透薄片 | [对照](../../web/ui-screenshots/icon-optics-07/g01-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g01-desktop.png) |
| g02 | 弧面折射 | [对照](../../web/ui-screenshots/icon-optics-07/g02-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g02-desktop.png) |
| g03 | 柔面反射 | [对照](../../web/ui-screenshots/icon-optics-07/g03-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g03-desktop.png) |
| g04 | 高透玻璃 | [对照](../../web/ui-screenshots/icon-optics-07/g04-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g04-desktop.png) |
| g05 | 烟色玻璃 | [对照](../../web/ui-screenshots/icon-optics-07/g05-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g05-desktop.png) |
| g06 | 环境映光 | [对照](../../web/ui-screenshots/icon-optics-07/g06-detail.png) | [桌面](../../web/ui-screenshots/icon-optics-07/g06-desktop.png) |

[01–03 主体](../../web/ui-screenshots/icon-optics-07/group-1-foreground.png) · [04–06 主体](../../web/ui-screenshots/icon-optics-07/group-2-foreground.png) · [02 光源变化](../../web/ui-screenshots/icon-optics-07/g02-angles.png) · [06 光源变化](../../web/ui-screenshots/icon-optics-07/g06-angles.png)。

## 验证

- `cd web && npm run build`：类型检查与生产构建通过；仍有已知的大 chunk 提示。
- `node .local/capture-icon-optics.mjs`：六版共 36 张 2× PNG，包含深浅底、去底板、完整 14 应用、32 / 48 / 64 px、三光源和实际桌面 / 启动器 / Dock。
- Chromium 无 pageerror；每个 canvas 检查渲染成功状态；六版分别比较左右光源后的 canvas 图像，确认结果实际变化。
- 修复裸主体在透明区域折射采样的暗点，并修正光源轴向，确认反射随左右光源移动。
- 一次截图因本机 Vite 收到 SIGTERM 退出而连接失败；重启后重拍成功，失败未计入成功证据。
- 截图为视觉评审数据：本地 session / 空列表，不作为后端验收。

开发入口：`/icon-optics.html?group=1` / `group=2`；单版 `?variant=g02` 等。实际桌面：`/?icon-study=g01` 至 `g06`。下一步依据用户对反射 / 折射强度与材质方向的反馈收敛。

