<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useEventListener, useWindowSize, usePreferredReducedMotion } from '@vueuse/core'
import { useQuery } from '@tanstack/vue-query'
import { motion } from 'motion-v'
import { apps } from '../app-host/registry'
import { session, api, APIError, type Instance } from '../services/api'
import { useDesktop } from './store'
import { json } from '../recovery/state'
import WindowFrame from './WindowFrame.vue'
import AppIcon from '../app-host/AppIcon.vue'
import UIIcon from '../app-host/UIIcon.vue'
import DesktopBar from './DesktopBar.vue'
import DockSurface from './DockSurface.vue'
import {DESKTOP_TOP} from './work-area'
import {themePalette} from '../appearance/palette'
import {wallpaperStyle} from '../appearance/wallpaper-style'
import type { Shortcut } from '../app-host/types'
defineEmits<{logout:[]}>()
const desktop=useDesktop(), launcher=ref(false), menu=ref<{x:number;y:number;shortcut?:Shortcut;appId?:string}>(), group=ref<string>(),workspaceMenu=ref(false),workspaceTitle=ref(''),workspaceBusy=ref(false),workspaceError=ref('')
// Review the selected artwork against neutral backdrops without changing saved preferences.
const reviewParams = import.meta.env.DEV ? new URLSearchParams(location.search) : undefined
const reviewBackdrop = reviewParams?.get('icon-study') === 'h04' ? reviewParams.get('icon-background') : null
const reviewBackground = reviewBackdrop === 'graphite'
  ? { backgroundImage: 'none', backgroundColor: '#222a31' }
  : reviewBackdrop === 'warm' ? { backgroundImage: 'none', backgroundColor: '#e9e5de' } : undefined
