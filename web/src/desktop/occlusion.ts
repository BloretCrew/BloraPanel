export type PaintRect={x:number;y:number;width:number;height:number}
import {SHADOW_PATCHES} from '../appearance/shadow-cache'
// Only native gesture translations may share a pre-paint clipping update.
// Unmarked application/layout writes still reveal their content immediately,
// including writes made inside an animation-frame callback.
const compositorTranslations=new WeakSet<HTMLElement>()
export function markCompositorTranslation(frame:HTMLElement){compositorTranslations.add(frame)}
// Subtract opaque rectangles, then retain a single conservative bounding box.
// Keeping extra pixels is safe; cutting a potentially visible pixel is not.
export function exposedRegions(rect:PaintRect,covers:PaintRect[]):PaintRect[]{
 let pieces=[rect]
 for(const cover of covers){
  const next:PaintRect[]=[]
  for(const p of pieces){
   const x=Math.max(p.x,cover.x),y=Math.max(p.y,cover.y),right=Math.min(p.x+p.width,cover.x+cover.width),bottom=Math.min(p.y+p.height,cover.y+cover.height)
   if(x>=right||y>=bottom){next.push(p);continue}
   if(y>p.y)next.push({x:p.x,y:p.y,width:p.width,height:y-p.y})
   if(bottom<p.y+p.height)next.push({x:p.x,y:bottom,width:p.width,height:p.y+p.height-bottom})
   if(x>p.x)next.push({x:p.x,y,width:x-p.x,height:bottom-y})
   if(right<p.x+p.width)next.push({x:right,y,width:p.x+p.width-right,height:bottom-y})
  }
  pieces=next
  // Pathological arrangements must not create unbounded geometry work.
  if(pieces.length>128)return [rect]
 }
 return pieces
}
export function exposedBounds(rect:PaintRect,covers:PaintRect[]):PaintRect|undefined{
 if(!covers.length)return rect
 const pieces=exposedRegions(rect,covers)
 if(!pieces.length)return undefined
 let x=Infinity,y=Infinity,right=-Infinity,bottom=-Infinity
 for(const p of pieces){x=Math.min(x,p.x);y=Math.min(y,p.y);right=Math.max(right,p.x+p.width);bottom=Math.max(bottom,p.y+p.height)}
 return {x,y,width:right-x,height:bottom-y}
}

// Retain the exact-strips geometry contracts/regressions for future renderer
// work. The installed product uses conservative bounds below: native vector
// masks were rejected by the whole-screen pixel preservation regression.
export function exposureClip(rect:PaintRect,regions:PaintRect[]):string{
 if(!regions.length)return 'inset(100%)'
 if(regions.length===1){
  const visible=regions[0]!,top=Math.max(0,visible.y-rect.y),left=Math.max(0,visible.x-rect.x),bottom=Math.max(0,rect.y+rect.height-visible.y-visible.height),right=Math.max(0,rect.x+rect.width-visible.x-visible.width)
  return top+left+bottom+right<.01?'':`inset(${top}px ${right}px ${bottom}px ${left}px)`
 }
 // Same-winding closed rectangles produce a union, including overlap in the
 // conservative gesture envelope. No application contents enter this mask.
 return `path('${regions.map(r=>`M${r.x-rect.x} ${r.y-rect.y}h${r.width}v${r.height}h${-r.width}Z`).join(' ')}')`
}
export function retainExposure(rect:PaintRect,previous:PaintRect[]|undefined,regions:PaintRect[]):PaintRect[]{
 let retained=previous?[...previous]:[]
 const contains=(outer:PaintRect,inner:PaintRect)=>outer.x<=inner.x&&outer.y<=inner.y&&outer.x+outer.width>=inner.x+inner.width&&outer.y+outer.height>=inner.y+inner.height
 for(const visible of regions){
  const x=Math.max(0,visible.x-rect.x-64),y=Math.max(0,visible.y-rect.y-64),right=Math.min(rect.width,visible.x+visible.width-rect.x+64),bottom=Math.min(rect.height,visible.y+visible.height-rect.y+64)
  const expanded={x,y,width:right-x,height:bottom-y}
  if(retained.some(r=>contains(r,expanded)))continue
  retained=retained.filter(r=>!contains(expanded,r));retained.push(expanded)
  // A long or pathological gesture falls back to painting the complete body.
  // Never allow mask complexity or accumulated exposure to grow unbounded.
  if(retained.length>32)return [{x:0,y:0,width:rect.width,height:rect.height}]
 }
 return retained
}

