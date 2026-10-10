import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {realLogin} from './login'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
test('task history page survives refresh and returns to latest tasks',async({page})=>{
  test.setTimeout(120000)
  await realLogin(page,fixture.admin)
  const instances=(await(await page.request.get('/api/v1/instances')).json()).items
  const session=await(await page.request.get('/api/v1/session')).json()
  let oldest='',latest=''
  const taskIds:string[]=[]
  for(let i=0;i<101;i++){
    const path='history-'+randomUUID()
    const result=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/files/actions`,{headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},data:{action:'mkdir',path,target:path,targetVersion:'missing'}})
    expect(result.status()).toBe(202)
    const taskId=(await result.json()).task.taskId
    taskIds.push(taskId)
    latest=taskId
    if(!oldest)oldest=taskId
  }
  await page.locator('.taskbar').getByRole('button',{name:/项后台任务/}).click()
  await expect(page.locator('.task-card')).toHaveCount(100)
  await expect(page.locator('.task-card').filter({hasText:oldest})).toHaveCount(0)
  await page.getByRole('button',{name:'更早任务',exact:true}).click()
  await expect(page.locator('.task-card').filter({hasText:oldest})).toHaveCount(1)
  await page.reload()
  await expect(page.locator('.task-card').filter({hasText:oldest})).toHaveCount(1)
  await expect(page.getByRole('button',{name:'最新任务',exact:true})).toBeEnabled()
  await page.getByRole('button',{name:'最新任务',exact:true}).click()
  await expect(page.locator('.task-card')).toHaveCount(100)
  await expect(page.locator('.task-card').filter({hasText:oldest})).toHaveCount(0)
  await expect.poll(async()=>(await(await page.request.get('/api/v1/tasks/'+latest)).json()).task.state,{timeout:60000}).toBe('SUCCEEDED')
  await expect.poll(async()=>Promise.all(taskIds.map(async id=>(await(await page.request.get('/api/v1/tasks/'+id)).json()).task.state)),{timeout:60000}).toEqual(taskIds.map(()=>'SUCCEEDED'))
  const filtered=page.waitForResponse(response=>{const url=new URL(response.url());return url.pathname==='/api/v1/tasks'&&url.searchParams.get('state')==='SUCCEEDED'})
  await selectStyledOption(page,page.getByLabel('任务状态筛选'),'SUCCEEDED')
  expect((await filtered).status()).toBe(200)
  await expect(page.locator('.task-card .status-chip').first()).toHaveText('成功')
  await page.reload()
  await expect(page.getByLabel('任务状态筛选')).toHaveValue('SUCCEEDED')
  const summary=await(await page.request.get('/api/v1/tasks/summary')).json()
  await expect(page.locator('.taskbar').getByRole('button',{name:new RegExp(summary.active+' 项后台任务')})).toBeVisible()
})
test('task completion notifies without task window and refresh never restores dismissed notice',async({page})=>{
  await realLogin(page,fixture.admin)
  await page.getByRole('button',{name:'站内通知',exact:true}).click()
  const panel=page.getByLabel('站内通知列表')
  await expect(panel.getByText('任务通知已连接',{exact:true})).toBeVisible()
  await panel.getByRole('button',{name:'关闭通知'}).click()
  const instances=(await(await page.request.get('/api/v1/instances')).json()).items
  const session=await(await page.request.get('/api/v1/session')).json()
  const path='notification-'+randomUUID()
  const result=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/files/actions`,{headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},data:{action:'mkdir',path,target:path,targetVersion:'missing'}})
  expect(result.status()).toBe(202)
  const taskId=(await result.json()).task.taskId
  await expect.poll(async()=>(await(await page.request.get('/api/v1/tasks/'+taskId)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')
  await expect(page.getByRole('heading',{name:'任务中心',exact:true})).toHaveCount(0)
  await page.getByRole('button',{name:'站内通知',exact:true}).click()
  const notice=panel.locator('article').filter({hasText:taskId})
  await expect(notice.getByText('任务成功',{exact:true})).toBeVisible()
  await page.reload()
  await page.getByRole('button',{name:'站内通知',exact:true}).click()
  await expect(notice).toHaveCount(1)
  await notice.getByRole('button',{name:'打开来源'}).click()
  await expect(page.getByRole('heading',{name:'任务中心',exact:true})).toBeVisible()
  await expect(page.locator('.task-card')).toHaveCount(1)
  await expect(page.locator('.task-card')).toContainText(taskId)
  await expect(page.locator('.task-timing')).toContainText(session.user.userId)
  await expect(page.locator('.task-timing')).toContainText('自接受起耗时（含等待）')
  const stages=page.getByLabel('任务阶段记录')
  await expect(stages.locator('li').first()).toContainText('SUCCEEDED')
  await expect(stages.locator('li').last()).toContainText('QUEUED')
  await page.getByRole('button',{name:'站内通知',exact:true}).click()
  await notice.getByRole('button',{name:'移除通知'}).click()
  await page.reload()
  await page.getByRole('button',{name:'站内通知',exact:true}).click()
  await expect(panel.getByText('任务通知已连接',{exact:true})).toBeVisible()
  await expect(notice).toHaveCount(0)
})

test('task completion reaches the browser notification API and its click opens the exact task',async({page})=>{
  await page.context().grantPermissions(['notifications'],{origin:fixture.url})
  await page.addInitScript(()=>{
    const NativeNotification=window.Notification
    const captured:Array<{title:string;options:NotificationOptions;notification:Notification}> = []
    Object.defineProperty(window,'__bloraCapturedNotifications',{value:captured,configurable:false})
    class ObservedNotification extends NativeNotification{
      constructor(title:string,options?:NotificationOptions){
        super(title,options)
        captured.push({title,options:options??{},notification:this})
      }
    }
    Object.defineProperty(ObservedNotification,'permission',{get:()=>NativeNotification.permission})
    Object.defineProperty(ObservedNotification,'requestPermission',{value:()=>NativeNotification.requestPermission()})
    Object.defineProperty(window,'Notification',{value:ObservedNotification,configurable:true})
  })
  await realLogin(page,fixture.admin)
  await expect.poll(()=>page.evaluate(()=>Notification.permission)).toBe('granted')
  await page.locator('[data-app="blora.tasks"]').click()
  await page.getByRole('button',{name:'启用系统通知',exact:true}).click()
  await expect(page.getByRole('status')).toContainText('系统通知已启用')

  const instances=(await(await page.request.get('/api/v1/instances')).json()).items
  const session=await(await page.request.get('/api/v1/session')).json()
  const taskPath='native-notification-'+randomUUID()
  const accepted=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/files/actions`,{headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},data:{action:'mkdir',path:taskPath,target:taskPath,targetVersion:'missing'}})
  expect(accepted.status()).toBe(202)
  const taskId=(await accepted.json()).task.taskId
  await expect.poll(async()=>(await(await page.request.get('/api/v1/tasks/'+taskId)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')
  await expect.poll(()=>page.evaluate(()=>((window as unknown as {__bloraCapturedNotifications:Array<{title:string;notification:Notification}>}).__bloraCapturedNotifications).map(item=>({title:item.title,body:item.notification.body,tag:item.notification.tag}))),{timeout:10000}).toEqual([{title:'任务成功',body:'file.mkdir',tag:'blora-task-'+taskId}])

  await page.evaluate(()=>{
    const item=(window as unknown as {__bloraCapturedNotifications:Array<{notification:Notification}>}).__bloraCapturedNotifications[0]
    item.notification.dispatchEvent(new Event('click'))
  })
  const taskWindow=page.locator('.app-window[data-window-mode="resource"]').filter({has:page.locator('.task-card').filter({hasText:taskId})})
  await expect(taskWindow).toHaveCount(1)
  await expect(taskWindow.getByRole('heading',{name:'任务中心',exact:true})).toBeVisible()
  await expect(taskWindow.locator('.task-card')).toHaveCount(1)
  await expect(taskWindow.locator('.task-card')).toContainText(taskId)
})
