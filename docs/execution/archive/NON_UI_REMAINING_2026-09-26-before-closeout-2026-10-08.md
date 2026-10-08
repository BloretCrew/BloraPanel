# Historical snapshot before the non-release closeout (2026-10-08)

This preserves the full prior record. Relative Markdown links were adjusted for this archive directory. Current status is in the parent document; pending items below are historical.

# UI 暂停后的剩余工作核对

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

最新产品并行独立终端解析版（源e1102342/产物3a360b78）正式Firefox90秒bounded **51ms FAIL**，2720事件、五目的校验；62584退出1/23258清理0。111单位、构建、Firefox7原生恢复/确认边界/遮挡像素通过。E08/R1仍开放，不能用旧短测50ms关闭；WebKit关键功能及最新三引擎/持续长测待齐。R2/R3/R4闭合、Windows不重复、发行延期。

最新同源长测：正式不可变终端日志Firefox300秒bounded **53ms FAIL**，8520事件（首120/58、steady8400/52）、23次目的校验，61233退出1/41824清理0；源b98487b1/产物32a54525。短轮90秒50ms不足以关闭E08/R1，继续检查原生绘制成本。111单位/构建/Firefox六项产品守卫通过；最新三引擎及严格功能尚未全部通过。R2/R3/R4闭合、Windows不重复、R5发行延期。以下运行中记录过期。

当前非发行剩余仍为E08/R1：正式不可变终端日志候选Firefox90秒bounded **50ms PASS**，2820事件、首120保留、八次目的校验及真实原负载，浏览器退出0/fixture清理0。111单位、构建、Firefox六项原生产品回归通过；539源b98487b1/168产物32a54525。同源Firefox300秒bounded正在执行，最终三引擎与严格功能尚未全部通过，不提前关闭R1。未采用的阴影sibling原型及严格失败证据保留为独立实验入口。R2/R3/R4闭合、Windows不重复，R5发行延期。以下历史“当前/运行中”不能作为最新状态。

2026-10-08最新有效状态：正式dedup产品Firefox90秒bounded p95 **55ms FAIL**，2645事件、首120/62ms、7次目的校验；28581退出1/97515夹具清理0。当前源016ec7cf/产物b206021c。恢复和终端守卫Firefox/WebKit各6/6通过，但E08/R1仍开放。内容空白阴影sibling原型严格RGBA比较失败，未进入产品；下一进行独立Firefox原生线程剖析，原负载/50ms/首120/恢复/ACK合同保持。当前无真实压力或夹具。R2/R3/R4闭合、Windows不重复，R5发行延期。以下历史运行状态已过期。

