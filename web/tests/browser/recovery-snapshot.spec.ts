import {test,expect} from '@playwright/test'

test('native durable terminal deltas restore cold state, latest journal and aborted transaction without publishing a partial head',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{isTerminalDelta}=await import('/src/recovery/terminal-delta.ts' as string),{openDB}=await import('/node_modules/.vite/deps/idb.js' as string)
  const dbName='native-terminal-delta-'+crypto.randomUUID(),service=new RecoveryService('delta','device','tab',sessionStorage,dbName)
  await service.restore()
  service.state.preferences.cold='cold editor body '.repeat(30000)
  service.commit([{kind:'set',path:['terminals','s:v'],value:{sessionId:'s',sequence:0,baseSequence:0,outputJournal:[],cols:80,rows:24,screen:'prompt',scroll:0}}]);await service.awaitPendingWrites()
  const append=(sequence:number)=>service.commit([{kind:'terminal-output',checkpointKey:'s:v',baseSequence:0,previousSequence:sequence-1,sequence,events:[{sequence,data:'A'.repeat(1024)}],scroll:sequence}])
  for(let sequence=1;sequence<=12;sequence++){append(sequence);await service.awaitPendingWrites()}
  const db=await openDB(dbName,1),before=await db.get('pointers',service.key),record=await db.get('snapshots',before.current)
  const compact=isTerminalDelta(record)&&record.entries.length===12&&JSON.stringify(record).length<String(service.state.preferences.cold).length/4
  sessionStorage.removeItem(`blora:tail:${service.key}`)
  const durable=new RecoveryService('delta','device','tab',sessionStorage,dbName);await durable.restore(false)
  const databaseOnly=durable.state.terminals['s:v']?.sequence===12&&durable.state.preferences.cold===service.state.preferences.cold
  append(13)
  const immediate=new RecoveryService('delta','device','tab',sessionStorage,dbName);await immediate.restore(false);await service.awaitPendingWrites()
  const latest=immediate.state.terminals['s:v']?.sequence===13
  const prior=await db.get('pointers',service.key),nativePut=IDBObjectStore.prototype.put
  IDBObjectStore.prototype.put=function(value:unknown,key?:IDBValidKey){const request=nativePut.call(this,value,key);if(this.transaction.db.name===dbName&&this.name==='snapshots'&&isTerminalDelta(value))this.transaction.abort();return request}
  try{append(14);await service.awaitPendingWrites()}finally{IDBObjectStore.prototype.put=nativePut}
  const aborted=!service.status.protected&&JSON.stringify(await db.get('pointers',service.key))===JSON.stringify(prior)
  const recovered=new RecoveryService('delta','device','tab',sessionStorage,dbName);await recovered.restore(false)
  const journalAfterAbort=recovered.status.protected&&recovered.state.terminals['s:v']?.sequence===14
  const keys=await db.getAllKeys('snapshots');db.close();await service.clear()
  return {compact,databaseOnly,latest,aborted,journalAfterAbort,records:keys.length}
 })
 expect(result).toMatchObject({compact:true,databaseOnly:true,latest:true,aborted:true,journalAfterAbort:true});expect(result.records).toBeLessThanOrEqual(4)
})

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
