import type {Page} from '@playwright/test'
import {test,expect} from '../helpers/management-fixture'

async function recordSummaryAborts(page:Page){
  await page.addInitScript(()=>{
    const fetch=window.fetch.bind(window)
    window.addEventListener('beforeunload',()=>localStorage.setItem('blora:test:leaving','true'))
    window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
      const url=typeof input==='string'?input:input instanceof URL?input.href:input.url
      if(new URL(url,location.href).pathname==='/api/v1/tasks/summary'){
        init?.signal?.addEventListener('abort',()=>{
          const events=JSON.parse(localStorage.getItem('blora:test:summary-aborts')||'[]') as string[]
          events.push(localStorage.getItem('blora:test:leaving')==='true'?'beforeunload':'other')
          localStorage.setItem('blora:test:summary-aborts',JSON.stringify(events))
        },{once:true})
      }
      return fetch(input,init)
    }
  })
}

test('pending task summary reads cancel before refresh while the latest task filter restores',async({page})=>{
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
    await page.goto('/')
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    await page.locator('[data-app="blora.tasks"]').click()
    await expect.poll(()=>waiting).toBe(true)
    // No settling delay after the last protected edit. The real refresh
    // must cancel the pending read without replaying a remote operation.
    await page.getByRole('combobox',{name:'任务状态筛选'}).selectOption('RUNNING')
    await page.reload()
    release()
    await expect(page.getByRole('combobox',{name:'任务状态筛选'})).toHaveValue('RUNNING')
    await expect(page.getByRole('button',{name:'7 项后台任务',exact:true})).toBeVisible()
    expect(await page.evaluate(()=>JSON.parse(localStorage.getItem('blora:test:summary-aborts')||'[]'))).toEqual(['beforeunload'])
    expect(writes).toEqual([])
    expect(errors).toEqual([])
  }finally{release()}
})

test('cancelled navigation keeps task summary polling and its last confirmed state',async({page})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await recordSummaryAborts(page)
  let summaries=0,waiting=false,release:()=>void=()=>{}
  const held=new Promise<void>(resolve=>{release=resolve})
  await page.context().route('**/api/v1/**',async route=>{
    const path=new URL(route.request().url()).pathname
    expect(route.request().method()).toBe('GET')
    if(path.endsWith('/tasks/summary')){
      summaries++
      if(summaries===2){waiting=true;await held;return}
      return route.fulfill({json:{active:summaries>2?7:0,states:{}}})
    }
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'query-stay',name:'查询继续测试',admin:true},csrfToken:'test-only'}:{items:[],nextBefore:-1}})
  })
  try{
    await page.goto('/')
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    await expect.poll(()=>waiting).toBe(true)
    // Stop new animation-frame callbacks at the navigation boundary. Reads
    // must recover without depending on painting, within the original 5s.
    // This is not evidence of native confirmation-dialog or bfcache support.
    await page.evaluate(()=>{
      const frame=window.requestAnimationFrame
      Object.assign(window,{__bloraRestoreFrames:()=>{window.requestAnimationFrame=frame}})
      window.requestAnimationFrame=()=>0
      const event=new Event('beforeunload',{cancelable:true});event.preventDefault();window.dispatchEvent(event)
    })
    expect(await page.evaluate(()=>JSON.parse(localStorage.getItem('blora:test:summary-aborts')||'[]'))).toEqual(['beforeunload'])
    await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
    await expect(page.getByRole('button',{name:'7 项后台任务',exact:true})).toBeVisible({timeout:5000})
    expect(errors).toEqual([])
  }finally{release();await page.evaluate(()=>(window as unknown as {__bloraRestoreFrames?:()=>void}).__bloraRestoreFrames?.()).catch(()=>{})}
})

test('queued task polling cannot start new reads during beforeunload and resumes in the surviving document',async({page})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
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
    window.addEventListener('beforeunload',()=>{leaving=true;setTimeout(()=>{leaving=false},0)})
    window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
      const url=typeof input==='string'?input:input instanceof URL?input.href:input.url
      const path=new URL(url,location.href).pathname
      if(leaving&&/^\/api\/v1\/tasks(?:\/|$)/.test(path))blocked.push(path)
      return fetch(input,init)
    }
    Object.assign(window,{__bloraQueuedTaskPolls:{callbacks,blocked}})
  })
  let summaries=0,lists=0
  const writes:string[]=[]
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(route.request().method()!=='GET')writes.push(path)
    if(path.endsWith('/tasks/summary'))return route.fulfill({json:{active:++summaries>1?7:0,states:{}}})
    if(path.endsWith('/tasks'))lists++
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'queued-poll',name:'离开轮询测试',admin:true},csrfToken:'test-only'}:{items:[],nextBefore:-1}})
  })
  await page.goto('/')
  await expect(page.getByRole('button',{name:'0 项后台任务',exact:true})).toBeVisible()
  await page.locator('[data-app="blora.tasks"]').click()
  await expect.poll(()=>lists).toBe(1)
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
  await expect.poll(()=>lists).toBeGreaterThanOrEqual(2)
  expect(await page.evaluate(()=>(window as unknown as {__bloraQueuedTaskPolls:{blocked:string[]}}).__bloraQueuedTaskPolls.blocked)).toEqual([])
  expect(writes).toEqual([])
  expect(errors).toEqual([])
})

test('instance log polling cancels pending reads, gates queued polls and resumes without input replay',async({page})=>{
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
    await page.goto('/')
    await page.locator('[data-app="blora.instances"]').click()
    await page.getByRole('button',{name:'轮询实例',exact:true}).click()
    await page.getByRole('button',{name:'控制台',exact:true}).click()
    const runs=page.getByRole('combobox',{name:'选择日志运行代次'})
    await expect(runs).toHaveValue('poll-run')
    await page.getByRole('textbox',{name:'尚未发送的实例命令'}).fill('保留草稿，不发送')
    // Establish the pending read by invoking the registered application timer
    // callbacks. A slow setup need not wait for the next wall-clock interval.
    await page.evaluate(()=>{
      const state=(window as unknown as {__bloraQueuedLogPolls:{callbacks:Map<number,()=>void>}}).__bloraQueuedLogPolls
      state.callbacks.forEach(callback=>callback())
    })
    await expect.poll(()=>waiting).toBe(true)
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
    await page.reload()
    await expect(page.getByRole('textbox',{name:'尚未发送的实例命令'})).toHaveValue('保留草稿，不发送')
    await expect(runs).toHaveValue('poll-run')
    expect(writes).toEqual([])
    expect(errors).toEqual([])
  }finally{release()}
})
