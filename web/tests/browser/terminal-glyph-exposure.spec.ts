import {test,expect} from '@playwright/test'

// Retain the exposed-pixel/idle-reveal regression that rejected the glyph
// upload/scissor candidates. Keep the complete original native renderer;
// compare exact pixels under equal native masks and on idle full reveal.
for(const scale of [1,1.25])test.describe(`native paint at DPR ${scale}`,()=>{
test.use({...nativeDeviceScale(scale)})
test('native WebGL exposure preserves coloured wide and combined glyphs and idle full reveal',async({page})=>{
 const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
 await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'glyph-exposure-test',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
 await page.goto('/')
 await page.evaluate(async()=>{
  const {Terminal}=await import('/node_modules/.vite/deps/@xterm_xterm.js' as string),{WebglAddon}=await import('/node_modules/.vite/deps/@xterm_addon-webgl.js' as string)
  await import('/src/services/terminals.ts' as string)
  const region=document.createElement('div');region.className='terminal-paint-region';region.style.cssText='position:fixed;left:80px;top:180px;width:720px;height:390px;z-index:10000;background:#111418'
  const viewport=document.createElement('div');viewport.className='terminal-paint-viewport'
  const host=document.createElement('div');host.className='terminal-container';host.style.cssText='width:720px;height:390px';viewport.append(host);region.append(viewport);document.body.append(region)
  const terminal=new Terminal({cols:70,rows:20,fontSize:14,fontFamily:'"Liberation Mono",monospace',allowProposedApi:true,cursorBlink:false,theme:{background:'#111418',foreground:'#e6ebf0'}})
  terminal.open(host);const addon=new WebglAddon(true);terminal.loadAddon(addon);await document.fonts.ready
  const canvas=addon._renderer._canvas as HTMLCanvasElement;canvas.dataset.nativeGlyphCanvas='true'
  ;(window as any).__glyphFixture={host,region,viewport,terminal,addon}
  await new Promise<void>(resolve=>terminal.write('\x1b[2J\x1b[H'+Array.from({length:18},(_,i)=>`\x1b[${31+i%7}m${String(i).padStart(2,'0')} 中文 字形 e\u0301 \x1b[3mItalic-overhang\x1b[23m ${'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.repeat(2)}\x1b[0m`).join('\r\n'),resolve))
  terminal.select(2,3,40);terminal.refresh(0,terminal.rows-1)
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
 })
 const canvas=page.locator('[data-native-glyph-canvas="true"]');await expect(canvas).toBeVisible()
 if(process.env.BLORA_E08_NATIVE_GLYPH_DIAGNOSTIC==='1')await page.evaluate(()=>{
  const f=(window as any).__glyphFixture;f.renderEvents=0
  f.terminal.onRender(()=>f.renderEvents++)
 })
 const native=await canvas.screenshot()
 const compare=async(a:Buffer,b:Buffer,width?:number)=>page.evaluate(async({a,b,width})=>{
  const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const ctx=c.getContext('2d')!;ctx.drawImage(image,0,0);return {pixels:ctx.getImageData(0,0,image.width,image.height).data,width:image.width,height:image.height}}
  const p=await load(a),q=await load(b);if(p.width!==q.width||p.height!==q.height)throw Error('Native canvas dimensions changed')
  let max=0,changed=0
  for(let y=0;y<p.height;y++)for(let x=0;x<Math.min(width??p.width,p.width);x++)for(let channel=0;channel<4;channel++){const i=(y*p.width+x)*4+channel,d=Math.abs(p.pixels[i]!-q.pixels[i]!);max=Math.max(max,d);if(d)changed++}
  return {max,changed}
 },{a:a.toString('base64'),b:b.toString('base64'),width})
 // Compare the same native CSS mask, including its fractional-DPR edge.
 // A partially covered physical pixel cannot equal an unmasked reference.
 await page.evaluate(()=>{(window as any).__glyphFixture.viewport.style.width='100px'})
 const nativeClipped=await canvas.screenshot()
 await page.evaluate(async()=>{
  const f=(window as any).__glyphFixture
  f.terminal.refresh(0,f.terminal.rows-1)
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
 })
 expect(await compare(nativeClipped,await canvas.screenshot())).toEqual({max:0,changed:0})
 await page.evaluate(async()=>{
  const f=(window as any).__glyphFixture;f.viewport.style.removeProperty('width')
  // Exposure alone changes: no output/write/resize or manual redraw is sent.
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
 })
 expect(await compare(native,await canvas.screenshot())).toEqual({max:0,changed:0})
 // Public refresh must respect native atomic synchronized-output mode. An
 // exposure event cannot publish cells that the application has not released.
 await page.evaluate(async()=>{
  const f=(window as any).__glyphFixture
  f.viewport.style.width='100px';f.terminal.refresh(0,f.terminal.rows-1)
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
  await new Promise<void>(resolve=>f.terminal.write('\x1b[?2026h\x1b[H\x1b[35mHELD SYNCHRONIZED UPDATE',resolve))
  f.viewport.style.removeProperty('width');f.terminal.refresh(0,f.terminal.rows-1)
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
 })
 expect(await compare(native,await canvas.screenshot())).toEqual({max:0,changed:0})
 await page.evaluate(async()=>{
  const f=(window as any).__glyphFixture;await new Promise<void>(resolve=>f.terminal.write('\x1b[?2026l',resolve))
  await new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve())))
 })
 const released=await canvas.screenshot()
 if(process.env.BLORA_E08_NATIVE_GLYPH_DIAGNOSTIC==='1')console.log('NATIVE_GLYPH_RELEASE',JSON.stringify(await page.evaluate(()=>{
  const f=(window as any).__glyphFixture,t=f.terminal,service=t._core._renderService,canvas=f.addon._renderer._canvas as HTMLCanvasElement
  const renderer=f.addon._renderer,browser=renderer._coreBrowserService,gl=renderer._gl
  return {renders:f.renderEvents,paused:service._isPaused,needsFullRefresh:service._needsFullRefresh,synchronized:t._core.coreService.decPrivateModes.synchronizedOutput,releasedBuffer:t.buffer.active.getLine(t.buffer.active.baseY).translateToString().startsWith('HELD SYNCHRONIZED UPDATE'),pendingFrame:!!service._renderDebouncer._animationFrame,baseY:t.buffer.active.baseY,viewportY:t.buffer.active.viewportY,dpr:devicePixelRatio,rendererDpr:renderer._devicePixelRatio,browserDpr:browser.dpr,browserWindowDpr:browser.window.devicePixelRatio,sameWindow:browser.window===window,contextLost:gl.isContextLost(),viewport:Array.from(gl.getParameter(gl.VIEWPORT)),canvasWidth:canvas.width,canvasHeight:canvas.height}
 })))
 expect((await compare(native,released)).changed).toBeGreaterThan(0)
 await page.evaluate(()=>{const f=(window as any).__glyphFixture;f.terminal.dispose();f.region.remove()})
 expect(errors).toEqual([])
})
})
import {nativeDeviceScale} from '../../playwright-browser'
