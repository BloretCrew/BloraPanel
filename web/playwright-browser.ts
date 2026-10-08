import type {LaunchOptions} from '@playwright/test'

export function browserLaunchOptions(browser:'chromium'|'firefox'|'webkit',nativeScale?:number|string):LaunchOptions{
 if(browser!=='chromium')return {}
 return {
  ...(process.env.BLORA_CHROMIUM?{executablePath:process.env.BLORA_CHROMIUM}:process.platform==='linux'?{executablePath:'/usr/bin/chromium-browser'}:{}),
  args:['--no-sandbox','--disable-dev-shm-usage',...(nativeScale!==undefined&&process.env.BLORA_E08_EMULATED_SCALE_ONLY!=='1'?[`--force-device-scale-factor=${nativeScale}`]:[])],
 }
}

// CDP's emulated window DPR alone can disagree with Chromium's native
// device-pixel-content-box. Use the same physical launch scale for native
// paint/glyph guards; the explicit emulation-only control retains that failure.
// Real mixed-load acceptance does not import this helper or change its launch.
export function nativeDeviceScale(scale:number){
 // Browser launch options are worker-scoped. Physical DPR belongs in a
 // project, while each describe group retains its original context DPR.
 return {deviceScaleFactor:scale}
}
