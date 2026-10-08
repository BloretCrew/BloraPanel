import type {Page} from '@playwright/test'

// Diagnostic only. Expand native optical edge pixels to their display size
// before interaction, so the software compositor need not scale a two-pixel
// source across every long window edge. No application pixels are read.
export async function installShadowResolutionDiagnostic(page:Page){
 await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
 await page.evaluate(async()=>{
  const cache=new Map<string,Promise<string>>(),urls:string[]=[],dpr=devicePixelRatio||1
  for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane')){
   const box=plane.getBoundingClientRect(),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner'))
   for(const name of ['n','s','w','e']){
    const piece=plane.querySelector<HTMLElement>('.shadow-piece.'+name);if(!piece)continue
    for(const state of ['normal','focused']){
     if(piece.style.getPropertyValue(`--shadow-${state}-display`)==='none')continue
     const horizontal=name==='n'||name==='s',width=horizontal?box.width-2*corner:parseFloat(piece.style.getPropertyValue(`--shadow-${state}-width`)),height=horizontal?parseFloat(piece.style.getPropertyValue(`--shadow-${state}-height`)):box.height-2*corner
     const source=piece.style.getPropertyValue(`--shadow-${state}`),match=/url\(["']?([^"')]+)["']?\)/.exec(source)
     if(!match||width<=0||height<=0)continue
     const w=Math.max(1,Math.round(width*dpr)),h=Math.max(1,Math.round(height*dpr)),key=JSON.stringify([source,w,h])
     if(!cache.has(key))cache.set(key,(async()=>{
      const image=new Image();image.src=match[1]!;await image.decode()
      const canvas=document.createElement('canvas');canvas.width=w;canvas.height=h;canvas.getContext('2d')!.drawImage(image,0,0,w,h)
      const blob=await new Promise<Blob>((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(Error('shadow resolution diagnostic unavailable'))))
      const url=URL.createObjectURL(blob);urls.push(url);const decoded=new Image();decoded.src=url;await decoded.decode();return `url("${url}")`
     })())
     piece.style.setProperty(`--shadow-${state}`,await cache.get(key)!)
    }
   }
  }
  window.addEventListener('pagehide',()=>{for(const url of urls)URL.revokeObjectURL(url)},{once:true})
 })
}
