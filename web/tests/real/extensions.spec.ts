import {readFileSync,readdirSync} from 'node:fs'
import {dirname,join,resolve} from 'node:path'
import {randomUUID} from 'node:crypto'
import {test, expect, type Page} from '@playwright/test'
import {realLogin} from './login'
import {clickPaintedExtensionButton} from '../extension-paint'

const fixture = JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!, 'utf8')) as {admin:{name:string;password:string};member:{name:string;password:string};extensionId?:string;nodeIds:string[]}
const extensionId = fixture.extensionId || 'fixture.reference'

test('accepted extension cancellation survives a lost browser receipt and reload',async({page})=>{
  test.setTimeout(90000)
  await realLogin(page,fixture.admin)
  const pkg=JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/reference.blora-extension.json',import.meta.url),'utf8'))
  const nodes=(await(await page.request.get('/api/v1/nodes')).json()).items
  const node=nodes.find((item:any)=>item.nodeId===fixture.nodeIds[0])
  const config=join(dirname(resolve(process.env.BLORA_E2E_CREDENTIALS!)),node.name+'.json')
  const matches=readdirSync('/proc').filter(id=>/^\d+$/.test(id)).filter(id=>{try{const args=readFileSync(`/proc/${id}/cmdline`,'utf8').split('\0');return args[0]?.endsWith('blora-daemon')&&args[1]==='--config'&&args[2]===config}catch{return false}})
  expect(matches).toHaveLength(1)
  const pid=Number(matches[0]),birth=()=>readFileSync(`/proc/${pid}/stat`,'utf8').split(') ').at(-1)!.split(' ')[19],originalBirth=birth()
  let paused=false,taskId='',cancelKey='',cancelCalls=0
  try{
    expect((await page.request.post('/api/v1/extensions/install-package',{headers:await headers(page),data:pkg})).status()).toBe(201)
    await openExtensions(page)
    await page.locator('[data-app-id="example.reference"]').getByRole('button',{name:'打开',exact:true}).click()
    const frames=page.locator('iframe[title="扩展 example.reference"]')
    await frames.first().contentFrame().getByLabel('节点标识').fill(node.nodeId)
    await clickPaintedExtensionButton(page,frames.first().contentFrame().getByRole('button',{name:'打开节点工作窗口'}))
    await expect(frames).toHaveCount(2)
    const target=frames.nth(1).contentFrame()
    expect(birth()).toBe(originalBirth);process.kill(pid,'SIGSTOP');paused=true
    const accepted=page.waitForResponse(r=>r.url().endsWith('/extensions/example.reference/tasks')&&r.request().method()==='POST')
    await clickPaintedExtensionButton(page,target.getByRole('button',{name:'提交节点任务'}))
    taskId=(await(await accepted).json()).task.taskId
    await page.route(`**/extensions/example.reference/tasks/${taskId}/cancel`,async route=>{
      cancelCalls++;cancelKey=route.request().headers()['idempotency-key']!
      const response=await route.fetch()
      expect(response.status()).toBe(202)
      expect((await response.json()).task.cancellationRequestId).toBe(cancelKey)
      await route.abort('connectionreset')
    })
    await target.getByRole('button',{name:'取消任务',exact:true}).click()
    await expect.poll(()=>cancelCalls).toBe(1)
    await expect.poll(async()=> (await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.cancellationRequested).toBe(true)
    await page.reload()
    await expect(target.getByRole('button',{name:'取消任务',exact:true})).toBeDisabled()
    expect(cancelCalls).toBe(1)
    expect(birth()).toBe(originalBirth);process.kill(pid,'SIGCONT');paused=false
    await expect.poll(async()=> (await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:45000}).toBe('CANCELLED')
    await expect(target.getByRole('status')).toContainText('任务已取消')
    const receipt=(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task
    expect(receipt.cancellationRequestId).toBe(cancelKey)
    expect(cancelCalls).toBe(1)
  }finally{
    test.setTimeout(test.info().timeout+15000)
    if(paused){expect(birth()).toBe(originalBirth);process.kill(pid,'SIGCONT')}
    await page.request.delete('/api/v1/extensions/example.reference?cleanup=true',{headers:await headers(page)})
  }
})

test('independent extension notification retains source and survives refresh without replay',async({page})=>{
  test.setTimeout(90000)
  let stage='login'
  await realLogin(page,fixture.admin)
  const pkg=JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/reference.blora-extension.json',import.meta.url),'utf8'))
  let sent=0
  let releaseData=()=>{}
  const dataGate=new Promise<void>(resolve=>{releaseData=resolve})
  await page.route('**/extensions/example.reference/data',async route=>{await dataGate;await route.continue()})
  page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/example.reference/notifications'))sent++})
  try{
    stage='install package'
    expect((await page.request.post('/api/v1/extensions/install-package',{headers:await headers(page),data:pkg})).status()).toBe(201)
    stage='open market'
    await openExtensions(page)
    stage='open extension'
    const extensionRow=page.locator('[data-app-id="example.reference"]')
    await expect(extensionRow).toBeVisible({timeout:5000})
    const openButton=extensionRow.getByRole('button',{name:'打开',exact:true})
    await expect(openButton).toBeEnabled({timeout:5000})
    await openButton.click({timeout:5000})
    stage='extension opened'
    const extension=page.frameLocator('iframe[title="扩展 example.reference"]').first()
    stage='wait for protected note'
    await expect(extension.getByLabel('工作笔记')).toBeDisabled()
    stage='release note request'
    releaseData()
    stage='fill and publish'
    await extension.getByLabel('工作笔记').fill('<b>inert notification text</b>')
    await clickPaintedExtensionButton(page,extension.getByRole('button',{name:'发送站内通知'}))
    stage='verify publish'
    await expect(extension.getByText('站内通知已保存',{exact:true})).toBeVisible()
    await expect(page.getByLabel('站内通知列表')).toHaveCount(0)
    stage='reload published note'
    await page.reload()
    stage='verify restored note'
    await expect(extension.getByLabel('工作笔记')).toHaveValue('<b>inert notification text</b>')
    expect(sent).toBe(1)
    stage='open notifications'
    await page.getByRole('button',{name:'站内通知',exact:true}).click()
    const panel=page.getByLabel('站内通知列表')
    await expect(panel.getByText('<b>inert notification text</b>',{exact:true})).toBeVisible()
    await expect(panel.locator('article small')).toContainText('example.reference')
    stage='open notification source'
    await panel.getByRole('button',{name:'打开来源'}).click()
    await expect(panel).toHaveCount(0)
    await expect(page.locator('iframe[title="扩展 example.reference"]')).toHaveCount(1)
    stage='remove notification'
    await page.getByRole('button',{name:'站内通知',exact:true}).click()
    await page.getByRole('button',{name:'移除通知'}).click()
    stage='reload after removal'
    await page.reload()
    stage='verify notification removed'
    await page.getByRole('button',{name:'站内通知',exact:true}).click()
    await expect(page.getByText('暂无通知',{exact:true})).toBeVisible()
    expect(sent).toBe(1)
  }finally{
    test.setTimeout(test.info().timeout+15000)
    console.log('extension notification cleanup after stage:',stage)
    releaseData()
    expect((await page.request.delete('/api/v1/extensions/example.reference?cleanup=true',{headers:await headers(page),timeout:5000})).status()).toBe(204)
  }
})

test('independent extension persists metadata draft and updates real instance',async({page})=>{
  test.setTimeout(120000)
  await realLogin(page,fixture.admin)
  const pkg=JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/reference.blora-extension.json',import.meta.url),'utf8'))
  const instances=(await (await page.request.get('/api/v1/instances')).json()).items
  const instance=instances[0]
  let writes=0
  page.on('request',r=>{if(r.method()==='PATCH'&&r.url().endsWith('/extensions/example.reference/resource'))writes++})
  try{
    expect((await page.request.post('/api/v1/extensions/install-package',{headers:await headers(page),data:pkg})).status()).toBe(201)
    await openExtensions(page)
    await page.locator('[data-app-id="example.reference"]').getByRole('button',{name:'打开',exact:true}).click()
    const overview=page.frameLocator('iframe[title="扩展 example.reference"]').first()
    await overview.getByLabel('节点标识').fill(instance.nodeId)
    await overview.getByLabel('实例标识',{exact:true}).fill(instance.instanceId)
    await overview.getByRole('button',{name:'打开实例工作窗口'}).click()
    const resource=page.frameLocator('iframe[title="扩展 example.reference"]').last()
    await expect(resource.getByLabel('实例名称',{exact:true})).toHaveValue(instance.name)
    const name='Extension metadata '+randomUUID().slice(0,8)
    await resource.getByLabel('实例名称',{exact:true}).fill(name)
    await page.reload()
    await expect(resource.getByLabel('实例名称',{exact:true})).toHaveValue(name)
    expect(writes).toBe(0)
    await page.evaluate(()=>{
      const original=Storage.prototype.setItem
      ;(window as any).__restoreNotificationStorage=()=>{Storage.prototype.setItem=original}
      Storage.prototype.setItem=function(key:string,value:string){if(this===sessionStorage&&key.startsWith('blora:tail:'))throw new Error('injected recovery failure');return original.call(this,key,value)}
    })
    await clickPaintedExtensionButton(page,resource.getByRole('button',{name:'保存实例名称',exact:true}))
    await expect(resource.getByText('Error: injected recovery failure',{exact:true})).toBeVisible()
    expect(writes).toBe(0)
    await page.evaluate(()=>{(window as any).__restoreNotificationStorage()})
    await clickPaintedExtensionButton(page,resource.getByRole('button',{name:'重试保存实例名称',exact:true}))
    await expect(resource.getByText('实例名称已保存',{exact:true})).toBeVisible()
    expect(writes).toBe(1)
    const updated=(await (await page.request.get('/api/v1/instances')).json()).items.find((x:any)=>x.instanceId===instance.instanceId)
    expect(updated.name).toBe(name)
    await page.reload()
    await expect(resource.getByLabel('实例名称',{exact:true})).toHaveValue(name)
    expect(writes).toBe(1)
  }finally{
    test.setTimeout(test.info().timeout+15000)
    const current=(await (await page.request.get('/api/v1/instances')).json()).items.find((x:any)=>x.instanceId===instance.instanceId)
    expect((await page.request.patch('/api/v1/instances/'+instance.instanceId,{headers:await headers(page),data:{revision:current.configRevision,name:instance.name}})).status()).toBe(200)
    expect((await page.request.delete('/api/v1/extensions/example.reference?cleanup=true',{headers:await headers(page)})).status()).toBe(204)
  }
})

async function openExtensions(page:Page) {
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button', {name:'应用市场', exact:true}).click()
}
async function headers(page:Page) {
  return {'X-CSRF-Token': (await (await page.request.get('/api/v1/session',{timeout:10000})).json()).csrfToken, 'Idempotency-Key': randomUUID()}
}

test('independent extension migrates server data and view state and preserves both after failure',async({page})=>{
  test.setTimeout(120000)
  await realLogin(page,fixture.admin)
  const load=(name:string)=>JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/'+name,import.meta.url),'utf8'))
  const base=load('reference.blora-extension.json'),v2=load('reference-v2.blora-extension.json'),v3=load('reference-v3.blora-extension.json')
  try{
    expect((await page.request.post('/api/v1/extensions/install-package',{headers:await headers(page),data:base})).status()).toBe(201)
    await openExtensions(page)
    await page.locator('[data-app-id="example.reference"]').getByRole('button',{name:'打开',exact:true}).click()
    const frame=page.frameLocator('iframe[title="扩展 example.reference"]')
    await frame.getByLabel('工作笔记').fill('用户服务器数据与未保存现场')
    await frame.getByRole('button',{name:'保存服务器笔记'}).click()
    await expect(frame.getByLabel('服务器笔记')).toContainText('用户服务器数据与未保存现场')
    expect((await page.request.post('/api/v1/extensions/example.reference/upgrade',{headers:await headers(page),data:v2})).status()).toBe(200)
    await page.reload()
    await expect(frame.getByLabel('工作笔记')).toHaveValue('用户服务器数据与未保存现场')
    await expect(frame.getByLabel('服务器笔记')).toContainText('"schemaVersion":2')
    await expect(frame.getByLabel('服务器笔记')).toContainText('"noteFormat":2')
    expect((await page.request.post('/api/v1/extensions/example.reference/upgrade',{headers:await headers(page),data:v3})).status()).toBe(409)
    await page.reload()
    await expect(frame.getByLabel('工作笔记')).toHaveValue('用户服务器数据与未保存现场')
    await expect(frame.getByLabel('服务器笔记')).toContainText('"schemaVersion":2')
    expect((await page.request.post('/api/v1/extensions/example.reference/rollback',{headers:await headers(page)})).status()).toBe(200)
    await page.reload()
    await expect(frame.getByLabel('工作笔记')).toHaveValue('用户服务器数据与未保存现场')
    await expect(frame.getByLabel('服务器笔记')).toContainText('"schemaVersion":1')
    await expect(frame.getByLabel('服务器笔记')).toContainText('用户服务器数据与未保存现场')
  }finally{
    expect((await page.request.delete('/api/v1/extensions/example.reference?cleanup=true',{headers:await headers(page)})).status()).toBe(204)
  }
})

test('independent package installs from the desktop and runs its WASI backend on a real node',async({page})=>{
  test.setTimeout(120000)
  let stage='login'
  const packagePath=new URL('../../../sdk/examples/reference-app/reference.blora-extension.json',import.meta.url)
  let submissions=0
  page.on('request',request=>{if(request.method()==='POST'&&request.url().endsWith('/extensions/example.reference/tasks'))submissions++})
  await realLogin(page,fixture.admin)
  await openExtensions(page)
  try{
    stage='install package'
    await page.getByLabel('扩展包').setInputFiles(packagePath.pathname)
    await page.getByRole('button',{name:'安装或升级包'}).click()
    const row=page.locator('[data-app-id="example.reference"]')
    await expect(row).toBeVisible()
    await row.getByRole('button',{name:'打开',exact:true}).click()
    const frames=page.locator('iframe[title="扩展 example.reference"]')
    stage='open resource'
    await frames.first().contentFrame().getByLabel('节点标识').fill(fixture.nodeIds[0]!)
    await clickPaintedExtensionButton(page,frames.first().contentFrame().getByRole('button',{name:'打开节点工作窗口'}))
    await expect(frames).toHaveCount(2)
    const target=frames.nth(1).contentFrame()
    await expect(target.getByLabel('当前资源')).toContainText(fixture.nodeIds[0]!)
    await target.getByLabel('工作笔记').fill('hello 世界')
    stage='submit and lose receipt'
    const lostKeys:string[]=[]
    let originalTaskId=''
    await page.route('**/extensions/example.reference/tasks',async route=>{
      lostKeys.push(route.request().headers()['idempotency-key']!)
      if(lostKeys.length===1){
        const response=await route.fetch()
        expect(response.status()).toBe(202)
        originalTaskId=(await response.json()).task.taskId
        await route.abort('connectionreset')
      }else await route.continue()
    })
    await clickPaintedExtensionButton(page,target.getByRole('button',{name:'提交节点任务'}))
    // The route deliberately drops a response after the server accepts it.
    // Chromium and WebKit use different network-error strings; require an
    // explicit error state, then prove the stable request identity recovers.
    await expect(target.getByRole('status')).toHaveText(/^Error:/)
    await page.reload()
    stage='restore and retry same request'
    await expect(target.getByLabel('工作笔记')).toHaveValue('hello 世界')
    expect(submissions).toBe(1)
    const receipt=page.waitForResponse(r=>r.url().endsWith('/extensions/example.reference/tasks')&&r.request().method()==='POST')
    await clickPaintedExtensionButton(page,target.getByRole('button',{name:'提交节点任务'}))
    const accepted=await receipt
    stage='verify real task'
    expect(accepted.status()).toBe(202)
    const {task}=await accepted.json()
    expect(task.taskId).toBe(originalTaskId)
    expect(lostKeys).toHaveLength(2)
    expect(lostKeys[0]).toBeTruthy()
    expect(lostKeys[1]).toBe(lostKeys[0])
    await expect.poll(async()=>{
      const response=await page.request.get(`/api/v1/tasks/${task.taskId}`)
      expect(response.status()).toBe(200)
      return (await response.json()).task.state
    },{timeout:75000}).toBe('SUCCEEDED')
    const completed=(await (await page.request.get(`/api/v1/tasks/${task.taskId}`)).json()).task
    expect(completed.result.result.characters).toBe(8)
    expect(completed.result.result.words).toBe(2)
    expect(completed.resource.nodeId).toBe(fixture.nodeIds[0])
    await expect(target.getByRole('status')).toContainText('任务已完成')
    await page.reload()
    await expect(frames.nth(1).contentFrame().getByRole('status')).toContainText('任务已完成')
    await expect(frames.nth(1).contentFrame().getByLabel('工作笔记')).toHaveValue('hello 世界')
    expect(submissions).toBe(2)
    stage='shortcut and move view'
    await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'添加当前资源到桌面'}))
    await expect(page.getByText('参考扩展资源',{exact:true})).toBeVisible()
    stage='close view'
    await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'移到新窗口'}))
    await expect(frames.nth(1).contentFrame().getByLabel('工作笔记')).toHaveValue('hello 世界')
    await page.reload()
    await expect(frames.nth(1).contentFrame().getByRole('status')).toContainText('任务已完成')
    await expect(page.getByText('参考扩展资源',{exact:true})).toBeVisible()
    await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'关闭当前视图'}))
    await expect(frames).toHaveCount(1)
    expect(submissions).toBe(2)
  }finally{
    test.setTimeout(test.info().timeout+15000)
    console.log('extension package cleanup after stage:',stage)
    const removed=await page.request.delete('/api/v1/extensions/example.reference',{headers:await headers(page)})
    expect([204,404]).toContain(removed.status())
  }
})

