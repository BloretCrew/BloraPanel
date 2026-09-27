# B08 Docker / Compose 模块验证 — 2026-09-09

范围：`internal/containers/**`，对应 F11、E02 的节点适配模块。此报告不是完整 Docker Center 前后端验收，也不把 Windows 交叉编译当作真机运行。

## 实现

- 直接使用 Docker Engine API 1.45，连接前检查服务端 API 上下界；Docker/Compose 依赖缺失显示 capability 原因，不隐式安装。
- 容器创建、启动、有限等待停止、先停后启重启、删除；镜像拉取和删除；卷、网络创建和删除。每次查询、执行和分片请求都要求该节点 `host.manage`。默认拒绝缺失授权函数/主体，取消上下文不产生新操作。
- 操作先写持久日记，稳定 task ID 绑定主体和完整参数。丢失 Engine 响应、进程中断后不重放创建/应用/删除；结果不确定显示 `INTERRUPTED/unknown`。宿主 Docker 变更在本适配器内串行，防止 Compose 和直接容器/卷入口交叉修改同一对象；实例隔离运行适配器独立。
- 容器、镜像、网络使用完整不可变 ID；卷使用名称、创建时间和内容指纹。删除卷在进度回调后再次检查身份，使用 `force=false`。**Engine 的卷删除没有原子 If-Match；模块串行和两次核对不能约束同时运行的外部 Docker CLI 在最后检查后替换同名卷。**
- Compose 保存不可变配置修订，保存与应用分开。应用记录配置校验、拉取、构建、创建、启动、健康阶段；失败保留实际资源和阶段结果，不承诺自动回滚。删除只删除固定项目容器和网络，默认保留卷；删除卷必须传入额外明确的目标身份。
- 不仅检查 Compose CLI 退出码，还在独立截止时间内检查实际服务数量、容器状态和健康状态。`starting` 不当作健康成功；配置的完成型依赖按实际退出码确认。
- Compose 正文上限 1 MiB，采用最多 64 KiB 的分片保存/读取，避免超过节点 RPC 大小。prepare 绑定 actor/saveID/projectID/base revision/长度/SHA-256；每片校验并 fsync 后确认；commit 前持久化 COMMITTING 日记。不可变修订保存提交归属和源哈希，重启/丢 ACK 后可核对早于当前项目版本的已提交修订。提交冲突不覆盖新正文，保存不执行 Compose。
- Docker 日志保留 stdout/stderr 和 Engine 原始时间戳，每次回调最多 32 KiB，支持 tail/since/follow 和取消。流空闲时每秒复权；撤权关闭实际 Engine 请求。日志时间窗口可能重叠或存在留存缺口，没有伪造终端事件游标。

## 环境与命令

环境：Fedora Linux amd64，Go 1.25.9；Docker Engine 29.4.1（服务端 API 1.54 / 最低 1.40），Compose CLI 5.1.3。模块协商并使用 API 1.45。

所有 Go 命令共用环境：

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E=1 go test -race -v ./internal/containers
```

最终结果：**exit 0，27.058s**；`TestRealDockerComposeLifecycle` 24.97s，其余故障/恢复/授权测试通过，race detector 无报告。

真实环境测试默认使用 `unix:///var/run/docker.sock` 和已构建专用镜像 `blora-isolated-e2e:20260909`；可用 `BLORA_CONTAINER_E2E_ENDPOINT`、`BLORA_CONTAINER_E2E_IMAGE` 指定受控环境。未设置 `BLORA_CONTAINER_E2E=1` 时此项明确跳过，跳过不计真实 Docker 验证通过。

