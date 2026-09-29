import {it,expect} from 'vitest'
import {Terminal} from '@xterm/headless'
import {captureTerminalControlState,restoreTerminalControlState} from '../src/services/terminal-control-state'

it('rejects a corrupt second buffer before changing either buffer or charset',async()=>{
  const terminal=new Terminal({cols:80,rows:24})
  try{
    await new Promise<void>(resolve=>terminal.write('\x1b(0\x1b[2;5r\x1b[3g\x1b[1;4H\x1bH',resolve))
    const before=captureTerminalControlState(terminal,80,24)
    const bad=structuredClone(before)
    bad.normal.scrollTop=0;bad.alternate.tabs=[-1]
    expect(()=>restoreTerminalControlState(terminal,bad,80,24)).toThrow('数值无效')
    expect(captureTerminalControlState(terminal,80,24)).toEqual(before)
    expect(()=>restoreTerminalControlState(terminal,{...before,protocol:{...before.protocol,mouseEncoding:'custom'}},80,24)).toThrow('枚举无效')
    expect(()=>restoreTerminalControlState(terminal,{...before,protocol:{...before.protocol,cursorHidden:'false'}},80,24)).toThrow('布尔值无效')
    expect(captureTerminalControlState(terminal,80,24)).toEqual(before)
    const invalidCharset=structuredClone(before)
    invalidCharset.charsets=[JSON.parse('{"__proto__":"x"}')]
    expect(()=>restoreTerminalControlState(terminal,invalidCharset,80,24)).toThrow('映射无效')
    expect(captureTerminalControlState(terminal,80,24)).toEqual(before)
  }finally{terminal.dispose()}
})

it('captures an independent serializable charset and rejects unknown state versions',async()=>{
  const terminal=new Terminal({cols:80,rows:24,allowProposedApi:true})
  try{
    await new Promise<void>(resolve=>terminal.write('\x1b(0',resolve))
    const state=captureTerminalControlState(terminal,80,24)
    const stored=JSON.parse(JSON.stringify(state))
    await new Promise<void>(resolve=>terminal.write('\x1b(B',resolve))
    restoreTerminalControlState(terminal,stored,80,24)
    await new Promise<void>(resolve=>terminal.write('qq',resolve))
    expect(terminal.buffer.active.getLine(0)?.translateToString(true)).toBe('──')
    expect(()=>restoreTerminalControlState(terminal,{...stored,version:2},80,24)).toThrow('版本无效')
  }finally{terminal.dispose()}
})
