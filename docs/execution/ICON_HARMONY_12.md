# 01A 图标与壁纸的配色协调预览

2026-09-23。用户认为01A实色底与浅色底的视觉重量不一致，且冰蓝壁纸不协调，要求同时提供多个图标和壁纸方向。本轮沿用01A图形、h04透明包边材质、底板及当前Material布局；只调整色彩关系和壁纸。

## 方案

| 图标版本 | 底板与主体处理 | 示例搭配 |
| --- | --- | --- |
| A 浅瓷底 `unify-porcelain` | 近白底板，蓝、绿、红等识别色移到主体 | W1 暖白细雾 |
| B 适中彩底 `unify-tonal` | 各底板保持独立色相，压小明度差，主体更深 | W2 石色纸层 |
| C 全彩底 `unify-color` | 全套采用彩底，内部图形转为浅色层次 | W3 雾紫暮色 |
| D 深底彩芯 `unify-smoke` | 深色底板承托各应用的彩色主体 | W5 岩白织面 |
| E 原01A `spectrum-silver` | 图标资产完全保留，仅更换壁纸 | W4 青灰流线 |

壁纸均为代码绘制SVG：暖白细雾是暖灰色域和细颗粒；石色纸层是低对比灰绿层面；雾紫暮色是浅紫与淡桃色域；青灰流线是连续的青灰色带；岩白织面是浅灰曲线和细纹。原冰蓝壁纸保留对照。没有图片生成或新增图标斜向反光。

参考[Microsoft应用图标设计指南](https://learn.microsoft.com/en-us/windows/apps/design/iconography/app-icon-design)中的清晰应用意象、小尺寸辨识及克制层次原则。最终配色以用户对01A的反馈为准；本轮不声称复刻某个系统图标规范。

## 实现与使用

- [`icon-harmony.ts`](../../web/src/app-host/icon-harmony.ts)定义四套完整的底板、主体和内部颜色；终端保留01A窗口几何，只替换fill/stroke。使用现有代码材质渲染器导出56张288×288 PNG，共749,685字节。
- [`preview.ts`](../../web/src/appearance/preview.ts)仅在开发冰蓝预览下读取壁纸白名单，[`ice.css`](../../web/src/appearance/ice.css)替换对应背景图。桌面、Dock、启动器和应用窗口的布局未调整。
- 示例开发入口：`/?appearance=ice&icon-colorway=unify-porcelain&wallpaper=warm`。图标可选上表四个新ID及 `spectrum-silver`；壁纸可选 `warm|strata|haze|flow|linen`，省略则为原冰蓝。
- 默认产品主题和默认h04资产保留。新增PNG、配色表与SVG壁纸不进入正式运行资产。

## 截图

[完整对照页](../../web/ui-screenshots/harmony-12/index.html)包含五组整体搭配，并可独立切换5套图标×6张壁纸，共30种启动器组合。共57张PNG：30张启动器、10张桌面、5张图标特写、5张Dock、5张实例窗口和2张对照板。

- [图标方向对照](../../web/ui-screenshots/harmony-12/00-icon-directions.png)
- [壁纸方向对照](../../web/ui-screenshots/harmony-12/00-wallpaper-directions.png)
- [A 实景](../../web/ui-screenshots/harmony-12/unify-porcelain--warm--launcher.png)
- [B 实景](../../web/ui-screenshots/harmony-12/unify-tonal--strata--launcher.png)
- [C 实景](../../web/ui-screenshots/harmony-12/unify-color--haze--launcher.png)
- [D 实景](../../web/ui-screenshots/harmony-12/unify-smoke--linen--launcher.png)
- [E 实景](../../web/ui-screenshots/harmony-12/spectrum-silver--flow--launcher.png)

## 验证与边界

- `cd web && npm run build`退出0，包含Vue/TypeScript检查；仅有既有大块产物提示。检查正式产物只有14张默认PNG，JS中无 `unify-porcelain` 配色ID。
- `node .local/export-selected-icons.mjs --colorway=<新版本>`四组导出通过。
- `node .local/capture-harmony.mjs`退出0，Chromium禁用WebGL，1600×1000视口、2倍像素，完成57张截图。30组启动器和5组实例窗口的测量几何分别一致；各组14种图标资产解码成功。图库30种组合切换通过，390px图库无页面横向溢出。零pageerror、零API写请求。见[完整验证记录](../../web/ui-screenshots/harmony-12/verification.json)。
- 首次截图时Vite内联SVG的URL含括号，旧解析截断URL导致解码失败；修复完整URL提取后重跑通过。续接时原Vite进程已退出，首次局部重拍报连接拒绝；重启本机Vite后恢复。
- 人工审图发现青灰壁纸有内部尖角，已将深色色带延伸出画面并修顺曲线。`node .local/capture-harmony.mjs --wallpaper=flow`退出0，更新7张相关实景及2张对照板；重新核实30种图库组合。见[青灰壁纸复核记录](../../web/ui-screenshots/harmony-12/verification-flow.json)。最终PNG数量仍为57。
- 使用实际Vue组件和隔离本地演示数据，写API被拒绝，WebSocket关闭；视觉截图不构成后端/远端资源验收证据。未操作真实节点或公开部署。
- 本轮浏览器和临时Vite已停止，无运行会话需接管。

下一动作：依据用户对图标方向和壁纸方向的独立反馈收敛搭配，不将任何新候选自动设为默认，也不恢复用户已否定的04A方向。
