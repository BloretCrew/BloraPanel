<script setup lang="ts">
import {computed,ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {api,type Task} from '../services/api'
import {useDesktop} from '../desktop/store'
import {copy,json} from '../recovery/state'
export interface BatchTarget {instanceId:string;nodeId:string;name:string;requestId:string;taskId?:string;error?:string}
interface Batch {action:string;targets:BatchTarget[]}
const props=defineProps<{viewTabId:string}>(),desktop=useDesktop(),recovery=desktop.recovery!,busy=ref(false)
const batch=computed<Batch|undefined>(()=>recovery.state.views[props.viewTabId]?.state.batchConfirmation as unknown as Batch|undefined)
const tasks=useQuery({queryKey:['tasks'],queryFn:()=>api<{items:Task[]}>('/tasks'),refetchInterval:2000})
function patch(value?:Batch){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state','batchConfirmation'],value:json(value||null)}])}
async function submit(){if(!batch.value||busy.value)return;busy.value=true;const fixed=copy(batch.value);try{for(const target of fixed.targets){if(target.taskId)continue;try{const known=tasks.data.value?.items.find(task=>task.requestId===target.requestId);const task=known||(await api<{task:Task}>(`/instances/${encodeURIComponent(target.instanceId)}/actions`,{method:'POST',headers:{'Idempotency-Key':target.requestId},body:JSON.stringify({action:fixed.action})})).task;target.taskId=task.taskId;target.error=undefined}catch(e){target.error=String(e)}patch(copy(fixed))}await tasks.refetch()}finally{busy.value=false}}
</script>
<template><div v-if="batch" class="in-window-dialog" role="dialog" aria-label="确认批量实例操作"><h3>批量 {{batch.action}}</h3><p>以下资源在发起时已固定，每个实例独立排队和校验权限。接受任务不代表执行成功，其他窗口仍可操作。</p><div v-for="target in batch.targets" :key="target.instanceId" class="batch-target"><strong>{{target.name}}</strong><small>{{target.nodeId}} / {{target.instanceId}}</small><p v-if="target.taskId">{{tasks.data.value?.items.find(task=>task.taskId===target.taskId)?.state||'已接受，待核对任务'}}<button @click="desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:target.taskId},title:target.name+' · '+batch!.action,disposition:'dedicated'})">查看此任务</button></p><p v-if="target.error" class="error">{{target.error}}</p></div><p class="muted">刷新仅恢复确认与已接受的任务；再次提交只核对尚未取得回执的原请求。</p><div class="action-row"><button :disabled="busy" @click="patch()">关闭批量确认</button><button class="primary" :disabled="busy||batch.targets.every(target=>target.taskId)" @click="submit">{{busy?'逐项提交中…':'确认提交未接受项'}}</button></div></div></template>
