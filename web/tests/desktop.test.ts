import 'fake-indexeddb/auto'
import {describe,it,expect,beforeEach,vi} from 'vitest'
import {createPinia,setActivePinia} from 'pinia'
import {defineComponent} from 'vue'
import {registerApp,apps} from '../src/app-host/registry'
import {useDesktop} from '../src/desktop/store'
import {RecoveryService} from '../src/recovery/service'
import {extensionDesktopAction} from '../src/app-host/extension-desktop'

beforeEach(()=>{setActivePinia(createPinia());apps.clear();registerApp({manifest:{appId:'test',title:'Reference',icon:'R',color:'#fff',packageVersion:'1.0.0',hostApiVersion:1,entrypoints:['overview','resource'],resourceHandlers:['instance'],permissions:[],dependencies:[],windowPolicy:'multiple',tabPolicy:{types:['overview','resource'],movable:true},stateSchemaVersion:1},component:defineComponent({}),captureState:s=>s,restoreState:s=>s,migrateState:s=>s,reconcileResource:async()=>{}})})
function setup(){const store=useDesktop(),storage={getItem:()=>null,setItem:()=>{},removeItem:()=>{},clear:()=>{},key:()=>null,length:0};const service=new RecoveryService('alice','device','tab',storage);store.recovery=service;store.state=service.state;return store}
describe('AppHost identity and atomic ownership',()=>{
  it('repeated focus avoids recovery writes but still raises and restores other windows',()=>{
    const store=setup()
    store.open({appId:'test'});const first=store.state!.order[0]!
    store.open({appId:'test',disposition:'new-window'});const second=store.state!.order.at(-1)!
    const writes=vi.spyOn(store.recovery!.storage,'setItem'),revision=store.state!.revision
    for(let i=0;i<100;i++)store.focus(second)
    expect(writes).not.toHaveBeenCalled()
    expect(store.state!.revision).toBe(revision)
    store.focus(first)
    expect(writes).toHaveBeenCalledTimes(1)
    expect(store.state!.order.at(-1)).toBe(first)
    expect(store.state!.activeWindowId).toBe(first)
    store.minimize(first);store.focus(first)
    expect(store.state!.windows[first]!.minimized).toBe(false)
    expect(store.state!.activeWindowId).toBe(first)
    // A restored state can have a stale active ID/order combination.
    store.state!.order=[first,second];store.focus(first)
    expect(store.state!.order.at(-1)).toBe(first)
  })
  it('scopes extension desktop actions to its own view and declared capabilities',()=>{
    const store=setup(),resource={kind:'instance',id:'one',nodeId:'node-a'}
    const view=store.open({appId:'test',resourceRef:resource,state:{note:'preserved'}})
    const source=store.state!.order[0]!
    const capabilities=['window.move','window.close','shortcut.create']
    expect(()=>extensionDesktopAction(store,'test',view,[],'window.close',null)).toThrow('未声明')
    expect(()=>extensionDesktopAction(store,'other',view,capabilities,'window.close',null)).toThrow('身份')
    expect(()=>extensionDesktopAction(store,'test',view,capabilities,'window.move',{targetWindowId:'missing'})).toThrow('已有窗口')
    extensionDesktopAction(store,'test',view,capabilities,'window.move',{})
    expect(store.state!.windows[source]).toBeUndefined()
    expect(store.state!.views[view]!.state.note).toBe('preserved')
    const result=extensionDesktopAction(store,'test',view,capabilities,'shortcut.create',{title:'Resource shortcut'}) as {shortcutId:string}
    expect(store.state!.shortcuts[result.shortcutId]!.resourceRef).toEqual(resource)
    extensionDesktopAction(store,'test',view,capabilities,'window.close',null)
    expect(store.state!.views[view]).toBeUndefined()
    expect(store.state!.closedViews[0]!.state.note).toBe('preserved')
    expect(store.state!.shortcuts[result.shortcutId]).toBeDefined()
  })
  it('opens and reuses dedicated resource windows separately from center views',()=>{
    const store=setup(),resourceRef={kind:'instance',id:'one',nodeId:'node-a'};const center=store.open({appId:'test'}),windowId=store.state!.order[0]!
    store.open({appId:'test',windowId,resourceRef,disposition:'new-tab'})
    const dedicated=store.open({appId:'test',resourceRef,disposition:'dedicated'})
    expect(store.state!.order).toHaveLength(2);expect(dedicated).not.toBe(center)
    expect(store.open({appId:'test',resourceRef,disposition:'dedicated'})).toBe(dedicated)
  })
  it('moves the same view with drafts/session identity and never clones resource operations',()=>{
    const store=setup(),view=store.open({appId:'test',resourceRef:{kind:'instance',id:'one'},state:{draftId:'draft',sessionId:'session'}})
    const source=store.state!.order[0]!,before=store.state!.revision
    store.moveView(view)
    expect(store.state!.revision).toBe(before+1);expect(store.state!.windows[source]).toBeUndefined();expect(Object.keys(store.state!.views)).toEqual([view]);expect(store.state!.views[view]!.state).toEqual({draftId:'draft',sessionId:'session'})
    const overview=store.open({appId:'test',disposition:'new-window'}),target=store.state!.order.at(-1)!
    store.moveView(view,target,0)
    expect(store.state!.windows[target]!.tabs).toEqual([view,overview]);expect(store.state!.windows[target]!.mode).toBe('center')
  })
  it('shortcut mutation and view closing only modify workspace identities',()=>{
    const store=setup(),ref={kind:'instance',id:'resource'};const view=store.open({appId:'test',resourceRef:ref,disposition:'dedicated'});store.addShortcut('test',ref,'Original');const shortcut=Object.keys(store.state!.shortcuts)[0]!
    store.commit([{kind:'set',path:['shortcuts',shortcut,'title'],value:'Alias'},{kind:'delete',path:['shortcuts',shortcut]}]);expect(store.state!.views[view]!.resourceRef).toEqual(ref)
    store.closeView(store.state!.order[0]!,view);expect(store.state!.closedViews[0]!.resourceRef).toEqual(ref)
  })
  it('restores the original closed view identity and focuses the next visible window on minimize',()=>{
    const store=setup(),first=store.open({appId:'test',state:{draftId:'kept'}}),one=store.state!.order[0]!
    store.open({appId:'test',disposition:'new-window'});const two=store.state!.order.at(-1)!
    store.minimize(two);expect(store.state!.activeWindowId).toBe(one)
    store.closeView(one,first);expect(store.state!.activeWindowId).toBeUndefined()
    store.reopenView();expect(store.state!.views[first]!.state.draftId).toBe('kept');expect(Object.keys(store.state!.views).filter(id=>id===first)).toHaveLength(1)
  })
  it('moves a tab into a minimized background window and raises its existing window',()=>{
    const store=setup(),first=store.open({appId:'test'}),one=store.state!.order[0]!
    store.open({appId:'test',disposition:'new-window'});const two=store.state!.order.at(-1)!;store.minimize(one)
    const third=store.open({appId:'test',windowId:two,disposition:'new-tab'});store.moveView(third,one)
    expect(store.state!.windows[one]!.tabs).toEqual([first,third]);expect(store.state!.windows[one]!.minimized).toBe(false);expect(store.state!.order.at(-1)).toBe(one)
  })
})
