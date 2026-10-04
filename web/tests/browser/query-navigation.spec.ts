import type {Page} from '@playwright/test'
import {test,expect} from '../helpers/management-fixture'

async function openDesktopWithSummary(page:Page,phase:(name:string)=>void){
  // Cold session/workspace initialization precedes the first task query. Wait
  // for that real HTTP response before asserting its rendering; navigation's
  // load event alone does not establish confirmed task data. This setup still
  // consumes the unchanged 45s case budget. The UI and recovery assertions
  // retain their original 5s limits, with no sleep, warmup or injected state.
  const [response]=await Promise.all([
    page.waitForResponse(response=>response.request().method()==='GET'&&new URL(response.url()).pathname==='/api/v1/tasks/summary'),
    page.goto('/'),
  ])
  expect(response.status()).toBe(200)
  expect(await response.finished()).toBeNull()
  phase('desktop bootstrap confirmed first task summary response')
}

async function openConfirmedTaskView(page:Page,phase:(name:string)=>void){
  const app=page.locator('[data-app="blora.tasks"]')
  await expect(app).toBeVisible()
  await expect(app).toBeEnabled()
  phase('activating task application through native Enter')
  // Opening an async application precedes its first list request. Confirm the
  // real response, not just a request counter, within the original case budget.
  // Native Enter uses the same visible application's normal activation path.
  const [response]=await Promise.all([
    page.waitForResponse(response=>response.request().method()==='GET'&&new URL(response.url()).pathname==='/api/v1/tasks'),
    app.press('Enter'),
  ])
  expect(response.status()).toBe(200)
  expect(await response.finished()).toBeNull()
  phase('task view confirmed first list response')
  const filter=page.getByRole('combobox',{name:'任务状态筛选'})
  await expect(filter).toBeVisible()
  await expect(filter).toBeEnabled()
  phase('task view filter visible and enabled')
}

async function recordSummaryAborts(page:Page){
  await page.addInitScript(()=>{
    const fetch=window.fetch.bind(window)
    let reads=0,leaving=false
    window.addEventListener('beforeunload',()=>{leaving=true})
    window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
      const url=typeof input==='string'?input:input instanceof URL?input.href:input.url
      if(new URL(url,location.href).pathname==='/api/v1/tasks/summary'){
        const read=++reads
        init?.signal?.addEventListener('abort',()=>{
          const events=JSON.parse(localStorage.getItem('blora:test:summary-aborts')||'[]') as {read:number;boundary:string}[]
          events.push({read,boundary:leaving?'beforeunload':'other'})
          localStorage.setItem('blora:test:summary-aborts',JSON.stringify(events))
        },{once:true})
      }
      return fetch(input,init)
    }
  })
}

async function expectPendingSummaryCancelled(page:Page){
  const events=await page.evaluate(()=>JSON.parse(localStorage.getItem('blora:test:summary-aborts')||'[]') as {read:number;boundary:string}[])
  // The second read is deliberately held by the route. Require its single
  // departure cancellation, not exactly one abort across every distinct read.
  // A slow real refresh can also cancel a read resumed by the bounded guard.
  expect(events.filter(event=>event.read===2)).toEqual([{read:2,boundary:'beforeunload'}])
  expect(events.every(event=>event.boundary==='beforeunload')).toBe(true)
}

test('pending task summary reads cancel before refresh while the latest task filter restores',async({page,validationPhase})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await recordSummaryAborts(page)
  let summaries=0,waiting=false,release:()=>void=()=>{}
  const writes:string[]=[]
  const held=new Promise<void>(resolve=>{release=resolve})
  await page.context().route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    if(route.request().method()!=='GET')writes.push(path)
    if(path.endsWith('/tasks/summary')){
      summaries++
      if(summaries===2){waiting=true;await held;return}
      return route.fulfill({json:{active:summaries>2?7:0,states:{}}})
    }
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'query-refresh',name:'查询恢复测试',admin:true},csrfToken:'test-only'}:{items:[],nextBefore:-1}})
  })
  try{
    validationPhase('opening desktop for pending task refresh')
    await openDesktopWithSummary(page,validationPhase)
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    await openConfirmedTaskView(page,validationPhase)
    await expect.poll(()=>waiting).toBe(true)
    validationPhase('pending summary established; editing latest task filter')
    // No settling delay after the last protected edit. The real refresh
    // must cancel the pending read without replaying a remote operation.
    await page.getByRole('combobox',{name:'任务状态筛选'}).selectOption('RUNNING')
    validationPhase('refreshing with pending task read')
    await page.reload()
    release()
    await expect(page.getByRole('combobox',{name:'任务状态筛选'})).toHaveValue('RUNNING')
    await expect(page.getByRole('button',{name:'7 项后台任务',exact:true})).toBeVisible()
    await expectPendingSummaryCancelled(page)
    expect(writes).toEqual([])
    expect(errors).toEqual([])
    validationPhase('task filter and confirmed summary restored; no writes or page errors')
  }finally{release()}
})

