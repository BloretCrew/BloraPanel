import {test,expect} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'

test('fully covered default terminal keeps parsing, protecting and acknowledging output, then reveals the latest screen',async({page})=>{
 await page.addInitScript(()=>{
  const original=HTMLCanvasElement.prototype.getContext
  HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,...args:any[]){if(args[0]==='webgl2'&&args[1]?.failIfMajorPerformanceCaveat)return null;return original.apply(this,args as any)} as typeof original
 })
 const session={sessionId:'occluded-terminal',state:'RUNNING',resource:{kind:'instance',id:'occluded-instance',nodeId:'occluded-node'},backend:'test-transport',cols:100,rows:28,archive:{maxBytes:16777216},maxSessions:64,maxAttachments:16}
 let output:(sequence:number,text:string)=>void=()=>{},ack=0,connections=0,inputs=0,resume=0
 const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
 await page.route('**/api/v1/**',route=>{
  const path=new URL(route.request().url()).pathname
  return route.fulfill({json:path.endsWith('/session')?{user:{userId:'terminal-occlusion-test',name:'测试',admin:true},csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:path.endsWith('/instances')?{items:[{instanceId:session.resource.id,nodeId:session.resource.nodeId,name:'遮挡终端实例',state:'RUNNING',revision:1,config:{mode:'native'}}]}:path.endsWith('/terminals')?{items:[session]}:{items:[]}})
 })
 await page.routeWebSocket(/\/api\/v1\/terminals\/occluded-terminal\/stream/,ws=>{
  connections++;resume=Number(new URL(ws.url()).searchParams.get('sequence'));let bytes=0
  const send=(type:MessageType,payload?:Uint8Array,sequence=0)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:session.sessionId,type,payload,sequence})))
  output=(sequence,text)=>{const payload=encodeJSON({sequence,kind:'output',data:Buffer.from(text).toString('base64')});bytes+=payload.byteLength;send(MessageType.Data,payload,bytes)}
  ws.onMessage(message=>{const frame=decodeEnvelope(new Uint8Array(message as Buffer));if(frame.type===MessageType.Ack)ack++;if(frame.type===MessageType.Data)inputs++})
  send(MessageType.OpenAck,encodeJSON({sessionId:session.sessionId,writable:false,earliest:1,latest:resume,state:'RUNNING'}));send(MessageType.Resume)
 })
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 await page.getByRole('button',{name:'遮挡终端实例',exact:true}).click();await page.getByRole('button',{name:'控制台',exact:true}).click()
 await page.getByRole('button',{name:'打开终端会话管理',exact:true}).click();await page.getByRole('button',{name:/occluded-terminal · RUNNING/}).click()
 const terminal=page.locator('.terminal-container')
 await expect(page.getByText('已连接 · 只读观察',{exact:true})).toBeVisible()
 await expect(terminal).toHaveAttribute('data-terminal-writable','false')
 await terminal.locator('.xterm-screen').evaluate(element=>{
  const observer=new IntersectionObserver(entries=>{(window as any).__terminalScreenIntersecting=entries.at(-1)!.isIntersecting},{threshold:0});observer.observe(element)
 })
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalScreenIntersecting)).toBe(true)
 output(1,'before-cover\r\n');await expect.poll(()=>ack).toBe(1)
 await page.locator('.dock-item[aria-label="设置"]').click()
 const cover=page.locator('.app-window.focused'),coverId=await cover.getAttribute('data-window-id');await cover.getByRole('button',{name:'最大化或还原窗口'}).click()
 await expect(terminal).toHaveAttribute('data-paint-occluded','true')
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalScreenIntersecting)).toBe(false)
 // Hidden hosts still follow their actual layout region on resize. The paint
 // viewport must not make Fit/lease sizing see a zero-size terminal.
 await terminal.evaluate(async element=>{
  const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop(),id=element.closest<HTMLElement>('.app-window')!.dataset.windowId!,r=desktop.state!.windows[id]!.rect
  desktop.geometry(id,{...r,width:r.width-40,height:r.height-30})
 })
 await expect.poll(()=>terminal.evaluate(element=>{
  const host=element as HTMLElement,region=host.closest<HTMLElement>('.terminal-paint-region')!
  return host.clientWidth===region.clientWidth&&host.clientHeight===region.clientHeight&&host.clientWidth>0&&host.clientHeight>0
 })).toBe(true)
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalScreenIntersecting)).toBe(false)
 const size=await terminal.evaluate(element=>({width:element.clientWidth,height:element.clientHeight}))
 // Cross the incremental journal budget while the actual xterm screen is
 // covered. A parsed/durable ACK, not a view repaint, governs stream credit.
 for(let sequence=2;sequence<=26;sequence++){
  output(sequence,('continued-output '.padEnd(100,'x')+'\r\n').repeat(160)+`hidden-checkpoint-${sequence}\r\n`)
  await expect.poll(()=>ack).toBe(sequence)
 }
 expect(await terminal.evaluate(element=>({width:element.clientWidth,height:element.clientHeight}))).toEqual(size)
 const checkpoint=async()=>page.evaluate(async()=>{
  // Restore from the actual database pointer plus synchronous session tail.
  // The most recent protected entry need not already be in the full snapshot.
  const {useDesktop}=await import('/src/desktop/store.ts' as string),{RecoveryService}=await import('/src/recovery/service.ts' as string)
  const current=useDesktop().recovery,restored=new RecoveryService(current.userId,current.deviceId,current.browserTabId,sessionStorage,'blora-workspaces',current.slot)
  try{
   const state=await restored.restore(false),values=Object.values(state.terminals) as any[]
   // The protected journal stores output bytes directly; unlike the wire
   // event it has no `kind` discriminator. Include the synchronous tail too.
   return restored.status.protected&&values.some(value=>value.sequence===26&&(value.screen.includes('hidden-checkpoint-26')||(value.outputJournal||[]).some((event:any)=>event.sequence===26&&atob(event.data).includes('hidden-checkpoint-26'))))
  }finally{restored.db?.close()}
 })
 await expect.poll(checkpoint).toBe(true)
 await cover.getByRole('button',{name:'最小化窗口'}).click()
 await expect(terminal).not.toHaveAttribute('data-paint-occluded','true')
 await expect.poll(()=>page.evaluate(()=>(window as any).__terminalScreenIntersecting)).toBe(true)
 await expect.poll(()=>terminal.locator('.xterm-screen').evaluate(element=>getComputedStyle(element).display)).not.toBe('none')
 // Check actual rendered rows, not just the independently protected parser.
 await expect(terminal.locator('.xterm-rows')).toContainText('hidden-checkpoint-26')
 await page.reload();await expect(terminal).toHaveAttribute('data-terminal-writable','false')
 await expect.poll(()=>connections).toBe(2);expect(resume).toBe(26);expect(inputs).toBe(0)
 await expect.poll(checkpoint).toBe(true)
 // An application notice moves the terminal while the outer window stays
 // unchanged. Hide the shorter host, then remove the notice: the newly
 // exposed terminal must be released without another drag/focus mutation.
 await terminal.evaluate(element=>{
  const notice=document.createElement('p');notice.dataset.layoutNotice='test';notice.className='notice';notice.style.cssText='height:160px;min-height:160px;margin:0;flex:none';notice.textContent='Controlled layout notice';element.closest('.terminal-paint-region')!.before(notice)
 })
 await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
 await page.evaluate(async id=>{
  const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop(),host=document.querySelector<HTMLElement>('.terminal-container[data-session-id="occluded-terminal"]')!,rect=host.getBoundingClientRect()
  desktop.snap(id!);desktop.geometry(id!,{x:rect.x-40,y:rect.y-8,width:rect.width+80,height:Math.max(250,rect.height+50)});desktop.focus(id!)
 },coverId)
 await expect(terminal).toHaveAttribute('data-paint-occluded','true')
 const frameBefore=await terminal.evaluate(element=>element.closest<HTMLElement>('.app-window')!.style.cssText)
 await terminal.evaluate(element=>element.closest('.terminal-app')!.querySelector('[data-layout-notice]')!.remove())
 await expect(terminal).not.toHaveAttribute('data-paint-occluded','true')
 expect(await terminal.evaluate(element=>element.closest<HTMLElement>('.app-window')!.style.cssText)).toBe(frameBefore)
 await expect(terminal.locator('.xterm-rows')).toContainText('hidden-checkpoint-26')
 expect(inputs).toBe(0);expect(errors).toEqual([])
})
