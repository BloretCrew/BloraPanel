import {decodeJSON} from './protocol'
import type {TerminalEvent} from './terminals'
// The byte envelope still accounts for the complete wire payload. Decoding
// does not parse terminal output or publish an ACK; TerminalModel retains its
// existing ordered parser, journal and checkpoint boundary for every event.
export function decodeTerminalEvents(payload:Uint8Array|undefined):TerminalEvent[]{
 const value=decodeJSON<unknown>(payload)
 if(!value||typeof value!=='object'||Array.isArray(value))throw Error('终端事件内容无效')
 if((value as {kind?:unknown}).kind!=='batch')return [value as TerminalEvent]
 const events=(value as {events?:unknown}).events
 if(!Array.isArray(events)||!events.length||events.length>128||events.some(event=>!event||typeof event!=='object'||Array.isArray(event)))throw Error('终端事件批次超出预算或内容无效')
 return events as TerminalEvent[]
}
