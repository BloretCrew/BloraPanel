# 活跃/后台编辑窗口立即刷新（2026-09-13）

关联A01。`desktop.spec.ts`新增两种焦点参数的latest IME paste：真实HTTPS登录、Monaco，Chromium CDP imeSetComposition/insertText提交中文组合；真实navigator.clipboard.writeText及Control+v粘贴，连续两次输入后撤销末步，实际鼠标拖窗。在活跃窗口或切任务中心后后台窗口状态下直接reload，不等保护标志/数据库快照或防抖时间。

刷新核对中文/粘贴/连续正文、末步撤销位置、原windowId与精确transform、焦点关系；redo恢复末步，再输入最后一句并直接reload，正文必须包含最新输入。没有测试替身后端或简单文本框。

fixture `.local/fixture-4013429189`，web目录设置私有凭据后运行 `npm run test:e2e:real -- --grep 'latest IME paste' --reporter=line`。16486退出0，2/2（6.0s）。fixture已Ctrl+C停止，生产代码未修改。此为Linux Chromium验收，不推断其他浏览器/Windows平台已经运行。