test('cancelled navigation keeps task summary polling and its last confirmed state',async({page,validationPhase})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await recordSummaryAborts(page)
  let summaries=0,waiting=false,release:()=>void=()=>{}
  const held=new Promise<void>(resolve=>{release=resolve})
  let resumeReply:()=>void=()=>{}
  const resumedReply=new Promise<void>(resolve=>{resumeReply=resolve})
  await page.context().route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    expect(route.request().method()).toBe('GET')
    if(path.endsWith('/tasks/summary')){
      summaries++
      if(summaries===2){waiting=true;await held;return}
      // Keep the resumed response pending until the retained-cache assertion
      // finishes. Immediate revalidation must not race that assertion.
      if(summaries===3)await resumedReply
      return route.fulfill({json:{active:summaries>2?7:0,states:{}}})
    }
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'query-stay',name:'查询继续测试',admin:true},csrfToken:'test-only'}:{items:[],nextBefore:-1}})
  })
  try{
    validationPhase('opening desktop for prevented navigation')
    await openDesktopWithSummary(page,validationPhase)
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    await expect.poll(()=>waiting).toBe(true)
    validationPhase('pending summary established; suppressing frames and preventing navigation')
    // Stop new animation-frame callbacks at the navigation boundary. Reads
    // must recover without depending on painting, within the original 5s.
    // This is not evidence of native confirmation-dialog or bfcache support.
    await page.evaluate(()=>{
      const frame=window.requestAnimationFrame
      Object.assign(window,{__bloraRestoreFrames:()=>{window.requestAnimationFrame=frame}})
      window.requestAnimationFrame=()=>0
      const event=new Event('beforeunload',{cancelable:true});event.preventDefault();window.dispatchEvent(event)
    })
    await expectPendingSummaryCancelled(page)
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    validationPhase('cancelled query retained its last confirmed count; releasing resumed response')
    resumeReply()
    await expect(page.getByRole('button',{name:'7 项后台任务',exact:true})).toBeVisible({timeout:5000})
    validationPhase('confirmed task count visible within original deadline without frames')
    expect(errors).toEqual([])
  }finally{resumeReply();release();await page.evaluate(()=>(window as unknown as {__bloraRestoreFrames?:()=>void}).__bloraRestoreFrames?.()).catch(()=>{})}
})

