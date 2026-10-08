// Diagnostic only: resample the existing wallpaper-only material. Never
// inspect/copy/paint app DOM, terminal data, glyphs, icons or live surfaces.
export async function installSmallMaterialAtlas(options?:{width?:number;opaque?:boolean;surface?:'window'|'wallpaper'|'desktop'|'face'|'menu'}){
 const root=document.documentElement,property=`--material-${options?.surface||'window'}`,original=root.style.getPropertyValue(property)
 const match=/^url\("([^\"]+)"\)$/.exec(original.trim())
 if(!match||root.dataset.materialCache!=='ready')throw Error('native wallpaper material unavailable')
 const image=new Image();image.src=match[1]!;await image.decode()
 const width=Math.min(options?.width||image.naturalWidth,image.naturalWidth),height=Math.max(1,Math.round(image.naturalHeight*width/image.naturalWidth))
 const canvas=document.createElement('canvas');canvas.width=width;canvas.height=height
 const x=canvas.getContext('2d',{alpha:root.dataset.materialCacheOpaque!=='true'});if(!x)throw Error('wallpaper diagnostic canvas unavailable')
 x.drawImage(image,0,0,width,height)
 let blob:Blob
 if(options?.opaque){
  // BMP24 has no alpha channel. Certify opacity before encoding only the
  // existing wallpaper samples; never flatten a transparent material.
  const pixels=x.getImageData(0,0,width,height).data,stride=Math.ceil(width*3/4)*4,bytes=new Uint8Array(54+stride*height),header=new DataView(bytes.buffer)
  bytes[0]=66;bytes[1]=77;header.setUint32(2,bytes.length,true);header.setUint32(10,54,true);header.setUint32(14,40,true);header.setInt32(18,width,true);header.setInt32(22,height,true);header.setUint16(26,1,true);header.setUint16(28,24,true);header.setUint32(34,stride*height,true)
  for(let row=0;row<height;row++)for(let column=0;column<width;column++){
   const source=(row*width+column)*4,target=54+(height-1-row)*stride+column*3
   if(pixels[source+3]!==255)throw Error('wallpaper material is not opaque')
   bytes[target]=pixels[source+2]!;bytes[target+1]=pixels[source+1]!;bytes[target+2]=pixels[source]!
  }
  blob=new Blob([bytes],{type:'image/bmp'})
 }else blob=await new Promise<Blob>((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(Error('wallpaper encoding unavailable'))))
 const bytes=new Uint8Array(await blob.arrayBuffer()),url=URL.createObjectURL(blob),decoded=new Image();decoded.src=url;await decoded.decode()
 root.style.setProperty(property,`url("${url}")`)
 const metadata={surface:options?.surface||'window',sourceWidth:image.naturalWidth,sourceHeight:image.naturalHeight,width,height,encoding:options?.opaque?'opaque-bmp':'png',pngColourType:options?.opaque?undefined:bytes[25]}
 const controller={metadata,restore(){if(root.style.getPropertyValue(property)===`url("${url}")`)root.style.setProperty(property,original)},dispose(){controller.restore();URL.revokeObjectURL(url)}}
 ;(window as any).__smallMaterialAtlas=controller
 window.addEventListener('pagehide',controller.dispose,{once:true})
 return metadata
}
