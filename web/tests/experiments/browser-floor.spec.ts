import {test,expect} from '@playwright/test'

// Scheduling attribution only: no application workload and never E08
// acceptance. Use the original trusted pointer trajectory and double RAF,
// default browser launch settings, and a separate fresh page per sample.
for(const mode of ['empty','native-flat-frame'] as const)for(let repeat=1;repeat<=3;repeat++){
 test(`native scheduling floor ${mode} repeat ${repeat}`,async({page,browserName})=>{
  await page.goto('about:blank')
  await page.setContent('<style>html,body{margin:0;height:100%;background:#eaf0f6}#frame{position:absolute;width:900px;height:620px;left:100px;top:100px;background:#fff}</style>'+(mode==='native-flat-frame'?'<div id="frame"></div>':''))
  await page.evaluate((mode)=>{
   const latencies:number[]=[],frames:number[]=[],phases:{input:number;first:number;second:number}[]=[]
   let running=true,last=performance.now(),trusted=0
   ;(window as any).__nativeFloor={latencies,frames,phases,stop(){running=false;return {latencies,frames,phases,trusted}}}
   window.addEventListener('pointermove',event=>{
    if(event.buttons!==1||!running)return
    if(event.isTrusted)trusted++
    const started=event.timeStamp,handled=performance.now()
    if(mode==='native-flat-frame')document.getElementById('frame')!.style.transform=`translate(${event.clientX-340}px,${event.clientY-115}px)`
    requestAnimationFrame(()=>{const first=performance.now();requestAnimationFrame(()=>{const finished=performance.now();latencies.push(finished-started);phases.push({input:handled-started,first:first-handled,second:finished-first})})})
   },true)
   const frame=(now:number)=>{if(!running)return;frames.push(now-last);last=now;requestAnimationFrame(frame)}
   requestAnimationFrame(frame)
  },mode)
  await page.mouse.move(340,115);await page.mouse.down()
  for(let i=0;i<120;i++)await page.mouse.move(340+Math.sin(i/10)*80,115+Math.cos(i/10)*30)
  await page.mouse.up()
  await page.evaluate(()=>(window as any).__nativeFloor.stop())
  // Let already requested callbacks finish. This does not alter timestamps
  // or add samples; idle frames are excluded from the frame distribution.
  await page.waitForTimeout(250)
  const result=await page.evaluate(()=>{
   const probe=(window as any).__nativeFloor.stop()
   const p95=(items:number[])=>items.sort((a,b)=>a-b)[Math.ceil(items.length*.95)-1]||0
   return {samples:probe.latencies.length,trusted:probe.trusted,p95:p95(probe.latencies),max:Math.max(...probe.latencies),frameP95:p95(probe.frames),inputP95:p95(probe.phases.map((p:any)=>p.input)),firstP95:p95(probe.phases.map((p:any)=>p.first)),secondP95:p95(probe.phases.map((p:any)=>p.second))}
  })
  console.log('NATIVE_SCHEDULING_FLOOR',JSON.stringify({browser:browserName,mode,repeat,...result}))
  expect(result.samples).toBe(120);expect(result.trusted).toBe(120)
 })
}
