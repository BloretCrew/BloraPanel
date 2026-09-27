import {test,expect,type Page} from '@playwright/test'
const source={instanceId:'clipboard-source',nodeId:'source-node',name:'来源实例',state:'STOPPED',config:{},revision:1},target={...source,instanceId:'clipboard-target',nodeId:'target-node',name:'目标实例'}
const file={name:'source.txt',path:'source.txt',size:12,isDir:false,kind:'file',version:'sha256:source-original'}
async function open(page:Page,name:string){await page.locator('[data-app="blora.instances"]').click();await page.locator('.app-window.focused').getByRole('button',{name,exact:true}).click();await page.locator('.app-window.focused').getByRole('button',{name:'文件',exact:true}).click();await page.locator('.app-window.focused').getByRole('button',{name:'在文件管理器打开',exact:true}).click()}
test('cross-instance paste preserves exact request after lost receipt and shows partial cancellation facts',async({page})=>{
  const posts:{key:string;body:unknown}[]=[],errors:string[]=[];let targetExists=false,cancelled=false
  const task=()=>({taskId:'clipboard-transfer',requestId:posts[0]?.key,action:'transfer.move',resource:{kind:'instance',id:target.instanceId,nodeId:target.nodeId},state:cancelled?'CANCELLED':'WAITING_NODE',phase:'verify_destination',revision:cancelled?2:1,cancellationRequested:cancelled})
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const url=new URL(route.request().url()),path=url.pathname;let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'clipboard-user',name:'剪贴板测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[source,target]}
    else if(path.endsWith('/files'))json={items:path.includes(source.instanceId)?[file]:[],total:path.includes(source.instanceId)?1:0,version:'directory-v1',nextOffset:0}
    else if(path.endsWith('/files/stat')){if(path.includes(source.instanceId))json=file;else if(targetExists)json={...file,version:'sha256:committed-target'};else return route.fulfill({status:404,json:{error:{code:'NOT_FOUND'}}})}
    else if(path.endsWith('/transfers')&&route.request().method()==='POST'){posts.push({key:route.request().headers()['idempotency-key']!,body:route.request().postDataJSON()});if(posts.length===1){targetExists=true;return route.abort('connectionreset')}json={task:task()}}
    else if(path.endsWith('/tasks'))json={items:posts.length?[task()]:[]}
    else if(path.endsWith('/clipboard-transfer/cancel')){cancelled=true;json={task:task()}}
    else if(path.endsWith('/transfers/clipboard-transfer'))json={task:task(),transfer:{stage:'verify_destination',entries:1,completed:1,total:12,committedBytes:12,currentOffset:12,destinationVerified:false,sourceDeleted:false,sourceOutcomeUnknown:cancelled,partial:true,cleanupPending:cancelled}}
    else if(path.endsWith('/clipboard-transfer/entries'))json={items:[{relative:'source.txt',kind:'file',size:12,status:'committed'}],total:1,nextOffset:-1}
    return route.fulfill({json})
  })
  await page.goto('/');await open(page,source.name);await page.locator('.file-row').click();await page.keyboard.press('Control+x');await expect(page.locator('.file-clipboard')).toContainText('剪切 1 项');await expect(page.locator('.file-row')).toHaveClass(/cut/)
  await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click();await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click();await open(page,target.name)
  await page.getByRole('button',{name:'粘贴',exact:true}).click();await expect(page.getByRole('dialog')).toContainText(source.instanceId);await expect(page.getByRole('dialog')).toContainText(target.instanceId);await page.reload();expect(posts).toEqual([]);await expect(page.getByRole('textbox',{name:'文件操作目标'})).toHaveValue('source.txt')
  await page.getByRole('button',{name:'确认文件操作',exact:true}).click();await expect(page.locator('.files-app .error')).toBeVisible();expect(posts).toHaveLength(1);await page.reload();await expect(page.getByRole('textbox',{name:'文件操作目标'})).toBeDisabled();expect(posts).toHaveLength(1)
  await page.getByRole('button',{name:'确认文件操作',exact:true}).click();await expect(page.getByRole('dialog')).toHaveCount(0);expect(posts).toHaveLength(2);expect(posts[1]).toEqual(posts[0]);expect(posts[0]!.body).toEqual({source:{instanceId:source.instanceId,path:file.path,version:file.version},target:{instanceId:target.instanceId,path:file.path,version:'missing'},move:true})
  await page.getByRole('button',{name:'详情',exact:true}).click();await expect(page.getByRole('region',{name:'文件传输阶段'})).toContainText('部分目标内容已经提交');await expect(page.getByRole('button',{name:'下一页传输条目',exact:true})).toBeDisabled();await page.getByRole('button',{name:'请求取消后续工作',exact:true}).click();await expect(page.getByRole('region',{name:'文件传输阶段'})).toContainText('源删除结果不明');await expect(page.getByRole('region',{name:'文件传输阶段'})).toContainText('临时内容仍待节点清理');expect(errors).toEqual([])
})
