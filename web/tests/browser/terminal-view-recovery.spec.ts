import {test,expect} from '@playwright/test'

test('one grouped wire payload preserves native split UTF-8, resize and durable-only terminal recovery',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{TerminalModel}=await import('/src/services/terminals.ts' as string),{decodeTerminalEvents}=await import('/src/services/terminal-events.ts' as string),{encodeJSON}=await import('/src/services/protocol.ts' as string)
  const dbName='grouped-native-view-'+crypto.randomUUID(),recovery=new RecoveryService('grouped','device','tab',sessionStorage,dbName)
  await recovery.restore();await recovery.enableBackgroundPersistence()
  const first=document.createElement('div'),second=document.createElement('div');first.style.cssText=second.style.cssText='width:700px;height:400px';document.body.append(first,second)
  const input:string[]=[],model=new TerminalModel(recovery,'s','v',(value:string)=>input.push(value));await model.mount(first);await model.restoreComplete()
  const events=[{sequence:1,kind:'output',data:btoa(String.fromCharCode(0xe4))},{sequence:2,kind:'output',data:btoa(String.fromCharCode(0xb8,0xad)+'\r\n\x1b[32m')},{sequence:3,kind:'resize',cols:80,rows:24},{sequence:4,kind:'output',data:btoa('grouped-native-output\x1b[0m')}]
  await model.receiveBatch(decodeTerminalEvents(encodeJSON({kind:'batch',events})))
  await recovery.awaitPendingWrites()
  const before=model.serialize.serialize({scrollback:3000}),mode=recovery.persistenceMode,sequence=model.sequence
  sessionStorage.removeItem(`blora:tail:${recovery.key}`)
  const restored=new RecoveryService('grouped','device','tab',sessionStorage,dbName);await restored.restore(false)
  const remounted=new TerminalModel(restored,'s','v',(value:string)=>input.push(value));await remounted.mount(second)
  const same=before===remounted.serialize.serialize({scrollback:3000}),dimensions=[remounted.terminal.cols,remounted.terminal.rows]
  model.dispose();remounted.dispose();first.remove();second.remove();await recovery.clear()
  return {mode,sequence,restoredSequence:remounted.sequence,same,dimensions,inputCount:input.length,containsUtf8:before.includes('中')}
 })
 expect(result).toEqual({mode:'worker',sequence:4,restoredSequence:4,same:true,dimensions:[80,24],inputCount:0,containsUtf8:true})
})

