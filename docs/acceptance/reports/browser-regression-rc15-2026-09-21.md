# RC15 当前源码真实浏览器回归

日期：2026-09-21。恢复值 `copy()` 优化已进入 RC15 当前源码后，使用全新 `--performance` fixture、真实 HTTPS Master/双 Daemon 和本机 Docker Engine 串行运行普通真实浏览器套件。独立一小时性能测试被排除；该场景另有[完整小时报告](performance-hour-recovery-copy-2026-09-21.md)。

夹具由以下命令启动：

```sh
./dist/blora-devfixture --performance \
  --docker-endpoint=unix:///var/run/docker.sock \
  --listen 127.0.0.1:9444
```

运行命令：

```sh
BLORA_E2E_CREDENTIALS=<private fixture credentials file> \
  npm run test:e2e:real -- --workers=1 --trace=off --reporter=line \
  --grep-invert='real eight-window mixed load measures pointer response with two PTYs and verified transfer'
```

Playwright 实际收集 **49 项**，使用 1 个 worker 串行运行，**49/49 passed**，退出码 0，耗时 11.5 分钟。真实 HTTPS fixture 在私有目录 `.local/fixture-2757846034` 创建；浏览器仅访问该 fixture 和 task-owned Docker 对象。

覆盖账号/资源授权、备份及实际分钟调度、Docker/Compose、桌面多窗和标签恢复、实例控制、扩展 SDK/WASI/权限边界、文件冲突/上传续传、PTY 输入与检查点、通知、旧快照/存储故障恢复、云工作区冲突和幂等实例创建。关闭浏览器续传场景确认原 `uploadId` 从 1,179,648B 已确认偏移继续上传 16MiB 文件，10,000 项解压继续完成且 `browserTabId` 改变；两个实际分钟槽位均生成唯一任务/归档。

Docker/Compose 测试的清理逻辑移除了其任务创建的项目/卷；随后 fixture 通过自己的监督进程 Ctrl+C 退出 0，Master/Daemon 和端口检查无残留，私有 fixture 目录已删除。

本结果覆盖 RC15 当前源码的一次全套常规真实浏览器回归。它不替代一小时 E08、Windows 真机、独立 systemd、远程 Docker、高 RTT/丢包或物理掉电验收。
