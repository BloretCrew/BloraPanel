export const wallpaperStyles = [
  {id:'palette',name:'随配色'},
  {id:'photo',name:'自然摄影'},
  {id:'poster',name:'印刷海报'},
  {id:'nocturne',name:'数字夜光'},
  {id:'ink',name:'水墨纸面'},
  {id:'material',name:'材质微距'},
  {id:'atmosphere',name:'柔和色场'},
  {id:'planes',name:'建筑平面'},
  {id:'topography',name:'等高线'},
  {id:'b-grid',name:'编辑网格'},
  {id:'b-monolith',name:'抽象雕塑'},
  {id:'b-horizon',name:'静谧地平线'},
] as const

export type WallpaperStyle = typeof wallpaperStyles[number]['id']

export function wallpaperStyle(preferences:{wallpaper?:unknown}):WallpaperStyle {
  return wallpaperStyles.find(style=>style.id===preferences.wallpaper)?.id ?? 'palette'
}
