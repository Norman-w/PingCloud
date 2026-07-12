<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { showSuccessToast, showConfirmDialog } from 'vant'
import { IconArrowLeft, IconEdit, IconTrash, IconMapPin, IconPhone, IconNotes } from '@tabler/icons-vue'

export interface LocData { id?: number; name: string; address: string; phone: string; notes: string; location_type: string }

const props = defineProps<{ loc?: LocData | null }>()
const emit = defineEmits<{ saved: [loc: LocData]; deleted: [id: number] }>()

const visible = defineModel<boolean>('visible', { default: false })
const mode = ref<'view'|'edit'|'create'>('view')

// Form fields
const name = ref(''); const address = ref(''); const phone = ref('')
const notes = ref(''); const locType = ref('球馆')
const typeOptions = ['球馆', '社区活动室', '学校', '公司', '户外', '其他']
const saving = ref(false)

watch(visible, (v) => {
  if (!v) return
  if (props.loc && props.loc.id) {
    // Edit mode
    const l = props.loc
    name.value = l.name; address.value = l.address; phone.value = l.phone
    notes.value = l.notes; locType.value = l.location_type || '球馆'
    mode.value = 'view'
  } else {
    // Create mode
    name.value = ''; address.value = ''; phone.value = ''
    notes.value = ''; locType.value = '球馆'
    mode.value = 'create'
  }
})

const selectedTypeLabel = computed(() => typeOptions.includes(locType.value) ? locType.value : '其他')

function startEdit() { mode.value = 'edit' }
function cancelEdit() {
  if (props.loc?.id) {
    const l = props.loc
    name.value = l.name; address.value = l.address; phone.value = l.phone
    notes.value = l.notes; locType.value = l.location_type || '球馆'
    mode.value = 'view'
  } else {
    visible.value = false
  }
}

async function save() {
  if (!name.value.trim()) return
  saving.value = true
  const data: LocData = { name: name.value.trim(), address: address.value.trim(), phone: phone.value.trim(), notes: notes.value.trim(), location_type: locType.value }
  try {
    const id = props.loc?.id
    const method = id ? 'PUT' : 'POST'
    const url = id ? `/api/locations/${id}` : '/api/locations'
    const r = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data) })
    if (!r.ok) return
    const saved = await r.json()
    showSuccessToast(id ? '已保存' : '场馆已创建')
    emit('saved', saved)
    visible.value = false
  } catch {} finally { saving.value = false }
}

