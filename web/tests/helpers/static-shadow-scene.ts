import type {Page} from '@playwright/test'

// Test-only scene precomposition: original wallpaper + isolated empty native
// shadows, never application/text/glyph/icon pixels. The active native window
// remains independent; static shadows over applications keep native painting.
export async function installStaticShadowScene(page:Page,hull=false){
 await page.waitForFunction(()=>document.documentElement.dataset.materialCache==='ready'&&[...document.querySelectorAll<HTMLElement>('.app-window>.window-shadow-plane')].every(p=>!!p.dataset.shadowCache))
 await page.evaluate(async(hull:boolean)=>{
  const root=document.documentElement,desktop=document.querySelector<HTMLElement>('.desktop')!
  ;(window as any).__staticShadowUnderlay?.dispose()
  const empty=()=>{(window as any).__staticShadowUnderlay={snapshot:()=>({eligible:0,opaquePixels:0}),setVisible:()=>{},dispose:()=>{}}}
  if(devicePixelRatio!==1||root.dataset.materialCacheOpaque!=='true'){empty();return}
  type Rect={x:number;y:number;width:number;height:number}
  type Tile={element:HTMLElement;rect:Rect;source:string;original:string;priority:string;child?:HTMLElement;pixels:number;clip?:string}
  type Frame={element:HTMLElement;plane:HTMLElement;rect:Rect;transform:string;width:string;height:string;z:string;key:string;tiles:Tile[]}
  const box=(b:DOMRect):Rect=>({x:b.x,y:b.y,width:b.width,height:b.height})
  const expand=(r:Rect,p:number):Rect=>{const x=Math.floor(r.x-p),y=Math.floor(r.y-p);return {x,y,width:Math.ceil(r.x+r.width+p)-x,height:Math.ceil(r.y+r.height+p)-y}}
  const intersect=(a:Rect,b:Rect):Rect|undefined=>{const x=Math.max(a.x,b.x),y=Math.max(a.y,b.y),right=Math.min(a.x+a.width,b.x+b.width),bottom=Math.min(a.y+a.height,b.y+b.height);return x<right&&y<bottom?{x,y,width:right-x,height:bottom-y}:undefined}
  const url=(s:string)=>/^url\("(blob:[^"]+)"\)$/.exec(s.trim())?.[1]
  const wallpaperSource=root.style.getPropertyValue('--material-desktop'),wallpaperURL=url(wallpaperSource),width=desktop.clientWidth,height=desktop.clientHeight
  if(!wallpaperURL){empty();return}
  const paper=document.createElement('canvas');paper.width=width;paper.height=height
  const wallpaper=new Image();wallpaper.src=wallpaperURL;await wallpaper.decode();paper.getContext('2d',{alpha:false})!.drawImage(wallpaper,0,0,width,height)
  const frames:Frame[]=[]
  for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window')){
   const plane=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane')!,p=box(plane.getBoundingClientRect()),pad=parseFloat(plane.style.getPropertyValue('--shadow-outset')),corner=parseFloat(plane.style.getPropertyValue('--shadow-corner')),tiles:Tile[]=[]
   for(const name of ['nw','n','ne','w','e','sw','s','se']){
    const element=plane.querySelector<HTMLElement>('.shadow-piece.'+name)!,source=url(element.style.getPropertyValue('--shadow-normal'))
    if(element.children.length||element.textContent)throw Error('Only isolated empty shadow sources are allowed')
    if(!source||element.style.getPropertyValue('--shadow-normal-display')==='none')continue
    const value=(key:string)=>parseFloat(element.style.getPropertyValue('--shadow-normal-'+key)),w=name==='n'||name==='s'?p.width-2*corner:value('width'),h=name==='w'||name==='e'?p.height-2*corner:value('height')
    const x=['nw','w','sw'].includes(name)?-pad+value('left'):['ne','e','se'].includes(name)?p.width+pad-value('right')-w:corner,y=['nw','n','ne'].includes(name)?-pad+value('top'):['sw','s','se'].includes(name)?p.height+pad-value('bottom')-h:corner
    const rect={x:p.x+x,y:p.y+y,width:w,height:h}
    if(!Object.values(rect).every(Number.isInteger)||w<=0||h<=0||w*h>1000000){empty();return}
    tiles.push({element,rect,source,original:element.style.getPropertyValue('background-image'),priority:element.style.getPropertyPriority('background-image'),pixels:0})
   }
   frames.push({element:frame,plane,rect:box(frame.getBoundingClientRect()),transform:frame.style.transform,width:frame.style.width,height:frame.style.height,z:frame.style.zIndex,key:plane.dataset.shadowCache!,tiles})
  }
  const active=frames.find(r=>r.element.classList.contains('focused'))
  if(!active||frames.some(r=>Number(r.z)>Number(active.z))){empty();return}
  const still=frames.filter(r=>r!==active),iconElements=[...desktop.querySelectorAll<HTMLElement>('.desktop-icon')],icons=iconElements.map(e=>box(e.getBoundingClientRect()))
  let blocks=[...still.map(r=>expand(r.rect,2)),...icons.map(r=>expand(r,12))]
  if(hull){
   if(!still.length||still.some(r=>r.tiles.some(t=>icons.some(icon=>!!intersect(t.rect,expand(icon,12)))))){empty();return}
   const x=Math.floor(Math.min(...still.map(r=>r.rect.x))-2),y=Math.floor(Math.min(...still.map(r=>r.rect.y))-2),right=Math.ceil(Math.max(...still.map(r=>r.rect.x+r.rect.width))+2),bottom=Math.ceil(Math.max(...still.map(r=>r.rect.y+r.rect.height))+2)
   blocks=[{x,y,width:right-x,height:bottom-y}]
  }
  const scene=document.createElement('canvas');scene.width=width;scene.height=height;const context=scene.getContext('2d',{alpha:false})!;context.drawImage(paper,0,0)
  const sources=new Map<string,Promise<HTMLImageElement>>()
  const decode=(source:string)=>{if(!sources.has(source))sources.set(source,(async()=>{const image=new Image();image.src=source;await image.decode();return image})());return sources.get(source)!}
  for(const frame of [...still].sort((a,b)=>Number(a.z)-Number(b.z)))for(const tile of frame.tiles){const r=tile.rect;context.drawImage(await decode(tile.source),r.x,r.y,r.width,r.height)}
  for(const block of blocks){const r=intersect(block,{x:0,y:0,width,height});if(r)context.drawImage(paper,r.x,r.y,r.width,r.height,r.x,r.y,r.width,r.height)}
  const stride=Math.ceil(width*3/4)*4,bytes=new Uint8Array(54+stride*height),header=new DataView(bytes.buffer),pixels=context.getImageData(0,0,width,height).data
  bytes[0]=66;bytes[1]=77;header.setUint32(2,bytes.length,true);header.setUint32(10,54,true);header.setUint32(14,40,true);header.setInt32(18,width,true);header.setInt32(22,height,true);header.setUint16(26,1,true);header.setUint16(28,24,true);header.setUint32(34,stride*height,true)
  for(let y=0;y<height;y++)for(let x=0;x<width;x++){const a=(y*width+x)*4,b=54+(height-1-y)*stride+x*3;if(pixels[a+3]!==255)throw Error('Scene must remain opaque');bytes[b]=pixels[a+2]!;bytes[b+1]=pixels[a+1]!;bytes[b+2]=pixels[a]!}
  const resource=URL.createObjectURL(new Blob([bytes],{type:'image/bmp'})),decoded=new Image();decoded.src=resource;await decoded.decode();const sceneImage=`url("${resource}")`
  // Native pieces retain their manager-owned outer clip. One inner native
  // image keeps only the portions where the opaque scene was left untouched.
  const opticalStyle=document.createElement('style');opticalStyle.textContent='[data-static-shadow-hull]{clip:var(--static-shadow-hull-clip)!important}';if(hull)document.head.append(opticalStyle)
  for(const frame of still)for(const tile of frame.tiles){
   const regions=blocks.flatMap(block=>{const r=intersect(tile.rect,block);return r?[r]:[]}),r=tile.rect
   if(regions.length>128){URL.revokeObjectURL(resource);empty();return}
   if(hull){const v=regions[0];tile.clip=v?`rect(${v.y-r.y}px,${v.x+v.width-r.x}px,${v.y+v.height-r.y}px,${v.x-r.x}px)`:'rect(0px,0px,0px,0px)'}
   else{const child=document.createElement('i');Object.assign(child.style,{position:'absolute',pointerEvents:'none',display:'none',inset:'0',backgroundImage:`url("${tile.source}")`,backgroundSize:'100% 100%',backgroundRepeat:'no-repeat',clipPath:regions.length?`path('${regions.map(v=>`M${v.x-r.x} ${v.y-r.y}h${v.width}v${v.height}h${-v.width}Z`).join(' ')}')`:'inset(100%)'});tile.child=child;tile.element.append(child)}
   tile.pixels=r.width*r.height
  }
  const background=desktop.style.getPropertyValue('background-image'),backgroundPriority=desktop.style.getPropertyPriority('background-image')
  let enabled=true,disposed=false,visible=false,eligible=0,opaquePixels=0
  const same=(r:Frame)=>r.element.isConnected&&r.element.style.display!=='none'&&r.element.style.width===r.width&&r.element.style.height===r.height&&r.element.style.zIndex===r.z&&r.plane.dataset.shadowCache===r.key
  const sync=()=>{
   const valid=enabled&&!disposed&&root.dataset.materialCache==='ready'&&root.style.getPropertyValue('--material-desktop')===wallpaperSource&&same(active)&&active.element.classList.contains('focused')&&still.every(r=>same(r)&&r.element.style.transform===r.transform&&!r.element.classList.contains('focused')&&!r.element.classList.contains('moving'))&&[...desktop.querySelectorAll<HTMLElement>(':scope>.app-window')].every(e=>frames.some(r=>r.element===e))
   if(valid===visible)return
   visible=valid;eligible=opaquePixels=0
   if(valid)desktop.style.setProperty('background-image',sceneImage,'important')
   else if(background)desktop.style.setProperty('background-image',background,backgroundPriority);else desktop.style.removeProperty('background-image')
   for(const frame of still)for(const tile of frame.tiles){
    if(valid){if(hull){tile.element.style.setProperty('--static-shadow-hull-clip',tile.clip!);tile.element.dataset.staticShadowHull='true'}else{tile.element.style.setProperty('background-image','none','important');tile.child!.style.display='block'}tile.element.dataset.staticShadowUnderlay='true';eligible++;opaquePixels+=tile.pixels}
    else{if(hull){delete tile.element.dataset.staticShadowHull;tile.element.style.removeProperty('--static-shadow-hull-clip')}else{if(tile.original)tile.element.style.setProperty('background-image',tile.original,tile.priority);else tile.element.style.removeProperty('background-image');tile.child!.style.display='none'}delete tile.element.dataset.staticShadowUnderlay}
   }
  }
  const changes=new MutationObserver(sync);for(const r of frames){changes.observe(r.element,{attributes:true,attributeFilter:['style','class']});changes.observe(r.plane,{attributes:true,attributeFilter:['data-shadow-cache']})}
  const appearance=new MutationObserver(sync);appearance.observe(root,{attributes:true,attributeFilter:['data-material-cache','data-theme','data-palette','data-translucency']})
  const structure=new MutationObserver(records=>{if(records.some(r=>r.target!==desktop))enabled=false;sync()});structure.observe(desktop,{childList:true});for(const icon of iconElements)structure.observe(icon,{attributes:true,attributeFilter:['style','class']})
  const resize=()=>{enabled=false;sync()};window.addEventListener('resize',resize)
  const controller={snapshot:()=>({eligible,opaquePixels,frames:frames.length,nativeFragments:hull?0:still.reduce((n,r)=>n+r.tiles.length,0),sceneWidth:width,sceneHeight:height}),setVisible(value:boolean){enabled=value;sync()},dispose(){if(disposed)return;enabled=false;sync();disposed=true;changes.disconnect();appearance.disconnect();structure.disconnect();window.removeEventListener('resize',resize);for(const r of still)for(const t of r.tiles)t.child?.remove();opticalStyle.remove();URL.revokeObjectURL(resource)}}
  ;(window as any).__staticShadowUnderlay=controller;window.addEventListener('pagehide',()=>controller.dispose(),{once:true});sync()
 },hull)
}
