<script setup lang="ts">
import { ref, watch } from 'vue'
import { IconSearch, IconPlus, IconMapPin, IconChevronRight } from '@tabler/icons-vue'
import LocationDetail, { type LocData } from './LocationDetail.vue'

const model = defineModel<string>('modelValue', { default: '' })
const visible = defineModel<boolean>('visible', { default: false })

const search = ref('')
const list = ref<LocData[]>([])
const showDetail = ref(false)
const detailLoc = ref<LocData | null>(null)

async function load(q?: string) {
  const url = q ? `/api/locations?q=${encodeURIComponent(q)}` : '/api/locations'
  try { const r = await fetch(url); if (r.ok) list.value = await r.json() } catch {}
}

watch(visible, async (v) => {
  if (v) { search.value = ''; await load() }
})

async function doSearch() { await load(search.value) }

function select(loc: LocData) { model.value = loc.name; visible.value = false }

function openCreate() {
  detailLoc.value = null // null = create mode
  showDetail.value = true
}

function openDetail(loc: LocData) {
  detailLoc.value = loc
  showDetail.value = true
}

function onSaved(loc: LocData) {
  // Auto-select after create/edit
  model.value = loc.name
  visible.value = false
}

function onDeleted() {
  // Refresh list after delete
  load(search.value)
}
</script>

<template>
  <!-- Selection sheet -->
  <div v-if="visible && !showDetail" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:3500;display:flex;align-items:flex-end;" @click.self="visible=false">
    <div style="background:#fff;border-radius:20px 20px 0 0;width:100%;max-height:70vh;display:flex;flex-direction:column;">
      <!-- Header -->
      <div style="padding:16px 20px;border-bottom:1px solid #f0f0f0;display:flex;align-items:center;gap:12px;flex-shrink:0;">
        <IconMapPin :size="20" :stroke-width="2" style="color:#1989fa;flex-shrink:0;" />
        <span style="font-weight:700;font-size:17px;flex:1;">选择场馆</span>
        <button @click="visible=false" style="background:none;border:none;font-size:22px;color:#bbb;cursor:pointer;padding:4px;">✕</button>
      </div>
      <!-- Search -->
      <div style="padding:12px 16px;flex-shrink:0;">
        <div style="position:relative;">
          <IconSearch :size="16" style="position:absolute;left:12px;top:50%;transform:translateY(-50%);color:#ccc;" />
          <input v-model="search" @input="doSearch" placeholder="搜索场馆" style="width:100%;padding:12px 12px 12px 36px;border:1px solid #e8e8e8;border-radius:12px;font-size:15px;outline:none;box-sizing:border-box;background:#f8f9fa;" />
        </div>
      </div>
      <!-- List -->
      <div style="flex:1;overflow-y:auto;padding:0 16px;">
        <div v-if="list.length===0" style="text-align:center;padding:40px 0;color:#bbb;font-size:14px;">
          {{ search.trim() ? '未找到匹配场馆' : '暂无场馆' }}
        </div>
        <div v-for="l in list" :key="l.id" style="display:flex;align-items:center;padding:14px 0;border-bottom:1px solid #f5f5f5;">
          <div @click="select(l)" style="flex:1;cursor:pointer;display:flex;align-items:center;gap:12px;min-width:0;">
            <div style="width:40px;height:40px;border-radius:10px;background:#e8f4ff;display:flex;align-items:center;justify-content:center;flex-shrink:0;">
              <IconMapPin :size="18" :stroke-width="2" style="color:#1989fa;" />
            </div>
            <div style="flex:1;min-width:0;">
              <div style="font-size:15px;font-weight:500;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ l.name }}</div>
              <div style="font-size:12px;color:#999;">{{ l.address || l.location_type || '' }}</div>
            </div>
            <div v-if="model===l.name" style="width:22px;height:22px;border-radius:50%;background:#07c160;display:flex;align-items:center;justify-content:center;flex-shrink:0;"><span style="color:#fff;font-size:12px;font-weight:700;">✓</span></div>
          </div>
          <button @click="openDetail(l)" style="background:none;border:none;padding:8px;cursor:pointer;flex-shrink:0;">
            <IconChevronRight :size="18" style="color:#ccc;" />
          </button>
        </div>
      </div>
      <!-- Create button -->
      <div style="padding:12px 16px;border-top:1px solid #f0f0f0;flex-shrink:0;">
        <button @click="openCreate" style="width:100%;padding:14px;background:none;border:2px dashed #1989fa;border-radius:14px;color:#1989fa;font-size:16px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:8px;">
          <IconPlus :size="20" :stroke-width="2.5" /> 新建场馆
        </button>
      </div>
      <div style="height:env(safe-area-inset-bottom);flex-shrink:0;"></div>
    </div>
  </div>

  <!-- Location detail (create/edit/view) -->
  <LocationDetail v-model:visible="showDetail" :loc="detailLoc" @saved="onSaved" @deleted="onDeleted" />
</template>
