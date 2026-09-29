// xterm's screen serializer does not include its streaming UTF-8 decoder.
// Only a final incomplete code point is needed: at most three pending bytes.
export function pendingUtf8(previous:Uint8Array,data:Uint8Array):Uint8Array{
  if(!data.length)return previous
  const suffix=data.subarray(Math.max(0,data.length-4))
  const tail=data.length>=4?suffix:new Uint8Array([...previous,...suffix])
  let start=tail.length-1
  while(start>=0&&(tail[start]!&0xc0)===0x80)start--
  if(start<0)return new Uint8Array()
  const lead=tail[start]!
  const width=(lead&0xe0)===0xc0?2:(lead&0xf0)===0xe0?3:(lead&0xf8)===0xf0?4:0
  return width>tail.length-start?tail.slice(start):new Uint8Array()
}
