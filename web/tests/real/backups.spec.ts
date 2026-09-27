import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {realLogin} from './login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8')) as {admin:{name:string;password:string};instanceIds:string[]}

async function headers(page:Page){return{'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
async function taskState(page:Page,id:string){return(await(await page.request.get(`/api/v1/tasks/${id}`)).json()).task as {state:string;result?:unknown}}
async function awaitTask(page:Page,id:string){await expect.poll(async()=> (await taskState(page,id)).state,{timeout:30000}).toBe('SUCCEEDED');return taskState(page,id)}

test('actual backup archive restore plan and schedule flow',async({page})=>{
  await realLogin(page,fixture.admin)
  const instance=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance as {instanceId:string;nodeId:string;name:string}
  const filesBase=`/api/v1/instances/${instance.instanceId}/files`,source=`browser-backup-${randomUUID()}.txt`,target=`browser-restored-${randomUUID()}.txt`
  const write=await page.request.put(`${filesBase}/content`,{headers:await headers(page),data:{path:source,text:'真实备份正文\n中文校验',version:'missing'}})
  expect(write.status()).toBe(202);await awaitTask(page,(await write.json()).task.taskId)
  const stat=await page.request.get(`${filesBase}/stat?path=${encodeURIComponent(source)}`);expect(stat.status()).toBe(200);const sourceStat=await stat.json()
  const backup=await page.request.post(`/api/v1/instances/${instance.instanceId}/backups`,{headers:await headers(page),data:{path:source,version:sourceStat.version,compression:'deflate',consistency:{mode:'files'}}})
  expect(backup.status()).toBe(202);const backupTask=await awaitTask(page,(await backup.json()).task.taskId);const result=(typeof backupTask.result==='string'?JSON.parse(backupTask.result):backupTask.result) as {snapshot?:{id:string;state:string}}
  expect(result.snapshot?.state).toBe('ready');expect(result.snapshot?.id).toBeTruthy()

  await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:instance.name,exact:true}).click();await page.getByRole('button',{name:'打开备份与计划',exact:true}).click()
  await expect(page.getByRole('heading',{name:`${instance.name} · 备份`,exact:true})).toBeVisible();await expect(page.getByText(source,{exact:true})).toBeVisible()
  await page.getByLabel('恢复目标路径',{exact:true}).fill(target);await page.getByRole('button',{name:'生成恢复计划',exact:true}).click();await expect(page.getByRole('status')).toContainText('恢复计划已生成')
  const restoreResponse=page.waitForResponse(response=>response.url().endsWith(`/api/v1/instances/${instance.instanceId}/restores`)&&response.request().method()==='POST')
  await page.getByRole('button',{name:'确认覆盖此恢复计划',exact:true}).click();const restore=await restoreResponse;expect(restore.status()).toBe(202);await awaitTask(page,(await restore.json()).task.taskId)
  const restored=await page.request.get(`${filesBase}/content?path=${encodeURIComponent(target)}`);expect(restored.status()).toBe(200);expect((await restored.json()).text).toBe('真实备份正文\n中文校验')

  await page.getByRole('button',{name:'＋ 新建计划',exact:true}).click();await page.getByLabel('计划备份路径',{exact:true}).fill(source);await page.getByLabel('定时任务 Cron',{exact:true}).fill('*/5 * * * *');await page.getByRole('button',{name:'保存定时任务',exact:true}).click();await expect(page.getByRole('status')).toContainText('定时任务已保存')
  const schedules=(await(await page.request.get('/api/v1/schedules')).json()).items as Array<{spec:{id:string;resource:{id:string}};configRevision:number}>,created=schedules.find(item=>item.spec.resource.id===instance.instanceId)
  expect(created).toBeTruthy();if(created)expect((await page.request.delete(`/api/v1/schedules/${created.spec.id}`,{headers:await headers(page),data:{revision:created.configRevision}})).status()).toBe(200)
})
