import {realLogin} from './login'
import {readFileSync} from 'node:fs'
import {createHash,randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {decodeEnvelope,MessageType} from '../../src/services/protocol'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8')) as {admin:{name:string;password:string};member:{name:string;password:string};instanceIds:string[]}
async function login(page:Page,role:'admin'|'member'='admin'){await realLogin(page,fixture[role])}
async function resource(page:Page,section:'文件'|'控制台'){
  const {items}=await(await page.request.get('/api/v1/instances')).json(),selected=items.find((item:{instanceId:string})=>item.instanceId===fixture.instanceIds[0]);await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:selected.name,exact:true}).click();await page.getByRole('button',{name:section,exact:true}).click();await page.getByRole('button',{name:section==='文件'?'在文件管理器打开':'打开终端会话管理',exact:true}).click()
}
async function headers(page:Page){return{'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
async function awaitTask(page:Page,taskId:string){await expect.poll(async()=>{const response=await page.request.get(`/api/v1/tasks/${taskId}`);return(await response.json()).task.state},{timeout:30000}).toBe('SUCCEEDED')}
const filesBase=()=>`/api/v1/instances/${fixture.instanceIds[0]}/files`
async function content(page:Page,path:string){const response=await page.request.get(`${filesBase()}/content?path=${encodeURIComponent(path)}`);return response.ok()?await response.json():undefined}

test('actual shared file windows retain local edits when another account changes server version',async({page,browser})=>{
  await login(page);await resource(page,'文件')
  const name=`shared-conflict-${randomUUID()}.txt`
  await page.getByRole('button',{name:'新建文本',exact:true}).click();await page.getByRole('textbox',{name:'文件操作目标'}).fill(name);await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('共同基线\n第二行')
  await page.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('共同基线\n第二行')
  await expect(page.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled()
  const firstId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!
  await page.locator('.app-window.focused [aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'复制视图到新窗口',exact:true}).click()
  const secondId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!
  const first=page.locator(`[data-window-id="${firstId}"]`),second=page.locator(`[data-window-id="${secondId}"]`)
  await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+End');await page.keyboard.insertText('，共享未保存中文')
  await expect(first.locator('.view-lines')).toContainText('共享未保存中文')
  const context=await browser.newContext({baseURL:new URL(page.url()).origin,ignoreHTTPSErrors:true}),member=await context.newPage()
  try{
    await login(member,'member');const baseline=await content(member,name)
    const external=await member.request.put(`${filesBase()}/content`,{headers:await headers(member),data:{path:name,text:'另一账号服务器修订',version:baseline.version}})
    expect(external.status()).toBe(202);await awaitTask(member,(await external.json()).task.taskId)
    await second.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect(second.locator('.error')).toContainText(/版本|changed|conflict/i,{timeout:30000})
    expect((await content(page,name)).text).toBe('另一账号服务器修订')
    await second.getByRole('button',{name:'比较服务器版本',exact:true}).click();await expect(second.getByRole('textbox',{name:'服务器版本正文'})).toHaveValue('另一账号服务器修订')
    await page.reload();await expect(first.locator('.view-lines')).toContainText('共享未保存中文');await expect(second.locator('.view-lines')).toContainText('共享未保存中文')
    await expect(second.getByRole('textbox',{name:'服务器版本正文'})).toHaveValue('另一账号服务器修订')
    await second.getByRole('button',{name:'合并后采用此基线',exact:true}).click();await second.getByRole('button',{name:'保存到服务器',exact:true}).click()
    await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('共同基线\n第二行，共享未保存中文')
    await expect(first.locator('.view-lines')).toContainText('共享未保存中文')
    await expect(second.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled()
    let release!:()=>void,received!:()=>void
    const gate=new Promise<void>(resolve=>release=resolve),fetched=new Promise<void>(resolve=>received=resolve)
    await page.route('**/files/content?*',async route=>{const response=await route.fetch();received();await gate;await route.fulfill({response})})
    try{
      await second.getByRole('button',{name:'重新读取…',exact:true}).click();await fetched
      await second.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.press('Control+End');await page.keyboard.insertText('，等待响应期间的新输入')
      release();await expect(second.getByRole('dialog',{name:'重新读取服务器正文'})).toBeVisible()
      await expect(first.locator('.view-lines')).toContainText('等待响应期间的新输入');await expect(second.locator('.view-lines')).toContainText('等待响应期间的新输入')
      await page.reload();await expect(second.getByRole('dialog',{name:'重新读取服务器正文'})).toBeVisible();await expect(second.locator('.view-lines')).toContainText('等待响应期间的新输入')
    }finally{release();await page.unroute('**/files/content?*')}
  }finally{await context.close()}
})

test('actual file revocation preserves unsaved draft but denies save and snapshot reload access',async({page,browser})=>{
  await login(page)
  const user=(await(await page.request.get('/api/v1/users')).json()).items.find((user:any)=>user.name===fixture.member.name)
  const instance=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  const grants=['file.write','file.read','instance.read'].map(action=>({userId:user.userId,resource:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},action}))
  const context=await browser.newContext({baseURL:new URL(page.url()).origin,ignoreHTTPSErrors:true}),member=await context.newPage()
  try{
    await login(member,'member');await resource(member,'文件')
    const name=`revoke-editor-${randomUUID()}.txt`
    await member.getByRole('button',{name:'新建文本',exact:true}).click();await member.getByRole('textbox',{name:'文件操作目标'}).fill(name);await member.getByRole('button',{name:'确认文件操作',exact:true}).click()
    await member.getByRole('textbox',{name:'文件正文编辑器'}).focus();await member.keyboard.insertText('授权期间正文')
    await member.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('授权期间正文')
    await expect(member.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled()
    await member.getByRole('textbox',{name:'文件正文编辑器'}).focus();await member.keyboard.press('Control+End');await member.keyboard.insertText('，撤权前未保存中文')
    for(const grant of grants)expect((await page.request.delete('/api/v1/grants',{headers:await headers(page),data:grant})).status()).toBe(200)
    await member.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect(member.locator('.editor-app .error')).toContainText(/权限|授权|FORBIDDEN/,{timeout:15000})
    expect((await content(page,name)).text).toBe('授权期间正文')
    await member.reload();await expect(member.locator('.monaco-editor:visible .view-lines')).toContainText('撤权前未保存中文')
    expect((await member.request.get(`${filesBase()}/content?path=${encodeURIComponent(name)}`)).status()).toBe(403)
    expect((await(await member.request.get('/api/v1/instances')).json()).items.some((item:any)=>item.instanceId===instance.instanceId)).toBe(false)
    await member.getByRole('textbox',{name:'文件正文编辑器'}).focus();await member.keyboard.press('Control+z');await expect(member.locator('.monaco-editor:visible .view-lines')).not.toContainText('撤权前未保存中文')
    await member.keyboard.press('Control+y');await expect(member.locator('.monaco-editor:visible .view-lines')).toContainText('撤权前未保存中文')
    const download=member.waitForEvent('download');await member.getByRole('button',{name:'导出',exact:true}).click()
    expect(readFileSync((await(await download).path())!,'utf8')).toBe('授权期间正文，撤权前未保存中文')
  }finally{
    for(const grant of grants)await page.request.post('/api/v1/grants',{headers:await headers(page),data:grant})
    await context.close()
  }
})

test('actual active browser PTY loses input and restored snapshot cannot regain revoked read',async({page,browser})=>{
  await login(page)
  const user=(await(await page.request.get('/api/v1/users')).json()).items.find((user:any)=>user.name===fixture.member.name)
  const instance=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}`)).json()).instance
  const ref={kind:'instance',id:instance.instanceId,nodeId:instance.nodeId}
  const grants=[{userId:user.userId,resource:ref,action:'terminal.read'},{userId:user.userId,resource:ref,action:'terminal.input'},{userId:user.userId,resource:{kind:'node',id:instance.nodeId},action:'host.manage'}]
  for(const grant of grants)expect((await page.request.post('/api/v1/grants',{headers:await headers(page),data:grant})).status()).toBe(200)
  const context=await browser.newContext({baseURL:new URL(page.url()).origin,ignoreHTTPSErrors:true}),member=await context.newPage()
  let sessionId=''
  try{
    await login(member,'member');await resource(member,'控制台');await member.getByRole('button',{name:'新建会话',exact:true}).click()
    await expect(member.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
    sessionId=(await member.locator('.terminal-container').getAttribute('data-session-id'))!
    const name=`revoke-pty-${randomUUID()}.txt`
    await member.locator('.xterm-helper-textarea').focus();await member.keyboard.type(`printf x >> ${name}\n`)
    await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('x')
    expect((await page.request.delete('/api/v1/grants',{headers:await headers(page),data:grants[1]})).status()).toBe(200)
    await expect(member.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','false',{timeout:15000})
    await member.locator('.xterm-helper-textarea').focus();await member.keyboard.type(`printf y >> ${name}\n`)
    expect((await content(page,name)).text).toBe('x')
    expect((await page.request.delete('/api/v1/grants',{headers:await headers(page),data:grants[0]})).status()).toBe(200)
    expect((await member.request.get(`/api/v1/instances/${instance.instanceId}/terminals`)).status()).toBe(403)
    await member.reload()
    await expect(member.locator('.terminal-app .error')).toContainText(/权限|FORBIDDEN|授权/,{timeout:15000})
    await expect(member.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','false')
    expect((await content(page,name)).text).toBe('x')
  }finally{
    if(sessionId)await page.request.post(`/api/v1/terminals/${sessionId}/close`,{headers:await headers(page),data:{}})
    for(const grant of grants)await page.request.delete('/api/v1/grants',{headers:await headers(page),data:grant})
    await context.close()
  }
})

test('actual TLS file creation, conflict recovery and verified multi-chunk upload',async({page})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message));await login(page);await resource(page,'文件')
  const name=`browser-${randomUUID()}.txt`
  await page.getByRole('button',{name:'新建文本',exact:true}).click();await page.getByRole('textbox',{name:'文件操作目标'}).fill(name);await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  const editor=page.getByRole('textbox',{name:'文件正文编辑器'});await editor.focus();await page.keyboard.insertText('节点正文中文');await page.getByRole('button',{name:'保存到服务器',exact:true}).click()
  await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('节点正文中文');await expect(page.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled()
  const baseline=await content(page,name),external=await page.request.put(`${filesBase()}/content`,{headers:await headers(page),data:{path:name,text:'节点外部修订',version:baseline.version}});expect(external.status()).toBe(202);await awaitTask(page,(await external.json()).task.taskId)
  await editor.focus();await page.keyboard.press('Control+End');await page.keyboard.insertText('，本地未保存');await page.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect(page.locator('.editor-app .error')).toContainText(/版本|changed|conflict/i,{timeout:30000})
  expect((await content(page,name)).text).toBe('节点外部修订');await expect(page.locator('.monaco-editor .view-lines')).toContainText('本地未保存')
  await page.getByRole('button',{name:'比较服务器版本',exact:true}).click();await expect(page.getByRole('textbox',{name:'服务器版本正文'})).toHaveValue('节点外部修订');await page.reload();await expect(page.getByRole('textbox',{name:'服务器版本正文'})).toHaveValue('节点外部修订');await expect(page.locator('.monaco-editor .view-lines')).toContainText('本地未保存')
  await page.getByRole('button',{name:'合并后采用此基线',exact:true}).click();await page.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('节点正文中文，本地未保存')
  await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click()
  const uploadName=`browser-${randomUUID()}.bin`,bytes=Buffer.alloc(3*65536);for(let i=0;i<bytes.length;i++)bytes[i]=i%251
  await page.locator('input[type="file"]').setInputFiles({name:uploadName,mimeType:'application/octet-stream',buffer:bytes});await page.getByRole('button',{name:'确认文件操作',exact:true}).click();await expect(page.locator('.upload-row').filter({hasText:uploadName})).toContainText('节点确认已提交',{timeout:30000})
  const downloaded=await page.request.get(`${filesBase()}/download?path=${encodeURIComponent(uploadName)}`);expect(downloaded.ok()).toBeTruthy();expect(createHash('sha256').update(await downloaded.body()).digest('hex')).toBe(createHash('sha256').update(bytes).digest('hex'));expect(errors).toEqual([])
})

async function checkpoint(page:Page,sessionId:string){return page.evaluate(async sessionId=>{
  const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
  const values=await new Promise<Record<string,unknown>[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close()
  const saved=values.flatMap(value=>Object.values(value.terminals||{}) as {sessionId:string;sequence:number;screen:string;cols:number;rows:number;outputJournal?:{data:string}[]}[]).filter(value=>value.sessionId===sessionId).sort((a,b)=>b.sequence-a.sequence)[0]
  // The persisted representation is a full screen plus bounded parsed output
  // since that baseline. After reload, restoreComplete folds it into a screen.
  return saved?{...saved,screen:saved.screen+(saved.outputJournal||[]).map(event=>new TextDecoder().decode(Uint8Array.from(atob(event.data),c=>c.charCodeAt(0)))).join('')}:undefined
},sessionId)}

test('actual PTY pointer tear-out and merge preserve session and never replay input',async({page})=>{
  const inputs:string[]=[],errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  page.on('websocket',socket=>{if(socket.url().includes('/terminals/'))socket.on('framesent',event=>{if(typeof event.payload!=='string'){const frame=decodeEnvelope(new Uint8Array(event.payload));if(frame.type===MessageType.Data)inputs.push(new TextDecoder().decode(frame.payload))}})})
  await login(page);await resource(page,'控制台')
  const windowId=(await page.locator('.app-window.focused').getAttribute('data-window-id'))!,viewId=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  await page.locator('.app-window.focused').getByRole('button',{name:'新建标签',exact:true}).click()
  const blank=(await page.locator('.app-window.focused .view-tab.selected').getAttribute('data-view-tab'))!
  await page.locator(`[data-view-tab="${viewId}"]`).click();await page.getByRole('button',{name:'新建会话',exact:true}).click()
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  const sessionId=(await page.locator('.terminal-container:visible').getAttribute('data-session-id'))!,name=`drag-pty-${randomUUID()}.txt`
  await page.locator('.xterm-helper-textarea:visible').focus();await page.keyboard.type(`printf x >> ${name}; printf 'PTY-DRAG-LIVE\\n'\n`)
  await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe('x')
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain('PTY-DRAG-LIVE')
  const sent=inputs.join(''),source=(await page.locator(`[data-view-tab="${viewId}"]`).boundingBox())!
  await page.mouse.move(source.x+25,source.y+12);await page.mouse.down();await page.mouse.move(1100,760,{steps:12});await page.mouse.up();await page.reload()
  await expect(page.locator(`[data-view-tab="${viewId}"]`)).toHaveCount(1)
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  expect(await page.locator('.terminal-container:visible:not([data-session-id=""])').getAttribute('data-session-id')).toBe(sessionId)
  expect(inputs.join('')).toBe(sent)
  const detached=(await page.locator(`[data-view-tab="${viewId}"]`).boundingBox())!,target=(await page.locator(`[data-tab-strip="${windowId}"]`).boundingBox())!
  await page.mouse.move(detached.x+25,detached.y+12);await page.mouse.down();await page.mouse.move(target.x+10,target.y+12,{steps:12});await page.mouse.up();await page.reload()
  await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  expect(await page.locator(`[data-tab-strip="${windowId}"] [data-view-tab]`).evaluateAll(elements=>elements.map(element=>element.getAttribute('data-view-tab')))).toEqual([viewId,blank])
  expect(await page.locator('.terminal-container:visible:not([data-session-id=""])').getAttribute('data-session-id')).toBe(sessionId)
  expect(inputs.join('')).toBe(sent);expect((await content(page,name)).text).toBe('x')
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain('PTY-DRAG-LIVE')
  const sessions=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}/terminals`)).json()).items
  expect(sessions.filter((session:any)=>session.sessionId===sessionId&&session.state==='running')).toHaveLength(1)
  await page.getByRole('button',{name:'结束会话…',exact:true}).click();await page.getByRole('button',{name:'确认结束会话',exact:true}).click()
  await expect(page.getByText('会话已结束 · 屏幕检查点保留',{exact:true})).toBeVisible({timeout:30000});expect(errors).toEqual([])
})

test('actual vim unsaved buffer and top survive refresh without replayed input',async({page})=>{
  const inputs:string[]=[],errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  page.on('websocket',socket=>{if(socket.url().includes('/terminals/'))socket.on('framesent',event=>{if(typeof event.payload!=='string'){const frame=decodeEnvelope(new Uint8Array(event.payload));if(frame.type===MessageType.Data)inputs.push(new TextDecoder().decode(frame.payload))}})})
  await login(page);await resource(page,'控制台');await page.getByRole('button',{name:'新建会话',exact:true}).click()
  const connected=()=>expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  await connected()
  const sessionId=(await page.locator('.terminal-container').getAttribute('data-session-id'))!,name=`vim-${randomUUID()}.txt`,marker=`VIM-${randomUUID()}`
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type(`vim -Nu NONE -n ${name}\n`)
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen,{timeout:30000}).toContain(name)
  await page.keyboard.type('i');await page.keyboard.insertText(marker+'中文现场');await page.keyboard.press('Escape')
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain('中文现场')
  const before=inputs.join('');await page.reload();await connected();expect(inputs.join('')).toBe(before)
  expect(await page.locator('.terminal-container').getAttribute('data-session-id')).toBe(sessionId)
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain(marker)
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type(':wq\n')
  await expect.poll(async()=>(await content(page,name))?.text,{timeout:30000}).toBe(marker+'中文现场\n')
  await page.keyboard.type('top -d 1\n')
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen,{timeout:30000}).toContain('load average')
  const topInput=inputs.join('');await page.reload();await connected();expect(inputs.join('')).toBe(topInput)
  await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain('load average')
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type('q');await page.keyboard.type(`printf 'TOP-EXIT-VERIFIED\\n' > ${name}.top\n`)
  await expect.poll(async()=>(await content(page,name+'.top'))?.text,{timeout:30000}).toBe('TOP-EXIT-VERIFIED\n')
  await page.getByRole('button',{name:'结束会话…',exact:true}).click();await page.getByRole('button',{name:'确认结束会话',exact:true}).click()
  await expect(page.getByText('会话已结束 · 屏幕检查点保留',{exact:true})).toBeVisible({timeout:30000});expect(errors).toEqual([])
})

