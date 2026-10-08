<script setup lang="ts">
import {computed,onBeforeUnmount,onMounted,ref,watch} from 'vue'
import {useDesktop} from '../desktop/store'
import {api,type Task} from '../services/api'
import {TerminalModel,type TerminalEvent} from '../services/terminals'
import {decodeTerminalEvents} from '../services/terminal-events'
import {ByteWindow,decodeEnvelope,decodeJSON,encodeEnvelope,encodeJSON,INTERACTIVE_WINDOW,MessageType,type Envelope} from '../services/protocol'
import {json} from '../recovery/state'
import type {ResourceRef} from '../app-host/types'
interface TerminalSession {sessionId:string;containerId?:string;state:string;resource:ResourceRef;backend:string;diagnostic?:string;cols:number;rows:number;archive:{maxBytes:number};maxSessions:number;maxAttachments:number}
interface EndRequest {sessionId:string;requestId:string;taskId?:string;state?:string}
const props=defineProps<{viewTabId:string;visible:boolean}>(),desktop=useDesktop(),recovery=desktop.recovery!
const container=ref<HTMLElement>(),error=ref(''),status=ref('未连接'),sessions=ref<TerminalSession[]>([]),busy=ref(false),writable=ref(false),connected=ref(false)
const view=computed(()=>recovery.state.views[props.viewTabId]!),sessionId=computed(()=>String(view.value.state.sessionId||'')),fontSize=computed(()=>Number(view.value.state.fontSize||13))
const selected=computed(()=>sessions.value.find(s=>s.sessionId===sessionId.value)),pendingTask=computed(()=>String(view.value.state.createTaskId||''))
const uiTheme=computed<'light'|'dark'>(()=>desktop.state?.preferences.theme==='dark'?'dark':'light')
const endRequest=computed<EndRequest|undefined>({get:()=>view.value.state.endRequest as unknown as EndRequest|undefined,set:value=>patch('endRequest',value||null)}),endDialog=computed({get:()=>!!view.value.state.endDialog,set:value=>patch('endDialog',value)})
let terminal:TerminalModel|undefined,socket:WebSocket|undefined,resize:ResizeObserver|undefined,taskTimer:ReturnType<typeof setTimeout>|undefined,endTimer:ReturnType<typeof setTimeout>|undefined,reconnectTimer:ReturnType<typeof setTimeout>|undefined,resizeFrame=0,disposed=false,connection=0
function patch(key:string,value:unknown){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
const hostContainer=computed(()=>view.value.resourceRef?.kind==='node'?String(view.value.state.containerId||''):'')
function terminalPath(create=false){const resource=view.value.resourceRef!;return resource.kind==='node'?create?`/nodes/${encodeURIComponent(resource.id)}/docker/containers/${encodeURIComponent(hostContainer.value)}/terminals`:`/nodes/${encodeURIComponent(resource.id)}/terminals`:`/instances/${encodeURIComponent(resource.id)}/terminals`}
async function refreshSessions(){if(!view.value.resourceRef)return;const result=await api<{items:TerminalSession[]}>(terminalPath());sessions.value=hostContainer.value?result.items.filter(s=>s.containerId===hostContainer.value):result.items}
async function load(){
  if(!view.value.resourceRef){status.value='请从实例打开终端，以明确节点和授权范围';return}
  try{
    await refreshSessions();if(disposed)return
    if(endRequest.value?.state==='SUBMITTING'||endRequest.value?.taskId&&!['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(endRequest.value.state||''))void trackEnd()
    if(pendingTask.value){await trackCreate(pendingTask.value);return}
    if(view.value.state.createRequestId){const {items}=await api<{items:Task[]}>('/tasks');const existing=items.find(task=>task.requestId===view.value.state.createRequestId);if(existing){patch('createTaskId',existing.taskId);await trackCreate(existing.taskId);return}status.value='创建结果尚未确认；重新点击新建会话会复用原请求身份'}
    if(sessionId.value)await attach(sessionId.value)
  }catch(e){error.value=String(e)}
}
async function create(){
  if(!view.value.resourceRef||busy.value||pendingTask.value)return
  const requestId=String(view.value.state.createRequestId||crypto.randomUUID());patch('createRequestId',requestId);busy.value=true;error.value=''
  try{
    const target=view.value.state.containerTarget as {createdAt?:string}|undefined
    const {task}=await api<{task:Task}>(terminalPath(true),{method:'POST',headers:{'Idempotency-Key':requestId},body:JSON.stringify({cols:100,rows:28,...(hostContainer.value?{createdAt:target?.createdAt}:{})})})
    patch('createTaskId',task.taskId);status.value='会话创建任务已接受';if(!disposed)await trackCreate(task.taskId)
  }catch(e){error.value=String(e)}finally{busy.value=false}
}
async function trackCreate(taskId:string){
  if(taskTimer)clearTimeout(taskTimer)
  try{
    const {task}=await api<{task:Task}>(`/tasks/${encodeURIComponent(taskId)}`);if(disposed)return
    status.value=`创建会话：${task.state} · ${task.phase}`
    if(task.state==='SUCCEEDED'){
      const result=task.result?.session as TerminalSession|undefined;if(!result?.sessionId)throw new Error('会话任务成功但缺少会话身份，请到任务中心核对')
      patch('createTaskId',null);patch('createRequestId',null);patch('sessionId',result.sessionId);await refreshSessions();await attach(result.sessionId)
    }else if(['FAILED','CANCELLED','INTERRUPTED'].includes(task.state)){patch('createTaskId',null);patch('createRequestId',null);error.value=task.error||`会话创建${task.state}`}
    else taskTimer=setTimeout(()=>void trackCreate(taskId),1000)
  }catch(e){error.value=String(e);if(!disposed)taskTimer=setTimeout(()=>void trackCreate(taskId),3000)}
}
function send(type:MessageType,payload?:Uint8Array,sequence=0,credit=0){
  if(!socket||socket.readyState!==WebSocket.OPEN)throw new Error('终端未连接，此次输入未发送')
  const frame=encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:sessionId.value,type,payload,sequence,credit})
  if(socket.bufferedAmount+frame.byteLength>INTERACTIVE_WINDOW)throw new Error('浏览器发送缓冲已满，此次操作未发送')
  socket.send(frame)
}
function stopConnection(){connection++;if(reconnectTimer)clearTimeout(reconnectTimer);resize?.disconnect();if(resizeFrame)cancelAnimationFrame(resizeFrame);resizeFrame=0;socket?.close();socket=undefined;terminal?.dispose();terminal=undefined;connected.value=false;writable.value=false}
async function attach(id:string){
  if(!container.value||disposed)return
  stopConnection();const generation=connection;patch('sessionId',id);error.value='';status.value='正在鉴权并挂载原会话'
  const bytes=new ByteWindow(),model=new TerminalModel(recovery,id,props.viewTabId,input=>{
    if(generation!==connection||disposed||!writable.value||!connected.value)return
    try{const data=new TextEncoder().encode(input);if(data.byteLength>32*1024)throw new Error('单次终端输入超过32 KiB，未发送，请分段粘贴');if(!socket||socket.bufferedAmount+data.length+512>INTERACTIVE_WINDOW)throw new Error('终端输入缓冲已满，此次输入未发送');const sequence=bytes.reserve(data.length);send(MessageType.Data,data,sequence)}catch(e){error.value=String(e)}
  },uiTheme.value)
  terminal=model;await model.mount(container.value);if(disposed||generation!==connection){model.dispose();return}
  const protocol=location.protocol==='https:'?'wss:':'ws:'
  const ws=new WebSocket(`${protocol}//${location.host}/api/v1/terminals/${encodeURIComponent(id)}/stream?viewId=${encodeURIComponent(props.viewTabId)}&sequence=${model.sequence}&encoding=events-v1`);socket=ws;ws.binaryType='arraybuffer'
  let processing=false,queuedBytes=0,canReconnect=true
  const pending:{frame:Envelope;size:number}[]=[]
  function fail(e:unknown){if(generation!==connection||disposed)return;error.value=String(e);status.value='连接已停止 · 原检查点保留';writable.value=false;connected.value=false;model.setLease(false);canReconnect=false;ws.close(4002,'terminal protocol failed')}
  async function process(frame:Envelope){
    if(generation!==connection||disposed||!canReconnect)return
    if(frame.protocolVersion!==1||frame.generation!==1||frame.channel!==2||frame.streamId!==id)throw new Error('终端消息身份或连接代次不匹配')
    if(frame.type===MessageType.OpenAck){
      const info=decodeJSON<{sessionId:string;writable:boolean;earliest:number;latest:number;state:string}>(frame.payload)
      if(info.sessionId!==id)throw new Error('终端应答资源不匹配')
      if(model.sequence+1<info.earliest)throw new Error('输出归档已超过保留范围，原屏幕保留；不能无缝恢复此检查点')
      writable.value=!!info.writable;model.setLease(writable.value);status.value=model.replaying?'正在回放原会话':writable.value?'已连接 · 输入控制者':'已连接 · 只读观察'
    }else if(frame.type===MessageType.Resume){await model.restoreComplete();if(generation!==connection||disposed||!canReconnect)return;connected.value=true;status.value=writable.value?'已连接 · 输入控制者':'已连接 · 只读观察';requestSize()}
    else if(frame.type===MessageType.Ack)bytes.acknowledge(frame.sequence||0,frame.credit||0)
    else if(frame.type===MessageType.Error){const detail=decodeJSON<{code:string;message:string}>(frame.payload);throw new Error(`${detail.code}：${detail.message}`)}
    else if(frame.type===MessageType.Ping)send(MessageType.Pong,frame.payload)
    else if(frame.type===MessageType.Close||frame.type===MessageType.Reset){canReconnect=false;status.value='会话已关闭或连接被撤销';writable.value=false;model.setLease(false);ws.close()}
  }
  async function drain(){
    if(processing)return
    processing=true
    try{
      while(pending.length&&generation===connection&&!disposed&&canReconnect){
        const first=pending.shift()!,group=[first]
        let size=first.size
        if(first.frame.type===MessageType.Data){
          // Consume only adjacent data frames, preserving control/resize order.
          // One xterm write can parse many small PTY reads without a timer and
          // full recovery transaction for every individual output event.
          while(pending[0]?.frame.type===MessageType.Data&&size+pending[0].size<=64*1024){const next=pending.shift()!;group.push(next);size+=next.size}
        }
        try{
          if(first.frame.type===MessageType.Data){
            const events=group.flatMap(({frame})=>{
              if(frame.protocolVersion!==1||frame.generation!==1||frame.channel!==2||frame.streamId!==id)throw new Error('终端消息身份或连接代次不匹配')
              bytes.receive(frame.sequence||0,frame.payload!.byteLength)
              return decodeTerminalEvents(frame.payload)
            })
            await model.receiveBatch(events)
            if(generation!==connection||disposed)return
            send(MessageType.Ack,undefined,group.at(-1)!.frame.sequence,size)
          }else await process(first.frame)
        }finally{queuedBytes-=size}
      }
    }catch(e){fail(e)}finally{
      processing=false
      if(!canReconnect||generation!==connection||disposed){pending.length=0;queuedBytes=0}
    }
  }
  ws.onmessage=event=>{
    if(generation!==connection||disposed||!canReconnect)return
    try{
      if(!(event.data instanceof ArrayBuffer))throw new Error('终端管理通道只接受Protobuf二进制消息')
      const frame=decodeEnvelope(new Uint8Array(event.data)),size=frame.payload?.byteLength||0;queuedBytes+=size
      if(queuedBytes>INTERACTIVE_WINDOW+65536)throw new Error('终端消费者缓冲超过上限，已关闭当前流')
      pending.push({frame,size});void drain()
    }catch(e){fail(e)}
  }
  ws.onclose=()=>{if(generation!==connection||disposed)return;connected.value=false;writable.value=false;model.setLease(false);if(!error.value)status.value='连接中断，正在重新核对原会话';if(canReconnect)reconnectTimer=setTimeout(()=>void attach(id),3000)}
  ws.onerror=()=>{if(generation===connection)error.value='无法连接终端流，请检查节点与授权；原会话和检查点保留'}
  resize=new ResizeObserver(requestSize);resize.observe(container.value)
}
function requestSize(){
  if(resizeFrame)return
  resizeFrame=requestAnimationFrame(()=>{resizeFrame=0;if(!props.visible||!connected.value||!writable.value||!terminal)return;const size=terminal.proposedSize();if(!size||size.cols<2||size.rows<1)return;if(size.cols===terminal.terminal.cols&&size.rows===terminal.terminal.rows)return;try{send(MessageType.Resize,encodeJSON({cols:Math.min(size.cols,1000),rows:Math.min(size.rows,1000)}))}catch(e){error.value=String(e)}})
}
function takeover(){try{send(MessageType.Open,encodeJSON({takeover:true}));status.value='正在请求输入和尺寸控制权'}catch(e){error.value=String(e)}}
function anotherView(){desktop.open({appId:'blora.terminal',resourceRef:view.value.resourceRef,title:view.value.title,disposition:'new-window',state:{sessionId:sessionId.value,containerId:hostContainer.value,containerTarget:view.value.state.containerTarget||null}})}
function askEnd(){endRequest.value={sessionId:sessionId.value,requestId:crypto.randomUUID()};endDialog.value=true}
async function endSession(){
  const fixed=endRequest.value;if(!fixed||busy.value)return;busy.value=true;error.value='';endRequest.value={...fixed,state:'SUBMITTING'}
  try{const {task}=await api<{task:Task}>(`/terminals/${encodeURIComponent(fixed.sessionId)}/close`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId},body:'{}'});endRequest.value={...fixed,taskId:task.taskId,state:task.state};endDialog.value=false;await trackEnd()}catch(e){error.value=String(e)}finally{busy.value=false}
}
async function trackEnd(){
  if(endTimer)clearTimeout(endTimer);const fixed=endRequest.value;if(!fixed||disposed)return
  try{
    const task=fixed.taskId?(await api<{task:Task}>(`/tasks/${encodeURIComponent(fixed.taskId)}`)).task:(await api<{items:Task[]}>('/tasks')).items.find(task=>task.requestId===fixed.requestId)
    if(!task){error.value='结束会话的提交回执尚未确认；可在确认框显式重试同一请求';return}
    if(disposed)return;endRequest.value={...fixed,taskId:task.taskId,state:task.state}
    if(task.state==='SUCCEEDED'){
      if(sessionId.value===fixed.sessionId){connection++;if(reconnectTimer)clearTimeout(reconnectTimer);socket?.close();socket=undefined;resize?.disconnect();writable.value=false;connected.value=false;terminal?.setLease(false);status.value='会话已结束 · 屏幕检查点保留'}
      await refreshSessions()
    }else if(['FAILED','CANCELLED','INTERRUPTED'].includes(task.state))error.value=task.error||`结束会话任务${task.state}`
    else endTimer=setTimeout(()=>void trackEnd(),1500)
  }catch(e){error.value=String(e);if(!disposed)endTimer=setTimeout(()=>void trackEnd(),3000)}
}
function zoom(delta:number){const value=Math.max(9,Math.min(28,fontSize.value+delta));patch('fontSize',value);terminal?.setFontSize(value);requestSize()}
async function copySelection(){const text=terminal?.terminal.getSelection();if(!text)return;try{await navigator.clipboard.writeText(text)}catch(e){error.value='浏览器拒绝复制：'+String(e)}}
watch(()=>props.visible,visible=>{if(visible)requestSize()})
watch(uiTheme,mode=>{void terminal?.setColorMode(mode).catch(e=>{error.value='终端主题恢复保护失败：'+String(e)})})
const notifyPaintHost=()=>container.value?.dispatchEvent(new CustomEvent('terminal-paint-host-changed',{bubbles:true}))
onMounted(()=>{notifyPaintHost();void load()})
onBeforeUnmount(()=>{notifyPaintHost();disposed=true;if(taskTimer)clearTimeout(taskTimer);if(endTimer)clearTimeout(endTimer);stopConnection()})
</script>
<template>
  <div class="terminal-app">
    <div class="editor-toolbar"><strong>{{status}}</strong><button @click="load">刷新会话</button><button class="primary" :disabled="!view.resourceRef || busy || !!pendingTask" @click="create">新建会话</button><button v-if="sessionId" @click="attach(sessionId)">重新挂载</button><button v-if="connected && !writable" @click="takeover">申请接管</button><button @click="zoom(-1)" aria-label="缩小终端字号">A−</button><button @click="zoom(1)" aria-label="放大终端字号">A＋</button><button @click="copySelection">复制选区</button></div>
    <p v-if="error" class="error notice" role="alert">{{error}}</p>
    <div v-if="sessionId" class="terminal-sessionbar"><select :value="sessionId" aria-label="选择终端会话" @change="attach(($event.target as HTMLSelectElement).value)"><option v-for="s in sessions" :key="s.sessionId" :value="s.sessionId">{{s.sessionId}} · {{s.state}}</option></select><button @click="anotherView">同会话新窗口</button><button :disabled="!!endRequest?.taskId&&!['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(endRequest.state||'')||selected?.state==='closed'" @click="askEnd">结束会话…</button><span v-if="endRequest?.taskId">结束任务：{{endRequest.state}}</span></div>
    <div v-if="!sessionId" class="terminal-sessions"><p class="muted">刷新与移窗只恢复原会话，不重发历史输入。创建会话通过后台任务执行。</p><button v-for="s in sessions" :key="s.sessionId" @click="attach(s.sessionId)">{{s.sessionId}} · {{s.state}} · {{s.backend}}</button></div>
    <div class="terminal-paint-region"><div class="terminal-paint-viewport"><div ref="container" class="terminal-container" :data-session-id="sessionId" :data-terminal-writable="writable"></div></div></div>
    <footer class="editor-status"><span>{{view.resourceRef?.nodeId}} · {{sessionId || '未选择会话'}}</span><span v-if="selected">{{selected.backend}} · 会话上限 {{selected.maxSessions}} · 归档 {{Math.round(selected.archive.maxBytes/1024/1024)}} MiB</span></footer>
    <form v-if="endDialog&&endRequest" class="in-window-dialog" role="dialog" aria-label="结束终端会话" @submit.prevent="endSession"><h3>结束终端会话</h3><p class="selectable">{{endRequest.sessionId}}</p><p>此操作结束这个交互会话、关联 Shell 及其前台程序。独立托管的实例仍按原状态运行。</p><p>仅关闭窗口可以保留会话以便稍后继续。</p><div class="action-row"><button type="button" @click="endDialog=false">返回</button><button class="primary" :disabled="busy">确认结束会话</button></div></form>
  </div>
</template>
