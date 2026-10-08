import {test,expect} from '@playwright/test'
import {nativeCssShadow,nativeCssPlaneShadow} from '../helpers/native-css-shadow'

for(const scale of [1,1.25])test.describe(`original CSS elevation at DPR ${scale}`,()=>{
 test.use({deviceScaleFactor:scale})
 test('native CSS shadow keeps native application pixels through overlap, motion and dark appearance',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-css-shadow',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'])}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.locator('.dock-item[aria-label="设置"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  const compare=async(label:string)=>{
   const original=await page.screenshot({animations:'disabled'})
   const style=await page.addStyleTag({content:process.env.BLORA_E08_CSS_SHADOW_PLANE==='1'?nativeCssPlaneShadow:nativeCssShadow}),native=await page.screenshot({animations:'disabled'})
   await style.evaluate(element=>element.parentNode?.removeChild(element))
   const difference=await page.evaluate(async({a,b,scale})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const c=canvas.getContext('2d')!;c.drawImage(image,0,0);return {pixels:c.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}}
    const p=await load(a),q=await load(b),box=document.querySelector('.app-window.focused')!.getBoundingClientRect();let max=0,changed=0,total=0,contentMax=0
    for(let i=0;i<p.pixels.length;i++){const d=Math.abs(p.pixels[i]!-q.pixels[i]!);max=Math.max(max,d);if(d)changed++;total+=d;const x=Math.floor(i/4)%p.width/scale,y=Math.floor(Math.floor(i/4)/p.width)/scale;if(x>box.left+32&&x<box.right-32&&y>box.top+32&&y<box.bottom-32)contentMax=Math.max(contentMax,d)}
    return {max,changed,total,contentMax,mean:total/p.pixels.length}
   },{a:original.toString('base64'),b:native.toString('base64'),scale})
   console.log('NATIVE_CSS_SHADOW',label,JSON.stringify(difference));await info.attach(`${label}-patches.png`,{body:original,contentType:'image/png'});await info.attach(`${label}-css.png`,{body:native,contentType:'image/png'})
   expect(difference.contentMax).toBe(0)
   if(process.env.BLORA_E08_CSS_SHADOW_EXACT==='1')expect(difference.max).toBe(0)
   else{expect(difference.max).toBeLessThanOrEqual(4);expect(difference.mean).toBeLessThanOrEqual(.1)}
  }
  await compare('overlap')
  const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(title.x+270,title.y+30);await compare('held');await page.mouse.up()
  await page.getByRole('combobox',{name:'主题',exact:true}).selectOption('dark')
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await compare('dark')
 })
})
