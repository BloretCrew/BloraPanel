import {SnapshotWriter,type SnapshotBatch} from './snapshot-writer'
type Request={id:number;kind:'init';dbName:string;key:string}|{id:number;kind:'write';batch:SnapshotBatch}
let writer:SnapshotWriter|undefined,pending=Promise.resolve()
self.onmessage=(event:MessageEvent<Request>)=>{
 const request=event.data
 pending=pending.then(async()=>{
  if(request.kind==='init'){
   if(writer)throw Error('恢复写入线程重复初始化')
   writer=new SnapshotWriter(request.dbName,request.key);await writer.initialize()
   self.postMessage({id:request.id})
  }else{
   if(!writer)throw Error('恢复写入线程尚未初始化')
   const revision=await writer.write(request.batch)
   self.postMessage({id:request.id,revision})
  }
 }).catch(error=>self.postMessage({id:request.id,error:error instanceof Error?error.message:'恢复写入失败'}))
}
