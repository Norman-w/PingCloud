<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showSuccessToast, showLoadingToast } from 'vant'
import { IconArrowLeft, IconPlayerPlay, IconPlayerStop, IconClock, IconChevronDown, IconChevronUp } from '@tabler/icons-vue'
import { myId, myName, checkAuth, logout as authLogout } from '../auth'
import LoginModal from '../components/LoginModal.vue'
import LocationPicker from '../components/LocationPicker.vue'
import PlayerPicker from '../components/PlayerPicker.vue'
import StartTrainingDialog from '../components/StartTrainingDialog.vue'

const route = useRoute(); const router = useRouter()
const skillId = Number(route.params.id)

// ── auth ──
const showLogin = ref(false)
function logout() { authLogout(); loadData() }

// ── skill info ──
const skillName = ref(''); const skillCategory = ref(''); const skillStatus = ref('none'); const skillTags = ref<string[]>([])

// ── indicators ──
interface HistEntry { id: number; date: string; duration_minutes: number; location: string; partner: string; notes: string; practice_amount: string; skill_notes: string; indicators: Record<string,number>; created_at: string }
const history = ref<HistEntry[]>([])
const indicatorMode = ref<'current'|'max'|'avg'>('current')
const defaultKeys = computed(() => Object.keys(defaults(skillId)))

const currentIndicators = computed(() => {
  const base = {...defaults(skillId)}
  if (history.value.length > 0 && history.value[0].indicators) {
    for (const [k,v] of Object.entries(history.value[0].indicators)) { if (k in base) base[k]=v }
  }
  return base
})
const maxIndicators = computed(() => {
  const base = {...defaults(skillId)}
  for (const h of history.value) {
    for (const [k,v] of Object.entries(h.indicators||{})) {
      if (k in base) base[k] = Math.max(base[k], v)
    }
  }
  return base
})
const avgIndicators = computed(() => {
  const base = {...defaults(skillId)}
  const sum: Record<string,number> = {}; const cnt: Record<string,number> = {}
  for (const k of defaultKeys.value) { sum[k]=0; cnt[k]=0 }
  for (const h of history.value) {
    for (const [k,v] of Object.entries(h.indicators||{})) {
      if (k in sum) { sum[k]+=v; cnt[k]++ }
    }
  }
  for (const k of defaultKeys.value) {
    if (cnt[k] > 0) base[k] = Math.round(sum[k]/cnt[k])
  }
  return base
})
const displayIndicators = computed(() => {
  if (indicatorMode.value === 'max') return maxIndicators.value
  if (indicatorMode.value === 'avg') return avgIndicators.value
  return currentIndicators.value
})

// ── history display ──
const showAllHistory = ref(false); const expandedHistory = ref<Set<number>>(new Set())
const visibleHistory = computed(() => showAllHistory.value ? history.value : history.value.slice(0, 1))
function toggleExpand(id: number) { const s = new Set(expandedHistory.value); s.has(id) ? s.delete(id) : s.add(id); expandedHistory.value = s }

// ── training mode ──
const training = ref(false)
const trainStart = ref(0); const trainElapsed = ref(0); let trainTimer: any = null
const targetSeconds = ref(0)
const showStartDialog = ref(false)
const startDialogRef = ref<InstanceType<typeof StartTrainingDialog>>()

function beep() {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    const osc = ctx.createOscillator(); const gain = ctx.createGain()
    osc.connect(gain); gain.connect(ctx.destination)
    osc.frequency.value = 880; osc.type = 'square'; gain.gain.value = 0.3
    osc.start(); osc.stop(ctx.currentTime + 0.15)
    setTimeout(() => { const osc2 = ctx.createOscillator(); osc2.connect(gain); osc2.frequency.value = 1200; osc2.type = 'square'; osc2.start(); osc2.stop(ctx.currentTime + 0.2) }, 200)
  } catch {}
}

function openStartDialog() {
  if (!myId.value) { showLogin.value = true; return }
  startDialogRef.value?.loadGoals?.(skillId)
  showStartDialog.value = true
}

