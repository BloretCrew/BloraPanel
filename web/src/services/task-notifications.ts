import type {useDesktop} from '../desktop/store'
import {decodeEnvelope,decodeJSON,encodeEnvelope,MessageType} from './protocol'
import {publishNotification} from './notifications'

export interface TaskCompletion {sequence:number;task:{taskId:string;state:string;action:string}}
const terminal:Record<string,string>={SUCCEEDED:'任务成功',FAILED:'任务失败',CANCELLED:'任务已取消',INTERRUPTED:'任务已中断'}
export function receiveTaskNotification(desktop:ReturnType<typeof useDesktop>,event:TaskCompletion){
  const before=desktop.state?.preferences.taskEventCursor
  if(!Number.isSafeInteger(event.sequence)||event.sequence<0||typeof before!=='number'||event.sequence<=before)return false
  if(!event.task||!terminal[event.task.state]||typeof event.task.taskId!=='string'||typeof event.task.action!=='string')throw new Error('任务通知事件无效')
  publishNotification(desktop,'blora.tasks',{title:terminal[event.task.state]!,message:event.task.action,key:event.task.taskId},undefined,{sequence:event.sequence,taskId:event.task.taskId})
  return true
}

export function subscribeTaskNotifications(desktop:ReturnType<typeof useDesktop>,status:(message:string)=>void){
  const workspace=desktop.state?.workspaceId,userId=desktop.state?.userId
  let stopped=false,socket:WebSocket|undefined,retry:ReturnType<typeof setTimeout>|undefined,heartbeat:ReturnType<typeof setInterval>|undefined,attempt=0
  const active=()=>!stopped&&desktop.state?.workspaceId===workspace&&desktop.state?.userId===userId
  function connect(){
    if(!active())return
    const cursor=desktop.state?.preferences.taskEventCursor
    const query=new URLSearchParams({notifications:'1'})
    if(typeof cursor==='number')query.set('after',String(cursor))
    const ws=new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/v1/events?${query}`)
    socket=ws;ws.binaryType='arraybuffer'
    let lastMessage=Date.now(),accepting=true
    ws.onopen=()=>{
      if(!active()){ws.close();return}
      status('正在同步任务通知')
      heartbeat=setInterval(()=>{
        if(Date.now()-lastMessage>45000){ws.close();return}
        if(ws.readyState===WebSocket.OPEN)ws.send(encodeEnvelope({protocolVersion:1,generation:1,channel:1,type:MessageType.Ping}))
      },15000)
    }
    ws.onmessage=event=>{
      if(!active()||!accepting)return
      try{
        const frame=decodeEnvelope(new Uint8Array(event.data as ArrayBuffer));lastMessage=Date.now()
        if(frame.type===MessageType.Snapshot){
          const sequence=frame.sequence||0,current=desktop.state?.preferences.taskEventCursor
          if(current===undefined||typeof current==='number'&&sequence>current)desktop.commit([{kind:'set',path:['preferences','taskEventCursor'],value:sequence}])
          if(desktop.recovery&&!desktop.recovery.status.protected)throw new Error(desktop.recovery.status.message||'通知游标保护失败')
          attempt=0;status('任务通知已连接')
        }else if(frame.type===MessageType.TaskEvent){
          const value=decodeJSON<TaskCompletion>(frame.payload)
          if(receiveTaskNotification(desktop,value)&&desktop.state?.preferences.systemNotifications===true&&globalThis.Notification?.permission==='granted'){
            // Persistent checkpoint precedes this optional, non-durable alert.
            // Reconnects and reloads never re-issue acknowledged system alerts.
            try{const alert=new Notification(terminal[value.task.state]!,{body:value.task.action,tag:'blora-task-'+value.task.taskId});alert.onclick=()=>{if(active()){window.focus();desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:value.task.taskId},disposition:'dedicated'})}alert.close()}}catch{/* The persistent in-app notification remains available. */}
          }
        }
      }catch(error){accepting=false;status('任务通知保存失败：'+String(error));ws.close()}
    }
    ws.onclose=()=>{
      if(heartbeat)clearInterval(heartbeat)
      if(active()){status('任务通知连接中断，正在重连');retry=setTimeout(connect,Math.min(30000,1000*2**Math.min(attempt++,5)))}
    }
  }
  connect()
  return ()=>{stopped=true;if(retry)clearTimeout(retry);if(heartbeat)clearInterval(heartbeat);socket?.close()}
}
