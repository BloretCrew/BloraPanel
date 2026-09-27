import {createApp,defineComponent,h} from 'vue'
import IconStudy from './app-host/IconStudy.vue'
import './icon-study.css'
const apps=[['launcher','启动器'],['instances','实例中心'],['backups','备份与计划'],['files','文件管理器'],['editor','编辑器'],['terminal','终端'],['tasks','任务中心'],['nodes','节点管理'],['monitor','监控与进程'],['docker','容器中心'],['users','用户与权限'],['settings','设置'],['extensions','应用市场'],['system','系统管理']]
const variants=[['a','A / 精简双色','清晰几何 · 圆形遮罩 · 无投影'],['b','B / 轻透层叠','浅色连续曲面 · 柔和色阶 · 轻微层次'],['c','C / 深色珐琅','深石墨底色 · 彩色主体 · 克制对比']]
createApp(defineComponent({setup:()=>()=>h('main',[
  h('header',[h('div',[h('small','BLORA / ICON STUDIES'),h('h1','更简洁，也更清晰。')]),h('p','三个方向 · 同一组应用 · 实际矢量组件')]),
  ...variants.map(([variant,title,subtitle])=>h('section',{id:variant},[
    h('div',{class:'caption'},[h('h2',title),h('p',subtitle)]),
    h('div',{class:'icons'},apps.map(([appId,title])=>h('div',{class:'app'},[h(IconStudy,{appId:appId!,variant}),h('span',title)]))),
  ])),h('footer','2026.09 / 设计候选，尚未设为默认图标。桌面与窗口布局保持一致。'),
])})).mount('#study')
