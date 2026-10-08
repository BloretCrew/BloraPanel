import {test,expect} from '@playwright/test'

for(const scale of [1,1.25])test.describe(`native SVG artwork retention at DPR ${scale}`,()=>{
 test.use({deviceScaleFactor:scale})
 test('native retained SVG application icons keep exact full-screen pixels in both appearances',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-icon',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.locator('.dock-item[aria-label="设置"]').click();await page.locator('.launcher-button').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  expect(await page.locator('.application-art').count()).toBeGreaterThan(20)
  const compare=async(label:string)=>{
   await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
   const stable=async()=>{let previous=await page.screenshot({animations:'disabled',caret:'hide'});for(let attempt=0;attempt<8;attempt++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const next=await page.screenshot({animations:'disabled',caret:'hide'});if(next.equals(previous))return next;previous=next}throw Error('Native artwork reference did not settle')}
   const reference=await stable(),style=await page.addStyleTag({content:'.application-art{will-change:transform}'})
   const candidate=await stable()
   const difference=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return context.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:reference.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_ICON_LAYERS',label,JSON.stringify(difference));await info.attach(`${label}-original.png`,{body:reference,contentType:'image/png'});await info.attach(`${label}-retained.png`,{body:candidate,contentType:'image/png'})
   expect(difference).toEqual({max:0,changed:0});await style.evaluate(element=>(element as HTMLElement).remove())
  }
  await compare('light')
  await page.locator('.launcher-button').click();await page.getByRole('combobox',{name:'主题',exact:true}).selectOption('dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await page.locator('.launcher-button').click();await compare('dark')
 })
})
