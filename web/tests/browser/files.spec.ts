import {test,expect,type Page} from '@playwright/test'
import {selectStyledOption} from '../helpers/styled-select'
import {createHash} from 'node:crypto'

async function openFiles(page:Page){
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'文件测试实例',exact:true}).click();await page.getByRole('button',{name:'文件',exact:true}).click();await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click()
}
const identity={user:{userId:'file-user',name:'文件测试',admin:true},csrfToken:'test-only'}
const instance={instanceId:'file-instance',nodeId:'file-node',nodeName:'文件测试节点',name:'文件测试实例',state:'STOPPED',config:{},revision:1}

test('virtual directory follows reduced node pages and restores scrolling across 10000 entries',async({page})=>{
  const requests:{offset:number;version:string;search:string;sort:string;order:string}[]=[],errors:string[]=[]
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',route=>{
    const url=new URL(route.request().url());let json:unknown={items:[]}
    if(url.pathname.endsWith('/session'))json=identity
    else if(url.pathname.endsWith('/instances'))json={items:[instance]}
    else if(url.pathname.endsWith('/files')){
      const offset=Number(url.searchParams.get('offset')),limit=Math.min(7,Number(url.searchParams.get('limit'))),search=url.searchParams.get('search')||'',sort=url.searchParams.get('sort')||'',order=url.searchParams.get('order')||'',version=url.searchParams.get('version')||''
      if(url.searchParams.has('version'))requests.push({offset,version,search,sort,order})
      const total=search?1:10000,size=Math.min(limit,total-offset)
      json={items:Array.from({length:size},(_,index)=>{const number=search?9876:offset+index;return{name:`file-${String(number).padStart(5,'0')}.txt`,path:`file-${String(number).padStart(5,'0')}.txt`,size:number,isDir:false}}),total,version:'directory-version',nextOffset:offset+size<total?offset+size:-1}
    }
    return route.fulfill({json})
  })
  await openFiles(page);await expect(page.getByRole('grid',{name:'节点文件列表'})).toHaveAttribute('aria-rowcount','10000')
  const marqueeViewport=page.getByRole('grid',{name:'节点文件列表'});await marqueeViewport.evaluate(node=>{const rect=node.getBoundingClientRect();const frame=(type:string,x:number,y:number)=>new PointerEvent(type,{bubbles:true,clientX:rect.left+x,clientY:rect.top+y,pointerId:17,button:0});node.dispatchEvent(frame('pointerdown',8,4));window.dispatchEvent(frame('pointermove',240,150));window.dispatchEvent(frame('pointerup',240,150))});await expect(page.locator('.file-row.selected')).toHaveCount(5);await page.locator('.file-row').first().click()
  await expect(page.locator('.file-row')).toHaveCount(100);expect(requests.slice(0,3).map(request=>request.offset)).toEqual([0,7,14]);expect(requests[1]!.version).toBe('directory-version')
  const viewport=page.getByRole('grid',{name:'节点文件列表'})
  await viewport.evaluate(element=>{element.scrollTop=9999*36})
  await expect(page.getByRole('button',{name:'file-09999.txt',exact:true})).toBeVisible();await expect(page.locator('.file-row')).toHaveCount(100)
  await page.reload();await expect(page.getByRole('button',{name:'file-09999.txt',exact:true})).toBeVisible()
  await page.getByRole('button',{name:'图标',exact:true}).click();await expect(page.getByRole('grid',{name:'节点文件图标视图'})).toBeVisible();await expect(page.locator('.file-icon-card')).toHaveCount(100)
  await page.getByRole('button',{name:'下一页',exact:true}).click();await expect(page.getByRole('gridcell',{name:/file-00100\.txt/})).toBeVisible();await page.reload();await expect(page.getByRole('gridcell',{name:/file-00100\.txt/})).toBeVisible();await page.getByRole('button',{name:'列表',exact:true}).click()
  await page.getByRole('textbox',{name:'搜索目录文件'}).fill('9876');await expect(viewport).toHaveAttribute('aria-rowcount','1');await expect(page.getByRole('button',{name:'file-09876.txt',exact:true})).toBeVisible()
  await selectStyledOption(page,page.getByRole('combobox',{name:'文件排序'}),'modified');await selectStyledOption(page,page.getByRole('combobox',{name:'排序方向'}),'desc');await expect.poll(()=>requests.at(-1)).toMatchObject({offset:0,search:'9876',sort:'modified',order:'desc'});expect(errors).toEqual([])
})

