import type {Page} from '@playwright/test'

// Test-only native positioning comparison. Keep the original app/terminal
// nodes, authored recovery transform and geometry; no captured app surface.
export async function installNativeWindowPosition(page:Page,retainLayer=true,identity=false){
 await page.evaluate(({retainLayer,identity})=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!,style=document.createElement('style')
  style.textContent=`@media (resolution:1dppx){.app-window[data-native-window-position="true"]{transform:${identity?'translate(0px,0px)':'none'}!important;${retainLayer?'':'will-change:auto!important;'}left:var(--native-window-x);top:var(--native-window-y)}}@property --native-window-x{syntax:"<length>";inherits:false;initial-value:0px}@property --native-window-y{syntax:"<length>";inherits:false;initial-value:0px}`
  document.head.append(style)
  const tracked=new Set<HTMLElement>();let enabled=true,disposed=false,updates=0
  function apply(frame:HTMLElement){
   const match=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(frame.style.transform)
   if(enabled&&devicePixelRatio===1&&match){
    const x=match[1]+'px',y=match[2]+'px'
    if(frame.style.getPropertyValue('--native-window-x')!==x){frame.style.setProperty('--native-window-x',x);updates++}
    if(frame.style.getPropertyValue('--native-window-y')!==y){frame.style.setProperty('--native-window-y',y);updates++}
    frame.dataset.nativeWindowPosition='true'
   }else delete frame.dataset.nativeWindowPosition
  }
  const mutations=new MutationObserver(entries=>{if(!disposed)for(const frame of new Set(entries.map(entry=>entry.target as HTMLElement)))apply(frame)})
  function scan(){if(disposed)return;for(const frame of tracked)if(!frame.isConnected)tracked.delete(frame);for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))if(!tracked.has(frame)){tracked.add(frame);mutations.observe(frame,{attributes:true,attributeFilter:['style']});apply(frame)}}
  const structure=new MutationObserver(scan);structure.observe(desktop,{childList:true});scan()
  const dispose=()=>{disposed=true;mutations.disconnect();structure.disconnect();for(const frame of tracked){delete frame.dataset.nativeWindowPosition;frame.style.removeProperty('--native-window-x');frame.style.removeProperty('--native-window-y')}tracked.clear();style.remove()}
  ;(window as any).__nativeWindowPosition={snapshot:()=>({frames:tracked.size,enabled:[...tracked].filter(frame=>frame.dataset.nativeWindowPosition==='true').length,updates}),setVisible(value:boolean){enabled=value;tracked.forEach(apply)},dispose}
  window.addEventListener('pagehide',dispose,{once:true})
 },{retainLayer,identity})
}
