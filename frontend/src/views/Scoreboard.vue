<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

// Pre-fill names from URL params
const started = ref(false)
const nameA = ref((route.query.a as string) || '')
const nameB = ref((route.query.b as string) || '')
const format = ref<'11' | 'golden' | '7'>('11')
const gamesToWin = ref(3)
const firstServer = ref<'A' | 'B'>('A')

const pointA = ref(0); const pointB = ref(0)
const gameA = ref(0); const gameB = ref(0)
const gameHistory = ref<{a:number,b:number}[]>([])
const finish = ref(false); const winner = ref('')

const serveSwitch = computed(() => (pointA.value>=10 && pointB.value>=10) ? 1 : 2)
const totalPoints = computed(() => pointA.value + pointB.value)
const server = computed(() => {
  const h = Math.floor(totalPoints.value / serveSwitch.value)
  return h % 2 === 0 ? firstServer.value : (firstServer.value === 'A' ? 'B' : 'A')
})

const timeoutsA = ref(1); const timeoutsB = ref(1)
const timeoutA = ref(false); const timeoutB = ref(false)
const timeoutCountdown = ref(0); let timer: any = null

const yellowA = ref(0); const redA = ref(0)
const yellowB = ref(0); const redB = ref(0)

const alertMsg = ref(''); let alertTimer: any = null

// Beep fallback (always works)
let audioCtx: AudioContext | null = null
function beep(freq: number, duration: number, count = 1) {
  try {
    if (!audioCtx) audioCtx = new AudioContext()
    let i = 0
    function play() {
      if (i >= count) return
      const osc = audioCtx!.createOscillator()
      const gain = audioCtx!.createGain()
      osc.connect(gain); gain.connect(audioCtx!.destination)
      osc.frequency.value = freq; osc.type = 'sine'
      gain.gain.value = 0.6
      osc.start(); osc.stop(audioCtx!.currentTime + duration / 1000)
      i++; setTimeout(play, duration + 100)
    }
    play()
  } catch(e) {}
}

function alertBeep(msg: string) {
  if (msg.includes('发球')) beep(880, 200, 1)
  else if (msg.includes('暂停')) beep(440, 250, 3)
  else if (msg.includes('擦网')) beep(660, 150, 2)
  else if (msg.includes('黄牌')) beep(520, 100, 2)
  else if (msg.includes('红牌')) beep(330, 200, 3)
  else beep(1000, 150, 1)
}
function showAlert(msg: string, ms = 1500) {
  clearTimeout(alertTimer); alertMsg.value = msg; alertBeep(msg)
  if (ms > 0) alertTimer = setTimeout(() => { alertMsg.value = '' }, ms)
}

const prevServer = ref('')
watch(server, (s) => { if (started.value && prevServer.value && s !== prevServer.value) showAlert(s==='A'?`${nameA.value} 发球`:`${nameB.value} 发球`); prevServer.value = s })

let wakeLock: any = null
async function keepAwake() { try { wakeLock = await (navigator as any).wakeLock?.request?.('screen') } catch(e){} }
onMounted(() => { document.addEventListener('visibilitychange', () => { if (wakeLock) keepAwake() }) })
onUnmounted(() => { wakeLock?.release?.(); clearInterval(timer) })

const container = ref<HTMLElement>()
async function enterFullscreen() { try { if (container.value?.requestFullscreen) await container.value.requestFullscreen(); await (screen.orientation as any)?.lock?.('landscape').catch(()=>{}) } catch(e){} }
async function toggleFullscreen() { if (document.fullscreenElement) { await document.exitFullscreen().catch(()=>{}) } else { if (container.value?.requestFullscreen) await container.value.requestFullscreen(); await (screen.orientation as any)?.lock?.('landscape').catch(()=>{}) } }

async function startMatch() {
  if (!nameA.value.trim()) nameA.value = '主队'
  if (!nameB.value.trim()) nameB.value = '客队'
  timeoutsA.value = 1; timeoutsB.value = 1; started.value = true
  keepAwake()
  // Pre-warm AudioContext + TTS (required for mobile browsers)
  try {
    audioCtx = new AudioContext()
    const o = audioCtx.createOscillator(); const g = audioCtx.createGain()
    o.connect(g); g.connect(audioCtx.destination)
    o.frequency.value = 880; g.gain.value = 0.01
    o.start(); o.stop(audioCtx.currentTime + 0.01)
  } catch(e) {}
  await nextTick(); enterFullscreen()
}

