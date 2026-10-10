import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {nativeDeviceScale} from '../../playwright-browser'
import {installNativeWindowPosition} from '../helpers/native-window-position'

async function stable(page:Page){let previous=await page.screenshot({animations:'disabled',caret:'hide'});for(let i=0;i<8;i++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const current=await page.screenshot({animations:'disabled',caret:'hide'});if(current.equals(previous))return current;previous=current}throw Error('Native positioning did not settle')}
for(const scale of [1,1.25])test.describe(`native window offsets DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('direct native positions keep exact pixels, application nodes and held geometry',async({page},info)=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-position',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<7;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.open({appId:'blora.settings',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}));(window as any).__positionViews=[...document.querySelectorAll('.app-view')]})
  await installNativeWindowPosition(page,process.env.BLORA_E08_POSITION_UNRETAINED!=='1',process.env.BLORA_E08_POSITION_IDENTITY==='1')
  const compare=async(label:string)=>{
   if(await page.locator('html').getAttribute('data-translucency')==='on')await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   const geometry=()=>page.locator('.app-window,.app-view').evaluateAll(es=>es.map(e=>{const r=e.getBoundingClientRect();return [r.x,r.y,r.width,r.height]}))
   await page.evaluate(()=>(window as any).__nativeWindowPosition.setVisible(false));const baseline=await stable(page),before=await geometry()
   await page.evaluate(()=>(window as any).__nativeWindowPosition.setVisible(true));const candidate=await stable(page),counts=await page.evaluate(()=>(window as any).__nativeWindowPosition.snapshot());expect(counts.enabled).toBe(scale===1?8:0)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:baseline.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_WINDOW_POSITION',label,JSON.stringify({...counts,...diff}));await info.attach(label+'-native.png',{body:baseline,contentType:'image/png'});await info.attach(label+'-offset.png',{body:candidate,contentType:'image/png'})
   expect(await geometry()).toEqual(before);expect(diff).toEqual({max:0,changed:0});expect(await page.evaluate(()=>(window as any).__positionViews.every((e:Element)=>e.isConnected))).toBe(true)
  }
  await compare('stack-light');const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!;await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+380,title.y+35);await compare('held');await page.mouse.move(title.x+90,title.y+20);await compare('reverse-held');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.geometry(d.state.activeWindowId,{x:400,y:210,width:661,height:611})});await compare('odd-resize')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark');await page.getByRole('switch',{name:'通透模式'}).uncheck();await compare('solid');expect(errors).toEqual([])
  await page.evaluate(()=>(window as any).__nativeWindowPosition.dispose())
 })
})