// A gesture's extra paint is a reserve, not an edge to move by one pixel on
// every event. Reuse it while it already contains the newly visible region;
// grow only at first reveal beyond that reserve. No exposed pixel waits.
export function retainExposureBounds(rect:PaintRect,previous:PaintRect|undefined,visible:PaintRect,margin=64):PaintRect{
 const x=Math.max(0,visible.x-rect.x),y=Math.max(0,visible.y-rect.y)
 const right=Math.min(rect.width,visible.x+visible.width-rect.x),bottom=Math.min(rect.height,visible.y+visible.height-rect.y)
 if(previous&&(previous.x<0||previous.y<0||previous.x+previous.width>rect.width||previous.y+previous.height>rect.height))previous=undefined
 if(previous&&previous.x<=x&&previous.y<=y&&previous.x+previous.width>=right&&previous.y+previous.height>=bottom)return previous
 const left=Math.max(0,x-margin),top=Math.max(0,y-margin),end=Math.min(rect.width,right+margin),down=Math.min(rect.height,bottom+margin)
 const retainedX=previous?Math.min(previous.x,left):left,retainedY=previous?Math.min(previous.y,top):top
 return {x:retainedX,y:retainedY,width:(previous?Math.max(previous.x+previous.width,end):end)-retainedX,height:(previous?Math.max(previous.y+previous.height,down):down)-retainedY}
}

