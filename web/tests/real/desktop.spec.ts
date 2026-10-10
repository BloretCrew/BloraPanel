import {realLogin} from './login'
import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
type Credentials={url:string;admin:{name:string;password:string};member:{name:string;password:string};instanceIds:string[];nodeIds:string[]}
const fixture:Credentials=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
async function login(page:Page,role:'admin'|'member'){await realLogin(page,fixture[role])}
async function csrf(page:Page){return (await (await page.request.get('/api/v1/session')).json()).csrfToken as string}

test('real shortcut keeps the original target after rename deletion and same-name replacement',async({page})=>{
  await login(page,'admin')
  const base=(await (await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  const originalName=`快捷入口目标-${randomUUID().slice(0,8)}`,renamed=`快捷入口改名-${randomUUID().slice(0,8)}`
  const token=await csrf(page)
  const headers=()=>({'X-CSRF-Token':token, 'Idempotency-Key':randomUUID()})
  const createdResponse=await page.request.post('/api/v1/instances',{headers:headers(),data:{nodeId:base.nodeId,name:originalName,config:base.config}})
  expect(createdResponse.status()).toBe(201)
  const temporary=(await createdResponse.json()).instance
  try{
    await page.locator('[data-app="blora.instances"]').click()
    const card=page.locator('.instance-card').filter({has:page.getByRole('button',{name:originalName,exact:true})})
    await expect(card).toHaveCount(1);await card.getByTitle('添加到桌面',{exact:true}).click();await page.getByRole('button',{name:'关闭窗口',exact:true}).click()
    const shortcut=page.locator('[data-shortcut]');await expect(shortcut).toContainText(originalName)
    const renamedResponse=await page.request.patch(`/api/v1/instances/${temporary.instanceId}`,{headers:headers(),data:{revision:temporary.configRevision,name:renamed}})
    expect(renamedResponse.status()).toBe(200)
    await expect.poll(async()=>shortcut.innerText(),{timeout:10000}).toContain(renamed)
    await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'刷新',exact:true}).click()
    const renamedCard=page.locator('.instance-card').filter({has:page.getByRole('button',{name:renamed,exact:true})})
    await expect(renamedCard).toHaveCount(1);await renamedCard.getByRole('button',{name:'删除',exact:true}).click()
    const deleteDialog=page.getByRole('dialog');await expect(deleteDialog).toContainText('删除会保留墓碑');await deleteDialog.getByRole('button',{name:'确认提交',exact:true}).click()
    await expect(renamedCard).toHaveCount(0);await page.getByRole('button',{name:'关闭窗口',exact:true}).click()
    const replacementResponse=await page.request.post('/api/v1/instances',{headers:headers(),data:{nodeId:base.nodeId,name:renamed,config:base.config}});expect(replacementResponse.status()).toBe(201)
    const replacement=(await replacementResponse.json()).instance
    try{
      await page.reload();await expect(page.locator('[data-shortcut]')).toHaveCount(1)
      const stale=page.locator('[data-shortcut]');await expect(stale).toContainText('不可访问');expect(await stale.innerText()).not.toContain(replacement.instanceId)
      await stale.click();await expect(page.locator('.empty-state')).toContainText('资源可能已删除或权限发生变化')
      expect((await page.request.get(`/api/v1/instances/${temporary.instanceId}`)).status()).toBe(404)
      expect((await page.request.get(`/api/v1/instances/${replacement.instanceId}`)).status()).toBe(200)
      await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click()
      await stale.click({button:'right'});await page.getByRole('menuitem',{name:'移除桌面入口',exact:true}).click();await expect(stale).toHaveCount(0)
    }finally{
      const cleanup=await page.request.delete(`/api/v1/instances/${replacement.instanceId}`,{headers:headers()});expect(cleanup.status()).toBe(200)
    }
  }finally{
    // A failed assertion can leave the temporary target behind; clean it by
    // its immutable ID when it is still visible to this administrator.
    if((await page.request.get(`/api/v1/instances/${temporary.instanceId}`)).status()===200){
      const cleanup=await page.request.delete(`/api/v1/instances/${temporary.instanceId}`,{headers:headers()});expect([200,404,409]).toContain(cleanup.status())
    }
  }
})

