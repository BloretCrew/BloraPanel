# Compose CLI 进程恢复

关联F09/F11/A10/E02，整项仍进行中。

Linux CLI启动前将随机归属令牌原子写入`containers/cli-runs/`并同步目录，启动后补写会话号。调用返回后核对会话、出生标识、环境令牌和pidfd，清理遗留后代；确认退出才删除归属记录。令牌扫描覆盖会话号尚未落盘的窗口，也覆盖保留令牌但改变会话的后代。记录内容和数量有界，未提交的`.pending-*`文件不作为已启动命令。

Daemon在连接Master、恢复任务和接收新工作之前执行RecoverCLI；记录损坏、取消或无法确认进程归属均保留证据并拒绝启动。运行期间无法确认CLI退出时，操作保留Unknown，管理器阻止后续Engine变更，直到启动恢复成功。资源检查一时成功不能解除这一限制。终止CLI不证明Engine已撤销先前接受的请求，原任务仍需按资源事实对账。

Windows沿用原有原子Job分配和kill-on-close，新增接口不以Linux证据替代Windows真机验证。

验证环境：本地Linux、受控独立Go测试进程、已有隔离Docker镜像。

- 独立执行器SIGKILL测试覆盖会话号已保存、模拟仅保存启动前令牌两条恢复路径：产品RecoverCLI结束遗留CLI，确认记录清除、重复恢复幂等，保留原输出且同任务不重放。
- 损坏归属记录、取消、空启动记录清理、禁止后续Engine变更；Daemon.Run在运行时和连接初始化前拒绝未确认归属，定向race1.035s通过。
- 首轮容器回归因`/proc/.../environ`返回ESRCH被误判检查失败而退出1；修正消失进程分支后通过。
- 最终`go test -race ./internal/containers ./internal/daemon -count=1`：containers 7.900s、daemon 30.518s，退出0。只启用普通套件，条件Docker入口不据此计通过。
- 显式`BLORA_CONTAINER_E2E=1 go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1`：真实Docker/Compose 27.047s退出0，随机标签对象由测试清理；该次在最后追加管理器拒绝后续变更门禁之前，正常执行路径已验证。
- 最终`GOCACHE=/tmp/blora-go-grant-idem make check build windows`退出0。

Go命令共用`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false`。没有活动fixture或测试进程。

完整Daemon被SIGKILL后重新连接Master的组合、物理掉电/ENOSPC、Windows真机与远程Engine仍待验证。本次PID落盘前测试通过重写自有记录模拟该持久状态，不能称为在fork后精确故障点注入。
