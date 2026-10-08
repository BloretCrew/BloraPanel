import type {Page} from '@playwright/test'

export type NativeRetention='software'|'app-views'|'glass-app-views'
// Browser-owned paint retention only. No app/text/glyph/icon capture exists.
export async function installNativeRetention(page:Page,mode:NativeRetention){
 await page.evaluate(mode=>{
  const style=document.createElement('style')
  const rule=mode==='software'?'.app-window{will-change:auto!important}':(mode==='glass-app-views'?':root[data-translucency="on"][data-material-cache="ready"][data-material-cache-opaque="true"] ':'')+'.app-window:not(.focused)>.window-body>.app-view{will-change:transform}'
  const exactRule=`@media (resolution:1dppx){${rule}}`
  style.textContent=exactRule;document.head.append(style)
  ;(window as any).__nativeRetention={setVisible(value:boolean){style.textContent=value?exactRule:''},dispose(){style.remove()}}
 },mode)
}
