import {expect, test} from '@playwright/test'

test('system and workspace reduced motion both keep the launcher stationary on entry',async({page})=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:new URL(route.request().url()).pathname.endsWith('/session')?{user:{userId:'motion-test',name:'动态效果测试',admin:true},csrfToken:'test-only'}:{items:[]}}))
  await page.emulateMedia({reducedMotion:'reduce'})
  await page.goto('/')
  // Observe actual entry frames, including inline updates by the animation library.
  await page.evaluate(()=>{
    const frames:number[][]=[]
    Object.assign(window,{launcherFrames:frames})
    new MutationObserver(()=>{
      const launcher=document.querySelector('.launcher')
      if(launcher){const style=getComputedStyle(launcher);frames.push([new DOMMatrixReadOnly(style.transform).m42,Number(style.opacity)])}
    }).observe(document.body,{subtree:true,childList:true,attributes:true,attributeFilter:['style']})
  })
  const stationary=async()=>{
    await page.evaluate(()=>{(window as unknown as {launcherFrames:number[][]}).launcherFrames.length=0})
    await page.locator('.launcher-button').click()
    await expect(page.locator('.launcher')).toBeVisible()
    const frames=await page.evaluate(()=>(window as unknown as {launcherFrames:number[][]}).launcherFrames)
    expect(frames.length).toBeGreaterThan(0)
    expect(frames.every(([y,opacity])=>Math.abs(y!)<.01&&opacity===1)).toBe(true)
  }
  await stationary()
  await page.locator('.launcher').getByRole('button',{name:'设置',exact:true}).click()
  await page.getByRole('checkbox',{name:'减少动态效果'}).check()
  await page.emulateMedia({reducedMotion:'no-preference'})
  await stationary()
})

test('desktop controls remain reachable on narrow screens and account actions work',async({page})=>{
  let signedIn=true
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    if(path.endsWith('/logout'))signedIn=false
    return route.fulfill({json:path.endsWith('/session')?{user:signedIn?{userId:'bar-test',name:'测试管理员',admin:true}:null,csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:{items:[]}})
  })
  await page.goto('/')
  const bar=page.getByRole('banner',{name:'桌面控制栏'})
  for(const width of [1440,390]){
    await page.setViewportSize({width,height:960})
    const bounds=(await bar.boundingBox())!
    for(const name of ['切换工作区','站内通知','账号菜单']){
      const box=(await bar.getByRole('button',{name,exact:true}).boundingBox())!
      expect(box.x).toBeGreaterThanOrEqual(bounds.x)
      expect(box.x+box.width).toBeLessThanOrEqual(bounds.x+bounds.width)
    }
    await bar.getByRole('button',{name:'账号菜单',exact:true}).click()
    await expect(page.getByRole('menu',{name:'账号菜单'})).toContainText('测试管理员')
    await page.keyboard.press('Escape')
    await expect(page.getByRole('menu',{name:'账号菜单'})).toBeHidden()
    await bar.getByRole('button',{name:'切换工作区'}).click()
    await expect(page.getByRole('dialog',{name:'工作区',exact:true})).toBeVisible()
    await page.keyboard.press('Escape')
  }
  await bar.getByRole('button',{name:'账号菜单',exact:true}).click()
  await page.getByRole('menuitem',{name:'桌面设置'}).click()
  await expect(page.locator('.app-window')).toHaveCount(1)
  await expect(page.locator('.window-title')).toHaveText('设置')
  await bar.getByRole('button',{name:'账号菜单',exact:true}).click()
  const downloading=page.waitForEvent('download')
  await page.getByRole('menuitem',{name:'导出工作现场'}).click()
  expect((await downloading).suggestedFilename()).toBe('blora-workspace.json')
  await bar.getByRole('button',{name:'账号菜单',exact:true}).click()
  page.once('dialog',dialog=>dialog.accept())
  await page.getByRole('menuitem',{name:'退出账号'}).click()
  await expect(page.locator('.login-screen')).toBeVisible()
})

test('workspace preferences persist and layout reset keeps a backup workspace',async({page})=>{
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    return route.fulfill({json:path.endsWith('/session')?{user:{userId:'settings-test',name:'设置测试',admin:true},csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:{items:[]}})
  })
  await page.goto('/')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'设置',exact:true}).click()
  await page.locator('.settings-panel label').filter({hasText:/^主题(?!配色)/}).locator('select').selectOption('dark')
  await page.getByLabel('字号').selectOption('1.1')
  await page.getByLabel('界面密度').selectOption('compact')
  await page.getByLabel('应用启动器快捷键').selectOption('off')
  await page.getByLabel('窗口切换快捷键').selectOption('off')
  await expect(page.locator('html')).toHaveAttribute('data-theme','dark')
  await expect(page.locator('html')).toHaveAttribute('data-density','compact')
  // Firefox quantizes layout values to fractions of a CSS pixel (14.2969px).
  // Still reject a wrong scale; tolerate only subpixel serialization rounding.
  await expect.poll(async()=>Math.abs(await page.locator('html').evaluate(el=>parseFloat(getComputedStyle(el).fontSize))-14.3)).toBeLessThan(0.01)
  page.once('dialog',dialog=>dialog.accept())
  await page.getByRole('button',{name:'恢复默认布局'}).click()
  await expect(page.getByRole('status')).toContainText('已恢复默认布局')
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme','dark')
  await expect(page.locator('html')).toHaveAttribute('data-density','compact')
  await page.locator('.launcher-button').click()
  await page.locator('.launcher').getByRole('button',{name:'设置',exact:true}).click()
  await expect(page.getByLabel('字号')).toHaveValue('1.1')
  await expect(page.getByLabel('应用启动器快捷键')).toHaveValue('off')
  await expect(page.getByLabel('窗口切换快捷键')).toHaveValue('off')
})
