<script setup lang="ts">
import { ref } from 'vue'
import { myId, doLogin as authLogin, checkName, completeLogin } from '../auth'

const visible = defineModel<boolean>('visible', { default: false })

const phone = ref(''); const code = ref('')
const step = ref<'phone'|'code'|'name'|'confirm'|'register'>('phone')
const sending = ref(false); const msg = ref('')
const countdown = ref(0); let timer: any = null

// Name step
const loginName = ref('')

// Confirm step
const matchedPlayers = ref<Array<{ id: number; name: string; gender: string; phone: string; current_rating: number; has_phone: boolean }>>([])

// Register step
const regName = ref('')
const regGender = ref('')
const regGrip = ref('')
const regRating = ref('')

const genderLabel = (g: string) => g === 'male' ? '男' : g === 'female' ? '女' : ''

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
  if (!result.ok) {
    msg.value = result.msg
    return
  }
  if (result.status === 'ok') {
    // Already bound phone — login directly
    visible.value = false; reset()
    return
  }
  // need_name — go to name step
  msg.value = ''
  step.value = 'name'
}

async function nextStep() {
  if (!loginName.value.trim()) { msg.value = '请输入姓名'; return }
  msg.value = ''
  try {
    const result = await checkName(phone.value.trim(), loginName.value.trim())
    if (result.matched && result.players.length > 0) {
      matchedPlayers.value = result.players
      step.value = 'confirm'
    } else {
      // No match — go to register with this name
      regName.value = loginName.value.trim()
      regGender.value = ''
      regGrip.value = ''
      regRating.value = ''
      step.value = 'register'
    }
  } catch (e: any) { msg.value = e.message || '查询失败' }
}

async function confirmBind(playerId: number) {
  const result = await completeLogin({ phone: phone.value.trim(), action: 'bind', bind_player_id: playerId })
  if (result.ok) {
    visible.value = false; reset()
  } else {
    msg.value = result.msg
  }
}

async function confirmNotMe() {
  // Auto-rename and go to register
  regName.value = loginName.value.trim() + '(2)'
  regGender.value = ''
  regGrip.value = ''
  regRating.value = ''
  step.value = 'register'
}

async function submitRegister() {
  if (!regName.value.trim()) { msg.value = '请输入姓名'; return }
  msg.value = ''
  const result = await completeLogin({
    phone: phone.value.trim(),
    action: 'create',
    name: regName.value.trim(),
    gender: regGender.value,
    grip: regGrip.value,
    initial_rating: regRating.value ? parseInt(regRating.value) : undefined,
  })
  if (result.ok) {
    visible.value = false; reset()
  } else {
    msg.value = result.msg
  }
}

function goBack() {
  if (step.value === 'code') { step.value = 'phone'; msg.value = '' }
  else if (step.value === 'name') { step.value = 'code'; msg.value = '' }
  else if (step.value === 'confirm') { step.value = 'name'; msg.value = '' }
  else if (step.value === 'register') { step.value = 'confirm'; msg.value = '' }
}

function reset() {
  step.value = 'phone'; phone.value = ''; code.value = ''; loginName.value = ''
  regName.value = ''; regGender.value = ''; regGrip.value = ''; regRating.value = ''
  matchedPlayers.value = []; msg.value = ''; countdown.value = 0
  if (timer) { clearInterval(timer); timer = null }
}

