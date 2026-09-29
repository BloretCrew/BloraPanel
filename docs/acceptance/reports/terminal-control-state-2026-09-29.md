# 终端持久控制状态恢复（2026-09-29）

起点`93ccace`。沿用RC28真实终端连续输出与检查点重建对照，新增四个已完成控制指令场景：字符集、滚动区域、保存光标和自定义制表位。四项在修复前全部失败（19.783秒）；例如图形字符集的`qq`恢复成普通字母，滚动区域失效后保留了本应被区域滚动移走的行。原失败日志在`web/.local/evidence/rc29/terminal-state-before.log`。首个未升级权限的启动因本机服务监听受限退出，未计产品失败或测试通过。

## 实现

锁定的xterm 6 SerializeAddon没有输出这些控制状态。新增独立适配器，在屏幕检查点旁保存版本化的`controlState`：两种缓冲区的光标/滚动区域/制表位/保存光标位置与属性，以及G0～G3字符集、当前字符集和切换层级。保存位置按缓冲区历史基线计算相对值，恢复时使用实际重建的历史基线。

恢复先完整校验数值、字符集映射、版本和两个缓冲区，再只赋值明确字段；保留原生属性对象，不批量合并不可信对象。屏幕回放后、UTF-8续接和原始日志之前恢复控制状态，并同步实际headless Worker。旧检查点没有该字段时仍沿用原读取路径。内部接口与xterm版本耦合，升级须执行真实行为对照。

UI、通透材质、原始日志预算和严格性能阈值未修改；终端回放仍禁止向远端发送历史输入。

## 当前证据

- 初始修复后的控制序列与四项新增场景14/14通过，31.975秒。
- 增加保存光标颜色、G1切换、备用屏幕返回、历史滚动、原点模式后，连同UTF-8/Worker回归30/30通过，88.501秒。
- 前端71/71单测通过，1.999秒；包含破坏第二缓冲区时禁止部分修改、非法映射与版本拒绝、JSON往返后真实字符集行为。首次新增单测未开启headless公开buffer的proposed API而失败，补与产品一致的配置后通过，旧日志保留。
- 类型检查及生产构建通过；从新的RecoveryService实际读取持久数据后19/19通过，55.934秒（`terminal-persistent.log`）。
- 真实HTTPS双节点文件/PTY刷新、移窗、权限撤销和Vim/top不重放7/7通过，80.621秒（`real-terminal.log`）。
- Firefox正式配置19/19通过，50.720秒（`firefox-terminal.log`）；WebKit首轮18/19通过，备用屏幕项在动态模块导入阶段报`Importing a module script failed`，未执行到终端断言（53.053秒，`webkit-terminal.log`）。保持用例和断言不变全组复核19/19通过，51.653秒（`webkit-terminal-repeat.log`）。首轮失败保留，模块加载问题未再次出现，不声称定位其根因。

## 发行验证

RC29六包构建62.202秒通过；独立校验六份外部摘要、2,227条内部文件记录及三份Web与当前构建一致性通过，5.519秒。`dist/releases/development-20260929-rc29/SHA256SUMS`自身SHA256为`891c0a45c2fcd9d08ffcf5f0a031cf66d4ecb1b5831b31518f3cc2f3da4b1176`。

独立SDK构建/签名、Master初始化/TLS/静态资源/登录、两个发行Daemon上线、停机恢复与兼容RC28回退通过，10.955秒。恢复核对身份、授权、回执与资源正文，确认快照后的改动不存在；所属进程全部停止。命令`python3 scripts/package-smoke.py dist/releases/development-20260929-rc29 --state-restore --rollback-release dist/releases/development-20260929-rc28`。对应日志`package-build.log`、`package-verify.log`、`package-smoke.log`及各自退出0的`.result.json`均保留。

浏览器与单测私有证据在`web/.local/evidence/rc29`，真实后端/发行证据在`.local/evidence/rc29`。对照覆盖上述明确状态，不声称所有VT扩展和自定义颜色/链接状态均已覆盖，也未把本次缺陷等同于原Vim偶发额外字节根因。E08与外部平台缺口保留。