function addPoint(side: 'A'|'B') {
  if (finish.value || timeoutA.value || timeoutB.value) return
  side === 'A' ? pointA.value++ : pointB.value++
  const pa = pointA.value; const pb = pointB.value; const w = format.value === '7' ? 7 : 11
  if (format.value === 'golden') {
    if (pa >= 10 && pb >= 10 && pa !== pb) advanceGame(pa > pb ? 'A' : 'B')
    else if (pa >= 11 && pa - pb >= 2) advanceGame('A')
    else if (pb >= 11 && pb - pa >= 2) advanceGame('B')
  } else {
    if ((pa >= w || pb >= w) && Math.abs(pa - pb) >= 2) advanceGame(pa > pb ? 'A' : 'B')
  }
}
function advanceGame(winner: 'A'|'B') {
  gameHistory.value.push({ a: pointA.value, b: pointB.value })
  winner === 'A' ? gameA.value++ : gameB.value++
  pointA.value = 0; pointB.value = 0; timeoutsA.value = 1; timeoutsB.value = 1
  showAlert(`${nameA.value} ${gameA.value}:${gameB.value} ${nameB.value}`)
}
function undoPoint() {
  if (pointA.value > 0) pointA.value--
  else if (pointB.value > 0) pointB.value--
  else { if (gameA.value > 0) { gameA.value-- } else if (gameB.value > 0) { gameB.value-- } }
}
watch([gameA, gameB], () => { if (gameA.value >= gamesToWin.value) { finish.value = true; winner.value = nameA.value } else if (gameB.value >= gamesToWin.value) { finish.value = true; winner.value = nameB.value } })

function startTimeout(side: 'A'|'B') {
  if (side === 'A' && timeoutsA.value <= 0) return
  if (side === 'B' && timeoutsB.value <= 0) return
  side === 'A' ? (timeoutA.value = true, timeoutsA.value--) : (timeoutB.value = true, timeoutsB.value--)
  timeoutCountdown.value = 60; showAlert(side==='A'?`${nameA.value} 暂停`:`${nameB.value} 暂停`, 0)
  clearInterval(timer)
  timer = setInterval(() => { timeoutCountdown.value--; if (timeoutCountdown.value <= 0) { clearInterval(timer); timeoutA.value = false; timeoutB.value = false; alertMsg.value = '' } }, 1000)
}
function endTimeout() { clearInterval(timer); timeoutA.value = false; timeoutB.value = false; alertMsg.value = '' }
function tapNetTouch() { showAlert('擦网！', 2000) }
function card(side: 'A'|'B', color: 'yellow'|'red') { const n = side==='A'?nameA.value:nameB.value; color==='yellow'?(side==='A'?yellowA.value++:yellowB.value++,showAlert(`黄牌！${n}`,2000)):(side==='A'?redA.value++:redB.value++,showAlert(`红牌！${n}`,3000)) }
function reset() { pointA.value=0;pointB.value=0;gameA.value=0;gameB.value=0;yellowA.value=0;redA.value=0;yellowB.value=0;redB.value=0;timeoutA.value=false;timeoutB.value=false;timeoutsA.value=1;timeoutsB.value=1;finish.value=false;winner.value='';gameHistory.value=[];clearInterval(timer);alertMsg.value='' }
function swapSides() { [nameA.value,nameB.value]=[nameB.value,nameA.value];[pointA.value,pointB.value]=[pointB.value,pointA.value];[gameA.value,gameB.value]=[gameB.value,gameA.value];[yellowA.value,yellowB.value]=[yellowB.value,yellowA.value];[redA.value,redB.value]=[redB.value,redA.value];[timeoutA.value,timeoutB.value]=[timeoutB.value,timeoutA.value];[timeoutsA.value,timeoutsB.value]=[timeoutsB.value,timeoutsA.value];firstServer.value=firstServer.value==='A'?'B':'A' }
function exitScoreboard() { clearInterval(timer); if (document.fullscreenElement) document.exitFullscreen().catch(()=>{}); router.back() }
</script>

