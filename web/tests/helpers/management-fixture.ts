import {test as base,expect} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,MessageType} from '../../src/services/protocol'

// These suites deliberately replace management transports. Include the
// desktop's independent notification stream so it cannot contact a real or
// absent daemon while the tested terminal/HTTP routes use their own doubles.
export const test=base.extend<{managementDiagnostics:void}>({
  managementDiagnostics:[async({page},use,testInfo)=>{
    const signals:unknown[]=[]
    const record=(value:Record<string,unknown>)=>{if(signals.length<120)signals.push({at:Date.now(),...value})}
    const requests=new WeakMap<object,number>()
    let requestId=0
    page.on('request',request=>{
      const path=new URL(request.url()).pathname
      const task=/^\/api\/v1\/tasks(?:\/|$)/.test(path),log=/^\/api\/v1\/instances\/[^/]+\/logs(?:\/|$)/.test(path)
      if(!task&&!log)return
      requests.set(request,++requestId)
      record({kind:task?'task-request':'log-request',id:requestId,method:request.method(),path})
    })
    page.on('requestfinished',request=>{const id=requests.get(request);if(id!==undefined)record({kind:new URL(request.url()).pathname.startsWith('/api/v1/tasks')?'task-request-finished':'log-request-finished',id})})
    page.on('framenavigated',frame=>{if(frame===page.mainFrame())record({kind:'navigation',path:new URL(frame.url()).pathname})})
    page.on('pageerror',error=>record({kind:'playwright-pageerror',name:error.name,message:error.message,stack:error.stack?.slice(0,2000)}))
    page.on('requestfailed',request=>record({kind:'requestfailed',id:requests.get(request),method:request.method(),path:new URL(request.url()).pathname,error:request.failure()?.errorText}))
    await page.exposeFunction('__bloraValidationError',(kind:string,message:string)=>record({kind,message:message.slice(0,2000)}))
    await page.addInitScript(()=>{
      const report=(kind:string,message:string)=>{void (window as unknown as {__bloraValidationError:(kind:string,message:string)=>Promise<void>}).__bloraValidationError(kind,message).catch(()=>{})}
      window.addEventListener('error',event=>{if(event instanceof ErrorEvent)report('window-error',event.message)})
      window.addEventListener('unhandledrejection',event=>report('unhandled-rejection',String(event.reason)))
      for(const event of ['beforeunload','pagehide','pageshow'])window.addEventListener(event,()=>report('lifecycle',event))
    })
    await page.context().routeWebSocket(/\/api\/v1\/events\?/,ws=>{
      const send=(type:MessageType)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:1,type,sequence:0})))
      ws.onMessage(message=>{if(decodeEnvelope(new Uint8Array(message as Buffer)).type===MessageType.Ping)send(MessageType.Pong)})
      send(MessageType.Snapshot)
    })
    await use()
    if(testInfo.status!==testInfo.expectedStatus){
      let surface:unknown
      try{
        surface=await Promise.race([
          page.evaluate(async()=>{
            const paint=await Promise.race([new Promise<boolean>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve(true)))),new Promise<boolean>(resolve=>setTimeout(()=>resolve(false),250))])
            const bounds=(element:Element)=>{const r=element.getBoundingClientRect();return{x:r.x,y:r.y,width:r.width,height:r.height}}
            return{visibility:document.visibilityState,focused:document.hasFocus(),paint,viewport:{width:innerWidth,height:innerHeight},dialogs:Array.from(document.querySelectorAll('[role="dialog"]')).map(element=>({bounds:bounds(element),scrollTop:element.scrollTop,clientHeight:element.clientHeight,scrollHeight:element.scrollHeight,buttons:Array.from(element.querySelectorAll('button')).map(button=>({bounds:bounds(button),disabled:button.disabled})),inputs:Array.from(element.querySelectorAll('input')).map(input=>({type:input.type,valid:input.validity.valid,focused:document.activeElement===input}))}))}
          }),
          new Promise(resolve=>setTimeout(()=>resolve({unavailable:'surface probe exceeded 3 seconds'}),3000)),
        ])
      }catch{surface={unavailable:'page closed or execution context unavailable'}}
      // No request bodies, field contents, storage dumps or traces. This JSON
      // distinguishes native JS errors from browser network diagnostics and
      // captures stalled paint/input geometry for the next Windows report.
      await testInfo.attach('blora-runtime-diagnostics',{body:Buffer.from(JSON.stringify({test:testInfo.title,signals,surface},null,2)),contentType:'application/json'})
    }
  },{auto:true}],
})
export {expect}
