import {test,expect} from '../helpers/management-fixture'
import type {Page} from '@playwright/test'

async function openSettings(page:Page){
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'设置',exact:true}).click()
}

// Browser doubles validate client decisions; real download/instance behavior
// is exercised separately by scripts/core-update-smoke.py.
test('release updates pin the preview and retry an uncertain request with the same receipt key',async({page})=>{
  const revision='b'.repeat(40),writes:{revision:string;key:string}[]=[]
  const status:any={component:'master',version:'0.1.0-beta.1',revision:'a'.repeat(40),platform:'linux/amd64',checking:false,
    source:{repository:'https://github.com/BloretCrew/BloraPanel',channel:'beta',apiUrl:'https://api.github.com/repos/BloretCrew/BloraPanel'}}
  await page.context().route('**/api/v1/**',async route=>{
    const request=route.request(),path=new URL(request.url()).pathname
    if(path==='/api/v1/session')return route.fulfill({json:{user:{userId:'update-admin',name:'测试管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/system/updates/check'){
      status.preview={revision,currentRevision:status.revision,version:'0.1.0-beta.2',compatible:true,upToDate:false,downloadBytes:1048576}
      return route.fulfill({json:status})
    }
    if(path==='/api/v1/system/updates/apply'){
      writes.push({revision:request.postDataJSON().revision,key:request.headers()['idempotency-key']!})
      if(writes.length===1)return route.abort('failed')
      status.job={id:'receipt',revision,state:'running',phase:'downloading',downloadedBytes:524288,totalBytes:1048576}
      return route.fulfill({json:status.job})
    }
    return route.fulfill({json:path==='/api/v1/system/updates'?status:{items:[]}})
  })
  await openSettings(page)
  const panel=page.getByRole('region',{name:'系统更新'})
  await expect(panel.getByRole('textbox',{name:'发行 API 镜像'})).toHaveValue('')
  await panel.getByRole('button',{name:'检查更新',exact:true}).click()
  await expect(panel).toContainText('0.1.0-beta.2')
  page.once('dialog',dialog=>dialog.accept())
  await panel.getByRole('button',{name:'应用此版本'}).click()
  await expect(panel.getByRole('button',{name:'重试同一更新请求'})).toBeEnabled()
  await panel.getByRole('button',{name:'重试同一更新请求'}).click()
  expect(writes).toHaveLength(2)
  expect(writes[0]).toEqual(writes[1])
  expect(writes[0]!.revision).toBe(revision)
  expect(writes[0]!.key).toBeTruthy()
  await expect(panel.getByRole('progressbar',{name:'发行包下载进度'})).toHaveAttribute('value','524288')
  await expect(panel.getByRole('button',{name:'应用此版本'})).toBeDisabled()
})

test('ordinary users never get the core update entry or polling requests',async({page})=>{
  let updateRequests=0
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(path.includes('/updates'))updateRequests++
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'update-member',name:'测试成员',admin:false},csrfToken:'test-only'}:{items:[]}})
  })
  await openSettings(page)
  await expect(page.getByRole('heading',{name:'工作区设置'})).toBeVisible()
  await expect(page.getByRole('region',{name:'系统更新'})).toHaveCount(0)
  expect(updateRequests).toBe(0)
})
