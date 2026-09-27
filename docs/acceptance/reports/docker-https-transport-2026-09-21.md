# 真实 Docker Engine 的 HTTPS 传输验收（2026-09-21）

关联 E02/F11。扩展 `TestRealDockerComposeLifecycle` 的 opt-in 测试路径：测试生成短期私有 CA 与 Engine 证书，只监听 `127.0.0.1`，通过 TLS 终止代理把明文请求转发到本机 Docker Unix socket。Blora Engine 客户端使用 `https://127.0.0.1:<port>` 和 CA；Compose CLI 使用同一 TLS 目录及 `DOCKER_TLS_VERIFY`。对象仍由实际 Docker Engine 29.4.1/overlayfs 执行。本测试没有开放 host-wide Docker TCP 端口。

验证命令：

```sh
env BLORA_CONTAINER_E2E=1 BLORA_CONTAINER_E2E_REMOTE_TLS=1 \
  GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go \
  GOCACHE=/tmp/blora-go-containers-remote-tls GOFLAGS=-buildvcs=false \
  go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1 -v -timeout 5m
```

退出 0，E2E 30.49 秒、包 31.52 秒。Engine API 1.45、Compose 5.1.3；容器创建/停止、stdout/stderr 日志、Compose 创建/更新/删除、健康失败状态、卷数据保留及 1 MiB tmpfs 实际 ENOSPC 均通过。随机 `blora-e2e-*` 资源经现有 defer cleanup 按测试标签清除。`go vet ./internal/containers` 通过。

这验证了真实 Engine 经已验证 HTTPS endpoint 及 Compose TLS 客户端的传输路径；TLS endpoint 回到同机 Unix socket，不证明远端 Docker 主机、高 RTT/丢包、远端存储/掉电或 Windows Engine。E02 的真正远端 Engine 验收仍待具备远端测试环境后完成。
