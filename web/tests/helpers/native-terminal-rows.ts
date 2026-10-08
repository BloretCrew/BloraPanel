// Test-only native visible-row control; never installed by the product.
import type {Terminal,IDisposable} from '@xterm/xterm'

export type TerminalPaintBounds={x:number;y:number;width:number;height:number}
type Controller={expose:(bounds:TerminalPaintBounds|undefined)=>void;dispose:()=>void}
const controllers=new WeakMap<HTMLElement,Controller>(),exposures=new WeakMap<HTMLElement,TerminalPaintBounds>()

// Geometry only. The original native renderer still owns every glyph/node;
// no terminal text, application snapshot or alternate renderer is retained.
export function exposeNativeTerminal(host:HTMLElement,bounds?:TerminalPaintBounds){
 if(bounds)exposures.set(host,bounds);else exposures.delete(host)
 controllers.get(host)?.expose(bounds)
}

// Native visible-row paint delegation. Parsing and protection remain complete;
// only rows outside the existing native overflow viewport skip DOM rebuilding.
// @xterm/xterm is pinned to 6.0.0. Confine the private native-render adapter
// here and fall back to complete native drawing if its required shape changes.
export function installNativeTerminalPaint(terminal:Terminal,host:HTMLElement):IDisposable{
 type NativeRenderer={renderRows:(start:number,end:number)=>void;_rowElements:HTMLElement[];_rowFactory:unknown;dimensions:{css:{cell:{width:number;height:number}}}}
 type NativeService={_renderer:{value?:NativeRenderer};_renderRows:(start:number,end:number)=>void;_needsFullRefresh?:boolean;_pausedResizeTask?:{flush:()=>void}}
 const core=(terminal as unknown as {_core?:{_renderService?:NativeService;coreService?:{decPrivateModes?:{synchronizedOutput?:boolean}}}})._core
 const service=core?._renderService,renderer=service?._renderer?.value
 if(!service||typeof service._renderRows!=='function'||typeof core?.coreService?.decPrivateModes?.synchronizedOutput!=='boolean'||!renderer||!Array.isArray(renderer._rowElements)||!renderer._rowFactory||typeof renderer.renderRows!=='function'||!terminal.element)return {dispose(){}}
 const original=renderer.renderRows,screen=terminal.element.querySelector<HTMLElement>('.xterm-screen')
 if(!screen)return {dispose(){}}
 let bounds:TerminalPaintBounds|undefined,dirty=false,complete=false,disposed=false
 // These native offsets are local to the unchanged full-size host. Window
 // translation and our paint viewport's inverse translation do not affect it.
 const screenLeft=terminal.element.offsetLeft+screen.offsetLeft,screenTop=terminal.element.offsetTop+screen.offsetTop
 const synchronized=()=>!!core?.coreService?.decPrivateModes?.synchronizedOutput
 const range=()=>{
  const cell=renderer.dimensions.css.cell
  if(!bounds||!Number.isFinite(cell.height)||cell.height<=0||!Number.isFinite(cell.width)||cell.width<=0)return {start:0,end:terminal.rows-1}
  // Retain two native rows for cursor/underline/italic and fractional-edge
  // overhang. Horizontal slits retain whole rows; never clip native glyphs.
  const right=screenLeft+cell.width*terminal.cols,bottom=screenTop+cell.height*terminal.rows
  if(bounds.x+bounds.width<screenLeft||bounds.x>right||bounds.y+bounds.height<screenTop||bounds.y>bottom||bounds.width<=0||bounds.height<=0)return {start:0,end:-1}
  return {start:Math.max(0,Math.floor((bounds.y-screenTop)/cell.height)-2),end:Math.min(terminal.rows-1,Math.ceil((bounds.y+bounds.height-screenTop)/cell.height)+2)}
 }
 const delegated:NativeRenderer['renderRows']=function(this:NativeRenderer,start,end){
  if(disposed||service._renderer.value!==renderer)return original.call(this,start,end)
  // Keep the native synchronized-output boundary even for direct native
  // focus/selection redraws. A reveal must never publish held buffer cells.
  if(synchronized()){dirty=true;return}
  const visible=complete?{start:0,end:terminal.rows-1}:range(),first=Math.max(start,visible.start),last=Math.min(end,visible.end)
  if(first>start||last<end)dirty=true
  if(first<=last)original.call(this,first,last)
  if(first===0&&last===terminal.rows-1)dirty=false
 }
 renderer.renderRows=delegated
 const redrawComplete=()=>{
  if((!dirty&&!service._needsFullRefresh)||synchronized()||disposed||service._renderer.value!==renderer)return
  complete=true
  try{service._pausedResizeTask?.flush();service._renderRows(0,terminal.rows-1)}finally{complete=false}
 }
 const controller:Controller={expose(next){
  const previous=bounds
  if(previous&&next&&previous.x===next.x&&previous.y===next.y&&previous.width===next.width&&previous.height===next.height)return
  bounds=next
  // Restore committed rows synchronously before the next native paint when
  // exposure expands. Native _renderRows enforces atomic output/selection.
  if((dirty||service._needsFullRefresh)&&(!next||!previous||next.x<previous.x||next.y<previous.y||next.x+next.width>previous.x+previous.width||next.y+next.height>previous.y+previous.height))redrawComplete()
 },dispose(){if(disposed)return;disposed=true;if(renderer.renderRows===delegated)renderer.renderRows=original;if(controllers.get(host)===controller)controllers.delete(host);delete host.dataset.terminalRowPaint;sync.dispose()}}
 // Before DECSET 2026 holds subsequent writes, refresh skipped committed
 // rows with the official renderer. Their existing native nodes then remain
 // the correct committed image on idle reveal during synchronized output.
 const sync=terminal.parser.registerCsiHandler({prefix:'?',final:'h'},params=>{if(params.some(value=>value===2026))redrawComplete();return false})
 controllers.set(host,controller);host.dataset.terminalRowPaint='native-visible-rows';controller.expose(exposures.get(host))
 return {dispose:controller.dispose}
}
