<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import { IconSwords, IconUsers, IconPingPong } from '@tabler/icons-vue'
import { api, type Player } from '../api'

const players = ref<Player[]>([])
const loading = ref(true)
const submitting = ref(false)

const selectedA = ref<Player | null>(null)
const selectedB = ref<Player | null>(null)
const scoreA = ref('')
const scoreB = ref('')
const format = ref(3) // 2=BO3, 3=BO5, 4=BO7

function formatLabel(wins: number) {
  if (wins === 2) return '三局两胜'
  if (wins === 4) return '七局四胜'
  return '五局三胜'
}

function quickScores(): [number, number][] {
  const w = format.value
  const scores: [number, number][] = []
  // Winner = player A
  for (let l = 0; l < w; l++) scores.push([w, l])
  // Winner = player B
  for (let l = 0; l < w; l++) scores.push([l, w])
  return scores
}

onMounted(async () => {
  try {
    players.value = await api.getPlayers()
  } catch (e: any) {
    showToast('加载球员失败')
  } finally {
    loading.value = false
  }
})

function selectPlayer(player: Player) {
  if (selectedA.value?.id === player.id) { selectedA.value = null; return }
  if (selectedB.value?.id === player.id) { selectedB.value = null; return }
  if (!selectedA.value) { selectedA.value = player }
  else if (!selectedB.value) {
    if (player.id === selectedA.value.id) { showToast('不能选同一个人'); return }
    selectedB.value = player
  } else { selectedA.value = player; selectedB.value = null; scoreA.value = ''; scoreB.value = '' }
}

function clearSelection() {
  selectedA.value = null; selectedB.value = null; scoreA.value = ''; scoreB.value = ''
}

function quickScore(a: number, b: number) { scoreA.value = String(a); scoreB.value = String(b) }

async function submitMatch() {
  if (!selectedA.value || !selectedB.value) return
  const sa = parseInt(scoreA.value), sb = parseInt(scoreB.value)
  if (isNaN(sa) || isNaN(sb) || sa === sb) { showToast('请输入有效比分'); return }
  submitting.value = true
  try {
    const match = await api.createMatch({
      player_a_id: selectedA.value.id, player_b_id: selectedB.value.id,
      score_a: sa, score_b: sb,
    })
    showSuccessToast(`${match.player_a_name} ${match.rating_change_a >= 0 ? '+' : ''}${match.rating_change_a} | ${match.player_b_name} ${match.rating_change_b >= 0 ? '+' : ''}${match.rating_change_b}`)
    players.value = await api.getPlayers()
    clearSelection()
  } catch (e: any) { showToast('提交失败: ' + e.message) }
  finally { submitting.value = false }
}

const showAdd = ref(false)
const newName = ref('')
const newRating = ref('')
const adding = ref(false)

async function quickAddPlayer() {
  if (!newName.value.trim()) { showToast('请输入姓名'); return }
  adding.value = true
  try {
    const rating = newRating.value ? parseInt(newRating.value) : undefined
    const player = await api.createPlayer({ name: newName.value.trim(), initial_rating: rating })
    players.value.push(player)
    showAdd.value = false; newName.value = ''; newRating.value = ''
    showSuccessToast(`已添加 ${player.name}`)
    selectPlayer(player)
  } catch (e: any) { showToast('添加失败: ' + e.message) }
  finally { adding.value = false }
}
</script>

