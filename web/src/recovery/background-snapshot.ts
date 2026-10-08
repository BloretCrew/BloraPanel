import SnapshotWorker from './snapshot.worker?worker'
import type {SnapshotBatch} from './snapshot-writer'
import {copy} from './state'
// Worker lifetime/transport failure is distinct from an error returned by an
// actual native storage transaction. Only the former may switch writers.
export class BackgroundSnapshotUnavailableError extends Error {}
export class BackgroundSnapshot {
 private worker=new SnapshotWorker()
 private next=0
 private closed=false
 private pending=new Map<number,{resolve:(revision?:number)=>void;reject:(error:Error)=>void;timer:ReturnType<typeof setTimeout>}>()
 constructor(){
  this.worker.onmessage=(event:MessageEvent<{id:number;revision?:number;error?:string}>)=>{
   const response=this.pending.get(event.data.id);if(!response)return
   this.pending.delete(event.data.id);clearTimeout(response.timer)
   if(event.data.error)response.reject(Error(event.data.error));else response.resolve(event.data.revision)
  }
  this.worker.onerror=()=>this.dispose();this.worker.onmessageerror=()=>this.dispose()
 }
 private request(value:Record<string,unknown>){
  if(this.closed)return Promise.reject(new BackgroundSnapshotUnavailableError('恢复写入线程不可用'))
  return new Promise<number|undefined>((resolve,reject)=>{
   const id=++this.next,timer=setTimeout(()=>this.dispose(),30000)
   this.pending.set(id,{resolve,reject,timer})
   try{this.worker.postMessage({...value,id})}catch(error){
    // App-supplied nested Vue proxies retain the existing JSON sanitization
    // fallback. Retry only synchronous clone errors, never storage failures.
    if(error instanceof DOMException&&error.name==='DataCloneError'){
     try{this.worker.postMessage({...copy(value),id});return}catch{/* Fail below. */}
    }
    this.pending.delete(id);clearTimeout(timer)
    if(error instanceof DOMException&&error.name!=='DataCloneError'){
     this.dispose();reject(new BackgroundSnapshotUnavailableError('恢复写入线程无法接收数据'))
    }else reject(error instanceof Error?error:Error('恢复数据无法传入写入线程'))
   }
  })
 }
 initialize(dbName:string,key:string){return this.request({kind:'init',dbName,key})}
 async write(batch:SnapshotBatch){const revision=await this.request({kind:'write',batch});if(revision!==batch.revision)throw Error('恢复写入线程确认序号不匹配')}
 dispose(){
  if(this.closed)return
  this.closed=true;this.worker.terminate()
  for(const response of this.pending.values()){clearTimeout(response.timer);response.reject(new BackgroundSnapshotUnavailableError('恢复写入线程不可用'))}
  this.pending.clear()
 }
}
