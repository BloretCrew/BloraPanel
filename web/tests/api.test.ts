import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest'
import {api,APIError,session} from '../src/services/api'
import {hideDocumentReads,pauseDocumentReadsUntilPaint,resumeDocumentReads} from '../src/services/read-lifecycle'

describe('API transport failures',()=>{
  afterEach(()=>{vi.restoreAllMocks();vi.unstubAllGlobals()})
  it('submits login when the browser exposes getRandomValues but not randomUUID',async()=>{
    vi.stubGlobal('crypto',{getRandomValues:(bytes:Uint8Array)=>{bytes.fill(0);return bytes}})
    const fetch=vi.spyOn(globalThis,'fetch').mockResolvedValue(Response.json({user:{userId:'admin'},csrfToken:'csrf'}))
    await api('/login',{method:'POST',body:JSON.stringify({name:'admin',password:'example'})})
    expect(fetch).toHaveBeenCalledOnce()
    expect(new Headers(fetch.mock.calls[0]![1]?.headers).get('Idempotency-Key')).toBe('00000000-0000-4000-8000-000000000000')
  })
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
    expect(fetch).not.toHaveBeenCalled()
  })
})

describe('read-only document lifecycle',()=>{
  let frames:FrameRequestCallback[]=[]
  const paint=()=>{for(let i=0;i<2;i++){const queued=frames.splice(0);queued.forEach(callback=>callback(0))}}
  beforeEach(()=>{
    frames=[]
    resumeDocumentReads()
    vi.stubGlobal('requestAnimationFrame',(callback:FrameRequestCallback)=>{frames.push(callback);return frames.length})
  })
  afterEach(()=>{resumeDocumentReads();vi.useRealTimers();vi.restoreAllMocks();vi.unstubAllGlobals();session.user=undefined})

  it.each([{}, {method:'POST',readOnly:true,body:'{}'}])('cancels a pending safe read and defers both its restart and queued reads until paint (%j)',async(options)=>{
    let originalSignal:AbortSignal|undefined
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementationOnce((_input,init)=>new Promise((_resolve,reject)=>{
      originalSignal=init!.signal!
      originalSignal.addEventListener('abort',()=>reject(originalSignal!.reason),{once:true})
    })).mockImplementation(async()=>Response.json({items:['confirmed']}))
    const pending=api('/instances/fixture/logs',options)
    await vi.waitFor(()=>expect(fetch).toHaveBeenCalledOnce())
    pauseDocumentReadsUntilPaint()
    expect(originalSignal?.aborted).toBe(true)
    const queued=api('/nodes/fixture/metrics')
    await Promise.resolve();await Promise.resolve()
    expect(fetch).toHaveBeenCalledOnce()
    paint()
    await expect(pending).resolves.toEqual({items:['confirmed']})
    await expect(queued).resolves.toEqual({items:['confirmed']})
    expect(fetch).toHaveBeenCalledTimes(3)
  })

  it('does not publish an empty success when navigation aborts response-body parsing',async()=>{
    let reading=false
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementationOnce(async(_input,init)=>({
      ok:true,status:200,json:()=>new Promise((_resolve,reject)=>{
        reading=true
        init!.signal!.addEventListener('abort',()=>reject(init!.signal!.reason),{once:true})
      }),
    } as Response)).mockImplementation(async()=>Response.json({items:['complete']}))
    const pending=api('/instances/fixture/logs')
    await vi.waitFor(()=>expect(reading).toBe(true))
    pauseDocumentReadsUntilPaint()
    paint()
    await expect(pending).resolves.toEqual({items:['complete']})
    expect(fetch).toHaveBeenCalledTimes(2)
  })

  it('keeps pagehide paused despite older frames and cancels an owner waiting for pageshow',async()=>{
    vi.useFakeTimers({toFake:['setTimeout','clearTimeout']})
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async()=>Response.json({items:[]}))
    pauseDocumentReadsUntilPaint()
    hideDocumentReads()
    const owner=new AbortController()
    const pending=api('/instances/fixture/logs',{signal:owner.signal})
    const cancelled=new DOMException('View closed','AbortError')
    const assertion=expect(pending).rejects.toBe(cancelled)
    paint()
    await vi.advanceTimersByTimeAsync(1000)
    expect(fetch).not.toHaveBeenCalled()
    owner.abort(cancelled)
    await assertion
    resumeDocumentReads()
    await expect(api('/nodes')).resolves.toEqual({items:[]})
    expect(fetch).toHaveBeenCalledOnce()
  })

  it('resumes cancelled and queued safe reads when animation frames stop in a surviving document',async()=>{
    vi.useFakeTimers({toFake:['setTimeout','clearTimeout']})
    let originalSignal:AbortSignal|undefined
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementationOnce((_input,init)=>new Promise((_resolve,reject)=>{
      originalSignal=init!.signal!
      originalSignal.addEventListener('abort',()=>reject(originalSignal!.reason),{once:true})
    })).mockImplementation(async()=>Response.json({items:['confirmed']}))
    const pending=api('/instances/fixture/logs')
    await vi.advanceTimersByTimeAsync(0)
    expect(fetch).toHaveBeenCalledOnce()
    pauseDocumentReadsUntilPaint()
    expect(originalSignal?.aborted).toBe(true)
    const queued=api('/nodes/fixture/metrics')
    await vi.advanceTimersByTimeAsync(249)
    expect(fetch).toHaveBeenCalledOnce()
    // No frame callback is run. Timers alone must restore read liveness.
    await vi.advanceTimersByTimeAsync(1)
    await expect(pending).resolves.toEqual({items:['confirmed']})
    await expect(queued).resolves.toEqual({items:['confirmed']})
    expect(fetch).toHaveBeenCalledTimes(3)
    paint()
    await vi.advanceTimersByTimeAsync(1000)
    expect(fetch).toHaveBeenCalledTimes(3)
  })

  it('does not let an older resume deadline release reads during a newer navigation',async()=>{
    vi.useFakeTimers({toFake:['setTimeout','clearTimeout']})
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementation(async()=>Response.json({items:[]}))
    pauseDocumentReadsUntilPaint()
    const pending=api('/nodes')
    await vi.advanceTimersByTimeAsync(100)
    pauseDocumentReadsUntilPaint()
    await vi.advanceTimersByTimeAsync(150)
    expect(fetch).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(100)
    await expect(pending).resolves.toEqual({items:[]})
    expect(fetch).toHaveBeenCalledOnce()
  })

  it('only repeats explicitly read-only POST queries; an interrupted mutation keeps its original identity and uncertainty',async()=>{
    let mutationSignal:AbortSignal|null|undefined
    const fetch=vi.spyOn(globalThis,'fetch').mockImplementationOnce((_input,init)=>new Promise((_resolve,reject)=>{
      mutationSignal=init?.signal
      pauseDocumentReadsUntilPaint()
      reject(new TypeError('Connection interrupted'))
    })).mockImplementation(async()=>Response.json({items:[]}))
    const mutation=await api('/terminals',{method:'POST',body:'{}',headers:{'Idempotency-Key':'original-terminal'}}).catch(error=>error)
    expect(mutation).toMatchObject({code:'NETWORK_ERROR',status:0})
    expect(mutationSignal).toBeUndefined()
    expect(new Headers(fetch.mock.calls[0]![1]!.headers).get('Idempotency-Key')).toBe('original-terminal')
    const read=api('/nodes/fixture/docker/query',{method:'POST',readOnly:true,body:'{}'})
    await Promise.resolve()
    expect(fetch).toHaveBeenCalledOnce()
    paint()
    await expect(read).resolves.toEqual({items:[]})
    expect(fetch).toHaveBeenCalledTimes(2)
    expect(fetch.mock.calls[1]![1]).not.toHaveProperty('readOnly')
  })

  it('rejects a deferred read after the account changes instead of fetching for its old view',async()=>{
    session.user={userId:'original',name:'original',admin:false,disabled:false,revision:1}
    const fetch=vi.spyOn(globalThis,'fetch')
    pauseDocumentReadsUntilPaint()
    const pending=api('/instances')
    const assertion=expect(pending).rejects.toMatchObject({name:'AbortError',message:'Account changed'})
    session.user={...session.user,userId:'replacement'}
    paint()
    await assertion
    expect(fetch).not.toHaveBeenCalled()
  })

  it('does not retry real read failures when there has been no navigation cancellation',async()=>{
    const fetch=vi.spyOn(globalThis,'fetch').mockRejectedValue(new TypeError('Load failed'))
    await expect(api('/instances/fixture/logs')).rejects.toMatchObject({code:'NETWORK_ERROR',status:0})
    expect(fetch).toHaveBeenCalledOnce()
  })
})
