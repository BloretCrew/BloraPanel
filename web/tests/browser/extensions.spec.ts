import {test, expect} from '@playwright/test'
import {readFileSync} from 'node:fs'
import {clickPaintedExtensionButton,clickPaintedExtensionControl} from '../extension-paint'

test('focusing an extension window preserves its iframe and uncommitted local input',async({page})=>{
  const manifest={appId:'example.focus',packageVersion:'1.0.0',hostApiVersion:1,title:'焦点扩展',icon:'✦',color:'#fff',permissions:[],entrypoints:['overview'],resourceHandlers:[],capabilities:['window.open'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['overview'],movable:true},stateSchemaVersion:1}
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(path.endsWith('/session'))return route.fulfill({json:{user:{userId:'focus-user',name:'焦点测试',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/extensions')return route.fulfill({json:{items:[{manifest,enabled:true}]}})
    if(path.endsWith('/example.focus/manifest'))return route.fulfill({json:manifest})
    if(path.endsWith('/example.focus/bundle'))return route.fulfill({contentType:'application/javascript',body:`export function start(){
      document.body.dataset.boot=crypto.randomUUID();
      const input=document.createElement('input');input.setAttribute('aria-label','扩展本地输入');document.body.append(input);
      const button=document.createElement('button');button.textContent='记录点击';button.onclick=()=>{document.body.dataset.clicks=String(Number(document.body.dataset.clicks||0)+1)};document.body.append(button);
    }`})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/');await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'应用市场',exact:true}).click();await page.getByRole('button',{name:'打开',exact:true}).click()
  const iframe=page.locator('iframe[title="扩展 example.focus"]'),extension=iframe.contentFrame()
  await extension.getByLabel('扩展本地输入').fill('尚未向宿主保存的正文')
  const boot=await extension.locator('body').getAttribute('data-boot')
  expect(boot).toBeTruthy()
  for(let index=0;index<3;index++){
    await page.locator('.dock-item[aria-label="应用市场"]').click()
    await page.locator('.dock-item[aria-label="焦点扩展"]').click()
    await expect(extension.locator('body')).toHaveAttribute('data-boot',boot!)
    await expect(extension.getByLabel('扩展本地输入')).toHaveValue('尚未向宿主保存的正文')
    await extension.getByRole('button',{name:'记录点击',exact:true}).click()
    await expect(extension.locator('body')).toHaveAttribute('data-clicks',String(index+1))
  }
})

test('sandbox controls only its own view and creates a persistent resource shortcut',async({page})=>{
  page.on('pageerror',error=>console.error(error.message))
  const manifest={appId:'example.desktop',packageVersion:'1.0.0',hostApiVersion:1,title:'桌面扩展',icon:'✦',color:'#fff',permissions:[],entrypoints:['overview'],resourceHandlers:['node'],capabilities:['window.open','window.move','window.close','shortcut.create'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['overview','resource'],movable:true},stateSchemaVersion:1}
  await page.route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    if(path.endsWith('/session'))return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/extensions')return route.fulfill({json:{items:[{manifest,enabled:true}]}})
    if(path.endsWith('/example.desktop/manifest'))return route.fulfill({json:manifest})
    if(path.endsWith('/example.desktop/bundle'))return route.fulfill({contentType:'application/javascript',body:`export async function start(call){
      const state=await call('state.capture',null);document.body.textContent=JSON.stringify(state);
      for(const [title,method,payload] of [['打开资源','window.open',{resource:{kind:'node',id:'test-node',nodeId:'test-node'},disposition:'new-window'}],['移到新窗口','window.move',{}],['创建快捷入口','shortcut.create',{title:'扩展资源入口'}],['关闭自身','window.close',null]]){
        const button=document.createElement('button');button.textContent=title;button.onclick=()=>call(method,payload);document.body.append(button);
      }
    }`})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'应用市场'}).click()
  await page.getByRole('button',{name:'打开',exact:true}).click()
  const frames=page.locator('iframe[title="扩展 example.desktop"]')
  await clickPaintedExtensionButton(page,frames.first().contentFrame().getByRole('button',{name:'打开资源'}))
  await expect(frames).toHaveCount(2)
  await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'创建快捷入口'}))
  await expect(page.getByText('扩展资源入口',{exact:true})).toBeVisible()
  await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'移到新窗口'}))
  await expect(frames.nth(1).contentFrame().locator('body')).toContainText('test-node')
  await page.reload()
  await expect(frames).toHaveCount(2)
  await expect(page.getByText('扩展资源入口',{exact:true})).toBeVisible()
  await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'关闭自身'}))
  await expect(frames).toHaveCount(1)
  await expect(page.getByText('扩展资源入口',{exact:true})).toBeVisible()
})