async function doDelete() {
  if (!props.loc?.id) return
  try {
    await showConfirmDialog({ title: '删除场馆', message: `确定删除「${props.loc.name}」吗？` })
    await fetch(`/api/locations/${props.loc.id}`, { method: 'DELETE' })
    showSuccessToast('已删除')
    emit('deleted', props.loc.id)
    visible.value = false
  } catch {}
}
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:4000;display:flex;align-items:flex-end;">
    <div style="background:#fff;border-radius:20px 20px 0 0;width:100%;max-height:85vh;display:flex;flex-direction:column;">
      <!-- Header -->
      <div style="padding:16px 20px;border-bottom:1px solid #f0f0f0;display:flex;align-items:center;gap:12px;flex-shrink:0;">
        <IconArrowLeft v-if="mode==='view'" :size="22" :stroke-width="2" @click="visible=false" style="cursor:pointer;color:#666;" />
        <IconArrowLeft v-else :size="22" :stroke-width="2" @click="cancelEdit" style="cursor:pointer;color:#666;" />
        <span style="font-weight:700;font-size:17px;flex:1;">
          {{ mode === 'create' ? '新建场馆' : mode === 'edit' ? '编辑场馆' : '场馆详情' }}
        </span>
        <button v-if="mode==='view'" @click="startEdit" style="background:none;border:none;padding:6px;cursor:pointer;color:#1989fa;">
          <IconEdit :size="20" :stroke-width="2" />
        </button>
      </div>

      <!-- Body -->
      <div style="flex:1;overflow-y:auto;padding:20px;">
        <!-- View mode -->
        <template v-if="mode==='view'">
          <div style="display:flex;align-items:flex-start;gap:16px;margin-bottom:24px;">
            <div style="width:56px;height:56px;border-radius:14px;background:linear-gradient(135deg,#1989fa,#1e88e5);display:flex;align-items:center;justify-content:center;flex-shrink:0;">
              <IconMapPin :size="28" :stroke-width="2" style="color:#fff;" />
            </div>
            <div style="flex:1;min-width:0;">
              <div style="font-size:20px;font-weight:700;margin-bottom:2px;">{{ name }}</div>
              <div style="font-size:13px;padding:2px 10px;background:#e8f4ff;color:#1989fa;border-radius:6px;display:inline-block;font-weight:500;">{{ selectedTypeLabel }}</div>
            </div>
          </div>
          <div v-if="address" style="margin-bottom:16px;display:flex;align-items:flex-start;gap:10px;">
            <IconMapPin :size="18" :stroke-width="1.5" style="color:#999;flex-shrink:0;margin-top:2px;" />
            <div><div style="font-size:12px;color:#999;margin-bottom:2px;">地址</div><div style="font-size:15px;color:#333;">{{ address }}</div></div>
          </div>
          <div v-if="phone" style="margin-bottom:16px;display:flex;align-items:flex-start;gap:10px;">
            <IconPhone :size="18" :stroke-width="1.5" style="color:#999;flex-shrink:0;margin-top:2px;" />
            <div><div style="font-size:12px;color:#999;margin-bottom:2px;">电话</div><div style="font-size:15px;color:#333;">{{ phone }}</div></div>
          </div>
          <div v-if="notes" style="margin-bottom:24px;display:flex;align-items:flex-start;gap:10px;">
            <IconNotes :size="18" :stroke-width="1.5" style="color:#999;flex-shrink:0;margin-top:2px;" />
            <div><div style="font-size:12px;color:#999;margin-bottom:2px;">备注</div><div style="font-size:14px;color:#666;line-height:1.5;">{{ notes }}</div></div>
          </div>
          <!-- Delete -->
          <button @click="doDelete" style="width:100%;padding:14px;border:1.5px solid #ee0a24;border-radius:12px;background:#fff;color:#ee0a24;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;">
            <IconTrash :size="18" /> 删除场馆
          </button>
        </template>

        <!-- Edit / Create mode -->
        <template v-else>
          <div style="margin-bottom:16px;">
            <label style="font-size:13px;font-weight:600;color:#555;display:block;margin-bottom:6px;">场馆名称 *</label>
            <input v-model="name" placeholder="例：XX社区乒乓球馆" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:15px;outline:none;box-sizing:border-box;" />
          </div>
          <div style="margin-bottom:16px;">
            <label style="font-size:13px;font-weight:600;color:#555;display:block;margin-bottom:6px;">类型</label>
            <div style="display:flex;flex-wrap:wrap;gap:8px;">
              <span v-for="t in typeOptions" :key="t" @click="locType = t"
                style="padding:8px 16px;border-radius:10px;font-size:13px;font-weight:600;cursor:pointer;transition:all .15s;"
                :style="locType===t?'background:#1989fa;color:#fff;':'background:#f0f2f5;color:#666;'">{{ t }}</span>
            </div>
          </div>
          <div style="margin-bottom:16px;">
            <label style="font-size:13px;font-weight:600;color:#555;display:block;margin-bottom:6px;">地址</label>
            <input v-model="address" placeholder="详细地址（可选）" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:15px;outline:none;box-sizing:border-box;" />
          </div>
          <div style="margin-bottom:16px;">
            <label style="font-size:13px;font-weight:600;color:#555;display:block;margin-bottom:6px;">电话</label>
            <input v-model="phone" placeholder="联系电话（可选）" type="tel" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:15px;outline:none;box-sizing:border-box;" />
          </div>
          <div style="margin-bottom:24px;">
            <label style="font-size:13px;font-weight:600;color:#555;display:block;margin-bottom:6px;">备注</label>
            <textarea v-model="notes" placeholder="补充说明（可选）" rows="3" style="width:100%;padding:14px;border:1.5px solid #e0e0e0;border-radius:12px;font-size:14px;outline:none;resize:vertical;box-sizing:border-box;font-family:inherit;"></textarea>
          </div>
          <button @click="save" :disabled="saving || !name.trim()"
            style="width:100%;padding:16px;background:linear-gradient(135deg,#1989fa,#1e88e5);color:#fff;border:none;border-radius:14px;font-size:17px;font-weight:700;cursor:pointer;box-shadow:0 4px 16px rgba(25,137,250,0.25);"
            :style="{opacity:(saving||!name.trim())?0.5:1}">
            {{ saving ? '保存中...' : mode === 'create' ? '创建场馆' : '保存修改' }}
          </button>
        </template>
      </div>
      <div style="height:env(safe-area-inset-bottom);"></div>
    </div>
  </div>
</template>