function close() { visible.value = false; reset() }
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:4000;display:flex;align-items:center;justify-content:center;" @click.self="close">
    <div style="background:#fff;border-radius:20px;padding:28px 24px;width:340px;max-width:90vw;max-height:85vh;overflow-y:auto;">
      <div style="text-align:center;font-size:40px;margin-bottom:8px;">🏓</div>
      <h3 style="text-align:center;margin-bottom:4px;font-size:18px;">登录乒云</h3>
      <div v-if="step==='phone'" style="text-align:center;color:#999;font-size:13px;margin-bottom:16px;">短信验证，安全快捷</div>
      <div v-else-if="step==='code'" style="text-align:center;color:#999;font-size:13px;margin-bottom:8px;">已发送至 {{ phone }}</div>
      <div v-else-if="step==='name'" style="text-align:center;color:#999;font-size:13px;margin-bottom:16px;">验证通过，请输入姓名</div>
      <div v-else-if="step==='confirm'" style="text-align:center;color:#999;font-size:13px;margin-bottom:16px;">找到同名球员</div>
      <div v-else-if="step==='register'" style="text-align:center;color:#999;font-size:13px;margin-bottom:16px;">完善信息，完成注册</div>

      <div v-if="msg && step!=='confirm'" style="text-align:center;font-size:12px;margin-bottom:8px;" :style="{color:msg.includes('错误')||msg.includes('失败')?'#e74c3c':'#07c160'}">{{ msg }}</div>

      <!-- Phone step -->
      <div v-if="step==='phone'">
        <input v-model="phone" placeholder="手机号" type="tel" maxlength="11" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;text-align:center;" />
        <button @click="send" :disabled="sending||countdown>0" style="width:100%;padding:14px;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;box-shadow:0 4px 12px rgba(25,137,250,0.3);" :style="{background:(sending||countdown>0)?'#ccc':'linear-gradient(135deg,#1989fa,#1e88e5)',cursor:(sending||countdown>0)?'not-allowed':'pointer'}">{{ sending?'发送中...':countdown>0?`重新发送(${countdown}s)`:'获取验证码' }}</button>
      </div>

      <!-- Code step -->
      <div v-else-if="step==='code'">
        <input v-model="code" placeholder="输入验证码" type="tel" maxlength="4" inputmode="numeric" pattern="[0-9]*" style="width:100%;padding:14px;border:1.5px solid #1989fa;border-radius:12px;font-size:22px;font-weight:700;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;letter-spacing:10px;" />
        <button @click="verify" style="width:100%;padding:14px;background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">验证</button>
        <button @click="send" :disabled="countdown>0" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;font-size:13px;cursor:pointer;" :style="{color:countdown>0?'#ccc':'#1989fa',cursor:countdown>0?'not-allowed':'pointer'}">{{ countdown>0?`${countdown}s后重发`:'重新发送' }}</button>
        <button @click="goBack" style="width:100%;padding:8px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">更换手机号</button>
      </div>

      <!-- Name step (new) -->
      <div v-else-if="step==='name'">
        <input v-model="loginName" placeholder="输入你的姓名" style="width:100%;padding:14px;border:1.5px solid #1989fa;border-radius:12px;font-size:18px;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;" />
        <button @click="nextStep" style="width:100%;padding:14px;background:linear-gradient(135deg,#1989fa,#1e88e5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;box-shadow:0 4px 12px rgba(25,137,250,0.3);">下一步</button>
        <button @click="goBack" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">返回</button>
      </div>

      <!-- Confirm step (new) — show matched players -->
      <div v-else-if="step==='confirm'">
        <div v-for="p in matchedPlayers" :key="p.id" style="background:#f8f9fa;border-radius:12px;padding:14px;margin-bottom:10px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px;">
            <span style="font-weight:700;font-size:16px;">{{ p.name }}</span>
            <span style="font-size:13px;color:#1989fa;font-weight:600;">{{ p.current_rating }} 分</span>
          </div>
          <div style="font-size:13px;color:#666;margin-bottom:8px;">
            {{ genderLabel(p.gender) || '未知性别' }}
            <span v-if="p.has_phone" style="color:#e74c3c;margin-left:8px;">📱 已绑定手机</span>
            <span v-else style="color:#07c160;margin-left:8px;">📱 未绑定手机</span>
          </div>
          <div v-if="p.has_phone" style="font-size:12px;color:#e74c3c;margin-bottom:6px;">该账号已绑定手机号，无法认领</div>
          <button v-else @click="confirmBind(p.id)" style="width:100%;padding:10px;background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;">这是我</button>
        </div>
        <div style="text-align:center;margin-top:4px;font-size:13px;color:#666;">这是你吗？</div>
        <button @click="confirmNotMe" style="width:100%;padding:12px;margin-top:10px;background:#f5f5f5;border:none;border-radius:12px;font-size:14px;color:#666;cursor:pointer;">都不是我，创建新账号</button>
        <button @click="goBack" style="width:100%;padding:8px;margin-top:4px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">返回修改姓名</button>
      </div>

      <!-- Register step (new) — phone readonly + registration form -->
      <div v-else-if="step==='register'">
        <!-- Phone (readonly) -->
        <div style="margin-bottom:12px;">
          <label style="font-size:13px;color:#999;display:block;margin-bottom:4px;">手机号（已验证）</label>
          <input :value="phone" disabled style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#f5f5f5;color:#999;text-align:center;" />
        </div>
        <!-- Name -->
        <div style="margin-bottom:12px;">
          <label style="font-size:13px;color:#999;display:block;margin-bottom:4px;">姓名</label>
          <input v-model="regName" placeholder="输入姓名" style="width:100%;padding:12px;border:1.5px solid #1989fa;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;text-align:center;" />
        </div>
        <!-- Gender -->
        <div style="margin-bottom:12px;">
          <label style="font-size:13px;color:#999;display:block;margin-bottom:6px;">性别</label>
          <div style="display:flex;gap:8px;">
            <button @click="regGender='male'" style="flex:1;padding:10px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;" :style="regGender==='male'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">男</button>
            <button @click="regGender='female'" style="flex:1;padding:10px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;" :style="regGender==='female'?{borderColor:'#ee0a24',color:'#ee0a24',background:'#fde8ef'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">女</button>
          </div>
        </div>
        <!-- Grip -->
        <div style="margin-bottom:12px;">
          <label style="font-size:13px;color:#999;display:block;margin-bottom:6px;">握拍方式</label>
          <div style="display:flex;gap:8px;">
            <button @click="regGrip='penhold'" style="flex:1;padding:10px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;" :style="regGrip==='penhold'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">直板</button>
            <button @click="regGrip='shakehand'" style="flex:1;padding:10px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;" :style="regGrip==='shakehand'?{borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'}:{borderColor:'#ddd',color:'#999',background:'#fff'}">横板</button>
          </div>
        </div>
        <!-- Initial rating -->
        <div style="margin-bottom:12px;">
          <label style="font-size:13px;color:#999;display:block;margin-bottom:4px;">初始积分（默认 1500）</label>
          <input v-model="regRating" type="number" placeholder="1500" style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;text-align:center;" />
        </div>
        <button @click="submitRegister" style="width:100%;padding:14px;background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">完成注册</button>
        <button @click="goBack" style="width:100%;padding:8px;margin-top:6px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">返回</button>
      </div>

      <button v-if="step==='phone'" @click="close" style="width:100%;padding:10px;margin-top:8px;background:none;border:none;color:#999;font-size:14px;cursor:pointer;">取消</button>
    </div>
  </div>
</template>
