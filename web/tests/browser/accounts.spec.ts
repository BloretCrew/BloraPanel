import {type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {test,expect} from '../helpers/management-fixture'
async function openApp(page:Page,name:string){await page.goto('/');await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name,exact:true}).click()}
async function recoveryRecords(page:Page){return page.evaluate(async()=>{
  const db=await new Promise<IDBDatabase>((resolve,reject)=>{const request=indexedDB.open('blora-workspaces',1);request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)})
  const data=await new Promise<unknown[]>((resolve,reject)=>{const request=db.transaction('snapshots').objectStore('snapshots').getAll();request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(request.error)});db.close();return JSON.stringify({session:Object.entries(sessionStorage),local:Object.entries(localStorage),data})
})}
const admin={userId:'account-admin',name:'管理测试员',admin:true,disabled:false,revision:3}

test('rejected login clears submitted credentials immediately',async({page})=>{
  let submitted=false
  await page.context().route('**/api/v1/**',route=>{if(route.request().url().endsWith('/login')){submitted=true;return route.fulfill({status:429,json:{error:{code:'RATE_LIMIT',message:'登录尝试过多，请稍后重试'}}})}return route.fulfill({status:401,json:{error:{code:'UNAUTHENTICATED'}}})})
  await page.goto('/');await page.getByRole('textbox',{name:'账号',exact:true}).fill('test-account');await page.getByLabel('密码',{exact:true}).fill('only-a-test-boundary-value');await page.getByRole('button',{name:'进入工作区 →'}).click();await expect(page.getByLabel('密码',{exact:true})).toHaveValue('');await expect(page.getByRole('alert')).toContainText('登录尝试过多');expect(submitted).toBe(true)
})

test('account and permission confirmations restore while passwords never enter workspace recovery',async({page})=>{
  const users=[admin],roles=[{roleId:'reader',name:'只读成员',actions:['instance.read','file.read'],nodeOnly:false,builtin:true,revision:1}],grants:{userId:string;resource:{kind:string;id:string};action:string}[]=[],writes:{path:string;body:Record<string,unknown>}[]=[],errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.context().route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname,method=route.request().method();let json:unknown={items:[]}
    if(method!=='GET')writes.push({path,body:route.request().postDataJSON()})
    if(path.endsWith('/session'))json={user:admin,csrfToken:'test-only'}
    else if(path.endsWith('/tasks/summary'))json={active:0,states:{}}
    else if(path.endsWith('/users')){if(method==='POST'){const body=route.request().postDataJSON(),user={userId:'created-member',name:body.name,admin:body.admin,disabled:false,revision:1};users.push(user);json={user}}else json={items:users}}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'authorized-instance',nodeId:'authorized-node',nodeName:'授权节点',name:'授权实例',state:'STOPPED',config:{mode:'native'},revision:1}]}
    else if(path.endsWith('/nodes'))json={items:[{nodeId:'authorized-node',name:'授权节点',state:'ONLINE',revision:1}]}
    else if(path.endsWith('/grants'))json={items:grants}
    else if(path.endsWith('/roles'))json={items:roles,semantics:'templates-expand-to-explicit-resource-grants'}
    else if(path.endsWith('/roles/reader/apply')){const body=route.request().postDataJSON();for(const action of roles[0]!.actions)grants.push({userId:body.userId,resource:body.resource,action});json={applied:true}}
    else if(path.includes('/roles/')&&method==='PUT'){const body=route.request().postDataJSON(),role={...body,roleId:path.split('/').at(-1),revision:1,builtin:false};roles.push(role);json={role}}
    return route.fulfill({json})
  })
  await openApp(page,'用户与权限');await page.getByRole('button',{name:'创建账号',exact:true}).click();await page.getByRole('textbox',{name:'新账号名称'}).fill('可恢复成员草稿')
  const secret='test-only-never-persist-93847';await page.getByLabel('初始密码',{exact:true}).fill(secret);await page.getByLabel('确认初始密码',{exact:true}).fill(secret);expect(await recoveryRecords(page)).not.toContain(secret)
  await page.reload();await expect(page.getByRole('textbox',{name:'新账号名称'})).toHaveValue('可恢复成员草稿');await expect(page.getByLabel('初始密码',{exact:true})).toHaveValue('');await expect(page.getByLabel('确认初始密码',{exact:true})).toHaveValue('');expect(writes).toEqual([])
  await page.bringToFront()
  await page.getByLabel('初始密码',{exact:true}).fill(secret);await page.getByLabel('确认初始密码',{exact:true}).fill(secret)
  await expect(page.getByLabel('初始密码',{exact:true})).toHaveValue(secret);await expect(page.getByLabel('确认初始密码',{exact:true})).toHaveValue(secret)
  const confirmCreate=page.getByRole('button',{name:'确认创建账号',exact:true})
  await confirmCreate.scrollIntoViewIfNeeded();await confirmCreate.click();await expect(page.locator('.user-grants h3')).toContainText('可恢复成员草稿')
  await selectStyledOption(page,page.getByRole('combobox',{name:'授权目标资源'}),'authorized-instance');await page.getByRole('button',{name:'预览模板授权',exact:true}).click();await page.reload();await expect(page.getByRole('dialog',{name:'确认资源授权'})).toContainText('授权实例');expect(writes).toHaveLength(1);await page.getByRole('button',{name:'确认授权',exact:true}).click();await expect(page.locator('.grant-row')).toHaveCount(2);expect(writes.filter(write=>write.path.endsWith('/apply'))).toHaveLength(1)
  await page.getByRole('button',{name:'创建自定义模板',exact:true}).click();await page.getByRole('textbox',{name:'模板名称'}).fill('文件维护模板草稿');await page.getByRole('checkbox',{name:'修改文件',exact:true}).check();await page.reload();await expect(page.getByRole('textbox',{name:'模板名称'})).toHaveValue('文件维护模板草稿');await expect(page.getByRole('checkbox',{name:'修改文件',exact:true})).toBeChecked();expect(writes).toHaveLength(2)
  await page.getByRole('button',{name:'保存模板',exact:true}).click();await expect(page.locator('.role-card').filter({hasText:'文件维护模板草稿'})).toHaveCount(1);expect(writes.at(-1)!.body).toMatchObject({revision:0,actions:['instance.read','file.write']});expect(await recoveryRecords(page)).not.toContain(secret);expect(errors).toEqual([])
})

