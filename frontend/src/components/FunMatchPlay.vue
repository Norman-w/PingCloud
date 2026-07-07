<script setup lang="ts">
import { ref, computed } from 'vue'
import { IconList } from '@tabler/icons-vue'
import FunScoreDialog from './FunScoreDialog.vue'
import FunCardDraw from './FunCardDraw.vue'
import PlayerStandings from './PlayerStandings.vue'

export interface FunDrawRecord {
  id: number
  card_type: string
  card_value: number | null
  card_detail: string
  player_id: number
  cancelled: boolean
  drawn_at: string
}

export interface FunMatchItem {
  id: number
  male_player_id: number
  female_player_id: number
  male_player_name: string
  female_player_name: string
  game1_score_male: number | null
  game1_score_female: number | null
  game2_score_male: number | null
  game2_score_female: number | null
  game3_score_male: number | null
  game3_score_female: number | null
  winner_id: number | null
  winner_team: string
  handicap_points: number
  played: boolean
  draws: FunDrawRecord[]
}

const props = defineProps<{
  mode: string
  sessionName: string
  sessionId: number
  maleWins: number
  femaleWins: number
  maleGameWins: number
  femaleGameWins: number
  malePoints: number
  femalePoints: number
  matches: FunMatchItem[]
  sessionPlayers: { id: number; current_rating: number; reference_rating: number }[]
  showEditDialog: boolean
  scoringMatch: FunMatchItem | null
  drawingCard: boolean
}>()

const emit = defineEmits<{
  (e: 'update:showEditDialog', v: boolean): void
  (e: 'openScoreEditor', m: FunMatchItem): void
  (e: 'submitScore', g1m: number, g1f: number, g2m: number, g2f: number, g3m?: number, g3f?: number): void
  (e: 'drawCard', matchId: number, playerId: number): void
  (e: 'completeSession'): void
  (e: 'backToList'): void
  (e: 'refresh'): void
  (e: 'cancelSession'): void
}>()

// Individual standings for wheel_rr mode
const isWheel = computed(() => props.mode === 'wheel_rr')

const individualStandings = computed(() => {
  if (!isWheel.value) return []
  const stats = new Map<number, { id: number; name: string; rating: number; wins: number; losses: number }>()
  for (const p of props.sessionPlayers) {
    stats.set(p.id, { id: p.id, name: '', rating: p.reference_rating > 0 ? p.reference_rating : p.current_rating, wins: 0, losses: 0 })
  }
  for (const m of props.matches) {
    if (!m.played) continue
    const a = stats.get(m.male_player_id)
    const b = stats.get(m.female_player_id)
    if (a) { a.name = m.male_player_name; if (m.winner_id === m.male_player_id) a.wins++; else a.losses++ }
    if (b) { b.name = m.female_player_name; if (m.winner_id === m.female_player_id) b.wins++; else b.losses++ }
  }
  // Also fill names from sessionPlayers for any missing
  for (const p of props.sessionPlayers) {
    const s = stats.get(p.id)
    if (s && !s.name) s.name = '' // fallback
  }
  return Array.from(stats.values()).filter(s => s.wins + s.losses > 0 || stats.size <= props.sessionPlayers.length)
    .sort((a, b) => b.wins - a.wins || b.rating - a.rating)
})

function playerRating(pid: number): number {
  const p = props.sessionPlayers.find(sp => sp.id === pid)
  if (!p) return 0
  return p.reference_rating > 0 ? p.reference_rating : p.current_rating
}

// Card draw state
const showCardDraw = ref(false)
const drawMatchId = ref(0)
const drawMaleName = ref('')
const drawFemaleName = ref('')
const drawMaleRating = ref(0)
const drawFemaleRating = ref(0)
const drawResult = ref<{ card_type: string; card_value: number | null; card_detail: string } | null>(null)
const confirmRedraw = ref(false)
const redrawMatch = ref<FunMatchItem | null>(null)

function openCardDraw(m: FunMatchItem) {
  // If already has draws, confirm redraw first
  if (m.draws && m.draws.length > 0) {
    redrawMatch.value = m
    confirmRedraw.value = true
    return
  }
  _startCardDraw(m)
}

function matchRating(playerId: number): number {
  const p = props.sessionPlayers.find(sp => sp.id === playerId)
  if (!p) return 0
  return p.reference_rating > 0 ? p.reference_rating : p.current_rating
}

function canDraw(m: FunMatchItem): boolean {
  const mr = matchRating(m.male_player_id)
  const fr = matchRating(m.female_player_id)
  return Math.abs(mr - fr) >= 50
}

function _startCardDraw(m: FunMatchItem) {
  drawMatchId.value = m.id
  drawMaleName.value = m.male_player_name
  drawFemaleName.value = m.female_player_name
  const mr = matchRating(m.male_player_id)
  const fr = matchRating(m.female_player_id)
  drawMaleRating.value = mr
  drawFemaleRating.value = fr
  drawResult.value = null
  showCardDraw.value = true
}

