import {describe,it,expect} from 'vitest'
import {pendingUtf8} from '../src/services/terminal-utf8'

describe('terminal streaming UTF-8 checkpoint suffix',()=>{
  it('retains only the incomplete code point through one-byte chunks',()=>{
    for(const text of ['¢','中','😀']){
      const bytes=new TextEncoder().encode(text)
      let pending:Uint8Array=new Uint8Array()
      for(let index=0;index<bytes.length;index++){
        pending=pendingUtf8(pending,bytes.slice(index,index+1))
        expect([...pending]).toEqual(index===bytes.length-1?[]:[...bytes.slice(0,index+1)])
      }
    }
  })
  it('discards an interrupted prefix and does not retain complete output',()=>{
    const prefix=new Uint8Array([0xe4,0xb8])
    expect([...pendingUtf8(prefix,new TextEncoder().encode('plain'))]).toEqual([])
    expect([...pendingUtf8(prefix,new TextEncoder().encode('😀中'))]).toEqual([])
    expect([...pendingUtf8(prefix,new Uint8Array([0xc2]))]).toEqual([0xc2])
    expect([...pendingUtf8(new Uint8Array(),new Uint8Array([0x80,0x80]))]).toEqual([])
  })
})
