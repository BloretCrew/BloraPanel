import {realLogin} from './login'
import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8')) as {url:string;admin:{name:string;password:string};instanceIds:string[]}
async function login(page:Page,name:string,password:string){await realLogin(page,{name,password})}
async function openApp(page:Page,name:string){await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name,exact:true}).click()}
async function requestHeaders(page:Page){return{'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}

test('real user creation, versioned role grants, revocation, password change and disabling',async({page,browser})=>{
  const errors:string[]=[];page.on('pageerror',error=>errors.push(error.message))
  const username=`browser-member-${randomUUID()}`,password=randomUUID()+randomUUID(),newPassword=randomUUID()+randomUUID(),roleName=`浏览器文件维护 ${randomUUID()}`
  await login(page,fixture.admin.name,fixture.admin.password);await openApp(page,'用户与权限');await page.getByRole('button',{name:'创建账号',exact:true}).click();await page.getByRole('textbox',{name:'新账号名称'}).fill(username);await page.getByLabel('初始密码',{exact:true}).fill(password);await page.getByLabel('确认初始密码',{exact:true}).fill(password);await page.getByRole('button',{name:'确认创建账号',exact:true}).click();await expect(page.locator('.user-grants h3')).toContainText(username)
  const users=(await(await page.request.get('/api/v1/users')).json()).items,user=users.find((user:{name:string})=>user.name===username);expect(user.admin).toBe(false)
  const memberContext=await browser.newContext({baseURL:fixture.url,ignoreHTTPSErrors:true,viewport:{width:1440,height:960}}),member=await memberContext.newPage();member.on('pageerror',error=>errors.push(error.message))
  try{
    await login(member,username,password);expect((await(await member.request.get('/api/v1/instances')).json()).items).toHaveLength(0)
    await page.locator('.role-card').filter({has:page.getByText('只读成员',{exact:true})}).getByRole('button',{name:'复制为自定义模板',exact:true}).click();await page.getByRole('textbox',{name:'模板名称'}).fill(roleName);await page.getByRole('checkbox',{name:'修改文件',exact:true}).check();await page.getByRole('button',{name:'保存模板',exact:true}).click();await expect(page.locator('.role-card').filter({hasText:roleName})).toHaveCount(1)
    const roles=(await(await page.request.get('/api/v1/roles')).json()).items,role=roles.find((role:{name:string})=>role.name===roleName);expect(role.revision).toBe(1)
    await selectStyledOption(page,page.getByRole('combobox',{name:'授权目标资源'}),fixture.instanceIds[0]!);await selectStyledOption(page,page.getByRole('combobox',{name:'授权角色模板'}),role.roleId);await page.getByRole('button',{name:'预览模板授权',exact:true}).click();await page.reload();await expect(page.getByRole('dialog',{name:'确认资源授权'})).toContainText(username);await page.getByRole('button',{name:'确认授权',exact:true}).click();await expect(page.locator('.grant-row')).toHaveCount(4)
    const visible=(await(await member.request.get('/api/v1/instances')).json()).items;expect(visible.map((instance:{instanceId:string})=>instance.instanceId)).toEqual([fixture.instanceIds[0]])
    const base=`/api/v1/instances/${fixture.instanceIds[0]}/files`,filename=`role-check-${randomUUID()}.txt`,write=await member.request.put(`${base}/content`,{headers:await requestHeaders(member),data:{path:filename,text:'成员按显式授权写入',version:'missing'}});expect(write.status()).toBe(202);const taskId=(await write.json()).task.taskId;await expect.poll(async()=>(await(await member.request.get(`/api/v1/tasks/${taskId}`)).json()).task.state,{timeout:30000}).toBe('SUCCEEDED')
    const forbidden=await member.request.post(`/api/v1/instances/${fixture.instanceIds[0]}/actions`,{headers:await requestHeaders(member),data:{action:'start'}});expect(forbidden.status()).toBe(403)
    await page.locator('.role-card').filter({hasText:roleName}).getByRole('button',{name:'编辑模板',exact:true}).click();await page.getByRole('checkbox',{name:'修改文件',exact:true}).uncheck();await page.getByRole('button',{name:'保存模板',exact:true}).click();await expect(page.locator('.role-card').filter({hasText:roleName})).toContainText('版本 2')
    let actual=(await(await page.request.get('/api/v1/grants')).json()).items;expect(actual.some((grant:{userId:string;action:string})=>grant.userId===user.userId&&grant.action==='file.write')).toBe(true)
    await page.locator('.grant-row').filter({hasText:'修改文件'}).getByRole('button',{name:'撤销此授权',exact:true}).click();await page.getByRole('button',{name:'确认撤销',exact:true}).click();await expect(page.locator('.grant-row')).toHaveCount(3)
    const denied=await member.request.put(`${base}/content`,{headers:await requestHeaders(member),data:{path:`denied-${randomUUID()}.txt`,text:'应被拒绝',version:'missing'}});expect(denied.status()).toBe(403)
    await openApp(member,'设置');await member.getByLabel('当前密码',{exact:true}).fill(password);await member.getByLabel('新密码',{exact:true}).fill(newPassword);await member.getByLabel('确认新密码',{exact:true}).fill(newPassword);await member.getByRole('button',{name:'保存新密码',exact:true}).click();await expect(member.getByRole('button',{name:'进入工作区 →',exact:true})).toBeVisible()
    const old=await member.request.post('/api/v1/login',{data:{name:username,password}});expect(old.status()).toBe(401);await login(member,username,newPassword)
    await page.getByRole('button',{name:'刷新账号和授权',exact:true}).click();await expect.poll(async()=>(await(await page.request.get('/api/v1/users')).json()).items.find((item:{userId:string})=>item.userId===user.userId).revision).toBe(2);await page.getByRole('button',{name:`编辑 ${username}`,exact:true}).click();await page.getByRole('checkbox',{name:'禁用账号',exact:true}).check();await page.getByRole('button',{name:'保存账号资料',exact:true}).click();await expect(page.locator('.user-card').filter({hasText:username})).toContainText('已禁用')
    expect((await member.request.get('/api/v1/session')).status()).toBe(401);expect(errors).toEqual([])
  }finally{await memberContext.close()}
})
