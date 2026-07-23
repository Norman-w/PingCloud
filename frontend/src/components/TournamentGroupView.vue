<script setup lang="ts">
//#region 导入/依赖
import { ref, computed } from 'vue'
import { showToast, showSuccessToast, showDialog } from 'vant'
import FunScoreDialog from './FunScoreDialog.vue'
import TournamentCardDraw from './TournamentCardDraw.vue'
//#endregion

//#region 模型/类型
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
	id: number; group_name: string; team_name: string
	group_wins: number; group_losses: number; group_points: number
	games_won: number; games_lost: number; points_scored: number
	group_rank: number | null; rank_manual: boolean
	players: { id: number; name: string; role: string; is_seed: boolean; reference_rating: number }[]
}

interface Detail {
	teams: Team[]; team_matches: TeamMatch[]
}
//#endregion

//#region 私有成员
const props = defineProps<{ detail: Detail; tournamentId: number; readonly?: boolean }>()
const emit = defineEmits<{ (e: 'refresh'): void; (e: 'back'): void }>()

const showScore = ref(false)
const scoringMatch = ref<Match | null>(null)
const scoringTeamMatch = ref<TeamMatch | null>(null)
const expandedTM = ref<number | null>(null)
const showCardDraw = ref(false)
const drawingTM = ref<TeamMatch | null>(null)
const drawResult = ref<{ card_type: string; card_detail: string } | null>(null)
const drawFailTick = ref(0)
const expandedTeams = ref<Set<number>>(new Set())
const manualRankGroup = ref<string | null>(null)
const manualRanks = ref<Record<number, number>>({})
const lineupTM = ref<TeamMatch | null>(null)
const lineupA = ref<{ A: number; B: number; C: number }>({ A: 0, B: 0, C: 0 })
const lineupB = ref<{ A: number; B: number; C: number }>({ A: 0, B: 0, C: 0 })

const matchTypeLabels: Record<number, string> = { 1: 'A单打', 2: 'BC双打', 3: 'C单打', 4: 'AB双打', 5: 'B单打' }
const roles = ['A', 'B', 'C'] as const
//#endregion

//#region 公开 API / 业务逻辑
const groupedTeams = computed(() => {
	const groups: Record<string, Team[]> = {}
	for (const t of props.detail.teams) {
		if (!groups[t.group_name]) groups[t.group_name] = []
		groups[t.group_name].push(t)
	}
	for (const g of Object.values(groups)) {
		g.sort((a, b) => {
			const ra = a.group_rank || 99
			const rb = b.group_rank || 99
			if (ra !== rb) return ra - rb
			return (b.group_points || 0) - (a.group_points || 0)
				|| (b.games_won - b.games_lost) - (a.games_won - a.games_lost)
				|| (b.points_scored || 0) - (a.points_scored || 0)
		})
	}
	return groups
})

function groupedMatches() {
	const groups: Record<string, TeamMatch[]> = {}
	for (const tm of props.detail.team_matches.filter(m => m.phase === 'group')) {
		if (!groups[tm.group_name]) groups[tm.group_name] = []
		groups[tm.group_name].push(tm)
	}
	return groups
}

function groupNeedsManualRank(groupName: string): boolean {
	const teams = groupedTeams.value[groupName] || []
	const played = (props.detail.team_matches || []).some(
		m => m.phase === 'group' && m.group_name === groupName && m.played,
	)
	if (!played) return false
	return teams.some(t => !t.group_rank)
}

