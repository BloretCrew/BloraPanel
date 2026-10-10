import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'

test.beforeEach(async({page})=>{
 await page.route('**/api/v1/**',route=>{const path=new URL(route.request().url()).pathname;return route.fulfill({json:path.endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}})})
})
test('late native Dock decoding cannot publish stale optics and failures retain the native path',async({page})=>{
 const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
 await page.addInitScript(()=>{
  const original=HTMLImageElement.prototype.decode,revoked=new Set<string>(),revoke=URL.revokeObjectURL
  URL.revokeObjectURL=function(value:string){revoked.add(value);return revoke.call(this,value)}
  let held=false
  HTMLImageElement.prototype.decode=async function(){
   const dock=this.src.startsWith('data:image/svg+xml')&&this.src.includes('foreignObject')&&this.src.includes('stroke-width')
   if(dock&&(window as any).__failDockDecode){(window as any).__dockDecodeFailures=((window as any).__dockDecodeFailures||0)+1;throw Error('injected native Dock decode failure')}
   await original.call(this)
   if(dock&&!held){held=true;(window as any).__heldDockDecode=true;await new Promise(resolve=>setTimeout(resolve,1500))}
  }
  ;(window as any).__dockRevoked=(value:string)=>revoked.has(value)
 })
 await page.goto('/');const surface=page.locator('.dock-surface')
 await expect.poll(()=>page.evaluate(()=>(window as any).__heldDockDecode===true)).toBe(true)
 await page.evaluate(()=>{const root=document.documentElement;root.dataset.theme='dark';root.dataset.palette='sand';root.dataset.translucency='off'})
 await expect(surface).toHaveAttribute('data-dock-cache','ready')
 const previous=await surface.locator('.dock-optical-cache').getAttribute('src')
 expect(previous).toMatch(/^blob:/)
 await page.evaluate(()=>{(window as any).__failDockDecode=true;document.documentElement.dataset.palette='rose'})
 await expect.poll(()=>page.evaluate(()=>(window as any).__dockDecodeFailures||0)).toBeGreaterThan(0)
 await expect(surface).not.toHaveAttribute('data-dock-cache','ready')
 await expect(surface.locator('svg')).toHaveCSS('visibility','visible')
 await expect.poll(()=>page.evaluate(value=>(window as any).__dockRevoked(value),previous)).toBe(true)
 await expect(surface.locator('.dock-optical-cache')).toBeHidden()
 await page.evaluate(()=>{(window as any).__failDockDecode=false;document.documentElement.dataset.palette='ice';document.documentElement.dataset.theme='light';document.documentElement.dataset.translucency='on'})
 await expect(surface).toHaveAttribute('data-dock-cache','ready')
 await expect.poll(()=>surface.locator('.dock-optical-cache').evaluate(element=>{const image=element as HTMLImageElement;return image.complete&&image.naturalWidth>0})).toBe(true)
 await page.locator('[data-app="blora.instances"]').click();await expect(page.locator('.app-window')).toBeVisible()
 expect(errors).toEqual([])
})
for(const scale of [1,1.25])test.describe(`native Dock optics at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('native contour and shadow stay intact through theme changes, fractional centering and live applications',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  const surface=page.locator('.dock-surface')
  const settle=async()=>{
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
   if(scale===1)await expect(surface).toHaveAttribute('data-dock-cache','ready')
   else await expect(surface).not.toHaveAttribute('data-dock-cache','ready')
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  }
  const compare=async(label:string)=>{
   await settle()
   const clip={x:0,y:800,width:1600,height:200},cached=await page.screenshot({clip,animations:'disabled'})
   const native=await page.addStyleTag({content:'.dock-surface[data-dock-cache="ready"]>svg{display:block!important;visibility:visible!important}.dock-optical-cache{display:none!important}'})
   const original=await page.screenshot({clip,animations:'disabled'});await native.evaluate(element=>(element as HTMLElement).remove())
   await info.attach(`${label}-cached.png`,{body:cached,contentType:'image/png'});await info.attach(`${label}-native.png`,{body:original,contentType:'image/png'})
   const difference=await page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return context.getImageData(0,0,canvas.width,canvas.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0,total=0
    for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta);total+=delta}
    return {changed,max,total}
   },{a:cached.toString('base64'),b:original.toString('base64')})
   console.log('NATIVE_DOCK_OPTICS',label,JSON.stringify(difference))
   // Preserve the exact-original comparison as an explicit negative control.
   // Native filter -> native PNG has small premultiplied-alpha quantization;
   // this optical-only boundary is not application or glyph rasterization.
   if(process.env.BLORA_E08_DOCK_EXACT_CONTROL==='1'||scale!==1)expect(difference).toEqual({changed:0,max:0,total:0})
   else {expect(difference.max).toBeLessThanOrEqual(4);expect(difference.total/(1600*200*4)).toBeLessThanOrEqual(.1)}
  }
  await compare('light')
  await page.locator('[data-app="blora.instances"]').click();await compare('live-app')
  await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'设置',exact:true}).click()
  const theme=page.getByRole('combobox',{name:'主题',exact:true}),palette=page.getByRole('combobox',{name:'主题配色'}),toggle=page.getByRole('switch',{name:'通透模式'})
  await selectStyledOption(page,theme,'dark');await selectStyledOption(page,palette,'sand');await compare('dark-sand')
  await toggle.uncheck();await compare('solid')
  await selectStyledOption(page,theme,'light');await selectStyledOption(page,palette,'ice');await toggle.check();await compare('restored')
  await selectStyledOption(page,theme,'dark');await selectStyledOption(page,palette,'rose');await selectStyledOption(page,theme,'light');await selectStyledOption(page,palette,'ice');await compare('rapid-change')
  await expect(page.locator('.taskbar').getByRole('button',{name:/实例中心/})).toBeVisible()
  expect(errors).toEqual([])
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
