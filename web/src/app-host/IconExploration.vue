<script setup lang="ts">
import {computed,useId} from 'vue'
import {LayoutGrid,Server,History,Folder,FilePenLine,Terminal,ListChecks,Network,ChartNoAxesCombined,Box,UsersRound,Settings,ShoppingBag,SlidersHorizontal} from '@lucide/vue'
import {APP_SURFACE_PATH} from './surface-geometry'
import {explorations,type ExplorationId} from './icon-explorations'
const props=withDefaults(defineProps<{appId:string;variant:ExplorationId;padded?:boolean}>(),{padded:true})
const clipId=useId().replace(/:/g,'')+'exploration-clip'
const kind=computed(()=>props.appId.replace(/^blora\./,''))
const family=computed(()=>explorations.find(v=>v.id===props.variant)!.family)
const icons={launcher:LayoutGrid,instances:Server,backups:History,files:Folder,editor:FilePenLine,terminal:Terminal,tasks:ListChecks,nodes:Network,monitor:ChartNoAxesCombined,docker:Box,users:UsersRound,settings:Settings,extensions:ShoppingBag,system:SlidersHorizontal}
const icon=computed(()=>icons[kind.value as keyof typeof icons]||Box)
const palettes:Record<string,string[]>={
  launcher:['#316bc3','#99bbec','#e7effa'],instances:['#365dd0','#a3b9f4','#e9edfc'],
  backups:['#168078','#83c8ba','#e2f1eb'],files:['#b27a20','#f0c670','#f8efdc'],
  editor:['#7656b5','#b7a0df','#eee8f7'],terminal:['#315977','#8cb8cb','#e3edf2'],
  tasks:['#387b61','#92c6a6','#e5f2e9'],nodes:['#5366b6','#a5b5e5','#e9edf9'],
  monitor:['#107e8c','#89cdd0','#e1f1f1'],docker:['#237fb5','#8cc8ea','#e3f0f8'],
  users:['#955b80','#d7adc7','#f4e9f0'],settings:['#546b80','#b0c1cf','#e9eef2'],
  extensions:['#4365d5','#a1b8f8','#e9effe'],system:['#73764c','#c4c69a','#f0f1e6'],
}
const c=computed(()=>palettes[kind.value]||palettes.extensions!)
const ink=computed(()=>props.variant==='e'?'#fff':props.variant==='l'?'#293e4a':c.value[0])
const accent=computed(()=>c.value[1])
// Individually composed two-plane silhouettes, on a common 24-unit grid.
const solids:Record<string,[string,string]>={
 launcher:['M3 3H11V11H3ZM13 13H21V21H13Z','M13 3H21V11H13ZM3 13H11V21H3Z'],
 instances:['M4 4H20V10H4ZM4 13H20V20H4Z','M6 6H8V8H6ZM6 15H8V17H6ZM12 6H18V8H12ZM12 15H18V17H12Z'],
 backups:['M4 8A9 9 0 1 1 3.4 16L6 15A6 6 0 1 0 6.5 9H10L3 14V4Z','M11 6H13V12L17 14L16 16L11 13Z'],
 files:['M2 6Q2 4 4 4H10L13 7H20Q22 7 22 9V19Q22 21 20 21H4Q2 21 2 19Z','M2 10H22V18Q22 20 20 20H4Q2 20 2 18Z'],
 editor:['M5 3H14L19 8V21H5Z','M8 17L9 13L18 4L21 7L12 16ZM8 19H16V20H8Z'],
 terminal:['M4 5L12 12L4 19L2 17L8 12L2 7Z','M13 17H22V20H13Z'],
 tasks:['M3 6L5 8L9 3L11 5L5 12L1 8ZM3 16L5 18L9 13L11 15L5 22L1 18Z','M14 5H23V8H14ZM14 15H23V18H14Z'],
 nodes:['M8 2H16V9H8ZM1 16H8V23H1ZM16 16H23V23H16Z','M11 9H13V12H20V16H18V14H6V16H4V12H11Z'],
 monitor:['M2 15H6V22H2ZM10 9H14V22H10ZM18 2H22V22H18Z','M1 11L8 5L13 7L20 0L22 2L14 11L8 9L3 14Z'],
 docker:['M12 1L23 7L12 13L1 7ZM1 10L11 15V24L1 18Z','M13 15L23 10V18L13 24Z'],
 users:['M8 3A4 4 0 1 1 8 11A4 4 0 1 1 8 3ZM1 21V19a7 7 0 0 1 14 0V21Z','M18 5A3 3 0 1 1 18 11A3 3 0 1 1 18 5ZM16 21V18A8 8 0 0 0 15 14A6 6 0 0 1 23 19V21Z'],
 settings:['M10 1H14L15 4L18 6L21 5L23 9L21 12V15L23 17L21 21L18 20L15 22L14 24H10L9 22L6 20L3 21L1 17L3 15V12L1 9L3 5L6 6L9 4ZM12 8A5 5 0 1 0 12 18A5 5 0 1 0 12 8Z','M12 10A3 3 0 1 1 12 16A3 3 0 1 1 12 10Z'],
 extensions:['M3 8H21L20 22H4ZM7 8V6a5 5 0 0 1 10 0V8H15V6a3 3 0 0 0-6 0V8Z','M11 11H13V16L15 14L17 16L12 21L7 16L9 14L11 16Z'],
 system:['M3 2H5V22H3ZM11 2H13V22H11ZM19 2H21V22H19Z','M1 7H7V13H1ZM9 15H15V21H9ZM17 3H23V9H17Z'],
}
// A separate geometric vocabulary for the more exploratory Fold/Interlace sets.
const ribbons:Record<string,[string,string]>={
 launcher:['M2 3H12V12H2ZM12 12H22V21H12Z','M12 3H22V12H12ZM2 12H12V21H2Z'],
 instances:['M2 3H22V10H2ZM2 14H22V21H2Z','M14 3H22V10H14ZM8 14H16V21H8Z'],
 backups:['M12 1A11 11 0 1 1 1 12H5A7 7 0 1 0 12 5Z','M1 2L10 11H1ZM11 7H14V12L18 15L16 18L11 14Z'],
 files:['M1 4H10L13 7H23V21H1Z','M1 10H23L18 21H1Z'],
 editor:['M3 1H16V23H3Z','M9 16L19 2L23 5L13 19L8 21Z'],
 terminal:['M2 3L12 12L2 21V15L6 12L2 9Z','M13 17H23V22H13Z'],
 tasks:['M1 3H6V8H1ZM1 16H6V21H1Z','M9 3H23V8H9ZM9 16H19V21H9Z'],
 nodes:['M8 1H16V9H8ZM1 15H9V23H1ZM15 15H23V23H15Z','M10 9H14V12H21V15H17V16H7V15H3V12H10Z'],
 monitor:['M1 16H6V23H1ZM9 9H14V23H9ZM17 2H22V23H17Z','M1 12L17 2H22L6 12Z'],
 docker:['M12 1L23 7L12 13L1 7ZM1 9L11 15V23L1 17Z','M13 15L23 9V17L13 23Z'],
 users:['M6 2A5 5 0 1 1 6 12A5 5 0 1 1 6 2ZM0 23V20a6 6 0 0 1 12 0V23Z','M18 2A5 5 0 1 1 18 12A5 5 0 1 1 18 2ZM12 23V20a6 6 0 0 1 12 0V23Z'],
 settings:['M12 1L23 7V18L12 24L1 18V7ZM12 7A5 5 0 1 0 12 17A5 5 0 1 0 12 7Z','M12 9A3 3 0 1 1 12 15A3 3 0 1 1 12 9Z'],
 extensions:['M1 1H11V11H1ZM13 13H23V23H13ZM1 13H11V23H1Z','M18 0L24 6L18 12L12 6Z'],
 system:['M1 5H23V9H1ZM1 16H23V20H1Z','M5 1H10V13H5ZM15 12H20V24H15Z'],
}
const paths=computed(()=>(family.value==='ribbon'?ribbons:solids)[kind.value]||solids.extensions!)
</script>
<template>
  <svg class="exploration-art" :class="`exploration-${variant}`" :viewBox="padded?'0 0 96 96':'8 8 80 80'" :style="{'--ink':c[0],'--accent':c[1]}" fill="none" aria-hidden="true">
    <defs v-if="variant==='k'"><clipPath :id="clipId"><path :d="APP_SURFACE_PATH"/></clipPath></defs>
    <path v-if="variant==='d'" :d="APP_SURFACE_PATH" fill="#fbfcfe" stroke="#dce3e9" stroke-width=".65"/>
    <path v-else-if="variant==='e'" :d="APP_SURFACE_PATH" :fill="c[0]"/>
    <g v-else-if="variant==='g'">
      <path d="M8 32Q8 8 32 8H64L88 32V64Q88 88 64 88H32Q8 88 8 64Z" :fill="c[2]"/>
      <path d="M64 8V32H88Z" :fill="c[1]"/>
    </g>
    <path v-else-if="variant==='h'" :d="APP_SURFACE_PATH" fill="#232e3c" stroke="#435062" stroke-width=".7"/>
    <g v-else-if="variant==='j'">
      <path :d="APP_SURFACE_PATH" fill="#ffffff56" stroke="#ffffffdd" stroke-width="1.2"/>
      <rect x="38" y="34" width="38" height="38" rx="12" :fill="c[1]" opacity=".3"/>
    </g>
    <g v-else-if="variant==='k'">
      <path :d="APP_SURFACE_PATH" :fill="c[0]"/>
      <rect x="28" y="8" width="60" height="80" fill="#fafbfd" :clip-path="`url(#${clipId})`"/>
    </g>
    <path v-else-if="variant==='l'" d="M32 8H64Q88 8 88 32V64Q88 88 64 88H8V32Q8 8 32 8Z" :fill="c[2]"/>
    <component v-if="family==='line'" :is="icon" class="exploration-line" :x="variant==='k'?32:25" y="25" :width="variant==='k'?45:46" height="46" :stroke-width="variant==='j'?1.65:1.8" :color="variant==='h'?c[1]:variant==='k'?'#2e404e':c[0]"/>
    <g v-else :transform="variant==='f'||variant==='i'?'translate(18 18) scale(2.5)':'translate(24 24) scale(2)'" fill-rule="evenodd">
      <path :d="paths[0]" :fill="ink"/>
      <path :d="paths[1]" :fill="accent"/>
    </g>
  </svg>
</template>
<style scoped>
.exploration-art{display:block;width:100%;height:100%;overflow:visible;flex-shrink:0}
.exploration-d .exploration-line :deep(> :nth-child(even)),.exploration-j .exploration-line :deep(> :last-child){stroke:var(--accent)}
.exploration-h .exploration-line :deep(> :nth-child(even)){stroke:#d6e4ed}
.exploration-k .exploration-line :deep(> :last-child){stroke:var(--ink)}
</style>
