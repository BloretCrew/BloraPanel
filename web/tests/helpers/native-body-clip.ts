import type {Page} from '@playwright/test'

// Explicit native DOM diagnostic. No application nodes are copied, moved,
// captured or rasterized. The browser alone draws each unchanged Vue view.
export async function installNativeBodyClip(page:Page,flat=false){
 await page.evaluate(flat=>{
  const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!
  const style=document.createElement('style')
  style.textContent='@property --native-body-clip{syntax:"*";inherits:false;initial-value:auto}.window-body[data-native-body-clip="true"]{position:relative;clip-path:none!important}.window-body[data-native-body-clip="true"]>.app-view{position:absolute;inset:0;--native-body-clip:inherit;clip:var(--native-body-clip)}.window-body[data-native-body-clip="true"]::before{--native-body-clip:inherit;clip:var(--native-body-clip)}'
  if(flat)style.textContent+='.window-body[data-native-body-clip="true"]::before{will-change:auto!important}'
  document.head.append(style)
  type Record={body:HTMLElement;width:number;height:number}
  const records=new Map<HTMLElement,Record>();let enabled=true,disposed=false
  function apply(record:Record){
   const {body,width,height}=record,clip=body.style.clipPath
   const match=/^inset\(([^)]+)\)$/.exec(clip),tokens=match?.[1]!.trim().split(/\s+/)
   const lengths=tokens&&tokens.length<=4&&tokens.every(value=>/^-?[\d.]+px$/.test(value))?tokens.map(parseFloat):undefined
   const sides=lengths?[lengths[0]!,lengths[1]??lengths[0]!,lengths[2]??lengths[0]!,lengths[3]??lengths[1]??lengths[0]!]:undefined
   const exact=Number.isInteger(devicePixelRatio||1)&&(!clip||clip==='inset(100%)'||!!sides)
   const value=clip==='inset(100%)'?'rect(0px, 0px, 0px, 0px)':sides?`rect(${sides[0]}px, ${width-sides[1]!}px, ${height-sides[2]!}px, ${sides[3]}px)`:'auto'
   if(body.style.getPropertyValue('--native-body-clip')!==value)body.style.setProperty('--native-body-clip',value)
   const opaqueGlass=root.dataset.translucency==='on'&&root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true'
   if(enabled&&exact&&opaqueGlass)body.dataset.nativeBodyClip='true';else delete body.dataset.nativeBodyClip
  }
  const mutations=new MutationObserver(entries=>{for(const body of new Set(entries.map(entry=>entry.target as HTMLElement))){const record=records.get(body);if(record)apply(record)}})
  const sizes=new ResizeObserver(entries=>{for(const entry of entries){const record=records.get(entry.target as HTMLElement);if(record){record.width=record.body.clientWidth;record.height=record.body.clientHeight;apply(record)}}})
  function scan(){
   if(disposed)return
   for(const [body] of records)if(!body.isConnected)records.delete(body)
   for(const body of desktop.querySelectorAll<HTMLElement>(':scope>.app-window>.window-body'))if(!records.has(body)){
    const record={body,width:body.clientWidth,height:body.clientHeight};records.set(body,record);apply(record)
    mutations.observe(body,{attributes:true,attributeFilter:['style']});sizes.observe(body)
   }
  }
  const structure=new MutationObserver(scan);structure.observe(desktop,{childList:true})
  const preferences=new MutationObserver(()=>records.forEach(apply));preferences.observe(root,{attributes:true,attributeFilter:['data-translucency','data-theme','data-material-cache']})
  const dispose=()=>{disposed=true;mutations.disconnect();sizes.disconnect();structure.disconnect();preferences.disconnect();for(const {body} of records.values()){delete body.dataset.nativeBodyClip;body.style.removeProperty('--native-body-clip')}records.clear();style.remove()}
  const snapshot=()=>({tracked:records.size,enabled:[...records.values()].filter(r=>r.body.dataset.nativeBodyClip==='true').length,partial:[...records.values()].filter(r=>r.body.style.clipPath&&r.body.style.clipPath!=='inset(100%)').length,viewMasks:[...records.values()].flatMap(r=>[...r.body.querySelectorAll<HTMLElement>(':scope>.app-view')].map(view=>getComputedStyle(view).clip)).filter(value=>value!=='auto').length})
  ;(window as any).__nativeBodyClip={dispose,snapshot,setVisible(value:boolean){enabled=value;records.forEach(apply)}}
  window.addEventListener('pagehide',dispose,{once:true});scan()
 },flat)
}
