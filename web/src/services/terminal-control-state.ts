// Supplement the pinned xterm 6 serializer with state it does not emit. Both
// browser and headless terminals use this adapter; upgrades require behavioral
// continuation tests, not just comparison of the serialized screen.
type Charset=Record<string,string>|undefined
type BufferState={x:number;y:number;scrollTop:number;scrollBottom:number;tabs:number[];savedX:number;savedY:number;fg:number;bg:number;savedCharset:Charset}
type CursorStyle='block'|'underline'|'bar'
type ProtocolState={mouseProtocol:'NONE'|'X10'|'VT200'|'DRAG'|'ANY';mouseEncoding:'DEFAULT'|'SGR'|'SGR_PIXELS';cursorHidden:boolean;cursorBlink:boolean;cursorStyle:CursorStyle;convertEol:boolean;overrideStyle?:CursorStyle;overrideBlink?:boolean}
export type TerminalControlState={version:1;glevel:number;charsets:(Charset|null)[];charset:Charset;normal:BufferState;alternate:BufferState;protocol?:ProtocolState}
type InternalBuffer={x:number;y:number;ybase:number;scrollTop:number;scrollBottom:number;tabs:Record<number,boolean|undefined>;savedX:number;savedY:number;savedCurAttrData:{fg:number;bg:number};savedCharset:Charset}
type Core={_charsetService:{glevel:number;_charsets:Charset[];charset:Charset};_bufferService:{buffers:{normal:InternalBuffer;alt:InternalBuffer}};coreMouseService:{activeProtocol:string;activeEncoding:string};coreService:{isCursorHidden:boolean;decPrivateModes:{cursorStyle?:CursorStyle;cursorBlink?:boolean}};optionsService:{rawOptions:{cursorBlink:boolean;cursorStyle:CursorStyle;convertEol:boolean};options:{cursorBlink:boolean;cursorStyle:CursorStyle;convertEol:boolean}}}
function core(terminal:unknown):Core{
  const value=(terminal as {_core?:Core})._core
  if(!value?._charsetService||!value._bufferService?.buffers?.normal||!value._bufferService.buffers.alt)throw new Error('终端控制状态接口不兼容')
  return value
}
function charset(value:unknown):Charset{
  if(value==null)return undefined
  if(typeof value!=='object'||Array.isArray(value)||Object.keys(value).length>128)throw new Error('终端字符集状态无效')
  const result:Record<string,string>={}
  for(const [key,text] of Object.entries(value)){
    if(key.length!==1||typeof text!=='string'||text.length>4)throw new Error('终端字符集映射无效')
    result[key]=text
  }
  return result
}
function integer(value:unknown,min:number,max:number):number{
  if(typeof value!=='number'||!Number.isInteger(value)||value<min||value>max)throw new Error('终端控制状态数值无效')
  return value
}
function choice<T extends string>(value:unknown,choices:readonly T[]):T{
  if(typeof value!=='string'||!choices.includes(value as T))throw new Error('终端协议状态枚举无效')
  return value as T
}
function boolean(value:unknown):boolean{
  if(typeof value!=='boolean')throw new Error('终端协议状态布尔值无效')
  return value
}
function protocol(value:unknown):ProtocolState{
  if(!value||typeof value!=='object')throw new Error('终端协议状态无效')
  const p=value as ProtocolState,styles=['block','underline','bar'] as const
  return {mouseProtocol:choice(p.mouseProtocol,['NONE','X10','VT200','DRAG','ANY'] as const),mouseEncoding:choice(p.mouseEncoding,['DEFAULT','SGR','SGR_PIXELS'] as const),cursorHidden:boolean(p.cursorHidden),cursorBlink:boolean(p.cursorBlink),cursorStyle:choice(p.cursorStyle,styles),convertEol:boolean(p.convertEol),...(p.overrideStyle===undefined?{}:{overrideStyle:choice(p.overrideStyle,styles)}),...(p.overrideBlink===undefined?{}:{overrideBlink:boolean(p.overrideBlink)})}
}
function buffer(value:unknown,cols:number,rows:number):BufferState{
  if(!value||typeof value!=='object')throw new Error('终端缓冲区状态无效')
  const b=value as BufferState
  const top=integer(b.scrollTop,0,rows-1),bottom=integer(b.scrollBottom,top,rows-1)
  if(!Array.isArray(b.tabs)||b.tabs.length>1000)throw new Error('终端制表位状态无效')
  return {x:integer(b.x,0,cols),y:integer(b.y,0,rows-1),scrollTop:top,scrollBottom:bottom,tabs:b.tabs.map(x=>integer(x,0,999)),savedX:integer(b.savedX,0,cols),savedY:integer(b.savedY,-3000,rows-1),fg:integer(b.fg,-2147483648,4294967295),bg:integer(b.bg,-2147483648,4294967295),savedCharset:charset(b.savedCharset)}
}
export function validateTerminalControlState(value:unknown,cols:number,rows:number):TerminalControlState{
  if(!value||typeof value!=='object')throw new Error('终端控制状态无效')
  const state=value as TerminalControlState
  if(state.version!==1||!Array.isArray(state.charsets)||state.charsets.length>4)throw new Error('终端控制状态版本无效')
  return {version:1,glevel:integer(state.glevel,0,3),charsets:state.charsets.map(x=>charset(x)??null),charset:charset(state.charset),normal:buffer(state.normal,cols,rows),alternate:buffer(state.alternate,cols,rows),...(state.protocol===undefined?{}:{protocol:protocol(state.protocol)})}
}
export function captureTerminalControlState(terminal:unknown,cols:number,rows:number):TerminalControlState{
  const c=core(terminal)
  const opts=c.optionsService.rawOptions,modes=c.coreService.decPrivateModes
  const capture=(b:InternalBuffer):BufferState=>({x:b.x,y:b.y,scrollTop:b.scrollTop,scrollBottom:b.scrollBottom,tabs:Object.keys(b.tabs).filter(x=>b.tabs[Number(x)]).map(Number),savedX:b.savedX,savedY:b.savedY-b.ybase,fg:b.savedCurAttrData.fg,bg:b.savedCurAttrData.bg,savedCharset:b.savedCharset})
  return validateTerminalControlState({version:1,glevel:c._charsetService.glevel,charsets:Array.from(c._charsetService._charsets,x=>x??null),charset:c._charsetService.charset,normal:capture(c._bufferService.buffers.normal),alternate:capture(c._bufferService.buffers.alt),protocol:{mouseProtocol:c.coreMouseService.activeProtocol,mouseEncoding:c.coreMouseService.activeEncoding,cursorHidden:c.coreService.isCursorHidden,cursorBlink:opts.cursorBlink,cursorStyle:opts.cursorStyle,convertEol:opts.convertEol,overrideStyle:modes.cursorStyle,overrideBlink:modes.cursorBlink}},cols,rows)
}
export function restoreTerminalControlState(terminal:unknown,value:unknown,cols:number,rows:number){
  // Validate the whole record before mutating either buffer. Assign only known
  // fields, retaining native AttributeData instances and their methods.
  const s=validateTerminalControlState(value,cols,rows),c=core(terminal)
  const restore=(b:InternalBuffer,s:BufferState)=>{
    b.x=s.x;b.y=s.y;b.scrollTop=s.scrollTop;b.scrollBottom=s.scrollBottom
    b.tabs=Object.fromEntries(s.tabs.map(x=>[x,true]))
    b.savedX=s.savedX;b.savedY=b.ybase+s.savedY
    b.savedCurAttrData.fg=s.fg;b.savedCurAttrData.bg=s.bg;b.savedCharset=s.savedCharset
  }
  restore(c._bufferService.buffers.normal,s.normal);restore(c._bufferService.buffers.alt,s.alternate)
  c._charsetService.glevel=s.glevel;c._charsetService._charsets=s.charsets.map(x=>x??undefined);c._charsetService.charset=s.charset
  if(s.protocol){
    const p=s.protocol
    // Native setters emit the option/protocol change notifications needed by
    // browser mouse bindings and rendering. Never restore a stale mouse event.
    c.optionsService.options.cursorBlink=p.cursorBlink;c.optionsService.options.cursorStyle=p.cursorStyle;c.optionsService.options.convertEol=p.convertEol
    c.coreService.isCursorHidden=p.cursorHidden
    c.coreService.decPrivateModes.cursorStyle=p.overrideStyle;c.coreService.decPrivateModes.cursorBlink=p.overrideBlink
    c.coreMouseService.activeEncoding=p.mouseEncoding;c.coreMouseService.activeProtocol=p.mouseProtocol
  }
}
