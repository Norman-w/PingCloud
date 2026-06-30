<script setup lang="ts">
import { IconChartBar, IconList, IconScoreboard } from '@tabler/icons-vue'
import ScoreDialog from './ScoreDialog.vue'
import AddPlayerDialog from './AddPlayerDialog.vue'
import { sessionChange, sessionDisplayRating, changeSign, type SessionDetail, type SessionMatch } from '../session-utils'
import type { Player } from '../api'

defineProps<{
  session: SessionDetail
  players: Player[]
  showEditDialog: boolean
  scoringMatch: SessionMatch | null
  showAddPlayerDialog: boolean
  addPlayerId: number
}>()

const emit = defineEmits<{
  (e: 'update:showEditDialog', v: boolean): void
  (e: 'update:showAddPlayerDialog', v: boolean): void
  (e: 'update:addPlayerId', v: number): void
  (e: 'openScoreEditor', m: SessionMatch): void
  (e: 'submitScore', scoreA: number, scoreB: number): void
  (e: 'forfeit', winnerId: number): void
  (e: 'addPlayer'): void
  (e: 'completeSession'): void
  (e: 'backToList'): void
  (e: 'editNameStart'): void
}>()
</script>

<template>
  <div>

    <!-- Session info -->
    <div style="background:#fff;border-radius:12px;padding:16px;margin:12px 16px;box-shadow:0 2px 12px rgba(0,0,0,0.06);">
      <div style="display:flex;align-items:center;justify-content:center;gap:8px;">
        <span style="font-weight:700;font-size:18px;">{{ session.name }}</span>
        <span @click="emit('editNameStart')" style="cursor:pointer;color:#c8c9cc;font-size:16px;">&#9998;</span>
      </div>
      <div style="font-size:13px;color:#969799;margin-top:6px;text-align:center;">
        {{ session.matches.filter(m=>m.played).length }} / {{ session.matches.length }} 场已完成
      </div>
    </div>

    <!-- Live standings -->
    <div style="font-size:16px;font-weight:600;padding:16px 16px 8px;display:flex;align-items:center;gap:6px;">
      <IconChartBar :size="18" :stroke-width="2" style="vertical-align:-3px;" />
      实时排名
    </div>
    <div style="background:#fff;border-radius:12px;margin:4px 16px 12px;box-shadow:0 2px 12px rgba(0,0,0,0.06);overflow:hidden;">
      <div v-for="(p,i) in session.players" :key="p.id"
        style="display:flex;align-items:center;padding:14px 16px;border-bottom:1px solid #f5f5f5;">
        <div style="width:28px;height:28px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:13px;font-weight:700;"
          :style="{background:i===0?'#fff3cd':i===1?'#e8e8e8':i===2?'#ffe8d6':'#f0f2f5',color:i===0?'#b8860b':i===1?'#666':i===2?'#b87333':'#969799'}">{{ i+1 }}</div>
        <div style="flex:1;margin-left:12px;">
          <div style="font-size:16px;font-weight:500;">{{ p.name }} <span style="font-size:11px;color:#c8c9cc;">#{{ i+1 }}</span></div>
          <div style="font-size:12px;color:#969799;">{{ p.wins }}胜 {{ p.losses }}负</div>
        </div>
        <div style="font-size:18px;font-weight:700;color:#1989fa;">{{ sessionDisplayRating(p, session.matches) }}</div>
      </div>
    </div>

    <!-- Match list -->
    <div style="font-size:16px;font-weight:600;padding:16px 16px 8px;display:flex;align-items:center;gap:6px;">
      <IconList :size="18" :stroke-width="2" style="vertical-align:-3px;" />
      对阵表
    </div>
    <div style="background:#fff;border-radius:12px;margin:4px 16px;box-shadow:0 2px 12px rgba(0,0,0,0.06);overflow:hidden;">
      <div v-for="(m,mi) in session.matches" :key="m.id"
        @click="emit('openScoreEditor', m)"
        style="display:flex;align-items:center;padding:14px 16px;border-bottom:1px solid #f5f5f5;cursor:pointer;">
        <span style="display:flex;flex-direction:column;align-items:center;gap:2px;flex-shrink:0;min-width:36px;">
          <a :href="`/#/scoreboard?a=${encodeURIComponent(m.player_a_name)}&b=${encodeURIComponent(m.player_b_name)}`" target="_blank" style="text-decoration:none;font-size:14px;line-height:1;" @click.stop title="记分牌">
            <IconScoreboard :size="16" :stroke-width="1.5" style="color:#1989fa;" />
          </a>
          <span style="font-size:12px;color:#c8c9cc;">#{{ mi+1 }}</span>
        </span>
        <div style="flex:1;text-align:right;font-weight:400;" :style="{fontWeight:m.winner_id===m.player_a_id?700:400}">
          <span style="font-size:10px;color:#c8c9cc;">{{ mi+1 }}号</span> {{ m.player_a_name }}
          <span v-if="m.played" style="font-size:11px;display:block;" :style="{color:m.rating_change_a>=0?'#07c160':'#ee0a24'}">{{ changeSign(m.rating_change_a) }}{{ m.rating_change_a }}</span>
        </div>
        <div style="width:64px;text-align:center;font-weight:700;font-size:16px;">
          <template v-if="m.forfeit"><span style="color:#ff976a;font-weight:600;font-size:14px;">弃权</span></template>
          <template v-else-if="m.played">{{ m.score_a }} : {{ m.score_b }}</template>
          <template v-else><span style="color:#c8c9cc;font-size:13px;">待录入</span></template>
        </div>
        <div style="flex:1;font-weight:400;" :style="{fontWeight:m.winner_id===m.player_b_id?700:400}">
          <span style="font-size:10px;color:#c8c9cc;">{{ mi+1 }}号</span> {{ m.player_b_name }}
          <span v-if="m.played" style="font-size:11px;display:block;" :style="{color:m.rating_change_b>=0?'#07c160':'#ee0a24'}">{{ changeSign(m.rating_change_b) }}{{ m.rating_change_b }}</span>
        </div>
      </div>
    </div>

    <!-- Add player button -->
    <div style="padding:8px 16px;">
      <button @click="emit('update:showAddPlayerDialog', true); emit('update:addPlayerId', 0)"
        style="width:100%;padding:12px;background:#fff;color:#1989fa;border:2px dashed #1989fa;border-radius:12px;font-size:15px;font-weight:500;cursor:pointer;">+ 拉人加入本场</button>
    </div>

    <!-- Action buttons -->
    <div style="padding:16px;display:flex;gap:12px;">
      <button @click="emit('completeSession')" :disabled="session.matches.some(m=>!m.played)"
        style="flex:1;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:16px;font-weight:600;cursor:pointer;"
        :style="{opacity:session.matches.some(m=>!m.played)?0.5:1,cursor:session.matches.some(m=>!m.played)?'not-allowed':'pointer'}">结束活动 · 查看排名</button>
      <button @click="emit('backToList')" style="padding:16px 24px;background:#f5f5f5;border:none;border-radius:24px;font-size:16px;cursor:pointer;color:#666;">返回</button>
    </div>

    <!-- Dialogs -->
    <ScoreDialog :show="showEditDialog" :player-a-id="scoringMatch?.player_a_id||0" :player-b-id="scoringMatch?.player_b_id||0"
      :player-a-name="scoringMatch?.player_a_name||''" :player-b-name="scoringMatch?.player_b_name||''"
      :initial-score-a="scoringMatch?.played?scoringMatch.score_a:undefined" :initial-score-b="scoringMatch?.played?scoringMatch.score_b:undefined"
      @update:show="emit('update:showEditDialog', $event)" @submit="(a:number,b:number) => emit('submitScore', a, b)" @forfeit="emit('forfeit', $event)" />

    <AddPlayerDialog :show="showAddPlayerDialog" :players="players"
      :exclude-ids="session.players.map(p=>p.id)" :session-name="session.name"
      @update:show="emit('update:showAddPlayerDialog', $event)" @add="(pid:number) => { emit('update:addPlayerId', pid); emit('addPlayer') }" />
  </div>
</template>
