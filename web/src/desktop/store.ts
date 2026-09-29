import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'
import { apps, getApp, loadExternalSandboxApp, isSandboxApp } from '../app-host/registry'
import { id, resourceKey, type AppWindow, type OpenRequest, type Rect, type ResourceRef, type ViewTab, type Workspace,type Json } from '../app-host/types'
import { json, copy, type Mutation } from '../recovery/state'
import { RecoveryService, browserIdentity } from '../recovery/service'
import { api, APIError } from '../services/api'
import {DESKTOP_TOP, DESKTOP_BOTTOM, DESKTOP_GUTTER} from './work-area'

export interface CloudWorkspaceMeta { workspaceId:string; title:string; deviceId:string; revision:number; schemaVersion:number; includeContent:boolean; updatedAt:string }

export const useDesktop = defineStore('desktop', () => {
  const recovery = shallowRef<RecoveryService>()
  const state = shallowRef<Workspace>()
  const workspaces = ref<{slot:string;title:string}[]>([])
  const cloudWorkspaces = ref<CloudWorkspaceMeta[]>([])
  const activeWorkspace = ref('main')
  let catalogKey = '', startup = 0
  async function ensureExternalApps(service:RecoveryService){
    const ids=new Set<string>()
    for(const view of Object.values(service.state.views)) if(!apps.has(view.appId)) ids.add(view.appId)
    for(const shortcut of Object.values(service.state.shortcuts)) if(!apps.has(shortcut.appId)) ids.add(shortcut.appId)
    for(const appId of ids){
      try{await loadExternalSandboxApp(appId)}catch(e){service.fail(new Error(`扩展 ${appId} 无法加载，已保留原始现场：${String(e)}`))}
    }
  }
  async function restoreAppStates(service:RecoveryService){
    const changes:Mutation[]=[]
    for(const view of Object.values(service.state.views)){
      const app=apps.get(view.appId);if(!app)continue
      try{
        if (view.resourceRef) {
          const reconciled = await app.reconcileResource(view.resourceRef)
          if (reconciled && JSON.stringify(reconciled)!==JSON.stringify(view.resourceRef)) {
            changes.push({kind:'set',path:['views',view.viewTabId,'resourceRef'],value:json(reconciled)})
          }
        }
        const version=view.stateSchemaVersion||1
        // Third-party migrations execute inside their iframe before start.
        // Preserve the snapshot until its conditional migration commit arrives.
        if(version!==app.manifest.stateSchemaVersion && isSandboxApp(view.appId))continue
        const migrated=version===app.manifest.stateSchemaVersion?copy(view.state):app.migrateState(copy(view.state),version)
        const restored=app.restoreState(migrated)
        if(JSON.stringify(restored)!==JSON.stringify(view.state) || version!==app.manifest.stateSchemaVersion)changes.push({kind:'set',path:['views',view.viewTabId,'state'],value:json(restored)},{kind:'set',path:['views',view.viewTabId,'stateSchemaVersion'],value:app.manifest.stateSchemaVersion})
      }catch(e){
        // A deleted or revoked resource is still a valid local workspace:
        // retain its view and draft so the user can export or re-authorize it.
        // Other failures (corrupt state or storage) still surface as a
        // protection error instead of silently rewriting the snapshot.
        if (e instanceof APIError && (e.status===403 || e.status===404)) continue
        service.fail(new Error(`应用 ${view.appId} 的原始现场已保留：${String(e)}`))
      }
    }
    if(changes.length)service.commit(changes)
  }
  async function initialize(userId: string) {
    const generation = ++startup
    const identity = await browserIdentity()
    catalogKey = `blora:workspace:${userId}:${identity.deviceId}:${identity.browserTabId}`
    const copiedKey = identity.copiedFrom ? `blora:workspace:${userId}:${identity.deviceId}:${identity.copiedFrom}` : ''
    const catalog = localStorage.getItem(catalogKey) || (copiedKey && localStorage.getItem(copiedKey))
    workspaces.value = catalog ? JSON.parse(catalog) : [{slot:'main',title:'个人桌面'}]
    activeWorkspace.value = sessionStorage.getItem(catalogKey) || (copiedKey && sessionStorage.getItem(copiedKey)) || 'main'
    if (!workspaces.value.some(w=>w.slot===activeWorkspace.value)) activeWorkspace.value='main'
    localStorage.setItem(catalogKey,JSON.stringify(workspaces.value))
    const service = new RecoveryService(userId, identity.deviceId, identity.browserTabId, sessionStorage, 'blora-workspaces', activeWorkspace.value)
    await service.restore()
    if (identity.copiedFrom && !service.state.revision) {
      const source = new RecoveryService(userId, identity.deviceId, identity.copiedFrom, sessionStorage, 'blora-workspaces', activeWorkspace.value)
      await source.restore(false)
      const fork = copy(source.state)
      fork.browserTabId = identity.browserTabId
      fork.workspaceId = id('workspace')
      Object.assign(service.state, fork)
      await service.flush()
    }
    await ensureExternalApps(service)
    await restoreAppStates(service)
    if (generation === startup) { recovery.value = service; state.value = service.state }
  }
  async function switchWorkspace(slot:string) {
    if(slot===activeWorkspace.value || !workspaces.value.some(w=>w.slot===slot))return
    const previous = recovery.value!
    await previous.flush()
    const service = new RecoveryService(previous.userId,previous.deviceId,previous.browserTabId,sessionStorage,'blora-workspaces',slot)
    await service.restore()
    await ensureExternalApps(service)
    await restoreAppStates(service)
    sessionStorage.setItem(catalogKey,slot)
    activeWorkspace.value=slot;recovery.value=service;state.value=service.state
  }
  async function createWorkspace(title:string,duplicate=false) {
    const slot=id('workspace'), previous=recovery.value!
    const service=new RecoveryService(previous.userId,previous.deviceId,previous.browserTabId,sessionStorage,'blora-workspaces',slot)
    await service.restore()
    if(duplicate){Object.assign(service.state,copy(previous.state));service.state.workspaceId=slot;await service.flush()}
    workspaces.value.push({slot,title:title.trim()||'新工作区'})
    localStorage.setItem(catalogKey,JSON.stringify(workspaces.value))
    await switchWorkspace(slot)
  }
  async function refreshCloudWorkspaces() {
    const response = await api<{items:CloudWorkspaceMeta[]}>('/workspaces')
    cloudWorkspaces.value = response.items
    return response.items
  }
  async function syncCloud(includeContent=false, title?:string) {
    const service = recovery.value
    if (!service) throw new Error('工作区尚未初始化')
    await service.flush()
    const pendingRequestId = String(service.state.preferences.cloudRequestId || '')
    const requestId = pendingRequestId || crypto.randomUUID()
    if (!pendingRequestId) {
      service.commit([{kind:'set',path:['preferences','cloudRequestId'],value:json(requestId)}])
      await service.flush()
    }
    const snapshot = copy(service.state)
    // The request identity belongs to the transport attempt, not the cloud
    // workspace document. Never upload it as user state.
    delete snapshot.preferences.cloudRequestId
    if (!includeContent) {
      snapshot.drafts = {}
      snapshot.terminals = {}
      snapshot.uploads = {}
      delete snapshot.preferences.notifications
    }
    const baseRevision = Number(service.state.preferences.cloudRevision || 0)
    try {
      const result = await api<{metadata:CloudWorkspaceMeta}>(`/workspaces/${encodeURIComponent(service.state.workspaceId)}`, {method:'PUT',headers:{'Idempotency-Key':requestId}, body:JSON.stringify({title:title||workspaces.value.find(item=>item.slot===activeWorkspace.value)?.title||'工作区',deviceId:service.deviceId,baseRevision,schemaVersion:snapshot.schemaVersion,includeContent,state:snapshot})})
      service.commit([{kind:'set',path:['preferences','cloudRevision'],value:json(result.metadata.revision)},{kind:'set',path:['preferences','cloudIncludeContent'],value:json(includeContent)},{kind:'set',path:['preferences','cloudRequestId'],value:json(null)}])
      await refreshCloudWorkspaces()
      return result.metadata
    } catch (error) {
      // A changed local payload cannot be replayed with the old request key;
      // leave the local workspace intact and let the next explicit sync start
      // a fresh idempotent operation.
      if (error instanceof APIError && error.code === 'IDEMPOTENCY_CONFLICT') {
        service.commit([{kind:'set',path:['preferences','cloudRequestId'],value:json(null)}])
        await service.flush()
      }
      throw error
    }
  }
  async function openCloudWorkspace(remoteId:string, title?:string) {
    const response = await api<{metadata:CloudWorkspaceMeta;state:Workspace}>(`/workspaces/${encodeURIComponent(remoteId)}`)
    await createWorkspace(title || response.metadata.title, false)
    const service = recovery.value!
    const imported = copy(response.state)
    imported.userId = service.userId
    imported.deviceId = service.deviceId
    imported.browserTabId = service.browserTabId
    imported.workspaceId = response.metadata.workspaceId
    imported.preferences = {...(imported.preferences||{}), cloudRevision:response.metadata.revision, cloudIncludeContent:response.metadata.includeContent}
    imported.drafts ||= {}
    imported.terminals ||= {}
    imported.uploads ||= {}
    Object.assign(service.state, imported)
    await ensureExternalApps(service)
    await restoreAppStates(service)
    await service.flush()
    state.value = service.state
    return response.metadata
  }
  const commit = (mutations: Mutation[]) => recovery.value!.commit(mutations)
  function focus(windowId: string) {
    const s = state.value!
    if (!s.windows[windowId]) return
    // Clicking/dragging the already active window needs no new recovery entry.
    // Check all three facts so restored/minimized or out-of-order windows heal.
    if (s.activeWindowId === windowId && s.order.at(-1) === windowId && !s.windows[windowId].minimized) return
    commit([{ kind: 'set', path: ['order'], value: json([...s.order.filter(x => x !== windowId), windowId]) }, { kind: 'set', path: ['activeWindowId'], value: windowId }, { kind: 'set', path: ['windows', windowId, 'minimized'], value: false }])
  }
  function open(request: OpenRequest) {
    const s = state.value!, app = getApp(request.appId)
    const disposition = request.disposition || 'default'
    let window = request.windowId ? s.windows[request.windowId] : undefined
    if (window && window.appId !== request.appId) throw new Error('目标窗口不兼容')
    if (!window && disposition === 'dedicated') window = [...s.order].reverse().map(id => s.windows[id]!).find(w => w.appId === request.appId && w.mode === 'resource' && w.tabs.length === 1 && resourceKey(s.views[w.tabs[0]!]!.resourceRef) === resourceKey(request.resourceRef))
    if (!window && disposition === 'default') window = [...s.order].reverse().map(id => s.windows[id]!).find(w => w.appId === request.appId && w.mode === 'center')
    if (window && (disposition==='dedicated' || (disposition==='default' && !request.resourceRef && !request.windowId))) { focus(window.windowId); return window.activeTabId }
    if (window && request.resourceRef && disposition !== 'new-tab') {
      const existing = window.tabs.find(tabId => resourceKey(s.views[tabId]?.resourceRef) === resourceKey(request.resourceRef))
      if (existing) { activate(window.windowId, existing); return existing }
    }
    const view: ViewTab = { viewTabId: id('view'), appId: request.appId, type: request.entrypoint || (request.resourceRef ? 'resource' : 'overview'), title: request.title || app.manifest.title, resourceRef: request.resourceRef, state: app.restoreState(copy(request.state || {})),stateSchemaVersion:app.manifest.stateSchemaVersion }
    const mutations: Mutation[] = [{ kind: 'set', path: ['views', view.viewTabId], value: json(view) }]
    if (!window || disposition === 'new-window') {
      const offset = (s.order.length % 7) * 26
      const windowId = id('window')
      window = { windowId, appId: request.appId, mode: request.resourceRef ? 'resource' : 'center', rect: { x: 132 + offset, y: DESKTOP_TOP + 18 + offset, width: 960, height: 620 }, minimized: false, tabs: [view.viewTabId], activeTabId: view.viewTabId }
      mutations.push({ kind: 'set', path: ['windows', windowId], value: json(window) }, { kind: 'set', path: ['order'], value: json([...s.order, windowId]) })
    } else mutations.push({ kind: 'set', path: ['windows', window.windowId, 'tabs'], value: json([...window.tabs, view.viewTabId]) }, { kind: 'set', path: ['windows', window.windowId, 'activeTabId'], value: view.viewTabId }, { kind: 'set', path: ['windows', window.windowId, 'mode'], value: 'center' })
    mutations.push({kind:'set',path:['order'],value:json([...s.order.filter(id=>id!==window!.windowId),window.windowId])},{ kind: 'set', path: ['activeWindowId'], value: window.windowId })
    commit(mutations)
    return view.viewTabId
  }
  function activate(windowId: string, viewTabId: string) { commit([{ kind: 'set', path: ['windows', windowId, 'activeTabId'], value: viewTabId }]); focus(windowId) }
  function patchView(viewTabId: string, key: string, value: unknown) {
    const view=state.value?.views[viewTabId];if(!view)return
    const captured=getApp(view.appId).captureState({...copy(view.state),[key]:json(value)})
    commit([{kind:'set',path:['views',viewTabId,'state'],value:json(captured)}])
  }
  function geometry(windowId: string, rect: Rect) { commit([{ kind: 'set', path: ['windows', windowId, 'rect'], value: json(rect) }]) }
  function minimize(windowId: string) {
    const s=state.value!,next=[...s.order].reverse().find(id=>id!==windowId && !s.windows[id]!.minimized)
    const changes:Mutation[]=[{kind:'set',path:['windows',windowId,'minimized'],value:true}]
    if(s.activeWindowId===windowId)changes.push(next?{kind:'set',path:['activeWindowId'],value:next}:{kind:'delete',path:['activeWindowId']})
    commit(changes)
  }
  function snap(windowId: string, mode?: 'left'|'right'|'max') {
    const window = state.value!.windows[windowId]!
    const width = globalThis.innerWidth || 1280, height = globalThis.innerHeight || 800
    const half=(width-DESKTOP_GUTTER*3)/2
    const rect = mode ? { x: mode === 'right' ? half+DESKTOP_GUTTER*2 : DESKTOP_GUTTER, y: DESKTOP_TOP, width: mode === 'max' ? width-DESKTOP_GUTTER*2 : half, height: Math.max(250,height-DESKTOP_TOP-DESKTOP_BOTTOM) } : window.restoreRect || window.rect
    const changes: Mutation[] = [{ kind:'set', path:['windows',windowId,'rect'],value:json(rect) }]
    if (mode && !window.snap) changes.push({ kind:'set',path:['windows',windowId,'restoreRect'],value:json(window.rect) })
    changes.push(mode ? { kind:'set',path:['windows',windowId,'snap'],value:mode } : { kind:'delete',path:['windows',windowId,'snap'] })
    commit(changes)
  }
  function closeView(windowId: string, viewTabId: string) {
    const s = state.value!, window = s.windows[windowId]!, remaining = window.tabs.filter(id => id !== viewTabId)
    const changes: Mutation[] = [{ kind:'set',path:['closedViews'],value:json([...s.closedViews.slice(-19),s.views[viewTabId]!]) },{ kind:'delete',path:['views',viewTabId] }]
    if (!remaining.length) changes.push({ kind:'delete',path:['windows',windowId] },{ kind:'set',path:['order'],value:json(s.order.filter(id => id !== windowId)) })
    else changes.push({ kind:'set',path:['windows',windowId,'tabs'],value:json(remaining) },{ kind:'set',path:['windows',windowId,'activeTabId'],value:window.activeTabId === viewTabId ? remaining.at(-1)! : window.activeTabId })
    if(!remaining.length && s.activeWindowId===windowId){const next=[...s.order].reverse().find(id=>id!==windowId && !s.windows[id]!.minimized);changes.push(next?{kind:'set',path:['activeWindowId'],value:next}:{kind:'delete',path:['activeWindowId']})}
    commit(changes)
  }
  function reopenView(index=state.value!.closedViews.length-1) {
    const s=state.value!,view=s.closedViews[index];if(!view)return
    const windowId=id('window')
    const window:AppWindow={windowId,appId:view.appId,mode:view.resourceRef?'resource':'center',rect:{x:180,y:90,width:900,height:590},minimized:false,tabs:[view.viewTabId],activeTabId:view.viewTabId}
    commit([{kind:'set',path:['views',view.viewTabId],value:json(view)},{kind:'set',path:['windows',windowId],value:json(window)},{kind:'set',path:['order'],value:json([...s.order,windowId])},{kind:'set',path:['activeWindowId'],value:windowId},{kind:'set',path:['closedViews'],value:json(s.closedViews.filter((_,i)=>i!==index))}])
  }
  function closeWindow(windowId: string) { for (const view of [...state.value!.windows[windowId]!.tabs]) closeView(windowId, view) }
  function moveView(viewTabId: string, targetWindowId?: string, index?: number, point?: {x:number;y:number}) {
    const s = state.value!, source = Object.values(s.windows).find(w => w.tabs.includes(viewTabId))!, view = s.views[viewTabId]!
    let target = targetWindowId ? s.windows[targetWindowId] : undefined
    if (target && !getApp(target.appId).manifest.tabPolicy.types.includes(view.type)) throw new Error('应用不支持此标签类型')
    if (target && target.appId !== view.appId) throw new Error('此版本只允许同应用标签迁移；请使用打开方式创建新视图')
    const changes: Mutation[] = []
    if (!target) {
      const windowId = id('window')
      target = { windowId, appId:view.appId,mode:view.resourceRef?'resource':'center',rect:{...source.rect,x:point?.x??source.rect.x+36,y:point?.y??source.rect.y+36},minimized:false,tabs:[],activeTabId:viewTabId }
      changes.push({kind:'set',path:['windows',windowId],value:json(target)},{kind:'set',path:['order'],value:json([...s.order,windowId])})
    }
    const tabs = target.tabs.filter(id=>id!==viewTabId)
    if(target.windowId===source.windowId && index!==undefined && source.tabs.indexOf(viewTabId)<index)index--
    tabs.splice(index ?? tabs.length,0,viewTabId)
    changes.push({kind:'set',path:['windows',target.windowId,'tabs'],value:json(tabs)},{kind:'set',path:['windows',target.windowId,'activeTabId'],value:viewTabId})
    if(tabs.length>1) changes.push({kind:'set',path:['windows',target.windowId,'mode'],value:'center'})
    if (source.windowId !== target.windowId) {
      const remaining = source.tabs.filter(id=>id!==viewTabId)
      if(remaining.length) changes.push({kind:'set',path:['windows',source.windowId,'tabs'],value:json(remaining)},{kind:'set',path:['windows',source.windowId,'activeTabId'],value:source.activeTabId===viewTabId?remaining[0]!:source.activeTabId})
      else changes.push({kind:'delete',path:['windows',source.windowId]},{kind:'set',path:['order'],value:json([...s.order.filter(id=>id!==source.windowId && id!==target!.windowId),target.windowId])})
    }
    const order=s.order.filter(id=>id!==target!.windowId && !(id===source.windowId && source.windowId!==target!.windowId && source.tabs.length===1))
    changes.push({kind:'set',path:['order'],value:json([...order,target.windowId])},{kind:'set',path:['windows',target.windowId,'minimized'],value:false},{kind:'set',path:['activeWindowId'],value:target.windowId})
    commit(changes)
  }
  function addShortcut(appId:string,resourceRef:ResourceRef,title:string,page?:string,shortcutState?:Record<string,Json>) {
    const shortcutId=id('shortcut'), count=Object.keys(state.value!.shortcuts).length
    commit([{kind:'set',path:['shortcuts',shortcutId],value:json({shortcutId,appId,resourceRef,title,page,state:shortcutState,x:24+Math.floor(count/6)*100,y:110+(count%6)*100})}])
    return shortcutId
  }
  async function resetLayout() {
    const service = recovery.value
    if (!service || !state.value) throw new Error('工作现场尚未初始化')
    await service.flush()
    // Keep a complete copy before changing geometry. Drafts, terminal checkpoints,
    // uploads and resource identities remain in the backup workspace for recovery.
    const backupSlot = id('workspace')
    const backup = new RecoveryService(service.userId, service.deviceId, service.browserTabId, sessionStorage, 'blora-workspaces', backupSlot)
    await backup.restore()
    Object.assign(backup.state, copy(service.state))
    backup.state.workspaceId = backupSlot
    await backup.flush()
    const title = `布局备份 ${new Date().toLocaleString()}`
    workspaces.value = [...workspaces.value, {slot:backupSlot, title}]
    if (catalogKey && typeof localStorage !== 'undefined') localStorage.setItem(catalogKey, JSON.stringify(workspaces.value))

    const changes:Mutation[] = []
    state.value.order.forEach((windowId,index) => {
      if (!state.value!.windows[windowId]) return
      const offset = (index % 6) * 24
      changes.push(
        {kind:'set',path:['windows',windowId,'rect'],value:json({x:132+offset,y:DESKTOP_TOP+18+offset,width:960,height:620})},
        {kind:'set',path:['windows',windowId,'minimized'],value:false},
        {kind:'delete',path:['windows',windowId,'snap']},
        {kind:'delete',path:['windows',windowId,'restoreRect']},
      )
    })
    Object.values(state.value.shortcuts).forEach((shortcut,index) => changes.push({kind:'set',path:['shortcuts',shortcut.shortcutId,'x'],value:24+Math.floor(index/6)*100},{kind:'set',path:['shortcuts',shortcut.shortcutId,'y'],value:110+(index%6)*100}))
    if (changes.length) commit(changes)
    return {backupSlot, title}
  }
  return { state,recovery,workspaces,cloudWorkspaces,activeWorkspace,initialize,switchWorkspace,createWorkspace,refreshCloudWorkspaces,syncCloud,openCloudWorkspace,resetLayout,commit,open,focus,activate,patchView,geometry,minimize,snap,closeView,closeWindow,reopenView,moveView,addShortcut }
})
