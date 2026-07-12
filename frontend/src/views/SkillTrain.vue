<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showSuccessToast } from 'vant'
import { IconArrowLeft, IconSearch, IconPlus, IconChevronDown, IconChevronUp, IconFlame, IconCrown, IconCircle } from '@tabler/icons-vue'

const route = useRoute(); const router = useRouter()
const skillId = Number(route.params.id)

// ── auth ──
const myId = ref(0); const myName = ref('')
const showLogin = ref(false)
const loginPhone = ref(''); const loginCode = ref(''); const loginSending = ref(false); const loginMsg = ref('')

// ── skill info ──
const skillName = ref(''); const skillCategory = ref('')
const skillStatus = ref('none'); const skillTags = ref<string[]>([])

// ── history ──
const history = ref<any[]>([]); const showAllHistory = ref(false)
const visibleHistory = computed(() => showAllHistory.value ? history.value : history.value.slice(0, 1))

// ── form ──
const formDate = ref(new Date().toISOString().slice(0, 10))
const formDuration = ref('60'); const formLocation = ref(''); const formPartner = ref('')
const formNotes = ref(''); const formPracticeAmount = ref(''); const formSkillNotes = ref('')
const formEnergyRating = ref(0); const formFeelRating = ref(0); const submitting = ref(false)

// ── location search ──
const locSearch = ref(''); const locResults = ref<{id:number;name:string}[]>([]); const showLocDropdown = ref(false)
async function searchLocations(q: string) {
  if (!q) { locResults.value = []; return }
  try { const r = await fetch(`/api/locations?q=${encodeURIComponent(q)}`); if (r.ok) locResults.value = await r.json() } catch {}
}
function selectLocation(name: string) { formLocation.value = name; showLocDropdown.value = false; locSearch.value = '' }
async function createLocation() {
  if (!locSearch.value.trim()) return
  try {
    const r = await fetch('/api/locations', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({name:locSearch.value.trim()}) })
    if (r.ok) { const l = await r.json(); formLocation.value = l.name; showLocDropdown.value = false; locSearch.value = '' }
  } catch {}
}

// ── player search ──
const playerSearch = ref(''); const playerResults = ref<{id:number;name:string}[]>([]); const showPlayerDropdown = ref(false)
async function searchPlayers(q: string) {
  if (!q) { playerResults.value = []; return }
  try { const r = await fetch(`/api/players?q=${encodeURIComponent(q)}`); if (r.ok) playerResults.value = await r.json() } catch {}
}
function selectPlayer(name: string) { formPartner.value = name; showPlayerDropdown.value = false; playerSearch.value = '' }

// ── radar chart ──
const indicators = ref<Record<string, number>>({})
const activeAxis = ref(-1); const dragging = ref(false)
const radarW = 300; const radarH = 300
const radarCX = 150; const radarCY = 140; const radarR = 100

const indicatorNames = computed(() => Object.keys(indicators.value))

function radarPoint(i: number, value: number): {x: number, y: number} {
  const n = indicatorNames.value.length; if (n === 0) return {x: radarCX, y: radarCY}
  const angle = (2 * Math.PI * i) / n - Math.PI / 2
  return { x: radarCX + (radarR * value / 5) * Math.cos(angle), y: radarCY + (radarR * value / 5) * Math.sin(angle) }
}

function axisEnd(i: number): {x: number, y: number} {
  const n = indicatorNames.value.length; if (n === 0) return {x: radarCX, y: radarCY}
  const angle = (2 * Math.PI * i) / n - Math.PI / 2
  return { x: radarCX + radarR * Math.cos(angle), y: radarCY + radarR * Math.sin(angle) }
}

function labelPos(i: number): {x: number, y: number, anchor: string} {
  const end = axisEnd(i); const dx = end.x - radarCX; const dy = end.y - radarCY
  const dist = Math.hypot(dx, dy); const nx = dx / dist; const ny = dy / dist
  const lx = radarCX + nx * (radarR + 24); const ly = radarCY + ny * (radarR + 20)
  let anchor = 'middle'
  if (Math.abs(nx) > 0.7) anchor = nx > 0 ? 'start' : 'end'
  return { x: lx, y: ly, anchor }
}

function gridPoints(level: number): string {
  return indicatorNames.value.map((_, i) => { const p = radarPoint(i, level); return `${p.x},${p.y}` }).join(' ')
}

