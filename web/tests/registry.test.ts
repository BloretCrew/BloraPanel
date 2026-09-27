import {afterEach,describe, expect, it,vi} from 'vitest'
import {defineComponent} from 'vue'
import {apps,loadExternalSandboxApp,registerExternalApp} from '../src/app-host/registry'
import type {AppManifest} from '../src/app-host/types'

const component=defineComponent({template:'<div />'})
const manifest=(overrides:Partial<AppManifest>={}):AppManifest=>({
  appId:'example.registry', packageVersion:'1.0.0', hostApiVersion:1, title:'Registry', icon:'✦', color:'#fff',
  entrypoints:['overview'], resourceHandlers:['instance'], permissions:[], capabilities:['window.open'], dependencies:[],
  windowPolicy:'multiple', tabPolicy:{types:['overview'],movable:true}, stateSchemaVersion:1, ...overrides,
})

describe('external app registry boundary',()=>{
  afterEach(()=>vi.restoreAllMocks())
  it('refreshes same-schema sandbox registration and preserves it on invalid upgrades',async()=>{
    const id=`example.upgrade-${crypto.randomUUID().slice(0,8)}`
    const fetch=vi.spyOn(globalThis,'fetch')
    fetch.mockResolvedValue(new Response(JSON.stringify(manifest({appId:id}))))
    await loadExternalSandboxApp(id)
    const old=apps.get(id)!
    fetch.mockResolvedValue(new Response(JSON.stringify(manifest({appId:id,packageVersion:'2.0.0',capabilities:['resource.read']}))))
    await loadExternalSandboxApp(id)
    const upgraded=apps.get(id)!
    expect(upgraded.component).not.toBe(old.component)
    expect(upgraded.manifest.capabilities).toEqual(['resource.read'])
    fetch.mockResolvedValue(new Response(JSON.stringify(manifest({appId:'another.app'}))))
    await expect(loadExternalSandboxApp(id)).rejects.toThrow('标识与请求不一致')
    expect(apps.get(id)).toBe(upgraded)
    fetch.mockResolvedValue(new Response(JSON.stringify(manifest({appId:id,stateSchemaVersion:2}))))
    await loadExternalSandboxApp(id)
    expect(apps.get(id)!.manifest.stateSchemaVersion).toBe(2)
    expect(()=>apps.get(id)!.migrateState({note:'original'},1)).toThrow('未提供')
    apps.delete(id)
  })
  it('rejects malformed identity, capability and window policy before registration',()=>{
    expect(()=>registerExternalApp(manifest({appId:'Example.Bad'}),component)).toThrow('应用标识')
    expect(()=>registerExternalApp(manifest({appId:'blora.files'}),component)).toThrow('保留')
    expect(()=>registerExternalApp(manifest({capabilities:['host.exec']}),component)).toThrow('能力')
    expect(()=>registerExternalApp(manifest({tabPolicy:{types:['overview'],movable:false}}),component)).toThrow('窗口策略')
  })
  it('accepts a validated external manifest once and rejects duplicate app ids',()=>{
    const id=`example.registry-${crypto.randomUUID().slice(0,8)}`
    const value=manifest({appId:id})
    expect(()=>registerExternalApp(value,component)).not.toThrow()
    expect(()=>registerExternalApp(value,component)).toThrow('应用标识已注册')
  })
})
