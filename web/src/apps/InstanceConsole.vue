<script setup lang="ts">
import StyledSelect from '../app-host/StyledSelect.vue'
import { randomUUID } from '../services/uuid'
import {computed,nextTick,onBeforeUnmount,onMounted,ref,watch} from 'vue'
import {useDesktop} from '../desktop/store'
import {api,session,type Instance,type Task} from '../services/api'
import {ByteWindow,decodeEnvelope,decodeJSON,encodeEnvelope,INTERACTIVE_WINDOW,MessageType,type Envelope} from '../services/protocol'
import {appendLog,emptyLog,logGap,logText,type LogCheckpoint,type LogEvent} from '../services/logs'
import {json} from '../recovery/state'
interface RunLog {runId:string;startedAt:string;backend:string;diagnostic?:string}
interface LogStatus {phase:string;earliest:number;latest:number;bytes:number;diagnostic?:string}
interface CommandAttempt {instanceId:string;runId:string;requestId:string;data:string;draft:string;taskId?:string;state:string;error?:string;bytesWritten?:number}
const props=defineProps<{viewTabId:string;instance:Instance;visible:boolean}>(),desktop=useDesktop(),recovery=desktop.recovery!,runs=ref<RunLog[]>([]),error=ref(''),status=ref('正在核对运行日志'),connected=ref(false),inputAvailable=ref(false),sending=ref(false),viewport=ref<HTMLElement>(),height=ref(280)
const view=computed(()=>recovery.state.views[props.viewTabId]!)
function patch(key:string,value:unknown){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
const runId=computed(()=>String(view.value.state.logRunId||'')),checkpoint=computed<LogCheckpoint>(()=>{const saved=view.value.state.logCheckpoint as unknown as LogCheckpoint|undefined;return saved?.runId===runId.value?saved:emptyLog(runId.value)})
const paused=computed({get:()=>!!view.value.state.logPaused,set:value=>patch('logPaused',value)}),search=computed({get:()=>String(view.value.state.logSearch||''),set:value=>{patch('logSearch',value);patch('logScroll',0)}}),commandDraft=computed({get:()=>String(view.value.state.consoleCommand||''),set:value=>patch('consoleCommand',value)})
const newline=computed({get:()=>view.value.state.consoleNewline!==false,set:value=>patch('consoleNewline',value)}),attempt=computed(()=>view.value.state.consoleAttempt as unknown as CommandAttempt|undefined)
const commandPending=computed(()=>!!attempt.value&&!['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(attempt.value.state))
const reads=new AbortController()
let refreshing=false,reconciling=false
function storeAttempt(value:CommandAttempt){if(!attempt.value||attempt.value.requestId===value.requestId)patch('consoleAttempt',value)}
function commandResult(fixed:CommandAttempt,task:Task){storeAttempt({...fixed,taskId:task.taskId,state:task.state,error:task.error,bytesWritten:typeof task.result?.bytesWritten==='number'?task.result.bytesWritten:undefined});if(task.state==='SUCCEEDED'&&commandDraft.value===fixed.draft)commandDraft.value=''}
async function reconcileCommand(){
  const fixed=attempt.value
  if(!fixed||sending.value||disposed||reconciling)return
  reconciling=true
  try{
    const task=fixed.taskId?(await api<{task:Task}>(`/tasks/${encodeURIComponent(fixed.taskId)}`,{signal:reads.signal})).task:(await api<{items:Task[]}>('/tasks',{signal:reads.signal})).items.find(task=>task.requestId===fixed.requestId&&task.action==='console.input')
    if(disposed)return
    if(task)commandResult(fixed,task)
    else storeAttempt({...fixed,state:'UNKNOWN',error:'尚未查到接收记录。可显式使用原请求核对；刷新不会再次发送。'})
  }catch(e){if(!disposed)storeAttempt({...fixed,error:String(e)})}
  finally{reconciling=false}
}
async function sendCommand(retry=false){
  if(sending.value||(!retry&&(!inputAvailable.value||!commandDraft.value||commandPending.value)))return
  const fixed:CommandAttempt=retry?{...attempt.value!}:{instanceId:props.instance.instanceId,runId:runId.value,requestId:randomUUID(),data:commandDraft.value+(newline.value?'\n':''),draft:commandDraft.value,state:'SUBMITTING'}
  const size=new TextEncoder().encode(fixed.data).length;if(size<1||size>8192){error.value='实例输入需要 1～8192 个 UTF-8 字节（包含所选换行）';return}
  patch('consoleAttempt',fixed);sending.value=true
  try{const {task}=await api<{task:Task}>(`/instances/${encodeURIComponent(fixed.instanceId)}/logs/${encodeURIComponent(fixed.runId)}/input`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId},body:JSON.stringify({data:fixed.data})});commandResult(fixed,task)}catch(e){storeAttempt({...fixed,state:'UNKNOWN',error:String(e)})}finally{sending.value=false}
}
const scroll=computed(()=>Number(view.value.state.logScroll||0)),allLines=computed(()=>logText(checkpoint.value.text).split('\n').map((text,index)=>({text,number:index+checkpoint.value.discardedLines+1}))),lines=computed(()=>search.value?allLines.value.filter(line=>line.text.toLowerCase().includes(search.value.toLowerCase())):allLines.value),first=computed(()=>Math.max(0,Math.floor(scroll.value/20)-4)),visibleLines=computed(()=>lines.value.slice(first.value,first.value+Math.min(100,Math.ceil(height.value/20)+10)))
let socket:WebSocket|undefined,generation=0,disposed=false,reconnect:ReturnType<typeof setTimeout>|undefined,refresh:ReturnType<typeof setInterval>|undefined,observer:ResizeObserver|undefined,autoScroll=false
async function refreshRuns(){
  if(disposed||refreshing)return
  refreshing=true
  try{
    const {items}=await api<{items:RunLog[]}>(`/instances/${encodeURIComponent(props.instance.instanceId)}/logs`,{signal:reads.signal})
    if(disposed)return
    runs.value=items.sort((a,b)=>Date.parse(b.startedAt)-Date.parse(a.startedAt))
    if(!runId.value){const initial=items.find(run=>run.runId===props.instance.runId)||items[0];if(initial)attach(initial.runId);else status.value='此实例尚无运行日志'}
  }catch(e){if(!disposed)error.value=String(e)}
  finally{refreshing=false}
}
function stop(){generation++;clearTimeout(reconnect);socket?.close();socket=undefined;connected.value=false;inputAvailable.value=false}
function persist(next:LogCheckpoint){const removed=next.discardedLines-checkpoint.value.discardedLines;patch('logCheckpoint',next);if(paused.value&&removed>0)patch('logScroll',Math.max(0,scroll.value-removed*20))}
async function attach(id:string){
  stop();const connection=generation,changed=runId.value!==id;patch('logRunId',id);error.value='';status.value='正在重新读取此运行代次';if(changed){persist(emptyLog(id));patch('logScroll',0)}
  const protocol=location.protocol==='https:'?'wss:':'ws:',ws=new WebSocket(`${protocol}//${location.host}/api/v1/instances/${encodeURIComponent(props.instance.instanceId)}/logs/${encodeURIComponent(id)}/stream?sequence=${checkpoint.value.sequence}`);socket=ws;ws.binaryType='arraybuffer'
  const bytes=new ByteWindow();let queued=0,processing=Promise.resolve(),retry=true
  function send(type:MessageType,sequence=0,credit=0){if(ws.readyState!==WebSocket.OPEN)throw new Error('日志连接已中断');ws.send(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:id,type,sequence,credit}))}
  function fail(reason:unknown){retry=false;connected.value=false;inputAvailable.value=false;status.value='连接已停止 · 保留已接收日志';error.value=String(reason);ws.close(4002,'log stream failed')}
  async function process(frame:Envelope){
    if(disposed||generation!==connection||!retry)return
    if(frame.protocolVersion!==1||frame.generation!==1||frame.channel!==2||frame.streamId!==id)throw new Error('日志流身份不匹配')
    if(frame.type===MessageType.Data){bytes.receive(frame.sequence||0,frame.payload!.length);persist(appendLog(checkpoint.value,decodeJSON<LogEvent>(frame.payload)));await recovery.awaitPendingWrites();if(!recovery.status.protected)throw new Error('日志检查点未保护，已暂停接收；可复制当前保留区后重试');if(disposed||generation!==connection)return;send(MessageType.Ack,frame.sequence,frame.payload!.length)}
    else if(frame.type===MessageType.OpenAck){const info=decodeJSON<{runId:string;status:LogStatus;inputAvailable:boolean}>(frame.payload);if(info.runId!==id)throw new Error('日志应答运行代次不匹配');connected.value=true;inputAvailable.value=info.inputAvailable===true;status.value=`${info.status.phase} · ${info.status.bytes} 字节归档`;if(info.status.diagnostic)error.value=info.status.diagnostic}
    else if(frame.type===MessageType.Reset){const gap=decodeJSON<{code:string;message:string;earliest:number}>(frame.payload);if(gap.code!=='OUTPUT_GAP')throw new Error(gap.message);persist(logGap(checkpoint.value,gap.earliest));error.value=gap.message+'；保留现有日志，并继续读取可用归档'}
    else if(frame.type===MessageType.Close){const final=decodeJSON<LogStatus>(frame.payload);retry=false;connected.value=false;inputAvailable.value=false;status.value=final.phase==='complete'?'日志归档已完成':`日志归档 ${final.phase}`;if(final.diagnostic)error.value=final.diagnostic;ws.close()}
    else if(frame.type===MessageType.Error){const detail=decodeJSON<{code:string;message:string}>(frame.payload);throw new Error(`${detail.code}：${detail.message}`)}
    else if(frame.type===MessageType.Ping)send(MessageType.Pong)
  }
  ws.onmessage=event=>{try{if(!(event.data instanceof ArrayBuffer))throw new Error('日志流需要Protobuf二进制消息');const frame=decodeEnvelope(new Uint8Array(event.data)),size=frame.payload?.length||0;queued+=size;if(queued>INTERACTIVE_WINDOW+65536)throw new Error('日志接收队列超出预算');processing=processing.then(()=>process(frame)).catch(fail).finally(()=>{queued-=size})}catch(e){fail(e)}}
  ws.onerror=()=>{if(generation===connection)error.value='日志连接不可用；原保留区与游标保持不变'}
  ws.onclose=()=>{if(disposed||generation!==connection)return;connected.value=false;inputAvailable.value=false;if(retry&&session.user?.userId===recovery.userId){status.value='日志连接中断，等待重核原运行代次';reconnect=setTimeout(()=>void attach(id),3000)}}
}
function onScroll(){if(!viewport.value)return;patch('logScroll',viewport.value.scrollTop);patch('logScrollLeft',viewport.value.scrollLeft);if(!autoScroll&&viewport.value.scrollHeight-viewport.value.scrollTop-viewport.value.clientHeight>24)paused.value=true}
async function position(){await nextTick();if(!viewport.value)return;autoScroll=true;viewport.value.scrollTop=paused.value?scroll.value:viewport.value.scrollHeight;viewport.value.scrollLeft=Number(view.value.state.logScrollLeft||0);patch('logScroll',viewport.value.scrollTop);requestAnimationFrame(()=>{autoScroll=false})}
function toggleFollow(){paused.value=!paused.value;if(!paused.value)void position()}
async function copyLog(){try{await navigator.clipboard.writeText(logText(checkpoint.value.text));status.value='当前保留日志已复制'}catch(e){error.value=String(e)}}
watch(()=>lines.value.length,()=>void position());watch(search,()=>void position());watch(()=>props.visible,visible=>{if(visible)void position()});watch(()=>props.instance.runId,()=>void refreshRuns())
onMounted(()=>{void refreshRuns();void reconcileCommand();if(runId.value)void attach(runId.value);observer=new ResizeObserver(()=>{height.value=viewport.value?.clientHeight||280});if(viewport.value)observer.observe(viewport.value);void position();refresh=setInterval(()=>{void refreshRuns();if(commandPending.value)void reconcileCommand()},3000)})
onBeforeUnmount(()=>{disposed=true;reads.abort();stop();clearInterval(refresh);observer?.disconnect()})
</script>
<template>
  <section class="instance-console">
    <div class="console-toolbar"><StyledSelect :value="runId" aria-label="选择日志运行代次" @change="attach(($event.target as HTMLSelectElement).value)"><option v-if="!runs.length" value="">暂无运行记录</option><option v-for="run in runs" :key="run.runId" :value="run.runId">{{new Date(run.startedAt).toLocaleString()}} · {{run.runId}}</option></StyledSelect><button :disabled="!runId" @click="attach(runId)">重新连接日志</button><button @click="refreshRuns">刷新运行记录</button><button @click="toggleFollow">{{paused?'恢复跟随输出':'暂停跟随输出'}}</button><button @click="copyLog">复制保留日志</button></div>
    <div class="filterbar"><input v-model="search" aria-label="搜索保留日志" placeholder="搜索当前保留区"><small>{{status}}{{connected?' · 已连接':''}}</small></div>
    <p v-if="error" class="error" role="alert">{{error}}</p><p v-if="checkpoint.truncated" class="warning">浏览器仅保留最近 48000 字符 / 3000 行；更早内容仍受节点归档保留策略约束。</p>
    <div ref="viewport" class="console-log-viewport" role="log" aria-label="实例运行日志" aria-live="off" tabindex="0" @scroll="onScroll"><div class="console-log-spacer" :style="{height:Math.max(20,lines.length*20)+'px'}"><div v-for="(line,index) in visibleLines" :key="line.number" class="console-log-line" :style="{top:(first+index)*20+'px'}"><span>{{line.number}}</span><code>{{line.text||' '}}</code></div></div></div>
    <div class="console-command"><input v-model="commandDraft" aria-label="尚未发送的实例命令" placeholder="独立命令草稿（尚未发送）"><button :disabled="!inputAvailable||!commandDraft||commandPending||sending" @click="sendCommand()">发送命令</button></div><label class="check-row"><input v-model="newline" type="checkbox">发送时追加换行</label><p class="muted">{{inputAvailable?'输入固定当前运行代次，任务只确认交付标准输入；程序处理结果请查看输出。':'此运行当前仅提供日志观察；无输入权限、历史运行或输入不可用时不能发送。'}} 未发草稿会恢复。</p>
    <div v-if="attempt" class="console-input-result"><p>{{attempt.state==='SUCCEEDED'?`已交付 ${attempt.bytesWritten??'—'} 字节标准输入（不代表程序处理成功）`:attempt.state==='UNKNOWN'?'输入接收结果待核对':`输入任务 ${attempt.state}`}} · {{attempt.runId}}</p><p v-if="attempt.error" class="error">{{attempt.error}}</p><button @click="reconcileCommand">核对发送结果</button><button v-if="attempt.state==='UNKNOWN'||attempt.state==='SUBMITTING'" :disabled="sending" @click="sendCommand(true)">用原请求核对发送</button><button v-if="attempt.taskId" @click="desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:attempt.taskId},title:'实例输入',disposition:'dedicated'})">查看输入任务</button></div>
    <button class="console-shell-link" @click="desktop.open({appId:'blora.terminal',resourceRef:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},title:instance.name+' · 终端',disposition:'new-window'})">打开终端会话管理</button>
  </section>
</template>
