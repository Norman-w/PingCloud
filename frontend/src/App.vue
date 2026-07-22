<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { IconHome, IconUserPlus, IconConfetti, IconLogin, IconSwords, IconUser, IconTrophy } from '@tabler/icons-vue'

const router = useRouter()
const route = useRoute()

// ── SMS login ──
const myId = ref(0)
const myName = ref('')
const showLogin = ref(false)
const loginPhone = ref('')
const loginCode = ref('')
const loginStep = ref<'phone'|'code'|'setpw'|'name'|'confirm'|'register'>('phone')
const loginPw = ref('')
const loginSending = ref(false)
const loginMsg = ref('')
const loginName = ref('')
const matchedPlayers = ref<Array<{ id: number; name: string; gender: string; phone: string; current_rating: number; has_phone: boolean }>>([])
const regName = ref('')
const regGender = ref('')
const regGrip = ref('')
const regRating = ref('')

function genderLabel(g: string) { return g === 'male' ? '男' : g === 'female' ? '女' : '' }

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
    if (d.status === 'ok') {
      // Phone already bound — login directly
      myId.value = d.player_id; myName.value = d.player_name
      if (d.need_setup) {
        loginStep.value = 'setpw'
      } else {
        showLogin.value = false; resetLogin()
      }
    } else {
      // need_name
      loginMsg.value = ''
      loginStep.value = 'name'
    }
  } catch { loginMsg.value = '验证失败' }
}

async function checkNameStep() {
  if (!loginName.value.trim()) { loginMsg.value = '请输入姓名'; return }
  loginMsg.value = ''
  try {
    const r = await fetch('/api/auth/check-name', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({phone:loginPhone.value, name:loginName.value.trim()}) })
    if (!r.ok) { loginMsg.value = await r.text(); return }
    const d = await r.json()
    if (d.matched && d.players.length > 0) {
      matchedPlayers.value = d.players
      loginStep.value = 'confirm'
    } else {
      regName.value = loginName.value.trim()
      regGender.value = ''; regGrip.value = ''; regRating.value = ''
      loginStep.value = 'register'
    }
  } catch { loginMsg.value = '查询失败' }
}

async function confirmBind(playerId: number) {
  try {
    const r = await fetch('/api/auth/complete', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({phone:loginPhone.value, action:'bind', bind_player_id:playerId}) })
    if (!r.ok) { loginMsg.value = await r.text(); return }
    const d = await r.json()
    myId.value = d.player_id; myName.value = d.player_name
    if (d.need_setup) { loginStep.value = 'setpw' }
    else { showLogin.value = false; resetLogin() }
  } catch { loginMsg.value = '绑定失败' }
}

async function confirmNotMe() {
  regName.value = loginName.value.trim() + '(2)'
  regGender.value = ''; regGrip.value = ''; regRating.value = ''
  loginStep.value = 'register'
}

async function submitRegister() {
  if (!regName.value.trim()) { loginMsg.value = '请输入姓名'; return }
  loginMsg.value = ''
  try {
    const r = await fetch('/api/auth/complete', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({phone:loginPhone.value, action:'create', name:regName.value.trim(), gender:regGender.value, grip:regGrip.value, initial_rating:regRating.value?parseInt(regRating.value):undefined}) })
    if (!r.ok) { loginMsg.value = await r.text(); return }
    const d = await r.json()
    myId.value = d.player_id; myName.value = d.player_name
    if (d.need_setup) { loginStep.value = 'setpw' }
    else { showLogin.value = false; resetLogin() }
  } catch { loginMsg.value = '注册失败' }
}

async function setupAccount() {
  if (!loginPw.value) return
  try {
    await fetch(`/api/admin/users`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: loginPw.value, display_name: myName.value, group_id: 4, player_id: myId.value }),
      credentials: 'include',
    })
    showLogin.value = false; resetLogin()
  } catch { loginMsg.value = '设置失败' }
}

