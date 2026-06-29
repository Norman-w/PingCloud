<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { CSS2DRenderer, CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js'

const router = useRouter()
const container = ref<HTMLElement>()
const loading = ref(true)

interface H2HRecord { opponent_id: number; opponent_name: string; wins: number; losses: number }
interface H2HPlayer { id: number; name: string; records: H2HRecord[] }

let scene: THREE.Scene, camera: THREE.PerspectiveCamera, renderer: THREE.WebGLRenderer, labelRenderer: CSS2DRenderer, controls: OrbitControls
let animId: number

function animate() {
  animId = requestAnimationFrame(animate)
  controls.update()
  renderer.render(scene, camera)
  labelRenderer.render(scene, camera)
}

function domColor(wins: number, losses: number, side: 'win' | 'lose'): number {
  const total = wins + losses
  if (total === 0) return 0.3
  if (side === 'win') {
    // Green intensity based on win rate
    return 0.3 + (wins / total) * 0.7
  } else {
    // Red intensity based on loss rate
    return 0.3 + (losses / total) * 0.7
  }
}

const error = ref('')

async function init() {
  try {
    const players: H2HPlayer[] = await fetch('/api/headtohead').then(r => r.json())
  } catch(e: any) { error.value = '加载失败'; loading.value = false; return }
  loading.value = false
  if (!container.value || players.length === 0) { error.value = '暂无球员数据'; return }

  const W = container.value.clientWidth
  const H = container.value.clientHeight

  // Scene
  scene = new THREE.Scene()
  scene.background = new THREE.Color(0x0a0a1a)
  scene.fog = new THREE.Fog(0x0a0a1a, 5, 30)

  // Camera
  camera = new THREE.PerspectiveCamera(50, W / H, 0.1, 100)
  camera.position.set(0, 4, 16)
  camera.lookAt(0, 0, 0)

  // Renderers
  renderer = new THREE.WebGLRenderer({ antialias: true })
  renderer.setSize(W, H)
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
  container.value.appendChild(renderer.domElement)

  labelRenderer = new CSS2DRenderer()
  labelRenderer.setSize(W, H)
  labelRenderer.domElement.style.position = 'absolute'
  labelRenderer.domElement.style.top = '0'
  labelRenderer.domElement.style.pointerEvents = 'none'
  container.value.appendChild(labelRenderer.domElement)

  // Controls
  controls = new OrbitControls(camera, renderer.domElement)
  controls.enableDamping = true; controls.dampingFactor = 0.05
  controls.autoRotate = true; controls.autoRotateSpeed = 0.3
  controls.minDistance = 5; controls.maxDistance = 30

  // Lighting
  scene.add(new THREE.AmbientLight(0x404060, 2))
  const light = new THREE.PointLight(0xffffff, 3)
  light.position.set(0, 10, 10)
  scene.add(light)

  // Wireframe sphere
  const wireGeo = new THREE.SphereGeometry(radius, 32, 16)
  const wireMat = new THREE.MeshBasicMaterial({ color: 0x1a2a4a, wireframe: true, transparent: true, opacity: 0.3 })
  scene.add(new THREE.Mesh(wireGeo, wireMat))

  // Fibonacci sphere: evenly distribute points on sphere surface
  const n = players.length
  const playerObjs: { pos: THREE.Vector3; id: number; name: string }[] = []
  const radius = 6
  const phi = Math.PI * (3 - Math.sqrt(5)) // golden angle

  players.forEach((p, i) => {
    const yn = 1 - (i / (n - 1 || 1)) * 2 // -1 to 1
    const r = Math.sqrt(1 - yn * yn)
    const theta = phi * i
    const x = Math.cos(theta) * r * radius
    const y = yn * radius
    const z = Math.sin(theta) * r * radius
    const pos = new THREE.Vector3(x, y, z)
    playerObjs.push({ pos, id: p.id, name: p.name })

    // Sphere
    const geo = new THREE.SphereGeometry(0.25, 32, 32)
    const mat = new THREE.MeshStandardMaterial({ color: 0x1989fa, roughness: 0.3, metalness: 0.6, emissive: 0x112244, emissiveIntensity: 0.5 })
    const sphere = new THREE.Mesh(geo, mat)
    sphere.position.copy(pos)
    scene.add(sphere)

    // Glow
    const glowGeo = new THREE.SphereGeometry(0.35, 32, 32)
    const glowMat = new THREE.MeshBasicMaterial({ color: 0x1989fa, transparent: true, opacity: 0.2 })
    const glow = new THREE.Mesh(glowGeo, glowMat)
    glow.position.copy(pos)
    scene.add(glow)

    // Label
    const div = document.createElement('div')
    div.textContent = p.name
    div.style.color = '#fff'
    div.style.fontSize = '12px'
    div.style.fontWeight = '600'
    div.style.textShadow = '0 0 8px rgba(25,137,250,0.8)'
    const label = new CSS2DObject(div)
    label.position.copy(pos.clone().add(new THREE.Vector3(0, 0.45, 0)))
    scene.add(label)
  })

  // Lines: only show dominant relationships (one arrow per pair, winner→loser)
  const drawn = new Set<string>()
  players.forEach((p, i) => {
    p.records.forEach(r => {
      const j = playerObjs.findIndex(po => po.id === r.opponent_id)
      if (j < 0 || r.wins + r.losses === 0) return

      // Skip if we already drew this pair
      const key = [i, j].sort().join('-')
      if (drawn.has(key)) return
      drawn.add(key)

      // Only draw if there's a clear winner (winRate ≠ 50%)
      const total = r.wins + r.losses
      const winRate = r.wins / total
      if (Math.abs(winRate - 0.5) < 0.01) return // skip 50-50

      // Determine winner/loser
      const winnerIsRow = winRate > 0.5
      const winnerPos = winnerIsRow ? playerObjs[i].pos : playerObjs[j].pos
      const loserPos = winnerIsRow ? playerObjs[j].pos : playerObjs[i].pos

      // Curve from winner toward loser
      const mid = winnerPos.clone().add(loserPos).multiplyScalar(0.5)
      const dir = loserPos.clone().sub(winnerPos).normalize()
      const perp = new THREE.Vector3(-dir.z, 0, dir.x)
      mid.add(perp.clone().multiplyScalar(1.2))

      const curve = new THREE.QuadraticBezierCurve3(winnerPos, mid, loserPos)
      const points = curve.getPoints(50)

      // Tube for thick visible lines
      const tubeGeo = new THREE.TubeGeometry(curve, 40, 0.04, 8, false)
      const intensity = 0.4 + Math.abs(winRate - 0.5) * 1.2
      const color = new THREE.Color().setHSL(0, 0.9, intensity * 0.5 + 0.15)
      const tubeMat = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: intensity, depthWrite: false })
      scene.add(new THREE.Mesh(tubeGeo, tubeMat))

      // Glow dots along curve
      const dotGeo = new THREE.SphereGeometry(0.08, 8, 8)
      const dotMat = new THREE.MeshBasicMaterial({ color: new THREE.Color().setHSL(0, 1, 0.6) })
      for (let k = 0; k < 2; k++) {
        const dot = new THREE.Mesh(dotGeo, dotMat)
        dot.userData = { curve, t: k * 0.5, speed: 0.003 + Math.random() * 0.004 }
        scene.add(dot)
      }
    })
  })

  // Particles animation
  function updateDots() {
    scene.children.forEach(child => {
      if (child.userData?.curve) {
        child.userData.t += child.userData.speed
        if (child.userData.t > 1) child.userData.t -= 1
        child.position.copy(child.userData.curve.getPoint(child.userData.t))
      }
    })
  }

  // Override animate
  const origAnimate = animate
  const _animate = () => {
    animId = requestAnimationFrame(_animate)
    controls.update()
    updateDots()
    renderer.render(scene, camera)
    labelRenderer.render(scene, camera)
  }
  _animate()

  // Handle resize
  window.addEventListener('resize', () => {
    if (!container.value) return
    const w = container.value.clientWidth
    const h = container.value.clientHeight
    camera.aspect = w / h; camera.updateProjectionMatrix()
    renderer.setSize(w, h); labelRenderer.setSize(w, h)
  })

}
onMounted(() => { init() })
onUnmounted(() => { cancelAnimationFrame(animId) })
</script>

