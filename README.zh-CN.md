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

Master 和 Daemon 均可**不带参数直接运行**。分别默认读取**可执行文件同目录**的 `master.json` 和 `daemon.json`，不依赖当前工作目录。配置中的相对文件/目录路径也按可执行文件所在目录解析；未填写时，数据目录默认为 `state/master` / `state/daemon`，Master 前端默认为 `web/dist`。需要另选配置文件时仍可用 `--config`。

Master 默认是 **`http://127.0.0.1:37861`**。可在 `master.json` 中把 `listen` 改为 `0.0.0.0:37861` 监听所有 IPv4 网卡；HTTP/WS 也可用于远程访问，HTTPS/WSS 则可由 Nginx 终止 TLS，常规启动无需准备 Master/Daemon 本机证书。

**版本范围：** [v0.1.0-beta.4](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.4) 支持直接远程 HTTP，并包含非安全 HTTP 页面登录所需的浏览器 UUID 回退。beta.2 和 beta.3 尚未同时包含这两项修复。

### 1. Master

二进制与 `web/dist` 放在 `/opt/blora/master`。将[生产配置示例](docs/operations/examples/master.json)保存为 **`/opt/blora/master/master.json`**：

```json
{
  "stateDir": "/var/lib/blora/master",
  "listen": "127.0.0.1:37861",
  "staticDir": "/opt/blora/master/web/dist",
  "extensionsDir": "/var/lib/blora/master/extensions",
  "adminName": "admin",
  "passwordFile": "/etc/blora/initial-password",
  "verifyProxyIPs": false
}
```

用本地编辑器创建配置指定的 `passwordFile`（12～72 字节，仅 Master 账号可读）。首次启动自动初始化管理员并提供面板，之后启动复用原数据库：

```sh
sudo -u blora-master /opt/blora/master/blora-master
```

也可在程序目录直接运行 `./blora-master`。本机打开 **`http://127.0.0.1:37861`**；登录后删除初始密码文件。配置/密码文件私有权限 `0600`，状态目录 `0700`，归运行账号所有。[便携模板](master.example.json)使用相对路径，需要允许该账号创建对应数据目录。

### 2. Daemon

每台被管理机器部署 Daemon，程序放在 `/opt/blora/daemon`。在面板生成一次性登记票据并私有保存；将 [daemon.json](docs/operations/examples/daemon.json) 放到可执行文件同目录：

```json
{
  "stateDir": "/var/lib/blora/node",
  "masterUrl": "http://127.0.0.1:37861",
  "enrollmentFile": "/etc/blora/node.enrollment",
  "allowPGIDFallback": false
}
```

```sh
sudo -u blora-daemon /opt/blora/daemon/blora-daemon
```

也可在程序目录直接运行 `./blora-daemon`。[便携模板](daemon.example.json)默认连接本机 Master，票据和状态使用相对路径。登记后复用节点自己的身份，可删除票据及 `enrollmentFile`；每个 Daemon 独立使用状态目录。

另一台机器的 Daemon 将 `masterUrl` 改为可访问的 Master 地址。优先使用 **Nginx HTTPS 地址**；也可使用 HTTP。Nginx 使用系统受信证书时不用填写 `caFile`；私有 CA 可按需配置。Daemon 主动连接，不需开放入站管理端口；管理 Master 所在主机也需要在该机运行 Daemon。

Linux 原生实例需要管理员准备可写、已委派的 `cgroupRoot`；Docker/Compose 需要 `dockerEndpoint` 和运行依赖。`allowPGIDFallback` 仅用于显式信任的工作负载，不是普通用户隔离。

### 3. 前端 / Nginx

Master 已提供 `web/dist`，**不需要额外 Node.js 或 Web 服务进程**。外部访问使用 [Nginx 示例](docs/operations/examples/nginx.conf)：证书和 HTTPS 由 Nginx 处理，API/WebSocket 转发至 **`http://127.0.0.1:37861`**，前端与 API 保持同源。若要独立提供 Web 包，修改 Nginx 静态 `root` 即可。

