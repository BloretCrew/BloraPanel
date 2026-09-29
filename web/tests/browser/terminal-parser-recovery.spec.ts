import {test,expect} from '@playwright/test'

const cases=[
  {name:'CSI colour and resize',prefix:'\x1b[31',suffix:'mR',resize:true},
  {name:'CSI embedded line feed',prefix:'line\x1b[31\n',suffix:'mR'},
  {name:'CSI cursor position',prefix:'first\r\n\x1b[1;',suffix:'1H!'},
  {name:'C1 CSI',prefix:'\u009b31',suffix:'mR'},
  {name:'OSC title',prefix:'\x1b]0;partial',suffix:' title\x07R'},
  {name:'OSC hyperlink',prefix:'\x1b]8;;https://example.invalid/pa',suffix:'th\x1b\\link\x1b]8;;\x1b\\'},
  {name:'DCS mode query',prefix:'\x1bP$q',suffix:'m\x1b\\R'},
  {name:'charset designation',prefix:'\x1b(',suffix:'0qq\x1b(BR'},
  {name:'completed charset',prefix:'\x1b(0',suffix:'qq\x1b(BR',complete:true},
  {name:'completed scroll region',prefix:'1\r\n2\r\n3\r\n4\x1b[2;4r\x1b[4;1H',suffix:'\r\nX',complete:true},
  {name:'saved cursor',prefix:'first\x1b7\r\nsecond',suffix:'\x1b8!',complete:true},
  {name:'custom tab stop',prefix:'\x1b[3g\x1b[1;4H\x1bH\r',suffix:'\tX',complete:true},
  {name:'saved cursor colour',prefix:'\x1b[31mfirst\x1b7\x1b[32m\r\nsecond',suffix:'\x1b8!',complete:true},
  {name:'shifted G1 charset',prefix:'\x1b)0\x0e',suffix:'qq\x0fR',complete:true},
  {name:'alternate buffer saved cursor',prefix:'normal\x1b7\x1b[?1049h\x1b[2;4Halt\x1b7\x1b[4;2H',suffix:'\x1b8!\x1b[?1049l\x1b8?',complete:true},
  {name:'saved cursor after scrollback',prefix:'line\r\n'.repeat(50)+'\x1b7tail',suffix:'\x1b8!',complete:true},
  {name:'origin mode cursor',prefix:'\x1b[2;5r\x1b[?6h\x1b[2;4H',suffix:'X\r\nY',complete:true},
  {name:'SGR mouse protocol',prefix:'\x1b[?1000h\x1b[?1006h',suffix:'\x1b[?1006$p',complete:true},
  {name:'pixel mouse protocol',prefix:'\x1b[?1003h\x1b[?1016h',suffix:'\x1b[?1016$p',complete:true},
  {name:'hidden cursor mode',prefix:'\x1b[?25l',suffix:'\x1b[?25$p',complete:true},
  {name:'blinking cursor mode',prefix:'\x1b[?12h',suffix:'\x1b[?12$p',complete:true},
  {name:'line feed conversion',prefix:'\x1b[20hAB',suffix:'\nX\x1b[20$p',complete:true},
  {name:'cursor style override',prefix:'\x1b[5 q',suffix:'X',complete:true},
]
for(const scenario of cases)test(`${scenario.name} survives checkpoint and remount`,async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async scenario=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts',controlModule='/src/services/terminal-control-state.ts'
    const {captureTerminalControlState}=await import(controlModule) as typeof import('../../src/services/terminal-control-state')
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const run=async(remount:boolean)=>{
      const database='parser-'+crypto.randomUUID()
      let recovery=new RecoveryService('parser','device','tab',sessionStorage,database)
      await recovery.restore()
      recovery.state.views.view={viewTabId:'view',appId:'blora.terminal',type:'session',title:'parser',stateSchemaVersion:1,state:{}}
      const container=document.createElement('div');container.style.cssText='width:800px;height:400px';document.body.append(container)
      const responses:string[]=[],titles:string[]=[]
      const create=()=>{const model=new TerminalModel(recovery,'session','view',value=>responses.push(value));model.terminal.onTitleChange(value=>titles.push(value));model.setLease(true);return model}
      let model=create(),sequence=0
      const output=async(text:string)=>model.receive({sequence:++sequence,kind:'output',data:btoa(String.fromCharCode(...new TextEncoder().encode(text)))})
      try{
        await model.mount(container);await model.restoreComplete()
        await output(scenario.prefix)
        if(scenario.resize)await model.receive({sequence:++sequence,kind:'resize',cols:80,rows:20})
        const compacted=await model.checkpoint()
        await recovery.awaitPendingWrites()
        const protectedSequence=recovery.state.terminals['session:view']?.sequence
        if(remount){model.dispose();recovery=new RecoveryService('parser','device','tab',sessionStorage,database);await recovery.restore();model=create();await model.mount(container);await model.restoreComplete()}
        const replayResponses=responses.splice(0)
        await output(scenario.suffix)
        await model.checkpoint()
        return {screen:model.serialize.serialize({scrollback:3000}),controls:captureTerminalControlState(model.terminal,model.terminal.cols,model.terminal.rows),cols:model.terminal.cols,rows:model.terminal.rows,responses,titles,replayResponses,compacted,protectedSequence}
      }finally{model.dispose();container.remove();await recovery.clear()}
    }
    return {baseline:await run(false),restored:await run(true)}
  },scenario)
  expect(result.restored).toEqual(result.baseline)
  expect(result.restored.compacted).toBe(!!scenario.complete)
  expect(result.restored.protectedSequence).toBe(scenario.resize?2:1)
  expect(result.restored.replayResponses).toEqual([])
})

