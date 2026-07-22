<script setup lang="ts">
import { ref } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import FunScoreDialog from './FunScoreDialog.vue'
import TournamentCardDraw from './TournamentCardDraw.vue'

// Reuse types from parent
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

interface TeamMatch {
  id: number; tournament_id: number; phase: string; round: number; group_name: string
  team_a_id: number; team_b_id: number; team_a_name: string; team_b_name: string
  team_a_wins: number; team_b_wins: number; winner_team_id: number | null
  played: boolean; matches: Match[]; cards: { id: number; team_match_id: number; team_id: number; card_type: string; drawn_at: string }[]
}

interface Team {
  id: number; group_name: string; team_name: string; group_wins: number; group_losses: number; group_rank: number | null
  players: { id: number; name: string; role: string; is_seed: boolean; reference_rating: number }[]
}

interface Detail {
  teams: Team[]; team_matches: TeamMatch[]
}

const props = defineProps<{ detail: Detail; tournamentId: number }>()
const emit = defineEmits<{ (e: 'refresh'): void; (e: 'back'): void }>()

const showScore = ref(false)
const scoringMatch = ref<Match | null>(null)
const scoringTeamMatch = ref<TeamMatch | null>(null)
const expandedTM = ref<number | null>(null)
const showCardDraw = ref(false)
const drawingTM = ref<TeamMatch | null>(null)
const drawResult = ref<{ card_type: string; card_detail: string } | null>(null)

const matchTypeLabels: Record<number, string> = { 1: 'A单打', 2: 'BC双打', 3: 'C单打', 4: 'AB双打', 5: 'B单打' }

function groupedTeams() {
  const groups: Record<string, Team[]> = {}
  for (const t of props.detail.teams) {
    if (!groups[t.group_name]) groups[t.group_name] = []
    groups[t.group_name].push(t)
  }
  // Sort each group by wins DESC
  for (const g of Object.values(groups)) {
    g.sort((a, b) => b.group_wins - a.group_wins || (a.group_rank || 99) - (b.group_rank || 99))
  }
  return groups
}

function groupedMatches() {
  const groups: Record<string, TeamMatch[]> = {}
  for (const tm of props.detail.team_matches.filter(m => m.phase === 'group')) {
    if (!groups[tm.group_name]) groups[tm.group_name] = []
    groups[tm.group_name].push(tm)
  }
  return groups
}

function scoreStr(m: Match) {
  if (!m.played) return '-'
  if (m.forfeit) return '弃权'
  const g1 = `${m.game1_score_a}:${m.game1_score_b}`
  const g2 = `${m.game2_score_a}:${m.game2_score_b}`
  let s = `${g1} ${g2}`
  if (m.game3_score_a != null) s += ` ${m.game3_score_a}:${m.game3_score_b}`
  return s
}

function matchWinnerLabel(m: Match) {
  if (!m.played) return ''
  if (m.winner_team_id === m.team_a_id) return m.player_a_name + (m.player_a2_name ? '/' + m.player_a2_name : '')
  return m.player_b_name + (m.player_b2_name ? '/' + m.player_b2_name : '')
}

function playerLabel(m: Match, side: 'a' | 'b') {
  if (side === 'a') return m.player_a_name + (m.player_a2_name ? '/' + m.player_a2_name : '')
  return m.player_b_name + (m.player_b2_name ? '/' + m.player_b2_name : '')
}

function teamMatchScore(tm: TeamMatch) {
  return `${tm.team_a_wins} : ${tm.team_b_wins}`
}

function openScoreEditor(tm: TeamMatch, m: Match) {
  scoringTeamMatch.value = tm
  scoringMatch.value = m
  showScore.value = true
}

