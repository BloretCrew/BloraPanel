import type {Page} from '@playwright/test'

// Test-only optical primitive: reproduce constant-axis native shadow pixels
// with the browser's gradient brush. App content is never read or retained.
export async function installLinearShadows(page:Page){
 await page.evaluate(async()=>{
  const records:{element:HTMLElement;normal:string;focused:string;linearNormal:string;linearFocused:string}[]=[]
  if(devicePixelRatio===1)for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane[data-shadow-cache]'))for(const name of ['n','s','w','e']){
   const element=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!,horizontal=name==='n'||name==='s'
   const images:string[]=[]
   for(const state of ['normal','focused']){
    const value=element.style.getPropertyValue('--shadow-'+state),match=/^url\("([^"]+)"\)$/.exec(value)
    if(!match){images.length=0;break}
    const image=new Image();image.src=match[1]!;await image.decode()
    const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height
    const context=canvas.getContext('2d')!;context.drawImage(image,0,0)
    const pixels=context.getImageData(0,0,image.width,image.height).data
    const count=horizontal?image.height:image.width,span=horizontal?image.width:image.height
    let constant=true
    for(let i=0;i<count&&constant;i++)for(let j=1;j<span&&constant;j++)for(let channel=0;channel<4;channel++){
     const first=(horizontal?i*image.width:i)*4+channel,at=(horizontal?i*image.width+j:j*image.width+i)*4+channel
     if(pixels[first]!==pixels[at]){constant=false;break}
    }
    if(!constant){images.length=0;break}
    const stops:string[]=[]
    for(let i=0;i<count;i++){
     const at=(horizontal?i*image.width:i)*4
     stops.push(`rgba(${pixels[at]},${pixels[at+1]},${pixels[at+2]},${(pixels[at+3]!/255).toFixed(10)}) ${i+.5}px`)
    }
    images.push(`linear-gradient(to ${horizontal?'bottom':'right'},${stops.join(',')})`)
   }
   if(images.length===2)records.push({element,normal:element.style.getPropertyValue('--shadow-normal'),focused:element.style.getPropertyValue('--shadow-focused'),linearNormal:images[0]!,linearFocused:images[1]!})
  }
  const toggle=(enabled:boolean)=>{for(const record of records){record.element.style.setProperty('--shadow-normal',enabled?record.linearNormal:record.normal);record.element.style.setProperty('--shadow-focused',enabled?record.linearFocused:record.focused)}}
  toggle(true)
  ;(window as any).__linearShadows={snapshot:()=>({edges:records.length}),setVisible:toggle,dispose(){toggle(false)}}
 })
}
