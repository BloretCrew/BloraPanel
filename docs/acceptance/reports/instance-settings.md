# B07 实例配置、模板与自动启动

2026-09-09，Fedora Linux amd64、Go1.25.9；真实本机 TLS Master、两个 Daemon、实际原生程序与目录。这里只记录子项证据，完整实例应用与全范围验收继续。

## 实现合同

`GET /api/v1/instance-templates` 提供 Linux/Windows 通用程序、Java 游戏、隔离容器的配置预填。模板不安装或执行程序。`PATCH /api/v1/instances/{id}` 接受独立 `configRevision` 的比较更新与 Idempotency-Key；同请求返回原回执，不再变更版本。名称、分组、标签与实例真实身份分别保存。

`instance.configure` 是独立权限，创建者获得该实例的配置权；修改原生运行配置还要求该节点的 `host.manage`。事务内重新验证当前账号/授权。启动字段变化要求已停止或启动失败，且没有本实例或以其为来源的未完成传输；修改停止策略和自动启动开关无需停机。保存与执行分开。新任务接受事务拒绝配置读取后已经发生更改的旧启动/文件/传输请求，已接受请求仍按原回执对账。

不同工作目录的上传/回收站记录按根分别保存。Daemon 文件服务持引用期间不会被配置切换关闭；第一次使用的根沿用旧状态目录，其余根保留单独记录。切换期间仍有旧读者时，新根请求明确返回冲突，可随后刷新，不能误返回旧根内容。节点事实只更新运行状态，不覆盖 Master 配置。

自动启动在启用后下一次 Daemon 进程启动触发。Daemon 先确认既有运行并发送恢复清单，Master 对每个实例/启动身份记录一次决定；控制连接重连或 Master 重启不会再次启动已手动停止的实例。触发检查节点维护、已有操作、实际状态以及策略所有者当前的 `instance.start`/`instance.configure`。执行通过同一持久任务队列，Daemon 执行前再次向 Master 核对完整任务身份及当前权限。

## 已运行

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test -race ./internal/master -run 'TestInstanceSettings|TestAutostart' -v -count=1 -timeout=80s
```

退出0，5.328s：配置同键去重、版本冲突、分组标签过滤、真实工作目录切换及原根保留、延迟旧配置任务拒绝、运行中启动字段拒绝、通过新停止命令完成实际退出；实际重启 Daemon 后自动启动、同进程管理重连不自启、撤销所有者启动权后下次 Daemon 启动不执行。

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go go test -race ./internal/storage ./internal/protocol ./internal/master -count=1 -timeout=180s
```

配置、根服务引用与执行前授权接入后全部退出0：storage 2.416s、protocol 1.269s、master 73.276s。包含文件、PTY、控制台、传输、管理和生命周期已有回归。另有 `TestTaskCentreKeepsParentAfterLargeTransfer` 1.736s：1001 个内部子任务不会把传输父任务挤出任务中心，子任务仍可按ID读取。

## 尚待验证/补全

配置/模板前端正在接线；Windows 真机配置和自动启动未验。计划内监控、有限系统管理、Docker/Compose、备份与调度、扩展 SDK 和整体故障性能验证仍继续。当前更改不提供跨节点迁移现有实例，也没有通过编辑节点字段改变现有资源身份；创建时选择节点，跨节点文件任务使用已实现传输流程。
