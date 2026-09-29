import {test,expect} from '@playwright/test'

const sequences=[{name:'ground',prefix:'',suffix:''},{name:'CSI',prefix:'\x1b[31',suffix:'m'},{name:'OSC',prefix:'\x1b]4;4;#12',suffix:'3456\x07'}]
for(const pending of sequences)for(const later of [false,true])test(`theme reset survives persistent restore at ${pending.name}, later colour ${later}`,async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  const result=await page.evaluate(async({pending,later})=>{
    const t='/src/services/terminals.ts',r='/src/recovery/service.ts'
    const {TerminalModel}=await import(t) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(r) as typeof import('../../src/recovery/service')
    const run=async(remount:boolean)=>{
      const db='theme-'+crypto.randomUUID();let recovery=new RecoveryService('theme','device','tab',sessionStorage,db);await recovery.restore()
      const container=document.createElement('div');document.body.append(container)
      const output:string[]=[]
      const create=()=>{const m=new TerminalModel(recovery,'session','view',v=>output.push(v),'dark');m.setLease(true);return m}
      let model=create(),sequence=0
      const send=async(data:string)=>model.receive({sequence:++sequence,kind:'output',data:btoa(data)})
      try{
        await model.mount(container);await model.restoreComplete()
        await send('\x1b]4;1;#123456\x07');await model.checkpoint()
        await send('\x1b]4;2;#654321\x07'+pending.prefix)
        await model.setColorMode('light');await model.setColorMode('dark');await recovery.awaitPendingWrites()
        if(later){await send(pending.suffix+'\x1b]4;3;#abcdef\x07');await recovery.awaitPendingWrites()}
        if(remount){model.dispose();recovery=new RecoveryService('theme','device','tab',sessionStorage,db);await recovery.restore();model=create();await model.mount(container);await model.restoreComplete()}
        const replay=output.splice(0)
        await send((later?'':pending.suffix)+'X\x1b]4;1;?\x07\x1b]4;2;?\x07\x1b]4;3;?\x07\x1b]4;4;?\x07')
        return {output,replay,screen:model.serialize.serialize()}
      }finally{model.dispose();container.remove();await recovery.clear()}
    }
    return {baseline:await run(false),restored:await run(true)}
  },{pending,later})
  expect(result.restored).toEqual(result.baseline)
  expect(result.restored.replay).toEqual([])
})

test('invalid theme reset position leaves the protected record unchanged',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  const result=await page.evaluate(async()=>{
    const t='/src/services/terminals.ts',r='/src/recovery/service.ts'
    const {TerminalModel}=await import(t) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(r) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('invalid-theme','device','tab',sessionStorage,'invalid-theme-'+crypto.randomUUID());await recovery.restore()
    const container=document.createElement('div');document.body.append(container)
    const output:string[]=[];let model=new TerminalModel(recovery,'session','view',v=>output.push(v))
    try{
      await model.mount(container);await model.restoreComplete()
      recovery.commit([{kind:'set',path:['terminals','session:view','colorResetSequence'],value:1}])
      const before=JSON.stringify(recovery.state.terminals['session:view'])
      model.dispose();model=new TerminalModel(recovery,'session','view',v=>output.push(v))
      let error='';try{await model.mount(container)}catch(value){error=String(value)}
      return {error,unchanged:JSON.stringify(recovery.state.terminals['session:view'])===before,output}
    }finally{model.dispose();container.remove();await recovery.clear()}
  })
  expect(result.error).toContain('主题恢复顺序无效')
  expect(result.unchanged).toBe(true);expect(result.output).toEqual([])
})
