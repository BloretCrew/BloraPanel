import type {Locator,Page} from '@playwright/test'

/** Choose an option from Blora’s custom combobox without invoking browser-native select UI. */
export async function selectStyledOption(page:Page,combobox:Locator,value:string){
  await combobox.click()
  const listId=await combobox.first().getAttribute('aria-controls')
  if(!listId)throw new Error('Custom combobox is missing its listbox id')
  const escaped=value.replaceAll('\\','\\\\').replaceAll('"','\\"')
  await page.locator(`[id="${listId}"]`).last().locator(`[role="option"][data-value="${escaped}"]`).click()
}
