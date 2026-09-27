import {describe,it,expect} from 'vitest'
import {ByteWindow,decodeEnvelope,encodeEnvelope,MessageType} from '../src/services/protocol'
describe('management.proto browser transport',()=>{
  it('uses the public protobuf field numbers and preserves UTF-8 raw payloads',()=>{
    const packet={protocolVersion:1,generation:1,channel:2,streamId:'s',type:MessageType.Data,sequence:2,payload:Uint8Array.of(104,105)}
    const golden=Uint8Array.of(8,1,16,1,24,2,50,1,115,56,3,64,2,74,2,104,105)
    expect(encodeEnvelope(packet)).toEqual(golden);expect(decodeEnvelope(golden)).toMatchObject(packet)
    const unicode={...packet,streamId:'session-节点',payload:new TextEncoder().encode('中文\u001b[?1049h')};expect(decodeEnvelope(encodeEnvelope(unicode)).payload).toEqual(unicode.payload)
  })
  it('rejects duplicate security fields, malformed wire types and unsafe integer counters',()=>{
    const valid=encodeEnvelope({protocolVersion:1,generation:1,channel:2,type:MessageType.Open})
    expect(()=>decodeEnvelope(Uint8Array.from([...valid,8,1]))).toThrow('重复')
    expect(()=>decodeEnvelope(Uint8Array.of(10,1,1))).toThrow('类型')
    expect(()=>decodeEnvelope(Uint8Array.from([...valid,64,255,255,255,255,255,255,255,31]))).toThrow('精确整数')
  })
  it('accepts forward compatible unknown non-group fields and rejects truncated lengths',()=>{
    const valid=encodeEnvelope({protocolVersion:1,generation:1,channel:2,type:MessageType.Open})
    expect(decodeEnvelope(Uint8Array.from([...valid,90,3,1,2,3])).type).toBe(MessageType.Open)
    expect(()=>decodeEnvelope(Uint8Array.from([...valid,90,4,1]))).toThrow('截断')
  })
  it('keeps independent direction byte offsets and replenishes only processed bytes',()=>{
    const window=new ByteWindow(10);expect(window.reserve(7)).toBe(7);expect(()=>window.reserve(4)).toThrow('未发送');window.acknowledge(3,3);expect(window.reserve(4)).toBe(11)
    expect(()=>window.acknowledge(11,9)).toThrow('确认');window.acknowledge(11,8);window.receive(5,5);expect(()=>window.receive(7,1)).toThrow('缺口')
  })
})
