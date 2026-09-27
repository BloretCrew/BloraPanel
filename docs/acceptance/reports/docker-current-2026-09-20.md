# 当前源码真实 Docker 验收（2026-09-20）

在提权测试环境中确认 Docker Engine `29.4.1`、`overlayfs` 驱动和隔离 fixture 镜像 `blora-isolated-e2e:20260909` 可用后，执行：

```text
BLORA_CONTAINER_E2E=1 go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1
```

命令退出 0，测试耗时 27.681 秒。测试覆盖真实容器、镜像、卷、内部网络、双流日志、Compose 创建/更新/删除、健康失败和卷保留，并验证普通用户越权拒绝。测试只使用带 `blora.dev/e2e` 标签的随机对象，结束时已清理所属对象，未触碰既有镜像或宿主资源。

这份证据覆盖当前源码与本机 Docker Engine；远程 Engine、Windows 容器、物理掉电和长期故障组合仍按矩阵保留未验证。
