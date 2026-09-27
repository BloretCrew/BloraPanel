import { iconColorways } from './icon-colorways'
import type { IconPalette } from './icon-ice-palette'
import type { IconLayer } from './classic-icon-layers'

type Role = IconLayer['role']
type TerminalInk = { backing: string; panel: string; header: string; symbol: string; cursor: string; detail: string; rear: string }
type MonitorInk = { backing: string; panel: string; area: string; trace: string; point: string }
type Direction = 'silver' | 'layers' | 'paper'

const rect = (x: number, y: number, width: number, height: number, rx: number, fill: string, role: Role = 'surface'): IconLayer => ({
  tag: 'rect', attrs: { x: String(x), y: String(y), width: String(width), height: String(height), rx: String(rx), fill }, role,
})
const path = (d: string, fill: string, role: Role = 'surface'): IconLayer => ({ tag: 'path', attrs: { d, fill }, role })
const line = (d: string, stroke: string, width: number): IconLayer => ({
  tag: 'path', attrs: { d, stroke, 'stroke-width': String(width), 'stroke-linecap': 'round', 'stroke-linejoin': 'round' }, role: 'ink',
})
const dot = (cx: number, cy: number, r: number, fill: string): IconLayer => ({
  tag: 'circle', attrs: { cx: String(cx), cy: String(cy), r: String(r), fill }, role: 'detail',
})

function terminalArtwork(direction: Direction, c: TerminalInk): IconLayer[] {
  if (direction === 'silver') return [
    rect(17, 24, 62, 49, 8, c.panel),
    // A flat header, fitted to the pane rather than an extra inset rectangle.
    path('M25 24h46a8 8 0 0 1 8 8v5H17v-5a8 8 0 0 1 8-8Z', c.header, 'ink'),
    line('M25 31h12', c.detail, 2.2), dot(71, 31, 1.5, c.detail),
    line('m29 45 7 6-7 6', c.symbol, 3.8),
    rect(42, 53, 11, 3.8, 1.4, c.cursor, 'ink'),
    line('M28 64h16m5 0h14', c.detail, 2),
  ]
  if (direction === 'layers') return [
    // Offset glass panes distinguish the app silhouette from a toolbar glyph.
    rect(29, 20, 49, 41, 7, c.rear),
    line('M37 27h21', c.detail, 2),
    rect(17, 32, 55, 43, 7, c.panel),
    path('M24 32h41a7 7 0 0 1 7 7v3H17v-3a7 7 0 0 1 7-7Z', c.header, 'ink'),
    dot(24, 37, 1.4, c.detail), line('M29 37h10', c.detail, 1.8),
    // A filled, carefully proportioned prompt is crisp at Dock dimensions.
    path('M27 48a1.6 1.6 0 0 1 2.3-.1l7.7 6.5a1.7 1.7 0 0 1 0 2.6l-7.7 6.5a1.6 1.6 0 0 1-2.2-2.4l6.2-5.4-6.2-5.3A1.6 1.6 0 0 1 27 48Z', c.symbol, 'ink'),
    rect(43, 61, 13, 3.4, 1.3, c.cursor, 'ink'),
  ]
  return [
    rect(20, 20, 56, 56, 8, c.panel),
    path('M28 20h40a8 8 0 0 1 8 8v7H20v-7a8 8 0 0 1 8-8Z', c.header, 'ink'),
    line('M28 28h17', c.detail, 2.3), dot(68, 28, 1.6, c.detail),
    // The command page uses a block cursor and two output lines, not a large >_.
    line('m29 43 6 5-6 5', c.symbol, 3.3),
    rect(43, 41, 7, 13, 1.3, c.cursor, 'ink'),
    line('M29 62h13m5 0h16M29 69h25', c.detail, 2.2),
  ]
}

const paints: Record<'spectrum' | 'pastel', Record<Direction, { terminal: TerminalInk; monitor: MonitorInk }>> = {
  spectrum: {
    silver: {
      terminal: { backing: '#edf0f5', panel: '#35414f', header: '#8392a6', symbol: '#fbfcff', cursor: '#83bcff', detail: '#c3cfdf', rear: '#b8c6df' },
      monitor: { backing: '#daf5e8', panel: '#fbfff9', area: '#cce8ce', trace: '#299b67', point: '#f5a96c' },
    },
    layers: {
      terminal: { backing: '#ede5fb', panel: '#8663c5', header: '#69469f', symbol: '#fffaff', cursor: '#ffc69b', detail: '#d9c9f0', rear: '#c5b0eb' },
      monitor: { backing: '#dfedff', panel: '#f8fcff', area: '#bdd7fa', trace: '#4b81d7', point: '#f5a757' },
    },
    paper: {
      terminal: { backing: '#e7effc', panel: '#ffffff', header: '#638bdd', symbol: '#395683', cursor: '#6d8fde', detail: '#9bb1d2', rear: '#cfddf4' },
      monitor: { backing: '#fff0d9', panel: '#fffaf1', area: '#f7ddb0', trace: '#c98935', point: '#e67e61' },
    },
  },
  pastel: {
    silver: {
      terminal: { backing: '#ecf0f5', panel: '#5c6a7b', header: '#a3b1c4', symbol: '#fffefd', cursor: '#bdccf1', detail: '#d5deea', rear: '#c9d4e7' },
      monitor: { backing: '#e5f5e7', panel: '#fffff9', area: '#d7eacc', trace: '#6caa72', point: '#edb884' },
    },
    layers: {
      terminal: { backing: '#efe8fc', panel: '#aa8cd3', header: '#886bb5', symbol: '#fffaff', cursor: '#ffe0b9', detail: '#e7d9f5', rear: '#d3c2ec' },
      monitor: { backing: '#e5efff', panel: '#fbfdff', area: '#cbdcf4', trace: '#7b9fda', point: '#edb778' },
    },
    paper: {
      terminal: { backing: '#eaf0fa', panel: '#fffffc', header: '#98b1dc', symbol: '#6580ac', cursor: '#9bacdc', detail: '#b4c4dc', rear: '#d8e3f1' },
      monitor: { backing: '#fff0dd', panel: '#fffdf5', area: '#f5e3c4', trace: '#c4a064', point: '#dfa081' },
    },
  },
}

function createUtilityVariants(): Record<string, IconPalette> {
  const variants: Record<string, IconPalette> = {}
  for (const base of ['spectrum', 'pastel'] as const) {
    for (const direction of ['silver', 'layers', 'paper'] as const) {
      const { terminal, monitor } = paints[base][direction]
      const original = iconColorways[base]
      const id = `${base}-${direction}`
      variants[id] = {
        ...original, id,
        backgrounds: { ...original.backgrounds, terminal: terminal.backing, monitor: monitor.backing },
        colors: {
          ...original.colors, terminal: {},
          monitor: { '#f7fcf6': monitor.panel, '#c0dac9': monitor.area, '#458678': monitor.trace, '#d9a175': monitor.point },
        },
        artwork: { terminal: terminalArtwork(direction, terminal) },
      }
    }
  }
  return variants
}

// This factory only allocates local preview data; production can remove it.
export const utilityVariants = /* @__PURE__ */ createUtilityVariants()

export function utilityColorway(id: string | null): IconPalette | undefined {
  return id && Object.prototype.hasOwnProperty.call(utilityVariants, id) ? utilityVariants[id] : undefined
}