test('unfinished control sequence over journal budget preserves the last protected boundary',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('budget','device','tab',sessionStorage,'budget-'+crypto.randomUUID())
    await recovery.restore()
    recovery.state.views.view={viewTabId:'view',appId:'blora.terminal',type:'session',title:'budget',stateSchemaVersion:1,state:{}}
    const container=document.createElement('div');document.body.append(container)
    const model=new TerminalModel(recovery,'session','view',()=>{})
    try{
      await model.mount(container);await model.restoreComplete()
      await model.receive({sequence:1,kind:'output',data:btoa('\x1b]0;pending')})
      const before=JSON.stringify(recovery.state.terminals['session:view'])
      let error=''
      try{await model.receive({sequence:2,kind:'output',data:btoa('a'.repeat(60*1024))})}catch(value){error=String(value)}
      return {error,unchanged:JSON.stringify(recovery.state.terminals['session:view'])===before,sequence:recovery.state.terminals['session:view']?.sequence}
    }finally{model.dispose();container.remove();await recovery.clear()}
  })
  expect(result.error).toContain('未完成控制序列超出保护预算')
  expect(result.unchanged).toBe(true)
  expect(result.sequence).toBe(1)
})

test('late Worker snapshot cannot replace output protected while it was pending',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}))
  await page.goto('/')
  const result=await page.evaluate(async()=>{
    const terminalModule='/src/services/terminals.ts',recoveryModule='/src/recovery/service.ts'
    const {TerminalModel}=await import(terminalModule) as typeof import('../../src/services/terminals')
    const {RecoveryService}=await import(recoveryModule) as typeof import('../../src/recovery/service')
    const recovery=new RecoveryService('late','device','tab',sessionStorage,'late-'+crypto.randomUUID())
    await recovery.restore()
    recovery.state.views.view={viewTabId:'view',appId:'blora.terminal',type:'session',title:'late',stateSchemaVersion:1,state:{}}
    const container=document.createElement('div');document.body.append(container)
    let model=new TerminalModel(recovery,'session','view',()=>{})
    let release=()=>{},ready=()=>{}
    const gate=new Promise<void>(resolve=>{release=resolve}),captured=new Promise<void>(resolve=>{ready=resolve})
    try{
      await model.mount(container);await model.restoreComplete()
      const mirror=(model as unknown as {snapshot:{screen:()=>Promise<string>}}).snapshot
      if(!mirror)throw new Error('real checkpoint Worker required')
      const screen=mirror.screen.bind(mirror)
      // Keep the real Worker result, delaying only its return to the caller.
      mirror.screen=async()=>{const result=await screen();ready();await gate;return result}
      const old=model.checkpoint()
      await captured
      await model.receive({sequence:1,kind:'output',data:btoa('new output')})
      release()
      const staleAccepted=await old
      const protectedSequence=recovery.state.terminals['session:view']?.sequence
      model.dispose();model=new TerminalModel(recovery,'session','view',()=>{})
      await model.mount(container);await model.restoreComplete()
      return {staleAccepted,protectedSequence,text:model.terminal.buffer.active.getLine(0)?.translateToString(true)}
    }finally{release();model.dispose();container.remove();await recovery.clear()}
  })
  expect(result).toEqual({staleAccepted:false,protectedSequence:1,text:'new output'})
})
