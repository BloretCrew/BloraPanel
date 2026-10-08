// Only wallpaper-derived optical canvases use this encoder. Live application
// content, glyphs, icons, terminal output and native shadows never enter it.
// Explicit RGB24 avoids an alpha-bearing PNG texture for a certified opaque
// surface. The native browser decoder preserves the original RGB samples.
export function encodeMaterial(canvas:HTMLCanvasElement,opaque:boolean):Promise<Blob>{
 if(!opaque)return new Promise((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(Error('material raster unavailable'))))
 const context=canvas.getContext('2d');if(!context)return Promise.reject(Error('material canvas unavailable'))
 const {width,height}=canvas,pixels=context.getImageData(0,0,width,height).data,stride=Math.ceil(width*3/4)*4
 const bytes=new Uint8Array(54+stride*height),header=new DataView(bytes.buffer)
 bytes[0]=66;bytes[1]=77;header.setUint32(2,bytes.length,true);header.setUint32(10,54,true);header.setUint32(14,40,true);header.setInt32(18,width,true);header.setInt32(22,height,true);header.setUint16(26,1,true);header.setUint16(28,24,true);header.setUint32(34,stride*height,true)
 for(let row=0;row<height;row++)for(let column=0;column<width;column++){
  const source=(row*width+column)*4,target=54+(height-1-row)*stride+column*3
  // A future source change must fall back instead of flattening transparency.
  if(pixels[source+3]!==255)return encodeMaterial(canvas,false)
  bytes[target]=pixels[source+2]!;bytes[target+1]=pixels[source+1]!;bytes[target+2]=pixels[source]!
 }
 return Promise.resolve(new Blob([bytes],{type:'image/bmp'}))
}
