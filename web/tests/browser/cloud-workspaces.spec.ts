import {test, expect} from '@playwright/test'

test('cloud workspace sync requires explicit content opt-in and restores a versioned copy', async ({page}) => {
  const puts: any[] = []
  const requests: string[] = []
  const putKeys: string[] = []
  let revision = 0
  let remoteState: any
  let remoteId = ''
  let dropNextPut = true
  await page.on('dialog', dialog => dialog.accept())
  await page.route('**/api/v1/**', async route => {
    const request = route.request()
    const path = new URL(request.url()).pathname
    requests.push(`${request.method()} ${path}`)
    if (path === '/api/v1/session') {
      return route.fulfill({json: {user: {userId: 'cloud-user', name: '云端用户', admin: true}, csrfToken: 'test-only'}})
    }
    if (path === '/api/v1/workspaces' && request.method() === 'GET') {
      return route.fulfill({json: {items: revision ? [{workspaceId: remoteId || 'cloud-copy', title: '云端测试', deviceId: 'device-remote', revision, schemaVersion: 1, includeContent: puts.at(-1)?.includeContent ?? false, updatedAt: new Date().toISOString()}] : []}})
    }
    if (path.startsWith('/api/v1/workspaces/') && request.method() === 'PUT') {
      putKeys.push((await request.headerValue('Idempotency-Key')) || '')
      if (dropNextPut) {
        dropNextPut = false
        return route.abort()
      }
      const body = JSON.parse(request.postData() || '{}')
      puts.push(body)
      revision += 1
      remoteState = body.state
      remoteId = decodeURIComponent(path.slice('/api/v1/workspaces/'.length))
      return route.fulfill({json: {metadata: {workspaceId: remoteId, title: body.title || '云端测试', deviceId: body.deviceId, revision, schemaVersion: body.schemaVersion, includeContent: body.includeContent, updatedAt: new Date().toISOString()}}})
    }
    if (path.startsWith('/api/v1/workspaces/') && request.method() === 'GET') {
      return route.fulfill({json: {metadata: {workspaceId: remoteId, title: '云端测试', deviceId: 'device-remote', revision, schemaVersion: 1, includeContent: true, updatedAt: new Date().toISOString()}, state: remoteState || {schemaVersion: 1, revision: 1, userId: 'cloud-user', deviceId: 'device-remote', browserTabId: 'remote-tab', workspaceId: remoteId, windows: {}, views: {}, order: [], shortcuts: {}, drafts: {}, terminals: {}, uploads: {}, preferences: {}, closedViews: []}}})
    }
    return route.fulfill({json: {items: []}})
  })

  await page.goto('/')
  await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox', {name: '文件正文编辑器'}).focus()
  await page.keyboard.insertText('仅本地草稿')
  await page.getByRole('button', {name: '切换工作区'}).click()
  await page.getByRole('button', {name: '同步布局/引用', exact: true}).click()
  await expect(page.locator('.workspace-menu .error')).toBeVisible()
  await page.reload()
  await page.getByRole('button', {name: '切换工作区'}).click()
  await page.getByRole('button', {name: '同步布局/引用', exact: true}).click()
  await expect.poll(() => puts).toHaveLength(1)
  expect(putKeys).toHaveLength(2)
  expect(putKeys[0]).toBeTruthy()
  expect(putKeys[0]).toBe(putKeys[1])
  expect(puts[0].includeContent).toBe(false)
  expect(puts[0].state.drafts).toEqual({})
  await page.getByRole('button', {name: '同步工作内容', exact: true}).click()
  await expect.poll(() => puts).toHaveLength(2)
  expect(puts[1].includeContent).toBe(true)
  expect(Object.keys(puts[1].state.drafts)).not.toHaveLength(0)
  await expect(page.getByRole('button', {name: /云端测试 · v2/ })).toBeVisible()
  await page.getByRole('button', {name: /云端测试 · v2/ }).click()
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('仅本地草稿')
  expect(requests.some(request => request.startsWith('PUT /api/v1/workspaces/'))).toBe(true)
  expect(requests.some(request => request.startsWith('GET /api/v1/workspaces/'))).toBe(true)
})
