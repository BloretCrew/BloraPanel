import {pressEditorKey} from '../helpers/editor-key'
import {test,expect} from '@playwright/test'

test.use({userAgent:'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36'})
for(const newline of ['LF','CRLF'] as const){
const eol=newline==='CRLF'?'\r\n':'\n'
test(`Windows browser preserves ${newline} and captured save without overwriting newer input after reload`,async({page})=>{
  const writes:{path:string;text:string;version:string}[]=[],errors:string[]=[]
  let saved=false,reads=0
  const requests=new Map<string,string>()
  page.on('pageerror',error=>errors.push(error.message))
  const task=(id='save-1')=>({taskId:id,requestId:requests.get(id),resource:{kind:'instance',id:'save-instance'},action:'file.write',state:saved?'SUCCEEDED':'RUNNING',phase:saved?'committed':'writing',result:saved?{stage:'committed',version:'saved-version-2'}:undefined})
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'save-user',name:'保存测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'save-instance',nodeId:'node-save',nodeName:'文件测试节点',name:'文件测试实例',state:'STOPPED',config:{},revision:1}]}
    else if(path.endsWith('/files'))json={items:[{name:'config.txt',path:'config.txt',isDir:false,size:18,version:'original-version'}],total:1,version:'directory-version'}
    else if(path.endsWith('/files/access'))json={instanceId:'save-instance',nodeId:'node-save',nodeName:'文件测试节点',nodeState:'ONLINE',canRead:true,canWrite:true}
    else if(path.endsWith('/files/content')){
      if(route.request().method()==='PUT'){
        writes.push(route.request().postDataJSON())
        const id=`save-${writes.length}`,requestId=route.request().headers()['idempotency-key']!
        expect(requestId).toBeTruthy();requests.set(id,requestId)
        json={task:task(id)}
      }
      else{reads++;json={text:`\ufeff服务器初始${eol}第二行`,version:'original-version',encoding:'UTF-8 BOM',newline,maxBytes:4194304}}
    }else if(path.includes('/tasks/save-'))json={task:task(path.split('/').at(-1))}
    // A reload can precede receipt persistence. Production reconciles the
    // accepted write by its original request key without submitting it again.
    else if(path==='/api/v1/tasks')json={items:[...requests.keys()].map(id=>task(id))}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'文件测试实例',exact:true}).click();await page.getByRole('button',{name:'文件',exact:true}).click();await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click();await page.getByRole('button',{name:'config.txt',exact:true}).click()
  await expect(page.locator('.editor-toolbar')).toContainText('文件测试节点 · config.txt · 可读写')
  await expect(page.locator('.editor-status')).toContainText(`UTF-8 BOM · ${newline} · 上限 4 MiB`)
  const editor=page.getByRole('textbox',{name:'文件正文编辑器'});await editor.focus();await pressEditorKey(page,'End');await page.keyboard.insertText('，本次提交');await page.getByRole('button',{name:'保存到服务器',exact:true}).click()
  expect(writes[0]!.text.startsWith('\ufeff')).toBe(true);expect(writes[0]!.text).toContain(eol)
  if(newline==='LF')expect(writes[0]!.text).not.toContain('\r')
  await editor.focus();await pressEditorKey(page,'End');await page.keyboard.insertText('，并行新输入');await page.reload();await expect(page.locator('.editor-status')).toContainText(`UTF-8 BOM · ${newline} · 上限 4 MiB`)
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('并行新输入');expect(writes).toHaveLength(1);expect(writes[0]!.text).not.toContain('并行新输入');expect(reads).toBe(1)
  saved=true;await expect(page.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled();await expect(page.locator('.editor-status')).toContainText('服务器尚未保存');await expect(page.locator('.monaco-editor .view-lines')).toContainText('并行新输入')
  await page.getByRole('button',{name:'保存到服务器',exact:true}).click();await expect.poll(()=>writes.length).toBe(2);expect(writes[1]!.version).toBe('saved-version-2');expect(writes[1]!.text).toContain('并行新输入');expect(errors).toEqual([])
})
}
