import type { IconLayer } from './classic-icon-layers'
import { darkIconPalette } from './icon-dark-palette'
import type { IconPalette } from './icon-ice-palette'
import type { IconOptics } from './icon-optics'

type DarkDirection = { palette: IconPalette; material: Partial<IconOptics> }

// These studies retain the approved 01A silhouettes. Pigments follow each
// app's existing layer order, so even the smallest details keep their role.
function direction(
  id: string,
  backgrounds: Record<string, string>,
  pigments: Record<string, string[]>,
  terminal: Record<string, string>,
  taskChecks: string,
  material: Partial<IconOptics>,
): DarkDirection {
  const colors: IconPalette['colors'] = {}
  for (const [kind, original] of Object.entries(darkIconPalette.colors)) {
    const keys = Object.keys(original)
    const replacements = pigments[kind]
    if (!backgrounds[kind] || !replacements || replacements.length !== keys.length) {
      throw new Error(`Incomplete dark icon direction ${id}: ${kind}`)
    }
    colors[kind] = Object.fromEntries(keys.map((key, i) => [key, replacements[i]!]))
  }

  const repaint = (layers: IconLayer[], replacements: Record<string, string>): IconLayer[] =>
    layers.map(layer => ({
      ...layer,
      attrs: Object.fromEntries(Object.entries(layer.attrs).map(([key, value]) => [
        key, key === 'fill' || key === 'stroke' ? replacements[value] || value : value,
      ])),
    }))

  const artwork = {
    ...darkIconPalette.artwork,
    terminal: repaint(darkIconPalette.artwork!.terminal!, terminal),
    tasks: repaint(darkIconPalette.artwork!.tasks!, { '#dce7dc': taskChecks }),
  }
  return { palette: { ...darkIconPalette, id, backgrounds, colors, artwork }, material }
}

