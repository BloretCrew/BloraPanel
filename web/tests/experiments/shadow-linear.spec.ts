import {test,expect,type Page} from '@playwright/test'
import {nativeDeviceScale} from '../../playwright-browser'
import {installLinearShadows} from '../helpers/shadow-linear'

async function stable(page:Page){let previous=await page.screenshot({animations:'disabled',caret:'hide'});for(let i=0;i<8;i++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const current=await page.screenshot({animations:'disabled',caret:'hide'});if(current.equals(previous))return current;previous=current}throw Error('Optical primitive did not settle')}
for(const scale of [1,1.25])test.describe(`linear native shadow at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('constant-axis brush keeps the native optical shadow and all live contents',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'linear-shadow',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<7;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.open({appId:'blora.settings',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}));(window as any).__linearViews=[...document.querySelectorAll('.app-view')]})
  const compare=async(label:string)=>{
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   await installLinearShadows(page);const count=await page.evaluate(()=>(window as any).__linearShadows.snapshot());expect(count.edges).toBe(scale===1?32:0)
   await page.evaluate(()=>(window as any).__linearShadows.setVisible(false));const baseline=await stable(page)
   await page.evaluate(()=>(window as any).__linearShadows.setVisible(true));const candidate=await stable(page)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return {pixels:ctx.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}};const p=await load(a),q=await load(b);let max=0,changed=0,total=0,contentMax=0;const bodies=[...document.querySelectorAll('.window-body')].map(e=>e.getBoundingClientRect());for(let i=0;i<p.pixels.length;i++){const d=Math.abs(p.pixels[i]!-q.pixels[i]!);max=Math.max(max,d);if(d)changed++;total+=d;const pixel=Math.floor(i/4),x=pixel%p.width,y=Math.floor(pixel/p.width);if(bodies.some(r=>x>r.x+2&&x<r.right-2&&y>r.y+2&&y<r.bottom-2))contentMax=Math.max(contentMax,d)}return {max,changed,mean:total/p.pixels.length,contentMax}},{a:baseline.toString('base64'),b:candidate.toString('base64')})
   console.log('LINEAR_SHADOW',label,JSON.stringify(diff));await info.attach(label+'-native.png',{body:baseline,contentType:'image/png'});await info.attach(label+'-linear.png',{body:candidate,contentType:'image/png'})
   expect(diff.max).toBeLessThanOrEqual(scale===1?4:0);expect(diff.mean).toBeLessThanOrEqual(scale===1?.1:0);expect(diff.contentMax).toBe(0);expect(await page.evaluate(()=>(window as any).__linearViews.every((e:Element)=>e.isConnected))).toBe(true)
   await page.evaluate(()=>(window as any).__linearShadows.dispose())
  }
  await compare('stack-light');const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!;await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+380,title.y+35);await compare('held');await page.mouse.up()
  await page.getByRole('combobox',{name:'主题',exact:true}).selectOption('dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark');await page.getByRole('switch',{name:'通透模式'}).uncheck();await compare('solid');expect(errors).toEqual([])
 })
})
