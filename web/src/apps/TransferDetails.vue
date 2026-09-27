<script setup lang="ts">
import {computed,ref,watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {api,type Task} from '../services/api'
interface Transfer {stage:string;entries:number;completed:number;total:number;committedBytes:number;currentPath?:string;currentOffset:number;destinationVerified:boolean;sourceDeleted:boolean;sourceOutcomeUnknown:boolean;partial:boolean;cleanupPending:boolean;error?:string}
interface Entry {relative:string;kind:string;size:number;status:string}
const props=defineProps<{task:Task}>(),offset=ref(0)
const detail=useQuery({queryKey:computed(()=>['transfer',props.task.taskId]),queryFn:()=>api<{task:Task;transfer:Transfer}>(`/transfers/${encodeURIComponent(props.task.taskId)}`),refetchInterval:()=>['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(props.task.state)?false:2000})
const entries=useQuery({queryKey:computed(()=>['transfer-entries',props.task.taskId,offset.value,detail.data.value?.transfer.completed,detail.data.value?.transfer.stage]),queryFn:()=>api<{items:Entry[];total:number;nextOffset:number}>(`/transfers/${encodeURIComponent(props.task.taskId)}/entries?offset=${offset.value}&limit=25`)})
const progress=computed(()=>detail.data.value?.transfer)
watch(()=>props.task.revision,()=>{void detail.refetch();void entries.refetch()})
const stages:Record<string,string>={preflight:'校验源、目标与目录关系',copy:'复制到目标',verify_destination:'验证目标内容',metadata:'应用文件元信息',verify_source:'再次验证源未改变',delete_source:'删除源',complete:'处理已完成',cleanup:'清理未提交内容',cleaned:'未提交内容已清理'}
</script>
<template>
  <section class="transfer-details" aria-label="文件传输阶段">
    <p v-if="detail.error.value||entries.error.value" class="error">{{detail.error.value?.message||entries.error.value?.message}}</p>
    <template v-if="progress"><h4>{{stages[progress.stage]||progress.stage}}</h4><progress :value="progress.committedBytes" :max="progress.total||1"></progress><p>已提交 {{progress.committedBytes}} / {{progress.total}} B · 已完成 {{progress.completed}} / {{progress.entries}} 项</p><p v-if="progress.currentPath" class="selectable">当前：{{progress.currentPath}} · 节点确认 {{progress.currentOffset}} B</p>
      <p>{{progress.destinationVerified?'目标内容已验证':'目标整体尚未验证'}} · {{progress.sourceDeleted?'源已删除':'未确认删除源'}}</p><p v-if="task.action==='transfer.move'" class="muted">移动依次复制、验证目标、重核源并删除源；期间可能同时存在两份内容。</p>
      <p v-if="progress.partial" class="warning">部分目标内容已经提交，取消不会自动撤销已提交文件。</p><p v-if="progress.sourceOutcomeUnknown" class="warning">源删除结果不明；必须核对源，不能认定源仍在或移动成功。</p><p v-if="progress.cleanupPending" class="warning">临时内容仍待节点清理。</p><p v-if="progress.error" class="error">{{progress.error}}</p>
    </template>
    <div v-for="entry in entries.data.value?.items" :key="entry.relative" class="task-row"><span class="selectable">{{entry.relative||'.'}}</span><span>{{entry.kind}} · {{entry.size}} B</span><span>{{entry.status}}</span></div>
    <div class="action-row"><button :disabled="offset===0" @click="offset=Math.max(0,offset-25)">上一页传输条目</button><small>{{entries.data.value?.total||0}} 项</small><button :disabled="!entries.data.value||entries.data.value.nextOffset<0" @click="offset=entries.data.value!.nextOffset">下一页传输条目</button><button @click="detail.refetch();entries.refetch()">核对传输结果</button></div>
  </section>
</template>
