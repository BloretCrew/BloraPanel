# 受控高延迟/丢包扩展任务验收（2026-09-22）

## 环境

私有 HTTPS Master、双 Daemon、官方 Playwright WebKit 共享任务专用 Docker network/PID namespace。该 namespace 的 loopback 通过 `tc/netem` 实际配置为 80ms 单向延迟、10ms 抖动、1% 丢包；没有触及宿主网络或外部注册源。

## 结果

`extensions.spec.ts --grep 'independent package installs'` **1/1 通过（49.4 秒）**，`.last-run.json` 为 `{"status":"passed","failedTests":[]}`。

场景从桌面安装独立签名包，打开节点资源的第二个扩展窗口并提交实际 WASI 节点任务。服务端接受第一次 POST 后，测试有意中止浏览器收到的回执；刷新后扩展保留笔记和稳定请求键，再次提交返回相同 task 身份，任务最终成功。随后核对资源身份、结果字符/词数、桌面快捷方式、多窗口移动、刷新恢复和关闭视图行为。

测试此前将 Chromium 的网络错误文本 `fetch` 误当为合同；现改为验证 `Error:` 状态，随后仍要求同一请求键恢复。WebKit 对同样的故意断开显示 `Error: Load failed`，这不是产品成功或静默丢失。

## 边界与清理

这是本机隔离 namespace 的受控网络证据，不是公网扩展注册源、跨主机远端网络或 Windows 验收。测试容器、网络规则与私有状态目录均已删除。
