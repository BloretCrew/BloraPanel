import {exposedBounds,type PaintRect} from '../../src/desktop/occlusion'

// Rejected optical-region prototype: explicitly installed only by native
// controls. Production never observes or executes this diagnostic.
export function installPartialShadowDiagnostic(){
 const desktop=document.querySelector<HTMLElement>('.desktop')!,root=document.documentElement
 let pending=0,disposed=false
 const set=(element:HTMLElement,name:string,value:string)=>{if(element.style.getPropertyValue(name)!==value){if(value)element.style.setProperty(name,value);else element.style.removeProperty(name)}}
 function apply(){
  pending=0;if(disposed)return
  const opaque=root.dataset.translucency==='off'||(root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true')
  const frames=[...desktop.querySelectorAll<HTMLElement>(':scope>.app-window')].map(frame=>({frame,style:getComputedStyle(frame),box:frame.getBoundingClientRect()})).filter(value=>value.style.display!=='none').sort((a,b)=>Number(b.style.zIndex)-Number(a.style.zIndex))
  const moving=frames.some(value=>value.frame.classList.contains('moving')),covers:PaintRect[]=[]
  for(const {frame,style,box} of frames){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane')
   for(const piece of plane?.querySelectorAll<HTMLElement>('.shadow-piece')||[]){
    set(piece,'clip','');set(piece,'clip-path','')
    if(!opaque||!plane?.dataset.shadowCache)continue
    const area=piece.getBoundingClientRect();if(!area.width||!area.height)continue
    const visible=exposedBounds(area,covers)
    if(!visible){set(piece,'clip-path','inset(100%)');continue}
    if(moving)continue
    const top=Math.max(0,visible.y-area.y),left=Math.max(0,visible.x-area.x),bottom=Math.max(0,area.y+area.height-visible.y-visible.height),right=Math.max(0,area.x+area.width-visible.x-visible.width)
    if(top+left+bottom+right<.01)continue
    if(devicePixelRatio===1)set(piece,'clip',`rect(${top-1}px ${area.width-right+1}px ${area.height-bottom+1}px ${left-1}px)`)
    else set(piece,'clip-path',`inset(${top-1}px ${right-1}px ${bottom-1}px ${left-1}px)`)
   }
   // Conservative opaque interior: never cover rounded AA edges or sample
   // application text/canvas/data.
   const inset=Math.max(1,parseFloat(style.borderTopLeftRadius)||0)+2
   if(opaque&&style.visibility==='visible'&&Number(style.opacity)===1&&box.width>2*inset&&box.height>2*inset)covers.push({x:box.x+inset,y:box.y+inset,width:box.width-2*inset,height:box.height-2*inset})
  }
 }
 const queue=()=>{if(!disposed&&!pending)pending=requestAnimationFrame(apply)}
 const mutations=new MutationObserver(records=>{if(records.some(record=>record.target instanceof Element&&(record.target.matches('.app-window,.window-shadow-plane')||(record.type==='childList'&&record.target===desktop))))queue()})
 mutations.observe(desktop,{attributes:true,attributeFilter:['style','class','data-shadow-cache'],subtree:true,childList:true})
 const preferences=new MutationObserver(queue);preferences.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-material-cache-opaque','data-translucency']})
 window.addEventListener('resize',queue);queue()
 const dispose=()=>{disposed=true;if(pending)cancelAnimationFrame(pending);mutations.disconnect();preferences.disconnect();window.removeEventListener('resize',queue);for(const piece of desktop.querySelectorAll<HTMLElement>('.shadow-piece')){set(piece,'clip','');set(piece,'clip-path','')}}
 window.addEventListener('pagehide',dispose,{once:true});return dispose
}
