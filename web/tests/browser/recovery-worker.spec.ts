import {test,expect} from '@playwright/test'

test('immediate journal restores edits before the scheduled background write and an explicit flush forces its boundary',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string)
  const dbName='worker-immediate-'+crypto.randomUUID(),service=new RecoveryService('immediate','device','tab',sessionStorage,dbName)
  await service.restore();await Promise.all([service.enableBackgroundPersistence(),service.enableBackgroundPersistence()])
  const mode=service.persistenceMode
  const nativeTimer=window.setTimeout
  window.setTimeout=((handler:TimerHandler,delay?:number,...args:any[])=>nativeTimer(handler,delay===64?10000:delay,...args)) as typeof window.setTimeout
  try{
   service.commit([{kind:'set',path:['preferences','text'],value:'immediate 中文'}])
   const immediatelyProtected=service.status.protected&&sessionStorage.getItem(`blora:tail:${service.key}`)!.includes('immediate 中文')
   // Simulate a refresh before the deferred database transaction has started.
   const fromTail=new RecoveryService('immediate','device','tab',sessionStorage,dbName);await fromTail.restore(false)
   const beforeTimer=fromTail.state.preferences.text==='immediate 中文'
   service.commit([{kind:'set',path:['preferences','second'],value:'same batch'}])
   await service.awaitPendingWrites()
   sessionStorage.removeItem(`blora:tail:${service.key}`)
   const fromDatabase=new RecoveryService('immediate','device','tab',sessionStorage,dbName);await fromDatabase.restore(false)
   const afterFlush=fromDatabase.state.preferences.text==='immediate 中文'&&fromDatabase.state.preferences.second==='same batch'&&fromDatabase.state.revision===service.state.revision
   await service.clear();return {mode,immediatelyProtected,beforeTimer,afterFlush}
  }finally{window.setTimeout=nativeTimer;service.stopBackgroundPersistence()}
 })
 expect(result).toMatchObject({mode:'worker',immediatelyProtected:true,beforeTimer:true,afterFlush:true})
})

test('unavailable background worker retains the foreground native transaction path',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),nativeWorker=window.Worker
  const service=new RecoveryService('fallback','device','tab',sessionStorage,'worker-unavailable-'+crypto.randomUUID());await service.restore()
  window.Worker=new Proxy(nativeWorker,{construct(){throw new DOMException('worker blocked','SecurityError')}})
  try{await service.enableBackgroundPersistence()}finally{window.Worker=nativeWorker}
  service.commit([{kind:'set',path:['preferences','text'],value:'native fallback'}]);await service.awaitPendingWrites()
  sessionStorage.removeItem(`blora:tail:${service.key}`)
  const restored=new RecoveryService('fallback','device','tab',sessionStorage,service.dbName);await restored.restore(false)
  const result={mode:service.persistenceMode,protectedState:service.status.protected,text:restored.state.preferences.text};await service.clear();return result
 })
 expect(result).toEqual({mode:'foreground',protectedState:true,text:'native fallback'})
})

test('a failed initialized background worker falls back without losing the immediate journal or durable recovery',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),nativeWorker=window.Worker
  const dbName='worker-runtime-fallback-'+crypto.randomUUID(),service=new RecoveryService('runtime','device','tab',sessionStorage,dbName)
  await service.restore()
  let worker:Worker|undefined
  window.Worker=new Proxy(nativeWorker,{construct(target,args){worker=Reflect.construct(target,args,target) as Worker;return worker}})
  try{await service.enableBackgroundPersistence()}finally{window.Worker=nativeWorker}
  const initialMode=service.persistenceMode
  service.commit([{kind:'set',path:['preferences','text'],value:'before worker failure'}]);await service.awaitPendingWrites()
  if(!worker)throw Error('The native recovery worker did not initialize')
  worker.dispatchEvent(new ErrorEvent('error',{message:'injected initialized worker failure'}))
  service.commit([{kind:'set',path:['preferences','text'],value:'after worker failure 中文'}])
  const immediatelyProtected=service.status.protected&&sessionStorage.getItem(`blora:tail:${service.key}`)!.includes('after worker failure 中文')
  const immediate=new RecoveryService('runtime','device','tab',sessionStorage,dbName);await immediate.restore(false)
  const immediateText=immediate.state.preferences.text
  await service.awaitPendingWrites()
  const mode=service.persistenceMode,protectedState=service.status.protected
  sessionStorage.removeItem(`blora:tail:${service.key}`)
  const databaseOnly=new RecoveryService('runtime','device','tab',sessionStorage,dbName);await databaseOnly.restore(false)
  const durableText=databaseOnly.state.preferences.text,revision=databaseOnly.state.revision
  await service.clear()
  return {initialMode,mode,immediatelyProtected,immediateText,protectedState,durableText,revision,expectedRevision:service.state.revision}
 })
 expect(result).toMatchObject({initialMode:'worker',mode:'foreground',immediatelyProtected:true,immediateText:'after worker failure 中文',protectedState:true,durableText:'after worker failure 中文'})
 expect(result.revision).toBe(result.expectedRevision)
})

