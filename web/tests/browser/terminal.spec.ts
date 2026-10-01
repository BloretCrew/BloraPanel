import {test,expect} from '../helpers/management-fixture'
import {decodeEnvelope,encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'

// Transport doubles are confined to this test. The xterm parser, serializer,
// browser recovery storage and binary protocol implementation are real.
for(const [forceFallback,forceSnapshotFallback] of [[false,false],[true,false],[true,true]])test(`xterm checkpoints resume the same session, ACK parsed bytes and never replay input (${forceFallback?'fallback':'automatic'} renderer, ${forceSnapshotFallback?'main':'worker'} checkpoints)`,async({page})=>{
  if(forceSnapshotFallback)await page.addInitScript(()=>{
    const Original=Worker
    window.Worker=class extends Original{
      constructor(url:string|URL,options?:WorkerOptions){
        if(String(url).includes('terminal-snapshot'))throw new Error('test: checkpoint Worker unavailable')
        super(url,options)
      }
    }
  })
  if(forceFallback)await page.addInitScript(()=>{
    const original=HTMLCanvasElement.prototype.getContext
    HTMLCanvasElement.prototype.getContext=function(this:HTMLCanvasElement,...args:any[]){
      if(args[0]==='webgl2'&&args[1]?.failIfMajorPerformanceCaveat)return null
      return original.apply(this,args as any)
    } as typeof original
  })
  const received:{type:MessageType;text?:string;sequence:number;credit:number}[]=[],resumeCursors:number[]=[],errors:string[]=[]
  let creates=0,connection=0,releaseReplay:()=>void=()=>{},liveOutput:(sequence:number,text:string)=>void=()=>{}
  let rejectStream:()=>void=()=>{},liveResize:(sequence:number)=>void=()=>{}
  page.on('pageerror',error=>errors.push(error.message))
  const resource={kind:'instance',id:'terminal-instance',nodeId:'terminal-node'}
  const session={sessionId:'terminal-fixture',state:'RUNNING',resource,backend:'test-transport',cols:100,rows:28,archive:{maxBytes:16777216},maxSessions:64,maxAttachments:16}
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(path.endsWith('/terminals')&&route.request().method()==='POST')creates++
    const json=path.endsWith('/session')?{user:{userId:'terminal-browser-test',name:'终端测试',admin:true},csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:path.endsWith('/instances')?{items:[{instanceId:resource.id,nodeId:resource.nodeId,nodeName:'终端测试节点',nodeState:'ONLINE',name:'终端测试实例',state:'RUNNING',revision:1,config:{mode:'native'}}]}:path.endsWith('/terminals')?{items:[session]}:{items:[]}
    return route.fulfill({json})
  })
  await page.routeWebSocket(/\/api\/v1\/terminals\/terminal-fixture\/stream/,ws=>{
    const index=++connection,after=Number(new URL(ws.url()).searchParams.get('sequence'));resumeCursors.push(after)
    let sentBytes=0,acknowledgedBytes=0
    const send=(type:MessageType,payload?:Uint8Array,sequence=0,credit=0)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:session.sessionId,type,payload,sequence,credit})))
    const info=(writable:boolean)=>send(MessageType.OpenAck,encodeJSON({sessionId:session.sessionId,writable,earliest:1,latest:1,state:'RUNNING'}))
    rejectStream=()=>send(MessageType.Error,encodeJSON({code:'OUTPUT_GAP',message:'测试归档缺口；检查点保留'}))
    ws.onMessage(message=>{
      const envelope=decodeEnvelope(new Uint8Array(message as Buffer));received.push({type:envelope.type,text:envelope.payload?new TextDecoder().decode(envelope.payload):undefined,sequence:envelope.sequence||0,credit:envelope.credit||0})
      if(envelope.type===MessageType.Ack){
        if(envelope.sequence!==acknowledgedBytes+(envelope.credit||0)||envelope.sequence>sentBytes)errors.push('ACK did not match parsed cumulative bytes')
        acknowledgedBytes=envelope.sequence||0
      }
      if(envelope.type===MessageType.Data)send(MessageType.Ack,undefined,envelope.sequence,envelope.payload!.length)
      if(envelope.type===MessageType.Open)info(true)
    })
    info(index===1)
    if(!after){const payload=encodeJSON({sequence:1,kind:'output',data:Buffer.from('\u001b[?1049h\u001b[2J\u001b[H中文屏幕\r\n\u001b[6n').toString('base64')});sentBytes+=payload.byteLength;send(MessageType.Data,payload,sentBytes)}
    liveOutput=(sequence,text)=>{const payload=encodeJSON({sequence,kind:'output',data:Buffer.from(text).toString('base64')});sentBytes+=payload.byteLength;send(MessageType.Data,payload,sentBytes)}
    liveResize=sequence=>{const payload=encodeJSON({sequence,kind:'resize',cols:80,rows:24});sentBytes+=payload.byteLength;send(MessageType.Data,payload,sentBytes)}
    if(index===1)releaseReplay=()=>send(MessageType.Resume)
    else send(MessageType.Resume)
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'终端测试实例',exact:true}).click();await page.getByRole('button',{name:'控制台',exact:true}).click();await page.getByRole('button',{name:'打开终端会话管理',exact:true}).click()
  await page.getByRole('button',{name:/terminal-fixture · RUNNING/}).click()
  await expect.poll(()=>received.filter(message=>message.type===MessageType.Ack).length).toBe(1)
  expect(received.filter(message=>message.type===MessageType.Data)).toEqual([])
  const ack=received.find(message=>message.type===MessageType.Ack)!;expect(ack.sequence).toBe(ack.credit)
  releaseReplay();await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','true')
  await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-checkpoint',forceSnapshotFallback?'main':'worker')
  if(forceFallback)await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-renderer','default')
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type('abc')
  await expect.poll(()=>received.filter(message=>message.type===MessageType.Data).map(message=>message.text).join('')).toBe('abc')
  liveOutput(2,'追加增量');await expect.poll(()=>received.filter(message=>message.type===MessageType.Ack).length).toBe(2)
  await page.reload();await expect(page.getByText('已连接 · 只读观察',{exact:true})).toBeVisible();expect(resumeCursors.at(-1)).toBe(2)
  await expect.poll(()=>page.evaluate(async()=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
    return snapshots.some(snapshot=>Object.values(snapshot.terminals||{}).some((value:any)=>value.sequence===2&&value.screen.includes('追加增量')))
  })).toBe(true)
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type('blocked');await page.getByRole('button',{name:'申请接管'}).click();await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible()
  expect(received.filter(message=>message.type===MessageType.Data).map(message=>message.text).join('')).toBe('abc')
  // Cross the incremental journal and synchronous tail budgets with actual
  // parsed output. ACK must await the durable full snapshot when necessary.
  for(let sequence=3;sequence<=26;sequence++){
    liveOutput(sequence,(sequence===3?'\u001b[?1049l':'')+`${'scrollback '.padEnd(100,'x')}\r\n`.repeat(160)+`checkpoint-${sequence}\r\n`)
    await expect.poll(()=>received.filter(message=>message.type===MessageType.Ack).length).toBe(sequence)
  }
  const viewId=await page.locator('.app-window.focused [data-view-tab]').getAttribute('data-view-tab')
  await page.locator('.app-window.focused [aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'移到新窗口',exact:true}).click();await expect(page.locator(`[data-view-tab="${viewId}"]`)).toHaveCount(1)
  await expect.poll(()=>connection).toBe(3);expect(resumeCursors.at(-1)).toBe(26);expect(creates).toBe(0);expect(received.filter(message=>message.type===MessageType.Data).map(message=>message.text).join('')).toBe('abc');expect(errors).toEqual([])
  await expect.poll(()=>page.evaluate(async()=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
    return snapshots.some(snapshot=>Object.values(snapshot.terminals||{}).some((value:any)=>value.sequence===26&&value.screen.length>128*1024&&value.screen.includes('checkpoint-26')))
  })).toBe(true)
  // A burst crosses an ordered resize event. Cumulative ACK credit must match
  // all parsed bytes, and a refresh must resume after the last original event.
  for(let sequence=27;sequence<=90;sequence++){
    if(sequence===40)liveResize(sequence)
    else liveOutput(sequence,'x'.repeat(512)+` burst-${sequence}\r\n`)
  }
  await expect.poll(()=>page.evaluate(async()=>{
    const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
    const snapshots=await new Promise<any[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
    return snapshots.some(snapshot=>Object.values(snapshot.terminals||{}).some((value:any)=>value.sequence===90&&value.cols===80))
  })).toBe(true)
  await page.reload();await expect(page.getByText('已连接 · 只读观察',{exact:true})).toBeVisible()
  expect(resumeCursors.at(-1)).toBe(90)
  const acknowledgements=received.filter(message=>message.type===MessageType.Ack)
  expect(acknowledgements.every(message=>message.credit>0&&message.credit<=message.sequence)).toBe(true)
  rejectStream()
  await expect(page.getByRole('alert')).toContainText('OUTPUT_GAP：测试归档缺口；检查点保留')
  await expect(page.getByText('连接已停止 · 原检查点保留',{exact:true})).toBeVisible()
  await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','false')
  expect(errors).toEqual([])
  // Failed local durability must withhold credit for the newly parsed batch.
  // Both browser storage layers fail; a successful synchronous tail must not
  // accidentally conceal the IndexedDB failure being exercised here.
  await page.getByRole('button',{name:'重新挂载',exact:true}).click()
  await expect(page.getByText('已连接 · 只读观察',{exact:true})).toBeVisible()
  const ackCount=received.filter(message=>message.type===MessageType.Ack).length
  await page.evaluate(()=>{
    const set=Storage.prototype.setItem,put=IDBObjectStore.prototype.put
    Storage.prototype.setItem=function(key,value){if(key.startsWith('blora:tail:'))throw new DOMException('terminal quota test','QuotaExceededError');return set.call(this,key,value)}
    IDBObjectStore.prototype.put=function(...args:Parameters<IDBObjectStore['put']>){if(this.name==='snapshots')throw new DOMException('terminal quota test','QuotaExceededError');return put.apply(this,args)}
  })
  for(let sequence=91;sequence<=110;sequence++)liveOutput(sequence,`unprotected-${sequence}\r\n`)
  await expect(page.getByRole('alert')).toContainText('终端输出尚未持久保护')
  expect(received.filter(message=>message.type===MessageType.Ack)).toHaveLength(ackCount)
  expect(errors).toEqual([])
})
