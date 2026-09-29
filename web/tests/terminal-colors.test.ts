import {it,expect} from 'vitest'
import {terminalColorSequence} from '../src/services/terminal-colors'

it('rejects escape injection, invalid channels and duplicated palette slots',()=>{
  for(const value of [null,{},'\x1b]52;clipboard',Array(260).fill({index:1,rgb:1}),[{index:259,rgb:0}],[{index:1,rgb:-1}],[{index:1,rgb:0x1000000}],[{index:1,rgb:'\x1b]52;clipboard'}],[{index:1,rgb:1},{index:1,rgb:2}]])expect(()=>terminalColorSequence(value)).toThrow()
})
it('emits only fixed native color setters for valid slots and full RGB range',()=>{
  expect(terminalColorSequence([{index:0,rgb:0},{index:255,rgb:0xffffff},{index:256,rgb:0x123456},{index:257,rgb:0xabcdef},{index:258,rgb:0x010203}])).toBe('\x1b]4;0;#000000\x1b\\\x1b]4;255;#ffffff\x1b\\\x1b]10;#123456\x1b\\\x1b]11;#abcdef\x1b\\\x1b]12;#010203\x1b\\')
})
