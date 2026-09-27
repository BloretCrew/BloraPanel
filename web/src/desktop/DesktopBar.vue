<script setup lang="ts">
import {computed, ref} from 'vue'
import {onClickOutside, useDateFormat, useEventListener, useNow} from '@vueuse/core'
import {session} from '../services/api'
import UIIcon from '../app-host/UIIcon.vue'
import NotificationTray from './NotificationTray.vue'

defineProps<{workspace:string;activeTitle:string;protected:boolean;protectionMessage:string;workspaceOpen:boolean}>()
const emit=defineEmits<{workspace:[];export:[];logout:[];settings:[]}>()
const now=useNow({interval:1000}),time=useDateFormat(now,'HH:mm'),date=useDateFormat(now,'M月D日')
const accountOpen=ref(false),account=ref<HTMLElement>()
const initial=computed(()=>Array.from(session.user?.name||'U')[0]?.toLocaleUpperCase())
onClickOutside(account,()=>accountOpen.value=false)
useEventListener(window,'keydown',event=>{if(event.key==='Escape')accountOpen.value=false})
function accountAction(action:'settings'|'export'|'logout'){accountOpen.value=false;if(action==='settings')emit('settings');else if(action==='logout')emit('logout');else emit('export')}
</script>

<template>
  <header class="topbar" role="banner" aria-label="桌面控制栏">
    <div class="bar-identity">
      <span class="brand-lockup"><svg class="brand-symbol" viewBox="0 0 32 32" aria-hidden="true"><path d="M7 3h9a8 8 0 0 1 0 16H7Z" fill="var(--primary)"/><path d="M7 14h11a8 8 0 0 1 0 16H7Z" fill="var(--brand-secondary)"/><path d="M7 14h9a8 8 0 0 1 7 4H7Z" fill="var(--primary-container)"/></svg><span>Blora</span></span>
      <span class="bar-separator" aria-hidden="true"></span>
      <span class="active-context">{{activeTitle}}</span>
    </div>
    <button class="workspace-switch" aria-label="切换工作区" :aria-expanded="workspaceOpen" @click="emit('workspace')"><UIIcon name="workspace"/><span>{{workspace}}</span><UIIcon name="down" class="workspace-chevron"/></button>
    <div class="bar-utilities">
      <time class="bar-clock"><span>{{date}}</span><strong>{{time}}</strong></time>
      <span class="bar-separator" aria-hidden="true"></span>
      <div class="bar-tools"><button class="bar-icon-button" :class="{warning:!protected}" :aria-label="protected?'现场已保护，导出工作区':'保护异常，导出工作区'" :title="protectionMessage" @click="emit('export')"><UIIcon name="shield"/></button><NotificationTray/></div>
      <div ref="account" class="bar-account">
        <button class="account-trigger" aria-label="账号菜单" :aria-expanded="accountOpen" :title="session.user?.name" @click="accountOpen=!accountOpen"><span>{{initial}}</span></button>
        <section v-if="accountOpen" class="account-popover" role="menu" aria-label="账号菜单">
          <header><span class="account-avatar">{{initial}}</span><div><strong>{{session.user?.name}}</strong><small>{{session.user?.admin?'管理员':'成员'}}</small></div></header>
          <button role="menuitem" @click="accountAction('settings')"><UIIcon name="settings"/>桌面设置</button>
          <button role="menuitem" @click="accountAction('export')"><UIIcon name="download"/>导出工作现场</button>
          <button role="menuitem" class="account-logout" @click="accountAction('logout')"><UIIcon name="logout"/>退出账号</button>
        </section>
      </div>
    </div>
  </header>
</template>

<style scoped>
.topbar{position:absolute;z-index:1000;inset:14px 20px auto;height:52px;display:grid;grid-template-columns:minmax(0,1fr) auto minmax(0,1fr);align-items:center;padding:0 14px 0 18px;border:1px solid transparent;border-radius:26px;background:var(--surface-low);box-shadow:none;color:var(--on-surface)}

.bar-identity,.brand-lockup,.bar-utilities,.bar-tools,.workspace-switch{display:flex;align-items:center;min-width:0}
.bar-identity{gap:19px}
.brand-lockup{gap:9px;flex-shrink:0;font-size:18px;font-weight:650;letter-spacing:-.65px;color:var(--on-surface)}
.brand-symbol{width:28px;height:28px}
.bar-separator{display:block;width:1px;height:17px;background:var(--surface-container);flex-shrink:0}
.active-context{font-size:13px;font-weight:500;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;color:var(--muted)}

