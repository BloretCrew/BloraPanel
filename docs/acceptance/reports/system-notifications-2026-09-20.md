# 浏览器系统通知 API 路径验收（2026-09-20）

## 验收范围

在真实 TLS Master/双 Daemon fixture 和 headed Chromium（Xvfb）中，授予该测试来源通知权限，使用任务中心开启系统通知，再由真实节点任务完成触发通知服务。测试确认实际 Chromium `Notification` 对象收到任务标题、正文和稳定 tag，并派发该对象的 click 事件后打开绑定该 taskId 的专用任务窗口。站内任务通知保持独立。

## 结果

命令：`BLORA_E2E_CREDENTIALS=<fixture 私有 browser-credentials.json 路径> xvfb-run -a npm run test:e2e:real -- tests/real/notifications.spec.ts --grep 'task completion reaches the browser notification API' --headed --workers=1 --trace=off --reporter=line`

最终复验 1/1 通过，3.1 秒。首次复验在功能已打开两个任务窗口后，通用标题定位因中心窗口和资源窗口各有一个“任务中心”标题而产生严格模式冲突；测试改为限定 `data-window-mode="resource"` 的目标窗口后通过。这是测试定位修正，不是产品行为失败。

新增真实浏览器用例位于 `web/tests/real/notifications.spec.ts`。仅增加验收代码，没有修改生产通知实现或 UI 样式。

## 验收边界

此结果证明真实浏览器的通知权限、原生 `Notification` 构造调用和应用点击处理路由；测试通过脚本派发 click 事件，不代表用户从桌面通知中心实际点击。Xvfb 是虚拟显示环境，没有可观察的桌面通知中心，因此系统桌面合成器中通知的实际绘制、声音/横幅策略、通知中心点击和浏览器退出后的生命周期仍未验证。F09 继续进行中。
