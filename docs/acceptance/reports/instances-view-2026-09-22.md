# 实例中心列表/卡片视图增量验收（2026-09-22）

实例中心现在提供列表与卡片两种呈现，默认保持原有列表密度。切换只改变当前实例总览的排布，筛选、选择、批量确认、资源身份和后端请求键不变；视图选择存入当前实例中心标签的现场，刷新后恢复。

实现位置：`web/src/apps/InstancesApp.vue` 的 `instanceViewMode` 与视图切换控件；`web/src/applications.css` 的 `.instance-grid.cards` 样式。资源标签页和实例操作流程保持原有实现。

## 验证

- `npm run check`：通过。
- `npm run test:e2e -- tests/browser/instance-batch.spec.ts`：1/1 通过（13.8s）。测试切换到卡片、刷新恢复，再切回列表后继续筛选、选择和部分接受批量启动；请求仍只提交一次并按原幂等键重试。

## 边界

卡片视图不改变实例查询数量或权限过滤；跨节点生命周期、远程网络和 Windows 行为继续由 F05/E09 的既有证据与环境缺口跟踪。
