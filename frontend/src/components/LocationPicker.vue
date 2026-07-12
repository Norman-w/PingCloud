<script setup lang="ts">
import { ref, watch } from 'vue'
import { IconSearch, IconPlus, IconMapPin } from '@tabler/icons-vue'

const model = defineModel<string>('modelValue', { default: '' })
const visible = defineModel<boolean>('visible', { default: false })

const search = ref('')
const list = ref<{ id: number; name: string }[]>([])
const creating = ref(false)

async function load(q?: string) {
  const url = q ? `/api/locations?q=${encodeURIComponent(q)}` : '/api/locations'
  try { const r = await fetch(url); if (r.ok) list.value = await r.json() } catch {}
}

watch(visible, async (v) => {
  if (v) { search.value = ''; await load() }
})

async function doSearch() { await load(search.value) }

function select(name: string) { model.value = name; visible.value = false }

async function create() {
  if (!search.value.trim()) return
  creating.value = true
  try {
    const r = await fetch('/api/locations', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: search.value.trim() }) })
    if (r.ok) { const l = await r.json(); model.value = l.name; visible.value = false }
  } catch {} finally { creating.value = false }
}
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:3500;display:flex;align-items:flex-end;" @click.self="visible=false">
    <div style="background:#fff;border-radius:20px 20px 0 0;width:100%;max-height:70vh;display:flex;flex-direction:column;">
      <div style="padding:16px 20px;border-bottom:1px solid #f0f0f0;display:flex;align-items:center;gap:12px;">
        <IconMapPin :size="20" :stroke-width="2" style="color:#1989fa;flex-shrink:0;" />
        <span style="font-weight:700;font-size:17px;flex:1;">选择场馆</span>
        <button @click="visible=false" style="background:none;border:none;font-size:20px;color:#999;cursor:pointer;padding:4px;">✕</button>
      </div>
      <!-- Search -->
      <div style="padding:12px 16px;display:flex;gap:8px;">
        <div style="flex:1;position:relative;">
          <IconSearch :size="16" style="position:absolute;left:10px;top:50%;transform:translateY(-50%);color:#ccc;" />
          <input v-model="search" @input="doSearch" placeholder="搜索场馆" style="width:100%;padding:10px 10px 10px 32px;border:1px solid #e8e8e8;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#f8f9fa;" />
        </div>
      </div>
      <!-- List -->
      <div style="flex:1;overflow-y:auto;padding:0 16px;">
        <div v-if="list.length===0" style="text-align:center;padding:40px 0;color:#bbb;">
          <div style="font-size:14px;margin-bottom:12px;">暂无匹配场馆</div>
          <button v-if="search.trim()" @click="create" :disabled="creating" style="padding:10px 24px;background:#1989fa;color:#fff;border:none;border-radius:20px;font-size:14px;font-weight:600;cursor:pointer;display:flex;align-items:center;gap:6px;margin:0 auto;">
            <IconPlus :size="16" /> {{ creating ? '创建中...' : `创建「${search.trim()}」` }}
          </button>
        </div>
        <div v-for="l in list" :key="l.id" @click="select(l.name)" style="padding:14px 0;cursor:pointer;border-bottom:1px solid #f5f5f5;display:flex;align-items:center;justify-content:space-between;">
          <div style="font-size:15px;font-weight:500;">{{ l.name }}</div>
          <div style="width:22px;height:22px;border-radius:50%;border:2px solid #ddd;display:flex;align-items:center;justify-content:center;" :style="model===l.name?'background:#1989fa;border-color:#1989fa;':''">
            <span v-if="model===l.name" style="color:#fff;font-size:12px;">✓</span>
          </div>
        </div>
      </div>
      <!-- Create button at bottom -->
      <div v-if="search.trim() && list.length>0" style="padding:12px 16px;border-top:1px solid #f0f0f0;">
        <button @click="create" :disabled="creating" style="width:100%;padding:12px;background:none;border:2px dashed #1989fa;border-radius:12px;color:#1989fa;font-size:15px;font-weight:600;cursor:pointer;display:flex;align-items:center;justify-content:center;gap:6px;">
          <IconPlus :size="18" /> 新增「{{ search.trim() }}」
        </button>
      </div>
      <!-- Safe area -->
      <div style="height:env(safe-area-inset-bottom);"></div>
    </div>
  </div>
</template>
