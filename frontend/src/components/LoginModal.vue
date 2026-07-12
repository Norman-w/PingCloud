<script setup lang="ts">
import { ref } from 'vue'
import { myId, doLogin as authLogin } from '../auth'

const visible = defineModel<boolean>('visible', { default: false })

const phone = ref(''); const code = ref('')
const step = ref<'phone'|'code'>('phone')
const sending = ref(false); const msg = ref('')
const countdown = ref(0); let timer: any = null

async function send() {
  if (!phone.value.trim()) return
  sending.value = true; msg.value = ''
  try {
    const r = await fetch('/api/auth/send-code', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({phone:phone.value.trim()}) })
    if (!r.ok) { msg.value = await r.text() } else {
      step.value = 'code'; countdown.value = 60
      timer = setInterval(() => { countdown.value--; if (countdown.value <= 0) { clearInterval(timer); timer = null } }, 1000)
    }
  } catch { msg.value = '发送失败' } finally { sending.value = false }
}

async function verify() {
  if (!code.value) return
  const result = await authLogin(phone.value.trim(), code.value)
  if (result.ok) {
    visible.value = false; reset()
  } else {
    msg.value = result.msg
  }
}

function reset() {
  step.value = 'phone'; phone.value = ''; code.value = ''; msg.value = ''; countdown.value = 0
  if (timer) { clearInterval(timer); timer = null }
}

function close() { visible.value = false; reset() }
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:4000;display:flex;align-items:center;justify-content:center;" @click.self="close">
    <div style="background:#fff;border-radius:20px;padding:28px 24px;width:320px;max-width:90vw;">
      <div style="text-align:center;font-size:40px;margin-bottom:8px;">🏓</div>
      <h3 style="text-align:center;margin-bottom:4px;font-size:18px;">登录乒云</h3>
      <div style="text-align:center;color:#999;font-size:13px;margin-bottom:16px;">短信验证，安全快捷</div>
      <div v-if="msg" style="text-align:center;font-size:12px;margin-bottom:8px;" :style="{color:msg.includes('错误')||msg.includes('失败')?'#e74c3c':'#07c160'}">{{ msg }}</div>
      <!-- Phone step -->
      <div v-if="step==='phone'">
        <input v-model="phone" placeholder="手机号" type="tel" maxlength="11" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;text-align:center;" />
        <button @click="send" :disabled="sending||countdown>0" style="width:100%;padding:14px;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;box-shadow:0 4px 12px rgba(25,137,250,0.3);" :style="{background:(sending||countdown>0)?'#ccc':'linear-gradient(135deg,#1989fa,#1e88e5)',cursor:(sending||countdown>0)?'not-allowed':'pointer'}">{{ sending?'发送中...':countdown>0?`重新发送(${countdown}s)`:'获取验证码' }}</button>
      </div>
      <!-- Code step -->
      <div v-else>
        <div style="font-size:13px;color:#666;text-align:center;margin-bottom:8px;">已发送至 {{ phone }}</div>
        <input v-model="code" placeholder="输入验证码" type="tel" maxlength="4" inputmode="numeric" pattern="[0-9]*" style="width:100%;padding:14px;border:1.5px solid #1989fa;border-radius:12px;font-size:22px;font-weight:700;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;letter-spacing:10px;" />
        <button @click="verify" style="width:100%;padding:14px;background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">验证登录</button>
        <button @click="send" :disabled="countdown>0" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;font-size:13px;cursor:pointer;" :style="{color:countdown>0?'#ccc':'#1989fa',cursor:countdown>0?'not-allowed':'pointer'}">{{ countdown>0?`${countdown}s后重发`:'重新发送' }}</button>
      </div>
      <button @click="close" style="width:100%;padding:10px;margin-top:8px;background:none;border:none;color:#999;font-size:14px;cursor:pointer;">取消</button>
    </div>
  </div>
</template>
