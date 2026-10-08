import type {Page} from '@playwright/test'

// Test-only native background topology. One original native pseudo-element
// paints the same shared atlas behind the unchanged frame contents. The older
// direct-background control remains executable after failing native edge AA.
// No app/image/text capture; no application nodes or geometry change.
export async function installNativeFrameMaterial(page:Page,strategy:'background'|'single-plane'='single-plane'){
 await page.evaluate(strategy=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!,root=document.documentElement,frames=new Set<HTMLElement>(),style=document.createElement('style')
  const selector=':root[data-translucency="on"][data-material-cache="ready"][data-material-cache-opaque="true"] .app-window[data-native-frame-material="true"]'
  style.textContent=strategy==='background'
   ?`${selector}{background-image:var(--material-window)!important;background-size:var(--material-size);background-position:var(--material-x) var(--material-y);background-repeat:no-repeat;background-origin:border-box;background-clip:padding-box;background-color:var(--surface-container)!important}${selector}::before,${selector}>.window-body::before{display:none!important}`
   :`${selector}::before{inset:0!important;height:auto!important;border-radius:calc(var(--material-frame-radius,var(--shape-xl)) - 1px)!important}${selector}>.window-body::before{display:none!important}`
  document.head.append(style)
  let enabled=true,disposed=false
  const scan=()=>{
   if(disposed)return
   for(const frame of frames)if(!frame.isConnected){delete frame.dataset.nativeFrameMaterial;frames.delete(frame)}
   for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))frames.add(frame)
   // Chromium changes native pixels when the independent body plane is
   // removed (the strict full-screen comparison fails). Keep its original
   // topology; it already passes the original mixed-load latency target.
   const nativeOwnership=!/(?:Chrome|Chromium|Edg)\//.test(navigator.userAgent)
   const active=enabled&&nativeOwnership&&Number.isInteger(devicePixelRatio||1)&&root.dataset.translucency==='on'&&root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true'
   for(const frame of frames)if(active)frame.dataset.nativeFrameMaterial='true';else delete frame.dataset.nativeFrameMaterial
  }
  const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
  const appearance=new MutationObserver(scan);appearance.observe(root,{attributes:true,attributeFilter:['data-translucency','data-material-cache','data-material-cache-opaque','data-theme','data-palette']})
  window.addEventListener('resize',scan)
  const controller={snapshot:()=>({frames:frames.size,enabled:[...frames].filter(frame=>frame.dataset.nativeFrameMaterial==='true').length}),setVisible(value:boolean){enabled=value;scan()},dispose(){if(disposed)return;disposed=true;children.disconnect();appearance.disconnect();window.removeEventListener('resize',scan);for(const frame of frames)delete frame.dataset.nativeFrameMaterial;frames.clear();style.remove()}}
  ;(window as any).__nativeFrameMaterial=controller
  window.addEventListener('pagehide',controller.dispose,{once:true});scan()
 },strategy)
}
