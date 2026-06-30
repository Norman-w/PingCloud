<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { showToast, showSuccessToast } from 'vant'
import { IconTournament, IconPlus, IconList, IconUsers, IconClipboardCheck, IconChartBar, IconTrophy, IconChevronRight } from '@tabler/icons-vue'
import SessionPlay from '../components/SessionPlay.vue'
import { api, type Player } from '../api'
import { unplayedCount, sessionChange, sessionDisplayRating, changeSign, selectedPlayers, type SessionMatch, type SessionPlayer, type SessionDetail } from '../session-utils'

const players = ref<Player[]>([])
const loadingPlayers = ref(true)

const step = ref<'list' | 'select' | 'confirm' | 'play' | 'result'>('list')
const selectedIDs = ref<Set<number>>(new Set())
const sessionName = ref('')

const currentSession = ref<SessionDetail | null>(null)
const sessions = ref<SessionDetail[]>([])

const scoringMatch = ref<SessionMatch | null>(null)
const showEditDialog = ref(false)
const submitting = ref(false)

const showAdd = ref(false)
const newName = ref('')
const newRating = ref('')
const adding = ref(false)
const loadingSession = ref(false)
const editingName = ref(false)
const editName = ref('')
const showAddPlayerDialog = ref(false)
const addPlayerId = ref(0)
const route = useRoute()

async function startEditName() {
  if (!currentSession.value) return
  editName.value = currentSession.value.name
  editingName.value = true
}

async function saveName() {
  if (!currentSession.value || !editName.value.trim()) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: editName.value.trim() }),
    })
    currentSession.value.name = editName.value.trim()
    editingName.value = false
  } catch (e: any) { showToast('修改失败') }
}

async function addPlayerToSession() {
  if (!currentSession.value || !addPlayerId.value) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}/players`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ player_id: addPlayerId.value }),
    })
    const detail = await fetch(`/api/sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
    showAddPlayerDialog.value = false
    showSuccessToast('已加入，对阵已更新')
  } catch (e: any) { showToast('添加失败') }
}

onMounted(async () => {
  await loadAll()
  if (route.query.new === '1') enterSelect()
})

async function loadAll() {
  loadingPlayers.value = true
  try {
    const [p, s] = await Promise.all([
      api.getPlayers(),
      fetch('/api/sessions').then(r => r.json()).catch(() => []),
    ])
    players.value = p; sessions.value = s
  } catch (e) { /* ignore */ }
  finally { loadingPlayers.value = false }
}

function enterSelect() { selectedIDs.value = new Set(); sessionName.value = ''; step.value = 'select' }
function togglePlayer(id: number) {
  const s = new Set(selectedIDs.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selectedIDs.value = s
}

async function quickAddPlayer() {
  if (!newName.value.trim()) { showToast('请输入姓名'); return }
  adding.value = true
  try {
    const rating = newRating.value ? parseInt(newRating.value) : undefined
    const p = await api.createPlayer({ name: newName.value.trim(), initial_rating: rating })
    players.value.push(p)
    const s = new Set(selectedIDs.value); s.add(p.id); selectedIDs.value = s
    showAdd.value = false; newName.value = ''; newRating.value = ''
    showSuccessToast(`已添加 ${p.name}`)
  } catch (e: any) { showToast('添加失败') }
  finally { adding.value = false }
}

function goConfirm() {
  if (selectedIDs.value.size < 2) { showToast('至少选 2 人'); return }
  if (!sessionName.value.trim()) {
    const names = Array.from(selectedIDs.value).map(id => players.value.find(p => p.id === id)?.name).filter(Boolean)
    sessionName.value = names.join('、') + ' 的对局'
  }
  step.value = 'confirm'
}

async function createSession() {
  submitting.value = true
  try {
    const res = await fetch('/api/sessions', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: sessionName.value.trim(), player_ids: Array.from(selectedIDs.value) }),
    })
    const { id } = await res.json()
    const detail = await fetch(`/api/sessions/${id}`).then(r => r.json())
    currentSession.value = detail; step.value = 'play'
  } catch (e: any) { showToast('创建失败: ' + e.message) }
  finally { submitting.value = false }
}

function openScoreEditor(match: SessionMatch) { scoringMatch.value = match; showEditDialog.value = true }

async function handleDeleteMatch(matchId: number) {
  if (!currentSession.value) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}/matches/${matchId}`, { method: 'DELETE' })
    const detail = await fetch(`/api/sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
  } catch (e: any) { showToast('删除失败') }
}

