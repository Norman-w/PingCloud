<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { IconUser, IconFlame, IconCrown, IconTarget, IconCircle, IconRun, IconFolder, IconFolderOpen, IconChevronRight } from '@tabler/icons-vue'
import { myId, myName, checkAuth, logout as authLogout } from '../auth'
import LoginModal from '../components/LoginModal.vue'

const router = useRouter()
const showLogin = ref(false)
const authReady = ref(false)

// ── training stats ──
interface TrainingStats { total_sessions: number; total_minutes: number; this_month_sessions: number; skill_frequencies: { skill_id: number; skill_name: string; category: string; count: number }[] }
const stats = ref<TrainingStats>({ total_sessions: 0, total_minutes: 0, this_month_sessions: 0, skill_frequencies: [] })

// ── skill mastery ──
interface SkillItem { id: number; name: string; category: string; tags: string[]; practice_count: number; total_duration_minutes: number; last_practiced: string; status: string }
interface SkillGroup { label: string; skills: SkillItem[] }
const stages = ref<SkillGroup[]>([]); const tagFilters = ref<SkillGroup[]>([])
const activeTags = ref<string[]>([]); const collapsedStages = ref<Set<string>>(new Set())

const totalHours = computed(() => { const h=Math.floor(stats.value.total_minutes/3600); const m=Math.floor((stats.value.total_minutes%3600)/60); if(h>0)return`${h}h${m}m`; return`${m}m` })
function allMastered(stage: SkillGroup) { return stage.skills.length>0 && stage.skills.every(s => s.status==='mastered') }
function stageMasteredCount(stage: SkillGroup) { return stage.skills.filter(s => s.status==='mastered').length }
function isCollapsed(stage: SkillGroup) { return collapsedStages.value.has(stage.label) }
function toggleStage(label: string) { const s=new Set(collapsedStages.value); s.has(label)?s.delete(label):s.add(label); collapsedStages.value=s }
function filteredSkills(skills: SkillItem[]): SkillItem[] { return activeTags.value.length===0?skills:skills.filter(item => activeTags.value.every(t => item.tags.includes(t))) }
function groupTag(label: string): string { const m: Record<string,string>={'正手技术':'正手','反手技术':'反手','进攻':'进攻','防守':'防守','上旋球':'上旋','下旋球':'下旋','短球':'短球','长球':'长球'}; return m[label]||label }
function setTag(label: string) { const t=groupTag(label); const idx=activeTags.value.indexOf(t); if(idx>=0)activeTags.value.splice(idx,1); else activeTags.value.push(t) }

// Sort stages: those with matching filtered skills first
const sortedStages = computed(() => {
  return [...stages.value].sort((a, b) => {
    const aMatch = filteredSkills(a.skills).length
    const bMatch = filteredSkills(b.skills).length
    if (aMatch > 0 && bMatch === 0) return -1
    if (bMatch > 0 && aMatch === 0) return 1
    return 0
  })
})

onMounted(async () => {
  await checkAuth()
  authReady.value = true
  await loadAll()
})

async function logout() { authLogout(); await loadAll() }
async function loadAll() {
  try { const r=await fetch('/api/training-stats'); if(r.ok)stats.value=await r.json() } catch {}
  try { const r=await fetch('/api/skill-mastery'); if(r.ok){const d=await r.json();stages.value=d.stages||[];tagFilters.value=d.tagFilters||[];const c=new Set<string>();for(const s of stages.value){if(allMastered(s))c.add(s.label)};collapsedStages.value=c} } catch {}
}

function statusStyle(status: string) {
  switch(status){case'mastered':return{bg:'linear-gradient(135deg,#e8f8ef,#d4f5e0)',border:'#07c160',text:'#07c160',badge:'🏆',badgeBg:'#07c160'};case'practicing':return{bg:'linear-gradient(135deg,#fff8e1,#fff3cd)',border:'#f5a623',text:'#b8860b',badge:'🔥',badgeBg:'#f5a623'};default:return{bg:'#f8f9fa',border:'#e0e0e0',text:'#999',badge:'',badgeBg:'#aaa'}}
}

