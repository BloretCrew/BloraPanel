import type {Page} from '@playwright/test'
import {installNativeFrameMaterial} from './native-frame-material'

// Test-only native paint structure. Keep each full opaque optical background
// independent of the original live app body's conservative exposure union.
// No app nodes/data/glyphs/images are copied, moved, sampled or rasterized.
export async function installNativeFrameRegions(page:Page){
 await installNativeFrameMaterial(page)
 await page.evaluate(()=>{
  type Rect={x:number;y:number;width:number;height:number}
  type Record={frame:HTMLElement;body:HTMLElement;box:Rect;area:Rect;x:number;y:number;width:string;height:string;dirty:boolean;regions?:Rect[]}
  const desktop=document.querySelector<HTMLElement>('.desktop')!,root=document.documentElement,records=new Map<HTMLElement,Record>(),base=(window as any).__nativeFrameMaterial,style=document.createElement('style')
  // Give the native app paint the identical opaque atlas backing directly.
  // Keep the full frame plane at joins; neither optical source contains app
  // contents. Independent retained backing changed SVG AA under a path mask.
  style.textContent='@property --native-frame-clip{syntax:"*";inherits:false;initial-value:none}.window-body[data-native-frame-regions="true"]{clip-path:var(--native-frame-clip)!important;background-image:var(--material-window)!important;background-size:var(--material-size);background-position:var(--material-x) var(--material-y);background-repeat:no-repeat;background-origin:border-box;background-color:var(--surface-container)!important}'
  document.head.append(style)
  let enabled=true,disposed=false
  const rect=(r:DOMRect):Rect=>({x:r.x,y:r.y,width:r.width,height:r.height})
  const contains=(a:Rect,b:Rect)=>a.x<=b.x&&a.y<=b.y&&a.x+a.width>=b.x+b.width&&a.y+a.height>=b.y+b.height
  function subtract(rect:Rect,covers:Rect[]){
   let pieces=[rect]
   for(const cover of covers){const next:Rect[]=[];for(const p of pieces){
    const x=Math.max(p.x,cover.x),y=Math.max(p.y,cover.y),right=Math.min(p.x+p.width,cover.x+cover.width),bottom=Math.min(p.y+p.height,cover.y+cover.height)
    if(x>=right||y>=bottom){next.push(p);continue}
    if(y>p.y)next.push({x:p.x,y:p.y,width:p.width,height:y-p.y})
    if(bottom<p.y+p.height)next.push({x:p.x,y:bottom,width:p.width,height:p.y+p.height-bottom})
    if(x>p.x)next.push({x:p.x,y,width:x-p.x,height:bottom-y})
    if(right<p.x+p.width)next.push({x:right,y,width:p.x+p.width-right,height:bottom-y})
   }pieces=next;if(pieces.length>128)return [rect]}
   return pieces
  }
  function retain(area:Rect,old:Rect[]|undefined,visible:Rect[]){
   let regions=old?[...old]:[]
   for(const part of visible){
    const local={...part,x:part.x-area.x,y:part.y-area.y}
    if(regions.some(r=>contains(r,local)))continue
    const x=Math.max(0,local.x-64),y=Math.max(0,local.y-64),right=Math.min(area.width,local.x+local.width+64),bottom=Math.min(area.height,local.y+local.height+64),expanded={x,y,width:right-x,height:bottom-y}
    regions=regions.filter(r=>!contains(expanded,r));regions.push(expanded)
    if(regions.length>32)return [{x:0,y:0,width:area.width,height:area.height}]
   }
   return regions
  }
  function clear(record:Record){delete record.body.dataset.nativeFrameRegions;record.body.style.removeProperty('--native-frame-clip');record.regions=undefined}
  const sizes=new ResizeObserver(entries=>{for(const entry of entries)for(const record of records.values())if(entry.target===record.frame||entry.target===record.body)record.dirty=true;update()})
  const changes=new MutationObserver(update)
  function scan(){
   if(disposed)return
   for(const [frame,record] of records)if(!frame.isConnected){clear(record);sizes.unobserve(frame);sizes.unobserve(record.body);records.delete(frame)}
   changes.disconnect()
   for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
    const body=frame.querySelector<HTMLElement>(':scope>.window-body');if(!body)continue
    if(!records.has(frame)){records.set(frame,{frame,body,box:{x:0,y:0,width:0,height:0},area:{x:0,y:0,width:0,height:0},x:0,y:0,width:'',height:'',dirty:true});sizes.observe(frame);sizes.observe(body)}
    changes.observe(frame,{attributes:true,attributeFilter:['style','class']})
   }
   update()
  }
  function update(){
   if(disposed)return
   const active=enabled&&Number.isInteger(devicePixelRatio||1)&&root.dataset.translucency==='on'&&root.dataset.materialCache==='ready'&&root.dataset.materialCacheOpaque==='true'&&!/(?:Chrome|Chromium|Edg)\//.test(navigator.userAgent)
   if(!active){for(const record of records.values())clear(record);return}
   // Read native geometry before changing any masks. Pure translation updates
   // cached world boxes arithmetically; terminal row mutations are unobserved.
   const windows=[...records.values()].map(record=>{
    const {frame}=record,translation=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(frame.style.transform),x=translation?Number(translation[1]):0,y=translation?Number(translation[2]):0
    if(record.dirty||!translation||record.width!==frame.style.width||record.height!==frame.style.height){record.box=rect(frame.getBoundingClientRect());record.area=rect(record.body.getBoundingClientRect());record.width=frame.style.width;record.height=frame.style.height;record.regions=undefined;record.dirty=false}
    else for(const r of [record.box,record.area]){r.x+=x-record.x;r.y+=y-record.y}
    record.x=x;record.y=y
    const css=getComputedStyle(frame);return {record,z:Number(css.zIndex)||0,display:css.display,radius:parseFloat(css.borderTopLeftRadius)||0,solid:css.visibility==='visible'&&Number(css.opacity)===1}
   }).filter(window=>window.display!=='none').sort((a,b)=>b.z-a.z)
   const moving=windows.some(window=>window.record.frame.classList.contains('moving')),covers:Rect[]=[]
   for(const window of windows){
    const {record}=window,{box,area,body}=record
    const regions=retain(area,moving?record.regions:undefined,subtract(area,covers));record.regions=regions
    const clip=regions.length?regions.some(r=>r.x===0&&r.y===0&&r.width===area.width&&r.height===area.height)?'none':`path('${regions.map(r=>`M${r.x} ${r.y}h${r.width}v${r.height}h${-r.width}Z`).join(' ')}')`:'inset(100%)'
    if(body.style.getPropertyValue('--native-frame-clip')!==clip)body.style.setProperty('--native-frame-clip',clip)
    body.dataset.nativeFrameRegions='true'
    const p=Math.max(1,Math.min(window.radius,box.width/2,box.height/2))
    if(window.solid&&box.width>p*2&&box.height>2)covers.push({x:box.x+p,y:box.y+1,width:box.width-p*2,height:box.height-2})
    if(window.solid&&box.width>2&&box.height>p*2)covers.push({x:box.x+1,y:box.y+p,width:box.width-2,height:box.height-p*2})
   }
  }
  const children=new MutationObserver(scan);children.observe(desktop,{childList:true})
  const appearance=new MutationObserver(()=>{for(const record of records.values())record.dirty=true;update()});appearance.observe(root,{attributes:true,attributeFilter:['data-translucency','data-material-cache','data-material-cache-opaque','data-theme','data-palette','data-density']})
  const resize=()=>{for(const record of records.values())record.dirty=true;update()};window.addEventListener('resize',resize)
  const controller={snapshot(){const value=base.snapshot();return {...value,regionMasks:[...records.values()].filter(r=>r.body.dataset.nativeFrameRegions==='true').length,rectangles:[...records.values()].reduce((n,r)=>n+(r.regions?.length||0),0),nativeArea:[...records.values()].reduce((n,r)=>n+r.area.width*r.area.height,0),regionArea:[...records.values()].reduce((n,r)=>n+(r.regions||[]).reduce((s,p)=>s+p.width*p.height,0),0)}},setVisible(value:boolean){enabled=value;base.setVisible(value);update()},dispose(){if(disposed)return;disposed=true;changes.disconnect();children.disconnect();appearance.disconnect();sizes.disconnect();window.removeEventListener('resize',resize);for(const record of records.values())clear(record);records.clear();style.remove();base.dispose()}}
  ;(window as any).__nativeFrameMaterial=controller
  window.addEventListener('pagehide',controller.dispose,{once:true});scan()
 })
}
