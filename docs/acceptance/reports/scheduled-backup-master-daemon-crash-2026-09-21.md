# 定时备份期间 Master 与 Daemon 强制终止（2026-09-21）

关联 E04、F12。使用 `development-20260921-rc15` 的独立发行包运行 Master 和两个 Daemon，建立随机临时资源及一份 32 MiB 本地源文件，创建真实 `* * * * *` UTC `backup.create` 计划。测试等待生产调度器实际接受 occurrence，并读取已持久化的 `lastTaskId` 与 task/request 回执；随后对 Master 和目标 Daemon 同时调用 `SIGKILL`，不走应用关闭流程。新进程用原状态目录恢复，另一 Daemon 在同一 Master 下继续连接。

恢复后目标 Daemon 的 `startupId` 改变，计划的 `lastTaskId` 未变，原 taskId/requestId 保持不变，接受的备份任务最终为 `SUCCEEDED`，该实例只产生一份归档。计划随后被禁用以避免测试跨入下一分钟触发；所有服务由测试 `finally` 停止。私有诊断目录留在 `/tmp/blora-release-smoke-or6m55rp`。

复现命令：

```sh
python3 scripts/package-smoke.py dist/releases/development-20260921-rc15 --scheduled-backup-crash
```

结果：退出 0；关键输出为 `Master and owning Daemon both SIGKILLed`、`same task/request receipt recovered terminal as SUCCEEDED`、`schedule did not create another occurrence`、`archives=1`。任务 ID 为 `39e263f2f3172944f793eb5fa21e4ef1`。测试脚本入口为 [package-smoke.py](../../../scripts/package-smoke.py)。首次脚本试跑发现等待函数作用域和更新计划 HTTP 方法错误，均已修正后才取得上述通过结果；这些是测试脚本错误，不是产品成功证据。

这证明实际发行包在已接受定时任务后的 Master/Daemon 进程级强制终止与恢复，不模拟掉电、文件系统损坏或 SQLite 事务提交中断；长时墙钟、物理掉电、Windows/systemd 和远端环境仍未验证。