test('queued task polling cannot start new reads during beforeunload and resumes in the surviving document',async({page,validationPhase})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  let afterDeparture=false,lists=0
  // Change fixture data only once the browser actually dispatches departure,
  // not before a potentially delayed automation round trip to that boundary.
  await page.exposeFunction('__bloraQueuedPollDeparture',()=>{afterDeparture=true})
  await page.addInitScript(()=>{
    const callbacks=new Map<number,()=>void>()
    const interval=window.setInterval.bind(window),clear=window.clearInterval.bind(window),fetch=window.fetch.bind(window)
    let leaving=false
    const blocked:string[]=[]
    // Retain the real QueryObserver timer callbacks. Invoke already-queued
    // callbacks after beforeunload in the same event turn, including its
    // microtasks. The observation ends on the next task, independently of RAF.
    window.setInterval=((handler:TimerHandler,delay?:number,...args:unknown[])=>{
      const id=interval(handler,delay,...args)
      if(typeof handler==='function'&&(delay===2000||delay===3000))callbacks.set(id,()=>handler(...args))
      return id
    }) as typeof window.setInterval
    window.clearInterval=((id?:number)=>{if(id!==undefined)callbacks.delete(id);clear(id)}) as typeof window.clearInterval
    window.addEventListener('beforeunload',()=>{
      leaving=true;setTimeout(()=>{leaving=false},0)
      void (window as unknown as {__bloraQueuedPollDeparture:()=>Promise<void>}).__bloraQueuedPollDeparture().catch(()=>{})
    })
    window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
      const url=typeof input==='string'?input:input instanceof URL?input.href:input.url
      const path=new URL(url,location.href).pathname
      if(leaving&&/^\/api\/v1\/tasks(?:\/|$)/.test(path))blocked.push(path)
      return fetch(input,init)
    }
    Object.assign(window,{__bloraQueuedTaskPolls:{callbacks,blocked}})
  })
  const writes:string[]=[]
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(route.request().method()!=='GET')writes.push(path)
    // A slow cold application must not make the expected recovered count
    // visible before the departure being tested. Only this boundary changes it.
    if(path.endsWith('/tasks/summary'))return route.fulfill({json:{active:afterDeparture?7:0,states:{}}})
    if(path.endsWith('/tasks'))lists++
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'queued-poll',name:'离开轮询测试',admin:true},csrfToken:'test-only'}:{items:[],nextBefore:-1}})
  })
  validationPhase('opening desktop and task view for queued polling')
  await openDesktopWithSummary(page,validationPhase)
  await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
  await openConfirmedTaskView(page,validationPhase)
  await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
  const baselineLists=lists
  validationPhase('task observers active; invoking queued callbacks during departure')
  const observation=await page.evaluate(()=>{
    const state=(window as unknown as {__bloraQueuedTaskPolls:{callbacks:Map<number,()=>void>;blocked:string[]}}).__bloraQueuedTaskPolls
    const queued=[...state.callbacks.values()]
    const event=new Event('beforeunload',{cancelable:true});event.preventDefault();window.dispatchEvent(event)
    queued.forEach(callback=>callback())
    return{callbacks:queued.length,blocked:[...state.blocked]}
  })
  expect(observation.callbacks).toBeGreaterThanOrEqual(2)
  expect(observation.blocked).toEqual([])
  await expect(page.getByRole('button',{name:'7 项后台任务',exact:true})).toBeVisible()
  await expect.poll(()=>lists).toBeGreaterThan(baselineLists)
  expect(await page.evaluate(()=>(window as unknown as {__bloraQueuedTaskPolls:{blocked:string[]}}).__bloraQueuedTaskPolls.blocked)).toEqual([])
  expect(writes).toEqual([])
  expect(errors).toEqual([])
  validationPhase('queued task reads resumed; departure fetches, writes and page errors absent')
})

