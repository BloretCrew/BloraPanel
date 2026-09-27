import {watch} from 'vue'
import {sha256} from '@noble/hashes/sha2.js'
import {bytesToHex} from '@noble/hashes/utils.js'
import {api,session,type Task} from './api'
import {dockerPath,type ComposeProject} from './containers'
import {json} from '../recovery/state'
import type {RecoveryService} from '../recovery/service'
import type {ComposeSave} from '../app-host/types'
interface SaveCheckpoint {id:string;projectId:string;expectedRevision:number;total:number;sha256:string;offset:number;state:string;project?:ComposeProject}
interface ProjectChunk {projectId:string;revision:number;offset:number;total:number;sha256:string;data:string}
const active=new Map<string,{userId:string;abort:AbortController}>(),timers=new Map<string,ReturnType<typeof setTimeout>>()
const key=(recovery:RecoveryService,draftId:string)=>`${recovery.key}:${draftId}`
const hash=(data:Uint8Array)=>bytesToHex(sha256(data))
function base64(bytes:Uint8Array){let binary='';for(const byte of bytes)binary+=String.fromCharCode(byte);return btoa(binary)}
function pending(recovery:RecoveryService,draftId:string){return recovery.state.drafts[draftId]?.composeSave}
function update(recovery:RecoveryService,draftId:string,value:ComposeSave){if(session.user?.userId===recovery.userId&&recovery.state.drafts[draftId]?.composeSave?.saveId===value.saveId)recovery.commit([{kind:'set',path:['drafts',draftId,'composeSave'],value:json(value)}])}
function checked(save:SaveCheckpoint,fixed:ComposeSave){if(save.id!==fixed.saveId||save.projectId!==fixed.projectId||save.expectedRevision!==fixed.expectedRevision||save.total!==fixed.total||save.sha256!==fixed.sha256||!Number.isSafeInteger(save.offset)||save.offset<0||save.offset>save.total)throw new Error('Compose 保存身份、校验和或节点检查点不匹配');return save}
function finish(recovery:RecoveryService,draftId:string,fixed:ComposeSave,task:Task){
  if(task.state==='SUCCEEDED'){
    const result=checked(task.result as unknown as SaveCheckpoint,fixed),project=result.project
    if(result.state!=='COMMITTED'||project?.id!==fixed.projectId||project.revision!==fixed.expectedRevision+1)throw new Error('任务未提供已确认的项目版本')
    if(session.user?.userId===recovery.userId&&pending(recovery,draftId)?.saveId===fixed.saveId)recovery.commit([{kind:'set',path:['drafts',draftId,'baseVersion'],value:String(project.revision)},{kind:'set',path:['drafts',draftId,'savedText'],value:fixed.text},{kind:'delete',path:['drafts',draftId,'composeSave']}]);return
  }
  update(recovery,draftId,{...fixed,taskId:task.taskId,state:task.state,error:task.error})
}
function schedule(recovery:RecoveryService,draftId:string){const identity=key(recovery,draftId);clearTimeout(timers.get(identity));timers.set(identity,setTimeout(()=>{timers.delete(identity);void reconcileComposeSave(recovery,draftId)},1800))}
export function composeSaveActive(recovery:RecoveryService,draftId:string){return active.has(key(recovery,draftId))}
export async function reconcileComposeSave(recovery:RecoveryService,draftId:string){
  const fixed=pending(recovery,draftId);if(!fixed||composeSaveActive(recovery,draftId)||session.user?.userId!==recovery.userId)return
  try{
    const task=fixed.taskId?(await api<{task:Task}>(`/tasks/${encodeURIComponent(fixed.taskId)}`)).task:(await api<{items:Task[]}>('/tasks')).items.find(task=>task.requestId===fixed.requestId&&task.action==='compose.save')
    if(task){finish(recovery,draftId,fixed,task);if(!['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state))schedule(recovery,draftId);return}
    const node=checked(await api<SaveCheckpoint>(`${dockerPath(fixed.nodeId)}/project-saves/${encodeURIComponent(fixed.saveId)}`),fixed)
    update(recovery,draftId,{...fixed,offset:node.offset,state:'WAITING_CLIENT',error:node.state==='COMMITTED'?'节点已有提交结果，等待用原请求核对保存任务。':'正文已在本地保护，请明确继续原保存。'})
  }catch(e){update(recovery,draftId,{...fixed,state:'WAITING_CLIENT',error:String(e)})}
}
export function pauseComposeSave(recovery:RecoveryService,draftId:string){const running=active.get(key(recovery,draftId));running?.abort.abort();const fixed=pending(recovery,draftId);if(fixed)update(recovery,draftId,{...fixed,state:'WAITING_CLIENT',error:'本机分片发送已暂停；节点已发生的提交不撤销。'})}
export async function saveCompose(recovery:RecoveryService,draftId:string,nodeId:string,projectId:string){
  const identity=key(recovery,draftId),draft=recovery.state.drafts[draftId];if(!draft||active.has(identity)||session.user?.userId!==recovery.userId)return
  let fixed=draft.composeSave
  if(!fixed||['FAILED','CANCELLED'].includes(fixed.state)){
    const bytes=new TextEncoder().encode(draft.text);if(bytes.length<1||bytes.length>1<<20)throw new Error('Compose 正文需要 1～1048576 个 UTF-8 字节')
    fixed={saveId:crypto.randomUUID(),requestId:crypto.randomUUID(),nodeId,projectId,expectedRevision:Number(draft.baseVersion||0),text:draft.text,total:bytes.length,sha256:hash(bytes),offset:0,state:'PREPARING'}
    recovery.commit([{kind:'set',path:['drafts',draftId,'composeSave'],value:json(fixed)}])
  }
  const abort=new AbortController();active.set(identity,{userId:recovery.userId,abort});clearTimeout(timers.get(identity))
  try{
    const bytes=new TextEncoder().encode(fixed.text);if(bytes.length!==fixed.total||hash(bytes)!==fixed.sha256)throw new Error('本地待保存正文指纹已变化，未续传')
    if(fixed.taskId){active.delete(identity);await reconcileComposeSave(recovery,draftId);return}
    const base=dockerPath(fixed.nodeId),response=checked(await api<SaveCheckpoint>(`${base}/project-saves`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId+':prepare'},body:JSON.stringify({saveId:fixed.saveId,projectId:fixed.projectId,expectedRevision:fixed.expectedRevision,total:fixed.total,sha256:fixed.sha256}),signal:abort.signal}),fixed)
    fixed={...fixed,offset:response.offset,state:'UPLOADING',error:undefined};update(recovery,draftId,fixed)
    while(fixed.offset<bytes.length){
      if(abort.signal.aborted||session.user?.userId!==recovery.userId)throw new Error('本机发送已暂停')
      const offset=fixed.offset,data=bytes.subarray(offset,Math.min(offset+65536,bytes.length)),checkpoint=checked(await api<SaveCheckpoint>(`${base}/project-saves/${encodeURIComponent(fixed.saveId)}/chunks`,{method:'PUT',headers:{'Idempotency-Key':`${fixed.requestId}:chunk:${offset}`},body:JSON.stringify({offset,data:base64(data),sha256:hash(data)}),signal:abort.signal}),fixed)
      if(checkpoint.offset<offset+data.length)throw new Error('节点没有确认已发送的完整分片')
      fixed={...fixed,offset:checkpoint.offset};update(recovery,draftId,fixed)
    }
    if(abort.signal.aborted||session.user?.userId!==recovery.userId)throw new Error('本机发送已暂停')
    fixed={...fixed,state:'COMMITTING'};update(recovery,draftId,fixed)
    const {task}=await api<{task:Task}>(`${base}/project-saves/${encodeURIComponent(fixed.saveId)}/commit`,{method:'POST',headers:{'Idempotency-Key':fixed.requestId},signal:abort.signal})
    finish(recovery,draftId,fixed,task);if(!['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'].includes(task.state))schedule(recovery,draftId)
  }catch(e){update(recovery,draftId,{...fixed,state:'WAITING_CLIENT',error:String(e)})}finally{active.delete(identity)}
}
export async function readComposeSource(nodeId:string,projectId:string,revision:number,signal?:AbortSignal){
  let offset=0,total=-1,digest='';const chunks:Uint8Array[]=[]
  do{const part=await api<ProjectChunk>(`${dockerPath(nodeId)}/projects/${encodeURIComponent(projectId)}/content?revision=${revision}&offset=${offset}&length=65536`,{signal});const bytes=Uint8Array.from(atob(part.data),char=>char.charCodeAt(0));if(part.projectId!==projectId||part.revision!==revision||part.offset!==offset||part.total<1||part.total>1<<20||!bytes.length||bytes.length>65536||offset+bytes.length>part.total||(digest&&digest!==part.sha256)||(total!==-1&&total!==part.total))throw new Error('Compose 版本分片不一致');total=part.total;digest=part.sha256;chunks.push(bytes);offset+=bytes.length}while(offset<total)
  const body=new Uint8Array(total);offset=0;for(const chunk of chunks){body.set(chunk,offset);offset+=chunk.length}if(hash(body)!==digest)throw new Error('Compose 正文整体校验失败');return new TextDecoder('utf-8',{fatal:true}).decode(body)
}
watch(()=>session.user?.userId,userId=>{for(const [identity,item] of active)if(item.userId!==userId){item.abort.abort();active.delete(identity)}for(const [identity,timer] of timers){clearTimeout(timer);timers.delete(identity)}})
