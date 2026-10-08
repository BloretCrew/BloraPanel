import {test,expect} from 'vitest'
import {blurMaterial} from '../src/appearance/material-blur'
test('bounded material convolution retains an opaque constant through edges',()=>{
 const original=new Uint8ClampedArray(20*12*4)
 for(let i=0;i<original.length;i+=4)original.set([40,100,180,255],i)
 expect(blurMaterial(original,20,12,6)).toEqual(original)
})
test('material blur spreads a bright feature symmetrically without alpha halos',()=>{
 const original=new Uint8ClampedArray(21*21*4)
 for(let i=3;i<original.length;i+=4)original[i]=255
 original[(10*21+10)*4]=255
 const result=blurMaterial(original,21,21,2)
 const at=(x:number,y:number)=>result[(y*21+x)*4]!
 expect(at(10,10)).toBeGreaterThan(0);expect(at(10,10)).toBeLessThan(255)
 expect(at(9,10)).toBe(at(11,10));expect(at(10,9)).toBe(at(10,11))
 for(let i=3;i<result.length;i+=4)expect(result[i]).toBe(255)
 expect(original[(10*21+10)*4]).toBe(255)
})
