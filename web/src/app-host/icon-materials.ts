export type IconMaterial={id:string;name:string;description:string;surface:number;tint:number;blur:number;diffuse:number;ink:number;light:number;edge:number;separation:number;backdrop:number;saturation:number;dark?:boolean;pearl?:boolean;flat?:boolean}
export const iconMaterials:IconMaterial[]=[
 {id:'m01',name:'01 · 原色柔雾',description:'原版配色，增加一层很轻的磨砂透色。',surface:.7,tint:.07,blur:10,diffuse:.1,ink:.97,light:.12,edge:.35,separation:.025,backdrop:6,saturation:1},
 {id:'m02',name:'02 · 清玻璃',description:'减少底板实色，让桌面颜色从玻璃后透入。',surface:.22,tint:.06,blur:11,diffuse:.15,ink:.96,light:.16,edge:.72,separation:.04,backdrop:9,saturation:1.02},
 {id:'m03',name:'03 · 有色磨砂',description:'雾面底层保留更多应用色，轮廓保持清晰。',surface:.48,tint:.22,blur:13,diffuse:.18,ink:.96,light:.18,edge:.45,separation:.035,backdrop:7,saturation:1.16},
 {id:'m04',name:'04 · 轻薄叠层',description:'正面色片轻微透叠，只在相邻层之间分离。',surface:.55,tint:.04,blur:5,diffuse:.08,ink:.88,light:.08,edge:.2,separation:.1,backdrop:3,saturation:1.08},
 {id:'m05',name:'05 · 缎面白',description:'接近白色的细腻底面，主体颜色更集中。',surface:.96,tint:0,blur:16,diffuse:.05,ink:1,light:.13,edge:.45,separation:.02,backdrop:2,saturation:1.08},
 {id:'m06',name:'06 · 乳白漫射',description:'更柔和的乳白材质，底层色彩在内部扩散。',surface:.68,tint:.12,blur:18,diffuse:.3,ink:.86,light:.18,edge:.6,separation:0,backdrop:12,saturation:1},
 {id:'m07',name:'07 · 清晰色片',description:'主体用原版平涂，只有承托层使用磨砂。',surface:.46,tint:.13,blur:10,diffuse:.12,ink:1,light:0,edge:.35,separation:.015,backdrop:8,saturation:1.2,flat:true},
 {id:'m08',name:'08 · 冷暖微光',description:'原色上增加很淡的冷暖环境透光。',surface:.52,tint:.09,blur:15,diffuse:.16,ink:.96,light:.14,edge:.52,separation:.025,backdrop:9,saturation:1,pearl:true},
 {id:'m09',name:'09 · 烟色磨砂',description:'深色半透明底，更清楚地观察透色与层次。',surface:.82,tint:.09,blur:11,diffuse:.22,ink:.98,light:.24,edge:.28,separation:.035,backdrop:9,saturation:1.05,dark:true},
 {id:'m10',name:'10 · 薄烟玻璃',description:'浅烟色玻璃，保留更多桌面背景的参与。',surface:.35,tint:.1,blur:12,diffuse:.15,ink:.98,light:.2,edge:.4,separation:.02,backdrop:12,saturation:1.08,dark:true},
 {id:'m11',name:'11 · 半透明主体',description:'将透感延伸到图形内部，减少实体块感。',surface:.62,tint:.05,blur:6,diffuse:.18,ink:.76,light:.2,edge:.44,separation:.075,backdrop:7,saturation:1.12},
 {id:'m12',name:'12 · 平衡磨砂',description:'主体保持分量，柔雾、透色与边缘都更克制。',surface:.6,tint:.08,blur:12,diffuse:.14,ink:.97,light:.1,edge:.4,separation:.04,backdrop:8,saturation:1.06},
]
export const originalMaterial:IconMaterial={id:'original',name:'00 · 原版对照',description:'最初的 Material 图标，保留原始造型和配色。',surface:1,tint:0,blur:0,diffuse:0,ink:1,light:0,edge:0,separation:0,backdrop:0,saturation:1,flat:true}
export const materialById=(id:string|null)=>iconMaterials.find(material=>material.id===id)
export const classicBackgrounds:Record<string,string>={launcher:'#e6e9e8',instances:'#e1e7ec',backups:'#dfece6',files:'#ede5d8',editor:'#e9e3e8',terminal:'#e0e7e9',tasks:'#ece8df',nodes:'#e2e6ed',monitor:'#dfebe7',extensions:'#ede3dd',users:'#e8e2ea',settings:'#e1e7ec',docker:'#dfe7ec',system:'#e9e7e0'}
export const classicAccents:Record<string,string>={launcher:'#779cad',instances:'#527faa',backups:'#63aa95',files:'#d9ac61',editor:'#a17da6',terminal:'#6186a6',tasks:'#bbaa77',nodes:'#899bd0',monitor:'#64aa95',extensions:'#ce8a7a',users:'#a180b1',settings:'#8ca1bb',docker:'#729dcc',system:'#b9ad8d'}
export function blend(a:string,b:string,amount:number){const ca=a.slice(1).match(/../g)!.map(x=>parseInt(x,16)),cb=b.slice(1).match(/../g)!.map(x=>parseInt(x,16));return '#'+ca.map((v,i)=>Math.round(v*(1-amount)+cb[i]!*amount).toString(16).padStart(2,'0')).join('')}
