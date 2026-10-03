// Read-only HTTP requests share the document lifecycle, including callbacks
// from independent application timers rather than only QueryObserver polls.
let paused:Promise<void>|undefined
let release:(()=>void)|undefined
let generation=0
let resumeDeadline:ReturnType<typeof setTimeout>|undefined
const active=new Set<AbortController>()
const departure=new DOMException('Document navigation','AbortError')

function pause(){if(!paused)paused=new Promise<void>(resolve=>{release=resolve})}
function cancelActive(){for(const controller of active)controller.abort(departure)}
function clearResumeDeadline(){clearTimeout(resumeDeadline);resumeDeadline=undefined}

export function resumeDocumentReads(){
  generation++
  clearResumeDeadline()
  const resume=release
  paused=undefined
  release=undefined
  resume?.()
}

export function pauseDocumentReadsUntilPaint(){
  pause()
  cancelActive()
  const current=++generation
  clearResumeDeadline()
  const resume=()=>{if(current===generation)resumeDocumentReads()}
  // A surviving document may stop painting even while its timers still run.
  // Keep a bounded departure guard, then allow safe reads without requiring
  // RAF. Pagehide cancels the deadline and invalidates both paths until
  // pageshow. This bounded guard is not a native navigation-cancel signal.
  resumeDeadline=setTimeout(resume,250)
  requestAnimationFrame(()=>requestAnimationFrame(resume))
}

export function hideDocumentReads(){
  generation++
  clearResumeDeadline()
  pause()
  cancelActive()
}

export async function waitForDocumentReads(signal?:AbortSignal|null){
  signal?.throwIfAborted()
  while(paused){
    const wait=paused
    await new Promise<void>((resolve,reject)=>{
      const abort=()=>{signal?.removeEventListener('abort',abort);reject(signal?.reason)}
      signal?.addEventListener('abort',abort,{once:true})
      void wait.then(()=>{signal?.removeEventListener('abort',abort);resolve()})
    })
    // A later navigation can pause again before this continuation executes.
    signal?.throwIfAborted()
  }
  signal?.throwIfAborted()
}

export async function readDocument<T>(signal:AbortSignal|null|undefined,read:(signal:AbortSignal)=>Promise<T>):Promise<T>{
  for(;;){
    await waitForDocumentReads(signal)
    if(paused)continue
    const controller=new AbortController()
    const cancel=()=>controller.abort(signal?.reason)
    signal?.addEventListener('abort',cancel,{once:true})
    active.add(controller)
    try{
      signal?.throwIfAborted()
      const result=await read(controller.signal)
      controller.signal.throwIfAborted()
      return result
    }catch(error){
      signal?.throwIfAborted()
      // Only our own navigation cancellation may restart a safe read in a
      // surviving document. Real network failures and owner cancellation keep
      // their existing errors. Mutations never use this retry path.
      if(controller.signal.reason!==departure)throw error
    }finally{
      active.delete(controller)
      signal?.removeEventListener('abort',cancel)
    }
  }
}
