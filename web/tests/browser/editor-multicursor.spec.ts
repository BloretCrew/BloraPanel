import {test,expect} from '@playwright/test'

test('multi-cursor edits remain one reversible protected history group after refresh',async({page})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    const json=path.endsWith('/session')?{user:{userId:'multi-user',name:'多光标测试',admin:true},csrfToken:'test-only'}:{items:[]}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click()
  const editor=page.getByRole('textbox',{name:'文件正文编辑器',exact:true}),content=page.locator('.monaco-editor .view-lines')
  await editor.focus();await page.keyboard.type('one two');await page.getByRole('button',{name:'查找',exact:true}).click()
  await page.getByRole('textbox',{name:'查找内容',exact:true}).fill('o');await page.getByRole('button',{name:'选择全部匹配',exact:true}).click();await page.keyboard.insertText('!')
  await expect(content).toContainText('!ne tw!')
  await page.reload();await expect(content).toContainText('!ne tw!')
  await editor.focus();await page.keyboard.press('Control+Z');await expect(content).toContainText('one two');await expect(content).not.toContainText('!ne tw!')
  await page.keyboard.press('Control+Y');await expect(content).toContainText('!ne tw!');expect(errors).toEqual([])
})
