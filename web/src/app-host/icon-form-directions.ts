import type { IconLayer } from './classic-icon-layers'
import type { IconPalette } from './icon-ice-palette'

type Role = IconLayer['role']
type Artwork = Record<string, IconLayer[]>
const r = (x: number, y: number, width: number, height: number, rx: number, fill: string, role: Role = 'surface'): IconLayer => ({ tag: 'rect', attrs: { x: `${x}`, y: `${y}`, width: `${width}`, height: `${height}`, rx: `${rx}`, fill }, role })
const p = (d: string, fill: string, role: Role = 'surface'): IconLayer => ({ tag: 'path', attrs: { d, fill }, role })
const c = (cx: number, cy: number, radius: number, fill: string, role: Role = 'surface'): IconLayer => ({ tag: 'circle', attrs: { cx: `${cx}`, cy: `${cy}`, r: `${radius}`, fill }, role })
const l = (d: string, stroke: string, width = 3, role: Role = 'ink'): IconLayer => ({ tag: 'path', attrs: { d, stroke, 'stroke-width': `${width}`, 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, role })
const turn = (layer: IconLayer, transform: string): IconLayer => ({ ...layer, attrs: { ...layer.attrs, transform } })
const hole = (d: string, fill: string): IconLayer => ({ tag: 'path', attrs: { d, fill, 'fill-rule': 'evenodd' }, role: 'surface' })

// Shared application colors and backing: the review compares silhouette,
// composition and metaphor, rather than another wallpaper/palette variation.
const C = {
  white: '#fffdf9', blue: '#528ddd', blueLight: '#bddcfa', blueDeep: '#3165ad',
  green: '#389e80', greenLight: '#bfe7d4', greenDeep: '#22755e',
  gold: '#edb14f', goldLight: '#ffe6ac', goldDeep: '#c38a35',
  pink: '#f0bdcf', rose: '#d582a5', purple: '#a47ac6', purpleDeep: '#76599a',
  lavender: '#d8c6ee', coral: '#df8078', coralLight: '#ffd8c4',
  cyan: '#5badcf', cyanLight: '#b6e3ec', cyanDeep: '#3d829e',
  slate: '#8296aa', silver: '#c4d1dc', pale: '#e3eaf0', dark: '#42566c', ink: '#607c95',
}
const backing = {
  launcher: '#f7f8f7', instances: '#f3f6fa', backups: '#f4f8f5', files: '#fcf7ef',
  editor: '#faf5f9', terminal: '#f3f5f7', tasks: '#fbf8f2', nodes: '#f5f6fb',
  monitor: '#f4f8f4', extensions: '#fcf4f2', users: '#f8f4fa', settings: '#f2f5f7',
  docker: '#f1f8fa', system: '#faf6ef',
}

function layered(): Artwork {
  return {
    launcher: [
      turn(r(25, 22, 39, 42, 11, C.cyanLight), 'rotate(-12 44 43)'),
      turn(r(36, 30, 37, 42, 11, C.pink), 'rotate(10 55 51)'),
      r(22, 40, 36, 35, 10, C.blue), r(29, 48, 8, 8, 2, C.white, 'ink'), r(42, 48, 8, 8, 2, C.goldLight, 'ink'), r(29, 61, 21, 5, 2, C.blueLight, 'ink'),
    ],
    instances: [
      r(20, 22, 40, 44, 8, C.blueLight), r(25, 29, 28, 5, 2, C.white, 'ink'),
      r(34, 32, 42, 43, 9, C.blue),
      ...[43, 54, 65].flatMap(y => [c(43, y, 2.3, y === 65 ? C.greenLight : C.white, 'ink'), r(50, y - 1.5, 16, 3, 1.5, C.blueLight, 'ink')]),
    ],
    backups: [
      turn(r(22, 23, 43, 48, 9, C.greenLight), 'rotate(-12 43 47)'),
      c(53, 49, 25, C.white, 'paper'), l('M38 31a24 24 0 1 1-8 27', C.green, 5, 'surface'),
      p('M28 27v15l14-5Z', C.green), l('M53 35v15l10 6', C.green, 3.7),
    ],
    files: [
      turn(p('M23 28h18l7 7h22a6 6 0 0 1 6 6v24H23Z', C.goldDeep), 'rotate(-10 49 47)'),
      r(29, 32, 41, 30, 4, C.white, 'paper'), r(35, 37, 25, 3, 1.5, C.goldLight, 'ink'),
      p('M18 43h60v25a8 8 0 0 1-8 8H26a8 8 0 0 1-8-8Z', C.gold),
      r(60, 63, 9, 3, 1.5, C.white, 'ink'),
    ],
    editor: [
      turn(r(25, 23, 41, 49, 6, C.pink), 'rotate(-10 45 48)'), r(32, 25, 39, 50, 6, C.white, 'paper'),
      r(38, 33, 18, 4, 1.5, C.rose, 'ink'), l('M38 44h22M38 51h17M38 58h10', C.pink, 2.6),
      p('M51 72l5-15 19-21a4 4 0 0 1 6 6L62 64Z', C.purple), p('m51 72 5-15 6 7Z', C.goldLight, 'ink'),
    ],
    terminal: [
      r(29, 22, 48, 38, 7, C.blueLight), r(36, 29, 17, 3, 1.5, C.white, 'ink'),
      r(18, 35, 56, 39, 8, C.dark),
      l('M27 44h12', C.silver, 2), c(65, 44, 1.5, C.blueLight, 'ink'),
      l('m28 53 6 5-6 5', C.white, 3.5), r(42, 60, 12, 3.5, 1.3, C.cyanLight, 'ink'),
    ],
    tasks: [
      turn(r(25, 20, 41, 51, 7, C.goldLight), 'rotate(-10 45 45)'), r(33, 27, 40, 49, 7, C.white, 'paper'),
      c(43, 41, 5, C.green), l('m40 41 2 2 4-5', C.white, 1.8),
      l('M55 41h10M55 55h10M43 67h22', C.slate, 2.4), c(43, 55, 4, C.coral, 'ink'),
    ],
    nodes: [
      l('M34 36h28v24H38', C.silver, 5), r(20, 21, 29, 28, 8, C.blue), r(52, 44, 27, 29, 8, C.purple),
      r(22, 56, 24, 21, 7, C.coralLight), c(34, 35, 5, C.blueLight), c(65, 58, 5, C.lavender), c(34, 66, 4, C.coral, 'ink'),
    ],
    monitor: [
      r(33, 22, 40, 38, 8, C.greenLight), r(42, 39, 5, 12, 2, C.white, 'ink'), r(51, 31, 5, 20, 2, C.white, 'ink'), r(60, 36, 5, 15, 2, C.white, 'ink'),
      r(19, 37, 55, 38, 8, C.white, 'paper'), p('M26 65V59l9-7 10 8 13-13 9 6v12Z', C.greenLight),
      l('m26 59 9-7 10 8 13-13 9 6', C.green, 3), c(58, 47, 3.3, C.gold, 'ink'),
    ],
    extensions: [
      l('M41 38V28a12 12 0 0 1 24 0v10', C.coral, 4, 'surface'),
      turn(r(30, 34, 42, 36, 7, C.coralLight), 'rotate(10 51 52)'),
      l('M29 42V31a11 11 0 0 1 22 0v11', C.rose, 4, 'surface'), r(19, 40, 45, 36, 8, C.coral),
      r(31, 50, 10, 10, 2.5, C.white), r(44, 50, 10, 10, 2.5, C.coralLight),
    ],
    users: [
      turn(r(36, 22, 36, 47, 8, C.lavender), 'rotate(9 54 45)'), c(55, 35, 7, C.purple),
      r(22, 34, 40, 43, 9, C.white, 'paper'), c(42, 47, 8, C.purple), p('M27 69a15 15 0 0 1 30 0v3H27Z', C.purpleDeep),
    ],
    settings: [
      ...gear(40, 42, 24, 19, 8, C.slate), c(40, 42, 12, C.white, 'paper'), c(40, 42, 5, C.blueLight),
      ...gear(64, 64, 15, 11, 6, C.blueLight), c(64, 64, 5, C.blueDeep),
    ],
    docker: [
      r(34, 22, 42, 31, 6, C.cyanLight), l('M43 28v19m9-19v19m9-19v19', C.white, 2),
      r(20, 42, 50, 31, 6, C.cyan), l('M29 49v17m10-17v17m10-17v17m10-17v17', C.cyanLight, 2.5),
      r(30, 76, 35, 3, 1.5, C.cyanDeep, 'ink'),
    ],
    system: [
      turn(r(27, 21, 43, 47, 8, C.goldLight), 'rotate(10 48 44)'), r(21, 34, 46, 43, 8, C.white, 'paper'),
      l('M31 44v23m13-23v23m13-23v23', C.silver, 2.5), r(27, 48, 8, 9, 3, C.green), r(40, 59, 8, 9, 3, C.coral), r(53, 44, 8, 9, 3, C.purple),
    ],
  }
}

function bold(): Artwork {
  return {
    launcher: [
      r(20, 20, 25, 25, 9, C.cyan), c(62, 33, 13, C.gold), p('M21 54a8 8 0 0 1 8-8h16v30H29a8 8 0 0 1-8-8Z', C.rose), r(51, 51, 25, 25, 7, C.blue),
    ],
    instances: [
      r(18, 24, 60, 49, 13, C.blue), r(26, 33, 44, 12, 5, C.blueLight), r(26, 52, 44, 12, 5, C.white, 'paper'),
      c(33, 39, 2.5, C.blueDeep, 'ink'), c(33, 58, 2.5, C.green, 'ink'), r(44, 37.5, 18, 3, 1.5, C.blue, 'ink'), r(44, 56.5, 18, 3, 1.5, C.blue, 'ink'),
    ],
    backups: [
      c(49, 49, 26, C.greenLight), l('M27 31a29 29 0 1 1-6 30', C.green, 8, 'surface'), p('M16 25v20l19-8Z', C.green),
      l('M49 33v17l12 7', C.greenDeep, 4.5),
    ],
    files: [
      p('M18 34a9 9 0 0 1 9-9h14l7 9h21a9 9 0 0 1 9 9v20a11 11 0 0 1-11 11H29a11 11 0 0 1-11-11Z', C.goldDeep),
      r(28, 34, 37, 21, 4, C.white, 'paper'), p('M18 45h60v18a11 11 0 0 1-11 11H29a11 11 0 0 1-11-11Z', C.gold), r(58, 61, 11, 4, 2, C.white, 'ink'),
    ],
    editor: [
      r(20, 24, 43, 50, 9, C.pink), r(29, 31, 25, 35, 4, C.white, 'paper'), l('M35 39h12M35 47h8', C.rose, 2.8),
      p('M47 73l5-20 17-24a6 6 0 0 1 10 7L62 60Z', C.purple), p('m47 73 5-20 10 7Z', C.goldLight, 'ink'), p('m47 73 2-8 5 4Z', C.purpleDeep, 'ink'),
    ],
    terminal: [
      r(17, 26, 62, 46, 12, C.dark), r(27, 33, 15, 3, 1.5, C.silver, 'ink'), c(69, 34.5, 2, C.cyanLight, 'ink'),
      p('M28 44a2 2 0 0 1 3-.2L40 51a2 2 0 0 1 0 3l-9 7a2 2 0 0 1-3-3l7-5.5-7-5.5a2 2 0 0 1 0-3Z', C.white, 'ink'),
      r(49, 46, 17, 4, 2, C.silver, 'ink'), r(49, 57, 8, 7, 2, C.cyanLight, 'ink'),
    ],
    tasks: [
      r(23, 19, 51, 59, 11, C.white, 'paper'), r(30, 19, 37, 11, 5, C.gold),
      c(38, 45, 9, C.green), l('m34 45 3 3 6-7', C.white, 2.7), r(53, 42, 13, 5, 2.5, C.greenLight, 'ink'),
      c(38, 65, 4, C.coral), r(49, 62.5, 17, 5, 2.5, C.silver, 'ink'),
    ],
    nodes: [
      l('M47 49 28 31m19 18 23-10M47 49l4 22', C.silver, 7), c(27, 29, 13, C.blue), c(70, 35, 11, C.coral), c(51, 71, 11, C.lavender),
      c(47, 48, 14, C.purple), c(47, 48, 6, C.white, 'paper'),
    ],
    monitor: [
      c(48, 48, 29, C.greenLight), hole('M22 62a29 29 0 1 1 52 0l-8-5a20 20 0 1 0-36 0Z', C.green),
      p('M42 54 66 29 54 59Z', C.coral), c(48, 55, 6, C.white, 'paper'), r(33, 70, 30, 4, 2, C.greenDeep, 'ink'),
    ],
    extensions: [
      l('M35 35v-7a13 13 0 0 1 26 0v7', C.coral, 5, 'surface'), r(20, 33, 56, 43, 11, C.coral),
      r(35, 45, 10, 10, 3, C.white), r(51, 45, 10, 10, 3, C.coralLight),
      r(35, 60, 10, 10, 3, C.goldLight), r(51, 60, 10, 10, 3, C.white),
    ],
    users: [
      r(22, 21, 44, 55, 15, C.purple), c(44, 38, 10, C.white, 'paper'), p('M28 66a16 16 0 0 1 32 0v4H28Z', C.lavender),
      c(70, 51, 9, C.coralLight), p('M59 75v-9a11 11 0 0 1 22 0v9Z', C.coral),
    ],
    settings: [
      ...gear(48, 48, 31, 25, 10, C.slate), c(48, 48, 17, C.white, 'paper'), c(48, 48, 9, C.blueLight), c(48, 48, 3, C.blueDeep, 'ink'),
    ],
    docker: [
      r(24, 24, 20, 21, 5, C.cyanLight), r(49, 24, 22, 21, 5, C.cyan),
      p('M17 49h62v6a22 22 0 0 1-22 22H39a22 22 0 0 1-22-22Z', C.cyanDeep),
      r(29, 57, 38, 7, 3, C.cyanLight, 'ink'), r(43, 19, 10, 13, 3, C.gold),
    ],
    system: [
      l('M28 23v50m20-50v50m20-50v50', C.silver, 6), r(19, 32, 18, 20, 7, C.green), r(39, 52, 18, 20, 7, C.coral), r(59, 24, 18, 20, 7, C.purple),
      l('M24 41h8m12 20h8m12-28h8', C.white, 2.4),
    ],
  }
}

function studio(): Artwork {
  return {
    launcher: [
      r(20, 23, 24, 51, 8, C.cyan), r(49, 23, 27, 22, 7, C.gold), r(49, 50, 27, 24, 7, C.rose),
      r(27, 33, 10, 3, 1.5, C.white, 'ink'), c(32, 62, 4, C.cyanLight),
    ],
    instances: [
      r(29, 18, 39, 60, 8, C.silver), r(33, 22, 31, 49, 5, C.blueDeep),
      ...[28, 41, 54].flatMap(y => [r(37, y, 23, 9, 2.5, C.blueLight), c(42, y + 4.5, 1.6, C.white, 'ink'), r(48, y + 3.5, 8, 2, 1, C.blueDeep, 'ink')]),
      c(48.5, 74, 1.6, C.green, 'ink'),
    ],
    backups: [
      r(25, 23, 46, 50, 8, C.greenLight), p('M25 34h46v-3a8 8 0 0 0-8-8H33a8 8 0 0 0-8 8Z', C.green),
      r(33, 42, 25, 3, 1.5, C.white, 'ink'), r(33, 49, 16, 3, 1.5, C.white, 'ink'),
      c(62, 63, 16, C.white, 'paper'), l('M51 54a15 15 0 1 1-3 17', C.green, 3.5), p('M47 49v10l10-3Z', C.green), l('M62 53v11l6 4', C.greenDeep, 2.6),
    ],
    files: [
      p('M21 30a6 6 0 0 1 6-6h17l7 9h21a6 6 0 0 1 6 6v28H21Z', C.goldDeep),
      r(27, 34, 43, 25, 3, C.goldLight, 'paper'), r(30, 40, 40, 25, 3, C.white, 'paper'),
      p('M18 45h61l-4 24a6 6 0 0 1-6 5H27a6 6 0 0 1-6-5Z', C.gold), r(39, 54, 25, 7, 2.5, C.goldLight, 'ink'),
    ],
    editor: [
      r(24, 19, 42, 58, 6, C.rose), r(30, 21, 33, 53, 4, C.white, 'paper'),
      l('M26 30h5m-5 11h5m-5 11h5m-5 11h5', C.purpleDeep, 2.5), l('M38 33h17M38 43h17M38 53h9', C.pink, 2.6),
      turn(r(66, 30, 7, 41, 3, C.purple), 'rotate(24 69 51)'), turn(p('m66 71 3.5 9 3.5-9Z', C.goldLight, 'ink'), 'rotate(24 69 51)'),
    ],
    terminal: [
      r(19, 24, 58, 43, 7, C.silver), r(23, 28, 50, 34, 4, C.dark),
      l('m30 38 7 6-7 6', C.white, 3), r(44, 46, 10, 3.3, 1.2, C.cyanLight, 'ink'), l('M30 55h20m5 0h9', C.slate, 2),
      p('M41 66h14l3 9H38Z', C.slate), r(31, 74, 35, 4, 2, C.silver),
    ],
    tasks: [
      r(25, 22, 47, 55, 6, C.gold), r(29, 28, 39, 44, 4, C.white, 'paper'), r(38, 19, 21, 12, 4, C.silver),
      r(35, 38, 9, 9, 3, C.green), l('m37 42 2 2 3-4', C.white, 1.5), l('M51 42h10M51 57h10', C.slate, 2.6),
      r(35, 53, 9, 9, 3, C.coralLight), c(39.5, 57.5, 1.8, C.coral, 'ink'),
    ],
    nodes: [
      l('M30 53v14h36V47M48 67v9', C.silver, 3.5),
      r(18, 25, 25, 33, 6, C.blue), r(53, 20, 25, 33, 6, C.purple),
      ...[33, 42].map(y => r(24, y, 13, 3, 1.5, C.blueLight, 'ink')),
      ...[28, 37].map(y => r(59, y, 13, 3, 1.5, C.lavender, 'ink')), c(48, 73, 7, C.coral),
    ],
    monitor: [
      r(20, 23, 57, 48, 8, C.green), r(25, 28, 47, 33, 4, C.white, 'paper'),
      l('M29 55h38', C.greenLight, 1.6), p('M30 55V46l8-6 8 8 9-15 11 9v13Z', C.greenLight), l('m30 46 8-6 8 8 9-15 11 9', C.greenDeep, 2.5),
      c(30, 66, 2, C.goldLight, 'ink'), r(58, 64.5, 11, 3, 1.5, C.greenLight, 'ink'), r(30, 73, 38, 4, 2, C.silver),
    ],
    extensions: [
      p('M21 38 48 28 75 38v33a5 5 0 0 1-5 5H26a5 5 0 0 1-5-5Z', C.coral),
      p('M21 38 48 48 75 38 67 26 48 32 29 26Z', C.coralLight), p('M48 32v16L21 38l8-12Z', C.white, 'paper'),
      r(38, 53, 20, 16, 4, C.white), r(43, 57, 5, 5, 1, C.coral, 'ink'), r(51, 57, 3, 8, 1, C.gold, 'ink'),
    ],
    users: [
      turn(r(38, 20, 36, 50, 6, C.coralLight), 'rotate(9 56 45)'), c(58, 37, 7, C.coral),
      r(22, 29, 38, 48, 7, C.white, 'paper'), r(32, 25, 18, 7, 3, C.purple), c(41, 46, 8, C.purple),
      p('M28 67a13 13 0 0 1 26 0v2H28Z', C.lavender),
    ],
    settings: [
      r(21, 22, 54, 53, 9, C.silver), r(26, 27, 44, 42, 6, C.white, 'paper'),
      c(40, 42, 10, C.slate), c(58, 57, 8, C.blueLight), l('M40 42v-5m18 20 3-3', C.white, 2.6),
      l('M30 59h13M55 36h8', C.silver, 3), c(65, 31, 1.7, C.green, 'ink'),
    ],
    docker: [
      r(19, 33, 58, 37, 5, C.cyan), r(23, 28, 50, 7, 2, C.cyanLight),
      ...[28, 39, 50, 61].map(x => r(x, 41, 5, 21, 1.5, C.cyanLight, 'ink')),
      r(24, 70, 10, 5, 1, C.cyanDeep, 'ink'), r(62, 70, 10, 5, 1, C.cyanDeep, 'ink'),
    ],
    system: [
      r(20, 24, 56, 48, 8, C.silver), r(26, 30, 44, 34, 5, C.white, 'paper'),
      l('M33 38v19m15-19v19m15-19v19', C.silver, 2.5), r(29, 39, 8, 8, 3, C.green), r(44, 49, 8, 8, 3, C.coral), r(59, 34, 8, 8, 3, C.purple),
      r(41, 68, 14, 2, 1, C.slate, 'ink'),
    ],
  }
}

function folded(): Artwork {
  return {
    launcher: [
      p('m20 28 27-10v29L20 57Z', C.cyan), p('m47 18 28 10v29L47 47Z', C.gold),
      p('m20 57 27-10v30L20 67Z', C.purple), p('m47 47 28 10v10L47 77Z', C.rose),
    ],
    instances: [
      p('m20 29 44-9 12 13-44 10Z', C.blueLight), p('M20 29 32 43v13L20 42Z', C.blueDeep), p('m32 43 44-10v13L32 56Z', C.blue),
      p('m20 50 44-9 12 13-44 10Z', C.blueLight), p('M20 50 32 64v13L20 63Z', C.blueDeep), p('m32 64 44-10v13L32 77Z', C.blue),
      l('m43 47 18-4m-18 26 18-4', C.white, 2.6),
    ],
    backups: [
      p('M28 24h33l12 12v35H28Z', C.greenLight), p('M61 24v12h12Z', C.white, 'paper'),
      l('M28 40a22 22 0 1 1-5 25', C.green, 6, 'surface'), p('M19 34v17l16-6Z', C.green),
      l('M45 44v14l10 5', C.greenDeep, 3.5),
    ],
    files: [
      p('M20 29h23l9 8h23v34H20Z', C.goldDeep), p('M29 30h30l9 10v20H29Z', C.white, 'paper'), p('M59 30v10h9Z', C.goldLight),
      p('M17 45h62l-5 29H22Z', C.gold), p('M17 45 48 61 22 74Z', C.goldLight), p('m79 45-31 16 26 13Z', C.gold),
    ],
    editor: [
      p('M25 20h31l15 15v42H25Z', C.white, 'paper'), p('M56 20v15h15Z', C.pink), p('M25 63 40 77H25Z', C.rose),
      l('M34 40h23M34 48h17M34 56h11', C.pink, 2.5),
      p('m45 71 9-23 20-17 6 7-20 17Z', C.purple), p('m45 71 9-23 6 7Z', C.goldLight, 'ink'),
    ],
    terminal: [
      p('M20 25h42l14 14v34H20Z', C.dark), p('M62 25v14h14Z', C.slate), p('M20 65h56v8H20Z', C.cyan),
      l('m30 39 7 6-7 6', C.white, 3.2), l('M45 44h16M30 58h23', C.cyanLight, 2.5),
    ],
    tasks: [
      p('M26 21h32l13 13v42H26Z', C.white, 'paper'), p('M58 21v13h13Z', C.goldLight),
      p('m17 46 11-9 15 15 23-25 10 10-33 34Z', C.green), p('m17 46 11-9 15 15-1 19Z', C.greenLight),
      r(49, 69, 15, 3, 1.5, C.silver, 'ink'),
    ],
    nodes: [
      l('M31 36 67 42 48 70Z', C.silver, 3.5),
      p('m19 24 13-7 13 7v18l-13 7-13-7Z', C.blue), p('m32 17 13 7v18l-13 7Z', C.blueLight),
      p('m58 30 11-6 11 6v16l-11 6-11-6Z', C.coral), p('m69 24 11 6v16l-11 6Z', C.coralLight),
      p('m36 62 13-7 13 7v15l-13 7-13-7Z', C.purple), p('m49 55 13 7v15l-13 7Z', C.lavender),
    ],
    monitor: [
      p('M21 73V25h39l15 15v33Z', C.white, 'paper'), p('M60 25v15h15Z', C.greenLight),
      p('m24 65 14-19 14 9 20-22v40H24Z', C.greenLight), p('m24 73 23-24 10 8 15-9v25Z', C.green),
      p('m57 57 15-9v25Z', C.greenDeep), c(72, 33, 3.5, C.gold, 'ink'),
    ],
    extensions: [
      l('M36 37V27a12 12 0 0 1 24 0v10', C.coral, 4, 'surface'), p('M23 35h49l5 41H18Z', C.coral),
      p('M23 35 48 56 18 76Z', C.coralLight), p('m72 35-24 21 29 20Z', C.rose),
      r(42, 44, 12, 12, 3, C.white),
    ],
    users: [
      c(62, 34, 10, C.coralLight), p('m51 47 13-4 14 10v22H48Z', C.coral), p('m64 43 14 10v22H64Z', C.coralLight),
      c(35, 32, 12, C.purple), p('m20 49 15-5 18 13v20H17Z', C.purpleDeep), p('m35 44 18 13v20H35Z', C.lavender),
    ],
    settings: [
      ...gear(48, 48, 29, 23, 8, C.slate, C.silver),
      c(48, 48, 14, C.white, 'paper'), c(48, 48, 7, C.blueLight),
    ],
    docker: [
      p('m20 31 28-13 29 13-29 14Z', C.cyanLight), p('M20 31 48 45v33L20 64Z', C.cyan), p('m48 45 29-14v33L48 78Z', C.cyanDeep),
      l('m28 44 0 14m10-9v14m20-15v15m10-20v15', C.cyanLight, 2.5),
    ],
    system: [
      p('m20 25 19-6v54l-19 6Z', C.greenLight), p('m39 19 19 6v54l-19-6Z', C.coralLight), p('m58 25 19-6v54l-19 6Z', C.lavender),
      l('M29 34v32m19-33v31m19-32v31', C.white, 3), r(24, 43, 10, 10, 3, C.green), r(43, 54, 10, 10, 3, C.coral), r(62, 34, 10, 10, 3, C.purple),
    ],
  }
}

// Rounded radial teeth, not a downloaded toolbar glyph.
function gear(cx: number, cy: number, outer: number, inner: number, teeth: number, fill: string, secondFace?: string): IconLayer[] {
  const points: string[] = []
  for (let i = 0; i < teeth * 4; i++) {
    const angle = i * Math.PI * 2 / (teeth * 4) - Math.PI / 2
    const radius = i % 4 === 1 || i % 4 === 2 ? outer : inner
    points.push(`${(cx + Math.cos(angle) * radius).toFixed(2)} ${(cy + Math.sin(angle) * radius).toFixed(2)}`)
  }
  const full: IconLayer = { tag: 'path', attrs: { d: `M${points.join('L')}Z`, fill, stroke: fill, 'stroke-width': '1.5', 'stroke-linejoin': 'round' }, role: 'surface' }
  // The color split is printed on one object; a second optical shell would
  // double the shared perimeter and create rippled, concentric highlights.
  return secondFace ? [full, p(`M${points.slice(0, teeth * 2 + 1).join('L')}Z`, secondFace, 'ink')] : [full]
}

function createDirections(): Record<string, IconPalette> {
  const directions = { 'form-layered': layered(), 'form-bold': bold(), 'form-studio': studio(), 'form-folded': folded() }
  return Object.fromEntries(Object.entries(directions).map(([id, artwork]) => {
    for (const kind of Object.keys(backing)) if (!artwork[kind]?.length) throw Error(`Missing ${id} artwork: ${kind}`)
    return [id, { id, backgrounds: { ...backing }, colors: {}, artwork }]
  }))
}
export const formDirections = /* @__PURE__ */ createDirections()
export function formDirection(id: string | null): IconPalette | undefined {
  return id && Object.prototype.hasOwnProperty.call(formDirections, id) ? formDirections[id] : undefined
}
