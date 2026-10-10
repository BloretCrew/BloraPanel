<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useAttrs, useSlots, watch, type VNode } from 'vue'
import { Comment, Fragment } from 'vue'

defineOptions({ inheritAttrs: false })
const props=defineProps<{modelValue?:string|number;value?:string|number;disabled?:boolean;required?:boolean;name?:string}>()
const emit=defineEmits<{(event:'update:modelValue',value:string):(void);(event:'update:value',value:string):(void);(event:'change',value:Event):(void)}>()
const attrs=useAttrs(),slots=useSlots(),root=ref<HTMLElement>(),trigger=ref<HTMLButtonElement>(),open=ref(false),activeIndex=ref(0),popupStyle=ref<Record<string,string>>({}),typeahead=ref('')
const options=computed(()=>{
  const flattened:VNode[]=[]
  const visit=(nodes:VNode[])=>{for(const node of nodes){if(node.type===Comment)continue;if(node.type===Fragment){if(Array.isArray(node.children))visit(node.children as VNode[]);continue}if(typeof node.type==='string'&&node.type.toLowerCase()==='option')flattened.push(node)}}
  visit(slots.default?.()||[])
  return flattened.map((node,index)=>({value:String(node.props?.value??''),label:String(node.props?.label??textOf(node.children)).trim(),disabled:node.props?.disabled!==undefined&&node.props.disabled!==false&&node.props.disabled!==null,index}))
})
function textOf(value:unknown):string{if(typeof value==='string'||typeof value==='number')return String(value);if(Array.isArray(value))return value.map(item=>item&&typeof item==='object'&&'children'in item?textOf((item as VNode).children):'').join('');return ''}
const current=computed(()=>String(props.modelValue??props.value??''))
const selectedIndex=computed(()=>options.value.findIndex(option=>option.value===current.value))
const selectedLabel=computed(()=>options.value[selectedIndex.value]?.label||'')
const triggerAttrs=computed(()=>Object.fromEntries(Object.entries(attrs).filter(([key])=>key!=='class'&&key!=='style')))
const listBaseId=`styled-select-${Math.random().toString(36).slice(2)}`,popupGeneration=ref(0),listId=computed(()=>`${listBaseId}-${popupGeneration.value}`)
const activeOptionId=computed(()=>`${listId.value}-option-${options.value[activeIndex.value]?.index??0}`)
function enabledIndex(start:number,direction:number){if(!options.value.length)return -1;for(let step=0;step<options.value.length;step++){const index=(start+direction*step+options.value.length*2)%options.value.length;if(!options.value[index]!.disabled)return index}return -1}
function position(){const rect=trigger.value?.getBoundingClientRect();if(!rect)return;const width=Math.min(Math.max(rect.width,180),innerWidth-16),left=Math.max(8,Math.min(rect.left,innerWidth-width-8)),desired=Math.min(320,Math.max(72,options.value.length*36+10)),up=rect.bottom+desired+12>innerHeight&&rect.top>desired;const maxHeight=Math.max(100,Math.min(320,up?rect.top-20:innerHeight-rect.bottom-16));popupStyle.value={left:`${left}px`,top:`${up?Math.max(8,rect.top-Math.min(desired,maxHeight)-5):rect.bottom+5}px`,width:`${width}px`,maxHeight:`${maxHeight}px`}}
function show(){if(props.disabled||open.value)return;popupGeneration.value++;activeIndex.value=selectedIndex.value>=0&&!options.value[selectedIndex.value]?.disabled?selectedIndex.value:enabledIndex(0,1);position();open.value=true;void nextTick(position)}
function hide(restore=true){if(!open.value)return;open.value=false;if(restore)void nextTick(()=>trigger.value?.focus({preventScroll:true}))}
function select(index:number){const option=options.value[index];if(!option||option.disabled)return;emit('update:modelValue',option.value);emit('update:value',option.value);const event=new Event('change',{bubbles:true});Object.defineProperty(event,'target',{configurable:true,value:{value:option.value}});emit('change',event);hide()}
function keydown(event:KeyboardEvent){if(props.disabled)return;if(['ArrowDown','ArrowUp','Enter',' ','Home','End','Escape'].includes(event.key)){event.preventDefault();event.stopPropagation()}
  if(event.key==='Escape'){hide();return}
  if(event.key==='Tab'&&open.value){hide(false);return}
  if(!open.value){if(['ArrowDown','ArrowUp','Enter',' '].includes(event.key)){show();return}return}
  if(event.key==='ArrowDown')activeIndex.value=enabledIndex(activeIndex.value+1,1)
  else if(event.key==='ArrowUp')activeIndex.value=enabledIndex(activeIndex.value-1,-1)
  else if(event.key==='Home')activeIndex.value=enabledIndex(0,1)
  else if(event.key==='End')activeIndex.value=enabledIndex(options.value.length-1,-1)
  else if(event.key==='Enter'||event.key===' ')select(activeIndex.value)
  else if(event.key.length===1&&!event.altKey&&!event.ctrlKey&&!event.metaKey){typeahead.value+=event.key.toLocaleLowerCase();const start=options.value.findIndex((option,index)=>index>activeIndex.value&&!option.disabled&&option.label.toLocaleLowerCase().startsWith(typeahead.value));if(start>=0)activeIndex.value=start;window.setTimeout(()=>typeahead.value='',700)}
}
function outside(event:PointerEvent){if(open.value&&event.target instanceof Node&&!root.value?.contains(event.target)&&!document.getElementById(listId.value)?.contains(event.target))hide(false)}
function invalid(event:Event){event.preventDefault();trigger.value?.focus({preventScroll:true})}
watch(options,items=>{if(activeIndex.value>=items.length)activeIndex.value=enabledIndex(0,1)})
onMounted(()=>{document.addEventListener('pointerdown',outside,true);window.addEventListener('resize',position);window.addEventListener('scroll',position,true)})
onBeforeUnmount(()=>{document.removeEventListener('pointerdown',outside,true);window.removeEventListener('resize',position);window.removeEventListener('scroll',position,true)})
</script>
<template>
  <div ref="root" class="styled-select" :class="attrs.class" :style="attrs.style">
    <button ref="trigger" type="button" class="styled-select-trigger" role="combobox" aria-haspopup="listbox" :aria-expanded="open" :aria-controls="listId" :data-value="current" :aria-activedescendant="open?activeOptionId:undefined" :disabled="disabled" v-bind="triggerAttrs" @click="open?hide(false):show()" @keydown="keydown">
      <span class="styled-select-value">{{selectedLabel}}</span><span class="styled-select-chevron" aria-hidden="true"></span>
    </button>
    <input v-if="required||name" class="styled-select-validation" type="text" :name="name" :value="current" :required="required" tabindex="-1" aria-hidden="true" autocomplete="off" @invalid="invalid">
    <Teleport to="body"><Transition name="select-pop"><div v-if="open" :id="listId" class="styled-select-popover" role="listbox" :aria-label="String(attrs['aria-label']||selectedLabel||'选项')" :style="popupStyle">
      <button v-for="option in options" :id="`${listId}-option-${option.index}`" :key="`${option.value}:${option.index}`" type="button" class="styled-select-option" role="option" :aria-selected="option.value===current" :aria-disabled="option.disabled||undefined" :data-value="option.value" :disabled="option.disabled" @pointerenter="activeIndex=option.index" @click="select(option.index)"><span class="styled-select-check" aria-hidden="true">{{option.value===current?'✓':''}}</span><span>{{option.label}}</span></button>
    </div></Transition></Teleport>
  </div>
</template>