function startTraining(durationMin: number) {
  targetSeconds.value = durationMin * 60
  training.value = true; trainStart.value = Date.now(); trainElapsed.value = 0
  trainTimer = setInterval(() => {
    trainElapsed.value = Math.floor((Date.now() - trainStart.value) / 1000)
    if (targetSeconds.value > 0 && trainElapsed.value >= targetSeconds.value) {
      clearInterval(trainTimer); trainTimer = null
      beep()
    }
  }, 1000)
}
function stopTraining() { training.value = false; if (trainTimer) { clearInterval(trainTimer); trainTimer = null }; openConfirm() }

// ── confirm/save ──
const showConfirm = ref(false)
const confirmDate = ref(''); const confirmDuration = ref(''); const confirmLoc = ref(''); const confirmPartner = ref(''); const confirmNotes = ref(''); const confirmAmount = ref('')
const confirmIndicators = ref<Record<string,number>>({})
const goalValues = ref<Record<string,number>>({})
const saving = ref(false)
const hasData = computed(() => history.value.length > 0)
const showLocPicker = ref(false); const showPlayerPicker = ref(false)
// Editable radar
const confirmKeys = computed(() => Object.keys(confirmIndicators.value))
const editW = 260; const editH = 280; const editCX = 130; const editCY = 125; const editR = 90
const svgRef = ref<SVGSVGElement>(); const activeAxis = ref(-1); const dragging = ref(false)
function editPt(i:number,v:number){const n=confirmKeys.value.length;if(!n)return{x:editCX,y:editCY};const a=(2*Math.PI*i)/n-Math.PI/2;return{x:editCX+(editR*v/5)*Math.cos(a),y:editCY+(editR*v/5)*Math.sin(a)}}
function editEnd(i:number){const n=confirmKeys.value.length;if(!n)return{x:editCX,y:editCY};const a=(2*Math.PI*i)/n-Math.PI/2;return{x:editCX+editR*Math.cos(a),y:editCY+editR*Math.sin(a)}}
function editLbl(i:number){const e=editEnd(i);const dx=e.x-editCX;const dy=e.y-editCY;const d=Math.hypot(dx,dy);const nx=dx/d;const ny=dy/d;return{x:editCX+nx*(editR+20),y:editCY+ny*(editR+16),anchor:Math.abs(nx)>0.7?(nx>0?'start':'end'):'middle'}}
function editGrid(lv:number){return confirmKeys.value.map((_,i)=>{const p=editPt(i,lv);return`${p.x},${p.y}`}).join(' ')}
function editPoly(){return confirmKeys.value.map((_,i)=>{const p=editPt(i,confirmIndicators.value[confirmKeys.value[i]]||1);return`${p.x},${p.y}`}).join(' ')}
function getPos(e:PointerEvent){const s=svgRef.value!;const r=s.getBoundingClientRect();return{x:(e.clientX-r.left)/r.width*editW,y:(e.clientY-r.top)/r.height*editH}}
function onDown(e:PointerEvent){const p=getPos(e);let best=-1;let minD=Infinity;for(let i=0;i<confirmKeys.value.length;i++){const ep=editEnd(i);const d=Math.hypot(p.x-ep.x,p.y-ep.y);if(d<minD){minD=d;best=i}};if(minD<45){activeAxis.value=best;dragging.value=true;svgRef.value?.setPointerCapture(e.pointerId)}}
function onMove(e:PointerEvent){if(!dragging.value||activeAxis.value<0)return;const p=getPos(e);const i=activeAxis.value;const n=confirmKeys.value.length;const a=(2*Math.PI*i)/n-Math.PI/2;const proj=(p.x-editCX)*Math.cos(a)+(p.y-editCY)*Math.sin(a);const v=Math.round(Math.max(1,Math.min(5,proj/editR*5)));confirmIndicators.value={...confirmIndicators.value,[confirmKeys.value[i]]:v}}
function onUp(){dragging.value=false;activeAxis.value=-1}

