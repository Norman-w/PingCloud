<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import { IconTournament, IconPlus, IconList, IconUsers, IconTrophy, IconChevronRight } from '@tabler/icons-vue'
import FunMatchPlay, { type FunMatchItem } from '../components/FunMatchPlay.vue'
import FunMatchHonors from '../components/FunMatchHonors.vue'
import { api, type Player } from '../api'

interface FunSessionSummary {
  id: number; name: string; male_count: number; female_count: number
  status: string; male_wins: number; female_wins: number
  male_game_wins: number; female_game_wins: number
  male_points: number; female_points: number
  match_count: number; unplayed_count: number; created_at: string
}

interface FunPlayer {
  id: number; name: string; current_rating: number; reference_rating: number; team: string
}

interface FunSessionDetail {
  id: number; name: string; male_count: number; female_count: number
  status: string; winning_team: string; male_wins: number; female_wins: number
  male_game_wins: number; female_game_wins: number
  male_points: number; female_points: number
  created_at: string; players: FunPlayer[]; matches: FunMatchItem[]
}

const players = ref<Player[]>([])
const loadingPlayers = ref(true)

const step = ref<'list' | 'select' | 'confirm' | 'play' | 'result'>('list')
const matchMode = ref('gender')
const maleIDs = ref<Set<number>>(new Set())
const femaleIDs = ref<Set<number>>(new Set())
const sessionName = ref('')

const isPimpleRR = computed(() => matchMode.value === 'pimple_rr')
const teamLabels = computed(() => {
  if (matchMode.value === 'rubber') return { a:'双反队', b:'颗粒队', ta:'反胶', tb:'颗粒' }
  if (matchMode.value === 'pimple_rr') return { a:'颗粒组', b:'', ta:'颗粒', tb:'' }
  return { a:'男队', b:'女队', ta:'男', tb:'女' }
})

const currentSession = ref<FunSessionDetail | null>(null)
const sessions = ref<FunSessionSummary[]>([])

const scoringMatch = ref<FunMatchItem | null>(null)
const showEditDialog = ref(false)
const loadingSession = ref(false)

onMounted(async () => { await loadAll() })

async function loadAll() {
  loadingPlayers.value = true
  try {
    const [p, s] = await Promise.all([
      api.getPlayers(),
      fetch('/api/fun-sessions').then(r => r.json()).catch(() => []),
    ])
    players.value = p; sessions.value = s
  } catch (e) { /* ignore */ }
  finally { loadingPlayers.value = false }
}

function enterSelect() {
  maleIDs.value = new Set(); femaleIDs.value = new Set()
  sessionName.value = ''; step.value = 'select'
}

function toggleMale(id: number) {
  const s = new Set(maleIDs.value); s.has(id) ? s.delete(id) : s.add(id)
  femaleIDs.value.delete(id)
  maleIDs.value = s
}

function toggleFemale(id: number) {
  const s = new Set(femaleIDs.value); s.has(id) ? s.delete(id) : s.add(id)
  maleIDs.value.delete(id)
  femaleIDs.value = s
}

function goConfirm() {
  if (maleIDs.value.size === 0) { showToast('请至少选择1人'); return }
  if (!isPimpleRR.value && femaleIDs.value.size === 0) { showToast('请至少选择1人'); return }
  if (!sessionName.value.trim()) {
    const mNames = Array.from(maleIDs.value).map(id => players.value.find(p => p.id === id)?.name).filter(Boolean)
    const fNames = Array.from(femaleIDs.value).map(id => players.value.find(p => p.id === id)?.name).filter(Boolean)
    sessionName.value = mNames.join('、') + ' VS ' + fNames.join('、')
  }
  step.value = 'confirm'
}

async function createSession() {
  try {
    const res = await fetch('/api/fun-sessions', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: sessionName.value.trim(),
        mode: matchMode.value,
        male_player_ids: Array.from(maleIDs.value),
        female_player_ids: Array.from(femaleIDs.value),
      }),
    })
    const { id } = await res.json()
    const detail = await fetch(`/api/fun-sessions/${id}`).then(r => r.json())
    currentSession.value = detail; step.value = 'play'
  } catch (e: any) { showToast('创建失败: ' + e.message) }
}