/** 并列队互战小分（不含场外队伍，如不含 A1） */
function miniTableForGroup(groupName: string) {
	const teams = (groupedTeams.value[groupName] || []).filter(t => !t.group_rank)
	if (teams.length < 2) return null
	const pts = teams[0].group_points
	if (!teams.every(t => t.group_points === pts)) return null
	const tiedIds = new Set(teams.map(t => t.id))
	const stats: Record<number, { pts: number; fw: number; fl: number; gw: number; gl: number; ps: number }> = {}
	for (const t of teams) stats[t.id] = { pts: 0, fw: 0, fl: 0, gw: 0, gl: 0, ps: 0 }

	for (const tm of props.detail.team_matches) {
		if (tm.phase !== 'group' || tm.group_name !== groupName || !tm.played) continue
		if (!tiedIds.has(tm.team_a_id) || !tiedIds.has(tm.team_b_id)) continue
		const w = tm.winner_team_id
		if (!w) continue
		stats[w].pts += 2
		const loser = w === tm.team_a_id ? tm.team_b_id : tm.team_a_id
		stats[loser].pts += 1
		stats[tm.team_a_id].fw += tm.team_a_wins
		stats[tm.team_a_id].fl += tm.team_b_wins
		stats[tm.team_b_id].fw += tm.team_b_wins
		stats[tm.team_b_id].fl += tm.team_a_wins
		for (const m of tm.matches || []) {
			if (!m.played || m.forfeit) continue
			const games: [number, number][] = [
				[m.game1_score_a || 0, m.game1_score_b || 0],
				[m.game2_score_a || 0, m.game2_score_b || 0],
			]
			if (m.game3_score_a != null) games.push([m.game3_score_a, m.game3_score_b || 0])
			for (const [sa, sb] of games) {
				stats[tm.team_a_id].ps += sa
				stats[tm.team_b_id].ps += sb
				if (sa > sb) { stats[tm.team_a_id].gw++; stats[tm.team_b_id].gl++ }
				else if (sb > sa) { stats[tm.team_b_id].gw++; stats[tm.team_a_id].gl++ }
			}
		}
	}
	return teams.map(t => ({
		team: t,
		...stats[t.id],
	}))
}

function teamById(id: number) {
	return props.detail.teams.find(t => t.id === id)
}

function readyToComplete(tm: TeamMatch) {
	return !tm.played && (tm.team_a_wins >= 3 || tm.team_b_wins >= 3)
}

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
		await showDialog({ title: '提交结束', message: `确认结束 ${tm.team_a_name} vs ${tm.team_b_name}？结束后将计入积分榜，改分需先重新打开。` })
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${tm.id}/complete`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('本场已结束')
		emit('refresh')
	} catch (e: any) { showToast('提交失败: ' + e.message) }
}

async function reopenTeamMatch(tm: TeamMatch) {
	try {
		await showDialog({ title: '重新打开', message: '打开后可改比分/出场，积分榜将暂时撤回本场结果。' })
	} catch { return }
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/team-matches/${tm.id}/reopen`, { method: 'POST' })
		if (!r.ok) { showToast(await r.text()); return }
		showSuccessToast('已重新打开')
		emit('refresh')
	} catch (e: any) { showToast('失败: ' + e.message) }
}

function toggleTeamExpand(id: number) {
	const next = new Set(expandedTeams.value)
	if (next.has(id)) next.delete(id)
	else next.add(id)
	expandedTeams.value = next
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
			const t = await r.text()
			showToast(t || '抽卡失败')
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

function toggleTeamMatch(id: number) {
	expandedTM.value = expandedTM.value === id ? null : id
}

function openManualRank(groupName: string) {
	const teams = groupedTeams.value[groupName] || []
	manualRankGroup.value = groupName
	const ranks: Record<number, number> = {}
	teams.forEach((t, i) => { ranks[t.id] = t.group_rank || i + 1 })
	manualRanks.value = ranks
}

async function saveManualRanks() {
	if (!manualRankGroup.value) return
	const ranks = Object.entries(manualRanks.value).map(([teamId, rank]) => ({
		team_id: Number(teamId), rank: Number(rank),
	}))
	const vals = ranks.map(r => r.rank)
	if (new Set(vals).size !== vals.length) {
		showToast('名次不能重复')
		return
	}
	try {
		const r = await fetch(`/api/tournaments/${props.tournamentId}/set-ranks`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ group_name: manualRankGroup.value, ranks }),
		})
		if (!r.ok) { const t = await r.text(); showToast(t); return }
		manualRankGroup.value = null
		showSuccessToast('名次已保存')
		emit('refresh')
	} catch (e: any) { showToast('保存失败: ' + e.message) }
}

