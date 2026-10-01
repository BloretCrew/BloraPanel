import {pressEditorKey} from '../helpers/editor-key'
import {test,expect} from '@playwright/test'

test('editor split panes share text while keeping both cursors and the split visible after refresh',async({page})=>{
  const errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    const json=path.endsWith('/session')?{user:{userId:'split-user',name:'分栏测试',admin:true},csrfToken:'test-only'}:{items:[]}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.editor"]').click()
  const panes=page.locator('.editor-panes .monaco-container'),primary=page.getByRole('textbox',{name:'文件正文编辑器',exact:true}),secondary=page.getByRole('textbox',{name:'文件正文编辑器分栏',exact:true})
  await primary.focus();await page.keyboard.insertText('alpha\nbeta');await page.getByRole('button',{name:'分栏',exact:true}).click();await expect(panes).toHaveCount(2)
  await secondary.focus();await pressEditorKey(page,'End');await page.keyboard.insertText('\ngamma')
  await expect(panes.nth(0).locator('.view-lines')).toContainText('gamma');await expect(panes.nth(1).locator('.view-lines')).toContainText('gamma')
  await primary.focus();await pressEditorKey(page,'Home');await page.getByRole('button',{name:'关闭分栏',exact:true}).click();await expect(panes).toHaveCount(1)
  await page.getByRole('button',{name:'分栏',exact:true}).click();await expect(panes).toHaveCount(2);await secondary.focus();await page.keyboard.insertText('!')
  await expect(panes.nth(0).locator('.view-lines')).toContainText('gamma!');await expect(panes.nth(1).locator('.view-lines')).toContainText('gamma!')
  await page.reload();await expect(page.getByRole('button',{name:'关闭分栏',exact:true})).toBeVisible();await expect(panes).toHaveCount(2)
  await expect(panes.nth(0).locator('.view-lines')).toContainText('gamma!');await expect(panes.nth(1).locator('.view-lines')).toContainText('gamma!');expect(errors).toEqual([])
})
