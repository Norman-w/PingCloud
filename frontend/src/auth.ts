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

export async function doLogin(phone: string, code: string): Promise<{ ok: boolean; msg: string }> {
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
    myId.value = d.player_id
    myName.value = d.player_name
    loaded = true
    return { ok: true, msg: '' }
  } catch {
    return { ok: false, msg: '验证失败' }
  }
}

export function logout() {
  myId.value = 0
  myName.value = ''
  loaded = false
  document.cookie = 'ping_id=;max-age=0;path=/'
}
