import {describe,expect,it,vi} from 'vitest'
import {createUUID,type UUIDProvider} from '../src/services/uuid'

describe('browser UUID generation',()=>{
  it('uses native randomUUID when available',()=>{
    const native=vi.fn(()=> 'native-id')
    const provider={randomUUID:native,getRandomValues:vi.fn()} as unknown as UUIDProvider
    expect(createUUID(provider)).toBe('native-id')
    expect(native).toHaveBeenCalledOnce()
  })

  it('creates an RFC 4122 v4 UUID when randomUUID is unavailable',()=>{
    const provider={getRandomValues:vi.fn((bytes:Uint8Array)=>{bytes.fill(0);return bytes})} as unknown as UUIDProvider
    expect(createUUID(provider)).toBe('00000000-0000-4000-8000-000000000000')
    expect(provider.getRandomValues).toHaveBeenCalledOnce()
  })
})
