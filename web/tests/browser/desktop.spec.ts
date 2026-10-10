import {pressEditorKey} from '../helpers/editor-key'
import {appearance as readAppearance} from '../helpers/appearance'
import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
// API fixtures are exclusively a browser test boundary. Production uses fetch against Master.
test.beforeEach(async({page})=>{
  await page.route('**/api/v1/**',route=>{const path=new URL(route.request().url()).pathname;return route.fulfill({json:path.endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}})})
})
test('cached optical surfaces settle after motion and rapid appearance changes',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.locator('.launcher-button').click()
  const launcher=page.locator('.launcher')
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await expect.poll(()=>launcher.evaluate(element=>{
    const box=element.getBoundingClientRect(),style=getComputedStyle(element)
    return Math.abs(parseFloat(style.getPropertyValue('--material-y'))+box.y)
  })).toBeLessThan(1)
  await launcher.getByRole('button',{name:'设置',exact:true}).click()
  const theme=page.getByRole('combobox',{name:'主题',exact:true}),palette=page.getByRole('combobox',{name:'主题配色'}),toggle=page.getByRole('switch',{name:'通透模式'})
  await selectStyledOption(page,theme,'dark');await selectStyledOption(page,palette,'sand');await toggle.uncheck()
  await selectStyledOption(page,theme,'light');await selectStyledOption(page,palette,'ice');await toggle.check()
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await expect(page.locator('html')).toHaveAttribute('data-theme','light')
  const current=await page.locator('.topbar').evaluate(element=>getComputedStyle(element).backgroundImage)
  await page.waitForTimeout(200)
  expect(await page.locator('.topbar').evaluate(element=>getComputedStyle(element).backgroundImage)).toBe(current)
  await toggle.uncheck()
  await expect(page.locator('.topbar')).toHaveCSS('backdrop-filter','none')
  expect(await page.locator('.topbar').evaluate(element=>getComputedStyle(element).backgroundImage)).not.toContain('blob:')
  expect(errors).toEqual([])
})

test('application row updates stay outside host material registration while dynamic menus remain located',async({page})=>{
 await page.addInitScript((forceGlobal:boolean)=>{
  let applicationQueries=0
  const matches=Element.prototype.matches,query=Element.prototype.querySelector,closest=Element.prototype.closest
  const inspect=(element:Element,selector:string)=>{if(selector.includes('.account-popover')&&closest.call(element,'.app-view'))applicationQueries++}
  Element.prototype.matches=function(this:Element,selector:string){inspect(this,selector);return matches.call(this,selector)} as typeof matches
  Element.prototype.querySelector=function(this:Element,selector:string){inspect(this,selector);return query.call(this,selector)} as typeof query
  if(forceGlobal){
   // Explicit negative control reproduces the former whole-desktop observer.
   // The same assertion must fail; it is never enabled in normal regression.
   const observe=MutationObserver.prototype.observe
   let materialObserver:MutationObserver|undefined
   MutationObserver.prototype.observe=function(target:Node,options?:MutationObserverInit){
    if(!materialObserver&&target instanceof Element&&matches.call(target,'.desktop')&&options?.childList)materialObserver=this
    return observe.call(this,target,this===materialObserver&&target instanceof Element&&matches.call(target,'.desktop')?{...options,subtree:true}:options)
   }
  }
  ;(window as any).__hostMaterialQueries=()=>applicationQueries
 },process.env.BLORA_E08_GLOBAL_MATERIAL_CONTROL==='1')
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 const before=await page.evaluate(()=>(window as any).__hostMaterialQueries())
 await page.locator('.app-window .app-view').evaluate(async view=>{
  const output=document.createElement('div');output.style.display='none';view.append(output)
  for(let round=0;round<12;round++){
   output.replaceChildren(...Array.from({length:80},()=>{const span=document.createElement('span');span.textContent='native row update';return span}))
   await Promise.resolve()
  }
  output.remove();await new Promise<void>(resolve=>requestAnimationFrame(()=>resolve()))
 })
 expect(await page.evaluate(()=>(window as any).__hostMaterialQueries())).toBe(before)
 const located=async(selector:string)=>{
  const surface=page.locator(selector);await expect(surface).toBeVisible()
  await expect.poll(()=>surface.evaluate(element=>{
   const rect=element.getBoundingClientRect(),style=getComputedStyle(element)
   // CSS serializes fractional coordinates with fewer decimals than a DOMRect.
   const error=Math.max(Math.abs(parseFloat(style.getPropertyValue('--material-x'))+rect.x),Math.abs(parseFloat(style.getPropertyValue('--material-y'))+rect.y))
   return {image:style.backgroundImage.includes('blob:'),located:error<.01}
  })).toEqual({image:true,located:true})
 }
 await page.getByRole('button',{name:'账号菜单',exact:true}).click();await located('.account-popover')
 await page.getByRole('button',{name:'站内通知',exact:true}).click();await located('.notification-panel')
 await page.locator('.launcher-button').click();await located('.launcher')
 await page.locator('.launcher-button').click();await page.locator('.launcher-button').click();await located('.launcher')
})
test('native shadow edge cache preserves elevation, contour and fallback through resize and theme changes',async({page})=>{
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const frame=page.locator('.app-window.focused'),plane=frame.locator('.window-shadow-plane')
 await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
 const piece=plane.locator('.shadow-piece.nw')
 const normal=await piece.evaluate(element=>getComputedStyle(element).getPropertyValue('--shadow-normal').trim()),focused=await piece.evaluate(element=>getComputedStyle(element).getPropertyValue('--shadow-focused').trim())
 expect(normal).toContain('blob:');expect(focused).toContain('blob:');expect(normal).not.toBe(focused)
 await expect(piece).toHaveCSS('background-image',focused)
 expect(await frame.evaluate(element=>getComputedStyle(element).boxShadow)).toBe('none')
 const contour=()=>frame.evaluate(element=>{const frame=element as HTMLElement,plane=frame.querySelector<HTMLElement>('.window-shadow-plane')!,a=frame.getBoundingClientRect(),b=plane.getBoundingClientRect();return Math.max(Math.abs(a.x-b.x),Math.abs(a.y-b.y),Math.abs(a.width-b.width),Math.abs(a.height-b.height),Math.abs(parseFloat(getComputedStyle(frame).borderTopLeftRadius)-parseFloat(getComputedStyle(plane).borderTopLeftRadius)))})
 expect(await contour()).toBeLessThan(.01)
 const box=(await frame.locator('.window-titlebar').boundingBox())!
 const patchArea=()=>plane.locator('.shadow-piece').evaluateAll(elements=>elements.reduce((sum,element)=>{const box=element.getBoundingClientRect();return sum+box.width*box.height},0))
 const focusedArea=await patchArea()
 await page.mouse.move(box.x+240,box.y+15);await page.mouse.down();await page.mouse.move(box.x+255,box.y+20)
 await expect(piece).toHaveCSS('background-image',normal)
 expect(await patchArea(),'Normal elevation must not draw the focused state’s larger transparent margins').toBeLessThan(focusedArea)
 await page.mouse.up();await expect(piece).toHaveCSS('background-image',focused)
 // A too-small contour must use the original native shadow, never compressed
 // corner patches. The real desktop state and layout are resized here.
 await frame.evaluate(async element=>{const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop(),id=(element as HTMLElement).dataset.windowId!;desktop.geometry(id,{...desktop.state!.windows[id]!.rect,width:360,height:320})})
 // A narrow mobile viewport cannot fit the curved corner influence.
 await page.setViewportSize({width:200,height:900})
 await expect(plane).not.toHaveAttribute('data-shadow-cache',/.+/)
 expect(await plane.evaluate(element=>getComputedStyle(element).boxShadow)).not.toBe('none')
 expect(await contour()).toBeLessThan(.01)
 await page.setViewportSize({width:1440,height:960})
 await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
 await page.locator('.dock-item[aria-label="设置"]').click()
 await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark')
 await selectStyledOption(page,page.getByRole('combobox',{name:'主题配色'}),'rose')
 await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'light')
 await selectStyledOption(page,page.getByRole('combobox',{name:'主题配色'}),'ice')
 await expect(page.locator('.window-shadow-plane').first()).toHaveAttribute('data-shadow-cache',/.+/)
 expect(errors).toEqual([])
})
test('a delayed shadow contour cannot replace a newer decoded contour',async({page})=>{
 await page.addInitScript(()=>{
  const original=Image.prototype.decode
  const gates:Array<()=>void>=[]
  let released=false,completed=0
  ;(window as any).__shadowRace={gates,release:()=>{released=true;for(const release of gates)release()},completed:()=>completed}
  Image.prototype.decode=async function(){
   const old=this.src.startsWith('data:image/svg+xml')&&decodeURIComponent(this.src).includes('border-radius:28px')
   if(old&&!released)await new Promise<void>(resolve=>gates.push(resolve))
   await original.call(this)
   if(released&&this.src.startsWith('blob:'))completed++
  }
 })
 const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const frame=page.locator('.app-window.focused'),plane=frame.locator('.window-shadow-plane')
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 await expect.poll(()=>page.evaluate(()=>(window as any).__shadowRace.gates.length)).toBeGreaterThan(0)
 // Resize just this frame, keeping the root cache generation unchanged.
 await frame.evaluate(element=>{const frame=element as HTMLElement;frame.style.borderRadius='16px';frame.style.width=frame.offsetWidth-1+'px'})
 const radius=()=>plane.evaluate(element=>{const key=(element as HTMLElement).dataset.shadowCache;return key?JSON.parse(key)[2]:null})
 await expect.poll(radius).toBe(16)
 const urls=await plane.locator('.shadow-piece').evaluateAll(elements=>elements.map(element=>(element as HTMLElement).style.getPropertyValue('--shadow-normal')))
 await page.evaluate(()=>(window as any).__shadowRace.release())
 await expect.poll(()=>page.evaluate(()=>(window as any).__shadowRace.completed())).toBeGreaterThanOrEqual(16)
 await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
 expect(await radius()).toBe(16)
 expect(await plane.locator('.shadow-piece').evaluateAll(elements=>elements.map(element=>(element as HTMLElement).style.getPropertyValue('--shadow-normal')))).toEqual(urls)
 expect(errors).toEqual([])
})

