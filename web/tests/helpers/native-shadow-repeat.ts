import type {Page} from '@playwright/test'

// Test-only: use the SAME decoded optical asset as a constant-axis native
// repeat brush. No application nodes/pixels, new images or glyph caches.
export async function installNativeShadowRepeat(page:Page){
 await page.evaluate(async()=>{
  type Record={element:HTMLElement;axis:'x'|'y';original:string|undefined}
  const records:Record[]=[],certified=new Map<string,Promise<boolean>>()
  const style=document.createElement('style')
  style.textContent='.shadow-piece[data-native-shadow-repeat="x"]{background-size:auto 100%!important;background-repeat:repeat-x!important}.shadow-piece[data-native-shadow-repeat="y"]{background-size:100% auto!important;background-repeat:repeat-y!important}'
  document.head.append(style)
  const certify=(url:string,horizontal:boolean)=>{
   const key=JSON.stringify([url,horizontal])
   if(!certified.has(key))certified.set(key,(async()=>{
    const image=new Image();image.src=url;await image.decode()
    const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height
    const context=canvas.getContext('2d')!;context.drawImage(image,0,0)
    const pixels=context.getImageData(0,0,image.width,image.height).data
    for(let y=0;y<image.height;y++)for(let x=0;x<image.width;x++)for(let c=0;c<4;c++){
     const first=(horizontal?y*image.width:x)*4+c
     if(pixels[(y*image.width+x)*4+c]!==pixels[first])return false
    }
    return true
   })())
   return certified.get(key)!
  }
  if(devicePixelRatio===1)for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane[data-shadow-cache]'))for(const name of ['n','s','w','e']){
   const element=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!,horizontal=name==='n'||name==='s'
   let valid=true
   for(const state of ['normal','focused']){
    if(element.style.getPropertyValue('--shadow-'+state+'-display')==='none')continue
    const url=/^url\("([^"]+)"\)$/.exec(element.style.getPropertyValue('--shadow-'+state))?.[1]
    if(!url||!await certify(url,horizontal)){valid=false;break}
   }
   if(valid)records.push({element,axis:horizontal?'x':'y',original:element.dataset.nativeShadowRepeat})
  }
  const toggle=(enabled:boolean)=>{for(const record of records){if(enabled)record.element.dataset.nativeShadowRepeat=record.axis;else if(record.original)record.element.dataset.nativeShadowRepeat=record.original;else delete record.element.dataset.nativeShadowRepeat}}
  toggle(true)
  const dispose=()=>{toggle(false);style.remove();certified.clear()}
  ;(window as any).__nativeShadowRepeat={snapshot:()=>({edges:records.length,assets:certified.size}),setVisible:toggle,dispose}
  window.addEventListener('pagehide',dispose,{once:true})
 })
}
