import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {installShadowTail} from '../helpers/shadow-tail'

async function stableScreenshot(page:Page){
 let previous=await page.screenshot({animations:'disabled',caret:'hide'})
 for(let attempt=0;attempt<8;attempt++){
  await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  const current=await page.screenshot({animations:'disabled',caret:'hide'})
  if(current.equals(previous))return current
  previous=current
 }
 throw Error('Native optical reference did not settle before comparison')
}
for(const scale of [1,1.25])test.describe(`native shadow alpha tail at DPR ${scale}`,()=>{
 test.use({deviceScaleFactor:scale})
 test('alpha-one shadow tail trim keeps native application paint exact and bounds optical quantization',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'shadow-tail',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.locator('.dock-item[aria-label="设置"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  const compare=async(label:string)=>{
   await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.app-window>.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   await installShadowTail(page)
   const removed=await page.evaluate(()=>(window as any).__nativeShadowTail.trimmedPixels);if(scale===1)expect(removed).toBeGreaterThan(0);else expect(removed).toBe(0)
   await page.evaluate(()=>(window as any).__nativeShadowTail.setVisible(false))
   const optical=await page.locator('.app-window>.window-shadow-plane>.shadow-piece').evaluateAll(elements=>elements.map(element=>{const r=element.getBoundingClientRect();return {x:r.x-1,y:r.y-1,right:r.right+1,bottom:r.bottom+1}}))
   const reference=await stableScreenshot(page)
   await page.evaluate(()=>(window as any).__nativeShadowTail.setVisible(true))
   const candidate=await stableScreenshot(page)
   const difference=await page.evaluate(async({a,b,scale,optical})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const c=canvas.getContext('2d')!;c.drawImage(image,0,0);return {pixels:c.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}}
    const p=await load(a),q=await load(b);let max=0,total=0,changed=0,contentMax=0
    for(let i=0;i<p.pixels.length;i++){const d=Math.abs(p.pixels[i]!-q.pixels[i]!);max=Math.max(max,d);total+=d;if(d)changed++;const x=Math.floor(i/4)%p.width/scale,y=Math.floor(Math.floor(i/4)/p.width)/scale;if(!optical.some(r=>x>=r.x&&x<=r.right&&y>=r.y&&y<=r.bottom))contentMax=Math.max(contentMax,d)}
    return {max,total,changed,contentMax,mean:total/p.pixels.length}
   },{a:reference.toString('base64'),b:candidate.toString('base64'),scale,optical})
   console.log('NATIVE_SHADOW_TAIL',label,JSON.stringify({...difference,removed}));await info.attach(`${label}-reference.png`,{body:reference,contentType:'image/png'});await info.attach(`${label}-candidate.png`,{body:candidate,contentType:'image/png'})
   expect(difference.contentMax).toBe(0);expect(difference.max).toBeLessThanOrEqual(4);expect(difference.mean).toBeLessThanOrEqual(.1)
   await page.evaluate(()=>(window as any).__nativeShadowTail.dispose())
  }
  await compare('overlap')
  const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(title.x+245,title.y+28);await compare('held');await page.mouse.up()
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark')
 })
})
