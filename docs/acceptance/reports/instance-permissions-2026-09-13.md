# 实例跨节点权限与创建（2026-09-13）

关联 A14。真实 HTTPS fixture `.local/fixture-3663732060` 使用 admin/member 两个账号和两个在线 Daemon。member 初始能查看两个实例但创建入口禁用、控制第二实例返回403；随后管理员只在节点2授予 `instance.create` 与 native 所需的 `host.manage`，`/nodes/creatable` 仅返回节点2，浏览器创建向导也只出现该节点。确认前撤销 `instance.create`，真实提交返回403且没有实例；重新授权后同一现场确认创建实例，返回201、目标节点为节点2且状态STOPPED，管理员按 immutable ID 删除测试实例并撤销临时授权。

命令：

```text
cd web && BLORA_E2E_CREDENTIALS=.local/fixture-3663732060/browser-credentials.json npm run test:e2e:real -- --workers=1 --trace=off --reporter=line desktop.spec.ts -g 'creation is node-scoped'
```

结果：1/1，3.5s，退出码0。该场景与现有两节点列表/过滤/运行状态测试共同证明查看节点不等于创建权限；Windows 真机及远程高延迟环境仍由 E09 保留。
