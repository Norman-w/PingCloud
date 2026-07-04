<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import * as THREE from 'three'

const router = useRouter()
const container = ref<HTMLElement>()
const loading = ref(true)
const error = ref('')

interface H2HPlayer { id: number; name: string; gender: string; records: { opponent_id: number; opponent_name: string; wins: number; losses: number }[] }

let scene: THREE.Scene, camera: THREE.PerspectiveCamera, renderer: THREE.WebGLRenderer, raycaster: THREE.Raycaster
let animId: number
let camDist = 24, rotY = 0, rotX = 0.4
let autoRotate = true, isDragging = false, dragMoved = 0
let iRotY = 0, iRotX = 0, inertiaOn = false
let spheres: THREE.Mesh[] = []
let lineGroups: { line: THREE.Line; dots: THREE.Sprite[]; winnerIdx: number; loserIdx: number; winnerPos: THREE.Vector3; loserPos: THREE.Vector3; intensity: number }[] = []
let activeIdx = 0
const viewMode = ref<'dominate' | 'feed'>('dominate')
let mode: 'dominate' | 'feed' = 'dominate'
let cycleTimer: any = null
let clickTimeout: any = null
let cycleCount = 0
const playerObjs: { pos: THREE.Vector3; id: number; name: string }[] = []
const nameSprites: { sprite: THREE.Sprite; canvas: HTMLCanvasElement; tex: THREE.CanvasTexture }[] = []
let allPlayers: H2HPlayer[] = []
let radius = 4

function clearLines() {
  lineGroups.forEach(g => {
    scene.remove(g.line)
    g.line.geometry.dispose()
    ;(g.line.material as THREE.LineBasicMaterial).dispose()
    g.dots.forEach(d => {
      scene.remove(d)
      if (d.userData.canvas) d.userData.canvas = null
      if ((d.material as THREE.SpriteMaterial).map) (d.material as THREE.SpriteMaterial).map!.dispose()
      ;(d.material as THREE.SpriteMaterial).dispose()
    })
  })
  lineGroups = []
}

function buildLinesForPlayer(idx: number) {
  clearLines()
  const p = allPlayers[idx]
  if (!p) return

  p.records.forEach(r => {
    const j = playerObjs.findIndex(po => po.id === r.opponent_id)
    if (j < 0 || r.wins + r.losses === 0) return

    const total = r.wins + r.losses
    const winRate = r.wins / total
    if (Math.abs(winRate - 0.5) < 0.01) return
    // 克制: 只显示当前球员克制谁 (winRate>0.5). 福星: 只显示谁是福星 (winRate<0.5)
    if (mode === 'dominate' && winRate < 0.5) return
    if (mode === 'feed' && winRate > 0.5) return

    const wIdx = winRate > 0.5 ? idx : j
    const lIdx = winRate > 0.5 ? j : idx
    const wPos = playerObjs[wIdx].pos.clone()
    const lPos = playerObjs[lIdx].pos.clone()

    const intensity = 0.4 + Math.abs(winRate - 0.5) * 1.2
    const lineGeo = new THREE.BufferGeometry().setFromPoints([wPos, lPos])
    const hsl = mode === 'dominate' ? { h: 0, s: 0.9, l: 0.08 } : { h: 0.33, s: 0.8, l: 0.08 }
    const initColor = new THREE.Color().setHSL(hsl.h, hsl.s, hsl.l)
    const lineMat = new THREE.LineBasicMaterial({ color: initColor, transparent: true, opacity: 0.7, depthTest: false, depthWrite: false })
    const line = new THREE.Line(lineGeo, lineMat)
    scene.add(line)

    const dots: THREE.Sprite[] = []
    for (let k = 0; k < 2; k++) {
      const canvas = document.createElement('canvas'); canvas.width = 64; canvas.height = 64
      const ctx = canvas.getContext('2d')!; ctx.fillStyle = mode === 'dominate' ? '#ff4444' : '#44dd44'; ctx.font = 'bold 40px sans-serif'; ctx.textAlign = 'center'; ctx.textBaseline = 'middle'
      ctx.fillText(mode === 'dominate' ? '克' : '福', 32, 32)
      const tex = new THREE.CanvasTexture(canvas); tex.minFilter = THREE.LinearFilter
      const spriteMat = new THREE.SpriteMaterial({ map: tex, transparent: true, opacity: 0.9, depthTest: false, depthWrite: false })
      const sprite = new THREE.Sprite(spriteMat); sprite.scale.set(0.4, 0.4, 1); sprite.renderOrder = 1
      sprite.position.lerpVectors(wPos, lPos, k * 0.5) // start on the line, not at origin
      sprite.userData = { a: wPos.clone(), b: lPos.clone(), t: k * 0.5, speed: 0.006 + Math.random() * 0.004, lineGrp: true, canvas, tex, char: mode === 'dominate' ? '克' : '福', fromW: mode === 'dominate' }
      scene.add(sprite)
      dots.push(sprite)
    }
    lineGroups.push({ line, dots, winnerIdx: wIdx, loserIdx: lIdx, winnerPos: wPos.clone(), loserPos: lPos.clone(), intensity })
  })
}

