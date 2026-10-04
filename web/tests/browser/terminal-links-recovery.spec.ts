import {test,expect} from '@playwright/test'

const start='\x1b]8;id=doc;https://example.invalid/docs\x1b\\',end='\x1b]8;;\x1b\\'
const cases=[
  {name:'closed',prefix:start+'docs'+end,suffix:' END'},
  {name:'open',prefix:start+'docs',suffix:' more'+end+' END'},
  {name:'open before first character',prefix:start,suffix:'docs'+end+' END'},
  {name:'wrapped wide characters',prefix:'x'.repeat(98)+start+'中文😀link'+end,suffix:' END'},
  {name:'scrollback',prefix:start+'docs'+end+'\r\nrow'.repeat(45),suffix:' END'},
  {name:'resize and reflow',prefix:'x'.repeat(98)+start+'中文😀link'+end+'\r\nrow'.repeat(30),suffix:' END',resize:true},
  {name:'alternate screen',prefix:'normal\x1b[?1049h'+start+'docs',suffix:' more'+end+' END'},
  {name:'return to normal screen',prefix:start+'normal'+end+'\x1b[?1049h'+start+'alternate',suffix:end+'\x1b[?1049l END'},
  {name:'disallowed URI',prefix:'\x1b]8;;javascript:alert(1)\x1b\\unsafe'+end,suffix:' END',count:0},
]
for(const scenario of cases)test(`OSC link remains discoverable after recovery: ${scenario.name}`,async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  const result=await page.evaluate(async scenario=>{
    const t='/src/services/terminals.ts',r='/src/recovery/service.ts'
    const {TerminalModel}=await import(t) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(r) as typeof import('../../src/recovery/service')
    const run=async(remount:boolean)=>{
      const database='links-'+crypto.randomUUID();let recovery=new RecoveryService('links','device','tab',sessionStorage,database);await recovery.restore()
      const container=document.createElement('div');document.body.append(container)
      const output:string[]=[],activated:string[]=[]
      const create=()=>{const model=new TerminalModel(recovery,'session','view',v=>output.push(v));model.terminal.options.linkHandler={activate:(_e,text)=>activated.push(text)};return model}
      let model=create(),sequence=0
      const send=async(data:string)=>model.receive({sequence:++sequence,kind:'output',data:btoa(String.fromCharCode(...new TextEncoder().encode(data)))})
      try{
        await model.mount(container);await model.restoreComplete()
        await send(scenario.prefix)
        if(scenario.resize)await model.receive({sequence:++sequence,kind:'resize',cols:40,rows:10})
        await model.checkpoint();await recovery.awaitPendingWrites()
        if(remount){model.dispose();recovery=new RecoveryService('links','device','tab',sessionStorage,database);await recovery.restore();model=create();await model.mount(container);await model.restoreComplete()}
        await send(scenario.suffix)
        const provider=(model.terminal as unknown as {_core:{_linkProviderService:{linkProviders:Array<{provideLinks:(y:number,callback:(links:Array<{text:string;range:unknown;activate:(event:MouseEvent,text:string)=>void}>|undefined)=>void)=>void}>}}})._core._linkProviderService.linkProviders[0]!
        const links:Array<{text:string;range:unknown;activate:(event:MouseEvent,text:string)=>void}>=[]
        for(let row=1;row<=model.terminal.buffer.active.length;row++)links.push(...await new Promise<typeof links>(resolve=>provider.provideLinks(row,value=>resolve(value||[]))))
        for(const link of links)link.activate(new MouseEvent('click'),link.text)
        return {links:links.map(({text,range})=>({text,range})),activated,output,screen:model.serialize.serialize(),worker:container.dataset.terminalCheckpoint==='worker'}
      }finally{model.dispose();container.remove();await recovery.clear()}
    }
    return {baseline:await run(false),restored:await run(true)}
  },scenario)
  if(scenario.count===0)expect(result.baseline.links).toHaveLength(0)
  else expect(result.baseline.links.length).toBeGreaterThan(0)
  expect(result.restored).toEqual(result.baseline)
  expect(result.restored.worker).toBe(true)
})

test('real pointer activates the same link through two persistent remounts',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  await page.evaluate(async()=>{
    const t='/src/services/terminals.ts',r='/src/recovery/service.ts'
    const {TerminalModel}=await import(t) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(r) as typeof import('../../src/recovery/service')
    const database='link-pointer-'+crypto.randomUUID();let recovery=new RecoveryService('link-pointer','device','tab',sessionStorage,database);await recovery.restore()
    const container=document.createElement('div');container.id='link-pointer';container.style.cssText='position:fixed;left:20px;top:100px;width:800px;height:400px;z-index:999999';document.body.append(container)
    const activated:string[]=[],output:string[]=[]
    const create=()=>{const m=new TerminalModel(recovery,'session','view',v=>output.push(v));m.terminal.options.linkHandler={activate:(_event,text)=>activated.push(text)};return m}
    let model=create();await model.mount(container);await model.restoreComplete()
    await model.receive({sequence:1,kind:'output',data:btoa('\x1b]8;;https://example.invalid/docs\x1b\\docs\x1b]8;;\x1b\\')})
    const probe={activated,output,restore:async()=>{
      await model.checkpoint();await recovery.awaitPendingWrites();model.dispose()
      recovery=new RecoveryService('link-pointer','device','tab',sessionStorage,database);await recovery.restore();model=create();await model.mount(container);await model.restoreComplete()
    },close:async()=>{model.dispose();container.remove();await recovery.clear()}}
    ;(window as unknown as {linkProbe:typeof probe}).linkProbe=probe
  })
  try{
    for(let cycle=0;cycle<3;cycle++){
      if(cycle)await page.evaluate(()=>(window as unknown as {linkProbe:{restore:()=>Promise<void>}}).linkProbe.restore())
      const screen=page.locator('#link-pointer .xterm-screen')
      await page.mouse.move(1000,700);await screen.hover({position:{x:8,y:8}})
      await expect(screen).toHaveClass(/xterm-cursor-pointer/)
      await screen.click({position:{x:8,y:8}})
      await expect.poll(()=>page.evaluate(()=>(window as unknown as {linkProbe:{activated:string[]}}).linkProbe.activated)).toEqual(Array(cycle+1).fill('https://example.invalid/docs'))
    }
    expect(await page.evaluate(()=>(window as unknown as {linkProbe:{output:string[]}}).linkProbe.output)).toEqual([])
  }finally{await page.evaluate(()=>(window as unknown as {linkProbe:{close:()=>Promise<void>}}).linkProbe.close())}
})
