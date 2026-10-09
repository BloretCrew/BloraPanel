<script setup lang="ts">
import { randomUUID } from '../services/uuid'
import {computed,ref} from 'vue'
import {useNow} from '@vueuse/core'
import {useQuery} from '@tanstack/vue-query'
import {api,session,type Task} from '../services/api'
import {readTaskQuery} from '../services/task-query-lifecycle'
import {useDesktop} from '../desktop/store'
import TransferDetails from './TransferDetails.vue'
import TaskStages from './TaskStages.vue'
import ComposeTaskOutput from './ComposeTaskOutput.vue'
import {taskElapsed,taskTimestamp} from '../services/task-time'
import {restoreUploadFromTask} from '../services/uploads'
const now=useNow({interval:1000})
const props=defineProps<{viewTabId:string}>(),desktop=useDesktop(),error=ref(''),cancelling=ref<string>()
const resuming=ref<string>()
const view=computed(()=>desktop.state!.views[props.viewTabId]!),filter=computed({get:()=>String(view.value.state.filter||''),set:value=>desktop.commit([{kind:'set',path:['views',props.viewTabId,'state','filter'],value},{kind:'set',path:['views',props.viewTabId,'state','taskHistoryBefore'],value:0},{kind:'set',path:['views',props.viewTabId,'state','taskHistoryStack'],value:[]}])})
const before=computed(()=>Number(view.value.state.taskHistoryBefore||0)),history=computed<number[]>(()=>view.value.state.taskHistoryStack as number[]||[])
const tasks=useQuery({queryKey:computed(()=>['tasks','view',view.value.resourceRef?.id||'',before.value,filter.value]),queryFn:({signal})=>readTaskQuery(signal,async()=>{
  const resource=view.value.resourceRef
  if(resource){
    const result=await api<{task?:Task}>(`/tasks/${encodeURIComponent(resource.id)}`,{signal})
    if(result.task?.taskId)return {items:[result.task],nextBefore:-1}
    // Keep a malformed or stale detail response from poisoning the reactive
    // list with `undefined`; the history endpoint is also authoritative for
    // the same user and lets a lost detail response recover the task.
    const history=await api<{items:Task[];nextBefore:number}>(`/tasks?before=0&limit=1000&state=`,{signal})
    const task=history.items.find(item=>item?.taskId===resource.id)
    if(task)return {items:[task],nextBefore:-1}
    throw new Error('任务不存在或未授权')
  }
  return api<{items:Task[];nextBefore:number}>(`/tasks?before=${before.value}&limit=100&state=${encodeURIComponent(filter.value)}`,{signal})
}),refetchInterval:2000})
const nextBefore=computed(()=>tasks.data.value?.nextBefore??-1)
const canReadEarlier=computed(()=>nextBefore.value>=0&&!tasks.isFetching.value)
function navigateHistory(direction:'next'|'previous'|'first'){
  let target=0,stack=[...history.value]
  if(direction==='next'){if(nextBefore.value<0)return;target=nextBefore.value;stack.push(before.value)}
  else if(direction==='previous'){target=stack.pop()||0}else stack=[]
  desktop.commit([{kind:'set',path:['views',props.viewTabId,'state','taskHistoryBefore'],value:target},{kind:'set',path:['views',props.viewTabId,'state','taskHistoryStack'],value:stack}])
}
const items=computed(()=>tasks.data.value?.items.filter((t):t is Task=>!!t?.taskId).filter(t=>(!filter.value||t.state===filter.value)&&(!view.value.resourceRef||t.taskId===view.value.resourceRef.id))||[])
interface CancelAttempt {requestId:string;error?:string}
const cancelAttempts=computed<Record<string,CancelAttempt>>(()=>view.value.state.cancelAttempts as unknown as Record<string,CancelAttempt>||{})
async function cancel(task:Task){
  error.value='';cancelling.value=task.taskId
  const fixed=cancelAttempts.value[task.taskId]||{requestId:randomUUID()}
  desktop.patchView(props.viewTabId,'cancelAttempts',{...cancelAttempts.value,[task.taskId]:{...fixed,error:undefined}})
  try{await api(`/tasks/${encodeURIComponent(task.taskId)}/cancel`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId}});await tasks.refetch()}
  catch(e){const message=String(e);error.value=message;desktop.patchView(props.viewTabId,'cancelAttempts',{...cancelAttempts.value,[task.taskId]:{...fixed,error:message}})}
  finally{cancelling.value=undefined}
}
interface RetryAttempt {requestId:string;taskId?:string;error?:string}
const retryAttempts=computed<Record<string,RetryAttempt>>(()=>view.value.state.retryAttempts as unknown as Record<string,RetryAttempt>||{})
const retryable=(task:Task)=>['instance.start','instance.stop','instance.restart','instance.kill'].includes(task.action)&&['FAILED','INTERRUPTED'].includes(task.state)
async function retry(task:Task){
  if(!retryable(task))return
  const fixed=retryAttempts.value[task.taskId]||{requestId:randomUUID()}
  desktop.patchView(props.viewTabId,'retryAttempts',{...retryAttempts.value,[task.taskId]:{...fixed,error:undefined}})
  try{
    const result=await api<{task:Task}>('/tasks/'+encodeURIComponent(task.taskId)+'/retry',{method:'POST',headers:{'Idempotency-Key':fixed.requestId},body:'{}'})
    desktop.patchView(props.viewTabId,'retryAttempts',{...retryAttempts.value,[task.taskId]:{...fixed,taskId:result.task.taskId,error:undefined}})
    await tasks.refetch()
  }catch(e){desktop.patchView(props.viewTabId,'retryAttempts',{...retryAttempts.value,[task.taskId]:{...fixed,error:e instanceof Error?e.message:String(e)}})}
}
function openTask(taskId:string){desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:taskId},disposition:'dedicated',title:'任务 '+taskId})}
async function resumeFile(task:Task){
  error.value='';resuming.value=task.taskId
  try{
    const record=await restoreUploadFromTask(desktop.recovery!,task)
    const slash=record.path.lastIndexOf('/')
    desktop.open({appId:'blora.files',resourceRef:{kind:'instance',id:record.instanceId,nodeId:record.nodeId},title:'上传 · '+record.sourceName,disposition:'dedicated',state:{path:slash<0?'':record.path.slice(0,slash)}})
  }catch(e){error.value=String(e)}finally{resuming.value=undefined}
}
function formatResult(value:unknown){try{return JSON.stringify(value,null,2)}catch{return String(value)}}
const notificationStatus=ref('')
async function notifications(){
  if(!globalThis.Notification){notificationStatus.value='此浏览器不支持系统通知，站内通知仍可使用';return}
  try{const permission=await Notification.requestPermission();desktop.commit([{kind:'set',path:['preferences','systemNotifications'],value:permission==='granted'}]);notificationStatus.value=permission==='granted'?'系统通知已启用':'系统通知未获授权，站内通知仍可使用'}catch(e){notificationStatus.value=String(e)}
}
function disableNotifications(){desktop.commit([{kind:'set',path:['preferences','systemNotifications'],value:false}]);notificationStatus.value='系统通知已关闭'}
const labels:Record<string,string>={QUEUED:'排队中',RUNNING:'执行中',WAITING_NODE:'等待节点',WAITING_CLIENT:'等待本机文件',CANCEL_REQUESTED:'正在取消',SUCCEEDED:'成功',FAILED:'失败',CANCELLED:'已取消',INTERRUPTED:'已中断'}
</script>
<template>
  <div class="application">
    <header class="app-heading"><div><span class="eyebrow">ACTIVITY</span><h2>任务中心</h2><p>接受、执行与结果分别记录，关闭视图不取消任务。</p></div></header>
    <div class="filterbar"><select v-model="filter" aria-label="任务状态筛选"><option value="">全部状态</option><option v-for="(label,state) in labels" :key="state" :value="state">{{label}}</option></select><button @click="tasks.refetch()">↻ 刷新</button><button @click="notifications()">启用系统通知</button><button v-if="desktop.state?.preferences.systemNotifications" @click="disableNotifications">关闭系统通知</button></div>
    <p v-if="notificationStatus" role="status">{{notificationStatus}}</p>
    <nav v-if="!view.resourceRef" class="action-row" aria-label="任务历史分页"><button :disabled="!before||tasks.isFetching.value" @click="navigateHistory('first')">最新任务</button><button :disabled="!history.length||tasks.isFetching.value" @click="navigateHistory('previous')">较新一页</button><button :disabled="!canReadEarlier" @click="navigateHistory('next')">更早任务</button></nav>
    <p v-if="tasks.error.value || error" class="error">{{tasks.error.value?.message || error}}</p>
    <article v-for="task in items" :key="task.taskId" class="task-card">
      <div><h3>{{task.action || task.taskId}}</h3><span class="status-chip" :data-status="task.state">{{labels[task.state]||task.state}}</span></div><p>{{task.phase}}</p>
      <small class="selectable">{{task.resource?.nodeId}} · {{task.resource?.id}} · {{task.taskId}}</small>
      <p class="task-timing">发起者：{{task.actorId||'未记录'}} · 自接受起耗时（含等待）：{{taskElapsed(task,now.getTime())}}</p>
      <small>接受：{{taskTimestamp(task.createdAt)}} · 最近更新：{{taskTimestamp(task.updatedAt)}} · 派发：{{taskTimestamp(task.dispatchedAt)}}</small>
      <button v-if="!view.resourceRef" @click="openTask(task.taskId)">查看任务详情</button>
      <small v-if="task.retryOf" class="muted">关联重试自 {{task.retryOf}}</small>
      <p v-if="task.cancellationRequested && !['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state)" class="warning">已请求取消后续工作，等待节点确认。已发生的变更不会自动撤销。</p>
      <p v-if="task.error" class="error">{{task.error}}</p>
      <p v-if="task.action==='image.pull'&&task.result?.layer&&typeof task.result.current==='number'" class="task-layer-progress">当前层 {{task.result.layer}}：{{task.result.current}} / {{task.result.total||'总量未知'}} · {{task.result.message}}</p>
      <details v-if="task.result"><summary>查看已确认结果</summary><pre class="selectable task-result">{{formatResult(task.result)}}</pre></details>
      <TransferDetails v-if="task.action==='transfer.copy'||task.action==='transfer.move'" :task="task" />
      <TaskStages v-if="view.resourceRef" :task-id="task.taskId" :view-tab-id="props.viewTabId" />
      <ComposeTaskOutput v-if="view.resourceRef&&task.resource.nodeId&&['compose.apply','compose.delete'].includes(task.action)" :node-id="task.resource.nodeId" :task-id="task.taskId" :view-tab-id="props.viewTabId" />
      <button v-if="['QUEUED','RUNNING','WAITING_NODE','WAITING_CLIENT'].includes(task.state)" :disabled="task.cancellationRequested || cancelling===task.taskId" @click="cancel(task)">{{task.cancellationRequested?'取消请求已接受':'请求取消后续工作'}}</button>
      <button v-if="task.action==='file.upload'&&task.state==='WAITING_CLIENT'&&task.actorId===session.user?.userId" :disabled="task.cancellationRequested||resuming===task.taskId" @click="resumeFile(task)">继续本机上传</button>
      <template v-if="retryable(task)"><button v-if="retryAttempts[task.taskId]?.taskId" @click="openTask(retryAttempts[task.taskId]!.taskId!)">查看重试任务</button><button v-else @click="retry(task)">重试此实例操作</button><p v-if="retryAttempts[task.taskId]?.error" class="error">{{retryAttempts[task.taskId]!.error}}</p></template>
    </article>
    <div v-if="!items.length" class="empty-state"><span>✓</span><h3>暂无符合筛选的任务</h3><p>实际操作提交后会显示在这里。</p></div>
  </div>
</template>
