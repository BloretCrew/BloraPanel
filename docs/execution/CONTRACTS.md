# 实施公共合同

2026-09-09 起实施。Go 模块 `blora.dev/panel`，共享实体在 `internal/model`。文件修改使用当前环境的 apply_patch；不使用 shell 重定向生成源码。没有与本任务模块匹配的专用编码技能；并行模块遵循根 AGENTS.md 及完整规划。

前端与 HTTP 的资源引用固定为 `{kind, id, nodeId?}`。所有 ID 为不携带授权含义的随机标识。HTTP JSON 使用 lowerCamelCase。API 前缀 `/api/v1`。错误为 `{error:{code,message,requestId?,resource?}}`。查询集合为 `{items:[...]}`。HTTP 修改使用 `X-CSRF-Token` 和 `Idempotency-Key`，返回真实任务（202）。

浏览器：`GET /api/v1/session` 返回 `{user,csrfToken}`；未登录返回 401。`POST /api/v1/login` 输入 `{name,password}`，同一响应合同；`POST /api/v1/logout`。身份确认后才读本地工作现场。基础查询 `/nodes`, `/instances`, `/tasks`。任务不能由恢复逻辑重新提交。

节点：单次登记、Ed25519 挑战认证、TLS WSS，控制/交互/批量通道物理分离。管理消息 Protobuf 包络含协议版本、连接代次、请求 ID、流 ID、序号、类型、负载和额度。业务消息负载使用版本化 JSON，终端/文件 DATA 为原始字节。节点服务之间不存在业务网络转发。

模块所有权：主代理负责 `internal/model`, `internal/storage`, `internal/protocol`, `internal/master`, `internal/daemon`, `cmd`, 根构建与整体文档；桌面代理负责 `web`；运行适配代理负责 `internal/runtime` 及其专属测试/说明。根 `go.mod/go.sum` 由主代理维护，依赖需求通过消息协调。独立模块完成后由主代理集成验证，模块单测不等于整体验收。