function openScoreEditor(match: FunMatchItem) {
  scoringMatch.value = match; showEditDialog.value = true
}

async function handleScoreSubmit(g1m: number, g1f: number, g2m: number, g2f: number, g3m?: number, g3f?: number) {
  if (!scoringMatch.value || !currentSession.value) return
  try {
    const body: any = {
      game1_score_male: g1m, game1_score_female: g1f,
      game2_score_male: g2m, game2_score_female: g2f,
    }
    if (g3m !== undefined) { body.game3_score_male = g3m; body.game3_score_female = g3f }
    await fetch(`/api/fun-sessions/${currentSession.value.id}/matches/${scoringMatch.value.id}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const detail = await fetch(`/api/fun-sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
    showEditDialog.value = false
  } catch (e: any) { showToast('提交失败') }
}

async function refreshSession() {
  if (!currentSession.value) return
  try {
    const detail = await fetch(`/api/fun-sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
  } catch (e) { /* ignore */ }
}

async function completeSession() {
  if (!currentSession.value) return
  try {
    await fetch(`/api/fun-sessions/${currentSession.value.id}/complete`, { method: 'POST' })
    const detail = await fetch(`/api/fun-sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail; step.value = 'result'
  } catch (e: any) { showToast('操作失败') }
}

async function viewSession(session: FunSessionSummary) {
  loadingSession.value = true
  try {
    const detail = await fetch(`/api/fun-sessions/${session.id}`).then(r => r.json())
    currentSession.value = detail
    step.value = session.status === 'completed' ? 'result' : 'play'
  } catch (e) { showToast('加载失败') }
  finally { loadingSession.value = false }
}

async function handleCancelSession() {
  if (!currentSession.value) return
  try {
    await fetch(`/api/fun-sessions/${currentSession.value.id}`, { method: 'DELETE' })
    backToList()
  } catch (e: any) { showToast('取消失败') }
}

function backToList() { step.value = 'list'; currentSession.value = null; loadAll() }

function selectedMalePlayers() {
  return Array.from(maleIDs.value).map(id => players.value.find(p => p.id === id)!).filter(Boolean)
}
function selectedFemalePlayers() {
  return Array.from(femaleIDs.value).map(id => players.value.find(p => p.id === id)!).filter(Boolean)
}

function teamPlayers(team: string): FunPlayer[] {
  if (!currentSession.value) return []
  return currentSession.value.players.filter(p => p.team === team)
}

function playerWins(pid: number): number {
  if (!currentSession.value) return 0
  return currentSession.value.matches.filter(m => m.played && m.winner_id === pid).length
}
function playerLosses(pid: number): number {
  if (!currentSession.value) return 0
  return currentSession.value.matches.filter(m => m.played && m.winner_id && m.winner_id !== pid && (m.male_player_id === pid || m.female_player_id === pid)).length
}

function drawLabel(d: any): string {
  switch (d.card_type) {
    case 'handicap': return `让${d.card_value}分`
    case 'spin': return d.card_detail === 'topspin' ? '上旋' : '下旋'
    case 'table': return d.card_detail === 'left' ? '左半台' : '右半台'
    case 'defense': return '防守'
    default: return '?'
  }
}
function drawColor(d: any): string {
  switch (d.card_type) {
    case 'handicap': return '#f5a623'
    case 'spin': return '#1989fa'
    case 'table': return '#9b59b6'
    case 'defense': return '#e74c3c'
    default: return '#666'
  }
}

function playerById(id: number): FunPlayer | undefined {
  return currentSession.value?.players.find(p => p.id === id)
}
</script>

<template>
  <div style="min-height: 100vh; background: #f0f2f5; padding-bottom: 80px;">

    <!-- Loading overlay -->
    <div v-if="loadingSession" style="position: fixed; inset: 0; background: rgba(255,255,255,0.9); z-index: 1000; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px;">
      <div style="width: 40px; height: 40px; border: 4px solid #ebedf0; border-top-color: #1989fa; border-radius: 50%; animation: spin 0.8s linear infinite;"></div>
      <span style="color: #969799; font-size: 14px;">加载中...</span>
    </div>

    <!-- Initial loading -->
    <div v-if="loadingPlayers" style="text-align: center; padding: 120px 20px;">
      <div style="width: 36px; height: 36px; border: 3px solid #ebedf0; border-top-color: #1989fa; border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 12px;"></div>
      <span style="color: #969799;">加载中...</span>
    </div>

    <template v-else>
      <!-- Hero -->
      <div style="background: linear-gradient(135deg, #ff6b9d, #c44569); color: #fff; padding: 24px 20px 20px;">
        <div style="font-size: 22px; font-weight: 700;">
          <IconTournament :size="26" :stroke-width="2" style="vertical-align: -5px; margin-right: 6px;" />
          趣味赛
        </div>
        <div style="font-size: 13px; opacity: 0.8; margin-top: 4px;">
          {{ step === 'list' ? '趣味团体赛' : step === 'select' ? `选${teamLabels.a}和${teamLabels.b}` : step === 'confirm' ? '确认对阵信息' : step === 'play' ? '逐场录入比分' : '比赛结果' }}
        </div>
      </div>

      <!-- ===== LIST ===== -->
      <template v-if="step === 'list'">
        <div style="padding: 16px;">
          <button @click="enterSelect" style="width: 100%; padding: 16px; background: linear-gradient(135deg, #ff6b9d, #c44569); color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer; box-shadow: 0 4px 16px rgba(255,107,157,0.3);">
            <IconPlus :size="20" :stroke-width="2" style="vertical-align: -4px; margin-right: 4px;" />
            创建趣味赛
          </button>
        </div>

        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconList :size="18" :stroke-width="2" style="vertical-align: -3px;" />
          历史趣味赛
        </div>

        <div v-if="sessions.length === 0" style="text-align: center; padding: 60px 20px; color: #969799;">
          <p>暂无趣味赛记录</p>
        </div>

        <div v-for="s in sessions" :key="s.id" @click="viewSession(s)"
          style="background: #fff; border-radius: 12px; padding: 16px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); cursor: pointer;">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <div style="font-weight: 600; font-size: 16px;">{{ s.name }}</div>
              <div style="font-size: 13px; color: #969799; margin-top: 2px;">
                男{{ s.male_count }}人 vs 女{{ s.female_count }}人 · {{ s.match_count }} 场
              </div>
              <div style="font-size: 13px; font-weight: 600; margin-top: 2px;" :style="{color: s.status === 'completed' ? (s.male_wins > s.female_wins ? '#1989fa' : '#ee0a24') : '#969799'}">
                {{ s.status === 'completed' ? `男 ${s.male_wins} : ${s.female_wins} 女` : `已完 ${s.match_count - s.unplayed_count}/${s.match_count}` }}
              </div>
            </div>
            <div style="display: flex; align-items: center; gap: 8px;">
              <span style="font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: 10px;"
                :style="s.status==='completed'?'background:#e8f8ef;color:#07c160;':'background:#e8f4ff;color:#1989fa;'">
                {{ s.status === 'completed' ? '已结束' : `剩${s.unplayed_count || 0}场` }}
              </span>
              <IconChevronRight :size="16" :stroke-width="2" style="color: #c8c9cc;" />
            </div>
          </div>
        </div>
      </template>

      <!-- ===== SELECT ===== -->
      <template v-if="step === 'select'">
        <!-- Mode selector -->
        <div style="padding:8px 16px;">
          <div style="display:flex;gap:6px;flex-wrap:wrap;">
            <button v-for="m in [{v:'gender',l:'男女对抗'},{v:'rubber',l:'胶皮大战'},{v:'pimple_rr',l:'全颗粒大循环'}]" :key="m.v"
              @click="matchMode=m.v; if(m.v==='rubber'){}"
              style="flex:1;padding:10px 8px;border-radius:10px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;text-align:center;min-width:0;"
              :style="matchMode===m.v?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'#fff',color:'#666',borderColor:'#ddd'}">
              {{ m.l }}</button>
          </div>
        </div>

        <!-- Male team -->
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconUsers :size="18" :stroke-width="2" style="vertical-align: -3px; color: #1989fa;" />
          {{ teamLabels.a }}（已选 {{ maleIDs.size }} 人）
        </div>
        <div style="background: #fff; border-radius: 12px; margin: 4px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-for="p in players" :key="'m'+p.id" @click="toggleMale(p.id)"
            style="display: flex; align-items: center; padding: 12px 16px; border-bottom: 1px solid #f5f5f5; cursor: pointer;"
            :style="{ background: maleIDs.has(p.id) ? '#e8f4ff' : '#fff', opacity: femaleIDs.has(p.id) ? 0.4 : 1 }">
            <input type="checkbox" :checked="maleIDs.has(p.id)" style="width: 18px; height: 18px; margin-right: 12px; accent-color: #1989fa;" />
            <div style="flex: 1;">
              <div style="font-size: 16px; font-weight: 500;">{{ p.name }}</div>
              <div style="font-size: 13px; color: #969799;">{{ p.current_rating }} 分</div>
            </div>
          </div>
        </div>

        <!-- Female / Team B (hidden for single-group mode) -->
        <template v-if="!isPimpleRR">
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px; margin-top: 8px;">
          <IconUsers :size="18" :stroke-width="2" style="vertical-align: -3px; color: #ee0a24;" />
          {{ teamLabels.b }}（已选 {{ femaleIDs.size }} 人）
        </div>
        <div style="background: #fff; border-radius: 12px; margin: 4px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-for="p in players" :key="'f'+p.id" @click="toggleFemale(p.id)"
            style="display: flex; align-items: center; padding: 12px 16px; border-bottom: 1px solid #f5f5f5; cursor: pointer;"
            :style="{ background: femaleIDs.has(p.id) ? '#fde8ef' : '#fff', opacity: maleIDs.has(p.id) ? 0.4 : 1 }">
            <input type="checkbox" :checked="femaleIDs.has(p.id)" style="width: 18px; height: 18px; margin-right: 12px; accent-color: #ee0a24;" />
            <div style="flex: 1;">
              <div style="font-size: 16px; font-weight: 500;">{{ p.name }}</div>
              <div style="font-size: 13px; color: #969799;">{{ p.current_rating }} 分</div>
            </div>
          </div>
        </div>

        <div style="padding: 16px;">
          <input v-model="sessionName" placeholder="趣味赛名称（例：移动杯第二届）"
            style="width: 100%; padding: 14px; border: 1px solid #ebedf0; border-radius: 12px; font-size: 15px; outline: none; margin-bottom: 16px; box-sizing: border-box;" />
          <button :disabled="maleIDs.size === 0 || femaleIDs.size === 0" @click="goConfirm"
            style="width: 100%; padding: 16px; background: linear-gradient(135deg, #ff6b9d, #c44569); color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;"
            :style="{ opacity: (maleIDs.size === 0 || femaleIDs.size === 0) ? 0.5 : 1 }">
            下一步（{{ isPimpleRR ? `已选${maleIDs.size}人` : `${teamLabels.ta}${maleIDs.size}人 ${teamLabels.tb}${femaleIDs.size}人` }}）
          </button>
          <div style="text-align: center; margin-top: 12px;">
            <button @click="step = 'list'" style="background: none; border: none; color: #969799; font-size: 14px; cursor: pointer;">返回列表</button>
          </div>
        </div>
        </template>
      </template>

      <!-- ===== CONFIRM ===== -->
      <template v-if="step === 'confirm'">
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px;">
          确认对阵信息
        </div>

        <div style="background: #fff; border-radius: 12px; padding: 16px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">
          <div style="font-weight: 600; margin-bottom: 12px; font-size: 16px;">{{ sessionName }}</div>
          <div style="margin-bottom: 8px;">
            <span style="font-size: 14px; color: #1989fa; font-weight: 600;">{{ teamLabels.a }} ({{ maleIDs.size }}人): </span>
            <span v-for="p in selectedMalePlayers()" :key="p.id" style="font-size: 14px; padding: 4px 8px; background: #e8f4ff; color: #1989fa; border-radius: 6px; margin: 2px; display: inline-block;">{{ p.name }} ({{ p.current_rating }})</span>
          </div>
          <div style="margin-bottom: 8px;">
            <span style="font-size: 14px; color: #ee0a24; font-weight: 600;">{{ teamLabels.b }} ({{ femaleIDs.size }}人): </span>
            <span v-for="p in selectedFemalePlayers()" :key="p.id" style="font-size: 14px; padding: 4px 8px; background: #fde8ef; color: #ee0a24; border-radius: 6px; margin: 2px; display: inline-block;">{{ p.name }} ({{ p.current_rating }})</span>
          </div>
          <div style="margin-top: 12px; font-size: 13px; color: #969799;">
            每位{{ teamLabels.ta }}队员 VS 每位{{ teamLabels.tb }}队员，共 {{ maleIDs.size * femaleIDs.size }} 场比赛<br/>
            分差≥50分触发趣味抽卡机制
          </div>
        </div>

        <div style="padding: 16px;">
          <button @click="createSession"
            style="width: 100%; padding: 16px; background: linear-gradient(135deg, #ff6b9d, #c44569); color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;">
            生成对阵表
          </button>
          <div style="text-align: center; margin-top: 12px;">
            <button @click="step = 'select'" style="background: none; border: none; color: #969799; font-size: 14px; cursor: pointer;">返回修改</button>
          </div>
        </div>
      </template>

      <!-- ===== PLAY ===== -->
      <template v-if="step === 'play' && currentSession">
        <FunMatchPlay
          :session-name="currentSession.name"
          :session-id="currentSession.id"
          :male-wins="currentSession.male_wins"
          :female-wins="currentSession.female_wins"
          :male-game-wins="currentSession.male_game_wins"
          :female-game-wins="currentSession.female_game_wins"
          :male-points="currentSession.male_points"
          :female-points="currentSession.female_points"
          :matches="currentSession.matches"
          :session-players="currentSession.players"
          :show-edit-dialog="showEditDialog"
          :scoring-match="scoringMatch"
          :drawing-card="false"
          @update:show-edit-dialog="showEditDialog = $event"
          @open-score-editor="openScoreEditor"
          @submit-score="handleScoreSubmit"
          @refresh="refreshSession"
          @cancel-session="handleCancelSession"
          @complete-session="completeSession"
          @back-to-list="backToList"
        />
      </template>

      <!-- ===== RESULT ===== -->
      <template v-if="step === 'result' && currentSession">
        <div style="background: #fff; border-radius: 12px; padding: 20px; margin: 12px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">
          <div style="display: flex; align-items: center; justify-content: center; gap: 8px;">
            <IconTrophy :size="22" :stroke-width="2"
              :style="{color: currentSession.winning_team === 'male' ? '#1989fa' : '#ee0a24'}" />
            <span style="font-weight: 700; font-size: 20px;">{{ currentSession.name }}</span>
          </div>
          <!-- Big score -->
          <div style="display:flex;align-items:center;justify-content:center;gap:16px;margin-top:16px;">
            <div style="text-align:center;" :style="{opacity: currentSession.winning_team === 'female' ? 0.3 : 1}">
              <div style="font-size:12px;color:#969799;">男队</div>
              <div style="font-size:48px;font-weight:800;color:#1989fa;">{{ currentSession.male_wins }}</div>
            </div>
            <div style="font-size:24px;font-weight:800;color:#c8c9cc;">:</div>
            <div style="text-align:center;" :style="{opacity: currentSession.winning_team === 'male' ? 0.3 : 1}">
              <div style="font-size:12px;color:#969799;">女队</div>
              <div style="font-size:48px;font-weight:800;color:#ee0a24;">{{ currentSession.female_wins }}</div>
            </div>
          </div>
          <div style="font-size: 13px; color: #969799; margin-top: 4px; text-align: center;">
            {{ currentSession.winning_team === 'male' ? '男队获胜！' : currentSession.winning_team === 'female' ? '女队获胜！' : '平局' }}
          </div>
          <!-- Detailed stats -->
          <div style="display:flex;justify-content:center;gap:24px;margin-top:12px;font-size:12px;color:#666;">
            <div>总局数 男{{ currentSession.male_game_wins }}:{{ currentSession.female_game_wins }}女</div>
            <div>总分数 男{{ currentSession.male_points }}:{{ currentSession.female_points }}女</div>
          </div>
        </div>

        <!-- 赛事荣誉 -->
        <FunMatchHonors :matches="currentSession.matches" :players="currentSession.players" :show="currentSession.status==='completed'" />

        <!-- Player performance in this session -->
        <div v-if="teamPlayers('male').length > 0" style="margin:8px 16px 4px;">
          <div style="font-weight:600;font-size:14px;color:#1989fa;margin-bottom:4px;">男队本场战绩</div>
          <div style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,0.06);">
            <div v-for="p in teamPlayers('male')" :key="p.id"
              style="display:flex;align-items:center;padding:10px 16px;border-bottom:1px solid #f5f5f5;">
              <span style="font-size:14px;font-weight:500;">{{ p.name }}</span>
              <span style="font-size:13px;margin-left:auto;"
                :style="{color: playerWins(p.id) >= playerLosses(p.id) ? '#07c160' : '#ee0a24'}">{{ playerWins(p.id) }}胜 {{ playerLosses(p.id) }}负</span>
            </div>
          </div>
        </div>
        <div v-if="teamPlayers('female').length > 0" style="margin:4px 16px 8px;">
          <div style="font-weight:600;font-size:14px;color:#ee0a24;margin-bottom:4px;">女队本场战绩</div>
          <div style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 12px rgba(0,0,0,0.06);">
            <div v-for="p in teamPlayers('female')" :key="p.id"
              style="display:flex;align-items:center;padding:10px 16px;border-bottom:1px solid #f5f5f5;">
              <span style="font-size:14px;font-weight:500;">{{ p.name }}</span>
              <span style="font-size:13px;margin-left:auto;"
                :style="{color: playerWins(p.id) >= playerLosses(p.id) ? '#07c160' : '#ee0a24'}">{{ playerWins(p.id) }}胜 {{ playerLosses(p.id) }}负</span>
            </div>
          </div>
        </div>

        <!-- All matches with card details -->
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px;">全部对阵</div>
        <div style="background: #fff; border-radius: 12px; margin: 4px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-for="m in currentSession.matches" :key="m.id"
            style="padding: 12px 16px; border-bottom: 1px solid #f5f5f5;">
            <!-- Card info -->
            <div v-if="m.draws && m.draws.length > 0" style="margin-bottom:4px;display:flex;flex-wrap:wrap;gap:4px;">
              <span v-for="d in m.draws.filter((x:any)=>!x.cancelled)" :key="d.id"
                style="font-size:10px;padding:2px 6px;border-radius:6px;color:#fff;display:flex;align-items:center;gap:3px;"
                :style="{background: drawColor(d)}">
                👤{{ playerById(d.player_id)?.name }} 🎯{{ drawLabel(d) }}
              </span>
              <span v-for="d in m.draws.filter((x:any)=>x.cancelled)" :key="d.id"
                style="font-size:10px;padding:2px 6px;border-radius:6px;background:#ccc;color:#999;text-decoration:line-through;">
                {{ drawLabel(d) }}
              </span>
            </div>
            <!-- Players and scores -->
            <div style="display: flex; align-items: center;">
              <div style="flex: 1; text-align: right; font-weight: 400;"
                :style="{ fontWeight: m.winner_id === m.male_player_id ? 700 : 400, color: m.winner_id && m.winner_id !== m.male_player_id && m.played ? '#999' : '#333' }">
                {{ m.male_player_name }}
              </div>
              <div style="width: 140px; text-align: center; font-weight: 700; font-size: 13px;">
                <template v-if="m.played && m.game3_score_male != null">
                  ①{{ m.game1_score_male }}:{{ m.game1_score_female }} ②{{ m.game2_score_male }}:{{ m.game2_score_female }} ③{{ m.game3_score_male }}:{{ m.game3_score_female }}
                </template>
                <template v-else-if="m.played">
                  ①{{ m.game1_score_male }}:{{ m.game1_score_female }} ②{{ m.game2_score_male }}:{{ m.game2_score_female }}
                </template>
                <template v-else>
                  <span style="color: #c8c9cc;">-</span>
                </template>
              </div>
              <div style="flex: 1; font-weight: 400;"
                :style="{ fontWeight: m.winner_id === m.female_player_id ? 700 : 400, color: m.winner_id && m.winner_id !== m.female_player_id && m.played ? '#999' : '#333' }">
                {{ m.female_player_name }}
              </div>
            </div>
          </div>
        </div>

        <div style="padding: 16px;">
          <button @click="backToList" style="width: 100%; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;">
            返回列表
          </button>
        </div>
      </template>
    </template>
  </div>
</template>
