<script setup lang="ts">
import {computed,ref,watch,onBeforeUnmount} from 'vue'
import UIIcon from '../app-host/UIIcon.vue'
import {useDesktop} from './store'
import {apps,loadExternalSandboxApp,isSandboxApp} from '../app-host/registry'
import {notifications,dismissNotification,type WorkspaceNotification} from '../services/notifications'
import {subscribeTaskNotifications} from '../services/task-notifications'
const desktop=useDesktop(),visible=ref(false),error=ref('')
const items=computed(()=>notifications(desktop))
const connection=ref('')
let unsubscribe:(()=>void)|undefined
watch(()=>[desktop.state?.userId,desktop.state?.workspaceId],()=>{unsubscribe?.();unsubscribe=subscribeTaskNotifications(desktop,value=>{connection.value=value})},{immediate:true})
onBeforeUnmount(()=>unsubscribe?.())
async function open(item:WorkspaceNotification){
  error.value=''
  try{
    if(!apps.has(item.appId)||isSandboxApp(item.appId))await loadExternalSandboxApp(item.appId)
    const view=item.viewTabId?desktop.state?.views[item.viewTabId]:undefined
    const window=view&&view.appId===item.appId?Object.values(desktop.state?.windows||{}).find(w=>w.tabs.includes(view.viewTabId)):undefined
    if(window&&view){desktop.commit([{kind:'set',path:['windows',window.windowId,'activeTabId'],value:view.viewTabId}]);desktop.focus(window.windowId)}
    else desktop.open({appId:item.appId,resourceRef:item.resourceRef,disposition:'dedicated'})
    visible.value=false
  }catch(e){error.value=String(e)}
}
</script>
<template>
  <button class="notification-toggle" aria-label="站内通知" :aria-expanded="visible" @click="visible=!visible"><UIIcon name="bell"/><span v-if="items.length" class="notification-count">{{items.length}}</span></button>
  <section v-if="visible" class="notification-panel" aria-label="站内通知列表">
    <header><strong>站内通知</strong><button @click="visible=false">关闭通知</button></header>
    <p role="status">{{connection}}</p>
    <p v-if="error" role="alert">{{error}}</p><p v-if="!items.length">暂无通知</p>
    <article v-for="item in items" :key="item.id">
      <small>{{item.appId}} · {{new Date(item.createdAt).toLocaleString()}}</small>
      <strong>{{item.title}}</strong><p>{{item.message}}</p>
      <small v-if="item.resourceRef">{{item.resourceRef.kind}} · {{item.resourceRef.id}}</small>
      <button @click="open(item)">打开来源</button><button @click="dismissNotification(desktop,item.id)">移除通知</button>
    </article>
  </section>
</template>
<style scoped>
.notification-panel{position:absolute;right:54px;top:62px;z-index:10000;width:min(360px,calc(100vw - 40px));max-height:65vh;overflow:auto;padding:18px;border:1px solid transparent;border-radius:var(--shape-xl);background:var(--surface-container);color:var(--on-surface);box-shadow:var(--menu-shadow)}
header{display:flex;align-items:center;justify-content:space-between;margin-bottom:14px;font-size:13px}header button{font-size:11px;color:var(--muted)}article{padding:14px;margin-top:8px;border-radius:var(--shape-lg);background:var(--surface-low)}article strong,article small{display:block}article strong{font-size:12px}article p{white-space:pre-wrap;overflow-wrap:anywhere;font-size:12px}article small{font-size:10px;color:var(--muted)}section>p{font-size:12px;color:var(--muted)}@media(max-width:640px){.notification-panel{right:0;top:59px}}
</style>
