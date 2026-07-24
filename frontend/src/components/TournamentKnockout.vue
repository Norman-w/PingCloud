<script setup lang="ts">
//#region 导入/依赖
import { ref, computed } from 'vue'
import { showToast, showSuccessToast, showDialog } from 'vant'
import FunScoreDialog from './FunScoreDialog.vue'
import TournamentCardDraw from './TournamentCardDraw.vue'
//#endregion

//#region 模型/类型
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

interface Team {
	id: number; team_name: string
	players: { id: number; name: string; role: string; is_seed: boolean }[]
}

interface Detail {
	teams: Team[]
	team_matches: TeamMatch[]
}
//#endregion

//#region 私有成员
const props = defineProps<{ detail: Detail; tournamentId: number; readonly?: boolean }>()
const emit = defineEmits<{ (e: 'refresh'): void; (e: 'back'): void }>()

const showScore = ref(false)
const scoringMatch = ref<Match | null>(null)
const showCardDraw = ref(false)
const drawingTM = ref<TeamMatch | null>(null)
const drawResult = ref<{ card_type: string; card_detail: string } | null>(null)
const drawFailTick = ref(0)
const expandedTM = ref<number | null>(null)

const drawingTeamName = computed(() => {
	const tm = drawingTM.value
	if (!tm) return ''
	const drawn = tm.cards.map(c => c.team_id)
	if (!drawn.includes(tm.team_a_id)) return tm.team_a_name
	if (!drawn.includes(tm.team_b_id)) return tm.team_b_name
	return ''
})
const lineupTM = ref<TeamMatch | null>(null)
const lineupA = ref<{ A: number; B: number; C: number }>({ A: 0, B: 0, C: 0 })
const lineupB = ref<{ A: number; B: number; C: number }>({ A: 0, B: 0, C: 0 })

const matchTypeLabels: Record<number, string> = { 1: 'A单打', 2: 'BC双打', 3: 'C单打', 4: 'AB双打', 5: 'B单打' }
const roles = ['A', 'B', 'C'] as const
//#endregion

//#region 公开 API / 业务逻辑
const sfMatches = computed(() => props.detail.team_matches.filter(m => m.phase === 'semifinal').sort((a, b) => a.round - b.round))
const finalMatch = computed(() => props.detail.team_matches.find(m => m.phase === 'final') || null)
const thirdMatch = computed(() => props.detail.team_matches.find(m => m.phase === 'third_place') || null)
const detailMatches = computed(() => [ ...sfMatches.value, finalMatch.value, thirdMatch.value ].filter(Boolean) as TeamMatch[])

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

function phaseTitle(phase: string) {
	if (phase === 'semifinal') return '半决赛'
	if (phase === 'final') return '决赛 · 决1-2名'
	if (phase === 'third_place') return '三四名决赛'
	return phase
}

function readyToComplete(tm: TeamMatch) {
	return !tm.played && (tm.team_a_wins >= 3 || tm.team_b_wins >= 3)
}

function teamById(id: number) {
	return props.detail.teams.find(t => t.id === id)
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
		if (!r.ok) { showToast(await r.text()); return }
		showScore.value = false
		showSuccessToast('比分已保存')
		emit('refresh')
	} catch (e: any) { showToast('保存失败: ' + e.message) }
}

async function clearMatchScore(tm: TeamMatch, m: Match) {
	try {
		await showDialog({
			title: '清除本场比分',
			message: `确认清除「${matchTypeLabels[m.match_order]}」比分？清除后变为未录入。`,
			showCancelButton: true,
		})
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/matches/${m.id}/clear`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('已清除，本场视为未打')
		emit('refresh')
	} catch (e: any) { showToast('清除失败: ' + e.message) }
}

function openCardDraw(tm: TeamMatch) { drawingTM.value = tm; drawResult.value = null; showCardDraw.value = true }

async function handleDrawCard() {
	if (!drawingTM.value) return
	const existingTeamIds = drawingTM.value.cards.map(c => c.team_id)
	let teamToDraw: number | null = null
	if (!existingTeamIds.includes(drawingTM.value.team_a_id)) teamToDraw = drawingTM.value.team_a_id
	else if (!existingTeamIds.includes(drawingTM.value.team_b_id)) teamToDraw = drawingTM.value.team_b_id
	if (!teamToDraw) {
		showToast('两队均已抽卡')
		drawFailTick.value++
		return
	}
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${drawingTM.value.id}/draw-card`, {
			method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ team_id: teamToDraw }),
		})
		if (!r.ok) {
			showToast((await r.text()) || '抽卡失败')
			drawFailTick.value++
			return
		}
		drawResult.value = await r.json()
	} catch (e: any) {
		showToast('抽卡失败: ' + e.message)
		drawFailTick.value++
	}
}

