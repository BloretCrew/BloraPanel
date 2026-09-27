# 真实浏览器备份、恢复与调度验收

记录日期：2026-09-14。环境为 Linux amd64、真实 Chromium、HTTPS Master、两个真实 Daemon 和隔离临时目录；fixture `fixture-4134305623` 已在测试后停止并清理。测试不访问生产资源。

命令：

```sh
BLORA_E2E_CREDENTIALS=/tmp/blora-backup-real/fixture-4134305623/browser-credentials.json \
  npm run test:e2e:real -- backups.spec.ts --workers=1 --trace=off
```

结果：1/1 通过，3.5s，退出码 0。场景先用真实文件 API 写入带中文的临时文件并核对版本，再通过备份任务生成 `ready` 快照；随后在默认实例中心打开“备份与计划”应用，读取真实快照、指定唯一目标路径、生成不可变恢复计划并提交覆盖任务，最后核对目标文件正文完全一致。相同窗口继续创建 `*/5 * * * *`、`Asia/Shanghai` 的备份计划，确认真实保存回执并按修订号删除。

该证据覆盖浏览器入口、任务状态与实际文件结果；物理空间耗尽/掉电、具体应用 save/pause/stop 钩子和长期离线调度仍按 E03/E04 保留未验证。