test('explicitly stopping an in-flight background writer does not republish a disposed account',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{openDB}=await import('/node_modules/.vite/deps/idb.js' as string),nativeWorker=window.Worker
  const dbName='worker-runtime-stop-'+crypto.randomUUID(),service=new RecoveryService('stopped','device','tab',sessionStorage,dbName)
  await service.restore()
  let worker:Worker|undefined
  window.Worker=new Proxy(nativeWorker,{construct(target,args){worker=Reflect.construct(target,args,target) as Worker;return worker}})
  try{await service.enableBackgroundPersistence()}finally{window.Worker=nativeWorker}
  if(!worker)throw Error('The native recovery worker did not initialize')
  const db=await openDB(dbName,1),before=await db.get('pointers',service.key),post=worker.postMessage.bind(worker)
  worker.postMessage=((value:any,...args:any[])=>{if(value.kind==='write'){service.stopBackgroundPersistence();return}post(value,args[0]??[])}) as Worker['postMessage']
  service.commit([{kind:'set',path:['preferences','text'],value:'must not publish after explicit stop'}]);await service.awaitPendingWrites()
  const unchanged=JSON.stringify(await db.get('pointers',service.key))===JSON.stringify(before),mode=service.persistenceMode
  // Account cleanup owns the stopped generation and must not recreate records.
  await service.clearAccount()
  const cleared=await db.get('pointers',service.key)===undefined&&(await db.getAllKeys('snapshots')).length===0
  db.close();return {unchanged,mode,cleared}
 })
 expect(result).toEqual({unchanged:true,mode:'foreground',cleared:true})
})

test('an explicit stop after a worker response prevents a late edit from being republished',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{openDB}=await import('/node_modules/.vite/deps/idb.js' as string),nativeWorker=window.Worker
  const dbName='worker-runtime-response-stop-'+crypto.randomUUID(),service=new RecoveryService('response-stopped','device','tab',sessionStorage,dbName)
  await service.restore()
  let worker:Worker|undefined
  window.Worker=new Proxy(nativeWorker,{construct(target,args){worker=Reflect.construct(target,args,target) as Worker;return worker}})
  try{await service.enableBackgroundPersistence()}finally{window.Worker=nativeWorker}
  if(!worker?.onmessage)throw Error('The native recovery worker did not initialize')
  const response=worker.onmessage.bind(worker);let stopped=false
  worker.onmessage=event=>{
   response(event)
   // Resolve the native transaction, then end this generation before the
   // awaited continuation can drain a final teardown edit through another writer.
   if(!stopped){stopped=true;service.commit([{kind:'set',path:['preferences','text'],value:'late teardown edit'}]);service.stopBackgroundPersistence()}
  }
  service.commit([{kind:'set',path:['preferences','text'],value:'last confirmed worker state'}]);const confirmedRevision=service.state.revision
  await service.awaitPendingWrites()
  const db=await openDB(dbName,1),pointer=await db.get('pointers',service.key),record=await db.get('snapshots',pointer.current)
  const result={stopped,mode:service.persistenceMode,durableText:record.preferences.text,durableRevision:record.revision,confirmedRevision,tailRetained:sessionStorage.getItem(`blora:tail:${service.key}`)!.includes('late teardown edit')}
  await service.clearAccount();db.close();return result
 })
 expect(result).toMatchObject({stopped:true,mode:'foreground',durableText:'last confirmed worker state',tailRetained:true})
 expect(result.durableRevision).toBe(result.confirmedRevision)
})