<template>
  <div class="safe-bottom">
    <div class="hero">
      <div class="hero-title">
        <IconSwords :size="26" :stroke-width="2" style="vertical-align: -5px; margin-right: 4px;" />
        快速记分
      </div>
      <div class="hero-sub">轻点选人 → 输入比分 → 提交</div>
    </div>

    <div class="select-hint" style="margin-top: 16px;">
      <div class="select-hint-item" :class="{ filled: selectedA }" @click="selectedA = null">
        <div class="select-hint-label">选手 A</div>
        <template v-if="selectedA">
          <div class="select-hint-name">{{ selectedA.name }}</div>
          <div class="select-hint-rating">{{ selectedA.current_rating }} 分</div>
        </template>
        <div v-else class="select-hint-empty">点击下方选择</div>
      </div>
      <div class="select-hint-item" :class="{ filled: selectedB }" @click="selectedB = null">
        <div class="select-hint-label">选手 B</div>
        <template v-if="selectedB">
          <div class="select-hint-name">{{ selectedB.name }}</div>
          <div class="select-hint-rating">{{ selectedB.current_rating }} 分</div>
        </template>
        <div v-else class="select-hint-empty">点击下方选择</div>
      </div>
    </div>

    <div v-if="selectedA && selectedB" class="card">
      <!-- Format selector -->
      <div style="display: flex; gap: 8px; margin-bottom: 16px;">
        <button v-for="w in [2,3,4]" :key="w" @click="format = w"
          style="flex: 1; padding: 10px 0; border-radius: 10px; font-size: 14px; font-weight: 600; border: 2px solid; cursor: pointer; transition: all 0.15s;"
          :style="format === w ? { background: '#1989fa', color: '#fff', borderColor: '#1989fa' } : { background: '#fff', color: '#666', borderColor: '#ebedf0' }">
          {{ formatLabel(w) }}
        </button>
      </div>

      <div style="display: flex; align-items: center; justify-content: center; gap: 16px; margin-bottom: 12px;">
        <div style="text-align: center;">
          <div style="font-size: 13px; color: #969799;">{{ selectedA.name }}</div>
          <input v-model="scoreA" type="number" min="0" :max="format" placeholder="0"
            style="width: 56px; height: 44px; text-align: center; font-size: 28px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
        <div style="font-size: 28px; font-weight: 800; color: #969799; padding-top: 16px;">:</div>
        <div style="text-align: center;">
          <div style="font-size: 13px; color: #969799;">{{ selectedB.name }}</div>
          <input v-model="scoreB" type="number" min="0" :max="format" placeholder="0"
            style="width: 56px; height: 44px; text-align: center; font-size: 28px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
      </div>

      <div style="display: flex; gap: 6px; flex-wrap: wrap; justify-content: center; margin-bottom: 12px;">
        <button v-for="(s, i) in quickScores()" :key="i"
          @click="quickScore(s[0], s[1])"
          style="padding: 6px 14px; border-radius: 16px; border: none; font-size: 14px; font-weight: 500; cursor: pointer;"
          :style="s[0] > s[1] ? { background: '#e8f8ef', color: '#07c160' } : { background: '#fde8e8', color: '#ee0a24' }">
          {{ s[0] }}:{{ s[1] }}
        </button>
      </div>

      <button @click="submitMatch" :disabled="submitting || !scoreA || !scoreB || scoreA === scoreB"
        style="width: 100%; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 16px; font-weight: 600; cursor: pointer;">
        {{ submitting ? '提交中...' : '确认提交' }}
      </button>
      <div style="text-align: center; margin-top: 10px;">
        <button @click="clearSelection" style="background: none; border: none; color: #969799; font-size: 14px; cursor: pointer;">重新选择</button>
      </div>
    </div>

    <van-loading v-if="loading" class="empty-state" />

    <div v-else>
      <div class="section-title">
        <IconUsers :size="18" :stroke-width="2" style="vertical-align: -3px; margin-right: 6px;" />
        {{ selectedA ? '选择选手 B' : '选择选手 A' }}
      </div>

      <div v-if="!loading && players.length === 0" class="empty-state">
        <IconPingPong :size="48" :stroke-width="1" style="color: #c8c9cc; margin-bottom: 12px;" />
        <p style="margin-bottom: 12px;">还没有球员</p>
        <van-button type="primary" round @click="showAdd = true">快速添加球员</van-button>
      </div>

      <div class="card" style="padding: 0; overflow: hidden;">
        <div v-for="p in players" :key="p.id" class="player-item"
          @click="selectPlayer(p)"
          :class="{ 'selected-a': selectedA?.id === p.id, 'selected-b': selectedB?.id === p.id }">
          <div class="player-info">
            <div class="player-name">{{ p.name }}</div>
            <div class="player-record">{{ p.current_rating }} 分 · 初始 {{ p.initial_rating }}</div>
          </div>
          <van-tag v-if="selectedA?.id === p.id" type="primary">A</van-tag>
          <van-tag v-else-if="selectedB?.id === p.id" type="success">B</van-tag>
        </div>
        <div class="player-item" style="justify-content: center; color: var(--c-primary); font-weight: 500;"
          @click="showAdd = true">
          快速添加新球员
        </div>
      </div>
    </div>

    <van-popup v-model:show="showAdd" position="bottom" round :style="{ padding: '24px 16px' }">
      <h3 style="text-align: center; margin-bottom: 20px; font-size: 18px;">快速添加球员</h3>
      <van-field v-model="newName" label="姓名" placeholder="输入姓名" clearable />
      <van-field v-model="newRating" label="初始积分" placeholder="默认 1500" type="number" clearable />
      <div style="padding: 20px 0 8px;">
        <van-button type="primary" block round :loading="adding" @click="quickAddPlayer" loading-text="添加中...">
          确认添加
        </van-button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.player-item.selected-a { background: #e8f4ff; border-left: 3px solid #1989fa; }
.player-item.selected-b { background: #e8f8ef; border-left: 3px solid #07c160; }
</style>
