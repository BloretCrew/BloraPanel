<p align="center">
  <img src="docs/assets/blora-mark.svg" width="80" height="80" alt="Blora 标志">
</p>

<h1 align="center">Blora Panel</h1>

<p align="center">
  <strong>为你的服务器准备的桌面工作区。</strong><br>
  在一个桌面中管理节点、实例、文件、终端、容器与备份。
</p>

<p align="center">
  <a href="README.md">English</a> · <strong>简体中文</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&amp;logo=go&amp;logoColor=white" alt="Go 1.25">
  <img src="https://img.shields.io/badge/Vue-3-42B883?style=flat-square&amp;logo=vuedotjs&amp;logoColor=white" alt="Vue 3">
  <img src="https://img.shields.io/badge/TypeScript-6-3178C6?style=flat-square&amp;logo=typescript&amp;logoColor=white" alt="TypeScript 6">
  <img src="https://img.shields.io/badge/Status-In_development-64748B?style=flat-square" alt="开发中">
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> ·
  <a href="#生产环境启动">生产环境启动</a> ·
  <a href="#功能概览">功能概览</a> ·
  <a href="#文档导航">文档导航</a> ·
  <a href="sdk/README.md">扩展 SDK</a>
</p>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="web/ui-screenshots/all-appearance-20260926/mineral-dark-on-launcher.png">
    <img src="web/ui-screenshots/all-appearance-20260926/mineral-light-on-launcher.png" width="100%" alt="Blora Panel 桌面与应用启动器">
  </picture>
</p>

<p align="center">
  <sub>深浅色外观 · 六种主题配色 · 独立的通透材质开关</sub><br>
  <sub>截图使用当前中文界面与隔离演示数据。</sub>
</p>

<details>
<summary>查看深浅色截图</summary>

| 浅色 | 深色 |
| :---: | :---: |
| ![浅色桌面](web/ui-screenshots/all-appearance-20260926/mineral-light-on-launcher.png) | ![深色桌面](web/ui-screenshots/all-appearance-20260926/mineral-dark-on-launcher.png) |

</details>

## 功能概览

Blora Panel 采用受 Material Design 3 启发的桌面界面，后端使用 Go Master/Daemon 架构，管理操作独立于浏览器运行。

| 模块 | 主要能力 |
| :--- | :--- |
| **桌面工作区** | 多窗口、多标签、标签移窗、资源快捷入口，以及刷新后的工作现场恢复。 |
| **实例与节点** | 聚合跨节点获准资源，控制实例生命周期，管理配额、节点登记与身份换钥。 |
| **文件与编辑器** | 文件浏览、跨节点传输、Monaco 编辑，以及未保存正文和撤销历史恢复。 |
| **终端与监控** | PTY 终端、有界日志归档、实时指标与持久任务跟踪。 |
| **Docker 与 Compose** | 通过节点授权操作管理容器和 Compose 项目，保留任务与恢复记录。 |
| **备份与计划** | 创建备份、审阅恢复计划、配置调度，以及执行管理员预设的应用钩子。 |
| **用户与扩展** | 资源权限、角色模板、应用安装升级，以及独立 SDK 扩展开发。 |

关闭窗口不会停止实例，打开资源快捷方式不会自动启动实例。通透材质、主题配色与深浅色模式分别设置。

## 快速开始

**已验证工具链：** Linux amd64、Go **1.25.9**、Node.js **24.15.0**、npm **11.12.1**。构建使用 GNU Make，发行打包另需 Python **3.9+**。依赖版本锁定于 `go.mod`、`go.sum` 与各 npm 锁文件。

```sh
git clone https://github.com/BloretCrew/BloraPanel.git
cd BloraPanel

make build
make web
make sdk
make fixture
```

| 命令 | 结果 |
| :--- | :--- |
| `make build` | 构建 `dist/blora-master` 与 `dist/blora-daemon`。 |
| `make web` | 安装锁定的前端依赖并生成 `web/dist`。 |
| `make sdk` | 构建 SDK 与三种参考扩展包。 |
| `make fixture` | 启动独立 Master 和两个 Daemon，准备隔离测试账号与资源。 |

打开测试环境输出的网址，默认为 **`https://127.0.0.1:9443`**。工具生成本机测试证书及私有 `.local/fixture-*` 目录；随机管理员与成员凭据只写入权限 `0600` 的 `browser-credentials.json`，没有可复用默认密码。按 **Ctrl+C** 停止本测试环境所属的资源与服务。

## 生产环境启动

Master 默认只监听 **`127.0.0.1:8443`**，不要求绑定域名或填写固定 `--origin`。初始化生成的证书覆盖 `localhost`、`127.0.0.1` 和 `::1`。可完全本地运行，需要外部入口时再交给 Nginx 反代。

