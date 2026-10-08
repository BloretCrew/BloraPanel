import type {Page} from '@playwright/test'

// Diagnostic only. Quantize just alpha=1 in the already isolated, native CSS
// shadow assets, then trim their now-zero margins. No application pixels enter
// these images. Stretch intervals and original source coordinates stay fixed.
export async function installShadowTail(page:Page){
 await page.evaluate(async()=>{
  const frames=[...document.querySelectorAll<HTMLElement>('.desktop>.app-window')]
  if(devicePixelRatio!==1){(window as any).__nativeShadowTail={setVisible:()=>{},dispose:()=>{},trimmedPixels:0};return}
  const urls:string[]=[],records:{frame:HTMLElement;piece:HTMLElement;original:Map<string,string>;candidate:Map<string,string>}[]=[]
  const encoded=new Map<string,Promise<{url:string;left:number;top:number;width:number;height:number;sourceWidth:number;sourceHeight:number;empty:boolean}>>()
  let trimmedPixels=0
  const image=async(source:string,stretchX:boolean,stretchY:boolean)=>{
   const key=source+'|'+stretchX+'|'+stretchY
   if(!encoded.has(key))encoded.set(key,(async()=>{
    const native=new Image();native.src=source;await native.decode()
    const canvas=document.createElement('canvas');canvas.width=native.naturalWidth;canvas.height=native.naturalHeight;const context=canvas.getContext('2d')!;context.drawImage(native,0,0)
    const pixels=context.getImageData(0,0,canvas.width,canvas.height)
    let left=canvas.width,top=canvas.height,right=0,bottom=0
    for(let y=0;y<canvas.height;y++)for(let x=0;x<canvas.width;x++){
     const offset=(y*canvas.width+x)*4
     if(pixels.data[offset+3]===1)pixels.data[offset+3]=0
     if(pixels.data[offset+3]){left=Math.min(left,x);right=Math.max(right,x+1);top=Math.min(top,y);bottom=Math.max(bottom,y+1)}
    }
    const empty=left>=right||top>=bottom
    if(empty){left=top=0;right=bottom=1}
    if(stretchX){left=0;right=canvas.width}if(stretchY){top=0;bottom=canvas.height}
    context.putImageData(pixels,0,0)
    const tile=document.createElement('canvas');tile.width=right-left;tile.height=bottom-top;tile.getContext('2d')!.drawImage(canvas,left,top,tile.width,tile.height,0,0,tile.width,tile.height)
    const blob=await new Promise<Blob>((resolve,reject)=>tile.toBlob(value=>value?resolve(value):reject(Error('Optical shadow encoding failed'))))
    const url=URL.createObjectURL(blob);urls.push(url);const decoded=new Image();decoded.src=url;await decoded.decode()
    return {url,left,top,width:tile.width,height:tile.height,sourceWidth:canvas.width,sourceHeight:canvas.height,empty}
   })())
   return encoded.get(key)!
  }
  for(const frame of frames){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!plane?.dataset.shadowCache)continue
   for(const piece of plane.querySelectorAll<HTMLElement>(':scope>.shadow-piece')){
    if(piece.children.length||piece.textContent)throw Error('Optical-only shadow asset required')
    const original=new Map<string,string>(),candidate=new Map<string,string>()
    for(const state of ['normal','focused']){
     const property='--shadow-'+state,source=/^url\("(.*)"\)$/.exec(piece.style.getPropertyValue(property))?.[1];if(!source)continue
     const stretchX=piece.classList.contains('n')||piece.classList.contains('s'),stretchY=piece.classList.contains('w')||piece.classList.contains('e')
     const asset=await image(source,stretchX,stretchY),number=(key:string)=>parseFloat(piece.style.getPropertyValue(property+'-'+key))
     const sx=number('width')/asset.sourceWidth,sy=number('height')/asset.sourceHeight
     const values={width:asset.width*sx,height:asset.height*sy,left:number('left')+asset.left*sx,top:number('top')+asset.top*sy,right:number('right')+(asset.sourceWidth-asset.left-asset.width)*sx,bottom:number('bottom')+(asset.sourceHeight-asset.top-asset.height)*sy}
     candidate.set(property,`url("${asset.url}")`)
     candidate.set(property+'-display',asset.empty?'none':'block')
     for(const [key,value] of Object.entries(values))candidate.set(property+'-'+key,value+'px')
     if(state==='normal')trimmedPixels+=Math.max(0,asset.sourceWidth*asset.sourceHeight-asset.width*asset.height)
    }
    for(const name of candidate.keys())original.set(name,piece.style.getPropertyValue(name))
    records.push({frame,piece,original,candidate})
   }
  }
  const setVisible=(enabled:boolean)=>{
   for(const {piece,original,candidate} of records)for(const [name,value] of enabled?candidate:original)piece.style.setProperty(name,value)
   // Notify the existing native geometry owner once, not per pointer event.
   for(const frame of new Set(records.map(record=>record.frame)))frame.style.setProperty('--shadow-tail-state',enabled?'1':'0')
  }
  ;(window as any).__nativeShadowTail={setVisible,trimmedPixels,dispose:()=>{setVisible(false);for(const frame of frames)frame.style.removeProperty('--shadow-tail-state');for(const url of urls)URL.revokeObjectURL(url)}}
  setVisible(true)
 })
}
