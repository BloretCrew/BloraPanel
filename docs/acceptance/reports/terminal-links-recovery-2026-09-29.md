# 终端超链接检查点恢复（2026-09-29）

起点 `9ff9887`。实际复现 OSC 8 链接经过检查点和持久存储重新挂载后只剩文字，原生链接提供器不再返回目标；已关闭与仍打开的链接两项均失败（10.870 秒，`links-before.log`）。

## 修复范围

为锁定的 xterm 6 增加可选链接恢复元数据，保存主屏幕、备用屏幕、历史行的链接范围和当前输出属性。仅扫描稀疏扩展属性，连续范围合并；没有链接时不扫描缓冲区。完整验证版本、范围、顺序、属性和链接索引后，使用原生链接注册及行标记恢复，滚出历史区后仍由原生机制清理。相同状态同步恢复到序列化 Worker，旧检查点仍兼容。

不拼接 URI 为终端控制字节，不改变原生 HTTP/HTTPS 限制或点击处理，也不产生历史终端输入。适配依赖当前锁定版本的内部缓冲区接口，升级 xterm 时需重新验证。

## 已取得证据

- Chromium 相关恢复、解析器、鼠标、主题及 Worker 组合 **61/61**，148.680 秒（`terminal-regression.log`）。链接覆盖关闭/打开/尚未输出文字、宽字符折行、历史行、备用屏幕及返回主屏幕、尺寸重排、非 HTTP URI 拒绝。
- 真实鼠标悬停和点击在连续两次持久恢复后目标不变、零终端输入，通过 7.206 秒（`links-pointer-fixed.log`）。首轮检查错了指针类所在的元素，11.905 秒失败；按 xterm 源码改为检查 screen 元素，点击断言不变，失败日志保留。
- 单元测试 **76/76**，3.048 秒（`unit-final.log`），包含非法记录修改前拒绝、普通终端不保存链接状态，以及恢复后的原生链接标记随历史行释放。

- Firefox 相关组合 **50/50**，135.929 秒（`firefox-links.log`）。
- WebKit 同组 **50/50**，129.442 秒（`webkit-links.log`）。
- RC33 类型/生产构建及六包构建通过，55.934 秒（`package-build.log`）。六份外部摘要、2,243 条内部文件记录和三份 Web 一致性通过，5.536 秒（`package-verify.log`）。`SHA256SUMS` 自身 SHA256：`67d20a3c04291c938b9605117d7af7e80d1810a173a46850a59c138009f22354`。
- 独立发行 SDK 构建/签名、Master 初始化/TLS/静态资源/登录、两个发行 Daemon 上线、停机恢复及兼容 RC32 回退通过，9.713 秒（`package-smoke.log`）。身份、授权、回执和资源正文一致，快照后的修改不存在，所属进程清理完成。命令：`python3 scripts/package-smoke.py dist/releases/development-20260929-rc33 --state-restore --rollback-release dist/releases/development-20260929-rc32`。

- 真实 HTTPS 双节点文件/PTY 权限撤销、上传、刷新、移窗、Vim/top 和 WSS Protobuf **7/7**，74.004 秒（`real-terminal.log`）；所属 fixture 随命令退出清理，附件 `/tmp/blora-rc33-real`。

浏览器私有证据 `web/.local/evidence/rc33`；发行和其他引擎证据 `.local/evidence/rc33`。本轮所有运行会话已结束。UI 未改；原 Vim 偶发此次未出现，不据此声称根因已修复；严格 E08 性能和外部平台验证缺口仍保留。
