# 扩展任务查询、取消与恢复

2026-09-12，覆盖 F09/F14、A08/A13、E06/E07 子项。

SDK 增加 createTask/readTask/cancelTask，返回持久任务记录。Master 的扩展范围查询和取消验证应用身份、发起者及当前 app.use/node.read；禁用后允许获准用户处理历史任务，撤权立即拒绝。Daemon 区分执行前取消与授权失败，模块下载中的取消同样进入 CANCELLED。参考扩展保存 taskId，刷新只恢复查询，展示真实结果和取消状态。

验证命令与结果（均退出 0）：

- `cd web && npm test -- --run tests/extension-runtime.test.ts`：8/8。
- `cd web && npm run test:e2e -- tests/browser/extensions.spec.ts --workers=1 --trace=off --reporter=line`：4/4，12.9s；后端路由替身，覆盖刷新不重提交、取消及迁移。
- `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master ./internal/daemon -run 'TestExtensionAdminAPIInstallsAndServesPackage|TestCancelAcceptedUploadBeforeNodePreparation|TestFirewallLeaseRollsBackOnCancellationAndExpiry' -count=1`：Master 55.528s、Daemon 1.331s。真实本机 TLS/WSS、双节点，覆盖错应用/他人任务、重复取消、禁用历史查询与撤权拒绝。
- `GOCACHE=/tmp/blora-go-grant-idem make build` 与 `cd web && npm run build`：通过。
- 真实浏览器：`.local/fixture-3028559533`，HTTPS 127.0.0.1:9444，真实 Master + 两个 Daemon；`BLORA_E2E_CREDENTIALS=<fixture私有凭据路径> npm run test:e2e:real -- tests/real/extensions.spec.ts --grep 'independent package' --workers=1 --trace=off --reporter=line`：1/1，8.6s。完整包上传、节点窗口、WASI 计算 hello 世界（8 字符/2 词）、SDK 显示完成、刷新恢复笔记和结果且仅一次提交。fixture 随后 Ctrl+C 停止。

首次新路由与已有目录路由存在 Go ServeMux 模式冲突，启动测试失败；查询改为 `/extensions/{id}/tasks/{taskId}/status` 后复验通过。取消路径为同任务 `/cancel`，OpenAPI 与 SDK 文档已同步。

本轮没有全仓 race、Windows 真机、远程断网取消证据，不提升整项验收为通过。后端用户数据迁移、清单签名绑定、完整桌面 SDK 能力与真实两版本迁移仍需继续实现。
