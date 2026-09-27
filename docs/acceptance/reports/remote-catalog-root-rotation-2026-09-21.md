# E07 注册源 CA 根轮换验证（2026-09-21）

## 命令与结果

在 Linux 本地真实 HTTPS 集成夹具上执行：

```sh
GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false \
  go test -race ./internal/master -run '^TestExtensionRemoteCatalogAdminAPI$' -count=1 -v
```

退出码 0，`TestExtensionRemoteCatalogAdminAPI` 通过（1.20 秒，race 检测开启；测试包总耗时 2.240 秒）。每个注册源 HTTPS 请求注入 80ms 延迟。

## 覆盖内容

- 原受信根下浏览 `v1.0.0`，并确认普通成员不能浏览管理注册源。
- 保持原 CA 不变、续期叶证书并强制新 TLS 握手后，仍可安装 `v1.0.0`。
- 将同一主机名下的注册源切换到新 CA 和 `v1.1.0` 后，旧信任配置下浏览失败、安装被拒，已安装的 `v1.0.0` 保持不变。
- 显式替换客户端 TLS 根并创建新 transport 后，才可浏览并安装 `v1.1.0`。

测试源码：[`internal/master/extensions_integration_test.go`](../../../internal/master/extensions_integration_test.go)，测试为 `TestExtensionRemoteCatalogAdminAPI`。

## 边界

这是 Master API 与本机 loopback HTTPS 测试注册源之间的集成验证。它证明旧根证书不会静默信任新 CA，并覆盖显式更新客户端信任后的成功路径；80ms 是受控单请求延迟，不能代表公网 RTT。公网注册源部署、真实证书/根更新操作链、进程启动参数或在线配置更新，以及 Windows 和长期远端网络均未由本测试验证，因此 E07 仍保持进行中。
