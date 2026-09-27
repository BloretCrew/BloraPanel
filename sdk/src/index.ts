export type ResourceRef = { kind: string; id: string; nodeId?: string }
export type Capability = 'resource.read' | 'resource.write' | 'task.create' | 'window.open' | 'window.move' | 'window.close' | 'shortcut.create' | 'data.read' | 'data.write' | 'notification.publish'
export type Json = null | boolean | number | string | Json[] | { [key:string]: Json }

export interface AppManifest {
  appId: string; packageVersion: string; hostApiVersion: 1; title: string
  entrypoints: string[]; resourceHandlers: string[]; capabilities: Capability[]
  dependencies: Record<string,string>; windowPolicy: 'multiple'
  tabPolicy: { types:string[]; movable:boolean }; stateSchemaVersion:number
  dataSchemaVersion?:number
}
export interface AppState { schemaVersion:number; state:Record<string,Json>; resource?:ResourceRef }
/** Export `migrate` from the frontend module when stateSchemaVersion changes.
 * The sandbox runs it before start; the host commits only a matching snapshot.
 */
export type StateMigration = (previous:AppState,targetVersion:number)=>AppState|Promise<AppState>
export interface ExtensionTask {taskId:string;state:string;phase:string;resource:ResourceRef;result?:Json;error?:string;cancellationRequested?:boolean}
export interface UserData {schemaVersion:number;revision:number;data:Json}
export interface AppHost {
  manifest: AppManifest
  open(input:{entrypoint?:string; resource?:ResourceRef; disposition?:'new-window'|'new-tab'|'dedicated'}): Promise<{windowId:string; viewTabId:string}>
  capture(): Promise<AppState>
  restore(state:AppState): Promise<void>
  readResource(resource?:ResourceRef): Promise<Json>
  updateResource(input:{revision:number;name?:string;group?:string;tags?:string[]},requestId:string,resource?:ResourceRef):Promise<Json>
  createTask(payload:Json,requestId:string):Promise<ExtensionTask>
  readTask(taskId:string):Promise<ExtensionTask>
  cancelTask(taskId:string,requestId:string):Promise<ExtensionTask>
  moveView(targetWindowId?:string):Promise<Json>
  closeView():Promise<Json>
  createShortcut(title?:string):Promise<Json>
  notify(input:{title:string;message?:string;key?:string}):Promise<Json>
  readData():Promise<UserData>
  writeData(data:Json,expectedRevision:number,requestId:string):Promise<UserData>
  capability<T>(name:Capability, fn:()=>Promise<T>): Promise<T>
}

export function validateManifest(m: AppManifest): void {
  if(m.appId.endsWith('.previous'))throw new Error('reserved extension app id suffix')
  if (!/^[-a-z0-9]+\.[-a-z0-9.]+$/.test(m.appId) || m.appId.length > 128 || m.appId.startsWith('blora.') || !/^v?\d+\.\d+\.\d+$/.test(m.packageVersion) || m.packageVersion.length > 64 || m.hostApiVersion !== 1) throw new Error('invalid app manifest')
  if (m.windowPolicy !== 'multiple' || !m.tabPolicy.movable || !Array.isArray(m.tabPolicy.types) || m.tabPolicy.types.some(t=>typeof t!=='string'||t.length>128) || m.entrypoints.length > 64 || m.resourceHandlers.length > 64) throw new Error('unsupported app policy')
  if (!Number.isInteger(m.stateSchemaVersion) || m.stateSchemaVersion < 1) throw new Error('invalid state schema version')
  if(m.dataSchemaVersion!==undefined&&(!Number.isInteger(m.dataSchemaVersion)||m.dataSchemaVersion<1||m.dataSchemaVersion>1_000_000))throw new Error('invalid data schema version')
  if (m.capabilities.some(c=>!['resource.read','resource.write','task.create','window.open','window.move','window.close','shortcut.create','data.read','data.write','notification.publish'].includes(c))) throw new Error('unknown capability')
  if (Object.keys(m.dependencies).length > 64 || Object.entries(m.dependencies).some(([id,version])=>!/^[-a-z0-9]+\.[-a-z0-9.]+$/.test(id)||id.startsWith('blora.')||!/^v?\d+\.\d+\.\d+$/.test(version))) throw new Error('invalid extension dependency')
}

export function createSandboxHost(manifest: AppManifest, transport: (method:string, payload:Json)=>Promise<Json>): AppHost {
  validateManifest(manifest)
  const allowed = new Set(manifest.capabilities)
  const gate = async <T>(name:Capability, fn:()=>Promise<T>) => {
    if (!allowed.has(name)) throw new Error(`capability denied: ${name}`)
    return fn()
  }
  const task=async(method:string,payload:Json):Promise<ExtensionTask>=>{
    const result=await transport(method,payload)
    if(!result||typeof result!=='object'||Array.isArray(result)||!result.task||typeof result.task!=='object'||Array.isArray(result.task)||typeof result.task.taskId!=='string'||typeof result.task.state!=='string')throw new Error('invalid task response')
    return result.task as unknown as ExtensionTask
  }
  return {
    manifest,
    open: input => gate('window.open', async()=>{
      const result = await transport('window.open', input as unknown as Json)
      if (!result || typeof result !== 'object' || !('windowId' in result) || !('viewTabId' in result)) throw new Error('invalid host response')
      return result as {windowId:string;viewTabId:string}
    }),
    capture: ()=>transport('state.capture', null).then(v=>v as unknown as AppState),
    restore: state=>transport('state.restore', state as unknown as Json).then(()=>undefined),
    readResource: resource=>gate('resource.read',()=>transport('resource.read',resource as unknown as Json ?? null)),
    updateResource: (input,requestId,resource)=>gate('resource.write',()=>transport('resource.write',{...input,requestId,...(resource?{resource:resource as unknown as Json}:{})})),
    createTask: (payload,requestId)=>gate('task.create',()=>task('task.create',{payload,requestId})),
    readTask: taskId=>gate('task.create',()=>task('task.read',{taskId})),
    cancelTask: (taskId,requestId)=>gate('task.create',()=>task('task.cancel',{taskId,requestId})),
    moveView: targetWindowId=>gate('window.move',()=>transport('window.move',targetWindowId?{targetWindowId}:{})),
    closeView: ()=>gate('window.close',()=>transport('window.close',null)),
    createShortcut: title=>gate('shortcut.create',()=>transport('shortcut.create',title?{title}:{})),
    notify: input=>gate('notification.publish',()=>transport('notification.publish',input)),
    readData: ()=>gate('data.read',async()=>await transport('data.read',null) as unknown as UserData),
    writeData: (data,expectedRevision,requestId)=>gate('data.write',async()=>await transport('data.write',{data,expectedRevision,requestId,schemaVersion:manifest.dataSchemaVersion||1}) as unknown as UserData),
    capability: gate
  }
}
