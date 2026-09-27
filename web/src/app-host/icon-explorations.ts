export const explorationApps = [
  ['launcher','启动器'],['instances','实例中心'],['backups','备份计划'],['files','文件管理'],
  ['editor','编辑器'],['terminal','终端'],['tasks','任务中心'],['nodes','节点管理'],
  ['monitor','监控'],['docker','容器中心'],['users','用户权限'],['settings','设置'],
  ['extensions','应用市场'],['system','系统管理'],
] as const
export const explorations = [
  {id:'d',name:'瓷白 / Porcelain',family:'line',description:'双色细线 + 白色底板，减少视觉重量。',note:'细化线宽与留白，偏精密工具感。'},
  {id:'e',name:'色域 / Colorfield',family:'solid',description:'整面纯色 + 镂空主体，以色彩强化辨识。',note:'提高小尺寸辨识度，不叠加渐变。'},
  {id:'f',name:'开放 / Open',family:'solid',description:'移除统一底板，保留独立的彩色轮廓。',note:'减少重复方块，让图形本身成为图标。'},
  {id:'g',name:'叠纸 / Fold',family:'ribbon',description:'几何切面 + 两层色块，强调构图。',note:'用真实遮挡表达层次，不使用投影。'},
  {id:'h',name:'夜航 / Instrument',family:'line',description:'深石墨面板 + 彩色线条，偏专业控制台。',note:'修正上一轮深色方案的过亮白色。'},
  {id:'i',name:'交织 / Interlace',family:'ribbon',description:'独立几何标记，双色穿插、无外框。',note:'更抽象的品牌感，作为大胆方向比较。'},
  {id:'j',name:'清透 / Clear',family:'line',description:'透明浅底 + 细边界，仅一层局部色片。',note:'撤掉 B 的渐变，让背景参与材质。'},
  {id:'k',name:'索引 / Index',family:'line',description:'彩色侧边 + 深色符号，类似工具分类标签。',note:'控制用色面积，同时保留分类辨识。'},
  {id:'l',name:'墨色 / Ink',family:'solid',description:'墨色主体 + 彩色切面，收敛整体饱和度。',note:'以不对称底形区别于常规圆角方块。'},
] as const
export type ExplorationId = typeof explorations[number]['id']
export const isExploration=(id:string|null):id is ExplorationId=>explorations.some(item=>item.id===id)
