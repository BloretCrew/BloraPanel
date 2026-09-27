import {test,expect} from '@playwright/test'
import {decodeEnvelope,encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'
test('run logs resume split UTF-8 checkpoints, bound rendered lines and never submit command drafts',async({page})=>{
  const cursors:number[]=[],received:MessageType[]=[],errors:string[]=[],writes:string[]=[]
  let connection=0,push:(sequence:number,text:string)=>void=()=>{},gap:()=>void=()=>{},finish:()=>void=()=>{}
  const encoded=new TextEncoder().encode('中文')
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{const path=new URL(route.request().url()).pathname;if(route.request().method()!=='GET')writes.push(path);return route.fulfill({json:path.endsWith('/session')?{user:{userId:'log-user',name:'日志观察者',admin:false},csrfToken:'test-only'}:path.endsWith('/instances')?{items:[{instanceId:'log-instance',nodeId:'log-node',nodeName:'日志节点',name:'日志实例',state:'RUNNING',runId:'log-run',config:{mode:'native'},revision:1}]}:path.endsWith('/logs')?{items:[{runId:'log-run',resource:{kind:'instance',id:'log-instance',nodeId:'log-node'},backend:'test-boundary',startedAt:'2026-09-09T00:00:00Z'}]}:{items:[]}})})
  await page.routeWebSocket(/\/instances\/log-instance\/logs\/log-run\/stream/,ws=>{
    const index=++connection;let sent=0;cursors.push(Number(new URL(ws.url()).searchParams.get('sequence')))
    const send=(type:MessageType,payload?:Uint8Array,sequence=0)=>ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:'log-run',type,payload,sequence})))
    const data=(sequence:number,bytes:Uint8Array)=>{const payload=encodeJSON({sequence,kind:'output',data:Buffer.from(bytes).toString('base64')});sent+=payload.length;send(MessageType.Data,payload,sent)}
    ws.onMessage(message=>received.push(decodeEnvelope(new Uint8Array(message as Buffer)).type))
    send(MessageType.OpenAck,encodeJSON({runId:'log-run',status:{phase:'capturing',earliest:1,latest:index,bytes:6},inputAvailable:false}));if(index===1)data(1,encoded.slice(0,5));else data(2,encoded.slice(5))
    push=(sequence,text)=>data(sequence,new TextEncoder().encode(text));gap=()=>send(MessageType.Reset,encodeJSON({code:'OUTPUT_GAP',message:'归档存在缺口',earliest:20,latest:20}));finish=()=>send(MessageType.Close,encodeJSON({phase:'complete',latest:20,earliest:20,bytes:1000}))
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'日志实例',exact:true}).click();await page.getByRole('button',{name:'控制台',exact:true}).click();await expect(page.getByRole('log',{name:'实例运行日志'})).toContainText('中');await expect.poll(()=>received.filter(type=>type===MessageType.Ack).length).toBe(1)
  await page.getByRole('textbox',{name:'尚未发送的实例命令'}).fill('独立命令草稿');await page.getByRole('textbox',{name:'搜索保留日志'}).fill('中');await page.getByRole('button',{name:'暂停跟随输出',exact:true}).click();await page.reload()
  await expect(page.getByRole('log',{name:'实例运行日志'})).toContainText('中文');await expect(page.getByRole('textbox',{name:'尚未发送的实例命令'})).toHaveValue('独立命令草稿');await expect(page.getByRole('textbox',{name:'搜索保留日志'})).toHaveValue('中');await expect(page.getByRole('button',{name:'恢复跟随输出',exact:true})).toBeVisible();expect(cursors).toEqual([0,1]);expect(received).not.toContain(MessageType.Data);await expect(page.getByRole('button',{name:'发送命令',exact:true})).toBeDisabled()
  await page.getByRole('textbox',{name:'搜索保留日志'}).fill('');push(3,'\n'+Array.from({length:10000},(_,index)=>`条目 ${index}`).join('\n'));await expect(page.locator('.instance-console>.warning')).toBeVisible();expect(await page.locator('.console-log-line').count()).toBeLessThanOrEqual(100)
  await page.getByRole('button',{name:'恢复跟随输出',exact:true}).click();await expect(page.getByRole('log',{name:'实例运行日志'})).toContainText('条目 9999');gap();push(20,'\n缺口后继续');await expect(page.getByRole('log',{name:'实例运行日志'})).toContainText('缺口后继续');await expect(page.locator('.instance-console .error')).toContainText('归档存在缺口');finish();await expect(page.locator('.instance-console')).toContainText('日志归档已完成');expect(writes).toEqual([]);expect(received).not.toContain(MessageType.Data);expect(errors).toEqual([])
})

