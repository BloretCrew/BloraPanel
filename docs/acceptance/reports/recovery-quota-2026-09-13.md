# 浏览器配额失败与旧记录恢复（2026-09-13）

关联 A11/F02/F08。新增 `web/tests/real/recovery-failure.spec.ts`，实际 HTTPS 登录与生产桌面、Monaco、Chromium IndexedDB。故障注入只替换测试页面 Storage.setItem 和 snapshots.put，使其抛出 QuotaExceededError；没有模拟后端响应，也没有耗尽宿主磁盘。

命令：web 目录设置 `BLORA_E2E_CREDENTIALS` 指向 `.local/fixture-1827070879/browser-credentials.json`，运行 `npm run test:e2e:real -- recovery-failure.spec.ts --reporter=line`。会话37838退出0，2/2通过（4.6s）。

- 同步日志与 IndexedDB 均写入失败后，最近中文输入仍在 Monaco 中，顶部显示“保护异常 · 导出”。实际下载的 JSON 包含完整最新草稿。恢复存储接口后追加输入，状态恢复保护；刷新后完整正文保留。
- 将真实数据库当前快照改为旧 schemaVersion 0/tabId 格式，刷新恢复正文且恢复保护；再改为不支持的999版本，显示异常，下载保留原始版本及正文。再次刷新仍保留999记录并可导出，未用空工作区覆盖。

这些证据覆盖指定写入异常和格式迁移场景；未模拟物理掉电，也未证明所有浏览器引擎的真实配额阈值。快照写入事务的中途失败/指针原子性仍需独立故障验证，不能仅凭此浏览器通过推断。生产恢复代码本轮没有修改。fixture已Ctrl+C退出0，无活动测试进程。

后续事务中断验证：新增在 pointers.put 已发出时调用真实 IDBTransaction.abort 的场景，此时 snapshots.put 已成功回调。读取实际数据库确认 pointers 和 snapshots 集合与故障前完全一致，最新正文仍可导出；刷新后同步日志恢复完整正文并恢复保护。首轮会话82235失败于未捕获 AbortError（另外两项通过），原因是请求 Promise 拒绝被捕获，但独立的 tx.done 拒绝没有观察者。

修复 `web/src/recovery/service.ts`：创建事务后立即观察完成 Promise；失败时主动中止尚未完成的事务并等待其结束，再报告保护异常，防止请求失败时部分事务继续提交或遗留未处理拒绝。最终真实浏览器会话75942退出0，3/3（6.7s），无 pageerror；生产构建及46项前端单元测试退出0。fixture `.local/fixture-2619400585` 已Ctrl+C退出0。该证据补齐上述事务中途失败场景，不延伸为物理掉电验证。
