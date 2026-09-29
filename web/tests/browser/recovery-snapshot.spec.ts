import {test,expect} from '@playwright/test'

test('native IndexedDB snapshots isolate concurrent edits and retain proxy fallback',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    const module='/src/recovery/service.ts',vueModule='/node_modules/.vite/deps/vue.js',idbModule='/node_modules/.vite/deps/idb.js'
    const {RecoveryService}=await import(module) as typeof import('../../src/recovery/service')
    const {reactive}=await import(vueModule) as typeof import('vue')
    const {openDB}=await import(idbModule) as typeof import('idb')
    const dbName='snapshot-boundary-'+crypto.randomUUID()
    const service=new RecoveryService('snapshot','device','tab',sessionStorage,dbName)
    await service.restore()
    const nativePut=IDBObjectStore.prototype.put;let injected=false
    IDBObjectStore.prototype.put=function(value:unknown,key?:IDBValidKey){
      const request=nativePut.call(this,value,key)
      if(this.name==='snapshots'&&!injected){injected=true;service.commit([{kind:'set',path:['preferences','text'],value:'after 中文'}])}
      return request
    }
    try{
      service.commit([{kind:'set',path:['preferences','text'],value:'before 中文'}])
      await service.awaitPendingWrites()
    }finally{IDBObjectStore.prototype.put=nativePut}
    const db=await openDB(dbName,1),pointer=await db.get('pointers',service.key)
    const current=await db.get('snapshots',pointer.current),previous=await db.get('snapshots',pointer.previous)
    db.close()
    service.state.preferences.nested=reactive({value:'proxy retained'})
    service.commit([{kind:'set',path:['preferences','another'],value:true}])
    await service.awaitPendingWrites()
    sessionStorage.removeItem(`blora:tail:${service.key}`)
    const restored=new RecoveryService('snapshot','device','tab',sessionStorage,dbName)
    await restored.restore()
    return {injected,current:{revision:current.revision,text:current.preferences.text},previous:{revision:previous.revision,text:previous.preferences.text},protected:service.status.protected,restored:restored.state.preferences}
  })
  expect(result.injected).toBe(true)
  expect(result.current.revision).toBe(result.previous.revision+1)
  expect(result.current.text).toBe('after 中文')
  expect(result.previous.text).toBe('before 中文')
  expect(result.protected).toBe(true)
  expect(result.restored).toEqual({text:'after 中文',nested:{value:'proxy retained'},another:true})
})
