import 'fake-indexeddb/auto'
import {beforeEach,expect,it,vi} from 'vitest'
const {request}=vi.hoisted(()=>({request:vi.fn()}))
vi.mock('../src/services/api',()=>({api:request,session:{user:{userId:'alice'}}}))
import {session,type Task} from '../src/services/api'
import {RecoveryService} from '../src/recovery/service'
import {restoreUploadFromTask} from '../src/services/uploads'

const task={taskId:'upload-task',actorId:'alice',requestId:'original-request',action:'file.upload',resource:{kind:'instance',id:'instance-a',nodeId:'node-a'},state:'WAITING_CLIENT',phase:'waiting_client_data',revision:2,createdAt:'2026-09-13T00:00:00Z'} as Task
function response(){return {task:{...task},upload:{spec:{id:'node-upload',ownerId:'alice',path:'dir/source.bin',total:131072,hash:'sha256:'+'1'.repeat(64),expectedVersion:'missing',sourceName:'source.bin',sourceModified:1700000000000,sourceFingerprint:'sha256:'+'1'.repeat(64)},offset:65536,stage:'receiving',chunkBytes:65536}}}
async function recovery(){
  const values=new Map<string,string>()
  const storage:Storage={get length(){return values.size},getItem:key=>values.get(key)??null,setItem:(key,value)=>{values.set(key,value)},removeItem:key=>{values.delete(key)},key:index=>[...values.keys()][index]??null,clear:()=>values.clear()}
  const service=new RecoveryService('alice','device','new-browser-tab',storage,'upload-recovery-'+crypto.randomUUID())
  await service.restore();return service
}
beforeEach(()=>{request.mockReset();session.user={userId:'alice'} as typeof session.user})

it('recovers a durable upload into a new workspace without creating a new request',async()=>{
  const service=await recovery();request.mockResolvedValue(response())
  const record=await restoreUploadFromTask(service,task)
  expect(record).toMatchObject({requestId:'original-request',taskId:'upload-task',instanceId:'instance-a',nodeId:'node-a',offset:65536,submitted:true,stage:'waiting_client'})
  expect(request).toHaveBeenCalledWith('/instances/instance-a/files/uploads/upload-task')
  expect((await restoreUploadFromTask(service,task)).uploadId).toBe(record.uploadId)
  expect(Object.keys(service.state.uploads)).toHaveLength(1)
  await service.flush()
})

it('rejects another actor and a late response after an account switch',async()=>{
  const service=await recovery()
  await expect(restoreUploadFromTask(service,{...task,actorId:'bob'})).rejects.toThrow('当前账号')
  expect(request).not.toHaveBeenCalled()
  let release!:(value:ReturnType<typeof response>)=>void
  request.mockReturnValue(new Promise(resolve=>{release=resolve}))
  const pending=restoreUploadFromTask(service,task)
  session.user={userId:'bob'} as typeof session.user;release(response())
  await expect(pending).rejects.toThrow('身份已改变')
  expect(service.state.uploads).toEqual({})
})

it('refuses changed resource, invalid byte checkpoint and completed work',async()=>{
  for(const mutate of [
    (value:ReturnType<typeof response>)=>{value.task={...value.task,resource:{...value.task.resource,id:'other'}}},
    (value:ReturnType<typeof response>)=>{value.task.action='file.save'},
    (value:ReturnType<typeof response>)=>{value.task.requestId='different-request'},
    (value:ReturnType<typeof response>)=>{value.upload.offset=999999},
    (value:ReturnType<typeof response>)=>{value.upload.spec.sourceFingerprint='different'},
    (value:ReturnType<typeof response>)=>{value.task.state='SUCCEEDED'},
  ]){
    const service=await recovery(),value=response();mutate(value);request.mockResolvedValue(value)
    await expect(restoreUploadFromTask(service,task)).rejects.toThrow()
    expect(service.state.uploads).toEqual({})
  }
})