test('sandbox migration commits a new state version and preserves state on failure',async({page})=>{
  let version=1
  const base={appId:'example.migration',packageVersion:'1.0.0',hostApiVersion:1,title:'迁移扩展',icon:'✦',color:'#fff',permissions:[],entrypoints:['overview'],resourceHandlers:[],capabilities:['window.open'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['overview'],movable:true}}
  await page.route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    const manifest={...base,packageVersion:`${version}.0.0`,stateSchemaVersion:version}
    if(path.endsWith('/session'))return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/extensions')return route.fulfill({json:{items:[{manifest,enabled:true}]}})
    if(path.endsWith('/example.migration/manifest'))return route.fulfill({json:manifest})
    if(path.endsWith('/example.migration/bundle')){
      const body=version===1?`export async function start(call){await call('state.restore',{schemaVersion:1,state:{note:'original'}});document.body.textContent='original'}`:
        version===2?`export function migrate(old,target){return {schemaVersion:target,state:{note:old.state.note+' migrated'}}} export async function start(call){document.body.textContent=JSON.stringify(await call('state.capture',null))}`:
        `export function migrate(){throw new Error('migration deliberately failed')} export function start(){throw new Error('must not start')}`
      return route.fulfill({contentType:'application/javascript',body})
    }
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'应用市场'}).click()
  await page.getByRole('button',{name:'打开',exact:true}).click()
  const frame=page.frameLocator('iframe[title="扩展 example.migration"]')
  await expect(frame.locator('body')).toHaveText('original')
  version=2
  await page.reload()
  await expect(frame.locator('body')).toContainText('original migrated')
  await expect(frame.locator('body')).toContainText('"schemaVersion":2')
  version=3
  await page.reload()
  await expect(page.getByRole('alert').filter({hasText:'migration deliberately failed'})).toBeVisible()
  version=2
  await page.reload()
  await expect(frame.locator('body')).toContainText('original migrated')
  await expect(frame.locator('body')).not.toContainText('migrated migrated')
})

