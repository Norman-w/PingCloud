<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { IconArrowBack, IconClock } from '@tabler/icons-vue'

const router = useRouter()

// Players
const nameA = ref('主队')
const nameB = ref('客队')

// Point scores (current game)
const pointA = ref(0)
const pointB = ref(0)

// Game scores
const gameA = ref(0)
const gameB = ref(0)
const gamesToWin = ref(3) // best of 5

// Cards
const yellowA = ref(0)
const redA = ref(0)
const yellowB = ref(0)
const redB = ref(0)

// Net touch
const netTouch = ref(false)
const netTouchSide = ref<'A' | 'B' | null>(null)

// Timeout
const timeoutA = ref(false)
const timeoutB = ref(false)

// Wakelock
let wakeLock: any = null

async function keepAwake() {
  try {
    wakeLock = await (navigator as any).wakeLock?.request?.('screen')
  } catch (e) { /* not supported */ }
}

onMounted(() => {
  keepAwake()
  document.addEventListener('visibilitychange', () => { if (wakeLock) keepAwake() })
})

onUnmounted(() => {
  wakeLock?.release?.()
})

// Brightness
function maxBrightness() {
  const style = document.createElement('style')
  style.innerHTML = 'html { filter: brightness(2); }'
  style.id = 'brightness-overlay'
  document.head.appendChild(style)
}
onMounted(() => maxBrightness())
onUnmounted(() => document.getElementById('brightness-overlay')?.remove())

// Point controls
function addPoint(side: 'A' | 'B') {
  if (finish.value) return
  if (side === 'A') pointA.value++
  else pointB.value++
  // Auto-advance game
  if ((pointA.value >= 11 || pointB.value >= 11) && Math.abs(pointA.value - pointB.value) >= 2) {
    if (pointA.value > pointB.value) gameA.value++
    else gameB.value++
    pointA.value = 0; pointB.value = 0
  }
}
function undoPoint() {
  if (pointA.value > 0) pointA.value--
  else if (pointB.value > 0) pointB.value--
  else {
    if (gameA.value > 0) { gameA.value--; pointA.value = 10; pointB.value = 12 }
    else if (gameB.value > 0) { gameB.value--; pointA.value = 12; pointB.value = 10 }
  }
}

const finish = ref(false)
const winner = ref('')

function checkWinner() {
  if (gameA.value >= gamesToWin.value) { finish.value = true; winner.value = nameA.value }
  else if (gameB.value >= gamesToWin.value) { finish.value = true; winner.value = nameB.value }
}

function toggleNetTouch(side: 'A' | 'B') {
  if (netTouch.value && netTouchSide.value === side) { netTouch.value = false; netTouchSide.value = null }
  else { netTouch.value = true; netTouchSide.value = side }
}

function card(side: 'A' | 'B', color: 'yellow' | 'red') {
  if (color === 'yellow') {
    if (side === 'A') yellowA.value++
    else yellowB.value++
  } else {
    if (side === 'A') redA.value++
    else redB.value++
  }
}

function toggleTimeout(side: 'A' | 'B') {
  if (side === 'A') timeoutA.value = !timeoutA.value
  else timeoutB.value = !timeoutB.value
}

function reset() {
  pointA.value = 0; pointB.value = 0
  gameA.value = 0; gameB.value = 0
  yellowA.value = 0; redA.value = 0
  yellowB.value = 0; redB.value = 0
  netTouch.value = false; netTouchSide.value = null
  timeoutA.value = false; timeoutB.value = false
  finish.value = false; winner.value = ''
}

function swapSides() {
  const tmpN = nameA.value; nameA.value = nameB.value; nameB.value = tmpN
  const tmpP = pointA.value; pointA.value = pointB.value; pointB.value = tmpP
  const tmpG = gameA.value; gameA.value = gameB.value; gameB.value = tmpG
  const tmpY = yellowA.value; yellowA.value = yellowB.value; yellowB.value = tmpY
  const tmpR = redA.value; redA.value = redB.value; redB.value = tmpR
  const tmpT = timeoutA.value; timeoutA.value = timeoutB.value; timeoutB.value = tmpT
}
</script>

