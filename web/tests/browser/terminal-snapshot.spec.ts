import {test,expect} from '@playwright/test'

test('terminal queries are silent during local and server replay but answer only live writable output',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('query-boundary','device','tab',sessionStorage,'query-'+crypto.randomUUID())
    await recovery.restore()
    recovery.state.views.view={viewTabId:'view',appId:'blora.terminal',type:'session',title:'queries',stateSchemaVersion:1,state:{}}
    // DECRQM produces the $y response seen in the intermittent real Vim test.
    const query='\x1b[?2004$p',responses:string[]=[]
    const container=document.createElement('div');container.style.cssText='width:800px;height:400px';document.body.append(container)
    let model=new TerminalModel(recovery,'session','view',value=>responses.push(value))
    try{
      await model.mount(container)
      await model.checkpoint()
      model.setLease(true)
      await model.receive({sequence:1,kind:'output',data:btoa(query)})
      const serverReplay=responses.splice(0)
      // Remount from the actual protected raw-output journal containing a query.
      const journalContainsQuery=recovery.state.terminals['session:view']?.outputJournal?.some(event=>atob(event.data).includes(query))
      model.dispose()
      model=new TerminalModel(recovery,'session','view',value=>responses.push(value))
      model.setLease(true)
      await model.mount(container)
      const localReplay=responses.splice(0)
      await model.restoreComplete()
      const completion=responses.splice(0)
      await model.receive({sequence:2,kind:'output',data:btoa(query)})
      const live=responses.splice(0)
      model.setLease(false)
      await model.receive({sequence:3,kind:'output',data:btoa(query)})
      const readOnly=responses.splice(0)
      model.setLease(true)
      await model.receive({sequence:4,kind:'output',data:btoa(query)})
      return {journalContainsQuery,serverReplay,localReplay,completion,live,readOnly,reacquired:responses}
    }finally{model.dispose();container.remove();await recovery.clear()}
  })
  expect(result.journalContainsQuery).toBe(true)
  expect(result.serverReplay).toEqual([])
  expect(result.localReplay).toEqual([])
  expect(result.completion).toEqual([])
  expect(result.live).toEqual(['\x1b[?2004;2$y'])
  expect(result.readOnly).toEqual([])
  expect(result.reacquired).toEqual(result.live)
})

test('worker and visible xterm checkpoints agree across UTF-8 boundaries, scrollback, modes and resize',async({page})=>{
  page.on('console',message=>{if(message.text().startsWith('Terminal checkpoint fallback'))console.log(message.text())})
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    // Import the actual Vite modules; no terminal/parser/Worker test doubles.
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('parity','device','tab',sessionStorage,'parity-'+crypto.randomUUID())
    await recovery.restore()
    recovery.state.views.parity={viewTabId:'parity',appId:'blora.terminal',type:'session',title:'parity',stateSchemaVersion:1,state:{}}
    const container=document.createElement('div');container.style.cssText='width:900px;height:500px';document.body.append(container)
    const model=new TerminalModel(recovery,'session','parity',()=>{})
    let sequence=0
    const checks:{worker:boolean;equal:boolean;stage:string}[]=[]
    const check=async(stage:string)=>{
      await model.checkpoint()
      checks.push({stage,worker:container.dataset.terminalCheckpoint==='worker',equal:recovery.state.terminals['session:parity'].screen===model.serialize.serialize({scrollback:3000})})
    }
    const output=async(data:Uint8Array)=>{
      for(let offset=0;offset<data.length;offset+=32*1024){
        let binary='';for(const byte of data.slice(offset,offset+32*1024))binary+=String.fromCharCode(byte)
        await model.receive({sequence:++sequence,kind:'output',data:btoa(binary)})
      }
    }
    const text=async(data:string)=>output(new TextEncoder().encode(data))
    try{
      await model.mount(container);await model.restoreComplete()
      await text(('history '+ '中文🙂'.repeat(3)+'\r\n').repeat(3100));await check('full scrollback')
      const utf8=new TextEncoder().encode('中文🙂')
      await output(utf8.slice(0,2));await check('partial UTF-8')
      await output(utf8.slice(2));await check('completed UTF-8')
      await text('\u001b[31;44;1m彩色\u001b[0m\u001b[?1049h\u001b[2J\u001b[H全屏编辑器\u001b[?25l\u001b[?2004h');await check('alternate screen and modes')
      await model.receive({sequence:++sequence,kind:'resize',cols:80,rows:24});await check('resize in alternate screen')
      await text('\u001b[?1049l\u001b[?25h\u001b[?2004l\r\n回到正文');await check('restored normal buffer')
      await recovery.awaitPendingWrites()
      return checks
    }finally{model.dispose();container.remove();await recovery.clear()}
  })
  expect(result).toHaveLength(6)
  for(const check of result){expect(check.worker,check.stage).toBe(true);expect(check.equal,check.stage).toBe(true)}
})

test('a killed checkpoint worker falls back without losing a split UTF-8 output or replaying input',async({page})=>{
  test.setTimeout(90000)
  await page.addInitScript(()=>{
    const Original=Worker
    ;(window as any).__checkpointWorkers=[]
    window.Worker=class extends Original{
      constructor(url:string|URL,options?:WorkerOptions){
        super(url,options)
        if(String(url).includes('terminal-snapshot'))(window as any).__checkpointWorkers.push(this)
      }
    }
  })
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('worker-loss','device','tab',sessionStorage,'worker-loss-'+crypto.randomUUID())
    await recovery.restore()
    recovery.state.views.view={viewTabId:'view',appId:'blora.terminal',type:'session',title:'worker-loss',stateSchemaVersion:1,state:{}}
    const container=document.createElement('div');container.style.cssText='width:800px;height:400px';document.body.append(container)
    let inputs=0
    const model=new TerminalModel(recovery,'session','view',()=>inputs++)
    let restored:InstanceType<typeof TerminalModel>|undefined
    try{
      await model.mount(container);await model.restoreComplete()
      const wasWorker=container.dataset.terminalCheckpoint==='worker'
      await model.receive({sequence:1,kind:'output',data:btoa(String.fromCharCode(0xe4,0xb8))})
      // Terminate the real parser process. No error is dispatched, so the
      // production request deadline and fallback must actually run.
      ;((window as any).__checkpointWorkers[0] as Worker).terminate()
      await model.receive({sequence:2,kind:'output',data:btoa(String.fromCharCode(0xad))})
      await model.receive({sequence:3,kind:'output',data:btoa('\r\n\x1b[32mkept\x1b[0m')})
      await model.checkpoint();await recovery.awaitPendingWrites()
      const saved=recovery.state.terminals['session:view']!
      const screen=model.serialize.serialize({scrollback:3000})
      const fallback=container.dataset.terminalCheckpoint==='main'
      const protectedOutput=recovery.status.protected&&saved.sequence===3&&saved.screen===screen
      model.dispose()
      restored=new TerminalModel(recovery,'session','view',()=>inputs++)
      await restored.mount(container);await restored.restoreComplete()
      return {wasWorker,fallback,protectedOutput,sameScreen:restored.serialize.serialize({scrollback:3000})===screen,screen,inputs}
    }finally{restored?.dispose();model.dispose();container.remove();await recovery.clear()}
  })
  expect(result.wasWorker).toBe(true)
  expect(result.fallback).toBe(true)
  expect(result.protectedOutput).toBe(true)
  expect(result.sameScreen).toBe(true)
  expect(result.screen).toContain('中')
  expect(result.screen).toContain('kept')
  expect(result.inputs).toBe(0)
})
