import {expect,type Locator,type Page} from '@playwright/test'

/** An opaque iframe can expose its DOM before its new/moved surface has
 * reached the parent's compositor. Wait for parent paint readiness, then use
 * the normal trusted mouse action; never dispatch a synthetic DOM click or
 * retry an operation that could have already reached the backend. */
export async function clickPaintedExtensionControl(page:Page,button:Locator) {
  await button.waitFor({state:'visible'})
  await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  // DOM visibility and RAF do not guarantee that the remote frame's input
  // region has arrived. Probe native hover without executing the action.
  await expect.poll(async()=>{
    await button.hover()
    return button.evaluate(element=>element.matches(':hover'))
  },{message:'The real mouse must reach the extension control before clicking',timeout:5000}).toBe(true)
  await button.click()
}

export const clickPaintedExtensionButton=clickPaintedExtensionControl
