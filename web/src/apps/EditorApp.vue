<script setup lang="ts">
import { randomUUID } from '../services/uuid'
import {computed,nextTick,onBeforeUnmount,onMounted,ref,watch} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useDesktop} from '../desktop/store'
import {createDraft,documentModel,monaco} from '../services/documents'
import {api,APIError,session,type Task,type Instance} from '../services/api'
import {copy,json,editorHistoryBudgetBytes} from '../recovery/state'
import {id,type PendingSave,type Draft} from '../app-host/types'
const props=defineProps<{viewTabId:string;windowId:string;visible:boolean}>(),desktop=useDesktop(),primaryContainer=ref<HTMLElement>(),splitContainer=ref<HTMLElement>(),error=ref(''),downloadSourceAfterOpenError=ref(false),busy=ref(false),ready=ref(false)
const recovery=desktop.recovery!,originalView=recovery.state.views[props.viewTabId]!
type DocumentContent={text:string;version:string;encoding?:string;newline?:string;maxBytes?:number}
type PreviewContent={text:string;version:string;offset:number;nextOffset:number;total:number;hasMore:boolean;encoding?:string;newline?:string;maxBytes?:number}
type FileAccess={instanceId:string;nodeId:string;nodeName?:string;nodeState?:string;canRead:boolean;canWrite:boolean}
function editableText(text:string){return text.startsWith('\ufeff')?text.slice(1):text}
function documentBody(text:string,encoding:string|undefined){return encoding==='UTF-8 BOM'&&!text.startsWith('\ufeff')?'\ufeff'+text:text}
// Closing removes the view record before Vue disposes query/watch effects.
// Keep that final local view readable during teardown; writes still check identity.
const view=computed(()=>recovery.state.views[props.viewTabId]||originalView),draftId=computed(()=>String(view.value.state.draftId||'')),draft=computed(()=>recovery.state.drafts[draftId.value]),dirty=computed(()=>draft.value?.text!==draft.value?.savedText),resourceInstanceId=computed(()=>String(view.value.state.instanceId||view.value.resourceRef?.id.split(':')[0]||''))
function patchView(key:string,value:unknown){if(recovery.state.views[props.viewTabId])recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
const splitEnabled=computed({get:()=>!!view.value.state.editorSplit,set:value=>patchView('editorSplit',value)})
const findOpen=computed({get:()=>!!view.value.state.findOpen,set:value=>patchView('findOpen',value)})
const findText=computed({get:()=>String(view.value.state.findText||''),set:value=>patchView('findText',value)})
const replaceText=computed({get:()=>String(view.value.state.replaceText||''),set:value=>patchView('replaceText',value)})
const findCase=computed({get:()=>!!view.value.state.findCase,set:value=>patchView('findCase',value)})
const findRegex=computed({get:()=>!!view.value.state.findRegex,set:value=>patchView('findRegex',value)})
const findInput=ref<HTMLInputElement>(),findMessage=ref('')
interface SaveAsForm {instanceId:string;path:string;overwrite:boolean;requestId:string}
const saveAsForm=computed<SaveAsForm|undefined>({get:()=>view.value.state.saveAsForm as unknown as SaveAsForm|undefined,set:value=>patchView('saveAsForm',value||null)})
const reloadForm=computed<(DocumentContent&{path:string})|undefined>({get:()=>view.value.state.reloadForm as (DocumentContent&{path:string})|undefined,set:value=>patchView('reloadForm',value||null)})
const destinations=useQuery({queryKey:['instances'],enabled:computed(()=>!!saveAsForm.value),queryFn:()=>api<{items:Instance[]}>('/instances')})
const fileAccess=useQuery({queryKey:computed(()=>['file-access',resourceInstanceId.value]),enabled:computed(()=>!!view.value.resourceRef&&!!resourceInstanceId.value),queryFn:()=>api<FileAccess>(`/instances/${encodeURIComponent(resourceInstanceId.value)}/files/access`),refetchInterval:30000})
const readonly=computed({get:()=>!!view.value.state.editorReadOnly,set:value=>patchView('editorReadOnly',value)})
const effectiveReadonly=computed(()=>readonly.value||(!!draft.value?.resourceRef&&(fileAccess.data.value?.canWrite===false||fileAccess.isError.value)))
const conflict=computed<DocumentContent|undefined>({get:()=>view.value.state.comparedServer as DocumentContent|undefined,set:value=>patchView('comparedServer',value||null)}),saveState=computed(()=>draft.value?.pendingSave?.state||''),savePending=computed(()=>!!draft.value?.pendingSave && !['FAILED','CANCELLED','INTERRUPTED'].includes(saveState.value))
const documentEncoding=computed(()=>draft.value?.encoding||(draft.value?.base.startsWith('\ufeff')?'UTF-8 BOM':'UTF-8'))
const documentNewline=computed(()=>draft.value?.newline||(draft.value?.base.includes('\r\n')?'CRLF':'LF'))
const textLimit=computed(()=>{const bytes=draft.value?.maxBytes;return bytes&&bytes>0?`${(bytes/1048576).toFixed(bytes%1048576?1:0)} MiB`:''})
function formatBytes(bytes:number){return bytes>=1048576?`${(bytes/1048576).toFixed(bytes%1048576?1:0)} MiB`:bytes>=1024?`${(bytes/1024).toFixed(bytes%1024?1:0)} KiB`:`${bytes} B`}
const historyStatus=computed(()=>{const bytes=draft.value?.historyBytes||0,budget=editorHistoryBudgetBytes(draft.value);return `撤销历史 ${formatBytes(bytes)} / ${formatBytes(budget)}${draft.value?.historyTrimmed?' · 较旧记录已按预算清理':''}`})
const canUndo=computed(()=>!!draft.value&&(draft.value.cursor||0)>0&&!effectiveReadonly.value),canRedo=computed(()=>!!draft.value&&draft.value.cursor<draft.value.history.length&&!effectiveReadonly.value)
const sourcePath=computed(()=>String(view.value.state.path||'')),sourceName=computed(()=>sourcePath.value.split('/').pop()||'原文件')
const sourceDownloadHref=computed(()=>downloadSourceAfterOpenError.value&&view.value.resourceRef&&sourcePath.value?`/api/v1/instances/${encodeURIComponent(resourceInstanceId.value)}/files/download?path=${encodeURIComponent(sourcePath.value)}`:'')
const preview=ref<PreviewContent>(),previewMode=ref(false),previewLoading=ref(false),previewError=ref(''),previewOffsets=ref<number[]>([0]),previewCursor=ref(0)
const previewRange=computed(()=>preview.value?`${formatBytes(preview.value.offset)}–${formatBytes(preview.value.nextOffset)} / ${formatBytes(preview.value.total)}`:'')
async function loadPreview(offset=0,remember=true):Promise<boolean>{
  if(!resourceInstanceId.value||!sourcePath.value)return false
  previewLoading.value=true;previewError.value=''
  try{
    const version=preview.value?.version
    const params=new URLSearchParams({path:sourcePath.value,offset:String(offset),limit:String(48<<10)})
    if(version)params.set('version',version)
    const page=await api<PreviewContent>(`/instances/${encodeURIComponent(resourceInstanceId.value)}/files/preview?${params.toString()}`)
    preview.value=page
    if(remember){
      const history=previewOffsets.value.slice(0,previewCursor.value+1),pageOffset=Number(page.offset)
      if(history.at(-1)!==pageOffset)history.push(pageOffset)
      previewOffsets.value=history;previewCursor.value=history.length-1
    }
    return true
  }catch(e){previewError.value=e instanceof Error?e.message:String(e);return false}
  finally{previewLoading.value=false}
}
async function previewNext(){if(preview.value?.hasMore&&!previewLoading.value)await loadPreview(preview.value.nextOffset)}
async function previewPrevious(){
  if(previewCursor.value<=0||previewLoading.value)return
  const target=previewOffsets.value[previewCursor.value-1]
  if(await loadPreview(target,false))previewCursor.value--
}
const documentLocation=computed(()=>{
  if(!draft.value?.resourceRef)return draft.value?.path||'本地未命名草稿'
  const access=fileAccess.data.value,node=access?.nodeName||access?.nodeId||draft.value.resourceRef.nodeId||'节点',nodeState=access?.nodeState&&access.nodeState!=='ONLINE'?` · ${access.nodeState}`:''
  const capability=fileAccess.isError.value?'权限状态不可用':access?.canWrite===true?(readonly.value?'只读模式':'可读写'):access?.canWrite===false?'只读':'权限检查中'
  return `${node}${nodeState} · ${draft.value.path||'未知路径'} · ${capability}`
})
let editor:monaco.editor.IStandaloneCodeEditor|undefined,splitEditor:monaco.editor.IStandaloneCodeEditor|undefined,activeEditor:monaco.editor.IStandaloneCodeEditor|undefined,observer:ResizeObserver|undefined,saveTimer:ReturnType<typeof setTimeout>|undefined,disposed=false
const disposables:monaco.IDisposable[]=[],splitDisposables:monaco.IDisposable[]=[]
function editorOptions(model:monaco.editor.ITextModel,ariaLabel='文件正文编辑器'):monaco.editor.IStandaloneEditorConstructionOptions{return{model,readOnly:effectiveReadonly.value,theme:desktop.state?.preferences.theme==='dark'?'vs-dark':'vs',fontFamily:'"SFMono-Regular",Consolas,"Liberation Mono",monospace',fontSize:14,minimap:{enabled:false},automaticLayout:false,scrollBeyondLastLine:false,padding:{top:16},tabSize:2,ariaLabel}}
function bindEditor(instance:monaco.editor.IStandaloneCodeEditor,stateKey:string,bag:monaco.IDisposable[]){
  const capture=()=>patchView(stateKey,instance.saveViewState())
  bag.push(instance.onDidChangeCursorSelection(capture),instance.onDidScrollChange(capture),instance.onDidChangeModelDecorations(()=>{if(instance.hasTextFocus())capture()}),instance.onDidFocusEditorText(()=>{activeEditor=instance}))
  instance.addCommand(monaco.KeyMod.CtrlCmd|monaco.KeyCode.KeyS,save)
  instance.addCommand(monaco.KeyMod.CtrlCmd|monaco.KeyCode.KeyF,showFind)
  instance.addCommand(monaco.KeyMod.CtrlCmd|monaco.KeyCode.KeyH,showFind)
}
function layoutEditors(){editor?.layout();splitEditor?.layout()}
function mountSplitEditor(){
  if(!splitEnabled.value||splitEditor||!splitContainer.value)return
  const model=editor?.getModel();if(!model)return
  splitEditor=monaco.editor.create(splitContainer.value,editorOptions(model,'文件正文编辑器分栏'))
  const saved=view.value.state.editorSplitView as unknown as monaco.editor.ICodeEditorViewState|undefined
  if(saved)splitEditor.restoreViewState(saved)
  bindEditor(splitEditor,'editorSplitView',splitDisposables)
  observer?.observe(splitContainer.value)
  splitEditor.layout()
}
function unmountSplitEditor(){
  if(!splitEditor)return
  patchView('editorSplitView',splitEditor.saveViewState())
  if(activeEditor===splitEditor)activeEditor=editor
  observer?.unobserve(splitContainer.value!)
  splitDisposables.splice(0).forEach(disposable=>disposable.dispose())
  splitEditor.dispose();splitEditor=undefined
}
async function initialize(){
  error.value='';downloadSourceAfterOpenError.value=false;preview.value=undefined;previewMode.value=false;previewError.value='';previewOffsets.value=[0];previewCursor.value=0
  try {
    if(!draftId.value){
      let existing=Object.values(recovery.state.drafts).find(d=>view.value.resourceRef && d.resourceRef?.kind===view.value.resourceRef.kind && d.resourceRef?.id===view.value.resourceRef.id && d.resourceRef?.nodeId===view.value.resourceRef.nodeId)
      if(!existing && view.value.resourceRef){
        const resource=view.value.resourceRef,path=String(view.value.state.path||resource.id)
        const file=view.value.state.newFile?{text:'',version:'missing',encoding:'UTF-8',newline:'LF'}:await api<DocumentContent>(`/instances/${encodeURIComponent(resourceInstanceId.value||resource.id)}/files/content?path=${encodeURIComponent(path)}`)
        const hasBOM=file.text.startsWith('\ufeff'),format={encoding:file.encoding||(hasBOM?'UTF-8 BOM':'UTF-8'),newline:file.newline||(file.text.includes('\r\n')?'CRLF':'LF'),maxBytes:file.maxBytes}
        patchView('draftId',createDraft(recovery,editableText(file.text),resource,path,file.version,format))
      }else patchView('draftId',existing?.draftId||createDraft(recovery))
    }
    const model=await documentModel(recovery,draftId.value)
    if(!primaryContainer.value)return
    editor=monaco.editor.create(primaryContainer.value,editorOptions(model));activeEditor=editor
    const saved=view.value.state.editorView as unknown as monaco.editor.ICodeEditorViewState
    if(saved)editor.restoreViewState(saved)
    bindEditor(editor,'editorView',disposables)
    observer=new ResizeObserver(layoutEditors);observer.observe(primaryContainer.value)
    ready.value=true
    if(splitEnabled.value){await nextTick();mountSplitEditor()}
    if(draft.value?.pendingSave)void reconcileSave()
    if(props.visible)editor.focus()
  }catch(e){
    error.value=e instanceof Error?e.message:String(e)
    downloadSourceAfterOpenError.value=e instanceof APIError&&(e.code==='TEXT_LIMIT'||e.code==='UNSUPPORTED_ENCODING')
    if(e instanceof APIError&&e.code==='TEXT_LIMIT'){
      previewMode.value=true
      if(await loadPreview(0))error.value=`${e.message} · 当前以只读分段预览显示；如需完整内容请下载原文件。`
      ready.value=true
      return
    }
  }
}
onMounted(initialize)
watch(splitEnabled,async enabled=>{if(!ready.value)return;if(enabled){await nextTick();mountSplitEditor()}else unmountSplitEditor()})
watch(()=>props.visible,visible=>{if(visible)requestAnimationFrame(layoutEditors)})
watch(effectiveReadonly,value=>{editor?.updateOptions({readOnly:value});splitEditor?.updateOptions({readOnly:value})})
watch(()=>desktop.state?.preferences.theme,theme=>monaco.editor.setTheme(theme==='dark'?'vs-dark':'vs'),{immediate:true})
watch(()=>draft.value?.pendingSave?.taskId,()=>{if(ready.value&&draft.value?.pendingSave)void reconcileSave()})
onBeforeUnmount(()=>{disposed=true;if(saveTimer)clearTimeout(saveTimer);if(editor)patchView('editorView',editor.saveViewState());unmountSplitEditor();disposables.forEach(d=>d.dispose());observer?.disconnect();editor?.dispose()})
async function save(){
  if(!draft.value)return
  if(draft.value.resourceRef&&effectiveReadonly.value){error.value=fileAccess.isError.value?'权限状态不可用，未提交写入':'此文件当前为只读状态，未提交写入';return}
  if(!draft.value.resourceRef){exportText();return}
  busy.value=true;error.value=''
  const currentId=draftId.value
  const captured:PendingSave=savePending.value?draft.value.pendingSave!:{requestId:randomUUID(),text:documentBody(draft.value.text,draft.value.encoding),editorText:draft.value.text,version:draft.value.baseVersion||'missing',path:draft.value.path!,instanceId:resourceInstanceId.value,state:'SUBMITTING'}
  recovery.commit([{kind:'set',path:['drafts',currentId,'pendingSave'],value:json(captured)}])
  try{
    const {task}=await api<{task:Task}>(`/instances/${encodeURIComponent(captured.instanceId)}/files/content`,{method:'PUT',headers:{'Idempotency-Key':captured.requestId},body:JSON.stringify({path:captured.path,text:captured.text,version:captured.version})})
    if(session.user?.userId!==recovery.userId)return
    recovery.commit([{kind:'set',path:['drafts',currentId,'pendingSave'],value:json({...captured,taskId:task.taskId,state:task.state})}])
    if(!disposed)await reconcileSave()
  }catch(e){error.value=e instanceof Error?e.message:String(e);if(e instanceof APIError&&e.status===403)void fileAccess.refetch()}finally{busy.value=false}
}
async function reconcileSave(){
  if(saveTimer)clearTimeout(saveTimer)
  const pending=draft.value?.pendingSave,currentId=draftId.value;if(!pending||disposed||session.user?.userId!==recovery.userId)return
  try{
    const task=pending.taskId?(await api<{task:Task}>(`/tasks/${encodeURIComponent(pending.taskId)}`)).task:(await api<{items:Task[]}>('/tasks')).items.find(task=>task.requestId===pending.requestId)
    if(disposed||session.user?.userId!==recovery.userId)return
    if(!task){error.value='保存是否已接受尚未确认。可显式重试原请求；当前新输入保留。';return}
    if(task.state==='SUCCEEDED'){
      const version=task.result?.version;if(typeof version!=='string')throw new Error('保存任务未提供已提交文件版本，请在任务中心核对')
      recovery.commit([{kind:'set',path:['drafts',currentId,'savedText'],value:pending.editorText??editableText(pending.text)},{kind:'set',path:['drafts',currentId,'baseVersion'],value:version},{kind:'delete',path:['drafts',currentId,'pendingSave']}]);return
    }
    recovery.commit([{kind:'set',path:['drafts',currentId,'pendingSave'],value:json({...pending,taskId:task.taskId,state:task.state,error:task.error})}])
    if(['FAILED','CANCELLED','INTERRUPTED'].includes(task.state)){error.value=task.error||`保存任务${task.state}`;return}
    saveTimer=setTimeout(()=>void reconcileSave(),1500)
  }catch(e){error.value=String(e);if(!disposed)saveTimer=setTimeout(()=>void reconcileSave(),3000)}
}
async function compareServer(){
  if(!draft.value?.resourceRef)return
  try{const file=await api<DocumentContent>(`/instances/${encodeURIComponent(resourceInstanceId.value)}/files/content?path=${encodeURIComponent(draft.value.path!)}`);conflict.value=file}catch(e){error.value=String(e)}
}
async function prepareReload(){
  if(!draft.value?.resourceRef||savePending.value)return;error.value=''
  try{const file=await api<DocumentContent>(`/instances/${encodeURIComponent(resourceInstanceId.value)}/files/content?path=${encodeURIComponent(draft.value.path!)}`);reloadForm.value={...file,path:draft.value.path!}}catch(e){error.value=String(e)}
}
function updateFormat(targetDraftId:string,file:DocumentContent){
  const changes=[
    {kind:'set' as const,path:['drafts',targetDraftId,'encoding'],value:file.encoding||(file.text.startsWith('\ufeff')?'UTF-8 BOM':'UTF-8')},
    {kind:'set' as const,path:['drafts',targetDraftId,'newline'],value:file.newline||(file.text.includes('\r\n')?'CRLF':'LF')}
  ]
  recovery.commit(changes)
  if(typeof file.maxBytes==='number')recovery.commit([{kind:'set',path:['drafts',targetDraftId,'maxBytes'],value:file.maxBytes}])
}
function applyReload(){
  const fixed=reloadForm.value,model=editor?.getModel();if(!fixed||!model||!draft.value||savePending.value||effectiveReadonly.value)return
  const text=editableText(fixed.text)
  editor!.pushUndoStop();editor!.executeEdits('blora.reload',[{range:model.getFullModelRange(),text}]);editor!.pushUndoStop()
  recovery.commit([{kind:'set',path:['drafts',draftId.value,'baseVersion'],value:fixed.version},{kind:'set',path:['drafts',draftId.value,'savedText'],value:text},{kind:'delete',path:['drafts',draftId.value,'pendingSave']}]);updateFormat(draftId.value,fixed);reloadForm.value=undefined;conflict.value=undefined;error.value=''
}
function prepareSaveAs(){saveAsForm.value={instanceId:resourceInstanceId.value,path:draft.value?.path?draft.value.path+'.copy':'未命名.txt',overwrite:false,requestId:randomUUID()}}
function saveAsField(key:'instanceId'|'path'|'overwrite',value:string|boolean){if(saveAsForm.value)saveAsForm.value={...saveAsForm.value,[key]:value}}
async function saveAs(){
  if(!draft.value||!saveAsForm.value||busy.value)return;busy.value=true;error.value='';const fixed=copy(saveAsForm.value),source=copy(draft.value)
  try{
    const destination=destinations.data.value?.items.find(instance=>instance.instanceId===fixed.instanceId);if(!destination)throw new Error('请选择可见目标实例，写入权限仍由服务器重新校验')
    const resource={kind:'file',id:`${fixed.instanceId}:${fixed.path}`,nodeId:destination.nodeId}
    if(Object.values(recovery.state.drafts).some(draft=>draft.resourceRef?.id===resource.id&&draft.resourceRef?.nodeId===resource.nodeId))throw new Error('目标已经有工作区草稿，请切换到该草稿处理，或选择其他路径')
    let version='missing'
    try{version=(await api<{version:string}>(`/instances/${encodeURIComponent(fixed.instanceId)}/files/stat?path=${encodeURIComponent(fixed.path)}`)).version;if(!fixed.overwrite)throw new Error('目标已存在，请改名或明确允许覆盖当前版本')}catch(e){if(!(e instanceof APIError&&e.status===404))throw e}
    const newId=id('draft'),pending:PendingSave={instanceId:fixed.instanceId,path:fixed.path,version,text:documentBody(source.text,source.encoding),editorText:source.text,requestId:fixed.requestId,state:'SUBMITTING'}
    const next:Draft={...source,draftId:newId,resourceRef:resource,path:fixed.path,baseVersion:version,pendingSave:pending}
    recovery.commit([{kind:'set',path:['drafts',newId],value:json(next)}]);saveAsForm.value=undefined
    desktop.open({appId:'blora.editor',resourceRef:resource,title:fixed.path.split('/').pop()!,disposition:'new-window',state:{draftId:newId,instanceId:fixed.instanceId,path:fixed.path}})
    try{const {task}=await api<{task:Task}>(`/instances/${encodeURIComponent(fixed.instanceId)}/files/content`,{method:'PUT',headers:{'Idempotency-Key':pending.requestId},body:JSON.stringify({path:pending.path,text:pending.text,version:pending.version})});if(session.user?.userId===recovery.userId)recovery.commit([{kind:'set',path:['drafts',newId,'pendingSave'],value:json({...pending,taskId:task.taskId,state:task.state})}])}
    catch(e){if(session.user?.userId===recovery.userId)recovery.commit([{kind:'set',path:['drafts',newId,'pendingSave'],value:json({...pending,error:String(e)})}])}
  }catch(e){error.value=String(e)}finally{busy.value=false}
}
function useComparedVersion(){if(!conflict.value||!draft.value||savePending.value)return;recovery.commit([{kind:'set',path:['drafts',draftId.value,'baseVersion'],value:conflict.value.version},{kind:'set',path:['drafts',draftId.value,'savedText'],value:editableText(conflict.value.text)},{kind:'delete',path:['drafts',draftId.value,'pendingSave']}]);updateFormat(draftId.value,conflict.value);conflict.value=undefined;error.value=''}
function exportText(){if(!draft.value)return;const url=URL.createObjectURL(new Blob([documentBody(draft.value.text,draft.value.encoding)],{type:'text/plain;charset=utf-8'}));const anchor=document.createElement('a');anchor.href=url;anchor.download=draft.value.path?.split('/').pop()||'未命名.txt';anchor.click();setTimeout(()=>URL.revokeObjectURL(url),1000)}
function undo(){const target=activeEditor||editor;if(!effectiveReadonly.value)void target?.getModel()?.undo();target?.focus()}
function redo(){const target=activeEditor||editor;if(!effectiveReadonly.value)void target?.getModel()?.redo();target?.focus()}
async function showFind(){
  const previousFocus=document.activeElement
  findOpen.value=true
  await nextTick()
  // Wait only for the input's DOM to exist. A delayed animation-frame focus
  // could steal typing after the user already moved to the replacement field.
  if(disposed||!findOpen.value||!findInput.value)return
  if(document.activeElement!==previousFocus&&document.activeElement!==document.body&&document.activeElement!==findInput.value)return
  findInput.value.focus();findInput.value.select()
}
function matches(){try{const result=editor?.getModel()?.findMatches(findText.value,false,findRegex.value,findCase.value,null,true)||[];findMessage.value=`${result.length} 处匹配`;return result}catch(e){findMessage.value=String(e);return []}}
function findNext(){const target=activeEditor||editor,found=matches();if(!found.length)return;const position=target?.getSelection()?.getEndPosition()||new monaco.Position(1,1);const next=found.find(match=>monaco.Position.isBeforeOrEqual(position,match.range.getStartPosition()))||found[0]!;target?.setSelection(next.range);target?.revealRangeInCenterIfOutsideViewport(next.range)}
function replaceAll(){const target=activeEditor||editor;if(effectiveReadonly.value)return;const found=matches();if(!found.length)return;target?.pushUndoStop();target?.executeEdits('blora.replace',found.map(match=>({range:match.range,text:replaceText.value})));target?.pushUndoStop();target?.focus()}
function selectMatches(){const target=activeEditor||editor,found=matches();if(!found.length)return;target?.setSelections(found.map(match=>new monaco.Selection(match.range.startLineNumber,match.range.startColumn,match.range.endLineNumber,match.range.endColumn)));target?.focus()}
</script>
<template>
  <div class="editor-app">
    <div class="editor-toolbar"><span>{{documentLocation}}{{dirty?' ●':''}}</span><button :disabled="!ready||!canUndo" @click="undo">↶ 撤销</button><button :disabled="!ready||!canRedo" @click="redo">↷ 重做</button><button :disabled="previewMode" @click="showFind">查找</button><button v-if="!previewMode" @click="exportText">导出</button><button :disabled="!ready||previewMode" :aria-pressed="splitEnabled" @click="splitEnabled=!splitEnabled">{{splitEnabled?'关闭分栏':'分栏'}}</button><button :disabled="!ready||previewMode" @click="prepareSaveAs">另存为…</button><button :disabled="previewMode||fileAccess.data.value?.canWrite===false||fileAccess.isError.value" @click="readonly=!readonly">{{fileAccess.isError.value?'权限状态不可用':fileAccess.data.value?.canWrite===false?'只读权限':readonly?'允许编辑':'只读查看'}}</button><button v-if="draft?.resourceRef" :disabled="savePending||effectiveReadonly" @click="prepareReload">重新读取…</button><button v-if="draft?.resourceRef" @click="compareServer">比较服务器版本</button><button class="primary" :disabled="previewMode||busy || !ready || (!!draft?.resourceRef&&effectiveReadonly) || (savePending && !!draft?.pendingSave?.taskId && saveState!=='WAITING_CLIENT')" @click="save">{{draft?.resourceRef?(effectiveReadonly?(fileAccess.isError.value?'权限状态不可用':fileAccess.data.value?.canWrite===false?'只读权限':'只读模式'):busy?'提交中…':savePending?'重试原保存':'保存到服务器'):'导出正文'}}</button></div>
    <form v-if="findOpen&&!previewMode" class="editor-find" aria-label="查找替换" @submit.prevent="findNext" @keydown.esc.prevent="findOpen=false;editor?.focus()">
      <input ref="findInput" v-model="findText" aria-label="查找内容" placeholder="查找" @input="matches()"><input v-model="replaceText" aria-label="替换内容" placeholder="替换为">
      <label><input v-model="findCase" type="checkbox">区分大小写</label><label><input v-model="findRegex" type="checkbox">正则</label>
      <button>下一个</button><button type="button" @click="selectMatches">选择全部匹配</button><button type="button" @click="replaceAll">全部替换</button><button type="button" aria-label="关闭查找" @click="findOpen=false;editor?.focus()">×</button><small>{{findMessage}}</small>
    </form>
    <p v-if="error||draft?.pendingSave?.error" class="error notice"><span>{{error||draft?.pendingSave?.error}}</span><a v-if="sourceDownloadHref" :href="sourceDownloadHref" :download="sourceName">下载原文件</a></p>
    <div v-if="conflict" class="editor-conflict"><strong>服务器当前版本（只读比较） · {{conflict.encoding||'UTF-8'}} · {{conflict.newline||'LF'}}</strong><textarea :value="editableText(conflict.text)" readonly aria-label="服务器版本正文"></textarea><p>先在编辑器中合并需要保留的内容，再采用此版本为保存基线。此操作不保存到服务器。</p><button :disabled="savePending" @click="useComparedVersion">合并后采用此基线</button><button @click="conflict=undefined">收起比较</button></div>
    <section v-if="previewMode" class="editor-preview" aria-label="只读分段预览">
      <header class="editor-preview-toolbar"><div><strong>只读分段预览</strong><span>{{previewRange}}</span></div><div class="action-row"><button :disabled="previewLoading||previewCursor<=0" @click="previewPrevious">上一段</button><button :disabled="previewLoading||!preview?.hasMore" @click="previewNext">下一段</button></div></header>
      <p v-if="previewError" class="error notice">{{previewError}} <a v-if="sourceDownloadHref" :href="sourceDownloadHref" :download="sourceName">下载原文件</a></p>
      <pre v-if="preview" class="editor-preview-text" tabindex="0">{{preview.text}}</pre>
      <div v-else class="empty-state"><strong>暂时无法读取分段</strong><p>文件仍保持只读状态，可下载原文件后使用专用工具打开。</p></div>
    </section>
    <div v-else class="editor-panes" :class="{split:splitEnabled}"><div ref="primaryContainer" class="monaco-container"></div><div v-if="splitEnabled" ref="splitContainer" class="monaco-container"></div></div>
    <form v-if="saveAsForm" class="in-window-dialog" role="dialog" aria-label="另存正文副本" @submit.prevent="saveAs"><h3>另存正文副本</h3><label>目标实例<select :value="saveAsForm.instanceId" aria-label="另存目标实例" required @change="saveAsField('instanceId',($event.target as HTMLSelectElement).value)"><option value="">选择目标实例</option><option v-for="instance in destinations.data.value?.items" :key="instance.instanceId" :value="instance.instanceId">{{instance.nodeName||instance.nodeId}} / {{instance.name}}</option></select></label><label>目标路径<input :value="saveAsForm.path" aria-label="另存目标路径" required @input="saveAsField('path',($event.target as HTMLInputElement).value)"></label><label class="check-row"><input type="checkbox" :checked="saveAsForm.overwrite" @change="saveAsField('overwrite',($event.target as HTMLInputElement).checked)">允许覆盖目标当前版本</label><p>当前正文和撤销历史复制到新文件视图；来源草稿保留。提交按目标版本校验，等待节点任务确认。</p><p v-if="destinations.error.value" class="error">{{destinations.error.value.message}}</p><div class="action-row"><button type="button" :disabled="busy" @click="saveAsForm=undefined">返回</button><button class="primary" :disabled="busy">确认另存正文</button></div></form>
    <div v-if="reloadForm" class="in-window-dialog" role="dialog" aria-label="重新读取服务器正文"><h3>重新读取服务器正文</h3><p>{{reloadForm.path}} · {{reloadForm.text.length}} 字符</p><p>将用刚读取的服务器版本替换当前共享草稿正文，所有打开此草稿的视图都会更新。原本地正文保留在撤销历史中，可撤销恢复；不会提交服务器写入。</p><div class="action-row"><button @click="reloadForm=undefined">返回</button><button class="primary" :disabled="savePending||readonly" @click="applyReload">采用服务器正文（可撤销）</button></div></div>
    <footer class="editor-status"><span :class="{'error':!recovery.status.protected}">{{recovery.status.message}}</span><span v-if="previewMode">只读分段预览 · {{previewRange}}{{preview?.encoding?' · '+preview.encoding:''}}</span><span v-else>{{draft?.text.length || 0}} 字符 · {{documentEncoding}} · {{documentNewline}}{{textLimit?' · 上限 '+textLimit:''}} · {{historyStatus}} · {{savePending?'保存任务 '+saveState:dirty?'服务器尚未保存':'基线未更改'}}</span></footer>
  </div>
</template>
