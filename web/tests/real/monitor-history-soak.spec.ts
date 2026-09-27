import {readFileSync,readdirSync} from 'node:fs'
import {dirname,join,resolve} from 'node:path'
import {test,expect} from '@playwright/test'
import {realLogin} from './login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))

test('real node sampling retains exactly the newest 120 points over sustained polling',async({page})=>{
  test.skip(process.platform!=='linux'||process.env.BLORA_E01_HISTORY_SOAK!=='1','requires Linux; set BLORA_E01_HISTORY_SOAK=1 for the opt-in 121-point real-node soak')
  const sampleCount=121
  const intervalMs=Number(process.env.BLORA_E01_SAMPLE_INTERVAL_MS||5000)
  expect(Number.isFinite(intervalMs)&&intervalMs>=1000&&intervalMs<=30000).toBe(true)
  test.setTimeout(sampleCount*intervalMs+120000)
  await realLogin(page,fixture.admin)
  const read=async(path:string)=>{const response=await page.request.get('/api/v1'+path);expect(response.status()).toBe(200);return response.json()}
  const node=(await read('/nodes')).items.find((item:any)=>item.state==='ONLINE')
  expect(node,'a real online node is required').toBeTruthy()
  const expected=join(dirname(resolve(process.env.BLORA_E2E_CREDENTIALS!)),node.name+'.json')
  const matches=readdirSync('/proc').filter(id=>/^\d+$/.test(id)).filter(id=>{try{const args=readFileSync(`/proc/${id}/cmdline`,'utf8').split('\0');return args[0]?.endsWith('blora-daemon')&&args[1]==='--config'&&args[2]===expected}catch{return false}})
  expect(matches,'the fixture-owned daemon config must resolve to one process').toHaveLength(1)
  const pid=Number(matches[0]),birth=()=>readFileSync(`/proc/${pid}/stat`,'utf8').split(') ').at(-1)!.split(' ')[19],identity=birth()
  const timestamps:string[]=[]
  let previous=0,paused=false
  try{
    for(let index=0;index<sampleCount;index++){
      if(index>0)await page.waitForTimeout(intervalMs)
      const point=await read(`/nodes/${encodeURIComponent(node.nodeId)}/metrics`)
      expect(point.stale).toBeFalsy()
      const observedAt=Date.parse(point.observedAt)
      expect(Number.isFinite(observedAt)&&observedAt>previous).toBe(true)
      previous=observedAt
      timestamps.push(point.observedAt)
    }
    const history=await read(`/nodes/${encodeURIComponent(node.nodeId)}/metrics/history`)
    expect(history.retention).toBe(120)
    expect(history.items).toHaveLength(120)
    expect(history.items[0].observedAt).toBe(timestamps[1])
    expect(history.items.at(-1).observedAt).toBe(timestamps.at(-1))
    for(let index=1;index<history.items.length;index++){
      expect(Date.parse(history.items[index].observedAt)).toBeGreaterThan(Date.parse(history.items[index-1].observedAt))
    }

    expect(birth()).toBe(identity)
    process.kill(pid,'SIGSTOP');paused=true
    await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:65000}).toBe('OFFLINE')
    const staleHistory=await read(`/nodes/${encodeURIComponent(node.nodeId)}/metrics/history`)
    expect(staleHistory.stale).toBe(true)
    expect(staleHistory.retention).toBe(120)
    expect(staleHistory.items).toHaveLength(120)
    expect(staleHistory.diagnostic).toContain('节点当前无法连接')
    for(let index=0;index<staleHistory.items.length;index++){
      expect(staleHistory.items[index].stale).toBe(true)
      expect(staleHistory.items[index].observedAt).toBe(timestamps[index+1])
    }

    const fixtureRoot=dirname(resolve(process.env.BLORA_E2E_CREDENTIALS!)),stateDir=join(fixtureRoot,'master'),fixtureHost=new URL(fixture.url).host
    const findMasterPIDs=()=>readdirSync('/proc').filter(id=>/^\d+$/.test(id)).map(Number).filter(candidate=>{try{const args=readFileSync(`/proc/${candidate}/cmdline`,'utf8').split('\0'),index=args.indexOf('--state-dir');return args[0]?.endsWith('blora-master')&&index>=0&&args[index+1]===stateDir}catch{return false}})
    const findSupervisorPIDs=()=>readdirSync('/proc').filter(id=>/^\d+$/.test(id)).map(Number).filter(candidate=>{try{const args=readFileSync(`/proc/${candidate}/cmdline`,'utf8').split('\0'),index=args.findIndex(arg=>arg==='-listen'||arg==='--listen');return args[0]?.endsWith('blora-devfixture')&&index>=0&&args[index+1]===fixtureHost}catch{return false}})
    const masterPIDs=findMasterPIDs(),supervisorPIDs=findSupervisorPIDs()
    expect(masterPIDs,'the fixture Master must be identified by its private state directory').toHaveLength(1)
    expect(supervisorPIDs,'the fixture supervisor must own the restart operation').toHaveLength(1)
    const previousMasterPID=masterPIDs[0]
    process.kill(supervisorPIDs[0],'SIGUSR1')
    await expect.poll(async()=>{
      const current=findMasterPIDs()
      if(current.length!==1||current[0]===previousMasterPID)return false
      const health=await page.request.get('/healthz',{timeout:1000}).catch(()=>undefined)
      return health?.status()===200
    },{timeout:20000}).toBe(true)
    const relogin=await page.request.post('/api/v1/login',{data:fixture.admin})
    expect(relogin.status()).toBe(200)
    await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:10000}).toBe('OFFLINE')
    const afterMasterRestart=await read(`/nodes/${encodeURIComponent(node.nodeId)}/metrics/history`)
    expect(afterMasterRestart.stale).toBe(true)
    expect(afterMasterRestart.items).toHaveLength(120)
    for(let index=0;index<afterMasterRestart.items.length;index++){
      expect(afterMasterRestart.items[index].stale).toBe(true)
      expect(afterMasterRestart.items[index].observedAt).toBe(timestamps[index+1])
    }
    expect(birth()).toBe(identity)
    process.kill(pid,'SIGCONT');paused=false
    await expect.poll(async()=>(await read('/nodes')).items.find((item:any)=>item.nodeId===node.nodeId).state,{timeout:30000}).toBe('ONLINE')
  }finally{
    if(paused){expect(birth()).toBe(identity);process.kill(pid,'SIGCONT');paused=false}
  }
})