测试对象使用随机 `blora-e2e-*` 名称和 `blora.dev/e2e=<本次随机身份>` 标签。只清理本次标签对象；不运行 prune，不删除共享 fixture 镜像，不挂载宿主目录。测试中新卷初始化容器显式使用容器内 UID 0，不赋予 privileged、宿主网络或 Docker socket。

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/blora-containers-windows.test.exe ./internal/containers
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go vet ./internal/containers
```

两命令均 **exit 0**。Windows 产物为 `/tmp/blora-containers-windows.test.exe`；没有 Windows 真机结果。

## 场景与实际结果

| 场景 | 证据/结果 |
| --- | --- |
| 真实宿主 Docker capability 与普通主体拒绝 | Engine/Compose 可用；普通主体查询得到 ErrForbidden |
| 容器 create/start/restart/stop/delete | 真实完整 ID 操作通过，重启先核对停止结果 |
| 日志 stdout/stderr | 读取真实时间戳和两个不同 stream 的 marker |
| 删除容器默认保留数据 | 删除后创建另一个仅挂载测试卷的容器，读到 `persistent-marker` |
| 镜像拉取与删除 | 临时 loopback registry 提供新生成、带测试标签的零层镜像；实际 Engine pull 后按不可变 image ID 删除。另从自己的停止容器 commit 独立镜像并删除，不删除 fixture |
| 镜像拉取失败 | `127.0.0.1:1/<随机测试名>:missing` 真实拒绝连接；返回 `INTERRUPTED/unknown=true` 与 Engine 诊断，没有报告成功 |
| 网络和卷创建/删除 | 只操作本次对象，显式卷身份删除通过 |
| Compose 保存与应用 | 保存版本 1 不部署，显式 apply 后实际运行，阶段含 validate-config/create/start/health |
| Compose 更新 | 保存版本 2 并显式 apply；服务完整容器 ID 改变，确实发生重新创建 |
| Compose 删除保留卷 | 删除项目后实际卷仍可 inspect，后续显式选择该卷才删除 |
| Compose 实际健康失败 | healthcheck `/bin/false`；返回 FAILED/health，保留 1 个实际容器和阶段事实，随后只清理本次项目 |
| Engine 丢 ACK | 故障注入 HTTP 服务接到创建后断连接；重开 Manager/重复同 task ID 后收到 ErrUnknown，创建计数保持 1 |
| 卷替换身份 | 故障注入在进度回调后替换同名卷指纹；ErrIdentity，DELETE 计数为 0 |
| 空闲日志流撤权 | 故障注入保留空闲真实 HTTP 流；撤权后约 1s 返回 ErrForbidden 并断开请求 |
| 分片与恢复 | 大正文逐片持久化，重复片 ACK、actor 隔离、权限撤销、重开 Manager 续交、整份 SHA-256 与分块读取一致 |
| 提交 ACK 丢失与后续版本 | 将已提交日记恢复为 COMMITTING，再保存下一修订，重开 Manager 后仍核对到早期提交，未重复产生新修订 |
| 保存冲突/未确认字节 | 未确认临时字节被覆盖；错误 offset/hash 拒绝；旧版本提交不覆盖新的项目正文 |

测试替身场景仅证明其对应故障处理，不替代上表真实 Engine/Compose 场景。

## 失败记录与修复

1. 第一轮真实测试 **exit 1，11.111s**：fixture 默认 UID 65534 无法写新卷根目录，后续读取没有数据。修正受控测试的卷初始化 UID 为容器内 0 后，真实生命周期首次通过 **exit 0，20.863s**。
2. 加入健康失败验收后 **exit 1，26.005s**：Compose 5.1.3 的 `start --wait --wait-timeout 2` 在本次环境返回 0，但 Engine 报告 health=`starting`。旧代码仅拒绝 `unhealthy`，造成错误成功。保留失败测试，增加实际服务/健康状态和独立截止时间验证后最终通过。
3. 审查卷删除发现第二次 inspect 后又触发一次进度回调。已把复核放到最后一次进度回调后，复核与 DELETE 之间不再调用外部通知；保留 Engine 不提供原子 If-Match 的实际限制。

## 尚未覆盖与集成边界

- 任意宿主容器交互 exec 已新增 `runtime.HostDockerExecutor` / `containerterm.HostAdapter`，复用现有 terminal session/lease 与 Docker PTY helper；原实例 owner/run/隔离检查保留。适配器经真实 session/lease/归档测试通过；产品 HTTP/Daemon 默认注册和 Docker Center 用户入口仍须主代理接线验收。
- Master/Daemon 权限、任务桥接和 Compose 分片 HTTP API、Docker Center 前端、配置草稿 F5 恢复由主代理集成并单独记录；此模块通过不能把 F11/E02 全项标为完成。
- Windows 原生 Docker/Compose/Job 运行和外部远程 HTTPS Engine/Compose 环境未运行。新增本地临时端点双向 TLS 验证见下；跨编通过只证明对应平台源码可编译。
- 真实 registry 测试验证本地固定镜像获取；认证 registry、多层大镜像下载、Compose build 项目、外部 CLI 并发替换资源未在本报告中验证。

实现 API 依据：[Docker Engine API 1.45](https://docs.docker.com/reference/api/engine/version/v1.45/)、[Compose start 的 wait 语义](https://docs.docker.com/reference/cli/docker/compose/start/)、[Compose up 的重新创建与卷保留语义](https://docs.docker.com/reference/cli/docker/compose/up/)。实际运行结果以上述命令和断言为准。

## 后续补充：TLS 与宿主终端适配

- `internal/dockerapi/tls.go` 为 Engine 和 host exec 复用 TLS 配置，Compose 使用同目录 `ca.pem/cert.pem/key.pem`；默认系统根证书，指定目录时追加 CA 并加载客户端证书，保持证书链/主机名校验和 TLS >=1.2。
- `TestEngineMutualTLSAndUntrustedServerRejected` 创建一小时有效期的临时 CA、服务端和客户端证书，在 loopback 上运行真正的 RequireAndVerifyClientCert TLS 服务。已验证正确双向握手、未知私有 CA 拒绝和非法 CA 文件拒绝，初轮 race 结果 **exit 0，1.382s**。未向系统证书库写入内容。
- Host 终端绑定 `HostDockerTarget{ContainerID,CreatedAt,StartedAt}`，`Reference()` 使用全部三字段 SHA-256，节点 ResourceRef 与 actor 一起存储；新 exec/input/resize 重新授权。独立 helper 清理仍仅持有已创建会话的 PID、出生和 token，不因用户后来撤权而失去清理能力。
- 主机/共享 PID namespace 的容器不能使用此 helper 路径；平台托管实例必须使用原实例终端入口。完整 container ID、创建时间或启动时间改变均拒绝。
- `exec-helper` 新增 child-subreaper 与父子树清理，覆盖 `setsid` 离开原会话的后代，保留 pidfd 出生核对及无关进程保护。宿主入口明确要求 `Probe.processTree=true`；旧 helper 不能被当作已经升级。
- helper 真实 Linux PTY 测试含错误出生、错误 token、原会话子进程、`setsid` 后代、其他无关进程和命令失败可见性；`go test -race ./internal/containerterm/helper` **exit 0，4.315s**。新增测试 helper 构建到 `/tmp/blora-host-exec-helper`，没有替换共享 fixture 镜像。
- 原实例 runtime 回归和新的宿主身份边界测试：`go test -race ./internal/runtime` **exit 0，2.652s**；Windows amd64 runtime 测试产物 `/tmp/blora-runtime-host-windows.test.exe` 交叉编译 **exit 0**，非真机运行。

### 宿主终端真实验证

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/blora-host-exec-helper ./cmd/exec-helper
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E=1 BLORA_HOST_EXEC_HELPER_BINARY=/tmp/blora-host-exec-helper go test -race -v ./internal/containerterm
```