test('instance log polling cancels pending reads, gates queued polls and resumes without input replay',async({page,validationPhase})=>{
  const errors:string[]=[],writes:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.addInitScript(()=>{
    const callbacks=new Map<number,()=>void>(),blocked:string[]=[],aborts:string[]=[]
    const interval=window.setInterval.bind(window),clear=window.clearInterval.bind(window),fetch=window.fetch.bind(window)
    let leaving=false
    window.setInterval=((handler:TimerHandler,delay?:number,...args:unknown[])=>{
      const id=interval(handler,delay,...args)
      if(typeof handler==='function'&&delay===3000)callbacks.set(id,()=>handler(...args))
      return id
    }) as typeof window.setInterval
    window.clearInterval=((id?:number)=>{if(id!==undefined)callbacks.delete(id);clear(id)}) as typeof window.clearInterval
    window.addEventListener('beforeunload',()=>{leaving=true;setTimeout(()=>{leaving=false},0)})
    window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
      const url=typeof input==='string'?input:input instanceof URL?input.href:input.url,path=new URL(url,location.href).pathname
      if(path==='/api/v1/instances/poll-instance/logs'){
        if(leaving)blocked.push(path)
        init?.signal?.addEventListener('abort',()=>aborts.push(leaving?'beforeunload':'other'),{once:true})
      }
      return fetch(input,init)
    }
    Object.assign(window,{__bloraQueuedLogPolls:{callbacks,blocked,aborts}})
  })
  let reads=0,waiting=false,release:()=>void=()=>{}
  const held=new Promise<void>(resolve=>{release=resolve})
  await page.context().route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    if(route.request().method()!=='GET')writes.push(path)
    if(path.endsWith('/instances/poll-instance/logs')){
      if(++reads===2){waiting=true;await held;return}
      return route.fulfill({json:{items:[{runId:'poll-run',startedAt:'2026-10-03T00:00:00Z',backend:'test-boundary'},...(reads>2?[{runId:'later-run',startedAt:'2026-10-03T01:00:00Z',backend:'test-boundary'}]:[])]}})
    }
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'log-poll',name:'日志轮询测试',admin:true},csrfToken:'test-only'}:path.endsWith('/instances')?{items:[{instanceId:'poll-instance',nodeId:'poll-node',name:'轮询实例',state:'RUNNING',runId:'poll-run',config:{mode:'native'},revision:1}]}:{items:[]}})
  })
  // This case isolates HTTP polling; the separate log/terminal suites exercise
  // real binary stream processing, ACKs, checkpoints and command delivery.
  await page.context().routeWebSocket(/\/instances\/poll-instance\/logs\/poll-run\/stream/,()=>{})
  try{
    validationPhase('opening desktop for instance log polling')
    await page.goto('/')
    // Native button keyboard activation builds the same real UI state without
    // coupling this HTTP lifecycle case to pointer stability's RAF checks.
    // No force click, dispatched click or store-based navigation is used.
    const applications=page.locator('[data-app="blora.instances"]')
    await expect(applications).toBeVisible()
    validationPhase('activating instance application with native Enter key')
    await applications.press('Enter')
    const instance=page.getByRole('button',{name:'轮询实例',exact:true})
    await expect(instance).toBeVisible()
    validationPhase('activating instance resource with native Enter key')
    await instance.press('Enter')
    const consoleButton=page.getByRole('button',{name:'控制台',exact:true})
    await expect(consoleButton).toBeVisible()
    validationPhase('activating console with native Enter key')
    await consoleButton.press('Enter')
    const runs=page.getByRole('combobox',{name:'选择日志运行代次'})
    await expect(runs).toHaveValue('poll-run')
    await page.getByRole('textbox',{name:'尚未发送的实例命令'}).fill('保留草稿，不发送')
    validationPhase('console and run selected; command draft protected')
    // Establish the pending read by invoking the registered application timer
    // callbacks. A slow setup need not wait for the next wall-clock interval.
    await page.evaluate(()=>{
      const state=(window as unknown as {__bloraQueuedLogPolls:{callbacks:Map<number,()=>void>}}).__bloraQueuedLogPolls
      state.callbacks.forEach(callback=>callback())
    })
    await expect.poll(()=>waiting).toBe(true)
    validationPhase('pending log read established; invoking queued departure callbacks')
    const observation=await page.evaluate(()=>{
      const state=(window as unknown as {__bloraQueuedLogPolls:{callbacks:Map<number,()=>void>;blocked:string[];aborts:string[]}}).__bloraQueuedLogPolls
      const queued=[...state.callbacks.values()]
      const event=new Event('beforeunload',{cancelable:true});event.preventDefault();window.dispatchEvent(event)
      queued.forEach(callback=>callback())
      return{callbacks:queued.length,blocked:[...state.blocked],aborts:[...state.aborts]}
    })
    expect(observation.callbacks).toBeGreaterThanOrEqual(2)
    expect(observation.blocked).toEqual([])
    expect(observation.aborts).toEqual(['beforeunload'])
    await expect(runs).toHaveValue('poll-run')
    await expect(runs.locator('option[value="later-run"]')).toHaveCount(1)
    validationPhase('pending log read cancelled and run list resumed; checking second departure')
    // Repeat the queued-callback boundary after the previous pending read has
    // completed, so the component's overlap guard cannot mask a new fetch.
    const queuedAfterResume=await page.evaluate(()=>{
      const state=(window as unknown as {__bloraQueuedLogPolls:{callbacks:Map<number,()=>void>;blocked:string[]}}).__bloraQueuedLogPolls
      const event=new Event('beforeunload',{cancelable:true});event.preventDefault();window.dispatchEvent(event)
      state.callbacks.forEach(callback=>callback())
      return[...state.blocked]
    })
    expect(queuedAfterResume).toEqual([])
    await expect.poll(()=>reads).toBeGreaterThanOrEqual(4)
    expect(await page.evaluate(()=>(window as unknown as {__bloraQueuedLogPolls:{blocked:string[]}}).__bloraQueuedLogPolls.blocked)).toEqual([])
    await expect(page.locator('.instance-console [role="alert"]')).toHaveCount(0)
    release()
    validationPhase('refreshing console; draft and run must restore without replay')
    await page.reload()
    await expect(page.getByRole('textbox',{name:'尚未发送的实例命令'})).toHaveValue('保留草稿，不发送')
    await expect(runs).toHaveValue('poll-run')
    expect(writes).toEqual([])
    expect(errors).toEqual([])
    validationPhase('log polling, draft recovery and no input replay verified')
  }finally{release()}
})
