# Chromium 圆角静态绘制问题收尾

2026-10-08。此项已修复成品绘制路径，未通过增加像素容差、隐藏区域或删减场景解决。

## 根因与修复

在 Chromium 原生 DPR 1 下，八窗口叠放后将设置窗口调整为 `(400,210,661,611)`，右下圆角的光学混合区会持续发生小幅颜色变化。布局、原生位置、遮挡范围和材质坐标均不变，原八次截图稳定检查仍失败；这不是单纯截图缩放问题。

独立对照排除了 1200ms 空闲等待、Playwright 每次截图的动画/光标样式处理，以及 Chromium surface/view 截图路径。只关闭阴影片的独立保留层仍失败；只让 `window-body::before` 的壁纸材质平面随原生窗口绘制，保留窗口、阴影片、正文及全部节点，则原六场景全部通过。证据定位到 **Chromium 的独立保留材质平面与祖先圆角 clip 在调整尺寸后的组合**；其内部 alpha/damage 处理是推断，未声称已修改 Chromium 本身。

生产修复位于 `web/src/appearance/material-cache.ts` 和 `web/src/appearance/cached-material.css`：Chromium 桌面标记 `data-material-body-paint="native"`，仅对应的共享壁纸材质伪元素使用 `will-change:auto`，不再强制单独保留该光学层。原壁纸样本、圆角、颜色、窗口位置、阴影缓存和应用正文仍由原生绘制；Firefox/WebKit 策略不变。缓存销毁时恢复之前的桌面标记。

`web/tests/browser/window-position.spec.ts` 保留八次稳定检查、完整整屏 RGBA 零差异、全部原场景、节点身份和窄屏回退；增加验证真实生产材质策略的断言，探索用的截图/等待/扁平化开关已全部移除。

## 对照证据

运行环境为官方 `mcr.microsoft.com/playwright:v1.63.0-noble`、匹配的 Chromium 153.0.8010.12、原生 DPR 1/1.25，单 worker，mock 接口下的真实浏览器 DOM/SVG/光学绘制。没有修改宿主机 GPU 或浏览器优先级。

| 独立运行 | 结果 |
| --- | --- |
| 空闲 1200ms 对照 | DPR 1 失败：42 通道，max 6，范围 `[1043,814,1052,820]`；DPR 1.25 与独立位置写入通过 |
| 截图不注入动画/光标样式 | 失败：53 通道，max 7，同一圆角区域 |
| Chromium 原生 view 截图 | 失败：42 通道，max 5，同一圆角区域 |
| 仅关闭阴影片 retention | 失败：42 通道，max 5，同一圆角区域 |
| 仅材质平面随原生窗口绘制 | 六场景通过；frame、shadow retention 保持原样 |
| 正式生产修复首次复核 | **3/3，通过，38.2s**；两 DPR 共十二次整屏 RGBA0，另含独立位置写入 |
| 正式生产修复与光学回归，独立重复两次 | **24/24，通过，183.4s**；零 skip、零 flaky |

最后一轮覆盖：两 DPR 的六种窗口场景（浅色叠放、按住拖动、反向拖动、奇数尺寸、深色、实色）、独立位置写入、阴影轮廓与回退、延迟阴影发布、Dock 延迟解码与光学切换、材质六配色×深浅色及透明样本/奇数行编码。窗口位置和材质样本继续执行原有 RGBA0；Dock 保留其原先已明确的原生 PNG alpha 量化边界，未改断言。

旧不稳定图与修复图的独立运行对照还确认设置正文内区 `(432..1029,330..789)` 变化通道为 0。旧错误的边缘像素被纠正，因此不声称旧错误画面与修复画面整屏相同。

复现最后一轮：

```sh
docker run --rm --network host --ipc host --user 0 \
  -v /data/instances/blora-panel:/work -w /work/web \
  -e BLORA_E2E_PORT=18574 -e BLORA_BROWSER=chromium \
  -e BLORA_CHROMIUM=/ms-playwright/chromium-1243/chrome-linux64/chrome \
  -e PLAYWRIGHT_JSON_OUTPUT_FILE=/work/web/.local/chromium-corner-optics-final-20261008.json \
  mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test --config=playwright.config.ts \
  tests/browser/window-position.spec.ts tests/browser/desktop.spec.ts \
  tests/browser/material-atlas.spec.ts tests/browser/dock-optics.spec.ts \
  --grep 'production window positioning|independent native position|native shadow edge cache|delayed shadow contour|wallpaper-only atlas|native Dock optics|late native Dock' \
  --workers=1 --repeat-each=2 --reporter=list,json \
  --output=.local/chromium-corner-optics-final-20261008-results
```

数值及截图证据：

- [修复前空闲对照](../../../web/.local/chromium-corner-idle-control-20261008.json)
- [截图样式对照](../../../web/.local/chromium-corner-style-control-20261008.json)
- [原生 view 对照](../../../web/.local/chromium-corner-native-view-control-20261008.json)
- [阴影片对照](../../../web/.local/chromium-corner-shadow-control-20261008.json)
- [材质平面正对照](../../../web/.local/chromium-corner-material-control-20261008.json)
- [正式修复首次验证](../../../web/.local/chromium-corner-product-fix-20261008.json)
- [最后 24 项结果](../../../web/.local/chromium-corner-optics-final-20261008.json)
- [DPR 1 奇数尺寸完整截图](../../../web/.local/chromium-corner-optics-final-20261008-results/window-position-production-aa703-des-and-responsive-fallback-chromium-native-dpr-1/odd-resize-native.png)
- [DPR 1.25 奇数尺寸完整截图](../../../web/.local/chromium-corner-optics-final-20261008-results/window-position-production-7cbcf-des-and-responsive-fallback-chromium-native-dpr-1-25/odd-resize-native.png)

本项没有重新启动 E08 优化，没有把旧源的性能数字写成新源验收结果；最终全功能真实后端回归和源码/产物指纹由本次总收尾报告统一登记。没有提交、推送或发行。
