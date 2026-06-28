<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { IconClock } from '@tabler/icons-vue'

const router = useRouter()

// Setup
const started = ref(false)
const nameA = ref('')
const nameB = ref('')
const format = ref<'11' | 'golden' | '7'>('11')
const gamesToWin = ref(3)

// Point scores
const pointA = ref(0)
const pointB = ref(0)
const gameA = ref(0)
const gameB = ref(0)

// Cards + net + timeout
const yellowA = ref(0); const redA = ref(0)
const yellowB = ref(0); const redB = ref(0)
const netTouch = ref(false); const netTouchSide = ref<'A' | 'B' | null>(null)
const timeoutA = ref(false); const timeoutB = ref(false)

// Wakelock + fullscreen
let wakeLock: any = null
const container = ref<HTMLElement>()

async function keepAwake() {
  try { wakeLock = await (navigator as any).wakeLock?.request?.('screen') } catch (e) {}
}
async function enterFullscreen() {
  try {
    if (container.value?.requestFullscreen) await container.value.requestFullscreen()
    await (screen.orientation as any)?.lock?.('landscape').catch(() => {})
  } catch (e) {}
}
async function startMatch() {
  if (!nameA.value.trim()) nameA.value = '主队'
  if (!nameB.value.trim()) nameB.value = '客队'
  started.value = true
  keepAwake()
  await nextTick()
  enterFullscreen()
}

onMounted(() => {
  document.addEventListener('visibilitychange', () => { if (wakeLock && started.value) keepAwake() })
})
onUnmounted(() => {
  wakeLock?.release?.()
  if (document.fullscreenElement) document.exitFullscreen().catch(() => {})
})

function addPoint(side: 'A' | 'B') {
  if (finish.value) return
  if (side === 'A') pointA.value++; else pointB.value++

  const w = format.value === '7' ? 7 : 11
  const pa = pointA.value; const pb = pointB.value

  if (format.value === 'golden') {
    // Golden goal: at 10-10, next point wins
    if (pa >= 10 && pb >= 10 && pa !== pb) advanceGame(pa > pb ? 'A' : 'B')
    else if (pa >= 11 && pa - pb >= 2) advanceGame('A')
    else if (pb >= 11 && pb - pa >= 2) advanceGame('B')
  } else {
    // Standard: first to w, must win by 2
    if ((pa >= w || pb >= w) && Math.abs(pa - pb) >= 2) advanceGame(pa > pb ? 'A' : 'B')
    // Also: if beyond w and leading by 2
  }
}

function advanceGame(winner: 'A' | 'B') {
  if (winner === 'A') gameA.value++; else gameB.value++
  pointA.value = 0; pointB.value = 0
}

function undoPoint() {
  if (pointA.value > 0) pointA.value--
  else if (pointB.value > 0) pointB.value--
  else {
    if (gameA.value > 0) { gameA.value-- }
    else if (gameB.value > 0) { gameB.value-- }
  }
}

const finish = ref(false); const winner = ref('')
function checkFinish() {
  if (gameA.value >= gamesToWin.value) { finish.value = true; winner.value = nameA.value }
  else if (gameB.value >= gamesToWin.value) { finish.value = true; winner.value = nameB.value }
}

function toggleNetTouch(side: 'A' | 'B') {
  if (netTouch.value && netTouchSide.value === side) { netTouch.value = false; netTouchSide.value = null }
  else { netTouch.value = true; netTouchSide.value = side }
}
function card(side: 'A' | 'B', color: 'yellow' | 'red') {
  if (color === 'yellow') side === 'A' ? yellowA.value++ : yellowB.value++
  else side === 'A' ? redA.value++ : redB.value++
}
function toggleTimeout(side: 'A' | 'B') {
  if (side === 'A') timeoutA.value = !timeoutA.value
  else timeoutB.value = !timeoutB.value
}
function reset() {
  pointA.value = 0; pointB.value = 0; gameA.value = 0; gameB.value = 0
  yellowA.value = 0; redA.value = 0; yellowB.value = 0; redB.value = 0
  netTouch.value = false; netTouchSide.value = null
  timeoutA.value = false; timeoutB.value = false
  finish.value = false; winner.value = ''
}
function swapSides() {
  [nameA.value, nameB.value] = [nameB.value, nameA.value]
  ;[pointA.value, pointB.value] = [pointB.value, pointA.value]
  ;[gameA.value, gameB.value] = [gameB.value, gameA.value]
  ;[yellowA.value, yellowB.value] = [yellowB.value, yellowA.value]
  ;[redA.value, redB.value] = [redB.value, redA.value]
  ;[timeoutA.value, timeoutB.value] = [timeoutB.value, timeoutA.value]
}
function exitScoreboard() {
  if (document.fullscreenElement) document.exitFullscreen().catch(() => {})
  router.back()
}
</script>

