import {readFileSync,writeFileSync} from 'node:fs'
import {createHash,randomUUID} from 'node:crypto'
import {test,expect,type APIRequestContext} from '@playwright/test'
import {realLogin} from './login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
const base=`/api/v1/instances/${fixture.instanceIds[0]}/files`
async function headers(request:APIRequestContext){return {'X-CSRF-Token':(await(await request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
async function state(request:APIRequestContext,id:string){return (await(await request.get(`/api/v1/tasks/${id}`)).json()).task.state}

test('closing the browser leaves extraction running and a resumable original local upload',async({page,context},info)=>{
  test.setTimeout(240000);page.setDefaultTimeout(15000)
  expect(fixture.performance,'Use --performance for the real 10000-file archive source').toBe(true)
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  await realLogin(page,fixture.admin)
  const oldBrowserTab=await page.evaluate(()=>sessionStorage.getItem('blora:tab'))
  const suffix=randomUUID(),archive=`a07-${suffix}.zip`,extracted=`a07-${suffix}-out`,uploadName=`a07-${suffix}.bin`
  const source=await(await context.request.get(`${base}/stat?path=ten-thousand`)).json()
  const compressed=await context.request.post(`${base}/actions`,{headers:await headers(context.request),data:{action:'compress',path:'ten-thousand',target:archive,version:source.version,targetVersion:'missing'}})
  expect(compressed.status(),await compressed.text()).toBe(202)
  const compressId=(await compressed.json()).task.taskId
  await expect.poll(()=>state(context.request,compressId),{timeout:90000}).toBe('SUCCEEDED')
  const instance=(await(await context.request.get('/api/v1/instances')).json()).items.find((item:{instanceId:string})=>item.instanceId===fixture.instanceIds[0])
  await page.locator('[data-app="blora.instances"]').click()
  await page.getByRole('button',{name:instance.name,exact:true}).click()
  await page.getByRole('button',{name:'文件',exact:true}).click()
  await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click()
  const bytes=Buffer.alloc(16<<20);for(let i=0;i<bytes.length;i++)bytes[i]=i%251
  const localFile=info.outputPath(uploadName);writeFileSync(localFile,bytes)
  const uploadAccepted=page.waitForResponse(response=>response.url().endsWith(`${base}/uploads`)&&response.request().method()==='POST')
  await page.locator('input[type="file"]').setInputFiles(localFile)
  await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  const opened=await uploadAccepted;expect(opened.status()).toBe(202)
  const uploadId=(await opened.json()).task.taskId
  await expect.poll(async()=>(await(await context.request.get(`${base}/uploads/${uploadId}`)).json()).upload.offset).toBeGreaterThanOrEqual(65536)
  await page.locator('.file-row').filter({hasText:archive}).click()
  await page.getByRole('button',{name:'解压',exact:true}).click()
  await page.getByRole('textbox',{name:'文件操作目标',exact:true}).fill(extracted)
  const extractionAccepted=page.waitForResponse(response=>response.url().endsWith(`${base}/actions`)&&response.request().method()==='POST')
  await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  const accepted=await extractionAccepted;expect(accepted.status()).toBe(202)
  const extractId=(await accepted.json()).task.taskId
  await expect.poll(()=>state(context.request,extractId),{timeout:15000}).toBe('RUNNING')
  expect(await state(context.request,uploadId)).toBe('WAITING_CLIENT')
  await page.close();expect(page.isClosed()).toBe(true)
  await expect.poll(()=>state(context.request,extractId),{timeout:90000}).toBe('SUCCEEDED')
  const directory=await(await context.request.get(`${base}?path=${encodeURIComponent(extracted+'/ten-thousand')}&limit=1`)).json()
  expect(directory.total).toBe(10000)
  const sample=await(await context.request.get(`${base}/content?path=${encodeURIComponent(extracted+'/ten-thousand/entry-09999.txt')}`)).json()
  expect(sample.text).toBe('Blora performance fixture entry 9999\n')
  const waiting=await(await context.request.get(`${base}/uploads/${uploadId}`)).json()
  expect(waiting.task.state).toBe('WAITING_CLIENT');expect(waiting.upload.offset).toBeGreaterThan(0);expect(waiting.upload.offset).toBeLessThan(bytes.length)
  const confirmed=waiting.upload.offset
  expect((await context.request.get(`${base}/stat?path=${uploadName}`)).status()).toBe(404)

  // A genuinely new browser tab has no copy of the old sessionStorage tail.
  const reopened=await context.newPage();reopened.setDefaultTimeout(15000)
  reopened.on('pageerror',error=>errors.push(error.message))
  await reopened.goto('/');await expect(reopened.locator('.desktop')).toBeVisible()
  expect(await reopened.evaluate(()=>sessionStorage.getItem('blora:tab'))).not.toBe(oldBrowserTab)
  await reopened.locator('.launcher-button').click()
  await reopened.locator('.launcher-grid').getByRole('button',{name:'任务中心',exact:true}).click()
  await reopened.locator('.task-card').filter({hasText:uploadId}).getByRole('button',{name:'继续本机上传',exact:true}).click()
  const row=reopened.locator('.upload-row').filter({hasText:uploadName})
  await expect(row).toContainText(`节点已确认 ${confirmed} / ${bytes.length} B`)
  const offsets:number[]=[],requestTasks:string[]=[]
  reopened.on('request',request=>{if(request.url().endsWith(`/uploads/${uploadId}/chunks`)&&request.method()==='PUT')offsets.push(request.postDataJSON().offset)})
  reopened.on('response',response=>{if(response.url().endsWith(`${base}/uploads`)&&response.request().method()==='POST')void response.json().then(value=>requestTasks.push(value.task.taskId))})
  const chooser=reopened.waitForEvent('filechooser')
  await row.getByRole('button',{name:'选择原文件续传',exact:true}).click()
  await(await chooser).setFiles(localFile)
  await expect(row).toContainText('节点确认已提交',{timeout:90000})
  expect(offsets[0]).toBe(confirmed);expect(requestTasks).toEqual([uploadId])
  expect(await state(context.request,uploadId)).toBe('SUCCEEDED')
  const final=await(await context.request.get(`${base}/stat?path=${uploadName}`)).json()
  expect(final.size).toBe(bytes.length);expect(final.version).toBe('sha256:'+createHash('sha256').update(bytes).digest('hex'))
  const tasks=(await(await context.request.get('/api/v1/tasks?limit=100')).json()).items
  expect(tasks.filter((task:{taskId:string})=>task.taskId===uploadId)).toHaveLength(1)
  expect(errors).toEqual([])
  console.log('BLORA_UPLOAD_CLOSE '+JSON.stringify({extractId,uploadId,confirmedOffset:confirmed,total:bytes.length,extractedEntries:directory.total,resumedOffset:offsets[0],browserTabChanged:true}))
  await reopened.close()
})
