import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {nativeDeviceScale} from '../../playwright-browser'

async function stable(page:Page){
 const capture=()=>page.screenshot({animations:'disabled',caret:process.env.BLORA_E08_NATIVE_CAPTURE_CARET==='initial'?'initial':'hide'})
 const state=()=>page.locator('.app-window').evaluateAll(es=>es.map(e=>{const f=e as HTMLElement,b=f.querySelector('.window-body') as HTMLElement,s=getComputedStyle(f),r=f.getBoundingClientRect();return {class:f.className,transform:s.transform,left:s.left,top:s.top,box:[r.x,r.y,r.width,r.height],clip:b.style.clipPath,header:f.style.getPropertyValue('--material-header-size'),material:[b.style.getPropertyValue('--material-x'),b.style.getPropertyValue('--material-y')],shadowLow:f.dataset.shadowLow}}))
 const states=[await state()]
 let previous=await capture(),last=previous
 for(let i=0;i<8;i++){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))));const current=await capture();if(current.equals(previous))return current;states.push(await state());last=previous;previous=current}
 const diagnostic=await page.evaluate(async({a,b})=>{
  const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const x=c.getContext('2d')!;x.drawImage(image,0,0);return {pixels:x.getImageData(0,0,c.width,c.height).data,width:c.width}}
  const p=await load(a),q=await load(b);let left=Infinity,top=Infinity,right=0,bottom=0,changed=0,max=0
  for(let i=0;i<p.pixels.length;i++){const d=Math.abs(p.pixels[i]!-q.pixels[i]!);if(!d)continue;changed++;max=Math.max(max,d);const at=Math.floor(i/4),x=at%p.width,y=Math.floor(at/p.width);left=Math.min(left,x);right=Math.max(right,x+1);top=Math.min(top,y);bottom=Math.max(bottom,y+1)}
  const owner=Number.isFinite(left)?document.elementsFromPoint((left+right)/2/devicePixelRatio,(top+bottom)/2/devicePixelRatio).map(e=>({tag:e.tagName,css:e.getAttribute('class'),type:e.getAttribute('type'),role:e.getAttribute('role')})):[]
  return {changed,max,bounds:[left,top,right,bottom],owner,focus:{tag:document.activeElement?.tagName,css:document.activeElement?.getAttribute('class')}}
 },{a:last.toString('base64'),b:previous.toString('base64')})
 await test.info().attach('unstable-native-previous.png',{body:last,contentType:'image/png'});await test.info().attach('unstable-native-current.png',{body:previous,contentType:'image/png'})
 console.log('WINDOW_POSITION_UNSTABLE',JSON.stringify({...diagnostic,layoutChanged:states.some(s=>JSON.stringify(s)!==JSON.stringify(states[0])),states:[states[0],states.at(-1)]}));throw Error('Native window reference did not settle')
}
for(const scale of [1,1.25])test.describe(`production window positioning DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('native software offsets match the original transform pixels and keep geometry, nodes and responsive fallback',async({page,browserName},info)=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-position-product',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  // Exercise the software branch independently of this browser's optional
  // renderer-identification extension. Only the first hardware probe returns
  // no context; native DOM/SVG/artwork paint and subsequent GL calls remain.
  // Real performance runs never install this deterministic feature control.
  await page.addInitScript(()=>{const original=HTMLCanvasElement.prototype.getContext;let first=true;HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,type:string,...args:any[]){if(first&&type==='webgl2'&&args[0]?.failIfMajorPerformanceCaveat){first=false;return null}return (original as any).call(this,type,...args)} as typeof original})
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-08T03:00:00Z'],newTarget)}})})
  await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();for(let i=0;i<7;i++)d.open({appId:'blora.instances',disposition:'new-window'});d.open({appId:'blora.settings',disposition:'new-window'});d.state.order.forEach((id:string,i:number)=>d.geometry(id,{x:132+(i%7)*26,y:100+(i%7)*26,width:960,height:620}));(window as any).__positionProductViews=[...document.querySelectorAll('.app-view')];const style=document.createElement('style');style.textContent='.app-window[data-window-position="offset"]{left:0!important;top:0!important;transform:translate(var(--window-x),var(--window-y))!important}';style.disabled=true;document.head.append(style);(window as any).__originalWindowTransformStyle=style})
  const compare=async(label:string)=>{
   // Optical comparisons use the window itself as their native focus owner;
   // focused controls/auto-focus from opening eight views are covered by the
   // separate immediate-input/restore suites, not this static paint oracle.
   if(process.env.BLORA_E08_STATIC_WINDOW_FOCUS==='blur')await page.evaluate(()=>{if(document.activeElement instanceof HTMLElement)document.activeElement.blur()})
   else await page.locator('.app-window.focused').focus()
   if(await page.locator('html').getAttribute('data-translucency')==='on'){
    await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
    if(browserName==='chromium'){
     await expect(page.locator('.desktop')).toHaveAttribute('data-material-body-paint','native')
     expect(await page.locator('.app-window.focused>.window-body').evaluate(e=>getComputedStyle(e,'::before').willChange)).toBe('auto')
    }
   }
   for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
   const geometry=()=>page.locator('.app-window,.app-view').evaluateAll(es=>es.map(e=>{const r=e.getBoundingClientRect();return [r.x,r.y,r.width,r.height]}))
   await page.evaluate(()=>(window as any).__originalWindowTransformStyle.disabled=false);const baseline=await stable(page),before=await geometry()
   await page.evaluate(()=>(window as any).__originalWindowTransformStyle.disabled=true);const candidate=await stable(page)
   const counts=await page.locator('.app-window').evaluateAll(es=>({frames:es.length,offsets:es.filter(e=>{const s=getComputedStyle(e),m=new DOMMatrixReadOnly(s.transform);return (e as HTMLElement).dataset.windowPosition==='offset'&&m.e===0&&m.f===0}).length}));expect(counts.offsets).toBe(scale===1&&browserName!=='chromium'?8:0)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:baseline.toString('base64'),b:candidate.toString('base64')})
   console.log('PRODUCT_WINDOW_POSITION',label,JSON.stringify({...counts,...diff}));await info.attach(label+'-original.png',{body:baseline,contentType:'image/png'});await info.attach(label+'-offset.png',{body:candidate,contentType:'image/png'});expect(await geometry()).toEqual(before);expect(diff).toEqual({max:0,changed:0});expect(await page.evaluate(()=>(window as any).__positionProductViews.every((e:Element)=>e.isConnected))).toBe(true)
   if(label==='odd-resize')await page.screenshot({path:info.outputPath('odd-resize-native.png'),animations:'disabled'})
  }
  await compare('stack-light');const frame=page.locator('.app-window.focused'),title=(await frame.locator('.window-titlebar').boundingBox())!;await page.mouse.move(title.x+180,title.y+15);await page.mouse.down();await page.mouse.move(title.x+380,title.y+35);await compare('held');await page.mouse.move(title.x+90,title.y+20);await compare('reverse-held');await page.mouse.up()
  await page.evaluate(async()=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.geometry(d.state.activeWindowId,{x:400,y:210,width:661,height:611})});await compare('odd-resize')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark');await expect(page.locator('html')).toHaveAttribute('data-theme','dark');await compare('dark');await page.getByRole('switch',{name:'通透模式'}).uncheck();await compare('solid')
  await page.setViewportSize({width:700,height:700});await expect(frame).toHaveCSS('transform','matrix(1, 0, 0, 1, 8, 74)');const narrow=await frame.boundingBox();expect(narrow?.x).toBe(8);expect(narrow?.y).toBe(74)
  expect(errors).toEqual([])
 })
})

test('independent native position writes reveal covered application content without being mistaken for a drag',async({page})=>{
 await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'position-write',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
 await page.goto('/');await expect(page.locator('.topbar')).toBeVisible();await page.locator('[data-app="blora.instances"]').click()
 const lowerId=await page.locator('.app-window').getAttribute('data-window-id'),lower=page.locator(`[data-window-id="${lowerId}"]`),body=await lower.locator('.window-body').boundingBox()
 await page.evaluate(async body=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),d=useDesktop();d.open({appId:'blora.instances',disposition:'new-window'});d.geometry(d.state.activeWindowId,{x:body!.x-40,y:body!.y-40,width:body!.width+80,height:body!.height+80})},body)
 const upper=page.locator('.app-window.focused');await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true');await expect.poll(()=>lower.locator('.window-body').evaluate(e=>(e as HTMLElement).style.clipPath)).toBe('inset(100%)')
 const saved=await upper.evaluate(e=>{const f=e as HTMLElement;return {transform:f.style.transform,x:f.style.getPropertyValue('--window-x'),offset:f.dataset.windowPosition==='offset'}})
 await upper.evaluate((e,offset)=>{const f=e as HTMLElement;if(offset)f.style.setProperty('--window-x','1250px');else f.style.transform='translate(1250px, 100px)'},saved.offset)
 await expect.poll(()=>lower.locator('.window-body').evaluate(e=>(e as HTMLElement).style.clipPath)).not.toBe('inset(100%)');expect((await upper.boundingBox())?.x).toBe(1250)
 if(saved.offset)expect(await upper.evaluate(e=>(e as HTMLElement).style.transform)).toBe(saved.transform)
 await upper.evaluate((e,saved)=>{const f=e as HTMLElement;if(saved.offset)f.style.setProperty('--window-x',saved.x);else f.style.transform=saved.transform},saved)
 await expect.poll(()=>lower.locator('.window-body').evaluate(e=>(e as HTMLElement).style.clipPath)).toBe('inset(100%)')
})