function openConfirm() {
  confirmDate.value = new Date().toISOString().slice(0,10)
  confirmDuration.value = String(trainElapsed.value || 60)
  confirmLoc.value = ''; confirmPartner.value = ''; confirmNotes.value = ''; confirmAmount.value = ''
  goalValues.value = {}
  const base = {...defaults(skillId)}
  if (hasData.value && history.value[0]?.indicators) {
    for (const [k,v] of Object.entries(history.value[0].indicators)) { if (k in base) base[k]=v }
  }
  confirmIndicators.value = JSON.parse(JSON.stringify(base))
  showConfirm.value = true
}

async function saveRecord() {
  saving.value = true
  const loadingToast = showLoadingToast({ message: '提交中...', forbidClick: true, duration: 0 })
  try {
    const body = { skill_id: skillId, date: confirmDate.value, duration_minutes: parseInt(confirmDuration.value)||60,
      location: confirmLoc.value, partner: confirmPartner.value, notes: confirmNotes.value,
      practice_amount: confirmAmount.value, skill_notes: '', energy_rating: 0, feel_rating: 0,
      indicators: confirmIndicators.value, goal_values: goalValues.value }
    const r = await fetch('/api/skill-train', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(body) })
    loadingToast.close()
    if (!r.ok) { showToast(await r.text() || '保存失败'); return }
    showSuccessToast('保存成功 ✅'); showConfirm.value = false; router.back()
  } catch (e: any) { loadingToast.close(); showToast('保存失败') }
  finally { saving.value = false }
}

// ── radar chart ──
const radarW = 320; const radarH = 320; const radarCX = 160; const radarCY = 150; const radarR = 85
const indicatorNames = computed(() => Object.keys(displayIndicators.value))
function radarPoint(i: number, val: number) {
  const n = indicatorNames.value.length; if (n===0) return {x:radarCX,y:radarCY}
  const a = (2*Math.PI*i)/n - Math.PI/2
  return {x:radarCX+(radarR*val/5)*Math.cos(a), y:radarCY+(radarR*val/5)*Math.sin(a)}
}
function axisEnd(i: number) {
  const n = indicatorNames.value.length; if (n===0) return {x:radarCX,y:radarCY}
  const a = (2*Math.PI*i)/n - Math.PI/2
  return {x:radarCX+radarR*Math.cos(a), y:radarCY+radarR*Math.sin(a)}
}
function labelPos(i: number) {
  const e = axisEnd(i); const dx=e.x-radarCX; const dy=e.y-radarCY; const d=Math.hypot(dx,dy); const nx=dx/d; const ny=dy/d
  return {x:radarCX+nx*(radarR+32), y:radarCY+ny*(radarR+26), anchor:Math.abs(nx)>0.4?(nx>0?'start':'end'):'middle'}
}
function gridPoints(lv: number) { return indicatorNames.value.map((_,i)=>{const p=radarPoint(i,lv);return`${p.x},${p.y}`}).join(' ') }
function polyPoints(vals: Record<string,number>) { return indicatorNames.value.map((_,i)=>{const p=radarPoint(i,vals[indicatorNames.value[i]]||1);return`${p.x},${p.y}`}).join(' ') }

// ── load ──
// ── goals ──
interface GoalProgress { id:number; label:string; unit:string; tier_1:number; tier_2:number; tier_3:number; tier_4:number; tier_5:number; min_stars:number; current_value:number; stars:number; passed:boolean }
const skillGoals = ref<GoalProgress[]>([])
const showGoals = ref(false)

const loading = ref(true)
onMounted(async () => { await checkAuth(); await loadData() })
async function loadData() {
  loading.value = true
  try {
    const r = await fetch(`/api/skill-train/${skillId}`)
    if (r.ok) { const d = await r.json(); skillName.value = d.skill_name; skillCategory.value = d.category; history.value = d.history || [] }
  } catch {}
  try {
    const r = await fetch('/api/skill-mastery')
    if (r.ok) { const d = await r.json(); const item = (d.skills||[]).find((s:any)=>s.id===skillId); if (item) { skillStatus.value = item.status; skillTags.value = item.tags||[] } }
  } catch {}
  try { const r = await fetch('/api/skill-goals/'+skillId); if (r.ok) skillGoals.value = await r.json() } catch {}
  loading.value = false
}
onUnmounted(() => { if (trainTimer) clearInterval(trainTimer) })

