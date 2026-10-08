import type {Page} from '@playwright/test'

// Diagnostic only: split content-free native shadow paint from the moving
// application compositing surface. No app subtree/text/glyph enters a clone.
export async function installShadowSiblings(page:Page){
 await page.addStyleTag({content:`
 .native-shadow-sibling{position:absolute;left:0;top:0;pointer-events:none;border:1px solid var(--outline-variant);contain:size layout style;will-change:transform}
 :root[data-theme="dark"] .native-shadow-sibling{border-color:var(--dark-window-edge)}
 :root[data-theme="dark"] .native-shadow-sibling.focused{border-color:var(--dark-window-edge-focus)}
 .native-shadow-sibling.focused>.window-shadow-plane{box-shadow:var(--window-shadow-focus)}
 .native-shadow-sibling[data-shadow-low="true"]>.window-shadow-plane{box-shadow:var(--window-shadow)}
 .native-shadow-sibling.focused>.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:var(--shadow-focused-display,block);background-image:var(--shadow-focused);--shadow-piece-width:var(--shadow-focused-width);--shadow-piece-height:var(--shadow-focused-height);--shadow-trim-left:var(--shadow-focused-left);--shadow-trim-top:var(--shadow-focused-top);--shadow-trim-right:var(--shadow-focused-right);--shadow-trim-bottom:var(--shadow-focused-bottom)}
 .native-shadow-sibling[data-shadow-low="true"]>.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:var(--shadow-normal-display,block);background-image:var(--shadow-normal);--shadow-piece-width:var(--shadow-normal-width);--shadow-piece-height:var(--shadow-normal-height);--shadow-trim-left:var(--shadow-normal-left);--shadow-trim-top:var(--shadow-normal-top);--shadow-trim-right:var(--shadow-normal-right);--shadow-trim-bottom:var(--shadow-normal-bottom)}
 `})
 await page.evaluate(()=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!
  const records=new Map<HTMLElement,{original:HTMLElement;shell:HTMLElement;plane:HTMLElement;observer:MutationObserver;visibility:string;borderColor:string}>()
  function register(frame:HTMLElement){
   if(records.has(frame))return
   const original=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!original)return
   if(original.children.length!==8||[...original.children].some(piece=>!piece.classList.contains('shadow-piece')||piece.children.length||piece.textContent))throw Error('Only empty native optical shadow pieces are eligible')
   const shell=document.createElement('div');shell.className='native-shadow-sibling';shell.setAttribute('aria-hidden','true')
   const plane=original.cloneNode(true) as HTMLElement;shell.append(plane)
   const visibility=original.style.visibility
   const borderColor=frame.style.borderColor
   let lastClass=''
   const sync=()=>{
    shell.style.transform=frame.style.transform;shell.style.width=frame.style.width;shell.style.height=frame.style.height;shell.style.zIndex=frame.style.zIndex;shell.style.display=frame.style.display
    if(lastClass!==frame.className){lastClass=frame.className;shell.style.borderRadius=getComputedStyle(frame).borderRadius}
    if(frame.style.borderRadius)shell.style.borderRadius=frame.style.borderRadius
    shell.classList.toggle('focused',frame.classList.contains('focused'))
    if(frame.dataset.shadowLow)shell.dataset.shadowLow=frame.dataset.shadowLow;else delete shell.dataset.shadowLow
    plane.style.cssText=original.style.cssText;plane.style.visibility=visibility
    if(original.dataset.shadowCache)plane.dataset.shadowCache=original.dataset.shadowCache;else delete plane.dataset.shadowCache
    for(let index=0;index<8;index++){
     const source=original.children[index] as HTMLElement,target=plane.children[index] as HTMLElement
     target.style.cssText=source.style.cssText
     if(source.dataset.shadowEmpty)target.dataset.shadowEmpty=source.dataset.shadowEmpty;else delete target.dataset.shadowEmpty
    }
   }
   // Native negative-z descendants paint after their parent's border. Move
   // both optical pieces together; copying only the shadow reverses that
   // ordering and changes rounded-edge anti-aliasing.
   sync();frame.before(shell);original.style.visibility='hidden';frame.style.borderColor='transparent'
   const observer=new MutationObserver(sync);observer.observe(frame,{attributes:true,attributeFilter:['style','class','data-shadow-low']});observer.observe(original,{subtree:true,attributes:true,attributeFilter:['style','data-shadow-cache','data-shadow-empty']})
   records.set(frame,{original,shell,plane,observer,visibility,borderColor})
  }
  const scan=()=>{
   for(const [frame,record] of records)if(!frame.isConnected){record.observer.disconnect();record.shell.remove();records.delete(frame)}
   for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))register(frame)
  }
  const structure=new MutationObserver(scan);structure.observe(desktop,{childList:true});scan()
  ;(window as any).__nativeShadowSiblingDiagnostic={setVisible:(enabled:boolean)=>{for(const [frame,record] of records){record.shell.style.visibility=enabled?'':'hidden';record.original.style.visibility=enabled?'hidden':record.visibility;frame.style.borderColor=enabled?'transparent':record.borderColor}},count:()=>records.size,dispose:()=>{structure.disconnect();for(const [frame,record] of records){record.observer.disconnect();record.original.style.visibility=record.visibility;frame.style.borderColor=record.borderColor;record.shell.remove()}records.clear()}}
 })
}
