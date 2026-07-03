<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { IconUsers, IconHistory, IconLogout, IconPlus, IconEdit, IconTrash, IconShield, IconPencil } from '@tabler/icons-vue'
import { api, type Player } from '../api'

const router = useRouter()

interface Perms { manage_admins: boolean; manage_players: boolean; manage_sessions: boolean; manage_funmatch: boolean; score_matches: boolean; edit_ratings: boolean; view_logs: boolean; view_data: boolean; participate: boolean }
interface AdminUser { id: number; username: string; display_name: string; role: string; group_id: number; group_name: string; permissions: Perms; created_at: string }
interface AdminGroup { id: number; name: string; description: string; permissions: Perms }
interface LogEntry { id: number; admin_id: number; admin_name: string; action: string; target_type: string; target_id: number; detail: string; ip: string; created_at: string }

const me = ref<AdminUser | null>(null)
const users = ref<AdminUser[]>([])
const groups = ref<AdminGroup[]>([])
const logs = ref<LogEntry[]>([])
const players = ref<Player[]>([])
const tab = ref<'users' | 'logs' | 'rating'>('users')

const showDialog = ref(false)
const editUser = ref<AdminUser | null>(null)
const formUser = ref(''); const formPass = ref(''); const formName = ref(''); const formGroup = ref(4)

const showRating = ref(false)
const ratingPlayerId = ref(0)
const ratingNew = ref('')
const ratingReason = ref('')
const ratingPlayerName = ref('')

const permKeys: (keyof Perms)[] = ['manage_admins','manage_players','manage_sessions','manage_funmatch','score_matches','edit_ratings','view_logs','view_data','participate']
const permLabels: Record<string, string> = {
  manage_admins: '管理操作人员', manage_players: '管理球员', manage_sessions: '管理排位赛',
  manage_funmatch: '管理趣味赛', score_matches: '录入比分', edit_ratings: '修改积分',
  view_logs: '查看操作记录', view_data: '查看数据', participate: '参加比赛',
}
const actionLabels: Record<string, string> = {
  create_user: '添加操作人员', update_user: '修改操作人员', delete_user: '删除操作人员',
  adjust_rating: '修改球员积分', login: '登录系统',
}

const selectedGroup = computed(() => groups.value.find(g => g.id === formGroup.value))

onMounted(async () => {
  try {
    const r = await fetch('/api/admin/me', { credentials: 'include' })
    if (!r.ok) { router.replace('/admin/login'); return }
    me.value = await r.json()
    await Promise.all([loadUsers(), loadLogs(), loadGroups()])
    if (me.value?.permissions.edit_ratings) {
      try { players.value = await api.getPlayers() } catch {}
    }
  } catch { router.replace('/admin/login') }
})

async function loadUsers() { try { const r = await fetch('/api/admin/users', { credentials: 'include' }); if (r.ok) users.value = await r.json() } catch {} }
async function loadLogs() { try { const r = await fetch('/api/admin/logs', { credentials: 'include' }); if (r.ok) logs.value = await r.json() } catch {} }
async function loadGroups() { try { const r = await fetch('/api/admin/groups', { credentials: 'include' }); if (r.ok) groups.value = await r.json() } catch {} }

function openCreate() {
  editUser.value = null; formUser.value = ''; formPass.value = ''; formName.value = ''; formGroup.value = 4
  showDialog.value = true
}
function openEdit(u: AdminUser) {
  editUser.value = u; formUser.value = u.username; formPass.value = ''; formName.value = u.display_name; formGroup.value = u.group_id || 4
  showDialog.value = true
}
async function saveUser() {
  try {
    if (editUser.value) {
      await fetch(`/api/admin/users/${editUser.value.id}`, {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ display_name: formName.value, group_id: formGroup.value, password: formPass.value || '' }),
        credentials: 'include',
      })
    } else {
      await fetch('/api/admin/users', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: formUser.value, password: formPass.value, display_name: formName.value, group_id: formGroup.value }),
        credentials: 'include',
      })
    }
    showDialog.value = false; await Promise.all([loadUsers(), loadLogs()])
  } catch {}
}
async function deleteUser(u: AdminUser) {
  if (!confirm(`确定删除 ${u.display_name || u.username}？`)) return
  await fetch(`/api/admin/users/${u.id}`, { method: 'DELETE', credentials: 'include' })
  await Promise.all([loadUsers(), loadLogs()])
}
async function adjustRating() {
  if (!ratingPlayerId.value || !ratingNew.value) return
  await fetch('/api/admin/rating', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ player_id: ratingPlayerId.value, new_rating: parseInt(ratingNew.value), reason: ratingReason.value }),
    credentials: 'include',
  })
  showRating.value = false; ratingReason.value = ''
  try { players.value = await api.getPlayers() } catch {}
}
function openRating(p: Player) {
  ratingPlayerId.value = p.id; ratingPlayerName.value = p.name; ratingNew.value = String(p.current_rating); ratingReason.value = ''
  showRating.value = true
}
async function logout() {
  await fetch('/api/admin/logout', { method: 'POST', credentials: 'include' })
  router.replace('/admin/login')
}
</script>

