# Blora Panel

基于 Go Master/Daemon 与 Vue 3/TypeScript 的桌面式多节点服务器管理面板。已有真实用户/节点/实例控制、文件与编辑器、PTY、角色模板、节点维护/换钥、监控、Docker/Compose、备份调度、有限系统管理、扩展安装生命周期、本地/HTTPS 注册源和云端工作区副本链路；全范围验收仍按细项区分已验证、未验证和环境缺失。逐项状态见[验收矩阵](docs/acceptance/ACCEPTANCE_MATRIX.md)，运行会话与下一步见[执行进度](docs/execution/PROGRESS.md)。

平台只承载管理通信，被管理服务的业务网络独立配置。

正式运行的目录布局、健康检查、停机一致性备份、升级与回退步骤见[运维手册](docs/operations/OPERATIONS.md)。

Windows、systemd、远程 Engine 和长期故障场景的真实验收入口见[平台验收运行手册](docs/operations/PLATFORM_VALIDATION.md)。

## 构建与本机验收

当前已验证环境为 Go 1.25.9、Node 24.15.0/npm 11.12.1、Linux amd64。依赖锁定于 go.mod/go.sum 和 web/package-lock.json。

```sh
make build
make sdk
cd web
npm ci
npm run build
cd ..
make fixture
```

构建产物为 `dist/blora-master`、`dist/blora-daemon` 与 `web/dist`。`make windows` 生成 Windows amd64 双二进制；构建通过不代表 Windows Job/ConPTY 真机验收通过。

源码仓库包含构建与验收入口，不包含本地凭据、运行数据库、依赖目录、发行归档和历史截图图库。`make sdk` 安装锁定依赖，并从参考扩展源码生成默认包及两个升级验证包；执行扩展相关验收前需完成此步骤。文字验收报告保留，历史截图链接对应本地开发归档。

本地发行包：先在 `web`、`sdk` 和 `sdk/examples/reference-app` 分别执行 `npm ci`，然后在根目录执行 `BLORA_VERSION=development-YYYYMMDD make package`（另需 Python 3.9+）。输出位于 `dist/releases/<版本>/`：Linux/Windows Master、Daemon，以及独立前端和 SDK 共六个包。Master 包含前端，Daemon 包含隔离镜像构建输入，SDK 包含参考扩展源码、构建产物与两平台签名工具。包不包含运行身份或 node_modules；SDK 解包后的构建步骤见 `START.txt`。

先用 `sha256sum -c SHA256SUMS` 验证归档，每个包内另有逐文件 `MANIFEST.json`。打包固定时间戳、文件顺序和权限；同一输入重复打包内容一致。已有版本内容不同会拒绝覆盖，源码或文档改变后应选择新版本。这是本地开发交付，不自动发布；具体验证边界见[发行包报告](docs/acceptance/reports/package-2026-09-13.md)。

`make package` 会从已编译的 Go 依赖与锁定的本地 npm 包收集上游许可证和通知，见 [依赖清单](docs/licenses/inventory.json) 与 [通知原文](docs/licenses/THIRD-PARTY-NOTICES.txt)。Master/Daemon包在`docs/licenses/`中保留这些材料，独立前端和SDK包根目录另带`THIRD-PARTY-NOTICES.txt`。收集器离线运行；依赖或许可证文件缺失会停止打包。若系统把Go许可证移到GOROOT之外，可用`BLORA_GO_LICENSE_DIR`指定实际安装包的许可证目录，Fedora默认识别`/usr/share/licenses/golang`。这些材料不替项目自身选择许可证。

fixture创建新的 `.local/fixture-*`，启动独立Master与两个Daemon，默认只监听 `https://127.0.0.1:9443`。管理员和成员密码每次随机生成，只写该目录权限0600的 `browser-credentials.json`；没有可复用默认密码。工具新建本机测试证书。

成员可见两个测试实例，仅对其中一个拥有部分控制/文件写权限，没有创建实例或主机Shell权限。实例初始停止，打开快捷方式不会启动实例。Ctrl+C先停止本fixture自己的实例，再退出三个自身进程；日志与目录保留用于诊断。

