import {test,expect,type Page} from '@playwright/test'

async function verifyDownloadAfterEditorRejects(page:Page,filename:string,status:number,code:'TEXT_LIMIT'|'UNSUPPORTED_ENCODING',message:string){
  await page.route('**/api/v1/**',route=>{
    const path=new URL(route.request().url()).pathname
    let json:unknown={items:[]}
    if(path.endsWith('/session'))json={user:{userId:'boundary-user',name:'边界测试',admin:true},csrfToken:'test-only'}
    else if(path.endsWith('/instances'))json={items:[{instanceId:'boundary-instance',nodeId:'boundary-node',nodeName:'文件边界节点',name:'文件边界实例',state:'STOPPED',config:{},revision:1}]}
    else if(path.endsWith('/instances/boundary-instance'))json={instance:{instanceId:'boundary-instance',nodeId:'boundary-node',nodeName:'文件边界节点',name:'文件边界实例',state:'STOPPED',config:{},revision:1}}
    else if(path.endsWith('/files'))json={items:[{name:filename,path:filename,isDir:false,size:status===413?4194305:8,version:'boundary-version'}],total:1,version:'directory-version'}
    else if(path.endsWith('/files/access'))json={instanceId:'boundary-instance',nodeId:'boundary-node',nodeName:'文件边界节点',nodeState:'ONLINE',canRead:true,canWrite:true}
    else if(path.endsWith('/files/content'))return route.fulfill({status,json:{error:{code,message}}})
    else if(path.endsWith('/files/preview')){const url=new URL(route.request().url()),offset=Number(url.searchParams.get('offset')||0);json={text:offset?'second preview':'first preview',version:'boundary-version',offset,nextOffset:offset+1024,total:2048,hasMore:offset===0,encoding:'UTF-8',newline:'LF',maxBytes:4194304}}
    return route.fulfill({json})
  })
  await page.goto('/');await page.locator('[data-app="blora.instances"]').click();await page.getByRole('button',{name:'文件边界实例',exact:true}).click();await page.getByRole('button',{name:'文件',exact:true}).click();await page.getByRole('button',{name:'在文件管理器打开',exact:true}).click();await page.getByRole('button',{name:filename,exact:true}).click()
  await expect(page.locator('.error.notice')).toContainText(message)
  const link=page.getByRole('link',{name:'下载原文件',exact:true})
  await expect(link).toHaveAttribute('href',`/api/v1/instances/boundary-instance/files/download?path=${encodeURIComponent(filename)}`)
  await expect(link).toHaveAttribute('download',filename)
}

test('oversized text opens with an original-file download action',async({page})=>{
  await verifyDownloadAfterEditorRejects(page,'large.txt',413,'TEXT_LIMIT','编辑器仅支持至多4 MiB的UTF-8文本；请使用下载')
})

test('binary content opens with an original-file download action',async({page})=>{
  await verifyDownloadAfterEditorRejects(page,'image.bin',415,'UNSUPPORTED_ENCODING','此文件不是可编辑UTF-8文本；请下载或使用专用工具')
})

test('oversized text supports bounded read-only preview navigation',async({page})=>{
  await verifyDownloadAfterEditorRejects(page,'large.txt',413,'TEXT_LIMIT','文件超过编辑器 4 MiB 上限')
  await expect(page.locator('.editor-preview')).toBeVisible()
  await expect(page.locator('.editor-preview-text')).toHaveText('first preview')
  await page.getByRole('button',{name:'下一段',exact:true}).click()
  await expect(page.locator('.editor-preview-text')).toHaveText('second preview')
  await expect(page.getByRole('button',{name:'下一段',exact:true})).toBeDisabled()
  await page.getByRole('button',{name:'上一段',exact:true}).click()
  await expect(page.locator('.editor-preview-text')).toHaveText('first preview')
})
