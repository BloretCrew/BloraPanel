# 真实 vim/top 终端恢复（2026-09-13）

关联 A06/F06，真实 Linux HTTPS Master、两个 Daemon 与 Chromium，fixture `.local/fixture-4180777667`。测试 `web/tests/real/files-terminal.spec.ts` 新增实际程序场景，无 API 或 PTY 替身。

命令：在 web 目录设置 `BLORA_E2E_CREDENTIALS` 指向上述 fixture 的私有 browser-credentials.json，运行 `npm run test:e2e:real -- --grep 'actual vim' --reporter=line`。最终退出 0，1/1 通过（9.0s）。

通过浏览器键盘在实际 PTY 启动 `vim -Nu NONE -n`，输入随机标识及中文未保存正文；刷新后保留相同 sessionId 和屏幕，WSS Data 输入未重发。执行 `:wq` 后通过真实文件 API 验证节点文件内容准确且没有重复。随后实际 `top -d 1`，刷新保留运行画面且不重发输入；按 q 退出并执行 shell 写文件，真实 API 验证退出后命令执行成功。最后显式结束会话，无 pageerror。

首轮失败保留：逐字符输入时 vim 会在字符之间输出光标/状态更新，原断言在未经终端解释的增量输出日志中寻找连续整行，超时。改为一次插入正文、刷新前等待中文输出，刷新后核对重建的完整屏幕，并用实际落盘正文判断缓冲区正确性。生产代码未因测试修改。不能用原始 ANSI 日志的字符串包含断言代替一般终端屏幕比较。

本轮覆盖 vim 未保存缓冲区和 top 刷新，不替代 Windows ConPTY、慢网长时输出和全部终端模式组合；仍按矩阵追踪。fixture 已收到 Ctrl+C，测试自建资源停止。

最终组合复验：新fixture `.local/fixture-79588922`，执行 `npm run test:e2e:real -- --grep 'actual vim|actual WSS Protobuf PTY' --reporter=line`（web目录及该fixture私有凭据）。5517退出0，2/2（26.7s）。除实际vim/top，还验证ANSI备用屏幕中文、重挂载URL的sequence精确等于刷新前持久检查点、sessionId未更换、WSS输入帧不增加且节点计数文件仅x；只读观察拒绝输入、接管后旧窗口失去租约、改字号导致真实尺寸更新、关视图不结束会话、结束确认刷新后固定原会话均通过。fixture已Ctrl+C退出，无活动测试。A06的Linux实际全屏/检查点/不重放组合据此更新，慢流量单独归A09，Windows运行归E09。
