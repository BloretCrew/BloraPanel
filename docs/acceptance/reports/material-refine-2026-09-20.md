# Material 3 圆角与图标配色细化 — 2026-09-20

按用户标注补齐内容区、表头和复选框圆角，并调整备份、终端、设置、应用市场的配色。保留上一轮扁平主题、桌面及窗口布局。源码与 `web/dist` 已更新。

## 官方依据

本轮联网查阅 Google Material / AndroidX 官方资料。M3 网站的部分规范页面需要 JavaScript，因此用官方组件文档及 AndroidX Material 3 实现核对具体形状和状态，没有用第三方截图猜数值。

- [Material 3 形状令牌](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/tokens/ShapeTokens.kt)：包含 4、8、12、16、20、28、32、48dp 及全圆角层级。本项目按桌面密度选用其中的层级，映射为 CSS px；这属于 Web 适配，并非 Google 要求所有平台使用这些像素值。
- [Material 3 / Expressive 列表令牌](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/tokens/ListTokens.kt)与[列表实现](https://github.com/androidx/androidx/blob/androidx-main/compose/material3/material3/src/commonMain/kotlin/androidx/compose/material3/ListItem.kt)：Expressive 列表支持分组轮廓，普通、悬停与选中状态使用不同形状，选中颜色采用 secondary container。此处据此实现外侧 16px、连接处 4px、悬停 12px、选中 16px；行间 2px 是本项目的紧凑桌面选择。
- [Material Checkbox 文档](https://github.com/material-components/material-components-android/blob/master/docs/components/Checkbox.md)：复选框是圆角方形，具有勾选、未勾选、混合及焦点等状态。本次保留原生输入语义和键盘行为，自定义其色面、勾选标记与焦点环。
- [Material 3 主题设计](https://developer.android.com/codelabs/m3-design-theming)：参考颜色与形状的角色关系。图标具体色值为项目自己的设计，不宣称是官方应用图标配色。

## 落地内容

- `theme.css` 定义统一形状层级。窗口及正文容器采用 28px 层级；实例和文件内容区显示在有色背景上，圆角处真正露出下层色面。保持原来的坐标、宽高和内边距。
- 实例表头两端改为 12px 圆角；行改为紧凑的分组色面，去掉生硬的分隔边线。文件表头和虚拟列表选中行、其他资源表头也补齐圆角。
- 复选框由浏览器默认外观改为 16px 方形、4px 圆角，使用主色勾选和键盘焦点环；实现禁用、混合状态及强制色模式原生外观回退。保留既有点击及表单逻辑。
- 备份图标采用薄荷绿与深青绿；终端采用浅灰蓝底、蓝色控制台和浅绿光标；设置采用灰蓝色阶；市场采用珊瑚色阶和暖白。保留多色平面图形，统一明度关系；其他图标不作整套重绘。
- Dock 的连续轮廓、14px 等距外框和工具图标库保持原实现。

## 验证

| 检查 | 实际结果 |
| --- | --- |
| `cd web && npm run build` | 退出 0，含 Vue 类型检查与生产构建。仍有原有 Monaco 大包提示。 |
| 相关浏览器回归 | `npm run test:e2e -- tests/browser/instance-batch.spec.ts tests/browser/instance-settings.spec.ts tests/browser/files.spec.ts tests/browser/desktop.spec.ts tests/browser/control-bar.spec.ts tests/browser/monitor.spec.ts tests/browser/accounts.spec.ts --workers=2`：17/17 通过，53.6 秒，会话 38223 退出 0。使用 API 替身；未将其计为后端全范围验证。 |
| 真实页面与键盘 | 本地隔离 HTTPS Master/双 Daemon，Chromium 截图 24 张。空格勾选/取消、选中计数、批量按钮启用及焦点通过；实际复选框 16×16px、4px 圆角。 |
| 布局保持 | 构建前直接采样旧页面、构建后采样新页面；13 组桌面/手机主要区域，0.5 CSS px 容差内无差异。列表内部增加的 2px 行间隔及 4px 表头间隔属于本轮微调，不在此宏观布局结论内。 |
| Dock 轮廓 | 7 种场景四边间距均 14px；曲线距离 13.99647–14.00050 CSS px，图标宽高畸变及横向越界均为 0。 |

复选框混合、禁用与强制色模式有样式实现，本轮没有逐态实机截图验收；实际键盘验证覆盖正常未选/选中及焦点状态。浏览器回归初次启动的自动审批超时，未执行测试；再次启动成功，上表是成功执行的实际结果。

## 截图与复现

- [实例中心及圆角](../screenshots/2026-09-20-material-refine/03b-instance-window.png)
- [勾选与键盘焦点](../screenshots/2026-09-20-material-refine/03c-instance-selection.png)
- [全套图标](../screenshots/2026-09-20-material-refine/02c-app-library-detail.png)
- [Dock 近景](../screenshots/2026-09-20-material-refine/01-dock.png)
- [文件内容区](../screenshots/2026-09-20-material-refine/05b-file-window.png)
- [桌面](../screenshots/2026-09-20-material-refine/02-desktop.png)、[应用市场](../screenshots/2026-09-20-material-refine/09-extensions.png)
- [布局比较](../screenshots/2026-09-20-material-refine/layout-comparison.json)、[键盘结果](../screenshots/2026-09-20-material-refine/checkbox-keyboard.json)、[Dock 测量](../screenshots/2026-09-20-material-refine/geometry.json)

启动 `./dist/blora-devfixture --listen 127.0.0.1:9444`，将 `BLORA_E2E_CREDENTIALS` 指向启动时输出的私有文件，运行 `node web/capture-ui.mjs` 和 `node web/capture-dock.mjs` 即可复现当前截图。`capture-layout.mjs` 通过 `BLORA_LAYOUT_STAGE=before/after` 比较改动两侧的稳定布局。凭据不进入报告。

截图浏览器、构建和测试均已结束；本轮 fixture 会话 85791 已停止并退出 0。旧 rc11 安装归档没有重打包。Windows 真机、独立 systemd、远程与物理故障等既有外部验证缺口保持原状态，本轮不提升其验收结论。
