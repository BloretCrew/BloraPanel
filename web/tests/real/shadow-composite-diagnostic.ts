// Explicit optical-only diagnostic. Compose the existing decoded native PNG
// perimeter patches into four strips; never read/capture a live app surface.
export async function installShadowCompositeDiagnostic(source:'grouped'|'partitioned'='grouped'){
 type Rect={x:number;y:number;width:number;height:number}
 type Piece=Rect&{url:string}
 type Asset={urls:string[];rects:Rect[]}
 type Record={frame:HTMLElement;plane:HTMLElement;images:HTMLImageElement[];ticket:number;key:string;signature:string;width:number;height:number;radius:number}
 const desktop=document.querySelector<HTMLElement>('.desktop');if(!desktop)return
 const names=['nw','n','ne','w','e','sw','s','se'],groups=[[0,1,2],[3],[4],[5,6,7]]
 const tracked=new Map<HTMLElement,Record>(),cache=new Map<string,Promise<Asset>>(),owned=new Set<string>(),root=document.documentElement,css=document.createElement('style')
 css.textContent='.window-shadow-plane[data-shadow-composite="ready"]>.shadow-piece{visibility:hidden!important}.shadow-composite-image{position:absolute;pointer-events:none;max-width:none}'
 document.head.append(css)
 let disposed=false,paused=false,generation=0
 const clear=(record:Record)=>{delete record.plane.dataset.shadowComposite;for(const image of record.images)image.style.display='none';record.key=''}
 function measure(record:Record){
  const {frame,plane}=record
  if(devicePixelRatio!==1||!frame.isConnected||!plane.dataset.shadowCache)return undefined
  const transform=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(frame.style.transform)
  if(!transform||!transform.slice(1).every(value=>Number.isInteger(Number(value))))return undefined
  const signature=[frame.style.width,frame.style.height,frame.style.borderRadius,innerWidth].join('|')
  if(signature!==record.signature){record.width=plane.offsetWidth;record.height=plane.offsetHeight;record.radius=parseFloat(getComputedStyle(frame).borderTopLeftRadius);record.signature=signature}
  const {width,height,radius}=record,pad=parseFloat(plane.style.getPropertyValue('--shadow-outset')),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner'))
  if(width<=0||height<=0||width>4096||height>4096||![width,height,pad,corner].every(Number.isInteger))return undefined
  const state=frame.classList.contains('focused')&&frame.dataset.shadowLow!=='true'?'focused':'normal',pieces:Piece[]=[]
  for(const name of names){
   const element=plane.querySelector<HTMLElement>('.shadow-piece.'+name);if(!element)return undefined
   const value=(key:string)=>parseFloat(element.style.getPropertyValue(`--shadow-${state}-${key}`))
   const url=/^url\("(blob:[^"]+)"\)$/.exec(element.style.getPropertyValue(`--shadow-${state}`))?.[1]
   if(!url||element.style.getPropertyValue(`--shadow-${state}-display`)==='none')return undefined
   const w=name==='n'||name==='s'?width-2*corner:value('width'),h=name==='w'||name==='e'?height-2*corner:value('height')
   const x=['nw','w','sw'].includes(name)?-pad+value('left'):['ne','e','se'].includes(name)?width+pad-value('right')-w:corner
   const y=['nw','n','ne'].includes(name)?-pad+value('top'):['sw','s','se'].includes(name)?height+pad-value('bottom')-h:corner
   if(![x,y,w,h].every(Number.isInteger)||w<=0||h<=0)return undefined
   pieces.push({x,y,width:w,height:h,url})
  }
  return {pieces,width,height,radius,key:JSON.stringify([source,width,height,radius,state,pieces])}
 }
 async function compile(value:{pieces:Piece[];width:number;height:number;radius:number},version:number):Promise<Asset>{
  const {pieces,width,height,radius}=value
  const urls:string[]=[],rects:Rect[]=[]
  try{
   const sources=await Promise.all(pieces.map(async piece=>{const image=new Image();image.src=piece.url;await image.decode();return image}))
   let perimeter:HTMLCanvasElement|undefined,partitions:Rect[]=[],origin={x:0,y:0}
   if(source==='partitioned'){
    if(!Number.isInteger(radius)||radius<0||radius>128)throw Error('unsupported native perimeter contour')
    const x=Math.min(...pieces.map(p=>p.x)),y=Math.min(...pieces.map(p=>p.y)),right=Math.max(...pieces.map(p=>p.x+p.width)),bottom=Math.max(...pieces.map(p=>p.y+p.height))
    origin={x,y};perimeter=document.createElement('canvas');perimeter.width=right-x;perimeter.height=bottom-y
    const context=perimeter.getContext('2d');if(!context)throw Error('native perimeter surface unavailable')
    pieces.forEach((p,index)=>context.drawImage(sources[index]!,p.x-x,p.y-y,p.width,p.height))
    const l=radius-x,t=radius-y,r=width-radius-x,b=height-radius-y
    if(l<0||t<0||r<=l||b<=t||r>perimeter.width||b>perimeter.height)throw Error('unsupported native perimeter partition')
    // Certify the removed centre from optical alpha, not from application
    // opacity or a guessed contour. Keep every nonzero shadow pixel.
    const middle=context.getImageData(l,t,r-l,b-t).data
    for(let i=3;i<middle.length;i+=4)if(middle[i])throw Error('native shadow centre is not empty')
    partitions=[{x:0,y:0,width:perimeter.width,height:t},{x:0,y:b,width:perimeter.width,height:perimeter.height-b},{x:0,y:t,width:l,height:b-t},{x:r,y:t,width:perimeter.width-r,height:b-t}]
   }
   for(let groupIndex=0;groupIndex<groups.length;groupIndex++){
    const group=groups[groupIndex]!
    const members=group.map(index=>pieces[index]!),x=Math.min(...members.map(p=>p.x)),y=Math.min(...members.map(p=>p.y)),right=Math.max(...members.map(p=>p.x+p.width)),bottom=Math.max(...members.map(p=>p.y+p.height))
    const partition=partitions[groupIndex],canvas=document.createElement('canvas');canvas.width=partition?partition.width:right-x;canvas.height=partition?partition.height:bottom-y
    const context=canvas.getContext('2d');if(!context)throw Error('native perimeter composition unavailable')
    if(perimeter&&partition)context.drawImage(perimeter,partition.x,partition.y,partition.width,partition.height,0,0,partition.width,partition.height)
    else for(const index of group){const p=pieces[index]!;context.drawImage(sources[index]!,p.x-x,p.y-y,p.width,p.height)}
    const pixels=context.getImageData(0,0,canvas.width,canvas.height).data
    let left=canvas.width,top=canvas.height,endX=0,endY=0
    for(let py=0;py<canvas.height;py++)for(let px=0;px<canvas.width;px++)if(pixels[(py*canvas.width+px)*4+3]){left=Math.min(left,px);top=Math.min(top,py);endX=Math.max(endX,px+1);endY=Math.max(endY,py+1)}
    if(endX<=left||endY<=top)throw Error('empty native perimeter strip')
    const tile=document.createElement('canvas');tile.width=endX-left;tile.height=endY-top
    const target=tile.getContext('2d');if(!target)throw Error('native perimeter strip unavailable')
    target.drawImage(canvas,left,top,tile.width,tile.height,0,0,tile.width,tile.height)
    const blob=await new Promise<Blob>((resolve,reject)=>tile.toBlob(blob=>blob?resolve(blob):reject(Error('native perimeter encoding unavailable'))))
    if(disposed||generation!==version)throw Error('superseded native perimeter')
    const url=URL.createObjectURL(blob);urls.push(url);const decoded=new Image();decoded.src=url;await decoded.decode()
    rects.push({x:(partition?origin.x+partition.x:x)+left,y:(partition?origin.y+partition.y:y)+top,width:tile.width,height:tile.height})
   }
   if(disposed||generation!==version)throw Error('superseded native perimeter')
   for(const url of urls)owned.add(url)
   return {urls,rects}
  }catch(error){for(const url of urls)URL.revokeObjectURL(url);throw error}
 }
 async function update(record:Record){
  if(disposed||paused)return
  const value=measure(record);if(!value){clear(record);return}if(record.key===value.key)return
  clear(record);const ticket=++record.ticket,version=generation
  if(!cache.has(value.key)&&cache.size>=32)return
  if(!cache.has(value.key))cache.set(value.key,compile(value,version))
  try{
   const asset=await cache.get(value.key)!
   if(disposed||paused||generation!==version||ticket!==record.ticket||measure(record)?.key!==value.key)return
   asset.rects.forEach((rect,index)=>{const image=record.images[index]!;image.src=asset.urls[index]!;image.style.left=rect.x+'px';image.style.top=rect.y+'px';image.style.width=rect.width+'px';image.style.height=rect.height+'px';image.style.display='block'})
   record.key=value.key;record.plane.dataset.shadowComposite='ready'
  }catch{if(!disposed&&ticket===record.ticket)clear(record)}
 }
 const mutations=new MutationObserver(records=>{for(const frame of new Set(records.map(r=>(r.target as HTMLElement).closest<HTMLElement>('.app-window')).filter((frame):frame is HTMLElement=>!!frame))){const record=tracked.get(frame);if(record)void update(record)}})
 const sizes=new ResizeObserver(entries=>{for(const entry of entries){const record=tracked.get(entry.target as HTMLElement);if(record){record.signature='';clear(record);void update(record)}}})
 for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
  const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!plane)continue
  const images=groups.map(()=>{const image=new Image();image.alt='';image.draggable=false;image.className='shadow-composite-image';image.style.display='none';plane.append(image);return image})
  const record={frame,plane,images,ticket:0,key:'',signature:'',width:0,height:0,radius:0};tracked.set(frame,record);sizes.observe(frame)
  mutations.observe(frame,{attributes:true,attributeFilter:['style','class','data-shadow-low']});mutations.observe(plane,{attributes:true,attributeFilter:['data-shadow-cache']})
 }
 function invalidate(){generation++;for(const record of tracked.values()){record.ticket++;clear(record)};cache.clear();for(const url of owned)URL.revokeObjectURL(url);owned.clear();for(const record of tracked.values())void update(record)}
 const appearance=new MutationObserver(invalidate);appearance.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-translucency']})
 const controller={pause(){paused=true;for(const record of tracked.values())clear(record)},resume(){paused=false;for(const record of tracked.values())void update(record)},dispose(){disposed=true;generation++;mutations.disconnect();sizes.disconnect();appearance.disconnect();for(const record of tracked.values()){record.ticket++;clear(record);for(const image of record.images)image.remove()};for(const url of owned)URL.revokeObjectURL(url);owned.clear();cache.clear();css.remove()}}
 ;(window as unknown as {__shadowCompositeDiagnostic:typeof controller}).__shadowCompositeDiagnostic=controller
 window.addEventListener('pagehide',controller.dispose,{once:true})
 await Promise.all([...tracked.values()].map(update))
}
