import {test,expect} from '@playwright/test'

test('editor reflects server file.read without file.write and prevents local edits and saves',async({page})=>{
  const writes:unknown[]=[],errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname,instance={instanceId:'reader-instance',nodeId:'reader-node',nodeName:'只读节点',name:'只读实例',state:'STOPPED',config:{},revision:1}
    let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'reader-user',name:'只读成员',admin:false},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[instance]}
    else if(path.endsWith('/instances/reader-instance'))json={instance}
    else if(path.endsWith('/files'))json={items:[{name:'secret.txt',path:'secret.txt',isDir:false,size:12,version:'read-version'}],total:1,version:'directory-version'}
    else if(path.endsWith('/files/access'))json={instanceId:'reader-instance',nodeId:'reader-node',nodeName:'只读节点',nodeState:'ONLINE',canRead:true,canWrite:false}
    else if(path.endsWith('/files/content')){
      if(route.request().method()==='PUT'){writes.push(route.request().postDataJSON());json={task:{taskId:'must-not-run',requestId:'must-not-run',state:'SUCCEEDED'}}}
      else json={text:'原始只读正文',version:'read-version',encoding:'UTF-8',newline:'LF',maxBytes:4194304}
    }
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'只读实例',exact:true}).click();await page.getByRole('button',{name:'文件',exact:true}).click();await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click();await page.getByRole('button',{name:'secret.txt',exact:true}).click()
  await expect(page.locator('.editor-toolbar')).toContainText('只读节点 · secret.txt · 只读')
  const readOnlyButtons=page.locator('.editor-toolbar').getByRole('button',{name:'只读权限',exact:true});await expect(readOnlyButtons).toHaveCount(2);await expect(readOnlyButtons.nth(0)).toBeDisabled();await expect(readOnlyButtons.nth(1)).toBeDisabled()
  const editor=page.getByRole('textbox',{name:'文件正文编辑器',exact:true});await editor.focus();await page.keyboard.insertText('不应写入')
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('原始只读正文');await expect(page.locator('.monaco-editor .view-lines')).not.toContainText('不应写入');expect(writes).toEqual([]);expect(errors).toEqual([])
})
