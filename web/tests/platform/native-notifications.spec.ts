import {readFileSync} from 'node:fs'
import {execFileSync} from 'node:child_process'
import {randomUUID} from 'node:crypto'
import {test,expect} from '@playwright/test'
import {realLogin} from '../real/login'

const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8'))
const run=(command:string,...args:string[])=>execFileSync(command,args,{encoding:'utf8',timeout:10000}).trim()

test('native Linux popup opens the exact task and retained popup can be dismissed after browser exit',async({page,browser},testInfo)=>{
  // No Notification subclass, synthetic DOM event, or injected click handler.
  await page.context().grantPermissions(['notifications'],{origin:fixture.url})
  await realLogin(page,fixture.admin)
  await page.locator('[data-app="blora.tasks"]').click()
  await page.getByRole('button',{name:'启用系统通知',exact:true}).click()
  await expect(page.getByRole('status')).toContainText('系统通知已启用')
  await expect.poll(()=>run('dunstctl','count','displayed')).toBe('0')
  const instances=(await(await page.request.get('/api/v1/instances')).json()).items
  const session=await(await page.request.get('/api/v1/session')).json()
  const path='native-popup-'+randomUUID()
  const accepted=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/files/actions`,{
    headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},
    data:{action:'mkdir',path,target:path,targetVersion:'missing'},
  })
  expect(accepted.status()).toBe(202)
  const taskId=(await accepted.json()).task.taskId
  await expect.poll(async()=>(await(await page.request.get('/api/v1/tasks/'+taskId)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')
  await expect.poll(()=>run('dunstctl','count','displayed'),{timeout:15000}).toBe('1')
  const windows=run('xdotool','search','--onlyvisible','--class','[Dd]unst').split('\n')
  expect(windows).toHaveLength(1)
  const geometry=run('xdotool','getwindowgeometry','--shell',windows[0])
  const width=Number(geometry.match(/^WIDTH=(\d+)$/m)?.[1])
  const height=Number(geometry.match(/^HEIGHT=(\d+)$/m)?.[1])
  expect(width).toBeGreaterThan(100)
  expect(height).toBeGreaterThan(20)
  const screenshot=testInfo.outputPath('native-popup.png')
  run('scrot',screenshot)
  await testInfo.attach('native-popup',{path:screenshot,contentType:'image/png'})
  // XTest mouse input targets the OS daemon's X window, outside Chromium.
  run('xdotool','mousemove','--sync','--window',windows[0],String(Math.floor(width/2)),String(Math.floor(height/2)),'click','1')
  const taskWindow=page.locator('.app-window[data-window-mode="resource"]').filter({has:page.locator('.task-card').filter({hasText:taskId})})
  await expect(taskWindow).toHaveCount(1)
  await expect(taskWindow.getByRole('heading',{name:'任务中心',exact:true})).toBeVisible()
  await expect(taskWindow.locator('.task-card')).toHaveCount(1)
  await expect(taskWindow.locator('.task-card')).toContainText(taskId)
  await expect.poll(()=>run('dunstctl','count','displayed')).toBe('0')
  const history=JSON.stringify(JSON.parse(run('dunstctl','history')))
  expect(history).toContain('任务成功')
  expect(history).toContain('file.mkdir')
  await page.reload()
  await expect(taskWindow).toHaveCount(1)
  await expect.poll(()=>run('dunstctl','count','displayed')).toBe('0')
  // Observe real browser termination with an outstanding non-persistent alert.
  // A new task is needed: reloading must not re-issue the acknowledged one.
  const secondPath='native-exit-'+randomUUID()
  const second=await page.request.post(`/api/v1/instances/${instances[0].instanceId}/files/actions`,{
    headers:{'X-CSRF-Token':session.csrfToken,'Idempotency-Key':randomUUID()},
    data:{action:'mkdir',path:secondPath,target:secondPath,targetVersion:'missing'},
  })
  expect(second.status()).toBe(202)
  await expect.poll(()=>run('dunstctl','count','displayed'),{timeout:15000}).toBe('1')
  await browser.close()
  expect(browser.isConnected()).toBe(false)
  // Dunst owns this already-delivered popup after Chromium exits. This is
  // platform characterization, not background delivery or reopen support.
  expect(run('dunstctl','count','displayed')).toBe('1')
  const retained=run('xdotool','search','--onlyvisible','--class','[Dd]unst').split('\n')
  expect(retained).toHaveLength(1)
  // Let X11 repaint the exposed notification after the browser window unmaps.
  await new Promise(resolve=>setTimeout(resolve,500))
  const afterExit=testInfo.outputPath('native-popup-after-browser-exit.png')
  run('scrot',afterExit)
  await testInfo.attach('native-popup-after-browser-exit',{path:afterExit,contentType:'image/png'})
  run('xdotool','mousemove','--sync','--window',retained[0],'40','40','click','3')
  await expect.poll(()=>run('dunstctl','count','displayed'),{timeout:10000}).toBe('0')
})
