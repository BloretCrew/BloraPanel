<script setup lang="ts">
import {computed,ref} from 'vue'
import AppIcon from '../app-host/AppIcon.vue'
import {useQuery} from '@tanstack/vue-query'
import {api} from '../services/api'
import {loadExternalSandboxApp} from '../app-host/registry'
import {useDesktop} from '../desktop/store'
interface Item{manifest:{appId:string;title:string;packageVersion:string;capabilities?:string[]};sha256:string;enabled:boolean}
interface CatalogItem{manifest:{appId:string;title:string;packageVersion:string;capabilities?:string[]};sha256:string;signaturePresent:boolean;size:number}
const props=defineProps<{viewTabId:string}>(),error=ref(''),items=useQuery({queryKey:['extensions'],queryFn:()=>api<{items:Item[]}>('/extensions')}),catalog=useQuery({queryKey:['extension-catalog'],queryFn:()=>api<{items:CatalogItem[]}>('/extensions/catalog')}),file=ref<File>(),desktop=useDesktop(),recovery=desktop.recovery!
const view=computed(()=>recovery.state.views[props.viewTabId]!),requestIds=computed<Record<string,string>>({get:()=>view.value.state.extensionRequestIds as unknown as Record<string,string>||{},set:value=>desktop.patchView(props.viewTabId,'extensionRequestIds',value)})
function requestKey(scope:string){return requestIds.value[scope]||(requestIds.value={...requestIds.value,[scope]:crypto.randomUUID()},requestIds.value[scope]!)}
function requestDone(scope:string){const next={...requestIds.value};delete next[scope];requestIds.value=next}
async function install(){
  if(!file.value)return
  try{
    if(file.value.size>16*1024*1024)throw new Error('包文件不能超过 16 MiB')
    const body=await file.value.text(),pkg=JSON.parse(body)
    if(typeof pkg.manifest?.appId!=='string'||typeof pkg.payload!=='string'||typeof pkg.sha256!=='string')throw new Error('请选择完整的 .blora-extension.json 包')
    const existing=items.data.value?.items.find(x=>x.manifest.appId===pkg.manifest.appId)
    const path=existing?`/extensions/${encodeURIComponent(pkg.manifest.appId)}/upgrade`:'/extensions/install-package'
    const scope=`package:${pkg.manifest.appId}:${pkg.manifest.packageVersion}`
    await api(path,{method:'POST',headers:{'Idempotency-Key':requestKey(scope)},body})
    requestDone(scope)
    file.value=undefined
    error.value=''
    await items.refetch()
  }catch(e){error.value=String(e)}
}
async function toggle(x:Item){const scope=`enabled:${x.manifest.appId}:${!x.enabled}`;try{await api(`/extensions/${encodeURIComponent(x.manifest.appId)}/enabled`,{method:'POST',headers:{'Idempotency-Key':requestKey(scope)},body:JSON.stringify({enabled:!x.enabled})});requestDone(scope);await items.refetch()}catch(e){error.value=String(e)}}
async function uninstall(x:Item){if(!confirm(`卸载 ${x.manifest.title}？用户数据将保留。`))return;const scope=`uninstall:${x.manifest.appId}`;try{await api(`/extensions/${encodeURIComponent(x.manifest.appId)}`,{method:'DELETE',headers:{'Idempotency-Key':requestKey(scope)}});requestDone(scope);await items.refetch()}catch(e){error.value=String(e)}}
async function rollback(x:Item){if(!confirm(`回滚 ${x.manifest.title} 到上一版本？`))return;const scope=`rollback:${x.manifest.appId}`;try{await api(`/extensions/${encodeURIComponent(x.manifest.appId)}/rollback`,{method:'POST',headers:{'Idempotency-Key':requestKey(scope)}});requestDone(scope);await items.refetch()}catch(e){error.value=String(e)}}
async function openExtension(x:Item){try{await loadExternalSandboxApp(x.manifest.appId);desktop.open({appId:x.manifest.appId,disposition:'new-window'})}catch(e){error.value=String(e)}}
async function installCatalog(x:CatalogItem){const scope=`catalog:${x.manifest.appId}:${x.manifest.packageVersion}`;try{await api('/extensions/catalog/install',{method:'POST',headers:{'Idempotency-Key':requestKey(scope)},body:JSON.stringify({appId:x.manifest.appId,version:x.manifest.packageVersion})});requestDone(scope);await items.refetch()}catch(e){error.value=String(e)}}
</script>
<template><div class="application market-app">
<header class="app-heading"><div><h2>应用市场</h2><p>浏览注册源中的应用，按需安装到你的工作空间。</p></div><button @click="catalog.refetch();items.refetch()">刷新</button></header>
<p v-if="error||items.error.value" class="error">{{error||items.error.value?.message}}</p>
<section><h3>可安装应用</h3><p v-if="catalog.error.value" class="error">无法读取应用源：{{catalog.error.value.message}}</p><p v-else-if="catalog.isPending.value" class="muted">正在读取应用源…</p><div class="market-grid"><article v-for="x in catalog.data.value?.items" :key="x.manifest.appId+'@'+x.manifest.packageVersion" class="market-card" :data-app-id="x.manifest.appId"><span class="market-symbol"><AppIcon :app-id="x.manifest.appId" /></span><h3>{{x.manifest.title}}</h3><p>版本 {{x.manifest.packageVersion}} · {{x.signaturePresent?'已签名':'未签名'}}</p><button class="primary" @click="installCatalog(x)">安装此版本</button></article></div><p v-if="!catalog.isPending.value&&!catalog.error.value&&!catalog.data.value?.items.length" class="muted">当前应用源没有可用应用。管理员可配置注册源或从本地安装应用包。</p></section>
<section class="market-section"><h3>已安装</h3><article v-for="x in items.data.value?.items" :key="x.manifest.appId" class="task-row" :data-app-id="x.manifest.appId"><span><strong>{{x.manifest.title}}</strong> · v{{x.manifest.packageVersion}} · {{x.enabled?'已启用':'已禁用'}}</span><span class="action-row"><button :disabled="!x.enabled" @click="openExtension(x)">打开</button><button @click="toggle(x)">{{x.enabled?'禁用':'启用'}}</button><button @click="rollback(x)">回滚</button><button @click="uninstall(x)">卸载</button></span></article><p v-if="!items.isPending.value&&!items.data.value?.items.length" class="muted">还没有安装应用。</p></section>
<form class="action-row market-upload" @submit.prevent="install"><label>从本地安装<input aria-label="扩展包" type="file" accept=".json" required @change="file=($event.target as HTMLInputElement).files?.[0]"></label><button>安装或升级包</button></form>
</div></template>
