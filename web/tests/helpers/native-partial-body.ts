import type {Page} from '@playwright/test'

// Attribution candidate: let the browser paint original native body contents
// through its normal rounded overflow contour. Only a partially hidden body
// loses its extra inset mask, and only behind proven opaque glass. Completely
// covered bodies, fractional DPR and all nonopaque/solid paths remain native.
// No application nodes, text, buffers, pixels or bitmap caches are created.
export async function installNativePartialBody(page:Page){
 await page.evaluate(()=>{
  const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!,bodies=new Set<HTMLElement>(),style=document.createElement('style')
  style.textContent='.window-body[data-native-partial-body="true"]{clip-path:none!important}'
  document.head.append(style)
  let enabled=true,disposed=false
  const active=()=>enabled&&devicePixelRatio===1&&root.dataset.translucency==='on'&&root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true'
  function apply(body:HTMLElement){
   const clip=body.style.clipPath
   if(active()&&clip&&clip!=='inset(100%)')body.dataset.nativePartialBody='true';else delete body.dataset.nativePartialBody
  }
  const attributes=new MutationObserver(entries=>{if(!disposed)for(const body of new Set(entries.map(entry=>entry.target as HTMLElement)))apply(body)})
  function scan(){
   if(disposed)return
   for(const body of bodies)if(!body.isConnected){delete body.dataset.nativePartialBody;bodies.delete(body)}
   for(const body of desktop.querySelectorAll<HTMLElement>(':scope>.app-window>.window-body'))if(!bodies.has(body)){bodies.add(body);attributes.observe(body,{attributes:true,attributeFilter:['style']});apply(body)}
  }
  const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
  const appearance=new MutationObserver(()=>bodies.forEach(apply));appearance.observe(root,{attributes:true,attributeFilter:['data-theme','data-palette','data-material-cache','data-material-cache-opaque','data-translucency']})
  const viewport=()=>bodies.forEach(apply);window.addEventListener('resize',viewport);scan()
  const dispose=()=>{disposed=true;attributes.disconnect();children.disconnect();appearance.disconnect();window.removeEventListener('resize',viewport);for(const body of bodies)delete body.dataset.nativePartialBody;bodies.clear();style.remove()}
  ;(window as any).__nativePartialBody={setVisible(value:boolean){enabled=value;bodies.forEach(apply)},snapshot(){return {tracked:bodies.size,unmasked:[...bodies].filter(body=>body.dataset.nativePartialBody==='true').length,covered:[...bodies].filter(body=>body.style.clipPath==='inset(100%)').length}},dispose}
  window.addEventListener('pagehide',dispose,{once:true})
 })
}
