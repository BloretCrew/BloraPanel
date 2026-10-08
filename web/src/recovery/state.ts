import type { Draft, Json, TextEdit, Workspace, TerminalJournalEvent } from '../app-host/types'
import {markRaw} from 'vue'

export const TERMINAL_OUTPUT_JOURNAL_BYTES = 128 * 1024
export const TERMINAL_OUTPUT_JOURNAL_EVENTS = 128
export const EDITOR_HISTORY_BUDGET_BYTES = 8 * 1024 * 1024
const historyEncoder = new TextEncoder()
export function editorHistoryBudgetBytes(draft?:Pick<Draft,'maxBytes'>){
  const editableLimit=draft?.maxBytes&&Number.isSafeInteger(draft.maxBytes)&&draft.maxBytes>0?draft.maxBytes:4*1024*1024
  return Math.min(EDITOR_HISTORY_BUDGET_BYTES,Math.max(1024,editableLimit*2))
}
export type Mutation = { kind: 'set'; path: string[]; value: Json } | { kind: 'delete'; path: string[] } | { kind: 'edit'; draftId: string; forward: TextEdit[]; reverse: TextEdit[];group?:string } | { kind: 'seal-history-group'; draftId:string; group:string } | { kind: 'history'; draftId: string; cursor: number } | {kind:'terminal-output';checkpointKey:string;baseSequence:number;previousSequence:number;sequence:number;events:TerminalJournalEvent[];scroll:number}
export function terminalJournalEvent(event:TerminalJournalEvent):TerminalJournalEvent{
  if(!Number.isSafeInteger(event.sequence)||event.sequence<1)throw new Error('终端本地输出日志序号无效')
  if(event.kind==='resize'){
    if(!Number.isInteger(event.cols)||!Number.isInteger(event.rows)||event.cols!<1||event.rows!<1||event.cols!>1000||event.rows!>1000)throw new Error('终端本地尺寸日志无效')
    return {sequence:event.sequence,kind:'resize',cols:event.cols,rows:event.rows}
  }
  if((event.kind!==undefined&&event.kind!=='output')||typeof event.data!=='string')throw new Error('终端本地输出日志内容无效')
  return {sequence:event.sequence,data:event.data}
}
export const terminalJournalBytes=(event:TerminalJournalEvent)=>event.kind==='resize'?64:(event.data?.length||0)*2
export type JournalEntry = { revision: number; mutations: Mutation[] }
// Output owned by recovery is immutable. Cache only the validated summary of
// arrays built here; imported/app-written arrays must pass the complete checks.
// Frozen plain values also avoid Vue wrapping every historical event in a
// proxy on each append. No output is delayed or dropped at this boundary.
const terminalJournals=new WeakMap<readonly TerminalJournalEvent[],{baseSequence:number;sequence:number;bytes:number}>()
function appendTerminalJournal(existing:TerminalJournalEvent[],baseSequence:number,sequence:number,events:TerminalJournalEvent[],nextSequence:number){
  const owned=terminalJournals.get(existing)
  let persistedSequence=baseSequence,existingBytes=0
  if(owned&&owned.baseSequence===baseSequence&&owned.sequence===sequence){persistedSequence=owned.sequence;existingBytes=owned.bytes}
  else for(const event of existing){
    terminalJournalEvent(event)
    if(event.sequence!==++persistedSequence)throw new Error('终端已有输出日志不连续')
    existingBytes+=terminalJournalBytes(event)
  }
  if(persistedSequence!==sequence||existingBytes>TERMINAL_OUTPUT_JOURNAL_BYTES)throw new Error('终端已有输出日志与检查点不匹配')
  let expected=sequence,addedBytes=0
  const added=events.map(event=>{
    const value=terminalJournalEvent(event)
    if(value.sequence!==++expected)throw new Error('终端本地输出日志不连续')
    addedBytes+=terminalJournalBytes(value)
    return Object.freeze(value)
  })
  if(nextSequence!==expected||existingBytes+addedBytes>TERMINAL_OUTPUT_JOURNAL_BYTES)throw new Error('终端本地输出日志超出保护预算')
  const retained=owned?existing:existing.map(event=>Object.freeze(terminalJournalEvent(event)))
  const journal=Object.freeze([...retained,...added])
  terminalJournals.set(journal,{baseSequence,sequence:nextSequence,bytes:existingBytes+addedBytes})
  return journal as unknown as TerminalJournalEvent[]
}
// Recovery values are structured data; avoid serializing large terminal output strings on every journal entry.
// Vue proxies and non-cloneable app values retain the JSON fallback used before this optimization.
export const copy = <T>(value: T): T => {
  try { return structuredClone(value) }
  catch { return JSON.parse(JSON.stringify(value)) as T }
}
// Keep the JSON boundary's sanitizing behavior (for example, removing undefined object properties).
export const json = (value: unknown) => JSON.parse(JSON.stringify(value)) as Json
export function editText(text: string, edits: TextEdit[]): string {
  for (const edit of [...edits].sort((a, b) => b.offset - a.offset)) {
    if (edit.offset < 0 || edit.offset + edit.length > text.length) throw new Error('编辑增量超出正文边界')
    text = text.slice(0, edit.offset) + edit.text + text.slice(edit.offset + edit.length)
  }
  return text
}
export function inverseEdits(text: string, edits: TextEdit[]): TextEdit[] {
  let displacement = 0
  return [...edits].sort((a,b) => a.offset - b.offset).map(edit => {
    const inverse = { offset: edit.offset + displacement, length: edit.text.length, text: text.slice(edit.offset, edit.offset + edit.length) }
    displacement += edit.text.length - edit.length
    return inverse
  })
}
function editStepBytes(step:Draft['history'][number]){return historyEncoder.encode(JSON.stringify(step)).byteLength}
function historyPayloadBytes(history:Draft['history']){return history.reduce((total,step)=>total+editStepBytes(step),0)}
// Committed undo operations are immutable values; only the owning history
// array/cursor changes. Copy before taking ownership so app-supplied operations
// cannot mutate a protected step, and keep Vue from retraversing every old
// offset/text pair on each new input. The marker is non-enumerable and is
// reconstructed after native/JSON cloning when a workspace is restored.
function ownedHistoryStep(step:Draft['history'][number]):Draft['history'][number]{
  const operations=(edits:TextEdit[])=>Object.freeze(edits.map(edit=>Object.freeze({offset:edit.offset,length:edit.length,text:edit.text})))
  return Object.freeze(markRaw({...step,forward:operations(step.forward),reverse:operations(step.reverse)})) as unknown as Draft['history'][number]
}
export function initializeDraftHistory(draft:Draft){
  draft.history=draft.history.map(ownedHistoryStep)
  draft.historyBytes=historyPayloadBytes(draft.history)
  draft.historyEpoch=Number.isSafeInteger(draft.historyEpoch)&&draft.historyEpoch!>=0?draft.historyEpoch:0
  draft.historyTrimmed=!!draft.historyTrimmed
  // A page may have closed while a large input group was still being journaled.
  // In that case its visible text was already rebased and remains safe as a base.
  draft.historyDroppedGroup=undefined
}
function groupEnd(history:Draft['history'],start:number){
  const group=history[start]?.group
  let end=start+1
  if(group)while(end<history.length&&history[end]?.group===group)end++
  return end
}
export function trimDraftHistory(draft:Draft,activeGroup?:string){
  if(draft.historyBytes===undefined)draft.historyBytes=historyPayloadBytes(draft.history)
  const budget=editorHistoryBudgetBytes(draft)
  let bytes=draft.historyBytes,trimmed=false
  while(bytes>budget&&draft.history.length){
    if(draft.cursor>0){
      const end=groupEnd(draft.history,0),removeCount=Math.min(end,draft.history.length)
      if(activeGroup&&draft.history[0]?.group===activeGroup){
        // Never cut the front off a group while Monaco is still producing it.
        // If one transaction alone exceeds the budget, make the visible body its
        // new base so undo can never restore only a suffix of that transaction.
        draft.base=draft.text.startsWith('\ufeff')?draft.text.slice(1):draft.text
        draft.history=[];draft.cursor=0;bytes=0;draft.historyDroppedGroup=activeGroup;trimmed=true;break
      }
      if(end>draft.cursor){
        // A partially undone group cannot be rebased without changing its meaning.
        // Fold the visible text into the new base and discard all dependent history.
        draft.base=draft.text.startsWith('\ufeff')?draft.text.slice(1):draft.text
        draft.history=[];draft.cursor=0;bytes=0;trimmed=true;break
      }
      let base=draft.base.startsWith('\ufeff')?draft.base.slice(1):draft.base
      const removed=draft.history.slice(0,removeCount)
      for(const step of removed)base=editText(base,step.forward)
      draft.base=base
      bytes-=removed.reduce((total,step)=>total+editStepBytes(step),0)
      draft.history.splice(0,removeCount);draft.cursor-=removeCount;trimmed=true
    }else{
      // With the document at its base, retain the nearest redo groups and drop
      // the farthest future group first.
      const last=draft.history.length-1,group=draft.history[last]?.group
      let start=last
      if(group)while(start>0&&draft.history[start-1]?.group===group)start--
      const removed=draft.history.splice(start)
      bytes-=removed.reduce((total,step)=>total+editStepBytes(step),0);trimmed=true
    }
  }
  draft.historyBytes=Math.max(0,bytes)
  if(trimmed){draft.historyTrimmed=true;draft.historyEpoch=(draft.historyEpoch||0)+1}
  return trimmed
}
export function applyEntry(state: Workspace, entry: JournalEntry) {
  if (entry.revision <= state.revision) return
  if (entry.revision !== state.revision + 1) throw new Error('恢复日志不连续，已保留原始记录')
  for (const op of entry.mutations) {
    if (op.kind === 'edit') {
      const draft = state.drafts[op.draftId]!
      if(draft.historyDroppedGroup&&draft.historyDroppedGroup===op.group){
        draft.text=editText(draft.text,op.forward)
        draft.base=draft.text.startsWith('\ufeff')?draft.text.slice(1):draft.text
        continue
      }
      if(draft.historyDroppedGroup)draft.historyDroppedGroup=undefined
      if(draft.historyBytes===undefined)draft.historyBytes=historyPayloadBytes(draft.history)
      const removed=draft.history.splice(draft.cursor)
      draft.historyBytes-=removed.reduce((total,step)=>total+editStepBytes(step),0)
      const step=ownedHistoryStep({forward:op.forward,reverse:op.reverse,group:op.group})
      draft.history.push(step)
      draft.historyBytes+=editStepBytes(step)
      draft.cursor++
      draft.text = editText(draft.text, op.forward)
      trimDraftHistory(draft,op.group)
    } else if(op.kind==='seal-history-group'){
      const draft=state.drafts[op.draftId]!
      if(draft.historyDroppedGroup===op.group)draft.historyDroppedGroup=undefined
      if(draft.history.some(step=>step.group===op.group))trimDraftHistory(draft)
    } else if (op.kind === 'history') {
      const draft: Draft = state.drafts[op.draftId]!
      while (draft.cursor > op.cursor) draft.text = editText(draft.text, draft.history[--draft.cursor]!.reverse)
      while (draft.cursor < op.cursor) draft.text = editText(draft.text, draft.history[draft.cursor++]!.forward)
    } else if (op.kind === 'terminal-output') {
      if (!op.checkpointKey || ['__proto__','prototype','constructor'].includes(op.checkpointKey)) throw new Error('无效终端检查点路径')
      const checkpoint = state.terminals[op.checkpointKey]
      if (!checkpoint) throw new Error('终端输出日志缺少屏幕检查点')
      const baseSequence = checkpoint.baseSequence ?? checkpoint.sequence
      if (!Number.isSafeInteger(checkpoint.sequence) || checkpoint.sequence < 0 || !Number.isSafeInteger(baseSequence) || baseSequence < 0 || baseSequence > checkpoint.sequence || !Number.isSafeInteger(op.baseSequence) || op.baseSequence !== baseSequence || !Number.isSafeInteger(op.previousSequence) || op.previousSequence !== checkpoint.sequence || !Number.isSafeInteger(op.sequence) || op.sequence <= checkpoint.sequence || !Number.isSafeInteger(op.scroll) || op.scroll < 0) throw new Error('终端输出日志检查点不连续')
      const existing = checkpoint.outputJournal ?? []
      if (!Array.isArray(existing) || existing.length > TERMINAL_OUTPUT_JOURNAL_EVENTS || !Array.isArray(op.events) || !op.events.length || existing.length + op.events.length > TERMINAL_OUTPUT_JOURNAL_EVENTS) throw new Error('终端本地输出日志超出保护预算')
      const journal=appendTerminalJournal(existing,baseSequence,checkpoint.sequence,op.events,op.sequence)
      checkpoint.baseSequence = baseSequence
      checkpoint.outputJournal = journal
      checkpoint.sequence = op.sequence
      checkpoint.scroll = op.scroll
    } else {
      if (!op.path.length || op.path.some(k => ['__proto__','prototype','constructor'].includes(k))) throw new Error('无效恢复路径')
      let target = state as unknown as Record<string, Json>
      for (const part of op.path.slice(0, -1)) target = target[part] as Record<string, Json>
      const key = op.path.at(-1)!
      if (op.kind === 'delete') delete target[key]
      else {
        const value=copy(op.value)
        // Save-as/workspace copies enter through the same journal boundary as
        // newly created drafts. Their copied history now has this owner too.
        if(op.path.length===2&&op.path[0]==='drafts'&&value&&typeof value==='object'&&!Array.isArray(value)&&Array.isArray(value.history))value.history=(value as unknown as Draft).history.map(ownedHistoryStep) as unknown as Json[]
        target[key]=value
      }
    }
  }
  state.revision = entry.revision
}
export function migrateWorkspace(value: Workspace): Workspace {
  const state = copy(value)
  if (state.schemaVersion === 0) {
    state.browserTabId ||= (state as unknown as { tabId?: string }).tabId ?? ''
    state.schemaVersion = 1
  }
  if (state.schemaVersion !== 1) throw new Error(`无法迁移工作现场版本 ${state.schemaVersion}；原记录保留，可导出`)
  state.closedViews ||= []
  state.uploads ||= {}
  for(const draft of Object.values(state.drafts)){initializeDraftHistory(draft)}
  return state
}