test('shadow raster failure retains the native shadow without a page error',async({page})=>{
 await page.addInitScript(()=>{HTMLCanvasElement.prototype.toBlob=function(callback){callback(null)}})
 const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const plane=page.locator('.app-window.focused .window-shadow-plane')
 await expect(plane).toBeAttached();await page.waitForTimeout(200)
 await expect(plane).not.toHaveAttribute('data-shadow-cache',/.+/)
 expect(await plane.evaluate(element=>getComputedStyle(element).boxShadow)).not.toBe('none')
 expect(errors).toEqual([])
})
for(const scale of [1,1.25])test.describe(`exposure strips at DPR ${scale}`,()=>{
 test.use({...nativeDeviceScale(scale)})
 test('bounded native list paint preserves exact pixels on scroll, replacement, held reveal and fallback',async({page},info)=>{
  // Fix only no-argument display dates. Date.now must remain native: freezing
  // it suppresses Vue's capture/bubble event timestamp guard on title drags.
  // Timers, event timestamps, RAF and paint all remain live.
  await page.addInitScript(()=>{window.Date=new Proxy(Date,{construct:(target,args,newTarget)=>Reflect.construct(target,args.length?args:['2026-10-07T12:00:00Z'],newTarget)})})
  await page.route('**/api/v1/instances',route=>route.fulfill({json:{items:Array.from({length:18},(_,i)=>({instanceId:'paint-'+i,nodeId:'paint-node',name:'Native 中文 list '+i,state:'STOPPED',group:'paint',tags:['exact']}))}}))
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  const lower=page.locator('.app-window').first(),id=await lower.getAttribute('data-window-id')
  await expect(lower.locator('.instance-identity')).toHaveCount(18)
  await lower.locator('[aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const upper=page.locator('.app-window.focused'),upperId=await upper.getAttribute('data-window-id')
  await page.mouse.move(1350,800)
  await page.evaluate(async({id,upperId})=>{
   const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop()
   desktop.geometry(id!,{x:180,y:180,width:720,height:520});desktop.geometry(upperId!,{x:250,y:210,width:800,height:620})
  },{id,upperId})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true')
  const listControl=process.env.BLORA_E08_NATIVE_LIST_CONTROL==='1'
  if(listControl){
   await page.evaluate(async()=>{const {installNativeListDiagnostic}=await import('/tests/helpers/native-list.ts' as string);installNativeListDiagnostic()})
   await expect.poll(()=>lower.locator('[data-app-paint-occluded="true"]').count()).toBeGreaterThan(0)
  }else await expect(page.locator('[data-app-paint-occluded="true"]')).toHaveCount(0)
  const compare=async()=>{
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const clipped=await page.screenshot({animations:'disabled'})
   const selector='[data-app-paint-occluded="true"]',previous=await page.locator(selector).evaluateAll(elements=>elements.map(element=>{
    const block=element as HTMLElement,old=block.style.clipPath;block.style.setProperty('clip-path','none','important');return old
   }))
   const complete=await page.screenshot({animations:'disabled'})
   await page.locator(selector).evaluateAll((elements,values)=>elements.forEach((element,index)=>(element as HTMLElement).style.setProperty('clip-path',values[index]!)),previous)
   await info.attach('native-list-clipped.png',{body:clipped,contentType:'image/png'});await info.attach('native-list-complete.png',{body:complete,contentType:'image/png'})
   return page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return context.getImageData(0,0,canvas.width,canvas.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0
    for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta)}
    return {changed,max}
   },{a:clipped.toString('base64'),b:complete.toString('base64')})
  }
  expect(await compare()).toEqual({changed:0,max:0})
  await lower.locator('.resource-main').evaluate(element=>{element.scrollTop=240})
  expect(await compare()).toEqual({changed:0,max:0})
  // Real default-app filtering replaces its native rows while occluded.
  await lower.getByRole('searchbox',{name:'搜索实例'}).evaluate(element=>{const input=element as HTMLInputElement;input.value='list 1';input.dispatchEvent(new Event('input',{bubbles:true}))})
  await expect(lower.locator('.instance-identity')).toHaveCount(9)
  expect(await compare()).toEqual({changed:0,max:0})
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  const hit=await page.evaluate(({x,y})=>{const element=document.elementFromPoint(x,y);return {tag:element?.tagName,class:element?.className,title:element?.closest<HTMLElement>('.window-titlebar')?.closest<HTMLElement>('.app-window')?.dataset.windowId}}, {x:title.x+220,y:title.y+15})
  expect(hit.title,JSON.stringify(hit)).toBe(upperId)
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down()
  await expect(upper).toHaveClass(/moving/)
  await page.mouse.move(1300,title.y+15)
  await expect(upper).toHaveClass(/moving/)
  expect(await compare()).toEqual({changed:0,max:0})
  await page.mouse.up()
  await page.evaluate(async id=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().focus(id!)},id)
  await expect(lower.locator('[data-app-paint-occluded="true"]')).toHaveCount(0)
  expect(await compare()).toEqual({changed:0,max:0})
  await page.evaluate(()=>{delete document.documentElement.dataset.materialCache;delete document.documentElement.dataset.materialCacheOpaque})
  await expect(page.locator('[data-app-paint-occluded="true"]')).toHaveCount(0)
  expect(await compare()).toEqual({changed:0,max:0})
 })
 test('partially covered native shadow patches preserve exact pixels through exposure and fallback',async({page},info)=>{
  // Freeze only the displayed no-argument date. Real Date.now/event timestamps
  // and RAF retain Vue's native capture/bubble protection and input timing.
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct(target,args,newTarget){return Reflect.construct(target,args.length?args:['2026-10-07T12:34:00Z'],newTarget)}})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  const lower=page.locator('.app-window').first(),id=await lower.getAttribute('data-window-id')
  await lower.locator('[aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const upper=page.locator('.app-window.focused'),upperId=await upper.getAttribute('data-window-id')
  await page.mouse.move(1350,800)
  await page.evaluate(async({id,upperId})=>{
   const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop()
   desktop.geometry(id!,{x:180,y:180,width:720,height:520})
   desktop.geometry(upperId!,{x:250,y:210,width:800,height:620})
  },{id,upperId})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true')
  await expect(lower.locator('.window-shadow-plane')).toHaveAttribute('data-shadow-cache',/.+/)
  const nativeLayer=process.env.BLORA_E08_SHADOW_LAYER_CONTROL
  if(nativeLayer==='plane')await page.addStyleTag({content:'.window-shadow-plane{isolation:isolate} '})
  if(nativeLayer==='pieces')await page.addStyleTag({content:'.shadow-piece{will-change:transform} '})
  const nearestControl=['1','all'].includes(process.env.BLORA_E08_SHADOW_NEAREST_CONTROL||'')
  if(nearestControl){
   // Explicit optical-only diagnostic. The complete reference below restores
   // normal native interpolation rather than comparing the candidate to itself.
   const edges=process.env.BLORA_E08_SHADOW_NEAREST_CONTROL==='all'?'':':is(.n,.s,.w,.e)'
   await page.addStyleTag({content:`@media (resolution:1dppx){.window-shadow-plane[data-shadow-cache]>.shadow-piece${edges}{image-rendering:pixelated}}`})
   await expect(lower.locator('.shadow-piece.n')).toHaveCSS('image-rendering',scale===1?'pixelated':'auto')
  }
  const partialControl=process.env.BLORA_E08_PARTIAL_SHADOW_CONTROL==='1'
  if(partialControl)await page.evaluate(async()=>{const {installPartialShadowDiagnostic}=await import('/tests/helpers/partial-shadow.ts' as string);installPartialShadowDiagnostic()})
  const cropControl=process.env.BLORA_E08_SHADOW_CROP_CONTROL==='1'
  if(cropControl){
   await page.evaluate(async()=>{const {installShadowCropDiagnostic}=await import('/tests/helpers/shadow-crop.ts' as string);(window as any).__nativeShadowCropControl=installShadowCropDiagnostic()})
   if(scale===1)await expect.poll(()=>lower.locator('.native-shadow-crop').count()).toBeGreaterThan(0)
   else await expect(lower.locator('.native-shadow-crop')).toHaveCount(0)
  }
  if(process.env.BLORA_E08_SHADOW_SEGMENTS_CONTROL==='1'){
   // Explicit rejected-renderer diagnostic; never enabled by the product.
   await page.evaluate(async()=>{const {installShadowSegmentsDiagnostic}=await import('/tests/helpers/shadow-segments.ts' as string);installShadowSegmentsDiagnostic()})
   if(scale===1)await expect.poll(()=>lower.locator('.shadow-segment').count()).toBeGreaterThan(0)
   else await expect(lower.locator('.shadow-segment')).toHaveCount(0)
  }
  await expect.poll(()=>lower.locator('.shadow-piece').evaluateAll(elements=>elements.some(element=>{
   const style=(element as HTMLElement).style;return style.clip.startsWith('rect(')||(style.clipPath.startsWith('inset(')&&style.clipPath!=='inset(100%)')
  }))).toBe(partialControl)
  await expect.poll(()=>lower.locator('.shadow-piece').evaluateAll(elements=>elements.some(element=>(element as HTMLElement).style.clip.startsWith('rect(')))).toBe(partialControl&&scale===1)
  const compare=async()=>{
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const clipped=await page.screenshot({animations:'disabled'})
   const previous=await page.locator('.shadow-piece').evaluateAll(elements=>elements.map(element=>{
    const piece=element as HTMLElement,old={path:piece.style.clipPath,legacy:piece.style.clip,visibility:piece.style.visibility};piece.style.setProperty('clip-path','none','important');piece.style.setProperty('clip','auto','important');piece.style.setProperty('visibility','visible','important');return old
   }))
   const nativeEdges=await page.addStyleTag({content:'.window-shadow-plane[data-shadow-cache]>.shadow-piece[data-shadow-segmented="true"]{background-image:var(--shadow-normal)!important}.app-window.focused>.window-shadow-plane[data-shadow-cache]>.shadow-piece[data-shadow-segmented="true"]{background-image:var(--shadow-focused)!important}.app-window[data-shadow-low="true"]>.window-shadow-plane[data-shadow-cache]>.shadow-piece[data-shadow-segmented="true"]{background-image:var(--shadow-normal)!important}.shadow-segment{display:none!important}'+(nearestControl?'.shadow-piece{image-rendering:auto!important}':'')})
   if(cropControl)await page.evaluate(()=>(window as any).__nativeShadowCropControl.pause())
   const complete=await page.screenshot({animations:'disabled'})
   await nativeEdges.evaluate(element=>(element as HTMLElement).remove())
   await page.locator('.shadow-piece').evaluateAll((elements,values)=>elements.forEach((element,index)=>{const style=(element as HTMLElement).style;style.setProperty('clip-path',values[index]!.path);style.setProperty('clip',values[index]!.legacy);style.setProperty('visibility',values[index]!.visibility)}),previous)
   if(cropControl)await page.evaluate(()=>(window as any).__nativeShadowCropControl.resume())
   await info.attach('native-clipped-shadow.png',{body:clipped,contentType:'image/png'});await info.attach('native-complete-shadow.png',{body:complete,contentType:'image/png'})
   return page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const context=canvas.getContext('2d')!;context.drawImage(image,0,0);return context.getImageData(0,0,canvas.width,canvas.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0
    for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta)}
    return {changed,max}
   },{a:clipped.toString('base64'),b:complete.toString('base64')})
  }
  expect(await compare()).toEqual({changed:0,max:0})
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(1300,title.y+15)
  await expect(upper).toHaveClass(/moving/)
  await expect.poll(()=>page.locator('.shadow-piece').evaluateAll(elements=>elements.some(element=>{const style=(element as HTMLElement).style;return !!style.clip||(!!style.clipPath&&style.clipPath!=='inset(100%)')}))).toBe(false)
  expect(await compare()).toEqual({changed:0,max:0})
  await page.mouse.up()
  await page.evaluate(async id=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().geometry(id!,{x:250,y:210,width:800,height:620})},upperId)
  await expect.poll(()=>lower.locator('.shadow-piece').evaluateAll(elements=>elements.some(element=>{const style=(element as HTMLElement).style;return style.clip.startsWith('rect(')||(style.clipPath.startsWith('inset(')&&style.clipPath!=='inset(100%)')}))).toBe(partialControl)
  expect(await compare()).toEqual({changed:0,max:0})
  await page.evaluate(async id=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().focus(id!)},id)
  await expect.poll(()=>lower.locator('.shadow-piece').evaluateAll(elements=>elements.every(element=>!(element as HTMLElement).style.clipPath&&!(element as HTMLElement).style.clip))).toBe(true)
  expect(await compare()).toEqual({changed:0,max:0})
  await lower.evaluate(element=>{const frame=element as HTMLElement;frame.style.borderRadius='120px';frame.style.width='200px'})
  await expect(lower.locator('.window-shadow-plane')).not.toHaveAttribute('data-shadow-cache',/.+/)
  await expect.poll(()=>lower.locator('.shadow-piece').evaluateAll(elements=>elements.every(element=>!(element as HTMLElement).style.clipPath&&!(element as HTMLElement).style.clip))).toBe(true)
  expect(await compare()).toEqual({changed:0,max:0})
 })
 test('separate visible strips preserve exact application pixels through held reveal and focus',async({page},info)=>{
  // Compare equal display minutes while native Date.now, event timestamps,
  // animation frames and gesture timing remain live (as in the list guard).
  await page.addInitScript(()=>{const NativeDate=Date;window.Date=new Proxy(NativeDate,{construct:(target,args,newTarget)=>Reflect.construct(target,args.length?args:['2026-10-08T07:00:00Z'],newTarget)})})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  await page.addStyleTag({content:'*,*::before,*::after{transition:none!important;animation:none!important}'})
  const lower=page.locator('.app-window').first(),id=await lower.getAttribute('data-window-id')
  await lower.getByRole('searchbox',{name:'搜索实例'}).fill('Chinese 中文 exposure strips')
  await lower.locator('[aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const upper=page.locator('.app-window.focused'),upperId=await upper.getAttribute('data-window-id')
  await page.mouse.move(1350,800)
  await page.evaluate(async({id,upperId})=>{
   const {useDesktop}=await import('/src/desktop/store.ts' as string),desktop=useDesktop()
   desktop.geometry(id!,{x:180,y:180,width:720,height:520})
   desktop.geometry(upperId!,{x:270,y:230,width:540,height:420})
  },{id,upperId})
  await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true')
  // Complete both native optical states before comparing the live screen.
  // Switching from the initial CSS shadow to its decoded patches between the
  // two captures would compare different material states, not body clipping.
  for(const plane of await page.locator('.window-shadow-plane').all())await expect(plane).toHaveAttribute('data-shadow-cache',/.+/)
  const exposedHits=()=>page.evaluate(()=>[[200,400],[870,400],[450,680],[450,400]].map(([x,y])=>document.elementFromPoint(x!,y!)?.closest<HTMLElement>('.app-window')?.dataset.windowId))
  await expect.poll(exposedHits).toEqual([id,id,id,upperId])
  const compare=async()=>{
   await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
   const masked=await page.screenshot({animations:'disabled'})
   // Compare the entire real screen against the original unmasked native body.
   const previous=await page.locator('.window-body').evaluateAll(elements=>elements.map(element=>{
    const body=element as HTMLElement,old=body.style.clipPath
    body.style.setProperty('clip-path','none','important');return old
   }))
   const all=await page.screenshot({animations:'disabled'})
   await page.locator('.window-body').evaluateAll((elements,values)=>elements.forEach((element,index)=>(element as HTMLElement).style.setProperty('clip-path',values[index]!)),previous)
   await info.attach('exposed-strips.png',{body:masked,contentType:'image/png'});await info.attach('complete-body.png',{body:all,contentType:'image/png'})
   return page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0
    for(let i=0;i<x.length;i++){const delta=Math.abs(x[i]!-y[i]!);if(delta)changed++;max=Math.max(max,delta)}
    return {changed,max}
   },{a:masked.toString('base64'),b:all.toString('base64')})
  }
  expect(await compare()).toEqual({changed:0,max:0})
  const title=(await upper.locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(1300,title.y+15)
  await expect(upper).toHaveClass(/moving/)
  await expect(lower.locator('.window-body')).toHaveCSS('clip-path','none')
  expect(await compare()).toEqual({changed:0,max:0})
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.up()
  await expect.poll(exposedHits).toEqual([id,id,id,upperId])
  await page.evaluate(async id=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().focus(id!)},id)
  await expect(lower.locator('.window-body')).toHaveCSS('clip-path','none')
  await expect(lower.getByRole('searchbox',{name:'搜索实例'})).toHaveValue('Chinese 中文 exposure strips')
 })
})
test('occluded application content is fully exposed again on window activation',async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const original=await page.locator('.app-window').getAttribute('data-window-id')
 for(let i=0;i<2;i++){
  await page.locator('.app-window.focused [aria-label^="标签菜单"]').click()
  await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
 }
 await expect(page.locator('.app-window')).toHaveCount(3)
 await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true')
 const lower=page.locator(`[data-window-id="${original}"]`),body=lower.locator('.window-body')
 await expect.poll(()=>body.evaluate(element=>(element as HTMLElement).style.clipPath)).not.toBe('')
 // Moving an opaque cover must reveal the lower content before the next
 // paint, not leave yesterday's clipping in place for an extra frame.
 const exposedOnNextPaint=await page.evaluate(id=>new Promise<string>(resolve=>{
  const frames=[...document.querySelectorAll<HTMLElement>('.app-window')].filter(frame=>frame.dataset.windowId!==id)
  const transforms=frames.map(frame=>frame.style.transform)
  requestAnimationFrame(()=>{
   frames.forEach(frame=>frame.style.transform='translate(2000px,2000px)')
   requestAnimationFrame(()=>{
    const clip=document.querySelector<HTMLElement>(`[data-window-id="${id}"] .window-body`)!.style.clipPath
    frames.forEach((frame,index)=>frame.style.transform=transforms[index]!)
    resolve(clip)
   })
  })
 }),original)
 expect(exposedOnNextPaint).toBe('')
 await page.locator('.dock-item[aria-label="实例中心"]').click()
 await page.locator('.window-picker button').first().click()
 await expect(lower).toHaveClass(/focused/)
 await expect.poll(()=>body.evaluate(element=>(element as HTMLElement).style.clipPath)).toBe('')
 await lower.getByRole('searchbox',{name:'搜索实例'}).fill('遮挡恢复后的筛选')
 await page.reload()
 await expect(lower.getByRole('searchbox',{name:'搜索实例'})).toHaveValue('遮挡恢复后的筛选')
 await expect.poll(()=>body.evaluate(element=>(element as HTMLElement).style.clipPath)).toBe('')
})
test('a large held native drag reveals lower application content by the next paint',async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const id=await page.locator('.app-window').getAttribute('data-window-id')
 const lower=page.locator(`[data-window-id="${id}"]`),search=lower.getByRole('searchbox',{name:'搜索实例'})
 await search.fill('大幅拖动露出后保留正文')
 await lower.locator('[aria-label^="标签菜单"]').click()
 await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
 const upper=page.locator('.app-window.focused'),title=upper.locator('.window-titlebar')
 // Enclose the complete lower body: a diagonal pair of equal windows leaves
 // exposed corner strips, whose conservative bounding box is intentionally full.
 const bodyBox=await lower.locator('.window-body').boundingBox()
 await upper.evaluate(async(element,box)=>{
  const {useDesktop}=await import('/src/desktop/store.ts' as string)
  useDesktop().geometry((element as HTMLElement).dataset.windowId!,{x:box!.x-40,y:box!.y-40,width:box!.width+80,height:box!.height+80})
 },bodyBox)
 await expect(page.locator('html')).toHaveAttribute('data-material-cache-opaque','true')
 await expect.poll(()=>lower.locator('.window-body').evaluate(element=>(element as HTMLElement).style.clipPath)).not.toBe('')
 const box=await title.boundingBox(),point=await search.boundingBox()
 expect(await page.evaluate(p=>document.elementFromPoint(p.x+p.width/2,p.y+p.height/2)?.closest<HTMLElement>('.app-window')?.dataset.windowId,point!)).not.toBe(id)
 await page.mouse.move(box!.x+180,box!.y+16);await page.mouse.down()
 await page.mouse.move(1250,box!.y+16)
 const hit=await page.evaluate(p=>new Promise<string|undefined>(resolve=>requestAnimationFrame(()=>resolve(document.elementFromPoint(p.x+p.width/2,p.y+p.height/2)?.closest<HTMLElement>('.app-window')?.dataset.windowId))),point!)
 expect(hit).toBe(id)
 await expect(upper).toHaveClass(/moving/)
 await expect(search).toHaveValue('大幅拖动露出后保留正文')
 await page.mouse.up()
 await search.fill('露出后仍可输入');await page.reload()
 await expect(search).toHaveValue('露出后仍可输入')
})
for(const phase of ['before','after'] as const)test(`native drag ownership preserves unowned contour and layout writes ${phase} its transform`,async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const win=page.locator('.app-window'),title=(await win.locator('.window-titlebar').boundingBox())!
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 await win.evaluate((element,phase)=>{
  const frame=element as HTMLElement
  const change=(event:PointerEvent)=>{
   if(event.buttons!==1)return
   frame.style.borderRadius='48px'
   const notice=document.createElement('div');notice.style.cssText='height:21px;flex:none';notice.dataset.motionLayoutProbe='true'
   frame.insertBefore(notice,frame.querySelector(':scope>.window-body'))
   frame.removeEventListener('pointermove',change,phase==='before')
  }
  frame.addEventListener('pointermove',change,phase==='before')
 },phase)
 await page.mouse.move(title.x+240,title.y+15);await page.mouse.down();await page.mouse.move(title.x+280,title.y+25)
 await expect(win).toHaveClass(/moving/)
 await expect.poll(()=>win.evaluate(element=>{
  const frame=element as HTMLElement,box=frame.getBoundingClientRect(),body=frame.querySelector('.window-body')!.getBoundingClientRect()
  return {radius:frame.style.getPropertyValue('--material-frame-radius'),headerError:Math.abs(parseFloat(frame.style.getPropertyValue('--material-header-size'))-(body.y-box.y+1)),probe:frame.querySelectorAll('[data-motion-layout-probe]').length}
 })).toEqual({radius:'48px',headerError:0,probe:1})
 await expect.poll(()=>win.evaluate(element=>parseFloat(getComputedStyle(element,'::before').borderTopLeftRadius))).toBe(47)
 await page.mouse.up();await expect(win).not.toHaveClass(/moving/)
})
test('pure window translations reuse layout while resize and narrow-screen contours remain accurate',async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const win=page.locator('.app-window')
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 // Let the initial ResizeObserver registration settle, then count actual
 // layout measurements while translating the same native element each frame.
 await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
 const reads=await win.evaluate(async element=>{
  const frame=element as HTMLElement,original=frame.style.transform,method=Element.prototype.getBoundingClientRect
  let reads=0
  Element.prototype.getBoundingClientRect=function(){if(this.matches('.app-window,.window-body'))reads++;return method.call(this)}
  try{
   for(let step=0;step<40;step++){
    frame.style.transform=`translate(${100+step}px,${100+step}px)`
    await new Promise<void>(resolve=>requestAnimationFrame(()=>resolve()))
   }
   const translationReads=reads,position=method.call(frame)
   // The original authored transform must still move native geometry, even
   // when a renderer uses absolute offsets internally. Do not count this
   // independent assertion's measurement as occlusion layout work.
   if(Math.abs(position.x-139)>.01||Math.abs(position.y-139)>.01)throw Error('Authored native translation did not move window geometry')
   return translationReads
  }finally{Element.prototype.getBoundingClientRect=method;frame.style.transform=original}
 })
 expect(reads).toBeLessThan(5)
 await page.setViewportSize({width:600,height:900})
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 await expect.poll(()=>win.evaluate(element=>{
  const style=getComputedStyle(element),header=getComputedStyle(element,'::before')
  return Math.abs(parseFloat(style.borderTopLeftRadius)-1-parseFloat(header.borderTopLeftRadius))
 })).toBeLessThan(.01)
 await expect(win.locator('.window-body')).toHaveCSS('border-bottom-left-radius','15px')
 await page.setViewportSize({width:1440,height:960})
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 await expect(win.getByRole('searchbox',{name:'搜索实例'})).toBeVisible()
})
test('Monaco text, undo/redo and original view identity survive immediate reload and moving windows',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click();await expect(page.locator('.monaco-editor')).toBeVisible()
  // Monaco uses native EditContext in Chromium. Its hidden readonly IME textarea
  // is not the editor input; focus the accessible editing surface instead.
  const editor=page.getByRole('textbox',{name:'文件正文编辑器'});await editor.focus();await page.keyboard.insertText('中文输入\nalpha');await page.keyboard.insertText('β');await pressEditorKey(page,'z')
  const original=await page.locator('[data-view-tab]').first().getAttribute('data-view-tab')
  await page.reload()
  await expect(page.locator('.monaco-editor')).toBeVisible();await expect(page.locator('.monaco-editor .view-lines')).toContainText('alpha');await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('β')
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'y');await expect(page.locator('.monaco-editor .view-lines')).toContainText('β')
  await page.locator('[aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'移到新窗口',exact:true}).click();await page.reload()
  await expect(page.locator(`[data-view-tab="${original}"]`)).toHaveCount(1);await expect(page.locator('.monaco-editor .view-lines')).toContainText('β')
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'z');await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('β');expect(errors).toEqual([])
})
test('window geometry and independent form drafts survive immediate reload',async({page})=>{
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();const win=page.locator('.app-window').first(),title=win.locator('.window-titlebar');const rect=await title.boundingBox();await page.mouse.move(rect!.x+240,rect!.y+15);await page.mouse.down();await page.mouse.move(rect!.x+370,rect!.y+60,{steps:5});await page.mouse.up();const transform=await win.evaluate(el=>(el as HTMLElement).style.transform);await win.getByRole('searchbox',{name:'搜索实例'}).fill('未提交筛选');await page.reload();await expect(page.locator('.app-window')).toHaveCSS('transform',/matrix/);expect(await page.locator('.app-window').evaluate(el=>(el as HTMLElement).style.transform)).toBe(transform);await expect(page.getByRole('searchbox',{name:'搜索实例'})).toHaveValue('未提交筛选')
})

