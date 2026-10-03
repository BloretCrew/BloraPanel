import {waitForDocumentReads} from './read-lifecycle'

export async function readTaskQuery<T>(signal:AbortSignal,read:()=>Promise<T>):Promise<T>{
  await waitForDocumentReads(signal)
  // Cancellation can win after the resume Promise settled but before this
  // continuation runs. Never call fetch with that obsolete signal.
  signal.throwIfAborted()
  return read()
}
