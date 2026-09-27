import {describe,it,expect,vi,afterEach} from 'vitest'
import {publishNotification,notifications,dismissNotification} from '../src/services/notifications'
import type {useDesktop} from '../src/desktop/store'

describe('workspace notifications',()=>{
  afterEach(()=>vi.restoreAllMocks())
  function fixture(){
    const state={preferences:{notifications:[]},views:{view:{viewTabId:'view',appId:'example.app',resourceRef:{kind:'instance',id:'one'}}}}
    return {state,commit:(ops:any[])=>{for(const op of ops)(state.preferences as any)[op.path[1]]=op.value}} as unknown as ReturnType<typeof useDesktop>
  }
  it('binds source view, deduplicates keys and rate limits repeated updates',()=>{
    const desktop=fixture()
    vi.spyOn(Date,'now').mockReturnValue(1000)
    expect(()=>publishNotification(desktop,'other.app',{title:'x',message:''},'view')).toThrow('身份')
    for(let i=0;i<5;i++)publishNotification(desktop,'example.app',{title:'x',message:String(i),key:'same'},'view')
    expect(notifications(desktop)).toHaveLength(1)
    expect(notifications(desktop)[0]).toMatchObject({appId:'example.app',viewTabId:'view',resourceRef:{kind:'instance',id:'one'},message:'4'})
    expect(()=>publishNotification(desktop,'example.app',{title:'x',message:'',key:'same'},'view')).toThrow('频繁')
    dismissNotification(desktop,'example.app:same')
    expect(notifications(desktop)).toHaveLength(0)
  })
  it('bounds retained messages without opening or executing a target',()=>{
    const desktop=fixture()
    let time=0
    vi.spyOn(Date,'now').mockImplementation(()=>time+=1001)
    for(let i=0;i<110;i++)publishNotification(desktop,'example.app',{title:'x',message:'plain',key:String(i)},'view')
    expect(notifications(desktop)).toHaveLength(100)
    expect(()=>publishNotification(desktop,'example.app',{title:'x',message:'a'.repeat(4097)})).toThrow('限制')
  })
})
