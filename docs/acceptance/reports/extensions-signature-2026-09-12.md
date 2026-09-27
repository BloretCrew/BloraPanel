# 清单与包体联合签名

2026-09-12，F14/E07 子项。此前 Ed25519 只认证 payload SHA-256，无法阻止清单能力、应用身份或版本被替换。现在 `internal/extensions/signature.go` 使用带版本域的规范消息，同时认证所有受支持 Manifest 字段和实际包体摘要；安装默认值归一化后持久读取仍可验证。旧包体单独签名明确拒绝，不提供降级验证分支。

新增 `cmd/extension-sign`：读取有界包 JSON 和 Ed25519 PKCS#8 PEM，验证原摘要，生成新签名包；拒绝覆盖已有输出，只显示公钥和输出路径。SDK README 提供命令与旧包重新签名说明。

真实验证（退出 0）：

- `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./cmd/extension-sign ./internal/extensions -run 'TestSignedPackage|TestTrustedPackageSignature|TestRegistry' -count=1`：签名工具 1.027s、扩展 1.322s。生成真实 Ed25519 密钥、运行签名工具、安装、重新打开注册表、篡改已存清单后读取拒绝；身份/版本/能力/权限/状态版本/入口/资源/依赖/包体篡改与旧签名拒绝。
- `go test ./internal/extensions ./internal/master -run 'TestTrustedPackageSignature|TestExtension' -count=1`（同缓存环境，本机 TLS 监听授权）：extensions 0.011s、Master 7.559s。
- `GOCACHE=/tmp/blora-go-grant-idem make check windows`：通过。Windows 仅交叉构建，未作真机签名/注册表验收。
- OpenAPI 经 python3/PyYAML 解析：3.1.0、99 路径、113 操作。

签名工具验证初次调用时测试文件未成功写入，显示 no test files，未作为测试通过；补齐文件后运行上述真实测试。没有公开发布签名包或使用生产密钥。签名策略变化要求原签名包由发布者重新签名；无可信公钥配置时的管理员本地无签名安装行为保持。后端数据迁移和 E07 其余完整组合继续进行中。
