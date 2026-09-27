# 扩展实例元数据写入（2026-09-12）

后续最终浏览器组合3/3（9.8s）通过，元数据分支新增同步恢复日志故障注入：写入现场失败时不发送PATCH，恢复存储后显式重试同一请求成功。宿主restoreState现在检查恢复服务的protected状态，避免内部记录错误却对扩展返回成功。独立包初始化控件也在恢复完成后才启用，防止最近输入被初始读取覆盖。详见 [通知组合报告](notifications-2026-09-12.md)。

关联F14/E06，整项继续进行中。

新增SDK updateResource和resource.write transport，PATCH扩展资源API只接受实例名称/分组/标签。服务端复核启用能力、app.use、instance.configure、资源节点身份；复用现有配置事务的configRevision条件更新、持久幂等与审计。读资源摘要增加configRevision、group、tags；写响应不含执行配置。未知config字段拒绝，禁用后拒绝新写入和重放。

验证：新增真实TLS测试覆盖缺少app.use、缺少instance.configure、授权写入、稳定请求重放、旧修订冲突、执行配置拒绝及禁用拒绝。最初两次测试分别因Package不序列化payload、误用disable路由失败，已修正请求后复验；最终结果见PROGRESS。前端transport10/10通过，包含当前实例身份与请求键绑定、未声明能力拒绝。SDK check/build、前端check/build、make check/build/windows通过。

独立样例已声明 resource.write，增加实例窗口与名称编辑。最近草稿、目标资源、修订号和稳定请求标识保存在视图现场；提交前等待现场写入完成，结果不明时仅显式重试同一请求，刷新不自动重发。重新读取由用户显式触发。

2026-09-12 实测：`make sdk` 干净构建 SDK 与三个版本样例包通过；随后最终示例 `npm run package && npm run package:fixtures` 通过。真实本地 HTTPS 双节点 fixture-115163169 上，`npm run test:e2e:real -- tests/real/extensions.spec.ts --grep 'persists metadata' --workers=1 --reporter=line` 1/1（4.7s）通过：未保存名称立即刷新恢复、真实后端改名、再次刷新没有重复 PATCH；测试最终恢复原实例名称并卸载样例。`go test -race ./internal/master -run TestIndependentReferencePackageResourceBridge -count=1` 通过（30.708s），测试管理器能力白名单同步新包。

此能力仅完成实例元数据分支。通知SDK、其他资源写入/控制能力和完整E06组合仍未完成；没有将本子项作为完整SDK完成证据。
