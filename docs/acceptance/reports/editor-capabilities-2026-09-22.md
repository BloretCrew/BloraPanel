# 编辑器能力增量验收（2026-09-22）

本报告覆盖文件编辑器的格式元数据、只读授权提示、分栏现场、多光标撤销恢复和有界撤销日志子项，不代表 F03 或 F08 整项通过。

## 本轮能力

- `/files/content` 返回的编码、换行符和最大字节数进入编辑器状态栏；UTF-8 BOM 在模型正文中剥离，保存/导出时恢复，CRLF 加载时保留。保存基线区分服务端序列化正文与编辑器可见正文，避免 BOM 使草稿误报未保存。
- 编辑器通过 `/files/access` 查询当前用户的节点状态、读取和写入能力。缺少写权限或权限查询失败时只读并禁止本地编辑/保存；导出仍可用，服务端授权仍是最终边界。
- 分栏的两个 Monaco 视图共享正文模型，各自恢复光标/视图状态；按等宽布局显示。
- 文件名/扩展名映射到 Monaco 语言，未知类型按纯文本处理。
- 多光标编辑的两处替换在普通刷新后仍可一次撤销并重做。
- 撤销历史使用与正文及刷新恢复相同的预算：最多 8 MiB，或文件字节上限的两倍（取较小者，最低 1 KiB）。超限压缩完整的旧操作组；当前输入组若自身超过预算，则整组并入新基线，避免撤销半次粘贴。
- 超过 4 MiB 和非 UTF-8/含 NUL 文件打开失败时，错误区给出原文件下载入口；超过上限继续被编辑 API 拒绝，不把下载误认为文本编辑。
- 超过 4 MiB 的文件现在可通过 `/files/preview` 读取受版本绑定的有限 UTF-8 分段；前端只读显示当前字节范围，支持上一段/下一段，仍保留原文件下载入口。Master 和浏览器都不缓存完整大文件。

## 验证

- `env GOCACHE=/tmp/blora-go-build go test ./internal/master ./internal/filesystem`：受限沙箱无法绑定 `httptest` IPv6 端口；同一文件集在获准本地回环环境以 `go test ./internal/master -run 'TestFileAPIRealNodesTasksAndConditionalSaves|TestFileActionsArchiveAndRecycleThroughDurableTasks' -count=1` 通过（4.111s），包含大文件首段/第二段预览和二进制拒绝。 
- `npm run test:e2e -- tests/browser/editor-boundaries.spec.ts`：3/3 通过（14.9s），覆盖超限下载入口、二进制下载入口，以及只读分段预览的上一段/下一段导航。
- `npm run check`：通过；`npm run test -- tests/recovery.test.ts`：14/14 通过（新增过大活动编辑组原子折叠）；此前 `tests/document-language.test.ts`：11/11 通过。
- `npm run build`：通过；Vite 仍报告 Monaco 编辑器分块约 2.67 MB、超过默认 500 KB 提示阈值。
- `docs/api/openapi.yaml` 以 PyYAML 解析：OpenAPI 3.1.0，105 paths / 122 operations（新增 `previewFile`）。

## 边界

目前编码仅区分 UTF-8 与 UTF-8 BOM，换行仅覆盖 LF/CRLF；分栏固定 50/50 且不可拖动；未知语言保持纯文本。4 MiB 上限仍是可编辑硬限制，超限文本仅提供受版本绑定的只读分段和下载，二进制不会按普通文本保存。单个编辑组大于预算时，正文保留为新基线且该组不可撤销；历史不会只保留该操作的尾部。浏览器客户端测试使用 API fixture；Windows、非 Chromium、远端及完整跨应用场景仍按矩阵分别跟踪。F03/F08 整体状态继续为“进行中”，桌面视觉改动按用户要求保持冻结。
