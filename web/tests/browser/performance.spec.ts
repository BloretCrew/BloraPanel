import {test, expect} from '@playwright/test'

// This is a repeatable client-side load probe. It deliberately reports the
// measured browser numbers instead of treating a successful build as a
// performance result; real Master/Daemon traffic is covered by the separate
// integration harness.
test('mixed desktop load probe records local interaction p95 and long frames', async ({page}) => {
  await page.route('**/api/v1/**', route => {
    const path = new URL(route.request().url()).pathname
    return route.fulfill({json: path.endsWith('/session')
      ? {user: {userId: 'perf-test', name: '性能测试', admin: true}, csrfToken: 'test-only'}
      : {items: []}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  const appButtons = page.locator('.launcher-grid > button')
  const count = await appButtons.count()
  expect(count).toBeGreaterThanOrEqual(8)
  for (let i = 0; i < 8; i++) {
    if (i > 0) await page.locator('.launcher-button').click()
    await appButtons.nth(i).click()
  }
  expect(await page.locator('.app-window').count()).toBeGreaterThanOrEqual(8)
  // Open a second terminal window through the same taskbar/window-picker path
  // users use; no fake terminal session is claimed by this local probe.
  const terminalGroup = page.locator('.taskbar-app').filter({hasText: '终端'})
  if (await terminalGroup.count()) {
    await terminalGroup.click({button: 'right'})
    await page.getByRole('menuitem', {name: '新建窗口'}).click()
    expect(await page.locator('.taskbar-app').filter({hasText: '终端'}).innerText()).toContain('2')
  }

  const sample = await page.evaluate(async () => {
    const rtts:number[] = []
    for (let i=0;i<5;i++) {
      const start=performance.now()
      await fetch('/api/v1/session', {credentials:'same-origin'})
      rtts.push(performance.now()-start)
    }
    const durations: number[] = []
    let longFrames = 0
    for (let i = 0; i < 60; i++) {
      const start = performance.now()
      document.body.dispatchEvent(new MouseEvent('mousemove', {bubbles: true, clientX: 100 + i, clientY: 120}))
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()))
      const elapsed = performance.now() - start
      durations.push(elapsed)
      if (elapsed > 50) longFrames++
    }
    durations.sort((a, b) => a - b)
    const p95 = durations[Math.ceil(durations.length * .95) - 1] || 0
    const memory = (performance as Performance & {memory?: {usedJSHeapSize: number}}).memory?.usedJSHeapSize
    rtts.sort((a,b)=>a-b)
    return {p95, max: durations[durations.length - 1] || 0, longFrames, memory, apiRttP95:rtts[Math.ceil(rtts.length*.95)-1]||0, apiRttMax:rtts[rtts.length-1]||0}
  })
  console.log(`BLORA_PERF ${JSON.stringify(sample)}`)
  expect(sample.p95).toBeLessThan(250)
})
