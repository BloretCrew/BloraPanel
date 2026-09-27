# 云端工作区副本验证

云端工作区使用独立的 `workspace_cloud` 命名空间，以账号作为所有者并用 revision 做条件写入。默认“同步布局/引用”会剥离正文、草稿、撤销历史和终端现场；用户明确选择“同步工作内容”后才上传正文。打开云端副本前仍由桌面确认，恢复失败保留本地现场。

后端入口：`GET/PUT/DELETE /api/v1/workspaces` 及 `/api/v1/workspaces/{id}`。PUT/DELETE 在账号鉴权后要求 `Idempotency-Key`，并在同一 SQLite 事务中完成 revision CAS、写入/删除、审计和回执；同键 PUT 重试复用首次 metadata。实现位于 `internal/master/workspaces.go`，集成测试为 `internal/master/workspaces_integration_test.go`。

验证命令：

```sh
GOCACHE=/tmp/blora-go-workspace-cache go test ./internal/master -run TestCloudWorkspace -count=1
cd web
npm run test:e2e -- --workers=1 --trace=off --grep 'cloud workspace' --reporter=line
BLORA_E2E_CREDENTIALS=/受控fixture/browser-credentials.json \
  npm run test:e2e:real -- --workers=1 --trace=off --grep 'real cloud workspace' --reporter=line
```

结果：Master 工作区集成测试通过；浏览器 mock 场景验证默认同步不包含 `state.drafts`、显式正文同步产生 revision 2、列表和打开动作；真实 HTTPS fixture `.local/fixture-1866434427` 的 `workspaces.spec.ts` **1/1 通过**。真实场景在 Monaco 中输入中文草稿，先验证布局/引用同步的云端正文为空，再显式同步正文，读取 revision 2 的副本并恢复正文。

未覆盖：跨设备长时间并发写入、浏览器存储耗尽、断电迁移和高延迟网络；这些场景仍按 F02/A11/E08 保持进行中。