function setActive(idx: number) {
  activeIdx = idx
  buildLinesForPlayer(idx)

  // Highlight active sphere
  spheres.forEach((s, i) => {
    const m = s.material as THREE.MeshStandardMaterial
    m.emissiveIntensity = i === idx ? 1.5 : 0.3
    m.emissive = new THREE.Color(i === idx ? 0x44aaff : 0x112244)
  })
}

function onPlayerClick(idx: number) {
  clearTimeout(clickTimeout)
  clearInterval(cycleTimer)
  setActive(idx)
  autoRotate = false
  clickTimeout = setTimeout(() => { autoRotate = true; startCycle() }, 8000)
}

function startCycle() {
  clearInterval(cycleTimer)
  cycleTimer = setInterval(() => {
    activeIdx = (activeIdx + 1) % playerObjs.length
    if (activeIdx === 0) {
      cycleCount++
      if (cycleCount % 2 === 0) {
        mode = mode === 'dominate' ? 'feed' : 'dominate'
        viewMode.value = mode
      }
    }
    setActive(activeIdx)
  }, 3000)
}

function setMode(m: 'dominate' | 'feed') {
  mode = m; viewMode.value = m; clearLines(); setActive(activeIdx)
}

async function init() {
  let players: H2HPlayer[]
  try { players = await fetch('/api/headtohead').then(r => r.json()) }
  catch (e: any) { error.value = '加载失败'; loading.value = false; return }
  allPlayers = players
  loading.value = false
  if (!container.value || players.length === 0) { error.value = '暂无球员'; return }

  const W = container.value.clientWidth; const H = container.value.clientHeight

  scene = new THREE.Scene()
  scene.background = new THREE.Color(0x0a0a1a)
  scene.fog = new THREE.Fog(0x0a0a1a, 15, 50)

  camera = new THREE.PerspectiveCamera(38, W / H, 0.1, 100)
  camera.position.set(0, 2, camDist)
  camera.lookAt(0, 0, 0)

  renderer = new THREE.WebGLRenderer({ antialias: true })
  renderer.setSize(W, H)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  container.value.appendChild(renderer.domElement)

  raycaster = new THREE.Raycaster()

  scene.add(new THREE.AmbientLight(0x404060, 3))
  const light = new THREE.PointLight(0xffffff, 4, 40)
  light.position.set(5, 8, 8)
  scene.add(light)

  // Wireframe sphere
  const wireGeo = new THREE.SphereGeometry(radius, 20, 10)
  const wireMat = new THREE.MeshBasicMaterial({ color: 0x1a3050, wireframe: true, transparent: true, opacity: 0.25 })
  scene.add(new THREE.Mesh(wireGeo, wireMat))

  // Fibonacci sphere distribution
  const n = players.length
  const phi = Math.PI * (3 - Math.sqrt(5))

  players.forEach((p, i) => {
    const yn = 1 - (i / (n - 1 || 1)) * 2
    const r = Math.sqrt(1 - yn * yn)
    const theta = phi * i
    const x = Math.cos(theta) * r * radius
    const y = yn * radius
    const z = Math.sin(theta) * r * radius
    const pos = new THREE.Vector3(x, y, z)
    playerObjs.push({ pos, id: p.id, name: p.name })

    const geo = new THREE.SphereGeometry(0.18, 16, 16)
    const isFemale = p.gender === 'female'
    const mat = new THREE.MeshStandardMaterial({ color: isFemale ? 0xff69b4 : 0x1989fa, roughness: 0.3, metalness: 0.6, emissive: isFemale ? 0x331122 : 0x112244, emissiveIntensity: 0.3 })
    const sphere = new THREE.Mesh(geo, mat)
    sphere.position.copy(pos)
    sphere.userData = { idx: i }
    scene.add(sphere)
    spheres.push(sphere)

    // Glow
    const glowG = new THREE.SphereGeometry(0.26, 16, 16)
    const glowM = new THREE.MeshBasicMaterial({ color: isFemale ? 0xff69b4 : 0x1989fa, transparent: true, opacity: 0.15 })
    const glow = new THREE.Mesh(glowG, glowM)
    glow.position.copy(pos)
    scene.add(glow)

    // Name label as Sprite (renders in 3D pipeline, no jitter)
    const nameCanvas = document.createElement('canvas'); nameCanvas.width = 256; nameCanvas.height = 64
    const nctx = nameCanvas.getContext('2d')!; nctx.fillStyle = '#ffffff'; nctx.font = 'bold 32px sans-serif'; nctx.textAlign = 'center'; nctx.textBaseline = 'middle'; nctx.fillText(p.name, 128, 32)
    const nameTex = new THREE.CanvasTexture(nameCanvas); nameTex.minFilter = THREE.LinearFilter
    const nameMat = new THREE.SpriteMaterial({ map: nameTex, transparent: true, depthTest: false, depthWrite: false })
    const nameSprite = new THREE.Sprite(nameMat); nameSprite.scale.set(1.8, 0.45, 1)
    nameSprite.position.copy(pos.clone().add(new THREE.Vector3(0, 0.6, 0)))
    nameSprite.renderOrder = 2
    scene.add(nameSprite)
    nameSprites.push({ sprite: nameSprite, canvas: nameCanvas, tex: nameTex })
  })

  // Lines built on-demand in setActive (lazy), not all upfront

  // Highlight first player
  setActive(0)
  startCycle()

  // Controls - use movementX/Y for smooth tracking
  let mHist: {dx:number,dy:number,ts:number}[] = []
  container.value.addEventListener('pointerdown', (e: PointerEvent) => {
    e.preventDefault(); isDragging = true; dragMoved = 0; autoRotate = false; inertiaOn = false
    iRotY = 0; iRotX = 0; mHist = []
    container.value!.setPointerCapture(e.pointerId)
  })
  container.value.addEventListener('pointermove', (e: PointerEvent) => {
    if (!isDragging) return
    rotY -= e.movementX * 0.004; rotX += e.movementY * 0.004
    rotX = Math.max(-1.4, Math.min(1.4, rotX))
    dragMoved += Math.abs(e.movementX) + Math.abs(e.movementY)
    mHist.push({dx: e.movementX, dy: e.movementY, ts: Date.now()})
    if (mHist.length > 5) mHist.shift()
  })
  container.value.addEventListener('pointerup', (e: PointerEvent) => {
    if (!isDragging) return; isDragging = false
    container.value!.releasePointerCapture(e.pointerId)
    if (mHist.length >= 2 && dragMoved > 5) {
      const f = mHist[0]; const l = mHist[mHist.length-1]; const dt = l.ts - f.ts
      if (dt > 0) {
        iRotY = -mHist.reduce((s,m)=>s+m.dx,0) / dt * 16 * 0.004
        iRotX = mHist.reduce((s,m)=>s+m.dy,0) / dt * 16 * 0.004
        if (Math.abs(iRotY) > 0.00001 || Math.abs(iRotX) > 0.00001) inertiaOn = true
      }
    }
    if (dragMoved < 5) {
      // Raycaster for sphere click
      const rect = container.value!.getBoundingClientRect()
      const mouse = new THREE.Vector2(
        ((e.clientX - rect.left) / rect.width) * 2 - 1,
        -((e.clientY - rect.top) / rect.height) * 2 + 1
      )
      raycaster.setFromCamera(mouse, camera)
      const hits = raycaster.intersectObjects(spheres)
      if (hits.length > 0) {
        const idx = hits[0].object.userData.idx
        if (idx !== undefined) onPlayerClick(idx)
      }
    }
    if (!inertiaOn) setTimeout(() => { if (!isDragging) autoRotate = true }, 2000)
  })
  container.value.addEventListener('wheel', (e: WheelEvent) => {
    e.preventDefault(); camDist += e.deltaY * 0.04; camDist = Math.max(14, Math.min(40, camDist))
  }, { passive: false })

  function animate() {
    animId = requestAnimationFrame(animate)
    if (inertiaOn) {
      rotY += iRotY; rotX += iRotX
      rotX = Math.max(-1.4, Math.min(1.4, rotX))
      iRotY *= 0.95; iRotX *= 0.95
      if (Math.abs(iRotY) < 0.00001 && Math.abs(iRotX) < 0.00001) { inertiaOn = false; autoRotate = true }
    } else if (autoRotate) {
      rotY += 0.002
    }

    camera.position.x = camDist * Math.cos(rotX) * Math.sin(rotY)
    camera.position.y = camDist * Math.sin(rotX)
    camera.position.z = camDist * Math.cos(rotX) * Math.cos(rotY)
    camera.lookAt(0, 0, 0)

    // Update dots
    scene.children.forEach(c => {
      if (c.userData?.lineGrp) {
        c.userData.t += c.userData.speed
        if (c.userData.t > 1) c.userData.t -= 1
        if (c.userData.fromW !== false) { c.position.lerpVectors(c.userData.a, c.userData.b, c.userData.t) }
        else { c.position.lerpVectors(c.userData.b, c.userData.a, c.userData.t) }
      }
    })

    // Update name sprites (active highlight)
    nameSprites.forEach((ns, i) => {
      const active = i === activeIdx
      const nctx = ns.canvas.getContext('2d')!
      nctx.clearRect(0, 0, 256, 64)
      nctx.fillStyle = active ? '#ffffff' : '#999999'
      nctx.font = active ? 'bold 40px sans-serif' : 'bold 32px sans-serif'
      nctx.textAlign = 'center'; nctx.textBaseline = 'middle'
      nctx.fillText(playerObjs[i].name, 128, 32)
      if (active) {
        nctx.shadowColor = '#ff6600'; nctx.shadowBlur = 8
        nctx.fillText(playerObjs[i].name, 128, 32)
        nctx.shadowBlur = 0
      }
      ns.tex.needsUpdate = true
    })

    renderer.render(scene, camera)
  }
  animate()

  window.addEventListener('resize', () => {
    if (!container.value) return
    const w = container.value.clientWidth; const h = container.value.clientHeight
    camera.aspect = w / h; camera.updateProjectionMatrix(); renderer.setSize(w, h)
  })
}