test('real shortcut becomes inaccessible after instance read permission is revoked',async({page,request})=>{
  await login(page,'member')
  const target=(await (await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  await page.locator('[data-app="blora.instances"]').click()
  const card=page.locator('.instance-card').filter({has:page.getByRole('button',{name:target.name,exact:true})})
  await expect(card).toHaveCount(1);await card.getByTitle('添加到桌面',{exact:true}).click();await page.getByRole('button',{name:'关闭窗口',exact:true}).click()
  const shortcut=page.locator('[data-shortcut]');await expect(shortcut).toContainText(target.name)
  const adminLogin=await request.post('/api/v1/login',{data:{name:fixture.admin.name,password:fixture.admin.password}});expect(adminLogin.status()).toBe(200)
  const adminSession=await request.get('/api/v1/session');const adminCSRF=(await adminSession.json()).csrfToken as string
  const users=await request.get('/api/v1/users');expect(users.status()).toBe(200);const memberID=((await users.json()).items as {userId:string;name:string}[]).find(user=>user.name===fixture.member.name)!.userId
  const grant={userId:memberID,resource:{kind:'instance',id:target.instanceId,nodeId:target.nodeId},action:'instance.read'}
  let revoked=false
  try{
    const revoke=await request.delete('/api/v1/grants',{headers:{'X-CSRF-Token':adminCSRF,'Idempotency-Key':randomUUID()},data:grant});expect(revoke.status()).toBe(200);revoked=true
    await page.reload();await expect(shortcut).toContainText('不可访问',{timeout:10000});expect((await page.request.get(`/api/v1/instances/${target.instanceId}`)).status()).toBe(403)
    await shortcut.click();await expect(page.locator('.empty-state')).toContainText('资源可能已删除或权限发生变化');await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click()
    await shortcut.click({button:'right'});await page.getByRole('menuitem',{name:'移除桌面入口',exact:true}).click();await expect(shortcut).toHaveCount(0)
  }finally{
    if(revoked){const restore=await request.post('/api/v1/grants',{headers:{'X-CSRF-Token':adminCSRF,'Idempotency-Key':randomUUID()},data:grant});expect(restore.status()).toBe(200)}
  }
})

test('real instance shortcut rename removal and explicit new window never mutate the resource',async({page})=>{
  await login(page,'admin');await page.locator('[data-app="blora.instances"]').click()
  const original=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  const existingTasks=new Set((await(await page.request.get('/api/v1/tasks')).json()).items.map((task:{taskId:string})=>task.taskId))
  const card=page.locator('.instance-card').filter({has:page.getByRole('button',{name:original.name,exact:true})})
  await card.getByTitle('添加到桌面',{exact:true}).click();await page.getByRole('button',{name:'关闭窗口',exact:true}).click()
  const shortcut=page.locator('[data-shortcut]');await shortcut.click()
  const dedicated=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!
  await expect(page.locator(`[data-window-id="${dedicated}"]`)).toHaveAttribute('data-window-mode','resource')
  await page.getByRole('button',{name:'最小化窗口',exact:true}).click();await shortcut.click()
  await expect(page.locator('.app-window')).toHaveCount(1);expect(await page.locator('.app-window.focused').getAttribute('data-window-id')).toBe(dedicated)
  await page.getByRole('button',{name:'最小化窗口',exact:true}).click();await shortcut.click({button:'right'});await page.getByRole('menuitem',{name:'在新窗口打开',exact:true}).click()
  await expect(page.locator('.app-window')).toHaveCount(2)
  await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click()
  await shortcut.click({button:'right'});await page.getByRole('menuitem',{name:'重命名入口',exact:true}).click()
  await page.getByRole('textbox',{name:'桌面入口名称'}).fill('只改本地入口');await page.getByRole('button',{name:'保存入口名称',exact:true}).click();await page.reload()
  await expect(shortcut).toContainText('只改本地入口')
  const current=(await(await page.request.get(`/api/v1/instances/${original.instanceId}`)).json()).instance
  expect(current.name).toBe(original.name);expect(current.state).toBe('STOPPED');expect(current.runId||'').toBe(original.runId||'')
  await shortcut.click({button:'right'});await page.getByRole('menuitem',{name:'移除桌面入口',exact:true}).click();await expect(shortcut).toHaveCount(0)
  expect((await page.request.get(`/api/v1/instances/${original.instanceId}`)).status()).toBe(200)
  const tasks=(await(await page.request.get('/api/v1/tasks')).json()).items
  expect(tasks.filter((task:any)=>!existingTasks.has(task.taskId)&&task.resource.id===original.instanceId&&['instance.start','instance.restart','instance.delete'].includes(task.action))).toHaveLength(0)
})
for(const background of [false,true])test(`latest IME paste and moved ${background?'background':'active'} editor survive immediate reload`,async({page,context})=>{
  await login(page,'admin');await page.locator('[data-app="blora.editor"]').click()
  const id=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,win=()=>page.locator(`[data-window-id="${id}"]`)
  await win().getByRole('textbox',{name:'文件正文编辑器'}).focus()
  const input=await context.newCDPSession(page)
  await input.send('Input.imeSetComposition',{text:'中文组合',selectionStart:4,selectionEnd:4});await input.send('Input.insertText',{text:'中文组合'})
  await context.grantPermissions(['clipboard-read','clipboard-write'],{origin:new URL(page.url()).origin})
  await page.evaluate(()=>navigator.clipboard.writeText('，真实粘贴'))
  await page.keyboard.press('Control+v');await page.keyboard.insertText('，连续编辑');await page.keyboard.insertText('，最后一步')
  await page.keyboard.press('Control+z')
  const title=(await win().locator('.window-titlebar').boundingBox())!
  await page.mouse.move(title.x+220,title.y+15);await page.mouse.down();await page.mouse.move(title.x+340,title.y+80,{steps:5});await page.mouse.up()
  const geometry=await win().evaluate(element=>(element as HTMLElement).style.transform)
  if(background){await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'任务中心',exact:true}).click()}
  await page.reload()
  await expect(win().locator('.view-lines')).toContainText('中文组合，真实粘贴，连续编辑');await expect(win().locator('.view-lines')).not.toContainText('最后一步')
  expect(await win().evaluate(element=>(element as HTMLElement).style.transform)).toBe(geometry)
  if(background)await expect(win()).not.toHaveClass(/focused/)
  else await expect(win()).toHaveClass(/focused/)
  await win().getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+y')
  await expect(win().locator('.view-lines')).toContainText('最后一步')
  await page.keyboard.insertText('，刷新前最后输入');await page.reload()
  await expect(win().locator('.view-lines')).toContainText('最后一步，刷新前最后输入')
})
test('accepted restart with dropped browser response restores original request without another run',async({page})=>{
  await login(page,'admin');await page.locator('[data-app="blora.instances"]').click()
  const instance=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  await page.getByRole('button',{name:instance.name,exact:true}).click()
  const win=page.locator('.app-window.focused'),endpoint=`/api/v1/instances/${instance.instanceId}/actions`
  await win.getByRole('button',{name:'启动',exact:true}).click();await win.getByRole('button',{name:'确认提交',exact:true}).click()
  await expect.poll(async()=>(await(await page.request.get(`/api/v1/instances/${instance.instanceId}`)).json()).instance.state,{timeout:30000}).toBe('RUNNING')
  let original:any,key='';const keys:string[]=[]
  await page.route(`**${endpoint}`,async route=>{
    keys.push(route.request().headers()['idempotency-key']!)
    const response=await route.fetch()
    if(!original){expect(response.status()).toBe(202);original=(await response.json()).task;key=keys[0]!;await route.abort('connectionreset')}
    else await route.fulfill({response})
  })
  try{
    await win.getByRole('button',{name:'重启',exact:true}).click();await win.getByRole('button',{name:'确认提交',exact:true}).click()
    await expect.poll(()=>original?.taskId).toBeTruthy();await expect(win.locator('.error')).toBeVisible()
    await page.reload();await expect(win.getByRole('button',{name:'确认提交',exact:true})).toBeVisible()
    await win.getByRole('button',{name:'确认提交',exact:true}).click()
    await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${original.taskId}`)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')
    const tasks=(await(await page.request.get('/api/v1/tasks')).json()).items.filter((task:any)=>task.action==='instance.restart'&&task.resource.id===instance.instanceId)
    expect(tasks).toHaveLength(1);expect(tasks[0].taskId).toBe(original.taskId);expect(tasks[0].requestId).toBe(key)
    expect(keys.every(value=>value===key)).toBe(true)
    const current=(await(await page.request.get(`/api/v1/instances/${instance.instanceId}`)).json()).instance
    expect(current.runId).toBe(tasks[0].result.runId)
  }finally{
    await page.unroute(`**${endpoint}`)
    const csrf=(await(await page.request.get('/api/v1/session')).json()).csrfToken
    await page.request.post(endpoint,{headers:{'X-CSRF-Token':csrf,'Idempotency-Key':crypto.randomUUID()},data:{action:'stop'}})
    await expect.poll(async()=>(await(await page.request.get(`/api/v1/instances/${instance.instanceId}`)).json()).instance.state,{timeout:30000}).toBe('STOPPED')
  }
})
test('shared Monaco windows restore independent cursor and scroll positions',async({page})=>{
  await login(page,'admin');await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText(Array.from({length:200},(_,index)=>`独立视图行${index+1}`).join('\n'))
  const firstId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,firstView=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  await page.keyboard.press('Control+Home')
  await page.locator('.app-window.focused [aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const secondId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,secondView=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  const first=page.locator(`[data-window-id="${firstId}"]`),second=page.locator(`[data-window-id="${secondId}"]`)
  await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+End');await page.keyboard.press('ArrowLeft')
  await expect(first.locator('.view-line').first()).toHaveText('独立视图行1');await expect(second.locator('.view-lines')).toContainText('独立视图行200')
  const saved=()=>page.evaluate(async ids=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
    const newest=snapshots.filter(snapshot=>ids.every(id=>snapshot.views[id])).sort((a,b)=>b.revision-a.revision)[0]
    return newest?ids.map(id=>({draftId:newest.views[id].state.draftId,cursor:newest.views[id].state.editorView?.cursorState,scroll:newest.views[id].state.editorView?.viewState})):[]
  },[firstView,secondView])
  await expect.poll(async()=>(await saved()).length).toBe(2)
  const before=await saved();expect(before[0].draftId).toBe(before[1].draftId);expect(before[0].cursor).not.toEqual(before[1].cursor);expect(before[0].scroll).not.toEqual(before[1].scroll)
  await page.reload();await expect(first.locator('.view-line').first()).toHaveText('独立视图行1');await expect(second.locator('.view-lines')).toContainText('独立视图行200')
  await expect.poll(async()=>(await saved()).map(view=>view.cursor)).toEqual(before.map(view=>view.cursor))
  await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('定位插入')
  await expect(second.locator('.view-lines')).toContainText('独立视图行20定位插入0')
  await expect(first.locator('.view-line').first()).toHaveText('独立视图行1')
})
test('two real cross-node instance centers share runtime state and serialize concurrent controls',async({page})=>{
  await page.setViewportSize({width:1920,height:1100});await login(page,'admin')
  const {items}=await(await page.request.get('/api/v1/instances')).json(),instances=fixture.instanceIds.map(id=>items.find((item:any)=>item.instanceId===id))
  expect(new Set(instances.map((instance:any)=>instance.nodeId)).size).toBe(2)
  const windows:string[]=[],views:string[]=[]
  for(let index=0;index<2;index++){
    if(index===0)await page.locator('[data-app="blora.instances"]').click()
    else{await page.locator('.taskbar-app').filter({hasText:'实例中心'}).click({button:'right'});await page.getByRole('menuitem',{name:'新建窗口',exact:true}).click()}
    const win=page.locator('.app-window.focused'),windowId=(await win.getAttribute('data-window-id'))!;windows.push(windowId)
    const heading=(await win.locator('.window-titlebar').boundingBox())!
    await page.mouse.move(heading.x+220,heading.y+15);await page.mouse.down();await page.mouse.move(index?1180:240,200,{steps:8});await page.mouse.up()
    await win.getByRole('button',{name:instances[1].name,exact:true}).click()
    await win.getByRole('button',{name:'新建标签',exact:true}).click()
    await win.getByRole('button',{name:instances[0].name,exact:true}).click()
    views.push((await win.locator('.view-tab.selected').getAttribute('data-view-tab'))!)
    expect(await win.locator('[data-view-tab]').count()).toBeGreaterThanOrEqual(4)
  }
  const first=page.locator(`[data-window-id="${windows[0]}"]`),second=page.locator(`[data-window-id="${windows[1]}"]`)
  await first.locator('.window-titlebar').click();await first.getByRole('button',{name:'文件',exact:true}).click();await expect(first.locator('.resource-nav .selected:visible')).toHaveText('文件');await expect(second.locator('.resource-nav .selected:visible')).toHaveText('概览')
  await first.getByRole('button',{name:'概览',exact:true}).click()
  await first.getByRole('button',{name:'启动',exact:true}).click({force:true});await second.getByRole('button',{name:'启动',exact:true}).click({force:true})
  const responses:any[]=[];page.on('response',async response=>{if(response.request().method()==='POST'&&response.url().endsWith(`/instances/${instances[0].instanceId}/actions`))responses.push({status:response.status(),body:await response.json()})})
  // Both already-confirmed target dialogs submit in the same browser task.
  await page.evaluate(ids=>{for(const id of ids){const button=[...document.querySelectorAll<HTMLButtonElement>(`[data-window-id="${id}"] button`)].find(button=>button.textContent==='确认提交');if(!button)throw new Error('missing fixed-target confirmation');button.click()}},windows)
  await expect.poll(()=>responses.length,{timeout:30000}).toBe(2)
  const tasks=responses.filter(response=>response.status===202).map(response=>response.body.task)
  expect(tasks.length).toBeGreaterThan(0)
  for(const response of responses)expect([202,409]).toContain(response.status)
  for(const task of tasks)await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${task.taskId}`)).json()).task.state,{timeout:30000}).toMatch(/SUCCEEDED|FAILED/)
  await expect(first.locator('.resource-heading:visible')).toContainText('RUNNING',{timeout:30000});await expect(second.locator('.resource-heading:visible')).toContainText('RUNNING',{timeout:30000})
  const running=(await(await page.request.get(`/api/v1/instances/${instances[0].instanceId}`)).json()).instance
  const completed=await Promise.all(tasks.map(async task=>(await(await page.request.get(`/api/v1/tasks/${task.taskId}`)).json()).task))
  expect(completed.filter(task=>task.state==='SUCCEEDED')).toHaveLength(1)
  for(const task of completed){expect(task.result.runId).toBe(running.runId);if(task.state==='FAILED')expect(task.error).toBe('existing running unit has not exited')}
  await expect(first.locator('.details-grid:visible')).toContainText(running.runId);await expect(second.locator('.details-grid:visible')).toContainText(running.runId)
  await first.getByRole('button',{name:'文件',exact:true}).click();await page.reload()
  await expect(page.locator(`[data-window-id="${windows[0]}"] .resource-nav .selected:visible`)).toHaveText('文件')
  await expect(page.locator(`[data-window-id="${windows[1]}"] .resource-nav .selected:visible`)).toHaveText('概览')
  for(const view of views)await expect(page.locator(`[data-view-tab="${view}"]`)).toHaveCount(1)
  const csrf=(await(await page.request.get('/api/v1/session')).json()).csrfToken
  const stop=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/actions`,{headers:{'X-CSRF-Token':csrf,'Idempotency-Key':crypto.randomUUID()},data:{action:'stop'}});expect(stop.status()).toBe(202)
  await expect.poll(async()=>(await(await page.request.get(`/api/v1/instances/${instances[0].instanceId}`)).json()).instance.state,{timeout:30000}).toBe('STOPPED')
})
test('real pointer tear-out and merge retain unsaved editor identity and undo on immediate reload',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await login(page,'admin');await page.locator('[data-app="blora.editor"]').click()
  const originalWindow=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,first=(await page.locator('[data-view-tab]').getAttribute('data-view-tab'))!
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('原标签独立正文')
  await page.getByRole('button',{name:'新建标签',exact:true}).click()
  const moved=(await page.locator('.view-tab.selected').getAttribute('data-view-tab'))!
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('拖动未保存中文');await page.keyboard.insertText('末次输入');await page.keyboard.press('Control+z')
  const tab=page.locator(`[data-view-tab="${moved}"]`),box=(await tab.boundingBox())!
  await page.mouse.move(box.x+30,box.y+12);await page.mouse.down();await page.mouse.move(1100,760,{steps:12});await page.mouse.up();await page.reload()
  await expect(page.locator('.app-window')).toHaveCount(2);await expect(page.locator(`[data-view-tab="${moved}"]`)).toHaveCount(1)
  const detached=page.locator('.app-window').filter({has:page.locator(`[data-view-tab="${moved}"]`)})
  expect(await detached.getAttribute('data-window-id')).not.toBe(originalWindow)
  await expect(detached.locator('.monaco-editor .view-lines')).toContainText('拖动未保存中文')
  await detached.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+y');await expect(detached.locator('.monaco-editor .view-lines')).toContainText('末次输入')
  const source=(await page.locator(`[data-view-tab="${moved}"]`).boundingBox())!,destination=(await page.locator(`[data-tab-strip="${originalWindow}"]`).boundingBox())!
  await page.mouse.move(source.x+30,source.y+12);await page.mouse.down();await page.mouse.move(destination.x+10,destination.y+12,{steps:12});await page.mouse.up();await page.reload()
  await expect(page.locator('.app-window')).toHaveCount(1)
  expect(await page.locator(`[data-tab-strip="${originalWindow}"] [data-view-tab]`).evaluateAll(elements=>elements.map(element=>element.getAttribute('data-view-tab')))).toEqual([moved,first])
  await expect(page.locator('.monaco-editor:visible .view-lines')).toContainText('末次输入')
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+z');await expect(page.locator('.monaco-editor:visible .view-lines')).not.toContainText('末次输入')
  await page.locator(`[data-view-tab="${first}"]`).click();await expect(page.locator('.monaco-editor:visible .view-lines')).toContainText('原标签独立正文');expect(errors).toEqual([])
})
test('real TLS login, cross-node dedicated windows and confirmed process control survive reload',async({page})=>{
  await login(page,'admin');await page.locator('[data-app="blora.instances"]').click()
  const response=await page.request.get('/api/v1/instances');expect(response.ok()).toBeTruthy();const {items}=await response.json()
  const seeded=fixture.instanceIds.map(instanceId=>items.find((item:{instanceId:string})=>item.instanceId===instanceId))
  expect(seeded.every(Boolean)).toBe(true)
  for(const instance of seeded)await expect(page.locator('.instance-card').filter({has:page.getByRole('button',{name:instance.name,exact:true})})).toHaveCount(1)
  expect(new Set(items.map((i:{nodeId:string})=>i.nodeId)).size).toBe(2)
  const selected=seeded[0]!
  const card=page.locator('.instance-card').filter({has:page.getByRole('button',{name:selected.name,exact:true})})
  await card.getByTitle('添加到桌面',{exact:true}).click();await card.getByRole('button',{name:'管理实例'}).click()
  const centerView=await page.locator('[data-active-view]').getAttribute('data-active-view')
  await page.getByRole('button',{name:'最小化窗口'}).click();await page.locator('[data-shortcut]').click();await expect(page.locator('.app-window')).toHaveCount(2)
  const dedicated=page.locator('.app-window.focused');expect(await dedicated.locator('[data-active-view]').getAttribute('data-active-view')).not.toBe(centerView)
  await dedicated.getByRole('button',{name:'启动',exact:true}).click();await dedicated.getByRole('button',{name:'确认提交',exact:true}).click()
  await expect.poll(async()=>{const r=await page.request.get(`/api/v1/instances/${selected.instanceId}`);return (await r.json()).instance.state},{timeout:30000}).toBe('RUNNING')
  const taskResponse=await page.request.get('/api/v1/tasks');const tasks=(await taskResponse.json()).items;const start=tasks.find((t:{action:string;resource:{id:string}})=>t.action==='instance.start'&&t.resource.id===selected.instanceId);expect(start).toBeTruthy()
  await page.reload();await expect(page.locator('.app-window')).toHaveCount(2);await expect(page.locator('.app-window.focused .resource-heading')).toContainText('RUNNING')
  const after=(await (await page.request.get('/api/v1/tasks')).json()).items;expect(after.filter((t:{requestId:string})=>t.requestId===start.requestId)).toHaveLength(1)
  await page.locator('.app-window.focused').getByRole('button',{name:'停止',exact:true}).click();await page.locator('.app-window.focused').getByRole('button',{name:'确认提交',exact:true}).click()
  await expect.poll(async()=>{const r=await page.request.get(`/api/v1/instances/${selected.instanceId}`);return (await r.json()).instance.state},{timeout:30000}).toBe('STOPPED')
  await page.screenshot({path:'real-test-results/real-tls-desktop.png'})
})
test('real limited account sees two authorized instances and cannot create or control the second',async({page})=>{
  await login(page,'member');await page.locator('[data-app="blora.instances"]').click();await expect(page.locator('.instance-card')).toHaveCount(2)
  await expect(page.getByRole('button',{name:'创建实例'})).toBeDisabled()
  const {items}=await (await page.request.get('/api/v1/instances')).json();const selected=items.find((i:{instanceId:string})=>i.instanceId===fixture.instanceIds[1])
  await page.getByRole('button',{name:selected.name,exact:true}).click();await page.getByRole('button',{name:'启动',exact:true}).click();await page.getByRole('button',{name:'确认提交',exact:true}).click()
  await expect(page.locator('.application .error')).toContainText(/权限|授权/)
  const state=(await (await page.request.get(`/api/v1/instances/${selected.instanceId}`)).json()).instance.state;expect(state).toBe('STOPPED')
})
test('real instance creation is node-scoped and rechecks permission at confirmation',async({page,request})=>{
  await login(page,'member')
  let adminLogin=await request.post('/api/v1/login',{data:{name:fixture.admin.name,password:fixture.admin.password}})
  if(adminLogin.status()===429){test.setTimeout(test.info().timeout+65000);test.info().annotations.push({type:'rate-limit',description:'真实登录返回429，等待61秒后只重试一次；未更改服务端限制。'});await page.waitForTimeout(61000);adminLogin=await request.post('/api/v1/login',{data:{name:fixture.admin.name,password:fixture.admin.password}})}
  expect(adminLogin.status()).toBe(200)
  const adminCSRF=(await (await request.get('/api/v1/session')).json()).csrfToken as string
  const users=await request.get('/api/v1/users');expect(users.status()).toBe(200);const memberID=((await users.json()).items as {userId:string;name:string}[]).find(user=>user.name===fixture.member.name)!.userId
  const nodeId=fixture.nodeIds[1]!,grants=['instance.create','host.manage'].map(action=>({userId:memberID,resource:{kind:'node',id:nodeId},action}))
  const mutate=async(method:'post'|'delete',grant:any)=>request[method]('/api/v1/grants',{headers:{'X-CSRF-Token':adminCSRF,'Idempotency-Key':randomUUID()},data:grant})
  for(const grant of grants){const response=await mutate('post',grant);expect(response.status()).toBe(200)}
  let createdID=''
  try{
    await page.locator('[data-app="blora.instances"]').click();await expect(page.getByRole('button',{name:'创建实例',exact:true})).toBeEnabled();await page.getByRole('button',{name:'创建实例',exact:true}).click()
    const dialog=page.getByRole('dialog',{name:'创建实例'});await dialog.getByRole('textbox',{name:'名称',exact:true}).fill(`权限创建-${randomUUID().slice(0,8)}`);const nodeSelect=dialog.getByRole('combobox',{name:'目标节点',exact:true});await nodeSelect.click();await expect(page.getByRole('option')).toHaveCount(2);await expect(page.locator('.styled-select-popover [role="option"]').evaluateAll(options=>options.map(option=>option.getAttribute('data-value')))).resolves.toEqual(expect.arrayContaining([nodeId]));await page.locator(`.styled-select-popover [role="option"][data-value="${nodeId}"]`).click();await selectStyledOption(page,dialog.getByRole('combobox',{name:'配置模板',exact:true}),'native-linux');await dialog.getByRole('textbox',{name:'可执行程序',exact:true}).fill('/bin/true');await dialog.getByRole('button',{name:'核对创建实例',exact:true}).click()
    const createGrant=grants[0]!;const revoked=await mutate('delete',createGrant);expect(revoked.status()).toBe(200);await dialog.getByRole('button',{name:'确认创建实例',exact:true}).click();await expect(dialog).toContainText(/权限|禁止|拒绝/);await dialog.getByRole('button',{name:'返回修改',exact:true}).click()
    await mutate('post',createGrant);await dialog.getByRole('button',{name:'核对创建实例',exact:true}).click();const response=page.waitForResponse(r=>r.url().endsWith('/api/v1/instances')&&r.request().method()==='POST');await dialog.getByRole('button',{name:'确认创建实例',exact:true}).click();const accepted=await response;expect(accepted.status()).toBe(201);createdID=(await accepted.json()).instance.instanceId;await expect(page.locator('.resource-heading')).toBeVisible()
    const created=(await (await request.get(`/api/v1/instances/${createdID}`)).json()).instance;expect(created.nodeId).toBe(nodeId);expect(created.state).toBe('STOPPED')
  }finally{
    if(createdID){const removed=await request.delete(`/api/v1/instances/${createdID}`,{headers:{'X-CSRF-Token':adminCSRF,'Idempotency-Key':randomUUID()}});expect([200,404]).toContain(removed.status())}
    for(const grant of grants){const response=await mutate('delete',grant);expect([200,400,404]).toContain(response.status())}
  }
})
test('real HTTPS Monaco draft survives reload and explicit logout removes the previous account workspace',async({page})=>{
  await login(page,'admin');await page.locator('[data-app="blora.editor"]').click();await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('验收专用未保存草稿')
  await page.reload();await expect(page.locator('.monaco-editor .view-lines')).toContainText('验收专用未保存草稿')
  await page.getByRole('button',{name:'账号菜单'}).click();page.once('dialog',dialog=>dialog.accept());await page.getByRole('menuitem',{name:'退出账号',exact:true}).click();await expect(page.getByRole('button',{name:'进入工作区 →'})).toBeVisible()
  await login(page,'member');await expect(page.locator('.app-window')).toHaveCount(0);await page.locator('[data-app="blora.editor"]').click();await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('验收专用未保存草稿')
})
