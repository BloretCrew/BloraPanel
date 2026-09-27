<script setup lang="ts">
import {computed, useId} from 'vue'
import {APP_SURFACE_PATH} from './surface-geometry'
const props=withDefaults(defineProps<{appId:string; variant?:string; padded?:boolean}>(),{variant:'a',padded:true})
const id=useId().replace(/:/g,'')
const kind=computed(()=>props.appId.replace(/^blora\./,''))
const palette:Record<string,string[]>={
  launcher:['#3168b0','#8baed9','#e4ecf8'],instances:['#355fc0','#81a5ee','#e2eafa'],
  backups:['#157c76','#6ebfb0','#dcf0e9'],files:['#ac701d','#edbe68','#f8edda'],
  editor:['#7955ac','#b59adb','#eee6f6'],terminal:['#344b68','#83b9ce','#e2eaf0'],
  tasks:['#427954','#91be8c','#e4eee2'],nodes:['#516ac3','#9aa9e2','#e7eafa'],
  monitor:['#117e86','#72c7c4','#dff2ef'],extensions:['#4264cd','#92aef7','#e5ecff'],
  users:['#8b527e','#cca0c1','#f1e5ee'],settings:['#536879','#a6b8c6','#e5ebef'],
  docker:['#2779b2','#89c5e9','#e1f0fa'],system:['#647353','#b6c29a','#edf0e3'],
}
const colors=computed(()=>palette[kind.value]||palette.extensions!)
const primary=computed(()=>props.variant==='c'?colors.value[1]:colors.value[0])
const secondary=computed(()=>props.variant==='c'?'#edf3fa':colors.value[1])
const base=computed(()=>props.variant==='c'?'#202b39':props.variant==='b'?'#f5f8fc':colors.value[2])
</script>
<template>
  <svg class="study-art" :class="`study-${variant}`" :viewBox="padded?'0 0 96 96':'8 8 80 80'" fill="none" aria-hidden="true">
    <defs>
      <linearGradient :id="id+'base'" x1="0" y1="0" x2=".65" y2="1">
        <stop :stop-color="variant==='c'?'#344253':'#ffffff'"/>
        <stop offset="1" :stop-color="base"/>
      </linearGradient>
      <linearGradient :id="id+'ink'" x1="0" y1="0" x2=".6" y2="1">
        <stop :stop-color="secondary"/><stop offset="1" :stop-color="primary"/>
      </linearGradient>
    </defs>
    <circle v-if="variant==='a'" cx="48" cy="48" r="40" :fill="base"/>
    <path v-else :d="APP_SURFACE_PATH" :fill="variant==='b'?base:`url(#${id}base)`" :stroke="variant==='c'?'#ffffff24':'#ffffff'" stroke-width=".8"/>
    <g :fill="primary" :stroke="primary" stroke-width="0" stroke-linecap="round" stroke-linejoin="round">
      <g v-if="kind==='launcher'">
        <rect x="26" y="26" width="19" height="19" rx="5"/><rect x="51" y="26" width="19" height="19" rx="5" :fill="secondary"/>
        <rect x="26" y="51" width="19" height="19" rx="5" :fill="secondary"/><rect x="51" y="51" width="19" height="19" rx="5"/>
      </g>
      <g v-else-if="kind==='instances'">
        <rect x="25" y="26" width="46" height="19" rx="5" :fill="secondary"/>
        <rect x="25" y="51" width="46" height="19" rx="5"/>
        <path d="M34 35.5h3m-3 25h3" stroke="white" stroke-width="3"/>
        <path d="M48 35.5h14m-14 25h14" stroke="white" stroke-opacity=".75" stroke-width="2.5"/>
      </g>
      <g v-else-if="kind==='backups'">
        <path d="M27 40a23 23 0 1 1 1 20" fill="none" stroke-width="5"/>
        <path d="M25 26v15h15" fill="none" stroke-width="5"/>
        <path d="M49 32v17l10 6" fill="none" :stroke="secondary" stroke-width="5"/>
      </g>
      <g v-else-if="kind==='files'">
        <path d="M24 33a6 6 0 0 1 6-6h13l8 8h15a6 6 0 0 1 6 6v23a6 6 0 0 1-6 6H30a6 6 0 0 1-6-6Z" :fill="secondary"/>
        <path d="M24 43h48v21a6 6 0 0 1-6 6H30a6 6 0 0 1-6-6Z"/>
        <path d="M33 57h14" stroke="white" stroke-width="3"/>
      </g>
      <g v-else-if="kind==='editor'">
        <rect x="27" y="24" width="38" height="49" rx="6" :fill="secondary" opacity=".55"/>
        <path d="m36 58 2-10 23-23a5.7 5.7 0 0 1 8 8L46 56Z"/>
        <path d="M37 65h20" stroke-width="3"/>
      </g>
      <g v-else-if="kind==='terminal'">
        <path d="m28 31 17 17-17 17" fill="none" stroke-width="5.5"/>
        <path d="M51 65h18" :stroke="secondary" stroke-width="5.5"/>
      </g>
      <g v-else-if="kind==='tasks'">
        <path d="m26 34 5 5 9-11m-14 31 5 5 9-11" fill="none" stroke-width="4.5"/>
        <path d="M49 34h21M49 59h21" :stroke="secondary" stroke-width="5"/>
      </g>
      <g v-else-if="kind==='nodes'">
        <path d="M48 31v18M28 64V50h40v14" fill="none" :stroke="secondary" stroke-width="4"/>
        <rect x="39" y="22" width="18" height="18" rx="5"/>
        <rect x="19" y="58" width="18" height="18" rx="5"/><rect x="59" y="58" width="18" height="18" rx="5"/>
      </g>
      <g v-else-if="kind==='monitor'">
        <path d="M25 68V48m15 20V35m15 33V23m15 45V42" :stroke="secondary" stroke-width="7"/>
        <path d="m23 55 15-9 17 7 18-23" fill="none" stroke-width="4"/>
      </g>
      <g v-else-if="kind==='extensions'">
        <path d="M25 37h46l-3 32H28Z"/>
        <path d="M37 38v-5a11 11 0 0 1 22 0v5" fill="none" :stroke="secondary" stroke-width="4"/>
        <path d="M48 47v13m-6-6 6 6 6-6" fill="none" stroke="white" stroke-width="3"/>
      </g>
      <g v-else-if="kind==='users'">
        <circle cx="59" cy="35" r="9" :fill="secondary"/><path d="M46 69v-7a14 14 0 0 1 28 0v7Z" :fill="secondary"/>
        <circle cx="38" cy="34" r="10"/><path d="M22 69v-7a16 16 0 0 1 32 0v7Z"/>
      </g>
      <g v-else-if="kind==='settings'">
        <path d="M42 23h12l2 7 6 4 7-2 6 10-5 6v7l5 5-6 10-7-2-6 4-2 7H42l-2-7-6-4-7 2-6-10 5-5v-7l-5-6 6-10 7 2 6-4Z" transform="translate(4 2) scale(.92)"/>
        <circle cx="48" cy="49" r="12" :fill="base"/>
        <circle cx="48" cy="49" r="6" :fill="secondary"/>
      </g>
      <g v-else-if="kind==='docker'">
        <path d="m48 23 25 13-25 13-25-13Z" :fill="secondary"/>
        <path d="m23 43 22 12v23L23 66Z"/><path d="m51 55 22-12v23L51 78Z" opacity=".7"/>
      </g>
      <g v-else>
        <path d="M31 24v48m17-48v48m17-48v48" fill="none" :stroke="secondary" stroke-width="3"/>
        <rect x="25" y="34" width="12" height="15" rx="4"/><rect x="42" y="52" width="12" height="15" rx="4"/><rect x="59" y="26" width="12" height="15" rx="4"/>
      </g>
    </g>
  </svg>
</template>
<style scoped>
.study-art{display:block;width:100%;height:100%;overflow:visible;flex-shrink:0}
.study-b{filter:drop-shadow(0 1px 1px #26375018)}
.study-c{filter:drop-shadow(0 1px 1px #26375022)}
</style>
