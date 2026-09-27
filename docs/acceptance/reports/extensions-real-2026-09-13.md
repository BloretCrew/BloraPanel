# 独立扩展全链路真实浏览器复验

2026-09-13，在 Linux Chromium、HTTPS `127.0.0.1:9444`、真实 Master 与双 Daemon fixture 上执行：

```sh
BLORA_E2E_CREDENTIALS=/data/instances/blora-panel/.local/fixture-2953159556/browser-credentials.json \
npm run test:e2e:real -- --workers=1 --trace=off --reporter=line extensions.spec.ts
```

退出码 0，5/5 通过，耗时 24.2s：

- 独立包通知保留来源、刷新不重复提交，并可从通知回到扩展窗口；
- 扩展元数据草稿在刷新后保留，恢复失败时不提交，重试后更新真实实例；
- 服务器用户数据与视图现场完成 v1→v2 迁移，拒绝不兼容 v3，失败后保留数据并可回滚；
- 独立包从桌面安装，在真实节点运行 WASI 后台任务，结果、资源窗口、快捷入口、移窗和关闭视图均保持；
- 本地注册源浏览 v1.0.0/v1.1.0，安装升级、禁用/启用门禁和 bundle 版本核验通过。

fixture 测试对象随后已停止。该证据覆盖 E06/E07/A13 的真实独立包成功链路；Windows 真机、远程高延迟注册源和物理掉电/ENOSPC 仍按矩阵保留未验证项，不能据此宣称对应整项完成。
