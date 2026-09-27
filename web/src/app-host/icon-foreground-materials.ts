export type ForegroundMaterial = {
  id:string;name:string;description:string;
  opacity:number;paper:number;blur:number;light:number;shade:number;rim:number;diffusion:number;chroma:number;cool:number;
  edgeMode?:'directional'|'uniform';edgeWidth?:number;
}
export const foregroundMaterials:ForegroundMaterial[] = [
  {id:'f01',name:'01 · 细磨砂',description:'有色主体里柔化光线，纸张透过文件夹前板。',opacity:.76,paper:.86,blur:2.2,light:.24,shade:.06,rim:.35,diffusion:.3,chroma:1.12,cool:0},
  {id:'f02',name:'02 · 清透色片',description:'更薄的面板、更清楚的层间透色，边缘轻薄。',opacity:.55,paper:.74,blur:.85,light:.28,shade:.08,rim:.58,diffusion:.22,chroma:1.38,cool:0},
  {id:'f03',name:'03 · 柔光雾面',description:'主体内部漫射更柔和，弱化硬反光。',opacity:.68,paper:.8,blur:4.2,light:.34,shade:.03,rim:.18,diffusion:.42,chroma:1.18,cool:0},
  {id:'f04',name:'04 · 克制缎面',description:'保留原版色彩的分量，只增加很浅的厚度与透叠。',opacity:.89,paper:.92,blur:1.6,light:.17,shade:.045,rim:.26,diffusion:.2,chroma:1.04,cool:0},
  {id:'f05',name:'05 · 冷调透光',description:'冷色环境光进入主体，彩色层与雾白层交叠。',opacity:.64,paper:.78,blur:2.8,light:.29,shade:.075,rim:.38,diffusion:.36,chroma:1.2,cool:.11},
  {id:'f06',name:'06 · 明快叠色',description:'保留饱满应用色，让磨砂主体更清楚、更有辨识度。',opacity:.78,paper:.87,blur:1.9,light:.23,shade:.075,rim:.42,diffusion:.3,chroma:1.45,cool:0},
]
// Isolate lighting from the accepted frosted/translucent layer composition.
// Every candidate below keeps f06's geometry, colour, blur and opacity unchanged.
const lightingBase=foregroundMaterials.find(m=>m.id==='f06')!
export const foregroundLightingMaterials:ForegroundMaterial[] = [
  {...lightingBase,id:'l01',name:'A1 · 微弱面光',description:'去掉光斑与底部压暗，仅保留极轻的上沿明度变化。',light:.065,shade:0,diffusion:0,rim:.12,edgeMode:'uniform',edgeWidth:.7},
  {...lightingBase,id:'l02',name:'A2 · 仅薄边缘',description:'主体使用平色，只留一圈很薄的透光边缘。',light:0,shade:0,diffusion:0,rim:.28,edgeMode:'uniform',edgeWidth:.7},
  {...lightingBase,id:'l03',name:'B · 无光影',description:'移除全部明暗渐变、光斑和亮边，仅保留模糊透叠。',light:0,shade:0,diffusion:0,rim:0},
]
export const foregroundById=(id:string|null)=>[...foregroundMaterials,...foregroundLightingMaterials].find(m=>m.id===id)
