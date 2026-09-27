# 实际拖拽与立即刷新（2026-09-13）

关联 A17/F01/F02/F08。真实 HTTPS fixture `.local/fixture-3185892968`、Chromium、生产 Monaco，新增 `web/tests/real/desktop.spec.ts` 的 `real pointer tear-out` 场景。

测试创建两个独立正文标签，对第二标签输入中文和末次编辑并撤销；实际 pointerdown/move/up 拖到桌面形成新窗口，立即刷新。验证视图 ID 唯一、窗口已改变、正文恢复，redo 恢复末次输入。随后实际拖回原窗口标签栏首位，立即刷新；核对只有原窗口、标签顺序为原第二/原第一 ID、正文及 undo 保留，切换第一标签仍是独立正文。没有用菜单点击代替鼠标拖出/合并。

命令：web 目录设置 `BLORA_E2E_CREDENTIALS` 为上述fixture私有凭据路径，运行 `npm run test:e2e:real -- --grep 'real pointer tear-out' --reporter=line`。最终会话86685退出0，1/1（3.8s），无pageerror。首轮54242失败于合并后页面保留两个Monaco模型、断言未限定可见编辑器造成strict mode；改为可见模型定位后通过，生产代码未修改。

该场景补齐未保存编辑器标签的真实拖拽组合。实际PTY此前覆盖菜单移窗及刷新，仍需补活跃终端实际拖出/合并的同一组合；A15实例中心跨节点多视图并发控制也仍在推进，不以本项代替。fixture已发送Ctrl+C停止，仅操作自建测试资源。

后续活跃PTY组合已补：`files-terminal.spec.ts` 的 `actual PTY pointer` 在真实PTY执行一次文件追加并打印标识，保留另一空终端标签，通过鼠标将活跃标签拖出、立即刷新，再拖回原标签栏首位、立即刷新。断言原viewTabId与sessionId、归属和顺序保留，屏幕标识恢复；所有WSS Data输入帧未增加，真实文件始终只有一个x，会话仍running且原ID唯一，最后显式结束会话，无pageerror。

命令同上，凭据改为 `.local/fixture-1635242649/browser-credentials.json`，grep为 `actual PTY pointer`。最终99231退出0，1/1（8.6s）。首轮59340定位同时匹配空终端容器而失败，限定已绑定会话的容器后通过，生产代码未改。测试fixture已Ctrl+C停止。编辑器与PTY这两个实际拖拽组合共同覆盖A17；不替代A15实例控制并发验证或Windows真机。
