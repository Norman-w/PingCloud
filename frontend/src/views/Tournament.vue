<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import TournamentGroupView from '../components/TournamentGroupView.vue'
import TournamentKnockout from '../components/TournamentKnockout.vue'

// ── types ──
interface TournamentSummary {
  id: number; name: string; group_count: number; teams_per_group: number
  players_per_team: number; max_participants: number; seed_enabled: boolean
  seed_count: number; status: string; phase: string
  confirmed_count: number; waitlisted_count: number
  registration_deadline: string; created_at: string
}

interface TournamentDetail {
  id: number; name: string; group_count: number; teams_per_group: number
  players_per_team: number; max_participants: number; seed_enabled: boolean
  seed_count: number; registration_deadline: string; status: string; phase: string
  created_at: string; completed_at: string
  teams: Team[]; registrations: Registration[]; team_matches: TeamMatch[]
}

interface Team {
  id: number; tournament_id: number; group_name: string; team_index: number
  team_name: string; knockout_seed: number | null; group_rank: number | null
  group_wins: number; group_losses: number; group_points: number
  players: TeamPlayer[]
}

interface TeamPlayer {
  id: number; name: string; current_rating: number; reference_rating: number
  role: string; is_seed: boolean
}

interface Registration {
  id: number; player_id: number; player_name: string; status: string
  waitlist_pos: number | null; registered_at: string
}

interface TeamMatch {
  id: number; tournament_id: number; phase: string; round: number; group_name: string
  team_a_id: number; team_b_id: number; team_a_name: string; team_b_name: string
  team_a_wins: number; team_b_wins: number; winner_team_id: number | null
  played: boolean; matches: Match[]; cards: Card[]
}

interface Match {
  id: number; team_match_id: number; phase: string; round: number; group_name: string
  team_a_id: number; team_b_id: number; match_order: number; match_type: string
  player_a_id: number; player_b_id: number; player_a_name: string; player_b_name: string
  player_a2_id: number | null; player_b2_id: number | null
  player_a2_name: string; player_b2_name: string
  game1_score_a: number | null; game1_score_b: number | null
  game2_score_a: number | null; game2_score_b: number | null
  game3_score_a: number | null; game3_score_b: number | null
  winner_id: number | null; winner_team_id: number | null; played: boolean; forfeit: boolean
}

interface Card { id: number; team_match_id: number; team_id: number; card_type: string; drawn_at: string }

// ── state ──
const step = ref<'list' | 'config' | 'registration' | 'draw' | 'play' | 'result'>('list')
const tournaments = ref<TournamentSummary[]>([])
const detail = ref<TournamentDetail | null>(null)
const allPlayers = ref<{ id: number; name: string; current_rating: number }[]>([])
const loading = ref(false)
const currentId = ref(0)
const viewingTeamMatch = ref<TeamMatch | null>(null)

// ── config form ──
const cfg = ref({ name: '锦标赛', group_count: 2, teams_per_group: 3, players_per_team: 3, max_participants: 18, seed_enabled: false, seed_count: 0, registration_deadline: '' })

// ── load ──
async function loadList() {
  try { tournaments.value = await fetch('/api/tournaments').then(r => r.json()) } catch { tournaments.value = [] }
}

async function loadPlayers() {
  try { allPlayers.value = await fetch('/api/players').then(r => r.json()) } catch { allPlayers.value = [] }
}

async function loadDetail(id: number) {
  loading.value = true
  try { detail.value = await fetch(`/api/tournaments/${id}`).then(r => r.json()) } catch { detail.value = null }
  loading.value = false
}

// ── actions ──
async function createTournament() {
  const body: any = { ...cfg.value }
  if (!body.registration_deadline) delete body.registration_deadline
  try {
    const r = await fetch('/api/tournaments', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
    if (!r.ok) { showToast('创建失败'); return }
    const d = await r.json()
    currentId.value = d.id
    await loadDetail(d.id)
    await loadPlayers()
    step.value = 'registration'
  } catch (e: any) { showToast('创建失败: ' + e.message) }
}

async function registerPlayer(playerId: number) {
  try {
    const r = await fetch(`/api/tournaments/${currentId.value}/register`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ player_id: playerId }) })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    await loadDetail(currentId.value)
  } catch (e: any) { showToast('报名失败: ' + e.message) }
}

