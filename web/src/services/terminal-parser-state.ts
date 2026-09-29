import type {Terminal} from '@xterm/xterm'

// The pinned xterm 6 parser exposes currentState internally, but its public
// serializer omits unfinished CSI/OSC/DCS state. Do not clone that parser or
// guess from output suffixes: compact only at its actual GROUND boundary.
// Dependency upgrades must keep the real-parser recovery tests green. An
// incompatible internal layout is a visible failure, never a lossy checkpoint.
export function terminalParserAtBoundary(terminal:Terminal):boolean{
  const parser=(terminal as unknown as {_core?:{_inputHandler?:{_parser?:{currentState?:number}}}})._core?._inputHandler?._parser
  const state=parser?.currentState
  if(!Number.isInteger(state)||state!<0||state!>13)throw new Error('终端解析器状态接口不兼容，无法安全生成检查点')
  return state===0
}
