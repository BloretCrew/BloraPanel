import {nativePaintFallback} from '../services/rendering-capabilities'
// Hardware composition keeps the existing transform path. Software native
// painting can use absolute offsets with an identity transform, preserving
// the window's stacking/containing context and all original DOM contents.
export function offsetWindowRenderer(userAgent:string,renderer:string|undefined,platform=''){
 if(!/Firefox\/|AppleWebKit\//.test(userAgent)||/(?:Chrome|Chromium|Edg)\//.test(userAgent))return false
 return nativePaintFallback(userAgent,renderer,platform)
}
let measured:boolean|undefined
export function nativeWindowOffsets(){
 if(measured!==undefined)return measured
 if(!/Firefox\/|AppleWebKit\//.test(navigator.userAgent)||/(?:Chrome|Chromium|Edg)\//.test(navigator.userAgent))return measured=false
 let renderer:string|undefined
 try{
  const probe=document.createElement('canvas').getContext('webgl2',{failIfMajorPerformanceCaveat:true})
  if(probe){
   try{const extension=probe.getExtension('WEBGL_debug_renderer_info');renderer=extension?String(probe.getParameter(extension.UNMASKED_RENDERER_WEBGL)):''}
   finally{probe.getExtension('WEBGL_lose_context')?.loseContext()}
  }
 }catch{/* No supported hardware context: keep all paint browser-owned. */}
 return measured=offsetWindowRenderer(navigator.userAgent,renderer,navigator.platform)
}
