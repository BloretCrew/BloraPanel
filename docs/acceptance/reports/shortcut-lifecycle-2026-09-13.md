# 真实实例快捷入口生命周期（2026-09-13）

关联A16。`desktop.spec.ts`新增real instance shortcut rename，实际HTTPS fixture `.local/fixture-1929604743`。从实例卡添加桌面入口，关闭中心后点击入口，必须为专用资源窗口；最小化后再次点击复用同windowId，显式右键“在新窗口打开”则新增一个窗口。

关闭新视图，原窗口最小化时改入口名称并刷新，入口名称保留；实际实例名称、STOPPED与runId均不变。移除入口后实例仍可GET，任务记录无该实例的start/restart/delete。没有通过打开入口自动启动资源。

命令：web目录设置该fixture私有凭据后运行 `npm run test:e2e:real -- --grep 'real instance shortcut rename' --reporter=line`，36210退出0，1/1（3.6s）。fixture已Ctrl+C停止，无生产修改。

补充目标身份场景：真实 fixture `.local/fixture-4074818222` 中创建专用测试实例，入口添加后通过真实 `PATCH /instances/{id}` 改目标名，入口跟随目标改名；随后管理员通过真实 `DELETE /instances/{id}` 生成停止资源墓碑，再创建同名新实例。刷新后旧入口显示“不可访问”，打开仍固定旧 `resourceRef` 并显示资源失效，不会解析为同名新实例；旧 ID GET 为404、新 ID GET为200。目标实例和替代实例均按 ID清理，浏览器 `desktop.spec.ts -g 'original target'` 1/1（7.0s）通过。

补充权限场景：真实双账号 fixture 同一入口由 member 创建，管理员撤销该实例的 `instance.read` 后刷新，入口显示“不可访问”，旧窗口打开为失效提示，实例 GET 返回403；恢复授权后清理入口。`desktop.spec.ts -g 'permission is revoked'` 1/1（3.3s）通过。

补充离线场景：真实暂停本 fixture Daemon 的精确进程身份，入口状态显示“节点失联”，实例页仍显示等待确认；恢复 Daemon 后同一 runId 和原 stop task 完成。`nodes.spec.ts -g 'paused Daemon'` 1/1（50.4s）通过。

删除入口仍只删除本地 shortcut；目标删除使用管理员受控软删除，停止且无活动任务才允许，服务端保留墓碑避免同名资源替代，节点重连快照忽略墓碑。相关存储删除/幂等/活动任务冲突测试通过。以上 Linux Chromium + 本地 HTTPS 双节点证据覆盖 A16 全部行为；Windows 浏览器真机仍归 E09 平台环境缺口。

当前源码复验补充：fixture `.local/fixture-115753733` 中先显式刷新实例中心，再从管理员界面点击“删除”，确认对话框说明墓碑与停止条件，删除后列表移除；随后继续同名替换与旧入口身份核对。该 UI 场景 1/1（7.0s）通过，fixture 已停止。