test('directory tree loads expanded branches on demand and restores the selected path',async({page})=>{
  const treeRequests:string[]=[];page.on('pageerror',error=>{throw error})
  await page.route('**/api/v1/**',route=>{
    const url=new URL(route.request().url());let json:unknown={items:[]}
    if(url.pathname.endsWith('/session'))json=identity
    else if(url.pathname.endsWith('/instances'))json={items:[instance]}
    else if(url.pathname.endsWith('/files')){
      const directory=url.searchParams.get('path')||'.',treeRequest=!url.searchParams.has('version')
      if(treeRequest)treeRequests.push(directory)
      if(directory==='.')json={items:[{name:'alpha',path:'alpha',isDir:true,size:0},{name:'root.txt',path:'root.txt',isDir:false,size:4}],total:2,version:'root-version',nextOffset:0}
      else if(directory==='alpha')json={items:[{name:'beta',path:'alpha/beta',isDir:true,size:0},{name:'alpha.txt',path:'alpha/alpha.txt',isDir:false,size:5}],total:2,version:'alpha-version',nextOffset:0}
      else if(directory==='alpha/beta')json={items:[{name:'nested.txt',path:'alpha/beta/nested.txt',isDir:false,size:6}],total:1,version:'beta-version',nextOffset:0}
      else json={items:[],total:0,version:'empty-version',nextOffset:0}
    }
    return route.fulfill({json})
  })
  await openFiles(page)
  await expect(page.getByRole('tree',{name:'目录树'})).toBeVisible()
  const alpha=page.getByRole('treeitem',{name:/alpha/}).first();await expect(alpha).toBeVisible()
  await page.getByRole('button',{name:'展开目录 alpha',exact:true}).click();await expect(page.getByRole('button',{name:'beta',exact:true})).toBeVisible()
  await page.getByRole('button',{name:'展开目录 beta',exact:true}).click();await expect(page.getByRole('button',{name:'nested.txt',exact:true})).toHaveCount(0)
  await page.getByRole('button',{name:'beta',exact:true}).click();await expect(page.getByRole('textbox',{name:'目录路径'})).toHaveValue('alpha/beta');await expect(page.getByRole('button',{name:'nested.txt',exact:true})).toBeVisible()
  expect(treeRequests).toEqual(expect.arrayContaining(['.','alpha','alpha/beta']))
  await page.reload();await expect(page.getByRole('textbox',{name:'目录路径'})).toHaveValue('alpha/beta');await expect(page.getByRole('button',{name:'nested.txt',exact:true})).toBeVisible()
})

async function pickFile(page:Page,changed=false){
  await page.locator('input[type="file"]').evaluate((input,changed)=>{
    const bytes=new Uint8Array(3*65536);for(let i=0;i<bytes.length;i++)bytes[i]=i%251;if(changed)bytes[0]=99
    const transfer=new DataTransfer();transfer.items.add(new File([bytes],'resume.bin',{lastModified:1700000000000}));(input as HTMLInputElement).files=transfer.files;input.dispatchEvent(new Event('change',{bubbles:true}))
  },changed)
}

