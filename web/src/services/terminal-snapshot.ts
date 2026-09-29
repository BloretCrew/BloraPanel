import SnapshotWorker from './terminal-snapshot.worker?worker'

// A single in-flight parser operation per view bounds the mirror queue. Its
// output is never forwarded to a PTY: this worker only computes checkpoints.
export class TerminalSnapshot {
  private worker=new SnapshotWorker()
  private next=0
  private pending=new Map<number,{resolve:(screen?:string)=>void;reject:(error:Error)=>void;timer:ReturnType<typeof setTimeout>}>()
  private closed=false
  constructor(){
    this.worker.onmessage=(event:MessageEvent<{id:number;screen?:string;error?:string}>)=>{
      const reply=this.pending.get(event.data.id)
      if(!reply)return
      this.pending.delete(event.data.id);clearTimeout(reply.timer)
      if(event.data.error)reply.reject(new Error(event.data.error))
      else reply.resolve(event.data.screen)
    }
    this.worker.onerror=()=>this.dispose()
    this.worker.onmessageerror=()=>this.dispose()
  }
  private request(value:Record<string,unknown>){
    if(this.closed)return Promise.reject(new Error('终端检查点 Worker 已关闭'))
    const id=++this.next
    return new Promise<string|undefined>((resolve,reject)=>{
      const timer=setTimeout(()=>this.dispose(),30000)
      this.pending.set(id,{resolve,reject,timer})
      try{this.worker.postMessage({...value,id})}catch{this.dispose()}
    })
  }
  initialize(cols:number,rows:number){return this.request({kind:'init',cols,rows})}
  write(data:string|Uint8Array){return this.request({kind:'write',data})}
  resize(cols:number,rows:number){return this.request({kind:'resize',cols,rows})}
  restoreControls(state:unknown,cols:number,rows:number){return this.request({kind:'controls',state,cols,rows})}
  async screen(){
    const screen=await this.request({kind:'snapshot'})
    if(typeof screen!=='string')throw new Error('终端检查点 Worker 未返回屏幕')
    return screen
  }
  dispose(){
    if(this.closed)return
    this.closed=true;this.worker.terminate()
    for(const {reject,timer} of this.pending.values()){
      clearTimeout(timer);reject(new Error('终端检查点 Worker 不可用'))
    }
    this.pending.clear()
  }
}
