# Historical snapshot before the non-release closeout (2026-10-08)

This preserves the full prior record. Relative Markdown links were adjusted for this archive directory. Current status is in the parent document; pending items below are historical.

# Blora Panel 全范围验收矩阵

## 2026-10-08 性能优化收尾（当前状态）

本段覆盖下方所有历史接管中的“正在运行 / 下一步 / R1开放”等措辞。用户明确接受罕见极端压力下约100ms，要求只再做少量尝试后收尾；本轮已完成，不继续追求全浏览器严格50ms。原50ms目标与失败记录仍完整保留，不写成全部通过。

当前产品同源正式复核：Chromium 38.8ms PASS、Firefox 66ms FAIL、WebKit 132ms FAIL及独立复测95ms FAIL（均为原始首120事件，FAIL只在未改动的50ms断言）。同源较早WebKit为89ms；因此记录89–132ms波动，不能承诺稳定≤100ms。八窗口、双真实PTY、万条目录与最终16MiB目的校验保持。最后只复测了一个已有flat-body候选，前次79ms未复现、此次139ms，无可靠收益而拒绝；未引入新产品实现。

保留545源4c8db790/169bundle5a04e608；生产构建、121单位、WebKit21项原生像素/几何/输入/保护ACK/恢复检查、最终类型检查和空白检查通过。新增Chromium DPR1静态圆角截图检查仍不稳定，旧定位同处复现，失败保留，不伪称修复或通过。1110部分body最终149ms、1112原transform对照109ms和1114自动行跳过0均未进产品。

本轮五个fresh夹具均确认清理退出0，浏览器测试均结束，无运行中的本任务服务。R1原严格50ms改为后续性能改进，不再作为本轮继续工作的阻塞项；R2/R3/R4历史状态不变，R5发行仍按用户要求延期，未提交或推送。最终数字、验证范围及限制见 [收尾报告](../reports/e08-bounded-closeout-2026-10-08.md)。

## 2026-10-08 历史接管（运行状态已被收尾记录覆盖）

2026-10-08 11:28 UTC 最新接管覆盖以下历史运行措辞：产品仍545源4c8db790cc97dcd6f1ba69b0d8007e40b0044f781cde9c9c9e87674c5ca1de9f/169bundle5a04e608c1958244f78594fc7539be40a95b19b8b5ff98cac486091c0021d1d3；1078构建/121单位通过、1080 WebKit21/21守卫通过。实际Linux WebKit为default/default与8offset，正式1082原始首120 89ms FAIL。1086软件frame无额外层仍89ms FAIL；1092隐藏正文80ms、1094去全部阴影111ms均是改变外观的归因对照，未采用；各目的校验/退出/夹具清理已收。1096 app-view保留层实色74通道/max1失败；1098限定opaque glass两个DPR十场景RGBA0通过，1100原始仍85ms FAIL（nativeRetentionGlassOnly=true、retainedAppViewLayers=8，PTY120931/115534B/s，max未ACK75165，最终16MiB11.389s），72113退出1/38282清理0，收益不足未采用。1102 flat-body原生clip两个DPR十六场景RGBA0通过，1104原始79ms FAIL（PTY157295/150685B/s，max未ACK76157，16MiB11.372s），41167退出1、59247已确认清理退出0，未采用。1106新部分body候选初始场景没有真实partial区，正分支激活断言失败保留；1108改用明确部分/完全覆盖场景，两DPR十八全屏RGBA0、文字/滚动/八窗held/实色fallback通过，完全覆盖1body仍保持原inset100%。1110此候选原始WebKit正在运行7778，fresh32990/fixture655150180；仅测试helper，产品未改；必须收最终目的校验与退出，再清理fixture。另1088/1090独立空白/flat可信120事件调度通常29–32ms（WK冷首轮65），诊断不能算E08。FF正式1058 50ms PASS/1062复测52ms FAIL非稳定；Chromium旧定位同处静态圆角不稳定，8次稳定/RGBA0不放宽。R1/E08开放、R2/R3/R4闭合，Windows不重复、发行延期、无提交/推送。

2026-10-08 08:58 UTC 本段覆盖下方历史运行措辞：0994原始Firefox120事件原生可见行归因诊断 **65ms FAIL**（max70、1/128长帧、双PTY117154/118082B/s、max未ACK86167、3.540s、最终16MiB目的校验）。25679退出1/18240夹具清理0；真实viewport只有27×7/27×33，行替换595→1075/636→1426，省略确实存在但未降低拖动延迟。产品可见行私有adapter和整数DPR64边层候选均撤回；更完整原生行控制仅留tests/helpers，原像素/同步输出/恢复断言与失败报告保留。当前恢复64px reserve/分数DPR完整原生回退，尚需重建，当前没有实际压力/夹具。下一独立验证仅窗口壁纸材质的物理像素RGB24预展开，避免以往PNG全分辨率实验的透明合成；不读取/缓存应用文字、图形、内容或数据、不改负载/50ms/首120/ACK。R1仍开放，R2/R3/R4闭合、Windows不重复、发行延期、无提交/推送。 0997恢复构建退出0，当前543源8d22bede/169bundle587b8154；0998 FF两DPR24光学条件2/2（31.1s）全部RGBA0，1000 WK2/2（47.2s）整数0/分数max1且mean≤.01、1002 Chromium2/2（26.0s）通过，RGB24与独立native PNG编码全部0。前次0225全分辨率诊断PNG colourType6，故此次RGB24是不透明合成的新归因对照；产品不接入新路径。功能测试全部停止，fresh夹具26586启动，下一1004原始Firefox首120/真实资源归因诊断。 1004 native-rgb24原始Firefox归因诊断120事件56ms FAIL（max67、2/126长帧、双PTY117416/120335B/s、max未ACK86410、3.095s、最终目的校验），24064退出1/26586夹具清理0；不采用为最终产品/不声称单轮收益。下一test-only直接原生frame背景结构：同一共享材质直接由原frame绘制，撤掉两块仅壁纸伪元素，无应用采样/节点移动/尺寸变化；fractional保留完整原结构。1007首次测试在应用恢复未完成即调用store.open，2FAIL未进入像素，保留证据；增加原topbar/material ready前置、不改断言后1009 Firefox两DPR运行17211，类型60313退出0。实际压力和夹具均已停止。 1009直接frame background原生参考整数DPR仅622通道/max6圆角边缘差异、fractional原结构0；保留FAIL且不放宽像素。更换为同一原header伪元素覆盖整个已知opaque frame/原圆角，正文伪元素撤掉；不改frame底色/边框/应用DOM。1011 FF2/2（21.8s）和1013 WK2/2（33.6s）两DPR八场景全部全屏RGBA0，实际八窗口均启用，fractional完整原回退。1015 Chromium运行35078、类型19468；当前无实际压力或夹具。新结构仅测试对照，产品543源8d22bede/169bundle587b8154保持。下一全部功能退出后fresh原始负载，原首120/50ms/所有资源保持。 1015 Chromium新结构整数DPR8484通道/max105 FAIL、fractional0，不采用该引擎新结构（其原始34.4ms已达标）；保留该失败。nativeOwnership fallback保持Chromium及fractional完整旧原生结构，全部原全屏RGBA0/几何/身份断言仍执行，1017两DPR2/2（26.7s）通过，类型19468退出0。FF/WK整数路径不变且原先已全部严格0。当前所有功能停止，下一fresh Firefox原始120单原生plane归因压力（BLORA_PERF_RENDER_EXPERIMENT=native-frame-material），不得称产品通过；产品仍不接入，仅测试helper。 1019 fresh原始单plane Firefox120事件61ms FAIL（max74、3/126长帧、双PTY118382/117007B/s、max未ACK66699、sample3.376s、八native frame实际启用、最终16MiB目的校验），24390退出1；30285停止请求返回运行中，随后确认退出0，fresh65654已ready（3354292383），未并行压力。1021只增加native-rgb24预展开组合归因诊断进行中；同源同产物、原轨迹/资源/断言保持，尚未采用产品且不称性能收益。 1021组合fresh原始Firefox120事件67ms FAIL（max103、2/127长帧、双PTY116521/115807B/s、max未ACK86167、sample3.688s、最终目的校验），85084退出1/65654夹具清理0；未采用。1024新原生精确保守区域结构首次FF1PASS/1FAIL：去掉正文原不透明backing后仅原SVG图形[132,284,147,365]129通道/max5差异，fractional原回退0；保留原严格像素断言和失败报告。1026保留正文原生不透明backing、由完整仅壁纸frame plane填补区域接缝的独立候选FF运行中；类型88206退出0。当前没有实际压力或夹具，新结构仍仅测试helper，产品/构建保持不变。

当前候选同时优化数据恢复和终端传输：桌面启用数据专用快照Worker，主线程即时sessionStorage日志仍立即保护每次输入；原生IDB原子事务确认后才裁剪日志。64ms仅批量唤醒数据库写入，不延迟PTY解析/ACK；停止输出后500ms空闲收敛完整Workspace。原有foreground兼容回退、至多四记录/单完整基线/64条128KiB终端增量、嵌套代理和直接应用元数据、并发提交、事务失败闭锁及恢复身份验证保持。应用DOM、字形、图标和PTY绘制没有进入自定义位图缓存。

新终端events-v1按现有75ms归档读取返回的原事件打包，每帧至多128事件/64KiB，不增加等待计时器、不削减输出、不改cursor、流量额度、授权/租约/重放或保护后ACK。未协商的客户端保留旧单事件格式。前端解包后仍走原有有序解析、UTF8/ANSI/resize和检查点链路。

0444 Worker未批量写入正式90秒58ms FAIL；0448批量Worker正式90秒52ms FAIL；0452同源正式300秒52ms FAIL（8520事件，首120/57、steady8400/52，22/15266长帧，23次16MiB目的校验，双PTY119213/119165B/s，max未ACK73320）。三个夹具均清理0，不能将单项线程减负当作E08达标。0501真实PTY分组/恢复/旧租约授权撤销Go检查通过；0503浏览器8PASS/3FAIL保留，发现输出停止后原生数据库仍是delta格式，500ms空闲完整收敛修复后0507 Firefox11/11通过（Worker真实abort/即时日志/数据库-only/原终端三模式/分组UTF8+resize）。0507单位113/113通过；0507构建两个新测试隐式any失败保留，显式类型修复后0510生产构建通过。

当前543源4dcc843cdc99f64e3a2656701f2d20f748bb60467539c8edd9b16208844b7dc3；169 bundle7689f722e8e67fa43b572ecf6f51c2f84809a6e7dcad44548c8c7cc573259301；实际Master二进制ce98b304f1ff3d3c0cf2d3ceb51fcc25398f5e5052287a2da09ab38b528ef22d。0512正式Firefox90秒bounded最终52ms FAIL（2720事件，首120/56、steady2600/51，4/4730长帧，7次16MiB目的校验，双PTY120467/120384B/s，max未ACK86316；实际worker、1633Data消息；同步journal2566写/29615415字符/68ms）。89896退出1/76753清理0。0517原生线程诊断54ms不是验收，72058退出1/27779清理0；profile68717058B仅数值/静态线程分类，部分线程buffer保留最后约32秒（不能解释首120），内容RefreshDriverTick p95 13.787ms、Styles2.007ms，Renderer CPU raw20045718081，主内容线程14440210826。应用仍全部原生绘制。

原生终端行CSS布局隔离仅在tests/experiments，未进入产品。0523选区叠放失败保留；选择实际显示时完整原生回退，0528 Firefox两DPR严格RGBA0/scroll/resize/同步输出通过。0526FF及0528WK held差异保留；WK完全原生对照0530两DPR通过，候选不能向WK推广。0531入口候选名白名单在开始负载前失败，无样本，34729退出1/27018清理0；已更正。0533 Firefox90秒bounded独立行隔离诊断51ms FAIL（2720事件，首120/57、steady2600/51，4/4739长帧，8次16MiB目的校验，双PTY120117/120306B/s，max未ACK60985；91293退出1/8631清理0）。仅1ms变化且WK严格像素未通过，未采用。

静态空白阴影预合成仅在tests/helpers，绝无应用/文字/字形/图标捕获。0544最大12自身边框失败保留；限定边框外两像素后0554FF/0557WK两DPR各2/2，按分区扩展0563FF/0566WK各2/2（内容0、光学max0/1）。0560/0568真实入口eligible0，采样前拒绝、无性能样本；60536/57827退出1，97620/52159清理0。0571真实数值确认DPR1/opaque true/8frame/64tile全部有效，但重叠阴影排除后fragments0；58204退出1/87620清理0。加入下层原空白阴影按z序预合成后0576FF2/2全屏RGBA0，0579WK2/2光学max4、内容0，含八窗/拖动/两DPR。0582真实90秒bounded片段方案63ms FAIL（2395事件，首120/69、steady2275/62，31/4251长帧，7目的校验，双PTY116820/116993B/s，max未ACK81950，实际71区域179580像素）；81143退出1/86938清理0，不采用。

静态单张壁纸scene原型0590FF两DPR2/2、六场景全屏RGBA0；0593WK1PASS/1FAIL八窗contentMax1，不推广。0596实际Firefox90秒bounded67ms FAIL（2320事件，首120/71、steady2200/66，28/4009长帧，7目的校验，双PTY118137/117842B/s，max未ACK69345）；88013退出1/59150清理0。简化单矩形hull、无新增fragment节点，0602FF/0605WK两DPR各2/2，FF0/WK光学max2、内容0，但0608实际90秒57ms FAIL（2545事件，首120/65、steady2425/56，11/4504长帧，4目的校验，双PTY111990/112608B/s，max未ACK85781）；74694退出1/62434清理0。两个均未采用，全部失败证据保留，不改产品图像质量/负载/50ms要求。

连续手势原生包络150ms候选0613构建/Firefox36PASS（Chromium专用IME在FF正常skip）；500ms候选0618构建/113单位/17项原守卫通过，但0622正式Firefox90秒bounded最终58ms FAIL（2495事件，首120/65、steady2375/57，13/4421长帧，4目的校验，双PTY115929/115949B/s，max未ACK76290；84649退出1/45663夹具清理0），无收益已完整撤回。0627回退构建通过，当前543源恢复4dcc843cdc99f64e3a2656701f2d20f748bb60467539c8edd9b16208844b7dc3，新169bundle1bf460d4d59e8307ea9b2acbbfc5b05e41ce808f6982704402de15c70c277f4c；Worker和events-v1仍保留，不能把旧产物与新产物混算。只读五秒主机汇总CPU busy13.92%、CPU等待avg10 .03%、约9.3GiB可用内存、内存等待0，无当前资源拥堵证据；没有触碰其他服务。当前无真实压力/夹具。下一独立测试原生矩形正文裁切的结构候选，应用/字形/图标全原生、原像素与50ms/负载要求不变，功能通过后才fresh实际收益验证。

原生正文矩形裁切只在tests/helpers/native-body-clip.ts独立诊断：沿用现有Vue节点/原尺寸/完整应用DOM，不生成应用位图，body圆角和overflow保持，旧父级inset转换为子视图及独立壁纸面的原生clip:rect；非继承局部变量不扩散进应用，实色/非整数DPR保留原路径。0644窗口未挂载、0647实色缓存前置、0650 inset规范化导致只7/8实际启用的失败均保留；修正前置/简写后0653FF两DPR2/2全屏RGBA0，0656WK实色13通道/max1失败保留，实色完整原生回退后0659WK两DPR2/2，各八场景、节点身份/尺寸/scroll/八窗held通过。0702真实Firefox90秒bounded最终54ms FAIL（2645事件，首120/62、steady2525/53，11/4645长帧，8目的校验，双PTY119544/119383B/s，max未ACK74879，实际八窗/七视图mask）；20266退出1/73828夹具清理0。没有性能收益，不采用；当前无真实压力/夹具。

原生双轴WAAPI平移只在tests/helpers/native-motion.ts诊断，原应用DOM/字形/材质完整原生，测试截屏明确animations:allow；0710FF及0713WK两DPR各2/2，三个held/reverse场景全屏RGBA0、实际DOM矩形/节点身份/释放/刷新现场通过，分数DPR旧原生回退。0718真实Firefox90秒bounded54ms FAIL（2670事件，首120/60、steady2550/53，6/4682长帧，7目的校验，双PTY119783/119669B/s，max未ACK77701，实际2674原生运动更新）；51266退出1/88529夹具清理0，无收益未采用。原生窗口软件绘制保留边界仅测试诊断：0730FF/0740WK两DPR各2/2全屏RGBA0；分数DPR完整原生回退，原0733WK差异和0736共享开发服务冲突的失败保留，测试入口改为BLORA_E2E_PORT=18573且同服务浏览器套件串行。原生后台app-view保留层0743FF整数DPR max1/4通道FAIL，未采用、不进入压力。0746软件绘制真实Firefox90秒bounded52ms FAIL（2720事件，首120/58、steady2600/52，6/4736长帧，7目的校验，双PTY119710/119765B/s，max未ACK66609，实际retainedWindowLayers0）；1144退出1/48837夹具清理0，无收益未采用。

0830～0839后台事件唤醒试验：消费ACK事件替代10ms轮询，持久归档事件替代75ms读取轮询，至多两个idle读者预留RPC控制容量。0830/0834 Go race及真实PTY均通过，0833全Go和0835最终受影响五模块均退出0；0836 Chromium11/11（44.095s）、0837 WebKit11/11（55.3s）恢复/原终端检查退出0，前端同源Firefox0507 11/11仍有效。但0839 fresh正式Firefox90秒bounded60ms FAIL（2395事件，首120/69、steady2275/59，36/4319长帧，六次16MiB目的校验，双PTY118205/117990B/s，max未ACK86203，sample94.376s），10310退出1/3879夹具清理0。Data7031及journal8542写/112751213字符/274ms，较0512原1633 Data/2566journal显著增加，立即读取增大浏览器频繁处理负担，不能作为收益采用。Manager/Archive/Daemon新增ReadWait/通知/capability及对应专用试验测试全部撤回，保留失败压力/正确性报告；原字节持久归档和75ms读取已完整恢复。

0846 ACK事件单独候选正式Firefox90秒bounded56ms FAIL（2595事件，首120/69、steady2475/55，16/4570长帧，七次16MiB目的校验，双PTY117321/117276B/s，max未ACK72057，sample94.424s），28893退出1/86263夹具清理0。输出节奏恢复后仍无收益，FlowLedger事件等待及Master使用也完整撤回；没有保留死代码，原protocol/Daemon/Manager/Archive源均恢复，84段期间失败压力和正确性报告仍保留。原实际消费后ACK门槛未变，真实PTY分组测试保留新增hold-first-ACK形成实际背压的检查。0854 Master回退构建30032退出0，二进制恢复ce98b304f1ff3d3c0cf2d3ceb51fcc25398f5e5052287a2da09ab38b528ef22d；前端543源4dcc843c/169bundle1bf460d4未变。

当前独立原生row-relevance仅tests/helpers：不跳过解析或改原xterm节点，不捕获应用/文字/图标，content-visibility:hidden只略过原viewport完全裁住的行子树，保留原行盒、两行/两像素可见边缘；几何改变立即原生恢复，分数DPR/其它renderer完整原生回退，观察器不观察PTY字符/span，断开行/host释放记录。0854类型退出0；0854 FF两项FAIL保留（四场景全屏RGBA0但release max45/6426通道；fractional原行高度浮点Set断言错误），改用稳定原生截图与完整原始/候选行DOMRect精确比较，hidden行内在高度取原计算CSS值。0857 FF两DPR2/2通过。0858 WK整数DPR通过、fractional resized-light max45/1620通道失败保留；0900完全禁用候选的独立原生控制重现同max45/1620，PIL只读差异边界[1145,245,1163,275]对应原生xterm滚动条，官方CSS淡出800ms。两张图同实际鼠标悬停保持原滚动条之后0902 WK两DPR2/2、五场景全屏RGBA0/完整行DOMRect/节点身份通过（12.3s），未改任何滚动条样式或像素阈值。0903最终FF含断开行清理/节点身份两DPR2/2（11.2s）及类型57211退出0。

0904原生row-relevance最终52ms FAIL（2670事件，首120/57、steady2550/51，0/4743长帧，五次16MiB目的校验，双PTY107091/107749B/s，max未ACK83345；实际tracked2/hiddenRows44/totalRows56）。69592退出1/7630夹具清理0，没有收益未进入产品。原严格像素/节点身份和失败压力报告均保留。

0912产品遮挡缓冲候选：原来每移动1px都把64px预留边界向外扩大，导致重复更新原生裁切/终端viewport；现在先检查实际可见像素是否仍在已绘制包络内，只有越界才立即扩展。应用/文字/字形/图标仍完整原生，解析/输出/ACK/恢复无改动，释放与结构失效仍原路径。0912 Firefox终端5PASS/2FAIL（官方WebGL缺少WebGL2，不计通过），0918 Desktop16/16通过；静态审查发现终端local/world转换误用，0920修正为area+bounds，115单位/构建再次退出0。原终端比较新增held首次露出/越过reserve/反向三个场景，0921 FF两DPR+covered3/3；0922 WK Desktop/partial/covered19/19（88.6s）通过。

0924 Chromium19PASS/2FAIL保留：Desktop首次比较未等阴影就绪，后增加两窗口已解码shadow前置，0928两DPR原RGBA0通过；终端fractional首次未拖动就底边350像素/1050通道/max7差异。原生几何确认host/region/viewport位置尺寸相等；0928保留auto全轴仍1050/max7，0932扩边与0934单轴clip均2100/max57，全部撤回。产品现在分数DPR终端也采用完整原生绘制回退，与既有正文/阴影相同；保持原字形/ACK/尺寸/像素，不削减实际负载。测试保留全部原零差异/恢复断言，分数DPR显式检查完整原生回退且不skip，整数DPR必须实际partial。0936 Chromium两DPR原partial+WebGL4/4（18.5s）；115单位和生产构建再次退出0。0937 Firefox最终两DPRpartial/covered3/3（21.0s），0939 WebKit同最终源3/3（27.8s），全部功能/单位/构建已退出0。当前543源68e6b5cfc29a73b3319164803b64ff9d907e1d4b4eac426f94a5b0775a76d334 /169bundlef98d48fe9071aad074d40877b270a635ee3294a76c9682008ff6ff742cefc199 /Masterce98b304 /Daemonad792afc。0941 fresh正式Firefox90秒bounded最终48ms PASS（2820事件，首120/51、steady2700/48，6/4967长帧，五次16MiB目的校验，双PTY108736/108586B/s，max未ACK86684，sample93.659s、API12ms；journal1974写/27008242字符/59ms）。62410退出0，84230夹具清理0；没有renderer/光学实验flags，原50ms/首120/全负载保持。不能用单轮PASS关闭R1。0945 fresh正式Firefox300秒bounded最终47ms PASS（8995事件，首120/57、steady8875/46，3/16099长帧，22次16MiB目的校验，双PTY116665/116678B/s，max未ACK77124，sample303.332s、API16ms）。71536退出0/57701清理0。但0953原始未扩展轨迹（没有soak/bounded或renderer flags）120事件61ms FAIL（2/126长帧，双PTY115553/116781B/s、max未ACK76290、sample3.413s），66199退出1；27408夹具清理退出0。不能靠追加稳态样本稀释初始压力来关闭R1。0955的192px reserve候选115单位/构建通过，首次像素失败保留；差异仅顶栏显示跨分钟，固定no-arg Date显示而保留Date.now/事件/RAF后0958两DPR原RGBA0通过。0960 fresh原始120事件64ms FAIL（max70，3/126长帧，双PTY116710/119226B/s，max未ACK65167，sample3.724s，API15ms）；67984退出1/20846夹具清理0。192px候选无收益完整撤回。0963单RAF合并原生平移候选115单位/构建通过，Firefox13检查12PASS/1FAIL：原有“首动画帧前的最新held drag立即提交”检查在requestAnimationFrame停滞时位置未立即变化，违反即时反馈合同；保留原断言，不进入实际压力，完整撤回。当前产品恢复64px reserve，543源68e6b5cf，dist仍0963旧候选需先重建；没有实际压力/夹具或构建/功能测试运行。下一检查原生绘制和样式传播路径，保留即时几何、原生内容和首120/50ms/全部资源/严格像素/恢复/ACK合同；R1仍开放。 0965恢复构建退出0，当前重新按显式sorted web-relative path+NUL+raw digest聚合的543源e5f021bb/169bundlef98d48fe（与0936完全同产物，源码hash差异不宣称同源）；0967 fresh原始WebKit120事件104ms FAIL（max131，6/27长帧，双PTY118952/116424B/s，sample1.046s，最终16MiB目的校验），51328退出1/50206夹具清理0。0969 fresh原始Chromium120事件34.4ms PASS（max39.9，0/137长帧，双PTY114587/122057B/s，sample2.283s，最终目的校验），37463退出0/31065清理0。0971独立阴影直边CSS gradient光学Firefox1PASS/1FAIL（integer光学max2/mean.00112但应用区域max2，fractional完全原回退0），不采用，不运行实际性能；类型检查0。0974新产品候选只对已证明完全覆盖的shadow-piece设visibility:hidden并移除空mask，露出/结构/外观失效立即原图恢复，无应用截图/新renderer/数据延迟；原完整阴影参考强制所有片visible（新增状态保存/恢复，RGBA0未放宽）。115单位/构建退出0，543源daa8fc0c/169bundlebad60ff4；Firefox原desktop16项目前运行74830，当前无实际压力/夹具。下一收功能→fresh原始120真实压力。 0974最终Firefox31PASS/1正常Chromium专用IME skip（3.0m），原RGBA0/即时恢复守卫全部通过；0976 fresh正式原始Firefox120事件65ms FAIL（max75，2/129长帧，双PTY118251/115413B/s，max未ACK86167），33784退出1/70929夹具清理0，visibility候选完整撤回，仅保留更完整的参考强制visible/状态恢复。0979各小阴影边片原生retention候选115单位/构建0、FF11/11（44.9s）原像素/恢复/即时反馈通过，543源4f6dd4d2/169bundlec1643a1d；0982 fresh正式原始120事件59ms FAIL（max66，0/125长帧，双PTY118935/116859B/s，实际64retained edges，max未ACK65359），85350退出1/30094夹具清理0，尚无可接受收益/未最终采用。0985新候选在现有完整native viewport/grid/保护/ACK不变的基础上，委托锁定xterm6原生DOM renderer仅重建可见行，露出立即同一原生renderer补齐；同步输出开启前补齐已提交的隐藏行，原paused native refresh也在首次露出补画，private ABI不匹配/其它renderer保持完整原生回退。115单位/构建0；FF8检查7PASS/1FAIL同步持有断言，失败trace显示持有后的截图/往返跨过xterm原生1000ms自动释放；0987独立同源同原断言两DPR2/2（9.9s）通过，不改超时/断言，初次FAIL保留。0988 WebKit八原检查workers1运行27058；没有实际压力/夹具。下一完整native正确性通过后fresh原始连续运动，并先确认真实可见范围/原生省略量；R1仍开放。 0988 WebKit八原native行/viewport/覆盖/原终端生命周期检查8/8（1.5m）通过；0990 Firefox加强完整参考（解除行委托+overflow，仅恢复当前committed native rows、不改变输出/时钟）两DPR2/2与类型检查0，所有原RGBA0不变。544源8cd2cc48/169bundle29a0f119；0992 Chromium原native行与加强partial参考四项运行36448。下一所有功能结束后fresh唯一实际原始轨迹，对照实际native省略量，源/产物/样本分别记录，未采用/未关闭R1。 0992 Chromium原native行与加强partial参考4/4（29.3s）退出0；全部功能/单位/构建已停止。0994 fresh原始Firefox首120真实混合负载以当前544源8cd2cc48/169bundle29a0f119运行25679；夹具18240（fixture-1421822390）独占。仅增加原生行绘制数值诊断，原轨迹/50ms/资源/ACK不变；此轮作为归因证据，若有效再做独立无该诊断正式验证。禁止并行实际压力或构建/功能测试，运行结束后先清理夹具。

