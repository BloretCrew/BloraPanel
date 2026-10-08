import type {Page} from '@playwright/test'

// Optical-only diagnostic: existing native PNG alpha, resolved original
// shadow colour, unchanged geometry/UV/occlusion. Never reads app contents.
export async function installShadowMasks(page:Page){
 await page.evaluate(()=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!
  const records=new Map<HTMLElement,{normal:string;focused:string}>()
  const colour=(shadow:string)=>{
   const values=[...shadow.matchAll(/rgba?\(([^)]+)\)/g)].map(match=>match[1]!.trim().split(/[,\s/]+/).slice(0,3).map(Number))
   if(!values.length||values.some(value=>value.length!==3||value.some(channel=>!Number.isFinite(channel)||channel<0||channel>255))||values.some(value=>value.some((channel,index)=>channel!==values[0]![index])))return undefined
   return `rgb(${values[0]!.join(',')})`
  }
  const observed=new WeakSet<HTMLElement>()
  const register=()=>{
   for(const plane of records.keys())if(!plane.isConnected||!plane.dataset.shadowCache){delete plane.dataset.nativeShadowMask;records.delete(plane)}
   for(const plane of desktop.querySelectorAll<HTMLElement>('.window-shadow-plane')){
    if(!observed.has(plane)){observer.observe(plane,{attributes:true,attributeFilter:['data-shadow-cache']});observed.add(plane)}
    if(!plane.dataset.shadowCache)continue
    const [normal,focused]=JSON.parse(plane.dataset.shadowCache!) as [string,string]
    const a=colour(normal),b=colour(focused)
    if(!a||!b){delete plane.dataset.nativeShadowMask;continue}
    plane.style.setProperty('--native-shadow-normal-color',a);plane.style.setProperty('--native-shadow-focused-color',b)
    plane.dataset.nativeShadowMask='ready';records.set(plane,{normal,focused})
   }
  }
  const observer=new MutationObserver(records=>{if(records.some(record=>record.type==='childList'||record.attributeName==='data-shadow-cache'))register()})
  // Windows are direct desktop children. Never rescan on every native xterm
  // row mutation; those nodes contain application data, not optical material.
  observer.observe(desktop,{childList:true});register()
  ;(window as any).__nativeShadowMaskDiagnostic={count:()=>records.size,setVisible:(enabled:boolean)=>{for(const plane of records.keys())plane.dataset.nativeShadowMask=enabled?'ready':'reference'},dispose:()=>{observer.disconnect();for(const plane of records.keys()){delete plane.dataset.nativeShadowMask;plane.style.removeProperty('--native-shadow-normal-color');plane.style.removeProperty('--native-shadow-focused-color')}records.clear()}}
 })
 await page.addStyleTag({content:`
 .window-shadow-plane[data-native-shadow-mask="ready"]>.shadow-piece{background-image:none!important;background-color:var(--native-shadow-normal-color)!important;mask-image:var(--shadow-normal);mask-mode:alpha;mask-repeat:no-repeat;mask-size:100% 100%;mask-origin:border-box;mask-clip:border-box}
 .app-window.focused>.window-shadow-plane[data-native-shadow-mask="ready"]>.shadow-piece{background-color:var(--native-shadow-focused-color)!important;mask-image:var(--shadow-focused)}
 .app-window[data-shadow-low="true"]>.window-shadow-plane[data-native-shadow-mask="ready"]>.shadow-piece{background-color:var(--native-shadow-normal-color)!important;mask-image:var(--shadow-normal)}
 `})
}