test('upload reload uses node checkpoint, verifies reselected source and keeps stable request identity',async({page})=>{
  const starts:{key:string;body:Record<string,unknown>}[]=[],chunks:number[]=[],chunkKeys:string[]=[],completeKeys:string[]=[],errors:string[]=[]
  let offset=0,failed=false,completed=false,spec:Record<string,unknown>={}
  const task=()=>({taskId:'upload-task',requestId:starts[0]?.key,resource:{kind:'instance',id:instance.instanceId},action:'file.upload',state:completed?'SUCCEEDED':'WAITING_CLIENT',phase:completed?'committed':'receiving'})
  const reply=()=>({task:task(),upload:{spec,offset,chunkBytes:65536,stage:completed?'committed':'receiving'}})
  page.on('pageerror',error=>errors.push(error.message))
  await page.route('**/api/v1/**',async route=>{
    const url=new URL(route.request().url()),method=route.request().method();let json:unknown={items:[]}
    if(url.pathname.endsWith('/session'))json=identity
    else if(url.pathname.endsWith('/instances'))json={items:[instance]}
    else if(url.pathname.endsWith('/files'))json={items:[],total:0,version:'empty-directory',nextOffset:0}
    else if(url.pathname.endsWith('/files/stat'))return route.fulfill({status:404,json:{error:{code:'NOT_FOUND',message:'目标尚不存在'}}})
    else if(url.pathname.endsWith('/files/uploads')){spec=route.request().postDataJSON();starts.push({key:route.request().headers()['idempotency-key']!,body:spec});json=reply()}
    else if(url.pathname.endsWith('/chunks')){
      const body=route.request().postDataJSON(),bytes=Buffer.from(body.data,'base64');expect(body.hash).toBe('sha256:'+createHash('sha256').update(bytes).digest('hex'));expect(body.offset).toBe(offset);chunkKeys.push(route.request().headers()['idempotency-key']!);chunks.push(body.offset);offset+=bytes.length
      // The node persisted this chunk but its HTTP acknowledgement was lost.
      if(offset===2*65536&&!failed){failed=true;return route.abort('connectionreset')}
      json=reply()
    }else if(url.pathname.endsWith('/complete')){completeKeys.push(route.request().headers()['idempotency-key']!);completed=true;json=reply()}
    else if(url.pathname.endsWith('/uploads/upload-task'))json=reply()
    else if(url.pathname.endsWith('/tasks'))json={items:starts.length?[task()]:[]}
    return route.fulfill({json})
  })
  await openFiles(page);await pickFile(page);await page.getByRole('button',{name:'确认文件操作',exact:true}).click()
  await expect(page.locator('.upload-row')).toContainText('等待本机文件或重试');expect(chunks).toEqual([0,65536]);await expect(page.locator('.upload-row')).toContainText('65536 / 196608')
  await page.reload();await expect(page.locator('.upload-row')).toContainText('131072 / 196608');expect(starts).toHaveLength(1)
  // Same metadata with changed bytes must never resume a different source.
  const chooser1=page.waitForEvent('filechooser');await page.getByRole('button',{name:'选择原文件续传',exact:true}).click();await chooser1;await pickFile(page,true)
  await expect(page.locator('.upload-row')).toContainText('内容指纹与原文件不符');expect(starts).toHaveLength(1)
  const chooser2=page.waitForEvent('filechooser');await page.getByRole('button',{name:'选择原文件续传',exact:true}).click();await chooser2;await pickFile(page)
  await expect(page.locator('.upload-row')).toContainText('节点确认已提交');expect(starts).toHaveLength(2);expect(starts[0]!.key).toBe(starts[1]!.key);expect(starts[0]!.body).toEqual(starts[1]!.body);expect(chunks).toEqual([0,65536,131072]);expect(chunkKeys).toEqual([`${starts[0]!.key}:chunk:0`,`${starts[0]!.key}:chunk:65536`,`${starts[0]!.key}:chunk:131072`]);expect(completeKeys).toEqual([`${starts[0]!.key}:complete`]);expect(errors).toEqual([])
})