<template>
  <div style="position:relative;width:100vw;height:100vh;overflow:hidden;background:#0a0a1a;">
    <!-- Top bar -->
    <div style="position:absolute;top:0;left:0;right:0;z-index:10;padding:12px 16px;display:flex;align-items:center;justify-content:space-between;">
      <button @click="router.back()" style="background:rgba(0,0,0,0.5);border:none;color:#fff;padding:6px 14px;border-radius:8px;font-size:14px;cursor:pointer;">&#8592; 返回</button>
      <span style="color:#fff;font-weight:600;">相生相克</span>
      <span style="font-size:11px;color:#666;">拖拽旋转 · 滚轮缩放</span>
    </div>
    <!-- Legend -->
    <div style="position:absolute;bottom:16px;left:50%;transform:translateX(-50%);z-index:10;display:flex;gap:16px;font-size:11px;color:#aaa;background:rgba(0,0,0,0.5);padding:6px 14px;border-radius:12px;">
      <span>箭头指向被克方 · 颜色越深越压制</span>
    </div>
    <!-- Loading -->
    <div v-if="loading" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#fff;font-size:16px;z-index:5;">加载中...</div>
    <!-- Error -->
    <div v-if="error" style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:#e74c3c;font-size:16px;z-index:5;">{{ error }}</div>
    <!-- Three.js container -->
    <div ref="container" style="width:100%;height:100%;"></div>
  </div>
</template>