<template>
  <div style="min-height: 100vh; background: #0a0a1a; color: #fff; user-select: none; -webkit-user-select: none; display: flex; flex-direction: column;">

    <!-- Top bar -->
    <div style="display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; background: #111;">
      <button @click="router.back()" style="background: none; border: none; color: #fff; font-size: 16px; cursor: pointer;">&#8592; 退出</button>
      <div style="font-size: 14px; font-weight: 600;">记分牌</div>
      <div style="display: flex; gap: 8px;">
        <button @click="swapSides" style="background: #333; border: none; color: #fff; padding: 4px 10px; border-radius: 6px; font-size: 12px; cursor: pointer;">⇄ 换边</button>
        <button @click="reset" style="background: #c0392b; border: none; color: #fff; padding: 4px 10px; border-radius: 6px; font-size: 12px; cursor: pointer;">重置</button>
      </div>
    </div>

    <!-- Winner overlay -->
    <div v-if="finish" style="position: fixed; inset: 0; background: rgba(0,0,0,0.85); z-index: 100; display: flex; flex-direction: column; align-items: center; justify-content: center;">
      <div style="font-size: 48px; font-weight: 900; color: #f1c40f;">🏆 {{ winner }}</div>
      <div style="font-size: 24px; margin-top: 8px; color: #aaa;">{{ gameA }} : {{ gameB }}</div>
      <button @click="reset" style="margin-top: 24px; background: #f1c40f; border: none; color: #000; padding: 14px 40px; border-radius: 24px; font-size: 18px; font-weight: 700; cursor: pointer;">重新开始</button>
    </div>

    <!-- Game score -->
    <div style="text-align: center; padding: 4px 0;">
      <span style="font-size: 32px; font-weight: 900;">{{ gameA }}</span>
      <span style="font-size: 24px; color: #666; margin: 0 8px;">:</span>
      <span style="font-size: 32px; font-weight: 900;">{{ gameB }}</span>
      <div style="font-size: 11px; color: #666; margin-top: 2px;">大局 {{ gamesToWin }} 局 {{ gamesToWin * 2 - 1 }} 胜</div>
    </div>

    <!-- Cards row -->
    <div style="display: flex; justify-content: center; gap: 40px; padding: 4px 0;">
      <div style="text-align: center;">
        <button @click="card('A', 'yellow')"
          style="width: 24px; height: 32px; background: #f1c40f; border: none; border-radius: 4px; cursor: pointer; display: block; margin: 0 auto 2px;"
          :style="{ opacity: yellowA > 0 ? 1 : 0.3 }"></button>
        <button @click="card('A', 'red')"
          style="width: 24px; height: 32px; background: #e74c3c; border: none; border-radius: 4px; cursor: pointer; display: block; margin: 0 auto 2px;"
          :style="{ opacity: redA > 0 ? 1 : 0.3 }"></button>
        <span style="font-size: 10px; color: #666;">{{ yellowA > 0 ? `黄×${yellowA}` : '' }} {{ redA > 0 ? `红×${redA}` : '' }}</span>
      </div>
      <div style="text-align: center;">
        <button @click="card('B', 'yellow')"
          style="width: 24px; height: 32px; background: #f1c40f; border: none; border-radius: 4px; cursor: pointer; display: block; margin: 0 auto 2px;"
          :style="{ opacity: yellowB > 0 ? 1 : 0.3 }"></button>
        <button @click="card('B', 'red')"
          style="width: 24px; height: 32px; background: #e74c3c; border: none; border-radius: 4px; cursor: pointer; display: block; margin: 0 auto 2px;"
          :style="{ opacity: redB > 0 ? 1 : 0.3 }"></button>
        <span style="font-size: 10px; color: #666;">{{ yellowB > 0 ? `黄×${yellowB}` : '' }} {{ redB > 0 ? `红×${redB}` : '' }}</span>
      </div>
    </div>

    <!-- Main score area -->
    <div style="flex: 1; display: flex; align-items: center; justify-content: center; position: relative;">

      <!-- Player A side (left half) -->
      <div @click="addPoint('A')" style="flex: 1; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; background: timeoutA ? '#1a1a00' : '#0d0d2b'; cursor: pointer; position: relative;">
        <div v-if="timeoutA" style="position: absolute; top: 8px; background: #f1c40f; color: #000; padding: 2px 10px; border-radius: 10px; font-size: 12px; font-weight: 700;">暂停</div>
        <div style="font-size: 18px; font-weight: 700; margin-bottom: 8px;">{{ nameA }}</div>
        <div style="font-size: 72px; font-weight: 900; line-height: 1;">{{ pointA }}</div>
        <button @click.stop="toggleTimeout('A')" style="margin-top: 12px; background: rgba(255,255,255,0.1); border: 1px solid rgba(255,255,255,0.3); color: #aaa; padding: 6px 16px; border-radius: 6px; font-size: 12px; cursor: pointer;">
          <IconClock :size="14" /> {{ timeoutA ? '继续' : '暂停' }}
        </button>
      </div>

      <!-- Center divider -->
      <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); text-align: center; pointer-events: none;">
        <div style="font-size: 20px; color: #444; font-weight: 900;">VS</div>
      </div>

      <!-- Player B side (right half) -->
      <div @click="addPoint('B')" style="flex: 1; height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; background: timeoutB ? '#1a1a00' : '#1a001a'; cursor: pointer; position: relative;">
        <div v-if="timeoutB" style="position: absolute; top: 8px; background: #f1c40f; color: #000; padding: 2px 10px; border-radius: 10px; font-size: 12px; font-weight: 700;">暂停</div>
        <div style="font-size: 18px; font-weight: 700; margin-bottom: 8px;">{{ nameB }}</div>
        <div style="font-size: 72px; font-weight: 900; line-height: 1;">{{ pointB }}</div>
        <button @click.stop="toggleTimeout('B')" style="margin-top: 12px; background: rgba(255,255,255,0.1); border: 1px solid rgba(255,255,255,0.3); color: #aaa; padding: 6px 16px; border-radius: 6px; font-size: 12px; cursor: pointer;">
          <IconClock :size="14" /> {{ timeoutB ? '继续' : '暂停' }}
        </button>
      </div>
    </div>

    <!-- Bottom controls -->
    <div style="padding: 8px 16px 16px; background: #111; display: flex; flex-direction: column; gap: 8px;">

      <!-- Net touch -->
      <div style="display: flex; justify-content: center; gap: 16px;">
        <button @click="toggleNetTouch('A')" style="padding: 6px 16px; border-radius: 8px; border: 2px solid; font-size: 13px; cursor: pointer;"
          :style="netTouch && netTouchSide === 'A' ? { background: '#e74c3c', color: '#fff', borderColor: '#e74c3c' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
          {{ nameA }} 擦网
        </button>
        <button @click="undoPoint" style="background: #333; border: none; color: #aaa; padding: 6px 16px; border-radius: 8px; font-size: 13px; cursor: pointer;">↩ 撤销</button>
        <button @click="toggleNetTouch('B')" style="padding: 6px 16px; border-radius: 8px; border: 2px solid; font-size: 13px; cursor: pointer;"
          :style="netTouch && netTouchSide === 'B' ? { background: '#e74c3c', color: '#fff', borderColor: '#e74c3c' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
          {{ nameB }} 擦网
        </button>
      </div>

      <!-- Player name inputs -->
      <div style="display: flex; gap: 8px;">
        <input v-model="nameA" placeholder="选手A" style="flex: 1; padding: 8px; background: #222; border: 1px solid #333; border-radius: 6px; color: #fff; font-size: 14px; text-align: center;" />
        <select v-model.number="gamesToWin" style="width: 80px; padding: 8px; background: #222; border: 1px solid #333; border-radius: 6px; color: #fff; font-size: 14px; text-align: center;">
          <option :value="2">BO3</option>
          <option :value="3">BO5</option>
          <option :value="4">BO7</option>
        </select>
        <input v-model="nameB" placeholder="选手B" style="flex: 1; padding: 8px; background: #222; border: 1px solid #333; border-radius: 6px; color: #fff; font-size: 14px; text-align: center;" />
      </div>
    </div>
  </div>
</template>
