import {createSandboxHost, type AppManifest, type ExtensionTask, type ResourceRef} from '@blora/app-sdk'
declare const REFERENCE_VERSION:string
declare const REFERENCE_SCHEMA:number

export const manifest:AppManifest={appId:'example.reference',packageVersion:typeof REFERENCE_VERSION==='string'?REFERENCE_VERSION:'0.3.0',hostApiVersion:1,title:'参考扩展',entrypoints:['overview'],resourceHandlers:['instance','node'],capabilities:['window.open','window.move','window.close','shortcut.create','data.read','data.write','resource.read','resource.write','notification.publish','task.create'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['overview','resource'],movable:true},stateSchemaVersion:typeof REFERENCE_SCHEMA==='number'?REFERENCE_SCHEMA:1,dataSchemaVersion:typeof REFERENCE_SCHEMA==='number'?REFERENCE_SCHEMA:1}
export function migrate(previous:{state:Record<string,unknown>},targetVersion:number){return {schemaVersion:targetVersion,state:{...previous.state,migratedTo:targetVersion}}}
export async function start(transport:(method:string,payload:any)=>Promise<any>){
  const host=createSandboxHost(manifest,transport)
  const root=document.getElementById('app') || document.body
  const title=document.createElement('h1')
  title.textContent=manifest.title
  const resourceSummary=document.createElement('p')
  resourceSummary.setAttribute('aria-label','当前资源')
  const nodeLabel=document.createElement('label')
  nodeLabel.textContent='节点标识 '
  const node=document.createElement('input')
  node.name='nodeId'
  node.disabled=true
  nodeLabel.append(node)
  const noteLabel=document.createElement('label')
  noteLabel.textContent='工作笔记 '
  const note=document.createElement('textarea')
  note.name='note'
  note.disabled=true
  noteLabel.append(note)
  const open=document.createElement('button')
  open.textContent='打开节点工作窗口'
  open.disabled=true
  const task=document.createElement('button')
  task.textContent='提交节点任务'
  task.disabled=true
  const cancelTask=document.createElement('button')
  cancelTask.textContent='取消任务'
  cancelTask.disabled=true
  const status=document.createElement('output')
  status.setAttribute('aria-live','polite')
  root.replaceChildren(title,resourceSummary,nodeLabel,document.createElement('br'),noteLabel,document.createElement('br'),open,task,cancelTask,status)
  const restored=await host.capture()
  const savedData=document.createElement('p')
  savedData.setAttribute('aria-label','服务器笔记')
  root.append(savedData)
  let dataRevision=0
  try{const data=await host.readData();dataRevision=data.revision;savedData.textContent=JSON.stringify(data)}catch(error){savedData.textContent=String(error)}
  node.value=typeof restored.state.nodeId==='string'?restored.state.nodeId:''
  note.value=typeof restored.state.note==='string'?restored.state.note:''
  let lastTaskId=typeof restored.state.lastTaskId==='string'?restored.state.lastTaskId:''
  let taskPending=restored.state.taskPending as {requestId:string;note:string}|undefined
  const cancelRequests={...((restored.state.cancelRequests||{}) as Record<string,string>)}
  let metadataResource:ResourceRef|undefined,metadataRevision=0
  const instanceName=document.createElement('input');instanceName.setAttribute('aria-label','实例名称')
  instanceName.value=typeof restored.state.instanceName==='string'?restored.state.instanceName:''
  let metadataPending=restored.state.metadataPending as {requestId:string;name:string;revision:number;resource:ResourceRef}|undefined
  try{
    const current=await host.readResource()
    resourceSummary.textContent=current?JSON.stringify(current):'总览窗口：选择节点后打开资源工作窗口'
    if(current && typeof current==='object' && !Array.isArray(current) && current.resource && typeof current.resource==='object' && !Array.isArray(current.resource)) {
      const ref=current.resource
      if(typeof ref.nodeId==='string') node.value=ref.nodeId
      if(ref.kind==='instance'&&typeof ref.id==='string'&&typeof current.configRevision==='number'){
        metadataResource=ref as ResourceRef;metadataRevision=current.configRevision
        if(typeof restored.state.instanceName!=='string')instanceName.value=String(current.name||'')
      }
    }
  }catch(error){resourceSummary.textContent=String(error)}
  let writes=Promise.resolve()
  const persist=()=>{
    const state={schemaVersion:manifest.stateSchemaVersion,state:{nodeId:node.value,note:note.value,lastTaskId,cancelRequests:{...cancelRequests},instanceName:instanceName.value,...(taskPending?{taskPending}:{}) ,...(metadataPending?{metadataPending:metadataPending as any}:{})}}
    writes=writes.catch(()=>undefined).then(()=>host.restore(state))
    void writes.catch(error=>{status.textContent=String(error)})
  }
  node.addEventListener('input',persist)
  note.addEventListener('input',persist)
  if(metadataResource){
    const saveMetadata=document.createElement('button');saveMetadata.textContent='保存实例名称'
    const reloadMetadata=document.createElement('button');reloadMetadata.textContent='重新读取实例'
    const sync=()=>{instanceName.disabled=!!metadataPending;saveMetadata.textContent=metadataPending?'重试保存实例名称':'保存实例名称'}
    sync();instanceName.addEventListener('input',persist)
    saveMetadata.addEventListener('click',async()=>{
      metadataPending??={requestId:crypto.randomUUID(),name:instanceName.value,revision:metadataRevision,resource:metadataResource!}
      sync();saveMetadata.disabled=true;reloadMetadata.disabled=true
      try{
        persist();await writes
        const pending=metadataPending
        const result=await host.updateResource({revision:pending.revision,name:pending.name},pending.requestId,pending.resource) as any
        metadataRevision=result.configRevision;metadataPending=undefined;persist();await writes
        status.textContent='实例名称已保存';resourceSummary.textContent=JSON.stringify(result)
      }catch(error){status.textContent=String(error)}finally{saveMetadata.disabled=false;reloadMetadata.disabled=false;sync()}
    })
    reloadMetadata.addEventListener('click',async()=>{
      try{const current=await host.readResource(metadataResource) as any;metadataRevision=current.configRevision;instanceName.value=current.name;metadataPending=undefined;persist();sync()}catch(error){status.textContent=String(error)}
    })
    root.insertBefore(instanceName,status);root.insertBefore(saveMetadata,status);root.insertBefore(reloadMetadata,status)
  }
  const instanceId=document.createElement('input');instanceId.setAttribute('aria-label','实例标识')
  const openInstance=document.createElement('button');openInstance.textContent='打开实例工作窗口'
  openInstance.addEventListener('click',async()=>{try{await host.open({resource:{kind:'instance',id:instanceId.value.trim(),nodeId:node.value.trim()},disposition:'new-window'})}catch(error){status.textContent=String(error)}})
  root.insertBefore(instanceId,status);root.insertBefore(openInstance,status)
  let poll:ReturnType<typeof setTimeout>|undefined
  let readingTask=false
  const refreshTask=document.createElement('button')
  refreshTask.textContent='刷新任务状态'
  refreshTask.disabled=!lastTaskId
  root.insertBefore(refreshTask,status)
  const terminal=new Set(['SUCCEEDED','FAILED','CANCELLED','INTERRUPTED'])
  const showTask=(current:ExtensionTask)=>{
    cancelTask.disabled=terminal.has(current.state)||!!current.cancellationRequested
    const label:Record<string,string>={SUCCEEDED:'任务已完成',FAILED:'任务失败',CANCELLED:'任务已取消',INTERRUPTED:'任务已中断',RUNNING:'任务执行中',QUEUED:'任务排队中',WAITING_NODE:'等待节点',CANCEL_REQUESTED:'已请求取消'}
    status.textContent=`${label[current.state]||current.state}：${current.error||JSON.stringify(current.result||{})}`
  }
  const follow=async()=>{
    const selected=lastTaskId
    if(!selected||readingTask)return
    if(poll)clearTimeout(poll)
    readingTask=true;refreshTask.disabled=true
    try{
      const current=await host.readTask(selected)
      if(lastTaskId!==selected)return
      showTask(current)
      if(!terminal.has(current.state))poll=setTimeout(follow,500)
    }catch(error){if(lastTaskId===selected){status.textContent=String(error);cancelTask.disabled=true}}
    finally{readingTask=false;refreshTask.disabled=!lastTaskId;if(lastTaskId!==selected)void follow()}
  }
  refreshTask.addEventListener('click',()=>void follow())
  cancelTask.addEventListener('click',async()=>{
    const selected=lastTaskId
    if(!selected)return
    cancelTask.disabled=true
    try{cancelRequests[selected]??=crypto.randomUUID();persist();await writes;const current=await host.cancelTask(selected,cancelRequests[selected]!);if(lastTaskId===selected)showTask(current)}
    catch(error){status.textContent=String(error)}
  })
  if(lastTaskId)void follow()
  const saveData=document.createElement('button')
  saveData.textContent='保存服务器笔记'
  saveData.addEventListener('click',async()=>{
    saveData.disabled=true
    try{const data=await host.writeData({note:note.value},dataRevision,crypto.randomUUID());dataRevision=data.revision;savedData.textContent=JSON.stringify(data)}catch(error){savedData.textContent=String(error)}finally{saveData.disabled=false}
  })
  root.insertBefore(saveData,status)
  const notify=document.createElement('button');notify.textContent='发送站内通知'
  notify.addEventListener('click',async()=>{try{await host.notify({title:'参考扩展笔记',message:note.value,key:'reference-note'});status.textContent='站内通知已保存'}catch(error){status.textContent=String(error)}})
  root.insertBefore(notify,status)
  for(const [label,action] of [
    ['移到新窗口',()=>host.moveView()],
    ['添加当前资源到桌面',()=>host.createShortcut('参考扩展资源')],
    ['关闭当前视图',()=>host.closeView()],
  ] as const){
    const button=document.createElement('button')
    button.textContent=label
    button.addEventListener('click',async()=>{
      try{await writes;await action()}catch(error){status.textContent=String(error)}
    })
    root.insertBefore(button,status)
  }
  open.addEventListener('click',async()=>{
    if(!node.value.trim()){status.textContent='请输入节点标识';return}
    try{
      const result=await host.open({resource:{kind:'node',id:node.value.trim(),nodeId:node.value.trim()},disposition:'new-window'})
      status.textContent=`已打开窗口 ${result.windowId}`
    }catch(error){status.textContent=String(error)}
  })
  task.addEventListener('click',async()=>{
    task.disabled=true
    try{
      taskPending??={requestId:crypto.randomUUID(),note:note.value}
      persist();await writes
      const result=await host.createTask({note:taskPending.note},taskPending.requestId)
      taskPending=undefined
      if(poll)clearTimeout(poll)
      lastTaskId=result.taskId
      persist()
      showTask(result)
      void follow()
    }catch(error){status.textContent=String(error)}
    finally{task.disabled=false}
  })
  node.disabled=false;note.disabled=false;open.disabled=false;task.disabled=false
  return host
}
export function enqueueExampleTask(transport:(method:string,payload:any)=>Promise<any>, payload:Record<string,unknown>,requestId:string){
  const host=createSandboxHost(manifest,transport)
  return host.capability('task.create',()=>transport('task.create',{payload,requestId}))
}
