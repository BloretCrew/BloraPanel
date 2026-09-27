<script setup lang="ts">
import {onBeforeUnmount, onMounted, ref} from 'vue'
import {createExtensionTransport, fetchExtensionBundle} from './extension-runtime'
import { computed } from 'vue'
import { useDesktop } from '../desktop/store'
import type { Json, OpenRequest, ResourceRef } from './types'
import {desktopMethods,extensionDesktopAction} from './extension-desktop'
import {publishNotification} from '../services/notifications'

const props=defineProps<{appId:string;viewTabId:string;capabilities?:string[];stateSchemaVersion?:number}>()
const desktop=useDesktop()
const nodeId=computed(()=>{
  const resource=desktop.state?.views[props.viewTabId]?.resourceRef
  return String(resource?.nodeId || (resource?.kind==='node'?resource.id:''))
})
const frame=ref<HTMLIFrameElement>()
const loadError=ref('')
let disposed=false,bundleSource=''

async function openWindow(payload:Json):Promise<Json> {
  const input=(payload && typeof payload==='object' ? payload : {}) as Record<string,Json>
  const resource=input.resource && typeof input.resource==='object' ? input.resource as unknown as ResourceRef : undefined
  const viewTabId=desktop.open({appId:props.appId,entrypoint:typeof input.entrypoint==='string'?input.entrypoint:undefined,resourceRef:resource,disposition:(input.disposition==='new-window'||input.disposition==='new-tab'||input.disposition==='dedicated')?input.disposition:'default'} as OpenRequest)
  const window=Object.values(desktop.state?.windows||{}).find(w=>w.tabs.includes(viewTabId))
  if(!window) throw new Error('扩展窗口创建失败')
  return {windowId:window.windowId,viewTabId}
}
async function captureState():Promise<Json>{
  const view=desktop.state?.views[props.viewTabId]
  const current=view?.state||{}
  const value=JSON.parse(JSON.stringify(current)) as Json
  return {schemaVersion:view?.stateSchemaVersion||1,state:value,...(view?.resourceRef?{resource:JSON.parse(JSON.stringify(view.resourceRef)) as Json}:{})}
}
async function restoreState(payload:Json):Promise<Json>{
  if(!payload || typeof payload!=='object') throw new Error('扩展现场格式无效')
  const version=(payload as Record<string,Json>).schemaVersion
  if(version!==(props.stateSchemaVersion||1) || version!==(desktop.state?.views[props.viewTabId]?.stateSchemaVersion||1))throw new Error('扩展现场版本不匹配，请先完成迁移')
  const state=(payload as Record<string,Json>).state
  if(!state || typeof state!=='object' || Array.isArray(state)) throw new Error('扩展现场状态无效')
  desktop.commit([{kind:'set',path:['views',props.viewTabId,'state'],value:state}])
  if(!desktop.recovery?.status.protected)throw new Error(desktop.recovery?.status.message||'本地现场保护失败')
  return null
}
async function migrateState(payload:Json):Promise<Json>{
  if(!payload||typeof payload!=='object'||Array.isArray(payload))throw new Error('迁移数据无效')
  const {from,to}=payload
  if(!from||!to||typeof from!=='object'||typeof to!=='object'||Array.isArray(from)||Array.isArray(to))throw new Error('迁移现场无效')
  const current=await captureState()
  if(JSON.stringify(current)!==JSON.stringify(from))throw new Error('迁移期间原始现场已变化，保留当前现场')
  if(to.schemaVersion!==(props.stateSchemaVersion||1)||!to.state||typeof to.state!=='object'||Array.isArray(to.state))throw new Error('迁移目标版本或状态无效')
  // A migration may reshape view state, but it cannot retarget the view to a
  // different node/instance.  Resource identity belongs to the host and is
  // checked again before committing the new schema version.
  if (to.resource !== undefined) {
    const fromResource = (current as Record<string,Json>).resource
    const targetResource = to.resource
    if (!fromResource || !targetResource || typeof fromResource !== 'object' || Array.isArray(fromResource) || typeof targetResource !== 'object' || Array.isArray(targetResource)) throw new Error('迁移不能改变资源身份')
    const a = fromResource as Record<string,Json>, b = targetResource as Record<string,Json>
    if (a.kind !== b.kind || a.id !== b.id || a.nodeId !== b.nodeId) throw new Error('迁移不能改变资源身份')
  }
  if(JSON.stringify(to).length>1024*1024)throw new Error('迁移现场超过 1 MiB')
  desktop.commit([
    {kind:'set',path:['views',props.viewTabId,'state'],value:to.state},
    {kind:'set',path:['views',props.viewTabId,'stateSchemaVersion'],value:to.schemaVersion},
  ])
  return null
}

function respond(id:string,ok:boolean,value:unknown){
  frame.value?.contentWindow?.postMessage({type:'blora-extension-result',id,ok,value},'*')
}
async function onMessage(event:MessageEvent){
  if(disposed||event.source!==frame.value?.contentWindow)return
  if(event.data?.type==='blora-extension-load-error'){loadError.value=String(event.data.message||'扩展加载失败');return}
  if(event.data?.type==='blora-extension-ready'){
    frame.value?.contentWindow?.postMessage({type:'blora-extension-init',bundle:bundleSource,stateSchemaVersion:props.stateSchemaVersion||1},'*')
    return
  }
  if(event.data?.type!=='blora-extension-call')return
  const {id,method,payload}=event.data as {id:string;method:string;payload:unknown}
  try{
    // Respond synchronously before Vue unmounts a moved/closed iframe. The
    // mutation already committed; restoration never repeats this call.
    if(desktopMethods.has(method)){
      respond(id,true,extensionDesktopAction(desktop,props.appId,props.viewTabId,props.capabilities||[],method,payload as Json))
      return
    }
    // Resolve the resource on every call. A view can be moved or retargeted
    // while its iframe remains mounted; task.create must never keep a stale
    // node identity from the first message.
    const workspace=desktop.state?.workspaceId,userId=desktop.state?.userId
    const notify=async(value:Json):Promise<Json>=>{
      if(disposed||desktop.state?.workspaceId!==workspace||desktop.state?.userId!==userId)throw new Error('原工作区已关闭')
      const input=value as {title:string;message:string;key?:string}
      return publishNotification(desktop,props.appId,input,props.viewTabId)
    }
    const transport=createExtensionTransport(props.appId,nodeId.value,{notify,open:openWindow,capture:captureState,restore:restoreState,migrate:migrateState,resource:()=>desktop.state?.views[props.viewTabId]?.resourceRef},props.capabilities)
    respond(id,true,await transport(method,payload as never))
  }catch(error){respond(id,false,error instanceof Error?error.message:String(error))}
}
onMounted(async()=>{
  window.addEventListener('message',onMessage)
  try{
    bundleSource=await fetchExtensionBundle(props.appId)
    const closeScript='<'+'/script>'
    if(!disposed&&frame.value)frame.value.srcdoc=`<!doctype html><meta charset="utf-8"><body><main id="app"></main><script src="/blora-extension-bootstrap.js">${closeScript}`
  }catch(error){if(!disposed)loadError.value=String(error)}
})
onBeforeUnmount(()=>{disposed=true;window.removeEventListener('message',onMessage);bundleSource=''})
</script>
<template><p v-if="loadError" role="alert" class="error">{{loadError}}</p><iframe v-else ref="frame" class="extension-sandbox" sandbox="allow-scripts" :title="`扩展 ${appId}`"></iframe></template>
