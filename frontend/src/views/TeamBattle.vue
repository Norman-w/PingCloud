<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import { IconTournament, IconPlus, IconList, IconUsers, IconRefresh } from '@tabler/icons-vue'
import { api, type Player } from '../api'

interface TBSummary { id:number;name:string;mode:string;group_a_name:string;group_b_name:string;status:string;a_wins:number;b_wins:number;created_at:string }
interface TBPlayer { id:number;name:string;current_rating:number;team:string }
interface TBMatch { id:number;match_type:string;a1_id:number;a2_id:number|null;b1_id:number;b2_id:number|null;a1_name:string;a2_name:string;b1_name:string;b2_name:string;score_a:number|null;score_b:number|null;winner_team:string;played:boolean }

const players = ref<Player[]>([])
const battles = ref<TBSummary[]>([])
const step = ref<'list'|'select'|'group'|'play'>('list')
const selectedIDs = ref<Set<number>>(new Set())
const groupA = ref<Set<number>>(new Set())
const groupB = ref<Set<number>>(new Set())
const sessionName = ref('')
const loading = ref(true)

const current = ref<{id:number;name:string;mode:string;group_a_name:string;group_b_name:string;status:string;a_wins:number;b_wins:number;players:TBPlayer[];matches:TBMatch[]}|null>(null)

// Score dialog
const scoreDialog = ref(false)
const scoreMatch = ref<TBMatch|null>(null)
const scoreA = ref('')
const scoreB = ref('')

onMounted(async () => {
  try {
    const [p, b] = await Promise.all([
      api.getPlayers(),
      fetch('/api/team-battles').then(r => r.json()).catch(() => []),
    ])
    players.value = p; battles.value = b
  } catch {}
  loading.value = false
})

function enterSelect() { selectedIDs.value = new Set(); sessionName.value = ''; step.value = 'select' }
function togglePlayer(id:number) { const s = new Set(selectedIDs.value); s.has(id)?s.delete(id):s.add(id); selectedIDs.value = s }

function doShuffle() {
  const all = Array.from(selectedIDs.value)
  const m:number[]=[]; const f:number[]=[]
  for (const id of all) { const p = players.value.find(x=>x.id===id); p?.gender==='female'?f.push(id):m.push(id) }
  const sh = (a:number[]) => { for(let i=a.length-1;i>0;i--){const j=Math.floor(Math.random()*(i+1));[a[i],a[j]]=[a[j],a[i]]} }; sh(m); sh(f)
  const ga=new Set<number>(); const gb=new Set<number>()
  for(const x of m){(ga.size<=gb.size?ga:gb).add(x)}
  for(const x of f){(ga.size<=gb.size?ga:gb).add(x)}
  groupA.value = ga; groupB.value = gb
}

function goGroup() { if(selectedIDs.value.size<2){showToast('至少选2人');return}; doShuffle(); step.value = 'group' }

async function createBattle() {
  try {
    const res = await fetch('/api/team-battles', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body:JSON.stringify({
        name:sessionName.value.trim()||'团体对抗赛',
        group_a_player_ids:Array.from(groupA.value),
        group_b_player_ids:Array.from(groupB.value),
        singles_on:true, doubles_on:true,
      }),
    })
    if (!res.ok) { const t = await res.text(); showToast(t); return }
    const {id} = await res.json()
    const d = await fetch(`/api/team-battles/${id}`).then(r=>r.json())
    current.value = d; step.value = 'play'
  } catch(e:any) { showToast('创建失败') }
}

async function loadBattle(b:TBSummary) {
  try {
    current.value = await fetch(`/api/team-battles/${b.id}`).then(r=>r.json())
    step.value = 'play'
  } catch { showToast('加载失败') }
}

function openScoreDialog(m:TBMatch) {
  scoreMatch.value = m
  scoreA.value = m.played && m.score_a != null ? String(m.score_a) : ''
  scoreB.value = m.played && m.score_b != null ? String(m.score_b) : ''
  scoreDialog.value = true
}