test('dragging retains full material and survives reload while Escape restores geometry',async({page})=>{
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  const win=page.locator('.app-window'),title=win.locator('.window-titlebar')
  const geometry=()=>win.evaluate(element=>(element as HTMLElement).style.transform)
  const materialAlignment=()=>win.evaluate(element=>{
    const frame=element as HTMLElement,body=frame.querySelector<HTMLElement>('.window-body')!,outer=frame.getBoundingClientRect(),inner=body.getBoundingClientRect()
    return Math.max(Math.abs(parseFloat(frame.style.getPropertyValue('--material-x'))-(inner.x-outer.x)-parseFloat(body.style.getPropertyValue('--material-x'))),Math.abs(parseFloat(frame.style.getPropertyValue('--material-y'))-(inner.y-outer.y)-parseFloat(body.style.getPropertyValue('--material-y'))))
  })
  await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  const opticalSurface=()=>win.evaluate(element=>({image:getComputedStyle(element.querySelector('.window-body')!,'::before').backgroundImage,header:getComputedStyle(element,'::before').backgroundImage,blur:getComputedStyle(document.documentElement).getPropertyValue('--material-blur').trim()}))
  const material=await opticalSurface()
  expect(material.image).toContain('blob:');expect(material.header).toContain('blob:');expect(material.blur).toBe('24px')
  const begin=async()=>{
    const box=await title.boundingBox()
    await page.mouse.move(box!.x+240,box!.y+15);await page.mouse.down()
    await page.mouse.move(box!.x+300,box!.y+45,{steps:3})
    await expect(win).toHaveClass(/moving/)
  }
  const original=await geometry()
  await begin();await expect.poll(geometry).not.toBe(original)
  expect(await opticalSurface()).toEqual(material)
  const moved=await geometry()
  // A menu/title/focus update may rerender this frame while geometry is still
  // owned by the held gesture. It must not snap back to the persisted rect.
  await win.locator('[aria-label^="标签菜单"]').evaluate(element=>(element as HTMLElement).click())
  await expect(win.locator('.tab-menu')).toBeVisible()
  await expect.poll(geometry).toBe(moved)
  await page.reload();await page.mouse.up()
  await expect.poll(geometry).toBe(moved)
  await begin();await expect.poll(geometry).not.toBe(moved)
  await page.keyboard.press('Escape');await page.mouse.up()
  await expect.poll(geometry).toBe(moved)
  await page.reload();await expect.poll(geometry).toBe(moved)
  await begin();await page.mouse.up()
  await expect.poll(materialAlignment).toBeLessThan(.01)
})

