export interface LogCheckpoint {runId:string;sequence:number;text:string;carry:number[];discardedLines:number;truncated:boolean}
export interface LogEvent {sequence:number;kind:string;data?:string}
export const LOG_TEXT_LIMIT=48000,LOG_LINE_LIMIT=3000
export const emptyLog=(runId:string):LogCheckpoint=>({runId,sequence:0,text:'',carry:[],discardedLines:0,truncated:false})
export function decodeLogBytes(carry:number[],bytes:Uint8Array){
  const joined=new Uint8Array(carry.length+bytes.length);joined.set(carry);joined.set(bytes,carry.length);let cut=joined.length
  // TextDecoder's streaming carry is private. Retain a possible partial UTF-8
  // suffix explicitly so an event checkpoint also protects split characters.
  for(let i=joined.length-1;i>=Math.max(0,joined.length-4);i--){const byte=joined[i]!;if((byte&0xc0)===0x80)continue;const required=byte>=0xc2&&byte<=0xdf?2:byte>=0xe0&&byte<=0xef?3:byte>=0xf0&&byte<=0xf4?4:1;if(required>joined.length-i)cut=i;break}
  return{text:new TextDecoder().decode(joined.subarray(0,cut)),carry:Array.from(joined.subarray(cut))}
}
export function appendLog(checkpoint:LogCheckpoint,event:LogEvent):LogCheckpoint{
  if(!Number.isSafeInteger(event.sequence)||event.sequence<1)throw new Error('日志事件序号无效')
  if(event.sequence<=checkpoint.sequence)return checkpoint
  if(event.sequence!==checkpoint.sequence+1)throw new Error('日志序号不连续，已保留最后检查点')
  if(event.kind!=='output'||typeof event.data!=='string')throw new Error('日志输出格式无效')
  const decoded=decodeLogBytes(checkpoint.carry,Uint8Array.from(atob(event.data),c=>c.charCodeAt(0)))
  let text=checkpoint.text+decoded.text,discardedLines=checkpoint.discardedLines,truncated=checkpoint.truncated
  if(text.length>LOG_TEXT_LIMIT){let cut=text.length-LOG_TEXT_LIMIT;if(text.charCodeAt(cut)>=0xdc00&&text.charCodeAt(cut)<=0xdfff)cut++;discardedLines+=text.slice(0,cut).split('\n').length-1;text=text.slice(cut);truncated=true}
  const lines=text.split('\n');if(lines.length>LOG_LINE_LIMIT){const count=lines.length-LOG_LINE_LIMIT;discardedLines+=count;text=lines.slice(count).join('\n');truncated=true}
  return{...checkpoint,sequence:event.sequence,text,carry:decoded.carry,discardedLines,truncated}
}
export function logGap(checkpoint:LogCheckpoint,earliest:number):LogCheckpoint{
  if(!Number.isSafeInteger(earliest)||earliest<1)throw new Error('日志保留范围无效')
  return{...checkpoint,sequence:earliest-1,text:checkpoint.text+'\n[较早日志存在归档缺口]\n',carry:[],truncated:true}
}
export function logText(text:string){return text.replace(/\x1b\][^\x07]*(?:\x07|\x1b\\)/g,'').replace(/\x1b\[[0-?]*[ -/]*[@-~]/g,'').replace(/\r\n/g,'\n').replace(/[^\n]*\r/g,'').replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g,'')}
