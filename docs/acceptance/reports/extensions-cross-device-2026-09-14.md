# 扩展跨设备现场与资源身份复验

2026-09-14，补充 Linux Chromium 浏览器场景，使用两个不同的本地 `deviceId`：第一台打开扩展资源窗口并同步“布局/引用”副本；第二台清理本地现场后打开该云端副本，同时把扩展清单从状态版本 1 提升到 2。

场景由 `web/tests/browser/extensions.spec.ts` 中的 `sandbox workspace copy on another device keeps resource identity and migrates view state` 覆盖。结果 1/1 通过（4.7s，退出码 0）：扩展现场在 opaque sandbox iframe 中完成 1→2 迁移，原有实例 `instance-1` 保留，旧节点提示 `node-old` 由资源读取返回的规范身份校准为 `node-new`；恢复后仍只有两个扩展窗口，没有创建重复窗口或重放任务。

同时修正恢复服务：默认应用和具备 `resource.read` 的扩展资源校准可以返回规范 `resourceRef`；实例被删除或权限被撤销时保留本地视图与草稿，不把恢复保护误报为存储失败。前端 `npm run check`、`npm test -- --run`（46/46）和 `npm run build` 均通过。

该证据覆盖跨设备副本、扩展状态迁移和资源身份校准的 Linux 浏览器组合；远程高延迟、长期运行、Windows 真机及完整未授权浏览器矩阵仍按 E06/E07/A13 保留未验证。
