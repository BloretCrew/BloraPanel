<script setup lang="ts">
// Shared Material-style icon geometry for legacy material comparisons.
withDefaults(defineProps<{kind:string;color?:(hex:string)=>string}>(),{color:(hex:string)=>hex})
</script>
<template>
    <g v-if="kind === 'launcher'">
      <path d="M22 45V34a12 12 0 0 1 24 0v11Z" :fill="color('#537977')"/>
      <path d="M51 22h11a12 12 0 0 1 0 24H51Z" :fill="color('#bc9874')"/>
      <path d="M74 51v11a12 12 0 0 1-24 0V51Z" :fill="color('#70849c')"/>
      <path d="M45 74H34a12 12 0 0 1 0-24h11Z" :fill="color('#9e8594')"/>
    </g>
    <g v-else-if="kind === 'instances'">
      <rect x="22" y="23" width="52" height="23" rx="7" :fill="color('#6e86a5')"/>
      <rect x="22" y="50" width="52" height="23" rx="7" :fill="color('#415e81')"/>
      <circle cx="33" cy="34.5" r="3" :fill="color('#fff5d9')"/><circle cx="33" cy="61.5" r="3" :fill="color('#d8eab3')"/>
      <path d="M46 34.5h16M46 61.5h16" :stroke="color('#f4f6ff')" stroke-width="3" stroke-linecap="round"/>
    </g>
    <g v-else-if="kind === 'backups'">
      <circle cx="48" cy="48" r="27" :fill="color('#528b7d')"/>
      <circle cx="48" cy="48" r="20" :fill="color('#f6fbf8')"/>
      <path d="M48 31.5V48l12.5 8" :stroke="color('#8cbaa9')" stroke-width="5.5" stroke-linecap="round" stroke-linejoin="round"/>
    </g>
    <g v-else-if="kind === 'files'">
      <path d="M20 64V30a6 6 0 0 1 6-6h14l7 9h23a6 6 0 0 1 6 6v25Z" :fill="color('#a78045')"/>
      <rect x="27" y="33" width="41" height="24" rx="3" :fill="color('#fff9eb')"/>
      <path d="M20 40h56v27a7 7 0 0 1-7 7H27a7 7 0 0 1-7-7Z" :fill="color('#caa264')"/>
      <path d="M59 62h7" :stroke="color('#fff9eb')" stroke-width="3" stroke-linecap="round"/>
    </g>
    <g v-else-if="kind === 'editor'">
      <path d="M28 21h24l12 12v41H28a5 5 0 0 1-5-5V26a5 5 0 0 1 5-5Z" :fill="color('#fffaf5')"/>
      <path d="M52 21v12h12" :fill="color('#c5b5c3')"/>
      <path d="M32 42h20M32 50h16M32 58h10" :stroke="color('#9d8797')" stroke-width="2.7" stroke-linecap="round"/>
      <path d="m49 64 4-11 19-25a4.5 4.5 0 0 1 7 5L60 59Z" :fill="color('#83657c')"/>
      <path d="m49 64 4-11 7 6Z" :fill="color('#bea07f')"/>
      <path d="m49 64 4-2-2-2Z" :fill="color('#655349')"/>
    </g>
    <g v-else-if="kind === 'terminal'">
      <rect x="21" y="26" width="54" height="45" rx="9" :fill="color('#4b637d')"/>
      <path d="m31 38 10 10-10 10" :stroke="color('#f6f9fc')" stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M51 59h12" :stroke="color('#b2daca')" stroke-width="4" stroke-linecap="round"/>
    </g>
    <g v-else-if="kind === 'tasks'">
      <rect x="22" y="22" width="52" height="52" rx="9" :fill="color('#fffaf1')"/>
      <circle cx="34" cy="35" r="6" :fill="color('#72925d')"/><circle cx="34" cy="59" r="6" :fill="color('#c58b65')"/>
      <path d="m31 35 2 2 4-4m-6 26 2 2 4-4" :stroke="color('#fffaf1')" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M47 35h17M47 59h13" :stroke="color('#a59a7e')" stroke-width="3" stroke-linecap="round"/>
    </g>
    <g v-else-if="kind === 'nodes'">
      <path d="m29 31 38 6-18 31Z" :stroke="color('#abb8d5')" stroke-width="4" stroke-linejoin="round"/>
      <circle cx="29" cy="31" r="11" :fill="color('#5776b6')"/><circle cx="67" cy="37" r="10" :fill="color('#c98a77')"/><circle cx="49" cy="68" r="12" :fill="color('#899bca')"/>
      <circle cx="29" cy="31" r="3" :fill="color('#e1e7f7')"/><circle cx="67" cy="37" r="3" :fill="color('#f9e3d2')"/>
    </g>
    <g v-else-if="kind === 'monitor'">
      <rect x="22" y="24" width="52" height="47" rx="9" :fill="color('#f7fcf6')"/>
      <path d="M32 58h8V46h9V35h9v23h7" :fill="color('#c0dac9')"/>
      <path d="M22 54h12l8-13 11 19 9-27 7 12h5" :stroke="color('#458678')" stroke-width="3.3" stroke-linecap="round" stroke-linejoin="round"/>
      <circle cx="62" cy="33" r="4" :fill="color('#d9a175')"/>
    </g>
    <g v-else-if="kind === 'extensions'">
      <rect x="23" y="38" width="50" height="36" rx="5" :fill="color('#cd8075')"/>
      <path d="m25 23-6 16v4a7 7 0 0 0 14 0 7 7 0 0 0 14 0 7 7 0 0 0 14 0 7 7 0 0 0 14 0v-4l-6-16Z" :fill="color('#fff5ed')"/>
      <path d="M35 23h10l-1 16H32Zm20 0h10l7 16H58Z" :fill="color('#b8625f')"/>
      <path d="M38 74V55a3 3 0 0 1 3-3h14a3 3 0 0 1 3 3v19Z" :fill="color('#fff5ed')"/>
      <path d="M48 55v19" :stroke="color('#cd8075')" stroke-width="2"/>
    </g>
    <g v-else-if="kind === 'users'">
      <circle cx="63" cy="36" r="10" :fill="color('#a78473')"/><path d="M50 72V60a13 13 0 0 1 26 0v12Z" :fill="color('#c4a594')"/>
      <circle cx="36" cy="33" r="12" :fill="color('#8d789a')"/><path d="M19 73V61a17 17 0 0 1 34 0v12Z" :fill="color('#705e7e')"/>
    </g>
    <g v-else-if="kind === 'settings'">
      <path d="M40 20h16l2 9 7 4 9-3 8 14-7 6v8l7 6-8 14-9-3-7 4-2 9H40l-2-9-7-4-9 3-8-14 7-6v-8l-7-6 8-14 9 3 7-4Z" transform="translate(5.8 2) scale(.88)" :fill="color('#72869c')" :stroke="color('#72869c')" stroke-width="2" stroke-linejoin="round"/>
      <circle cx="48" cy="49.5" r="12" :fill="color('#f5f8fc')"/>
      <circle cx="48" cy="49.5" r="5" :fill="color('#b3c8dc')"/>
    </g>
    <g v-else-if="kind === 'docker'">
      <rect x="23" y="37" width="14" height="13" rx="2" :fill="color('#6d97c7')"/><rect x="41" y="37" width="14" height="13" rx="2" :fill="color('#8db5d9')"/><rect x="59" y="37" width="14" height="13" rx="2" :fill="color('#d9aa71')"/>
      <rect x="41" y="20" width="14" height="13" rx="2" :fill="color('#466f9e')"/>
      <path d="M19 54h59c-4 15-13 21-31 21-14 0-23-6-28-21Z" :fill="color('#466f9e')"/>
      <path d="M27 64h13" :stroke="color('#dce9f5')" stroke-width="3" stroke-linecap="round"/>
    </g>
    <g v-else-if="kind === 'system'">
      <path d="M29 25v48m19-48v48m19-48v48" :stroke="color('#cabfa4')" stroke-width="4" stroke-linecap="round"/>
      <rect x="22" y="33" width="14" height="17" rx="5" :fill="color('#70845c')"/><rect x="41" y="51" width="14" height="17" rx="5" :fill="color('#be8272')"/><rect x="60" y="26" width="14" height="17" rx="5" :fill="color('#8897b9')"/>
    </g>
    <g v-else>
      <path d="M22 36h20v-9a8 8 0 0 1 16 0v9h16v18h-9a8 8 0 0 0 0 16h9v4H22Z" :fill="color('#8d789a')"/>
      <circle cx="32" cy="58" r="7" :fill="color('#eac0a0')"/>
    </g>
</template>

