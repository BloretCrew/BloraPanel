<script setup lang="ts">
import {computed,onMounted,onBeforeUnmount,ref,watch} from 'vue'
import {useDesktop} from '../desktop/store'
import {createDraft,documentModel,monaco} from '../services/documents'
import {saveCompose,reconcileComposeSave,readComposeSource,pauseComposeSave} from '../services/compose'
import {json} from '../recovery/state'
const props=defineProps<{viewTabId:string;nodeId:string;projectId:string;revision:number}>(),desktop=useDesktop(),recovery=desktop.recovery!,element=ref<HTMLElement>(),error=ref(''),busy=ref(false)
const view=computed(()=>recovery.state.views[props.viewTabId]),draftId=computed(()=>String(view.value?.state.composeDraftId||'')),draft=computed(()=>recovery.state.drafts[draftId.value])
let editor:monaco.editor.IStandaloneCodeEditor|undefined,observer:ResizeObserver|undefined,disposed=false
function patch(key:string,value:unknown){if(view.value)recovery.commit([{kind:'set',path:['views',props.viewTabId,'state',key],value:json(value)}])}
onMounted(async()=>{
  try{
    if(!draft.value){
      const resource={kind:'compose',id:props.projectId,nodeId:props.nodeId}
      const existing=Object.values(recovery.state.drafts).find(d=>d.resourceRef?.kind==='compose'&&d.resourceRef.id===props.projectId&&d.resourceRef.nodeId===props.nodeId)
      const text=existing?'':props.revision?await readComposeSource(props.nodeId,props.projectId,props.revision):'services:\n  app:\n    image: busybox:latest\n    command: ["sleep", "3600"]\n'
      if(disposed)return
      patch('composeDraftId',existing?.draftId||createDraft(recovery,text,resource,'compose.yaml',String(props.revision)))
    }
    const model=await documentModel(recovery,draftId.value)
    if(disposed||!element.value)return
    editor=monaco.editor.create(element.value,{model,theme:desktop.state?.preferences.theme==='dark'?'vs-dark':'vs',minimap:{enabled:false},ariaLabel:'Compose 配置正文',scrollBeyondLastLine:false})
    const saved=view.value?.state.composeEditorView
    if(saved)editor.restoreViewState(saved as unknown as monaco.editor.ICodeEditorViewState)
    const capture=()=>patch('composeEditorView',editor?.saveViewState())
    editor.onDidChangeCursorSelection(capture)
    editor.onDidScrollChange(capture)
    editor.addCommand(monaco.KeyMod.CtrlCmd|monaco.KeyCode.KeyS,save)
    observer=new ResizeObserver(()=>editor?.layout())
    observer.observe(element.value)
    if(draft.value?.composeSave)void reconcileComposeSave(recovery,draftId.value)
  }catch(e){error.value=String(e)}
})
watch(()=>desktop.state?.preferences.theme,theme=>monaco.editor.setTheme(theme==='dark'?'vs-dark':'vs'),{immediate:true})
async function save(){busy.value=true;error.value='';try{await saveCompose(recovery,draftId.value,props.nodeId,props.projectId)}catch(e){error.value=String(e)}finally{busy.value=false}}
onBeforeUnmount(()=>{disposed=true;patch('composeEditorView',editor?.saveViewState());observer?.disconnect();editor?.dispose()})
</script>
<template><section class="compose-editor"><div class="action-row"><strong>{{projectId}} · 已保存版本 {{draft?.baseVersion||revision}}</strong><span>{{draft?.text!==draft?.savedText?'有未保存修改':'正文已保存'}}</span><button :disabled="busy" @click="save">{{draft?.composeSave?'继续原保存':'保存配置'}}</button><button v-if="draft?.composeSave" @click="pauseComposeSave(recovery,draftId)">暂停发送</button></div><p class="muted">保存只记录配置版本。返回项目列表后，选择明确版本执行应用；刷新不会部署。</p><p v-if="error||draft?.composeSave?.error" class="error" role="alert">{{error||draft?.composeSave?.error}}</p><p v-if="draft?.composeSave" role="status">保存 {{draft.composeSave.state}} · {{draft.composeSave.offset}} / {{draft.composeSave.total}} 字节</p><div ref="element" class="compose-monaco"></div></section></template>
<style scoped>.compose-editor{display:flex;flex-direction:column;min-height:400px;flex:1}.compose-monaco{flex:1;min-height:330px}</style>
