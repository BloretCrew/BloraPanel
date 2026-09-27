// Wire-compatible with api/management.proto and internal/protocol/envelope.go.
// Browser transport counters deliberately stop at MAX_SAFE_INTEGER instead of
// silently rounding the protocol's uint64 sequence numbers.
export enum MessageType {Open=1,OpenAck,Data,Credit,Ack,Resize,Close,Reset,Error,Resume,Ping,Pong,TaskEvent,Snapshot}
export interface Envelope {protocolVersion:number;generation:number;channel:number;requestId?:string;runId?:string;streamId?:string;type:MessageType;sequence?:number;payload?:Uint8Array;credit?:number}
export const INTERACTIVE_WINDOW=256*1024
const MAX_MESSAGE=1024*1024,encoder=new TextEncoder(),decoder=new TextDecoder('utf-8',{fatal:true})
export const encodeJSON=(value:unknown)=>encoder.encode(JSON.stringify(value))
export const decodeJSON=<T>(bytes:Uint8Array|undefined):T=>JSON.parse(decoder.decode(bytes))
function validate(e:Envelope){
  if(e.protocolVersion!==1 || e.generation<1 || ![1,2,3].includes(e.channel) || e.type<1 || e.type>14)throw new Error('不支持的管理协议版本、连接代次或消息类型')
  for(const value of [e.generation,e.sequence||0,e.credit||0])if(!Number.isSafeInteger(value)||value<0)throw new Error('管理协议计数超出浏览器精确整数范围')
  for(const value of [e.requestId,e.runId,e.streamId])if(value && encoder.encode(value).length>256)throw new Error('管理协议资源标识过长')
  if([3,4,5,6].includes(e.type) && (!e.streamId || e.channel===1))throw new Error('管理数据缺少流身份或通道不匹配')
  if(e.type===MessageType.Data && !e.payload?.length)throw new Error('不接受空管理数据包')
  if((e.payload?.length||0)>(e.channel===1?64*1024:MAX_MESSAGE))throw new Error('管理消息超过预算')
}
export function encodeEnvelope(e:Envelope):Uint8Array<ArrayBuffer>{
  validate(e);const out:number[]=[]
  const varint=(value:number)=>{let remaining=BigInt(value);while(remaining>=128n){out.push(Number(remaining&127n)|128);remaining>>=7n}out.push(Number(remaining))}
  const integer=(field:number,value:number|undefined)=>{if(value){varint(field*8);varint(value)}}
  const bytes=(field:number,value:Uint8Array|undefined)=>{if(value?.length){varint(field*8+2);varint(value.length);for(const byte of value)out.push(byte)}}
  integer(1,e.protocolVersion);integer(2,e.generation);integer(3,e.channel);bytes(4,e.requestId?encoder.encode(e.requestId):undefined);bytes(5,e.runId?encoder.encode(e.runId):undefined);bytes(6,e.streamId?encoder.encode(e.streamId):undefined);integer(7,e.type);integer(8,e.sequence);bytes(9,e.payload);integer(10,e.credit)
  if(out.length>(e.channel===1?64*1024:MAX_MESSAGE))throw new Error('管理消息超过预算')
  return Uint8Array.from(out)
}
export function decodeEnvelope(bytes:Uint8Array):Envelope{
  if(bytes.length>MAX_MESSAGE)throw new Error('管理消息超过预算')
  let offset=0;const values:Record<number,number|Uint8Array>={},seen=new Set<number>()
  const varint=()=>{let value=0n;for(let index=0;index<10;index++){if(offset>=bytes.length)throw new Error('截断的管理消息');const byte=bytes[offset++]!;if(index===9 && byte>1)throw new Error('管理消息整数溢出');value|=BigInt(byte&127)<<BigInt(index*7);if(!(byte&128)){if(value>BigInt(Number.MAX_SAFE_INTEGER))throw new Error('管理协议计数超出浏览器精确整数范围');return Number(value)}}throw new Error('无效管理消息整数')}
  while(offset<bytes.length){
    const tag=varint(),field=Math.floor(tag/8),wire=tag%8
    if(field<1 || field>0x1fffffff || ![0,1,2,5].includes(wire))throw new Error('无效管理消息字段')
    if(field<=10){if(seen.has(field))throw new Error('重复的管理消息字段');seen.add(field);if(([4,5,6,9].includes(field)?2:0)!==wire)throw new Error('管理消息字段类型错误')}
    let value:number|Uint8Array
    if(wire===0)value=varint()
    else{const size=wire===2?varint():wire===1?8:4;if(offset+size>bytes.length)throw new Error('截断的管理消息');value=bytes.slice(offset,offset+size);offset+=size}
    if(field<=10)values[field]=value
  }
  const integer=(field:number)=>Number(values[field]||0),string=(field:number)=>values[field]?decoder.decode(values[field] as Uint8Array):undefined
  const result:Envelope={protocolVersion:integer(1),generation:integer(2),channel:integer(3),requestId:string(4),runId:string(5),streamId:string(6),type:integer(7),sequence:integer(8),payload:values[9] as Uint8Array|undefined,credit:integer(10)}
  validate(result);if(result.channel===1 && bytes.length>64*1024)throw new Error('控制消息超过预算');return result
}
export class ByteWindow {
  sent=0;acknowledged=0;received=0
  constructor(readonly limit=INTERACTIVE_WINDOW){}
  reserve(bytes:number){if(!Number.isSafeInteger(bytes)||bytes<=0||bytes>this.limit-(this.sent-this.acknowledged))throw new Error('终端输入额度不足，此次输入未发送');this.sent+=bytes;if(!Number.isSafeInteger(this.sent))throw new Error('连接计数达到上限，请重新挂载');return this.sent}
  acknowledge(sequence:number,credit:number){if(sequence<this.acknowledged||sequence>this.sent||credit!==sequence-this.acknowledged)throw new Error('终端输入确认序号无效');this.acknowledged=sequence}
  receive(sequence:number,bytes:number){if(bytes<=0||sequence!==this.received+bytes)throw new Error('终端连接字节序号有缺口');this.received=sequence}
}