function cardLabel(t: string) { return t === 'edge_double' ? '擦边翻倍卡' : t === 'net_deduction' ? '擦网扣分卡' : t }

function toggleTM(id: number) { expandedTM.value = expandedTM.value === id ? null : id }

function openLineup(tm: TeamMatch) {
	lineupTM.value = tm
	const ta = teamById(tm.team_a_id)
	const tb = teamById(tm.team_b_id)
	const m1 = tm.matches.find(m => m.match_order === 1)
	const m2 = tm.matches.find(m => m.match_order === 2)
	lineupA.value = {
		A: m1?.player_a_id || ta?.players.find(p => p.role === 'A')?.id || ta?.players[0]?.id || 0,
		B: m2?.player_a_id || ta?.players.find(p => p.role === 'B')?.id || ta?.players[1]?.id || 0,
		C: m2?.player_a2_id || ta?.players.find(p => p.role === 'C')?.id || ta?.players[2]?.id || 0,
	}
	lineupB.value = {
		A: m1?.player_b_id || tb?.players.find(p => p.role === 'A')?.id || tb?.players[0]?.id || 0,
		B: m2?.player_b_id || tb?.players.find(p => p.role === 'B')?.id || tb?.players[1]?.id || 0,
		C: m2?.player_b2_id || tb?.players.find(p => p.role === 'C')?.id || tb?.players[2]?.id || 0,
	}
}

async function saveLineup() {
	if (!lineupTM.value) return
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${lineupTM.value.id}/lineup`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ team_a: lineupA.value, team_b: lineupB.value }),
		})
		if (!r.ok) { showToast(await r.text()); return }
		lineupTM.value = null
		showSuccessToast('出场已保存')
		emit('refresh')
	} catch (e: any) { showToast('保存失败: ' + e.message) }
}

async function completeTeamMatch(tm: TeamMatch) {
	try {
		await showDialog({ title: '提交结束', message: `确认结束 ${tm.team_a_name} vs ${tm.team_b_name}？`, showCancelButton: true })
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${tm.id}/complete`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('本场已结束')
		emit('refresh')
	} catch (e: any) { showToast('提交失败: ' + e.message) }
}

async function reopenTeamMatch(tm: TeamMatch) {
	const tip = tm.phase === 'semifinal'
		? '重开半决赛将作废已生成的决赛/三四名对阵，确认继续？'
		: '打开后可改比分/出场。'
	try {
		await showDialog({ title: '重新打开', message: tip, showCancelButton: true })
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${tm.id}/reopen`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('已重新打开')
		emit('refresh')
	} catch (e: any) { showToast('失败: ' + e.message) }
}

async function completeTournament() {
	try {
		await showDialog({ title: '结束赛事', message: '决赛与三四名均已打完，确认结束本届混合团体赛？', showCancelButton: true })
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/complete`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('赛事已结束')
		emit('refresh')
	} catch (e: any) { showToast('失败: ' + e.message) }
}

const canCompleteTournament = computed(() => {
	const f = finalMatch.value
	const t = thirdMatch.value
	return !!(f?.played && t?.played)
})
//#endregion
</script>

