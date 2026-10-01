<script setup lang="ts">
import UIIcon from '../app-host/UIIcon.vue'
import { computed, ref } from 'vue'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { api, session, type Instance, type Node, type Task } from '../services/api'
import {readTaskQuery} from '../services/task-query-lifecycle'
import { useDesktop } from '../desktop/store'
import InstanceConsole from './InstanceConsole.vue'
import InstanceSettings from './InstanceSettings.vue'
import InstanceCreate from './InstanceCreate.vue'
import InstanceBatch from './InstanceBatch.vue'
import MonitorApp from './MonitorApp.vue'
const props=defineProps<{viewTabId:string;windowId:string;visible:boolean}>(),desktop=useDesktop(),query=useQueryClient()
const view=computed(()=>desktop.state!.views[props.viewTabId]!),search=computed({get:()=>String(view.value.state.search||''),set:value=>desktop.patchView(props.viewTabId,'search',value)}),nodeFilter=computed({get:()=>String(view.value.state.nodeFilter||''),set:value=>desktop.patchView(props.viewTabId,'nodeFilter',value)})
const instanceViewMode=computed<'list'|'cards'>({get:()=>view.value.state.instanceViewMode==='cards'?'cards':'list',set:value=>desktop.patchView(props.viewTabId,'instanceViewMode',value)})
const instances=useQuery({queryKey:['instances'],queryFn:()=>api<{items:Instance[]}>('/instances'),refetchInterval:3000})
const nodes=useQuery({queryKey:['nodes'],queryFn:()=>api<{items:Node[]}>('/nodes'),refetchInterval:5000})
const creatable=useQuery({queryKey:['nodes','creatable'],queryFn:()=>api<{items:Node[]}>('/nodes/creatable')})
const tasks=useQuery({queryKey:['tasks'],queryFn:({signal})=>readTaskQuery(signal,()=>api<{items:Task[]}>('/tasks',{signal})),refetchInterval:3000})
const stateFilter=computed({get:()=>String(view.value.state.stateFilter||''),set:value=>desktop.patchView(props.viewTabId,'stateFilter',value)}),groupFilter=computed({get:()=>String(view.value.state.groupFilter||''),set:value=>desktop.patchView(props.viewTabId,'groupFilter',value)}),tagFilter=computed({get:()=>String(view.value.state.tagFilter||''),set:value=>desktop.patchView(props.viewTabId,'tagFilter',value)})
const nodeChoices=computed(()=>[...new Map((instances.data.value?.items||[]).map(instance=>[instance.nodeId,{nodeId:instance.nodeId,name:instance.nodeName||nodes.data.value?.items.find(node=>node.nodeId===instance.nodeId)?.name||instance.nodeId}])).values()])
const groupChoices=computed(()=>[...new Set((instances.data.value?.items||[]).map(instance=>instance.group).filter(Boolean))]),tagChoices=computed(()=>[...new Set((instances.data.value?.items||[]).flatMap(instance=>instance.tags||[]))]),stateChoices=computed(()=>[...new Set((instances.data.value?.items||[]).map(instance=>instance.state))])
const items=computed(()=>instances.data.value?.items.filter(i=>[i.name,i.group,...i.tags||[]].join(' ').toLowerCase().includes(search.value.toLowerCase())&&(!nodeFilter.value||i.nodeId===nodeFilter.value)&&(!stateFilter.value||(stateFilter.value==='NODE_OFFLINE'?i.nodeState==='OFFLINE':i.state===stateFilter.value))&&(!groupFilter.value||i.group===groupFilter.value)&&(!tagFilter.value||i.tags?.includes(tagFilter.value)))||[])
const selected=computed(()=>view.value.state.selectedInstances as string[]||[])
function selectInstance(id:string){desktop.patchView(props.viewTabId,'selectedInstances',selected.value.includes(id)?selected.value.filter(value=>value!==id):[...selected.value,id])}
function selectAll(){const ids=items.value.map(instance=>instance.instanceId);desktop.patchView(props.viewTabId,'selectedInstances',ids.every(id=>selected.value.includes(id))?selected.value.filter(id=>!ids.includes(id)):[...new Set([...selected.value,...ids])])}
function batch(action:string){desktop.patchView(props.viewTabId,'batchConfirmation',{action,targets:(instances.data.value?.items||[]).filter(instance=>selected.value.includes(instance.instanceId)).map(instance=>({...copy(instance),requestId:crypto.randomUUID()}))})}
const instance=computed(()=>instances.data.value?.items.find(i=>i.instanceId===view.value.resourceRef?.id))
const page=computed(()=>String(view.value.state.page||'overview'))
type Confirmation={instance:Pick<Instance,'instanceId'|'name'|'nodeId'>;action:string;requestId?:string}
const error=ref(''),pending=ref(false),confirmation=computed<Confirmation|undefined>({get:()=>view.value.state.confirmation as Confirmation|undefined,set:value=>desktop.patchView(props.viewTabId,'confirmation',value||null)}),createOpen=computed({get:()=>!!view.value.state.createOpen,set:value=>desktop.patchView(props.viewTabId,'createOpen',value)})
const targetNode=(id:string)=>{const node=nodes.data.value?.items.find(n=>n.nodeId===id);const resource=instances.data.value?.items.find(i=>i.nodeId===id);return node || (resource?{name:resource.nodeName||id,state:resource.nodeState||'UNKNOWN'}:undefined)}
function copy(i:Instance){return {instanceId:i.instanceId,name:i.name,nodeId:i.nodeId}}
function openInstance(i:Instance,newWindow=false){desktop.open({appId:'blora.instances',windowId:newWindow?undefined:props.windowId,resourceRef:{kind:'instance',id:i.instanceId,nodeId:i.nodeId},title:`${i.name} · ${targetNode(i.nodeId)?.name || i.nodeId}`,disposition:newWindow?'new-window':'new-tab'})}
function shortcut(i:Instance){desktop.addShortcut('blora.instances',{kind:'instance',id:i.instanceId,nodeId:i.nodeId},i.name,page.value)}
function requestDelete(i:Instance){confirmation.value={instance:copy(i),action:'delete'}}
async function action(){const fixed=confirmation.value;if(!fixed)return;const requestId=fixed.requestId||crypto.randomUUID();confirmation.value={...fixed,requestId};pending.value=true;error.value='';try{if(fixed.action==='delete'){await api(`/instances/${encodeURIComponent(fixed.instance.instanceId)}`,{method:'DELETE',headers:{'Idempotency-Key':requestId}});await query.invalidateQueries({queryKey:['instances']});confirmation.value=undefined;return}const known=tasks.data.value?.items.find(t=>t.requestId===requestId);const task=known || (await api<{task:Task}>(`/instances/${encodeURIComponent(fixed.instance.instanceId)}/actions`,{method:'POST',headers:{'Idempotency-Key':requestId},body:JSON.stringify({action:fixed.action})})).task;desktop.patchView(props.viewTabId,'lastTaskId',task.taskId);await query.invalidateQueries({queryKey:['tasks']});confirmation.value=undefined}catch(e){error.value=e instanceof Error?e.message:String(e)}finally{pending.value=false}}
async function created(instance:Instance){createOpen.value=false;await query.invalidateQueries({queryKey:['instances']});openInstance(instance)}
const resourceTasks=computed(()=>tasks.data.value?.items.filter(t=>t.resource?.id===instance.value?.instanceId)||[])
</script>
<template>
  <div class="application instances-app">
    <div v-if="!view.resourceRef" class="instances-browser">
      <aside class="resource-sidebar" aria-label="实例导航">
        <div class="sidebar-group-label">资源库</div>
        <button class="sidebar-item" :class="{selected:!nodeFilter&&!stateFilter}" @click="nodeFilter='';stateFilter=''"><UIIcon name="workspace"/><span>全部实例</span><small>{{instances.data.value?.items.length??0}}</small></button>
        <button class="sidebar-item" :class="{selected:stateFilter==='RUNNING'}" @click="nodeFilter='';stateFilter='RUNNING'"><UIIcon name="play"/><span>运行中</span></button>
        <button class="sidebar-item" :class="{selected:stateFilter==='STOPPED'}" @click="nodeFilter='';stateFilter='STOPPED'"><UIIcon name="stop"/><span>已停止</span></button>
        <div class="sidebar-group-label">节点</div>
        <button v-for="node in nodeChoices" :key="node.nodeId" class="sidebar-item" :class="{selected:nodeFilter===node.nodeId}" @click="nodeFilter=node.nodeId;stateFilter=''"><UIIcon name="monitor"/><span>{{node.name}}</span></button>
        <div class="sidebar-footnote">仅显示已获授权的资源</div>
      </aside>
      <main class="resource-main">
        <header class="app-heading"><div><h2>全部可见实例 <span class="count">{{instances.data.value?.items.length ?? '—'}}</span></h2><p>集中管理节点上的服务</p></div><button class="primary" :disabled="!creatable.data.value?.items.length" :title="creatable.error.value?.message || '在获准节点上创建实例'" @click="createOpen=true">创建实例</button></header>
        <div class="filterbar instance-primary-filters"><input v-model="search" type="search" placeholder="搜索实例名称…" aria-label="搜索实例"><select v-model="nodeFilter" aria-label="节点筛选"><option value="">全部节点</option><option v-for="node in nodeChoices" :key="node.nodeId" :value="node.nodeId">{{node.name}}</option></select><button @click="instances.refetch()" aria-label="刷新">刷新</button><div class="instance-view-switch" role="group" aria-label="实例视图"><button type="button" :class="{selected:instanceViewMode==='list'}" :aria-pressed="instanceViewMode==='list'" @click="instanceViewMode='list'">列表</button><button type="button" :class="{selected:instanceViewMode==='cards'}" :aria-pressed="instanceViewMode==='cards'" @click="instanceViewMode='cards'">卡片</button></div></div>
        <div class="filterbar instance-extra-filters"><select v-model="stateFilter" aria-label="实例状态筛选"><option value="">全部状态</option><option value="NODE_OFFLINE">节点失联</option><option v-for="state in stateChoices" :key="state" :value="state">{{state}}</option></select><select v-model="groupFilter" aria-label="实例分组筛选"><option value="">全部分组</option><option v-for="group in groupChoices" :key="group" :value="group">{{group}}</option></select><select v-model="tagFilter" aria-label="实例标签筛选"><option value="">全部标签</option><option v-for="tag in tagChoices" :key="tag" :value="tag">{{tag}}</option></select></div>
        <div class="instance-batch-toolbar"><button :disabled="!items.length" @click="selectAll">选择当前筛选全部</button><small>已选 {{selected.length}} 项</small><button :disabled="!selected.length" @click="batch('start')">批量启动</button><button :disabled="!selected.length" @click="batch('stop')">批量停止</button><button :disabled="!selected.length" @click="batch('restart')">批量重启</button></div>
        <p v-if="instances.error.value" class="error">{{instances.error.value.message}}</p>
        <div v-if="instanceViewMode==='list'" class="instance-list-header"><span>名称 / 节点</span><span>状态</span><span>操作</span></div>
        <div class="instance-grid" :class="{cards:instanceViewMode==='cards'}">
          <article v-for="i in items" :key="i.instanceId" class="instance-card" @dblclick="openInstance(i)">
            <input type="checkbox" :aria-label="'选择实例 '+i.name" :checked="selected.includes(i.instanceId)" @click.stop @change="selectInstance(i.instanceId)">
            <UIIcon class="instance-resource-icon" name="server"/>
            <div class="instance-identity"><h3><button @click="openInstance(i)">{{i.name}}</button></h3><p>{{targetNode(i.nodeId)?.name || i.nodeId}}</p><small>{{i.group||'未分组'}}<span v-if="i.tags?.length"> · {{i.tags.join(', ')}}</span></small></div>
            <span class="status-chip" :data-status="i.state">{{targetNode(i.nodeId)?.state==='OFFLINE'?'节点失联 · 待确认':i.state}}</span>
            <footer><button class="instance-open" @click="openInstance(i)">管理实例</button><button class="icon-button" title="在新窗口打开" @click="openInstance(i,true)"><UIIcon name="external"/></button><button class="icon-button" title="添加到桌面" @click="shortcut(i)"><UIIcon name="pin-desktop"/><span class="sr-only">添加到桌面</span></button><button v-if="session.user?.admin" class="icon-button danger" title="删除实例（需已停止且无活动任务）" @click.stop="requestDelete(i)"><UIIcon name="trash"/><span class="sr-only">删除</span></button></footer>
          </article>
        </div>
        <div v-if="!items.length && !instances.isPending.value && !instances.error.value" class="empty-state"><h3>这里会显示获准的实例</h3><p>在有创建权限的节点上创建实例，或由管理员分配已有资源。</p></div>
        <footer class="resource-statusbar">{{items.length}} 个实例<span>关闭窗口不会停止正在运行的服务</span></footer>
      </main>
    </div>
    <template v-else-if="instance"><header class="resource-heading"><div><span class="eyebrow">{{targetNode(instance.nodeId)?.name || instance.nodeId}}</span><h2>{{instance.name}} <span class="status-chip" :data-status="instance.state">{{instance.state}}</span></h2></div><div class="action-row"><button @click="shortcut(instance)">＋ 添加到桌面</button><button v-if="session.user?.admin" title="删除实例（需已停止且无活动任务）" @click="requestDelete(instance)">删除实例</button></div></header><nav class="resource-nav"><button v-for="p in [['overview','概览'],['console','控制台'],['files','文件'],['monitor','监控'],['settings','设置']]" :key="p[0]" :class="{selected:page===p[0]}" @click="desktop.patchView(viewTabId,'page',p[0])">{{p[1]}}</button></nav><div v-if="targetNode(instance.nodeId)?.state==='OFFLINE'" class="warning notice">节点失联，运行结果等待重新确认。最近记录不代表当前进程已停止。</div><div v-if="page==='overview'" class="resource-section"><div class="action-row"><button v-for="a in [['start','启动'],['stop','停止'],['restart','重启'],['kill','强制结束']]" :key="a[0]" @click="confirmation={instance:copy(instance),action:a[0]!}">{{a[1]}}</button></div><dl class="details-grid"><dt>实例标识</dt><dd>{{instance.instanceId}}</dd><dt>节点</dt><dd>{{instance.nodeId}}</dd><dt>运行代次</dt><dd>{{instance.runId || '尚无运行'}}</dd><dt>实际状态</dt><dd>{{instance.state}}</dd></dl><h3>关联任务</h3><div v-for="task in resourceTasks" :key="task.taskId" class="task-row"><span>{{task.action}}</span><span>{{task.state}} · {{task.phase}}</span><button @click="desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:task.taskId},disposition:'dedicated',title:task.action})">查看</button></div><p v-if="!resourceTasks.length" class="muted">暂无任务。关闭此窗口不会影响实例。</p></div><div v-else-if="page==='files'" class="resource-section"><h3>实例文件</h3><button class="primary" @click="desktop.open({appId:'blora.files',resourceRef:{kind:'directory',id:instance.instanceId,nodeId:instance.nodeId},title:instance.name+' · 文件',disposition:'new-window'})">在文件管理器打开</button></div><InstanceConsole v-else-if="page==='console'" :view-tab-id="viewTabId" :instance="instance" :visible="visible" /><MonitorApp v-else-if="page==='monitor'" :view-tab-id="viewTabId" :visible="visible" /><InstanceSettings v-else-if="page==='settings'" :view-tab-id="viewTabId" :instance="instance" /><div v-else class="resource-section"><h3>运行事实</h3><pre class="selectable">{{JSON.stringify(instance,null,2)}}</pre></div></template>
    <div v-else class="empty-state"><h3>{{instances.isPending.value?'正在核对资源…':'无法读取此实例'}}</h3><p>{{instances.error.value?.message || '资源可能已删除或权限发生变化。窗口保留原资源身份，不会打开同名替代资源。'}}</p><button @click="instances.refetch()">重新检查</button></div>
    <button v-if="instance" class="backup-launch" @click="desktop.open({appId:'blora.backups',resourceRef:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},title:instance.name+' · 备份',disposition:'new-window'})">打开备份与计划</button>
    <p v-if="error" class="error notice">{{error}}</p>
    <div v-if="confirmation" class="in-window-dialog" role="dialog" aria-modal="true"><h3>{{confirmation.action==='delete'?'确认删除实例':'确认 '+confirmation.action}}</h3><p>实例：<strong>{{confirmation.instance.name}}</strong></p><p>节点：{{confirmation.instance.nodeId}}</p><p class="muted">{{confirmation.action==='delete'?'删除会保留墓碑以阻止旧入口指向同名资源；实例必须已停止且没有活动任务。':'操作固定绑定此目标，实际执行结果在任务中更新。'}}</p><div class="action-row"><button @click="confirmation=undefined">返回</button><button class="primary" :disabled="pending" @click="action">{{pending?(confirmation.action==='delete'?'删除中…':'提交中…'):'确认提交'}}</button></div></div>
    <InstanceBatch :view-tab-id="viewTabId" />
    <InstanceCreate v-if="createOpen" :view-tab-id="viewTabId" :nodes="creatable.data.value?.items||[]" @close="createOpen=false" @created="created" />
  </div>
</template>
