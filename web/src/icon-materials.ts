import {createApp,defineComponent,h,ref} from 'vue'
import MaterialAppIcon from './app-host/MaterialAppIcon.vue'
import {iconMaterials,originalMaterial,materialById,type IconMaterial} from './app-host/icon-materials'
import {explorationApps} from './app-host/icon-explorations'
import './icon-materials.css'
const selected=materialById(new URLSearchParams(location.search).get('variant'))
const group=Number(new URLSearchParams(location.search).get('group'))
const shown=group>=1&&group<=3?iconMaterials.slice((group-1)*4,group*4):iconMaterials
const materialIcon=(appId:string,id:string)=>h(MaterialAppIcon,{appId,material:id})
const apps=(id:string,className='')=>h('div',{class:['material-icons',className]},explorationApps.map(([appId,title])=>h('div',{class:'material-cell'},[materialIcon(appId,id),h('span',title)])))
const sampleApps=['instances','files','editor','terminal','extensions','monitor','settings']
const row=(id:string,size:number)=>h('div',{class:'size-row'},[h('span',`${size}px`),...sampleApps.map(appId=>h('div',{style:{width:`${size}px`,height:`${size}px`}},materialIcon(appId,id)))])
createApp(defineComponent({setup(){
  const backdrop=ref('wallpaper')
  return()=>h('main',{class:[selected?'detail':'overview',group?'group-view':'',`backdrop-${backdrop.value}`]},[
    h('header',[h('small','BLORA / MATERIAL STUDIES 04'),h('div',{class:'heading'},[h('h1',selected?selected.name:'同一套图标，只比较材质。'),h('div',{class:'backdrop-controls'},['wallpaper','light','dark'].map((value,index)=>h('button',{onClick:()=>backdrop.value=value,'aria-pressed':backdrop.value===value},['壁纸','浅底','深底'][index])))]),h('p',selected?selected.description:'保留原版轮廓、构图与色彩关系 · 磨砂 / 透色 / 高斯漫射 / 轻叠层')]),
    selected?h('div',[
      h('section',{class:'detail-comparison'},[
        h('div',[h('h2','00 · 原版'),h('div',{class:'hero-row'},sampleApps.slice(0,5).map(appId=>h('div',materialIcon(appId,'original'))))]),
        h('div',[h('h2',selected.name),h('div',{class:'hero-row'},sampleApps.slice(0,5).map(appId=>h('div',materialIcon(appId,selected.id))))]),
      ]),
      h('section',{class:'full-set'},[h('h2','整套应用'),apps(selected.id)]),
      h('section',{class:'small-sizes'},[h('h2','实际尺寸'),row(selected.id,48),row(selected.id,32),row(selected.id,20)]),
      h('section',{class:'background-check'},[h('h2','深浅背景对照'),h('div',{class:'background-rows'},['light','dark'].map(mode=>h('div',{class:mode},sampleApps.map(appId=>h('div',materialIcon(appId,selected.id))))))]),
    ]):h('div',[
      h('section',{class:'baseline'},[h('h2',originalMaterial.name),apps('original','baseline-icons')]),
      h('div',{class:'material-grid'},shown.map(m=>h('section',{id:m.id,class:'material-card'},[
        h('div',{class:'material-title'},[h('h2',m.name),h('a',{href:`/icon-materials.html?variant=${m.id}`},'放大 ↗')]),h('p',m.description),apps(m.id),
      ]))),
    ]),h('footer','SVG / CSS 实际渲染 · 高斯模糊仅用于材质底层，主体保持清晰 · 当前为候选预览'),
  ])
}})).mount('#materials')
