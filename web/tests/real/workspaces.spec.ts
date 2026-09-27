import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test, expect, type Page} from '@playwright/test'
import {realLogin} from './login'

const fixture = JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!, 'utf8')) as {admin:{name:string;password:string}}
async function headers(page:Page) {
  return {'X-CSRF-Token': (await (await page.request.get('/api/v1/session')).json()).csrfToken, 'Idempotency-Key': randomUUID()}
}

test('real cloud workspace keeps layout-only sync private and restores explicit content', async ({page}) => {
  page.on('dialog', dialog => dialog.accept())
  await realLogin(page, fixture.admin)
  await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox', {name:'文件正文编辑器'}).focus()
  await page.keyboard.insertText('真实云端工作区草稿')
  const instance = (await (await page.request.get('/api/v1/instances')).json()).items[0] as {name:string}
  await page.locator('[data-app="blora.instances"]').click()
  const instanceWindow = page.locator('.app-window.focused')
  await instanceWindow.getByRole('button', {name:instance.name, exact:true}).click()
  await instanceWindow.getByRole('button', {name:'控制台', exact:true}).click()
  await instanceWindow.getByRole('button', {name:'打开终端会话管理', exact:true}).click()
  const terminalWindow = page.locator('.app-window.focused')
  await terminalWindow.getByRole('button', {name:'新建会话', exact:true}).click()
  await expect(terminalWindow.getByText('已连接 · 输入控制者', {exact:true})).toBeVisible({timeout:30000})
  const terminalSessionId = await terminalWindow.locator('.terminal-container').getAttribute('data-session-id')
  expect(terminalSessionId).toBeTruthy()
  await page.getByRole('button', {name:'切换工作区'}).click()
  await page.getByRole('button', {name:'同步布局/引用', exact:true}).click()
  await expect.poll(async() => (await (await page.request.get('/api/v1/workspaces')).json()).items[0], {timeout:15000}).toMatchObject({includeContent:false})
  const first = (await (await page.request.get('/api/v1/workspaces')).json()).items[0]
  expect(first.includeContent).toBe(false)
  const stripped = await (await page.request.get(`/api/v1/workspaces/${encodeURIComponent(first.workspaceId)}`)).json()
  expect(stripped.state.drafts).toEqual({})
  expect(stripped.state.terminals).toEqual({})
  await page.getByRole('button', {name:'同步工作内容', exact:true}).click()
  await expect.poll(async() => (await (await page.request.get('/api/v1/workspaces')).json()).items[0], {timeout:15000}).toMatchObject({revision:2,includeContent:true})
  const second = (await (await page.request.get('/api/v1/workspaces')).json()).items[0]
  expect(second.workspaceId).toBe(first.workspaceId)
  expect(second.revision).toBe(2)
  expect(second.includeContent).toBe(true)
  const withContent = await (await page.request.get(`/api/v1/workspaces/${encodeURIComponent(second.workspaceId)}`)).json()
  expect(Object.keys(withContent.state.drafts)).not.toHaveLength(0)
  expect(Object.keys(withContent.state.terminals)).not.toHaveLength(0)
  try {
    await expect(page.getByRole('button', {name:/个人桌面 · v2/})).toBeVisible()
    await page.getByRole('button', {name:/个人桌面 · v2/}).click()
    await expect(page.locator('.monaco-editor .view-lines')).toContainText('真实云端工作区草稿')
  } finally {
    if (terminalSessionId) {
      const close = await page.request.post(`/api/v1/terminals/${encodeURIComponent(terminalSessionId)}/close`, {headers:await headers(page), data:{}})
      expect([202, 404]).toContain(close.status())
    }
    const removed = await page.request.delete(`/api/v1/workspaces/${encodeURIComponent(second.workspaceId)}`, {headers:await headers(page)})
    expect([204, 404]).toContain(removed.status())
  }
})

test('real cloud workspace rejects a stale device revision without replacing local work', async ({page, browser}) => {
  page.on('dialog', dialog => dialog.accept())
  await realLogin(page, fixture.admin)
  await page.locator('[data-app="blora.editor"]').click()
  await page.getByRole('textbox', {name:'文件正文编辑器'}).focus()
  await page.keyboard.insertText('设备一的未同步现场')
  await page.getByRole('button', {name:'切换工作区'}).click()
  await page.getByRole('button', {name:'同步布局/引用', exact:true}).click()
  await expect.poll(async() => (await (await page.request.get('/api/v1/workspaces')).json()).items[0], {timeout:15000}).toMatchObject({revision:1,includeContent:false})
  const metadata = (await (await page.request.get('/api/v1/workspaces')).json()).items[0] as {workspaceId:string;revision:number}
  const remote = await (await page.request.get(`/api/v1/workspaces/${encodeURIComponent(metadata.workspaceId)}`)).json() as {state:Record<string,unknown>}

  // Reuse the authenticated session but give the second browser context a
  // distinct device identity. It writes a newer server revision first.
  const deviceTwo = await browser.newContext({storageState: await page.context().storageState()})
  await deviceTwo.addInitScript(() => localStorage.setItem('blora:device', 'device-two-real-conflict'))
  const devicePage = await deviceTwo.newPage()
  try {
    await devicePage.goto('/')
    await expect(devicePage.locator('.desktop')).toBeVisible()
    const session = await (await devicePage.request.get('/api/v1/session')).json() as {csrfToken:string}
    const changedState = structuredClone(remote.state)
    const preferences = (changedState.preferences || {}) as Record<string, unknown>
    preferences.remoteDeviceMarker = '设备二已提交'
    changedState.preferences = preferences
    const remoteUpdate = await devicePage.request.put(`/api/v1/workspaces/${encodeURIComponent(metadata.workspaceId)}`, {
      headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},
      data:{title:'冲突验证',deviceId:'device-two-real-conflict',baseRevision:1,schemaVersion:1,includeContent:false,state:changedState},
    })
    expect(remoteUpdate.status()).toBe(200)
  } finally {
    await deviceTwo.close()
  }

  await page.getByRole('button', {name:'同步布局/引用', exact:true}).click()
  await expect(page.getByText('云端副本已被其他设备更新；当前本地现场保留，请刷新列表后选择要打开的副本。')).toBeVisible()
  await expect(page.locator('.monaco-editor .view-lines')).toContainText('设备一的未同步现场')
  const after = (await (await page.request.get(`/api/v1/workspaces/${encodeURIComponent(metadata.workspaceId)}`)).json()) as {metadata:{revision:number};state:Record<string,unknown>}
  expect(after.metadata.revision).toBe(2)
  expect((after.state.preferences as Record<string,unknown>).remoteDeviceMarker).toBe('设备二已提交')
  try {
    const removed = await page.request.delete(`/api/v1/workspaces/${encodeURIComponent(metadata.workspaceId)}`, {headers:await headers(page)})
    expect([204, 404]).toContain(removed.status())
  } catch { /* Fixture cleanup is best effort after the conflict assertion. */ }
})
