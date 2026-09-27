# Master 任务终态审计补齐

2026-09-12，F03/F09 子项。`ReconcileTask` 接受真实 Daemon 终态时，在同一事务写入审计；派发前取消的 `mutateTask` 路径也补终态审计。审计只包含主体、节点、资源、动作、请求、状态、时间，不写任务输入或结果正文。重复终态上报被已有去重条件拦截，不重复产生审计。

`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/storage -count=1`：通过，2.404s。新增 SQLite trigger 故障注入证明审计失败会回滚终态、事件和版本；解除故障后重试产生一条终态记录。

相同环境 `go test ./internal/master -run 'TestIndependentReferencePackageResourceBridge|TestCancelAcceptedUploadBeforeNodePreparation' -count=1`：本机 TLS 授权运行，通过，4.402s；真实 WASI 节点终态在 Master 审计中恰有一条。`GOCACHE=/tmp/blora-go-grant-idem make check build` 通过。没有将此定向验证当作全仓 race 或完整审计保留策略验收。