test('independent native and recovery parsers start together but publish output only after both finish',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{TerminalModel}=await import('/src/services/terminals.ts' as string),{TerminalSnapshot}=await import('/src/services/terminal-snapshot.ts' as string)
  const recovery=new RecoveryService('parallel','device','tab',sessionStorage,'terminal-parallel-'+crypto.randomUUID());await recovery.restore()
  const host=document.createElement('div');host.style.cssText='position:fixed;left:40px;top:90px;width:640px;height:400px';document.body.append(host)
  let inputs=0;const model=new TerminalModel(recovery,'session','view',()=>inputs++)
  await model.mount(host);await model.restoreComplete()
  const originalWrite=model.terminal.write.bind(model.terminal),originalMirror=TerminalSnapshot.prototype.write
  let finishVisible:()=>void=()=>{},finishMirror:()=>void=()=>{},visibleParsed:()=>void=()=>{},mirrorParsed:()=>void=()=>{},mirrorStarted=false,resolved=false
  const visibleDone=new Promise<void>(resolve=>visibleParsed=resolve),mirrorDone=new Promise<void>(resolve=>mirrorParsed=resolve)
  model.terminal.write=(data:string|Uint8Array,callback?:()=>void)=>originalWrite(data,()=>{finishVisible=()=>callback?.();visibleParsed()})
  TerminalSnapshot.prototype.write=function(data:string|Uint8Array){mirrorStarted=true;return originalMirror.call(this,data).then(()=>new Promise<void>(resolve=>{finishMirror=resolve;mirrorParsed()}))}
  const receive=model.receive({sequence:1,kind:'output',data:btoa('first parsed output\r\n')}).then(()=>resolved=true)
  try{
   await visibleDone
   const startedTogether=mirrorStarted,beforeBoth=!resolved&&model.sequence===0&&recovery.state.terminals[model.checkpointKey]!.sequence===0
   finishVisible();await mirrorDone;await Promise.resolve()
   const waitsForMirror=!resolved&&model.sequence===0&&recovery.state.terminals[model.checkpointKey]!.sequence===0
   finishMirror();await receive
   const protectedOutput=recovery.status.protected&&model.sequence===1&&recovery.state.terminals[model.checkpointKey]!.sequence===1
   model.terminal.write=originalWrite;TerminalSnapshot.prototype.write=originalMirror
   // Adjacent API calls remain serial despite concurrent independent parsers.
   const a=model.receive({sequence:2,kind:'output',data:btoa('\x1b[32msecond ')}),b=model.receive({sequence:3,kind:'output',data:btoa('third\x1b[0m\r\n')}),c=model.receive({sequence:4,kind:'resize',cols:90,rows:24})
   await Promise.all([a,b,c]);await model.checkpoint();await recovery.awaitPendingWrites()
   const saved=recovery.state.terminals[model.checkpointKey]!,screen=model.serialize.serialize({scrollback:3000}),ordered=model.sequence===4&&saved.sequence===4&&saved.cols===90&&saved.rows===24&&saved.screen===screen
   model.dispose();host.replaceChildren();sessionStorage.removeItem(`blora:tail:${recovery.key}`)
   const restored=new RecoveryService('parallel','device','tab',sessionStorage,recovery.dbName);await restored.restore(false)
   const next=new TerminalModel(restored,'session','view',()=>inputs++);await next.mount(host)
   const remounted=next.sequence===4&&next.serialize.serialize({scrollback:3000})===screen
   next.dispose();host.remove();await recovery.clear()
   return {startedTogether,beforeBoth,waitsForMirror,protectedOutput,ordered,remounted,inputs}
  }finally{model.terminal.write=originalWrite;TerminalSnapshot.prototype.write=originalMirror;finishVisible();finishMirror();model.dispose();host.remove()}
 })
 expect(result).toEqual({startedTogether:true,beforeBoth:true,waitsForMirror:true,protectedOutput:true,ordered:true,remounted:true,inputs:0})
})

test('duplicate native scroll notifications do not rewrite recovery, while changed scroll and selection survive remount',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
 const result=await page.evaluate(async()=>{
  const {RecoveryService}=await import('/src/recovery/service.ts' as string),{TerminalModel}=await import('/src/services/terminals.ts' as string)
  const dbName='terminal-view-'+crypto.randomUUID(),recovery=new RecoveryService('view','device','tab',sessionStorage,dbName)
  await recovery.restore()
  const host=document.createElement('div');host.style.cssText='position:fixed;left:40px;top:90px;width:640px;height:400px';document.body.append(host)
  let inputs=0
  const model=new TerminalModel(recovery,'session','view',()=>inputs++)
  await model.mount(host);await model.restoreComplete()
  await model.receive({sequence:1,kind:'output',data:btoa(Array.from({length:65},(_,i)=>`history-${i}`).join('\r\n'))})
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())));await recovery.awaitPendingWrites()
  const key=model.checkpointKey,revision=recovery.state.revision
  // The pinned native emitter models the renderer's delayed duplicate scroll
  // callback, without changing parser/buffer/viewport or suppressing events.
  for(let i=0;i<12;i++)model.terminal._core._onScroll.fire({position:model.terminal.buffer.active.viewportY})
  const duplicate=recovery.state.revision===revision
  model.terminal.scrollToLine(0);model.terminal.select(1,2,5)
  const changed=recovery.state.revision>revision&&recovery.state.terminals[key].scroll===0&&!!recovery.state.terminals[key].selection
  const selected=JSON.stringify(recovery.state.terminals[key].selection),screen=model.serialize.serialize({scrollback:3000})
  await recovery.awaitPendingWrites();model.dispose();host.replaceChildren()
  sessionStorage.removeItem(`blora:tail:${recovery.key}`)
  const restored=new RecoveryService('view','device','tab',sessionStorage,dbName);await restored.restore(false)
  const next=new TerminalModel(restored,'session','view',()=>inputs++)
  await next.mount(host)
  const remounted=next.sequence===1&&next.terminal.buffer.active.viewportY===0&&JSON.stringify(next.terminal.getSelectionPosition())===selected&&next.serialize.serialize({scrollback:3000})===screen
  next.dispose();host.remove();await recovery.clear()
  return {duplicate,changed,remounted,inputs}
 })
 expect(result).toEqual({duplicate:true,changed:true,remounted:true,inputs:0})
})
