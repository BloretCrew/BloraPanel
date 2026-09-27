export type IconOptics={id:string;name:string;description:string;tint:number;bevel:number;refraction:number;reflection:number;roughness:number;curve:number;smoke:number;structure?:'clear'|'absorption'|'core'|'rim'|'laminate'|'soft'|'balanced'|'lens';shell?:number;paper?:number;depth?:number}
export const iconOptics:IconOptics[]=[
 {id:'g01',name:'01 · 清透薄片',description:'窄反射边，清晰透色；轻薄的彩色玻璃。',tint:.72,bevel:1.25,refraction:1.8,reflection:.52,roughness:48,curve:0,smoke:0},
 {id:'g02',name:'02 · 弧面折射',description:'稍宽的弧面边缘，纸张在边缘轻微弯折。',tint:.57,bevel:2.8,refraction:4.4,reflection:.64,roughness:42,curve:0,smoke:0},
 {id:'g03',name:'03 · 柔面反射',description:'反射更柔和，保持颜色与图形的分量。',tint:.77,bevel:1.9,refraction:2.5,reflection:.38,roughness:19,curve:0,smoke:0},
 {id:'g04',name:'04 · 高透玻璃',description:'降低色片遮挡，反射边缘更清晰可见。',tint:.37,bevel:2.15,refraction:4.7,reflection:.85,roughness:62,curve:0,smoke:0},
 {id:'g05',name:'05 · 烟色玻璃',description:'较深的有色玻璃，突出明亮的局部反射。',tint:.74,bevel:1.8,refraction:3.1,reflection:.75,roughness:55,curve:0,smoke:.2},
 {id:'g06',name:'06 · 清晰色片',description:'已移除旧版斜向环境映光，仅保留边缘效果。',tint:.69,bevel:1.7,refraction:2.8,reflection:.5,roughness:38,curve:0,smoke:0},
]
export const glassMaterials:IconOptics[]=[
 {id:'h01',name:'01 · 清透色片',description:'清晰的透叠关系，保持薄片中央平整。',structure:'clear',tint:.44,bevel:1.2,refraction:1.25,reflection:.2,roughness:48,curve:0,smoke:0,paper:.74},
 {id:'h02',name:'02 · 有色玻璃',description:'用玻璃的吸色表现厚度，保留更饱满的颜色。',structure:'absorption',tint:.82,bevel:2.3,refraction:2,reflection:.28,roughness:50,curve:0,smoke:0,paper:.78},
 {id:'h03',name:'03 · 边缘透镜',description:'通透的中心，边缘对下层内容产生折射。',structure:'lens',tint:.38,bevel:3.4,refraction:5.8,reflection:.34,roughness:58,curve:0,smoke:0,paper:.72},
 {id:'h04',name:'04 · 透明包边',description:'彩色主体外保留可见的透明薄壳。',structure:'core',tint:.83,bevel:1.65,refraction:2.4,reflection:.4,roughness:58,curve:0,smoke:0,shell:2.8,paper:.85},
 {id:'h05',name:'05 · 彩边清玻',description:'色彩集中在边缘，中央允许更多下层颜色透入。',structure:'rim',tint:.34,bevel:2.9,refraction:3.1,reflection:.3,roughness:55,curve:0,smoke:0,shell:3.5,paper:.72},
 {id:'h06',name:'06 · 夹层玻璃',description:'透明表层与色片分离，用轻微错位表现层次。',structure:'laminate',tint:.58,bevel:1.5,refraction:1.8,reflection:.28,roughness:55,curve:0,smoke:0,shell:1.7,paper:.8,depth:1.4},
 {id:'h07',name:'07 · 雾白透色',description:'少量乳白材质混合清晰透色，边缘保持轻薄。',structure:'soft',tint:.48,bevel:1.5,refraction:1.4,reflection:.17,roughness:30,curve:0,smoke:0,paper:.72},
 {id:'h08',name:'08 · 平衡薄壳',description:'保留应用色，结合清晰透叠与适量壳层厚度。',structure:'balanced',tint:.69,bevel:2,refraction:2.6,reflection:.3,roughness:52,curve:0,smoke:0,shell:1.5,paper:.84},
]
export const opticsById=(id:string|null)=>[...iconOptics,...glassMaterials].find(m=>m.id===id)