async function submitScore() {
  if (!scoreMatch.value || !current.value) return
  const a = Number(scoreA.value); const b = Number(scoreB.value)
  if (isNaN(a) || isNaN(b)) { showToast('请输入有效比分'); return }
  await fetch(`/api/team-battles/${current.value.id}/matches/${scoreMatch.value.id}`, {
    method:'POST', headers:{'Content-Type':'application/json'},
    body:JSON.stringify({score_a:a, score_b:b}),
  })
  current.value = await fetch(`/api/team-battles/${current.value.id}`).then(r=>r.json())
  scoreDialog.value = false
}

async function forfeitMatch(winner: string) {
  if (!scoreMatch.value || !current.value) return
  await fetch(`/api/team-battles/${current.value.id}/matches/${scoreMatch.value.id}/forfeit`, {
    method:'POST', headers:{'Content-Type':'application/json'},
    body:JSON.stringify({winner_team:winner}),
  })
  current.value = await fetch(`/api/team-battles/${current.value.id}`).then(r=>r.json())
  scoreDialog.value = false
}

async function completeBattle() {
  if (!current.value) return
  try {
    const res = await fetch(`/api/team-battles/${current.value.id}/complete`, {method:'POST'})
    if (!res.ok) { const t = await res.text(); showToast(t); return }
    current.value = await fetch(`/api/team-battles/${current.value.id}`).then(r=>r.json())
  } catch { showToast('操作失败') }
}

async function deleteBattle() {
  if (!current.value) return
  if (!confirm('确定删除这场团体赛吗？')) return
  await fetch(`/api/team-battles/${current.value.id}`, {method:'DELETE'})
  backToList()
}

async function backToList() {
  step.value = 'list'; current.value = null
  try { battles.value = await fetch('/api/team-battles').then(r=>r.json()) } catch {}
}

function matchLabel(m:TBMatch):string {
  if (m.match_type==='doubles') return `${m.a1_name}${m.a2_name?'+'+m.a2_name:''} vs ${m.b1_name}${m.b2_name?'+'+m.b2_name:''}`
  return `${m.a1_name} vs ${m.b1_name}`
}

const unplayedCount = computed(() => current.value?.matches.filter(m => !m.played).length ?? 0)
const allPlayed = computed(() => unplayedCount.value === 0)
</script>

