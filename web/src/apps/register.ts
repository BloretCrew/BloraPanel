import { defineAsyncComponent, type Component } from 'vue'
import { registerApp } from '../app-host/registry'
import { api } from '../services/api'
import type { AppManifest } from '../app-host/types'
import {copy} from '../recovery/state'
import AppLoading from '../app-host/AppLoading.vue'
const definitions: {appId:string;title:string;icon:string;color:string;resources:string[];component:()=>Promise<{default:Component}>}[] = [
  {appId:'blora.instances',title:'实例中心',icon:'▣',color:'#55cbaa',resources:['instance'],component:()=>import('./InstancesApp.vue')},
  {appId:'blora.backups',title:'备份与计划',icon:'⟳',color:'#d6a875',resources:['instance'],component:()=>import('./BackupApp.vue')},
  {appId:'blora.files',title:'文件管理器',icon:'▰',color:'#e7bd67',resources:['directory'],component:()=>import('./FilesApp.vue')},
  {appId:'blora.editor',title:'编辑器',icon:'⌘',color:'#92a7f5',resources:['file'],component:()=>import('./EditorApp.vue')},
  {appId:'blora.terminal',title:'终端',icon:'>_',color:'#86d0b5',resources:['terminal'],component:()=>import('./TerminalApp.vue')},
  {appId:'blora.tasks',title:'任务中心',icon:'☷',color:'#dea681',resources:['task'],component:()=>import('./TasksApp.vue')},
  {appId:'blora.nodes',title:'节点管理',icon:'⬡',color:'#7db9d8',resources:['node'],component:()=>import('./NodesApp.vue')},
  {appId:'blora.monitor',title:'监控与进程',icon:'◒',color:'#c8a8ed',resources:['node','instance'],component:()=>import('./MonitorApp.vue')},
  {appId:'blora.docker',title:'容器中心',icon:'▦',color:'#72b9e5',resources:['docker','compose'],component:()=>import('./DockerApp.vue')},
  {appId:'blora.users',title:'用户与权限',icon:'♙',color:'#d5a9dd',resources:['user'],component:()=>import('./UsersApp.vue')},
  {appId:'blora.settings',title:'设置',icon:'⚙',color:'#aab8ad',resources:[],component:()=>import('./SettingsApp.vue')},
  {appId:'blora.extensions',title:'应用市场',icon:'✦',color:'#e3a5c4',resources:[],component:()=>import('./ExtensionsApp.vue')},
  {appId:'blora.system',title:'系统管理',icon:'⌁',color:'#d6b67a',resources:['node'],component:()=>import('./SystemApp.vue')},
]
for(const definition of definitions){
  const manifest:AppManifest={appId:definition.appId,title:definition.title,icon:definition.icon,color:definition.color,packageVersion:'0.1.0',hostApiVersion:1,entrypoints:['overview','resource'],resourceHandlers:definition.resources,permissions:['blora.users','blora.extensions','blora.system'].includes(definition.appId)?['platform.admin']:[],dependencies:[],protected:true,windowPolicy:'multiple',tabPolicy:{types:['overview','resource'],movable:true},stateSchemaVersion:1}
  registerApp({manifest,component:defineAsyncComponent({loader:definition.component,loadingComponent:AppLoading,delay:0}),captureState:state=>copy(state),restoreState:state=>copy(state),migrateState:(state,from)=>{if(from!==1)throw new Error('应用状态版本不兼容');return state},reconcileResource:async resource=>{
    if(resource?.kind!=='instance') return undefined
    const response=await api<{instance:{instanceId:string;nodeId:string}}>(`/instances/${encodeURIComponent(resource.id)}`)
    return response.instance?.nodeId ? {...resource,nodeId:response.instance.nodeId} : undefined
  }})
}
