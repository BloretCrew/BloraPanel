import {test,expect} from 'vitest'
import {exposedBounds,exposedRegions,exposureClip,retainExposure,retainExposureBounds} from '../src/desktop/occlusion'
const rect={x:0,y:0,width:100,height:100}
test('opaque cover only clips the actually covered side',()=>{
 expect(exposedBounds(rect,[{x:20,y:-1,width:120,height:102}])).toEqual({x:0,y:0,width:20,height:100})
 expect(exposedBounds(rect,[{x:-1,y:-1,width:102,height:102}])).toBeUndefined()
})
test('disjoint visible sides and overlapping opaque covers stay conservative',()=>{
 expect(exposedBounds(rect,[{x:25,y:-1,width:50,height:102}])).toEqual(rect)
 expect(exposedBounds(rect,[{x:20,y:-1,width:50,height:102},{x:50,y:-1,width:70,height:102}])).toEqual({x:0,y:0,width:20,height:100})
 expect(exposedBounds(rect,[{x:100,y:0,width:20,height:100}])).toEqual(rect)
})
test('every uncovered pixel centre remains inside conservative exposure bounds',()=>{
 let seed=1701
 const random=()=>{seed=(Math.imul(seed,1664525)+1013904223)>>>0;return seed/4294967296}
 for(let sample=0;sample<60;sample++){
  const covers=Array.from({length:8},()=>({x:random()*140-20,y:random()*140-20,width:random()*110,height:random()*110}))
  const visible=exposedBounds(rect,covers),regions=exposedRegions(rect,covers)
  for(let y=.5;y<100;y+=5)for(let x=.5;x<100;x+=5){
   if(covers.some(c=>x>=c.x&&x<c.x+c.width&&y>=c.y&&y<c.y+c.height))continue
   expect(regions.some(p=>x>=p.x&&x<p.x+p.width&&y>=p.y&&y<p.y+p.height)).toBe(true)
   expect(visible,`uncovered ${x},${y}`).toBeDefined()
   expect(x).toBeGreaterThanOrEqual(visible!.x);expect(x).toBeLessThanOrEqual(visible!.x+visible!.width)
   expect(y).toBeGreaterThanOrEqual(visible!.y);expect(y).toBeLessThanOrEqual(visible!.y+visible!.height)
  }
 }
})
test('separate exposed strips omit the hidden centre without changing their bounds',()=>{
 const covers=[{x:10,y:10,width:100,height:100}],regions=exposedRegions(rect,covers)
 expect(exposedBounds(rect,covers)).toEqual(rect)
 expect(regions.reduce((sum,p)=>sum+p.width*p.height,0)).toBe(1900)
 expect(regions.some(p=>50>=p.x&&50<p.x+p.width&&50>=p.y&&50<p.y+p.height)).toBe(false)
})
test('separate exposure masks retain visible strips without repainting the enclosing hidden centre',()=>{
 const regions=exposedRegions(rect,[{x:10,y:10,width:100,height:100}])
 expect(exposureClip(rect,regions)).toBe("path('M0 0h100v10h-100Z M0 10h10v90h-10Z')")
 expect(exposureClip(rect,[])).toBe('inset(100%)')
 expect(exposureClip(rect,[rect])).toBe('')
})
test('gesture exposure retains all previously exposed pixels, expands ahead and has bounded complexity',()=>{
 const body={x:50,y:70,width:1600,height:800}
 let previous=retainExposure(body,undefined,[{x:50,y:70,width:10,height:800}])
 expect(previous).toEqual([{x:0,y:0,width:74,height:800}])
 const left=previous[0]!
 previous=retainExposure(body,previous,[{x:1640,y:70,width:10,height:800}])
 expect(previous).toContainEqual(left)
 expect(previous.some(r=>r.x<=800&&r.x+r.width>=800)).toBe(false)
 for(let i=0;i<200;i++)previous=retainExposure(body,previous,[{x:50+i*7,y:70+i%3*300,width:1,height:1}])
 expect(previous.length).toBeLessThanOrEqual(32)
 expect(previous.some(r=>r.x<=left.x&&r.y<=left.y&&r.x+r.width>=left.x+left.width&&r.y+r.height>=left.y+left.height)).toBe(true)
})

test('a moving reveal reuses its reserve and expands before the first uncovered pixel',()=>{
 const body={x:120,y:80,width:1000,height:600}
 const initial=retainExposureBounds(body,undefined,{x:120,y:80,width:10,height:600})
 for(let width=11;width<=74;width++)expect(retainExposureBounds(body,initial,{x:120,y:80,width,height:600})).toBe(initial)
 const expanded=retainExposureBounds(body,initial,{x:120,y:80,width:75,height:600})
 expect(expanded).not.toBe(initial);expect(expanded).toEqual({x:0,y:0,width:139,height:600})
 expect(retainExposureBounds(body,expanded,{x:120,y:80,width:2,height:600})).toBe(expanded)
 const resized={...body,width:60}
 expect(retainExposureBounds(resized,expanded,{...resized,width:1})).toEqual({x:0,y:0,width:60,height:600})
})

test('translation and reverse motion never omit a fractional visible point or repeatedly grow buffered bounds',()=>{
 let previous:ReturnType<typeof retainExposureBounds>|undefined,changes=0
 for(let step=0;step<600;step++){
  const body={x:70+step*.25,y:80-step*.5,width:900,height:640},edge=2+Math.abs(300-step)*.5
  const visible={x:body.x+edge,y:body.y+100.25,width:20.5,height:75.25}
  const next=retainExposureBounds(body,previous,visible)
  if(next!==previous)changes++
  expect(next.x).toBeLessThanOrEqual(visible.x-body.x);expect(next.y).toBeLessThanOrEqual(visible.y-body.y)
  expect(next.x+next.width).toBeGreaterThanOrEqual(visible.x+visible.width-body.x);expect(next.y+next.height).toBeGreaterThanOrEqual(visible.y+visible.height-body.y)
  if(previous){expect(next.x).toBeLessThanOrEqual(previous.x);expect(next.y).toBeLessThanOrEqual(previous.y);expect(next.x+next.width).toBeGreaterThanOrEqual(previous.x+previous.width);expect(next.y+next.height).toBeGreaterThanOrEqual(previous.y+previous.height)}
  previous=next
 }
 expect(changes).toBeLessThan(8)
})
