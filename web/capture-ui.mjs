import {chromium,expect} from '@playwright/test'
import {readFileSync,mkdirSync,writeFileSync} from 'node:fs'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS,'utf8'))
const out='/data/instances/blora-panel/docs/acceptance/screenshots/2026-09-20-material-modern'
mkdirSync(out,{recursive:true})
const browser=await chromium.launch({executablePath:'/usr/bin/chromium-browser',args:['--no-sandbox','--disable-dev-shm-usage']})
const page=await browser.newPage({ignoreHTTPSErrors:true,viewport:{width:1600,height:1000},deviceScaleFactor:1.5,locale:'zh-CN'})
const shot=async name=>{await page.mouse.move(1150,80);await page.waitForTimeout(1000);await page.screenshot({path:`${out}/${name}.png`});console.log(name)}
const open=async title=>{await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:title,exact:true}).click()}
const close=async()=>{while(await page.getByRole('button',{name:'关闭窗口',exact:true}).count())await page.getByRole('button',{name:'关闭窗口',exact:true}).last().click()}
try{
 await page.goto(fixture.url)
 await shot('01-login')
 await page.getByRole('textbox',{name:'账号',exact:true}).fill(fixture.admin.name)
 await page.getByLabel('密码',{exact:true}).fill(fixture.admin.password)
 await page.getByRole('button',{name:'进入工作区 →',exact:true}).click()
 await page.locator('.desktop').waitFor()
 await shot('02-desktop')
 await page.locator('.launcher-button').click();await shot('02b-app-library');await page.locator('.launcher').screenshot({path:`${out}/02c-app-library-detail.png`});await page.locator('.launcher-button').click()
 await open('实例中心');await shot('03-instances')
 await page.locator('.topbar').screenshot({path:`${out}/00-topbar.png`})
 await page.locator('.app-window.focused').screenshot({path:`${out}/03b-instance-window.png`})
 const selection=page.locator('.instance-card input[type="checkbox"]').first()
 await selection.focus();await page.keyboard.press('Space')
 await expect(selection).toBeChecked()
 await expect(page.locator('.instance-batch-toolbar')).toContainText('已选 1 项')
 await expect(page.getByRole('button',{name:'批量启动',exact:true})).toBeEnabled()
 await expect(selection).toBeFocused()
 await page.locator('.app-window.focused').screenshot({path:`${out}/03c-instance-selection.png`})
 const control=await selection.evaluate(input=>{const s=getComputedStyle(input);return {width:s.width,height:s.height,radius:s.borderRadius,focusVisible:input.matches(':focus-visible'),outline:s.outlineStyle}})
 await page.keyboard.press('Space');await expect(selection).not.toBeChecked()
 await expect(page.locator('.instance-batch-toolbar')).toContainText('已选 0 项')
 writeFileSync(`${out}/checkbox-keyboard.json`,JSON.stringify({spaceSelects:true,spaceClears:true,batchStateUpdated:true,control},null,2)+'\n')
 await page.getByRole('button',{name:'创建实例',exact:true}).click()
 await page.getByRole('dialog',{name:'创建实例',exact:true}).getByRole('textbox',{name:'名称',exact:true}).focus()
 await shot('03d-create-dialog')
 await page.getByRole('button',{name:'收起',exact:true}).click()
 const instances=(await(await page.request.get(fixture.url+'/api/v1/instances')).json()).items
 await page.getByRole('button',{name:instances[0].name,exact:true}).click()
 await shot('04-instance-details')
 await page.getByRole('button',{name:'文件',exact:true}).click()
 await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click()
 await shot('05-files-multiwindow')
 await page.locator('.app-window.focused').screenshot({path:`${out}/05b-file-window.png`})
 await close();await open('节点管理');await shot('06-nodes')
 await close();await open('监控与进程')
 const nodes=(await(await page.request.get(fixture.url+'/api/v1/nodes')).json()).items
 await page.getByRole('combobox',{name:'监控节点'}).selectOption(nodes[0].nodeId)
 await shot('07-monitor')
 await close();await open('用户与权限');await shot('08-permissions')
 await close();await open('应用市场');await shot('09-extensions')
 await page.locator('.taskbar').screenshot({path:`${out}/11-dock-detail.png`})
 await page.getByRole('button',{name:'账号菜单',exact:true}).click();await shot('09b-account-menu');await page.locator('.account-popover').screenshot({path:`${out}/12-account-detail.png`});await page.keyboard.press('Escape')
 await page.getByRole('button',{name:'站内通知',exact:true}).click();await shot('13-notifications');await page.getByRole('button',{name:'关闭通知',exact:true}).click()
 await page.getByRole('button',{name:'切换工作区',exact:true}).click();await shot('14-workspaces');await page.keyboard.press('Escape')
 for(const [title,name] of [['备份与计划','15-backups'],['编辑器','16-editor'],['终端','17-terminal'],['任务中心','18-tasks'],['容器中心','19-containers'],['系统管理','20-system'],['设置','21-settings']]){await close();await open(title);await shot(name)}
 await close();await open('应用市场')
 await page.setViewportSize({width:390,height:844});await shot('10-mobile-market')
}finally{await browser.close()}
