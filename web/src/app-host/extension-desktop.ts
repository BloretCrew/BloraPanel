import type {useDesktop} from '../desktop/store'
import type {Json} from './types'

export const desktopMethods=new Set(['window.move','window.close','shortcut.create'])

/** The calling view is supplied by the host, never by the extension payload. */
export function extensionDesktopAction(desktop:ReturnType<typeof useDesktop>,appId:string,viewTabId:string,capabilities:readonly string[],method:string,payload:Json):Json {
  if(!desktopMethods.has(method)||!capabilities.includes(method))throw new Error(`扩展未声明能力：${method}`)
  const view=desktop.state?.views[viewTabId]
  const source=Object.values(desktop.state?.windows||{}).find(w=>w.tabs.includes(viewTabId))
  if(!view||view.appId!==appId||!source||source.appId!==appId)throw new Error('扩展视图不存在或身份不匹配')
  const input=payload&&typeof payload==='object'&&!Array.isArray(payload)?payload:{}
  if(method==='window.close') {
    desktop.closeView(source.windowId,viewTabId)
    return {viewTabId,closed:true}
  }
  if(method==='window.move') {
    const target=input.targetWindowId
    if(target!==undefined&&(typeof target!=='string'||!target))throw new Error('目标窗口无效')
    if(typeof target==='string'&&desktop.state?.windows[target]?.appId!==appId)throw new Error('只能移入当前应用的已有窗口')
    desktop.moveView(viewTabId,target as string|undefined)
    const window=Object.values(desktop.state?.windows||{}).find(w=>w.tabs.includes(viewTabId))!
    return {viewTabId,windowId:window.windowId}
  }
  if(!view.resourceRef)throw new Error('当前标签没有资源，不能创建资源快捷入口')
  if(input.title!==undefined&&(typeof input.title!=='string'||!input.title.trim()||input.title.length>128))throw new Error('快捷入口名称无效')
  const title=typeof input.title==='string'?input.title.trim():view.title
  const shortcutId=desktop.addShortcut(appId,view.resourceRef,title)
  return {shortcutId}
}
