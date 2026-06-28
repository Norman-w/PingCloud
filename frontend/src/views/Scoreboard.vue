<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// Setup
const started = ref(false)
const nameA = ref('')
const nameB = ref('')
const format = ref<'11' | 'golden' | '7'>('11')
const gamesToWin = ref(3)
const firstServer = ref<'A' | 'B'>('A')

// Scores
const pointA = ref(0); const pointB = ref(0)
const gameA = ref(0); const gameB = ref(0)

// Server (auto-switch every 2 points)
const totalPoints = computed(() => pointA.value + pointB.value)
const server = computed(() => {
  const half = Math.floor(totalPoints.value / 2)
  return half % 2 === 0 ? firstServer.value : (firstServer.value === 'A' ? 'B' : 'A')
})

// Timeouts (1 per player per game)
const timeoutsA = ref(1); const timeoutsB = ref(1)
const timeoutA = ref(false); const timeoutB = ref(false)
const timeoutCountdown = ref(0); let timer: any = null

// Cards
const yellowA = ref(0); const redA = ref(0)
const yellowB = ref(0); const redB = ref(0)

// Net touch
const netTouch = ref(false); const netTouchSide = ref<'A' | 'B' | null>(null)

// Full-screen alert
const alertMsg = ref('')
let alertTimer: any = null
function showAlert(msg: string, ms = 1500) {
  clearTimeout(alertTimer)
  alertMsg.value = msg
  if (ms > 0) alertTimer = setTimeout(() => { alertMsg.value = '' }, ms)
}

// Wakelock
let wakeLock: any = null
async function keepAwake() {
  try { wakeLock = await (navigator as any).wakeLock?.request?.('screen') } catch (e) {}
}
onMounted(() => { document.addEventListener('visibilitychange', () => { if (wakeLock) keepAwake() }) })
onUnmounted(() => { wakeLock?.release?.(); clearInterval(timer) })

// Fullscreen
const container = ref<HTMLElement>()
async function enterFullscreen() {
  try {
    if (container.value?.requestFullscreen) await container.value.requestFullscreen()
    await (screen.orientation as any)?.lock?.('landscape').catch(() => {})
  } catch (e) {}
}
async function startMatch() {
  if (!nameA.value.trim()) nameA.value = '主队'
  if (!nameB.value.trim()) nameB.value = '客队'
  timeoutsA.value = 1; timeoutsB.value = 1
  started.value = true
  keepAwake()
  await nextTick()
  enterFullscreen()
}

function addPoint(side: 'A' | 'B') {
  if (finish.value || timeoutA.value || timeoutB.value) return
  if (side === 'A') { pointA.value++ } else { pointB.value++ }
  const pa = pointA.value; const pb = pointB.value; const w = format.value === '7' ? 7 : 11

  if (format.value === 'golden') {
    if (pa >= 10 && pb >= 10 && pa !== pb) advanceGame(pa > pb ? 'A' : 'B')
    else if (pa >= 11 && pa - pb >= 2) advanceGame('A')
    else if (pb >= 11 && pb - pa >= 2) advanceGame('B')
  } else {
    if ((pa >= w || pb >= w) && Math.abs(pa - pb) >= 2) advanceGame(pa > pb ? 'A' : 'B')
  }
}
function advanceGame(winner: 'A' | 'B') {
  if (winner === 'A') { gameA.value++ } else { gameB.value++ }
  pointA.value = 0; pointB.value = 0
  timeoutsA.value = 1; timeoutsB.value = 1
  showAlert(`${nameA.value} ${gameA.value}:${gameB.value} ${nameB.value}`)
}
function undoPoint() {
  if (pointA.value > 0) pointA.value--
  else if (pointB.value > 0) pointB.value--
  else { if (gameA.value > 0) { gameA.value--; pointA.value = 10; pointB.value = 12 } else if (gameB.value > 0) { gameB.value--; pointA.value = 12; pointB.value = 10 } }
}

// Finish
const finish = ref(false); const winner = ref('')
watch([gameA, gameB], () => {
  if (gameA.value >= gamesToWin.value) { finish.value = true; winner.value = nameA.value }
  else if (gameB.value >= gamesToWin.value) { finish.value = true; winner.value = nameB.value }
})

