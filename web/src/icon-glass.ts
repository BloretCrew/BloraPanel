import {createApp,defineComponent,h,ref} from 'vue'
import OpticalAppIcon from './app-host/OpticalAppIcon.vue'
import ForegroundAppIcon from './app-host/ForegroundAppIcon.vue'
import {glassMaterials} from './app-host/icon-optics'
import {explorationApps} from './app-host/icon-explorations'
import './icon-foreground.css'
import './icon-glass.css'
const params=new URLSearchParams(location.search)
const group=Number(params.get('group'))||1
const selected=glassMaterials.find(m=>m.id===params.get('variant'))
const samples=['instances','files','terminal','editor','extensions','nodes','settings']
const name=(id:string)=>explorationApps.find(a=>a[0]===id)?.[1]||id
createApp(defineComponent({setup(){
 const bare=ref(params.has('bare')),dark=ref(false)
 const icon=(appId:string,id:string,isolated=bare.value,probe:'plain'|'grid'|'colour'='plain')=>id==='l03'?h(ForegroundAppIcon,{appId,material:id,bare:isolated}):h(OpticalAppIcon,{appId,material:id,bare:isolated,lightX:-.6,lightY:-.5,probe})
 const icons=(id:string,apps=samples,isolated=bare.value)=>h('div',{class:'icon-row'},apps.map(appId=>h('div',{class:'icon-cell'},[h('div',{class:'glyph'},icon(appId,id,isolated)),h('span',name(appId))])))
 const row=(m:{id:string;name:string;description:string})=>h('section',{class:'study-row',id:m.id},[h('div',{class:'row-label'},[h('h2',m.name),h('p',m.description),m.id==='l03'?null:h('a',{href:`/icon-glass.html?variant=${m.id}`},'放大查看 ↗')]),icons(m.id)])
 const probes=(id:string)=>h('section',{class:'probes'},[h('h2','材质观察 · 三种衬底'),h('p','网格和彩色块仅用于观察透色与边缘折射，正常桌面不使用这些衬底。'),h('div',{class:'probe-grid'},(['plain','grid','colour'] as const).map((probe,i)=>h('div',[h('span',['原底板','网格 · 观察边缘弯折','色块 · 观察下层透色'][i]),h('div',{class:'probe-icons'},['instances','files','terminal'].map(appId=>h('div',icon(appId,id,false,probe))))])))])
 return()=>h('main',{class:['glass-review',{detail:!!selected,'dark-review':dark.value}]},[
  h('header',[h('div',[h('small','BLORA / GLASS MATERIAL · 08'),h('h1',selected?selected.name:'只留下玻璃的通透。'),h('p','相同造型与底板 · 比较透色、折射和壳层 · 已移除斜向反光带')]),h('div',{class:'glass-controls'},[h('button',{onClick:()=>bare.value=!bare.value},bare.value?'显示底板':'仅看主体'),h('button',{onClick:()=>dark.value=!dark.value},dark.value?'浅底对比':'深底对比')])]),
  h('div',{class:'comparisons'},[row({id:'l03',name:'对照 · 原透叠',description:'上一轮无光影版本，作为对照。'}),...(selected?[selected]:glassMaterials.slice((group-1)*4,group*4)).map(row)]),
  selected?h('div',[probes(selected.id),h('section',{class:'complete'},[h('h2','完整应用图标'),icons(selected.id,explorationApps.map(a=>a[0]))]),h('section',{class:'size-check'},[h('h2','实际尺寸'),...[64,48,32].map(size=>h('div',{class:'size-row'},[h('span',size+' px'),...samples.map(appId=>h('div',{style:{width:size+'px',height:size+'px'}},icon(appId,selected.id,false)))]))])]):null,
  h('footer',[h('span','默认采用 04「透明包边」并保留底板；其他候选仅供开发评审。'),h('nav',[h('a',{href:'/icon-glass.html?group=1'},'01—04'),h('a',{href:'/icon-glass.html?group=2'},'05—08')])]),
 ])
}})).mount('#glass')
