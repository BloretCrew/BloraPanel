import { openDB, type IDBPDatabase } from 'idb'
import { reactive, toRaw } from 'vue'
import { id, type Workspace } from '../app-host/types'
import { applyEntry, copy, migrateWorkspace, trimDraftHistory, type JournalEntry, type Mutation } from './state'

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
  readonly key: string
  constructor(readonly userId: string, readonly deviceId: string, readonly browserTabId: string, readonly storage: Storage = sessionStorage, readonly dbName = 'blora-workspaces', readonly slot = 'main') {
    this.key = `${userId}:${deviceId}:${browserTabId}${slot === 'main' ? '' : ':' + slot}`
    this.state = reactive({ schemaVersion: 1, revision: 0, userId, deviceId, browserTabId, workspaceId: id('workspace'), windows: {}, views: {}, order: [], shortcuts: {}, drafts: {}, terminals: {},uploads:{}, preferences: {}, closedViews: [] })
  }
  async restore(persist=true) {
    try {
      this.db = await openDB(this.dbName, 1, { upgrade(db) { db.createObjectStore('snapshots'); db.createObjectStore('pointers') } })
      const pointer = await this.db.get('pointers', this.key)
      if (pointer) {
        const snapshot = await this.db.get('snapshots', pointer.current)
        this.retained.push(snapshot)
        const restored = migrateWorkspace(snapshot)
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
    const entry = { revision: this.state.revision + 1, mutations: copy(mutations) }
    applyEntry(this.state, entry)
    this.tail.push(entry)
    if(this.restoreFailure){this.fail(new Error('原恢复记录无法读取，已禁止覆盖；请导出保留记录和当前内容'));return}
    try {
      const tail = JSON.stringify(this.tail)
      if (tail.length * 2 > TAIL_LIMIT) throw new Error('最近输入超过同步保护预算，正在写入本地数据库；完成前请勿刷新，可导出当前现场')
      this.storage.setItem(`blora:tail:${this.key}`, tail)
      this.status.protected = true
      this.status.message = '本地工作现场已保护'
    } catch (error) { this.fail(error) }
    void this.flush()
  }
  awaitPendingWrites():Promise<void>{return this.pending}
  flush(): Promise<void> {
    if(this.flushing)return this.pending
    this.flushing=true
    this.pending = Promise.resolve().then(async () => {
      try {
        if (!this.db || this.restoreFailure) return
        do {
          // Output arriving while IDB is busy is in the synchronous journal;
          // coalesce its next snapshot instead of copying identical revisions.
          // Workspace values are JSON-shaped. Native cloning avoids encoding
          // and parsing every large editor/terminal string for each snapshot.
          // Retain the JSON path for nested proxies supplied by an app.
          let snapshot:Workspace
          try { snapshot = structuredClone(toRaw(this.state)) }
          catch { snapshot = copy(toRaw(this.state)) }
          const revision = snapshot.revision
          this.status.saving = true
          let transaction: {abort():void;done:Promise<void>} | undefined
          try {
            const tx = this.db.transaction(['snapshots', 'pointers'], 'readwrite')
            transaction=tx
            // Individual requests and transaction completion reject separately.
            // Observe completion immediately, even when an earlier request fails.
            void tx.done.catch(()=>{})
            const old = await tx.objectStore('pointers').get(this.key)
            const current = `${this.key}:${revision}`
            await tx.objectStore('snapshots').put(snapshot, current)
            await tx.objectStore('pointers').put({ current, previous: old?.current }, this.key)
            if (old?.previous && old.previous !== current) await tx.objectStore('snapshots').delete(old.previous)
            await tx.done
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
            if(transaction){
              try{transaction.abort()}catch{/* Already completed or aborted. */}
              await transaction.done.catch(()=>{})
            }
            this.fail(error);return
          }
          finally { this.status.saving = false }
          if(this.state.revision===revision)break
        } while(true)
      } finally {this.flushing=false}
    })
    return this.pending
  }
  fail(error: unknown) { this.status.protected = false; this.status.message = error instanceof Error ? error.message : String(error) }
  export() { return JSON.stringify({ state: this.state, tail: this.tail, retained:this.retained }, null, 2) }
  async clear() {
    await this.pending
    this.storage.removeItem(`blora:tail:${this.key}`)
    if (!this.db) return
    const tx = this.db.transaction(['pointers','snapshots'], 'readwrite')
    for (const key of await tx.objectStore('snapshots').getAllKeys()) if (String(key).startsWith(this.key + ':')) await tx.objectStore('snapshots').delete(key)
    await tx.objectStore('pointers').delete(this.key)
    await tx.done
  }
  async clearAccount() {
    await this.pending
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
