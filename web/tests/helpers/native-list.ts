import {exposedBounds,type PaintRect} from '../../src/desktop/occlusion'

// Rejected bounded native-list clipping, preserved only as an explicit pixel
// control. Product source neither imports this helper nor observes list rows.
export function installNativeListDiagnostic(){
 const desktop=document.querySelector<HTMLElement>('.desktop')!,root=document.documentElement
 type Record={body:HTMLElement;elements:HTMLElement[];revealed:Set<HTMLElement>;structure:MutationObserver;sizes:ResizeObserver}
 const tracked=new Map<HTMLElement,Record>(),clipped=new Set<HTMLElement>()
 let pending=0,disposed=false
 const release=(element:HTMLElement)=>{if(clipped.delete(element)){element.style.removeProperty('clip-path');delete element.dataset.appPaintOccluded}}
 const queue=()=>{if(!disposed&&!pending)pending=requestAnimationFrame(apply)}
 function register(frame:HTMLElement,record:Record){
  record.structure.disconnect();record.sizes.disconnect()
  const elements:HTMLElement[]=[],structural=new Set<HTMLElement>([record.body]),layout=new Set<HTMLElement>([frame,record.body])
  for(const element of record.body.querySelectorAll<HTMLElement>('.instance-identity,.file-row>span,.resource-table td,.sidebar-item span,.file-tree-label span,.audit-row span,.audit-row strong')){
   if(elements.length>=64)break
   if(element.closest('.terminal-container,.monaco-editor,iframe')||elements.some(parent=>parent.contains(element)))continue
   elements.push(element);layout.add(element)
   for(let parent=element.parentElement;parent&&parent!==frame;parent=parent.parentElement){
    structural.add(parent);layout.add(parent)
    for(const child of parent.children)if(child instanceof HTMLElement)layout.add(child)
   }
  }
  for(const old of record.elements)if(!elements.includes(old)){release(old);record.revealed.delete(old)}
  record.elements=elements
  for(const element of structural)record.structure.observe(element,{childList:true})
  for(const element of layout)record.sizes.observe(element)
 }
 function apply(){
  pending=0;if(disposed)return
  for(const [frame,record] of tracked)if(!frame.isConnected){record.structure.disconnect();record.sizes.disconnect();for(const element of record.elements)release(element);tracked.delete(frame)}
  for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))if(!tracked.has(frame)){
   const body=frame.querySelector<HTMLElement>(':scope>.window-body');if(!body)continue
   const record={} as Record
   Object.assign(record,{body,elements:[],revealed:new Set<HTMLElement>(),structure:new MutationObserver(()=>{register(frame,record);queue()}),sizes:new ResizeObserver(queue)})
   tracked.set(frame,record);register(frame,record)
  }
  const opaque=root.dataset.translucency==='off'||(root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true')
  const frames=[...tracked].map(([frame,record])=>({frame,record,style:getComputedStyle(frame),box:frame.getBoundingClientRect(),islands:record.elements.map(element=>{
   const style=getComputedStyle(element),box=element.getBoundingClientRect()
   const bounded=['hidden','clip','auto','scroll'].includes(style.overflowX)&&['hidden','clip','auto','scroll'].includes(style.overflowY)&&style.boxShadow==='none'&&style.filter==='none'&&style.display!=='inline'&&(style.clipPath==='none'||clipped.has(element))
   return {element,bounded,rect:{x:box.x-2,y:box.y-2,width:box.width+4,height:box.height+4}}
  })})).filter(value=>value.style.display!=='none').sort((a,b)=>Number(b.style.zIndex)-Number(a.style.zIndex))
  const moving=frames.some(value=>value.frame.classList.contains('moving')),covers:PaintRect[]=[]
  for(const {record,style,box,islands} of frames){
   if(!moving)record.revealed.clear()
   for(const island of islands){
    const hidden=opaque&&island.bounded&&!island.element.contains(document.activeElement)&&!exposedBounds(island.rect,covers)
    if(moving&&!hidden)record.revealed.add(island.element)
    if(hidden&&!record.revealed.has(island.element)){if(island.element.style.clipPath!=='inset(100%)')island.element.style.clipPath='inset(100%)';island.element.dataset.appPaintOccluded='true';clipped.add(island.element)}else release(island.element)
   }
   const p=Math.max(1,Math.min(parseFloat(style.borderTopLeftRadius)||0,box.width/2,box.height/2))
   if(opaque&&style.visibility==='visible'&&Number(style.opacity)===1){
    if(box.width>p*2&&box.height>2)covers.push({x:box.x+p,y:box.y+1,width:box.width-p*2,height:box.height-2})
    if(box.width>2&&box.height>p*2)covers.push({x:box.x+1,y:box.y+p,width:box.width-2,height:box.height-p*2})
   }
  }
 }
 const mutations=new MutationObserver(records=>{if(records.some(record=>record.target instanceof Element&&(record.target.matches('.app-window,.app-view')||(record.type==='childList'&&record.target===desktop))))queue()})
 mutations.observe(desktop,{attributes:true,attributeFilter:['style','class','data-active-view'],subtree:true,childList:true})
 const preferences=new MutationObserver(queue);preferences.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-material-cache-opaque','data-translucency','data-theme','data-palette','data-density']})
 const scrolled=(event:Event)=>{if(event.target instanceof Element&&!event.target.closest('.terminal-container,.monaco-editor'))queue()}
 desktop.addEventListener('scroll',scrolled,true);desktop.addEventListener('focusin',queue);window.addEventListener('resize',queue);queue()
 const dispose=()=>{disposed=true;if(pending)cancelAnimationFrame(pending);mutations.disconnect();preferences.disconnect();desktop.removeEventListener('scroll',scrolled,true);desktop.removeEventListener('focusin',queue);window.removeEventListener('resize',queue);for(const record of tracked.values()){record.structure.disconnect();record.sizes.disconnect()}for(const element of clipped)release(element);tracked.clear()}
 window.addEventListener('pagehide',dispose,{once:true});return dispose
}
