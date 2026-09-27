# B08 Docker/Compose 的真实节点 API

2026-09-09。Fedora Linux amd64、Go1.25.9、Docker29.4.1、Compose5.1.3。专用镜像`blora-isolated-e2e:20260909`；所有变更仅涉及随机命名的本任务新对象。测试未更改既有容器、卷、网络或防火墙。

## 接线与边界

`GET /nodes/managed`按当前`host.manage`过滤。`POST /nodes/{id}/docker/query`读取固定资源或有界列表；`POST .../actions`记录持久化节点任务。Daemon有独立的两个host操作槽，不占实例停止预算，执行前核对完整任务身份和当前授权。容器/镜像/网络使用完整ID；卷使用名字、出生时间与配置指纹。托管实例的容器生命周期要求通过实例控制入口，保留run代次、停止策略及归属退出确认。

Compose正文最多1MiB，准备保存绑定actor、saveId、projectId、基准revision、总长度与裸十六进制SHA256；bulk通道每片64KiB、持久化offset。提交为`compose.save`任务，其control正文仅含saveId。项目正文按不可变revision分片读取，独立`compose.apply`与`compose.delete`绑定所选版本，保存不自动应用。

Docker日志WSS是每秒更新的保留时间窗口，明确`mode=timestamp-window`和`possibleGap=true`，不提供虚构连续事件游标。每次替换的窗口最多48KiB/256帧；浏览器确认消费后才请求下一窗口，45秒慢消费截止；实时撤销host权限关闭流。Daemon现将成功窗口以节点私有目录原子归档，最多100个窗口，并用 Daemon 锁串行化同一节点的读改写，通过`GET /nodes/{id}/docker/containers/{containerId}/logs/history`提供带`possibleGap`标记的历史读取。归档仍未完成跨节点云端副本和完整实例生命周期绑定。

## 已运行

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 go test -race ./internal/master -run TestContainerNodeAPI -v -count=1 -timeout=90s
```

首次真实API两项4.647s通过；追加真实WSS中文输出、持有ACK时停止、host-only节点发现及撤权断流后两项3.518s通过。实际创建、启动、查询、停止、删除同一个完整ID容器，清理无错误；无Engine配置时任务明确失败，普通实例成员调用host接口403，任务同key复用。

```sh
env GOFLAGS=-buildvcs=false GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go BLORA_CONTAINER_E2E_IMAGE=blora-isolated-e2e:20260909 go test -race ./internal/master -run TestComposeProjectSave -v -count=1 -timeout=120s
```

退出0，5.011s。约480KiB中文注释与真实Compose正文分片保存，首片后实际重启Daemon并重复同片；提交同key返回原任务，按固定revision读回整文hash一致。保存后没有容器，显式apply创建一个服务，随后delete清理。首轮测试错误使用文件模块的`sha256:`前缀被身份检查拒绝；按项目模块裸hex合同修正后通过，没有削弱校验。

底层模块的真实Docker/Compose、镜像拉取、保留卷、健康失败、日记恢复与mTLS状态见[模块报告](containers-2026-09-09.md)。`make windows`与`go vet ./...`本轮退出0，仅说明构建/静态检查；Windows真机仍缺失。

## 继续工作

Docker Center/Compose前端正在接线，host container exec复用PTY适配器在实现，mTLS共享配置在补全。日志窗口的节点私有归档与历史 API 已接入；容器删除得到 Engine 成功确认后清理对应归档，失败或结果不明保留诊断历史。实例生命周期绑定、跨节点/云端副本仍待实现。备份/调度、监控、有限系统管理及扩展SDK继续。F11/E02整体尚未通过，当前真实子链路不能替代全范围验收。