function polygonPoints(vals?: Record<string, number>): string {
  const v = vals || indicators.value
  return indicatorNames.value.map((_, i) => { const p = radarPoint(i, v[indicatorNames.value[i]] || 1); return `${p.x},${p.y}` }).join(' ')
}

function setIndicatorValue(i: number, val: number) {
  const name = indicatorNames.value[i]; if (!name) return
  indicators.value = { ...indicators.value, [name]: Math.round(Math.max(1, Math.min(5, val))) }
}

// Pointer handling
const svgRef = ref<SVGSVGElement>()
function getPos(e: PointerEvent) {
  const svg = svgRef.value!; const r = svg.getBoundingClientRect()
  return { x: (e.clientX - r.left) / r.width * radarW, y: (e.clientY - r.top) / r.height * radarH }
}
function onDown(e: PointerEvent) {
  const p = getPos(e); let best = -1; let minD = Infinity
  for (let i = 0; i < indicatorNames.value.length; i++) {
    const ep = axisEnd(i); const d = Math.hypot(p.x - ep.x, p.y - ep.y)
    if (d < minD) { minD = d; best = i }
  }
  if (minD < 45) { activeAxis.value = best; dragging.value = true; updateFromPos(p); svgRef.value?.setPointerCapture(e.pointerId) }
}
function onMove(e: PointerEvent) { if (dragging.value && activeAxis.value >= 0) updateFromPos(getPos(e)) }
function onUp() { dragging.value = false; activeAxis.value = -1 }
function updateFromPos(p: {x: number, y: number}) {
  const i = activeAxis.value; if (i < 0) return
  const n = indicatorNames.value.length; const angle = (2 * Math.PI * i) / n - Math.PI / 2
  const proj = (p.x - radarCX) * Math.cos(angle) + (p.y - radarCY) * Math.sin(angle)
  setIndicatorValue(i, Math.round(proj / radarR * 5))
}

// ── login ──
async function checkLogin() {
  try {
    const r = await fetch('/api/auth/me')
    if (r.ok) { const d = await r.json(); if (d?.player_id) { myId.value = d.player_id; myName.value = d.player_name } }
  } catch {}
}
async function sendCode() {
  if (!loginPhone.value) return; loginSending.value = true; loginMsg.value = ''
  try {
    const r = await fetch('/api/auth/send-code', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({phone:loginPhone.value}) })
    if (!r.ok) { loginMsg.value = await r.text(); return }
    loginMsg.value = '验证码已发送'
  } catch { loginMsg.value = '发送失败' }
  finally { loginSending.value = false }
}
async function verifyCode() {
  if (!loginCode.value) return
  try {
    const r = await fetch('/api/auth/verify', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({phone:loginPhone.value, code:loginCode.value}) })
    if (!r.ok) { loginMsg.value = '验证码错误'; return }
    const d = await r.json()
    myId.value = d.player_id; myName.value = d.player_name
    showLogin.value = false; loginPhone.value = ''; loginCode.value = ''; loginMsg.value = ''
    // Reload data now that we're logged in
    await loadSkillData()
  } catch { loginMsg.value = '验证失败' }
}

// ── load ──
const loading = ref(true)
onMounted(async () => {
  await checkLogin()
  await loadSkillData()
})

async function loadSkillData() {
  loading.value = true
  try {
    const r = await fetch(`/api/skill-train/${skillId}`)
    if (r.ok) {
      const d = await r.json()
      skillName.value = d.skill_name; skillCategory.value = d.category; history.value = d.history || []
      indicators.value = (d.history?.[0]?.indicators && Object.keys(d.history[0].indicators).length > 0)
        ? d.history[0].indicators : defaults(skillId)
    }
  } catch {}
  try {
    const r = await fetch('/api/skill-mastery')
    if (r.ok) {
      const d = await r.json()
      const item = (d.skills || []).find((s:any) => s.id === skillId)
      if (item) { skillStatus.value = item.status; skillTags.value = item.tags || [] }
    }
  } catch {}
  loading.value = false
}

