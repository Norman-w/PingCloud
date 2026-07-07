<script setup lang="ts">
import { ref, watch, onBeforeUnmount, computed } from 'vue'

export interface CardDef {
  id: number
  card_type: string
  card_value: number | null
  card_detail: string
  color: string
}

const props = defineProps<{
  show: boolean
  maleName: string
  femaleName: string
  maleRating: number
  femaleRating: number
  drawing: boolean
  mode?: string
  result: { card_type: string; card_value: number | null; card_detail: string } | null
}>()

const emit = defineEmits<{
  (e: 'draw'): void
  (e: 'close'): void
}>()

const WHEEL_CARDS: CardDef[] = [
  { id: 1, card_type: 'handicap', card_value: 2, card_detail: '', color: '#f5a623' },
  { id: 2, card_type: 'handicap', card_value: 3, card_detail: '', color: '#f5a623' },
  { id: 3, card_type: 'handicap', card_value: 4, card_detail: '', color: '#f5a623' },
  { id: 4, card_type: 'handicap', card_value: 5, card_detail: '', color: '#f5a623' },
  { id: 5, card_type: 'spin', card_value: null, card_detail: '', color: '#1989fa' },
  { id: 6, card_type: 'table', card_value: null, card_detail: '', color: '#9b59b6' },
  { id: 7, card_type: 'defense', card_value: null, card_detail: '', color: '#e74c3c' },
]

const ALL_CARDS: CardDef[] = [
  { id: 1, card_type: 'handicap', card_value: 2, card_detail: '', color: '#f5a623' },
  { id: 2, card_type: 'handicap', card_value: 3, card_detail: '', color: '#f5a623' },
  { id: 3, card_type: 'handicap', card_value: 4, card_detail: '', color: '#f5a623' },
  { id: 4, card_type: 'handicap', card_value: 5, card_detail: '', color: '#f5a623' },
  { id: 5, card_type: 'spin', card_value: null, card_detail: 'topspin', color: '#1989fa' },
  { id: 6, card_type: 'spin', card_value: null, card_detail: 'backspin', color: '#1989fa' },
  { id: 7, card_type: 'table', card_value: null, card_detail: 'left', color: '#9b59b6' },
  { id: 8, card_type: 'table', card_value: null, card_detail: 'right', color: '#9b59b6' },
  { id: 9, card_type: 'defense', card_value: null, card_detail: '', color: '#e74c3c' },
]

const activeCards = computed(() => props.mode === 'wheel_rr' ? WHEEL_CARDS : ALL_CARDS)

const started = ref(false)
const spinning = ref(false)
const highlightIdx = ref(-1)
const resultIdx = ref(-1)
const spinTimer = ref<ReturnType<typeof setInterval> | null>(null)
const phaseTimeout = ref<ReturnType<typeof setTimeout> | null>(null)
const spinStartTime = ref(0)
const MIN_SPIN_MS = 4000 // 4 seconds minimum

const pendingResultIdx = ref(-1)
let revealCheckTimer: ReturnType<typeof setInterval> | null = null

// Circle layout: compute card positions around a circle
const circleRadius = 140
const containerSize = circleRadius * 2 + 80 // enough for card + offset

const cardPositions = computed(() => {
  const count = activeCards.value.length
  const angleStep = (2 * Math.PI) / count
  const startAngle = -Math.PI / 2 // start from top
  return activeCards.value.map((_, i) => {
    const angle = startAngle + i * angleStep
    return {
      x: circleRadius * Math.cos(angle),
      y: circleRadius * Math.sin(angle),
      angle: angle * (180 / Math.PI),
    }
  })
})

watch(() => props.show, (v) => {
  if (v) {
    started.value = false
    spinning.value = false
    highlightIdx.value = -1
    resultIdx.value = -1
    pendingResultIdx.value = -1
    spinStartTime.value = 0
    if (revealCheckTimer) { clearInterval(revealCheckTimer); revealCheckTimer = null }
  }
})

function tryReveal() {
  if (pendingResultIdx.value < 0) return
  if (Date.now() - spinStartTime.value >= MIN_SPIN_MS) {
    stopSpin(pendingResultIdx.value)
  }
}

watch(() => props.result, (r) => {
  if (!r || !spinning.value) return
  const idx = activeCards.value.findIndex(c =>
    c.card_type === r.card_type && c.card_value === r.card_value && c.card_detail === r.card_detail)
  if (idx < 0) return

  if (Date.now() - spinStartTime.value >= MIN_SPIN_MS) {
    stopSpin(idx)
  } else {
    pendingResultIdx.value = idx
    if (!revealCheckTimer) {
      revealCheckTimer = setInterval(tryReveal, 200)
    }
  }
})

function cardLabel(c: CardDef): string {
  switch (c.card_type) {
    case 'handicap': return `让${c.card_value}分`
    case 'spin': return c.card_detail === 'topspin' ? '上旋' : c.card_detail === 'backspin' ? '下旋' : '旋转'
    case 'table': return c.card_detail === 'left' ? '左半台' : c.card_detail === 'right' ? '右半台' : '半台'
    case 'defense': return '防守'
    default: return '?'
  }
}

