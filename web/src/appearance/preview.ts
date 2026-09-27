// A single opt-in appearance study; saved preferences and production remain unchanged.
export const iceAppearance = import.meta.env.DEV && new URLSearchParams(location.search).get('appearance') === 'ice'
if (iceAppearance) {
  document.documentElement.dataset.appearance = 'ice'
  document.documentElement.dataset.palette = 'ice'
}
// Wallpaper review stays independent of icon color; paths are served only by
// the development server and are not imported into the production bundle.
if (import.meta.env.DEV && iceAppearance) {
  const wallpaper = new URLSearchParams(location.search).get('wallpaper')
  if (wallpaper && ['warm', 'strata', 'haze', 'flow', 'linen'].includes(wallpaper)) {
    document.documentElement.dataset.wallpaper = wallpaper
    document.documentElement.style.setProperty('--review-wallpaper', `url('/src/appearance/wallpapers/${wallpaper}.svg')`)
  }
}
