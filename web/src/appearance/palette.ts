export const themePalettes = [
  {id:'ice',name:'冰蓝'},
  {id:'mineral',name:'青灰'},
  {id:'sand',name:'暖砂'},
  {id:'sage',name:'苔绿'},
  {id:'iris',name:'雾紫'},
  {id:'rose',name:'烟粉'},
] as const
export type ThemePalette = typeof themePalettes[number]['id']
export function isThemePalette(value: unknown): value is ThemePalette {
  return themePalettes.some(palette=>palette.id===value)
}

// Existing workspaces saved only the translucency switch. Preserve their
// current blue/teal appearance until they explicitly choose a palette.
export function themePalette(preferences: {palette?: unknown;translucency?: unknown}): ThemePalette {
  if (isThemePalette(preferences.palette)) return preferences.palette
  return preferences.translucency === false ? 'mineral' : 'ice'
}
