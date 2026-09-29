import 'fake-indexeddb/auto'
import { describe,it,expect } from 'vitest'
import {reactive,toRaw} from 'vue'
import {RecoveryService} from '../src/recovery/service'
import {copy,inverseEdits,json,editText,migrateWorkspace,applyEntry,EDITOR_HISTORY_BUDGET_BYTES,editorHistoryBudgetBytes} from '../src/recovery/state'
import type {Draft,Workspace} from '../src/app-host/types'
import {openDB} from 'idb'

class MemoryStorage implements Storage {
  values=new Map<string,string>();fail=false
  get length(){return this.values.size}
  clear(){this.values.clear()}
  getItem(key:string){return this.values.get(key)??null}
  key(index:number){return [...this.values.keys()][index]??null}
  removeItem(key:string){this.values.delete(key)}
  setItem(key:string,value:string){if(this.fail)throw new DOMException('quota exhausted','QuotaExceededError');this.values.set(key,value)}
}
async function setup(){const storage=new MemoryStorage(),dbName='test-'+crypto.randomUUID(),service=new RecoveryService('alice','device','tab',storage,dbName);await service.restore();return {storage,service,dbName}}
const draft:Draft={draftId:'d',base:'alpha\n你好',text:'alpha\n你好',history:[],cursor:0,savedText:'alpha\n你好'}
describe('recovery value copies',()=>{
  it('deep-copies terminal payloads, falls back for reactive proxies, and keeps JSON cleanup at the app boundary',()=>{
    const payload={events:[{sequence:1,data:'terminal-output-'.repeat(2048)}]}
    const clone=copy(payload)
    payload.events[0]!.data='changed'
    expect(clone.events[0]!.data).toBe('terminal-output-'.repeat(2048))
    expect(clone.events).not.toBe(payload.events)
    const proxy=reactive({nested:{label:'restored'}})
    expect(copy(proxy)).toEqual({nested:{label:'restored'}})
    expect(json({kept:'value',omitted:undefined})).toEqual({kept:'value'})
  })
})
describe('synchronous tail and transactional snapshots',()=>{
  it('keeps the snapshot revision and contents atomic when edits arrive just after put',async()=>{
    const {service,dbName}=await setup()
    const nativePut=IDBObjectStore.prototype.put;let injected=false
    IDBObjectStore.prototype.put=function(this:IDBObjectStore,value:unknown,key?:IDBValidKey){
      const request=nativePut.call(this,value,key)
      if(!injected&&this.name==='snapshots'){
        injected=true
        service.commit([{kind:'set',path:['preferences','value'],value:'after put'}])
      }
      return request
    }
    try{
      service.commit([{kind:'set',path:['preferences','value'],value:'before put'}])
      await service.awaitPendingWrites()
    }finally{IDBObjectStore.prototype.put=nativePut}
    const db=await openDB(dbName,1),pointer=await db.get('pointers',service.key)
    const current=await db.get('snapshots',pointer.current),previous=await db.get('snapshots',pointer.previous)
    expect(injected).toBe(true)
    expect(current.revision).toBe(service.state.revision)
    expect(current.preferences.value).toBe('after put')
    expect(previous.revision).toBe(current.revision-1)
    expect(previous.preferences.value).toBe('before put')
    db.close()
    const restored=new RecoveryService('alice','device','tab',new MemoryStorage(),dbName)
    await restored.restore();expect(restored.state.preferences.value).toBe('after put')
  })
  it('retains the nested reactive app value fallback when IndexedDB cannot clone it',async()=>{
    const {service,dbName}=await setup()
    service.state.preferences.nested=reactive({text:'中文 draft'})
    service.commit([{kind:'set',path:['preferences','other'],value:'saved'}])
    await service.awaitPendingWrites()
    expect(service.status.protected).toBe(true)
    const restored=new RecoveryService('alice','device','tab',new MemoryStorage(),dbName)
    await restored.restore()
    expect(restored.state.preferences).toEqual({nested:{text:'中文 draft'},other:'saved'})
  })
  it('coalesces a burst while preserving the final synchronous tail and awaited snapshot',async()=>{
    const {storage,service,dbName}=await setup()
    service.commit([{kind:'set',path:['preferences','sample'],value:0}])
    const pending=service.awaitPendingWrites()
    for(let sample=1;sample<=100;sample++)service.commit([{kind:'set',path:['preferences','sample'],value:sample}])
    expect(service.awaitPendingWrites()).toBe(pending)
    expect(storage.getItem(`blora:tail:${service.key}`)).toContain('100')
    await pending
    // Restore using only the durable snapshot, with no shared in-memory state
    // or synchronous journal to conceal a lost final write.
    const restored=new RecoveryService('alice','device','tab',new MemoryStorage(),dbName)
    await restored.restore();expect(restored.state.preferences.sample).toBe(100)
    expect(restored.state.revision).toBe(service.state.revision)
  })
  it('protects the final edit before an asynchronous flush and preserves undo/redo position',async()=>{
    const {storage,service,dbName}=await setup()
    service.commit([{kind:'set',path:['drafts','d'],value:json(draft)}]);await service.flush()
    const forward=[{offset:0,length:5,text:'βeta'},{offset:8,length:0,text:'，世界'}]
    service.commit([{kind:'edit',draftId:'d',forward,reverse:inverseEdits(service.state.drafts.d!.text,forward)}])
    expect(storage.getItem(`blora:tail:${service.key}`)).toContain('世界')
    service.commit([{kind:'history',draftId:'d',cursor:0}])
    const restored=new RecoveryService('alice','device','tab',storage,dbName);await restored.restore()
    expect(restored.state.drafts.d!.text).toBe('alpha\n你好')
    restored.commit([{kind:'history',draftId:'d',cursor:1}])
    expect(restored.state.drafts.d!.text).toBe('βeta\n你好，世界')
    await restored.flush()
  })
  it('includes a new edit arriving during snapshot completion in the same awaited write',async()=>{
    const {storage,service,dbName}=await setup()
    const set=storage.setItem.bind(storage);let injected=false
    storage.setItem=(key,value)=>{
      set(key,value)
      if(value==='[]'&&!injected){injected=true;service.commit([{kind:'set',path:['preferences','late'],value:'latest'}])}
    }
    service.commit([{kind:'set',path:['preferences','first'],value:'first'}])
    await service.awaitPendingWrites();expect(injected).toBe(true)
    const restored=new RecoveryService('alice','device','tab',new MemoryStorage(),dbName)
    await restored.restore();expect(restored.state.preferences).toEqual({first:'first',late:'latest'})
  })
  it('never writes an oversized tail when edits outrun an older IndexedDB snapshot',async()=>{
    const {storage,service,dbName}=await setup()
    const set=storage.setItem.bind(storage);let largestTailBytes=0,oversizedWrites=0,injected=false
    storage.setItem=(key,value)=>{
      if(key===`blora:tail:${service.key}`){const bytes=value.length*2;largestTailBytes=Math.max(largestTailBytes,bytes);if(bytes>256*1024)oversizedWrites++}
      set(key,value)
    }
    const nativeGet=IDBObjectStore.prototype.get
    IDBObjectStore.prototype.get=function(this:IDBObjectStore,key:IDBValidKey){
      const request=nativeGet.call(this,key)
      if(!injected&&this.name==='pointers'&&key===service.key){
        request.addEventListener('success',()=>{
          injected=true
          for(let i=0;i<100;i++)service.commit([{kind:'set',path:['preferences','payload'],value:`${'x'.repeat(3000)}-${i}`}])
        })
      }
      return request
    }
    try{
      service.commit([{kind:'set',path:['preferences','initial'],value:'snapshot in flight'}])
      await service.awaitPendingWrites()
    }finally{IDBObjectStore.prototype.get=nativeGet}
    expect(injected).toBe(true)
    expect(oversizedWrites).toBe(0)
    expect(largestTailBytes).toBeLessThanOrEqual(256*1024)
    expect(storage.getItem(`blora:tail:${service.key}`)).toBe('[]')
    expect(service.status.protected).toBe(true)
    const restored=new RecoveryService('alice','device','tab',storage,dbName)
    await restored.restore()
    expect(restored.state.preferences.payload).toBe(`${'x'.repeat(3000)}-99`)
    expect(restored.state.revision).toBe(service.state.revision)
  })
  it('appends bounded terminal output deltas and rejects a mismatched checkpoint sequence',async()=>{
    const {storage,service,dbName}=await setup()
    const checkpoint={sessionId:'session',viewTabId:'view',sequence:40,baseSequence:38,outputJournal:[{sequence:39,data:'b2xk'},{sequence:40,data:'c2F2ZWQ='}],cols:80,rows:24,screen:'prompt',scroll:0}
    service.commit([{kind:'set',path:['terminals','session:view'],value:json(checkpoint)}]);await service.flush()
    service.commit([{kind:'terminal-output',checkpointKey:'session:view',baseSequence:38,previousSequence:40,sequence:42,events:[{sequence:41,data:'b25l'},{sequence:42,data:'dHdv'}],scroll:2}])
    service.commit([{kind:'terminal-output',checkpointKey:'session:view',baseSequence:38,previousSequence:42,sequence:43,events:[{sequence:43,data:'dGhyZWU='}],scroll:3}])
    const saved=service.state.terminals['session:view']!
    expect(saved).toMatchObject({baseSequence:38,sequence:43,scroll:3,outputJournal:[{sequence:39,data:'b2xk'},{sequence:40,data:'c2F2ZWQ='},{sequence:41,data:'b25l'},{sequence:42,data:'dHdv'},{sequence:43,data:'dGhyZWU='}]})
    // IndexedDB must be able to clone the raw snapshot directly. Array spreads
    // of reactive entries otherwise embed proxies and force a full JSON copy.
    expect(structuredClone(toRaw(service.state)).terminals['session:view']).toEqual(toRaw(saved))
    const revision=service.state.revision
    expect(()=>service.commit([{kind:'terminal-output',checkpointKey:'session:view',baseSequence:39,previousSequence:43,sequence:44,events:[{sequence:44,data:'bWFsaWNpb3Vz'}],scroll:4}])).toThrow('检查点不连续')
    expect(service.state.revision).toBe(revision)
    expect(()=>service.commit([{kind:'terminal-output',checkpointKey:'session:view',baseSequence:38,previousSequence:43,sequence:44,events:[{sequence:44,data:'A'.repeat(70_000)}],scroll:4}])).toThrow('保护预算')
    expect(service.state.revision).toBe(revision)
    await service.flush()
    const restored=new RecoveryService('alice','device','tab',storage,dbName);await restored.restore()
    expect(restored.state.terminals['session:view']).toMatchObject({baseSequence:38,sequence:43,scroll:3,outputJournal:saved.outputJournal})
  })
  it('persists ordered terminal resize deltas and rejects invalid sizes before mutation',async()=>{
    const {storage,service,dbName}=await setup()
    service.commit([{kind:'set',path:['terminals','s:v'],value:json({sessionId:'s',sequence:0,cols:100,rows:28,screen:'',scroll:0})}])
    service.commit([{kind:'terminal-output',checkpointKey:'s:v',baseSequence:0,previousSequence:0,sequence:3,events:[{sequence:1,data:'G1szMQ=='},{sequence:2,kind:'resize',cols:80,rows:20},{sequence:3,data:'bVI='}],scroll:0}])
    const expected=structuredClone(toRaw(service.state)).terminals['s:v']
    expect(expected?.outputJournal?.[1]).toEqual({sequence:2,kind:'resize',cols:80,rows:20})
    const revision=service.state.revision
    expect(()=>service.commit([{kind:'terminal-output',checkpointKey:'s:v',baseSequence:0,previousSequence:3,sequence:4,events:[{sequence:4,kind:'resize',cols:0,rows:20}],scroll:0}])).toThrow('尺寸日志无效')
    expect(service.state.revision).toBe(revision)
    expect(service.state.terminals['s:v']).toEqual(expected)
    await service.flush()
    const restored=new RecoveryService('alice','device','tab',storage,dbName);await restored.restore()
    expect(restored.state.terminals['s:v']).toEqual(expected)
  })
  it('keeps current text and reports actual quota failure without claiming protection',async()=>{
    const {storage,service}=await setup();service.commit([{kind:'set',path:['drafts','d'],value:json(draft)}]);await service.flush();storage.fail=true
    service.commit([{kind:'edit',draftId:'d',forward:[{offset:0,length:0,text:'last'}],reverse:[{offset:0,length:4,text:''}]}])
    expect(service.status.protected).toBe(false);expect(service.state.drafts.d!.text).toBe('lastalpha\n你好');expect(service.export()).toContain('lastalpha');await service.flush();expect(service.status.protected).toBe(false)
  })
  it('never restores a different account workspace',async()=>{
    const {storage,service,dbName}=await setup();service.commit([{kind:'set',path:['drafts','d'],value:json(draft)}]);await service.flush()
    const bob=new RecoveryService('bob','device','tab',storage,dbName);await bob.restore();expect(bob.state.drafts).toEqual({})
  })
  it('keeps named workspace snapshots separate from the original browser-tab workspace',async()=>{
    const {storage,service,dbName}=await setup();service.commit([{kind:'set',path:['drafts','d'],value:json(draft)}]);await service.flush()
    const separate=new RecoveryService('alice','device','tab',storage,dbName,'second');await separate.restore();expect(separate.state.drafts).toEqual({})
    separate.commit([{kind:'set',path:['preferences','title'],value:'second'}]);await separate.flush()
    const original=new RecoveryService('alice','device','tab',storage,dbName);await original.restore();expect(original.state.drafts.d!.text).toBe(draft.text);expect(original.state.preferences.title).toBeUndefined()
  })
  it('retains incompatible database snapshots and refuses to overwrite them with a fresh empty workspace',async()=>{
    const {storage,service,dbName}=await setup();const db=await openDB(dbName,1);const pointer=await db.get('pointers',service.key)
    const old={...service.state,schemaVersion:77,drafts:{d:draft}};await db.put('snapshots',json(old),pointer.current)
    const restored=new RecoveryService('alice','device','tab',storage,dbName);await restored.restore();expect(restored.status.protected).toBe(false)
    restored.commit([{kind:'set',path:['preferences','theme'],value:'dark'}]);await restored.flush()
    expect((await db.get('snapshots',pointer.current)).schemaVersion).toBe(77);expect(restored.export()).toContain('alpha');expect(restored.status.protected).toBe(false);db.close()
  })
  it('migrates old browser-tab identity and retains unsupported input unchanged',async()=>{
    const {service}=await setup();const old={...service.state,schemaVersion:0,browserTabId:'',tabId:'legacy-tab'}
    expect(migrateWorkspace(old).browserTabId).toBe('legacy-tab');const unknown={...old,schemaVersion:99};expect(()=>migrateWorkspace(unknown)).toThrow('无法迁移');expect(unknown.schemaVersion).toBe(99)
  })
  it('inverts multiple simultaneous edits and Unicode without losing offsets',()=>{
    const source='ABC\n中文\n😀xyz',forward=[{offset:0,length:2,text:'π'},{offset:4,length:2,text:'汉字测试'},{offset:9,length:3,text:'last'}]
    expect(editText(editText(source,forward),inverseEdits(source,forward))).toBe(source)
  })
  it('rebases complete oldest undo groups within one durable history budget',()=>{
    expect(editorHistoryBudgetBytes({maxBytes:4096})).toBe(8192)
    const state:Workspace={schemaVersion:1,revision:0,userId:'alice',deviceId:'device',browserTabId:'tab',workspaceId:'workspace',windows:{},views:{},order:[],shortcuts:{},drafts:{d:{...draft,history:[],cursor:0,historyBytes:0,historyEpoch:0}},terminals:{},uploads:{},preferences:{},closedViews:[]}
    const first='a'.repeat(4*1024*1024+128),second='b'.repeat(4*1024*1024+128)
    const firstForward=[{offset:state.drafts.d!.text.length,length:0,text:first}]
    applyEntry(state,{revision:1,mutations:[{kind:'edit',draftId:'d',forward:firstForward,reverse:inverseEdits(state.drafts.d!.text,firstForward),group:'first'}]})
    const secondForward=[{offset:state.drafts.d!.text.length,length:0,text:second}]
    applyEntry(state,{revision:2,mutations:[{kind:'edit',draftId:'d',forward:secondForward,reverse:inverseEdits(state.drafts.d!.text,secondForward),group:'second'}]})
    const compacted=state.drafts.d!
    expect(compacted.historyBytes).toBeLessThanOrEqual(EDITOR_HISTORY_BUDGET_BYTES)
    expect(compacted.historyTrimmed).toBe(true);expect(compacted.historyEpoch).toBe(1);expect(compacted.cursor).toBe(1);expect(compacted.history).toHaveLength(1)
    expect(compacted.base).toBe('alpha\n你好'+first);expect(compacted.text).toBe('alpha\n你好'+first+second)
    applyEntry(state,{revision:3,mutations:[{kind:'seal-history-group',draftId:'d',group:'second'}]})
    applyEntry(state,{revision:4,mutations:[{kind:'history',draftId:'d',cursor:0}]})
    expect(compacted.text).toBe('alpha\n你好'+first)
    applyEntry(state,{revision:5,mutations:[{kind:'history',draftId:'d',cursor:1}]})
    expect(compacted.text).toBe('alpha\n你好'+first+second)
    const restored=migrateWorkspace(copy(state)).drafts.d!
    expect(restored).toMatchObject({base:compacted.base,text:compacted.text,cursor:1,historyBytes:compacted.historyBytes,historyTrimmed:true})
  })
  it('keeps an over-budget live edit group atomic instead of undoing only its suffix',()=>{
    const state:Workspace={schemaVersion:1,revision:0,userId:'alice',deviceId:'device',browserTabId:'tab',workspaceId:'workspace',windows:{},views:{},order:[],shortcuts:{},drafts:{d:{...draft,history:[],cursor:0,historyBytes:0,historyEpoch:0,maxBytes:512}},terminals:{},uploads:{},preferences:{},closedViews:[]}
    for(let index=0;index<32;index++){
      const text=state.drafts.d!.text,forward=[{offset:text.length,length:0,text:'x'}]
      applyEntry(state,{revision:index+1,mutations:[{kind:'edit',draftId:'d',forward,reverse:[{offset:text.length,length:1,text:''}],group:'large-paste'}]})
    }
    const active=state.drafts.d!
    expect(active.historyTrimmed).toBe(true);expect(active.historyDroppedGroup).toBe('large-paste')
    expect(active.text).toBe('alpha\n你好'+'x'.repeat(32));expect(active.base).toBe(active.text)
    expect(active.history).toHaveLength(0);expect(active.cursor).toBe(0);expect(active.historyBytes).toBe(0)
    applyEntry(state,{revision:33,mutations:[{kind:'seal-history-group',draftId:'d',group:'large-paste'}]})
    expect(active.historyDroppedGroup).toBeUndefined();expect(active.text).toBe(active.base)
    applyEntry(state,{revision:34,mutations:[{kind:'history',draftId:'d',cursor:0}]})
    expect(active.text).toBe('alpha\n你好'+'x'.repeat(32))
  })
})
