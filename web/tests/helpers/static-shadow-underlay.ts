import type {Page} from '@playwright/test'

// Optical-only prototype. Bake a native empty shadow over the original
// opaque wallpaper only where no other window/shadow or desktop icon can
// contribute. Focused/moving/changed/unsupported geometry keeps native PNGs.
// No application node, pixel, text, font, glyph or icon enters a canvas.
export async function installStaticShadowUnderlay(page:Page){
 await page.waitForFunction(()=>document.documentElement.dataset.materialCache==='ready'&&[...document.querySelectorAll<HTMLElement>('.app-window>.window-shadow-plane')].every(p=>!!p.dataset.shadowCache))
 await page.evaluate(async()=>{
  const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!
  ;(window as any).__staticShadowUnderlay?.dispose()
  if(devicePixelRatio!==1||root.dataset.materialCacheOpaque!=='true'){(window as any).__staticShadowUnderlay={snapshot:()=>({eligible:0,opaquePixels:0,dpr:devicePixelRatio,opaque:root.dataset.materialCacheOpaque==='true',frames:0,tiles:0,fragments:0,rejectedTiles:0}),setVisible:()=>{},dispose:()=>{}};return}
  type Rect={x:number;y:number;width:number;height:number}
  type Fragment={element:HTMLElement;rect:Rect;opaque:boolean}
  type Tile={piece:HTMLElement;rect:Rect;source:string;image?:string;original:string;priority:string;fragments?:Fragment[]}
  type Frame={frame:HTMLElement;plane:HTMLElement;rect:Rect;transform:string;width:string;height:string;pad:number;tiles:Tile[];key:string;z:number}
  const box=(b:DOMRect):Rect=>({x:b.x,y:b.y,width:b.width,height:b.height})
  const url=(s:string)=>/^url\("(blob:[^"]+)"\)$/.exec(s.trim())?.[1]
  const wallpaperSource=root.style.getPropertyValue('--material-desktop'),wallpaperURL=url(wallpaperSource),wallpaper=new Image(),urls:string[]=[],frames:Frame[]=[]
  let enabled=true,disposed=false,rejectedTiles=0
  if(!wallpaperURL)throw Error('Opaque native wallpaper is required')
  wallpaper.src=wallpaperURL;await wallpaper.decode()
  const width=desktop.clientWidth,height=desktop.clientHeight
  const paper=document.createElement('canvas');paper.width=width;paper.height=height
  const painted=paper.getContext('2d',{alpha:false})!;painted.drawImage(wallpaper,0,0,width,height)
  const bmp=(canvas:HTMLCanvasElement)=>{
   const w=canvas.width,h=canvas.height,pixels=canvas.getContext('2d')!.getImageData(0,0,w,h).data,stride=Math.ceil(w*3/4)*4,bytes=new Uint8Array(54+stride*h),header=new DataView(bytes.buffer)
   bytes[0]=66;bytes[1]=77;header.setUint32(2,bytes.length,true);header.setUint32(10,54,true);header.setUint32(14,40,true);header.setInt32(18,w,true);header.setInt32(22,h,true);header.setUint16(26,1,true);header.setUint16(28,24,true);header.setUint32(34,stride*h,true)
   for(let y=0;y<h;y++)for(let x=0;x<w;x++){const a=(y*w+x)*4,b=54+(h-1-y)*stride+x*3;if(pixels[a+3]!==255)throw Error('Optical underlay must remain opaque');bytes[b]=pixels[a+2]!;bytes[b+1]=pixels[a+1]!;bytes[b+2]=pixels[a]!}
   return new Blob([bytes],{type:'image/bmp'})
  }
  const names=['nw','n','ne','w','e','sw','s','se']
  for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!plane?.dataset.shadowCache)continue
   const rect=box(frame.getBoundingClientRect()),p=box(plane.getBoundingClientRect()),pad=parseFloat(plane.style.getPropertyValue('--shadow-outset')),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner')),tiles:Tile[]=[]
   for(const name of names){
    const piece=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!;if(piece.children.length||piece.textContent)throw Error('Only empty native shadow pieces are eligible')
    const source=url(piece.style.getPropertyValue('--shadow-normal'));if(!source||piece.style.getPropertyValue('--shadow-normal-display')==='none')continue
    const value=(key:string)=>parseFloat(piece.style.getPropertyValue('--shadow-normal-'+key)),w=name==='n'||name==='s'?p.width-2*corner:value('width'),h=name==='w'||name==='e'?p.height-2*corner:value('height')
    const x=['nw','w','sw'].includes(name)?-pad+value('left'):['ne','e','se'].includes(name)?p.width+pad-value('right')-w:corner,y=['nw','n','ne'].includes(name)?-pad+value('top'):['sw','s','se'].includes(name)?p.height+pad-value('bottom')-h:corner
    const area={x:p.x+x,y:p.y+y,width:w,height:h}
    if(!Object.values(area).every(Number.isInteger)||w<=0||h<=0||w*h>1000000||area.x<0||area.y<0||area.x+w>width||area.y+h>height){rejectedTiles++;continue}
    tiles.push({piece,rect:area,source,original:piece.style.getPropertyValue('background-image'),priority:piece.style.getPropertyPriority('background-image')})
   }
   frames.push({frame,plane,rect,transform:frame.style.transform,width:frame.style.width,height:frame.style.height,pad,tiles,key:plane.dataset.shadowCache,z:Number(frame.style.zIndex)})
  }
  const intersects=(a:Rect,b:Rect)=>a.x<b.x+b.width&&a.x+a.width>b.x&&a.y<b.y+b.height&&a.y+a.height>b.y
  let icons:Rect[]=[]
  const readIcons=()=>{icons=[...desktop.querySelectorAll<HTMLElement>('.desktop-icon')].map(icon=>{const r=box(icon.getBoundingClientRect());return {x:r.x-12,y:r.y-12,width:r.width+24,height:r.height+24}})}
  readIcons()
  const expand=(r:Rect,pad:number):Rect=>({x:r.x-pad,y:r.y-pad,width:r.width+2*pad,height:r.height+2*pad})
  const partition=(tile:Rect,covers:Rect[])=>{
   let areas:{rect:Rect;opaque:boolean}[]=[{rect:tile,opaque:true}]
   for(const cover of covers){
    const next:typeof areas=[]
    for(const area of areas){
     const r=area.rect,x=Math.max(r.x,cover.x),y=Math.max(r.y,cover.y),right=Math.min(r.x+r.width,cover.x+cover.width),bottom=Math.min(r.y+r.height,cover.y+cover.height)
     if(!area.opaque||x>=right||y>=bottom){next.push(area);continue}
     if(y>r.y)next.push({rect:{x:r.x,y:r.y,width:r.width,height:y-r.y},opaque:true})
     if(bottom<r.y+r.height)next.push({rect:{x:r.x,y:bottom,width:r.width,height:r.y+r.height-bottom},opaque:true})
     if(x>r.x)next.push({rect:{x:r.x,y,width:x-r.x,height:bottom-y},opaque:true})
     if(right<r.x+r.width)next.push({rect:{x:right,y,width:r.x+r.width-right,height:bottom-y},opaque:true})
     next.push({rect:{x,y,width:right-x,height:bottom-y},opaque:false})
    }
    if(next.length>128)return [{rect:tile,opaque:false}]
    areas=next
   }
   return areas
  }
  const sources=new Map<string,Promise<HTMLImageElement>>()
  const decode=(source:string)=>{if(!sources.has(source))sources.set(source,(async()=>{const native=new Image();native.src=source;await native.decode();return native})());return sources.get(source)!}
  for(const record of frames)for(const tile of record.tiles){
   const t=tile.rect,areas=partition(t,[...frames.map(other=>expand(other.rect,2)),...icons])
   if(!areas.some(area=>area.opaque))continue
   const canvas=document.createElement('canvas');canvas.width=tile.rect.width;canvas.height=tile.rect.height;const context=canvas.getContext('2d',{alpha:false})!
   context.drawImage(paper,tile.rect.x,tile.rect.y,canvas.width,canvas.height,0,0,canvas.width,canvas.height)
   // Reproduce lower empty shadow planes in their native stacking order.
   // Windows, borders, icons and all application pixels are absent from this
   // optical-only canvas, and remain excluded from every opaque fragment.
   for(const lower of frames.filter(other=>other.z<=record.z).sort((a,b)=>a.z-b.z))for(const shadow of lower.tiles){
    if(!intersects(shadow.rect,tile.rect))continue
    context.drawImage(await decode(shadow.source),shadow.rect.x-tile.rect.x,shadow.rect.y-tile.rect.y,shadow.rect.width,shadow.rect.height)
   }
   const resource=URL.createObjectURL(bmp(canvas));urls.push(resource);const decoded=new Image();decoded.src=resource;await decoded.decode();tile.image=`url("${resource}")`
   // The native frame's antialiased border blends with its shadow. Preserve
   // that exact ordering, including two pixels around the rectangular frame.
   // Only empty optical regions outside it may become opaque wallpaper.
   tile.fragments=[]
   for(const area of areas){
    if(area.rect.width<=0||area.rect.height<=0)continue
    const fragment=document.createElement('i'),a={...area.rect,x:area.rect.x-t.x,y:area.rect.y-t.y}
    fragment.dataset.staticShadowFragment='true';Object.assign(fragment.style,{position:'absolute',pointerEvents:'none',display:'none',left:a.x+'px',top:a.y+'px',width:a.width+'px',height:a.height+'px',backgroundImage:area.opaque?tile.image:`url("${tile.source}")`,backgroundSize:t.width+'px '+t.height+'px',backgroundPosition:-a.x+'px '+-a.y+'px',backgroundRepeat:'no-repeat'})
    tile.fragments.push({element:fragment,rect:area.rect,opaque:area.opaque});tile.piece.append(fragment)
   }
  }
  const translation=(s:string)=>{const m=/^translate\((-?[\d.]+)px,\s*(-?[\d.]+)px\)$/.exec(s);return m?{x:Number(m[1]),y:Number(m[2])}:undefined}
  const live=(r:Frame):Rect|undefined=>{
   if(!r.frame.isConnected||r.frame.style.display==='none')return undefined
   const a=translation(r.transform),b=translation(r.frame.style.transform)
   if(!a||!b||r.width!==r.frame.style.width||r.height!==r.frame.style.height)return box(r.frame.getBoundingClientRect())
   return {...r.rect,x:r.rect.x+b.x-a.x,y:r.rect.y+b.y-a.y}
  }
  let eligible=0,opaquePixels=0
  const sync=()=>{
   eligible=opaquePixels=0
   const appearances=root.dataset.materialCache==='ready'&&root.style.getPropertyValue('--material-desktop')===wallpaperSource,positions=new Map(frames.map(r=>[r,live(r)]))
   const unknown=[...desktop.querySelectorAll<HTMLElement>(':scope>.app-window')].some(frame=>!frames.some(r=>r.frame===frame))
   for(const record of frames){
    const matches=(r:Frame)=>{const p=positions.get(r);return !!p&&p.x===r.rect.x&&p.y===r.rect.y&&p.width===r.rect.width&&p.height===r.rect.height&&r.plane.dataset.shadowCache===r.key&&Number(r.frame.style.zIndex)===r.z&&!r.frame.classList.contains('focused')&&!r.frame.classList.contains('moving')}
    const stable=matches(record)&&frames.filter(other=>other.z<record.z).every(matches)
    for(const tile of record.tiles){
     const image=enabled&&!disposed&&appearances&&!unknown&&stable?tile.image:undefined
     if(image){
      if(tile.piece.style.getPropertyValue('background-image')!=='none')tile.piece.style.setProperty('background-image','none','important');tile.piece.dataset.staticShadowUnderlay='true'
      for(const fragment of tile.fragments||[]){
       const clear=fragment.opaque&&!icons.some(icon=>intersects(fragment.rect,icon))&&!frames.some(other=>{if(other===record)return false;const r=positions.get(other);return !!r&&intersects(fragment.rect,expand(r,2))})
       const value=clear?image:`url("${tile.source}")`
       if(fragment.element.style.backgroundImage!==value)fragment.element.style.backgroundImage=value
       if(fragment.element.style.display!=='block')fragment.element.style.display='block'
       if(clear){eligible++;opaquePixels+=fragment.rect.width*fragment.rect.height}
      }
     }
     else{if(tile.piece.style.getPropertyValue('background-image')!==tile.original||tile.piece.style.getPropertyPriority('background-image')!==tile.priority){if(tile.original)tile.piece.style.setProperty('background-image',tile.original,tile.priority);else tile.piece.style.removeProperty('background-image')}delete tile.piece.dataset.staticShadowUnderlay;for(const fragment of tile.fragments||[])if(fragment.element.style.display!=='none')fragment.element.style.display='none'}
    }
   }
  }
  const changes=new MutationObserver(()=>sync())
  for(const record of frames){changes.observe(record.frame,{attributes:true,attributeFilter:['style','class']});changes.observe(record.plane,{attributes:true,attributeFilter:['data-shadow-cache']})}
  const appearance=new MutationObserver(()=>sync());appearance.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-palette','data-theme','data-translucency']})
  const structure=new MutationObserver(()=>{readIcons();sync()});structure.observe(desktop,{childList:true})
  for(const icon of desktop.querySelectorAll<HTMLElement>('.desktop-icon'))structure.observe(icon,{attributes:true,attributeFilter:['style','class']})
  const resize=()=>{enabled=false;sync()};window.addEventListener('resize',resize)
  const controller={snapshot:()=>({eligible,opaquePixels,dpr:devicePixelRatio,opaque:root.dataset.materialCacheOpaque==='true',frames:frames.length,tiles:frames.reduce((n,r)=>n+r.tiles.length,0),fragments:frames.reduce((n,r)=>n+r.tiles.reduce((m,t)=>m+(t.fragments?.length||0),0),0),rejectedTiles}),setVisible(value:boolean){enabled=value;sync()},dispose(){if(disposed)return;enabled=false;sync();disposed=true;changes.disconnect();appearance.disconnect();structure.disconnect();window.removeEventListener('resize',resize);for(const record of frames)for(const tile of record.tiles)for(const fragment of tile.fragments||[])fragment.element.remove();for(const resource of urls)URL.revokeObjectURL(resource)}}
  ;(window as any).__staticShadowUnderlay=controller;window.addEventListener('pagehide',()=>controller.dispose(),{once:true});sync()
 })
}
