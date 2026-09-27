import {describe,it,expect} from 'vitest'
import {appendLog,emptyLog,logGap,logText,LOG_TEXT_LIMIT,LOG_LINE_LIMIT} from '../src/services/logs'
const event=(sequence:number,bytes:Uint8Array)=>({sequence,kind:'output',data:Buffer.from(bytes).toString('base64')})
describe('resumable bounded log decoding',()=>{
  it('persists a partial UTF-8 suffix at the same event cursor',()=>{const bytes=new TextEncoder().encode('中文🙂'),first=appendLog(emptyLog('run'),event(1,bytes.slice(0,5))),restored=JSON.parse(JSON.stringify(first)),done=appendLog(restored,event(2,bytes.slice(5)));expect(first.text).toBe('中');expect(first.carry).toHaveLength(2);expect(done.text).toBe('中文🙂');expect(done.carry).toEqual([])})
  it('keeps a bounded tail including many short lines and long single lines',()=>{let record=appendLog(emptyLog('run'),event(1,new TextEncoder().encode('x\n'.repeat(10000))));expect(record.text.split('\n').length).toBeLessThanOrEqual(LOG_LINE_LIMIT);expect(record.discardedLines).toBeGreaterThan(0);record=appendLog(record,event(2,new TextEncoder().encode('a'.repeat(100000))));expect(record.text.length).toBeLessThanOrEqual(LOG_TEXT_LIMIT);expect(record.truncated).toBe(true)})
  it('requires explicit gap notification before jumping event cursors',()=>{const record=emptyLog('run');expect(()=>appendLog(record,event(10,new TextEncoder().encode('late')))).toThrow('不连续');expect(appendLog(logGap(record,10),event(10,new TextEncoder().encode('late'))).text).toContain('late')})
  it('shows terminal control output as text without markup or raw escape codes',()=>{expect(logText('\u001b[32m中文\u001b[0m\r\nold\rnew')).toBe('中文\nnew')})
})
