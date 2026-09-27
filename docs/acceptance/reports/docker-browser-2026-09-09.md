# Docker 应用浏览器边界验证

命令：

```sh
npm run test:e2e -- --workers=1 --trace=off --grep 'Docker app queries' --reporter=line
```

结果：退出 0，1/1 通过（5.4 秒）。Playwright API fixture 返回完整节点、容器查询和任务回执；场景验证 Docker 应用选择管理节点、展示容器、打开显式确认对话框、提交 `container.start` 并带自动生成的 `Idempotency-Key`，随后读取任务完成状态。

该场景覆盖浏览器到 Master API 边界，不替代真实 Docker daemon 生命周期报告；真实容器/Compose/日志/卷保留证据见 `containers-e2e-2026-09-09.md`。
