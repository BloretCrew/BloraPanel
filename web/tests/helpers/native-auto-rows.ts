import type {Page} from '@playwright/test'

// Native browser relevance, not a custom terminal renderer. Official xterm
// still parses/commits every byte and creates all original rows and glyphs.
// No per-output-row observers or private render-service adapters are used.
export async function installNativeAutoRows(page:Page){
 await page.evaluate(()=>{
  const hosts=new Map<HTMLElement,{rows:HTMLElement;release:()=>void}>(),style=document.createElement('style')
  style.textContent='@media (resolution:1dppx){.terminal-container[data-native-auto-rows="true"] .xterm-rows>div{content-visibility:auto;contain-intrinsic-block-size:auto var(--native-auto-row-height)}}'
  document.head.append(style)
  let enabled=true,disposed=false
  function apply(host:HTMLElement,rows:HTMLElement){
   const row=rows.firstElementChild,active=enabled&&devicePixelRatio===1&&CSS.supports('content-visibility','auto')&&host.dataset.terminalRenderer==='default'&&row instanceof HTMLElement
   if(!active){delete host.dataset.nativeAutoRows;host.style.removeProperty('--native-auto-row-height');return}
   const height=getComputedStyle(row).height
   if(!/^\d+(?:\.\d+)?px$/.test(height)||parseFloat(height)<=0){delete host.dataset.nativeAutoRows;return}
   if(host.style.getPropertyValue('--native-auto-row-height')!==height)host.style.setProperty('--native-auto-row-height',height)
   host.dataset.nativeAutoRows='true'
  }
  function scan(){
   if(disposed)return
   for(const [host,record] of hosts)if(!host.isConnected||!record.rows.isConnected){record.release();hosts.delete(host)}
   for(const host of document.querySelectorAll<HTMLElement>('.terminal-container')){
    if(hosts.has(host))continue
    const rows=host.querySelector<HTMLElement>('.xterm-rows');if(!rows)continue
    const resize=new ResizeObserver(()=>apply(host,rows));resize.observe(host);resize.observe(rows)
    const attrs=new MutationObserver(()=>apply(host,rows));attrs.observe(host,{attributes:true,attributeFilter:['data-terminal-renderer']})
    hosts.set(host,{rows,release(){resize.disconnect();attrs.disconnect();delete host.dataset.nativeAutoRows;host.style.removeProperty('--native-auto-row-height')}});apply(host,rows)
   }
  }
  // Host lifecycle only; output/span mutations never trigger a rescan.
  const desktop=document.querySelector('.desktop')||document.body,structure=new MutationObserver(scan);structure.observe(desktop,{childList:true})
  desktop.addEventListener('terminal-paint-host-changed',scan)
  const resize=()=>{for(const [host,record] of hosts)apply(host,record.rows)};window.addEventListener('resize',resize);scan()
  const dispose=()=>{disposed=true;structure.disconnect();desktop.removeEventListener('terminal-paint-host-changed',scan);window.removeEventListener('resize',resize);for(const record of hosts.values())record.release();hosts.clear();style.remove()}
  ;(window as any).__nativeAutoRows={setVisible(value:boolean){enabled=value;resize()},snapshot(){const rows=[...hosts.values()].flatMap(record=>[...record.rows.children] as HTMLElement[]);return {tracked:hosts.size,active:[...hosts.keys()].filter(host=>host.dataset.nativeAutoRows==='true').length,rows:rows.length,automatic:rows.filter(row=>getComputedStyle(row).contentVisibility==='auto').length,skipped:rows.filter(row=>typeof row.checkVisibility==='function'&&!row.checkVisibility({contentVisibilityAuto:true})).length}},dispose}
  window.addEventListener('pagehide',dispose,{once:true})
 })
}