test('native capture loss releases input protection and protects the latest geometry',async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const win=page.locator('.app-window'),title=win.locator('.window-titlebar')
 await title.evaluate(element=>element.addEventListener('pointerdown',e=>(element as HTMLElement).dataset.testPointerId=String((e as PointerEvent).pointerId),{once:true}))
 const box=(await title.boundingBox())!,before=await win.evaluate(element=>(element as HTMLElement).style.transform)
 await page.mouse.move(box.x+240,box.y+15);await page.mouse.down();await page.mouse.move(box.x+280,box.y+25)
 await expect.poll(()=>win.evaluate(element=>(element as HTMLElement).style.transform)).not.toBe(before)
 const latest=await win.evaluate(element=>(element as HTMLElement).style.transform)
 await title.evaluate(element=>{const target=element as HTMLElement;target.releasePointerCapture(Number(target.dataset.testPointerId))})
 await page.mouse.move(box.x+290,box.y+25)
 await expect(win).not.toHaveClass(/moving/);await expect(win.locator('.window-body')).not.toHaveCSS('pointer-events','none')
 await page.mouse.up()
 expect(await win.evaluate(element=>(element as HTMLElement).style.transform)).toBe(latest)
 await win.getByRole('searchbox',{name:'搜索实例'}).fill('捕获丢失后可输入')
 await page.reload()
 await expect.poll(()=>win.evaluate(element=>(element as HTMLElement).style.transform)).toBe(latest)
 await expect(win.getByRole('searchbox',{name:'搜索实例'})).toHaveValue('捕获丢失后可输入')
})

