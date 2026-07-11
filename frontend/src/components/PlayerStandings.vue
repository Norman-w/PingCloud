<script setup lang="ts">
import { IconChartBar } from '@tabler/icons-vue'

export interface StandingPlayer {
  id: number
  name: string
  wins: number
  losses: number
  rating: number
  ratingLabel?: string
  ratingChange?: number
  forfeits?: number
  points?: number
  gameWins?: number
  gameLosses?: number
  pointsFor?: number
  pointsAgainst?: number
}

defineProps<{
  title?: string
  players: StandingPlayer[]
  ratingLabel?: string
}>()
</script>

<template>
  <div style="background:#fff;border-radius:12px;margin:4px 16px 12px;box-shadow:0 2px 12px rgba(0,0,0,0.06);overflow:hidden;">
    <div v-if="title" style="font-size:16px;font-weight:600;padding:12px 16px 4px;display:flex;align-items:center;gap:6px;">
      <IconChartBar :size="18" :stroke-width="2" style="vertical-align:-3px;" />
      {{ title }}
    </div>
    <div v-for="(p, i) in players" :key="p.id"
      style="display:flex;align-items:center;padding:12px 16px;border-bottom:1px solid #f5f5f5;">
      <div style="width:28px;height:28px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:13px;font-weight:700;"
        :style="{background:i===0?'#fff3cd':i===1?'#e8e8e8':i===2?'#ffe8d6':'#f0f2f5',color:i===0?'#b8860b':i===1?'#666':i===2?'#b87333':'#969799'}">{{ i + 1 }}</div>
      <div style="flex:1;margin-left:12px;">
        <div style="font-size:15px;font-weight:500;">{{ p.name }} <span style="font-size:11px;color:#c8c9cc;">#{{ i + 1 }}</span></div>
        <div style="font-size:12px;color:#969799;">
          {{ p.wins }}胜 {{ p.losses }}负
          <template v-if="p.points !== undefined"><span style="color:#1989fa;font-weight:500;">· {{ p.points }}分</span></template>
          <template v-if="p.gameWins !== undefined">· 局{{ p.gameWins }}胜{{ p.gameLosses }}负</template>
          <template v-if="p.pointsFor !== undefined">· 小分+{{ p.pointsFor }}/-{{ p.pointsAgainst }}</template>
          <template v-if="p.forfeits">· 弃权{{ p.forfeits }}</template>
        </div>
      </div>
      <div style="text-align:right;">
        <div style="font-size:16px;font-weight:700;color:#1989fa;">{{ p.rating }}</div>
        <div v-if="p.ratingChange !== undefined" style="font-size:11px;"
          :style="{color: p.ratingChange >= 0 ? '#07c160' : '#ee0a24', fontWeight: 600}">
          {{ p.ratingChange >= 0 ? '+' : '' }}{{ p.ratingChange }}
        </div>
        <div v-if="ratingLabel" style="font-size:10px;color:#969799;">{{ ratingLabel }}</div>
      </div>
    </div>
    <div v-if="players.length === 0" style="text-align:center;padding:20px;color:#c8c9cc;font-size:13px;">
      暂无数据
    </div>
  </div>
</template>
