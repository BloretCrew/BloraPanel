import {test,expect} from '@playwright/test'
// Test-only row delegation is not installed by the shipped terminal renderer.

for(const scale of [1,1.25])test.describe(`native row exposure at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('native visible-row delegation preserves first reveal, committed atomic output, selection and exact glyph pixels',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-row-test',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.goto('/')
  await page.evaluate(async()=>{
   const {Terminal}=await import('/node_modules/.vite/deps/@xterm_xterm.js' as string),{installNativeTerminalPaint,exposeNativeTerminal}=await import('/tests/helpers/native-terminal-rows.ts' as string)
   await import('/src/services/terminals.ts' as string)
   const values:any[]=[],original=Element.prototype.replaceChildren
   Element.prototype.replaceChildren=function(this:Element,...nodes:(Node|string)[]){if(this.parentElement?.classList.contains('xterm-rows')){const host=this.closest<HTMLElement>('.terminal-container'),value=values.find(value=>value.host===host);if(value)value.paints++}return original.apply(this,nodes)}
   for(let index=0;index<2;index++){
    const region=document.createElement('div');region.className='terminal-paint-region';region.dataset.nativeRows=String(index);region.style.cssText=`position:fixed;left:${40+index*760}px;top:160px;width:720px;height:390px;z-index:10000;background:#111917`
    const viewport=document.createElement('div');viewport.className='terminal-paint-viewport'
    const host=document.createElement('div');host.className='terminal-container';host.style.cssText='width:720px;height:390px';viewport.append(host);region.append(viewport);document.body.append(region)
    const terminal=new Terminal({cols:70,rows:20,fontSize:14,fontFamily:'"Liberation Mono",monospace',allowProposedApi:true,cursorBlink:false,theme:{background:'#111917',foreground:'#e6ebf0'}});terminal.open(host)
    values.push({region,viewport,host,terminal,paints:0,originalRender:terminal._core._renderService._renderer.value.renderRows,adapter:index===0?installNativeTerminalPaint(terminal,host):undefined})
   }
   await document.fonts.ready
   const f={values,exposeNativeTerminal,restore(){Element.prototype.replaceChildren=original;for(const value of values){value.adapter?.dispose();value.terminal.dispose();value.region.remove()}}}
   ;(window as any).__nativeRows=f
  })
  const write=async(data:string)=>page.evaluate(async data=>{const f=(window as any).__nativeRows;await Promise.all(f.values.map((v:any)=>new Promise<void>(resolve=>v.terminal.write(data,resolve))))},data)
  const settle=()=>page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  const picture=async()=>{
   const a=await page.locator('[data-native-rows="0"]').screenshot(),b=await page.locator('[data-native-rows="1"]').screenshot()
   const diff=await page.evaluate(async({a,b})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const x=c.getContext('2d')!;x.drawImage(image,0,0);return x.getImageData(0,0,c.width,c.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0;for(let i=0;i<x.length;i++){const d=Math.abs(x[i]!-y[i]!);if(d)changed++;max=Math.max(max,d)}return {changed,max}
   },{a:a.toString('base64'),b:b.toString('base64')});expect(diff).toEqual({changed:0,max:0})
  }
  const partial=()=>page.evaluate(()=>{
   const f=(window as any).__nativeRows
   for(const v of f.values){v.viewport.style.cssText='top:280px;height:75px';v.host.style.top='-280px';v.paints=0}
   f.exposeNativeTerminal(f.values[0].host,{x:0,y:280,width:720,height:75})
  })
  await expect(page.locator('[data-native-rows="0"] .terminal-container')).toHaveAttribute('data-terminal-row-paint','native-visible-rows')
  await write('\x1b[?25l\x1b[2J\x1b[H'+Array.from({length:20},(_,i)=>`\x1b[${31+i%7}m${i.toString().padStart(2,'0')} 中文 e\u0301 \x1b[3mItalic-overhang\x1b[23m ABCDEFGHIJKLMNOPQRSTUVWXYZ\x1b[0m`).join('\r\n'));await settle();await picture()
  await partial()
  await write('\x1b[H'+Array.from({length:20},(_,i)=>`\x1b[${31+i%7}m${i.toString().padStart(2,'0')} updated 中文 e\u0301 \x1b[3mItalic-overhang\x1b[23m ABCDEFGHIJKLMNOPQRSTUVWXYZ\x1b[0m`).join('\r\n'));await settle();await picture()
  const paints=await page.evaluate(()=>(window as any).__nativeRows.values.map((v:any)=>v.paints));expect(paints[0]).toBeGreaterThan(0);expect(paints[0]).toBeLessThan(paints[1]*.8)
  // Synchronously assert native rows at the first reveal animation frame.
  // A later screenshot/RAF cannot hide a one-frame stale committed image.
  const reveal=()=>page.evaluate(()=>new Promise<boolean>(resolve=>{
   const f=(window as any).__nativeRows;for(const v of f.values){v.viewport.style.cssText='';v.host.style.top=''}f.exposeNativeTerminal(f.values[0].host)
   requestAnimationFrame(()=>resolve(f.values[0].host.querySelector('.xterm-rows').textContent===f.values[1].host.querySelector('.xterm-rows').textContent))
  }))
  expect(await reveal()).toBe(true);await picture()
  await partial();await write('\x1b[1;1H\x1b[36mCOMMITTED HIDDEN UPDATE\x1b[0m');await settle()
  await write('\x1b[?2026h\x1b[1;1H\x1b[35mHELD ATOMIC OUTPUT\x1b[0m')
  expect(await reveal()).toBe(true);await picture()
  expect(await page.locator('[data-native-rows="0"] .xterm-rows').textContent()).not.toContain('HELD ATOMIC OUTPUT')
  await write('\x1b[?2026l');await settle();await picture();await expect(page.locator('[data-native-rows="0"] .xterm-rows')).toContainText('HELD ATOMIC OUTPUT')
  await page.evaluate(()=>{for(const v of (window as any).__nativeRows.values)v.terminal.select(2,3,40)});await settle();await picture()
  expect(await page.evaluate(()=>{const f=(window as any).__nativeRows,v=f.values[0];v.adapter.dispose();const restored=v.terminal._core._renderService._renderer.value.renderRows===v.originalRender;f.restore();return restored})).toBe(true)
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