<template>
	<!--#region 视图层 -->
	<div style="padding: 16px;">
		<div style="text-align: center; margin-bottom: 20px;">
			<div style="font-weight: 700; font-size: 16px; margin-bottom: 16px;">🏆 半决赛 / 决赛</div>

			<!-- Semifinals -->
			<div style="display: flex; justify-content: center; gap: 12px; margin-bottom: 16px;">
				<div v-for="sf in sfMatches" :key="sf.id"
					@click="toggleTM(sf.id)"
					style="flex: 1; max-width: 160px; background: #fff; border-radius: 12px; padding: 12px; box-shadow: var(--c-shadow); cursor: pointer;"
					:style="{ border: expandedTM === sf.id ? '2px solid #1989fa' : '2px solid transparent' }">
					<div style="font-size: 11px; font-weight: 600; color: #969799; margin-bottom: 6px;">半决赛 {{ sf.round }}</div>
					<div style="font-size: 13px; font-weight: 600;">{{ sf.team_a_name }}</div>
					<div style="font-size: 12px; color: #969799; margin: 4px 0;">vs</div>
					<div style="font-size: 13px; font-weight: 600;">{{ sf.team_b_name }}</div>
					<div v-if="sf.played && sf.winner_team_id" style="margin-top: 6px;">
						<span class="badge badge-success">{{ sf.winner_team_id === sf.team_a_id ? sf.team_a_name : sf.team_b_name }} 进决赛</span>
					</div>
					<div v-else-if="readyToComplete(sf)" style="font-size: 12px; color: #07c160; margin-top: 4px;">待提交 {{ sf.team_a_wins }}:{{ sf.team_b_wins }}</div>
					<div v-else style="font-size: 12px; color: #ff976a; margin-top: 4px;">{{ sf.team_a_wins }}:{{ sf.team_b_wins }}</div>
				</div>
			</div>

			<div style="font-size: 24px; color: #ccc; margin-bottom: 12px;">↓</div>

			<!-- Final + Third place side by side -->
			<div style="display: flex; justify-content: center; gap: 12px; flex-wrap: wrap;">
				<div v-if="finalMatch" @click="toggleTM(finalMatch.id)"
					style="flex: 1; min-width: 140px; max-width: 200px; background: linear-gradient(135deg, #fff9e6, #fff3cc); border-radius: 14px; padding: 14px; box-shadow: 0 4px 16px rgba(245,166,35,0.2); cursor: pointer;"
					:style="{ border: expandedTM === finalMatch.id ? '2px solid #f5a623' : '2px solid #f5a623' }">
					<div style="font-size: 13px; font-weight: 700; color: #f5a623; margin-bottom: 8px;">🏆 决赛 · 1-2名</div>
					<div style="font-size: 14px; font-weight: 700;">{{ finalMatch.team_a_name || '待定' }}</div>
					<div style="font-size: 12px; color: #969799; margin: 4px 0;">vs</div>
					<div style="font-size: 14px; font-weight: 700;">{{ finalMatch.team_b_name || '待定' }}</div>
					<div v-if="finalMatch.played && finalMatch.winner_team_id" style="margin-top: 8px;">
						<span style="font-size: 16px; font-weight: 800; color: #f5a623;">👑 {{ finalMatch.winner_team_id === finalMatch.team_a_id ? finalMatch.team_a_name : finalMatch.team_b_name }}</span>
					</div>
					<div v-else-if="readyToComplete(finalMatch)" style="font-size: 12px; color: #07c160; margin-top: 4px;">待提交 {{ finalMatch.team_a_wins }}:{{ finalMatch.team_b_wins }}</div>
					<div v-else style="font-size: 12px; color: #ff976a; margin-top: 4px;">{{ finalMatch.team_a_wins }}:{{ finalMatch.team_b_wins }}</div>
				</div>
				<div v-else style="flex: 1; min-width: 140px; max-width: 200px; background: #f5f5f5; border-radius: 14px; padding: 14px; color: #969799; font-size: 13px;">
					🏆 决赛 · 1-2名<br><span style="font-size: 12px;">半决赛结束后生成</span>
				</div>

				<div v-if="thirdMatch" @click="toggleTM(thirdMatch.id)"
					style="flex: 1; min-width: 140px; max-width: 200px; background: #f0f7ff; border-radius: 14px; padding: 14px; box-shadow: var(--c-shadow); cursor: pointer;"
					:style="{ border: expandedTM === thirdMatch.id ? '2px solid #1989fa' : '2px solid #c8e0ff' }">
					<div style="font-size: 13px; font-weight: 700; color: #1989fa; margin-bottom: 8px;">🥉 三四名决赛</div>
					<div style="font-size: 14px; font-weight: 700;">{{ thirdMatch.team_a_name || '待定' }}</div>
					<div style="font-size: 12px; color: #969799; margin: 4px 0;">vs</div>
					<div style="font-size: 14px; font-weight: 700;">{{ thirdMatch.team_b_name || '待定' }}</div>
					<div v-if="thirdMatch.played && thirdMatch.winner_team_id" style="margin-top: 8px;">
						<span style="font-size: 14px; font-weight: 700; color: #1989fa;">{{ thirdMatch.winner_team_id === thirdMatch.team_a_id ? thirdMatch.team_a_name : thirdMatch.team_b_name }} 获第3</span>
					</div>
					<div v-else-if="readyToComplete(thirdMatch)" style="font-size: 12px; color: #07c160; margin-top: 4px;">待提交 {{ thirdMatch.team_a_wins }}:{{ thirdMatch.team_b_wins }}</div>
					<div v-else style="font-size: 12px; color: #ff976a; margin-top: 4px;">{{ thirdMatch.team_a_wins }}:{{ thirdMatch.team_b_wins }}</div>
				</div>
				<div v-else style="flex: 1; min-width: 140px; max-width: 200px; background: #f5f5f5; border-radius: 14px; padding: 14px; color: #969799; font-size: 13px;">
					🥉 三四名决赛<br><span style="font-size: 12px;">半决赛结束后生成</span>
				</div>
			</div>
		</div>

		<!-- Expanded match details -->
		<div v-for="tm in detailMatches" :key="tm.id">
			<div v-if="expandedTM === tm.id" style="margin-top: 8px; background: #fff; border-radius: 12px; padding: 12px; box-shadow: var(--c-shadow);">
				<div style="font-weight: 700; font-size: 15px; margin-bottom: 10px; text-align: center;">
					{{ phaseTitle(tm.phase) }} · {{ tm.team_a_name }} vs {{ tm.team_b_name }}
				</div>
				<div style="text-align: center; font-size: 18px; font-weight: 800; color: #1989fa; margin-bottom: 10px;">
					{{ teamMatchScore(tm) }}
				</div>

				<div v-if="!readonly" style="display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; margin-bottom: 10px;">
					<button v-if="!tm.played" @click="openLineup(tm)"
						style="padding: 8px 14px; background: #fff; border: 1.5px solid #1989fa; color: #1989fa; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">👥 配置出场 ABC</button>
					<button v-if="tm.cards.length < 2 && !tm.played" @click="openCardDraw(tm)"
						style="padding: 8px 14px; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🎴 抽趣味卡</button>
					<button v-if="readyToComplete(tm)" @click="completeTeamMatch(tm)"
						style="padding: 8px 14px; background: linear-gradient(135deg, #07c160, #06ad56); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">✅ 提交结束本场</button>
					<button v-if="tm.played" @click="reopenTeamMatch(tm)"
						style="padding: 8px 14px; background: #fff; border: 1.5px solid #ed6a0c; color: #ed6a0c; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🔓 重新打开改分</button>
				</div>
				<div v-if="!readonly && readyToComplete(tm)" style="text-align: center; font-size: 12px; color: #07c160; margin-bottom: 8px;">
					已达 3 胜，核对比分后提交结束
				</div>

				<div v-if="tm.cards.length > 0" style="display: flex; gap: 6px; justify-content: center; margin-bottom: 8px;">
					<span v-for="c in tm.cards" :key="c.id" style="font-size: 11px; padding: 2px 8px; border-radius: 10px; background: #fff3e0; color: #e65100; font-weight: 600;">
						{{ c.team_id === tm.team_a_id ? tm.team_a_name : tm.team_b_name }}: {{ cardLabel(c.card_type) }}
					</span>
				</div>

				<div v-for="m in tm.matches" :key="m.id"
					style="display: flex; align-items: center; padding: 10px 12px; margin-bottom: 6px; background: #fafafa; border-radius: 10px;">
					<span style="width: 56px; font-size: 12px; font-weight: 600; color: #969799;">{{ matchTypeLabels[m.match_order] }}</span>
					<div style="flex: 1;">
						<div style="font-size: 13px; font-weight: 500;">{{ playerLabel(m, 'a') }} vs {{ playerLabel(m, 'b') }}</div>
						<div style="font-size: 11px; color: #969799;" v-if="m.played">{{ scoreStr(m) }}</div>
					</div>
					<div style="display: flex; gap: 6px; flex-shrink: 0;">
						<button v-if="!readonly && !tm.played && m.played" @click="clearMatchScore(tm, m)"
							style="padding: 6px 10px; background: #fff; color: #ee0a24; border: 1.5px solid #ee0a24; border-radius: 14px; font-size: 12px; font-weight: 600; cursor: pointer;">
							清除
						</button>
						<button v-if="!readonly && !tm.played" @click="openScoreEditor(m)"
							style="padding: 6px 12px; background: #1989fa; color: #fff; border: none; border-radius: 14px; font-size: 12px; font-weight: 600; cursor: pointer;">
							{{ m.played ? '改分' : '录入' }}
						</button>
						<span v-else-if="m.played" style="font-size: 11px; font-weight: 600; color: #07c160;">✓</span>
					</div>
				</div>
			</div>
		</div>

		<div v-if="!readonly && canCompleteTournament" style="margin-top: 16px;">
			<button @click="completeTournament"
				style="width: 100%; padding: 16px; border: none; border-radius: 14px; font-size: 16px; font-weight: 700; cursor: pointer; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff;">
				🏁 结束混合团体赛
			</button>
		</div>
	</div>

	<!-- Lineup editor -->
	<div v-if="lineupTM" style="position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 600; display: flex; align-items: center; justify-content: center;" @click.self="lineupTM = null">
		<div style="background: #fff; border-radius: 16px; padding: 20px; width: 92%; max-width: 400px; max-height: 85vh; overflow: auto;">
			<div style="font-weight: 700; font-size: 16px; margin-bottom: 4px;">配置出场 ABC</div>
			<div style="font-size: 12px; color: #969799; margin-bottom: 14px;">{{ lineupTM.team_a_name }} vs {{ lineupTM.team_b_name }} · {{ phaseTitle(lineupTM.phase) }}</div>

			<div style="margin-bottom: 14px;">
				<div style="font-weight: 700; color: #1989fa; margin-bottom: 8px;">{{ lineupTM.team_a_name }}</div>
				<div v-for="role in roles" :key="'a'+role" style="display: flex; align-items: center; gap: 10px; margin-bottom: 8px;">
					<span style="width: 24px; font-weight: 700;">{{ role }}</span>
					<select v-model.number="lineupA[role]" style="flex: 1; padding: 8px; border: 1.5px solid #ebedf0; border-radius: 8px;">
						<option v-for="p in (teamById(lineupTM.team_a_id)?.players || [])" :key="p.id" :value="p.id">{{ p.name }}</option>
					</select>
				</div>
			</div>
			<div style="margin-bottom: 14px;">
				<div style="font-weight: 700; color: #ee0a24; margin-bottom: 8px;">{{ lineupTM.team_b_name }}</div>
				<div v-for="role in roles" :key="'b'+role" style="display: flex; align-items: center; gap: 10px; margin-bottom: 8px;">
					<span style="width: 24px; font-weight: 700;">{{ role }}</span>
					<select v-model.number="lineupB[role]" style="flex: 1; padding: 8px; border: 1.5px solid #ebedf0; border-radius: 8px;">
						<option v-for="p in (teamById(lineupTM.team_b_id)?.players || [])" :key="p.id" :value="p.id">{{ p.name }}</option>
					</select>
				</div>
			</div>
			<div style="display: flex; gap: 10px;">
				<button @click="lineupTM = null" style="flex: 1; padding: 12px; border: 1.5px solid #ccc; border-radius: 12px; background: #fff; cursor: pointer;">取消</button>
				<button @click="saveLineup" style="flex: 2; padding: 12px; border: none; border-radius: 12px; background: #1989fa; color: #fff; font-weight: 700; cursor: pointer;">保存出场</button>
			</div>
		</div>
	</div>

	<FunScoreDialog
		:show="showScore"
		:male-name="scoringMatch?.player_a_name || ''"
		:female-name="scoringMatch?.player_b_name || ''"
		:handicap-points="0"
		:allow-forfeit="false"
		@update:show="showScore = $event"
		@submit="handleScore"
	/>

	<TournamentCardDraw
		:show="showCardDraw"
		:drawing="drawResult === null"
		:result="drawResult"
		:fail-tick="drawFailTick"
		:team-a-name="drawingTM?.team_a_name || ''"
		:team-b-name="drawingTM?.team_b_name || ''"
		:drawing-team-name="drawingTeamName"
		@draw="handleDrawCard"
		@close="showCardDraw = false; emit('refresh')"
	/>
	<!--#endregion -->
</template>