async function cancelRegistration(playerId: number) {
  try {
    const r = await fetch(`/api/tournaments/${currentId.value}/cancel`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ player_id: playerId }) })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    await loadDetail(currentId.value)
  } catch (e: any) { showToast('取消失败: ' + e.message) }
}

async function drawTeams() {
  try {
    const r = await fetch(`/api/tournaments/${currentId.value}/draw-teams`, { method: 'POST' })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    await loadDetail(currentId.value)
    step.value = 'draw'
  } catch (e: any) { showToast('抽签失败: ' + e.message) }
}

async function generateGroup() {
  try {
    const r = await fetch(`/api/tournaments/${currentId.value}/generate-group`, { method: 'POST' })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    await loadDetail(currentId.value)
    step.value = 'play'
  } catch (e: any) { showToast('生成赛程失败: ' + e.message) }
}

async function advanceKnockout() {
  try {
    const r = await fetch(`/api/tournaments/${currentId.value}/advance-knockout`, { method: 'POST' })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    await loadDetail(currentId.value)
  } catch (e: any) { showToast('晋级失败: ' + e.message) }
}

async function viewTournament(id: number) {
  currentId.value = id
  await loadDetail(id)
  if (!detail.value) return
  if (detail.value.status === 'registration') await loadPlayers()
  const s = detail.value.status
  const p = detail.value.phase
  if (s === 'registration') step.value = 'registration'
  else if (s === 'group_stage' || s === 'knockout') step.value = (p === 'completed' || s === 'completed') ? 'result' : 'play'
  else if (s === 'completed') step.value = 'result'
}

function requiredPlayers(): number {
  if (!detail.value) return 0
  // 每组至少 2 队才能循环；总人数须整除每队人数
  return detail.value.group_count * 2 * detail.value.players_per_team
}

function confirmedCount(): number {
  if (!detail.value) return 0
  return detail.value.registrations.filter(r => r.status === 'confirmed').length
}

function canDrawTeams(): boolean {
  if (!detail.value) return false
  const n = confirmedCount()
  const ppt = detail.value.players_per_team
  if (ppt <= 0 || n < requiredPlayers()) return false
  return n % ppt === 0
}

function drawBlockReason(): string {
  if (!detail.value) return ''
  const n = confirmedCount()
  const ppt = detail.value.players_per_team
  const min = requiredPlayers()
  if (n < min) return `至少需要 ${min} 人（每组≥2队）才能抽签（当前 ${n}/${min}）`
  if (n % ppt !== 0) return `报名人数需为每队 ${ppt} 人的整数倍（当前 ${n} 人，差 ${ppt - (n % ppt)} 人凑整）`
  return ''
}

function onDrawTeamsClick() {
  const reason = drawBlockReason()
  if (reason) {
    showToast(reason)
    return
  }
  drawTeams()
}

function drawPreviewHint(): string {
  if (!detail.value || !canDrawTeams()) return ''
  const n = confirmedCount()
  const ppt = detail.value.players_per_team
  const gc = detail.value.group_count
  const totalTeams = n / ppt
  const base = Math.floor(totalTeams / gc)
  const extra = totalTeams % gc
  if (extra === 0) return `将组成 ${totalTeams} 队，均分到 ${gc} 组（每组 ${base} 队）`
  return `将组成 ${totalTeams} 队，随机分配到 ${gc} 组（${base} 队 / ${base + 1} 队）`
}

function backToList() { step.value = 'list'; detail.value = null }

onMounted(() => { loadList() })

// ── format helpers ──
function fmtDate(s: string) { if (!s) return ''; return s.replace('T', ' ').substring(0, 16) }
function phaseLabel(p: string) { const m: Record<string, string> = { registration: '报名中', group: '小组赛', semifinal: '半决赛', final: '决赛', completed: '已结束' }; return m[p] || p }
function cardLabel(t: string) { return t === 'edge_double' ? '擦边翻倍卡' : t === 'net_deduction' ? '擦网扣分卡' : t }
</script>

