# 主题切换与终端输出恢复顺序（2026-09-29）

起点`4aec17f`。实际复现：先收到自定义调色板并检查点，再收到原始日志中的第二种自定义颜色，用户切换浅色/深色主题后刷新，原有自定义色重新出现。正常边界与未完成CSI两项均失败（11.240秒，`theme-before.log`）。

## 修复

主题重置纳入与已解析远端输出相同的顺序队列，并将发生位置作为可选`colorResetSequence`恢复元数据保存。挂载也纳入队列，避免初始化覆盖随后选择的主题。恢复时先检查标记必须落在检查点基线与已保护序号之间，再在正确位置调用原生主题setter，后续远端颜色仍可生效。

重置是本地调色板操作，不将ESC字节插入半段CSI/OSC，不产生远端输入，也不改变原始事件序号、日志预算或用户选定的UI配色。下一次完整检查点会自然压缩掉旧标记。界面监听异步保护失败并显示错误，不能静默丢弃恢复保护失败。

首个修复复核仍失败（10.005秒），原因是把与初始化同一个主题对象引用再次赋值，xterm跳过了原生setter更新。恢复时传入新的等值主题对象后，原两项2/2通过（7.938秒，`theme-fixed-final.log`）。不改写中间失败记录。

## 当前证据

- 73/73单测通过，2.733秒（`unit.log`）；类型检查通过。
- Chromium原有状态/颜色/鼠标/UTF-8/Worker及四项主题顺序组合48/48通过，125.766秒（`terminal-regression.log`）。
- 完整边界/半段CSI/半段OSC × 后续有无新颜色六项通过，15.975秒（`theme-expanded.log`）。
- RC32生产构建和六包构建通过48.920秒；六份外部摘要、2,239条内部文件及三份Web一致性校验通过5.197秒。`SHA256SUMS`自身SHA256为`67c6862fe6cdeed577970512d9a11a02eb379786b6bd01a520fc1f5e85bb2d30`。
- 非法恢复标记拒绝且原记录保持、零输入1/1通过，6.362秒（`theme-invalid.log`）。
- 真实HTTPS双节点文件/PTY权限、刷新、移窗和Vim/top不重放7/7通过，76.300秒（`real-terminal.log`）。
- 独立发行SDK构建/签名、Master初始化/TLS/静态资源/登录、两个发行Daemon上线、停机恢复及兼容RC31回退通过，11.944秒（`package-smoke.log`）；身份、授权、回执和资源正文一致，快照后的修改不存在，所属进程清理完成。命令`python3 scripts/package-smoke.py dist/releases/development-20260929-rc32 --state-restore --rollback-release dist/releases/development-20260929-rc31`。
- Firefox最终组合40/40通过，107.223秒（`firefox-theme.log`）；WebKit40/40通过，108.485秒（`webkit-theme.log`）。

浏览器证据`web/.local/evidence/rc32`，发行和真实后端证据`.local/evidence/rc32`。链接恢复、原Vim偶发、严格E08及外部平台缺口仍保留。这里不把本次明确交错场景的修复推广为所有终端交互已验证。
