# 冻结 UI 后的真实混合负载定位（2026-09-26）

关联 F02/F06/A06/A09/E08。当前 **E08 仍未通过**，不以功能测试或单一改动代替 ≤50ms 性能验收。用户已暂停 UI 发散，本轮没有改变配色、图标、壁纸、圆角或通透材质。

## 改动与针对性验证

- 同一版本的 xterm 6.0.0 headless parser 在专用 Worker 中镜像每次真实输出/尺寸事件并计算完整检查点；屏幕仍由原 xterm 绘制。每个视图单一在途操作，只有 Worker 解析完成、对应输出完成本地保护后才返回原 WSS ACK 路径。Worker 初始化/运行失败或超时会销毁 Worker，使用原主线程序列化；不会重放历史输入或新建 PTY。
- `@xterm/headless 6.0.0` 的已发布 ESM 文件位于 `lib-headless/xterm-headless.mjs`，包内 module 字段有误，Vite 显式解析该锁定文件；[上游问题](https://github.com/xtermjs/xterm.js/issues/6052)。Worker 环境适配仅让 xterm 的 `window.requestIdleCallback` 特性探测使用 Worker 全局，实际选择原生 timer 后备队列，不提供 DOM 或模拟终端。
- 终端恢复日志改为全部序号/预算校验后原位追加。旧代码展开 Vue 代理数组后，会把代理元素嵌进 raw 快照，导致原生克隆 `DataCloneError`，退回整份工作区 JSON 编解码。新增针对性断言在修复前真实失败，修复后通过；保留最近输出即时保护、128KiB/128事件预算及失败不确认流量规则。
- 二进制输出解码去掉逐字节 callback，单块输出不再复制一次 joined buffer，数据、顺序与64KiB解析预算保持一致。

实际结果：前端单测 **63/63**、生产构建含类型检查退出0；真实 xterm/Worker 等价性与刷新/移窗/不可用回退浏览器 **4/4（48.9s）**通过。等价性覆盖3000行历史、分块中文/emoji、UTF-8不完整边界的当前屏幕一致性、备用屏幕/ANSI/模式、尺寸变化。传输替身仅存在于该浏览器测试；parser/serializer/Worker/恢复存储均真实。它不单独证明任意 TUI 私有状态或设备掉电恢复。

## 实测记录

所有性能运行保留八窗口、双独立 PTY、10,000项目录、真实16MiB跨节点传输和两次 RAF 的指针响应测量，输出脚本每10ms约1KiB；没有降低验收阈值。宿主为 Linux/Fedora43、Ryzen7 5800H、12逻辑CPU、50,500,485,120B内存；页面没有可用硬件 WebGL，终端均为默认 renderer。

| 运行 | p95 / max | 样本 / 长帧 | 双 PTY B/s | 结果 |
| --- | --- | --- | --- | --- |
| 宿主 Firefox 冻结 UI 初始基线，60s soak | 1519 / 1973ms | 216 / 全部长帧 | 约97KB/s，各流真实持续 | 失败 |
| 宿主 Firefox，仅测试侧关闭 backdrop-filter、尚未迁移 Worker | 1172 / 2452ms | 336 | 约58/60KB/s | 失败，不进入产品 |
| 宿主 Firefox，Worker 启动修正后 | 1227 / 1545ms | 240 / 247长帧 | 107069.6 / 106961.8 | 失败；6轮已验证传输，队列峰值86261B |
| 宿主 Firefox，Worker + raw日志数组修正 | 1212 / 1280ms | 239 / 248长帧 | 108794.5 / 108958.5 | 失败；6轮已验证传输，队列峰值87234B |
| 官方容器 Firefox155.0，同一当前产品 | 1230 / 1325ms | 239 / 247长帧 | 110847.1 / 110956.7 | 失败；6轮已验证传输，队列峰值87098B |
| 官方容器 Firefox，当前产品仅测试侧关闭 backdrop-filter | 297 / 359ms | 648 / 649长帧（908帧） | 111285.1 / 110979.4 | 失败；8轮已验证传输，队列峰值86052B；不进入产品 |
| 官方容器 Chromium153.0.8010.12，同一当前产品 | 1240.1 / 1517.2ms | 240 / 250长帧（262帧） | 109475.6 / 109363.1 | 失败；6轮已验证传输，队列峰值80370B；SwiftShader软件合成 |
| 官方容器 Chromium，仅测试侧 will-change/终端绘制包含 | 1104.7 / 1604.5ms | 240 / 250长帧（264帧） | 113063.4 / 113116.1 | 失败；6轮已验证传输，队列峰值55036B；不进入产品 |
| 官方容器 WebKit26.6，同一当前产品 | 739 / 1065ms | 1752 / 204长帧（210帧） | 101713.5 / 102393.7 | 失败；6轮已验证传输，队列峰值86010B；两终端webgl/Worker |
| 官方容器 Chromium，仅测试侧去模糊并启用诊断/CPU采样 | 58 / 126.1ms | 1296 / 27长帧（3174帧） | 107670.0 / 107759.0 | 失败；8轮已验证传输，队列峰值78981B；非正式基线 |

“60s soak”是循环声明的最短持续时间，指针操作完成后才检查截止时间，实际采样更长。例如当前官方Firefox基线真实采样132.114s，去材质对照76.319s。两条流持续、传输目标核对、错误/队列断言均正常；性能阈值明确失败。官方容器与宿主结果接近，不能把宿主 ABI 当作唯一原因。去材质后改善说明合成是主要成本之一，但该对照仍未达到50ms，不代表删掉材质就完成验收。

原生 GeckoMain profile 在迁移前主要热点是主线程完整终端序列化；迁移后该热点消失，IndexedDB put、克隆/JSON及绘制仍有成本。空白可见页面的宿主Firefox帧时钟实测180帧/3s、p50 16.66ms、p95 17.4ms。原始profile含浏览器内部URL，保持0600私有，仅提取函数名/计数用于分析，不作为可公开产物。

Chromium去模糊诊断路径 `/tmp/blora-perf-chromium-no-glass-profile-20260926/results`：真实采样65.928s，同步tail共3292次写入，p95 1.5ms、max8.3ms，仅一条72ms主线程longtask；CPU采样仍见IDB put/克隆、DOM行替换与渲染开销。这一测试侧对照不能当作当前材质的通过证据，亦不把模糊关闭或改变测量阈值作为产品修复。

## 入口、证据与清理

宿主入口：

```sh
./dist/blora-devfixture --performance --listen 127.0.0.1:9443 --state-parent /tmp
# 在web中，仅传私有文件路径，不输出其中凭据。
BLORA_E2E_CREDENTIALS=<本次fixture>/browser-credentials.json \
BLORA_BROWSER=firefox BLORA_PERF_SOAK_SECONDS=60 \
npx playwright test --config playwright.real.config.ts tests/real/performance.spec.ts \
  --reporter=line --output=<本次独立结果目录>
```

官方容器版本为已安装 `mcr.microsoft.com/playwright:v1.63.0-noble`，镜像摘要 `eff16c30e6f3f4af0a03fa4b706120d5e9b0891c344a27d64559aff5900a4a27`。每次使用 `--network none`，Master/双Daemon/浏览器均在同一容器私有loopback内运行；项目只读，仅挂载对应独立输出目录，不挂载宿主PID或Docker socket。脚本在EXIT只向其直接启动的fixture发送SIGINT并等待清理，容器以`--rm`退出。

宿主结果目录：`/tmp/blora-perf-firefox-baseline-20260926`、`/tmp/blora-perf-firefox-worker-fixed-20260926`、`/tmp/blora-perf-firefox-native-copy-20260926`。官方容器结果：`/tmp/blora-perf-firefox-container-20260926/results`、`/tmp/blora-perf-firefox-no-glass-worker-20260926/results`。诊断开关 `BLORA_PERF_RENDER_EXPERIMENT` 仅存在测试侧，不改变默认产品。

Chromium/WebKit 当前完整材质的严格基线亦失败；现有RC15/9月22日性能数据只对应各自代码时点，不覆盖本轮冻结 UI。Chromium 分层对照没有足够收益，不进入生产样式。对应目录为 `/tmp/blora-perf-chromium-current-20260926/results`、`/tmp/blora-perf-chromium-layers-20260926/results`、`/tmp/blora-perf-webkit-current-20260926/results`；三容器与所属fixture均自动退出清理。WebKit的WebGL字符串“Apple GPU”来自Linux移植引擎，不能据此声称该机器为Apple硬件或有实际GPU加速。

仍需继续定位保持既定视觉的合成/持久化开销；Windows、硬件加速设备、远端网络和完整跨平台组合保持未验证。性能失败不阻止完成独立的功能回归和发行验证，也不以那些结果覆盖E08。

## 2026-09-27 继续定位

新增仅测试侧 `glass-layer` 对照，把相同 blur/saturate/透色表面放在窗口独立伪元素中，保持真实负载、流控与≤50ms断言。官方Chromium153/SwiftShader实跑125.767s、263指针样本，p95 **1003.8ms**、max1446.8ms、272/297长帧，两个PTY分别115308.8/115315.7B/s、3轮真实传输完成且目标校验正确、未确认队列峰值66291B；命令164.670s、退出1。运行时另一个独立浏览器正在做低频监控历史/失联功能回归，因此此项只是定位记录，不能据微小差值推断正式收益，更没有达到阈值。当前布局或材质实现未替换为该对照；不再重复这个无足够收益的方向。私有证据位于 `.local/evidence/rc24/perf-glass-layer`，隔离容器 `--network none` 无宿主PID/Engine挂载，所属fixture和容器已退出清理；独立24h监督不受影响。

进一步仅测试侧通过 `BLORA_PERF_SOFTWARE_COMPOSITING=1` 添加Chromium `--disable-gpu` 启动参数，当前产品/完整材质不变，且不与其他功能浏览器并发。命令02:44:47～02:47:37 UTC、169.799s、退出1；真实采样129.071s、264样本，p95 **1005ms**、max1516.1ms、272/291长帧，双PTY116187/116274B/s、4轮目标验证通过，队列61998B、heap72.2MB。最后探测的WebGL字符串仍为SwiftShader，因此**没有证明得到独立的CPU合成后端，也不能称为硬件/纯CPU比较或正式通过**。≤50ms断言仍失败，无产品改动；私有证据 `.local/evidence/rc24/perf-cpu-compositing` 和对应日志，隔离容器及所属fixture已自动退出。

新增不含后台/终端/恢复存储的微型绘制探针，只画八个同尺寸24px模糊窗口和移动最上层，使用同官方Chromium153的有窗口Xvfb环境比较已有后端。五秒RAF样本：默认SwiftShader17样本/p95 400ms；明确OpenGL后的Mesa llvmpipe（LLVM20.1.2）41样本/p95 433.3ms/max1533.3ms；Vulkan参数仍SwiftShader21样本/p95 383.3ms。它没有完整真实混合负载、两次RAF输入测量或100样本，**仅定位合成成本，不属于E08验收**；即便没有终端/持久化活动，本机软件玻璃合成仍明显耗时。宿主没有 `/dev/dri`，没有验证硬件加速。Xvfb包装器未进入Node，首个无认证直接调用也未启动浏览器；使用该隔离显示的认证路径后实际探针02:59:36～03:00:14 UTC执行37.830s、退出0（诊断程序结束，不代表性能通过）。证据 `.local/evidence/rc24/renderer-probe-direct.log`；随后仅停止该唯一命名的容器，包装器退出137，未计作通过。没有修改或删去当前通透材质、没有采用较差的OpenGL对照。
