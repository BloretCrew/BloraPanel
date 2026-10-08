// WebKit deliberately returns a fixed "Apple GPU" debug identifier on every
// platform (WebGLRenderingContextBase::getParameter). On Linux it cannot
// certify acceleration. Keep browser-owned default paint rather than opting
// into software GL; identified hardware and the existing Apple paths remain.
export function nativePaintFallback(userAgent:string,renderer:string|undefined,platform=''){
 if(renderer===undefined||/SwiftShader|llvmpipe|softpipe|software|Microsoft Basic Render/i.test(renderer))return true
 // A WebKit port can expose a Safari/macOS UA on Linux. Use the independent
 // platform hint too; no browser API or acceleration setting is overridden.
 const linuxWebKit=/AppleWebKit\//.test(userAgent)&&/Linux|X11/.test(platform+' '+userAgent)&&!/(?:Chrome|Chromium|Edg)\//.test(userAgent)
 return linuxWebKit&&renderer==='Apple GPU'
}
