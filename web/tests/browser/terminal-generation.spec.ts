import {test,expect,type WebSocketRoute} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'

test('a late parse failure from a replaced terminal connection cannot revoke the new connection',async({page})=>{
  const resource={kind:'instance',id:'generation-instance',nodeId:'generation-node'}
  const session={sessionId:'generation-terminal',state:'RUNNING',resource,backend:'test-transport',cols:100,rows:28,archive:{maxBytes:16777216},maxSessions:64,maxAttachments:16}
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'generation-user',name:'测试',admin:true},csrfToken:'test-only'}:path.endsWith('/instances')?{items:[{instanceId:resource.id,nodeId:resource.nodeId,nodeName:'测试节点',nodeState:'ONLINE',name:'代次测试实例',state:'RUNNING',revision:1,config:{mode:'native'}}]}:path.endsWith('/terminals')?{items:[session]}:{items:[]}})
  })
  const connections:WebSocketRoute[]=[],inputs:string[]=[]
  const send=(ws:WebSocketRoute,type:MessageType,payload?:Uint8Array,sequence=0)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:session.sessionId,type,payload,sequence})))
  await page.routeWebSocket(/\/api\/v1\/terminals\/generation-terminal\/stream/,ws=>{
    connections.push(ws)
    ws.onMessage(message=>{const frame=decodeEnvelope(new Uint8Array(message as Buffer));if(frame.type===MessageType.Data)inputs.push(new TextDecoder().decode(frame.payload))})
    send(ws,MessageType.OpenAck,encodeJSON({sessionId:session.sessionId,writable:true,earliest:1,latest:0,state:'RUNNING'}))
    send(ws,MessageType.Resume)
  })
  await page.goto('/')
  await page.evaluate(async()=>{
    const path='/src/services/terminals.ts'
    const {TerminalModel}=await import(path) as typeof import('../../src/services/terminals')
    const original=TerminalModel.prototype.receiveBatch
    const probe={entered:false,release:()=>{}}
    TerminalModel.prototype.receiveBatch=function(events){
      if(events.some(event=>event.kind==='output'&&event.data===btoa('held-output'))){
        probe.entered=true
        return new Promise<void>((_resolve,reject)=>{probe.release=()=>reject(new Error('old connection parse failure'))})
      }
      return original.call(this,events)
    }
    ;(window as unknown as {generationProbe:typeof probe}).generationProbe=probe
  })
  await page.locator('[data-app="blora.instances"]').click()
  await page.getByRole('button',{name:'代次测试实例',exact:true}).click()
  await page.getByRole('button',{name:'控制台',exact:true}).click()
  await page.getByRole('button',{name:'打开终端会话管理',exact:true}).click()
  await page.getByRole('button',{name:/generation-terminal · RUNNING/}).click()
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible()
  const payload=encodeJSON({sequence:1,kind:'output',data:Buffer.from('held-output').toString('base64')})
  send(connections[0]!,MessageType.Data,payload,payload.length)
  await expect.poll(()=>page.evaluate(()=>(window as unknown as {generationProbe:{entered:boolean}}).generationProbe.entered)).toBe(true)
  await page.getByRole('button',{name:'重新挂载',exact:true}).click()
  await expect.poll(()=>connections.length).toBe(2)
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible()
  await page.evaluate(async()=>{
    ;(window as unknown as {generationProbe:{release:()=>void}}).generationProbe.release()
    await new Promise<void>(resolve=>setTimeout(resolve,0))
  })
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible()
  await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','true')
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type('live')
  await expect.poll(()=>inputs.join('')).toBe('live')
  send(connections[1]!,MessageType.Error,encodeJSON({code:'CURRENT_FAILURE',message:'current connection failed'}))
  await expect(page.getByText('连接已停止 · 原检查点保留',{exact:true})).toBeVisible()
  await expect(page.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','false')
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type('blocked')
  expect(inputs.join('')).toBe('live')
})