async function handleCardDraw() {
  if (drawMatchId.value === 0) return
  const m = props.matches.find(x => x.id === drawMatchId.value)
  if (!m) return
  const higherId = drawMaleRating.value >= drawFemaleRating.value ? m.male_player_id : m.female_player_id

  try {
    const res = await fetch(`/api/fun-sessions/${props.sessionId}/draw-card/${drawMatchId.value}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ player_id: higherId }),
    })
    drawResult.value = await res.json()
    // Refresh session to get updated handicap_points
    emit('refresh')
  } catch (e) { /* ignore */ }
}

function handleCardDrawClose() {
  showCardDraw.value = false
  drawMatchId.value = 0
  drawResult.value = null
}

function matchScoreSummary(m: FunMatchItem): string {
  if (!m.played) return '待录入'
  const g1 = `${m.game1_score_male}:${m.game1_score_female}`
  const g2 = `${m.game2_score_male}:${m.game2_score_female}`
  if (m.game3_score_male != null) return `${g1} ${g2} ${m.game3_score_male}:${m.game3_score_female}`
  return `${g1} ${g2}`
}

function matchWinnerName(m: FunMatchItem): string {
  if (!m.played || !m.winner_id) return ''
  return m.winner_id === m.male_player_id ? m.male_player_name : m.female_player_name
}

function drawLabel(d: FunDrawRecord): string {
  switch (d.card_type) {
    case 'handicap': return `让${d.card_value}分`
    case 'spin': return d.card_detail === 'topspin' ? '上旋' : '下旋'
    case 'table': return d.card_detail === 'left' ? '左半台' : '右半台'
    case 'defense': return '防守'
    default: return '?'
  }
}

function drawColor(d: FunDrawRecord): string {
  switch (d.card_type) {
    case 'handicap': return '#f5a623'
    case 'spin': return '#1989fa'
    case 'table': return '#9b59b6'
    case 'defense': return '#e74c3c'
    default: return '#666'
  }
}
</script>

<template>
  <div>
    <!-- Session info -->
    <div style="background:#fff;border-radius:12px;padding:16px;margin:12px 16px;box-shadow:0 2px 12px rgba(0,0,0,0.06);">
      <div style="text-align:center;">
        <span style="font-weight:700;font-size:18px;">{{ sessionName }}</span>
      </div>
      <div style="font-size:13px;color:#969799;margin-top:6px;text-align:center;">
        {{ matches.filter(m=>m.played).length }} / {{ matches.length }} 场已完成
      </div>
    </div>

    <!-- Team Score Banner (non-wheel modes) -->
    <div v-if="!isWheel" style="padding:12px 16px;background:#fff;margin:0 16px;border-radius:12px;box-shadow:0 2px 12px rgba(0,0,0,0.06);">
      <div style="display:flex;align-items:center;justify-content:center;gap:20px;">
        <div style="text-align:center;">
          <div style="font-size:12px;color:#969799;">男队</div>
          <div style="font-size:32px;font-weight:800;color:#1989fa;">{{ maleWins }}</div>
        </div>
        <div style="font-size:24px;font-weight:800;color:#c8c9cc;">:</div>
        <div style="text-align:center;">
          <div style="font-size:12px;color:#969799;">女队</div>
          <div style="font-size:32px;font-weight:800;color:#ee0a24;">{{ femaleWins }}</div>
        </div>
      </div>
      <div style="display:flex;justify-content:center;gap:16px;margin-top:6px;font-size:11px;color:#999;">
        <span>局数 {{ maleGameWins }}:{{ femaleGameWins }}</span>
        <span>分数 {{ malePoints }}:{{ femalePoints }}</span>
      </div>
    </div>

    <!-- Individual Standings (wheel_rr mode) -->
    <PlayerStandings v-if="isWheel" title="实时排名" :players="individualStandings" rating-label="开球网参考积分" />

    <!-- Match list -->
    <div style="font-size:16px;font-weight:600;padding:16px 16px 8px;display:flex;align-items:center;gap:6px;">
      <IconList :size="18" :stroke-width="2" style="vertical-align:-3px;" />
      对阵表
    </div>
    <div style="background:#fff;border-radius:12px;margin:4px 16px;box-shadow:0 2px 12px rgba(0,0,0,0.06);overflow:hidden;">
      <div v-for="(m, mi) in matches" :key="m.id"
        style="padding:10px 16px;border-bottom:1px solid #f5f5f5;">
        <!-- Top bar: card badges + draw button -->
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px;">
          <span style="font-size:12px;color:#c8c9cc;">#{{ mi+1 }}</span>
          <div style="display:flex;gap:4px;flex-wrap:wrap;justify-content:flex-end;">
            <template v-for="d in m.draws" :key="d.id">
              <span style="font-size:10px;font-weight:600;padding:1px 6px;border-radius:8px;color:#fff;position:relative;"
                :style="d.cancelled ? {background:'#ccc',color:'#999',textDecoration:'line-through'} : {background: drawColor(d)}">{{ drawLabel(d) }}</span>
            </template>
          </div>
          <button v-if="!m.played && canDraw(m)" @click="openCardDraw(m)"
            style="font-size:10px;padding:2px 8px;border-radius:8px;background:#fff3cd;color:#b8860b;border:1px solid #f5a623;cursor:pointer;margin-left:4px;flex-shrink:0;">
            {{ m.draws && m.draws.length > 0 ? '🔄 重抽' : '🎰 抽卡' }}
          </button>
        </div>
        <div @click="emit('openScoreEditor', m)" style="display:flex;align-items:center;cursor:pointer;">
          <div style="flex:1;text-align:right;font-weight:400;"
            :style="{fontWeight: m.winner_id === m.male_player_id ? 700 : 400}">
            {{ m.male_player_name }}<span v-if="isWheel" style="font-size:10px;color:#969799;display:block;">{{ playerRating(m.male_player_id) }}</span>
          </div>
          <div style="width:120px;text-align:center;font-weight:700;font-size:14px;">
            <template v-if="m.played">
              {{ matchScoreSummary(m) }}
              <span style="font-size:10px;display:block;"
                :style="{color: m.winner_team === 'male' ? '#1989fa' : '#ee0a24'}">{{ matchWinnerName(m) }} 胜</span>
            </template>
            <template v-else>
              <span style="color:#c8c9cc;font-size:13px;">待录入</span>
            </template>
          </div>
          <div style="flex:1;font-weight:400;"
            :style="{fontWeight: m.winner_id === m.female_player_id ? 700 : 400}">
            {{ m.female_player_name }}<span v-if="isWheel" style="font-size:10px;color:#969799;display:block;">{{ playerRating(m.female_player_id) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Cancel session (only when no matches played) -->
    <div v-if="!matches.some(m=>m.played)" style="padding:0 16px 8px;">
      <button @click="emit('cancelSession')" style="width:100%;padding:12px;background:#fff;color:#e74c3c;border:1px solid #e74c3c;border-radius:24px;font-size:14px;font-weight:500;cursor:pointer;">取消活动</button>
    </div>

    <!-- Action buttons -->
    <div style="padding:16px;display:flex;gap:12px;">
      <button @click="emit('completeSession')" :disabled="matches.some(m=>!m.played)"
        style="flex:1;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:16px;font-weight:600;cursor:pointer;"
        :style="{opacity:matches.some(m=>!m.played)?0.5:1}">结束活动</button>
      <button @click="emit('backToList')" style="padding:16px 24px;background:#f5f5f5;border:none;border-radius:24px;font-size:16px;cursor:pointer;color:#666;">返回</button>
    </div>

    <!-- Score Dialog -->
    <FunScoreDialog
      :show="showEditDialog"
      :male-name="scoringMatch?.male_player_name || ''"
      :female-name="scoringMatch?.female_player_name || ''"
      :handicap-points="scoringMatch?.handicap_points || 0"
      @update:show="emit('update:showEditDialog', $event)"
      @submit="(g1m:number,g1f:number,g2m:number,g2f:number,g3m?:number,g3f?:number) => emit('submitScore', g1m, g1f, g2m, g2f, g3m, g3f)"
    />

    <!-- Redraw confirmation -->
    <div v-if="confirmRedraw" style="position:fixed;inset:0;background:rgba(0,0,0,0.6);z-index:1050;display:flex;align-items:center;justify-content:center;" @click.self="confirmRedraw = false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:300px;text-align:center;">
        <div style="font-size:16px;font-weight:700;margin-bottom:8px;">🔄 重新抽卡</div>
        <div style="font-size:13px;color:#666;margin-bottom:4px;">
          {{ redrawMatch?.male_player_name }} vs {{ redrawMatch?.female_player_name }}
        </div>
        <div style="font-size:12px;color:#f5a623;margin-bottom:4px;">
          已抽过 {{ redrawMatch?.draws?.length || 0 }} 次：
          <span v-for="d in redrawMatch?.draws" :key="d.id"
            style="display:inline-block;margin:2px;padding:2px 6px;border-radius:6px;color:#fff;font-size:10px;"
            :style="d.cancelled ? {background:'#ccc',color:'#999',textDecoration:'line-through'} : {background: drawColor(d)}">{{ drawLabel(d) }}</span>
        </div>
        <div style="font-size:12px;color:#999;margin-bottom:16px;line-height:1.5;">
          让分会累加到已有分数上<br/>确认重新抽取？
        </div>
        <div style="display:flex;gap:12px;">
          <button @click="confirmRedraw = false"
            style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:24px;font-size:14px;cursor:pointer;">取消</button>
          <button @click="confirmRedraw = false; _startCardDraw(redrawMatch!)"
            style="flex:1;padding:12px;background:#f5a623;color:#fff;border:none;border-radius:24px;font-size:14px;font-weight:600;cursor:pointer;">确认重抽</button>
        </div>
      </div>
    </div>

    <!-- Card Draw Dialog -->
    <FunCardDraw
      :show="showCardDraw"
      :male-name="drawMaleName"
      :female-name="drawFemaleName"
      :male-rating="drawMaleRating"
      :female-rating="drawFemaleRating"
      :drawing="drawingCard"
      :result="drawResult"
      @draw="handleCardDraw"
      @close="handleCardDrawClose"
    />
  </div>
</template>