**版本范围：** 自动识别本地来源及 `--trusted-proxies` 是已发布 `v0.1.0-beta.1` 之后的修改。以下步骤使用当前源码（`make build && make web`），将 `dist/blora-master`、`dist/blora-daemon` 和 `web/dist` 放到对应目录。旧 beta 可显式加 `--origin https://localhost:8443` 本地启动，但没有新的可信反代来源/IP 识别功能；原 beta 标签和附件不会覆盖。

程序和状态分开：Master 放 `/opt/blora/master`（二进制及 `web/dist`），节点 Daemon 放 `/opt/blora/daemon`。先准备运行账号 `blora-master` / `blora-daemon`，私有状态/配置目录归对应账号所有（目录 `0700`，敏感文件 `0600`）。Windows 使用对应路径和私有 ACL。不要使用 fixture 测试身份或状态。

### 1. Master：初始化一次，然后本地常驻

用编辑器创建 `/etc/blora/initial-password`：12～72 字节，仅 Master 账号可读。先初始化一次，再运行不带 `--init` 的常驻进程：

```sh
sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master --init \
  --password-file /etc/blora/initial-password

sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master --listen 127.0.0.1:8443 \
  --static-dir /opt/blora/master/web/dist \
  --extensions-dir /var/lib/blora/master/extensions
```

本机打开 **`https://localhost:8443`** 或 **`https://127.0.0.1:8443`**。核对生成的 `/var/lib/blora/master/tls.crt` 后，在浏览器/系统中信任它；不要关闭证书校验或分发 `tls.key`。使用 `admin` 和自己的初始密码登录，初始化成功后删除密码文件，重启保留状态目录。Master 已带前端。`--extensions-dir` 开启扩展安装持久目录，可选注册源、公钥参数见 `blora-master --help`。

直接访问时，自动模式只接受本地回环主机名/IP，并按请求的实际 HTTPS 主机和端口校验 Origin。若希望限制为固定入口，仍可显式填写 `--origin`。

### 2. Daemon：登记后连接 Master

在节点应用生成一次性登记票据，私有保存为 `/etc/blora/node.enrollment`，将 **Master 的公开 `tls.crt`** 复制为 `/etc/blora/master-ca.pem`。与 Master 同机的 Daemon 配置 `/etc/blora/node.json`：

```json
{
  "stateDir": "/var/lib/blora/node",
  "masterUrl": "https://127.0.0.1:8443",
  "caFile": "/etc/blora/master-ca.pem",
  "enrollmentFile": "/etc/blora/node.enrollment",
  "allowPGIDFallback": false
}
```

```sh
sudo -u blora-daemon /opt/blora/daemon/blora-daemon --config /etc/blora/node.json
```

每个节点独立使用状态目录和身份。上线后删除票据文件及 `enrollmentFile` 字段，后续复用已登记身份。另一台机器的 Daemon 使用可访问的 Nginx HTTPS 入口作为 `masterUrl`，私有证书配置该入口 CA，系统受信证书可省略 `caFile`。`127.0.0.1` 始终指 Daemon 自己所在机器。Daemon 主动连接，不需要开放入站管理端口；管理 Master 所在机器，也要运行一个 Daemon。

Linux 原生实例需要管理员准备可写、已委派的 `cgroupRoot`；Docker/Compose 需要可访问的 `dockerEndpoint` 和运行依赖。账号权限按节点能力设置。`allowPGIDFallback` 是可信工作负载的显式后备方式，不是普通用户隔离。详见[运维手册](docs/operations/OPERATIONS.md)。

### 3. 前端 / Nginx：可选的外部入口

上面的 Master 已向浏览器提供前端，**无需单独下载 Web 包，也无需 Node.js、开发服务器或预览服务器进程**。

Nginx 与 Master 同机时，Master 仍只监听回环地址，显式信任代理连接后端使用的 IP：

```sh
sudo -u blora-master /opt/blora/master/blora-master \
  --state-dir /var/lib/blora/master --listen 127.0.0.1:8443 \
  --trusted-proxies 127.0.0.1/32 \
  --static-dir /opt/blora/master/web/dist \
  --extensions-dir /var/lib/blora/master/extensions
```

[Nginx 示例](docs/operations/examples/nginx.conf)独立提供静态文件，将 API/WSS 和健康检查转发到本地 Master。静态根目录默认 `/opt/blora/master/web/dist`；若使用同版本独立 Web 包，解压到 `/opt/blora/web` 并修改 `root`。若希望页面也由 Master 提供，使用相同代理设置配置 `location /`，替换静态 location 即可。

将 Master 公开证书复制为 `/etc/blora/backend-ca.pem`，使用 `proxy_ssl_name localhost` 校验后端。**Nginx 外部入口**的证书按访问它的 IP 或主机名配置，Master 无需绑定或提前知道该域名。示例放入 Nginx 的 `http {}` 上下文，检查后启动：

