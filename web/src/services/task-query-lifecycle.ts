import type {QueryClient} from '@tanstack/vue-query'
import {onDocumentReadsResumed,waitForDocumentReads} from './read-lifecycle'

export function resumeTaskQueriesOnDocumentResume(query:QueryClient){
  // beforeunload cancels query-owned reads and retains their confirmed cache.
  // Once this document survives, revalidate active views immediately rather
  // than waiting up to another polling interval. Do not replace a queued read
  // that has already resumed, or refetch disabled/inactive query observers.
  return onDocumentReadsResumed(()=>{void query.refetchQueries({queryKey:['tasks'],type:'active'},{cancelRefetch:false})})
}

export async function readTaskQuery<T>(signal:AbortSignal,read:()=>Promise<T>):Promise<T>{
  await waitForDocumentReads(signal)
  // Cancellation can win after the resume Promise settled but before this
  // continuation runs. Never call fetch with that obsolete signal.
  signal.throwIfAborted()
  return read()
}