async function handleScore(g1m: number, g1f: number, g2m: number, g2f: number, g3m?: number, g3f?: number) {
  if (!scoringMatch.value) return
  const body: any = {
    game1_score_a: g1m, game1_score_b: g1f,
    game2_score_a: g2m, game2_score_b: g2f,
  }
  if (g3m !== undefined && g3f !== undefined) {
    body.game3_score_a = g3m
    body.game3_score_b = g3f
  }
  try {
    const r = await fetch(`/api/tournaments/${props.tournamentId}/matches/${scoringMatch.value.id}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
    })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    showScore.value = false
    showSuccessToast('比分已保存')
    emit('refresh')
  } catch (e: any) { showToast('保存失败: ' + e.message) }
}

async function handleForfeit(winnerIsA: boolean) {
  if (!scoringMatch.value) return
  const winnerTeamId = winnerIsA ? scoringMatch.value.team_a_id : scoringMatch.value.team_b_id
  try {
    const r = await fetch(`/api/tournaments/${props.tournamentId}/matches/${scoringMatch.value.id}/forfeit`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ winner_team_id: winnerTeamId }),
    })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    showScore.value = false
    showSuccessToast('弃权已记录')
    emit('refresh')
  } catch (e: any) { showToast('失败: ' + e.message) }
}

function openCardDraw(tm: TeamMatch) {
  drawingTM.value = tm
  drawResult.value = null
  showCardDraw.value = true
}

async function handleDrawCard() {
  if (!drawingTM.value) return
  // Determine which team hasn't drawn yet
  const existingTeamIds = drawingTM.value.cards.map(c => c.team_id)
  const teamToDraw = existingTeamIds.includes(drawingTM.value.team_a_id) ? drawingTM.value.team_b_id : drawingTM.value.team_a_id
  if (!teamToDraw) { showToast('两队均已抽卡'); return }
  try {
    const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${drawingTM.value.id}/draw-card`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ team_id: teamToDraw }),
    })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    drawResult.value = await r.json()
  } catch (e: any) { showToast('抽卡失败: ' + e.message) }
}

function cardLabel(t: string) { return t === 'edge_double' ? '擦边翻倍卡' : t === 'net_deduction' ? '擦网扣分卡' : t }

function toggleTeamMatch(id: number) {
  expandedTM.value = expandedTM.value === id ? null : id
}
</script>