```sh
sudo nginx -t
sudo systemctl enable --now nginx
# 已运行时，在配置检查通过后 reload。
```

Nginx 覆盖 `X-Forwarded-Host`、`X-Forwarded-Proto`、`X-Forwarded-For` 和 `X-Real-IP`。Master 仅在直连来源属于显式配置的可信 IP/CIDR 时读取这些头，其他来源的代理头忽略。原始 HTTPS 主机/端口用于 API 和全部浏览器 WebSocket 来源校验；真实客户端 IP 用于登录限流。IP 链从靠近 Master 的一侧向客户端检查，遇到首个不可信节点停止；单层示例直接用 `$remote_addr` 覆盖 XFF，避免客户端伪造前缀。多层反代需配置各可信节点，并在入口清洗原始主机/协议头。歧义、无效的可信代理头和非 HTTPS 转发来源会被拒绝。

前端与 API 保持同一 HTTPS 来源：前端使用相对 `/api/v1` 和同源 Cookie。后端仍必须使用 TLS，代理头不能绕过 TLS 要求。头部转发依据 [Nginx 官方文档](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header)。开放 Nginx 管理入口不等于开放被管理应用的业务端口。

<details>
<summary><strong>Linux 后台服务 / Windows 启动方式</strong></summary>

准备账号/目录并初始化 Master 后，使用 [Master unit](docs/operations/examples/blora-master.service)和 [Daemon unit](docs/operations/examples/blora-daemon.service)，分别放到所在机器的 `/etc/systemd/system/`。启用本机反代时，给 Master 的 `ExecStart` 加 `--trusted-proxies 127.0.0.1/32`。不要同时用同一状态目录运行前台副本和服务。

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now blora-master.service  # Master 机器
sudo systemctl enable --now blora-daemon.service  # 配置好的节点
sudo journalctl -u blora-master.service -f        # 或 blora-daemon.service
```

Daemon 的 `KillMode=process` 保留独立归属的实例/日志辅助进程，停止服务不是停止所有资源；一致备份/升级前显式停止实例、PTY 并等待任务结束。示例不自动准备 cgroup 或 Docker。

Windows 准备私有状态/密码/配置目录，当前编译程序及前端放在 `C:\Blora\master` / `C:\Blora\daemon`：

```powershell
# 仅初始化一次：
& 'C:\Blora\master\blora-master.exe' --state-dir 'C:\Blora\state\master' --init --password-file 'C:\Blora\private\initial-password'
# 本地常驻：
& 'C:\Blora\master\blora-master.exe' --state-dir 'C:\Blora\state\master' --listen '127.0.0.1:8443' --static-dir 'C:\Blora\master\web\dist' --extensions-dir 'C:\Blora\state\master\extensions'
# 节点，node.json 使用 Windows 路径：
& 'C:\Blora\daemon\blora-daemon.exe' --config 'C:\Blora\private\node.json'
```

JSON 路径例如 `C:/Blora/state/node`，同样复制并信任公开 Master CA。Windows 是控制台程序，不能用 `sc create` 直接安装为 SCM 服务；外部进程管理器需保留私有状态和独立辅助进程。本次源码修改只有交叉编译，不冒充新的真机认证。

</details>

本机可用 `curl --fail --cacert /etc/blora/master-ca.pem https://127.0.0.1:8443/healthz` 校验健康，再确认节点状态、权限和实际操作。`/healthz` 只检查 Master 数据库。文档不自动变更生产机器。

## 技术架构

| 组件 | 职责 |
| :--- | :--- |
| **前端 / 桌面** · Vue 3 / TypeScript | 在浏览器中运行，负责窗口、标签、默认应用、扩展宿主、交互和本地工作现场恢复；以静态文件交付，由 Master 或 Web 服务器托管。 |
| **Master** · Go | 中央管理服务，负责账号/会话、服务端授权、节点登记、资源索引、持久任务协调及 API/WSS 网关；通常一套部署运行一个，也可直接托管前端。 |
| **Daemon** · Go | 每台被管理机器上的执行服务，负责本机进程生命周期、PTY、文件、Docker/Compose、备份、监控及任务执行和回报；管理哪台机器，就需要在那台机器运行。 |

节点主动向 Master 建立管理连接。HTTPS 与 WSS 只承载管理数据，被管理服务的业务网络独立配置。资源授权由服务端落实，第三方扩展默认不获得宿主机命令执行权限。

日常管理链路是 **浏览器前端 → Master → 目标机器的 Daemon**。SDK 是扩展开发工具，不是第四个常驻服务。关闭浏览器不会停止服务端实例、计划和任务；节点失联按离线/结果待确认呈现，不冒充实例已停止。