<template>
<div ref="container" style="min-height:100vh;background:#0a0a1a;color:#fff;">

  <!-- SETUP -->
  <div v-if="!started" style="min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:24px;gap:20px;">
    <div style="font-size:28px;font-weight:900;">记分牌</div>
    <div style="width:100%;max-width:320px;display:flex;flex-direction:column;gap:14px;">
      <input v-model="nameA" placeholder="选手A" style="padding:14px;background:#111;border:1px solid #333;border-radius:10px;color:#fff;font-size:16px;text-align:center;outline:none;">
      <div style="text-align:center;color:#666;font-size:20px;">VS</div>
      <input v-model="nameB" placeholder="选手B" style="padding:14px;background:#111;border:1px solid #333;border-radius:10px;color:#fff;font-size:16px;text-align:center;outline:none;">
      <div style="font-size:13px;color:#aaa;text-align:center;">先手发球</div>
      <div style="display:flex;gap:8px;">
        <button @click="firstServer='A'" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;" :style="firstServer==='A'?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{nameA||'选手A'}} 先发</button>
        <button @click="firstServer='B'" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;" :style="firstServer==='B'?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{nameB||'选手B'}} 先发</button>
      </div>
      <div style="font-size:13px;color:#aaa;text-align:center;">计分规则</div>
      <div style="display:flex;gap:6px;">
        <button v-for="o in [{k:'11',t:'11分制',s:'先到11领先2分'},{k:'golden',t:'金球制',s:'10平一球决胜'},{k:'7',t:'抢7制',s:'先到7分赢'}]" :key="o.k" @click="format=o.k as any" style="flex:1;padding:10px 6px;border-radius:10px;border:2px solid;font-size:13px;font-weight:600;cursor:pointer;" :style="format===o.k?{background:o.k==='golden'?'#f1c40f':'#1989fa',color:o.k==='golden'?'#000':'#fff',borderColor:o.k==='golden'?'#f1c40f':'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{o.t}}<br><span style="font-size:10px;font-weight:400;">{{o.s}}</span></button>
      </div>
      <div style="font-size:13px;color:#aaa;text-align:center;">大局</div>
      <div style="display:flex;gap:8px;">
        <button v-for="n in [{v:1,l:'单局'},{v:2,l:'三局两胜'},{v:3,l:'五局三胜'},{v:4,l:'七局四胜'}]" :key="n.v" @click="gamesToWin=n.v" style="flex:1;padding:12px;border-radius:10px;border:2px solid;font-size:15px;font-weight:600;cursor:pointer;" :style="gamesToWin===n.v?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'transparent',color:'#666',borderColor:'#333'}">{{n.l}}</button>
      </div>
      <button @click="startMatch" style="margin-top:8px;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:18px;font-weight:700;cursor:pointer;">开始比赛（全屏横屏）</button>
    </div>
  </div>

  <!-- SCOREBOARD -->
  <template v-else>
    <div style="display:flex;flex-direction:column;height:100vh;user-select:none;-webkit-user-select:none;">

      <!-- Alert overlay -->
      <div v-if="alertMsg" style="position:fixed;inset:0;background:rgba(0,0,0,0.85);z-index:200;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:16px;">
        <div style="font-size:64px;font-weight:900;text-align:center;">{{alertMsg}}</div>
        <div v-if="timeoutCountdown>0" style="font-size:96px;font-weight:900;color:#f1c40f;">{{timeoutCountdown}}</div>
        <button v-if="timeoutCountdown>0" @click="endTimeout" style="padding:12px 32px;background:#e74c3c;border:none;color:#fff;border-radius:12px;font-size:16px;font-weight:700;cursor:pointer;">提前结束</button>
      </div>

      <!-- Winner -->
      <div v-if="finish" style="position:fixed;inset:0;background:rgba(0,0,0,0.9);z-index:100;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;">
        <div style="font-size:56px;font-weight:900;color:#f1c40f;">{{winner}} 胜！</div>
        <div style="font-size:28px;color:#aaa;">总比分 {{gameA}}:{{gameB}}</div>
        <div style="font-size:16px;color:#666;display:flex;gap:16px;flex-wrap:wrap;justify-content:center;">
          <span v-for="(g,i) in gameHistory" :key="i" style="background:#222;padding:6px 14px;border-radius:8px;">第{{i+1}}局 {{g.a}}:{{g.b}}</span>
        </div>
        <button @click="reset" style="margin-top:16px;padding:14px 36px;background:#f1c40f;border:none;color:#000;border-radius:20px;font-size:16px;font-weight:700;cursor:pointer;">重新开始</button>
      </div>

      <!-- Top bar -->
      <div style="display:flex;align-items:center;justify-content:space-between;padding:2px 8px;background:#111;flex-shrink:0;">
        <button @click="exitScoreboard" style="background:none;border:none;color:#fff;font-size:12px;cursor:pointer;">&#8592;</button>
        <span style="font-size:12px;font-weight:600;">{{format==='11'?'11分':format==='golden'?'金球':'抢7'}} · {{gamesToWin===1?'单局':gamesToWin===2?'三局两胜':gamesToWin===3?'五局三胜':'七局四胜'}}</span>
        <div style="display:flex;gap:4px;">
          <button @click="toggleFullscreen" style="background:#333;border:none;color:#fff;padding:3px 6px;border-radius:4px;font-size:10px;cursor:pointer;">⛶</button>
          <button @click="swapSides" style="background:#333;border:none;color:#fff;padding:3px 6px;border-radius:4px;font-size:10px;cursor:pointer;">⇄换边</button>
          <button @click="reset" style="background:#c0392b;border:none;color:#fff;padding:3px 6px;border-radius:4px;font-size:10px;cursor:pointer;">重置</button>
        </div>
      </div>

      <!-- Game score -->
      <div style="display:flex;align-items:center;justify-content:center;gap:12px;padding:4px 0;flex-shrink:0;">
        <span style="font-size:44px;font-weight:900;">{{gameA}}</span>
        <span style="font-size:24px;color:#444;">:</span>
        <span style="font-size:44px;font-weight:900;">{{gameB}}</span>
      </div>

      <!-- Cards -->
      <div style="display:flex;justify-content:center;gap:32px;padding:1px 0;flex-shrink:0;">
        <div style="text-align:center;">
          <div style="display:flex;gap:2px;justify-content:center;margin-bottom:1px;">
            <button @click="card('A','yellow')" style="width:16px;height:22px;background:#f1c40f;border:none;border-radius:2px;cursor:pointer;" :style="{opacity:yellowA>0?1:0.3}"></button>
            <button @click="card('A','red')" style="width:16px;height:22px;background:#e74c3c;border:none;border-radius:2px;cursor:pointer;" :style="{opacity:redA>0?1:0.3}"></button>
          </div>
          <span style="font-size:9px;color:#555;">黄{{yellowA}} 红{{redA}}</span>
        </div>
        <div style="text-align:center;">
          <div style="display:flex;gap:2px;justify-content:center;margin-bottom:1px;">
            <button @click="card('B','yellow')" style="width:16px;height:22px;background:#f1c40f;border:none;border-radius:2px;cursor:pointer;" :style="{opacity:yellowB>0?1:0.3}"></button>
            <button @click="card('B','red')" style="width:16px;height:22px;background:#e74c3c;border:none;border-radius:2px;cursor:pointer;" :style="{opacity:redB>0?1:0.3}"></button>
          </div>
          <span style="font-size:9px;color:#555;">黄{{yellowB}} 红{{redB}}</span>
        </div>
      </div>

      <!-- Main score -->
      <div style="flex:1;display:flex;align-items:center;min-height:0;">
        <div @click="addPoint('A')" style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutA?'#1a1a00':'#0d0d2b';cursor:pointer;">
          <div style="font-size:22px;font-weight:700;margin-bottom:8px;display:flex;align-items:center;gap:8px;">
            {{nameA}}<span v-if="server==='A'" style="font-size:24px;animation:serve-pulse 1.5s infinite;">🏓</span>
          </div>
          <div style="font-size:130px;font-weight:900;line-height:1;">{{pointA}}</div>
          <button @click.stop="startTimeout('A')" :disabled="timeoutsA<=0" style="margin-top:12px;padding:6px 16px;border-radius:6px;font-size:12px;cursor:pointer;font-weight:600;" :style="timeoutsA>0?{background:'rgba(255,255,255,0.1)',border:'1px solid rgba(255,255,255,0.3)',color:'#aaa'}:{background:'rgba(255,255,255,0.03)',border:'1px solid rgba(255,255,255,0.1)',color:'#333',cursor:'not-allowed'}">暂停 {{timeoutsA}}</button>
        </div>
        <div @click="addPoint('B')" style="flex:1;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;background:timeoutB?'#1a1a00':'#1a001a';cursor:pointer;">
          <div style="font-size:22px;font-weight:700;margin-bottom:8px;display:flex;align-items:center;gap:8px;">
            {{nameB}}<span v-if="server==='B'" style="font-size:24px;animation:serve-pulse 1.5s infinite;">🏓</span>
          </div>
          <div style="font-size:130px;font-weight:900;line-height:1;">{{pointB}}</div>
          <button @click.stop="startTimeout('B')" :disabled="timeoutsB<=0" style="margin-top:12px;padding:6px 16px;border-radius:6px;font-size:12px;cursor:pointer;font-weight:600;" :style="timeoutsB>0?{background:'rgba(255,255,255,0.1)',border:'1px solid rgba(255,255,255,0.3)',color:'#aaa'}:{background:'rgba(255,255,255,0.03)',border:'1px solid rgba(255,255,255,0.1)',color:'#333',cursor:'not-allowed'}">暂停 {{timeoutsB}}</button>
        </div>
      </div>

      <!-- Bottom -->
      <div style="padding:4px 8px 8px;background:#111;flex-shrink:0;">
        <div style="display:flex;justify-content:center;gap:10px;">
          <button @click="tapNetTouch()" style="padding:6px 18px;border-radius:6px;border:2px solid #333;font-size:12px;cursor:pointer;background:transparent;color:#666;">擦网</button>
          <button @click="undoPoint" style="background:#333;border:none;color:#aaa;padding:6px 12px;border-radius:6px;font-size:12px;cursor:pointer;">↩撤销</button>
        </div>
      </div>
    </div>
  </template>
</div>
</template>

<style>
@keyframes serve-pulse { 0%,100%{opacity:1} 50%{opacity:0.3} }
</style>