const tagIcons: Record<string,string> = {
  '正手':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>`,
  '反手':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" transform="scale(-1,1) translate(-24,0)"/></svg>`,
  '进攻':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M8 16V8a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"/><line x1="14" y1="7" x2="20" y2="7"/><line x1="14" y1="11" x2="20" y2="11"/></svg>`,
  '防守':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2l8 5v6c0 5.5-3.8 10.7-8 12-4.2-1.3-8-6.5-8-12V7l8-5z"/></svg>`,
  '上旋':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><ellipse cx="6" cy="13" rx="4" ry="5"/><line x1="9" y1="9" x2="14" y2="4"/></svg>`,
  '下旋':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><ellipse cx="7" cy="12" rx="4" ry="5"/><line x1="11" y1="13" x2="18" y2="10"/></svg>`,
  '短球':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="8"/><circle cx="12" cy="12" r="3"/></svg>`,
  '长球':`<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="8"/><line x1="12" y1="4" x2="12" y2="8"/><line x1="12" y1="16" x2="12" y2="20"/></svg>`,
}
function tagColor(t: string) { const m: Record<string,string>={'正手':'#1989fa','反手':'#e67e22','进攻':'#ee0a24','防守':'#07c160','上旋':'#9b59b6','下旋':'#3498db','短球':'#1abc9c','长球':'#e74c3c','左侧旋':'#f39c12','右侧旋':'#2ecc71'}; return m[t]||'#666' }
</script>

<template>
  <div class="safe-bottom">
    <!-- Loading skeleton -->
    <div v-if="!authReady" style="text-align:center;padding:120px 20px;">
      <div style="width:36px;height:36px;border:3px solid #ebedf0;border-top-color:#1989fa;border-radius:50%;animation:spin 0.8s linear infinite;margin:0 auto 16px;"></div>
      <div style="color:#969799;font-size:14px;">加载中...</div>
    </div>
    <template v-else>
    <!-- Header -->
    <div class="hero" style="padding-bottom:24px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <div>
          <div class="hero-title"><IconUser :size="28" :stroke-width="2" style="vertical-align:-5px;margin-right:4px;" />{{ myName || '我的' }}</div>
          <div class="hero-sub">{{ myName ? '个人训练中心' : '记录每次练功，见证每步成长' }}</div>
        </div>
        <button v-if="!myName" @click="showLogin=true"
          style="background:rgba(255,255,255,0.25);backdrop-filter:blur(8px);border:1.5px solid rgba(255,255,255,0.4);color:#fff;padding:10px 20px;border-radius:24px;font-size:15px;font-weight:700;cursor:pointer;white-space:nowrap;transition:all .2s;box-shadow:0 2px 8px rgba(0,0,0,0.1);">🔑 登录</button>
        <button v-else @click="logout"
          style="background:rgba(255,255,255,0.15);backdrop-filter:blur(8px);border:1.5px solid rgba(255,255,255,0.3);color:#fff;padding:8px 16px;border-radius:20px;font-size:13px;font-weight:600;cursor:pointer;transition:all .2s;">退出</button>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row">
      <div class="stat-card"><div class="stat-value">{{ stats.total_sessions }}</div><div class="stat-label">累计练功</div></div>
      <div class="stat-card"><div class="stat-value">{{ totalHours }}</div><div class="stat-label">总时长</div></div>
      <div class="stat-card"><div class="stat-value">{{ stats.this_month_sessions }}</div><div class="stat-label">本月次数</div></div>
    </div>

    <!-- Skill frequency -->
    <div v-if="stats.skill_frequencies&&stats.skill_frequencies.length>0" style="margin:0 16px 12px;">
      <div class="section-title" style="padding-left:0;"><IconFlame :size="18" :stroke-width="2" style="vertical-align:-3px;margin-right:6px;" />常练技能</div>
      <div class="card" style="margin:0;display:flex;flex-wrap:wrap;gap:8px;">
        <span v-for="(sk,i) in stats.skill_frequencies.slice(0,10)" :key="sk.skill_id" style="font-size:13px;padding:4px 10px;border-radius:6px;font-weight:500;" :style="{background:i<3?'#e8f4ff':'#f5f5f5',color:i<3?'#1989fa':'#646566'}">{{ sk.skill_name }} ×{{ sk.count }}</span>
      </div>
    </div>

    <!-- Skill Mastery -->
    <div v-if="stages.length>0" style="margin:0 16px 12px;">
      <div class="section-title" style="padding-left:0;justify-content:space-between;"><span><IconTarget :size="18" :stroke-width="2" style="vertical-align:-3px;margin-right:6px;" />技能掌握</span></div>

      <!-- Attribute filter chips -->
      <div style="display:flex;flex-wrap:wrap;gap:6px;margin-bottom:16px;">
        <span v-for="g in tagFilters" :key="g.label" @click="setTag(g.label)" style="font-size:12px;padding:4px 10px;border-radius:12px;cursor:pointer;font-weight:500;transition:all .15s;white-space:nowrap;" :style="activeTags.includes(groupTag(g.label))?'background:#1989fa;color:#fff;':'background:#f0f2f5;color:#646566;'">{{ g.label }} · {{ g.skills.length }}</span>
        <span v-if="activeTags.length>0" @click="activeTags=[]" style="font-size:12px;padding:4px 10px;border-radius:12px;cursor:pointer;background:#fde8e8;color:#ee0a24;font-weight:500;">✕ 清除</span>
      </div>

      <!-- Stage sections -->
      <div v-for="stage in sortedStages" :key="stage.label" style="margin-bottom:16px;">
        <div @click="toggleStage(stage.label)" style="display:flex;align-items:center;justify-content:space-between;padding:10px 14px;border-radius:12px;cursor:pointer;margin-bottom:6px;transition:all .15s;" :style="allMastered(stage)?'background:#e8f8ef;':'background:#f8f9fa;'">
          <div style="display:flex;align-items:center;gap:8px;">
            <IconFolderOpen v-if="!isCollapsed(stage)" :size="18" :stroke-width="2" :style="{color:allMastered(stage)?'#07c160':'#1989fa'}" />
            <IconFolder v-else :size="18" :stroke-width="2" :style="{color:allMastered(stage)?'#07c160':'#1989fa'}" />
            <span style="font-weight:700;font-size:15px;" :style="{color:allMastered(stage)?'#07c160':'#333'}">{{ stage.label }}</span>
            <span style="font-size:12px;color:#999;">{{ stageMasteredCount(stage) }}/{{ stage.skills.length }}</span>
          </div>
          <div style="display:flex;align-items:center;gap:6px;">
            <IconCrown v-if="allMastered(stage)" :size="16" style="color:#f5a623;" />
            <IconChevronRight :size="16" style="color:#ccc;" :style="{transform:isCollapsed(stage)?'rotate(0deg)':'rotate(90deg)',transition:'transform .2s'}" />
          </div>
        </div>
        <div v-if="!isCollapsed(stage)" style="display:grid;grid-template-columns:repeat(auto-fill,minmax(140px,1fr));gap:8px;">
          <div v-for="item in filteredSkills(stage.skills)" :key="item.id" @click="router.push({name:'SkillTrain',params:{id:item.id}})"
            style="border-radius:12px;padding:12px 10px;cursor:pointer;transition:all .2s;position:relative;overflow:hidden;"
            :style="{background:statusStyle(item.status).bg,border:'2px solid '+statusStyle(item.status).border,opacity:item.status==='none'?0.5:1,filter:item.status==='none'?'grayscale(0.7)':'none',}">
            <div v-if="statusStyle(item.status).badge" style="position:absolute;top:8px;right:8px;font-size:14px;">{{ statusStyle(item.status).badge }}</div>
            <div style="font-weight:700;font-size:14px;margin-bottom:4px;padding-right:28px;" :style="{color:statusStyle(item.status).text}">{{ item.name }}</div>
            <div style="display:flex;flex-wrap:wrap;gap:3px;margin-bottom:6px;">
              <span v-for="t in item.tags.slice(0,3)" :key="t" style="display:flex;align-items:center;gap:2px;padding:1px 5px;border-radius:4px;color:#fff;font-size:10px;font-weight:500;" :style="{background:tagColor(t)}"><span v-html="tagIcons[t]" style="display:flex;align-items:center;"></span>{{ t }}</span>
            </div>
            <div style="font-size:10px;color:#999;">
              <template v-if="item.practice_count>0">{{ item.practice_count }}次 · {{ Math.floor(item.total_duration_minutes/3600) }}h{{ Math.floor((item.total_duration_minutes%3600)/60) }}m</template>
            </div>
            <div v-if="item.practice_count>0" style="margin-top:6px;height:2px;background:#e0e0e0;border-radius:1px;overflow:hidden;"><div style="height:100%;border-radius:1px;" :style="{width:Math.min(100,item.practice_count*20)+'%',background:statusStyle(item.status).border}"></div></div>
          </div>
        </div>
      </div>
    </div>
    <div style="height:8px;"></div>
    </template>
  </div>
  <LoginModal v-model:visible="showLogin" />
</template>