test('sandbox workspace copy on another device keeps resource identity and migrates view state',async({page})=>{
  test.setTimeout(60000)
  let version=1,remoteState:any,remoteId='',revision=0
  await page.addInitScript(()=>{if(window.top===window)localStorage.setItem('blora:device','device-a')})
  await page.route('**/api/v1/**',async route=>{
    const request=route.request(),path=new URL(request.url()).pathname
    const manifest={appId:'example.crossdevice',packageVersion:`${version}.0.0`,hostApiVersion:1,title:'跨设备扩展',icon:'✦',color:'#fff',permissions:[],entrypoints:['overview','resource'],resourceHandlers:['instance'],capabilities:['window.open','resource.read'],dependencies:{},windowPolicy:'multiple',tabPolicy:{types:['overview','resource'],movable:true},stateSchemaVersion:version}
    if(path.endsWith('/session'))return route.fulfill({json:{user:{userId:'cross-user',name:'跨设备用户',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/extensions')return route.fulfill({json:{items:[{manifest,enabled:true}]}})
    if(path.endsWith('/example.crossdevice/manifest'))return route.fulfill({json:manifest})
    if(path.endsWith('/example.crossdevice/bundle')){
      const body=version===1?`export async function start(call){const current=await call('state.capture',null);if(!current.state.note)await call('state.restore',{schemaVersion:1,state:{note:'跨设备现场'}});if(!current.resource){const button=document.createElement('button');button.textContent='打开资源';button.onclick=()=>call('window.open',{resource:{kind:'instance',id:'instance-1',nodeId:'node-old'},disposition:'new-window'});document.body.append(button)}}`:`export function migrate(previous,target){return {schemaVersion:target,state:{...previous.state,note:previous.state.note+' · 已迁移'}}} export async function start(call){const captured=await call('state.capture',null);document.body.textContent=JSON.stringify(captured)}`
      return route.fulfill({contentType:'application/javascript',body})
    }
    if(path.endsWith('/example.crossdevice/resource'))return route.fulfill({json:{resource:{kind:'instance',id:'instance-1',nodeId:'node-new'},name:'跨设备实例',state:'RUNNING'}})
    if(path==='/api/v1/workspaces'&&request.method()==='GET')return route.fulfill({json:{items:revision?[{workspaceId:remoteId,title:'跨设备副本',deviceId:'device-a',revision,schemaVersion:1,includeContent:false,updatedAt:new Date().toISOString()}]:[]}})
    if(path.startsWith('/api/v1/workspaces/')&&request.method()==='PUT'){
      const body=JSON.parse(request.postData()||'{}');remoteState=body.state;remoteId=decodeURIComponent(path.slice('/api/v1/workspaces/'.length));revision++
      return route.fulfill({json:{metadata:{workspaceId:remoteId,title:'跨设备副本',deviceId:body.deviceId,revision,schemaVersion:body.schemaVersion,includeContent:body.includeContent,updatedAt:new Date().toISOString()}}})
    }
    if(path.startsWith('/api/v1/workspaces/')&&request.method()==='GET')return route.fulfill({json:{metadata:{workspaceId:remoteId,title:'跨设备副本',deviceId:'device-a',revision,schemaVersion:1,includeContent:false,updatedAt:new Date().toISOString()},state:remoteState}})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'应用市场'}).click();await page.getByRole('button',{name:'打开',exact:true}).click()
  const overview=page.frameLocator('iframe[title="扩展 example.crossdevice"]').first()
  await clickPaintedExtensionButton(page,overview.getByRole('button',{name:'打开资源',exact:true}))
  const frames=page.locator('iframe[title="扩展 example.crossdevice"]');await expect(frames).toHaveCount(2)
  await page.getByRole('button',{name:'切换工作区'}).click();await page.getByRole('button',{name:'同步布局/引用',exact:true}).click();await expect.poll(()=>revision).toBe(1)
  version=2
  await page.evaluate(()=>{sessionStorage.clear();localStorage.clear();localStorage.setItem('blora:device','device-b')})
  await page.reload()
  await page.getByRole('button',{name:'切换工作区'}).click();await page.getByRole('button',{name:/跨设备副本 · v1/}).click();await expect(page.getByRole('dialog',{name:'工作区'})).toHaveCount(0)
  const restoredFrames=page.locator('iframe[title="扩展 example.crossdevice"]');await expect(restoredFrames).toHaveCount(2)
  const restored=restoredFrames.last().contentFrame()
  await expect(restored.locator('body')).toContainText('跨设备现场 · 已迁移')
  await expect(restored.locator('body')).toContainText('node-new')
})

test('independently packaged reference extension renders and restores its note', async ({page}) => {
  const pkg=JSON.parse(readFileSync(new URL('../../../sdk/examples/reference-app/reference.blora-extension.json', import.meta.url),'utf8'))
  let installed=false
  let createdTasks=0,taskState='RUNNING'
  const cancelKeys:string[]=[]
  let failTaskRead=false
  await page.route('**/api/v1/**', async route => {
    const path=new URL(route.request().url()).pathname
    if(path.endsWith('/session')) return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/extensions') return route.fulfill({json:{items:installed?[{manifest:pkg.manifest,sha256:pkg.sha256,enabled:true}]:[]}})
    if(path==='/api/v1/extensions/install-package') {
      const submitted=route.request().postDataJSON()
      expect(submitted.sha256).toBe(pkg.sha256)
      expect(submitted.payload).toBe(pkg.payload)
      installed=true
      return route.fulfill({status:201,json:{manifest:pkg.manifest,enabled:true}})
    }
    if(path.endsWith('/example.reference/manifest')) return route.fulfill({json:pkg.manifest})
    if(path.endsWith('/example.reference/bundle')) {
      const bytes=Buffer.from(pkg.payload,'base64').toString('utf8')
      const body=bytes.startsWith('BLORA-BUNDLE-1\n')?JSON.parse(bytes.slice('BLORA-BUNDLE-1\n'.length)).frontend:bytes
      return route.fulfill({contentType:'application/javascript',body})
    }
    if(path.endsWith('/example.reference/resource')) return route.fulfill({json:{resource:{kind:'node',id:'node-test',nodeId:'node-test'},name:'参考测试节点',state:'ONLINE'}})
    if(path==='/api/v1/extensions/example.reference/tasks'){
      createdTasks++
      expect(route.request().postDataJSON().nodeId).toBe('node-test')
      return route.fulfill({status:202,json:{task:{taskId:'reference-task',state:taskState,phase:'running',resource:{kind:'node',id:'node-test',nodeId:'node-test'}}}})
    }
    if(path.startsWith('/api/v1/extensions/example.reference/tasks/reference-task/')){
      if(path.endsWith('/status')&&failTaskRead)return route.abort('connectionreset')
      if(path.endsWith('/cancel')){
        cancelKeys.push(route.request().headers()['idempotency-key']!)
        if(cancelKeys.length===1)return route.abort('connectionreset')
        taskState='CANCELLED'
      }
      return route.fulfill({json:{task:{taskId:'reference-task',state:taskState,phase:taskState==='CANCELLED'?'cancelled':'running',resource:{kind:'node',id:'node-test',nodeId:'node-test'}}}})
    }
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'应用市场'}).click()
  await page.getByLabel('扩展包').setInputFiles({name:'reference.blora-extension.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify(pkg))})
  await page.getByRole('button',{name:'安装或升级包'}).click()
  await page.getByRole('button',{name:'打开',exact:true}).click()
  const frame=page.frameLocator('iframe[title="扩展 example.reference"]')
  await expect(frame.getByRole('heading',{name:'参考扩展'})).toBeVisible()
  await expect(frame.getByLabel('工作笔记')).toBeEnabled()
  await clickPaintedExtensionControl(page,frame.getByLabel('工作笔记'))
  await frame.getByLabel('工作笔记').fill('独立包中的未保存笔记')
  await page.reload()
  await expect(frame.getByLabel('工作笔记')).toHaveValue('独立包中的未保存笔记')
  await clickPaintedExtensionControl(page,frame.getByLabel('节点标识'))
  await frame.getByLabel('节点标识').fill('node-test')
  await clickPaintedExtensionButton(page,frame.getByRole('button',{name:'打开节点工作窗口'}))
  const frames=page.locator('iframe[title="扩展 example.reference"]')
  await expect(frames).toHaveCount(2)
  const resourceFrame=frames.nth(1).contentFrame()
  await expect(resourceFrame.getByLabel('当前资源')).toContainText('参考测试节点')
  await expect(resourceFrame.getByLabel('节点标识')).toHaveValue('node-test')
  await expect(resourceFrame.getByLabel('工作笔记')).toHaveValue('')
  await clickPaintedExtensionButton(page,resourceFrame.getByRole('button',{name:'提交节点任务'}))
  await expect(resourceFrame.getByRole('status')).toContainText('任务执行中')
  await page.reload()
  await expect(frames.nth(1).contentFrame().getByRole('status')).toContainText('任务执行中')
  expect(createdTasks).toBe(1)
  failTaskRead=true
  await clickPaintedExtensionButton(page,resourceFrame.getByRole('button',{name:'刷新任务状态'}))
  await expect(resourceFrame.getByRole('status')).toContainText('网络连接失败')
  await expect(resourceFrame.getByRole('button',{name:'取消任务',exact:true})).toBeDisabled()
  failTaskRead=false
  await clickPaintedExtensionButton(page,resourceFrame.getByRole('button',{name:'刷新任务状态'}))
  await expect(resourceFrame.getByRole('status')).toContainText('任务执行中')
  expect(createdTasks).toBe(1)
  expect(cancelKeys).toHaveLength(0)
  await clickPaintedExtensionButton(page,frames.nth(1).contentFrame().getByRole('button',{name:'取消任务'}))
  await expect.poll(()=>cancelKeys.length).toBe(1)
  await page.reload()
  await expect(resourceFrame.getByRole('status')).toContainText('任务执行中')
  expect(cancelKeys).toHaveLength(1)
  await clickPaintedExtensionButton(page,resourceFrame.getByRole('button',{name:'取消任务'}))
  await expect(frames.nth(1).contentFrame().getByRole('status')).toContainText('任务已取消')
  expect(cancelKeys).toHaveLength(2)
  expect(cancelKeys[0]).toBeTruthy()
  expect(cancelKeys[1]).toBe(cancelKeys[0])
  expect(createdTasks).toBe(1)
})

test('extension management installs, toggles and rolls back an extension', async ({page}) => {
  const requests: {path:string; method:string; body:string}[] = []
  page.on('dialog', dialog => dialog.accept())
  await page.route('**/api/v1/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname
    if (request.method() !== 'GET') requests.push({path, method:request.method(), body:request.postData() || ''})
    if (path.endsWith('/session')) return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if (path === '/api/v1/extensions') return route.fulfill({json:{items:[{manifest:{appId:'example.api',title:'示例扩展',packageVersion:'1.0.0',capabilities:['window.open']},sha256:'00',enabled:true}]}})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button', {name:'应用市场'}).click()
  await expect(page.getByText('示例扩展')).toBeVisible()
  await page.getByRole('button', {name:'禁用'}).click()
  await page.getByRole('button', {name:'回滚'}).click()
  await page.getByRole('button', {name:'卸载'}).click()
  expect(requests.map(x=>x.path)).toEqual(expect.arrayContaining([
    '/api/v1/extensions/example.api/enabled',
    '/api/v1/extensions/example.api/rollback',
    '/api/v1/extensions/example.api',
  ]))
  const toggle = requests.find(x=>x.path.endsWith('/enabled'))
  expect(JSON.parse(toggle!.body).enabled).toBe(false)
})

test('enabled extension bundle runs in opaque sandbox and uses restricted state bridge', async ({page}) => {
  await page.route('**/api/v1/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname
    if (path.endsWith('/session')) return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if (path === '/api/v1/extensions') return route.fulfill({json:{items:[{manifest:{appId:'example.api',title:'示例扩展',packageVersion:'1.0.0',capabilities:['window.open']},sha256:'00',enabled:true}]}})
    if (path === '/api/v1/extensions/example.api/manifest') return route.fulfill({json:{appId:'example.api',packageVersion:'1.0.0',hostApiVersion:1,title:'示例扩展',icon:'✦',color:'#8ec5ff',entrypoints:['overview'],resourceHandlers:[],permissions:[],capabilities:['window.open'],dependencies:[],windowPolicy:'multiple',tabPolicy:{types:['overview'],movable:true},stateSchemaVersion:1}})
    if (path === '/api/v1/extensions/example.api/bundle') return route.fulfill({
      headers:{'Content-Type':'application/javascript'},
      body:"export const start=async call=>{const state=await call('state.capture',null);let denied=false;try{await call('host.unlisted',null)}catch(e){denied=true}document.body.dataset.bridge=JSON.stringify({state,denied})}"
    })
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button', {name:'应用市场'}).click()
  await page.getByRole('button', {name:'打开',exact:true}).click()
  const frame=page.locator('iframe[title="扩展 example.api"]')
  await expect(frame).toBeVisible()
  await expect.poll(async()=>frame.contentFrame().locator('body').getAttribute('data-bridge')).toContain('"denied":true')
  await expect.poll(async()=>frame.contentFrame().locator('body').getAttribute('data-bridge')).toContain('"schemaVersion":1')
})
