# systemd与委派cgroup真实补验（2026-09-29）

起点 `a1478ef`。用户要求持续完成本地可完成工作。本轮发现原平台手册虽然提到systemd集成验证，但源码仅有命令参数/输出替身测试，不能支持真实生命周期完成结论。现补专属容器入口并实际运行，未修改UI、通透材质或性能阈值。

## 实际发现与修复

1. 原服务列表只调用`systemctl list-units --all`。真实服务停止并被manager卸载后会从列表消失。首轮专属服务启动/重启通过，停止后生产枚举断言失败，日志 `systemd-lifecycle.log`，退出1。
2. 修复服务后，真实未启动的timer在生产列表缺失。原实现只查询`list-timers --all`，无法枚举未加载但已安装的timer。失败日志 `systemd-service-fixed.log`，退出1。

当前服务和timer列表合并已安装与已加载单元，先按名称排序/去重/分页，再用一次`systemctl show`批量读取该页实际属性。不把enabled/disabled推断成运行状态；别名使用manager返回的Names/Id匹配；无实例的模板不作为可操作服务。timer返回真实启用状态，未加载timer保留计划属性。变更只涉及Linux枚举，API与存储schema不变。

## 隔离与运行入口

`scripts/systemd-e2e.py`及`systemd-e2e-init.sh`使用已有官方Playwright v1.63.0-noble镜像中的systemd。随机容器无网络、私有PID/mount/cgroup空间；只读挂载测试二进制及脚本，无宿主服务总线、cgroup、设备、仓库或Docker socket挂载。容器内需要SYS_ADMIN与独立seccomp配置，未采用privileged；仅改写容器私有子树。初始化建立明确的测试标记及专属slice。

测试程序必须包含指定用例，且必须实际PASS；SKIP和“没有测试”均拒绝。Go测试也检查专属标记、Docker环境和真实systemd PID 1。每次均清理自身unit/标记，再删除唯一所属容器；只有清理成功才打印PASSED。

```sh
go test -c -o /tmp/blora-systeminfo.test ./internal/systeminfo
go test -c -o /tmp/blora-runtime.test ./internal/runtime
python3 scripts/systemd-e2e.py --binary /tmp/blora-systeminfo.test --runtime-binary /tmp/blora-runtime.test
```

## 最终真实结果

日志位于私有目录`.local/evidence/rc26`，每次包装器记录实际UTC时间、单调耗时和退出码；不公开凭据。

| 验证 | 结果与证据 |
| --- | --- |
| 服务启动、重启、停止，真实PID切换/停止，规范名和别名一致可见 | `TestRealSystemdServiceAndTimerLifecycle`通过 |
| 已安装未启动timer可见，enable/disable/enable与列表状态一致，实际oneshot触发 | 同一真实systemd用例通过 |
| cgroup在exec前分配，`/proc/<pid>/cgroup`包含原runId，整组kill确认 | `TestDelegatedCgroupRealLaunch`通过 |
| 32MiB memory.max、8个pids.max、半核cpu.max真实写入，cpu.stat发生限流 | `TestDelegatedCgroupResourceLimits`通过；不冒充OOM/进程数耗尽测试 |
| 上述最终组合及所属资源清理 | `systemd-all-final.log`，3项实际执行、退出0，8.199秒 |
| 拒绝错误测试程序 | `systemd-wrong-binary.log`，1.553秒退出1，明确缺少目标测试，清理后结束 |
| runtime/systeminfo race回归 | `runtime-systeminfo-final-race.log`，退出0，5.140秒；默认真实环境用例跳过不重复计通过 |
| Master/Daemon系统管理、主机权限和防火墙相关回归 | `system-gates-race.log`，7个顶层用例通过，退出0，63.583秒 |

早期无额外能力的容器启动因`/run/lock`挂载失败退出255；隔离能力配置完善后真实manager启动成功。第一次限额补验的父层只委派memory/pids，写入cpu subtree control失败，日志`systemd-cgroup-limits.log`、`systemd-cgroup-delegated.log`、`systemd-controller-diagnosis.log`保留。读取实际controller后，专属slice显式配置CPUQuota，最终根与slice均返回cpu/memory/pids，CPU限流通过；未修改宿主controller。中间`systemd-cgroup-final.log`只证明启动/整组结束，不当作限额通过。

## 范围与后续

RC26交付已完成：`GOMAXPROCS=4 BLORA_VERSION=development-20260929-rc26 make package`退出0，81.235秒；六包外部SHA256、2,215条内部路径/模式/长度/摘要、三份Web与当前构建一致性及依赖许可证检查通过（4.694秒）。`SHA256SUMS`自身摘要为`1251c3fc22219eb96ce298a7417a3aaca4d752df88b5fa1cebacd0c780e5c5b1`。

`python3 scripts/package-smoke.py dist/releases/development-20260929-rc26 --state-restore --rollback-release dist/releases/development-20260929-rc25`退出0，10.320秒：独立SDK构建/签名、Master初始化/TLS/静态资源/登录、两个打包Daemon ONLINE、停机恢复后原TLS/节点身份/只读授权/任务回执/资源正文及兼容RC25回退全部通过。所属进程已停止。日志`package-build.log`、`package-verify.log`、`package-smoke.log`；产物`dist/releases/development-20260929-rc26`。没有重跑与本次后端修改无关的全UI截图或性能长测。

F13/E05的Linux systemd适配及E09/A03的上述委派cgroup子项已获得真实本机证据，不再笼统标为“完全没有可运行manager”。Windows、生产部署、其他内核/跨主机、OOM/进程数耗尽、物理掉电及系统通知中心仍分别验收。严格E08和Vim原偶发问题状态不变，不用本次平台通过代替其他场景。
