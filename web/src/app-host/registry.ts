import { defineAsyncComponent, defineComponent, markRaw, h, shallowReactive } from 'vue'
import type { AppDefinition, AppManifest } from './types'
import type { Component } from 'vue'
import { api, APIError } from '../services/api'
export const apps = shallowReactive(new Map<string, AppDefinition>())
const sandboxApps=new Set<string>()
export const isSandboxApp=(appId:string)=>sandboxApps.has(appId)
export function registerApp(app: AppDefinition) {
  if (app.manifest.hostApiVersion !== 1) throw new Error('应用宿主 API 版本不兼容')
  if (apps.has(app.manifest.appId)) throw new Error('应用标识已注册')
  apps.set(app.manifest.appId, { ...app, component: markRaw(app.component) })
}
/** Register a package supplied by the controlled extension loader. */
export function registerExternalApp(manifest: AppManifest, component: Component, hooks?: Partial<Pick<AppDefinition,'captureState'|'restoreState'|'migrateState'|'reconcileResource'>>) {
  validateExternalManifest(manifest)
  registerApp({manifest, component, captureState:hooks?.captureState||((state)=>state), restoreState:hooks?.restoreState||((state)=>state), migrateState:hooks?.migrateState||((state,from)=>{if(from!==manifest.stateSchemaVersion)throw new Error('扩展未提供此版本的状态迁移');return state}), reconcileResource:hooks?.reconcileResource|| (async()=>undefined)})
}
export async function loadExternalApp(appId: string, component: Component, hooks?: Partial<Pick<AppDefinition,'captureState'|'restoreState'|'migrateState'|'reconcileResource'>>) {
  const response = await fetch(`/api/v1/extensions/${encodeURIComponent(appId)}/manifest`, { credentials: 'same-origin' })
  if (!response.ok) throw new Error(`扩展清单读取失败：${response.status}`)
  const manifest = normalizeExternalManifest(await response.json() as AppManifest)
  registerExternalApp(manifest, component, hooks)
  return manifest
}
/** Load an enabled package into an opaque-origin iframe without evaluating it
 * in the desktop document. The iframe bridge is limited to SDK transport. */
