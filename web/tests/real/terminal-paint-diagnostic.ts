// Numeric-only native paint attribution. Delegate both native entry points
// unchanged; never copy/serialize terminal text, nodes or application pixels.
export function installNativeTerminalPaintDiagnostic(){
 type Box={x:number;y:number;width:number;height:number}
 const box=(value:DOMRectReadOnly):Box=>({x:value.x,y:value.y,width:value.width,height:value.height})
 type Counts={host:HTMLElement;rows:number;intersecting?:boolean;intersection?:Box;callbacks:number}
 const values:Counts[]=[],tracked=new WeakMap<HTMLElement,Counts>()
 const lookup=(host:HTMLElement)=>{let value=tracked.get(host);if(!value){value={host,rows:0,callbacks:0};tracked.set(host,value);values.push(value)}return value}
 const NativeObserver=window.IntersectionObserver,replace=Element.prototype.replaceChildren
 window.IntersectionObserver=class extends NativeObserver{
  constructor(callback:IntersectionObserverCallback,options?:IntersectionObserverInit){
   super((entries,observer)=>{
    for(const entry of entries)if(entry.target.matches('.xterm-screen')){const host=entry.target.closest<HTMLElement>('.terminal-container');if(host){const value=lookup(host);value.intersecting=entry.isIntersecting;value.intersection=box(entry.intersectionRect);value.callbacks++}}
    callback(entries,observer)
   },options)
  }
 }
 Element.prototype.replaceChildren=function(this:Element,...nodes:(Node|string)[]){
  if(this.parentElement?.classList.contains('xterm-rows')){const host=this.closest<HTMLElement>('.terminal-container');if(host)lookup(host).rows++}
  return replace.apply(this,nodes)
 }
 const controller={reset(){for(const value of values)value.rows=0},snapshot(){return values.filter(value=>value.host.isConnected).map(value=>({nativeRowReplacements:value.rows,nativeScreenIntersecting:value.intersecting,nativeIntersection:value.intersection,screen:value.host.querySelector('.xterm-screen')?box(value.host.querySelector('.xterm-screen')!.getBoundingClientRect()):undefined,viewport:value.host.closest('.terminal-paint-viewport')?box(value.host.closest('.terminal-paint-viewport')!.getBoundingClientRect()):undefined,intersectionCallbacks:value.callbacks,renderer:value.host.dataset.terminalRenderer||'unknown'}))},dispose(){window.IntersectionObserver=NativeObserver;Element.prototype.replaceChildren=replace}}
 ;(window as unknown as {__nativeTerminalPaintDiagnostic:typeof controller}).__nativeTerminalPaintDiagnostic=controller
 window.addEventListener('pagehide',controller.dispose,{once:true})
}
