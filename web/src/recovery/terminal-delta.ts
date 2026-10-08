import type {Workspace} from '../app-host/types'
import {applyEntry,migrateWorkspace,type JournalEntry} from './state'

// Durable representation only. Cloud/export state remains a complete Workspace.
// A delta always references one complete base, never another delta. Its small
// bounded payload avoids cloning cold editor/window state for every PTY batch.
export const TERMINAL_DELTA_FORMAT='blora-terminal-delta-v1'
export const TERMINAL_DELTA_ENTRIES=64
export const TERMINAL_DELTA_BYTES=128*1024
export type TerminalDelta={format:typeof TERMINAL_DELTA_FORMAT;base:string;baseRevision:number;revision:number;entries:JournalEntry[]}
export function isTerminalDelta(value:unknown):value is TerminalDelta{
 return !!value&&typeof value==='object'&&(value as {format?:unknown}).format===TERMINAL_DELTA_FORMAT
}
function terminalEntries(entries:JournalEntry[],baseRevision:number,revision:number){
 return entries.length>0&&entries.length<=TERMINAL_DELTA_ENTRIES&&Number.isSafeInteger(baseRevision)&&baseRevision>=0&&Number.isSafeInteger(revision)&&revision===baseRevision+entries.length&&entries.every((entry,index)=>entry&&entry.revision===baseRevision+index+1&&Array.isArray(entry.mutations)&&entry.mutations.length>0&&entry.mutations.every(op=>op?.kind==='terminal-output'))
}
export function terminalDelta(base:string,baseRevision:number,entries:JournalEntry[],revision:number):TerminalDelta|undefined{
 if(!terminalEntries(entries,baseRevision,revision)||JSON.stringify(entries).length*2>TERMINAL_DELTA_BYTES)return undefined
 return {format:TERMINAL_DELTA_FORMAT,base,baseRevision,revision,entries}
}
export function restoreTerminalDelta(delta:TerminalDelta,base:unknown,key:string,current:string):Workspace{
 if(delta.base!==`${key}:${delta.baseRevision}`||delta.base===current||current!==`${key}:${delta.revision}`||!Array.isArray(delta.entries)||!terminalDelta(delta.base,delta.baseRevision,delta.entries,delta.revision)||isTerminalDelta(base))throw Error('终端增量快照无效，已保留原恢复记录')
 const state=migrateWorkspace(base as Workspace)
 if(state.revision!==delta.baseRevision)throw Error('终端增量快照基线不匹配，已保留原恢复记录')
 for(const entry of delta.entries)applyEntry(state,entry)
 return state
}
