// xterm 6's screen serializer preserves indexed cell references, not the
// palette changed by OSC. Keep only differences from the current theme so
// restoring a checkpoint does not replace the user's theme defaults.
type Color={rgba:number}
type Palette={foreground:Color;background:Color;cursor:Color;ansi:Color[]}
export type TerminalColors=Array<{index:number;rgb:number}>
function validate(value:unknown):TerminalColors{
  if(!Array.isArray(value)||value.length>259)throw new Error('终端颜色恢复记录无效')
  const seen=new Set<number>()
  return value.map(entry=>{
    if(!entry||typeof entry!=='object')throw new Error('终端颜色恢复条目无效')
    const {index,rgb}=entry as {index:number;rgb:number}
    if(!Number.isInteger(index)||index<0||index>258||seen.has(index)||!Number.isInteger(rgb)||rgb<0||rgb>0xffffff)throw new Error('终端颜色恢复数值无效')
    seen.add(index);return {index,rgb}
  })
}
export function captureTerminalColors(terminal:unknown):TerminalColors{
  const theme=(terminal as {_core?:{_themeService?:{colors:Palette;_restoreColors:Palette}}})._core?._themeService
  if(!theme||theme.colors.ansi.length!==256||theme._restoreColors.ansi.length!==256)throw new Error('终端颜色状态接口不兼容')
  const flatten=(palette:Palette)=>[...palette.ansi,palette.foreground,palette.background,palette.cursor]
  const current=flatten(theme.colors),defaults=flatten(theme._restoreColors)
  return validate(current.flatMap((color,index)=>color.rgba===defaults[index]!.rgba?[]:[{index,rgb:color.rgba>>>8}]))
}
export function terminalColorSequence(value:unknown):string{
  // Never accept arbitrary escape strings from a stored checkpoint. Rebuild
  // only the fixed OSC set operations after validating every palette entry.
  return validate(value).map(({index,rgb})=>`\x1b]${index<256?'4;'+index:10+index-256};#${rgb.toString(16).padStart(6,'0')}\x1b\\`).join('')
}