test('password changes clear hidden fields, exclude recovery and require a new login after confirmation',async({page})=>{
  const writes:{path:string;body:Record<string,unknown>}[]=[],secret='test-only-password-change-443'
  await page.context().route('**/api/v1/**',route=>{const path=new URL(route.request().url()).pathname;if(route.request().method()!=='GET')writes.push({path,body:route.request().postDataJSON()});return route.fulfill({json:path.endsWith('/session')?{user:admin,csrfToken:'test-only'}:path.endsWith('/tasks/summary')?{active:0,states:{}}:path.endsWith('/password')?{changed:true,sessionsRevoked:true}:{items:[]}})})
  await openApp(page,'设置');await page.getByLabel('当前密码',{exact:true}).fill('test-only-current-password');await page.getByLabel('新密码',{exact:true}).fill(secret);await page.getByLabel('确认新密码',{exact:true}).fill(secret);expect(await recoveryRecords(page)).not.toContain(secret)
  await page.getByRole('button',{name:'最小化窗口',exact:true}).click();await page.locator('.taskbar-app').click();await expect(page.getByLabel('新密码',{exact:true})).toHaveValue('');await expect(page.getByLabel('当前密码',{exact:true})).toHaveValue('')
  await page.getByLabel('当前密码',{exact:true}).fill('test-only-current-password');await page.getByLabel('新密码',{exact:true}).fill(secret);await page.getByLabel('确认新密码',{exact:true}).fill(secret);await page.reload();await expect(page.getByLabel('新密码',{exact:true})).toHaveValue('');expect(writes).toEqual([])
  await page.getByLabel('当前密码',{exact:true}).fill('test-only-current-password');await page.getByLabel('新密码',{exact:true}).fill(secret);await page.getByLabel('确认新密码',{exact:true}).fill(secret);await page.getByRole('button',{name:'保存新密码',exact:true}).click();await expect(page.getByRole('button',{name:'进入工作区 →',exact:true})).toBeVisible();expect(writes).toHaveLength(1);expect(writes[0]!.body.revision).toBe(3);expect(await recoveryRecords(page)).not.toContain(secret)
})
