// Real backend only: no routing, mocks, traces, screenshots or credential logs.
import {readFileSync} from 'node:fs'
import {chromium,expect} from '@playwright/test'

let browser
let phase='browser-start'
const step=value=>{phase=value;console.log('RUN browser phase: '+phase)}
async function closeBrowser(){try{await browser?.close()}catch{process.exitCode=1;console.error('FAIL browser phase: browser-close')}}
const deadline=setTimeout(async()=>{
  process.exitCode=1
  await closeBrowser()
},90000)
deadline.unref()
try {
  step('read-private-fixture')
  const fixture=JSON.parse(readFileSync(process.argv[2],'utf8'))
  step('browser-launch')
  browser=await chromium.launch({headless:true})
  step('browser-context')
  const context=await browser.newContext({baseURL:fixture.url,ignoreHTTPSErrors:true})
  const page=await context.newPage()
  const errors=[]
  page.on('pageerror',()=>errors.push('page-error'))
  step('login')
  await page.goto('/')
  await page.getByRole('textbox',{name:'账号',exact:true}).fill(fixture.name)
  await page.getByLabel('密码',{exact:true}).fill(fixture.password)
  const logged=page.waitForResponse(r=>r.url().endsWith('/api/v1/login')&&r.request().method()==='POST')
  await page.getByRole('button',{name:'进入工作区 →',exact:true}).click()
  if((await logged).status()!==200)throw new Error('login')
  await expect(page.locator('.desktop')).toBeVisible({timeout:30000})
  step('open-resource')
  const data=await(await page.request.get('/api/v1/instances')).json()
  const selected=data.items.find(i=>i.instanceId===fixture.instance)
  if(!selected)throw new Error('instance identity')
  await page.locator('[data-app="blora.instances"]').click()
  await page.getByRole('button',{name:selected.name,exact:true}).click()
  await page.getByRole('button',{name:'文件',exact:true}).click()
  await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click()
  step('create-file')
  await page.getByRole('button',{name:'新建文本',exact:true}).click()
  await page.getByRole('textbox',{name:'文件操作目标'}).fill('browser-device.txt')
  await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  await page.getByRole('textbox',{name:'文件正文编辑器'}).focus()
  await page.keyboard.insertText('真实浏览器未保存草稿\n')
  step('draft-refresh')
  await page.reload()
  await expect(page.locator('.view-lines')).toContainText('真实浏览器未保存草稿',{timeout:30000})
  step('save-readback')
  await page.getByRole('button',{name:'保存到服务器',exact:true}).click()
  await expect.poll(async()=>{
    const r=await page.request.get(`/api/v1/instances/${fixture.instance}/files/content?path=browser-device.txt`)
    return r.ok()?(await r.json()).text:''
  },{timeout:30000}).toBe('真实浏览器未保存草稿\n')
  step('final-refresh')
  await page.reload()
  await expect(page.locator('.view-lines')).toContainText('真实浏览器未保存草稿',{timeout:30000})
  if(errors.length)throw new Error('page error')
  await context.close()
} catch {
  // Intentionally omit Playwright exception text: locators can include passwords.
  process.exitCode=1
  console.error('FAIL browser phase: '+phase)
} finally {
  clearTimeout(deadline)
  await closeBrowser()
}
