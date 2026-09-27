# 轻拟物图标与界面密度微调

用户反馈：应用图标过度拟物，历史记录等自绘线条图标形状不统一，账号菜单留白偏大、文字偏小。本轮保留已确认的桌面布局，只更新图标和小幅调整密度。

- 重写 AppIcon.vue 的应用图形，使用简洁的填充轮廓、半透明叠层和轻微高光；移除机箱槽位、金属齿轮、显示器支架等细节。应用库、桌面入口、Dock、窗口标题共用同一组件，应用市场仍在桌面和 Dock 中。
- 将 UIIcon.vue 的自绘路径替换为官方 `@lucide/vue`，精确锁定 1.47.0；历史、通知、盾牌、窗口控制、账号菜单、实例/文件列表和侧栏使用同一图标库，笔画统一为 1.75。品牌和应用插画仍由各自的 SVG 组件负责。
- 账号菜单宽度 240→232px，内边距 8→7px，菜单行高 37→35px，菜单字 12→13px，账号名 13→14px。应用正文、窗口标题等增加约 1px；内容边距和标题下方间距减少约 2px。此次没有再次重排桌面布局或改变后端行为。

参考用户图 3，以及 [Apple 的 Liquid Glass 适配说明](https://developer.apple.com/documentation/technologyoverviews/adopting-liquid-glass)、[Icon Composer 图标分层说明](https://developer.apple.com/documentation/Xcode/creating-your-app-icon-using-icon-composer)。图标库接口核对 [Lucide Vue 官方文档](https://lucide.dev/guide/vue/getting-started)。

## 验证

- `cd web && npm run build`：最终构建退出 0，包含 Vue/TypeScript 检查和生产构建。Vite 仍提示 Monaco 等既有大分包，不是构建失败。
- `npm run test:e2e -- tests/browser/control-bar.spec.ts tests/browser/desktop.spec.ts tests/browser/files.spec.ts --workers=2`：11/11 通过，52.9 秒。覆盖桌面控制栏/账号菜单/窄屏边界、窗口与工作区操作、文件操作与现场恢复。该组使用 API 替身；日志中的默认 WSS 8443 拒绝连接不作为真实后端证据。
- Chromium 连接真实隔离 HTTPS Master/双 Daemon `127.0.0.1:9444`，没有 API 模拟，保存 17 张截图。应用列表和实例为测试夹具资源。桌面截图视口 1600×1000、像素比 1.5，窄屏为 390×844。复现入口为 `web/capture-ui.mjs`；凭据由 `BLORA_E2E_CREDENTIALS` 指向本地私有文件，不写入报告。
- 人工检查应用库、Dock、账号菜单、实例窗口和窄屏的实际截图。截图浏览器和本轮 fixture 已关闭（fixture 会话 44040 退出 0）。本轮没有重新执行全部后端或跨平台测试。

## 截图

[应用图标近景](../screenshots/2026-09-20-glass/02c-app-library-detail.png) · [Dock 与历史图标](../screenshots/2026-09-20-glass/11-dock-detail.png) · [账号菜单近景](../screenshots/2026-09-20-glass/12-account-detail.png) · [顶栏](../screenshots/2026-09-20-glass/00-topbar.png) · [实例窗口](../screenshots/2026-09-20-glass/03-instances.png) · [应用市场](../screenshots/2026-09-20-glass/09-extensions.png) · [窄屏](../screenshots/2026-09-20-glass/10-mobile-market.png)。

源码、依赖锁文件和 web/dist 已更新；旧 rc11 发行归档未重新打包。Windows 真机、独立 systemd、远程部署和物理故障等原验证缺口保持原状态，此报告只覆盖本次视觉修订。
