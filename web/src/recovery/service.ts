import { openDB, type IDBPDatabase } from 'idb'
import { reactive, toRaw, watch, type WatchHandle } from 'vue'
import { id, type Workspace } from '../app-host/types'
import { applyEntry, copy, migrateWorkspace, trimDraftHistory, type JournalEntry, type Mutation } from './state'
import {isTerminalDelta,restoreTerminalDelta,terminalDelta,type TerminalDelta} from './terminal-delta'
import {BackgroundSnapshot,BackgroundSnapshotUnavailableError} from './background-snapshot'

const TAIL_LIMIT = 256 * 1024
export class RecoveryService {
  state: Workspace
  status = reactive({ protected: true, message: '本地工作现场已保护', saving: false })
  private tail: JournalEntry[] = []
  private pending: Promise<void> = Promise.resolve()
  private flushing = false
  private db?: IDBPDatabase
  private restoreFailure = false
  private retained: unknown[] = []
  private coldVersion=0
  private externalVersion=0
  private applying=false
  private coldWatchers:WatchHandle[]=[]
  private committing=false
  private commitQueue:Mutation[][]=[]
  private background?:BackgroundSnapshot
  private mirrored?:{revision:number;externalVersion:number}
  private flushTimer?:ReturnType<typeof setTimeout>
  private idleFullTimer?:ReturnType<typeof setTimeout>
  private outputSinceIdle=false
  private forceFull=false
  private enabling?:Promise<void>
  private backgroundGeneration=0
  private durable?:{current:string;base:string;baseRevision:number;revision:number;entries:JournalEntry[];coldVersion:number}
  readonly key: string
  constructor(readonly userId: string, readonly deviceId: string, readonly browserTabId: string, readonly storage: Storage = sessionStorage, readonly dbName = 'blora-workspaces', readonly slot = 'main') {
    this.key = `${userId}:${deviceId}:${browserTabId}${slot === 'main' ? '' : ':' + slot}`
    this.state = reactive({ schemaVersion: 1, revision: 0, userId, deviceId, browserTabId, workspaceId: id('workspace'), windows: {}, views: {}, order: [], shortcuts: {}, drafts: {}, terminals: {},uploads:{}, preferences: {}, closedViews: [] })
    // Existing callers may prepare reactive app metadata before committing.
    // Such changes must not disappear merely because the next commit is PTY
    // output. Track cold branches separately; output journal/sequence/scroll
    // changes do not traverse editor history or terminal output on each chunk.
    const invalidate=()=>{this.coldVersion++;if(!this.applying)this.externalVersion++}
    for(const key of ['schemaVersion','userId','deviceId','browserTabId','workspaceId','windows','views','order','activeWindowId','shortcuts','drafts','uploads','preferences','closedViews'] as const)this.coldWatchers.push(watch(()=>this.state[key],invalidate,{deep:true,flush:'sync'}))
    const outputKeys=new Set(['sequence','baseSequence','outputJournal','scroll'])
    this.coldWatchers.push(watch(()=>Object.fromEntries(Object.keys(this.state.terminals).map(key=>{
      const checkpoint=this.state.terminals[key]! as unknown as Record<string,unknown>
      return [key,Object.fromEntries(Object.keys(checkpoint).filter(name=>!outputKeys.has(name)).map(name=>[name,checkpoint[name]]))]
    })),invalidate,{deep:true,flush:'sync'}))
  }
  async restore(persist=true) {
    try {
      this.db = await openDB(this.dbName, 1, { upgrade(db) { db.createObjectStore('snapshots'); db.createObjectStore('pointers') } })
      const pointer = await this.db.get('pointers', this.key)
      if (pointer) {
        const snapshot = await this.db.get('snapshots', pointer.current)
        this.retained.push(snapshot)
        let restored:Workspace
        if(isTerminalDelta(snapshot)){
          if(pointer.base!==snapshot.base)throw new Error('终端增量指针不匹配，已保留原恢复记录')
          const base=await this.db.get('snapshots',snapshot.base)
          this.retained.push(base)
          restored=restoreTerminalDelta(snapshot,base,this.key,pointer.current)
        }else restored = migrateWorkspace(snapshot)
        if (restored.userId !== this.userId || restored.deviceId !== this.deviceId || restored.browserTabId !== this.browserTabId) throw new Error('工作现场身份不匹配')
        Object.assign(this.state, restored)
      }
      const raw = this.storage.getItem(`blora:tail:${this.key}`)
      if(raw)this.retained.push(raw)
      this.tail = raw ? JSON.parse(raw) : []
      for (const entry of this.tail) applyEntry(this.state, entry)
      for(const draft of Object.values(this.state.drafts))trimDraftHistory(draft)
      this.retained=[]
      if(persist)await this.flush()
    } catch (error) { this.restoreFailure=true;this.fail(error) }
    return this.state
  }
  commit(mutations: Mutation[]) {
    // A synchronous Vue watcher can commit while an outer applyEntry has not
    // yet advanced the revision or protected its tail. Drain those mutations
    // after the outer entry is fully applied/protected, with distinct revisions.
    // Every queued entry is still synchronously protected before this outer
    // commit returns; no microtask, database wait or remote operation is added.
    this.commitQueue.push(copy(mutations))
    if(this.committing)return
    this.committing=true
    try{while(this.commitQueue.length)this.commitEntry(this.commitQueue.shift()!)}
    finally{this.committing=false;this.commitQueue.length=0}
  }
  private commitEntry(mutations:Mutation[]) {
    if(this.idleFullTimer)clearTimeout(this.idleFullTimer)
    this.idleFullTimer=undefined
    if(mutations.some(op=>op.kind==='terminal-output'))this.outputSinceIdle=true
    const entry = { revision: this.state.revision + 1, mutations }
    this.applying=true
    // One edit updates history, cursor, body and budget metadata together.
    // Deep observers must see its completed transaction, rather than traverse
    // the growing history after every intermediate property/array mutation.
    // Resume synchronously while applying remains true: external direct writes
    // still invalidate the worker mirror, and this commit stays journal-owned.
    for(const observer of this.coldWatchers)observer.pause()
    try{applyEntry(this.state, entry)}finally{
      try{for(const observer of this.coldWatchers)observer.resume()}
      finally{this.applying=false}
    }
    this.tail.push(entry)
    if(this.restoreFailure){this.fail(new Error('原恢复记录无法读取，已禁止覆盖；请导出保留记录和当前内容'));return}
    try {
      const tail = JSON.stringify(this.tail)
      if (tail.length * 2 > TAIL_LIMIT) throw new Error('最近输入超过同步保护预算，正在写入本地数据库；完成前请勿刷新，可导出当前现场')
      this.storage.setItem(`blora:tail:${this.key}`, tail)
      this.status.protected = true
      this.status.message = '本地工作现场已保护'
    } catch (error) { this.fail(error) }
    // Every entry is already synchronously durable in sessionStorage. Bound
    // background database wakeups, rather than waking another thread for each
    // PTY batch. Explicit flush/await, quota pressure and clearing force this
    // transaction immediately; output parsing and ACKs are never delayed.
    if(this.background){if(!this.flushing&&!this.flushTimer)this.flushTimer=setTimeout(()=>{this.flushTimer=undefined;void this.flush()},64)}
    else void this.flush()
  }
  awaitPendingWrites():Promise<void>{return this.flushTimer?this.flush():this.pending}
  get persistenceMode(){return this.background?'worker':'foreground'}
  enableBackgroundPersistence(){return this.enabling||=(this.startBackgroundPersistence().finally(()=>{this.enabling=undefined}))}
  private async startBackgroundPersistence(){
    const generation=this.backgroundGeneration
    await this.pending
    if(this.background||!this.db||this.restoreFailure||typeof Worker==='undefined')return
    let background:BackgroundSnapshot|undefined
    try{
      background=new BackgroundSnapshot();await background.initialize(this.dbName,this.key)
      if(generation!==this.backgroundGeneration){background.dispose();return}
      this.background=background;this.mirrored=undefined
      await this.flush()
    }catch{background?.dispose();this.background=undefined;this.mirrored=undefined}
  }
  stopBackgroundPersistence(){this.backgroundGeneration++;if(this.flushTimer)clearTimeout(this.flushTimer);if(this.idleFullTimer)clearTimeout(this.idleFullTimer);this.flushTimer=undefined;this.idleFullTimer=undefined;this.outputSinceIdle=false;this.forceFull=false;this.background?.dispose();this.background=undefined;this.mirrored=undefined;this.durable=undefined}
  flush(): Promise<void> {
    if(this.flushTimer)clearTimeout(this.flushTimer)
    this.flushTimer=undefined
    if(this.flushing)return this.pending
    this.flushing=true
    this.pending = Promise.resolve().then(async () => {
      try {
        if (!this.db || this.restoreFailure) return
        do {
          // Output arriving while IDB is busy remains in the synchronous journal.
          let revision = this.state.revision
          this.status.saving = true
          let transaction: {abort():void;done:Promise<void>} | undefined
          let backgroundAttempt:BackgroundSnapshot|undefined
          const backgroundGeneration=this.backgroundGeneration
          try {
            if(this.background){
              backgroundAttempt=this.background
              const forceFull=this.forceFull;this.forceFull=false
              revision=this.state.revision
              const externalVersion=this.externalVersion,previous=this.mirrored
              const state=forceFull||!previous||previous.externalVersion!==externalVersion?toRaw(this.state):undefined
              const entries=previous?this.tail.filter(entry=>entry.revision>previous.revision&&entry.revision<=revision):[]
              // The foreground has already protected this exact revision in
              // sessionStorage. Only complete native transaction confirmation
              // allows pruning its tail; mutations arriving meanwhile remain.
              await backgroundAttempt.write({revision,state,entries})
              if(backgroundGeneration!==this.backgroundGeneration||this.background!==backgroundAttempt)return
              this.mirrored={revision,externalVersion}
            }else{
            const tx = this.db.transaction(['snapshots', 'pointers'], 'readwrite')
            transaction=tx
            // Individual requests and transaction completion reject separately.
            // Observe completion immediately, even when an earlier request fails.
            void tx.done.catch(()=>{})
            const old = await tx.objectStore('pointers').get(this.key)
            // put() clones synchronously. Capture the revision immediately before
            // it, after the pointer read, without first duplicating the workspace.
            revision = this.state.revision
            const coldVersion=this.coldVersion
            const current = `${this.key}:${revision}`
            const previous=this.durable
            // Only an unbroken run of terminal-output commits can reuse a cold
            // base. Mixed UI/editor/checkpoint writes and ownership mismatch
            // keep the existing complete native snapshot path.
            const incoming=previous?this.tail.filter(entry=>entry.revision>previous.revision&&entry.revision<=revision):[]
            const delta:TerminalDelta|undefined=previous&&previous.coldVersion===coldVersion&&old?.current===previous.current&&incoming.length?terminalDelta(previous.base,previous.baseRevision,[...previous.entries,...incoming],revision):undefined
            let snapshotWrite: Promise<IDBValidKey>
            try { snapshotWrite = tx.objectStore('snapshots').put(delta||toRaw(this.state), current) }
            catch (error) {
              // A nested app-supplied Vue proxy can still require JSON cleanup.
              // Only a synchronous clone failure is safe to retry in this tx.
              if (!(error instanceof DOMException) || error.name !== 'DataCloneError') throw error
              snapshotWrite = tx.objectStore('snapshots').put(copy(delta||toRaw(this.state)), current)
            }
            await snapshotWrite
            const prior=current===old?.current?old?.previous:old?.current
            const priorBase=current===old?.current?old?.previousBase:old?.base
            const pointer={current,previous:prior,...(delta?{base:delta.base}:{}),...(priorBase?{previousBase:priorBase}:{})}
            await tx.objectStore('pointers').put(pointer, this.key)
            // Current and previous deltas must keep their complete bases. All
            // publication and cleanup stay inside this one native transaction.
            const retained=new Set([current,prior,delta?.base,priorBase])
            for(const stale of new Set([old?.current,old?.previous,old?.base,old?.previousBase]))if(stale&&!retained.has(stale))await tx.objectStore('snapshots').delete(stale)
            await tx.done
            this.durable={current,base:delta?.base||current,baseRevision:delta?.baseRevision??revision,revision,entries:delta?.entries||[],coldVersion}
            }
            this.tail = this.tail.filter(entry => entry.revision > revision)
            const remainingTail = JSON.stringify(this.tail)
            if (remainingTail.length * 2 > TAIL_LIMIT) {
              if (this.state.revision === revision) throw new Error('恢复日志超出同步保护预算，且快照未收敛')
              this.fail(new Error('最近输入超过同步保护预算，正在写入本地数据库；完成前请勿刷新，可导出当前现场'))
              // The snapshot predates edits that arrived while IndexedDB was busy.
              // Keep the synchronous-protection warning visible and snapshot the
              // latest state before clearing this oversized in-memory tail.
              continue
            }
            this.storage.setItem(`blora:tail:${this.key}`, remainingTail)
            this.status.protected = true
            this.status.message = '本地工作现场已保护'
          } catch (error) {
            if(error instanceof BackgroundSnapshotUnavailableError&&backgroundAttempt){
              // An explicit stop/logout owns the new generation. Never retry
              // its disposed writer or publish the old account's state.
              if(backgroundGeneration!==this.backgroundGeneration||this.background!==backgroundAttempt)return
              // Native IDB errors are ordinary errors, not this transport-only
              // signal. Keep the journal until a complete foreground transaction
              // confirms the latest state; no output is acknowledged here.
              this.stopBackgroundPersistence()
              continue
            }
            if(transaction){
              try{transaction.abort()}catch{/* Already completed or aborted. */}
              await transaction.done.catch(()=>{})
            }
            this.fail(error);return
          }
          finally { this.status.saving = false }
          if(this.state.revision===revision)break
        } while(true)
      } finally {
        this.flushing=false
        if(this.background&&this.status.protected&&this.outputSinceIdle){
          this.outputSinceIdle=false
          if(this.idleFullTimer)clearTimeout(this.idleFullTimer)
          // Deltas are already durable. A quiet terminal additionally converges
          // to a complete native Workspace record for inspection/recovery;
          // continuous output never waits for this optional compaction.
          this.idleFullTimer=setTimeout(()=>{this.idleFullTimer=undefined;this.forceFull=true;void this.flush()},500)
        }
      }
    })
    return this.pending
  }
  fail(error: unknown) { this.status.protected = false; this.status.message = error instanceof Error ? error.message : String(error) }
  export() { return JSON.stringify({ state: this.state, tail: this.tail, retained:this.retained }, null, 2) }
  async clear() {
    await this.awaitPendingWrites()
    this.stopBackgroundPersistence()
    this.storage.removeItem(`blora:tail:${this.key}`)
    if (!this.db) return
    const tx = this.db.transaction(['pointers','snapshots'], 'readwrite')
    for (const key of await tx.objectStore('snapshots').getAllKeys()) if (String(key).startsWith(this.key + ':')) await tx.objectStore('snapshots').delete(key)
    await tx.objectStore('pointers').delete(this.key)
    await tx.done
  }
  async clearAccount() {
    await this.awaitPendingWrites()
    this.stopBackgroundPersistence()
    const prefix = `${this.userId}:`
    for (let i = this.storage.length - 1; i >= 0; i--) {
      const key = this.storage.key(i)
      if (key?.startsWith(`blora:tail:${prefix}`) || key?.startsWith(`blora:workspace:${prefix}`)) this.storage.removeItem(key)
    }
    if (!this.db) throw new Error('本地数据库不可用，无法确认账号现场已清除')
    const tx = this.db.transaction(['pointers', 'snapshots'], 'readwrite')
    for (const store of ['pointers', 'snapshots']) {
      for (const key of await tx.objectStore(store).getAllKeys()) if (String(key).startsWith(prefix)) await tx.objectStore(store).delete(key)
    }
    await tx.done
    for (let i = localStorage.length - 1; i >= 0; i--) {
      const key = localStorage.key(i)
      if (key?.startsWith(`blora:workspace:${prefix}`)) localStorage.removeItem(key)
    }
  }
}

let identity: Promise<{deviceId:string;browserTabId:string;copiedFrom?:string}> | undefined
export function browserIdentity() { return identity ||= acquireBrowserIdentity() }
async function acquireBrowserIdentity() {
  const deviceId = localStorage.getItem('blora:device') || id('device')
  localStorage.setItem('blora:device', deviceId)
  let browserTabId = sessionStorage.getItem('blora:tab') || id('browserTab')
  const previousId = browserTabId
  // Holding a Web Lock detects copied sessionStorage in a duplicated browser tab.
  if (navigator.locks) {
    const acquire = (tabId:string) => new Promise<boolean>((resolve,reject) => {
      void navigator.locks.request(`blora-tab:${tabId}`, { ifAvailable: true }, async lock => {
        resolve(!!lock)
        if (lock) await new Promise<void>(() => {})
      }).catch(reject)
    })
    if (!await acquire(browserTabId)) {
      browserTabId = id('browserTab')
      await acquire(browserTabId)
    }
  }
  sessionStorage.setItem('blora:tab', browserTabId)
  return { deviceId, browserTabId, copiedFrom: browserTabId !== previousId ? previousId : undefined }
}
