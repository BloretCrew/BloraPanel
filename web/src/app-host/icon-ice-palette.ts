import type { IconLayer } from './classic-icon-layers'

export type IconPalette = {
  id: string
  backgrounds: Record<string, string>
  colors: Record<string, Record<string, string>>
  // A review can replace individual app artwork without mutating the baseline.
  artwork?: Record<string, IconLayer[]>
}

// Colors from the cold white / blue reference, applied to the existing h04 shapes.
export const iceIconPalette: IconPalette = {
  id: 'ice',
  backgrounds: {
    launcher: '#f7fbff', instances: '#408cdd', backups: '#3db0a2', files: '#f8fcff',
    editor: '#fbfaff', terminal: '#25343f', tasks: '#fcfdff', nodes: '#ecf3fc',
    monitor: '#253740', extensions: '#438fdd', users: '#f3effa', settings: '#becad9',
    docker: '#f4faff', system: '#f4f7fb',
  },
  colors: {
    launcher: { '#537977': '#67c7d0', '#bc9874': '#edb287', '#70849c': '#749de7', '#9e8594': '#bc99d9' },
    instances: { '#6e86a5': '#cbe8fc', '#415e81': '#e2f1fc', '#fff5d9': '#ffffff', '#d8eab3': '#f0fff8', '#f4f6ff': '#72a5d8' },
    backups: { '#f6fbf8': '#f5fffc', '#528b7d': '#bee9df', '#8cbaa9': '#4cae9e' },
    files: { '#a78045': '#66b8df', '#fff9eb': '#fbfeff', '#caa264': '#3f9dd7' },
    editor: { '#fffaf5': '#fcfcff', '#c5b5c3': '#d7cbea', '#9d8797': '#baafce', '#83657c': '#a481d0', '#bea07f': '#c4b9d2', '#655349': '#7e688f' },
    terminal: { '#4b637d': '#273d4b', '#f6f9fc': '#f2f8ff', '#b2daca': '#8bd7c8' },
    tasks: { '#fffaf1': '#ffffff', '#72925d': '#70b9df', '#c58b65': '#e2a386', '#a59a7e': '#b2c3d3' },
    nodes: { '#abb8d5': '#bacde5', '#5776b6': '#6599dc', '#c98a77': '#72bfb8', '#899bca': '#9595d7', '#e1e7f7': '#f2f8ff', '#f9e3d2': '#ecfffb' },
    monitor: { '#f7fcf6': '#253e47', '#c0dac9': '#365c64', '#458678': '#74ccbd', '#d9a175': '#a0dfcb' },
    extensions: { '#cd8075': '#d1e9fd', '#fff5ed': '#ffffff', '#b8625f': '#87b7e6' },
    users: { '#a78473': '#b7c6e0', '#c4a594': '#c5d4eb', '#8d789a': '#a994cc', '#705e7e': '#a18bc4' },
    settings: { '#72869c': '#f3f7fd', '#f5f8fc': '#94a9c1', '#b3c8dc': '#d8e6f3' },
    docker: { '#6d97c7': '#5ba4df', '#8db5d9': '#8ac8e9', '#d9aa71': '#aecddd', '#466f9e': '#478ecb', '#dce9f5': '#f0f9ff' },
    system: { '#cabfa4': '#c5d0df', '#70845c': '#85c7be', '#be8272': '#a795d0', '#8897b9': '#79a8dd' },
  },
}