test('reload commits the latest held drag even before its first animation frame',async({page})=>{
 await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
 const win=page.locator('.app-window'),title=win.locator('.window-titlebar')
 await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
 const before=(await win.boundingBox())!,box=(await title.boundingBox())!,original=await win.evaluate(element=>(element as HTMLElement).style.transform)
 await page.mouse.move(box.x+240,box.y+15);await page.mouse.down()
 await expect(win).toHaveClass(/moving/)
 // Stall painting, not pointer delivery or the synchronous unload commit.
 await page.evaluate(()=>{let id=100000;window.requestAnimationFrame=()=>++id})
 await page.mouse.move(box.x+340,box.y+25)
 // Translation feedback remains immediate even when frame callbacks stall.
 expect(await win.evaluate(element=>(element as HTMLElement).style.transform.replace(/\s+/g,''))).toBe(`translate(${Math.round(before.x+100)}px,${Math.round(before.y+10)}px)`)
 await page.reload();await page.mouse.up()
 await expect.poll(()=>win.evaluate(element=>(element as HTMLElement).style.transform.replace(/\s+/g,''))).toBe(`translate(${Math.round(before.x+100)}px,${Math.round(before.y+10)}px)`)
})

test('offscreen windows retain opaque material while dragging back, resizing and reloading',async({page})=>{
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  const root=page.locator('html'),win=page.locator('.app-window'),title=win.locator('.window-titlebar')
  await expect(root).toHaveAttribute('data-material-cache-opaque','true')
  const west=await win.locator('[data-resize="w"]').boundingBox()
  await page.mouse.move(west!.x+west!.width/2,west!.y+west!.height/2);await page.mouse.down()
  await page.mouse.move(-450,west!.y+west!.height/2);await page.mouse.up()
  expect((await win.boundingBox())!.x).toBeLessThan(0)
  const box=await title.boundingBox(),windowBox=(await win.boundingBox())!,x=Math.min(1300,box!.x+box!.width-170)
  await page.mouse.move(x,box!.y+15);await page.mouse.down();await page.mouse.move(x+100,box!.y+25)
  await expect(win).toHaveClass(/moving/)
  // Native pointer dispatch can return before a scheduled draw in WebKit.
  // Compare the actually requested/painted geometry, never a stale transform.
  await expect.poll(()=>win.evaluate(element=>(element as HTMLElement).style.transform.replace(/\s+/g,''))).toBe(`translate(${Math.round(windowBox.x+100)}px,${Math.round(windowBox.y+10)}px)`)
  const moving=await win.evaluate(element=>(element as HTMLElement).style.transform)
  const planes=await win.evaluate(element=>{
    const header=getComputedStyle(element,'::before'),body=getComputedStyle(element.querySelector('.window-body')!,'::before')
    return [header,body].map(style=>({colour:style.backgroundColor,image:style.backgroundImage}))
  })
  for(const plane of planes){expect(plane.colour).not.toMatch(/(?:transparent|rgba\([^)]*,\s*0\))/);expect(plane.image).toContain('blob:')}
  await page.reload();await page.mouse.up()
  await expect(root).toHaveAttribute('data-material-cache','ready')
  expect(await win.evaluate(element=>(element as HTMLElement).style.transform)).toBe(moving)
  const east=await win.locator('[data-resize="e"]').boundingBox(),width=(await win.boundingBox())!.width
  await page.mouse.move(east!.x+east!.width/2,east!.y+east!.height/2);await page.mouse.down()
  await page.mouse.move(east!.x+east!.width/2+50,east!.y+east!.height/2);await page.mouse.up()
  expect((await win.boundingBox())!.width).toBeGreaterThan(width)
  await win.getByRole('searchbox',{name:'搜索实例'}).fill('离屏恢复仍可操作')
  await page.reload();await expect(win.getByRole('searchbox',{name:'搜索实例'})).toHaveValue('离屏恢复仍可操作')
})

