# B05 容器终端与 B04 能力探测验证

日期：2026-09-09。范围是 runtime Docker exec、containerterm Adapter、隔离镜像内的 exec-helper；不代表全产品验收完成。既有 native PTY 证据见 terminal-2026-09-09.md。

## 实现与权限边界

- `runtime.Manager.StartDockerExec(ctx, Record, DockerExecOptions)` 返回真实 `io.Reader/io.Writer/Resize/Wait/Close`；`InspectDockerExec` 校验 exec 对应容器、运行出生时间及 Blora instance/run/ownership 标签。
- `containerterm.Adapter` 要求实例资源、节点及 run ID，经可信 Resolver 取得 Record；普通请求不能传任意容器 ID、Docker 用户、特权、挂载或 socket。实际容器还须非 root 数字 uid:gid、只读根、私有 PID namespace、cap-drop ALL、no-new-privileges。
- 固定 `/usr/local/lib/blora/exec-helper` 必须通过真实版本与 pidfd 功能探测。缺 helper 或能力不符返回 CAPABILITY_UNAVAILABLE，不启动 Shell。默认镜像 `Dockerfile.isolated` 包含静态 Go helper；`make -f cmd/exec-helper/Makefile` 可单独构建。
- exec 使用一次 Docker HTTP Upgrade 连接；服务端持久化 exec ID 与 helper PID/出生身份后才发送一次启动确认。浏览器连接和输入租约由 terminal.Manager 管理；Daemon 重启失去 PTY 时标失效，不再启动替代 Shell。
- Close 启动固定的容器内 helper 关闭指令，先核实 PID 出生时间、会话身份和秘密所有权标记，再用 pidfd 发送信号。终端主管确认自己会话中的进程消失，最后检查 Engine 原 exec 已退出。不会停止或删除整个实例容器。清理不确定返回错误并保留 close_unknown。
- 用户命令自身非零退出以终端正文报告；helper 的退出状态表示清理是否确认。只有清理确认才是成功关闭。
- `Manager.Capabilities(ctx)` 实际查询 Docker version/info 和 native 运行能力。Linux 回退明确标识 pgid_trusted、缺少硬限制；不把无委派 cgroup 说成完整隔离。能力探测只创建并回收自身标记的短暂 native 运行。

## 验证命令及结果

统一 Go 环境：`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false`。

1. `go test -race -v ./internal/runtime ./internal/containerterm/... ./cmd/exec-helper`：runtime 20 项通过，3 项因未提供真实 Docker/委派 cgroup 环境跳过；Adapter 通过。HTTP 替身覆盖真实 Upgrade、辅助程序探测、所有权拒绝、尺寸与关闭错误。首次 sandbox 无权创建 httptest socket，批准后原命令通过。
2. `go test -race -v ./internal/containerterm/helper`：真实 Linux PTY、中文/ANSI、出生身份和 token 不匹配拒绝、清理自身后台进程且保留另一控制进程、命令 exit 7 与清理状态分离、真实 pidfd probe 通过，3.058 秒。
3. `GOOS=windows GOARCH=amd64 go test -c -o /tmp/blora-runtime-exec-windows.test.exe ./internal/runtime`：当前源码交叉编译通过。Docker helper 是 Linux 容器程序；Windows 主机运行和 ConPTY 仍需 Windows 真机，不能以交编替代。

实际 Engine：Docker 29.4.1，API 1.54，最小 API 1.40，Linux。客户端使用受支持的 1.45。首次真实启动暴露 `/version` 实际字段为 `Os`，原 `OSType` 解码为空而误报能力不可用；已修复客户端和测试夹具，无放宽 Linux 检查。

主代理创建专用镜像 `blora-isolated-e2e:20260909`，记录 manifest `sha256:6486725488fcea62b3d2a2da0a395434a359a7642b536098a6cacfa04b0526c9`。下面命令通过自动权限审查后执行：

```sh
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false BLORA_TEST_DOCKER_ENDPOINT=unix:///var/run/docker.sock BLORA_TEST_DOCKER_IMAGE=blora-isolated-e2e:20260909 BLORA_TEST_DOCKER_INSTANCE_ROOT=/data/instances/blora-panel/.local go test -race -v ./internal/runtime -run '^TestDockerExecRealIsolatedSession$' -count=1
```

通过（测试 0.88 秒，包 1.897 秒）：中文和 ANSI 原始输出、输入只出现一次、Resize 后容器内 `stty size` 实际返回 42×132、同容器两条 exec 会话、关闭第一条后第二条继续交互且主实例仍运行。仅创建和清理随机测试对象；run `f6b1acd4e4fe8041e756fe81a89764b3`、container `f8b5dac435a1c40b6ea1d707c6e97a3de3562631c3d5cd8d627669e5ef7dd5f5` 清理后精确 GET 返回 404。未触碰既有容器，未删除主代理提供的测试镜像。

```sh
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false BLORA_TEST_DOCKER_ENDPOINT=unix:///var/run/docker.sock go test -race -v ./internal/runtime -run '^TestDockerCapabilitiesRealEngine$' -count=1
```

通过（1.034 秒），实际探测得到 isolated runtime API 1.45 可用。

## 验证边界

没有 Windows 真机结果。未验证远程 HTTPS Engine；代码沿用已有证书验证客户端，不假设远端能访问本机 helper 路径。当前默认路线为镜像内预装 helper，未实现管理员额外 helper 挂载。终端关闭范围是本次 PTY 会话；主动建立新 session 的进程仍受实例容器隔离，不能宣称 PTY 本身是敌对程序沙箱。Daemon/主服务的权限、WSS 与 UI 集成由主代理另行记录。