test('lost upload creation and cancellation receipts remain unconfirmed across reload',async({page})=>{
  let spec:Record<string,unknown>={},key='',starts=0,cancels=0,cancelled=false
  const task=()=>({taskId:'uncertain-upload',requestId:key,resource:{kind:'instance',id:instance.instanceId},action:'file.upload',state:cancelled?'CANCELLED':'WAITING_CLIENT',phase:'receiving'})
  const reply=()=>({task:task(),upload:{spec,offset:0,chunkBytes:65536,stage:cancelled?'cancelled':'receiving'}})
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname;let json:unknown={items:[]}
    if(path.endsWith('/session'))json=identity
    else if(path.endsWith('/instances'))json={items:[instance]}
    else if(path.endsWith('/files'))json={items:[],total:0,version:'empty',nextOffset:0}
    else if(path.endsWith('/files/stat'))return route.fulfill({status:404,json:{error:{code:'NOT_FOUND'}}})
    else if(path.endsWith('/files/uploads')){starts++;spec=route.request().postDataJSON();key=route.request().headers()['idempotency-key']!;return route.abort('connectionreset')}
    else if(path.endsWith('/tasks'))json={items:starts?[task()]:[]}
    else if(path.endsWith('/uploads/uncertain-upload')){if(route.request().method()==='DELETE'){cancels++;if(cancels===1)return route.abort('connectionreset');cancelled=true}json=reply()}
    return route.fulfill({json})
  })
  await openFiles(page);await pickFile(page);await page.getByRole('button',{name:'确认文件操作',exact:true}).click();await expect(page.locator('.upload-row')).toContainText('等待本机文件或重试')
  await page.getByRole('button',{name:'请求取消',exact:true}).click();await expect.poll(()=>cancels).toBe(1);await expect(page.locator('.upload-row')).toContainText('cancel_requested');await expect(page.locator('.upload-row')).not.toContainText('已取消')
  await page.reload();await expect(page.locator('.upload-row')).toContainText('cancel_requested');expect(starts).toBe(1);expect(cancels).toBe(1)
  await page.getByRole('button',{name:'请求取消',exact:true}).click();await expect(page.locator('.upload-row')).toContainText('已取消');expect(cancels).toBe(2)
})

test('directory shortcuts preserve distinct resources and rename drafts without changing node files',async({page})=>{
  const writes:string[]=[]
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname;if(route.request().method()!=='GET')writes.push(path)
    return route.fulfill({json:path.endsWith('/session')?identity:path.endsWith('/instances')?{items:[instance]}:path.endsWith('/files')?{items:[],total:0,version:'empty-directory',nextOffset:0}:{items:[]}})
  })
  await openFiles(page)
  for(const directory of ['alpha','beta']){await page.getByRole('textbox',{name:'目录路径'}).fill(directory);await page.getByRole('button',{name:'转到',exact:true}).click();await page.getByRole('button',{name:'固定目录到桌面',exact:true}).click()}
  await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click();await page.locator('.app-window.focused').getByRole('button',{name:'关闭窗口',exact:true}).click()
  const alpha=page.locator('[data-shortcut]').filter({hasText:'alpha'}),beta=page.locator('[data-shortcut]').filter({hasText:'beta'})
  await alpha.click();await expect(page.locator('.app-window.focused').getByRole('textbox',{name:'目录路径'})).toHaveValue('alpha');await beta.click({position:{x:5,y:10}});await expect(page.locator('.app-window.focused').getByRole('textbox',{name:'目录路径'})).toHaveValue('beta');await expect(page.locator('.app-window')).toHaveCount(2)
  await alpha.click({button:'right',position:{x:5,y:10}});await page.getByRole('menuitem',{name:'重命名入口',exact:true}).click();await page.getByRole('textbox',{name:'桌面入口名称'}).fill('测试目录入口草稿');await page.reload();await expect(page.getByRole('textbox',{name:'桌面入口名称'})).toHaveValue('测试目录入口草稿');await page.getByRole('button',{name:'保存入口名称',exact:true}).click();await expect(page.locator('[data-shortcut]').filter({hasText:'测试目录入口草稿'})).toHaveCount(1)
  await page.locator('[data-shortcut]').filter({hasText:'测试目录入口草稿'}).click({position:{x:5,y:10}});await expect(page.locator('.app-window.focused').getByRole('textbox',{name:'目录路径'})).toHaveValue('alpha');await expect(page.locator('.app-window')).toHaveCount(2);expect(writes).toEqual([])
})
