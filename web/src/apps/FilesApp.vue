<script setup lang="ts">
import UIIcon from '../app-host/UIIcon.vue'
import {computed,nextTick,onBeforeUnmount,onMounted,ref,watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useDesktop} from '../desktop/store'
import {api,APIError,type Task} from '../services/api'
import {readTaskQuery} from '../services/task-query-lifecycle'
import {copy,json} from '../recovery/state'
import {readDirectoryBlock,type DirectoryPage,type FileEntry} from '../services/files'
import {createUpload,resumeUpload,pauseUpload,cancelUpload,reconcileUpload,uploadIsActive} from '../services/uploads'
interface TrashItem {id:string;originalPath:string;version:string;deletedAt:string;expiresAt:string;stage:string}
interface Submission {endpoint:string;body:Record<string,unknown>;requestId:string;taskId?:string}
interface FileDialog {action:string;sources:FileEntry[];target:string;requestId:string;instanceId:string;overwrite:boolean;trashId?:string;sourceInstanceId?:string;sourceNodeId?:string;targetNodeId?:string;clipboardId?:string;submissions?:Submission[]}
interface FileClipboard {clipboardId:string;mode:'copy'|'move';instanceId:string;nodeId?:string;sources:FileEntry[];taskIds?:string[]}
interface TreeRow {path:string;name:string;depth:number;expanded:boolean;hasChildren:boolean}
interface Marquee {startX:number;startY:number;currentX:number;currentY:number;add:boolean;scrollTop:number}
const props=defineProps<{viewTabId:string;windowId:string}>(),desktop=useDesktop(),recovery=desktop.recovery!,error=ref(''),busy=ref(false),viewport=ref<HTMLElement>(),pickedFile=ref<File>(),resumeId=ref<string>(),fileInput=ref<HTMLInputElement>()
const marquee=ref<Marquee>()
const view=computed(()=>recovery.state.views[props.viewTabId]!)
function patch(key:string,value:unknown){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
const path=computed(()=>String(view.value.state.path||'.')),pathInput=computed({get:()=>String(view.value.state.pathInput??path.value),set:value=>patch('pathInput',value)})
const search=computed({get:()=>String(view.value.state.search||''),set:value=>{patch('search',value);resetPage()}}),sort=computed({get:()=>String(view.value.state.sort||'name'),set:value=>{patch('sort',value);resetPage()}}),order=computed({get:()=>String(view.value.state.order||'asc'),set:value=>{patch('order',value);resetPage()}})
const viewMode=computed<'list'|'grid'>({get:()=>view.value.state.viewMode==='grid'?'grid':'list',set:value=>{patch('viewMode',value);resetPage()}})
const scroll=computed(()=>Number(view.value.state.scroll||0)),gridOffset=computed(()=>Math.max(0,Number(view.value.state.gridOffset||0))),offset=computed(()=>viewMode.value==='grid'?gridOffset.value:Math.floor(scroll.value/36/100)*100),selected=computed(()=>view.value.state.selected as string[]||[]),dialog=computed<FileDialog|undefined>({get:()=>view.value.state.fileDialog as unknown as FileDialog|undefined,set:value=>patch('fileDialog',value||null)})
const trashOpen=computed({get:()=>!!view.value.state.trashOpen,set:value=>patch('trashOpen',value)})
const trashOffset=computed(()=>Number(view.value.state.trashOffset||0))
const resourceInstanceId=computed(()=>String(view.value.state.instanceId||view.value.resourceRef?.id||''))
const clipboard=computed<FileClipboard|undefined>(()=>recovery.state.preferences.fileClipboard as unknown as FileClipboard|undefined)
function setClipboard(value:FileClipboard|undefined){recovery.commit([{kind:'set',path:['preferences','fileClipboard'],value:json(value||null)}])}
const clipboardPending=computed(()=>!!clipboard.value?.taskIds?.some(id=>!tasks.data.value?.items.some(task=>task.taskId===id&&['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state))))
const endpoint=computed(()=>`/instances/${encodeURIComponent(resourceInstanceId.value)}/files`)
const files=useQuery({queryKey:computed(()=>['files',resourceInstanceId.value,path.value,search.value,sort.value,order.value,offset.value]),enabled:computed(()=>!!view.value.resourceRef),queryFn:({signal})=>readDirectoryBlock(endpoint.value,{path:path.value,offset:offset.value,search:search.value,sort:sort.value,order:order.value},signal)})
const treeExpanded=computed(()=>{const value=view.value.state.treeExpanded;const paths=Array.isArray(value)?value.filter((item):item is string=>typeof item==='string'&&item.length>0):[];return paths.includes('.')?paths:['.',...paths]})
const tree=useQuery({queryKey:computed(()=>['file-tree',resourceInstanceId.value,treeExpanded.value.join('|')]),enabled:computed(()=>!!view.value.resourceRef),queryFn:async({signal})=>{const entries=await Promise.all(treeExpanded.value.map(async directory=>{const query=new URLSearchParams({path:directory,offset:'0',limit:'100',search:'',sort:'name',order:'asc'});return[directory,await api<DirectoryPage>(`${endpoint.value}?${query}`,{signal})] as const}));return Object.fromEntries(entries)}})
const trash=useQuery({queryKey:computed(()=>['trash',resourceInstanceId.value,trashOffset.value]),enabled:computed(()=>!!view.value.resourceRef&&trashOpen.value),queryFn:()=>api<{items:TrashItem[];total:number;nextOffset:number}>(`${endpoint.value}/trash?offset=${trashOffset.value}&limit=10`)})
const tasks=useQuery({queryKey:['tasks'],queryFn:({signal})=>readTaskQuery(signal,()=>api<{items:Task[]}>('/tasks',{signal})),refetchInterval:2000})
const taskIds=computed(()=>view.value.state.fileTasks as string[]||[]),fileTasks=computed(()=>tasks.data.value?.items.filter(task=>taskIds.value.includes(task.taskId))||[])
const uploads=computed(()=>Object.values(recovery.state.uploads).filter(upload=>upload.instanceId===resourceInstanceId.value)),total=computed(()=>files.data.value?.total??Number(view.value.state.total||0))
const treeRows=computed(()=>{const rows:TreeRow[]=[];const walk=(directory:string,depth:number)=>{if(depth>8)return;for(const entry of tree.data.value?.[directory]?.items||[]){if(!(entry.isDir||entry.kind==='directory'))continue;const expanded=treeExpanded.value.includes(entry.path);rows.push({path:entry.path,name:entry.name,depth,expanded,hasChildren:true});if(expanded)walk(entry.path,depth+1)}};walk('.',0);return rows})
watch(()=>files.data.value,data=>{if(data){patch('total',data.total??data.items.length);void nextTick(()=>{if(viewport.value&&Math.abs(viewport.value.scrollTop-scroll.value)>2)viewport.value.scrollTop=scroll.value})}})
watch(()=>fileTasks.value.map(task=>`${task.taskId}:${task.state}`).join(','),()=>{if(fileTasks.value.some(task=>['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state))){void files.refetch();if(trashOpen.value)void trash.refetch()}})
watch(()=>tasks.data.value,()=>{const clip=clipboard.value;if(clip?.mode==='move'&&clip.taskIds?.length===clip.sources.length&&clip.taskIds.every(id=>tasks.data.value?.items.some(task=>task.taskId===id&&task.state==='SUCCEEDED')))setClipboard(undefined)})
function resetPage(){patch('scroll',0);patch('gridOffset',0);patch('selected',[]);patch('total',0);if(viewport.value)viewport.value.scrollTop=0}
function moveGridPage(direction:'previous'|'next'){
  if(direction==='previous')patch('gridOffset',Math.max(0,gridOffset.value-100))
  else if(files.data.value?.nextOffset&&files.data.value.nextOffset>gridOffset.value)patch('gridOffset',files.data.value.nextOffset)
  patch('selected',[])
}
function ensureTreePath(target:string){const clean=target.trim().replace(/^\.\//,'').replace(/\/$/,'');const next=['.',...clean.split('/').filter(Boolean).reduce<string[]>((paths,part)=>{const parent=paths.at(-1)!;paths.push(parent==='.'?part:`${parent}/${part}`);return paths},[])];const merged=[...new Set([...treeExpanded.value,...next])];if(merged.length!==treeExpanded.value.length||merged.some((item,index)=>item!==treeExpanded.value[index]))patch('treeExpanded',merged)}
function toggleTree(directory:string){const next=treeExpanded.value.includes(directory)?treeExpanded.value.filter(item=>item!==directory):[...treeExpanded.value,directory];patch('treeExpanded',[...new Set(next)]);if(!treeExpanded.value.includes(directory))void tree.refetch()}
function treeNavigate(directory:string){ensureTreePath(directory);navigate(directory)}
function navigate(target:string,historyIndex?:number){
  const value=target.trim()||'.',history=(view.value.state.history as string[]||['.']),index=Number(view.value.state.historyIndex||0)
  if(historyIndex===undefined){patch('history',[...history.slice(0,index+1),value]);patch('historyIndex',index+1)}else patch('historyIndex',historyIndex)
  patch('path',value);patch('pathInput',value);ensureTreePath(value);resetPage()
}
function historyMove(delta:number){const history=view.value.state.history as string[]||['.'],index=Number(view.value.state.historyIndex||0)+delta;if(index>=0&&index<history.length)navigate(history[index]!,index)}
const join=(directory:string,name:string)=>directory==='.'?name:directory.replace(/\/$/,'')+'/'+name
function open(file:FileEntry){if(file.isDir||file.kind==='directory'){navigate(file.path);return}openEditor(file.path,file.name)}
function openEditor(filePath:string,title:string,newFile=false){desktop.open({appId:'blora.editor',resourceRef:{kind:'file',id:`${resourceInstanceId.value}:${filePath}`,nodeId:view.value.resourceRef!.nodeId},title,disposition:'new-window',state:{path:filePath,instanceId:resourceInstanceId.value,newFile}})}
function pinDirectory(){desktop.addShortcut('blora.files',{kind:'directory',id:`${resourceInstanceId.value}:${path.value}`,nodeId:view.value.resourceRef?.nodeId},path.value==='.'?'实例根目录':path.value.split('/').pop()!,undefined,{instanceId:resourceInstanceId.value,path:path.value,history:[path.value],historyIndex:0})}
function select(event:MouseEvent,file:FileEntry){
  const rows=files.data.value?.items||[];let next=event.ctrlKey||event.metaKey?selected.value.filter(path=>path!==file.path):[file.path]
  if((event.ctrlKey||event.metaKey)&&!selected.value.includes(file.path))next.push(file.path)
  if(event.shiftKey){const anchor=rows.findIndex(row=>row.path===view.value.state.selectionAnchor),index=rows.findIndex(row=>row.path===file.path);if(anchor>=0)next=rows.slice(Math.min(anchor,index),Math.max(anchor,index)+1).map(row=>row.path)}else patch('selectionAnchor',file.path)
  patch('selected',next)
}
function beginMarquee(event:PointerEvent){
  if(viewMode.value!=='list'||event.button!==0||!viewport.value||!(event.target instanceof Element)||event.target.closest('.file-row'))return
  const rect=viewport.value.getBoundingClientRect(),x=event.clientX-rect.left,y=event.clientY-rect.top
  marquee.value={startX:x,startY:y,currentX:x,currentY:y,add:event.ctrlKey||event.metaKey,scrollTop:viewport.value.scrollTop}
  try{viewport.value.setPointerCapture?.(event.pointerId)}catch{/* Synthetic/browser-limited pointer events may not support capture. */}
  window.addEventListener('pointermove',updateMarquee)
  window.addEventListener('pointerup',finishMarquee,{once:true})
  event.preventDefault()
}
function updateMarquee(event:PointerEvent){
  const active=marquee.value,element=viewport.value;if(!active||!element)return
  const rect=element.getBoundingClientRect();active.currentX=Math.max(0,Math.min(rect.width,event.clientX-rect.left));active.currentY=Math.max(0,Math.min(rect.height,event.clientY-rect.top))
}
function finishMarquee(){
  const active=marquee.value,element=viewport.value;if(!active||!element)return
  const y1=Math.min(active.startY,active.currentY)+active.scrollTop,y2=Math.max(active.startY,active.currentY)+active.scrollTop
  if(Math.abs(active.currentX-active.startX)>3||Math.abs(active.currentY-active.startY)>3){
    const picked=(files.data.value?.items||[]).filter((_,index)=>{const top=(offset.value+index)*36;return top+36>=y1&&top<=y2}).map(file=>file.path)
    patch('selected',active.add?[...new Set([...selected.value,...picked])]:picked)
  }else if(!active.add)patch('selected',[])
  marquee.value=undefined;window.removeEventListener('pointermove',updateMarquee)
}
function marqueeStyle(){const active=marquee.value;if(!active)return {};return {left:Math.min(active.startX,active.currentX)+'px',top:(Math.min(active.startY,active.currentY)+active.scrollTop)+'px',width:Math.abs(active.currentX-active.startX)+'px',height:Math.abs(active.currentY-active.startY)+'px'}}
onBeforeUnmount(()=>{window.removeEventListener('pointermove',updateMarquee);window.removeEventListener('pointerup',finishMarquee)})
async function stat(filePath:string,instanceId=resourceInstanceId.value):Promise<FileEntry|undefined>{try{return await api<FileEntry>(`/instances/${encodeURIComponent(instanceId)}/files/stat?path=${encodeURIComponent(filePath)}`)}catch(e){if(e instanceof APIError&&e.status===404)return undefined;throw e}}
async function captureClipboard(mode:'copy'|'move',instanceId=resourceInstanceId.value,nodeId=view.value.resourceRef?.nodeId,paths=[...selected.value]){
  if(!paths.length)throw new Error('请先选择文件');const sources:FileEntry[]=[]
  for(const path of paths){const source=await stat(path,instanceId);if(!source?.version)throw new Error(`无法读取来源版本：${path}`);sources.push(source)}
  const clip:FileClipboard={clipboardId:crypto.randomUUID(),mode,instanceId,nodeId,sources};setClipboard(clip);return clip
}
async function copySelection(mode:'copy'|'move'){if(busy.value)return;busy.value=true;error.value='';try{await captureClipboard(mode)}catch(e){error.value=String(e)}finally{busy.value=false}}
async function paste(clip=clipboard.value){
  if(!clip||busy.value)return;busy.value=true;error.value=''
  const instanceId=resourceInstanceId.value,targetNodeId=view.value.resourceRef?.nodeId,directory=path.value
  try{
    if(clip.taskIds?.some(id=>!tasks.data.value?.items.some(task=>task.taskId===id&&['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state))))throw new Error('上一次粘贴仍待核对，请先查看任务结果')
    // The clipboard already contains captured identities and versions. Open and
    // protect its local confirmation immediately; recheck source at submission.
    const only=clip.sources.length===1?clip.sources[0]:undefined;let target=only?join(directory,only.name):directory
    if(only&&clip.instanceId===instanceId&&target===only.path){if(clip.mode==='move')throw new Error('来源已在当前目录，无需移动');target+='.copy'}
    dialog.value={action:clip.mode,sources:copy(clip.sources),target,requestId:crypto.randomUUID(),instanceId,sourceInstanceId:clip.instanceId,sourceNodeId:clip.nodeId,targetNodeId,clipboardId:clip.clipboardId,overwrite:false}
  }catch(e){error.value=String(e)}finally{busy.value=false}
}
async function prepare(action:string,item?:TrashItem){
  if(!view.value.resourceRef)return;busy.value=true;error.value=''
  const instanceId=resourceInstanceId.value,paths=[...selected.value]
  try{
    const sources:FileEntry[]=[]
    if(!['mkdir','newfile','upload','restore'].includes(action))for(const filePath of paths){const source=await stat(filePath);if(!source)throw new Error(`来源不存在：${filePath}`);sources.push(source)}
    if(!['mkdir','newfile','upload','restore'].includes(action)&&!sources.length)throw new Error('请先选择文件')
    const target=action==='restore'?item!.originalPath:action==='mkdir'?join(path.value,'新建目录'):action==='newfile'?join(path.value,'未命名.txt'):action==='upload'?join(path.value,pickedFile.value?.name||''):sources.length===1?join(path.value,sources[0]!.name+(action==='compress'?'.tar.gz':action==='extract'?'.extracted':'.copy')):path.value
    dialog.value={action,sources,target,requestId:crypto.randomUUID(),instanceId,overwrite:false,trashId:item?.id}
  }catch(e){error.value=String(e)}finally{busy.value=false}
}
function dialogField(key:'target'|'overwrite',value:string|boolean){if(dialog.value&&!dialog.value.submissions)dialog.value={...dialog.value,[key]:value}}
async function submit(){
  const fixed=dialog.value?copy(dialog.value):undefined;if(!fixed||busy.value)return;busy.value=true;error.value=''
  try{
    if(fixed.action==='newfile'){openEditor(fixed.target,fixed.target.split('/').pop()!,true);dialog.value=undefined;return}
    const base=`/instances/${encodeURIComponent(fixed.instanceId)}/files`
    if(fixed.action==='upload'){
      if(!pickedFile.value)throw new Error('刷新后需要重新选择原本机文件')
      const existing=await stat(fixed.target);if(existing&&!fixed.overwrite)throw new Error('目标已存在，请改名或明确允许覆盖当前版本')
      createUpload(recovery,fixed.instanceId,view.value.resourceRef?.nodeId,fixed.target,pickedFile.value,existing?.version||'missing');pickedFile.value=undefined;dialog.value=undefined;return
    }
    if(!fixed.submissions){
      if(fixed.sourceInstanceId)for(const source of fixed.sources){const current=await stat(source.path,fixed.sourceInstanceId);if(current?.version!==source.version)throw new Error(`来源已改变或不存在，请重新选择：${source.path}`)}
      fixed.submissions=[];const sources=fixed.sources.length?fixed.sources:[undefined]
      for(const [index,source] of sources.entries()){
        const target=fixed.sources.length>1?join(fixed.target,source!.name):fixed.target,existing=['mkdir','delete'].includes(fixed.action)?undefined:await stat(target,fixed.instanceId)
        if(existing&&!fixed.overwrite)throw new Error(`目标已存在，尚未覆盖：${target}`)
        const crossInstance=!!fixed.sourceInstanceId&&fixed.sourceInstanceId!==fixed.instanceId
        const body=crossInstance?{source:{instanceId:fixed.sourceInstanceId,path:source!.path,version:source!.version},target:{instanceId:fixed.instanceId,path:target,version:existing?.version||'missing'},move:fixed.action==='move'}:{action:fixed.action==='rename'?'move':fixed.action,path:fixed.action==='mkdir'?fixed.target:source?.path,target,version:source?.version,targetVersion:existing?.version||'missing',trashId:fixed.trashId}
        fixed.submissions.push({endpoint:crossInstance?'/transfers':`${base}/actions`,body,requestId:`${fixed.requestId}:${index}`})
      }
      // Persist every exact request before its first possible remote side effect.
      dialog.value=copy(fixed)
    }
    for(const submission of fixed.submissions){
      if(submission.taskId)continue
      const {task}=await api<{task:Task}>(submission.endpoint,{method:'POST',headers:{'Idempotency-Key':submission.requestId},body:JSON.stringify(submission.body)})
      submission.taskId=task.taskId;dialog.value=copy(fixed);patch('fileTasks',[...new Set([...taskIds.value,task.taskId])].slice(-100))
      if(fixed.clipboardId===clipboard.value?.clipboardId)setClipboard({...clipboard.value!,taskIds:fixed.submissions.flatMap(item=>item.taskId?[item.taskId]:[])})
    }
    dialog.value=undefined;await tasks.refetch()
  }catch(e){error.value=String(e)}finally{busy.value=false}
}
function download(){for(const filePath of selected.value){const a=document.createElement('a');a.href=`/api/v1${endpoint.value}/download?path=${encodeURIComponent(filePath)}`;a.download=filePath.split('/').pop()!;a.click()}}
async function chooseFile(event:Event){const input=event.target as HTMLInputElement,file=input.files?.[0];input.value='';if(!file)return;if(resumeId.value){void resumeUpload(recovery,resumeId.value,file);resumeId.value=undefined;return}pickedFile.value=file;await prepare('upload')}
function selectUpload(uploadId?:string){resumeId.value=uploadId;fileInput.value?.click()}
const dragType='application/x-blora-files'
function dragStart(event:DragEvent,file:FileEntry){if(!event.dataTransfer)return;event.dataTransfer.effectAllowed='copyMove';event.dataTransfer.setData(dragType,JSON.stringify({workspace:recovery.key,instanceId:resourceInstanceId.value,nodeId:view.value.resourceRef?.nodeId,paths:selected.value.includes(file.path)?selected.value:[file.path]}))}
async function drop(event:DragEvent){
  if(!event.dataTransfer)return;event.preventDefault()
  if(event.dataTransfer.types.includes(dragType)){
    try{const source=JSON.parse(event.dataTransfer.getData(dragType));if(source.workspace!==recovery.key||typeof source.instanceId!=='string'||!Array.isArray(source.paths)||source.paths.length>100||source.paths.some((path:unknown)=>typeof path!=='string'))throw new Error('拖放来源不属于此账号工作区');const clip=await captureClipboard(event.shiftKey?'move':'copy',source.instanceId,source.nodeId,source.paths);await paste(clip)}catch(e){error.value=String(e)}
  }else if(event.dataTransfer.files.length){pickedFile.value=event.dataTransfer.files[0];await prepare('upload')}
}
function keyboard(event:KeyboardEvent){if((event.target as HTMLElement).closest('input,textarea,select'))return;if(event.key==='F2'){event.preventDefault();void prepare('rename')}else if(event.key==='Delete'){event.preventDefault();void prepare('delete')}else if(event.ctrlKey||event.metaKey){const key=event.key.toLowerCase();if(['a','c','x','v'].includes(key))event.preventDefault();if(key==='a')patch('selected',files.data.value?.items.map(file=>file.path)||[]);else if(key==='c'||key==='x')void copySelection(key==='c'?'copy':'move');else if(key==='v')void paste()}}
onMounted(()=>{ensureTreePath(path.value);for(const upload of uploads.value)if(!uploadIsActive(recovery,upload.uploadId)&&!['succeeded','cancelled'].includes(upload.stage))void reconcileUpload(recovery,upload.uploadId)})
</script>
<template>
  <div class="application files-app" tabindex="0" @keydown="keyboard" @dragover.prevent @drop="drop">
    <header class="app-heading"><div><span class="eyebrow">FILES</span><h2>文件管理器</h2><p>{{view.resourceRef?.nodeId || '从实例窗口打开文件以绑定授权目录。'}}</p></div></header>
    <div v-if="view.resourceRef" class="files-browser">
      <aside class="resource-sidebar" aria-label="文件导航"><div class="sidebar-group-label">位置</div><button class="sidebar-item" :class="{selected:path==='.'&&!trashOpen}" @click="navigate('.');trashOpen=false"><UIIcon name="folder"/><span>实例根目录</span></button><div class="file-tree" role="tree" aria-label="目录树"><div v-for="row in treeRows" :key="row.path" class="file-tree-row" role="treeitem" :aria-level="row.depth+1" :aria-expanded="row.hasChildren?row.expanded:undefined" :aria-selected="path===row.path&&!trashOpen" :style="{paddingLeft:6+row.depth*14+'px'}"><button type="button" class="file-tree-toggle" :aria-label="row.expanded?'收起目录 '+row.name:'展开目录 '+row.name" :disabled="!row.hasChildren" @click.stop="toggleTree(row.path)">{{row.hasChildren?(row.expanded?'⌄':'›'):''}}</button><button type="button" class="file-tree-label" :class="{selected:path===row.path&&!trashOpen}" @click="treeNavigate(row.path)"><UIIcon name="folder"/><span>{{row.name}}</span></button></div><small v-if="tree.isPending.value" class="file-tree-status">正在读取目录树…</small><small v-else-if="tree.error.value" class="file-tree-status error">目录树读取失败</small></div><button class="sidebar-item" :class="{selected:trashOpen}" @click="trashOpen=!trashOpen"><UIIcon name="trash"/><span>回收区</span></button><div class="sidebar-group-label">当前实例</div><span class="sidebar-resource-name">{{resourceInstanceId}}</span><div class="sidebar-footnote">文件操作由节点确认<br>支持拖放与快捷键</div></aside>
      <div class="files-content">
      <form class="filterbar file-pathbar" @submit.prevent="navigate(pathInput)"><button type="button" class="icon-button" aria-label="后退目录" @click="historyMove(-1)"><UIIcon name="left"/></button><button type="button" class="icon-button" aria-label="前进目录" @click="historyMove(1)"><UIIcon name="right"/></button><input v-model="pathInput" aria-label="目录路径"><button>转到</button><button type="button" @click="files.refetch()">刷新</button><button type="button" @click="pinDirectory">固定目录到桌面</button></form>
      <div class="file-actions"><button @click="prepare('newfile')">新建文本</button><button @click="prepare('mkdir')">新建目录</button><button @click="selectUpload()">上传文件</button><button :disabled="!selected.length" @click="download">下载</button><button :disabled="selected.length!==1" @click="prepare('rename')">重命名</button><button :disabled="!selected.length||busy" @click="copySelection('copy')">复制</button><button :disabled="!selected.length||busy" @click="copySelection('move')">剪切</button><button :disabled="!clipboard||clipboardPending||busy" @click="paste()">粘贴</button><button :disabled="!selected.length" @click="prepare('copy')">复制到…</button><button :disabled="!selected.length" @click="prepare('move')">移动到…</button><button :disabled="!selected.length" @click="prepare('delete')">移入回收区</button><button :disabled="!selected.length" @click="prepare('compress')">压缩</button><button :disabled="selected.length!==1" @click="prepare('extract')">解压</button></div>
      <div class="filterbar"><input v-model="search" aria-label="搜索目录文件" placeholder="在此目录搜索名称"><select v-model="sort" aria-label="文件排序"><option value="name">名称</option><option value="size">大小</option><option value="modified">修改时间</option><option value="kind">类型</option></select><select v-model="order" aria-label="排序方向"><option value="asc">升序</option><option value="desc">降序</option></select><div class="file-view-switch" role="group" aria-label="文件视图"><button type="button" :class="{selected:viewMode==='list'}" :aria-pressed="viewMode==='list'" @click="viewMode='list'">列表</button><button type="button" :class="{selected:viewMode==='grid'}" :aria-pressed="viewMode==='grid'" @click="viewMode='grid'">图标</button></div></div>
      <p v-if="clipboard" class="file-clipboard muted">内部剪贴板：{{clipboard.mode==='move'?'剪切':'复制'}} {{clipboard.sources.length}} 项 · {{clipboard.nodeId}} / {{clipboard.instanceId}} <span v-if="clipboardPending">· 等待粘贴任务确认</span><button :disabled="clipboardPending" @click="setClipboard(undefined)">清空剪贴板</button></p>
      <p v-if="files.error.value||error" class="error" role="alert">{{files.error.value?.message||error}}</p>
      <template v-if="viewMode==='list'">
        <div class="file-list-head"><span>名称</span><span>大小</span><span>修改时间</span></div>
        <div ref="viewport" class="file-viewport" role="grid" aria-label="节点文件列表" :aria-rowcount="total" @pointerdown="beginMarquee" @scroll="patch('scroll',($event.target as HTMLElement).scrollTop)">
          <div class="file-spacer" :style="{height:Math.max(total*36,120)+'px'}">
            <div v-for="(file,index) in files.data.value?.items" :key="file.path" class="file-row" :class="{selected:selected.includes(file.path),cut:clipboard?.mode==='move'&&clipboard.instanceId===resourceInstanceId&&clipboard.sources.some(source=>source.path===file.path)}" draggable="true" @pointerdown.stop @dragstart="dragStart($event,file)" role="row" tabindex="0" :aria-selected="selected.includes(file.path)" :style="{top:(offset+index)*36+'px'}" @click="select($event,file)" @dblclick="open(file)" @keydown.enter="open(file)"><span><button @click.stop="open(file)"><UIIcon v-if="file.isDir||file.kind==='directory'" class="file-entry-icon folder" name="folder"/><UIIcon v-else class="file-entry-icon document" name="file"/>{{file.name}}</button></span><span>{{file.isDir||file.kind==='directory'?'—':file.size+' B'}}</span><span>{{file.modified?new Date(file.modified).toLocaleString():'—'}}</span></div>
          </div>
          <div v-if="marquee" class="file-marquee" :style="marqueeStyle()" aria-hidden="true"></div>
        </div>
      </template>
      <template v-else>
        <div class="file-icon-grid" role="grid" aria-label="节点文件图标视图" :aria-rowcount="total">
          <article v-for="file in files.data.value?.items" :key="file.path" class="file-icon-card" :class="{selected:selected.includes(file.path),cut:clipboard?.mode==='move'&&clipboard.instanceId===resourceInstanceId&&clipboard.sources.some(source=>source.path===file.path)}" draggable="true" role="gridcell" tabindex="0" :aria-selected="selected.includes(file.path)" @dragstart="dragStart($event,file)" @click="select($event,file)" @dblclick="open(file)" @keydown.enter="open(file)">
            <button class="file-icon-preview" type="button" @click.stop="open(file)"><UIIcon v-if="file.isDir||file.kind==='directory'" class="file-entry-icon folder" name="folder"/><UIIcon v-else class="file-entry-icon document" name="file"/></button>
            <strong>{{file.name}}</strong><small>{{file.isDir||file.kind==='directory'?'目录':file.size+' B'}}</small>
          </article>
        </div>
        <div class="file-grid-pager" aria-label="图标视图分页"><button type="button" :disabled="gridOffset===0||files.isFetching.value" @click="moveGridPage('previous')">上一页</button><span>{{gridOffset+1}}–{{Math.min(gridOffset+(files.data.value?.items.length||0),total)}} / {{total}}</span><button type="button" :disabled="!files.data.value||files.data.value.nextOffset<=gridOffset||files.isFetching.value" @click="moveGridPage('next')">下一页</button></div>
      </template>
      <small class="muted">{{total}} 项 · 已选 {{selected.length}} 项 {{files.isFetching.value?'· 正在读取当前区域…':''}}</small>
      <section v-if="trashOpen" class="file-task-list"><h3>回收区</h3><p v-if="trash.error.value" class="error">{{trash.error.value.message}}</p><div v-for="item in trash.data.value?.items" :key="item.id" class="task-row"><span>{{item.originalPath}}</span><span>{{item.stage}}</span><button @click="prepare('restore',item)">恢复</button></div><div class="action-row"><button :disabled="trashOffset===0" @click="patch('trashOffset',Math.max(0,trashOffset-10))">上一页回收记录</button><span>{{trash.data.value?.total||0}} 项</span><button :disabled="!trash.data.value||trash.data.value.nextOffset<=0" @click="patch('trashOffset',trash.data.value!.nextOffset)">下一页回收记录</button></div></section>
      <section v-if="uploads.length" class="file-task-list"><h3>本机上传</h3><article v-for="upload in uploads" :key="upload.uploadId" class="upload-row"><strong>{{upload.sourceName}}</strong><span>{{upload.stage==='hashing'?'正在校验本机来源':upload.stage==='waiting_client'?'等待本机文件或重试':upload.stage==='succeeded'?'节点确认已提交':upload.stage==='cancelled'?'已取消':upload.stage}}</span><progress :value="upload.stage==='hashing'?upload.hashOffset:upload.offset" :max="upload.total||1"></progress><small>节点已确认 {{upload.offset}} / {{upload.total}} B</small><p v-if="upload.error" class="error">{{upload.error}}</p><div><button v-if="['hashing','uploading'].includes(upload.stage)" @click="pauseUpload(recovery,upload.uploadId)">暂停</button><button v-if="!['succeeded','cancelled','hashing','uploading'].includes(upload.stage)" @click="selectUpload(upload.uploadId)">选择原文件续传</button><button v-if="upload.taskId" @click="reconcileUpload(recovery,upload.uploadId)">核对进度</button><button v-if="!['succeeded','cancelled'].includes(upload.stage)" @click="cancelUpload(recovery,upload.uploadId)">请求取消</button></div></article></section>
      <section v-if="fileTasks.length" class="file-task-list"><h3>文件任务</h3><div v-for="task in fileTasks" :key="task.taskId" class="task-row"><span>{{task.action}}</span><span>{{task.state}} · {{task.phase}}</span><button @click="desktop.open({appId:'blora.tasks',resourceRef:{kind:'task',id:task.taskId},title:task.action,disposition:'dedicated'})">详情</button></div></section>
      <input ref="fileInput" type="file" class="file-picker" aria-label="选择本机文件" @change="chooseFile">
      </div>
    </div>
    <div v-else class="empty-state"><h3>选择实例文件入口</h3><button class="primary" @click="desktop.open({appId:'blora.instances'})">打开实例中心</button></div>
    <form v-if="dialog" class="in-window-dialog" role="dialog" aria-label="确认文件操作" @submit.prevent="submit"><h3>{{({mkdir:'新建目录',newfile:'新建文本',copy:'复制',move:'移动',rename:'重命名',delete:'移入回收区',compress:'压缩',extract:'解压',restore:'恢复',upload:'上传'} as Record<string,string>)[dialog.action]}}</h3><p>目标实例：{{dialog.targetNodeId}} / {{dialog.instanceId}}</p><p v-if="dialog.sourceInstanceId">来源实例：{{dialog.sourceNodeId}} / {{dialog.sourceInstanceId}}</p><p v-for="source in dialog.sources" :key="source.path" class="selectable">{{source.path}}</p><label v-if="dialog.action!=='delete'">{{dialog.sources.length>1?'目标目录':'目标路径'}}<input :value="dialog.target" :disabled="!!dialog.submissions" required aria-label="文件操作目标" @input="dialogField('target',($event.target as HTMLInputElement).value)"></label><label v-if="dialog.action!=='delete'&&dialog.action!=='mkdir'&&dialog.action!=='newfile'" class="check-row"><input type="checkbox" :disabled="!!dialog.submissions" :checked="dialog.overwrite" @change="dialogField('overwrite',($event.target as HTMLInputElement).checked)">允许覆盖目标当前版本</label><p class="muted">{{dialog.action==='delete'?'此操作将指定目标移入节点回收区。':'操作会固定绑定以上实例和路径，任务结果由节点确认。'}}</p><p v-if="dialog.sourceInstanceId&&dialog.sourceInstanceId!==dialog.instanceId" class="warning">{{dialog.action==='move'?'跨实例移动先复制、验证目标，再验证并删除源。部分结果可能已提交；取消不会回滚。目录重叠由节点在写入前校验。':'跨实例复制会校验来源与目标版本；部分结果可能已提交，取消不会回滚。'}}</p><p v-if="dialog.submissions" class="warning">已固定此次请求。刷新只恢复现场；再次确认使用原请求身份和原版本核对，已接受的任务不会重复创建。</p><div class="action-row"><button type="button" :disabled="busy" @click="dialog=undefined">返回</button><button class="primary" :disabled="busy">{{busy?'提交中…':'确认文件操作'}}</button></div></form>
  </div>
</template>
