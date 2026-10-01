<script setup lang="ts">
import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useDesktop} from '../desktop/store'
import {dockerQuery} from '../services/containers'
import {readTaskQuery} from '../services/task-query-lifecycle'
const props=defineProps<{nodeId:string;taskId:string;viewTabId:string}>(),desktop=useDesktop()
const offset=computed(()=>Number(desktop.state?.views[props.viewTabId]?.state.composeOutputOffset||0))
const output=useQuery({queryKey:computed(()=>['tasks',props.taskId,'compose-output',offset.value]),queryFn:({signal})=>readTaskQuery(signal,()=>dockerQuery(props.nodeId,{kind:'operation-output',taskId:props.taskId,outputOffset:offset.value},signal)),refetchInterval:3000})
const canNext=computed(()=>!output.isFetching.value&&(output.data.value?.nextOutput??-1)>=0)
function change(value:number){desktop.patchView(props.viewTabId,'composeOutputOffset',value)}
function text(value?:string|null){return new TextDecoder().decode(Uint8Array.from(atob(value||''),character=>character.charCodeAt(0)))}
</script>
<template>
  <section aria-label="Compose命令输出">
    <h3>Compose 命令输出</h3><p class="muted">显示各命令阶段已保存的输出；部署结果以任务状态和资源检查为准。</p>
    <div class="action-row"><button :disabled="!offset||output.isFetching.value" @click="change(offset-1)">上一阶段输出</button><button :disabled="!canNext" @click="change(output.data.value!.nextOutput!)">下一阶段输出</button><button @click="output.refetch()">刷新输出</button></div>
    <p v-if="output.error.value" role="alert" class="error">{{output.error.value.message}}</p>
    <template v-else-if="output.data.value?.output">
      <p>{{output.data.value.output.phase}} · {{new Date(output.data.value.output.recordedAt).toLocaleString()}}</p>
      <p>{{output.data.value.output.completed ? '命令已退出' : '尚未记录命令退出；显示最近保存的输出，当前运行状态以任务详情为准。'}}</p>
      <p v-if="output.data.value.output.commandError" class="error">{{output.data.value.output.commandError}}</p>
      <h4>标准输出</h4><pre>{{text(output.data.value.output.stdout)}}</pre><small v-if="output.data.value.output.stdoutTruncated">标准输出已截断，保留前32 KiB。</small>
      <h4>标准错误</h4><pre>{{text(output.data.value.output.stderr)}}</pre><small v-if="output.data.value.output.stderrTruncated">标准错误已截断，保留前32 KiB。</small>
    </template><p v-else>尚无已保存的命令输出。</p>
  </section>
</template>
<style scoped>pre{white-space:pre-wrap;overflow-wrap:anywhere;max-height:22rem;overflow:auto}</style>
