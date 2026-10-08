import type {Page} from '@playwright/test'

// Raster URLs are ephemeral resources, not wallpaper identity. Read the
// original stylesheet source atomically, without painting an uncached frame.
export async function appearance(page:Page){
 return page.evaluate(()=>{
  const root=document.documentElement,desktop=document.querySelector('.desktop')!,bar=document.querySelector('.topbar')!
  const ready=root.dataset.materialCache
  const displayed=getComputedStyle(desktop).backgroundImage
  const material=root.dataset.translucency==='on'&&ready==='ready'&&getComputedStyle(bar).backgroundImage.includes('blob:')?'cached-blur('+getComputedStyle(root).getPropertyValue('--material-blur')+')':getComputedStyle(bar).backdropFilter
  let wallpaper:string
  try{delete root.dataset.materialCache;wallpaper=getComputedStyle(desktop).backgroundImage}
  finally{if(ready!==undefined)root.dataset.materialCache=ready}
  return {primary:getComputedStyle(root).getPropertyValue('--primary').trim(),wallpaper,background:wallpaper,displayed,material}
 })
}