test('instance command delivery preserves the original run and request without replay on refresh',async({page})=>{
  const posts:{key:string;path:string;body:{data:string}}[]=[],sent:MessageType[]=[];let completed=false,known=false
  const task=()=>({taskId:'input-task',requestId:posts[0]?.key,action:'console.input',resource:{kind:'instance',id:'input-instance'},state:completed?'SUCCEEDED':'RUNNING',phase:'stdin',result:completed?{runId:'input-run',bytesWritten:Buffer.byteLength(posts[0]!.body.data),deliveryOnly:true}:undefined})
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname;let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'input-user',name:'输入测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'input-instance',nodeId:'input-node',name:'输入实例',state:'RUNNING',runId:'input-run',config:{},revision:1}]}
    else if(path.endsWith('/logs'))json={items:[{runId:'input-run',startedAt:'2026-09-09T00:00:00Z',backend:'test-boundary'}]}
    else if(path.endsWith('/input')){posts.push({key:route.request().headers()['idempotency-key']!,path,body:route.request().postDataJSON()});if(posts.length===1)return route.abort('connectionreset');known=true;json={task:task()}}
    else if(path.endsWith('/tasks'))json={items:known?[task()]:[]}
    else if(path.endsWith('/tasks/input-task'))json={task:task()}
    return route.fulfill({json})
  })
  await page.routeWebSocket(/\/logs\/input-run\/stream/,ws=>{ws.onMessage(message=>sent.push(decodeEnvelope(new Uint8Array(message as Buffer)).type));ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:'input-run',type:MessageType.OpenAck,payload:encodeJSON({runId:'input-run',status:{phase:'capturing',earliest:1,latest:0,bytes:0},inputAvailable:true})})))})
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'输入实例',exact:true}).click();await page.getByRole('button',{name:'控制台',exact:true}).click();await page.getByRole('textbox',{name:'尚未发送的实例命令'}).fill('原始中文输入');await page.getByRole('button',{name:'发送命令',exact:true}).click();await expect(page.locator('.console-input-result')).toContainText('输入接收结果待核对');expect(posts).toHaveLength(1)
  await page.getByRole('textbox',{name:'尚未发送的实例命令'}).fill('下一条尚未发送');await page.reload();await expect(page.getByRole('textbox',{name:'尚未发送的实例命令'})).toHaveValue('下一条尚未发送');await expect(page.getByRole('button',{name:'发送命令',exact:true})).toBeDisabled();expect(posts).toHaveLength(1)
  await page.getByRole('button',{name:'用原请求核对发送',exact:true}).click();await expect(page.locator('.console-input-result')).toContainText('输入任务 RUNNING');expect(posts[1]).toEqual(posts[0]);expect(posts[0]!.body).toEqual({data:'原始中文输入\n'});expect(posts[0]!.path).toBe('/api/v1/instances/input-instance/logs/input-run/input')
  completed=true;await page.getByRole('button',{name:'核对发送结果',exact:true}).click();await expect(page.locator('.console-input-result')).toContainText('不代表程序处理成功');await expect(page.getByRole('textbox',{name:'尚未发送的实例命令'})).toHaveValue('下一条尚未发送');await expect(page.getByRole('button',{name:'发送命令',exact:true})).toBeEnabled();await page.reload();expect(posts).toHaveLength(2);expect(sent).not.toContain(MessageType.Data)
})