默认 `verifyProxyIPs` 为 **false**，读取转发头时不要求代理 IP 列表。请求头不能证明一定来自 Nginx，所以保持后端只监听本机，并像示例一样由 Nginx 覆盖 `X-Forwarded-Host/Proto/For` 和 `X-Real-IP`。原始主机/协议用于浏览器来源校验和 Cookie 安全属性，客户端 IP 用于登录限流；错误或歧义头会拒绝。

需要限制可信代理 IP 时，在 `master.json` 修改这两个字段，再用相同命令重启：

```json
{
  "verifyProxyIPs": true,
  "trustedProxies": ["127.0.0.1/32", "::1/128"]
}
```

开关开启时必须填写非空 IP/CIDR 列表，其他连接来源的代理头被忽略。多层反代需要清洗转发链；单层示例直接覆盖 XFF 为真实客户端地址，不拼接未经信任的输入。

<details>
<summary><strong>可选 Master 配置</strong></summary>

| 配置项 | 用途 |
| :--- | :--- |
| `stateDir`、`listen`、`staticDir` | 持久数据、监听地址、构建好的前端。 |
| `adminName`、`passwordFile` | 首次启动的管理员及私有密码文件路径。 |
| `verifyProxyIPs`、`trustedProxies` | 可选的严格代理 IP 校验。 |
| `origin` | 可选固定浏览器来源，通常自动识别。 |
| `extensionsDir` | 持久扩展注册目录，省略时不开启安装。 |
| `extensionsCatalog`、`extensionsCatalogUrl` | 可选本地 / HTTPS 扩展注册源。 |
| `extensionsPublicKey` | 32 字节十六进制 Ed25519 公钥，启用包强制签名校验。 |
| `https`、`tlsCert`、`tlsKey` | 可选直接 HTTPS 兼容模式，Nginx 终止 TLS 时通常不用填写。 |

配置修改后重启生效；未知 JSON 字段或无效设置在打开数据库前拒绝。`--init` 仍可用于只初始化后退出，拒绝覆盖已有管理员；旧 fixture 的参数启动方式继续兼容。

</details>

<details>
<summary><strong>Linux 服务 / Windows 启动</strong></summary>

使用 [Master unit](docs/operations/examples/blora-master.service)和 [Daemon unit](docs/operations/examples/blora-daemon.service)。准备对应账号、同目录 JSON 和私有状态目录后安装：

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now blora-master.service
sudo systemctl enable --now blora-daemon.service
```

Daemon 的 `KillMode=process` 保留独立归属的实例/日志辅助进程；在线更新只切换管理子进程。一致性状态备份或不兼容维护升级仍按停机流程排空。

Windows 将 `master.json` 放在 `blora-master.exe` 同目录，`daemon.json` 放在 `blora-daemon.exe` 同目录。绝对路径改为 `C:/Blora/state/master` 等 Windows 路径，或使用便携模板的相对路径：

```powershell
& 'C:\Blora\master\blora-master.exe'
& 'C:\Blora\daemon\blora-daemon.exe'
```

程序为控制台程序，不能直接当作 SCM 原生服务安装；外部守护需保留私有状态及独立辅助进程。当前 Windows 构建为交叉编译，未新增这次修改的真机认证。

</details>

本机健康检查用 `curl --fail http://127.0.0.1:37861/healthz`，并通过配置的外部 Master 地址检查。健康接口仅检查 Master 数据库，不覆盖全部节点能力。文档没有自动修改生产机器。

### 在线更新

管理员在**设置**更新 Master，在**节点管理**更新对应 Daemon。下载已构建好的 Release，运行机器不需要 Git、Go、Node.js 或现场编译。先检查版本和兼容性，再应用固定目标。运行中的托管实例保留原进程身份；管理连接会重连，交互终端可能断开，浏览器上传可能需要续传。

两份配置文件均可填写相同的可选来源：

```json
{
  "updates": {
    "repository": "https://github.com/BloretCrew/BloraPanel",
    "channel": "beta"
  }
}
```

不填写时默认本项目、`beta` 通道；`stable` 排除预发布。可选 `updates.apiUrl` 指向 HTTPS **GitHub 兼容 API 镜像**，需同时提供 Release、附件下载、标签和提交接口，普通文件目录不适用。来源供应可执行程序，需要选择可信项目/镜像。在面板保存的来源持久化于私有更新状态目录，优先于 JSON 中的初始默认值。

