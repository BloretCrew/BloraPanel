import type {Page} from '@playwright/test'

// Opt-in diagnostics, not an acceptance measurement. Aggregate only timing
// metadata: never persist trace arguments (URLs, account or terminal content).
export async function startRenderTrace(page:Page){
  const browser=page.context().browser()
  if(!browser||browser.browserType().name()!=='chromium')throw new Error('BLORA_PERF_RENDER_TRACE requires Chromium')
  const systemSession=await browser.newBrowserCDPSession()
  let compositorFeatures:Record<string,string>|undefined
  try{const system=await systemSession.send('SystemInfo.getInfo');compositorFeatures=system.gpu.featureStatus}
  finally{await systemSession.detach()}
  const session=await page.context().newCDPSession(page)
  const names=new Map<string,string>()
  const totals=new Map<string,{thread:string;event:string;count:number;totalMs:number;maxMs:number}>()
  let events=0,dropped=0,first=Infinity,last=0
  session.on('Tracing.dataCollected',({value})=>{
    for(const raw of value){
      const item=raw as unknown as {pid:number;tid:number;ph:string;name:string;ts:number;dur:number;args?:{name?:string}}
      const thread=`${item.pid}:${item.tid}`
      if(item.ph==='M'&&item.name==='thread_name')names.set(thread,String(item.args?.name||thread))
      if(item.ph!=='X'||!Number.isFinite(item.dur)||item.dur<0)continue
      events++
      first=Math.min(first,item.ts);last=Math.max(last,item.ts+item.dur)
      const key=`${thread}/${item.name}`
      if(!totals.has(key)&&totals.size>=2048){dropped++;continue}
      const row=totals.get(key)||{thread,event:String(item.name),count:0,totalMs:0,maxMs:0}
      row.count++;row.totalMs+=item.dur/1000;row.maxMs=Math.max(row.maxMs,item.dur/1000)
      totals.set(key,row)
    }
  })
  await session.send('Tracing.start',{categories:'devtools.timeline,cc,gpu,viz',options:'record-until-full',bufferUsageReportingInterval:1000})
  let bufferPeak=0
  session.on('Tracing.bufferUsage',event=>{bufferPeak=Math.max(bufferPeak,event.percentFull||event.value||0)})
  return async()=>{
    let timer:ReturnType<typeof setTimeout>|undefined
    try{
      const completed=new Promise<void>((resolve,reject)=>{
        session.once('Tracing.tracingComplete',()=>resolve())
        timer=setTimeout(()=>reject(new Error('Render trace did not complete within 30 seconds')),30000)
      })
      await session.send('Tracing.end')
      await completed
      return {diagnosticOnly:true,compositorFeatures,events,dropped,bufferPeak,spanMs:Number.isFinite(first)?(last-first)/1000:0,
        note:'Inclusive event durations overlap across nested events and threads; do not sum them as wall time.',
        timings:[...totals.values()].map(row=>({...row,thread:names.get(row.thread)||row.thread})).sort((a,b)=>b.totalMs-a.totalMs).slice(0,60)}
    }finally{clearTimeout(timer);await session.detach()}
  }
}
