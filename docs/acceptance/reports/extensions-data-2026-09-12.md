# 扩展用户数据与升级迁移

2026-09-12，F14/E06/E07 子项。

`internal/extensions/userdata.go` 提供按认证用户隔离的 JSON 文档、revision 条件替换与持久请求回执；重复请求不覆盖后续写入。每文档 16 KiB，全部文档及回执共 4 MiB，超额拒绝而不修改旧文件。Master `/extensions/{id}/data` GET/PUT 再验证 app.use、启用状态及 data.read/data.write；owner 不从客户端获取，管理员也只读取自己的文档。SDK 和独立包提供真实读写入口。

Manifest 新增受签名保护的 dataSchemaVersion。安装保留数据、升级或回滚时，目标包 WASI 后台转换每位用户的当前数据；编译结果复用，每次新建独立内存实例，没有宿主文件、环境或网络。整批迁移 60 秒、每次执行 5 秒。所有结果校验后才与包事务共同提交，失败保留原包和数据。回滚转换当前数据，避免丢弃升级后的写入。注册表事务新增可选数据快照，兼容旧四文件恢复记录。

验证证据（退出 0）：

- 真实 TLS `TestExtensionDataAPIUsesAuthenticatedOwnerAndCurrentGrant`：0.162s；不同用户隔离、客户端伪造 owner 拒绝、重复/旧修订请求、撤权及禁用拒绝。
- `go test ./internal/extensions ./internal/master -run 'TestRegistry|TestUserData|TestExtensionData|TestIndependentReference' -count=1`：6.086s/4.607s。
- 编译复用后 `go test -race ./internal/extensions ./internal/master -run 'TestUserData|TestRegistry|TestBackendSandbox|TestExtensionData' -count=1`：138.095s/2.166s；含真实 Go WASI guest、双用户迁移/回滚/失败保留、沙箱输出及时间限制。
- 新增实际子进程中断覆盖数据与包，`go test -race ./internal/extensions -run TestRegistryRecoversInterruptedTransactions -count=1`：1.396s；安装/升级多个阶段退出，未提交恢复旧数据、已提交令牌保留新数据。
- `go test -race ./internal/extensions -run TestUserDataConcurrentCASAndQuotaKeepCommittedDocument -count=1`：2.661s；并发同修订只有一个成功，单文档超限拒绝，实际近 4 MiB 回执文件上的超额写入保留原字节。
- 前端 `npm test`：33/33；transport 9/9，验证稳定请求键及 undeclared 拒绝。SDK、前端 build、`make check build windows` 通过。OpenAPI 3.1 解析为 100 路径。
- 真实浏览器 `.local/fixture-1932133697`，HTTPS 127.0.0.1:9444，Master + 两个 Daemon：`BLORA_E2E_CREDENTIALS=<私有路径> npm run test:e2e:real -- tests/real/extensions.spec.ts --grep 'independent extension migrates' --workers=1 --trace=off --reporter=line` 1/1（11.7s）。独立构建 0.3.0/schema1、0.4.0/schema2、故意失败 0.5.0/schema3；保存服务器笔记，升级后窗口和服务器数据均恢复，失败升级保留两者，回滚后内容保留。fixture 已 Ctrl+C 停止。

Go 命令使用 Makefile 缓存设置及 GOFLAGS=-buildvcs=false；本机 TLS 测试在工具授权后运行。全仓 race 正在收取，结果单独补记。物理掉电/ENOSPC、Windows 真机、远程掉线迁移、大规模用户配额管理与升级异步任务体验仍未完整验证；本次不是 F14/E07 整项完成。

最终收取：`GOCACHE=/tmp/blora-go-grant-idem make test` 退出 0，Master 194.482s、extensions 144.750s，其余包通过。新增进程中断与配额测试在该命令编译后加入，已分别通过上述定向 race。`make sdk` 首次默认沙箱在 esbuild 安装校验的 spawnSync 遇 EPERM 退出 2；授权环境重跑退出 0，SDK 与示例干净依赖安装、类型检查、基础包和两个迁移样例包全部生成。没有跳过安装脚本或把首次失败计作通过。
