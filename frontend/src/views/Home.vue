<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { IconPingPong, IconSwords, IconUserPlus, IconTrophy, IconScoreboard, IconSpeakerphone } from '@tabler/icons-vue'
import { api, type RankingEntry } from '../api'
import { unreadCount } from '../bulletins'

const router = useRouter()
const rankings = ref<RankingEntry[]>([])
const loading = ref(true)

const stats = computed(() => {
  const total = rankings.value.length
  const matches = rankings.value.reduce((s, p) => s + p.matches_played, 0)
  const avgRating = total > 0
    ? Math.round(rankings.value.reduce((s, p) => s + p.current_rating, 0) / total)
    : 0
  return { total, matches, avgRating }
})

onMounted(async () => {
  try {
    rankings.value = await api.getRankings()
  } catch (e: any) {
    showToast('加载失败')
  } finally {
    loading.value = false
  }
})

function rankClass(index: number) {
  if (index === 0) return 'r1'
  if (index === 1) return 'r2'
  if (index === 2) return 'r3'
  return ''
}

function goPlayer(id: number) {
  router.push({ name: 'PlayerDetail', params: { id } })
}

function winRate(p: RankingEntry) {
  if (p.matches_played === 0) return '-'
  return p.win_rate.toFixed(0) + '%'
}
</script>

<template>
  <div class="safe-bottom">
    <div class="hero">
      <div class="hero-title">
        <IconPingPong :size="28" :stroke-width="2" style="vertical-align: -5px; margin-right: 4px;" />
        乒云
      </div>
      <div class="hero-sub" style="display:flex;align-items:center;justify-content:space-between;">
        <span>乒乓球积分排名系统</span>
        <span @click="router.push({name:'Bulletin'})" style="cursor:pointer;position:relative;font-size:12px;background:rgba(255,255,255,0.2);padding:4px 10px;border-radius:10px;">
          <IconSpeakerphone :size="14" :stroke-width="2" style="vertical-align:-2px;margin-right:4px;" />公告
          <span v-if="unreadCount()>0" style="position:absolute;top:-6px;right:-6px;background:#ee0a24;color:#fff;font-size:10px;min-width:16px;height:16px;border-radius:8px;display:flex;align-items:center;justify-content:center;font-weight:700;">{{unreadCount()}}</span>
        </span>
      </div>
    </div>

    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ stats.total }}</div>
        <div class="stat-label">球员总数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.matches }}</div>
        <div class="stat-label">比赛场次</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.avgRating }}</div>
        <div class="stat-label">平均积分</div>
      </div>
    </div>

    <div class="quick-actions">
      <button class="quick-action qa-match" @click="router.push({ name: 'SessionView', query: { new: '1' } })">
        <IconSwords :size="28" :stroke-width="2" />
        <span>快速记分</span>
      </button>
      <button class="quick-action qa-player" @click="router.push({ name: 'AddPlayer' })">
        <IconUserPlus :size="28" :stroke-width="2" />
        <span>添加球员</span>
      </button>
      <button class="quick-action" style="background: linear-gradient(135deg, #0a0a2e, #1a1a4e);" @click="router.push({ name: 'Scoreboard' })">
        <IconScoreboard :size="28" :stroke-width="2" />
        <span>记分牌</span>
      </button>
    </div>

    <div class="section-title" style="justify-content: space-between;">
      <span>
        <IconTrophy :size="18" :stroke-width="2" style="vertical-align: -3px; margin-right: 6px;" />
        积分排行
      </span>
      <span @click="router.push({ name: 'Rules' })" style="font-size: 13px; color: #1989fa; font-weight: 600; cursor: pointer; background: #e8f4ff; padding: 4px 12px; border-radius: 12px;">
        积分规则 ▸
      </span>
    </div>

    <van-loading v-if="loading" class="empty-state" />

    <div v-else-if="rankings.length === 0" class="empty-state">
      <IconPingPong :size="48" :stroke-width="1" style="color: #c8c9cc; margin-bottom: 12px;" />
      <p style="margin-bottom: 8px;">还没有球员</p>
      <p class="text-hint">点击上方按钮开始</p>
    </div>

    <div v-else class="card" style="padding: 0; overflow: hidden;">
      <div
        v-for="(p, i) in rankings"
        :key="p.id"
        class="player-item"
        @click="goPlayer(p.id)"
      >
        <div class="player-rank" :class="rankClass(i)">{{ i + 1 }}</div>
        <div class="player-info">
          <div class="player-name">{{ p.name }}</div>
          <div class="player-record">
            {{ p.matches_played }} 场 · {{ p.wins }} 胜 {{ p.losses }} 负
            <template v-if="p.forfeits"> · 弃权 {{ p.forfeits }}</template>
            · 胜率 {{ winRate(p) }}
          </div>
        </div>
        <div class="player-rating">
          {{ p.current_rating }}
        </div>
      </div>
    </div>

    <div style="height: 8px;"></div>
  </div>
</template>
