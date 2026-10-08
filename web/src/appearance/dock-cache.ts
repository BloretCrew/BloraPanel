// Cache only the native Dock contour and its resolved optical shadow. Icons,
// labels, badges, buttons and wallpaper remain separate live application layers.
export function installDockCache(surface:HTMLElement){
 const svg=surface.querySelector<SVGSVGElement>(':scope>svg'),path=svg?.querySelector('path')
 if(!svg||!path)return ()=>{}
 const root=document.documentElement,image=document.createElement('img')
 image.className='dock-optical-cache';image.alt='';image.draggable=false
 image.style.cssText='position:absolute;pointer-events:none;max-width:none;display:none'
 surface.append(image)
 let disposed=false,version=0,frame=0,busy=false,pending=false,key='',ownedUrl:string|undefined
 const escape=(value:string)=>value.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;')
 function clear(){
  delete surface.dataset.dockCache;image.style.display='none';image.removeAttribute('src')
  if(ownedUrl)URL.revokeObjectURL(ownedUrl);ownedUrl=undefined;key=''
 }
 function measure(){
  // Fractional DPR retains the original vector/filter path. The native phase
  // at DPR 1 is included in the raster so even half-pixel centered Docks keep
  // their original antialiasing instead of receiving a second image resample.
  if(devicePixelRatio!==1||!surface.isConnected)return undefined
  const box=surface.getBoundingClientRect(),d=path!.getAttribute('d')||'',style=getComputedStyle(svg!),face=getComputedStyle(path!)
  if(!d||box.width<=0||box.height<=0||box.width>4096||box.height>512)return undefined
  const filter=style.filter
  if(filter!=='none'&&!/^drop-shadow\(/.test(filter))return undefined
  const lengths=[...filter.matchAll(/(-?[\d.]+)px\b/g)].map(match=>Number(match[1]))
  if(lengths.some(value=>!Number.isFinite(value)||Math.abs(value)>256)||lengths.length%3)return undefined
  let extent=0
  for(let i=0;i<lengths.length;i+=3)extent+=Math.max(Math.abs(lengths[i]!),Math.abs(lengths[i+1]!))+3*Math.max(0,lengths[i+2]!)
  const pad=Math.ceil(extent)+2,phaseX=box.x-Math.floor(box.x),phaseY=box.y-Math.floor(box.y)
  const value={width:box.width,height:box.height,d,fill:face.fill,stroke:face.stroke,strokeWidth:face.strokeWidth,filter,pad,phaseX,phaseY}
  return {...value,key:JSON.stringify(value)}
 }
 async function update(){
  if(disposed)return
  if(busy){pending=true;return}
  const value=measure();if(!value){clear();return}if(key===value.key)return
  clear();busy=true;pending=false;const ticket=version
  let url:string|undefined
  try{
   const {width,height,d,fill,stroke,strokeWidth,filter,pad,phaseX,phaseY}=value
   const pixelsW=Math.ceil(width+2*pad+phaseX),pixelsH=Math.ceil(height+2*pad+phaseY)
   // Isolated numeric SVG: no app DOM, fonts, user data or external resources.
   const source=`<svg xmlns="http://www.w3.org/2000/svg" width="${pixelsW}" height="${pixelsH}"><foreignObject width="${pixelsW}" height="${pixelsH}" style="overflow:visible"><div xmlns="http://www.w3.org/1999/xhtml" style="position:absolute;left:${pad+phaseX}px;top:${pad+phaseY}px;width:${width}px;height:${height}px"><svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" style="overflow:visible;filter:${escape(filter)}"><path d="${escape(d)}" fill="${escape(fill)}" stroke="${escape(stroke)}" stroke-width="${escape(strokeWidth)}"/></svg></div></foreignObject></svg>`
   const native=new Image();native.src='data:image/svg+xml;charset=utf-8,'+encodeURIComponent(source);await native.decode()
   const canvas=document.createElement('canvas');canvas.width=pixelsW;canvas.height=pixelsH
   const context=canvas.getContext('2d');if(!context)throw Error('Dock optical canvas unavailable')
   context.drawImage(native,0,0)
   const data=context.getImageData(0,0,pixelsW,pixelsH).data
   let left=pixelsW,top=pixelsH,right=0,bottom=0
   for(let y=0;y<pixelsH;y++)for(let x=0;x<pixelsW;x++)if(data[(y*pixelsW+x)*4+3]){left=Math.min(left,x);top=Math.min(top,y);right=Math.max(right,x+1);bottom=Math.max(bottom,y+1)}
   if(right<=left||bottom<=top)throw Error('Dock optical raster is empty')
   // Trim only zero alpha; faint shadow pixels and the native phase are kept.
   const tile=document.createElement('canvas');tile.width=right-left;tile.height=bottom-top
   const tileContext=tile.getContext('2d');if(!tileContext)throw Error('Dock optical tile unavailable')
   tileContext.drawImage(canvas,left,top,tile.width,tile.height,0,0,tile.width,tile.height)
   const blob=await new Promise<Blob>((resolve,reject)=>tile.toBlob(value=>value?resolve(value):reject(Error('Dock optical encoding unavailable'))))
   url=URL.createObjectURL(blob);const decoded=new Image();decoded.src=url;await decoded.decode()
   if(disposed||ticket!==version||measure()?.key!==value.key)return
   image.src=url;image.style.left=`${left-pad-phaseX}px`;image.style.top=`${top-pad-phaseY}px`
   image.style.width=`${tile.width}px`;image.style.height=`${tile.height}px`;image.style.display='block'
   ownedUrl=url;url=undefined;key=value.key;surface.dataset.dockCache='ready'
  }catch{if(!disposed&&ticket===version)clear()}
  finally{if(url)URL.revokeObjectURL(url);busy=false;if(pending&&!disposed){pending=false;queue()}}
 }
 function queue(){
  version++;clear();if(busy)pending=true
  if(!frame)frame=requestAnimationFrame(()=>{frame=0;void update()})
 }
 const sizes=new ResizeObserver(queue);sizes.observe(surface)
 const contour=new MutationObserver(queue);contour.observe(path,{attributes:true,attributeFilter:['d']});contour.observe(svg,{attributes:true,attributeFilter:['viewBox']})
 const appearance=new MutationObserver(queue);appearance.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-translucency']})
 window.addEventListener('resize',queue);queue()
 return ()=>{disposed=true;version++;if(frame)cancelAnimationFrame(frame);sizes.disconnect();contour.disconnect();appearance.disconnect();window.removeEventListener('resize',queue);clear();image.remove()}
}