async function confirmManualHint(groupName: string) {
	try {
		await showDialog({
			title: '无法自动判定名次',
			message: `${groupName}组存在积分/局分完全相同的循环局面，请手动指定名次。`,
			confirmButtonText: '去设置',
		})
		openManualRank(groupName)
	} catch { /* cancelled */ }
}
//#endregion
</script>

<template>
	<!--#region 视图层 -->
	<!-- Team standings per group -->
	<div v-for="(teams, groupName) in groupedTeams" :key="groupName" style="margin: 12px 16px;">
		<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
			<div style="font-weight: 700; font-size: 15px; color: #333;">{{ groupName }}组 积分榜</div>
			<button v-if="groupNeedsManualRank(String(groupName)) && !readonly" @click="confirmManualHint(String(groupName))"
				style="padding: 4px 10px; border: 1px solid #ed6a0c; background: #fff7e8; color: #ed6a0c; border-radius: 12px; font-size: 11px; font-weight: 600; cursor: pointer;">
				需手动排名
			</button>
		</div>
		<div style="background: #fff; border-radius: 12px; overflow: hidden; box-shadow: var(--c-shadow);">
			<div style="display: flex; padding: 10px 10px; background: #f8f9fa; font-size: 11px; font-weight: 600; color: #969799;">
				<span style="flex: 2.2;">队伍</span>
				<span style="flex: 1; text-align: center;">积分</span>
				<span style="flex: 1.2; text-align: center;">局比</span>
				<span style="flex: 1; text-align: center;">总分</span>
				<span style="flex: 0.8; text-align: center;">排名</span>
			</div>
			<div v-for="t in teams" :key="t.id" style="border-top: 1px solid #f0f2f5;">
				<div @click="toggleTeamExpand(t.id)"
					style="display: flex; padding: 12px 10px; align-items: center; cursor: pointer;">
					<span style="flex: 2.2; font-weight: 600; display: flex; align-items: center; gap: 4px;">
						<span style="font-size: 10px; color: #969799;">{{ expandedTeams.has(t.id) ? '▼' : '▶' }}</span>
						{{ t.team_name }}
					</span>
					<span style="flex: 1; text-align: center; color: #1989fa; font-weight: 700;">{{ t.group_points || 0 }}</span>
					<span style="flex: 1.2; text-align: center; font-weight: 600; font-size: 12px;">{{ t.games_won || 0 }}:{{ t.games_lost || 0 }}</span>
					<span style="flex: 1; text-align: center; font-weight: 600; font-size: 12px;">{{ t.points_scored || 0 }}</span>
					<span style="flex: 0.8; text-align: center; font-weight: 700; color: #f5a623;">{{ t.group_rank || '-' }}</span>
				</div>
				<div v-if="expandedTeams.has(t.id)" style="padding: 0 12px 10px 28px; display: flex; flex-wrap: wrap; gap: 6px;">
					<span v-for="p in t.players" :key="p.id"
						:style="{
							padding: '4px 10px', borderRadius: '14px', fontSize: '12px', fontWeight: 600,
							background: p.is_seed ? '#fff3e0' : '#f0f2f5',
							color: p.is_seed ? '#e65100' : '#333',
							border: p.is_seed ? '1px solid #ffcc80' : '1px solid transparent',
						}">
						{{ p.role }}·{{ p.name }}
						<span v-if="p.is_seed" style="font-size: 10px;">⭐</span>
					</span>
					<span v-if="!t.players?.length" style="font-size: 12px; color: #969799;">暂无队员</span>
				</div>
			</div>
		</div>
		<div style="font-size: 11px; color: #969799; margin-top: 6px; padding: 0 2px;">
			积分：胜场 2 分 / 负场 1 分 · 并列先看互战（不含场外队），互战再同分则看局比/总分
		</div>
		<div v-if="miniTableForGroup(String(groupName))" style="margin-top: 8px; background: #fff7e8; border-radius: 10px; padding: 10px 12px;">
			<div style="font-size: 12px; font-weight: 700; color: #ed6a0c; margin-bottom: 6px;">互战小分（仅并列队之间）</div>
			<div v-for="row in miniTableForGroup(String(groupName))" :key="row.team.id"
				style="display: flex; font-size: 12px; padding: 3px 0; color: #646566;">
				<span style="flex: 1.5; font-weight: 600;">{{ row.team.team_name }}</span>
				<span style="flex: 1;">互战积分 {{ row.pts }}</span>
				<span style="flex: 1;">场 {{ row.fw }}:{{ row.fl }}</span>
				<span style="flex: 1;">局 {{ row.gw }}:{{ row.gl }}</span>
				<span style="flex: 1;">分 {{ row.ps }}</span>
			</div>
		</div>
	</div>

	<!-- Manual rank editor -->
	<div v-if="manualRankGroup" style="position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 600; display: flex; align-items: center; justify-content: center;" @click.self="manualRankGroup = null">
		<div style="background: #fff; border-radius: 16px; padding: 20px; width: 90%; max-width: 360px;">
			<div style="font-weight: 700; font-size: 16px; margin-bottom: 12px;">{{ manualRankGroup }}组 · 手动排名</div>
			<div v-for="t in (groupedTeams[manualRankGroup] || [])" :key="t.id"
				style="display: flex; align-items: center; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f0f2f5;">
				<span style="font-weight: 600;">{{ t.team_name }}</span>
				<select v-model.number="manualRanks[t.id]" style="padding: 6px 10px; border: 1.5px solid #ebedf0; border-radius: 8px;">
					<option v-for="n in (groupedTeams[manualRankGroup] || []).length" :key="n" :value="n">第 {{ n }} 名</option>
				</select>
			</div>
			<div style="display: flex; gap: 10px; margin-top: 16px;">
				<button @click="manualRankGroup = null" style="flex: 1; padding: 12px; border: 1.5px solid #ccc; border-radius: 12px; background: #fff; cursor: pointer;">取消</button>
				<button @click="saveManualRanks" style="flex: 2; padding: 12px; border: none; border-radius: 12px; background: #1989fa; color: #fff; font-weight: 700; cursor: pointer;">保存名次</button>
			</div>
		</div>
	</div>

	<!-- Matches per group -->
	<div v-for="(matches, groupName) in groupedMatches()" :key="'m'+groupName" style="margin: 16px;">
		<div style="font-weight: 700; font-size: 15px; margin-bottom: 8px; color: #333;">{{ groupName }}组 对阵</div>

		<div v-for="tm in matches" :key="tm.id"
			style="background: #fff; border-radius: 12px; margin-bottom: 8px; overflow: hidden; box-shadow: var(--c-shadow);">
			<div @click="toggleTeamMatch(tm.id)"
				style="display: flex; align-items: center; justify-content: space-between; padding: 14px; cursor: pointer;">
				<div style="display: flex; align-items: center; gap: 8px;">
					<span style="font-size: 18px; font-weight: 800; color: #1989fa;">{{ teamMatchScore(tm) }}</span>
					<span style="font-weight: 600;">{{ tm.team_a_name }} vs {{ tm.team_b_name }}</span>
				</div>
				<div style="display: flex; gap: 8px; align-items: center;">
					<span v-if="tm.played" class="badge badge-success">已结束</span>
					<span v-else-if="readyToComplete(tm)" class="badge" style="background: #e8f8ef; color: #07c160;">待提交</span>
					<span v-else class="badge badge-warning">进行中</span>
					<span style="font-size: 12px; color: #969799;">{{ expandedTM === tm.id ? '▲' : '▼' }}</span>
				</div>
			</div>

			<div v-if="tm.cards.length > 0" style="padding: 0 14px 8px; display: flex; gap: 6px;">
				<span v-for="c in tm.cards" :key="c.id"
					style="font-size: 11px; padding: 2px 8px; border-radius: 10px; background: #fff3e0; color: #e65100; font-weight: 600;">
					{{ c.team_id === tm.team_a_id ? tm.team_a_name : tm.team_b_name }}: {{ cardLabel(c.card_type) }}
				</span>
			</div>

			<div v-if="expandedTM === tm.id" style="border-top: 1px solid #f0f2f5;">
				<div v-if="!readonly" style="padding: 10px 14px; display: flex; flex-wrap: wrap; gap: 8px; justify-content: center;">
					<button v-if="!tm.played" @click="openLineup(tm)"
						style="padding: 8px 14px; background: #fff; border: 1.5px solid #1989fa; color: #1989fa; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">👥 配置出场 ABC</button>
					<button v-if="tm.cards.length < 2 && !tm.played" @click="openCardDraw(tm)"
						style="padding: 8px 14px; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🎴 抽趣味卡</button>
					<button v-if="readyToComplete(tm)" @click="completeTeamMatch(tm)"
						style="padding: 8px 14px; background: linear-gradient(135deg, #07c160, #06ad56); color: #fff; border: none; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">✅ 提交结束本场</button>
					<button v-if="tm.played" @click="reopenTeamMatch(tm)"
						style="padding: 8px 14px; background: #fff; border: 1.5px solid #ed6a0c; color: #ed6a0c; border-radius: 18px; font-size: 13px; font-weight: 700; cursor: pointer;">🔓 重新打开改分</button>
				</div>
				<div v-if="!readonly && readyToComplete(tm)" style="text-align: center; font-size: 12px; color: #07c160; padding: 0 14px 8px;">
					已达 3 胜，请核对比分后点击「提交结束本场」计入积分榜
				</div>

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
					<button v-if="!readonly && !tm.played" @click="openScoreEditor(tm, m)"
						style="padding: 6px 12px; background: #1989fa; color: #fff; border: none; border-radius: 14px; font-size: 12px; font-weight: 600; cursor: pointer;">
						{{ m.played ? '改分' : '录入' }}
					</button>
					<span v-else-if="m.played" style="font-size: 11px; font-weight: 600; color: #07c160;">{{ matchWinnerLabel(m) }} 胜</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Lineup editor -->
	<div v-if="lineupTM" style="position: fixed; inset: 0; background: rgba(0,0,0,0.45); z-index: 600; display: flex; align-items: center; justify-content: center;" @click.self="lineupTM = null">
		<div style="background: #fff; border-radius: 16px; padding: 20px; width: 92%; max-width: 400px; max-height: 85vh; overflow: auto;">
			<div style="font-weight: 700; font-size: 16px; margin-bottom: 4px;">配置出场 ABC</div>
			<div style="font-size: 12px; color: #969799; margin-bottom: 14px;">{{ lineupTM.team_a_name }} vs {{ lineupTM.team_b_name }} · 每场单独设置</div>

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

			<div style="font-size: 11px; color: #969799; margin-bottom: 12px;">
				对阵顺序：A单 / BC双 / C单 / AB双 / B单
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
		@update:show="showScore = $event"
		@submit="handleScore"
		@forfeit="handleForfeit"
	/>

	<TournamentCardDraw
		:show="showCardDraw"
		:drawing="drawResult === null"
		:result="drawResult"
		:fail-tick="drawFailTick"
		:team-a-name="drawingTM?.team_a_name || ''"
		:team-b-name="drawingTM?.team_b_name || ''"
		@draw="handleDrawCard"
		@close="showCardDraw = false; emit('refresh')"
	/>
	<!--#endregion -->
</template>
