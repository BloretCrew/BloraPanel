// Retain the browser's own CSS shadow raster, then draw its eight edge patches.
// Empty window interiors are never large translucent shadow/compositor quads.
export function shadowGeometry(normal:string,focused:string,radius:number){
 if(!Number.isFinite(radius)||radius<0||radius>128)return undefined
 let extent=0,reach=0
 for(const value of [normal,focused]){
  if(value==='none')continue
  let depth=0,start=0
  const components:string[]=[]
  for(let i=0;i<=value.length;i++){
   if(value[i]==='(')depth++;if(value[i]===')')depth--
   if(i===value.length||(value[i]===','&&depth===0)){components.push(value.slice(start,i));start=i+1}
  }
  for(const component of components){
   if(/\binset\b/.test(component))return undefined
   const lengths=[...component.matchAll(/(-?[\d.]+)px\b/g)].map(m=>Number(m[1]))
   if(lengths.length<2||lengths.length>4||lengths.some(n=>!Number.isFinite(n)||Math.abs(n)>128))return undefined
   const [x=0,y=0,blur=0,spread=0]=lengths;if(blur<0)return undefined
   const tail=1.5*blur+Math.abs(spread)
   extent=Math.max(extent,tail+Math.max(Math.abs(x),Math.abs(y)))
   reach=Math.max(reach,tail+Math.max(Math.abs(x),Math.abs(y)))
  }
 }
 const pad=Math.ceil(extent)+2,corner=Math.ceil(radius+reach)+1,box=2*corner+2
 return {pad,corner,box,size:box+2*pad,slice:pad+corner}
}
export const SHADOW_PATCHES=['nw','n','ne','w','e','sw','s','se'] as const
// Rasterize CSS at its final pixel resolution. WebKit does not consistently
// scale a foreignObject's shadow with an SVG viewBox at fractional DPR: the
// shadow can be displaced by the unscaled padding. Scaling the resolved CSS
// lengths directly avoids that path; colours and relative elevation stay put.
export function scaleShadowPixels(shadow:string,scale:number){
 return shadow.replace(/(-?[\d.]+)px\b/g,(_,value:string)=>`${Number((Number(value)*scale).toFixed(6))}px`)
}
type Bounds={x:number;y:number;width:number;height:number}
// Normal/focused images use one union crop: not even alpha=1 is discarded.
// Native zero-alpha margins need neither decoding space nor blend quads.
export function shadowPixelBounds(normal:Uint8ClampedArray,focused:Uint8ClampedArray,width:number,height:number):Bounds|undefined{
 let left=width,top=height,right=0,bottom=0
 for(let y=0;y<height;y++)for(let x=0;x<width;x++){
  const alpha=(y*width+x)*4+3
  if(!normal[alpha]&&!focused[alpha])continue
  left=Math.min(left,x);right=Math.max(right,x+1);top=Math.min(top,y);bottom=Math.max(bottom,y+1)
 }
 return right>left&&bottom>top?{x:left,y:top,width:right-left,height:bottom-top}:undefined
}
type Patch={bounds:Bounds;sourceWidth:number;sourceHeight:number;empty:boolean}
type Asset={urls:string[];patches:Patch[][];geometry:NonNullable<ReturnType<typeof shadowGeometry>>;scale:number}
export function installShadowCache(desktop:HTMLElement){
 const root=document.documentElement,tracked=new Map<HTMLElement,HTMLElement>(),cache=new Map<string,Promise<Asset>>(),assets=new Set<Asset>()
 // Chromium requires stable compact edges for exact alpha rounding. Other
 // engines keep their original ownership; retaining their edges separately
 // did not improve the original mixed-load gesture.
 const stableNativeLayers=()=>/(?:Chrome|Chromium|Edg)\//.test(navigator.userAgent)
 const requests=new WeakMap<HTMLElement,number>()
 let generation=0,requestVersion=0,disposed=false,resizeTimer:ReturnType<typeof setTimeout>|undefined
 const dimensions=['width','height','left','top','right','bottom']
 const pieceProperties=['will-change','--shadow-normal','--shadow-focused','--shadow-piece-width','--shadow-piece-height','--shadow-trim-left','--shadow-trim-top','--shadow-trim-right','--shadow-trim-bottom',...['normal','focused'].flatMap(state=>['display',...dimensions].map(name=>`--shadow-${state}-${name}`))]
 const clear=(plane:HTMLElement)=>{delete plane.dataset.shadowCache;for(const name of ['--shadow-width','--shadow-outset','--shadow-corner'])plane.style.removeProperty(name);for(const piece of plane.querySelectorAll<HTMLElement>('.shadow-piece')){delete piece.dataset.shadowEmpty;for(const name of pieceProperties)piece.style.removeProperty(name)}}
 const escape=(value:string)=>value.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;')
 async function raster(normal:string,focused:string,radius:number,geometry:Asset['geometry'],scale:number,version:number){
  const urls:string[]=[]
  try{
   const states:HTMLCanvasElement[]=[]
   for(const shadow of [normal,focused]){
    const {size,pad,box}=geometry,pixels=Math.ceil(size*scale),pixelScale=pixels/size
    // No application content, network images, fonts or user fields enter this
    // isolated SVG. The native CSS renderer supplies the unchanged blur/color.
    const svg=`<svg xmlns="http://www.w3.org/2000/svg" width="${pixels}" height="${pixels}"><foreignObject width="${pixels}" height="${pixels}"><div xmlns="http://www.w3.org/1999/xhtml" style="position:absolute;left:${pad*pixelScale}px;top:${pad*pixelScale}px;width:${box*pixelScale}px;height:${box*pixelScale}px;border-radius:${radius*pixelScale}px;box-shadow:${escape(scaleShadowPixels(shadow,pixelScale))}"></div></foreignObject></svg>`
    const image=new Image();image.src='data:image/svg+xml;charset=utf-8,'+encodeURIComponent(svg);await image.decode()
    const canvas=document.createElement('canvas');canvas.width=pixels;canvas.height=pixels
    const context=canvas.getContext('2d');if(!context)throw Error('shadow canvas unavailable')
    context.drawImage(image,0,0)
    if(shadow!=='none'&&!context.getImageData(0,0,pixels,pixels).data.some((value,index)=>index%4===3&&value>0))throw Error('native shadow raster is empty')
    states.push(canvas)
   }
   const pixels=states[0]!.width,pixelScale=pixels/geometry.size,cut=Math.round(geometry.slice*pixelScale),middle=pixels-2*cut
   const columns=[{x:0,width:cut},{x:cut,width:middle},{x:cut+middle,width:cut}],rows=columns
   const positions=[[0,0],[1,0],[2,0],[0,1],[2,1],[0,2],[1,2],[2,2]]
   // Each elevation owns its zero-alpha crop. Reusing the larger focused
   // rectangle would leave needless blend pixels in all seven normal windows
   // and every window during a held gesture. Both states still decode before
   // publication and keep the same native source/stretched interval mapping.
   const patches:Patch[][]=states.map(canvas=>positions.map(([column,row])=>{
    const x=columns[column!]!,y=rows[row!]!
    const pixels=canvas.getContext('2d')!.getImageData(x.x,y.x,x.width,y.width).data
    const bounds=shadowPixelBounds(pixels,pixels,x.width,y.width)
    // Preserve the whole source interval on each stretch axis. Only margins
    // perpendicular to that axis can be removed without shifting its mapping.
    if(bounds){if(column===1){bounds.x=0;bounds.width=x.width}if(row===1){bounds.y=0;bounds.height=y.width}}
    return {bounds:bounds||{x:0,y:0,width:1,height:1},sourceWidth:x.width,sourceHeight:y.width,empty:!bounds}
   }))
   for(let state=0;state<states.length;state++){
    const canvas=states[state]!
    for(let index=0;index<positions.length;index++){
     const [column,row]=positions[index]!,x=columns[column!]!,y=rows[row!]!,patch=patches[state]![index]!,b=patch.bounds
     const tile=document.createElement('canvas');tile.width=b.width;tile.height=b.height
     const tileContext=tile.getContext('2d');if(!tileContext)throw Error('shadow patch unavailable')
     if(!patch.empty)tileContext.drawImage(canvas,x.x+b.x,y.x+b.y,b.width,b.height,0,0,b.width,b.height)
     const blob=await new Promise<Blob>((resolve,reject)=>tile.toBlob(value=>value?resolve(value):reject(Error('shadow raster unavailable'))))
     if(disposed||version!==generation)throw Error('superseded shadow raster')
     const url=URL.createObjectURL(blob);urls.push(url)
     const decoded=new Image();decoded.src=url;await decoded.decode()
    }
   }
   if(disposed||version!==generation)throw Error('superseded shadow raster')
   const slice=cut/pixelScale
   const asset={urls,patches,geometry:{...geometry,slice,corner:slice-geometry.pad},scale:pixelScale};assets.add(asset);return asset
  }catch(error){for(const url of urls)URL.revokeObjectURL(url);throw error}
 }
 async function apply(frame:HTMLElement,plane:HTMLElement){
  if(disposed)return
  const ticket=++requestVersion;requests.set(plane,ticket)
  const version=generation,style=getComputedStyle(frame),radius=parseFloat(style.borderTopLeftRadius)||0
  // Resolve tokens through native CSS, including color-mix and all palettes.
  // Resolve outside the frame: inserting measurement nodes into an observed
  // window would retrigger its structural/resize registration during delivery.
  const probe=document.createElement('span');probe.style.cssText='position:absolute;visibility:hidden';probe.style.color=style.color;probe.style.boxShadow=style.getPropertyValue('--window-shadow');document.body.append(probe)
  let normal:string,focused:string
  try{normal=getComputedStyle(probe).boxShadow;probe.style.boxShadow=style.getPropertyValue('--window-shadow-focus');focused=getComputedStyle(probe).boxShadow}finally{probe.remove()}
  const geometry=shadowGeometry(normal,focused,radius)
  // Patches must not overlap. Small/mobile/unsupported contours retain the
  // native shadow instead of distorting the window or changing its layout.
  if(!geometry||frame.offsetWidth<geometry.corner*2+2||frame.offsetHeight<geometry.corner*2+2){clear(plane);return}
  const scale=Math.min(devicePixelRatio||1,2,1024/geometry.size),key=JSON.stringify([normal,focused,radius,scale])
  if(plane.dataset.shadowCache===key)return
  // A breakpoint/new contour immediately falls back to its current native
  // shadow until the matching decoded patches are ready.
  if(plane.dataset.shadowCache)clear(plane)
  if(!cache.has(key)&&cache.size>=8){clear(plane);return}
  if(!cache.has(key))cache.set(key,raster(normal,focused,radius,geometry,scale,version))
  try{
   const asset=await cache.get(key)!
   if(disposed||version!==generation||requests.get(plane)!==ticket||tracked.get(frame)!==plane||!frame.isConnected)return
   // A resize may make this old request inapplicable before decoding finishes.
   if(frame.offsetWidth<asset.geometry.corner*2+2||frame.offsetHeight<asset.geometry.corner*2+2){clear(plane);return}
   const {pad,slice,corner}=asset.geometry
   SHADOW_PATCHES.forEach((name,index)=>{
    const piece=plane.querySelector<HTMLElement>(`.shadow-piece.${name}`)!
    if(stableNativeLayers())piece.style.willChange='transform';else piece.style.removeProperty('will-change')
    piece.style.setProperty('--shadow-normal',`url("${asset.urls[index]}")`);piece.style.setProperty('--shadow-focused',`url("${asset.urls[index+8]}")`)
    for(const [state,stateIndex] of [['normal',0],['focused',1]] as const){
     const patch=asset.patches[stateIndex]![index]!,b=patch.bounds
     const values=[b.width,b.height,b.x,b.y,patch.sourceWidth-b.x-b.width,patch.sourceHeight-b.y-b.height]
     dimensions.forEach((name,i)=>piece.style.setProperty(`--shadow-${state}-${name}`,`${values[i]!/asset.scale}px`))
     piece.style.setProperty(`--shadow-${state}-display`,patch.empty?'none':'block')
    }
    if(asset.patches[0]![index]!.empty&&asset.patches[1]![index]!.empty)piece.dataset.shadowEmpty='true';else delete piece.dataset.shadowEmpty
   })
   plane.style.setProperty('--shadow-width',`${slice}px`);plane.style.setProperty('--shadow-outset',`${pad}px`);plane.style.setProperty('--shadow-corner',`${corner}px`)
   plane.dataset.shadowCache=key
  }catch{if(!disposed&&version===generation&&requests.get(plane)===ticket&&tracked.get(frame)===plane)clear(plane)}
 }
 const sizes=new ResizeObserver(entries=>{for(const entry of entries){const frame=entry.target as HTMLElement,plane=tracked.get(frame);if(plane)void apply(frame,plane)}})
 function scan(){
  for(const [frame,plane] of tracked)if(!frame.isConnected){sizes.unobserve(frame);clear(plane);tracked.delete(frame)}
  for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!plane)continue
   if(!tracked.has(frame)){tracked.set(frame,plane);sizes.observe(frame);void apply(frame,plane)}
  }
 }
 function invalidate(){
  generation++;for(const plane of tracked.values())clear(plane);cache.clear()
  for(const asset of assets)for(const url of asset.urls)URL.revokeObjectURL(url);assets.clear()
  for(const [frame,plane] of tracked)void apply(frame,plane)
 }
 const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
 const preferences=new MutationObserver(invalidate);preferences.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-translucency']})
 const resize=()=>{if(resizeTimer)clearTimeout(resizeTimer);resizeTimer=setTimeout(invalidate,80)}
 window.addEventListener('resize',resize);scan()
 return ()=>{disposed=true;generation++;if(resizeTimer)clearTimeout(resizeTimer);children.disconnect();preferences.disconnect();sizes.disconnect();window.removeEventListener('resize',resize);for(const plane of tracked.values())clear(plane);for(const asset of assets)for(const url of asset.urls)URL.revokeObjectURL(url);assets.clear();tracked.clear();cache.clear()}
}
