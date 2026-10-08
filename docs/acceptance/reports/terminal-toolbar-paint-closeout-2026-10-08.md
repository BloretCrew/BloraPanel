# 终端工具栏原生绘制差异收尾

2026-10-08。此项修复成品 Chromium 的绘制归属，没有更换终端渲染器、裁掉按钮、忽略正文或放宽整屏像素断言。

## 原失败与有界对照

完整 mock 回归及其独立复验在原生 DPR 1 的 `terminal-partial-paint` 最后一种 `bottom-cover` 场景均失败：9 通道、max 6，仅新建会话按钮的三个圆角像素 `(769,235)`、`(767,245)`、`(766,246)` 改变。工具栏在 CSS y≈235，远离本场景 y≥470 的遮挡窗口；前五种场景均整屏 RGBA0，DPR 1.25 全部通过。

增加与窗口位置守卫相同的最多八次整屏稳定捕获前置后，同一成对比较仍失败。它排除了连续静态截图尚未稳定的解释，**没有单独修复按钮 AA 差异**，也未证明所谓“稍后原生重绘”就是原因。

唯一正对照只给 `.terminal-app>.editor-toolbar` 增加 `will-change:transform`，其余 DOM、几何、终端区域、窗口/材质/阴影路径不变。两个 DPR 的原六种场景整屏零差异，输出保护、只读观察、完整终端网格和 reload/resume 断言继续通过。对照结果为 2/2 PASS，30.434s。

证据定位到 Chromium 对同一窗口中工具栏和兄弟终端裁切区域的共同绘制归属：终端 viewport clip 的切换会改变工具栏按钮的原生 AA，而让不变工具栏保留自己的层后差异消失。Chromium 内部 damage/raster 具体机制属于推断，未声称已证明或修改浏览器内部实现。

## 最小生产修复

`web/src/applications.css` 仅为现有 `data-material-body-paint="native"` Chromium 桌面下的终端工具栏增加 `will-change:transform`。规则不作用于 Firefox/WebKit，也不改变窗口几何/圆角、主题配色、终端网格、行节点、输出解析与恢复协议。临时控制环境开关已删除；正式守卫直接验证生产规则。

`web/tests/browser/terminal-partial-paint.spec.ts` 保留原完整六种场景、整屏 RGBA0、网格/输出/恢复/只读断言和八次稳定捕获前置；注释明确稳定前置本身并未修复差异。

正式修复在两个 DPR 各重复两次：**4/4 PASS，55.459s，零 skip、零 flaky**。共 24 对完整屏幕截图均为 `changed=0,max=0`。

随后单 worker 串行运行两 DPR 的原窗口位置/圆角、desktop partial-shadow、Dock optics，以及 Dock 延迟解码回退，共 **7/7 PASS，60.0s，零 skip、零 flaky**。窗口位置各六场景与 partial-shadow 各五场景整屏 RGBA0；Dock 两 DPR 的六种主题/应用场景也全部实测 RGBA0，未改变其原有光学量化边界。这与前面的终端正式复验合计 11 个测试执行；没有将重复执行或场景数混称为 11 个独立功能。

跨独立运行的旧截图与修复截图不保证整屏相同，原生文本/边缘栅格化也可能不同，因此此报告不声称“旧错误画面与修复画面整屏零差异”。严格零差异证明的是修复后相同场景、相同现场中部分终端 viewport 与完整原生 viewport 的配对比较。

## 证据与清理

- [完整 mock 原失败](../../../.local/nonrelease-20261008-final-mock.json)
- [稳定截图前置仍失败](../../../.local/nonrelease-20261008-terminal-stable.json)
- [仅工具栏独立层正对照](../../../.local/terminal-toolbar-layer-control-20261008.json)
- [正式生产修复双 DPR 各重复两次](../../../.local/terminal-toolbar-product-fix-20261008.json)
- [相关窗口位置、partial-shadow 与 Dock 回归](../../../.local/terminal-toolbar-optics-regression-20261008.json)

所有运行使用官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 与匹配的 full Chromium 153.0.8010.12，显式路径 `/ms-playwright/chromium-1243/chrome-linux64/chrome`，单 worker；未使用 headless-shell 替换版本。测试使用真实浏览器 DOM/原生软件终端，HTTP/WSS 为测试传输，不将模拟场景称为真机资源测试。

没有重开 E08 优化、没有修改产品默认开关、没有提交/推送/发行。测试产物保留于 `.local`，不进入产品默认流程；专用 18688 服务随测试容器退出。
