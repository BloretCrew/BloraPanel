import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {installOutsideShadows} from '../helpers/shadow-outside'

// Unadopted structural prototype. Preserve its exact native comparison and
// failed evidence separately from the installed product's browser regressions.
// Run explicitly with playwright.experiments.config.ts; no tolerance/skip.
for(const scale of [1,1.25])test.describe(`disjoint outside optical shadow at DPR ${scale}`,()=>{
 test.use({deviceScaleFactor:scale})
test('outside shadow composition preserves full-screen pixels through motion and dark appearance',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'optical-sibling',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.locator('.dock-item[aria-label="设置"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await expect.poll(()=>page.locator('.window-shadow-plane[data-shadow-cache]').count()).toBe(2)
  await installOutsideShadows(page)
  expect(await page.evaluate(()=>(window as any).__nativeOutsideShadowDiagnostic.count())).toBe(process.env.BLORA_BROWSER==='firefox'&&scale===1?2:0)
  const compare=async()=>{
   await page.evaluate(()=>{(window as any).__nativeOutsideShadowDiagnostic.setVisible(false)})
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const native=await page.screenshot()
   await page.evaluate(()=>{(window as any).__nativeOutsideShadowDiagnostic.setVisible(true)})
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const sibling=await page.screenshot()
   await info.attach('native-outside-reference.png',{body:native,contentType:'image/png'});await info.attach('native-outside-candidate.png',{body:sibling,contentType:'image/png'})
   return page.evaluate(async({a,b})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return context.getImageData(0,0,image.width,image.height).data}
    const p=await load(a),q=await load(b);let max=0,changed=0
    for(let i=0;i<p.length;i++){const difference=Math.abs(p[i]!-q[i]!);if(difference)changed++;max=Math.max(max,difference)}
    return {max,changed}
   },{a:native.toString('base64'),b:sibling.toString('base64')})
  }
  expect(await compare()).toEqual({max:0,changed:0})
  const frame=page.locator('.app-window.focused'),box=(await frame.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(box.x+220,box.y+15);await page.mouse.down();await page.mouse.move(box.x+245,box.y+28)
  expect(await compare()).toEqual({max:0,changed:0});await page.mouse.up()
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark')
  await expect(page.locator('html')).toHaveAttribute('data-theme','dark')
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await expect.poll(()=>page.locator('.app-window>.window-shadow-plane[data-shadow-cache]').evaluateAll(elements=>elements.every(element=>{
   const key=JSON.parse((element as HTMLElement).dataset.shadowCache!) as [string,string,number,number]
   const frame=element.parentElement!,probe=document.createElement('span');probe.style.boxShadow=getComputedStyle(frame).getPropertyValue('--window-shadow');document.body.append(probe)
   try{return key[0]===getComputedStyle(probe).boxShadow}finally{probe.remove()}
  }))).toBe(true)
  expect(await compare()).toEqual({max:0,changed:0})
  await page.evaluate(()=>{(window as any).__nativeOutsideShadowDiagnostic.dispose()})
 })
})
