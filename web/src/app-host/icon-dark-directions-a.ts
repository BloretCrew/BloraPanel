import type { IconLayer } from './classic-icon-layers'
import { darkIconPalette } from './icon-dark-palette'
import type { IconPalette } from './icon-ice-palette'
import type { IconOptics } from './icon-optics'

type Direction = { palette: IconPalette; material: Partial<IconOptics> }

// All three studies keep the approved 01A shapes. Each array follows the
// original paint order in darkIconPalette; even the smallest marks are tuned
// against their own tile rather than brightened by one global filter.
function direction(
  id: string,
  backgrounds: Record<string, string>,
  pigments: Record<string, string[]>,
  terminal: Record<string, string>,
  taskCheck: string,
  material: Partial<IconOptics>,
): Direction {
  const colors: IconPalette['colors'] = {}
  for (const [kind, original] of Object.entries(darkIconPalette.colors)) {
    const keys = Object.keys(original)
    const paint = pigments[kind]
    if (!backgrounds[kind] || !paint || paint.length !== keys.length) {
      throw new Error(`Incomplete dark icon direction ${id}: ${kind}`)
    }
    colors[kind] = Object.fromEntries(keys.map((key, index) => [key, paint[index]!]))
  }
  const recolor = (layer: IconLayer): IconLayer => ({
    ...layer,
    attrs: Object.fromEntries(Object.entries(layer.attrs).map(([key, value]) => [
      key, key === 'fill' || key === 'stroke' ? terminal[value] || value : value,
    ])),
  })
  const terminalArtwork = darkIconPalette.artwork!.terminal!.map(recolor)
  const taskArtwork = darkIconPalette.artwork!.tasks!.map(layer =>
    layer.attrs.d?.startsWith('m31 35')
      ? { ...layer, attrs: { ...layer.attrs, stroke: taskCheck } }
      : layer,
  )
  return {
    palette: {
      id,
      backgrounds,
      colors,
      artwork: { ...darkIconPalette.artwork, terminal: terminalArtwork, tasks: taskArtwork },
    },
    material,
  }
}

// Colour-forward tiles: brighter grounds and luminous, readable app artwork.
const liftedChroma = direction('lifted-chroma', {
  launcher: '#536f7d', instances: '#477bb7', backups: '#428b78', files: '#967448',
  editor: '#846b8c', terminal: '#657e94', tasks: '#7c715f', nodes: '#6781a4',
  monitor: '#478e79', extensions: '#a25768', users: '#806a93', settings: '#678097',
  docker: '#528ca6', system: '#84735b',
}, {
  launcher: ['#89dde5', '#f3be83', '#a2b9f3', '#e1a1ca'],
  instances: ['#b9dcfa', '#86baf0', '#fff8e8', '#daf1b8', '#f8fcff'],
  backups: ['#effff6', '#a8e5cd', '#38987e'],
  files: ['#d6a163', '#fff6dc', '#ffc16c'],
  editor: ['#fff8fc', '#e6cce9', '#b991c0', '#d292d8', '#efbf92', '#806079'],
  terminal: [],
  tasks: ['#fbf7ec', '#4eaf88', '#df8c6c', '#a6bab5'],
  nodes: ['#a9bfdc', '#7db6f4', '#f0aa8c', '#b198e4', '#f5faff', '#ffe9d5'],
  monitor: ['#f1fff7', '#a9ddc6', '#3ca581', '#e8bd80'],
  extensions: ['#e2a2aa', '#fff1e4', '#b86676'],
  users: ['#edb49b', '#f6d3b4', '#bea0de', '#9978bf'],
  settings: ['#c0d2e2', '#59758d', '#e8f2f9'],
  docker: ['#9fd5ec', '#c2e8ed', '#efc390', '#6cafd3', '#f6fcff'],
  system: ['#ccbb9d', '#90d1ad', '#e1a080', '#b1a7dc'],
}, {
  '#2d3b4c': '#304c63', '#72869b': '#7398b7', '#d8e4ef': '#f8fcff',
  '#7fb8af': '#a8f0d6', '#a8bdce': '#c9dce9',
}, '#fafff4', {
  structure: 'core', tint: .94, paper: .95, bevel: 1.35, refraction: 1.35,
  reflection: .12, roughness: 46, shell: 1.1, smoke: 0,
})

