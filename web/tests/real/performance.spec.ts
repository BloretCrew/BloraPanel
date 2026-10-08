import {readFileSync} from 'node:fs'
import {cpus,totalmem} from 'node:os'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {realLogin} from './login'
import {decodeEnvelope,MessageType} from '../../src/services/protocol'
import {startRenderTrace} from './render-trace'
import {installShadowBandDiagnostic} from './shadow-band-diagnostic'
import {installShadowResolutionDiagnostic} from './shadow-resolution-diagnostic'
import {installBodyViewportDiagnostic} from './body-viewport-diagnostic'
import {installShadowCropDiagnostic} from '../helpers/shadow-crop'
import {installWindowMaterialDiagnostic} from './window-material-diagnostic'
import {installBarShadowDiagnostic} from './bar-shadow-diagnostic'
import {installShadowCompositeDiagnostic} from './shadow-composite-diagnostic'
import {installShadowAlphaDiagnostic} from './shadow-alpha-diagnostic'
import {installNativeTerminalPaintDiagnostic} from './terminal-paint-diagnostic'
import {installSmallMaterialAtlas} from './material-atlas-diagnostic'
import {installShadowSiblings} from '../helpers/shadow-sibling'
import {installShadowMasks} from '../helpers/shadow-mask'
import {nativeCssPlaneShadow} from '../helpers/native-css-shadow'
import {installOutsideShadows} from '../helpers/shadow-outside'
import {installShadowTail} from '../helpers/shadow-tail'
import {installStaticShadowUnderlay} from '../helpers/static-shadow-underlay'
import {installStaticShadowScene} from '../helpers/static-shadow-scene'
import {installNativeBodyClip} from '../helpers/native-body-clip'
import {installNativeMotion} from '../helpers/native-motion'
import {installNativeRetention} from '../helpers/native-retention'
import {installNativeRowRelevance} from '../helpers/native-row-relevance'
import {installNativeFrameMaterial} from '../helpers/native-frame-material'
import {installNativeWindowPosition} from '../helpers/native-window-position'
import {installNativePartialBody} from '../helpers/native-partial-body'

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
  const terminalRendererExperiment=process.env.BLORA_PERF_TERMINAL_RENDERER||'baseline'
  expect(['baseline','default','native-webgl-software','native-dom-equality']).toContain(terminalRendererExperiment)
  info.annotations.push({type:'terminal-renderer-experiment',description:terminalRendererExperiment})
  if(terminalRendererExperiment==='native-dom-equality')await page.addInitScript(()=>{
    // Retain existing native nodes only when the official factory produced an
    // exactly equal row. No custom row renderer, snapshot or output buffering.
    const original=Element.prototype.replaceChildren
    const counts={calls:0,reused:0,replaced:0}
    ;(window as any).__nativeDomEquality=counts
    Element.prototype.replaceChildren=function(this:Element,...nodes:(Node|string)[]){
      if(this.parentElement?.classList.contains('xterm-rows')){
        counts.calls++
        if(nodes.length===this.childNodes.length&&nodes.every((node,index)=>node instanceof Node&&node.isEqualNode(this.childNodes[index]!))&&!this.querySelector('.xterm-cursor-blink')){counts.reused++;return}
        counts.replaced++
      }
      original.apply(this,nodes)
    }
  })
  if(terminalRendererExperiment==='native-webgl-software')await page.addInitScript(()=>{
    // Diagnostic only: allow the pinned official WebGL addon to load on a
    // software device. Do not modify the context used by the actual addon,
    // its renderer, terminal parser, output, protection or native glyphs.
    const original=HTMLCanvasElement.prototype.getContext
    HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,type:string,...args:any[]){
      if(type==='webgl2'&&args[0]?.failIfMajorPerformanceCaveat){
        const context=(original as any).call(this,type,{...args[0],failIfMajorPerformanceCaveat:false}) as WebGL2RenderingContext|null
        if(context){
          const extension=context.getExtension.bind(context)
          context.getExtension=((name:string)=>name==='WEBGL_debug_renderer_info'?null:extension(name)) as typeof context.getExtension
        }
        return context
      }
      return (original as any).call(this,type,...args)
    } as typeof original
  })
  if(terminalRendererExperiment==='default')await page.addInitScript(()=>{
    const original=HTMLCanvasElement.prototype.getContext
    HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,type:string,...args:any[]){
      if(type==='webgl'||type==='webgl2'||type==='experimental-webgl')return null
      return (original as any).call(this,type,...args)
    } as typeof original
  })
  expect(fixture.performance,'Run devfixture with --performance; missing real data is not a passing probe').toBe(true)
  const configuredSoak=Number(process.env.BLORA_PERF_SOAK_SECONDS||0)
  const soakSeconds=Number.isFinite(configuredSoak)&&configuredSoak>0?Math.min(3600,Math.floor(configuredSoak)):0
  const terminalPaintDiagnostic=process.env.BLORA_PERF_TERMINAL_PAINT_DIAGNOSTIC==='1'
  if(terminalPaintDiagnostic){expect(soakSeconds).toBe(0);await page.addInitScript(installNativeTerminalPaintDiagnostic)}
  // Additional long stability runs must keep the gesture visible throughout;
  // the original 60-second acceptance trajectory remains exactly unchanged.
  const longMotion=process.env.BLORA_PERF_LONG_MOTION||'original'
  expect(['original','bounded']).toContain(longMotion)
  if(longMotion==='bounded')expect(soakSeconds).toBeGreaterThan(60)
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
  let terminalDataMessages=0
  page.on('pageerror',error=>errors.push(error.message))
  page.on('websocket',socket=>{if(!socket.url().includes('/terminals/'))return;let received=0,acknowledged=0;socket.on('framesent',event=>{
    if(typeof event.payload==='string')return
    const frame=decodeEnvelope(new Uint8Array(event.payload));if(frame.type===MessageType.Ack)acknowledged=frame.sequence||0
  });socket.on('framereceived',event=>{
    if(typeof event.payload==='string')return
    const frame=decodeEnvelope(new Uint8Array(event.payload))
    if(frame.type===MessageType.Data){terminalDataMessages++;bytes.set(socket.url(),(bytes.get(socket.url())||0)+(frame.payload?.length||0));received=frame.sequence||0;maxUnacknowledgedBytes=Math.max(maxUnacknowledgedBytes,received-acknowledged)}
  })})
  await realLogin(page,fixture.admin)
  const renderExperiment=process.env.BLORA_PERF_RENDER_EXPERIMENT||'baseline'
  const nativeRetentionGlassOnly=process.env.BLORA_PERF_APP_VIEW_GLASS_ONLY==='1'
  if(nativeRetentionGlassOnly)expect(renderExperiment).toBe('native-appview-retention')
  const nativeFlatBody=process.env.BLORA_PERF_NATIVE_FLAT_BODY==='1'
  if(nativeFlatBody)expect(renderExperiment).toBe('native-body-clip')
  const nativePartialBody=process.env.BLORA_PERF_NATIVE_PARTIAL_BODY==='1'
  if(nativePartialBody){expect(renderExperiment).toBe('native-body-clip');expect(nativeFlatBody).toBe(false)}
  const nativeTransformControl=process.env.BLORA_PERF_NATIVE_TRANSFORM_CONTROL==='1'
  if(nativeTransformControl){
    expect(renderExperiment).toBe('baseline')
    await page.addStyleTag({content:'.app-window[data-window-position="offset"]{left:0!important;top:0!important;transform:translate(var(--window-x),var(--window-y))!important}'})
  }
  const windowMaterialExperiment=['native-scale','native-rgb24'].includes(process.env.BLORA_PERF_WINDOW_MATERIAL||'')
  const smallMaterialExperiment=process.env.BLORA_PERF_SMALL_MATERIAL_ATLAS==='1'
  expect(['baseline','native-window-position','native-frame-material','native-row-relevance','native-software-paint','native-appview-retention','native-animation-motion','native-body-clip','static-shadow-underlay','static-shadow-scene','static-shadow-hull','native-terminal-row-layout','native-icon-layers','native-shadow-tail','native-shadow-outside','native-shadow-mask','native-css-plane','native-shadow-sibling','contain','layers','window-layers','active-layer','active-app-layer','chrome-layers','shadow-plane-layers','body-viewport','shadow-overflow','bar-shadow-cache','bar-shadow-native','native-shadow-composite','native-shadow-partition','native-shadow-alpha','material-nearest','shadow-nearest','shadow-nearest-all','native-window-surface','shadow-bands','shadow-native-resolution','dock-flat-shadow','native-dock-filter','full-native-list-paint','full-native-body-paint','direct-body-material','body-layers','shadow-layers','active-shadow-plane','focused-body-layer','flat-active-material','terminal-paint-containment','opaque-frame','glass-layer','no-glass','no-shadow','no-window-shadow','no-chrome-shadow','no-content','no-body-layer','border-shadow','full-native-shadow-patches']).toContain(renderExperiment)
  // Diagnostic comparisons only: each run keeps the same real resources,
  // terminal output, interaction sampling and acceptance threshold.
  if(!['baseline','native-window-position','native-frame-material','native-row-relevance','native-software-paint','native-appview-retention','native-animation-motion','native-body-clip','static-shadow-underlay','static-shadow-scene','static-shadow-hull','native-shadow-tail','native-shadow-outside','native-shadow-mask','native-css-plane','native-shadow-sibling','no-content','border-shadow','shadow-bands','shadow-native-resolution','body-viewport','shadow-overflow','bar-shadow-cache','bar-shadow-native','native-shadow-composite','native-shadow-partition','native-shadow-alpha'].includes(renderExperiment))await page.addStyleTag({content:renderExperiment==='contain'
    ?'.terminal-container .xterm-rows{contain:layout style}'
    :renderExperiment==='native-terminal-row-layout'?'.xterm-screen:not(:has(.xterm-selection>div)) .xterm-rows>div{contain:layout style}'
    :renderExperiment==='native-icon-layers'?'.application-art{will-change:transform}'
    :renderExperiment==='layers'?'.app-window{will-change:transform}.terminal-container{contain:layout paint style}'
    :renderExperiment==='window-layers'?'.app-window{will-change:transform}'
    :renderExperiment==='active-layer'?'.app-window.focused{will-change:transform}'
    :renderExperiment==='active-app-layer'?'.app-window.focused>.window-body{will-change:transform}'
    :renderExperiment==='chrome-layers'?'.topbar,.taskbar{will-change:transform}'
    :renderExperiment==='shadow-plane-layers'?'.window-shadow-plane[data-shadow-cache]{will-change:transform!important}'
    :renderExperiment==='shadow-nearest'?'@media (resolution:1dppx){.window-shadow-plane[data-shadow-cache]>.shadow-piece:is(.n,.s,.w,.e){image-rendering:pixelated}}'
    :renderExperiment==='shadow-nearest-all'?'@media (resolution:1dppx){.window-shadow-plane[data-shadow-cache]>.shadow-piece{image-rendering:pixelated}}'
    :renderExperiment==='no-window-shadow'?'.window-shadow-plane{display:none!important}'
    :renderExperiment==='no-chrome-shadow'?'.topbar{box-shadow:none!important}.dock-surface>svg{filter:none!important}'
    :renderExperiment==='material-nearest'?':root[data-translucency="on"][data-material-cache="ready"] .app-window::before,:root[data-translucency="on"][data-material-cache="ready"] .app-window>.window-body::before{image-rendering:pixelated}'
    :renderExperiment==='native-window-surface'?'.app-window.focused{translate:0px 0px 0px;backface-visibility:hidden}'
    :renderExperiment==='dock-flat-shadow'?'.dock-surface>svg{filter:none!important}'
    :renderExperiment==='native-dock-filter'?'.dock-surface>svg{display:block!important;visibility:visible!important}.dock-optical-cache{display:none!important}'
    :renderExperiment==='full-native-shadow-patches'?'.shadow-piece:not([style*="inset(100%)"]){clip-path:none!important;clip:auto!important}'
    :renderExperiment==='full-native-body-paint'?'.window-body:not([style*="inset(100%)"]){clip-path:none!important}'
    :renderExperiment==='full-native-list-paint'?'[data-app-paint-occluded="true"]{clip-path:none!important}'
    :renderExperiment==='direct-body-material'?':root[data-translucency="on"][data-material-cache="ready"] .app-window>.window-body::before{display:none!important}:root[data-translucency="on"][data-material-cache="ready"] .app-window>.window-body{background-image:var(--material-window)!important;background-size:var(--material-size);background-position:var(--material-x,0px) var(--material-y,0px);background-repeat:no-repeat;background-origin:border-box;background-color:var(--surface-container)!important}'
    :renderExperiment==='shadow-layers'?'.app-window.focused>.window-shadow-plane[data-shadow-cache]>.shadow-piece{will-change:transform}'
    :renderExperiment==='active-shadow-plane'?'.app-window.focused>.window-shadow-plane{will-change:transform!important;transform:translateZ(0)}'
    :renderExperiment==='focused-body-layer'?':root[data-translucency="on"][data-material-cache="ready"] .app-window:not(.focused)>.window-body::before{will-change:auto!important}'
    :renderExperiment==='flat-active-material'?':root[data-translucency="on"][data-material-cache="ready"] .app-window.focused>.window-body::before{will-change:auto!important}'
    :renderExperiment==='terminal-paint-containment'?'.terminal-container{contain:layout paint style}'
    :renderExperiment==='opaque-frame'?':root[data-translucency="on"][data-material-cache="ready"][data-material-cache-opaque="true"] .app-window{background-color:var(--surface-container)!important}'
    :renderExperiment==='no-body-layer'?'.app-window>.window-body::before{will-change:auto!important}'
    :renderExperiment==='body-layers'?':root[data-material-cache="ready"][data-material-cache-opaque="true"] .app-window>.window-body{background-color:var(--surface-container);will-change:transform}'
    :renderExperiment==='glass-layer'?':root[data-translucency="on"] .app-window{backdrop-filter:none!important;background:transparent!important}:root[data-translucency="on"] .app-window::before{content:"";position:absolute;inset:0;border-radius:inherit;z-index:-1;pointer-events:none;background:var(--glass-window);backdrop-filter:blur(24px) saturate(1.05)}:root[data-theme="dark"][data-translucency="on"] .app-window::before{backdrop-filter:blur(20px) saturate(1.03)}'
    :renderExperiment==='no-shadow'?'*,*::before,*::after{box-shadow:none!important;filter:none!important}.window-shadow-plane{display:none!important}'
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
    await expect(terminal.locator('.xterm-helper-textarea')).toBeFocused()
    const sessionId=await terminal.getAttribute('data-session-id')
    const produced=()=>[...bytes].filter(([url])=>new URL(url).pathname===`/api/v1/terminals/${sessionId}/stream`).reduce((sum,[,count])=>sum+count,0)
    const before=produced()
    await page.keyboard.type("i=0; while [ $i -lt "+outputIterations+" ]; do printf '%01024d\\n' $i; i=$((i+1)); sleep "+outputDelay+"; done\n")
    // Establish both real commands before collecting latency. Connection/UI
    // readiness alone cannot prove that the shell received a complete command.
    await expect.poll(produced,{timeout:15000}).toBeGreaterThan(before+32*1024)
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
  // This intentionally altered visual diagnostic starts only after all real
  // resources are ready. Parser/streams/protection remain active; it is never
  // a product acceptance result even if its original latency threshold passes.
  if(renderExperiment==='no-content')await page.addStyleTag({content:'.window-body>.app-view{visibility:hidden!important}'})
  if(renderExperiment==='shadow-bands')await installShadowBandDiagnostic(page)
  if(renderExperiment==='shadow-native-resolution')await installShadowResolutionDiagnostic(page)
  if(renderExperiment==='body-viewport')await installBodyViewportDiagnostic(page)
  if(renderExperiment==='native-body-clip'){
    if(nativePartialBody){
      await installNativePartialBody(page)
      const native=await page.evaluate(()=>(window as any).__nativePartialBody.snapshot())
      console.log('BLORA_PERF_NATIVE_PARTIAL_BODY '+JSON.stringify(native))
      expect(native.tracked).toBe(8);expect(native.unmasked).toBeGreaterThan(0)
    }else{
      await installNativeBodyClip(page,nativeFlatBody)
      const native=await page.evaluate(()=>(window as any).__nativeBodyClip.snapshot())
      console.log('BLORA_PERF_NATIVE_BODY_CLIP '+JSON.stringify(native))
      expect(native.tracked).toBe(8);expect(native.enabled).toBe(8);expect(native.viewMasks).toBeGreaterThan(0)
    }
  }
  if(renderExperiment==='shadow-overflow')await page.evaluate(installShadowCropDiagnostic)
  if(['bar-shadow-cache','bar-shadow-native'].includes(renderExperiment)){
    await page.evaluate<void,'png'|'native-svg'|'native-full'>(installBarShadowDiagnostic,renderExperiment==='bar-shadow-native'?'native-full':'png')
    await expect(page.locator('.topbar')).toHaveAttribute('data-bar-shadow-cache','ready')
  }
  if(['native-shadow-composite','native-shadow-partition'].includes(renderExperiment)){
    await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
    await page.evaluate<void,'grouped'|'partitioned'>(installShadowCompositeDiagnostic,renderExperiment==='native-shadow-partition'?'partitioned':'grouped')
    await expect(page.locator('.window-shadow-plane[data-shadow-composite="ready"]')).toHaveCount(8)
  }
  if(renderExperiment==='native-shadow-alpha'){
    await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
    await page.evaluate<void,2|4>(installShadowAlphaDiagnostic,process.env.BLORA_E08_SHADOW_ALPHA_LEAVES==='4'?4:2)
    await expect(page.locator('[data-native-alpha-bands="ready"]')).toHaveCount(32)
  }
  const windowMaterialSampling=windowMaterialExperiment?await installWindowMaterialDiagnostic(page,process.env.BLORA_PERF_WINDOW_MATERIAL==='native-rgb24'?'opaque-rgb24':'native-png'):undefined
  if(renderExperiment==='native-window-position'){
    await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
    await installNativeWindowPosition(page,true,true)
    await expect(page.locator('.app-window[data-native-window-position="true"]')).toHaveCount(8)
  }
  if(renderExperiment==='native-frame-material'){
    await page.waitForFunction(()=>document.documentElement.dataset.materialCache==='ready'&&document.documentElement.dataset.materialCacheOpaque==='true')
    await installNativeFrameMaterial(page)
    await expect(page.locator('.app-window[data-native-frame-material="true"]')).toHaveCount(8)
  }
  const smallMaterialAtlas=[]
  if(smallMaterialExperiment){
    const surfaces:('window'|'wallpaper'|'desktop'|'face'|'menu')[]=process.env.BLORA_PERF_ALL_MATERIALS==='1'?['window','wallpaper','desktop','face','menu']:['window']
    for(const surface of surfaces)smallMaterialAtlas.push(await page.evaluate(installSmallMaterialAtlas,{width:Number(process.env.BLORA_PERF_MATERIAL_WIDTH||0),opaque:process.env.BLORA_PERF_OPAQUE_MATERIAL==='1',surface}))
    console.log('BLORA_PERF_SMALL_MATERIAL_ATLAS '+JSON.stringify(smallMaterialAtlas))
  }
  if(renderExperiment==='border-shadow'){
    // Diagnostic only: the same content-free native shadow, drawn through one
    // border-image object with no filled centre or forced full-size layer.
    await page.waitForFunction(()=>[...document.querySelectorAll<HTMLElement>('.window-shadow-plane')].every(plane=>!!plane.dataset.shadowCache))
    await page.evaluate(async()=>{
      const urls:string[]=[]
      for(const plane of document.querySelectorAll<HTMLElement>('.window-shadow-plane')){
        const [normal,focused,radius,scale]=JSON.parse(plane.dataset.shadowCache!) as [string,string,number,number]
        const style=getComputedStyle(plane),slice=parseFloat(style.getPropertyValue('--shadow-width')),pad=parseFloat(style.getPropertyValue('--shadow-outset')),corner=parseFloat(style.getPropertyValue('--shadow-corner')),size=2*slice+2,box=2*corner+2,pixels=Math.ceil(size*scale)
        for(const [name,shadow] of [['normal',normal],['focused',focused]]){
          const escape=(value:string)=>value.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;')
          const pixelScale=pixels/size,physicalShadow=shadow!.replace(/(-?[\d.]+)px\b/g,(_,value:string)=>`${Number(value)*pixelScale}px`)
          const svg=`<svg xmlns="http://www.w3.org/2000/svg" width="${pixels}" height="${pixels}"><foreignObject width="${pixels}" height="${pixels}"><div xmlns="http://www.w3.org/1999/xhtml" style="position:absolute;left:${pad*pixelScale}px;top:${pad*pixelScale}px;width:${box*pixelScale}px;height:${box*pixelScale}px;border-radius:${radius*pixelScale}px;box-shadow:${escape(physicalShadow)}"></div></foreignObject></svg>`
          const image=new Image();image.src='data:image/svg+xml;charset=utf-8,'+encodeURIComponent(svg);await image.decode()
          const canvas=document.createElement('canvas');canvas.width=pixels;canvas.height=pixels;canvas.getContext('2d')!.drawImage(image,0,0)
          const blob=await new Promise<Blob>((resolve,reject)=>canvas.toBlob(value=>value?resolve(value):reject(Error('diagnostic native shadow unavailable'))))
          const url=URL.createObjectURL(blob),decoded=new Image();decoded.src=url;await decoded.decode();urls.push(url)
          plane.style.setProperty(`--diagnostic-shadow-${name}`,`url("${url}")`)
        }
        plane.style.setProperty('--diagnostic-shadow-slice',String(slice*scale))
      }
      window.addEventListener('pagehide',()=>{for(const url of urls)URL.revokeObjectURL(url)},{once:true})
    })
    await page.addStyleTag({content:'.window-shadow-plane[data-shadow-cache]{border:1px solid transparent;border-image-source:var(--diagnostic-shadow-normal);border-image-slice:var(--diagnostic-shadow-slice);border-image-width:var(--shadow-width);border-image-outset:var(--shadow-outset);border-image-repeat:stretch;will-change:auto!important}.app-window.focused>.window-shadow-plane[data-shadow-cache]{border-image-source:var(--diagnostic-shadow-focused)}.app-window[data-shadow-low="true"]>.window-shadow-plane[data-shadow-cache]{border-image-source:var(--diagnostic-shadow-normal)}.window-shadow-plane[data-shadow-cache]>.shadow-piece{display:none!important}'})
  }
  if(renderExperiment==='native-shadow-sibling')await installShadowSiblings(page)
  if(renderExperiment==='native-shadow-mask')await installShadowMasks(page)
  if(renderExperiment==='native-css-plane')await page.addStyleTag({content:nativeCssPlaneShadow})
  if(renderExperiment==='native-shadow-outside')await installOutsideShadows(page)
  if(renderExperiment==='native-shadow-tail')await installShadowTail(page)
  if(['static-shadow-underlay','static-shadow-scene','static-shadow-hull'].includes(renderExperiment)){
    if(renderExperiment==='static-shadow-underlay')await installStaticShadowUnderlay(page);else await installStaticShadowScene(page,renderExperiment==='static-shadow-hull')
    const optical=await page.evaluate(()=>(window as any).__staticShadowUnderlay.snapshot())
    console.log('BLORA_PERF_STATIC_SHADOW_UNDERLAY '+JSON.stringify(optical))
    expect(optical.eligible).toBeGreaterThan(0);expect(optical.opaquePixels).toBeGreaterThan(0)
  }
  console.log('BLORA_PERF_RENDER_EXPERIMENT '+renderExperiment)
  if(renderExperiment==='native-row-relevance'){
    await installNativeRowRelevance(page)
    const native=await page.evaluate(()=>(window as any).__nativeRowRelevance.snapshot())
    console.log('BLORA_PERF_NATIVE_ROW_RELEVANCE '+JSON.stringify(native))
    expect(native.tracked).toBeGreaterThanOrEqual(2);expect(native.hiddenRows).toBeGreaterThan(0)
  }
  if(renderExperiment==='native-software-paint'||renderExperiment==='native-appview-retention'){
    await installNativeRetention(page,renderExperiment==='native-software-paint'?'software':nativeRetentionGlassOnly?'glass-app-views':'app-views')
    if(renderExperiment==='native-software-paint')await expect(page.locator('.app-window.focused')).toHaveCSS('will-change','auto')
    else expect(await page.locator('.app-window:not(.focused)>.window-body>.app-view').evaluateAll(elements=>elements.filter(element=>getComputedStyle(element).willChange==='transform').length)).toBeGreaterThanOrEqual(7)
  }
  if(renderExperiment==='native-animation-motion'){
    await installNativeMotion(page)
    const native=await page.evaluate(()=>(window as any).__nativeMotion.snapshot())
    expect(native.tracked).toBe(8);expect(native.eligible).toBe(8)
  }
  const beforeBytes=new Map(bytes),sampleStart=Date.now()
  const beforeDataMessages=terminalDataMessages
  const profileSoak=!!process.env.BLORA_PERF_PROFILE_SOAK
  const perfDiagnostics=!!process.env.BLORA_PERF_DIAGNOSTICS
  if(terminalPaintDiagnostic)await page.evaluate(()=>(window as any).__nativeTerminalPaintDiagnostic.reset())
  const profiler=process.env.BLORA_PERF_PROFILE||profileSoak?await page.context().newCDPSession(page):undefined
  if(profiler){await profiler.send('Profiler.enable');await profiler.send('Profiler.start')}
  const finishRenderTrace=process.env.BLORA_PERF_RENDER_TRACE?await startRenderTrace(page):undefined
  await page.evaluate((diagnostics:boolean)=>{
    const probe={latencies:[] as number[],initialLatencies:[] as number[],steadyLatencies:[] as number[],initialPhaseEnd:Infinity,frames:[] as number[],slowPointers:[] as Array<{latencyMs:number;inputDelayMs:number;firstFrameMs:number;secondFrameMs:number;frameAfterHandlerMs:number;eventTime:number;x:number;y:number;buttons:number}>,longTasks:[] as Array<{startTime:number;duration:number;name:string}>,longTaskObserver:undefined as PerformanceObserver|undefined,storageWrites:{count:0,totalChars:0,totalMs:0,maxChars:0,maxMs:0,durations:[] as number[]},databaseWrites:{count:0,totalMs:0,maxMs:0,durations:[] as number[],deltaCount:0,fullCount:0,deltaMs:0,fullMs:0},restoreStorage:undefined as (()=>void)|undefined,restoreDatabase:undefined as (()=>void)|undefined,running:true,last:performance.now()}
    ;(window as any).__bloraPerf=probe
    ;(probe as any).startedAt=probe.last
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
      const originalPut=IDBObjectStore.prototype.put
      IDBObjectStore.prototype.put=function(value,key){
        if(this.name!=='snapshots')return originalPut.call(this,value,key)
        const started=performance.now()
        try{return originalPut.call(this,value,key)}finally{
          const duration=performance.now()-started
          probe.databaseWrites.count++;probe.databaseWrites.totalMs+=duration;probe.databaseWrites.maxMs=Math.max(probe.databaseWrites.maxMs,duration);probe.databaseWrites.durations.push(duration)
          const delta=value&&typeof value==='object'&&value.format==='blora-terminal-delta-v1'
          if(delta){probe.databaseWrites.deltaCount++;probe.databaseWrites.deltaMs+=duration}else{probe.databaseWrites.fullCount++;probe.databaseWrites.fullMs+=duration}
        }
      }
      probe.restoreDatabase=()=>{IDBObjectStore.prototype.put=originalPut}
    }catch{probe.longTaskObserver?.disconnect();probe.longTaskObserver=undefined}
    window.addEventListener('pointermove',event=>{
      if(event.buttons!==1||!probe.running)return
      if(!diagnostics){const start=event.timeStamp;requestAnimationFrame(()=>requestAnimationFrame(()=>probe.latencies.push(performance.now()-start)));return}
      const start=event.timeStamp,handlerTime=performance.now(),inputDelayMs=Math.max(0,handlerTime-start),x=event.clientX,y=event.clientY,buttons=event.buttons
      requestAnimationFrame(()=>{const firstFrame=performance.now();requestAnimationFrame(()=>{
        const latencyMs=performance.now()-start
        probe.latencies.push(latencyMs)
        ;(start<=probe.initialPhaseEnd?probe.initialLatencies:probe.steadyLatencies).push(latencyMs)
        if(latencyMs>=50){
          probe.slowPointers.push({latencyMs,inputDelayMs,firstFrameMs:firstFrame-handlerTime,secondFrameMs:performance.now()-firstFrame,frameAfterHandlerMs:performance.now()-handlerTime,eventTime:start,x,y,buttons})
          probe.slowPointers.sort((a,b)=>b.latencyMs-a.latencyMs)
          if(probe.slowPointers.length>20)probe.slowPointers.length=20
        }
      })})
    },true)
    function frame(now:number){if(!probe.running)return;probe.frames.push(now-probe.last);probe.last=now;requestAnimationFrame(frame)}
    requestAnimationFrame(frame)
  },perfDiagnostics)
  const win=page.locator('.app-window.focused'),title=await win.locator('.window-titlebar').boundingBox()
  const original=await win.evaluate(element=>(element as HTMLElement).style.transform)
  await page.mouse.move(title!.x+240,title!.y+15);await page.mouse.down()
  for(let i=0;i<120;i++)await page.mouse.move(title!.x+240+Math.sin(i/10)*80,title!.y+15+Math.cos(i/10)*30)
  await page.mouse.up()
  if(finishRenderTrace){
    const rendering=await finishRenderTrace()
    console.log('BLORA_PERF_RENDER_TRACE '+JSON.stringify(rendering))
    await info.attach('render-timing-summary.json',{body:JSON.stringify(rendering,null,2),contentType:'application/json'})
  }
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
  expect(await win.evaluate(element=>{(window as any).__bloraPerf.initialPhaseEnd=performance.now();return (element as HTMLElement).style.transform})).not.toBe(original)
  if(terminalPaintDiagnostic)console.log('BLORA_PERF_TERMINAL_NATIVE_PAINT_INITIAL '+JSON.stringify(await page.evaluate(()=>(window as any).__nativeTerminalPaintDiagnostic.snapshot())))
  const soakHeap:number[]=[]
  // Additional long-run diagnostics use numeric native heap counters only;
  // no snapshots, app values or forced GC. Original 60s acceptance is unchanged.
  const preciseHeap=process.env.BLORA_PERF_PRECISE_HEAP==='1'
  if(preciseHeap){expect(soakSeconds).toBeGreaterThan(60);expect(browser.browserType().name()).toBe('chromium')}
  const heapProbe=preciseHeap?await page.context().newCDPSession(page):undefined
  if(soakSeconds>0){
    const soakUntil=Date.now()+soakSeconds*1000
    let lastProgress=Date.now()
    let lastPointerSample=0
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
        if(longMotion==='bounded')await page.mouse.move(x,y)
        await page.mouse.up()
      }
      const heap=heapProbe?Number((await heapProbe.send('Runtime.getHeapUsage')).usedSize):await page.evaluate(()=>Number((performance as any).memory?.usedJSHeapSize||0))
      if(heap>0)soakHeap.push(heap)
      if(Date.now()-lastProgress>=60000){
        const terminalIntervalBytes=[...bytes].map(([key,value])=>value-(lastTerminalBytes.get(key)||0)).filter(value=>value>0)
        const pointerInterval=await page.evaluate(start=>{
          const all=(window as any).__bloraPerf.latencies,values=(all.slice(start) as number[]).sort((a,b)=>a-b)
          return {end:all.length,samples:values.length,p95:values[Math.ceil(values.length*.95)-1]||0,max:values.at(-1)||0}
        },lastPointerSample)
        lastPointerSample=pointerInterval.end
        console.log('BLORA_PERF_PROGRESS '+JSON.stringify({elapsedSeconds:Math.round((Date.now()-sampleStart)/1000),completedTransfers,activeTransferSamples,heapBytes:heap,maxUnacknowledgedBytes,terminalIntervalBytes,pointerInterval}))
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
  if(heapProbe){await heapProbe.detach();expect(soakHeap.length).toBeGreaterThan(1)}
  if(profileSoak)await finishProfile()
  if(soakSeconds>0)await maintainTransfer()
  const stateAfter=(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state
  const sampleSeconds=(Date.now()-sampleStart)/1000
  const terminalBytesPerSecond=[...bytes].map(([key,value])=>(value-(beforeBytes.get(key)||0))/sampleSeconds).filter(value=>value>0)
  const sample=await page.evaluate(async(diagnostics:boolean)=>{
    const probe=(window as any).__bloraPerf;probe.running=false;probe.longTaskObserver?.disconnect()
    probe.restoreStorage?.();probe.restoreDatabase?.()
    const rtts:number[]=[]
    for(let i=0;i<5;i++){const start=performance.now();const response=await fetch('/api/v1/session');if(!response.ok)throw Error('RTT request failed');await response.json();rtts.push(performance.now()-start)}
    const p95=(values:number[])=>values.sort((a,b)=>a-b)[Math.ceil(values.length*.95)-1]||0
    const writes=probe.storageWrites,database=probe.databaseWrites
    return {samples:probe.latencies.length,p95:p95(probe.latencies),max:Math.max(...probe.latencies),frames:probe.frames.length,longFrames:probe.frames.filter((value:number)=>value>50).length,...(diagnostics?{phaseTiming:{timeOriginMs:performance.timeOrigin,initialFromMs:probe.startedAt,initialUntilMs:probe.initialPhaseEnd},pointerPhases:{initial:{samples:probe.initialLatencies.length,p95:p95(probe.initialLatencies)},steady:{samples:probe.steadyLatencies.length,p95:p95(probe.steadyLatencies)}},frameP95Ms:p95(probe.frames),slowPointerResponses:probe.slowPointers,longTasks:probe.longTasks,recoveryJournalWrites:{count:writes.count,totalChars:writes.totalChars,totalMs:writes.totalMs,maxChars:writes.maxChars,maxMs:writes.maxMs,p95Ms:p95(writes.durations)},recoverySnapshotClones:{count:database.count,totalMs:database.totalMs,maxMs:database.maxMs,p95Ms:p95(database.durations),deltaCount:database.deltaCount,fullCount:database.fullCount,deltaMs:database.deltaMs,fullMs:database.fullMs}}:{}),heapBytes:(performance as any).memory?.usedJSHeapSize,apiRttP95:p95(rtts)}
  },perfDiagnostics)
  const soak={seconds:soakSeconds,longMotion,outputDelaySeconds:outputDelay,completedTransfers,activeTransferSamples,maxTransferRestartMs,heapMeasurement:preciseHeap?'native-cdp':soakHeap.length?'performance-memory-rounded':'unavailable',heapSamples:soakHeap.length,heapFirst:soakHeap[0],heapLast:soakHeap.at(-1),heapMin:soakHeap.length?Math.min(...soakHeap):undefined,heapMax:soakHeap.length?Math.max(...soakHeap):undefined,heapDelta:soakHeap.length?Math.max(...soakHeap)-Math.min(...soakHeap):undefined}
  console.log('BLORA_REAL_PERF_SAMPLE '+JSON.stringify({...sample,terminalBytesPerSecond,stateAfter,soak}))
  if(renderExperiment==='native-animation-motion'){
    const native=await page.evaluate(()=>(window as any).__nativeMotion.snapshot())
    console.log('BLORA_PERF_NATIVE_MOTION '+JSON.stringify(native))
    expect(native.updates).toBeGreaterThanOrEqual(120);expect(native.maxGeometryError).toBeLessThan(.01)
  }
  await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:transferTimeout}).toBe('SUCCEEDED')
  const detail=await(await page.request.get(`/api/v1/transfers/${taskId}`)).json()
  expect(detail.transfer.destinationVerified).toBe(true)
  const graphics=await page.evaluate(()=>{const gl=document.createElement('canvas').getContext('webgl2'),extension=gl?.getExtension('WEBGL_debug_renderer_info');const renderer=extension?gl!.getParameter(extension.UNMASKED_RENDERER_WEBGL):'unavailable';gl?.getExtension('WEBGL_lose_context')?.loseContext();return renderer})
  const shadowCache=await page.locator('.window-shadow-plane').evaluateAll(elements=>({cached:elements.filter(element=>(element as HTMLElement).dataset.shadowCache).length,total:elements.length}))
  const retainedWindowLayers=await page.locator('.app-window').evaluateAll(elements=>elements.filter(element=>getComputedStyle(element).willChange.split(',').some(value=>value.trim()==='transform')).length)
  const retainedAppViewLayers=await page.locator('.app-window>.window-body>.app-view').evaluateAll(elements=>elements.filter(element=>getComputedStyle(element).willChange.split(',').some(value=>value.trim()==='transform')).length)
  const retainedShadowEdges=await page.locator('.shadow-piece').evaluateAll(elements=>elements.filter(element=>getComputedStyle(element).display!=='none'&&getComputedStyle(element).willChange.split(',').some(value=>value.trim()==='transform')).length)
  const retainedShadowPlanes=await page.locator('.window-shadow-plane').evaluateAll(elements=>elements.filter(element=>getComputedStyle(element).willChange.split(',').some(value=>value.trim()==='transform')).length)
  const shadowPatchOcclusion=await page.locator('.window-shadow-plane[data-shadow-cache]>.shadow-piece:not([data-shadow-empty="true"])').evaluateAll(elements=>{
    // CSS geometry only; this is not a claimed GPU allocation measurement.
    let totalPixels=0,culledPixels=0,culled=0,clipped=0,clippedPixels=0
    for(const element of elements){
      const box=element.getBoundingClientRect(),pixels=box.width*box.height,style=(element as HTMLElement).style,clip=style.clipPath;totalPixels+=pixels
      if((element as HTMLElement).dataset.shadowOccluded==='true'||clip==='inset(100%)'){culled++;culledPixels+=pixels;clippedPixels+=pixels;continue}
      const match=/^inset\(([-\d.]+)px ([-\d.]+)px ([-\d.]+)px ([-\d.]+)px\)$/.exec(clip)
      if(match){
        const [top,right,bottom,left]=match.slice(1).map(Number),visible=Math.max(0,box.width-Math.max(0,left!)-Math.max(0,right!))*Math.max(0,box.height-Math.max(0,top!)-Math.max(0,bottom!))
        if(visible<pixels){clipped++;clippedPixels+=pixels-visible}
      }
      const native=/^rect\(([-\d.]+)px,?\s+([-\d.]+)px,?\s+([-\d.]+)px,?\s+([-\d.]+)px\)$/.exec(style.clip)
      if(native){const [top,right,bottom,left]=native.slice(1).map(Number),visible=Math.max(0,Math.min(box.width,right!)-Math.max(0,left!))*Math.max(0,Math.min(box.height,bottom!)-Math.max(0,top!));if(visible<pixels){clipped++;clippedPixels+=pixels-visible}}
    }
    return {total:elements.length,culled,clipped,totalPixels,culledPixels,clippedPixels}
  })
  const result={...sample,renderExperiment,maxUnacknowledgedBytes,graphics,terminalRenderers:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalRenderer).filter(Boolean)),terminalPaintOcclusion:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.paintOccluded==='true')),terminalTextRasters:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalTextRaster||'native')),materialCache:await page.evaluate(()=>document.documentElement.dataset.materialCache||'fallback'),terminalCheckpoints:await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalCheckpoint).filter(Boolean)),terminalBytesPerSecond,sampleSeconds,stateAfter,soak,windows:await page.locator('.app-window').count(),directoryEntries:10000,transferBytes:stat.size,transferSeconds:(Date.now()-transferStart)/1000,browser:browser.version(),cpu:cpus()[0]?.model,logicalCPUs:cpus().length,hostMemoryBytes:totalmem()}
  const nativeListPaintOcclusion=await page.locator('[data-app-paint-occluded="true"]').evaluateAll(elements=>({blocks:elements.length,cssPixels:elements.reduce((sum,element)=>{const box=element.getBoundingClientRect();return sum+box.width*box.height},0)}))
  const shadowSegmentOcclusion=await page.locator('.shadow-segment').evaluateAll(elements=>({segments:elements.length,culled:elements.filter(element=>(element as HTMLElement).dataset.shadowSegmentOccluded==='true').length,culledCssPixels:elements.reduce((sum,element)=>{if((element as HTMLElement).dataset.shadowSegmentOccluded!=='true')return sum;const box=element.getBoundingClientRect();return sum+box.width*box.height},0)}))
  const dockOpticalCache=await page.locator('.dock-surface').evaluate(element=>{const image=element.querySelector<HTMLImageElement>('.dock-optical-cache');return {ready:(element as HTMLElement).dataset.dockCache==='ready',decoded:!!image?.complete&&!!image.naturalWidth,width:image?.naturalWidth||0,height:image?.naturalHeight||0}})
  Object.assign(result,{shadowCache,shadowPatchOcclusion,nativeListPaintOcclusion,shadowSegmentOcclusion,dockOpticalCache,terminalRendererExperiment,retainedWindowLayers,retainedShadowEdges,retainedShadowPlanes,nativeRetentionGlassOnly,retainedAppViewLayers,nativeFlatBody,nativePartialBody,nativeTransformControl})
  Object.assign(result,{materialCacheEncoding:await page.evaluate(()=>document.documentElement.dataset.materialCacheEncoding||'fallback')})
  Object.assign(result,{recoveryPersistenceMode:await page.locator('.desktop').getAttribute('data-recovery-writer')||'foreground'})
  Object.assign(result,{terminalDataMessages:terminalDataMessages-beforeDataMessages})
  const terminalPartialPaint=await terminals.evaluateAll(elements=>elements.map(element=>{
    const host=element as HTMLElement,region=host.closest<HTMLElement>('.terminal-paint-region'),viewport=host.closest<HTMLElement>('.terminal-paint-viewport')
    const full=region?.getBoundingClientRect(),paint=viewport?.getBoundingClientRect()
    return {partial:host.dataset.paintPartial==='true',fullPixels:full?full.width*full.height:0,paintPixels:paint?paint.width*paint.height:0}
  }))
  const terminalNativeRowPaint=await terminals.evaluateAll(elements=>elements.map(element=>(element as HTMLElement).dataset.terminalRowPaint||'native-full'))
  Object.assign(result,{terminalPartialPaint,terminalNativeRowPaint})
  const windowPositioning=await page.locator('.app-window').evaluateAll(elements=>({selectedOffsets:elements.filter(element=>(element as HTMLElement).dataset.windowPosition==='offset').length,nativeOffsets:elements.filter(element=>{const style=getComputedStyle(element),matrix=new DOMMatrixReadOnly(style.transform);return (element as HTMLElement).dataset.windowPosition==='offset'&&matrix.e===0&&matrix.f===0}).length}))
  Object.assign(result,{windowPositioning})
  if(renderExperiment==='native-row-relevance')Object.assign(result,{nativeRowRelevance:await page.evaluate(()=>(window as any).__nativeRowRelevance.snapshot())})
  if(renderExperiment==='native-frame-material')Object.assign(result,{nativeFrameMaterial:await page.evaluate(()=>(window as any).__nativeFrameMaterial.snapshot())})
  if(renderExperiment==='native-window-position')Object.assign(result,{nativeWindowPosition:await page.evaluate(()=>(window as any).__nativeWindowPosition.snapshot())})
  if(terminalPaintDiagnostic){
    const terminalNativePaint=await page.evaluate(()=>(window as any).__nativeTerminalPaintDiagnostic.snapshot())
    console.log('BLORA_PERF_TERMINAL_NATIVE_PAINT '+JSON.stringify(terminalNativePaint));Object.assign(result,{terminalNativePaint})
  }
  if(['native-shadow-composite','native-shadow-partition'].includes(renderExperiment)){
    const shadowComposition=await page.locator('.shadow-composite-image').evaluateAll(elements=>({strips:elements.length,decoded:elements.filter(element=>{const image=element as HTMLImageElement;return image.complete&&image.naturalWidth>0}).length,cssPixels:elements.reduce((sum,element)=>{const rect=element.getBoundingClientRect();return sum+rect.width*rect.height},0)}))
    expect(shadowComposition.decoded).toBe(32);Object.assign(result,{shadowComposition})
    if(renderExperiment==='native-shadow-partition')expect(shadowComposition.cssPixels).toBeLessThan(shadowPatchOcclusion.totalPixels)
  }
  if(renderExperiment==='native-shadow-alpha'){
    const shadowAlpha=await page.locator('[data-native-alpha-bands="ready"]').evaluateAll(elements=>{
      let originalPixels=0,partitionPixels=0,visiblePixels=0,bands=0
      for(const element of elements){const rect=element.getBoundingClientRect();originalPixels+=rect.width*rect.height;for(const child of element.children){const box=child.getBoundingClientRect(),pixels=box.width*box.height;partitionPixels+=pixels;bands++;if((element as HTMLElement).style.clipPath!=='inset(100%)')visiblePixels+=pixels}}
      return {corners:elements.length,bands,originalPixels,partitionPixels,visiblePixels}
    })
    expect(shadowAlpha.corners).toBe(32);expect(shadowAlpha.partitionPixels).toBeLessThan(shadowAlpha.originalPixels)
    console.log('BLORA_PERF_SHADOW_ALPHA '+JSON.stringify(shadowAlpha))
    Object.assign(result,{shadowAlpha})
  }
  if(windowMaterialSampling)Object.assign(result,{windowMaterialSampling,windowMaterialExperiment:process.env.BLORA_PERF_WINDOW_MATERIAL})
  if(renderExperiment==='shadow-overflow'){
    const nativeShadowCrop=await page.locator('.native-shadow-crop').evaluateAll(elements=>({viewports:elements.length,savedCssPixels:elements.reduce((sum,element)=>{const full=element.parentElement!.getBoundingClientRect(),crop=element.getBoundingClientRect();return sum+Math.max(0,full.width*full.height-crop.width*crop.height)},0)}))
    expect(nativeShadowCrop.viewports).toBeGreaterThan(0)
    expect(nativeShadowCrop.savedCssPixels).toBeGreaterThan(0)
    Object.assign(result,{nativeShadowCrop})
  }
  if(terminalRendererExperiment==='native-dom-equality'){
    const nativeDomEquality=await page.evaluate(()=>(window as any).__nativeDomEquality as {calls:number;reused:number;replaced:number})
    expect(nativeDomEquality.reused).toBeGreaterThan(0)
    Object.assign(result,{nativeDomEquality})
  }
  if(['shadow-nearest','shadow-nearest-all'].includes(renderExperiment)){
    const selector=renderExperiment==='shadow-nearest-all'?'':':is(.n,.s,.w,.e)'
    const nativeShadowSampling=await page.locator('.window-shadow-plane[data-shadow-cache]>.shadow-piece'+selector).evaluateAll(elements=>({edges:elements.length,pointSampled:elements.filter(element=>getComputedStyle(element).imageRendering==='pixelated').length}))
    expect(nativeShadowSampling.pointSampled).toBeGreaterThan(0)
    Object.assign(result,{nativeShadowSampling})
  }
  await info.attach('real-performance.json',{body:JSON.stringify(result,null,2),contentType:'application/json'})
  if(['static-shadow-underlay','static-shadow-scene','static-shadow-hull'].includes(renderExperiment))Object.assign(result,{staticShadowUnderlay:await page.evaluate(()=>(window as any).__staticShadowUnderlay.snapshot())})
  console.log('BLORA_REAL_PERF '+JSON.stringify(result))
  expect(errors).toEqual([])
  expect(stateAfter).toBe('RUNNING')
  expect(terminalBytesPerSecond).toHaveLength(2)
  expect(maxUnacknowledgedBytes).toBeLessThanOrEqual(256*1024+65536)
  expect(sample.samples).toBeGreaterThanOrEqual(100)
  expect(sample.p95).toBeLessThanOrEqual(50)
})
