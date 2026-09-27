# 真实节点失联监控与缓存进程提示

关联F10/E01。新增`web/tests/real/monitor-offline.spec.ts`，实际TLS Master/双Daemon/Chromium，按私有fixture精确配置路径和PID出生身份定位所属Daemon；启动所属实例、采集节点和实例指标，SIGSTOP等待真实心跳超时，再SIGCONT重连，finally恢复并停止所属实例。

首轮54.8s退出1：后端节点/实例指标已设置stale，保留原observedAt与运行代次并附诊断，界面节点指标也显示旧数据；但进程列表仍无条件显示“实时列表”。这属于产品提示错误，未删除失败断言。

`MonitorApp.vue`现根据进程查询错误或节点旧数据状态显示“旧进程列表（当前查询不可用）”，保留最近成功读取时间，禁用旧列表上的新终止入口及翻页，继续自动查询。服务端PID+出生标识核验保持不变。

生产构建通过，修正后实际复验1/1（53.4s）退出0：两个缓存observedAt在重复读取中保持不变且不晚于暂停时刻，诊断可见、实例runId不变、非空进程旧列表动作禁用；另一节点仍能实时采样；恢复后observedAt前进、旧数据提示清除、进程列表重新显示实时。fixture-3280126874仅所属对象被暂停/恢复，最后停止；无生产进程被终止。Windows、远程机器和长期采样负载不由该场景证明。

运行命令：`BLORA_E2E_CREDENTIALS=<私有fixture凭据文件> npm run test:e2e:real -- tests/real/monitor-offline.spec.ts --workers=1`。当前修复晚于rc3，待统一发行新包，不改写rc3包含内容。

随后隔离运行`npm run test:e2e -- tests/browser/monitor.spec.ts --workers=1`，1/1（4.8s）通过；保留原有搜索、内存排序、分页/刷新恢复、切换节点后确认仍使用原nodeId和startTicks断言。该场景使用HTTP替身，与上述真实失联证据分别记录。

## 2026-09-20 全套回归顺序复验

新的49项真实浏览器全套运行发现，这个离线场景在SIGSTOP前没有等待首次进程查询成功；错误上下文证明当时正确显示“进程查询不可用”，因为没有可缓存的列表。测试现加入前置断言，先等待真实列表、最近成功读取时间和非空进程行，再暂停所属Daemon。修正后命令 `BLORA_E2E_CREDENTIALS=<隔离fixture私有凭据路径> npm run test:e2e:real -- tests/real/monitor-offline.spec.ts --workers=1 --trace=off --reporter=line` 在 fixture-1831668373 上1/1通过（54.0s）；缓存列表转为旧数据提示、终止/翻页禁用、指标时间戳固定，恢复后转回实时。此次仅修正测试前置条件，未改产品代码；夹具已停止并清理。
