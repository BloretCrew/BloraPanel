import {createApp,defineComponent,h,ref} from 'vue'
import OpticalAppIcon from './app-host/OpticalAppIcon.vue'
import ForegroundAppIcon from './app-host/ForegroundAppIcon.vue'
import {iconOptics,opticsById} from './app-host/icon-optics'
import {explorationApps} from './app-host/icon-explorations'
import './icon-foreground.css'
import './icon-optics.css'
const params=new URLSearchParams(location.search),selected=opticsById(params.get('variant'))
const group=Number(params.get('group'))||1
const sampleApps=['instances','files','editor','terminal','extensions','settings']
const name=(id:string)=>explorationApps.find(a=>a[0]===id)?.[1]||id
createApp(defineComponent({setup(){
 const bare=ref(params.has('bare')),light=ref(0),backdrop=ref('light')
 const lights:[[number,number],[number,number],[number,number]]=[[-.65,-.5],[0,0],[.75,.5]]
 const icon=(appId:string,id:string,isolated=bare.value,lightIndex=light.value)=>id==='l03'?h(ForegroundAppIcon,{appId,material:id,bare:isolated}):h(OpticalAppIcon,{appId,material:id,bare:isolated,lightX:lights[lightIndex]![0],lightY:lights[lightIndex]![1]})
 const icons=(id:string,apps=sampleApps,isolated=bare.value)=>h('div',{class:'icon-row'},apps.map(appId=>h('div',{class:'icon-cell'},[h('div',{class:'glyph'},icon(appId,id,isolated)),h('span',name(appId))])))
 const row=(m:{id:string;name:string;description:string})=>h('section',{id:m.id,class:'study-row'},[h('div',{class:'row-label'},[h('h2',m.name),h('p',m.description),m.id==='l03'?null:h('a',{href:`/icon-optics.html?variant=${m.id}`},'放大 ↗')]),icons(m.id)])
 return()=>h('main',{class:['optics-review',{detail:!!selected},'backdrop-'+backdrop.value]},[
  h('header',[h('div',[h('small','BLORA / OPTICAL MATERIAL · 07'),h('h1',selected?selected.name:'让反射沿着图形走。'),h('p','主体材质对照 · 同一图标造型与底板 · 移动指针可改变反射')]),h('div',{class:'optical-controls'},[
   h('div',{class:'light-controls'},['左侧光','正面光','右侧光'].map((label,i)=>h('button',{'aria-pressed':light.value===i,onClick:()=>light.value=i},label))),
   h('button',{onClick:()=>bare.value=!bare.value},bare.value?'显示底板':'仅看主体'),
   h('button',{onClick:()=>backdrop.value=backdrop.value==='light'?'dark':'light'},backdrop.value==='light'?'深底对比':'浅底对比'),
  ])]),
  h('div',{class:'comparisons'},[row({id:'l03',name:'上一版 · 无光影',description:'模糊透叠，作为本轮对照。'}),...(selected?[selected]:iconOptics.slice((group-1)*3,group*3)).map(row)]),
  selected?h('div',[
   h('section',{class:'angle-comparison'},[h('h2','同一材质 · 三个光照位置'),h('div',{class:'angle-grid'},lights.map((_,i)=>h('div',[h('span',['左侧光','正面光','右侧光'][i]),h('div',{class:'angle-icons'},['instances','files','terminal'].map(appId=>h('div',icon(appId,selected.id,false,i))))])))]),
   h('section',{class:'complete'},[h('h2','完整应用图标'),icons(selected.id,explorationApps.map(a=>a[0]))]),
   h('section',{class:'size-check'},[h('h2','实际尺寸'),...[64,48,32].map(size=>h('div',{class:'size-row'},[h('span',size+' px'),...sampleApps.map(appId=>h('div',{style:{width:size+'px',height:size+'px'}},icon(appId,selected.id,false)))]))]),
  ]):null,
  h('footer',[h('span','代码渲染的反射 / 边缘折射候选；未替换默认图标。'),h('nav',[h('a',{href:'/icon-optics.html?group=1'},'01—03'),h('a',{href:'/icon-optics.html?group=2'},'04—06')])]),
 ])
}})).mount('#optics')
