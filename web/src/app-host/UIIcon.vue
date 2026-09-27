<script setup lang="ts">
import {computed} from 'vue'
import {X, Minus, Maximize, Plus, ChevronDown, Bell, ShieldCheck, LogOut, History, Activity, Search, PanelLeft, UserRound, LayoutGrid, Download, SlidersHorizontal, Folder, Trash2, ChevronLeft, ChevronRight, Play, Square, Monitor, ExternalLink, MonitorUp, Server, FileText} from '@lucide/vue'

const props=defineProps<{name:string}>()
// Shared line-icon vocabulary. Import only used Lucide components.
const icons={close:X,minus:Minus,expand:Maximize,plus:Plus,down:ChevronDown,bell:Bell,shield:ShieldCheck,logout:LogOut,history:History,activity:Activity,search:Search,sidebar:PanelLeft,user:UserRound,workspace:LayoutGrid,download:Download,settings:SlidersHorizontal,folder:Folder,trash:Trash2,left:ChevronLeft,right:ChevronRight,play:Play,stop:Square,monitor:Monitor,external:ExternalLink,'pin-desktop':MonitorUp,server:Server,file:FileText}
const icon=computed(()=>icons[props.name as keyof typeof icons]||Plus)
</script>
<template><component :is="icon" class="ui-icon" :size="18" :stroke-width="1.75" aria-hidden="true"/></template>