async function handleDeleteSession(sessionId: number) {
  try {
    await fetch(`/api/sessions/${sessionId}`, { method: 'DELETE' })
    await loadAll()
  } catch (e: any) { showToast('删除失败') }
}

async function handleForfeit(winnerId: number) {
  if (!scoringMatch.value || !currentSession.value) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}/matches/${scoringMatch.value.id}/forfeit`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ winner_id: winnerId }),
    })
    const detail = await fetch(`/api/sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
    showEditDialog.value = false
  } catch (e: any) { showToast('操作失败') }
}

async function handleScoreSubmit(scoreA: number, scoreB: number) {
  if (!scoringMatch.value || !currentSession.value) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}/matches/${scoringMatch.value.id}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ score_a: scoreA, score_b: scoreB }),
    })
    const detail = await fetch(`/api/sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail
    showEditDialog.value = false
  } catch (e: any) { showToast('提交失败') }
}

async function completeSession() {
  if (!currentSession.value) return
  try {
    await fetch(`/api/sessions/${currentSession.value.id}/complete`, { method: 'POST' })
    const detail = await fetch(`/api/sessions/${currentSession.value.id}`).then(r => r.json())
    currentSession.value = detail; step.value = 'result'
  } catch (e: any) { showToast('操作失败') }
}

async function viewSession(session: SessionDetail) {
  loadingSession.value = true
  try {
    const detail = await fetch(`/api/sessions/${session.id}`).then(r => r.json())
    currentSession.value = detail
    step.value = session.status === 'completed' ? 'result' : 'play'
  } catch (e) { showToast('加载失败') }
  finally { loadingSession.value = false }
}

function backToList() { step.value = 'list'; currentSession.value = null; loadAll() }
function myPlayers() { return selectedPlayers(players.value, selectedIDs.value) }
function playerIndex(pid: number): number {
  if (!currentSession.value) return pid
  const idx = currentSession.value.players.findIndex(p => p.id === pid)
  return idx >= 0 ? idx + 1 : pid
}
function matchIndex(mid: number): number {
  if (!currentSession.value) return mid
  const idx = currentSession.value.matches.findIndex(m => m.id === mid)
  return idx >= 0 ? idx + 1 : mid
}
</script>

<template>
  <div style="min-height: 100vh; background: #f0f2f5; padding-bottom: 80px;">

    <!-- Full-page loading overlay for session view -->
    <div v-if="loadingSession" style="position: fixed; inset: 0; background: rgba(255,255,255,0.9); z-index: 1000; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px;">
      <div style="width: 40px; height: 40px; border: 4px solid #ebedf0; border-top-color: #1989fa; border-radius: 50%; animation: spin 0.8s linear infinite;"></div>
      <span style="color: #969799; font-size: 14px; font-weight: 500;">加载活动中...</span>
    </div>

    <!-- Initial loading -->
    <div v-if="loadingPlayers" style="text-align: center; padding: 120px 20px;">
      <div style="width: 36px; height: 36px; border: 3px solid #ebedf0; border-top-color: #1989fa; border-radius: 50%; animation: spin 0.8s linear infinite; margin: 0 auto 12px;"></div>
      <span style="color: #969799;">加载中...</span>
    </div>

    <template v-else>
      <!-- Hero -->
      <div style="background: linear-gradient(135deg, #1a56db, #1e88e5, #00bfa5); color: #fff; padding: 24px 20px 20px;">
        <div style="font-size: 22px; font-weight: 700;">
          <IconTournament :size="26" :stroke-width="2" style="vertical-align: -5px; margin-right: 6px;" />
          活动
        </div>
        <div style="font-size: 13px; opacity: 0.8; margin-top: 4px;">
          {{ step === 'list' ? '创建或继续一场对局' : step === 'select' ? '选择参赛球员' : step === 'confirm' ? '确认对阵信息' : step === 'play' ? '逐场录入比分' : '最终排名' }}
        </div>
      </div>

      <!-- ===== LIST ===== -->
      <template v-if="step === 'list'">
        <div style="padding: 16px;">
          <button @click="enterSelect" style="width: 100%; padding: 16px; background: linear-gradient(135deg, #1989fa, #1e88e5); color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer; box-shadow: 0 4px 16px rgba(25,137,250,0.3);">
            <IconPlus :size="20" :stroke-width="2" style="vertical-align: -4px; margin-right: 4px;" />
            创建新活动
          </button>
        </div>

        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconList :size="18" :stroke-width="2" style="vertical-align: -3px;" />
          历史活动
        </div>

        <div v-if="sessions.length === 0" style="text-align: center; padding: 60px 20px; color: #969799;">
          <p>暂无活动记录</p>
        </div>

        <div v-for="s in sessions" :key="s.id" @click="viewSession(s)"
          style="background: #fff; border-radius: 12px; padding: 16px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); cursor: pointer;">
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <div style="font-weight: 600; font-size: 16px;">{{ s.name }}</div>
              <div style="font-size: 13px; color: #969799; margin-top: 2px;">{{ s.player_count || 0 }} 人 · {{ s.match_count || 0 }} 场</div>
            </div>
            <div style="display: flex; align-items: center; gap: 8px;">
              <span style="font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: 10px;"
                :style="s.status==='completed'?'background:#e8f8ef;color:#07c160;':'background:#e8f4ff;color:#1989fa;'">
                {{ s.status === 'completed' ? '已结束' : `剩${s.unplayed_count || 0}场` }}
              </span>
              <IconChevronRight :size="16" :stroke-width="2" style="color: #c8c9cc;" />
              <span @click.stop="handleDeleteSession(s.id)" style="font-size:12px;color:#c8c9cc;cursor:pointer;margin-left:4px;" title="删除活动">✕</span>
            </div>
          </div>
        </div>
      </template>

      <!-- ===== SELECT ===== -->
      <template v-if="step === 'select'">
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconUsers :size="18" :stroke-width="2" style="vertical-align: -3px;" />
          选择参赛球员（已选 {{ selectedIDs.size }} 人）
        </div>

        <div style="background: #fff; border-radius: 12px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-if="players.length === 0" style="text-align: center; padding: 40px; color: #969799;">
            <p style="margin-bottom: 12px;">还没有球员</p>
          </div>
          <div v-for="p in players" :key="p.id" @click="togglePlayer(p.id)"
            style="display: flex; align-items: center; padding: 14px 16px; border-bottom: 1px solid #f5f5f5; cursor: pointer;"
            :style="{ background: selectedIDs.has(p.id) ? '#e8f4ff' : '#fff' }">
            <input type="checkbox" :checked="selectedIDs.has(p.id)" style="width: 18px; height: 18px; margin-right: 12px; accent-color: #1989fa;" />
            <div style="flex: 1;">
              <div style="font-size: 16px; font-weight: 500;">{{ p.name }}</div>
              <div style="font-size: 13px; color: #969799;">{{ p.current_rating }} 分</div>
            </div>
          </div>
          <div @click="showAdd = true" style="display: flex; align-items: center; justify-content: center; padding: 14px; color: #1989fa; font-weight: 500; cursor: pointer;">
            <IconPlus :size="18" :stroke-width="2" style="vertical-align: -3px; margin-right: 4px;" />
            快速添加新球员
          </div>
        </div>

        <div style="padding: 16px;">
          <input v-model="sessionName" placeholder="活动名称（例：周五晚乒乓）"
            style="width: 100%; padding: 14px; border: 1px solid #ebedf0; border-radius: 12px; font-size: 15px; outline: none; margin-bottom: 16px; box-sizing: border-box;" />
          <button :disabled="selectedIDs.size < 2" @click="goConfirm"
            style="width: 100%; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;"
            :style="{ opacity: selectedIDs.size < 2 ? 0.5 : 1, cursor: selectedIDs.size < 2 ? 'not-allowed' : 'pointer' }">
            下一步（已选 {{ selectedIDs.size }} 人）
          </button>
          <div style="text-align: center; margin-top: 12px;">
            <button @click="step = 'list'" style="background: none; border: none; color: #969799; font-size: 14px; cursor: pointer;">返回列表</button>
          </div>
        </div>

        <!-- Quick add popup - simple overlay -->
        <div v-if="showAdd" style="position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 500; display: flex; align-items: flex-end;" @click.self="showAdd = false">
          <div style="background: #fff; border-radius: 16px 16px 0 0; padding: 24px 16px 80px; width: 100%;">
            <h3 style="text-align: center; margin-bottom: 20px; font-size: 18px;">快速添加球员</h3>
            <input v-model="newName" placeholder="姓名" style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 15px; margin-bottom: 8px; outline: none; box-sizing: border-box;" />
            <input v-model="newRating" type="number" placeholder="初始积分（默认 1500）" style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 15px; outline: none; box-sizing: border-box;" />
            <div style="display: flex; gap: 12px; margin-top: 20px;">
              <button @click="showAdd = false" style="flex: 1; padding: 14px; background: #f5f5f5; border: none; border-radius: 24px; font-size: 15px; cursor: pointer;">取消</button>
              <button @click="quickAddPlayer" :disabled="adding" style="flex: 2; padding: 14px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 15px; font-weight: 600; cursor: pointer;">
                {{ adding ? '添加中...' : '确认添加并选中' }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- ===== CONFIRM ===== -->
      <template v-if="step === 'confirm'">
        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconClipboardCheck :size="18" :stroke-width="2" style="vertical-align: -3px;" />
          确认对阵信息
        </div>

        <div style="background: #fff; border-radius: 12px; padding: 16px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">
          <div style="font-weight: 600; margin-bottom: 12px; font-size: 16px;">{{ sessionName }}</div>
          <div style="display: flex; flex-wrap: wrap; gap: 8px;">
            <span v-for="p in myPlayers()" :key="p.id" style="font-size: 14px; padding: 6px 12px; background: #e8f4ff; color: #1989fa; border-radius: 8px; font-weight: 500;">
              {{ p.name }} ({{ p.current_rating }})
            </span>
          </div>
          <div style="margin-top: 12px; font-size: 13px; color: #969799;">将自动生成循环赛对阵表（每人互相打一场）</div>
        </div>

        <div style="padding: 16px;">
          <button @click="createSession" :disabled="submitting"
            style="width: 100%; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;">
            {{ submitting ? '生成中...' : '生成对阵表' }}
          </button>
          <div style="text-align: center; margin-top: 12px;">
            <button @click="step = 'select'" style="background: none; border: none; color: #969799; font-size: 14px; cursor: pointer;">返回修改</button>
          </div>
        </div>
      </template>

      <!-- ===== PLAY ===== -->
      <template v-if="step === 'play' && currentSession">
        <SessionPlay
          :session="currentSession"
          :players="players"
          :show-edit-dialog="showEditDialog"
          :scoring-match="scoringMatch"
          :show-add-player-dialog="showAddPlayerDialog"
          :add-player-id="addPlayerId"
          @update:show-edit-dialog="showEditDialog = $event"
          @update:show-add-player-dialog="showAddPlayerDialog = $event"
          @update:add-player-id="addPlayerId = $event"
          @open-score-editor="openScoreEditor"
          @submit-score="handleScoreSubmit"
          @forfeit="handleForfeit"
          @add-player="addPlayerToSession"
          @complete-session="completeSession"
          @back-to-list="backToList"
          @edit-name-start="startEditName"
          @delete-match="handleDeleteMatch"
        />
        <!-- Name editing (in-play) -->
        <template v-if="editingName">
          <div style="padding:0 16px 8px;display:flex;gap:8px;justify-content:center;">
            <input v-model="editName" style="font-size:14px;font-weight:700;border:2px solid #1989fa;border-radius:8px;padding:6px 10px;outline:none;width:200px;text-align:center;" />
            <button @click="saveName" style="background:#1989fa;color:#fff;border:none;border-radius:6px;padding:6px 12px;font-size:13px;cursor:pointer;">保存</button>
            <button @click="editingName = false" style="background:#f5f5f5;border:none;border-radius:6px;padding:6px 12px;font-size:13px;cursor:pointer;">取消</button>
          </div>
        </template>
      </template>
      <!-- ===== RESULT ===== -->
      <template v-if="step === 'result' && currentSession">
        <div style="background: #fff; border-radius: 12px; padding: 20px; margin: 12px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">
          <div style="display: flex; align-items: center; justify-content: center; gap: 8px;">
            <IconTrophy :size="22" :stroke-width="2" style="color: #f5a623;" />
            <template v-if="editingName">
              <input v-model="editName" style="font-size: 16px; font-weight: 700; border: 2px solid #1989fa; border-radius: 8px; padding: 6px 10px; outline: none; width: 200px; text-align: center;" />
              <button @click="saveName" style="background: #1989fa; color: #fff; border: none; border-radius: 6px; padding: 6px 12px; font-size: 13px; cursor: pointer;">保存</button>
              <button @click="editingName = false" style="background: #f5f5f5; border: none; border-radius: 6px; padding: 6px 12px; font-size: 13px; cursor: pointer;">取消</button>
            </template>
            <template v-else>
              <span style="font-weight: 700; font-size: 20px;">{{ currentSession.name }}</span>
              <span @click="startEditName" style="cursor: pointer; color: #c8c9cc; font-size: 16px;">&#9998;</span>
            </template>
          </div>
          <div style="font-size: 13px; color: #969799; margin-top: 4px; text-align: center;">最终排名</div>
        </div>

        <div style="background: #fff; border-radius: 12px; margin: 8px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-for="(p, i) in currentSession.players" :key="p.id"
            style="display: flex; align-items: center; padding: 14px 16px; border-bottom: 1px solid #f5f5f5;"
            :style="{ background: i === 0 ? '#fffdf0' : i === 1 ? '#f8f9fa' : i === 2 ? '#fdf3f0' : '' }">
            <div style="width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 800;"
              :style="{ background: i===0?'#fff3cd':i===1?'#e8e8e8':i===2?'#ffe8d6':'#f0f2f5', color: i===0?'#b8860b':i===1?'#666':i===2?'#b87333':'#969799' }">
              {{ i + 1 }}
            </div>
            <div style="flex: 1; margin-left: 12px;">
              <div style="font-size: 16px; font-weight: 500;" :style="{ fontSize: i < 3 ? '18px' : '16px', fontWeight: i < 3 ? 700 : 500 }">{{ p.name }} <span style="font-size:11px;color:#c8c9cc;">#{{ i + 1 }}</span></div>
              <div style="font-size: 12px; color: #969799;">{{ p.wins }}胜 {{ p.losses }}负 <template v-if="p.forfeits">· 弃权{{ p.forfeits }}</template></div>
            </div>
            <div style="text-align: right;">
              <div style="font-size: 18px; font-weight: 700; color: #1989fa;">{{ sessionDisplayRating(p, currentSession?.matches) }}</div>
              <div style="font-size: 11px;">
                <span style="color: #969799;">{{ p.starting_rating }}</span>
                <span :style="{ color: sessionChange(currentSession!.matches, p.id) >= 0 ? '#07c160' : '#ee0a24', fontWeight: 600 }">
                  {{ changeSign(sessionChange(currentSession!.matches, p.id)) }}{{ sessionChange(currentSession!.matches, p.id) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div style="font-size: 16px; font-weight: 600; padding: 16px 16px 8px; display: flex; align-items: center; gap: 6px;">
          <IconList :size="18" :stroke-width="2" style="vertical-align: -3px;" />
          全部对阵
        </div>
        <div style="background: #fff; border-radius: 12px; margin: 4px 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06); overflow: hidden;">
          <div v-for="m in currentSession.matches" :key="m.id"
            style="display: flex; align-items: center; padding: 14px 16px; border-bottom: 1px solid #f5f5f5;">
            <span style="width:32px;font-size:12px;color:#c8c9cc;flex-shrink:0;">#{{ matchIndex(m.id) }}</span>
            <div style="flex: 1; text-align: right; font-weight: 400;" :style="{ fontWeight: m.winner_id === m.player_a_id ? 700 : 400 }">
              <span style="font-size:10px;color:#c8c9cc;">{{ playerIndex(m.player_a_id) }}号</span> {{ m.player_a_name }}
              <span style="font-size:10px;display:block;" :style="{ color: m.rating_change_a >= 0 ? '#07c160' : '#ee0a24' }">{{ changeSign(m.rating_change_a) }}{{ m.rating_change_a }}</span>
            </div>
            <div style="width: 64px; text-align: center; font-weight: 700; font-size: 16px;"><template v-if="m.forfeit"><span style="color:#ff976a;font-weight:600;">弃权</span></template><template v-else>{{ m.score_a }} : {{ m.score_b }}</template></div>
            <div style="flex: 1; font-weight: 400;" :style="{ fontWeight: m.winner_id === m.player_b_id ? 700 : 400 }">
              <span style="font-size:10px;color:#c8c9cc;">{{ playerIndex(m.player_b_id) }}号</span> {{ m.player_b_name }}
              <span style="font-size:10px;display:block;" :style="{ color: m.rating_change_b >= 0 ? '#07c160' : '#ee0a24' }">{{ changeSign(m.rating_change_b) }}{{ m.rating_change_b }}</span>
            </div>
          </div>
        </div>

        <div style="padding: 16px;">
          <button @click="backToList" style="width: 100%; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 17px; font-weight: 600; cursor: pointer;">
            返回活动列表
          </button>
        </div>
      </template>
    </template>
  </div>
</template>

