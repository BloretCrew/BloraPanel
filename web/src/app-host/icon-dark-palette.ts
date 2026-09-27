import type { IconLayer } from './classic-icon-layers'
import { classicIconLayers } from './classic-icon-layers'
import type { IconPalette } from './icon-ice-palette'
import { utilityVariants } from './icon-utility-variants'

// Keep the approved 01A app silhouettes and their distinct colour identities.
// Dark mode keeps the 01A shapes and middle-value backplates. Foreground planes
// are muted separately from their small marks so the glyph still reads at 55px.
const base = utilityVariants['spectrum-silver']!

const pigments: Record<string, string[]> = {
  launcher: ['#67b2bb', '#d59b68', '#889bd5', '#bf779e'],
  instances: ['#80abd5', '#6e9aca', '#e6d9c2', '#cbdca7', '#bfd4e7'],
  backups: ['#b4d1c5', '#58aa95', '#338878'],
  files: ['#b5844b', '#d1bea4', '#c99348'],
  editor: ['#bfb2c4', '#aa96b4', '#70657e', '#aa72aa', '#c49d7c', '#675064'],
  terminal: [], // The terminal has its own pane artwork below.
  tasks: ['#adbfba', '#328a75', '#b86e5b', '#5f7580'],
  nodes: ['#899db6', '#7199cd', '#c48470', '#917db6', '#d0dce9', '#ddcbbf'],
  monitor: ['#a5c2b7', '#7cae9b', '#267967', '#c79669'],
  extensions: ['#b66770', '#d7b8b0', '#944f5c'],
  users: ['#c18f7a', '#cea58a', '#9678b3', '#80639c'],
  settings: ['#91a8be', '#cbd8e5', '#7792ae'],
  docker: ['#6ba0c8', '#80b4c7', '#cc9f69', '#447ba7', '#c8dfe8'],
  system: ['#aea28a', '#67a98e', '#bf866e', '#928dbb'],
}

const colors = Object.fromEntries(Object.entries(base.colors).map(([kind, original]) => {
  const keys = Object.keys(original)
  const paint = pigments[kind]
  if (!paint || paint.length !== keys.length) throw new Error(`Incomplete dark icon palette: ${kind}`)
  return [kind, Object.fromEntries(keys.map((key, i) => [key, paint[i]!]))]
}))

const terminalPaint: Record<string, string> = {
  '#35414f': '#2d3b4c',
  '#8392a6': '#72869b',
  '#fbfcff': '#d8e4ef',
  '#83bcff': '#7fb8af',
  '#c3cfdf': '#a8bdce',
}
const terminalArtwork = base.artwork!.terminal!.map(layer => {
  const attrs = Object.fromEntries(Object.entries(layer.attrs).map(([key, value]) => [
    key, key === 'fill' || key === 'stroke' ? terminalPaint[value] || value : value,
  ]))
  if (attrs.d === 'm29 45 7 6-7 6') {
    attrs.d = 'm27 43 9 8-9 8'
    attrs['stroke-width'] = '4.3'
  }
  if (layer.tag === 'rect' && attrs.x === '42' && attrs.y === '53') {
    attrs.x = '43'
    attrs.y = '55'
    attrs.width = '15'
    attrs.height = '4.4'
  }
  return { ...layer, attrs } as IconLayer
})

// The checklist's board and its check strokes share one classic paint key.
// Separate them so a quieter board does not erase the checkmarks at Dock size.
const taskArtwork = classicIconLayers.tasks!.map(layer => layer.attrs.d?.startsWith('m31 35')
  ? { ...layer, attrs: { ...layer.attrs, stroke: '#dce7dc' } }
  : layer)

export const darkIconPalette: IconPalette = {
  ...base,
  id: 'spectrum-night',
  backgrounds: {
    launcher: '#566879',
    instances: '#3d6091',
    backups: '#357865',
    files: '#6c563d',
    editor: '#65536b',
    terminal: '#536678',
    tasks: '#665e50',
    nodes: '#566783',
    monitor: '#397768',
    extensions: '#884a58',
    users: '#665575',
    settings: '#586e86',
    docker: '#48728a',
    system: '#71614c',
  },
  colors,
  artwork: { ...base.artwork, terminal: terminalArtwork, tasks: taskArtwork },
}
