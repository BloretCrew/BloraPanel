<script setup lang="ts">
import {computed,useId} from 'vue'
import ClassicIconArtwork from './ClassicIconArtwork.vue'
import {APP_SURFACE_PATH} from './surface-geometry'
import {blend,classicBackgrounds,classicAccents,materialById,originalMaterial} from './icon-materials'
import {classicColors} from './classic-icon-colors'
const props=withDefaults(defineProps<{appId:string;material:string;padded?:boolean}>(),{padded:true})
const kind=computed(()=>props.appId.replace(/^blora\./,''))
const m=computed(()=>materialById(props.material)||originalMaterial)
const uid=useId().replace(/:/g,'')+'material'
const ref=(name:string)=>`url(#${uid}-${name})`
const palette=computed(()=>classicColors[kind.value]||classicColors.launcher!)
const accent=computed(()=>classicAccents[kind.value]||'#8ba1ba')
const mask=computed(()=>`url("data:image/svg+xml,${encodeURIComponent(`<svg xmlns='http://www.w3.org/2000/svg' viewBox='${props.padded?'0 0 96 96':'8 8 80 80'}'><path d='${APP_SURFACE_PATH}' fill='white'/></svg>`)}")`)
const darkColor=(hex:string)=>m.value.dark?blend(hex,'#eaf3fa',.14):hex
const color=(hex:string)=>m.value.flat?darkColor(hex):ref('tone-'+hex.slice(1))
const bottom=computed(()=>m.value.dark?'#233448':'#ffffff')
</script>
<template>
  <span class="material-app-art" :data-material="material" :data-material-app="kind">
    <span v-if="m.backdrop" class="material-backdrop" :style="{maskImage:mask,WebkitMaskImage:mask,backdropFilter:`blur(${m.backdrop}px) saturate(1.1)`,WebkitBackdropFilter:`blur(${m.backdrop}px) saturate(1.1)`}"></span>
    <svg :viewBox="padded?'0 0 96 96':'8 8 80 80'" fill="none" aria-hidden="true">
      <defs>
        <clipPath :id="uid+'-clip'"><path :d="APP_SURFACE_PATH"/></clipPath>
        <g :id="uid+'-art'"><ClassicIconArtwork :kind="kind" :color="color"/></g>
        <linearGradient v-for="hex in palette" :key="hex" :id="uid+'-tone-'+hex.slice(1)" x1=".15" y1="0" x2=".8" y2="1">
          <stop :stop-color="blend(darkColor(hex),'#ffffff',m.light)"/>
          <stop offset=".65" :stop-color="darkColor(hex)"/>
          <stop offset="1" :stop-color="blend(darkColor(hex),'#345578',m.separation*.55)"/>
        </linearGradient>
        <linearGradient :id="uid+'-edge'" x1="0" y1="0" x2=".65" y2="1"><stop stop-color="white" :stop-opacity="m.edge"/><stop offset=".46" stop-color="white" :stop-opacity="m.edge*.22"/><stop offset="1" stop-color="white" :stop-opacity="m.edge*.5"/></linearGradient>
        <filter :id="uid+'-diffusion'" x="-30%" y="-30%" width="160%" height="160%" color-interpolation-filters="sRGB"><feGaussianBlur :stdDeviation="m.blur"/></filter>
        <filter :id="uid+'-finish'" x="-5%" y="-5%" width="110%" height="110%" color-interpolation-filters="sRGB">
          <feColorMatrix type="saturate" :values="String(m.saturation)"/>
          <feDropShadow dx="0" dy=".6" stdDeviation=".65" flood-color="#29475f" :flood-opacity="m.separation"/>
        </filter>
      </defs>
      <g :clip-path="ref('clip')">
        <path :d="APP_SURFACE_PATH" :fill="material==='original'?(classicBackgrounds[kind]||'#e8e2ea'):bottom" :fill-opacity="m.surface"/>
        <g v-if="m.tint||m.pearl" :filter="ref('diffusion')">
          <ellipse cx="66" cy="71" rx="37" ry="33" :fill="accent" :opacity="m.tint"/>
          <ellipse v-if="m.pearl" cx="29" cy="24" rx="26" ry="20" fill="#d9c8ef" opacity=".22"/>
          <ellipse v-if="m.pearl" cx="75" cy="75" rx="22" ry="28" fill="#ccefe8" opacity=".3"/>
        </g>
        <use v-if="m.diffuse" :href="'#'+uid+'-art'" :filter="ref('diffusion')" :opacity="m.diffuse" transform="translate(1 3)"/>
        <use :href="'#'+uid+'-art'" :filter="material==='original'?undefined:ref('finish')" :opacity="m.ink"/>
      </g>
      <path v-if="m.edge" :d="APP_SURFACE_PATH" :stroke="ref('edge')" stroke-width=".85"/>
    </svg>
  </span>
</template>
<style scoped>
.material-app-art{display:block;position:relative;width:100%;height:100%;flex-shrink:0;isolation:isolate}
.material-app-art>svg{position:relative;display:block;width:100%;height:100%;overflow:visible}
.material-backdrop{position:absolute;inset:0;pointer-events:none;mask-size:100% 100%;-webkit-mask-size:100% 100%;mask-repeat:no-repeat;-webkit-mask-repeat:no-repeat}
</style>