结果：**exit 0，5.980s**，真实测试 4.96s。测试仅向本次随机标签容器复制新 helper；共享的 `blora-isolated-e2e:20260909` 镜像未改动。

- node ResourceRef + 固定完整容器 ID/创建时间/启动时间绑定；可信主体经 `SpawnSpec.OwnerID` 传递。
- 同一 session 两个视图，只有持有输入租约者能写；真实 `stty size` 输出 `42 132`。
- detach/重新 attach 读到原归档，`host:once` 仅出现一次，没有重放输入。
- 两个独立 exec 同时存在；关闭第一个后，第二个真实响应 `host:still-alive`。
- 权限撤销后输入拒绝；维护关闭仍能按原 PID/birth/token 清理自己的 helper。关闭所有 exec 后目标容器仍 running。
- 使用变化的 StartedAt 无法启动 exec。
- 旧 helper 返回明确缺少 `processTree/subreaper` 能力；把 helper 目录覆盖为本测试 tmpfs 的容器返回固定路径不存在/exit 127，用户命令未获得启动许可。随后使用单独构建的 `/tmp/blora-host-exec-helper`（具备 processTree/subreaper）复制到本次隔离容器，运行 Master 双主体真实 API 测试通过：

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E=1 BLORA_HOST_EXEC_HELPER_BINARY=/tmp/blora-host-exec-helper BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 go test -race ./internal/master -run '^TestHostDockerTerminalAPIActorBindingAndRefresh$' -count=1 -timeout=180s
```

结果：`ok blora.dev/panel/internal/master 4.341s`。覆盖两个主体在同一容器的独立会话/授权、稳定任务重试、ANSI 与中文输入、断线后检查点恢复、撤权断流、关闭 PTY 不停止容器，以及旧 helper 能力不足时明确失败。

初次编译记录：HostAdapter 使用主代理已确认的 OwnerID 合同时，原 `terminal.SpawnSpec` 尚未添加该字段，编译 **exit 1**。主代理随后因远程 compact 网络错误不可运行，模块代理仅补齐已明确指定的两处：`SpawnSpec.OwnerID` 和 `Manager.Create` 构造 SpawnSpec 时从可信 `CreateRequest.OwnerID` 传值，没有改 Server/Daemon 行为。

### 最终相关模块回归

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E=1 BLORA_HOST_EXEC_HELPER_BINARY=/tmp/blora-host-exec-helper BLORA_TEST_DOCKER_ENDPOINT=unix:///var/run/docker.sock BLORA_TEST_DOCKER_IMAGE=blora-isolated-e2e:20260909 go test -race -v ./internal/containers ./internal/containerterm ./internal/containerterm/helper ./internal/runtime ./internal/terminal ./internal/dockerapi
```