// Timeout
function startTimeout(side: 'A' | 'B') {
  if (side === 'A' && timeoutsA.value <= 0) return
  if (side === 'B' && timeoutsB.value <= 0) return

  if (side === 'A') { timeoutA.value = true; timeoutsA.value-- }
  else { timeoutB.value = true; timeoutsB.value-- }

  timeoutCountdown.value = 60
  showAlert(side === 'A' ? `${nameA.value} 暂停` : `${nameB.value} 暂停`, 0)
  clearInterval(timer)
  timer = setInterval(() => {
    timeoutCountdown.value--
    if (timeoutCountdown.value <= 0) {
      clearInterval(timer)
      timeoutA.value = false; timeoutB.value = false
      alertMsg.value = ''
    }
  }, 1000)
}
function endTimeout() {
  clearInterval(timer)
  timeoutA.value = false; timeoutB.value = false
  alertMsg.value = ''
}

// Net touch
function toggleNetTouch(side: 'A' | 'B') {
  if (netTouch.value && netTouchSide.value === side) { netTouch.value = false; netTouchSide.value = null }
  else {
    netTouch.value = true; netTouchSide.value = side
    showAlert(`擦网！${side === 'A' ? nameA.value : nameB.value}`)
  }
}

// Cards
function card(side: 'A' | 'B', color: 'yellow' | 'red') {
  let n: string
  if (color === 'yellow') {
    if (side === 'A') yellowA.value++; else yellowB.value++
    n = side === 'A' ? nameA.value : nameB.value
    showAlert(`🟡 黄牌！${n}`, 2000)
  } else {
    if (side === 'A') redA.value++; else redB.value++
    n = side === 'A' ? nameA.value : nameB.value
    showAlert(`🔴 红牌！${n}`, 3000)
  }
}

function reset() {
  pointA.value = 0; pointB.value = 0; gameA.value = 0; gameB.value = 0
  yellowA.value = 0; redA.value = 0; yellowB.value = 0; redB.value = 0
  netTouch.value = false; netTouchSide.value = null
  timeoutA.value = false; timeoutB.value = false
  timeoutsA.value = 1; timeoutsB.value = 1
  finish.value = false; winner.value = ''
  clearInterval(timer); alertMsg.value = ''
}
function swapSides() {
  [nameA.value, nameB.value] = [nameB.value, nameA.value]
  ;[pointA.value, pointB.value] = [pointB.value, pointA.value]
  ;[gameA.value, gameB.value] = [gameB.value, gameA.value]
  ;[yellowA.value, yellowB.value] = [yellowB.value, yellowA.value]
  ;[redA.value, redB.value] = [redB.value, redA.value]
  ;[timeoutA.value, timeoutB.value] = [timeoutB.value, timeoutA.value]
  ;[timeoutsA.value, timeoutsB.value] = [timeoutsB.value, timeoutsA.value]
  firstServer.value = firstServer.value === 'A' ? 'B' : 'A'
}
function exitScoreboard() {
  clearInterval(timer)
  if (document.fullscreenElement) document.exitFullscreen().catch(() => {})
  router.back()
}
</script>

