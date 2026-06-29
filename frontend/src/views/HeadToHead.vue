<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import * as THREE from 'three'

const router = useRouter()
const container = ref<HTMLElement>()
const loading = ref(true)
const error = ref('')

interface H2HRecord { opponent_id: number; opponent_name: string; wins: number; losses: number }
interface H2HPlayer { id: number; name: string; records: H2HRecord[] }

let scene: THREE.Scene, camera: THREE.PerspectiveCamera, renderer: THREE.WebGLRenderer
let animId: number
let autoRotate = true
let isDragging = false, prevX = 0, prevY = 0
let camDist = 14
let rotY = 0, rotX = 0.4

async function init() {
  let players: H2HPlayer[]
  try {
    players = await fetch('/api/headtohead').then(r => r.json())
  } catch(e: any) { error.value = '加载失败'; loading.value = false; return }
  loading.value = false
  if (!container.value || players.length === 0) { error.value = '暂无球员'; return }

  const W = container.value.clientWidth; const H = container.value.clientHeight
  const radius = 6

  scene = new THREE.Scene()
  scene.background = new THREE.Color(0x0a0a1a)
  scene.fog = new THREE.Fog(0x0a0a1a, 10, 40)

  camera = new THREE.PerspectiveCamera(45, W / H, 0.1, 100)
  camera.position.set(0, 3, 18)
  camera.lookAt(0, 0, 0)

  renderer = new THREE.WebGLRenderer({ antialias: true })
  renderer.setSize(W, H)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  container.value.appendChild(renderer.domElement)

  // Lights
  scene.add(new THREE.AmbientLight(0x404060, 3))
  const light = new THREE.PointLight(0xffffff, 4, 30)
  light.position.set(5, 8, 8)
  scene.add(light)

  // Wireframe sphere
  const wireGeo = new THREE.SphereGeometry(radius, 24, 12)
  const wireMat = new THREE.MeshBasicMaterial({ color: 0x1a2a4a, wireframe: true, transparent: true, opacity: 0.3 })
  scene.add(new THREE.Mesh(wireGeo, wireMat))

  // Fibonacci sphere distribution
  const n = players.length
  const playerObjs: { pos: THREE.Vector3; id: number; name: string }[] = []
  const phi = Math.PI * (3 - Math.sqrt(5))

  const labelDivs: HTMLDivElement[] = []

  players.forEach((p, i) => {
    const yn = 1 - (i / (n - 1 || 1)) * 2
    const r = Math.sqrt(1 - yn * yn)
    const theta = phi * i
    const x = Math.cos(theta) * r * radius
    const y = yn * radius
    const z = Math.sin(theta) * r * radius
    const pos = new THREE.Vector3(x, y, z)
    playerObjs.push({ pos, id: p.id, name: p.name })

    // Player sphere
    const geo = new THREE.SphereGeometry(0.28, 32, 32)
    const mat = new THREE.MeshStandardMaterial({ color: 0x1989fa, roughness: 0.3, metalness: 0.6, emissive: 0x112244, emissiveIntensity: 0.5 })
    const sphere = new THREE.Mesh(geo, mat); sphere.position.copy(pos)
    sphere.userData = { name: p.name, pos: pos.clone() }
    scene.add(sphere)

    // Glow
    const glowG = new THREE.SphereGeometry(0.38, 32, 32)
    const glowM = new THREE.MeshBasicMaterial({ color: 0x1989fa, transparent: true, opacity: 0.2 })
    const glow = new THREE.Mesh(glowG, glowM); glow.position.copy(pos); scene.add(glow)

    // HTML label
    const div = document.createElement('div')
    div.textContent = p.name
    div.style.cssText = 'position:absolute;color:#fff;font-size:12px;font-weight:600;text-shadow:0 0 8px #1989fa;pointer-events:none;transform:translate(-50%,-50%);'
    container.value!.appendChild(div)
    labelDivs.push(div)
  })

  // Only show dominant lines (one per pair, winner→loser)
  const drawn = new Set<string>()
  players.forEach((p, i) => {
    p.records.forEach(r => {
      const j = playerObjs.findIndex(po => po.id === r.opponent_id)
      if (j < 0 || r.wins + r.losses === 0) return
      const key = [i, j].sort().join('-'); if (drawn.has(key)) return; drawn.add(key)

      const total = r.wins + r.losses; const winRate = r.wins / total
      if (Math.abs(winRate - 0.5) < 0.01) return

      const winnerPos = (winRate > 0.5 ? playerObjs[i].pos : playerObjs[j].pos).clone()
      const loserPos = (winRate > 0.5 ? playerObjs[j].pos : playerObjs[i].pos).clone()

      // Straight line from winner to loser (piercing through the sphere)
      const intensity = 0.4 + Math.abs(winRate - 0.5) * 1.2
      const color = new THREE.Color().setHSL(0, 0.9, intensity * 0.45 + 0.2)
      const lineGeo = new THREE.BufferGeometry().setFromPoints([winnerPos, loserPos])
      const lineMat = new THREE.LineBasicMaterial({ color, transparent: true, opacity: 0.5 + Math.abs(winRate - 0.5) * 0.5, depthTest: false })
      scene.add(new THREE.Line(lineGeo, lineMat))

      // Flow dots along straight line
      const dotG = new THREE.SphereGeometry(0.08, 8, 8)
      const dotM = new THREE.MeshBasicMaterial({ color: new THREE.Color().setHSL(0, 1, 0.55) })
      for (let k = 0; k < 2; k++) {
        const dot = new THREE.Mesh(dotG, dotM)
        dot.userData = { a: winnerPos.clone(), b: loserPos.clone(), t: k * 0.5, speed: 0.004 + Math.random() * 0.004 }
        scene.add(dot)
      }
    })
  })

  // Mouse/touch controls (prevent page scroll)
  container.value.addEventListener('pointerdown', (e: PointerEvent) => { e.preventDefault(); isDragging = true; prevX = e.clientX; prevY = e.clientY; autoRotate = false })
  window.addEventListener('pointermove', (e: PointerEvent) => {
    if (!isDragging) return; rotY -= (e.clientX - prevX) * 0.005; rotX += (e.clientY - prevY) * 0.005
    rotX = Math.max(-1.4, Math.min(1.4, rotX)); prevX = e.clientX; prevY = e.clientY
  })
  window.addEventListener('pointerup', () => { isDragging = false; setTimeout(() => { if (!isDragging) autoRotate = true }, 2000) })
  container.value.addEventListener('wheel', (e: WheelEvent) => { e.preventDefault(); camDist += e.deltaY * 0.03; camDist = Math.max(8, Math.min(28, camDist)) }, { passive: false })

  // Animation loop
  function animate() {
    animId = requestAnimationFrame(animate)
    if (autoRotate) rotY += 0.003

    // Spherical coordinates: camera orbits around origin at distance camDist
    camera.position.x = camDist * Math.cos(rotX) * Math.sin(rotY)
    camera.position.y = camDist * Math.sin(rotX)
    camera.position.z = camDist * Math.cos(rotX) * Math.cos(rotY)
    camera.lookAt(0, 0, 0)

    // Update flow dots (linear interpolation)
    scene.children.forEach(c => { if (c.userData?.a) { c.userData.t += c.userData.speed; if (c.userData.t > 1) c.userData.t -= 1; c.position.lerpVectors(c.userData.a, c.userData.b, c.userData.t) } })

    // Update HTML labels
    labelDivs.forEach((div, i) => {
      const wp = playerObjs[i].pos.clone().project(camera)
      div.style.left = ((wp.x + 1) / 2 * W) + 'px'
      div.style.top = ((-wp.y + 1) / 2 * H) + 'px'
      div.style.display = wp.z > 1 ? 'none' : ''
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
onUnmounted(() => { cancelAnimationFrame(animId) })
</script>

<template>
  <div style="position:relative;width:100vw;height:100vh;overflow:hidden;background:#0a0a1a;touch-action:none;-webkit-user-select:none;user-select:none;">
    <!-- Top bar -->
    <div style="position:absolute;top:0;left:0;right:0;z-index:10;padding:10px 16px;display:flex;align-items:center;justify-content:space-between;">
      <button @click="router.back()" style="background:rgba(0,0,0,0.5);border:none;color:#fff;padding:6px 14px;border-radius:8px;font-size:14px;cursor:pointer;">&#8592; 返回</button>
      <span style="color:#fff;font-weight:600;">相生相克 · 3D</span>
      <span style="font-size:11px;color:#666;">拖拽旋转 · 滚轮缩放</span>
    </div>
    <!-- Legend -->
    <div style="position:absolute;bottom:16px;left:50%;transform:translateX(-50%);z-index:10;font-size:11px;color:#aaa;background:rgba(0,0,0,0.5);padding:6px 14px;border-radius:12px;">
      箭头指向被克方 · 颜色越深越压制
    </div>
    <div v-if="loading" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#fff;z-index:5;">加载中...</div>
    <div v-if="error" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#e74c3c;z-index:5;">{{ error }}</div>
    <div ref="container" style="width:100%;height:100%;"></div>
  </div>
</template>
