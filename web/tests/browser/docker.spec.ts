import {test, expect} from '@playwright/test'
import {encodeEnvelope,encodeJSON,MessageType} from '../../src/services/protocol'

test('Docker app queries containers and submits an explicit idempotent action', async ({page}) => {
  const requests:{path:string;method:string;body:string}[]=[]
  await page.route('**/api/v1/**', async route => {
    const request=route.request(), path=new URL(request.url()).pathname
    if(request.method()!=='GET') requests.push({path,method:request.method(),body:request.postData()||''})
    if(path.endsWith('/session')) return route.fulfill({json:{user:{userId:'admin',name:'管理员',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/nodes/managed') return route.fulfill({json:{items:[{nodeId:'docker-node',name:'Docker 节点',platform:'linux',state:'ONLINE'}]}})
    if(path==='/api/v1/nodes/docker-node/docker/query') return route.fulfill({json:{kind:'containers',observedAt:'2026-09-09T00:00:00Z',items:[{name:'web',state:'running',health:'healthy',image:'fixture:latest',target:{kind:'container',id:'a'.repeat(64),createdAt:'2026-09-09T00:00:00Z',fingerprint:'fixture'}}]}})
    if(path==='/api/v1/nodes/docker-node/docker/actions') return route.fulfill({status:202,json:{task:{taskId:'docker-task',requestId:request.headers()['idempotency-key'],action:'container.start',resource:{kind:'node',id:'docker-node',nodeId:'docker-node'},state:'QUEUED',phase:'accepted',revision:1}}})
    if(path==='/api/v1/tasks/docker-task') return route.fulfill({json:{task:{taskId:'docker-task',requestId:'req',action:'container.start',resource:{kind:'node',id:'docker-node',nodeId:'docker-node'},state:'SUCCEEDED',phase:'completed',revision:2}}})
    return route.fulfill({json:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'容器中心'}).click()
  await expect(page.getByText('web',{exact:true})).toBeVisible()
  await page.getByRole('button',{name:/启动/}).click()
  await expect(page.getByRole('dialog',{name:'确认 Docker 操作'})).toContainText('container.start')
  await page.getByRole('dialog',{name:'确认 Docker 操作'}).getByRole('button',{name:'确认执行'}).click()
  await expect(page.getByRole('status')).toContainText('docker-task')
  expect(requests.some(x=>x.path==='/api/v1/nodes/docker-node/docker/actions'&&JSON.parse(x.body).action==='container.start')).toBe(true)
})

test('Docker app opens bounded archived log windows and returns to live logs', async ({page}) => {
  const containerId='a'.repeat(64)
  await page.route('**/api/v1/**', async route => {
    const request=route.request(), path=new URL(request.url()).pathname
    if(path.endsWith('/session')) return route.fulfill({json:{user:{userId:'docker-logs-user',name:'日志观察者',admin:true},csrfToken:'test-only'}})
    if(path==='/api/v1/nodes/managed') return route.fulfill({json:{items:[{nodeId:'docker-node',name:'Docker 节点',platform:'linux',state:'ONLINE'}]}})
    if(path==='/api/v1/nodes/docker-node/docker/query') return route.fulfill({json:{kind:'containers',observedAt:'2026-09-09T00:00:00Z',items:[{name:'web',state:'running',health:'healthy',image:'fixture:latest',target:{kind:'container',id:containerId,createdAt:'2026-09-09T00:00:00Z',fingerprint:'fixture'}}]}})
    if(path===`/api/v1/nodes/docker-node/docker/containers/${containerId}/logs/history`) return route.fulfill({json:{items:[{containerId,observedAt:'2026-09-09T00:00:00Z',frames:[{stream:'stdout',data:Buffer.from('归档日志内容\n').toString('base64')}],truncated:false,possibleGap:true}],retention:100,possibleGap:true}})
    return route.fulfill({json:{items:[]}})
  })
  await page.routeWebSocket(new RegExp(`/api/v1/nodes/docker-node/docker/containers/${containerId}/logs\\?`),ws=>{
    ws.send(Buffer.from(encodeEnvelope({protocolVersion:1,generation:1,channel:2,streamId:containerId,type:MessageType.OpenAck,payload:encodeJSON({containerId,mode:'timestamp-window',possibleGap:true})})))
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'容器中心'}).click()
  await expect(page.getByText('web',{exact:true})).toBeVisible()
  await page.getByRole('button',{name:'日志',exact:true}).click()
  await expect(page.getByRole('button',{name:'查看归档',exact:true})).toBeVisible()
  await page.getByRole('button',{name:'查看归档',exact:true}).click()
  await expect(page.locator('[data-log-mode="history"]')).toContainText('归档日志内容')
  await expect(page.locator('[data-log-mode="history"]')).toContainText('可能存在缺口')
  await page.getByRole('button',{name:'返回实时日志',exact:true}).click()
  await expect(page.locator('[data-log-mode="live"]')).toBeVisible()
})