校验 `SHA256SUMS`、`CORE-UPDATE.json`、归档路径和逐文件 `MANIFEST.json` 后，再读取二进制内嵌兼容信息。稳定启动器先准备独立版本目录，原管理子进程退出后才切换；配置、身份、数据库和实例目录保留原位置。候选启动失败保留旧版本。默认捆绑的前端随 Master 更新；自定义独立前端目录由管理员维护。

版本号相近不代表兼容：数据库迁移指纹不同、对端协议不适配或无法确认实例归属时拒绝在线更新。不兼容架构/存储变更需要维护流程，不自动降级数据库或执行补丁链。缺少新元数据的旧包（包括 `v0.1.0-beta.1`）不能作为在线更新目标；首次人工安装 beta.2，后续兼容 Release 可通过网页更新。新增 Windows 在线更新行为尚需真机验证。

生产建议独立安装到 `/data/instances/blora-panel-runtime` 等目录，不把真实运行状态混入当前开发工作树。详见[更新与生产目录合同](docs/plan/Blora-03-在线更新与生产部署.md)。

## 技术架构

| 组件 | 职责 |
| :--- | :--- |
| **前端 / 桌面** · Vue 3 / TypeScript | 在浏览器中运行，负责窗口、标签、默认应用、扩展宿主、交互和本地工作现场恢复；以静态文件交付，由 Master 或 Web 服务器托管。 |
| **Master** · Go | 中央管理服务，负责账号/会话、服务端授权、节点登记、资源索引、持久任务协调及 API/WSS 网关；通常一套部署运行一个，也可直接托管前端。 |
| **Daemon** · Go | 每台被管理机器上的执行服务，负责本机进程生命周期、PTY、文件、Docker/Compose、备份、监控及任务执行和回报；管理哪台机器，就需要在那台机器运行。 |

节点主动向 Master 建立管理连接。Master 支持 HTTPS/WSS 和 HTTP/WS 管理访问；被管理服务的业务网络独立配置。资源授权由服务端落实，第三方扩展默认不获得宿主机命令执行权限。

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

将 [master.example.json](master.example.json) 复制为编译程序同目录的 `dist/master.json`。源码构建布局将 `staticDir` 设为 `../web/dist`，解包后的 Master 已有同目录 `web/dist`。再用本地编辑器创建配置指定的 `dist/initial-password` 文件（权限 0600，12～72 字节密码），不把密码写入命令行或仓库。准备完后直接启动：

```sh
./dist/blora-master
```

首次启动按私有密码文件初始化管理员，后续启动复用数据库。Master 默认提供本机 HTTP，外部 HTTPS 和证书交给 Nginx。本任务未公开部署。

登录后在节点应用生成一次性登记票据，保存为 Daemon 所在机器的私有文本文件。将独立配置保存为可执行文件同目录的 `daemon.json`，例如：

```json
{
  "stateDir": "/absolute/private/blora-node-a",
  "masterUrl": "http://127.0.0.1:37861",
  "enrollmentFile": "/absolute/private/node-a.enrollment",
  "allowPGIDFallback": false
}
```

```sh
./dist/blora-daemon
```

每个 Daemon 使用独立状态目录、票据和身份。本机 HTTP 不需要 CA 文件；远程使用 Nginx HTTPS，私有 CA 才需要配置 caFile。已有身份不重新消费登记票据。

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

钩子不会经过 shell，资源身份和 hook ID 通过 `BLORA_BACKUP_*` 环境变量传入；未配置的策略会明确返回能力不可用。扩展目录在 Master JSON 中可用 `extensionsCatalog` 指定本地目录，或用 `extensionsCatalogUrl` 指定 HTTPS 注册源 拉取 `index.json` 与版本包；两者不能同时配置，远程源必须是 HTTPS 且拒绝重定向。

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

当前预构建预发布版本为 **[v0.1.0-beta.4](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.4)**，包含零参数启动、核心 Release 更新、直接远程 HTTP，以及非安全 HTTP 页面登录所需的浏览器 UUID 回退。自动化及代理执行的验证不能代替项目负责人的亲自体验，不宣称达到正式稳定版标准。

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
