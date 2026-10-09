# 双程序零参数启动与本机 HTTP 反代

日期：2026-10-09。用户最新要求替代上一轮“后端必用 HTTPS、必须显式信任代理 IP”的默认值。这是 beta.1 之后的源码更新，不覆盖原标签或附件。

## 实现

- Master 默认 `http://127.0.0.1:37861`；常规启动不生成或要求后端证书。外部 HTTPS/WSS 与证书由 Nginx 处理。
- Master 默认读可执行文件同目录 `master.json`，Daemon 默认读同目录 `daemon.json`。相对配置文件名、配置内路径与默认数据路径都按实际可执行文件目录解析，独立于调用者 cwd。状态默认 `state/master` / `state/daemon`，静态文件默认 `web/dist`，密码文件默认 `initial-password`。
- 两程序不带参数即可启动。Master 首次普通启动按私有密码文件初始化账号并常驻，后续不重建账号、不要求保留密码文件。`--init` 仍可只初始化后退出并拒绝覆盖账号。
- 默认 `verifyProxyIPs:false` 信任转发头，无须 IP 列表；开启后必须配置有效非空 `trustedProxies`。默认模式不能识别 Nginx 身份，因此保持回环监听，入口覆盖客户端传入的头；严格模式保留逐跳与不可信头忽略。
- HTTP 只允许实际回环 peer，不能用伪造 HTTPS 转发头放行远程明文。Daemon 也校验 HTTP 主机及实际拨号地址为回环，远程仍经 Nginx HTTPS。控制/数据连接均支持相应 WS/WSS。
- API、事件、终端、日志、容器日志及节点换钥采用统一传输/来源判断。本机 HTTP Cookie 不设 Secure；原始 HTTPS 即使后端 HTTP 也保持 Secure。本机 HTTP 页面的 CSP 放行 WS。
- 配置限制大小、拒绝未知字段、错误类型、非对象和多重 JSON；Master 校验监听端口、TLS 配对、公钥和严格代理列表。旧参数入口兼容 TLS fixture，不隐式继承用户同目录配置。
- 中英 README、两份 systemd unit、Nginx HTTP 后端、Vite 默认目标、四份配置模板和未来包 START 同步。源码布局的 `dist/master.json` 用 `staticDir: "../web/dist"`，独立发行布局用同目录 `web/dist`。

## 验证

```sh
env GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test -race ./cmd/master ./cmd/daemon ./internal/bootstrap ./internal/master ./internal/daemon -count=1
```

受影响后端全包竞态退出 0：Master 命令包 8.506s、Daemon 命令包 1.042s、Master 服务包 284.657s、Daemon 服务包 32.054s。bootstrap 无独立测试，由命令包及真实启动覆盖。之后的空代理 IP/帮助小改动由最终命令包竞态覆盖（7.077s / 1.029s），不将两阶段称为同一源码全包重跑。

守卫包括同目录默认配置、外置配置的相对路径基准、省略字段、旧参数、错误配置/空 IP、可选 HTTPS 初始化/重启、实际 HTTP 后端 + HTTPS 反代登录与 Secure Cookie、事件 WSS/跨来源拒绝、严格/全信任代理、非回环明文拒绝，以及真实 Daemon HTTP 登记、签名控制 WS 和 Bulk 数据 WS。

实际 Linux 二进制验收：将两份当前 `dist` 程序复制到自有私有临时目录，从第三个无关 cwd 以零启动参数运行。省略 stateDir、Master 端口、Daemon masterUrl 和 Master passwordFile，核对同目录状态、首次自动初始化、HTTP 页面/会话及没有生成 TLS 文件。删除密码/票据后重启，节点 ONLINE 且 nodeId 不变；通过 JSON 切换默认全信任及严格允许/拒绝 IP，符合 200/403。整链路退出 0，所属进程正常关闭、私有目录清理。

手动验证纠正保留：早期误用 `/api/v1/auth/login` 得到 404，改用真实 `/api/v1/login`；另一轮将原始 HTTPS 的 Secure Cookie 用于 HTTP 得到正确的 401，随后恢复独立本机会话并按登记接口真实 201 重跑。这些是验收客户端/路由错误，不作为产品修复计数，最终整链路通过。

`make check build windows` 通过；前端 `npm run build`（含 TypeScript）通过，10.15s，原有大 chunk 提示保留。111 个 README/运维本地链接、10 个 JSON 文档块、四份 JSON 模板、Python 包脚本语法及 diff 检查通过。

本地开发包 `.local/config-startup-packages/zero-args-startup-final/` 不是公开发行。`python3 scripts/package.py --version zero-args-startup-final --output .local/config-startup-packages` 退出 0；六包共 2458 个清单载荷、外部/内部摘要、四份实际二进制、两模板、四份零参数 START 及两份 Master 内完整前端核验通过。包标记修改工作树，不冒充已发布版本；本条最终证据更新晚于该本地打包，不将其当作已提交新版本附件。

## 边界

本机没有 Nginx，使用真实 HTTPS Go 反代复现覆盖头及 HTTP 后端，不记为 Nginx 部署 PASS。没有安装/启动 systemd、改防火墙、公开监听或触碰生产资源。Windows 当前代码交叉编译通过，未新增真机 PASS；没有重开 UI/E08 或要求用户重跑已有设备验收。
