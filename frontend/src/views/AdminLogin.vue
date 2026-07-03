<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function login() {
  if (!username.value || !password.value) { error.value = '请输入用户名和密码'; return }
  loading.value = true; error.value = ''
  try {
    const res = await fetch('/api/admin/login', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
      credentials: 'include',
    })
    if (!res.ok) { const t = await res.text(); throw new Error(t) }
    router.replace('/admin')
  } catch (e: any) { error.value = '登录失败: ' + (e.message || '') }
  finally { loading.value = false }
}
</script>

<template>
  <div style="min-height:100vh;display:flex;align-items:center;justify-content:center;background:linear-gradient(135deg,#1a1a2e,#16213e,#0f3460);">
    <div style="background:#fff;border-radius:16px;padding:32px 24px;width:340px;box-shadow:0 8px 32px rgba(0,0,0,0.3);">
      <div style="text-align:center;margin-bottom:24px;">
        <div style="font-size:28px;font-weight:800;color:#1a1a2e;">🏓 PingCloud</div>
        <div style="font-size:14px;color:#999;margin-top:4px;">管理员登录</div>
      </div>

      <div v-if="error" style="background:#fff0f0;color:#e74c3c;padding:10px;border-radius:8px;font-size:13px;margin-bottom:12px;text-align:center;">{{ error }}</div>

      <div style="margin-bottom:12px;">
        <input v-model="username" placeholder="用户名" @keyup.enter="login"
          style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;" />
      </div>
      <div style="margin-bottom:20px;">
        <input v-model="password" type="password" placeholder="密码" @keyup.enter="login"
          style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;" />
      </div>

      <button @click="login" :disabled="loading"
        style="width:100%;padding:14px;background:linear-gradient(135deg,#1a56db,#1e88e5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;"
        :style="{opacity:loading?0.6:1}">
        {{ loading ? '登录中...' : '登录' }}
      </button>
    </div>
  </div>
</template>
