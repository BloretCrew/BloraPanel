# 活跃浏览器PTY撤权（2026-09-13）

关联A08。`files-terminal.spec.ts` 新增actual active browser PTY场景，实际HTTPS fixture `.local/fixture-1366325914`，管理员和成员分别使用独立浏览器上下文。只对自建测试账号/节点临时授予原生PTY所需terminal.read/input及host.manage，finally撤销新增授权并关闭原会话。

成员通过真实终端执行一次文件追加，管理员API读取文件确认为x。撤销terminal.input后，浏览器data-terminal-writable在15秒预算内变为false，继续键入追加命令未改变真实文件；再撤销terminal.read，成员新会话列表请求403。刷新原浏览器现场后，界面显示权限错误、输入仍禁用，文件仍仅x；旧sessionId快照没有恢复授权。

运行：web目录设置 `BLORA_E2E_CREDENTIALS` 为该fixture私有凭据路径，执行 `npm run test:e2e:real -- --grep 'actual active browser PTY' --reporter=line`。81846退出0，1/1（6.1s）。fixture已Ctrl+C停止，未修改生产代码或非测试账号。

这补齐活跃PTY与旧现场撤权的浏览器组合。文件写入撤权和账号切换另有accounts/desktop真实测试；资源查看撤权期间编辑器草稿完整保留等组合仍需逐项核对，A08不因此自动全部通过。

后续文件组合通过：新增 `actual file revocation`，成员创建真实节点文件并保存正文，追加未保存中文后管理员撤销file.write/file.read/instance.read。保存界面显示权限错误、服务器正文仍为原版本；刷新保留未保存正文及undo/redo，直接读取API403，实例总览排除该资源；实际下载导出的文本精确包含最新草稿。finally恢复本fixture原有授权。

在 `.local/fixture-708836550` 运行 `npm run test:e2e:real -- --grep 'actual file revocation' --reporter=line`，66268退出0，1/1（6.7s）。随后复验 `real HTTPS Monaco draft survives reload and explicit logout`，84359退出0，1/1（3.5s），账号退出及后续成员登录不保留前账号工作区。fixture已Ctrl+C停止，生产代码无修改。这些真实浏览器证据与前述PTY撤权共同覆盖A08主要组合；它们并不声称能远程擦除用户已经合法下载的数据。
