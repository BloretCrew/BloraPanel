import {api} from '../services/api'
import type {Json, ResourceRef} from './types'

/**
 * Creates the narrow transport given to an installed extension. The SDK still
 * performs its manifest capability check; this transport adds the browser
 * session, CSRF and idempotency headers through the shared API client and
 * exposes only the server-side task bridge.
 */
export interface ExtensionTransportHooks {
  notify?: (payload:Json) => Promise<Json>
  open?: (payload:Json) => Promise<Json>
  capture?: () => Promise<Json>
  restore?: (payload:Json) => Promise<Json>
  migrate?: (payload:Json) => Promise<Json>
  resource?: () => ResourceRef | undefined
}

export async function fetchExtensionBundle(appId:string):Promise<string> {
  if (!appId) throw new Error('读取扩展包需要应用标识')
  const response = await fetch(`/api/v1/extensions/${encodeURIComponent(appId)}/bundle`, {credentials:'same-origin', cache:'no-store'})
  if (!response.ok) throw new Error(`扩展包读取失败：${response.status}`)
  return response.text()
}

export function createExtensionTransport(appId:string, nodeId:string, hooks:ExtensionTransportHooks = {}, capabilities?: readonly string[]) {
  if (!appId) throw new Error('扩展任务桥接需要应用标识')
  const allowed=capabilities ? new Set(capabilities) : undefined
  return async (method:string, payload:Json):Promise<Json> => {
    if(method==='notification.publish'){
      if(!allowed?.has(method))throw new Error('扩展未声明能力：notification.publish')
      if(!hooks.notify)throw new Error('通知服务不可用')
      if(!payload||typeof payload!=='object'||Array.isArray(payload))throw new Error('通知参数无效')
      const {title,message='',key}=payload
      if(typeof title!=='string'||!title.trim()||typeof message!=='string'||(key!==undefined&&(typeof key!=='string'||!key||key.length>128)))throw new Error('通知参数无效')
      const result=await api<Json>(`/extensions/${encodeURIComponent(appId)}/notifications`,{method:'POST',body:JSON.stringify({title,message})})
      if(!result||typeof result!=='object'||Array.isArray(result)||result.appId!==appId)throw new Error('通知来源无效')
      return hooks.notify({...result,...(key?{key}:{})})
    }
    if(method==='resource.write'){
      if(!allowed?.has('resource.write'))throw new Error('扩展未声明能力：resource.write')
      if(!payload||typeof payload!=='object'||Array.isArray(payload))throw new Error('资源更新参数无效')
      const {requestId,resource:inputResource,...patch}=payload
      const resource=inputResource??hooks.resource?.()
      if(!resource||typeof resource!=='object'||Array.isArray(resource)||resource.kind!=='instance'||typeof resource.id!=='string')throw new Error('需要实例资源身份')
      if(typeof requestId!=='string'||!requestId||requestId.length>128)throw new Error('资源写入需要稳定请求标识')
      return api<Json>(`/extensions/${encodeURIComponent(appId)}/resource`,{method:'PATCH',headers:{'Idempotency-Key':requestId},body:JSON.stringify({...patch,resource})})
    }
    if(method==='data.read'||method==='data.write'){
      if(!allowed?.has(method))throw new Error(`扩展未声明能力：${method}`)
      const path=`/extensions/${encodeURIComponent(appId)}/data`
      if(method==='data.read')return api<Json>(path)
      if(!payload||typeof payload!=='object'||Array.isArray(payload))throw new Error('数据写入参数无效')
      const {requestId,...body}=payload
      if(typeof requestId!=='string'||!requestId||requestId.length>128)throw new Error('数据写入需要稳定请求标识')
      return api<Json>(path,{method:'PUT',headers:{'Idempotency-Key':requestId},body:JSON.stringify(body)})
    }
    if (allowed && method === 'window.open' && !allowed.has('window.open')) throw new Error('扩展未声明能力：window.open')
    if (allowed && ['task.create','task.read','task.cancel'].includes(method) && !allowed.has('task.create')) throw new Error('扩展未声明能力：task.create')
    if(method==='task.read'||method==='task.cancel'){
      const taskId=payload&&typeof payload==='object'&&!Array.isArray(payload)?payload.taskId:undefined
      if(typeof taskId!=='string'||!taskId||taskId.length>128)throw new Error('需要任务标识')
      const path=`/extensions/${encodeURIComponent(appId)}/tasks/${encodeURIComponent(taskId)}`
      if(method==='task.read')return api<Json>(`${path}/status`)
      const requestId=payload&&typeof payload==='object'&&!Array.isArray(payload)?payload.requestId:undefined
      if(typeof requestId!=='string'||!requestId.trim()||requestId.length>128)throw new Error('任务取消需要稳定请求标识')
      return api<Json>(`${path}/cancel`,{method:'POST',headers:{'Idempotency-Key':requestId}})
    }
    if (allowed && method === 'resource.read' && !allowed.has('resource.read')) throw new Error('扩展未声明能力：resource.read')
    if (method === 'resource.read') {
      const resource=payload && typeof payload==='object' && !Array.isArray(payload) ? payload as unknown as ResourceRef : hooks.resource?.()
      if(!resource) return null
      if(!['node','instance'].includes(resource.kind) || typeof resource.id!=='string' || !resource.id) throw new Error('无效资源标识')
      const query=new URLSearchParams({kind:resource.kind,id:resource.id})
      if(resource.nodeId) query.set('nodeId',resource.nodeId)
      return api<Json>(`/extensions/${encodeURIComponent(appId)}/resource?${query}`)
    }
    if (method === 'window.open' && hooks.open) return hooks.open(payload)
    if (method === 'state.capture' && hooks.capture) return hooks.capture()
    if (method === 'state.restore' && hooks.restore) return hooks.restore(payload)
    if (method === 'state.migrate' && hooks.migrate) return hooks.migrate(payload)
    if (method !== 'task.create') throw new Error(`宿主未提供能力：${method}`)
    if (!nodeId) throw new Error('扩展任务桥接缺少节点标识')
    if(!payload||typeof payload!=='object'||Array.isArray(payload)||typeof payload.requestId!=='string'||!payload.requestId.trim()||payload.requestId.length>128||!('payload' in payload))throw new Error('任务创建需要稳定请求标识和任务内容')
    return await api<Json>(`/extensions/${encodeURIComponent(appId)}/tasks`, {
      method:'POST',
      headers:{'Idempotency-Key':payload.requestId},
      body:JSON.stringify({nodeId, payload:payload.payload}),
    })
  }
}
