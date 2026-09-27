import { utilityVariants } from './icon-utility-variants'
import type { IconPalette } from './icon-ice-palette'

// Keep 01A's geometry and h04 optical material. These studies change the
// relationship between backing and foreground across the entire collection.
function createHarmonyPalettes(): Record<string, IconPalette> {
  const base = utilityVariants['spectrum-silver']!
  function palette(id: string, backgrounds: Record<string, string>, pigments: Record<string, string[]>, terminal: Record<string, string>): IconPalette {
    const colors = { ...base.colors }
    for (const kind of Object.keys(base.backgrounds)) {
      if (!backgrounds[kind]) throw new Error(`Missing ${id} backing: ${kind}`)
      if (kind === 'terminal') continue
      const keys = Object.keys(base.colors[kind]!)
      if (pigments[kind]?.length !== keys.length) throw new Error(`Incomplete ${id} artwork colors: ${kind}`)
      colors[kind] = Object.fromEntries(keys.map((key, index) => [key, pigments[kind]![index]!]))
    }
    return { id, backgrounds, colors, artwork: {
      terminal: base.artwork!.terminal!.map(layer => ({ ...layer, attrs: Object.fromEntries(Object.entries(layer.attrs).map(([key, value]) => [key, (key === 'fill' || key === 'stroke') ? terminal[value] || value : value])) })),
    } }
  }
  const porcelain = palette('unify-porcelain', {
    launcher: '#f7f8f7', instances: '#f3f6fa', backups: '#f4f8f5', files: '#fcf7ef',
    editor: '#faf5f9', terminal: '#f3f5f7', tasks: '#fbf8f2', nodes: '#f5f6fb',
    monitor: '#f4f8f4', extensions: '#fcf4f2', users: '#f8f4fa', settings: '#f2f5f7',
    docker: '#f1f8fa', system: '#faf6ef',
  }, {
    launcher: ['#45afbf', '#e7a05b', '#6484d0', '#cf79a1'],
    instances: ['#598fce', '#3066a7', '#fff8df', '#d8efaa', '#f1f7ff'],
    backups: ['#fcfdf8', '#25896d', '#60b293'],
    files: ['#b17b30', '#fffbed', '#e5a84e'],
    editor: ['#fffcfb', '#e0c6db', '#a37d9d', '#a34d9a', '#d9aa7a', '#755267'],
    tasks: ['#fffefa', '#36a278', '#d6795e', '#8b98a5'],
    nodes: ['#adbadb', '#4b81c9', '#df8b70', '#9874ca', '#f8fbff', '#fff6ea'],
    monitor: ['#fbfff9', '#c6e5d1', '#248c60', '#edaa64'],
    extensions: ['#cc6462', '#fff9ef', '#e59182'],
    users: ['#e1a077', '#f1cba6', '#9670c1', '#7454a2'],
    settings: ['#6a7e92', '#fcfdff', '#9fbbd3'],
    docker: ['#3992c7', '#73c2d9', '#ecb675', '#3275ad', '#f8fcff'],
    system: ['#a5afba', '#4aa28f', '#db8c70', '#8b80c0'],
  }, { '#35414f': '#526276', '#8392a6': '#98a8b9', '#fbfcff': '#ffffff', '#83bcff': '#9cc9f1', '#c3cfdf': '#d5deea' })

  const tonal = palette('unify-tonal', {
    launcher: '#d7e6e5', instances: '#b8d2ee', backups: '#b6ded1', files: '#efdab5',
    editor: '#dfc8e0', terminal: '#cbd6e2', tasks: '#e9dfc7', nodes: '#ccd3ee',
    monitor: '#c4dfc5', extensions: '#e9c5bf', users: '#d9cbee', settings: '#cad4dc',
    docker: '#bcdde8', system: '#e3d3c7',
  }, {
    launcher: ['#3a98a9', '#d69a57', '#4f78ba', '#bc7594'],
    instances: ['#4f84bc', '#2e5f95', '#fff6d8', '#d5edac', '#edf6ff'],
    backups: ['#f0f9f2', '#31876d', '#70b199'],
    files: ['#a77831', '#fffbed', '#d4a04e'],
    editor: ['#fffafd', '#ceabc9', '#986e95', '#955797', '#d8ad7e', '#6f5067'],
    tasks: ['#fffcf2', '#6d9b6a', '#c7805f', '#9d947b'],
    nodes: ['#94a4c8', '#527ebd', '#d89279', '#9174c1', '#f6faff', '#fff2e7'],
    monitor: ['#f6fbed', '#a9cea9', '#518b59', '#dbab62'],
    extensions: ['#b86864', '#fff6e9', '#d89587'],
    users: ['#d2997b', '#edc7ab', '#9675bb', '#735298'],
    settings: ['#60768a', '#f3f8fb', '#a2bacb'],
    docker: ['#467eaa', '#77b1c8', '#dcad6c', '#36648b', '#f2f9fb'],
    system: ['#a89d92', '#5d9e8d', '#c78368', '#8d85b7'],
  }, { '#35414f': '#52677c', '#8392a6': '#90a4b9', '#fbfcff': '#fafcff', '#83bcff': '#a8d4e8', '#c3cfdf': '#d0ddea' })

  const chromatic = palette('unify-color', {
    launcher: '#577f96', instances: '#4979ba', backups: '#388e75', files: '#c48f45',
    editor: '#9877ab', terminal: '#64788c', tasks: '#738f83', nodes: '#797db2',
    monitor: '#4b8b74', extensions: '#c87370', users: '#977eb4', settings: '#6f8395',
    docker: '#458ca8', system: '#aa8c6d',
  }, {
    launcher: ['#bde4db', '#f2cb92', '#c2d5f0', '#ecc2d5'],
    instances: ['#a9caec', '#dbe9f7', '#fff8db', '#d5f6b2', '#5287c3'],
    backups: ['#f4fff6', '#c5ead8', '#367d66'],
    files: ['#e8bf7c', '#fffbec', '#ffe1a0'],
    editor: ['#fffafb', '#e3c9e6', '#ae8baa', '#6f4985', '#ecc495', '#624f61'],
    tasks: ['#fbf9ec', '#559d7c', '#d58968', '#9b9f8d'],
    nodes: ['#bac7ec', '#c4ddfa', '#ffccac', '#d9c7f4', '#6381b2', '#c77b60'],
    monitor: ['#f4fff4', '#c6e4c8', '#32836a', '#ecb871'],
    extensions: ['#f9c6b8', '#fff6e5', '#c8716b'],
    users: ['#f7c19e', '#ffe0bb', '#e2d4f1', '#c8b3e3'],
    settings: ['#e9f0f6', '#617c92', '#b9d1e2'],
    docker: ['#c9eafa', '#ebfbfc', '#ffd79d', '#d3e7f3', '#3f79a2'],
    system: ['#e3d6c7', '#b8e4d4', '#ffd0b3', '#e3d6ff'],
  }, { '#35414f': '#e1eaf3', '#8392a6': '#adbed0', '#fbfcff': '#435b75', '#83bcff': '#487cac', '#c3cfdf': '#7f96af' })

  const smoke = palette('unify-smoke', {
    launcher: '#35454e', instances: '#30445f', backups: '#2f4c44', files: '#51483b',
    editor: '#4a3f50', terminal: '#374553', tasks: '#444b42', nodes: '#3e4359',
    monitor: '#304b41', extensions: '#563f41', users: '#463d52', settings: '#374650',
    docker: '#304b56', system: '#4b443f',
  }, {
    launcher: ['#85d8d5', '#ebbc83', '#9db8e9', '#dda6c8'],
    instances: ['#92bce7', '#618fbd', '#fff4d5', '#d2ebb3', '#ecf5ff'],
    backups: ['#d7eee2', '#8acdb1', '#428d74'],
    files: ['#ce9e53', '#fff2d4', '#e9be74'],
    editor: ['#f5e9f1', '#c4a9c3', '#9a709a', '#bc8ac2', '#e6bc8e', '#765768'],
    tasks: ['#e9eddf', '#609879', '#ce927a', '#86988e'],
    nodes: ['#8996b7', '#8eafe7', '#e7af91', '#b39bdb', '#f5f7ff', '#fff3e4'],
    monitor: ['#d9eee0', '#94c7a3', '#508b65', '#e8b781'],
    extensions: ['#dd9b95', '#fff1dd', '#b2686c'],
    users: ['#e1ac90', '#f2cfb3', '#b89ed8', '#997ec0'],
    settings: ['#a9bed0', '#506779', '#cfdfeb'],
    docker: ['#82bbd6', '#bce1e6', '#e8bd80', '#629aba', '#eaf8fb'],
    system: ['#969ea7', '#97d0bd', '#e7b098', '#b3a8dd'],
  }, { '#35414f': '#8fa5b8', '#8392a6': '#b5c5d4', '#fbfcff': '#fbfeff', '#83bcff': '#c1e2f5', '#c3cfdf': '#ddeaf3' })
  return Object.fromEntries([porcelain, tonal, chromatic, smoke].map(item => [item.id, item]))
}

export const harmonyPalettes = /* @__PURE__ */ createHarmonyPalettes()
export function harmonyColorway(id: string | null): IconPalette | undefined {
  return id && Object.prototype.hasOwnProperty.call(harmonyPalettes, id) ? harmonyPalettes[id] : undefined
}