test('real local catalog browses two versions, upgrades, disables and executes the sandbox bundle', async ({page}) => {
  await realLogin(page, fixture.admin)
  await openExtensions(page)
  try {
    const catalog = page.locator('.market-app > section').filter({has:page.getByRole('heading',{name:'可安装应用'})})
    const catalogCards = catalog.locator(`.market-card[data-app-id="${extensionId}"]`)
    await expect(catalogCards).toHaveCount(2)
    await expect(catalogCards.first()).toBeVisible()
    await catalogCards.filter({hasText:'版本 1.0.0'}).getByRole('button', {name:'安装此版本', exact:true}).click()
    const installed = page.locator('.application section').filter({hasText:'已安装'})
    await expect(installed.locator('.task-row').filter({hasText:'v1.0.0'})).toHaveCount(1)
    await catalogCards.filter({hasText:'版本 1.1.0'}).getByRole('button', {name:'安装此版本', exact:true}).click()
    await expect(installed.locator('.task-row').filter({hasText:'v1.1.0'})).toHaveCount(1)
    const bundle = await page.request.get(`/api/v1/extensions/${extensionId}/bundle`)
    expect(bundle.status()).toBe(200)
    expect(await bundle.text()).toContain('1.1.0')
    const enablePath = `/api/v1/extensions/${extensionId}/enabled`
    expect((await page.request.post(enablePath, {headers:await headers(page), data:{enabled:false}})).status()).toBe(200)
    expect((await page.request.get(`/api/v1/extensions/${extensionId}/bundle`)).status()).toBe(409)
    expect((await page.request.post(enablePath, {headers:await headers(page), data:{enabled:true}})).status()).toBe(200)
    await page.reload()
    await openExtensions(page)
    await installed.locator('.task-row').filter({hasText:'v1.1.0'}).getByRole('button', {name:'打开', exact:true}).click()
    const frame = page.locator(`iframe[title="扩展 ${extensionId}"]`)
    await expect(frame).toBeVisible()
    await expect.poll(async() => frame.contentFrame().locator('body').getAttribute('data-blora-extension')).toContain('1.1.0')
  } finally {
    const remove = await page.request.delete(`/api/v1/extensions/${extensionId}`, {headers:await headers(page)})
    expect([204, 404]).toContain(remove.status())
  }
})