test('actual WSS Protobuf PTY input, alternate screen, refresh and moving retain one session',async({page})=>{
  const errors:string[]=[],inputs:string[]=[],attachments:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  page.on('websocket',socket=>{if(!socket.url().includes('/terminals/'))return;attachments.push(socket.url());socket.on('framesent',event=>{if(typeof event.payload==='string')return;const frame=decodeEnvelope(new Uint8Array(event.payload));if(frame.type===MessageType.Data)inputs.push(new TextDecoder().decode(frame.payload))})})
  await login(page);await resource(page,'控制台');await page.getByRole('button',{name:'新建会话',exact:true}).click();await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000})
  const sessionId=(await page.locator('.terminal-container').getAttribute('data-session-id'))!,countFile=`browser-pty-${randomUUID()}.txt`
  const command=`printf x >> ${countFile}; printf '\\033[?1049h\\033[2J\\033[H\\344\\270\\255\\346\\226\\207-PTY-BROWSER\\n'\n`
  await page.locator('.xterm-helper-textarea').focus();await page.keyboard.type(command)
  await expect.poll(async()=>(await content(page,countFile))?.text,{timeout:30000}).toBe('x');await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen,{timeout:30000}).toContain('中文-PTY-BROWSER')
  const sent=inputs.join(''),before=await checkpoint(page,sessionId);await page.reload();await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000});await expect.poll(async()=>(await checkpoint(page,sessionId))?.screen).toContain('中文-PTY-BROWSER')
  expect(new URL(attachments.at(-1)!).searchParams.get('sequence')).toBe(String(before!.sequence));expect(inputs.join('')).toBe(sent);expect((await content(page,countFile)).text).toBe('x')
  const viewId=await page.locator('.app-window.focused [data-view-tab]').getAttribute('data-view-tab');await page.locator('.app-window.focused [aria-label^="标签菜单"]').click();await page.getByRole('button',{name:'移到新窗口',exact:true}).click();await expect(page.locator(`[data-view-tab="${viewId}"]`)).toHaveCount(1);await expect(page.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible({timeout:30000});expect(await page.locator('.terminal-container').getAttribute('data-session-id')).toBe(sessionId);expect(inputs.join('')).toBe(sent);expect((await content(page,countFile)).text).toBe('x')
  const sessions=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}/terminals`)).json()).items;expect(sessions.filter((session:{sessionId:string})=>session.sessionId===sessionId)).toHaveLength(1)
  await page.screenshot({path:'real-test-results/real-tls-terminal.png'})
  await page.getByRole('button',{name:'同会话新窗口',exact:true}).click();const observer=page.locator('.app-window.focused');await expect(observer.getByText('已连接 · 只读观察',{exact:true})).toBeVisible({timeout:30000});await observer.locator('.xterm-helper-textarea').focus();await page.keyboard.type('blocked');expect(inputs.join('')).toBe(sent)
  await observer.getByRole('button',{name:'申请接管',exact:true}).click();await expect(observer.getByText('已连接 · 输入控制者',{exact:true})).toBeVisible();const original=page.locator('.app-window').filter({has:page.locator(`[data-view-tab="${viewId}"]`)});await expect(original.locator('.terminal-container')).toHaveAttribute('data-terminal-writable','false',{timeout:15000})
  const priorCols=(await checkpoint(page,sessionId))!.cols;await observer.getByRole('button',{name:'放大终端字号'}).click();await expect.poll(async()=>(await checkpoint(page,sessionId))?.cols).toBeLessThan(priorCols)
  await observer.getByRole('button',{name:'关闭窗口',exact:true}).click();const stillRunning=(await(await page.request.get(`/api/v1/instances/${fixture.instanceIds[0]}/terminals`)).json()).items.find((session:{sessionId:string})=>session.sessionId===sessionId);expect(stillRunning.state).toBe('running');expect(inputs.join('')).toBe(sent)
  await original.getByRole('button',{name:'结束会话…',exact:true}).click();await page.reload();await expect(page.getByRole('dialog',{name:'结束终端会话'})).toBeVisible();await page.getByRole('button',{name:'确认结束会话',exact:true}).click();await expect(page.getByText('会话已结束 · 屏幕检查点保留',{exact:true})).toBeVisible({timeout:30000});expect((await content(page,countFile)).text).toBe('x');expect(inputs.join('')).toBe(sent);expect(errors).toEqual([])
})
