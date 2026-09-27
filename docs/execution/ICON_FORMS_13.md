# 应用图标造型发散：四套完整图形

2026-09-23。用户认为上轮各方案的图标变化仍不明显，要求图标本身向多个方向发散。本轮从配色研究转为轮廓、构图及应用意象研究，每套重新绘制14个图标；沿用接受的h04材质和浅色底板，并固定暖白壁纸，便于比较主体变化。

## 四个方向

| 方向 | 实例中心 | 终端 | 监控 | 应用市场 |
| --- | --- | --- | --- | --- |
| F1 透明叠片 | 错位层叠的服务面板 | 前后命令窗口 | 两张叠放图表 | 错位购物袋 |
| F2 饱满大形 | 整块圆润机箱 | 饱满面板与命令/输出块 | 圆形仪表 | 大购物袋与四个应用格子 |
| F3 精致器物 | 窄立式机柜 | 带支架的终端屏幕 | 小型监控屏幕 | 打开的应用包装盒 |
| F4 彩色折面 | 折面层板 | 折角命令页 | 两片面积图 | 折叠购物袋 |

其余文件、编辑器、备份、任务、节点、用户、设置、容器、系统及应用入口也按对应方向重画；不是只更换上述四个重点应用。保持前后层面的轻微透色，不添加斜向反光带、夸张面内光照或图片生成资产。

参考[Android自适应图标](https://developer.android.com/develop/ui/compose/system/icon_design_adaptive)的前景/背景分层和主体安全区思路，以及[Windows应用图标指南](https://learn.microsoft.com/en-us/windows/apps/design/iconography/app-icon-design)的清楚意象、轮廓辨识与小尺寸细节控制。作为探索，F4额外尝试折面与轻微角度；本轮不宣称原样遵循任何单一平台图标规范。

## 源码与预览

- [`icon-form-directions.ts`](../../web/src/app-host/icon-form-directions.ts)集中定义四套图形。四套共用完全一致的底板值、各应用识别色与h04材质；通过 `artwork` 替换全部14种主体路径。
- [`icon-optics-renderer.ts`](../../web/src/app-host/icon-optics-renderer.ts)为图层绘制补齐绕指定中心旋转与evenodd填充。没有修改光学shader及h04材质参数。正常桌面使用预渲染PNG，无须WebGL。
- 开发入口：`/?appearance=ice&wallpaper=warm&icon-colorway=form-layered|form-bold|form-studio|form-folded`；正式默认图标保持原样。
- [完整对照页](../../web/ui-screenshots/forms-13/index.html)包含两个7应用对照板、14应用单套近看、32/48/64px小尺寸比较，以及桌面、启动器、Dock和实例窗口切换。

| 方向 | 全套图标 | 实际启动器 | 实际Dock |
| --- | --- | --- | --- |
| F1 | [查看](../../web/ui-screenshots/forms-13/form-layered--collection.png) | [查看](../../web/ui-screenshots/forms-13/form-layered--launcher.png) | [查看](../../web/ui-screenshots/forms-13/form-layered--dock.png) |
| F2 | [查看](../../web/ui-screenshots/forms-13/form-bold--collection.png) | [查看](../../web/ui-screenshots/forms-13/form-bold--launcher.png) | [查看](../../web/ui-screenshots/forms-13/form-bold--dock.png) |
| F3 | [查看](../../web/ui-screenshots/forms-13/form-studio--collection.png) | [查看](../../web/ui-screenshots/forms-13/form-studio--launcher.png) | [查看](../../web/ui-screenshots/forms-13/form-studio--dock.png) |
| F4 | [查看](../../web/ui-screenshots/forms-13/form-folded--collection.png) | [查看](../../web/ui-screenshots/forms-13/form-folded--launcher.png) | [查看](../../web/ui-screenshots/forms-13/form-folded--dock.png) |

## 验证记录

- 最终 `cd web && npm run build` 退出0，Vue/TypeScript检查通过，仅保留既有大块产物提示。正式产物仍只有14张默认PNG，JS内无 `form-layered` 预览ID。
- `node .local/export-selected-icons.mjs --colorway=<版本>`完成四套导出：56张288×288图标，共814,685字节。
- `node .local/capture-forms.mjs`退出0：禁用WebGL的Chromium产生33张截图，包含四套新图标及上轮A对照。五种场景各五组、五张全套近看、两张主体对照和一张尺寸对照。
- 图形验证剔除fill/stroke颜色后比较路径、尺寸、旋转和构图，确认56个主体全部区别于01A几何，四套之间也无同应用的重复几何。四套底板与上轮A完全一致。见[完整验证记录](../../web/ui-screenshots/forms-13/verification.json)。
- 每套14类资产解码成功，五组启动器/五组实例窗口的布局分别一致；图库25个场景切换通过，390px宽度无页面横向溢出，零pageerror、零写API。
- 人工审图修正F2购物袋容易形成卡通脸的排列，以及F4设置图标类似风扇的轮廓；改为四个应用格子和结构清楚的双面齿轮后重新导出并完整重拍。随后发现齿轮两层同轮廓产生重复光学边缘，将分色层改成同一物体的平面色区；最终 `node .local/capture-forms.mjs --variant=form-folded` 退出0，更新5张实景、5张全套近看和3张对照图，25个场景切换复核通过。见[最终局部复核](../../web/ui-screenshots/forms-13/verification-form-folded.json)。最终图片总数仍为33。
- 使用真实Vue桌面组件、隔离本地演示数据，写API拒绝且WebSocket关闭。未执行真实节点操作、改变产品默认外观或公开部署；不计后端验收。
- 浏览器、导出、构建和临时Vite会话均已停止，无运行会话需接管。

下一动作：依据用户对F1～F4主体造型的反馈细化，可与上一轮选定的底板配色或壁纸组合；保留现有全部候选，不自动将新方案设为默认。