test('member browser context cannot cross extension capability or node grants', async ({page, browser}) => {
  test.setTimeout(120000)
  await realLogin(page, fixture.admin)
  const pkg = JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/reference.blora-extension.json', import.meta.url), 'utf8'))
  const users = (await (await page.request.get('/api/v1/users')).json()).items as {userId:string; name:string}[]
  const member = users.find(user => user.name === fixture.member.name)
  expect(member).toBeTruthy()
  const extensionResource = `/api/v1/extensions/example.reference/resource?kind=node&id=${encodeURIComponent(fixture.nodeIds[0]!)}&nodeId=${encodeURIComponent(fixture.nodeIds[0]!)}`
  const extensionTask = '/api/v1/extensions/example.reference/tasks'
  const memberContext = await browser.newContext({ignoreHTTPSErrors: true})
  const memberPage = await memberContext.newPage()
  const appGrant = {userId: member!.userId, resource: {kind: 'extension', id: 'example.reference'}, action: 'app.use'}
  const nodeGrant = {userId: member!.userId, resource: {kind: 'node', id: fixture.nodeIds[0], nodeId: fixture.nodeIds[0]}, action: 'node.read'}
  try {
    expect((await page.request.post('/api/v1/extensions/install-package', {headers: await headers(page), data: pkg})).status()).toBe(201)
    await realLogin(memberPage, fixture.member)
    expect((await memberPage.request.get(extensionResource)).status()).toBe(403)
    expect((await memberPage.request.post(extensionTask, {headers: await headers(memberPage), data: {nodeId: fixture.nodeIds[0], payload: {probe: 'no-app-grant'}}})).status()).toBe(403)

    expect((await page.request.post('/api/v1/grants', {headers: await headers(page), data: appGrant})).status()).toBe(200)
    expect((await memberPage.request.get(extensionResource)).status()).toBe(403)
    expect((await memberPage.request.post(extensionTask, {headers: await headers(memberPage), data: {nodeId: fixture.nodeIds[0], payload: {probe: 'no-node-grant'}}})).status()).toBe(403)

    expect((await page.request.post('/api/v1/grants', {headers: await headers(page), data: nodeGrant})).status()).toBe(200)
    expect((await memberPage.request.get(extensionResource)).status()).toBe(200)
    expect((await page.request.delete('/api/v1/grants', {headers: await headers(page), data: appGrant})).status()).toBe(200)
    expect((await memberPage.request.get(extensionResource)).status()).toBe(403)
  } finally {
    await page.request.delete('/api/v1/grants', {headers: await headers(page), data: appGrant})
    await page.request.delete('/api/v1/grants', {headers: await headers(page), data: nodeGrant})
    const removed = await page.request.delete('/api/v1/extensions/example.reference?cleanup=true', {headers: await headers(page)})
    expect([204, 404]).toContain(removed.status())
    await memberContext.close()
  }
})
