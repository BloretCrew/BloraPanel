<script setup lang="ts">
import {computed,ref} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {api,APIError,type Node,type Instance} from '../services/api'
import {useDesktop} from '../desktop/store'
import {copy,json} from '../recovery/state'
import InstanceConfigFields from './InstanceConfigFields.vue'
interface Creation {name:string;nodeId:string;templateId?:string;config:Record<string,unknown>}
interface Submission {requestId:string;body:Creation;state:string;error?:string}
const props=defineProps<{viewTabId:string;nodes:Node[]}>(),emit=defineEmits<{created:[instance:Instance];close:[]}>(),desktop=useDesktop(),recovery=desktop.recovery!,busy=ref(false)
const view=computed(()=>recovery.state.views[props.viewTabId]!)
function patch(key:string,value:unknown){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
const draft=computed<Creation>(()=>{const saved=view.value.state.createDraft as Record<string,unknown>||{};return{name:String(saved.name||''),nodeId:String(saved.nodeId||''),templateId:String(saved.templateId||''),config:saved.config as Record<string,unknown>||{mode:saved.mode||'container',command:[saved.command||''],directory:saved.directory||'/workspace',image:saved.image||'',uid:1000,gid:1000,stopSeconds:30,killSeconds:10,escalate:true,autostart:false}}})
const submission=computed<Submission|undefined>({get:()=>view.value.state.createSubmission as unknown as Submission|undefined,set:value=>patch('createSubmission',value||null)})
const templates=useQuery({queryKey:['instance-templates'],queryFn:()=>api<{items:{templateId:string;name:string;platform:string;description:string;config:Record<string,unknown>}[]}>('/instance-templates')})
function field(key:string,value:unknown){patch('createDraft',{...draft.value,[key]:value})}
function template(id:string){const chosen=templates.data.value?.items.find(item=>item.templateId===id);if(chosen)patch('createDraft',{...draft.value,templateId:id,config:copy(chosen.config)})}
function review(){submission.value={requestId:crypto.randomUUID(),body:{name:draft.value.name,nodeId:draft.value.nodeId,config:copy(draft.value.config)},state:'PREPARED'}}
async function submit(){if(!submission.value||busy.value)return;const fixed=copy(submission.value);busy.value=true;submission.value={...fixed,state:'SUBMITTING'};try{const {instance}=await api<{instance:Instance}>('/instances',{method:'POST',headers:{'Idempotency-Key':fixed.requestId},body:JSON.stringify(fixed.body)});submission.value=undefined;emit('created',instance)}catch(e){submission.value={...fixed,state:e instanceof APIError&&e.status<500?'REJECTED':'UNKNOWN',error:String(e)}}finally{busy.value=false}}
</script>
<template>
  <form class="in-window-dialog instance-create" role="dialog" aria-label="创建实例" @submit.prevent="submission?submit():review()">
    <template v-if="!submission"><h3>创建实例</h3><label>名称<input :value="draft.name" required maxlength="120" @input="field('name',($event.target as HTMLInputElement).value)"></label><label>目标节点<select :value="draft.nodeId" required @change="field('nodeId',($event.target as HTMLSelectElement).value)"><option value="">选择获准节点</option><option v-for="node in nodes" :key="node.nodeId" :value="node.nodeId">{{node.name}} · {{node.platform}}</option></select></label><label>配置模板<select :value="draft.templateId||''" @change="template(($event.target as HTMLSelectElement).value)"><option value="">手动配置</option><option v-for="item in templates.data.value?.items" :key="item.templateId" :value="item.templateId">{{item.name}} · {{item.platform}}</option></select></label><p v-if="draft.templateId" class="muted">{{templates.data.value?.items.find(item=>item.templateId===draft.templateId)?.description}}</p><p v-if="templates.error.value" class="error">{{templates.error.value.message}}</p><InstanceConfigFields :config="draft.config" @change="field('config',$event)" /><p class="muted">只创建配置；启动需要单独操作。提交时重新校验节点在线、模式、权限和配额。</p><div class="action-row"><button type="button" @click="emit('close')">收起</button><button class="primary">核对创建实例</button></div></template>
    <template v-else><h3>确认创建实例</h3><p>节点：{{nodes.find(node=>node.nodeId===submission!.body.nodeId)?.name||submission.body.nodeId}}</p><pre>{{JSON.stringify(submission.body,null,2)}}</pre><p>确认只创建以上配置。刷新不会提交；回执不明时用相同请求核对，不重复创建。</p><p v-if="submission.error" class="error">{{submission.error}}</p><div class="action-row"><button type="button" :disabled="busy||['SUBMITTING','UNKNOWN'].includes(submission.state)" @click="submission=undefined">返回修改</button><button class="primary" :disabled="busy">确认创建实例</button></div></template>
  </form>
</template>
