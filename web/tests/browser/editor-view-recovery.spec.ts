import {test,expect} from '../helpers/management-fixture'
import {pressEditorKey} from '../helpers/editor-key'

test('shared Monaco windows restore independent cursor and scroll positions',async({page},info)=>{
  test.setTimeout(60000)
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  // This is a management-transport double, including the independent desktop
  // stream/poll. Keep HTTP interception owned by the context across refresh,
  // as in the navigation guards, and provide the actual summary schema.
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'editor-view-user',name:'视图恢复测试',admin:true},csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:{items:[]}})
  })
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus()
  const inputStart=Date.now()
  await page.keyboard.insertText(Array.from({length:200},(_,index)=>`独立视图行${index+1}`).join('\n'))
  await info.attach('editor-native-input-duration',{body:JSON.stringify({milliseconds:Date.now()-inputStart,lines:200}),contentType:'application/json'})
  const firstId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,firstView=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  await pressEditorKey(page,'Home')
  await page.locator('.app-window.focused [aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const secondId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,secondView=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  const first=page.locator(`[data-window-id="${firstId}"]`),second=page.locator(`[data-window-id="${secondId}"]`)
  await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await pressEditorKey(page,'End');await page.keyboard.press('ArrowLeft')
  await expect(first.locator('.view-line').first()).toHaveText('独立视图行1');await expect(second.locator('.view-lines')).toContainText('独立视图行200')
  const saved=()=>page.evaluate(async ids=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
    const newest=snapshots.filter(snapshot=>snapshot.views&&ids.every(id=>snapshot.views[id])).sort((a,b)=>b.revision-a.revision)[0]
    return newest?ids.map(id=>({draftId:newest.views[id].state.draftId,cursor:newest.views[id].state.editorView?.cursorState,scroll:newest.views[id].state.editorView?.viewState})):[]
  },[firstView,secondView])
  await expect.poll(async()=>(await saved()).length).toBe(2)
  const before=await saved();expect(before[0].draftId).toBe(before[1].draftId);expect(before[0].cursor).not.toEqual(before[1].cursor);expect(before[0].scroll).not.toEqual(before[1].scroll)
  await page.reload();await expect(first.locator('.view-line').first()).toHaveText('独立视图行1');await expect(second.locator('.view-lines')).toContainText('独立视图行200')
  await expect.poll(async()=>(await saved()).map(view=>view.cursor)).toEqual(before.map(view=>view.cursor))
  await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('定位插入')
  await expect(second.locator('.view-lines')).toContainText('独立视图行20定位插入0')
  await expect(first.locator('.view-line').first()).toHaveText('独立视图行1')
  expect(errors).toEqual([])
})
