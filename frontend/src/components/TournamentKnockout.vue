<script setup lang="ts">
import { ref } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import FunScoreDialog from './FunScoreDialog.vue'
import TournamentCardDraw from './TournamentCardDraw.vue'

interface Match {
  id: number; team_match_id: number; phase: string; match_order: number; match_type: string
  player_a_id: number; player_b_id: number; player_a_name: string; player_b_name: string
  player_a2_id: number | null; player_b2_id: number | null; player_a2_name: string; player_b2_name: string
  team_a_id: number; team_b_id: number
  game1_score_a: number | null; game1_score_b: number | null
  game2_score_a: number | null; game2_score_b: number | null
  game3_score_a: number | null; game3_score_b: number | null
  winner_id: number | null; winner_team_id: number | null; played: boolean; forfeit: boolean
}

interface TeamMatch {
  id: number; phase: string; round: number; team_a_name: string; team_b_name: string
  team_a_id: number; team_b_id: number; team_a_wins: number; team_b_wins: number
  winner_team_id: number | null; played: boolean; matches: Match[]
  cards: { id: number; team_id: number; card_type: string; drawn_at: string }[]
}

interface Detail {
  team_matches: TeamMatch[]
}

const props = defineProps<{ detail: Detail; tournamentId: number }>()
const emit = defineEmits<{ (e: 'refresh'): void; (e: 'back'): void }>()

const showScore = ref(false)
const scoringMatch = ref<Match | null>(null)
const showCardDraw = ref(false)
const drawingTM = ref<TeamMatch | null>(null)
const drawResult = ref<{ card_type: string; card_detail: string } | null>(null)
const expandedTM = ref<number | null>(null)

const matchTypeLabels: Record<number, string> = { 1: 'A单打', 2: 'BC双打', 3: 'C单打', 4: 'AB双打', 5: 'B单打' }

const sfMatches = computed(() => props.detail.team_matches.filter(m => m.phase === 'semifinal').sort((a, b) => a.round - b.round))
const finalMatch = computed(() => props.detail.team_matches.find(m => m.phase === 'final'))

function scoreStr(m: Match) {
  if (!m.played) return '-'
  if (m.forfeit) return '弃权'
  const g1 = `${m.game1_score_a}:${m.game1_score_b}`
  const g2 = `${m.game2_score_a}:${m.game2_score_b}`
  let s = `${g1} ${g2}`
  if (m.game3_score_a != null) s += ` ${m.game3_score_a}:${m.game3_score_b}`
  return s
}

function teamMatchScore(tm: TeamMatch) {
  return `${tm.team_a_wins} : ${tm.team_b_wins}`
}

function playerLabel(m: Match, side: 'a' | 'b') {
  if (side === 'a') return m.player_a_name + (m.player_a2_name ? '/' + m.player_a2_name : '')
  return m.player_b_name + (m.player_b2_name ? '/' + m.player_b2_name : '')
}

function openScoreEditor(m: Match) { scoringMatch.value = m; showScore.value = true }

