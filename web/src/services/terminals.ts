import {Terminal} from '@xterm/xterm'
import {SerializeAddon} from '@xterm/addon-serialize'
import {FitAddon} from '@xterm/addon-fit'
import {WebglAddon} from '@xterm/addon-webgl'
import {TerminalSnapshot} from './terminal-snapshot'
import type {RecoveryService} from '../recovery/service'
import {json,TERMINAL_OUTPUT_JOURNAL_BYTES,TERMINAL_OUTPUT_JOURNAL_EVENTS} from '../recovery/state'
import '@xterm/xterm/css/xterm.css'
export interface TerminalEvent {sequence:number;kind:'output'|'resize';data?:string;cols?:number;rows?:number}
function decodeOutput(value:string){
  const binary=atob(value),bytes=new Uint8Array(binary.length)
  for(let index=0;index<binary.length;index++)bytes[index]=binary.charCodeAt(index)
  return bytes
}
const terminalColors={
  light:{background:'#111917',foreground:'#d6e6de'},
  dark:{background:'#111418',foreground:'#e6ebf0'},
}
export class TerminalModel {
  readonly terminal:Terminal
  readonly serialize=new SerializeAddon()
  readonly fit=new FitAddon()
  replaying=true
  writable=false
  sequence=0
  private parsed=Promise.resolve()
  private parsing=false
  private disposed=false
  private snapshot?:TerminalSnapshot
  private container?:HTMLElement
  private outputJournal:{sequence:number;data:string}[]=[]
  private journalBytes=0
  readonly checkpointKey:string
  constructor(readonly recovery:RecoveryService,readonly sessionId:string,readonly viewTabId:string,readonly onInput:(value:string)=>void,mode:'light'|'dark'='light'){
    this.checkpointKey=`${sessionId}:${viewTabId}`
    const saved=recovery.state.terminals[this.checkpointKey]||recovery.state.terminals[sessionId]
    this.sequence=saved?.sequence||0
    this.terminal=new Terminal({cols:saved?.cols||100,rows:saved?.rows||28,scrollback:3000,allowProposedApi:true,disableStdin:true,fontSize:Number(recovery.state.views[viewTabId]?.state.fontSize||13),fontFamily:'Consolas,"Liberation Mono",monospace',theme:terminalColors[mode]})
    this.terminal.loadAddon(this.serialize);this.terminal.loadAddon(this.fit)
    this.terminal.onData(data=>{if(!this.replaying && this.writable)onInput(data)})
    this.terminal.onScroll(()=>{if(!this.replaying && !this.parsing)this.protectView()})
    this.terminal.onSelectionChange(()=>{if(!this.replaying && !this.parsing)this.protectView()})
  }
  async mount(container:HTMLElement){
    this.container=container
    this.terminal.open(container)
    try{
      this.snapshot=new TerminalSnapshot()
      await this.snapshot.initialize(this.terminal.cols,this.terminal.rows)
      container.dataset.terminalCheckpoint='worker'
    }catch(error){if(import.meta.env.DEV)console.debug('Terminal checkpoint fallback',error);this.disableSnapshot()}
    container.dataset.terminalRenderer='default'
    try{
      // Software WebGL can be much slower than xterm's default renderer. Do
      // not opt into it on machines where hardware acceleration is unavailable.
      const probe=document.createElement('canvas').getContext('webgl2',{failIfMajorPerformanceCaveat:true})
      if(probe){
        const diagnostic=probe.getExtension('WEBGL_debug_renderer_info')
        const renderer=diagnostic?String(probe.getParameter(diagnostic.UNMASKED_RENDERER_WEBGL)):''
        // Some ANGLE configurations accept failIfMajorPerformanceCaveat even
        // while reporting a software Vulkan device. Honor that actual result.
        const software=/SwiftShader|llvmpipe|softpipe|software|Microsoft Basic Render/i.test(renderer)
        probe.getExtension('WEBGL_lose_context')?.loseContext()
        if(!software){
          const webgl=new WebglAddon()
          webgl.onContextLoss(()=>{webgl.dispose();container.dataset.terminalRenderer='default'})
          try{this.terminal.loadAddon(webgl);container.dataset.terminalRenderer='webgl'}catch(error){webgl.dispose();throw error}
        }
      }
    }catch{/* xterm's default renderer remains available. */}
    const saved=this.recovery.state.terminals[this.checkpointKey]||this.recovery.state.terminals[this.sessionId]
    if(saved){
      if(saved.outputJournal&&(!Array.isArray(saved.outputJournal)||saved.outputJournal.length>TERMINAL_OUTPUT_JOURNAL_EVENTS))throw new Error('终端本地输出日志超出条目预算')
      await this.write(saved.screen)
      let sequence=saved.baseSequence??saved.sequence
      for(const event of saved.outputJournal||[]){
        if(event.sequence!==sequence+1||typeof event.data!=='string')throw new Error('终端本地输出日志不连续，原记录已保留')
        this.journalBytes+=event.data.length*2
        if(this.journalBytes>TERMINAL_OUTPUT_JOURNAL_BYTES)throw new Error('终端本地输出日志超出保护预算')
        await this.write(decodeOutput(event.data))
        sequence=event.sequence
      }
      if(sequence!==saved.sequence)throw new Error('终端本地输出日志与检查点序号不匹配')
      this.outputJournal=(saved.outputJournal||[]).map(event=>({...event}))
      this.terminal.scrollToLine(saved.scroll);if(saved.selection){const {start,end}=saved.selection;this.terminal.select(start.x,start.y,(end.y-start.y)*saved.cols+end.x-start.x)}
    }
    // Only the server's replay-complete message enables outbound terminal responses.
  }
  private disableSnapshot(){this.snapshot?.dispose();this.snapshot=undefined;if(this.container)this.container.dataset.terminalCheckpoint='main'}
  private async mirror(operation:(snapshot:TerminalSnapshot)=>Promise<unknown>){
    if(this.snapshot)try{await operation(this.snapshot)}catch(error){if(import.meta.env.DEV)console.debug('Terminal checkpoint fallback',error);this.disableSnapshot()}
  }
  private async write(data:string|Uint8Array){
    await new Promise<void>(resolve=>this.terminal.write(data,resolve))
    await this.mirror(snapshot=>snapshot.write(data))
  }
  receive(event:TerminalEvent){return this.receiveBatch([event])}
  receiveBatch(events:TerminalEvent[]){
    this.parsed=this.parsed.then(async()=>{
      if(this.disposed)return
      let expected=this.sequence
      const fresh:TerminalEvent[]=[]
      for(const event of events){
        if(!Number.isSafeInteger(event.sequence)||event.sequence<1)throw new Error('终端归档事件序号无效')
        if(event.sequence<=this.sequence)continue
        if(event.sequence!==++expected)throw new Error('终端归档事件有缺口，原检查点已保留')
        if(event.kind==='output'){if(typeof event.data!=='string')throw new Error('终端输出数据缺失')}
        else if(event.kind==='resize'){if(!Number.isInteger(event.cols)||!Number.isInteger(event.rows)||!event.cols||!event.rows||event.cols<1||event.rows<1||event.cols>1000||event.rows>1000)throw new Error('终端尺寸事件无效')}
        else throw new Error('不支持的终端归档事件')
        fresh.push(event)
      }
      for(let index=0;index<fresh.length;){
        const group:TerminalEvent[]=[]
        const first=fresh[index]!
        this.parsing=true
        try{
          if(first.kind==='resize'){this.terminal.resize(first.cols!,first.rows!);await this.mirror(snapshot=>snapshot.resize(first.cols!,first.rows!));group.push(first);index++}
          else{
            const chunks:Uint8Array[]=[];let size=0
            while(index<fresh.length&&fresh[index]!.kind==='output'){
              const event=fresh[index]!,data=decodeOutput(event.data!)
              if(data.length>64*1024)throw new Error('终端输出块超过解析预算')
              if(size+data.length>64*1024)break
              chunks.push(data);group.push(event);size+=data.length;index++
            }
            const joined=chunks.length===1?chunks[0]!:new Uint8Array(size)
            if(chunks.length>1){let offset=0;for(const chunk of chunks){joined.set(chunk,offset);offset+=chunk.length}}
            await this.write(joined)
          }
          this.sequence=group.at(-1)!.sequence
        }finally{this.parsing=false}
        await this.protectOutputs(group)
        if(!this.recovery.status.protected){
          await this.recovery.awaitPendingWrites()
          if(!this.recovery.status.protected)throw new Error('终端输出尚未持久保护，已停止确认流量')
        }
      }
    })
    return this.parsed
  }
  setLease(writable:boolean){this.writable=writable;this.terminal.options.disableStdin=this.replaying||!writable}
  async restoreComplete(){
    await this.parsed;await this.checkpoint()
    if(!this.recovery.status.protected){await this.recovery.awaitPendingWrites();if(!this.recovery.status.protected)throw new Error('终端恢复检查点尚未持久保护')}
    this.replaying=false;this.setLease(this.writable)
  }
  proposedSize(){return this.fit.proposeDimensions()}
  setFontSize(value:number){this.terminal.options.fontSize=value}
  setColorMode(mode:'light'|'dark'){this.terminal.options.theme=terminalColors[mode]}
  private protectView(){
    if(this.disposed||!this.recovery.state.terminals[this.checkpointKey])return
    const selection=this.terminal.getSelectionPosition()
    this.recovery.commit([
      {kind:'set',path:['terminals',this.checkpointKey,'scroll'],value:this.terminal.buffer.active.viewportY},
      selection?{kind:'set',path:['terminals',this.checkpointKey,'selection'],value:json(selection)}:{kind:'delete',path:['terminals',this.checkpointKey,'selection']},
    ])
  }
  private async protectOutputs(events:TerminalEvent[]){
    const saved=this.recovery.state.terminals[this.checkpointKey]
    const addedBytes=events.reduce((sum,event)=>sum+(event.data?.length||0)*2,0)
    if(events.some(event=>event.kind!=='output')||!saved||this.outputJournal.length+events.length>TERMINAL_OUTPUT_JOURNAL_EVENTS||this.journalBytes+addedBytes>TERMINAL_OUTPUT_JOURNAL_BYTES){await this.checkpoint();return}
    // Persist parsed remote output immediately, without serializing all 3000
    // scrollback lines or recopying the output journal on every chunk. Only
    // output is replayed locally; user input is never journaled or replayed.
    this.outputJournal.push(...events.map(event=>({sequence:event.sequence,data:event.data!})))
    this.journalBytes+=addedBytes
    this.recovery.commit([
      {kind:'terminal-output',checkpointKey:this.checkpointKey,baseSequence:saved.baseSequence??saved.sequence,previousSequence:saved.sequence,sequence:this.sequence,events:events.map(event=>({sequence:event.sequence,data:event.data!})),scroll:this.terminal.buffer.active.viewportY},
    ])
  }
  async checkpoint(){
    if(this.parsing || this.disposed)return
    // xterm serialization includes alternate screen, cursor and modes in this pinned version.
    // Parse callbacks establish the sequence boundary; no historical input is stored.
    let screen:string|undefined
    if(this.snapshot)try{screen=await this.snapshot.screen()}catch(error){if(import.meta.env.DEV)console.debug('Terminal checkpoint fallback',error);this.disableSnapshot()}
    if(this.disposed)return
    screen??=this.serialize.serialize({scrollback:3000})
    const checkpoint={sessionId:this.sessionId,viewTabId:this.viewTabId,sequence:this.sequence,cols:this.terminal.cols,rows:this.terminal.rows,screen,scroll:this.terminal.buffer.active.viewportY,selection:this.terminal.getSelectionPosition()}
    this.recovery.commit([{kind:'set',path:['terminals',this.checkpointKey],value:json(checkpoint)}])
    this.outputJournal=[];this.journalBytes=0
  }
  dispose(){this.disposed=true;this.snapshot?.dispose();this.snapshot=undefined;this.terminal.dispose()}
}
