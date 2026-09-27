# 扩展注册表中断恢复

安装、升级、回滚与卸载使用持久撤销记录。修改包体/清单前，记录当前清单、当前包体、上一版清单、上一版包体的完整快照及文件是否存在；事务记录以临时文件、文件同步、重命名和目录同步写入，包含完整性摘要及随机事务令牌。完成全部文件修改后持久化提交令牌，再移除撤销记录；重启遇到匹配的已提交令牌时保留新版本，不会因为撤销记录尚未删除而误回滚。

启动时先检查并恢复未提交记录，再开放扩展注册表。恢复可重复执行；普通写入失败也执行同一恢复路径。恢复记录损坏或超过上限时阻止注册表启动，不据此覆盖现有文件。尚有待恢复记录的包不能被读取为正常安装状态。回滚快照自身也在同一事务内恢复。安装不会把损坏元数据误当作“未安装”覆盖。

新增应用 ID 保留后缀 `.previous`：当前平面文件布局中它会与其他应用的回滚快照命名冲突。Master、SDK 和 AppHost 均拒绝该后缀。

验证：

```sh
GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem go test -race ./internal/extensions -run 'TestRegistry|TestRollbackFilename' -count=1
```

子进程在持久记录之后、包体写入后、清单写入后、回滚快照写入后、删除文件后、提交标记已写但撤销记录未删时直接 `os.Exit(91)`；父进程重新创建 Manager。覆盖初次安装和已有版本两类状态，未提交时恢复全部旧文件/缺失状态，提交后保留新版本；同时验证写入错误恢复、记录损坏拒绝和重复打开。此为实际进程中断及磁盘文件测试，不是物理断电、ENOSPC 或 Windows 真机证据。

完整扩展模块 `go test ./internal/extensions -count=1` 通过（7.169s，允许本机 TLS）；真实 TLS `TestIndependentReferencePackageResourceBridge` 通过（4.075s），包含独立大包安装/升级与 WASI 计算。最终定向 race、静态检查和平台构建结果在进度记录中追加。

最终结果：中断恢复/命名冲突/生命周期定向 race 通过（1.385s）；最新扩展普通定向通过（0.119s）、Master TLS 独立包回归通过（4.453s）；`make check build windows` 退出 0，前端类型检查与 SDK 构建通过。无活动测试会话。本轮未重跑全仓 race，不能把这些针对性证据扩张到未覆盖模块。

Windows 适配新增 `MoveFileEx(REPLACE_EXISTING | WRITE_THROUGH)`，处理长路径与 UNC；删除先用同一接口把文件移出有效命名空间，再清理临时文件。提交令牌在撤销记录清理后仍保留。接口语义依据 [Microsoft MoveFileEx 文档](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw)，交叉构建结果不等同于 Windows 真机耐久性验证。

限制：按单 Master/单注册表进程使用；不承诺多进程同时写同一目录。Linux 目录同步已执行，真实掉电/文件系统耐久性仍需分别验证。后端用户数据迁移与 SDK 任务查询/取消仍待继续实现。