test('background native transactions protect mixed workspace changes, bounded terminal deltas and abort recovery',async({page})=>{
 // Inject a genuine native transaction abort in the separate worker realm.
 // The original foreground-path abort/clone guards remain unchanged.
 await page.route('**/snapshot.worker.ts?**',async route=>{
  const response=await route.fetch(),source=await response.text()
  const preamble=`const workerNativePut=IDBObjectStore.prototype.put;IDBObjectStore.prototype.put=function(value,key){const request=workerNativePut.call(this,value,key);if(this.transaction.db.name.startsWith('worker-boundary-')&&this.name==='snapshots'&&value?.format==='blora-terminal-delta-v1'&&value.revision===15)this.transaction.abort();return request};\n`
  await route.fulfill({response,body:preamble+source})
 })
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{isTerminalDelta}=await import('/src/recovery/terminal-delta.ts' as string),{openDB}=await import('/node_modules/.vite/deps/idb.js' as string),{reactive}=await import('/node_modules/.vite/deps/vue.js' as string)
  const dbName='worker-boundary-'+crypto.randomUUID(),service=new RecoveryService('worker','device','tab',sessionStorage,dbName)
  await service.restore();await service.enableBackgroundPersistence()
  const mode=service.persistenceMode
  service.state.preferences.cold='unchanged editor body '.repeat(30000)
  service.commit([{kind:'set',path:['terminals','s:v'],value:{sessionId:'s',sequence:0,baseSequence:0,outputJournal:[],cols:80,rows:24,screen:'prompt',scroll:0}}]);await service.awaitPendingWrites()
  const append=(sequence:number)=>service.commit([{kind:'terminal-output',checkpointKey:'s:v',baseSequence:0,previousSequence:sequence-1,sequence,events:[{sequence,data:'YQ=='.repeat(256)}],scroll:sequence}])
  for(let sequence=1;sequence<=12;sequence++){append(sequence);await service.awaitPendingWrites()}
  const db=await openDB(dbName,1),prior=await db.get('pointers',service.key),compactRecord=await db.get('snapshots',prior.current)
  const compact=isTerminalDelta(compactRecord)&&compactRecord.entries.length===12
  sessionStorage.removeItem(`blora:tail:${service.key}`)
  const durable=new RecoveryService('worker','device','tab',sessionStorage,dbName);await durable.restore(false)
  const databaseOnly=durable.state.terminals['s:v']?.sequence===12&&durable.state.preferences.cold===service.state.preferences.cold
  // Revision 15 fails in the worker. Its immediate foreground tail must
  // restore the latest output while the durable head remains revision 14.
  append(13);await service.awaitPendingWrites()
  const beforeAbort=await db.get('pointers',service.key)
  append(14);await service.awaitPendingWrites()
  const aborted=!service.status.protected&&JSON.stringify(await db.get('pointers',service.key))===JSON.stringify(beforeAbort)
  const fromTail=new RecoveryService('worker','device','tab',sessionStorage,dbName);await fromTail.restore(false)
  const immediate=fromTail.state.terminals['s:v']?.sequence===14
  // Retry includes both the already mirrored failed revision and new edits.
  service.state.preferences.nested=reactive({text:'prepared proxy'})
  service.commit([{kind:'set',path:['preferences','late'],value:'concurrent 中文'}]);append(15)
  await service.awaitPendingWrites()
  sessionStorage.removeItem(`blora:tail:${service.key}`)
  const restored=new RecoveryService('worker','device','tab',sessionStorage,dbName);await restored.restore(false)
  const latest=restored.state.preferences.nested?.text==='prepared proxy'&&restored.state.preferences.late==='concurrent 中文'&&restored.state.terminals['s:v']?.sequence===15&&restored.state.revision===service.state.revision
  const records=(await db.getAllKeys('snapshots')).length,protectedState=service.status.protected,finalMode=service.persistenceMode
  db.close();await service.clear()
  return {mode,finalMode,compact,databaseOnly,aborted,immediate,latest,records,protectedState}
 })
 expect(result).toMatchObject({mode:'worker',finalMode:'worker',compact:true,databaseOnly:true,aborted:true,immediate:true,latest:true,protectedState:true})
 expect(result.records).toBeLessThanOrEqual(4)
})
