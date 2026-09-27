import {readFileSync,readdirSync} from 'node:fs'
import {dirname,join,resolve} from 'node:path'
import {randomUUID} from 'node:crypto'
import {test,expect} from '@playwright/test'
import {realLogin} from './login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))

test('real node loss preserves metric timestamps and marks cached readings until reconnection',async({page})=>{
  test.setTimeout(150000)
  await realLogin(page,fixture.admin)
  const read=async(path:string)=>{const response=await page.request.get('/api/v1'+path);expect(response.status()).toBe(200);return response.json()}
  const instanceId=fixture.instanceIds[0],instance=(await read('/instances/'+instanceId)).instance
  const nodes=(await read('/nodes')).items,node=nodes.find((item:any)=>item.nodeId===instance.nodeId)
  const expected=join(dirname(resolve(process.env.BLORA_E2E_CREDENTIALS!)),node.name+'.json')
  const matches=readdirSync('/proc').filter(id=>/^\d+$/.test(id)).filter(id=>{try{const args=readFileSync(`/proc/${id}/cmdline`,'utf8').split('\0');return args[0]?.endsWith('blora-daemon')&&args[1]==='--config'&&args[2]===expected}catch{return false}})
  expect(matches).toHaveLength(1)
  const pid=Number(matches[0]),birth=()=>readFileSync(`/proc/${pid}/stat`,'utf8').split(') ').at(-1)!.split(' ')[19],identity=birth()
  const act=async(action:string)=>{
    const csrf=(await read('/session')).csrfToken
    const response=await page.request.post(`/api/v1/instances/${instanceId}/actions`,{headers:{'X-CSRF-Token':csrf,'Idempotency-Key':randomUUID()},data:{action}})
    expect(response.status()).toBe(202)
    const task=(await response.json()).task
    await expect.poll(async()=>(await read('/tasks/'+task.taskId)).task.state,{timeout:30000}).toBe('SUCCEEDED')
  }
  let paused=false,started=false
  try{
    expect(instance.state).toBe('STOPPED');await act('start');started=true
    const run=(await read('/instances/'+instanceId)).instance.runId
    const initial=await read(`/instances/${instanceId}/metrics`)
    expect(initial.runId).toBe(run);expect(initial.stale).toBeFalsy()
    await page.locator('[data-app="blora.instances"]').click()
    await page.locator('.app-window.focused').getByRole('button',{name:instance.name,exact:true}).click()
    const instanceWindow=page.locator('.app-window').filter({has:page.getByRole('heading',{name:instance.name+' RUNNING',exact:true})})
    await instanceWindow.getByRole('button',{name:'监控',exact:true}).click()
    await expect(instanceWindow.locator('.monitor-summary')).toContainText('采集于')
    await page.locator('.launcher-button').click()
    await page.locator('.launcher').getByRole('button',{name:'监控与进程',exact:true}).click()
    const nodeWindow=page.locator('.app-window.focused')
    await nodeWindow.getByRole('combobox',{name:'监控节点'}).selectOption(node.nodeId)
    await expect(nodeWindow.locator('.monitor-summary')).toContainText('采集于')
    // Ensure the process page has a successful live result to cache before
    // disconnecting the node; otherwise the correct offline state is simply
    // "query unavailable" rather than a stale cached list.
    await expect(nodeWindow.getByText('实时列表 · PID 0 之后',{exact:true})).toBeVisible({timeout:20000})
    await expect(nodeWindow.getByText(/最近成功读取于/)).toBeVisible({timeout:20000})
    expect(await nodeWindow.getByRole('button',{name:'请求终止',exact:true}).count()).toBeGreaterThan(0)
    const before=await read(`/nodes/${node.nodeId}/metrics`)
    expect(before.stale).toBeFalsy()
    expect(birth()).toBe(identity);const pausedAt=Date.now();process.kill(pid,'SIGSTOP');paused=true
    await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:65000}).toBe('OFFLINE')
    const cachedNode=await read(`/nodes/${node.nodeId}/metrics`),cachedInstance=await read(`/instances/${instanceId}/metrics`)
    for(const point of [cachedNode,cachedInstance]){expect(point.stale).toBe(true);expect(Date.parse(point.observedAt)).toBeLessThanOrEqual(pausedAt);expect(point.diagnostic).toContain('采样')}
    expect(cachedInstance.runId).toBe(run)
    expect((await read(`/nodes/${node.nodeId}/metrics`)).observedAt).toBe(cachedNode.observedAt)
    expect((await read(`/instances/${instanceId}/metrics`)).observedAt).toBe(cachedInstance.observedAt)
    await expect(nodeWindow.locator('.monitor-summary')).toContainText('旧数据（节点失联，禁止视为实时成功）',{timeout:20000})
    await expect(nodeWindow.locator('.monitor-summary')).toContainText(cachedNode.diagnostic)
    // A cached process page must not still advertise itself as a live list.
    await expect(nodeWindow.getByText('实时列表 · PID 0 之后',{exact:true})).not.toBeVisible()
    await expect(nodeWindow.getByText('旧进程列表（当前查询不可用） · PID 0 之后',{exact:true})).toBeVisible()
    await expect(nodeWindow.getByText(/最近成功读取于/)).toBeVisible()
    expect(await nodeWindow.getByRole('button',{name:'请求终止',exact:true}).count()).toBeGreaterThan(0)
    for(const button of await nodeWindow.getByRole('button',{name:'请求终止',exact:true}).all())await expect(button).toBeDisabled()
    await expect(instanceWindow.locator('.monitor-summary')).toContainText('旧数据（当前采样不可用）',{timeout:20000})
    await expect(instanceWindow.locator('.monitor-summary')).toContainText(cachedInstance.diagnostic)
    const other=nodes.find((item:any)=>item.nodeId!==node.nodeId)
    expect((await read(`/nodes/${other.nodeId}/metrics`)).stale).toBeFalsy()
    expect(birth()).toBe(identity);process.kill(pid,'SIGCONT');paused=false
    await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:30000}).toBe('ONLINE')
    await expect.poll(async()=>{const point=await read(`/nodes/${node.nodeId}/metrics`);return !point.stale&&Date.parse(point.observedAt)>pausedAt},{timeout:15000}).toBe(true)
    await expect(nodeWindow.locator('.monitor-summary')).not.toContainText('旧数据',{timeout:20000})
    await expect(nodeWindow.getByText('实时列表 · PID 0 之后',{exact:true})).toBeVisible({timeout:20000})
    await expect(instanceWindow.locator('.monitor-summary')).not.toContainText('旧数据',{timeout:20000})
    expect((await read(`/instances/${instanceId}/metrics`)).runId).toBe(run)
  }finally{
    test.setTimeout(test.info().timeout+45000)
    if(paused){expect(birth()).toBe(identity);process.kill(pid,'SIGCONT')}
    if(started){await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:30000}).toBe('ONLINE');await act('stop')}
  }
})
