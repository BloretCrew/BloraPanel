import {it,expect} from 'vitest'
import {decodeTerminalEvents} from '../src/services/terminal-events'
import {encodeJSON} from '../src/services/protocol'
it('preserves legacy events and grouped output, split UTF-8 bytes, resize and sequence order',()=>{
 const events=[{sequence:1,kind:'output',data:'5A=='},{sequence:2,kind:'output',data:'uK0='},{sequence:3,kind:'resize',cols:80,rows:24},{sequence:4,kind:'output',data:'G1szMm0='}]
 expect(decodeTerminalEvents(encodeJSON(events[0]))).toEqual([events[0]])
 expect(decodeTerminalEvents(encodeJSON({kind:'batch',events}))).toEqual(events)
})
it('rejects malformed or unbounded batches before native terminal parsing',()=>{
 for(const value of [null,[],{kind:'batch'},{kind:'batch',events:[]},{kind:'batch',events:[null]},{kind:'batch',events:Array.from({length:129},()=>({sequence:1,kind:'output',data:'YQ=='}))}])expect(()=>decodeTerminalEvents(encodeJSON(value))).toThrow()
})