整体 **exit 0**：containers 27.557s；containerterm 复用前述 5.980s 实测的测试缓存；helper 4.360s；runtime 4.464s；terminal 1.193s；dockerapi 无独立测试，其 TLS 经 containers 真握手覆盖。实际隔离实例 Docker PTY、隔离容器生命周期、宿主 Linux PTY 和相关故障测试仍通过。`TestDelegatedCgroupRealLaunch` 因未配置明确委派目录而 **SKIP**，不计 cgroup 真实验收通过。

最终追加修正：任务结果按实际资源身份更新/去重，避免 Compose 多个阶段把同一个容器重复计数。新增真实断言后 `BLORA_CONTAINER_E2E=1 go test -race -v ./internal/containers` **exit 0，27.114s**，生命周期 24.68s，包含完整 mTLS 拒绝分支（缺客户端证书、错误主机名、未知 CA、非法 CA 文件）。

`go vet ./internal/containers ./internal/containerterm/... ./internal/runtime ./internal/terminal ./internal/dockerapi` 与 `go build ./cmd/master ./cmd/daemon` 均 **exit 0**。Windows amd64 containers/containerterm 测试交叉编译均 **exit 0**，产物 `/tmp/blora-containers-windows.test.exe`、`/tmp/blora-containerterm-windows.test.exe`；构建和交编不是产品全链路或 Windows 真机运行证据。

### 主代理接管合同与下一步

1. Daemon 注册 backend `docker-host`，构造 `runtime.NewHostDockerExecutor(runtime.HostDockerOptions{StateRoot, Endpoint, TLSCertPath})`；关闭时先收敛 terminal Manager，再释放此 executor 的空闲连接。
2. 创建前调用 `containers.Manager.InspectExecTarget(ctx, actor, containers.Target)` 得到真实完整身份，再构造 `runtime.HostDockerTarget{ContainerID, CreatedAt, StartedAt}`。其 `Reference()` 为全部三字段 SHA-256，作为 terminal RunID。
3. 持久记录按 **actor + ":" + runRef** 绑定主体、node ResourceRef 和完整出生信息，不能仅以 runRef 存单个 actor；同容器多个管理员可以各自创建会话。
4. 注册 `containerterm.HostAdapter{Executor, Resolve}`，`Resolve` 签名为 `func(context.Context, actor string, model.ResourceRef, runRef string) (containerterm.HostBinding, error)`；返回 `HostBinding{Target, Authorize func(context.Context) error}`。Resolve 从受保护记录取主体/身份并复核，闭包每次重查当前节点 `host.manage`。
5. node 类型会话所有列表、流、输入、尺寸、关闭、续租检查都映射当前 `host.manage`；不要直接把 `terminal.input` 当成节点主机权限。既有托管实例继续实例终端入口和原 owner/run 检查。
6. 新建 `terminal.CreateRequest{OwnerID:可信主体, Resource:{Kind:"node",ID:nodeID,NodeID:nodeID}, RunID:runRef, Backend:"docker-host", ...}`。稳定后台任务与创建意图先记账，丢 ACK 查询原会话，不重发 exec。允许绝对容器目录，不能提供 host/user/privileged/helper 路径覆盖。
7. 补 Master/Daemon HTTP 与真实浏览器 Docker Center exec 入口，并连同现有 Compose 分片保存、WSS 日志窗口做端到端验证。当前已存在这些集成源码的部分实现，不凭本报告替代其验收。
8. 本次所有子进程/Go 执行会话均已结束，测试对象按自己的完整身份清理；仅 `/tmp` 的测试构建产物保留。根代理/桌面代理远程 compact 网络错误是续接原因，完整产品范围未缩减。
