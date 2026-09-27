# 镜像分层进度（2026-09-12）

F09/F11增量，整项仍进行中。

发现image.pull虽然解析Engine progressDetail，却只保存状态文本，Current/Total始终为空。现在共享emitProgress保留当前层Layer、Current/Total、状态消息，写入节点操作日志后再通知Daemon；Daemon任务修订及任务视图接入同一层身份与计数。保留原有200毫秒采样与最近256条操作进度上限，不据此编造整个镜像百分比。负数或超出浏览器精确整数范围的计数被丢弃。

新增HTTP Engine协议替身测试覆盖分层解析、日志保存、通知回调、无效计数，以及最终镜像检查失败时不因进度存在而报告成功。首次测试漏掉/version协商，未产生操作日志而失败；补齐测试协议后定向race1.024s通过。它是真实HTTP和生产解析/持久路径，Engine仍为测试替身，不算真实镜像下载验证。make check/build/windows与前端生产构建通过；容器模块完整race结果见PROGRESS。

仍需处理采样期间最后一条计数保留、完整进度诊断浏览入口、Compose成功执行输出持久化与流式诊断，以及真实拉取/远程故障证据。当前Compose已有阶段记录和有界stderr/stdout缓冲，但成功输出尚未持久化；不能将阶段存在称作完整执行日志。

最终`go test -race ./internal/containers -count=1`退出0（2.553s）。未设置真实Docker验收入口的环境开关，本次全包回归不代替真实Engine下载或Compose部署测试。

后续拉取读取边界：正常EOF前保存采样间隔内最后一条观测，避免短流的末条层计数消失。读取使用32 MiB加1字节探针，达到超限时明确失败且保留Unknown，不把人为限流产生的EOF作为远端结束，也不继续检查可能已存在的旧镜像。新增真实HTTP替身在原上限处填满合法JSON/空白、上限后发送失败记录，验证不会执行后续镜像检查；定向`go test -race ./internal/containers -run TestPullProgress -count=1`通过（3.743s），包含末条计数持久化。最终make check/build/windows通过。

Compose成功输出仍未持久化；当前stdout/stderr各保留最多64 KiB的缓冲，错误返回才使用其中正文。下一步应记录有界输出与截断标识、提供获准读取和实际CLI验证，而不是将当前缓冲宣称日志已交付。
