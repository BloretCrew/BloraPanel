// Optical-only topology experiment. Native corner PNGs keep their original
// texels, sampling and coordinates. Partition the nonzero alpha footprint
// into disjoint native background rectangles; never read live application
// nodes, text, glyphs, fields or surfaces into a bitmap.
export async function installShadowAlphaDiagnostic(maxLeaves:2|4=2){
 type Rect={x:number;y:number;width:number;height:number}
 type Plan={width:number;height:number;rects:Rect[];url:string}
 type Corner={element:HTMLElement;ticket:number;key:string;nodes:HTMLElement[]}
 type Frame={frame:HTMLElement;plane:HTMLElement;corners:Corner[]}
 const desktop=document.querySelector<HTMLElement>('.desktop');if(!desktop)return
 const root=document.documentElement,tracked=new Map<HTMLElement,Frame>(),plans=new Map<string,Promise<Plan>>()
 const css=document.createElement('style');css.textContent='.shadow-piece[data-native-alpha-bands="ready"]{background-image:none!important}.native-shadow-alpha-band{position:absolute;display:block;pointer-events:none;overflow:hidden}.native-shadow-alpha-band>i{position:absolute;display:block;background-repeat:no-repeat;background-size:100% 100%}'
 document.head.append(css)
 let disposed=false,paused=false,generation=0
 const clear=(corner:Corner)=>{corner.ticket++;corner.key='';delete corner.element.dataset.nativeAlphaBands;for(const node of corner.nodes)node.remove();corner.nodes=[]}
 const measure=(record:Frame,corner:Corner)=>{
  if(devicePixelRatio!==1||!record.frame.isConnected||!record.plane.dataset.shadowCache)return
  const position=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(record.frame.style.transform)
  if(!position||!position.slice(1).every(v=>Number.isInteger(Number(v))))return
  const state=record.frame.classList.contains('focused')&&record.frame.dataset.shadowLow!=='true'?'focused':'normal'
  const element=corner.element,url=/^url\("(blob:[^"]+)"\)$/.exec(element.style.getPropertyValue('--shadow-'+state))?.[1]
  const width=parseFloat(element.style.getPropertyValue('--shadow-'+state+'-width')),height=parseFloat(element.style.getPropertyValue('--shadow-'+state+'-height'))
  if(!url||element.style.getPropertyValue('--shadow-'+state+'-display')==='none'||![width,height].every(v=>Number.isInteger(v)&&v>0&&v<=1024))return
  return {url,width,height,key:JSON.stringify([url,width,height,maxLeaves])}
 }
 async function compile(value:{url:string;width:number;height:number}):Promise<Plan>{
  const {url,width,height}=value,image=new Image();image.src=url;await image.decode()
  if(image.naturalWidth!==width||image.naturalHeight!==height)throw Error('native shadow sampling is not one to one')
  const canvas=document.createElement('canvas');canvas.width=width;canvas.height=height
  const context=canvas.getContext('2d');if(!context)throw Error('native optical alpha is unavailable')
  context.drawImage(image,0,0);const pixels=context.getImageData(0,0,width,height).data
  const bounds=(area:Rect):Rect|undefined=>{
   let left=area.x+area.width,top=area.y+area.height,right=area.x,bottom=area.y
   for(let y=area.y;y<area.y+area.height;y++)for(let x=area.x;x<area.x+area.width;x++)if(pixels[(y*width+x)*4+3]){left=Math.min(left,x);top=Math.min(top,y);right=Math.max(right,x+1);bottom=Math.max(bottom,y+1)}
   return right>left&&bottom>top?{x:left,y:top,width:right-left,height:bottom-top}:undefined
  }
  const split=(area:Rect)=>{
   let best:{gain:number;parts:Rect[]}|undefined
   for(const axis of ['x','y'] as const){
    const size=axis==='x'?area.width:area.height
    const strips=Array.from({length:size},(_,i)=>bounds(axis==='x'?{x:area.x+i,y:area.y,width:1,height:area.height}:{x:area.x,y:area.y+i,width:area.width,height:1}))
    const union=(a:Rect|undefined,b:Rect|undefined):Rect|undefined=>{if(!a)return b;if(!b)return a;const x=Math.min(a.x,b.x),y=Math.min(a.y,b.y);return {x,y,width:Math.max(a.x+a.width,b.x+b.width)-x,height:Math.max(a.y+a.height,b.y+b.height)-y}}
    const prefix:(Rect|undefined)[]=[],suffix:(Rect|undefined)[]=[]
    for(let i=0;i<size;i++)prefix[i]=union(prefix[i-1],strips[i])
    for(let i=size-1;i>=0;i--)suffix[i]=union(suffix[i+1],strips[i])
    for(let cut=1;cut<size;cut++){
     const first=prefix[cut-1],second=suffix[cut]
     const parts=[first,second].filter((r):r is Rect=>!!r),gain=area.width*area.height-parts.reduce((sum,r)=>sum+r.width*r.height,0)
     if(parts.length===2&&gain>0&&(!best||gain>best.gain))best={gain,parts}
    }
   }
   return best
  }
  const initial=bounds({x:0,y:0,width,height});if(!initial)throw Error('native shadow corner is empty')
  let rects=[initial]
  while(rects.length<maxLeaves){
   let choice:{index:number;gain:number;parts:Rect[]}|undefined
   for(let index=0;index<rects.length;index++){const next=split(rects[index]!);if(next&&(!choice||next.gain>choice.gain))choice={index,...next}}
   if(!choice)break
   rects.splice(choice.index,1,...choice.parts)
  }
  // Certify every original nonzero texel appears in exactly one rectangle.
  // This does not sample or make assumptions about the live app underneath.
  for(let y=0;y<height;y++)for(let x=0;x<width;x++)if(pixels[(y*width+x)*4+3]){
   if(rects.filter(r=>x>=r.x&&x<r.x+r.width&&y>=r.y&&y<r.y+r.height).length!==1)throw Error('native shadow alpha partition is incomplete')
  }
  return {url,width,height,rects}
 }
 async function update(record:Frame,corner:Corner){
  if(disposed||paused)return
  const value=measure(record,corner);if(!value){clear(corner);return}if(corner.key===value.key)return
  clear(corner);const ticket=corner.ticket,version=generation
  if(!plans.has(value.key)&&plans.size>=64)return
  if(!plans.has(value.key))plans.set(value.key,compile(value))
  try{
   const plan=await plans.get(value.key)!
   if(disposed||paused||version!==generation||corner.ticket!==ticket||measure(record,corner)?.key!==value.key)return
   const nodes=plan.rects.map(r=>{
    const node=document.createElement('i');node.className='native-shadow-alpha-band'
    node.style.cssText=`left:${r.x}px;top:${r.y}px;width:${r.width}px;height:${r.height}px`
    // Retain the original full native image quad and UV mapping inside an
    // overflow viewport. WebKit's independently resized background rectangles
    // altered alpha rounding by one; that failed direct-band control stays in
    // its separate report. Nothing is reencoded and no threshold is changed.
    const image=document.createElement('i');image.style.cssText=`left:${-r.x}px;top:${-r.y}px;width:${plan.width}px;height:${plan.height}px;background-image:url("${plan.url}")`;node.append(image)
    return node
   })
   corner.nodes=nodes;corner.element.append(...nodes);corner.key=value.key;corner.element.dataset.nativeAlphaBands='ready'
  }catch{if(!disposed&&corner.ticket===ticket)clear(corner)}
 }
 const updateFrame=(record:Frame)=>Promise.all(record.corners.map(corner=>update(record,corner)))
 const mutations=new MutationObserver(records=>{for(const frame of new Set(records.map(r=>(r.target as HTMLElement).closest<HTMLElement>('.app-window')).filter((f):f is HTMLElement=>!!f))){const record=tracked.get(frame);if(record)void updateFrame(record)}})
 function scan(){
  mutations.disconnect()
  for(const [frame,record] of tracked)if(!frame.isConnected){for(const corner of record.corners)clear(corner);tracked.delete(frame)}
  for(const frame of desktop!.querySelectorAll<HTMLElement>(':scope>.app-window')){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!plane)continue
   if(!tracked.has(frame))tracked.set(frame,{frame,plane,corners:[...plane.querySelectorAll<HTMLElement>(':scope>.shadow-piece:is(.nw,.ne,.sw,.se)')].map(element=>({element,ticket:0,key:'',nodes:[]}))})
   mutations.observe(frame,{attributes:true,attributeFilter:['style','class','data-shadow-low']});mutations.observe(plane,{attributes:true,attributeFilter:['data-shadow-cache']})
  }
  for(const record of tracked.values())void updateFrame(record)
 }
 const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
 const invalidate=()=>{generation++;for(const record of tracked.values())for(const corner of record.corners)clear(corner);plans.clear();for(const record of tracked.values())void updateFrame(record)}
 const appearance=new MutationObserver(invalidate);appearance.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-translucency']})
 const controller={pause(){paused=true;for(const record of tracked.values())for(const corner of record.corners)clear(corner)},async resume(){paused=false;await Promise.all([...tracked.values()].map(updateFrame))},dispose(){disposed=true;generation++;mutations.disconnect();children.disconnect();appearance.disconnect();for(const record of tracked.values())for(const corner of record.corners)clear(corner);tracked.clear();plans.clear();css.remove();window.removeEventListener('resize',invalidate)}}
 ;(window as unknown as {__shadowAlphaDiagnostic:typeof controller}).__shadowAlphaDiagnostic=controller
 window.addEventListener('resize',invalidate);window.addEventListener('pagehide',controller.dispose,{once:true});scan()
 await Promise.all([...tracked.values()].map(updateFrame))
}
