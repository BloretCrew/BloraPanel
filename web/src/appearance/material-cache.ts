// One wallpaper-only optical raster. Live text, terminal output and extension
// frames never enter the cache, and keep their own independent render lifecycle.
import {blurMaterial} from './material-blur'
import {encodeMaterial} from './material-encoding'
const surfaces='.topbar,.launcher,.account-popover,.workspace-menu,.context-menu,.window-picker,.notification-panel,.shortcut-rename-dialog,.dock-backdrop'
const branches='.topbar,.taskbar'
export function installMaterialCache(desktop:HTMLElement){
 const root=document.documentElement
 // Chromium's separately retained wallpaper plane can accumulate alpha at
 // a resized ancestor's rounded clip. Keep only that optical plane in its
 // existing native window surface; application nodes and shadow layers stay.
 const previousBodyPaint=desktop.dataset.materialBodyPaint
 if(/(?:Chrome|Chromium|Edg)\//.test(navigator.userAgent))desktop.dataset.materialBodyPaint='native'
 let generation=0,disposed=false,urls:string[]=[],timer:ReturnType<typeof setTimeout>|undefined
 const invalidate=()=>{delete root.dataset.materialCache;delete root.dataset.materialCacheOpaque;delete root.dataset.materialCacheEncoding}
 const tracked=new Set<HTMLElement>()
 const locate=(element:HTMLElement)=>{const b=element.getBoundingClientRect();element.style.setProperty('--material-x',`${-b.x}px`);element.style.setProperty('--material-y',`${-b.y}px`)}
 const sizes=new ResizeObserver(entries=>{for(const entry of entries)locate(entry.target as HTMLElement)})
 const scan=()=>{
  // Host materials exist at the desktop root or inside the bar/Dock. Never
  // subscribe to application descendants: default terminal row replacement
  // and Monaco decorations can produce thousands of irrelevant records.
  children.disconnect();children.observe(desktop,{childList:true})
  const candidates=new Set<HTMLElement>()
  for(const child of desktop.children){
   if(!(child instanceof HTMLElement))continue
   if(child.matches(surfaces))candidates.add(child)
   if(child.matches(branches)){
    children.observe(child,{childList:true,subtree:true})
    for(const element of child.querySelectorAll<HTMLElement>(surfaces))candidates.add(element)
   }
  }
  for(const element of tracked)if(!candidates.has(element)){sizes.unobserve(element);tracked.delete(element)}
  for(const element of candidates)if(!tracked.has(element)){tracked.add(element);locate(element);sizes.observe(element)}
 }
 // Terminal row updates are not surface changes. Avoid querying the desktop
 // for every output chunk; only new/removed optical surfaces need registration.
 const children=new MutationObserver(records=>{
  if(records.some(r=>[...r.addedNodes].some(n=>n instanceof HTMLElement&&(n.matches(surfaces)||n.matches(branches)||(r.target!==desktop&&!!n.querySelector(surfaces))))||[...r.removedNodes].some(n=>[...tracked].some(e=>n===e||n.contains(e)))))scan()
 })
 async function render(){
  const version=++generation
  if(root.dataset.translucency==='off'){invalidate();return}
  const style=getComputedStyle(desktop),width=desktop.clientWidth,height=desktop.clientHeight
  const matches=[...style.backgroundImage.matchAll(/url\((?:"([^"]*)"|'([^']*)'|([^)]*))\)/g)]
  if(matches.length!==1){invalidate();return}
  try{
   const match=matches[0]!,image=new Image();image.src=match[1]||match[2]||match[3]!;await image.decode()
   if(disposed||version!==generation)return
   const scale=Math.min(.5,2048/Math.max(width,height)),sigma=(root.dataset.theme==='dark'?20:24)*scale,pad=Math.ceil(sigma*3),w=Math.ceil(width*scale),h=Math.ceil(height*scale)
   const source=document.createElement('canvas');source.width=w+pad*2;source.height=h+pad*2
   // A solid desktop is fully opaque. Preserve that fact in the raster instead
   // of making the software compositor blend a redundant alpha plane per
   // window. Transparent custom desktops retain their original alpha path.
   const opaque=/^rgb\(/i.test(style.backgroundColor)||/^rgba\([^)]*,\s*1(?:\.0*)?\)$/i.test(style.backgroundColor)
   const ctx=source.getContext('2d',{alpha:!opaque});if(!ctx)throw Error('canvas unavailable')
   ctx.fillStyle=style.backgroundColor;ctx.fillRect(0,0,source.width,source.height)
   const ratio=Math.max(w/image.naturalWidth,h/image.naturalHeight),iw=image.naturalWidth*ratio,ih=image.naturalHeight*ratio
   ctx.drawImage(image,pad+(w-iw)/2,pad+(h-ih)/2,iw,ih)
   const veil=style.backgroundImage.match(/linear-gradient\((rgba?\([^)]*\)|#[a-f0-9]{8}),\s*\1\)/i)
   if(veil){ctx.fillStyle=veil[1]!;ctx.fillRect(pad,pad,w,h)}
   ctx.drawImage(source,pad,pad,1,h,0,pad,pad,h)
   ctx.drawImage(source,pad+w-1,pad,1,h,pad+w,pad,pad,h)
   ctx.drawImage(source,0,pad,source.width,1,0,0,source.width,pad)
   ctx.drawImage(source,0,pad+h-1,source.width,1,0,pad+h,source.width,pad)
   const output=document.createElement('canvas');output.width=w;output.height=h
   const out=output.getContext('2d',{alpha:!opaque});if(!out)throw Error('canvas unavailable')
   if('filter' in out)out.filter=`blur(${sigma}px) saturate(1.05)`
   else{
    const raster=ctx.getImageData(0,0,source.width,source.height)
    raster.data.set(blurMaterial(raster.data,source.width,source.height,sigma))
    for(let i=0;i<raster.data.length;i+=4){const r=raster.data[i]!,g=raster.data[i+1]!,b=raster.data[i+2]!,l=.213*r+.715*g+.072*b;raster.data[i]=l+(r-l)*1.05;raster.data[i+1]=l+(g-l)*1.05;raster.data[i+2]=l+(b-l)*1.05}
    ctx.putImageData(raster,0,0)
   }
   out.drawImage(source,-pad,-pad)
   // Bake each stable surface tint into the wallpaper once. The software
   // compositor then samples one opaque image, rather than blending a second
   // full-size translucent gradient on every title/body repaint.
   const variants=['wallpaper','window','face','menu','desktop'] as const,candidates:string[]=[],encodings:string[]=[]
   try{
    for(const variant of variants){
     let canvas=output
     if(variant==='desktop'){
      // Keep the visible wallpaper itself as a full-resolution bitmap too.
      // Re-exposing the original filtered SVG beneath moving windows can make
      // engines rerasterize its gradients/filters on every desktop repaint.
      const resolution=Math.min(devicePixelRatio||1,2,4096/Math.max(width,height))
      canvas=document.createElement('canvas');canvas.width=Math.ceil(width*resolution);canvas.height=Math.ceil(height*resolution)
      const painted=canvas.getContext('2d',{alpha:!opaque});if(!painted)throw Error('canvas unavailable')
      painted.fillStyle=style.backgroundColor;painted.fillRect(0,0,canvas.width,canvas.height)
      const fit=Math.max(canvas.width/image.naturalWidth,canvas.height/image.naturalHeight),iw=image.naturalWidth*fit,ih=image.naturalHeight*fit
      painted.drawImage(image,(canvas.width-iw)/2,(canvas.height-ih)/2,iw,ih)
      if(veil){painted.fillStyle=veil[1]!;painted.fillRect(0,0,canvas.width,canvas.height)}
     }else if(variant!=='wallpaper'){
      canvas=document.createElement('canvas');canvas.width=w;canvas.height=h
      const painted=canvas.getContext('2d',{alpha:!opaque});if(!painted)throw Error('canvas unavailable')
      painted.drawImage(output,0,0)
      const probe=document.createElement('span');probe.style.cssText=`position:absolute;visibility:hidden;background:var(--glass-${variant})`;desktop.append(probe)
      try{painted.fillStyle=getComputedStyle(probe).backgroundColor}finally{probe.remove()}
     painted.fillRect(0,0,w,h)
     }
     const blob=await encodeMaterial(canvas,opaque)
     encodings.push(blob.type)
     if(disposed||version!==generation){for(const candidate of candidates)URL.revokeObjectURL(candidate);return}
     const candidate=URL.createObjectURL(blob);candidates.push(candidate)
     const decoded=new Image();decoded.src=candidate;await decoded.decode()
    }
   }catch(error){for(const candidate of candidates)URL.revokeObjectURL(candidate);throw error}
   if(disposed||version!==generation){for(const candidate of candidates)URL.revokeObjectURL(candidate);return}
   const previous=urls;urls=candidates
   variants.forEach((variant,index)=>root.style.setProperty(`--material-${variant}`,`url("${urls[index]}")`));root.style.setProperty('--material-size',`${width}px ${height}px`)
   root.style.setProperty('--material-blur',root.dataset.theme==='dark'?'20px':'24px')
   root.dataset.materialCacheOpaque=String(opaque);root.dataset.materialCache='ready';scan();for(const old of previous)URL.revokeObjectURL(old)
   root.dataset.materialCacheEncoding=encodings.every(value=>value==='image/bmp')?'opaque-rgb24':encodings.every(value=>value==='image/png')?'native-png':'mixed'
  }catch{if(!disposed&&version===generation)invalidate()}
 }
  const schedule=()=>{generation++;invalidate();if(timer)clearTimeout(timer);timer=setTimeout(()=>void render(),80)}
 const preferences=new MutationObserver(schedule)
 preferences.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-wallpaper-choice','data-translucency']})
 const resize=()=>{for(const element of tracked)locate(element);schedule()}
 const motionDone=(event:Event)=>{const element=event.target;if(element instanceof HTMLElement&&tracked.has(element))locate(element)}
 desktop.addEventListener('transitionend',motionDone);desktop.addEventListener('animationend',motionDone);desktop.addEventListener('material-position-settled',motionDone)
 window.addEventListener('resize',resize);void render();scan()
 return ()=>{disposed=true;generation++;if(timer)clearTimeout(timer);children.disconnect();preferences.disconnect();sizes.disconnect();window.removeEventListener('resize',resize);desktop.removeEventListener('transitionend',motionDone);desktop.removeEventListener('animationend',motionDone);desktop.removeEventListener('material-position-settled',motionDone);for(const url of urls)URL.revokeObjectURL(url);delete root.dataset.materialCache;delete root.dataset.materialCacheOpaque;delete root.dataset.materialCacheEncoding;if(previousBodyPaint===undefined)delete desktop.dataset.materialBodyPaint;else desktop.dataset.materialBodyPaint=previousBodyPaint;for(const variant of ['wallpaper','window','face','menu','desktop'])root.style.removeProperty(`--material-${variant}`);root.style.removeProperty('--material-size');root.style.removeProperty('--material-blur')}
}
