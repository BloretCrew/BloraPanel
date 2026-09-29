# 本地性能与发行收尾（2026-09-29）

起点为 `658c521`。用户确认完整保留静止、拖动及缩放时的通透效果；本轮不调整配色、图标、材质参数或严格性能阈值。**当前E08仍未通过**，下面阶段结果不能作为全范围完成声明。

## 实现与恢复边界

恢复服务原先先 `structuredClone` 完整工作区，再交给IndexedDB做一次结构化克隆。候选改为读取pointer后、同步 `put` 前绑定revision，由原生 `put` 捕获快照，省去前一次复制。同步DataCloneError仍走原JSON清理回退，异步请求失败不重试；快照、pointer及旧版本清理仍在同一事务中，最新同步日志与ACK顺序保留。规范依据为 [IndexedDB add/put](https://w3c.github.io/IndexedDB/#dom-idbobjectstore-put)，锁定idb 8.0.3源码亦确认同步错误原样抛出。

已聚焦、位于层级末尾且未最小化的窗口再次聚焦时不生成恢复事务。三个条件缺一时仍更新，保留最小化恢复和不一致层级修正。

直接DOM拖动在9月21日已有两次无稳定收益的记录，本轮不重复采用；关闭模糊、降低输出和软件后端切换均不作为产品修复。

## 当前验证

- 新增快照写入后立即出现下一次编辑：当前/前一快照的revision与正文保持一致，独立恢复得到最后版本；嵌套Vue代理值仍可恢复。恢复单测16/16、完整前端单测66/66通过，含原有配额失败、异步快照期间新输入与日志预算边界。
- 原生IndexedDB同一新增场景在官方Playwright v1.63.0的Chromium153、Firefox155、WebKit26.6各1/1通过；并发编辑、前一版本、无同步tail恢复与代理回退均通过。日志 `.local/evidence/rc25/native-snapshot-engines.log`。
- 焦点回归在修复前实际失败，观察到100次额外同步存储写入；修复后100次重复聚焦为0次写入，其他窗口聚焦、最小化恢复和层级校准通过。
- 类型检查与生产构建通过。未修改公共API、协议或持久化schema。

首轮完整mock与真实Docker套件同时执行，mock为54/59、退出1（355.586s）。四个失败场景的trace记录静态模块请求 `ERR_NETWORK_CHANGED`，包括整页空白及刷新后应用仍加载；与真实Docker创建/清理网络同期，不能据此认定工作区丢失，也不能抹掉失败。另一个参考扩展开窗未出现第二个iframe，单独复核。后续完整mock改为与Docker生命周期串行，保留原超时、原生鼠标与断言；原trace/截图保存在 `.local/evidence/rc25/mock-first-artifacts/`。

单进程定向6项复核中加载/恢复相关5项通过；参考包第二次在刷新后取消按钮未触发成功状态。该测试遗漏已有的 `clickPaintedExtensionButton`：现补齐iframe按钮的原生hover命中等待后再单次click，未增加动作重试、合成点击或放宽断言。相同完整参考包场景连续3/3通过（61.350s，退出0）。Docker套件停止后，独立完整mock **59/59通过（301.162s，退出0）**，日志 `mock-isolated.log`，包括拖动期间材质不降级/即时刷新/取消还原及Worker运行中终止回退。Firefox与WebKit真实HTTPS恢复故障各3/3通过，覆盖配额错误后导出/恢复、旧schema迁移/不支持版本保留、真实IDB事务abort后原子回滚；属于故障注入，不能当作设备掉电验证。

完整真实HTTPS/Docker套件51项首轮 **50通过、1失败**（844.294s，退出1，无skip）：Vim刷新时严格输入字节比对多出一次`$y`尾部。锁定xterm源码的DECRPM应答确实具有此后缀，但仅凭尾部不能证明来源，未据此过滤字节或放宽断言。临时记录协议收发方向的新夹具重复3/3通过，没有再捕获失败；移除诊断后，原始Vim/top用例在另一新夹具连续3/3通过（46.573s，退出0）。这属于分段复验，**不能写为完整套件51/51一次通过，也不能声称该偶发现象根因已修复**。命令为 `bash /tmp/blora-rc25-real.sh --grep-invert 'real eight-window mixed load'` 和 `--grep 'actual vim unsaved' --repeat-each=3`，私有日志 `real-final.log`、`vim-diagnostic.log`、`vim-original.log`。原用例和终端产品代码均未修改，输入不重放断言保留。

实际套件包含双节点权限、Docker/Compose、WASI参考包、文件冲突/撤权、PTY及标签移窗、跨节点传输、121点监控及同库Master重启、节点失联、通知、存储故障、两个真实分钟调度slot、关闭浏览器后万文件解压与16MiB上传续传、工作区修订冲突与幂等创建。本次监控采样间隔为1000ms，不能等同于此前5000ms或24h场景。测试包装器仅停止所属fixture及其资源；生产服务未操作。

## 三组原生快照对照

同一Chromium进程使用原服务和候选服务，交换执行顺序；每组各35次提交，前5次预热，剩余30次统计。固定大型正文及保存基线，提交相同的微小偏好变更；每次都等待真实IndexedDB完成并检查保护状态。该专项不是E08混合负载。

| 组 | 原版中位 / p95（ms） | 候选中位 / p95（ms） | 额外JS工作区复制次数（原/新） |
| --- | --- | --- | --- |
| 1 | 85.9 / 121.9 | 55.7 / 73.9 | 35 / 0 |
| 2（逆序） | 76.2 / 93.9 | 35.4 / 46.8 | 35 / 0 |
| 3 | 85.4 / 96.6 | 37.8 / 55.5 | 35 / 0 |

日志 `.local/evidence/rc25/snapshot-bench.log`，私有复现入口 `web/.local/recovery-bench.spec.ts`，原服务从起点Git读取。快照专项中位耗时改善35%～56%；不外推为整页交互的同等改善。

## 严格混合负载

本机Ryzen7 5800H、12逻辑CPU，Chromium使用SwiftShader，没有可用`/dev/dri`。每次独立容器、私有loopback、真实八窗口/双PTY/万项目录/16MiB循环传输，60秒最短soak，测量与≤50ms断言保持原样；性能运行串行且不同时运行功能浏览器。

| 浏览器 | 原版 p95 / 样本 | 最终候选 p95 / 样本 | ≤50ms |
| --- | --- | --- | --- |
| Chromium 153 | 1016.6ms / 239 | 1003.9ms / 263 | 失败 |
| Firefox 155 | 1080ms / 240 | 1108ms / 239 | 失败 |
| WebKit 26.6 | 516ms / 2256 | 545ms / 2304 | 失败 |

六次测试均输出完整指标后在原≤50ms断言退出1。双PTY约106～116KB/s，队列峰值均低于87KB，末尾状态RUNNING，16MiB目标内容校验通过，各完成2～3轮循环传输。WebKit返回的脱敏WebGL名称“Apple GPU”不代表本机存在该硬件。本次Chromium仅恢复候选p95 984.3ms（239样本）亦失败；整页数据没有稳定改善证据。不能用恢复快照专项收益覆盖E08，亦不启动以短测通过为前提的一小时复验。

## RC25发行与独立恢复

```sh
GOMAXPROCS=4 BLORA_VERSION=development-20260929-rc25 make package
python3 scripts/package-smoke.py dist/releases/development-20260929-rc25 \
  --state-restore --rollback-release dist/releases/development-20260927-rc24
```

六包构建开始03:14:53 UTC、结束03:15:48 UTC，54.618s、退出0，包含Linux/Windows Master与Daemon、Web、SDK；类型检查、前端构建、SDK/参考包、106项第三方通知收集通过。归档验证4.827s、退出0：六个外部SHA256、**2,207条内部路径/长度/模式/摘要**、路径安全与重复检查、两个Master包和独立Web包共三份Web精确匹配当前构建。SHA256SUMS自身摘要：`08a8f1a22ef281216f4e88a4f812350c052f175988e43b4106206290102faa71`。产物在 `dist/releases/development-20260929-rc25`，依照现有策略不纳入源码Git。

独立包验证开始03:16:53 UTC、结束03:17:04 UTC，11.063s、退出0。在专属目录解压SDK并独立构建/签名，Master初始化/TLS/static/login及两个打包Daemon ONLINE均通过。停止所属服务形成快照、制造后续修改、恢复后核对TLS、两节点身份、用户登录、只读授权、原任务回执与资源内容；后续修改消失。旧RC24打包Master/Daemon实际打开同一兼容快照，保持身份与权限；不承诺任意schema降级。所属进程均停止，生产服务未触碰。

逐命令起止UTC、耗时、退出码位于 `.local/evidence/rc25/*.log.result.json`；私有证据不纳入Git。既有24h长测已通过，见[终态报告](local-endurance-24h-2026-09-27.md)。当前剩余：E08严格完整材质性能仍失败；本次Vim偶发额外字节虽六次复核未再现，来源尚未定位。Windows真机、独立systemd、跨主机网络/Engine/注册源、设备掉电及原生通知中心仍缺对应环境证据。只有上述问题及环境证据补齐后，才能将全范围标为完成。
