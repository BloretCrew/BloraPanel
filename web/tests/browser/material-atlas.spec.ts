import {test,expect} from '@playwright/test'

for(const scale of [1,1.25])test.describe(`wallpaper-only atlas at DPR ${scale}`,()=>{
 test.use({viewport:{width:1440,height:960},...nativeDeviceScale(scale)})
 test('product native RGB wallpaper matches independent PNG samples across light and dark palettes',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'small-atlas-test',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.goto('/');await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await page.evaluate(()=>{const surface=document.createElement('div');surface.dataset.smallAtlasSurface='true';surface.style.cssText='position:fixed;inset:0;z-index:10000;background-image:var(--material-window);background-size:var(--material-size);background-repeat:no-repeat;background-color:var(--surface-container);pointer-events:none';document.body.append(surface)})
  for(const theme of ['light','dark'])for(const palette of ['ice','mineral','sand','sage','iris','rose']){
   await page.evaluate(async({theme,palette})=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().commit([{kind:'set',path:['preferences','theme'],value:theme},{kind:'set',path:['preferences','palette'],value:palette}])},{theme,palette})
   await expect(page.locator('html')).toHaveAttribute('data-theme',theme);await expect(page.locator('html')).toHaveAttribute('data-palette',palette);await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready');await page.mouse.move(0,0)
   await expect(page.locator('html')).toHaveAttribute('data-material-cache-encoding','opaque-rgb24')
   // The product RGB surface is compared with a separately encoded native
   // PNG containing the identical original wallpaper samples, not itself.
   const metadata=await page.evaluate(async()=>{const {installSmallMaterialAtlas}=await import('/tests/real/material-atlas-diagnostic.ts' as string);return installSmallMaterialAtlas({opaque:false})})
   expect(metadata.width).toBe(metadata.sourceWidth);expect(metadata.encoding).toBe('png')
   const original=await page.locator('[data-small-atlas-surface]').screenshot()
   await page.evaluate(()=>(window as any).__smallMaterialAtlas.dispose())
   const candidate=await page.locator('[data-small-atlas-surface]').screenshot()
   const diff=await page.evaluate(async({a,b})=>{
    const load=async(data:string)=>{const image=new Image();image.src='data:image/png;base64,'+data;await image.decode();const c=document.createElement('canvas');c.width=image.width;c.height=image.height;const x=c.getContext('2d')!;x.drawImage(image,0,0);return x.getImageData(0,0,c.width,c.height).data}
    const x=await load(a),y=await load(b);let changed=0,max=0,total=0;for(let i=0;i<x.length;i++){const d=Math.abs(x[i]!-y[i]!);if(d)changed++;max=Math.max(max,d);total+=d}return {changed,max,mean:total/x.length}
   },{a:original.toString('base64'),b:candidate.toString('base64')})
   console.log('WALLPAPER_ATLAS_OPTICS',theme,palette,JSON.stringify({metadata,diff}));expect(diff).toEqual({changed:0,max:0,mean:0})
  }
  await page.locator('[data-small-atlas-surface]').evaluate(element=>(element as HTMLElement).remove())
 })
 test('native material encoding preserves transparent samples and odd-width row alignment',async({page})=>{
  await page.goto('/')
  const results=await page.evaluate(async()=>{
   const {encodeMaterial}=await import('/src/appearance/material-encoding.ts' as string)
   const decode=async(blob:Blob)=>{
    const url=URL.createObjectURL(blob)
    try{const image=new Image();image.src=url;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return {width:image.width,height:image.height,pixels:Array.from(ctx.getImageData(0,0,canvas.width,canvas.height).data)}}finally{URL.revokeObjectURL(url)}
   }
   const results=[]
   for(const width of [3,4,5])for(const transparent of [false,true]){
    const canvas=document.createElement('canvas');canvas.width=width;canvas.height=3
    const ctx=canvas.getContext('2d')!,data=ctx.createImageData(width,3)
    for(let i=0;i<data.data.length;i+=4){data.data[i]=(i*17)%256;data.data[i+1]=(i*31+7)%256;data.data[i+2]=(i*47+13)%256;data.data[i+3]=transparent&&i===4?127:255}
    ctx.putImageData(data,0,0)
    const reference=await encodeMaterial(canvas,false),candidate=await encodeMaterial(canvas,true)
    const a=await decode(reference),b=await decode(candidate)
    results.push({width,transparent,type:candidate.type,equal:JSON.stringify(a)===JSON.stringify(b),alpha:b.pixels[7]})
   }
   return results
  })
  expect(results).toHaveLength(6)
  for(const result of results){expect(result.equal).toBe(true);expect(result.type).toBe(result.transparent?'image/png':'image/bmp');expect(result.alpha).toBe(result.transparent?127:255)}
 })
})
import {nativeDeviceScale} from '../../playwright-browser'
