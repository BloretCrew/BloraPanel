# 扩展桌面 SDK 能力

2026-09-12，F01/F14、A13/A17、E06 子项。

SDK 新增 `moveView`、`closeView`、`createShortcut`，分别要求 `window.move`、`window.close`、`shortcut.create` 声明。Master 安装白名单与 AppHost 清单验证同步。宿主固定调用者的 appId/viewTabId，扩展只能移动/关闭自身标签，目标窗口必须存在且属于同应用并兼容标签。快捷入口绑定当前资源，总览拒绝创建；关闭视图保留关闭历史和后台任务。桌面变更直接使用统一恢复事务；移动/关闭回执在 Vue 卸载 iframe 前发送。

独立示例包升级为 0.2.0，新增移窗、添加桌面入口和关闭按钮，执行桌面动作前等待当前笔记保护请求完成。安装升级权限仍由已有能力提升检查处理，不自动批准新增能力。

验证结果（退出 0，未特别说明均本地 Linux）：

- `cd web && npm test`：32/32；包含真实 desktop store 的自身身份/能力拒绝、无效目标拒绝、移窗状态保留和关闭不删除入口。
- 新增浏览器场景首轮快捷入口未出现，4 个原场景通过。未查明首轮原因；保留为未复现的时序风险。定向复跑 1/1（4.5s），再 `npm run test:e2e -- tests/browser/extensions.spec.ts --grep 'sandbox controls' --repeat-each=5 --workers=1 --trace=off --reporter=line` 5/5（21.3s），均为后端路由替身。
- SDK build、独立包 `npm run package`、前端生产 build、`make check windows`、Linux `make build` 通过。
- 真实双节点 fixture `.local/fixture-2811626890`，HTTPS 127.0.0.1:9444：`BLORA_E2E_CREDENTIALS=<私有凭据路径> npm run test:e2e:real -- tests/real/extensions.spec.ts --grep 'independent package' --workers=1 --trace=off --reporter=line` 1/1（9.6s）。完整独立包安装、WASI 计算、SDK 结果回读、移窗笔记恢复、刷新任务及快捷入口恢复、关闭自身视图，任务提交次数始终为 1。fixture 已 Ctrl+C 停止。

后端独立包测试首轮因固定升级目标 0.2.0 与新包相同而返回 409；改为从当前主版本生成下一主版本，保留真实升级/幂等语义，定向复验通过（4.299s）。最终扩展浏览器组合 5/5（17.2s）通过。全仓 race 正在另行运行，不能提前计作通过。后端用户数据迁移、通知与其余资源能力、跨设备/长时恢复和 Windows 真机仍未完整验证。

全仓 race 最终结果：`GOCACHE=/tmp/blora-go-grant-idem make test` 退出 2，Master 177.795s 仅失败于修复前编译的同版本升级用例，其他包通过。修复后 `go test -race ./internal/master -run TestIndependentReferencePackageResourceBridge -count=1`（相同缓存与授权环境）通过，33.921s。此处记录失败及针对性复验，不把原全仓命令改写为通过。所有测试及 fixture 已结束。
