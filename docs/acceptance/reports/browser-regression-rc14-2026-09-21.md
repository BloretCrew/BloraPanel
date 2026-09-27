# RC14 当前源码真实浏览器回归（2026-09-21）

## 环境与命令

使用全新私有 fixture `.local/fixture-1611252974`，真实 HTTPS Master、双 Daemon、性能测试用 10,000 项目录及 16 MiB 传输源；Daemon 通过本机 Docker socket 连接已安装的隔离测试镜像。fixture 由以下命令启动：

```sh
./dist/blora-devfixture --performance --docker-endpoint=unix:///var/run/docker.sock
```

随后串行执行当前源码普通真实浏览器套件，排除单独运行的一小时性能场景：

```sh
BLORA_E2E_CREDENTIALS=<private fixture credentials file> \
  npm run test:e2e:real -- --workers=1 --trace=off --reporter=line \
  --grep-invert='real eight-window mixed load measures pointer response with two PTYs and verified transfer'
```

Playwright 报告 `Running 49 tests using 1 worker`，最终 **49/49 passed**，退出码 0，耗时 11.4 分钟。

## 覆盖与清理

本轮通过了账号与权限、备份和真实分钟调度、Docker/Compose、桌面窗口和实例控制、刷新/撤权恢复、扩展 SDK/WASI/注册源、文件传输、双 PTY、日志、节点失联恢复、通知、云工作区和实际实例创建等 49 个场景。Docker/Compose 测试的 `finally` 清理了本测试随机创建的项目和卷；随后 fixture 收到 Ctrl+C 并以退出码 0 停止，特权只读进程检查未发现 fixture Master/Daemon 或 Playwright 残留。测试不包含单独的一小时 E08 性能场景；该场景当前源码 p95 超标仍见[E08 小时报告](performance-hour-rc14-2026-09-21.md)。

这次结果将常规真实浏览器套件从“各场景分别有通过证据”提升为 RC14 当前源码的一次完整 49/49 通过；它不覆盖 Windows 真机、独立 systemd、远程 Engine/高延迟、物理掉电或 E08 小时门槛。
