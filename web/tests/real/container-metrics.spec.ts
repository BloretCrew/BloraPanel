import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {realLogin} from './login'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
async function headers(page:Page){return {'X-CSRF-Token':(await(await page.request.get('/api/v1/session',{timeout:10000})).json()).csrfToken,'Idempotency-Key':randomUUID()}}
test('isolated instance metrics respect resource grants and render in the instance monitor tab',async({page,browser})=>{
  test.setTimeout(120000)
  await realLogin(page,fixture.admin)
  const name='metrics-'+randomUUID().slice(0,8),nodeId=fixture.nodeIds[0]
  const created=await page.request.post('/api/v1/instances',{headers:await headers(page),data:{nodeId,name,config:{mode:'container',uid:65534,gid:65534,image:'blora-isolated-e2e:20260909',command:['/bin/sh','-c','sleep 300'],memoryBytes:134217728,stopSeconds:1,killSeconds:2,escalate:true}}})
  expect(created.status(),await created.text()).toBe(201)
  const instance=(await created.json()).instance,id=instance.instanceId
  async function action(action:string){const response=await page.request.post(`/api/v1/instances/${id}/actions`,{headers:await headers(page),data:{action}});expect(response.status()).toBe(202);const task=(await response.json()).task.taskId;await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${task}`)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')}
  const memberContext=await browser.newContext({baseURL:fixture.url,ignoreHTTPSErrors:true}),member=await memberContext.newPage()
  let stage='start'
  try{
    await action('start')
    const response=await page.request.get(`/api/v1/instances/${id}/metrics`);expect(response.status()).toBe(200)
    const point=await response.json();expect(point.runId).toBeTruthy();expect(point.memoryUsedBytes).toBeGreaterThan(0);expect(point.memoryTotalBytes).toBe(134217728);expect(point.unavailable).not.toContain('cpu');expect(point.stale).toBe(false)
    stage='member-login'
    const login=await member.request.post('/api/v1/login',{data:fixture.member,timeout:15000})
    expect(login.status()).toBe(200)
    await member.goto('/')
    await expect(member.locator('.desktop')).toBeVisible()
    expect((await member.request.get(`/api/v1/instances/${id}/metrics`)).status()).toBe(403)
    const users=(await(await page.request.get('/api/v1/users')).json()).items,user=users.find((u:any)=>u.name===fixture.member.name)
    const grant={userId:user.userId,resource:{kind:'instance',id,nodeId},action:'instance.read'}
    const granted=await page.request.post('/api/v1/grants',{headers:await headers(page),data:grant});expect(granted.status()).toBe(200)
    expect((await member.request.get(`/api/v1/instances/${id}/metrics`)).status()).toBe(200)
    expect((await member.request.get(`/api/v1/nodes/${nodeId}/metrics`)).status()).toBe(403)
    const hostRequests:string[]=[]
    member.on('request',request=>{if(new RegExp(`/nodes/${nodeId}/(metrics|processes)`).test(request.url()))hostRequests.push(request.url())})
    stage='open-monitor';await member.reload()
    await member.locator('.launcher-button').click();await member.locator('.launcher').getByRole('button',{name:'实例中心',exact:true}).click()
    await member.locator('.instance-card').filter({hasText:name}).getByRole('button',{name,exact:true}).click()
    await member.getByRole('button',{name:'监控',exact:true}).click()
    await expect(member.getByText('实例指标',{exact:true})).toBeVisible()
    await expect(member.getByText('容器内存使用量包含文件缓存；内存上限 128 MiB。')).toBeVisible()
    await member.reload()
    await expect(member.getByText('容器内存使用量包含文件缓存；内存上限 128 MiB。')).toBeVisible()
    expect(hostRequests).toEqual([])
    await expect(member.getByText('网络接收',{exact:true})).toBeVisible()
    await expect(member.locator('dt').filter({hasText:'网络接收'}).locator('xpath=following-sibling::dd[1]')).toHaveText(point.unavailable?.includes('network')?'不可用':(point.networkRxBytes/1048576).toFixed(2)+' MiB（累计）')
    stage='visible-poll';await member.waitForResponse(response=>response.url().endsWith(`/instances/${id}/metrics`)&&response.status()===200,{timeout:15000})
    let metricRequests=0
    member.on('request',request=>{if(request.url().endsWith(`/instances/${id}/metrics`))metricRequests++})
    const window=member.locator('.app-window').filter({has:member.getByText('实例指标',{exact:true})})
    stage='minimize';await window.getByRole('button',{name:'最小化窗口',exact:true}).click()
    await expect(window).toBeHidden()
    await member.waitForTimeout(7000)
    expect(metricRequests).toBe(0)
    stage='restore'
    await Promise.all([
      member.waitForResponse(response=>response.url().endsWith(`/instances/${id}/metrics`)&&response.status()===200,{timeout:10000}),
      (async()=>{await member.locator('.taskbar-app').filter({hasText:'实例中心'}).click();const picker=member.locator('.window-picker');if(await picker.isVisible())await picker.getByRole('button').filter({hasText:name}).click({timeout:5000})})(),
    ])
    await expect(window).toBeVisible()
    stage='stop';await action('stop')
    const cached=await(await page.request.get(`/api/v1/instances/${id}/metrics`)).json();expect(cached.stale).toBe(true);expect(cached.runId).toBe(point.runId)
    await expect(member.getByText('旧数据（当前采样不可用）',{exact:false})).toBeVisible({timeout:15000})
    const revoked=await page.request.delete('/api/v1/grants',{headers:await headers(page),data:grant});expect(revoked.status()).toBe(200)
    expect((await member.request.get(`/api/v1/instances/${id}/metrics`)).status()).toBe(403)
  }catch(error){console.error('container metrics stage:',stage,String(error));throw error}finally{test.setTimeout(test.info().timeout+45000);await memberContext.close();await action('kill');const removed=await page.request.delete(`/api/v1/instances/${id}`,{headers:await headers(page)});expect([200,404]).toContain(removed.status())}
})
