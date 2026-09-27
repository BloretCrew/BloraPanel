# Firefox / WebKit 真实恢复故障验收（2026-09-26）

关联 F02/F08/A11。使用既有 `web/tests/real/recovery-failure.spec.ts`，本轮没有改动测试断言、产品源码或 UI。补齐两个非 Chromium 引擎的真实 HTTPS 登录、生产桌面、Monaco 与 IndexedDB 恢复故障证据；Chromium 既有三项证据见[原恢复报告](recovery-quota-2026-09-13.md)。

## 环境与隔离

- 宿主：Fedora 43，Linux `6.19.12-200.fc43.x86_64`，x86_64；Node `v24.15.0`，npm `11.12.1`。
- 锁定测试包：`@playwright/test 1.63.0`。
- Firefox：宿主 Playwright `firefox-1543`，实际启动报告版本 `155.0`。
- WebKit：已有官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 容器，实际启动报告版本 `26.6`。项目只读挂载，只有本任务临时目录可写；不挂载 Docker socket、不共享宿主 PID namespace。
- 私有 fixture：真实 Master / 双 Daemon，独立 `127.0.0.1:9444`；状态目录 `/tmp/blora-recovery-engines-20260926/fixture-1878308687`。凭据仅由私有 JSON 文件读取，没有输出凭据值。每个测试均使用独立浏览器上下文和真实数据库。
- 运行端口、结果目录、浏览器容器名称均与并行的性能任务分开。

## 可复现命令

仓库根目录启动专用 fixture，等待其报告就绪；将 `FIXTURE_DIR` 设为实际创建的私有目录。本轮为上述 `fixture-1878308687`。

```sh
./dist/blora-devfixture -listen 127.0.0.1:9444 \
  -state-parent /tmp/blora-recovery-engines-20260926
```

Firefox 在 `web` 目录运行：

```sh
BLORA_E2E_CREDENTIALS="$FIXTURE_DIR/browser-credentials.json" \
BLORA_BROWSER=firefox \
npx playwright test --config playwright.real.config.ts \
  tests/real/recovery-failure.spec.ts --workers=1 --trace=off \
  --reporter=line --output=/tmp/blora-recovery-engines-20260926/firefox-results
```

WebKit 在仓库根目录运行：

```sh
docker run --rm --name blora-recovery-webkit-20260926 --network host \
  -e BLORA_E2E_CREDENTIALS="$FIXTURE_DIR/browser-credentials.json" \
  -e BLORA_BROWSER=webkit \
  -v /data/instances/blora-panel:/work:ro \
  -v /tmp/blora-recovery-engines-20260926:/tmp/blora-recovery-engines-20260926 \
  -w /work/web mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test --config playwright.real.config.ts \
  tests/real/recovery-failure.spec.ts --workers=1 --trace=off \
  --reporter=line --output=/tmp/blora-recovery-engines-20260926/webkit-results
```

## 实际结果与断言

| 引擎 | 结果 | Playwright 实际耗时 | 退出码 | 运行会话 |
| --- | --- | --- | --- | --- |
| Firefox 155.0 | 3/3 passed | 37.5s | 0 | 25810 |
| WebKit 26.6 | 3/3 passed | 50.8s | 0 | 50141 |

两套输出目录的 `.last-run.json` 都在清理前实际读取，内容为：

```json
{"status":"passed","failedTests":[]}
```

两个引擎均通过相同三项严格断言：

1. **配额写入失败与恢复**：测试页面的 `Storage.setItem`（仅恢复尾部键）及真实 IndexedDB `snapshots.put` 被注入 `QuotaExceededError`。最近中文输入仍显示在 Monaco；顶部转为“保护异常，导出工作区”；实际下载的 JSON 包含完整最新正文。恢复原始接口、再追加中文后转为已保护；刷新后完整正文保留。
2. **旧 schema 迁移与失败保留**：在真实 IndexedDB 中将当前快照改成 schemaVersion `0` / `tabId` 旧格式，刷新后完整正文恢复、保护正常；再改成不支持的 `999`，刷新后报告保护异常。实际下载含原始 `999` 记录及正文，再次刷新仍保留并可导出，没有被空工作区覆盖。
3. **真实事务中止与指针原子性**：在 `pointers.put` 请求已发出时调用该真实 `IDBTransaction.abort()`，先前的快照写入也回滚。实际读取的 `pointers` 与 `snapshots` 两集合和故障前深度相等；中止计数大于零。最新中文输入仍能导出，刷新后由同步尾部恢复完整正文且重新显示已保护；收集的 `pageerror` 严格等于空数组。

## 清理与边界

fixture 会话 `70874` 收到 Ctrl+C 后退出 `0`。两次检查确认 `9444` 已无监听、没有匹配本任务状态路径的进程；WebKit 测试容器及版本探针容器均以 `--rm` 删除，按容器名前缀复核无残留。随后只删除本任务创建的 `/tmp/blora-recovery-engines-20260926`，包括私有凭据和浏览器结果目录。

这些是两个实际浏览器引擎的恢复失败处理证据。配额异常采用定向 API 故障注入，没有耗尽宿主磁盘，也没有测量浏览器真实配额阈值；事务中止使用真实 IndexedDB 事务。结果不代表物理掉电、浏览器清除站点数据或 Windows 真机恢复已验证。UI 保持冻结。
