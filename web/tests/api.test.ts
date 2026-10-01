import {afterEach,describe,expect,it,vi} from 'vitest'
import {api,APIError} from '../src/services/api'

describe('API transport failures',()=>{
  afterEach(()=>vi.restoreAllMocks())
  it.each(['Failed to fetch','Load failed','NetworkError when attempting to fetch resource.'])('keeps an accepted operation uncertain across browser error %s',async(message)=>{
    const fetch=vi.spyOn(globalThis,'fetch').mockRejectedValue(new TypeError(message))
    const failure=await api('/tasks',{method:'POST',headers:{'Idempotency-Key':'original-request'},body:'{}'}).catch(error=>error)
    expect(failure).toBeInstanceOf(APIError)
    if(!(failure instanceof APIError))throw failure
    expect(failure).toMatchObject({code:'NETWORK_ERROR',status:0})
    expect(failure.message).toContain('操作结果尚未确认')
    expect(fetch).toHaveBeenCalledOnce()
    expect(new Headers(fetch.mock.calls[0]![1]?.headers).get('Idempotency-Key')).toBe('original-request')
  })
  it('preserves caller cancellation rather than presenting it as a failed network operation',async()=>{
    const controller=new AbortController()
    const cancelled=new DOMException('Caller cancelled','AbortError')
    controller.abort(cancelled)
    const fetch=vi.spyOn(globalThis,'fetch').mockRejectedValue(cancelled)
    await expect(api('/tasks',{signal:controller.signal})).rejects.toBe(cancelled)
    expect(fetch).toHaveBeenCalledOnce()
  })
})
