<script setup lang="ts">
import { computed, ref } from 'vue'
import { useElementSize } from '@vueuse/core'
import { APP_CORNER_RATIO, continuousRectPath } from '../app-host/surface-geometry'

const props = defineProps<{ inset: number }>()
const surface = ref<HTMLElement>()
const { width, height } = useElementSize(surface)
const contour = computed(() => {
  const w = width.value - 2 * props.inset, h = height.value - 2 * props.inset
  if (w <= 0 || h <= 0) return ''
  // Both material layers use the same unchanged continuous outer contour.
  return continuousRectPath(w, h, h * APP_CORNER_RATIO, props.inset, props.inset, props.inset)
})
</script>

<template>
  <div ref="surface" class="dock-surface" aria-hidden="true">
    <div class="dock-backdrop" :style="{clipPath:contour ? `path('${contour}')` : undefined}"></div>
    <svg :viewBox="`0 0 ${width || 1} ${height || 1}`"><path :d="contour" fill="var(--dock-surface,var(--surface-low))" stroke="var(--dock-edge,transparent)" stroke-width="1"/></svg>
  </div>
</template>

<style scoped>
.dock-surface,.dock-surface>svg{position:absolute;inset:0;width:100%;height:100%;pointer-events:none}
.dock-surface{z-index:0}
.dock-backdrop{position:absolute;inset:0;backdrop-filter:var(--dock-backdrop-filter,none)}
.dock-surface>svg{overflow:visible;filter:var(--dock-shadow,none)}
</style>
