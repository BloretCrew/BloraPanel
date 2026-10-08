import {test,expect,type Page} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'

// Require settled native paints, as in the window-position oracle: identical
// consecutive full-screen captures within the same eight-attempt bound.
// This prerequisite alone did not fix the paired toolbar AA difference;
// the production toolbar paint ownership is verified separately below.
async function stableNativePaint(page:Page){
 let previous=await page.screenshot({animations:'disabled'})
 for(let attempt=0;attempt<8;attempt++){
  await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  const current=await page.screenshot({animations:'disabled'})
  if(current.equals(previous))return current
  previous=current
 }
 throw Error('Native terminal reference did not settle within eight captures')
}

for(const scale of [1,1.25])test.describe(`native terminal viewport at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('partially covered native terminal retains exact pixels, full grid and protected output through reveal and reload',async({page},info)=>{
  await page.addInitScript(()=>{
   const original=HTMLCanvasElement.prototype.getContext
   HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,...args:any[]){if(['webgl','webgl2','experimental-webgl'].includes(args[0]))return null;return original.apply(this,args as any)} as typeof original
   const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})
  })
  const session={sessionId:'partial-terminal',state:'RUNNING',resource:{kind:'instance',id:'partial-instance',nodeId:'partial-node'},backend:'test-transport',cols:100,rows:28,archive:{maxBytes:16777216},maxSessions:64,maxAttachments:16}
  let output:(sequence:number,text:string)=>void=()=>{},ack=0,connections=0,resume=0,inputs=0
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
   const path=new URL(route.request().url()).pathname
   return route.fulfill({json:path.endsWith('/session')?{user:{userId:'partial-terminal-test',name:'测试',admin:true},csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:path.endsWith('/instances')?{items:[{instanceId:session.resource.id,nodeId:session.resource.nodeId,name:'局部终端实例',state:'RUNNING',revision:1,config:{mode:'native'}}]}:path.endsWith('/terminals')?{items:[session]}:{items:[]}})
  })
  await page.routeWebSocket(/\/api\/v1\/terminals\/partial-terminal\/stream/,ws=>{
   connections++;resume=Number(new URL(ws.url()).searchParams.get('sequence'));let bytes=0
   const send=(type:MessageType,payload?:Uint8Array,sequence=0)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:session.sessionId,type,payload,sequence})))
   output=(sequence,text)=>{const payload=encodeJSON({sequence,kind:'output',data:Buffer.from(text).toString('base64')});bytes+=payload.byteLength;send(MessageType.Data,payload,bytes)}
   ws.onMessage(message=>{const frame=decodeEnvelope(new Uint8Array(message as Buffer));if(frame.type===MessageType.Ack)ack++;if(frame.type===MessageType.Data)inputs++})
   send(MessageType.OpenAck,encodeJSON({sessionId:session.sessionId,writable:false,earliest:1,latest:resume,state:'RUNNING'}));send(MessageType.Resume)
  })
  await page.goto('/');await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'局部终端实例',exact:true}).click();await page.getByRole('button',{name:'控制台',exact:true}).click();await page.getByRole('button',{name:'打开终端会话管理',exact:true}).click();await page.getByRole('button',{name:/partial-terminal · RUNNING/}).click()
  const terminal=page.locator('.terminal-container[data-session-id="partial-terminal"]')
  if(await page.locator('.desktop').getAttribute('data-material-body-paint')==='native')await expect(page.locator('.terminal-app>.editor-toolbar')).toHaveCSS('will-change','transform')
  await expect(page.getByText('已连接 · 只读观察',{exact:true})).toBeVisible();await expect(terminal).toHaveAttribute('data-terminal-renderer','default')
  await terminal.evaluate(async element=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const d=useDesktop();d.geometry(element.closest<HTMLElement>('.app-window')!.dataset.windowId!,{x:70,y:130,width:960,height:700})})
  output(1,'\x1b[?25l\x1b[2J\x1b[H'+Array.from({length:20},(_,i)=>`\x1b[${31+i%7}m${String(i).padStart(2,'0')} 中文 e\u0301 \x1b[3mItalic-overhang\x1b[23m ${'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.repeat(2)}\x1b[0m`).join('\r\n'))
  await expect.poll(()=>ack).toBe(1);await expect(terminal.locator('.xterm-rows')).toContainText('Italic-overhang')
  await page.locator('.dock-item[aria-label="设置"]').click();const upper=page.locator('.app-window.focused'),upperId=await upper.getAttribute('data-window-id')
  const position=async(rect:{x:number;y:number;width:number;height:number})=>page.evaluate(async({id,rect})=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const d=useDesktop();d.geometry(id!,rect);d.focus(id!)},{id:upperId,rect})
  await position({x:350,y:80,width:1000,height:850})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  const geometry=async()=>terminal.evaluate(element=>{
   const host=element as HTMLElement,region=host.closest<HTMLElement>('.terminal-paint-region')!,viewport=host.closest<HTMLElement>('.terminal-paint-viewport')!,a=host.getBoundingClientRect(),b=region.getBoundingClientRect(),p=viewport.getBoundingClientRect()
   return {full:host.clientWidth===region.clientWidth&&host.clientHeight===region.clientHeight&&host.clientWidth>0&&host.clientHeight>0,position:Math.abs(a.x-b.x)<.01&&Math.abs(a.y-b.y)<.01,partial:p.width*p.height<b.width*b.height,cols:host.querySelectorAll('.xterm-rows>div').length}
  })
  const compare=async(label:string,held=false)=>{
   // Fractional displays keep the complete native viewport to preserve its
   // edge AA. Check that explicit fallback as well as every original zero-
   // difference pixel/recovery assertion; do not skip the fractional case.
   if(Number.isInteger(scale))await expect(terminal).toHaveAttribute('data-paint-partial','true')
   else await expect(terminal).not.toHaveAttribute('data-paint-partial','true')
   await expect.poll(geometry).toMatchObject({full:true,position:true,partial:Number.isInteger(scale)})
   if(!held)await page.mouse.move(0,0);await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const partial=await stableNativePaint(page)
   const paintGeometry=()=>terminal.evaluate(element=>[element,element.closest('.terminal-paint-region')!,element.closest('.terminal-paint-viewport')!].map(element=>{const e=element as HTMLElement,r=e.getBoundingClientRect(),s=getComputedStyle(e);return {x:r.x,y:r.y,width:r.width,height:r.height,cssWidth:s.width,cssHeight:s.height,left:s.left,top:s.top,bottom:s.bottom,overflow:s.overflow,boxSizing:s.boxSizing}}))
   const beforeGeometry=process.env.BLORA_E08_NATIVE_VIEWPORT_DIAGNOSTIC==='1'?await paintGeometry():undefined
   // Native full viewport reference, same live nodes, parser and window masks.
   const reference=await page.addStyleTag({content:'.terminal-paint-viewport{left:0!important;top:0!important;width:auto!important;height:auto!important;overflow:hidden!important}.terminal-paint-viewport>.terminal-container{left:0!important;top:0!important;width:auto!important;height:auto!important}'})
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const full=await stableNativePaint(page)
   if(beforeGeometry)console.log('NATIVE_TERMINAL_GEOMETRY',label,JSON.stringify({partial:beforeGeometry,full:await paintGeometry()}))
   await reference.evaluate(element=>(element as HTMLElement).remove())
   await info.attach(`${label}-partial.png`,{body:partial,contentType:'image/png'});await info.attach(`${label}-full.png`,{body:full,contentType:'image/png'})
   const difference=await page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const x=c.getContext('2d')!;x.drawImage(image,0,0);return x.getImageData(0,0,c.width,c.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0;for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta)}return {changed,max}
   },{a:partial.toString('base64'),b:full.toString('base64')})
   console.log('NATIVE_TERMINAL_PARTIAL',label,JSON.stringify(difference));expect(difference).toEqual({changed:0,max:0})
  }
  await compare('right-cover')
  // The internal owned viewport must not invalidate its own held envelope
  // through ResizeObserver. Exposed native rows stay retained until release.
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(title.x+240,title.y+15)
  await expect(upper).toHaveClass(/moving/);await expect.poll(geometry).toMatchObject({full:true,position:true,partial:Number.isInteger(scale)})
  await compare('held-exposure',true)
  // Cross the reserved edge, then reverse while held. Compare actual native
  // pixels as well as dimensions: local/world-coordinate errors can retain a
  // wide viewport yet silently crop the newly exposed terminal columns.
  await page.mouse.move(title.x+320,title.y+15);await compare('held-reserve-crossing',true)
  await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  const exposed=await terminal.evaluate(element=>element.closest('.terminal-paint-viewport')!.getBoundingClientRect().width)
  await page.mouse.move(title.x+210,title.y+15);await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  expect(await terminal.evaluate(element=>element.closest('.terminal-paint-viewport')!.getBoundingClientRect().width)).toBeGreaterThanOrEqual(exposed)
  await compare('held-reverse',true)
  await page.mouse.up()
  await position({x:-650,y:80,width:1200,height:850});await compare('left-cover')
  await position({x:40,y:470,width:1100,height:450});await compare('bottom-cover')
  output(2,'\x1b[H\x1b[35mprotected-partial-update\x1b[0m');await expect.poll(()=>ack).toBe(2)
  const protectedOutput=async()=>page.evaluate(async()=>{
   const {useDesktop}=await import('/src/desktop/store.ts' as string),{RecoveryService}=await import('/src/recovery/service.ts' as string),current=useDesktop().recovery
   const restored=new RecoveryService(current.userId,current.deviceId,current.browserTabId,sessionStorage,'blora-workspaces',current.slot)
   try{const state=await restored.restore(false);return restored.status.protected&&Object.values(state.terminals).some((v:any)=>v.sequence===2&&(v.screen.includes('protected-partial-update')||(v.outputJournal||[]).some((event:any)=>event.sequence===2&&atob(event.data).includes('protected-partial-update'))))}finally{restored.db?.close()}
  })
  await expect.poll(protectedOutput).toBe(true)
  await upper.getByRole('button',{name:'最小化窗口'}).click();await expect(terminal).not.toHaveAttribute('data-paint-partial','true');await expect.poll(geometry).toMatchObject({full:true,position:true,partial:false});await expect(terminal.locator('.xterm-rows')).toContainText('protected-partial-update')
  await page.reload();await expect(terminal).toHaveAttribute('data-terminal-writable','false');await expect.poll(()=>connections).toBe(2);expect(resume).toBe(2);await expect.poll(protectedOutput).toBe(true);await expect(terminal.locator('.xterm-rows')).toContainText('protected-partial-update')
  expect(inputs).toBe(0);expect(errors).toEqual([])
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