// A tinted satin backing makes the silhouette legible through light/dark
// separation; these tiles deliberately stop well short of white.
const softPearl = direction('soft-pearl', {
  launcher: '#8aa0aa', instances: '#8aa3c0', backups: '#8fb4a8', files: '#baa07c',
  editor: '#aa95b0', terminal: '#99a9b8', tasks: '#b2ac9c', nodes: '#9baac2',
  monitor: '#91b2a3', extensions: '#b4929b', users: '#a897b8', settings: '#9caebb',
  docker: '#90b0be', system: '#b4a492',
}, {
  launcher: ['#277d8b', '#b67a45', '#506fad', '#a95d8a'],
  instances: ['#416e9f', '#285589', '#ffffff', '#d8f2bd', '#eaf5ff'],
  backups: ['#ecfff4', '#257d63', '#64b397'],
  files: ['#91602e', '#fff8e8', '#e89d42'],
  editor: ['#fff2fb', '#dec5e0', '#996f9d', '#a254aa', '#e1a875', '#6b506e'],
  terminal: [],
  tasks: ['#fff9ed', '#288e6c', '#c0735e', '#688392'],
  nodes: ['#748dac', '#3d7cc8', '#d78068', '#8266b6', '#f5fbff', '#ffe8d5'],
  monitor: ['#effbed', '#9ad1ae', '#287d62', '#dc9f62'],
  extensions: ['#af6370', '#fff4e7', '#d58987'],
  users: ['#cd8f79', '#ebbea1', '#9d7dbf', '#7657a0'],
  settings: ['#5c7389', '#ecf5fa', '#a6c2d2'],
  docker: ['#4e8eb6', '#7db8cb', '#d9a66c', '#39719b', '#f2fbfd'],
  system: ['#96908b', '#55927b', '#c6866c', '#897caf'],
}, {
  '#2d3b4c': '#40566c', '#72869b': '#6f879d', '#d8e4ef': '#f8fcff',
  '#7fb8af': '#9ae9d4', '#a8bdce': '#d2e2eb',
}, '#ffffff', {
  structure: 'soft', tint: .92, paper: .96, bevel: 1.15, refraction: 1.05,
  reflection: .08, roughness: 32, shell: .8, smoke: 0,
})

// Quiet charcoal tiles let the actual 01A shapes carry the colour. Surface
// opacity is highest here so thin details survive the 55px Dock presentation.
const clearForms = direction('clear-forms', {
  launcher: '#35454e', instances: '#2d4055', backups: '#30463f', files: '#4b4034',
  editor: '#403a48', terminal: '#344552', tasks: '#44443b', nodes: '#363f52',
  monitor: '#314640', extensions: '#4b3941', users: '#41394d', settings: '#374753',
  docker: '#334853', system: '#484037',
}, {
  launcher: ['#73dce3', '#f1b77a', '#92b0f1', '#e29bc5'],
  instances: ['#93c9f6', '#63a3e5', '#ffffff', '#d7f2b7', '#faffff'],
  backups: ['#eefff6', '#7ad3af', '#289c78'],
  files: ['#dfa151', '#fff4dc', '#ffbd5a'],
  editor: ['#fff9fd', '#e8c9e8', '#b781b8', '#cf7dcc', '#ecc08b', '#876284'],
  terminal: [],
  tasks: ['#fffcf1', '#43b88a', '#eb886e', '#bbc9c8'],
  nodes: ['#a1bfde', '#69acf5', '#f1a087', '#a98ae2', '#faffff', '#fff2e4'],
  monitor: ['#effff4', '#ade4c4', '#37b589', '#ffd17f'],
  extensions: ['#dc8892', '#fff4df', '#bb6070'],
  users: ['#eda98b', '#ffcfaa', '#b695db', '#8d69bf'],
  settings: ['#c9dfed', '#476379', '#edf7ff'],
  docker: ['#84cae9', '#bee8ec', '#ffcb85', '#5ca8d2', '#f4fcff'],
  system: ['#cbc5af', '#90d7ba', '#f1ac89', '#b0a7e6'],
}, {
  '#2d3b4c': '#263846', '#72869b': '#688aa4', '#d8e4ef': '#ffffff',
  '#7fb8af': '#80e5ca', '#a8bdce': '#d5e5ee',
}, '#ffffff', {
  structure: 'clear', tint: .98, paper: .98, bevel: 1.05, refraction: .85,
  reflection: .06, roughness: 48, shell: .7, smoke: 0,
})

export const darkDirectionsA: Record<string, Direction> = {
  'lifted-chroma': liftedChroma,
  'soft-pearl': softPearl,
  'clear-forms': clearForms,
}
