import {it,expect} from 'vitest'
import {Terminal} from '@xterm/headless'
import {SerializeAddon} from '@xterm/addon-serialize'
import type {Terminal as BrowserTerminal} from '@xterm/xterm'
import {captureTerminalLinks,restoreTerminalLinks} from '../src/services/terminal-links'

it('rejects invalid link spans before changing links, attributes or buffers',async()=>{
  const terminal=new Terminal({cols:80,rows:24,allowProposedApi:true})
  try{
    await new Promise<void>(resolve=>terminal.write('\x1b]8;;https://example.invalid/\x1b\\link\x1b]8;;\x1b\\',resolve))
    const before=captureTerminalLinks(terminal)!
    for(const bad of [
      {...before,version:2},
      {...before,normal:[{...before.normal[0]!,row:24}]},
      {...before,normal:[{...before.normal[0]!,to:81}]},
      {...before,normal:[before.normal[0]!,before.normal[0]!]},
      {...before,current:{...before.current,link:999}},
      {...before,links:[{uri:17}]},
    ]){
      expect(()=>restoreTerminalLinks(terminal,bad,80,24)).toThrow()
      expect(captureTerminalLinks(terminal)).toEqual(before)
    }
  }finally{terminal.dispose()}
})

it('omits metadata for an ordinary terminal with no OSC links',()=>{
  const terminal=new Terminal({cols:80,rows:24})
  try{expect(captureTerminalLinks(terminal)).toBeUndefined()}finally{terminal.dispose()}
})

it('native link markers are released when restored linked rows leave scrollback',async()=>{
  const source=new Terminal({cols:80,rows:3,scrollback:2,allowProposedApi:true}),target=new Terminal({cols:80,rows:3,scrollback:2,allowProposedApi:true})
  const serializer=new SerializeAddon();serializer.activate(source as unknown as BrowserTerminal)
  const write=(terminal:Terminal,data:string)=>new Promise<void>(resolve=>terminal.write(data,resolve))
  try{
    await write(source,'\x1b]8;;https://example.invalid/\x1b\\link\x1b]8;;\x1b\\')
    const state=captureTerminalLinks(source)!
    await write(target,serializer.serialize());restoreTerminalLinks(target,state,80,3)
    expect(captureTerminalLinks(target)?.links).toEqual(state.links)
    await write(target,'\r\nnew'.repeat(10))
    expect(captureTerminalLinks(target)).toBeUndefined()
  }finally{source.dispose();target.dispose()}
})
