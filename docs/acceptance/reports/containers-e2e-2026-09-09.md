# 真实 Docker/Compose 生命周期验证

环境：Docker Engine 29.4.1，API 1.45，Compose 5.1.3；隔离测试镜像 `blora-isolated-e2e:20260909`，宿主 PTY 使用 `/tmp/blora-host-exec-helper`。

命令：

```sh
BLORA_CONTAINER_E2E=1 BLORA_HOST_EXEC_HELPER_BINARY=/tmp/blora-host-exec-helper \
BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 \
GOCACHE=/tmp/blora-go-container-cache \
go test ./internal/containers -run TestRealDockerComposeLifecycle -count=1 -v
```

结果：退出 0，25.651 秒。真实验证覆盖容器创建/运行/日志帧、镜像拉取及不可用 registry 失败、卷与网络创建/查询/删除、Compose 应用/更新/删除、健康检查失败诊断，以及删除容器/Compose 项目时显式保留数据卷。失败场景标记为 `INTERRUPTED`/`unknown=true` 或 `FAILED`，没有伪造成功。

限制：测试在单机 Docker daemon 完成，未覆盖远程 Engine、磁盘真实 ENOSPC、Windows 容器和完整浏览器到 Master 的 Docker UI 链路。

## 2026-09-10 停止响应与真实浏览器复验

真实浏览器 fixture 在 Compose 应用更新/删除流程中确认 Docker stop 可能在服务端保持响应最长 300 秒。客户端 `internal/containers/engine.go` 将 transport `ResponseHeaderTimeout` 调整为 5 分钟，具体操作仍由 context 提供更短截止；`GOCACHE=/tmp/blora-go-container-regression go test ./internal/containers -count=1` 退出码 0。

新的隔离 fixture `.local/fixture-3172017811` 上，Docker 浏览器用例从启动器打开 Docker 应用，提交 inline Compose 草稿并执行真实 apply/update/delete，**1/1 通过，43.6 秒**。随后真实全套浏览器回归 **13/13 通过，2.2 分钟**；fixture 已停止并清理自身测试资源。该证据仍限于本机 Docker daemon，不代表远程 Engine、ENOSPC 或 Windows 容器已验证。

## 2026-09-10 Docker-enabled 全量 fixture

新的 `.local/fixture-1866434427` 以 `--docker-endpoint unix:///var/run/docker.sock` 启动，真实 HTTPS Master、两个 Daemon 和本机 Docker 29.4.1 同时运行。完整命令：

```sh
BLORA_E2E_CREDENTIALS=/data/instances/blora-panel/.local/fixture-1866434427/browser-credentials.json \
npm run test:e2e:real -- --workers=1 --trace=off --reporter=line
```

结果：**15/15 通过，2.2 分钟，退出码 0**。Docker 场景实际创建/应用/更新/删除 Compose 项目并回读任务事实；同一套件还验证容器草稿刷新、扩展 sandbox 和云工作区，确保 Docker endpoint 不会破坏管理连接。fixture 结束后以 Ctrl+C 停止，测试对象按随机标签清理。

限制与前述相同：仅覆盖本机 Docker Engine，未覆盖远程 TLS Engine、真实 ENOSPC/掉电、Windows 容器和长时网络故障。

2026-09-14 复验：在 Docker Engine 29.4.1、API 1.45、Compose 5.1.3 上重新执行同一测试，使用 `BLORA_CONTAINER_E2E=1 BLORA_HOST_EXEC_HELPER_BINARY=/tmp/blora-host-exec-helper BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 go test ./internal/containers -run TestRealDockerComposeLifecycle -count=1 -v -timeout 5m`，25.063 秒退出 0。随机标签对象覆盖容器/镜像/卷/网络、双流日志、真实镜像拉取与失败、Compose 健康失败/更新/删除和卷保留；清理后未留下测试对象。限制仍为单机 Engine，远程 Engine、真实 ENOSPC/掉电、Windows 容器及长时网络故障未覆盖。

同日真实浏览器补验：启动任务自有双节点 HTTPS fixture（`--docker-endpoint unix:///var/run/docker.sock`），执行 `BLORA_E2E_CREDENTIALS=/tmp/blora-e02/fixture-2657977220/browser-credentials.json npm run test:e2e:real -- tests/real/docker.spec.ts --workers=1 --trace=off`，1/1 通过（44.5 秒）。覆盖容器中心草稿刷新、Compose 源保存与显式部署分离、中文输出/阶段分页、任务详情刷新、删除部署及卷保留；fixture Ctrl+C 正常退出，Docker 随机对象无残留。仍未覆盖远程 Engine、真实 ENOSPC/掉电、Windows 容器及长时网络故障。
