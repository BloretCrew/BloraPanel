<script setup lang="ts">
import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {api} from '../services/api'
import {readTaskQuery} from '../services/task-query-lifecycle'
import {useDesktop} from '../desktop/store'
const props=defineProps<{taskId:string;viewTabId:string}>(),desktop=useDesktop()
interface Stage{revision:number;recordedAt:number;state:string;phase:string;error?:string;truncated:boolean}
const before=computed(()=>Number(desktop.state?.views[props.viewTabId]?.state.taskStageBefore||0))
const stages=useQuery({queryKey:computed(()=>['tasks',props.taskId,'stages',before.value]),queryFn:({signal})=>readTaskQuery(signal,()=>api<{items:Stage[];nextBefore:number}>(`/tasks/${encodeURIComponent(props.taskId)}/events?before=${before.value}&limit=50`,{signal})),refetchInterval:computed(()=>before.value?false:3000)})
const canEarlier=computed(()=>!stages.isFetching.value&&(stages.data.value?.nextBefore??-1)>=0)
function change(cursor:number){desktop.patchView(props.viewTabId,'taskStageBefore',cursor)}
</script>
<template>
  <section aria-label="任务阶段记录">
    <h3>阶段记录</h3><p class="muted">按最近记录排列，保留任务接受、状态变化和诊断。</p>
    <div class="action-row"><button :disabled="!before||stages.isFetching.value" @click="change(0)">最新阶段</button><button :disabled="!canEarlier" @click="change(stages.data.value!.nextBefore)">更早阶段</button><button @click="stages.refetch()">刷新阶段</button></div>
    <p v-if="stages.error.value" role="alert" class="error">{{stages.error.value.message}}</p>
    <ol v-else class="task-stages"><li v-for="stage in stages.data.value?.items" :key="stage.revision"><small>{{new Date(stage.recordedAt).toLocaleString()}} · #{{stage.revision}}</small><p>{{stage.state}} · {{stage.phase}}</p><p v-if="stage.error" class="error">{{stage.error}}</p><small v-if="stage.truncated">本条诊断已截断</small></li></ol>
  </section>
</template>
