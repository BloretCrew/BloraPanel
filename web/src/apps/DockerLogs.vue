<script setup lang="ts">
import {onMounted,onBeforeUnmount,ref} from 'vue'
import {api} from '../services/api'
import {decodeEnvelope,decodeJSON,encodeEnvelope,MessageType,ByteWindow} from '../services/protocol'

interface LogFrame {stream:string;data:string}
interface LogWindow {containerId:string;observedAt:string;frames:LogFrame[];truncated:boolean;possibleGap?:boolean}
interface LogHistory {items:LogWindow[];retention:number;possibleGap:boolean}

const props=defineProps<{nodeId:string;containerId:string}>()
const text=ref(''),status=ref('正在连接'),error=ref(''),historyError=ref('')
const history=ref<LogWindow[]>([]),historyRetention=ref(100),historyMode=ref(false),historyLoading=ref(false)
let socket:WebSocket|undefined,disposed=false

function decodeFrame(data:string){return new TextDecoder().decode(Uint8Array.from(atob(data),c=>c.charCodeAt(0)))}
function closeSocket(){const current=socket;socket=undefined;current?.close()}
function connect(){
  if(disposed)return
  closeSocket();historyMode.value=false;error.value='';status.value='正在连接'
  const ws=new WebSocket(`${location.protocol==='https:'?'wss:':'ws:'}//${location.host}/api/v1/nodes/${encodeURIComponent(props.nodeId)}/docker/containers/${encodeURIComponent(props.containerId)}/logs?tail=200`)
  socket=ws;ws.binaryType='arraybuffer';const bytes=new ByteWindow()
  ws.onmessage=event=>{try{
    const f=decodeEnvelope(new Uint8Array(event.data))
    if(f.protocolVersion!==1||f.generation!==1||f.channel!==2||f.streamId!==props.containerId)throw new Error('日志连接身份不匹配')
    if(f.type===MessageType.Data){
      bytes.receive(f.sequence||0,f.payload?.length||0);const window=decodeJSON<LogWindow>(f.payload)
      if(window.containerId!==props.containerId)throw new Error('日志资源身份不匹配')
      text.value=window.frames.map(frame=>decodeFrame(frame.data)).join('')
      status.value=`观察时间 ${new Date(window.observedAt).toLocaleString()}${window.truncated?' · 窗口已截断':''}`
      ws.send(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:props.containerId,type:MessageType.Ack,sequence:f.sequence,credit:f.payload?.length||0}))
    }else if(f.type===MessageType.Error)throw new Error(decodeJSON<{message:string}>(f.payload).message)
  }catch(e){error.value=String(e);ws.close()}}
  ws.onclose=()=>{if(socket===ws){socket=undefined;status.value='连接已关闭；最后观察窗口保留'}}
  ws.onerror=()=>{if(socket===ws)error.value='日志连接失败，请核对节点、权限和 Docker 日志驱动'}
}
async function loadHistory(){
  historyLoading.value=true;historyError.value='';error.value='';closeSocket();status.value='正在读取归档';historyMode.value=true
  try{const result=await api<LogHistory>(`/nodes/${encodeURIComponent(props.nodeId)}/docker/containers/${encodeURIComponent(props.containerId)}/logs/history?limit=100`);history.value=result.items||[];historyRetention.value=result.retention||100;status.value=`已载入 ${history.value.length} 个归档窗口`}
  catch(e){history.value=[];historyError.value=String(e);status.value='归档读取失败'}
  finally{historyLoading.value=false}
}
function showLive(){historyMode.value=false;historyError.value='';connect()}
function renderWindow(window:LogWindow){return window.frames.map(frame=>decodeFrame(frame.data)).join('')}
onMounted(connect);onBeforeUnmount(()=>{disposed=true;closeSocket()})
</script>
<template>
  <section class="docker-logs" :data-log-mode="historyMode?'history':'live'">
    <header class="docker-log-toolbar"><div><p>{{status}}</p><p class="muted">实时窗口会整体替换；Docker 日志驱动可能重叠或存在保留缺口。</p></div><div class="action-row"><button v-if="!historyMode" :disabled="historyLoading" @click="loadHistory">{{historyLoading?'读取归档…':'查看归档'}}</button><button v-else @click="showLive">返回实时日志</button></div></header>
    <p v-if="error" class="error" role="alert">{{error}}</p><p v-if="historyError" class="error" role="alert">{{historyError}}</p>
    <template v-if="historyMode">
      <p class="muted">保留上限 {{historyRetention}} 个窗口。归档只用于诊断，窗口之间可能存在缺口。</p><p v-if="!history.length&&!historyError" class="empty-state">暂无已归档日志窗口。</p>
      <article v-for="(window,index) in history" :key="`${window.observedAt}-${index}`" class="docker-log-window"><header><strong>{{new Date(window.observedAt).toLocaleString()}}</strong><span v-if="window.truncated||window.possibleGap" class="muted">{{window.truncated?'窗口已截断':''}}{{window.truncated&&window.possibleGap?' · ':''}}{{window.possibleGap?'可能存在缺口':''}}</span></header><pre class="docker-log selectable">{{renderWindow(window)}}</pre></article>
    </template>
    <pre v-else class="docker-log selectable">{{text}}</pre>
  </section>
</template>
<style scoped>
.docker-log-toolbar{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}
.docker-log-window{margin:12px 0;padding:10px 12px;border:1px solid var(--line);border-radius:var(--shape-md);background:var(--surface-low)}
.docker-log-window header{display:flex;justify-content:space-between;gap:10px;align-items:baseline;margin-bottom:8px}
.docker-log{white-space:pre-wrap;overflow:auto;max-height:420px;background:var(--inverse-surface);color:var(--inverse-on-surface);padding:14px;user-select:text}
</style>
