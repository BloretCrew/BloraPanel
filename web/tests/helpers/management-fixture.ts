import {test as base,expect} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,MessageType} from '../../src/services/protocol'

// These suites deliberately replace management transports. Include the
// desktop's independent notification stream so it cannot contact a real or
// absent daemon while the tested terminal/HTTP routes use their own doubles.
export const test=base.extend<{validationPhase:(name:string)=>void}>({
  validationPhase:[async({page},use,testInfo)=>{
    const started=Date.now(),phases:{name:string;elapsedMs:number}[]=[]
    const signals:unknown[]=[]
    const record=(value:Record<string,unknown>)=>{if(signals.length<120)signals.push({at:Date.now(),...value})}
    const assetRequests=new WeakMap<object,{path:string;started:number}>()
    const slowestAssets:{path:string;durationMs:number}[]=[]
    let assetCount=0
    const requests=new WeakMap<object,number>()
    let requestId=0
    page.on('request',request=>{
      const path=new URL(request.url()).pathname
      if(/^\/(?:src|assets|node_modules\/\.vite|@vite)\//.test(path)&&/\.(?:js|ts|vue|css)$/.test(path))assetRequests.set(request,{path,started:Date.now()})
      const task=/^\/api\/v1\/tasks(?:\/|$)/.test(path),log=/^\/api\/v1\/instances\/[^/]+\/logs(?:\/|$)/.test(path)
      if(!task&&!log)return
      requests.set(request,++requestId)
      record({kind:task?'task-request':'log-request',id:requestId,method:request.method(),path})
    })
    page.on('requestfinished',request=>{
      const asset=assetRequests.get(request)
      if(asset){assetCount++;slowestAssets.push({path:asset.path,durationMs:Date.now()-asset.started});slowestAssets.sort((a,b)=>b.durationMs-a.durationMs);slowestAssets.length=Math.min(slowestAssets.length,20)}
      const id=requests.get(request);if(id!==undefined)record({kind:new URL(request.url()).pathname.startsWith('/api/v1/tasks')?'task-request-finished':'log-request-finished',id})
    })
    page.on('framenavigated',frame=>{if(frame===page.mainFrame())record({kind:'navigation',path:new URL(frame.url()).pathname})})
    page.on('pageerror',error=>record({kind:'playwright-pageerror',name:error.name,message:error.message,stack:error.stack?.slice(0,2000)}))
    page.on('requestfailed',request=>record({kind:'requestfailed',id:requests.get(request),method:request.method(),path:new URL(request.url()).pathname,error:request.failure()?.errorText}))
    await page.exposeFunction('__bloraValidationError',(kind:string,message:string)=>record({kind,message:message.slice(0,2000)}))
    await page.exposeFunction('__bloraValidationTaskSignal',(value:Record<string,unknown>)=>{
      // Whitelist numeric milestones only. No JSON bodies, labels, input text
      // or storage contents cross this diagnostic binding.
      if(!['summary-json-consumed','task-count-rendered','document-lifecycle'].includes(String(value.kind)))return
      if(typeof value.browserAt!=='number'||!Number.isFinite(value.browserAt))return
      const signal:Record<string,unknown>={kind:value.kind,browserAt:value.browserAt}
      if(typeof value.read==='number'&&Number.isSafeInteger(value.read))signal.read=value.read
      if(typeof value.activeCount==='number'&&Number.isSafeInteger(value.activeCount)&&value.activeCount>=0)signal.activeCount=value.activeCount
      if(value.unconfirmed===true)signal.unconfirmed=true
      if(['beforeunload','pagehide','pageshow'].includes(String(value.event)))signal.event=value.event
      record(signal)
    })
    await page.addInitScript(()=>{
      const report=(kind:string,message:string)=>{void (window as unknown as {__bloraValidationError:(kind:string,message:string)=>Promise<void>}).__bloraValidationError(kind,message).catch(()=>{})}
      const milestone=(value:Record<string,unknown>)=>{void (window as unknown as {__bloraValidationTaskSignal:(value:Record<string,unknown>)=>Promise<void>}).__bloraValidationTaskSignal({...value,browserAt:Date.now()}).catch(()=>{})}
      window.addEventListener('error',event=>{if(event instanceof ErrorEvent)report('window-error',event.message)})
      window.addEventListener('unhandledrejection',event=>report('unhandled-rejection',String(event.reason)))
      for(const event of ['beforeunload','pagehide','pageshow'])window.addEventListener(event,()=>milestone({kind:'document-lifecycle',event}))
      const fetch=window.fetch.bind(window)
      let reads=0
      window.fetch=(input:RequestInfo|URL,init?:RequestInit)=>{
        const url=typeof input==='string'?input:input instanceof URL?input.href:input.url
        const summary=new URL(url,location.href).pathname==='/api/v1/tasks/summary'&&(init?.method|| (input instanceof Request?input.method:'GET')).toUpperCase()==='GET'
        const read=summary?++reads:0
        const request=fetch(input,init)
        if(!summary)return request
        return request.then(response=>{
          // Observe the application's own body consumption. Do not read a
          // clone, change responses, await the binding, or publish query state.
          const json=response.json.bind(response)
          response.json=async()=>{const value=await json();milestone({kind:'summary-json-consumed',read,activeCount:value?.active});return value}
          return response
        })
      }
      const observe=()=>{
        let previous:string|undefined
        const check=()=>{
          const label=Array.from(document.querySelectorAll('.taskbar button')).map(button=>button.getAttribute('aria-label')||'').find(label=>/^\d+ 项后台任务$/.test(label)||label==='后台任务待确认')
          if(!label||label===previous)return
          previous=label
          milestone(label==='后台任务待确认'?{kind:'task-count-rendered',unconfirmed:true}:{kind:'task-count-rendered',activeCount:Number(label.split(' ')[0])})
        }
        new MutationObserver(check).observe(document.body,{childList:true,subtree:true,attributes:true,attributeFilter:['aria-label']})
        check()
      }
      if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',observe,{once:true});else observe()
    })
    await page.context().routeWebSocket(/\/api\/v1\/events\?/,ws=>{
      const send=(type:MessageType)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:1,type,sequence:0})))
      ws.onMessage(message=>{if(decodeEnvelope(new Uint8Array(message as Buffer)).type===MessageType.Ping)send(MessageType.Pong)})
      send(MessageType.Snapshot)
    })
    await use(name=>{
      // Callers supply static operation labels, never user data or input text.
      if(phases.length>=30)return
      const phase={name:name.slice(0,120),elapsedMs:Date.now()-started};phases.push(phase)
      console.log(`[browser phase +${phase.elapsedMs}ms] ${phase.name}`)
    })
    if(testInfo.status!==testInfo.expectedStatus){
      let surface:unknown
      try{
        surface=await Promise.race([
          page.evaluate(async()=>{
            const paint=await Promise.race([new Promise<boolean>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve(true)))),new Promise<boolean>(resolve=>setTimeout(()=>resolve(false),250))])
            const bounds=(element:Element)=>{const r=element.getBoundingClientRect();return{x:r.x,y:r.y,width:r.width,height:r.height}}
            const taskLabel=Array.from(document.querySelectorAll('.taskbar button')).map(button=>button.getAttribute('aria-label')||'').find(label=>/^\d+ 项后台任务$/.test(label)||label==='后台任务待确认')
            const consoles=Array.from(document.querySelectorAll<HTMLButtonElement>('.resource-nav button')).filter(button=>button.textContent==='控制台').map(button=>({bounds:bounds(button),disabled:button.disabled,focused:document.activeElement===button,selected:button.classList.contains('selected')}))
            return{visibility:document.visibilityState,focused:document.hasFocus(),paint,taskSummary:taskLabel==='后台任务待确认'?{unconfirmed:true}:taskLabel?{activeCount:Number(taskLabel.split(' ')[0])}:undefined,consoles,viewport:{width:innerWidth,height:innerHeight},dialogs:Array.from(document.querySelectorAll('[role="dialog"]')).map(element=>({bounds:bounds(element),scrollTop:element.scrollTop,clientHeight:element.clientHeight,scrollHeight:element.scrollHeight,buttons:Array.from(element.querySelectorAll('button')).map(button=>({bounds:bounds(button),disabled:button.disabled})),inputs:Array.from(element.querySelectorAll('input')).map(input=>({type:input.type,valid:input.validity.valid,focused:document.activeElement===input}))}))}
          }),
          new Promise(resolve=>setTimeout(()=>resolve({unavailable:'surface probe exceeded 3 seconds'}),3000)),
        ])
      }catch{surface={unavailable:'page closed or execution context unavailable'}}
      // No request bodies, field contents, storage dumps or traces. This JSON
      // distinguishes native JS errors from browser network diagnostics and
      // captures stalled paint/input geometry for the next Windows report.
      await testInfo.attach('blora-runtime-diagnostics',{body:Buffer.from(JSON.stringify({test:testInfo.title,browserServerMode:process.env.BLORA_E2E_PREVIEW==='1'?'production-preview':'development',phases,assets:{completed:assetCount,slowest:slowestAssets},signals,surface},null,2)),contentType:'application/json'})
    }
  },{auto:true}],
})
export {expect}
