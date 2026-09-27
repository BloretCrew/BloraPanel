<script setup lang="ts">
import AppIcon from '../app-host/AppIcon.vue'
import UIIcon from '../app-host/UIIcon.vue'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { apps } from '../app-host/registry'
import UnavailableApp from '../app-host/UnavailableApp.vue'
import { useDesktop } from './store'
import {DESKTOP_TOP} from './work-area'
import type { Rect } from '../app-host/types'
const props=defineProps<{windowId:string;zIndex:number}>(), desktop=useDesktop()
const win=computed(()=>desktop.state!.windows[props.windowId]!), app=computed(()=>apps.get(win.value.appId)||{manifest:{title:win.value.appId,icon:'◇',color:'#aab8ad'},component:UnavailableApp}), active=computed(()=>desktop.state!.activeWindowId===props.windowId), tabMenu=ref<string>(), preview=ref<'left'|'right'|'max'>(), moving=ref(false), frame=ref<HTMLElement>(), gestureError=ref('')
const title=computed(()=>desktop.state!.views[win.value.activeTabId]?.title || app.value.manifest.title)
let abort:AbortController|undefined,tabAbort:AbortController|undefined,lastFocus:HTMLElement|undefined
watch(()=>active.value && !win.value?.minimized,async focused=>{
  if(!focused)return
  await nextTick()
  if(frame.value?.contains(document.activeElement))return
  if(lastFocus?.isConnected)lastFocus.focus({preventScroll:true})
  else (frame.value?.querySelector<HTMLElement>('[data-active-view] [role="textbox"],[data-active-view] input') || frame.value)?.focus({preventScroll:true})
})
const directions=['n','ne','e','se','s','sw','w','nw']
function gesture(event:PointerEvent,direction?:string){
  if(event.button!==0 || (event.target as HTMLElement).closest('button,input'))return
  event.preventDefault();desktop.focus(props.windowId)
  const target=event.currentTarget as HTMLElement;target.setPointerCapture(event.pointerId)
  const original={...win.value.rect},originalSnap=win.value.snap
  if(win.value.snap)desktop.snap(props.windowId)
  const start={x:event.clientX,y:event.clientY,rect:{...win.value.rect}}
  moving.value=true;abort?.abort();abort=new AbortController();let animation=0,pending:Rect|undefined
  function move(e:PointerEvent){const dx=e.clientX-start.x,dy=e.clientY-start.y,r={...start.rect}
    if(!direction){r.x=Math.max(-r.width+120,Math.min(innerWidth-120,r.x+dx));r.y=Math.max(DESKTOP_TOP,Math.min(innerHeight-90,r.y+dy));preview.value=e.clientY<DESKTOP_TOP+10?'max':e.clientX<18?'left':e.clientX>innerWidth-18?'right':undefined}
    else{if(direction.includes('e'))r.width=Math.max(360,r.width+dx);if(direction.includes('s'))r.height=Math.max(250,r.height+dy);if(direction.includes('w')){r.width=Math.max(360,r.width-dx);r.x=start.rect.x+start.rect.width-r.width}if(direction.includes('n')){r.height=Math.max(250,r.height-dy);r.y=start.rect.y+start.rect.height-r.height}}
    pending=r;if(!animation)animation=requestAnimationFrame(()=>{animation=0;if(pending)win.value.rect=pending})
  }
  function end(){if(animation)cancelAnimationFrame(animation);if(pending)desktop.geometry(props.windowId,pending);if(preview.value)desktop.snap(props.windowId,preview.value);preview.value=undefined;moving.value=false;if(target.hasPointerCapture(event.pointerId))target.releasePointerCapture(event.pointerId);abort?.abort()}
  target.addEventListener('pointermove',move,{signal:abort.signal});target.addEventListener('pointerup',end,{signal:abort.signal});target.addEventListener('pointercancel',end,{signal:abort.signal})
  // A reload during a gesture preserves the current visible geometry synchronously.
  window.addEventListener('beforeunload',()=>{if(pending)desktop.geometry(props.windowId,pending)},{signal:abort.signal})
  window.addEventListener('keydown',e=>{if(e.key==='Escape'){preview.value=undefined;pending=original;end();if(originalSnap)desktop.snap(props.windowId,originalSnap)}},{signal:abort.signal})
}
function dragTab(event:PointerEvent,viewTabId:string){
  if(event.button!==0 || (event.target as HTMLElement).closest('button'))return
  const start={x:event.clientX,y:event.clientY},target=event.currentTarget as HTMLElement;let dragged=false
  tabAbort?.abort();tabAbort=new AbortController()
  target.setPointerCapture(event.pointerId)
  const move=(e:PointerEvent)=>{if(Math.abs(e.clientX-start.x)+Math.abs(e.clientY-start.y)>12)dragged=true}
  const cancel=()=>{if(target.hasPointerCapture(event.pointerId))target.releasePointerCapture(event.pointerId);tabAbort?.abort()}
  const up=(e:PointerEvent)=>{cancel();if(!dragged){desktop.activate(props.windowId,viewTabId);return}try{const destination=document.elementFromPoint(e.clientX,e.clientY)?.closest<HTMLElement>('[data-tab-strip]');if(destination){const windowId=destination.dataset.tabStrip!;const tabs=[...destination.querySelectorAll<HTMLElement>('[data-view-tab]')];const index=tabs.findIndex(el=>e.clientX<el.getBoundingClientRect().left+el.clientWidth/2);desktop.moveView(viewTabId,windowId,index<0?undefined:index)}else desktop.moveView(viewTabId,undefined,undefined,{x:Math.max(0,e.clientX-80),y:Math.max(DESKTOP_TOP,e.clientY-20)})}catch(e){gestureError.value=String(e)}}
  target.addEventListener('pointermove',move,{signal:tabAbort.signal});target.addEventListener('pointerup',up,{signal:tabAbort.signal});target.addEventListener('pointercancel',cancel,{signal:tabAbort.signal})
  window.addEventListener('keydown',e=>{if(e.key==='Escape')cancel()},{signal:tabAbort.signal})
}
onBeforeUnmount(()=>{abort?.abort();tabAbort?.abort()})
</script>
<template>
  <div v-if="preview" class="snap-preview" :class="preview" :style="{zIndex:zIndex+1}"></div>
  <section v-show="!win.minimized" ref="frame" class="app-window" :class="{focused:active,moving,snapped:win.snap}" :data-window-id="windowId" :data-window-mode="win.mode" role="region" tabindex="-1" :aria-label="title" :style="{transform:`translate(${win.rect.x}px,${win.rect.y}px)`,width:win.rect.width+'px',height:win.rect.height+'px',zIndex}" @pointerdown.capture="!active && desktop.focus(windowId)" @focusin="lastFocus=$event.target as HTMLElement" @keydown.esc="tabMenu=undefined;gestureError=''">
    <header class="window-titlebar" @pointerdown="gesture($event)" @dblclick.self="desktop.snap(windowId,win.snap?undefined:'max')"><div class="window-heading"><span class="window-app-icon"><AppIcon :app-id="win.appId" /></span><span class="window-title">{{title}}</span></div><div class="window-controls"><button class="window-minimize" aria-label="最小化窗口" title="最小化" @click="desktop.minimize(windowId)"><UIIcon name="minus"/></button><button class="window-maximize" aria-label="最大化或还原窗口" title="最大化 / 还原" @click="desktop.snap(windowId,win.snap?undefined:'max')"><UIIcon name="expand"/></button><button class="window-close" aria-label="关闭窗口" title="关闭视图" @click="desktop.closeWindow(windowId)"><UIIcon name="close"/></button></div></header>
    <nav class="view-tabs" role="tablist" :data-tab-strip="windowId"><div v-for="viewId in win.tabs" :key="viewId" class="view-tab" :class="{selected:win.activeTabId===viewId}" role="tab" tabindex="0" :aria-selected="win.activeTabId===viewId" :data-view-tab="viewId" @pointerdown="dragTab($event,viewId)" @keydown.enter="desktop.activate(windowId,viewId)" @keydown.delete="desktop.closeView(windowId,viewId)" @contextmenu.prevent="tabMenu=tabMenu===viewId?undefined:viewId"><span>{{desktop.state!.views[viewId]?.pinned?'• ':''}}{{desktop.state!.views[viewId]?.title}}</span><button :aria-label="'标签菜单 '+desktop.state!.views[viewId]?.title" @click="tabMenu=tabMenu===viewId?undefined:viewId"><UIIcon name="down"/></button><button aria-label="关闭标签" @click="desktop.closeView(windowId,viewId)"><UIIcon name="close"/></button></div><button class="new-tab" aria-label="新建标签" @click="desktop.open({appId:win.appId,windowId,disposition:'new-tab'})"><UIIcon name="plus"/></button></nav>
    <div v-if="tabMenu" class="tab-menu"><button @click="desktop.moveView(tabMenu);tabMenu=undefined">移到新窗口</button><button v-for="target in Object.values(desktop.state!.windows).filter(w=>w.windowId!==windowId && w.appId===win.appId)" :key="target.windowId" @click="desktop.moveView(tabMenu!,target.windowId);tabMenu=undefined">移入 {{desktop.state!.views[target.activeTabId]?.title}}</button><button @click="desktop.open({appId:win.appId,resourceRef:desktop.state!.views[tabMenu]?.resourceRef,title:desktop.state!.views[tabMenu]?.title,state:desktop.state!.views[tabMenu]?.state,disposition:'new-window'});tabMenu=undefined">复制视图到新窗口</button><button @click="desktop.commit([{kind:'set',path:['views',tabMenu,'pinned'],value:!desktop.state!.views[tabMenu]?.pinned}]);tabMenu=undefined">固定 / 取消固定</button><button @click="tabMenu=undefined">关闭菜单</button></div>
    <p v-if="gestureError" class="error notice" role="alert">{{gestureError}}</p>
    <div class="window-body"><div v-for="viewId in win.tabs" v-show="win.activeTabId===viewId" :key="viewId" class="app-view" :data-active-view="win.activeTabId===viewId?viewId:undefined"><component :is="app.component" :view-tab-id="viewId" :window-id="windowId" :visible="win.activeTabId===viewId && !win.minimized" /></div></div>
    <div v-for="direction in directions" :key="direction" class="resize-handle" :class="direction" :data-resize="direction" @pointerdown="gesture($event,direction)"></div>
  </section>
</template>
