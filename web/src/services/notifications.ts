import type {useDesktop} from '../desktop/store'
import type {Json,ResourceRef} from '../app-host/types'
import {json} from '../recovery/state'

export interface WorkspaceNotification {
  id:string; appId:string; title:string; message:string; createdAt:number
  viewTabId?:string; resourceRef?:ResourceRef
}
const rates=new WeakMap<object,Map<string,number[]>>()
export function notifications(desktop:ReturnType<typeof useDesktop>):WorkspaceNotification[]{
  const value=desktop.state?.preferences.notifications
  return Array.isArray(value)?value as unknown as WorkspaceNotification[]:[]
}
// Shared by trusted applications and the checked extension transport. Store only
// inert text and an application/resource identity, never executable actions/URLs.
export function publishNotification(desktop:ReturnType<typeof useDesktop>,appId:string,input:{title:string;message:string;key?:string},viewTabId?:string,taskCheckpoint?:{sequence:number;taskId:string}):Json{
  if(!desktop.state)throw new Error('工作区未加载')
  const view=viewTabId?desktop.state.views[viewTabId]:undefined
  if(viewTabId&&view?.appId!==appId)throw new Error('通知视图身份不匹配')
  if(!input.title.trim()||input.title.length>256||input.message.length>4096)throw new Error('通知内容超出限制')
  const rows=notifications(desktop),now=Date.now()
  const rate=rates.get(desktop.state)||new Map<string,number[]>()
  rates.set(desktop.state,rate)
  const recent=(rate.get(appId)||[]).filter(time=>now-time<1000)
  if(!taskCheckpoint&&recent.length>=5)throw new Error('通知过于频繁，请稍后重试')
  rate.set(appId,[...recent,now])
  const id=`${appId}:${input.key||crypto.randomUUID()}`
  const item:WorkspaceNotification={id,appId,title:input.title,message:input.message,createdAt:now,...(view?{viewTabId:view.viewTabId,...(view.resourceRef?{resourceRef:view.resourceRef}:{})}:{}),...(taskCheckpoint?{resourceRef:{kind:'task',id:taskCheckpoint.taskId}}:{})}
  desktop.commit([{kind:'set',path:['preferences','notifications'],value:json([item,...rows.filter(n=>n.id!==id)].slice(0,100))},...(taskCheckpoint?[{kind:'set' as const,path:['preferences','taskEventCursor'],value:taskCheckpoint.sequence}]:[])])
  if(desktop.recovery&&!desktop.recovery.status.protected)throw new Error(desktop.recovery.status.message||'通知现场保护失败')
  return {notificationId:id}
}
export function dismissNotification(desktop:ReturnType<typeof useDesktop>,id:string){
  desktop.commit([{kind:'set',path:['preferences','notifications'],value:json(notifications(desktop).filter(n=>n.id!==id))}])
}
