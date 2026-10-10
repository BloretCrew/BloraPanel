import {readFileSync} from 'node:fs'
import {randomUUID} from 'node:crypto'
import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {realLogin} from './login'
const fixture=JSON.parse(readFileSync(process.env.BLORA_E2E_CREDENTIALS!,'utf8')) as {admin:{name:string;password:string}}
async function headers(page:Page){return {'X-CSRF-Token':(await(await page.request.get('/api/v1/session')).json()).csrfToken,'Idempotency-Key':randomUUID()}}
test('actual Docker and Compose UI preserves drafts and separates save from explicit deployment',async({page})=>{
  test.setTimeout(120000);await realLogin(page,fixture.admin);const suffix=randomUUID().slice(0,8),projectId='browser-'+suffix,volumeName='blora-browser-'+suffix
  await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'容器中心',exact:true}).click();const select=page.getByLabel('Docker 管理节点');await expect(select).not.toHaveValue('');const node=await select.inputValue(),base=`/api/v1/nodes/${node}/docker`
  async function query(kind:string){const response=await page.request.post(`${base}/query`,{headers:await headers(page),data:{kind,limit:100}});expect(response.status()).toBe(200);return response.json()}
  async function action(operation:unknown){const response=await page.request.post(`${base}/actions`,{headers:await headers(page),data:operation});expect(response.status()).toBe(202);const id=(await response.json()).task.taskId;await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${id}`)).json()).task.state,{timeout:40000}).toBe('SUCCEEDED')}
  try{
    await selectStyledOption(page,page.getByLabel('Docker 资源类别'),'volumes');await page.getByLabel('Docker 新资源名称').fill(volumeName);await page.getByRole('button',{name:'创建…',exact:true}).click();await page.reload();await page.getByRole('button',{name:'确认执行',exact:true}).click();await expect(page.locator(`[data-docker-id="${volumeName}"]`)).toBeVisible({timeout:30000});
    await selectStyledOption(page,page.getByLabel('Docker 资源类别'),'projects');await page.getByLabel('Docker 新资源名称').fill(projectId);await page.getByRole('button',{name:'新建配置草稿',exact:true}).click();const editor=page.getByRole('textbox',{name:'Compose 配置正文'});await editor.focus();await page.keyboard.press('Control+A');await page.keyboard.insertText('services: {app: {image: "blora-isolated-e2e:20260909", entrypoint: ["/bin/sh", "-c"], command: ["printf 中文-Compose; sleep 3600"]}} # 未保存中文草稿');await page.reload();await expect(page.locator('.view-lines')).toContainText('未保存中文草稿');expect((await query('projects')).projects?.find((p:{id:string})=>p.id===projectId)).toBeUndefined();await page.getByRole('button',{name:'保存配置',exact:true}).click();await expect(page.locator('.compose-editor')).toContainText('已保存版本 1',{timeout:30000});expect((await query('containers')).items?.some((c:{labels?:Record<string,string>})=>c.labels?.['com.docker.compose.project']?.includes(projectId))).toBeFalsy();
    await page.getByRole('button',{name:'返回项目列表',exact:true}).click();const row=page.locator(`[data-project-id="${projectId}"]`);await row.getByRole('button',{name:'应用此版本…',exact:true}).click();await page.getByRole('button',{name:'确认执行',exact:true}).click();await expect(row).toContainText('已应用版本 1',{timeout:40000});
    const applied=(await query('projects')).projects.find((p:any)=>p.id===projectId)
    let cursor=0,sample:any,sampleOffset=0
    do{
      const response=await page.request.post(base+'/query',{headers:await headers(page),data:{kind:'operation-output',taskId:applied.lastTaskId,outputOffset:cursor}})
      expect(response.status()).toBe(200)
      const output=await response.json();expect(Buffer.byteLength(JSON.stringify(output))).toBeLessThan(96<<10)
      if(output.output&&(output.output.stdout||output.output.stderr)&&!sample){sample=output.output;sampleOffset=cursor}
      cursor=output.nextOutput??-1
    }while(cursor>=0)
    expect(sample).toBeTruthy()
    expect(sample.completed).toBe(true)
    await page.locator('.taskbar').getByRole('button',{name:/项后台任务/}).click()
    await page.locator('.task-card').filter({hasText:applied.lastTaskId}).getByRole('button',{name:'查看任务详情'}).click()
    const outputPanel=page.getByLabel('Compose命令输出')
    await expect(outputPanel).toBeVisible()
    for(let index=0;index<sampleOffset;index++)await outputPanel.getByRole('button',{name:'下一阶段输出'}).click()
    const sampleText=Buffer.from(sample.stdout||sample.stderr,'base64').toString('utf8').trim()
    await expect(outputPanel).toContainText(sampleText)
    await expect(outputPanel).toContainText('命令已退出')
    await page.reload()
    await expect(outputPanel).toContainText(sampleText)
    await page.locator('.taskbar-app').filter({hasText:'容器中心'}).click()
    await row.getByRole('button',{name:'删除部署…',exact:true}).click();await page.getByRole('button',{name:'确认执行',exact:true}).click();await expect(row).toContainText('已删除部署',{timeout:40000});expect((await query('volumes')).items.some((v:{name:string})=>v.name===volumeName)).toBe(true)
  }finally{const projects=await query('projects'),project=projects.projects?.find((p:{id:string})=>p.id===projectId);if(project&&!project.deleted)await action({action:'compose.delete',projectId,revision:project.revision});const volume=(await query('volumes')).items?.find((v:{name:string})=>v.name===volumeName);if(volume)await action({action:'volume.delete',target:volume.target})}
})

test('real Docker log view switches between live and bounded archive',async({page})=>{
  test.setTimeout(90000);await realLogin(page,fixture.admin);const suffix=randomUUID().slice(0,8),name='browser-log-'+suffix
  await page.locator('.launcher-button').click();await page.locator('.launcher').getByRole('button',{name:'容器中心',exact:true}).click();const select=page.getByLabel('Docker 管理节点');await expect(select).not.toHaveValue('');const node=await select.inputValue(),base=`/api/v1/nodes/${node}/docker`
  async function query(){const response=await page.request.post(`${base}/query`,{headers:await headers(page),data:{kind:'containers',limit:100}});expect(response.status()).toBe(200);return response.json()}
  async function action(operation:unknown){const response=await page.request.post(`${base}/actions`,{headers:await headers(page),data:operation});expect(response.status()).toBe(202);const id=(await response.json()).task.taskId;await expect.poll(async()=>(await(await page.request.get(`/api/v1/tasks/${id}`)).json()).task.state,{timeout:40000}).toBe('SUCCEEDED')}
  let target:{kind:string;id:string}|undefined
  try{
    await action({action:'container.create',name,config:{Image:'blora-isolated-e2e:20260909',Cmd:['/bin/sh','-c',"printf '中文-Compose\\n'; sleep 3600"],HostConfig:{NetworkMode:'none'}}})
    const container=(await query()).items.find((item:{name:string})=>item.name===name) as {target:{kind:string;id:string}}|undefined;expect(container).toBeTruthy();target=container!.target;await action({action:'container.start',target})
    await selectStyledOption(page,page.getByLabel('Docker 资源类别'),'containers');await page.getByRole('button',{name:'刷新资源',exact:true}).click();const row=page.locator(`[data-docker-id="${target.id}"]`);await expect(row).toBeVisible({timeout:30000});await row.getByRole('button',{name:'日志',exact:true}).click();await expect(page.getByRole('button',{name:'查看归档',exact:true})).toBeVisible({timeout:30000});await expect(page.locator('[data-log-mode="live"]')).toContainText('中文-Compose',{timeout:30000});await page.getByRole('button',{name:'查看归档',exact:true}).click();await expect(page.locator('[data-log-mode="history"]')).toContainText('中文-Compose',{timeout:30000});await expect(page.locator('[data-log-mode="history"]')).toContainText('可能存在缺口');await page.getByRole('button',{name:'返回实时日志',exact:true}).click();await expect(page.locator('[data-log-mode="live"]')).toBeVisible()
  }finally{if(target){await action({action:'container.stop',target,stopSeconds:1});await action({action:'container.delete',target})}}
})
