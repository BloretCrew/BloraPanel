# Historical snapshot before the non-release closeout (2026-10-08)

This preserves the full prior record. Relative Markdown links were adjusted for this archive directory. Current status is in the parent document; pending items below are historical.

# Blora Panel 执行进度与续接记录

## 2026-10-08 性能优化收尾（当前状态）

本段覆盖下方所有历史接管中的“正在运行 / 下一步 / R1开放”等措辞。用户明确接受罕见极端压力下约100ms，要求只再做少量尝试后收尾；本轮已完成，不继续追求全浏览器严格50ms。原50ms目标与失败记录仍完整保留，不写成全部通过。

当前产品同源正式复核：Chromium 38.8ms PASS、Firefox 66ms FAIL、WebKit 132ms FAIL及独立复测95ms FAIL（均为原始首120事件，FAIL只在未改动的50ms断言）。同源较早WebKit为89ms；因此记录89–132ms波动，不能承诺稳定≤100ms。八窗口、双真实PTY、万条目录与最终16MiB目的校验保持。最后只复测了一个已有flat-body候选，前次79ms未复现、此次139ms，无可靠收益而拒绝；未引入新产品实现。

保留545源4c8db790/169bundle5a04e608；生产构建、121单位、WebKit21项原生像素/几何/输入/保护ACK/恢复检查、最终类型检查和空白检查通过。新增Chromium DPR1静态圆角截图检查仍不稳定，旧定位同处复现，失败保留，不伪称修复或通过。1110部分body最终149ms、1112原transform对照109ms和1114自动行跳过0均未进产品。

本轮五个fresh夹具均确认清理退出0，浏览器测试均结束，无运行中的本任务服务。R1原严格50ms改为后续性能改进，不再作为本轮继续工作的阻塞项；R2/R3/R4历史状态不变，R5发行仍按用户要求延期，未提交或推送。最终数字、验证范围及限制见 [收尾报告](../../acceptance/reports/e08-bounded-closeout-2026-10-08.md)。

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

## 当前接管点（2026-10-08，不以历史记录的“运行中”为准）

当前产品新增独立native/Worker解析并行启动，仍等待二者完成后才保护输出/推进sequence，receive queue/resize顺序/ACK/UTF8边界不变。旧串行原生控制报告0225确认只startedTogether=false，其余保护/有序/恢复通过；新版本0228 Firefox7/7(27.734s)、111单位/生产构建通过。539源e1102342811fb71db25010c24ba76d0b7497c21d297ad228b11b26a1b49249b1；168bundle3a360b786d3655a84b386df60abf9a6533b9cb42c45bf512a09b4d3411f9bb33。正式0231 Firefox90秒bounded **51ms FAIL**（2720事件，首120/52、steady2600/51，6/4770长帧，五目的校验，双PTY约118kB/s、max未ACK86950）；62584退出1/23258夹具清理0，不关闭R1。现在WebKit关键功能88349运行；下一原生CSS窗口自身shadow诊断的独立像素比较，未进入产品。

原生alpha-mask实验Firefox光学两DPR通过，但WebKit1.25 max48失败，且真实90秒Firefox62ms FAIL（2445事件、七目的校验、双PTY约119kB/s）。未采用，失败比较保留；93900退出1/49476夹具清理0。以下不可变日志300秒53ms是上一产品结果，不能混算新源。

E08/R1仍开放。不可变终端输出日志候选在正式Firefox90秒bounded达标：**p95 50ms**，2820事件（首120/54ms，steady2700/49ms），5/4916长帧，8次16MiB目的校验，双PTY120392/120751B/s、max未ACK66348。74455退出0/85387夹具清理0；这是单轮阈值边缘PASS，不是全量完成。

111单位、生产构建通过；浏览器产品原生恢复/实际选区/covered/partial六项通过。仅由恢复系统生成的不可变数组可复用序列/字节校验，外部替换/导入仍完整校验；立即journal与ACK门控不变。539源 **b98487b14d4f8c9ae1c3e1a54316715143a0dfd5cae615fc5888799c41bffe3f**，168bundle **32a54525756ceb5eb9803d80eebd120e568ad1869da7aa128cacce47aba6830a**。数据only增量和真实view-change去重保留。

同源正式Firefox300秒bounded最终 **p95 53ms FAIL**：8520事件，首120/58ms，steady8400/52ms，38/15285长帧，23次16MiB目的校验，双PTY118998/118988B/s、max未ACK63670。61233退出1，41824/fixture-354449102清理退出0；报告 `web/.local/e08-20261008-0205-immutable-firefox-300-report.json`。单轮90秒PASS不足以证明持续达标。下一仅原生alpha-mask阴影光学诊断，应用正文/字形和既有零像素守卫保持；通过独立光学比较后才测实际收益。压力期间不并行构建、功能测试或剖析。

原native-shadow-sibling原型仍失败（最新FF两DPR max34），未进入产品。严格原比较及失败证据完整保留在tests/experiments和playwright.experiments.config.ts独立入口；成品原有正文/字形/shadow-exposure断言不动。0150原生剖析是诊断，原60轨迹49ms不计正式验收；初120在部分线程已滚出buffer，不能用其解释冷启动。原始profile私有，仅数值/静态标签分析。

R2/R3/R4闭合，Windows不重复；R5发行延期，未提交/推送。用户报告的select斜角/文字裁切修复已三引擎深浅截图与键盘验证。

## 历史续接记录

2026-10-08 最新有效结果：正式 dedup 产品 Firefox 90 秒 bounded **p95 55ms FAIL**（2645 事件，首120/62ms，steady2525/54ms，19/4629长帧，7次16MiB目的校验，双PTY118726/118958B/s，max未ACK66398）。浏览器28581退出1；fixture97515清理退出0，当前无真实压力/fixture。原工作负载、首120、50ms、恢复/ACK与全部零像素断言保持。当前539源016ec7cf69939c319c00abbd36212dbf6f66cfa32dbd42c352c483b5ae18292e /168bundleb206021c9b6093e89b6b309a4b03765c3c5badab64fd52df0bf94eeffa7fd49d。Firefox/WebKit原生恢复、实际scroll/selection、covered/partial各6/6通过；本次结果不足以关闭E08。新增仅内容空白光学阴影 sibling 对照在DPR1及1.25均出现RGBA差异，原型未进入产品，不计PASS；失败报告保留。下一具体动作：独立 Firefox 原生线程剖析，在原八窗口/双PTY/万目录/连续校验传输负载下定位主线程与绘制成本，只输出数值摘要，剖析结果不作为性能验收。R1开放；R2/R3/R4闭合，Windows不重复；R5发行延期。下方历史“正在运行”不代表当前。

2026-10-08增量首轮正式结果：select-delta Firefox90秒bounded p95 52ms FAIL，2695事件，首120/58、steady2575/51，7/4764长帧、两目的校验、PTY约106.5/106.9KiB/s、max未ACK71539。实际1388原生snapshot put中688增量只耗70ms，700完整耗1052ms；journal2639写/总67ms。3160退出1/3664清理0。增量收益不足，继续排查完整保存来源，不据此关闭E08。TerminalModel protectView原先对相同scroll/selection也提交set/delete并迫使完整快照；现在逐字段比对，仅实际改变立即提交，输出保护/解析/流/ACK不变。新原生重复scroll事件、实际scroll/selection、数据库only恢复+重新挂载检查通过；Firefox4880及WebKit8063各6/6、原恢复/covered/partial守卫保留。38228构建通过。最新539源016ec7cf69939c319c00abbd36212dbf6f66cfa32dbd42c352c483b5ae18292e/168bundleb206021c9b6093e89b6b309a4b03765c3c5badab64fd52df0bf94eeffa7fd49d。fresh fixture97515/fixture-169697150已ready，正式Firefox90秒bounded已启动，无实验flags/原负载阈值与首120保持；下一报告→退出→fixture清理0，再按收益继续。R1开放，Windows不重复/发行延期/未提交推送。

2026-10-08最新续接覆盖下方历史当前/运行：0822正式RGB产品Firefox300秒bounded p95 51ms FAIL，8645事件（首120/56，steady8525/51），20目的校验，PTY116073/115945B/s，最大未ACK81460；94576退出1/96168清理0。0847环境中断只有空输出、没有JSON，旧11547等会话不存在，不计执行/PASS。当前候选为完整基线+64条/128KiB有界terminal-only IDB增量（无增量链、原原子事务、同步journal/ACK/云端导出保持），冷状态变化/到界/所有权不一致回原完整快照。108单位、正式构建通过；0839 Firefox/WebKit原生恢复及covered/partial各5/5，当前Chromium两个原生恢复通过。原软件WebGL DPR1.25像素守卫仍失败；released buffer=true/renders3/pending=false，不弱化原断言或宣称功能全PASS。

用户最新下拉框问题已直接复现，是WebKit原生appearance:auto使12px角斜切及28px文字裁切，不是截图缩放。CSS仅接管普通单选框外壳，主题填充圆角/箭头、显式文字空间，菜单语义键盘原生，紧凑三筛选仍28px/forced colours原生回退。三浏览器×深浅色六次实际截图/文字空间/ArrowDown+Enter选择全部退出0；截图固定合成示例。539源3324bac7d908fb758440eb5e8065a9948ed0259e4aa48c57ee8739e76535589a、168bundle38a5239050423ec180b37ef8092a35a1775fd445146fdcf3394329f6ee591def；Vite38869。3160正式Firefox90秒bounded正在运行，fresh fixture3664/fixture-1186847609，原八窗口/双PTY/万目录/16MiB循环目的校验/首120/50ms保持，无renderer/optical诊断flag；只读快照delta/full计数完善。下一报告及最终目的校验→浏览器退出→fixture Ctrl+C确认清理0，再继续原60秒三引擎/持续稳定性或新结构优化。R1开放，R2R3R4闭合/Windows不重复/R5发行延期/未提交推送。

2026-10-08 0808Firefox70pass/1仅Chromium原生IMEskip、无fail/flaky（359.846s）；0809WebKit70pass/1同skip/2fail（395.685s），平移6reads≥5及workspace刷新浏览器tasks/summary access-control页面错误。0817仅两失败独立各3次6/6(54.663s)，首失败仍保留，因果未确立不合成全PASS。0815Chromium40pass/2fail（195.920s），native-list零差异在DPR1.25 held比较changed340/max39，nativeWebGL DPR1.25同步释放后changed0未满足>0；0820各DPR三次控制重现，待独立native数值诊断，原断言不变。源/产物仍6e8eeb6b/5ead3af9（产品未改）。所有功能进程已退出，现在96168 fresh fixture3117097121及0822Firefox300秒bounded正式产品运行；无RGB/window/shadow/renderer诊断flags，全部原资源/首120/50ms保持，只收取数值，不作成品闭合。下一末轮目的校验/报告→浏览器结束→fixture Ctrl+C确认清理0，再解决像素/启动失效问题；不与压力并行跑构建/功能/剖析。Vite29529；R1仍开放/Windows不重复/发行延期/未提交推送。

2026-10-08 RGB产品最终0807构建23184退出0、102/102单位16939退出0；538源6e8eeb6bcc7a749820699cc3cc9413c9a3946d9588a99fc34a6d579798ffee20，168bundle5ead3af9ca46c6c1e4b6b4ef5eb3645f93b51986c63a41ae3a5b61a1b501476a，指纹按web相对路径+NUL+rawSHA256排序聚合。0758Firefox23634退出0、4/4(26.677s)，24全光学RGBA0。新增原生透明fallback/奇数行stride回归，并统一ready/opaque/encoding失效。0746 fixture84218已确认清理0。现在0808Firefox功能99151、0809WebKit功能74443依次启动（只功能，不作性能声明），原全屏正文/shadow/glyph/即时恢复守卫保留。下一收取两者，然后Chromium功能；全部完成后fresh正式同源三引擎原60秒和Firefox持续长测，无optical/renderer实验flags，首120不排除。R1开放/Windows不重复/发行延期/未提交推送；Vite29529。

2026-10-08 all RGB光学0746完整Firefox90秒bounded50ms PASS：2745事件、initial55/steady50、7/4817长帧、六目的校验、双PTY约116KiB/s/max未ACK82545；38913退出0，84218请求清理。五variant实际native BMP24 decode、原720×480/desktop1440×960、不减分辨率/颜色/阴影/负载。不是成品验收：这是明确RGB诊断，无app内容捕获；不能关闭E08。开始集成仅material-cache使用的lossless opaque encoder，逐样本alpha255认证，透明源回退原PNG，现有decode发布/生命周期/回退保持。下一正式构建、全部材质/像素/原生终端恢复与三引擎原60/持续长测，必要时再推进增量持久化，不把阈值边缘PASS当稳定性闭合。R1开放/Windows不重复/发行延期/未提交推送；Vite29529。

2026-10-08 壁纸RGB原生decode候选：0730 atlas180px测试虽2/2，但四个palette字符串错误而fallback ice，只证明默认光学，不能作六主题证据；保留并修正palette枚举与正向data-theme/data-palette断言。发现native opaque Canvas PNG仍colorType6，增加仅既有壁纸RGBA且逐样本alpha255认证的lossless BMP24（不捕获app/字形/图标/shadow）。0738 Firefox30920 2/2(21.521s)、0741 WebKit42103 2/2(30.767s)，六palette×深浅×DPR1/1.25共每引擎24完整光学表面RGBA0；48990/42347类型0。无性能收益声明。撤回row adapter后源537 e3ce5f6081eeb5ca5617a71fc19a070cd0c9d28a70bbbe85512eb858017966e7/168bundle31b9d78fcfeb7c597b49c11202a3adf2257fac2a4a2d508bd03f05504a8f8c94，0727构建80573退出0。84218 freshfixture1316866127，0746 Firefox90秒bounded明确all五壁纸光学variant RGB诊断开始（原分辨率、不含正文/图标/shadow/字形捕获），仅基于数字metadata核实实际decode与50ms。下一收取最后目标校验/性能、确认fixture清理0；若收益成立才集成并完成同源全验收。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 visible-row正式0715 Firefox90秒bounded仍54ms FAIL（2695事件、initial61/steady53、13/4725长帧、八目的校验、双PTY约117KiB/s/max未ACK69090）；实际两个native-visible-rows已安装，不因终端单独正确性通过宣称收益。94353退出1/97007清理0，源385d7a9a/bundle63d310d8失败保留；原生行委托从产品撤回并移入tests/helpers/native-terminal-rows.ts，新独立native参照/首RAF/2026/DPR守卫保留。partial viewport生产仍未证明延迟收益。下一只壁纸光学共享atlas180px候选，不接触app内容或shadow，先测原生光学差异，再fresh明确诊断/50ms，不计成品验收。R1开放/Windows不重复/发行延期/未提交推送；Vite29529，无压力/fixture。

2026-10-08 native visible rows新结构candidate：0656数值几何确认两原生screen底部只露7/33CSS pixels，原isIntersecting持续true；首120末计数1624/1680行，末轮9184/8400（8394退出1/47931清理0，短诊断不计正式验收）。增加单独pinned6.0.0 default原生renderRows范围委托，只生成可能露出行并保留两行overhang；无app内容副本、无自定义字形、无输入/解析/恢复/ACK改动。露出扩大同步恢复原生完整已提交行；DECSET2026之前恢复隐藏的已提交行，原native节点保持atomic held图像；接口不符/WebGL仍完整原生路径。0701类型61886退出0，0702 Firefox63057原全隐藏与partial DPR1/1.25三项通过。0705类型37022/Firefox91735两DPR新增独立原生参照RGBA0、首RAF露出、hidden update/2026 held/release/selection/真正行数量减小及dispose恢复通过2/2。下一WebKit完整五终端回归和正式构建，然后fresh正式90秒bounded性能，不混旧产品PASS。R1开放/Windows不重复/发行延期/未提交推送；Vite29529，无真实fixture/压力。

2026-10-08 native paint0648独立zero-soak诊断完成：原首120事件62ms FAIL，采样3.43秒/测试24.286秒/最终目的校验11.555秒、双PTY约119KiB/s、最大未ACK86052；79663退出1/47255清理0。从gesture开始至最终目的校验，两个实际default renderer仍收到native IO isIntersecting=true，分别7700/7504次原生行replaceChildren（该计数不是仅3.43秒区间）。入口原样委托，不读出正文/节点/应用像素，短测不计完整E08。源e3ce5f60/产物873f41a9，下一记录实际screen/overflow交集数值及首120末时计数，确认暂停条件再改生产；所有原负载/解析/恢复/ACK/零RGBA/50ms保持。0656类型3293退出0，fresh fixture2351031612/47931及数值短诊断已启动；下一收取目的校验/报告、清理0后继续生产路径。R1开放/Windows不重复/发行延期/未提交推送；Vite29529。

2026-10-08 terminal partial0643正式Firefox90秒bounded54ms FAIL：2620事件、initial56/steady54、10/4573长帧、六目的校验、双PTY约114KiB/s、最大未ACK85505。72237退出1/78468清理0。真实前两host full410024分别paint189/891CSS pixels，第三空host378601未裁；nativeScreen实际IO暂停/行更新未测，不能只以dataset.paintOccluded=false证明仍全屏绘制。源e3ce5f60/产物873f41a9未达50ms，当前candidate不计达标；下一新增独立zero-soak numeric native IntersectionObserver+replaceChildren委托计数诊断（不改输出/解析/ACK），确认是否背景原生默认renderer仍更新；仅原首120/完整真实资源与末轮校验，不能替代60秒验收。若实际屏幕已暂停则转向其它真实开销，不因小bbox假定因果。Vite29529，无压力/fixture；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 terminal partial0638 Firefox24997 3/3(26.199s)/WebKit28816 3/3(35.236s)通过：DPR1/1.25右/左/下局部原生默认终端均整屏RGBA0、full host/world position、ACK/保护/refresh/idler露出及原全隐藏链路保持。构建98732因新增测试Node.remove类型失败，50358六几何单位通过；修正为HTMLElement，并排除仅管理器拥有的paintviewport尺寸RO自反馈，region/host/祖先RO保留。新增held露出包络单调回归。0641构建74305退出0、Firefox82883 9/9(48.156s)/WebKit4318 9/9(61.174s)，含动态菜单登记、未受控写入前后、纯translation/narrow/八resize/即时reload及三终端检查。当前537源e3ce5f6081eeb5ca5617a71fc19a070cd0c9d28a70bbbe85512eb858017966e7/168bundle873f41a9d87e2746bc7c15c7f9a5142e3872e75fd03d3d2947efe1875e5f6d58。78468 fresh fixture-143686774，0643正式Firefox90秒bounded原负载开启，无光学/renderer诊断，实际partial面积数值新增只读记录。下一收取末轮校验/性能，清理fixture确认0；若无收益继续结构候选，若有收益完整同源三引擎/原60+长测/功能。Vite29529，R1仍开放/Windows不重复/发行延期/未提交推送。

2026-10-08 alpha-two0632 Firefox90秒bounded55ms FAIL（2645事件、initial68/steady54、14/4614长帧、七目的校验、双PTY约117KiB/s），80253退出1/13676清理0。32corner/64bands的959708→454406CSS pixels，实际可见341066，约减半仍没收益，不集成。进入产品native partial terminal viewport候选：复用已有原生overflowviewport缩小绘制范围，full host宽高/world position、grid/滚动/解析/worker保护/ACK/恢复不变；手势中露出包络只扩张，几何/结构失效重测，向外1px保守AA。原全隐藏回归及所有原零像素断言保留，新DPR1/1.25右/左/下局部遮挡整屏RGBA0+actual default renderer/ACK/恢复/idle露出检验待运行。下一类型/构建、两引擎原守卫及新默认字形像素检查，再fresh正式Firefox90秒bounded。不带alpha诊断，不混旧源PASS。Vite29529，无压力/fixture；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 alpha初版0629 Firefox61820 2/2整屏RGBA0，但WebKit4208 DPR1 changed41086/max1 FAIL，小数DPR原回退0通过；未放宽。改为每band保留原完整native quad/UV并只加overflow视口，0631 Firefox87012 2/2(12.685s)、WebKit86775 2/2(21.238s)整屏全部RGBA0、深浅/重叠/held reveal/resize/小数DPR回退通过。0629类型/构建38906退出0；rollback后537源381c79645cd20ae4191f67d4867ade3fa4d41fd607c7a8fa04dea0f1340a5fe6/168bundle9a6a341450f2cf97fedcc66c00861b8ea60708fdc7681c17b88b79dcafd07798。13676 fresh fixture-22047854，0632 alpha-two Firefox90秒bounded80253运行，诊断保持全资源/首轮采样/50ms，实际32corner/面积缩小必须成立，不计产品PASS。下一收取结果/末轮目的校验并确认fixture清理0，再按收益决定是否集成，若无收益继续终端部分原生overflow区域候选。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native motion driver0623正式Firefox90秒bounded54ms FAIL：2695事件、initial71/steady52、11/4722长帧、八次目的校验、双PTY约117KiB/s、最大未ACK67620。31610退出1/84522清理0；直接driver/owned style stamp撤回，没有证明收益，保留新增同事件capture/bubble未受控写入回归及安全私有几何对象复用。0525 WebKit13/13、102/102单位通过；0530实际终端/held-menu/离屏Firefox3通过、两强制WebGL像素测试因环境WebGL2 unavailable失败，不算通过；WebKit5/5全通过。新native-shadow-alpha仅诊断：原native corner PNG按非零alpha最小矩形切为2/4band，使用原URL/原1:1采样，无重编码/无app捕获，全部非零texel恰好一个band认证；原parent完全遮挡裁剪保留，其它缩放/unsupported回退。下一两引擎完整RGBA0/实际面积对照，再fresh真实90秒bounded；不改原守卫/50ms/负载，不混诊断和成品PASS。Vite29529，无压力/fixture；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native-shadow-partition0450 Firefox90秒bounded55ms FAIL（2620事件、initial62/steady53、12/4572长帧、六目的校验、双PTY约116KiB/s），9027退出1/12112清理0。实际32strip1,490,102CSS pixels确低于原1,628,332，但高于原剔除27hidden后的1,037,081，不采用。进入产品新的native motion driver：已知纯translation直接发布私有缓存矩形并复用对象，观察器只对完全相同的已签认style跳过通用序列化；拖动前drain先前frame/其它window/结构/appearance变更，未签认/晚到write仍走原立即测量路径。恢复、控制、解析、50ms和原零RGBA守卫保持。新增同一原生指针事件capture/bubble阶段未受控contour+layout变化回归，不接受漏读。0520构建87041退出0、六遮挡单测10788退出0、Firefox41434定向回归退出0；current537源8a2c311caade49ab8eed39e64e5b4e9d5b5137cd2ffa49402154f93c2564a295、168bundle86c5ce427016c3937e30fdc4244492b81d48fa4e6918c8727f10fa75b8e6f1b8。0525 WebKit及全单测运行，下一二者完成后fresh正式Firefox90秒bounded，无shadow/bar/renderer诊断，不混旧来源PASS。bar光学测试默认改为已正确的native-full，PNG/cropped失败控制仍显式可运行，不改阈值。Vite29529，无压力/fixture；R1仍开放/Windows不重复/发行延期/未提交推送。

2026-10-08 grouped shadow-composite0427 Firefox90秒bounded54ms FAIL（2670事件、initial63/steady52、12/4666长帧、七目的校验、双PTY约117KiB/s），69280退出1/88370清理0。实际32strip面积3,105,164CSS pixels高于原64patch1,628,332（其中27完全隐藏），解释对象少却更慢，未采用。新partitioned模式先组合全部原native阴影到隔离Canvas，再认证round-contour中心每个alpha=0，拆top/bottom/left/right非重叠四周边并只裁alpha0，保留全部非零影；没有app采样。0435 Firefox95207 2/2(13.324s)、0438 WebKit24958 2/2(21.094s)均前台0/光学max≤4 mean≤.1，小数DPR全RGBA0回退；类型初次19246因旧无arg Page.evaluate失败，修正显式模式后50769类型0。新增独立native-shadow-partition性能枚举和实际32decode/面积小于原patch总面积的正向assert。下一fresh真实90秒bounded，不改50ms和原零RGBA守卫。产品99a20129/303bdb97仍52ms/R1开放，Vite29529；Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-composite原型0415 WebKit90486 2/2(19.371s)、Firefox71043 DPR1.25回退1/1(8.256s)通过；合并0410 FirefoxDPR1本候选两引擎均有重叠/held reveal/resize/dark光学max≤4、mean≤.1、前台正文RGBA0及小数DPR原八patch全RGBA0。只有shadow PNG进入额外Canvas；不会捕获app/文字/字段。四条完整perimeter strip可能增加覆盖区域，性能收益尚未成立；没有采用、不能以这些新光学对照替代原全画面exposure零RGBA守卫。新增native-shadow-composite明确32条decode/八窗口ready正向前置、最终记录实际strip面积；所有原负载与50ms不变。0420类型与fresh fixture准备中，下一真实Firefox90秒bounded独立对照。产品99a20129/303bdb97仍52ms，R1开放，Vite29529；Windows不重复/发行延期/未提交推送。

2026-10-08 bar-shadow-native0404真实Firefox90秒bounded51ms FAIL（2745事件、initial56/steady50、5/4790长帧、七目的校验、双PTY约117KiB/s），51324退出1/38049清理0，不采用；0355 Firefox54517 native-full2/2通过、类型首次54340因Page.evaluate字面量推断失败，明确泛型修正后75642类型0。新shadow-composite-diagnostic将原生decoded PNG八周边合成四条native strips，仅已有光学图/数字几何，不取app内容；DPR1整数几何/cache decoded才启用，其它立即原八patch回退，appearance代次/32key上限/URL清理。0410类型2993退出0；Firefox96813 DPR1重叠/held reveal/resize/dark正文0，光学max2/255且mean<.1；DPR1.25在初始化前desktop.state未就绪而setup失败，保留、补topbar可见前置，不改断言。下一WebKit及DPR1.25复核，再fresh真实native-shadow-composite；仍只候选，产品99a20129/303bdb97的52ms，R1开放。Vite29529，无压力fixture；Windows不重复/发行延期/未提交推送。

2026-10-08 原生full SVG光学0350 WebKit23389 2/2(17.864s)通过：DPR1正文/字形0，光学max2/255且全区域总差275～314，仅轮廓零星差；DPR1.25全RGBA0/native fallback，深浅/材质/odd width/活菜单及按钮身份通过。之前0345 cropped viewBox nativeSVG max23失败，保留且不采用。新native-full保留原生SVG原完整物理尺寸/坐标，不使用viewBox、不进行PNG二次alpha转换；SVG仅空div原CSS shadow/contour，没有背景/内容/字段。0355 Firefox与类型复核运行，bar-shadow-native仅显式真实diag；下一fresh90秒bounded真实对照，不以光学通过等同延迟收益。产品source99a20129/303bdb97仍52ms，R1开放，Vite29529，无压力/fixture；Windows不重复/发行延期/未提交推送。

2026-10-08 bar-shadow-cache0340 Firefox90秒bounded50ms（2770事件、initial55/steady50、5/4838长帧、七目的校验、双PTY约117KiB/s），93277退出0/29188清理0，显式prototype不是成品验收；产品仍52ms/R1开放。修正settle前置后0337 Firefox33594两缩放2/2；0335 WebKit25908 DPR1.25 native回退全RGBA0，DPR1深色光学max3/255但mean.113超过.1而FAIL，未提高阈值。新native-svg仅同内容无关光学document、native1:1裁剪，避免PNG二次alpha转换；其对照仍原内容RGBA0/原光学边界，未采用。下一两引擎验证native-svg，同时准备原native影patch合并的结构候选；全资源/字形/恢复/50ms不变。Vite29529，无fixture/压力浏览器；Windows不重复/发行延期/未提交推送。

2026-10-08 no-window-shadow独立0319 Firefox90秒bounded48ms（2895事件、initial58/steady48、3/5067长帧、八目的校验、双PTY约117KiB/s），95911退出0/10605清理0；这是改外观诊断，R1不关闭。完整产品仍99a20129/303bdb97的52ms。新增bar-shadow-diagnostic仅仅读数值轮廓/原生CSS阴影，isolated无内容div原生decode→PNG，无app DOM/文字/字段捕获；DPR1/同边框统一圆角才使用，失败/小数DPR/native none回退。86882类型0。0330 FirefoxDPR1光学最大4/255、正文0通过，DPR1.25 native回退live-app截图前window cache未settle导致失败；保留失败，修复仅增加实际材质/window影缓存与两RAF稳定前置，不放宽像素断言。0335 WebKit25908/0337 Firefox回归运行，原型未采用；下一通过后fresh真实90秒bounded bar-shadow-cache对照。所有原应用/字形/阴影exposure零RGBA守卫不变，Vite29529，无压力fixture；Windows不重复/发行延期/未提交推送。

2026-10-08 no-chrome-shadow固定顶部栏/Dock纯诊断0315 Firefox90秒bounded50ms（2745事件、initial52/steady50、8/4814长帧、七目的校验、双PTY约117KiB/s，原窗口64patch/27完全遮挡保持），52286退出0/18675清理0。不算成品PASS，Dock光学PNG因移除filter变641×74，不能只归因顶部栏。下一fresh no-window-shadow对照，同时准备仅固定顶部栏native outer/inset shadow的内容无关单层cache原型，保留实际背景/文字/图标/按钮/菜单活跃，不custom捕获应用/用户字段。原阴影/shape/颜色和像素守卫保持，新原型未采用/未验证。源99a20129/bundle303bdb97，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-nearest-all光学对照：Firefox37517 2/2(16.789s)、WebKit17804 2/2(23.222s)、13338类型0，仍native auto full reference RGBA0。但真实0310 Firefox90秒bounded54ms FAIL（2645事件、initial59/steady53、5/4596长帧、七目的校验、双PTY约117KiB/s），55703退出1/44986清理0，不采用任何point采样CSS。下一进一步拆分no-window-shadow与no-chrome-shadow固定顶部栏/Dock诊断；原no-shadow37ms同时影响二者且DockPNG尺寸改变，不能仅归因窗口PNG。保留完整52ms原生手势产品及所有原生应用/资源/像素/恢复标准。源99a20129/bundle303bdb97，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-nearest四边缘显式诊断0304 Firefox90秒bounded52ms FAIL（2720事件、initial60/steady51、7/4744长帧、八目的校验、双PTY约117KiB/s），76207退出1/14912清理0；与完整52ms无可靠收益，不进产品。下一仅测试corner也按DPR1 native1:1源point采样的光学控制，完整reference仍auto，DPR1.25全保留auto；新增shadow-nearest-all显式枚举/实际采样数断言，不算产品PASS。0307 Firefox/WebKit两项像素守卫与类型检查运行，source99a20129/bundle303bdb97不变，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-nearest WebKit66159 2/2(17.307s)通过，DPR1原生point采样对原auto reference RGBA0、DPR1.25 auto fallback保持。首次0303命令凭据路径误填，90356退出1/无JSON/未运行测试，保留独立启动日志，不算产品性能失败。实际14912 freshfixture-1100293144尚未使用，0304更正实际路径，独立唯一Firefox90秒bounded shadow-nearest显式诊断开始；必须32边缘采样实际生效，原负载/50ms/目的校验不变。源99a20129/bundle303bdb97未变，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native gesture正式0255 Firefox90秒bounded52ms FAIL（2670事件、initial60/steady52、6/4644长帧、七目的校验、双PTY约117KiB/s），31347退出1/76148清理0，未达50ms。shadow-nearest仅光学诊断开启DPR1四stretch边缘原生point采样，完整reference强制auto平滑采样，零RGBA阈值未改；Firefox60402 2/2(10.676s)，DPR1实际pixelated、DPR1.25原生auto、held/release/focus/fallback及全画面RGBA0通过；74231类型0。下一WebKit相同两项检查，再fresh真实90秒bounded shadow-nearest对照；不以诊断PASS关闭E08、不改应用/字形。source99a20129/bundle303bdb97，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native gesture正式构建56785退出0、Firefox49878 15/15(58.421s)通过；新moving局部所有权保持原生阴影elevation、RGBA0、held-menu期间几何、capture loss、八resize及即时reload。当前537源99a201290317900b5a6dc61340ab4dc9ccd179649f1058033494b08c3fdfdc73、168bundle303bdb97030728e3e031e449406d311bc18de9314aa67015fd5898034d38a051；76148 freshfixture-1302676915、唯一0255 Firefox90秒bounded正式负载开始。下一阴影stretch-axis原生采样仅候选：每种normal/focused原PNG沿stretch轴逐像素一致才可能改采样；DPR/坐标/原RGBA必须验证，不用近似掩盖正文/字形问题。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 no-shadow装饰隔离0247 Firefox90秒bounded37ms（3145事件、initial44/steady36、0/5493长帧、五目的校验、双PTY约117KiB/s），60142退出0/35074清理0；移除阴影/滤镜也改变Dock尺寸，不算产品PASS，不能只归因窗口八块阴影。与no-content41ms分别保存，不组合为完整验收。下一正式产品将gesture moving显示状态改为非响应式局部所有权，立即维护同一native frame的moving class；其它Vue更新仍读局部geometry/moving，结束/取消/reload保护不改。56785正式构建0、0251 Firefox原生阴影/零像素/held-menu/恢复等15项回归运行中。所有正文/字形仍原生，无custom应用数据采样/捕获。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 no-content纯诊断0243 Firefox90秒bounded41ms（3095事件、initial42/steady41、1/5416长帧、七目的校验、双PTY约118KiB/s），26807退出0/63828清理0；仅原生visibility隐藏应用绘制，真实资源/解析/保护/ACK保持，但外观改变且原生终端可停止画图，明确不算E08产品PASS。与同源完整54ms对照说明应用原生绘制存在重要成本；不删除内容、不减输出。下一fresh no-shadow装饰隔离诊断及原生手势显示状态所有权候选。源34a76f7b/产物1aaef90b未改，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 原生shadow可见决策/全正文包络去重正式0240 Firefox90秒bounded54ms FAIL（2695事件、initial63/steady53、7/4690长帧、七目的校验、双PTY约117KiB/s），71694退出1/27870清理0。两引擎各14/14和六单测通过，不等于延迟改善；source34a76f7b/bundle1aaef90b保持。下一fresh只诊断no-content及no-shadow隔离原生应用绘制与装饰成本，保持全部真实资源/输出/目录/传输/采样/50ms，但外观对照明确不算产品验收，不据此删内容或阴影。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 新shadow可见决策/全正文包络计算去重：构建32205退出0、六遮挡单测59351退出0；Firefox17100 14/14(54.981s)及WebKit31220 14/14(69.008s)既有零RGBA像素/partial完整patch/正文strips/held reveal/激活/纯translation/八resize/即时reload/Escape/capture loss/离屏/真实terminal遮挡保护通过。明确web相对路径指纹537源34a76f7b00a0875328bbc7d24678db081df2c7b6c6ed22898a15915457fde86f，168bundle1aaef90b6cdb52e8f2e763c2653d9ccd5b8f008c7f261c834490aa1134bf1981；实际当前候选，不混入旧源。27870 freshfixture-1670858593、唯一0240正式Firefox90秒bounded负载开始，无CSS/renderer/光学诊断flags。Vite29529；R1仍开放/Windows不重复/发行延期/未提交推送。

2026-10-08 正式geometry去重版0232 Firefox90秒bounded53ms FAIL（2695事件、initial61/steady51、11/4722长帧、七次目的校验、双PTY约117KiB/s），49364退出1/58365清理0。不将51ms稳态或旧源60秒PASS当最终达标。当前指纹98cb5188/61f38fba使用明确新方法；旧17e8c75c的聚合方式尚未复核，不能宣称已证明只是路径序列差异，历史标识保留但不混算。下一产品改动仅消除冗余遮挡计算：拖动中已露出的完整原生shadow patch保留可见布尔值，未露出的继续逐次检查；正文保守包络已覆盖全矩形时不再重复减去覆盖物，隐藏/部分正文仍计算，结束手势重新收紧。exposedBounds改单次边界归约，保留128片段保守回退。必须通过原零像素/held reveal/八方向/恢复及同源实负载，未声称性能收益。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 单window原生尺寸atlas对照0225 Firefox90秒bounded54ms FAIL（2670事件、initial59/steady53、12/4697长帧、八目的校验、双PTY约117KiB/s），26744退出1/35048清理0，不采用纹理放大。发现desktop.geometry即使完成手势返回同一持久坐标仍提交新rect；产品现四维全部相同时保留原对象/原保护回执，真实变化仍同步提交，不改变任何远程作用或数据恢复。33842正式类型构建0；Firefox17592八个既有真实移动/保持拖动下一paint/reload/Escape/capture loss/offscreen/八方向resize回归8/8、26.644s。新指纹明确采用web相对路径+NUL+文件rawSHA256再聚合：537源98cb5188d6dc36587e5a53277c5f408cb7c54175f735695ccb52c551a07eb199，168产物61f38fba43880fda961d5ad562fa1cb51d09ff2a9cb721da2b5a46c322c7e824；旧17e8c75c记录保留；其聚合方法尚未重建，不宣称仅路径序列不同，不混算。下一fresh正式90秒bounded Firefox0232，无材质/renderer诊断；Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native-scale0222诊断入口12.351s失败，无性能采样/无最终目的校验，不算产品延迟结果。新光学atlas已decode但非必要fetch原始blob的元数据读取触发NetworkError；实际CSP img-src允许blob、connect-src仅self/wss，不放宽策略。删除该fetch，只读本地新Blob PNG头；98177退出1/44778清理0。35048 freshfixture-1769381978、独立唯一0225真实Firefox90秒bounded重跑中；source17e8c75c/bundlea402d394未变，Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 缓存shadow-overflow WebKit65406 2/2、16.971s通过，但真实Firefox90秒bounded54ms FAIL（2620事件、initial/steady均54、8/4593长帧、七目的校验、双PTY约112KiB/s；实际21视口、减少311977CSS pixels），14853退出1、64905清理0，不进入产品。下一只将window一个共享atlas预先native-scale重采样；其余三滤色atlas不扩大，不生成逐窗口数据/文字纹理，原模糊/配色/几何/阴影/负载与50ms不改。window-material-diagnostic.ts仅读取光学blob、原生canvas采样，记录尺寸/PNGcolourType数值；native-scale为显式诊断，不算产品验收。1414类型0、44778 freshfixture-3291603599与Docker真实90秒bounded运行，唯一0222报告。source17e8c75c/bundlea402d394、Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-crop原型首版Firefox18278 2/2、10.229s及WebKit57412 2/2、16.841s通过：DPR1实际partial viewports>0，DPR1.25原生fallback无viewports，重叠/held reveal/复位/focus/过小native fallback全部RGBA0/full-reference保持。随后仅诊断helper加入缓存native geometry、纯translation算术及root/size/focus/cache代次失效；self-contained installer通过原生browser protocol可作用于正式bundle。50231/74177类型0，缓存版Firefox43726 2/2、10.041s；WebKit65406复核中。正式产品source17e8c75c/bundlea402d394未改，不将纯helper当产品或性能PASS。下一WebKit缓存版像素通过后fresh真实Firefox90秒bounded shadow-overflow，保留8窗/双PTY/10k目录/16MiB目的校验及50ms。Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native-dom-equality确实复用111154/117664行、仅6510次替换，但Firefox90秒bounded仍53ms FAIL（2695事件、initial61/steady52、8/4685长帧、七目的校验、双PTY约117KiB/s），70178退出1、75793清理0，不作为产品改动。下一content-free原生shadow overflow viewport，在tests/helpers/shadow-crop.ts保持每块完整原图quad/UV/颜色/阴影，仅单矩形overflow裁掉保守不可见区域；不复制应用内容、不用新增向量mask，DPR1以外保留原生完整patch。新增显式BLORA_E08_SHADOW_CROP_CONTROL，原有DPR1/1.25零差异/held reveal/focus/fallback断言不放宽，独立完整native reference暂停裁剪；42479类型0，Firefox18278功能像素检查中。source17e8c75c/bundlea402d394不变，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 official native-webgl-software对照结束53ms/2695事件/八目的校验，但两个真实renderer仍default，说明该Firefox环境没有成功装载WebGL，不是WebGL与DOM性能对照，不据此改产品策略。72909退出1、25320清理0。下一native-dom-equality显式诊断：原生RowFactory仍生成完整行，只在原生isEqualNode确认全部子节点/属性/正文严格相同时跳过replaceChildren；光标blink行始终原替换，不缓存/光栅化应用文字。实际复用次数必须>0，原负载/50ms保持。51851类型0，75793 freshfixture-3938814211、Docker70178运行90秒bounded、唯一0153报告；source17e8c75c/bundlea402d394、Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 Firefox原生Gecko诊断完成：0140真实90秒bounded59ms（采样器有开销，不作验收），2495事件/32长帧/八目的校验；19271退出1、95859清理0。私有profile70,306,168字节，本地仅静态/数值汇总：诊断时间区间内content RefreshDriverTick p95 13.86ms、Styles1.81ms，JS采样中xterm原生renderRows312/createRow296，另有软件Renderer图形开销。两个真实终端实际default，所有3个terminal host未完全遮挡，不借完全遮挡降低负载。下一原生官方WebGL软件设备对照；只显式测试初始化探针允许software，不改实际addon context/字形/解析/保护/ACK，注解明确native-webgl-software。类型22959通过；25320 fixture-3986027524与Docker72909真实90秒bounded运行，报告唯一0150。source17e8c75c/产物a402d394仍不变，Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 wallpaper-only material-nearest Firefox90秒bounded53ms FAIL（2695事件、initial60/steady52、9/4696长帧、八目的校验、双PTY约117KiB/s），96123退出1、96448 fixture-2654306122清理0；body-viewport53803清理0确认。采样改变无收益，不进入产品。下一Firefox原生Gecko主线程/合成/渲染诊断，startup采样只用于归因、不算验收。0139启动因凭据路径误抄退出1/无JSON/零用例，保留日志且不计产品失败；0140更正实际freshfixture-2304037819，95859及Docker19271运行，私有profile仅读取静态/数值汇总不输出用户状态。当前源17e8c75c/产物a402d394、Vite29529；R1仍开放/Windows不重复/发行延期/未提交推送。

2026-10-08 body-viewport结构原型Firefox90秒bounded52ms FAIL（2695事件、initial58/steady51、6/4707长帧、七目的校验、双PTY约117KiB/s），21823退出1、53803 fixture-503709807末轮校验后请求清理；不作为成品/验收，native DOM移动原型未经独立像素/多视图生命周期验证，源码仍17e8c75c/a402d394。下一wallpaper-only image sampling诊断，CSS仅光学伪元素使用pixelated采样，源模糊/tint/不透明度/窗口内容/原几何/恢复/阴影/50ms保持。用户允许wallpaper-only光学近似，不捕获应用；如果有收益仍须独立采样误差与完整native0正文/字形/三引擎/持续负载验证。Vite29529、R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native body viewport纯test结构原型在web/tests/real/body-viewport-diagnostic.ts：只移动实际应用DOM、不复制数据/捕获应用纹理，原尺寸通过完整content层保持，原壁纸光学material沿用；外层原生overflow按原保守bbox裁剪替代body clip-path。仅显式body-viewport、真实8窗创建完后安装，不进入产品。55204含全部test的类型检查0；当前53803 freshfixture-503709807、90秒bounded Firefox真实诊断开始。91933上轮清理0已确认。若收益有效，必须先独立native0像素/尺寸/滚动/身份/恢复/移窗验证，再在Vue原生模板接入，不把手动VueDOM原型当成成品。源17e8c75c/产物a402d394、Vite29529；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 shadow-plane-layers Firefox90秒bounded53ms FAIL（2670事件、initial61/steady52、7/4670长帧、七目的校验、双PTY约117KiB/s，实际native shadowPlanes8），17871退出1，91933 fixture-2336063514末轮验证后请求清理。上轮chrome85364退出1/67705清理0确认。所有新will-change/3D候选无收益，不进入产品。下一结构原型native body paint viewport：原应用完整尺寸/DOM/身份和材质保持，原生overflow可见区域替代合成clip-path，先纯测试helper90秒bounded诊断；成功才进入源码并严格核对全部native像素/恢复/移窗/大小/三引擎及长测。source17e8c75c/bundlea402d394、Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 chrome-layers Firefox90秒bounded53ms FAIL（2670事件、initial59/steady53、5/4644长帧、七目的校验、双PTY约117KiB/s），不加入产品；85364运行末轮退出检查，67705 fixture-362606154清理已请求。下一fresh shadow-plane-layers仅缓存原native八块装饰阴影为浏览器合成层，不custom捕获应用/改变视觉/关闭负载/50ms，source17e8c75c/bundlea402d394不变；Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 no-body-layer Firefox90秒bounded53ms FAIL（2695事件、initial61/steady53、5/4715长帧、八目的校验、双PTY约117KiB/s、最大未ACK66264），89201退出1、49083 fixture-3662299258末轮验证后请求清理；不采用，无明显收益。前28770夹具清理退出0已确认。下一fresh chrome-layers90秒bounded仅原生固定topbar/taskbar绘制层，内容/材质/负载/50ms保持；source17e8c75c/bundlea402d394、Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 full-native-body-paint Firefox90秒bounded83ms FAIL（1945事件、initial89/steady82、352/3171长帧、五目的校验、双PTY约114KiB/s、最大未ACK86318），68604退出1、28770 fixture-1137384118末轮验证后请求清理；不采用，保留产品正文保守bbox裁剪。高成本来自更多原生应用绘制，不能用删应用来回避。下一fresh90秒bounded no-body-layer只改变body::before的冗余will-change hint、现有源/像素/负载/恢复/50ms不改；source17e8c75c/bundlea402d394、Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 native-window-surface原生3D路径Firefox90秒bounded62ms FAIL（2420事件、initial69/steady61、18/4152长帧、七目的校验、双PTY约117KiB/s），39011退出1，4303 fixture-801381765末轮验证后清理0，不进入产品。该report名与较早13:27:23的旧60秒71ms/1464事件失败发生碰撞；当前原JSON已被覆盖，仅保留从已读取结果提取的旧数值，不伪称旧raw完整保留，当前完整JSON另以实际17:02:42时间命名保存；后续所有新artifact先核对不存在且带时间标签。下一fresh90秒bounded full-native-body-paint，只恢复部分正文完整原生绘制，完全覆盖正文/terminal仍省画；现有整屏native full-reference已覆盖无mask绘制外观，不删图形/数据/负载/50ms。Vite29529、源17e8c75c/产物a402d394，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 active-app-layer Firefox90秒bounded诊断56ms FAIL（2620事件、initial68/steady54、11/4574长帧、五目的校验、双PTY约116KiB/s）；45021退出1，54370 fixture-362993294末轮验证后清理0，不加入产品。上轮60238夹具清理也确认0。当前源/构建仍17e8c75c/a402d394，继续隔离原生3D合成路径及冗余body光学layer成本，先native-window-surface相同90秒bounded对照；所有负载/50ms/数据合同保持，拒绝diagnostic不算产品验收。Vite29529运行，无压力browser/fixture；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08同源17e8c75c/a402d394 Firefox300秒bounded较强持续负载53ms FAIL（8370事件含initial120/62、steady8250/52、30/15028长帧、21目的校验、双PTY约116KiB/s、最大未ACK81312）。66265退出1，末轮校验已完成；60238 fixture-3353403589 Ctrl+C清理已请求。原60秒同源三引擎30.9/49/49及600秒Chromium31.1 PASS仍保留，但不足以收尾，不隐藏bounded首分钟54/次分钟53等持续开销。下一诊断active-app-layer：仅活动window-body保留原生浏览器绘制层（没有custom应用采样/删阴影/减負載），原代码不变，fresh Firefox90秒同样bounded作对照；成功才进入产品并重跑同源证据。Vite29529，R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08同源17e8c75c/a402d394 Chromium600秒持续负载PASS：19495native事件p9531.1ms、2/36164长帧、48次目的校验、双PTY约116KiB/s且逐分钟均持续、最大未ACK60544。775次原生CDP heap数字25.13–71.90MB、first52.44/last47.62MB（未强制GC/无内容快照），九分钟记录p9530.2；档案轮转终值15,767,440/16,614,832B均≤16MiB，额外较早三点也在限内；容器聚合内存采样1.102/1.159GiB（不是JS heap或完整峰值）。86320退出0，46976 fixture-2142794047末轮校验后清理0。该bounded长测不替代原60秒三引擎轨迹，只加强持续性。下一fresh Firefox/WebKit各较长负载、最终三引擎功能和单位/类型/文档检查；R1尚不关闭，Vite29529；Windows不重复/发行延期/未提交推送。

2026-10-08正式源码/产物17e8c75c/a402d394不再改动，开始600秒Chromium持续负载：fresh46976 fixture-2142794047，原数据/输出速率/校验/50ms保持；仅额外长测用bounded归原位手势避免离开屏幕减负载，不替代原60秒轨迹PASS。仅长测显式PRECISE_HEAP=1取Runtime.getHeapUsage数字（不读快照/内容/不强制GC），区分原performance.memory舍入值，记录first/last/min/max。Vite29529运行；后续还需同源Firefox/WebKit较长稳定性、最终功能/单位/链接检查；R1尚不关闭/Windows不重复/发行延期/未提交推送。

2026-10-08同源17e8c75c/a402d394原60秒三引擎全部达到≤50ms：Chromium30.9/Firefox49/WebKit49。Chromium96865退出0，2184样本含initial120/32.4、steady2064/30.8、0/3741长帧、五目的校验、双PTY约118KiB/s、最大未ACK86774；55165 fixture-29800656清理0。未降低负载/放宽阈值/排除初始样本/混源/启用diag。当前下一动作同源较长持续负载及最终三引擎功能，之后才关闭R1；Vite29529保留，无压力browser/fixture；Windows不重复、发行延期、未提交推送。

2026-10-08同源17e8c75c/a402d394正式原60秒Firefox49ms PASS（1644事件含initial120/64、steady1524/42，3/3681长帧、五目的校验、双PTY约118KiB/s、最大未ACK62168）。46351退出0、24060 fixture-3189678340清理0。同源正式WebKit49/Firefox49已通过，但Chromium/最终功能/长测仍缺，不关闭R1；Vite29529运行，无压力browser/fixture；不请求Windows/发行延期/未提交推送。

2026-10-08当前正式source17e8c75c/bundlea402d394原60秒WebKit49ms PASS：4032事件含initial120/88ms、steady3912/43ms，9/3846长帧，五目的校验、双PTY约117KiB/s、最大未ACK86010、native list masks0/partialshadow0。17261退出0、5682 fixture-2085301826末轮验证后清理退出0。未启用任何诊断style/prototype，无新预热/删初始/降负载/松50ms。下一同一源/产物fresh Firefox，再Chromium/最终功能/较长稳定性；三引擎未齐仍不关闭R1。Vite29529保留，无压力browser/fixture；Windows不重复/发行延期/未提交推送。

2026-10-08撤回文字/list裁剪正式候选：正常与显式prototype两DPR全屏RGBA0/滚动/筛选替换/held露出/focus/fallback分别2/2（67077 21.6秒；2007 20.4秒）通过，42144类型/生产构建0。source537文件SHA256 17e8c75c50bb7480735b9fec09bb375117a1626329460bec5dca5e4bec33d6d4，bundle168文件a402d394814416246ff6c533d93f52a4533208e870d30d94ce05a86b856500c7。当前fresh5682 fixture-2085301826原60秒WebKit正式baseline进行（无diagstyle/prototype），末轮目的验证后退出browser、夹具Ctrl+C确认0再启下引擎。Vite29529运行；同源三引擎/长测缺、R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08 full-native-list-paint WebKit原60秒诊断50ms PASS（4032事件、initial113/steady43、9/3832长帧、五目的校验、双PTY约118KiB/s），21416退出0、73658 fixture-791398994清理0。该诊断未算产品验收。正式代码现在撤回无足够收益的文字/list逐块clip与注册/测量/监听；原生正文、原窗口/终端遮挡、数据/ACK/恢复、负载/阈值保持。拒绝prototype移到tests/helpers/native-list.ts，仅显式BLORA_E08_NATIVE_LIST_CONTROL=1；保留全部positive存在及两DPR全屏RGBA0/滚动/替换/held reveal/focus/fallback守卫，普通产品额外断言不存在逐块mask。下一具体动作正常与显式prototype守卫、构建、新同源原负载WebKit，再Firefox/Chromium及长测。Vite29529运行，无压力browser/fixture；R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08稳定shadow正式WebKit52ms FAIL（4008事件、initial120/steady42、12/3822长帧、五目的校验、双PTY约117KiB/s），45989退出1、64821夹具fixture-121011267清理0。当前源05711dbc/产物d3f32c08不能称达标，不混入旧49/48。慢样本input约0–16ms但first/second RAF有125/129ms等待，存储保护write max3ms、snapshotclone p95 2ms，继续定位冷绘制而非删初始样本/加等待。新full-native-list-paint仅不裁既有完全遮住的文字/list块，保持原DOM/遮挡/全部真实负载与50ms；既有零差异reference已覆盖完整原生文字，属于隔离诊断、不能算产品验收。fresh73658 fixture-791398994当前原60秒WebKit对照运行，Vite29529；最终同源三引擎/长测缺、R1开放/Windows不重复/发行延期/未提交推送。

2026-10-08当前稳定native单patch版：正常与显式partialprototype两DPR严格全参考/held/focus/fallback各2/2通过（61985 18.7秒；16937 17.2秒），全部新增守卫保留，测试prototype没有进入产品。生产32795已构建，开始冻结该源/产物用于原完整同源引擎实测：source537文件SHA256 05711dbc1a4f6797c1efac400f2ae60f20555f035a0904d1625f0a536a1627c3，bundle168文件 d3f32c087a7c09ac6732933114fa9c55fe3b2f4fd097751fa692b4b1f7a53f14。当前fresh64821 fixture-121011267原60秒WebKit baseline运行（未启用prototype/诊断style），报告web/.local/e08-stable-native-shadow-webkit-report.json，Vite29529保留。之后须末轮目标验证+清理0、同源Firefox/Chromium、当前功能和较长稳定性；E08/R1未关闭、Windows不重复、发行延期、未提交推送。

2026-10-08（本地日期）500ms候选原60秒WebKit55ms FAIL（3960事件、initial116/steady45、15/3818长帧、五目的校验、双PTY约117KiB/s），88987退出1、61666夹具fixture-1229498487清理0。产品partial shadow优化和仅光学空闲timer现撤回，保留全部隐蔽patch省画及原native单patch/image UV；不采样应用/删阴影/改负载/阈值。拒绝的native rect/fractional inset partialprototype移到tests/helpers/partial-shadow.ts，只显式BLORA_E08_PARTIAL_SHADOW_CONTROL=1执行；原先存在性断言在control仍要求positivepartial、DPR1 rect/fractional path，各whole-screenRGBA0/held无partial/再曝光/focus/fallback守卫完整保留，普通产品额外要求partial不存在。61985正常产品两DPR共2/2（18.7秒）RGBA0，32795生产构建通过。显式prototype守卫正在验证，freshfixture准备，Vite29529；本次尚无正式三引擎/长测，R1/50ms开放/Windows不重复/发行延期/未提交推送。

2026-10-07原180ms空闲恢复产品正式WebKit68ms FAIL（3984事件、initial126/steady60、28/3776长帧、五目的校验、双PTY约117KiB/s），80473退出1、47014夹具fixture-2757535965清理0。原压力轨迹每轮有250ms原等待，180ms会在短输入间隙重建，产品改仅500ms安静后回收部分光学省画，测试250ms/指针/负载/50ms保持不变。95451首次DPR1 held screenshot165通道max200，附件定位仅顶栏时钟bbox(1252,36)-(1260,45)跨分钟；只该像素test固定new Date()显示值、Date.now/事件/RAF保持原生，46897两DPR整屏RGBA0/held无partial/quiet恢复/focus/fallback共2/2（19.2秒）通过、失败证据不删。71705最新生产构建通过。fresh61666 fixture-1229498487当前500ms候选原60秒WebKit实测，Vite29529；最终同源三引擎/长测未齐，R1/50ms开放/Windows不重复/发行延期/未提交推送。若仍无收益，撤回产品动态partial，仅保留原全隐蔽省画及partial原生prototype守卫。

2026-10-07动态native rect正式WebKit76ms FAIL（4032事件、initial113/steady69、52/3720长帧、五目的校验、双PTY约117KiB/s），12222退出1、29574夹具fixture-1509349156清理0。并非只有clip-path语法成本，动态部分裁剪合成区域重建无收益。新产品按输入生命周期稳定背景原生阴影区域：移动期间保留原完整部分patch，只裁全隐蔽，结束180ms空闲再恢复部分省画；新拖动取消仅此光学空闲timer，数据/交互/存储/解析/ACK不延迟，完全不采样应用内容。idle DPR1 native rect、fractional原path，所有native原全绘参考/held清除partial/空闲恢复/focus/fallback43414共2/2（25.1秒）RGBA0通过，50958生产构建通过。fresh47014准备原60秒WebKit，Vite29529；最终同源三引擎/长测未齐，50ms/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07定位WebKit回退：full-native-shadow-patches只停部分clip-path原全负载48ms（4008事件、initial90/steady42、8/3812长帧、五校验、双PTY约117KiB/s），37408退出0、23920夹具fixture-1808069879清理0；诊断不关闭E08，但69→48表明部分通用mask成本。新产品部分shadow在DPR1用原生CSS rect clipping，原full image/UV不变，无填充/新图层/应用采样，fullhidden仍原裁剪；fractional DPR保留已验证clip-path（首次统一rect在1.25为419通道max1失败，完整证据保留）。新增明确DPR1原生rect存在、1.25旧path回退，原完整绘制参考同时移除两种clip，整屏RGBA0/held/focus/fallback77858共2/2（18.2秒）通过，不删/放宽原守卫；32247类型/生产构建通过。下一动作当前新rect同源fresh原60秒WebKit，再Firefox/Chromium与功能/长测；当前无压力browser、Vite29529；R1/50ms开放、Windows不重复、发行延期、未提交推送。

2026-10-07native-dock-filter完整WebKit对照72ms FAIL（4032事件、initial116/steady65、45/3732长帧、五目的校验、双PTY约117KiB/s），10891退出1、11646夹具fixture-3692852158清理0，不采用任何WebKit专用Dock/UA分支。恢复原Dock滤镜并未消除回退，下一动作原60秒full-native-shadow-patches对照：只恢复部分patch完整原生绘制，完全隐蔽patch仍裁剪；已严native0像素参考验证该完整绘制外观一致，不删除阴影/业务/负载/50ms。当前源码与4322生产dist一致（仅test新诊断字段），Vite29529；R1/同源三引擎/长测开放、Windows不重复、发行延期、未提交推送。

2026-10-07Dock display:none正式WebKit仍69ms FAIL（4080事件、initial108/steady64、35/3751长帧、五目的校验、双PTY约118KiB/s），59064退出1、98848夹具fixture-1619557135清理确认退出0。该优化不解释引擎差异，当前生产4322完整新Dock+原小atlas。新增纯测试native-dock-filter对照恢复原nativeSVG/drop-shadow并隐藏decoded光学img（不改业务/输入/负载/阈值，不删除阴影）；另一full-native-shadow-patches对照仅不做部分patch clip、全遮挡保留。fresh11646 fixture-3692852158，当前原60秒WebKit native-dock-filter对照运行；最终原baseline/同源三引擎及长测尚缺，不以诊断计产品PASS。Vite29529，50ms/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07恢复小atlas正式WebKit71ms FAIL（4032事件、initial112/steady67、42/3755长帧、五目的校验、双PTY约117KiB/s），不是atlas主因；46246退出1、85592夹具fixture-3794491670清理退出0。新Dock发布后用display:none真正移除原SVG绘制层（此前visibility:hidden），解码/失效仍原生SVG回退。新增held异步解码/实际故障计数/旧URL撤销/恢复/真实按钮操作，以及DPR1/1.25六种光学对照WebKit34689共3/3（28.5秒）通过，最大通道2/255、fractional回退RGBA0，未放宽正文/字形/遮挡守卫。4322生产构建通过；freshfixture原60秒WebKit准备。Vite29529运行，无别的压力browser；最终同源三引擎与长测缺，50ms/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07同源新Dock+display-resolution atlas正式WebKit69ms FAIL（4056事件、initial98/steady66、38/3759长帧、五目的校验、双PTY约118KiB/s），cache实际decoded692×127；6867退出1、68857夹具fixture-1213685340清理退出0。Firefox49只单引擎PASS，不关闭R1。高分辨率四材质atlas此前Firefox52无收益且增加4倍像素，现只撤回这项产品pre-expand、恢复原半分辨率wallpaper-only滤波缓存；所有新测试/失败证据/Dock native缓存保留。构建进行，下一freshfixture原60秒WebKit；当前无压力browser/fixture，Vite29529；最终同源三引擎/长测未齐，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07内容无关Dock原生缓存正式原60秒Firefox49ms PASS（1644事件、initial59/steady42、2/3686长帧、五目的校验、双PTY约118KiB/s、最大未ACK86010）；cache实际ready/decoded、681×116px。42568退出0、46087夹具fixture-1696061879清理退出0。生产93875构建3746模块2.11秒通过，所有当前产品源码未再改。新Dock只native face/path/颜色/边缘/drop-shadow，不采样图标/应用/文字；DPR1包含worldphase，原SVG失败/其他DPR回退、仅裁alpha=0、单URL/代际/卸载清理。独立WebKit两display guards10477共2/2通过（裁零透明边之前、最大通道2/255、均值≤0.1/255），Firefox裁边后96886共2/2（15.0秒、最大1/255）；DPR1.25原生回退六对照均RGBA0。严格光学RGBA0负对照失败保留，不宣称shadowPNG绝对像素一致，正文/字形/遮挡原0容差守卫未放宽。当前无压力browser/fixture，Vite29529；下一动作同一生产dist freshfixture原60秒Chromium/WebKit，再当前三引擎功能守卫和较长稳定性。尚不能关闭50ms/R1；Windows不重复、发行延期、未提交推送。

2026-10-07 Dock滤镜隔离结束：仅去除Dock SVG滤镜的原60秒Firefox50ms（1632事件、initial64/steady43、4/3671长帧、五目的校验、双PTY约118KiB/s）属于改变视觉的diagnostic，不能关闭E08；52321退出0、86307夹具fixture-3986061560清理退出0。当前无压力browser/fixture，Vite29529继续。新内容无关Dock face+shadow原生PNG缓存正在验证，禁止采样应用DOM/图标/用户数据；DPR1含原世界坐标小数phase，其他DPR原生SVG回退，单current/pending URL与代际归属，尺寸/外观重建。初次WebKit对照DPR1最大6通道级差异，strict零差异失败证据保留，DPR1.25原生回退三对照零差异但solid模式测试误等材质ready（产品关闭材质正常）已修正；下一动作比较foreignObject原生CSS filter栅格化，再评估视觉/生命周期及原全负载。生产dist72227还是旧材质候选，当前新源码尚未build，无正式同源三引擎/长测，50ms/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07显示分辨率wallpaper-only材质正式原60秒Firefox52ms FAIL（1620事件、initial64/steady45、3/3647长帧、五校验、双PTY约118KiB/s），末轮目的验证完成、28075夹具退出0；六原生/生命周期守卫51969共6/6（38.7秒）通过。没有足够收益，不据此提升验收；下一隔离只取消Dock SVG原生drop-shadow以定位残留滤镜成本，窗口阴影/背景材质/业务和原测量保持、明确 altered-visual diagnostic不计PASS。生产dist72227仍该native-resolution材质候选，暂无新压力browser；freshfixture准备，Vite29529保持。之后按证据决定是否做仅光学Dock脸/阴影缓存，不采样应用内容。最终同源三引擎与长测缺，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07材质原生分辨率新候选：保持既有半分辨率预过滤/颜色，把四种wallpaper-only atlas预采样为显示物理分辨率（DPR≤2、最长4096px），共享/完整decode后原子发布；不采样应用DOM/字体/Canvas/用户数据，不新增每窗纹理，不改恢复/业务。新有界列表曝光集合清理已移除DOM引用，避免长手势/替换保留无界旧对象。分段产品已完全撤回，复现helper只显式test控制加载，原native pixel参考保留。类型/生产72227构建通过（3745模块、2.78秒）；WebKit材质生命周期/host隔离/列表及阴影DPR1/1.25六项51969实测，下一动作收取全结果，再freshfixture原60秒Firefox实测。Vite29529运行；当前源三引擎/长测未齐、50ms/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07分段候选实测结束：原60秒Firefox55ms FAIL（1572事件、initial64/steady49、4/3577长帧、四校验、双PTY约117KiB/s），272段151隐蔽、约359004CSSpx省画，但图形数量开销反而更慢；末轮目的验证完成、91937夹具退出0。已撤回分段产品执行与CSS，prototype移到tests/helpers/shadow-segments.ts作为显式BLORA_E08_SHADOW_SEGMENTS_CONTROL原生诊断，新增严格原单patch参考/像素守卫全部保留、不删除失败。当前无压力browser/fixture，Vite29529；生产dist仍分段旧候选31122，下一实测前必须重新构建。下一动作按原生软件绘制证据继续材质采样/合成架构；原50ms/R1未达标，同源三引擎/长测仍缺、Windows不重复、发行延期、未提交推送。

2026-10-07新阴影分段候选：长边保留原full原生纹理quad及UV，按64物理像素对齐小rect仅裁内部边界，被遮段隐藏。首次DPR1.25直接局部图片3200通道/最大3，保留full quad+overflow9288/12，保留outerAA后2927/2，失败完整留存。因此产品快路径明确限DPR1且原阴影几何全整数，其他显示比例/几何使用原decoded单patch；不牺牲像素。不降低已有整屏零差异断言，新增确认DPR1分段存在、DPR1.25原生回退无分段；对照参考更严格恢复原单patch而非仅取消裁剪。DPR1/1.25整屏0差异/held露出/focus/fallback2/2（80564，20.0秒），生产类型构建31122（3746模块，2.44秒）通过。新数字指标只量分段数量/CSS几何，不计GPU字节。Vite29529，下一步freshfixture原60秒Firefox评估该新架构，仍无同源三引擎与长测PASS。50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07阴影原生分辨率隔离结束：Firefox原60秒54ms FAIL（1608事件、initial66/steady46、2/3645长帧、五次校验、双PTY约117KiB/s），末轮目的校验完成、38840夹具退出0。没有收益，不采用诊断；生产仍有界列表+部分阴影rect。当前无压力browser/fixture，Vite29529运行。下一优化将研究长阴影边按物理像素对齐分段的原生图片绘制，只跳过完全遮挡的段、不新建大透明平面；必须先与原单patch整屏零差异比对，再压力实测。不采样应用内容、不改50ms/原八窗口/真实PTY/传输/恢复合同，三引擎与长测尚缺、R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07当前：有界原生文字/列表遮挡DPR1/1.25滚动/替换/held露出/focus/cache回退整屏零差异2/2（66727，20.4秒）通过；首次219通道变化位于topbar时钟跨分钟，固定显示日期时Date.now一起冻结会触发Vue捕获/冒泡时间戳保护，已改只固定new Date()显示值、Date.now/真实事件/RAF保持原生，保留失败证据。类型85011与生产39799（3745模块，2.48秒）通过。产品原60秒Firefox52ms FAIL（1608事件、initial62/steady46、4/3637长帧、四校验、双PTY约117KiB/s），结束75942夹具0。该负载只裁2块约1893CSSpx，无足够延迟收益，不宣称达标。当前唯一隔离诊断30038：既有内容无关RGBA阴影4条伸展边预放大到实际物理像素，测原60秒Firefox，fresh38840 fixture-36296260，报告web/.local/e08-shadow-native-resolution-firefox-report.json；不采样应用内容、不换原阴影颜色/elevation、不改产品/原测量阈值；候选视觉产品化仍须原生对照。Vite29529继续。下一动作收取目的校验/结果并结束夹具，再按证据推进，最终同源三引擎与长测缺，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07续接：direct-body-material原60秒Firefox56ms FAIL（1620事件、initial67/steady47、4/3632长帧、四次校验），末轮验证完成、45980夹具退出0，未采用。当前没有压力夹具/浏览器，Vite29529仍运行。新候选只对既有overflow有界默认应用文字/列表块做完全遮挡原生inset(100%)裁剪，每窗最多64块、几何缓存平移、手势一旦露出不再回裁；列表结构/尺寸/原生scroll/focus失效，不观察终端/Monaco内部，不采样应用内容、不停数据/布局/恢复/ACK。新增DPR1/1.25整屏零差异/滚动/列表替换/held露出/回退回归待执行；类型检查27418运行。下一动作先校验原生像素与现有守卫，再原60秒Firefox；当前源三引擎与长测仍缺，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07最新实测：部分原生阴影矩形裁剪产品原60秒Firefox52ms FAIL（1620事件、initial67/steady46、1/3641长帧、五次16MiB校验、双PTY约117KiB/s）；64块中27全裁/8部分裁，CSS区域800188/1628332约49%省画，仅几何数字不算GPU分配。末轮校验通过，29509夹具清理0；准确像素收益保持，但不宣称延迟达标。新隔离诊断将同一wallpaper-only正文材质改交body背景原生绘制，去掉独立负z pseudo表面；不采样应用内容/不改配色、形状、业务、恢复、解析或测量。生产尚未采用该诊断。当前唯一压力Firefox direct-body-material原60秒，fresh45980 fixture-602226898，报告web/.local/e08-direct-body-material-firefox-report.json；生产构建12987/Vite29529保持。下一动作收取实测/末轮校验并结束夹具，再依据收益校验产品化/原生像素；同源三引擎及长测未齐，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07当前产品候选：4阴影周边带完整原60秒Firefox64ms FAIL（1524样本、initial75/steady55、两校验循环、22371清理0），不采用；改为既有8原生shadow patch的保守部分矩形裁剪，缓存本地两elevation几何随纯平移更新，cache发布/回退由边界观察失效，手势曝光只扩不缩并提前32px。首次DPR1.25新增严格全屏对照出现528通道/最大5差异，定位为新rect clip在原native图片的fractional外边缘重复AA，裁剪每边再保守外扩1CSSpx后，DPR1/1.25零差异、held露出/focus/fallback2/2（28995，18.3秒）通过；原六项native轮廓/latepub/回退/布局/正文像素6870共6/6、27.3秒。所有原断言不变/失败证据保留，应用内容未采样。最新类型构建12987（3745模块，2.29秒）与数字面积类型检查92695通过。当前唯一正式baseline Firefox60秒实测，fresh29509 fixture-2426095182，报告web/.local/e08-partial-shadow-firefox-report.json；Vite29529。下一动作收取实测/目的校验并结束夹具，再按证据保留/继续架构优化，最终同源三引擎及长测仍缺，R1开放/50ms不变/Windows不重复/发行延期/未提交推送。

2026-10-07最新：focused独立三维零平移正式原60秒Firefox71ms FAIL（1464样本、35/3390长帧、initial86/steady64、四校验循环，93487夹具清理0），未采用；材质交界已重叠的精确条带重试仍DPR1/1.25全画面4407/9496通道变化、最大4/25，83630 exit1，产品已回退保守bbox，原两项零容忍回归与geometry helpers完整保留。官方Mozilla符号库该固定Playwright libxul标识404，本地ELF仅两导出函数，不据最近符号错认热点。新隔离诊断只合并已有原生阴影8图为4周边带，不读取应用内容；同样颜色/elevation、两状态完整解码，应用渲染/材质/输出与原60秒测量保持。类型83972 exit0。当前唯一压力shadow-bands Firefox原60秒运行，fresh22371 fixture-403990417，报告web/.local/e08-shadow-bands-firefox-report.json，Vite29529。生产未启用，最终同源三引擎/长测仍未齐，50ms/R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07原生首轮定位：零soak诊断已完成，原120指针85ms、4.484秒，不计完整验收；双PTY持续、最后16MiB目的校验通过，63352夹具清理exit0。私有54MB profile只解析数字与原生标签白名单。首轮实际4.462秒内Renderer CPU3891.8ms（ns正确换算ms），content GeckoMain2504ms；Renderer主要RenderThread::UpdateAndRender，原生软件绘制/合成成为直接证据，不能归因于Monaco冷启动。marker仅phase1且end>=start作为有效区间；早先未区分phase的错误聚合不用于结论。当前唯一原60秒Firefox native-window-surface诊断22043，fresh93487 fixture-3936916850，CSS独立translate三维零偏移/backface只focused，不改原2D几何/外观/内容。生产未采用，Vite29529，无其他压力浏览器。下一动作收取正式负载诊断/末轮校验，清理夹具，再继续原生合成架构优化；50ms不变/R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07当前接管：宿主材质登记隔离产品正式Firefox原60秒53ms FAIL（1632样本、3/3668长帧、五16MiB校验循环、双PTY约118KiB/s、初始120事件64ms/后续1512事件44ms，98835清理0）。隔离范围有真实减少无关查询收益但不宣称延迟达标；生产构建21582保持。原生Firefox profiler full60诊断56ms/初始69/后续48，末轮验证与25253清理0；raw私有文件web/.local/e08-host-material-gecko-profile.json67MB，不输出/上传，需读取数字/原生标签白名单。4M buffer只保留native时间39～82秒，初始11.62～15.434秒已滚出，不能用它证明初始根因。下一动作fresh63352 fixture准备，仅0soak原首120指针原生profiling诊断，双PTY/8窗口/万目录/实际传输仍齐但不计E08/60秒验收；唯一新profile文件名，原full60失败保留。Vite29529，最终当前源三引擎/功能/长测仍待，50ms/所有原断言不变、不等待editor readiness、R1开放/Windows不重复/发行延期/未提交推送。

2026-10-07最新接管：当前contained bundle全8window常驻层诊断Firefox55ms FAIL（1620样本、4/3647长帧、三校验循环、双PTY约116KiB/s）；初始120事件p9566ms、后续1500事件p9545ms，未加等待/删首轮样本，末轮校验和36474 fixture清理0，extra all-window hints不采用。新产品候选按真实host ownership隔离材质登记：观察desktop直接子级与topbar/taskbar分支，不再接收终端/Monaco整个树MutationRecords；新增动态账号/通知/launcher定位及80行×12轮更新不会触发host查询回归。Firefox新回归/既有cache lifecycle2/2通过（64102，10.3秒）；精确0坐标首失败为CSS序列化6e-6px，改新定位断言<.01px，原RGBA零差异不改。负对照只恢复旧material MO全树观察，同原断言失败：query4→1926，31838 exit1，独立保留非产品FAIL。类型instrument overload两错已修复，最终构建21582通过3745模块。fresh98835 fixture准备中、Vite29529，下一动作当前正式原60秒Firefox、再依结果继续冷启动/最终三引擎/长测。恢复/终端输出/ACK不改、50ms不变、R1开放、Windows不重复/发行延期/未提交推送。

2026-10-07当前实际状态：材质交界重叠产品正式原60秒WebKit46ms PASS（3936样本、9/3826长帧、五校验循环、双PTY约117KiB/s、末轮校验通过/11056清理0）；同源Firefox55ms FAIL（1620样本、5/3634长帧、四循环、双PTY约117KiB/s，40003清理0）。不把旧Chromium/FF诊断合成最终全通过。下一隔离诊断仅terminal-container contain:layout paint style，检验默认renderer damage成本，不改输出/解析/保护/ACK/测量标准；fresh fixture47133准备中、Vite29529。baseline生产构建不变，新增可选长测bounded轨迹仅>60秒使用，每周期增加真实回到原点pointer，保留原60秒轨迹不变，避免长测窗口逐渐离屏而无法持续交互。102单位/构建/新全画面native像素2/2已通过，最后三引擎功能/延迟/长测仍未补。50ms不变、R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07最新接管，覆盖以下候选的运行中措辞：独立活动shadow plane完整60秒WebKit55ms FAIL、只扁平化活动body材质51ms FAIL；两轮均双PTY持续、末轮16MiB目的校验完成，5187/92422 fixture已清理exit0，不采用额外图层诊断。当前产品候选把wallpaper-only header背景延伸到body交界下2px，消除fractional DPR独立AA交界泄漏，并移除交界安全带以恢复已知不透明遮挡；应用内容没有缓存/改变，保守正文bbox及所有新回归保留。WebKit DPR1/1.25全画面零差异/held reveal/focus2/2（43111，12.8秒），102单位（38436）与生产类型构建3745模块（21101）通过。fresh fixture11056（.local/e08-fixture/fixture-3429568536）、当前正式baseline WebKit原60秒96187运行，报告web/.local/e08-overlapped-material-webkit-report.json；Vite29529。下一动作收取完整实测和末轮校验、结束fixture，再决定保留并完成同源三引擎与稳定性证据。50ms/真实負载不变、R1开放、Windows不重复、发行延期、未提交推送。

2026-10-07当前接管（覆盖下方候选）：CSS path和SVG crisp rect精确可见条带均未通过逐像素对照，DPR1/1.25原path最大4/185，保留内层材质交界±1px安全带后最大2/25；SVG最大3/5。它们均撤回，不运行/宣称性能PASS，不放宽全画面零差异。新增pixel/held reveal/focus及纯geometry两项单位完整保留；原native shadow/终端/恢复断言不改。自动审批拒绝同时删除geometry回归的回退，已改安全产品回退并保留两测试/helper，无需审批/阻塞。恢复保守bbox后WebKit DPR1/1.25全画面与原无mask零差异、held露出/focus/布局3/3（87177，21.2秒）；最新类型构建50498通过3745模块。当前新隔离诊断只令活动原shadow-plane native transform层常驻，不改图像/颜色/轮廓/业务状态；完整原60秒WebKit诊断使用fresh5187 fixture（.local/e08-fixture/fixture-1780489900），报告web/.local/e08-contained-active-shadow-plane-webkit-report.json，Vite29529。下一动作收取诊断/末轮校验并清理fixture，再决定采用/撤回及最终同版本三引擎/稳定性。102单位已通过，当前R1仍开放/50ms不变/Windows不重复/发行延期/未提交推送。

2026-10-07最新接管：两种阴影alpha crop的union遮挡正式原负载WebKit仍53ms FAIL（3888样本、12/3818长帧、五校验循环、双PTY约116KiB/s），22/64阴影tile共490970/1489972像素停画，较完整tile增至33%却无延迟收益；66991结束、89298 fixture清理0。两轮失败报告保留，整个新shadow cull及其通过专属测试/单位已撤回，原阴影与终端/恢复正确性回归保持。CSS paint containment+overflow-clip-margin的WebKit native探针实际不支持（两语法false），直接开启截断阴影，不入产品。新候选改正文遮挡从单个bbox为分离可见条带同向CSS path union，仍以已知opaque圆角内区建立遮挡/1pxAA排除；手势露出各条带提前64px、只扩大、不反复收缩、最多32块则安全退回完整正文。解析/ACK/保护/尺寸/恢复均不变。102单位16975和类型/生产构建61786通过3745模块，WebKit原遮挡/布局/covered terminal及DPR1/1.25全画面严格对照93926运行，Vite29529，无压力fixture。下一动作收取逐像素/held露出/保护回归，再fresh正式原60秒WebKit负载；最终三引擎/稳定性尚未补，不闭R1/50ms不变，不重复Windows/发行延期/未提交推送。

2026-10-07续接：双固定阴影opacity候选正式WebKit原60秒负载74ms FAIL（3672样本、21/3700长帧、4次真实校验传输、双PTY约115KiB/s）；20323结束，22924 fixture清理exit0，候选全部撤回，失败报告保留web/.local/e08-shadow-opacity-webkit-report.json。新候选由既有缓存窗口几何计算完整原生阴影tile区间，只隐藏被上层已知不透明圆角内区完全覆盖的tile；平移不增加布局测量，手势内露出后不再隐藏直到release，晚cache发布/回退/尺寸变更立即失效。101单位通过，生产类型构建98811通过3745模块。初次新功能测试覆盖窗未包含未裁剪透明边界，实际只6块可隐藏，修正测试覆盖窗尺寸而未改隐藏安全条件；原阴影轮廓/晚发布/失败回退/有界布局4项已通过。当前新逐像素/held露出测试25756运行，Vite29529，无真实压力fixture。下一动作收取新测试后fresh原60秒严格WebKit验收，必要时撤回候选。仍需最终同版本三引擎/功能/稳定性证据，50ms不变、R1开放、不重复Windows、不发行/提交/推送。

2026-10-07当前接管：完整原负载WebKit border-image71ms FAIL（3768样本、22/3715长帧、5校验循环），focused-only body层63ms FAIL（3816样本、19/3748长帧、5校验循环），双流持续且末轮校验及fixture清理均0；不进入产品。当前候选两套已解码原生阴影裁剪分别固定在b元素，只由opacity 0/1切换，没有动画/新参数/内容截取；仅活动16个variant有opacity提示，所有后台无提示。修改WindowFrame/CSS及原阴影回归，使断言真实active图像/opacity和绘制面积，晚发布/失败回退/原轮廓等原语义保留。WebKit五项原生shadow/晚发布/失败/held露出/有界布局70318通过24秒。原100单位与cache来源/恢复服务不改；当前类型构建及DPR1.25八态原生截图对照并行中，Vite29529，无压力fixture。下一动作收取截图/构建，再fresh原负载测试，最终三引擎/稳定性/功能未补，50ms不变、R1开放，不重复Windows/发行延期/未提交推送。

2026-10-07当前接管：原生阴影8层产品正式WebKit54ms FAIL（3888样本、12/3805长帧、五校验循环、双PTY约117KiB/s），53边缘报告保留；1536 fixture清理0，额外边缘层已撤回。新候选把拖动/resize输入保护改为稳定透明shield，仅自身active显示，不改变应用DOM树继承pointer-events；nativecapture、即时位置/resize、退出/Escape/捕获丢失/刷新保护保持。WebKit真实iframe输入保护/退出、captureloss、drag/Escape/刷新和8resize四项28106通过19.3秒，100单位59573通过1.20秒，类型/生产构建30297通过3745模块。fresh fixture3264（.local/e08-fixture/fixture-3326981641），正式baseline WebKit60秒报告web/.local/e08-gesture-shield-webkit-perf-report.json实测中，Vite29529。下一动作收取结果/末轮校验并结束fixture，再决定保留及最终三引擎/功能/视觉/较长稳定性。E08仍未达、50ms不变，Windows不重复/发行延期/未提交推送。

2026-10-07当前接管：外移阴影面诊断104ms FAIL已撤回，failed报告web/.local/e08-contained-shadow-siblings-webkit-report.json保留且33298清理0。活动8边缘层对照49ms，尚无稳定收益；产品候选仅加focused已解码8原生边缘will-change，仍保留原图像/形状/elevation且不升所有后台窗口。类型/生产构建13092通过3745模块。fresh fixture1536（.local/e08-fixture/fixture-3802243326）、正式baseline WebKit60秒报告web/.local/e08-shadow-edge-retained-webkit-report.json运行中，Vite29529。下一动作收取真实结果/末轮校验并清fixture，再依收益保留或撤回；最终三引擎/功能/视觉/稳定性仍待，50ms和负载不变，不重复Windows/不做发行/未提交推送。

2026-10-07当前接管：正式驻留focused/moving层产品WebKit60秒51ms FAIL，全部非延迟/末轮目的校验通过，29485完成/16491清理0；此前49ms不能混入当前全通过。完整同负载no-shadow诊断38ms（4080样本、5轮校验、双PTY约117KiB/s）且清理0，只是关闭阴影/滤镜的诊断不能作为产品PASS。下一候选仅活动窗口的8个原生阴影贴片独立transform层，保留每个原生像素/原颜色/elevation/内容，产品CSS尚未采用；fresh fixture22993（.local/e08-fixture/fixture-2868000369）已就绪，报告web/.local/e08-contained-shadow-layers-webkit-report.json正在运行。Vite29529，R1开放，最终三引擎/稳定性/功能视觉待补；Worker/scissor撤回且恢复与HEAD一致，不重复Windows，发行延期，未提交/推送。

2026-10-07当前接管：当前containment原生产品的Firefox活动层驻留完整对照47ms PASS（1656样本、1/3676长帧、五次16MiB校验、双PTY约118KiB/s、63.465秒；原native复制仍2919ms），报告web/.local/e08-contained-active-firefox-report.json，39223退出0/48617 fixture清理0。以此证据产品保留仅focused/moving窗口原生transform层，不给所有后台窗口驻留；正式同源码三引擎/长测仍待确认，旧Wk层失败不升级本候选。最新保留100单测23638通过1.06秒、类型/生产构建38666通过3745模块。当前完整WebKit桌面/原生IDB/终端像素及隐藏恢复39612为26pass/1仅Chromium原生IMEskip，2.3分钟，零失败。Vite29529，fresh HTTPS9444 fixture16491（.local/e08-fixture/fixture-562637233），新产品baseline完整WebKit60秒正启动，报告web/.local/e08-retained-contained-webkit-report.json。下一动作收取末轮验证/正式延迟，结束fixture，补Chromium及Firefox产品baseline、较长稳定性和最终全引擎功能/视觉。50ms不变，R1开放，恢复service与HEAD一致，自定义worker/scissor均撤回；只保留官方WebglAddon(true)像素正确性修复。未提交/推送，Windows不重复、发行延期。

2026-10-07最新接管：SnapshotWorker真实完整Firefox52ms FAIL（1632样本、3/3672长帧、五轮真实16MiB校验、双PTY约114KiB/s、63.999秒，末轮通过且fixture65644清理0），虽实际副本Worker1、前台复制0且2043次IPC仅143ms/426seed/1827entries，但没改善目标。Worker及五专属通过单位/新增专属浏览器故障用例已撤回；恢复service和原生snapshot浏览器测试再次与HEAD一致，失败报告web/.local/e08-snapshot-worker-firefox-verified-report.json保留。首个6.062秒尝试仅诊断URL未就绪错误无延迟结果，独立保留。空窗口原时间戳/doubleRAF控件两次（passive真/假）各1920样本均p95 33ms，input p95 16、firstRAF1、secondRAF18；不属真实验收，说明FF仍有桌面额外帧等待，passive不是收益来源。当前产品仍containment+官方preserve buffer：记录Chromium30.8/WebKit49 PASS、FF51FAIL；Vite29529无压力。最新原生产品构建正在重做，新fixture准备中，下一动作仅当前contained窗口的FF活动层驻留对照（以前旧源码WK驻留失败不提升本对照）后依证据保留/撤回。R1开放、最终功能/视觉/稳定性待补；不降低50ms，未提交/推送，不重复Windows，发行延期。

2026-10-07当前接管（覆盖以下历史会话）：窗口containment+官方保留原生缓冲区构建完整60秒原负载WebKit49ms PASS/3960样本、Chromium30.8ms PASS/2184样本、Firefox51ms FAIL/1644样本；双真实PTY持续、8窗口、万目录、真实16MiB循环和最终校验全部满足，三fresh fixture均清理exit0。仍不关闭R1。为处理Firefox输入排队，新增专用快照Worker：仅已验证terminal-output日志增量复用副本，其他commit/显式flush完整播种保留直接应用值/代理回退；完整快照克隆和同格式current/previous IDB事务在worker执行。同步session journal即时写、解析/终端ACK不改；RPC实际持久修订后才清尾，worker中断先终止再原生事务回退，持久错误保留未保护/日志，闲置30秒释放副本，清理排空后终止线程。五个原始格式/连续日志/旧副本禁止覆盖/原子quota失败/正文历史单位新增，总105/105通过1.38秒；类型失败为新增测试postMessage重载类型，修正转发签名后最新类型/构建1722通过3746模块，检查52802退出0。Firefox真实worker与startupfallback并发/代理、transport与persistence失败/重试/跨账号清理、原隐藏终端5/5通过14.1秒。fresh fixture HTTPS9444会话5975（.local/e08-fixture/fixture-3436688196），原负载Firefox60秒69730正在运行，报告web/.local/e08-snapshot-worker-firefox-report.json；Vite18573仍29529。下一动作收取延迟和持久修订诊断/末轮校验并结束fixture；新worker没有性能PASS，最终三引擎/视觉/稳定性待补，未提交/未推送，Windows不重复、发行延期。

2026-10-07当前接管（下方同日会话为历史）：自定义原生scissor已全部撤回；保留的更强原生像素回归在没有适配器、且独立于窗口containment的原始WebglAddon上也复现闲置露出/同步输出像素错误。仅恢复官方WebglAddon(true)缓冲区保留选项作为正确性修复，不修改解析/模型/保护/ACK或恢复服务。两个实际DPR1/1.25严格整图零差异、idle露出和同步输出回归40686共2/2通过10秒，最新类型/生产构建68240通过3745模块。显式窗口contain:size layout style候选此前8项几何/材质回退/遮挡终端回归通过，100单测通过；本次fresh HTTPS9444 fixture89776（.local/e08-fixture/fixture-858651894），完整原负载WebKit60秒24389正在运行，报告web/.local/e08-window-containment-webkit-report.json。Vite18573仍29529。下一动作收取正式延迟和最终传输校验、结束fixture后依结果推进；不将原生正确性修复算作性能收益。50ms未达，最终三引擎/视觉/稳定性待补，R1开放、未提交，Windows不重复、发行延期。

2026-10-07最新接管：preserved-scissor完整WebKit59ms FAIL（3984样本、15/3791长帧、五轮真实16MiB校验、双PTY约117KiB/s、61.553秒），两实际native renderer安装适配但末态完全遮挡，非延迟断言全部通过，84131 exit1/69940 fixture清理exit0。无收益已撤回helper/原生preserve设置及四个专属通过单位，保留DPR1/1.25原生像素/idle露出/同步输出严格回归（零容忍，不删失败记录）。终端与恢复服务再次与HEAD一致。新候选显式尺寸absolute窗口contain:size layout style，不做paint containment避免切断阴影；100单测85682（1.13秒）、类型/构建3598（3745模块）通过。WebKit九项native像素/大幅held/平移布局/离屏/材质失败回退/default遮挡/立即刷新3764运行，新fixture准备中，Vite29529，无压力浏览器。下一动作功能通过后fresh完整WebKit60秒；保持50ms及完整真实资源负载，R1开放、未提交，Windows不重复、发行延期。

2026-10-07最新接管：preserved原生scissor DPR1/1.25两项12276严格通过10秒，比较相同原生CSS裁剪条件下整图（包含半像素边缘），idle全露出、DEC2026held曝光不得提前呈现、释放后改变且与解除hook原生完全一致，GPU实际scissor断言不放宽。初始1.25参考错误把部分遮罩物理边缘像素与未遮罩参考比较（21通道差），已改为原生同遮罩全图零容忍比较。前held/default遮挡回归通过，104单测/最新构建通过。fresh HTTPS9444 fixture69940，私有目录.local/e08-fixture/fixture-3820044611；完整原负载WebKit60秒正在启动，Vite29529。下一动作收取正式性能/末轮校验并结束fixture，再按结果保留或撤回，没有性能PASS声明，50ms及全部真实负载保持；R1开放、未提交、不重复Windows、发行延期。

2026-10-07最新接管：新scissor严格闲置露出发现WebKit呈现纹理仍截断（71736通道差/max195），但画前GPU像素与baseline完全一致；布局强制读/全框scissor/GLclear/额外RAF/canvas图层/public刷新独立均无效，失败记录保留。官方原生WebglAddon(true)保留绘图缓冲区后DPR1暴露与idle完整画面严格零差异通过；适配器要求实际preserveDrawingBuffer=true，否则保持原生。当前通过public terminal.refresh请求重绘，保留DEC2026同步输出门禁/selection事件，撤回直接render/额外RAF/层/clear方案，不再为收缩曝光发额外重绘。104单测48574（1.05秒）、类型/构建34967（3746模块）通过；新DPR1/1.25零差异+同步输出保持/释放+解除hook原生对照、隐藏默认和held拖动36675运行中，Vite29529，新fixture准备中，尚无压力。下一动作确认像素全部通过再完整原负载性能。R1仍OPEN/50ms不变，未提交，Windows不重复、发行延期。

2026-10-07最新接管：同源码active-layer完整WebKit56ms FAIL（3936样本、12/3804长帧、五次真实16MiB校验、双PTY约116.7KiB/s、61.321秒），无收益不采用；33311 exit1/82770 fixture清理exit0。新原生GL scissor只限制遮挡区域fragment着色，不改model/atlas/完整字形上传/解析/恢复/ACK；曝光扩张同步原生全行重画、每次恢复GL状态，未知几何/不兼容渲染器保持原生全量，销毁/失去context解除适配。暴露区域由既有缓存几何提供，输出不读取DOM布局；四项真实边界单测，104/104单位63374（1.08秒）及类型/构建77856（3746模块）通过。新WebKit原生全屏baseline像素/实际scissor/idle重新露出、隐藏默认终端和大幅held拖动三项63237运行中，Vite29529，无真实fixture/压力。下一动作收取严格像素结果，再fresh完整WebKit60秒比较；失败不得放宽原显示断言。50ms仍未达、R1开放、未提交，Windows不重复、发行延期。

2026-10-07最新接管：纯平移绘制前合并遮挡完整WebKit56ms FAIL，3888样本、16/3785长帧、五次真实16MiB校验、双PTY约116.7KiB/s、61.219秒，全部非延迟断言通过，23175 exit1/18104 fixture清理exit0。功能10947六项通过一项新增布置失败（两错位窗口实际留下角落、保守bbox应完整）；用真实桌面geometry完全包住下层正文后，首帧held露出/输入刷新及原隐藏终端持续解析/保护/ACK/notice布局23800两项通过，像素露出原生WebGL回归也已通过。新候选暂保留以对照反复手势图层重建，下一动作同源码fresh active-layer完整负载对照；不重复旧层诊断当正式通过。100单测/构建仍对应当前产品，Vite29529，真实新fixture正在准备，无压力浏览器。50ms仍未达，R1开放、未提交，Windows不重复、发行延期。

2026-10-07 最新接管（以下运行状态均为历史）：100ms后台快照合并完整WebKit65ms FAIL，3840样本、16/3746长帧、五次16MiB目的端校验、双PTY约115KiB/s、61.549秒，全部非延迟断言通过，fixture11838结束0。快照复制1885→641、前台耗时2280→809ms却无延迟收益，已撤回该候选及其两个专属通过测试，恢复服务再次与HEAD一致；未删失败证据。该候选完整WebKit功能79284为23通过/1原生Chromium IME跳过，包含保留的原生WebGL局部露出/无新输出全屏恢复像素回归。新候选将原生纯平移标记为可合并遮挡更新，最新坐标仍事件内立即发布，下一paint前统一处理；焦点/布局/尺寸/材质/结构及非手势写入仍立即更新，清理取消未执行帧。新增真实大幅held拖动首帧露出和刷新保留输入回归。最新100/100单测45576（1.12秒）、类型/生产构建88137（3745模块）通过；WebKit定向10947运行中，Vite18573会话29529，无真实fixture/压力浏览器。下一动作收取定向回归，fresh fixture完整WebKit60秒测合并遮挡；保留50ms/八窗口/双PTY/万项目录/校验传输/恢复合同，R1开放，未提交、不重复Windows、发行延期。

2026-10-07 E08直接拖动最终双流实测：WebKit58ms（3768样本、13/3774长帧、四次16MiB校验循环、双PTY约115KiB/s、61.662秒），Firefox58ms（1572样本、3/3608长帧、两次校验循环、双PTY约101KiB/s、64.149秒），均仅50ms失败，fresh fixture各完整结束0。新候选合并已由同步日志保护的完整快照复制：100ms有界批处理、64KiB日志提前唤醒、异常/显式flush立即，IDB事务内不等待timer；同步日志/终端ACK时序不改，同一awaitPendingWrites包含途中新修订，原原子快照/预算/配额测试不放宽。新增延迟窗口内真实日志恢复及未推进假时钟的预算唤醒验证；最新102/102单测1.87秒、类型/构建45710 exit0（3745模块）。新fresh fixture准备中，下一动作完整WebKit100ms候选后同引擎对照/Firefox/Chromium，最终全部功能视觉。Vite29529，旧72340/45826 fixture已结束0，旧压力38460/89495已exit1；R1开放、未提交、Windows不重复，发行延期。

2026-10-07 E08最新状态：字形曝光适配器69ms与同构建原生上传78ms对照均FAIL；新真实WebGL局部非空字形上传/全量闲置露出像素验证发现露出错误，候选连同专属新增helper/测试已撤回，失败保留，不修改显示正确性断言。终端服务现再次与HEAD完全一致，解析/保护/ACK保持。纯拖动改为事件内直接发布非响应式transform，只有实际改变布局的resize继续RAF合批；原停住RAF并立即刷新恢复用例通过。首轮完整WebKit测43ms但仅一路PTY持续，整体FAILED，不能算通过；测试前置新增每条真实shell命令输出>32KiB及真实textarea焦点，确认双流再采样，原速率/50ms/60秒/最终双流断言不变。当前源码100/100单测935ms、类型/生产构建20341 exit0（3745模块），新完整WebKit正在运行，fresh fixture HTTPS9444会话72340，私有目录.local/e08-fixture/fixture-759429526；Vite18573会话29529，其他fixture/browser已结束。下一动作收取完整结果/确认目标文件后结束fixture，依次Firefox/Chromium及最终完整功能/视觉与长测，R1开放、未提交，Windows不重复、发行延期。

2026-10-07 E08最新接管（覆盖下方历史运行会话）：新终端原生WebGL字形上传适配器完整WebKit60秒69ms，仍FAIL；3744样本、19/3746长帧、四次真实16MiB校验循环，双PTY约114KiB/s、最大未确认86123B，全部非延迟断言通过。两个活跃宿主实际安装native-prefix，但末态完全遮挡（绘制0/full3426），尚不能据此证明局部上传收益；下一步同构建原生上传对照与局部露出像素/无新输出恢复验证。原解析、完整模型、ACK和同步保护语义保持。此前独立正常/活动alpha裁剪86ms、面积减少约34%无明确延迟收益；border-image短诊断118ms/原生默认终端20秒诊断112ms均不采用、不计完整验收。修复WebKit DPR1.25原生阴影SVG viewBox导致的错位，用实际物理像素布局和原长度缩放；八种深浅/材质/held原生对照无页面错误，阴影区域最大通道差2～12，非像素完全一致。Firefox完整桌面/隐藏终端回归21通过、1失败、1原生Chromium IME跳过；真实离屏指针流定位到屏幕外mouseup遗漏，新下一次屏内buttons=0同步结束手势，原离屏/resize/立即刷新恢复用例1/1通过10.5秒。102单测及类型/生产构建通过（单测早于最后两处小guard）；最终全引擎功能证据待补。目前无fixture/压力浏览器，上一94102性能exit1、75044 fixture清理exit0；Vite18573会话29529。50ms和全部真实负载保持，R1开放、未提交，不重跑Windows、发行延期。

2026-10-07 E08最新接管：已保留源码的共有alpha裁剪三引擎36.9/79/84ms；后续仅活动窗口常驻层WebKit101ms不采用，IDB同步clone p95 3ms/max9ms不足以支持为延迟重写恢复；完整真实资源但隐藏正文的诊断74ms（改变外观，不计通过），取消正文独立材质层115ms不采用。旧PTY/fixture均已结束，无压力浏览器。新源码将正常/活动两态原生阴影各自裁掉alpha为0的边缘，避免拖动全部正常高度仍承担活动高度的空白矩形；不改任何非零alpha/原生阴影tokens/伸缩轴区间/解码发布和生命周期。类型/生产构建1667 exit0（3745模块），99/99单测7530通过1.78秒；WebKit10项缓存/回退/竞争/纯平移/捕获丢失/拖动和立即刷新/离屏/隐藏终端resize及持续解析ACK恢复65048正在运行，新高度回归同时要求正常状态实际绘制矩形面积小于活动状态。Vite18573会话86104。下一动作收取65048，深浅材质开关×静态/held阴影与原生对照（含DPR1.25），fresh真实fixture测新候选完整WebKit/Firefox/Chromium；最后完整跨引擎功能证据和文档。50ms及真实负载保持，R1开放，未提交，不重复Windows、发行延期。

2026-10-07 E08当前接管：原生阴影alpha裁剪正式三引擎同构建60秒测量Chromium36.9ms PASS（1944样本、0长帧）、Firefox79ms FAIL、WebKit84ms FAIL；各4次16MiB真实校验循环、双PTY约110KiB/s、所有非延迟断言通过。阴影四态原生窗口区域对照最大通道差3～6，细微差异非像素完全一致，无页面错误。后续阴影遮挡106ms（22/64片、约36%面积实际跳过）无收益，连同专属测试/helper撤回；其2/2浏览器、100单测只覆盖撤回候选。恢复alpha裁剪构建48867已exit0；新isolated WebKit active-layer60秒对照81221运行，新增真实IDB同步clone耗时诊断，不改保护/解析/50ms等合同。fresh fixture HTTPS9444会话30847，私有目录.local/e08-fixture/fixture-1192931148；Vite18573会话86104。先收取81221完整终态，结束fixture，再按证据优化；最终源码完整跨引擎功能证据待补。R1仍开放、未提交，Windows不重复、发行延期；所有较早fixture/压力容器已结束。

2026-10-07 E08最新接管（覆盖历史会话）：无透明中央面的八片缓存阴影/独立正文材质完整WebKit92ms仍失败；合并光学面151ms、常驻不透明正文图层110ms、手势输入挡板111ms均无收益，全部撤回。当前新候选仅按正常/活动阴影alpha非零像素的联合边界裁掉透明边缘，保留alpha=1及伸缩轴映射，不改任何原阴影token。另独立修复原拖动lostpointercapture未退出的缺陷。生产类型/构建22759及99/99单测10912已通过；当前源码WebKit10/10材质/阴影fallback/竞争/几何/拖动刷新/捕获丢失/隐藏终端持续解析ACK与resize回归57111通过1.0分钟。视觉原生对照26338运行中，Vite18573会话86104，无真实fixture或压力浏览器；上一fixture47065已清理退出0、性能89638退出1。下一动作收取阴影视觉对照，fresh fixture测裁剪候选完整WebKit60秒，再同源码Chromium/Firefox及完整功能。50ms/8窗口/2真实PTY/万条目录/真实校验传输/double-RAF合同不变，R1开放，未提交，不重复Windows、发行延期。

2026-10-07 E08最新接管（覆盖下方运行会话）：无中央透明面八片阴影完整WebKit93ms；局部活动高度样式99ms；修正请求竞争/DPR切分/测量节点归属后独立正文材质平面92ms（3384样本、4次16MiB校验传输、双PTY约113KiB/s、55/3595长帧）；同源码全窗口常驻图层对照99ms，无收益不采用。全部保留50ms/8窗口/2PTY/万条目录/真实传输/double-RAF断言，R1未关闭。WebKit材质/阴影fallback/几何/持久刷新/隐藏终端resize首轮6/9：三个新测试前置错误是初始圆角误用32而实际28、CSS transform序列化空格（两项），修正测试不放宽几何/竞争/恢复断言。下一候选将不透明壁纸材质和缓存原生阴影合并同一 retained optical layer，不捕获应用内容、不改颜色/阴影token/正文/PTY；小窗口或缓存失败仍原生回退。源码已完成，当前类型/构建98936与WebKit九项回归81905运行；fresh真实fixture HTTPS9444会话88881（私有目录.local/e08-fixture/fixture-142705984），尚未发起PTY负载。Vite18573会话86104。前fixture32317/1327、性能13449/9316和浏览器22617已结束。下一具体动作收取构建/回归，测合并光学面完整WebKit负载，之后最新三引擎和视觉/完整回归证据；未提交，Windows不重复、发行延期。

2026-10-07 E08渲染架构第二阶段：WebKit完整fresh去阴影/滤镜对照77ms（非产品PASS），原生阴影拆层178ms，再原生CSS计算结果缓存为PNG边框121ms（8/8窗口实际缓存），拆为无中央透明面八片绘制93ms（3432样本、双真实PTY约114KiB/s、4次16MiB校验传输、60/3589长帧），以上正式结果仍仅50ms断言失败。阴影缓存不采样应用内容，保留正常/活动原始tokens，8种变体上限/异常原生回退/两态全部解码再发布/代次和URL清理；8项模糊/遮挡/阴影边界单测通过。四种WebKit截图初轮发现ResizeObserver循环，测量probe移到窗口外后复验exit0无页面错误；阴影区域有细微差异，非逐像素一致。新候选去掉全桌面:has(moving)样式失效，由既有同步遮挡器仅在活动frame标记原正常拖动高度，连中途焦点变化语义都保持；源码完成、构建会话待接管。前真实fixture11424及浏览器49408已退出；新fixture刚启动，Vite86104仍18573。下一动作测局部样式候选、再三引擎隐藏resize/缓存回退/可见像素及最终完整性能。不重复Windows，不更改PTY/恢复/50ms/8窗口等合同，发行延期，仍未提交。

2026-10-07 E08专项接管更新：stable region/overflow viewport＋保守手势曝光完整WebKit60秒p95 332ms（3360样本、双PTY约110KiB/s、四次真实校验传输）；提前准备仅活动窗口图层独立fresh测325ms（3144样本、双PTY约107KiB/s、三次校验传输），均仅延迟失败，不视为有效收益或关闭R1。96/96单测、最新类型/生产构建通过；只读专项审查未发现额外布局/清理/租约缺陷。隐藏终端回归已新增隐藏期间实际窗口resize断言，待三引擎执行。无效复用旧fixture的去阴影诊断15849已主动中止exit137（不计证据），对应fixture33660及容器停止；新fresh fixture HTTPS9444会话48691（私有目录.local/e08-fixture/fixture-3300614678）、完整去阴影WebKit诊断32718正在执行，Vite18573会话86104。新增仅诊断的第一/第二RAF分项，不改事件起点/double-RAF/50ms/输出或真实资源合同。下一动作收取隔离阴影开销证据后实施对应渲染架构，再最新三引擎性能/隐藏resize/视觉回归。未提交，Windows不重复，发行延期。

2026-10-07 E08关键新发现：WebKit原生IntersectionObserver忽略clip-path和legacy clip，两种候选都在真实xterm-screen可见性断言失败（分别26657/80314退出1），原终端paint-occluded标记不能证明该引擎停止绘图。新增固定布局region＋独立overflow viewport，仅隐藏viewport，xterm宿主宽高保留region真实CSS尺寸，外层布局不变；尺寸变化仍传入现有Fit/租约流程，不改TerminalModel/解析/ACK/保护。WebKit严格隐藏/持续ACK26/真实恢复内容/零输入重发/应用notice改变布局再露出回归1/1（18.9秒）通过（75225已结束）。几何缓存后完整WebKit344ms仍失败；去阴影/滤镜短对照156ms、120样本，非正式验收，未删除产品阴影。源码另保守扩张拖动中的正文可见区域，以减少逐像素裁剪重画，结束必收紧；Chromium三项新增/补强回归3/3（21.7秒），结构注册同Vue几何写入时全量测量避免丢observer记录；全部已执行容器结束。下一动作构建最新bundle并测WebKit完整60秒，再Firefox/Chromium及可见像素验证。当前fresh fixture HTTPS9444会话83785、私有目录.local/e08-fixture/fixture-340093597，尚未启动真实负载；Vite18573会话86104。R1仍开放、Windows不重复、发行延期，未提交。

2026-10-07 E08实时接管：原union路径裁剪实测Firefox221ms，明显恶化，已撤回；保留普通inset和终端表面完整遮挡裁剪，Firefox完整81ms（1356样本、双PTY约93KiB/s、两次校验传输、56/3278长帧），仍失败。独立原生IntersectionObserver验证xterm-screen隐藏/恢复，超过增量预算的输出仍持久保护且ACK推进，重新露出最新内容、刷新同会话/游标26、零输入重发，Chromium1/1（10.4秒）通过；初次断言误把保护日志结构当wire事件，诊断确认输出26已真实保存，已修正结构读取而未放宽内容断言。下一候选独立测稳定材质纹理一次预展开到显示分辨率；生产bundle该候选已构建通过，WebKit完整60秒会话32566，fresh fixture HTTPS9444会话55867，私有凭据.local/e08-fixture/fixture-3643436006/browser-credentials.json。Vite18573会话86104。此前所有真实fixture/容器已结束。正在源码实现有界布局失效和几何缓存，避免每次纯拖动重新读8个窗口布局，并修窄屏缓存圆角；不在当前测试中重建bundle。未提交，R1开放，Windows不重复，发行延期。

2026-10-07 E08续接：中断期间临时二进制/测试进程被环境清理，重建任务自有Master/Daemon/devfixture并使用fresh资源。可见壁纸高清栅格化Firefox完整102ms仍失败；加正文layout/paint/style边界完整90ms（1320样本、双PTY约109KiB/s、4次校验传输、119/3063长帧），仅≤50ms失败，不关闭R1。当前生产bundle为该contain候选，WebKit完整60秒正在执行，会话42429；fixture HTTPS9444会话86746，私有目录.local/e08-fixture/fixture-740189152。源码另加实际露出区域union裁剪候选（保守1px重叠/不支持inset回退/复杂度128上限），类型检查及缓存/遮挡6/6单测通过；尚未build/adopt。测试壁纸身份改读原样式源，保留真实缓存显示检查，不把临时blob URL当内容身份；已补离屏拖回/缩放/刷新回归待运行。未提交，Vite未启动，无其他浏览器。下一动作收取当前WebKit结果，然后构建/测新union候选，完成视觉/恢复及三引擎严格验收；不重复Windows组，发行延期。见[架构记录](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-06 E08后续诊断：预混材质Chromium完整60秒严格PASS 47.3ms（1440样本、4次校验传输），同架构Firefox174ms/WebKit403ms均只延迟失败，R1仍开放。原生绝对矩形绘制裁剪短诊断Firefox127ms而WebKit747ms且仅99样本，已撤回；恢复普通inset裁剪，保留先批量读尺寸再写样式和离屏材质不透明底色修复。WebKit自动GPU降级完整415ms，后台ASCII字形/行缓存Firefox短诊断129ms、实际已cached-rows，均无明确收益，代码全部撤回，原终端服务与HEAD逐字一致，不改变输出/解析/恢复。新候选仅将现有可见壁纸也先栅格化为高清位图（不换内容/配色），以避免SVG渐变/滤镜在软件重绘时反复求值；生产构建3744模块通过，正在Firefox完整60秒严格验收。当前fixture HTTPS9444会话32247（.local/e08-fixture/fixture-3965315738），容器测试会话36558，Vite18573会话70819；此前fixture均清理退出，无其他浏览器。下一动作收取新Firefox结果，再WebKit/Chromium复验、离屏拖回/可见像素/刷新恢复回归及文档收尾；尚未提交，不要求Windows复测，发行延期。详见[架构记录](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-06 E08架构优化继续：不透明壁纸缓存/保守正文遮挡99.9ms，拆分标题与正文光学平面82.3ms，统一拖动期间材质坐标、移除父窗口重复填色、遮挡在MutationObserver当前绘制周期内同步后，正式Chromium完整60秒负载p95 50.6ms（1392样本、双真实PTY约105KiB/s、4次校验传输、14/3459长帧），仅50ms断言失败，不能按四舍五入关闭R1。构建通过，最新缓存/遮挡恢复/拖动刷新/八向尺寸/主题独立7/7通过42.6秒；刚补同帧露出回归需复跑。下一候选将不变材质底色预混入缓存，避免每帧重复alpha混合，正在正式完整负载测量；未提交/未关闭，仍需三引擎及可见像素验证。当前fixture HTTPS9444会话86999（.local/e08-fixture/fixture-2068197407），正式测试容器59509，Vite18573会话70819；旧fixture和容器均已退出。下一动作收取本候选结果、补严格同帧回归与像素比对，随后三引擎验收；不要求用户重复Windows组，发行延期。见[架构工作记录](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-06 E08专项续接：共享壁纸材质缓存、手势非响应式绘制/即时恢复提交、物理像素对齐候选完整Chromium60秒负载最新p95 139.3ms（600样本、双PTY约94KiB/s、3次真实校验传输、仅延迟断言失败），仍不能关闭R1。已撤回过慢的应用内部强制分层（226.6ms）、永久窗口分层（163.4ms）、终端canvas初版（575.1ms）及合并通知版（164.4ms）；未降低50ms/输出速率/double-RAF，也未改原终端解析/持久化语义。材质/拖动/八向尺寸等Chromium6/6及模糊/遮挡几何初始4项已通过；遮挡随机可见像素校验已加待最终复跑。下一候选将已不透明壁纸编码为无冗余alpha栅格，并仅裁剪被上层不透明窗口确实遮挡的应用绘制（圆角/边缘保守排除、保留暴露区域单一包围矩形、不影响后台任务）。正在完整复验；未提交，需性能/可见像素/刷新恢复回归后再采用。详见[架构工作记录](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。当前fixture HTTPS9444会话69954（.local/e08-fixture/fixture-4290099296）、测试容器会话53586、Vite18573会话70819；此前会话已退出。下一动作检查新完整负载结果并复核遮挡可见性，不要求用户重复已完成Windows组；发行延期。

2026-10-06 E08 专项正在实施：用户明确授权超出微调的架构优化。候选将逐窗口递归 backdrop-filter 改为单张壁纸光学缓存（只采样壁纸，窗口重叠不再采样其他应用），并将拖动过程几何绘制与持久状态提交分离；不改配色、图标、形状、50ms/8窗口/2PTY/万条目录/真实传输/double-RAF合同。八窗口无PTY短诊断约47～50ms，但完整双PTY+持续校验传输两轮 p95 147.4/139.7ms，均失败，不能据短诊断关闭R1。正在进一步采集真实主线程/渲染分项；新缓存与手势改动尚未完成浏览器回归和视觉验收，未提交。候选生产构建与缓存模糊单元2项已通过。活动fixture HTTPS9444会话17431（私有凭据目录.local/e08-fixture/fixture-1577937187）、渲染CPU诊断容器会话16031、Vite18573会话70819；下一动作分析完整负载长帧并继续降低渲染开销，随后独立正式验收。R4已关闭，发行延期。

2026-10-06 R4普通Windows真机验证完成：收到e546c8f0唯一ZIP，12文件逐项读取、11项完整checksum/固定source及runner SHA吻合。Windows/amd64八阶段exit0、14必需检查各唯一PASS，无缺失/失败，整体CHECKS_PASSED_WITH_SCOPE_LIMITS；本地时间11:53:56～11:54:55，真实检查7.411秒。LF浏览器刷新保存、停止状态重启及cleanup均实际通过，见[真机通过报告](../../acceptance/reports/windows-device-passed-2026-10-06.md)。R4关闭，无需再给用户同组脚本；此前失败保留历史，不提升未测环境。更新当前剩余表和Windows入口文档。下一具体工作仅R1 E08严格延迟诊断优化，仍不满足50ms且不降低目标/改变冻结UI；R5发行继续按用户要求延期。当前无活动运行会话，本轮只核验报告及记录，无产品改动。

2026-10-06 收到cf83c5ea唯一Windows报告：归档SHA/全覆盖checksum/runner固定源码身份吻合；源码下载3.99秒及全部构建通过，11项真实功能和cleanup通过，browser save-readback失败，stopped-state restart未执行，整体12PASS/1FAIL/1未执行。以当前真实后端和Windows-UA Chromium本地复现HTTP200但unexpected-crlf，确认Monaco默认EOL按Windows平台选CRLF，而LF分支未显式设置。documentModel和history rebuild现在双向设置LF/CRLF，保持逐字断言；新增安全readback状态/类别日志。类型生产构建、Windows-UA LF/CRLF保存刷新2/2（48.4秒）、真实后端生产包Windows-UA完整14/14（exit0）、JS语法/diff检查通过。见[报告](../../acceptance/reports/windows-device-editor-eol-2026-10-06.md)。所属浏览器/服务/容器已退出，无活动接管；下一动作交付固定新版整段{}命令确认Windows修复。R4尚需真机新证据，R1 E08未满足，发行延期。

2026-10-06 第二次源码阶段失败：用户ba68e5d的-ArchiveOnly下载30.2秒exit -1，无child输出，14项未开始；不是脚本150秒超时，未收到ZIP，不确定网络或外部终止根因。移除stage和download的嵌套EncodedCommand，改任务自有ps1文件和UTF8 JSON参数；日志保留PID/启动/下载开始完成标记，helper/spec不进报告。隔离PS7 harness通过路径空格/中文特殊参数/stderr/exit7/缺命令1/启动标记/归档安全/校验隐私；精确新版下载函数经新版stage实取ba68e5d ZIP，4.0秒完成、1219文件安全解压。见[追加报告](../../acceptance/reports/windows-device-source-2026-10-06.md)。本地容器已退出、无活动会话；不更改产品/UI/14项合同、不提升Windows状态。下一动作交付固定新版-ArchiveOnly命令收取唯一报告；R1严格延迟和R4 Windows真实证据仍未完成，发行延期。

2026-10-06 Windows新真实联调入口在源码获取阶段失败：用户提供固定2bfac47的console，clone30.3秒/exit -1，尚未执行build或14项产品检查；未收到ZIP，不推定具体Git网络根因。新wrapper补结构化参数、UTF8双流、native stderr非终止、child exception/阶段detail保留及明确exit code；新增-ArchiveOnly按固定提交HTTPS ZIP获取源码，default Git失败也fallback且保留原FAILED/恢复标记。全归档先校验固定root/遍历/Windows路径/重复/符号链接及容量上限，script-source SHA仍必须相同，报告记录sourceMethod/archive SHA。隔离PS7 harness通过stderr/child7/缺命令1/字面参数/安全归档/报告checksum隐私，实际固定GitHub ZIP下载和1218文件安全解压通过，见[本轮报告](../../acceptance/reports/windows-device-source-2026-10-06.md)。不改产品/UI/14项checker、不重跑已通过Go或旧浏览器组，不记Windows新PASS、不做发行；所属测试容器已退出。下一动作交付同一固定新版源码和脚本的-ArchiveOnly命令，普通设备运行一次并回传唯一ZIP；R4实际结果和R1 E08仍待完成。

2026-10-05 非发行本地增量完成：用户要求继续本地能完成的工作并排除发行；R2防火墙读取/变更门禁已修，仅Linux firewalld支持预览应用，状态读取保留，类型检查和浏览器5/5通过。R3新增cmd/device-validation、真实Chromium编辑刷新脚本及windows-device-validation.ps1，单节点真实Master/Daemon/HTTPS/WSS、初始化账号/权限拒绝/UTF8文件/备份恢复/监控/扩展升级移除/实例启停/终端实际文件副作用与重挂/浏览器中文草稿F5保存/停止资源后重启身份正文及同任务回执/明确清理共14/14，最新Linux源码及本轮正式frontend实跑11.388秒、exit0。Go安全2/2、vet、Windows交叉构建、JS语法、PS7报告harness通过；固定40hex、script-source SHA一致、fresh私有目录、每阶段进度与15秒心跳、缺失/重复/清理失败/非Windows拒绝、唯一ZIP与全覆盖校验，凭据/状态/进程日志/截图不进ZIP。新入口不重跑旧visible WebKit/native/race组，不做宿主权限修改和发行。R1新布局隔离ABBA探针像素一致但约923～953ms，不足以采用，保持UI/完整材质/50ms目标且原失败不改PASS。详见[本轮报告](../../acceptance/reports/local-remaining-2026-10-05.md)、[性能诊断](../../acceptance/reports/render-layout-diagnostic-2026-10-05.md)和[更新剩余表](../NON_UI_REMAINING_2026-09-26.md)。全部所属fixture/浏览器/容器已退出，无会话接管。下一具体动作交付固定提交的新普通Windows命令并收取一次14项真实结果；E08未达到，发行按用户要求延期，不称全部完成。

2026-10-05 三子代理审查与用户范围调整完成：见[审查报告](../../acceptance/reports/scope-review-2026-10-05.md)和[最新剩余表](../NON_UI_REMAINING_2026-09-26.md)。本轮用户明确要求不再强求磁盘耗尽/虚拟机等难提供环境，三个只读审查覆盖存储备份调度/Engine失败、Windows平台/脚本与终端/性能兼容性；主代理复核提交4eb5eee源码和既有真实ENOSPC、SIGKILL、24h、TLS/CA证据。对应物理/独立公网/全桌面组合改为本次非阻塞未实测，Windows额外headless/IME/通知中心/特权服务任务修改也仅作兼容性记录；旧Vim/一次撤权超时转历史观察，不称根因修复。Windows规则应用/回滚并未支持，删除该错误必测要求；确认SystemApp误把netsh读取能力当预览/应用能力，仍待本地修复。当前只保留R1 E08目标、R2门禁、R3普通Windows真实后端与发行恢复/回退脚本、R4一次合并真机结果、R5最终发行收尾。此次仅改范围/证据文档，不改产品/脚本，不运行新测试或服务，无会话/容器接管。下一具体动作R2及R3，不重复已关闭visible WebKit组，不要求用户提供危险环境；E08已失败不能由审查变PASS。

2026-10-05 Windows第十六轮通过并结束本组补测：见[报告](../../acceptance/reports/windows-sixteenth-report-2026-10-05.md)。固定8d6b14e，ZIP16条/15条checksum完整匹配，run.json与最终来源/时间、阶段检查点一致。Windows真实headed精确剩余同例3/3（14.882/19.153/11.334秒），无retry/skip/flaky/全局错误；十个阶段全PASS，正式构建16.35秒、浏览器51.43秒、JSON49.721秒，fresh端口12408、原45/5秒保留。结合第十五轮其他四项各3/3，用户要求的visible WebKit补测组完成，不再要求重复本组脚本；原14/15 FAILED与本轮3/3按不同固定版本独立保留，不伪造新版单批15/15。仅更新验收文档，不改产品/UI/脚本、不重跑已通过测试、不提升全Windows/E08状态。本轮无启动服务、容器或测试会话。下一动作只在有具体新复现、可区分性能假设或独立缺失环境时推进对应范围；完整缺口见最新剩余表，不再把本组查询恢复当待补项目。

2026-10-05 第十五轮交付前核对：1,158个本地文档链接无缺失，差异检查通过，所有本轮所属容器/会话均已退出。源码、脚本和验收记录采用同一固定提交交付。最新下一动作是用户Windows运行交付版本的`-Followup -BrowsersOnly -WebKitOnly -Headed -TaskRecoveryOnly`，只补剩余同例三次，返回含版本/时间的新ZIP；无需再收集已正确回传的d75b825旧结果。未将Linux或脚本harness计为Windows新增通过，完整平台/E08/外部缺口保留。

2026-10-05 第十五轮本地收尾：类型与正式构建通过（3740模块/Vite3.88秒），WebKit headed四导航例各3次12/12、49.300秒；Firefox4/4、17.237秒；Chromium4/4、13.974秒；引擎依次一worker，fresh正式preview，无skip/retry/flaky/全局错误且实际模式/45秒预算JSON核对。真实脚本AST生成的精简参数WebKit headless准确同例3/3、16.251秒，通过title/mode/次数/独立server核验。最终PS7.4 harness通过新3次缺失/错误标题拒绝、旧模式/构建/显示/资源边界、精准CollectRef及恢复checksum/保留FAILED/中断不造PASS；真实回传由脚本title/display函数仍确认14/15。证据`.local/evidence/windows-report15-2026-10-05/`；本轮所属容器和会话均结束，早期GUI启动未开始测试及两次harness夹具纠正单独记录，不算Windows新失败或产品PASS。下一具体动作差异/文档链接检查、提交推送固定版本，只交付-TaskRecoveryOnly Windows同一项3次回传；自然轮询时延、默认headless、完整原生/E08/外部缺口保持。

2026-10-05 Windows第十五轮已核对：见[报告](../../acceptance/reports/windows-fifteenth-report-2026-10-05.md)。固定d75b825、15/15摘要及16个ZIP条目完整匹配，真实headed注释15条；正式构建18.33秒PASS，WebKit14/15。刷新/queued/日志/Worker终端各3/3，唯一次prevented-navigation准备阶段waiting断言5秒失败，尚未执行departure；同例后两次通过，不把准备失败归为恢复失败。此前同日上传与第十四轮ZIP逐字节相同，不计新执行。测试现通过应用已注册的实际轮询回调建立held read，原45秒总计/5秒恢复及全部取消/缓存/无帧断言保持，产品/UI不改。新增-TaskRecoveryOnly严格同例3次，报告独立模式及标题，遗漏范围不导入旧PASS；文件名含版本/时间，CollectRef按实际记录版本选择，保留原结果并重建收集校验清单。下一具体动作完成本地回归、文档及固定提交交付，仅补Windows剩余同例3次；完整原生/E08/默认无界面及外部缺口保持。

2026-10-04 第十四轮 Windows 回传核对及本地修复完成：见[报告](../../acceptance/reports/windows-fourteenth-report-2026-10-04.md)，固定63cacc0、17/17摘要及完整ZIP匹配，Windows正式构建24.36秒PASS，OS端口3570 fresh preview成功，WebKit12/15；无帧/真实queued/日志各3/3。剩filter元素5秒缺失、held read准备迟到和初次导航35秒后接管耗尽45秒，Windows原因未定。新增可选-Headed仅同五项3次环境对照，真实resolved launch fixture逐执行注释验证、缺失/重复/不一致BLOCKED，PWDEBUG隔离恢复，原45/5秒不变。初次Linux headed12/15、175.945秒及私有两次同例Worker探针0/1保留，确认reactive controls导致postMessage DataCloneError；controls/links现转为纯数据，既有终端两次真实刷新及controls/link续接增加Worker保持断言，UI/协议/ACK/不重放不改，不声称解释Windows失败。新类型构建3740模块/Vite7.73秒、最终类型、90单测/17文件/1.850秒、PS7最终harness通过。本地精确五项3次headed15/15、181.922秒和headless15/15、91.029秒；WebKit六文件终端恢复62/62、231.936秒（源码import/fresh开发server），Firefox/Chromium两文件正式preview7/7、66.054/51.943秒，浏览器依次一worker，无skip/retry/flaky/全局错误。真实JSON/finally核对模式及标题次数，上传12/15仍FAILED、两不完整标题/三诊断复制且不导入未选Go/native/SDK；真实headed15/15可过，删一launch注释即INCOMPLETE。证据`.local/evidence/windows-report14-2026-10-04/`，全部本轮所属容器/会话结束，无接管句柄；初次失败及后续成功独立保留。1,151本地文档链接无缺失、差异检查通过。下一产品验证动作用户Windows运行交付命令指定的固定源码/脚本 `-Followup -BrowsersOnly -WebKitOnly -Headed`，正式构建及严格15/15回传；完整原生/E08/外部缺口和默认无界面可靠性保持未关闭。

2026-10-04 第十三轮 Windows 回传核对与本地修订完成：见[报告](../../acceptance/reports/windows-thirteenth-report-2026-10-04.md)，固定c63bcb4，14/14摘要及完整ZIP集合匹配，正式构建24.04秒PASS；5173被占用导致WebKit启动失败、实际零执行，不是15个产品失败，也不证明前轮修订通过。脚本现每个引擎启动前选择OS空闲loopback端口，同步baseURL/server端口，strictPort/fresh保持，不杀/不复用原监听；报告保留有界全局错误及实际URL/选择端口。产品/UI/用例/诊断/45及5秒保持。真实PS7 harness/OS端口选择/类型检查通过；本地复现原启动失败、选中端口抢占失败关闭及6种非法输入拒绝；保持5173占用时WebKit精确Windows五项各3次15/15、97.036秒，Firefox正式preview7/7、66.612秒，Chromium开发server7/7、78.294秒，引擎依次一worker，无skip/retry/flaky/全局错误。真实AST函数核对本地JSON标题次数/server元数据，实际回传零执行由真实finally拒绝、保留FAILED/启动错误/URL及未选范围、生成独立私有ZIP。harness另验证名义15PASS不能掩盖全局错误。既有loopback teardown错误保留，不计新本地构建/Go/native/SDK/单测PASS。1,145本地文档链接及差异检查通过，远端发布前复核仍c63bcb4；私有证据`.local/evidence/windows-report13-2026-10-04/`，本轮所属容器/会话结束，无接管句柄。脚本/配置/报告同一固定提交交付，下一具体动作用户Windows运行新版`-Followup -BrowsersOnly -WebKitOnly`，正式构建PASS且同五项各3次严格15/15回传；前轮Windows任务恢复/完整原生/E08/外部缺口保持。

2026-10-04 第十二轮收尾：最终差异检查和1,137条本地文档链接通过，远端main仍为该回传90658d0；本轮所属测试容器和会话均已结束。源码、脚本和本条验收记录在同一提交交付。下一条具体验证动作是用户在Windows运行本轮固定版本`-Followup -BrowsersOnly -WebKitOnly`并回传ZIP，要求正式构建PASS和原五项各3次严格15/15；本地结果不关闭Windows剩余恢复超时及完整原生/E08/外部缺口。

2026-10-04 第十二轮 Windows 回传核对与本地修订完成：见[报告](../../acceptance/reports/windows-twelfth-report-2026-10-04.md)，固定90658d0，17/17摘要及完整ZIP集合匹配；正式构建22.71秒PASS，WebKit12/15，无skip/retry/flaky，终端/日志/无帧恢复各3/3。剩任务窗口一次准备耗尽45秒、一次首列表未就绪、一次导航后计数原5秒失败，后一Windows原因未确立；旧queued预先普通轮询变为7，也不能独立证明该次1/3的计数是导航恢复。仅修订测试准备/诊断：可见可用任务应用原生Enter，确认真实首列表HTTP/body及筛选控件；queued mock等浏览器实际beforeunload后才切换，baseline0/list基线计数严格验证，防止自动化传递迟滞时提前满足恢复。原45/5秒/queued/不写入/ACK/不重放/错误断言保持，产品/UI不改。被动诊断用实际JSON消费、DOM数字计数变化及浏览器departure时间，严格数字/静态事件白名单、120上限，无body/字段/存储，不计绘制通过。类型及新正式构建通过（3740模块、Vite5.70秒），类型复验退出0。最终两文件7项WebKit21/21、114.264秒，Firefox7/7、42.794秒，Chromium7/7、41.167秒，独立fresh preview、一worker、引擎依次，无skip/retry/flaky。受控真实模块直到第三次普通summary轮询6.453秒释放，列表6.538秒/边界6.590秒且baseline0，6.748秒完成原恢复断言，独立1/1、9.729秒，不冒充Windows因果；边界绑定精进前21/21、113.237秒保留为初轮。真实PS7脚本harness通过，实际上传/本地JSON标题次数/12/15拒绝及三份诊断真实复制通过；故意失败探针0/1、7.096秒，2823字节诊断验证body消费0/7、DOM0/7、browserAt/beforeunload及任意字段/类型排除，独立保留不计PASS。实际CLI --list严格五标题各3次15项仅选择核验。证据`.local/evidence/windows-report12-2026-10-04/`；本轮所属容器/会话均已结束，无接管句柄，既有大chunk警告保留，不计Go/SDK/单测新PASS。下一具体动作差异/文档链接及远端版本复核后提交推送同一源码/脚本固定版本，用户Windows运行`-Followup -BrowsersOnly -WebKitOnly`（正式构建且严格15/15）回传ZIP；Windows实效/冷启动/指针绘制/完整原生/E08/外部缺口保持。

2026-10-04 第十一轮 Windows 回传核对与本地修订完成：见[报告](../../acceptance/reports/windows-eleventh-report-2026-10-04.md)，固定`7426801`，16/16摘要匹配，正式构建17.67秒PASS；WebKit13/15，无skip/retry/flaky，日志3/3。剩一次初次桌面计数断言（未到导航边界）及一次终端控制台准备45秒超时（未挂载终端），不确立Windows原因。仅测试准备修订：首次真实summary HTTP200及body后原5秒渲染断言，冷初始化计原45秒；终端五步可见/可用按钮原生Enter，后续指针及原协议/worker/ACK/不重放/刷新/存储失败保持，不计准备指针/冷启动5秒性能。产品/UI未改。初轮WebKit36/36、240.615秒；Firefox初轮6/7、79.796秒，两条取消记录及新文档session前的第三读取说明原总计1断言未区分请求，不能确立双记录身份。现每文档读取编号/重置离开标识，严格held read2恰一次离开后取消且其他取消均在离开后，queued/不写入等检查保持；初轮失败JSON/trace独立保留不计PASS。最终受影响两文件7项：WebKit21/21、182.344秒，Firefox7/7、65.855秒，Chromium7/7、49.663秒，浏览器依次运行，无skip/retry/flaky；类型检查、真实脚本harness、实际上传/本地JSON标题次数/13/15拒绝及两份诊断真实复制通过；CLI --list严格五标题各3次15项仅选择核验。1,133本地文档链接无缺失、差异检查通过。证据`.local/evidence/windows-report11-2026-10-04/`；本轮所属容器与会话均已结束，无接管句柄，既有loopback/teardown警告保留。生产源/依赖/构建配置未改，使用既有正式bundle，不冒充新构建或Go/SDK/单测新PASS。下一具体动作提交推送固定新版源码/脚本后用户运行`-Followup -BrowsersOnly -WebKitOnly`，正式构建且严格15/15回传；Windows实效/冷启动/指针绘制/原生/E08/外部缺口保持。

2026-10-04 第十轮 Windows 回传核对与修订完成：见[报告](../../acceptance/reports/windows-tenth-report-2026-10-04.md)，固定`3ef289a`，17条摘要全部匹配，正式构建17.42秒PASS；WebKit12/15，无skip/retry/flaky，终端3次31.485/33.691/34.322秒通过，剩两次控制台click稳定等待45秒总超时（未发日志读取/未到导航边界），及一次无帧汇总原5秒断言失败（后续HTTP完成），具体Windows原因未确立。现存活文档读取恢复立即重查活动任务查询，保留已确认缓存，cancelRefetch:false避免替换已排队读取，应用销毁退订；pagehide/旧帧截止失效与原250ms恢复保持。实际QueryObserver无订阅基线0/1，在250ms仍两次读取，接入后回归通过；不当作上传失败复现。日志测试用实际原生Enter打开同一应用/实例/控制台，保留可见控制、真实草稿输入/interval回调/两种queued边界/刷新/无写入及原错误断言，不计这三步指针覆盖，不加force/sleep/retry/skip/限时。四项任务/日志均补静态阶段，失败诊断新增数字任务计数及按钮几何。90/90单测（17文件、1.960秒）、类型构建（3740模块、Vite6.23秒）、正式WebKit36/36（210.400秒）、Firefox12/12（82.874秒）、Chromium12/12（71.432秒）通过；Firefox起段与WebKit末段约21秒重叠，正确性结果不冒充隔离性能/E08。脚本harness、真实本地/回传JSON标题次数/12/15拒绝与三份诊断解码通过，真实CLI --list严格五标题各3次15项，不算另一次运行PASS。探索暂停应用时钟未复现指针停顿（click成功使预测断言失败），初次两失败探针独立保留；随后故意诊断断言单独0/1、2.967秒、857字节，真实复制包含两阶段/计数7/选中控制台几何，不计产品PASS。证据`.local/evidence/windows-report10-2026-10-04/`；本轮所属容器和会话均已结束，无接管句柄，既有代理/teardown及大chunk警告保留。下一具体动作差异/链接复核后提交推送同一源码/脚本固定版本，用户运行`-Followup -BrowsersOnly -WebKitOnly`回传（正式构建且15/15）。UI冻结；Go/SDK/其他Windows引擎未选择不导入旧PASS，Windows修复实效、指针/绘制性能、原生/E08/外部缺口保持。

2026-10-03 第九轮 Windows 回传核对与补测改进完成：见[报告](../../acceptance/reports/windows-ninth-report-2026-10-03.md)，固定`3e6c31c`，14条摘要全部匹配；WebKit14/15，无skip/retry/flaky，四项导航/日志用例各3次通过，前轮两项超时已获真机通过。剩第一次fallback/worker终端45秒总时限耗尽，之后18.810/18.762秒通过，诊断无页面/请求错误，首次刷新约32.1秒才开始；上传不能确认具体慢动作/Windows因果。新增独立production build阶段和fresh preview，缺失/失败构建不计成功，禁止端口回退/旧服务器复用；保留原UI动作/45秒限时/5秒恢复/worker/ACK/输入不重放，实时终端阶段和有界静态资源耗时加入失败诊断，产品/UI未改。相同隔离冷启动终端单例各1/1，开发模式case21.625秒、正式14.766秒，不冒充Windows因果/E08验证；类型构建3740模块/Vite3.10秒通过。正式WebKit36/36（181.604秒）、Firefox12/12（67.804秒）、Chromium12/12（55.533秒）全无skip/retry/flaky；最终PowerShell harness、实际JSON标题/次数/返回14/15失败判定及诊断解码复制通过；真实CLI --list严格五标题各3次15项，非另一次运行PASS。诊断探针首次因嵌套配置cwd错误未启动，单独保留setup JSON；仅改本地探针cwd后故意断言失败0/1，1401字节诊断包含两阶段及五资源耗时，正确不计产品PASS。证据`.local/evidence/windows-report9-2026-10-03/`，本轮所属容器和测试会话全部结束，无接管句柄；替身/teardown代理及大chunk警告保留，单测无生产改动未重复。文档链接664条无断链、差异检查通过；交付固定同一提交源码/脚本，下一验证动作是用户Windows运行`-Followup -BrowsersOnly -WebKitOnly`先构建且严格15/15回传ZIP。未选择Go/SDK/其他浏览器旧PASS不导入，Windows实效及原生/E08/外部环境缺口未关闭。

2026-10-03 第八轮 Windows 回传修复与本地验证完成：见[报告](../../acceptance/reports/windows-eighth-report-2026-10-03.md)，ZIP SHA256 `92da4d5f6239347e11e8e13fdeb0ffec61de5c1e842a0a8eb2ea38d4f2eae467`，15条摘要全部匹配，测试源码 `57d4fb6e18977f78c92015a00df45cdf0c92e612`。仅WebKit13/15，无skip/retry/flaky；原fallback/worker终端3次通过。剩一次任务汇总恢复5秒超时及一次日志待取消读取建立超时（44.405秒/原45秒限时，首日志读取36.371秒才开始，诊断visible/focused但paint=false），不是同一个终端fetch错误。禁止帧回调的可控基线0/1（14.543秒）复现原5秒断言失败；正确只模拟timeout后的API基线11pass/2fail，初次同时模拟RAF的测试设置错误保留说明。共享安全读取现增加250ms无帧恢复兜底，pagehide取消并失效旧帧/旧截止，新导航重置代次；queued观察改为事件turn及microtasks，日志待取消读取由实际interval回调建立，不加sleep/force/retry/skip/限时。89单测（16文件、1.320秒）、类型构建（3740模块、Vite3.22秒）通过；WebKit36/36（226.946秒）、Firefox12/12（87.737秒）、Chromium19/19（103.257秒）加独立补验4/4（33.451秒），全无skip/retry/flaky，后4项为初次三个文件名模式未匹配而补验，未拼成单次全套。PowerShell harness、实际本地/回传JSON计数和失败诊断解码通过；真实CLI --list确认五标题各3次严格15个执行，不冒充已运行PASS。文档链接0断链、差异检查通过；证据`.local/evidence/windows-report8-2026-10-03/`，本轮所属容器/测试会话全部结束，无运行句柄接管，替身/teardown代理和大chunk警告保留。交付使用固定同一提交源码和脚本，下一条验证动作是用户Windows运行`-Followup -BrowsersOnly -WebKitOnly`回传ZIP核对两项超时修复；不导入未选择Go/SDK/其他浏览器旧PASS。UI冻结；有界兜底不证明所有慢导航/native dialog/bfcache序列，Windows实效、原生/E08/外部环境缺口保持。

2026-10-03 第七轮 Windows 回传修复与本地验证完成：见[报告](../../acceptance/reports/windows-seventh-report-2026-10-03.md)，ZIP SHA256 `aa2456fc3474ad6edc1ba58c0b147c2f80b5d1e5300bcfd2e503c695dbc8e3b5`，14条摘要全部匹配，测试源码 `8ad386a97cf214caa9a5eafa75d7a194e136ed87`。仅WebKit11/12（无skip/retry/flaky），三项任务查询导航各3次通过；剩余fallback/worker终端诊断来自InstanceConsole独立日志轮询，在beforeunload后9ms、pagehide前1ms。可控回调基线0/1、24.702秒，复现新的日志fetch；初步修复3/3、39.659秒，最终用例再验证无旧在途读取时的排队回调。安全API读取现共享导航取消/等待，仅导航中断的只读请求可在存活页面恢复，body取消不能变成空成功，pagehide阻止旧帧恢复，控制台销毁取消和防重叠；原输入/写入不重放断言保持。最终WebKit36/36（244.049秒）、Firefox12/12（89.141秒）、Chromium23/23（142.687秒，含Docker/批量/文件/保存/系统任务）无skip/retry/flaky；87单测（1.11秒）、类型构建（3740模块、Vite5.36秒）通过。PowerShell harness首次抓到旧9项断言后同步为10项，最终全通过；真实浏览器JSON标题次数及基线诊断解码通过，实际CLI --list确认缩小选择5项各3次15个执行，清单不冒充新运行PASS。证据`.local/evidence/windows-report7-2026-10-03/`；所有本轮测试会话和所属容器已结束，无句柄接管，Vite替身/teardown代理警告保留。下一动作差异/链接审阅后提交推送固定源码与脚本，用户运行`-Followup -BrowsersOnly -WebKitOnly`回传ZIP；未选择Go/SDK/其他浏览器不导入旧PASS。UI冻结，Windows修复实效、完整平台/E08及外部环境缺口保留。

2026-10-01 第六轮 Windows 回传修复验证完成：见[报告](../../acceptance/reports/windows-sixth-report-2026-10-01.md)，原 ZIP SHA256 `3f42d2492cea0b700c1922a4b42ba9d2d0fd976779d1b9044a83d8e99d3b638a`，20文件校验值全部匹配；测试源码 `f98b078c63df8f7ba3f5b9ce2cc3a84018c3fb65`、followup-browser。Chromium/Firefox各8/8，WebKit23/24，无skip/retry/flaky；两项查询导航检查各在WebKit3次通过。剩第一次fallback/worker终端执行的 `/tasks` 诊断发生在beforeunload后8ms、pagehide前8ms，未记录新HTTP/native exception。可控真实QueryObserver排队回调复现既有取消后仍发起summary/list新请求，基线0/1（12.295秒）；新增只读任务查询门禁，存活页面绘制/pageshow恢复，pagehide取消等待，修复后单例3/3、组合Chromium9/9（78.202秒）、Firefox9/9（100.685秒）、WebKit27/27（267.153秒）通过。共享批量/上传续传2/2（28.094秒），SystemApp防火墙任务状态接入同一取消族后相关状态/确认/恢复与不可用能力2/2（33.824秒），80单测/最终类型构建通过。最终PowerShell harness及真实JSON解析/基线诊断解码通过；实际CLI --list验证剩余4项各3次精确12个执行，没有把选择清单当测试PASS。新增 `-Followup -BrowsersOnly -WebKitOnly`，不导入已通过Go/其他浏览器，不放宽原错误断言/跳过/重试。证据`.local/evidence/windows-report6-2026-10-01/`，本轮测试会话及所属容器均结束，无句柄接管；Vite替身/teardown代理警告保留。下一动作最终差异/链接审阅后提交推送固定源码与脚本，用户Windows运行缩小补测并回传ZIP核对修复实效；UI冻结，完整平台/E08及外部环境缺口保留。

2026-10-01 第五轮 Windows 回传修复完成：见[报告](../../acceptance/reports/windows-fifth-report-2026-10-01.md)，原 ZIP SHA256 `bae0894c4730ae6e76c0502d75fb7fc82f196d19d772bbb680640fb54c31b810`，36文件校验值全部匹配，测试源码 `3c7abb3ef89f30d4bebd366019a72fe63c7f7eea`。11 必需原生检查通过，三项 Go 定向回归各3次通过，Master/runlog Windows race87pass/0fail/4skip；Chromium/Firefox各6/6，WebKit16pass/2fail，失败仅第三轮两种终端配置的 `/tasks/summary` 访问控制诊断，账号全部通过。可控旧请求阻塞再立即刷新复现读取未取消，修复前0/1、后WebKit3/3；任务汇总/列表/详情/阶段与共享实例/文件/Compose输出查询接入signal，在beforeunload/pagehide仅取消只读查询。最终Chromium8/8（79.608秒）、Firefox8/8（115.302秒）、WebKit24/24（263.222秒）无skip/retry/flaky；批量部分接受与文件续传另2/2，80单测/类型检查/生产构建、PowerShell harness及真实8/8/24 JSON解析/诊断解码通过。新增 `-Followup -BrowsersOnly`，不重复已验证Go/GCC/SDK，报告显式记录未选择项且不导入旧PASS。本地Vite teardown仍有代理警告，Windows原报错根因及修复实效尚待真机补测；完整Windows/E08及外部环境缺口保留。全部本轮测试会话及所属容器已结束，无运行句柄；差异和文档链接校验已通过。交付固定提交版本的GitHub脚本与源码，下一条产品验证动作是用户在Windows运行 `-Followup -BrowsersOnly` 并回传ZIP；继续核验实际源码、每个场景次数及失败诊断。UI冻结。

2026-10-01 第四轮 Windows 回传与修复：见[报告](../../acceptance/reports/windows-fourth-report-2026-10-01.md)，原 ZIP SHA256 `9199f2b17f927b533040a4ca89a86e9db6db55aaa12c6641ab4458845437b213`、源码 `08db099`。27必需Go/11原生全过，Chromium/Firefox各32，WebKit29pass3fail，全race356pass2fail14skip。阻止未派发客户端上传取消误入通用调度；日志辅助退出后重读同birth/token完成记录，未知/失败仍拒绝；真实 stdin 结果检查等待进程退出，Linux ESRCH 作为已退出处理。三项定向race8轮通过，Master/runlog完整race通过（426.988/2.542秒）。浏览器保留全部断言，补独立通知流、正确汇总数据、上下文级HTTP替身、账号正常按钮前景/滚动检查及失败诊断。初次并行本地WebKit16pass2fail/Firefox5pass1fail保留；最终分别复验Chromium6/6（59.158秒）、Firefox6/6（66.321秒）、WebKit18/18（169.378秒），无skip/retry/flaky。Go vet、两个修改模块Windows交叉编译、前端类型检查及PowerShell脚本harness通过；脚本也已读取实际三引擎JSON与两份失败诊断，验证每项执行次数及附件解码。本轮测试会话和容器均已结束，无运行句柄接管。本轮修复交付使用固定提交版本 -Followup 命令，下一条验证动作是用户在Windows运行补测并回传 Blora-Windows-Report.zip，不再原样重复第三轮全套。UI冻结，E08/完整平台及外部环境缺口保持。

2026-10-01 交付同步：修复代码、验收记录及新版脚本已提交并推送 `08db0995e6221933b77e202834d6a187a11649b1` 到 `BloretCrew/BloraPanel/main`，推送退出0。本轮 Windows 命令固定该提交版本，使用 `-Retest -InstallRaceCompiler`；不需要用户重复旧版原样全套。下一动作等待用户在 Windows 上运行新版脚本并回传 `Blora-Windows-Report.zip`，逐项复核真实结果；无本轮运行测试句柄。后续本地可复现的新失败继续处理，不把 Windows 编译、Linux 回归或脚本 harness 代替真机通过。

2026-10-01 第三轮 Windows 本地修复与回归完成：见[修复报告](../../acceptance/reports/windows-report-fixes-2026-10-01.md)。目录属性/ConPTY 标准句柄、可移植真实子进程夹具、Windows 权限与锁定根边界、ZIP 膨胀/协议截止夹具、Monaco 身份编辑/历史游标、浏览器 UA 键位、扩展即时检查点/真实焦点、网络结果未确认均已处理。新增实测连接 EOF 分类及查找栏延迟抢焦点修复；完整 Go race 退出0（Master296.382s）、四项定向race连续三轮通过；最终 Unicode/多光标/查找/撤销场景 Chromium/Firefox/WebKit 各3/3（34.1/41.6/43.5s）；前序九文件 Chromium33/33、Firefox32pass/1nativeIMEskip、WebKit31pass/1fail/1skip及后续修复单独记录，不拼成虚构全套结果。最后80单测/构建、Go vet、四包Windows交叉编译、两次PowerShell harness与文档链接/差异检查均通过；所有本轮测试会话已退出，无运行句柄。新版 -Retest 选27个Go回归、11必需原生项、每引擎32个浏览器场景，并用哈希验证的便携GCC补原来阻塞的全race。下一动作提交推送并交付固定提交版本命令，等待用户Windows真机ZIP；Windows目录属性/ConPTY及新夹具运行仍未验证，不提升整个平台验收；UI冻结/E08和外部环境缺口保持。

2026-09-30 第三轮 Windows 报告：见[详细记录](../../acceptance/reports/windows-third-report-2026-09-30.md)。Go/SDK 打包执行成功；Chromium120/120，Firefox116通过3失败1跳过，WebKit112通过7失败1跳过；全Go事件330pass/22fail/15skip，原生关键9/10仍ConPTY失败。Compose检查点/容器日志归档通过，目录metadata仍Access denied，不能称修复通过。GCC缺失race阻塞。下一动作本地定位目录metadata权限、编辑器恢复和Windows夹具等后再定向复测，不再让用户无变化重复44分钟全套。

2026-09-30 第二轮 Windows 报告已分析，见[报告](../../acceptance/reports/windows-second-report-2026-09-30.md)。Go PATH 缺失导致原生/全 Go 未运行，SDK WASI 包失败；浏览器 Chromium118/119、Firefox115/119、WebKit111/119，进度与 ZIP 已恢复。修正 Go 探测刷新/缺依赖记录、字号断言舍入、CDP IME 与跨浏览器恢复测试分离；PowerShell harness 与字号 Firefox 定向通过。仍须核查编辑器恢复、WebKit 超时/任务请求、ConPTY 等；不得算 Windows Go 修复复测通过。

2026-09-30 新版脚本交付验证：PowerShell 7.4 隔离容器 harness 退出 0，覆盖 remaining 模式记录、参数、损坏进度宿主、失败/跳过/截止、三类报告结果、ZIP 边界和旧报告恢复。实际 Windows 行为仍待回传；本次不把脚本测试算作 Windows 产品通过。

2026-09-30 用户要求提供剩余项新版脚本：新增 `-Remaining`，保留 SDK 打包、Windows 原生、全 Go 回归/可用 race、三浏览器；省略上一轮已通过的独立构建/vet/Web 单测，报告记录模式和覆盖缺口，不导入旧 PASS。仍包含已知未解决测试以收集现场，不宣称全量修复。上一轮 Git 暂存自动审批两次超时，修复尚未同步 GitHub；本轮正在重新验证与交付。

2026-09-30用户Windows报告已读取：见[首轮报告](../../acceptance/reports/windows-first-report-2026-09-30.md)，原ZIP hash记录，未公开私有日志。原生关键9/10通过，ConPTY失败；全Go316pass/26fail/14skip事件，containers发生runtime.semawakeup；Web76单测/生产构建通过、浏览器未运行、GCC缺失。修复SDK fileURLToPath与先生成默认/v2/v3包再Go测试的顺序、Windows CLI句柄重复关闭、metadata缺READ_ATTRIBUTES、日志目录sync平台适配。Linux定向三包通过，三包Windows测试交叉编译、实际参考扩展打包及PowerShell harness通过；所有本轮会话结束。Windows重新执行仍未验证。下一动作审阅提交这些确定修正，继续ConPTY标准句柄/启动现场、Master /bin/sh夹具移植、Windows权限断言/根目录重命名/ZIP膨胀夹具及协议20ms初始传输测试；不能让用户重复原样全套并把同一失败当环境问题。

2026-09-30 Windows脚本故障修复：用户截图显示Chromium下载完成后Windows PowerShell Write-Progress抛IndexOutOfRangeException，且Compress-Archive内部同类调用阻止报告ZIP。移除父脚本Write-Progress，子shell设置SilentlyContinue，改15秒普通文本状态及.NET ZIP；新增CollectLatest/CollectRun仅收集原日志、明确不重跑/不提升验收。新增模拟损坏进度宿主、子进度抑制和恢复打包测试；Windows实机原结果仍需用户回传，不能将脚本修复当产品Windows验收通过。

2026-09-29 Windows远端自助验收入口：用户有Windows设备但不开放连接，新增`scripts/windows-validation.ps1`及[运行说明](../../operations/WINDOWS_VALIDATION.md)。独立GitHub拉取/commit记录，可选winget工具安装，原生必需测试PASS事件校验、全Go/可用GCC的race、SDK/Web构建及三引擎浏览器，实时双流日志、阶段截止、失败继续、结果/skip/gap/ZIP报告。清除继承的BLORA环境防止连接旧Engine。Playwright mock配置不再在Windows写死Linux Chromium路径，runner强制新Vite服务。官方PowerShell7.4隔离容器语法/字面参数/失败/skip/超时/三类汇总/ZIP排除私有目录测试通过；首轮测试harness的Start-Sleep参数形式错误已修正，非Windows实机失败。`npm run check`及新服务模式真实浏览器终端代次用例1/1（7.9秒）通过。Windows实机仍待用户回传ZIP；脚本明确不覆盖服务/任务变更生命周期、系统通知中心、Linux依赖的真实浏览器fixture、物理故障/E08/远程Engine等，不提升整个平台验收。无本轮运行句柄，下一动作提交推送并交付运行命令。

2026-09-29 RC34最终：完整真实功能51/51通过897.040秒（session90325退出0），首轮RC33的49/51失败保留；三浏览器终端组合各4/4、六包2,251条记录/三份Web、停机恢复及兼容RC33回退均通过，见[报告](../../acceptance/reports/terminal-generation-2026-09-29.md)。修复旧连接异步错误污染新连接；未证明原Vim偶发/撤权超时根因。玻璃遮挡裁剪像素不一致，未采用。fixture由trap停止，诊断Vite78601已Ctrl+C退出130，无本轮活跃测试句柄。下一动作审阅文档/提交推送；严格E08与Windows/跨主机/物理故障等未通过项继续保留，UI冻结。

2026-09-29 RC34最新接管：三引擎终端组合各4/4、六包构建/2,251条记录/独立恢复及兼容RC33回退均已通过，详见[报告](../../acceptance/reports/terminal-generation-2026-09-29.md)。完整真实51项session90325仍确认存活，已进入第35项121点监控；前轮两处扩展失败本轮均通过，尚待全套终态。日志.local/evidence/rc34/real-full.log，附件/tmp/blora-rc34-real-full；不可重启替代原会话。玻璃遮挡裁剪诊断已退出0，但连续原版截图零差异、候选35,344像素改变（455像素通道差>8），不采用、不计性能改善；日志web/.local/evidence/rc34/opaque-control.log，图片/tmp/blora-opaque-{before,control,after}.png。诊断Vite session78601待停止，产品未加入裁剪。下一动作收取完整回归、补验收与剩余清单并提交推送；严格E08/原偶发/外部平台缺口保留。

RC34续接：RC33完整51项session47037已退出1，49通过/2扩展失败，879.929秒，所属fixture结束；新增旧代次竞态与原终端三配置4/4通过46.248秒，72520已结束。现在RC34包构建session6350、Firefox4项session14140，日志.local/evidence/rc34；须先收取，再串行WebKit、六包校验/恢复及修复后完整51项回归。产品仅TerminalApp连接代次保护，UI未变；不得把首轮49/51写成通过。

2026-09-29 RC34进行中：实际复现旧连接解析失败晚返回会禁用新连接，terminal-generation首次失败25.664秒、代次校验修复后通过12.250秒；当前补当前连接错误继续禁用输入，与原terminal.spec三配置组合session72520运行。证据web/.local/evidence/rc34，见[报告](../../acceptance/reports/terminal-generation-2026-09-29.md)。RC33完整51项session47037仍在运行，已有扩展通知/名称保存两项失败，保留原附件；测试补既有真实hover命中前置条件待复验。产品源码已改但web/dist仍是RC33，必须等原全套结束后再构建RC34并验证修复，不能混算版本。新竞态未被认定为原Vim偶发根因。

2026-09-29恢复本地任务：起点3f0f9eb，RC33当前源码完整真实51项回归已启动session47037，日志.local/evidence/rc33/real-full.log，附件/tmp/blora-rc33-real-full。必须收取同一会话终态，不能用旧RC26通过代替。本轮开始新的受阻审计计数，不沿用此前blocked。源码审查发现TerminalApp旧代次异步失败可改写当前连接状态，新增terminal-generation.spec.ts可控延迟复现，尚未运行/确认，不宣称原Vim根因；产品源码未改。等待Docker相关测试完成后再运行该浏览器探针，避免已知网络变更干扰。UI冻结及严格E08阈值保持。

2026-09-29 RC33后复核：产品HEAD dfb0c14、工作树起始干净；核对发行/终端/矩阵/平台报告后，将[剩余工作页](../NON_UI_REMAINING_2026-09-26.md)整理成当前清单，旧过程保留本文件及Git历史，避免旧RC23/长测RUNNING等条目误导执行。再次确认本机无/dev/dri，未发现QEMU/KVM入口；没有新确定缺陷、可区分性能方案或活动测试句柄。本次仅文档整理，不重复构建、不提升验收状态。下一动作校验链接并提交推送；后续需要新失败附件/性能假设或对应隔离环境，严格E08及未定位偶发保留。

2026-09-29 RC33最终：OSC 8链接持久恢复修复完成，Chromium61/61、Firefox/WebKit各50/50、76单测及真实双节点PTY7/7（74.004秒）通过；RC33六包2,243条记录/三份Web与独立恢复/兼容RC32回退通过，见[报告](../../acceptance/reports/terminal-links-recovery-2026-09-29.md)。所有本轮会话已结束，fixture/浏览器容器/发行冒烟所属进程均已清理，无运行句柄接管。下一动作提交推送源码与验收记录；后续仅对有新证据或可区分假设的剩余缺口继续定位，不把重复通过视作原Vim偶发根因修复，不无目的扩展VT实现范围。严格E08、本机缺硬件图形设备、Windows/跨主机/物理故障等缺口仍保留，UI冻结未改。

RC33后续收取：Firefox50/50（135.929秒）、WebKit50/50（129.442秒）、六包构建55.934秒、2,243条记录和三份Web校验5.536秒、独立恢复及兼容RC32回退9.713秒均通过。原27764/21367/59520/40197/84901已结束；当前仅真实双节点文件/PTY组合session61149运行，日志.local/evidence/rc33/real-terminal.log，附件/tmp/blora-rc33-real。先收取该句柄，再更新最终证据和提交推送；不重复已通过检查。

2026-09-29 RC33进行中：起点9ff9887，实际复现OSC 8链接跨检查点丢失，新增经完整验证的稀疏链接/当前属性记录及原生行标记恢复，同步Worker。Chromium61/61、单测76/76、实际鼠标连续两次恢复点击通过，见[报告](../../acceptance/reports/terminal-links-recovery-2026-09-29.md)。构建session27764、Firefox组合session21367运行中，日志.local/evidence/rc33；先收取同一会话终态，再串行WebKit和真实双节点PTY、六包校验及兼容RC32回退。前序本轮Chromium/单测会话均已结束；UI未改，原Vim/E08/外部平台缺口保留。

2026-09-29 RC32已验证：起点4aec17f，确认主题切换后旧调色板跨刷新复活，新增受校验的重置序号元数据，与输出/挂载串行并在恢复正确位置直接调用原生主题setter，不插入控制字节。第一修复因对象引用相同未触发setter仍失败，改用等值新对象后原两项通过；73单测/类型、Chromium48项、最终CSI/OSC交错六项及非法标记拒绝1项通过。真实PTY7/7、RC32六包2,239条记录/三份Web、独立恢复/兼容RC31回退通过；Firefox/WebKit最终各40/40通过107.223/108.485秒。所有本轮测试/构建会话已结束，fixture及发行冒烟所属进程已清理，无活跃句柄待接管。证据web/.local/evidence/rc32与.local/evidence/rc32，见[报告](../../acceptance/reports/terminal-theme-order-2026-09-29.md)。下一步提交推送，再独立核查链接等未覆盖终端状态；UI未改，原Vim/E08/外部平台缺口保留。

2026-09-29 RC31已验证：起点5fb2ea0，实际复现四种OSC自定义颜色跨检查点丢失，新增有界/严格验证的颜色差异记录，通过原生OSC恢复且不改主题默认值。三浏览器各33项（Chromium71.933秒、Firefox83.659秒、WebKit90.427秒）、73单测、真实双节点PTY7项77.046秒通过；RC31构建48.723秒、六包2,235条记录/三份Web校验5.342秒、独立停机恢复/兼容RC30回退12.244秒通过。证据web/.local/evidence/rc31及.local/evidence/rc31，见[报告](../../acceptance/reports/terminal-colour-state-2026-09-29.md)。所有本轮句柄已结束、fixture/发行冒烟所属进程已清理，无活跃会话接管。下一动作提交推送本轮修复与证据，再独立验证用户切换主题与未完成远端指令交错及链接状态；UI未变，原Vim/E08/外部平台缺口保留。

2026-09-29 RC30最终收取：WebKit session85510退出0，27/27、74.612秒；三浏览器各27项、71单测、类型/生产构建、真实双节点7项复核和RC30六包2,231条记录/恢复/兼容RC29回退均完成，见[报告](../../acceptance/reports/terminal-protocol-state-2026-09-29.md)。所有本轮运行句柄已结束，fixture与发行冒烟所属进程清理完成，无活跃会话待接管。源码/UI未在验证后改变；下一动作提交推送本轮修复，再继续独立检查链接/自定义颜色等持久终端状态。首轮权限撤销超时原因未定位，原Vim偶发、严格E08和外部平台缺口保留，不称全量完成。

RC30续接最新状态：真实双节点串行复核session69918退出0，7/7、73.001秒；独立发行恢复/兼容RC29回退79184退出0，13.567秒；Firefox7416退出0，27/27、76.149秒。当前仅WebKit同组27项session85510运行中，日志.local/evidence/rc30/webkit-protocol.log，附件.local/evidence/rc30/webkit-results。先收取该句柄终态，失败则保留附件定位；通过后补齐矩阵/报告、审阅并提交推送。本轮UI未动，首轮真实权限撤销超时原因仍未定位，不能将复核通过说成根因已修复。

RC30中间收取：构建session60229退出0（51.135秒），六包校验40790退出0（4.759秒）。实际PTY首轮14424退出1，6/7通过、权限撤销项60秒总超时，清理阶段headers获取会话报错，原因未定位；旧日志和/tmp/blora-rc25-real-results保留。当前原断言/超时串行全组复核session69918，日志.local/evidence/rc30/real-terminal-repeat.log，附件/tmp/blora-rc30-real-repeat。先收取该句柄，再启动Firefox/WebKit容器及独立发行恢复，避免增加无关并发干扰。

2026-09-29 RC30进行中：起点200dfab，四项实际协议查询复现鼠标SGR/像素编码、光标隐藏/闪烁状态跨检查点丢失，新增可选协议状态并使用原生setter恢复，保留程序覆盖与用户选项的区别。25项初测、27项完整控制状态/实际鼠标点击比较、71单测及类型检查通过。当前真实双节点PTY session14424、RC30构建60229运行中，日志.local/evidence/rc30；须收取原会话再进行跨引擎与六包校验、停机恢复/兼容RC29回退。浏览器证据web/.local/evidence/rc30。见[报告](../../acceptance/reports/terminal-protocol-state-2026-09-29.md)。UI未改，原Vim偶发与E08等缺口未关闭。

2026-09-29 RC29最终收取：WebKit原用例/断言全组复核19/19通过51.653秒，session37094已退出0；首轮模块加载失败保留且未再次出现。71单测、Chromium30项及持久重开19项、Firefox19项、真实双节点PTY7项和RC29六包/恢复回退均通过，详见[报告](../../acceptance/reports/terminal-control-state-2026-09-29.md)。所有本轮浏览器、fixture、构建和发行验证会话已结束，无运行句柄待接管。下一动作：提交推送本轮源码与验收记录；继续审计终端尚未覆盖的持久协议模式/链接/颜色等状态，不把本轮四类修复推广为全部VT正确。严格E08、原Vim偶发和外部平台缺口继续保留，UI未修改。

RC29后续收取：Firefox19/19通过50.720秒；六包2,227条记录及三份Web校验5.519秒，独立停机恢复/兼容RC28回退10.955秒通过，原60422/22569/51357已结束。WebKit首轮18/19，备用屏幕项动态模块加载失败、尚未执行终端断言；原64222已退出1。现按原断言重跑同组session37094，日志.local/evidence/rc29/webkit-terminal-repeat.log，附件持久目录.local/evidence/rc29/webkit-results。须先收取该句柄，不重启活跃任务；如再次失败检查持久trace定位加载原因。

2026-09-29 RC29控制状态恢复进行中：起点93ccace，实际复现字符集/滚动区域/保存光标/制表位四项检查点后丢失，新增版本化controlState并恢复到主终端与Worker；全对象先验证，旧检查点兼容。30项浏览器组合、从持久存储重开19项、71单测、真实HTTPS双节点PTY7项通过，类型及RC29六包构建62.202秒通过。证据见[报告](../../acceptance/reports/terminal-control-state-2026-09-29.md)，浏览器/单测web/.local/evidence/rc29，后端/发行.local/evidence/rc29。当前Firefox session60422、六包校验22569、独立恢复回退51357已启动，须收取同一会话终态；之后串行WebKit，补最终证据并提交推送。此前所有本轮Chromium/真实fixture/build会话已完成。UI未变，原Vim偶发、严格E08和外部平台缺口不关闭。

2026-09-29 RC28终态（覆盖下条进行中记录）：Firefox/WebKit真实解析器组合各21/21通过（92.575/79.382秒），正式仓库配置入口各2/2通过（11.590/12.498秒）；69单测、Chromium24项加晚返回1项、真实双节点PTY7项、类型及生产构建通过。RC28六包47.211秒、2,223条内部记录及三份Web校验4.332秒、独立SDK/双节点/停机恢复/兼容RC27回退10.055秒全部退出0，见[完整报告](../../acceptance/reports/terminal-parser-checkpoint-2026-09-29.md)。先前exec58587、74062、27039、34822均已完成，发行冒烟所属进程已停止，无这些会话需要接管。私有证据.local/evidence/rc28；UI与严格阈值未改。下一动作：提交推送本轮源码与证据，继续独立审计已完成VT指令的持久状态恢复；原Vim偶发及严格E08和外部平台缺口不关闭。

2026-09-29 RC28控制序列恢复进行中：起点76f8a55，真实复现半段CSI颜色指令跨检查点后显示为文字；当前修复仅在锁定xterm解析器GROUND时压缩，未完成序列保留有界原始日志，日志支持尺寸事件且严格校验；Worker旧快照返回前若序号/解析状态变化则拒绝覆盖。初始控制序列9/9、相关浏览器24/24（90.286秒）、额外真实Worker晚返回1/1（6.487秒）、前端69/69、生产构建及真实HTTPS/双节点文件PTY7/7（74.241秒）通过。现在Firefox真实解析器/UTF-8/Worker组合在exec58587运行，日志.local/evidence/rc28/firefox-terminal.log；最终类型检查exec74062。下一动作：先收取这两个同一会话终态，再串行补WebKit同组合，按当前源码构建RC28、六包摘要/三份Web核对和兼容RC27恢复回退；不能把未完成运行或RC27旧包计入本轮。UI/通透效果/性能阈值未变，原Vim额外字节仍未定位。证据.local/evidence/rc28，所有此前本轮fixture已结束。

2026-09-29 RC27终端UTF-8恢复修复：起点d8c5fd8，真实复现半个“中”经检查点重建后丢失（期待“中”、实际空串），新增可选最多3字节的pendingUtf8续接状态，在原始输出日志前恢复；旧记录兼容、非法续接拒绝，UI未改。六个分割位置与既有Worker/查询共9/9、另检查点+日志混合1/1、前端68/68、实际双节点文件/PTY7/7及类型/构建通过。RC27六包49.255秒、2,219条记录/三份当前Web校验5.051秒、独立SDK/双节点/停机恢复/兼容RC26回退10.024秒通过，见[修复及发行报告](../../acceptance/reports/terminal-utf8-checkpoint-2026-09-29.md)。私有证据.local/evidence/rc27，早期失败日志在web/.local/evidence/rc26；所有fixture/浏览器/发行冒烟已结束。下一动作：提交已验证修复与记录，然后独立审计未完成VT控制序列的检查点恢复，不能把UTF-8通过推广为全部解析器状态正确。原Vim偶发、严格E08及Windows/跨主机/物理故障缺口不关闭。

2026-09-29 RC26最终功能与平台补验完成：完整HTTPS/双节点/Docker功能51/51通过，905.725秒，原exec44632已退出0；新Vim导航延迟2秒诊断5/5及本次完整套件通过仍未定位原偶发，临时诊断路由已移除。systemd/cgroup补充32MiB组内OOM与8进程上限实际触发，最终5项通过，9.652秒，所属组/容器清理成功。Linux私有Xvfb/DBus/Dunst原生通知实际绘制、系统鼠标点击正确任务、文字、刷新去重和退出后用户关闭通过，最终9.666秒；早期端口不支持及“退出自动消失”预期失败分别保留，Dunst实际保留已投递弹窗，不声称退出后重启跳转。见[平台报告](../../acceptance/reports/systemd-cgroup-2026-09-29.md)和[原生通知报告](../../acceptance/reports/native-notifications-2026-09-29.md)。类型检查、Python/Shell语法及差异检查通过。产品/UI未再改，RC26不因测试/文档更新重复打包。私有证据`.local/evidence/rc26`，全部所属fixture、浏览器和容器已结束，无运行会话接管。下一动作：最终审阅测试/验收记录后提交并推送；后续仅针对能取得新证据的Vim诊断或不降级UI的性能方案继续，不能将严格E08、Windows/其他桌面、跨主机、宿主级耗尽与物理故障缺口写为通过。

2026-09-29 RC26平台修复与交付完成：起点a1478ef，独立systemd容器发现并修复停止后卸载服务、已安装未启动timer从列表消失的问题；现合并单元并分页读取真实属性，别名及启用状态核对通过。真实systemd生命周期、cgroup出生归属/整组结束、限制文件/CPU限流3项通过，8.199秒；错误二进制明确拒绝。runtime/systeminfo race和Master/Daemon相关7项通过。RC26六包构建81.235秒、2,215条内部记录/三份Web校验4.694秒、独立SDK/双节点/停机恢复/兼容RC25回退10.320秒均退出0，见[平台与发行报告](../../acceptance/reports/systemd-cgroup-2026-09-29.md)。产物dist/releases/development-20260929-rc26，私有证据.local/evidence/rc26；所有所属测试容器/进程已结束，无会话需接管。UI及通透效果未变，未触碰宿主manager。下一动作：继续审计其余本地可运行平台边界及Vim偶发，保留E08、Windows、跨主机、设备故障和原生通知中心等实际缺口；不重复已无收益的软件合成试验，不宣称全量完成。

2026-09-29 RC25后诊断完成：产品与UI未修改。真实Vim/top连续10/10通过但原偶发未复现，新增失败时连接/方向/阶段/十六进制字节附件，严格输入相等断言不变；最终版本再复验3/3通过（36.582秒）。新增历史查询/本地raw日志回放零应答、live查询正确应答及只读抑制检查，连同现有终端Worker/回退/恢复定向6/6通过（77.347秒）；新检查独立重复3/3通过，最终类型检查通过。绘制线程跟踪明确软件合成瓶颈；显式启用SwiftShader合成的两个微探针更慢，均未采纳。最终60秒循环传输跟踪复核240样本p95 1062.5ms，双PTY/6次传输/目标校验正常，在原50ms断言失败；带跟踪数据不能替代RC25正式三引擎结果。见[诊断报告](../../acceptance/reports/local-diagnostics-2026-09-29.md)。私有证据 `.local/evidence/rc26` 只是本轮目录名，没有生成RC26发行包；所属fixture/浏览器均结束，无会话需接管。下一具体动作：如真实Vim再失败，从新附件判定旧连接/回放/恢复后查询来源后修复；在可用硬件合成设备补测冻结UI严格E08，不重复已无收益的软件后端试验。Windows/systemd/跨主机/物理故障等环境缺口保留，不宣称全量完成。

2026-09-29 RC25本地收尾已交付：起点 `658c521`，完整保留通透与选定UI；恢复快照重复复制已移除，重复聚焦100次不再产生100次同步存储写入。66单测、59/59完整mock、原生IDB三引擎各1项与Firefox/WebKit真实恢复故障各3项通过。真实完整套件50/51，Vim刷新偶发额外字节在诊断/原始用例各三次复核均未再现，来源仍未定位，首轮失败保留；参考包补已有原生hover命中等待后连续3/3及全mock通过。快照专项中位耗时改善35%～56%，但严格E08 Chromium/Firefox/WebKit p95为1003.9/1108/545ms，均失败，不启动条件不满足的一小时复验。RC25六包构建、2,207条内部记录/三份当前Web校验、独立双节点/停机恢复/兼容RC24回退全部通过，见[本轮报告](../../acceptance/reports/local-closeout-2026-09-29.md)。所属浏览器/fixture/容器/包测试均结束，无运行会话需接管；私有证据 `.local/evidence/rc25`，产物 `dist/releases/development-20260929-rc25`。既有24h已通过，未重复长测。下一动作：进一步定位Vim偶发应答来源及完整材质合成瓶颈，在具有硬件加速和Windows/systemd/跨主机/设备故障等隔离环境补验；未把全范围标为完成。源码、回归和验收记录提交推送到既有GitHub仓库。

2026-09-29 本地完成度复核：原先写作“运行中”的独立24h长测已从私有原始证据重新校验为 **PASSED_24H**，且[最终报告](../../acceptance/reports/local-endurance-24h-2026-09-27.md)与原始终态完全一致；实际86,415.480秒、17,262次采样、1,440个分钟槽、两次故障恢复、末尾真实备份恢复与所属资源清理通过。已同步[验收矩阵](../../acceptance/ACCEPTANCE_MATRIX.md)及[非UI剩余工作](../NON_UI_REMAINING_2026-09-26.md)，旧“RUNNING”条目仅为当时过程记录。本轮未改产品源码、未重跑已通过的功能或发行检查，当前仓库起点为 `6f48131`；无长测/fixture会话需接管。Linux本地功能51/51、mock56/56及RC24发行证据保留；严格E08当前UI三引擎p95 1240.1/1230/739ms，均高于≤50ms目标，本机无`/dev/dri`，不将旧UI或软件合成对照视为通过。下一具体动作：仅在能区分绘制热点且保持用户选定UI的条件下继续源码级性能优化并定向复验；取得可用硬件加速及Windows/systemd/跨主机/物理故障等隔离环境后执行对应真实验证。全范围仍未完成。

2026-09-28 README双语与首页排版：按用户要求将默认 `README.md` 改为英文，新增 `README.zh-CN.md` 双向语言跳转；使用现有品牌几何制作静态标志，增加技术徽章、随深浅色切换的真实界面截图、功能表、快速开始、架构与文档导航，高级配置/验收折叠保留。中文版保留原运行指南；其余工程文档仍为中文，英文页已明确说明。仅将两张既有演示PNG纳入Git，其他历史图库继续忽略，未使用图片生成或修改产品UI。全部本地链接存在，17个Shell示例与4个JSON示例语法校验、GitHub原生Markdown表格/picture/details渲染通过；实际GitHub匿名Chromium验证英文浅/深、中文版、390px手机四场景，3200px原图加载、语言链接及折叠比较均通过。首次深色图片等待超时、手机页面水合期间节点替换与整页load等待超时均保留为检查入口失败，改为等待实际README内容/图片后定向复核，未改变产品或验收阈值。首页更新已推送 `main`，提交 `43d80ea`；本轮是文档交付，不提升任何功能/性能/平台验收状态。

2026-09-27 GitHub 源码交付：按用户明确授权，将项目上传至公开仓库 https://github.com/BloretCrew/BloraPanel 的 `main` 分支，首次提交 `212a5ddc7ada7195c1b96d5206bca63eaf4c49d9`。共1,137个源码、测试、文档、锁文件及代码使用的图标资产，远端逐项核对路径、文件模式与blob摘要一致。完善 `.gitignore`，排除凭据、运行数据库、依赖、发行归档、生成的参考扩展包和历史截图；这些本地文件没有删除。README补齐首次克隆后的 `make sdk` 步骤，三个参考包可从源码生成。仓库Description由用户自行填写。产品UI及验收状态不变，独立24h长测和只读收取器继续运行；下一动作仍为收取实际24h终态并保留性能/外部平台验证缺口。

2026-09-27 11:14 长测18h阶段核对通过：第2次故障03:13:25 UTC完成；所属Daemon SIGSTOP后明确OFFLINE，Master重启前03:13:20/重启后03:13:21均核对原120点stale历史，同一Daemon恢复后原runId `7d43fbe2017b5c6b1aa1e017a24eb45c`仍RUNNING。最新Master为PID1372175/startTicks16367598、Daemon598122/startTicks9882575、worker597961/startTicks9882482，同bootId；任何接管仍应读最新checkpoint确认身份。当前实际64,869s/12,956采样/1,081成功slot，两个声明故障阶段完成，RSS Master/Daemon34.04/34.01MB、FD14/19、归档16,776,441B、状态/证据峰值205,456,312B、普通采样gap6.067s，继续使用原预算，不提前标24h通过。独立收取器1366399仍运行，将自动更新[终态报告](../../acceptance/reports/local-endurance-24h-2026-09-27.md)；预计17:12后由长测worker执行实际恢复与所属资源清理，收取器核验后结束。RC24与真实51项/完整mock56项功能证据已闭合；严格E08本机完整材质仍未达标，软件后端比较无足够收益，Windows等外部环境仍未验证。新收取入口不改变本次不可变二进制或产品UI。下一动作：收取24h实际终态及边界，保留性能/外部平台缺口；无需为结果收取重新启动服务或要求用户回复“继续”。

2026-09-27 11:08 结果自动收取：新增只读 `scripts/endurance-report.py`，当前真实短测终态正确识别为PASSED_SHORT；旧短测缺新清理字段明确拒绝，短测伪标24h/缺实际恢复/遗留服务/私有字段不导出四项拒绝边界验证通过。独立收取器PID1366399/startTicks16319189、bootId同长测，03:05:16 UTC启动，持续更新[24h动态报告](../../acceptance/reports/local-endurance-24h-2026-09-27.md)并在实际终态结束，不需要用户回复“继续”；日志和身份 `.local/evidence/rc24/endurance-collector.*`。报告不是预先通过声明，超时/检查点停更也以未完成退出且不改服务。最新长测17h54m、12,883采样、1,074成功slot，状态RUNNING，18h阶段约11:12，24h终态17:12后；实际disk峰值204,685,379B仍在原256MiB预算，SQLite任务/事件累积继续真实计入预算，未调整上限或删除历史。新无终端/存储活动的八玻璃微探针：SwiftShader p95 400ms、明确Mesa llvmpipe 433.3ms、Vulkan仍SwiftShader383.3ms，非E08验收，未采纳对照；Xvfb包装器未进入探针，直接带所属显示认证后的诊断37.830s完成；唯一容器已停止，包装器137不算通过。产品UI和RC24产物未变化。下一动作：核对18h故障后原runId与120点历史，持续保留24h真实终态及严格E08未达标；Windows等外部环境仍未验证。

2026-09-27 10:57 RC24发行闭合：六包构建退出0（49.479s），外部SHA256和2,191条包内路径/模式/长度/SHA256记录全部通过，两个Master及独立Web包精确匹配当前浏览器构建。独立SDK/参考包签名、Master/TLS/双Daemon、停机快照恢复及兼容RC23实际回退 **退出0（15.359s）**，所属进程已停止，见[RC24报告](../../acceptance/reports/release-rc24-2026-09-27.md)。初次打包因headless包缺LICENSE退出2，现严格绑定同源commit/MIT/原文hash补齐通知；失败记录保留。当前56项完整mock、另Worker终止1项、完整真实51项及分段Go race通过，不以这些功能结果覆盖E08。完整材质的禁用GPU启动参数诊断仍p95 1005ms、退出1，最终WebGL仍SwiftShader，不算得到独立CPU后端；产品未改。接下来仅私有隔离容器比较已有图形后端是否能改善合成；probe exec77336，不访问宿主设备/网络/凭据。独立24h `.local/endurance-zs_dzpzj` 最新elapsed63,438s/12,679采样/1,057成功slot，RUNNING；下一18h故障约11:12，24h终态约17:12，勿停止。正式服务和生产节点未触碰，Windows/systemd等外部平台仍未验证。

2026-09-27 10:43 本地完整功能回归闭合：当前56项全mock浏览器通过；最终原生鼠标hover命中探针后的两条扩展链路各重复3次共6/6通过，真实完整51项 **全部通过（809.066s，退出0，无skip）**，含121点监控/Master重启、两个真实分钟调度slot、真实Docker/Compose/扩展WASI、PTY全屏/刷新/撤权/移窗与16MiB原上传续接。新增真实终止运行中的检查点Worker后等待原生产30秒截止、中文UTF-8分块/屏幕保护/重挂不重放输入 **1/1通过（35.553s）**。详见[本地功能报告](../../acceptance/reports/local-functional-2026-09-27.md)。同材质独立绘制伪元素性能诊断仍p95 1003.8ms、退出1，无足够收益，不进入产品；当前三引擎E08失败继续保留。真实fixture9444 exec36317已Ctrl+C退出0，Vite5173 exec50786已停止，新Worker测试自管Vite也已退出；性能容器已`--rm`清理。独立24h监督 `.local/endurance-zs_dzpzj` 不停止，最新status仍RUNNING。下一动作：RC24六包构建、全部外部/包内摘要校验及独立停机恢复/兼容RC23回退，更新发行台账；Windows/systemd等实际运行仍未验证。

2026-09-27 10:30 功能复验进展：当前56项完整mock浏览器回归全部通过（365.482s、退出0）；绘制就绪两次RAF后仍使用可信鼠标，参考包实际WASI/丢202回执同幂等键恢复/快捷入口/移窗关闭及成员权限2/2（31.694s、退出0）通过。临时指针事件打印、body状态诊断与样式对照已删除，未合成click或重试副作用。首轮真实参考包未包裹绘制等待的开窗点击实际复现失败后修正；登录前一次ERR_CERT_VERIFIER_CHANGED单独保留。两轮短暂启动的完整套件分别因需独立性能入口及fixture已有备份导致隔离不足而中断，退出130，不能计通过；旧fixture均Ctrl+C退出0。现在唯一完整真实功能套件为exec52431（51项，性能严格另验），fixture9444 exec36317，私有凭据 `/tmp/fixture-1296441527/browser-credentials.json` 勿输出内容；Vite仍5173 exec50786。独立24h长测保持运行。见[本地收尾结果与失败记录](../../acceptance/reports/local-functional-2026-09-27.md)。下一动作：收取51项终态，继续非视觉合成成本对照（仅测试侧），通过后停止所属fixture并重新打包RC24、验证独立恢复与RC23回退；E08已知失败不被mock17ms覆盖。

2026-09-27 10:20 本地收尾更新：保持 UI 冻结，修复 Compose 保存流程先解析参数后授权导致无 requestId 的未授权请求返回400，现先查权限、执行时重查；真实 Docker 分阶段保存/Daemon重启/显式应用及未授权回归通过。修复按焦点顺序重排窗口 DOM 会重新加载扩展 iframe、丢本地输入的问题，改为稳定创建顺序和 CSS 层叠；新增实际输入/三次焦点切换回归通过。修复旧测试主题选择器、保存回执对账替身、异步重新读取正文确认后刷新时序，以及监控 fixture `--listen` PID 边界识别，未放宽产品合同。全 Go race 首轮其余包通过、Master两项失败；独立真实PTY重复10次通过，修复后完整Master race **通过（435.868s）**。实际浏览器51项首轮42通过9失败，修复后的12项定向复验 **10通过2失败**：备份、容器指标、扩展取消/通知/名称保存/迁移/目录、监控121点历史与Master重启、实例设置和编辑器均通过；扩展参考包鼠标点击及其清理失败引起的成员409仍在定位，不能据定向通过声称完整回归通过。新版前端mock25项22通过3失败，焦点新增项通过；现用系统和缓存浏览器对照确认剩余 iframe 真实点击未进入子文档，移除诊断前需取得复验。当前服务：Vite5173 exec50786；真实fixture9444 exec66125，私有凭据路径 `/tmp/fixture-2906434352/browser-credentials.json`（勿输出内容），只能清理所属fixture。独立24h长测 `.local/endurance-zs_dzpzj` 已17h04m、1025成功分钟slot、12287采样，仍RUNNING；预计17:12后收取，勿随构建/fixture清理停止。各本轮命令实际终态保存在私有 `.local/evidence/rc24/*.log.result.json`。下一动作：定位 iframe 命中/点击问题，完成全套前端与真实浏览器回归，再RC24打包和独立恢复/RC23回退；严格E08三引擎失败及Windows等平台未验证继续保留。

2026-09-27 09:43 续接本地收尾：独立长测已连续16h24m、跨UTC日期，984个分钟备份slot、11805采样，6h失联/同库Master重启后保持原runId；仍RUNNING，预计17:12后收取，不能提前算通过。当前Master已由长测监督器重启为1020809，身份必须使用status读取。昨日tool会话失效，未取得完整回归终态，不据文件时间戳计作通过。今日加入私有磁盘日志/真实退出码并重跑：`make check build windows`退出0（4.90s）、前端63/63、干净SDK与三个参考包退出0。前端55项首轮受5173服务中断影响已取消，持有独立Vite后复验；发现旧`主题`测试定位匹配两个控件，修正选择器，结束后定向复验。Go首轮缺少显式exec-helper测试配置，取消后以新构建helper重跑全套race（exec48147）。真实51项串行浏览器已启用监控121点soak（exec46948），fixture为9444、`/tmp/fixture-172569861`、exec16295；Vite exec50786、前端exec26075。昨日性能当前完整材质基线三引擎均失败：Chromium1240.1ms、Firefox1230ms、WebKit739ms；测试侧去模糊/分层对照不进入产品，见[当前性能报告](../../acceptance/reports/performance-current-2026-09-26.md)。下一动作：修复回归发现并复验，完成RC24打包与独立恢复/回退冒烟，继续保留严格E08失败和24h进行中。

2026-09-26 本地收尾更新：Firefox/WebKit 恢复故障各 3/3 通过；新增真实墙钟监督入口，两轮完整五分钟短测及 stop/resume/监督中断安全清理通过，见[长测入口报告](../../acceptance/reports/local-endurance-2026-09-26.md)。独立 24h 运行 `.local/endurance-zs_dzpzj` 已从北京时间 9/26 17:12:32 开始，预计 9/27 17:12 后收取终态；必须用 `python3 scripts/local-endurance.py status .local/endurance-zs_dzpzj` 读取最新身份，不能在运行中记为通过。终端序列化 Worker 已实现，真实 xterm 一致性/刷新/移窗/不可用回退浏览器 4/4 通过；单独迁移的 Firefox 混合负载 p95 1227ms，仍失败。新增断言实际复现终端日志数组嵌套 Vue 代理造成 `DataCloneError`，修正为校验后原位追加，63/63 前端单测及生产构建通过；严格负载正在复验。UI、负载、恢复 ACK 顺序及 ≤50ms 阈值保持不变。性能 fixture 仍为 exec 48529、9443、`/tmp/fixture-872806179`；Vite 测试 exec 34771、5173；只 Ctrl+C 清理本任务服务，独立长测保留。下一动作：收取性能结果，必要时继续定位，再完成当前全套回归、RC24 构建及独立发行恢复/回退冒烟。

2026-09-26 本地收尾执行中：用户已授权完成本地可推进工作，保持 UI 冻结。Firefox 155.0 和 WebKit 26.6 的恢复故障三项各 3/3 通过，见[跨引擎报告](../../acceptance/reports/recovery-engines-2026-09-26.md)，不代表真实物理掉电或浏览器实际配额阈值通过。当前冻结 UI 的 Firefox 严格双 PTY 60 秒基线 p95 1519ms，失败证据位于 `/tmp/blora-perf-firefox-baseline-20260926`；开始仅测试侧窗口合成对照，输出负载、输入时序及 ≤50ms 阈值不变。根代理临时性能 fixture：127.0.0.1:9443，exec 会话 48529，凭据文件路径 `/tmp/fixture-872806179/browser-credentials.json`（不得输出内容），仅 Ctrl+C 清理所属 fixture。另有独立监督长测入口正在准备，先短时真实墙钟验证，再启动 24 小时跨日运行；未跑完前不得计作通过。下一动作：比较玻璃合成/终端隔离绘制开销，完成必要修复，然后当前版本真实回归、重新打包及独立发行冒烟。

2026-09-26 UI 暂停与非 UI 剩余工作盘点：用户明确要求 UI 先优化到这里，本轮只读审计源码/台账/最新报告，未启动测试或实施功能修改。主要功能已有代码与 Linux 本地真实证据；明确仍未解决的质量问题是 WebKit 严格双 PTY p95 78ms、Firefox 144ms，超过 ≤50ms。本地仍可推进性能热点定位、冻结版本真实全套回归与重新发行：最近完整回归为 RC15，最近发行 RC23 的包内 index.html 与当前 web/dist 摘要不同。其余严格验证缺口包括 24 小时/更长期边界、Windows 真机、独立 systemd、跨主机 Engine/注册源/网络、设备断电与系统通知中心真实呈现，详见[剩余工作核对](../NON_UI_REMAINING_2026-09-26.md)。没有把已有 ENOSPC、进程崩溃、权限撤销或恢复通过项重新列为功能缺失；整体验收状态保持原台账。本轮无运行会话需接管。下一动作：停止视觉发散，非 UI 工作优先定位非 Chromium 绘制性能，再做冻结版本回归与发行收尾。

2026-09-26 全配色深浅色截图交付：按用户要求重新拍摄当前六套配色（冰蓝、青灰、暖砂、苔绿、雾紫、烟粉），每套覆盖浅/深色、通透开/关和桌面/启动器/实例中心三个场景，共 **72 张 3200×2000 原图**。使用真实 Vue 设置选择外观、24 个隔离浏览器 context 和本地 GET-only 四实例/双节点数据；未改产品源码。资源 SHA256 核对当前暖砂深色壁纸及新版浅/深备份图标，材质开关保持壁纸/图标一致，页面错误、API 写入和外网请求均为 0，见[验证记录](../../../web/ui-screenshots/all-appearance-20260926/verification.json)。生成每配色/材质一张深浅并排的三场景图板，向远程用户直接嵌入 PNG；[72 张原图 ZIP](../../../web/ui-screenshots/all-appearance-20260926/blora-all-appearance-originals-20260926.zip)已校验 72 项及 CRC。捕获脚本为 `web/.local/capture-all-appearance-20260926.mjs`，临时 Vite/Chromium 已退出。本轮只提供视觉复查，不改变产品/外部环境验收状态。下一动作：依据用户对完整配色与场景的反馈进行定向调整。

2026-09-25 暖砂深色壁纸中段留白修订：用户指出浅色版顺眼，但深色版上方约 70% 是一整块近黑底。定位到 `sand-dark.svg` 的 `#1b1b1a` 全幅底色和仅从 y≈700 开始的下方曲面，无额外 CSS 蒙层。本轮只调整深色暖砂 SVG：底色改为清晰暖炭色，加入低对比的上方宽曲面，并让下方沙丘提前进入画面，保持纯色几何、不加渐变；浅色壁纸、主题/材质切换及应用图标未改。真实 Vue 以隔离 GET 数据重拍浅色/深色和深色通透 `on/off` 共 **6 张**完整桌面/启动器，加一张[对照总览](../../../web/ui-screenshots/sand-wallpaper-rework/00-overview.png)；[验证记录](../../../web/ui-screenshots/sand-wallpaper-rework/verification.json)确认浅/深资源不同、材质开关不改壁纸/图标、零页面错误/API 写入/外网请求。`cd web && npm run build`（含类型检查）通过，暖砂深色/自选壁纸聚焦 Playwright **1/1** 通过；此 UI 微调不提升全范围验收状态。下一动作：根据用户对新版深色实景的观感继续定向调整。

2026-09-25 备份与计划图标收口：用户指出旧深色图标左侧回转缺口切断外圈、浅色钟面越过内径，观感像白色溢出。保持现有 01A 底板和浅/深配色体系，前景改为闭合圆形表盘、内收钟面与简洁指针；浅色外环单独拉开明度，更新共享矢量几何、默认/正式浅色/03 深色三份预渲染资源。`cd web && npm run build`（含类型检查）通过；主题/通透与深色图标切换 Playwright **2/2** 通过。真实 Vue 在暖砂浅色/深色、通透 `on` 下拍摄桌面/启动器与桌面 58px、Dock 44px、启动器 55px 实际尺寸共 **11 张**，确认三入口均引用更新资源、零页面错误/API 写入/外网请求；见[实际尺寸对照](../../../web/ui-screenshots/backup-icon-final/00-comparison.png)、[验证记录](../../../web/ui-screenshots/backup-icon-final/verification.json)。本轮仅改图标，不改变其他主题、壁纸或全范围验收状态。下一动作：根据用户对新图标实景的反馈继续微调。

2026-09-24 暖砂深色壁纸再次修订与远程截图交付：用户明确否定上轮青蓝壁纸，并指出远程对话无法打开工作机本地 HTML。比较真实桌面中的石墨、莓红、赤陶、香槟砂候选后，将正式 `sand-dark.svg` 改成近黑底与浅米金纯色曲面，保留既有几何和细轮廓，不用渐变或伪光影；其他五套配色、03 深色图标、浅色壁纸和通透开关未改。`cd web && npm run build` 通过；主题、图标、壁纸独立性 Playwright **3/3** 通过。六套深色完整桌面/启动器实景均已重拍，16:10、2× DPR、本地 GET-only fixture；[暖砂桌面](../../../web/ui-screenshots/dark-wallpaper-final/sand-dark-desktop.png)、[暖砂启动器](../../../web/ui-screenshots/dark-wallpaper-final/sand-dark-launcher.png)、[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)。六壁纸互异、通透开关不改变壁纸，零页面错误/API 写入/外网请求。向用户交付时直接在对话嵌入六张完整 PNG，不再仅给本地 HTML 链接。下一步根据用户对米金版及六配色的视觉反馈继续定向改进；全范围验收状态不因本轮 UI 截图而改变。

2026-09-24 暖砂深色壁纸配色修订：用户认为上轮可见度增强后的灰棕色块浑浊难看。用真实 Vue 桌面分别对照冷蓝、冷灰蓝、深蓝绿三种保留原曲线的方案，选深蓝绿替换正式 `sand-dark.svg`；去掉灰棕色及半透明灰叠色，底部曲面采用明确的青绿层次，保留一条细轮廓线。没有改动布局、03 默认深色图标、浅色壁纸、主题配色或通透开关。`cd web && npm run build` 通过；主题/图标/壁纸独立性 Playwright **3/3** 通过。真实 Vue 只读演示数据重拍[暖砂深色桌面](../../../web/ui-screenshots/dark-wallpaper-final/sand-dark-desktop.png)、[启动器](../../../web/ui-screenshots/dark-wallpaper-final/sand-dark-launcher.png)及[六配色总览](../../../web/ui-screenshots/dark-wallpaper-final/index.html)，通透开关不改变壁纸、六壁纸互异、全部加载 03 图标、零页面错误/API 写入/外网请求；[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)。图标方向对照图库也已按新正式壁纸重拍。下一步依据用户对新实景的观感继续定向微调；本次仅属 UI 改善，不改变全范围验收状态。

2026-09-24 深色壁纸可见度修复：用户指出暖砂深色桌面几乎纯色；定位到原浅色 SVG 前的 `#11171ce6` 遮罩（约 90% 不透明）把色块差压到约 1–4 RGB。保留六套主题原有几何，为冰蓝、青灰、暖砂、苔绿、雾紫、烟粉各新增深色专用 SVG，去掉这层遮罩；浅色壁纸、03 默认暗色图标和通透开关不变。青灰主题的暗色默认壁纸规则同时限定在“随配色”，不再盖掉用户自选壁纸。`cd web && npm run build`（含类型检查）通过；主题、壁纸选择、03 图标与通透独立性相关 Playwright **4/4** 通过。真实 Vue 在本地只读 fixture 下重拍六配色暗色桌面/启动器及暖砂浅色/实色共 14 张原图，六壁纸互异、通透 on/off 背景完全一致、六组均加载 14 枚 03 图标、零页面错误/API 写入/外网请求；见[总览](../../../web/ui-screenshots/dark-wallpaper-final/00-overview.png)、[全图索引](../../../web/ui-screenshots/dark-wallpaper-final/index.html)与[验证记录](../../../web/ui-screenshots/dark-wallpaper-final/verification.json)。既有 03 图标七向对照图库也已重拍成新壁纸状态。下一步继续根据用户对暗色壁纸亮度和配色的实际观感定向微调；本 UI 修复不扩大全范围验收结论。

2026-09-24 深色图标方向定稿：用户从六套预览中选定 **03「清晰造型」**。深色模式默认资源已切为 `h04-spectrum-night-clear-forms` 的 14 枚图标；浅色资源、六套主题配色与通透开关未变。旧深色版作为仅开发对照继续可复现；24 场景截图脚本的默认深色资源断言已同步。`cd web && npm run build`（含类型检查）通过，主题/通透独立性与深色图标切换 Playwright **2/2** 通过；七向真实截图重拍并确认 7×14 图标加载、相同布局、零页面错误与 API 写入，见[预览图库](../../../web/ui-screenshots/dark-icon-directions/index.html)和[验证记录](../../../web/ui-screenshots/dark-icon-directions/verification.json)。下一步在后续 UI 反馈中继续检查其他深色组件；当前所选图标已为正式默认。

2026-09-24 深色图标六方向预览：针对用户指出的 Dock 图标偏暗、辨识度不足，保留现状作对照，另备色彩提亮、柔亮底板、清晰造型、双色强调、暖色材质、亮色中心六套开发预览；**默认图标未切换**。[七向对比页](../../../web/ui-screenshots/dark-icon-directions/index.html)、[Dock 55px 总览](../../../web/ui-screenshots/dark-icon-directions/00-comparison-dock.png)与[启动器总览](../../../web/ui-screenshots/dark-icon-directions/00-comparison-panel.png)包含每向完整桌面、Dock 局部和启动器面板实拍。真实 Vue 深色/暖砂/通透开场景使用隔离 GET 数据：每向 14 张 288px 图标均加载，七向布局一致，零页面错误、零 API 写入和外网请求；详见[验证记录](../../../web/ui-screenshots/dark-icon-directions/verification.json)。`cd web && npm run build` 通过，主题配色与通透独立性、深色图标切换两项 Playwright 测试 **2/2** 通过。下一步由用户选择偏好的方向，再据此细化并决定是否切换默认方案。

2026-09-24 深色通透与亮边复核：用户所示启动器截图确为通透模式 `on`，但此前深色面材质遮盖约 96–97%，因此视觉上接近实色。本轮将深色 `on` 的启动器/菜单与顶栏调整为约 80% 遮盖、窗口约 89%、Dock 约 78%，缩小模糊半径并收减内侧反射与投影；深色窗口、启动器和 Dock 外圈使用暗壳边，启动器底部分隔线同步压暗。14 枚深色应用图标继续逐个调色，并降低仅深色 h04 导出中的反射与壳层亮度，浅色资源及六套主题配色保持独立。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过；[六配色×深浅×六场景通透 `on` 图库](../../../web/ui-screenshots/dark-mode-review/index.html)有 **72 张**实景，另拍暖砂 `off` **12 张**对照，可直接查看[暖砂深色启动器 `on`](../../../web/ui-screenshots/dark-mode-review/sand-on/08-dark-launcher.png)与[`off`](../../../web/ui-screenshots/dark-mode-review/sand-off/08-dark-launcher.png)。本地隔离 fixture 记录零页面错误、零 API 写入；这些视觉证据不替代真实节点、后端或外部平台验收。下一动作：根据用户对通透度、暗边与图标亮度的实景反馈继续定向微调，保留全范围验收矩阵中的未验证项。

2026-09-24 深色窗口边界与图标亮度收口：按用户指出的实例中心截图，深色窗口外框由带主题色的明显描边改为中性细线（普通/焦点分别为 5%/7% 前景色），保留阴影区分窗口；实例中心内容区原有四角圆角在左下角露出壳体，现仅在深色模式将两个底角归零，消除侧栏底部的月牙形接缝。14 枚深色应用图标进一步按暗背景调低高亮、调整底板与主体图层对比，保留各应用造型和辨识色；浅色模式、六套主题配色与壁纸选择未改。`cd web && npm run build`（含类型检查）通过，主题相关 Playwright **4/4** 通过。[六配色×深浅×六场景图库](../../../web/ui-screenshots/dark-mode-review/index.html)共 72 张实景；[暖砂实例中心左下角局部](../../../web/ui-screenshots/dark-mode-review/sand-on/09-dark-instances-corner.png)用于复核接缝。图库的本地隔离 fixture 记录为零页面错误、零 API 写入；它不替代真实节点、后端或外部平台验收。下一动作：按用户对本轮暗色边界和图标亮度的实际反馈继续精修，不恢复已暂停的壁纸探索。

2026-09-24 深色模式去浑浊修订：用户指出六套深色截图整体发脏、图标难辨。本轮将暗色大面积窗口、菜单、Dock 和表单统一到中性炭灰层级，各主题的强调色保留在按钮、焦点和选中态；壁纸暗色显示层不再分别染棕/绿/紫，壁纸资源、选择项和浅色呈现未改。暗色材质面提高到约94–97%不透明并减轻阴影、卡片内描边；终端暗色画布由绿黑改为中性黑灰，已打开终端随主题即时切换。14枚选定应用图标重新以清晰浅彩底板和更强主体轮廓导出，保留原有01A造型与应用颜色。`npm run build`（含类型检查）通过，聚焦桌面 Playwright **4/4** 和终端恢复回归 **1/1** 通过；[重新拍摄的六配色×六场景深浅对照图库](../../../web/ui-screenshots/dark-mode-review/index.html)共72张实景及六张概览，核验每套14枚浅色/深色图标、零页面错误和零API写入。截图使用隔离演示数据，不代表真实节点/后端验收；下一动作按用户对新暗色观感的反馈继续定向微调，暂不探索新壁纸。

2026-09-24 深色模式视觉适配：按用户最新反馈暂停壁纸探索，本轮没有修改壁纸资源、壁纸选项或六套主题配色。为选定的 01A 应用图标新增 14 枚逐个调整的深色资源；沿用原图形与应用辨识色，主题切换时立即替换，配色和通透开关不改变图标造型。修正深色顶栏、Dock、窗口、菜单、表单、列表、状态标签、应用卡片和日志表面的层次与边界；Monaco 编辑器在已打开时也随主题切换。冰蓝深色状态色与禁用文字对比单独调整。`npm run build`（含类型检查）及聚焦 Playwright **4/4** 通过，覆盖深色图标/编辑器实时切换、六配色和通透材质独立及刷新恢复；截图脚本语法检查通过。[六配色深浅色对照图库](../../../web/ui-screenshots/dark-mode-review/index.html)共 **72 张**真实 Vue 截图，覆盖桌面、启动器、实例中心、设置、应用市场、监控与终端六种场景，另有六张概览。截图核验每套均加载 14 枚浅色与 14 枚深色图标，页面错误、API 写入均为 0。截图使用隔离演示数据，不扩大后端、节点或外部平台验收结论。下一动作：依据用户对暗色图标和各场景的反馈做定向微调；壁纸方向保持暂停。

2026-09-24 壁纸艺术媒介重做：用户指出此前六张只是同一套低饱和扁平插画的构图变化。本轮新增五个真正跨媒介方向：国家公园真实摄影、瑞士套印海报、深色数字光学、水墨纸张、压纹玻璃材质；保留此前壁纸与“随配色”默认值，不强制切换用户现有选择。真实摄影取自[NPS 公有领域原图](https://npgallery.nps.gov/AssetDetail/204B23EC-1DD8-B71B-0B21B310C2A73B98)，裁切及镜像等处理见[来源记录](../../../web/src/appearance/wallpapers/media/PHOTO_SOURCE.md)；其余四张为代码绘制的 SVG，无 AI 图片生成。照片的桌面图标文字采用局部半透明底衬，夜光图用浅色文字以保持可读。壁纸独立选择仍不改变主题、通透材质或统一的 01A 彩色应用图标。`npm run build`（含类型检查）通过，聚焦 Playwright **3/3** 覆盖全部11种自选壁纸、配色/材质独立及刷新恢复。实拍[五种艺术媒介的四场景图库](../../../web/ui-screenshots/wallpaper-art/index.html)，含桌面、启动器、实例中心、深色桌面共20张原图和4张对照图；五种浅/深色壁纸均各异、14应用图标资源及配色/材质固定，页面错误和API写入均为0。截图使用隔离演示数据，不扩大后端验收范围；临时服务与浏览器已停止。下一动作：根据用户对艺术方向的反馈选择和细化；不再把单纯换构图/配色描述为不同画风。

2026-09-24 六种独立壁纸风格探索：针对用户指出原六套壁纸仅是同一弧线构图换色，新增柔和色场、建筑平面、等高线、编辑网格、抽象雕塑、静谧地平线六张不同构图的 1600×1000 代码 SVG。设置新增独立“壁纸风格”选择，“随配色”为默认且保留原壁纸；指定风格时更换壁纸不改变主题主色、浅/深色、通透材质或选定的 01A 彩色应用图标。聚焦 Playwright **3/3** 覆盖旧配色/通透独立性和六张壁纸唯一性、配色/材质独立、刷新持久化；`npm run build`（含类型检查）通过。用真实 Vue 桌面和隔离演示数据拍摄[六种壁纸的四场景图库](../../../web/ui-screenshots/wallpaper-styles/index.html)，含桌面、启动器、实例中心、深色桌面共24张原图与4张对照图；核验六种浅/深色壁纸各异、配色/材质/14图标资源一致，0 页面错误、0 API 写入。截图不扩大后端验收范围；临时服务已停止。下一动作：按用户对壁纸方向的选择收敛，而非继续仅更换弧线配色。

2026-09-24 六套主题配色实景对照：在冰蓝、青灰之外新增暖砂、苔绿、雾紫、烟粉；每套具有独立的浅/深色语义色和代码绘制壁纸，保留用户选定的 01A 彩色应用图标。主题配色、浅/深色和通透材质仍分别切换，通透开关不改变当前配色、壁纸或图标。`npm run build`（含 `vue-tsc --noEmit`）通过；聚焦 Playwright **2/2** 验证四套新增配色及旧两套的材质独立性与刷新持久化。[六套主题图库](../../../web/ui-screenshots/palette-showcase/index.html)包含六配色×通透开/关×桌面、启动器、实例中心、设置、深色桌面、深色启动器，共 **72 张真实 Vue 截图**及六张概览图；截图脚本核对六种主色/壁纸各异、同配色两材质主色/壁纸一致、12 组合的 14 个应用图标资源相同、材质确实变化，页面错误和远端写入均为 0。截图使用隔离演示数据，不扩大后端验收结论；本地截图服务和浏览器已停止。下一动作：根据用户对六套配色的选择，细调选中方案的壁纸与界面色。

2026-09-24 通透材质与主题配色解耦：按用户新要求，工作区设置新增独立“主题配色”（冰蓝 / 青灰）选择；通透模式现在只控制顶部栏、窗口、菜单与 Dock 的通透材质，切换时保留配色、壁纸和统一的 01A 应用图标。配色选择负责色彩令牌与配套壁纸，不会改变通透开关；原有“浅色 / 深色”主题仍单独保留。旧工作区尚无配色字段时，按原通透开关推断并在首次切换前固化旧视觉选择，避免迁移后突然换色。样式将冰蓝色值/壁纸与调色板无关的通透层分离，通透层改用当前配色色值混合，不再硬编码冰蓝。聚焦 Playwright 覆盖四组合、主色/壁纸/图标不随材质变化、两设置独立持久化 **1/1**；`npm run build`（含 `vue-tsc --noEmit`）通过。用真实设置控件与隔离演示数据实拍[四组合×25场景图库](../../../web/ui-screenshots/palette-matrix/index.html)共100张及7张四宫格，包含深色设置页；校验同配色开/关主色、壁纸与14图标资源一致，顶部栏材质有差异；四组无页面错误或远端写操作。截图不扩大后端验收结论。下一动作：按用户对四组合效果的反馈微调材质强度与配色；不再将通透开关和配色合并。

2026-09-24 双模式应用图标统一：按用户反馈，实色模式不再切换到较简化的旧 h04 图标；通透与实色模式均使用选定的 01A（`h04-spectrum-silver`）彩色应用图标，开关只改变界面材质与壁纸。设置说明同步更新。用真实设置开关重拍[通透24场景](../../../web/ui-screenshots/translucency-on/index.html)、[实色24场景](../../../web/ui-screenshots/translucency-off/index.html)及[24组对照](../../../web/ui-screenshots/translucency-comparison/index.html)；实际渲染的14个图标资源路径在两组中完全相同，零页面错误、零远端写入。`npm run build`（含 `vue-tsc --noEmit`）通过，Playwright 通透开关/刷新后图标一致性测试 **1/1** 通过。截图仍使用隔离演示数据，不扩大后端验收结论；临时截图服务已停止。下一动作：按用户对双模式统一图标的实景反馈微调。

2026-09-23 实例中心控件回退与通透模式开关：按用户图1撤回上两轮放大搜索条/白底下拉框的改动，恢复直接搜索输入和紧凑填充控件；下排三筛选框回到28px/11px并保持等宽。实际截图的实例窗口区域与既有图1/图3候选截图近乎逐像素一致，控件区域三通道平均绝对差均小于0.35/255。工作区设置新增可访问的“通透模式”开关，偏好随工作现场持久化，默认开启；开启切换冰蓝通透材质、曲面壁纸与已选01A彩色图标，关闭切换用户图3的青灰实色表面、默认壁纸与默认图标。补齐两种模式深色桌面的壁纸压暗与可读文字，开启时深色窗口/Dock仅保留克制通透。用真实设置开关和隔离演示数据分别实拍[开启24场景](../../../web/ui-screenshots/translucency-on/index.html)、[关闭24场景](../../../web/ui-screenshots/translucency-off/index.html)，并生成[24组并排对照](../../../web/ui-screenshots/translucency-comparison/index.html)；两组均14图标、零pageerror/远端写操作，图库筛选和48张原始PNG完整性核验通过。`npm run build`（含`vue-tsc --noEmit`）、`npm run test:e2e -- tests/browser/desktop.spec.ts tests/browser/instance-batch.spec.ts` **8/8**、截图脚本语法检查通过。截图为本地隔离演示数据，不扩大后端验收结论；Vite和浏览器会话已结束。下一动作：根据用户对两种模式实景对照的反馈微调；不要再自动把实例筛选放大。

2026-09-23 实例中心节点筛选圆角修正：用户指出“全部节点”下拉框看起来变成直角；前一轮照搬4px小圆角到56px高控件，实际视觉比例不合适。现将该下拉框调为16px圆角、下排40px高的三个筛选框调为12px圆角，保留原生选择语义、高度和布局。检查无其他CSS覆盖形状；重新实拍[实例列表](../../../web/ui-screenshots/01a-flow-contexts/05-instances.png)与[1366宽列表](../../../web/ui-screenshots/01a-flow-contexts/23-laptop-instances.png)，并重拍完整24场景图库（14张选定图标、零pageerror/远端写操作）；`npm run build`（含`vue-tsc --noEmit`）通过。截图使用隔离演示数据，不计后端验收；临时Vite与浏览器已停止，无运行会话。下一动作：按用户对圆角比例的反馈微调，其余视觉继续沿用选定原版。

2026-09-23 实例中心表单形状核对与修正：用户质疑截图中的输入框是否符合 Material Design。核对 Google 官方 Material Web [文本框](https://github.com/material-components/material-web/blob/main/docs/components/text-field.md)、[选择框](https://github.com/material-components/material-web/blob/main/docs/components/select.md)和 Android [搜索](https://github.com/material-components/material-components-android/blob/master/docs/components/Search.md)文档后确认，原搜索框约28px、第二排筛选约24px且视觉角色混同，不能称为按 M3 默认组件实现。现将实例搜索改为带搜索图标的56px搜索条，节点下拉改为56px/4px角的选择框，第二排改为40px/8px角的桌面密度选择框；保留真实原生选择语义、筛选状态和布局。当前是基于官方角色与形状令牌的桌面适配，不宣称引入官方 Material Web 组件库或每个像素都与 Android 默认控件相同。已重拍[实例列表](../../../web/ui-screenshots/01a-flow-contexts/05-instances.png)、[卡片](../../../web/ui-screenshots/01a-flow-contexts/06-instance-cards.png)和[1366宽列表](../../../web/ui-screenshots/01a-flow-contexts/23-laptop-instances.png)，并修正窄屏截图误拍资源详情的问题；全24张图库零pageerror/写API。Playwright `desktop.spec.ts`＋`instance-batch.spec.ts` **7/7**通过，`npm run build`（含 `vue-tsc --noEmit`）与截图脚本语法检查通过。截图仍是隔离演示数据，不计后端验收；临时Vite与浏览器已停止，无运行会话。下一动作：依据用户对新比例的视觉反馈继续微调，不把其他页面自动认定为已完成 Material 3 审计。

2026-09-23 选定原版的多场景截图：按用户要求继续展示 **01A (`spectrum-silver`) 图标＋W4青灰流线 (`flow`) 壁纸＋现有Material布局**，复用真实 Vue 组件截图库并新增 `--01a --flow` 参数，生成[24张可筛选图库](../../../web/ui-screenshots/01a-flow-contexts/index.html)：桌面/启动器/Dock、实例列表与卡片、文件和编辑器多窗口、终端/监控、应用市场、节点/任务、设置/账号、深色主题、1366×768与1024×768。新浅色启动器与用户选定的原图同为3200×2000，RGB逐像素完全一致。抽查时发现原深色预览因浅壁纸导致桌面标签几乎不可读，已在预览专用 `ice.css` 中为同一壁纸加均匀暗色遮罩；深色图库明确标记为适配预览。14张选定图标均加载，24张PNG/27张图库图片与分类筛选核实；零pageerror、零API写请求，终端只读样本协议连接1次。`node --check .local/capture-04a-contexts.mjs`、`npm run check`、`npm run build`通过；见[验证结果](../../../web/ui-screenshots/01a-flow-contexts/verification.json)。截图使用隔离本地演示数据，不增加后端验收结论，也没有将预览图标/壁纸切换为产品默认。临时Vite和浏览器已停止，无运行会话。下一动作：按用户对这些实景截图的反馈细调选定视觉基线，不恢复已被否定的其他图标方向。

2026-09-23 用户选回原版作为设计基线：用户在看过F1～F4后表示“感觉还是这个原版好看”。附图与 `harmony-12/spectrum-silver--flow--launcher.png` 的3200×2000 RGB像素完全一致，明确对应 **01A (`spectrum-silver`) 原版图标＋W4青灰流线 (`flow`) 壁纸＋冷白冰蓝界面/现有Material布局**。后续以此为基线，保留各应用独立颜色、图标底板与h04轻透材质；F1～F4及其他配色保留为探索档案，不作为当前选定方向。已将[对照页](../../../web/ui-screenshots/harmony-12/index.html)默认组合改为该选择，并标出基线及[原图](../../../web/ui-screenshots/harmony-12/spectrum-silver--flow--launcher.png)。本轮仅记录选择并修改图库默认展示，未把偏好反馈扩展为正式产品发布或默认主题切换；无需重跑产品构建。没有启动服务，无运行会话。下一动作：围绕此基线接受具体细调要求，不继续无目标扩散图标风格。

2026-09-23 图标主体四方向发散：用户明确要求图标本身产生多个方向。新增F1透明叠片、F2饱满大形、F3精致器物、F4彩色折面，每套14个主体全部重画；固定上轮A浅底板、暖白壁纸与h04材质。渲染器补齐旋转和evenodd填充，导出56张288×288资产共814,685字节。最终类型检查/构建通过，正式产物仍仅14张默认图标；33张Chromium实拍含全套/尺寸/桌面/启动器/Dock/实例窗口。去除颜色后比较确认56个几何全部改变、方向间无同应用重复，底板一致；布局、25种图库切换及390px图库验证通过，零pageerror/写API。人工修正市场类似卡通脸、设置类似风扇及分色齿轮重复边缘后重拍对应截图。见[造型与验证记录](../ICON_FORMS_13.md)、[四套对照页](../../../web/ui-screenshots/forms-13/index.html)。仅开发预览启用；使用隔离本地演示数据，不计后端验收。全部本轮浏览器、Vite与构建会话已结束。下一动作：按用户对主体方向的反馈继续细化，再与已认可配色/壁纸组合，不自动切换默认。

2026-09-23 01A 图标/壁纸协调方案：针对实色底与浅色底割裂，新增浅瓷底、适中彩底、全彩底、深底彩芯四套完整配色，主体和内部图层同步调整；保留01A图形、h04材质和Material布局。另有暖白、纸层、雾紫、青灰流线、浅灰纹理五张代码壁纸，可与原版一起独立切换30种组合。56张288×288图标共749,685字节，仅开发预览启用。类型检查/构建通过，正式产物仅14张默认PNG；实际Chromium实拍57张，30组启动器/5组窗口几何一致，零pageerror/写API，图库切换与390px宽度验证通过。审图后修顺青灰壁纸尖角并重拍7张实景和2张对照。见[实现与验证记录](../ICON_HARMONY_12.md)、[完整对照页](../../../web/ui-screenshots/harmony-12/index.html)。截图为隔离本地演示数据，不增加后端验收结论；默认外观未切换，临时Vite与浏览器已停止，无运行会话。下一动作：依据用户对图标和壁纸的独立选择继续收敛。

2026-09-23 改用01A预览：用户否定04A实景方向并要求查看01A。沿用原始 `spectrum-silver` 资产，在与上一轮相同的24个场景、固定时间和窗口内容下重新截图；未重绘图标或修改产品布局/主题。新启动器特写与最初01A特写逐RGB像素一致（1240×904）；14张01A资产全部加载，零pageerror、零API写请求，图库27张图片和3/3/4/24分类验证通过。见[01A完整图库](../../../web/ui-screenshots/01a-contexts/index.html)、[验证结果](../../../web/ui-screenshots/01a-contexts/verification.json)。复用命令：`cd web && node .local/capture-04a-contexts.mjs --01a`（需先启动本机Vite）。截图继续使用隔离演示数据，不计后端验收；仅截图脚本新增01A参数，无需重新构建产品。临时Vite和浏览器已停止，无运行会话。下一动作：依据用户对01A的反馈继续，04A已被用户否定，不能作为新的默认方向。

2026-09-23 04A 多场景视觉评审：按用户选中的 `pastel-silver` 扩展到桌面/启动器、实例列表与卡片、文件编辑多窗口、终端与监控、市场、节点、任务、设置、账号菜单、深色主题及两档小屏，共24张2倍像素截图。保留图标与产品源码，未切换默认主题。全部渲染图标为04A，14张资产加载；实际Vue/Monaco/xterm组件使用隔离只读视觉样本，零pageerror、零API写请求，图库27张图片加载及四类筛选通过。记录现有深色主题配浅壁纸造成桌面文字对比度不足。见[截图与验证记录](../ICON_04A_CONTEXTS_11.md)、[24张图库](../../../web/ui-screenshots/04a-contexts/index.html)。临时Vite和浏览器已停止，无本轮运行会话。下一动作：依据用户对04A实际场景的反馈继续收敛，不自行恢复其他环境验收。

2026-09-23 终端与监控六组细化：用户认可01/04整体配色，要求改终端造型及两应用颜色。基于两套基础分别新增A银白窗口＋绿白监控、B紫色叠片＋蓝白监控、C浅色命令页＋暖橙监控，共六组。新代码绘制三种终端几何，继续沿用h04材质和底板；每组另外12个图标与对应基础版PNG逐字节一致。84张288×288候选资产共1,095,755字节，仅开发参数启用。最终类型检查/构建通过，正式产物仅14张默认图标；Chromium禁用WebGL完成30张实景/局部＋2张对照，六组13应用图标加载、三类布局比较与对照页切换均通过，零pageerror。见[细化与验证记录](../ICON_UTILITIES_10.md)、[六组对照页](../../../web/ui-screenshots/icon-utilities/index.html)。截图使用只读本地演示数据，不计真实终端或后端验收；所有本轮进程已停止。下一动作：按用户对01A/B/C和04A/B/C的选择继续收敛。

2026-09-23 图标独立配色四版：用户认可 h04 图标质感但要求摆脱统一冰蓝色。参考 Apple App icons 与 Liquid Glass 官方文档，保留图标几何、材质、底板和冷白冰蓝桌面，逐应用/逐图层制作「应用原色」「浅底彩芯」「明暗交错」「彩色薄片」四组；56张代码渲染PNG共734,796字节。开发参数 `?appearance=ice&icon-colorway=...` 可切换，默认版不变。禁用WebGL的Chromium实拍四版各桌面/启动器/特写/Dock/实例中心及并排总览21张，13个已安装应用图标逐版加载，三组场景几何一致，零pageerror。首次构建发现预览PNG误入正式产物，已改为仅开发服务器加载；复验 `npm run build` 通过，正式产物只含14张默认图标PNG。对照页切换和原冰蓝预览资产加载复核通过。见[设计、截图与验证记录](../ICON_COLORWAYS_09.md)、[可切换对照页](../../../web/ui-screenshots/icon-colorways/index.html)。截图仅用只读本地演示数据，不计后端验收；Chromium及临时Vite已停止，无运行会话。下一动作是根据用户具体偏好收敛各应用颜色。

2026-09-23 冷白冰蓝单版预览：按用户新参考图保留当前 Material 形状、圆角、布局与 h04 图标底板，新增开发参数 `?appearance=ice`；用冷白/蓝灰表面、蓝/青绿/紫/深灰图标、柔和阴影和代码曲面壁纸复现参考氛围，未替换默认主题。14个图标仅改原图层颜色，保留轮廓与材质参数，导出288×288资产共182,848字节；正常显示不依赖WebGL，没有图片生成或斜向反光条。最终类型检查/构建通过；禁用WebGL的Chromium实拍12张新版＋3张默认对照，七组场景的几何/字号/间距/圆角/Dock路径一致、零pageerror。已检查菜单透出与文字对比度并更新截图。见[实现与验证记录](../MATERIAL_ICE_PREVIEW.md)和[交互对照页](../../../web/ui-screenshots/material-ice/index.html)。截图使用本地演示数据，不计后端验收；浏览器和临时Vite均已停止。下一动作：依据用户对这一个版本的反馈收敛外观，继续保留Material形状与图标底板。

2026-09-22 h04 保留底板设为默认：用户选择“保留底版好一些”。已将「透明包边＋保留底板」统一用于桌面、启动器、Dock、窗口标题栏和应用市场。原材质/图形/配色未改；从原代码渲染器导出14个288×288资产（185,447字节），默认显示不再依赖实时WebGL，保留原viewBox与Dock比例。类型检查/正式构建通过；本地正式构建在禁用WebGL的Chromium中验证14类资产全部加载、零WebGL请求、零pageerror，桌面→启动器→实例中心和生产忽略开发参数检查通过，另有5张实际截图。见[默认图标记录](../ICON_GLASS_H04_DEFAULT.md)。测试使用本地演示数据，无远程节点操作或公开部署；临时开发/预览服务及浏览器已停止。下一动作：后续视觉细化以本版为基线，保留底板，不恢复斜向反光或其他被否定方向。

2026-09-22 h04「透明包边」实际场景扩展：用户明确选定 h04，要求多种实际界面截图。本轮保持 h04 图形/配色/材质参数，不新增风格；开发预览增加 `icon-surface=bare` 对齐所选裸主体截图，`icon-background=graphite|warm` 提供深浅背景对照。24 张真实组件实拍涵盖完整桌面、13应用启动器、实例列表/卡片、文件/编辑器、多窗口、应用市场、节点/任务、设置/账号、深色、暖灰、1366×768和1024×768、字号及保留底板对照。最终脚本24张零pageerror、所有光学图标ready，PNG与浏览页链接核实；生产构建/类型检查通过。首轮主题选择器不匹配已修正并完整重拍。API使用本地演示数据且拒绝写请求，未操作真实节点，不新增后端验收结论。默认图标未切换；禁止斜向反光和图片生成继续有效。见[场景记录](../ICON_GLASS_H04_CONTEXTS.md)及[24图浏览页](../../../web/ui-screenshots/glass-h04-contexts/index.html)。本轮浏览器已退出，临时Vite已停止，无运行会话需接管。下一动作：依据用户对实际场景的反馈收敛h04；不擅自加强光影或恢复被否定的其他方案。

2026-09-22 图标玻璃材质第八轮：用户明确禁止斜向反光带。已从shader和配置删除该效果，旧g系列后续渲染也移除；今后不再引入斜向反光、扫光条或面内大块方向光影。新增h01～h08八种透色/吸色/折射/包边/彩边/夹层/乳白/平衡壳层候选，保留原图形与底板；细线壳宽限制为笔画20%，避免备份箭头变淡。构建通过；八版前景内部像素验证确认实际透色，平整中央无方向光色差。并发构建/浏览器超时后改为顺序执行复核。46张代码实拍涵盖八版全套、小尺寸、桌面/启动器/Dock与测试衬底；默认图标未切换。见[玻璃通透记录](../ICON_GLASS_08.md)。下一动作：按用户选定的玻璃方向收敛，严格保留禁止斜向反光带和禁止图片生成要求。

2026-09-22 图标光学材质第七轮：用户指出薄边缘不可见，要求更多反射玻璃方向。查阅Apple Liquid Glass官方说明，新增共享WebGL渲染器：保留原14应用/63部件，轮廓生成法线/距离纹理，逐层折射采样与移动光源反射，6版主体不使用Gaussian blur。开发 `?icon-study=g01`～`g06`，默认图标不变；36张实际Chromium截图含裸主体、深浅底、完整图标、32/48/64px、三光源和桌面/启动器/Dock。类型检查/构建通过；canvas渲染状态与六版左右光源图像变化均核实，无pageerror。修复透明采样暗点及光源轴向；一次Vite被SIGTERM后的连接失败已恢复重拍。属于网页近似材质候选，非Apple原生或跨平台性能验收。见[光学候选记录](../ICON_OPTICS_07.md)。下一动作：按用户选择收敛，继续直接代码渲染，不用图片生成。

2026-09-22 图标光影第六轮：用户认可主体模糊透叠，但要求优化夸张渐变及提供无光影版。以f06固定几何/颜色/透明度/blur制作A1微弱面光、A2仅薄边缘、B无光影；B的SVG不生成任何渐变/光斑/亮边，仅保留图层模糊透色。修复填色与描边重复合成的暗边。默认图标未切换，开发参数 `?icon-study=l01`～`l03`；15张真实Chromium截图覆盖放大、全套、小尺寸与桌面/启动器/Dock。类型检查/构建通过，截图脚本验证B零光影定义、模糊仍存在、底板一致且无pageerror。见[光影对照记录](../ICON_LIGHTING_06.md)。下一动作：按用户选择收敛；不恢复无关阻塞验收。

2026-09-22 图标主体材质第五轮：用户指出上一轮只改底板。本轮新增原版14应用/63部件的逐层材质合成，Gaussian diffusion裁剪于主体面板，文件夹前板实际透出后方纸张，笔身和店铺等分别处理，6版共用原版不透明底板。备份等粗线使用alpha mask，终端符号保留清晰实色。全部SVG/Vue代码，无图片生成；仅开发 `?icon-study=f01`～`f06` 启用。类型检查/生产构建及Chromium截图验证通过，28张包含去底板主体对比、放大、20/32/48px和实际桌面/启动器/Dock；中途线条裁剪问题已修复并重拍。见[主体材质记录](../ICON_FOREGROUND_05.md)。下一动作：依据用户对主体材质的反馈收敛，默认图标尚未切换，不恢复无关环境阻塞验收。

2026-09-22 图标材质第四轮：用户否定第三轮拟物化，明确以最初Material图标为基础，仅增加适量通透与高斯模糊材质。新 `ClassicIconArtwork.vue` 提取原版图形，源码比较确认几何不变；`MaterialAppIcon.vue` 实现12种 Gaussian blur底层透色/CSS backdrop-filter/微弱层次，非整图模糊。全部手写SVG/CSS，无图片生成。仅开发 `?icon-study=m01` 至 `m12` 启用；`icon-materials.html` 包含原版对照、深浅底、20/32/48像素和分组。修正SVG属性类型后类型检查与构建通过；Chromium12套实际桌面/启动器无pageerror，53张2倍像素截图核实齐全。见[材质方案与官方资料](../ICON_MATERIALS_04.md)。下一动作：依据用户选择的材质方向细化；不自行改变默认风格或恢复环境阻塞验收。

2026-09-22 应用图标第三轮：用户指出 D–L 偏工具栏功能符号，并明确要求停止图片生成、直接编写代码后截图。已遵从：新 `NativeAppIcon.vue` 14图标全部使用手写 SVG，未引入生成图片资产；文件夹/文档笔/安装袋/服务器托盘/包装箱采用独立轮廓，终端与监控使用完整面板，有限局部渐变只表现材质面。仅开发 `?icon-study=native` 启用，默认方案保持评审前状态。`native-icons.html` 展示四个代表图标、全套及 Dock；截图脚本 `web/.local/capture-native-icons.mjs` 输出8张2倍像素图，目录 `web/ui-screenshots/native-icons-03/`。`npm run check`、`npm run build` 通过；Chromium 桌面→启动器→实例中心正常，无 pageerror。固定 API 演示数据不构成后端验收。下一动作：依据用户对完整应用造型的反馈细化；后续图标工作直接使用 SVG/CSS，不使用图片生成。

2026-09-22 图标设计第二轮：用户否定 B 渐变并要求更多方向。移除 B 主体/底色渐变；新增 D–L 九套14图标候选（双色线条、全幅纯色、无底板、折面几何、石墨面板、几何标记、透明浅底、侧边索引和墨色主体）。四套线条方向统一 Lucide，其余为本地双色矢量；默认图标与桌面布局保留。开发预览 `?icon-study=d` 至 `l`，总览 `/icon-explorations.html`。类型检查和生产构建通过；Chromium生成37张截图，包含九套真实桌面/启动器及20/32/48像素和深底检查。见[设计与截图索引](../ICON_EXPLORATIONS_02.md)。下一动作：等待用户评价具体方向再细化，未恢复环境阻塞的其他任务。

2026-09-22 图标三方向设计评审：按用户最新要求重启图标视觉探索，保留桌面布局，新增 `IconStudy.vue` 三套14图标候选：A双色圆形、B浅色层叠、C深石墨双色。仅开发模式 `?icon-study=a|b|c` 使用候选，默认图标未切换；`icon-study.html` 提供并排比较。参考 Android 官方 adaptive icons 和 Apple HIG app-icons 的遮罩、分层与清晰轮廓原则，非系统原生材质复刻。`npm run check`、`npm run build` 通过，Chromium 已生成10张2倍像素截图，目录 `web/ui-screenshots/icon-study/`，包括三套实际桌面/启动器及图标对比。截图使用固定登录测试数据，不作为后端验收证据。下一动作：根据用户选定方向细化，不自行恢复其他环境阻塞任务。

2026-09-22 E08 Firefox 双 PTY 实测：官方 Firefox 155.0 在强化后的真实8窗口/双PTY/万项目录/16MiB循环传输60秒组合中，两条PTY分别116,029.8/116,120.3B/s、四轮传输、队列和实例状态均正常，但1,020个指针样本 p95 **144ms**，故按≤50ms阈值失败；Firefox无可用WebGL，两个终端使用默认渲染器。该结果与WebKit78ms、Chromium42ms并列记录，不能用Chromium覆盖非Chromium失败；见[Firefox报告](../../acceptance/reports/firefox-performance-dual-pty-2026-09-22.md)。下一具体动作：本机已补齐三种可用浏览器的严格双PTY短时组合，停止无信息增益的浏览器重跑，仅在获得可区分绘制热点或外部平台时继续。

2026-09-22 E08 双 PTY 负载复核：修正真实性能测试不稳定地按前两个终端 DOM 容器输入的问题，改为在每条新会话仍处于前台时向其唯一可写 PTY 注入有界输出；每分钟严格要求两个 WebSocket 都有新增字节。官方 WebKit 26.6 的60秒真实8窗口/万项目录/16MiB循环传输组合中，两条 PTY 分别约115.8KB/s、四轮传输与队列均正常，但3,072个指针样本 p95仍为**78ms**，故按≤50ms阈值失败；同一修正场景官方 Chromium 双PTY p95 **42ms**通过。后台终端画面暂停无收益，动画帧合批又会破坏 Vim 后紧接输入的时序，均已撤回；回退后的真实 Vim/刷新/不重放输入1/1（9.1s）通过。UI未改；见[双PTY复核](../../acceptance/reports/webkit-performance-dual-pty-short-2026-09-22.md)。下一具体动作：不重跑已知失败的一小时WebKit负载；本机只剩需要新的可区分性能热点或外部平台的严格验收。

2026-09-22 E02 rootless Engine 安全替代未启动：官方 `docker:29-dind-rootless` 未使用 `--privileged`；精确把任务证书目录交给其 UID/GID 1000、模式700后，mTLS证书成功生成，但 rootlesskit 受当前 seccomp 的 `/proc/self/exe` 执行限制拒绝（`operation not permitted`）。未尝试 `seccomp=unconfined` 或其他安全策略放宽；已删除失败容器、私有证书/密钥目录和拉取镜像。独立 Engine 仍需明确授权高权限/放宽安全策略，或使用外部测试主机。

2026-09-22 E02 独立 TCP Engine 探索未执行：为补足现有 loopback TLS 代理之外的独立 Engine 协议链路，计划启动任务专用 mTLS Docker-in-Docker 并仅映射临时 loopback 端口、导入现有测试镜像；平台自动审批拒绝 `--privileged` Docker-in-Docker，理由是其扩大到宿主内核与容器边界的高权限范围。没有执行特权容器、端口映射、镜像导出或证书写入；复核无残留。当前安全替代不能提供独立 Engine，因此 E02 的跨主机/独立 Engine 缺口继续保留，等待明确授权或外部测试主机。

2026-09-22 真实 Chromium 容器回退修正：`playwright.real.config.ts` 在未设置 `BLORA_CHROMIUM` 时改由锁定 Playwright 解析自带 Chromium，显式变量仍优先系统路径；官方容器此前因硬编码 `/usr/bin/chromium-browser` 无法启动。类型检查通过；在80ms/10ms抖动/1%丢包的隔离 namespace 中，官方 Chromium E01完整组合 **1/1通过（3.6m）**，与既有WebKit证据合并见[netem监控历史报告](../../acceptance/reports/netem-monitor-history-2026-09-22.md)。fixture、规则和状态均已清理。下一具体动作：不重复已覆盖的本机netem浏览器组合，保留剩余跨主机、Windows、systemd、物理故障与24小时边界。

2026-09-22 E07 受控网络补验：80ms单向延迟、10ms抖动、1%丢包的任务专用Docker namespace 中，WebKit 本地注册源两版本浏览、安装、升级、禁用/启用与 sandbox bundle 执行 **1/1通过（13.6s）**；见[netem注册源报告](../../acceptance/reports/netem-extension-catalog-2026-09-22.md)。fixture/规则/状态目录已清理。下一具体动作：停止重复单机netem组合，回到仍需外部环境的跨主机、Windows、systemd、物理故障与24小时验收缺口。

2026-09-22 E06 受控网络补验：同一任务专用 Docker namespace 的80ms单向延迟、10ms抖动、1%丢包下，WebKit 独立扩展包安装、多窗口、真实WASI节点任务、服务端接受后浏览器回执丢失的同键恢复、快捷方式和刷新现场 **1/1通过（49.4s）**；扩展测试改为断言明确 `Error:` 状态，兼容 WebKit 的 `Load failed` 与 Chromium 的错误文本，仍严格核对同请求键恢复。报告见[netem扩展任务报告](../../acceptance/reports/netem-extension-task-2026-09-22.md)，测试容器和状态已清理。下一具体动作：继续只运行不同的可隔离验收组合，不用网络场景替代Windows、跨主机或物理故障。

2026-09-22 E04 受控网络补验：任务专用 Docker namespace 的 `tc/netem` 80ms单向延迟、10ms抖动、1%丢包下，官方WebKit的真实分钟计划场景 **1/1通过（1.5m）**；相邻两个槽各唯一接受一次，第二槽绑定更新后的源正文，两份归档均恢复核对。fixture、namespace规则和状态目录已清理；见[netem分钟调度报告](../../acceptance/reports/netem-schedule-wall-clock-2026-09-22.md)。不替代跨主机、24小时或物理掉电验证。下一具体动作：继续审计是否还有可在隔离环境中补充且不重复的严格验收组合。

2026-09-22 E01 受控网络补验：任务专用 Docker network/PID namespace 的 `lo` 用 `tc/netem` 实际施加80ms单向延迟、10ms抖动和1%丢包；官方 WebKit 与私有HTTPS Master/双Daemon fixture 上，1秒间隔121点→Daemon离线stale→Master重启缓存→Daemon恢复完整场景 **1/1通过（3.6m）**。容器销毁即删除私有netem规则，fixture状态目录也已清理；详见[netem监控历史报告](../../acceptance/reports/netem-monitor-history-2026-09-22.md)。这增加受控网络证据，不替代跨主机远端Engine、Windows或24小时墙钟。下一具体动作：继续只选择能在隔离环境实际运行的缺口；不重跑已覆盖的普通浏览器与本地Docker场景。

2026-09-22 E08 WebKit 失败后诊断：官方容器空白页双 `requestAnimationFrame` 180次为 p50 32ms/p95 33ms；60秒真实混合诊断中恢复日志写入p95 1ms、没有 long task，慢响应集中在事件处理后的绘制阶段。尝试将 WebKit 的xterm从WebGL改为默认渲染器，真实p95恶化至130ms，已撤销，源码不保留无收益改动；详见[WebKit小时报告](../../acceptance/reports/webkit-performance-hour-2026-09-22.md)。下一具体动作：只有在具备可区分WebKit绘制热点的独立profile时才继续性能代码优化；其余严格验收继续等待Windows、独立systemd、远端高RTT/丢包、物理故障和24小时环境，避免重复长测。

2026-09-22 E08 WebKit 一小时真实复验：官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 容器中的 WebKit 26.6 以私有 HTTPS 双 Daemon `--performance` fixture 运行 3,603.841 秒，8窗口/双PTY/10,000项目录/16MiB循环传输均持续完成；539轮传输、末轮目标、两路终端和实例状态均正常。但 71,496 个指针样本 p95 **78ms**，超过≤50ms阈值，Playwright 退出1；详见[WebKit小时报告](../../acceptance/reports/webkit-performance-hour-2026-09-22.md)。已删除容器、停止fixture并清理目录。该失败保留为非Chromium实际证据，不覆盖 Chromium 的48.1ms通过；下一具体动作：不重复本地长测，整理当前严格验收中只能在 Windows、独立systemd、远端高RTT/丢包、物理故障和24小时环境完成的入口与缺口。

2026-09-22 E01/F10 WebKit 真实补验：Fedora 宿主无法直接运行 Ubuntu fallback WebKit（缺 ABI 依赖），因此使用版本匹配的官方 `mcr.microsoft.com/playwright:v1.63.0-noble` 隔离容器；私有 HTTPS Master/双 Daemon fixture 上，以1秒间隔运行121点真实采样→Daemon离线 stale→Master重启缓存→Daemon恢复链路，WebKit **1/1通过**。容器按 `--rm` 删除，fixture Ctrl+C 后进程、9443监听和临时目录均清理；报告见[WebKit监控历史](../../acceptance/reports/webkit-monitor-history-2026-09-22.md)。这补齐 E01/F10 的 WebKit 浏览器证据，不替代 Windows、24小时墙钟、远端网络或高 RTT/丢包。

2026-09-22 B10 当前源码缺口复审：重新对照 `ACTION_GUIDE.md` 的 B00～B10 完成标准、两份规划和验收矩阵；生产代码扫描未发现 TODO、固定成功/假数据或未实现占位，`CAPABILITY_UNAVAILABLE` 仅出现在明确的能力边界（未开放通道、无容器运行时和由所属应用处理的重试）。A01～A08、A10～A12、A14～A17 的实现列为已实现，A09/A13 以及 F/E 整项的剩余条件均指向 Windows、systemd、远端网络、物理故障、长时或非 Chromium 环境。当前没有新的安全本地生产改动可补；继续重复现有短测不会改变验收状态。

2026-09-22 E05 私有 systemd 探针收口：在 `/tmp` 临时运行目录中尝试启动隔离 user manager；`unshare --user` 的 uid 映射和 `/proc` 挂载均被当前沙箱拒绝（`Operation not permitted`），直接启动私有 `systemd --user` 后也无法通过本地 user bus 连接（同样为 `Operation not permitted`）。未接触宿主服务、cgroup 或 D-Bus；探针目录、进程和 socket 已清理，复核无残留。E05 的真实 service/timer 生命周期仍需可操作的 systemd 环境，本机没有可再推进的安全路径。

2026-09-22 B10 测试夹具清理：复核无 `blora-devfixture`、Master、Daemon、Vite、Playwright 活动进程后，删除本任务生成的 120 个 `.local/fixture-*` 停止夹具目录；`.local` 从约 3.3GiB 可读残留降为 0，源码、`web/dist` 和 RC23 发行包未触碰。当前没有遗留 socket 或夹具会话。

2026-09-22 A09/E02/E08 网络边界复核：审查现有慢消费者、Docker HTTPS 代理和扩展远程目录测试后，确认本地可复用的证据已经覆盖有界日志/归档、控制请求不被阻塞、HTTPS Engine/Compose 传输，以及目录源的 80ms 延迟和证书轮换；没有把 loopback 代理伪称远端高 RTT/丢包。当前源码 `TestExtensionRemoteCatalogAdminAPI` 在沙箱外以 `go test -race` 复验通过（1.08s，退出0）；沙箱内失败仅因 `[::1]` loopback 绑定受限。由于本机没有 `tc/netem`、远端 Engine 或更高层网络故障环境，本地没有新的安全生产改动可补，剩余项转为相应平台验收手册入口。

2026-09-22 B10 环境复探与文档链审计：重新检查 `wine`/Windows 工具、systemd user bus、Docker endpoint、Chromium/Firefox/WebKit 和 `tc`；仍无 Windows、可操作 systemd manager、沙箱内 Docker socket 或 WebKit，Firefox 已有真实监控证据。验收矩阵、执行进度、README 和两份运维文档共 **5** 份 Markdown 的相对链接审计 **0** 条断链。没有因环境未变重复长时间测试。

2026-09-22 B10 本地收口审计：用户要求暂停视觉重构后继续核对可独立完成的功能。源码扫描未发现生产路径 TODO、固定成功响应或空实现；B07 的实例列表/卡片视图、批量操作、任务中心/站内通知，B08 的 Docker/Compose、监控、备份调度和有限系统管理，B09 的独立扩展包、沙箱、任务/资源桥接与迁移均已有实际入口和对应证据。本轮复跑 `cd web && npm run check`、`npm test -- --run` **63/63**、`npm run build` 和 `make check` 均退出0；当前源码 `make windows` 交叉构建也退出0。进度表中的旧“前端配置和模板正在接线”等描述已按当前源码修正；新增[平台验收运行手册](../../operations/PLATFORM_VALIDATION.md)，把剩余 Windows、systemd、远程 Engine 和长期故障入口固定下来。剩余工作主要是矩阵严格验收仍缺的真实环境证据，不能由本机替代测试冒充完成。

2026-09-22 RC23 当前源码发行复核：`BLORA_VERSION=development-20260922-rc23 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`sha256sum -c SHA256SUMS` 六项通过，SHA256SUMS SHA-256 为 `5208cc267914f11144df11586ec4f3a7c936d92ff0a02c590fe745cf00dd5f82`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/static/login、双 Daemon、停机快照恢复和 RC22 兼容回退；见[RC23发行报告](../../acceptance/reports/release-rc23-2026-09-22.md)。

2026-09-22 RC22 当前源码发行复核：`BLORA_VERSION=development-20260922-rc22 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`sha256sum -c SHA256SUMS` 六项通过，SHA256SUMS SHA-256 为 `eb6556276146b764795666924321d881b7c7dbdd0c8ad2e7f718459749b277cf`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/static/login、双 Daemon、停机快照恢复和 RC21 兼容回退；见[RC22发行报告](../../acceptance/reports/release-rc22-2026-09-22.md)。

2026-09-22 Docker 日志归档入口增量：后端已有的有界 `GET /api/v1/nodes/{id}/docker/containers/{containerId}/logs/history` 现在接入 `DockerLogs.vue`，实时窗口与归档窗口明确分离，归档显示保留上限、观察时间、截断/可能缺口标记，并可返回实时日志；读取失败不伪造历史。`npm run check`、前端单测 **63/63**、`npm run build`、`make check` 均通过；Docker 浏览器链路 **2/2（11.1s）**通过，覆盖实时日志入口、归档内容、缺口提示和返回实时状态；真实隔离 Docker Engine 的 Master/Daemon 生命周期测试 **2.79s，退出0**，新增验证归档由 Master API 读取、未授权403和 `limit=101` 400；真实 Chromium Docker 日志 UI **1/1（7.6s）**通过，覆盖真实容器启动、实时正文、归档正文、缺口提示、返回实时及停止/删除清理；原有真实 Compose 长流程 **1/1（45.6s）**复验通过，配置草稿、显式应用、阶段输出和清理链路无回归。详见[Docker日志归档报告](../../acceptance/reports/docker-log-history-2026-09-22.md)。这补齐了当前 Linux/浏览器/Docker 可推进的 F11 子项；远端 Engine、Windows 容器、主机掉电/存储耗尽和长时故障仍需相应环境。

2026-09-22 RC21 当前源码发行复核：预览模式控件禁用修正后的 `BLORA_VERSION=development-20260922-rc21 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `0c6c2a2d02c44c953d38c62aa1da9ed65b9759700c44be061b7a57eabaf86d9c`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC20 回退；见[RC21发行报告](../../acceptance/reports/release-rc21-2026-09-22.md)。本地可推进的文件/桌面增量已继续完成，剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合等环境证据。

2026-09-22 F08 大文件只读分段预览增量：新增 `GET /api/v1/instances/{id}/files/preview`，按节点确认的文件版本读取不超过60KiB的 UTF-8 分段，带边界重叠、分片校验、偏移/总量/下一段状态，不把完整大文件装入 Master 或浏览器；二进制/NUL/非法 UTF-8 仍明确拒绝并保留下载入口。编辑器在超过4MiB时进入只读预览，支持上一段/下一段，保存/撤销/编辑控件不会误作用于预览。真实 TLS 文件集成包含大文件首段/第二段与二进制拒绝，定向浏览器边界 **3/3（14.9s）**通过；`make check`、前端 **63/63** 单测、`npm run check`、`npm run build` 和 OpenAPI 解析 **105 paths / 122 operations**通过；改动已纳入 RC21。详见[编辑器能力增量报告](../../acceptance/reports/editor-capabilities-2026-09-22.md)。F08仍因完整跨应用/故障交叉及外部平台证据保持进行中。

2026-09-22 RC19 当前源码发行复核：设置页窄屏两列修正后的 `BLORA_VERSION=development-20260922-rc19 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `94590b0a100b2e1b60af204894fc8526752ff34d0729bf7264dd544ef65440fe`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC18 回退；见[RC19发行报告](../../acceptance/reports/release-rc19-2026-09-22.md)。当前本地可推进的功能与回归已经继续完成，剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合等环境证据。

2026-09-22 RC18 当前源码发行复核：加入快捷键偏好后的 `BLORA_VERSION=development-20260922-rc18 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；`npm run check`、前端单测 **63/63**、`npm run build` 和控制栏/设置浏览器回归 **3/3（9.9s）**通过；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `02f61fe468ba62d784df1d76f38cb79927cbd2685a4af3ba5c18624676d4c21c`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC17 回退；见[RC18发行报告](../../acceptance/reports/release-rc18-2026-09-22.md)。下一步剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合；这些不是本地代码停滞，而是当前环境没有对应运行条件。

2026-09-22 工作区偏好与布局保护增量：设置页补齐主题、字号、界面密度和减少动态效果的工作区持久偏好；新增“恢复默认布局”，先保存完整布局备份工作区，再重置窗口几何、吸附状态、最小化状态和桌面入口坐标，草稿/终端检查点/资源身份仍保留在备份副本。`npm run check`、前端单测 **63/63**、`npm run build` 和设置/布局浏览器回归 **3/3（9.3s）**通过。当前源码重新打包为 `development-20260922-rc17`，六包 SHA 校验及独立包冒烟通过，见[RC17发行报告](../../acceptance/reports/release-rc17-2026-09-22.md)。下一步剩余主要是 Windows 真机、systemd、远程网络、物理掉电和长时跨平台组合；这些不是本地代码停滞，而是当前环境没有对应运行条件。

2026-09-22 RC16 当前源码发行复核：`BLORA_VERSION=development-20260922-rc16 make package` 退出0，Linux/Windows Master、Daemon、SDK、Web 六包生成；逐项 `sha256sum -c SHA256SUMS` 通过，SHA256SUMS SHA-256 为 `c3199c01de67de1e8da40b677caefad8e7b1960d4a2f879fe875dc76e97906ca`。独立包级冒烟退出0，验证 SDK/参考扩展签名、Master TLS/登录、双 Daemon、停机快照恢复和兼容 RC15 回退；见[RC16发行报告](../../acceptance/reports/release-rc16-2026-09-22.md)。这轮没有停在 UI：当前源码中可本地闭环的文件目录树、列表/图标视图、列表框选、实例列表/卡片视图均已实现并有定向浏览器证据。下一步只继续处理尚无本地证据的 B07/B08 小缺口或整理全范围外部验证清单；Windows 真机、systemd、远程网络、物理掉电和长时跨平台仍需相应环境，不能用本地包冒烟替代。

2026-09-22 B07 实例中心视图增量：补齐规划要求的列表/卡片视图。两种排布共用实例筛选、选择、权限和幂等批量操作，视图选择写入标签现场并在刷新后恢复。新增报告[实例视图增量](../../acceptance/reports/instances-view-2026-09-22.md)。`npm run check`通过；`npm run test:e2e -- tests/browser/instance-batch.spec.ts` **1/1通过（13.8s）**，含切换、刷新恢复和部分接受批量请求回归。下一步继续审计 B07/B08 尚无真实证据的本地功能，不重复已通过套件。

2026-09-22 B06 文件管理器视图增量：补齐规划要求的图标/列表视图，并为列表空白区域加入框选。列表保持按滚动窗口读取的虚拟化路径，图标视图按最多100项分页并提供前后页；视图和页码写入现场，刷新后恢复。新增报告[文件视图增量](../../acceptance/reports/files-view-2026-09-22.md)。`npm run check`通过；前端单测 **63/63通过**；`npm run build`通过；`npm run test:e2e -- tests/browser/files.spec.ts` **5/5通过（19.0s）**，同时回归目录树、10,000项虚拟列表、框选、上传续传/取消和目录快捷方式。F07仍进行中：跨节点/物理故障/Windows场景按矩阵跟踪。下一步继续审计剩余 B07/B08 本地功能，不重复已通过文件套件。

2026-09-22 B06 文件管理器目录树增量：补齐规划要求的目录树入口。根目录及展开分支按需读取首分页块，未展开目录不递归请求；展开路径写入视图现场，点击目录复用路径历史，刷新后恢复树分支和当前路径。新增报告[文件目录树增量](../../acceptance/reports/files-tree-2026-09-22.md)。`npm run check`通过；`npm run test:e2e -- tests/browser/files.spec.ts` **5/5通过（16.7s）**，同时回归10,000项虚拟列表、上传续传/取消和目录快捷方式。F07仍进行中：完整跨节点/物理故障/Windows场景仍按矩阵跟踪。下一步继续按B07审计任务中心、托盘通知和后台任务关闭网页后的本地恢复证据，不重复已通过的文件套件。

2026-09-22 F03/F08 编辑器授权、边界与恢复增量（分段预览加入前的记录）：新增文件访问能力查询与 fail-closed 只读编辑器，补充 UTF-8/BOM、LF/CRLF、最大字节数元数据往返、扩展名语言映射、共享正文分栏与多光标刷新恢复；超 4 MiB/非 UTF-8 错误页提供原始下载。编辑器撤销日志现按最多 8 MiB/文件上限两倍执行；活动输入组在操作结束前不被切开，单组超预算时完整落为新基线，避免只撤销部分粘贴。修复前的实测确曾暴露“撤销后正文仍是粘贴尾部”的缺陷，修复后恢复单测 **14/14**、编辑器 Playwright 合组 **8/8（36.1s）**、`npm run check` 与 `npm run build` 通过；后端真实 TLS Master/双 Daemon 文件 API race **11.302s**通过，覆盖精确 4 MiB 可读、超限/二进制拒绝和源字节下载。B06 本机文件系统 race 套件 **1.332s**通过（符号链接/FIFO/目录交换隔离），Windows amd64 `internal/filesystem` 测试二进制交叉构建成功（5.9 MiB，未在 Windows 运行）。OpenAPI 3.1.0 当时为 **104 paths / 121 operations**。详细证据和限制见[编辑器能力增量报告](../../acceptance/reports/editor-capabilities-2026-09-22.md)。当时 F03/F08仍进行中：分栏固定50/50、编码只支持UTF-8/BOM、分段预览尚未实现；随后已由上方 F08 增量补齐，完整跨应用/故障组合和外部平台仍待验；视觉改动按用户要求冻结。

2026-09-21 E01 Firefox真实浏览器补验：`web/playwright.real.config.ts` 新增 `BLORA_BROWSER` 引擎选择（默认Chromium），`npm run check`通过。安装Playwright Firefox后，在隔离真实HTTPS双节点 fixture 上以1秒间隔跑完整121点采样→失联/stale→Master重启缓存→Daemon恢复链路 **1/1通过（2.9分钟）**。首轮 fixture 少显式 `-listen` 导致 supervisor 不能定位，已补参数复验；Firefox普通沙箱 profile错误后经批准沙箱外运行通过。Ctrl+C清理fixture，退出0，进程无残留，9443/9444不再响应。Windows、WebKit与24小时墙钟采样仍未验证；UI保持冻结。证据见[监控历史报告](../../acceptance/reports/monitor-history-soak-2026-09-21.md)。

2026-09-21 E01跨UTC日边界补验：扩展现有节点/实例历史测试，让125个带时间戳样本跨越午夜；最新120点从次日00:00:01 UTC开始并有序保留至00:02:00 UTC。定向 `go test -race` 两项测试通过（2.718s）。这验证时间戳 key 跨日排序和裁剪，不代表运行24小时以上的实时采样；后者、Windows和非Chromium仍缺。沙箱拒绝 HTTPS fixture 所需 IPv6 loopback 后，获准在沙箱外执行相同测试并通过。UI保持冻结。

2026-09-21 E03/F12 备份恢复进程中断：新增 `internal/backup/restore_crash_test.go` 和[恢复进程崩溃报告](../../acceptance/reports/backup-restore-process-crash-2026-09-21.md)。独立恢复子进程在真实 8 MiB 文件已有1,048,576B写入目标暂存区后被强制终止；重开原备份/目标状态后，原ID明确停在 `interrupted`，已完成目录作为部分进度保留、上传取消、暂存清除、正文未发布；同ID重放不变，生成新计划后重新恢复并核对清单SHA成功。定向 race 连续5次12.408s、备份全包 race 6.864s、`go vet ./internal/backup`与Windows amd64测试二进制交叉编译通过。只证明Linux进程级恢复，不覆盖物理掉电和Windows运行；UI保持冻结。下一步按矩阵继续找当前环境可真实执行、且不重复既有套件的验收缺口。

2026-09-21 E01 节点失联及 Master 重启缓存组合补验：fixture Daemon 在121点真实采样后按 `/proc` 身份精确 `SIGSTOP`，Master判为 `OFFLINE` 后仍读到120个 `stale` 点；测试再经fixture supervisor专用`SIGUSR1`请求优雅重启 Master，核实进程 PID 更换、健康检查/重新登录成功，原 Daemon 仍暂停时缓存的120个 stale 点及全部采样时间戳保持一致。随后 `SIGCONT` 原 Daemon并等待重新上线。1秒间隔组合真实Chromium 1/1通过（2.7分钟）；`cmd/devfixture` Linux编译测试、Windows amd64交叉构建、`web/npm run check`通过。SIGINT清理 fixture 后进程及9443/9444端口无残留。证据见[长时监控报告](../../acceptance/reports/monitor-history-soak-2026-09-21.md)，F10/E01矩阵已更新。此项不覆盖Windows运行或跨日采样；UI保持冻结。

2026-09-21 E01 本机真实监控历史长采样：新增默认跳过、需 `BLORA_E01_HISTORY_SOAK=1` 显式开启的 `web/tests/real/monitor-history-soak.spec.ts`。在隔离 HTTPS 双节点 fixture 上按5秒间隔调用真实节点指标121次；每点非stale且时间严格递增，最终历史恰保留最新120点，首点等于第2样本、末点等于第121样本。沙箱外真实Chromium 1/1通过，10.1分钟；`web/npm run check`退出0。证据见[报告](../../acceptance/reports/monitor-history-soak-2026-09-21.md)，验收矩阵F10/E01已更新。SIGINT停止专用fixture，进程/9443/9444端口复核无残留。此项只补约10分钟Linux在线采样，不覆盖跨日、Windows、非Chromium或Master重启后持久缓存读取；UI保持冻结。下一步继续处理剩余跨平台与外部环境验收，不重复RC15已通过套件。

2026-09-21 E04/F12 实际 Scheduler.Tick 阶段崩溃：新增 `internal/backup/scheduler_stage_crash_test.go` 与[阶段崩溃报告](../../acceptance/reports/scheduler-staged-process-crash-2026-09-21.md)。独立子进程分别在 pending slot 已持久化但尚未 BuildTask、prepared fire/taskId/requestId 已持久化但尚未 Accept 时被 SIGKILL；SQLite/WAL 重开后，前者恰好构造并接受一个任务，后者复用原 taskId/requestId且不重建，重复 Tick 均不重放。定向 race 两场景通过（包1.224s），backup 全包race 8.017s、`go vet ./internal/backup`通过；Windows amd64 测试二进制交叉编译成功。它们覆盖实际调度代码的提交阶段之间，不是设备写入/SQLite事务内部的掉电测试。

2026-09-21 E09 Windows 交叉编译补核：HTTPS Docker E2E 测试辅助代码连同 `internal/containers` 测试包通过 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c`，产物 `/tmp/blora-containers-remote-tls.test.exe`。storage/backup 测试包本轮也均已交叉编译成功。上述只证明测试代码可构建，不是 Windows Docker/Job/ConPTY 真机验收。

2026-09-21 E02/F11 HTTPS Engine 传输子项：E2E 新增 `BLORA_CONTAINER_E2E_REMOTE_TLS=1`，自动生成短期测试CA，在 loopback TLS 代理与本机 Docker Unix socket 间转发；Master Engine 客户端与 Compose CLI 均走真实 HTTPS/TLS。Docker Engine 29.4.1 上生命周期 race E2E 30.49s、包31.52s通过，覆盖容器/镜像/卷/网络、双流日志、Compose更新删除、健康失败、数据保留及tmpfs ENOSPC；随机标签资源通过测试 defer 清理。`go vet ./internal/containers`通过。此项证实 HTTPS 传输与 Compose TLS，不等同远端主机或公网测试；见[HTTPS传输报告](../../acceptance/reports/docker-https-transport-2026-09-21.md)。

2026-09-21 E09 Windows 测试构建增量：新增 crash test binaries 均成功交叉编译：`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ... ./internal/storage` 与 `./internal/backup`，产物位于 `/tmp/blora-storage-crash.test.exe`、`/tmp/blora-backup-archive-crash.test.exe`。只证明新增测试代码可编译到 Windows，未执行真机行为。`tc` 不在当前环境，暂不能用 `netem` 做真实内核 RTT/丢包；后续核对可用隔离 Engine/代理是否能提供不重复的真实远端组合。UI保持冻结。

2026-09-21 E03/F12 备份归档中断恢复：新增 [归档进程崩溃报告](../../acceptance/reports/backup-archive-process-crash-2026-09-21.md) 与 `internal/backup/archive_crash_test.go`。子进程以真实 `Manager.Create` 打包 64 MiB 源文件，父进程观察 pending ZIP 写入1,048,627B时 SIGKILL。用原 ID 重试后返回 `interrupted/unknown`、未发布快照并清除部分 ZIP；再用新 ID 完成备份且 Inspect 验证通过。定向race通过（2.26s），备份全包race 6.524s及 `go vet ./internal/backup`通过。未模拟掉电/设备缓存丢失或 Windows。下一步继续核查本机可运行的远程/系统环境模拟价值，避免重跑已有浏览器套件；UI保持冻结。

2026-09-21 E04/F12 SQLite 事务进程崩溃边界：新增 [元数据事务崩溃报告](../../acceptance/reports/metadata-transaction-crash-2026-09-21.md) 与 `internal/storage/metadata_transaction_crash_test.go`。独立子进程经真实 `Store.MetadataMutation` 在同一 WAL 事务中写 schedule/fire 两项后停在 commit 前，父进程 `SIGKILL`，重开数据库确认业务记录、幂等回执、审计均未部分落盘，既有管理员记录保留，`integrity_check=ok`。定向race通过（用例0.09s），storage全包race 20.570s、`go vet ./internal/storage`通过。该证据只覆盖共用存储事务实现，不覆盖实际 Scheduler.Tick 回调/设备掉电；这些与长期墙钟仍待验。下一步继续挑选可在当前Linux环境真实执行的剩余故障组合；UI保持冻结。

2026-09-21 A09/E01 验收续接：A09 当前源码真实 TLS 慢日志消费者 30 秒逐秒采样通过；未消费字节峰值33,048B/262,144B，接收队列峰值28,320B/2,097,152B，随后跨节点传输、5秒内停止、归档重读与目标校验均通过，`go vet ./internal/master`通过。60秒探索超过协议45秒慢消费者截止期，连接按预期关闭，不计通过；已有测试验证诊断与游标重连，证据见[30秒报告](../../acceptance/reports/slow-consumer-soak-2026-09-21.md)。E01 源码与现有真实浏览器证据已核对：节点/实例指标、进程树、120点有界历史、失联 stale、重连采样和搜索/排序/授权终止均已有覆盖；短时本机重复采样无新增价值，跨平台及长期监控专项仍待补。下一步优先在可用平台/远端测试环境补验；UI继续冻结。

2026-09-21 E04 发行包进程级崩溃复验：扩展 `scripts/package-smoke.py --scheduled-backup-crash`，用 RC15 Master/双 Daemon 和真实分钟计划创建定时备份；任务接受后同时 `SIGKILL` Master 与所属 Daemon，再用原数据目录重启。目标 Daemon startupId 更新，原 schedule/taskId/requestId 保持，备份最终SUCCEEDED且归档唯一1份，脚本退出0。初始 harness 的等待函数作用域及 PUT 方法问题均已修正，最终证据见[Master/Daemon崩溃报告](../../acceptance/reports/scheduled-backup-master-daemon-crash-2026-09-21.md)。这补足进程级全服务崩溃后的 accepted occurrence 恢复，不覆盖掉电或事务中途断电；长期墙钟和Windows/systemd仍待补，UI继续冻结。

2026-09-21 E02 真实 Docker ENOSPC 子场景：在 `TestRealDockerComposeLifecycle` 中为随机标签的 Compose 服务配置 1 MiB tmpfs，容器实际写入触发内核 `ENOSPC`；当前源码真实 Engine/Compose race E2E 29.17s 通过。确认 apply 返回 `FAILED/health`、exit code 1、`Unknown=false`、保留容器事实与 “No space left on device” 日志，然后通过 Compose delete 清理。`go test -race ./internal/containers -count=1` 在获准 loopback 环境8.248s通过，`go vet ./internal/containers`通过；普通沙箱首轮因 httptest IPv6 loopback 权限失败，不是测试断言失败。报告见[Docker ENOSPC 验收](../../acceptance/reports/docker-enospc-2026-09-21.md)，矩阵E02/F11已更新。此项只覆盖容器 tmpfs，不代表 Engine 主机存储耗尽、远程 Engine、掉电或 Windows 容器已通过；UI继续冻结。

2026-09-21 E08 CPU热点优化与一小时复验完成：60秒 CPU profile 将 `recovery/state.ts` 的 JSON 深拷贝定位为主要应用热点；内部恢复值复制改为 `structuredClone`，Vue代理/不可克隆值仍回退 JSON，`json()` 保留原清洗语义。`web/npm run check`、前端单测 **50/50**、生产构建通过；桌面/文件/PTY/编辑器/恢复/云工作区定向真实浏览器 **27/27**。无 profiler 短时真实混合场景 64.954秒 **1/1**，p95 **45.2ms**、max71.9ms、3/3565长帧、双PTY约112kB/s、万条目录/16MiB校验传输完成、未确认峰值79,006B。完整一小时真实混合场景 **1/1通过**：3,604.492秒、48,840个响应样本、p95 **48.1ms**、max160.4ms、731/197,137长帧，8窗口/双PTY/万条目录、165轮传输均完成；队列峰值86,180B，堆35.8–124.6MB。62组RSS/归档采样中57组在浏览器负载期；Master/Daemon RSS峰值39.4/32.0MB，单会话归档最大16,774,920B（低于16MiB预算），无采样竞态。报告与原始指标见[恢复克隆优化小时报告](../../acceptance/reports/performance-hour-recovery-copy-2026-09-21.md)和[JSON](../../acceptance/reports/performance-hour-recovery-copy-2026-09-21.json)。夹具/浏览器/采样器退出并清理，9443/9444空闲。本机E08一小时p95目标已达标；远端高RTT/丢包、Windows及非Chromium尚未覆盖。此前RC14发行包不含本轮克隆优化；UI按用户要求冻结。

2026-09-21 RC15 交付复核：`BLORA_VERSION=development-20260921-rc15 make package` 退出 0，Linux/Windows Master、Daemon、SDK、Web 六包 `sha256sum -c SHA256SUMS` 全部通过，SHA256SUMS SHA-256 `b5aa3b0c1ec273010b8439521678aa880a1ce133d181320861a670b7f57ada2d`。独立包冒烟在沙箱内首轮受 `esbuild` `EPERM` 阻止；同一测试在沙箱外重跑退出0，SDK/参考扩展签名、Master初始化/TLS/登录、双Daemon、停机快照恢复与兼容RC14回退均通过。见[RC15发行报告](../../acceptance/reports/release-rc15-2026-09-21.md)。私有冒烟目录 `/tmp/blora-release-smoke-pmravw8_` 保留，所有服务/测试进程已停止。Windows真机/systemd等外部验收仍未覆盖。

2026-09-21 RC15 当前源码普通真实浏览器全套：`fixture-2757846034`（真实 HTTPS Master/双Daemon/本机 Docker Engine）串行运行49项，排除单独的一小时场景，Playwright **49/49 passed**，退出0，11.5分钟。账号/备份/调度/Docker/Compose/桌面/实例/扩展/文件/PTY/通知/恢复/云工作区均通过；续传从1,179,648B确认偏移恢复16MiB上传且10,000项解压完成。fixture收到Ctrl+C并清理，Master/Daemon/Playwright及9444端口无残留；完整命令与边界见[RC15浏览器回归报告](../../acceptance/reports/browser-regression-rc15-2026-09-21.md)。独立E08小时测试另见上条。

2026-09-21 A09 RC15 当前源码定向复验：`GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-grant-idem GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop$' -count=1` 退出0（13.918s）。普通沙箱因 HTTPS 测试监听无法绑定 IPv6 loopback 而失败，获准在沙箱外重跑同一命令后通过。真实TLS慢日志消费者/8.4MB跨节点传输/停止及目标与归档校验子场景当前源码未回归；长时全连接和跨平台资源验证仍缺，见[混合流报告](../../acceptance/reports/mixed-stream-control-2026-09-13.md)。

2026-09-21 E04 定时调度强制终止子项：新增独立 Go 测试进程，在计划 occurrence/task/receipt 持久提交后由父进程 `Process.Kill`；另一个进程重开原 SQLite，核对 accepted fire、taskId/requestId、schedule LastTaskID/slot 不变、旧 occurrence 不重放，并确认撤权后下一触发不创建任务。定向 race 1.130s、备份全包 race 2.351s、Windows amd64 测试程序交叉编译通过；最终 `make check`（`go vet ./...`）退出0。全包首轮露出测试输入受当前 `umask 0077` 影响（期望0640、实际0600），已在该测试显式Chmod固定，不改生产行为。证据见[E04进程崩溃报告](../../acceptance/reports/scheduler-process-crash-2026-09-21.md)。这不是事务中途掉电或 Master+Daemon 整体硬崩溃验证；物理掉电、长时墙钟、Windows/systemd/远程环境仍缺。UI保持冻结。

2026-09-21 可执行环境与验收矩阵对照完成：当前 runner 为 Linux x86_64；`wine`、`firefox` 不存在，Playwright 缓存只有 Chromium，因此 Windows 真机及非 Chromium 验证不能在此环境执行。`systemctl` 命令存在，但可操作 systemd manager 的缺失已由[环境报告](../../acceptance/reports/systemd-environment-2026-09-20.md)记录。RC15 49/49、E08 一小时、A09 定向、E04 scheduler crash，以及本地 Docker、备份/ENOSPC、firewalld、扩展目录等可执行子项均有当前源码证据；未发现应为形式而重复的本机短测。仍缺 A09/E01/E04 的长期与跨平台证据、E02 远端 Engine/长时故障、E03/E04 物理掉电、E05 systemd/Windows 管理器、E06/E07 真实远端运营和 Windows、E08 高 RTT/丢包/Windows/非 Chromium、E09 Windows 真机与 systemd。下一步应直接补入可用的 Windows、systemd 和独立远端测试环境，再针对未覆盖组合验证；若环境仍未提供，则继续做有明确收益的长期故障场景，不重复已通过套件。当前无测试子进程或 fixture 在运行，UI继续冻结。

2026-09-21 RC14 当前源码常规真实浏览器套件：隔离 `--performance --docker-endpoint=unix:///var/run/docker.sock` fixture 上串行运行49项（排除单独的一小时性能测试），Playwright **49/49 passed**，退出0，11.4分钟。覆盖账号/权限、备份/真实分钟调度、Docker/Compose、桌面与实例控制、扩展安装/WASI、文件/PTY/日志/传输、监控/通知、恢复故障注入和工作区等；完整命令及边界见[浏览器回归报告](../../acceptance/reports/browser-regression-rc14-2026-09-21.md)。fixture、容器任务和浏览器均已停止；`.local/fixture-1611252974`已清理，定向进程检查无 Master/Daemon/Playwright 残留。此次不覆盖 E08 单独一小时门槛，也不验证 Windows 真机、systemd、远程高延迟或物理故障。

2026-09-21 E08 拖窗优化尝试已回退：13项真实窗口几何测试通过，但真实60秒混合场景在两种直接DOM样式写入变体下 p95 分别为57.0ms和54.9ms，均超50ms；因此未保留该改动。恢复原RAF响应式更新后 `npm run build` 退出0，最终源码继续以RC14 49/49回归所覆盖的窗口实现为准。失败实验和环境见[拖窗优化尝试报告](../../acceptance/reports/performance-drag-experiment-rc14-2026-09-21.md)。性能fixture与Playwright均已停止。

2026-09-21 当前源码交付复核：`BLORA_VERSION=development-20260920-rc14 make package` 退出0，Linux/Windows Master 和 Daemon、SDK、Web 六个归档均通过 `sha256sum -c SHA256SUMS`。包级独立冒烟 `python3 scripts/package-smoke.py dist/releases/development-20260920-rc14 --state-restore --rollback-release dist/releases/development-20260920-rc13` 退出0：SDK示例签名、Master初始化/TLS/前端/登录、双Daemon上线、停机快照中身份/权限/任务回执/资源正文恢复以及兼容rc13包回退均通过。服务已停止，无测试进程残留；私有诊断目录保留于`/tmp/blora-release-smoke-h9uvvo0v`。证据见[rc14报告](../../acceptance/reports/release-rc14-2026-09-21.md)，SHA256SUMS哈希`46cb8a9d566f0158708e9f68741131aad5c2c96d7731a80ea2b18da7b5c20768`。此包已包含rc13之后的恢复日志/终端快照性能修复，但未代表Windows运行验收或全范围完成。UI按用户要求保持冻结。

2026-09-21 E08 当前源码一小时复验已结束但未通过：3606.506秒、45,072次响应采样，p95 **51.7ms（目标≤50ms）**、max159.4ms、1312/190831长帧；8窗口、双PTY、万条目录和16MiB循环传输运行完成，162轮传输及末轮目标校验通过。未确认队列峰值86,193B，浏览器堆样本36.9–135.9MB，归档最高观测16,775,268B后持续轮转；RSS/归档为定期样本而非瞬时上界。唯一失败为p95断言；不能用旧源码通过结果覆盖。完整环境与数值见[小时报告](../../acceptance/reports/performance-hour-rc14-2026-09-21.md)及[原始JSON](../../acceptance/reports/performance-hour-rc14-2026-09-21.json)。当前源码两次60秒诊断1/1通过但不计性能验收：p95均47.4ms，分别14/3468、11/3634长帧；慢输入同时来自事件前排队和后续渲染，恢复日志同步`setItem` p95低于1ms，profile热点分散于IndexedDB、xterm、列表更新和GC，详见[诊断报告](../../acceptance/reports/performance-diagnostic-rc14-2026-09-21.md)。直接DOM拖窗改动实测p95为57.0/54.9ms，未达到目标并已回退，详见[实验报告](../../acceptance/reports/performance-drag-experiment-rc14-2026-09-21.md)。所有浏览器和fixture均已停止。E07 CA根切换定向race已通过（测试1.20秒/包2.240秒）：旧根信任新CA时失败关闭，显式换根后安装新版本；详见[CA根轮换报告](../../acceptance/reports/remote-catalog-root-rotation-2026-09-21.md)。下一步继续定位可重复、可量化的非视觉优化；未取得短时达标证据前不重跑小时场景。UI按用户要求保持冻结。

2026-09-20 E08恢复写放大优化：修复 `flush()` 在旧快照后无预算校验写回超256KiB恢复尾日志的竞态；IndexedDB快照追赶期间遇到超额增量时保持未保护提示并立即提交最新状态。新增 `terminal-output` 增量恢复操作，旧 `outputJournal`检查点仍兼容，避免每个终端块复制完整累计输出数组。真实混合负载profiling指出xterm全屏序列化仍是主要热点后，将快照日志预算从64KiB调整为128KiB，仍限定128条事件，逐条输出依旧同步提交，ACK顺序不变。`npm test` 49/49、`npm run check`、`npm run build`通过；fake-indexeddb并发超预算/旧检查点连续追加测试通过；真实WSS PTY刷新/移窗1/1（18.7s），无输入重放。当前干净8窗口/双PTY/万条目录/16MiB传输样本64.63s，交互p95 **45.1ms**、max72.9ms、2/3636长帧、双PTY约114KB/s，性能测试1/1通过。带诊断的上一样本日志`setItem`累计字符由约1.57亿降至2,421万，时间2.96s降至0.57s；诊断扰动不算验收。fixture `fixture-2507872921`、`fixture-1933717844`、`fixture-3066140007`均已停止并清理，当前无测试/fixture进程。UI保持冻结。下一步继续处理其它可完成的功能块与外部验收缺口；E08需在最终源码上补长期复验，Windows/独立远程环境及掉电仍不能由本机短测替代。详细数据见[持续性能报告](../../acceptance/reports/performance-continuous-2026-09-19.md)。

2026-09-20 E08 CPU采样跟进：带 profiler/长任务观察器的60秒诊断p95 53.2ms、max127.5ms、3个55/66/59ms长任务、7/3506长帧，API RTT p95 17.5ms；慢样本同时包含事件处理前排队与处理后等待渲染。采样热点为同步 `setItem` 3.09s、xterm `_nextCell` 2.04s/`serialize` 1.89s、IndexedDB `put` 1.70s、`_diffStyle` 1.39s及虚拟列表 `replaceChildren`/`createRow`。它们对应恢复日志、终端缓冲/快照和列表渲染路径，但工具本身扰动测量，不能断定唯一根因或作验收。双PTY约114KB/s、4次核验传输，队列有界且无pageerror；fixture已退出清理。未改生产逻辑或UI。下一步在保留同步可恢复语义及存储失败停ACK保证的前提下，量化RecoveryService每次提交日志写入成本并寻找可安全降复杂度的点；E08仍进行中。见[持续性能报告](../../acceptance/reports/performance-continuous-2026-09-19.md)。

2026-09-20 当前源码真实浏览器全套增量：排除独立性能场景后 Playwright 实际收集49项（矩阵旧计数48已修正）。本轮同夹具完整串行47/49通过；两失败分别是监控失联用例未等待首个实时进程列表，以及上传续传用例缺少 `--performance` 数据。监控用例添加“先成功读取实时列表/最近成功时间/进程行”的前置断言后，在真实SIGSTOP/CONT双节点夹具复验1/1（54.0s）；上传关闭续传改用性能夹具后复验1/1（1.2m），1万文件解压和16MiB上传从655,360B确认偏移续传均通过。因此每个常规场景都有一次通过证据，但未在一次全套运行中达到49/49；不为形式目标重复整套测试。两夹具和所属测试实例均已停止/清理。报告索引见验收矩阵；监控证据更新见[E01](../../acceptance/reports/monitor-offline-2026-09-19.md)。未改 UI 或生产逻辑。

2026-09-20 E08短时性能诊断复测：将主线程长任务/慢输入采集改为 `BLORA_PERF_DIAGNOSTICS=1` 显式启用，默认保持原低开销路径；`web/npm run check` 通过。带诊断样本p95 54.8ms、最大87.3ms，3个54–73ms长任务，最慢输入主要延迟在事件处理器之前；不作为验收基线。随后全新夹具未启用诊断复跑真实8窗口/双PTY/万条目录/传输60秒混合负载：65.522s、p95 54.4ms（未达50ms目标）、最大80.6ms、14/3478长帧、双PTY约112KB/s、队列峰值67,680B、两次16MiB目标校验通过。唯一失败为p95阈值；无pageerror。真实浏览器/夹具会话均已退出并清理。本轮只改性能测试诊断及证据，不动生产逻辑或UI。见[持续性能报告](../../acceptance/reports/performance-continuous-2026-09-19.md)。下一步按矩阵归纳剩余验收，并优先继续当前环境可完成的 Linux 真实组合；E08需定位输入排队来源，原一小时46.5ms证据与本次短时失败并存。

2026-09-20 F09 系统通知真实浏览器补验：新增真实 TLS 双节点任务通知用例；headed Chromium/Xvfb 在授予来源通知权限后收到真实 `Notification` 对象，核对标题/正文/tag，并派发 click 事件后打开精确 taskId 资源窗口，最终1/1（3.1s）通过。首次只因两个“任务中心”标题使通用定位严格模式冲突，收窄至 `data-window-mode="resource"` 后复验通过。报告：[系统通知 API 验收](../../acceptance/reports/system-notifications-2026-09-20.md)。虚拟显示无法验证桌面通知中心的实际绘制或物理点击，故F09仍进行中；用户要求暂停的视觉样式未改。fixture进程已退出，两个本轮创建的临时夹具目录已清理。

2026-09-20 E07远程注册源受控故障补验：增强 `TestExtensionRemoteCatalogAdminAPI`，Master 从 loopback HTTPS 源读取清单并安装参考包；每次注册源响应延迟80ms，浏览与安装之间续期同一受信CA签发的服务端叶证书，并关闭空闲连接强制再次握手。清单读取和安装均断言经过受控延迟，成员读取被拒绝；`GOCACHE=/tmp/blora-go-build GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOFLAGS=-buildvcs=false go test -race ./internal/master -run '^TestExtensionRemoteCatalogAdminAPI$' -count=1 -v` 退出0，race测试总耗时1.986s。详细范围和限制见[远程注册源验收](../../acceptance/reports/remote-catalog-rotation-2026-09-20.md)。这不是公网远端、高RTT、CA根更换或 Master 二进制参数的验收；E07继续进行中。未改生产逻辑或 UI。

2026-09-20 当前源码 rc13 回归：按用户要求暂停 UI 样式工作，只给新版应用市场卡片/已安装项和窗口增加非视觉 `data-app-id`/`data-window-mode` 识别属性，并将真实测试同步到现有标题、版本和账号菜单。`npm run build` 通过；真实双节点常规浏览器套件（排除独立一小时性能场景）48项首轮43项通过、5项因旧断言/测试数据隔离失败，修正后受影响6项定向复验6/6通过，因此48个常规场景均至少各有一次通过证据，但没有单次48/48的全绿套件运行。`development-20260920-rc13` 六包构建、SHA校验、停机状态恢复及兼容rc12回退冒烟退出0；Docker浏览器2/2通过。fixture和测试Docker对象已清理。详见[rc13报告](../../acceptance/reports/rc13-current-source-2026-09-20.md)。矩阵整体仍未完成；下一步只能在真实 Windows、独立 systemd、远端故障及物理掉电环境补证，不能以本地替代测试覆盖。

2026-09-20 最新现代 Material 3 Expressive 整理：壁纸替换为青灰抽象曲面，主题改为中性雾灰/矿物青绿并收敛图标色彩；统一菜单、弹窗、表单、卡片和编辑/终端外框，清理 Docker 局部旧色。修复 JS 启动器未遵循减少动态效果的问题，新增双来源偏好实测。完整浏览器38/38通过（1.9分钟），最后焦点样式修正另3/3通过（10.8秒），最终构建通过；真实截图34张，13组主要布局无差异，Dock四边14px。见[现代视觉报告](../../acceptance/reports/material-modern-2026-09-20.md)。源码/web/dist最新；所有进程结束，fixture 83928退出0；旧rc11未重打包。后续以2026-09-20-material-modern为视觉基准，保留布局；外部平台验收缺口及第三方应用内容边界见报告。

2026-09-20 最新 Material 3 细化：联网核对官方 Shape/List Expressive/Checkbox 文档，补齐内容区、表头、分组列表与复选框的圆角层级，调整备份/终端/设置/市场图标配色。构建通过，相关浏览器17/17通过（53.6秒，API替身），真实本地截图24张及空格选择/取消验证通过；13组主要布局无差异，Dock四边14px。见[细化报告](../../acceptance/reports/material-refine-2026-09-20.md)。源码及web/dist最新；所有构建/测试/截图会话结束，fixture 85791退出0，旧rc11未重打包。下一步视觉调整以2026-09-20-material-refine为基准；保留布局、多色扁平图标及既有外部验收边界。

2026-09-20 最新 Material You 重构：按用户“布局不变、整套改扁平、多色图标”的要求，新增统一颜色角色，改写桌面/窗口/侧栏/菜单/表单样式及整套图标，删除玻璃材质文件，替换平面几何壁纸。最终构建退出0，完整浏览器37/37通过（1.8分钟，API替身），真实Master/双Daemon截图22张。13组桌面/手机区域稳定态布局比较无差异，Dock四边14px、曲线误差小于0.01px；首次布局采样的入场动画差异已保留并修正采样后复验。见[本轮报告](../../acceptance/reports/material-you-2026-09-20.md)。源码和web/dist最新，构建/测试/截图全部结束，fixture 78272退出0，旧rc11未重打包。下一步视觉调整以2026-09-20-material-you为基准，保留用户要求的布局；全范围外部验证仍按已有环境缺口处理。

2026-09-20 最新薄片图标修订：按用户新提供的照片/App Store/邮件近景，去除此前被否定的追光、光晕和厚重高光，重新绘制薄片交叠前景、平缓底色及轻微接触阴影。最终构建退出0，控制栏/桌面浏览器7/7通过（44.6秒，API替身）；真实隔离Master/双Daemon截图22张，7种场景的Dock四边均14px、曲线误差小于0.01px。见[本轮报告](../../acceptance/reports/thin-glass-icons-2026-09-20.md)。源码和web/dist已更新；所有本轮构建/测试/截图会话结束，fixture 94030退出0，旧rc11未重打包。下一步若继续视觉调整，以2026-09-20-thin-glass截图与源码为基准；全范围外部验证仍按既有环境缺口处理。

2026-09-20 最新圆角对齐：依据 Apple 连续曲率/同心嵌套文档，应用图标和 Dock 共用连续轮廓，外框沿图标轮廓法线等距外扩14px；移除透明SVG留边对Dock布局的干扰，上下左右可见间距统一，悬停和窄屏保持关系。最终构建通过，控制栏/桌面浏览器7/7通过（38.5秒）；真实fixture直接测量7种场景的渲染轮廓，四边均14px、曲线偏差小于0.01px，无畸变或越界。截图5张与数值见[圆角报告](../../acceptance/reports/concentric-corners-2026-09-20.md)。源码及web/dist已更新，截图和fixture会话70529已结束，旧rc11未重打包。下一步如继续视觉调整，以此共享轮廓与截图为基准；全范围外部验证仍按原环境缺口处理。

2026-09-20 最新轻拟物修订：按用户最新图例重写应用图标为简洁的玻璃叠层，线条工具图标改用精确锁定的 @lucide/vue 1.47.0；账号菜单文字增加1px、宽度240→232px、行高37→35px，内容间距小幅收紧。最终生产构建通过，选定控制栏/桌面/文件浏览器11/11通过（52.9秒，API替身）；真实本地Master/双Daemon截图17张，见[本轮报告](../../acceptance/reports/glass-icons-2026-09-20.md)。源码与web/dist已更新，旧rc11归档未重打包；截图浏览器和fixture会话44040已退出0，没有待接管测试进程。下一步若继续调视觉，以本轮截图和源码为准；全范围剩余验证仍需先前记录的专用外部环境。

2026-09-20 最新桌面视觉修订：按用户对顶栏/旧macOS仿制感的反馈，新增独立DesktopBar组件和三列排版，统一工具图标尺度，账号菜单收纳退出/设置；移除三色窗口控制，重写窗口与Dock光影、背景及应用材质，调整共用窗口工作区域。build和47单测通过；本轮选定16浏览器场景经15+1修复复验通过。真实fixture截图14张，见[本轮报告](../../acceptance/reports/modern-desktop-2026-09-20.md)。新源码与web/dist已更新，截图fixture交付前停止，旧归档未含本轮UI。

2026-09-20 原生桌面 UI 重写：依据用户后续反馈移除旧 style.css 内容及 desktop-light.css 补丁层，重写桌面/Dock/窗口/应用样式；具象SVG应用图标、左侧三色窗口控制、无常驻标题Dock、资源侧栏与紧凑列表已落地。构建与47单测通过，浏览器首轮35通过+同名入口定位修复后1复验通过，36个场景全部验证。真实隔离Master/双Daemon截图11张及证据见[UI报告](../../acceptance/reports/native-ui-2026-09-20.md)。本轮测试进程已结束，截图fixture在交付前停止；源码与web/dist最新，rc11归档仍是旧UI。下一步如用户要求打包，基于当前源码生成新版归档；原环境阻塞项未改变。

2026-09-20 UI 重构：按用户反馈改为浅色桌面、蓝色操作重点、白色窗口与居中浮动任务栏；移除桌面营销文案，统一桌面/启动器/标题栏 SVG 图标，Monaco 使用浅色主题。原桌面 slice(0,6) 隐藏了扩展入口，现显式固定七个入口并将扩展管理更名为应用市场；市场展示真实注册源版本、已安装应用与本地安装，保留管理员权限和安装生命周期。当前源码及 web/dist 已更新，旧 rc11 归档未重新打包。

本轮验证：`cd web && npm run build` 退出0；单测47/47；`npm run test:e2e -- tests/browser/desktop.spec.ts tests/browser/extensions.spec.ts --workers=1` 最后整组11通过、1因上传输入框名称变化失败，恢复 aria-label 后以 `--grep 'independently packaged'` 单独复验1/1通过（合计12个场景通过，并非一次整组12/12）。初轮应用按钮模糊定位冲突已改为启动器专用定位。Chromium连接真实本地Master/Daemon fixture，在无API模拟的情况下保存[九张截图](../../acceptance/screenshots/2026-09-20-redesign/)，脚本为 `web/capture-ui.mjs`；市场中的条目为fixture参考扩展，不代表公共应用生态。最后截图后只新增上传框可访问名称，不影响画面。原Windows/systemd等环境验证缺口保持不变；后续如交付安装包，需基于此源码重新打包。

最新交付rc11：包含当前生产源码、前端回归修复、Docker 真实复验及同步后的验收台账；六包外部SHA/归档内MANIFEST、独立SDK签名/Master初始化TLS/双节点冒烟全部通过，包级冒烟退出0。SHA256SUMS 哈希与六个归档逐项结果见[rc11报告](../../acceptance/reports/release-rc11-2026-09-20.md)。

2026-09-20 续接审计：当前树重新执行 `make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及 47/47 单测退出0。源码扫描未发现生产路径空实现或遗留 TODO；当前无活动测试、服务或包冒烟进程。

2026-09-20 文档证据审计：检查 112 份 Markdown 的相对链接，缺失链接为 0；修正一份历史报告对已清理截图的失实链接并明确其证据边界。

2026-09-20 当前源码真实 Docker 复验：确认 Engine 29.4.1/overlayfs 与隔离 fixture 可用，`BLORA_CONTAINER_E2E=1 go test -race ./internal/containers -run '^TestRealDockerComposeLifecycle$' -count=1` 退出0（27.681s）；真实容器、镜像、卷、网络、双流日志、Compose 更新/删除、健康失败、卷保留和普通用户拒权均通过，带标签对象已清理。见[当前 Docker 验收](../../acceptance/reports/docker-current-2026-09-20.md)。

2026-09-20 systemd 环境复核：宿主 systemd user bus 返回 `Operation not permitted`；不挂载宿主 cgroup 的隔离 Debian systemd 容器退出255，特权宿主 cgroup 方案被安全审查拒绝，未执行。未修改宿主服务；真实 systemd 生命周期继续记录为环境阻塞，见[systemd 环境报告](../../acceptance/reports/systemd-environment-2026-09-20.md)。

2026-09-20 浏览器回归收口：修复三个真实失败后，`web/npm run test:e2e -- --workers=1` 完整 36/36（2.5分钟）通过；实例批量、跨实例传输和扩展跨设备场景均复验，见[默认应用浏览器回归](../../acceptance/reports/browser-regression-2026-09-20.md)。

2026-09-20 当前源码完整竞态回归：提权环境执行 `make test`（`go test -race ./...`）退出0；containers 8.485s、extensions 152.757s、master 277.018s、protocol 1.230s、runlog 2.279s、runtime 3.097s及其余Go包全部通过。受限沙箱首轮仅因IPv6回环监听和helper子进程限制失败，未计作代码失败。见[全仓竞态报告](../../acceptance/reports/full-race-2026-09-20.md)。

2026-09-20 续接复验：`make check` 退出0；`make windows` 重新生成 Linux/Windows 交叉构建产物并退出0；`web/npm run check && npm test -- --run` 的类型检查和 47/47 单测退出0。未启动常驻服务或留下测试进程。

2026-09-20 定时备份一致性补齐：Master 调度校验现接受 `save`、`pause`、`stop`、`hooks`，校验固定钩子引用格式，保持 `files` 模式禁止钩子引用；真实调度触发的保存程序已验证前后钩子、资源身份、最新正文归档和任务回执，定向集成测试及 race 均退出0（7.625s）。

默认备份应用的定时任务表单已同步暴露一致性策略及前后钩子引用，编辑时回填已有参数；类型检查、47/47 单测和备份界面浏览器 1/1（5.0s，实际断言 save 引用进入请求）通过，增量进入 rc7。

扩展用户数据迁移复验：真实 WASI guest 双用户迁移、失败保留旧包、回滚迁移、事务中断恢复和迁移期间并发写入保护通过，`go test -race ./internal/extensions -run 'TestUserDataIsolationCASMigrationAndRecovery|TestMigrationDoesNotBlockRegistryAndRejectsConcurrentWrite' -count=1` 退出0（85.404s），见[扩展数据迁移报告](../../acceptance/reports/extensions-data-2026-09-20.md)。

最终会话检查点：实际分钟调度完整复跑 **44800退出0，1/1（1.6分钟）通过**，每槽唯一任务、新来源版本和两份归档恢复正文均核对成功。首次78155仅因新增测试误期望201为200失败，已保留记录。所属计划已删除，fixture **98737退出0**。当前无活动测试或服务会话待接管；小时35594和冷构建已完成，不重复运行。见[实际分钟调度](../../acceptance/reports/schedule-wall-clock-2026-09-19.md)。

补充 F02/A01：扩展 `workspaces.spec.ts` 真实终端检查点内容策略并加入双浏览器设备冲突场景，fixture **1902302221** 会话 **27729** 2/2（6.9s）通过。默认布局同步剥离 `terminals`，显式工作内容同步携带真实终端检查点；第二设备先提交 revision 2，第一设备旧 revision 同步被拒绝并显示冲突，本地未同步 Monaco 正文保留。fixture 已停止。见[多设备冲突](../../acceptance/reports/workspaces-2026-09-19.md)。

续接后的本地回归：`make check` 退出0；`web/npm run check && npm test -- --run` 类型检查及47/47单测通过；提权隔离 `go test ./internal/systeminfo ./internal/runtime ./internal/terminal ./internal/runlog` 退出0；工作区后端 `go test -race ./internal/master -run 'TestCloudWorkspace' -count=1` 退出0。受限沙箱首次运行的回环绑定失败未计作代码失败，已用隔离权限复跑并通过。当前无活动测试、fixture或构建会话。

E09 平台适配修复：发现并修正 `internal/runtime/native_windows.go` 多核 Job Object CPU 配额被错误覆盖的问题，新增 `internal/runtime/cpu_quota_test.go` 锁定一核/八核/小配额/超容量边界。`go test -race ./internal/runtime`、`make check`、`make windows` 与 Windows runtime 测试二进制交叉编译均退出0。Windows 真机仍未提供，不能将该证据称为运行验证；修复随后已进入 rc5，rc4 保留为修复前的历史包。

最新交付已升级为 **rc5**：`BLORA_VERSION=development-20260919-rc5 make package` 退出0；六包 SHA256SUMS 逐项校验通过（文件哈希 `6a683cfbd4433036e2f088cf4e64673f6d5139e0231d7ea9697fce6cdabe73d5`），独立 `package-smoke.py` 会话 **60112退出0**，SDK/参考包签名、Master 初始化/TLS/登录及双 Daemon ONLINE 均通过。见[rc5报告](../../acceptance/reports/release-rc5-2026-09-20.md)。当前生产修复已进入 rc5；Windows 真机与 systemd 等外部环境缺口仍未关闭。

全范围仍不能宣布完成：Windows真机、独立systemd服务/计时器生命周期、独立远程环境与物理故障等矩阵未验证项仍保留。下一条具体动作是取得用户提供的专用测试环境访问配置后，运行对应平台验收入口并补证；已异步询问，尚无回复，不触碰当前生产宿主。以下为按时间倒序保留的历史过程记录，旧“当前”句不代表活动会话。

最新终态与当前任务：第二小时 **35594退出0**，p95 **46.5ms**、60528次交互、220次轮换与末轮校验通过，双PTY全程持续；945.7ms最大延迟保留，100/207818长帧。优化后完整JSON/61次资源样本/心跳已写`performance-hour-optimized*-2026-09-19.json`，10351采样自然退出0，31985夹具退出0。rc3对应小时证据通过，不覆盖其他平台。

后续E01监控真实失联首轮 **22756退出1（54.8s）**：指标stale/时间/诊断正确，但缓存进程页仍显示“实时列表”。已修正`MonitorApp.vue`：旧进程列表明确提示、显示最近成功读取时间、旧列表禁用终止入口和翻页。新增测试要求非空旧列表禁用按钮且恢复后转实时。build通过。当前复验 **17778**，fixture **17481**（`.local/fixture-3280126874/browser-credentials.json`），继续原复验；此前finally已恢复Daemon并停止实例。此新增生产修复不在rc3，复验/监控浏览器回归通过后需统一rc4并独立冒烟，不为仅报告结果循环重打包。小时性能路径未再更改，无需对监控文案单独再跑一小时。

第二轮小时当前心跳2064秒、128次轮换正常，主测试 **35594** /采样 **10351** /fixture **31985** 不变，预计15:17 UTC结束。33次已读样本内归档未超预算，最新浏览器堆44.6MB；不是最终通过。继续poll同一35594并将JSON追加functions.store相应2180714798键。生产源码/静态构建未再变化；新增monitor-offline真实测试仍未执行，必须等此小时终态后在新夹具运行。若最终p95仍超50ms，保留原数据并继续定位，不放宽阈值。

小时等待期间仅补测试源码：新增`web/tests/real/monitor-offline.spec.ts`，尚未运行。使用私有fixture精确配置路径+PID出生身份定位所属Daemon，实际SIGSTOP/CONT验证节点与实例缓存时间/诊断/原runId、另一节点仍采样、恢复后清除旧数据标记，finally恢复并停止所属实例。另断言失联时缓存进程列表不能继续显示“实时列表”；源码MonitorApp目前无条件显示该标签，可能需真实复现后修复，不能把新测试算通过。须在35594小时终态并停止旧夹具后，开新fixture执行此测试，不能暂停当前性能节点。当前生产代码/静态构建/rc3均未因这段测试而改变。

当前最新会话（2026-09-19 14:17 UTC）：优化后小时复验 **35594**，fixture **31985**（`.local/fixture-2180714798/browser-credentials.json`），RSS/归档只读采样 **10351**。原普通权限采样61436看不到/proc进程，已停止130，改为提权只读采样；主测试未重启。命令`BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，无CPU profiler、默认0.01秒输出、原50ms断言。预计15:17 UTC终态，必须继续原35594，不构建web/dist、不并行重负载。functions.store采样键`blora-perf-progress-2180714798`和`blora-perf-resources-2180714798`；终态保存第二轮独立JSON，再停10351/31985。

当前发行已经完成：`development-20260919-rc3`包含终端合批/关闭码/提示、Linuxunit边界、原生快照复制和依赖通知。六包外部hash/内部每文件manifest/许可文本逐字节全部通过；SHA256SUMS哈希`263567cb00344b3c6686fc1e08517bae1811341d916ea35e4ed657869cbfab38`。独立解包SDK/签名/双节点以及停机恢复分别使用rc3自身和兼容rc2均退出0，5671/18622均已终止并清理所属进程，详见`release-rc3-2026-09-19.md`。不用为了补写该包报告循环重新打包。

快照优化回归：check、47项单测、build通过；真实恢复配额/schema/指针事务中止3/3（7.8s），桌面编辑器终端组合8/9后末项额外导航导致故障注入失效（疑似同时另一测试报告触发开发服务器监听，保留失败），隔离终端2/2（17.6s）通过。优化前60秒CPU JSON复制7.291s，优化后2.059s、p95 45.4ms、堆36.3～79.4MB，短测不覆盖小时失败。旧23641/98648夹具均已停止，53406/28190诊断已结束。第一小时失败50.3ms的完整JSON及60次样本已保存，不能覆盖或改称通过。当前仅35594/10351/31985活动。

小时86432已退出1：全程3600秒/180次轮换、双PTY持续输出与末轮目标校验完成，唯一失败p95 **50.3ms > 50ms**。不得降低阈值或计通过。完整JSON、60次资源样本及已保留心跳已写`docs/acceptance/reports/performance-hour*-2026-09-19.json`；旧fixture96517退出0、采样86738退出130。新增终端双存储失败不ACK断言已实际2/2（18.0s）通过，57397退出0。当前定位性能：新fixture **23641**（`.local/fixture-2836489518`），CPU采样测试 **53406**（60秒，`BLORA_PERF_PROFILE_SOAK=1`），继续原会话取CPU摘要，结束后停止夹具。新开关仅测试诊断，非生产改动。发行rc3/通知验证/停机恢复与兼容rc2回退仍待执行；暂先定位50ms未达标原因。

当前有效小时测试仍为 **86432**，夹具 **96517**（`.local/fixture-2713429467`），只读采样 **86738**。最新2435秒/126次传输轮换正常，双PTY各约6.5MB/分钟，41次资源样本内归档未超16MiB。尚未到终态，不能计通过；继续原会话，不重建web/dist或另起重负载。终态后保存完整JSON及functions.store中的progress/resources数组，再停采样与夹具。随后运行最新terminal.spec.ts双存储失败不ACK断言，统一构建rc3并校验六包与通知，运行`package-smoke.py --state-restore --rollback-release dist/releases/development-20260919-rc2`。恢复脚本语法和参数拒绝检查通过，但实际恢复尚未运行；运维手册已补入口。异步询问的Windows/systemd专用环境尚无回复。

E09后续入口已写、尚未执行：`scripts/package-smoke.py --state-restore [--rollback-release <兼容旧包目录>]`，只管理自身私有目录，停止全部所属进程后复制Master/双Daemon及配置；重开后制造用户/文件变更，再停止、保留变更后状态、恢复旧快照，检查TLS/节点身份、用户登录、只读授权/控制403和文件正文。禁止与crash参数混用。三份Python脚本语法检查通过，**未冒充真实恢复演练通过**。小时86432在1156秒仍正常（63次轮换），86738采样/96517夹具保持。先完成小时与新增存储失败浏览器断言，再统一新包并运行该恢复入口（可用rc2作为兼容回退目标）。

E09独立进展：新增离线`third-party-notices.py`并接入Makefile/package.py，真实收集104条依赖通知（17Go模块+Go安装包+86npm包）、约490KB，缺少依赖/许可证即失败；README和`dependency-notices-2026-09-19.md`记录来源与边界。未运行重打包，不影响小时性能测试。当前86432在911秒仍正常、53次轮换；96517夹具和86738采样保持。小时结束后还需运行新增双存储失败不ACK浏览器断言，并统一打包核验通知与本轮生产修复；rc2均未包含这些增量。

最新427秒心跳：28次传输轮换，双PTY各约6.9MB/分钟，暂无错误；归档首段约20900，16MiB预算内。小时会话86432/采样86738/fixture96517不变。另已向`web/tests/browser/terminal.spec.ts`追加双存储写入失败时不ACK新批次的断言，**尚未运行这段最新断言**；小时结束并停止夹具后，运行`npm run test:e2e -- tests/browser/terminal.spec.ts --workers=1`，再统一验证/发行本轮生产增量。此前2/2（18.3s）仅覆盖追加前的版本，不冒充新断言已通过。

小时复验246秒心跳正常：17次已验证轮换，两路每分钟约7MB；四分钟资源采样归档各16.2MB且首段推进，Master约34.5MB RSS，两Daemon约27MB。主测试86432、采样86738、fixture96517继续运行。`functions.store`内累计记录键为 `blora-perf-progress-2713429467` 和 `blora-perf-resources-2713429467`；后续poll输出可按JSON追加，最终汇总实际采样峰值，不能把过程记录当通过。已异步询问用户是否有专用Windows/systemd测试环境，尚无答复，本机工作继续。

小时复验附属采样会话 **86738**：`python3 scripts/performance-resources.py .local/fixture-2713429467 --samples 62 --interval 60`；仅只读RSS/归档，每分钟输出JSON。主测试 **86432**、夹具 **96517**。终态后先保留结果，再停止86738与96517。README已补可复现长时入口；不为仅文档变化重新构建。

当前唯一小时复验：Playwright **86432**，夹具 **96517**，目录 `.local/fixture-2713429467`，命令 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`（默认0.01秒输出间隔）。包含终端合批、合法关闭码、断流状态提示与逐轮错误检查。上一高输出51451仍在3.1分钟OUTPUT_GAP失败，吞吐约三倍但不能计通过；fixture73328已退出0。最终前端build和终端浏览器2/2（18.3s）通过。继续原86432至终态，不重建静态资源、不重复启动、不因观察超时重启。成功后保存完整JSON/资源采样，再停止96517；失败则按诊断修复。源码增量尚未发行。

终端消费修复当前检查点：高输出10205在1.7分钟明确失败`OUTPUT_GAP`，旧fixture12256已退出0。现TerminalApp合并相邻DATA（64KiB载荷），TerminalModel合并相邻输出解析/保护并保留resize顺序与原序号，累计ACK仅在保护后发送；原缓冲/归档预算不变。关闭码4002修复同在源码。前端build通过，突发输出/resize/刷新序号90/精确ACK浏览器2/2（17.0s）通过。当前真实高输出复验 **51451**：180秒、输出间隔0.001；fixture **73328**，`.local/fixture-3865775071/browser-credentials.json`。继续原51451，终态后记录并停止73328；不要重建静态资源。报告 `terminal-batching-2026-09-19.md`。成功后仍需默认速率长时复验；本轮前端生产增量和之前unit边界均未进入rc2。

最新诊断（2026-09-19）：600秒复现20194退出1（10.4分钟），33次传输轮换、p95 42.8ms，但两次浏览器异常明确指出 `WebSocket.close(1002)` 为脚本不允许的关闭码。已将TerminalApp/InstanceConsole错误关闭改为4002，终端浏览器回归新增服务端OUTPUT_GAP后检查错误显示/只读/无pageerror，2/2（14.8s）通过；生产前端build通过。该修复尚未定位原始断流原因。旧fixture11505已退出0。当前新fixture **12256**，`.local/fixture-2178575800`；诊断测试 **10205**，`BLORA_PERF_SOAK_SECONDS=180 BLORA_PERF_OUTPUT_DELAY_SECONDS=0.001`（提高输出速率加速积压，不能代替默认速率小时验收）。新增逐轮检查终端错误和有限JSON附件，防止末分钟漏检；下一步接管10205查看原始错误，再修复。不要重建静态资源或启动重复负载。当前源码两个WebSocket关闭码修复也未纳入rc2。

600秒故障复现已启动：Playwright **20194**，fixture **11505**（`.local/fixture-30796204`）。65秒心跳正常，两路各约3.89MB/分钟，3次传输轮换；等待同一20194的终态与`BLORA_PERF_TERMINAL_FAILURE`诊断，不重复启动。原小时6442确已失败并停止对应夹具。新测试不修改生产代码、不放宽持续输出断言。

最新失败与接管（2026-09-19）：小时测试6442在9.3分钟退出1，双PTY连续一分钟接收增量为0，持续输出断言失败；最后成功心跳492秒、27次传输轮换。节点session仍running、归档首段继续前进，不能归因于负载生成器已结束，也不能计通过。已停止旧夹具78200（退出0）及采样20793（退出130）；旧句柄不要接管。新增失败时仅采集终端状态/错误和pageerror的诊断，不保存凭据或完整终端正文。新夹具 **11505**，目录 `.local/fixture-30796204`；正在启动600秒定向复现，凭据文件路径该目录下`browser-credentials.json`。下一步读取复现终态诊断、定位浏览器/流/恢复保护故障并修复；不是忽略失败直接重跑小时。只读采样脚本已实际运行并验证归档轮转。全范围仍未完成。

当前资源观测（2026-09-19 12:35 UTC）：新增只读入口 `python3 scripts/performance-resources.py .local/fixture-4166028878 --samples 60 --interval 60`，采样会话 **20793**，与原性能会话 **6442**、夹具 **78200** 并行。前两次磁盘采样显示两路归档各15.8–16.3MB、首段序号从约14544前进至20415，均小于16MiB；Master RSS约33–35MB、两Daemon约26–28MB。性能最新372秒、19次已验证轮换、双PTY持续新增输出。以上均为过程证据，不计小时通过。下一步继续原6442和20793，不重建前端、不启动重复负载；6442终态后保存结果，再停止20793及78200。单元边界生产修复尚未纳入rc2。

修正后的小时级首个心跳：会话6442在67秒仍运行，已完成2次传输轮换；两路PTY该观测区间分别新增6,224,580/6,150,304B，持续输出断言通过，堆53,849,655B、未确认峰值87,404B。尚未结束，继续原句柄；该进度不计小时级通过。

当前有效小时级运行：新fixture-4166028878已就绪，夹具会话 **78200**；修正后Playwright会话 **6442**，命令 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，私有凭据路径 `.local/fixture-4166028878/browser-credentials.json`。继续轮询6442，每分钟进度必须有两路terminalIntervalBytes；不要启动第三份测试、不重建静态资源。终态后保留最终JSON/失败信息并停止78200。旧91740/63313已明确终止，不再接管。当前尚无小时级结果。

纠正小时级负载：发现PTY生成器固定12000次输出会提前结束。已主动中止91740（退出130）并停止63313（退出0），不是因观察超时重启；旧fixture-950062738的约249秒不能计小时混合负载通过。测试现按soak时长设置有限输出次数，并每分钟断言两路PTY都有新字节。正在启动全新私有fixture重跑；旧句柄不再轮询。只读旧夹具内存样本约Master33.6MB、Daemon26.6/28.3MB，非最终小时证据。

小时级验收最新已观测进度：会话91740在66秒输出3次已完成传输轮换、44次活跃检查、当前堆53,724,785B、未确认峰值87,400B；进程仍在运行，非终态。保留原会话继续等待，不算一小时通过，也不重启负载。

小时级测试接管句柄：Playwright会话 **91740**，夹具会话 **63313**。继续轮询91740；测试终态后先保留JSON结果与失败诊断，再停止63313。不要重复启动测试或在其运行期间重建/替换前端静态资源。尚不能计小时级通过。

当前活动E08验收：持续传输入口已改为每轮成功并核验目标后，以目标当前版本覆盖同一测试文件，避免单次传输提前完成后继续空载采样；soak上限3600秒，每分钟打印无凭据进度。30秒真实Chromium通过（59.4s整测），两次传输轮换、p95 42.8ms、队列未确认峰值87,400B、堆38.3–92.0MB。旧fixture-816325812已退出0。新fixture-950062738已就绪，夹具会话63313，私有凭据 `.local/fixture-950062738/browser-credentials.json`；已启动 `BLORA_PERF_SOAK_SECONDS=3600 npm run test:e2e:real -- tests/real/performance.spec.ts`，一小时结果尚未产生，必须继续轮询原测试句柄，不能将观察超时当失败或另起重复负载。生产代码无新增变化，上一轮单元边界修复仍未纳入rc2。

最新生产修复（2026-09-19）：Linux服务动作此前可传入`.target/.mount/.socket`，timer动作可带文件路径；现将执行适配器限制为明确`.service`和无路径`.timer`，Windows名称规则不变。修改 `internal/systeminfo/service_actions.go`、`tasks.go`，新增Linux适配器与真实TLS双Daemon边界测试，运维说明同步完整unit名。systeminfo全模块race1.147s、原API权限1.759s、新Daemon拒绝边界2.688s通过，所有测试会话已退出。测试空PATH不调用宿主systemctl；不能把该证据代替真实systemd生命周期。本轮为实质进展；修复尚未纳入rc2，后续统一发行时须包含，勿称rc2已有此边界。下一步继续E05独立systemd验收环境/E08长期负载，或汇总生产增量后统一发行验证。全范围目标仍未完成。

最新续接结果：私有firewalld验收新增显式确认与实际Daemon重开组合，通过race32.103s；确认后两套规则保留，普通用户403，同键回放200、异键409且终态修订不变，最终恢复原快照。会话99317已退出0，测试所属进程/挂载/网络已清理。改动仅 `firewall_namespace_linux_test.go` 与验收记录；生产代码及rc2不变，无需重复打包。上一轮及本轮均为实质进展，目标保持未完成。下一步继续E05独立systemd服务/计划任务环境或E08持续混合负载；Windows、物理掉电、远程及小时级负载缺口仍保留。

最新检查点：定时备份接受后真实Master/SQLite关闭重开、双Daemon重连、任务身份保持和单一归档race3.198s通过（`TestScheduledBackupMasterReopenKeepsAcceptedOccurrence`）。会话89276已退出0，无活动验收会话。矩阵与README已加入真实ENOSPC、钩子、离线调度和私有firewalld入口，`go vet ./internal/master`通过。未确认可用独立systemd环境，不访问宿主服务/计划任务；Windows、物理掉电、远程和小时级负载仍未验证。后续从E05独立系统管理环境或E08持续混合负载入口继续，保留rc2及本轮源码测试，不重复初始化、不为文档记录循环打包。

当前总检查点（2026-09-19）：全范围仍未完成。生产修复（控制接收队列饱和背压、此前SDK请求键与查询恢复）已纳入rc2且独立解包验证通过；后续新增真实ENOSPC、save钩子、节点离线调度、私有firewalld断链回滚测试均通过并留在源码。矩阵A12更新为已实现并明确Linux证据；E03/E04/E05继续按子项记录。Windows真机、系统服务/计划任务实际主机组合、物理掉电、远程与长期负载仍有缺口。无活动会话待接管，不为仅同步本条记录重打包。下一步针对E05系统服务/计划任务检查可用独立systemd环境；如无法提供真实管理器，保持未验证并继续E08/E09可用场景，不调用当前生产宿主服务。

最新E05验收完成：私有user/mount/net namespace + 独立D-Bus/firewalld + 真实TLS Master/双Daemon，防火墙runtime/permanent差异、未托管范围保留、内核规则核对、实际阻断Master端口后期限回滚及解除后FAILED诊断均通过；race25.998s。普通用户403。首两次环境启动失败已修正并记录，见[真实防火墙报告](../../acceptance/reports/firewall-private-2026-09-19.md)。会话11284已退出0，所属服务/挂载/规则已清理，无活动测试会话。当前rc2生产代码未再变化；新增ENOSPC、钩子、离线调度、防火墙测试在源码。E05系统服务/计划任务和Windows、物理故障、长期远程组合仍待核对，下一步完善这些增量在矩阵与运行文档中的状态。

当前活动验收：E05私有防火墙环境首次运行，会话81768；`BLORA_TEST_FIREWALL_NAMESPACE=1 go test ./internal/master -run '^TestPrivateFirewalldApplyAndRestore$' -count=1 -v`。新增 `firewall_namespace_linux_test.go` 在显式启用后自启动unshare user/mount/net子进程，核对父子namespace不同、私有/run隐藏宿主总线，再启动任务所属dbus-broker/firewalld。测试真实端口应用、内核nft规则、runtime/permanent区别和回滚；结果尚未确定，不能计通过。工具仅探测私有namespace能力成功，没有调用宿主firewall-cmd。

合并复验完成：启用 `BLORA_TEST_ENOSPC=1` 的备份/调度/传输空间不足race组合退出0，20.324s，覆盖新增钩子成功/失败、真实目标空间不足、真实节点离线补跑/撤权及原有接口。没有新增生产代码改动；rc2仍是当前发行产物。下一步评估E05能否用私有net/user/mount namespace + 独立D-Bus/firewalld实现安全的真实防火墙演练；现有宿主防火墙绝不调用，尚未启动该环境。

E03钩子增量：固定argv独立保存程序接入真实Master/双Daemon，before原子发布新正文、after确认；归档恢复得到保存后的正文，失败分支明确FAILED并补偿，同键不重复钩子。race6.174s通过，新增 `backup_hooks_integration_test.go`。正在运行启用真实ENOSPC的备份/调度合并race回归，命令 `BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'Test(Backup|Schedules|TransferTargetENOSPC)' -count=1`；完毕记录结果。此前发布rc2生产代码未再变化，新增测试单独留在源码。

E04增量已完成：新增真实Daemon离线/重连调度组合，受控时钟五次到期不积累任务，恢复只补跑一次且实际备份最新源版本；离线撤权则无任务并报告authorization_denied。加强账本零任务断言后race2.869s通过，证据已写入备份调度报告。当前没有运行会话；下一步继续E03固定命令save前后钩子的真实备份/恢复全链路，已有runner和库测试不替代该组合。

最新验收（2026-09-19）：A12/E03真实ENOSPC已补齐Linux子项。私有user/mount namespace内的1MiB tmpfs：跨节点移动失败保留源并清理暂存，备份仓库满不发布归档，恢复目标满保留当前正文/其他文件/原备份。传输+备份创建race组合2/2、6.505s；恢复race3.678s通过。新入口 `BLORA_TEST_ENOSPC=1 go test -race ./internal/master -run 'ENOSPC' -count=1 -v`，代码为两个 `*_enospc_linux_test.go`；未启用时明确skip，不算通过。报告见[真实空间不足](../../acceptance/reports/enospc-2026-09-19.md)。当前无活动测试服务或挂载，生产代码未再变化，rc2不为仅增加这些测试重复打包。下一步继续E03/E04应用钩子及调度故障组合的证据核对；物理掉电/Windows/远程长期缺口保留。

rc2交付完成：`dist/releases/development-20260919-rc2` 构建、独立SDK构建签名、Master初始化/TLS/登录、双Daemon ONLINE和六个归档摘要全部通过；摘要文件SHA-256 `9cf78d121d1c28a58ba700b2e0ee5210c7a7670996ab397a7068115c8795dc62`。冒烟 `/tmp/blora-release-smoke-r1q90jz1` 所属进程已停止。继续A12真实ENOSPC：私有user/mount命名空间探测通过，新增 `transfers_enospc_linux_test.go`，仅显式 `BLORA_TEST_ENOSPC=1`运行，自启动隔离子进程并核对挂载namespace分离；1MiB tmpfs装在测试夹具目录，检查移动失败源保留/目标不发布/原数据保留/清理。该测试正在首次运行，尚无结果；新增测试未纳入rc2，不为仅补文档/测试再次打包。

当前交付检查点（2026-09-19）：协议背压修复后全仓Go普通回归退出0（Master129.494s、terminal16.944s），`make check`通过；协议race与真实101任务浏览器均通过。正在统一构建 `development-20260919-rc2`，包含扩展查询恢复、真实取消证据、节点测试入口修复和控制队列背压。下一步独立解包冒烟与归档摘要校验；不为装入包自身验证记录再重复打包。当前无活动fixture。

最新（2026-09-19）：控制通道有界背压修复后的协议race 1.192s通过；重编译Master/Daemon后，新fixture-1287262529的通知/任务分页2/2（8.2s）通过，加强为全部101任务SUCCEEDED后再次2/2（7.9s）。两个本轮fixture均已停止，退出0。完整 `go test -p 1 ./... -count=1` 会话45734正在运行；完成后记录结果并统一生成包含协议修复及扩展查询恢复增量的发行包。详见[控制背压报告](../../acceptance/reports/control-backpressure-2026-09-19.md)。剩余全范围缺口不变，不把局部修复等同于全部完成。

当前检查点（2026-09-19）：私有夹具支持的40项真实浏览器回归38通过/2失败（6.3m）。快捷入口测试错误地要求历史启动任务为零，已改为检查本次新增任务，定向1/1（3.6s）通过。任务分页测试101次mkdir中6项为 `INTERRUPTED/node_acceptance_missing`，不是环境缺失；发现控制连接接收队列饱和即断链，正在修复为控制通道有界背压并添加真实TLS慢消费者回归。`internal/protocol/conn.go` 和 `protocol_test.go` 已改，race会话52739；正在重建双二进制用于新fixture复验。旧fixture-1890745423已Ctrl+C退出0。下一步确认协议race、启动新fixture，重跑任务历史与节点/控制场景，再更新证据；此增量尚未发行。

最新故障修复：2026-09-19 节点维护测试误用未显示在桌面上的应用图标，已改为从实际应用菜单打开；添加动作超时与主错误诊断。私有 fixture-1890745423 定向 1/1（4.7s）及完整 nodes.spec.ts 2/2（53.1s）通过，包括真实 Daemon 暂停/恢复和维护/票据下载。详见[全套回归报告](../../acceptance/reports/full-browser-regression-2026-09-13.md)。当前继续同夹具支持的真实浏览器回归；夹具会话 83987，凭据路径 `.local/fixture-1890745423/browser-credentials.json`（禁止输出内容），完成后停止所属进程。

## 2026-09-19 续接：扩展任务请求身份

最新：真实取消丢回执/刷新组合已补验，Chromium 1/1（4.6s）通过；暂停已核对身份的私有 Daemon，真实 Master 接受取消后丢弃 HTTP 响应，刷新无重发，恢复节点后任务/界面确认 CANCELLED，取消键保持。fixture-1780818324 已恢复并停止。详细命令与范围见扩展任务重试报告；本项不证明运行中 WASI 中断，也不清除 Windows/远程长期缺口。下一步核对矩阵其余可执行场景，合并增量后统一发行验证。

参考扩展任务查询失败恢复入口已补齐：“刷新任务状态”只发起读取，查询期间去重，旧任务结果不覆盖新任务；定向 Chromium 1/1（8.5s）通过，断网后恢复查询不重发创建/取消。独立包构建通过，增量尚未进入 rc1。下一步继续真实取消丢回执组合，再统一纳入后续发行包；本轮测试已结束，无活动会话待接管。

发行增量已落地 `dist/releases/development-20260919-rc1`：Linux/Windows 构建、前端生产构建、SDK 与参考扩展三版本构建通过；独立提权冒烟通过 SDK 构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE；六个归档摘要全部 OK。冒烟目录 `/tmp/blora-release-smoke-_xl3psca` 所属进程已停止。本条为产物生成后的外部验证记录，不为把该条自身装入包再次打包。下一步继续矩阵尚缺的真实取消故障组合和平台验收；当前无测试会话待接管。

独立参考扩展浏览器取消恢复场景 1/1（8.7s）通过：路由中止首次取消，刷新不自动重发，显式重试复用原取消键，任务创建总数仍为 1。使用真实 Chromium 与模拟 HTTP 后端；真实节点接受取消后丢回执仍未单独演练。参考包重新生成。下一步按验收矩阵继续真实故障组合及统一发行验证，无运行测试会话待接管。

取消服务端补验完成：真实 TLS 双节点 `TestExtensionAdminAPIInstallsAndServesPackage` 7.018s 通过，同键取消保留原回执，CANCELLED 后重试修订不变，原扩展授权与撤权断言仍通过。浏览器取消回执丢失/刷新组合仍待验证；不能将服务端集成当作该 UI 场景完成。

取消链路也已接入显式稳定键：SDK `cancelTask(taskId, requestId)` 传递调用者键，宿主拒绝缺失/空白/超长键，参考扩展先持久化按任务保存的取消键再发送。SDK→宿主传输回归覆盖响应丢失后重建宿主并同键重试，定向 12/12 通过；前端类型检查、47/47 单测及 SDK/参考扩展构建通过。SDK 文档同步新签名；取消的真实浏览器故障场景尚未验证，下一步补此场景及最终发行包。

补验完成：SDK 到宿主的丢响应回归通过，前端 47/47；真实 HTTPS 双 Daemon Chromium 场景 1/1（9.2s）通过，后端接受后中止首次响应，刷新不重发，显式同键重试返回原任务 ID，WASI 结果正确。首轮仅末尾旧请求计数断言失败，修正为两次后通过。fixture-950580611 已停止（退出 0）。详细证据见[扩展任务重试](../../acceptance/reports/extensions-task-retry-2026-09-19.md)。下一步核对其他扩展任务调用入口和取消请求身份，完成该能力块后统一生成发行包。

核对当前源码发现 SDK `createTask(payload, requestId)` 实现丢弃第二参数，宿主随后自动生成新请求键。已改为 SDK 传递 `{payload, requestId}`，宿主校验请求键并放入 HTTP 头，任务正文保持原值；参考扩展在发送前持久化待提交内容和键，确认收到任务后清理。前端类型检查与 46/46 单测、SDK 和参考扩展 TypeScript 构建通过。定向传输测试覆盖正文与请求键分离、缺失键拒绝；尚需补充 SDK 到宿主的丢响应重试及真实扩展浏览器验证。此次源码尚未进入 final54 发行包；后续先完成这些验证，再统一打包，避免只为补写包自身报告重复生成归档。全范围验收仍未完成，原矩阵缺口继续有效。

记录日期：2026-09-14
当前阶段：B00～B09 的主要代码链路已接入；B10 正在进行全范围回归、故障边界和交付证据整理。
当前目标：连续完成全部明确范围；验收矩阵仍按缺失环境和未覆盖组合保留“进行中”，不把局部通过写成整项完成。

云工作区布局/引用 PUT 的请求键已写入窗口恢复状态，网络响应丢失后刷新并重试仍复用同一键，成功后清理；正文同步保持显式选择。`cloud workspace sync` 提权 Chromium 场景 1/1 通过（8.0s），路由中止首次写入后第二次同键只产生一次成功写入。节点登记、换钥、撤销动作也保存待处理键；上传分片按偏移派生稳定键，Master 对分片/完成入口在副作用前拒绝缺失键，上传重连/取消定向集成提权通过（2.180s）。`npm run check`、前端 46/46 单测通过。final53 已完成打包、提权独立冒烟和六个归档校验；冒烟目录 `/tmp/blora-release-smoke-zsjybjny` 已停止所属进程，`SHA256SUMS` SHA-256 为 `6a52012c8afb1335d97cd9792f73e7ff8793574547f38825b91084db824c7484`。Windows、远程节点/Engine、物理故障与长期组合仍按矩阵保留进行中。

上传浏览器恢复场景追加分片/完成键断言，提权 Chromium 1/1（5.6s）通过；三次分片请求按偏移复用同一上传键，完成请求使用固定后缀，响应中止后刷新仍只创建一个任务。final54 已完成打包、提权独立冒烟和六个归档校验；冒烟目录 `/tmp/blora-release-smoke-sgmac81e` 已停止所属进程，`SHA256SUMS` SHA-256 为 `265c1a8b4863d73212e8c94001bce95f4d83b4ec738d466a25bf5a47b33867fc`。其余平台和长期组合缺口不变。

final54 之后的提权串行全仓 Go 回归退出 0（Master 125.232s、extensions 16.389s、terminal 16.885s，其余包通过），节点/上传请求键改动未引入其他包失败。资源复核无 Blora/Vite/Playwright 进程和 `blora.test=1` Docker 对象；`/tmp` 可用约18GiB。最新请求键证据已写入工作区，发行包中的文档按 final54 记录。

调度更新/删除请求回执收口：`backup.Scheduler.SaveMutation` 与 `DeleteMutation` 将配置版本 CAS、记录删除和 `UserMetadataMutation` 回执放入同一 SQLite 事务；同键更新/删除重试稳定回放，不同载荷返回 `IDEMPOTENCY_CONFLICT`。前端定时任务删除将稳定请求键保存在窗口恢复状态中。定向调度 race 通过；随后提权全仓普通回归 Master 128.030s、竞态回归 Master 267.872s，扩展 149.546s，存储 13.581s，终端 18.646s，均退出 0；`make check`、前端构建及 final46/final47 发行包复验均完成。

final46 交付完成：调度请求回执修正与全仓回归证据同步后生成 `BLORA_VERSION=development-20260914-final46`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-94mpn3me` 已停止进程，`SHA256SUMS` SHA-256 为 `18cc8b4e1013c968924f4159eecfeb1925fff78b14e57906461b928989c3b537`。

final47 交付完成：纳入调度重启级回执测试与 OpenAPI 删除契约后的 `development-20260914-final47` 打包、独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-jzdy74x5` 已停止进程，`SHA256SUMS` SHA-256 为 `24e852ccf0fe2a6cba417ddb45217ca25cae98a75c087e2a1b7107a814af4ca2`。

默认系统/监控应用请求键现场收口：服务/计划任务动作、防火墙应用与确认按目标保存稳定键，进程终止确认框保存键，成功后清理，失败或丢响应时重试复用。提权 Chromium 系统管理浏览器场景 3/3（14.6s）通过；前端单测 46/46、类型检查和生产构建通过，修正已纳入 final48 发行包。

final48 交付完成：默认系统/监控应用请求键现场修正、调度回执与重启测试证据同步后生成 `BLORA_VERSION=development-20260914-final48`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-b1wqtbhw` 已停止进程，`SHA256SUMS` SHA-256 为 `580112f9bd3cc8738162a91fe800fb151bf8ac9c3b5f027e4c890ac63b48aee4`。

默认备份清理、用户创建、改密和扩展管理界面请求键现场收口：待处理键分别保存在备份窗口恢复状态、用户创建草稿和密码/扩展组件生命周期中，成功后清理，失败或丢响应时重试复用；密码和扩展包正文仍不写入工作区恢复。提权 Chromium 备份、账户/密码和扩展管理浏览器场景 4/4（9.4s）通过，前端 `npm run check`、46/46 单测和生产构建通过。下一步将该增量和证据纳入新的发行包并复验，Windows、远程节点/Engine、物理故障与长期组合仍按矩阵保留进行中。

final49 交付完成：上述界面请求键修正后的 `development-20260914-final49` 打包、提权独立冒烟和六个归档校验均退出 0；冒烟目录 `/tmp/blora-release-smoke-yb1y0em8` 已停止所属进程，`SHA256SUMS` SHA-256 为 `1a2ad9de3ce7951df30aa0aece4e51703a7cb20f9f7155d14dc99d1c3bc87aca`。受限冒烟首轮的 esbuild `EPERM` 已按环境限制单独记录；当前无活动 Blora、Vite/Playwright 进程或测试 Docker 对象。

定时任务删除成功后清理旧请求键的修正纳入 final50：前端检查、46/46 单测、构建、提权独立发行包冒烟和六个归档校验均通过；冒烟目录 `/tmp/blora-release-smoke-rptovpg1` 已停止所属进程，`SHA256SUMS` SHA-256 为 `8cd89bb3825e86bcac3f348ef7e443ab3e3d6bb767a43dc3ae60064b73c22f3d`。当前无活动 Blora、Vite/Playwright 进程或测试 Docker 对象，验收矩阵仍如实保留 Windows、远程、物理故障和长期组合缺口。

扩展管理请求键现已写入窗口恢复状态，刷新后仍可重试启停、回滚、卸载和目录安装；扩展管理 Chromium 场景 1/1、前端检查、46/46 单测和构建通过。final51 独立冒烟及六个归档校验通过，冒烟目录 `/tmp/blora-release-smoke-yratxpey` 已停止所属进程，`SHA256SUMS` SHA-256 为 `f87459b1eb291633fd6d31c156dfe7e72611dfc5312287ac39852cc50baaa1c1`。Windows、远程、物理故障和长期组合仍按矩阵保留进行中。

任务取消收据幂等收口：`model.Task` 持久化首次取消请求键，普通任务、上传和传输取消均在 SQLite 事务中记录；并发取消只推进一次修订，任务进入 `CANCELLED` 后重试仍返回原收据。任务中心为每个任务保存稳定取消键。存储 `Test(Cancellation|ConcurrentCancellation)` 与真实 TLS Master 集成 `TestTaskCancellationReplayAfterTerminalKeepsReceipt` 通过（0.139s），OpenAPI Task schema 已声明 `cancellationRequestId`。改动后提权全仓普通回归 124.958s、竞态复跑 255.904s 均退出 0（首次竞态仅 runlog `/proc` 时序波动，隔离复跑 2.345s 通过），`make check` 与前端 46/46 通过；修正已纳入 final36 发行包。

节点登记票据幂等收口：`EnrollmentMutation` 将管理员、请求键和名称写入受保护的元数据回执，同键重试复用原票据，不同名称返回 `ErrRequestMismatch`/`IDEMPOTENCY_CONFLICT`；原有无请求键的内部测试入口保持随机票据语义。存储 `TestEnrollmentMutationReplaysTicketByRequest`、真实 TLS 双 Daemon 管理回归和 `make check` 通过。登记修正后的提权全仓普通回归 126.994s、竞态回归 261.687s 均退出 0，前端 46/46 通过。

final38 交付完成：将登记幂等修正、全仓普通/竞态证据和矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final38`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-rrefkg26` 已停止进程，`SHA256SUMS` SHA-256 为 `ce76c1e52dfeb38627f2181c8751cb15048c538f0a025f05c568d7c2ad240628`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

节点换钥票据幂等收口：`NewRotationTicketMutation` 将管理员、节点、请求键和 `reactivate` 写入受保护元数据回执，同键重试复用原票据，不同载荷返回 `IDEMPOTENCY_CONFLICT` 且不覆盖待确认记录；真实 TLS 换钥集成覆盖回放、冲突、身份保持和旧连接栅栏。改动后提权全仓普通回归 128.965s、竞态回归 262.770s 均退出 0；下一步生成包含换钥修正和证据的发行包。

final39 交付完成：将换钥票据幂等修正、全仓普通/竞态证据和矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final39`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-wsor4qlc` 已停止进程，`SHA256SUMS` SHA-256 为 `e63dd2e0ae2c9fbd25e7fba2abf245fb8027984f23d5c07f13939eedc8647728`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final40 交付完成：将 final39 的换钥证据、进度与矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final40`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-ghsj5ufc` 已停止进程，`SHA256SUMS` SHA-256 为 `1739590e081a8fc043f914e82f5adda7c998e759df1e9fb3fb50a01c787d9fce`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

请求键边界再收口：实例创建、实例控制、终端创建和授权写入均在读取请求体前拒绝缺失/非法键；定向回归 6.241s、最新 Master 完整普通回归 127.825s、竞态回归 265.970s 均退出 0。下一步生成包含该边界证据的发行包。

final41 交付完成：将请求键边界证据、进度与矩阵同步进包后生成 `BLORA_VERSION=development-20260914-final41`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-rwgm48yh` 已停止进程，`SHA256SUMS` SHA-256 为 `37996802de1cc4510380974f8d60cbab69ca9513e9371c952be80dd631c7c286`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

Compose 项目保存 prepare/chunk 入口补齐请求键，前端按准备键和分片偏移发送稳定键；定向容器管理回归 0.377s（Compose 真实镜像用例因缺少显式镜像而按规则跳过），前端 46/46 通过。最新 Master 普通回归 129.598s、竞态回归 264.964s 均退出 0；下一步生成包含 Compose 修正和证据的发行包。

final42 交付完成：将 Compose 请求键修正、前端分片键和最新 Master 回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final42`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-k7gfk8j5` 已停止进程，`SHA256SUMS` SHA-256 为 `b6a8cc499522d141cbe05e11fdd7b1878b9ddc5e17d6f465ff4be9ca7eddf73e`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final43 交付完成：将 Compose OpenAPI 请求键声明同步进包后生成 `BLORA_VERSION=development-20260914-final43`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-zg38nbet` 已停止进程，`SHA256SUMS` SHA-256 为 `78520382738ab277e285a3cd084fd22b9ea8214d2f1aa802a9e828652fcee24d`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

换钥票据同键载荷冲突错误码明确为 `IDEMPOTENCY_CONFLICT`（409），真实 TLS 集成回归 0.209s 通过；下一步生成包含该错误码修正的发行包。

final44 交付完成：将换钥冲突错误码修正和回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final44`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-nn8w4hzn` 已停止进程，`SHA256SUMS` SHA-256 为 `15f56cc3151cc38b24ab0d2a916bff6342a9be27055f008bb259231a68deed70`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

管理写入口的请求键冲突错误码统一为 `IDEMPOTENCY_CONFLICT`；定向用户/角色/节点回归 1.198s，通过。最新 Master 完整普通回归 127.672s、竞态回归 263.528s 均退出 0；首次并行回归因 `/tmp` 空间耗尽，清理本轮缓存后重跑通过。下一步生成包含该统一错误码修正的发行包。

final45 交付完成：将管理写入口统一幂等冲突错误码及最新回归证据同步进包后生成 `BLORA_VERSION=development-20260914-final45`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 与六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-8pjxh2wo` 已停止进程，`SHA256SUMS` SHA-256 为 `e332a19b9c80ea6fdecdfc1638eaca3583dd547ac3e8aa787203439dec120195`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

2026-09-14 持久化请求键扩展：用户创建/改密、节点登记/身份轮换、云端工作区 PUT/DELETE 及通用任务/上传取消在授权后统一检查 `Idempotency-Key`；用户与工作区写入进入 SQLite 事务幂等回执，工作区 CAS、用户创建重放、改密和传输取消边界已加入集成测试。`go test -race` 定向 Master/Storage、传输取消回归、`make check` 和 OpenAPI 3.1 解析均通过。详见[请求键报告](../../acceptance/reports/request-keys-2026-09-14.md)。

防火墙二阶段确认幂等收口：确认请求键随 RPC 传到 Daemon，并写入租约记录和成功任务结果；同键在租约仍存或任务已完成时复用确认结果，其他键返回冲突且不重复修改主机规则。Daemon 定向 race 29.354s、Master 定向 race 2.096s 通过；改动后串行全仓普通测试（Master 126.258s）与 race（Master 256.038s、extensions 136.535s、terminal 18.627s）均退出 0，`make check`、前端46/46亦通过。随后补齐重启后无内存租约的双键 CAS 并发确认：只允许一个键成功，获胜键可回放，另一键返回 `FIREWALL_LEASE_CONFLICT`；定向 race 1.137s，最新串行全仓普通测试 Master 126.926s、race Master 256.432s 均退出 0。下一步生成包含该修正和证据的 final35 发行包并复验。

final34 交付完成：防火墙确认请求键修正后的 `BLORA_VERSION=development-20260914-final34 make package`、独立 `package-smoke.py` 和六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-ukbhphef` 已停止进程，`SHA256SUMS` SHA-256 为 `83dcf17fc84253c0e8731056497a7b0f83b4e2eb5f1768ede6d033260dfccd15`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final35 交付完成：重启后无内存租约的确认 CAS 并发修正纳入 `BLORA_VERSION=development-20260914-final35 make package`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-okvsdiwf` 已停止进程，`SHA256SUMS` SHA-256 为 `de8c519f6af02db48e41129150d41cef7805bd4bdeaa14b08db99ec3bcb1dd4d`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final36 交付完成：任务取消收据幂等修正（首次 `cancellationRequestId` 持久化、并发取消单次提交、终态重试复用）纳入 `BLORA_VERSION=development-20260914-final36 make package`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-7o0b93de` 已停止进程，`SHA256SUMS` SHA-256 为 `50fb039b87dc00bef14d1006f0e6ebcf929f407a1790efe51426d0d77ff80913`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

final37 交付完成：将 final36 进度与矩阵记录同步进包后生成 `BLORA_VERSION=development-20260914-final37`；独立冒烟、SDK/参考扩展构建签名、Master 初始化/TLS/登录、双 Daemon ONLINE 和六个归档校验均退出 0。冒烟临时目录 `/tmp/blora-release-smoke-jna5wn75` 已停止进程，`SHA256SUMS` SHA-256 为 `4fbae56157ee6ed406d6f8deb91e584be61fe44b130f05d16669a42e691a59a3`。当前无活动测试服务或带 `blora.test=1` 标签的 Docker 对象；Windows、远程节点/Engine、物理故障和长期调度组合仍按矩阵保留未验证。

当前源码串行全仓复验已明确退出 0：`go test -p 1 ./... -count=1`，Master 126.349s，runlog、runtime、storage、terminal、extensions 及其余包全部通过；该结果与此前并行 runlog 时序失败记录并存，仍不覆盖 Windows、远程和物理故障环境。

随后以独立缓存执行 `go test -race -p 1 ./... -count=1`，明确退出 0；Master、extensions、terminal、runlog、runtime、storage 及其余 Go 包全部通过。该竞态证据覆盖当前源码，不覆盖 Windows 真机、远程高延迟和物理故障场景。

随后收口文件入口：文件保存、文件动作和新建上传在实例/写权限确认后立即拒绝空白或超过 128 字节的请求键，避免读取正文、生成上传身份或联系节点；新增空白保存键与 129 字节动作键回归 2.693s、竞态回归 5.406s 均通过，完整 Master 126.836s、`make check` 均退出 0。上传分片/完成仍按已确认检查点与任务状态推进，前端 `api()` 会为其自动附带请求键。
随后发现并修复两个上传取消历史用例未携带请求键的问题；定向上传取消回归 1.562s、完整 Master 套件 125.538s 均退出 0。全仓其余包此前已通过，最终交付前继续保持串行全仓证据与发行包校验。

同日全仓与交付复验：提权环境 `go test ./...` 首轮仅因 Master 备份用例出现一次 `terminal access denied` 而失败；该备份用例单独复验通过，随后完整 `go test ./internal/master -count=1` 通过（125.544s），其余 Go 包也通过。受限环境的本机监听/辅助进程失败不计产品失败。前端 `npm test -- --run` 10 文件 46/46、`make check` 均通过。`BLORA_VERSION=development-20260914-final22 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS` 全部通过；`SHA256SUMS` SHA-256 为 `d9f80088796e1655df5fd9c3e4a46b92904a5675b2b1a895aa51d021fa60ca5c`，临时冒烟目录 `/tmp/blora-release-smoke-6svj66pc` 已停止进程但按约定保留私有诊断目录。

同日 final23 文档同步交付：将请求键 race 最终耗时同步后执行 `BLORA_VERSION=development-20260914-final23 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-he8gr2sa` 已停止进程，`SHA256SUMS` SHA-256 为 `7bab5ab5c69dda7cc0fef737ee7885af9b7caaebec3c9fc883ee5164d2eed2cd`。

同日 final24 记录同步交付：修正全仓测试一次环境波动说明后执行 `BLORA_VERSION=development-20260914-final24 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-uf0cgu2x` 已停止进程，`SHA256SUMS` SHA-256 为 `ed434701601b269098053b261a69ce0e785aea7d6021f971367839424996cefc`。

同日 final25 取消请求键交付：通用任务与上传取消新增授权后请求键检查，传输取消回归通过；执行 `BLORA_VERSION=development-20260914-final25 make package`、独立 `package-smoke.py` 和六个归档 `sha256sum -c SHA256SUMS`，均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-luefmwwc` 已停止进程，`SHA256SUMS` SHA-256 为 `466f6d3e6a2da7c3ab37beb389ca45e3c4dae5f87785cbf0788b738a8311a5fa`。

同日 final26 收口：修正两个上传取消历史测试未携带请求键的问题；上传取消定向回归 1.562s、完整 Master 套件 125.538s 退出 0。`BLORA_VERSION=development-20260914-final26 make package`、独立 `package-smoke.py` 与六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-ola9vpox` 已停止进程，`SHA256SUMS` SHA-256 为 `272c3f3bef10d4a380394ba23d2f99692b2328af4f63a12378ae79958f8a6e55`。

同日 final27 文档归档：将 final26 记录纳入包内后生成 `development-20260914-final27`，打包、独立冒烟与六个归档校验退出 0；冒烟临时目录 `/tmp/blora-release-smoke-bqwy03_2` 已停止进程，`SHA256SUMS` SHA-256 为 `a89541bb7657d7174d454828a7cea77477c9aaa73649b37ff85d55fe10f5faa`。

同日改密幂等收口：将自助当前密码校验移入首次事务回调，撤销会话后同键重试可稳定复用回执；定向改密 1.044s、完整 Master 127.375s 退出 0。下一步生成包含本修正与证据的 final28 并复验。

同日 final28 交付：改密幂等修正后的 `development-20260914-final28` 打包、独立冒烟与六个归档校验均退出 0；冒烟临时目录 `/tmp/blora-release-smoke-_y6cvl_e` 已停止进程，`SHA256SUMS` SHA-256 为 `efc9e4dda49e234b47d2037990331f726af7b674094b62667aa941c4c0c4d030`。

收尾复验：存储包单测 0.579s、`make check`（go vet）通过；调试残留、Blora 测试进程及测试标签容器均无活动对象。

同日 final29 交付：将最终收尾证据纳入归档后生成 `development-20260914-final29`；打包、独立冒烟和六个归档校验退出 0，冒烟临时目录 `/tmp/blora-release-smoke-k43e8lci` 已停止进程，`SHA256SUMS` SHA-256 为 `12171c6ed4aa430da966bb63e1b6ca3b20767985df2871667f17b36fba63af2a`。

同日 final30 交付：文件保存、文件动作、新建上传入口请求键边界收口后生成 `development-20260914-final30`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-dl1fg2bz` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `f7630947960072aded20d64f7a64397c8737bf125599f28fd8f661d143a5802d`。

随后将文件场景竞态回归证据纳入归档，生成 `development-20260914-final31`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-m3ccimab` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `18ff79189be48c49c93bb8c0a5330827472726f8f242b54eb7a44c6c28db3e80`。

随后将串行全仓 Go 回归证据纳入归档，生成 `development-20260914-final32`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-4osq9cd9` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `06e336ee2ec13240f2d7b86f6c83b638d5ec42db8cf2d9b0de35289aaae25c96`。

随后将串行全仓 Go 竞态回归证据纳入归档，生成 `development-20260914-final33`；独立冒烟完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE，临时目录 `/tmp/blora-release-smoke-9tnfnent` 已停止进程；六个归档校验退出 0，`SHA256SUMS` SHA-256 为 `84931b88163ef6182ca4504b9d26105b427c4c16fe569f783c2807223abbe34e`。

当前收尾检查：工作区无 Blora/fixture/Vite/Playwright 活动进程，Docker `blora.test=1` 标签对象为空；final33 归档校验已复核。下一步仅推进验收矩阵中仍可获得环境的缺口（Windows 真机、远程高延迟/Engine、物理 ENOSPC/掉电、小时级调度与完整跨平台组合），未因本地串行测试和 Linux 冒烟通过而宣称 F01～F14/A01～A17/E01～E09 全部完成。

最新普通 `go test ./... -count=1`：Master 127.270s 及其余包通过，runlog 输入时序用例首轮空输入失败；隔离重跑 0.052s 通过。该非确定性结果已保留，不能宣称本轮全仓无条件通过；生产代码与发行包不受此测试时序波动影响。

2026-09-14 备份/调度竞态复验：提权隔离环境执行 `go test -race ./internal/backup ./internal/master -run 'Test(Schedule|Backup)' -count=1` 退出0（backup 1.856s、Master 5.816s），覆盖真实 TLS API 归档/条件恢复/撤权/取消、计划触发与 SQLite 回执恢复；`go test -race ./internal/daemon -run 'TestBackupHook' -count=1` 退出0（2.039s），固定 argv Hook 的环境绑定、稳定回执及未知副作用不重放通过。具体应用 hook 语义、长期离线调度、物理 ENOSPC/掉电仍按 E03/E04 保留。

同日真实备份浏览器验收：`backups.spec.ts` 在 HTTPS 双节点 fixture-4134305623 中 1/1 通过（3.5s）；真实文件写入/版本核对、备份归档任务、默认备份窗口恢复计划与覆盖任务、中文正文核对，以及 Asia/Shanghai 定时计划保存/修订删除均通过。fixture 已停止并清理。证据见[备份浏览器报告](../../acceptance/reports/backup-browser-2026-09-14.md)。物理 ENOSPC/掉电、具体应用 hook 语义和长期离线调度仍未验证。

同日 final20 交付复验：因真实备份浏览器报告更新执行 `BLORA_VERSION=development-20260914-final20 make package`，`python3 scripts/package-smoke.py dist/releases/development-20260914-final20` 与 `cd dist/releases/development-20260914-final20 && sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-p00oohac`，`SHA256SUMS` SHA-256 为 `4e25104368473ca06daaae78870b9ee8b015e5e0ebbda665ad26be116a45c60b`。

同日 final21 交付收口：将最新备份/调度报告同步进包内后执行 `BLORA_VERSION=development-20260914-final21 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final21` 和六个归档 `sha256sum -c SHA256SUMS`，均退出0；冒烟临时目录 `/tmp/blora-release-smoke-klzy0dtj`，`SHA256SUMS` SHA-256 为 `9cb7f53ce64e6703003054bdcce2af95a7b5c392b59b0de9debde1e9b2e5282d`。

2026-09-14 请求键边界扩展：扩展包安装/升级/回滚/启停/卸载、目录安装和防火墙租约确认统一在授权后检查 `Idempotency-Key`；新增定向 race 77.219s 和 OpenAPI 3.1 解析断言通过，见[请求键报告](../../acceptance/reports/request-keys-2026-09-14.md)。前端非 GET 请求继续由 `api()` 自动生成合法键。

同日 final19 交付复验：因上述证据更新执行 `BLORA_VERSION=development-20260914-final19 make package`，`python3 scripts/package-smoke.py dist/releases/development-20260914-final19` 和 `cd dist/releases/development-20260914-final19 && sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-urc6k8iz`，`SHA256SUMS` SHA-256 为 `7e365582aad6a03cc84e1d7b7342ce795da2f903036f0db059f37875385bc9b2`。包内含最新 docs/acceptance/reports/backup-scheduler-race-2026-09-14.md。

2026-09-14 512MiB 传输批量步进复验：重编译夹具纳入 `transferChunksPerStep=16` 后启动私有 fixture-1569035732，真实 Chromium 混合场景 1/1 通过（约2.0分钟）。八窗口、双PTY、10,000项目录与512MiB跨节点传输全程并行，目标校验成功；536,870,912B 用时110.868s（约4.62MiB/s），反馈p95 39.0ms、0/285长帧、JS heap53,729,723B、API RTT p95 43.7ms，未确认字节峰值86,306B。fixture已停止且无测试残留。首次旧夹具结果仅保留为索引基线；长期内存/远程网络/归档轮转、Windows及物理故障仍按矩阵保留。证据见[大文件传输报告](../../acceptance/reports/transfer-proof-2026-09-13.md)和[真实性能报告](../../acceptance/reports/performance-real-2026-09-13.md)。

同日长时性能采样：性能测试新增可选 `BLORA_PERF_SOAK_SECONDS`（每250ms堆采样并持续窗口交互），重编译夹具 fixture-2822501719 设置60秒后真实场景1/1通过（约2.0分钟）；采样65.697s、p95 43.3ms、1/3770长帧、堆34,569,872–97,894,433B、未确认峰值86,867B，512MiB传输109.245s。该模式默认关闭，测试超时预算已包含soak时长；小时级稳定性仍未验证。详情见[真实性能报告](../../acceptance/reports/performance-real-2026-09-13.md)。

同日传输改动后的全仓竞态复验：`go test -race -p 1 ./...` 退出0，Master 248.686s、extensions 136.133s、terminal 18.695s，其余 Go 包通过；使用独立 `/tmp/blora-go-race-final14` 缓存，未出现空间或监听权限错误。

同日 final14 交付复验：`BLORA_VERSION=development-20260914-final14 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final14` 均退出0。独立解包完成 SDK/参考扩展构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE；`SHA256SUMS` SHA-256 为 `63a3318876d538a4e91ea52ea7318485e424931f23b5162285e4ede38da81099`，临时冒烟目录 `/tmp/blora-release-smoke-y8dorqs1`。

由于发行脚本包含 README/docs，证据更新后生成 final15：`BLORA_VERSION=development-20260914-final15 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final15` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-7wbln31e`，`SHA256SUMS` SHA-256 为 `746ed95e6d047933374c35b1d3b119902d96a9a8ac90904acbe1abb3b9f1c183`。

加入长时采样开关及60秒证据后生成 final16：`BLORA_VERSION=development-20260914-final16 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final16` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-be_8yxox`，`SHA256SUMS` SHA-256 为 `d2acfcb7b76896028bf701bc1e96778e2de9c29fc0909b4212b707af384d96c7`。

长时采样测试超时预算修正后生成 final17：`BLORA_VERSION=development-20260914-final17 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final17` 和归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-y3ghbju2`，`SHA256SUMS` SHA-256 为 `de8908dc5ea53aca45f154d5413aab5950eddeb04c9bbf7f394928f1dc4df081`。

同日 Windows 适配编译复验：在 `GOOS=windows GOARCH=amd64 CGO_ENABLED=0` 下，`go test -c` 分别生成 runtime、runlog、systeminfo 测试入口（`/tmp/blora-runtime-windows.test.exe`、`/tmp/blora-runlog-windows.test.exe`、`/tmp/blora-systeminfo-windows.test.exe`），全部退出0；无 Wine/Windows 真机，仍不计运行验收。

同日 final18 交付复验：`BLORA_VERSION=development-20260914-final18 make package`、`python3 scripts/package-smoke.py dist/releases/development-20260914-final18` 与六个归档 `sha256sum -c SHA256SUMS` 均退出0；冒烟临时目录 `/tmp/blora-release-smoke-rb_wc5my`，`SHA256SUMS` SHA-256 为 `adb76c7bac48b4c47728c43bd25b91fbb1d81eec930de17eac0aa583a0a30928`。

同日交付检查：`make check`（`go vet ./...`）与 `cd web && npm test -- --run`（10个文件、46/46）退出0；final14 构建已同时完成前端类型检查和生产构建。

同日传输定向 race：`go test -race ./internal/master -run '^TestTransfer' -count=1` 退出0（64.894s），覆盖当前批量步进实现的跨节点传输集成回归。

2026-09-14 E02 真实 Docker 复验：Docker Engine 29.4.1/API 1.45/Compose 5.1.3 上，`TestRealDockerComposeLifecycle` 使用随机标签测试对象完整执行 25.063s 退出 0，覆盖容器/镜像/卷/网络、双流日志、真实拉取及失败、Compose 健康失败/更新/删除和卷保留；清理后无测试对象。远程 Engine、真实 ENOSPC/掉电、Windows 容器与长时网络故障仍未验证。

同日 E02 浏览器补验：任务自有双节点 HTTPS fixture 连接本机 Docker Engine，`tests/real/docker.spec.ts` 1/1 通过（44.5s）；覆盖容器中心草稿刷新、Compose 保存与显式部署、中文输出/阶段分页、任务详情刷新、删除部署和卷保留。fixture 与随机 Docker 对象均清理，远程 Engine、ENOSPC/掉电、Windows 容器和长时故障仍未验证。

同日 A09 内存采样补验：真实 Linux PTY 无消费者滚动场景延长至约17.5秒（32批×2MiB、0.5秒间隔），每5ms采样归档和 `/proc/<pid>/status` RSS；定向 race 退出0，3435次采样，归档峰值65,484B、磁盘54,729B，PTY RSS 741,376B→3,764,224B，未超过基线+32MiB，最终Gap与末尾标识通过。证据见[混合流报告](../../acceptance/reports/mixed-stream-control-2026-09-13.md)；跨平台、全连接长期及更长生产负载仍未验证。

同日 A13/E06 未授权浏览器边界：真实 HTTPS fixture 的 `extensions.spec.ts` 成员独立浏览器上下文 1/1 通过（3.5s）。无 `app.use`、仅有 `app.use` 无 `node.read` 时资源摘要/任务均 403；补齐 `node.read` 后资源摘要 200，撤销 `app.use` 后立即回到403；finally 撤销授权、卸载扩展并清理 fixture。证据见[扩展未授权浏览器报告](../../acceptance/reports/extensions-unauthorized-browser-2026-09-14.md)。Windows、远程高延迟和长期组合仍未验证。

同日扩展真实套件复验：新增场景并入 `extensions.spec.ts` 后完整真实 HTTPS 套件 6/6 通过（22.6s），fixture Ctrl+C 停止且无残留进程/测试 Docker 对象。

2026-09-14 请求键边界收口：`requireRequestID` 现在统一拒绝空白及超过128字节的 `Idempotency-Key`；实例控制在动作解析、资源授权和节点门禁后、任务接纳前检查。后端定向集成回归、前端类型检查与 46/46 单测通过。`BLORA_VERSION=development-20260914-final10 make package` 和 `python3 scripts/package-smoke.py dist/releases/development-20260914-final10` 均退出 0，完成独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录及双 Daemon ONLINE。

同日最终复验：清理任务缓存后在提权环境以 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-race-final13 GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 退出 0，Master 251.165s、extensions 136.422s、terminal 18.729s，其余 Go 包通过。此前空间耗尽的复跑失败（`no space left on device`）保留为环境记录，未计作代码失败。

幂等键边界回归新增 `TestRequireRequestIDBounds`（空白/缺失/129 字节拒绝，128 字节接受）并通过；随后以 `BLORA_VERSION=development-20260914-final11 make package` 重新生成发行包，独立冒烟 `python3 scripts/package-smoke.py dist/releases/development-20260914-final11` 退出 0。

恢复实例控制的资源授权先后顺序后，以 `BLORA_VERSION=development-20260914-final12 make package` 重新生成发行包；`python3 scripts/package-smoke.py dist/releases/development-20260914-final12` 退出 0，独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 均通过。

A09 长时采样测试改动后，以 `BLORA_VERSION=development-20260914-final13 make package` 重新生成发行包；`python3 scripts/package-smoke.py dist/releases/development-20260914-final13` 退出 0（临时目录 `/tmp/blora-release-smoke-60hxav7x`），独立 SDK/参考扩展构建签名、Master 初始化/TLS/登录和双 Daemon ONLINE 均通过。

最终检查：`go vet ./...` 退出0；`cd web && npm run check && npm test -- --run` 退出0（10个文件、46/46）；项目测试/fixture及带 `blora.test=1` 标签的 Docker 对象均无活动残留。发行包 `dist/releases/development-20260914-final13/SHA256SUMS` SHA-256 为 `09865538d7e776eba64424d8f73766a0734c1235ae0a46015de2d2f764ed8cfc`。

## 2026-09-14 扩展跨设备现场与资源身份复验

- AppHost 恢复校准现在可以返回规范化 `resourceRef`：默认实例应用按稳定实例 ID 读取服务端节点身份，具备 `resource.read` 的扩展通过受控资源摘要接口校准实例引用；节点或实例被删除/撤权时保留原窗口和草稿，不把可恢复的离线现场标成恢复存储失败。涉及 [应用类型](../../../web/src/app-host/types.ts)、[扩展注册表](../../../web/src/app-host/registry.ts)、[桌面恢复](../../../web/src/desktop/store.ts)、[默认应用注册](../../../web/src/apps/register.ts)。
- 新增跨设备浏览器副本场景：两个不同 `deviceId` 打开同一云端布局/引用副本，扩展视图从状态版本 1 迁移到 2，旧节点提示校准到 `node-new`，恢复后保持两个扩展窗口，1/1 通过（4.7s）。证据：[扩展跨设备报告](../../acceptance/reports/extensions-cross-device-2026-09-14.md)。
- 本轮 `npm run check`、`npm test`（46/46）、`npm run build` 和定向 Playwright 场景均退出 0。Windows 真机、远程高延迟、物理 ENOSPC/掉电及完整未授权组合仍按矩阵保留进行中。

## 2026-09-14 最新回归结果

- 统一请求键边界扩展到已持久化的实例/节点设置、用户/角色变更、节点撤销、扩展数据与资源写入、备份计划/恢复和传输取消；相关 Master 集成测试全部通过。OpenAPI 的对应修改操作现在声明 `IdempotencyKey`。
- 提权串行全仓 race 在上述入口改动后再次通过：`go test -race -p 1 ./...` 退出 0，Master 250.611s，其余包通过；受限沙箱重跑仅因本地 TLS 监听权限失败，未计作代码失败。
- A09 慢日志/批量传输/停止并行测试在当前源码下重新通过：`go test -race ./internal/master -run TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop -count=1 -v` 退出 0（13.07s）；真实 TLS Master/双 Daemon、8.4MB 跨节点传输和未确认日志接收端同时运行，未消费峰值 82,635B、接收队列 79,507B，停止和归档/目标校验均成功。全进程长期内存采样仍未覆盖，A09 不提升为整项完成。
- 扩展浏览器套件 6/6 通过（21.1s，含新增跨设备副本场景）；调度 API 删除缺少请求键时返回 400，带键删除及 race 定向测试通过。一次完整 `make test` 在并行资源压力下有 4 个既有时序失败（Daemon 防火墙命令、Master 审计登录、runtime 后代条件、PTY 输出）；随后逐项低并发重跑均通过，其中 Master TLS 场景在提权环境通过。该次全仓失败原样保留，不改写为全仓通过。
- 随后以 `GOMODCACHE=/tmp/blora-go-mod GOPATH=/tmp/blora-go GOCACHE=/tmp/blora-go-cache GOFLAGS=-buildvcs=false go test -race -p 1 ./...` 串行完成全仓 race，退出 0：Master 259.481s，其余包通过。并行命令的失败事实仍保留，串行结果作为最新可复核全仓证据。

## 2026-09-13 扩展全链路真实复验

- E06/E07/A13 独立扩展成功链路完成新 fixture 复验：Linux Chromium HTTPS、真实 Master + 双 Daemon，`extensions.spec.ts` 5/5、24.2s，覆盖通知去重、元数据草稿/真实实例写入、用户数据迁移与失败回滚、WASI 任务与资源窗口/快捷入口/移窗/关闭、本地注册源双版本安装升级及禁用门禁。证据：[扩展全链路报告](../../acceptance/reports/extensions-real-2026-09-13.md)。fixture `fixture-2953159556` 已停止。
- 仍按矩阵保留 Windows 真机、远程高延迟注册源、物理掉电/ENOSPC 及完整未授权组合，不能把本次成功链路当作 E06/E07 整项完成。

## 2026-09-13 全套浏览器回归

- 最终 Linux 构建在基础双节点 fixture `fixture-2323261112` 执行 43 项真实 Chromium 场景，37 项通过。6 项按环境条件保留失败：容器指标/Docker 需要可用 Docker Engine，性能与上传关闭需要 `--performance` fixture，节点维护长套件一次连接超时造成污染；实例设置单独新 fixture 重跑 1/1（29.8s）通过。详见[全套浏览器回归](../../acceptance/reports/full-browser-regression-2026-09-13.md)。fixture 已停止。

## 2026-09-13 性能与关闭网页复验

- 带 `--performance` 的最终构建 fixture `fixture-2936034595`：性能 1/1（27.8s），p95 36.2ms、最大44.4ms、0/298长帧、8窗口/双PTY/10,000目录/16MiB传输全程并行；关闭网页上传续传 1/1（1.4m），10,000项解压、16MiB uploadId 从917,504B续传并完成摘要核验。证据：[真实性能](../../acceptance/reports/performance-real-2026-09-13.md)、[关闭网页](../../acceptance/reports/upload-close-2026-09-13.md)。fixture 已停止；长期稳定性、512MiB吞吐、远程网络和Windows仍未验证。

## 2026-09-13 最终 race 与扩展边界修复

- `make test`（`go test -race ./...`）退出 0：Master 267.898s，extensions 157.362s，其余包通过。新增软删除与任务接纳竞态保护、扩展迁移资源身份不变式均已通过定向验证；最终 Master/Daemon/devfixture 构建、Windows 交叉构建、前端 check/test/build 与 `make sdk` 均通过。

控制台输入请求键边界加入后再次运行 `make test`（`go test -race ./...`）退出 0：Master 257.810s，其余包通过。

## 2026-09-13 最终发行包冒烟

- `BLORA_VERSION=development-20260913-final make package` 退出 0，生成 Linux/Windows Master、Daemon、SDK 和 web 六个归档；随后 `python3 scripts/package-smoke.py dist/releases/development-20260913-final` 在受限提升环境退出 0，完成解包后的 SDK/参考扩展离线安装构建、临时签名、Master 初始化/TLS/登录及两个 Daemon ONLINE 检查。测试进程已停止，证据已补入[发行包报告](../../acceptance/reports/package-2026-09-13.md)。Windows 真机、真实业务运行和物理故障仍按矩阵保留未验证。

随后在任务入口幂等键、慢消费者诊断和云端冲突提示改动后，`BLORA_VERSION=development-20260913-final2 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final2` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，详见[发行包报告](../../acceptance/reports/package-2026-09-13.md)。

监控 stale 诊断和前端提示更新后，`BLORA_VERSION=development-20260913-final3 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final3` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../../acceptance/reports/package-2026-09-13.md)。

系统管理按钮加入能力禁用状态后，`BLORA_VERSION=development-20260913-final4 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final4` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../../acceptance/reports/package-2026-09-13.md)。

资源身份校准、撤权现场保留和调度删除请求键契约更新后，`BLORA_VERSION=development-20260914-final6 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final6` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../../acceptance/reports/package-2026-09-13.md)。

调度创建/更新入口统一使用非空白 `Idempotency-Key` 校验；`BLORA_VERSION=development-20260914-final7 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final7` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过。

调度请求键校验提前到请求体解析前，避免无键请求触发解析或资源访问；定向 race 通过。`BLORA_VERSION=development-20260914-final8 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final8` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过。

统一持久化修改请求键后的 `BLORA_VERSION=development-20260914-final9 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260914-final9` 均退出 0，独立 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过；发行包包含实例/节点/用户/角色/扩展/备份/传输入口的最终校验。

控制台输入入口补齐请求键校验和 OpenAPI 声明后，`BLORA_VERSION=development-20260913-final5 make package` 与 `python3 scripts/package-smoke.py dist/releases/development-20260913-final5` 均退出 0；新包独立解包 SDK/参考扩展、Master/TLS/登录和双 Daemon ONLINE 复验通过，测试进程已停止，详见[发行包报告](../../acceptance/reports/package-2026-09-13.md)。

## 2026-09-13 实例删除用户入口

- 实例中心和实例专用窗口新增管理员“删除实例”入口，使用现有确认对话框、CSRF/幂等请求和刷新反馈；服务端仍在持久化接纳阶段检查 STOPPED、活动任务和墓碑身份。前端 `npm test` 46/46、`npm run check`、生产构建均通过；当前源码下原有 A16 目标改名/软删除/同名替换回归在 fixture-1914707629 1/1（7.3s）通过。普通成员不呈现特权入口，删除失败保留原视图与错误。
- 实例中心和实例专用窗口新增管理员“删除实例”入口，使用现有确认对话框、CSRF/幂等请求和刷新反馈；服务端仍在持久化接纳阶段检查 STOPPED、活动任务和墓碑身份。前端 `npm test` 46/46、`npm run check`、生产构建均通过；当前源码下 A16 目标改名/软删除/同名替换回归改为真实点击管理员删除入口，在 fixture-115753733 中 1/1（7.0s）通过。普通成员不呈现特权入口，删除失败保留原视图与错误。

## 2026-09-13 任务入口幂等键边界

- 为主机终端、容器操作、Compose 应用、系统服务/计划任务/防火墙应用、进程终止、实例终端创建/关闭、扩展任务和跨节点传输补充统一的 `REQUEST_ID_REQUIRED` 校验。此前这些入口虽会在存储层拒绝空键，但错误不稳定且可能先触发远端检查；现在在授权后、任务接纳前明确失败，前端 `api()` 自动生成的请求键保持兼容。
- `gofmt` 后以提权本地 TLS 监听运行 `go test ./internal/master -run 'Test(System|Container|Host|Transfer|Extension|Process|Terminal)' -count=1`，退出 0（39.554s）。沙箱内首次运行因禁止监听端口失败，未计作代码失败。

控制台输入入口随后补齐同一 `REQUEST_ID_REQUIRED` 早期校验，OpenAPI 为实例终端、主机终端、Docker/系统任务、Compose 提交、跨节点传输和控制台输入补充必填请求键声明；`TestConsoleInputHasOneDeliveryAndGracefulStopUsesIndependentStdin` 定向运行退出 0（1.620s），覆盖空键 400、合法请求单次交付和独立停止输入。

OpenAPI 3.1 文档重新解析通过（103 paths），并逐项核对上述任务入口的 `IdempotencyKey` 引用。

## 2026-09-13 系统能力不可用反馈

- SystemApp 根据节点 `system/capabilities` 结果停止对不可用服务/计划任务轮询；对应操作按钮随能力禁用，能力缺失时仍给出明确原因，防火墙预览/应用同样不会提交注定失败的请求。可用能力路径保持原确认、幂等和二阶段租约逻辑。
- `npm run check`、`npm test`（46/46）、`npm run build` 和系统管理浏览器场景 2/2（54.2s）通过；浏览器场景分别覆盖 `systemd/firewalld` 可用和 Windows 能力不可用，使用受控 API，未改变真实主机。

## 2026-09-13 慢消费者诊断交付竞态

- Master 日志流在发送 `SLOW_CONSUMER` 诊断后立即关闭 WSS，客户端偶尔只能观察到 EOF。`internal/master/logs.go` 现在在错误帧成功写入后等待最多 100ms 或连接结束，再释放流，保持有界关闭语义并确保诊断可见。
- `go test ./internal/master -run TestLogSlowConsumerDeadlineAndCursorReattach -count=2` 在提权本地监听环境连续通过（91.364s）；修复前同测试 3 次中 2 次出现 45s 后 EOF。
- 同一场景 `go test -race ./internal/master -run TestLogSlowConsumerDeadlineAndCursorReattach -count=1` 通过（47.369s）。

`make build` 与 `make windows` 在本次后端改动后均退出 0，Linux Master/Daemon 及 Windows amd64 交叉产物已重新生成。

## 2026-09-13 云端工作区冲突提示

- Desktop 在云端工作区同步收到 `WORKSPACE_CONFLICT` 时明确提示“本地现场保留”，并引导刷新列表后选择副本；不会覆盖本地草稿或把旧云端版本误报为成功。`web/src/desktop/Desktop.vue` 类型检查通过。

## 1. 已知事实

- 已通过 DevSpace 打开 /data/instances/blora-panel；放入本次文件之前，该目录为空。
- 用户已经执行完整启动提示；保留原规划，现已有 Go Master/Daemon、运行适配、协议/存储和 Vue 桌面源码。
- Go/前端依赖已锁；Master/Daemon 与前端生产构建通过。真实 HTTPS fixture 已覆盖 43 项浏览器场景中的大部分核心链路（基础 fixture 37/43；扩展、性能和关闭网页负载另有独立通过证据），整项 F/A/E 仍受 Windows、危险主机、ENOSPC/掉电、远程负载等条件约束。
- 两个签名 Daemon、两个权限账号的 TLS API 集成已运行；任务去重、实际进程重启、管理重连、权限撤销流有子项证据。
- 最近两套本机浏览器验收 fixture 均已 Ctrl+C 停止；没有公开部署或操作生产节点。运行测试会话以第 7 节最新检查点为准。
- DevSpace 工作区标识可能随会话变化；接手时使用当前工具返回的实际工作区，不硬编码旧标识。

入口：[行动指导](../../../ACTION_GUIDE.md) · [技术架构](../../plan/Blora-01-技术架构.md) · [功能交互](../../plan/Blora-02-功能与交互.md) · [验收矩阵](../../acceptance/ACCEPTANCE_MATRIX.md)

## 2. 范围锁定

M0～M3 为本次完整任务的内部实施顺序。F01～F14、A01～A17、E01～E09 均需跟踪，不能只完成 M1 后将其他功能列为下一阶段。

保留的核心要求：管理通信与业务网络分离；默认/扩展应用统一宿主；跨节点全部获准实例与单独创建授权；应用多窗口多标签/移窗；实例桌面入口与专用窗口；完整刷新恢复；有界停止重启、真实退出确认；前后端分离、多用户多节点。

条件项和排除项以验收矩阵第 1 节为准。不得用“首版”“原型”“时间有限”自动删减明确范围。

## 3. 实施块台账

实现：未开始 / 进行中 / 已实现 / 阻塞。验证：未运行 / 通过 / 失败 / 环境缺失。每块可能覆盖多个 F/A/E，最终以验收矩阵细项为准。

| 块 | 内容 | 实现 | 验证 | 证据/下一动作 |
| --- | --- | --- | --- | --- |
| B00 | 工程、依赖、环境与最小构建 | 已实现 | 基线构建通过 | [核心报告](../../acceptance/reports/core-2026-09-09.md)，生产运行文档在继续补全 |
| B01 | 共享协议、存储、迁移、持久化任务 | 已实现 | 子项通过 | 接受/派发持久化、去重、状态CAS、取消竞态与事件；故障组合仍在 B10 验证 |
| B02 | 桌面、AppHost、多窗口标签、恢复 | 已实现 | 子项通过 | Chromium Monaco中文/撤销重做/立即刷新/移窗、云端布局/内容选择和真实后端浏览器均有证据；完整跨设备/长时恢复仍在补 |
| B03 | 身份权限、节点通信与基础默认应用 | 进行中 | 用户/角色/维护/换钥子项通过 | [管理报告](../../acceptance/reports/administration.md)；Users/Nodes真实浏览器通过，完整审计查看/保留等继续 |
| B04 | Linux/Windows/隔离运行与生命周期 | 进行中 | Linux子项通过；部分环境缺失 | [runtime报告](../../acceptance/reports/runtime-2026-09-09.md)，Linux真实两节点重启；Docker/cgroup/Windows真运行未验 |
| B05 | 日志、PTY、终端及会话恢复 | 进行中 | 真实PTY/浏览器/Docker/原生日志与stdin子项通过 | 独立helper+16MiB归档、固定run输入任务；容器日志事件适配及完整分栏继续 |
| B06 | 文件、编辑、传输与草稿恢复 | 进行中 | 真实文件/上传/跨节点与浏览器子项通过 | [传输报告](../../acceptance/reports/transfers.md)，含同宿主根别名/硬链接/源出生检查；ENOSPC等继续 |
| B07 | 实例中心、默认应用与任务体验 | 进行中 | 实例配置、列表/卡片视图、批量操作、任务历史/通知和自动启动子项通过 | [实例视图报告](../../acceptance/reports/instances-view-2026-09-22.md)、[任务/通知证据](../../acceptance/reports/notifications-2026-09-12.md)、[实例设置报告](../../acceptance/reports/instance-settings.md)；整项仍受跨平台与故障组合验收条件约束 |
| B08 | Docker/Compose、监控、备份调度、系统管理 | 进行中 | Linux/Docker/Compose、监控历史、备份调度/恢复、受控系统管理子项通过 | [Docker日志归档报告](../../acceptance/reports/docker-log-history-2026-09-22.md)、[长时监控报告](../../acceptance/reports/monitor-history-soak-2026-09-21.md)、[调度阶段崩溃报告](../../acceptance/reports/scheduler-staged-process-crash-2026-09-21.md)；Windows 主机适配、独立 systemd、远程/物理危险故障仍待相应环境 |
| B09 | 扩展 SDK、沙箱、包与安装生命周期 | 进行中 | 独立包安装/升级/回滚/卸载、沙箱能力门禁、WASI 任务/资源桥接、用户数据迁移子项通过 | [扩展全链路报告](../../acceptance/reports/extensions-real-2026-09-13.md)、[扩展数据迁移报告](../../acceptance/reports/extensions-data-2026-09-20.md)；远程目录运维、Windows 和长期跨平台组合仍待补 |
| B10 | 故障、性能、全范围集成和交付 | 进行中 | 真实套件通过，条件项未清零 | 真实 HTTPS + Docker fixture 15/15；继续整理 OpenAPI、故障/性能、Windows 与受控系统变更证据 |

当前产品统计：按“整项所有条件均有证据”的严格口径，功能 0/14、原规划场景 0/17、补充场景 0/9；大量子场景已通过，具体证据和未覆盖条件见验收矩阵，不把文档或测试替身计入通过。

## 4. 接手后的第一个具体动作

当用户明确发起实现任务时：

1. 阅读根目录 AGENTS.md、ACTION_GUIDE.md、两份规划和验收矩阵全文。
2. 检查项目目录的现状，尊重用户在本记录之后新增的文件和更改。
3. 检查 Go、Node、包管理器、系统、Docker/Compose 与浏览器测试能力；分别记录可用、缺失、未检查。
4. 根据固定技术基线建立 B00 工程和验证入口，完成后继续 B01～B10。
5. 按每个块的真实结果更新此文件和验收矩阵，不等待用户逐块确认。

若用户当前只要求修改规划，则只执行该文档任务，不因本文件而启动开发。

## 5. 环境、命令与证据

以下均为待填写槽位，不是已经运行的命令或成功结果。

| 项目 | 实际值 |
| --- | --- |
| 工作目录 | /data/instances/blora-panel |
| 主机系统/架构 | Fedora Linux 内核 6.19.12-200.fc43.x86_64；cgroup v2 当前路径不可写 |
| Go / Node / 包管理器版本 | go1.25.9 linux/amd64；Node v24.15.0；npm 11.12.1（选定npm） |
| 浏览器及 E2E 驱动 | Chromium 151.0.7922.71，Playwright CLI 1.60.0；真实HTTPS文件与PTY测试2/2通过13.3s，更多桌面证据见desktop报告 |
| Docker / Compose | Docker29.4.1、API1.54/min1.40、Compose5.1.3；容器、镜像、卷、网络、Compose 生命周期、日志/PTY 与桌面入口均已接入并有真实/隔离测试；远程 Engine、真实 ENOSPC 和危险主机组合仍未验证 |
| Windows 构建工具及真机运行环境 | Go交叉编译入口将提供；未发现Windows真机，运行验收仍环境缺失 |
| 后端构建命令 | make build；根Makefile设缓存及GOFLAGS=-buildvcs=false |
| 前端类型检查与构建命令 | cd web && npm run build；npm run check |
| 本地联合启动/初始化方式 | make fixture（先构建web）；仅本机9443，私有随机凭据见fixture目录 |
| 静态/单元/集成/浏览器测试命令 | make test/check；cd web && npm test / npm run test:e2e，需本机监听和Chromium权限 |
| 平台测试和发布命令 | `make windows`（Windows amd64 交叉编译）；Windows Job/ConPTY 真机待环境 |
| 扩展构建/安装测试命令 | `make sdk`；`go test ./internal/extensions ./internal/master -run 'Extension' -count=1`；真实 fixture `tests/real/extensions.spec.ts` |
| 证据目录 | `docs/acceptance/reports/` 下核心、桌面、Docker、SDK、系统、工作区和 OpenAPI 报告；真实凭据仅保留在 `.local/fixture-*` |

## 6. 活跃工作检查点

每完成一块、遇到重要失败或接近上下文边界时更新：

- 当前块和验收编号：B10 全范围验收与交付证据，关联 F01～F14、A01～A17、E01～E09；B00～B09 主链路已接入。
- 当前具体任务：B10 全范围回归与证据收口；UI 样式按用户要求冻结。Docker 日志归档入口的本地实现、浏览器替身、真实 Master/Daemon API 与真实 Chromium/Docker UI 已闭环；原 Compose 长流程也已复验通过。本轮完成 B07～B09 本地入口审计、前端/Go/Windows 构建复核；下一步转为维护外部环境验收清单，不重复已通过的 UI 视觉套件。
- 本轮修改/证据文件：`web/src/apps/DockerLogs.vue`、`web/tests/browser/docker.spec.ts`、`web/tests/real/docker.spec.ts`、`internal/master/containers_integration_test.go`、`docs/acceptance/reports/docker-log-history-2026-09-22.md`；验收矩阵与本进度文件。真实 fixture 已 Ctrl+C 停止，遗留的本轮 Compose 容器/卷按标签清理，当前没有活动测试或 fixture 进程。
- 最新成功验证：真实 Docker 日志归档 UI 1/1（7.6s）；原真实 Compose 长流程 1/1（45.6s）；真实 Master/Daemon 归档 API race 2.79s；前端 `vue-tsc`、63/63 单测、构建、`make check` 仍通过。
- 最新失败验证：曾有一次把 Compose 容器选取/日志换行/容器启动条件写错的验收脚本失败，均已修正并在最终真实场景通过；不记录为产品失败。矩阵仍保留真实远程 Engine、Windows 容器和物理故障等未验证边界。
- 尚未解决的缺口：F01～F14均未整项通过；A09/A13仍进行中；E01～E09均未整项通过。跨平台缺Windows Job/ConPTY真机和容器；E05缺可操作独立systemd manager；还缺远程 Engine/真实高延迟、物理掉电/Engine主机存储耗尽、长期墙钟故障及非Chromium平台组合。E08 当前源码一小时本地目标已达到 p95 48.1ms，但公网/高 RTT/丢包、Windows 和 WebKit 仍未验证。详见验收矩阵，不以构建或替身覆盖环境缺口。
- 下一条具体动作：当前环境中未发现还应补写生产代码的本地缺口；若提供 Windows 真机、独立 systemd manager 或远程受控节点，按矩阵执行对应验收。Windows Job/ConPTY、独立 systemd、远程高 RTT/丢包、物理掉电和长期跨平台场景继续维持环境缺失记录，不用 Linux 本地结果替代。
- 运行中服务/测试会话及接管方式：当前无活动 fixture；真实回归使用 `.local/fixture-*` 私有凭据并在结束后 Ctrl+C 停止。宿主 PTY 测试需显式 `BLORA_HOST_EXEC_HELPER_BINARY` 与隔离镜像；防火墙真实改动不在本机执行。
- 2026-09-09 扩展升级安全性：升级前保存旧清单与包体，写入失败恢复旧包体；版本限定为 `v?数字.数字.数字` 并拒绝降级；管理员新增 `POST /api/v1/extensions/{id}/rollback`，卸载/重新安装会清理旧回滚状态，生命周期回滚测试及 Master TLS 集成测试通过。验证：`go test ./internal/extensions -count=1`、`go test ./internal/master -run TestExtensionAdminAPIInstallsAndServesPackage -count=1`（提权后通过）。
- 同日扩展契约校正：Master 注册源允许 SDK 使用的 `resource.write`/`task.create` 能力，并对入口、资源处理器、窗口/标签策略做边界校验；全仓编译与 vet 复验通过。
- 同日系统管理增量：计划任务支持跨平台 enable/disable 命令、任务名路径穿越防护、Master `POST /api/v1/nodes/{id}/system/tasks/actions` 和 Daemon 持久任务执行器；普通用户仍受 `host.manage` 门禁。边界及相关模块回归通过，真实主机变更和失联回滚未执行。
- 同日扩展禁用门禁：`Manager.AuthorizeCapability` 每次从持久清单重新读取启用状态和能力，禁用后新能力请求立即拒绝；生命周期测试覆盖该语义。真实扩展任务桥接仍待接入。
- 同日独立包交付：参考扩展新增 `npm run package`，生成 `reference.blora-extension.json`（清单、SHA-256、base64 包体）；Master 新增受限 `POST /api/v1/extensions/install-package` 原始包安装入口，TLS 集成测试验证独立包安装与完整性校验。
- 同日外部清单兼容：注册源持久化 icon/color/permissions、入口和窗口策略字段并拒绝 protected 外部清单，避免 AppHost 加载时因缺少权限字段崩溃；参考包已重新生成并通过安装集成测试。
- 同日参考扩展能力示例：示例源码加入 `task.create` 能力声明和 `enqueueExampleTask` 门禁调用，重新执行 `npm run package` 成功；该示例仍等待真实宿主任务桥接，未将 transport 替身计为 E06 完成。
- 同日容器日志生命周期：Daemon 仅在 `container.delete` 得到成功状态后删除对应节点归档，失败/未知结果保留历史；归档清理测试与相关集成回归通过。
- 同日浏览器回归：`npm run test:e2e -- --workers=1 --reporter=line` 完整 21/21 通过（约 1.5m）；此前中断的输出已由明确终态补证。该套件覆盖刷新、窗口/标签、编辑、文件、日志、终端、传输和权限交互，但不代替真实 Master/Daemon 混合负载性能测试。
- 同日全仓 Go 回归：`go test ./...` 全部通过（Master 48.213s，含现有集成场景）；未将测试替身或交叉编译当作生产环境验收。
- 同日 E08 探针：新增 Playwright 8 窗口本地负载测试；独立运行 p95 50.8ms，完整 22 项套件内 p95 75.4ms（heap 约68MB，4～5/60 长帧），显示负载波动并高于 50ms 目标。完整浏览器套件 22/22 通过；报告明确不覆盖真实终端/万条节点目录/后台传输与 RTT。E08 仍进行中。
- 同日 E08 探针增量：通过任务栏真实右键“新建窗口”建立第二个终端窗口，探针测得 p95 69.5ms、最大73.9ms、5/60 长帧、heap 约64MB；窗口未绑定后端会话，不能替代两条活跃终端验收。
- 同日防火墙受控应用：Linux firewalld 端口规则支持 256 条以内、60 秒确认期限和逐条失败反向回滚；Master/Daemon 任务权限与过期确认测试通过，全仓编译通过。未在当前宿主执行危险改动，租约级失联自动回滚仍待实现。
- 同日防火墙取消语义：应用 reload 成功后若任务上下文已取消，仍执行反向回滚并报告取消；系统信息模块、Master 集成与全仓编译复验通过。
- 同日防火墙回滚测试增强：引入仅测试可替换的命令执行器，隔离模拟 firewalld 在删除规则中途失败，验证已完成新增/删除操作按逆序恢复；生产默认仍使用 `exec.CommandContext`。
- 同日持久化边界：容器日志归档及扩展清单/包体 rename 后同步父目录（Windows 跳过不支持的目录 Sync），相关模块测试和 `make windows` 通过。
- 同日系统信息 race 回归：`go test -race ./internal/systeminfo -count=1` 通过，覆盖防火墙输入与隔离命令回滚测试。
- 同日防火墙诊断：回滚命令失败不再静默忽略，错误同时保留原始操作失败和 `firewall rollback failed` 诊断；隔离测试覆盖该分支。
- 同日平台回归：`make windows` 在新增 firewalld 分支后仍成功生成 Master/Daemon Windows 产物；Windows 实机行为仍未验证。
- 同日系统模块完整回归：`go test ./internal/systeminfo ./internal/master ./internal/daemon -count=1` 与 `go vet ./...` 通过；未执行当前宿主的危险防火墙变更。
- 同日系统管理前端：SystemApp 新增计划任务启用/停用和防火墙受控应用确认按钮，提交 60 秒确认期限；`npm run check && npm run build` 通过。真实主机变更仍未执行。
- 同日系统管理浏览器证据：新增 `web/tests/browser/system-management.spec.ts`，用完整 API 合同验证节点选择、计划任务启用和防火墙确认提交；Playwright 定向场景 1/1 通过。证据记录于 `docs/acceptance/reports/system-management-2026-09-09.md`；真实危险主机变更与租约失联回滚仍未验证。
- 同日全量浏览器复跑：新增场景后尝试 23 项串行套件时，测试开发服务器在首个用例后连接被拒（23 项均报 `ERR_CONNECTION_REFUSED`/trace 资源错误），因此不计为通过；定向系统管理场景仍稳定 1/1，通过既有 22/22 基线保持有效。后续需在稳定浏览器服务环境重跑全套。
- 同日全量浏览器复跑修复：清理本轮 trace 产物并以 `--trace=off` 串行执行，完整 23/23 通过（约 1.2 分钟）；性能场景 p95 16.9ms、最大 71.7ms、1/60 长帧、heap 约64.7MB。前次连接拒绝归因于本机 trace/开发服务器环境竞争，E08 仍受真实后端负载场景限制。
- 同日 Go 回归复核：首次 `go test ./...` 出现两个瞬态集成竞态（进程 `/proc` 已退出、备份节点通道尚未连接）；对应单独测试及完整 `go test ./internal/master -count=1` 随后通过（48.062s）。未把首次失败隐藏，后续应继续降低测试启动同步敏感性。
- 同日测试同步修复：`fileFixture.restartNode` 现在同时等待新 `StartupID` 和 bulk 通道协商完成，避免重启后首个备份/文件请求偶发 503；`TestBackupAPIRealArchiveConditionalRestoreAndRevocation` 定向通过。
- 同日扩展管理浏览器证据：新增 `web/tests/browser/extensions.spec.ts`，管理员生命周期操作禁用、回滚、卸载均走真实 API 边界并断言禁用请求体；定向 Playwright 1/1 通过。真实扩展代码加载与任务桥接仍待实现。
- 同日扩展任务桥接：新增 `POST /api/v1/extensions/{id}/tasks`，服务端复核启用状态与 `task.create`、节点 `host.manage`，payload 限制 64 KiB；Daemon 仅返回不执行的扩展数据结果并拒绝越界。扩展 API 集成测试已通过，SDK README 记录调用合同。
- 同日扩展拒权回归：集成测试增加“禁用扩展后新 task.create 返回 409”断言，验证禁用门禁覆盖任务入口；测试通过。
- 同日前端宿主 transport：新增 `web/src/app-host/extension-runtime.ts`，将 SDK `task.create` 限定映射到 Master 扩展任务 API，并复用 CSRF/幂等 API 客户端；未知能力和缺失 app/node 在浏览器侧拒绝。`npm run check && npm test`（20/20）通过。
- 同日扩展桥接静态审计：`go vet ./...` 通过；清理不可达的通用 `app.use` 映射，`extension.task` 仅由专用授权分支校验扩展能力和节点管理权限，集成测试复验通过。
- 同日扩展边界回归：集成测试新增普通用户无 `host.manage` 时 403、payload 超过 64 KiB 时 413 的断言；扩展任务桥接测试通过。
- 同日交付说明校准：README 更新为当前已接入监控、系统管理和扩展生命周期的实际状态，并保留动态加载、完整沙箱恢复及网络注册源等明确缺口。
- 同日监控浏览器证据：新增 `web/tests/browser/monitor.spec.ts`，验证节点选择、进程搜索、内存排序及带 `startTicks` 的终止提交；定向 Playwright 1/1 通过，报告已更新。
- 同日监控实例视图：MonitorApp 增加实例资源引用下的 `/instances/{id}/metrics` 查询，展示实例 CPU/内存和 stale 标记；`npm run check` 与监控浏览器场景均通过。
- 同日监控前端构建回归：`npm run build` 成功（Vite 1755 modules）且 `npm test` 5 个文件 22/22 通过；仅保留既有大 chunk 警告，不影响构建结果。
- 同日监控缓存修复：实例指标 queryKey 改为响应式，窗口内切换实例不再复用旧实例采样；`npm run check` 与监控浏览器场景 1/1 通过。
- 同日标准检查入口：`make check`（`go vet ./...`）通过。
- 同日浏览器全量回归：纳入扩展管理、监控和系统管理新增场景后，`npm run test:e2e -- --workers=1 --trace=off --reporter=line` 完整 25/25 通过（约 1.2 分钟）；性能探针 p95 17.1ms、最大54.3ms、1/60长帧、heap约64.1MB。
- 同日性能测量增量：探针增加 5 次 API 请求 RTT 采样；最新独立运行 API RTT p95/max 8.7ms，交互 p95 16.8ms、最大61.1ms、1/60长帧、heap约65.1MB。明确标注为本地 fixture 边界，不替代远程 RTT/带宽负载。
- 同日 Master race 全量回归：`go test -race ./internal/master -count=1` 通过（74.668s），覆盖扩展任务、节点重启、权限、文件/终端/备份/容器集成场景。
- 同日进程消失竞态加固：增加 `processGone`，兼容 `/proc` 的 ENOENT/ESRCH 和系统错误文本，避免日志停止时快速退出被误报；runtime 定向测试及三项 Master 生命周期场景重复 3 次均通过。
- 同日回归确认：此前全仓回归中失败的 `TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop` 在修复后重复 3 次通过（2.101s）；后续全仓测试仍需再跑一次以更新总状态。
- 同日全仓 Go 回归完成：`go test ./...` 通过（Master 46.624s，其余包通过/缓存），确认进程竞态和扩展最小权限改动未引入总回归。
- 同日 Linux 进程竞态修复：运行时在弱隔离进程扫描、cgroup 成员读取和 pidfd 信号前统一使用 `os.IsNotExist` 识别进程瞬时退出，避免 `/proc/<pid>/stat` 消失被误报为 restart 失败；runtime 全测试和双节点生命周期集成测试均通过。
- 同日竞态 race 复验：`go test -race ./internal/runtime -count=1` 与 `go test -race ./internal/master -run TestTwoNodesAuthorizationAndDurableLifecycle -count=1` 均通过，分别 2.603s/5.868s。
- 同日 Windows 产物复建：`make windows` 成功生成 `dist/blora-master.exe` 与 `dist/blora-daemon.exe`；Windows Job/ConPTY 真机运行仍属环境缺失，交叉编译不计作运行验收。
- 同日宿主 transport 单测：新增 `web/tests/extension-runtime.test.ts`，覆盖 `task.create` API 路径、CSRF、payload 保真及未知能力/缺失 scope 的本地拒绝；`npm test` 5 个文件 22/22 通过。
- 同日宿主合同扩展：`createExtensionTransport` 支持宿主显式注入 `window.open`、`state.capture`、`state.restore` hooks，未注入能力仍本地拒绝；`npm run check && npm test` 通过（5 文件 23/23）。
- 同日扩展最小权限修正：扩展任务不再要求 `host.manage`；新增扩展资源 `app.use` 授权类型并结合目标节点 `node.read`，集成测试验证获准普通用户可提交、未获准用户仍 403。
- 同日扩展包加载边界：新增认证 `GET /extensions/{id}/bundle`，仅已启用扩展可读取已校验包体，禁用返回 409；前端新增 `fetchExtensionBundle`，为 sandbox loader 保留 no-store API 边界。Master 集成与前端 `npm run check && npm test`（24/24）通过。
- 同日扩展 sandbox loader：新增 `ExternalSandboxApp.vue`/`loadExternalSandboxApp`，在 `sandbox="allow-scripts"` opaque-origin iframe 执行 bundle，以 `postMessage` 接入受限 transport；前端检查和 24/24 单测通过。资源标签/状态恢复端到端仍待验证。
- 同日 sandbox 构建修复：避免 SFC script 中直接出现 `</script>` 导致模板解析截断，改用运行时拼接闭合标签；`npm run build` 与 `npm test`（24/24）通过。
- 同日 SDK 安全文档：README 明确 bundle 只能经认证接口读取并放入 `sandbox="allow-scripts"` iframe，通过 `postMessage` transport，禁止注入宿主主页面。
- 同日编辑器恢复稳定性复验：IME、多光标、查找替换及即时刷新场景单独重复 5 次全部通过；此前完整套件中一次撤销时序失败未能复现。扩展 sandbox loader 改动后的完整套件已有关闭 trace 的 25/25 通过记录，后续若资源允许再做一次完整串行复跑。
- 同日最终 Go 回归：使用临时 GOCACHE 并允许 httptest/WSS 本地监听后，`go test ./...` 全部通过（Master 47.150s）；无权限的默认沙箱运行仅受监听策略阻断，非代码失败。
- 同日扩展恢复链路：工作区初始化扫描未知外部 appId 并经认证 manifest 动态注册 sandbox host；sandbox 组件按 viewTab resourceRef 动态解析 nodeId，支持跨节点标签复用。`npm run check`、`npm test`（24/24）与 `npm run build` 通过。
- 同日扩展浏览器回归：`npm run test:e2e -- --workers=1 --trace=off --grep 'extension management installs' --reporter=line` 1/1 通过。
- 同日扩展入口补齐：扩展管理窗口新增已启用扩展“打开”动作，复用动态 sandbox 注册并避免重复注册；前端检查与 24/24 单测通过。
- 同日扩展能力门禁：sandbox transport 根据 manifest 在调用宿主 hook/API 前检查 `window.open` 与 `task.create` 声明；新增拒绝测试，前端 `npm test` 25/25 通过。
- 同日防火墙回滚修复：ApplyFirewall 在父 context 取消后使用独立 10 秒回滚 context，避免断线取消让恢复命令全部提前失败；`go test ./internal/systeminfo -count=1` 通过。
- 同日 clean 交付构建：`make sdk` 完成 SDK 与 reference app 的 clean `npm ci`、类型检查和构建；`make web` 完成前端 clean `npm ci` 与生产构建，均退出 0。
- 同日真实 Docker/Compose E2E：Docker 29.4.1/API 1.45、Compose 5.1.3 下运行 `TestRealDockerComposeLifecycle`，25.651s 退出 0；覆盖容器/镜像/卷/网络、Compose 应用更新删除、日志、失败诊断与卷保留，报告见 `containers-e2e-2026-09-09.md`。
- 同日宿主 Docker PTY E2E：使用隔离镜像与 exec-helper 运行 `TestDockerExecProtocolOwnershipPTYAndClose`，退出 0；验证 PTY 协议所有权与关闭路径，结果已写入 runtime 报告。
- 同日实例监控失联回退：新增每实例 120 点持久指标缓存；节点断连时 `/instances/{id}/metrics` 返回最近点并标记 `stale`，避免旧值被误认为实时。监控定向集成测试通过（0.924s），矩阵和报告已更新。
- 同日 Docker 浏览器边界：新增节点选择、容器查询、显式启动确认、幂等任务提交和完成回读场景；`npm run test:e2e -- --workers=1 --trace=off --grep 'Docker app queries' --reporter=line` 1/1 通过（5.4s），报告见 `docker-browser-2026-09-09.md`。
- 同日监控 race 复验：`go test -race ./internal/master -run 'TestMetrics|TestMetricHistory|TestInstanceMetric' -count=1` 通过（3.239s）。
- 同日跨节点传输回归：`GOCACHE=/tmp/blora-go-transfer-cache go test ./internal/master -run 'TestTransfer' -count=1` 11 项通过（25.830s）；更新 F07/A12 证据，真实 ENOSPC 与浏览器关闭组合仍待验证。
- 同日标准 race 入口：`make test`（`go test -race ./...`）在允许本地 TLS/WSS 与辅助进程环境中完成，backup、containers、daemon、extensions、filesystem 等包均通过；未将环境缺失的 Windows/危险主机场景计入通过。
- 同日扩展现场能力接线：sandbox 将 `window.open`、`state.capture`、`state.restore` 连接真实桌面 store，返回实际窗口/标签身份并持久化 view 状态；总览扩展无节点时仍允许开窗，任务调用明确拒绝缺少 nodeId。前端检查、25/25 单测和构建通过。
- 同日扩展总览边界测试：新增无 nodeId 场景，验证总览扩展可开窗/恢复但任务调用被拒绝；前端 `npm test` 26/26 通过。
- 同日扩展节点身份修复：sandbox 每次 bridge 调用重新读取当前 viewTab resourceRef，避免跨节点切换/移窗后复用旧 nodeId；`npm run check && npm test`（26/26）与构建通过。
- 同日扩展 sandbox E2E：新增真实 bundle/opaque iframe/`postMessage` 场景，发现并修复 Vue 响应式状态无法结构化克隆导致的 bridge 中断；浏览器测试 1/1 通过（3.5s），SDK 报告已更新。
- 同日扩展浏览器组合回归：管理生命周期与 sandbox bundle 两项场景串行 2/2 通过（5.0s）。
- 同日 B10 浏览器全量回归：新增 Docker 与 sandbox 场景后完整串行套件 27/27 通过（约 1.3 分钟，退出码 0）；性能探针交互 p95 16.8ms、最大 60.3ms、1/60 长帧、heap 64,318,483 字节，API RTT p95/max 32.7ms，性能报告已更新。
- 同日资源恢复复核：桌面恢复/工作区切换现在为每个带 resourceRef 的 view 调用 `reconcileResource`，复核失败保留原现场并记录异常；改动后完整浏览器套件再次 27/27 通过，性能 p95 16.8ms、最大 62ms、1/60 长帧、API RTT p95/max 4.5ms。
- 同日最终后端 race 复验：`GOCACHE=/tmp/blora-go-race-final go test -race ./...` 退出 0，Master 75.886s，其余所有包通过；确认实例指标缓存和资源恢复相关改动无数据竞争。
- 同日监控改动后全仓 Go 回归：`GOCACHE=/tmp/blora-go-final-cache-2 go test ./...` 退出 0，Master 47.206s，所有包通过。
- 本轮创建的测试数据、进程、容器、网络规则及清理方式：fixture三个二进制与自身两个实例；停止用Ctrl+C，仅清理自身实例进程。新增专用镜像blora-isolated-e2e:20260909，manifest sha256:6486725488fcea62b3d2a2da0a395434a359a7642b536098a6cacfa04b0526c9；真实测试容器按随机run标签清理，未报清理错误。未改已有生产容器或防火墙。镜像与fixture目录保留作复现产物。
- 外部阻塞及影响范围：本机网络监听/浏览器/宿主测试需要工具提权，由auto_review按授权处理；Windows真机与可写cgroup委派仍缺失。Docker只读版本可用，真实容器待获准测试镜像环境验证。
- 2026-09-10 真实浏览器全链路复验：`cmd/devfixture` 在隔离 `.local/fixture-3172017811` 启动 Master 与两个 Daemon，并接入 Docker 29.4.1；`files-terminal.spec.ts` 定向 2/2（27.1s），`npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` 完整 13/13 通过（2.2m，退出码 0）。覆盖真实登录、节点/实例、文件、终端/PTY、日志、设置、模板创建、账号权限、Docker 浏览器边界及扩展链路；fixture 已用 Ctrl+C 停止，未留下活动会话。
- 2026-09-10 Docker 停止边界修复：Engine transport 的 `ResponseHeaderTimeout` 从 10 秒上调至 5 分钟，覆盖 Docker stop 的 300 秒服务端等待上限，单请求仍由 context 截止；`go test ./internal/containers -count=1` 通过。真实 Docker 浏览器 Compose 用例改为先从启动器打开 Docker 应用并提交 inline Compose 草稿，真实 apply/update/delete 1/1 通过。
- 2026-09-10 WSS 重连竞态修复：Master `nodeCall` 在控制连接已恢复但 bulk/interactive 数据链路尚未登记时等待最多 2 秒；控制连接消失立即返回 `NODE_UNREACHABLE`，请求取消仍尊重调用方 context。`go test ./internal/master -count=1` 与完整真实 13/13 套件通过；远程高延迟链路仍需单独测量。
- 2026-09-10 扩展回滚元数据修复：`Manager.List` 跳过 `.previous.json` 回滚快照，`read` 拒绝清单内 `appId` 与文件名不一致，避免升级回滚文件被误报为第二个已安装扩展；`go test ./internal/extensions -count=1` 通过。
- 2026-09-10 本地注册源与真实扩展回归：`cmd/devfixture` 预置 `fixture.reference` 的 `v1.0.0`/`v1.1.0` 包；真实 HTTPS 场景安装两版、读取启用 bundle、禁用返回 409、重新启用、opaque sandbox 执行并卸载，`extensions.spec.ts` 1/1 通过。静态 bootstrap 改为从同源 `/blora-extension-bootstrap.js` 加载，再在 opaque iframe 内创建 bundle URL；Master CSP 保持禁止 inline script，仅允许 `self` 与 `blob`。
- 2026-09-10 云端工作区回归：新增布局/引用默认同步、显式正文同步和云端副本打开的 mock 与真实 HTTPS 场景；`workspaces.spec.ts` 真实 1/1 通过，默认同步剥离草稿，显式同步后 revision=2 并恢复正文。
- 2026-09-10 OpenAPI 交付描述：新增 `docs/api/openapi.yaml`，覆盖 94 个路径、108 个操作以及 session/CSRF/idempotency、任务、扩展、系统和工作区合同；Python `yaml.safe_load` 校验通过，README 与验收矩阵已链接。
- 2026-09-10 防火墙租约终态修复：直接执行 `runFirewallApply` 的过期/取消回滚现在持久化终态阶段；回滚开始前先写入 `rollback_in_progress`，重启可继续识别并恢复。普通与 race 定向租约测试均通过。
- 2026-09-10 B10 真实全量复验：在 `.local/fixture-1866434427` 启动真实 HTTPS Master、两个 Daemon 和 Docker 29.4.1，`BLORA_E2E_CREDENTIALS=... npm run test:e2e:real -- --workers=1 --trace=off --reporter=line` **15/15 通过，2.2 分钟，退出码 0**；随后 Ctrl+C 停止 fixture。场景包含账号、桌面/权限、文件/PTY、日志/传输、节点维护、设置/编辑器、Docker/Compose、扩展两版本 sandbox 和云工作区。
- 2026-09-10 全仓验证：提权允许本地监听后 `make test` 的 `go test -race ./...` 通过（Master 79.757s），`make check`、前端 `npm run check && npm test`（28/28）、`make windows` 和 Linux 三个二进制构建均通过。默认沙箱运行 race 测试时的 `httptest` IPv6 监听失败属于权限限制，未计作代码失败。
- 2026-09-10 审计与节点维护增量：新增 `003_audit_node.sql`，把审计时间统一为毫秒并回填节点身份；`Store.Audits` 和 `GET /api/v1/audit` 支持管理员按用户、节点、资源、动作、RFC3339 时间及分页检索，UsersApp 增加筛选界面。节点撤销改为管理员确认后的事务幂等回执，递增代次、保留实例资源并写入节点审计，NodesApp 增加“撤销管理”入口；节点设置增加分组/标签且心跳保留管理员字段；新增存储和 Master TLS 集成测试均退出 0。新增问题修复：扩展 sandbox 不再把实例 ID 当作 nodeId；OpenAPI 更新为 95 路径、109 操作。证据见 `docs/acceptance/reports/audit-2026-09-10.md`。
- 2026-09-10 审计横向回归：任务接受和终态写入审计事件但不写入任务正文；节点/实例资源保留节点维度。`make test` 提权 race 全仓通过（Master 77.778s），`make check`、前端 28/28 与最终生产构建均通过；审计证据报告已补充。
- 2026-09-10 接口合同校正：OpenAPI 将节点撤销改为实际同步 `200` JSON 响应并补充 `400/404/409` 错误，登记者票使用 `201`、退出使用 `200`；YAML 重复键检查及路径/操作计数复验为 OpenAPI 3.1、95 路径、109 操作。
- 2026-09-10 备份与定时任务界面增量：默认注册 `blora.backups`，实例窗口提供真实入口；创建备份先读取源版本，恢复先生成不可变计划并要求 `overwritePlanHash` 显式覆盖确认，保留策略与定时任务使用真实 API、任务回执、幂等键和修订号。新增浏览器场景 1/1 通过；前端 28/28、check 和生产构建通过。Daemon 新增可选固定 argv `backupHookCommands` HookRunner，未配置时 save/pause/stop/hooks 仍明确返回能力不可用；真实 ENOSPC/掉电/长时调度和具体应用 hook 语义继续未验证。证据见 `docs/acceptance/reports/backup-ui-2026-09-10.md`。
- 2026-09-10 远程扩展目录源增量：新增 `extensions.CatalogSource` 与可配置 `RemoteCatalog`，只接受无凭据/无查询的 HTTPS 基址，固定读取 `index.json` 和版本包，限制响应大小、拒绝重定向、校验目录身份/摘要/重复项；Master `Options.ExtensionCatalogURL` 与 `cmd/master --extensions-catalog-url` 接入，目录与 URL 互斥且仍由管理员 API 调用，安装再次经 Manager 的完整性/签名校验。`GOCACHE=/tmp/blora-go-catalog2 go test ./internal/extensions ./internal/master -run 'TestRemoteCatalog|TestExtensionRemoteCatalogAdminAPI|TestExtensionCatalogBrowseAcquireAndUpgrade' -count=1` 通过；完整 `go test -race ./...` 通过，前端完整浏览器套件本轮更新为 30/30 通过。真实远程部署、高延迟证书轮换仍未验证。
- 2026-09-10 任务重试增量：任务中心新增失败/中断实例生命周期操作的关联重试 API。Master 仅重建当前实例配置并重新授权 `start/stop/restart/kill`，拒绝文件、终端、传输、备份和主机任务的通用重放；重试使用 `RetryOf`、独立幂等键和持久任务回执，权限边界与实例资源身份再次核对。新增 TLS Master 集成测试验证真实失败、关联任务、重复请求复用和未授权隐藏；任务中心 UI 显示已确认结果、关联重试入口和刷新后保留的请求身份。`GOCACHE=/tmp/blora-go-retry2 go test ./internal/master -run TestTaskRetryCreatesRelatedInstanceAttemptAndIsIdempotent -count=1` 通过；OpenAPI 更新为 96 路径、110 操作。远程节点故障下的长期重试组合仍待真实环境验证。
- 2026-09-10 任务列表/传输索引增量：任务中心 GET 支持 `offset`/`limit`（输出上限 1000 与 `nextOffset`），新增 `/transfers` 根任务列表并逐项复核传输源读与目标写权限，保留终态用于对账；真实传输集成测试覆盖新列表路由。列表不会绕过任务中心既有授权，也不会把内部传输子任务泄露为根任务。
- 2026-09-10 最终回归：备份列表/恢复计划分页拒绝负偏移，创建/清理/恢复明确要求 `Idempotency-Key`；真实备份归档/恢复定向测试通过。随后 `GOCACHE=/tmp/blora-go-race-final2 make test` 的全仓 race 通过（Master 79.536s），`make check` 与 OpenAPI 3.1 解析复核（96 路径、110 操作）通过；前端最新完整回归为 Vitest 28/28、Playwright 30/30，性能探针交互 p95 16.9ms、最大71.9ms、1/60 长帧、heap 63,210,217 字节、API RTT p95/max 6.6ms。进度文档同步修正 Docker/Compose 已接入的历史误记；Windows 真机、远程/故障负载和整项验收缺口仍按矩阵保留。

## 7. 决策记录

2026-09-13 A16入口基础组合：desktop.spec.ts新增真实入口专用窗口复用/显式新开，改名刷新保留但实际实例名/状态/runId不变，删除入口不删资源且无启动重启删除任务。36210退出0，1/1（3.6s），fixture1929604743会话9013已发送Ctrl+C，收取退出；无生产改动。下一动作：新建专用测试实例用于目标改名/删除/同名替换，验证旧resourceRef不会指向新资源，并补成员撤权/离线入口显示；不删除fixture基础实例影响其他场景。A16尚未全项完成。

2026-09-13 A16完整入口身份闭环：新增受控管理员 `DELETE /instances/{id}` 软删除墓碑，停止且无活动任务才可执行；列表/读取隐藏墓碑，Daemon重连快照忽略已删身份，配额统计排除墓碑，审计保留 instance.delete。存储删除幂等/运行中及活动任务冲突测试通过。真实 Chromium fixture-4074818222：目标改名/删除/同名替换旧resourceRef失效1/1（7.0s），双账号撤权入口失效1/1（3.3s），暂停精确Daemon入口显示节点失联并原runId恢复1/1（50.4s）；与基础专用窗口复用/显式新开/入口改名删除1/1（3.6s）合计覆盖A16全部条件，矩阵更新Linux已实现。fixture会话61153已Ctrl+C停止，无活动服务。下一具体动作：按矩阵审计A13/A14/A09/A12及E项剩余证据，优先补可推进的真实权限/资源迁移与长期负载；重新生成最终发行包，Windows真机和物理故障仍如实保留。
2026-09-13 A14创建权限闭环：真实 fixture-3663732060 双节点双账号浏览器1/1（3.5s）验证 member 可见两实例但创建入口禁用；仅节点2授予 `instance.create` 与 native 所需 `host.manage` 后 `/nodes/creatable`/创建向导仅列节点2，确认前撤权真实403且无实例，恢复授权后创建STOPPED实例并管理员按ID清理。报告与矩阵更新A14已实现（Linux Chromium）；fixture会话89877已Ctrl+C停止，无活动服务。下一具体动作：继续A13扩展完整能力/迁移、A09长期内存与A12 ENOSPC/物理故障等可推进验证，再更新最终打包与全量回归；Windows/远程环境仍未验证。

2026-09-13 A01立即刷新闭环：desktop.spec.ts新增活跃/后台参数场景，真实CDP IME提交、系统剪贴板CtrlV、连续输入撤销、鼠标拖窗、切后台或保持活跃后直接reload，正文/精确geometry/焦点/redo保留，再末次输入直接reload通过。16486退出0，2/2（6.0s）；fixture4013429189会话65859已发送Ctrl+C，收取退出。A01矩阵更新Linux覆盖，无生产改动。下一动作：A06实际vim/top与ANSI序号旧报告核对后更新覆盖；A13独立SDK资源任务恢复/未授权能力及A14/A16入口权限完整组合逐项审核。A09/A12和E项仍有明确未验证条件。

2026-09-13 A05传输边界闭环：收紧TestTransferDisconnectAndMidstreamPermissionRevocation为RUNNING且0<offset<total才断真实bulk连接，恢复后Verified/全字节一致/固定上传提交子任务恰一，96813 race退出0（12.490s），同时中途撤权未发布目标通过。结合restart断线原请求及实际OFFLINE/WAITING_NODE浏览器，A05矩阵更新Linux完整分支。无生产改动/活动测试。下一动作：核对A01活跃/后台窗口最近输入立即刷新、A06实际模式/序号以及A13/A14/A16已有证据缺口，按真实当前实现补齐；A09长时内存、A12物理ENOSPC及E项仍保留，不能宣布全量完成。

2026-09-13 A05任务卡等待：扩展actual paused Daemon，真实OFFLINE期间提交stop202，原任务WAITING_NODE并生产任务中心完整taskId卡显示等待节点；finallySIGCONT后原任务SUCCEEDED/卡成功/实际STOPPED，runId保持。30032退出0，1/1（50.7s），fixture3110999777会话42086已发送Ctrl+C，收取退出。无生产改动。下一动作：汇总A05重启/传输实际断线证据与新浏览器呈现，确认各阶段无重复副作用，再处理A01/A06/A13/A14/A16未归档完整条件；E项和全量最终回归/重新打包仍未完成。

2026-09-13 A05真实浏览器失联：nodes.spec.ts新增actual paused Daemon，完整配置路径/出生标识唯一识别自身进程，SIGSTOP后真实OFFLINE与页面等待提示，不伪造STOPPED；finallySIGCONT后ONLINE/原runId，再实际stop。32702退出0，1/1（50.5s）；fixture3470920766会话72955已发送Ctrl+C，收取退出。无生产修改。下一动作：任务中心在节点失联期间的排队/等待确认可见状态及重连结果，不仅验证资源页告警；结合restart/transfer后端断线证据审计A05。A01/A06/A13/A14/A16与E项仍须继续。

2026-09-13 A10身份边界闭环：新增真实存活PID错误StartTicks记录，Observe/ForceStop均ErrUnknown且正确身份原运行存活；结合未知token拒绝、丢PID的marker恢复、manager重建四项67725 race退出0（1.058s）。该项明确为过期身份注入，不声称实际PID复用。结合独立Daemon稳定运行/STOPPING SIGKILL，A10矩阵更新Linux覆盖；无生产改动，无活动测试。下一动作：A05实际浏览器节点断开/恢复的失联待确认呈现，结合已有restart/transfer断线避免重复副作用；A01/A06/A13/A14/A16及E项仍需核对，不宣布全量完成。

2026-09-13 A10任务中途SIGKILL：package-smoke新增--daemon-task-crash，27806退出0；/tmp/blora-release-smoke-fnx8gqb6，实际STOPPING阶段kill所属Daemon并确认-9，原目录恢复后原任务INTERRUPTED/daemon_restarted_reconcile_required，同key仍原任务且不重放，原runId/计数x保持；finally显式stop SUCCEEDED并退出所有自建进程，无活动会话。下一动作：核对A10运行身份防PID误认直接测试与恢复状态UI证据，之后A05浏览器失联/等待而非假停止组合；其他范围与最终重打包仍继续。

2026-09-13 A10独立Daemon SIGKILL：package-smoke新增--daemon-crash，44796退出0；/tmp/blora-release-smoke-zoh3sha_，原运行计数x，所属Daemon kill确认-9，同配置重启新startupId/同nodeId上线，原runId RUNNING及x保留。随后真实restart+Master SIGKILL恢复原任务，计数xx；最终停止实例及全部自身进程。无活动测试，无生产改动。下一动作：Daemon在restart/文件任务中途SIGKILL后的明确终态及不重复副作用，不能把当前稳定运行崩溃证据当作中途任务也已验；A05浏览器等待状态及其他范围继续保留。

2026-09-13 A03停止边界：新增真实双节点TestFailedStopBlocksReplacementButOtherNodeRemainsUsable，忽略TERM实例1秒停止失败、再启动失败且原runId/计数x不变，另一节点期间可启动停止；90516 race退出0（4.375s）。底层主PID先退/后代持管道、升级确认全后代退出、不升级有限失败三项41872 race通过（1.302s）。A03矩阵更新Linux真实覆盖，Windows/cgroup仍环境限制；无生产修改/活动测试。下一动作：A10独立Daemon进程SIGKILL后重启并识别原运行、活动任务结局，复用package-smoke自建进程管理但不要把Daemon.Close重建冒称崩溃。A05浏览器节点失联提示及其他范围仍待继续。

2026-09-13 A04响应丢失闭环：desktop.spec.ts新增真实restart202后route.abort丢浏览器响应，刷新保留确认并显式重试，原requestId/taskId、仅一个restart任务、实际runId一致；73971退出0，1/1（4.4s）。结合同key集成/管理断线/数据库重开/独立SIGKILL/A15资源串行证据，A04矩阵更新。finally实际STOPPED，fixture1671058723会话93296已发送Ctrl+C，收取退出。无生产修改。下一动作：A03完整异常后代/停止超时全局可用组合及A05浏览器节点断线等待状态、A10实际Daemon崩溃对账；同时E09最终重新打包/全套回归仍未完成。

2026-09-13 A04独立Master SIGKILL：scripts/package-smoke.py新增--master-crash，97578退出0；解包development-20260913归档、全新/tmp/blora-release-smoke-wkzbiotd，真实Master子进程在restart RUNNING时kill并确认-9，原目录启动后同key复用任务，成功后计数xx，完成重试不改runId。finally实际停止有限测试实例并停止所有自身进程，无活动会话。生产代码未改。下一具体动作：补客户端已经接受但响应真正丢失的重启重试验证（当前明确测试管理断线与Master硬中断，不冒称浏览器响应丢失）；再审计A04全部条件并推进A03/A05/A10。全范围仍未完成。

2026-09-13 A04服务数据库重开：新增TestRestartTaskSurvivesMasterAndDatabaseReopen，实际双Daemon+真实进程，restart RUNNING时关闭Master服务和SQLite后重开，同key复用任务，原任务成功；启动标记从x到xx，第二次重开已完成重试仍xx且runId一致。69010 race测试通过（8.12s），已收取最终状态，无活动测试。生产无改动。下一动作：补独立Master二进制进程SIGKILL中断而不是同进程服务对象重开，使用自建fixture/临时状态和启动计数标记，继续A04硬中断证据；其余完整范围仍未满足。

2026-09-13 A04/A05任务中断线：加强TestTwoNodesAuthorizationAndDurableLifecycle，restart任务RUNNING时关闭真实Daemon管理连接，等新generation后同key重试仍原taskId，原任务成功及runId在后续重连不变、最终停止；6548 race退出0（7.392s），无活动测试。未将Master RUNNING边界称为精确Daemon停止调用断点，生产无改动。下一动作：实现/验证Master进程中途重启后的同requestId对账入口，优先检查已有可重开store/atomic handler fixture与任务恢复逻辑，再补A04直接证据；其他完整范围仍待继续。

2026-09-13 A02独立视图闭环：新增真实Monaco200行共享模型双窗口，键盘首尾定位、IDB同draftId不同cursor/viewState、刷新后首行精确1/末尾200与原光标、继续输入落在原位置且另一窗口滚动不变。最终81566退出0，1/1（12.7s）；A02结合真实双账号冲突/迟到响应证据更新已实现。fixture3227456034会话87506已发送Ctrl+C，收取退出；无生产改动。下一动作：重新核对A01最近输入/后台窗口、A03/A04/A05/A10运行故障和A13/A14/A16旧矩阵证据，将缺少的真实组合逐项补齐；全范围及长期内存仍未完成。

2026-09-13 A02迟到响应增量：加强实际shared file windows，route.fetch拿到真实Master响应后gate暂缓交付，期间追加中文，释放后只出现重新读取确认而不改正文，两窗口/刷新都保留最新输入。33078退出0，1/1（9.6s），前序双账号冲突一并通过。fixture136228777会话80655已发送Ctrl+C，收取退出；无生产改动。下一具体动作：补共享Monaco模型的两窗口独立光标/滚动位置及刷新恢复，以真实键盘/滚动和持久editorView观测，不通过测试专用模型实现代替。A02仍进行中，长期内存与其他范围仍保留。

2026-09-13 A02真实外部修改增量：files-terminal.spec.ts新增双编辑窗口共享未保存中文、另一成员账号真实修改节点文件、旧基线冲突/服务器新版本保留、刷新保留两视图正文与比较版本、显式采用基线后保存。97420退出0，1/1（7.7s），fixture446296949会话59363已发送Ctrl+C，收取最终退出；生产代码未改。下一动作：补同模型两视图独立光标/滚动的实际浏览器断言，以及迟到服务器读取响应与新输入竞态，A02尚未全项完成。见shared-editor-conflict报告。

2026-09-13 A08文件与账号闭环：新增文件查看/写入/实例查看撤权真实浏览器，66268退出0，1/1（6.7s），保存拒绝/原服务器正文保留，刷新本地草稿undo redo/导出完整，新读取403及实例列表过滤；84359退出账号隔离复验1/1（3.5s）。结合前述实际PTY撤权，A08矩阵更新。测试恢复fixture原授权，fixture708836550会话22573已发送Ctrl+C，收取最终退出；无生产修改。下一具体动作：A02同用户双窗口同文件与另一账号外部修改的完整共享正文/独立光标滚动/版本冲突组合，检查现有editor-copy/settings-editor测试后补真实缺口。长期内存及其他A/E项仍未完成。

2026-09-13 A08活跃PTY：新增真实双账号独立上下文撤权浏览器测试，81846退出0，1/1（6.1s）；成员原生PTY单次文件追加，撤input后租约false/再输入无副作用，撤read后新请求403，刷新旧快照显示权限错误且不能输入。临时host.manage等授权finally撤销，原会话关闭；fixture1366325914会话41642已发送Ctrl+C，收取退出。无生产修改。下一动作：核对A08资源查看撤权期间编辑器草稿保留/账号切换证据，再推进其余仍进行中场景；长期内存采样仍未完成。见browser-revoke报告。

2026-09-13 A09慢端期限：新增TestLogSlowConsumerDeadlineAndCursorReattach，保持真实45秒期限，实际WSS读取不返信用，45.03165s收到SLOW_CONSUMER而非仅断线，原runId/cursor1重挂载收到后续第二标识。51543 race退出0（47.329s），自建实例Cleanup停止，无活动测试。生产Send确认等待socket写入，未发现错误帧仅排队就关闭的问题，无生产改动。下一具体动作：量化长时间混合流全进程内存，或按现有报告推进仍缺直接证据的A08运行中撤权浏览器组合；不要将协议错误送达自动升级为浏览器可见提示已测。证据在mixed-stream-control-2026-09-13.md。

2026-09-13 A09连接预算增量：混合流测试新增20次/50ms Conn.Stats采样，未消费<=256KiB、接收队列<=2MiB/64帧；最终29591 race退出0（15.959s），峰值82590B/79302B，批量传输与5秒停止也通过。7241首次错误假设填满信用窗口，实际日志按批次ACK更早背压；同时暴露失败时临时目录与无限日志进程清理竞态，已将显式stop注册到本测试Cleanup，最终复验通过。生产代码未改，无活动测试。下一动作：覆盖45秒慢消费者期限与可见SLOW_CONSUMER错误、重挂载游标恢复，并记录较长运行内存采样；目前短采样不等于全范围A09完成。见混合流报告。

2026-09-13 A09归档滚动增量：新增真实PTY无读取端64MiB输出、64KiB预算/16KiB分段测试，10470 race退出0（3.321s）；实际2.300s，380次采样峰值65358B，磁盘63317B与计数一致，迟到读Gap及末尾标识通过。见[混合流报告](../../acceptance/reports/mixed-stream-control-2026-09-13.md)。无活动测试/fixture。下一具体动作：给真实混合流测试采集Conn.Stats的队列/信用峰值并覆盖慢端饱和，结合持续运行进程内存观测补A09，不把短时64MiB归档测试称为长期稳定性。其余未满足验收继续保留。

2026-09-13 A09真实混合流增量：加强logs_integration_test.go慢日志测试，4097B循环输出、不归还首帧信用、同源节点8.4MB跨节点传输确认进度后发stop，5秒内真实STOPPED，完成归档可读、传输Verified且目标全字节一致。69124定向race退出0（14.120s）；57388首次新增字段名编译错误已修正复验。生产代码未改，无活动测试/fixture。见[混合流报告](../../acceptance/reports/mixed-stream-control-2026-09-13.md)。下一具体动作：补A09长期有界队列/归档滚动与慢端可见背压的实测，核对现有ArchiveRetention测试和协议队列指标入口；同时保留A01/A02/A03/A04/A05/A08/A10/A13/A14/A16及E项未完成组合，不能因新子项通过结束全范围任务。

2026-09-13 A15多视图闭环：`desktop.spec.ts` 新增双中心跨节点多标签/重复资源、独立导航、同浏览器任务提交两启动确认；最终31449退出0，1/1（7.9s），仅一个成功、失败限定旧运行未退出、结果runId一致，自动状态同步及刷新导航保持。早期36852被窗口遮挡桌面图标、59146匹配隐藏导航均已修正测试定位。实例最后STOPPED，fixture4012957923会话11603已发送Ctrl+C；收取最终状态。A15矩阵已更新，生产代码未改。下一具体动作：重新审计A01～A17剩余进行中行与最新专项报告，按明确缺口补真实故障组合，尤其A08/A09断线慢流和A03/A04运行归属；不要以当前几个绿色场景宣布全量完成。

2026-09-13 A17活跃终端闭环：新增actual PTY pointer真实鼠标拖出/合并立即刷新，原视图/会话/标签顺序保留，WSS Data输入不增加，真实文件副作用一次，最终99231退出0，1/1（8.6s）。首轮59340因空终端容器同时匹配失败，修正定位后通过。A17矩阵更新为已实现、Linux真实组合通过；生产代码未改。fixture1635242649会话4227已发送Ctrl+C，收取最终退出后继续A15。下一具体动作：检查InstancesApp资源标签开窗/导航/控制接口，新增两个实例中心、跨节点标签、同实例重复视图与并发控制的实际浏览器组合。全范围仍未完成。

2026-09-13 A17真实鼠标增量：`desktop.spec.ts` 新增双编辑标签实际pointer拖出/合并，每次松开立即刷新，原ID唯一/原窗口归属/标签顺序/中文正文/undo redo/另一正文独立均通过；86685退出0，1/1（3.8s）。首轮54242失败为隐藏Monaco也被定位的strict mode，修正可见选择器后通过，生产代码未改。fixture3185892968会话46036已Ctrl+C退出0，无活动测试；见[拖拽报告](../../acceptance/reports/drag-recovery-2026-09-13.md)。下一动作：补实际PTY拖出/合并立即刷新与WSS输入不重放，再实现A15真实跨节点双实例中心多视图并发控制组合，不能以菜单移窗旧证据代替新组合。

2026-09-13 A11事务中断闭环：真实Chromium在snapshots写入后、pointers请求时abort事务，数据库集合保持旧值，最新正文导出/刷新日志恢复通过；首轮82235发现未捕获tx.done AbortError，已修复RecoveryService立即观察done并在失败时abort/等待，最终75942为3/3（6.7s），79602前端生产构建及46/46单元通过。A11矩阵已更新实际浏览器覆盖，物理掉电不在本证据中。fixture2619400585会话72052已Ctrl+C退出0，无活动测试。下一条具体动作：A15/A17真实跨节点实例多窗口共享状态、独立视图、标签拖出/合并立即刷新组合；先核对现有桌面/实例实现和真实测试避免重复。全范围仍未完成。

2026-09-13 A11浏览器故障增量：新增 `recovery-failure.spec.ts`；实际HTTPS/Monaco中Storage同步日志与IDB snapshots.put同时QuotaExceededError，最新正文仍可真实下载导出，恢复写入后刷新保留；schema0恢复与schema999原始记录两次刷新可导出均通过。会话37838退出0，2/2（4.6s）；fixture1827070879会话73966已Ctrl+C退出0，无活动测试。见[恢复报告](../../acceptance/reports/recovery-quota-2026-09-13.md)。下一条具体动作：补IDB事务在快照写入后/指针切换前中断的浏览器验证，确认当前/旧快照不混合，再继续A15/A17真实跨节点多窗口移窗立即刷新；全范围尚未完成。

2026-09-13 A06 实际程序增量：`files-terminal.spec.ts` 新增 vim 未保存中文/刷新/真实保存及 top 刷新/退出后真实写文件，最终会话65248退出0，1/1（9.0s）。首轮97839因对原始增量ANSI日志断言连续整行失败，修正测试观测方式后通过，生产代码无改动。fixture4180777667会话36955已Ctrl+C停止；独立发行包smoke39175也退出0且进程停止。见[终端报告](../../acceptance/reports/terminal-programs-2026-09-13.md)。下一条具体动作：检查并补A11真实浏览器存储配额故障/恢复迁移验证，再推进A15/A17跨节点多窗口移窗立即刷新组合；全范围仍未完成。

2026-09-13 解包独立链路：新增 `scripts/package-smoke.py`，会话39175退出0；临时目录 `/tmp/blora-release-smoke-m3u7kpav`，SDK/参考扩展离线干净依赖安装、源码构建、归档签名工具运行、Master初始化/TLS/HTML/登录及两个解包Daemon ONLINE均通过，三个进程已停止。下一步正在真实vim/top刷新恢复，fixture `.local/fixture-4180777667` 监听9444，会话36955；浏览器完成后必须Ctrl+C停止fixture。报告见发行包报告，Windows真机仍缺失。

2026-09-13 本地发行包检查点：`scripts/package.py` 与 `make package` 已接入，六个归档生成于 `dist/releases/development-20260913`。会话45917构建退出0，5282重复打包及独立清单/摘要校验退出0；Master每平台183文件、Daemon78、SDK21、web115。同输入归档逐字节一致，版本不同内容拒绝覆盖，Windows仍仅交叉编译。README与[发行包报告](../../acceptance/reports/package-2026-09-13.md)已补充；这些后续文档不在旧归档里，最终需新版本再打包。无活动测试或fixture。下一条具体动作：将这组Linux Master/Daemon和SDK解包至本任务临时目录，验证独立初始化/健康检查及SDK离开仓库重新安装依赖和构建，然后继续A06实际vim/top、A11浏览器配额、A15/A17真实组合；全范围未完成。

2026-09-13 A07最终通过：TasksApp接回原本机上传，复核账号/任务/动作/requestId/资源及原文件身份；前端46/46单元、类型/生产构建通过。真实关闭标签后万条文件解压完成，16MiB上传WAITING_CLIENT；新browserTabId从983,040B原偏移复用原taskId续传成功，目标摘要正确，约1.1分钟通过。fixture-3124651407已停止，无活动会话。A07标已实现并明确Linux通过，Windows留E09。最终storage版本保护后的check/build/windows也已通过。下一具体动作：产出独立Master/Daemon与SDK发行包及校验清单，再按矩阵补A06实际全屏程序、A11浏览器配额、A15/A17等真实组合；仍不称全量完成。见[上传关闭证据](../../acceptance/reports/upload-close-2026-09-13.md)。

2026-09-13 A07继续：发现新浏览器标签丢失本地上传入口而服务端检查点仍在，新增TasksApp“继续本机上传”及restoreUploadFromTask，GET再次核对当前账号/原任务/资源/请求身份和文件指纹，重建本地引用后复用原幂等请求；不创建新上传任务。3项定向单元通过（430ms），前端生产构建及类型检查通过；最后追加action/requestId核对尚待最终复验。真实upload-close.spec已编写：万条目录压缩→解压与16MiB本机上传并行→关闭实际标签→新标签从任务中心选择原文件续传。当前fixture-3124651407监督会话70696，浏览器测试20890；需接收结果。上轮最终make check/build/windows已通过，无其他构建会话。下一动作：修复真实A07失败并记录结果，随后发行包与其余完整场景。

2026-09-13 512MiB实际传输通过：fixture-4036895406同一任务659de8466bcbe9ef91aed784ce047b90完成，543.324s（约0.942MiB/s），源目标独立SHA256一致。首轮小文件观察预算不足手动中断，后台未取消；新测试验证资源/源版本后续接原任务，约5分钟通过，p95 42.4ms、2/357长帧、未确认峰值87,404B，8窗口/2PTY/万条目录/512MiB传输全程采样并行。filesystem完整race1.300s；fixture已停止，无活动浏览器。另补storage未知/不连续迁移拒绝启动，完整存储race17.216s通过，新增运维手册。最终check/build/windows会话52400待接收。下一动作：记录最终构建结果，然后按矩阵继续A07并行解压/本机上传关闭网页场景，以及发行包与长期/平台验证；不要把现有大量旧矩阵行当作已全覆盖。见[传输证据](../../acceptance/reports/transfer-proof-2026-09-13.md)、[升级保护](../../acceptance/reports/upgrade-storage-2026-09-13.md)。

2026-09-13 大文件优化执行中：新增filesystem/read_proof.go的64KiB块SHA索引（每服务64条/8MiB上限），原ReadChunk不变；Daemon新增file.transfer.stat/chunk，Master中间读取使用索引并在发布前完整核对源版本。文件定向race1.045s、原传输TLS回归46.307s、新增“已读前缀改写并恢复mtime不能发布”与实际源Daemon重启重建索引25.963s通过。fixture增加performance-transfer-mib（1..1024）。当前无fixture，make check/build/windows和fixture二进制正在构建；下一动作：接收构建结果，启动--performance --performance-transfer-mib 512的私有fixture，运行真实performance.spec记录吞吐和本地响应，再补完整filesystem回归。见[校验索引报告](../../acceptance/reports/transfer-proof-2026-09-13.md)。

2026-09-13 E08短时真实场景最终通过：识别SwiftShader软件WebGL并降级默认终端渲染器；不透明窗口和拖动轻效果、64KiB/128条输出增量及大检查点持久化确认已接入。最终fixture-746611301串行PTY17.9s+性能35.7s，2/2通过（54.7s）；p95 38.8ms/max48.5ms，0/302长帧，8窗口/2PTY/万条目录/16MiB传输全程并行，未确认峰值87,353B、目标校验成功。浏览器大屏幕恢复/降级2/2（13.2s），前端43单元及构建通过。make check/windows最终通过；所有fixture已停止，无活动会话。下一具体动作：ReadChunk每块两次整文件哈希及Master每8块再次Stat导致512MiB吞吐极低；设计不削弱版本/源变更/目的校验合同的有界读取优化并做真实大文件复验，之后继续长期E08及其他未验证组合。完整证据见[性能报告](../../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 E08绘制诊断续接：CPU诊断主要落在原生program；窗口整体不透明后p95降至132.4ms但仍失败。增加硬件WebGL能力探测（实测仍选WebGL，降级未验证）及未确认字节峰值观测，下一样本169.1ms、峰值87,310B，仍失败。全部前端单元43/43通过；128条/64KiB终端增量上限、类型和生产构建通过。新改动拖动期间轻阴影/关闭背景模糊，正在fixture-1196973665（监督会话67716）运行性能测试39784；需接收结果及图形设备字段。之前4028490118/719092740/2394455507均已停止。下一动作：判定绘制改动效果、验证渲染降级分支及最终PTY回归，再继续大文件哈希性能与E08未达标项。

2026-09-13 E08第二次优化：终端恢复增加64KiB输出增量日志，完整屏幕低频合并，旧记录兼容；浏览器7.7s、真实PTY19.4s、生产构建/类型检查通过。混合负载42.5s仍失败：p95 219.6ms、272/288长帧，heap降至43.9MB；双PTY、万条目录及16MiB传输全程并行已确认。fixture-1188777270已停止，新增可选BLORA_PERF_PROFILE只汇总登录后CPU函数时间。新诊断fixture会话47116正在启动；下一动作：读取其凭据路径，带BLORA_PERF_PROFILE=1运行performance.spec，区分脚本与合成绘制瓶颈后继续优化。见[更新证据](../../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 E08实测未通过：真实8窗口、2PTY、万条目录和16MiB跨节点传输已运行；反馈p95 278.2ms，303/319长帧，传输验证成功但先于采样结束，不满足全程并行。RecoveryService合并在途重复快照，恢复测试9/9、前端构建及PTY/协议定向race通过；终端逐事件完整屏幕序列化仍需优化，512MiB反复全源哈希导致吞吐严重偏低。所有本轮测试结束，fixture-2336190821/3188205192已Ctrl+C停止，无活动会话。下一具体动作：终端恢复输出增量化，保留立即刷新及不重发输入，复验真实PTY刷新后重新测量全程负载；继续大文件传输性能。见[真实性能失败证据](../../acceptance/reports/performance-real-2026-09-13.md)。

2026-09-13 原生进程树指标：Daemon按运行成员出生身份聚合CPU/RSS，最多1024成员，成员变化拒绝旧快照并最多三次重采样；Windows Job枚举补句柄归属和出生身份，界面显示范围/成员数及RSS共享页口径。真实忙子进程TLS聚合2.380s、最终daemon/monitor/Master指标race1.324s/1.270s/5.963s、check/build/windows及前端构建通过。无活动测试/fixture。见[进程树证据](../../acceptance/reports/process-tree-metrics-2026-09-13.md)。下一具体动作：检查现有performance.spec与fixture，补E08真实两终端、万条目录、后台传输和八窗口并行测量；保留短命成员历史账本、Windows真机及第二账号登录等待问题边界。

2026-09-13 fixture清理/原生CPU：编译监督二进制直接运行fixture-404818613，启动实例后Ctrl+C退出0，数据库证明start及两次kill成功、两实例STOPPED；make fixture改为直接执行编译二进制。Linux真实CPU双次采样、出生身份与RunID基线已实现，主机统计不重复guest，CPU基准在API/界面区分。monitor/Master定向race1.269s/4.667s、最终check/build/windows和前端构建通过。无活动会话/fixture。下一动作：原生实例所属进程树CPU/RSS聚合与成员变化边界（当前是主进程），随后E08实际并行负载及既有登录超时诊断。见[CPU/清理证据](../../acceptance/reports/native-cpu-2026-09-13.md)，不把主进程采样称为整个运行总量。

2026-09-13 网络/可见性收尾：真实Docker浏览器最终1/1（29.9s）验证网络值、七秒最小化无原五秒轮询及恢复即时采样；修正单窗口直接恢复/多窗口选择器测试。第二账号界面登录曾超时，专项使用真实API登录，问题保留。最终构建和指标定向race通过。fixture-75425925/515580840/2134529472均停止；四个失败轮次容器按记录RunID核对后停止/删除，成功轮次任务已清理，无活动测试会话。Linux fixture子服务新增Setsid避免Ctrl+C过早停止服务，尚需验证信号清理。下一具体动作：先实测新fixture信号清理，随后补Linux原生进程CPU采样（当前明确不可用）；继续E08完整负载和登录超时诊断。见[可见性报告](../../acceptance/reports/monitor-visibility-2026-09-13.md)。

2026-09-13 监控网络/可见性执行中：节点与实例显示网络收发累计MiB，缺失或不可用不显示零；Point保留真实零网络计数，原生进程未采集的网络/磁盘标不可用。MonitorApp接visible，后台指标30s/历史60s，恢复立即读取对应范围；实例嵌入传递visible。发现Linux原生进程CPU未采集且先前默认为0，先改为不可用，下一项补真实采样。check/build/windows、前端构建、monitor/runtime定向race1.208s/1.189s通过。首次浏览器最小化七秒无请求成立，恢复步骤因测试错误任务栏定位而失败；旧fixture-75425925已停止并清理。最终fixture-515580840运行会话26367，浏览器复验会话34435仍待收取，需先完成验证并停止fixture再继续。

2026-09-13 监控代次缓存修复：采样后重读实例和授权，拒绝变化或不匹配RunID；缓存按当前代次选取，不跨代次回退。真实TLS原生实例断开/重建Daemon链路、同代次旧值时间戳、断链撤权403、重连实时恢复、实例重启后无新采样503均通过；最终监控定向race5.043s、check/build/windows、OpenAPI103路径通过。无活动测试/fixture。见[监控证据](../../acceptance/reports/container-metrics-2026-09-13.md)。下一具体动作：按F10/E08核对节点/实例网络指标的界面呈现，以及最小化窗口的绘制/采样行为和真实负载证据；继续全范围验收，不将Daemon.Close测试等同物理断网或Windows真机。

2026-09-13 F10真实指标端到端：实例监控标签接MonitorApp、实例范围不发主机请求；修复全新节点首次容器启动缺少InstanceRoot初始化。真实双Daemon/Docker浏览器1/1（15.4s）验证授予/撤销权限、CPU/128MiB内存、刷新、停止后stale和无主机请求；全新根目录真实Docker2.950s、runtime/daemon race2.982s/30.584s、check/build/windows、前端构建及41/41通过。旧fixture-1599939518与最终fixture-3062496594均已停止，自有运行资源清理，无活动会话。失败与修复详见[容器指标证据](../../acceptance/reports/container-metrics-2026-09-13.md)。下一动作：监控跨运行代次缓存及实际节点断开/重连的语义和可见诊断，随后继续全范围缺口；不把实例停止后的缓存当节点失联。

2026-09-13 F10容器实例指标增量：发现并替换Daemon固定不可用分支，运行适配器采集Docker双样本CPU、内存usage/limit和有界网络计数，采样前后核对归属与StartedAt；界面显示真实内存口径。HTTP身份/重启/计数边界race最终1.040s、真实隔离Docker生命周期含指标2.747s、最终check/build/windows和前端构建退出0。无活动测试/fixture。下一条具体动作：真实Master/双Daemon fixture创建隔离实例，验证监控API权限、浏览器CPU/内存口径和历史/失联标记；然后继续全范围缺口审查。见[容器指标报告](../../acceptance/reports/container-metrics-2026-09-13.md)，F10/E01不计整项完成。

2026-09-13 Linux Compose CLI恢复完成增量：启动前持久令牌、启动后会话号，运行结束/启动恢复按身份与pidfd清理，未确认保留证据并阻止新Engine变更；Daemon连接前恢复门禁已接。中断后核对追加测试已落盘、无遗留测试命令，最终containers/daemon race7.900s/30.518s、check/build/windows退出0；真实Docker生命周期27.047s通过。无活动会话/fixture。见[CLI恢复证据](../../acceptance/reports/compose-cli-recovery-2026-09-13.md)。下一动作：核对F10容器模式实例指标是否完整接入真实Docker统计，补明确代码缺口；其余SDK按规划公共能力语义审查，不凭旧“其余能力”笼统描述无限扩张。完整Daemon硬崩溃/平台及故障组合仍继续。

2026-09-12 Compose检查点硬崩溃/浏览器复验：Linux独立执行器SIGKILL后日志逐字节保留、查询INTERRUPTED/Unknown、相同任务禁止重放，race最终1.381s退出0。实际Docker浏览器输出/退出标识及刷新1/1（45.9s）通过；fixture-4124677712已Ctrl+C停止，测试部署及卷清理，无活动测试。OpenAPI3.1解析103路径。下一条具体动作：补Linux Compose CLI运行身份持久记录和Daemon启动遗留进程恢复，防止旧CLI仍执行时新任务并行修改同一Engine；当前测试按专属随机令牌清理不能替代产品恢复。继续其余SDK能力和全范围平台/故障验收，见[输出报告](../../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose运行中输出检查点：每250ms仅保存变化、每流32KiB与截断标识；命令启动前记录未完成，退出后completed，界面不把缺失退出记录误称仍在运行。落盘失败取消并等待CLI，保留Unknown。受控真实CLI/跨管理器磁盘读取及取消race2.470s、containers全包6.520s、追加目录故障定向race1.952s通过；make check/build/windows、前端构建退出0。无活动会话或fixture。下一动作：Daemon/执行器硬崩溃的日志保留与遗留CLI处理审查、真实浏览器复验，然后其余SDK资源能力与全范围平台/故障验收。详情见[输出报告](../../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose输出最终构建收取：会话9398前端生产构建退出0，源码/产物包含最终说明文案。无活动测试、fixture或测试资源待清理。下一动作保持为运行中有界输出检查点与崩溃窗口验证，随后继续完整范围；详见 [Compose输出报告](../../acceptance/reports/compose-output-2026-09-12.md)。

2026-09-12 Compose阶段输出闭环：成功/失败stdout/stderr前32KiB、独立截断标识、最多16阶段写入有8MiB保护的操作日志；operation-output单阶段base64查询保持96KiB，Master/Daemon主机权限与负页码校验，任务详情分页/刷新接入。实际CLI进程双满流/中文/NUL/磁盘重开定向race2.120s、containers全包6.004s、Master TLS1.917s，真实Docker/Compose浏览器1/1（45.9s）、前端41/41、check/build/windows与OpenAPI103路径通过，见 [输出证据](../../acceptance/reports/compose-output-2026-09-12.md)。fixture-2005453535已Ctrl+C停止，测试部署及卷已清理；最终文案前端重建尚待收取，此外无活动测试/fixture。下一动作：运行期间有界流式输出与崩溃前检查点，保持命令结果不明语义；然后其余SDK资源能力与全范围平台/故障验收。

2026-09-12 拉取EOF/末条观测闭环：正常EOF刷新最后一条采样，32MiB+1探针辨别超限，不继续旧镜像检查并保留Unknown。真实HTTP协议替身32MiB边界/迟到错误、末条计数与持久化定向race3.743s，最终check/build/windows通过，见 [容器进度证据](../../acceptance/reports/container-progress-2026-09-12.md)。无活动测试或fixture。下一动作：Compose阶段成功/失败stdout、stderr有界持久记录与截断标识、受host.manage保护的读取/浏览入口；查询保持96KiB边界，不把整个操作输出塞入控制消息。随后继续流式诊断、真实CLI/Engine与全范围SDK/平台/故障验收。

2026-09-12 容器进度回归收取：会话42746的containers全包race退出0（2.553s），真实Docker条件入口未启用，不将跳过计作真实拉取。无活动测试/fixture。下一动作仍为镜像采样末条计数与限流EOF边界，随后Compose输出持久化/受控读取及完整任务诊断；见 [进度报告](../../acceptance/reports/container-progress-2026-09-12.md)。全范围目标不变。

2026-09-12 镜像进度增量：修复Engine progressDetail被丢弃，emitProgress/节点日志/Daemon任务修订保留Layer与Current/Total，任务界面显示分层计数。协议替身首次漏/version协商失败已修正，定向race1.024s、check/build/windows、前端生产构建通过，见 [容器进度报告](../../acceptance/reports/container-progress-2026-09-12.md)。正在容器全包race回归，无活动fixture或生产操作。下一动作：收取全包结果，修复采样末条进度保留与流量上限EOF语义，再补Compose实际输出的有界持久记录/浏览入口及真实验证；全范围继续。

2026-09-12 任务阶段详情闭环：发起者/时间/含等待耗时、终态冻结与缺失时间边界、专用详情和持久阶段分页接入。真实TLS57修订/诊断限长/负载排除/撤权race1.973s，前端41/41，真实文件任务详情浏览器1/1（3.8s），check/build/windows及OpenAPI3.1/103路径通过，见 [阶段记录证据](../../acceptance/reports/task-stages-2026-09-12.md)。fixture-186544590会话65750已Ctrl+C停止，无活动测试或fixture。下一动作：审查各任务执行器的实际进度与输出记录，优先补Docker镜像拉取/Compose等长任务的有界持久诊断与实时阶段，继续其余SDK受控资源能力及全范围平台/故障验收。阶段记录不等于程序stdout，不能因此提升F09整项通过。

2026-09-12 任务阶段记录进行中：TaskStagePage投影持久task_events，按任务修订号倒序分页，phase/error限512/4096字符并标截断，不返回payload/result；GET /tasks/{id}/events复核当前canTask。任务列表增加详情入口，专用详情显示发起者、接受/派发/更新时间、含等待耗时和阶段页；终态耗时冻结、无效时间明确不可用。真实TLS57修订跨页/诊断/撤权race1.973s、时间边界2/2、check/build/windows与前端生产构建通过。fixture-186544590会话65750活动，正在真实文件任务详情浏览器验证；前端全单元与OpenAPI结果待收取。下一动作收取、修复失败、补报告；阶段记录只代表真实任务状态历史，不冒充执行程序stdout。

2026-09-12 任务统计/筛选闭环：定向TLS race3.499s、真实101任务/后端状态筛选/刷新/任务栏一致/通知详情组合2/2（9.8s）、前端39/39、生产构建、make check/build/windows、OpenAPI3.1/102路径通过，见 [任务历史报告](../../acceptance/reports/task-history-2026-09-12.md)。WAITING_NODE测试样本自动派发造成的两次失败已保留；WAITING_CLIENT稳定样本复验通过。fixture-4241791769会话99611已Ctrl+C停止，无活动测试或fixture。下一动作：默认任务显示actorId、创建/更新时间与实际耗时；通过已持久task_events提供获准任务阶段记录分页和真实详情入口，不把阶段日志伪称进程stdout；然后继续完整SDK资源能力、平台/故障及其余验收缺口。全范围仍进行中，系统通知真桌面及Windows真机未验证。

2026-09-12 任务统计/全历史筛选进行中：GET /tasks/summary按500条活动根任务页扫描全部当前授权记录，5秒超时返回错误不返回部分总数；任务栏读取该API且错误显示待确认。before历史模式增加服务端state条件，UI筛选重置游标并恢复筛选。check/build/windows、前端生产构建、OpenAPI3.1/102路径通过。新增TLS测试前两次因WAITING_NODE样本被真实派发器改变状态失败（管理员计数1非2），改WAITING_CLIENT稳定样本复验中会话97929；新fixture会话99611待就绪。下一动作收取定向race与fixture路径，运行notifications.spec.ts含101真实任务/筛选刷新/完整任务栏计数，再更新证据。无生产操作；整项范围不缩减。

2026-09-12 任务历史闭环：最终storage游标race12.905s、Master授权1.883s、真实101任务分页/刷新与通知详情组合2/2（10.9s）、前端39/39、check/build/windows通过，见 [历史报告](../../acceptance/reports/task-history-2026-09-12.md)。已记录模板语法、通知定位和测试异步队列期限三个失败/修复；fixture-501473755会话67184已Ctrl+C停止，无活动测试/fixture。下一动作：任务栏完整获准未终态统计（目前旧1000条列表会漏旧未完成任务）、任务状态筛选传到历史查询；随后补默认任务元信息/日志及其余全范围缺口。上一全仓race通过，当前增量定向通过，不将Windows构建称真机验收。

2026-09-12 任务历史进行中：RootTaskPage稳定before游标、权限过滤、真实界面分页与现场、专用详情直接GET接入；1005任务插入稳定性定向race最终12.905s，Master分页授权1.883s通过。101个真实mkdir任务浏览器分页/立即刷新通过；同套件后续通知测试因前一测试尚在完成的同类任务产生40条通知而定位歧义失败，现通知显示资源身份，测试按本次taskId精确定位（不取第一条）。最终前端与两端重建进行中；fixture-501473755会话67184仍活动。上一全仓race78624已退出0，Master175.982s/extensions164.323s通过；本历史分页在全仓之后修改，单独证据不冒充全仓覆盖。下一动作收取构建并复验真实两场景，更新报告/矩阵；其余任务统计/迁移与全范围缺口继续。

2026-09-12 通知最终组合收取：修复初始化竞态后3/3（9.8s）通过；含真实元数据存储故障禁止PATCH/显式重试、独立通知来源和刷新、真实文件任务完成通知。fixture-1098943595会话52328已Ctrl+C停止，无活动fixture。全仓race78624仍运行，extensions164.323s通过，其余已输出包通过；待收取Master。详见 [通知证据](../../acceptance/reports/notifications-2026-09-12.md)。下一动作：任务中心历史读取突破RootTasks固定最近1000条，加入稳定游标与权限过滤，专用任务详情直接查询；实现真实界面分页/刷新后再验证。系统通知真桌面展示、全范围平台与故障组合仍保留。

2026-09-12 通知组合失败修复：首轮三场景2/3通过（含注入恢复日志失败时禁止元数据PATCH、恢复后同请求显式重试）；独立样例输入在初始化readData完成前可用，随后被恢复覆盖导致笔记丢失。现将初始输入/按钮禁用到恢复和监听绑定结束，三个包重建通过；通知测试用受控未完成数据请求确定性覆盖禁用时机。最终三场景复验进行中，fixture-1098943595会话52328仍活动，全仓race会话78624仍运行。新增通知日志错误单元2/2通过。无生产操作；继续收取并修复失败，不能写成全量完成。

2026-09-12 任务通知进行中：WSS通知模式、首次当前快照、续传游标、服务端授权/内部子任务过滤/终态最小摘要接入；默认桌面持续订阅，通知和游标同日志提交，显式系统通知启用/关闭及点击详情代码接入。真实WSS游标/权限/race2.321s、真实文件任务浏览器1/1（3.6s）、前端38/38、check/build/windows通过。发现RecoveryService.commit失败仅记录状态，扩展restore/通知桥接现已检查protected并拒绝成功回执；最终前端构建通过。正在fixture-37470389后继fixture会话52328验证本地日志故障时禁止元数据PATCH与恢复后显式重试；旧fixture均已停止。下一动作收取新fixture私有路径/定向浏览器组合与新增错误测试，更新证据，然后全仓race及其余F/A/E缺口。系统通知在真实桌面操作系统的展示仍未验证。

2026-09-12 通知子项最终构建收取：会话77057的make check/build/windows退出0，无活动测试或fixture。后台任务通知下一步复用已有持久task_events和WSS/Protobuf事件通道，增加通知订阅的起始快照/持久游标，避免只轮询最近1000任务漏掉快速完成与旧任务；通知及游标同一恢复日志提交，刷新/断线不重放已读历史。尚未开始该代码，上一通知报告中的范围限制仍有效。

2026-09-12 通知SDK子项：notification.publish/host.notify、服务端实时能力/app.use校验、工作区持久通知托盘与来源跳转接入。独立包浏览器1/1（4.7s）、最终TLS通知/云端策略race2.195s、前端37/37及构建、SDK三个包、OpenAPI101路径通过，见 [通知报告](../../acceptance/reports/notifications-2026-09-12.md)。fixture-3796281576与早前fixture-115163169均已Ctrl+C停止；最终make check/build/windows会话77057待收取。下一动作：接入后台任务终态通知（关闭任务窗口仍工作），按已授权真实任务变化触发，持久去重、点击跳任务详情；系统通知必须用户显式启用且不重放历史完成项。其他全范围缺口继续保留。

2026-09-12 独立扩展元数据浏览器闭环：实例窗口、名称草稿、稳定请求现场与显式重试接入，提交前等待本地现场持久化。真实双节点浏览器1/1（4.7s）、独立包Master定向race30.708s通过，SDK与三个版本样例构建通过，见 [元数据报告](../../acceptance/reports/extensions-metadata-2026-09-12.md)。fixture-115163169 会话93883暂保留供后续验证，无生产操作；下一动作：通知SDK与带应用来源的站内通知能力，然后继续其余受控资源与全范围验收。

2026-09-12 SDK元数据写入增量：updateResource/resource.write与PATCH扩展资源API接入，实例name/group/tags、显式configRevision、事务幂等、app.use及instance.configure复核，输出不含执行配置。transport10/10、SDK与前端检查/构建、最终check/build/windows及OpenAPI解析通过。TLS新增测试的包序列化和启用路由错误已修正，最终race1.909s通过（含修订号缺失拒绝）。见 [元数据报告](../../acceptance/reports/extensions-metadata-2026-09-12.md)。下一动作：独立扩展示例增加实例元数据编辑与稳定请求恢复，跑真实浏览器链路；随后通知SDK及其他受控资源能力。无活动测试、fixture或生产资源操作。

2026-09-12 计划任务分页贯通：Linux.timer完整排序、Windows COM完整路径Ordinal保留后续最小集合，limit/after经API/RPC传递，界面分页和立即刷新恢复。205项跨三页race1.036s、Master参数TLS1.963s、浏览器1/1（6.6s）、最终systeminfo全包race1.246s、check/build/windows、前端生产构建与OpenAPI解析通过，见 [系统报告](../../acceptance/reports/system-adapters-2026-09-12.md)。Windows COM运行仍未验证。下一动作：复核剩余SDK受控资源写入/通知能力及生产任务语义，补真实安装示例，再继续全范围故障和平台验收。无活动测试、fixture或主机变更。

2026-09-12 服务分页：Linux/Windows名称游标、Daemon/Master传递与SystemApp首批/下一批/刷新恢复接入，解决只返回前100/200项。systeminfo race1.187s、浏览器1/1（6.2s）、check/build/windows、前端check/build及OpenAPI解析通过，见 [系统报告](../../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：计划任务分页，Linux.timer和Windows完整COM路径分别排序并稳定游标；补服务器参数验证和浏览器恢复，然后继续全范围剩余项。无活动测试、fixture或主机变更。

2026-09-12 Windows服务操作增量：原生SCM启停与状态轮询取代sc.exe，重启等STOPPED后才启动、启动等RUNNING，整体30秒并检查取消；合法Unicode/空格服务身份与Linux命令参数边界已补。systeminfo race1.169s及最终make check/build/windows通过，真机启停/超时尚未验证，见 [系统报告](../../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：补服务/计划任务完整分页（目前返回前100/200项）、视图查询现场及状态错误呈现，再继续全范围权限/扩展/故障组合。无活动测试、fixture或主机变更。

2026-09-12 Windows计划任务增量：固定PowerShell/Task Scheduler COM脚本替换schtasks本地化LIST，完整路径/数值状态/下次时间JSON；Enabled变更后回读确认。Unicode/空格身份与数据参数隔离、20秒/1MiB边界、实际依赖探测接入，systeminfo race1.200s、Windows测试程序交叉编译及最终make check/build/windows通过。真机COM脚本尚未运行，不能算真实Windows验收。见 [系统报告](../../acceptance/reports/system-adapters-2026-09-12.md)。下一动作：系统服务/计划任务完整分页及Windows服务操作合法名称/完成确认，随后继续权限/扩展/故障组合缺口。无活动测试、fixture或主机变更。

2026-09-12 系统适配增量：Windows服务查询改原生SCM最小权限枚举/状态，补只读真机测试入口；Linux定时器返回.timer身份，禁止通过计划任务入口启停.service并隔离命令参数。systeminfo全包race1.172s、Windows测试程序交叉编译及最终make check/build/windows通过，见 [系统报告](../../acceptance/reports/system-adapters-2026-09-12.md)。Windows真机未验证。下一动作：Windows计划任务以结构化原生/COM数据替换schtasks本地化LIST解析，补合法Unicode/空格任务身份和操作验证；随后完善系统列表分页及其余全范围缺口。没有活动测试、fixture或主机变更。

2026-09-12 旧防火墙接口清理完成：删除旧ApplyFirewall/RollbackFirewall、全局reload及无用单份比较函数；无双快照的变更明确拒绝。原错误/取消/恢复失败/外部冲突测试迁移到Snapshot接口，最终systeminfo/Daemon定向race1.168s/29.595s通过（含独立命令进程），make check/build/windows通过。见 [快照证据](../../acceptance/reports/firewall-snapshots-2026-09-12.md)。无活动测试/fixture，无主机规则操作。下一动作：Windows系统服务查询改用SCM或结构化原生数据，计划任务替换本地化文本解析；修复Linux定时器名称误取service列，补平台构建和测试入口，再继续全范围剩余项。

2026-09-12 Daemon命令进程证据：新增Linux临时firewall-cmd包装器/独立测试子进程，走真实Daemon Snapshot生产路径；过期摘要不落变更租约、双快照持久化、取消恢复与终态race通过（28.946s），make check通过。详见 [快照报告](../../acceptance/reports/firewall-snapshots-2026-09-12.md)。这是命令进程链路，firewalld仍为测试替身，不计真实防火墙环境验收。无活动测试/fixture，未触碰主机规则。下一动作：迁移删除旧systeminfo ApplyFirewall/RollbackFirewall函数及对应旧测试，保留取消/补偿失败的验证含义；再推进Windows系统服务/计划任务真实适配和全范围剩余能力。

2026-09-12 双份预览/提交条件：UI展示区域及运行/永久差异，planHash绑定完整双快照与目标集合；Master要求摘要，Daemon变更前重新采集核对。systeminfo/Daemon定向race1.124s/1.913s、摘要绑定race1.032s、浏览器1/1（7.5s）、check/build/windows、前端生产构建和OpenAPI100路径解析通过。Master预览必填/期限/权限定向TLS也通过。见 [快照证据](../../acceptance/reports/firewall-snapshots-2026-09-12.md)。无活动测试或fixture。下一动作：删除/迁移旧systeminfo Apply/Rollback兼容实现，补Daemon真实命令进程链路与快照冲突拒绝；随后继续Windows系统服务/计划任务等全范围缺口。

2026-09-12 双配置快照增量：新生产防火墙租约固定区域并持久运行/永久快照，分别应用和恢复、读回核对，不使用reload；部分进度可恢复，外部不同规则修改拒绝覆盖。定向race systeminfo1.216s/Daemon2.074s及check/build/windows通过，见 [快照证据](../../acceptance/reports/firewall-snapshots-2026-09-12.md)。旧租约缺少永久快照改为保留诊断，不猜测恢复；最后保护改动的Daemon定向race及重建也通过。下一步：预览显示区域与双份差异、提交条件核对，清理旧兼容Apply/Rollback函数，补Daemon真实命令测试入口；再继续Windows系统适配。无活动测试、fixture或主机变更。

2026-09-12 预览/任务显示增量：端口预览与实际受管集合统一，系统界面按实际任务阶段显示排队/等待确认，确认绑定原节点并恢复草稿；最终定向浏览器1/1（6.3s）、systeminfo全包race1.074s、check/build/windows及前端生产构建通过。构建首次因新增按钮模板实体表达式解析失败，改为computed后退出0。见 [预览证据](../../acceptance/reports/firewall-preview-2026-09-12.md)。下一动作仍是firewalld区域及运行/永久双快照，分别应用与恢复，移除reload对无关运行规则的影响；该后端缺口尚未修复。无活动测试、fixture或主机变更。

2026-09-12 回滚耐久顺序续接：增加rolled_back持久阶段，规则恢复后先记任务终态再清理；已恢复原规则不重复变更，应用错误保留prepared证据。任务更新故障注入验证恢复记录、启动失败与不重放规则；定向race退出0（1.618s），随后追加规则恢复后/回执前中断场景也通过。详见 [防火墙证据](../../acceptance/reports/firewall-confirm-2026-09-12.md)。最终make check/build/windows退出0，无活动测试、fixture或宿主规则变更。下一动作：修复firewalld预览与apply使用不同规则集、运行/永久状态混用和reload影响其他规则的合同；仍须完整实现Windows系统服务/计划任务及其余范围。

2026-09-12 防火墙确认增量：成功终态先持久化、随后删除确认租约；失败保留恢复记录，启动失败不覆盖成普通中断。SQLite故障注入与确认/取消/超时/重启定向race退出0（1.532s），最终make check/build/windows退出0，未变更宿主防火墙。见 [确认恢复证据](../../acceptance/reports/firewall-confirm-2026-09-12.md)。下一具体动作：继续回滚终态的耐久顺序及 firewalld运行/永久规则合同，再补Windows系统服务/计划任务解析；其余SDK受控能力仍保留。没有活动测试或fixture。

2026-09-12 最新检查点：扩展数据转换已移到注册表锁外，重获锁后比对包/数据/提交令牌并重新验证签名信任与依赖。单 Manager 同时一批迁移，冲突保留当前包和最新数据。`go test -race ./internal/extensions ./internal/master -run 'TestMigrationDoesNotBlock|TestUserData|TestRegistry|TestExtensionData|TestIndependentReference' -count=1` 退出0（extensions101.070s、Master36.693s）。迁移 API 仍同步，原全局授权阻塞缺口已修复。

同轮 Windows 原生监控/进程、Linux pidfd 安全终止及完整进程分页已接入，见 [监控证据](../../acceptance/reports/monitoring-2026-09-12.md)。monitor race1.291s、最终真实TLS Master2.246s、浏览器确认/分页/恢复1/1（5.4s）、check/windows/build及前端生产构建通过；Windows测试程序交叉编译通过，真机未运行。首次新增TLS测试响应解码编译错误已修正复验。没有活动测试或fixture。下一动作：继续 Windows 系统服务/计划任务本地化解析、防火墙持久化边界及其余SDK受控能力；全范围仍未完成。

2026-09-12 数据迁移最新检查点：全仓 `GOCACHE=/tmp/blora-go-grant-idem make test` 退出0（Master194.482s、extensions144.750s），随后新增的数据进程中断1.396s及并发/配额2.661s定向race通过。make sdk默认沙箱EPERM后提权重跑通过，干净安装及三个样例包生成完成。真实迁移浏览器1/1（11.7s）、前端33/33、OpenAPI100路径、Linux/Windows构建均已有证据；当前正复建最后校验改动后的二进制，除此无活动测试/fixture。下一步Windows监控实际适配，不能把交叉编译当真机验收。详见 [数据迁移报告](../../acceptance/reports/extensions-data-2026-09-12.md)。

2026-09-12 数据迁移故障证据：编译复用/双用户 WASI 迁移定向 race 138.095s、Master 数据授权 2.166s；实际子进程退出恢复（含数据、提交令牌）1.396s；并发 CAS 与近 4 MiB 配额文件失败保留2.661s通过。报告见 [用户数据](../../acceptance/reports/extensions-data-2026-09-12.md)。全仓 race 会话90976仍运行，extensions已通过144.750s，待Master最终结果。make sdk默认沙箱在esbuild安装验证spawnSync遇EPERM退出2，已提权重跑；不能把首次失败计为通过。fixture均停止。当前下一动作：收取这两个运行会话，更新构建证据，然后补Windows主机监控真实代码（现有!linux文件仍是不可用返回）。

2026-09-12 用户数据迁移进行中：Manager 增加按登录用户隔离的有界 JSON 文档、CAS/稳定请求回执；data.read/data.write SDK 与 API 已接入。升级/回滚/重新安装保留数据时，由目标 WASI 后台逐用户迁移；编译一次、实例内存独立，整批 60 秒。数据与包共同事务恢复，旧事务缺少数据字段时不删现有数据。独立 schema1/2/故意失败3 包已生成；真实 fixture-1932133697 三版本服务器数据与窗口现场迁移/失败保留/回滚 1/1（11.7s）通过并已 Ctrl+C 停止。API 用户隔离/撤权 TLS 0.162s、transport 9/9、前端 33/33、OpenAPI 100 路径、Linux/Windows 构建通过。编译复用后的多用户定向 race 会话 42661 尚在运行，随后已启动全仓 race；收取会话结果后补报告。下一步补实际进程中断包含用户数据的恢复证据、迁移并发/配额边界；不能把早期仅包四文件的证据代替新增数据恢复。

2026-09-12 最新收尾检查点：修正后的独立包升级定向 race 通过（33.921s），原全仓 race 唯一失败项已在相同条件复验通过；原 make test 退出 2 的事实保留，不改写为成功。其余包均在该全仓运行中通过；最终扩展浏览器 5/5（17.2s）、真实独立包组合 1/1（9.6s）、前端 32/32、SDK/包及 Linux/Windows 构建已记录。无活动测试或 fixture。下一条具体动作：扩展按用户隔离的有界 JSON 数据读写、修订/CAS/幂等，升级以受限迁移模块转换数据并与包事务共同提交，失败保留旧包和旧数据；随后补真实两版本迁移。不要将既有 UserDataDir 空目录当成此功能已实现。

2026-09-12 全仓 race 收取：34837 已退出 2，所有其他包通过，Master（177.795s）仅失败于编译时尚未修复的独立包升级版本冲突。修复后普通定向 4.299s 已通过，正在同条件定向 race 复验；不把失败的 make test 写成通过。最终扩展浏览器 5/5（17.2s）通过，fixture 均停止。下一步完成本项 race 后继续按用户隔离的数据读写与升级迁移。

2026-09-12 SDK 桌面真实验证：独立包 0.2.0 的计算/结果/移窗/笔记与任务刷新恢复/快捷入口/关闭组合 1/1（9.6s）通过，fixture-2811626890 已 Ctrl+C 停止。前端 32/32，新增桌面场景连续 5/5；首轮未复现的快捷入口缺失保留记录。升级测试固定版本冲突已改为生成下一主版本，定向复验中。全仓 race 会话 34837 在修复前已启动，需收取最终结果，若只失败旧测试版本则复验该项并更新总状态；下一步继续后端用户数据迁移。见 [桌面 SDK 报告](../../acceptance/reports/extensions-desktop-2026-09-12.md)。

2026-09-12 SDK 桌面能力进行中：新增 window.move/window.close/shortcut.create，沙箱只操作调用者自身标签，移入目标限同应用兼容窗口；快捷入口限当前资源。独立包更新为 0.2.0 并增加按钮，SDK/包、Linux/Windows 与前端构建通过，前端 32/32；新增浏览器首轮快捷入口未出现，尚未确定原因，复跑一次及连续五次通过（4.5s/21.3s）。正在真实 fixture-2811626890（127.0.0.1:9444，会话 54633）验证组合；结束后需 Ctrl+C 停止。尚不能把定向替身浏览器当作真实桌面 SDK 全链路证据。

2026-09-12 终态审计补齐：Master ReconcileTask 与派发前取消现在同事务写入终态审计；storage 全包 race 2.404s、故障注入原子性/重放去重及真实节点集成 4.402s 通过，make check/build 通过。无活动会话。下一条具体动作：SDK 自身标签移窗/关闭、快捷入口能力及沙箱身份限制，之后继续后端用户数据迁移。见 [审计报告](../../acceptance/reports/task-audit-2026-09-12.md)。

2026-09-12 清单签名绑定：修复仅认证 payload 摘要的缺口，改为带版本域的完整 Manifest + payload SHA-256 签名；新增 extension-sign 工具，旧签名要求发布者重签。签名工具/安装/重开/落盘篡改拒绝与清单字段变更的定向 race 通过（1.027s/1.322s），真实 TLS 扩展回归通过（7.559s），make check/windows 通过。OpenAPI 为 99 路径/113 操作。无活动测试/fixture；下一条具体动作：补 Master ReconcileTask 终态审计的事务写入与重放去重证据，然后继续扩展后端用户数据迁移及完整 SDK 桌面能力。见 [签名报告](../../acceptance/reports/extensions-signature-2026-09-12.md)。

2026-09-12 扩展任务闭环：SDK create/read/cancel、应用范围任务授权、执行前/模块获取取消分类、示例任务 ID 恢复已接入。transport 8/8、扩展浏览器 4/4（12.9s）、定向 Master/Daemon race（55.528s/1.331s）通过；真实双节点 fixture 的独立包计算、SDK 结果与刷新不重提交 1/1（8.6s）通过。Linux/前端构建通过，fixture-3028559533 已 Ctrl+C 停止。首次路由冲突已修复并复验；下一步修复包签名未认证 manifest 的缺口，然后继续后端用户数据迁移/完整 SDK 能力。见 [任务报告](../../acceptance/reports/extensions-tasks-2026-09-12.md)。

2026-09-12 注册表事务恢复：安装/升级/回滚/卸载增加四文件旧快照及完整性摘要的持久撤销记录，随机提交令牌区分已提交与待恢复事务，启动前恢复未提交事务，读者拒绝待恢复包；损坏元数据不再被当作未安装覆盖。最终实际子进程退出/提交中断恢复定向 race 通过（1.385s），完整扩展模块此前通过（7.169s），最终 TLS 独立包安装/升级/计算通过（4.453s）。Windows 新增 MoveFileEx 写穿替换/删除及长路径处理；`make check build windows`、前端类型检查和 SDK 构建通过。新增 `.previous` 保留后缀避免 ID 与回滚快照冲突。当前无活动测试/fixture；本轮未重跑全仓 race，物理掉电、ENOSPC、Windows 真机耐久性仍未验证。证据见 [恢复报告](../../acceptance/reports/extensions-recovery-2026-09-12.md)。

2026-09-12 生命周期/迁移增量：Manager 公共读写锁、一致清单/包体 Snapshot、回滚覆盖前校验、禁用状态保持与依赖门禁已接入；定向 race 通过（1.094s），真实 TLS 扩展后台回归通过（10.771s）。AppHost 响应式更新清单/能力/组件，SDK 增加沙箱 `migrate`，宿主对旧现场/资源身份做条件检查并一次提交新状态和版本，失败保留旧现场。前端单元 30/30、扩展浏览器 4/4（20.1s）、最终迁移定向 1/1（6.9s）、类型检查/SDK 和前端生产构建通过；`make check` 通过。当前无活动测试/fixture；本轮未重跑全仓 race，上一轮全仓结果不得替代本轮未覆盖场景。下一步补包写入的掉电事务恢复、后端数据迁移及真实两版本迁移证据。证据见 [生命周期报告](../../acceptance/reports/extensions-lifecycle-2026-09-12.md)。

2026-09-12 WASI 链路收尾：完整独立包已在真实浏览器上传到 HTTPS Master，打开节点资源窗口并执行实际 WASI 笔记计算，1/1 通过（9.4s，`.local/fixture-1553746562`）；fixture 已 Ctrl+C 停止，按精确 fixture 标识检查无残留进程。最终 `GOCACHE=/tmp/blora-go-grant-idem make test` 全仓 race 通过（Master 191.796s，extensions 87.402s），Linux/Windows 二进制及前端生产构建通过，OpenAPI 3.1 解析通过（97 路径）。当前无活动测试/fixture。WASI 获取/编译/执行分别限时 30/30/5 秒，输入输出各 32 KiB；旧输入回传实现已删除。下一步按上方具体动作继续扩展能力和生命周期，不标记全量完成。证据见 [复核报告](../../acceptance/reports/resume-2026-09-12.md)。

2026-09-12 WASI 生产链路：`Bundle` 前后端包格式、Master 版本绑定、Daemon 32 KiB 模块分块获取与沙箱执行已接入，独立示例后台计算笔记统计及摘要。真实 TLS 定向测试已通过（21.585s），扩展管理完整包上传/资源窗口浏览器 3/3 通过（12.3s）。全仓 race 首次因冷编译预算和测试等待不匹配失败，已分离编译/执行期限并加入单模块解码缓存，正在定向 race 复验；不能计作全仓通过。原回传生产代码已删除；下一步完成 race 修复、真实浏览器后台任务结果链路和完整升级/状态迁移。

2026-09-12 资源桥接增量：SDK `readResource()` 与 Master 扩展资源摘要 API 已接入，独立生成包在真实 TLS Master 安装和权限边界测试通过（0.156s）；transport 7/7、浏览器资源窗口与笔记独立 3/3、类型和包构建通过。WASI JSON 后台执行器锁定 wazero v1.12.0（间接升级 x/sys v0.44.0）；真实 Go WASI guest 的计算/文件隔离/空环境/输出上限/循环中止测试通过（11.649s），扩展模块完整回归通过（提权本机 TLS，7.373s），Windows 交叉构建通过，OpenAPI 3.1 解析为 97 路径。执行器尚未接入生产任务，不能替代现有回传实现或计作 E06 通过。当前无活动测试；下一步增加受完整性校验的前后端包格式、Daemon 分块制品获取与任务版本绑定，然后替换回传实现并验证独立包真实后台计算。

2026-09-12 续接修复：授权变更事务幂等、主机任务节点分发与普通用户扩展任务可见性已修复，三项真实 TLS 集成测试通过（1.482s）；证据见 [续接报告](../../acceptance/reports/resume-2026-09-12.md)。独立参考扩展新增 esbuild 0.25.12 浏览器打包及真实 DOM 界面，`npm run package` 已通过；读取实际生成包的浏览器加载/立即刷新及原扩展场景 3/3 通过（12.4s，后端路由替身），相关 Go 包 vet 通过。无活动测试会话；下一步用独立包接入真实安装测试并补资源桥接、后台受限执行和迁移，不能用此局部证据结束 B09。

| 日期 | 决定 | 原因/影响 |
| --- | --- | --- |
| 2026-09-09 | 两份 v0.3 规划完整保留，另加全范围执行解释 | 用户要求把完整规划放入项目并供后续模型实施，不删改已确认产品行为 |
| 2026-09-09 | M0～M3 作为同一任务实施顺序 | 用户希望一次持续任务覆盖明确功能，之后再调细节 |
| 2026-09-09 | 真实实现与真实验证分开计数 | 缺少 Windows/Docker 等环境不能用桩函数或虚假“通过”掩盖 |
| 2026-09-09 | 本次只准备文档 | 用户本轮要求放置规划与行动指导，尚未要求运行完整开发或部署 |
| 2026-09-09 | 工作位置更正为 /data/instances/blora-panel | 按用户更正同步文档路径；新目录作为后续开发入口 |
| 2026-09-09 | 用户明确执行完整启动提示，开始产品实现 | 保留此前准备记录作为历史；全范围持续任务已启动 |
| 2026-09-09 | 按AGENTS条件分配desktop/runtime/protocol模块 | 当前技能列表无匹配的通用编码模块技能；主代理控制公共合同、依赖及整体验收 |

追加技术决定时注明影响、替代方案和关联验收，不覆盖历史记录。

## 8. 交接规则

后续模型先读当前记录再看必要源码和失败证据，继续未完成内容，不重新初始化已有工程，不要求用户重复已经确认的需求。

如果平台强制中断，交接必须能回答：做到哪里、真实通过什么、什么失败、当前代码可否运行、下一步执行什么。不能只留下“后续完善”。如果全范围尚有缺口，最终回复如实列出；不能为了结束任务把缺口改写成视觉细节优化。
