import type {Page} from '@playwright/test'

// Explicit structural diagnostic only. Keep the actual Vue app nodes and
// dimensions; neither serialize nor rasterize any app text/data/contents.
export async function installBodyViewportDiagnostic(page:Page){
 await page.evaluate(()=>{
  const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!
  const style=document.createElement('style');style.textContent='.window-body.native-body-viewport{clip-path:none!important}.window-body.native-body-viewport::before{display:none!important}.native-paint-viewport{position:absolute;overflow:hidden}.native-paint-content{position:absolute}.native-body-material{position:absolute;inset:0;z-index:-1;pointer-events:none;background-image:var(--material-window);background-size:var(--material-size);background-repeat:no-repeat;background-origin:border-box;will-change:transform}'
  document.head.append(style)
  const records:Array<{body:HTMLElement;viewport:HTMLElement;content:HTMLElement;material:HTMLElement;width:number;height:number}>=[]
  const set=(element:HTMLElement,name:string,value:string)=>{if(element.style.getPropertyValue(name)!==value)element.style.setProperty(name,value)}
  function apply(record:typeof records[number]){
   const {body,viewport,content,material,width,height}=record,clip=body.style.clipPath
   const match=/^inset\(([-\d.]+)px ([-\d.]+)px ([-\d.]+)px ([-\d.]+)px\)$/.exec(clip)
   const [top,right,bottom,left]=clip==='inset(100%)'?[height,width,height,width]:match?match.slice(1).map(Number):[0,0,0,0]
   set(viewport,'inset',`${top}px ${right}px ${bottom}px ${left}px`)
   set(content,'left',`${-left!}px`);set(content,'top',`${-top!}px`);set(content,'width',`${width}px`);set(content,'height',`${height}px`)
   const ready=root.dataset.translucency==='on'&&root.dataset.materialCache==='ready'
   set(material,'display',ready?'block':'none')
   set(material,'background-position',`${body.style.getPropertyValue('--material-x')||'0px'} ${body.style.getPropertyValue('--material-y')||'0px'}`)
   set(material,'background-color',root.dataset.materialCacheOpaque==='true'?'var(--surface-container)':'transparent')
  }
  const sizes=new ResizeObserver(entries=>{for(const entry of entries){const record=records.find(value=>value.body===entry.target);if(record){record.width=record.body.clientWidth;record.height=record.body.clientHeight;apply(record)}}})
  const mutations=new MutationObserver(entries=>{for(const body of new Set(entries.map(value=>value.target as HTMLElement))){const record=records.find(value=>value.body===body);if(record)apply(record)}})
  for(const body of desktop.querySelectorAll<HTMLElement>(':scope>.app-window>.window-body')){
   const viewport=document.createElement('div'),content=document.createElement('div'),material=document.createElement('div')
   viewport.className='native-paint-viewport';content.className='native-paint-content';material.className='native-body-material';material.setAttribute('aria-hidden','true')
   const children=[...body.children];for(const child of children)content.append(child)
   content.prepend(material);viewport.append(content);body.append(viewport);body.classList.add('native-body-viewport')
   const record={body,viewport,content,material,width:body.clientWidth,height:body.clientHeight};records.push(record);apply(record)
   sizes.observe(body);mutations.observe(body,{attributes:true,attributeFilter:['style']})
  }
  const appearance=new MutationObserver(()=>records.forEach(apply));appearance.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-material-cache-opaque','data-translucency']})
  window.addEventListener('pagehide',()=>{sizes.disconnect();mutations.disconnect();appearance.disconnect()},{once:true})
 })
}