onMounted(() => init())
onUnmounted(() => { cancelAnimationFrame(animId); clearInterval(cycleTimer); clearTimeout(clickTimeout); clearLines() })
</script>

<template>
  <div style="position:relative;width:100vw;height:100dvh;overflow:hidden;background:#0a0a1a;touch-action:none;-webkit-user-select:none;user-select:none;">
    <div style="position:absolute;top:0;left:0;right:0;z-index:10;padding:10px 16px;display:flex;align-items:center;justify-content:space-between;">
      <button @click="router.back()" style="background:rgba(0,0,0,0.5);border:none;color:#fff;padding:6px 14px;border-radius:8px;font-size:14px;cursor:pointer;">&#8592; 返回</button>
      <div style="display:flex;background:rgba(0,0,0,0.5);border-radius:10px;overflow:hidden;border:1px solid #444;">
        <button @click="setMode('dominate')" style="padding:6px 16px;border:none;font-size:13px;font-weight:700;cursor:pointer;background:transparent;"
          :style="viewMode==='dominate'?{background:'#e74c3c',color:'#fff'}:{background:'transparent',color:'#999'}">克制</button>
        <button @click="setMode('feed')" style="padding:6px 16px;border:none;font-size:13px;font-weight:700;cursor:pointer;background:transparent;"
          :style="viewMode==='feed'?{background:'#07c160',color:'#fff'}:{background:'transparent',color:'#999'}">福星</button>
      </div>
      <span style="font-size:11px;color:#666;">点击球员锁定</span>
    </div>
    <div v-if="loading" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#fff;z-index:5;">加载中...</div>
    <div v-if="error" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#e74c3c;z-index:5;">{{ error }}</div>
    <div ref="container" style="width:100%;height:100%;"></div>
  </div>
</template>
