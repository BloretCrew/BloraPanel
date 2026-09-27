<script setup lang="ts">
import {computed,ref} from 'vue'
import {useDesktop} from '../desktop/store'
import {session} from '../services/api'
import PasswordForm from './PasswordForm.vue'
import {isThemePalette,themePalette,themePalettes} from '../appearance/palette'
import {wallpaperStyle,wallpaperStyles} from '../appearance/wallpaper-style'
defineProps<{visible:boolean}>()
const desktop=useDesktop(),preferences=computed(()=>desktop.state!.preferences),notice=ref(''),error=ref(''),busy=ref(false)
const palette=computed(()=>themePalette(preferences.value))
const wallpaper=computed(()=>wallpaperStyle(preferences.value))
function set(key:string,value:string|boolean|number){
  const mutations:{kind:'set';path:string[];value:string|boolean|number}[]=[]
  if(key==='translucency' && !isThemePalette(preferences.value.palette)) mutations.push({kind:'set',path:['preferences','palette'],value:palette.value})
  mutations.push({kind:'set',path:['preferences',key],value})
  desktop.commit(mutations)
}
async function resetLayout(){
  if(busy.value)return
  if(!confirm('会先保存当前工作区副本，再恢复窗口和桌面入口的默认布局。继续吗？'))return
  busy.value=true;notice.value='';error.value=''
  try{const result=await desktop.resetLayout();notice.value=`已恢复默认布局，当前现场副本为“${result.title}”。`}
  catch(e){error.value=String(e)}finally{busy.value=false}
}
</script>
<style scoped>
.glass-switch-row span{display:grid;gap:3px}
.glass-switch-row small{color:var(--on-surface-variant);font-weight:400}
.glass-switch-row .glass-switch{display:inline-flex;align-items:center;justify-content:flex-start;width:44px;height:26px;min-width:44px;padding:2px;border:2px solid var(--outline);border-radius:var(--shape-full);background:var(--surface-highest);transition:background var(--motion-fast),border-color var(--motion-fast)}
.glass-switch-row .glass-switch::after{content:"";display:block;width:18px;height:18px;border-radius:50%;clip-path:none;opacity:1;background:var(--on-surface-variant);transition:transform var(--motion-fast),background var(--motion-fast)}
.glass-switch-row .glass-switch:checked{background:var(--primary);border-color:var(--primary)}
.glass-switch-row .glass-switch:checked::after{transform:translateX(18px);background:var(--on-primary)}
@media(forced-colors:active){.glass-switch-row .glass-switch{appearance:auto;width:20px;height:20px;min-width:20px;padding:0}.glass-switch-row .glass-switch::after{display:none}}
</style>
<template><div class="application"><header class="app-heading"><div><span class="eyebrow">PREFERENCES</span><h2>工作区设置</h2><p>设置随此工作现场保存在本浏览器。</p></div></header><section class="settings-panel"><h3>界面偏好</h3><label>主题<select :value="String(preferences.theme||'light')" @change="set('theme',($event.target as HTMLSelectElement).value)"><option value="light">浅色</option><option value="dark">深色</option></select></label><label>主题配色<select aria-label="主题配色" :value="palette" @change="set('palette',($event.target as HTMLSelectElement).value)"><option v-for="item in themePalettes" :key="item.id" :value="item.id">{{item.name}}</option></select><small class="muted">改变界面色彩；壁纸选“随配色”时也会匹配对应配色。</small></label><label>壁纸风格<select aria-label="壁纸风格" :value="wallpaper" @change="set('wallpaper',($event.target as HTMLSelectElement).value)"><option v-for="item in wallpaperStyles" :key="item.id" :value="item.id">{{item.name}}</option></select><small class="muted">可单独更换壁纸，不影响主题配色、通透材质或应用图标。</small></label><label>字号<select :value="String(preferences.fontScale||'1')" @change="set('fontScale',($event.target as HTMLSelectElement).value)"><option value="0.9">较小</option><option value="1">标准</option><option value="1.1">较大</option></select></label><label>界面密度<select :value="String(preferences.density||'comfortable')" @change="set('density',($event.target as HTMLSelectElement).value)"><option value="comfortable">舒适</option><option value="compact">紧凑</option></select></label><label>应用启动器快捷键<select :value="String(preferences.launcherShortcut||'ctrl-alt-n')" @change="set('launcherShortcut',($event.target as HTMLSelectElement).value)"><option value="ctrl-alt-n">Ctrl + Alt + N</option><option value="off">关闭</option></select></label><label>窗口切换快捷键<select :value="String(preferences.windowCycleShortcut||'alt-backtick')" @change="set('windowCycleShortcut',($event.target as HTMLSelectElement).value)"><option value="alt-backtick">Alt + `</option><option value="off">关闭</option></select></label><label class="settings-row glass-switch-row"><span>通透模式<small>只切换表面通透材质，不改变主题配色、壁纸或应用图标。</small></span><input class="glass-switch" type="checkbox" role="switch" aria-label="通透模式" :checked="preferences.translucency!==false" @change="set('translucency',($event.target as HTMLInputElement).checked)"></label><label class="settings-row">减少动态效果<input type="checkbox" :checked="!!preferences.reduceMotion" @change="set('reduceMotion',($event.target as HTMLInputElement).checked)"></label><p v-if="notice" class="notice" role="status">{{notice}}</p><p v-if="error" class="error" role="alert">{{error}}</p></section><section class="settings-panel"><h3>布局</h3><p class="muted">恢复默认布局前会自动保留完整工作区副本，草稿、终端检查点和资源现场不会被删除。</p><button :disabled="busy" @click="resetLayout">{{busy?'正在保存副本…':'恢复默认布局'}}</button></section><h3>恢复保护</h3><p :class="desktop.recovery!.status.protected?'muted':'error'">{{desktop.recovery!.status.message}}</p><p class="muted">正文草稿不会自动修改服务器文件。同步尾部预算为 256 KiB；超出或存储失败时会显示未保护状态。</p><dl class="details-grid"><dt>工作区</dt><dd>{{desktop.state!.workspaceId}}</dd><dt>浏览器标签</dt><dd>{{desktop.state!.browserTabId}}</dd><dt>本地版本</dt><dd>{{desktop.state!.revision}}</dd></dl><section class="account-settings"><PasswordForm v-if="session.user" :user="session.user" :visible="visible"/></section></div></template>
