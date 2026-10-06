import * as monaco from 'monaco-editor/editor/editor.api'
import 'monaco-editor/basic-languages/monaco.contribution'
import 'monaco-editor/language/json/monaco.contribution'
import EditorWorker from 'monaco-editor/editor/editor.worker?worker'
import JsonWorker from 'monaco-editor/language/json/json.worker?worker'
import { id, type Draft, type ResourceRef } from '../app-host/types'
import { editText, inverseEdits, json } from '../recovery/state'
import type { RecoveryService } from '../recovery/service'
import { documentLanguage } from './document-language'

self.MonacoEnvironment = { getWorker(_id,label) { return label === 'json' ? new JsonWorker() : new EditorWorker() } }
export { monaco }
type DocumentModel = { model: monaco.editor.ITextModel; recovering: boolean; group?:string;historyEpoch:number;historyRebuildPending:boolean }
const models = new Map<string, DocumentModel>()
export function releaseDocuments(userId:string) {
  for(const [key,entry] of models)if(key.startsWith(userId+':')){entry.model.dispose();models.delete(key)}
}
export function createDraft(recovery:RecoveryService,text='',resourceRef?:ResourceRef,path?:string,baseVersion?:string,format?:Pick<Draft,'encoding'|'newline'|'maxBytes'>) {
  const draftId=id('draft')
  const draft:Draft={draftId,resourceRef,path,baseVersion,base:text,text,history:[],cursor:0,savedText:text,historyBytes:0,historyEpoch:0,historyTrimmed:false,...format}
  recovery.commit([{kind:'set',path:['drafts',draftId],value:json(draft)}])
  return draftId
}
function replayEdits(model:monaco.editor.ITextModel,draft:Draft){
  for(const [index,step] of draft.history.entries()){
    const operations=step.forward.map(edit=>({range:monaco.Range.fromPositions(model.getPositionAt(edit.offset),model.getPositionAt(edit.offset+edit.length)),text:edit.text}))
    model.pushEditOperations([],operations,()=>[])
    // A native undo can stop inside a recorded group (including old records
    // with identity edits). Keep the persisted cursor a stack boundary.
    if(index+1===draft.cursor||!step.group||step.group!==draft.history[index+1]?.group)model.pushStackElement()
  }
}
async function restoreHistory(model:monaco.editor.ITextModel,draft:Draft){
  replayEdits(model,draft)
  for(let index=draft.history.length;index>draft.cursor;){const group=draft.history[index-1]!.group;index--;while(group&&index>draft.cursor&&draft.history[index-1]!.group===group)index--;await model.undo()}
  if(model.getValue()!==draft.text)throw new Error('草稿历史与正文不一致，原始记录已保留，请导出恢复记录')
}
async function rebuildHistory(recovery:RecoveryService,draftId:string,entry:DocumentModel){
  const draft=recovery.state.drafts[draftId],model=entry.model
  if(!draft||model.isDisposed())return
  const viewStates=monaco.editor.getEditors().filter(editor=>editor.getModel()===model).map(editor=>[editor,editor.saveViewState()] as const)
  entry.recovering=true
  try{
    model.setValue(draft.base.startsWith('\ufeff')?draft.base.slice(1):draft.base)
    model.setEOL((draft.newline|| (draft.base.includes('\r\n')?'CRLF':'LF'))==='CRLF'?monaco.editor.EndOfLineSequence.CRLF:monaco.editor.EndOfLineSequence.LF)
    await restoreHistory(model,draft)
    for(const [editor,state] of viewStates)editor.restoreViewState(state)
    entry.historyEpoch=draft.historyEpoch||0
  }catch(error){recovery.fail(error)}
  finally{entry.recovering=false}
}
export async function documentModel(recovery:RecoveryService,draftId:string) {
  const cacheKey=`${recovery.key}:${draftId}`
  const cached=models.get(cacheKey);if(cached)return cached.model
  const draft=recovery.state.drafts[draftId];if(!draft)throw new Error('草稿正文缺失，未用服务器文件替代')
  const language=documentLanguage(draft.path)
  const modelBase=draft.base.startsWith('\ufeff')?draft.base.slice(1):draft.base
  const model=monaco.editor.createModel(modelBase,language,monaco.Uri.parse(`blora://draft/${encodeURIComponent(recovery.key)}/${draftId}/${encodeURIComponent(draft.path||'untitled.txt')}`))
  // Monaco's default EOL depends on the browser platform (CRLF on Windows).
  // Explicitly restore either LF or CRLF from the document before replaying
  // before replaying offset-based edits so CRLF offsets and the protected body
  // remain byte-for-byte aligned.
  model.setEOL((draft.newline|| (draft.base.includes('\r\n')?'CRLF':'LF'))==='CRLF'?monaco.editor.EndOfLineSequence.CRLF:monaco.editor.EndOfLineSequence.LF)
  const entry:DocumentModel={model,recovering:true,historyEpoch:draft.historyEpoch||0,historyRebuildPending:false};models.set(cacheKey,entry)
  try{await restoreHistory(model,draft)}catch(error){model.dispose();models.delete(cacheKey);throw error}
  entry.recovering=false
  model.onDidChangeContent(event=>{
    if(entry.recovering)return
    const current=recovery.state.drafts[draftId]!
    if(event.isUndoing || event.isRedoing){
      let cursor=current.cursor,text=current.text
      const expected=model.getValue()
      while(text!==expected && (event.isUndoing?cursor>0:cursor<current.history.length)){
        text=event.isUndoing?editText(text,current.history[--cursor]!.reverse):editText(text,current.history[cursor++]!.forward)
      }
      if(text!==expected){recovery.fail(new Error('无法核对编辑器撤销位置，原正文与日志已保留'));return}
      recovery.commit([{kind:'history',draftId,cursor}])
    }
    else {
      const previousHistoryEpoch=current.historyEpoch||0
      const forward=event.changes.map(change=>({offset:change.rangeOffset,length:change.rangeLength,text:change.text}))
      // Firefox's textarea input path may emit an identity replacement after
      // the semantic edit. It must not truncate redo history or form an undo step.
      if(editText(current.text,forward)===current.text)return
      entry.group ||= id('edit')
      recovery.commit([{kind:'edit',draftId,forward,reverse:inverseEdits(current.text,forward),group:entry.group}])
      if((recovery.state.drafts[draftId]!.historyEpoch||0)!==previousHistoryEpoch)entry.historyRebuildPending=true
      // The content event fires inside Monaco's edit transaction, before its new
      // stack element is fully installed. Close it at the end of this event turn.
      // The edit delta above is already synchronously protected before this point.
      const group=entry.group
      queueMicrotask(()=>{
        if(!group||entry.group!==group)return
        entry.group=undefined
        recovery.commit([{kind:'seal-history-group',draftId,group}])
        if(model.isDisposed())return
        if(entry.historyRebuildPending||(recovery.state.drafts[draftId]!.historyEpoch||0)!==entry.historyEpoch){entry.historyRebuildPending=false;void rebuildHistory(recovery,draftId,entry)}
        else model.pushStackElement()
      })
    }
    if(model.getValue()!==recovery.state.drafts[draftId]!.text)recovery.fail(new Error('编辑历史保护状态不一致；请保留页面并导出正文'))
  })
  return model
}
