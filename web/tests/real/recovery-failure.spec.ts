import {readFileSync} from 'node:fs'
import {test,expect,type Page} from '@playwright/test'
import {realLogin} from './login'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
async function edit(page:Page,text:string){await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+End');await page.keyboard.insertText(text)}
async function exported(page:Page){const pending=page.waitForEvent('download');await page.getByRole('button',{name:'保护异常，导出工作区',exact:true}).click();const download=await pending;return JSON.parse(readFileSync((await download.path())!,'utf8'))}

test('browser quota failures retain latest draft for export and recover after storage returns',async({page})=>{
  await realLogin(page,fixture.admin);await page.locator('[data-app="blora.editor"]').click();await edit(page,'配额前正文')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  await page.evaluate(()=>{
    const set=Storage.prototype.setItem,put=IDBObjectStore.prototype.put
    ;(window as any).restoreWrites=()=>{Storage.prototype.setItem=set;IDBObjectStore.prototype.put=put}
    Storage.prototype.setItem=function(key,value){if(key.startsWith('blora:tail:'))throw new DOMException('injected quota exhausted','QuotaExceededError');return set.call(this,key,value)}
    IDBObjectStore.prototype.put=function(...args:Parameters<IDBObjectStore['put']>){if(this.name==='snapshots')throw new DOMException('injected quota exhausted','QuotaExceededError');return put.apply(this,args)}
  })
  await edit(page,'，最后中文输入');await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  const saved=await exported(page);expect(Object.values(saved.state.drafts).some((draft:any)=>draft.text==='配额前正文，最后中文输入')).toBeTruthy()
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('最后中文输入')
  await page.evaluate(()=>(window as any).restoreWrites());await edit(page,'，恢复后')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible();await page.reload()
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('配额前正文，最后中文输入，恢复后')
})

test('old snapshot migrates and unsupported snapshot remains exportable',async({page})=>{
  await realLogin(page,fixture.admin);await page.locator('[data-app="blora.editor"]').click();await edit(page,'旧版本保留正文')
  async function changeVersion(version:number){
    // The protection indicator also covers the synchronous journal. Confirm a
    // real draft snapshot exists before injecting a historical/corrupt schema.
    await expect.poll(()=>page.evaluate(async()=>{
      const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
      const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
      db.close();return snapshots.some(snapshot=>Object.values(snapshot.drafts||{}).some((draft:any)=>draft.text==='旧版本保留正文'))
    })).toBe(true)
    // End the product document and its owned Worker before changing the DB.
    // /healthz is a real same-origin read-only JSON endpoint with no app script.
    const response=await page.goto('/healthz');expect(response?.status()).toBe(200)
    const changed=await page.evaluate(async version=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const tx=db.transaction(['pointers','snapshots'],'readwrite'),pointers=tx.objectStore('pointers'),snapshots=tx.objectStore('snapshots')
    let changed=0
    const request=pointers.openCursor();request.onsuccess=()=>{const cursor=request.result;if(!cursor)return;const value=snapshots.get(cursor.value.current);value.onsuccess=()=>{const snapshot=value.result;snapshot.schemaVersion=version;if(version===0){snapshot.tabId=snapshot.browserTabId;delete snapshot.browserTabId}snapshots.put(snapshot,cursor.value.current);changed++};cursor.continue()}
    await new Promise<void>((resolve,reject)=>{tx.oncomplete=()=>resolve();tx.onerror=()=>reject(tx.error);tx.onabort=()=>reject(tx.error)});db.close();return changed
    },version)
    expect(changed).toBeGreaterThan(0)
  }
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  await changeVersion(0);await page.goto('/');await expect(page.locator('.monaco-editor .view-lines')).toContainText('旧版本保留正文')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  await changeVersion(999);await page.goto('/');await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  const result=await exported(page);expect(result.retained.some((snapshot:any)=>snapshot.schemaVersion===999&&Object.values(snapshot.drafts).some((draft:any)=>draft.text==='旧版本保留正文'))).toBeTruthy()
  await page.reload();await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  expect((await exported(page)).retained.some((snapshot:any)=>snapshot.schemaVersion===999)).toBeTruthy()
})

test('aborted pointer transaction keeps old snapshot and replays latest synchronous draft',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  // Production commits use a dedicated Worker realm. Abort its real native IDB
  // transaction; a page-only prototype patch cannot intercept these writes.
  await page.route('**/assets/snapshot.worker-*.js',async route=>{
    const response=await route.fetch()
    const preamble=`let pointerAbortEnabled=false,pointerAbortCount=0;
const pointerAbortNativePut=IDBObjectStore.prototype.put;
IDBObjectStore.prototype.put=function(...args){const request=pointerAbortNativePut.apply(this,args);if(pointerAbortEnabled&&this.name==='pointers'){pointerAbortCount++;this.transaction.abort()}return request};
self.addEventListener('message',event=>{const message=event.data;if(!message?.__bloraPointerAbortTest)return;event.stopImmediatePropagation();if(message.action==='enable')pointerAbortEnabled=true;self.postMessage({__bloraPointerAbortTest:true,nonce:message.nonce,count:pointerAbortCount,enabled:pointerAbortEnabled})});\n`
    await route.fulfill({response,body:preamble+await response.text()})
  })
  await page.addInitScript(()=>{
    const NativeWorker=window.Worker
    window.Worker=new Proxy(NativeWorker,{construct(target,args){
      const worker=Reflect.construct(target,args,target) as Worker
      if(/\/snapshot\.worker-[^/]+\.js(?:\?|$)/.test(String(args[0])))(window as any).pointerAbortWorker=worker
      return worker
    }})
  })
  const pointerFault=(action:'enable'|'status')=>page.evaluate(action=>new Promise<{count:number;enabled:boolean}>((resolve,reject)=>{
    const worker=(window as any).pointerAbortWorker as Worker|undefined
    if(!worker){reject(Error('Production snapshot Worker was not initialized'));return}
    const nonce=crypto.randomUUID()
    const timer=setTimeout(()=>{worker.removeEventListener('message',listener);reject(Error('Snapshot Worker fault control was not acknowledged'))},5000)
    function listener(event:MessageEvent){if(!event.data?.__bloraPointerAbortTest||event.data.nonce!==nonce)return;clearTimeout(timer);worker!.removeEventListener('message',listener);resolve({count:event.data.count,enabled:event.data.enabled})}
    worker.addEventListener('message',listener);worker.postMessage({__bloraPointerAbortTest:true,nonce,action})
  }),action)
  await realLogin(page,fixture.admin);await page.locator('[data-app="blora.editor"]').click();await edit(page,'事务前正文')
  const database=()=>page.evaluate(async()=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const tx=db.transaction(['pointers','snapshots']),read=(store:string)=>new Promise<any[]>((resolve,reject)=>{const request=tx.objectStore(store).getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const [pointers,snapshots]=await Promise.all([read('pointers'),read('snapshots')]);db.close();return{pointers,snapshots}
  })
  await expect.poll(async()=>JSON.stringify((await database()).snapshots)).toContain('事务前正文')
  const before=await database()
  expect(await pointerFault('enable')).toEqual({count:0,enabled:true})
  await edit(page,'，事务中断后的最新输入')
  await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  expect((await pointerFault('status')).count).toBeGreaterThan(0)
  expect(await database()).toEqual(before)
  const saved=await exported(page);expect(Object.values(saved.state.drafts).some((draft:any)=>draft.text==='事务前正文，事务中断后的最新输入')).toBeTruthy()
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('事务前正文，事务中断后的最新输入')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  expect(errors).toEqual([])
})