function cardDesc(c: CardDef): string {
  switch (c.card_type) {
    case 'handicap': return `对手开局领先${c.card_value}分`
    case 'spin': return c.card_detail ? (c.card_detail === 'topspin' ? '只允许发上旋球' : '只允许发下旋球') : '由低分方指定发球旋转'
    case 'table': return c.card_detail ? (c.card_detail === 'left' ? '只允许发球到左半台' : '只允许发球到右半台') : '由低分方指定发球半台'
    case 'defense': return '只允许防守不允许进攻'
    default: return ''
  }
}

function cardIcon(c: CardDef): string {
  switch (c.card_type) {
    case 'handicap': return String(c.card_value ?? '')
    case 'spin': return c.card_detail === 'topspin' ? '↑' : '↓'
    case 'table': return c.card_detail === 'left' ? '◧' : '◨'
    case 'defense': return '🛡'
    default: return '?'
  }
}

function startDraw() {
  started.value = true
  spinning.value = true
  resultIdx.value = -1
  highlightIdx.value = 0
  spinStartTime.value = Date.now()
  pendingResultIdx.value = -1
  emit('draw')

  // Marquee: faster cadence, 3-5 phases, ~5-8s total
  let phaseIdx = 0
  const phases: Array<[number, number]> = [
    [60, 1200 + Math.random() * 800],    // fast ~1.2-2s
    [90, 1000 + Math.random() * 800],    // ~1-1.8s
    [150, 1200 + Math.random() * 1000],  // slowing ~1.2-2.2s
    [80, 800 + Math.random() * 600],     // speed burst! ~0.8-1.4s
    [200, 1500 + Math.random() * 1000],  // winding down ~1.5-2.5s
    [350, 999999],                        // crawl
  ]

  function runNext() {
    if (!spinning.value) return
    if (phaseIdx >= phases.length) return

    const [speed, duration] = phases[phaseIdx]
    if (spinTimer.value) clearInterval(spinTimer.value)
    spinTimer.value = setInterval(() => {
      highlightIdx.value = (highlightIdx.value + 1) % activeCards.value.length
    }, speed)

    phaseIdx++
    if (phaseIdx < phases.length) {
      phaseTimeout.value = setTimeout(runNext, duration)
    }
  }

  runNext()
}

function stopSpin(idx: number) {
  if (spinTimer.value) { clearInterval(spinTimer.value); spinTimer.value = null }
  if (phaseTimeout.value) { clearTimeout(phaseTimeout.value); phaseTimeout.value = null }
  if (revealCheckTimer) { clearInterval(revealCheckTimer); revealCheckTimer = null }
  pendingResultIdx.value = -1
  spinning.value = false
  highlightIdx.value = idx
  resultIdx.value = idx
}

function onClose() {
  if (spinTimer.value) { clearInterval(spinTimer.value); spinTimer.value = null }
  if (phaseTimeout.value) { clearTimeout(phaseTimeout.value); phaseTimeout.value = null }
  if (revealCheckTimer) { clearInterval(revealCheckTimer); revealCheckTimer = null }
  pendingResultIdx.value = -1
  spinning.value = false
  emit('close')
}

onBeforeUnmount(() => {
  if (spinTimer.value) clearInterval(spinTimer.value)
  if (phaseTimeout.value) clearTimeout(phaseTimeout.value)
  if (revealCheckTimer) clearInterval(revealCheckTimer)
})
</script>

