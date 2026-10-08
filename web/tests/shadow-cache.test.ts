import {test,expect} from 'vitest'
import {shadowGeometry,shadowPixelBounds,scaleShadowPixels} from '../src/appearance/shadow-cache'
test('native shadow raster scales physical lengths without changing colours or layer order',()=>{
 const shadow='rgba(76, 107, 138, .09) -2px 10px 30px 0px, color(srgb .2 .3 .4 / .04) 0px 2px 6px -1px'
 expect(scaleShadowPixels(shadow,1)).toBe(shadow)
 expect(scaleShadowPixels(shadow,1.25)).toBe('rgba(76, 107, 138, .09) -2.5px 12.5px 37.5px 0px, color(srgb .2 .3 .4 / .04) 0px 2.5px 7.5px -1.25px')
 expect(scaleShadowPixels('none',2)).toBe('none')
})
test('shadow patch guard accounts for every layer, offset, spread and contour',()=>{
 const g=shadowGeometry('rgba(76, 107, 138, .09) 0px 10px 30px 0px, rgba(76, 107, 138, .04) 0px 2px 6px 0px','rgba(69, 101, 128, .14) 0px 20px 56px 0px',32)!
 expect(g.pad).toBeGreaterThanOrEqual(84+20);expect(g.corner).toBeGreaterThanOrEqual(32+84+20)
 expect(g.size-2*g.slice).toBe(2);expect(g.box-2*g.corner).toBe(2)
})
test('unsupported or excessive shadows keep native painting',()=>{
 for(const shadow of ['inset rgb(0, 0, 0) 0px 1px 10px','rgb(0, 0, 0) 0px 0px 1000px','rgb(0, 0, 0) 0px 0px -1px','invalid'])expect(shadowGeometry(shadow,'none',32)).toBeUndefined()
 expect(shadowGeometry('none','none',NaN)).toBeUndefined()
 expect(shadowGeometry('color(srgb 0 0 0 / .1) -2px 4px 8px -3px','none',16)).toBeDefined()
})
test('common shadow crop retains every nonzero normal and focused alpha, including alpha one',()=>{
 const width=23,height=19,normal=new Uint8ClampedArray(width*height*4),focused=new Uint8ClampedArray(normal.length)
 normal[(3*width+2)*4+3]=1;focused[(15*width+20)*4+3]=255
 expect(shadowPixelBounds(normal,focused,width,height)).toEqual({x:2,y:3,width:19,height:13})
 // Independent elevation crops preserve the native alpha-one pixel without
 // allocating the other state's larger transparent rectangle.
 expect(shadowPixelBounds(normal,normal,width,height)).toEqual({x:2,y:3,width:1,height:1})
 expect(shadowPixelBounds(focused,focused,width,height)).toEqual({x:20,y:15,width:1,height:1})
 let seed=1973
 for(let sample=0;sample<50;sample++){
  normal.fill(0);focused.fill(0)
  for(const raster of [normal,focused])for(let point=0;point<10;point++){
   seed=(Math.imul(seed,1664525)+1013904223)>>>0;const x=seed%width
   seed=(Math.imul(seed,1664525)+1013904223)>>>0;const y=seed%height
   raster[(y*width+x)*4+3]=point+1
  }
  const bounds=shadowPixelBounds(normal,focused,width,height)!
  for(let y=0;y<height;y++)for(let x=0;x<width;x++){
   if(x>=bounds.x&&x<bounds.x+bounds.width&&y>=bounds.y&&y<bounds.y+bounds.height)continue
   expect(normal[(y*width+x)*4+3]).toBe(0);expect(focused[(y*width+x)*4+3]).toBe(0)
  }
 }
 normal.fill(0);focused.fill(0);expect(shadowPixelBounds(normal,focused,width,height)).toBeUndefined()
})
