import {test, expect} from '@playwright/test'

test.setTimeout(90_000)

for (const [platform, backend] of [['windows', 'netsh'], ['darwin', 'pf'], ['linux', 'nftables']]) {
  test(`system management keeps ${backend} status readable without offering unsupported mutations`, async ({page}) => {
    const posts: string[] = []
    await page.route('**/api/v1/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname
      if (request.method() === 'POST') posts.push(path)
      if (path.endsWith('/session')) return route.fulfill({json: {user: {userId: 'admin', name: '管理员', admin: true}, csrfToken: 'test-only'}})
      if (path === '/api/v1/nodes') return route.fulfill({json: {items: [{nodeId: 'read-only-node', name: '只读防火墙节点', platform}]}})
      if (path.endsWith('/system/capabilities')) return route.fulfill({json: {platform, services: 'unavailable', firewall: backend, scheduledTasks: 'unavailable'}})
      if (path.endsWith('/system/firewall')) return route.fulfill({json: {backend, state: 'enabled', detail: 'Read-only firewall status remains available'}})
      return route.fulfill({json: {items: []}})
    })
    await page.goto('/')
    await page.locator('.launcher-button').click()
    await page.locator('.launcher').getByRole('button', {name: '系统管理'}).click()
    await page.locator('select[aria-label="系统节点"]').selectOption('read-only-node')
    await expect(page.getByText(`${backend} · enabled`, {exact: true})).toBeVisible()
    await page.getByRole('textbox', {name: '目标防火墙规则'}).fill('443/tcp')
    await expect(page.getByRole('button', {name: '计算差异'})).toBeDisabled()
    await expect(page.getByRole('button', {name: '确认并应用'})).toBeDisabled()
    await page.reload()
    await expect(page.getByText(`${backend} · enabled`, {exact: true})).toBeVisible()
    await expect(page.getByRole('button', {name: '计算差异'})).toBeDisabled()
    expect(posts).toEqual([])
  })
}

