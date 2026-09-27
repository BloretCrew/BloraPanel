# Blora App SDK

`host.notify({title,message?,key?})` 需要 `notification.publish`，服务端每次复核扩展当前启用状态、能力及用户 `app.use`。标题最多256 UTF-8字节，正文最多4096字节；来源使用宿主绑定的 appId，不能发送 HTML、可执行动作或外部跳转。通知固定关联调用视图及其资源，点击“打开来源”才激活原标签或重新打开同应用资源。

站内通知保存在当前账号工作区的恢复日志，刷新不会重新发送；最多保留100条，每个应用每秒最多5次，同一可选key替换旧通知。该接口不请求浏览器系统通知授权，也不发送给其他账号。仅同步布局的云端副本排除通知正文；明确同步工作内容才包含通知。独立示例的“发送站内通知”可验证这条链路。

`host.updateResource({revision,name?,group?,tags?}, requestId, resource?)` 使用 `resource.write` 修改实例元数据；省略resource时绑定当前视图实例。先从`readResource()`取得`configRevision`作为revision（读取另需resource.read），请求标识由应用保存并在重试时复用。服务端还要求app.use与instance.configure；不接受执行配置、宿主命令或节点修改。成功返回最小元数据与新configRevision，旧修订返回冲突。实例启停及其他资源写入仍为独立能力范围，不由此接口默认授予。

桌面能力：声明 `window.move` 后可用 `host.moveView(targetWindowId?)` 将调用者自身标签移入同应用兼容窗口，省略目标则移出新窗口；未知窗口和其他应用窗口被拒绝。声明 `window.close` 后 `host.closeView()` 仅关闭自身视图，草稿与后台任务继续保留。声明 `shortcut.create` 后 `host.createShortcut(title?)` 为当前资源创建桌面入口；总览没有资源时拒绝。扩展不能通过 payload 指定其他应用或其他标签作为操作对象。快捷入口只保存真实资源引用，打开时沿用宿主重新授权；不复制状态或启动资源。移窗可能重新挂载 iframe，新视图通过 capture 恢复现场，不能在 start 中自动重复移动或提交任务。

发布者签名覆盖完整受支持清单与包体摘要，使用域标识 `BLORA-EXTENSION-SIGNATURE-2`。从项目根目录运行 `go run ./cmd/extension-sign --package sdk/examples/reference-app/reference.blora-extension.json --key /安全路径/publisher.pem --out /输出路径/reference.signed.json`；密钥文件为 Ed25519 PKCS#8 PEM，输出必须是新文件。工具只输出签名包路径与公钥，公钥可配置为 Master 的 `--extensions-public-key`。私钥不应加入扩展包或项目仓库。

签名使用 Go `extensions.SignatureMessage` 生成的规范消息：安装默认值归一化后的 Manifest 与实际包体 SHA-256 经 Go JSON 序列化，并加上述域前缀和换行；依赖对象键排序。请使用签名工具，避免不同语言的 JSON 表示差异。旧的仅包体摘要签名不能认证能力清单，升级后不再接受；已有签名包需由发布者重新签名。管理员本地无签名模式仍仅用于明确允许的可信包安装，不能代表发布者认证。

独立扩展包通过版本化 `AppManifest` 声明入口、资源类型、窗口/标签策略和能力。`createSandboxHost` 将能力检查放在每次调用前；宿主传输层仍需对节点、资源、用户和任务权限做服务端复核。参考包位于 `examples/reference-app`，不修改桌面核心即可单独构建。

```sh
cd sdk
npm install
npm run check
npm run build
cd examples/reference-app
npm install
npm run build
npm run package  # 生成可由 Master /extensions/install-package 接收的包文件
```

包文件是 JSON 信封，包含清单、SHA-256 和 base64 包体；包体仍由服务端重新校验，不能凭客户端清单获得权限。参考扩展导出 `start` 和 `enqueueExampleTask`：后台任务必须经过 `task.create` 能力门禁，宿主传输层仍需在服务端重新授权。

`npm run package` 使用锁定的 esbuild 把 SDK 和示例合为浏览器 ESM 包。示例提供笔记、节点资源开窗与任务提交界面。`host.readResource()` 读取当前视图资源；总览没有绑定资源时返回 `null`。也可传入 `{kind:'instance'|'node', id, nodeId?}` 读取明确目标。资源摘要来自 `GET /api/v1/extensions/{appId}/resource`，服务端校验已启用扩展的 `resource.read`、调用者 `app.use` 和目标 `instance.read`/`node.read`；实例节点提示必须与服务端记录一致。摘要包含资源身份、名称、状态和修订号，不包含执行配置或凭据。此接口不授予文件读取、实例控制或主机管理权限。

宿主传输层提交任务时调用 `POST /api/v1/extensions/{appId}/tasks`，请求体为 `{nodeId, payload}`，并携带 CSRF 与幂等请求头。服务端只接受已启用、声明 `task.create` 且包含 WASI 后台模块的扩展，调用者须有 `app.use` 与目标 `node.read`。输入为最多 32 KiB 的 JSON，任务绑定包摘要和模块摘要。Daemon 通过批量通道分块读取模块，每块复核权限/版本/取消状态，执行前再次授权；升级后未执行的旧版本任务明确失败，不改为执行新代码。同请求键的回放仍复用原任务。