## 独立初始化和运行

先用本地编辑器创建权限0600的初始密码文件，内容为12～72字节密码，不把密码写入命令行或仓库。然后初始化一次：

```sh
./dist/blora-master --state-dir .local/master --init --password-file /absolute/private/initial-password
./dist/blora-master --state-dir .local/master --listen 127.0.0.1:8443 --origin https://localhost:8443 --static-dir web/dist
```

重复初始化会拒绝覆盖已有账号。初始化生成本机TLS证书；部署可用 `--tls-cert`、`--tls-key` 指定管理域名证书，`--origin` 必须与浏览器来源一致。本任务未公开部署。

登录后在节点应用生成一次性登记票据，保存为 Daemon 所在机器的私有文本文件。准备独立配置，例如：

```json
{
  "stateDir": "/absolute/private/blora-node-a",
  "masterUrl": "https://localhost:8443",
  "caFile": "/absolute/private/master-ca.crt",
  "enrollmentFile": "/absolute/private/node-a.enrollment",
  "allowPGIDFallback": false
}
```

```sh
./dist/blora-daemon --config /absolute/private/node-a.json
```

每个Daemon使用独立stateDir、票据和身份。caFile是受信管理CA证书，本机测试可用Master生成的tls.crt。节点主动连接Master并校验证书，已有身份不重新消费登记票据。

Linux原生实例需要已委派且可写的 `cgroupRoot`。信任执行内容的管理员可显式选择 `allowPGIDFallback:true`；进程组模式不提供普通用户宿主机隔离。普通用户使用隔离容器，需要管理员配置 `dockerEndpoint`、准备镜像并授予资源权限。

节点维护、配额、角色模板、禁用/改密和换钥见[管理能力报告](docs/acceptance/reports/administration.md)。换钥保持nodeId与关联资源。独立日志helper和Windows Job keeper由Daemon自身内部入口启动，不依赖网页保持打开。

## 隔离容器镜像

```sh
docker build -f Dockerfile.isolated -t blora/isolated:local .
```

镜像预装固定路径exec-helper并使用非root用户。运行适配器限制容器权限、资源和绑定目录；PTY关闭会核对出生身份，只结束对应exec。真实命令与边界见[容器终端报告](docs/acceptance/reports/terminal-container-2026-09-09.md)。Docker Center/Compose、容器日志有界归档、备份恢复和调度 API 已接入；远程 Engine、ENOSPC/掉电和长期故障场景仍需验收。

Linux Daemon启动时会先核对并清理自身遗留的Compose CLI。若提示`recover Compose CLI ownership`，需保留`containers/cli-runs/`记录，排查文件完整性与进程检查权限，再重新启动；无法确认旧CLI退出时不会接受新的Engine变更。CLI退出也不代表远程资源已回滚，应查看原任务诊断并核对实际资源。详见[恢复证据](docs/acceptance/reports/compose-cli-recovery-2026-09-13.md)。

## 默认应用与验证

已接 AppHost 的入口包括实例中心、文件、编辑器、终端、任务、节点、用户与权限、设置、Docker/Compose、监控、有限系统管理和扩展管理。独立扩展 SDK 与参考包位于 [`sdk/`](sdk/README.md)，可从独立目录构建；启用扩展通过认证 bundle 接口在 opaque-origin `sandbox="allow-scripts"` iframe 中加载，并经受控任务桥接、窗口和状态合同接入桌面。管理员可以浏览本地已校验注册源、获取包并安装或升级；网络注册源和 Windows/远程高延迟场景仍按验收矩阵记录。

备份与恢复从实例窗口的“打开备份与计划”进入，使用服务端版本核对、不可变恢复计划、显式覆盖确认和持久任务。Daemon 的 `backupHookCommands` 是可选的管理员固定 argv 映射；例如：

