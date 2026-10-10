import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
test('save-as copies undo history and explicit server reload remains reversible after refresh',async({page})=>{
  const writes:{path:string;text:string;version:string}[]=[],errors:string[]=[];let text='服务器原文',version='v1'
  page.on('pageerror',error=>errors.push(error.message))
  const task=()=>({taskId:'copy-task',requestId:'copy-request',resource:{kind:'instance',id:'copy-instance'},action:'file.write',state:'SUCCEEDED',phase:'committed',result:{version}})
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname;let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'copy-user',name:'另存测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'copy-instance',nodeId:'copy-node',name:'另存目标',state:'STOPPED',config:{},revision:1}]}
    else if(path.endsWith('/files/stat'))return route.fulfill({status:404,json:{error:{code:'NOT_FOUND'}}})
    else if(path.endsWith('/files/content')){if(route.request().method()==='PUT'){const body=route.request().postDataJSON();writes.push(body);text=body.text;version='v2';json={task:task()}}else json={text,version}}
    else if(path.endsWith('/tasks/copy-task'))json={task:task()}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click();await page.getByRole('textbox',{name:'文件正文编辑器'}).focus();await page.keyboard.insertText('原始草稿');await page.keyboard.insertText('第二步')
  await page.getByRole('button',{name:'另存为…',exact:true}).click();await selectStyledOption(page,page.getByRole('combobox',{name:'另存目标实例'}),'copy-instance');await page.getByRole('textbox',{name:'另存目标路径'}).fill('copied.txt');await page.reload();await expect(page.getByRole('textbox',{name:'另存目标路径'})).toHaveValue('copied.txt');expect(writes).toEqual([])
  await page.getByRole('button',{name:'确认另存正文',exact:true}).click();await expect(page.locator('.app-window')).toHaveCount(2);const copyWindow=page.locator('.app-window.focused');await expect(copyWindow.getByRole('button',{name:'保存到服务器',exact:true})).toBeEnabled();expect(writes).toEqual([{path:'copied.txt',text:'原始草稿第二步',version:'missing'}])
  await copyWindow.getByRole('button',{name:'↶ 撤销',exact:true}).click();await expect(copyWindow.locator('.view-lines')).toContainText('原始草稿');await expect(copyWindow.locator('.view-lines')).not.toContainText('第二步');await expect(page.locator('.app-window:not(.focused) .view-lines')).toContainText('原始草稿第二步')
  text='服务器后续版本';version='v3';await copyWindow.getByRole('button',{name:'重新读取…',exact:true}).click()
  // The confirmation is populated by an asynchronous server read. Reload
  // immediately after it exists, so this checks form recovery rather than
  // relying on an unfinished HTTP response winning the navigation race.
  await expect(page.getByRole('dialog',{name:'重新读取服务器正文'})).toBeVisible()
  await page.reload();await expect(page.getByRole('dialog',{name:'重新读取服务器正文'})).toBeVisible();expect(writes).toHaveLength(1);await page.getByRole('button',{name:'采用服务器正文（可撤销）',exact:true}).click();await expect(page.locator('.app-window.focused .view-lines')).toContainText('服务器后续版本');await page.reload();await page.locator('.app-window.focused').getByRole('button',{name:'↶ 撤销',exact:true}).click();await expect(page.locator('.app-window.focused .view-lines')).toContainText('原始草稿');await expect(page.locator('.app-window.focused .view-lines')).not.toContainText('服务器后续版本');expect(writes).toHaveLength(1);await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click();await expect(page.locator('.app-window')).toHaveCount(1);expect(errors).toEqual([])
})
