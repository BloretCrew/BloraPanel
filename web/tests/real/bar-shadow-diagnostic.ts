// Explicit content-free optical diagnostic, not installed by the product.
// The isolated native rectangle contains only resolved CSS shadow/contour.
export async function installBarShadowDiagnostic(source:'png'|'native-svg'|'native-full'='png'){
 const bar=document.querySelector<HTMLElement>('.topbar')
 if(!bar)return
 const root=document.documentElement,image=document.createElement('img'),css=document.createElement('style')
 image.className='bar-shadow-diagnostic';image.alt='';image.draggable=false
 image.style.cssText='position:absolute;pointer-events:none;max-width:none;display:none;z-index:-1'
 css.textContent='.topbar[data-bar-shadow-cache="ready"]{box-shadow:none!important}'
 document.head.append(css);bar.append(image)
 let disposed=false,paused=false,version=0,frame=0,busy=false,pending=false,key='',owned:string|undefined
 const escape=(value:string)=>value.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;')
 function clear(){delete bar!.dataset.barShadowCache;image.style.display='none';image.removeAttribute('src');if(owned)URL.revokeObjectURL(owned);owned=undefined;key=''}
 function measure(){
  if(devicePixelRatio!==1||!bar!.isConnected)return undefined
  // Read the original native shadow, never our own substituted rule. This
  // synchronous read cannot paint an intervening fallback frame.
  const previous=bar!.dataset.barShadowCache;delete bar!.dataset.barShadowCache
  // Computed declarations are live. Copy every input before restoring the
  // substitution, with one native style read and no intervening paint.
  let shadow:string,radii:string[],borderX:number,borderY:number
  try{const style=getComputedStyle(bar!);shadow=style.boxShadow;radii=[style.borderTopLeftRadius,style.borderTopRightRadius,style.borderBottomLeftRadius,style.borderBottomRightRadius];borderX=parseFloat(style.borderLeftWidth);borderY=parseFloat(style.borderTopWidth)}finally{if(previous)bar!.dataset.barShadowCache=previous}
  const box=bar!.getBoundingClientRect(),radius=parseFloat(radii[0]!)
  if(shadow==='none'||borderX!==borderY||!radii.every(r=>r===radii[0])||!Number.isFinite(radius)||radius<0||radius>128||box.width<=0||box.width>4096||box.height<=0||box.height>128)return undefined
  const lengths=[...shadow.matchAll(/(-?[\d.]+)px\b/g)].map(m=>Number(m[1]))
  if(lengths.some(n=>!Number.isFinite(n)||Math.abs(n)>256))return undefined
  // Conservative bound includes every blur/offset/spread; trim only alpha0.
  const pad=Math.ceil(lengths.reduce((sum,n)=>sum+Math.abs(n)*3,0))+2
  if(pad>512)return undefined
  const value={width:box.width,height:box.height,radius,shadow,borderX,borderY,pad,phaseX:box.x-Math.floor(box.x),phaseY:box.y-Math.floor(box.y)}
  return {...value,key:JSON.stringify(value)}
 }
 async function update(){
  if(disposed||paused)return
  if(busy){pending=true;return}
  const value=measure();if(!value){clear();return}if(key===value.key)return
  clear();busy=true;pending=false;const ticket=version
  let candidate:string|undefined
  try{
   const {width,height,radius,shadow,borderX,borderY,pad,phaseX,phaseY}=value
   const w=Math.ceil(width+2*pad+phaseX),h=Math.ceil(height+2*pad+phaseY)
   const svg=`<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}"><foreignObject width="${w}" height="${h}"><div xmlns="http://www.w3.org/1999/xhtml" style="position:absolute;left:${pad+phaseX}px;top:${pad+phaseY}px;width:${width}px;height:${height}px;box-sizing:border-box;border:${borderY}px solid transparent;border-radius:${radius}px;box-shadow:${escape(shadow)}"></div></foreignObject></svg>`
   const native=new Image();native.src='data:image/svg+xml;charset=utf-8,'+encodeURIComponent(svg);await native.decode()
   const canvas=document.createElement('canvas');canvas.width=w;canvas.height=h
   const context=canvas.getContext('2d');if(!context)throw Error('native optical canvas unavailable')
   context.drawImage(native,0,0)
   const data=context.getImageData(0,0,w,h).data
   let left=w,top=h,right=0,bottom=0
   for(let y=0;y<h;y++)for(let x=0;x<w;x++)if(data[(y*w+x)*4+3]){left=Math.min(left,x);top=Math.min(top,y);right=Math.max(right,x+1);bottom=Math.max(bottom,y+1)}
   if(right<=left||bottom<=top)throw Error('native optical raster is empty')
   const tile=document.createElement('canvas');tile.width=right-left;tile.height=bottom-top
   const target=tile.getContext('2d');if(!target)throw Error('native optical tile unavailable')
   target.drawImage(canvas,left,top,tile.width,tile.height,0,0,tile.width,tile.height)
   if(source==='native-full')candidate=URL.createObjectURL(new Blob([svg],{type:'image/svg+xml'}))
   else if(source==='native-svg'){
    // Same native optical document, cropped 1:1 without a second PNG alpha
    // conversion. No live surface or application data enters the document.
    const cropped=svg.replace(`width="${w}" height="${h}">`,`width="${tile.width}" height="${tile.height}" viewBox="${left} ${top} ${tile.width} ${tile.height}">`)
    candidate=URL.createObjectURL(new Blob([cropped],{type:'image/svg+xml'}))
   }else{
    const blob=await new Promise<Blob>((resolve,reject)=>tile.toBlob(blob=>blob?resolve(blob):reject(Error('native optical encoding unavailable'))))
    candidate=URL.createObjectURL(blob)
   }
   const decoded=new Image();decoded.src=candidate;await decoded.decode()
   if(disposed||paused||ticket!==version||measure()?.key!==value.key)return
   image.src=candidate;image.style.left=`${(source==='native-full'?0:left)-pad-phaseX-borderX}px`;image.style.top=`${(source==='native-full'?0:top)-pad-phaseY-borderY}px`
   image.style.width=`${source==='native-full'?w:tile.width}px`;image.style.height=`${source==='native-full'?h:tile.height}px`;image.style.display='block'
   owned=candidate;candidate=undefined;key=value.key;bar!.dataset.barShadowCache='ready'
  }catch{if(!disposed&&ticket===version)clear()}
  finally{if(candidate)URL.revokeObjectURL(candidate);busy=false;if(pending&&!disposed){pending=false;queue()}}
 }
 function queue(){version++;clear();if(busy)pending=true;if(!frame)frame=requestAnimationFrame(()=>{frame=0;void update()})}
 const sizes=new ResizeObserver(queue);sizes.observe(bar)
 const appearance=new MutationObserver(queue);appearance.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-translucency','data-density']})
 window.addEventListener('resize',queue)
 const controller={pause(){paused=true;delete bar!.dataset.barShadowCache;image.style.display='none'},resume(){paused=false;if(owned&&measure()?.key===key){image.style.display='block';bar!.dataset.barShadowCache='ready'}else queue()},dispose(){disposed=true;version++;if(frame)cancelAnimationFrame(frame);sizes.disconnect();appearance.disconnect();window.removeEventListener('resize',queue);clear();image.remove();css.remove()}}
 ;(window as unknown as {__barShadowDiagnostic:typeof controller}).__barShadowDiagnostic=controller
 window.addEventListener('pagehide',controller.dispose,{once:true})
 await update()
}
