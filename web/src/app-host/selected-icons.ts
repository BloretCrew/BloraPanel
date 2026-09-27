// Both material modes use the approved 01A geometry. Colour theme selects its
// light or dark paints; dark mode uses the chosen 03 clear-forms artwork.
const assets = import.meta.glob<string>('../assets/app-icons/h04/*.png', {
  eager: true,
  query: '?url',
  import: 'default',
})
const sharedAssets = import.meta.glob<string>('../assets/app-icons/h04-spectrum-silver/*.png', {
  eager: true,
  query: '?url',
  import: 'default',
})
const darkAssets = import.meta.glob<string>('../assets/app-icons/h04-spectrum-night-clear-forms/*.png', {
  eager: true,
  query: '?url',
  import: 'default',
})
const darkPreview = import.meta.env.DEV ? new URLSearchParams(location.search).get('dark-icon-preview') : null
const darkPreviewIds = new Set(['baseline', 'lifted-chroma', 'soft-pearl', 'clear-forms', 'duotone-punch', 'warm-material', 'luminous-core'])
export function selectedIcon(kind: string, preview: string | null = null, theme: 'light' | 'dark' = 'light'): string {
  const safeKind = assets[`../assets/app-icons/h04/${kind}.png`] ? kind : 'launcher'
  if (theme === 'dark' && darkPreview && darkPreviewIds.has(darkPreview)) {
    const previewDirectory = darkPreview === 'baseline' ? 'h04-spectrum-night' : `h04-spectrum-night-${darkPreview}`
    return `/src/assets/app-icons/${previewDirectory}/${safeKind}.png`
  }
  if (theme === 'dark' && (!preview || preview === 'spectrum-silver')) return darkAssets[`../assets/app-icons/h04-spectrum-night-clear-forms/${safeKind}.png`] || darkAssets['../assets/app-icons/h04-spectrum-night-clear-forms/launcher.png']!
  if (preview === 'spectrum-silver') return sharedAssets[`../assets/app-icons/h04-spectrum-silver/${safeKind}.png`] || sharedAssets['../assets/app-icons/h04-spectrum-silver/launcher.png']!
  // The review assets are served directly by Vite in development. They are
  // deliberately absent from the production module graph and release bundle.
  if (import.meta.env.DEV && preview) return `/src/assets/app-icons/h04-${preview}/${safeKind}.png`
  return assets[`../assets/app-icons/h04/${safeKind}.png`] || assets['../assets/app-icons/h04/launcher.png']!
}
