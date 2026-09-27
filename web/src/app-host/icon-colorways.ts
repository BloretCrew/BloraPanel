import { iceIconPalette, type IconPalette } from './icon-ice-palette'

// Four app-by-app color studies. Each array follows the original layer colors
// recorded by icon-ice-palette; only pigment and backing change. The h04 paths,
// optical material, backing silhouette and icon size remain identical.
function colorway(id: string, backgrounds: Record<string, string>, pigments: Record<string, string[]>): IconPalette {
  const colors: IconPalette['colors'] = {}
  for (const [kind, original] of Object.entries(iceIconPalette.colors)) {
    const keys = Object.keys(original)
    const replacements = pigments[kind]
    if (!backgrounds[kind] || !replacements || replacements.length !== keys.length) {
      throw new Error(`Incomplete ${id} icon colorway: ${kind}`)
    }
    colors[kind] = Object.fromEntries(keys.map((key, index) => [key, replacements[index]!]))
  }
  return { id, backgrounds, colors }
}

export const iconColorways = {
  spectrum: colorway('spectrum', {
    launcher: '#f8fbff', instances: '#427deb', backups: '#2bb68d', files: '#fff5e7',
    editor: '#fcf1fa', terminal: '#232d39', tasks: '#fffaf3', nodes: '#edf5ff',
    monitor: '#183740', extensions: '#e95861', users: '#f0e8fc', settings: '#d4dae4',
    docker: '#e5f9ff', system: '#fff2dc',
  }, {
    launcher: ['#52c5d5', '#f4a55f', '#777ce8', '#e57aa6'],
    instances: ['#bfddff', '#eaf5ff', '#fffaf0', '#ddffac', '#5b96dd'],
    backups: ['#f4fff9', '#8ddfba', '#116f5d'],
    files: ['#b9792c', '#fffdf3', '#e9a84c'],
    editor: ['#fffafd', '#edcee9', '#b98aaf', '#bb4fa8', '#dbaf82', '#76516d'],
    terminal: ['#293b48', '#f8fcff', '#5be0bd'],
    tasks: ['#fffdfa', '#26b489', '#f0785c', '#879bb0'],
    nodes: ['#adc4e8', '#4c85e8', '#f28d70', '#8f6ce5', '#f4fbff', '#fff2e9'],
    monitor: ['#203f48', '#2e5b58', '#77e5c6', '#ffd17d'],
    extensions: ['#a93346', '#fff5ef', '#ffad9f'],
    users: ['#f8aa7b', '#ffcda5', '#8b66c5', '#7751ae'],
    settings: ['#607083', '#f8fbff', '#97aed0'],
    docker: ['#178bcf', '#48c8df', '#f7b757', '#2262ae', '#f7fcff'],
    system: ['#9da8bd', '#32b69b', '#f28c68', '#7a74d3'],
  }),
  porcelain: colorway('porcelain', {
    launcher: '#fbfdff', instances: '#f2f6ff', backups: '#f1fbf4', files: '#fff7ec',
    editor: '#fff5fb', terminal: '#f2f0fc', tasks: '#fffdfa', nodes: '#f5f4ff',
    monitor: '#eefbf8', extensions: '#fff2f0', users: '#f8f2ff', settings: '#eff3f6',
    docker: '#effaff', system: '#fff8ee',
  }, {
    launcher: ['#43bfca', '#f4a45f', '#5189e5', '#db7fbc'],
    instances: ['#70a0ef', '#355fc4', '#fffbe9', '#e3ff9e', '#f2f7ff'],
    backups: ['#f9fffc', '#169a66', '#56d1a2'],
    files: ['#bc7724', '#fffdf0', '#efaa4e'],
    editor: ['#fffefd', '#e3b6d6', '#a978a6', '#bd4695', '#e9bc8c', '#725067'],
    terminal: ['#392d70', '#fffaff', '#ff9bce'],
    tasks: ['#ffffff', '#27b88e', '#ed7b67', '#8794a6'],
    nodes: ['#a9b1e5', '#4c80dc', '#f08069', '#9866d0', '#fffaf6', '#fff1e6'],
    monitor: ['#f7fffc', '#b0e9d3', '#118d80', '#ff9f65'],
    extensions: ['#d75655', '#fffdf4', '#f18d79'],
    users: ['#ef9d8d', '#ffcabc', '#9d73d0', '#7248ad'],
    settings: ['#657c94', '#f8fbfd', '#9cb8d5'],
    docker: ['#237bc1', '#58bcdc', '#ecb268', '#2859ad', '#f5fbff'],
    system: ['#9aa5b5', '#1f9e7f', '#e87d65', '#696ed4'],
  }),
  contrast: colorway('contrast', {
    launcher: '#f8fbff', instances: '#1c2c5a', backups: '#176f5b', files: '#dcefff',
    editor: '#33303e', terminal: '#16232c', tasks: '#f8bd5b', nodes: '#282c54',
    monitor: '#edfff8', extensions: '#f46a53', users: '#efe3ff', settings: '#354759',
    docker: '#197db6', system: '#f5f6f7',
  }, {
    launcher: ['#50bfce', '#f4ae5d', '#6d88de', '#dc77a4'],
    instances: ['#8bbaff', '#cce6ff', '#fff9e2', '#d4f7ae', '#315fb5'],
    backups: ['#f6fff9', '#c8ffdf', '#6de5b7'],
    files: ['#267abb', '#ffffff', '#46a2dc'],
    editor: ['#faf8ff', '#e2c6e5', '#9a85ab', '#d477ca', '#d9ae8a', '#76516c'],
    terminal: ['#243746', '#fcffff', '#6ae6cb'],
    tasks: ['#fffefa', '#3a9e72', '#e17c55', '#898d87'],
    nodes: ['#758bbb', '#69acff', '#f68d7a', '#a58bf5', '#f5fbff', '#fff2ec'],
    monitor: ['#f8fffc', '#d1f3e5', '#168879', '#f19661'],
    extensions: ['#b64542', '#fff4e7', '#ffd68b'],
    users: ['#dc9b93', '#f5c7b8', '#a879d1', '#8055af'],
    settings: ['#d6e2f0', '#52657b', '#8aaed1'],
    docker: ['#b9e5ff', '#effcff', '#ffc06b', '#f2faff', '#225c94'],
    system: ['#b4baca', '#39b99b', '#e9906d', '#7d79d0'],
  }),
  pastel: colorway('pastel', {
    launcher: '#fafbff', instances: '#d9e8ff', backups: '#ddf7eb', files: '#fff0cd',
    editor: '#f5e6f9', terminal: '#27303e', tasks: '#fff5eb', nodes: '#e8e7ff',
    monitor: '#173e41', extensions: '#ffe1dc', users: '#eee3fb', settings: '#dce4ed',
    docker: '#e1f8ff', system: '#f9e8df',
  }, {
    launcher: ['#52c4cf', '#f9aa65', '#6e92e8', '#df87bd'],
    instances: ['#699fe8', '#3774c7', '#fffef7', '#dfffad', '#ecf9ff'],
    backups: ['#fafffa', '#32a879', '#8bd6b4'],
    files: ['#c98b35', '#fffdf4', '#f5b550'],
    editor: ['#fffcff', '#eac6e3', '#ac83b4', '#a653c3', '#e5b68c', '#775177'],
    terminal: ['#293d4c', '#fffaff', '#95e6d0'],
    tasks: ['#fffdfa', '#76bd7b', '#e98f68', '#ac9c87'],
    nodes: ['#aabbe7', '#5c8ce1', '#ee957e', '#a284e1', '#f8faff', '#ffeddf'],
    monitor: ['#24484a', '#426b64', '#99dfb4', '#efbb7c'],
    extensions: ['#d06d70', '#fff8eb', '#f2a9a7'],
    users: ['#e9a28b', '#f7c6b2', '#a785cf', '#7b63b2'],
    settings: ['#718baa', '#f7fbff', '#b1d2e8'],
    docker: ['#5b9bd2', '#92c6e8', '#e8b786', '#4184c1', '#f7fcff'],
    system: ['#bbaaa6', '#5cb4a1', '#df8f77', '#8a94c8'],
  }),
} as const

export type IconColorwayId = keyof typeof iconColorways

export function iconColorway(id: string | null): IconPalette | undefined {
  return id && Object.prototype.hasOwnProperty.call(iconColorways, id)
    ? iconColorways[id as IconColorwayId]
    : undefined
}