async function handleScore(g1m: number, g1f: number, g2m: number, g2f: number, g3m?: number, g3f?: number) {
  if (!scoringMatch.value) return
  const body: any = {
    game1_score_a: g1m, game1_score_b: g1f,
    game2_score_a: g2m, game2_score_b: g2f,
  }
  if (g3m !== undefined && g3f !== undefined) { body.game3_score_a = g3m; body.game3_score_b = g3f }
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

function openCardDraw(tm: TeamMatch) { drawingTM.value = tm; drawResult.value = null; showCardDraw.value = true }

async function handleDrawCard() {
  if (!drawingTM.value) return
  const existingTeamIds = drawingTM.value.cards.map(c => c.team_id)
  const teamToDraw = existingTeamIds.includes(drawingTM.value.team_a_id) ? drawingTM.value.team_b_id : drawingTM.value.team_a_id
  try {
    const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${drawingTM.value.id}/draw-card`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ team_id: teamToDraw }),
    })
    if (!r.ok) { const t = await r.text(); showToast(t); return }
    drawResult.value = await r.json()
  } catch (e: any) { showToast('抽卡失败: ' + e.message) }
}

function cardLabel(t: string) { return t === 'edge_double' ? '擦边翻倍卡' : t === 'net_deduction' ? '擦网扣分卡' : t }

function toggleTM(id: number) { expandedTM.value = expandedTM.value === id ? null : id }

import { computed } from 'vue'
</script>

<template>
  <div style="padding: 16px;">
    <!-- Bracket visualization -->
    <div style="text-align: center; margin-bottom: 20px;">
      <div style="font-weight: 700; font-size: 16px; margin-bottom: 16px;">🏆 淘汰赛对阵</div>

      <!-- Semifinals row -->
      <div style="display: flex; justify-content: center; gap: 20px; margin-bottom: 16px;">
        <div v-for="sf in sfMatches" :key="sf.id"
          @click="toggleTM(sf.id)"
          style="flex: 1; max-width: 160px; background: #fff; border-radius: 12px; padding: 12px; box-shadow: var(--c-shadow); cursor: pointer;">
          <div style="font-size: 11px; font-weight: 600; color: #969799; margin-bottom: 6px;">半决赛 {{ sf.round }}</div>
          <div style="font-size: 13px; font-weight: 600;">{{ sf.team_a_name }}</div>
          <div style="font-size: 12px; color: #969799; margin: 4px 0;">vs</div>
          <div style="font-size: 13px; font-weight: 600;">{{ sf.team_b_name }}</div>
          <div v-if="sf.played && sf.winner_team_id" style="margin-top: 6px;">
            <span class="badge badge-success">{{ sf.winner_team_id === sf.team_a_id ? sf.team_a_name : sf.team_b_name }} 晋级</span>
          </div>
          <div v-else style="font-size: 12px; color: #ff976a; margin-top: 4px;">
            {{ sf.team_a_wins }}:{{ sf.team_b_wins }}
          </div>
        </div>
      </div>

      <!-- Arrow -->
      <div style="font-size: 24px; color: #ccc; margin-bottom: 12px;">↓</div>

      <!-- Final -->
      <div v-if="finalMatch" @click="toggleTM(finalMatch.id)"
        style="display: inline-block; min-width: 180px; background: linear-gradient(135deg, #fff9e6, #fff3cc); border-radius: 14px; padding: 14px; box-shadow: 0 4px 16px rgba(245,166,35,0.2); cursor: pointer; border: 2px solid #f5a623;">
        <div style="font-size: 13px; font-weight: 700; color: #f5a623; margin-bottom: 8px;">🏆 决赛</div>
        <div style="font-size: 14px; font-weight: 700;">{{ finalMatch.team_a_name || '待定' }}</div>
        <div style="font-size: 12px; color: #969799; margin: 4px 0;">vs</div>
        <div style="font-size: 14px; font-weight: 700;">{{ finalMatch.team_b_name || '待定' }}</div>
        <div v-if="finalMatch.played && finalMatch.winner_team_id" style="margin-top: 8px;">
          <span style="font-size: 18px; font-weight: 800; color: #f5a623;">👑 {{ finalMatch.winner_team_id === finalMatch.team_a_id ? finalMatch.team_a_name : finalMatch.team_b_name }}</span>
        </div>
      </div>
    </div>

    <!-- Expanded team match details -->
    <div v-for="tm in [...sfMatches, finalMatch].filter(Boolean)" :key="tm!.id">
      <div v-if="expandedTM === tm!.id" style="margin-top: 8px;">
        <!-- Card draw -->
        <div v-if="tm!.cards.length < 2 && !tm!.played" style="text-align: center; margin-bottom: 8px;">
          <button @click="openCardDraw(tm!)"
            style="padding: 8px 20px; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🎴 抽趣味卡</button>
        </div>
        <div v-if="tm!.cards.length > 0" style="display: flex; gap: 6px; justify-content: center; margin-bottom: 8px;">
          <span v-for="c in tm!.cards" :key="c.id" style="font-size: 11px; padding: 2px 8px; border-radius: 10px; background: #fff3e0; color: #e65100; font-weight: 600;">
            {{ c.team_id === tm!.team_a_id ? tm!.team_a_name : tm!.team_b_name }}: {{ cardLabel(c.card_type) }}
          </span>
        </div>
        <div style="text-align: center; font-size: 13px; font-weight: 600; margin-bottom: 6px;">
          {{ teamMatchScore(tm!) }}
        </div>

        <!-- 5 sub-matches -->
        <div v-for="m in tm!.matches" :key="m.id"
          style="display: flex; align-items: center; padding: 10px 14px; margin-bottom: 6px; background: #fff; border-radius: 10px; box-shadow: var(--c-shadow);"
          :style="{ opacity: m.played ? 1 : 0.8 }">
          <span style="width: 56px; font-size: 12px; font-weight: 600; color: #969799;">{{ matchTypeLabels[m.match_order] }}</span>
          <div style="flex: 1;">
            <div style="font-size: 13px; font-weight: 500;">{{ playerLabel(m, 'a') }} vs {{ playerLabel(m, 'b') }}</div>
            <div style="font-size: 11px; color: #969799;" v-if="m.played">{{ scoreStr(m) }}</div>
          </div>
          <button v-if="!m.played" @click="openScoreEditor(m)"
            style="padding: 6px 12px; background: #1989fa; color: #fff; border: none; border-radius: 14px; font-size: 12px; font-weight: 600; cursor: pointer;">录入</button>
          <span v-else style="font-size: 11px; font-weight: 600; color: #07c160;">✓</span>
        </div>
      </div>
    </div>
  </div>

  <FunScoreDialog
    :show="showScore"
    :male-name="scoringMatch?.player_a_name || ''"
    :female-name="scoringMatch?.player_b_name || ''"
    :handicap-points="0"
    @update:show="showScore = $event"
    @submit="handleScore"
    @forfeit="handleForfeit"
  />

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
