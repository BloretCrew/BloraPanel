# 终端长时消费与错误关闭

关联F06/A06/A09/E08。默认速率小时场景在9.3分钟失败，600秒复现最终也退出1。提高输出速率（`BLORA_PERF_OUTPUT_DELAY_SECONDS=0.001`）后，真实TLS双节点/Chromium在1.7分钟捕获`OUTPUT_GAP`；两路PTY仍运行，节点有界归档继续轮转。逐条xterm解析和恢复写入导致消费者积压，输出被归档保留策略淘汰后正确停止拼接。浏览器错误路径还错误使用`WebSocket.close(1002)`，引发第二个异常。

修改：

- `web/src/apps/TerminalApp.vue`：有界队列合并相邻DATA帧，单批最多64KiB载荷，控制帧不跨越；每帧验证身份与字节顺序，解析和保护完成后发累计ACK。原消费者总预算保持不变。
- `web/src/services/terminals.ts`：保留原事件序号，按顺序处理resize和相邻输出，最多64KiB原始字节合并为一次xterm写入；只保护已解析边界，不重放输入。同步恢复日志仍保留每个原始事件序号，超预算生成完整检查点。
- TerminalApp/InstanceConsole错误关闭改用浏览器允许的应用码4002。
- 性能入口每轮检查终端错误，失败仅导出会话身份、状态、错误及pageerror，不保存账号字段/终端正文。输出间隔可调，默认0.01秒不变。

验证：`npm run build`通过；`npm run test:e2e -- tests/browser/terminal.spec.ts --workers=1`，2/2（17.0s）通过。真实xterm、IndexedDB及两种渲染模式，传输为测试替身：突发64个事件夹带resize、最后原序号90刷新恢复、累计ACK精确字节、错误显示及禁用输入、无浏览器异常；原中文、备用屏幕、移窗和不重放输入回归保留。

真实高输出180秒复验退出1（3.1分钟）：首分钟两路接收18.8/19.5MB，较修复前约6.3/6.4MB提升约三倍；128秒仍各约19MB/分钟，之后仍出现OUTPUT_GAP，无pageerror。这组0.001秒输出的超载不能计通过，也不承诺任意输出速率均可无损消费；原默认0.01秒输出负载正在重新进行小时验证。

断流后仍显示“已连接”的文本也已改为“连接已停止 · 原检查点保留”，日志控制台相应显示停止并保留已接收日志。最终前端构建通过，加入状态提示断言的终端浏览器2/2（18.3s）通过。运行句柄见PROGRESS。默认速率小时、Windows及远程网络场景仍未完成。
# 最新终态

默认输出速率小时测试完成全程，双PTY持续输出、无OUTPUT_GAP/浏览器异常，180次传输轮换及末轮目标校验完成；整体仍失败于交互p95 50.3ms超过50ms，见[小时报告](performance-continuous-2026-09-19.md)，不能把终端稳定子项写成整体性能通过。

最新`npm run test:e2e -- tests/browser/terminal.spec.ts --workers=1`实际2/2（18.0s）通过，包含此前未运行的双存储故障：sessionStorage尾日志和IndexedDB快照均注入QuotaExceededError，20个新增事件后明确显示未持久保护，未发送新增ACK，未出现pageerror。传输仍为测试替身，xterm解析及浏览器恢复存储是真实实现。
