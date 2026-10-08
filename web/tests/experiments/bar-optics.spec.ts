import {test,expect} from '@playwright/test'
// Explicit unadopted bar-shadow candidate; retain strict failure comparisons.
import {installBarShadowDiagnostic} from '../real/bar-shadow-diagnostic'

test.beforeEach(async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}}))
 // Freeze displayed clock text only; event timestamps, RAF and Date.now stay
 // native so this cannot alter gesture/performance timing.
 await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
})
for(const scale of [1,1.25])test.describe(`native fixed bar shadow at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('content stays native while cached contour matches its original shadow',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  const bar=page.locator('.topbar');await expect(bar).toBeVisible()
  await page.evaluate(()=>{(window as any).__originalBarButton=document.querySelector('.workspace-switch')})
  await page.evaluate<void,'png'|'native-svg'|'native-full'>(installBarShadowDiagnostic,process.env.BLORA_E08_BAR_SOURCE==='native-svg'?'native-svg':process.env.BLORA_E08_BAR_SOURCE==='png'?'png':'native-full')
  const settle=async()=>{
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
   for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   if(scale===1&&await page.locator('html').getAttribute('data-translucency')==='on')await expect(bar).toHaveAttribute('data-bar-shadow-cache','ready')
   else await expect(bar).not.toHaveAttribute('data-bar-shadow-cache','ready')
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  }
  const compare=async(label:string)=>{
   await settle()
   if(scale===1&&await bar.getAttribute('data-bar-shadow-cache')==='ready')expect(await bar.locator('.bar-shadow-diagnostic').evaluate(element=>{const image=element as HTMLImageElement;return image.complete&&image.naturalWidth>0})).toBe(true)
   const clip={x:0,y:0,width:1600,height:130},cached=await page.screenshot({clip,animations:'disabled'})
   await page.evaluate(()=>(window as any).__barShadowDiagnostic.pause())
   const native=await page.screenshot({clip,animations:'disabled'})
   await page.evaluate(()=>(window as any).__barShadowDiagnostic.resume());await settle()
   await info.attach(`${label}-cached.png`,{body:cached,contentType:'image/png'});await info.attach(`${label}-native.png`,{body:native,contentType:'image/png'})
   const difference=await page.evaluate(async({a,b,scale})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return {pixels:context.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}}
    const x=await load(a),y=await load(b);let changed=0,max=0,total=0,contentMax=0
    const box=document.querySelector('.topbar')!.getBoundingClientRect()
    for(let i=0;i<x.pixels.length;i++){const delta=Math.abs(x.pixels[i]!-y.pixels[i]!);if(delta)changed++;max=Math.max(max,delta);total+=delta
     const px=Math.floor(i/4)%x.width/scale,py=Math.floor(Math.floor(i/4)/x.width)/scale
     if(px>box.left+32&&px<box.right-32&&py>box.top+8&&py<box.bottom-8)contentMax=Math.max(contentMax,delta)
    }return {changed,max,total,contentMax}
   },{a:cached.toString('base64'),b:native.toString('base64'),scale})
   console.log('NATIVE_BAR_OPTICS',label,JSON.stringify(difference))
   // Content/glyphs retain strict equality. Only the content-free native
   // shadow may incur the already-declared PNG alpha quantization boundary.
   expect(difference.contentMax).toBe(0)
   if(process.env.BLORA_E08_BAR_EXACT_CONTROL==='1'||scale!==1)expect(difference).toEqual({changed:0,max:0,total:0,contentMax:0})
   else {expect(difference.max).toBeLessThanOrEqual(4);expect(difference.total/(1600*130*4*scale*scale)).toBeLessThanOrEqual(.1)}
  }
  await compare('light')
  await page.locator('[data-app="blora.instances"]').click();await compare('live-app')
  await bar.getByRole('button',{name:'切换工作区',exact:true}).click();await expect(page.locator('.workspace-menu')).toBeVisible();await bar.getByRole('button',{name:'切换工作区',exact:true}).click()
  await bar.getByRole('button',{name:'账号菜单',exact:true}).click();await expect(page.getByRole('menu',{name:'账号菜单'})).toBeVisible();await page.keyboard.press('Escape')
  expect(await page.evaluate(()=>(window as any).__originalBarButton===document.querySelector('.workspace-switch'))).toBe(true)
  await page.evaluate(()=>{document.documentElement.dataset.theme='dark';document.documentElement.dataset.palette='sand'})
  await compare('dark-sand')
  await page.evaluate(()=>document.documentElement.dataset.translucency='off');await compare('solid-native')
  await page.evaluate(()=>{const root=document.documentElement;root.dataset.theme='light';root.dataset.palette='ice';root.dataset.translucency='on'});await compare('restored')
  await page.setViewportSize({width:1599,height:1000});await compare('odd-width')
  await page.evaluate(()=>(window as any).__barShadowDiagnostic.dispose());await expect(bar).not.toHaveAttribute('data-bar-shadow-cache','ready');await expect(bar.locator('.bar-shadow-diagnostic')).toHaveCount(0)
  expect(errors).toEqual([])
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
