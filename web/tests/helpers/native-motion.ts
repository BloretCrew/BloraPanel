import type {Page} from '@playwright/test'

// Native animation transport experiment only. Every application stays in its
// original live DOM and browser paint ownership; no app image is generated.
export async function installNativeMotion(page:Page){
 await page.evaluate(()=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!,sample=desktop.querySelector<HTMLElement>('.app-window')!
  let owner=Object.getPrototypeOf(sample.style),original:PropertyDescriptor|undefined
  while(owner&&!(original=Object.getOwnPropertyDescriptor(owner,'transform')))owner=Object.getPrototypeOf(owner)
  if(!original?.get||!original.set)throw Error('Native transform accessor unavailable')
  const get=original.get,set=original.set
  type Record={frame:HTMLElement;style:CSSStyleDeclaration;origin?:{x:number;y:number};point?:{x:number;y:number};alias?:string;animations?:[Animation,Animation];resizing:boolean}
  const records=new Map<HTMLElement,Record>(),styles=new WeakMap<CSSStyleDeclaration,Record>()
  const parse=(value:string)=>{const m=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(value);return m?{x:Number(m[1]),y:Number(m[2])}:undefined}
  let enabled=true,disposed=false,updates=0,serial=0
  const property=document.createElement('style');property.textContent='@property --material-motion{syntax:"<integer>";inherits:false;initial-value:0}';document.head.append(property)
  function release(record:Record){
   if(!record.animations)return
   const value=record.alias!;record.alias=undefined;record.origin=undefined;record.point=undefined
   set.call(record.style,value);for(const animation of record.animations)animation.cancel();record.animations=undefined;record.style.removeProperty('--material-motion')
  }
  function arm(record:Record){
   if(record.animations)return true
   const point=parse(get.call(record.style));if(!point)return false
   const matrix=new DOMMatrixReadOnly(getComputedStyle(record.frame).transform)
   if(!matrix.is2D||matrix.a!==1||matrix.b!==0||matrix.c!==0||matrix.d!==1||Math.abs(matrix.e-point.x)>.01||Math.abs(matrix.f-point.y)>.01)return false
   const x=record.frame.animate([{transform:'translateX(-16384px)'},{transform:'translateX(16384px)'}],{duration:32768,fill:'both',composite:'add'}),y=record.frame.animate([{transform:'translateY(-16384px)'},{transform:'translateY(16384px)'}],{duration:32768,fill:'both',composite:'add'})
   for(const animation of [x,y]){animation.pause();animation.currentTime=16384}
   record.animations=[x,y];record.origin=point;record.point=point;record.alias=get.call(record.style)
   return true
  }
  const getter=function(this:CSSStyleDeclaration){return styles.get(this)?.alias??get.call(this)}
  const setter=function(this:CSSStyleDeclaration,value:string){
   const record=styles.get(this),point=parse(value)
   if(record&&enabled&&Number.isInteger(devicePixelRatio||1)&&record.frame.classList.contains('moving')&&!record.resizing&&point&&arm(record)){
    const dx=point.x-record.origin!.x,dy=point.y-record.origin!.y
    if(Math.abs(dx)<=16384&&Math.abs(dy)<=16384){
     record.alias=value;record.point=point;record.animations![0].currentTime=16384+dx;record.animations![1].currentTime=16384+dy
     // The original ownership/occlusion observers receive a signed geometry
     // publication. This unused non-inherited variable carries no app style.
     this.setProperty('--material-motion',String(++serial));updates++;return
    }
   }
   if(record)release(record);set.call(this,value)
  }
  Object.defineProperty(owner,'transform',{...original,get:getter,set:setter})
  function scan(){
   for(const [frame,record] of records)if(!frame.isConnected){release(record);records.delete(frame);styles.delete(record.style)}
   for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))if(!records.has(frame)){const record={frame,style:frame.style,resizing:false};records.set(frame,record);styles.set(frame.style,record)}
  }
  // Filter native frame/resize ownership only; never observe terminal rows or
  // application text mutations. Registration is limited to desktop children.
  const frames=new MutationObserver(entries=>{for(const entry of entries){const record=records.get(entry.target as HTMLElement);if(record&&!record.frame.classList.contains('moving'))release(record)}})
  const children=new MutationObserver(()=>{scan();frames.disconnect();for(const frame of records.keys())frames.observe(frame,{attributes:true,attributeFilter:['class']})})
  const pointer=(event:PointerEvent)=>{const frame=(event.target as Element)?.closest<HTMLElement>('.app-window'),record=frame&&records.get(frame);if(record)record.resizing=!!(event.target as Element).closest('.resize-handle')}
  desktop.addEventListener('pointerdown',pointer,true);scan();for(const frame of records.keys())frames.observe(frame,{attributes:true,attributeFilter:['class']});children.observe(desktop,{childList:true})
  const dispose=()=>{if(disposed)return;disposed=true;frames.disconnect();children.disconnect();desktop.removeEventListener('pointerdown',pointer,true);for(const record of records.values())release(record);records.clear();if(Object.getOwnPropertyDescriptor(owner,'transform')?.get===getter)Object.defineProperty(owner,'transform',original!);property.remove()}
  const snapshot=()=>{let maxGeometryError=0;for(const record of records.values())if(record.animations){const matrix=new DOMMatrixReadOnly(getComputedStyle(record.frame).transform);maxGeometryError=Math.max(maxGeometryError,Math.abs(matrix.e-record.point!.x),Math.abs(matrix.f-record.point!.y))}return {tracked:records.size,eligible:Number.isInteger(devicePixelRatio||1)?records.size:0,updates,active:[...records.values()].filter(r=>!!r.animations).length,maxGeometryError}}
  ;(window as any).__nativeMotion={dispose,snapshot,setVisible(value:boolean){enabled=value;if(!value)for(const record of records.values())release(record)}}
  window.addEventListener('pagehide',dispose,{once:true})
 })
}
