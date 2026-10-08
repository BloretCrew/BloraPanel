import {test,expect} from '@playwright/test'
import {installShadowMasks} from '../helpers/shadow-mask'

for(const scale of [1,1.25])test.describe(`native optical alpha masks at DPR ${scale}`,()=>{
 test.use({deviceScaleFactor:scale})
 test('keeps native application pixels and bounds optical quantization through focus and motion',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'optical-mask',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.locator('.dock-item[aria-label="设置"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await expect.poll(()=>page.locator('.window-shadow-plane[data-shadow-cache]').count()).toBe(2)
  await installShadowMasks(page);await expect(page.locator('[data-native-shadow-mask="ready"]')).toHaveCount(2)
  const compare=async(label:string)=>{
   await page.evaluate(()=>{(window as any).__nativeShadowMaskDiagnostic.setVisible(false)})
   const native=await page.screenshot({animations:'disabled'})
   await page.evaluate(()=>{(window as any).__nativeShadowMaskDiagnostic.setVisible(true)})
   const mask=await page.screenshot({animations:'disabled'})
   const difference=await page.evaluate(async({a,b,scale})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return {pixels:context.getImageData(0,0,image.width,image.height).data,width:image.width}}
    const p=await load(a),q=await load(b);let max=0,changed=0,total=0,contentMax=0
    // Another window's shadow can overlap native content behind it. Only the
    // declared optical rectangles may quantize; all pixels elsewhere stay exact.
    const optical=[...document.querySelectorAll('.shadow-piece')].filter(element=>getComputedStyle(element).display!=='none').map(element=>element.getBoundingClientRect())
    for(let i=0;i<p.pixels.length;i++){const d=Math.abs(p.pixels[i]!-q.pixels[i]!);if(d)changed++;max=Math.max(max,d);total+=d;const x=Math.floor(i/4)%p.width/scale,y=Math.floor(Math.floor(i/4)/p.width)/scale;if(!optical.some(box=>x>=box.left-1&&x<=box.right+1&&y>=box.top-1&&y<=box.bottom+1))contentMax=Math.max(contentMax,d)}
    return {max,changed,total,mean:total/p.pixels.length,contentMax}
   },{a:native.toString('base64'),b:mask.toString('base64'),scale})
   console.log('NATIVE_SHADOW_MASK',label,JSON.stringify(difference))
   await info.attach(`${label}-reference.png`,{body:native,contentType:'image/png'});await info.attach(`${label}-mask.png`,{body:mask,contentType:'image/png'})
   expect(difference.contentMax).toBe(0)
   if(process.env.BLORA_E08_MASK_EXACT_CONTROL==='1')expect(difference.max).toBe(0)
   else{expect(difference.max).toBeLessThanOrEqual(4);expect(difference.mean).toBeLessThanOrEqual(.1)}
  }
  await compare('overlap')
  const frame=page.locator('.app-window.focused'),box=(await frame.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(box.x+220,box.y+15);await page.mouse.down();await page.mouse.move(box.x+245,box.y+28);await compare('held');await page.mouse.up()
  await page.getByRole('combobox',{name:'主题',exact:true}).selectOption('dark')
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');await expect(page.locator('[data-native-shadow-mask="ready"]')).toHaveCount(2)
  await compare('dark')
  await page.evaluate(()=>{(window as any).__nativeShadowMaskDiagnostic.dispose()})
 })
})
