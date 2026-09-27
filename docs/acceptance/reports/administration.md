# B03 用户、角色模板和节点维护/换钥

2026-09-09；以下为子场景证据，完整F03/F04仍需所有UI、故障和审计交付验证。

## 已实现的后台行为

- 用户资料、管理员标志、禁用状态使用revision前置条件；元数据修改及幂等回执同一SQLite事务。修改会撤销该账号会话/现有流，事务保证至少一个未禁用平台管理员。管理员创建用户也要求请求键，同键重试复用首次用户身份，密码摘要只用于比较请求内容。
- 自助改密要求当前密码，平台管理员可重置其他账号密码；密码12～72字节，bcrypt存储，成功撤销全部会话。密码和原密码不进入任务或审计正文；新密码仅以不可逆摘要参与幂等比较，改密请求键在授权后校验并写入同一事务回执。
- 内置只读成员、实例操作员、节点管理员模板，自定义模板带CAS版本和数量上限。模板按指定真实资源展开成显式grants；修改模板不静默改变既有授权。host.manage/instance.create只适用节点。应用/撤销模板重新鉴权并撤销旧流。
- 节点名称/分组/标签/维护/配额使用独立configRevision，心跳只更新观察字段并保留管理员分组标签；维护拒绝新创建/启动/重启，已有运行继续，停止仍允许。配额在创建事务再次检查。
- 管理员生成十分钟一次性换钥票据；新私钥在发送请求前写入Daemon受保护pending记录。Master验证票据与新密钥签名，事务替换公钥并提升连接代次。响应丢失重试同一密钥，已使用票据不能切换成另一把密钥。普通换钥保持nodeId与关联资源；已吊销节点须显式reactivate授权。

## 实际命令与结果

各Go命令均使用Makefile对应缓存和GOFLAGS=-buildvcs=false。测试监听临时本机TLS端口，使用两个真实Daemon对象、签名WSS、临时SQLite/进程，不打印随机密码、票据或私钥。

```sh
go test -race ./internal/master ./internal/storage -run 'TestRolesRemain|TestUserChanges' -v -count=1 -timeout=60s
go test -race ./internal/master -run '^TestNodeMaintenance' -v -count=1 -timeout=60s
go test -race ./internal/master -run '^TestNodeKeyRotation' -v -count=1 -timeout=60s
```

均退出0：角色/账号master2.941s、storage1.058s；节点维护2.164s；换钥1.747s。角色范围不泄漏到其他实例或主机；并发禁用两个管理员只允许其中一个成功；改密旧会话失效、新密码可登录。节点维护运行ID不变且可实际停止；重放旧心跳保留最新管理设置。换钥实际关闭旧控制/bulk连接，新Daemon以同ID重新建立WSS，已提交票据重试不生成第二把钥匙，第二次显式换钥也成功。

维护测试第一次发现配额超限统一返回403权限错误；改为独立容量冲突409后复验通过。完整新UI浏览器验收在desktop报告继续追加，以上后台测试不能替代UI。

## 换钥操作入口

从节点管理接口 `POST /api/v1/nodes/{nodeId}/rotation` 获取票据JSON（请求体 `{"reactivate":false}`；普通成员拒绝）。将响应JSON以0600保存到节点上的私有文件，在Daemon私有配置中设置 `rotationFile` 指向它，再重启管理Daemon。新建 `rotation-pending.json` 与现有 `identity.json` 均为0600。确认换钥/新连接成功后，可移除配置的rotationFile引用并删除已消费的票据文件；保留pending记录便于同票据重试，下一次已确认换钥可替换它。

管理Daemon重启不等于停止托管实例。PTY在Daemon进程重启后的恢复边界仍按会话能力展示，不能偷偷创建新会话。当前尚需针对传输/日志的Daemon进程崩溃验收，以及新增UI、完整审计查看/保留和用户现场隔离组合。

2026-09-10 增量：节点卡片支持管理员确认后的“撤销管理”，Master 以事务幂等回执递增代次、保留实例资源并断开旧连接；同一请求重试不重复递增。节点分组与逗号分隔标签通过节点设置 API 保存，Daemon 心跳不会覆盖这些字段。相关集成测试见 `audit_integration_test.go` 和 `node_settings_integration_test.go`；完整失联/Windows 组合仍未验证。
