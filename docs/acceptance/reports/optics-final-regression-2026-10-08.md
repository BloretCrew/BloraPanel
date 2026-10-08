# 最终 Chromium 光学回归复核

2026-10-08。以下区分成品守卫与未采用的显式光学实验；保留原失败证据，不放宽像素边界。

后续终端工具栏的正式修复及独立 4/4 + 7/7 复验见[终端绘制收尾](terminal-toolbar-paint-closeout-2026-10-08.md)。下列完整运行失败属于修复前记录，未改写为通过；总交接状态见[整合报告](non-release-closeout-2026-10-08.md)。

## 完整运行的失败定位

完整 mock 回归使用官方 Playwright 1.63 容器和匹配的 full Chromium 153.0.8010.12（`chromium-1243/chrome-linux64/chrome`），单 worker。结果为 175 PASS、6 FAIL，零 skip、零 flaky。原窗口位置/圆角的两个原生 DPR 守卫均通过；此轮没有改变之前的生产圆角修复。

| 守卫 | 原失败差异及位置 | 定位范围 |
| --- | --- | --- |
| 成品 desktop partial-shadow，DPR 1 | held 场景整屏 1 通道、max 1；`(1278,407)` 红通道 238/239 | 远离下层窗口被裁掉的阴影片；原差异仍保留，尚未证明根因 |
| 成品 Dock optics，DPR 1.25 | restored 场景 60478 通道、max 4；物理裁图范围 `[272,0,1363,22]` | 实际屏幕 CSS y=800..817.6，位于设置窗口下缘而非 Dock 本体；此 DPR 原生回退没有 Dock PNG 缓存，不能把差异归因于 PNG 量化 |
| 成品 terminal partial-paint，DPR 1 | bottom-cover 场景 9 通道、max 6；三个像素 `(769,235)`、`(767,245)`、`(766,246)` | 上层终端工具栏的新建会话按钮；主代理独立复验仍失败，单独处理 |
| bar-optics 候选，DPR 1 | 第一对截图内区 glyph 边缘 max 1 | 测试显式安装的未采用 bar 合成候选，并非成品 bar 绘制 |
| shadow-composite 候选，DPR 1 | 第一对 overlap 内区 max 1，整屏 max 91 | 测试显式安装的未采用分组阴影片候选，整屏差异不能只解释为 glyph 量化；没有修改候选来制造通过 |
| offscreen/network | 网络执行错误 | 主代理独立复验已通过，未由本项修改 |

原 24 项光学回归也是同 full Chromium 153，但没有运行 bar-optics 或 shadow-composite 候选。不能以原 24 PASS 证明这两个候选通过。

## 成品模块独立复验

在所有其他浏览器停止后，新建本任务专用容器，保持同一 Chromium、原生 DPR、原断言和全部场景，串行运行真正的 desktop partial-shadow 与 Dock 守卫。**5/5 PASS，31.227s，零 skip、零 flaky**：

- desktop partial-shadow：DPR 1 / 1.25 的初始、按住拖动、释放、切换焦点和原生回退，全屏 RGBA 零差异。
- Dock optics：DPR 1 / 1.25 的浅色、实际应用、深色暖砂、实色、恢复通透、连续主题切换，全部 `changed=0,max=0,total=0`。DPR 1 原有允许光学量化边界未更改，实际结果仍为零。
- Dock 延迟解码、过期发布及解码失败回退：通过。

本次没有修改成品源码或测试断言，也没有增加等待、重试、忽略区域或删除场景。独立 PASS 只能说明此轮相同成品路径通过；它**不足以证明完整套件中的偶发像素变化已根治**。原 FAIL 与这一复验都登记为证据，由总收尾报告明确剩余局限，不能将第一次完整运行改写为通过。

复现命令（JSON 和 PNG 输出均保留在不进入产品的 `.local`）：

```sh
docker run --name blora-optics-closeout --rm \
  -v /data/instances/blora-panel:/workspace -w /workspace/web \
  -e BLORA_E2E_PORT=18687 -e BLORA_E2E_FRESH_SERVER=1 \
  -e BLORA_CHROMIUM=/ms-playwright/chromium-1243/chrome-linux64/chrome \
  -e PLAYWRIGHT_JSON_OUTPUT_NAME=/workspace/.local/optics-closeout-initial-20261008.json \
  mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test tests/browser/desktop.spec.ts tests/browser/dock-optics.spec.ts \
  --grep 'partially covered native shadow patches|native contour and shadow|late native Dock decoding' \
  --workers=1 --reporter=list,json --output=.local/optics-closeout-initial-output
```

证据：[完整 mock 原结果](../../../.local/nonrelease-20261008-final-mock.json)、[5 项独立复验](../../../.local/optics-closeout-initial-20261008.json)。截图附件保留于对应 JSON，未将真实运行快照或私有运行日志写入报告。

专用容器运行完成后自动退出删除，18687 测试服务结束。没有重新运行 E08 压力优化，没有提交、推送或发行。