test('system management submits authorized task and firewall confirmations', async ({page}) => {
  const requests: {path: string; body: string}[] = []
  let phase='queued',state='QUEUED'
  page.on('dialog', dialog => dialog.accept())
  await page.route('**/api/v1/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname
    if (request.method() === 'POST') requests.push({path, body: request.postData() || ''})
    if (path.endsWith('/session')) return route.fulfill({json: {user: {userId: 'admin', name: '管理员', admin: true}, csrfToken: 'test-only'}})
    if (path === '/api/v1/nodes') return route.fulfill({json: {items: [{nodeId: 'node-1', name: '测试节点', platform: 'linux'},{nodeId:'node-2',name:'第二节点',platform:'linux'}]}})
    if(path==='/api/v1/tasks/firewall-task-1') return route.fulfill({json:{task:{taskId:'firewall-task-1',state,phase}}})
    if (path.endsWith('/system/capabilities')) return route.fulfill({json: {platform: 'linux', services: 'systemd', firewall: 'firewalld', scheduledTasks: 'systemd'}})
    if (path.endsWith('/system/services')) return route.fulfill({json:new URL(request.url()).searchParams.get('after')==='a.service'?{items:[{name:'z.service',state:'running'}]}:{items:[{name:'a.service',state:'stopped'}],nextAfter:'a.service'}})
    if (path.endsWith('/system/tasks')) return route.fulfill({json:new URL(request.url()).searchParams.get('after')==='demo.timer'?{items:[{name:'last.timer',state:'ready'}]}:{items: [{name: 'demo.timer', state: 'active', schedule: '* * * * *'}],nextAfter:'demo.timer'}})
    if (path.endsWith('/system/firewall')) return route.fulfill({json: {backend: 'firewalld', state: 'running'}})
    if (path.endsWith('/system/firewall/preview')) return route.fulfill({json: {backend: 'firewalld',zone:'public',planHash:'a'.repeat(64), permanentAdd:['443/tcp'],permanentRemove:['53/udp'], current: ['22/tcp'], add: ['80/tcp'], remove: [], apply: true}})
    if (path.endsWith('/system/firewall/apply')) return route.fulfill({json: {task: {taskId: 'firewall-task-1', state: 'RUNNING', phase: 'awaiting_confirmation'}}})
    if (path.endsWith('/system/firewall/confirm')) return route.fulfill({json: {taskId: 'firewall-task-1', confirmed: true}})
    return route.fulfill({json: {items: []}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button', {name: '系统管理'}).click()
  await page.locator('select[aria-label="系统节点"]').selectOption('node-1')
  await expect(page.getByText('demo.timer')).toBeVisible()
  await page.getByRole('button',{name:'下一批计划任务'}).click()
  await expect(page.getByText('last.timer',{exact:true})).toBeVisible({timeout: 15000})
  await page.reload()
  await expect(page.getByText('last.timer',{exact:true})).toBeVisible({timeout: 15000})
  await expect(page.getByRole('button',{name:'下一批计划任务'})).toBeDisabled()
  await page.getByRole('button',{name:'计划任务首批'}).click()
  await expect(page.getByText('demo.timer',{exact:true})).toBeVisible()
  await page.getByRole('button',{name:'下一批服务'}).click()
  await expect(page.getByText('z.service · running',{exact:true})).toBeVisible({timeout: 15000})
  await page.reload()
  await expect(page.getByText('z.service · running',{exact:true})).toBeVisible({timeout: 15000})
  await expect(page.getByRole('button',{name:'下一批服务'})).toBeDisabled()
  await page.getByRole('button',{name:'服务首批'}).click()
  await expect(page.getByText('a.service · stopped',{exact:true})).toBeVisible()
  await page.getByRole('button', {name: '启用'}).click()
  await expect(page.getByRole('button',{name:'确认并应用'})).toBeDisabled()
  await page.getByRole('button',{name:'计算差异'}).click()
  await expect(page.getByText('public',{exact:true})).toBeVisible()
  await expect(page.getByText('53/udp',{exact:true})).toBeVisible()
  await page.getByRole('button', {name: '确认并应用'}).click()
  await expect(page.getByText('排队中',{exact:false})).toBeVisible()
  await expect(page.getByRole('button', {name: '确认保留'})).toBeDisabled()
  await page.locator('select[aria-label="系统节点"]').selectOption('node-2')
  phase='awaiting_confirmation';state='RUNNING'
  await page.reload()
  await expect(page.locator('select[aria-label="系统节点"]')).toHaveValue('node-2',{timeout: 15000})
  await page.getByRole('button', {name: '确认保留'}).click()
  await expect.poll(() => requests.map(x => x.path)).toEqual(expect.arrayContaining([
    '/api/v1/nodes/node-1/system/tasks/actions',
    '/api/v1/nodes/node-1/system/firewall/apply',
    '/api/v1/nodes/node-1/system/firewall/confirm',
  ]))
  const firewall = requests.find(x => x.path.endsWith('/system/firewall/apply'))
  expect(JSON.parse(firewall!.body).desired).toEqual([])
  expect(JSON.parse(firewall!.body).planHash).toBe('a'.repeat(64))
})

test('system management disables controls when node capabilities are unavailable', async ({page}) => {
  const posts: string[] = []
  page.on('dialog', dialog => dialog.accept())
  await page.route('**/api/v1/**', async route => {
    const request = route.request(), path = new URL(request.url()).pathname
    if (request.method() === 'POST') posts.push(path)
    if (path.endsWith('/session')) return route.fulfill({json: {user: {userId: 'admin', name: '管理员', admin: true}, csrfToken: 'test-only'}})
    if (path === '/api/v1/nodes') return route.fulfill({json: {items: [{nodeId: 'node-1', name: '受限节点', platform: 'windows'}]}})
    if (path.endsWith('/system/capabilities')) return route.fulfill({json: {platform: 'windows', services: 'unavailable', firewall: 'unsupported', scheduledTasks: 'unavailable'}})
    return route.fulfill({json: {items: []}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button', {name: '系统管理'}).click()
  await page.locator('select[aria-label="系统节点"]').selectOption('node-1')
  await expect(page.getByText('unavailable', {exact: true}).first()).toBeVisible()
  await expect(page.getByRole('button', {name: '计算差异'})).toBeDisabled()
  await expect(page.getByRole('button', {name: '启动'})).toHaveCount(0)
  await expect(page.getByRole('button', {name: '启用'})).toHaveCount(0)
  expect(posts).toEqual([])
})
