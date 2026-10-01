<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import { useEventListener } from '@vueuse/core'
import { api, checkSession, login, session } from './services/api'
import { pauseTaskReadsUntilPaint, resumeTaskReads } from './services/task-query-lifecycle'
import { useDesktop } from './desktop/store'
import Desktop from './desktop/Desktop.vue'
const desktop = useDesktop(), query=useQueryClient(), name = ref(''), password = ref(''), busy = ref(false), error = ref(''), initialized = ref('')
onMounted(checkSession)
watch(() => session.user?.userId, async (userId,previousUserId) => {
  initialized.value = ''
  await nextTick()
  query.clear()
  if(previousUserId){await desktop.recovery?.flush();const {releaseDocuments}=await import('./services/documents');releaseDocuments(previousUserId)}
  if (userId) { try{await desktop.initialize(userId);if(session.user?.userId===userId)initialized.value = userId}catch(e){error.value=String(e)} }
})
async function submit() { if(busy.value)return;const submittedPassword=password.value;password.value='';busy.value = true; error.value = ''; try { await login(name.value,submittedPassword) } catch(e) { error.value = e instanceof Error ? e.message : String(e) } finally { busy.value = false } }
async function logout() {
  if (!confirm('退出会清除此账号在本浏览器的本地工作区和草稿。请先导出需要保留的未同步内容。确认退出？')) return
  session.ready=false
  try{
    await api('/logout',{method:'POST'})
    const recovery=desktop.recovery
    session.user=undefined
    await nextTick()
    await recovery?.clearAccount()
  }catch(e){error.value='退出清理未完成：'+String(e)}finally{session.ready=true}
}
// Cancel read-only task polling while this document can still abort fetches.
// Leave task mutations and terminal/input recovery owned by their existing
// lifecycles. Gate new reads from timer callbacks already queued at departure.
// A surviving document resumes on paint or pageshow after restoration.
const cancelTaskReads=()=>{void query.cancelQueries({queryKey:['tasks']})}
useEventListener(window,'beforeunload',(event:BeforeUnloadEvent)=> { pauseTaskReadsUntilPaint();if(desktop.recovery && !desktop.recovery.status.protected) { event.preventDefault(); event.returnValue = '' };cancelTaskReads() })
useEventListener(window,'pagehide',cancelTaskReads)
useEventListener(window,'pageshow',resumeTaskReads)
</script>
<template>
  <main v-if="!session.user" class="login-screen">
    <div class="login-brand"><span class="brand-mark">B</span><p>BLORA PANEL</p><h1>让每一项工作，<br>都有自己的空间。</h1><p class="muted">连接节点，管理实例，接着上次的工作。</p></div>
    <form class="login-card" @submit.prevent="submit"><span class="eyebrow">欢迎回来</span><h2>登录管理桌面</h2><p class="muted">使用平台管理员为你创建的账号。</p><label>账号<input v-model="name" name="username" autocomplete="username" required autofocus></label><label>密码<input v-model="password" name="password" type="password" autocomplete="current-password" required></label><p v-if="error || session.error" role="alert" class="error">{{error || session.error}}</p><button class="primary" :disabled="busy || !session.ready">{{busy?'正在登录…':'进入工作区 →'}}</button><small>登录后恢复属于你的本地工作现场。</small></form>
  </main>
  <Desktop v-else-if="initialized === session.user.userId" @logout="logout" />
  <main v-else class="loading"><p v-if="error" role="alert" class="error">{{error}}</p><span v-else>正在恢复工作现场…</span></main>
</template>
