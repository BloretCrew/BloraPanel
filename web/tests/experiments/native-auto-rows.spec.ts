import {test,expect,type Page} from '@playwright/test'
import {nativeDeviceScale} from '../../playwright-browser'
import {installNativeAutoRows} from '../helpers/native-auto-rows'

async function settled(page:Page){await page.evaluate(()=>new Promise<void>(resolve=>requestAnimationFrame(()=>requestAnimationFrame(()=>resolve()))))}
async function stable(page:Page){let before=await page.screenshot({caret:'hide'});for(let n=0;n<8;n++){await settled(page);const next=await page.screenshot({caret:'hide'});if(next.equals(before))return next;before=next}throw Error('Native terminal did not settle')}
for(const scale of [1,1.25])test.describe(`native automatic row relevance at DPR ${scale}`,()=>{
 test.use(nativeDeviceScale(scale))
 test('native clipped rows retain glyphs, geometry, selection and synchronized output on reveal',async({page},info)=>{
  await page.route('**/api/v1/**',route=>route.fulfill({json:{items:[]}}));await page.goto('/')
  await page.evaluate(async()=>{
   const {Terminal}=await import('/node_modules/.vite/deps/@xterm_xterm.js' as string);await import('/src/services/terminals.ts' as string)
   const viewport=document.createElement('div');viewport.className='terminal-paint-viewport';viewport.style.cssText='position:fixed;left:80px;top:140px;width:860px;height:540px;overflow:hidden;z-index:10000;background:#111418'
   const host=document.createElement('div');host.className='terminal-container';host.dataset.terminalRenderer='default';host.dataset.paintPartial='true';host.style.cssText='position:absolute;left:0;top:0;width:860px;height:540px';viewport.append(host);document.body.append(viewport)
   const terminal=new Terminal({cols:80,rows:24,fontSize:14,fontFamily:'"Liberation Mono",monospace',allowProposedApi:true,cursorBlink:false,theme:{background:'#111418',foreground:'#e6ebf0'}});terminal.open(host);await document.fonts.ready
   ;(window as any).__rowRelevance={terminal,host,viewport,rows:[...host.querySelectorAll('.xterm-rows>div')]}
  })
  await installNativeAutoRows(page)
  const write=async(data:string)=>{await page.evaluate(data=>new Promise<void>(resolve=>(window as any).__rowRelevance.terminal.write(data,resolve)),data);await settled(page)}
  const compare=async(label:string)=>{
   const geometry=()=>page.evaluate(()=>{const f=(window as any).__rowRelevance;return [...f.host.querySelectorAll('.xterm-rows>div')].map((row:any)=>{const r=row.getBoundingClientRect();return {x:r.x,y:r.y,width:r.width,height:r.height}})})
   // Keep xterm's original scrollbar fully hovered for both references.
   // Its upstream 800ms fade otherwise changes pixels even in the explicitly
   // disabled control, independently of row containment or terminal output.
   await page.mouse.move(928,180)
   const originalNodes=await page.evaluate(()=>{const f=(window as any).__rowRelevance;f.referenceNodes=[...f.host.querySelectorAll('.xterm-rows>div')];return f.referenceNodes.length})
   await page.evaluate(()=>(window as any).__nativeAutoRows.setVisible(false));await settled(page);const original=await stable(page),originalGeometry=await geometry()
   await page.evaluate(()=>(window as any).__nativeAutoRows.setVisible(true));await settled(page);const candidate=await stable(page)
   const diff=await page.evaluate(async({a,b})=>{const load=async(value:string)=>{const image=new Image();image.src='data:image/png;base64,'+value;await image.decode();const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;const ctx=canvas.getContext('2d')!;ctx.drawImage(image,0,0);return ctx.getImageData(0,0,canvas.width,canvas.height).data};const p=await load(a),q=await load(b);let max=0,changed=0;for(let i=0;i<p.length;i++){const d=Math.abs(p[i]!-q[i]!);max=Math.max(max,d);if(d)changed++}return {max,changed}},{a:original.toString('base64'),b:candidate.toString('base64')})
   console.log('NATIVE_ROW_RELEVANCE',label,JSON.stringify(diff));await info.attach(label+'-original.png',{body:original,contentType:'image/png'});await info.attach(label+'-candidate.png',{body:candidate,contentType:'image/png'});expect(diff).toEqual({max:0,changed:0})
   expect(await geometry()).toEqual(originalGeometry)
   expect(await page.evaluate(()=>{const f=(window as any).__rowRelevance,rows=[...f.host.querySelectorAll('.xterm-rows>div')];return f.referenceNodes.filter((node:Element,i:number)=>node.isConnected&&node===rows[i]).length})).toBe(originalNodes)
  }
  await write('\x1b[?25l'+Array.from({length:24},(_,i)=>`\x1b[${31+i%7}m${i} 中文 e\u0301 \x1b[1;3mItalic-overhang\x1b[0m ${'ABCDEFGH'.repeat(4)}`).join('\r\n'))
  await page.evaluate(()=>{const f=(window as any).__rowRelevance;f.viewport.style.height='80px'});await compare('top-strip')
  const relevance=await page.evaluate(()=>(window as any).__nativeAutoRows.snapshot())
  if(scale===1){expect(relevance.active).toBe(1);expect(relevance.automatic).toBe(24);expect(relevance.skipped).toBeGreaterThan(0);expect(relevance.skipped).toBeLessThan(24)}else expect(relevance.automatic).toBe(0)
  await page.evaluate(()=>{const f=(window as any).__rowRelevance;f.host.style.top='-180px';f.terminal.select(2,3,700)});await compare('middle-selection')
  await write('\r\n'+Array.from({length:45},(_,i)=>'scroll-history-'+i).join('\r\n'));await page.evaluate(()=>(window as any).__rowRelevance.terminal.scrollToLine(0));await compare('scroll')
  await page.evaluate(()=>{const f=(window as any).__rowRelevance;f.terminal.clearSelection();f.terminal.scrollToBottom();f.terminal.resize(90,25);f.terminal.options.theme={background:'#f9fafc',foreground:'#192129'};f.host.style.top='-210px'});await write('\x1b[2J\x1b[H\x1b[4mRESIZED 中文 e\u0301\x1b[0m');await compare('resized-light')
  await write('\x1b[?2026h\x1b[HHELD ATOMIC OUTPUT');await page.evaluate(()=>{const f=(window as any).__rowRelevance;f.viewport.style.height='540px';f.host.style.top='0px'});await settled(page)
  await expect(page.locator('.xterm-rows')).not.toContainText('HELD ATOMIC OUTPUT');await write('\x1b[?2026l');await compare('released-atomic');await expect(page.locator('.xterm-rows')).toContainText('HELD ATOMIC OUTPUT')
  await page.evaluate(()=>{(window as any).__nativeAutoRows.dispose();const f=(window as any).__rowRelevance;f.terminal.dispose();f.viewport.remove()})
 })
})