function defaults(id: number): Record<string,number> {
  const s: Record<number,string[]> = {}
  const serve = ['弧线','落点','旋转','速度','隐蔽性']; [15,16,17].forEach(k=>s[k]=serve)
  const loopBase = ['跟球不漏','脚踩路线','击球时机','重心下压','转体','收小臂','手指位置','板型固定','球受力点','旋转']
s[3]=loopBase; s[4]=loopBase; s[5]=loopBase;
  const chop = ['旋转','控制','弧线','落点','手感']; [6,7,8,9].forEach(k=>s[k]=chop)
  s[1]=['跟球不漏','击球时机','重心下压','转体','收小臂','手指位置','板型固定','球受力点']; s[10]=['速度','落点','时机','手感']; s[11]=['旋转','弧线','落点','手感']
  s[12]=['反应','控制','落点','节奏']; s[13]=['速度','力量','落点','时机']; s[14]=['力量','速度','落点','时机']
  s[18]=['判断','反应','落点','旋转']; s[19]=['速度','灵活','到位','体能']
  s[20]=['执行度','落点','时机','速度']; s[21]=['判断','反应','旋转','落点']; s[22]=['稳定性','旋转','速度','落点']
  s[23]=['时机','落点','速度','衔接']; s[24]=['速度','力量','落点','时机']; s[25]=['控制','时机','落点','旋转']
  s[26]=['落点','速度','旋转','时机']; s[27]=['落点','变化','节奏','时机']; s[28]=['旋转','节奏','变化','控制']; s[29]=['反应','稳定性','落点','衔接']
  for (const k of [30,31,32,33,34]) s[k]=['理解度','兴趣','掌握']
  for (const k of [35,36,37,38,39]) s[k]=['理解','应用','感受']
  const names = s[id]; if (!names) { const o:Record<string,number>={}; o['综合']=1; return o }
  const o:Record<string,number>={}; names.forEach(n=>o[n]=1); return o
}

function formatTime(s: number) { const h=Math.floor(s/3600); const m=Math.floor((s%3600)/60); const sec=s%60; if(h>0)return`${h}时${m}分${sec}秒`; if(m>0)return`${m}分${sec}秒`; return`${sec}秒` }
</script>