组合包体以 `BLORA-BUNDLE-1\n` 开头，随后为 `{frontend: string, backend: base64}` JSON；外层包的 SHA-256/签名覆盖整个包体。前端由 esbuild 打包，参考后台 `backend/main.go` 使用 `GOOS=wasip1 GOARCH=wasm` 构建，运行环境需要 Go 工具链。后台从 stdin 读取 JSON、向 stdout 输出单个 JSON 值；不提供宿主文件、环境、网络或命令执行。执行期限 5 秒、线性内存最多 64 MiB、stdout/stderr 各最多 32 KiB。示例计算笔记字符数、词数和 SHA-256，结果通过持久任务中心查看。旧的纯 JavaScript 包仍可打开前端，但没有后台模块时创建任务返回明确错误。

`host.createTask(payload, requestId)` 返回 `ExtensionTask`；调用前保存请求键和正文，丢响应后以相同参数显式重试。SDK 向宿主传递 `{payload, requestId}`，宿主将键放入 HTTP 头，不加入 WASI 输入。用 `host.readTask(taskId)` 查询实际状态/结果，`host.cancelTask(taskId, requestId)` 请求取消。取消键应与创建键分开生成，随任务 ID 保存，重试沿用原取消键。三个方法均要求声明 `task.create`。查询与取消绑定原任务 ID，不随当前窗口的节点切换而改变目标；服务端校验应用身份、任务发起者及当前 `app.use`/`node.read` 授权。禁用或卸载后，仍获准的用户可通过任务中心查询和取消历史任务；撤权后立即拒绝。取消回执可能仍是运行中，只有节点确认后的 `CANCELLED` 才表示已取消，已经成功的任务不会回滚。参考扩展保存待提交请求和任务 ID，刷新不自动重新提交。

扩展管理中选择 `reference.blora-extension.json` 后点击“安装或升级包”，系统保留原清单和摘要。首次安装走完整包 API，同 ID 的新版本走升级 API；包文件上限 16 MiB。

应用 ID 不得使用 `blora.` 前缀或 `.previous` 后缀。注册表的安装、升级、回滚和卸载会保存持久事务记录；Master 重启时自动恢复未提交修改，并保留已提交版本。遇到“事务需要恢复”时，修复磁盘/目录权限问题后重启服务；不要删除事务记录绕过恢复。恢复记录完整性损坏时需从已验证的注册表备份恢复。受控 JSON 用户数据与四个包文件共同参与事务恢复；旧恢复记录没有用户数据字段时不会删除已有数据。

状态版本变化时，从前端模块导出 `migrate(previous: AppState, targetVersion: number): AppState | Promise<AppState>`。它在 opaque 沙箱中、`start` 之前运行，输入为旧现场的副本；返回 `{schemaVersion: targetVersion, state: ...}`。宿主检查旧现场仍未变化、目标版本正确后，将正文状态和版本一次提交。迁移缺失、抛错、超过 10 秒或现场冲突时保留旧现场并显示错误，不启动新版应用；不能把旧状态简单改号冒充迁移。同状态版本的升级会更新窗口组件与能力声明。服务器数据遵守下面独立的迁移合同。

声明 `data.read`/`data.write` 后，用 `host.readData()` 读取当前用户文档，用 `host.writeData(data, expectedRevision, requestId)` 条件保存。返回 `{schemaVersion,revision,data}`。服务端只取登录身份，不接受客户端指定 owner；管理员也不会通过此接口读取其他用户数据。接口为 `/extensions/{id}/data` 的 GET/PUT，仍要求当前 `app.use` 和已启用能力。`requestId` 应在首次发送前随现场保存，结果不明时使用同一请求身份；重复请求返回当前文档，不覆盖后续保存。修订冲突先读取比较，不能静默覆盖。

每用户正文上限 16 KiB，扩展所有用户文档及持久幂等记录合计上限 4 MiB。额度不足时拒绝新增写入并保留原数据；幂等记录不会偷偷淘汰。卸载默认保留数据，显式清理会删除该扩展所有用户数据和回执。此存储为受控小型 JSON 数据，不能当作无限文件仓库。

Manifest 的 `dataSchemaVersion` 默认 1。安装（包括重新安装保留数据）、升级或回滚遇到不同数据版本时，目标包 WASI 后台收到 `{operation:"blora.data.migrate",fromVersion,toVersion,data}`，必须返回 `{schemaVersion:toVersion,data:转换后的JSON}`。每个用户在独立内存实例中运行，不传入其他用户数据或身份；一批迁移复用编译模块，整批期限 60 秒，单次执行仍为 5 秒。所有转换先完成并校验，再与包文件共同提交。任何转换失败或超额都保留当前包与数据。回滚也转换当前数据，不能靠恢复旧快照丢掉升级后用户的新写入；目标旧包不支持迁移时回滚明确失败。

`make sdk` 同时生成默认 0.3.0/schema 1、`reference-v2.blora-extension.json`（0.4.0/schema 2）及 `reference-v3.blora-extension.json`（0.5.0/schema 3）。后者故意拒绝服务器数据迁移，用于失败保留测试；不要当作可用升级发布。示例支持保存服务器笔记，并在 schema 1/2 之间保留 note、更新 noteFormat，同时迁移窗口现场。

数据转换在注册表锁外运行，其他应用授权和用户写入可继续；重新取得锁后校验包、数据与提交令牌快照，再次验证当前签名信任和依赖。并发变更导致条件提交冲突，保留当前包与最新数据，可重新读取后重试。每个 Manager 同时只接受一批迁移，第二批明确拒绝；生命周期 API 当前仍同步等待结果。

桌面宿主通过认证的 `GET /api/v1/extensions/{appId}/bundle` 读取已启用包体，并应放入 `sandbox="allow-scripts"` 的 iframe，通过 `postMessage` 连接 transport；不要把 bundle 直接注入宿主页面。`loadExternalSandboxApp` 和 `ExternalSandboxApp.vue` 提供了当前宿主实现，禁用扩展的 bundle 请求返回 409。