```json
{
  "backupRoot": "/absolute/private/node-a-backups",
  "backupHookCommands": {
    "backup.save.before": ["/absolute/private/bin/app-save", "--instance"],
    "backup.save.after": ["/absolute/private/bin/app-resume", "--instance"]
  }
}
```

钩子不会经过 shell，资源身份和 hook ID 通过 `BLORA_BACKUP_*` 环境变量传入；未配置的策略会明确返回能力不可用。扩展目录既可用 `--extensions-catalog /absolute/private/catalog` 读取本地目录，也可用 `--extensions-catalog-url https://registry.example.invalid/blora/` 拉取 `index.json` 与版本包；两者不能同时配置，远程源必须是 HTTPS 且拒绝重定向。

HTTP 管理接口的请求、授权、幂等、任务、WebSocket 和结构化错误合同见 [OpenAPI 描述](docs/api/openapi.yaml)。文档中的 `sessionCookie`、`X-CSRF-Token` 与 `Idempotency-Key` 对应真实 Master 校验；OpenAPI 的通用 JSON 响应只表示模块返回的具体资源由相应应用定义。

```sh
make check
make test
cd web
npm test
npm run test:e2e
```

真实浏览器测试读取fixture私有凭据文件，具体命令见[桌面报告](docs/acceptance/reports/desktop-2026-09-09.md)。测试只管理自己创建的资源；Windows、cgroup委派、Docker、磁盘满等环境缺口分别记录。

持续混合负载入口：先运行 `./dist/blora-devfixture --performance`，然后在另一终端进入 `web`，执行 `BLORA_E2E_CREDENTIALS=/absolute/path/to/fixture/browser-credentials.json BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`。使用夹具实际输出的私有凭据路径，勿复制凭据内容。该场景打开八个窗口、两个持续输出的真实PTY、万项目录并循环执行已验证的跨节点复制；逐轮检查终端错误，每分钟报告进度。运行期间不要替换前端构建或并行执行重负载测试。Linux可用 `python3 scripts/performance-resources.py .local/fixture-实际编号 --samples 62 --interval 60` 只读采样节点RSS和归档轮转。测试终态后停止采样并对夹具按Ctrl+C；观察心跳或短时通过不等于一小时通过。当前结果与已知失败见[持续负载](docs/acceptance/reports/performance-continuous-2026-09-19.md)和[终端合批](docs/acceptance/reports/terminal-batching-2026-09-19.md)。

Linux真实空间不足验收可显式运行 `BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'ENOSPC' -count=1 -v`。需要系统允许私有user/mount namespace；测试自行隔离后仅在所属临时目录挂载1MiB tmpfs，验证跨节点移动与备份创建/恢复，不填满宿主磁盘。默认未启用时跳过，不计为通过；见[ENOSPC证据](docs/acceptance/reports/enospc-2026-09-19.md)。

独立防火墙验收：`BLORA_TEST_FIREWALL_NAMESPACE=1 go test -race ./internal/master -run '^TestPrivateFirewalldApplyAndRestore$' -count=1 -v`。需要Linux私有user/mount/net namespace及firewalld、firewall-cmd、nft、ip、dbus-broker-launch。入口先验证隔离并隐藏宿主总线，仅操作测试网络；包含真实管理连接阻断后的租约回滚。默认跳过不计通过，见[防火墙证据](docs/acceptance/reports/firewall-private-2026-09-19.md)。

证据：[核心](docs/acceptance/reports/core-2026-09-09.md)、[运行适配](docs/acceptance/reports/runtime-2026-09-09.md)、[流与日志](docs/acceptance/reports/stream-integration.md)、[文件API](docs/acceptance/reports/files-api.md)、[Docker/Compose](docs/acceptance/reports/containers-api.md)、[备份与调度](docs/acceptance/reports/backup-scheduler-library.md)。原始需求以[技术架构](docs/plan/Blora-01-技术架构.md)、[功能交互](docs/plan/Blora-02-功能与交互.md)及[行动指导](ACTION_GUIDE.md)为准。
