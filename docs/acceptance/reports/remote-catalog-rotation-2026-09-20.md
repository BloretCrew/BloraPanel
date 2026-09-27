# 扩展远程注册源的受控延迟与证书续期验收（2026-09-20）

## 验收范围

扩展 `TestExtensionRemoteCatalogAdminAPI`：真实 Master HTTP API 从独立 HTTPS 注册源读取清单并安装包。注册源通过 loopback 提供服务，每个响应增加 80ms 延迟；客户端信任专用测试 CA。浏览清单后，将服务端叶证书切换为由同一 CA 签发的新证书，关闭客户端空闲连接强制重新握手，再从新连接获取包并完成安装。另验证普通成员不能读取注册源。

## 结果

命令：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestExtensionRemoteCatalogAdminAPI$' -count=1 -v`

退出码 0；race 测试通过，总耗时 1.986s，测试本身 0.95s。浏览清单和安装阶段都断言服务端受控延迟已经经过；TLS 回调分别确认清单请求使用首张证书、安装请求使用续期叶证书。安装器仍校验清单身份和 SHA-256 后才保存包。

## 尚未覆盖

这是本机 loopback 上的受控 TLS 集成，不代表公网远程部署或真实高 RTT。轮换只更换同一受信 CA 签发的服务端叶证书，没有测试 CA 根更换、吊销或生产证书链部署；测试在 Master 集成层直接注入已构造的 RemoteCatalog，也没有启动 Master 二进制验证 `--extensions-catalog-url` 部署配置。因此 E07 的远端部署和高延迟/证书运维场景仍未完成。
