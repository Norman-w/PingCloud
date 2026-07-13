<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { IconPingPong, IconSwords, IconTrophy, IconScoreboard, IconSpeakerphone, IconTournament, IconPlus, IconChevronRight } from '@tabler/icons-vue'
import { api, type RankingEntry } from '../api'
import { unreadCount } from '../bulletins'
import { myId } from '../auth'

const router = useRouter()
const rankings = ref<RankingEntry[]>([])
const loading = ref(true)
const showAll = ref(false)
const hideInactive = ref(false)

const displayRankings = computed(() => {
  let list = rankings.value
  if (hideInactive.value) list = list.filter(p => p.matches_played > 0)
  if (!showAll.value) return list.slice(0, 10)
  return list
})
const hiddenCount = computed(() => {
  let list = rankings.value
  if (hideInactive.value) list = list.filter(p => p.matches_played > 0)
  return Math.max(0, list.length - 10)
})

// Sessions on home page
interface SessionSummary {
  id: number; name: string; status: string; player_count: number; match_count: number; unplayed_count: number
}
const homeSessions = ref<SessionSummary[]>([])
const loadingSessions = ref(false)

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
  // Load sessions
  loadingSessions.value = true
  try {
    homeSessions.value = await fetch('/api/sessions').then(r => r.json()).catch(() => [])
  } catch { /* ignore */ }
  finally { loadingSessions.value = false }
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
        <span>快速开局</span>
      </button>
      <button class="quick-action" style="background: linear-gradient(135deg, #e67e22, #f39c12);" @click="router.push({ name: 'HeadToHead' })">
        <IconSwords :size="28" :stroke-width="2" />
        <span>相生相克</span>
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

    <!-- Toggle: hide inactive -->
    <div v-else style="display:flex;align-items:center;gap:8px;padding:4px 16px 8px;font-size:12px;">
      <span @click="hideInactive=!hideInactive" style="display:flex;align-items:center;gap:4px;cursor:pointer;color:#969799;user-select:none;">
        <span style="width:18px;height:18px;border-radius:5px;border:2px solid #d0d0d0;display:flex;align-items:center;justify-content:center;transition:all .15s;" :style="hideInactive?'background:#1989fa;border-color:#1989fa;':''">
          <span v-if="hideInactive" style="color:#fff;font-size:11px;font-weight:700;">✓</span>
        </span>
        隐藏未参与者
      </span>
    </div>

    <div class="card" style="padding: 0; overflow: hidden;">
      <div
        v-for="(p, i) in displayRankings"
        :key="p.id"
        class="player-item"
        :style="myId>0 && p.id===myId ? {background:'#e8f4ff',borderLeft:'3px solid #1989fa'} : {}"
        @click="goPlayer(p.id)"
      >
        <div class="player-rank" :class="rankClass(i)">{{ i + 1 }}</div>
        <div class="player-info">
          <div class="player-name" :style="myId>0&&p.id===myId?{color:'#1989fa',fontWeight:700}:{}">{{ p.name }}{{ myId>0&&p.id===myId?' 👈':'' }}</div>
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

    <!-- Load more / show less -->
    <div v-if="hiddenCount > 0 && !showAll" @click="showAll=true"
      style="text-align:center;padding:12px;color:#1989fa;font-size:14px;font-weight:600;cursor:pointer;background:#fff;margin:0 16px;border-radius:0 0 12px 12px;box-shadow:var(--c-shadow);margin-top:-1px;">
      加载更多 ({{ hiddenCount }}人) ▾
    </div>
    <div v-else-if="showAll && rankings.length > 10" @click="showAll=false"
      style="text-align:center;padding:10px;color:#969799;font-size:13px;cursor:pointer;">
      收起 ▴
    </div>

    <!-- Recent Sessions -->
    <div class="section-title" style="justify-content: space-between; margin-top: 8px;">
      <span>
        <IconTournament :size="18" :stroke-width="2" style="vertical-align: -3px; margin-right: 6px;" />
        近期活动
      </span>
      <span @click="router.push({ name: 'SessionView', query: { new: '1' } })" style="font-size: 13px; color: #1989fa; font-weight: 600; cursor: pointer; background: #e8f4ff; padding: 4px 12px; border-radius: 12px;">
        <IconPlus :size="14" :stroke-width="2" style="vertical-align:-2px;" /> 新建
      </span>
    </div>

    <van-loading v-if="loadingSessions" style="padding: 20px; text-align: center;" />

    <div v-else-if="homeSessions.length === 0" style="text-align:center;padding:24px 20px;color:var(--c-text-hint);font-size:14px;">
      暂无活动，<span @click="router.push({ name: 'SessionView', query: { new: '1' } })" style="color:#1989fa;cursor:pointer;">创建第一个</span>
    </div>

    <div v-else v-for="s in homeSessions.slice(0, 5)" :key="s.id"
      @click="router.push({ name: 'SessionView' })"
      class="card" style="margin: 8px 16px; padding: 14px 16px; cursor: pointer;">
      <div style="display: flex; justify-content: space-between; align-items: center;">
        <div style="flex:1;min-width:0;">
          <div style="font-weight: 600; font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ s.name }}</div>
          <div style="font-size: 12px; color: var(--c-text-hint); margin-top: 3px;">{{ s.player_count }} 人 · {{ s.match_count }} 场</div>
        </div>
        <div style="display: flex; align-items: center; gap: 8px; flex-shrink: 0;">
          <span style="font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: 10px;"
            :style="s.status==='completed'?'background:#e8f8ef;color:#07c160;':'background:#e8f4ff;color:#1989fa;'">
            {{ s.status === 'completed' ? '已结束' : '进行中' }}
          </span>
          <IconChevronRight :size="16" :stroke-width="2" style="color: #c8c9cc;" />
        </div>
      </div>
    </div>

    <div v-if="homeSessions.length > 5" @click="router.push({ name: 'SessionView' })"
      style="text-align:center;padding:12px;color:#1989fa;font-size:14px;cursor:pointer;">
      查看全部 {{ homeSessions.length }} 个活动 ▸
    </div>

    <div style="height: 8px;"></div>
  </div>
</template>
