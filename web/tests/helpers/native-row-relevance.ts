import type {Page} from '@playwright/test'

// Diagnostic only. Native xterm still updates every original node and parses
// every byte. Only wholly clipped row subtrees lose browser layout/paint work;
// their unchanged outer line boxes remain in the native terminal geometry.
export async function installNativeRowRelevance(page:Page,control=false){
 await page.evaluate(control=>{
  type Saved={visibility:string;intrinsic:string}
  const saved=new Map<HTMLElement,Saved>(),hosts=new Map<HTMLElement,()=>void>()
  let enabled=true,disposed=false,updates=0
  const restore=(row:HTMLElement)=>{const old=saved.get(row);if(!old)return;row.style.contentVisibility=old.visibility;row.style.containIntrinsicSize=old.intrinsic;saved.delete(row)}
  function update(host:HTMLElement){
   if(disposed)return
   for(const row of saved.keys())if(!row.isConnected)restore(row)
   const viewport=host.closest<HTMLElement>('.terminal-paint-viewport'),screen=host.querySelector<HTMLElement>('.xterm-screen'),rows=host.querySelector<HTMLElement>('.xterm-rows')
   if(!viewport||!screen||!rows)return
   const children=[...rows.children].filter((element):element is HTMLElement=>element instanceof HTMLElement)
   const active=!control&&enabled&&CSS.supports('content-visibility','hidden')&&Number.isInteger(devicePixelRatio||1)&&host.dataset.terminalRenderer==='default'&&host.dataset.paintPartial==='true'
   if(!active){for(const row of children)restore(row);return}
   // Take every native row measurement before changing any containment. A
   // two-row vertical and two-pixel horizontal guard retain glyph overhang,
   // cursor/underline and fractional edge antialiasing around the true clip.
   const clip=viewport.getBoundingClientRect(),boxes=children.map(row=>row.getBoundingClientRect()),heights=children.map(row=>getComputedStyle(row).height),height=boxes[0]?.height||0
   if(!height||boxes.some(box=>Math.abs(box.height-height)>.01)){for(const row of children)restore(row);return}
   for(let i=0;i<children.length;i++){
    const row=children[i]!,box=boxes[i]!
    const hidden=box.bottom<clip.top-2*height||box.top>clip.bottom+2*height||box.right<clip.left-2||box.left>clip.right+2
    if(hidden){
     if(!saved.has(row))saved.set(row,{visibility:row.style.contentVisibility,intrinsic:row.style.containIntrinsicSize})
     const intrinsic=`none ${heights[i]}`
     if(row.style.containIntrinsicSize!==intrinsic)row.style.containIntrinsicSize=intrinsic
     if(row.style.contentVisibility!=='hidden')row.style.contentVisibility='hidden'
    }else restore(row)
   }
   updates++
  }
  function scan(){
   for(const [host,release] of hosts)if(!host.isConnected){release();hosts.delete(host)}
   for(const host of document.querySelectorAll<HTMLElement>('.terminal-container')){
    if(hosts.has(host))continue
    const viewport=host.closest<HTMLElement>('.terminal-paint-viewport'),rows=host.querySelector<HTMLElement>('.xterm-rows')
    if(!viewport||!rows)continue
    const changed=new MutationObserver(()=>update(host));changed.observe(viewport,{attributes:true,attributeFilter:['style']});changed.observe(host,{attributes:true,attributeFilter:['style','data-paint-partial','data-paint-occluded','data-terminal-renderer']});changed.observe(rows,{childList:true})
    const sizes=new ResizeObserver(()=>update(host));sizes.observe(viewport);sizes.observe(host)
    hosts.set(host,()=>{changed.disconnect();sizes.disconnect();for(const row of rows.children)if(row instanceof HTMLElement)restore(row)})
    update(host)
   }
  }
  const desktop=document.querySelector('.desktop')||document.body
  // Never observe PTY text/span mutations across the application tree.
  const children=new MutationObserver(scan);children.observe(desktop,{childList:true});desktop.addEventListener('terminal-paint-host-changed',scan);scan()
  ;(window as any).__nativeRowRelevance={
   setVisible(value:boolean){enabled=value;for(const host of hosts.keys())update(host)},
   snapshot(){return {tracked:hosts.size,hiddenRows:[...saved.keys()].filter(row=>row.isConnected&&row.style.contentVisibility==='hidden').length,totalRows:[...hosts.keys()].reduce((sum,host)=>sum+host.querySelectorAll('.xterm-rows>div').length,0),updates}},
   dispose(){if(disposed)return;enabled=false;for(const host of hosts.keys())update(host);disposed=true;children.disconnect();desktop.removeEventListener('terminal-paint-host-changed',scan);for(const release of hosts.values())release();hosts.clear();for(const row of saved.keys())restore(row)},
  }
 },control)
}