test('theme palette and translucency remain independent across all four combinations and reload',async({page})=>{
  await page.goto('/')
  const root=page.locator('html')
  await expect(root).toHaveAttribute('data-translucency','on')
  await expect(root).toHaveAttribute('data-palette','ice')
  const appIcon=page.locator('.desktop-icon .application-art').first()
  await expect(appIcon).toHaveAttribute('data-icon-style','h04-spectrum-silver')
  const artwork=await appIcon.locator('image').getAttribute('href')
  await page.locator('.dock-item[aria-label="设置"]').click()
  const toggle=page.getByRole('switch',{name:'通透模式'})
  const palette=page.getByRole('combobox',{name:'主题配色'})
  const appearance=()=>readAppearance(page)
  await expect(toggle).toBeChecked()
  await expect(palette).toHaveAttribute('data-value','ice')
  const blueGlass=await appearance()
  expect(blueGlass.wallpaper).toContain('data:image/svg+xml')
  expect(blueGlass.material).toContain('blur(')
  await toggle.uncheck()
  await expect(root).toHaveAttribute('data-translucency','off')
  await expect(root).toHaveAttribute('data-palette','ice')
  const blueSolid=await appearance()
  expect(blueSolid.primary).toBe(blueGlass.primary)
  expect(blueSolid.wallpaper).toBe(blueGlass.wallpaper)
  expect(blueSolid.material).toBe('none')
  await expect(appIcon).toHaveAttribute('data-icon-style','h04-spectrum-silver')
  await expect(appIcon.locator('image')).toHaveAttribute('href',artwork!)
  await page.reload()
  await expect(toggle).not.toBeChecked()
  await expect(palette).toHaveAttribute('data-value','ice')
  await expect(root).toHaveAttribute('data-translucency','off')
  expect((await appearance()).wallpaper).toBe(blueGlass.wallpaper)
  await selectStyledOption(page,palette,'mineral')
  await expect(root).toHaveAttribute('data-palette','mineral')
  await expect(root).toHaveAttribute('data-translucency','off')
  const mineralSolid=await appearance()
  expect(mineralSolid.primary).not.toBe(blueGlass.primary)
  expect(mineralSolid.wallpaper).toContain('/wallpaper.svg')
  expect(mineralSolid.material).toBe('none')
  await toggle.check()
  await expect(root).toHaveAttribute('data-translucency','on')
  await expect(root).toHaveAttribute('data-palette','mineral')
  const mineralGlass=await appearance()
  expect(mineralGlass.primary).toBe(mineralSolid.primary)
  expect(mineralGlass.wallpaper).toBe(mineralSolid.wallpaper)
  expect(mineralGlass.material).toContain('blur(')
  await page.reload()
  await expect(toggle).toBeChecked()
  await expect(palette).toHaveAttribute('data-value','mineral')
  await expect(root).toHaveAttribute('data-palette','mineral')
  await expect(root).toHaveAttribute('data-translucency','on')
  await expect(appIcon).toHaveAttribute('data-icon-style','h04-spectrum-silver')
  await expect(appIcon.locator('image')).toHaveAttribute('href',artwork!)
  expect((await appearance()).wallpaper).toBe(mineralSolid.wallpaper)
})

