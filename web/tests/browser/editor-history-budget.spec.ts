import {test,expect,type Page} from '@playwright/test'

const budget=1024
type StoredDraft={path?:string;text:string;base:string;cursor:number;history:{group?:string}[];historyBytes:number;historyTrimmed?:boolean}

async function persistedDraft(page:Page){
  return page.evaluate(async({userId,path})=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces');request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    try{
      const key=`${userId}:${localStorage.getItem('blora:device')}:${sessionStorage.getItem('blora:tab')}`
      const pointer=await new Promise<{current:string}|undefined>((resolve,reject)=>{const request=db.transaction('pointers').objectStore('pointers').get(key);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
      if(!pointer)return undefined
      const snapshot=await new Promise<{drafts?:Record<string,StoredDraft>}|undefined>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').get(pointer.current);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
      const draft=Object.values(snapshot?.drafts||{}).find(item=>item.path===path)
      if(!draft)return undefined
      return {textLength:draft.text.length,first:draft.text[0],last:draft.text.at(-1),allSame:draft.text===draft.text[0]?.repeat(draft.text.length),baseLength:draft.base.length,baseLast:draft.base.at(-1),cursor:draft.cursor,historyLength:draft.history.length,historyGroups:draft.history.map(step=>step.group),historyBytes:draft.historyBytes,historyTrimmed:draft.historyTrimmed}
    }finally{db.close()}
  },{userId:'history-user',path:'budget.txt'})
}

test('undo history keeps the same byte budget before and after refresh',async({page})=>{
  test.setTimeout(120000)
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'history-user',name:'历史预算测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'history-instance',nodeId:'history-node',nodeName:'历史节点',name:'历史实例',state:'STOPPED',config:{},revision:1}]}
    else if(path.endsWith('/instances/history-instance'))json={instance:{instanceId:'history-instance',nodeId:'history-node',nodeName:'历史节点',name:'历史实例',state:'STOPPED',config:{},revision:1}}
    else if(path.endsWith('/files'))json={items:[{name:'budget.txt',path:'budget.txt',isDir:false,size:4,version:'history-version'}],total:1,version:'directory-version'}
    else if(path.endsWith('/files/access'))json={instanceId:'history-instance',nodeId:'history-node',nodeName:'历史节点',nodeState:'ONLINE',canRead:true,canWrite:true}
    else if(path.endsWith('/files/content'))json={text:'seed',version:'history-version',encoding:'UTF-8',newline:'LF',maxBytes:512}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'历史实例',exact:true}).click();await page.getByRole('button',{name:'文件',exact:true}).click();await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click();await page.getByRole('button',{name:'budget.txt',exact:true}).click()
  const editor=page.getByRole('textbox',{name:'文件正文编辑器',exact:true});await editor.focus()
  for(const letter of 'ABCDEFGHIJKL'){
    await page.keyboard.press('Control+A');await page.keyboard.type(letter)
    await expect.poll(async()=>({letter:(await persistedDraft(page))?.last,protected:(await page.locator('.editor-status').innerText()).includes('本地工作现场已保护')}),{timeout:15000}).toEqual({letter,protected:true})
  }
  const current=await persistedDraft(page)
  expect(current).toMatchObject({textLength:1,first:'L',last:'L',allSame:true,historyTrimmed:true})
  expect(current!.historyBytes).toBeLessThanOrEqual(budget);expect(current!.cursor).toBe(current!.historyLength)
  expect(current!.historyLength).toBeGreaterThanOrEqual(2)
  await expect(page.locator('.editor-status')).toContainText('撤销历史');await expect(page.locator('.editor-status')).toContainText('旧记录已按预算清理')
  await expect(page.getByRole('button',{name:'↶ 撤销',exact:true})).toBeEnabled();await page.getByRole('button',{name:'↶ 撤销',exact:true}).click()
  await expect.poll(async()=>({last:(await persistedDraft(page))?.last,cursor:(await persistedDraft(page))?.cursor}), {timeout:10000}).toEqual({last:'K',cursor:current!.cursor-1})
  await page.getByRole('button',{name:'↶ 撤销',exact:true}).click()
  await expect.poll(async()=>(await persistedDraft(page))?.last).toBe('J')
  await page.reload();await expect.poll(async()=>(await persistedDraft(page))?.last).toBe('J')
  await expect(page.getByRole('button',{name:'↷ 重做',exact:true})).toBeEnabled();await page.getByRole('button',{name:'↷ 重做',exact:true}).click();await expect.poll(async()=>(await persistedDraft(page))?.last).toBe('K')
  await page.getByRole('button',{name:'↷ 重做',exact:true}).click();await expect.poll(async()=>(await persistedDraft(page))?.last).toBe('L')
  expect(errors).toEqual([])
})
