import {test,expect,type Page} from '@playwright/test'
import {installStaticShadowUnderlay} from '../helpers/static-shadow-underlay'
import {installStaticShadowScene} from '../helpers/static-shadow-scene'
import {nativeDeviceScale} from '../../playwright-browser'

async function stable(page:Page){
 let before=await page.screenshot({animations:'disabled',caret:'hide'})
 for(let n=0;n<8;n++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const next=await page.screenshot({animations:'disabled',caret:'hide'});if(next.equals(before))return next;before=next}
 throw Error('Native optical reference did not settle')
}
for(const scale of [1,1.25])test.describe(`opaque static shadow underlay at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('wallpaper-only underlay keeps application pixels exact and safely falls back on movement, resize and structure',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'static-shadow',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.open({appId:'blora.instances',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:110,y:150,width:760,height:590});d.open({appId:'blora.settings',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:450,y:250,width:740,height:580})})
  const compare=async(label:string)=>{
   await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.app-window>.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   await page.evaluate(()=>(window as any).__staticShadowUnderlay?.dispose());if(process.env.BLORA_E08_SHADOW_SCENE==='1')await installStaticShadowScene(page,process.env.BLORA_E08_SHADOW_SCENE_HULL==='1');else await installStaticShadowUnderlay(page)
   const counts=await page.evaluate(()=>(window as any).__staticShadowUnderlay.snapshot());if(scale===1&&['overlap','stack','stack-held'].includes(label)){expect(counts.eligible).toBeGreaterThan(0);expect(counts.opaquePixels).toBeGreaterThan(0)}if(scale!==1)expect(counts.eligible).toBe(0)
   await page.evaluate(()=>(window as any).__staticShadowUnderlay.setVisible(false));const original=await stable(page)
   await page.evaluate(()=>(window as any).__staticShadowUnderlay.setVisible(true));const candidate=await stable(page)
   const difference=await page.evaluate(async({a,b,scale})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return {data:ctx.getImageData(0,0,canvas.width,canvas.height).data,width:canvas.width}}
    const p=await load(a),q=await load(b),bodies=[...document.querySelectorAll('.app-window')].map(e=>{const r=e.getBoundingClientRect();return {x:r.x+32,y:r.y+32,right:r.right-32,bottom:r.bottom-32}}),icons=[...document.querySelectorAll('.desktop-icon')].map(e=>{const r=e.getBoundingClientRect();return {x:r.x-8,y:r.y-8,right:r.right+8,bottom:r.bottom+8}});let max=0,total=0,changed=0,contentMax=0
    const contents=[...bodies,...icons]
    for(let i=0;i<p.data.length;i++){const d=Math.abs(p.data[i]!-q.data[i]!);if(!d)continue;max=Math.max(max,d);total+=d;changed++;const x=Math.floor(i/4)%p.width/scale,y=Math.floor(Math.floor(i/4)/p.width)/scale;if(contents.some(r=>x>=r.x&&x<=r.right&&y>=r.y&&y<=r.bottom))contentMax=Math.max(contentMax,d)}return {max,mean:total/p.data.length,changed,contentMax}
   },{a:original.toString('base64'),b:candidate.toString('base64'),scale})
   console.log('STATIC_SHADOW_UNDERLAY',label,JSON.stringify({...counts,...difference}));await info.attach(label+'-original.png',{body:original,contentType:'image/png'});await info.attach(label+'-candidate.png',{body:candidate,contentType:'image/png'})
   expect(difference.contentMax).toBe(0);expect(difference.max).toBeLessThanOrEqual(4);expect(difference.mean).toBeLessThanOrEqual(.1)
   if(scale!==1)expect(difference.changed).toBe(0)
  }
  await compare('overlap')
  const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+280,title.y+35);await compare('held');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.geometry(d.state.activeWindowId,{x:400,y:210,width:660,height:610})});await compare('resize')
  await page.getByRole('combobox',{name:'主题',exact:true}).selectOption('dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark')
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<6;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}))});await expect(page.locator('.app-window')).toHaveCount(8);await compare('stack')
  const stacked=(await page.locator('.app-window.focused .window-titlebar').boundingBox())!;await page.mouse.move(stacked.x+180,stacked.y+15);await page.mouse.down();await page.mouse.move(stacked.x+225,stacked.y+35);await compare('stack-held');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().open({appId:'blora.files',disposition:'new-window'})})
  await expect(page.locator('[data-static-shadow-underlay]')).toHaveCount(0)
  await page.evaluate(()=>(window as any).__staticShadowUnderlay.dispose());await expect(page.locator('[data-static-shadow-underlay]')).toHaveCount(0)
 })
})