<template>
  <div v-if="show" class="card-draw-overlay" @click.self="onClose">
    <!-- Title -->
    <div style="font-size: 18px; font-weight: 800; color: #f5a623; margin-bottom: 2px; z-index: 10;">
      🎰 高手紧箍咒
    </div>
    <div style="font-size: 12px; color: #999; z-index: 10; text-align: center; line-height: 1.8;">
      {{ maleName }} <span :style="{color: maleRating >= femaleRating ? '#f5a623' : '#999'}">{{ maleRating }}</span>
      vs
      {{ femaleName }} <span :style="{color: femaleRating >= maleRating ? '#f5a623' : '#999'}">{{ femaleRating }}</span>
    </div>
    <div style="font-size: 14px; font-weight: 700; z-index: 10; margin-bottom: 4px;"
      :style="{color: '#f5a623'}">
      👉 {{ maleRating >= femaleRating ? maleName : femaleName }} 抽卡
    </div>

    <!-- Circular card ring -->
    <div class="circle-container" :style="{ width: containerSize+'px', height: containerSize+'px' }">
      <!-- Center: start button, spinning highlight, or result -->
      <div v-if="!started" class="center-start">
        <button @click="startDraw" class="start-btn">🎰<br/>抽卡</button>
      </div>
      <div v-else-if="resultIdx >= 0" class="center-result" :style="{ borderColor: activeCards[resultIdx].color }">
        <div class="result-icon" :style="{ color: activeCards[resultIdx].color }">{{ cardIcon(activeCards[resultIdx]) }}</div>
        <div style="font-size: 20px; font-weight: 800; color: #fff;">⭐ 抽中!</div>
        <div style="font-size: 16px; font-weight: 700; color: #ccc;">{{ cardLabel(activeCards[resultIdx]) }}</div>
        <div style="font-size: 12px; color: #999; margin-top: 4px;">{{ cardDesc(activeCards[resultIdx]) }}</div>
      </div>
      <div v-else class="center-spinning" :style="{ borderColor: activeCards[highlightIdx].color }">
        <div class="result-icon" :style="{ color: activeCards[highlightIdx].color }">{{ cardIcon(activeCards[highlightIdx]) }}</div>
        <div style="font-size: 14px; font-weight: 700; color: '#fff';">{{ cardLabel(activeCards[highlightIdx]) }}</div>
      </div>

      <!-- Ring of cards -->
      <div v-for="(c, i) in activeCards" :key="c.id"
        class="ring-card"
        :style="{
          borderColor: c.color,
          background: highlightIdx === i ? c.color : 'rgba(255,255,255,0.06)',
          color: highlightIdx === i ? '#fff' : c.color,
          opacity: resultIdx >= 0 && resultIdx !== i ? 0.2 : 1,
          boxShadow: highlightIdx === i ? `0 0 16px ${c.color}, 0 0 32px ${c.color}` : 'none',
          transform: highlightIdx === i
            ? `translate(${cardPositions[i].x}px, ${cardPositions[i].y}px) scale(1.15)`
            : `translate(${cardPositions[i].x}px, ${cardPositions[i].y}px) scale(1)`,
          zIndex: highlightIdx === i ? 20 : 1,
          transition: 'transform 0.1s, box-shadow 0.1s, opacity 0.3s',
        }">
        <span style="font-size: 18px; font-weight: 800;">{{ cardIcon(c) }}</span>
        <span style="font-size: 9px; font-weight: 600; margin-top: 2px;">{{ cardLabel(c) }}</span>
      </div>

      <!-- Marquee trail dots -->
      <div v-for="(_, i) in activeCards" :key="'dot'+i"
        class="ring-dot"
        :style="{
          transform: `translate(${cardPositions[i].x}px, ${cardPositions[i].y}px)`,
          background: activeCards[i].color,
          opacity: (() => {
            if (!spinning || highlightIdx < 0) return 0
            const dist = (i - highlightIdx + activeCards.length) % activeCards.length
            if (dist === 0) return 0
            return Math.max(0, 1 - dist / activeCards.length)
          })(),
          width: (4 + 2 * (1 - ((i - highlightIdx + activeCards.length) % activeCards.length) / activeCards.length))+'px',
          height: (4 + 2 * (1 - ((i - highlightIdx + activeCards.length) % activeCards.length) / activeCards.length))+'px',
        }">
      </div>
    </div>

    <!-- Status / confirm -->
    <div style="display: flex; gap: 12px; min-height: 46px; z-index: 10; margin-top: 8px;">
      <div v-if="spinning" style="color:#f5a623;font-size:15px;font-weight:700;padding:12px;">
        🎰 抽卡中...
      </div>
      <button v-if="resultIdx >= 0" @click="onClose"
        style="padding:14px 48px;background:#07c160;color:#fff;border:none;border-radius:24px;font-size:18px;font-weight:700;cursor:pointer;box-shadow:0 4px 16px rgba(7,193,96,0.4);">
        确认
      </button>
    </div>
  </div>
</template>

<style scoped>
.card-draw-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.93);
  z-index: 1100;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}
.circle-container {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.center-start,
.center-spinning,
.center-result {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  z-index: 30;
  transition: all 0.3s;
}
.center-start {
  width: 120px;
  height: 120px;
}
.start-btn {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #f5a623, #ffd700);
  color: #333;
  font-size: 20px;
  font-weight: 800;
  cursor: pointer;
  box-shadow: 0 0 30px rgba(245,166,35,0.5);
  animation: bounce 1s ease-in-out infinite alternate;
  line-height: 1.3;
}
.center-spinning {
  width: 130px;
  height: 130px;
  border: 2px solid;
  background: rgba(0,0,0,0.5);
}
.center-result {
  width: 160px;
  height: 180px;
  border: 3px solid;
  background: rgba(0,0,0,0.7);
  animation: popIn 0.4s ease-out;
}
.result-icon {
  font-size: 36px;
  font-weight: 800;
}
.ring-card {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 52px;
  height: 62px;
  margin-left: -26px;
  margin-top: -31px;
  border-radius: 10px;
  border: 1.5px solid;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transition: transform 0.12s, box-shadow 0.12s, background 0.2s;
}
.ring-dot {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 4px;
  height: 4px;
  margin-left: -2px;
  margin-top: -2px;
  border-radius: 50%;
  transition: opacity 0.3s;
}
@keyframes popIn {
  from { transform: translate(-50%, -50%) scale(0.5); opacity: 0; }
  to { transform: translate(-50%, -50%) scale(1); opacity: 1; }
}
@keyframes bounce {
  from { transform: translateY(0); }
  to { transform: translateY(-6px); }
}
</style>
