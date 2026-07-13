<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { IconClock, IconTarget, IconPlayerPlay, IconStar, IconStarFilled } from '@tabler/icons-vue'

const visible = defineModel<boolean>('visible', { default: false })
const emit = defineEmits<{ start: [duration: number] }>()

const durations = [15, 30, 45, 60, 90, 120]
const customDuration = ref('')
const selectedDuration = ref(60)

interface Goal { id: number; label: string; goal_type: string; current_value: number; stars: number; passed: boolean; star_1: number; star_2: number; star_3: number; star_4: number; star_5: number; min_stars: number }
const goals = ref<Goal[]>([])
const loadingGoals = ref(false)

const effectiveDuration = computed(() => {
  if (customDuration.value) return parseInt(customDuration.value) || selectedDuration.value
  return selectedDuration.value
})

const durationLabel = computed(() => {
  const m = effectiveDuration.value
  if (m < 60) return `${m}分钟`
  const h = Math.floor(m/60); const r = m%60
  return r > 0 ? `${h}小时${r}分` : `${h}小时`
})

async function loadGoals(skillId: number) {
  loadingGoals.value = true
  try { const r = await fetch('/api/skill-goals/' + skillId); if (r.ok) goals.value = await r.json() } catch {}
  loadingGoals.value = false
}

function starArray(stars: number) { return Array.from({length:5}, (_,i) => i < stars) }

function start() { emit('start', effectiveDuration.value); visible.value = false }

watch(visible, async (v) => { if (v) { selectedDuration.value = 60; customDuration.value = '' } })

defineExpose({ loadGoals })
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.5);z-index:2500;display:flex;align-items:flex-end;">
    <div style="background:#fff;border-radius:20px 20px 0 0;width:100%;max-height:80vh;display:flex;flex-direction:column;">
      <div style="padding:20px;border-bottom:1px solid #f0f0f0;display:flex;align-items:center;justify-content:space-between;">
        <span style="font-weight:700;font-size:18px;">⚡ 开始训练</span>
        <button @click="visible=false" style="background:none;border:none;font-size:22px;color:#bbb;cursor:pointer;">✕</button>
      </div>

      <div style="flex:1;overflow-y:auto;padding:16px 20px;">
        <!-- Duration picker -->
        <div style="margin-bottom:20px;">
          <div style="font-size:14px;font-weight:600;color:#555;margin-bottom:10px;display:flex;align-items:center;gap:6px;">
            <IconClock :size="18" :stroke-width="2" style="color:#1989fa;" /> 目标时长
          </div>
          <div style="display:flex;flex-wrap:wrap;gap:8px;margin-bottom:10px;">
            <span v-for="d in durations" :key="d" @click="selectedDuration=d; customDuration=''"
              style="padding:10px 16px;border-radius:10px;font-size:14px;font-weight:600;cursor:pointer;transition:all .15s;"
              :style="selectedDuration===d && !customDuration ? 'background:#1989fa;color:#fff;' : 'background:#f0f2f5;color:#666;'">
              {{ d >= 60 ? Math.floor(d/60)+'h'+(d%60>0?d%60+'m':'') : d+'m' }}
            </span>
          </div>
          <input v-model="customDuration" type="number" placeholder="自定义(分钟)" min="1"
            style="width:100%;padding:12px;border:1px solid #e0e0e0;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;" />
        </div>

        <!-- Goals display -->
        <div v-if="goals.length > 0" style="margin-bottom:16px;">
          <div style="font-size:14px;font-weight:600;color:#555;margin-bottom:10px;display:flex;align-items:center;gap:6px;">
            <IconTarget :size="18" :stroke-width="2" style="color:#f5a623;" /> 训练目标
          </div>
          <div v-for="g in goals" :key="g.id" style="background:#f8f9fa;border-radius:12px;padding:14px;margin-bottom:8px;">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px;">
              <span style="font-weight:600;font-size:14px;">{{ g.label }}</span>
              <div style="display:flex;gap:2px;">
                <template v-for="i in 5" :key="i">
                  <IconStarFilled v-if="i <= g.stars" :size="14" style="color:#f5a623;" />
                  <IconStar v-else :size="14" style="color:#ddd;" />
                </template>
              </div>
            </div>
            <div style="font-size:12px;color:#999;">
              当前 {{ g.current_value }} ·
              升星: {{ g.star_1 }}→{{ g.star_2 }}→{{ g.star_3 }}→{{ g.star_4 }}→{{ g.star_5 }}
              · 需{{ g.min_stars }}★过关
              <span v-if="g.passed" style="color:#07c160;font-weight:600;"> ✓ 已达标</span>
              <span v-else style="color:#ff976a;"> 未达标</span>
            </div>
            <div style="margin-top:6px;height:4px;background:#e0e0e0;border-radius:2px;overflow:hidden;">
              <div style="height:100%;border-radius:2px;transition:width .3s;"
                :style="{width:Math.min(100, g.current_value/Math.max(1,g.star_2)*100)+'%',background:g.passed?'#07c160':'#1989fa'}"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Start button -->
      <div style="padding:16px 20px;border-top:1px solid #f0f0f0;">
        <div style="text-align:center;font-size:24px;font-weight:800;color:#07c160;margin-bottom:4px;">{{ durationLabel }}</div>
        <button @click="start" style="width:100%;padding:16px;background:linear-gradient(135deg,#07c160,#00bfa5);color:#fff;border:none;border-radius:14px;font-size:18px;font-weight:700;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:8px;">
          <IconPlayerPlay :size="22" /> 开始训练
        </button>
      </div>
      <div style="height:env(safe-area-inset-bottom);"></div>
    </div>
  </div>
</template>
