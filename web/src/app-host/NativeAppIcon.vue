<script setup lang="ts">
import {computed,useId} from 'vue'
import {continuousRectPath} from './surface-geometry'
const props=withDefaults(defineProps<{appId:string;padded?:boolean}>(),{padded:true})
const kind=computed(()=>props.appId.replace(/^blora\./,''))
const uid=useId().replace(/:/g,'')+'native'
const paint=(name:string)=>`url(#${uid}-${name})`
const tile=continuousRectPath(108,108,37,0,10,10)
const inner=continuousRectPath(96,86,20,0,16,26)
const gear=(cx:number,cy:number,outer:number,root:number,count:number)=>{
  const points:number[][]=[]
  for(let n=0;n<count;n++)for(const [offset,r] of [[-.43,root],[-.26,root],[-.19,outer],[.19,outer],[.26,root],[.43,root]]){
    const angle=(n+offset!)/count*Math.PI*2;points.push([cx+Math.sin(angle)*r!,cy-Math.cos(angle)*r!])
  }
  return points.map(([x,y],i)=>`${i?'L':'M'}${x!.toFixed(2)} ${y!.toFixed(2)}`).join(' ')+'Z'
}
const gearPath=gear(64,64,43,36,16)
</script>
<template>
  <svg class="native-app-art" :data-native-app="kind" :viewBox="padded?'0 0 128 128':'8 8 112 112'" fill="none" aria-hidden="true">
    <defs>
      <linearGradient :id="uid+'-paper'" x1="0" y1="0" x2=".35" y2="1"><stop stop-color="#fff"/><stop offset=".74" stop-color="#f7f9fb"/><stop offset="1" stop-color="#e5eaf0"/></linearGradient>
      <linearGradient :id="uid+'-silver'" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#f5f7f9"/><stop offset=".48" stop-color="#dce3e9"/><stop offset="1" stop-color="#bec9d4"/></linearGradient>
      <linearGradient :id="uid+'-blue'" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#56b8f3"/><stop offset=".45" stop-color="#3da5ed"/><stop offset="1" stop-color="#2587d6"/></linearGradient>
      <linearGradient :id="uid+'-indigo'" x1="0" y1="0" x2=".2" y2="1"><stop stop-color="#467fe2"/><stop offset="1" stop-color="#2859b5"/></linearGradient>
      <linearGradient :id="uid+'-charcoal'" x1="0" y1="0" x2=".4" y2="1"><stop stop-color="#3c4652"/><stop offset="1" stop-color="#242d38"/></linearGradient>
      <linearGradient :id="uid+'-screen'" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#202c38"/><stop offset="1" stop-color="#17232e"/></linearGradient>
      <linearGradient :id="uid+'-teal'" x1="0" y1="0" x2=".3" y2="1"><stop stop-color="#6eb7a2"/><stop offset="1" stop-color="#368c7d"/></linearGradient>
      <filter :id="uid+'-separation'" x="-20%" y="-20%" width="140%" height="150%" color-interpolation-filters="sRGB"><feDropShadow dx="0" dy="1.2" stdDeviation=".85" flood-color="#152f46" flood-opacity=".16"/></filter>
    </defs>

    <!-- The folder itself is the outer silhouette: no generic icon tile. -->
    <g v-if="kind==='files'">
      <path d="M12 39V30a7 7 0 0 1 7-7h28l11 12h48a10 10 0 0 1 10 10v51a10 10 0 0 1-10 10H22a10 10 0 0 1-10-10Z" fill="#2975b6"/>
      <path d="M14 35v-5a6 6 0 0 1 6-6h26l12 12h48a8 8 0 0 1 8 8v3H14Z" fill="#6dbbed"/>
      <rect x="24" y="32" width="71" height="60" rx="3" fill="#d1e8f7" transform="rotate(-5 59 62)"/>
      <rect x="25" y="37" width="76" height="55" rx="3" :fill="paint('paper')" transform="rotate(2 63 64)"/>
      <path d="M11 48q0-5 5-5h32l9 7h54q6 0 6 6l-3 43a9 9 0 0 1-9 8H22a9 9 0 0 1-9-8Z" :fill="paint('blue')" stroke="#2c8bcc" stroke-width=".8" :filter="paint('separation')"/>
      <path d="M16 44h31l10 7h54" stroke="#bde9ff" stroke-opacity=".85" stroke-width="1"/>
      <path d="M18 101q2 4 7 4h77q6 0 7-4" stroke="#145eac" stroke-opacity=".3" stroke-width="1"/>
    </g>

    <g v-else-if="kind==='editor'">
      <path d="M25 13h56l18 18v78a6 6 0 0 1-6 6H25a6 6 0 0 1-6-6V19a6 6 0 0 1 6-6Z" :fill="paint('paper')" stroke="#c9d3dd" stroke-width="1" :filter="paint('separation')"/>
      <path d="M81 13v17a3 3 0 0 0 3 3h15" fill="#dce6ef"/>
      <path d="M81 13v17q0 3 3 3h15" stroke="#b9c8d7" stroke-width=".8"/>
      <path d="M30 24v80" stroke="#bad4e9" stroke-width="1"/>
      <path d="M39 40h30M39 49h42M39 58h39M39 67h30M39 76h40M39 85h25M39 94h30" stroke="#aab7c6" stroke-width="2.2" stroke-linecap="round"/>
      <g transform="rotate(35 88 63)" :filter="paint('separation')">
        <path d="M82 19q6-4 12 0v64H82Z" fill="#3d588b"/>
        <path d="M82 19q6-4 12 0v12H82Z" fill="#233652"/>
        <path d="M82 31h12v3H82Z" :fill="paint('silver')"/>
        <path d="M84 35v45" stroke="#7c95bf" stroke-width="1.7"/>
        <path d="M82 83h12l-3 11-3 9-3-9Z" fill="#e7edf2"/>
        <path d="M88 89v14l-3-9Z" fill="#74849c"/><circle cx="88" cy="90" r="1.2" fill="#596a83"/>
        <path d="M91 20v16" stroke="#c6d3e4" stroke-width="1.8" stroke-linecap="round"/>
      </g>
    </g>

    <g v-else-if="kind==='terminal'">
      <path :d="tile" :fill="paint('charcoal')" stroke="#566374" stroke-width="1"/>
      <path :d="inner" :fill="paint('screen')" stroke="#161f2b" stroke-width=".8"/>
      <path d="M30 20h3m5 0h3m5 0h3" stroke="#8393a5" stroke-width="2.6" stroke-linecap="round"/>
      <path d="M28 48 47 65 28 82" stroke="#f3f7fa" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M62 83h27" stroke="#88cfbd" stroke-width="5" stroke-linecap="round"/>
      <path d="M37 11h54" stroke="#82909d" stroke-opacity=".5"/>
    </g>

    <g v-else-if="kind==='extensions'">
      <path d="M43 38V27a21 21 0 0 1 42 0v11" stroke="#75a7e8" stroke-width="7"/>
      <path d="M44 35V27a20 20 0 0 1 40 0v8" stroke="#c3daf5" stroke-width="2"/>
      <path d="M25 33h78l8 68q1 12-12 12H29q-13 0-12-12Z" :fill="paint('indigo')" stroke="#2d5bae" stroke-width="1"/>
      <path d="M26 34h76" stroke="#91b7ed" stroke-width="1"/>
      <path d="M101 40l8 61q0 9-9 10l-3-16Z" fill="#174b9f" opacity=".3"/>
      <g :filter="paint('separation')">
        <path d="m64 51 25 14-25 14-25-14Z" fill="#f0f7ff"/>
        <path d="m39 65 23 13v25L39 89Z" fill="#c9e0f9"/>
        <path d="m66 78 23-13v24l-23 14Z" fill="#9cc6f0"/>
        <path d="m64 51 25 14-12 7-25-14Z" fill="#ffffff" opacity=".55"/>
        <path d="m64 77 12-7v11l-12 7Z" fill="#fff" opacity=".46"/>
      </g>
      <path d="M24 105q1 6 8 6h64" stroke="#174995" stroke-opacity=".4"/>
    </g>

    <g v-else-if="kind==='instances'">
      <path d="m20 27 13-10h63l13 10v23l-9 7 9 8v30l-13 13H32L19 97V66l9-9-8-7Z" fill="#3b536f"/>
      <path d="m20 27 13-10h63l13 10Z" fill="#c4d5e4"/>
      <rect x="20" y="27" width="89" height="28" rx="6" :fill="paint('silver')" stroke="#91a7bd" stroke-width=".8"/>
      <rect x="20" y="62" width="89" height="35" rx="6" :fill="paint('silver')" stroke="#91a7bd" stroke-width=".8"/>
      <path d="M25 30h78M25 65h78" stroke="#fff" stroke-opacity=".7"/>
      <rect x="28" y="35" width="49" height="11" rx="3" fill="#6c84a2"/>
      <rect x="28" y="72" width="49" height="14" rx="3" fill="#536d90"/>
      <path d="M33 39h37m-37 4h37M33 77h37m-37 4h37" stroke="#9cb0c7" stroke-width="1.1"/>
      <circle cx="95" cy="40" r="3.5" fill="#548bac"/><circle cx="95" cy="78" r="3.5" fill="#5fa48b"/>
      <circle cx="95" cy="39" r="1.2" fill="#b8dfea"/><circle cx="95" cy="77" r="1.2" fill="#c4ead8"/>
      <path d="M31 103h67" stroke="#647e98" stroke-width="2" stroke-linecap="round"/>
    </g>

    <g v-else-if="kind==='backups'">
      <rect x="22" y="14" width="83" height="99" rx="17" :fill="paint('silver')" stroke="#aabac9" stroke-width="1"/>
      <circle cx="63.5" cy="61" r="31" fill="#eff3f6" stroke="#c1cdd7" stroke-width="1.4"/>
      <circle cx="63.5" cy="61" r="16" fill="#d6dfe7"/><circle cx="63.5" cy="61" r="5" fill="#93a7b8"/>
      <path d="M39 99h27" stroke="#a4b6c5" stroke-width="3" stroke-linecap="round"/>
      <circle cx="91" cy="100" r="2.5" fill="#5d9f8b"/>
      <path d="M40 43a29 29 0 1 1-2 32" stroke="#479d88" stroke-width="7" stroke-linecap="round"/>
      <path d="M27 29v23h23Z" fill="#479d88"/>
    </g>

    <g v-else-if="kind==='monitor'">
      <path :d="tile" :fill="paint('charcoal')" stroke="#526475" stroke-width="1"/>
      <path d="M22 30h84v71H22Z" :fill="paint('screen')"/>
      <path d="M23 44h82M23 62h82M23 80h82M42 31v69M63 31v69M84 31v69" stroke="#426075" stroke-opacity=".4" stroke-width=".8"/>
      <path d="M24 86h17l12-37 14 48 15-58 12 31h11v31H24Z" fill="#6fbca4" opacity=".1"/>
      <path d="M24 86h17l12-37 14 48 15-58 12 31h11" stroke="#80d0b8" stroke-width="3.5" stroke-linejoin="round" stroke-linecap="round"/>
      <path d="M27 20h18" stroke="#889aa9" stroke-width="2" stroke-linecap="round"/>
      <circle cx="99" cy="20" r="2" fill="#6bbc9e"/>
    </g>

    <g v-else-if="kind==='tasks'">
      <rect x="22" y="13" width="90" height="104" rx="15" fill="#bccbd8"/>
      <rect x="18" y="10" width="90" height="104" rx="15" :fill="paint('paper')" stroke="#c9d5df" stroke-width="1"/>
      <path d="M33 11v102" stroke="#dbe4eb" stroke-width="1"/>
      <path d="M18 37h90M18 63h90M18 89h90" stroke="#e7edf2" stroke-width=".8"/>
      <circle cx="48" cy="36" r="7" fill="#6a9f9a"/><circle cx="48" cy="62" r="7" fill="#7a92ba"/><circle cx="48" cy="88" r="7" fill="#c49a64"/>
      <path d="m45 36 2 2 4-5m-6 29 2 2 4-5" stroke="#fff" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
      <path d="M65 35h26M65 61h20M65 87h25" stroke="#8b9aa8" stroke-width="3" stroke-linecap="round"/>
      <path d="M65 42h18M65 68h26M65 94h16" stroke="#cbd4dc" stroke-width="1.7" stroke-linecap="round"/>
    </g>

    <g v-else-if="kind==='settings'">
      <path :d="tile" :fill="paint('silver')" stroke="#bdc9d3" stroke-width="1"/>
      <circle cx="64" cy="64" r="43" fill="#657689"/>
      <path :d="gearPath" :fill="paint('paper')" stroke="#b5c1ce" stroke-width=".75" stroke-linejoin="round" :filter="paint('separation')"/>
      <circle cx="64" cy="64" r="27" fill="#a7b6c4" stroke="#e7eef4" stroke-width="1"/>
      <circle cx="64" cy="64" r="22" fill="#718399"/>
      <path d="m61 41 10 3-6 16 17 9-5 9-17-9-14 12-7-8 16-14Z" :fill="paint('silver')" stroke="#c5d1dc" stroke-width=".8"/>
      <circle cx="62" cy="63" r="7" fill="#e2eaf1" stroke="#7e93a8" stroke-width="1"/>
      <circle cx="62" cy="63" r="2" fill="#8396a9"/>
    </g>

    <g v-else-if="kind==='docker'">
      <path d="m64 13 46 24v55l-46 24-46-24V37Z" fill="#346f9a"/>
      <path d="m64 13 46 24-46 25-46-25Z" fill="#a0ceec"/>
      <path d="m18 37 46 25v54L18 92Z" fill="#5ba1d0"/>
      <path d="m64 62 46-25v55l-46 24Z" fill="#367eaf"/>
      <path d="M64 15v44M22 37l42 23 43-23" stroke="#d5edf9" stroke-width="1" stroke-opacity=".8"/>
      <path d="m41 25 46 24v18l-12 7V55L29 31Z" fill="#e5f1f9" opacity=".85"/>
      <path d="m27 63 28 15v10L27 73Z" fill="#d4eafa" opacity=".78"/>
      <path d="m27 81 18 9" stroke="#d8edfa" stroke-width="2" stroke-linecap="round"/>
    </g>

    <g v-else-if="kind==='nodes'">
      <path d="M64 40v19M32 82V59h64v23" stroke="#738ca8" stroke-width="5" stroke-linejoin="round"/>
      <rect x="43" y="12" width="42" height="32" rx="7" :fill="paint('indigo')" stroke="#31548d" stroke-width="1"/>
      <rect x="47" y="17" width="34" height="21" rx="3" fill="#c5e3f5"/>
      <path d="M55 25h18M55 30h9" stroke="#639bbc" stroke-width="2" stroke-linecap="round"/>
      <rect x="9" y="77" width="46" height="35" rx="7" :fill="paint('silver')" stroke="#a8bacb" stroke-width="1"/>
      <rect x="73" y="77" width="46" height="35" rx="7" :fill="paint('silver')" stroke="#a8bacb" stroke-width="1"/>
      <rect x="15" y="82" width="34" height="22" rx="3" :fill="paint('screen')"/>
      <rect x="79" y="82" width="34" height="22" rx="3" :fill="paint('screen')"/>
      <path d="m22 95 5-5 5 5 7-7m48 7 5-5 5 5 7-7" stroke="#8bc6ba" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>
      <circle cx="64" cy="59" r="5" fill="#90b2d2" stroke="#e4eef6" stroke-width="2"/>
    </g>

    <g v-else-if="kind==='users'">
      <path d="M30 13h66q12 0 12 12v79q0 12-12 12H30q-12 0-12-12V25q0-12 12-12Z" fill="#79658b"/>
      <path d="M33 13v103" stroke="#ac98bc" stroke-width="1"/>
      <path d="M108 31h7v16h-7M108 58h7v16h-7M108 84h7v16h-7" fill="#b09cbf"/>
      <circle cx="69" cy="47" r="15" fill="#e4dbe9"/>
      <path d="M43 96V85a26 26 0 0 1 52 0v11Z" fill="#c2b2d0"/>
      <path d="M22 29h15m-15 24h15m-15 24h15m-15 24h15" stroke="#d6c9e0" stroke-width="3" stroke-linecap="round"/>
      <path d="M39 113h56" stroke="#554966" stroke-opacity=".4"/>
    </g>

    <g v-else-if="kind==='system'">
      <path :d="tile" :fill="paint('charcoal')" stroke="#5f6e7d" stroke-width="1"/>
      <path d="M37 27v74M64 27v74M91 27v74" stroke="#101d29" stroke-width="6" stroke-linecap="round"/>
      <path d="M38 27v74M65 27v74M92 27v74" stroke="#53687c" stroke-width="1"/>
      <path d="M27 39h3m-3 14h3m-3 14h3m-3 14h3m34-42h3m-3 14h3m-3 14h3m-3 14h3" stroke="#8191a2" stroke-width="1"/>
      <rect x="28" y="42" width="18" height="16" rx="4" :fill="paint('silver')" stroke="#e8edf2" stroke-width=".6"/>
      <rect x="55" y="73" width="18" height="16" rx="4" fill="#7eadd1" stroke="#c0d7e8" stroke-width=".6"/>
      <rect x="82" y="32" width="18" height="16" rx="4" fill="#80b29e" stroke="#cee7dc" stroke-width=".6"/>
      <path d="M31 49h12M58 80h12M85 39h12" stroke="#456279" stroke-width="1"/>
    </g>

    <g v-else>
      <path :d="tile" :fill="paint('paper')" stroke="#d0dae3" stroke-width="1"/>
      <rect x="27" y="27" width="31" height="32" rx="9" fill="#629fc7"/>
      <rect x="69" y="27" width="31" height="32" rx="9" fill="#86b8a6"/>
      <rect x="27" y="70" width="31" height="32" rx="9" fill="#aea0c5"/>
      <rect x="69" y="70" width="31" height="32" rx="9" fill="#d4b181"/>
      <path d="M30 30q2-2 7-2h10m25 2q2-2 7-2h10M30 73q2-2 7-2h10m25 2q2-2 7-2h10" stroke="#fff" stroke-opacity=".45" stroke-width="1" stroke-linecap="round"/>
    </g>
  </svg>
</template>
<style scoped>
.native-app-art{display:block;width:100%;height:100%;overflow:visible;flex-shrink:0}
</style>