<template>
  <!-- ===== LIST ===== -->
  <div v-if="step === 'list'">
    <div style="background: linear-gradient(135deg, #1a56db 0%, #1e88e5 50%, #00bfa5 100%); color: #fff; padding: 28px 20px 24px;">
      <h2 style="font-size: 22px; font-weight: 800; margin: 0 0 4px;">🏆 锦标赛</h2>
      <p style="font-size: 13px; opacity: 0.85; margin: 0;">小组循环 + 淘汰晋级 · 五场三胜制团体赛</p>
    </div>

    <div style="padding: 16px;">
      <button @click="step = 'config'"
        style="width: 100%; padding: 16px; background: linear-gradient(135deg, #1989fa, #1e88e5); color: #fff; border: none; border-radius: 16px; font-size: 17px; font-weight: 700; cursor: pointer; box-shadow: 0 4px 16px rgba(25,137,250,0.3);">
        ＋ 创建锦标赛
      </button>
    </div>

    <div v-if="tournaments.length === 0" style="text-align: center; padding: 60px 20px; color: #969799;">
      <div style="font-size: 48px; margin-bottom: 12px;">🏓</div>
      <div style="font-size: 15px;">暂无锦标赛记录</div>
    </div>

    <div v-for="t in tournaments" :key="t.id" class="card" style="cursor: pointer;" @click="viewTournament(t.id)">
      <div style="display: flex; justify-content: space-between; align-items: flex-start;">
        <div>
          <div style="font-weight: 700; font-size: 16px;">{{ t.name }}</div>
          <div style="font-size: 12px; color: #969799; margin-top: 4px;">
            {{ t.group_count }}组×{{ t.teams_per_group }}队×{{ t.players_per_team }}人
            <span v-if="t.seed_enabled"> · 种子{{ t.seed_count }}</span>
          </div>
        </div>
        <span class="badge" :class="t.status === 'completed' ? 'badge-success' : t.status === 'registration' ? 'badge-primary' : 'badge-warning'">
          {{ phaseLabel(t.phase) }}
        </span>
      </div>
      <div style="display: flex; gap: 16px; margin-top: 8px; font-size: 12px; color: #969799;">
        <span>✅ {{ t.confirmed_count }}/{{ t.max_participants }}</span>
        <span v-if="t.waitlisted_count > 0">⏳ 候补 {{ t.waitlisted_count }}</span>
        <span>{{ fmtDate(t.created_at) }}</span>
      </div>
    </div>
  </div>

  <!-- ===== CONFIG ===== -->
  <div v-if="step === 'config'">
    <div style="background: linear-gradient(135deg, #1a56db 0%, #1e88e5 50%, #00bfa5 100%); color: #fff; padding: 20px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <button @click="step = 'list'" style="background: none; border: none; color: #fff; font-size: 24px; cursor: pointer; padding: 0;">←</button>
        <h2 style="font-size: 18px; font-weight: 700; margin: 0;">创建锦标赛</h2>
      </div>
    </div>

    <div class="card">
      <div style="margin-bottom: 14px;">
        <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">赛事名称</label>
        <input v-model="cfg.name" style="width: 100%; padding: 12px; border: 1.5px solid #ebedf0; border-radius: 10px; font-size: 15px; outline: none; box-sizing: border-box;" placeholder="输入赛事名称" />
      </div>

      <div style="display: flex; gap: 10px; margin-bottom: 14px;">
        <div style="flex: 1;">
          <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">组数</label>
          <input v-model.number="cfg.group_count" type="number" min="1" max="8"
            style="width: 100%; padding: 12px; border: 2px solid #1989fa; border-radius: 10px; font-size: 20px; font-weight: 700; text-align: center; outline: none; box-sizing: border-box;" />
        </div>
        <div style="flex: 1;">
          <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">参考队数/组</label>
          <input v-model.number="cfg.teams_per_group" type="number" min="2" max="8"
            style="width: 100%; padding: 12px; border: 2px solid #ebedf0; border-radius: 10px; font-size: 20px; font-weight: 700; text-align: center; outline: none; box-sizing: border-box;" />
        </div>
        <div style="flex: 1;">
          <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">每队人数</label>
          <input v-model.number="cfg.players_per_team" type="number" min="2" max="5"
            style="width: 100%; padding: 12px; border: 2px solid #ebedf0; border-radius: 10px; font-size: 20px; font-weight: 700; text-align: center; outline: none; box-sizing: border-box;" />
        </div>
      </div>

      <div style="margin-bottom: 14px;">
        <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">报名上限</label>
        <input v-model.number="cfg.max_participants" type="number" min="1"
          style="width: 100%; padding: 12px; border: 1.5px solid #ebedf0; border-radius: 10px; font-size: 15px; outline: none; box-sizing: border-box;" />
        <div style="font-size: 11px; color: #969799; margin-top: 2px;">参考总人数 = {{ cfg.group_count }}组 × {{ cfg.teams_per_group }}队 × {{ cfg.players_per_team }}人 = {{ cfg.group_count * cfg.teams_per_group * cfg.players_per_team }}人；抽签时按实际报名人数组队，组间队数可不均（如 3队+4队）</div>
      </div>

      <div style="margin-bottom: 14px;">
        <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">报名截止时间（可选）</label>
        <input v-model="cfg.registration_deadline" type="datetime-local"
          style="width: 100%; padding: 12px; border: 1.5px solid #ebedf0; border-radius: 10px; font-size: 15px; outline: none; box-sizing: border-box;" />
      </div>

      <div style="margin-bottom: 14px;">
        <label style="display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; font-weight: 600;">
          <input type="checkbox" v-model="cfg.seed_enabled" style="width: 20px; height: 20px;" />
          启用种子选手
        </label>
        <div v-if="cfg.seed_enabled" style="margin-top: 8px;">
          <label style="font-size: 13px; font-weight: 600; color: #646566; display: block; margin-bottom: 4px;">种子数量</label>
          <input v-model.number="cfg.seed_count" type="number" min="1" :max="cfg.group_count * cfg.teams_per_group"
            style="width: 100%; padding: 12px; border: 1.5px solid #ebedf0; border-radius: 10px; font-size: 15px; outline: none; box-sizing: border-box;" />
          <div style="font-size: 11px; color: #969799; margin-top: 2px;">按开球网积分排名，前 {{ cfg.seed_count || 0 }} 名将分配到不同队伍</div>
        </div>
      </div>

      <button @click="createTournament"
        style="width: 100%; padding: 16px; background: linear-gradient(135deg, #07c160, #06ad56); color: #fff; border: none; border-radius: 14px; font-size: 17px; font-weight: 700; cursor: pointer;">
        创建赛事并进入报名
      </button>
    </div>
  </div>

  <!-- ===== REGISTRATION ===== -->
  <div v-if="step === 'registration' && detail">
    <div style="background: linear-gradient(135deg, #1a56db 0%, #1e88e5 50%, #00bfa5 100%); color: #fff; padding: 20px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <button @click="backToList" style="background: none; border: none; color: #fff; font-size: 24px; cursor: pointer; padding: 0;">←</button>
        <div>
          <h2 style="font-size: 18px; font-weight: 700; margin: 0;">{{ detail.name }}</h2>
          <div style="font-size: 12px; opacity: 0.85;">{{ detail.group_count }}组×{{ detail.teams_per_group }}队×{{ detail.players_per_team }}人 · 上限 {{ detail.max_participants }} 人</div>
        </div>
      </div>
      <!-- Progress bar -->
      <div style="margin-top: 12px; background: rgba(255,255,255,0.2); border-radius: 8px; height: 8px; overflow: hidden;">
        <div :style="{ width: (detail.registrations.filter(r => r.status === 'confirmed').length / detail.max_participants * 100) + '%', background: '#fff', height: '100%', borderRadius: '8px', transition: 'width 0.3s' }"></div>
      </div>
      <div style="display: flex; justify-content: space-between; font-size: 12px; margin-top: 4px;">
        <span>✅ {{ detail.registrations.filter(r => r.status === 'confirmed').length }}/{{ detail.max_participants }}</span>
        <span>⏳ 候补 {{ detail.registrations.filter(r => r.status === 'waitlisted').length }}</span>
      </div>
    </div>

    <!-- Confirmed -->
    <div class="section-title">已报名 ({{ detail.registrations.filter(r => r.status === 'confirmed').length }})</div>
    <div v-for="reg in detail.registrations.filter(r => r.status === 'confirmed')" :key="reg.id" class="card" style="margin-top: 0; margin-bottom: 0; border-bottom: 1px solid #f5f5f5;">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <span style="font-weight: 600;">{{ reg.player_name }}</span>
        <button @click="cancelRegistration(reg.player_id)"
          style="padding: 6px 14px; background: none; border: 1.5px solid #ee0a24; border-radius: 16px; color: #ee0a24; font-size: 12px; font-weight: 600; cursor: pointer;">取消</button>
      </div>
    </div>

    <!-- Waitlisted -->
    <div v-if="detail.registrations.filter(r => r.status === 'waitlisted').length > 0" class="section-title">候补队列 ({{ detail.registrations.filter(r => r.status === 'waitlisted').length }})</div>
    <div v-for="reg in detail.registrations.filter(r => r.status === 'waitlisted')" :key="reg.id" class="card" style="margin-top: 0; margin-bottom: 0; opacity: 0.6;">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <span>#{{ reg.waitlist_pos }} {{ reg.player_name }}</span>
        <button @click="cancelRegistration(reg.player_id)"
          style="padding: 4px 12px; background: none; border: 1.5px solid #ccc; border-radius: 16px; color: #999; font-size: 12px; cursor: pointer;">取消</button>
      </div>
    </div>

    <!-- Player picker -->
    <div class="section-title">选择球员报名</div>
    <div style="padding: 0 16px 16px; display: flex; flex-wrap: wrap; gap: 8px;">
      <button v-for="p in allPlayers" :key="p.id"
        :disabled="detail.registrations.some(r => r.player_id === p.id)"
        @click="registerPlayer(p.id)"
        :style="{
          padding: '8px 14px', borderRadius: '20px', border: '1.5px solid ' + (detail.registrations.some(r => r.player_id === p.id) ? '#e0e0e0' : '#1989fa'),
          background: detail.registrations.some(r => r.player_id === p.id) ? '#f5f5f5' : '#fff',
          color: detail.registrations.some(r => r.player_id === p.id) ? '#ccc' : '#1989fa',
          fontSize: '13px', fontWeight: 600, cursor: detail.registrations.some(r => r.player_id === p.id) ? 'default' : 'pointer'
        }">
        {{ p.name }} <span style="font-weight: 400; font-size: 11px;">{{ p.current_rating }}</span>
      </button>
    </div>

    <!-- Actions -->
    <div style="padding: 16px;">
      <div v-if="!canDrawTeams()" style="font-size: 13px; color: #ed6a0c; margin-bottom: 8px; text-align: center;">
        {{ drawBlockReason() }}
      </div>
      <div v-else style="font-size: 13px; color: #07c160; margin-bottom: 8px; text-align: center;">
        {{ drawPreviewHint() }}
      </div>
      <button @click="onDrawTeamsClick"
        :style="{
          width: '100%', padding: '16px', border: 'none', borderRadius: '14px', fontSize: '16px', fontWeight: 700,
          cursor: canDrawTeams() ? 'pointer' : 'not-allowed',
          background: canDrawTeams() ? 'linear-gradient(135deg, #f5a623, #e8961a)' : '#c8c9cc',
          color: '#fff', opacity: canDrawTeams() ? 1 : 0.7
        }">🎲 抽签组队</button>
    </div>
  </div>

  <!-- ===== DRAW ===== -->
  <div v-if="step === 'draw' && detail">
    <div style="background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; padding: 20px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <button @click="backToList" style="background: none; border: none; color: #fff; font-size: 24px; cursor: pointer; padding: 0;">←</button>
        <h2 style="font-size: 18px; font-weight: 700; margin: 0;">抽签结果 · {{ detail.name }}</h2>
      </div>
    </div>

    <div v-for="team of detail.teams" :key="team.id" class="card" style="margin-bottom: 8px;">
      <div style="display: flex; align-items: center; gap: 10px; margin-bottom: 8px;">
        <span style="background: linear-gradient(135deg, #1989fa, #1e88e5); color: #fff; padding: 4px 12px; border-radius: 14px; font-size: 14px; font-weight: 700;">{{ team.team_name }}</span>
        <span style="font-size: 12px; color: #969799;">{{ team.group_name }}组</span>
      </div>
      <div style="display: flex; gap: 6px; flex-wrap: wrap;">
        <span v-for="p in team.players" :key="p.id"
          :style="{ padding: '6px 12px', borderRadius: '18px', fontSize: '13px', fontWeight: 600, background: p.is_seed ? '#fff3e0' : '#f0f2f5', color: p.is_seed ? '#e65100' : '#333', border: p.is_seed ? '1.5px solid #ffcc80' : '1.5px solid transparent' }">
          {{ p.role }}·{{ p.name }}
          <span style="font-size: 10px; font-weight: 400; opacity: 0.7;">{{ p.reference_rating || p.current_rating }}</span>
          <span v-if="p.is_seed" style="font-size: 10px;">⭐</span>
        </span>
      </div>
    </div>

    <div style="padding: 16px; display: flex; gap: 10px;">
      <button @click="step = 'registration'"
        style="flex: 1; padding: 14px; border: 1.5px solid #ccc; border-radius: 14px; background: #fff; font-size: 15px; cursor: pointer;">重新抽签</button>
      <button @click="generateGroup"
        style="flex: 2; padding: 14px; border: none; border-radius: 14px; font-size: 16px; font-weight: 700; cursor: pointer; background: linear-gradient(135deg, #07c160, #06ad56); color: #fff;">✅ 确认并生成赛程</button>
    </div>
  </div>

  <!-- ===== PLAY ===== -->
  <div v-if="step === 'play' && detail">
    <div style="background: linear-gradient(135deg, #1989fa, #1e88e5); color: #fff; padding: 16px 20px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <button @click="backToList" style="background: none; border: none; color: #fff; font-size: 24px; cursor: pointer; padding: 0;">←</button>
        <div>
          <div style="font-size: 16px; font-weight: 700;">{{ detail.name }}</div>
          <div style="font-size: 12px; opacity: 0.85;">{{ phaseLabel(detail.phase) }}</div>
        </div>
      </div>
    </div>

    <!-- Group stage -->
    <TournamentGroupView v-if="detail.phase === 'group'" :detail="detail" :tournament-id="currentId" @refresh="loadDetail(currentId)" @back="backToList" />

    <!-- Knockout -->
    <TournamentKnockout v-if="detail.phase === 'semifinal' || detail.phase === 'final'" :detail="detail" :tournament-id="currentId" @refresh="loadDetail(currentId)" @back="backToList" />

    <!-- Admin advance button -->
    <div style="padding: 16px;">
      <button v-if="detail.phase === 'group' && detail.team_matches.filter(m => m.phase === 'group').every(m => m.played)"
        @click="advanceKnockout"
        style="width: 100%; padding: 16px; border: none; border-radius: 14px; font-size: 16px; font-weight: 700; cursor: pointer; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff;">
        🏆 晋级淘汰赛
      </button>
    </div>
  </div>

  <!-- ===== RESULT ===== -->
  <div v-if="step === 'result' && detail">
    <div style="background: linear-gradient(135deg, #f5a623 0%, #e8961a 50%, #ee0a24 100%); color: #fff; padding: 24px 20px; text-align: center;">
      <div style="font-size: 40px; margin-bottom: 8px;">🏆</div>
      <h2 style="font-size: 20px; font-weight: 800; margin: 0 0 4px;">{{ detail.name }}</h2>
      <div style="font-size: 13px; opacity: 0.85;">赛事已结束</div>
    </div>

    <!-- Final standings -->
    <div class="section-title">最终排名</div>
    <div v-for="team of detail.teams.filter(t => t.group_rank).sort((a, b) => (a.group_rank || 99) - (b.group_rank || 99))" :key="team.id" class="card" style="margin-bottom: 6px;">
      <div style="display: flex; align-items: center; gap: 10px;">
        <span style="font-size: 24px; font-weight: 800; color: #f5a623;">#{{ team.group_rank }}</span>
        <div>
          <div style="font-weight: 700;">{{ team.team_name }}</div>
          <div style="font-size: 12px; color: #969799;">{{ team.group_wins }}胜 {{ team.group_losses }}负</div>
        </div>
      </div>
    </div>

    <div style="padding: 16px;">
      <button @click="backToList" style="width: 100%; padding: 14px; border: none; border-radius: 14px; font-size: 16px; font-weight: 700; cursor: pointer; background: #1989fa; color: #fff;">返回列表</button>
    </div>
  </div>

  <!-- Loading -->
  <div v-if="loading" style="text-align: center; padding: 60px; color: #969799;">加载中...</div>
</template>
