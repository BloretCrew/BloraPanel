import {test,expect,type Page} from '@playwright/test'
import {installNativeMotion} from '../helpers/native-motion'
import {nativeDeviceScale} from '../../playwright-browser'

async function stable(page:Page){let before=await page.screenshot({animations:'allow',caret:'hide'});for(let i=0;i<8;i++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const next=await page.screenshot({animations:'allow',caret:'hide'});if(next.equals(before))return next;before=next}throw Error('Native motion reference did not settle')}
for(const scale of [1,1.25])test.describe(`native animation motion at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('native animation and original CSS motion preserve full screen, pointer geometry, release and refresh',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-motion',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<7;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.open({appId:'blora.settings',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}));await new Promise<void>(resolve=>requestAnimationFrame(()=>resolve()));(window as any).__motionOriginalViews=[...document.querySelectorAll('.app-view')]})
  await expect(page.locator('.app-window')).toHaveCount(8);await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  await installNativeMotion(page)
  const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!,id=await frame.getAttribute('data-window-id')
  const compare=async(label:string)=>{
   const state=await page.evaluate(()=>(window as any).__nativeMotion.snapshot());expect(state.maxGeometryError).toBeLessThan(.01);if(scale===1)expect(state.active).toBe(1);else expect(state.active).toBe(0)
   const box=await frame.boundingBox(),candidate=await stable(page)
   await page.evaluate(()=>(window as any).__nativeMotion.setVisible(false));const original=await stable(page);expect(await frame.boundingBox()).toEqual(box)
   expect(await frame.evaluate(element=>element.getAnimations().length)).toBe(0)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:original.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_MOTION',label,JSON.stringify({...state,...diff}));await info.attach(label+'-native-css.png',{body:original,contentType:'image/png'});await info.attach(label+'-animation.png',{body:candidate,contentType:'image/png'});expect(diff).toEqual({max:0,changed:0});expect(await page.evaluate(()=>(window as any).__motionOriginalViews.every((view:Element)=>view.isConnected))).toBe(true)
   await page.evaluate(()=>(window as any).__nativeMotion.setVisible(true))
  }
  await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+237,title.y+46);await compare('held-diagonal');await page.mouse.move(title.x+114,title.y+36);await compare('reverse-reveal');await page.mouse.move(title.x+236,title.y+18);await compare('held-horizontal');await page.mouse.up()
  await expect(frame).not.toHaveClass(/moving/);expect(await frame.evaluate(element=>element.getAnimations().length)).toBe(0)
  const committed=await frame.boundingBox();await page.reload();await expect(page.locator(`[data-window-id="${id}"]`)).toBeVisible();expect(await page.locator(`[data-window-id="${id}"]`).boundingBox()).toEqual(committed)
  expect(errors).toEqual([])
 })
})
