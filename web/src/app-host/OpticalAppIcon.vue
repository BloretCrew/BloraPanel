<script setup lang="ts">
import {computed,onBeforeUnmount,onMounted,ref,watch} from 'vue'
import {opticsById,iconOptics} from './icon-optics'
import {renderOpticalIcon} from './icon-optics-renderer'
const props=withDefaults(defineProps<{appId:string;material:string;padded?:boolean;bare?:boolean;lightX?:number;lightY?:number;interactive?:boolean;probe?:'plain'|'grid'|'colour'}>(),{padded:true,bare:false,lightX:0,lightY:0,interactive:true,probe:'plain'})
const element=ref<HTMLCanvasElement>()
const ready=ref(false),error=ref('')
const m=computed(()=>opticsById(props.material)||iconOptics[0]!)
let frame=0
function draw(x=props.lightX,y=props.lightY){if(!element.value)return;try{renderOpticalIcon(element.value,props.appId,m.value,[x,y],props.bare,props.padded,props.probe);ready.value=true;error.value=''}catch(e){error.value=String(e);ready.value=false}}
function move(event:PointerEvent){if(!props.interactive||matchMedia('(prefers-reduced-motion: reduce)').matches)return;const b=element.value!.getBoundingClientRect();cancelAnimationFrame(frame);frame=requestAnimationFrame(()=>draw((event.clientX-b.left)/b.width*2-1,(event.clientY-b.top)/b.height*2-1))}
onMounted(()=>draw())
watch(()=>[props.appId,props.material,props.lightX,props.lightY,props.bare,props.padded,props.probe],()=>draw())
onBeforeUnmount(()=>cancelAnimationFrame(frame))
</script>
<template><canvas ref="element" class="optical-app-art" :data-optics="material" :data-ready="ready" :data-render-error="error||undefined" aria-hidden="true" @pointermove="move" @pointerleave="draw()"/></template>
<style scoped>.optical-app-art{display:block;width:100%;height:100%;flex-shrink:0}</style>
