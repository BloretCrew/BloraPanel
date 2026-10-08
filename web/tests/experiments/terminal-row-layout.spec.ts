import {test,expect} from '@playwright/test'
import {nativeDeviceScale} from '../../playwright-browser'

// Native CSS layout boundaries only: no paint clipping, alternate renderer,
// captured app image, skipped native row or delayed output is installed.
for(const scale of [1,1.25])test.describe(`native terminal row layout at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('layout containment retains native glyphs, selection, scroll, resize and atomic output exactly',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  await page.evaluate(async()=>{
   const {Terminal}=await import('/node_modules/.vite/deps/@xterm_xterm.js' as string)
   await import('/src/services/terminals.ts' as string)
   const host=document.createElement('div');host.dataset.rowLayoutHost='true';host.style.cssText='position:fixed;left:80px;top:140px;width:1000px;height:600px;z-index:10000;background:#111418';document.body.append(host)
   const terminal=new Terminal({cols:80,rows:24,fontSize:14,fontFamily:'"Liberation Mono",monospace',allowProposedApi:true,cursorBlink:false,theme:{background:'#111418',foreground:'#e6ebf0'}});terminal.open(host);await document.fonts.ready
   ;(window as any).__rowLayout={terminal,host}
   const style=document.createElement('style');style.textContent='.terminal-row-layout-experiment .xterm-screen:not(:has(.xterm-selection>div)) .xterm-rows>div{contain:layout style}';document.head.append(style)
  })
  if(process.env.BLORA_E08_ROW_LAYOUT_CONTROL==='1')await page.addStyleTag({content:'.terminal-row-layout-experiment .xterm-screen .xterm-rows>div{contain:none!important}'})
  const host=page.locator('[data-row-layout-host="true"]')
  const settle=()=>page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))
  const write=async(data:string)=>{await page.evaluate(data=>new Promise<void>(resolve=>(window as any).__rowLayout.terminal.write(data,resolve)),data);await settle()}
  const compare=async(label:string,baseline?:Buffer)=>{
   let original=baseline
   if(!original){await host.evaluate(element=>element.classList.remove('terminal-row-layout-experiment'));await settle();original=await host.screenshot()}
   await host.evaluate(element=>element.classList.add('terminal-row-layout-experiment'));await settle()
   if(label==='selection')expect(await host.locator('.xterm-rows>div').first().evaluate(element=>getComputedStyle(element).contain)).toBe('none')
   if(['coloured-glyphs','resized-light','released-atomic'].includes(label))expect(await host.locator('.xterm-rows>div').first().evaluate(element=>getComputedStyle(element).contain)).toBe(process.env.BLORA_E08_ROW_LAYOUT_CONTROL==='1'?'none':'layout style')
   const contained=await host.screenshot()
   const diff=await page.evaluate(async({a,b})=>{
    const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data}
    const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}
   },{a:original.toString('base64'),b:contained.toString('base64')})
   console.log('ROW_LAYOUT_PIXELS',label,JSON.stringify({...diff,...await page.evaluate(()=>({sync:(window as any).__rowLayout.terminal._core.coreService.decPrivateModes.synchronizedOutput}))}))
   await info.attach(label+'-original.png',{body:original,contentType:'image/png'});await info.attach(label+'-contained.png',{body:contained,contentType:'image/png'});expect(diff).toEqual({max:0,changed:0});return contained
  }
  await write('\x1b[?25l'+Array.from({length:24},(_,i)=>`\x1b[${31+i%7}m${i} 中文 e\u0301 \x1b[1;3mItalic-overhang\x1b[0m ${'ABCDEFGH'.repeat(4)}`).join('\r\n'));await compare('coloured-glyphs')
  await page.evaluate(()=>(window as any).__rowLayout.terminal.select(2,3,170));await settle();await compare('selection')
  await write('\r\n'+Array.from({length:45},(_,i)=>'scroll-history-'+i).join('\r\n'));await page.evaluate(()=>(window as any).__rowLayout.terminal.scrollToLine(0));await settle();await compare('scroll')
  await page.evaluate(()=>{const t=(window as any).__rowLayout.terminal;t.resize(90,25);t.clearSelection();t.scrollToBottom();t.options.theme={background:'#f9fafc',foreground:'#192129'}});await write('\x1b[2J\x1b[H\x1b[4mRESIZED 中文 e\u0301\x1b[0m');const committed=await compare('resized-light')
  // Compare against the committed native image taken before DECSET 2026.
  // Two extra screenshot encodes can exceed xterm's unchanged 1000ms timeout.
  await write('\x1b[?2026h\x1b[HHELD ATOMIC OUTPUT');await compare('held-atomic',committed)
  expect(await page.evaluate(()=>(window as any).__rowLayout.terminal._core.coreService.decPrivateModes.synchronizedOutput)).toBe(true)
  await expect(host.locator('.xterm-rows')).not.toContainText('HELD ATOMIC OUTPUT')
  await write('\x1b[?2026l');await compare('released-atomic');await expect(host.locator('.xterm-rows')).toContainText('HELD ATOMIC OUTPUT')
  await page.evaluate(()=>{const f=(window as any).__rowLayout;f.terminal.dispose();f.host.remove()})
 })
})
