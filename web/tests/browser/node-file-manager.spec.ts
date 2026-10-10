import {test,expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'

test('file manager opens independently and switches whole-disk scope between daemon nodes',async({page})=>{
  const directories:string[]=[]
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'node-files-user',name:'节点文件测试',admin:true},csrfToken:'test-only'}
    else if(path==='/api/v1/nodes/files')json={items:[{nodeId:'node-a',name:'节点 A',state:'ONLINE'},{nodeId:'node-b',name:'节点 B',state:'ONLINE'}]}
    else if(path==='/api/v1/nodes/node-a/files'){directories.push('node-a:'+new URL(route.request().url()).searchParams.get('path'));json={items:[{name:'etc',path:'etc',isDir:true,size:0},{name:'node-a-root.txt',path:'node-a-root.txt',isDir:false,size:11}],total:2,version:'node-a-root',nextOffset:0}}
    else if(path==='/api/v1/nodes/node-b/files'){directories.push('node-b:'+new URL(route.request().url()).searchParams.get('path'));json={items:[{name:'var',path:'var',isDir:true,size:0},{name:'node-b-root.txt',path:'node-b-root.txt',isDir:false,size:11}],total:2,version:'node-b-root',nextOffset:0}}
    return route.fulfill({json})
  })
  await page.goto('/')
  await expect(page.locator('select,option')).toHaveCount(0)
  await page.locator('[data-app="blora.files"]').click()
  const node=page.getByRole('combobox',{name:'选择文件管理节点'})
  await expect(node).toBeVisible()
  await selectStyledOption(page,node,'node-a')
  await expect(page.getByRole('button',{name:'node-a-root.txt',exact:true})).toBeVisible()
  await expect(page.getByText('按 daemon 节点浏览整台主机的文件系统。')).toBeVisible()
  await selectStyledOption(page,node,'node-b')
  await expect(page.getByRole('button',{name:'node-b-root.txt',exact:true})).toBeVisible()
  await expect(page.locator('select,option')).toHaveCount(0)
  expect(directories).toEqual(expect.arrayContaining(['node-a:.','node-b:.']))
})

test('custom combobox supports keyboard operation and exposes a listbox instead of native options',async({page})=>{
  await page.route('**/api/v1/**',route=>{const path=new URL(route.request().url()).pathname;return route.fulfill({json:path.endsWith('/session')?{user:{userId:'select-user',name:'下拉框测试',admin:true},csrfToken:'test-only'}:{items:[]}})})
  await page.goto('/')
  await page.locator('.dock-item[aria-label="设置"]').click()
  const combo=page.getByRole('combobox',{name:'主题',exact:true})
  await expect(page.locator('select,option')).toHaveCount(0)
  await combo.focus();await page.keyboard.press('Space')
  const popup=page.getByRole('listbox')
  await expect(popup).toBeVisible();await expect(popup.getByRole('option')).toHaveCount(2)
  await page.keyboard.press('ArrowDown');await page.keyboard.press('Enter')
  await expect(page.locator('html')).toHaveAttribute('data-theme','dark')
  await expect(popup).toBeHidden()
})