2026-10-08最新：仅E08/R1继续；正式RGB产品300秒p95 51ms未达标，旧90秒诊断PASS不能关闭目标。当前完整基线+有界终端增量IDB候选已通过108单位、Firefox/WebKit各5原生恢复/终端守卫及Chromium两原生恢复；同源正式Firefox90秒正在补验，0847环境中断无报告不计结果。软件WebGL DPR1.25原像素守卫失败继续保留，未降低阈值。源3324bac7/产物38a52390，包含用户最新要求的下拉框斜角/文字裁切修复（三浏览器×深浅色实际截图/键盘选择通过）。三引擎原60秒及持续稳定性仍待齐，不导入旧版本PASS。R2/R3/R4闭合、Windows不重复、R5发行延期。[最新证据](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-08 最新有效状态：仅R1 E08继续，R2/R3/R4关闭，Windows不重复，R5发行延期。0715可见行委托54ms FAIL后已撤回，不再是当前产品。0746五种壁纸光学材质RGB24诊断首次完整90秒50ms PASS，但不是最终成品验收。当前产品只加入无损不透明壁纸RGB编码，透明源保留原PNG；0807构建及102单位通过，源6e8eeb6b/产物5ead3af9。六配色深浅/DPR光学24次RGBA0通过，最终功能守卫与同源三引擎原60秒、持续长测待齐。50ms/首120事件/原负载/原生正文与字形/恢复与ACK都保持，E08不能提前关闭。[证据](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-08当前状态覆盖下方历史“最新/运行中”：alpha-two55ms、partial viewport54ms均FAIL。现在只推进E08/R1原生可见行结构candidate：原生DOM renderer保留字体/字形/布局/输入/解析/恢复/ACK，仅减少隐藏行生成；首次露出同步恢复已提交native行，2026保持未释放图像。Firefox原3+新2和WebKit5/5严格终端回归、102单位及最终构建通过。当前538源385d7a9a/168产物63d310d8，fresh正式Firefox90秒bounded0715进行中，不能称达标。最终同源三引擎/完整原60与持续稳定性仍待验证；R2/R3/R4已关闭，Windows不重跑，R5发行延期。没有其他新增任务范围。[证据](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-08 E08最新：原60秒三引擎30.9/49/49ms PASS和Chromium600秒31.1ms PASS是历史不同源证据，Firefox300秒bounded53ms FAIL；新native motion driver90秒也54ms FAIL（initial71/steady52、2695样本、八次目的校验），没有证明收益的直接driver已撤回。当前源381c7964/产物9a6a3414仍待最终完整验证，不混旧PASS。下一结构候选native-shadow-alpha按原生阴影非零alpha拆分且保留完整原图采样，Firefox/WebKit各2/2整屏RGBA0，正在真实90秒诊断而非成品验收。严格50ms、8窗/双PTY/万目录/持续目标校验、恢复和原零正文/字形像素差异保持。仅R1非发行工作开放，Windows不重复；本段覆盖下方R1历史进展。[当前证据](../../acceptance/reports/e08-render-architecture-2026-10-06.md)。

2026-10-06 最新结论：R4普通Windows真机真实联调已于固定e546c8f0 **14/14通过**，报告完整性和八阶段exit0核验通过，见[报告](../../acceptance/reports/windows-device-passed-2026-10-06.md)。R2、R3和R4均已关闭；**当前非发行剩余只有R1 E08严格延迟**。R5发行仍由用户延期，无需再运行任何已关闭Windows补测组。以下历史R4待执行措辞不再代表当前任务。

更新：2026-10-05，按用户最新要求由三个子代理只读审查并收缩难提供环境的必测范围，见[范围调整与审查报告](../../acceptance/reports/scope-review-2026-10-05.md)。保留原路径供已有链接使用；本页“尚未完成”的 R1～R5 是当前阻塞清单，下面历史报告中的旧补测指令不再代表下一动作。已验证的发行基线为 RC34，后续源码修复尚需最终发行收尾，不把历史51/51冒充最新全平台验收。

物理掉电、缓存丢失、宿主/NTFS实际磁盘耗尽、浏览器真实配额阈值/驱逐、独立公网Engine/注册源/证书运维、多平台全组合及超过既有24h的长测，改为本次不强制的环境兼容性项。Windows管理员服务/任务实际变更、通知中心/系统IME/原生对话框、额外headless/自然周期/冷启动/指针/绘制与容器组合也不要求逐项提供环境。已有安全ENOSPC、SIGKILL、权限和恢复测试保留；审查不记作对应实测PASS。Windows防火墙仅支持读取状态，规则应用/回滚不是当前Windows支持子集。

最新第十六轮：[固定8d6b14e报告](../../acceptance/reports/windows-sixteenth-report-2026-10-05.md)确认Windows实际headed剩余取消导航用例3/3，正式构建及全部阶段PASS，完整checksum/ZIP/来源及阶段检查点匹配，原45/5秒和恢复断言保留。结合[第十五轮](../../acceptance/reports/windows-fifteenth-report-2026-10-05.md)刷新/queued/日志/Worker终端各3/3，用户要求的visible WebKit补测组结束，无需再跑同组脚本。两份报告按原版本独立记录，不导入旧PASS或伪称新版单批15/15，旧14/15 FAILED保留。未测环境仍不是PASS，但是否阻塞本次交付以本页最新范围为准；E08和普通Windows真实联调/发行验收仍保留。

## 历史增量（保留结果，不作为当前补测要求）

第十四轮 Windows 增量：[固定63cacc0的报告](../../acceptance/reports/windows-fourteenth-report-2026-10-04.md)确认17/17摘要及完整ZIP匹配，构建PASS、OS端口3570 fresh preview成功，WebKit12/15。无帧/真实queued/日志各3/3；剩filter恢复5秒、held read准备迟到、初次导航慢后接管45秒耗尽，Windows原因未定。脚本新增可选-Headed同五项3次环境对照和逐执行实际模式证据，原45/5秒及既有断言保持。初次Linux headed12/15及两次Worker探针0/1保留，发现reactive controls跨Worker克隆失败；现controls/links转为纯数据，真实刷新后验证Worker保持，UI/协议/ACK/不重放不变。新类型构建、90单测、脚本harness/实际JSON/finally通过；Linux headed/headless精确各15/15，WebKit六文件终端恢复62/62、Firefox/Chromium正式preview各7/7，无skip/retry/flaky/全局错误且引擎依次。回传仍保留FAILED，不用本地修复假称Windows三处超时已解决。下一次固定新提交运行`-Followup -BrowsersOnly -WebKitOnly -Headed`，保持桌面解锁且允许自动窗口运行，正式构建PASS及五项3次15/15回传；可见模式结果不自动关闭默认无界面可靠性，完整原生/E08/外部缺口保持。

第十三轮 Windows 增量：[固定c63bcb4的报告](../../acceptance/reports/windows-thirteenth-report-2026-10-04.md)确认14/14摘要及完整ZIP集合匹配，正式构建24.04秒PASS；5173端口占用使测试启动失败，五项各3次均未执行。原监听归属未知，不杀/不复用；补测脚本现每个引擎启动前选择OS空闲loopback端口，同步测试/server地址，strictPort/fresh保持，报告新增有界全局错误和实际URL/端口。原任务恢复缺口未因零执行关闭，产品/UI/用例/45及5秒不改。本地端口占用/抢占失败关闭/6种非法配置及PS7 harness/类型通过；保持5173占用时WebKit精确15/15、97.036秒，Firefox正式preview7/7、66.612秒及Chromium开发server7/7、78.294秒通过。真实JSON/finally保留回传零执行FAILED及启动错误，1,145文档链接/差异检查通过，所属容器和会话结束。下一次仍固定新版`-Followup -BrowsersOnly -WebKitOnly`，正式构建PASS且五项各3次严格15/15回传，未选择项不导入旧PASS。

第十二轮 Windows 增量：[固定90658d0的报告](../../acceptance/reports/windows-twelfth-report-2026-10-04.md)确认17/17摘要和完整ZIP集合匹配，正式构建PASS、WebKit12/15；终端/日志/无帧恢复各3/3。剩任务窗口准备耗尽45秒、首列表未就绪及导航后原5秒计数失败，后者具体Windows原因未确立。测试现原生Enter打开可见可用任务应用，确认首列表HTTP/body和筛选控件；queued数据只在浏览器实际departure后切换，防止旧普通轮询提前满足恢复断言。诊断补真实JSON消费/DOM计数/浏览器时间的数字及静态事件白名单，非绘制验证。产品/UI不改，原45/5秒/ACK/不重放/错误断言保持；最终WebKit21/21、Firefox7/7、Chromium7/7及冷视图三次真实轮询后释放1/1、类型构建/harness/实际JSON和诊断复制通过，严格15次CLI选择仅核验、不计另一次运行；故意失败0/1诊断探针独立保留。交付固定新版`-Followup -BrowsersOnly -WebKitOnly`，正式构建且同五项各3次严格15/15；Windows实效/完整原生/E08/外部缺口保持，未选择Go/SDK/其他Windows引擎不导入旧PASS。第十一轮结果和修订保留于[历史报告](../../acceptance/reports/windows-eleventh-report-2026-10-04.md)。

第十轮 Windows 增量：[固定3ef289a的报告](../../acceptance/reports/windows-tenth-report-2026-10-04.md)确认正式构建通过、WebKit12/15，终端3次通过；剩两次控制台点击稳定等待45秒超时（尚无日志读取），及一次无帧汇总5秒断言失败（之后有HTTP完成）。Windows具体绘制/数据发布原因仍未确立。安全读取恢复后现立即重查活动任务查询，保留确认缓存，不重放写入；日志生命周期测试以原生Enter键打开同一实际UI，明确不计这三步指针稳定覆盖。所有查询/日志补实时阶段和有界数字计数/按钮几何诊断。90单测/类型构建、WebKit36/36、Firefox12/12、Chromium12/12及脚本harness/真实JSON/严格15次选择核验通过；探索暂停时钟未复现指针停顿，独立保留，不冒充修复因果。下一次仍`-Followup -BrowsersOnly -WebKitOnly`五项各3次且构建PASS、15/15，不重复Go/GCC/SDK/其他引擎，不导入旧PASS；UI冻结，Windows实效和原生/E08/外部缺口保持。

第九轮 Windows 增量：[固定3e6c31c的报告](../../acceptance/reports/windows-ninth-report-2026-10-03.md)确认WebKit14/15，四项导航/日志用例各3次通过，前轮两项超时已获真机通过。仅第一次终端耗尽原45秒总时限，之后两次约19秒通过，无页面/请求错误；首次刷新约32秒才开始，不能从上传确认具体慢动作/Windows因果。补测改为先单独构建正式frontend，再在fresh preview运行，新增实时阶段耗时和有界静态资源诊断，保留原断言/时限，UI不改。下一次仍是`-Followup -BrowsersOnly -WebKitOnly`五项各3次，必须构建通过且15/15；不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS。本地验证详情见报告，Windows实效和完整原生/E08/外部环境缺口保持。

第八轮 Windows 增量：[固定57d4fb6的报告](../../acceptance/reports/windows-eighth-report-2026-10-03.md)确认WebKit13/15，原终端配置3次通过，剩两次读取恢复/测试建立超时。可控无帧基线0/1复现汇总恢复超时，API只读门禁现增加250ms独立恢复兜底，pagehide阻止旧帧/旧截止恢复，新导航重置代次；日志测试用真实interval回调建立待取消读取，queued观察不依赖绘制，原限时/错误/ACK/输入不重放保持。89单测、类型构建及本地WebKit36/36、Firefox12/12、Chromium19/19加独立补验4/4、脚本harness/真实JSON和精确15次选择通过。下一次仍仅`-Followup -BrowsersOnly -WebKitOnly`五项各3次，不重复Go/GCC/SDK/其他浏览器，不计未选择项PASS；Windows实效、原生/E08/外部环境缺口保持。

第七轮 Windows 增量：[固定8ad386a的报告](../../acceptance/reports/windows-seventh-report-2026-10-03.md)确认WebKit11/12，三项任务查询导航各3次通过；剩一次fallback/worker终端的错误来自独立控制台日志轮询。可控基线0/1复现页面离开后新日志fetch，安全API读取现统一取消/等待，仅导航取消的只读请求可在存活页面恢复，body取消不能变成空成功，控制台销毁取消所属读取；输入/写入不重放。87单测、类型构建及本地Chromium23/23、Firefox12/12、WebKit36/36通过，最终PowerShell harness/真实JSON解析和精确选择核验通过。下一次仅`-Followup -BrowsersOnly -WebKitOnly`，五项各3次严格15/15；不重复已通过Go/GCC/SDK/其他浏览器，不计未选择项PASS；修复实效、完整平台/E08及外部环境缺口仍待验证。

第六轮 Windows 增量：[固定f98b078的报告](../../acceptance/reports/windows-sixth-report-2026-10-01.md)确认Chromium/Firefox各8/8、WebKit23/24，两项导航查询检查也各3次通过；只剩一次fallback/worker终端的任务列表请求在离开窗口期失败。可控排队轮询基线0/1，新增读取门禁后组合本地Chromium/Firefox各9/9、WebKit27/27，原错误断言保持。下一次仅运行 `-Followup -BrowsersOnly -WebKitOnly`，剩余终端配置加三项导航各3次（严格12/12）；不重复Go、GCC、SDK及其他浏览器，不把本轮未选择项当作新PASS。Windows修复实效仍待实际回传，完整平台/E08和外部环境缺口保留。

第五轮 Windows 增量：[固定3c7abb3的报告](../../acceptance/reports/windows-fifth-report-2026-10-01.md)确认11必需原生、3项Go回归各3次通过，Master/runlog race87pass/0fail/4skip；前轮两处Go竞争及账号恢复已获得真机通过证据。Chromium/Firefox各6/6，WebKit16/18，剩两次任务汇总请求诊断。任务读取与共享列表已补离开前取消与AbortSignal；最终本地Chromium/Firefox各8/8、WebKit24/24及80单测/类型检查/构建通过，Windows实效仍待验证。下一次仅运行 `-Followup -BrowsersOnly` 的8/8/24浏览器补测，不再重复Go/GCC/SDK；不代表Windows全范围通过。

第四轮 Windows 增量：[固定08db099的报告](../../acceptance/reports/windows-fourth-report-2026-10-01.md)已确认27定向Go/11原生项全部通过、Chromium/Firefox各32/32，原来的目录metadata和ConPTY已有真机PASS。全race两项新竞争已在本地修正并通过受影响模块完整race；浏览器夹具修正后本地Chromium/Firefox各6/6、WebKit18/18通过。新修复仍待 -Followup 真机复测，不代表Windows全范围通过。

UI、完整通透材质及 p95 ≤ 50ms 的目标保持不变。证据只覆盖实际验证范围；用户本次排除的特殊环境不再阻塞交付，不宣称原全环境认证通过。

## 已完成的本地增量

| 项目 | 当前证据及范围 |
| --- | --- |
| 功能回归与发行 | RC34真实完整功能51/51、897.040秒通过；RC33首轮49/51两处扩展点击失败保留，测试补既有原生hover就绪前置条件，业务断言不变。RC34六包2,251条记录/三份Web、独立启动/停机恢复/兼容RC33回退通过，见[本轮报告](../../acceptance/reports/terminal-generation-2026-09-29.md)。Windows包不等于Windows运行通过 |
| 24小时稳定性 | [真实长测](../../acceptance/reports/local-endurance-24h-2026-09-27.md)86,415.480秒、17,262次采样、1,440分钟槽、两次故障恢复与末尾备份恢复通过；不代替更长期或跨主机/跨平台组合 |
| 跨引擎存储故障 | [Firefox/WebKit](../../acceptance/reports/recovery-engines-2026-09-26.md)各3项通过；不代替设备掉电或实际浏览器配额阈值 |
| 终端恢复 | RC27～33补齐UTF-8续接、未完成VT序列、控制状态、鼠标/光标协议、颜色/主题顺序和OSC 8链接。RC33单测76/76、Chromium61/61、Firefox/WebKit各50/50；不声称所有VT扩展已验证 |
| Linux系统管理与限制 | [独立systemd/cgroup](../../acceptance/reports/systemd-cgroup-2026-09-29.md)服务/定时器、归属/整组结束、CPU限流、32MiB OOM与8进程上限实际触发通过；停止服务和未加载timer遗漏已修复 |
| Linux原生通知 | [X11/DBus/Dunst](../../acceptance/reports/native-notifications-2026-09-29.md)实际绘制/点击/刷新去重/退出后遗留弹窗关闭通过；不覆盖其他桌面或退出后后台投递 |
| R2：防火墙能力门禁 | 状态读取与仅Linux firewalld支持的规则变更已分开；类型检查及定向浏览器5/5通过，见[本地报告](../../acceptance/reports/local-remaining-2026-10-05.md) |
| R3：普通Windows真实联调入口 | 新脚本使用真实Master/Daemon/HTTPS/WSS及Vue/Monaco，临时数据与确认清理、独立报告；当前Linux最新源码实跑14/14、安全单测2/2、Windows交叉构建和PowerShell报告harness通过。Windows实际执行不计PASS，见[入口说明](../../operations/WINDOWS_VALIDATION.md) |

## 尚未完成

| 项目 | 当前事实 | 下一步动作 |
| --- | --- | --- |
| R1：E08严格延迟 | **本轮已按用户最新范围收尾**。同源原始首120：Chromium38.8ms达到原50ms；Firefox66ms、WebKit132ms/95ms未达原50ms。较早同源WebKit89ms，因此仍有89–132ms波动，不能保证稳定≤100ms。最后已有flat-body候选独立复测139ms无稳定收益，未进入产品；生产构建、121单位、21项WebKit守卫及最终类型检查通过。新增Chromium DPR1静态截图不稳定仍是未解决诊断，不计通过；无本任务服务运行。见[收尾报告](../../acceptance/reports/e08-bounded-closeout-2026-10-08.md)。 | 用户2026-10-08接受极端负载下约100ms，要求少量尝试后收尾；已完成一次最后候选复测及同源三浏览器/一次WebKit独立复核。严格全浏览器50ms保留为将来改进，不再阻塞本轮；不继续自动优化，不重复Windows测试，发行按原要求延期。 |
| R5：最终发行（用户延期） | RC34历史发行与51/51保留，后续修复不自动获得同一发行/全范围结论；用户明确要求本轮不做发行 | 发行打包、兼容回退和发布留到用户要求时执行；本轮仍整理当前源码的普通回归和验收证据 |

## 非阻塞历史观察与环境限制

Vim刷新偶发额外字节和[RC30一次撤权超时](../../acceptance/reports/terminal-protocol-state-2026-09-29.md)保留原失败与未定位根因。随后原断言多次通过，三子代理终端代次/回放/撤权审查未确认新缺陷，因此不再作为需要无限补测的交付阻塞；若出现新失败附件则重新调查，不能称已确定根因修复。

物理设备、实际宿主磁盘、公网多主机、所有通知/IME/桌面组合及更长期运行未实测的事实继续保留在[范围报告](../../acceptance/reports/scope-review-2026-10-05.md)，状态为本次不强制/部署兼容性待抽验，不是PASS。

## 继续执行条件

本轮已复现并修复旧终端连接解析失败晚返回禁用新连接的问题，三引擎定向组合各4/4通过；不能将它认定为原Vim偶发或撤权超时根因。完整回归session90325已退出0，无活动测试句柄，日志`.local/evidence/rc34/real-full.log`。玻璃遮挡裁剪的像素等价检查未通过，不纳入产品。旧表中的RC23发行、24h仍RUNNING及Linux systemd/通知完全缺环境已不再是当前待办。

本轮R2修复及R3新真实入口已完成本地验证，相关证据见[本轮报告](../../acceptance/reports/local-remaining-2026-10-05.md)。R4已由普通Windows实际14/14报告关闭，见[真机报告](../../acceptance/reports/windows-device-passed-2026-10-06.md)。R1新增布局隔离探针截图一致但仍约923～953ms，收益不足，未采用，见[性能诊断](../../acceptance/reports/render-layout-diagnostic-2026-10-05.md)；不降低50ms或改变UI。下一具体工作是R1有针对性的延迟诊断优化；R5用户明确延期。无活动测试/服务/容器和会话接管。无需危险环境，不重复已关闭补测组，不将Linux或静态审查冒充Windows实跑，不称本次交付全部完成。