<template>
  <!-- Team standings per group -->
  <div v-for="(teams, groupName) in groupedTeams()" :key="groupName" style="margin: 12px 16px;">
    <div style="font-weight: 700; font-size: 15px; margin-bottom: 8px; color: #333;">{{ groupName }}组 积分榜</div>
    <div style="background: #fff; border-radius: 12px; overflow: hidden; box-shadow: var(--c-shadow);">
      <div style="display: flex; padding: 10px 14px; background: #f8f9fa; font-size: 12px; font-weight: 600; color: #969799;">
        <span style="flex: 3;">队伍</span>
        <span style="flex: 1; text-align: center;">胜</span>
        <span style="flex: 1; text-align: center;">负</span>
        <span style="flex: 1; text-align: center;">排名</span>
      </div>
      <div v-for="t in teams" :key="t.id"
        style="display: flex; padding: 12px 14px; align-items: center; border-top: 1px solid #f0f2f5;">
        <span style="flex: 3; font-weight: 600;">{{ t.team_name }}</span>
        <span style="flex: 1; text-align: center; color: #07c160; font-weight: 700;">{{ t.group_wins }}</span>
        <span style="flex: 1; text-align: center; color: #ee0a24; font-weight: 700;">{{ t.group_losses }}</span>
        <span style="flex: 1; text-align: center; font-weight: 700; color: #f5a623;">{{ t.group_rank || '-' }}</span>
      </div>
    </div>
  </div>

  <!-- Matches per group -->
  <div v-for="(matches, groupName) in groupedMatches()" :key="'m'+groupName" style="margin: 16px;">
    <div style="font-weight: 700; font-size: 15px; margin-bottom: 8px; color: #333;">{{ groupName }}组 对阵</div>

    <div v-for="tm in matches" :key="tm.id"
      style="background: #fff; border-radius: 12px; margin-bottom: 8px; overflow: hidden; box-shadow: var(--c-shadow);">
      <!-- Team match header -->
      <div @click="toggleTeamMatch(tm.id)"
        style="display: flex; align-items: center; justify-content: space-between; padding: 14px; cursor: pointer;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 18px; font-weight: 800; color: #1989fa;">{{ teamMatchScore(tm) }}</span>
          <span style="font-weight: 600;">{{ tm.team_a_name }} vs {{ tm.team_b_name }}</span>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
          <span v-if="tm.played" class="badge badge-success">已结束</span>
          <span v-else class="badge badge-warning">进行中</span>
          <span style="font-size: 12px; color: #969799;">{{ expandedTM === tm.id ? '▲' : '▼' }}</span>
        </div>
      </div>

      <!-- Cards display -->
      <div v-if="tm.cards.length > 0" style="padding: 0 14px 8px; display: flex; gap: 6px;">
        <span v-for="c in tm.cards" :key="c.id"
          style="font-size: 11px; padding: 2px 8px; border-radius: 10px; background: #fff3e0; color: #e65100; font-weight: 600;">
          {{ c.team_id === tm.team_a_id ? tm.team_a_name : tm.team_b_name }}: {{ cardLabel(c.card_type) }}
        </span>
      </div>

      <!-- Sub-matches (expandable) -->
      <div v-if="expandedTM === tm.id" style="border-top: 1px solid #f0f2f5;">
        <!-- Card draw row -->
        <div v-if="tm.cards.length < 2 && !tm.played" style="padding: 10px 14px; text-align: center;">
          <button @click="openCardDraw(tm)"
            style="padding: 8px 20px; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🎴 抽趣味卡</button>
        </div>

        <!-- 5 matches -->
        <div v-for="m in tm.matches" :key="m.id"
          style="display: flex; align-items: center; padding: 10px 14px; border-top: 1px solid #f5f5f5;"
          :style="{ background: m.played ? '#fff' : '#fafafa' }">
          <span style="width: 56px; font-size: 12px; font-weight: 600; color: #969799;">{{ matchTypeLabels[m.match_order] }}</span>
          <div style="flex: 1;">
            <div style="font-size: 13px; font-weight: 500;">
              <span :style="{ color: m.winner_team_id === m.team_a_id ? '#1989fa' : '#333' }">{{ playerLabel(m, 'a') }}</span>
              <span style="margin: 0 4px; color: #ccc;">vs</span>
              <span :style="{ color: m.winner_team_id === m.team_b_id ? '#ee0a24' : '#333' }">{{ playerLabel(m, 'b') }}</span>
            </div>
            <div style="font-size: 11px; color: #969799;" v-if="m.played">{{ scoreStr(m) }} <span v-if="m.forfeit" style="color: #ff976a;">(弃权)</span></div>
          </div>
          <button v-if="!m.played" @click="openScoreEditor(tm, m)"
            style="padding: 6px 12px; background: #1989fa; color: #fff; border: none; border-radius: 14px; font-size: 12px; font-weight: 600; cursor: pointer;">录入</button>
          <span v-else style="font-size: 11px; font-weight: 600; color: #07c160;">{{ matchWinnerLabel(m) }} 胜</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Score dialog -->
  <FunScoreDialog
    :show="showScore"
    :male-name="scoringMatch?.player_a_name || ''"
    :female-name="scoringMatch?.player_b_name || ''"
    :handicap-points="0"
    @update:show="showScore = $event"
    @submit="handleScore"
    @forfeit="handleForfeit"
  />

  <!-- Card draw overlay -->
  <TournamentCardDraw
    :show="showCardDraw"
    :drawing="drawResult === null"
    :result="drawResult"
    :team-a-name="drawingTM?.team_a_name || ''"
    :team-b-name="drawingTM?.team_b_name || ''"
    @draw="handleDrawCard"
    @close="showCardDraw = false; emit('refresh')"
  />
</template>
