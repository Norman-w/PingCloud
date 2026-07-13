import { ref } from 'vue'

// Shared auth state across all components
export const myId = ref(0)
export const myName = ref('')

let loaded = false

export async function checkAuth() {
  if (loaded) return
  try {
    const r = await fetch('/api/auth/me')
    if (r.ok) {
      const d = await r.json()
      if (d?.player_id) { myId.value = d.player_id; myName.value = d.player_name }
    }
  } catch {}
  loaded = true
}

// Verify SMS code. Returns { status: "ok"|"need_name", ... }
export async function doLogin(phone: string, code: string): Promise<{ ok: boolean; status?: string; msg: string; player_id?: number; player_name?: string; need_setup?: boolean; phone?: string }> {
  try {
    const r = await fetch('/api/auth/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone, code }),
    })
    if (!r.ok) {
      const t = await r.text()
      return { ok: false, msg: t || '验证码错误' }
    }
    const d = await r.json()
    if (d.status === 'ok') {
      // Phone already bound — login directly
      myId.value = d.player_id
      myName.value = d.player_name
      loaded = true
      return { ok: true, status: 'ok', msg: '', player_id: d.player_id, player_name: d.player_name, need_setup: d.need_setup }
    }
    // Phone not bound — need name
    return { ok: true, status: 'need_name', msg: '', phone: d.phone }
  } catch {
    return { ok: false, msg: '验证失败' }
  }
}

// Check if name matches existing players (requires verified phone)
export async function checkName(phone: string, name: string): Promise<{ matched: boolean; players: Array<{ id: number; name: string; gender: string; phone: string; current_rating: number; has_phone: boolean }> }> {
  const r = await fetch('/api/auth/check-name', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phone, name }),
  })
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}

// Complete login: bind to existing player or create new one
export async function completeLogin(params: {
  phone: string
  action: 'bind' | 'create'
  bind_player_id?: number
  name?: string
  gender?: string
  grip?: string
  initial_rating?: number
}): Promise<{ ok: boolean; msg: string }> {
  try {
    const r = await fetch('/api/auth/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(params),
    })
    if (!r.ok) {
      const t = await r.text()
      return { ok: false, msg: t || '操作失败' }
    }
    const d = await r.json()
    myId.value = d.player_id
    myName.value = d.player_name
    loaded = true
    return { ok: true, msg: '' }
  } catch {
    return { ok: false, msg: '操作失败' }
  }
}

export function logout() {
  myId.value = 0
  myName.value = ''
  loaded = false
  document.cookie = 'ping_id=;max-age=0;path=/'
}
