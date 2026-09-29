# 本地终端与绘制诊断补验（2026-09-29）

起点 `16f2262`。本轮只修改测试和诊断入口；产品源码、UI、完整通透效果、发行包及≤50ms标准不变。RC25仍为当前发行产物，不能将本报告解释为全范围完成。

## 终端应答边界

真实Vim/top刷新用例在新隔离双节点fixture连续10/10通过，107.582秒，退出0，证据 `.local/evidence/rc26/vim-diagnosis.log`。未复现RC25的额外字节，不能宣布根因已解决。

原用例继续逐字节比较刷新前后的全部输入，未过滤查询应答或重试用户命令。增加有界的收发诊断：连接编号、单调时间、Vim/top阶段、DATA/RESUME消息及十六进制载荷；仅比较失败时附到测试结果，避免普通终端日志解释ESC后只剩`$y`。原始输出属于隔离测试数据，不进入公开文档。

终端定向浏览器回归 **6/6通过（77.347秒，退出0）**，覆盖新查询边界、Worker屏幕一致性、运行中Worker真实终止回退，以及自动渲染/默认渲染/主线程检查点三种恢复组合。日志 `terminal-browser-final.log`。初次类型检查发现测试事件重载与CDP协议泛型类型不匹配，已改为明确的收发监听和trace事件结构；不是产品编译回归。

新增真实xterm/Worker/IndexedDB边界测试：服务端历史查询、保存的原始输出日志重挂及恢复完成均零应答；恢复后的新DECRQM查询精确回复`ESC[?2004;2$y`；只读期间零应答，重新取得输入权限后正常回复。最初测试未建立屏幕基线，首批输出直接成为屏幕检查点，日志存在性断言3/3失败；补上真实基线后连续3/3通过（12.499秒）。该修改是补正测试前置条件，不放宽产品断言。首轮日志 `query-boundary.log`，复验 `query-boundary-fixed.log`。

## 软件合成诊断

新增可选 `BLORA_PERF_RENDER_TRACE=1`，在真实八窗口负载的首轮拖动记录Chromium绘制线程计时；只输出线程/事件/耗时汇总和图形功能状态，不保存trace参数、URL或终端正文。默认关闭；带跟踪运行是诊断，不能替代未加跟踪的E08测量。线程间及嵌套事件耗时重叠，不得相加当成墙钟时间。

初次跟踪跨度59.070秒，152,447条完整耗时事件，聚合未丢条目，trace缓冲峰值10.99%。主要结果：

| 线程/事件 | 累计包含耗时 | 单次最大 |
| --- | ---: | ---: |
| VizCompositorThread / SoftwareRenderer::DoDrawQuad | 57.782秒 | 99.792ms |
| VizCompositorThread / Display::DrawAndSwap | 58.157秒 | 641.007ms |
| CrRendererMain / FunctionCall | 2.651秒 | 9.701ms |

这是软件合成成为主要瓶颈的直接证据，不能仅凭WebGL的SwiftShader字符串推断页面正在使用GPU合成。初次命令没有开启持续传输soak，传输在采样末尾已完成，因此原`stateAfter === RUNNING`断言失败；p95 1125.2ms也不满足50ms。该次不是通过的E08，失败日志 `render-trace.log` 原样保留。

进一步在相同八玻璃窗口的5秒微探针比较三种启动方式（与完整产品验收不同）：

| 模式 | 实际gpu_compositing | 样本数 | p95 |
| --- | --- | ---: | ---: |
| 默认 | disabled_software | 15 | 433.4ms |
| 显式SwiftShader合成 | enabled | 12 | 1416.6ms |
| 显式SwiftShader合成及GPU光栅化 | enabled | 15 | 783.3ms |

显式启动参数确实改变了合成路径，但本机微探针更慢，未加入产品或默认测试配置。日志 `.local/evidence/rc26/compositor-backends.log`，19.610秒，退出0仅代表探针正常执行，不代表性能达标。本机仍没有可用硬件加速设备。

最终增加60秒循环传输的诊断复核在155.528秒后按原≤50ms断言失败，日志 `render-trace-final.log`：240样本p95 **1062.5ms**；两个PTY各约111.5KB/s，队列峰值81,801字节，采样末尾任务仍RUNNING，循环传输6次，最终16MiB目标校验通过，页面错误为零。真实图形状态再次明确`gpu_compositing=disabled_software`。此次完整负载条件满足，失败点仍是交互延迟；启用跟踪和系统Chromium151会影响可比性，不将此值与官方Chromium153的RC25基线作优化幅度比较。

## 复现入口与剩余边界

最终收尾：包含新附件监听的真实Vim/top用例再连续 **3/3通过（36.582秒）**，日志 `vim-final.log`；最终`npm run check`退出0（16.248秒），日志 `typecheck-complete.log`。所属fixture、浏览器和探针全部退出，未启动生产服务。本轮无产品二进制变更，因此未重复打包或给RC25产物重新编号。

```sh
cd web
npx playwright test tests/browser/terminal-snapshot.spec.ts tests/browser/terminal.spec.ts --workers=1
# 先启动独立 devfixture --performance，凭据路径只放环境变量。
npx playwright test --config playwright.real.config.ts --grep 'actual vim unsaved' --repeat-each=10
BLORA_PERF_RENDER_TRACE=1 BLORA_PERF_SOAK_SECONDS=60 npx playwright test --config playwright.real.config.ts tests/real/performance.spec.ts
```

不要将mock浏览器与实际Docker网络创建/销毁并行，以免已知的`ERR_NETWORK_CHANGED`干扰。完整通透材质的正式E08仍采用RC25三引擎1003.9/1108/545ms失败结论；本报告不覆盖它。真实Vim偶发额外字节仍待复现并判定连接/回放边界；Windows、独立systemd、跨主机及设备故障等缺口保持原状态。