<template>
  <div ref="container" style="min-height: 100vh; background: #0a0a1a; color: #fff;">

    <!-- ========== SETUP SCREEN ========== -->
    <div v-if="!started" style="min-height: 100vh; display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 24px; gap: 24px;">
      <div style="font-size: 28px; font-weight: 900;">📊 记分牌设置</div>

      <div style="width: 100%; max-width: 320px; display: flex; flex-direction: column; gap: 16px;">
        <input v-model="nameA" placeholder="选手A 姓名" style="padding: 14px; background: #111; border: 1px solid #333; border-radius: 10px; color: #fff; font-size: 16px; text-align: center; outline: none;" />
        <div style="text-align: center; color: #666; font-size: 20px;">VS</div>
        <input v-model="nameB" placeholder="选手B 姓名" style="padding: 14px; background: #111; border: 1px solid #333; border-radius: 10px; color: #fff; font-size: 16px; text-align: center; outline: none;" />

        <div style="font-size: 14px; color: #aaa; text-align: center; margin-top: 8px;">计分规则</div>
        <div style="display: flex; gap: 8px;">
          <button @click="format = '11'" style="flex: 1; padding: 12px; border-radius: 10px; border: 2px solid; font-size: 14px; font-weight: 600; cursor: pointer;"
            :style="format === '11' ? { background: '#1989fa', color: '#fff', borderColor: '#1989fa' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
            11 分制<br><span style="font-size:11px;font-weight:400;">先到11且领先2分</span>
          </button>
          <button @click="format = 'golden'" style="flex: 1; padding: 12px; border-radius: 10px; border: 2px solid; font-size: 14px; font-weight: 600; cursor: pointer;"
            :style="format === 'golden' ? { background: '#f1c40f', color: '#000', borderColor: '#f1c40f' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
            金球制<br><span style="font-size:11px;font-weight:400;">10平后下一球决胜</span>
          </button>
          <button @click="format = '7'" style="flex: 1; padding: 12px; border-radius: 10px; border: 2px solid; font-size: 14px; font-weight: 600; cursor: pointer;"
            :style="format === '7' ? { background: '#07c160', color: '#fff', borderColor: '#07c160' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
            抢 7 制<br><span style="font-size:11px;font-weight:400;">先到7分即胜</span>
          </button>
        </div>

        <div style="font-size: 14px; color: #aaa; text-align: center;">大局</div>
        <div style="display: flex; gap: 8px;">
          <button v-for="n in [1,3,5]" :key="n" @click="gamesToWin = n" style="flex: 1; padding: 12px; border-radius: 10px; border: 2px solid; font-size: 14px; font-weight: 600; cursor: pointer;"
            :style="gamesToWin === n ? { background: '#1989fa', color: '#fff', borderColor: '#1989fa' } : { background: 'transparent', color: '#666', borderColor: '#333' }">
            {{ n === 1 ? '单局' : `BO${n}` }}
          </button>
        </div>

        <button @click="startMatch" style="margin-top: 12px; padding: 16px; background: #1989fa; color: #fff; border: none; border-radius: 12px; font-size: 18px; font-weight: 700; cursor: pointer;">
          开始比赛（全屏横屏）
        </button>
      </div>
    </div>

    <!-- ========== SCOREBOARD ========== -->
    <template v-else>
      <div style="display: flex; flex-direction: column; height: 100vh; user-select: none; -webkit-user-select: none;">

        <!-- Top bar -->
        <div style="display: flex; align-items: center; justify-content: space-between; padding: 4px 8px; background: #111; flex-shrink: 0;">
          <button @click="exitScoreboard" style="background: none; border: none; color: #fff; font-size: 13px; cursor: pointer;">&#8592;</button>
          <div style="font-size: 13px; font-weight: 600;">{{ format === '11' ? '11分制' : format === 'golden' ? '金球制' : '抢7制' }} · BO{{ gamesToWin * 2 - 1 || 1 }}</div>
          <div style="display: flex; gap: 4px;">
            <button @click="swapSides" style="background: #333; border: none; color: #fff; padding: 3px 8px; border-radius: 4px; font-size: 11px; cursor: pointer;">⇄换边</button>
            <button @click="reset" style="background: #c0392b; border: none; color: #fff; padding: 3px 8px; border-radius: 4px; font-size: 11px; cursor: pointer;">重置</button>
          </div>
        </div>

        <!-- Winner -->
        <div v-if="finish" style="position: fixed; inset: 0; background: rgba(0,0,0,0.9); z-index: 100; display: flex; flex-direction: column; align-items: center; justify-content: center;">
          <div style="font-size: 56px; font-weight: 900; color: #f1c40f;">{{ winner }}</div>
          <div style="font-size: 28px; margin-top: 4px; color: #aaa;">{{ gameA }} : {{ gameB }}</div>
          <button @click="reset" style="margin-top: 20px; background: #f1c40f; border: none; color: #000; padding: 12px 32px; border-radius: 20px; font-size: 16px; font-weight: 700; cursor: pointer;">重新开始</button>
        </div>

        <!-- Game score -->
        <div style="text-align: center; padding: 2px 0; flex-shrink: 0;">
          <span style="font-size: 26px; font-weight: 900;">{{ gameA }}</span>
          <span style="font-size: 18px; color: #444; margin: 0 6px;">:</span>
          <span style="font-size: 26px; font-weight: 900;">{{ gameB }}</span>
        </div>

        <!-- Cards -->
        <div style="display: flex; justify-content: center; gap: 32px; padding: 2px 0; flex-shrink: 0;">
          <div style="text-align: center;">
            <button @click="card('A','yellow')" style="width:20px;height:26px;background:#f1c40f;border:none;border-radius:3px;cursor:pointer;display:block;margin:0 auto 1px;opacity:0.3;" :style="{opacity:yellowA>0?1:0.3}"></button>
            <button @click="card('A','red')" style="width:20px;height:26px;background:#e74c3c;border:none;border-radius:3px;cursor:pointer;display:block;margin:0 auto 1px;opacity:0.3;" :style="{opacity:redA>0?1:0.3}"></button>
            <span style="font-size:9px;color:#555;">{{ yellowA||redA ? `黄${yellowA}红${redA}` : '' }}</span>
          </div>
          <div style="text-align: center;">
            <button @click="card('B','yellow')" style="width:20px;height:26px;background:#f1c40f;border:none;border-radius:3px;cursor:pointer;display:block;margin:0 auto 1px;opacity:0.3;" :style="{opacity:yellowB>0?1:0.3}"></button>
            <button @click="card('B','red')" style="width:20px;height:26px;background:#e74c3c;border:none;border-radius:3px;cursor:pointer;display:block;margin:0 auto 1px;opacity:0.3;" :style="{opacity:redB>0?1:0.3}"></button>
            <span style="font-size:9px;color:#555;">{{ yellowB||redB ? `黄${yellowB}红${redB}` : '' }}</span>
          </div>
        </div>

        <!-- Main score area -->
        <div style="flex: 1; display: flex; align-items: center; position: relative; min-height: 0;">
          <div @click="addPoint('A')" style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutA?'#1a1a00':'#0d0d2b';cursor:pointer;position:relative;">
            <div v-if="timeoutA" style="position:absolute;top:4px;background:#f1c40f;color:#000;padding:1px 8px;border-radius:8px;font-size:11px;font-weight:700;">暂停</div>
            <div style="font-size:15px;font-weight:700;margin-bottom:4px;">{{ nameA }}</div>
            <div style="font-size:60px;font-weight:900;line-height:1;">{{ pointA }}</div>
            <button @click.stop="toggleTimeout('A')" style="margin-top:8px;background:rgba(255,255,255,0.08);border:1px solid rgba(255,255,255,0.2);color:#888;padding:4px 12px;border-radius:4px;font-size:11px;cursor:pointer;">
              <IconClock :size="12" /> {{ timeoutA ? '继续' : '暂停' }}
            </button>
          </div>
          <div @click="addPoint('B')" style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutB?'#1a1a00':'#1a001a';cursor:pointer;position:relative;">
            <div v-if="timeoutB" style="position:absolute;top:4px;background:#f1c40f;color:#000;padding:1px 8px;border-radius:8px;font-size:11px;font-weight:700;">暂停</div>
            <div style="font-size:15px;font-weight:700;margin-bottom:4px;">{{ nameB }}</div>
            <div style="font-size:60px;font-weight:900;line-height:1;">{{ pointB }}</div>
            <button @click.stop="toggleTimeout('B')" style="margin-top:8px;background:rgba(255,255,255,0.08);border:1px solid rgba(255,255,255,0.2);color:#888;padding:4px 12px;border-radius:4px;font-size:11px;cursor:pointer;">
              <IconClock :size="12" /> {{ timeoutB ? '继续' : '暂停' }}
            </button>
          </div>
        </div>

        <!-- Bottom controls -->
        <div style="padding: 4px 8px 8px; background: #111; flex-shrink: 0;">
          <div style="display: flex; justify-content: center; gap: 10px; margin-bottom: 4px;">
            <button @click="toggleNetTouch('A')" style="padding:5px 14px;border-radius:6px;border:2px solid;font-size:12px;cursor:pointer;"
              :style="netTouch&&netTouchSide==='A'?{background:'#e74c3c',color:'#fff',borderColor:'#e74c3c'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{ nameA }} 擦网</button>
            <button @click="undoPoint" style="background:#333;border:none;color:#aaa;padding:5px 14px;border-radius:6px;font-size:12px;cursor:pointer;">↩ 撤销</button>
            <button @click="toggleNetTouch('B')" style="padding:5px 14px;border-radius:6px;border:2px solid;font-size:12px;cursor:pointer;"
              :style="netTouch&&netTouchSide==='B'?{background:'#e74c3c',color:'#fff',borderColor:'#e74c3c'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{ nameB }} 擦网</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
