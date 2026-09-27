# 任务中心重试增量

记录日期：2026-09-10。

任务中心现在可以为失败或中断的实例生命周期任务创建一条具有关联身份的新任务。Master 只允许 `instance.start`、`instance.stop`、`instance.restart` 和 `instance.kill` 走通用重试；它读取当前实例配置，重新检查资源身份、节点维护状态和当前账号权限，再以新的 `Idempotency-Key` 接受任务。`RetryOf` 写入任务正文并在桌面任务卡显示。文件、终端输入、传输、备份和主机变更仍由各自应用的核对或补偿流程处理，通用任务中心不会重放这些副作用。

验证命令：

```sh
GOCACHE=/tmp/blora-go-retry2 go test ./internal/master -run TestTaskRetryCreatesRelatedInstanceAttemptAndIsIdempotent -count=1
```

结果：退出码 0。测试使用真实 TLS Master、两个 Daemon 和临时文件根目录；故意配置不可执行的实例命令，先观察原始 `instance.start` 任务进入 `FAILED`，再创建关联重试，重复相同请求键复用同一重试任务，普通用户对管理员任务得到相同的未授权隐藏（404），并用 `limit=1` 核对任务列表分页和有界回执，最后核对重试任务也记录真实失败。测试不把 HTTP 接受当作执行成功。

桌面任务中心在刷新后保留重试请求键和关联任务 ID，显示已确认的结果正文，并提供专用任务窗口入口。完整前端类型检查、28 个单元测试和 30 个串行浏览器场景均已通过；当前浏览器套件没有真实 Daemon 故障注入，因此远程节点失联、Master 重启和长期重试组合仍未验收。
