# 桌面控制栏与现代视觉修订

用户反馈：顶栏排版与图标尺度混乱，上一版过度模仿旧 macOS。此次按独立桌面视觉重新组织。

- 新建 DesktopBar.vue，使用三列布局：品牌/当前应用、工作区、时间/状态/账号。工具图标统一18px、相同笔画和32px点击区域；移除中英文星期混排；退出与设置收进账号菜单。顶栏样式独立，不再受旧全局按钮选择器影响。
- 移除红黄绿窗口控制；标题、窗口按钮、标签按钮统一布局与SVG符号。窗口与Dock使用轻透表面、边缘高光和柔和阴影；替换桌面背景，降低应用图标金属/高光对比。
- 顶栏增加高度后，使用共用 work-area 常量调整窗口新建、拖动和贴边区域，最大化不遮挡顶栏或Dock。保留窗口及工作内容恢复行为。
- 删除重复的任务Dock项来源，所有应用入口与权限保持原逻辑。

验证：`cd web && npm run build` 通过；`npm test -- --run` 47/47通过。`npm run test:e2e -- tests/browser/control-bar.spec.ts tests/browser/desktop.spec.ts tests/browser/accounts.spec.ts tests/browser/extensions.spec.ts --workers=2` 首轮15通过、1因header在main内缺少显式banner角色失败。补充可访问角色后 `control-bar.spec.ts` 1/1通过；16个选定场景均已验证，不声称此轮重新跑过全部浏览器用例。新增用例覆盖1440/390px顶栏控件边界、账号菜单、Escape、工作区、设置入口、导出与退出。

真实 Chromium 连接隔离本地Master/双Daemon，无API模拟截图14张；测试实例和应用条目属于fixture。截图由 `web/capture-ui.mjs` 复现，设备像素比1.5。[顶栏近景](../screenshots/2026-09-20-modern/00-topbar.png)、[实例窗口](../screenshots/2026-09-20-modern/03-instances.png)、[窗口近景](../screenshots/2026-09-20-modern/03b-instance-window.png)、[应用库](../screenshots/2026-09-20-modern/02b-app-library.png)、[账号菜单](../screenshots/2026-09-20-modern/09b-account-menu.png)、[窄屏](../screenshots/2026-09-20-modern/10-mobile-market.png)。截图后仅补充banner语义，不影响画面。

源码和web/dist更新，旧发行归档未重打包。原后端与外部环境验收边界未改变。
