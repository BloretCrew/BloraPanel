import {createApp,defineComponent,h} from 'vue'
import IconExploration from './app-host/IconExploration.vue'
import {explorations,explorationApps,isExploration,type ExplorationId} from './app-host/icon-explorations'
import './icon-explorations.css'
const requested=new URLSearchParams(location.search).get('variant')
const selected=isExploration(requested)?explorations.find(item=>item.id===requested):undefined
const icon=(appId:string,variant:ExplorationId)=>h(IconExploration,{appId,variant})
const fullGrid=(variant:ExplorationId)=>h('div',{class:'icon-grid'},explorationApps.map(([id,label])=>h('div',{class:'icon-cell'},[icon(id,variant),h('span',label)])))
const sizeRow=(variant:ExplorationId,size:number)=>h('div',{class:'size-row'},[h('span',`${size} px`),...explorationApps.map(([id])=>h('div',{style:{width:`${size}px`,height:`${size}px`}},icon(id,variant)))])
createApp(defineComponent({setup:()=>()=>h('main',{class:selected?'detail':'overview'},[
  h('header',[h('div',[h('small','BLORA / ICON EXPLORATIONS 02'),h('h1',selected?selected.name:'九个方向，重新比较。')]),h('p',selected?selected.description:'重新审视形状、线条、层次与用色 · 所有方案均无渐变')]),
  selected?h('article',[
    h('p',{class:'detail-note'},selected.note),
    fullGrid(selected.id),
    h('section',{class:'scale-panel'},[h('h2','实际尺寸'),sizeRow(selected.id,48),sizeRow(selected.id,32),sizeRow(selected.id,20)]),
    h('section',{class:'contrast-panel'},[h('h2','深色背景'),h('div',{class:'on-dark'},explorationApps.map(([id])=>h('div',icon(id,selected.id))))]),
  ]):h('div',{class:'directions'},explorations.map(item=>h('section',{id:item.id,class:'direction'},[
    h('div',{class:'direction-heading'},[h('b',item.id.toUpperCase()),h('h2',item.name)]),
    h('p',item.description),fullGrid(item.id),h('footer',item.note),
  ]))),
  h('footer',{class:'page-footer'},'2026.09 / 原生 SVG 与 Lucide 组件 · 设计评审候选 · 桌面布局未变'),
])})).mount('#explorations')
