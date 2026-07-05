<script setup lang="ts">
import { computed } from 'vue'
import { IconSwords, IconFlame, IconCrown } from '@tabler/icons-vue'
import type { FunMatchItem } from './FunMatchPlay.vue'

interface PlayerItem { id: number; name: string }

const props = defineProps<{
  matches: FunMatchItem[]
  players: PlayerItem[]
  show: boolean
}>()

interface Honor { title: string; tag: string; html: string; border: string; bg: string; icon: any; iconColor: string }

const honors = computed<Honor[]>(() => {
  if (!props.show) return []
  const played = props.matches.filter(m => m.played)
  if (played.length === 0) return []

  // 旗鼓相当
  let bm = played[0], bmTotal = 0, bmFound = false
  for (const m of played) {
    const g1d = Math.abs((m.game1_score_male||0) - (m.game1_score_female||0))
    const g2d = Math.abs((m.game2_score_male||0) - (m.game2_score_female||0))
    const hasG3 = m.game3_score_male != null
    const g3d = hasG3 ? Math.abs(m.game3_score_male! - m.game3_score_female!) : 0
    if (g1d !== 2 || g2d !== 2) continue
    if (hasG3 && g3d !== 2) continue
    const t = (m.game1_score_male||0)+(m.game2_score_male||0)+(m.game3_score_male||0) + (m.game1_score_female||0)+(m.game2_score_female||0)+(m.game3_score_female||0)
    if (t > bmTotal) { bmTotal = t; bm = m; bmFound = true }
  }
  if (!bmFound) {
    let bestScore = -1
    for (const m of played) {
      const g = m.game3_score_male != null ? 3 : 2
      const d = Math.abs((m.game1_score_male||0)+(m.game2_score_male||0)+(m.game3_score_male||0) - (m.game1_score_female||0)-(m.game2_score_female||0)-(m.game3_score_female||0))
      const t = (m.game1_score_male||0)+(m.game2_score_male||0)+(m.game3_score_male||0) + (m.game1_score_female||0)+(m.game2_score_female||0)+(m.game3_score_female||0)
      const s = g * 10000 - d * 100 + t
      if (s > bestScore) { bestScore = s; bm = m; bmTotal = t }
    }
  }
  const bmHas3 = bm.game3_score_male != null
  const bmGames = bmHas3
    ? `<div style="display:flex;flex-direction:column;align-items:center;gap:10px;">
        <div style="display:flex;align-items:center;gap:16px;">
          <div style="text-align:center;"><div style="font-size:11px;color:#999;margin-bottom:2px;">G1</div><div style="font-size:32px;font-weight:900;color:#333;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bm.game1_score_male}<span style="color:#ccc;font-weight:300;margin:0 2px;">:</span>${bm.game1_score_female}</div></div>
          <div style="text-align:center;"><div style="font-size:11px;color:#999;margin-bottom:2px;">G2</div><div style="font-size:32px;font-weight:900;color:#333;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bm.game2_score_male}<span style="color:#ccc;font-weight:300;margin:0 2px;">:</span>${bm.game2_score_female}</div></div>
          ${bmHas3 ? `<div style="text-align:center;"><div style="font-size:11px;color:#999;margin-bottom:2px;">G3</div><div style="font-size:32px;font-weight:900;color:#333;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bm.game3_score_male}<span style="color:#ccc;font-weight:300;margin:0 2px;">:</span>${bm.game3_score_female}</div></div>` : ''}
        </div>
      </div>`
    : `<div style="display:flex;align-items:center;gap:16px;">
        <div style="text-align:center;"><div style="font-size:11px;color:#999;margin-bottom:2px;">G1</div><div style="font-size:32px;font-weight:900;color:#333;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bm.game1_score_male}<span style="color:#ccc;font-weight:300;margin:0 2px;">:</span>${bm.game1_score_female}</div></div>
        <div style="text-align:center;"><div style="font-size:11px;color:#999;margin-bottom:2px;">G2</div><div style="font-size:32px;font-weight:900;color:#333;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bm.game2_score_male}<span style="color:#ccc;font-weight:300;margin:0 2px;">:</span>${bm.game2_score_female}</div></div>
      </div>`

  // 难分高下
  let bg = { m: played[0], n: 1, t: 0, sm: 0, sf: 0 }
  for (const m of played) {
    const gs = [{s:m.game1_score_male||0,r:m.game1_score_female||0},{s:m.game2_score_male||0,r:m.game2_score_female||0},{s:m.game3_score_male||0,r:m.game3_score_female||0}]
    for (let i = 0; i < 3; i++) {
      if (gs[i].s+gs[i].r > bg.t) bg = { m, n: i+1, t: gs[i].s+gs[i].r, sm: gs[i].s, sf: gs[i].r }
    }
  }

  // 一骑绝尘
  const ps = new Map<number,{w:number,d:number}>()
  for (const m of played) {
    if (!m.winner_id) continue
    const e = ps.get(m.winner_id) || {w:0,d:0}; e.w++; e.d += Math.abs(((m.game1_score_male||0)+(m.game2_score_male||0)+(m.game3_score_male||0)) - ((m.game1_score_female||0)+(m.game2_score_female||0)+(m.game3_score_female||0)))
    ps.set(m.winner_id, e)
  }
  let bp = { id: 0, w: 0, d: 0 }
  for (const [id, s] of ps) { if (s.w > bp.w || (s.w === bp.w && s.d > bp.d)) bp = { id, w: s.w, d: s.d } }
  const bpn = props.players.find(p => p.id === bp.id)?.name || ''
  const bpm = played.filter(m => m.winner_id === bp.id)
  let bpTotalDiff = 0
  for (const m of bpm) {
    bpTotalDiff += Math.abs(((m.game1_score_male||0)+(m.game2_score_male||0)+(m.game3_score_male||0)) - ((m.game1_score_female||0)+(m.game2_score_female||0)+(m.game3_score_female||0)))
  }
  const bpAvgDiff = bpm.length > 0 ? (bpTotalDiff / bpm.length).toFixed(1) : '0'

  return [
    {
      title: '旗鼓相当', tag: '最胶着对局', icon: IconSwords, iconColor: '#c8960c',
      border: '#c8960c', bg: 'linear-gradient(180deg, #fffef5 0%, #fff8e1 40%, #fef3c7 100%)',
      html: `<div style="text-align:center;margin-bottom:14px;"><div style="font-size:18px;font-weight:700;color:#1a1a1a;">${bm.male_player_name}</div><div style="font-size:13px;color:#aaa;margin:4px 0;">对阵</div><div style="font-size:18px;font-weight:700;color:#1a1a1a;">${bm.female_player_name}</div></div>${bmGames}<div style="margin-top:14px;font-size:12px;color:#999;text-align:center;line-height:1.6;">${bmFound ? '每一小局均以 2 分之差决出胜负<br/>旗鼓相当，难分伯仲' : '本场局数最多且总分最接近的一组对阵'}</div>`,
    },
    {
      title: '难分高下', tag: '最高分单局', icon: IconFlame, iconColor: '#c0392b',
      border: '#c0392b', bg: 'linear-gradient(180deg, #fff8f8 0%, #ffe8e8 40%, #ffd6d6 100%)',
      html: `<div style="text-align:center;margin-bottom:10px;"><span style="font-size:15px;font-weight:700;color:#1a1a1a;">${bg.m.male_player_name}</span><span style="color:#bbb;margin:0 6px;">vs</span><span style="font-size:15px;font-weight:700;color:#1a1a1a;">${bg.m.female_player_name}</span></div>
        <div style="font-size:11px;color:#999;margin-bottom:12px;text-align:center;">第 ${bg.n} 局</div>
        <div style="display:flex;align-items:center;justify-content:center;gap:12px;margin-bottom:6px;">
          <span style="font-size:56px;font-weight:900;color:#c0392b;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bg.sm}</span>
          <span style="font-size:28px;font-weight:200;color:#ddd;">:</span>
          <span style="font-size:56px;font-weight:900;color:#c0392b;font-family:'SF Mono','JetBrains Mono',monospace;line-height:1;">${bg.sf}</span>
        </div>
        <div style="text-align:center;font-size:12px;color:#999;">单局总分全场最高，攻势如潮</div>`,
    },
    {
      title: '一骑绝尘', tag: '最佳球员', icon: IconCrown, iconColor: '#1a56db',
      border: '#1a56db', bg: 'linear-gradient(180deg, #f8faff 0%, #e8f0ff 40%, #d6e4ff 100%)',
      html: bpn ? `<div style="text-align:center;margin-bottom:14px;"><div style="font-size:24px;font-weight:800;color:#1a56db;">${bpn}</div></div>
        <div style="display:flex;justify-content:center;gap:24px;">
          <div style="text-align:center;"><div style="font-size:36px;font-weight:900;color:#1a1a1a;font-family:'SF Mono','JetBrains Mono',monospace;">${bp.w}</div><div style="font-size:11px;color:#999;margin-top:2px;">胜场</div></div>
          <div style="text-align:center;"><div style="font-size:36px;font-weight:900;color:#1a1a1a;font-family:'SF Mono','JetBrains Mono',monospace;">${bpAvgDiff}</div><div style="font-size:11px;color:#999;margin-top:2px;">平均领先</div></div>
          <div style="text-align:center;"><div style="font-size:36px;font-weight:900;color:#1a1a1a;font-family:'SF Mono','JetBrains Mono',monospace;">${bpTotalDiff}</div><div style="font-size:11px;color:#999;margin-top:2px;">净胜总分</div></div>
        </div>` : `<div style="font-size:14px;color:#999;text-align:center;">暂无数据</div>`,
    },
  ]
})
</script>

<template>
  <div v-if="honors.length>0" style="margin:12px 16px;">
    <div style="text-align:center;margin-bottom:20px;">
      <div style="font-size:11px;font-weight:600;color:#bbb;letter-spacing:3px;text-transform:uppercase;">Honors</div>
    </div>
    <div style="display:flex;flex-direction:column;gap:24px;">
      <div v-for="h in honors" :key="h.title"
        style="border-radius:20px;overflow:hidden;box-shadow:0 2px 4px rgba(0,0,0,0.04),0 8px 24px rgba(0,0,0,0.08);"
        :style="{background:h.bg, border:'1px solid '+h.border+'18'}">
        <!-- Accent bar -->
        <div style="height:4px;" :style="{background:'linear-gradient(90deg,'+h.border+','+h.border+'88)'}"></div>
        <!-- Card body -->
        <div style="padding:20px 20px 24px;">
          <!-- Header -->
          <div style="display:flex;align-items:center;justify-content:center;gap:10px;margin-bottom:16px;">
            <component :is="h.icon" :size="24" :stroke-width="2" :color="h.border" />
            <span style="font-size:20px;font-weight:800;letter-spacing:1px;" :style="{color:h.border}">{{ h.title }}</span>
          </div>
          <div style="font-size:11px;color:#999;text-align:center;margin-bottom:16px;letter-spacing:1px;">{{ h.tag }}</div>
          <!-- Dynamic content -->
          <div v-html="h.html"></div>
        </div>
      </div>
    </div>
  </div>
</template>
