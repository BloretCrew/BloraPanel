import {test, expect} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'

test('monitoring searches, sorts and submits process identity', async ({page}) => {
  const requests:{path:string;body:string}[] = []
  await page.route('**/api/v1/**', async route => {
    const request=route.request(), path=new URL(request.url()).pathname
    if(request.method()==='POST') requests.push({path,body:request.postData()||''})
    if(path.endsWith('/session')) return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/nodes') return route.fulfill({json:{items:[{nodeId:'node-1',name:'节点一',platform:'linux',state:'ONLINE'},{nodeId:'node-2',name:'节点二',platform:'windows',state:'ONLINE'}]}})
    if(path.endsWith('/metrics')) return route.fulfill({json:{observedAt:new Date().toISOString(),cpuPercent:4.2,memoryUsedBytes:10,memoryTotalBytes:20,diskUsedBytes:1,diskTotalBytes:2,stale:false}})
    if(path.endsWith('/processes')) return route.fulfill({json:new URL(request.url()).searchParams.get('afterPid')==='22'?{items:[{pid:33,startTicks:303,command:'worker-later',rssBytes:7000}]}:{items:[{pid:11,startTicks:101,command:'small',rssBytes:1000},{pid:22,startTicks:202,command:'worker',rssBytes:5000}],nextAfterPid:22}})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'监控与进程'}).click()
  await selectStyledOption(page,page.getByRole('combobox',{name:'监控节点'}),'node-1')
  await expect(page.getByText('worker')).toBeVisible()
  await selectStyledOption(page,page.getByRole('combobox',{name:'进程排序'}),'rss')
  await page.getByRole('textbox',{name:'搜索进程'}).fill('worker')
  await page.getByRole('button',{name:'请求终止'}).click()
  await expect(page.getByRole('dialog',{name:'确认终止进程'})).toContainText('202')
  await selectStyledOption(page,page.getByRole('combobox',{name:'监控节点'}),'node-2')
  await expect(page.getByRole('dialog',{name:'确认终止进程'})).toContainText('node-1')
  await page.getByRole('button',{name:'确认终止'}).click()
  await expect.poll(()=>requests.map(x=>x.path)).toContain('/api/v1/nodes/node-1/processes/22/terminate')
  const body=requests.find(x=>x.path.endsWith('/processes/22/terminate'))!.body
  expect(JSON.parse(body).startTicks).toBe(202)
  expect(requests.some(x=>x.path==='/api/v1/nodes/node-2/processes/22/terminate')).toBe(false)
  await page.reload()
  await expect(page.getByRole('combobox',{name:'监控节点'})).toHaveAttribute('data-value','node-2')
  await expect(page.getByRole('textbox',{name:'搜索进程'})).toHaveValue('worker')
  await expect(page.getByRole('combobox',{name:'进程排序'})).toHaveAttribute('data-value','rss')
  await page.getByRole('button',{name:'下一批进程'}).click()
  await expect(page.getByText('worker-later',{exact:false})).toBeVisible()
  await expect(page.getByRole('button',{name:'下一批进程'})).toBeDisabled()
  await page.reload()
  await expect(page.getByText('worker-later',{exact:false})).toBeVisible()
  await page.getByRole('button',{name:'返回首批'}).click()
  await expect(page.getByText('worker-later',{exact:false})).toHaveCount(0)
})
