import {test,expect} from '@playwright/test'
// Explicit unadopted alpha-partition candidate, not a production paint path.
import {installShadowAlphaDiagnostic} from '../real/shadow-alpha-diagnostic'

test.beforeEach(async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}}))
 await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
})
for(const scale of [1,1.25])test.describe(`native alpha footprint at DPR ${scale}`,()=>{
 test.use({viewport:{width:1600,height:1000},...nativeDeviceScale(scale)})
 test('native alpha partitions preserve every original pixel through overlap, held reveal, resize and fallback',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const d=useDesktop();d.open({appId:'blora.instances',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:70,y:130,width:900,height:700});d.open({appId:'blora.settings',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:350,y:230,width:900,height:650})})
  const windows=page.locator('.app-window'),upper=page.locator('.app-window.focused')
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  for(const plane of await windows.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await page.evaluate<void,2|4>(installShadowAlphaDiagnostic,process.env.BLORA_E08_SHADOW_ALPHA_LEAVES==='4'?4:2)
  const settle=async()=>{
   if(scale===1)await expect(page.locator('[data-native-alpha-bands="ready"]')).toHaveCount(8)
   else await expect(page.locator('[data-native-alpha-bands="ready"]')).toHaveCount(0)
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  }
  const compare=async(label:string)=>{
   await settle();const cached=await page.screenshot({animations:'disabled'})
   await page.evaluate(()=>(window as any).__shadowAlphaDiagnostic.pause());const original=await page.screenshot({animations:'disabled'})
   await page.evaluate(()=>(window as any).__shadowAlphaDiagnostic.resume());await settle()
   await info.attach(`${label}-partitioned.png`,{body:cached,contentType:'image/png'});await info.attach(`${label}-original.png`,{body:original,contentType:'image/png'})
   const difference=await page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const x=c.getContext('2d')!;x.drawImage(image,0,0);return x.getImageData(0,0,c.width,c.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0;for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta)}return {changed,max}
   },{a:cached.toString('base64'),b:original.toString('base64')})
   console.log('NATIVE_SHADOW_ALPHA',label,JSON.stringify(difference));expect(difference).toEqual({changed:0,max:0})
  }
  await compare('overlap')
  if(scale===1){
   const area=await page.locator('[data-native-alpha-bands="ready"]').evaluateAll(elements=>{
    let original=0,partitioned=0
    for(const element of elements){const rect=element.getBoundingClientRect();original+=rect.width*rect.height;for(const node of element.children){const band=node.getBoundingClientRect();partitioned+=band.width*band.height}}
    return {original,partitioned}
   });expect(area.partitioned).toBeLessThan(area.original)
  }
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(1350,title.y+15);await expect(upper).toHaveClass(/moving/);await compare('held-reveal');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);const d=useDesktop();d.geometry(d.state.activeWindowId,{x:250,y:210,width:800,height:620})});await compare('resized')
  await page.evaluate(()=>{document.documentElement.dataset.theme='dark';document.documentElement.dataset.palette='sand'})
  for(const plane of await windows.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await compare('dark')
  await page.evaluate(()=>(window as any).__shadowAlphaDiagnostic.dispose());await expect(page.locator('[data-native-alpha-bands="ready"]')).toHaveCount(0);await expect(page.locator('.native-shadow-alpha-band')).toHaveCount(0)
  expect(errors).toEqual([])
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
