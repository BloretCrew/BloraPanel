# 扩展未授权浏览器边界（2026-09-14）

关联 A13/E06。使用真实 HTTPS Master、两个真实 Daemon 和 fixture 自有管理员/成员账号运行 `web/tests/real/extensions.spec.ts` 的 `member browser context cannot cross extension capability or node grants` 场景；扩展包从独立参考包安装，所有请求均指向真实 Master API。

定向命令：`BLORA_E2E_CREDENTIALS=<fixture 私有 browser-credentials.json> npm run test:e2e:real -- extensions.spec.ts --workers=1 --trace=off --grep 'cannot cross extension capability'`，Playwright 1/1 通过（3.5s）。随后不带 grep 运行完整 `extensions.spec.ts`，6/6 通过（22.6s），包含通知、元数据、迁移、WASI、目录升级和本场景。

场景顺序及实际断言：

1. 成员无 `app.use` 时，扩展节点资源摘要和扩展任务均返回 403。
2. 仅授予扩展 `app.use`、未授予节点 `node.read` 时，两者仍返回 403，未创建任务。
3. 临时授予节点 `node.read` 后，资源摘要返回 200；随后撤销 `app.use`，同一资源立即恢复 403。
4. 测试 finally 撤销两项授权、卸载扩展并关闭成员浏览器上下文；fixture Ctrl+C 停止后无 Master/Daemon/fixture 进程残留。

这补充了真实浏览器上下文中的扩展能力与节点授权组合证据；未将成员直接看到管理员扩展管理列表或 Windows/远程长期环境误计为通过。
