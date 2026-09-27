import {readFileSync} from 'node:fs'
import {cpus,totalmem} from 'node:os'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {realLogin} from './login'
import {decodeEnvelope,MessageType} from '../../src/services/protocol'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
async function headers(page:Page){return {'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
async function openResource(page:Page,section:'文件'|'控制台'){
  const instance=(await(await page.request.get('/api/v1/instances')).json()).items.find((item:{instanceId:string})=>item.instanceId===fixture.instanceIds[0])
  await page.locator('[data-app="blora.instances"]').click()
  const win=page.locator('.app-window.focused')
  const entry=win.getByRole('button',{name:instance.name,exact:true})
  await expect(entry.or(win.getByRole('button',{name:section,exact:true}))).toBeVisible()
  if(await entry.count())await entry.click()
  await win.getByRole('button',{name:section,exact:true}).click()
  await win.getByRole('button',{name:section==='文件'?'在文件管理器打开':'打开终端会话管理',exact:true}).click()
}

test('real eight-window mixed load measures pointer response with two PTYs and verified transfer',async({page,browser},info)=>{
  test.setTimeout(300000)
  page.setDefaultTimeout(15000)
  expect(fixture.performance,'Run devfixture with --performance; missing real data is not a passing probe').toBe(true)
  const configuredSoak=Number(process.env.BLORA_PERF_SOAK_SECONDS||0)
  const soakSeconds=Number.isFinite(configuredSoak)&&configuredSoak>0?Math.min(3600,Math.floor(configuredSoak)):0
  const outputDelay=Number(process.env.BLORA_PERF_OUTPUT_DELAY_SECONDS||0.01)
  expect(outputDelay).toBeGreaterThanOrEqual(0.001)
  expect(outputDelay).toBeLessThanOrEqual(1)
  const outputIterations=Math.max(12000,Math.ceil((soakSeconds+120)/outputDelay))
  const bytes=new Map<string,number>(),errors:string[]=[]
  const terminalDiagnostics=async()=>page.locator('.terminal-container').evaluateAll(elements=>elements.map(element=>{
    const app=element.closest('.app-window')
    return {sessionId:element.getAttribute('data-session-id'),status:app?.querySelector('.editor-toolbar strong')?.textContent,errors:[...app?.querySelectorAll('[role="alert"]')||[]].map(alert=>alert.textContent)}
  }))
  let maxUnacknowledgedBytes=0
  page.on('pageerror',error=>errors.push(error.message))
  page.on('websocket',socket=>{if(!socket.url().includes('/terminals/'))return;let received=0,acknowledged=0;socket.on('framesent',event=>{
    if(typeof event.payload==='string')return
    const frame=decodeEnvelope(new Uint8Array(event.payload));if(frame.type===MessageType.Ack)acknowledged=frame.sequence||0
  });socket.on('framereceived',event=>{
    if(typeof event.payload==='string')return
    const frame=decodeEnvelope(new Uint8Array(event.payload))
    if(frame.type===MessageType.Data){bytes.set(socket.url(),(bytes.get(socket.url())||0)+(frame.payload?.length||0));received=frame.sequence||0;maxUnacknowledgedBytes=Math.max(maxUnacknowledgedBytes,received-acknowledged)}
  })})
  await realLogin(page,fixture.admin)
  const renderExperiment=process.env.BLORA_PERF_RENDER_EXPERIMENT||'baseline'
  expect(['baseline','contain','layers','glass-layer','no-glass']).toContain(renderExperiment)
  // Diagnostic comparisons only: each run keeps the same real resources,
  // terminal output, interaction sampling and acceptance threshold.
  if(renderExperiment!=='baseline')await page.addStyleTag({content:renderExperiment==='contain'
    ?'.terminal-container .xterm-rows{contain:layout style}'
    :renderExperiment==='layers'?'.app-window{will-change:transform}.terminal-container{contain:layout paint style}'
    :renderExperiment==='glass-layer'?':root[data-translucency="on"] .app-window{backdrop-filter:none!important;background:transparent!important}:root[data-translucency="on"] .app-window::before{content:"";position:absolute;inset:0;border-radius:inherit;z-index:-1;pointer-events:none;background:var(--glass-window);backdrop-filter:blur(24px) saturate(1.05)}:root[data-theme="dark"][data-translucency="on"] .app-window::before{backdrop-filter:blur(20px) saturate(1.03)}'
    :'*,*::before,*::after{backdrop-filter:none!important;-webkit-backdrop-filter:none!important}'})
  await openResource(page,'文件')
  await page.getByRole('textbox',{name:'目录路径',exact:true}).fill('ten-thousand')
  await page.getByRole('button',{name:'转到',exact:true}).click()
  await expect(page.getByRole('grid',{name:'节点文件列表'})).toHaveAttribute('aria-rowcount','10000',{timeout:30000})
  expect(await page.locator('.file-row').count()).toBeLessThanOrEqual(300)
  await openResource(page,'控制台')
  await page.locator('.app-window.focused').getByRole('button',{name:'新建会话',exact:true}).click()
  await expect(page.locator('.app-window.focused').getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  const startOutput=async(terminal:ReturnType<Page['locator']>)=>{
    await terminal.locator('.xterm-helper-textarea').focus()
    await page.keyboard.type("i=0; while [ $i -lt "+outputIterations+" ]; do printf '%01024d\\n' $i; i=$((i+1)); sleep "+outputDelay+"; done\n")
  }
  await startOutput(page.locator('.app-window.focused .terminal-container[data-terminal-writable="true"]'))
  await page.locator('.app-window.focused').getByRole('button',{name:'同会话新窗口',exact:true}).click()
  await page.locator('.app-window.focused').getByRole('button',{name:'新建会话',exact:true}).click()
  await expect(page.locator('.app-window.focused').getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  await startOutput(page.locator('.app-window.focused .terminal-container[data-terminal-writable="true"]'))
  const terminals=page.locator('.terminal-container')
  const terminalSessionIds=[...new Set((await terminals.evaluateAll(elements=>elements.map(element=>element.getAttribute('data-session-id')).filter((id):id is string=>!!id))))]
  expect(terminalSessionIds).toHaveLength(2)
  let index=0
  while(await page.locator('.app-window').count()<8){
    await page.locator('.launcher-button').click()
    await page.locator('.launcher-grid > button').nth(index++).click()
    expect(index).toBeLessThan(20)
  }
  const source=await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}/files/stat?path=transfer-load.bin`)
  expect(source.status()).toBe(200)
  const stat=await source.json()
  const transferTimeout=Math.max(240000,Math.ceil(stat.size/(512*1024))*1000)
  test.setTimeout(transferTimeout+120000+soakSeconds*1000)
  let transferStart=Date.now(),taskId=process.env.BLORA_PERF_TRANSFER_TASK
  let targetPath=`performance-${randomUUID()}.bin`,completedTransfers=0,activeTransferSamples=0,maxTransferRestartMs=0
  const startTransfer=async(version:string)=>{
    const accepted=await page.request.post('/api/v1/transfers',{headers:await headers(page),data:{source:{instanceId:fixture.instanceIds[0],path:'transfer-load.bin',version:stat.version},target:{instanceId:fixture.instanceIds[1],path:targetPath,version},move:false}})
    expect(accepted.status(),await accepted.text()).toBe(202)
    return (await accepted.json()).task.taskId as string
  }
  if(taskId){
    const response=await page.request.get(`/api/v1/transfers/${encodeURIComponent(taskId)}`);expect(response.status()).toBe(200)
    const {task}=await response.json()
    expect(task.action).toBe('transfer.copy')
    expect(task.payload.source).toEqual({instanceId:fixture.instanceIds[0],path:'transfer-load.bin',version:stat.version})
    expect(task.payload.target.instanceId).toBe(fixture.instanceIds[1])
    expect(task.payload.target.path).toMatch(/^performance-[a-f0-9-]+\.bin$/)
    targetPath=task.payload.target.path
    transferStart=Date.parse(task.createdAt);expect(Number.isFinite(transferStart)).toBe(true)
  }else{
    taskId=await startTransfer('missing')
  }
  const maintainTransfer=async()=>{
    const response=await page.request.get(`/api/v1/transfers/${taskId}`)
    expect(response.status()).toBe(200)
    const current=await response.json()
    expect(['RUNNING','SUCCEEDED']).toContain(current.task.state)
    if(current.task.state==='SUCCEEDED'){
      expect(current.transfer.destinationVerified).toBe(true)
      completedTransfers++
      const restartAt=Date.now()
      const target=await page.request.get(`/api/v1/instances/${fixture.instanceIds[1]}/files/stat?path=${encodeURIComponent(targetPath)}`)
      expect(target.status()).toBe(200)
      taskId=await startTransfer((await target.json()).version)
      await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:30000}).toBe('RUNNING')
      maxTransferRestartMs=Math.max(maxTransferRestartMs,Date.now()-restartAt)
    }
    activeTransferSamples++
  }
  await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:30000}).toBe('RUNNING')
  console.log('BLORA_PERF_RENDER_EXPERIMENT '+renderExperiment)
  const beforeBytes=new Map(bytes),sampleStart=Date.now()
  const profileSoak=!!process.env.BLORA_PERF_PROFILE_SOAK
  const perfDiagnostics=!!process.env.BLORA_PERF_DIAGNOSTICS
  const profiler=process.env.BLORA_PERF_PROFILE||profileSoak?await page.context().newCDPSession(page):undefined
  if(profiler){await profiler.send('Profiler.enable');await profiler.send('Profiler.start')}
  await page.evaluate((diagnostics:boolean)=>{
    const probe={latencies:[] as number[],frames:[] as number[],slowPointers:[] as Array<{latencyMs:number;inputDelayMs:number;frameAfterHandlerMs:number;eventTime:number;x:number;y:number;buttons:number}>,longTasks:[] as Array<{startTime:number;duration:number;name:string}>,longTaskObserver:undefined as PerformanceObserver|undefined,storageWrites:{count:0,totalChars:0,totalMs:0,maxChars:0,maxMs:0,durations:[] as number[]},restoreStorage:undefined as (()=>void)|undefined,running:true,last:performance.now()}
    ;(window as any).__bloraPerf=probe
    if(diagnostics)try{
      probe.longTaskObserver=new PerformanceObserver(list=>{
        for(const entry of list.getEntries()){
          probe.longTasks.push({startTime:entry.startTime,duration:entry.duration,name:entry.name})
          if(probe.longTasks.length>2000)probe.longTasks.shift()
        }
      })
      probe.longTaskObserver.observe({entryTypes:['longtask']})
      const originalSetItem=Storage.prototype.setItem
      Storage.prototype.setItem=function(key,value){
        if(!key.startsWith('blora:tail:'))return originalSetItem.call(this,key,value)
        const started=performance.now()
        try{return originalSetItem.call(this,key,value)}finally{
          const duration=performance.now()-started
          probe.storageWrites.count++;probe.storageWrites.totalChars+=value.length
          probe.storageWrites.totalMs+=duration;probe.storageWrites.maxChars=Math.max(probe.storageWrites.maxChars,value.length)
          probe.storageWrites.maxMs=Math.max(probe.storageWrites.maxMs,duration);probe.storageWrites.durations.push(duration)
        }
      }
      probe.restoreStorage=()=>{Storage.prototype.setItem=originalSetItem}
    }catch{probe.longTaskObserver?.disconnect();probe.longTaskObserver=undefined}
    window.addEventListener('pointermove',event=>{
      if(event.buttons!==1||!probe.running)return
      if(!diagnostics){const start=event.timeStamp;requestAnimationFrame(()=>requestAnimationFrame(()=>probe.latencies.push(performance.now()-start)));return}
      const start=event.timeStamp,handlerTime=performance.now(),inputDelayMs=Math.max(0,handlerTime-start),x=event.clientX,y=event.clientY,buttons=event.buttons
      requestAnimationFrame(()=>requestAnimationFrame(()=>{
        const latencyMs=performance.now()-start
        probe.latencies.push(latencyMs)
        if(latencyMs>=50){
          probe.slowPointers.push({latencyMs,inputDelayMs,frameAfterHandlerMs:performance.now()-handlerTime,eventTime:start,x,y,buttons})
          probe.slowPointers.sort((a,b)=>b.latencyMs-a.latencyMs)
          if(probe.slowPointers.length>20)probe.slowPointers.length=20
        }
      }))
    },true)
    function frame(now:number){if(!probe.running)return;probe.frames.push(now-probe.last);probe.last=now;requestAnimationFrame(frame)}
    requestAnimationFrame(frame)
  },perfDiagnostics)
  const win=page.locator('.app-window.focused'),title=await win.locator('.window-titlebar').boundingBox()
  const original=await win.evaluate(element=>(element as HTMLElement).style.transform)
  await page.mouse.move(title!.x+240,title!.y+15);await page.mouse.down()
  for(let i=0;i<120;i++)await page.mouse.move(title!.x+240+Math.sin(i/10)*80,title!.y+15+Math.cos(i/10)*30)
  await page.mouse.up()
  const finishProfile=async()=>{if(profiler){
    const {profile}=await profiler.send('Profiler.stop'),names=new Map(profile.nodes.map((node:any)=>{
      const frame=node.callFrame;let source=frame.url||''
      try{const parsed=new URL(source);source=parsed.pathname||parsed.href}catch{}
      return [node.id,`${frame.functionName||'(anonymous)'} ${source}:${(frame.lineNumber||0)+1}:${(frame.columnNumber||0)+1}`]
    })),durations=new Map<string,number>()
    profile.samples?.forEach((id:number,index:number)=>{const name=String(names.get(id));durations.set(name,(durations.get(name)||0)+(profile.timeDeltas?.[index]||0)/1000)})
    console.log('BLORA_PERF_CPU '+JSON.stringify([...durations].sort((a,b)=>b[1]-a[1]).slice(0,30).map(([frame,durationMs])=>({frame,durationMs}))))
    await profiler.detach()
  }}
  if(!profileSoak)await finishProfile()
  expect(await win.evaluate(element=>(element as HTMLElement).style.transform)).not.toBe(original)
  const soakHeap:number[]=[]
  if(soakSeconds>0){
    const soakUntil=Date.now()+soakSeconds*1000
    let lastProgress=Date.now()
    let lastTerminalBytes=new Map(bytes)
    while(Date.now()<soakUntil){
      // Keep a real, verified transfer workload throughout long runs. Reuse
      // only this test's destination with its exact version to bound disk use.
      await maintainTransfer()
      const diagnostics=await terminalDiagnostics()
      if(diagnostics.some(terminal=>terminal.errors.length)){
        console.log('BLORA_PERF_TERMINAL_FAILURE '+JSON.stringify({diagnostics,pageErrors:errors}))
        await info.attach('terminal-failure.json',{body:JSON.stringify({diagnostics,pageErrors:errors},null,2),contentType:'application/json'})
        expect(diagnostics.flatMap(terminal=>terminal.errors),'Live terminals must not fail between progress samples').toEqual([])
      }
      const box=await win.locator('.window-titlebar').boundingBox()
      if(box){
        const x=box.x+Math.min(240,Math.max(40,box.width-40)),y=box.y+15
        await page.mouse.move(x,y);await page.mouse.down()
        for(let i=0;i<24;i++)await page.mouse.move(x+Math.sin(i/4)*24,y+Math.cos(i/4)*12)
        await page.mouse.up()
      }
      const heap=await page.evaluate(()=>Number((performance as any).memory?.usedJSHeapSize||0))
      if(heap>0)soakHeap.push(heap)
      if(Date.now()-lastProgress>=60000){
        const terminalIntervalBytes=[...bytes].map(([key,value])=>value-(lastTerminalBytes.get(key)||0)).filter(value=>value>0)
        console.log('BLORA_PERF_PROGRESS '+JSON.stringify({elapsedSeconds:Math.round((Date.now()-sampleStart)/1000),completedTransfers,activeTransferSamples,heapBytes:heap,maxUnacknowledgedBytes,terminalIntervalBytes}))
        if(terminalIntervalBytes.length!==2){
          // Preserve product diagnostics before Playwright disposes the page.
          // Do not capture account fields, terminal contents, or full storage.
          const diagnostics=await terminalDiagnostics()
          console.log('BLORA_PERF_TERMINAL_FAILURE '+JSON.stringify({diagnostics,pageErrors:errors}))
          await info.attach('terminal-failure.json',{body:JSON.stringify({diagnostics,pageErrors:errors},null,2),contentType:'application/json'})
        }
        expect(terminalIntervalBytes,'Both real PTYs must keep producing throughout the soak').toHaveLength(2)
        lastTerminalBytes=new Map(bytes)
        lastProgress=Date.now()
      }
      await page.waitForTimeout(250)
    }
  }
  if(profileSoak)await finishProfile()
  if(soakSeconds>0)await maintainTransfer()
  const stateAfter=(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state
  const sampleSeconds=(Date.now()-sampleStart)/1000
  const terminalBytesPerSecond=[...bytes].map(([key,value])=>(value-(beforeBytes.get(key)||0))/sampleSeconds).filter(value=>value>0)
  const sample=await page.evaluate(async(diagnostics:boolean)=>{
    const probe=(window as any).__bloraPerf;probe.running=false;probe.longTaskObserver?.disconnect()
    probe.restoreStorage?.()
    const rtts:number[]=[]
    for(let i=0;i<5;i++){const start=performance.now();const response=await fetch('/api/v1/session');if(!response.ok)throw Error('RTT request failed');await response.json();rtts.push(performance.now()-start)}
    const p95=(values:number[])=>values.sort((a,b)=>a-b)[Math.ceil(values.length*.95)-1]||0
    const writes=probe.storageWrites
    return {samples:probe.latencies.length,p95:p95(probe.latencies),max:Math.max(...probe.latencies),frames:probe.frames.length,longFrames:probe.frames.filter((value:number)=>value>50).length,...(diagnostics?{slowPointerResponses:probe.slowPointers,longTasks:probe.longTasks,recoveryJournalWrites:{count:writes.count,totalChars:writes.totalChars,totalMs:writes.totalMs,maxChars:writes.maxChars,maxMs:writes.maxMs,p95Ms:p95(writes.durations)}}:{}),heapBytes:(performance as any).memory?.usedJSHeapSize,apiRttP95:p95(rtts)}
  },perfDiagnostics)
  const soak={seconds:soakSeconds,outputDelaySeconds:outputDelay,completedTransfers,activeTransferSamples,maxTransferRestartMs,heapSamples:soakHeap.length,heapMin:soakHeap.length?Math.min(...soakHeap):undefined,heapMax:soakHeap.length?Math.max(...soakHeap):undefined,heapDelta:soakHeap.length?Math.max(...soakHeap)-Math.min(...soakHeap):undefined}
  console.log('BLORA_REAL_PERF_SAMPLE '+JSON.stringify({...sample,terminalBytesPerSecond,stateAfter,soak}))
  await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:transferTimeout}).toBe('SUCCEEDED')
  const detail=await(await page.request.get(`/api/v1/transfers/${taskId}`)).json()
  expect(detail.transfer.destinationVerified).toBe(true)
  const graphics=await page.evaluate(()=>{const gl=document.createElement('canvas').getContext('webgl2'),extension=gl?.getExtension('WEBGL_debug_renderer_info');const renderer=extension?gl!.getParameter(extension.UNMASKED_RENDERER_WEBGL):'unavailable';gl?.getExtension('WEBGL_lose_context')?.loseContext();return renderer})
  const result={...sample,renderExperiment,maxUnacknowledgedBytes,graphics,terminalRenderers:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalRenderer).filter(Boolean)),terminalCheckpoints:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalCheckpoint).filter(Boolean)),terminalBytesPerSecond,sampleSeconds,stateAfter,soak,windows:await page.locator('.app-window').count(),directoryEntries:10000,transferBytes:stat.size,transferSeconds:(Date.now()-transferStart)/1000,browser:browser.version(),cpu:cpus()[0]?.model,logicalCPUs:cpus().length,hostMemoryBytes:totalmem()}
  await info.attach('real-performance.json',{body:JSON.stringify(result,null,2),contentType:'application/json'})
  console.log('BLORA_REAL_PERF '+JSON.stringify(result))
  expect(errors).toEqual([])
  expect(stateAfter).toBe('RUNNING')
  expect(terminalBytesPerSecond).toHaveLength(2)
  expect(maxUnacknowledgedBytes).toBeLessThanOrEqual(256*1024+65536)
  expect(sample.samples).toBeGreaterThanOrEqual(100)
  expect(sample.p95).toBeLessThanOrEqual(50)
})