## 开发与验证

```sh
make check
make test
npm --prefix web test
npm --prefix web run test:e2e
```

浏览器测试默认使用 `/usr/bin/chromium-browser`，可通过 `BLORA_CHROMIUM` 指定其他 Chromium 可执行文件。race 测试需要受支持的本机平台及 C 工具链。更多配置、真实浏览器测试、故障验收与发行步骤见下面的完整运行指南。

<details>
<summary><strong>展开完整运行指南：节点登记、容器、备份、扩展、发行与故障验收</strong></summary>

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

</details>

## 项目状态

**Blora Panel 正在持续开发。** 实现状态与验证状态分别记录在[验收矩阵](docs/acceptance/ACCEPTANCE_MATRIX.md)中。

首个公开版本为 **[v0.1.0-beta.1](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1)**，是用于初次实际体验的预发布版本。已有自动化及代理执行的验证，项目负责人尚未亲自试用，不宣称达到正式稳定版标准。[beta 验收报告](docs/acceptance/reports/beta-0.1.0-2026-10-08.md)记录包与源码一致性、解包后的实际启动/恢复、兼容升级回退及远端附件核对结果。

[非发行源码收尾](docs/acceptance/reports/non-release-closeout-2026-10-08.md)包含后端竞态、构建、123 项前端单测、完整运行与后续单项复验合并覆盖的 51 个真实功能场景、恢复修复和严格原生绘制校验。原失败与合并证据的限制仍保留；beta 包制作与验证已随后完成，正式稳定版等待本人实际体验。

[Windows 普通真机验证](docs/acceptance/reports/windows-device-passed-2026-10-06.md)的十四项已全部通过；此前 Job/ConPTY 与可见 WebKit 结果保留各自固定源码范围，新 beta Windows 包尚未在真机重跑。[当前剩余工作](docs/execution/NON_UI_REMAINING_2026-09-26.md)区分已交付 beta 与等待体验的正式版。额外物理故障和部署组合按本次范围免测，未执行不记为通过。

[E08 已按用户调整范围收尾](docs/acceptance/reports/e08-bounded-closeout-2026-10-08.md)：Chromium 为 38.8ms、Firefox 为 66ms、WebKit 为 89～132ms，最后一次为 95ms。所有浏览器统一 50ms 的原目标仍未通过；这些固定源码的压力结果不保证每轮都低于 100ms。

真实 systemd 服务/定时器生命周期、委派 cgroup 进程管理、CPU 限流及有界内存/PID耗尽已有[隔离 Linux 容器证据](docs/acceptance/reports/systemd-cgroup-2026-09-29.md)；[Linux原生通知绘制与鼠标点击](docs/acceptance/reports/native-notifications-2026-09-29.md)也已在 X11/Dunst 验证。各报告只覆盖其注明环境，可按[平台指南](docs/operations/PLATFORM_VALIDATION.md)复现；本地回归不认证生产部署和其他平台组合。

仓库包含源码、测试、依赖锁文件、文档及两张 README 展示截图。凭据、运行数据库、已安装依赖、发行归档、生成的参考包与其他历史截图图库不进入版本控制。

## 文档导航

| 文档 | 内容 |
| :--- | :--- |
| [运维手册](docs/operations/OPERATIONS.md) | 状态布局、健康检查、备份、升级与回退。 |
| [平台验收](docs/operations/PLATFORM_VALIDATION.md) | Windows、systemd、远端 Engine 与故障验证流程。 |
| [本地回归](docs/operations/LOCAL_REGRESSION.md) | 当前源码构建、浏览器守卫、真实功能自动回归与独立实验入口。 |
| [API 说明](docs/api/openapi.yaml) | 请求、授权、幂等、任务、WebSocket 与结构化错误。 |
| [扩展 SDK](sdk/README.md) | 应用清单、沙箱能力、签名与生命周期合同。 |
| [验收矩阵](docs/acceptance/ACCEPTANCE_MATRIX.md) | 已实现、已验证、待验证及环境相关要求。 |
| [执行进度](docs/execution/PROGRESS.md) | 最新工程决定、证据与剩余工作。 |
| [技术架构](docs/plan/Blora-01-技术架构.md) | Master/Daemon 架构与核心技术合同。 |
| [功能与交互](docs/plan/Blora-02-功能与交互.md) | 产品范围、应用与工作区行为。 |

## 第三方许可

Blora Panel 采用 **GNU General Public License 第 3 版（仅限第 3 版）**，SPDX 标识为 `GPL-3.0-only`，全文见 [LICENSE](LICENSE)。第三方组件保留各自许可证；原始文本及通知见[依赖清单](docs/licenses/inventory.json)和[第三方通知](docs/licenses/THIRD-PARTY-NOTICES.txt)，随发行包提供。
