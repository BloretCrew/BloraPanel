// A timer callback can already be queued when beforeunload cancels the current
// queries. Keep those subsequent reads behind the document lifecycle too.
let paused:Promise<void>|undefined
let release:(()=>void)|undefined
let generation=0

export function resumeTaskReads(){
  generation++
  const resume=release
  paused=undefined
  release=undefined
  resume?.()
}

export function pauseTaskReadsUntilPaint(){
  if(!paused)paused=new Promise<void>(resolve=>{release=resolve})
  const current=++generation
  // Accepted navigation destroys/hides this document and pagehide cancels its
  // waiting queries. A cancelled navigation can paint again and resume reads.
  // pageshow also releases the gate when the same document is restored.
  requestAnimationFrame(()=>requestAnimationFrame(()=>{if(current===generation)resumeTaskReads()}))
}

export async function readTaskQuery<T>(signal:AbortSignal,read:()=>Promise<T>):Promise<T>{
  signal.throwIfAborted()
  const wait=paused
  if(wait)await new Promise<void>((resolve,reject)=>{
    const abort=()=>{signal.removeEventListener('abort',abort);reject(signal.reason)}
    signal.addEventListener('abort',abort,{once:true})
    void wait.then(()=>{signal.removeEventListener('abort',abort);resolve()})
  })
  // Cancellation can win after the resume Promise settled but before this
  // continuation runs. Never call fetch with that obsolete signal.
  signal.throwIfAborted()
  return read()
}
