import type {Page} from '@playwright/test'

// Diagnostic only. Merge the existing decoded, content-free native shadow
// pixels into four perimeter bands; never read or capture application pixels.
export async function installShadowBandDiagnostic(page:Page){
 await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
 await page.evaluate(async()=>{
  const names=['nw','n','ne','w','e','sw','s','se'],bands=['top','left','right','bottom'],urls:string[]=[]
  for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane')){
   const computed=getComputedStyle(plane),pad=parseFloat(computed.getPropertyValue('--shadow-outset')),corner=parseFloat(computed.getPropertyValue('--shadow-corner'))
   const width=plane.offsetWidth,height=plane.offsetHeight,cut=pad+corner,dpr=devicePixelRatio||1
   const dimensions=[{x:-pad,y:-pad,width:width+2*pad,height:cut},{x:-pad,y:corner,width:cut,height:height-2*corner},{x:width-corner,y:corner,width:cut,height:height-2*corner},{x:-pad,y:height-corner,width:width+2*pad,height:cut}]
   if(dimensions.some(r=>r.width<=0||r.height<=0))continue
   const elements=dimensions.map((r,index)=>{
    const element=document.createElement('i');element.className='shadow-band '+bands[index]
    element.style.cssText=`position:absolute;left:${r.x}px;top:${r.y}px;width:${r.width}px;height:${r.height}px;background-size:100% 100%;background-repeat:no-repeat`
    return element
   })
   for(const state of ['normal','focused']){
    const canvases=dimensions.map(r=>{const canvas=document.createElement('canvas');canvas.width=Math.round(r.width*dpr);canvas.height=Math.round(r.height*dpr);return canvas})
    for(let index=0;index<names.length;index++){
     const name=names[index]!,piece=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!,style=getComputedStyle(piece)
     if(style.getPropertyValue(`--shadow-${state}-display`).trim()==='none')continue
     const value=(key:string)=>parseFloat(style.getPropertyValue(`--shadow-${state}-${key}`))
     const intrinsicWidth=value('width'),intrinsicHeight=value('height'),w=['n','s'].includes(name)?width-2*corner:intrinsicWidth,h=['w','e'].includes(name)?height-2*corner:intrinsicHeight
     const x=['nw','w','sw'].includes(name)?-pad+value('left'):['ne','e','se'].includes(name)?width+pad-value('right')-w:corner
     const y=['nw','n','ne'].includes(name)?-pad+value('top'):['sw','s','se'].includes(name)?height+pad-value('bottom')-h:corner
     const image=new Image(),match=/url\(["']?([^"')]+)["']?\)/.exec(style.getPropertyValue(`--shadow-${state}`));if(!match)throw Error('shadow diagnostic source missing')
     image.src=match[1]!;await image.decode()
     const band=index<3?0:index===3?1:index===4?2:3,canvas=canvases[band]!,context=canvas.getContext('2d')!,region=dimensions[band]!
     context.drawImage(image,(x-region.x)*dpr,(y-region.y)*dpr,w*dpr,h*dpr)
    }
    for(let index=0;index<canvases.length;index++){
     const blob=await new Promise<Blob>((resolve,reject)=>canvases[index]!.toBlob(value=>value?resolve(value):reject(Error('shadow diagnostic encoding failed'))))
     const url=URL.createObjectURL(blob);urls.push(url);const image=new Image();image.src=url;await image.decode()
     elements[index]!.style.setProperty(`--shadow-band-${state}`,`url("${url}")`)
    }
   }
   plane.append(...elements);plane.dataset.shadowBands='ready'
  }
  window.addEventListener('pagehide',()=>{for(const url of urls)URL.revokeObjectURL(url)},{once:true})
 })
 await page.addStyleTag({content:'.window-shadow-plane[data-shadow-bands="ready"]>.shadow-piece{display:none!important}.shadow-band{background-image:var(--shadow-band-normal)}.app-window.focused>.window-shadow-plane>.shadow-band{background-image:var(--shadow-band-focused)}.app-window[data-shadow-low="true"]>.window-shadow-plane>.shadow-band{background-image:var(--shadow-band-normal)}'})
}
