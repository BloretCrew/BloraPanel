# 终端连接代次隔离与当前版本回归

起点 `3f0f9eb`（产品 `dfb0c14`）。本轮按用户要求继续本地任务，补 RC33 完整真实功能回归，并针对源码中的异步连接边界建立可控复现。

## 已确认的产品缺陷

旧连接处理输出期间重新挂载，新连接已经完成恢复并可输入，随后旧解析操作失败。旧 `fail()` 未检查连接代次，覆盖共享的状态、错误与输入权限，导致新连接失去输入能力。隔离浏览器测试真实使用 Vue 终端应用、xterm、恢复存储和 Protobuf；仅替换网络端与延迟解析失败，首次在预期的“新连接仍可用”断言失败，25.664秒（`generation-before.log`）。

新增代次/销毁状态检查，覆盖输入回调、异步恢复完成、错误和消息入口。旧连接不得改写当前连接；当前连接发生错误仍保留检查点并停止输入。首个复现原断言修复后通过，12.250秒（`generation-fixed.log`）。这证明了一个确定竞态，尚不能证明它就是此前 Vim 额外字节或权限撤销单次超时的根因。

## 当前完整回归

RC33完整真实51项已结束，**49通过、2失败**，879.929秒；日志 `.local/evidence/rc33/real-full.log`，附件 `/tmp/blora-rc33-real-full`。扩展通知发布与实例名称保存两处点击后无预期结果；首轮失败保留。对应测试增加既有 `clickPaintedExtensionButton` 的原生鼠标hover命中就绪检查，然后仍只执行一次正常点击，原业务断言/权限/失败保护不变。最终复验见下文，不归因为已确定产品缺陷。Vim与权限撤销此次通过，但未定位此前偶发根因。

新增连接回归最终还覆盖当前连接错误仍禁用输入；与原终端流控/恢复三种配置组合 **4/4通过**，46.248秒（`terminal-browser.log`）。Firefox同组4/4、66.584秒，WebKit4/4、64.482秒。私有浏览器日志 `web/.local/evidence/rc34`，跨引擎日志 `.local/evidence/rc34`。UI及E08标准未变。

## RC34发行验证

- 类型/生产构建和六包构建通过55.948秒，`package-build.log`。
- 六份外部摘要、2,251条内部记录及三份Web一致性通过5.485秒，`package-verify.log`；`SHA256SUMS`自身SHA256为`389e1079ad879364d0c85e57d3f8b97f8646b76eb4e1f1bef0b2df63bb17f140`。
- 独立SDK构建/签名、Master初始化/TLS/静态资源/登录、两发行Daemon上线、停机恢复及兼容RC33回退通过10.655秒，`package-smoke.log`；所属进程清理完成。命令`python3 scripts/package-smoke.py dist/releases/development-20260929-rc34 --state-restore --rollback-release dist/releases/development-20260929-rc33`。

修复后RC34完整真实功能回归 **51/51通过，无skip，897.040秒，退出0**。日志`.local/evidence/rc34/real-full.log`及同名`.result.json`，附件`/tmp/blora-rc34-real-full`；2026-09-29 14:13:19～14:28:16 UTC。两处扩展首轮失败在最终套件均通过；仍不把测试就绪前置条件视为确定产品根因。包括真实HTTPS双节点、Docker/Compose、WASI扩展、文件/PTY撤权及刷新移窗、121点历史/Master重启、14:26和14:27 UTC两个实际分钟备份与恢复、关闭浏览器后10,000项解压及16MiB原上传续传、云工作区冲突和幂等创建。启动脚本退出时停止所属fixture。

命令：`bash /tmp/blora-rc25-real.sh --grep-invert 'real eight-window mixed load' --workers=1 --output=/tmp/blora-rc34-real-full`，外层使用`/tmp/blora-record-rc25.py`记录真实时长/退出码。脚本启用实际Docker端点与`BLORA_E01_HISTORY_SOAK=1`、采样间隔1000ms，执行`playwright.real.config.ts`；版本名沿用的启动脚本不改变RC34当前构建。此51项是功能回归，明确不包含独立E08负载门禁，不能覆盖已有延迟失败。

## 独立渲染诊断（未采用）

尝试仅裁掉被不透明应用正文完全覆盖的窗口玻璃区域，仍保留原模糊参数。使用实际Vue应用开八个窗口，固定动画/时间，先连续截两张原版，再截图裁剪候选。两张原版像素完全一致；候选35,344像素变化，其中455像素最大通道差超过8，平均绝对通道差0.026229。候选未通过严格画面一致性前置检查，未进入产品，也未将理论减少的滤镜面积当作E08收益。诊断日志`web/.local/evidence/rc34/opaque-control.log`，退出0仅指诊断正常运行；不代表优化通过。
