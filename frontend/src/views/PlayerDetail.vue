<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { IconListDetails } from '@tabler/icons-vue'
import { api, type Player, type Match } from '../api'

const route = useRoute()
const router = useRouter()
const player = ref<Player | null>(null)
const matches = ref<Match[]>([])
const wins = ref(0)
const losses = ref(0)
const forfeitWins = ref(0)
const forfeits = ref(0)
const loading = ref(true)

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) { router.replace({ name: 'Home' }); return }
  try {
    const data = await api.getPlayer(id)
    player.value = data.player
    matches.value = data.matches
    wins.value = data.wins ?? 0
    losses.value = data.losses ?? 0
    forfeitWins.value = data.forfeit_wins ?? 0
    forfeits.value = data.forfeits ?? 0
  } catch (e: any) {
    showToast('加载失败')
    router.replace({ name: 'Home' })
  } finally { loading.value = false }
})

function ratingColor(change: number) {
  if (change > 0) return 'up'
  if (change < 0) return 'down'
  return ''
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function isWin(m: Match, playerId: number) { return m.winner_id === playerId }

const totalMatches = () => wins.value + losses.value
const winRate = () => totalMatches() > 0 ? (wins.value / totalMatches() * 100).toFixed(0) + '%' : '-'
</script>

<template>
  <div class="safe-bottom">
    <van-nav-bar title="球员详情" left-arrow @click-left="router.back" fixed placeholder />

    <van-loading v-if="loading" class="empty-state" />

    <template v-else-if="player">
      <div class="detail-hero">
        <div class="detail-name">{{ player.name }}</div>
        <div class="detail-rating">{{ player.current_rating }}</div>
        <div class="detail-rating-change">
          初始 {{ player.initial_rating }}
          ·
          <span :class="player.current_rating - player.initial_rating >= 0 ? 'mp-rating-change up' : 'mp-rating-change down'" style="font-size:14px;">
            {{ player.current_rating - player.initial_rating >= 0 ? '+' : '' }}{{ player.current_rating - player.initial_rating }}
          </span>
        </div>
      </div>

      <div class="detail-stats">
        <div class="detail-stat">
          <div class="ds-val">{{ totalMatches() }}</div>
          <div class="ds-label">总场次</div>
        </div>
        <div class="detail-stat">
          <div class="ds-val ds-win">{{ wins }}</div>
          <div class="ds-label">胜</div>
        </div>
        <div class="detail-stat">
          <div class="ds-val ds-loss">{{ losses }}</div>
          <div class="ds-label">负</div>
        </div>
        <div class="detail-stat">
          <div class="ds-val ds-rate">{{ winRate() }}</div>
          <div class="ds-label">胜率</div>
        </div>
        <div class="detail-stat" v-if="forfeits">
          <div class="ds-val" style="color:#ff976a;">{{ forfeits }}</div>
          <div class="ds-label">弃权</div>
        </div>
      </div>

      <div class="section-title">
        <IconListDetails :size="18" :stroke-width="2" style="vertical-align: -3px; margin-right: 6px;" />
        对战记录
      </div>

      <div v-if="matches.length === 0" class="empty-state" style="padding: 40px 20px;">
        <p>暂无对战记录</p>
      </div>

      <div style="padding: 0 16px;" v-else>
        <div v-for="m in matches" :key="m.id" class="match-card">
          <div class="match-header">
            <div class="match-date">{{ formatDate(m.played_at) }}</div>
            <span class="match-result-tag" :class="m.forfeit ? '' : (isWin(m, player.id) ? 'win' : 'loss')"
              :style="m.forfeit ? 'background:#fff3ed;color:#ff976a;' : ''">
              {{ m.forfeit ? '弃权' : (isWin(m, player.id) ? '胜' : '负') }}
            </span>
          </div>
          <div class="match-body">
            <div class="match-player" :class="{ winner: m.winner_id === m.player_a_id && !m.forfeit }">
              <div class="mp-name">{{ m.player_a_name }}</div>
              <div class="mp-rating-change" :class="ratingColor(m.rating_change_a)">
                {{ m.rating_change_a >= 0 ? '+' : '' }}{{ m.rating_change_a }}
              </div>
            </div>
            <div class="match-score-badge" :style="m.forfeit ? 'color:#ff976a;' : ''">
              {{ m.forfeit ? '弃权' : `${m.score_a} : ${m.score_b}` }}
            </div>
            <div class="match-player" :class="{ winner: m.winner_id === m.player_b_id && !m.forfeit }">
              <div class="mp-name">{{ m.player_b_name }}</div>
              <div class="mp-rating-change" :class="ratingColor(m.rating_change_b)">
                {{ m.rating_change_b >= 0 ? '+' : '' }}{{ m.rating_change_b }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div style="height: 16px;"></div>
    </template>
  </div>
</template>
