# Docker 日志归档入口增量（2026-09-22）

## 结论

现有 Master/Daemon 已经按容器身份持久化最多 100 个有界日志窗口，并通过 `GET /api/v1/nodes/{id}/docker/containers/{containerId}/logs/history` 返回归档和 `possibleGap` 标记。本轮把这个能力接入 `DockerLogs.vue`：实时流仍然使用管理协议和字节确认；用户点击“查看归档”后先关闭实时连接，再读取最多 100 个窗口，按观察时间展示每个窗口的截断/可能缺口状态；“返回实时日志”会重新建立实时流。读取失败只显示错误，不把空结果当成成功历史。

## 代码入口

- `web/src/apps/DockerLogs.vue`：实时/归档模式、API 请求、窗口解码、缺口提示和生命周期关闭。
- `internal/master/container_logs.go`、`internal/daemon/container_logs.go`：现有有界归档接口及持久化实现，本轮未扩大后端预算。
- `web/tests/browser/docker.spec.ts`：归档浏览器链路。
- `web/tests/real/docker.spec.ts`：真实 Docker 容器的实时/归档 UI 链路。

## 验证

- `cd web && npm run check`：退出 0。
- `cd web && npm run test -- --run`：11 个文件、63/63 通过。
- `cd web && npm run build`：退出 0；仅保留既有 Monaco 大包提示。
- `make check`：`go vet ./...` 退出 0。
- `cd web && npm run test:e2e -- --grep 'Docker app'`：**2/2 通过，11.1s**。测试覆盖容器列表进入日志、受限历史 API、归档正文、`可能存在缺口` 提示和返回实时模式。
- `env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 go test -race ./internal/master -run '^TestContainerNodeAPIRealEngineLifecycle$' -count=1 -v -timeout=180s`：**2.79s，退出 0**。真实隔离 Docker Engine 中验证日志窗口持久化后由 Master 历史 API 读取、未授权访问返回 403、`limit=101` 返回 400，且归档容器身份和 `possibleGap` 保持一致。
- `BLORA_E2E_CREDENTIALS=<隔离 fixture 私有凭据> npm run test:e2e:real -- tests/real/docker.spec.ts --grep 'real Docker log view' --workers=1 --trace=off --reporter=line`：**1/1 通过，7.6s**。真实 Chromium、HTTPS Master/Daemon 和 Docker Engine 中创建并启动测试容器，验证实时中文日志、归档窗口正文、`可能存在缺口` 提示、返回实时模式及最终停止/删除清理。
- 同一真实 fixture 的原有 Compose 长流程 `--grep 'actual Docker and Compose'`：**1/1 通过，45.6s**，配置草稿刷新、显式应用、阶段输出分页/刷新恢复、删除部署和独立卷清理保持通过。

## 边界

归档是 Docker 驱动保留的时间窗口，不是连续事件日志；窗口可能重叠、截断或有缺口，前端明确保留这些状态。当前证据仍是 Linux/Chromium 本地浏览器和已有真实 Docker 后端测试；远程 Engine、高 RTT/丢包、Windows 容器、主机存储耗尽/物理掉电及长时故障组合仍按 F11/E02 记录为未验证。旧 Compose 长流程验收超时留下的本轮容器/卷已按本轮 Compose 标签识别并清理，未触碰其他 Docker 资源。