const systemMotion=usePreferredReducedMotion()
const reducedMotion=computed(()=>systemMotion.value==='reduce'||!!desktop.state!.preferences.reduceMotion)
const currentWorkspace=computed(()=>desktop.workspaces.find(w=>w.slot===desktop.activeWorkspace)?.title || '个人桌面')
const renameDraft=computed<{shortcutId:string;title:string}|undefined>({get:()=>desktop.state!.preferences.shortcutRename as {shortcutId:string;title:string}|undefined,set:value=>desktop.commit([{kind:'set',path:['preferences','shortcutRename'],value:json(value||null)}])})
function applyPreferences(){
  const preferences=desktop.state!.preferences,root=document.documentElement
  root.dataset.reduceMotion=String(!!preferences.reduceMotion)
  root.dataset.theme=String(preferences.theme||'light')
  root.dataset.density=String(preferences.density||'comfortable')
  root.dataset.translucency=preferences.translucency===false?'off':'on'
  root.dataset.palette=themePalette(preferences)
  root.dataset.wallpaperChoice=wallpaperStyle(preferences)
  const scale=Number(preferences.fontScale||1)
  root.style.setProperty('--ui-font-scale',String(scale>=.8&&scale<=1.2?scale:1))
}
watch(()=>[desktop.state!.preferences.reduceMotion,desktop.state!.preferences.theme,desktop.state!.preferences.density,desktop.state!.preferences.fontScale,desktop.state!.preferences.translucency,desktop.state!.preferences.palette,desktop.state!.preferences.wallpaper],applyPreferences,{immediate:true})
const appList=computed(()=>[...apps.values()].filter(app=>!app.manifest.permissions.includes('platform.admin') || session.user?.admin))
const groups=computed(()=>[...new Set(desktop.state!.order.map(id=>desktop.state!.windows[id]!.appId))])
const dockApps=computed(()=>[...new Set(['blora.instances','blora.files','blora.editor','blora.terminal','blora.extensions','blora.monitor','blora.settings',...groups.value])].filter(id=>id!=='blora.tasks'&&appList.value.some(app=>app.manifest.appId===id)))
const viewport=useWindowSize()
const dockMetrics=computed(()=>{
  const compact=viewport.width.value<=760
  const count=dockApps.value.length+2+(!compact&&desktop.state!.closedViews.length?1:0)
  const inset=14,hitPadding=compact?3:5,gap=compact?1:5,divider=compact?7:13
  const available=viewport.width.value-(compact?16:32)
  const fixed=2*(inset-hitPadding)+2*divider+(count+1)*gap+2*count*hitPadding
  // Fit every icon equally. Independent flex shrinking would change the corner
  // shape while leaving a fixed-height frame and break concentricity.
  const icon=Math.max(8,Math.min(compact?35:44,(available-fixed)/count))
  return {inset,style:{'--dock-inset':`${inset}px`,'--dock-hit-padding':`${hitPadding}px`,'--dock-icon-size':`${icon}px`,'--dock-gap':`${gap}px`,'--dock-divider-margin':`${(divider-1)/2}px`}}
})
const activeAppTitle=computed(()=>apps.get(desktop.state!.windows[desktop.state!.activeWindowId||'']?.appId||'')?.manifest.title||'桌面')
const taskSummary=useQuery({queryKey:['tasks','summary'],queryFn:({signal})=>api<{active:number;states:Record<string,number>}>('/tasks/summary',{signal}),refetchInterval:3000})
const shortcutResources=useQuery({queryKey:['instances'],enabled:computed(()=>Object.values(desktop.state!.shortcuts).some(shortcut=>shortcut.resourceRef.kind==='instance')),queryFn:()=>api<{items:Instance[]}>('/instances'),refetchInterval:3000})
function shortcutInstance(shortcut:Shortcut){return shortcut.resourceRef.kind==='instance'?shortcutResources.data.value?.items.find(instance=>instance.instanceId===shortcut.resourceRef.id&&instance.nodeId===shortcut.resourceRef.nodeId):undefined}
function shortcutTitle(shortcut:Shortcut){return shortcut.customTitle?shortcut.title:shortcutInstance(shortcut)?.name||shortcut.title}
function shortcutStatus(shortcut:Shortcut){if(!apps.has(shortcut.appId))return '应用不可用';if(shortcut.resourceRef.kind!=='instance')return '';if(shortcutResources.error.value||!shortcutResources.data.value)return '状态待确认';const instance=shortcutInstance(shortcut);return !instance?'不可访问':instance.nodeState==='OFFLINE'?'节点失联':instance.state}
const running=computed(()=>taskSummary.error.value||!taskSummary.data.value?'后台任务待确认':`${taskSummary.data.value.active} 项后台任务`)
function context(event:MouseEvent,shortcut?:Shortcut,appId?:string) { event.preventDefault(); menu.value={x:Math.min(event.clientX,innerWidth-240),y:Math.min(event.clientY,innerHeight-260),shortcut,appId}; launcher.value=false }
function openShortcut(s:Shortcut,newWindow=false) { desktop.open({appId:s.appId,resourceRef:s.resourceRef,title:shortcutTitle(s),disposition:newWindow?'new-window':'dedicated',state:{page:s.page||'overview',...s.state}});menu.value=undefined }
function renameShortcut() { const s=menu.value?.shortcut;if(s)renameDraft.value={shortcutId:s.shortcutId,title:shortcutTitle(s)};menu.value=undefined }
function saveShortcutName(){const draft=renameDraft.value;if(!draft?.title.trim())return;if(desktop.state!.shortcuts[draft.shortcutId])desktop.commit([{kind:'set',path:['shortcuts',draft.shortcutId,'title'],value:draft.title.trim()},{kind:'set',path:['shortcuts',draft.shortcutId,'customTitle'],value:true}]);renameDraft.value=undefined}
function removeShortcut() { const s=menu.value?.shortcut;if(s)desktop.commit([{kind:'delete',path:['shortcuts',s.shortcutId]}]);menu.value=undefined }
function exportWorkspace() { const url=URL.createObjectURL(new Blob([desktop.recovery!.export()],{type:'application/json'}));const a=document.createElement('a');a.href=url;a.download='blora-workspace.json';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000) }
async function changeWorkspace(slot?:string,duplicate=false){workspaceBusy.value=true;workspaceError.value='';try{if(slot)await desktop.switchWorkspace(slot);else await desktop.createWorkspace(workspaceTitle.value,duplicate);workspaceMenu.value=false;workspaceTitle.value=''}catch(e){workspaceError.value=String(e)}finally{workspaceBusy.value=false}}
async function toggleWorkspaceMenu(){workspaceMenu.value=!workspaceMenu.value;if(workspaceMenu.value){try{await desktop.refreshCloudWorkspaces()}catch(e){workspaceError.value=String(e)}}}
async function syncCloud(includeContent=false){if(includeContent&&!confirm('同步工作内容会上传当前草稿、终端屏幕和上传检查点到账号云端副本，继续？'))return;workspaceBusy.value=true;workspaceError.value='';try{await desktop.syncCloud(includeContent,currentWorkspace.value);await desktop.refreshCloudWorkspaces()}catch(e){workspaceError.value=e instanceof APIError&&e.code==='WORKSPACE_CONFLICT'?'云端副本已被其他设备更新；当前本地现场保留，请刷新列表后选择要打开的副本。':String(e)}finally{workspaceBusy.value=false}}
async function openCloud(item:{workspaceId:string;title:string}){workspaceBusy.value=true;workspaceError.value='';try{await desktop.openCloudWorkspace(item.workspaceId,item.title);workspaceMenu.value=false}catch(e){workspaceError.value=String(e)}finally{workspaceBusy.value=false}}
function taskbarClick(appId:string){launcher.value=false;const windows=Object.values(desktop.state!.windows).filter(w=>w.appId===appId);if(!windows.length){desktop.open({appId});return}if(windows.length===1){const w=windows[0]!;if(desktop.state!.activeWindowId===w.windowId && !w.minimized)desktop.minimize(w.windowId);else desktop.focus(w.windowId)}else group.value=group.value===appId?undefined:appId}
function dragShortcut(event:PointerEvent,s:Shortcut) {
  if(event.button!==0)return
  const target=event.currentTarget as HTMLElement,start={x:event.clientX,y:event.clientY,sx:s.x,sy:s.y};let moved=false
  target.setPointerCapture(event.pointerId)
  const move=(e:PointerEvent)=>{if(Math.abs(e.clientX-start.x)+Math.abs(e.clientY-start.y)>5)moved=true;if(moved){s.x=Math.max(0,start.sx+e.clientX-start.x);s.y=Math.max(DESKTOP_TOP,start.sy+e.clientY-start.y)}}
  const up=()=>{target.removeEventListener('pointermove',move);target.removeEventListener('pointerup',up);if(moved){desktop.commit([{kind:'set',path:['shortcuts',s.shortcutId],value:json(s)}]);target.addEventListener('click',e=>e.stopPropagation(),{once:true,capture:true})}}
  target.addEventListener('pointermove',move);target.addEventListener('pointerup',up)
}
useEventListener(window,'keydown',(e:KeyboardEvent)=>{
  if(e.key==='Escape'){menu.value=undefined;launcher.value=false;group.value=undefined;workspaceMenu.value=false}
  if((e.ctrlKey||e.metaKey) && e.shiftKey && e.key.toLowerCase()==='t' && !((e.target as HTMLElement).closest('input,textarea,[role="textbox"],.xterm'))){e.preventDefault();desktop.reopenView()}
  if(String(desktop.state!.preferences.windowCycleShortcut||'alt-backtick')==='alt-backtick' && e.altKey && e.key==='`'){e.preventDefault();const ids=desktop.state!.order;if(ids.length)desktop.focus(ids[0]!)}
  if(String(desktop.state!.preferences.launcherShortcut||'ctrl-alt-n')==='ctrl-alt-n' && e.ctrlKey && e.altKey && e.key.toLowerCase()==='n'){e.preventDefault();launcher.value=!launcher.value}
  if(menu.value && ['ArrowDown','ArrowUp','Home','End'].includes(e.key)){
    e.preventDefault();const items=[...document.querySelectorAll<HTMLButtonElement>('.context-menu button')];const index=items.indexOf(document.activeElement as HTMLButtonElement);items[e.key==='Home'?0:e.key==='End'?items.length-1:(index+(e.key==='ArrowDown'?1:-1)+items.length)%items.length]?.focus()
  }
})
useEventListener(window,'resize',()=>{if(innerWidth<=760)return;for(const w of Object.values(desktop.state!.windows)){if(w.snap)desktop.snap(w.windowId,w.snap);else desktop.geometry(w.windowId,{...w.rect,x:Math.max(-w.rect.width+120,Math.min(w.rect.x,innerWidth-120)),y:Math.max(DESKTOP_TOP,Math.min(w.rect.y,innerHeight-100))})}})
</script>
<template>
  <main class="desktop" :style="reviewBackground" @contextmenu.self="context($event)" @pointerdown.self="menu=undefined;launcher=false;group=undefined">
    <DesktopBar :workspace="currentWorkspace" :active-title="activeAppTitle" :protected="desktop.recovery!.status.protected" :protection-message="desktop.recovery!.status.message" :workspace-open="workspaceMenu" @workspace="toggleWorkspaceMenu" @export="exportWorkspace" @logout="$emit('logout')" @settings="desktop.open({appId:'blora.settings'})"/>
    <div v-if="workspaceMenu" class="workspace-menu" role="dialog" aria-label="工作区"><button v-for="workspace in desktop.workspaces" :key="workspace.slot" :disabled="workspaceBusy" :aria-current="workspace.slot===desktop.activeWorkspace" @click="changeWorkspace(workspace.slot)">{{workspace.title}} {{workspace.slot===desktop.activeWorkspace?'✓':''}}</button><form @submit.prevent="changeWorkspace()"><label>新工作区名称<input v-model="workspaceTitle" aria-label="新工作区名称" placeholder="例如：测试节点" required></label><div class="action-row"><button :disabled="workspaceBusy" class="primary">创建工作区</button><button type="button" :disabled="workspaceBusy || !workspaceTitle.trim()" @click="changeWorkspace(undefined,true)">复制当前现场</button></div></form><div class="workspace-cloud"><h3>账号云端副本</h3><p class="muted">默认只同步布局和资源引用；正文、终端与上传检查点需明确选择。</p><div class="action-row"><button :disabled="workspaceBusy" @click="syncCloud(false)">同步布局/引用</button><button :disabled="workspaceBusy" @click="syncCloud(true)">同步工作内容</button></div><button v-for="item in desktop.cloudWorkspaces" :key="item.workspaceId" :disabled="workspaceBusy" @click="openCloud(item)">{{item.title||item.workspaceId}} · v{{item.revision}} · {{item.includeContent?'含工作内容':'仅布局引用'}}</button><p v-if="!desktop.cloudWorkspaces.length" class="muted">尚无云端副本。</p></div><p v-if="workspaceError" class="error">{{workspaceError}}</p></div>
    <form v-if="renameDraft" class="shortcut-rename-dialog" role="dialog" aria-label="重命名桌面入口" @submit.prevent="saveShortcutName" @keydown.esc.stop="renameDraft=undefined"><h3>重命名桌面入口</h3><label>入口名称<input :value="renameDraft.title" aria-label="桌面入口名称" required maxlength="120" @input="renameDraft={...renameDraft,title:($event.target as HTMLInputElement).value}"></label><p class="muted">名称仅用于本工作区的快捷入口。</p><div class="action-row"><button type="button" @click="renameDraft=undefined">返回</button><button class="primary">保存入口名称</button></div></form>

    <nav class="desktop-apps" aria-label="默认应用"><button v-for="app in appList.filter(app=>['blora.instances','blora.files','blora.editor','blora.terminal','blora.tasks','blora.backups','blora.extensions'].includes(app.manifest.appId))" :key="app.manifest.appId" class="desktop-icon" :data-app="app.manifest.appId" @click="desktop.open({appId:app.manifest.appId})" @contextmenu="context($event,undefined,app.manifest.appId)"><span class="app-icon" :style="{'--app-color':app.manifest.color}"><AppIcon :app-id="app.manifest.appId" /></span><span>{{app.manifest.title}}</span></button></nav>
    <button v-for="shortcut in desktop.state!.shortcuts" :key="shortcut.shortcutId" class="desktop-icon resource-shortcut" :style="{left:shortcut.x+'px',top:shortcut.y+'px'}" :data-shortcut="shortcut.shortcutId" @pointerdown="dragShortcut($event,shortcut)" @click="openShortcut(shortcut)" @contextmenu="context($event,shortcut)"><span class="app-icon"><AppIcon :app-id="shortcut.appId"/><small>↗</small></span><span>{{shortcutTitle(shortcut)}}</span><small v-if="shortcutStatus(shortcut)" class="shortcut-resource-status">{{shortcutStatus(shortcut)}}</small></button>
    <!-- Keep mounted window DOM in creation order. Moving an iframe's parent
         during focus reloads its browsing context and can discard the click,
         extension state or an in-flight bridge call. Stacking is CSS-only. -->
    <WindowFrame v-for="windowId in Object.keys(desktop.state!.windows)" :key="desktop.recovery!.key+':'+windowId" :window-id="windowId" :z-index="desktop.state!.order.indexOf(windowId)+10" />
    <motion.div v-if="launcher" class="launcher" :initial="reducedMotion?false:{opacity:0,y:8}" :animate="{opacity:1,y:0}" :transition="reducedMotion?{duration:0}:{type:'spring',stiffness:460,damping:34,mass:.7}"><header><h2>应用</h2><span>{{appList.length}} 个已安装</span></header><div class="launcher-grid"><button v-for="app in appList" :key="app.manifest.appId" :aria-label="app.manifest.title" @click="desktop.open({appId:app.manifest.appId});launcher=false" @contextmenu="context($event,undefined,app.manifest.appId)"><span class="app-icon" aria-hidden="true" :style="{'--app-color':app.manifest.color}"><AppIcon :app-id="app.manifest.appId" /></span>{{app.manifest.title}}</button></div><footer>Ctrl + Alt + N 打开应用 · Alt + ` 切换窗口</footer></motion.div>
    <div v-if="group" class="window-picker"><button v-for="w in Object.values(desktop.state!.windows).filter(w=>w.appId===group)" :key="w.windowId" @click="desktop.focus(w.windowId);group=undefined">{{desktop.state!.views[w.activeTabId]?.title}} <span>{{w.minimized?'已最小化':'打开中'}}</span></button><button @click="desktop.open({appId:group,disposition:'new-window'});group=undefined">＋ 新建窗口</button></div>
    <footer class="taskbar" aria-label="应用 Dock" :style="dockMetrics.style"><DockSurface :inset="dockMetrics.inset"/><button class="launcher-button dock-item" aria-label="应用" :aria-expanded="launcher" @click="launcher=!launcher"><AppIcon app-id="blora.launcher" :padded="false"/><span class="dock-tooltip">应用</span></button><div class="taskbar-divider"></div><button v-for="appId in dockApps" :key="appId" class="dock-item" :class="{'taskbar-app':groups.includes(appId),running:groups.includes(appId),active:desktop.state!.windows[desktop.state!.activeWindowId||'']?.appId===appId}" :aria-label="apps.get(appId)?.manifest.title" @click="taskbarClick(appId)" @contextmenu="context($event,undefined,appId)"><AppIcon :app-id="appId" :padded="false"/><span class="dock-tooltip">{{apps.get(appId)?.manifest.title}}</span><small v-if="Object.values(desktop.state!.windows).filter(w=>w.appId===appId).length>1" class="dock-count">{{Object.values(desktop.state!.windows).filter(w=>w.appId===appId).length}}</small></button><div class="taskbar-divider"></div><button v-if="desktop.state!.closedViews.length" class="dock-item dock-utility" aria-label="重新打开最近标签" @click="desktop.reopenView()"><UIIcon name="history"/><span class="dock-tooltip">最近关闭</span></button><button class="dock-item" :aria-label="running" @click="desktop.open({appId:'blora.tasks'})"><AppIcon app-id="blora.tasks" :padded="false"/><span class="dock-tooltip">{{running}}</span></button></footer>
    <div v-if="menu" role="menu" tabindex="-1" class="context-menu" :style="{left:menu.x+'px',top:menu.y+'px'}"><template v-if="menu.shortcut"><button role="menuitem" @click="openShortcut(menu.shortcut)">打开管理窗口</button><button role="menuitem" @click="openShortcut(menu.shortcut,true)">在新窗口打开</button><button role="menuitem" @click="renameShortcut">重命名入口</button><button role="menuitem" @click="removeShortcut">移除桌面入口</button></template><template v-else-if="menu.appId"><button role="menuitem" @click="desktop.open({appId:menu.appId,disposition:'new-window'});menu=undefined">新建窗口</button></template><template v-else><button role="menuitem" @click="launcher=true;menu=undefined">打开应用</button><button role="menuitem" @click="exportWorkspace();menu=undefined">导出工作现场</button></template></div>
  </main>
</template>
