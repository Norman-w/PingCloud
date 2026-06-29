<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

interface H2HRecord { opponent_id: number; opponent_name: string; wins: number; losses: number }
interface H2HPlayer { id: number; name: string; records: H2HRecord[] }

const players = ref<H2HPlayer[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    players.value = await fetch('/api/headtohead').then(r => r.json())
  } catch (e) { /* ignore */ }
  finally { loading.value = false }
})

function cellColor(wins: number, losses: number): string {
  if (wins + losses === 0) return '#f5f5f5'
  if (wins > losses) return '#e8f8ef'
  if (wins < losses) return '#fde8e8'
  return '#fffbe6'
}
function cellText(r: H2HRecord): string {
  if (r.wins + r.losses === 0) return '-'
  return `${r.wins}:${r.losses}`
}
</script>

<template>
  <div style="min-height:100vh;background:#f0f2f5;padding-bottom:80px;">
    <div class="hero">
      <div style="display:flex;align-items:center;gap:12px;">
        <span @click="router.back()" style="cursor:pointer;font-size:22px;">&#8592;</span>
        <div class="hero-title">相生相克</div>
      </div>
      <div class="hero-sub">横向是自己的胜负记录</div>
    </div>

    <div v-if="loading" style="text-align:center;padding:60px;color:#969799;">加载中...</div>

    <div v-else style="padding:8px;overflow-x:auto;">
      <table style="border-collapse:collapse;font-size:11px;min-width:100%;">
        <thead>
          <tr>
            <th style="padding:6px 4px;background:#f8f9fa;position:sticky;left:0;z-index:1;">姓名</th>
            <th v-for="p in players" :key="p.id" style="padding:4px 2px;writing-mode:vertical-lr;font-size:10px;color:#666;max-width:20px;background:#f8f9fa;">{{ p.name }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in players" :key="row.id">
            <td style="padding:6px 8px;font-weight:700;font-size:12px;background:#f8f9fa;position:sticky;left:0;white-space:nowrap;">{{ row.name }}</td>
            <td v-for="col in players" :key="col.id"
              style="text-align:center;padding:4px 2px;"
              :style="{background: row.id===col.id ? '#f0f0f0' : cellColor(row.records.find(r=>r.opponent_id===col.id)?.wins||0, row.records.find(r=>r.opponent_id===col.id)?.losses||0)}">
              <template v-if="row.id===col.id">-</template>
              <template v-else>
                <span style="font-weight:600;">{{ cellText(row.records.find(r=>r.opponent_id===col.id)!) }}</span>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
      <div style="display:flex;gap:16px;margin-top:16px;justify-content:center;font-size:12px;color:#969799;">
        <span><span style="display:inline-block;width:12px;height:12px;background:#e8f8ef;border-radius:3px;vertical-align:middle;margin-right:4px;"></span>胜率优势</span>
        <span><span style="display:inline-block;width:12px;height:12px;background:#fde8e8;border-radius:3px;vertical-align:middle;margin-right:4px;"></span>胜率劣势</span>
        <span><span style="display:inline-block;width:12px;height:12px;background:#fffbe6;border-radius:3px;vertical-align:middle;margin-right:4px;"></span>五五开</span>
        <span><span style="display:inline-block;width:12px;height:12px;background:#f5f5f5;border-radius:3px;vertical-align:middle;margin-right:4px;"></span>没打过</span>
      </div>
    </div>

    <div style="padding:16px;">
      <button @click="router.back()" style="width:100%;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:17px;font-weight:600;cursor:pointer;">返回首页</button>
    </div>
  </div>
</template>
