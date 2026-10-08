import {test,expect} from '@playwright/test'
// Explicit unadopted grouped-shadow candidate; preserve the original oracle.
import {installShadowCompositeDiagnostic} from '../real/shadow-composite-diagnostic'

test.beforeEach(async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}}))
 await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
})
for(const scale of [1,1.25])test.describe(`native perimeter composition at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('native optical strips retain visible content through overlap, held reveal, focus and fallback',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const desktop=useDesktop();desktop.open({appId:'blora.instances',disposition:'new-window'});const lower=desktop.state.activeWindowId;desktop.geometry(lower,{x:70,y:130,width:900,height:700});desktop.open({appId:'blora.settings',disposition:'new-window'});const upper=desktop.state.activeWindowId;desktop.geometry(upper,{x:350,y:230,width:900,height:650})})
  const windows=page.locator('.app-window'),upper=page.locator('.app-window.focused')
  for(const plane of await windows.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await page.evaluate<void,'grouped'|'partitioned'>(installShadowCompositeDiagnostic,process.env.BLORA_E08_SHADOW_PARTITION==='1'?'partitioned':'grouped')
  const settle=async()=>{
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   if(scale===1)for(const plane of await windows.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-composite','ready')
   else await expect(page.locator('[data-shadow-composite="ready"]')).toHaveCount(0)
  }
  const compare=async(label:string)=>{
   await settle();const cached=await page.screenshot({animations:'disabled'})
   await page.evaluate(()=>(window as any).__shadowCompositeDiagnostic.pause())
   const original=await page.screenshot({animations:'disabled'})
   await page.evaluate(()=>(window as any).__shadowCompositeDiagnostic.resume());await settle()
   await info.attach(`${label}-composed.png`,{body:cached,contentType:'image/png'});await info.attach(`${label}-native.png`,{body:original,contentType:'image/png'})
   const difference=await page.evaluate(async({a,b,scale})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return {pixels:context.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}}
    const x=await load(a),y=await load(b),box=document.querySelector('.app-window.focused')!.getBoundingClientRect();let changed=0,max=0,total=0,contentMax=0
    for(let i=0;i<x.pixels.length;i++){const delta=Math.abs(x.pixels[i]!-y.pixels[i]!);if(delta)changed++;max=Math.max(max,delta);total+=delta;const px=Math.floor(i/4)%x.width/scale,py=Math.floor(Math.floor(i/4)/x.width)/scale;if(px>box.left+32&&px<box.right-32&&py>box.top+32&&py<box.bottom-32)contentMax=Math.max(contentMax,delta)}
    return {changed,max,total,contentMax}
   },{a:cached.toString('base64'),b:original.toString('base64'),scale})
   console.log('NATIVE_SHADOW_COMPOSITION',label,JSON.stringify(difference));expect(difference.contentMax).toBe(0)
   if(scale!==1||process.env.BLORA_E08_COMPOSITE_EXACT_CONTROL==='1')expect(difference).toEqual({changed:0,max:0,total:0,contentMax:0})
   else {expect(difference.max).toBeLessThanOrEqual(4);expect(difference.total/(1600*1000*4)).toBeLessThanOrEqual(.1)}
  }
  await compare('overlap')
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(1350,title.y+15);await expect(upper).toHaveClass(/moving/);await compare('held-reveal');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const desktop=useDesktop();desktop.geometry(desktop.state.activeWindowId,{x:250,y:210,width:800,height:620})});await compare('resized')
  await page.evaluate(()=>{const root=document.documentElement;root.dataset.theme='dark';root.dataset.palette='sand'})
  for(const plane of await windows.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await compare('dark')
  await page.evaluate(()=>(window as any).__shadowCompositeDiagnostic.dispose());await expect(page.locator('[data-shadow-composite="ready"]')).toHaveCount(0);await expect(page.locator('.shadow-composite-image')).toHaveCount(0)
  expect(errors).toEqual([])
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
