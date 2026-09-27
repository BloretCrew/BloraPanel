# 受控高延迟/丢包扩展注册源验收（2026-09-22）

在任务专用 Docker network/PID namespace 中，私有 HTTPS fixture 与官方 WebKit 经过 `tc/netem delay 80ms 10ms distribution normal loss 1%` 通信。`extensions.spec.ts --grep 'real local catalog browses two versions'` **1/1 通过（13.6 秒）**：本地注册源显示两个版本，安装 v1.0.0 后升级至 v1.1.0，禁用时 bundle 请求明确拒绝，重新启用后沙箱 iframe 成功执行 v1.1.0 bundle。

容器、netem qdisc 与私有状态目录均已删除。该结果是受控单机 namespace 证据，不替代公网注册源、跨主机网络、真实证书运维流程或 Windows 验收。
