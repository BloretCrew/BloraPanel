import {openDB,type IDBPDatabase} from 'idb'
import type {Workspace} from '../app-host/types'
import {applyEntry,type JournalEntry} from './state'
import {terminalDelta,type TerminalDelta} from './terminal-delta'

export interface SnapshotBatch {revision:number;state?:Workspace;entries:JournalEntry[]}
// Owns structured recovery data only. Never renders or receives user input.
// One awaited batch bounds its queue; all publications retain the native
// current/previous/base transaction used by the foreground writer.
export class SnapshotWriter {
 private db?:IDBPDatabase
 private state?:Workspace
 private durable?:{current:string;base:string;baseRevision:number;revision:number;entries:JournalEntry[]}
 constructor(readonly dbName:string,readonly key:string){}
 async initialize(){this.db=await openDB(this.dbName,1,{upgrade(db){db.createObjectStore('snapshots');db.createObjectStore('pointers')}})}
 async write(batch:SnapshotBatch){
  if(!this.db)throw Error('恢复数据库尚未初始化')
  if(batch.state){
   if(batch.state.schemaVersion!==1||batch.state.revision!==batch.revision)throw Error('恢复快照边界无效')
   this.state=batch.state;this.durable=undefined
  }else{
   if(!this.state)throw Error('恢复快照缺少完整基线')
   for(const entry of batch.entries)applyEntry(this.state,entry)
  }
  if(this.state!.revision!==batch.revision)throw Error('恢复快照序号不连续')
  const revision=batch.revision,current=`${this.key}:${revision}`,previous=this.durable
  const tx=this.db.transaction(['snapshots','pointers'],'readwrite')
  void tx.done.catch(()=>{})
  try{
   const old=await tx.objectStore('pointers').get(this.key)
   const incoming=previous?batch.entries.filter(entry=>entry.revision>previous.revision&&entry.revision<=revision):[]
   const delta:TerminalDelta|undefined=previous&&old?.current===previous.current&&incoming.length?terminalDelta(previous.base,previous.baseRevision,[...previous.entries,...incoming],revision):undefined
   await tx.objectStore('snapshots').put(delta||this.state,current)
   const prior=current===old?.current?old?.previous:old?.current,priorBase=current===old?.current?old?.previousBase:old?.base
   await tx.objectStore('pointers').put({current,previous:prior,...(delta?{base:delta.base}:{}),...(priorBase?{previousBase:priorBase}:{})},this.key)
   const retained=new Set([current,prior,delta?.base,priorBase])
   for(const stale of new Set([old?.current,old?.previous,old?.base,old?.previousBase]))if(stale&&!retained.has(stale))await tx.objectStore('snapshots').delete(stale)
   await tx.done
   this.durable={current,base:delta?.base||current,baseRevision:delta?.baseRevision??revision,revision,entries:delta?.entries||[]}
   return revision
  }catch(error){try{tx.abort()}catch{/* Transaction already settled. */}await tx.done.catch(()=>{});throw error}
 }
 close(){this.db?.close();this.db=undefined;this.state=undefined;this.durable=undefined}
}
