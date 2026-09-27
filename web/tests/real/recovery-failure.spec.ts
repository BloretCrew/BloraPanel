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
  async function changeVersion(version:number){await page.evaluate(async version=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const tx=db.transaction(['pointers','snapshots'],'readwrite'),pointers=tx.objectStore('pointers'),snapshots=tx.objectStore('snapshots')
    const request=pointers.openCursor();request.onsuccess=()=>{const cursor=request.result;if(!cursor)return;const value=snapshots.get(cursor.value.current);value.onsuccess=()=>{const snapshot=value.result;snapshot.schemaVersion=version;if(version===0){snapshot.tabId=snapshot.browserTabId;delete snapshot.browserTabId}snapshots.put(snapshot,cursor.value.current)};cursor.continue()}
    await new Promise<void>((resolve,reject)=>{tx.oncomplete=()=>resolve();tx.onerror=()=>reject(tx.error);tx.onabort=()=>reject(tx.error)});db.close()
  },version)}
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  await changeVersion(0);await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('旧版本保留正文')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  await changeVersion(999);await page.reload();await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  const result=await exported(page);expect(result.retained.some((snapshot:any)=>snapshot.schemaVersion===999&&Object.values(snapshot.drafts).some((draft:any)=>draft.text==='旧版本保留正文'))).toBeTruthy()
  await page.reload();await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  expect((await exported(page)).retained.some((snapshot:any)=>snapshot.schemaVersion===999)).toBeTruthy()
})

test('aborted pointer transaction keeps old snapshot and replays latest synchronous draft',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await realLogin(page,fixture.admin);await page.locator('[data-app="blora.editor"]').click();await edit(page,'事务前正文')
  const database=()=>page.evaluate(async()=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const tx=db.transaction(['pointers','snapshots']),read=(store:string)=>new Promise<any[]>((resolve,reject)=>{const request=tx.objectStore(store).getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const [pointers,snapshots]=await Promise.all([read('pointers'),read('snapshots')]);db.close();return{pointers,snapshots}
  })
  await expect.poll(async()=>JSON.stringify((await database()).snapshots)).toContain('事务前正文')
  const before=await database()
  await page.evaluate(()=>{
    const put=IDBObjectStore.prototype.put
    ;(window as any).abortedPointers=0
    IDBObjectStore.prototype.put=function(...args:Parameters<IDBObjectStore['put']>){
      const request=put.apply(this,args)
      if(this.name==='pointers'){(window as any).abortedPointers++;this.transaction.abort()}
      return request
    }
  })
  await edit(page,'，事务中断后的最新输入')
  await expect(page.getByRole('button',{name:'保护异常，导出工作区',exact:true})).toBeVisible()
  expect(await page.evaluate(()=>(window as any).abortedPointers)).toBeGreaterThan(0)
  expect(await database()).toEqual(before)
  const saved=await exported(page);expect(Object.values(saved.state.drafts).some((draft:any)=>draft.text==='事务前正文，事务中断后的最新输入')).toBeTruthy()
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('事务前正文，事务中断后的最新输入')
  await expect(page.getByRole('button',{name:'现场已保护，导出工作区',exact:true})).toBeVisible()
  expect(errors).toEqual([])
})