<template>
  <div style="min-height:100vh;background:#f0f2f5;padding-bottom:80px;">
    <div style="background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;padding:24px 20px 20px;">
      <div style="font-size:22px;font-weight:700;"><IconTournament :size="26" style="vertical-align:-5px;margin-right:6px;" />团体对抗赛</div>
      <div style="font-size:13px;opacity:0.8;margin-top:4px;">{{ step==='list'?'A组 vs B组 双单交替':step==='select'?'选择参赛选手':'抽签分组' }}</div>
    </div>

    <template v-if="loading"><div style="text-align:center;padding:60px;color:#999;">加载中...</div></template>

    <!-- LIST -->
    <template v-if="step==='list' && !loading">
      <div style="padding:16px;"><button @click="enterSelect" style="width:100%;padding:16px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:24px;font-size:17px;font-weight:600;cursor:pointer;box-shadow:0 4px 16px rgba(102,126,234,0.3);"><IconPlus :size="20" style="vertical-align:-4px;margin-right:4px;" />创建团体对抗赛</button></div>
      <div style="padding:8px 16px;font-size:14px;font-weight:600;color:#666;"><IconList :size="16" style="vertical-align:-3px;margin-right:4px;" />历史记录</div>
      <div v-if="battles.length===0" style="text-align:center;padding:60px 20px;color:#969799;">暂无团体赛记录</div>
      <div v-for="b in battles" :key="b.id" @click="loadBattle(b)" style="background:#fff;border-radius:12px;padding:16px;margin:8px 16px;box-shadow:0 2px 8px rgba(0,0,0,0.06);cursor:pointer;">
        <div style="font-weight:600;font-size:16px;">{{ b.name }}</div>
        <div style="font-size:13px;color:#969799;margin-top:4px;">{{ b.group_a_name }} {{ b.a_wins }}:{{ b.b_wins }} {{ b.group_b_name }} · {{ b.status==='completed'?'已结束':'进行中' }}</div>
      </div>
    </template>

    <!-- SELECT -->
    <template v-if="step==='select'">
      <div style="font-size:16px;font-weight:600;padding:16px 16px 8px;"><IconUsers :size="18" style="vertical-align:-3px;color:#1989fa;" /> 选择参赛选手（已选 {{ selectedIDs.size }} 人）</div>
      <div style="background:#fff;border-radius:12px;margin:4px 16px;box-shadow:0 2px 8px rgba(0,0,0,0.06);overflow:hidden;">
        <div v-for="p in players" :key="p.id" @click="togglePlayer(p.id)" style="display:flex;align-items:center;padding:12px 16px;border-bottom:1px solid #f5f5f5;cursor:pointer;" :style="{background:selectedIDs.has(p.id)?'#e8f4ff':'#fff'}">
          <input type="checkbox" :checked="selectedIDs.has(p.id)" style="width:18px;height:18px;margin-right:12px;accent-color:#1989fa;" />
          <div><div style="font-size:16px;font-weight:500;">{{ p.name }}</div><div style="font-size:13px;color:#969799;">{{ p.current_rating }}分 · {{ p.gender==='female'?'女':'男' }}</div></div>
        </div>
      </div>
      <div style="padding:16px;">
        <input v-model="sessionName" placeholder="比赛名称（默认：团体对抗赛）" style="width:100%;padding:14px;border:1px solid #ebedf0;border-radius:12px;font-size:15px;outline:none;margin-bottom:16px;box-sizing:border-box;" />
        <button :disabled="selectedIDs.size<2" @click="goGroup" style="width:100%;padding:16px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:24px;font-size:17px;font-weight:600;cursor:pointer;" :style="{opacity:selectedIDs.size<2?0.5:1}">下一步 · 抽签分组</button>
        <div style="text-align:center;margin-top:12px;"><button @click="step='list'" style="background:none;border:none;color:#969799;font-size:14px;cursor:pointer;">返回列表</button></div>
      </div>
    </template>

    <!-- GROUP -->
    <template v-if="step==='group'">
      <div style="text-align:center;font-size:16px;font-weight:600;padding:16px;">自动分组（按性别均匀分配）</div>
      <div style="display:flex;gap:8px;margin:8px 16px;">
        <div style="flex:1;background:#fff;border-radius:12px;padding:12px;box-shadow:0 2px 6px rgba(0,0,0,0.06);">
          <div style="font-weight:700;color:#1989fa;margin-bottom:8px;text-align:center;">A组 ({{ groupA.size }}人)</div>
          <div v-for="id in Array.from(groupA)" :key="id" style="padding:6px 8px;font-size:13px;border-bottom:1px solid #f5f5f5;">{{ players.find(p=>p.id===id)?.name }}</div>
        </div>
        <div style="flex:1;background:#fff;border-radius:12px;padding:12px;box-shadow:0 2px 6px rgba(0,0,0,0.06);">
          <div style="font-weight:700;color:#ee0a24;margin-bottom:8px;text-align:center;">B组 ({{ groupB.size }}人)</div>
          <div v-for="id in Array.from(groupB)" :key="id" style="padding:6px 8px;font-size:13px;border-bottom:1px solid #f5f5f5;">{{ players.find(p=>p.id===id)?.name }}</div>
        </div>
      </div>
      <div style="padding:0 16px;display:flex;gap:8px;">
        <button @click="doShuffle" style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:24px;font-size:14px;cursor:pointer;"><IconRefresh :size="16" style="vertical-align:-3px;" /> 重新抽签</button>
        <button @click="step='select'" style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:24px;font-size:14px;cursor:pointer;">返回修改</button>
      </div>
      <div style="padding:16px;"><button @click="createBattle" style="width:100%;padding:16px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:24px;font-size:17px;font-weight:600;cursor:pointer;">确认分组 · 生成对阵</button></div>
    </template>

    <!-- PLAY -->
    <template v-if="step==='play' && current">
      <div style="background:#fff;border-radius:12px;padding:16px;margin:12px 16px;text-align:center;box-shadow:0 2px 8px rgba(0,0,0,0.06);">
        <div style="font-weight:700;font-size:18px;">{{ current.name }}</div>
        <div style="display:flex;justify-content:center;gap:20px;margin-top:12px;">
          <div><div style="font-size:12px;color:#969799;">{{ current.group_a_name }}</div><div style="font-size:32px;font-weight:800;color:#1989fa;">{{ current.a_wins }}</div></div>
          <div style="font-size:24px;font-weight:800;color:#c8c9cc;">:</div>
          <div><div style="font-size:12px;color:#969799;">{{ current.group_b_name }}</div><div style="font-size:32px;font-weight:800;color:#ee0a24;">{{ current.b_wins }}</div></div>
        </div>
        <div v-if="current.status==='completed'" style="font-size:13px;color:#07c160;font-weight:600;margin-top:4px;">已结束</div>
      </div>

      <!-- Players by team -->
      <div style="display:flex;gap:8px;margin:4px 16px 12px;">
        <div style="flex:1;background:#fff;border-radius:10px;padding:10px;box-shadow:0 1px 4px rgba(0,0,0,0.04);">
          <div style="font-size:12px;font-weight:700;color:#1989fa;margin-bottom:6px;">{{ current.group_a_name }}</div>
          <div v-for="p in current.players.filter(x=>x.team==='A')" :key="p.id" style="font-size:12px;color:#666;padding:2px 0;">{{ p.name }}</div>
        </div>
        <div style="flex:1;background:#fff;border-radius:10px;padding:10px;box-shadow:0 1px 4px rgba(0,0,0,0.04);">
          <div style="font-size:12px;font-weight:700;color:#ee0a24;margin-bottom:6px;">{{ current.group_b_name }}</div>
          <div v-for="p in current.players.filter(x=>x.team==='B')" :key="p.id" style="font-size:12px;color:#666;padding:2px 0;">{{ p.name }}</div>
        </div>
      </div>

      <!-- Match list -->
      <div style="font-size:16px;font-weight:600;padding:8px 16px;">对阵表（剩{{ unplayedCount }}场）</div>
      <div style="background:#fff;border-radius:12px;margin:4px 16px;box-shadow:0 2px 8px rgba(0,0,0,0.06);overflow:hidden;">
        <div v-for="(m,i) in current.matches" :key="m.id" @click="current.status!=='completed' && openScoreDialog(m)" style="padding:12px 16px;border-bottom:1px solid #f5f5f5;display:flex;align-items:center;gap:8px;" :style="{cursor: current.status==='completed'?'default':'pointer'}">
          <span style="font-size:10px;padding:2px 6px;border-radius:6px;font-weight:600;" :style="{background:m.match_type==='doubles'?'#e8f4ff':'#f0f2f5',color:m.match_type==='doubles'?'#1989fa':'#969799'}">{{ m.match_type==='doubles'?'双打':'单打' }}</span>
          <span style="font-size:12px;color:#c8c9cc;">#{{ i+1 }}</span>
          <div style="flex:1;text-align:center;font-size:13px;">{{ matchLabel(m) }}</div>
          <span v-if="m.played" style="font-weight:700;font-size:14px;" :style="{color:m.winner_team==='A'?'#1989fa':'#ee0a24'}">{{ m.score_a }}:{{ m.score_b }}</span>
          <span v-else style="color:#c8c9cc;font-size:12px;">待录入</span>
        </div>
      </div>

      <div style="padding:16px;display:flex;gap:12px;" v-if="current.status!=='completed'">
        <button @click="completeBattle" :disabled="!allPlayed" style="flex:1;padding:16px;background:#07c160;color:#fff;border:none;border-radius:24px;font-size:16px;font-weight:600;cursor:pointer;" :style="{opacity:allPlayed?1:0.5}">结束比赛</button>
        <button @click="deleteBattle" style="padding:16px 24px;background:#fff;border:1px solid #ee0a24;border-radius:24px;font-size:14px;cursor:pointer;color:#ee0a24;">删除</button>
      </div>
      <div style="padding:16px;">
        <button @click="backToList" style="width:100%;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:16px;font-weight:600;cursor:pointer;">返回列表</button>
      </div>
    </template>

    <!-- Score Dialog -->
    <div v-if="scoreDialog" style="position:fixed;inset:0;background:rgba(0,0,0,0.5);z-index:2000;display:flex;align-items:center;justify-content:center;" @click.self="scoreDialog=false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:300px;">
        <h3 style="text-align:center;margin:0 0 4px;">录入比分</h3>
        <div v-if="scoreMatch" style="text-align:center;font-size:14px;color:#666;margin-bottom:16px;">
          {{ matchLabel(scoreMatch) }}
          <span style="font-size:11px;padding:2px 6px;border-radius:6px;font-weight:600;margin-left:4px;" :style="{background:scoreMatch.match_type==='doubles'?'#e8f4ff':'#f0f2f5',color:scoreMatch.match_type==='doubles'?'#1989fa':'#969799'}">{{ scoreMatch.match_type==='doubles'?'双打':'单打' }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:20px;">
          <div style="flex:1;text-align:center;">
            <div style="font-size:12px;color:#1989fa;font-weight:600;margin-bottom:4px;">{{ current?.group_a_name }}</div>
            <input v-model="scoreA" type="number" inputmode="numeric" style="width:80px;padding:12px;border:2px solid #1989fa;border-radius:12px;font-size:24px;font-weight:700;text-align:center;outline:none;" />
          </div>
          <div style="font-size:20px;font-weight:800;color:#c8c9cc;">:</div>
          <div style="flex:1;text-align:center;">
            <div style="font-size:12px;color:#ee0a24;font-weight:600;margin-bottom:4px;">{{ current?.group_b_name }}</div>
            <input v-model="scoreB" type="number" inputmode="numeric" style="width:80px;padding:12px;border:2px solid #ee0a24;border-radius:12px;font-size:24px;font-weight:700;text-align:center;outline:none;" />
          </div>
        </div>
        <div style="display:flex;gap:8px;margin-bottom:8px;">
          <button @click="scoreDialog=false" style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:12px;font-size:15px;cursor:pointer;">取消</button>
          <button @click="submitScore" style="flex:1;padding:12px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:12px;font-size:15px;font-weight:600;cursor:pointer;">确认</button>
        </div>
        <div style="display:flex;gap:8px;">
          <button @click="forfeitMatch('A')" style="flex:1;padding:10px;background:#fff;border:1.5px solid #ff976a;border-radius:10px;color:#ff976a;font-size:13px;font-weight:600;cursor:pointer;">{{ current?.group_b_name }} 弃权</button>
          <button @click="forfeitMatch('B')" style="flex:1;padding:10px;background:#fff;border:1.5px solid #ff976a;border-radius:10px;color:#ff976a;font-size:13px;font-weight:600;cursor:pointer;">{{ current?.group_a_name }} 弃权</button>
        </div>
      </div>
    </div>
  </div>
</template>
