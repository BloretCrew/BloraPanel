import {test,expect} from '@playwright/test'

for(const encoding of [1006,1016])test(`real pointer retains mouse encoding ${encoding} after persistent restore`,async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  await page.evaluate(async encoding=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const database='mouse-'+crypto.randomUUID()
    let recovery=new RecoveryService('mouse','device','tab',sessionStorage,database)
    await recovery.restore()
    const container=document.createElement('div');container.id='mouse-terminal';container.style.cssText='position:fixed;z-index:999999;left:20px;top:100px;width:800px;height:400px';document.body.append(container)
    const output:string[]=[]
    const create=()=>{const value=new TerminalModel(recovery,'session','view',text=>output.push(text));value.setLease(true);return value}
    let model=create()
    await model.mount(container);await model.restoreComplete()
    await model.receive({sequence:1,kind:'output',data:btoa(`\x1b[?1000h\x1b[?${encoding}h`)})
    await model.checkpoint();await recovery.awaitPendingWrites()
    const state={output,restore:async()=>{
      model.dispose();recovery=new RecoveryService('mouse','device','tab',sessionStorage,database);await recovery.restore();model=create();await model.mount(container);await model.restoreComplete()
    },close:async()=>{model.dispose();container.remove();await recovery.clear()}}
    ;(window as unknown as {mouseProbe:typeof state}).mouseProbe=state
  },encoding)
  try{
    const screen=page.locator('#mouse-terminal .xterm-screen')
    await screen.click({position:{x:40,y:35}})
    const before=await page.evaluate(()=>{const s=(window as unknown as {mouseProbe:{output:string[]}}).mouseProbe;return s.output.splice(0)})
    expect(before).toHaveLength(2)
    expect(before[0]).toMatch(/^\x1b\[<0;\d+;\d+M$/)
    expect(before[1]).toMatch(/^\x1b\[<0;\d+;\d+m$/)
    const replay=await page.evaluate(async()=>{const s=(window as unknown as {mouseProbe:{restore:()=>Promise<void>;output:string[]}}).mouseProbe;await s.restore();return s.output.splice(0)})
    expect(replay).toEqual([])
    await screen.click({position:{x:40,y:35}})
    const after=await page.evaluate(()=>(window as unknown as {mouseProbe:{output:string[]}}).mouseProbe.output)
    expect(after).toEqual(before)
  }finally{await page.evaluate(()=>(window as unknown as {mouseProbe:{close:()=>Promise<void>}}).mouseProbe.close())}
})
