import {afterEach, describe, expect, it, vi} from 'vitest'
import {createExtensionTransport, fetchExtensionBundle} from '../src/app-host/extension-runtime'
import {session} from '../src/services/api'
import {createSandboxHost, type AppManifest} from '../../sdk/src/index'

describe('extension runtime transport', () => {
  it('preserves the SDK request identity across a lost response and a recreated host',async()=>{
    const manifest:AppManifest={appId:'example.retry',packageVersion:'1.0.0',hostApiVersion:1,title:'Retry',entrypoints:['main'],resourceHandlers:[],capabilities:['task.create'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['main'],movable:true},stateSchemaVersion:1}
    const receipts=new Map<string,string>()
    let loseResponse=true
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async(_url,init)=>{
      const key=new Headers(init?.headers).get('Idempotency-Key')!
      const body=String(init?.body)
      if(receipts.has(key))expect(receipts.get(key)).toBe(body)
      else receipts.set(key,body)
      if(loseResponse){loseResponse=false;throw new TypeError('response lost')}
      return new Response(JSON.stringify({task:{taskId:'accepted-task',state:'RUNNING'}}))
    })
    const host=()=>createSandboxHost(manifest,createExtensionTransport(manifest.appId,'node-a',{},manifest.capabilities))
    await expect(host().createTask({note:'original'},'persisted-request')).rejects.toMatchObject({code:'NETWORK_ERROR',status:0})
    await expect(host().createTask({note:'original'},'persisted-request')).resolves.toMatchObject({taskId:'accepted-task'})
    expect(receipts.size).toBe(1)
    expect(receipts.get('persisted-request')).toBe(JSON.stringify({nodeId:'node-a',payload:{note:'original'}}))
    expect(fetch).toHaveBeenCalledTimes(2)
    loseResponse=true
    await expect(host().cancelTask('accepted-task','persisted-cancel')).rejects.toMatchObject({code:'NETWORK_ERROR',status:0})
    await expect(host().cancelTask('accepted-task','persisted-cancel')).resolves.toMatchObject({taskId:'accepted-task'})
    expect(receipts.size).toBe(2)
    expect(fetch).toHaveBeenCalledTimes(4)
  })
  it('authorizes notifications before the local hook and refuses stale capability',async()=>{
    const notify=vi.fn(async()=>({notificationId:'one'}))
    const fetch=vi.spyOn(globalThis,'fetch').mockResolvedValue(new Response(JSON.stringify({appId:'example.app',title:'Hello',message:'World'})))
    const transport=createExtensionTransport('example.app','',{notify},['notification.publish'])
    await transport('notification.publish',{title:'Hello',message:'World',key:'one',appId:'blora.admin'})
    expect(JSON.parse(String(fetch.mock.calls[0]![1]?.body))).toEqual({title:'Hello',message:'World'})
    expect(notify).toHaveBeenCalledWith({appId:'example.app',title:'Hello',message:'World',key:'one'})
    fetch.mockResolvedValue(new Response('{}',{status:403}))
    await expect(transport('notification.publish',{title:'Denied'})).rejects.toThrow()
    await expect(createExtensionTransport('example.app','',{notify},[])('notification.publish',{title:'Denied'})).rejects.toThrow('未声明')
    expect(notify).toHaveBeenCalledOnce()
  })
  it('binds metadata writes to resource and stable request while denying undeclared writes',async()=>{
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async()=>new Response('{}',{headers:{'Content-Type':'application/json'}}))
    const resource={kind:'instance',id:'instance-a',nodeId:'node-a'}
    const transport=createExtensionTransport('metadata.example','',{resource:()=>resource},['resource.write'])
    await transport('resource.write',{revision:3,name:'renamed',requestId:'stable-meta'})
    expect(fetch.mock.calls[0]![0]).toBe('/api/v1/extensions/metadata.example/resource')
    expect(JSON.parse(String(fetch.mock.calls[0]![1]?.body))).toEqual({revision:3,name:'renamed',resource})
    expect(new Headers(fetch.mock.calls[0]![1]?.headers).get('Idempotency-Key')).toBe('stable-meta')
    await expect(createExtensionTransport('metadata.example','',{},[])('resource.write',{})).rejects.toThrow('未声明')
    expect(fetch).toHaveBeenCalledOnce()
  })
  it('uses a caller supplied stable request ID for scoped user data writes',async()=>{
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async()=>new Response('{"schemaVersion":1,"revision":1,"data":{}}',{headers:{'Content-Type':'application/json'}}))
    const transport=createExtensionTransport('data.example','',{},['data.read','data.write'])
    await transport('data.write',{requestId:'stable-data-request',schemaVersion:1,expectedRevision:0,data:{note:'draft'}})
    expect(fetch.mock.calls[0]![0]).toBe('/api/v1/extensions/data.example/data')
    expect(new Headers(fetch.mock.calls[0]![1]?.headers).get('Idempotency-Key')).toBe('stable-data-request')
    expect(JSON.parse(String(fetch.mock.calls[0]![1]?.body))).not.toHaveProperty('requestId')
    await expect(createExtensionTransport('data.example','',{},[])('data.read',null)).rejects.toThrow('未声明')
    expect(fetch).toHaveBeenCalledOnce()
  })
  afterEach(() => vi.restoreAllMocks())
  it('keeps task reads and cancellation bound to task and application identity',async()=>{
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async()=>new Response(JSON.stringify({task:{taskId:'saved-task',state:'RUNNING'}}),{headers:{'Content-Type':'application/json'}}))
    const transport=createExtensionTransport('example.app','',{ },['task.create'])
    await transport('task.read',{taskId:'saved-task'})
    await transport('task.cancel',{taskId:'saved-task',requestId:'cancel-request'})
    expect(fetch.mock.calls[0]![0]).toBe('/api/v1/extensions/example.app/tasks/saved-task/status')
    expect(fetch.mock.calls[1]![0]).toBe('/api/v1/extensions/example.app/tasks/saved-task/cancel')
    expect(fetch.mock.calls[1]![1]?.method).toBe('POST')
    expect(new Headers(fetch.mock.calls[1]![1]?.headers).get('Idempotency-Key')).toBe('cancel-request')
    await expect(transport('task.cancel',{taskId:'saved-task'})).rejects.toThrow('稳定请求标识')
    await expect(transport('task.cancel',null)).rejects.toThrow('任务标识')
    await expect(createExtensionTransport('example.app','',{},[])('task.read',{taskId:'saved-task'})).rejects.toThrow('未声明')
    expect(fetch).toHaveBeenCalledTimes(2)
  })
  it('reads the current resource through the extension API and rejects undeclared access', async () => {
    let current={kind:'instance',id:'first',nodeId:'node-a'}
    const fetch=vi.spyOn(globalThis,'fetch').mockResolvedValue(new Response('{}',{status:200,headers:{'Content-Type':'application/json'}}))
    const transport=createExtensionTransport('example.app','node-a',{resource:()=>current},['resource.read'])
    await transport('resource.read',null)
    expect(fetch.mock.calls[0]![0]).toBe('/api/v1/extensions/example.app/resource?kind=instance&id=first&nodeId=node-a')
    current={kind:'instance',id:'second',nodeId:'node-b'}
    fetch.mockResolvedValue(new Response('{}',{status:200,headers:{'Content-Type':'application/json'}}))
    await transport('resource.read',null)
    expect(fetch.mock.calls[1]![0]).toBe('/api/v1/extensions/example.app/resource?kind=instance&id=second&nodeId=node-b')
    const denied=createExtensionTransport('example.app','node-a',{resource:()=>current},[])
    await expect(denied('resource.read',null)).rejects.toThrow('未声明能力')
    expect(fetch).toHaveBeenCalledTimes(2)
  })
  it('maps task.create to the scoped API and preserves payload', async () => {
    session.csrfToken = 'test-csrf'
    const fetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({task:{taskId:'t1'}}), {status:202, headers:{'Content-Type':'application/json'}}))
    const transport = createExtensionTransport('example.app', 'node-1')
    await expect(transport('task.create', {requestId:'stable-task',payload:{kind:'demo', value:3}})).resolves.toEqual({task:{taskId:'t1'}})
    expect(fetch).toHaveBeenCalledOnce()
    const [url, init] = fetch.mock.calls[0]!
    expect(url).toBe('/api/v1/extensions/example.app/tasks')
    expect((init as RequestInit).method).toBe('POST')
    expect(JSON.parse(String((init as RequestInit).body))).toEqual({nodeId:'node-1', payload:{kind:'demo', value:3}})
    expect(new Headers((init as RequestInit).headers).get('X-CSRF-Token')).toBe('test-csrf')
    expect(new Headers((init as RequestInit).headers).get('Idempotency-Key')).toBe('stable-task')
    await expect(transport('task.create',{payload:null})).rejects.toThrow('稳定请求标识')
  })
  it('rejects unsupported methods and missing scope before fetch', async () => {
    expect(() => createExtensionTransport('', 'node-1')).toThrow()
    const fetch = vi.spyOn(globalThis, 'fetch')
    const transport = createExtensionTransport('example.app', 'node-1')
    await expect(transport('window.open', null)).rejects.toThrow('宿主未提供能力')
    expect(fetch).not.toHaveBeenCalled()
  })
  it('forwards only explicitly injected view and state hooks', async () => {
    const transport = createExtensionTransport('example.app', 'node-1', {
      open: async payload => ({opened: payload}),
      capture: async () => ({schemaVersion:1}),
      restore: async payload => ({restored: payload}),
    })
    await expect(transport('window.open', {entrypoint:'overview'})).resolves.toEqual({opened:{entrypoint:'overview'}})
    await expect(transport('state.capture', null)).resolves.toEqual({schemaVersion:1})
    await expect(transport('state.restore', {schemaVersion:1})).resolves.toEqual({restored:{schemaVersion:1}})
  })
  it('enforces declared capabilities before invoking host hooks or API', async () => {
    const fetch = vi.spyOn(globalThis, 'fetch')
    const transport = createExtensionTransport('example.app', 'node-1', {open: async () => ({opened:true})}, ['task.create'])
    await expect(transport('window.open', null)).rejects.toThrow('未声明能力')
    expect(fetch).not.toHaveBeenCalled()
  })
  it('allows overview window and state hooks without a node while keeping tasks node-bound', async () => {
    const transport = createExtensionTransport('example.app', '', {
      open: async payload => ({opened: payload}),
      capture: async () => ({schemaVersion: 1}),
      restore: async () => null,
    }, ['window.open','task.create'])
    await expect(transport('window.open', {entrypoint:'overview'})).resolves.toEqual({opened:{entrypoint:'overview'}})
    await expect(transport('state.capture', null)).resolves.toEqual({schemaVersion:1})
    await expect(transport('task.create', null)).rejects.toThrow('缺少节点标识')
  })
  it('fetches only the enabled-extension bundle endpoint', async () => {
    const fetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('export const start=()=>{}', {status:200}))
    await expect(fetchExtensionBundle('example.app')).resolves.toContain('start')
    expect(fetch.mock.calls[0]![0]).toBe('/api/v1/extensions/example.app/bundle')
    await expect(fetchExtensionBundle('')).rejects.toThrow()
  })
})
