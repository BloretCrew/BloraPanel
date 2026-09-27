# 真实节点失联与浏览器呈现（2026-09-13）

关联A05。新增 `nodes.spec.ts` 的actual paused Daemon，真实HTTPS双节点fixture与运行实例。测试从/proc查找命令行严格匹配本fixture配置文件的Daemon，要求唯一并记录进程出生标识；SIGSTOP暂停它，等待真实Master将节点判为OFFLINE。页面必须显示“节点失联，运行结果等待重新确认。最近记录不代表当前进程已停止。”，API实例runId保留且不伪造STOPPED。

finally再次确认出生标识后SIGCONT，等待ONLINE并核对runId与失联前相同，再实际stop等待STOPPED。未对其他进程发送信号，也未通过路由替身伪造节点状态。

fixture `.local/fixture-3470920766`，web目录设置其私有凭据路径运行 `npm run test:e2e:real -- --grep 'actual paused Daemon' --reporter=line`，32702退出0，1/1（50.5s）。fixture已Ctrl+C停止。生产代码未修改。

此场景证明真实失联状态与原运行恢复的浏览器呈现。重启/传输中途断线已有后端专项证据，但浏览器具体任务等待阶段仍需核对；不把稳定运行失联场景冒称全部任务组合也已通过。

任务等待后续通过：同一真实失联测试在OFFLINE期间通过实际API提交stop（202），等待原任务WAITING_NODE，打开生产任务中心并按完整taskId定位，必须显示“等待节点”。finally恢复Daemon，原taskId达到SUCCEEDED、任务卡显示“成功”，实例实际STOPPED且runId仍原值，没有另建停止任务代替等待任务。

fixture `.local/fixture-3110999777`，同命令30032退出0，1/1（50.7s）；fixture已Ctrl+C停止，生产无修改。该场景补齐任务中心等待/成功可见状态；重启任务中断线及传输断线组合分别见restart-disconnect和transfer-proof/transfers报告，不混同为本浏览器单场景实测。

传输中途边界最终复核：`TestTransferDisconnectAndMidstreamPermissionRevocation`加强为任务RUNNING且0<CurrentOffset<总长度时才关闭实际目标bulk连接；恢复后同父任务SUCCEEDED、CommittedBytes准确、DestinationVerified，直接目标字节完全一致，固定上传提交子请求计数必须恰为1。后续同测试还验证中途撤权失败且未发布目标、清理确认与权限过滤。

`go test -race ./internal/master -run '^TestTransferDisconnectAndMidstreamPermissionRevocation$' -count=1 -v`（既有缓存环境、本机TLS授权），96813退出0（12.490s）。结合重启RUNNING断管理连接原请求复用的restart-disconnect报告，以及本报告实际浏览器OFFLINE/WAITING_NODE/同任务恢复，A05各明确分支已有Linux直接证据；不能将其扩展为所有远程网络或Windows运行条件均已测。
