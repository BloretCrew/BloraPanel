import {realLogin} from './login'
import {readFileSync,readdirSync} from 'node:fs'
import {dirname,join,resolve} from 'node:path'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8')) as {admin:{name:string;password:string};instanceIds:string[]}
async function headers(page:Page){return{'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
async function allNodes(page:Page){return(await(await page.request.get('/api/v1/nodes')).json()).items}

test('actual paused Daemon shows offline uncertainty and resumes the original run',async({page})=>{
  test.setTimeout(120000)
  await realLogin(page,fixture.admin)
  const instanceId=fixture.instanceIds[0]!,instance=(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance,node=(await allNodes(page)).find((node:any)=>node.nodeId===instance.nodeId)
  const expected=join(dirname(resolve(process.env.BLORA_E2E_CREDENTIALS!)),node.name+'.json')
  const matches=readdirSync('/proc').filter(id=>/^\d+$/.test(id)).filter(id=>{try{const args=readFileSync(`/proc/${id}/cmdline`,'utf8').split('\0');return args[0]?.endsWith('blora-daemon')&&args[1]==='--config'&&args[2]===expected}catch{return false}})
  expect(matches).toHaveLength(1)
  const pid=Number(matches[0]),birth=()=>readFileSync(`/proc/${pid}/stat`,'utf8').split(') ').at(-1)!.split(' ')[19],identity=birth()
  const act=async(action:string)=>page.request.post(`/api/v1/instances/${instanceId}/actions`,{headers:await headers(page),data:{action}})
  await act('start');await expect.poll(async()=>(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance.state,{timeout:30000}).toBe('RUNNING')
  const original=(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance.runId
  await page.locator('[data-app="blora.instances"]').click();
  const card=page.locator('.instance-card').filter({has:page.getByRole('button',{name:instance.name,exact:true})})
  await card.getByTitle('添加到桌面',{exact:true}).click();await card.getByRole('button',{name:instance.name,exact:true}).click()
  let pendingTask=''
  try{
    expect(birth()).toBe(identity);process.kill(pid,'SIGSTOP')
    await expect.poll(async()=>(await allNodes(page)).find((item:any)=>item.nodeId===node.nodeId).state,{timeout:65000}).toBe('OFFLINE')
    await expect(page.getByText('节点失联，运行结果等待重新确认。最近记录不代表当前进程已停止。',{exact:true})).toBeVisible({timeout:15000})
    await expect(page.locator('[data-shortcut] .shortcut-resource-status')).toHaveText('节点失联',{timeout:15000})
    const uncertain=(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance
    expect(uncertain.runId).toBe(original);expect(uncertain.state).not.toBe('STOPPED')
    const response=await act('stop');expect(response.status()).toBe(202);pendingTask=(await response.json()).task.taskId
    await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${pendingTask}`)).json()).task.state,{timeout:15000}).toBe('WAITING_NODE')
    await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'任务中心',exact:true}).click()
    await expect(page.locator('.task-card').filter({hasText:pendingTask})).toContainText('等待节点')
  }finally{
    expect(birth()).toBe(identity);process.kill(pid,'SIGCONT')
    await expect.poll(async()=>(await allNodes(page)).find((item:any)=>item.nodeId===node.nodeId).state,{timeout:30000}).toBe('ONLINE')
    expect((await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance.runId).toBe(original)
    if(pendingTask){await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${pendingTask}`)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED');await expect(page.locator('.task-card').filter({hasText:pendingTask})).toContainText('成功')}
    else await act('stop')
    await expect.poll(async()=>(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance.state,{timeout:30000}).toBe('STOPPED')
  }
})
async function storageContains(page:Page,value:string){return page.evaluate(async value=>{
  const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
  const values=await new Promise<unknown[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
  return JSON.stringify({session:Object.entries(sessionStorage),local:Object.entries(localStorage),values}).includes(value)
},value)}

test('real node configuration keeps processes running during maintenance and tickets only leave by explicit download',async({page})=>{
  page.setDefaultTimeout(10000)
  await realLogin(page,fixture.admin)
  const instanceId=fixture.instanceIds[0]!,instance=(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance,node=(await allNodes(page)).find((node:{nodeId:string})=>node.nodeId===instance.nodeId),nextName=`浏览器维护节点 ${randomUUID().slice(0,8)}`
  const act=async(action:string)=>page.request.post(`/api/v1/instances/${instanceId}/actions`,{headers:await headers(page),data:{action}}),state=async()=>(await(await page.request.get(`/api/v1/instances/${instanceId}`)).json()).instance.state
  try{
    expect((await act('start')).status()).toBe(202);await expect.poll(state,{timeout:30000}).toBe('RUNNING')
    await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'节点管理',exact:true}).click();const card=page.locator(`[data-node-id="${node.nodeId}"]`);await card.getByRole('button',{name:'配置节点',exact:true}).click();await page.getByRole('textbox',{name:'节点名称',exact:true}).fill(nextName);await page.getByRole('spinbutton',{name:'节点实例配额'}).fill('1');await page.getByRole('checkbox',{name:'维护模式',exact:true}).check();await page.reload();await expect(page.getByRole('textbox',{name:'节点名称',exact:true})).toHaveValue(nextName);await expect(page.getByRole('checkbox',{name:'维护模式',exact:true})).toBeChecked();expect((await allNodes(page)).find((entry:{nodeId:string})=>entry.nodeId===node.nodeId).maintenance).toBe(false)
    await page.getByRole('button',{name:'保存节点配置',exact:true}).click();await expect(card).toContainText('MAINTENANCE');expect(await state()).toBe('RUNNING');expect((await act('restart')).status()).toBe(409);expect((await act('stop')).status()).toBe(202);await expect.poll(state,{timeout:30000}).toBe('STOPPED')
    await card.getByRole('button',{name:'配置节点',exact:true}).click();await page.getByRole('checkbox',{name:'维护模式',exact:true}).uncheck();await page.getByRole('button',{name:'保存节点配置',exact:true}).click();await expect(card).toContainText('ONLINE')
    const excess=await page.request.post('/api/v1/instances',{headers:await headers(page),data:{nodeId:node.nodeId,name:'超过测试配额的实例',config:{mode:'native',command:['/bin/true'],directory:'',stopSeconds:2,killSeconds:2,escalate:true}}});expect(excess.status()).toBe(409)
    await card.getByRole('button',{name:'轮换节点密钥',exact:true}).click();await page.reload();await expect(page.getByRole('dialog',{name:'节点密钥轮换'})).toContainText(node.nodeId)
    const rotationDownload=page.waitForEvent('download');await page.getByRole('button',{name:'生成并下载换钥票据',exact:true}).click();const rotation=await rotationDownload,privatePath=await rotation.path();expect(rotation.suggestedFilename()).toBe('blora-node-rotation.json');const ticket=JSON.parse(readFileSync(privatePath!,'utf8'));expect(ticket.nodeId===node.nodeId&&typeof ticket.ticket==='string'&&ticket.ticket.length===64).toBe(true);expect(await storageContains(page,ticket.ticket)).toBe(false);expect((await page.locator('body').innerText()).includes(ticket.ticket)).toBe(false);await rotation.delete()
    const before=(await allNodes(page)).length;await page.getByRole('button',{name:'登记新节点',exact:true}).click();await page.getByRole('textbox',{name:'登记节点名称'}).fill('浏览器下载登记验证');const enrollmentDownload=page.waitForEvent('download');await page.getByRole('button',{name:'生成并下载登记票据',exact:true}).click();const enrollment=await enrollmentDownload,secret=readFileSync((await enrollment.path())!,'utf8');expect(secret.length>0).toBe(true);expect(await storageContains(page,secret)).toBe(false);await enrollment.delete();expect((await allNodes(page)).length).toBe(before)
  }catch(error){
    console.error('Node maintenance primary failure:',error instanceof Error?error.message:String(error))
    throw error
  }finally{
    // Preserve the original failure instead of exhausting the test budget
    // while restoring task-owned resources.
    test.setTimeout(test.info().timeout+15000)
    const current=(await allNodes(page)).find((entry:{nodeId:string})=>entry.nodeId===node.nodeId);if(current)await page.request.patch(`/api/v1/nodes/${node.nodeId}`,{headers:await headers(page),data:{name:node.name,quota:node.quota,maintenance:node.maintenance,revision:current.configRevision}})
    if(await state()==='RUNNING'){await act('stop');await expect.poll(state,{timeout:30000}).toBe('STOPPED')}
  }
})
