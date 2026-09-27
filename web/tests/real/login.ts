import {test,expect,type Page} from '@playwright/test'
// Real suites use the same loopback address and retain the production login
// limit. One explicit retry waits out its one-minute window instead of bypassing
// authentication or changing the server's protection for the test environment.
export async function realLogin(page:Page,account:{name:string;password:string}){
  await page.goto('/')
  for(let attempt=0;attempt<2;attempt++){
    await page.getByRole('textbox',{name:'账号',exact:true}).fill(account.name);await page.getByLabel('密码',{exact:true}).fill(account.password)
    const response=page.waitForResponse(response=>response.url().endsWith('/api/v1/login')&&response.request().method()==='POST')
    await page.getByRole('button',{name:'进入工作区 →',exact:true}).click();const result=await response
    if(result.status()!==429||attempt===1){expect(result.status()).toBe(200);break}
    test.setTimeout(test.info().timeout+65000);test.info().annotations.push({type:'rate-limit',description:'真实登录返回429，等待61秒后只重试一次；未更改服务端限制。'})
    await page.waitForTimeout(30000);await page.waitForTimeout(30000);await page.waitForTimeout(1000)
  }
  await expect(page.locator('.desktop')).toBeVisible()
}
