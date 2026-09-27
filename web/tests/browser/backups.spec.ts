import {test,expect} from '@playwright/test'

test('backup UI binds source versions, explicit restore plans and schedule revisions',async({page})=>{
  const writes:{path:string;body:Record<string,unknown>}[]=[]
  const instance={instanceId:'backup-instance',nodeId:'backup-node',nodeName:'备份节点',name:'备份实例',group:'',tags:[],state:'STOPPED',configRevision:1,revision:1,config:{mode:'native',command:['/bin/true'],directory:'.',stopSeconds:30,killSeconds:10}}
  const snapshot={id:'backup-1',ownerId:'admin',source:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},path:'.',capturedVersion:'sha256:'+'1'.repeat(64),created:'2026-09-10T10:00:00Z',entries:1,total:12,archiveHash:'sha256:'+'2'.repeat(64),archiveBytes:256,state:'ready',consistency:{mode:'files'}}
  const plan={hash:'sha256:'+'3'.repeat(64),archiveHash:snapshot.archiveHash,archiveObjectId:'object-1',targetRootId:'root-1',total:1,created:'2026-09-10T10:01:00Z',entries:[{path:'.',kind:'directory',size:0,mode:493,modified:'2026-09-10T10:00:00Z',targetPath:'restored',targetVersion:'missing'}],request:{id:'plan-1',ownerId:'admin',backupId:snapshot.id,target:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},path:'restored',version:'missing'}}
  const task={taskId:'backup-task-1',requestId:'request-1',action:'backup.create',resource:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},state:'QUEUED',phase:'accepted',revision:1,createdAt:'2026-09-10T10:00:00Z'}
  await page.route('**/api/v1/**',route=>{
    const url=new URL(route.request().url()),path=url.pathname,method=route.request().method();let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'admin',name:'管理员',admin:true,disabled:false,revision:1},csrfToken:'test-only'}
    else if(path.endsWith('/instances/backup-instance'))json={instance}
    else if(path.endsWith('/instances'))json={items:[instance]}
    else if(path.endsWith('/nodes')||path.endsWith('/nodes/creatable'))json={items:[{nodeId:instance.nodeId,name:instance.nodeName,state:'ONLINE',platform:'linux',generation:1,revision:1,capabilities:{}}]}
    else if(path.endsWith('/tasks'))json={items:[]}
    else if(path.endsWith('/backups')&&method==='GET')json={items:[snapshot],nextOffset:-1}
    else if(path.endsWith('/files/stat'))json={path:url.searchParams.get('path')||'restored',kind:'directory',size:0,version:'missing'}
    else if(path.endsWith('/backups')&&method==='POST'){writes.push({path,body:route.request().postDataJSON()});json={task}}
    else if(path.endsWith('/restore-plans')&&method==='POST'){writes.push({path,body:route.request().postDataJSON()});json={plan,entryCount:1,nextOffset:-1}}
    else if(path.endsWith('/restores')&&method==='POST'){writes.push({path,body:route.request().postDataJSON()});json={task:{...task,taskId:'restore-task-1',action:'backup.restore'}}}
    else if(path.endsWith('/schedules')&&method==='GET')json={items:[]}
    else if(path.endsWith('/schedules')&&method==='POST'){writes.push({path,body:route.request().postDataJSON()});json={schedule:{spec:{id:'schedule-1',resource:{kind:'instance',id:instance.instanceId,nodeId:instance.nodeId},action:'backup.create',args:{path:'data',compression:'deflate',consistency:{mode:'files'}},cron:'0 3 * * *',timezone:'Asia/Shanghai',misfire:'skip',overlap:'skip',enabled:true},configRevision:1,nextRun:'2026-09-11T03:00:00Z',skipped:0}}}
    return route.fulfill({json})
  })
  await page.goto('/')
  await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:instance.name,exact:true}).click();await page.getByRole('button',{name:'打开备份与计划',exact:true}).click()
  await expect(page.getByRole('heading',{name:'备份实例 · 备份',exact:true})).toBeVisible();await expect(page.getByText('backup-1')).toBeVisible()
  await page.getByLabel('备份来源路径',{exact:true}).fill('data');await page.getByRole('button',{name:'读取版本并预览',exact:true}).click();await expect(page.getByRole('dialog',{name:'确认创建备份'})).toContainText('data');await page.getByRole('button',{name:'确认创建备份',exact:true}).click();await expect(page.getByRole('status')).toContainText('备份任务已接受')
  await page.getByRole('button',{name:'生成恢复计划',exact:true}).click();await expect(page.getByText('计划 plan-1')).toBeVisible();await page.getByRole('button',{name:'确认覆盖此恢复计划',exact:true}).click();await expect(page.getByRole('status')).toContainText('恢复任务已接受')
  await page.getByRole('button',{name:'＋ 新建计划',exact:true}).click();await page.getByLabel('计划备份一致性策略',{exact:true}).selectOption('save');await page.getByLabel('计划保存前钩子引用',{exact:true}).fill('backup.save.before');await page.getByLabel('计划保存后钩子引用',{exact:true}).fill('backup.save.after');await page.getByRole('button',{name:'保存定时任务',exact:true}).click();await expect(page.getByRole('status')).toContainText('定时任务已保存');const scheduleWrite=writes.find(item=>item.path==='/api/v1/schedules');expect(scheduleWrite?.body).toMatchObject({args:{consistency:{mode:'save',before:'backup.save.before',after:'backup.save.after'}}});expect(writes.map(item=>item.path)).toEqual(expect.arrayContaining(['/api/v1/instances/backup-instance/backups','/api/v1/instances/backup-instance/restore-plans','/api/v1/instances/backup-instance/restores','/api/v1/schedules']))
})