function defaults(id: number): Record<string, number> {
  // Variable indicator counts per skill (3-6 axes, not forced to 5)
  const sets: Record<number, string[]> = {
    // 基础入门 (knowledge, 3 indicators)
    30: ['理解度','兴趣','掌握'], 31: ['理解度','兴趣','掌握'], 32: ['理解度','兴趣','掌握'],
    33: ['理解度','兴趣','掌握'], 34: ['理解度','兴趣','掌握'],
    // 物理学原理 (theory, 3-4 indicators)
    35: ['理解','应用','观察'], 36: ['理解','应用','观察'],
    37: ['理解','应用','感受'], 38: ['理解','应用','精准'], 39: ['理解','应用','感受'],
    // 基本功 — attack
    1:  ['速度','力量','落点','旋转'],           // 正手攻球
    3:  ['旋转','速度','弧线','力量','落点'],    // 正手前冲弧圈
    4:  ['旋转','弧线','落点','控制'],            // 正手加转弧圈
    5:  ['旋转','速度','弧线','力量','落点'],    // 反手拉球
    10: ['速度','落点','时机','手感'],            // 台内挑打
    11: ['旋转','弧线','落点','手感'],            // 反手拧拉
    13: ['速度','力量','落点','时机'],            // 反手弹击
    14: ['力量','速度','落点','时机'],            // 正手扣杀
    // 基本功 — defense/control
    2:  ['控制','落点','稳定性','节奏'],          // 反手拨球
    6:  ['旋转','控制','弧线','落点'],            // 正手搓球
    7:  ['旋转','控制','弧线','落点'],            // 反手搓球
    8:  ['落点','控制','手感','时机'],            // 摆短
    9:  ['落点','速度','旋转','弧线'],            // 劈长
    12: ['反应','控制','落点','节奏'],            // 正手快带
    // 基本功 — serve/receive/footwork
    15: ['弧线','落点','旋转','速度','隐蔽性'],  // 正手发球
    16: ['弧线','落点','旋转','速度','隐蔽性'],  // 反手发球
    17: ['弧线','旋转','落点','隐蔽性'],          // 逆旋转发球
    18: ['判断','反应','落点','旋转'],            // 接发球
    19: ['速度','灵活','到位','体能'],            // 步法
    // 技战术
    20: ['执行度','落点','时机','速度'],          // 发球抢攻
    21: ['判断','反应','旋转','落点'],            // 接发球抢攻
    22: ['稳定性','旋转','速度','落点'],          // 反手相持对拉
    23: ['时机','落点','速度','衔接'],            // 反手相持转正手
    24: ['速度','力量','落点','时机'],            // 侧身抢攻
    25: ['控制','时机','落点','旋转'],            // 摆短控制+抢先上手
    26: ['落点','速度','旋转','时机'],            // 劈长压反手底线
    27: ['落点','变化','节奏','时机'],            // 落点变化
    28: ['旋转','节奏','变化','控制'],            // 节奏旋转变化
    29: ['反应','稳定性','落点','衔接'],          // 相持转换与反拉防守
  }
  const names = sets[id]
  if (!names) { const obj: Record<string,number> = {}; obj['综合']=1; return obj }
  const obj: Record<string,number> = {}; names.forEach(n => obj[n]=1); return obj
}

const isLoggedIn = computed(() => myId.value > 0)
const canRecord = computed(() => isLoggedIn.value && skillStatus.value !== 'none')
async function submit() {
  if (!isLoggedIn.value) { showToast('请先登录'); return }
  if (!canRecord.value) { showToast('该技能尚未练习，无法记录'); return }
  if (!formDate.value || !formDuration.value) { showToast('请填写日期和时长'); return }
  submitting.value = true
  try {
    const r = await fetch('/api/skill-train', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({ skill_id:skillId, date:formDate.value, duration_minutes:parseInt(formDuration.value)||60,
        location:formLocation.value, partner:formPartner.value, energy_rating:formEnergyRating.value,
        feel_rating:formFeelRating.value, notes:formNotes.value, practice_amount:formPracticeAmount.value,
        skill_notes:formSkillNotes.value, indicators:indicators.value }),
    })
    if (!r.ok) { showToast(await r.text() || '保存失败'); return }
    showSuccessToast('已保存'); router.back()
  } catch (e: any) { showToast('保存失败') }
  finally { submitting.value = false }
}

function statusColor(s: string) { return s==='mastered'?'#07c160':s==='practicing'?'#f5a623':'#aaa' }
</script>