function loginGoBack() {
  if (loginStep.value === 'code') { loginStep.value = 'phone'; loginMsg.value = '' }
  else if (loginStep.value === 'name') { loginStep.value = 'code'; loginMsg.value = '' }
  else if (loginStep.value === 'confirm') { loginStep.value = 'name'; loginMsg.value = '' }
  else if (loginStep.value === 'register') { loginStep.value = 'confirm'; loginMsg.value = '' }
}

function resetLogin() {
  loginPhone.value = ''; loginCode.value = ''; loginPw.value = ''; loginName.value = ''
  regName.value = ''; regGender.value = ''; regGrip.value = ''; regRating.value = ''
  matchedPlayers.value = []; loginStep.value = 'phone'; loginMsg.value = ''
}

function openLogin() { resetLogin(); showLogin.value = true }
function logout() { myId.value = 0; myName.value = ''; document.cookie = 'ping_id=;max-age=0;path=/' }

onMounted(() => { loadMe() })

const tabs = [
  { name: 'Home', label: '排位', component: IconHome },
  { name: 'TeamBattle', label: '团体赛', component: IconSwords },
  { name: 'FunMatch', label: '趣味赛', component: IconConfetti },
  { name: 'Tournament', label: '锦标赛', component: IconTrophy },
  { name: 'AddPlayer', label: '球员', component: IconUserPlus },
  { name: 'Me', label: '我的', component: IconUser },
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
      <div style="background:#fff;border-radius:16px;padding:24px;width:320px;max-width:90vw;max-height:85vh;overflow-y:auto;">
        <h3 style="text-align:center;margin-bottom:6px;">短信验证登录</h3>
        <div v-if="loginMsg && loginStep!=='confirm'" style="text-align:center;color:#e74c3c;font-size:12px;margin-bottom:8px;">{{ loginMsg }}</div>

        <!-- Phone step -->
        <div v-if="loginStep==='phone'">
          <input v-model="loginPhone" placeholder="输入手机号" type="tel" maxlength="11" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="sendCode" :disabled="loginSending" style="width:100%;padding:14px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">
            {{ loginSending ? '发送中...' : '获取验证码' }}</button>
        </div>

        <!-- Code step -->
        <div v-else-if="loginStep==='code'">
          <div style="font-size:13px;color:#666;text-align:center;margin-bottom:8px;">已发送至 {{ loginPhone }}</div>
          <input v-model="loginCode" placeholder="输入4位验证码" type="tel" maxlength="4" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:20px;font-weight:700;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;letter-spacing:8px;" />
          <button @click="verifyCode" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;margin-bottom:8px;">验证</button>
          <button @click="loginGoBack" style="width:100%;padding:10px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">更换手机号</button>
        </div>

        <!-- Name step (new) -->
        <div v-else-if="loginStep==='name'">
          <div style="font-size:13px;color:#666;text-align:center;margin-bottom:12px;">验证通过，请输入你的姓名</div>
          <input v-model="loginName" placeholder="输入姓名" style="width:100%;padding:12px;border:1.5px solid #1989fa;border-radius:10px;font-size:18px;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="checkNameStep" style="width:100%;padding:14px;background:linear-gradient(135deg,#1989fa,#1e88e5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">下一步</button>
          <button @click="loginGoBack" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">返回</button>
        </div>

        <!-- Confirm step (new) -->
        <div v-else-if="loginStep==='confirm'">
          <div style="font-size:14px;color:#666;text-align:center;margin-bottom:4px;">找到同名球员，这是你吗？</div>
          <div v-for="p in matchedPlayers" :key="p.id" style="background:#f8f9fa;border-radius:10px;padding:12px;margin-bottom:8px;margin-top:8px;">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px;">
              <span style="font-weight:700;font-size:15px;">{{ p.name }}</span>
              <span style="font-size:13px;color:#1989fa;font-weight:600;">{{ p.current_rating }} 分</span>
            </div>
            <div style="font-size:12px;color:#666;margin-bottom:6px;">
              {{ genderLabel(p.gender) || '未知性别' }}
              <span v-if="p.has_phone" style="color:#e74c3c;margin-left:6px;">📱 已绑定手机</span>
              <span v-else style="color:#07c160;margin-left:6px;">📱 未绑定手机</span>
            </div>
            <div v-if="p.has_phone" style="font-size:11px;color:#e74c3c;margin-bottom:6px;">该账号已绑定其他手机号，无法认领</div>
            <button v-else @click="confirmBind(p.id)" style="width:100%;padding:8px;background:#07c160;color:#fff;border:none;border-radius:8px;font-size:13px;font-weight:600;cursor:pointer;">这是我</button>
          </div>
          <button @click="confirmNotMe" style="width:100%;padding:10px;margin-top:6px;background:#f5f5f5;border:none;border-radius:10px;font-size:13px;color:#666;cursor:pointer;">都不是我，创建新账号</button>
          <button @click="loginGoBack" style="width:100%;padding:8px;margin-top:4px;background:none;border:none;color:#999;font-size:12px;cursor:pointer;">返回修改姓名</button>
        </div>

        <!-- Register step (new) -->
        <div v-else-if="loginStep==='register'">
          <div style="font-size:13px;color:#666;text-align:center;margin-bottom:8px;">完善信息，完成注册</div>
          <div style="margin-bottom:10px;">
            <label style="font-size:12px;color:#999;display:block;margin-bottom:3px;">手机号（已验证）</label>
            <input :value="loginPhone" disabled style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;background:#f5f5f5;color:#999;text-align:center;" />
          </div>
          <div style="margin-bottom:10px;">
            <label style="font-size:12px;color:#999;display:block;margin-bottom:3px;">姓名</label>
            <input v-model="regName" placeholder="输入姓名" style="width:100%;padding:10px;border:1.5px solid #1989fa;border-radius:8px;font-size:15px;outline:none;box-sizing:border-box;text-align:center;" />
          </div>
          <div style="margin-bottom:10px;">
            <label style="font-size:12px;color:#999;display:block;margin-bottom:4px;">性别</label>
            <div style="display:flex;gap:8px;">
              <button @click="regGender='male'" style="flex:1;padding:8px;border-radius:8px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;" :style="regGender==='male'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">男</button>
              <button @click="regGender='female'" style="flex:1;padding:8px;border-radius:8px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;" :style="regGender==='female'?{borderColor:'#ee0a24',color:'#ee0a24',background:'#fde8ef'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">女</button>
            </div>
          </div>
          <div style="margin-bottom:10px;">
            <label style="font-size:12px;color:#999;display:block;margin-bottom:4px;">握拍方式</label>
            <div style="display:flex;gap:8px;">
              <button @click="regGrip='penhold'" style="flex:1;padding:8px;border-radius:8px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;" :style="regGrip==='penhold'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">直板</button>
              <button @click="regGrip='shakehand'" style="flex:1;padding:8px;border-radius:8px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;" :style="regGrip==='shakehand'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">横板</button>
            </div>
          </div>
          <div style="margin-bottom:12px;">
            <label style="font-size:12px;color:#999;display:block;margin-bottom:3px;">初始积分（默认 1500）</label>
            <input v-model="regRating" type="number" placeholder="1500" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;text-align:center;" />
          </div>
          <button @click="submitRegister" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">完成注册</button>
          <button @click="loginGoBack" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">返回</button>
        </div>

        <!-- Setup step (first time login) -->
        <div v-else-if="loginStep==='setpw'">
          <div style="font-size:13px;color:#07c160;text-align:center;margin-bottom:8px;">🎉 首次登录！请设置登录账号</div>
          <input v-model="loginPw" placeholder="设置登录账号（英文/拼音）" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="setupAccount" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">确认设置</button>
        </div>
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
