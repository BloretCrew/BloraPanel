// Rejected performance candidate retained only for explicit native diagnostics.
import type {PaintRect} from '../../src/desktop/occlusion'
export type ShadowSegment={element:HTMLElement;normal:PaintRect;focused:PaintRect}
type Entry={key:string;elements:HTMLElement[]}
const retained=new WeakMap<HTMLElement,Entry>()
export function clearShadowSegments(piece:HTMLElement){
 const entry=retained.get(piece);if(entry)for(const element of entry.elements)element.remove()
 retained.delete(piece);delete piece.dataset.shadowSegmented
}
// Only content-free optical images are subdivided. All children sample the
// same original image coordinates. Internal edges land on physical pixels,
// avoiding a second antialias/blend contribution at a fractional seam.
export function segmentShadowPatch(piece:HTMLElement,normal:PaintRect|undefined,focused:PaintRect|undefined):ShadowSegment[]{
 const horizontal=piece.classList.contains('n')||piece.classList.contains('s'),vertical=piece.classList.contains('w')||piece.classList.contains('e')
 if(!normal||!focused||(!horizontal&&!vertical)){clearShadowSegments(piece);return []}
 // Other display scales retain the original decoded native patch. Native
 // fractional-DPR interpolation under independent clips is not bit-identical
 // to the original image quad; never trade visible pixels for this fast path.
 if(devicePixelRatio!==1||[normal,focused].some(area=>[area.x,area.y,area.width,area.height].some(value=>Math.abs(value-Math.round(value))>.001))){clearShadowSegments(piece);return []}
 const axis=horizontal?'x':'y',extent=horizontal?'width':'height',native=piece.getBoundingClientRect(),dpr=devicePixelRatio||1
 // The native layout engine quantizes CSS side offsets independently. Use
 // its actual stretch-axis origin/length rather than recomputing that width
 // from unquantized custom-property floats.
 normal={...normal,[axis]:native[axis],[extent]:native[extent]};focused={...focused,[axis]:native[axis],[extent]:native[extent]}
 const start=normal[axis],length=normal[extent]
 if(length<128||Math.abs(start-focused[axis])>.001||Math.abs(length-focused[extent])>.001){clearShadowSegments(piece);return []}
 const stride=Math.ceil(Math.max(64,length/32)*dpr),cuts=[0]
 for(let edge=(Math.floor(start*dpr)+stride)/dpr;edge<start+length;edge+=stride/dpr)if(edge>start)cuts.push(edge-start)
 cuts.push(length)
 const key=JSON.stringify([horizontal,normal.width,normal.height,focused.width,focused.height,cuts.map(value=>Number(value.toFixed(5)))])
 let entry=retained.get(piece)
 if(entry?.key!==key){
  clearShadowSegments(piece)
  const elements=cuts.slice(0,-1).map((offset,index)=>{
   const element=document.createElement('i');element.className='shadow-segment'
   const size=cuts[index+1]!-offset
   element.style.cssText=horizontal?`left:${offset}px;top:0;width:${size}px;height:100%`:`left:0;top:${offset}px;width:100%;height:${size}px`
   // Clip only internal subdivision boundaries. Original optical image
   // edges have native coverage beyond their fractional box; retain it.
   const first=index===0?-2:0,last=index===cuts.length-2?-2:0
   element.style.clipPath=horizontal?`inset(-2px ${last}px -2px ${first}px)`:`inset(${first}px -2px ${last}px -2px)`
   // Preserve the original full image quad and UV mapping under a small
   // rectangular overflow clip. Repositioning/resizing the image itself can
   // change native bilinear rounding at fractional DPR.
   const image=document.createElement('i');image.style.cssText=horizontal?`left:${-offset}px;top:0;width:${length}px;height:100%`:`left:0;top:${-offset}px;width:100%;height:${length}px`;element.append(image)
   return element
  })
  piece.append(...elements);piece.dataset.shadowSegmented='true';entry={key,elements};retained.set(piece,entry)
 }
 return entry.elements.map((element,index)=>{
  const offset=cuts[index]!,size=cuts[index+1]!-offset
  const region=(area:PaintRect)=>horizontal?{...area,x:area.x+offset,width:size}:{...area,y:area.y+offset,height:size}
  return {element,normal:region(normal),focused:region(focused)}
 })
}

export function installShadowSegmentsDiagnostic(){
 const style=document.createElement('style');style.textContent='.shadow-piece[data-shadow-segmented="true"]{background-image:none!important}.shadow-segment{position:absolute;display:block}.shadow-segment>i{position:absolute;display:block;background-image:var(--shadow-normal);background-size:100% 100%;background-repeat:no-repeat}.app-window.focused>.window-shadow-plane>.shadow-piece>.shadow-segment>i{background-image:var(--shadow-focused)}.app-window[data-shadow-low="true"]>.window-shadow-plane>.shadow-piece>.shadow-segment>i{background-image:var(--shadow-normal)}';document.head.append(style)
 for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane[data-shadow-cache]')){
  const area=plane.getBoundingClientRect(),pad=parseFloat(plane.style.getPropertyValue('--shadow-outset')),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner'))
  for(const name of ['n','s','w','e']){
   const piece=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!,states:Record<string,PaintRect>={}
   for(const state of ['normal','focused']){
    const value=(key:string)=>parseFloat(piece.style.getPropertyValue(`--shadow-${state}-${key}`)),horizontal=name==='n'||name==='s',width=horizontal?area.width-2*corner:value('width'),height=horizontal?value('height'):area.height-2*corner
    const x=horizontal?corner:name==='w'?-pad+value('left'):area.width+pad-value('right')-width,y=!horizontal?corner:name==='n'?-pad+value('top'):area.height+pad-value('bottom')-height
    states[state]={x:area.x+x,y:area.y+y,width,height}
   }
   segmentShadowPatch(piece,states.normal,states.focused)
  }
 }
}
