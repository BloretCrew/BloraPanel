<script setup lang="ts">
import { randomUUID } from '../services/uuid'
import {computed,onBeforeUnmount,ref,watch} from 'vue'
import {api,session} from '../services/api'

interface Source {repository:string;channel:'beta'|'stable';apiUrl?:string}
interface Preview {revision:string;currentRevision:string;version:string;compatible:boolean;reason?:string;upToDate:boolean;checkedAt:string;tag?:string;releaseUrl?:string;downloadBytes?:number}
interface Job {id:string;revision:string;state:string;phase:string;detail?:string;updatedAt:string;downloadedBytes?:number;totalBytes?:number}
interface Status {component:string;version:string;revision:string;platform:string;source:Source;preview?:Preview;job?:Job;checking:boolean;checkError?:string}
const props=withDefaults(defineProps<{nodeId?:string;visible?:boolean}>(),{visible:true})
const base=computed(()=>props.nodeId?`/nodes/${encodeURIComponent(props.nodeId)}/updates`:'/system/updates')
const status=ref<Status>(),preview=ref<Preview>(),repository=ref(''),channel=ref<'beta'|'stable'>('beta'),apiUrl=ref(''),error=ref(''),notice=ref(''),busy=ref(false)
const pending=ref<{revision:string;requestId:string}>(),sourceDirty=ref(false)
const enabled=computed(()=>props.visible&&session.user?.admin===true)
const active=computed(()=>!!status.value?.checking||status.value?.job?.state==='running'||status.value?.job?.state==='ready')
const phases:Record<string,string>={checking:'检查发行版',fetching:'获取发行信息',downloading:'下载发行包',verifying:'校验发行包',extracting:'准备程序文件',validating:'验证程序兼容性',preflight:'检查实例与版本适配',staging:'准备版本目录',ready:'准备切换',activating:'正在切换管理服务',complete:'更新完成',failed:'更新失败',interrupted:'更新中断'}
let timer:ReturnType<typeof setTimeout>|undefined,controller:AbortController|undefined,generation=0
function editableAPI(source:Source){
  try{const url=new URL(source.repository),parts=url.pathname.replace(/\.git\/?$/,'').split('/').filter(Boolean);if(url.hostname==='github.com'&&parts.length===2&&source.apiUrl===`https://api.github.com/repos/${parts.join('/')}`)return ''}catch{/* Display an explicit source unchanged if its address cannot be derived. */}
  return source.apiUrl||''
}
function stop(){generation++;if(timer)clearTimeout(timer);controller?.abort();controller=undefined}
async function refresh(){
  if(!enabled.value)return
  const current=generation
  controller?.abort();const read=new AbortController();controller=read
  try{
    const result=await api<Status>(props.nodeId?`${base.value}/status`:base.value,{signal:read.signal})
    if(current!==generation||!enabled.value)return
    status.value=result
    preview.value=result.preview
    if(result.preview&&notice.value.startsWith('后台正在检查发行版'))notice.value=result.preview.upToDate?'当前已是此频道的最新发行版。':'已检查指定发行版；应用时不会改为另一个版本。'
    if(!sourceDirty.value){repository.value=result.source.repository;channel.value=result.source.channel;apiUrl.value=editableAPI(result.source)}
    error.value=''
  }catch(e){if(!read.signal.aborted&&current===generation)error.value=String(e)}
  finally{if(current===generation&&enabled.value)timer=setTimeout(refresh,3000)}
}
watch([enabled,base,()=>session.user?.userId],()=>{stop();status.value=undefined;preview.value=undefined;pending.value=undefined;sourceDirty.value=false;busy.value=false;error.value='';notice.value='';if(enabled.value)void refresh()},{immediate:true})
onBeforeUnmount(stop)
async function saveSource(){
  if(!enabled.value||busy.value)return
  const current=generation
  busy.value=true;error.value='';notice.value=''
  try{const result=await api<Status>(`${base.value}/source`,{method:'PUT',body:JSON.stringify({repository:repository.value.trim(),channel:channel.value,apiUrl:apiUrl.value.trim()})});if(current!==generation)return;status.value=result;sourceDirty.value=false;preview.value=undefined;pending.value=undefined;notice.value='更新源已保存。请重新检查更新。'}catch(e){if(current===generation)error.value=String(e)}finally{if(current===generation)busy.value=false}
}
async function check(){
  if(!enabled.value||busy.value)return
  const current=generation
  busy.value=true;error.value='';notice.value=''
  try{const result=await api<Status>(`${base.value}/check`,{method:'POST'});if(current!==generation)return;status.value=result;preview.value=result.preview;pending.value=undefined;notice.value='后台正在检查发行版，结果会自动显示。应用时会固定到检查得到的版本。'}catch(e){if(current===generation)error.value=String(e)}finally{if(current===generation)busy.value=false}
}
async function apply(){
  const target=preview.value
  if(!enabled.value||busy.value||active.value||!target?.compatible||target.upToDate||sourceDirty.value)return
  if(!pending.value&&!confirm(`更新到 ${target.version}（${target.revision.slice(0,12)}）？运行中的实例继续运行。管理连接会短暂重连，终端会话可能关闭；不适配或无法保留实例时会拒绝更新。`))return
  pending.value??={revision:target.revision,requestId:randomUUID()}
  const fixed=pending.value
  const current=generation
  busy.value=true;error.value='';notice.value=''
  try{const job=await api<Job>(`${base.value}/apply`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId},body:JSON.stringify({revision:fixed.revision})});if(current!==generation)return;if(status.value)status.value.job=job;pending.value=undefined;notice.value='后台已接受更新；关闭此窗口不会取消任务。'}catch(e){if(current===generation)error.value=String(e)}finally{if(current===generation)busy.value=false}
}
</script>

