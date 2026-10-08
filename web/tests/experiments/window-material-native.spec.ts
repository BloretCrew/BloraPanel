import {test,expect} from '@playwright/test'
import {nativeDeviceScale} from '../../playwright-browser'
import {installWindowMaterialDiagnostic} from '../real/window-material-diagnostic'

for(const scale of [1,1.25])test.describe(`opaque native window atlas DPR ${scale}`,()=>{
 test.use({viewport:{width:1440,height:960},...nativeDeviceScale(scale)})
 test('preexpanded RGB24 preserves native optical samples and changes no application ownership',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'native-atlas-test',name:'测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.goto('/');await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
  await page.locator('[data-app="blora.instances"]').click()
  await page.evaluate(()=>{
   ;(window as any).__nativeAtlasApp=document.querySelector('.app-window')
   const surface=document.createElement('div');surface.dataset.nativeAtlasSurface='true';surface.style.cssText='position:fixed;inset:0;z-index:10000;background-image:var(--material-window);background-size:var(--material-size);background-repeat:no-repeat;background-color:var(--surface-container);pointer-events:none';document.body.append(surface)
  })
  const surface=page.locator('[data-native-atlas-surface]')
  for(const theme of ['light','dark'])for(const palette of ['ice','mineral','sand','sage','iris','rose']){
   await page.evaluate(async({theme,palette})=>{const {useDesktop}=await import('/src/desktop/store.ts' as string);useDesktop().commit([{kind:'set',path:['preferences','theme'],value:theme},{kind:'set',path:['preferences','palette'],value:palette}])},{theme,palette})
   await expect(page.locator('html')).toHaveAttribute('data-theme',theme);await expect(page.locator('html')).toHaveAttribute('data-palette',palette);await expect(page.locator('html')).toHaveAttribute('data-material-cache','ready')
   const original=await surface.screenshot()
   const meta=await installWindowMaterialDiagnostic(page,'opaque-rgb24')
   expect(meta).toMatchObject({width:Math.ceil(1440*scale),height:Math.ceil(960*scale),encoding:'image/bmp'})
   const candidate=await surface.screenshot()
   await page.evaluate(()=>(window as any).__windowMaterialDiagnostic.dispose())
   const png=await installWindowMaterialDiagnostic(page,'native-png');expect(png.encoding).toBe('image/png')
   const reference=await surface.screenshot()
   const difference=await page.evaluate(async({a,b,c})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const x=canvas.getContext('2d')!;x.drawImage(image,0,0);return x.getImageData(0,0,canvas.width,canvas.height).data}
    const old=await load(a),next=await load(b),reference=await load(c);let max=0,total=0,encodingMax=0,encodingChanged=0
    for(let i=0;i<old.length;i++){const delta=Math.abs(old[i]!-next[i]!);max=Math.max(max,delta);total+=delta;const encodingDelta=Math.abs(next[i]!-reference[i]!);encodingMax=Math.max(encodingMax,encodingDelta);if(encodingDelta)encodingChanged++}
    return {max,mean:total/old.length,encodingMax,encodingChanged}
   },{a:original.toString('base64'),b:candidate.toString('base64'),c:reference.toString('base64')})
   console.log('NATIVE_WINDOW_ATLAS',theme,palette,JSON.stringify({meta,difference}))
   // Same strict RGB encoding equality. Only wallpaper-only optical sampling
   // uses the declared 4/255 maximum and .1/255 mean boundary.
   expect(difference.encodingMax).toBe(0);expect(difference.encodingChanged).toBe(0)
   expect(difference.max).toBeLessThanOrEqual(4);expect(difference.mean).toBeLessThanOrEqual(.1)
   await page.evaluate(()=>(window as any).__windowMaterialDiagnostic.dispose())
  }
  expect(await page.evaluate(()=>(window as any).__nativeAtlasApp===document.querySelector('.app-window'))).toBe(true)
  await surface.evaluate(element=>(element as HTMLElement).remove())
 })
})
