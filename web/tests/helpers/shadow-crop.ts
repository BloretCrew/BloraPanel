import type {PaintRect} from '../../src/desktop/occlusion'

// Explicit native optical prototype. Each unchanged original image quad is
// retained inside a simple overflow viewport; no vector mask or app capture.
export function installShadowCropDiagnostic(){
 // Self-contained so the exact same installer can run through the native
 // browser protocol against the real production bundle, without a dev server.
 function exposedBounds(rect:PaintRect,covers:PaintRect[]):PaintRect|undefined{
  let pieces=[rect]
  for(const cover of covers){const next:PaintRect[]=[];for(const p of pieces){
   const x=Math.max(p.x,cover.x),y=Math.max(p.y,cover.y),right=Math.min(p.x+p.width,cover.x+cover.width),bottom=Math.min(p.y+p.height,cover.y+cover.height)
   if(x>=right||y>=bottom){next.push(p);continue}
   if(y>p.y)next.push({x:p.x,y:p.y,width:p.width,height:y-p.y})
   if(bottom<p.y+p.height)next.push({x:p.x,y:bottom,width:p.width,height:p.y+p.height-bottom})
   if(x>p.x)next.push({x:p.x,y,width:x-p.x,height:bottom-y})
   if(right<p.x+p.width)next.push({x:right,y,width:p.x+p.width-right,height:bottom-y})
  }pieces=next;if(pieces.length>128)return rect}
  if(!pieces.length)return undefined
  const x=Math.min(...pieces.map(p=>p.x)),y=Math.min(...pieces.map(p=>p.y)),right=Math.max(...pieces.map(p=>p.x+p.width)),bottom=Math.max(...pieces.map(p=>p.y+p.height));return {x,y,width:right-x,height:bottom-y}
 }
 const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!
 const entries=new Map<HTMLElement,{viewport:HTMLElement;image:HTMLElement;bounds?:PaintRect;state:string}>()
 type Metrics={signature:string;fast:boolean;position?:{x:number;y:number};frame:HTMLElement;style:{display:string;zIndex:string;visibility:string;opacity:string;borderTopLeftRadius:string};box:PaintRect;pieces:{piece:HTMLElement;box:PaintRect;cached:boolean}[]}
 const metrics=new Map<HTMLElement,Metrics>()
 const asRect=(box:DOMRect):PaintRect=>({x:box.x,y:box.y,width:box.width,height:box.height})
 function measure(frame:HTMLElement):Metrics{
  const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane')
  const signature=[...frame.style].filter(key=>key!=='transform'&&!key.startsWith('--material-')).map(key=>key+':'+frame.style.getPropertyValue(key)).join(';')+'|'+[...frame.classList].filter(key=>key!=='moving').join(' ')+'|'+frame.dataset.shadowLow+'|'+plane?.dataset.shadowCache
  const match=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(frame.style.transform),position=match?{x:Number(match[1]),y:Number(match[2])}:undefined,old=metrics.get(frame)
  if(old?.fast&&old.position&&position&&old.signature===signature){const dx=position.x-old.position.x,dy=position.y-old.position.y;old.box.x+=dx;old.box.y+=dy;for(const p of old.pieces){p.box.x+=dx;p.box.y+=dy}old.position=position;return old}
  const native=getComputedStyle(frame),matrix=new DOMMatrixReadOnly(native.transform==='none'?undefined:native.transform)
  const next:Metrics={signature,position,fast:!!position&&matrix.is2D&&matrix.a===1&&matrix.b===0&&matrix.c===0&&matrix.d===1&&Math.abs(matrix.e-position.x)<.01&&Math.abs(matrix.f-position.y)<.01,frame,style:{display:native.display,zIndex:native.zIndex,visibility:native.visibility,opacity:native.opacity,borderTopLeftRadius:native.borderTopLeftRadius},box:asRect(frame.getBoundingClientRect()),pieces:[...frame.querySelectorAll<HTMLElement>(':scope>.window-shadow-plane>.shadow-piece')].map(piece=>({piece,box:asRect(piece.getBoundingClientRect()),cached:!!piece.parentElement?.dataset.shadowCache}))}
  metrics.set(frame,next);return next
 }
 const frames=new MutationObserver(queue),planes=new MutationObserver(queue)
 let pending=0,paused=false,disposed=false
 const style=document.createElement('style');style.textContent='.shadow-piece[data-native-shadow-crop]{background-image:none!important}.native-shadow-crop{position:absolute;display:block;overflow:hidden}.native-shadow-crop>i{position:absolute;display:block;background-image:var(--shadow-normal);background-size:100% 100%;background-repeat:no-repeat}.app-window.focused>.window-shadow-plane>.shadow-piece>.native-shadow-crop>i{background-image:var(--shadow-focused)}.app-window[data-shadow-low="true"]>.window-shadow-plane>.shadow-piece>.native-shadow-crop>i{background-image:var(--shadow-normal)}';document.head.append(style)
 const set=(element:HTMLElement,key:string,value:string)=>{if(element.style.getPropertyValue(key)!==value)element.style.setProperty(key,value)}
 const remove=(piece:HTMLElement)=>{entries.get(piece)?.viewport.remove();entries.delete(piece);delete piece.dataset.nativeShadowCrop}
 function apply(){
  pending=0;if(disposed||paused)return
  const opaque=root.dataset.translucency==='off'||(root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true')
  const rows=[...desktop.querySelectorAll<HTMLElement>(':scope>.app-window')].map(measure).filter(row=>row.style.display!=='none').sort((a,b)=>Number(b.style.zIndex)-Number(a.style.zIndex))
  const moving=rows.some(row=>row.frame.classList.contains('moving')),covers:PaintRect[]=[]
  // Complete native reads before adding/mutating viewports.
  for(const {frame,style:appearance,box,pieces} of rows){
   const state=frame.classList.contains('focused')&&!frame.dataset.shadowLow?'focused':'normal'
   for(const {piece,box:area,cached} of pieces){
    if(!opaque||!cached||devicePixelRatio!==1||!area.width||!area.height){remove(piece);continue}
    const visible=exposedBounds(area,covers)
    if(!visible){remove(piece);continue} // Product already clips fully hidden patches.
    const margin=moving?34:2
    let bounds={x:Math.max(-2,Math.floor(visible.x-area.x)-margin),y:Math.max(-2,Math.floor(visible.y-area.y)-margin),width:0,height:0}
    let right=Math.min(area.width+2,Math.ceil(visible.x+visible.width-area.x)+margin),bottom=Math.min(area.height+2,Math.ceil(visible.y+visible.height-area.y)+margin)
    const old=entries.get(piece)
    if(moving&&old?.bounds&&old.state===state){right=Math.max(right,old.bounds.x+old.bounds.width);bottom=Math.max(bottom,old.bounds.y+old.bounds.height);bounds.x=Math.min(bounds.x,old.bounds.x);bounds.y=Math.min(bounds.y,old.bounds.y)}
    bounds.width=right-bounds.x;bounds.height=bottom-bounds.y
    if(bounds.width>=area.width+3.99&&bounds.height>=area.height+3.99){remove(piece);continue}
    let entry=old
    if(!entry){const viewport=document.createElement('i'),image=document.createElement('i');viewport.className='native-shadow-crop';viewport.append(image);piece.append(viewport);piece.dataset.nativeShadowCrop='true';entry={viewport,image,state};entries.set(piece,entry)}
    entry.state=state;entry.bounds=moving?bounds:undefined
    for(const [key,value] of Object.entries({left:bounds.x,top:bounds.y,width:bounds.width,height:bounds.height}))set(entry.viewport,key,`${value}px`)
    for(const [key,value] of Object.entries({left:-bounds.x,top:-bounds.y,width:area.width,height:area.height}))set(entry.image,key,`${value}px`)
   }
   const radius=Math.max(1,Math.min(parseFloat(appearance.borderTopLeftRadius)||0,box.width/2,box.height/2))
   if(opaque&&appearance.visibility==='visible'&&Number(appearance.opacity)===1){
    if(box.width>radius*2&&box.height>2)covers.push({x:box.x+radius,y:box.y+1,width:box.width-radius*2,height:box.height-2})
    if(box.width>2&&box.height>radius*2)covers.push({x:box.x+1,y:box.y+radius,width:box.width-2,height:box.height-radius*2})
   }
  }
  for(const piece of entries.keys())if(!piece.isConnected)remove(piece)
 }
 function queue(){if(!disposed&&!paused&&!pending)pending=requestAnimationFrame(apply)}
 function scan(){frames.disconnect();planes.disconnect();metrics.clear();for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){frames.observe(frame,{attributes:true,attributeFilter:['style','class','data-shadow-low']});const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(plane)planes.observe(plane,{attributes:true,attributeFilter:['data-shadow-cache']})}queue()}
 const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
 const invalidate=()=>{metrics.clear();queue()}
 const preferences=new MutationObserver(invalidate);preferences.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-material-cache-opaque','data-translucency','data-theme','data-palette']})
 const pause=()=>{paused=true;if(pending)cancelAnimationFrame(pending);pending=0;for(const piece of entries.keys())remove(piece)}
 const resume=()=>{paused=false;apply()}
 const dispose=()=>{disposed=true;pause();metrics.clear();frames.disconnect();planes.disconnect();children.disconnect();preferences.disconnect();window.removeEventListener('resize',invalidate);style.remove()}
 window.addEventListener('resize',invalidate);window.addEventListener('pagehide',dispose,{once:true});scan()
 return {pause,resume,dispose}
}
