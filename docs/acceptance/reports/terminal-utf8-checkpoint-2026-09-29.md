# 终端UTF-8检查点续接修复与RC27（2026-09-29）

起点`d8c5fd8`。继续审计本地可完成的恢复边界时，发现已有测试只覆盖持续运行中的分块UTF-8，没有覆盖半个字符到达后保存检查点并重建终端。

## 原失败与修复

真实xterm/Worker/RecoveryService收到“中”的前两个字节`e4 b8`后生成检查点，销毁并重新挂载模型，再收到`ad`。期望屏幕为“中”，实际为空串；首轮退出1，6.304秒，原始日志`web/.local/evidence/rc26/terminal-split-refresh.log`保留。

xterm的屏幕序列化不包括流式UTF-8解码器的未完成字节。当前实现只在二进制输出写入后跟踪末尾不完整码点，最多保留3字节；检查点增加可选`pendingUtf8`，在恢复屏幕后、回放原始输出日志前重新送入解码器。旧检查点不含此字段仍可读取；长度、编码和不完整前缀不符合要求时保留原记录并报错。该状态不包含键盘输入，不发送远端副作用，也不改变输出序号或ACK保护顺序。

尾部检查只读取当前块末尾最多4字节及上一段最多3字节，不复制整个输出块。屏幕镜像Worker收到相同续接数据，Worker失效时继续使用主线程序列化路径。

## 验证

| 验证 | 实际结果 |
| --- | --- |
| 初始修复与既有真实xterm/Worker场景 | 4/4通过，日志`.local/evidence/rc26/terminal-split-fixed.log` |
| 二/三/四字节字符全部6个分割位置、回放查询静默/live应答、Worker屏幕一致、实际杀死Worker后回退 | 9/9通过，50.851秒，`.local/evidence/rc26/terminal-utf8-boundaries.log` |
| 检查点保留第1字节、原始输出日志保留第2字节，恢复后收到最后2字节 | 1/1通过，7.023秒，`.local/evidence/rc27/terminal-utf8-journal.log` |
| 前端单测（含逐字节续接、完整码点、不合法中断边界） | 68/68通过，2.223秒，`unit.log` |
| 实际HTTPS双节点文件/PTY：撤权、刷新、移窗、同会话、Vim未保存中文/top | 7/7通过，75.135秒，`real-terminal.log` |
| 类型检查与生产构建 | 通过；初始Uint8Array泛型类型不匹配已修正，未忽略类型错误 |

除特别注明外，日志在`.local/evidence/rc27`，包装器另记录UTC起止时间、单调耗时和退出码。用户选定UI、材质、图标、配色和性能阈值均未改变。

## RC27发行

`GOMAXPROCS=4 BLORA_VERSION=development-20260929-rc27 make package`通过，49.255秒；六包外部摘要、2,219条内部路径/权限/长度/摘要、三份Web与当前构建及依赖许可证校验通过，5.051秒。`SHA256SUMS`自身SHA256为`53c411078125a9cb2161451e1d8e698bc8892a943d35233a9d0452b4799e46d1`。

独立发行验证：

```sh
python3 scripts/package-smoke.py dist/releases/development-20260929-rc27 --state-restore --rollback-release dist/releases/development-20260929-rc26
```

退出0，10.024秒，包含独立SDK构建签名、TLS Master/双Daemon启动、停机恢复与兼容RC26回退；原身份、授权、任务回执和资源正文保留，所属进程清理成功。日志`package-build.log`、`package-verify.log`、`package-smoke.log`；产物`dist/releases/development-20260929-rc27`。

## 未关闭的边界

此修复有确定的UTF-8失败复现，不是此前Vim额外`$y`字节的根因证明。后者仍未定位，现有原始字节附件与严格输入不重放断言保留。RC26完整51项回归是前一版本证据，本轮只将7项真实文件/PTY和上述定向结果计入RC27，不冒称重跑了全套。

UTF-8通过不能推及未完成VT控制序列的检查点状态；下一步应独立审计此边界。严格E08仍未达标，Windows、跨主机及物理故障仍有环境缺口，整体验收保持未完成。
