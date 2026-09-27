<script setup lang="ts">
import {computed,useId} from 'vue'
import {classicIconLayers,type IconLayer} from './classic-icon-layers'
import {foregroundById,foregroundMaterials} from './icon-foreground-materials'
import {blend,classicBackgrounds} from './icon-materials'
import {APP_SURFACE_PATH} from './surface-geometry'
const props=withDefaults(defineProps<{appId:string;material:string;padded?:boolean;bare?:boolean}>(),{padded:true,bare:false})
const kind=computed(()=>props.appId.replace(/^blora\./,''))
const layers=computed(()=>classicIconLayers[kind.value]||classicIconLayers.launcher!)
const material=computed(()=>foregroundById(props.material)||foregroundMaterials[0]!)
const background=computed(()=>classicBackgrounds[kind.value]||'#e6e9e8')
const uid=useId().replace(/:/g,'')+'foreground'
const paint=(id:string)=>`url(#${uid}-${id})`
const isFace=(layer:IconLayer)=>layer.role==='surface'||layer.role==='paper'||(layer.role==='ink'&&['backups','system','nodes'].includes(kind.value)&&Number(layer.attrs['stroke-width'])>=4)
const tint=(hex:string)=>{
  const rgb=hex.slice(1).match(/../g)!.map(v=>parseInt(v,16));const avg=rgb.reduce((a,b)=>a+b,0)/3
  const tone='#'+rgb.map(v=>Math.round(Math.min(255,Math.max(0,avg+(v-avg)*material.value.chroma))).toString(16).padStart(2,'0')).join('')
  return blend(tone,'#b4d5f0',material.value.cool)
}
const face=(layer:IconLayer)=>tint(layer.attrs.fill||layer.attrs.stroke||'#8ba1b5')
const opacity=(layer:IconLayer)=>layer.role==='paper'?material.value.paper:material.value.opacity
const hasFaceLight=computed(()=>material.value.light>0||material.value.shade>0)
const facePaint=(layer:IconLayer,index:number)=>hasFaceLight.value?paint('face-'+index):face(layer)
const field=(layer:IconLayer,index:number)=>{
  const a=layer.attrs
  if(layer.tag==='rect')return {cx:Number(a.x)+Number(a.width)*.26,cy:Number(a.y)+Number(a.height)*.16,rx:Number(a.width)*.46,ry:Number(a.height)*.22}
  if(layer.tag==='circle')return {cx:Number(a.cx)-Number(a.r)*.32,cy:Number(a.cy)-Number(a.r)*.55,rx:Number(a.r),ry:Number(a.r)*.38}
  const locations:Record<string,number[][]>={
    launcher:[[30,25],[55,26],[56,53],[29,53]],files:[[28,27],[],[34,42]],
    editor:[[34,24],[54,25],[],[70,35]],extensions:[[],[39,24],[42,26],[43,55]],
    users:[[],[58,51],[],[29,49]],docker:[[],[],[],[],[36,56]],
  }
  const point=locations[kind.value]?.[index]
  return {cx:point?.[0]||34,cy:point?.[1]||29,rx:kind.value==='editor'&&index===3?7:19,ry:6}
}
// Geometry and backing are fixed. Gaussian diffusion is clipped to each foreground
// plane and receives the preceding glyph layers (e.g. the folder's actual paper).
// Small marks remain crisp. Nothing blurs the complete icon or its background tile.
const lowerAttrs=(layer:IconLayer)=>({...layer.attrs,...(layer.attrs.fill?{fill:tint(layer.attrs.fill)}:{}),...(layer.attrs.stroke?{stroke:tint(layer.attrs.stroke)}:{})})
</script>
<template>
  <svg class="foreground-app-art" :data-material="material.id" :data-foreground-app="kind" :viewBox="padded?'0 0 96 96':'8 8 80 80'" fill="none" aria-hidden="true">
    <defs>
      <filter :id="uid+'-diffuse'" x="-30%" y="-30%" width="160%" height="160%" color-interpolation-filters="sRGB"><feGaussianBlur :stdDeviation="material.blur"/></filter>
      <filter v-if="material.diffusion>0" :id="uid+'-soft-field'" x="-50%" y="-50%" width="200%" height="200%" color-interpolation-filters="sRGB"><feGaussianBlur stdDeviation="7"/></filter>
      <linearGradient v-if="material.rim>0&&material.edgeMode!=='uniform'" :id="uid+'-rim'" x1="0" y1="0" x2=".4" y2="1"><stop stop-color="white" :stop-opacity="material.rim"/><stop offset=".5" stop-color="white" stop-opacity=".04"/><stop offset="1" stop-color="#354d65" :stop-opacity="material.rim*.25"/></linearGradient>
      <template v-for="(layer,i) in layers" :key="i">
        <clipPath :id="uid+'-plane-'+i"><component :is="layer.tag" v-bind="layer.attrs" :fill="layer.attrs.fill?'white':'none'" :stroke="layer.attrs.stroke?'white':undefined"/></clipPath>
        <mask v-if="!layer.attrs.fill&&isFace(layer)" :id="uid+'-stroke-plane-'+i" maskUnits="userSpaceOnUse" x="0" y="0" width="96" height="96" style="mask-type:alpha"><component :is="layer.tag" v-bind="layer.attrs" fill="none" stroke="white"/></mask>
        <linearGradient v-if="isFace(layer)&&hasFaceLight" :id="uid+'-face-'+i" x1=".18" y1="0" x2=".72" y2="1">
          <stop :stop-color="blend(face(layer),'#ffffff',material.light)"/>
          <stop offset=".6" :stop-color="face(layer)"/>
          <stop offset="1" :stop-color="blend(face(layer),'#273c58',material.shade)"/>
        </linearGradient>
      </template>
    </defs>
    <path v-if="!bare" class="fixed-icon-background" :d="APP_SURFACE_PATH" :fill="background"/>
    <g v-for="(layer,i) in layers" :key="i" :data-layer-role="layer.role">
      <template v-if="isFace(layer)">
        <g :clip-path="layer.attrs.fill?paint('plane-'+i):undefined" :mask="!layer.attrs.fill?paint('stroke-plane-'+i):undefined">
          <g :filter="paint('diffuse')">
            <rect width="96" height="96" :fill="blend(face(layer),background,.55)"/>
            <component v-for="(lower,k) in layers.slice(0,i)" :key="k" :is="lower.tag" v-bind="lowerAttrs(lower)"/>
          </g>
          <component :is="layer.tag" v-bind="layer.attrs" :fill="layer.attrs.fill?facePaint(layer,i):'none'" :stroke="layer.attrs.stroke?facePaint(layer,i):undefined" :opacity="opacity(layer)"/>
          <g v-if="material.diffusion>0" data-light-field :filter="paint('soft-field')" :opacity="material.diffusion">
            <ellipse v-bind="field(layer,i)" fill="white"/>
            <ellipse cx="68" cy="69" rx="16" ry="18" :fill="face(layer)" opacity=".7"/>
          </g>
          <component v-if="layer.attrs.fill&&material.rim>0" :is="layer.tag" v-bind="layer.attrs" data-light-edge fill="none" :stroke="material.edgeMode==='uniform'?'white':paint('rim')" :stroke-opacity="material.edgeMode==='uniform'?material.rim:1" :stroke-width="material.edgeWidth||1.1"/>
          <path v-if="kind==='editor'&&i===3&&material.rim>0" data-light-edge d="m56 52 17-22" stroke="white" :stroke-width="material.edgeWidth||1.25" stroke-linecap="round" :opacity="material.rim*.7"/>
        </g>
      </template>
      <component v-else :is="layer.tag" v-bind="lowerAttrs(layer)"/>
    </g>
  </svg>
</template>
<style scoped>
.foreground-app-art{display:block;width:100%;height:100%;overflow:visible;flex-shrink:0}
</style>
