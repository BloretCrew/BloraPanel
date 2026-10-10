import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {installNativeRetention,type NativeRetention} from '../helpers/native-retention'
import {nativeDeviceScale} from '../../playwright-browser'

async function stable(page:Page){let before=await page.screenshot({animations:'disabled',caret:'hide'});for(let i=0;i<8;i++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const next=await page.screenshot({animations:'disabled',caret:'hide'});if(next.equals(before))return next;before=next}throw Error('Native paint retention did not settle')}
const mode:NativeRetention=process.env.BLORA_E08_APP_VIEW_RETENTION==='glass'?'glass-app-views':process.env.BLORA_E08_APP_VIEW_RETENTION==='1'?'app-views':'software'
for(const scale of [1,1.25])test.describe(`native ${mode} paint retention at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('retaining native paint by ownership preserves the whole desktop through motion, resize and appearance changes',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-retention',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<7;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.open({appId:'blora.settings',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}));await new Promise<void>(resolve=>requestAnimationFrame(()=>resolve()));(window as any).__retentionOriginalViews=[...document.querySelectorAll('.app-view')]})
  await expect(page.locator('.app-window')).toHaveCount(8);await installNativeRetention(page,mode)
  const compare=async(label:string)=>{
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   await page.evaluate(()=>(window as any).__nativeRetention.setVisible(false));const original=await stable(page)
   await page.evaluate(()=>(window as any).__nativeRetention.setVisible(true));const candidate=await stable(page)
   if(mode==='software')await expect(page.locator('.app-window.focused')).toHaveCSS('will-change',scale===1?'auto':'transform');else{
    const active=scale===1&&(mode!=='glass-app-views'||await page.locator('html').getAttribute('data-translucency')==='on')
    expect(await page.locator('.app-window:not(.focused)>.window-body>.app-view').evaluateAll((elements,active)=>elements.every(element=>getComputedStyle(element).willChange===(active?'transform':'auto')),active)).toBe(true)
   }
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:original.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_RETENTION',label,JSON.stringify(diff));await info.attach(label+'-original.png',{body:original,contentType:'image/png'});await info.attach(label+'-candidate.png',{body:candidate,contentType:'image/png'});expect(diff).toEqual({max:0,changed:0});expect(await page.evaluate(()=>(window as any).__retentionOriginalViews.every((view:Element)=>view.isConnected))).toBe(true)
  }
  await compare('eight-window-stack')
  const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!;await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+380,title.y+35);await compare('held-reveal');await page.mouse.up()
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark')
  await page.getByRole('switch',{name:'通透模式'}).uncheck();await compare('solid');await page.getByRole('switch',{name:'通透模式'}).check()
  await frame.evaluate(async element=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.geometry((element as HTMLElement).dataset.windowId,{x:360,y:150,width:661,height:611})});await compare('odd-resize')
  await page.evaluate(()=>(window as any).__nativeRetention.dispose());expect(errors).toEqual([])
 })
})