test('additional colour themes keep their own colours and wallpapers when material changes',async({page})=>{
  await page.goto('/')
  await page.locator('.dock-item[aria-label="设置"]').click()
  const root=page.locator('html')
  const palette=page.getByRole('combobox',{name:'主题配色'})
  const toggle=page.getByRole('switch',{name:'通透模式'})
  const icon=page.locator('.desktop-icon .application-art image').first()
  const artwork=await icon.getAttribute('href')
  const colours=new Set<string>(),wallpapers=new Set<string>()
  const appearance=()=>readAppearance(page)
  for(const id of ['sand','sage','iris','rose']){
    await selectStyledOption(page,palette,id)
    await expect(root).toHaveAttribute('data-palette',id)
    await expect(toggle).toBeChecked()
    const glass=await appearance()
    expect(glass.material).toContain('blur(')
    colours.add(glass.primary);wallpapers.add(glass.wallpaper)
    await toggle.uncheck()
    await expect(root).toHaveAttribute('data-translucency','off')
    const solid=await appearance()
    expect(solid.primary).toBe(glass.primary)
    expect(solid.wallpaper).toBe(glass.wallpaper)
    expect(solid.material).toBe('none')
    await expect(icon).toHaveAttribute('href',artwork!)
    await toggle.check()
  }
  expect(colours.size).toBe(4)
  expect(wallpapers.size).toBe(4)
  await page.reload()
  await expect(palette).toHaveAttribute('data-value','rose')
  await expect(toggle).toBeChecked()
  await expect(root).toHaveAttribute('data-palette','rose')
})

test('dark theme updates app artwork and an open editor while palette and material stay independent',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/')
  const root=page.locator('html'),icon=page.locator('.desktop-icon .application-art image').first()
  const lightArtwork=await icon.getAttribute('href')
  await page.locator('[data-app="blora.editor"]').click()
  await expect(page.locator('.monaco-editor.vs')).toBeVisible()
  await page.locator('.dock-item[aria-label="设置"]').click()
  const settings=page.locator('.settings-panel')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark')
  await expect(root).toHaveAttribute('data-theme','dark')
  await expect(page.locator('.monaco-editor.vs-dark')).toHaveCount(1)
  const darkArtwork=await icon.getAttribute('href')
  expect(darkArtwork).not.toBe(lightArtwork)
  expect(darkArtwork).toContain('h04-spectrum-night-clear-forms')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题配色'}),'rose')
  await page.getByRole('switch',{name:'通透模式'}).uncheck()
  await expect(root).toHaveAttribute('data-palette','rose')
  await expect(root).toHaveAttribute('data-translucency','off')
  await expect(icon).toHaveAttribute('href',darkArtwork!)
  await page.reload()
  await expect(root).toHaveAttribute('data-theme','dark')
  await expect(root).toHaveAttribute('data-palette','rose')
  await expect(icon).toHaveAttribute('href',darkArtwork!)
  await expect(page.locator('.monaco-editor.vs-dark')).toHaveCount(1)
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'light')
  await expect(root).toHaveAttribute('data-theme','light')
  await expect(icon).toHaveAttribute('href',lightArtwork!)
  await expect(page.locator('.monaco-editor.vs')).toHaveCount(1)
  expect(errors).toEqual([])
})

test('dark palette wallpapers keep their shapes and custom wallpaper overrides mineral',async({page})=>{
  await page.goto('/')
  await page.locator('.dock-item[aria-label="设置"]').click()
  const root=page.locator('html')
  await selectStyledOption(page,page.getByRole('combobox',{name:'主题',exact:true}),'dark')
  const palette=page.getByRole('combobox',{name:'主题配色'})
  const wallpaper=page.getByRole('combobox',{name:'壁纸风格'})
  const material=page.getByRole('switch',{name:'通透模式'})
  const background=async()=>(await readAppearance(page)).wallpaper
  const darkWallpapers=new Set<string>()
  for(const value of ['ice','mineral','sand','sage','iris','rose']){
    await selectStyledOption(page,palette,value)
    const artwork=await background()
    expect(artwork).not.toContain('linear-gradient(')
    darkWallpapers.add(artwork)
  }
  expect(darkWallpapers.size).toBe(6)
  await selectStyledOption(page,palette,'sand')
  const sand=await background()
  await material.uncheck()
  expect(await background()).toBe(sand)
  await material.check()
  expect(await background()).toBe(sand)
  await selectStyledOption(page,palette,'mineral')
  const mineral=await background()
  await selectStyledOption(page,wallpaper,'photo')
  await expect(root).toHaveAttribute('data-wallpaper-choice','photo')
  expect(await background()).not.toBe(mineral)
  await page.reload()
  await expect(root).toHaveAttribute('data-theme','dark')
  await expect(root).toHaveAttribute('data-palette','mineral')
  expect(await background()).not.toBe(mineral)
  await expect(page.locator('.desktop-icon .application-art image').first()).toHaveAttribute('href',/h04-spectrum-night-clear-forms/)
})

test('wallpaper styles are distinct and stay independent of palette, material and application icons',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/')
  await page.locator('.dock-item[aria-label="设置"]').click()
  const root=page.locator('html')
  const wallpaper=page.getByRole('combobox',{name:'壁纸风格'})
  const palette=page.getByRole('combobox',{name:'主题配色'})
  const material=page.getByRole('switch',{name:'通透模式'})
  const icon=page.locator('.desktop-icon .application-art image').first()
  const artwork=await icon.getAttribute('href')
  const appearance=()=>readAppearance(page)
  await expect(wallpaper).toHaveAttribute('data-value','palette')
  const original=await appearance(),backgrounds=new Set<string>()
  for(const id of ['photo','poster','nocturne','ink','material','atmosphere','planes','topography','b-grid','b-monolith','b-horizon']){
    await selectStyledOption(page,wallpaper,id)
    await expect(root).toHaveAttribute('data-wallpaper-choice',id)
    const current=await appearance()
    expect(current.primary).toBe(original.primary)
    expect(current.background).not.toBe(original.background)
    backgrounds.add(current.background)
    await expect(icon).toHaveAttribute('href',artwork!)
  }
  expect(backgrounds.size).toBe(11)
  const selected=await appearance()
  await selectStyledOption(page,palette,'iris')
  await material.uncheck()
  await expect(root).toHaveAttribute('data-palette','iris')
  await expect(root).toHaveAttribute('data-translucency','off')
  expect((await appearance()).background).toBe(selected.background)
  await page.reload()
  await expect(wallpaper).toHaveAttribute('data-value','b-horizon')
  await expect(palette).toHaveAttribute('data-value','iris')
  await expect(material).not.toBeChecked()
  expect((await appearance()).background).toBe(selected.background)
  await expect(icon).toHaveAttribute('href',artwork!)
  await selectStyledOption(page,wallpaper,'palette')
  const themeBackground=(await appearance()).background
  expect(themeBackground).not.toBe(selected.background)
  await selectStyledOption(page,palette,'ice')
  expect((await appearance()).background).toBe(original.background)
  expect(errors).toEqual([])
})

