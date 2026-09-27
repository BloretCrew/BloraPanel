import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect} from '@playwright/test'
import {realLogin} from './login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))

test('actual minute ticker accepts each backup slot once and captures the current source version',async({page})=>{
  test.setTimeout(240000)
  await realLogin(page,fixture.admin)
  const read=async(path:string)=>{const response=await page.request.get('/api/v1'+path);expect(response.status()).toBe(200);return response.json()}
  const headers=async()=>({'X-CSRF-Token':(await read('/session')).csrfToken,'Idempotency-Key':randomUUID()})
  const instance=(await read('/instances/'+fixture.instanceIds[0])).instance,base='/instances/'+instance.instanceId
  const source='wall-schedule-'+randomUUID()+'.txt'
  const waitTask=async(id:string)=>{await expect.poll(async()=>(await read('/tasks/'+id)).task.state,{timeout:30000,intervals:[500]}).toBe('SUCCEEDED');return(await read('/tasks/'+id)).task}
  const write=async(text:string,version:string)=>{
    const response=await page.request.put('/api/v1'+base+'/files/content',{headers:await headers(),data:{path:source,text,version}})
    expect(response.status()).toBe(202);await waitTask((await response.json()).task.taskId)
    return(await read(base+'/files/stat?path='+encodeURIComponent(source))).version as string
  }
  const firstVersion=await write('first real minute\n中文', 'missing')
  let scheduleId=''
  const tasks:any[]=[],versions=[firstVersion]
  try{
    const response=await page.request.post('/api/v1/schedules',{headers:await headers(),data:{revision:0,resource:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},action:'backup.create',args:{path:source,compression:'deflate',consistency:{mode:'files'}},cron:'* * * * *',timezone:'Asia/Shanghai',misfire:'run_once',overlap:'wait_one',enabled:true}})
    expect(response.status()).toBe(201)
    const created=(await response.json()).schedule;scheduleId=created.spec.id
    const current=async()=>(await read('/schedules')).items.find((item:any)=>item.spec.id===scheduleId)
    let previous=''
    for(let index=0;index<2;index++){
      await expect.poll(async()=>{const schedule=await current();return !!schedule.lastTaskId&&schedule.lastTaskId!==previous},{timeout:85000,intervals:[1000]}).toBe(true)
      const schedule=await current(),task=await waitTask(schedule.lastTaskId)
      expect(task.action).toBe('backup.create');expect(task.payload.scheduleId).toBe(scheduleId)
      expect(task.payload.create.version).toBe(versions[index])
      expect(task.payload.scheduleSlot).toBe(schedule.lastSlot)
      expect(task.requestId).toMatch(/^schedule:/)
      expect(Date.parse(schedule.lastSlot)%60000).toBe(0)
      const snapshot=task.result.snapshot
      expect(snapshot.state).toBe('ready');expect(snapshot.id).toBeTruthy()
      tasks.push(task);previous=task.taskId
      console.log('BLORA_WALL_SCHEDULE '+JSON.stringify({slot:schedule.lastSlot,taskId:task.taskId,snapshotId:snapshot.id}))
      if(index===0){versions.push(await write('second real minute\n修改后正文',firstVersion));await page.reload()}
    }
    expect(Date.parse(tasks[1].payload.scheduleSlot)-Date.parse(tasks[0].payload.scheduleSlot)).toBe(60000)
    expect(versions[1]).not.toBe(firstVersion)
    const schedule=await current()
    const removed=await page.request.delete('/api/v1/schedules/'+scheduleId,{headers:await headers(),data:{revision:schedule.configRevision}})
    expect(removed.status()).toBe(200)
    const accepted=(await read('/tasks')).items.filter((item:any)=>item.payload?.scheduleId===scheduleId)
    expect(accepted.map((item:any)=>item.taskId).sort()).toEqual(tasks.map(item=>item.taskId).sort())
    scheduleId=''
    for(let index=0;index<2;index++){
      const planResponse=await page.request.post('/api/v1'+base+'/restore-plans',{headers:await headers(),data:{backupId:tasks[index].result.snapshot.id,path:source+'.restored-'+index,version:'missing'}})
      expect(planResponse.status()).toBe(201)
      const plan=(await planResponse.json()).plan
      const restored=await page.request.post('/api/v1'+base+'/restores',{headers:await headers(),data:{planId:plan.request.id,planHash:plan.hash,overwritePlanHash:plan.hash}})
      expect(restored.status()).toBe(202);await waitTask((await restored.json()).task.taskId)
      expect((await read(base+'/files/content?path='+encodeURIComponent(source+'.restored-'+index))).text).toBe(index===0?'first real minute\n中文':'second real minute\n修改后正文')
    }
  }finally{
    if(scheduleId){const current=(await read('/schedules')).items.find((item:any)=>item.spec.id===scheduleId);if(current)expect((await page.request.delete('/api/v1/schedules/'+scheduleId,{headers:await headers(),data:{revision:current.configRevision}})).status()).toBe(200)}
  }
})
