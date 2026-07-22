<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  teamAName: string
  teamBName: string
  drawing: boolean
  result: { card_type: string; card_detail: string } | null
}>()

const emit = defineEmits<{
  (e: 'draw'): void
  (e: 'close'): void
}>()

const CARD_DEFS = [
  { type: 'edge_double', label: '擦边翻倍卡', desc: '发球擦边得2分', color: '#f5a623', icon: '🔁' },
  { type: 'net_deduction', label: '擦网扣分卡', desc: '除发球外擦网扣1分', color: '#ee0a24', icon: '🕸️' },
]

const spinning = ref(false)
const currentCard = ref(0)
const chosenIdx = ref(-1)
const spinTimer = ref<ReturnType<typeof setInterval> | null>(null)
const spinStartTime = ref(0)
const MIN_SPIN_MS = 2000
const pendingIdx = ref(-1)
let checkTimer: ReturnType<typeof setInterval> | null = null

function startSpinning() {
  spinning.value = true
  chosenIdx.value = -1
  spinStartTime.value = Date.now()
  let interval = 60
  let phase = 0

  function tick() {
    currentCard.value = (currentCard.value + 1) % 2
    const elapsed = Date.now() - spinStartTime.value
    if (elapsed > 800) phase = 1; interval = 120
    if (elapsed > 1600) phase = 2; interval = 250
    if (elapsed > 2500) phase = 3; interval = 400

    if (spinning.value) {
      spinTimer.value = setTimeout(tick, interval)
    }
  }
  tick()

  // Poll for result
  checkTimer = setInterval(() => {
    if (pendingIdx.value >= 0) {
      const elapsed = Date.now() - spinStartTime.value
      if (elapsed >= MIN_SPIN_MS) {
        stopSpin(pendingIdx.value)
      }
    }
  }, 150)
}

function stopSpin(idx: number) {
  if (spinTimer.value) clearTimeout(spinTimer.value)
  if (checkTimer) clearInterval(checkTimer)
  chosenIdx.value = idx
  currentCard.value = idx
  spinning.value = false
  spinTimer.value = null
}

function onDraw() {
  if (spinning.value) return
  emit('draw')
  startSpinning()
}

function onClose() {
  if (spinTimer.value) clearTimeout(spinTimer.value)
  if (checkTimer) clearInterval(checkTimer)
  spinning.value = false
  chosenIdx.value = -1
  pendingIdx.value = -1
  currentCard.value = 0
  emit('close')
}

// Watch for result from parent
watch(() => props.result, (r) => {
  if (r && spinning.value) {
    const idx = CARD_DEFS.findIndex(c => c.type === r.card_type)
    if (idx >= 0) {
      const elapsed = Date.now() - spinStartTime.value
      if (elapsed >= MIN_SPIN_MS) {
        stopSpin(idx)
      } else {
        pendingIdx.value = idx
      }
    }
  }
})

watch(() => props.show, (v) => {
  if (!v) {
    spinning.value = false
    chosenIdx.value = -1
    pendingIdx.value = -1
  }
})
</script>

<template>
  <div v-if="show" style="position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 500; display: flex; align-items: center; justify-content: center;" @click.self="onClose">
    <div style="background: #fff; border-radius: 16px; padding: 24px 20px; width: 90%; max-width: 340px; text-align: center;">
      <h3 style="font-size: 18px; font-weight: 700; margin-bottom: 4px;">🎴 趣味卡抽取</h3>
      <div style="font-size: 12px; color: #969799; margin-bottom: 16px;">每队赛前各抽一张，全队生效</div>

      <!-- Two cards side by side -->
      <div style="display: flex; gap: 12px; justify-content: center; margin-bottom: 20px;">
        <div v-for="(card, i) in CARD_DEFS" :key="card.type"
          :style="{
            flex: 1, maxWidth: '140px', padding: '16px 12px', borderRadius: '14px',
            background: chosenIdx === i ? card.color : '#f8f9fa',
            color: chosenIdx === i ? '#fff' : '#666',
            border: currentCard === i && spinning ? '3px solid ' + card.color : '2px solid transparent',
            transform: currentCard === i && spinning ? 'scale(1.05)' : 'scale(1)',
            transition: 'all 0.15s',
            opacity: chosenIdx >= 0 && chosenIdx !== i ? 0.5 : 1,
          }">
          <div style="font-size: 32px; margin-bottom: 8px;">{{ card.icon }}</div>
          <div style="font-size: 14px; font-weight: 700;">{{ card.label }}</div>
          <div style="font-size: 11px; opacity: 0.8; margin-top: 4px;">{{ card.desc }}</div>
        </div>
      </div>

      <!-- Draw button -->
      <button v-if="!spinning && chosenIdx < 0" @click="onDraw"
        style="padding: 14px 40px; background: linear-gradient(135deg, #f5a623, #e8961a); color: #fff; border: none; border-radius: 24px; font-size: 16px; font-weight: 700; cursor: pointer; box-shadow: 0 4px 12px rgba(245,166,35,0.3);">
        🎲 抽卡
      </button>

      <!-- Result -->
      <div v-if="chosenIdx >= 0 && !spinning" style="animation: popIn 0.3s ease-out;">
        <div style="font-size: 40px; margin-bottom: 4px;">{{ CARD_DEFS[chosenIdx].icon }}</div>
        <div :style="{ fontSize: '18px', fontWeight: 800, color: CARD_DEFS[chosenIdx].color, marginBottom: '4px' }">
          {{ CARD_DEFS[chosenIdx].label }}
        </div>
        <div style="font-size: 12px; color: #969799; margin-bottom: 16px;">{{ CARD_DEFS[chosenIdx].desc }}</div>
        <button @click="onClose"
          style="padding: 12px 32px; background: #07c160; color: #fff; border: none; border-radius: 24px; font-size: 15px; font-weight: 700; cursor: pointer;">确认</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes popIn {
  from { transform: scale(0.8); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
