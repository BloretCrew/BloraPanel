import {createApp,defineComponent,h,ref} from 'vue'
import ForegroundAppIcon from './app-host/ForegroundAppIcon.vue'
import {foregroundLightingMaterials} from './app-host/icon-foreground-materials'
import {explorationApps} from './app-host/icon-explorations'
import './icon-foreground.css'
import './icon-lighting.css'
const params=new URLSearchParams(location.search)
const selected=foregroundLightingMaterials.find(m=>m.id===params.get('variant'))
const focused=params.has('focus')
const samples=focused?['instances','files','terminal']:['instances','files','editor','terminal','extensions','monitor','settings']
const baseline={id:'f06',name:'上一版 · 对照',description:'面内渐变、光斑与亮暗边缘叠加。'}
const name=(id:string)=>explorationApps.find(a=>a[0]===id)?.[1]||id
const icon=(appId:string,material:string,bare=false)=>h(ForegroundAppIcon,{appId,material,bare})
const icons=(id:string,bare:boolean,apps=samples)=>h('div',{class:'icon-row'},apps.map(appId=>h('div',{class:'icon-cell'},[h('div',{class:'glyph'},icon(appId,id,bare)),h('span',name(appId))])))
const row=(m:{id:string;name:string;description:string},bare:boolean)=>h('section',{class:'study-row',id:m.id},[h('div',{class:'row-label'},[h('h2',m.name),h('p',m.description)]),icons(m.id,bare)])
createApp(defineComponent({setup(){
 const bare=ref(params.has('bare'))
 return()=>h('main',{class:['lighting-review',{focus:focused,detail:!!selected}]},[
  h('header',[h('div',[h('small','BLORA / LIGHTING REVIEW · 06'),h('h1',selected?selected.name:'保留通透感，收掉夸张光影。'),h('p','同一造型、配色、透明度与模糊强度 · 仅比较光影处理')]),h('button',{onClick:()=>bare.value=!bare.value},bare.value?'显示底板':'仅看主体')]),
  h('div',{class:'comparisons'},[row(baseline,bare.value),...(selected?[selected]:foregroundLightingMaterials).map(m=>row(m,bare.value))]),
  selected?h('div',[
   h('section',{class:'isolated'},[h('h2','主体放大'),icons(selected.id,true)]),
   h('section',{class:'complete'},[h('h2','完整应用图标'),icons(selected.id,bare.value,explorationApps.map(a=>a[0]))]),
   h('section',{class:'size-check'},[h('h2','实际尺寸'),...[48,32,20].map(size=>h('div',{class:'size-row'},[h('span',size+' px'),...samples.map(appId=>h('div',{style:{width:size+'px',height:size+'px'}},icon(appId,selected.id)))]))]),
  ]):null,
  h('footer','B 版保留层间模糊带来的色彩过渡；没有人为添加的方向光、明暗渐变、光斑或亮边。'),
 ])
}})).mount('#lighting')