<template>
  <div style="min-height:100vh;background:var(--c-bg);padding-bottom:100px;">

    <!-- Header -->
    <div style="background:linear-gradient(135deg,#1a56db,#1e88e5);color:#fff;padding:16px 20px;display:flex;align-items:center;gap:12px;">
      <IconArrowLeft :size="24" :stroke-width="2" @click="router.back()" style="cursor:pointer;flex-shrink:0;" />
      <div style="flex:1;min-width:0;">
        <div style="font-size:20px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ skillName }}</div>
        <div style="font-size:12px;opacity:0.85;">{{ skillCategory }}</div>
      </div>
    </div>

    <van-loading v-if="loading" class="empty-state" />
    <template v-else>

      <!-- Radar Chart -->
      <div class="card" style="margin:16px;padding:8px;display:flex;flex-direction:column;align-items:center;">
        <!-- Toggle -->
        <div style="display:flex;gap:4px;margin-bottom:4px;background:#f0f2f5;border-radius:10px;padding:3px;">
          <span v-for="m in [{k:'current',l:'当前'},{k:'max',l:'最高'},{k:'avg',l:'平均'}]" :key="m.k"
            @click="indicatorMode = m.k as any"
            style="padding:6px 16px;border-radius:8px;font-size:13px;font-weight:600;cursor:pointer;transition:all .15s;"
            :style="indicatorMode===m.k?'background:#1989fa;color:#fff;':'color:#666;'">{{ m.l }}</span>
        </div>
        <!-- No data -->
        <div v-if="!hasData" style="text-align:center;padding:40px 20px;color:#bbb;">
          <div style="font-size:48px;margin-bottom:8px;">📊</div>
          <div style="font-size:15px;font-weight:600;color:#999;">暂无训练数据</div>
          <div style="font-size:13px;color:#ccc;margin-top:4px;">开始训练后，指标数据将在此展示</div>
        </div>
        <!-- Radar with data -->
        <svg v-else :viewBox="`0 0 ${radarW} ${radarH}`" width="300" height="300" style="max-width:100%;">
          <polygon v-for="lv in [1,2,3,4,5]" :key="lv" :points="gridPoints(lv)" fill="none" :stroke="lv===5?'#d0d0d0':'#e8e8e8'" stroke-width="1" :stroke-dasharray="lv===5?'0':'4,3'" />
          <line v-for="(_,i) in indicatorNames" :key="'ax'+i" :x1="radarCX" :y1="radarCY" :x2="axisEnd(i).x" :y2="axisEnd(i).y" stroke="#e0e0e0" stroke-width="1" />
          <polygon :points="polyPoints(displayIndicators)" fill="rgba(25,137,250,0.12)" stroke="#1989fa" stroke-width="2" />
          <circle v-for="(name,i) in indicatorNames" :key="'pt'+i" :cx="radarPoint(i,displayIndicators[name]).x" :cy="radarPoint(i,displayIndicators[name]).y" r="5" fill="#fff" stroke="#1989fa" stroke-width="2" />
          <text v-for="(name,i) in indicatorNames" :key="'v'+i" :x="radarPoint(i,displayIndicators[name]).x" :y="radarPoint(i,displayIndicators[name]).y-10" text-anchor="middle" font-size="10" font-weight="700" fill="#1989fa">{{ displayIndicators[name].toFixed(2) }}</text>
          <text v-for="(name,i) in indicatorNames" :key="'l'+i" :x="labelPos(i).x" :y="labelPos(i).y" :text-anchor="labelPos(i).anchor" font-size="12" font-weight="600" fill="#333">{{ name }}</text>
        </svg>
      </div>

      <!-- Goals / 评定标准 (always visible) -->
      <div v-if="skillGoals.length>0" style="padding:0 16px;margin-bottom:12px;">
        <div style="background:#fff;border-radius:12px;padding:12px 16px;box-shadow:0 1px 4px rgba(0,0,0,0.04);">
          <div style="font-weight:600;font-size:14px;color:#333;margin-bottom:10px;">📋 评定标准 · {{ skillGoals.filter(g=>g.passed).length }}/{{ skillGoals.length }} 达标</div>
          <div v-for="g in skillGoals" :key="g.id" style="padding:8px 0;border-bottom:1px solid #f5f5f5;">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;">
              <span style="font-weight:600;font-size:13px;">{{ g.label }} <span style="font-weight:400;color:#999;">({{ g.unit }})</span></span>
              <span style="font-size:11px;" :style="{color:g.passed?'#07c160':'#ff976a'}">{{ g.passed?'✓ 已达标':'需'+g.min_stars+'★'}}</span>
            </div>
            <div style="font-size:11px;color:#999;margin-bottom:4px;">
              当前 {{ g.current_value }}{{ g.unit }} ·
              <span v-for="t in [g.tier_1,g.tier_2,g.tier_3,g.tier_4,g.tier_5]" :key="t" style="margin:0 2px;"
                :style="{color:g.current_value>=t?'#f5a623':'#ccc',fontWeight:g.current_value>=t?600:400}">
                {{ t>0?'★'+t:'—' }}
              </span>
            </div>
            <div style="height:3px;background:#e8e8e8;border-radius:2px;overflow:hidden;">
              <div style="height:100%;border-radius:2px;transition:width .3s;"
                :style="{width:Math.min(100,g.current_value/Math.max(1,g.tier_3)*100)+'%',background:g.passed?'#07c160':'#f5a623'}"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Training button (if logged in) -->
      <div v-if="myId" style="padding:0 16px;margin-bottom:16px;">
        <div v-if="!training" @click="openStartDialog"
          style="background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border-radius:16px;padding:28px;text-align:center;cursor:pointer;box-shadow:0 6px 20px rgba(7,193,96,0.3);transition:transform .1s;">
          <IconPlayerPlay :size="40" :stroke-width="2" style="display:block;margin:0 auto 8px;" />
          <div style="font-size:22px;font-weight:800;">开练</div>
          <div style="font-size:13px;opacity:0.85;margin-top:4px;">点击开始记录本次训练</div>
        </div>
        <div v-else
          style="background:linear-gradient(135deg,#ee0a24,#ff4757);color:#fff;border-radius:16px;padding:28px;text-align:center;box-shadow:0 6px 20px rgba(238,10,36,0.3);">
          <IconClock :size="40" :stroke-width="2" style="display:block;margin:0 auto 8px;" />
          <div style="font-size:36px;font-weight:800;font-variant-numeric:tabular-nums;">{{ formatTime(trainElapsed) }}</div>
          <div style="font-size:13px;opacity:0.85;margin:4px 0 4px;">训练进行中...</div>
          <div v-if="targetSeconds>0" style="font-size:13px;opacity:0.7;margin-bottom:16px;">⏱ 目标 {{ formatTime(targetSeconds) }} · {{ trainElapsed>=targetSeconds?'✅ 已完成！':'剩余 '+formatTime(targetSeconds-trainElapsed) }}</div>
          <div v-else style="margin-bottom:16px;"></div>
          <button @click="stopTraining"
            style="background:#fff;color:#ee0a24;border:none;padding:14px 40px;border-radius:24px;font-size:17px;font-weight:700;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:8px;margin:0 auto;">
            <IconPlayerStop :size="20" /> 结束训练
          </button>
        </div>
      </div>
      <div v-else style="padding:0 16px;margin-bottom:16px;">
        <div @click="showLogin=true" style="background:linear-gradient(135deg,#f0f2f5,#e5e7eb);color:#999;border-radius:16px;padding:24px;text-align:center;cursor:pointer;border:2px dashed #ddd;">
          <div style="font-size:18px;font-weight:700;margin-bottom:4px;">🔒 登录后开始训练</div>
          <div style="font-size:13px;">点击登录，记录每一次进步</div>
        </div>
      </div>

      <!-- History -->
      <div v-if="history.length > 0" style="padding:0 16px;">
        <div style="font-size:15px;font-weight:600;margin-bottom:10px;display:flex;justify-content:space-between;">
          <span>训练记录 · {{ history.length }}次</span>
          <span v-if="history.length > 1" @click="showAllHistory=!showAllHistory" style="color:#1989fa;font-size:13px;cursor:pointer;display:flex;align-items:center;gap:4px;">
            {{ showAllHistory?'收起':'加载更多' }} <IconChevronDown :size="16" />
          </span>
        </div>
        <div v-for="h in visibleHistory" :key="h.id" style="background:#fff;border-radius:12px;margin-bottom:8px;overflow:hidden;box-shadow:0 1px 4px rgba(0,0,0,0.04);">
          <div @click="toggleExpand(h.id)" style="padding:14px;cursor:pointer;display:flex;justify-content:space-between;align-items:center;">
            <div>
              <div style="font-weight:600;font-size:14px;">{{ new Date(h.created_at).toLocaleDateString('zh-CN',{month:'long',day:'numeric',weekday:'short'}) }}</div>
              <div style="font-size:11px;color:#bbb;">{{ new Date(h.created_at).toLocaleTimeString('zh-CN',{hour:'2-digit',minute:'2-digit'}) }}</div>
              <div style="font-size:12px;color:#999;margin-top:2px;">{{ formatTime(h.duration_minutes) }}<template v-if="h.partner"> · {{ h.partner }}</template><template v-if="h.location"> · {{ h.location }}</template></div>
            </div>
            <IconChevronDown :size="18" style="color:#ccc;" :style="{transform:expandedHistory.has(h.id)?'rotate(180deg)':'rotate(0deg)',transition:'transform .2s'}" />
          </div>
          <div v-if="expandedHistory.has(h.id)" style="padding:0 14px 14px;border-top:1px solid #f5f5f5;">
            <div v-if="h.indicators && Object.keys(h.indicators).length>0" style="display:flex;flex-wrap:wrap;gap:6px;margin:10px 0;">
              <span v-for="(v,k) in h.indicators" :key="k" style="font-size:12px;padding:3px 10px;background:#e8f4ff;border-radius:6px;color:#1989fa;font-weight:500;">{{ k }} {{ v }}</span>
            </div>
            <div v-if="h.skill_notes" style="font-size:13px;color:#666;line-height:1.5;">{{ h.skill_notes }}</div>
            <div v-if="h.notes" style="font-size:13px;color:#666;line-height:1.5;margin-top:4px;">{{ h.notes }}</div>
            <div v-if="h.practice_amount" style="font-size:12px;color:#999;margin-top:6px;">练习量: {{ h.practice_amount }}</div>
          </div>
        </div>
      </div>
      <div v-else style="text-align:center;padding:20px;color:#bbb;font-size:14px;">暂无训练记录</div>

    </template>

    <!-- Confirm modal -->
    <div v-if="showConfirm" style="position:fixed;inset:0;background:rgba(0,0,0,0.5);z-index:3000;overflow-y:auto;" @click.self="showConfirm=false">
      <div style="background:#fff;min-height:100vh;padding:0 0 80px;">
        <div style="background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;padding:16px 20px;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:10;">
          <button @click="showConfirm=false" style="background:none;border:none;color:#fff;font-size:16px;cursor:pointer;">取消</button>
          <span style="font-weight:700;font-size:18px;">训练完成 🎉</span>
          <button @click="saveRecord" :disabled="saving" style="background:rgba(255,255,255,0.25);border:none;color:#fff;font-size:15px;font-weight:600;padding:8px 20px;border-radius:16px;cursor:pointer;">{{ saving?'保存中...':'保存' }}</button>
        </div>
        <div style="padding:16px;">
          <!-- Summary -->
          <div style="text-align:center;padding:20px;background:#f8f9fa;border-radius:12px;margin-bottom:8px;">
            <div style="font-size:32px;font-weight:800;color:#07c160;">{{ formatTime(trainElapsed) }}</div>
            <div style="font-size:14px;color:#666;margin-top:4px;">本次训练时长</div>
          </div>

          <!-- Editable radar chart -->
          <div style="margin-bottom:16px;display:flex;flex-direction:column;align-items:center;">
            <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">📊 拖拽调整指标</div>
            <svg ref="svgRef" :viewBox="`0 0 ${editW} ${editH}`" width="260" height="280" style="max-width:100%;touch-action:none;cursor:pointer;" @pointerdown="onDown" @pointermove="onMove" @pointerup="onUp" @pointerleave="onUp">
              <polygon v-for="lv in [1,2,3,4,5]" :key="lv" :points="editGrid(lv)" fill="none" :stroke="lv===5?'#d0d0d0':'#e8e8e8'" stroke-width="1" :stroke-dasharray="lv===5?'0':'4,3'" />
              <line v-for="(_,i) in confirmKeys" :key="'ax'+i" :x1="editCX" :y1="editCY" :x2="editEnd(i).x" :y2="editEnd(i).y" stroke="#e0e0e0" stroke-width="1" />
              <polygon :points="editPoly()" fill="rgba(7,193,96,0.15)" stroke="#07c160" stroke-width="2" />
              <circle v-for="(name,i) in confirmKeys" :key="'pt'+i" :cx="editPt(i,confirmIndicators[name]).x" :cy="editPt(i,confirmIndicators[name]).y" :r="activeAxis===i?10:6" fill="#fff" stroke="#07c160" stroke-width="2.5" style="cursor:grab;" />
              <text v-for="(name,i) in confirmKeys" :key="'v'+i" :x="editPt(i,confirmIndicators[name]).x" :y="editPt(i,confirmIndicators[name]).y-10" text-anchor="middle" font-size="10" font-weight="700" fill="#07c160">{{ confirmIndicators[name] }}</text>
              <text v-for="(name,i) in confirmKeys" :key="'l'+i" :x="editLbl(i).x" :y="editLbl(i).y" :text-anchor="editLbl(i).anchor" font-size="11" font-weight="600" fill="#333">{{ name }}</text>
            </svg>
          </div>
          <!-- Editable fields -->
          <div style="display:flex;gap:12px;margin-bottom:14px;">
            <div style="flex:1;"><div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">日期</div><input v-model="confirmDate" type="date" style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;" /></div>
            <div style="flex:1;"><div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">时长(秒)</div><input v-model="confirmDuration" type="number" style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;" /></div>
          </div>
          <div style="display:flex;gap:12px;margin-bottom:14px;">
            <div style="flex:1;"><div style="font-size:12px;font-weight:600;color:#888;margin-bottom:4px;">地点</div>
              <button @click="showLocPicker=true" style="width:100%;padding:11px;border:1px solid #e0e0e0;border-radius:10px;font-size:14px;text-align:left;background:#fff;cursor:pointer;outline:none;box-sizing:border-box;" :style="{color:confirmLoc?'#333':'#999'}">{{ confirmLoc || '点击选择场馆' }}</button>
            </div>
            <div style="flex:1;"><div style="font-size:12px;font-weight:600;color:#888;margin-bottom:4px;">陪练</div>
              <button @click="showPlayerPicker=true" style="width:100%;padding:11px;border:1px solid #e0e0e0;border-radius:10px;font-size:14px;text-align:left;background:#fff;cursor:pointer;outline:none;box-sizing:border-box;" :style="{color:confirmPartner?'#333':'#999'}">{{ confirmPartner || '点击选择球员' }}</button>
            </div>
          </div>
          <!-- Goal quantities -->
          <div v-if="skillGoals.length>0" style="margin-bottom:14px;">
            <div style="font-size:13px;font-weight:600;color:#555;margin-bottom:8px;">训练数量</div>
            <div v-for="g in skillGoals" :key="g.id" style="margin-bottom:10px;background:#f8f9fa;border-radius:10px;padding:10px 12px;">
              <div style="font-size:12px;font-weight:600;color:#666;margin-bottom:6px;">{{ g.label }} ({{ g.unit }})</div>
              <div style="display:flex;flex-wrap:wrap;gap:4px;margin-bottom:6px;">
                <span v-for="v in [10,20,50,100,200,500]" :key="v" @click="goalValues[g.label]=(goalValues[g.label]||0)+v"
                  style="padding:4px 10px;border-radius:6px;font-size:12px;font-weight:600;cursor:pointer;background:#e8f4ff;color:#1989fa;">+{{ v }}</span>
                <span @click="goalValues[g.label]=Math.max(0,(goalValues[g.label]||0)-10)"
                  style="padding:4px 10px;border-radius:6px;font-size:12px;cursor:pointer;background:#fde8e8;color:#ee0a24;">-10</span>
                <span @click="goalValues[g.label]=0"
                  style="padding:4px 10px;border-radius:6px;font-size:12px;cursor:pointer;background:#f0f0f0;color:#999;">清零</span>
              </div>
              <input v-model.number="goalValues[g.label]" type="number" :placeholder="'本次'+g.label"
                style="width:100%;padding:8px 10px;border:1px solid #e0e0e0;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;text-align:center;" />
            </div>
          </div>
          <div style="margin-bottom:14px;"><div style="font-size:13px;font-weight:600;color:#555;margin-bottom:6px;">备注</div><textarea v-model="confirmNotes" placeholder="训练心得..." rows="2" style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:14px;outline:none;resize:vertical;box-sizing:border-box;font-family:inherit;"></textarea></div>
        </div>
      </div>
    </div>

    <LoginModal v-model:visible="showLogin" />
  <StartTrainingDialog ref="startDialogRef" v-model:visible="showStartDialog" @start="startTraining" />
  <LocationPicker v-model="confirmLoc" v-model:visible="showLocPicker" />
  <PlayerPicker v-model="confirmPartner" v-model:visible="showPlayerPicker" />
  </div>
</template>