<template>
  <div style="min-height:100vh;background:#f0f2f5;">
    <!-- 顶栏 -->
    <div style="background:linear-gradient(135deg,#1a1a2e,#1e3a5f);color:#fff;padding:16px 20px;display:flex;align-items:center;justify-content:space-between;">
      <div>
        <div style="font-size:18px;font-weight:700;">🏓 乒云管理后台</div>
        <div v-if="me" style="font-size:12px;opacity:0.7;">{{ me.group_name }}</div>
      </div>
      <button @click="logout" style="background:rgba(255,255,255,0.15);border:none;color:#fff;padding:8px 14px;border-radius:8px;cursor:pointer;display:flex;align-items:center;gap:4px;">
        <IconLogout :size="16" /> 退出登录
      </button>
    </div>

    <!-- 标签页 -->
    <div style="display:flex;background:#fff;border-bottom:1px solid #eee;">
      <button @click="tab='users'" style="flex:1;padding:14px;border:none;background:none;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;"
        :style="{color:tab==='users'?'#1989fa':'#969799',borderBottom:tab==='users'?'2px solid #1989fa':'2px solid transparent'}">
        <IconUsers :size="18" /> 操作人员</button>
      <button v-if="me?.permissions.view_logs" @click="tab='logs'" style="flex:1;padding:14px;border:none;background:none;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;"
        :style="{color:tab==='logs'?'#1989fa':'#969799',borderBottom:tab==='logs'?'2px solid #1989fa':'2px solid transparent'}">
        <IconHistory :size="18" /> 操作记录</button>
      <button v-if="me?.permissions.edit_ratings" @click="tab='rating'" style="flex:1;padding:14px;border:none;background:none;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;"
        :style="{color:tab==='rating'?'#1989fa':'#969799',borderBottom:tab==='rating'?'2px solid #1989fa':'2px solid transparent'}">
        <IconPencil :size="18" /> 球员积分</button>
    </div>

    <!-- ===== 操作人员 ===== -->
    <div v-if="tab==='users'" style="padding:12px 16px;">
      <!-- 权限组说明 -->
      <div v-if="groups.length" style="margin-bottom:12px;">
        <div style="font-size:13px;font-weight:600;color:#666;margin-bottom:6px;">权限组</div>
        <div style="display:flex;flex-direction:column;gap:6px;">
          <div v-for="g in groups" :key="g.id" style="background:#fff;border-radius:10px;padding:10px 12px;box-shadow:0 1px 4px rgba(0,0,0,0.06);cursor:pointer;" @click="formGroup=g.id">
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <span style="font-weight:600;font-size:14px;">{{ g.name }}</span>
              <span style="font-size:11px;color:#c8c9cc;">{{ g.description || '' }}</span>
            </div>
            <div style="display:flex;flex-wrap:wrap;gap:3px;margin-top:4px;">
              <span v-for="k in permKeys.filter(pk => g.permissions[pk])" :key="k"
                style="font-size:10px;padding:1px 6px;border-radius:6px;background:#e8f4ff;color:#1989fa;">
                {{ permLabels[k] }}</span>
              <span v-if="permKeys.filter(pk=>g.permissions[pk]).length===0" style="font-size:10px;color:#ccc;">无特殊权限</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="me?.permissions.manage_admins" style="margin-bottom:12px;">
        <button @click="openCreate" style="width:100%;padding:12px;background:#1989fa;color:#fff;border:none;border-radius:12px;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;">
          <IconPlus :size="18" /> 添加操作人员</button>
      </div>

      <div style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.06);">
        <div v-for="u in users" :key="u.id"
          style="display:flex;align-items:center;padding:14px 16px;border-bottom:1px solid #f5f5f5;">
          <div style="flex:1;">
            <div style="font-size:15px;font-weight:500;">{{ u.display_name || u.username }}</div>
            <div style="font-size:12px;color:#969799;">账号 {{ u.username }} · {{ u.group_name || '未分组' }} · {{ u.created_at?.slice(0,10) }}</div>
            <div style="display:flex;flex-wrap:wrap;gap:2px;margin-top:3px;">
              <span v-for="k in permKeys.filter(pk => u.permissions[pk])" :key="k"
                style="font-size:9px;padding:1px 5px;border-radius:5px;background:#f0f2f5;color:#999;">
                {{ permLabels[k] }}</span>
            </div>
          </div>
          <template v-if="me?.permissions.manage_admins">
            <button @click="openEdit(u)" style="background:none;border:none;color:#1989fa;cursor:pointer;padding:6px;margin-left:4px;"><IconEdit :size="16" /></button>
            <button @click="deleteUser(u)" style="background:none;border:none;color:#e74c3c;cursor:pointer;padding:6px;"><IconTrash :size="16" /></button>
          </template>
        </div>
      </div>
    </div>

    <!-- ===== 操作记录 ===== -->
    <div v-if="tab==='logs'" style="padding:12px 16px;">
      <div style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.06);">
        <div v-for="l in logs" :key="l.id" style="padding:12px 16px;border-bottom:1px solid #f5f5f5;">
          <div style="display:flex;align-items:center;justify-content:space-between;">
            <span style="font-size:14px;font-weight:500;">{{ l.admin_name }}</span>
            <span style="font-size:11px;color:#c8c9cc;">{{ l.created_at?.slice(0,19)?.replace('T',' ') }}</span>
          </div>
          <div style="font-size:13px;color:#666;margin-top:2px;">
            {{ actionLabels[l.action] || l.action }}
            <span v-if="l.detail" style="color:#999;font-size:12px;"> · {{ l.detail }}</span>
          </div>
        </div>
        <div v-if="logs.length===0" style="text-align:center;padding:40px;color:#999;">暂无操作记录</div>
      </div>
    </div>

    <!-- ===== 球员积分 ===== -->
    <div v-if="tab==='rating'" style="padding:12px 16px;">
      <div style="font-size:12px;color:#e74c3c;margin-bottom:8px;background:#fff0f0;padding:8px 12px;border-radius:8px;">
        ⚠ 修改积分会立即生效，操作将被记录
      </div>
      <div style="background:#fff;border-radius:12px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.06);padding:8px 0;">
        <div v-for="p in players" :key="p.id"
          @click="openRating(p)"
          style="display:flex;align-items:center;padding:12px 16px;border-bottom:1px solid #f5f5f5;cursor:pointer;">
          <div style="flex:1;">
            <div style="font-size:15px;font-weight:500;">{{ p.name }}</div>
            <div style="font-size:12px;color:#969799;">开球网参考 {{ p.reference_rating || '-' }}</div>
          </div>
          <div style="font-size:22px;font-weight:800;color:#1989fa;">{{ p.current_rating }}</div>
        </div>
      </div>
    </div>

    <!-- ===== 添加/编辑 弹窗 ===== -->
    <div v-if="showDialog" style="position:fixed;inset:0;background:rgba(0,0,0,0.4);z-index:1000;display:flex;align-items:center;justify-content:center;" @click.self="showDialog=false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:340px;max-height:80vh;overflow-y:auto;">
        <h3 style="margin-bottom:16px;">{{ editUser ? '修改操作人员' : '添加操作人员' }}</h3>

        <div v-if="!editUser" style="margin-bottom:10px;">
          <div style="font-size:13px;color:#969799;margin-bottom:4px;">登录账号</div>
          <input v-model="formUser" placeholder="英文或拼音" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;" />
        </div>
        <div style="margin-bottom:10px;">
          <div style="font-size:13px;color:#969799;margin-bottom:4px;">{{ editUser?'新密码（不填不修改）':'登录密码' }}</div>
          <input v-model="formPass" :placeholder="editUser?'留空则不修改':'至少6位'" type="password" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;" />
        </div>
        <div style="margin-bottom:10px;">
          <div style="font-size:13px;color:#969799;margin-bottom:4px;">显示名称</div>
          <input v-model="formName" placeholder="中文名或昵称" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;" />
        </div>

        <div style="margin-bottom:12px;">
          <div style="font-size:13px;color:#969799;margin-bottom:6px;">权限组</div>
          <div style="display:flex;flex-direction:column;gap:6px;">
            <div v-for="g in groups" :key="g.id"
              @click="formGroup=g.id"
              style="padding:10px 12px;border-radius:10px;border:2px solid;cursor:pointer;"
              :style="{borderColor:formGroup===g.id?'#1989fa':'#eee',background:formGroup===g.id?'#e8f4ff':'#fff'}">
              <div style="display:flex;align-items:center;gap:8px;">
                <div style="width:18px;height:18px;border-radius:50%;border:2px solid;display:flex;align-items:center;justify-content:center;"
                  :style="{borderColor:formGroup===g.id?'#1989fa':'#ccc'}">
                  <div v-if="formGroup===g.id" style="width:10px;height:10px;border-radius:50%;background:#1989fa;"></div>
                </div>
                <span style="font-weight:600;font-size:14px;">{{ g.name }}</span>
                <span v-if="g.description" style="font-size:11px;color:#999;">- {{ g.description }}</span>
              </div>
              <div style="display:flex;flex-wrap:wrap;gap:3px;margin-top:4px;margin-left:26px;">
                <span v-for="k in permKeys.filter(pk => g.permissions[pk])" :key="k"
                  style="font-size:10px;padding:2px 6px;border-radius:6px;background:#e8f8ef;color:#07c160;">
                  ✓ {{ permLabels[k] }}</span>
                <span v-if="permKeys.filter(pk=>g.permissions[pk]).length===0" style="font-size:10px;color:#ccc;">无管理权限</span>
              </div>
            </div>
          </div>
        </div>

        <div style="display:flex;gap:12px;">
          <button @click="showDialog=false" style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:24px;font-size:14px;cursor:pointer;">取消</button>
          <button @click="saveUser" style="flex:2;padding:12px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:14px;font-weight:600;cursor:pointer;">确认{{ editUser?'修改':'添加' }}</button>
        </div>
      </div>
    </div>

    <!-- ===== 修改积分弹窗 ===== -->
    <div v-if="showRating" style="position:fixed;inset:0;background:rgba(0,0,0,0.4);z-index:1000;display:flex;align-items:center;justify-content:center;" @click.self="showRating=false">
      <div style="background:#fff;border-radius:16px;padding:24px;width:320px;">
        <h3 style="margin-bottom:8px;">修改积分 - {{ ratingPlayerName }}</h3>
        <div style="margin-bottom:10px;">
          <div style="font-size:13px;color:#969799;margin-bottom:4px;">新积分</div>
          <input v-model="ratingNew" type="number" placeholder="输入新积分" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:18px;font-weight:700;outline:none;box-sizing:border-box;" />
        </div>
        <div style="margin-bottom:16px;">
          <div style="font-size:13px;color:#969799;margin-bottom:4px;">修改原因（必填）</div>
          <input v-model="ratingReason" placeholder="例如：数据迁移、录入修正" style="width:100%;padding:10px;border:1px solid #ddd;border-radius:8px;font-size:14px;outline:none;box-sizing:border-box;" />
        </div>
        <div style="display:flex;gap:12px;">
          <button @click="showRating=false" style="flex:1;padding:12px;background:#f5f5f5;border:none;border-radius:24px;font-size:14px;cursor:pointer;">取消</button>
          <button @click="adjustRating" style="flex:2;padding:12px;background:#e74c3c;color:#fff;border:none;border-radius:24px;font-size:14px;font-weight:600;cursor:pointer;">确认修改</button>
        </div>
      </div>
    </div>
  </div>
</template>
