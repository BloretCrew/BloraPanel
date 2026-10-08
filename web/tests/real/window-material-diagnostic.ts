import type {Page} from '@playwright/test'

// Resample just the shared window optical atlas once. No application DOM,
// glyph, icon, canvas or data ever enters this diagnostic material.
export async function installWindowMaterialDiagnostic(page:Page,encoding:'native-png'|'opaque-rgb24'='native-png'){
 return page.evaluate(async encoding=>{
  const root=document.documentElement
  ;(window as any).__windowMaterialDiagnostic?.dispose()
  if(root.dataset.materialCache!=='ready'||root.dataset.materialCacheOpaque!=='true')throw Error('Decoded opaque material required')
  const original=root.style.getPropertyValue('--material-window'),match=/url\("([^"]+)"\)/.exec(original)
  if(!match)throw Error('Shared native window atlas required')
  const source=new Image();source.src=match[1]!;await source.decode()
  const size=getComputedStyle(root).getPropertyValue('--material-size').match(/[\d.]+/g)!.map(Number),ratio=devicePixelRatio||1
  const width=Math.ceil(size[0]!*ratio),height=Math.ceil(size[1]!*ratio)
  const canvas=document.createElement('canvas');canvas.width=width;canvas.height=height
  const context=canvas.getContext('2d',{alpha:false});if(!context)throw Error('Native optical context unavailable')
  context.drawImage(source,0,0,width,height)
  let blob:Blob
  if(encoding==='opaque-rgb24'){
   const pixels=context.getImageData(0,0,width,height).data,stride=Math.ceil(width*3/4)*4,bytes=new Uint8Array(54+stride*height),header=new DataView(bytes.buffer)
   bytes[0]=66;bytes[1]=77;header.setUint32(2,bytes.length,true);header.setUint32(10,54,true);header.setUint32(14,40,true);header.setInt32(18,width,true);header.setInt32(22,height,true);header.setUint16(26,1,true);header.setUint16(28,24,true);header.setUint32(34,stride*height,true)
   for(let row=0;row<height;row++)for(let column=0;column<width;column++){
    const a=(row*width+column)*4,b=54+(height-1-row)*stride+column*3
    if(pixels[a+3]!==255)throw Error('Native window material must remain opaque')
    bytes[b]=pixels[a+2]!;bytes[b+1]=pixels[a+1]!;bytes[b+2]=pixels[a]!
   }
   blob=new Blob([bytes],{type:'image/bmp'})
  }else blob=await new Promise<Blob>((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(Error('Native optical encoding failed'))))
  const candidate=URL.createObjectURL(blob),decoded=new Image();decoded.src=candidate
  try{await decoded.decode()}catch(error){URL.revokeObjectURL(candidate);throw error}
  const value=`url("${candidate}")`;root.style.setProperty('--material-window',value)
  let disposed=false
  const controller={dispose(){if(disposed)return;disposed=true;if(root.style.getPropertyValue('--material-window')===value)root.style.setProperty('--material-window',original);URL.revokeObjectURL(candidate)}}
  ;(window as any).__windowMaterialDiagnostic=controller
  window.addEventListener('pagehide',controller.dispose,{once:true})
  // Do not fetch the original blob: its image permission does not imply a
  // network/connect permission under the real production CSP.
  const candidateHeader=new Uint8Array(await blob.arrayBuffer()).slice(0,32)
  return {sourceWidth:source.naturalWidth,sourceHeight:source.naturalHeight,width,height,encoding:blob.type,candidatePngColourType:blob.type==='image/png'?candidateHeader[25]:undefined}
 },encoding)
}
