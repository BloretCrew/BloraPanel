import type {Page} from '@playwright/test'

// Optical-only experiment: retain the native shadow's inside-frame paint
// order, and move only its disjoint outside rectangles to sibling surfaces.
// Never copy application DOM, text, terminal cells or icon contents.
export async function installOutsideShadows(page:Page){
 // The Firefox native path is independently pixel-verified. Other engines
 // and fractional DPR keep their original DOM and paint ownership, including
 // no hidden sibling insertions that could trigger a native repaint.
 const eligible=await page.evaluate(()=>/Firefox\//.test(navigator.userAgent)&&Number.isInteger(devicePixelRatio))
 if(!eligible){await page.evaluate(()=>{(window as any).__nativeOutsideShadowDiagnostic={setVisible:()=>{},count:()=>0,dispose:()=>{}}});return}
 await page.addStyleTag({content:`
 .native-shadow-outside{position:absolute;left:0;top:0;border:1px solid transparent;pointer-events:none;contain:size layout style;will-change:transform}
 .native-shadow-outside.focused>.window-shadow-plane{box-shadow:var(--window-shadow-focus)}
 .native-shadow-outside[data-shadow-low="true"]>.window-shadow-plane{box-shadow:var(--window-shadow)}
 .native-shadow-outside.focused>.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:var(--shadow-focused-display,block);background-image:var(--shadow-focused);--shadow-piece-width:var(--shadow-focused-width);--shadow-piece-height:var(--shadow-focused-height);--shadow-trim-left:var(--shadow-focused-left);--shadow-trim-top:var(--shadow-focused-top);--shadow-trim-right:var(--shadow-focused-right);--shadow-trim-bottom:var(--shadow-focused-bottom)}
 .native-shadow-outside[data-shadow-low="true"]>.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:var(--shadow-normal-display,block);background-image:var(--shadow-normal);--shadow-piece-width:var(--shadow-normal-width);--shadow-piece-height:var(--shadow-normal-height);--shadow-trim-left:var(--shadow-normal-left);--shadow-trim-top:var(--shadow-normal-top);--shadow-trim-right:var(--shadow-normal-right);--shadow-trim-bottom:var(--shadow-normal-bottom)}
 `})
 await page.evaluate(()=>{
  const desktop=document.querySelector<HTMLElement>('.desktop')!
  const records=new Map<HTMLElement,{original:HTMLElement;shells:HTMLElement[];planes:HTMLElement[];observer:MutationObserver;clip:string}>()
  let enabled=true
  const set=(element:HTMLElement,name:string,value:string)=>{if(element.style.getPropertyValue(name)!==value)element.style.setProperty(name,value)}
  function register(frame:HTMLElement){
   if(records.has(frame))return
   const original=frame.querySelector<HTMLElement>(':scope>.window-shadow-plane');if(!original)return
   if(original.children.length!==8||[...original.children].some(piece=>!piece.classList.contains('shadow-piece')||piece.children.length||piece.textContent))throw Error('Only empty native optical shadow pieces are eligible')
   const shells:HTMLElement[]=[],planes:HTMLElement[]=[],clip=original.style.clip
   for(let edge=0;edge<4;edge++){
    const shell=document.createElement('div');shell.className='native-shadow-outside';shell.setAttribute('aria-hidden','true')
    const plane=original.cloneNode(true) as HTMLElement;shell.append(plane);shells.push(shell);planes.push(plane);frame.before(shell)
   }
   let layoutKey='',width=0,height=0,radius=''
   const sync=()=>{
    // Native fractional ownership remains the unchanged product fallback.
    const active=enabled&&Number.isInteger(devicePixelRatio)&&!!original.dataset.shadowCache
    const nextKey=frame.style.width+'|'+frame.style.height+'|'+frame.style.borderRadius+'|'+[...frame.classList].filter(name=>name!=='moving'&&name!=='focused').join(' ')
    if(nextKey!==layoutKey){layoutKey=nextKey;width=frame.offsetWidth;height=frame.offsetHeight;radius=getComputedStyle(frame).borderRadius}
    const pad=parseFloat(original.style.getPropertyValue('--shadow-outset'))||0
    const bounds=[[-pad,width+pad,0,-pad],[height,width+pad,height+pad,-pad],[0,0,height,-pad],[0,width+pad,height,width]]
    const inside=`rect(0px, ${width}px, ${height}px, 0px)`
    set(original,'clip',active?inside:clip)
    for(let edge=0;edge<4;edge++){
     const shell=shells[edge]!,plane=planes[edge]!
     set(shell,'transform',frame.style.transform);set(shell,'width',frame.style.width);set(shell,'height',frame.style.height);set(shell,'z-index',frame.style.zIndex);set(shell,'display',active?frame.style.display:'none')
     set(shell,'border-radius',radius)
     shell.classList.toggle('focused',frame.classList.contains('focused'))
     if(frame.dataset.shadowLow)shell.dataset.shadowLow=frame.dataset.shadowLow;else delete shell.dataset.shadowLow
     if(plane.dataset.shadowCache!==original.dataset.shadowCache){if(original.dataset.shadowCache)plane.dataset.shadowCache=original.dataset.shadowCache;else delete plane.dataset.shadowCache}
     for(const name of ['--shadow-width','--shadow-outset','--shadow-corner'])set(plane,name,original.style.getPropertyValue(name))
     set(plane,'clip',`rect(${bounds[edge]!.map(value=>value+'px').join(', ')})`)
     for(let index=0;index<8;index++){
      const source=original.children[index] as HTMLElement,target=plane.children[index] as HTMLElement
      if(target.style.cssText!==source.style.cssText)target.style.cssText=source.style.cssText
      if(source.dataset.shadowEmpty)target.dataset.shadowEmpty=source.dataset.shadowEmpty;else delete target.dataset.shadowEmpty
     }
    }
   }
   const observer=new MutationObserver(sync);observer.observe(frame,{attributes:true,attributeFilter:['style','class','data-shadow-low']});observer.observe(original,{subtree:true,attributes:true,attributeFilter:['style','data-shadow-cache','data-shadow-empty']})
   records.set(frame,{original,shells,planes,observer,clip});sync()
  }
  const scan=()=>{
   for(const [frame,record] of records)if(!frame.isConnected){record.observer.disconnect();for(const shell of record.shells)shell.remove();records.delete(frame)}
   for(const frame of desktop.querySelectorAll<HTMLElement>(':scope>.app-window'))register(frame)
  }
  const structure=new MutationObserver(scan);structure.observe(desktop,{childList:true});scan()
  ;(window as any).__nativeOutsideShadowDiagnostic={setVisible:(value:boolean)=>{enabled=value;for(const [frame,record] of records){if(value)frame.style.setProperty('--outside-shadow-toggle','1');else{record.original.style.clip=record.clip;for(const shell of record.shells)shell.style.display='none';frame.style.removeProperty('--outside-shadow-toggle')}}},count:()=>records.size,dispose:()=>{structure.disconnect();for(const record of records.values()){record.observer.disconnect();record.original.style.clip=record.clip;for(const shell of record.shells)shell.remove()}records.clear()}}
 })
}
