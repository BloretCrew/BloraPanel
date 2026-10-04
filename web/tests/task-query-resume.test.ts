import {afterEach,expect,it,vi} from 'vitest'
import {QueryClient,QueryObserver} from '@tanstack/vue-query'
import {readTaskQuery,resumeTaskQueriesOnDocumentResume} from '../src/services/task-query-lifecycle'
import {hideDocumentReads,pauseDocumentReadsUntilPaint,resumeDocumentReads} from '../src/services/read-lifecycle'

afterEach(()=>{resumeDocumentReads();vi.useRealTimers();vi.unstubAllGlobals()})

it('revalidates cancelled active task queries on timer-only resume, retains cache and leaves hidden/disabled queries alone',async()=>{
  vi.useFakeTimers({toFake:['setTimeout','clearTimeout','setInterval','clearInterval']})
  vi.stubGlobal('requestAnimationFrame',()=>0)
  const client=new QueryClient({defaultOptions:{queries:{retry:false}}})
  let reads=0
  const observer=new QueryObserver(client,{queryKey:['tasks','summary'],queryFn:({signal})=>readTaskQuery(signal,async()=>{
    if(++reads===2)return new Promise<{active:number}>((_resolve,reject)=>signal.addEventListener('abort',()=>reject(signal.reason),{once:true}))
    return{active:reads===1?0:7}
  }),refetchInterval:3000})
  const disabledRead=vi.fn(async()=>({items:[]})),unrelatedRead=vi.fn(async()=>({items:[]}))
  const disabled=new QueryObserver(client,{queryKey:['tasks','disabled'],queryFn:disabledRead,enabled:false})
  const unrelated=new QueryObserver(client,{queryKey:['nodes'],queryFn:unrelatedRead})
  const unsubscribe=[observer.subscribe(()=>{}),disabled.subscribe(()=>{}),unrelated.subscribe(()=>{})]
  const stop=resumeTaskQueriesOnDocumentResume(client)
  try{
    await vi.advanceTimersByTimeAsync(0)
    expect(observer.getCurrentResult().data).toEqual({active:0})
    void observer.refetch()
    await vi.advanceTimersByTimeAsync(0)
    expect(reads).toBe(2)
    pauseDocumentReadsUntilPaint()
    await client.cancelQueries({queryKey:['tasks']})
    expect(observer.getCurrentResult().data).toEqual({active:0})
    await vi.advanceTimersByTimeAsync(249)
    expect(reads).toBe(2)
    await vi.advanceTimersByTimeAsync(1)
    expect(reads).toBe(3)
    expect(observer.getCurrentResult().data).toEqual({active:7})
    expect(disabledRead).not.toHaveBeenCalled()
    expect(unrelatedRead).toHaveBeenCalledOnce()
    resumeDocumentReads()
    await vi.advanceTimersByTimeAsync(0)
    expect(reads).toBe(3)
    pauseDocumentReadsUntilPaint();hideDocumentReads()
    await vi.advanceTimersByTimeAsync(1000)
    expect(reads).toBe(3)
    resumeDocumentReads()
    await vi.advanceTimersByTimeAsync(0)
    expect(reads).toBe(4)
    stop()
    pauseDocumentReadsUntilPaint()
    await vi.advanceTimersByTimeAsync(250)
    expect(reads).toBe(4)
  }finally{stop();unsubscribe.forEach(fn=>fn());client.clear()}
})