0343 outside-only实际59ms FAIL、0414阴影alpha=1尾部实际53ms FAIL、0430原生SVG保留层实际53ms FAIL，均未采用。原严像素/失败证据全部保留；0412Chromium原两DPR列表/阴影/正文条带/官方WebGL字形8/8通过，分数DPR保留完整原生应用/阴影绘制。下拉框切角和文字裁切是实际WebKit原生UI问题，修复后三引擎深浅色截图和原生键盘通过。

R1/E08仍开放；R2/R3/R4已关闭，Windows不重复；发行延期、未提交/推送。以下所有“当前/运行中”描述为历史记录，不代表本段状态。
## 2026-10-08 最新有效接管状态

2026-10-08 接管更新：普通下拉框斜角/文字裁切属于 WebKit 原生外壳的实际 UI 问题，已修复并通过三引擎深浅色截图及键盘验证。当前产品增加分数 DPR 的应用/阴影完整原生绘制回退；539 源 a9608d4df8bcbfcc3c58acb9823724215c97f43b79a0638db4cef6d45a744391、168 bundle a3600502f55f9c6eae6abff3d5bd244d17db8f454f1dec8ea4e8dfcfa2e6c618。0310 构建/0311 的111单位通过；0332 Firefox15项原生恢复/scroll/覆盖/窗口像素检查通过。上一源 WebKit0238关键7/7通过，不能混算为新源全通过。

Chromium 的原 WebGL1.25失败确认为模拟 DPR 与物理设备像素盒不一致：viewport700×400而canvas560×320；同物理缩放控制0246为700×400且原像素断言通过。浏览器功能配置现在按 DPR 项目启动相同实际缩放；真实压力配置没有修改。0300配置错误（describe内worker级launchOptions）及0302类型构建失败保留，已修复。0312四个 list/glyph原检查通过；0316完整exposure六项中四通过，两阴影比较失败（2通道/max1）；0327原断言重复1/4通过、3/4同差异失败，未放宽或删除断言。分数 DPR 的 separate-strip 大差异已由完整绘制回退消除，其余继续查明。

正式上一产品0231 Firefox90秒bounded51ms FAIL仍是最新有效产品压力结果；原CSS shadow-plane诊断0252为73ms FAIL，26045退出1/89883夹具清理0，未采用。新outside-only阴影拆分保留框内原生阴影/边框绘制顺序；Firefox0321、0340两DPR全屏零差异各2/2，真实收益未确立。WebKit初原型0325失败，后仅Firefox整数DPR启用、其他浏览器/DPR原DOM回退；0331/0337仍在深色切换的原生Dock区出现差异（原型未启用），不是可忽略的PASS。0343真实Firefox90秒bounded诊断正在运行，39565浏览器、76971夹具fixture-1706727654；不并行构建/功能/剖析。下一收报告→浏览器退出→夹具Ctrl+C确认清理0，继续严格像素与稳定低于50ms的结构优化。

E08/R1保持开放；R2/R3/R4闭合，Windows不重复；发行延期，未提交/推送。以下历史运行状态不代表当前。

最新产品539源e1102342/168产物3a360b78：独立native/Worker终端解析并行启动，但序号/保护/ACK仍等两边完成。111单位、构建、Firefox7原生产品守卫通过；正式90秒bounded **51ms FAIL**，2720事件含首120、五目的校验、原双PTY/万目录/八窗。62584退出1/23258夹具清理0。E08/R1继续开放，最新WebKit和最终三引擎/长测待验；此前不可变日志300秒53ms及alpha-mask62ms不同轮次独立保留。原严格像素/恢复/流合同不改，Windows不重复、发行延期。

2026-10-08最新同源长测：不可变终端日志正式Firefox300秒bounded **53ms FAIL**，8520事件（首120/58、steady8400/52），23次16MiB目的校验、原双PTY/万目录/八窗口；61233退出1/41824夹具清理0。源b98487b1/产物32a54525仍有效。90秒50ms只是短轮阈值边缘PASS，R1继续开放。下一独立原生alpha-mask光学诊断，产品应用/字形与既有零像素断言不改；R2/R3/R4闭合、Windows不重复、R5延期。以下运行中均为历史。[最新证据](../reports/e08-render-architecture-2026-10-06.md)。

