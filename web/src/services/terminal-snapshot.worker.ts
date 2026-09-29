import './terminal-worker-env'
import {Terminal as HeadlessTerminal} from '@xterm/headless'
import {SerializeAddon} from '@xterm/addon-serialize'
import type {Terminal} from '@xterm/xterm'
import {restoreTerminalControlState} from './terminal-control-state'
import {restoreTerminalLinks} from './terminal-links'

type Request = {id:number;kind:'init';cols:number;rows:number}
  | {id:number;kind:'write';data:string|Uint8Array}
  | {id:number;kind:'resize';cols:number;rows:number}
  | {id:number;kind:'controls';state:unknown;cols:number;rows:number}
  | {id:number;kind:'links';state:unknown}
  | {id:number;kind:'snapshot'}
let terminal:HeadlessTerminal|undefined
const serialize=new SerializeAddon()
let pending=Promise.resolve()
self.onmessage=(event:MessageEvent<Request>)=>{
  const request=event.data
  pending=pending.then(async()=>{
    if(request.kind==='init'){
      if(terminal)throw new Error('重复初始化终端检查点')
      terminal=new HeadlessTerminal({cols:request.cols,rows:request.rows,scrollback:3000,allowProposedApi:true})
      // Both terminals are pinned to the same upstream 6.0.0 commit. The
      // string serializer uses their shared buffer/parser API, without DOM.
      serialize.activate(terminal as unknown as Terminal)
    }else{
      if(!terminal)throw new Error('终端检查点尚未初始化')
      if(request.kind==='write')await new Promise<void>(resolve=>terminal!.write(request.data,resolve))
      else if(request.kind==='resize')terminal.resize(request.cols,request.rows)
      else if(request.kind==='controls')restoreTerminalControlState(terminal,request.state,request.cols,request.rows)
      else if(request.kind==='links')restoreTerminalLinks(terminal,request.state,terminal.cols,terminal.rows)
      else{
        self.postMessage({id:request.id,screen:serialize.serialize({scrollback:3000})})
        return
      }
    }
    self.postMessage({id:request.id})
  }).catch(error=>{self.postMessage({id:request.id,error:error instanceof Error?error.message:'终端检查点 Worker 失败'})})
}
