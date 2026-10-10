import { randomUUID } from './uuid'
import {sha256} from '@noble/hashes/sha2.js'
import {watch} from 'vue'
import {bytesToHex} from '@noble/hashes/utils.js'
import HashWorker from './hash.worker?worker'
import {api,session,type Task} from './api'
import {id,type BrowserUpload} from '../app-host/types'
import {json} from '../recovery/state'
import type {RecoveryService} from '../recovery/service'
interface NodeUpload {spec:{id:string;ownerId:string;expectedVersion:string;path:string;total:number;hash:string;sourceName:string;sourceModified:number;sourceFingerprint:string};offset:number;stage:string;version?:string;chunkBytes:number}
interface UploadReply {task:Task;upload?:NodeUpload;waiting?:{code:string;message:string}}
const active=new Map<string,AbortController>()
const polling=new Map<string,ReturnType<typeof setTimeout>>()
const key=(recovery:RecoveryService,uploadId:string)=>`${recovery.key}:${uploadId}`
watch(()=>session.user?.userId,()=>{for(const abort of active.values())abort.abort();for(const timer of polling.values())clearTimeout(timer);polling.clear()})
export function uploadIsActive(recovery:RecoveryService,uploadId:string){return active.has(key(recovery,uploadId))}
function patch(recovery:RecoveryService,uploadId:string,values:Partial<BrowserUpload>){if(session.user?.userId!==recovery.userId)return;const record=recovery.state.uploads[uploadId];if(record)recovery.commit([{kind:'set',path:['uploads',uploadId],value:json({...record,...values})}])}
export function createUpload(recovery:RecoveryService,instanceId:string,nodeId:string|undefined,path:string,file:File,version='missing',resourceKind:'instance'|'node'='instance'){
  if(active.size>=2)throw new Error('同时上传上限为2个，请等待或暂停当前上传')
  if(file.size>1024*1024*1024)throw new Error('当前上传上限为1 GiB')
  const uploadId=id('upload')
  const record:BrowserUpload={uploadId,requestId:randomUUID(),instanceId,resourceKind,nodeId,path,sourceName:file.name,sourceModified:file.lastModified,total:file.size,hash:'',version,offset:0,hashOffset:0,stage:'waiting_client'}
  recovery.commit([{kind:'set',path:['uploads',uploadId],value:json(record)}]);void resumeUpload(recovery,uploadId,file);return uploadId
}
function hashFile(file:File,signal:AbortSignal,progress:(offset:number)=>void){return new Promise<string>((resolve,reject)=>{
  const worker=new HashWorker(),close=()=>{worker.terminate();signal.removeEventListener('abort',abort)},abort=()=>{close();reject(new DOMException('上传已暂停','AbortError'))}
  signal.addEventListener('abort',abort,{once:true});worker.onerror=event=>{close();reject(new Error(event.message))};worker.onmessage=(event:MessageEvent<{offset?:number;hash?:string;error?:string}>)=>{if(event.data.error){close();reject(new Error(event.data.error))}else if(event.data.hash){close();resolve(event.data.hash)}else if(event.data.offset!==undefined)progress(event.data.offset)};worker.postMessage({file})
})}
function verify(record:BrowserUpload,upload:NodeUpload){if(upload.spec.path!==record.path||upload.spec.total!==record.total||upload.spec.hash!==record.hash||upload.spec.sourceName!==record.sourceName||upload.spec.sourceModified!==record.sourceModified||upload.spec.sourceFingerprint!==record.hash||!Number.isSafeInteger(upload.offset)||upload.offset<0||upload.offset>record.total)throw new Error('节点上传检查点与原文件身份不匹配')}
function base(record:BrowserUpload){return `/${record.resourceKind==='node'?'nodes':'instances'}/${encodeURIComponent(record.instanceId)}/files/uploads`}
export async function restoreUploadFromTask(recovery:RecoveryService,task:Task){
  if(task.action!=='file.upload'||!['instance','node'].includes(task.resource.kind)||task.actorId!==recovery.userId||session.user?.userId!==recovery.userId)throw new Error('只能恢复当前账号自己的本机上传')
  const fileScope=task.resource.kind==='node'?'nodes':'instances';const result=await api<UploadReply>(`/${fileScope}/${encodeURIComponent(task.resource.id)}/files/uploads/${encodeURIComponent(task.taskId)}`)
  if(session.user?.userId!==recovery.userId||result.task.actorId!==recovery.userId||result.task.taskId!==task.taskId||result.task.action!=='file.upload'||!result.task.requestId||result.task.requestId!==task.requestId||result.task.resource.kind!==task.resource.kind||result.task.resource.id!==task.resource.id||result.task.resource.nodeId!==task.resource.nodeId)throw new Error('上传账号或资源身份已改变')
  if(result.task.state!=='WAITING_CLIENT'||result.task.cancellationRequested||!result.upload)throw new Error('该任务已不再等待本机文件，请刷新任务状态')
  const spec=result.upload.spec
  if(spec.ownerId!==recovery.userId||!spec.sourceName||spec.sourceName.length>512||!Number.isSafeInteger(spec.total)||spec.total<0||spec.total>1024*1024*1024||!Number.isSafeInteger(spec.sourceModified)||!/^sha256:[a-f0-9]{64}$/.test(spec.hash)||spec.sourceFingerprint!==spec.hash||!spec.path||!spec.expectedVersion)throw new Error('上传检查点缺少可核对的原文件身份')
  const existing=Object.values(recovery.state.uploads).find(record=>record.taskId===task.taskId)
  if(existing){verify(existing,result.upload);return existing}
  const record:BrowserUpload={uploadId:id('upload'),requestId:result.task.requestId,taskId:task.taskId,submitted:true,instanceId:task.resource.id,resourceKind:task.resource.kind as 'instance'|'node',nodeId:task.resource.nodeId,path:spec.path,sourceName:spec.sourceName,sourceModified:spec.sourceModified,total:spec.total,hash:spec.hash,version:spec.expectedVersion,offset:result.upload.offset,hashOffset:spec.total,stage:'waiting_client'}
  verify(record,result.upload)
  recovery.commit([{kind:'set',path:['uploads',record.uploadId],value:json(record)}])
  return record
}
export async function reconcileUpload(recovery:RecoveryService,uploadId:string){
  const record=recovery.state.uploads[uploadId];if(!record||session.user?.userId!==recovery.userId)return
  const pollingKey=key(recovery,uploadId);clearTimeout(polling.get(pollingKey));polling.delete(pollingKey)
  try{
    let taskId=record.taskId
    if(!taskId&&record.submitted){const tasks=await api<{items:Task[]}>('/tasks');taskId=tasks.items.find(task=>task.requestId===record.requestId)?.taskId;if(taskId)patch(recovery,uploadId,{taskId})}
    if(!taskId){if(record.submitted)patch(recovery,uploadId,{error:'上传提交回执尚未确认；重选原文件可核对同一请求'});return}
    const result=await api<UploadReply>(`${base(record)}/${encodeURIComponent(taskId)}`)
    if(result.upload)verify(record,result.upload)
    if(result.task.state==='SUCCEEDED')patch(recovery,uploadId,{stage:'succeeded',offset:record.total,error:''})
    else if(result.task.state==='CANCELLED')patch(recovery,uploadId,{stage:'cancelled',error:''})
    else if(['FAILED','INTERRUPTED'].includes(result.task.state))patch(recovery,uploadId,{stage:'failed',error:result.task.error||result.task.state})
    else{
      patch(recovery,uploadId,{stage:record.stage==='cancel_requested'?'cancel_requested':result.task.state==='WAITING_CLIENT'?'waiting_client':'verifying',offset:result.upload?.offset??record.offset,error:''})
      if(result.task.state!=='WAITING_CLIENT'&&session.user?.userId===recovery.userId)polling.set(pollingKey,setTimeout(()=>void reconcileUpload(recovery,uploadId),1500))
    }
  }catch(e){patch(recovery,uploadId,{stage:record.stage==='cancel_requested'?'cancel_requested':'waiting_client',error:String(e)})}
}
export async function resumeUpload(recovery:RecoveryService,uploadId:string,file:File){
  const record=recovery.state.uploads[uploadId];if(!record||active.has(key(recovery,uploadId)))return
  if(active.size>=2){patch(recovery,uploadId,{stage:'waiting_client',error:'同时上传上限为2个，请稍后续传'});return}
  if(file.name!==record.sourceName||file.size!==record.total||file.lastModified!==record.sourceModified){patch(recovery,uploadId,{error:'重新选择的文件名称、大小或修改时间与原文件不符，未续传'});return}
  const abort=new AbortController();active.set(key(recovery,uploadId),abort);const signal=abort.signal
  try{
    patch(recovery,uploadId,{stage:'hashing',error:''})
    const hash=await hashFile(file,signal,offset=>patch(recovery,uploadId,{hashOffset:offset}))
    if(signal.aborted||session.user?.userId!==recovery.userId)throw new DOMException('上传已暂停','AbortError')
    if(record.hash && record.hash!==hash)throw new Error('重新选择的文件内容指纹与原文件不符，未续传')
    patch(recovery,uploadId,{hash,stage:'uploading'})
    const current=recovery.state.uploads[uploadId]!
    patch(recovery,uploadId,{submitted:true})
    const opened=await api<UploadReply>(base(current),{method:'POST',signal,headers:{'Idempotency-Key':current.requestId},body:JSON.stringify({path:current.path,total:current.total,hash,version:current.version,sourceName:current.sourceName,sourceModified:current.sourceModified,sourceFingerprint:hash})})
    patch(recovery,uploadId,{taskId:opened.task.taskId})
    if(opened.waiting)throw new Error(opened.waiting.message)
    if(opened.task.state==='SUCCEEDED'){patch(recovery,uploadId,{stage:'succeeded',offset:current.total});return}
    if(opened.task.state!=='WAITING_CLIENT'){await reconcileUpload(recovery,uploadId);return}
    if(!opened.upload)throw new Error('节点尚未提供持久化检查点')
    verify(current,opened.upload);let offset=opened.upload.offset
    patch(recovery,uploadId,{offset})
    while(offset<current.total){
      if(signal.aborted||session.user?.userId!==recovery.userId)throw new DOMException('上传已暂停','AbortError')
      const bytes=new Uint8Array(await file.slice(offset,offset+64*1024).arrayBuffer());let binary='';for(const byte of bytes)binary+=String.fromCharCode(byte)
      const result=await api<UploadReply>(`${base(current)}/${encodeURIComponent(opened.task.taskId)}/chunks`,{method:'PUT',signal,headers:{'Idempotency-Key':`${current.requestId}:chunk:${offset}`},body:JSON.stringify({offset,data:btoa(binary),hash:'sha256:'+bytesToHex(sha256(bytes))})})
      if(!result.upload)throw new Error('节点未确认上传分片');verify(current,result.upload)
      if(result.upload.offset<offset+bytes.length)throw new Error('节点上传检查点未推进，未跳过此分片')
      offset=result.upload.offset;patch(recovery,uploadId,{offset})
    }
    await api<UploadReply>(`${base(current)}/${encodeURIComponent(opened.task.taskId)}/complete`,{method:'POST',headers:{'Idempotency-Key':`${current.requestId}:complete`},signal})
    patch(recovery,uploadId,{stage:'verifying'});await reconcileUpload(recovery,uploadId)
  }catch(e){if(!['cancelled','cancel_requested'].includes(recovery.state.uploads[uploadId]?.stage||''))patch(recovery,uploadId,{stage:'waiting_client',error:e instanceof DOMException&&e.name==='AbortError'?'已暂停；保留节点确认的检查点':String(e)})}
  finally{active.delete(key(recovery,uploadId))}
}
export function pauseUpload(recovery:RecoveryService,uploadId:string){active.get(key(recovery,uploadId))?.abort()}
export async function cancelUpload(recovery:RecoveryService,uploadId:string){
  let record=recovery.state.uploads[uploadId];if(!record)return;patch(recovery,uploadId,{stage:'cancel_requested'});pauseUpload(recovery,uploadId)
  if(!record.taskId){if(record.submitted){patch(recovery,uploadId,{error:'原提交结果尚未确认；正在核对任务，尚未取消节点上传'});await reconcileUpload(recovery,uploadId);record=recovery.state.uploads[uploadId]!;if(!record.taskId)return}else{patch(recovery,uploadId,{stage:'cancelled'});return}}
  try{const result=await api<UploadReply>(`${base(record)}/${encodeURIComponent(record.taskId)}`,{method:'DELETE',headers:{'Idempotency-Key':`${record.requestId}:cancel`}});patch(recovery,uploadId,{stage:result.task.state==='CANCELLED'?'cancelled':'cancel_requested',error:result.waiting?.message||''})}catch(e){patch(recovery,uploadId,{stage:'cancel_requested',error:String(e)})}
}