<template>
  <section v-if="session.user?.admin" class="settings-panel core-updates" aria-label="系统更新">
    <h3>{{nodeId?'Daemon 更新':'Master 系统更新'}}</h3>
    <p class="muted">下载与你的平台匹配的发行版，校验后切换管理服务。运行中的实例保持运行；配置和资源数据独立保留。</p>
    <dl v-if="status" class="details-grid"><dt>当前版本</dt><dd>{{status.version||'开发版本'}} · {{status.revision?.slice(0,12)||'提交未知'}}</dd><dt>平台</dt><dd>{{status.platform}}</dd></dl>
    <form @submit.prevent="saveSource">
      <label>发行仓库<input v-model="repository" aria-label="发行仓库" type="url" required placeholder="https://github.com/BloretCrew/BloraPanel" :disabled="busy||active" @input="sourceDirty=true"></label>
      <label>更新频道<select v-model="channel" aria-label="更新频道" :disabled="busy||active" @change="sourceDirty=true"><option value="beta">包含测试发行版</option><option value="stable">仅正式发行版</option></select></label>
      <label>发行 API 镜像<input v-model="apiUrl" aria-label="发行 API 镜像" type="url" placeholder="留空使用 GitHub 官方源" :disabled="busy||active" @input="sourceDirty=true"></label>
      <small class="muted">镜像需提供兼容的发行列表与下载地址。仅使用你信任的发行仓库或镜像。</small>
      <div class="action-row"><button :disabled="busy||active||!sourceDirty||!repository.trim()">保存更新源</button><button type="button" :disabled="busy||active||sourceDirty||!status" @click="check">{{busy?'正在处理…':'检查更新'}}</button></div>
    </form>
    <div v-if="preview" class="update-preview"><p>目标：{{preview.tag||preview.version}} · {{preview.revision.slice(0,12)}}</p><p v-if="preview.downloadBytes" class="muted">下载约 {{(preview.downloadBytes/1048576).toFixed(1)}} MiB</p><p v-if="!preview.compatible" class="warning">{{preview.reason||'此更新与当前运行环境不兼容'}}</p><p v-else-if="preview.upToDate" class="muted">当前已是最新发行版。</p><button v-else class="primary" :disabled="busy||active||sourceDirty" @click="apply">{{pending?'重试同一更新请求':'应用此版本'}}</button></div>
    <p v-if="status?.job" class="update-progress" role="status">{{phases[status.job.phase]||status.job.phase}} · {{status.job.revision.slice(0,12)}}<span v-if="status.job.detail"> — {{status.job.detail}}</span></p>
    <div v-if="status?.job?.phase==='downloading'&&status.job.totalBytes" class="download-progress"><progress aria-label="发行包下载进度" :value="status.job.downloadedBytes||0" :max="status.job.totalBytes"/><small>{{((status.job.downloadedBytes||0)/1048576).toFixed(1)}} / {{(status.job.totalBytes/1048576).toFixed(1)}} MiB</small></div>
    <p v-if="status?.checking" class="muted" role="status">正在后台检查更新源，请稍候…</p><p v-if="status?.checkError" class="error" role="alert">{{status.checkError}}</p>
    <p v-if="notice" class="notice" role="status">{{notice}}</p><p v-if="error" class="error" role="alert">{{error}}</p>
  </section>
</template>
<style scoped>
.core-updates{margin-top:20px}.core-updates form{display:grid;gap:12px}.core-updates input{width:100%;box-sizing:border-box}.core-updates dd{overflow-wrap:anywhere}.update-preview{margin-top:16px}.update-progress{overflow-wrap:anywhere;color:var(--on-surface-variant)}.download-progress{display:flex;align-items:center;gap:12px}.download-progress progress{flex:1;min-width:0;accent-color:var(--primary)}
</style>
