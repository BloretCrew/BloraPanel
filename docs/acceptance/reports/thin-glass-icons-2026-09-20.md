# 薄片微立体图标修订 — 2026-09-20

用户否定此前过强的塑料质感，提供照片、App Store、邮件三张近景作为本轮参考。当前实现以清晰底色、薄片交叠和克制的接触阴影表现微立体；此前 `2026-09-20-liquid` 目录保留为已被用户否定的历史试稿，不能作为当前视觉基准。本轮不宣称复制了 Apple 的专有材质或达到其设计质量。

## 实现

- `web/src/app-host/IconMaterial.vue` 移除镜面光照、追光、光晕、焦散等滤镜；底板为平缓的两段渐变，白色前景使用薄片材质，仅保留两种轻微接触/悬浮阴影。
- `web/src/app-host/AppIcon.vue` 重新绘制应用前景：实例卡片、文件夹、纸笔、市场购物袋等改为薄层结构；启动器用四片局部重叠的彩色薄片；设置改为单层细齿开孔轮廓，移除镜头般的同心高光。取消指针驱动的图标照明。
- 沿用共享连续轮廓和 Dock 14px 等距框架；工具图标继续使用已锁定的 Lucide，图标 Dock、应用市场入口和现有字号/间距保持有效。
- 更新两个截图入口的输出目录，移除针对已删除追光效果的采样，保留真实渲染轮廓测量。

## 验证

| 检查 | 结果与边界 |
| --- | --- |
| `cd web && npm run build` | 退出 0；Vue 类型检查及 Vite 生产构建通过。原有 Monaco 大块提示仍存在。 |
| `cd web && npm run test:e2e -- tests/browser/control-bar.spec.ts tests/browser/desktop.spec.ts --workers=2` | 7/7 通过，44.6 秒；该套件使用 API 替身，不能替代真实后端验收。 |
| 本地真实截图 | `./dist/blora-devfixture --listen 127.0.0.1:9444` 提供隔离 Master/双 Daemon；Chromium 通过 HTTPS 登录，两个截图脚本退出 0，无 API 模拟，共 22 张 PNG。 |
| 渲染几何 | 桌面、悬停、包含历史项和 1280/760/390/320px 视口共 7 种场景；四边均为 14px，曲线距离为 13.99647–14.00049 CSS px，图标宽高误差与横向越界均为 0。 |

截图通过 `BLORA_E2E_CREDENTIALS` 环境变量引用 fixture 私有凭据文件，运行 `node web/capture-ui.mjs` 和 `node web/capture-dock.mjs`。凭据不属于公开证据。Dock 脚本使用 DPR 3，常规界面脚本使用 DPR 1.5；几何数据以 CSS px 记录。

证据：[全部截图](../screenshots/2026-09-20-thin-glass/)、[Dock 近景](../screenshots/2026-09-20-thin-glass/01-dock.png)、[应用列表近景](../screenshots/2026-09-20-thin-glass/02c-app-library-detail.png)、[桌面](../screenshots/2026-09-20-thin-glass/02-desktop.png)、[应用市场](../screenshots/2026-09-20-thin-glass/09-extensions.png)、[几何数据](../screenshots/2026-09-20-thin-glass/geometry.json)。

## 交付状态

源码和 `web/dist` 已更新；构建、测试和截图浏览器均结束，fixture 会话 94030 已停止并退出 0，无本轮测试服务待接管。旧 rc11 发行归档未重新打包。Windows 真机、独立 systemd、远程及物理故障等全范围外部验证缺口保持原状态，本轮只证明上述前端和本地界面结果。

下一步若继续调整视觉，以本轮截图、简化后的材质和既有等距轮廓为基础；若要求发行安装包，则从当前源码重新打包并进行包级验证。