export function installWindowOcclusion(desktop:HTMLElement){
 type Shadow={element:HTMLElement;normal?:PaintRect;focused?:PaintRect}
 type Metrics={rect:PaintRect;bodyRect:PaintRect;shadows:Shadow[];terminals:{element:HTMLElement;rect:PaintRect;viewport?:HTMLElement;width:string;height:string}[];z:number;display:string;radius:number;solid:boolean}
 type Tracked={frame:HTMLElement;body:HTMLElement;hosts:HTMLElement[];dirty:boolean;moving:boolean;metrics?:Metrics;exposure?:PaintRect;terminalExposure?:Map<HTMLElement,PaintRect>;shadowExposure?:Map<HTMLElement,{state:string;visible:boolean}>;translation?:{x:number;y:number};fast:boolean;transform:string;atlasX:string;atlasY:string;style:string;className:string;structure:MutationObserver;views:MutationObserver;sizes:ResizeObserver}
 const root=document.documentElement,tracked=new Map<HTMLElement,Tracked>()
 const clippedTerminals=new Set<HTMLElement>()
 let animation=0,disposed=false
 const rect=(value:DOMRect):PaintRect=>({x:value.x,y:value.y,width:value.width,height:value.height})
 const translation=(value:string)=>{const match=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(value);return match?{x:Number(match[1]),y:Number(match[2])}:undefined}
 // Window enter/leave classes animate only child chrome and the shadow plane;
 // they never change the frame/body geometry or its opaque paint. Ignore those
 // presentation-only mutations so they do not trigger an unnecessary layout
 // read while the desktop compositor is measuring a held translation.
 const className=(frame:HTMLElement)=>[...frame.classList].filter(name=>name!=='moving'&&!/^window-(?:enter|leave)-(?:from|to|active)$/.test(name)).join(' ')
 const authoredStyle=(frame:HTMLElement)=>{
  const position=translation(frame.style.transform),offsets=position&&frame.dataset.windowPosition==='offset'&&parseFloat(frame.style.getPropertyValue('--window-x'))===position.x&&parseFloat(frame.style.getPropertyValue('--window-y'))===position.y
  return [...frame.style].filter(name=>name!=='transform'&&!name.startsWith('--material-')&&!(offsets&&(name==='--window-x'||name==='--window-y'))).map(name=>`${name}:${frame.style.getPropertyValue(name)}!${frame.style.getPropertyPriority(name)}`).join(';')
 }
 const set=(element:HTMLElement,name:string,value:string)=>{if(element.style.getPropertyValue(name)!==value){if(value)element.style.setProperty(name,value);else element.style.removeProperty(name)}}
 function releaseTerminal(host:HTMLElement){if(clippedTerminals.delete(host)){for(const name of ['clip-path','width','height','left','top'])host.style.removeProperty(name);const viewport=host.closest<HTMLElement>('.terminal-paint-viewport');if(viewport)for(const name of ['width','height','left','top'])viewport.style.removeProperty(name);delete host.dataset.paintOccluded;delete host.dataset.paintPartial}}
 function release(record:Tracked){compositorTranslations.delete(record.frame);record.structure.disconnect();record.views.disconnect();record.sizes.disconnect();for(const host of record.hosts)releaseTerminal(host);for(const shadow of record.metrics?.shadows||[])for(const name of ['clip-path','clip'])shadow.element.style.removeProperty(name);delete record.frame.dataset.shadowLow;for(const name of ['--material-header-size','--material-frame-radius'])record.frame.style.removeProperty(name);for(const name of ['clip-path','--material-x','--material-y'])record.body.style.removeProperty(name)}
 function register(record:Tracked){
  record.structure.disconnect();record.views.disconnect();record.sizes.disconnect()
  const hosts=[...record.body.querySelectorAll<HTMLElement>('.terminal-container')]
  for(const old of record.hosts)if(!hosts.includes(old))releaseTerminal(old)
  record.hosts=hosts;record.dirty=true
  const plane=record.frame.querySelector<HTMLElement>(':scope>.window-shadow-plane')
  if(plane)record.views.observe(plane,{attributes:true,attributeFilter:['data-shadow-cache']})
  const structural=new Set<HTMLElement>([record.frame,record.body]),layout=new Set<HTMLElement>([record.frame,record.body])
  for(const view of record.body.querySelectorAll<HTMLElement>(':scope>.app-view')){
   structural.add(view);layout.add(view)
   record.views.observe(view,{attributes:true,attributeFilter:['style','class','data-active-view']})
  }
  for(const host of hosts){
   layout.add(host)
   // Observe the host's layout ancestors and immediate peers, never terminal
   // rows. A wrapping toolbar, inserted notice or tab activation can move a
   // host without changing the outer window's geometry.
   for(let parent=host.parentElement;parent&&parent!==record.frame;parent=parent.parentElement){
    structural.add(parent);layout.add(parent)
    for(const child of parent.children)if(child instanceof HTMLElement)layout.add(child)
   }
  }
  // Native lists/text remain in their original paint ownership. The rejected
  // per-block clipping prototype is isolated in tests/helpers/native-list.ts;
  // it saved little area while adding observers and native compositing masks.
  for(const element of structural)record.structure.observe(element,{childList:true})
  // This internal paint viewport is owned here and deliberately resized by
  // clipping. Its full region/host and layout ancestors remain observed;
  // reacting to our own viewport sizes would discard the gesture envelope.
  for(const element of layout)if(!element.classList.contains('terminal-paint-viewport'))record.sizes.observe(element)
 }
 function create(frame:HTMLElement,body:HTMLElement):Tracked{
  const record={} as Tracked
  Object.assign(record,{frame,body,hosts:[],dirty:true,moving:false,fast:false,transform:'',atlasX:'',atlasY:'',style:'',className:'',
   structure:new MutationObserver(()=>{if(!disposed){register(record);update()}}),
   views:new MutationObserver(()=>{if(!disposed){record.dirty=true;update()}}),
   sizes:new ResizeObserver(()=>{if(!disposed){record.dirty=true;update()}})})
  register(record)
  return record
 }
 function advance(record:Tracked,next:{x:number;y:number},transform:string){
  const dx=next.x-record.translation!.x,dy=next.y-record.translation!.y,move=(r:PaintRect)=>{r.x+=dx;r.y+=dy}
  move(record.metrics!.rect);move(record.metrics!.bodyRect)
  for(const shadow of record.metrics!.shadows){if(shadow.normal)move(shadow.normal);if(shadow.focused)move(shadow.focused)}
  for(const terminal of record.metrics!.terminals)move(terminal.rect)
  record.transform=transform;record.translation=next
 }
 function attributeChanges(records:MutationRecord[]){
  if(disposed)return
  let changed=false,canDefer=true
  for(const frame of new Set(records.map(record=>record.target as HTMLElement))){
   const record=tracked.get(frame);if(!record)continue
   const owned=compositorTranslations.delete(frame)
   const position=translation(frame.style.transform)
   // A direct authored transform remains a real native geometry write. Keep
   // its offset aliases in sync only when they still hold the old position;
   // independently changed aliases remain layout invalidations, not gestures.
   if(!owned&&position&&record.translation&&frame.dataset.windowPosition==='offset'&&frame.style.transform!==record.transform&&parseFloat(frame.style.getPropertyValue('--window-x'))===record.translation.x&&parseFloat(frame.style.getPropertyValue('--window-y'))===record.translation.y){
    frame.style.setProperty('--window-x',position.x+'px');frame.style.setProperty('--window-y',position.y+'px')
   }
   const style=authoredStyle(frame),classes=className(frame),transform=frame.style.transform,next=translation(transform)
   const atlasX=frame.style.getPropertyValue('--material-x'),atlasY=frame.style.getPropertyValue('--material-y')
   const moving=frame.classList.contains('moving')
   // Our optical variables and unchanged Vue style writes do not invalidate
   // layout. Pure translation reuses local body/host rectangles and z/radius.
   if(!record.dirty&&style===record.style&&classes===record.className&&transform===record.transform){
    // Committing a held drag updates the wallpaper sample without changing
    // the already drawn transform. Header and body must settle together.
    if(atlasX!==record.atlasX||atlasY!==record.atlasY){record.atlasX=atlasX;record.atlasY=atlasY;changed=true;canDefer=false}
    if(moving!==record.moving){record.moving=moving;changed=true;canDefer=false}
    continue
   }
   if(!record.dirty&&record.fast&&record.metrics&&next&&record.translation&&style===record.style&&classes===record.className){
    canDefer=canDefer&&owned&&moving===record.moving&&atlasX===record.atlasX&&atlasY===record.atlasY
    advance(record,next,transform)
    record.atlasX=atlasX;record.atlasY=atlasY
    record.moving=moving
   }else{record.dirty=true;canDefer=false}
   changed=true
  }
  if(changed){if(canDefer)schedule();else update()}
 }
 const attributes=new MutationObserver(attributeChanges)
 function scan(){
  attributes.disconnect()
  for(const frame of tracked.keys())compositorTranslations.delete(frame)
  for(const [frame,record] of tracked)if(!frame.isConnected){release(record);tracked.delete(frame)}
  // A structural desktop change can arrive in the same Vue patch as geometry
  // writes. Disconnecting the attribute observer discards pending records, so
  // remeasure this infrequent registration boundary conservatively.
  for(const record of tracked.values())record.dirty=true
  for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
   const body=frame.querySelector<HTMLElement>(':scope>.window-body')
   if(body&&!tracked.has(frame))tracked.set(frame,create(frame,body))
   attributes.observe(frame,{attributes:true,attributeFilter:['style','class']})
  }
  update()
 }
 function schedule(){if(!disposed&&!animation)animation=requestAnimationFrame(()=>{animation=0;update()})}
 function invalidate(){for(const record of tracked.values())record.dirty=true;schedule()}
 function measure(record:Tracked){
  if(!record.dirty&&record.metrics)return record.metrics
  record.terminalExposure=undefined
  const style=getComputedStyle(record.frame),box=record.frame.getBoundingClientRect()
  record.style=authoredStyle(record.frame);record.className=className(record.frame);record.transform=record.frame.style.transform;record.translation=translation(record.transform)
  record.atlasX=record.frame.style.getPropertyValue('--material-x');record.atlasY=record.frame.style.getPropertyValue('--material-y')
  record.moving=record.frame.classList.contains('moving')
  const matrix=new DOMMatrixReadOnly(style.transform==='none'?undefined:style.transform),position=record.translation
  // Media rules (including the narrow-screen !important transform) and other
  // non-translation transforms cannot use authored-position arithmetic.
  const offset=!!position&&record.frame.dataset.windowPosition==='offset'&&matrix.e===0&&matrix.f===0&&Math.abs(parseFloat(style.left)-position.x)<.01&&Math.abs(parseFloat(style.top)-position.y)<.01
  record.fast=!!position&&matrix.is2D&&matrix.a===1&&matrix.b===0&&matrix.c===0&&matrix.d===1&&(offset||Math.abs(matrix.e-position.x)<.01&&Math.abs(matrix.f-position.y)<.01)
  const plane=record.frame.querySelector<HTMLElement>(':scope>.window-shadow-plane'),shadows:Shadow[]=[]
  if(plane){
   const cached=!!plane.dataset.shadowCache,pad=parseFloat(plane.style.getPropertyValue('--shadow-outset')),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner'))
   const area=cached?plane.getBoundingClientRect():undefined
   for(const name of SHADOW_PATCHES){
    const element=plane.querySelector<HTMLElement>('.shadow-piece.'+name);if(!element)continue
    const shadow:Shadow={element}
    if(area&&Number.isFinite(pad)&&Number.isFinite(corner))for(const state of ['normal','focused'] as const){
     if(element.style.getPropertyValue(`--shadow-${state}-display`)==='none')continue
     const value=(key:string)=>parseFloat(element.style.getPropertyValue(`--shadow-${state}-${key}`))
     const width=name==='n'||name==='s'?area.width-2*corner:value('width'),height=name==='w'||name==='e'?area.height-2*corner:value('height')
     const x=['nw','w','sw'].includes(name)?-pad+value('left'):['ne','e','se'].includes(name)?area.width+pad-value('right')-width:corner
     const y=['nw','n','ne'].includes(name)?-pad+value('top'):['sw','s','se'].includes(name)?area.height+pad-value('bottom')-height:corner
     if([x,y,width,height].every(Number.isFinite)&&width>0&&height>0)shadow[state]={x:area.x+x,y:area.y+y,width,height}
    }
    shadows.push(shadow)
   }
  }
  record.metrics={rect:rect(box),bodyRect:rect(record.body.getBoundingClientRect()),shadows,terminals:record.hosts.map(element=>{
   const region=element.closest<HTMLElement>('.terminal-paint-region'),viewport=element.closest<HTMLElement>('.terminal-paint-viewport')||undefined,layout=region||element,sizing=getComputedStyle(layout)
   return {element,rect:rect(layout.getBoundingClientRect()),viewport,width:sizing.width,height:sizing.height}
  }),z:Number(style.zIndex)||0,display:style.display,radius:parseFloat(style.borderTopLeftRadius)||0,solid:style.visibility==='visible'&&Number(style.opacity)===1}
  record.dirty=false;return record.metrics
 }
 function update(){
  if(disposed)return
  // A focus/resize/structural invalidation supersedes a queued translation.
  // The cached geometry already includes every intervening pointer position.
  if(animation){cancelAnimationFrame(animation);animation=0}
  const opaque=root.dataset.translucency==='off'||(root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true')
  // Fractional native pixels can change raster ownership/edge AA when these
  // application, shadow or terminal clips are introduced. Keep their original
  // full native paint there: identical boxes alone do not prove identical AA.
  const exactPaintClipping=Number.isInteger(devicePixelRatio||1)
  // Finish geometry/style reads before writing any clipping or optical state.
  // Interleaving getComputedStyle with writes makes some engines repeatedly
  // rebuild the styles of application/terminal descendants in the same frame.
  const windows=[...tracked.values()].map(record=>({record,frame:record.frame,body:record.body,...measure(record)})).filter(window=>window.display!=='none').sort((a,b)=>b.z-a.z)
  const moving=windows.some(window=>window.frame.classList.contains('moving'))
  const anyMoving=[...tracked.keys()].some(frame=>frame.classList.contains('moving'))
  const covers:PaintRect[]=[]
  for(const window of windows){
   // Preserve gesture elevation without an ancestor :has selector that makes
   // every application's descendants restyle when a gesture starts/ends.
   if(anyMoving&&window.frame.classList.contains('focused')){if(window.frame.dataset.shadowLow!=='true')window.frame.dataset.shadowLow='true'}else delete window.frame.dataset.shadowLow
   // Bound each unchanged native shadow patch by the pixels actually exposed
   // above it. Rectangular native clips avoid vector masks and never sample
   // application contents. Geometry stays cached across pure translations.
   const elevation=window.frame.classList.contains('focused')&&!anyMoving?'focused':'normal'
   if(!moving)window.record.shadowExposure=undefined
   for(const shadow of window.shadows){
    const area=shadow[elevation]
    if(!opaque||!area||!exactPaintClipping){set(shadow.element,'clip-path','');set(shadow.element,'clip','');continue}
    let visible:boolean
    if(moving){
     const retained=window.record.shadowExposure||=new Map(),old=retained.get(shadow.element)
     // A patch is always painted whole once any of it has been exposed in
     // this gesture. Its retained rectangle never affected native paint:
     // keep that decision, not repeated subtraction/allocation on every move.
     // Still-hidden patches are checked every time so first reveal is timely.
     visible=old?.state===elevation&&old.visible||!!exposedBounds(area,covers)
     if(!old||old.state!==elevation||old.visible!==visible)retained.set(shadow.element,{state:elevation,visible})
    }else visible=!!exposedBounds(area,covers)
    if(!visible){set(shadow.element,'clip','');set(shadow.element,'clip-path','inset(100%)')}
    else{
     // Stable original image regions outperform partial native masks under
     // the complete workload. Completely hidden patches still skip paint.
     set(shadow.element,'clip-path','');set(shadow.element,'clip','')
    }
   }
   const radius=`${window.radius}px`
   set(window.frame,'--material-frame-radius',radius)
   if(!moving)window.record.terminalExposure=undefined
   // The exposed window may contain only sidebar/toolbar strips while the
   // terminal screen itself is completely covered. Clip that host separately
   // so xterm's native IntersectionObserver can skip invisible drawing. Its
   // parser, PTY stream, ACKs, checkpoint and dimensions are untouched.
   for(const terminal of window.terminals){
    if(!exactPaintClipping){releaseTerminal(terminal.element);continue}
    const area=terminal.rect
    let visible=exposedBounds(area,covers)
    // Keep the original full-size terminal host and native renderer. Only its
    // existing overflow viewport becomes smaller; the host retains the same
    // world position, grid, scrolling, parser, lease, journal and ACK chain.
    // Exposure only expands during a held gesture, preventing repeated native
    // layout/paint changes for already revealed rows. Structural invalidation
    // discards the old envelope before measuring current region dimensions.
    if(moving&&visible){
     const retained=window.record.terminalExposure||=new Map(),old=retained.get(terminal.element)
     const x=area.x+Math.floor(visible.x-area.x),y=area.y+Math.floor(visible.y-area.y),right=area.x+Math.ceil(visible.x+visible.width-area.x),bottom=area.y+Math.ceil(visible.y+visible.height-area.y)
     const bounds=retainExposureBounds(area,old,{x,y,width:right-x,height:bottom-y});retained.set(terminal.element,bounds)
     visible={...bounds,x:area.x+bounds.x,y:area.y+bounds.y}
    }
    const hidden=area.width>0&&area.height>0&&!visible
    if(hidden&&(clippedTerminals.has(terminal.element)||!terminal.element.style.clipPath)){
     // WebKit ignores both clip-path and legacy clip in native intersection
     // reports. Collapse only this overflow viewport while keeping the actual
     // xterm host at the region's unchanged layout dimensions. Resize remains
     // driven by the region; parsing, leases and checkpoints never pause.
     if(terminal.viewport){set(terminal.element,'width',terminal.width);set(terminal.element,'height',terminal.height);set(terminal.element,'left','0px');set(terminal.element,'top','0px');set(terminal.viewport,'left','0px');set(terminal.viewport,'top','0px');set(terminal.viewport,'width','0px');set(terminal.viewport,'height','0px')}
     else set(terminal.element,'clip-path','inset(100%)')
     terminal.element.dataset.paintOccluded='true';delete terminal.element.dataset.paintPartial;clippedTerminals.add(terminal.element)
    }
    else if(visible&&terminal.viewport&&area.width>0&&area.height>0&&(clippedTerminals.has(terminal.element)||!terminal.element.style.clipPath)){
     // Round outward and retain one CSS pixel for native fractional-edge AA.
     // Conservative extra paint is safe; never cut a potentially exposed row.
     const left=Math.max(0,Math.floor(visible.x-area.x)-1),top=Math.max(0,Math.floor(visible.y-area.y)-1)
     const right=Math.min(area.width,Math.ceil(visible.x+visible.width-area.x)+1),bottom=Math.min(area.height,Math.ceil(visible.y+visible.height-area.y)+1)
     if(left===0&&top===0&&right===area.width&&bottom===area.height){releaseTerminal(terminal.element);continue}
     set(terminal.element,'width',terminal.width);set(terminal.element,'height',terminal.height);set(terminal.element,'left',`${-left}px`);set(terminal.element,'top',`${-top}px`)
     set(terminal.viewport,'left',`${left}px`);set(terminal.viewport,'top',`${top}px`);set(terminal.viewport,'width',`${right-left}px`);set(terminal.viewport,'height',`${bottom-top}px`)
     delete terminal.element.dataset.paintOccluded;terminal.element.dataset.paintPartial='true';clippedTerminals.add(terminal.element)
    }else if(!hidden)releaseTerminal(terminal.element)
   }
   // Extend the wallpaper-only header plane below the body join. Fractional
   // device pixels otherwise leave two independently antialiased edges over
   // lower application content. An opaque overlap restores the native single
   // material's coverage without capturing or changing application paint.
   const headerSize=`${Math.max(0,window.bodyRect.y-window.rect.y+1)}px`
   set(window.frame,'--material-header-size',headerSize)
   // Split optical planes so clipping application paint also clips its cached
   // material; an unchanged full-window bitmap must not undo the savings.
   const atlasX=parseFloat(window.frame.style.getPropertyValue('--material-x')),atlasY=parseFloat(window.frame.style.getPropertyValue('--material-y'))
   const x=`${(Number.isFinite(atlasX)?atlasX:-window.rect.x)-(window.bodyRect.x-window.rect.x)}px`,y=`${(Number.isFinite(atlasY)?atlasY:-window.rect.y)-(window.bodyRect.y-window.rect.y)}px`
   set(window.body,'--material-x',x);set(window.body,'--material-y',y)
   const r=window.bodyRect
   const held=window.record.exposure
   // Once the conservative gesture envelope covers the complete body, no
   // subtraction can enlarge it. Keep native ownership and skip only this
   // redundant calculation; partial/hidden bodies still track first reveal.
   const full=moving&&held&&held.x<=0&&held.y<=0&&held.x+held.width>=r.width&&held.y+held.height>=r.height
   let visible=!exactPaintClipping||full?r:exposedBounds(r,covers)
   if(moving){
    // Keep a conservative exposure envelope for the duration of a gesture.
    // Expanding ahead never hides visible content, and avoids repainting every
    // lower application just because its clip moved by one pixel this frame.
    if(visible&&!full){
     window.record.exposure=retainExposureBounds(r,window.record.exposure,visible)
    }
    const retained=window.record.exposure
    if(retained)visible={...retained,x:r.x+retained.x,y:r.y+retained.y}
   }else window.record.exposure=undefined
   if(!visible)set(window.body,'clip-path','inset(100%)')
   else{
    const top=Math.max(0,visible.y-r.y),left=Math.max(0,visible.x-r.x),bottom=Math.max(0,r.y+r.height-visible.y-visible.height),right=Math.max(0,r.x+r.width-visible.x-visible.width)
    set(window.body,'clip-path',top+left+bottom+right<.01?'':`inset(${top}px ${right}px ${bottom}px ${left}px)`)
   }
   // These two rectangles lie wholly inside the unchanged rounded contour.
   // Exclude a pixel at the edge: antialiasing and translucent border pixels
   // never establish opaque coverage of the window underneath.
   const b=window.rect,p=Math.max(1,Math.min(window.radius,b.width/2,b.height/2))
   const coversBelow=opaque&&window.solid
   const cover=(region:PaintRect)=>covers.push(region)
   if(coversBelow&&b.width>p*2&&b.height>2)cover({x:b.x+p,y:b.y+1,width:b.width-p*2,height:b.height-2})
   if(coversBelow&&b.width>2&&b.height>p*2)cover({x:b.x+1,y:b.y+p,width:b.width-2,height:b.height-p*2})
  }
 }
 const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
 const preferences=new MutationObserver(()=>{for(const record of tracked.values())record.dirty=true;update()});preferences.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-material-cache-opaque','data-translucency','data-theme','data-palette','data-density']})
 const hostChanged=(event:Event)=>{const frame=(event.target as HTMLElement).closest<HTMLElement>('.app-window');queueMicrotask(()=>{if(disposed)return;const record=frame&&tracked.get(frame);if(record){register(record);update()}else scan()})}
 desktop.addEventListener('terminal-paint-host-changed',hostChanged)
 const scrolled=(event:Event)=>{const target=event.target;if(!(target instanceof HTMLElement))return;const frame=target.closest<HTMLElement>('.app-window'),record=frame&&tracked.get(frame);if(record?.hosts.length&&!target.closest('.terminal-container,.monaco-editor')){record.dirty=true;update()}}
 desktop.addEventListener('scroll',scrolled,true)
 const focused=(event:Event)=>{const target=event.target;if(!(target instanceof HTMLElement))return;const frame=target.closest<HTMLElement>('.app-window'),record=frame&&tracked.get(frame);if(record?.hosts.length){record.dirty=true;update()}}
 desktop.addEventListener('focusin',focused)
 window.addEventListener('resize',invalidate);scan()
 return ()=>{disposed=true;if(animation)cancelAnimationFrame(animation);children.disconnect();attributes.disconnect();preferences.disconnect();window.removeEventListener('resize',invalidate);desktop.removeEventListener('terminal-paint-host-changed',hostChanged);desktop.removeEventListener('scroll',scrolled,true);desktop.removeEventListener('focusin',focused);for(const record of tracked.values())release(record);for(const terminal of clippedTerminals)releaseTerminal(terminal);tracked.clear()}
}