.workspace-switch{justify-content:center;gap:9px;padding:0 12px;height:32px;font-size:13px;max-width:240px;border:1px solid transparent;border-radius:var(--shape-full);color:var(--on-secondary-container);background:var(--secondary-container)}
.workspace-switch span{min-width:0;text-overflow:ellipsis;overflow:hidden;white-space:nowrap}
.workspace-switch:hover,.workspace-switch[aria-expanded=true]{background:var(--primary-container);border-color:transparent;color:var(--on-primary-container)}
.workspace-switch .workspace-chevron{width:12px;height:12px;margin-left:5px;color:var(--muted)}
.workspace-switch>.ui-icon:first-child{color:var(--muted)}

.bar-utilities{justify-content:flex-end;gap:15px}
.bar-clock{display:flex;align-items:baseline;gap:10px;font-size:11px;white-space:nowrap;color:var(--muted);font-variant-numeric:tabular-nums}
.bar-clock strong{font-size:13px;font-weight:550;color:var(--on-surface);letter-spacing:.1px}
.bar-tools{gap:3px}
.bar-icon-button,.bar-tools :deep(.notification-toggle){width:32px;height:32px;display:grid;place-items:center;padding:0;border:1px solid transparent;border-radius:50%;color:var(--on-surface-variant);background:transparent}
.bar-icon-button:hover,.bar-tools :deep(.notification-toggle:hover){background:var(--secondary-container);color:var(--on-secondary-container)}
.bar-tools :deep(.ui-icon){width:18px;height:18px;stroke-width:1.75}
.bar-icon-button.warning{color:var(--warning)}
.bar-account{position:relative;flex-shrink:0}
.account-trigger{height:32px;width:32px;padding:3px;border:1px solid transparent;border-radius:50%;background:var(--primary-container);box-shadow:none;display:grid;place-items:center}
.account-trigger span{font-size:11px;font-weight:600;color:var(--on-primary-container)}
.account-trigger:hover,.account-trigger[aria-expanded=true]{border-color:var(--primary)}

.account-popover{position:absolute;right:-3px;top:43px;width:232px;padding:7px;background:var(--surface-container);border:1px solid transparent;border-radius:var(--shape-lg-increased);box-shadow:var(--menu-shadow)}
.account-popover header{display:flex;align-items:center;gap:10px;padding:10px 9px 13px;margin-bottom:5px;border-bottom:1px solid var(--line)}
.account-avatar{display:grid;place-items:center;width:34px;height:34px;border-radius:50%;background:var(--primary-container);color:var(--on-primary-container);font-size:14px}
.account-popover strong,.account-popover small{display:block}
.account-popover strong{font-size:14px;font-weight:550}
.account-popover small{font-size:12px;color:var(--muted);margin-top:2px}
.account-popover button{display:flex;align-items:center;gap:10px;width:100%;height:35px;border-radius:var(--shape-full);padding:0 10px;font-size:13px;text-align:left;color:var(--on-surface)}
.account-popover button .ui-icon{width:16px;height:16px}
.account-popover button:hover{background:var(--secondary-container)}
.account-popover .account-logout{border-top:1px solid transparent;border-radius:var(--shape-full);margin-top:4px;padding-top:4px;height:38px;color:var(--error)}

@media(max-width:1000px){.bar-identity{gap:12px}
.bar-utilities{gap:9px}
.bar-clock>span{display:none}
.active-context{max-width:100px}
}

@media(max-width:640px){.topbar{inset:10px 10px auto;height:50px;padding:0 10px;grid-template-columns:auto minmax(0,1fr) auto;gap:9px;border-radius:var(--shape-lg)}
.brand-lockup>span,.bar-identity>.bar-separator,.active-context,.bar-clock,.bar-utilities>.bar-separator,.bar-tools>.bar-icon-button{display:none}
.brand-symbol{width:25px;height:25px}
.bar-utilities{gap:8px}
.workspace-switch{justify-self:start;max-width:100%;padding:0 3px;gap:6px}
.workspace-switch>.ui-icon:first-child{display:none}
.workspace-switch .workspace-chevron{margin-left:2px}
.bar-tools{gap:0}
.account-trigger{width:30px;height:30px}
.account-popover{right:-2px;top:42px}
.bar-tools :deep(.notification-toggle){width:30px;height:30px}
}

</style>
