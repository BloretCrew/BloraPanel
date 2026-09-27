import {createApp,defineComponent,h,ref} from 'vue'
import ForegroundAppIcon from './app-host/ForegroundAppIcon.vue'
import MaterialAppIcon from './app-host/MaterialAppIcon.vue'
import ClassicIconArtwork from './app-host/ClassicIconArtwork.vue'
import {foregroundMaterials,foregroundById} from './app-host/icon-foreground-materials'
import {explorationApps} from './app-host/icon-explorations'
import './icon-foreground.css'
const params=new URLSearchParams(location.search)
const selected=foregroundById(params.get('variant'))
const group=Number(params.get('group'))||1
const shown=foregroundMaterials.slice((group-1)*3,group*3)
const sampleApps=['instances','files','editor','terminal','extensions','monitor','settings']
const name=(id:string)=>explorationApps.find(a=>a[0]===id)?.[1]||id
const icon=(appId:string,id:string,bare=false)=>id==='original'?(bare?h('svg',{viewBox:'0 0 96 96',fill:'none',class:'original-bare'},[h(ClassicIconArtwork,{kind:appId})]):h(MaterialAppIcon,{appId,material:'original'})):h(ForegroundAppIcon,{appId,material:id,bare})
const icons=(id:string,bare:boolean,apps=sampleApps)=>apps.map(appId=>h('div',{class:'icon-cell'},[h('div',{class:'glyph'},[icon(appId,id,bare)]),h('span',name(appId))]))
createApp(defineComponent({setup(){
 const bare=ref(params.has('bare'))
 return()=>h('main',{class:{detail:!!selected,bare:bare.value}},[
  h('header',[h('div',[h('small','BLORA / FOREGROUND MATERIAL · 05'),h('h1',selected?selected.name:'材质，做到图形里面。'),h('p','原版造型 · 相同底板 · 单独处理面板、纸张、笔身和店铺的透色与漫射')]),h('button',{onClick:()=>bare.value=!bare.value},bare.value?'显示底板':'仅看主体')]),
  selected?h('div',{class:'detail-content'},[
   h('section',{class:'compare'},['original',selected.id].map(id=>h('div',{class:'comparison-row'},[h('h2',id==='original'?'原版 · 实色主体':selected.name),h('div',{class:'icon-row'},icons(id,bare.value,sampleApps.slice(0,5)))]))),
   h('section',{class:'isolated'},[h('h2','去掉底板 · 观察主体'),h('div',{class:'icon-row'},icons(selected.id,true))]),
   h('section',{class:'complete'},[h('h2','完整应用图标'),h('div',{class:'icon-row'},icons(selected.id,bare.value,explorationApps.map(a=>a[0])))]),
   h('section',{class:'size-check'},[h('h2','实际尺寸'),...[48,32,20].map(size=>h('div',{class:'size-row'},[h('span',size+' px'),...sampleApps.map(appId=>h('div',{style:{width:size+'px',height:size+'px'}},[icon(appId,selected.id)]))]))]),
  ]):h('div',{class:'comparisons'},[
   h('section',{class:'study-row original'},[h('div',{class:'row-label'},[h('h2','00 · 原版'),h('p','保留原始平涂，作为对照。')]),h('div',{class:'icon-row'},icons('original',bare.value))]),
   ...shown.map(m=>h('section',{class:'study-row',id:m.id},[h('div',{class:'row-label'},[h('h2',m.name),h('p',m.description),h('a',{href:`/icon-foreground.html?variant=${m.id}`},'放大查看 ↗')]),h('div',{class:'icon-row'},icons(m.id,bare.value))])),
  ]),
  h('footer',[h('span',bare.value?'底板已隐藏；当前看到的是图形自身的材质。':'所有版本使用完全相同的不透明底板。'),h('nav',[h('a',{href:'/icon-foreground.html?group=1'},'01—03'),h('a',{href:'/icon-foreground.html?group=2'},'04—06')])]),
 ])
}})).mount('#foreground')