test('workspaces keep independent windows and model bodies through copy, switch and reload',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('原工作区草稿')
  const original=await page.locator('[data-view-tab]').getAttribute('data-view-tab')
  await page.getByRole('button',{name:'切换工作区'}).click();await page.getByRole('textbox',{name:'新工作区名称'}).fill('独立副本');await page.getByRole('button',{name:'复制当前现场'}).click()
  await expect(page.getByRole('dialog',{name:'工作区'})).toBeHidden()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'End');await page.keyboard.insertText('，副本新输入')
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('副本新输入')
  await page.getByRole('button',{name:'切换工作区'}).click();await page.getByRole('dialog',{name:'工作区'}).getByRole('button',{name:'个人桌面',exact:true}).click()
  await expect(page.locator(`[data-view-tab="${original}"]`)).toHaveCount(1);await expect(page.locator('.monaco-editor .view-lines')).toContainText('原工作区草稿');await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('副本新输入')
  await page.getByRole('button',{name:'关闭窗口',exact:true}).click();await page.getByRole('button',{name:'重新打开最近标签'}).click()
  await expect(page.locator(`[data-view-tab="${original}"]`)).toHaveCount(1);await expect(page.locator('.monaco-editor .view-lines')).toContainText('原工作区草稿');expect(errors).toEqual([])
})

test('all eight resize handles and snapped restore preserve geometry without remote operations',async({page})=>{
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click()
  const win=page.locator('.app-window'),initial=await win.boundingBox()
  for(const direction of ['n','ne','e','se','s','sw','w','nw']){
    const handle=await win.locator(`[data-resize="${direction}"]`).boundingBox();const start={x:handle!.x+handle!.width/2,y:handle!.y+handle!.height/2}
    await page.mouse.move(start.x,start.y);await page.mouse.down();await page.mouse.move(start.x+(direction.includes('w')?-9:direction.includes('e')?9:0),start.y+(direction.includes('n')?-4:direction.includes('s')?4:0));await page.mouse.up()
  }
  expect((await win.boundingBox())!.width).toBeGreaterThan(initial!.width)
  const title=await win.locator('.window-titlebar').boundingBox()
  await page.mouse.move(title!.x+220,title!.y+19);await page.mouse.down();await page.mouse.move(5,200);await expect(page.locator('.snap-preview.left')).toBeVisible();await page.mouse.up()
  await expect(win).toHaveClass(/snapped/);await page.reload();await expect(page.locator('.app-window')).toHaveClass(/snapped/)
  await page.getByRole('button',{name:'最大化或还原窗口'}).click();await expect(page.locator('.app-window')).not.toHaveClass(/snapped/)
  await page.locator('.taskbar-app').click();await expect(page.locator('.app-window')).toBeHidden();await page.locator('.taskbar-app').click();await expect(page.locator('.app-window')).toBeVisible()
})

test('copying a browser tab creates an isolated branch and immediate typing keeps original unchanged',async({page,context})=>{
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click();await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('来源草稿')
  const storage=await page.evaluate(()=>Object.fromEntries(Object.entries(sessionStorage)))
  const duplicate=await context.newPage()
  await duplicate.route('**/api/v1/**',route=>route.fulfill({json:route.request().url().endsWith('/session')?{user:{userId:'browser-test',name:'测试用户',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await duplicate.addInitScript(values=>{if(!sessionStorage.getItem('branch-seeded')){for(const [key,value] of Object.entries(values))sessionStorage.setItem(key,value);sessionStorage.setItem('branch-seeded','1')}},storage)
  await duplicate.goto('/');await expect(duplicate.locator('.monaco-editor .view-lines')).toContainText('来源草稿')
  expect(await duplicate.evaluate(()=>sessionStorage.getItem('blora:tab'))).not.toBe(await page.evaluate(()=>sessionStorage.getItem('blora:tab')))
  await duplicate.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(duplicate,'End');await duplicate.keyboard.insertText('分支输入');await duplicate.reload();await expect(duplicate.locator('.monaco-editor .view-lines')).toContainText('分支输入')
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('分支输入');await duplicate.close()
})

test('native Chromium IME composition survives immediate reload',async({page,context,browserName})=>{
  test.skip(browserName!=='chromium','Native composition injection requires CDP; Firefox/WebKit native IME remains unverified.')
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click();await page.getByRole('textbox',{name:'文件正文编辑器'}).focus()
  const input=await context.newCDPSession(page)
  await input.send('Input.imeSetComposition',{text:'中文组合',selectionStart:4,selectionEnd:4})
  await input.send('Input.insertText',{text:'中文组合'})
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('中文组合')
})

test('committed Unicode, simultaneous multi-cursor edits and find state survive immediate reload',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click();await page.getByRole('textbox',{name:'文件正文编辑器'}).focus()
  await page.keyboard.insertText('中文组合')
  await page.keyboard.insertText('\nmatch\nmatch')
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('中文组合')
  await page.getByRole('button',{name:'查找',exact:true}).click();await page.getByRole('textbox',{name:'查找内容',exact:true}).fill('match');await page.getByRole('textbox',{name:'替换内容',exact:true}).fill('替换草稿')
  await expect(page.getByRole('textbox',{name:'查找内容',exact:true})).toHaveValue('match');await expect(page.getByRole('textbox',{name:'替换内容',exact:true})).toHaveValue('替换草稿')
  await page.getByRole('button',{name:'选择全部匹配',exact:true}).click();await page.keyboard.insertText('双光标')
  // Native EditContext can emit one semantic edit per character for a multi-cursor
  // insertion. Compare the actual available undo history before and after reload.
  const lines=page.locator('.monaco-editor .view-lines')
  await expect(lines).toContainText('双光标');const beforeUndo=await lines.innerText()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'z')
  // Monaco paints asynchronously in WebKit. Reading innerText immediately
  // after the key can capture the pre-undo frame as the expected undo result.
  await expect(lines).not.toHaveText(beforeUndo,{useInnerText:true});const priorUndo=await lines.innerText()
  await pressEditorKey(page,'y');await expect(lines).toHaveText(beforeUndo,{useInnerText:true})
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('中文组合');await expect(page.locator('.monaco-editor .view-lines')).toContainText('双光标')
  await expect(page.getByRole('textbox',{name:'查找内容',exact:true})).toHaveValue('match');await expect(page.getByRole('textbox',{name:'替换内容',exact:true})).toHaveValue('替换草稿')
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'z');await expect(page.locator('.monaco-editor .view-lines')).toHaveText(priorUndo,{useInnerText:true})
  await pressEditorKey(page,'y');await expect(page.locator('.monaco-editor .view-lines')).toContainText('双光标');expect(errors).toEqual([])
})
import {nativeDeviceScale} from '../../playwright-browser'
