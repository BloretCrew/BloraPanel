import {test,expect} from '@playwright/test'

test('task center retries a failed instance operation with a durable related task',async({page})=>{
  const retryKeys:string[]=[]
  let retried=false
  const failed={taskId:'failed-start',requestId:'original-request',action:'instance.start',resource:{kind:'instance',id:'instance-1',nodeId:'node-1'},state:'FAILED',phase:'process_exit',revision:3,createdAt:'2026-09-10T00:00:00Z',updatedAt:'2026-09-10T00:00:01Z',error:'command exited with status 127',result:{runId:'run-unknown',confirmed:false}}
  const retry={...failed,taskId:'retry-start',requestId:'retry-request',retryOf:failed.taskId,state:'QUEUED',phase:'accepted',revision:1,error:undefined,result:undefined}
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname,method=route.request().method()
    if(path.endsWith('/session'))return route.fulfill({json:{user:{userId:'task-user',name:'任务测试员',admin:true},csrfToken:'test-only'}})
    if(path.endsWith('/tasks/'+failed.taskId+'/retry')){retryKeys.push(route.request().headers()['idempotency-key']||'');retried=true;return route.fulfill({status:202,json:{task:retry}})}
    if(path.endsWith('/tasks'))return route.fulfill({json:{items:retried?[failed,retry]:[failed]}})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'任务中心',exact:true}).click()
  await expect(page.getByRole('heading',{name:'任务中心',exact:true})).toBeVisible()
  await expect(page.locator('.task-card')).toContainText('command exited with status 127')
  await expect(page.getByText('查看已确认结果')).toBeVisible()
  await page.getByRole('button',{name:'重试此实例操作',exact:true}).click()
  await expect(page.getByRole('button',{name:'查看重试任务',exact:true})).toBeVisible()
  expect(retryKeys).toHaveLength(1)
  expect(retryKeys[0]).toBeTruthy()
  await page.reload()
  await expect(page.locator('.task-card').filter({hasText:'关联重试自 failed-start'})).toHaveCount(1)
  await expect(page.getByRole('button',{name:'查看重试任务',exact:true})).toBeVisible()
})
