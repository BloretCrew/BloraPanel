# Docker/Compose 实际 ENOSPC 验收（2026-09-21）

在本机 Docker Engine 29.4.1、API 1.45、overlayfs 和 Compose 5.1.3 上，扩展 `TestRealDockerComposeLifecycle`，通过随机带标签的临时 Compose 项目创建一个 1 MiB 容器 tmpfs，并在真实容器内尝试写入 2 MiB。测试使用预先存在的隔离镜像 `blora-isolated-e2e:20260909`，只给这个容器挂载 `/data` tmpfs，没有挂载宿主目录。

Docker 内核实际返回 `No space left on device`，容器以 exit code 1 退出。Compose apply 明确返回 `FAILED`、阶段 `health`、`Unknown=false`，同时保留 `Changed=true` 和唯一已创建容器事实；重读容器详情得到实际 exit code，容器日志保留 ENOSPC 诊断，未出现虚假的成功写入标记。随后通过 Compose delete 移除项目；测试自身按随机 `blora.dev/e2e` 标签回收本轮新建对象。

验证命令及结果：

- `env BLORA_CONTAINER_E2E=1 go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1 -v -timeout 5m`：退出 0；真实 E2E 29.17 秒，包总耗时 30.191 秒。
- `go test -race ./internal/containers -count=1`：沙箱内第一次因 `httptest` 无法绑定 IPv6 loopback 失败；在获准可绑定 loopback 的环境执行同一命令，退出 0，8.248 秒。
- `go vet ./internal/containers`：退出 0。

测试入口：[containers_e2e_test.go](../../../internal/containers/containers_e2e_test.go)。本证据证明真实 Docker 容器 tmpfs 耗尽时 Compose 的失败、部分资源和诊断语义，不代表 Engine 主机 overlay 数据根耗尽、物理掉电、远程 Engine、Windows 容器或长期网络故障已通过；这些仍按 E02 保留。
