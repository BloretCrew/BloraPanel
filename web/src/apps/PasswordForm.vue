<script setup lang="ts">
import {computed,onBeforeUnmount,ref,watch} from 'vue'
import {api,session,type User} from '../services/api'
import {passwordProblem} from '../services/accounts'
const props=defineProps<{user:User;visible:boolean}>(),emit=defineEmits<{changed:[]}>()
// Credentials intentionally live only in this component's memory. This form
// never calls workspace capture/commit, including on errors or unmount.
const currentPassword=ref(''),password=ref(''),confirmation=ref(''),requestId=ref(crypto.randomUUID()),busy=ref(false),error=ref(''),self=computed(()=>props.user.userId===session.user?.userId)
function clear(){currentPassword.value='';password.value='';confirmation.value=''}
watch(()=>props.visible,visible=>{if(!visible)clear()});watch(()=>props.user.userId,()=>{requestId.value=crypto.randomUUID();clear()});watch(()=>props.user.revision,()=>{requestId.value=crypto.randomUUID()});onBeforeUnmount(clear)
async function submit(){
  error.value=passwordProblem(password.value,confirmation.value);if(error.value||busy.value)return
  const userId=props.user.userId,actorId=session.user?.userId,isSelf=self.value,revision=props.user.revision;busy.value=true
  try{
    await api(`/users/${encodeURIComponent(userId)}/password`,{method:'POST',headers:{'Idempotency-Key':requestId.value},body:JSON.stringify({password:password.value,currentPassword:isSelf?currentPassword.value:undefined,revision})})
    requestId.value=crypto.randomUUID();clear();if(session.user?.userId!==actorId)return
    if(isSelf){session.error='密码已更改，原登录会话已撤销，请重新登录。';session.user=undefined}else emit('changed')
  }catch(e){error.value=String(e)}finally{clear();busy.value=false}
}
</script>
<template><form class="password-form" aria-label="更新账号密码" @submit.prevent="submit"><h3>{{self?'修改我的密码':'重置用户密码'}}</h3><p>{{user.name}} · {{user.userId}}</p><label v-if="self">当前密码<input v-model="currentPassword" type="password" aria-label="当前密码" autocomplete="current-password" required :disabled="busy"></label><label>新密码<input v-model="password" type="password" aria-label="新密码" autocomplete="new-password" required :disabled="busy"></label><label>确认新密码<input v-model="confirmation" type="password" aria-label="确认新密码" autocomplete="new-password" required :disabled="busy"></label><p class="muted">12～72字节。保存后撤销该账号的全部登录会话。密码字段在提交、隐藏或刷新后清空。</p><p v-if="error" class="error" role="alert">{{error}}</p><button class="primary" :disabled="busy">{{busy?'正在更新密码…':'保存新密码'}}</button></form></template>