<template>
  <div style="min-height:100vh;background:var(--c-bg);padding-bottom:80px;">

    <!-- Header -->
    <div style="background:linear-gradient(135deg,#1a56db,#1e88e5);color:#fff;padding:20px;display:flex;align-items:center;gap:12px;">
      <IconArrowLeft :size="24" :stroke-width="2" @click="router.back()" style="cursor:pointer;flex-shrink:0;" />
      <div style="flex:1;min-width:0;">
        <div style="font-size:20px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ skillName }}</div>
        <div style="font-size:13px;opacity:0.85;margin-top:2px;">{{ skillCategory }} · {{ skillStatus === 'mastered' ? '已练成' : skillStatus === 'practicing' ? '训练中' : '未开始' }}</div>
      </div>
    </div>

    <van-loading v-if="loading" class="empty-state" />
    <template v-else>

      <!-- Radar Chart Card -->
      <div class="card" style="margin:16px;padding:12px 8px 8px;display:flex;justify-content:center;background:#fff;">
        <svg ref="svgRef" :viewBox="`0 0 ${radarW} ${radarH}`" width="300" height="300"
          style="max-width:100%;touch-action:none;cursor:pointer;"
          @pointerdown="onDown" @pointermove="onMove" @pointerup="onUp" @pointerleave="onUp">
          <!-- Grid -->
          <polygon v-for="lv in [1,2,3,4,5]" :key="lv" :points="gridPoints(lv)"
            fill="none" :stroke="lv===5?'#d0d0d0':'#e8e8e8'" stroke-width="1"
            :stroke-dasharray="lv===5?'0':'4,3'" />
          <!-- Axes -->
          <line v-for="(_,i) in indicatorNames" :key="'ax'+i"
            :x1="radarCX" :y1="radarCY" :x2="axisEnd(i).x" :y2="axisEnd(i).y"
            stroke="#e0e0e0" stroke-width="1" />
          <!-- Data area -->
          <polygon :points="polygonPoints()" :fill="canRecord ? 'rgba(25,137,250,0.15)' : 'rgba(170,170,170,0.1)'"
            :stroke="canRecord ? '#1989fa' : '#aaa'" stroke-width="2" />
          <!-- Data points -->
          <circle v-for="(name,i) in indicatorNames" :key="'pt'+i"
            :cx="radarPoint(i, indicators[name]).x" :cy="radarPoint(i, indicators[name]).y"
            :r="activeAxis===i?10:6" fill="#fff" :stroke="canRecord?'#1989fa':'#aaa'" stroke-width="2.5"
            :style="canRecord?{cursor:'grab',transition:'r .15s'}:{}" />
          <!-- Value labels -->
          <text v-for="(name,i) in indicatorNames" :key="'v'+i"
            :x="radarPoint(i, indicators[name]).x" :y="radarPoint(i, indicators[name]).y - 12"
            text-anchor="middle" font-size="11" font-weight="700" :fill="canRecord?'#1989fa':'#999'">{{ indicators[name] }}</text>
          <!-- Axis labels -->
          <text v-for="(name,i) in indicatorNames" :key="'l'+i"
            :x="labelPos(i).x" :y="labelPos(i).y" :text-anchor="labelPos(i).anchor"
            font-size="12" font-weight="600" fill="#333">{{ name }}</text>
        </svg>
      </div>

      <!-- Not-logged-in notice -->
      <div v-if="!isLoggedIn" style="margin:0 16px 16px;padding:20px 16px;background:linear-gradient(135deg,#e8f4ff,#dbeafe);border-radius:14px;text-align:center;">
        <div style="font-size:40px;margin-bottom:8px;">🏓</div>
        <div style="font-size:16px;font-weight:700;color:#1a56db;margin-bottom:6px;">登录后开始训练</div>
        <div style="font-size:13px;color:#666;margin-bottom:16px;line-height:1.6;">登录即可记录训练数据、追踪技能成长</div>
        <button @click="showLogin=true" style="padding:12px 32px;background:linear-gradient(135deg,#1989fa,#1e88e5);color:#fff;border:none;border-radius:24px;font-size:16px;font-weight:600;cursor:pointer;box-shadow:0 4px 12px rgba(25,137,250,0.3);">📱 短信登录</button>
      </div>

      <!-- Not-practiced notice -->
      <div v-else-if="!canRecord" style="margin:0 16px 16px;padding:16px;background:linear-gradient(135deg,#fef9e7,#fdebd0);border-radius:14px;text-align:center;">
        <div style="font-size:32px;margin-bottom:6px;">🌱</div>
        <div style="font-size:15px;font-weight:700;color:#7d6608;margin-bottom:4px;">尚未开始训练</div>
        <div style="font-size:13px;color:#997404;line-height:1.6;">在「练功记录」中练习这项技能后，这里就能记录详细数据了</div>
      </div>

      <!-- Form -->
      <div style="padding:0 16px;">
        <!-- Date & Duration -->
        <div style="display:flex;gap:12px;margin-bottom:14px;">
          <div style="flex:1;">
            <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">日期</div>
            <input type="date" v-model="formDate" :disabled="!canRecord"
              style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#fff;"
              :style="{ opacity: canRecord?1:0.55 }" />
          </div>
          <div style="flex:1;">
            <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">时长(分钟)</div>
            <input type="number" v-model="formDuration" :disabled="!canRecord"
              style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#fff;"
              :style="{ opacity: canRecord?1:0.55 }" />
          </div>
        </div>

        <!-- Location -->
        <div style="margin-bottom:14px;position:relative;">
          <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">地点</div>
          <input type="text" v-model="formLocation" :disabled="!canRecord" placeholder="搜索或输入地点"
            @focus="showLocDropdown=true" @input="searchLocations(($event.target as any).value)"
            style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#fff;"
            :style="{ opacity: canRecord?1:0.55 }" />
          <div v-if="showLocDropdown && canRecord" style="position:absolute;top:100%;left:0;right:0;background:#fff;border:1px solid #e8e8e8;border-radius:10px;max-height:180px;overflow-y:auto;z-index:100;box-shadow:0 4px 16px rgba(0,0,0,0.08);margin-top:2px;">
            <div v-for="l in locResults" :key="l.id" @click="selectLocation(l.name)"
              style="padding:12px 14px;cursor:pointer;font-size:14px;border-bottom:1px solid #f5f5f5;">{{ l.name }}</div>
            <div v-if="formLocation && locResults.length===0" @click="createLocation()"
              style="padding:12px 14px;cursor:pointer;color:#1989fa;font-size:14px;display:flex;align-items:center;gap:8px;">
              <IconPlus :size="16" /> 创建「{{ formLocation }}」
            </div>
          </div>
        </div>

        <!-- Partner -->
        <div style="margin-bottom:14px;position:relative;">
          <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">陪练</div>
          <input type="text" v-model="formPartner" :disabled="!canRecord" placeholder="搜索球员"
            @focus="showPlayerDropdown=true" @input="searchPlayers(($event.target as any).value)"
            style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#fff;"
            :style="{ opacity: canRecord?1:0.55 }" />
          <div v-if="showPlayerDropdown && canRecord && playerResults.length>0" style="position:absolute;top:100%;left:0;right:0;background:#fff;border:1px solid #e8e8e8;border-radius:10px;max-height:180px;overflow-y:auto;z-index:100;box-shadow:0 4px 16px rgba(0,0,0,0.08);margin-top:2px;">
            <div v-for="p in playerResults" :key="p.id" @click="selectPlayer(p.name)"
              style="padding:12px 14px;cursor:pointer;font-size:14px;border-bottom:1px solid #f5f5f5;">{{ p.name }}</div>
          </div>
        </div>

        <!-- Practice amount -->
        <div style="margin-bottom:14px;">
          <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">练习量</div>
          <input type="text" v-model="formPracticeAmount" :disabled="!canRecord" placeholder="例：5组、200球"
            style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#fff;"
            :style="{ opacity: canRecord?1:0.55 }" />
        </div>

        <!-- Notes -->
        <div style="margin-bottom:14px;">
          <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">备注</div>
          <textarea v-model="formNotes" :disabled="!canRecord" placeholder="训练心得..." rows="2"
            style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:14px;outline:none;resize:vertical;box-sizing:border-box;font-family:inherit;background:#fff;"
            :style="{ opacity: canRecord?1:0.55 }"></textarea>
        </div>
      </div>

      <!-- History -->
      <div v-if="history.length > 0" style="padding:0 16px;margin-top:4px;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px;">
          <div style="font-size:15px;font-weight:600;color:#333;">历史记录 · {{ history.length }}次</div>
          <div v-if="history.length > 1" @click="showAllHistory=!showAllHistory"
            style="color:#1989fa;font-size:13px;cursor:pointer;display:flex;align-items:center;gap:4px;font-weight:500;">
            {{ showAllHistory?'收起':'展开全部' }}
            <IconChevronUp v-if="showAllHistory" :size="16" /><IconChevronDown v-else :size="16" />
          </div>
        </div>
        <div v-for="h in visibleHistory" :key="h.id"
          style="background:#fff;border-radius:12px;padding:14px;margin-bottom:8px;box-shadow:0 1px 4px rgba(0,0,0,0.04);border:1px solid #f0f0f0;">
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px;">
            <span style="font-weight:600;font-size:14px;color:#333;">{{ h.date }}</span>
            <span style="font-size:13px;color:#999;">{{ h.duration_minutes }}分钟</span>
          </div>
          <div v-if="h.partner || h.location" style="font-size:12px;color:#999;margin-bottom:6px;">
            <template v-if="h.partner">陪练: {{ h.partner }}</template>
            <template v-if="h.partner && h.location"> · </template>
            <template v-if="h.location">{{ h.location }}</template>
          </div>
          <div v-if="h.indicators && Object.keys(h.indicators).length>0" style="display:flex;flex-wrap:wrap;gap:6px;margin-top:8px;">
            <span v-for="(v,k) in h.indicators" :key="k"
              style="font-size:12px;padding:3px 10px;background:#e8f4ff;border-radius:6px;color:#1989fa;font-weight:500;">{{ k }} {{ v }}</span>
          </div>
          <div v-if="h.skill_notes" style="font-size:13px;color:#666;margin-top:6px;line-height:1.5;">{{ h.skill_notes }}</div>
        </div>
      </div>

      <div v-else style="text-align:center;padding:32px;color:#bbb;font-size:14px;">暂无训练记录</div>

    </template>

    <!-- Save button -->
    <div v-if="canRecord" style="position:fixed;bottom:0;left:0;right:0;padding:12px 16px max(12px, env(safe-area-inset-bottom));background:rgba(255,255,255,0.97);border-top:1px solid #eee;z-index:100;">
      <button @click="submit" :disabled="submitting"
        style="width:100%;padding:16px;background:linear-gradient(135deg,#1989fa,#1e88e5);color:#fff;border:none;border-radius:14px;font-size:17px;font-weight:700;cursor:pointer;box-shadow:0 4px 16px rgba(25,137,250,0.25);transition:transform .1s;"
        :style="{ opacity: submitting?0.7:1 }">
        {{ submitting?'保存中...':'💾 保存记录' }}
      </button>
    </div>

    <!-- Login modal -->
    <div v-if="showLogin" style="position:fixed;inset:0;background:rgba(0,0,0,0.4);z-index:3000;display:flex;align-items:center;justify-content:center;" @click.self="showLogin=false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:300px;max-width:90vw;">
        <h3 style="text-align:center;margin-bottom:16px;">📱 短信验证登录</h3>
        <div v-if="loginMsg" style="text-align:center;font-size:12px;margin-bottom:8px;" :style="{color:loginMsg.includes('发送')?'#1989fa':'#e74c3c'}">{{ loginMsg }}</div>
        <div v-if="!loginCode">
          <input v-model="loginPhone" placeholder="输入手机号" type="tel" maxlength="11" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:16px;outline:none;box-sizing:border-box;margin-bottom:12px;" />
          <button @click="sendCode" :disabled="loginSending" style="width:100%;padding:14px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">{{ loginSending?'发送中...':'获取验证码' }}</button>
        </div>
        <div v-else>
          <div style="font-size:13px;color:#666;text-align:center;margin-bottom:8px;">已发送至 {{ loginPhone }}</div>
          <input v-model="loginCode" placeholder="输入验证码" type="tel" maxlength="4" style="width:100%;padding:12px;border:1px solid #ddd;border-radius:10px;font-size:20px;font-weight:700;text-align:center;outline:none;box-sizing:border-box;margin-bottom:12px;letter-spacing:8px;" />
          <button @click="verifyCode" style="width:100%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;">验证登录</button>
        </div>
        <button @click="showLogin=false" style="width:100%;padding:10px;margin-top:8px;background:none;border:none;color:#999;font-size:13px;cursor:pointer;">取消</button>
      </div>
    </div>
  </div>
</template>