当前E08/R1仍开放：正式不可变终端日志候选Firefox90秒bounded **50ms PASS**（2820事件含首120，8次目的校验，原双PTY/万目录/八窗口；浏览器退出0、夹具清理0）。111单位与构建通过，Firefox六项原生恢复/终端回归通过。当前源b98487b1/产物32a54525，同源300秒bounded正在验证；三引擎最终原60秒及严格功能证据尚未齐，单轮PASS不等于完成。未采用的shadow sibling原型严格比较仍失败，保留独立实验入口和全部原断言；成品没有启用。R2/R3/R4闭合，Windows不重复，R5发行延期。以下历史“当前/运行中”已过期。[最新证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08 最新有效结果：正式 dedup 产品 Firefox 90 秒 bounded **p95 55ms FAIL**（2645 事件，首120/62ms，steady2525/54ms，19/4629长帧，7次16MiB目的校验，双PTY118726/118958B/s，max未ACK66398）。浏览器28581退出1；fixture97515清理退出0，当前无真实压力/fixture。原工作负载、首120、50ms、恢复/ACK与全部零像素断言保持。当前539源016ec7cf69939c319c00abbd36212dbf6f66cfa32dbd42c352c483b5ae18292e /168bundleb206021c9b6093e89b6b309a4b03765c3c5badab64fd52df0bf94eeffa7fd49d。Firefox/WebKit原生恢复、实际scroll/selection、covered/partial各6/6通过；本次结果不足以关闭E08。新增仅内容空白光学阴影 sibling 对照在DPR1及1.25均出现RGBA差异，原型未进入产品，不计PASS；失败报告保留。下一具体动作：独立 Firefox 原生线程剖析，在原八窗口/双PTY/万目录/连续校验传输负载下定位主线程与绘制成本，只输出数值摘要，剖析结果不作为性能验收。R1开放；R2/R3/R4闭合，Windows不重复；R5发行延期。下方历史“正在运行”不代表当前。

2026-10-08最新有效状态：正式RGB产品0822 Firefox300秒p95 **51ms FAIL**（8645事件，20目的校验，fixture清理0），因此E08/R1未闭合。现在新增完整基线+有界终端增量IDB事务候选，108单位通过，Firefox/WebKit各5原生恢复/终端守卫通过，Chromium两原生恢复通过；Chromium软件WebGL DPR1.25原像素守卫依然失败，所有原断言保持。0847中断没有JSON，不计结果。当前539源3324bac7d908fb758440eb5e8065a9948ed0259e4aa48c57ee8739e76535589a /168bundle38a5239050423ec180b37ef8092a35a1775fd445146fdcf3394329f6ee591def已构建，本轮同源正式Firefox90秒正在运行。原生下拉框斜角/文字裁切已修复，三引擎×深浅色实际截图与原生键盘选择通过。R2/R3/R4闭合，Windows不重复，R5发行延期。[最新证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08 最新有效状态（覆盖下方历史候选）：0715原生可见行委托正式Firefox90秒54ms FAIL，已从产品撤回，独立原生参照守卫保留在tests/helpers；不宣称其延迟收益。0746仅壁纸五种光学材质无损RGB24诊断完整Firefox90秒bounded首次50ms PASS（2745事件、首120全部计入、六次目的校验、双PTY约116KiB/s），fixture清理0；诊断不能关闭E08。现在RGB已集成仅material-cache，原尺寸/颜色/阴影/正文/字形/恢复/ACK及全部负载保持，透明像素回退原PNG。0758 Firefox四项通过，六配色×深浅×DPR1/1.25的24次完整光学比较RGBA0。最终0807构建、102单位通过：538源6e8eeb6bcc7a749820699cc3cc9413c9a3946d9588a99fc34a6d579798ffee20，168产物5ead3af9ca46c6c1e4b6b4ef5eb3645f93b51986c63a41ae3a5b61a1b501476a。当前正在最终三引擎功能与同源原60秒/持续压力验收，R1仍开放；R2/R3/R4闭合，Windows不重复，R5发行延期。[完整证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08当前E08（本段覆盖下方同日期历史“最新/运行中”记录）：alpha-two正式诊断55ms FAIL、partial viewport正式54ms FAIL，均不算达标。原生数值诊断确认屏幕只露底部7/33px时，default renderer仍重建全部行；新增pinned官方DOM renderRows可见行委托，正文/字形仍归原生所有，无live-app捕获或自定义字形。首RAF露出/隐藏更新/2026同步输出/选区/DPR1及1.25整屏与独立native参照RGBA0、解析/保护/ACK/refresh：Firefox原3+新2项、WebKit5/5通过，102单位通过。0708构建因显式函数this类型缺失退出2，修正后0711构建65328退出0。当前538源385d7a9a67c260788f310392c9fc2d79c2f6a96133dc0abedd8dfc6e5f666671（src文件加package.json/package-lock.json），168bundle63d310d8440670f582bc0fabce101ca0cbe3b0dc3519a9eff4853546fef007da；0715 fresh Firefox90秒bounded完整压力运行中，无renderer/光学诊断flags。最终同源三引擎/原60秒/持续负载尚未齐，R1开放，不能混合旧源PASS。WindowsR4已闭合不重复、发行延期。50ms、首120和全部原资源/像素守卫不变。[证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08最新E08状态覆盖下方历史候选：native motion driver完整Firefox90秒bounded54ms FAIL（2695样本、initial71/steady52、八次目的校验、双PTY约117KiB/s），直接driver/owned stamp撤回，102单位和两引擎既有窗口/恢复检查通过不等于延迟达标。当前537源381c79645cd20ae4191f67d4867ade3fa4d41fd607c7a8fa04dea0f1340a5fe6、168产物9a6a341450f2cf97fedcc66c00861b8ea60708fdc7681c17b88b79dcafd07798；最终同源三引擎及持续负载仍未满足，R1开放。新仅诊断native-shadow-alpha保留原PNG/1:1UV，在原完全遮挡裁剪下划分非零alpha区域；直接背景band的WebKit max1失败保留，改原完整quad+overflow后Firefox/WebKit各2/2整屏RGBA0，深浅/held reveal/resize及小数DPR回退通过。真实90秒对照进行中，不采用/不记成品PASS。Firefox额外强制WebGL测试因环境WebGL2不支持失败，原默认renderer解析/保护/ACK/露出回归通过；不宣称WebGL像素通过。50ms/全部负载/原零RGBA守卫不变；WindowsR4闭合不重复，发行延期。[证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08最新状态优先于下方历史记录：原17e8c75c/产物a402d394的60秒三引擎PASS保留，但额外Firefox300秒bounded53ms FAIL，不能关闭E08。新明确指纹法的geometry去重源98cb5188/产物61f38fba，八交互回归通过，真实90秒bounded仍53ms FAIL（initial61/steady51，2695事件、七目的校验、双PTY约117KiB/s），浏览器退出1/夹具清理0。正在消除拖动中已全部露出shadow/body的冗余遮挡计算，完整负载与50ms、正文/字形零像素差异和恢复合同不变；最终同源三引擎/持续负载未满足。历史指纹算法尚未复核，不混算旧新结果。R1开放/WindowsR4关闭不重复/发行延期。[证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-08当前源17e8c75c/产物a402d394：原60秒完整负载Chromium30.9ms、Firefox49ms、WebKit49ms，三个正式baseline均PASS，初始120事件全部计入，不混入诊断/旧源。八窗口、双真实PTY、万目录、持续目标校验传输和50ms保持。现进行同源较长稳定性及最终三引擎功能守卫；尚不关闭E08/R1。旧失败与候选记录保留，以下历史“正在验证”不代表当前。WindowsR4关闭/不重复，发行延期/未提交推送。[本轮证据](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07当前：wallpaper-only材质交界产品正式原60秒WebKit46ms PASS；新宿主材质登记观察器隔离后Firefox正式53ms FAIL（初始120事件64ms、后续1512事件44ms），旧PASS不与此版合成全通过。全8window层55ms、terminal paint51ms、opaque underlay54ms未采用。新登记/动态菜单和既有cache lifecycle正常2/2通过；旧material全树观察负对照同断言失败4→1926查询，不计产品失败/通过。102单位/类型构建及DPR1/1.25全画面零差异保持；Firefox原生剖析正在定位冷启动/大幅位移成本，最终同源三引擎/长测未齐，E08/R1开放/50ms不变，Windows不重复/发行延期。

2026-10-07最新：阴影遮挡完整/crop union均53ms、活动shadow plane55ms、活动body材质扁平化51ms，全部FAIL/不采用；精确path/SVG正文mask未通过严格像素对照也撤回，保留全部新增回归。当前仅wallpaper-only header/body材质交界重叠候选：WebKit DPR1/1.25整屏RGBA零差异、held reveal/focus2/2及102单位/类型构建通过，正式原60秒WebKit八窗口/双真实PTY/万目录/校验跨节点传输运行中。最终同源三引擎和较长稳定性未通过，E08/R1仍开放/50ms不变。WindowsR4关闭，不重复；发行延期。此段覆盖以下历史候选的“正在验证”措辞。[实测与失败记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07当前状态优先于下方历史候选的“进行中”措辞：双固定阴影opacity74ms FAIL已撤回，focused-only正文63ms FAIL未采用，缓存几何完整tile遮挡53ms FAIL（3936样本/五校验循环/双PTY约117KiB/s/末轮校验/清理0），均不关闭R1。新候选使用两种原生alpha裁剪的union缩小遮挡判断，平移无额外布局读取、手势露出只扩大、cache发布/回退/尺寸失效。101单测、类型/构建和原生shadow/布局回归通过，DPR1/1.25全画面RGBA停画与完整绘制零差异，包括held露出/release/fallback；完整原负载WebKit60秒正在验证。最终同源三引擎/功能/稳定性尚未补齐，50ms不变。WindowsR4关闭/不重复，发行延期/未提交推送。[真实架构与失败记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07当前E08（覆盖下面历史Worker候选）：Worker完整Firefox52ms FAIL已撤回；恢复服务及原快照回归与HEAD一致。当前仅活动/移动窗口保留transform绘制层、显式size/layout/style隔离及官方原生缓冲区正确性修复。Firefox同源活动层诊断47ms，但正式当前WebKit60秒51ms FAIL（3912样本、五次校验循环、双PTY约117KiB/s），不能关闭R1。100单测/类型/构建通过，WebKit功能26pass及1个仅Chromium原生IME跳过；继续隔离绘制开销，50ms及真实负载不变。WindowsR4不重测，发行延期。[记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07最新E08：窗口containment及官方原生缓冲区正确性修复的完整原负载WebKit49ms PASS、Chromium30.8ms PASS、Firefox51ms FAIL，所有非延迟断言通过；仍须全部引擎及稳定性，R1开放。新候选把高频已保护终端日志副本及同格式原子数据库快照放到专用Worker，其他提交/显式flush完整播种，保留立即同步日志、解析/ACK、proxy回退、未保护错误及清理；105单测/类型/构建通过，Firefox真实Worker/fallback/故障与原遮挡终端5/5通过。新候选Firefox完整实测中，不能混合旧版本PASS为新版本全通过；最终功能/视觉/稳定性待补。[记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07当前E08（覆盖下方历史候选）：自定义scissor已撤回。严格原生像素回归在未修改绘制器、且独立于窗口containment的fixture也发现丢弃缓冲区导致idle露出/同步输出错误，现仅保留官方WebglAddon(true)选项；DPR1/1.25整图零差异及同步输出2/2通过，解析/保护/ACK和恢复服务不改。100单测与最新类型/生产构建通过，窗口contain:size layout style的8项功能回归通过；fresh完整原负载WebKit60秒实测中，没有性能PASS。50ms及全部真实负载保持，R1开放、最终全引擎/视觉/稳定性待补，已通过Windows不重复，发行延期。[架构记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07当前E08：绘制前合并纯平移遮挡56ms、同源活动层56ms；原生preserved-scissor经DPR1/1.25零差异/idle/同步输出验证正确，但完整原负载59ms无收益，已连同新增helper/preserve设置撤回。快照合并65ms也已撤回，恢复/终端服务再次与HEAD一致，失败证据保留。新窗口尺寸/布局/style containment候选100单测与类型/生产构建通过，功能和原负载实测中；50ms未达，最终三引擎/视觉/长测待补，R1开放。先前已通过Windows不重复，发行延期。

2026-10-07最新E08：直接事件内拖动正式双PTY完整负载WebKit58ms/Firefox58ms均FAIL，全部非延迟条件通过。100ms快照合并虽然复制1885→641次但WebKit65ms无收益，已撤回；恢复与终端服务均与HEAD一致。保留原生WebGL局部露出/闲置全屏像素回归，WebKit通过；撤回快照候选完整功能23pass/1native Chromium IME skip不代表最终源码证据。新候选在事件内立即平移、绘制前合并纯平移遮挡更新，其他布局/焦点变更立即处理；100单测及类型/生产构建通过，真实大幅拖动露出回归及完整原负载验证中。50ms及所有负载不变，最终全引擎/视觉/长测仍待补，R1开放。

2026-10-07当前E08源码以指针事件内直接发布非响应式拖动transform为新候选，resize保留RAF合批；字形曝光上传候选因真实闲置露出像素错误已连同新增专属helper/测试撤回，终端服务与HEAD完全一致。首轮WebKit43ms但仅一条PTY持续，整体FAILED、不计达标；新增双命令各>32KiB真实输出前置，原双PTY速率、八窗口、万条目录、循环传输、double-RAF/50ms和60秒不变，fresh完整复测中。100/100保留单测、类型/构建通过，最终功能/视觉/跨引擎与长测待补，R1仍开放。以下字形69ms记录属于已撤回候选。

2026-10-07最新E08候选：按真实字形露出区域裁减原生WebGL上传，完整WebKit60秒 **69ms FAIL**；3744样本、四次真实16MiB校验循环、双PTY约114KiB/s、队列有界，非延迟断言通过。同构建原生上传对照、局部露出像素/无输出重绘和最终全引擎验收待补，不宣称因果收益。独立正常/活动原生阴影裁剪86ms无明确收益；DPR1.25错位已修复，八态原生视觉对照无页面错误，非像素完全一致。Firefox离屏mouseup遗漏已按真实指针序列修复，原离屏/缩放/立即刷新用例1/1通过。102单测及类型/生产构建通过，最终源码全引擎回归待补。原50ms/真实负载/恢复合同不变，R1开放；下段36.9/79/84ms是此前构建证据，不能升级新源码PASS。见[架构记录](../reports/e08-render-architecture-2026-10-06.md)。

2026-10-07 E08架构专项仍进行中：用户授权壁纸光学缓存（仅采样壁纸）及渲染架构优化。当前保留的原生阴影八片alpha裁剪构建在完整60秒负载下Chromium **36.9ms PASS**、Firefox **79ms FAIL**、WebKit **84ms FAIL**；三个引擎均四次真实16MiB循环传输校验成功，双PTY约110KiB/s持续、队列有界。原50ms/8窗口/双真实PTY/万项目录/double-RAF不变，E08不能关闭。99单测、类型/生产构建和该源码WebKit10项针对性回归通过；阴影窗口区域深浅色×材质开关原生对照最大通道差3～6，非完全像素一致。合并光学面151ms、正文常驻层110ms、手势挡板111ms、阴影遮挡106ms均撤回；阴影遮挡的两项通过和100单测仅属已撤回候选，不提升当前功能证据。仅活动窗口常驻层的新隔离对照与同步快照复制诊断待收取，最终源码完整跨引擎功能回归待补。详见[架构及真实测量记录](../reports/e08-render-architecture-2026-10-06.md)。下方旧E08行是历史失败与环境范围，不代替本段最新实测；R2/R3/R4关闭不变，R5发行延期。

2026-10-06 R4普通Windows真机入口完成：e546c8f0报告固定source/runner身份及全checksum吻合，八阶段exit0、14必需真实检查唯一PASS，无缺失或失败；浏览器LF保存核对、停止状态重启和资源清理通过，见[真机证据](../reports/windows-device-passed-2026-10-06.md)。R4关闭，不重跑同组或已关闭native/WebKit组。当前非发行剩余只有R1 E08严格延迟，发行按用户延期；未测特殊环境不升级PASS，历史FAILED保留。下方旧R4待验证描述为历史，不作为下一动作。

2026-10-06 R4 cf83c5ea真机报告身份及全归档校验通过：下载/构建已通过，11项真实功能+cleanup通过，browser正文读回失败、停机重启未执行，整体仍FAILED。Windows-UA真实后端复现LF变CRLF，修复模型创建/历史重建显式EOL；本地当前真实后端生产包Windows-UA14/14及LF/CRLF定向2/2通过，见[报告](../reports/windows-device-editor-eol-2026-10-06.md)。UA复现不计Windows原生PASS，新版需设备确认，E08未达及发行延期保持。

2026-10-06 R4再次源码阶段阻塞：ba68e5d -ArchiveOnly下载30.2秒exit -1，14项未开始，不计产品失败或Windows通过。删除两层EncodedCommand，使用文件启动及JSON参数，新增child PID/start/download边界日志。隔离PS7完整wrapper回归通过；精确下载路径实取固定GitHub ZIP4.0秒、1219文件安全解压通过，见[追加证据](../reports/windows-device-source-2026-10-06.md)。当前Windows根因及新版设备结果未验证，R1/延期发行状态保持。

2026-10-06 R4源码获取阻塞：用户固定2bfac47的clone阶段30.3秒exit -1，14项真实联调尚未开始，不计产品14fail或Windows PASS。wrapper增加完整child错误输出及固定提交HTTPS源码ZIP路径，-ArchiveOnly避开该Git失败；原failed stage/恢复标记、script-source SHA和14项必需断言保持。隔离PowerShell回归与实际GitHub固定ZIP1218文件安全解压通过，见[报告](../reports/windows-device-source-2026-10-06.md)。原Windows失败根因未凭console确定；新版Windows效果仍须回传，E08/发行状态不变。

2026-10-05 用户排除发行后的本地增量：R2门禁已修并通过5/5浏览器及类型检查；R3普通Windows真实联调入口已实现，Linux最新源码实际14/14通过（11.388秒）、Go安全2/2/vet/Windows交叉编译、正式frontend与PowerShell报告harness通过，见[本轮证据](../reports/local-remaining-2026-10-05.md)。Windows真实后端结果仍须一次新普通设备脚本回传，不从Linux提升Windows状态、不重跑已关闭补测组。R1新布局隔离仅微探针、截图一致但仍远超50ms，未采纳，见[诊断](../reports/render-layout-diagnostic-2026-10-05.md)；原E08失败保留。R5发行打包/兼容回退/发布按用户最新要求延期。**当前非发行剩余只有R1严格延迟与R4新入口的Windows真实执行证据**，下方R2/R3待做文字是历史记录，不能重新加入当前清单。未强制环境和历史失败继续保留，不称原全环境认证通过。

2026-10-05 用户要求调整本次必测范围：已由三个子代理分别审查存储/备份/传输/调度、Windows平台能力和终端恢复/性能，主代理核对源码与既有证据，见[范围调整报告](../reports/scope-review-2026-10-05.md)。物理掉电/缓存丢失、宿主/NTFS实际写满、浏览器真实驱逐、公网多主机/证书运维、超过24h及多桌面全组合，以及Windows通知/IME/额外headless/特权服务任务修改等改为本次不强制，保持未实测事实而不记PASS。普通错误处理/权限/源保护及已有安全测试不删除。Windows防火墙规则变更本就不属支持子集，不能再要求用户验证其回滚；新增确认的读写门禁错误仍待修复。旧Vim字节/单次撤权超时转为非阻塞历史观察，不声称根因已修复。**当前阻塞清单只保留R1～R5：E08目标、门禁修复、Windows正常真实后端/发行演练入口、一次正常真机结果、最终发行验收收尾**，详见[最新剩余表](../../execution/NON_UI_REMAINING_2026-09-26.md)。下方历史范围/未测说明不自动重新成为本次阻塞项；不宣称原全环境认证完成。

2026-10-05 Windows第十六轮收尾：见[报告](../reports/windows-sixteenth-report-2026-10-05.md)，固定8d6b14e、完整hash/ZIP集合及来源/阶段检查点匹配；实际headed剩余同例3/3，原45/5秒、取消身份/保留缓存/无帧恢复断言通过，无retry/skip/flaky/全局错误。构建16.35秒及所有十阶段PASS，fresh正式preview端口12408。第十五轮其他四例各3/3保持独立版本证据，当前requested visible WebKit补测组结束，无需重复；不把这两份报告伪合成新版单批15/15，不修改旧FAILED。Go/native/SDK本轮未选不记新PASS，默认headless、自然周期/冷启动/指针绘制、完整平台/物理故障/E08仍未因本轮关闭。

2026-10-05 Windows第十五轮增量：见[报告](../reports/windows-fifteenth-report-2026-10-05.md)，固定d75b825、完整hash/ZIP集合及15条实际headed证据匹配。构建PASS，WebKit14/15；刷新、真实queued、日志、Worker终端各3/3。剩一次prevented-navigation第二个held read准备5秒失败，未到恢复边界，同例其余2/2通过；Windows调度原因不作推定。仅测试准备通过已注册轮询回调建立实际held read；全部原恢复/取消/缓存/无帧断言及45/5秒不变。-TaskRecoveryOnly只同例三次、缺失/错误标题/次数失败关闭，历史通过不导入当前报告。ZIP新增版本和时间，CollectRef防选错、收集后完整checksum覆盖。此前重复上传不计新一轮或新PASS；本地验证和实际Windows剩余结果分开，完整平台及E08等缺口未关闭。

2026-10-04 Windows第十四轮增量：见[报告](../reports/windows-fourteenth-report-2026-10-04.md)，固定63cacc0、17/17摘要及完整ZIP集合匹配；Windows正式构建24.36秒PASS、OS端口3570实际fresh preview成功，WebKit12/15。无帧恢复/真实queued/日志各3/3，剩filter元素5秒缺失、held read准备迟到、初次导航35秒后接管耗尽45秒，Windows根因未定。新增可选-Headed同五项环境对照，实际逐执行launch fixture注释验证，缺失/重复/不一致BLOCKED，PWDEBUG隔离恢复，原45及5秒保持。初次Linux headed12/15、175.945秒保留；私有同例两次0/1确认reactive controls导致Worker postMessage DataCloneError，恢复现将controls/links转为纯数据，既有用例新增两次刷新后Worker仍在断言，协议/ACK/输入不重放/UI不改，不声称此修复解释Windows失败。新类型构建3740模块/Vite7.73秒、90单测/17文件/1.850秒、最终类型及脚本harness通过。本地精确五项3次headed15/15（181.922秒）、headless15/15（91.029秒），六文件终端恢复WebKit62/62（231.936秒，源码import开发server），Firefox/Chromium两文件正式preview各7/7（66.054/51.943秒），引擎依次、无skip/retry/flaky/全局错误。真实JSON/finally保留上传FAILED、两个缺少完整次数标题及三份诊断；真实headed15/15经finally通过，删一条实际launch注释即INCOMPLETE。所属容器/会话结束。下一步固定新源码/脚本`-Followup -BrowsersOnly -WebKitOnly -Headed`，正式构建及15/15回传；此轮Linux及可见窗口比较不关闭Windows默认模式可靠性、完整原生/E08/外部缺口。

2026-10-04 Windows第十三轮增量：见[报告](../reports/windows-thirteenth-report-2026-10-04.md)，固定c63bcb4，14/14摘要及完整ZIP集合匹配，正式构建24.04秒PASS；5173已有监听导致WebKit启动失败、零用例执行，不是15个产品失败，也不能确认前轮任务恢复修订通过。脚本现每个引擎启动前由OS选择空闲loopback端口，baseURL/server同步，strictPort/fresh不变，不杀/不复用原监听；报告保留有界全局启动错误与实际URL/端口。原五标题各3次严格15/15及45/5秒、产品/UI/诊断均不改。本地真实PS7 harness、类型及端口选择通过；原端口占用可复现，选中端口抢占仍失败关闭、6种非法配置拒绝；保持5173占用时WebKit精确15/15（97.036秒）、Firefox正式preview7/7（66.612秒）、Chromium开发server7/7（78.294秒），引擎依次、无skip/retry/flaky。真实AST函数核对本地JSON及实际回传零执行，真实finally保留FAILED/全局启动错误/URL和未选范围并生成独立私有ZIP；1,145文档链接及差异检查通过，容器/会话均结束。不计新的Windows有效性或本地构建/Go/native/SDK/单测/E08/外部缺口通过，下一步固定新版Windows正式构建及15/15回传。

2026-10-04 Windows第十二轮增量：见[报告](../reports/windows-twelfth-report-2026-10-04.md)，固定90658d0，17/17摘要及完整ZIP集合匹配，正式构建PASS；WebKit12/15，无skip/retry/flaky，终端/日志/无帧恢复各3/3。剩任务窗口准备耗尽45秒、首列表未就绪和导航后计数原5秒失败，后一Windows原因仍未确立。仅修订测试准备/诊断：可见可用任务应用原生Enter，真实首列表HTTP/body及filter就绪；queued mock只在浏览器实际beforeunload后变为7，避免普通预先轮询误判恢复，原45/5秒及queued/不写入/错误/终端协议断言不变。真实JSON消费和DOM计数变化补浏览器时间及严格数字白名单，不计绘制通过；产品/UI未改。最终WebKit21/21（114.264秒）、Firefox7/7（42.794秒）、Chromium7/7（41.167秒），独立模块等到第三次真实轮询后才释放的冷视图1/1（9.729秒）及类型构建/脚本harness通过；故意失败0/1探针验证2823字节诊断真实复制/任意字段排除，独立保留不计PASS。真实本地/上传JSON标题次数/12/15拒绝/三份诊断复制及CLI严格15次选择核验通过。Windows实效待固定新版正式构建及五项各3次15/15回传；本轮Go/SDK/其他Windows引擎未选择不导入旧PASS，原生/E08/外部缺口保持。

2026-10-04 Windows第十一轮增量：见[报告](../reports/windows-eleventh-report-2026-10-04.md)，固定`7426801`，16/16摘要匹配，正式构建PASS，WebKit13/15，无skip/retry/flaky，日志3次通过。一次首次计数断言未到导航边界，一次终端45秒失败未挂载终端，不确立Windows原因。仅修订测试准备：首次真实summary HTTP/body后原5秒渲染、冷初始化计原45秒；终端五步可见/可用按钮原生Enter，同一UI，原刷新/worker/协议/ACK/身份/不重放/存储失败/后续指针保持；取消记录按每文档读取编号，held read2须恰一次离开后取消，其他取消也须在离开后，避免慢刷新不同请求混计。产品/UI未改，不计准备指针/冷启动5秒性能。最终WebKit21/21、Firefox7/7、Chromium7/7、类型检查/harness/真实JSON/严格15次选择通过；Firefox初轮6/7及trace保留不计PASS，初轮WebKit36/36为记录修订前证据。Windows实效仍待固定新版正式构建PASS、五项各3次严格15/15；本轮Go/SDK/其他Windows引擎未选择不导入旧PASS，完整原生/E08/外部缺口未关闭。

2026-10-04 Windows第十轮增量：见[回传报告](../reports/windows-tenth-report-2026-10-04.md)，固定`3ef289a`，摘要17/17匹配，正式构建PASS，WebKit12/15，无skip/retry/flaky；fallback/worker终端3次通过，剩两次控制台指针稳定等待45秒超时（未开始日志请求）及一次无帧汇总5秒断言失败（后有HTTP完成），具体Windows绘制/数据发布原因未确立。现读取恢复立即重查活动任务查询，保留缓存且不替换已恢复读取；日志生命周期现场以原生Enter打开同一实际UI，明确不计这三步指针稳定覆盖。原45秒/5秒、实际刷新/worker/ACK/身份/无输入重放及错误断言保持，UI未改；实时查询/日志阶段、有界数字计数和按钮几何加入失败诊断。本地90单测、类型构建、WebKit36/36、Firefox12/12、Chromium12/12、脚本harness/实际JSON/严格15次选择核验通过；探索暂停时钟未复现指针停顿，独立保留，不冒充Windows因果；独立故意失败探针验证857字节诊断真实复制，不计产品PASS。下一动作固定新版`-Followup -BrowsersOnly -WebKitOnly`，构建成功且五项各3次15/15回传；Windows实效仍未验证，Go/SDK/其他引擎本轮未选择不导入旧PASS，完整平台/E08/外部环境保持未关闭。

2026-10-03 Windows第九轮增量：见[回传报告](../reports/windows-ninth-report-2026-10-03.md)，固定`3e6c31c`，摘要14/14匹配，WebKit14/15，无skip/retry/flaky；四项导航/日志用例各3次通过，前轮两项读取恢复/建立超时已获得真机通过证据。仅第一次fallback/worker终端达到45秒总时限，之后18.810/18.762秒通过，诊断无页面/请求错误；首次刷新约32.1秒才开始，具体慢动作/Windows因果仍未确立。补测现独立类型构建正式bundle，fresh preview禁止端口回退/服务器复用，缺失或失败构建阻止成功；实时终端阶段及有界静态资源耗时加入失败诊断，原UI动作/45秒时限/5秒恢复/worker/ACK/输入不重放断言保持，产品和UI未改。本地正式WebKit36/36、Firefox12/12、Chromium12/12、类型构建及最终脚本harness/实际JSON/严格15项选择核验通过；隔离冷启动单例case21.625秒→14.766秒，不冒充Windows因果/E08。故意失败探针确认有界诊断实际进入报告，其初次cwd错误独立保留，不计产品PASS。Windows实效仍待固定新版`-Followup -BrowsersOnly -WebKitOnly`五项各3次回传。Go/SDK/其他浏览器本轮Windows未选择，不导入旧PASS；原生/E08/外部环境整体验收保持未关闭。

2026-10-03 Windows第八轮增量：见[回传报告](../reports/windows-eighth-report-2026-10-03.md)，固定`57d4fb6`，摘要15/15匹配，WebKit13/15，无skip/retry/flaky；原fallback/worker终端各3次通过，剩一次汇总恢复5秒超时及一次日志待取消读取建立超时，后者诊断visible/focused但paint=false。可控无帧基线0/1在原5秒断言失败，共享只读API门禁新增250ms独立恢复兜底，pagehide取消并失效旧帧/旧截止，新导航重置代次。日志测试通过实际interval回调建立待取消读取，queued观察不再依赖paint结束；原错误/限时/ACK/输入不重放断言保持。89单测、类型构建、本地WebKit36/36、Firefox12/12、Chromium19/19加独立补验4/4、PowerShell harness/实际JSON与精确15次选择通过；Windows修复实效仍待新版`-Followup -BrowsersOnly -WebKitOnly`（五项各3次严格15/15）回传。Go/SDK/其他浏览器未在本轮Windows选择中执行，旧PASS不导入；有界兜底不证明所有慢导航/native dialog/bfcache序列，完整平台/E08及外部环境整体验收保持未关闭。

2026-10-03 Windows第七轮增量：见[回传报告](../reports/windows-seventh-report-2026-10-03.md)，固定`8ad386a`、摘要14/14匹配，WebKit11/12，无skip/retry/flaky；三项任务查询导航各3次获得Windows通过证据。剩余fallback/worker终端错误来自InstanceConsole独立日志轮询，在beforeunload后9ms、pagehide前1ms。可控日志轮询基线0/1，已增加共享安全API读取导航取消/等待、response body取消检查、pagehide旧帧失效和组件销毁取消，原终端/ACK/不重放断言保持；87单测/类型构建、本地Chromium23/23、Firefox12/12、WebKit36/36通过，最终PowerShell harness/真实JSON解析及精确选择清单核验通过。下一次仅5项各3次严格15/15的`-Followup -BrowsersOnly -WebKitOnly`；Go/SDK/其他浏览器不导入旧PASS，修复实效仍待Windows回传；本条不提升完整平台/E08或外部环境整体验收。

2026-10-01 第六轮 Windows 实机证据：固定 `f98b078`、followup-browser，20文件校验值全部匹配；Chromium/Firefox各8/8，WebKit23/24，仅一次fallback/worker终端任务列表请求在beforeunload后8ms、pagehide前8ms报访问控制诊断。前轮两项查询导航检查在Windows WebKit各3次通过，Go/native/SDK本轮未选，不能导入旧PASS。可控真实QueryObserver排队回调复现取消后仍启动新读取，基线0/1；新增只读任务查询门禁与存活页面恢复，组合本地Chromium9/9、Firefox9/9、WebKit27/27，无skip/retry/flaky，80单测/类型构建及脚本harness通过。新增 `-Followup -BrowsersOnly -WebKitOnly`，只跑剩余终端配置与三项导航，各3次严格12/12，不重复已验证Go/其他浏览器；见[第六轮报告](../reports/windows-sixth-report-2026-10-01.md)。本地复现关闭的是取消后排队新读取缺口，Windows原诊断实效及完整平台/E08/外部环境仍待实际验证。

2026-10-01 第五轮 Windows 实机证据：固定 `3c7abb3` 的11必需原生检查、3项Go回归各3次通过；Master/runlog完整race87pass/0fail/4skip，前轮两处Go竞争修复已有真机PASS，WebKit账号场景也全部三轮通过。Chromium/Firefox各6/6，WebKit16/18，仅第三轮两种终端配置的后台任务汇总访问控制诊断仍失败。复现页面离开前读取未取消，任务查询及共享列表接入AbortSignal，在beforeunload/pagehide取消只读查询；最终本地Chromium8/8（79.608秒）、Firefox8/8（115.302秒）、WebKit24/24（263.222秒）通过，无skip/retry/flaky；80单测/类型检查/构建及PowerShell harness通过。新增 `-Followup -BrowsersOnly`，严格要求8/8/24执行，不重复已验证Go/GCC/SDK也不导入旧PASS；见[第五轮报告](../reports/windows-fifth-report-2026-10-01.md)。Windows剩余诊断及完整平台/E08/外部环境缺口仍待验证，不能把Linux通过当真机修复通过。

2026-10-01 第四轮 Windows 实机证据：固定 `08db099` 的 27 项定向 Go 与 11 项必需原生全部通过，目录 metadata/ConPTY Unicode 实际执行、Windows 可移植生命周期与 source restart 已获得真机证据；Chromium/Firefox 各32/32，WebKit29/32。全race事件356pass/2fail/14skip，新增取消上传的调度竞争与日志收尾记录竞争已在本地修正，Master/runlog 完整race退出0；三项修正定向race连续8轮通过。最终浏览器夹具本地Chromium/Firefox各6/6、WebKit18/18通过，原始失败和首轮本地失败独立保留，不能因 Linux 通过提升 Windows 整项。`-Followup` 限定两个 Go 模块及两份浏览器文件，校验实际次数并带失败诊断；详见[第四轮报告](../reports/windows-fourth-report-2026-10-01.md)。E08、完整 Windows/特权生命周期和外部环境缺口不提升为完成。

2026-09-30 Windows首次用户实机证据：Job/keeper/daemon退出恢复、日志管道、监控、SCM/Task Scheduler查询等九项关键原生通过，ConPTY关键项失败；全Go仍有产品兼容缺陷和Linux夹具不兼容，浏览器尚未运行、race缺GCC。Web构建与76单测通过。首批修正及准确边界见[Windows首轮报告](../reports/windows-first-report-2026-09-30.md)，本地Linux回归/Windows交叉编译不是修复后Windows运行证据，不提升Windows整体验收。

2026-09-29 F06/RC34：复现旧连接异步解析失败污染新连接，增加连接代次/销毁保护；当前连接错误仍停止输入。Chromium/Firefox/WebKit定向组合各4/4通过；六包2,251条记录/三份Web、独立停机恢复及兼容RC33回退通过，见[报告](../reports/terminal-generation-2026-09-29.md)。RC33完整复验49/51，两项扩展失败保留；补原生鼠标hover命中前置后RC34完整真实功能51/51通过897.040秒、无skip。玻璃遮挡裁剪像素等价检查失败未采用，UI未变；功能套件不含独立E08负载门禁，严格E08及原偶发/外部平台缺口保留。

2026-09-29 F06/RC33：实际复现并修复OSC 8链接跨检查点丢失，保留主/备用/历史屏幕范围与当前输出属性，并使用原生标记管理生命周期。Chromium61项、Firefox/WebKit各50项、76单测、真实双节点PTY7项通过；实际鼠标点击和两次持久恢复目标一致，原生URI限制保留。RC33六包2,243条记录/三份Web及独立恢复/兼容RC32回退通过，见[报告](../reports/terminal-links-recovery-2026-09-29.md)。UI未改；原Vim偶发、严格E08和外部平台缺口未关闭。

2026-09-29 F06/RC32：修复主题切换后检查点/原始日志中的旧调色板复活，记录本地重置在输出流中的位置，直接重置原生调色板而不向半段CSI/OSC插入字节。Chromium48项组合、最终交错六项及非法标记1项、Firefox/WebKit各40项、73单测与真实双节点PTY7项通过；RC32六包2,239条记录/三份Web及独立恢复/兼容RC31回退通过，见[报告](../reports/terminal-theme-order-2026-09-29.md)。UI未改，链接恢复/原Vim偶发/严格E08/外部平台缺口保留。

2026-09-29 F06/RC31：实际复现并修复OSC索引色/前景/背景/光标颜色恢复丢失，采用最多259槽的数值差异记录与固定指令恢复，保留主题默认重置语义。三浏览器各33项、73单测、真实双节点PTY7项、RC31六包2,235条记录及三份Web一致性、独立停机恢复与兼容RC30回退通过，见[报告](../reports/terminal-colour-state-2026-09-29.md)。其他终端恢复边界、原Vim偶发、严格E08与外部平台未据此关闭。

2026-09-29 F06/RC30：修复鼠标SGR/像素编码、光标显隐/闪烁/样式及换行转换跨检查点丢失。真实查询首轮四项失败，修复后Chromium/Firefox/WebKit各27项（含实际鼠标按下/释放字节对照）、71单测通过；真实双节点PTY首轮6/7超时，保持原断言和60秒限时串行复核7/7通过。RC30六包2,231条记录及兼容RC29恢复回退通过，见[报告](../reports/terminal-protocol-state-2026-09-29.md)。不据此关闭原Vim偶发、严格E08及外部平台缺口。

2026-09-29 F06/RC29：实际复现并修复字符集、滚动区域、保存光标和制表位跨检查点丢失，新增两缓冲区及字符集版本化状态并同步Worker。Chromium30项组合、持久存储重开19项、Firefox19项、71单测和真实双节点PTY7项通过；RC29六包2,227条记录及兼容RC28恢复/回退通过。WebKit首轮18/19（一项模块加载失败），原断言全组复核19/19通过，首轮失败保留，见[报告](../reports/terminal-control-state-2026-09-29.md)。UI未改，E08/原Vim偶发/外部平台缺口不关闭。

2026-09-29 F06/RC28：修复未完成VT控制序列跨检查点丢失，按真实解析器边界压缩，保留有界输出/resize日志并拒绝Worker过时快照。Chromium组合24项与晚返回1项、Firefox/WebKit各21项、正式配置各2项、69单测和真实双节点PTY7项通过；六包2,223条记录及恢复/兼容RC27回退通过，见[报告](../reports/terminal-parser-checkpoint-2026-09-29.md)。原Vim偶发、已完成控制指令的其他持久状态、严格E08与外部平台缺口未据此关闭。

2026-09-29 F06/恢复增量：真实复现并修复终端未完成UTF-8码点跨检查点丢失，保留最多3字节解码续接状态；六个分割位置/Worker/查询9项、检查点+日志混合1项、68单测、真实双节点文件/PTY7项通过。RC27六包、2,219条记录/三份Web、独立启动/停机恢复/兼容RC26回退通过，见[报告](../reports/terminal-utf8-checkpoint-2026-09-29.md)。这是新的确定恢复缺陷，不是Vim原偶发的根因证明；尚未据此证明VT控制序列全部边界，严格E08与外部平台缺口保持。

2026-09-29 RC26最终补验：完整HTTPS/双节点/Docker功能 **51/51、905.725秒通过**；新增专属子组OOM及PID上限实际触发，与原systemd/cgroup场景合计5项通过、9.652秒，见[平台报告](../reports/systemd-cgroup-2026-09-29.md)。Linux X11/DBus/Dunst原生弹窗、文字、系统鼠标点击、刷新去重及浏览器退出后遗留弹窗可关闭通过，最终9.666秒，见[通知报告](../reports/native-notifications-2026-09-29.md)。退出即自动消失的早期预期失败保留，未声称后台投递或退出后重启跳转。当前UI与RC26产品未再变更；严格E08、原Vim偶发来源、Windows、跨主机和物理故障等缺口保留。

2026-09-29 F13/E05/A03/E09平台增量：独立私有systemd容器真实运行3项通过，覆盖服务启动/重启/停止及规范名/别名列表、timer可见性/enable/disable/实际触发、cgroup进程出生归属/整组终止/三项配额文件/CPU实际限流。修复原停止服务与已安装未加载timer列表遗漏。runtime/systeminfo race及Master/Daemon相关7项通过；细节和失败保留见[平台报告](../reports/systemd-cgroup-2026-09-29.md)。不再将Linux对应子项写作完全缺环境；未计Windows、OOM/进程耗尽、生产部署、跨主机和物理故障通过。RC26六包/2,215条内部记录/独立启动、停机恢复和兼容RC25回退已通过，原E08/Vim缺口不变。

2026-09-29 终端/合成诊断增量：真实Vim/top10/10与最终版本3/3通过但未复现原偶发；新增失败字节附件，不放宽输入不重放断言。真实xterm/Worker/IDB查询回放/权限边界与现有终端定向6/6、最终类型检查通过。跟踪明确主要耗时位于SoftwareRenderer合成；显式SwiftShader合成更慢未采用。60秒真实循环传输诊断240样本p95 1062.5ms，在原50ms断言失败，不能替代RC25正式三引擎E08结果；详见[补验报告](../reports/local-diagnostics-2026-09-29.md)。无产品/UI/发行包变更，不将未复现或缺环境视为完成。

2026-09-29 RC25非视觉收尾：恢复快照去除重复复制、重复聚焦不再写恢复日志，66项单测、59/59完整mock、三引擎原生IDB边界及Firefox/WebKit真实恢复故障各3/3通过。真实HTTPS/Docker完整套件为50/51（Vim刷新额外字节一次失败）；带诊断和移除诊断后的原用例各连续3/3通过，保留偶发未定位记录，不能写成首轮全过。RC25六包构建、外部摘要与2,207条包内记录、三份当前Web一致性、独立启动/停机恢复/兼容RC24回退通过，见[完整本轮报告](../reports/local-closeout-2026-09-29.md)。当前严格E08 Chromium/Firefox/WebKit分别1003.9/1108/545ms，均大于50ms；快照专项35%～56%的中位改善不代表整页达标。UI和通透效果全程保留。24h终态证据沿用下条已通过结果；Windows/systemd等环境缺口及整体验收状态不变。

2026-09-29 E01/E04/A09 本地24小时长测终态复核：私有原始检查点与事件经独立只读收取器重新校验，状态 **PASSED_24H**、证据缺口0，生成的[最终报告](../reports/local-endurance-24h-2026-09-27.md)与现有文件完全一致。实际单调时长86,415.480秒、跨2个UTC日期、17,262次真实采样、1,440个成功分钟槽，6h/18h两次失联和同库重启后保留原runId；终态SQLite完整性、最新真实备份正文恢复、所属计划与服务清理均通过。Master/Daemon RSS峰值34,041,856/34,013,184字节，FD峰值14/20，归档峰值16,776,441字节，状态与证据峰值265,899,979字节，均在该次声明预算内。下方9月26～27日的“运行中”是历史阶段记录，现以本条终态为准。此结果只关闭本机这一次24小时场景；当前E08三引擎严格性能仍失败，Windows真机、独立systemd、跨主机、物理掉电和系统通知中心仍缺实际验证，不将全范围标为完成。

2026-09-27 11:14 长测阶段证据：独立24h运行的6h与18h两次实际故障阶段均完成，管理端同库重启前后120点stale历史精确一致，Daemon恢复后保持原实例runId；当前18h01m/12,956采样/1,081个成功分钟slot，仍RUNNING，不计24h通过。新只读收取器已独立启动，会在实际终态核对恢复、SQLite完整性、所属资源清理及不可变二进制，并更新[动态报告](../reports/local-endurance-24h-2026-09-27.md)；其短测与拒绝错误证据测试通过，详见[监督/收取入口](../reports/local-endurance-2026-09-26.md)。整项与外部平台仍不提升为完成。

2026-09-27 当前冻结UI复核：完整mock浏览器56/56、另新增运行中Worker实际终止1/1、原失败扩展链路重复6/6、完整真实HTTPS/Docker功能 **51/51**（无skip）通过，见[本地功能报告](../reports/local-functional-2026-09-27.md)；Go race各包与修复后的完整Master分段复验通过，未将首轮失败改写为通过。RC24六包构建/外部SHA及2,191条包内记录核验、独立SDK/参考包签名、TLS/双Daemon、停机恢复及兼容RC23回退全部退出0，见[最新发行报告](../reports/release-rc24-2026-09-27.md)。E08严格八窗口/双真实PTY/万项目录/16MiB并行传输三引擎仍未达标（Chromium1240.1ms、Firefox1230ms、WebKit739ms），新合成对照亦失败，见[当前性能报告](../reports/performance-current-2026-09-26.md)；9月21～22日通过不覆盖本版本。Firefox/WebKit恢复故障各3/3通过，独立24h跨日监督仍RUNNING，不能提前计作通过；Windows等外部平台仍未验证，整体验收不提升为完成。

2026-09-26 F02/F08/A11 非 Chromium 恢复故障补验：Firefox 155.0 **3/3（37.5s）**、WebKit 26.6 **3/3（50.8s）**、退出0，真实 HTTPS/Monaco/IndexedDB 验证最近中文正文导出/恢复、schema 0 迁移与不支持版本原样保留，以及真实事务 abort 后快照/指针原子回滚；详见[跨引擎恢复报告](../reports/recovery-engines-2026-09-26.md)。配额异常是定向故障注入，不是实际磁盘/浏览器配额耗尽；不代表物理掉电通过。

2026-09-26 E01/E04/A09 长测入口：新增独立受监督的 `scripts/local-endurance.py`，两个完整五分钟真实短测通过，包含真实分钟调度唯一性、120点历史、Daemon失联、同库Master重启、原runId恢复、最新备份正文恢复和准确清理；stop/resume 与监督 worker 中断后安全接管也通过。24小时运行 `.local/endurance-zs_dzpzj` 已启动但**尚未通过**，需跨真实UTC日期并满足墙钟/单调时钟、采样覆盖、故障阶段、RSS/FD/归档上限及末尾恢复核对，详见[长测报告](../reports/local-endurance-2026-09-26.md)。整项与外部平台状态不提升。

2026-09-26 F02 全外观截图证据：当前六配色 × 浅/深色 × 通透开/关 × 桌面/启动器/实例中心共 **72 张**真实 Vue 原图，原图分辨率 3200×2000，正式图标和暖砂深色壁纸 SHA256 与当前资源一致；本地 GET-only 演示数据核验材质开关不改壁纸/图标、零页面错误/API 写入/外网请求，见[验证记录](../../../web/ui-screenshots/all-appearance-20260926/verification.json)及[原图 ZIP](../../../web/ui-screenshots/all-appearance-20260926/blora-all-appearance-originals-20260926.zip)。本轮未改产品源码，截图只用于远程视觉复查，不提升 F02/A01/A02 或后端/外部平台整体验收状态。

2026-09-25 F02 暖砂深色壁纸视觉增量：旧 `sand-dark.svg` 的近黑底覆盖约 70% 画面，现改为暖炭色并增加上方曲面、上移沙丘层次，浅色壁纸与图标、主题/通透逻辑保持独立。真实 Vue [深浅场景对照](../../../web/ui-screenshots/sand-wallpaper-rework/00-overview.png)和[验证记录](../../../web/ui-screenshots/sand-wallpaper-rework/verification.json)覆盖浅色、深色通透/实色共六张桌面与启动器，零页面错误/API 写入。`cd web && npm run build`（含类型检查）通过，暖砂深色/自选壁纸聚焦 Playwright **1/1** 通过；此视觉改善不改变 F02/A01/A02 整体验收状态。

2026-09-25 F02 备份与计划应用图标视觉增量：去掉旧回转圆环左侧断口和外溢钟面，使用闭合表盘及内收钟面；同步正式浅/深色 01A 图标。生产构建通过，聚焦 Playwright **2/2** 通过；[真实界面实际尺寸对照](../../../web/ui-screenshots/backup-icon-final/00-comparison.png)与[验证记录](../../../web/ui-screenshots/backup-icon-final/verification.json)覆盖桌面、Dock、启动器，核实正式资源加载及零页面错误/API 写入。此视觉证据不提升 F02/A01/A02 整体验收状态。

2026-09-24 F02 暖砂深色壁纸二次配色修订：用户否定青蓝版后，正式深色暖砂改为近黑底与浅米金曲面，原几何、03 图标及通透/配色独立性保持。生产构建通过，相关 Playwright **3/3** 通过；[六配色完整截图目录](../../../web/ui-screenshots/dark-wallpaper-final/index.html)和[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)证实六壁纸互异、通透开关不改壁纸、零页面错误/API 写入。远程用户须直接收到 PNG 嵌图；此 UI 证据不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 暖砂深色壁纸配色修订：灰棕旧版改为深蓝绿曲面，保留原几何以及已选 03 图标，浅色、配色选择与通透材质逻辑不变。`cd web && npm run build` 通过，相关 Playwright **3/3** 通过；[真实桌面与启动器图库](../../../web/ui-screenshots/dark-wallpaper-final/index.html)和[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)证实六套壁纸互异、材质开关不改壁纸、零页面错误/API 写入。该 UI 修订不改变 F02/A01/A02 整体验收状态。

2026-09-24 F02 深色壁纸可见度增量：六套默认主题改用保留原几何的深色专用 SVG，修复约 90% 遮罩令壁纸近乎纯色的问题；青灰深色自选壁纸不再被默认规则覆盖。浅色主题、通透开关与已选 03 深色图标独立。`cd web && npm run build` 通过，相关桌面浏览器测试 **4/4** 通过；[六配色实景图库](../../../web/ui-screenshots/dark-wallpaper-final/index.html)与[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)核验暗色壁纸互异、通透 on/off 背景一致、零页面错误/API 写入。此 UI 证据不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 深色应用图标定稿：用户选择 03「清晰造型」作为默认深色图标；14 枚正式资源已切换，浅色资源、配色和通透开关独立。生产构建及桌面主题/材质聚焦 Playwright **2/2** 通过；[七向实景与验证记录](../../../web/ui-screenshots/dark-icon-directions/index.html)可复查。此 UI 决定不改变 F02/A01/A02 的整体验收状态。

2026-09-24 F02 深色通透材质与亮边增量：用户截图对应通透 `on`，此前暗色面约 96–97% 遮盖；现启动器/顶栏约 80%、窗口约 89%、Dock 约 78%，内侧反射、模糊和投影收减，深色窗口/启动器/Dock 外圈改用暗壳边，启动器分隔线压暗。14 枚深色图标改色，并降低仅深色 h04 导出中的反射与壳层亮度；浅色资源及六配色独立性保留。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[通透 `on` 的六配色×深浅×六场景图库](../../../web/ui-screenshots/dark-mode-review/index.html)为 **72 张**，另有暖砂 `off` **12 张**，可对照[深色启动器 `on`](../../../web/ui-screenshots/dark-mode-review/sand-on/08-dark-launcher.png)和[`off`](../../../web/ui-screenshots/dark-mode-review/sand-off/08-dark-launcher.png)。本地隔离 fixture 为零页面错误、零 API 写入；此 UI 增量不提升 F02/A01/A02 的完整状态，亦不覆盖真实节点、后端或外部平台验收。

2026-09-24 F02 深色窗口与图标增量：深色窗口外框改为弱中性边缘，实例中心内容区底角不再露出壳体形成接缝；14 枚深色应用图标针对暗背景调整高亮、底板和主体对比，浅色界面及六套配色、壁纸选择不变。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[六配色×深浅×六场景图库](../../../web/ui-screenshots/dark-mode-review/index.html)有 72 张实景及[实例中心左下角局部](../../../web/ui-screenshots/dark-mode-review/sand-on/09-dark-instances-corner.png)，本地隔离 fixture 核验零页面错误、零 API 写入。该 UI 证据不提升 F02/A01/A02 的完整状态，也不覆盖真实节点、后端或外部平台验收。

2026-09-24 F02 深色外观清晰度修订：六套配色共用中性暗色表面，强调色仍独立；调整暗色壁纸呈现、近不透明窗口/Dock、列表边界及终端画布，重绘14枚更易辨认的01A深色图标。原壁纸资源、选择项与浅色模式不变。生产构建（含类型检查）、聚焦桌面浏览器 **4/4** 与终端恢复回归 **1/1** 通过；[六配色×六场景深浅对照](../../../web/ui-screenshots/dark-mode-review/index.html)重拍72张原图与六张概览，逐套核验14枚浅/深色图标，0页面错误、0 API写入。截图为本地隔离演示数据，不提升F02/A01/A02或外部平台的完整验收状态。

2026-09-24 F02 工作区深色外观增量：保持现有六套配色和壁纸选项不变，选定 01A 应用图标新增 14 枚随浅/深色切换的暗色资源；深色窗口、Dock、菜单、输入/选择控件、资源列表、状态、应用卡片、日志及已打开的 Monaco 编辑器同步适配。生产构建（含类型检查）、聚焦浏览器 **4/4** 通过；[六配色×六场景×深浅两版实景图库](../../../web/ui-screenshots/dark-mode-review/index.html)包含 72 张原图、六张概览，每套分别核验 14 枚浅色/深色图标、0 页面错误和 0 API 写入。截图为本地隔离演示数据，不提升 F02/A01/A02 或外部平台的完整验收状态；壁纸未在本轮改动。

2026-09-24 F02 工作区壁纸艺术媒介增量：针对旧方案画风相同的问题，新添真实摄影、印刷海报、数字夜光、水墨纸张及压纹玻璃五个方向；摄影使用[NPS 公有领域素材](https://npgallery.nps.gov/AssetDetail/204B23EC-1DD8-B71B-0B21B310C2A73B98)，来源和本地处理见[记录](../../../web/src/appearance/wallpapers/media/PHOTO_SOURCE.md)。原壁纸与默认值保留；当前11种自选壁纸不改配色、通透材质或图标。生产构建及浏览器聚焦 **3/3** 通过；[五方向真实界面图库](../../../web/ui-screenshots/wallpaper-art/index.html)共20张原图/4张对照图，核验页面错误和API写入均为0。截图为隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区壁纸偏好增量：新增六种不同构图的 SVG 壁纸，并与主题配色、浅/深色和通透材质分开选择；“随配色”为兼容原有外观的默认值。聚焦浏览器测试 **3/3**、生产构建通过；[六种壁纸实景图库](../../../web/ui-screenshots/wallpaper-styles/index.html)含24张真实界面截图及4张对照图，核验六张各异、配色与14个应用图标不变、零页面错误或 API 写入。截图使用隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区外观配色增量：主题配色由冰蓝/青灰扩展为冰蓝、青灰、暖砂、苔绿、雾紫、烟粉六套，各有浅/深色令牌及配套壁纸；材质开关、配色和浅/深色保持独立，应用图标一致。当前聚焦浏览器测试 **2/2**、生产构建通过；[六套配色实景图库](../../../web/ui-screenshots/palette-showcase/index.html)共72张，覆盖两种材质和六类场景，核验六套各异主色/壁纸、两材质同配色/图标及零页面错误/远端写入。截图为本地隔离演示数据，不提升 F02/A01/A02 或外部平台验收状态。

2026-09-24 F02 工作区外观偏好增量：通透材质开关与冰蓝/青灰主题配色已独立持久化；原有浅/深色主题独立保留。当前浏览器聚焦场景覆盖四组合和刷新 **1/1**，生产构建通过；[25场景四组合图库](../../../web/ui-screenshots/palette-matrix/index.html)共100张本地截图，验证同配色开/关主色、壁纸和14图标资源一致，材质不同。截图使用隔离演示数据，不提升 F02/A01/A02 等工作现场恢复或外部平台验收状态。

2026-09-22 RC23 当前源码发行复核：`development-20260922-rc23` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC22 兼容回退均通过，见[RC23发行报告](../reports/release-rc23-2026-09-22.md)。包内含 Docker 日志归档实时/历史入口、F08 只读分段预览、预览模式控件保护和最新验收台账；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 RC22 当前源码发行复核：`development-20260922-rc22` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC21 兼容回退均通过，见[RC22发行报告](../reports/release-rc22-2026-09-22.md)。包内含 Docker 日志归档实时/历史入口和 RC21 的 F08 只读分段预览；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 Docker 日志归档入口增量：`DockerLogs.vue` 已接入现有有界历史接口，支持实时/归档切换、归档窗口观察时间、截断与可能缺口标记；Docker 浏览器 **2/2（11.1s）**、真实隔离 Docker Engine Master/Daemon 归档 API **2.79s**及 `npm run check`、63/63 单测、生产构建、`make check` 通过，真实链路还覆盖未授权403和 `limit=101` 400，详见[Docker日志归档报告](../reports/docker-log-history-2026-09-22.md)。这只补齐 Linux/浏览器/Docker 可验证子项，不改变 F11/E02 对远程 Engine、Windows 容器、物理故障和长时组合的进行中状态。

2026-09-22 RC21 当前源码发行复核：`development-20260922-rc21` 六包构建、`sha256sum -c SHA256SUMS`、独立 SDK/参考扩展签名、Master/TLS/双 Daemon、停止态快照恢复及 RC20 兼容回退均通过，见[RC21发行报告](../reports/release-rc21-2026-09-22.md)。包内含 F08 只读分段预览、预览模式控件保护与 OpenAPI `previewFile`；发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合。

2026-09-22 F08 大文件只读分段预览：新增受文件版本绑定的 `/files/preview` 有界 UTF-8 分片与浏览器上一段/下一段导航；真实文件 API 覆盖超限文本首段、第二段及二进制拒绝，编辑器边界浏览器 **3/3** 通过。完整 F08 交叉场景和外部平台证据仍按下表保持进行中，详见[编辑器能力增量](../reports/editor-capabilities-2026-09-22.md)。

2026-09-22 RC19 当前源码发行复核：`development-20260922-rc19` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC18 兼容回退，见[RC19发行报告](../reports/release-rc19-2026-09-22.md)。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC18 当前源码发行复核：加入快捷键偏好后的 `development-20260922-rc18` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC17 兼容回退，见[RC18发行报告](../reports/release-rc18-2026-09-22.md)。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC17 当前源码发行复核：`development-20260922-rc17` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC16 兼容回退，见[RC17发行报告](../reports/release-rc17-2026-09-22.md)。本轮还验证工作区主题/字号/密度偏好与默认布局备份恢复浏览器场景3/3。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-22 RC16 当前源码发行复核：`development-20260922-rc16` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建和 `sha256sum -c SHA256SUMS` 全部通过；独立包级冒烟验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复及 RC15 兼容回退，见[RC16发行报告](../reports/release-rc16-2026-09-22.md)。包内含当前文件目录树/列表与图标视图/框选、实例列表与卡片视图改动。发行证据仍不覆盖 Windows 真机、systemd、远程网络、物理故障或长时跨平台组合；全范围各项按下表继续保留进行中。

2026-09-21 最新当前源码复核：`development-20260921-rc15` 的 Linux/Windows Master 与 Daemon、SDK、Web 六包构建、SHA256 校验和独立包启动/TLS/登录/双 Daemon/停机快照恢复/兼容 RC14 回退冒烟均通过，见[RC15发行报告](../reports/release-rc15-2026-09-21.md)。RC15 当前源码常规真实浏览器套件 **49/49**（11.5分钟）通过，见[RC15浏览器回归报告](../reports/browser-regression-rc15-2026-09-21.md)；独立一小时 E08 也已通过，本机 p95 为48.1ms，见[恢复克隆优化小时报告](../reports/performance-hour-recovery-copy-2026-09-21.md)。全范围各项继续按下表保留进行中：这些本机结果不覆盖 Windows 真机、systemd、远程网络或物理故障。历史 RC14 失败与 RC13 证据仍保留在下方及[RC13报告](../reports/rc13-current-source-2026-09-20.md)。

2026-09-21 A09 RC15 当前源码定向 race 复验：真实TLS慢日志消费者、8.4MB跨节点传输、限时停止及归档/目标核对通过（13.918s），见[混合流报告](../reports/mixed-stream-control-2026-09-13.md)。长时全连接内存边界和跨平台负载仍未验证。

2026-09-21 E04 Linux进程崩溃子项：独立测试进程接受定时 occurrence/task 后遭强制终止，另一个进程重开 SQLite，accepted receipt 与任务/计划身份不变，旧 occurrence 不重放，撤权后的下一次触发被阻止；定向 race 1.130s、备份全包 race 2.351s通过，Windows amd64 测试程序交叉编译通过。见[定时调度进程强制终止报告](../reports/scheduler-process-crash-2026-09-21.md)。交叉编译不代表 Windows 运行；本测试也不覆盖事务中途掉电或 Master+Daemon 整体进程崩溃，E04/F12仍进行中。

2026-09-21 E03/F12 Linux恢复进程崩溃子项：真实 `Manager.Restore` 在8 MiB文件已有1 MiB写入暂存区时遭SIGKILL；重开服务后原ID为明确部分进度 `interrupted`，部分上传取消且清理、目标正文未发布，同ID不重放，新恢复计划重新恢复并核对SHA成功。定向race连续5次12.408s、备份全包race6.864s、`go vet ./internal/backup`及Windows测试程序交叉编译通过，见[恢复进程崩溃报告](../reports/backup-restore-process-crash-2026-09-21.md)。Windows运行、设备写入中断及物理掉电仍未验证。

2026-09-21 E04 时区边界证据复核：当前源码 `TestSchedulerTimezoneDSTAndFiveFieldContract` 覆盖纽约春季不存在的 02:30 不补跑、秋季重复的 01:30 对应两个不同 UTC slot；`TestSchedulerSkipPoliciesAndBoundedLongDowntime` 覆盖40天停机后的有限跳过策略。它们包含在本轮通过的 `go test -race ./internal/backup` 中，但不替代多日真实墙钟运行或设备掉电验收。

2026-09-21 E01 UTC跨日key顺序补验：节点和实例历史现有125点测试跨过UTC午夜，核实裁剪后的120点时间范围从次日00:00:01到00:02:00且最新实例样本未被错序；`go test -race` 定向两项通过（2.718s），见[历史测试源码](../../../internal/master/monitor_integration_test.go)。这是确定性时间戳边界验证，不是24小时墙钟采样；Windows和非Chromium仍未验证。

2026-09-21 E01 Firefox真实浏览器：Firefox 155.0/Playwright v1543 上，真实节点121点采样、离线 stale、Master重启后持久历史读取和 Daemon 恢复组合1/1通过（2.9分钟）；`BLORA_BROWSER=firefox` 经真实 HTTPS 双节点 fixture 验证。Chromium仍为默认引擎，Firefox临时profile需在沙箱外启动。见[长时监控报告](../reports/monitor-history-soak-2026-09-21.md)；Windows、WebKit、24小时实时运行仍缺。

2026-09-21 rc14 当前源码一小时混合负载已实际完成，但 E08 未通过：3606.506秒/45,072次响应采样，p95 **51.7ms > 50ms**、max159.4ms、1312/190831长帧；8窗口、双PTY、10,000项目录、16MiB循环复制均运行，162轮传输和末轮目标均通过。队列峰值86,193B，heap采样36.9–135.9MB；过程RSS/归档采样见[本轮报告](../reports/performance-hour-rc14-2026-09-21.md)。随后两次诊断工具开启的60秒样本均p95 47.4ms，但有测量扰动，不能覆盖小时失败；CPU热点分散，见[诊断报告](../reports/performance-diagnostic-rc14-2026-09-21.md)。直接写窗口几何的非视觉优化尝试，短测p95 57.0/54.9ms且仍超标，已回退；细节见[拖窗优化尝试报告](../reports/performance-drag-experiment-rc14-2026-09-21.md)。当前源码49项常规真实浏览器套件一次49/49通过，E08仍需找到可重复优化再决定小时复验。

2026-09-21 后续当前源码复验已达成本机 E08 一小时目标：恢复值深拷贝热点改用原生 `structuredClone` 并保留代理及 JSON 清洗回退；类型检查、50/50前端单测、生产构建、27/27相关真实浏览器测试通过。无 profiler 一小时场景 3,604.492秒、48,840响应样本，p95 **48.1ms**（≤50ms）、max160.4ms、731/197,137长帧，165轮跨节点传输均校验成功；62组资源采样含57组负载中样本，RSS峰值Master 39.4MB/Daemon 32.0MB，单终端归档16,774,920B、队列峰值86,180B，无采样/轮转竞态。详见[恢复克隆优化小时报告](../reports/performance-hour-recovery-copy-2026-09-21.md)及[原始指标](../reports/performance-hour-recovery-copy-2026-09-21.json)。这次通过取代RC14的本机p95失败，但不覆盖公网远端、高RTT/丢包、Windows或非Chromium场景。

2026-09-21 当前源码已重新打包为 `development-20260921-rc15`：Linux/Windows Master、Daemon、SDK、Web 六个归档 SHA 校验通过；独立包级 smoke 验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、状态备份恢复及兼容 RC14 包回退通过。详见[RC15发行报告](../reports/release-rc15-2026-09-21.md)。打包/快照证据不替代 Windows 真机、systemd、远端网络和物理故障验证。

2026-09-21 RC15 当前源码常规真实浏览器回归：Docker-enabled HTTPS 双节点隔离 fixture 上49项串行运行 **49/49 passed**（11.5分钟），含账号/备份/真实分钟调度/Docker/Compose/扩展/桌面/文件/PTY/通知/恢复/工作区；一小时 E08 单独执行，见[RC15浏览器报告](../reports/browser-regression-rc15-2026-09-21.md)。这次不验证 Windows 真机、独立systemd、远程高延迟、物理故障或E08以外的平台组合。

2026-09-21 E08优化阶段记录（历史过程态，已由后续完整复验更新）：CPU profile 将高成本定位到恢复模块 JSON 深拷贝，`copy()` 改为原生 `structuredClone` 并保留 JSON/代理回退语义；类型检查、50/50前端单测、构建及27/27恢复相关真实浏览器测试通过。无 profiler 的64.954秒真实混合样本1/1通过，p95 **45.2ms**、3/3565长帧、两个PTY及16MiB校验传输完成。当时准备继续完整一小时测试；最终小时结果为3,604.492秒、p95 **48.1ms**并通过，详见上方[恢复克隆优化小时报告](../reports/performance-hour-recovery-copy-2026-09-21.md)，E08不再处于等待复验状态。

2026-09-20 最新现代 Material 3 Expressive 整理：更换抽象壁纸及中性青灰配色，统一公共形状/颜色/状态，整理13个内置应用样式并修复减少动态效果对JS启动器的约束。完整浏览器38/38通过（1.9分钟），最后输入框修正复验3/3通过（10.8秒）；最终构建通过，真实截图34张，13组主要布局无差异，Dock四边14px。见[现代视觉报告](../reports/material-modern-2026-09-20.md)。本轮不提升后端和外部平台验收状态，旧rc11未包含最新UI。

2026-09-20 最新 Material 3 细化：依据官方形状/Expressive列表/复选框资料，统一内容区及表头圆角、列表状态、复选框外观，优化四个应用图标配色。构建通过，相关浏览器17/17通过（53.6秒，API替身）；真实本地截图24张及键盘勾选验证通过，13组主要布局无差异，Dock四边14px。见[细化报告](../reports/material-refine-2026-09-20.md)。源码/web/dist已更新，本轮未提升后端或外部验证状态，旧rc11未包含最新UI。

2026-09-20 最新 Material You 重构：整套扁平色面、胶囊控件、多色应用图标与几何壁纸落地，保留原布局。最终构建及完整浏览器37/37通过（1.8分钟，API替身）；真实本地截图22张，13组稳定态布局比较无差异，7种Dock场景四边14px、曲线误差小于0.01px。初轮启动器动画采样差异保留并修正后复验，详见[本轮报告](../reports/material-you-2026-09-20.md)。源码与web/dist已更新，本轮未提升后端/外部环境验证状态，旧rc11未包含本轮UI。

2026-09-20 最新薄片图标修订：按用户最新近景参考简化图标材质、重绘薄片前景，移除追光/光晕及厚重高光；构建退出0，选定浏览器7/7通过（44.6秒，API替身），真实本地截图22张。7种场景Dock四边均14px，曲线误差小于0.01px。见[薄片图标报告](../reports/thin-glass-icons-2026-09-20.md)。源码及web/dist已更新，本轮未提升后端或外部环境验证状态，旧rc11归档未含本轮UI。

2026-09-20 最新圆角对齐：图标和Dock共用连续轮廓，外框等距外扩14px；构建及选定浏览器7/7通过，真实渲染7种场景四边均14px、曲线距离误差小于0.01px，截图5张。见[圆角报告](../reports/concentric-corners-2026-09-20.md)。原外部环境验证边界及旧rc11归档状态不变。

2026-09-20 最新轻拟物修订：应用图标改为简洁玻璃叠层，工具图标使用 Lucide，账号菜单及内容字号/间距微调；最终构建及选定浏览器11/11通过，真实本地截图17张。见[图标与密度报告](../reports/glass-icons-2026-09-20.md)。本轮未重新验证全部后端/外部环境，旧rc11归档未含本轮UI。

2026-09-20 最新桌面视觉修订：独立三列顶栏、统一工具图标、账号菜单、单色窗口控制及光影/工作区域更新；构建和47单测通过，选定16浏览器场景经15+1修复复验通过，真实本地截图14张。见[本轮报告](../reports/modern-desktop-2026-09-20.md)。此轮未重新验证全部后端/外部环境，旧发行归档未含本轮UI。

2026-09-20 原生桌面UI重写（上一版记录）：旧主题与补丁层已替换，图标Dock、具象SVG图标、统一三色窗口控制、实例/文件侧栏列表完成；build、47单测通过，36浏览器场景经35+1修复复验通过，真实本地Master/双Daemon截图11张。见[UI报告](../reports/native-ui-2026-09-20.md)。未改变下述外部环境验证边界，旧rc11归档未含此UI。

2026-09-20 当前源码回归：提权环境执行 `make test`（`go test -race ./...`）退出0，所有Go包通过；受限沙箱的IPv6回环/helper失败不计作代码失败。见[全仓竞态报告](../reports/full-race-2026-09-20.md)。该证据仍不覆盖Windows真机、独立systemd、远程高延迟和物理掉电。

2026-09-20 续接审计：当前树重新执行 `make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及 47/47 单测退出0。源码扫描未发现生产路径空实现或遗留 TODO；未改变矩阵中外部环境缺口。

2026-09-20 文档证据审计：检查 112 份 Markdown 的相对链接，缺失链接为 0；历史截图已清理的报告改为明确记录“截图未保留”，不再提供失效证据链接。

2026-09-20 当前源码真实 Docker 复验：Engine 29.4.1/overlayfs 与隔离 fixture 可用，`TestRealDockerComposeLifecycle` race 退出0（27.681s）；真实容器/镜像/卷/网络、双流日志、Compose 更新删除、健康失败、卷保留和普通用户拒权均通过，随机带标签对象已清理。见[当前 Docker 验收](../reports/docker-current-2026-09-20.md)。远程 Engine、Windows 容器、物理掉电和长期故障组合仍未验证。

2026-09-20 systemd 环境复核：宿主 user bus 无法连接；不挂载宿主 cgroup 的隔离 systemd 容器退出255，特权宿主 cgroup 方案被安全审查拒绝且未执行。未修改宿主服务，真实 systemd service/timer 生命周期仍为环境阻塞，见[systemd 环境报告](../reports/systemd-environment-2026-09-20.md)。

2026-09-20 rc8浏览器与发行：前端完整 Chromium 替身回归 36/36（2.5分钟）通过，包含实例批量、传输详情和扩展跨设备修复；六包构建、外部 SHA、归档 MANIFEST 与独立冒烟退出0。见[浏览器回归](../reports/browser-regression-2026-09-20.md)和[rc8报告](../reports/release-rc8-2026-09-20.md)。

2026-09-20 rc9发行：同步最新验收台账后的六包构建、外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc9报告](../reports/release-rc9-2026-09-20.md)。

2026-09-20 rc10发行：续接审计后的当前源码和验收台账重新生成六包；外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc10报告](../reports/release-rc10-2026-09-20.md)。

2026-09-20 rc11发行：纳入当前源码 Docker 真实复验和最新台账；六包构建、外部 SHA、归档 MANIFEST 与独立冒烟全部退出0，见[rc11报告](../reports/release-rc11-2026-09-20.md)。

2026-09-20 构建复验：`make check`、`make windows`、前端类型检查及 47/47 单测均退出0；交叉编译只证明产物可构建，不替代 Windows 真机运行证据。

2026-09-19 F12/E04实际分钟调度：不注入时钟，两个相邻分钟每槽唯一备份任务、源版本更新、两归档分别恢复正文完整通过，1/1（1.6分钟），44800退出0，所属计划与服务已停止。见[实际分钟证据](../reports/schedule-wall-clock-2026-09-19.md)。多天墙钟和物理故障仍未验证，不将该子项当整项通过。

2026-09-19 F02/A01多设备与终端现场：真实 HTTPS 两浏览器上下文以不同设备身份竞争同一工作区，旧 revision 被拒绝且本地正文保留；同场景确认布局同步剥离终端检查点、显式内容同步携带终端检查点，`workspaces.spec.ts` 2/2（6.9s，27729）通过。见[多设备冲突证据](../reports/workspaces-2026-09-19.md)。断电迁移、存储耗尽和高延迟组合仍未验证。

2026-09-19 E09 Windows运行代码修复：修正多核 Windows Job Object CPU 配额换算，新增跨平台换算边界单测；Linux race、Windows交叉编译和静态检查通过，Windows Job/ConPTY 真机仍受环境阻塞。见[系统适配报告](../reports/system-adapters-2026-09-12.md)。

2026-09-20 E09发行验证：包含上述修复的 `development-20260919-rc5` 六包重新打包，SHA256SUMS 与独立包级冒烟（60112退出0）通过；Windows运行仍未验证。见[rc5发行报告](../reports/release-rc5-2026-09-20.md)。

2026-09-20 E09发行验证：包含定时备份一致性修复的 `development-20260920-rc6` 六包构建、SHA256SUMS/355项MANIFEST校验和独立包级冒烟（24537退出0）通过；Windows运行仍未验证。见[rc6发行报告](../reports/release-rc6-2026-09-20.md)。

2026-09-20 E09发行验证：包含默认备份应用定时一致性 UI 的 `development-20260920-rc7` 六包构建、SHA256SUMS/852项MANIFEST校验和独立包级冒烟（36781退出0）通过；Windows运行仍未验证。见[rc7发行报告](../reports/release-rc7-2026-09-20.md)。

2026-09-20 F12/E04定时一致性：Master 调度入口现接受 `save`、`pause`、`stop`、`hooks` 策略并限制固定钩子引用，`files` 仍拒绝钩子字段；真实固定 argv 保存程序由定时触发，前后钩子、资源身份、最新正文归档与任务结果均通过，定向集成 race 退出0（7.625s）。

2026-09-19 E08优化后小时复验：rc3默认速率/无profiler实际通过，8窗口、双PTY、万条目录与循环传输，p95 **46.5ms**、60,528次交互、220轮已验证传输及末轮校验；61次RSS/归档样本有界。保留945.7ms最大延迟及100/207818长帧，不称无卡顿；见[完整小时证据](../reports/performance-continuous-2026-09-19.md)。此前50.3ms失败仍保留；Windows和独立远程网络环境缺口不由本机结果覆盖。

2026-09-19 E09发行与恢复：rc3六包外部SHA-256、内部MANIFEST全部文件和依赖许可文本校验通过；独立SDK/示例构建签名、Master初始化/TLS/双Daemon、停机快照恢复分别在rc3和兼容rc2回退路径实际通过，身份/授权/任务回执/文件正文保留，备份后变更消失，详见[rc3报告](../reports/release-rc3-2026-09-19.md)与[依赖通知](../reports/dependency-notices-2026-09-19.md)。Windows仍仅构建，任意schema降级和物理故障未由此证明。

2026-09-19 E08小时终态：双PTY持续输出、180次传输轮换和末轮目标校验通过，但交互p95 **50.3ms** 超过50ms，整体退出1；完整数据与60次RSS/归档采样见[持续负载报告](../reports/performance-continuous-2026-09-19.md)。随后依据CPU采样优化完整恢复快照复制，保持同步日志和ACK时机；类型检查、47项单测、实际恢复故障3/3及隔离终端2/2通过，60秒诊断p95 45.4ms，JSON复制耗时由7.29秒降至2.06秒。见[快照优化](../reports/recovery-clone-2026-09-19.md)。仍需优化后的小时复验，不提升E08整体状态。

2026-09-19 F06/A06/A09/E08诊断修复：真实长时断流确认OUTPUT_GAP，前端逐条消费不足；相邻输出现有界合批解析/保护并累计ACK，保留resize顺序和原事件游标。TerminalApp/InstanceConsole非法浏览器关闭码及断流状态提示已修复。构建和突发输出/精确ACK/刷新/缺口提示浏览器2/2（18.3s）通过；高输出压力仍超出消费能力，默认输出速率小时复验正在运行，不提升整体状态。见[终端合批报告](../reports/terminal-batching-2026-09-19.md)。

2026-09-19 E08持续负载：30秒循环传输场景通过（p95 42.8ms），修正PTY负载持续时间后一小时首轮在9.3分钟失败：两路浏览器接收连续一分钟无增量，节点PTY和归档仍运行。后续600秒复现及提高输出速率的诊断已结束并保留失败，原因与修复见上条；当前重新运行默认速率小时场景，不能沿用短时结果声称小时通过。只读RSS/归档采样验证观察点内归档轮转且各低于16MiB，完整峰值仍未定。见[持续负载报告](../reports/performance-continuous-2026-09-19.md)。

2026-09-19 F13/E05执行目标边界：Linux服务动作仅接受`.service`，计划任务动作仅接受无路径`.timer`；拒绝通过有限服务接口操作其他unit类型或安装外部timer路径。系统适配race1.147s、真实TLS权限1.759s、Daemon拒绝非法目标回执2.688s通过，见[系统适配报告](../reports/system-adapters-2026-09-12.md)。真实systemd生命周期仍未验，此生产改动尚未进入rc2。

2026-09-19 F13/E05确认路径增量：真实私有firewalld应用后显式确认、实际Daemon重开仍保留runtime/permanent规则；普通用户403，同键回放200、异键409、终态修订不变。包含原失联回滚的完整场景race32.103s通过，见[私有防火墙报告](../reports/firewall-private-2026-09-19.md)。系统服务/计划任务、Windows及生产拓扑仍未由此证明。

2026-09-19 E03/E04补验：固定argv实际保存程序、保存失败补偿及归档恢复最新正文race6.174s；真实Daemon离线/重连、单次补跑及撤权race2.869s；定时备份接受后真实Master/SQLite重开、双节点重连且任务/归档唯一race3.198s。详见[备份调度证据](../reports/backup-scheduler-race-2026-09-14.md)。受控时钟不代替长期墙钟，服务重开不代替进程硬崩溃或掉电。

2026-09-19 F13/E05 真实防火墙：私有网络/挂载/用户namespace、独立D-Bus与firewalld中，真实TLS双Daemon应用后阻断测试Master端口，未确认租约按期限恢复runtime/permanent原配置，连接恢复后任务明确FAILED/confirmation_expired_rolled_back；普通账号403，原服务/端口范围保留。race25.998s通过，详见[私有防火墙报告](../reports/firewall-private-2026-09-19.md)。未接触宿主防火墙，Windows及系统服务/计划任务真机组合不由此证明。

2026-09-19 A12/E03 Linux真实ENOSPC：私有1MiB tmpfs、真实TLS双Daemon竞态验证跨节点移动、备份创建、备份恢复的空间不足；源和既有目标保留、未完成数据不发布、暂存清理均通过。传输+备份创建6.505s、恢复3.678s，详见[空间不足报告](../reports/enospc-2026-09-19.md)。不以tmpfs替代物理掉电、Windows或远程文件系统证据。

2026-09-19 F04/F09/A09 突发任务修复：控制接收队列满时改为有界等待，避免短时存储消费落后导致断链及任务接受记录缺失；协议真实TLS race 1.192s通过，重编译双节点后真实Chromium通知/分页2/2（7.9s），明确核对101任务全部SUCCEEDED。见[控制背压证据](../reports/control-backpressure-2026-09-19.md)。长期及跨平台边界仍未验证。

2026-09-19 F04/A05 节点真实浏览器复验：修正测试的应用启动入口后，维护、配额、配置草稿和票据下载定向 1/1（4.7s）；完整节点文件 2/2（53.1s）通过，包含真实 Daemon 暂停后的失联/WAITING_NODE及恢复。详见[全套回归报告](../reports/full-browser-regression-2026-09-13.md)。不改变其他平台和长期组合的未验证状态。

2026-09-19 F14/E06 真实取消恢复：私有双节点 HTTPS + Chromium 1/1（4.6s）通过，服务端接受取消后中止浏览器响应，刷新不重放，恢复暂停节点后任务与界面为 CANCELLED，原取消键保留。详见[扩展任务重试](../reports/extensions-task-retry-2026-09-19.md)。该项补齐此前取消丢回执/刷新缺口；未要求 WASI 已开始执行，Windows、远程长期与运行中中断组合仍不由此证明。

2026-09-19 F14/E06 查询恢复增量：参考扩展任务 GET 失败后可显式刷新状态，读取请求去重且不重放创建/取消；独立包 Chromium 1/1（8.5s）通过，使用模拟 HTTP 后端。证据见[扩展任务重试](../reports/extensions-task-retry-2026-09-19.md)，不改变真实故障组合的未验证状态。

2026-09-19 F14/A13/E06 请求身份增量：SDK 创建/取消任务均显式传递稳定请求键，参考扩展在发送前保存待提交内容和键。创建任务真实 HTTPS 双节点浏览器丢响应、刷新、同键重试 1/1（9.2s）通过；取消回执真实 TLS 集成 7.018s 通过，终态回放不推进修订；取消浏览器恢复使用模拟 HTTP 后端，单场景 1/1（8.7s）。更新后的完整扩展浏览器套件 6/6（20.1s）、前端 47/47、SDK/参考扩展构建通过。证据见[扩展任务重试](../reports/extensions-task-retry-2026-09-19.md)。真实节点取消丢回执/刷新组合、Windows 与远程长期场景仍未验证，整项状态不提升。

2026-09-14 F12/E04 调度写入请求回执收口：计划更新与删除的版本 CAS、记录变更和 SQLite 请求回执在同一事务内提交；同键更新/删除可回放，不同载荷返回 `IDEMPOTENCY_CONFLICT`，前端删除保留稳定请求键。调度定向 race、提权全仓普通回归（Master 128.030s）和竞态回归（Master 267.872s、extensions 149.546s）均退出 0，详见[请求键报告](../reports/request-keys-2026-09-14.md)。Windows、长期离线与物理故障场景仍按 F12/E04 行保留未验证。

2026-09-14 final46 交付复验：调度回执事务修正及全仓回归证据纳入 `dist/releases/development-20260914-final46`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `18cc8b4e1013c968924f4159eecfeb1925fff78b14e57906461b928989c3b537`，冒烟临时目录 `/tmp/blora-release-smoke-94mpn3me` 已停止进程。

2026-09-14 final47 交付复验：重启级调度回执测试与 OpenAPI 删除契约纳入 `dist/releases/development-20260914-final47`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `24e852ccf0fe2a6cba417ddb45217ca25cae98a75c087e2a1b7107a814af4ca2`，冒烟临时目录 `/tmp/blora-release-smoke-jzdy74x5` 已停止进程。

2026-09-14 F13/E05 默认应用请求键现场：系统服务/计划任务、防火墙应用/确认及进程终止前端均保存稳定请求键，提权 Chromium 系统管理浏览器场景 3/3（14.6s）、前端 46/46、类型检查和生产构建通过；真实危险主机和 Windows 行为仍按矩阵保留未验证。

2026-09-14 final48 交付复验：默认应用请求键修正及调度回执证据纳入 `dist/releases/development-20260914-final48`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 全部退出 0。`SHA256SUMS` SHA-256 为 `580112f9bd3cc8738162a91fe800fb151bf8ac9c3b5f027e4c890ac63b48aee4`，冒烟临时目录 `/tmp/blora-release-smoke-b1wqtbhw` 已停止进程。

2026-09-14 F09/A09 取消请求幂等收口：任务收据持久化首次 `cancellationRequestId`，普通任务、上传和跨节点传输取消均在事务中记录；并发取消只提交一次，终态重试返回原收据。存储定向取消测试与真实 TLS Master/双 Daemon 集成回归通过，任务中心为每个任务保留稳定请求键，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 F03/F04 节点登记幂等收口：登记票据生成绑定管理员、请求键和节点名称，同键重试只返回原票据，不同名称明确返回 `IDEMPOTENCY_CONFLICT`；存储和真实 TLS 双 Daemon 管理回归通过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 登记回执修正后的全仓复验：提权 `go test -p 1 ./... -count=1` 与 `go test -race -p 1 ./... -count=1` 均退出 0（Master 普通 126.994s、竞态 261.687s），`make check` 与前端 46/46 通过；未留下测试资源。

2026-09-14 final38 交付复验：登记票据幂等修正及全仓证据同步后生成 `dist/releases/development-20260914-final38`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `ce76c1e52dfeb38627f2181c8751cb15048c538f0a025f05c568d7c2ad240628`，冒烟临时目录 `/tmp/blora-release-smoke-rrefkg26` 已停止进程。

2026-09-14 F04 请求键继续收口：节点换钥票据以管理员、节点、请求键和 `reactivate` 绑定持久化回执，同键回放原票据，不同载荷返回 `IDEMPOTENCY_CONFLICT`；真实 TLS 换钥集成及换钥后的代次/旧连接栅栏通过。改动后全仓普通 128.965s、竞态 262.770s 均退出 0，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 final39 交付复验：换钥票据幂等修正及全仓证据同步后生成 `dist/releases/development-20260914-final39`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `e63dd2e0ae2c9fbd25e7fba2abf245fb8027984f23d5c07f13939eedc8647728`，冒烟临时目录 `/tmp/blora-release-smoke-wsor4qlc` 已停止进程。

2026-09-14 final40 交付复验：将 final39 换钥证据与文档同步后生成 `dist/releases/development-20260914-final40`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `1739590e081a8fc043f914e82f5adda7c998e759df1e9fb3fb50a01c787d9fce`，冒烟临时目录 `/tmp/blora-release-smoke-ghsj5ufc` 已停止进程。

2026-09-14 F03/F05/F06/E09 请求键边界再收口：实例创建、实例控制、终端创建和授权写入均在读取请求体前拒绝缺失/非法键；定向回归 6.241s，最新 Master 普通 127.825s、竞态 265.970s 均退出 0，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 final41 交付复验：请求键边界证据及文档同步后生成 `dist/releases/development-20260914-final41`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `37996802de1cc4510380974f8d60cbab69ca9513e9371c952be80dd631c7c286`，冒烟临时目录 `/tmp/blora-release-smoke-rwgm48yh` 已停止进程。

2026-09-14 E02/F11 请求键边界补齐：Compose 项目保存 prepare/chunk 写入在解析正文前要求稳定键，前端按准备键/分片偏移发送；容器管理定向回归 0.377s，前端 46/46，最新 Master 普通 129.598s、竞态 264.964s 均退出 0。需要显式 Docker 镜像的 Compose 真实端到端用例因环境缺失跳过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 final42 交付复验：Compose 请求键修正及最新回归证据同步后生成 `dist/releases/development-20260914-final42`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `b6a8cc499522d141cbe05e11fdd7b1878b9ddc5e17d6f465ff4be9ca7eddf73e`，冒烟临时目录 `/tmp/blora-release-smoke-k7gfk8j5` 已停止进程。

2026-09-14 final43 交付复验：Compose OpenAPI 请求键声明同步后生成 `dist/releases/development-20260914-final43`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `78520382738ab277e285a3cd084fd22b9ea8214d2f1aa802a9e828652fcee24d`，冒烟临时目录 `/tmp/blora-release-smoke-zg38nbet` 已停止进程。

2026-09-14 F04 换钥冲突错误码修正：同键不同载荷明确返回 `IDEMPOTENCY_CONFLICT`（409），真实 TLS 换钥集成回归 0.209s 通过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 final44 交付复验：换钥冲突错误码修正同步后生成 `dist/releases/development-20260914-final44`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `15f56cc3151cc38b24ab0d2a916bff6342a9be27055f008bb259231a68deed70`，冒烟临时目录 `/tmp/blora-release-smoke-nn8w4hzn` 已停止进程。

2026-09-14 F03/F04 管理写入错误码统一：用户、角色、节点设置和换钥等持久化请求键冲突统一返回 `IDEMPOTENCY_CONFLICT`（409）；定向回归 1.198s，最新 Master 普通 127.672s、竞态 263.528s 均退出 0。首次并行回归的 `/tmp` 空间耗尽已清理后复跑通过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 final45 交付复验：管理写入口幂等冲突错误码统一后生成 `dist/releases/development-20260914-final45`；独立 `package-smoke.py`、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c SHA256SUMS` 均退出 0。`SHA256SUMS` SHA-256 为 `e332a19b9c80ea6fdecdfc1638eaca3583dd547ac3e8aa787203439dec120195`，冒烟临时目录 `/tmp/blora-release-smoke-8pjxh2wo` 已停止进程。

2026-09-14 取消收据改动后全仓复验：提权 `go test -p 1 ./... -count=1` 退出 0（Master 124.958s）；完整 `go test -race -p 1 ./... -count=1` 复跑退出 0（Master 255.904s、extensions 137.279s、terminal 18.616s、runlog 2.136s），`make check` 与前端 46/46 通过。首轮竞态的 runlog `/proc` 时序失败已隔离重跑通过并如实保留。

2026-09-14 F03/F04/F06/E09 请求键补齐：用户创建/改密、节点登记/身份轮换、云端工作区 PUT/DELETE 及通用任务/上传取消在授权后统一拒绝缺失或非法 `Idempotency-Key`；用户与工作区事务写入支持稳定回放。Master/Storage 定向 race、传输取消回归、`make check` 与 OpenAPI 3.1（103 路径、120 操作）解析通过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 防火墙确认幂等收口：确认请求键传入并持久化于 Daemon 租约和成功任务结果；同键在租约清理前后均可安全回放，其他键明确冲突且不重复触碰主机规则。Daemon 定向 race 29.354s、Master 定向 race 2.096s、改动后串行全仓普通与 race 均退出 0；详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 防火墙确认 CAS 并发补验：Daemon 重启后无内存租约的双键并发确认仅一个成功，获胜键可回放，另一键返回 `FIREWALL_LEASE_CONFLICT`；定向 race 1.137s、改动后串行全仓普通与 race 均退出 0，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 当前源码串行全仓复验：`go test -p 1 ./... -count=1` 明确退出 0，Master 126.349s，runlog、runtime、storage、terminal、extensions 及其余 Go 包全部通过；不把此前并行时序失败记录改写为无条件全绿。

2026-09-14 当前源码串行全仓竞态复验：`go test -race -p 1 ./... -count=1` 明确退出 0，Master、extensions、terminal、runlog、runtime、storage 及其余 Go 包全部通过；Windows、远程节点与物理故障仍按矩阵保留未验证状态。

2026-09-14 文件请求键边界收口：文件保存、文件动作、新建上传均在实例写权限确认后、请求体解析前统一检查 `Idempotency-Key`；空白保存键与 129 字节动作键集成回归及竞态回归（5.406s）通过，完整 Master 126.836s 与 `make check` 退出 0，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 上传取消测试契约同步：两个历史集成用例补充稳定请求键；上传取消定向回归 1.562s、完整 `go test ./internal/master -count=1` 125.538s 退出 0。

2026-09-14 改密同键回放修正：当前密码只在首次事务回调校验，撤销会话后重新登录仍能复用原回执；改密定向回归 1.044s、完整 Master 套件 127.375s 退出 0。

2026-09-14 final22 交付复验：提权环境完整 Master suite（125.544s）及其余 Go 包通过；全仓首轮仅因备份用例一次 `terminal access denied` 失败，单独复验和后续 Master 重跑通过。前端 46/46 单测与 `make check` 通过；`development-20260914-final22` 打包、独立冒烟及六个归档校验通过，见[发行包报告](../reports/package-2026-09-13.md)。

2026-09-14 final23 文档同步交付：请求键 race 最终时间写入证据后重新生成 `development-20260914-final23`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `7bab5ab5c69dda7cc0fef737ee7885af9b7caaebec3c9fc883ee5164d2eed2cd`）。

2026-09-14 final24 记录同步交付：修正全仓测试一次环境波动的说明后重新生成 `development-20260914-final24`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `ed434701601b269098053b261a69ce0e785aea7d6021f971367839424996cefc`）。

2026-09-14 final25 取消请求键交付：通用任务与上传取消新增授权后请求键检查，传输取消回归通过；重新生成 `development-20260914-final25`，打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `466f6d3e6a2da7c3ab37beb389ca45e3c4dae5f87785cbf0788b738a8311a5fa`）。

2026-09-14 final26 交付复验：同步上传取消历史用例请求键后，完整 Master 套件 125.538s 通过；重新生成 `development-20260914-final26`，打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `272c3f3bef10d4a380394ba23d2f99692b2328af4f63a12378ae79958f8a6e55`）。

2026-09-14 final27 文档归档复验：将 final26 记录纳入发行包后重新生成 `development-20260914-final27`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `a89541bb7657d7174d454828a7cea77477c9aaa73649b37ff85d55fe10f5faa`）。

2026-09-14 final28 改密幂等交付：自助改密回执在会话撤销后可同键重放；改密定向与完整 Master 套件通过。`development-20260914-final28` 打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `efc9e4dda49e234b47d2037990331f726af7b674094b62667aa941c4c0c4d030`）。

同日收尾复验：`go test ./internal/storage -count=1` 0.579s、`make check`（`go vet ./...`）均退出 0；调试残留及活动测试资源检查为空。

2026-09-14 final29 归档：将最终改密收尾证据纳入包内后生成 `development-20260914-final29`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `12171c6ed4aa430da966bb63e1b6ca3b20767985df2871667f17b36fba63af2a`）。

2026-09-14 final30 归档：文件保存、文件动作、新建上传入口的请求键边界收口后生成 `development-20260914-final30`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `f7630947960072aded20d64f7a64397c8737bf125599f28fd8f661d143a5802d`）。

2026-09-14 final31 归档：将文件请求键竞态回归证据纳入包内后生成 `development-20260914-final31`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `18ff79189be48c49c93bb8c0a5330827472726f8f242b54eb7a44c6c28db3e80`）。

2026-09-14 final32 归档：将串行全仓 Go 回归证据纳入包内后生成 `development-20260914-final32`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `06e336ee2ec13240f2d7b86f6c83b638d5ec42db8cf2d9b0de35289aaae25c96`）。

2026-09-14 final33 归档：将串行全仓 Go 竞态回归证据纳入包内后生成 `development-20260914-final33`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `84931b88163ef6182ca4504b9d26105b427c4c16fe569f783c2807223abbe34e`）。

2026-09-14 final34 归档：防火墙确认请求键持久化与终态回放修正后生成 `development-20260914-final34`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `83dcf17fc84253c0e8731056497a7b0f83b4e2eb5f1768ede6d033260dfccd15`）。

2026-09-14 final35 归档：重启后无内存租约的确认 CAS 并发修正及回归证据纳入 `development-20260914-final35`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `de8c519f6af02db48e41129150d41cef7805bd4bdeaa14b08db99ec3bcb1dd4d`）。

2026-09-14 final36 归档：任务取消收据幂等修正及最新全仓普通/竞态证据纳入 `development-20260914-final36`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `50fb039b87dc00bef14d1006f0e6ebcf929f407a1790efe51426d0d77ff80913`）。

2026-09-14 final37 归档：将 final36 进度/矩阵记录同步进包后生成 `development-20260914-final37`；打包、独立冒烟和六个归档校验全部通过（`SHA256SUMS` SHA-256 `4fbae56157ee6ed406d6f8deb91e584be61fe44b130f05d16669a42e691a59a3`）。

2026-09-14 收尾资源检查：无 Blora/fixture/Vite/Playwright 活动进程，Docker `blora.test=1` 标签对象为空；这只证明本轮测试清理完成，不改变 Windows 真机、远程节点/Engine、物理故障、长期调度和跨平台组合的未验证状态。

2026-09-14 最新普通全仓回归：Master（127.270s）及其余 Go 包通过；`internal/runlog/TestStdinSurvivesDaemonAndOutputEOF` 首次并行时序出现空输入，单独重跑 0.052s 通过，未改写为全仓无条件通过。

2026-09-14 请求键边界收口：统一入口拒绝空白及超过128字节的 `Idempotency-Key`，实例控制在动作解析、资源授权和节点门禁后、任务接纳前检查；后端定向回归、前端类型检查与 46/46 单测通过。`development-20260914-final10` 发行包构建及独立冒烟通过（SDK/参考扩展离线构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE）。

2026-09-14 最终 race 复验：清理任务缓存后提权环境 `go test -race -p 1 ./...` 退出 0（Master 251.165s、extensions 136.422s、terminal 18.729s，其余 Go 包通过）；此前一次 `/tmp` 空间耗尽的复跑保留为环境失败记录。

2026-09-14 传输步进优化后 race 复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-race-final14 GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 退出 0（Master 248.686s、extensions 136.133s、terminal 18.695s，其余 Go 包通过）。

2026-09-14 幂等键边界回归：`TestRequireRequestIDBounds` 覆盖缺失、空白、129 字节拒绝和128字节上限接受；`development-20260914-final11` 发行包构建及独立冒烟通过。

2026-09-14 授权顺序修正后的 E09 复验：`development-20260914-final12` 发行包构建及独立冒烟通过；实例控制保持动作解析和资源授权优先，再执行请求键与任务接纳检查。
2026-09-14 A09 长时采样后的 E09 发行包复验：`development-20260914-final13` 构建及独立冒烟通过，独立 SDK/参考扩展、Master 初始化/TLS/登录和双 Daemon ONLINE 均完成；发行包校验见 `dist/releases/development-20260914-final13/SHA256SUMS`。
2026-09-14 A13/E06 浏览器权限复验：真实 `extensions.spec.ts` 完整套件 6/6（22.6s）通过，新增成员上下文授权边界场景同时通过；fixture 已停止并完成清理。

2026-09-14 E09 传输优化后发行包复验：`development-20260914-final14` 构建及独立冒烟通过；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 完成，校验见 `dist/releases/development-20260914-final14/SHA256SUMS`。

2026-09-14 E09 文档同步发行包复验：因 README/docs 更新重新生成 `development-20260914-final15`；独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档 `sha256sum -c` 均通过，校验见 `dist/releases/development-20260914-final15/SHA256SUMS`。

2026-09-14 E09 长时性能证据发行包复验：加入 `BLORA_PERF_SOAK_SECONDS` 说明与60秒采样证据后生成 `development-20260914-final16`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final16/SHA256SUMS`。

2026-09-14 E09 采样超时预算修正发行包复验：`development-20260914-final17` 构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final17/SHA256SUMS`。

2026-09-14 E09 Windows 适配编译复验：`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c` 分别编译 `internal/runtime`、`internal/runlog`、`internal/systeminfo` 测试入口，均退出0；当前环境无 Wine/Windows 真机，运行行为仍未验证。

2026-09-14 E09 交叉编译证据发行包复验：`development-20260914-final18` 构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，校验见 `dist/releases/development-20260914-final18/SHA256SUMS`。

2026-09-14 E03/E04 备份与调度竞态复验：提权隔离环境 `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（backup 1.856s、Master 5.816s），固定 argv Hook `go test -race ./internal/daemon -run 'TestBackupHook' -count=1` 退出0（2.039s）；真实 Chromium 备份/恢复/调度入口 1/1、3.5s 通过；长期离线、具体应用 hook 语义和物理 ENOSPC/掉电仍保留，见[浏览器证据](../reports/backup-browser-2026-09-14.md)。

2026-09-14 E09 最新交付包：因备份/调度证据更新生成 `development-20260914-final19`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `7e365582aad6a03cc84e1d7b7342ce795da2f903036f0db059f37875385bc9b2`。

2026-09-14 E09 最新交付包：因真实备份浏览器证据更新生成 `development-20260914-final20`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `4e25104368473ca06daaae78870b9ee8b015e5e0ebbda665ad26be116a45c60b`。

2026-09-14 E09 交付包收口：将备份/调度报告纳入包内后生成 `development-20260914-final21`；构建、独立冒烟、SDK/参考扩展签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均通过，`SHA256SUMS` SHA-256 为 `9cb7f53ce64e6703003054bdcce2af95a7b5c392b59b0de9debde1e9b2e5282d`。

2026-09-14 F03/F13/F14 请求键边界修复：扩展安装/升级/回滚/启停/卸载、目录安装和防火墙租约确认均在授权后统一拒绝无效 `Idempotency-Key`；Master 定向 race 77.219s、OpenAPI 3.1 解析通过，详见[请求键报告](../reports/request-keys-2026-09-14.md)。

2026-09-14 E08 批量步进复验：重编译夹具纳入 `transferChunksPerStep=16` 后，真实 Chromium 混合场景 1/1 通过（fixture-1569035732，约2.0分钟），512MiB 用时110.868s（约4.62MiB/s），反馈p95 39.0ms、0/285长帧、目标校验成功；此前旧夹具133.246s仅作索引基线；长期内存/远程网络/Windows及物理故障仍保留。

2026-09-14 E08 长时采样：`BLORA_PERF_SOAK_SECONDS=60` 的真实 Chromium 混合场景 1/1 通过（fixture-2822501719），采样65.697s、p95 43.3ms、1/3770长帧、堆34.6–97.9MB、未确认队列峰值86,867B，512MiB传输109.245s；小时级稳定性、远程网络、Windows及物理故障仍保留。

2026-09-14 增量：扩展跨设备副本场景 1/1 通过（Linux Chromium，状态版本 1→2）；恢复校准返回规范资源身份，删除/撤权资源保留窗口与草稿。详见[扩展跨设备报告](../reports/extensions-cross-device-2026-09-14.md)。该证据不覆盖 Windows、远程高延迟和物理故障；未授权浏览器组合另见[权限报告](../reports/extensions-unauthorized-browser-2026-09-14.md)。

2026-09-14 回归记录：最新一次完整 `make test` 在并行资源压力下未通过（Daemon 防火墙命令、Master 审计登录、runtime 后代条件、PTY 输出各一项时序失败）；四项随后低并发定向 race 均通过，不能将该次全仓命令记为通过。此前全仓通过证据仍保留其对应代码时点。

2026-09-14 race 复验：改用 `go test -race -p 1 ./...` 串行执行，退出 0（Master 259.481s，其余包通过）；该结果与上条并行压力失败记录同时保留。

2026-09-14 发行包与请求键收口：调度创建/更新入口统一拒绝空白 `Idempotency-Key`；`development-20260914-final7` 打包及独立冒烟均通过，包含 SDK/参考扩展离线构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE。

2026-09-14 调度请求边界：请求键校验提前到请求体解析前，定向 race 通过；`development-20260914-final8` 打包及独立冒烟均通过，产物包含该最终源码。

2026-09-14 持久化修改请求键复验：实例/节点设置、用户/角色变更、节点撤销、扩展数据/资源写入、备份计划/恢复和传输取消均在授权后统一拒绝空白 `Idempotency-Key`，相关 Master 集成和提权串行全仓 race（Master 252.584s）通过；OpenAPI 同步声明对应参数。

2026-09-14 E09 增量：统一请求键改动后的 `development-20260914-final9` 发行包构建与独立冒烟通过；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功。

2026-09-14 A09 并行流复验：`TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop` 在真实 TLS Master/双 Daemon 下 race 通过（13.07s）；慢日志未消费峰值 82,635B、接收队列 79,507B，8.4MB 跨节点传输与停止并行，停止和归档核验成功。长时内存采样仍未覆盖，A09 继续保留进行中。

2026-09-13 A16完整入口身份复验：真实目标改名、管理员停止资源墓碑删除/同名替换后旧resourceRef显示不可访问且不指向新资源1/1（7.0s），双账号撤销instance.read后旧入口失效1/1（3.3s），暂停Daemon时入口显示节点失联并恢复原runId1/1（50.4s）；结合专用窗口复用/显式新开和入口改名删除1/1（3.6s），Linux Chromium覆盖全部A16，见[入口报告](../reports/shortcut-lifecycle-2026-09-13.md)。

2026-09-13 E06/E07/A13真实扩展复验：Linux Chromium HTTPS fixture 上独立扩展全套 5/5（24.2s）通过，覆盖通知来源与刷新去重、元数据草稿与真实实例写入、v1→v2 数据迁移/不兼容拒绝/失败保留/回滚、真实节点 WASI 任务与资源窗口/快捷入口/移窗/关闭、本地注册源 v1.0.0/v1.1.0 安装升级及禁用门禁；见[扩展全链路报告](../reports/extensions-real-2026-09-13.md)。Windows 真机、远程高延迟源和物理故障仍保留。

2026-09-13 E09最终发行包复验：加入实例删除入口后使用不可变标签 `development-20260913-final` 重新生成六个归档；独立解包冒烟在受限提升环境 1/1 通过，SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，见[发行包报告](../reports/package-2026-09-13.md)。Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-13 E09发行包最终复验：监控 stale 诊断和前端提示更新后使用不可变标签 `development-20260913-final3` 重新生成六个归档；独立解包冒烟在受限提升环境 1/1 通过，SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，见[发行包报告](../reports/package-2026-09-13.md)。Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-13 系统管理能力门禁复验：系统服务、计划任务和防火墙操作按钮现在随节点能力探测禁用，能力不可用时不再允许提交；前端 check/test/build、系统管理浏览器 2/2（54.2s，分别覆盖能力可用与不可用）及 `development-20260913-final4` 独立发行包冒烟通过。危险主机、Windows 真机及远程故障组合仍保留。

2026-09-13 任务接纳边界：主机终端、容器/Compose、系统管理、进程终止、实例终端、扩展任务和跨节点传输入口统一在授权后要求 `Idempotency-Key`，缺失时返回 `REQUEST_ID_REQUIRED`，防止空键请求先触发远端检查或产生不可对账任务；定向 Master 集成回归退出 0。

2026-09-13 控制台输入幂等键补齐：实例日志控制台输入也在授权后、读取运行代次前要求 `Idempotency-Key`，并同步 OpenAPI 必填参数；定向测试覆盖空键 400、合法请求单次交付和独立停止输入，退出 0（1.620s）。

2026-09-13 E09发行包收口：控制台输入请求键边界加入后，以不可变标签 `development-20260913-final5` 重新生成六个归档并完成独立解包冒烟；SDK/参考扩展离线构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均成功，Windows 真机与真实业务/物理故障组合仍未验证。

2026-09-14 全仓竞态复验：控制台输入边界加入后，提权 `make test`（`go test -race ./...`）退出 0，Master 257.810s，其余 Go 包通过；Windows 真机、物理 ENOSPC/掉电、远程高延迟和长期负载仍按各项保留。

2026-09-13 系统能力反馈：SystemApp 按节点能力探测停止不可用服务/计划任务轮询，并对防火墙不可用状态给出原因、阻止预览/应用提交；可用能力路径及系统管理浏览器 1/1 保持通过。

2026-09-13 日志流边界修复：慢消费者超过 45 秒时，Master 在 `SLOW_CONSUMER` 错误帧成功写入后保留 100ms 有界交付窗口，避免客户端只收到 EOF；慢消费者游标重连测试连续 2 次通过，race 复验亦通过。

2026-09-13 云端工作区冲突提示：同步遇到 `WORKSPACE_CONFLICT` 时前端保留本地现场并提示用户刷新列表后选择副本，不覆盖草稿或误报旧云端版本成功。

2026-09-13 A05任务等待增量：真实离线节点提交stop后原任务WAITING_NODE，生产任务卡显示等待节点；恢复后同taskId成功且实例实际停止，浏览器1/1（50.7s）通过，见[失联报告](../reports/node-offline-browser-2026-09-13.md)。

2026-09-13 A05浏览器增量：暂停精确识别的fixture Daemon至真实心跳失联，页面显示等待确认而非实例停止，恢复上线runId不变，1/1（50.5s），见[失联浏览器报告](../reports/node-offline-browser-2026-09-13.md)。

2026-09-13 A10任务中途增量：独立Daemon实际STOPPING阶段SIGKILL后，原任务INTERRUPTED并保留诊断阶段，同key重试不重放，原runId/启动计数保留，脚本退出0，见[Daemon崩溃报告](../reports/daemon-crash-2026-09-13.md)。

2026-09-13 A10增量：独立Daemon SIGKILL后同身份目录启动，新startupId恢复原runId、启动计数不增加，随后可重启原运行；独立包脚本退出0，任务中途Daemon崩溃仍待补，见[Daemon崩溃报告](../reports/daemon-crash-2026-09-13.md)。

2026-09-13 A04硬中断增量：独立解包Master在restart RUNNING时SIGKILL，原状态重启后同requestId复用taskId、任务成功且实际启动计数仅初次+一次重启；脚本退出0，见[重启恢复报告](../reports/restart-disconnect-2026-09-13.md)。

2026-09-13 A04/A05增量：实际重启任务RUNNING时断开源节点管理连接，新generation重连后同requestId仍复用原taskId，原任务成功、后续重连不改变runId；定向race7.392s通过，见[重启断线报告](../reports/restart-disconnect-2026-09-13.md)。Master进程中断等组合仍保留。

2026-09-13 A02增量：真实双编辑窗口共享正文、另一账号修改服务器版本、冲突拒绝/刷新保留/显式采用基线保存1/1（7.7s）通过；独立光标滚动及迟到响应组合仍待补，见[共享编辑报告](../reports/shared-editor-conflict-2026-09-13.md)。

2026-09-13 A08增量：真实成员活跃PTY撤销input/read，输入租约失效、真实文件未增加副作用、新请求403、刷新旧快照仍拒权，1/1（6.1s）通过；见[浏览器撤权报告](../reports/browser-revoke-2026-09-13.md)。

2026-09-13 A09并行增量：真实慢WSS日志读取端、8.4MB跨节点传输及停止控制同时运行，停止在5秒预算内完成、归档可重读且传输目标逐字节验证通过；定向race14.120s，长期内存/归档预算仍待量化，见[混合流报告](../reports/mixed-stream-control-2026-09-13.md)。

2026-09-13 A17编辑器增量：实际鼠标拖出/合并并立即刷新，视图ID、归属、顺序、中文未保存正文及撤销重做保持，1/1（3.8s）通过；活跃PTY同组合仍待补，见[拖拽报告](../reports/drag-recovery-2026-09-13.md)。

2026-09-13 A11 增量：真实桌面配额写入失败时最新草稿导出、存储恢复后刷新、旧schema0迁移及未知schema999原记录跨刷新保留，2/2通过（4.6s）；快照事务中断组合仍待补，见[恢复故障报告](../reports/recovery-quota-2026-09-13.md)。

2026-09-13 A06 增量：真实 vim 中文未保存缓冲区刷新后保存准确、top 刷新与退出、WSS 输入不重放 1/1 通过（9.0s），见[交互程序报告](../reports/terminal-programs-2026-09-13.md)。E09 解包 SDK 离线干净安装/构建/签名及 Master/双 Daemon 独立上线也通过，见发行包报告；Windows 真机仍未验证。

2026-09-13 E09 交付增量：六个本地发行包构建、重复打包逐字节一致及归档清单/摘要核验通过；解包独立运行和 Windows 真机仍待验证，见[发行包报告](../reports/package-2026-09-13.md)。

2026-09-13 A07 Linux真实场景通过：解压与16MiB本机上传并行关闭标签，解压10,000文件独立完成；新标签从任务中心恢复原上传，于983,040B继续并验证成功。新增账号/请求/资源/文件身份保护，前端46单元与构建通过，见[关闭网页证据](../reports/upload-close-2026-09-13.md)。

2026-09-13 A07/E08最终构建复验：带 `--performance` fixture 的真实关闭网页/上传续传 1/1（1.4m），原 uploadId 从917,504B续传、16MiB摘要和10,000项解压均核对；混合负载 1/1（27.8s），p95 36.2ms、0/298长帧、八窗口/双PTY/万条目录/16MiB传输全程并行，见[关闭网页证据](../reports/upload-close-2026-09-13.md)与[真实性能报告](../reports/performance-real-2026-09-13.md)。

2026-09-13 F07/A12/E08大文件增量：512MiB真实跨节点传输和独立摘要验证通过，约0.942MiB/s；同一后台任务跨浏览器中断续接，采样p95 42.4ms。源前缀改写并恢复mtime在发布前拒绝，实际源Daemon重启重建索引通过，见[传输证据](../reports/transfer-proof-2026-09-13.md)。新增[数据库版本保护与运维文档](../reports/upgrade-storage-2026-09-13.md)；完整存储race通过，物理掉电和Windows真机仍保留。

2026-09-13 E08短时真实场景通过：最终p95 38.8ms、0/302长帧，八窗口/双PTY/万条目录/16MiB跨节点传输全程并行；终端增量恢复和软件GPU降级真实复验通过。大文件吞吐及长期/远程网络仍待补，整项保持进行中，见[最终证据与失败历史](../reports/performance-real-2026-09-13.md)。

2026-09-13 E08真实负载未通过：8窗口/2PTY/万条目录/16MiB真实传输已运行，p95 278.2ms、303/319长帧，传输成功但未覆盖全采样区间；恢复写入合并及9项恢复测试通过，仍需终端输出增量化与大文件性能优化，见[失败证据](../reports/performance-real-2026-09-13.md)。

2026-09-13 F10/E01增量：原生实例CPU/RSS聚合当前运行进程树，出生身份核对和成员变化重采样已接入；真实忙子进程、FIFO新增成员、TLS链路回归与构建通过。Windows真机、短命成员历史累计和E08全负载仍保留，见[进程树证据](../reports/process-tree-metrics-2026-09-13.md)。

2026-09-13 F10/E01增量：Linux原生主进程CPU真实双采样、出生身份与代次基线保护、CPU计算口径标识已接入并通过真实耗时测试和节点回归。fixture信号清理实际通过。整个运行进程树聚合、Windows真机和E08完整负载仍待补，见[证据](../reports/native-cpu-2026-09-13.md)。

2026-09-13 F10/E01/E08增量：网络累计量展示、原生未采集项不可用标识、后台降频/可见恢复已实现；真实Docker浏览器29.9s和构建通过。Linux原生CPU尚待采集，完整并发负载、fixture新清理流程和一次第二账号界面登录超时仍未解决或未验证，见[可见性证据](../reports/monitor-visibility-2026-09-13.md)。

2026-09-13 F10/E01缓存增量：指标采样后复核身份/权限，旧值仅匹配当前RunID。真实TLS断开/重建Daemon、断链撤权和新代次无缓存拒绝通过，监控race5.043s及构建通过；物理断网/采样中精确竞态与跨平台负载仍保留，见[指标证据](../reports/container-metrics-2026-09-13.md)。

2026-09-13 F10/E01端到端：实例监控标签已接真实指标，实例授权范围不请求主机信息；修复新节点首次容器启动目录初始化。实际双Daemon/Docker浏览器1/1验证读权限、撤权、CPU/内存、刷新与停止后旧值，runtime/daemon回归及构建通过。真实断链/跨代次缓存仍待验证，见[容器指标证据](../reports/container-metrics-2026-09-13.md)，整项仍进行中。

2026-09-13 F10/E01增量：容器实例指标从固定不可用改为实际Docker采样，身份/启动时间复核、CPU与内存及有界网络计数已实现；HTTP边界race和真实隔离容器指标通过。Master/Daemon与浏览器新增组合仍待验证，见[容器指标证据](../reports/container-metrics-2026-09-13.md)，整项保持进行中。

2026-09-13 F09/F11/A10/E02增量：Linux Compose CLI持久归属、遗留进程恢复及未确认时禁止新变更已实现；执行器SIGKILL、启动拒绝、容器/Daemon race及真实Docker生命周期通过。完整Daemon重连组合和Windows真机仍未验证，见[恢复证据](../reports/compose-cli-recovery-2026-09-13.md)，不计整项完成。

2026-09-12 F09/F11/A10/E02增量：容器执行器SIGKILL后输出检查点保留、任务不明且不重放，真实进程race通过；Docker浏览器输出和命令退出标识1/1通过。Linux遗留Compose CLI的Daemon启动恢复仍有代码缺口，完整A10未通过，见[输出证据](../reports/compose-output-2026-09-12.md)。

2026-09-12 F09/F11/E02增量：Compose运行中每250ms有变化才持久化有界输出，退出标识与任务资源结果分开；实际CLI未退出时跨管理器读取、取消及目录写入故障测试通过，容器race与构建通过。Daemon硬崩溃、Windows真机及完整组合仍未验证，见[输出证据](../reports/compose-output-2026-09-12.md)，整项保持进行中。

2026-09-12 扩展resource.write实例元数据分支已接SDK/API/CAS幂等和独立授权，transport与构建通过；真实独立示例浏览器、通知和其余资源能力仍待补，E06保持进行中。见 [元数据证据](../reports/extensions-metadata-2026-09-12.md)。

2026-09-12 计划任务分页已贯通Linux/Windows、RPC/API和界面；205项三页、参数TLS与浏览器恢复子项通过，Windows COM分页真机仍未验证，见 [系统报告](../reports/system-adapters-2026-09-12.md)。

2026-09-12 系统服务名称分页已贯通Linux/Windows适配、RPC/API与界面，游标刷新恢复；模块race、替身浏览器和构建通过。[系统报告](../reports/system-adapters-2026-09-12.md)保留计划任务分页、Windows真机与完整性能场景缺口。

2026-09-12 Windows服务操作已改原生SCM并确认STOPPED/RUNNING，30秒截止与取消检查；名称边界race和构建通过，真机状态组合仍未验证。[系统报告](../reports/system-adapters-2026-09-12.md)记录细项与分页缺口。

2026-09-12 Windows计划任务已换为固定脚本/COM结构化读写，支持完整Unicode路径并回读Enabled；路径/输出边界race与交叉编译通过，真机COM执行及完整分页未验证或未完成。见 [系统适配](../reports/system-adapters-2026-09-12.md)。

2026-09-12 系统适配：Windows服务SCM查询及只读真机入口、Linux.timer身份与操作边界已接入；Linux race及Windows交叉编译通过，Windows真机/计划任务结构化适配/完整分页仍保留。见 [系统报告](../reports/system-adapters-2026-09-12.md)。

2026-09-12 旧单份防火墙Apply/Rollback及reload路径已删除，错误/取消/恢复失败测试迁移到Snapshot并通过race；含Daemon命令子进程回归29.595s。[快照证据](../reports/firewall-snapshots-2026-09-12.md)已更新，真实危险环境和Windows仍保留。

2026-09-12 Daemon防火墙命令进程链路：独立子进程测试覆盖生产Snapshot路径的摘要拒绝、持久快照、取消恢复与终态，race28.946s通过；firewalld状态仍为测试替身，真实规则和失联环境不计通过。见 [快照证据](../reports/firewall-snapshots-2026-09-12.md)。

2026-09-12 双份预览补充：运行/永久差异与区域已显示，提交摘要绑定双快照和目标，执行前重新核对。定向race、浏览器和构建证据见 [快照报告](../reports/firewall-snapshots-2026-09-12.md)。真实命令/防火墙环境与旧兼容清理仍待补，E05不计整项完成。

2026-09-12 新防火墙租约已接入固定区域、运行/永久双快照与分别应用/恢复，新生产路径无全局reload；不同原配置、部分失败恢复和外部冲突的状态化命令替身race通过。[快照证据](../reports/firewall-snapshots-2026-09-12.md)说明旧租约、预览与真实环境缺口，E05仍进行中。

2026-09-12 防火墙预览/界面增量：实际受管端口差异、任务排队与确认阶段、原节点绑定及刷新恢复已有代码和定向race/浏览器证据；运行/永久规则混用仍待修复，E05保持进行中。见 [证据](../reports/firewall-preview-2026-09-12.md)。

2026-09-12 回滚补充：规则恢复完成持久记录、任务终态后清理、重启不重放已恢复规则以及应用失败保留证据已接入；SQLite任务更新故障注入与租约race通过。[防火墙证据](../reports/firewall-confirm-2026-09-12.md)已补充，真实规则合同与环境缺口仍保留。

2026-09-12 防火墙确认成功路径改为先写持久任务终态、再清理恢复记录；SQLite故障注入与租约定向race通过。真实防火墙环境、回滚终态故障组合仍未完成，E05保持进行中，见 [证据](../reports/firewall-confirm-2026-09-12.md)。

2026-09-12 监控增量：Windows 原生采集/进程及真机入口、Linux pidfd核对并确认退出、完整枚举搜索与PID分页、确认目标和视图恢复已接入；Linux race、Windows构建和浏览器子项通过，真机与跨处理器组CPU等缺口保留。见 [监控证据](../reports/monitoring-2026-09-12.md)。扩展迁移已移到锁外并按包/数据/令牌条件提交，授权不阻塞及并发写冲突定向race通过；E07仍不计整项通过。

版本：v1.0 · 2026-09-09
当前基线：完整实施已开始；部分核心与桌面子场景有证据，整项验收尚未完成。

2026-09-12 复核更正：B08/B09 尚有代码缺口，不能全部归因为环境缺失。主机任务分发和扩展任务查看已修复并通过真实 TLS 子项；独立 SDK 示例正在补实际浏览器包和界面，后台扩展任务仍仅回传输入，完整受限后台执行、资源桥接和升级状态迁移未完成。此前 fixture 演示包不证明独立 SDK 示例验收。见 [复核报告](../reports/resume-2026-09-12.md)。

同日增量：独立包浏览器界面与资源摘要桥接已实现，真实 TLS 独立包安装/资源授权和浏览器资源窗口/笔记独立分别验证通过；有界 WASI 后台执行器已通过真实 guest 测试但尚未接入生产任务，不能计作后台扩展全链路通过。完整资源写入/任务能力、状态迁移及原环境缺口继续保留。

后续 WASI 链路增量：组合包、包/模块摘要绑定、批量通道分块模块获取已接入生产任务，原输入回传代码已替换。独立包真实 TLS 计算与浏览器完整包上传分别有证据；首次全仓 race 暴露冷编译限时与测试等待问题，修复后正在复验。真实浏览器后台任务结果、完整迁移/升级故障和其余能力边界仍未完成，不能提升 E06/E07 整项状态。

上述复验已结束：独立包真实浏览器上传、节点窗口和 WASI 计算 1/1 通过（9.4s）；最终全仓 race 通过（Master 191.796s）。E06/E07 仍保留完整状态迁移、生命周期并发/失败恢复及其余能力缺口；详细命令与失败修复过程见同日复核报告。

后续生命周期增量：公共读写锁、一致包快照、损坏回滚前置校验、禁用状态保持和依赖门禁已通过定向 race；视图迁移在 opaque iframe 执行，宿主条件提交状态/版本，浏览器成功/失败保留场景已通过。尚缺包跨文件写入的掉电恢复、后端用户数据迁移、真实两版本部署及完整能力组合，E06/E07 继续进行中。见 [生命周期报告](../reports/extensions-lifecycle-2026-09-12.md)。

注册表恢复增量：跨文件撤销记录、持久提交令牌和启动恢复已实现，实际子进程在多阶段退出的恢复测试通过；Windows 写穿替换/删除及长路径代码交叉构建通过。物理断电、ENOSPC、Windows 真机耐久性、后端数据迁移及其余 E06/E07 组合仍未验证或未完成。见 [恢复报告](../reports/extensions-recovery-2026-09-12.md)。

## 1. 范围与使用方法

2026-09-12 用户数据增量：按用户隔离的 JSON 读写/CAS/持久回执、data.read/write SDK、WASI 数据版本迁移和包/数据共同事务恢复已实现。真实独立三版本浏览器迁移/失败保留/回滚 1/1（11.7s）、双用户 TLS、定向 race、实际子进程中断恢复及并发/配额保护通过。物理故障、Windows 真机、其余 SDK 能力及完整组合仍保留，E06/E07 不提升为整项通过。见 [用户数据报告](../reports/extensions-data-2026-09-12.md)。

桌面 SDK 增量：受控自身标签移窗/关闭、当前资源快捷入口和独立包 0.2.0 已实现，真实双节点浏览器组合通过（9.6s），前端 32/32。首轮替身浏览器入口缺失未复现，完整资源/通知能力与后端数据迁移仍继续。见 [桌面 SDK 报告](../reports/extensions-desktop-2026-09-12.md)。

任务审计增量：真实 Daemon 终态回读及派发前取消补充事务审计，故障注入/去重与真实 TLS 验证通过，F03/F09 整项仍进行中。见 [审计报告](../reports/task-audit-2026-09-12.md)。

清单签名增量：发布者认证现覆盖清单全部受支持字段与包体摘要，旧的包体单独签名拒绝；签名工具、重启后验证、存储清单/能力篡改拒绝及真实 TLS 回归通过。E07 仍待后端数据迁移和完整组合，不提升整项状态。见 [签名报告](../reports/extensions-signature-2026-09-12.md)。

2026-09-12 扩展任务增量：SDK 查询/取消及刷新恢复已实现；真实双节点 TLS 授权/取消 race 和真实独立包浏览器 SDK 结果回读通过。禁用历史查询、撤权拒绝和刷新不重提交有子项证据；清单签名绑定、后端用户数据迁移和其余能力仍待补，整项保持进行中。见 [任务报告](../reports/extensions-tasks-2026-09-12.md)。

本矩阵补充而不替代[技术架构](../../plan/Blora-01-技术架构.md)和[功能与交互](../../plan/Blora-02-功能与交互.md)的正文。表内描述是追踪入口，不是删减详细要求的依据。

F01～F14 追踪功能交付；A01～A17 沿用原规划的关键场景；E01～E09 补足运维、扩展和最终交付验证。实施块见 [ACTION_GUIDE.md](../../../ACTION_GUIDE.md)。

### 1.1 本次必须完成

| 范围 | 解释 |
| --- | --- |
| M0～M3 | 同一开发任务内的依赖顺序；不是只做 MVP 的分期授权 |
| 默认应用 | 桌面宿主统一注册，主要功能以应用提供；实例中心、文件、编辑器、终端、任务、监控、节点、用户、设置和运维应用均有真实入口 |
| 多窗口与多标签 | 主要应用可多开；每个窗口可管理多个节点资源；标签可移窗；同资源可以有多个视图 |
| 实例快捷方式 | 独立资源入口；默认打开或激活专用窗口；可显式再开窗口 |
| Linux 与 Windows | 实际平台运行适配、文件、PTY 和停止控制；平台差异明确，不用空实现冒充支持 |
| 运维增强 | Docker/Compose、完整主机监控与进程视图、跨节点移动、备份恢复、调度、有限系统管理 |
| 扩展生态 | 公共 SDK、沙箱与受控能力、注册源/目录、可安装包、权限、升级迁移、禁用卸载和真实示例 |
| 刷新恢复 | 普通 F5/工具栏刷新保留完整已登记工作现场；最近输入和撤销重做属于核心能力 |

### 1.2 保留原规划边界

| 事项 | 本次处理方式 |
| --- | --- |
| 被管理服务的业务端口、穿透、VPN、游戏协议网关 | 用户独立配置；可以成为未来扩展应用，核心不实现业务中继 |
| 1Panel 的全部产品功能 | 只实现两份规划明确列出的管理能力，不无限扩张为网站托管、邮件等完整产品集合 |
| 公网商业应用市场运营 | 不要求运营公网服务、支付、审核团队；要求可工作的目录/注册源和包分发生命周期，本地测试源可验收 |
| 多 Master、高可用、外部数据库及额外部署形态 | 原规划中按需求增加的事项，当前保留合理模块边界，不强行纳入 |
| Windows 上所有发行版/Engine/系统功能 | 按平台能力声明和独立适配，不能把 Linux 命令视为 Windows 支持；缺少某 Engine 不应阻止其他功能 |
| 跨设备工作现场 | 恢复已同步副本；默认同步布局和引用，工作内容同步由用户开启；不保证未上传状态凭空出现 |
| 操作系统文件选择授权、密码和私钥字段 | 不伪造重新授权、不保存敏感字段；提供真实续传或重新选择流程 |
| 浏览器数据被清除、存储拒绝、磁盘或配额耗尽 | 明示保护失败并保留可导出数据，不声称任何条件下绝对不丢失 |
| 任意 TUI 无检查点重建、浏览器完整离线启动、操作系统级全局快捷键 | 不作不受浏览器与会话条件限制的承诺；正常会话恢复和桌面内快捷操作仍必须实现 |
| 视觉细节后续优化 | 可调颜色、间距、动画参数和文案；窗口模型、拖放、焦点、即时反馈不能留作细节待办 |

## 2. 状态和证据规则

实现状态：未开始 / 进行中 / 已实现 / 阻塞。
验证状态：未运行 / 通过 / 失败 / 环境缺失 / 不适用（必须写明平台能力依据）。

“已实现”不等于“通过”。只交叉编译 Windows 不等于验证 Windows 运行；只有模拟后端的浏览器测试不等于真实链路通过。环境缺失不计入通过率，不得为了全绿改成不适用。

每条证据至少记录：编号、覆盖子场景、代码位置、可重复命令、测试环境和时间、实际结果/退出码、日志或截图路径、尚未覆盖项。可在本文件追加证据索引，详细报告存放在 docs/acceptance/reports/，但必须是真实生成的报告。

更新表格的实现/验证列并附证据链接。若只验证部分条件，写“进行中”或明确子项结果，不把整行标为通过。功能入口、后端权限、失败行为和持久化缺一时不能完成该 F 项。

## 3. 功能交付 F01～F14

2026-09-12 F09/F11/E02 Compose输出：成功/失败阶段输出有界持久化、截断与单阶段受控查询、任务详情浏览/刷新接入。实际CLI进程race、真实Docker/Compose浏览器1/1（45.9s）、Master权限测试及构建通过，见 [输出证据](../reports/compose-output-2026-09-12.md)。运行中流式输出、崩溃窗口及完整平台/故障组合继续。

2026-09-12 F09/F11拉取边界：保存EOF前末条观测，超32MiB流不再因限流EOF进入成功检查，定向真实HTTP协议替身race3.743s及平台构建通过，见 [进度证据](../reports/container-progress-2026-09-12.md)。真实下载、Compose输出持久化和完整组合继续。

2026-09-12 F09/F11镜像进度增量：保留Engine分层身份与Current/Total并写入节点日志/任务修订，HTTP协议替身定向race与构建通过，见 [分层进度证据](../reports/container-progress-2026-09-12.md)。真实拉取及完整输出日志未验，不提升整项。

2026-09-12 F09阶段详情：发起者/时间/耗时、专用任务详情、持久状态阶段分页与当前授权接入。真实TLS/race、前端41/41、真实任务浏览器1/1（3.8s）通过，见 [阶段证据](../reports/task-stages-2026-09-12.md)。各执行器完整输出/进度与长期故障组合仍保留，整项不提升。

2026-09-12 F09统计/筛选增量：全部获准未终态根任务统计与数据库状态筛选、刷新恢复已接入；超过1000条新历史遮蔽旧等待任务的权限TLS/race、真实浏览器组合2/2（9.8s）、前端39/39通过，见 [任务历史证据](../reports/task-history-2026-09-12.md)。完整任务元信息/阶段日志、系统通知真桌面和故障组合继续。

2026-09-12 F09历史增量：后端稳定游标读取突破最近1000条，服务端当前权限过滤、桌面翻页/刷新及专用详情直接查询接入；1005项race、真实TLS及101个真实任务浏览器组合2/2通过，见 [任务历史证据](../reports/task-history-2026-09-12.md)。后台数量聚合、全历史条件筛选和完整故障组合继续。

2026-09-12 F09终态通知续接：持久事件WSS游标、首次边界与重连、授权过滤、默认桌面通知及精确任务跳转、显式系统通知启用/关闭接入。最终浏览器组合3/3（9.8s）、WSS race2.321s、恢复日志故障保护通过，见 [通知证据](../reports/notifications-2026-09-12.md)。操作系统原生通知展示/点击未验，整项继续进行中。

2026-09-12 F09/F14/E06通知增量：受控扩展通知API/SDK、持久站内托盘、来源绑定与点击跳转、立即刷新不重发接入。独立包浏览器1/1、TLS通知/云端选择策略race、前端37/37通过，见 [通知证据](../reports/notifications-2026-09-12.md)。后台任务自动完成通知与系统通知发送继续，整项不提升。

2026-09-12 F14/E06元数据增量：独立包实例编辑、立即刷新草稿恢复、真实后端名称保存与刷新不重放，浏览器1/1（4.7s）通过；独立包资源桥接race30.708s通过。见 [元数据证据](../reports/extensions-metadata-2026-09-12.md)。其余SDK能力及完整组合仍保留，整项状态不提升。

| 编号 | 必须交付的用户能力和边界 | 主要验收 | 实现 | 验证/证据 |
| --- | --- | --- | --- | --- |
| F01 | 桌面、任务栏、窗口移动缩放吸附/层级/焦点；应用多开、跨节点资源标签、标签移窗、快捷方式、菜单和键盘操作；可复制文本区域保留选择能力 | A01、A13、A15～A17 | 进行中 | 完整场景未通过；本轮专用实例窗口、真实跨节点双中心、账号菜单退出/现场隔离定向场景通过；跨窗口测试修正吸附区重叠后复验通过，见[rc13报告](../reports/rc13-current-source-2026-09-20.md)；[当前核心证据](../reports/core-2026-09-09.md)、[运行适配子项](../reports/runtime-2026-09-09.md) |
| F02 | 布局、窗口/标签归属、正文、撤销重做、表单与视图现场、终端检查点；本地日志与快照、云端副本和版本迁移；账号/设备/浏览器标签隔离 | A01、A02、A06、A08、A11、A17 | 进行中 | 云端布局/引用默认同步、正文显式同步和打开恢复已有真实 1/1；同步 PUT 在中止后刷新重试复用同一请求键并只成功写入一次，真实 Chromium `cloud workspace sync` 1/1（8.0s）通过；配额失败、schema迁移和事务中止实际浏览器已通过（A11）；2026-09-19真实双设备 stale revision 与终端检查点内容策略 2/2（6.9s）通过且本地正文保留；2026-09-20并发快照超256KiB尾日志防护、终端输出增量日志及旧检查点兼容单测通过；RC17补齐主题/字号/密度偏好与默认布局备份恢复，设置浏览器3/3通过，见[RC17发行报告](../reports/release-rc17-2026-09-22.md)、[多设备冲突证据](../reports/workspaces-2026-09-19.md)与[性能恢复报告](../reports/performance-continuous-2026-09-19.md)；断电和存储耗尽仍待补，见[工作区报告](../reports/workspaces-2026-09-10.md)、[请求键报告](../reports/request-keys-2026-09-14.md) |
| F03 | 登录退出、多用户、角色与资源授权；查看/创建/控制/文件/终端输入/主机管理分权；服务端列表、统计、请求及流均授权 | A08、A14、E04、E07 | 进行中 | 新增管理员审计检索（用户/节点/资源/动作/时间/分页）及普通用户 403 集成证据；用户创建/改密界面保存待处理请求键；本轮 `/files/access` 返回当前用户文件读写能力，真实 TLS 集成确认管理员可写、普通用户只读和撤权后拒绝，浏览器只读界面 1/1 通过；整项场景仍未通过，见[编辑器能力增量](../reports/editor-capabilities-2026-09-22.md)、[审计与撤销证据](../reports/audit-2026-09-10.md)、[请求键报告](../reports/request-keys-2026-09-14.md)及[rc13报告](../reports/rc13-current-source-2026-09-20.md) |
| F04 | 节点登记、身份轮换吊销、管理连接、心跳、版本/能力协商、在线/异常/离线/维护状态、入口；只管理通信 | A05、A08～A10 | 进行中 | 节点管理新增持久撤销、连接代次递增、幂等重试、分组/标签设置和界面入口；登记/换钥/撤销界面待处理请求键随窗口恢复，节点 TLS 管理集成通过；完整失联/平台组合仍未通过，见[审计与撤销证据](../reports/audit-2026-09-10.md)、[请求键报告](../reports/request-keys-2026-09-14.md) |
| F05 | 聚合全部获准实例、获准节点创建向导、搜索筛选统计/批量操作；多窗口多标签和专用资源窗口；真实启动停止重启与阶段诊断 | A03～A05、A10、A14～A17 | 进行中 | 完整场景未通过；实例中心现补齐可恢复的列表/卡片视图，筛选、选择和批量请求身份不受切换影响，`instance-batch.spec.ts` 1/1通过，见[实例视图报告](../reports/instances-view-2026-09-22.md)；本轮真实跨节点双中心/专用实例窗口和节点范围创建复验通过；测试按夹具实例身份核验，不依赖临时测试实例的总数，见[rc13报告](../reports/rc13-current-source-2026-09-20.md)；[当前核心证据](../reports/core-2026-09-09.md)、[运行适配子项](../reports/runtime-2026-09-09.md) |
| F06 | 实例控制台、PTY、只读观察/受权输入、多窗口标签/分栏、会话重新挂载、输出流控与检查点；主机 Shell 独立高权限 | A06、A08、A09、A17 | 进行中 | 完整场景未通过；2026-09-20真实WSS终端刷新/移窗后保持原会话、无输入重放1/1（17.6s）；增量输出检查点兼容旧格式、ACK仍等待本地保护，见[性能恢复报告](../reports/performance-continuous-2026-09-19.md)；[当前核心证据](../reports/core-2026-09-09.md)、[运行适配子项](../reports/runtime-2026-09-09.md) |
| F07 | 多选、重命名、新建、复制移动、回收/删除、压缩解压、上传下载、跨节点复制与移动；真实根目录隔离、冲突和传输续接 | A07、A08、A09、A12 | 进行中 | 实例中心新增管理员软删除入口并沿用停止/活动任务/幂等约束；跨节点持久协调、权限撤销和源对象保护通过；浏览器上传分片按偏移使用稳定请求键，Master 在分片/完成前拒绝缺失键；重编译夹具后的512MiB分块校验与批量步进传输真实完成（110.868s、目标校验成功），并有[并行关闭网页续传](../reports/upload-close-2026-09-13.md)证据。文件管理器本轮补齐按需目录树、列表空白区框选、可恢复的图标/列表视图和目录快捷方式，列表虚拟滚动/图标分页、根目录→子目录→孙目录刷新恢复、上传续传/取消均在文件浏览器套件5/5通过，见[目录树报告](../reports/files-tree-2026-09-22.md)与[视图报告](../reports/files-view-2026-09-22.md)。Linux目标tmpfs真实ENOSPC已通过（A12）；本轮 Linux `internal/filesystem` race 套件1.332s通过，包含符号链接/FIFO/目录交换隔离；Windows amd64 测试包交叉构建通过但未运行。物理掉电及其余故障组合仍单列待补 |
| F08 | 真实编辑器、语法能力、共享正文与独立视图、草稿/撤销重做恢复、显式保存、原子写入/版本冲突；大文件二进制有明确边界 | A01、A02、A08、A11、A17 | 进行中 | 完整场景未通过；UTF-8/BOM 与 LF/CRLF 元数据往返、4 MiB 上限显示、扩展名语言映射、共享正文的独立分栏视图、多光标编辑刷新后撤销重做，以及只读用户禁止编辑/保存已补；撤销历史按最多 8 MiB/文件上限两倍执行，过大活动编辑组整体落为基线以避免部分撤销；真实节点 API 验证 4 MiB 可读、超限文本/二进制拒绝且原字节可下载；新增受版本绑定的有限 UTF-8 只读分段预览，真实文件 API覆盖首段/第二段/二进制拒绝，浏览器边界套件3/3通过。完整跨应用/故障交叉场景仍未通过，详见[编辑器能力增量](../reports/editor-capabilities-2026-09-22.md)；[当前核心证据](../reports/core-2026-09-09.md)、[运行适配子项](../reports/runtime-2026-09-09.md) |
| F09 | 持久化后台任务、阶段/进度/日志、等待、取消、重试和结果；站内通知与系统通知授权；关闭网页不取消服务端任务 | A04、A05、A07、A09、A10 | 进行中 | 任务中心显示持久结果、有界分页及失败/中断生命周期任务关联重试；真实TLS覆盖幂等、失败和授权边界。2026-09-29独立X11/DBus/Dunst中真实节点任务触发原生弹窗，系统鼠标点击正确taskId，标题/正文、刷新去重及浏览器退出后用户关闭均通过，见[原生通知报告](../reports/native-notifications-2026-09-29.md)。不承诺退出后后台投递或重新启动；其他桌面平台与完整故障组合未由Linux结果覆盖。历史证据见[通知API](../reports/system-notifications-2026-09-20.md)、[重试](../reports/tasks-retry-2026-09-10.md)、[核心](../reports/core-2026-09-09.md) |
| F10 | 节点/实例指标、历史采样策略与旧数据标识、完整系统进程视图、诊断和授权终止；窗口最小化降低绘制开销 | E01、E08 | 进行中 | [监控报告](../reports/monitoring-2026-09-09.md)；节点/实例实时指标、历史、进程搜索排序和出生标识终止子项通过，2026-09-19真实Daemon失联/重连、旧指标时间/诊断和旧进程禁用入口1/1（53.4s）通过，见[失联监控报告](../reports/monitor-offline-2026-09-19.md)；本轮 Linux 真实5秒间隔采样121点、Master缓存离线回读与Master重启后120个stale点恢复均通过，详见[长时监控报告](../reports/monitor-history-soak-2026-09-21.md)；Windows采样仍待补 |
| F11 | 容器、镜像、卷、网络和 Compose 项目；日志/终端、配置草稿、显式应用和分阶段结果；数据删除与容器删除区分 | E02 | 进行中 | 当前源码在真实 Docker Engine 29.4.1/overlayfs 上的 Compose 生命周期 race 27.681s 通过，覆盖容器/镜像/卷/网络、双流日志、更新/删除、健康失败、卷保留和普通用户拒权；本轮容器指标及 Compose 浏览器场景2/2通过；宿主 PTY 与 Docker 浏览器边界也已通过；已有后端有界日志归档现已接入前端实时/归档切换，原有真实 Compose 长流程1/1（45.6s）复验通过，Docker 浏览器链路2/2、真实隔离 Docker Engine 历史 API 2.79s及真实 Chromium 日志 UI 1/1（7.6s）通过，真实链路覆盖归档持久化、未授权403、`limit=101` 400、实时/归档正文和缺口提示，详见[Docker日志归档报告](../reports/docker-log-history-2026-09-22.md)；2026-09-21容器内1MiB tmpfs实际触发内核ENOSPC，Compose错误状态/exit code/日志及部分资源事实保持通过，见[ENOSPC报告](../reports/docker-enospc-2026-09-21.md)；当前源码另经 loopback HTTPS TLS 代理对真实 Engine/Compose 完成同一生命周期 E2E 30.49s，通过已验证 HTTPS endpoint，但不是远端主机证据，详见[HTTPS传输报告](../reports/docker-https-transport-2026-09-21.md)；远程 Engine、Engine主机存储耗尽/掉电及Windows容器仍待补，见[当前 Docker 验收](../reports/docker-current-2026-09-20.md) |
| F12 | 可恢复备份、目标/范围、保留策略与一致性钩子；平台调度、时区、错过执行/重叠策略、触发重授权和任务结果 | E03、E04 | 进行中 | [备份与调度模块](../reports/backup-scheduler-library.md)、[备份界面增量](../reports/backup-ui-2026-09-10.md)、[真实备份浏览器](../reports/backup-browser-2026-09-14.md)、[定时一致性验证](../reports/scheduled-consistency-2026-09-20.md)；默认桌面已接入真实备份/恢复计划/定时任务入口，清理动作也保存稳定请求键，产品 HookRunner 已提供固定 argv 配置入口，实际固定argv保存程序成功/失败补偿、定时保存钩子、真实ENOSPC及节点离线/服务重开调度已通过；当前源码已验证备份 ZIP 写入中断后的归档清理/新任务恢复（[归档崩溃报告](../reports/backup-archive-process-crash-2026-09-21.md)），以及恢复上传中途强制终止后的暂存清理、原ID中断对账和新计划成功恢复（[恢复崩溃报告](../reports/backup-restore-process-crash-2026-09-21.md)）；scheduler独立进程及真实发行包 Master+所属Daemon 在已接受 occurrence 后强制终止、原 task/request 回执恢复且归档唯一，分别见[调度器进程崩溃报告](../reports/scheduler-process-crash-2026-09-21.md)和[Master/Daemon 崩溃报告](../reports/scheduled-backup-master-daemon-crash-2026-09-21.md)；真实 `Scheduler.Tick` 的 pending/prepared 两个已提交阶段中断后均会以同一 occurrence 恢复且无重复，见[调度阶段崩溃报告](../reports/scheduler-staged-process-crash-2026-09-21.md)。设备写入中断、长期墙钟和物理掉电仍待补 |
| F13 | 按平台能力提供有限系统服务、防火墙和主机计划任务管理；管理员专用、变更差异、管理连接保护及回滚 | E05 | 进行中 | 能力探测、服务/计划任务任务链路、防火墙预览及 Linux firewalld 端口 apply 子集已实现；两阶段 60 秒租约、确认、取消/过期回滚和 Daemon 重启恢复测试通过；真实私有网络firewalld失联回滚已通过，见[私有环境报告](../reports/firewall-private-2026-09-19.md)；不操作生产宿主，系统服务/计划任务与Windows环境仍待补 |
| F14 | 默认与扩展应用共同宿主；SDK、状态合同、受控能力、沙箱、注册源和真实安装包；安装授权、版本兼容、升级、禁用、卸载 | A13、E06、E07 | 进行中 | [SDK报告](../reports/sdk-2026-09-09.md)；独立 `sdk/`、参考包、签名校验、安装/升级/回退/禁用/卸载、动态 bundle sandbox 加载和受控任务桥接已有 API/浏览器及真实 15/15 fixture 证据；本轮扩展真实浏览器7个独立场景最终均通过，涵盖通知、任务取消、草稿/迁移、WASI、目录两版本升级和权限隔离；`data-app-id` 仅供稳定定位，不改可见 UI；扩展管理界面现按扩展/版本保存生命周期请求键；HTTPS 远程目录源已有 TLS 单元/集成覆盖，资源标签/跨设备状态迁移已有A13证据；真实远程部署、高延迟证书轮换仍待补，见[rc13报告](../reports/rc13-current-source-2026-09-20.md) |

## 4. 原规划关键场景 A01～A17

以下是可执行验证要求，不预设任何测试已经通过。立即刷新测试不得人为等待自动保存完成。

| 编号 | 场景与通过标准 | 实现 | 验证/证据 |
| --- | --- | --- | --- |
| A01 | 输入、中文输入提交、粘贴、连续编辑和拖窗后立即 F5/工具栏刷新；正文、窗口布局和应用状态一致，撤销/重做可沿刷新前历史继续；分别验证活跃与后台窗口 | 已实现 | Linux Chromium真实IME提交/剪贴板粘贴/连续编辑/拖窗/活跃与后台窗口reload两场景2/2（6.0s），末次输入无等待刷新、精确布局/焦点及undo redo通过，见[立即刷新报告](../reports/immediate-refresh-2026-09-13.md) |
| A02 | 同用户两窗口打开同文件，随后外部用户修改服务器版本；共享正文策略正确、局部光标/滚动独立，保存时检测版本冲突并保留各版本，不用旧响应覆盖新输入 | 已实现 | Linux Chromium真实双账号文件冲突及迟到响应1/1（9.6s），共享模型独立光标/滚动/刷新后准确插入1/1（12.7s）；旧版本不覆盖新输入、两版本保留与显式解决通过，见[共享编辑报告](../reports/shared-editor-conflict-2026-09-13.md) |
| A03 | 真实实例忽略正常停止、派生后代、主 PID 提前退出或后代持有管道；有限等待后升级或明确失败，全局仍可用，归属运行未退出不启动新代次 | 已实现 | 2026-09-29 本机私有委派cgroup出生归属、整组kill确认与CPU实际限流通过，见[平台报告](../reports/systemd-cgroup-2026-09-29.md)。Linux真实双节点停止失败/拒绝新运行/另一节点可控race4.375s，真实主PID提前退出及管道后代/强制升级/截止失败三项race1.302s通过，见[停止边界报告](../reports/stop-boundary-2026-09-13.md)。Windows/cgroup平台运行限制见E09 |
| A04 | 重复重启请求、服务端已接受但客户端丢响应、Master 中途重启；同 requestId 对账复用结果，同资源不同请求按策略串行，不重复产生运行 | 已实现 | 同key重启/管理断线race通过，服务与SQLite重开通过，独立Master SIGKILL后原任务及精确启动计数通过；真实浏览器202响应丢失后刷新重试1/1（4.4s），见[重启报告](../reports/restart-disconnect-2026-09-13.md)。不同请求资源串行见A15，Windows运行见E09 |
| A05 | 重启和传输途中断开节点连接并恢复；标为失联/待确认，不伪造停止或成功；按实际执行记录校准，不因重连重复副作用 | 已实现 | 重启RUNNING管理断线原请求复用、传输中途bulk断线目标验证/提交唯一race12.490s、真实浏览器OFFLINE与WAITING_NODE原任务恢复1/1（50.7s），见[失联与恢复证据](../reports/node-offline-browser-2026-09-13.md)及[重启断线](../reports/restart-disconnect-2026-09-13.md)。其他平台见E09 |
| A06 | 在实际 PTY 中运行 vim/top 等全屏程序、中文/ANSI/光标模式后刷新；检查点和输出序号衔接正确，继续使用原会话，回放绝不重发历史输入 | 已实现 | Linux真实vim/top及WSS ANSI/备用屏幕/精确恢复sequence/原sessionId/输入不重放组合2/2（26.7s），真实文件副作用单次，见[终端程序报告](../reports/terminal-programs-2026-09-13.md)。Windows运行见E09 |
| A07 | 同时发起节点端解压和来自浏览器本地文件的上传，再关闭网页；解压独立完成，上传保留已确认检查点并在需要时等待重新选择源文件，状态与真实结果一致 | 已实现 | Linux真实场景通过：10,000文件解压独立完成，16MiB上传在新browserTabId从983,040B复用原任务续传并核验摘要，见[完整场景证据](../reports/upload-close-2026-09-13.md)。Windows平台由E09另行验证 |
| A08 | 使用过程中撤销查看、写入或终端能力；新请求拒绝、现有流及时终止/降权；切换账号不能访问前账号草稿；旧快照不能恢复授权 | 已实现 | Linux真实浏览器PTY撤权1/1（6.1s）、文件查看/写入撤权与旧草稿恢复1/1（6.7s）、账号退出现场隔离1/1（3.5s）通过；新请求403、输入租约失效、服务器正文不变、旧现场不能恢复授权，见[撤权报告](../reports/browser-revoke-2026-09-13.md) |
| A09 | 大日志、慢接收端与批量文件传输并行；控制请求持续取得服务，队列、内存、归档磁盘有可测上限，背压/丢弃策略可见，无无限累积 | 进行中 | 真实 TLS Master/双 Daemon 并行复验通过（13.07s）：8.4MB 传输、慢日志未消费峰值82,635B、接收队列79,507B，停止≤5s且归档/目标校验成功；2026-09-21 RC15当前源码同一 race 组合复验13.918s通过；本轮当前源码真实TLS慢读者30秒逐秒采样：未消费峰值33,048B、接收队列峰值28,320B，随后文件传输、停止、归档重读和目标校验通过，见[30秒慢消费者报告](../reports/slow-consumer-soak-2026-09-21.md)；真实PTY约17.5s、64MiB无消费者滚动采样归档峰值65,484B，PTY进程RSS基线741,376B/峰值3,764,224B（基线+32MiB内）；2026-09-27本机独立24h真实采样与归档峰值在声明预算内，详见[最终报告](../reports/local-endurance-24h-2026-09-27.md)；跨平台、全连接更长期混合负载仍待补，见[混合流报告](../reports/mixed-stream-control-2026-09-13.md) |
| A10 | Daemon 崩溃后重启，包含已启动进程与中途任务；按运行身份而非仅 PID 对账，既有进程/任务结局明确，不可恢复的操作标为中断 | 已实现 | 独立Daemon运行中及STOPPING中SIGKILL后恢复原运行/任务明确INTERRUPTED不重放；真实PID错误出生标识拒认拒信号及运行标记恢复4项race1.058s，见[Daemon崩溃报告](../reports/daemon-crash-2026-09-13.md)。Windows真机见E09 |
| A11 | 注入浏览器配额写入失败并迁移旧恢复记录；不清空草稿、不显示虚假保护成功；快照完整切换，迁移失败可保留/导出旧记录 | 已实现 | Linux Chromium实际桌面故障注入3/3通过：配额失败导出最新正文、schema0迁移、schema999原始记录保留、指针事务中止前后数据库一致及同步日志刷新恢复；修复未捕获AbortError，见[恢复故障报告](../reports/recovery-quota-2026-09-13.md)。2026-09-26 Firefox 155.0和WebKit 26.6对应真实浏览器各3/3、退出0，见[跨引擎恢复报告](../reports/recovery-engines-2026-09-26.md)。实际配额阈值、物理掉电及浏览器清除站点数据仍不由故障注入结果覆盖 |
| A12 | 跨节点复制/移动遇到目标磁盘满、断线、来源变化；记录可诊断阶段和清理对象，未验证目标前不删除源；源改变后不错误删除新内容 | 已实现 | 传输断线、来源改变及删除前再次核对已有真实TLS证据；2026-09-19私有1MiB目标tmpfs真实ENOSPC竞态验收4.371s通过：FAILED/诊断明确，源完整保留、目标不发布、既有文件保留及暂存清理，见[空间不足报告](../reports/enospc-2026-09-19.md)和[传输报告](../reports/transfers.md)。Windows/远程文件系统及物理掉电另行未验 |
| A13 | 默认应用和参考扩展通过同一 AppHost 能力合同注册；无需改桌面核心可开窗口/标签、注册任务并恢复现场；扩展不能获得未授予能力 | 进行中 | 独立包真实浏览器已覆盖通知、元数据、迁移、WASI 任务、资源窗口/快捷入口/移窗/关闭及双版本目录安装；跨设备副本状态迁移与资源身份校准1/1；新增真实成员浏览器上下文未授权组合1/1：无 `app.use`、仅 `app.use` 无 `node.read` 均403，补齐节点读取后资源摘要200，撤销后立即403；本轮七个扩展场景在真实 HTTPS 双节点 fixture 上最终全通过；Windows/远程长期组合仍待补，见[扩展全链路报告](../reports/extensions-real-2026-09-13.md)、[跨设备报告](../reports/extensions-cross-device-2026-09-14.md)、[未授权浏览器报告](../reports/extensions-unauthorized-browser-2026-09-14.md)及[rc13报告](../reports/rc13-current-source-2026-09-20.md) |
| A14 | 两节点多实例、查看与创建权限不同；实例中心聚合全部可见资源，筛选/统计不泄漏；看见节点不等于可创建，提交再次校验权限/配额 | 已实现 | Linux Chromium真实双节点双账号：member 查看两实例但创建/第二实例控制受限；仅节点2临时授予创建与所需host.manage后向导仅列节点2，确认前撤权返回403且无实例，恢复授权后创建STOPPED实例并按ID清理，1/1（3.5s）；见[实例权限报告](../reports/instance-permissions-2026-09-13.md)。Windows/远程环境见E09 |
| A15 | 两个实例中心窗口包含多个跨节点实例标签，重复查看同一实例；真实运行状态同步，视图导航独立；两视图并发控制仍由后端统一去重/串行 | 已实现 | 双节点双实例中心真实浏览器1/1（7.9s），并发启动仅一成功、另一拒绝且runId一致，两窗口状态同步、独立导航刷新保留，见[多视图报告](../reports/instance-multiview-2026-09-13.md)；请求去重与故障归属另见[运行适配](../reports/runtime-2026-09-09.md) |
| A16 | 添加实例到桌面，单击启动入口、显式新开、改入口名称、删除入口、目标改名/删除/离线/权限撤销；默认打开专用管理窗口且可复用已有专用窗口，不创建/启动/删除实例，不用同名资源替代 | 已实现 | Linux Chromium真实双节点：专用窗口复用/显式新开/入口改名删除1/1（3.6s）、管理员从实例中心点击删除确认后目标改名/软删除/同名替换固定旧resourceRef1/1（7.0s）、成员撤权入口失效1/1（3.3s）、Daemon离线入口状态与恢复1/1（50.4s）；软删除仅管理员且停止/无活动任务，见[入口报告](../reports/shortcut-lifecycle-2026-09-13.md)。Windows真机见E09 |
| A17 | 把含未保存草稿或活跃终端的标签拖出新窗口，再移入兼容窗口并立即刷新；原视图 ID、归属、顺序、正文、撤销历史与会话引用保留，不重复视图/任务/输入 | 已实现 | Linux Chromium真实编辑器拖拽1/1（3.8s）、实际PTY拖拽1/1（8.6s）通过；逐次立即刷新核对身份/归属/顺序/撤销重做/会话和输入帧，真实文件副作用仅一次，见[拖拽报告](../reports/drag-recovery-2026-09-13.md)。其他平台见E09 |

## 5. 补充验收 E01～E09

2026-09-22 E08 非 Chromium 双 PTY 复核：强化夹具后，官方 Firefox 155.0 在真实8窗口/双PTY/10,000项目录/16MiB循环传输的60秒组合中两条PTY均约116KB/s、传输和队列正常，但1,020个指针样本 p95 **144ms**，超过≤50ms；此前官方WebKit同组合为78ms，Chromium为42ms。见[Firefox报告](../reports/firefox-performance-dual-pty-2026-09-22.md)与[WebKit/Chromium复核](../reports/webkit-performance-dual-pty-short-2026-09-22.md)。E08保持进行中。

| 编号 | 必须通过的完整场景 | 实现 | 验证/证据 |
| --- | --- | --- | --- |
| E01 | 主机与实例指标来自真实节点；系统进程搜索/排序/详情及授权操作可用，终止只命中指定测试进程；失联标明采样时间，不把旧值当实时；采样和历史保留有边界 | 进行中 | 实时/历史/身份终止子项通过；节点与实例指标均有 120 点有界持久缓存，节点失联回退标记 `stale`；浏览器已验证进程搜索、排序和 `startTicks` 提交 | [监控报告](../reports/monitoring-2026-09-09.md)；真实节点SIGSTOP/CONT的采样时间保持、旧数据诊断、缓存进程旧标记/禁用终止和重连更新1/1（2026-09-20，54.0s）通过；2026-09-21 Linux隔离HTTPS真实Master/Daemon按5秒间隔采样121次，在线历史精确保留120点（10.1m, 1/1）；1秒间隔完整组合中离线 stale、同库 Master 重启后的120点持久回读和 Daemon 恢复在 Chromium 1/1（2.7m）、Firefox 1/1（2.9m）和官方隔离容器 WebKit 1/1（2026-09-22）均通过，见[长时监控报告](../reports/monitor-history-soak-2026-09-21.md)与[WebKit报告](../reports/webkit-monitor-history-2026-09-22.md)；2026-09-22 受控 Docker namespace 的 `tc/netem` 80ms单向延迟、10ms抖动、1%丢包下，WebKit 与官方Playwright Chromium 同一完整组合均1/1通过（各3.6m），见[netem报告](../reports/netem-monitor-history-2026-09-22.md)；节点/实例缓存单测另以125个UTC时间戳跨午夜验证裁剪后仍严格保序（定向race 2.718s，见[测试源码](../../../internal/master/monitor_integration_test.go)）。2026-09-27本机独立24h真实采样、两次失联/重启与120点stale历史保持已通过，详见[最终报告](../reports/local-endurance-24h-2026-09-27.md)；仍不覆盖跨主机远端网络或Windows运行 |
| E02 | 在真实 Docker/Compose 测试环境创建项目、应用配置、查看日志/终端、更新和删除；镜像/卷/网络操作可追踪；镜像拉取失败/依赖缺失/部分成功明确；删容器默认不误删卷数据，配置草稿刷新保留 | 进行中 | 当前源码在 Docker Engine 29.4.1/overlayfs 上的 `TestRealDockerComposeLifecycle` race 27.681s 退出0，覆盖真实容器、镜像、卷、网络、双流日志、Compose创建/更新/删除、健康失败、卷保留和普通用户拒权；真实 Master/Daemon 容器测试另验证归档窗口持久化后的历史 API、403 权限边界和 limit 边界（2.79s）；真实 Chromium 日志 UI 1/1（7.6s）验证启动容器、实时/归档正文、缺口提示和清理；2026-09-21扩展后当前源码真实E2E race 29.17s通过，其中1MiB容器tmpfs实际触发内核ENOSPC，应用保留FAILED/health、exit code 1、容器事实及日志，见[ENOSPC报告](../reports/docker-enospc-2026-09-21.md)；另以私有证书、loopback HTTPS 代理连接实际 Docker Engine，Engine 客户端和 Compose TLS 生命周期 race E2E 30.49s通过，详见[HTTPS传输报告](../reports/docker-https-transport-2026-09-21.md)。这不是远端主机验证；远程 Engine、Engine主机存储耗尽/掉电、Windows容器和长时故障仍待补 | [当前 Docker 验收](../reports/docker-current-2026-09-20.md)、[HTTPS传输场景](../reports/docker-https-transport-2026-09-21.md)、[ENOSPC场景](../reports/docker-enospc-2026-09-21.md)、[Docker/Compose API](../reports/containers-api.md)、[真实 Docker 生命周期报告](../reports/containers-e2e-2026-09-09.md)、[Docker 浏览器报告](../reports/docker-browser-2026-09-09.md)、[Docker日志归档报告](../reports/docker-log-history-2026-09-22.md) |
| E03 | 创建有可核对内容的备份并恢复到受控目标，验证文件内容/元信息和权限；保留策略只清理自身范围；一致性钩子成功/失败分别呈现；中断、空间不足、恢复覆盖均有明确处理 | 进行中 | 子项通过 | [备份与调度模块](../reports/backup-scheduler-library.md)、[备份界面增量](../reports/backup-ui-2026-09-10.md)、[竞态复验](../reports/backup-scheduler-race-2026-09-14.md)、[真实浏览器](../reports/backup-browser-2026-09-14.md)；默认界面已接入版本核对、恢复计划和显式覆盖确认，真实 Master/Daemon/Chromium 已核对归档、恢复正文和任务回执，Linux真实备份仓库及恢复目标ENOSPC竞态已通过，见[空间不足报告](../reports/enospc-2026-09-19.md)；实际固定argv保存程序、保存失败补偿和恢复最新正文race6.174s已通过；新增 Linux 真实归档写入中 SIGKILL 场景：原任务显式转为 `interrupted/unknown`，未发布 snapshot，部分 ZIP 按对象身份清理，新 ID 备份成功，见[归档进程崩溃报告](../reports/backup-archive-process-crash-2026-09-21.md)；新增 Linux 真实恢复上传中 SIGKILL 场景：分片取消并清理、目标文件不发布、原ID保持部分中断结果，同ID不重放，新计划完成恢复并校验正文SHA，见[恢复进程崩溃报告](../reports/backup-restore-process-crash-2026-09-21.md)。物理掉电、设备缓存丢失及 Windows 运行仍未验证 |
| E04 | 时区含夏令时、节点离线、Master 重启、错过触发和同资源重叠；按声明策略执行且不重复补跑；创建者失去权限后新触发拒绝；节点离线仅继续已接受任务，不从陈旧计划无限产生特权工作 | 进行中 | 子项通过 | `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（备份1.856s，Master 5.816s）；真实 Chromium 保存/删除 `Asia/Shanghai` 计划 1/1、3.5s 通过；2026-09-19真实Daemon关闭/重连、受控时钟五次到期合并补跑及离线撤权race2.869s通过；2026-09-19定时任务接受后Master/SQLite实际关闭重开、双节点重连且任务/归档唯一race3.198s通过；2026-09-21 scheduler独立进程在accepted occurrence提交后强制终止、SQLite重开与无重复/新触发撤权子项 race 1.130s通过；同日真实 RC15发行包按真实分钟触发计划后同时 SIGKILL Master 与所属 Daemon，重启后 startupId变化、taskId/requestId不变、原任务SUCCEEDED且仅一份归档，退出0，详见[Master/Daemon 崩溃报告](../reports/scheduled-backup-master-daemon-crash-2026-09-21.md)；当前源码实际 `Scheduler.Tick` 的 pending slot 已保存但尚未 BuildTask、prepared fire/task 身份已保存但尚未 Accept 两个阶段均以 SIGKILL 中断；重启后分别只构造并接受一次、复用相同 taskId/requestId，重复 Tick 不重放，完整验证见[调度阶段崩溃报告](../reports/scheduler-staged-process-crash-2026-09-21.md)；另以独立子进程在共用 `metadataMutation` 事务中写入 schedule/fire 记录后、回执/审计/提交前 SIGKILL，SQLite重开未见部分记录且完整性检查通过，见[元数据事务崩溃报告](../reports/metadata-transaction-crash-2026-09-21.md)；2026-09-22 同一受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 两个真实分钟槽各只接受一次、第二槽绑定新源版本且两份快照恢复校验通过（1.5m），见[netem分钟调度报告](../reports/netem-schedule-wall-clock-2026-09-22.md)。上述是进程级/受控网络边界，不是 SQLite 设备写入中的断电；2026-09-27本机独立24h真实墙钟、1,440个成功分钟槽与末尾最新备份恢复已通过，详见[最终报告](../reports/local-endurance-24h-2026-09-27.md)；更长期、物理掉电、跨主机网络及Windows仍未验证 |
| E05 | 在独立且获准的测试环境操作支持的服务/计划任务/防火墙子集；只修改目标对象，展示差异、保留既有规则；模拟管理链路失联后按期限自动回滚，再连接可诊断；普通用户不能调用主机特权 | 进行中 | 2026-09-29独立systemd容器服务/定时器生命周期与列表真实验证通过，修复停止服务及未加载timer遗漏，见[平台报告](../reports/systemd-cgroup-2026-09-29.md)。跨平台能力探测、有界service控制、timer启停、firewalld及host.manage API已有实现；持久租约、确认、取消/过期回滚、失败诊断与重启恢复通过。2026-09-19私有网络真实firewalld管理端口阻断和期限回滚race25.998s通过，见[防火墙报告](../reports/firewall-private-2026-09-19.md)。9月20～22日无法启动manager的记录仅为历史环境失败；当前Linux对应子项已有证据。Windows及其他实际部署环境仍待验证 |
| E06 | 按 SDK 文档从独立示例目录构建新扩展；不修改桌面核心即可安装、开多窗口/标签、注册资源入口和后台任务、保存迁移现场；前后端沙箱与能力桥接实际拒绝未授权访问 | 进行中 | SDK 构建通过；`POST /extensions/{id}/tasks` 受 `app.use`（扩展资源）+ `node.read` 最小权限保护，bundle 在 sandbox iframe opaque origin 执行，Master/Daemon 集成和真实浏览器覆盖成功/拒权；真实 WASI guest 用户数据迁移、失败保留、回滚和事务恢复 race 通过，见[扩展数据迁移报告](../reports/extensions-data-2026-09-20.md)；真实成员浏览器上下文未授权组合1/1通过，跨设备状态迁移/资源身份校准1/1；2026-09-22 受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 独立包安装、多窗口、WASI节点任务、已接受回执丢失后的稳定请求键恢复、快捷方式与刷新现场均通过（49.4s），见[netem扩展报告](../reports/netem-extension-task-2026-09-22.md)；Windows 与长期跨主机远程场景仍待补，见[扩展跨设备报告](../reports/extensions-cross-device-2026-09-14.md)、[未授权浏览器报告](../reports/extensions-unauthorized-browser-2026-09-14.md) |
| E07 | 真实测试注册源提供至少两个版本扩展；浏览/获取/验证/授权/安装/启用/禁用/升级/失败回退/卸载跑通；篡改包、版本不兼容、能力提升均按策略阻止；禁用不再接受新任务，卸载默认保留用户数据并给出清理选项 | 进行中 | fixture 本地注册源提供 `v1.0.0`/`v1.1.0`；签名/完整性/回退/能力提升拒绝、启用/禁用/卸载数据保留测试及真实浏览器两版安装和 bundle 执行通过；后端用户数据迁移、失败保留、回滚和事务恢复 race 通过，见[扩展数据迁移报告](../reports/extensions-data-2026-09-20.md)；`TestExtensionRemoteCatalogAdminAPI` 验证 Master API 经 loopback HTTPS 浏览/安装、普通成员拒权、同一 CA 的叶证书续期，以及切换新 CA 后旧信任失败关闭、显式更新信任后安装新版本，race 通过，见[原注册源报告](../reports/remote-catalog-rotation-2026-09-20.md)和[CA 根轮换报告](../reports/remote-catalog-root-rotation-2026-09-21.md)；2026-09-22 受控 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 本地注册源两版本浏览、安装、升级、禁用/启用和沙箱bundle执行通过（13.6s），见[netem注册源报告](../reports/netem-extension-catalog-2026-09-22.md)；公网远端部署、跨主机高 RTT、证书信任更新的实际运维流程和 Windows 仍待补 |
| E08 | 记录参考设备、浏览器、RTT、日志速率和带宽；8个窗口、2个活跃终端、万条目录与后台传输并行；原规划p95 ≤50ms独立保留，未达标如实记录；2026-10-08用户接受约100ms并明确要求有界尝试后收尾 | **按用户调整范围收尾** | 当前同源原始首120事件：Chromium38.8ms PASS；Firefox66ms、WebKit132ms及独立95ms仅在原50ms断言FAIL。较早同源WebKit89ms，完整记录89–132ms波动，不保证每轮≤100ms。原工作负载、保护ACK、恢复和最终16MiB目的校验未放宽；生产构建/121单位/WebKit21项功能守卫/最终类型检查通过。最后flat-body候选79→139ms未获稳定收益，未采用。新增Chromium DPR1静态圆角截图不稳定（旧定位同处复现）仍记未解决，不计PASS。全部本轮fixture清理退出0；原严格50ms作为后续非阻塞优化。详见[收尾报告](../reports/e08-bounded-closeout-2026-10-08.md)，[完整历史](../reports/e08-render-architecture-2026-10-06.md)保留。 |
| E09 | 从干净依赖环境按说明构建并初始化 Master/Daemon/前端；至少两个节点和不同权限用户跑通；交付 Linux 与 Windows 产物/构建入口，分别记录 Linux 运行、Windows Job/ConPTY 真机结果；验证配置/数据迁移、备份恢复、升级回退和示例扩展开发说明可复现 | 进行中 | 2026-09-29 RC26六包/2,215条内部记录/三份Web与独立启动、停机恢复及兼容RC25回退通过；Linux systemd与委派cgroup（含CPU实际限流）已有本机隔离证据，见[平台与发行报告](../reports/systemd-cgroup-2026-09-29.md)。Windows完整启动/迁移/升级回退及其他未覆盖组合仍待独立验证。Linux 构建/初始化与多节点权限子项已有证据；`make sdk`、`make web` clean 依赖安装和构建退出 0，`make windows` 及 runtime/runlog/systeminfo Windows 测试入口交叉编译通过，OpenAPI YAML 当前为 105 路径 / 122 操作；storage、backup（含本轮 scheduler stage/restore crash）、containers HTTPS E2E 和本轮 `internal/filesystem` 测试程序均以 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c` 编译成功（仅构建证据，不代表Windows运行）；rc3独立发行包的停机快照恢复及兼容rc2回退实际通过，见[rc3报告](../reports/release-rc3-2026-09-19.md)；2026-09-19从全新Go模块/构建/npm缓存完成Linux/Windows、前端、SDK及参考包构建，产物与rc4一致，见[冷构建报告](../reports/cold-build-2026-09-19.md)；rc5含多核 Windows CPU 配额修复、rc7含定时备份一致性 UI 且包冒烟通过；rc13六包SHA校验、停机恢复与兼容rc12回退通过，见[rc13报告](../reports/rc13-current-source-2026-09-20.md)；rc14包证据见[RC14报告](../reports/release-rc14-2026-09-21.md)；RC23证据见[历史报告](../reports/release-rc23-2026-09-22.md)。2026-09-29 **RC25** 六包/2,207条内部记录/三份Web一致性及独立启动/停机恢复/兼容RC24回退通过，见[本轮报告](../reports/local-closeout-2026-09-29.md)。历史2026-09-27 **RC24** 已包含冻结UI、检查点Worker、Compose授权与稳定窗口DOM修复：六包构建/外部SHA及2,191条包内记录核验、全部三份Web与当前构建一致、独立SDK/参考包签名、TLS/双Daemon、停机快照与兼容RC23回退通过，见[RC24报告](../reports/release-rc24-2026-09-27.md)；完整真实功能51/51见[功能报告](../reports/local-functional-2026-09-27.md)。Windows Job/ConPTY等11项必需原生检查已于[第五轮真机](../reports/windows-fifth-report-2026-10-01.md)验证；2026-10-05[第十六轮](../reports/windows-sixteenth-report-2026-10-05.md)补齐visible WebKit查询恢复子项。Windows完整启动/迁移/升级回退及其他未覆盖组合仍待独立证据，任意schema降级不作承诺 |

E05 的真实防火墙验证不授权改动当前生产宿主机。E09 的 Windows 真机缺失必须写“环境缺失”，不影响继续实现代码，但不能计作全部验证完成。E08 的设计目标不代表已测性能。

Windows Job/ConPTY/监控/系统适配、独立 systemd manager、远程 Engine/网络故障和长期物理故障的可复现实验入口见[平台验收运行手册](../../operations/PLATFORM_VALIDATION.md)。该手册只提供真实环境运行步骤；当前 runner 缺少对应环境，不能把命令存在或交叉编译结果计作通过。

## 6. 横向核对

- 默认应用和扩展应用均服从账号/节点/资源权限及状态版本协议；窗口标题、确认框与任务记录显示实际目标。
- 服务端成功结果到达前，界面只能显示已提交/执行中；本地拖动、输入和导航立即反馈。
- 非文本桌面区域使用应用式选择、框选和拖放；编辑器、终端、日志等需要复制的区域保留受控文本选择。不得全局封禁浏览器全部快捷键。
- 节点后台任务、PTY 会话、真实实例、前端窗口和浏览器标签的生命周期分别处理。
- 同一用户多个浏览器标签/设备不能静默覆盖彼此较新的状态；显式退出处理未同步草稿与本地数据隔离。
- 不将私钥、令牌、密码、真实用户草稿或不必要的文件正文写进验收日志。
- 模拟后端仅可作为测试替身；默认产品流程不得使用固定成功响应、虚构节点/实例或假进度。

## 7. 执行台账与完成判定

历史初始统计（2026-09-09）：14 个功能项、17 个原规划场景、9 个补充场景当时均尚未整项通过。此统计不能当成2026-10-05当前待办数量；最新证据和用户调整后的本次完成判定以本文件顶部及最新剩余表为准。部分测试通过仍不等于未执行的实测通过。

2026-09-09增量证据：F03/F04见[用户角色及节点维护换钥](../reports/administration.md)；F05及A04/A14配置与真实自动启动子项见[实例设置](../reports/instance-settings.md)；F06/F07/F08及A01/A02/A06/A07/A17见[真实数据流集成](../reports/stream-integration.md)、[文件API](../reports/files-api.md)、[桌面浏览器](../reports/desktop-2026-09-09.md)；A12跨节点/同宿主关系保护与11项真实TLS测试见[传输报告](../reports/transfers.md)；F11/E02见[Docker/Compose](../reports/containers-api.md)，宿主 docker-host PTY 双主体测试见[容器报告](../reports/containers-2026-09-09.md)；F12/E03/E04见[备份与调度](../reports/backup-scheduler-library.md)及 Master API 集成测试；A13/E06见[SDK报告](../reports/sdk-2026-09-09.md)；F13/E05 的系统管理浏览器边界见[系统管理报告](../reports/system-management-2026-09-09.md)。当时真实磁盘满、完整故障组合、容器实例日志归档、监控、系统管理危险变更、扩展安装生命周期和 Windows 真机仍未验证；其中 Docker 日志归档已在 2026-09-22 补上前端入口和浏览器证据，其他边界继续按各项记录。均为已明确列出的子场景，未将交叉编译、替身测试或尚未运行的组合计为整项通过。

2026-09-10最新真实复验：在隔离 fixture `.local/fixture-1866434427` 以真实 HTTPS Master、两个 Daemon 和 Docker 29.4.1 运行，`files-terminal.spec.ts` 2/2 通过（27.1s），`npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` **15/15 通过（2.2m，退出码 0）**。本轮覆盖真实登录、节点/实例、文件、PTY/终端、日志/传输、设置/编辑器、模板创建、账号权限、Docker/Compose、扩展两版本 sandbox 和云工作区；证据详见[桌面真实复验](../reports/desktop-2026-09-09.md)、[工作区报告](../reports/workspaces-2026-09-10.md)和[SDK报告](../reports/sdk-2026-09-09.md)。Docker stop 的 300 秒服务端等待已由 Engine transport 5 分钟 header timeout 覆盖，容器回归与真实 Compose 浏览器 1/1 通过，详见[真实 Docker 报告](../reports/containers-e2e-2026-09-09.md)。Master 数据链接重连窗口也已修复并由完整真实套件验证。该证据不改变矩阵中 Windows、真实 ENOSPC/掉电、远程/长时负载、危险主机变更、扩展网络注册源分发和完整资源迁移等未覆盖项。

2026-09-10交付与安全增量：`docs/api/openapi.yaml` 通过 Python `yaml.safe_load`（OpenAPI 3.1，94 路径/108 操作；审计与节点维护增量后更新为 95 路径/109 操作，本轮任务重试后为 96 路径/110 操作）；`Manager.List` 跳过升级回滚快照并拒绝清单身份错配；防火墙租约回滚开始前持久化 `rollback_in_progress`，过期/取消直接执行路径写入终态，普通/race 租约测试通过。`make check`、提权 `make test`（`go test -race ./...`）、前端 `npm run check && npm test`（26/26）、`make windows` 和 Linux 三个二进制构建均退出 0。

2026-09-10审计与节点维护增量：新增 `003_audit_node.sql` 将审计记录统一为毫秒时间并回填节点身份；`Store.Audits` 提供管理员可分页筛选，UsersApp 增加用户/节点/资源/动作/时间检索。节点撤销改为事务幂等回执并保留实例资源，NodesApp 增加管理员确认入口。`go test` 存储/ Master 定向集成、`make check`、`cd web && npm run check` 通过；OpenAPI 更新为 95 路径，证据见[审计与撤销报告](../reports/audit-2026-09-10.md)。

实现阶段请在此追加证据索引，并在 [PROGRESS.md](../../execution/PROGRESS.md) 同步当前状态。建议每条记录格式：

- 编号及子场景：
- 代码/入口：
- 运行命令与环境：
- 结果、退出码和时间：
- 日志/截图/产物：
- 未覆盖范围及下一步：

最终报告必须明确区分：已实现且已验证、已实现但未验证、尚未实现、环境阻塞。全部必需功能与关键场景具有真实证据且缺口清零，才可以宣布全范围完成；后续仅剩视觉和非关键体验细化时也要具体列出。
# Windows returned evidence update (2026-09-30)

See [second Windows report](../reports/windows-second-report-2026-09-30.md): three
browser suites executed with 118/119, 115/119 and 111/119 passes respectively;
Go unavailable meant native and full Go checks did not run. Windows acceptance
remains incomplete; prior native results are not new validation.

## Third report corrections (2026-10-01)

The [third Windows report](../reports/windows-third-report-2026-09-30.md) ran
Chromium 120/120, Firefox 116 pass / 3 fail / 1 skip, WebKit 112 pass / 7 fail /
1 skip; Go events were 330 pass / 22 fail / 15 skip. Nine of ten required native
checks passed, while ConPTY failed and race was blocked by a missing compiler.
See [corrections and local regression evidence](../reports/windows-report-fixes-2026-10-01.md)
for directory handle access, ConPTY stdio, portable real-process fixtures,
editor history, browser input and network/transfer outcome fixes. Local Go race,
frontend units and Windows cross-compilation pass; Windows runtime retest is
still pending. The new runner checks 27 named regressions, 11 native cases and
affected browser scenarios, and can supply a verified portable race compiler.
No entire matrix item is promoted to complete from these local results.