export async function loadExternalSandboxApp(appId:string) {
  const response=await fetch(`/api/v1/extensions/${encodeURIComponent(appId)}/manifest`,{credentials:'same-origin'})
  if(!response.ok)throw new Error(`扩展清单读取失败：${response.status}`)
  const manifest=normalizeExternalManifest(await response.json() as AppManifest)
  validateExternalManifest(manifest)
  if(manifest.appId!==appId)throw new Error('扩展清单标识与请求不一致')
  const previous=apps.get(appId)
  if(previous){
    if(!sandboxApps.has(appId))throw new Error('不能替换其他来源注册的应用')
    if(JSON.stringify(previous.manifest)===JSON.stringify(manifest))return manifest
  }
  const ExternalSandboxApp=defineAsyncComponent(()=>import('./ExternalSandboxApp.vue'))
  const component=defineComponent({name:'ExternalSandboxHost',props:{viewTabId:{type:String,required:true}},setup:(props)=>()=>h(ExternalSandboxApp,{appId,viewTabId:props.viewTabId,capabilities:manifest.capabilities||[],stateSchemaVersion:manifest.stateSchemaVersion})})
  const reconcileResource = async (resource?: import('./types').ResourceRef) => {
    if (!resource || !manifest.capabilities?.includes('resource.read') || !['node','instance'].includes(resource.kind)) return undefined
    const query = new URLSearchParams({kind:resource.kind,id:resource.id})
    // Instance IDs are stable across node presentation changes; omitting a
    // stale node hint lets the Master return the canonical resource tag.
    if (resource.kind === 'node' && resource.nodeId) query.set('nodeId', resource.nodeId)
    try {
      const result = await api<{resource: import('./types').ResourceRef}>(`/extensions/${encodeURIComponent(appId)}/resource?${query}`)
      return result.resource
    } catch (error) {
      // Revocation/deletion keeps the local view and draft; the desktop store
      // renders it as unavailable until the resource is authorized again.
      if (error instanceof APIError && (error.status===403 || error.status===404)) return undefined
      throw error
    }
  }
  if(previous) apps.set(appId,{...previous,manifest,component:markRaw(component),reconcileResource,migrateState:(state,from)=>{if(from!==manifest.stateSchemaVersion)throw new Error('扩展未提供同步迁移，等待沙箱迁移');return state}})
  else registerExternalApp(manifest,component,{reconcileResource})
  sandboxApps.add(appId)
  return manifest
}
function normalizeExternalManifest(manifest:AppManifest):AppManifest {
  return {...manifest,
    permissions:manifest.permissions||[], capabilities:manifest.capabilities||[],
    entrypoints:manifest.entrypoints||[], resourceHandlers:manifest.resourceHandlers||[],
    dependencies:manifest.dependencies||[], tabPolicy:manifest.tabPolicy||{types:[],movable:true},
    stateSchemaVersion:manifest.stateSchemaVersion||1,
  }
}
const externalAppID=/^[-a-z0-9]+\.[-a-z0-9.]+$/
const externalVersion=/^v?\d+\.\d+\.\d+$/
const externalCapabilities=new Set(['resource.read','resource.write','task.create','window.open','window.move','window.close','shortcut.create','data.read','data.write','notification.publish'])
function validateExternalManifest(manifest:AppManifest) {
  if (manifest.protected) throw new Error('外部扩展不能声明 protected')
  if (!externalAppID.test(manifest.appId) || manifest.appId.length>128) throw new Error('扩展应用标识无效')
  if(manifest.appId.startsWith('blora.'))throw new Error('扩展不能使用保留应用标识')
  if(manifest.appId.endsWith('.previous'))throw new Error('扩展标识使用保留后缀')
  if (!externalVersion.test(manifest.packageVersion) || manifest.packageVersion.length>64) throw new Error('扩展版本无效')
  if (manifest.hostApiVersion !== 1) throw new Error('应用宿主 API 版本不兼容')
  if (!Array.isArray(manifest.entrypoints) || manifest.entrypoints.length>64 || manifest.entrypoints.some(item=>typeof item!=='string'||item.length>128)) throw new Error('扩展入口声明无效')
  if (!Array.isArray(manifest.resourceHandlers) || manifest.resourceHandlers.length>64 || manifest.resourceHandlers.some(item=>typeof item!=='string'||item.length>128)) throw new Error('扩展资源处理器声明无效')
  if (!Array.isArray(manifest.permissions) || manifest.permissions.some(item=>typeof item!=='string'||!item||item.length>128)) throw new Error('扩展权限声明无效')
  if (!Array.isArray(manifest.capabilities) || manifest.capabilities.some(item=>typeof item!=='string'||!externalCapabilities.has(item))) throw new Error('扩展能力声明无效')
  if (manifest.windowPolicy !== 'multiple' || !manifest.tabPolicy || manifest.tabPolicy.movable !== true || !Array.isArray(manifest.tabPolicy.types)) throw new Error('扩展窗口策略无效')
  if (manifest.tabPolicy.types.some(item=>typeof item!=='string'||item.length>128)) throw new Error('扩展标签类型无效')
  if (!Number.isInteger(manifest.stateSchemaVersion) || manifest.stateSchemaVersion<1 || manifest.stateSchemaVersion>1_000_000) throw new Error('扩展状态版本无效')
  if (Array.isArray(manifest.dependencies)) {
    if (manifest.dependencies.some(item=>typeof item!=='string'||item.length>128)) throw new Error('扩展依赖声明无效')
  } else if (!manifest.dependencies || typeof manifest.dependencies!=='object' || Object.entries(manifest.dependencies).some(([id,version])=>!externalAppID.test(id)||typeof version!=='string'||version.length>128)) {
    throw new Error('扩展依赖声明无效')
  }
}
export function getApp(appId: string) { const app = apps.get(appId); if (!app) throw new Error(`应用未安装：${appId}`); return app }
