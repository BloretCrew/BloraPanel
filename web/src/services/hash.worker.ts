/// <reference lib="webworker" />
import {sha256} from '@noble/hashes/sha2.js'
import {bytesToHex} from '@noble/hashes/utils.js'
self.onmessage=async(event:MessageEvent<{file:File}>)=>{
  try{
    const file=event.data.file,hash=sha256.create();let offset=0,lastProgress=0
    while(offset<file.size){const chunk=new Uint8Array(await file.slice(offset,offset+256*1024).arrayBuffer());hash.update(chunk);offset+=chunk.length;if(offset-lastProgress>=1024*1024 || offset===file.size){self.postMessage({offset});lastProgress=offset}}
    self.postMessage({hash:'sha256:'+bytesToHex(hash.digest())})
  }catch(e){self.postMessage({error:String(e)})}
}
