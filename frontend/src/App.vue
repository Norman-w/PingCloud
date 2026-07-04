<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { IconHome, IconTournament, IconUserPlus, IconConfetti, IconLogin } from '@tabler/icons-vue'

const router = useRouter()
const route = useRoute()

// ── SMS login ──
const myId = ref(0)
const myName = ref('')
const showLogin = ref(false)
const loginPhone = ref('')
const loginCode = ref('')
const loginStep = ref<'phone'|'code'|'setpw'>('phone')
const loginPw = ref('')
const loginSending = ref(false)
const loginMsg = ref('')

function loadMe() {
  fetch('/api/auth/me').then(r => r.ok ? r.json() : null).then(d => {
    if (d?.player_id) { myId.value = d.player_id; myName.value = d.player_name }
  }).catch(() => {})
}

async function sendCode() {
  if (!loginPhone.value) return
  loginSending.value = true; loginMsg.value = ''
  try {
    const r = await fetch('/api/auth/send-code', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({phone:loginPhone.value}) })
    if (!r.ok) { const t = await r.text(); loginMsg.value = t; return }
    loginStep.value = 'code'
  } catch { loginMsg.value = '发送失败' }
  finally { loginSending.value = false }
}

async function verifyCode() {
  if (!loginCode.value) return
  try {
    const r = await fetch('/api/auth/verify', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({phone:loginPhone.value, code:loginCode.value}) })
    if (!r.ok) { loginMsg.value = '验证码错误'; return }
    const d = await r.json()
    myId.value = d.player_id; myName.value = d.player_name
    if (d.need_setup) {
      loginStep.value = 'setup'
    } else {
      showLogin.value = false; loginPhone.value = ''; loginCode.value = ''; loginStep.value = 'phone'; loginMsg.value = ''
    }
  } catch { loginMsg.value = '验证失败' }
}

async function setupAccount() {
  // Called after user picks a username on first login
  if (!loginPw.value) return
  try {
    await fetch(`/api/admin/users`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: loginPw.value, display_name: myName.value, group_id: 4, player_id: myId.value }),
      credentials: 'include',
    })
    showLogin.value = false; loginPhone.value = ''; loginCode.value = ''; loginPw.value = ''; loginStep.value = 'phone'; loginMsg.value = ''
  } catch { loginMsg.value = '设置失败' }
}

function openLogin() { loginStep.value = 'phone'; loginPhone.value = ''; loginCode.value = ''; loginMsg.value = ''; showLogin.value = true }
function logout() { myId.value = 0; myName.value = ''; document.cookie = 'ping_id=;max-age=0;path=/' }

onMounted(() => { loadMe() })

const tabs = [
  { name: 'Home', label: '排位', component: IconHome },
  { name: 'FunMatch', label: '趣味赛', component: IconConfetti },
  { name: 'SessionView', label: '活动', component: IconTournament },
  { name: 'AddPlayer', label: '球员', component: IconUserPlus },
]

const active = computed(() => {
  const name = route.name as string
  return tabs.find(t => t.name === name) ? name : ''
})

const hideTabbar = computed(() => {
  const n = route.name as string
  return n === 'Admin' || n === 'AdminLogin'
})

function onTabChange(name: string) {
  router.push({ name })
}
</script>

<template>
  <div id="app-shell">
    <!-- Login indicator -->
    <div v-if="!hideTabbar" style="position:fixed;top:0;right:0;z-index:100;padding:6px 12px;">
      <button v-if="myId>0" @click="logout" style="background:rgba(255,255,255,0.9);border:1px solid #e0e0e0;border-radius:12px;padding:3px 10px;font-size:11px;color:#1989fa;cursor:pointer;box-shadow:0 1px 4px rgba(0,0,0,0.06);">🏓 {{ myName }}</button>
      <button v-else @click="openLogin" style="background:rgba(255,255,255,0.9);border:1px dashed #ccc;border-radius:12px;padding:3px 10px;font-size:11px;color:#999;cursor:pointer;">
        <IconLogin :size="14" style="vertical-align:-2px;" /> 登录</button>
    </div>

    <div class="app-body">
      <router-view />
    </div>
    <nav v-if="!hideTabbar" class="tabbar">
      <button
        v-for="tab in tabs"
        :key="tab.name"
        class="tabbar-item"
        :class="{ active: active === tab.name }"
        @click="onTabChange(tab.name)"
      >
        <component :is="tab.component" :size="24" :stroke-width="active === tab.name ? 2.5 : 1.5" />
        <span class="tab-label">{{ tab.label }}</span>
      </button>
    </nav>

    <!-- Login modal -->
    <div v-if="showLogin" style="position:fixed;inset:0;background:rgba(0,0,0,0.4);z-index:2000;display:flex;align-items:center;justify-content:center;" @click.self="showLogin=false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:300px;">
        <h3 style="text-align:center;margin-bottom:6px;">短信验证登录</h3>
        <div v-if="loginMsg" style="text-align:center;color:#e74c3c;font-size:12px;margin-bottom:8px;">{{ loginMsg }}</div>
        <!-- Phone step -->
        <template v-if="loginStep==='phone'">
          <input v-model="loginPhone" placeholder="输入手机号" type="tel" maxlength="11" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="sendCode" :disabled="loginSending" style="width:100%;padding:14px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">
            {{ loginSending ? '发送中...' : '获取验证码' }}</button>
        </template>
        <!-- Setup step (first time login) -->
        <template v-if="loginStep==='setup'">
          <div style="font-size:13px;color:#07c160;text-align:center;margin-bottom:8px;">🎉 首次登录！请设置登录账号</div>
          <input v-model="loginPw" placeholder="设置登录账号（英文/拼音）" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="setupAccount" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">确认设置</button>
        </template>

        <!-- Code step -->
        <template v-if="loginStep==='code'">
          <div style="font-size:13px;color:#666;text-align:center;margin-bottom:8px;">已发送至 {{ loginPhone }}</div>
          <input v-model="loginCode" placeholder="输入4位验证码" type="tel" maxlength="4" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:20px;font-weight:700;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;letter-spacing:8px;" />
          <button @click="verifyCode" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;margin-bottom:8px;">验证登录</button>
          <button @click="loginStep='phone'" style="width:100%;padding:10px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">重新发送</button>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
#app-shell {
  min-height: 100vh;
  background: #f0f2f5;
}

.app-body {
  padding-bottom: 64px;
}

.tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  padding: 6px 0 max(8px, env(safe-area-inset-bottom));
  z-index: 999;
}

.tabbar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 6px 4px;
  cursor: pointer;
  background: none;
  border: none;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  color: #969799;
  transition: color 0.2s;
}

.tabbar-item.active {
  color: #1989fa;
}

.tab-label {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.2px;
}

.tabbar-item.active .tab-label {
  font-weight: 700;
}
</style>
