# 当前源码回归与 rc13 候选包验收 — 2026-09-20

本轮暂停视觉调整，只修正自动化验收与新版应用市场之间的定位差异。应用市场现在以产品名称和自然版本文案显示扩展；真实浏览器测试不应依赖内部 appId 可见文本。给已安装行和目录卡片增加 `data-app-id`，并给桌面窗口增加 `data-window-mode`，供真实验收按稳定身份定位；这些属性不改变布局、样式或用户可见文案。跨窗口测试的拖放目标移出桌面吸附带，账号退出测试改为通过当前账号菜单操作。

## 验证结果

- `cd web && npm run build` 通过（vue-tsc、Vite 生产构建，退出码 0）。Vite 仍提示 Monaco/文档分片大于 500 KB；这是已有分片警告，本轮不涉及分片策略。
- 真实 HTTPS 双节点桌面定向场景通过：专用实例窗口、双窗口跨节点状态/控制、账号退出后现场隔离。跨窗口场景此前因测试拖入吸附区而相互遮挡，调整测试位置后单项复验通过。
- 扩展真实浏览器共 7 个独立场景均已通过：通知保存/刷新/来源跳转/删除（1/1，4.4 秒）；取消回执、元数据草稿、用户数据迁移、WASI 任务、成员权限隔离五项在组合运行中通过；双版本目录安装、升级、禁用和运行（修正旧目录选择器后 1/1，4.1 秒）。组合首跑的目录选择器按旧 UI 查找 appId 和 `v1.0.0`，未触发该场景业务操作；已按当前可见版本文案和稳定 `data-app-id` 重跑通过。
- Docker 真实浏览器覆盖容器指标权限/刷新与 Compose 草稿、应用、输出、删除部署，2/2 通过（约 1.3 分钟），运行环境为本机 Docker Engine 29.4.1。任务对象与部署、卷均按用例清理；结束后只读检查未发现 `dev.blora.run` 测试容器或卷。
- 普通真实浏览器套件（排除已有独立一小时性能场景）共 48 项，`--workers=1 --trace=off --reporter=line` 首轮运行 11.4 分钟，43 项通过、5 项失败。失败分别是指标用例没有删除临时实例导致后续总数断言不稳、管理员 API 登录遇 429、以及三项恢复测试仍查旧顶栏无障碍名称；这些均由错误上下文确认是测试隔离/定位问题。
- 修正后定向重跑受影响的 6 项（包含“指标用例清理后立即进入桌面用例”的顺序验证）6/6 通过（40.8 秒）。与首轮合并，48 个常规场景各有至少一次通过结果；这不是一次 48/48 全绿运行，完整套件当前不需为同一测试问题重复 11 分钟。现有 3600 秒性能测试沿用独立实测，不在本轮重复。
- `BLORA_VERSION=development-20260920-rc13 make package` 通过，产出 Linux/Windows Master 和 Daemon、SDK 与 Web 共六个归档。六个归档逐项 `sha256sum -c SHA256SUMS` 全部通过；`SHA256SUMS` 文件 SHA-256：`bbbf742184f958166fb4673acc9517c4de2b388eca44e98bfbaa340548390991`。
- `python3 scripts/package-smoke.py dist/releases/development-20260920-rc13 --state-restore --rollback-release dist/releases/development-20260920-rc12` 退出码 0。包内 SDK/参考扩展构建和签名、Master TLS/登录、双 Daemon ONLINE、停机快照恢复，以及 rc12 兼容包读取恢复快照后的身份、权限与数据均通过。

首轮失败均保留在 Playwright 结果目录并已按上下文修复，没有把它们伪写成产品回归通过。新建的 HTTPS fixture 已停止并清理，Docker 测试对象为空；没有 Blora/Playwright 测试进程遗留。包级私有诊断目录保留在 `/tmp/blora-release-smoke-r33udswv`。

## 仍未完成

本报告增加局部真实链路证据，不提升全范围状态。验收矩阵中的 F01–F14 和 E01–E09 仍保持“进行中”；A01–A17 中已有若干 Linux 场景通过，但跨平台和综合组合仍有缺口。后续优先补齐 Windows Job Object/ConPTY 与 Windows 系统管理真机验证、独立 systemd 服务/定时器环境验证、真实远端高延迟及证书轮换/远端 Docker 场景、物理掉电和长周期调度故障验证。现有一小时本机性能场景虽已通过，最大交互延迟 945.7 ms 仍需保留在报告中；远端及 Windows 性能组合尚未验证。

环境边界：本机无 Windows 运行环境；宿主 systemd user bus 返回 `Operation not permitted`，隔离容器没有可操作的宿主 cgroup/systemd manager，特权宿主 cgroup 方案被安全审查拒绝，因此未修改宿主服务；没有独立远端测试节点或可安全模拟物理掉电的环境。相应状态仍记录为未验证/环境缺失，不能由交叉编译、本地 Docker 或普通进程模拟替代。

## 2026-09-20 真实浏览器计数复核

系统通知真实浏览器用例加入后，当前 Playwright 常规真实场景从此前报告的48项增加为49项。`BLORA_E2E_CREDENTIALS=<隔离fixture私有凭据路径> npm run test:e2e:real -- --workers=1 --trace=off --reporter=line --grep-invert='real eight-window mixed load measures pointer response with two PTYs and verified transfer'` 收集49项，首轮47/49通过。两项失败均为测试准备问题：监控用例未等待初次进程页读取，上传续传用例未使用性能fixture。监控测试补齐前置等待后1/1（54.0s）通过；上传关闭续传用正确`--performance`夹具1/1（1.2m）通过。故每项均有单独通过证据，但没有一次运行达到49/49；为避免无意义重复，没有重跑其余47项。RC13包及其内容未因此改变。
