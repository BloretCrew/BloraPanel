# 空依赖缓存构建与产物一致性

关联E09。开始前`test ! -e /tmp/blora-cold-20260919-1527`退出0，使用全新Go模块缓存、Go构建缓存、GOPATH和npm缓存；不是继续依赖已有模块/构建缓存。工具链沿用本机已安装Go/Node，不声称换成了新的操作系统镜像。

实际命令（均在仓库，SDK命令在相应目录）：

```sh
env GOCACHE=/tmp/blora-cold-20260919-1527/go-build GOMODCACHE=/tmp/blora-cold-20260919-1527/go-mod GOPATH=/tmp/blora-cold-20260919-1527/go GOTOOLCHAIN=local GOFLAGS=-buildvcs=false go build -trimpath -o /tmp/blora-cold-20260919-1527/bin/ ./cmd/master ./cmd/daemon ./cmd/extension-sign
env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOCACHE=/tmp/blora-cold-20260919-1527/go-build GOMODCACHE=/tmp/blora-cold-20260919-1527/go-mod GOPATH=/tmp/blora-cold-20260919-1527/go GOTOOLCHAIN=local GOFLAGS=-buildvcs=false go build -trimpath -o /tmp/blora-cold-20260919-1527/windows/ ./cmd/master ./cmd/daemon ./cmd/extension-sign
env GOMODCACHE=/tmp/blora-cold-20260919-1527/go-mod GOPATH=/tmp/blora-cold-20260919-1527/go GOTOOLCHAIN=local go mod verify
npm ci --cache /tmp/blora-cold-20260919-1527/npm-cache --no-audit --no-fund
```

Linux和Windows三程序构建均退出0，实际日志显示从公开模块源下载锁定依赖，最终`all modules verified`。首次在Windows依赖尚下载时从受限网络运行verify失败于DNS权限，未计通过；等构建完成后以同样允许访问公开模块源的环境复验退出0。六个冷构建二进制与rc4对应输入逐字节一致。

在`web`、`sdk`、`sdk/examples/reference-app`分别执行上面的npm ci，实际安装120/1/4包，随后分别运行`npm run build`、`npm run build`、`npm run package`及`npm run package:fixtures`，均退出0。前端115个产物文件SHA-256与rc4独立前端包MANIFEST全部一致；SDK/参考包源码及生成文件亦按rc4 SDK MANIFEST逐项比较。没有为了产生相同产物再生成新的发行版本。

相同字节的Linux发行包已独立初始化/TLS/登录/双Daemon运行通过，见[rc4报告](release-rc4-2026-09-19.md)。Windows仍为构建证据，Job/ConPTY/SCM运行需要Windows真机；其他操作系统工具链组合不由本次结果覆盖。
