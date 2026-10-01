import type {Page} from '@playwright/test'

// Monaco chooses its bindings from the browser UA, not the test runner's OS.
// WebKit's default UA is Macintosh even on Windows/Linux Playwright hosts.
export async function pressEditorKey(page:Page,key:string){
  const mac=await page.evaluate(()=>/Macintosh|iPad|iPhone/.test(navigator.userAgent))
  const letter=key.toLowerCase()
  const binding=mac&&letter==='y'?'Meta+Shift+z':mac&&letter==='end'?'Meta+ArrowDown':mac&&letter==='home'?'Meta+ArrowUp':`${mac?'Meta':'Control'}+${key}`
  await page.keyboard.press(binding)
}