export const darkDirectionsB: Record<string, DarkDirection> = {
  // Dark, nearly neutral tiles make the colorful glyph planes and tiny marks
  // read as a strong two-value silhouette at Dock size.
  'duotone-punch': direction('duotone-punch', {
    launcher: '#344759', instances: '#253f66', backups: '#205347', files: '#513b2c',
    editor: '#4a374f', terminal: '#3d5968', tasks: '#514536', nodes: '#344363',
    monitor: '#235547', extensions: '#653845', users: '#4e3a5f', settings: '#394d67',
    docker: '#2d5066', system: '#594736',
  }, {
    launcher: ['#70d3d8', '#ffbc78', '#8dabff', '#ef91bf'],
    instances: ['#83bfff', '#4b93e3', '#fff5cf', '#d5ffd4', '#eef8ff'],
    backups: ['#edfff6', '#26bd98', '#087e70'],
    files: ['#c98235', '#fff1d1', '#ffb850'],
    editor: ['#fce9f8', '#d8b4d9', '#815d8c', '#c46ed0', '#ffc78c', '#713c71'],
    terminal: [],
    tasks: ['#fff2d9', '#21ae82', '#e8795d', '#5d6f6a'],
    nodes: ['#bfd6eb', '#5fa8ff', '#ffad81', '#af90ef', '#f8fcff', '#fff1df'],
    monitor: ['#e5fff1', '#a3d9c3', '#137d6c', '#ffc384'],
    extensions: ['#d95066', '#ffeddf', '#b9354f'],
    users: ['#efab87', '#d88469', '#cf9ef5', '#b78ae5'],
    settings: ['#afcef5', '#344f6e', '#dcecff'],
    docker: ['#5bb9fa', '#9bdce9', '#ffc17e', '#5bb8ec', '#dff5ff'],
    system: ['#ead8bd', '#4cc1a7', '#ef9676', '#9ca8fa'],
  }, {
    '#2d3b4c': '#1b3344', '#72869b': '#618dab', '#d8e4ef': '#f6fbff',
    '#7fb8af': '#79e8cb', '#a8bdce': '#bfd8e7',
  }, '#203e3b', {
    structure: 'balanced', tint: .94, paper: .97, bevel: 1.15, shell: 1.15,
    refraction: 1.2, reflection: .12, roughness: 32, smoke: 0,
  }),

  // Warmer backplates and creamy object planes keep color identity while
  // reducing the cold, glassy feel of the current dark collection.
  'warm-material': direction('warm-material', {
    launcher: '#5a5960', instances: '#4d5c6b', backups: '#4b665b', files: '#715843',
    editor: '#675567', terminal: '#56616b', tasks: '#696252', nodes: '#575b6f',
    monitor: '#4e665c', extensions: '#795057', users: '#65586d', settings: '#586473',
    docker: '#4e6876', system: '#70604d',
  }, {
    launcher: ['#78c4bd', '#ebba87', '#a2b6df', '#d3a0b8'],
    instances: ['#a9cee9', '#78a8d2', '#fff0cb', '#e2efc2', '#e8f5fa'],
    backups: ['#f7f4de', '#48a786', '#397f71'],
    files: ['#dca767', '#ffefd1', '#f8c27b'],
    editor: ['#f8e7dc', '#dbc4c3', '#947986', '#b779a1', '#e5b481', '#6c5665'],
    terminal: [],
    tasks: ['#f8ebd5', '#438c70', '#ca765e', '#7d8275'],
    nodes: ['#c6cbd8', '#89acd8', '#e3a184', '#ac92c5', '#f7f4e9', '#fff0de'],
    monitor: ['#f4f0d9', '#c8ddc8', '#418d79', '#e1ac77'],
    extensions: ['#c46e75', '#fae7d7', '#a85666'],
    users: ['#dfad8e', '#d9a185', '#d3b4e2', '#c4a2dd'],
    settings: ['#c6d5df', '#60788e', '#e5edf0'],
    docker: ['#78acd2', '#a5d0de', '#e7b986', '#76b8e0', '#e6f0e8'],
    system: ['#e8ddc4', '#7cb59b', '#d58f7c', '#9eadd1'],
  }, {
    '#2d3b4c': '#344754', '#72869b': '#8b9da8', '#d8e4ef': '#fff0da',
    '#7fb8af': '#a6d8c6', '#a8bdce': '#d1d8d4',
  }, '#40554e', {
    structure: 'soft', tint: .91, paper: .97, bevel: 1.2, shell: 1.1,
    refraction: 1.1, reflection: .13, roughness: 24, smoke: 0,
  }),

  // Mid-tone brand-colored tiles frame a bright focal glyph. Keep the shell
  // thin and its specular response low so the center supplies the contrast.
  'luminous-core': direction('luminous-core', {
    launcher: '#536b82', instances: '#416c9f', backups: '#327c69', files: '#816440',
    editor: '#795d80', terminal: '#526e7f', tasks: '#817159', nodes: '#5b6f9b',
    monitor: '#307963', extensions: '#995361', users: '#765d91', settings: '#5c789a',
    docker: '#4e83a0', system: '#897151',
  }, {
    launcher: ['#9aeded', '#ffd49c', '#aec3ff', '#ffb2d6'],
    instances: ['#b2dcff', '#78bdff', '#fffbd9', '#e9ffd1', '#ffffff'],
    backups: ['#f2fff1', '#21b98d', '#157c69'],
    files: ['#e9a658', '#fff8df', '#ffca73'],
    editor: ['#fff2fb', '#e7caed', '#a682b6', '#e99beb', '#ffcb92', '#8d5a92'],
    terminal: [],
    tasks: ['#fff4de', '#269d77', '#d9745c', '#72857b'],
    nodes: ['#d9e8ff', '#87c5ff', '#ffc19d', '#c3a9ff', '#ffffff', '#fff4e5'],
    monitor: ['#ebfff0', '#8cd6ba', '#116c59', '#ffd295'],
    extensions: ['#ee8194', '#fff4e9', '#d84e69'],
    users: ['#ffd0a7', '#f0ad8d', '#e0c1ff', '#d1abf5'],
    settings: ['#d5e9ff', '#4e7194', '#ffffff'],
    docker: ['#a5dcff', '#bceef5', '#ffd797', '#6fbbef', '#ffffff'],
    system: ['#fff0d0', '#91e3c6', '#ffb69d', '#c3c7ff'],
  }, {
    '#2d3b4c': '#274657', '#72869b': '#9bbbc8', '#d8e4ef': '#ffffff',
    '#7fb8af': '#9ff6d9', '#a8bdce': '#e0f1f3',
  }, '#326453', {
    structure: 'core', tint: .98, paper: .99, bevel: .9, shell: .85,
    refraction: .95, reflection: .07, roughness: 22, smoke: 0,
  }),
}
