import {createApp,defineComponent,h} from 'vue'
import NativeAppIcon from './app-host/NativeAppIcon.vue'
import {explorationApps} from './app-host/icon-explorations'
import './native-icons.css'
const heroes=[['files','文件管理器','文件夹本身就是图标，纸张与前板形成轻微层次。'],['editor','编辑器','纸张、正文和金属笔尖组成独立轮廓。'],['terminal','终端','完整的深色面板，提示符与光标成为视觉中心。'],['extensions','应用市场','蓝色安装袋与软件包，以材质区分前后层。']]
const icon=(appId:string)=>h(NativeAppIcon,{appId})
createApp(defineComponent({setup:()=>()=>h('main',[
  h('header',[h('small','BLORA / DESKTOP APPLICATION ICONS'),h('h1','回到应用图标。'),h('p','独立造型 · 完整构图 · 克制的材质层次')]),
  h('section',{class:'heroes'},heroes.map(([id,name,description])=>h('article',[icon(id!),h('h2',name),h('p',description)]))),
  h('section',{class:'collection'},[h('h2','整套应用'),h('div',{class:'app-grid'},explorationApps.map(([id,name])=>h('div',{class:'app-item'},[icon(id),h('span',name)])))]),
  h('section',{class:'dock-compare'},[h('h2','Dock 尺寸'),h('div',{class:'dock'},['launcher','instances','files','editor','terminal','extensions','monitor','settings','tasks'].map(id=>h('div',icon(id))))]),
  h('footer','SVG / CSS 实际浏览器渲染 · 未使用图片生成资产 · 设计评审预览'),
])})).mount('#native-icons')
