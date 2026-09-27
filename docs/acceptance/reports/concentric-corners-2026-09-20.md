# 连续圆角与 Dock 等距嵌套

用户要求 Dock 外框与内部应用图标的轮廓平行，曲线和四边间隔一致，并依据 Apple 的设计文档调整。

依据：[Apple WWDC25 设计系统说明](https://developer.apple.com/videos/play/wwdc2025/356/)讲解了同心嵌套、圆角与内边距的关系；[RoundedCornerStyle.continuous](https://developer.apple.com/documentation/swiftui/roundedcornerstyle/continuous)明确区分连续曲率与四分之一圆弧；[ConcentricRectangle](https://developer.apple.com/documentation/swiftui/concentricrectangle)说明相邻内外角之间的几何关联。公开文档提供几何原则，本次的连续曲线参数是项目自己的 Web 实现。

## 实现

- 新增共享 `surface-geometry.ts`：两段连续曲率贝塞尔曲线组成一个角，接入直边处曲率为零。应用图标统一使用这一轮廓；描边也从原轮廓内偏移取得。
- 新增 `DockSurface.vue`：以两端图标的实际位置和尺寸为基准，沿轮廓法线向外偏移 14px 生成外框，玻璃裁切和边缘描边使用同一来源。连续曲线的等距外扩不是将另一条曲线等比放大。
- Dock 图标使用不带透明留边的 SVG 视口，实际可见边缘参与布局计算；移除上下不等的内边距及独立固定外圆角。桌面图标可见尺寸保持 44px，Dock 高度为 72px。
- 窄屏按可用宽度和当前应用数量统一缩放图标，外框同步计算，所有图标保持正方形。悬停改为轻微亮度变化，保持几何关系；运行指示点、历史按钮、应用打开/恢复逻辑保留。

## 验证与证据

`cd web && npm run build` 退出 0，类型检查和生产构建通过。既有 Monaco 大分包提示仍存在。

`npm run test:e2e -- tests/browser/control-bar.spec.ts tests/browser/desktop.spec.ts --workers=2`：7/7 通过，38.5 秒。该组使用 API 替身，验证控制栏/账号操作、窗口/Dock 与恢复行为；默认 WSS 8443 连接拒绝日志不作为真实后端证据。

`web/capture-dock.mjs` 在真实隔离 HTTPS Master/双 Daemon 上登录，Chromium 像素比 3。脚本直接读取浏览器渲染后的 SVG 轮廓与坐标矩阵，测量两端共 388 个轮廓采样点到外框线段的最近距离，没有导入生产几何计算函数。完整数值见 [geometry.json](../screenshots/2026-09-20-concentric/geometry.json)。

| 场景 | 图标边长 px | Dock 尺寸 px | 四边可见间距 px | 曲线距离范围 px |
| --- | --- | --- | --- | --- |
| 1600px 桌面 | 44 | 580×72 | 均为14 | 13.9964–14.0005 |
| 悬停 | 44 | 580×72 | 均为14 | 13.9964–14.0005 |
| 出现历史按钮 | 44 | 639×72 | 均为14 | 13.9964–14.0005 |
| 1280px | 44 | 639×72 | 均为14 | 13.9964–14.0005 |
| 760px | 35 | 415×63 | 均为14 | 13.9966–14.0004 |
| 390px | 30.4375 | 373.9375×58.4375 | 均为14 | 13.9966–14.0005 |
| 320px | 22.65625 | 303.90625×50.65625 | 均为14 | 13.9965–14.0003 |

这些场景均无横向越界、无图标长宽畸变；曲线等距误差小于 0.01 CSS px。人工检查桌面、两端放大和窄屏截图。

[完整 Dock](../screenshots/2026-09-20-concentric/01-dock.png) · [右端放大](../screenshots/2026-09-20-concentric/02-right-corner.png) · [左端放大](../screenshots/2026-09-20-concentric/03-left-corner.png) · [带历史按钮](../screenshots/2026-09-20-concentric/04-dock-with-history.png) · [窄屏 Dock](../screenshots/2026-09-20-concentric/05-mobile-dock.png)。

截图浏览器已关闭，fixture 会话 70529 退出 0。源码及 web/dist 更新，旧 rc11 归档未重打包；本次局部视觉证据不改变全范围矩阵中的外部环境验证缺口。
