# 跨节点实例中心多视图（2026-09-13）

关联A15。`web/tests/real/desktop.spec.ts` 新增 `two real cross-node`，使用实际HTTPS fixture `.local/fixture-4012957923`、两个Daemon、Chromium及真实原生测试实例，无API替身。

两个实例中心窗口各打开两个不同节点的实例标签，均保留多个标签；两个窗口重复查看节点一实例。第一窗口切文件页，第二仍为概览。回到概览后分别打开固定目标启动确认，在同一个浏览器任务中点击两确认按钮，让不同requestId的控制并发到达后端。

最终断言：接受的任务全部进入终态，只有一个SUCCEEDED；失败任务必须是 `existing running unit has not exited`，所有结果runId与实际实例一致。两个窗口均自动显示RUNNING及相同runId。第一窗口切文件后刷新，两个原视图ID唯一，导航分别保持文件/概览；最后实际停止实例并等待STOPPED。

命令：web目录设置 `BLORA_E2E_CREDENTIALS` 为该fixture私有凭据路径，运行 `npm run test:e2e:real -- --grep 'two real cross-node' --reporter=line`。最终31449退出0，1/1（7.9s）。此前较弱结果断言26946也通过（6.1s），随后读取实际SQLite确认一成功一拒绝及相同runId并将此约束加入最终测试。首轮36852桌面图标被测试窗口遮挡，改用任务栏入口；59146隐藏标签导航也被断言匹配，限定可见导航后修正。生产代码无修改。

本场景验证不同请求的资源级串行；相同requestId去重、重连丢响应和故障运行归属需结合已有后端专项证据，不将此一次浏览器并发当作全部故障证明。fixture已Ctrl+C停止，仅操作自建实例。Windows真机仍见E09。
