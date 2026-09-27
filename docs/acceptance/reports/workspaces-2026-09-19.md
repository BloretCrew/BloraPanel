# 多设备工作区冲突验证

关联 F02/A01。真实 HTTPS fixture `fixture-1902302221` 使用一个账号的两个独立浏览器上下文：第二上下文通过不同的 `blora:device` 身份提交 revision 2，第一上下文仍持有 revision 1。第一设备再次同步时服务端返回 `WORKSPACE_CONFLICT`，桌面显示“云端副本已被其他设备更新”，本地 Monaco 草稿“设备一的未同步现场”保持可见；随后读取云端确认 revision 2 和第二设备标记仍未被覆盖。第一场景还创建真实终端会话，确认默认布局同步剥离 `terminals`，显式工作内容同步携带终端检查点，同时保留草稿正文。

验证命令（2026-09-19，隔离 loopback HTTPS）：

```sh
BLORA_E2E_CREDENTIALS=<私有fixture凭据文件> \
  npm run test:e2e:real -- tests/real/workspaces.spec.ts --workers=1
```

结果：2/2 通过（6.9s，Playwright 会话 27729），包含终端检查点内容策略和多设备 stale revision 场景；fixture 进程随后以 Ctrl+C 停止（退出130），未保留测试服务。

该证据覆盖真实版本 CAS、冲突提示和本地现场保留，不替代浏览器存储耗尽、断电迁移或高延迟网络组合。
