// Pinned xterm 6: SerializeAddon emits link text, but omits OSC link metadata.
// Use sparse extended-attribute cells, not a scan of every scrollback column.
type Link={uri:string;id?:string}
type Attributes={fg:number;bg:number;ext:number;link?:number}
type Span=Attributes&{row:number;from:number;to:number;link:number}
export type TerminalLinks={version:1;links:Link[];normal:Span[];alternate:Span[];current:Attributes}
type Extended={_ext:number;ext:number;urlId:number;clone:()=>Extended}
type Cell={fg:number;bg:number;extended:Extended;hasExtendedAttrs:()=>number;updateExtended:()=>void}
type Line={length:number;_extendedAttrs:Record<number,Extended>;loadCell:(x:number,cell:Cell)=>Cell;setCell:(x:number,cell:Cell)=>void}
type Buffer={y:number;ybase:number;lines:{length:number;get:(y:number)=>Line|undefined};getNullCell:()=>Cell}
type Core={_bufferService:{cols:number;rows:number;buffers:{normal:Buffer;alt:Buffer;_activeBuffer:Buffer}};_oscLinkService:{_dataByLinkId:Map<number,unknown>;getLinkData:(id:number)=>Link|undefined;registerLink:(data:Link)=>number;addLineToLink:(id:number,row:number)=>void};_inputHandler:{_curAttrData:Cell}}
function core(terminal:unknown):Core{
  const c=(terminal as {_core?:Core})._core
  if(!c?._oscLinkService||!c._bufferService?.buffers||!c._inputHandler?._curAttrData)throw new Error('终端链接状态接口不兼容')
  return c
}
function newCell(buffer:Buffer):Cell{
  const Constructor=buffer.getNullCell().constructor as new()=>Cell
  return new Constructor()
}
export function captureTerminalLinks(terminal:unknown):TerminalLinks|undefined{
  const c=core(terminal),links:Link[]=[],ids=new Map<number,number>()
  if(c._oscLinkService._dataByLinkId.size===0)return undefined
  const link=(id:number)=>{
    if(ids.has(id))return ids.get(id)!
    const data=c._oscLinkService.getLinkData(id)
    if(!data)return undefined
    const index=links.length;links.push({...data});ids.set(id,index);return index
  }
  const attrs=(cell:Cell):Attributes=>({fg:cell.fg,bg:cell.bg,ext:cell.extended._ext,...(cell.hasExtendedAttrs()&&cell.extended.urlId?{link:link(cell.extended.urlId)}:{})})
  const capture=(buffer:Buffer):Span[]=>{
    const spans:Span[]=[],cell=newCell(buffer)
    for(let row=0;row<buffer.lines.length;row++){
      const line=buffer.lines.get(row)!
      for(const key of Object.keys(line._extendedAttrs)){
        const x=Number(key);if(x>=line.length)continue
        line.loadCell(x,cell)
        if(!cell.hasExtendedAttrs()||!cell.extended.urlId)continue
        const a=attrs(cell);if(a.link===undefined)continue
        const previous=spans.at(-1)
        if(previous&&previous.row===row-buffer.ybase&&previous.to===x&&previous.link===a.link&&previous.fg===a.fg&&previous.bg===a.bg&&previous.ext===a.ext)previous.to++
        else spans.push({...a,link:a.link,row:row-buffer.ybase,from:x,to:x+1})
      }
    }
    return spans
  }
  const normal=capture(c._bufferService.buffers.normal),alternate=capture(c._bufferService.buffers.alt),current=attrs(c._inputHandler._curAttrData)
  return links.length?validate({version:1,links,normal,alternate,current},c,c._bufferService.cols,c._bufferService.rows):undefined
}
function number(value:unknown,min:number,max:number):number{
  if(typeof value!=='number'||!Number.isInteger(value)||value<min||value>max)throw new Error('终端链接状态数值无效')
  return value
}
function validate(value:unknown,c:Core,cols:number,rows:number):TerminalLinks{
  const s=value as TerminalLinks,max=(3000+2*rows)*cols
  if(!s||s.version!==1||!Array.isArray(s.links)||s.links.length>max+1)throw new Error('终端链接恢复记录无效')
  const links=s.links.map(data=>{
    if(!data||typeof data.uri!=='string'||data.uri.length>10000000||(data.id!==undefined&&(typeof data.id!=='string'||data.id.length>10000000)))throw new Error('终端链接目标无效')
    return {uri:data.uri,...(data.id===undefined?{}:{id:data.id})}
  })
  const attrs=(a:Attributes):Attributes=>{
    if(!a||typeof a!=='object')throw new Error('终端链接属性无效')
    return {fg:number(a.fg,-2147483648,4294967295),bg:number(a.bg,-2147483648,4294967295),ext:number(a.ext,-2147483648,4294967295),...(a.link===undefined?{}:{link:number(a.link,0,links.length-1)})}
  }
  const spans=(input:Span[],buffer:Buffer):Span[]=>{
    if(!Array.isArray(input)||input.length>max)throw new Error('终端链接范围无效')
    let previousRow=-3001,previousEnd=0
    return input.map(s=>{
      const a=attrs(s),row=number(s.row,-3000,rows-1),from=number(s.from,0,cols-1),to=number(s.to,from+1,cols)
      if(a.link===undefined||row<previousRow||(row===previousRow&&from<previousEnd)||!buffer.lines.get(buffer.ybase+row))throw new Error('终端链接范围顺序无效')
      previousRow=row;previousEnd=to;return {...a,link:a.link,row,from,to}
    })
  }
  return {version:1,links,normal:spans(s.normal,c._bufferService.buffers.normal),alternate:spans(s.alternate,c._bufferService.buffers.alt),current:attrs(s.current)}
}
export function restoreTerminalLinks(terminal:unknown,value:unknown,cols:number,rows:number){
  const c=core(terminal),s=validate(value,c,cols,rows),buffers=c._bufferService.buffers,active=buffers._activeBuffer
  const mappings=new Map<Buffer,Map<number,number>>()
  const register=(buffer:Buffer,index:number,row:number)=>{
    let map=mappings.get(buffer);if(!map){map=new Map();mappings.set(buffer,map)}
    const oldY=buffer.y
    // Registration uses the active buffer for marker ownership. Switch only
    // the internal lookup during this synchronous call, without triggering
    // buffer activation (which would erase alternate-screen content).
    buffers._activeBuffer=buffer;buffer.y=row-buffer.ybase
    try{
      let id=map.get(index)
      if(id===undefined){id=c._oscLinkService.registerLink(s.links[index]!);map.set(index,id)}
      else c._oscLinkService.addLineToLink(id,row)
      return id
    }finally{buffer.y=oldY;buffers._activeBuffer=active}
  }
  const apply=(cell:Cell,a:Attributes,id:number)=>{
    cell.fg=a.fg;cell.bg=a.bg;cell.extended=cell.extended.clone();cell.extended.ext=a.ext;cell.extended.urlId=id;cell.updateExtended()
  }
  const restore=(buffer:Buffer,spans:Span[])=>{
    const cell=newCell(buffer)
    for(const span of spans){
      const row=buffer.ybase+span.row,line=buffer.lines.get(row)!,id=register(buffer,span.link,row)
      for(let x=span.from;x<span.to;x++){line.loadCell(x,cell);apply(cell,span,id);line.setCell(x,cell)}
    }
  }
  restore(buffers.normal,s.normal);restore(buffers.alt,s.alternate)
  apply(c._inputHandler._curAttrData,s.current,s.current.link===undefined?0:register(active,s.current.link,active.ybase+active.y))
}