<template>
  <div ref="container" style="min-height: 100vh; background: #0a0a1a; color: #fff;">

    <!-- ========== SETUP ========== -->
    <div v-if="!started" style="min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:24px;gap:24px;">
      <div style="font-size:28px;font-weight:900;">记分牌</div>
      <div style="width:100%;max-width:320px;display:flex;flex-direction:column;gap:14px;">
        <input v-model="nameA" placeholder="选手A" style="padding:14px;background:#111;border:1px solid #333;border-radius:10px;color:#fff;font-size:16px;text-align:center;outline:none;">
        <div style="text-align:center;color:#666;font-size:20px;">VS</div>
        <input v-model="nameB" placeholder="选手B" style="padding:14px;background:#111;border:1px solid #333;border-radius:10px;color:#fff;font-size:16px;text-align:center;outline:none;">

        <div style="font-size:13px;color:#aaa;text-align:center;">先手发球</div>
        <div style="display:flex;gap:8px;">
          <button @click="firstServer='A'" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;"
            :style="firstServer==='A'?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{ nameA||'选手A' }} 先发</button>
          <button @click="firstServer='B'" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;"
            :style="firstServer==='B'?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{ nameB||'选手B' }} 先发</button>
        </div>

        <div style="font-size:13px;color:#aaa;text-align:center;">计分规则</div>
        <div style="display:flex;gap:6px;">
          <button v-for="opt in [{k:'11',t:'11分制',s:'先到11领先2分'},{k:'golden',t:'金球制',s:'10平一球决胜'},{k:'7',t:'抢7制',s:'先到7分赢'}]" :key="opt.k"
            @click="format=opt.k as any" style="flex:1;padding:10px 6px;border-radius:10px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;"
            :style="format===opt.k?{background:opt.k==='golden'?'#f1c40f':'#1989fa',color:opt.k==='golden'?'#000':'#fff',borderColor:opt.k==='golden'?'#f1c40f':'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">
            {{ opt.t }}<br><span style="font-size:10px;font-weight:400;">{{ opt.s }}</span>
          </button>
        </div>

        <div style="font-size:13px;color:#aaa;text-align:center;">大局</div>
        <div style="display:flex;gap:8px;">
          <button v-for="n in [{v:1,l:'单局'},{v:2,l:'BO3'},{v:3,l:'BO5'},{v:4,l:'BO7'}]" :key="n.v" @click="gamesToWin=n.v" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;"
            :style="gamesToWin===n.v?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{ n.l }}</button>
        </div>

        <button @click="startMatch" style="margin-top:8px;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:18px;font-weight:700;cursor:pointer;">开始比赛（全屏横屏）</button>
      </div>
    </div>

    <!-- ========== SCOREBOARD ========== -->
    <template v-else>
      <div style="display:flex;flex-direction:column;height:100vh;user-select:none;-webkit-user-select:none;">

        <!-- Alert overlay -->
        <div v-if="alertMsg" style="position:fixed;inset:0;background:rgba(0,0,0,0.8);z-index:200;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:16px;">
          <div style="font-size:48px;font-weight:900;text-align:center;">{{ alertMsg }}</div>
          <div v-if="timeoutCountdown>0" style="font-size:80px;font-weight:900;color:#f1c40f;">{{ timeoutCountdown }}</div>
          <button v-if="timeoutCountdown>0" @click="endTimeout" style="padding:12px 32px;background:#e74c3c;border:none;color:#fff;border-radius:12px;font-size:16px;font-weight:700;cursor:pointer;">提前结束</button>
        </div>

        <!-- Winner -->
        <div v-if="finish" style="position:fixed;inset:0;background:rgba(0,0,0,0.9);z-index:100;display:flex;flex-direction:column;align-items:center;justify-content:center;">
          <div style="font-size:64px;font-weight:900;color:#f1c40f;">{{ winner }} 胜！</div>
          <div style="font-size:32px;margin-top:4px;color:#aaa;">{{ gameA }} : {{ gameB }}</div>
          <button @click="reset" style="margin-top:24px;padding:14px 36px;background:#f1c40f;border:none;color:#000;border-radius:20px;font-size:16px;font-weight:700;cursor:pointer;">重新开始</button>
        </div>

        <!-- Top bar -->
        <div style="display:flex;align-items:center;justify-content:space-between;padding:2px 8px;background:#111;flex-shrink:0;">
          <button @click="exitScoreboard" style="background:none;border:none;color:#fff;font-size:12px;cursor:pointer;">&#8592;</button>
          <span style="font-size:12px;font-weight:600;">{{ format==='11'?'11分':format==='golden'?'金球':'抢7' }} · {{ gamesToWin===1?'单局':`BO${gamesToWin*2-1}` }}</span>
          <div style="display:flex;gap:4px;">
            <button @click="swapSides" style="background:#333;border:none;color:#fff;padding:3px 6px;border-radius:4px;font-size:10px;cursor:pointer;">⇄换边</button>
            <button @click="reset" style="background:#c0392b;border:none;color:#fff;padding:3px 6px;border-radius:4px;font-size:10px;cursor:pointer;">重置</button>
          </div>
        </div>

        <!-- Server + Game score -->
        <div style="display:flex;align-items:center;justify-content:center;gap:20px;padding:4px 0;flex-shrink:0;">
          <div v-if="server===`A`" style="font-size:16px;color:#4fc3f7;font-weight:800;text-align:right;width:80px;animation:serve-pulse 1.5s infinite;">🏓<span style="font-size:14px;"> 发球</span></div>
          <div v-else style="width:60px;"></div>
          <div style="display:flex;align-items:center;gap:12px;">
            <span style="font-size:44px;font-weight:900;">{{ gameA }}</span>
            <span style="font-size:24px;color:#444;">:</span>
            <span style="font-size:44px;font-weight:900;">{{ gameB }}</span>
          </div>
          <div v-if="server===`B`" style="font-size:16px;color:#4fc3f7;font-weight:800;width:80px;animation:serve-pulse 1.5s infinite;">🏓<span style="font-size:14px;"> 发球</span></div>
          <div v-else style="width:60px;"></div>
        </div>

        <!-- Cards row -->
        <div style="display:flex;justify-content:center;gap:32px;padding:1px 0;flex-shrink:0;">
          <div style="text-align:center;">
            <div style="display:flex;gap:2px;justify-content:center;margin-bottom:1px;">
              <button @click="card('A','yellow')" style="width:16px;height:22px;background:#f1c40f;border:none;border-radius:2px;cursor:pointer;opacity:0.3;" :style="{opacity:yellowA>0?1:0.3}"></button>
              <button @click="card('A','red')" style="width:16px;height:22px;background:#e74c3c;border:none;border-radius:2px;cursor:pointer;opacity:0.3;" :style="{opacity:redA>0?1:0.3}"></button>
            </div>
            <span style="font-size:9px;color:#555;">黄{{yellowA}} 红{{redA}}</span>
          </div>
          <div style="text-align:center;">
            <div style="display:flex;gap:2px;justify-content:center;margin-bottom:1px;">
              <button @click="card('B','yellow')" style="width:16px;height:22px;background:#f1c40f;border:none;border-radius:2px;cursor:pointer;opacity:0.3;" :style="{opacity:yellowB>0?1:0.3}"></button>
              <button @click="card('B','red')" style="width:16px;height:22px;background:#e74c3c;border:none;border-radius:2px;cursor:pointer;opacity:0.3;" :style="{opacity:redB>0?1:0.3}"></button>
            </div>
            <span style="font-size:9px;color:#555;">黄{{yellowB}} 红{{redB}}</span>
          </div>
        </div>

        <!-- Main score area - MAXIMIZED -->
        <div style="flex:1;display:flex;align-items:center;position:relative;min-height:0;">
          <div @click="addPoint('A')"
            style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutA?'#1a1a00':'#0d0d2b';cursor:pointer;position:relative;"
            :class="{ 'net-blink': netTouch && netTouchSide === 'A' }">
            <div style="font-size:22px;font-weight:700;margin-bottom:8px;letter-spacing:2px;">{{ nameA }}</div>
            <div style="font-size:130px;font-weight:900;line-height:1;font-variant-numeric:tabular-nums;">{{ pointA }}</div>
            <div style="display:flex;align-items:center;gap:8px;margin-top:12px;">
              <button @click.stop="startTimeout('A')" :disabled="timeoutsA<=0"
                style="padding:6px 16px;border-radius:6px;font-size:12px;cursor:pointer;font-weight:600;"
                :style="timeoutsA>0?{background:'rgba(255,255,255,0.1)',border:'1px solid rgba(255,255,255,0.3)',color:'#aaa'}:{background:'rgba(255,255,255,0.03)',border:'1px solid rgba(255,255,255,0.1)',color:'#333',cursor:'not-allowed'}">
                暂停 {{ timeoutsA }}
              </button>
            </div>
          </div>
          <div @click="addPoint('B')"
            style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutB?'#1a1a00':'#1a001a';cursor:pointer;position:relative;"
            :class="{ 'net-blink': netTouch && netTouchSide === 'B' }">
            <div style="font-size:22px;font-weight:700;margin-bottom:8px;letter-spacing:2px;">{{ nameB }}</div>
            <div style="font-size:130px;font-weight:900;line-height:1;font-variant-numeric:tabular-nums;">{{ pointB }}</div>
            <div style="display:flex;align-items:center;gap:8px;margin-top:12px;">
              <button @click.stop="startTimeout('B')" :disabled="timeoutsB<=0"
                style="padding:6px 16px;border-radius:6px;font-size:12px;cursor:pointer;font-weight:600;"
                :style="timeoutsB>0?{background:'rgba(255,255,255,0.1)',border:'1px solid rgba(255,255,255,0.3)',color:'#aaa'}:{background:'rgba(255,255,255,0.03)',border:'1px solid rgba(255,255,255,0.1)',color:'#333',cursor:'not-allowed'}">
                暂停 {{ timeoutsB }}
              </button>
            </div>
          </div>
        </div>

        <!-- Bottom controls -->
        <div style="padding:4px 8px 8px;background:#111;flex-shrink:0;">
          <div style="display:flex;justify-content:center;gap:10px;">
            <button @click="netTouch=!netTouch" style="padding:6px 18px;border-radius:6px;border:2px solid;font-size:12px;cursor:pointer;"
              :style="netTouch?{background:'#e74c3c',color:'#fff',borderColor:'#e74c3c'}:{background:'transparent',color:'#666',borderColor:'#333'}">擦网</button>
            <button @click="undoPoint" style="background:#333;border:none;color:#aaa;padding:6px 12px;border-radius:6px;font-size:12px;cursor:pointer;">↩ 撤销</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style>
.net-blink { animation: net-blink 0.5s ease-in-out 2; }
@keyframes net-blink {
  0%,100% { background: inherit; }
  50% { background: rgba(231,76,60,0.3); }
}
@keyframes serve-pulse {
  0%,100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>

<style>
.net-blink { animation: net-blink 0.5s ease-in-out 2; }
@keyframes net-blink {
  0%,100% { background: inherit; }
  50% { background: rgba(231,76,60,0.3); }
}
@keyframes serve-pulse {
  0%,100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>
