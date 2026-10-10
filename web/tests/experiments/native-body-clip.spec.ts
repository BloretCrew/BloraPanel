import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {installNativeBodyClip} from '../helpers/native-body-clip'
import {nativeDeviceScale} from '../../playwright-browser'

async function stable(page:Page){
 let before=await page.screenshot({animations:'disabled',caret:'hide'})
 for(let n=0;n<8;n++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const next=await page.screenshot({animations:'disabled',caret:'hide'});if(next.equals(before))return next;before=next}
 throw Error('Native reference did not settle')
}
for(const scale of [1,1.25])test.describe(`native body rectangle clip at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('original DOM, dimensions, content and round contour remain exact in independent full-native comparisons',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-clip',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.open({appId:'blora.instances',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:110,y:150,width:760,height:590});d.open({appId:'blora.settings',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:450,y:250,width:740,height:580});await new Promise<void>(resolve=>requestAnimationFrame(()=>resolve()));(window as any).__nativeOriginalViews=[...document.querySelectorAll('.app-view')];const body=document.querySelector<HTMLElement>('.window-body')!;const content=document.createElement('div');content.dataset.nativeClipContent='true';content.style.cssText='position:absolute;inset:70px 22px 0;background:var(--surface-container);overflow:auto';content.innerHTML=Array.from({length:80},(_,i)=>`<p style="margin:0;height:32px;color:${['#b8472c','#258965','#6977c5'][i%3]}">${i} 中文 é <i>italic native text</i><b> original glyphs</b></p>`).join('');body.querySelector('.app-view')!.append(content)})
  await installNativeBodyClip(page,process.env.BLORA_E08_NATIVE_FLAT_BODY==='1')
  const geometry=()=>page.locator('.app-view').evaluateAll(elements=>elements.map(element=>{const r=element.getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height,scroll:element.querySelector('[data-native-clip-content]')?.scrollTop}}))
  const compare=async(label:string)=>{
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.app-window>.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   await page.evaluate(()=>(window as any).__nativeBodyClip.setVisible(false));const original=await stable(page),before=await geometry()
   // Explicitly prove that this reference is full native, not the new child
   // clip accidentally left active under an overridden parent clip-path.
   expect(await page.locator('.app-view').evaluateAll(elements=>elements.every(element=>getComputedStyle(element).clip==='auto'))).toBe(true)
   await page.evaluate(()=>(window as any).__nativeBodyClip.setVisible(true));const candidate=await stable(page),counts=await page.evaluate(()=>(window as any).__nativeBodyClip.snapshot())
   expect(await geometry()).toEqual(before);expect(counts.enabled).toBe(scale===1&&label!=='solid'?counts.tracked:0)
   if(scale===1&&label==='eight-window-stack')expect(counts.viewMasks).toBeGreaterThan(0)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:original.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_BODY_CLIP',label,JSON.stringify({...counts,...diff}));await info.attach(label+'-original.png',{body:original,contentType:'image/png'});await info.attach(label+'-candidate.png',{body:candidate,contentType:'image/png'});expect(diff).toEqual({max:0,changed:0})
   expect(await page.evaluate(()=>(window as any).__nativeOriginalViews.every((view:Element)=>view.isConnected))).toBe(true)
  }
  await compare('overlap')
  const title=(await page.locator('.app-window.focused .window-titlebar').boundingBox())!;await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+275,title.y+35);await compare('held-reveal');await page.mouse.up()
  await page.locator('[data-native-clip-content]').evaluate(element=>{element.scrollTop=342});await compare('native-scroll')
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.geometry(d.state.activeWindowId,{x:400,y:210,width:661,height:611})});await compare('odd-resize')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark')
  await page.getByRole('switch',{name:'通透模式'}).uncheck();await compare('solid')
  await page.getByRole('switch',{name:'通透模式'}).check()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<6;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}))});await expect(page.locator('.app-window')).toHaveCount(8);await compare('eight-window-stack')
  const stacked=(await page.locator('.app-window.focused .window-titlebar').boundingBox())!;await page.mouse.move(stacked.x+180,stacked.y+15);await page.mouse.down();await page.mouse.move(stacked.x+225,stacked.y+35);await compare('stack-held');await page.mouse.up()
  await page.evaluate(()=>(window as any).__nativeBodyClip.dispose());expect(errors).toEqual([])
 })
})
