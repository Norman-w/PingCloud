<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { IconSearch, IconUsers, IconUser } from '@tabler/icons-vue'
import { myId } from '../auth'

const model = defineModel<string>('modelValue', { default: '' })
const visible = defineModel<boolean>('visible', { default: false })

const search = ref('')
const rawList = ref<{ id: number; name: string }[]>([])

// Filter out self
const list = computed(() => rawList.value.filter(p => p.id !== myId.value))

async function load(q?: string) {
  const url = q ? `/api/players?q=${encodeURIComponent(q)}` : '/api/players'
  try { const r = await fetch(url); if (r.ok) rawList.value = await r.json() } catch {}
}

watch(visible, async (v) => {
  if (v) { search.value = ''; await load() }
})

async function doSearch() { await load(search.value) }

function select(name: string) { model.value = name; visible.value = false }
</script>

<template>
  <div v-if="visible" style="position:fixed;inset:0;background:rgba(0,0,0,0.45);z-index:3500;display:flex;align-items:flex-end;" @click.self="visible=false">
    <div style="background:#fff;border-radius:20px 20px 0 0;width:100%;max-height:70vh;display:flex;flex-direction:column;">
      <div style="padding:16px 20px;border-bottom:1px solid #f0f0f0;display:flex;align-items:center;gap:12px;">
        <IconUsers :size="20" :stroke-width="2" style="color:#1989fa;flex-shrink:0;" />
        <span style="font-weight:700;font-size:17px;flex:1;">选择陪练</span>
        <button @click="visible=false" style="background:none;border:none;font-size:20px;color:#999;cursor:pointer;padding:4px;">✕</button>
      </div>
      <!-- Search -->
      <div style="padding:12px 16px;">
        <div style="position:relative;">
          <IconSearch :size="16" style="position:absolute;left:10px;top:50%;transform:translateY(-50%);color:#ccc;" />
          <input v-model="search" @input="doSearch" placeholder="搜索球员" style="width:100%;padding:10px 10px 10px 32px;border:1px solid #e8e8e8;border-radius:10px;font-size:15px;outline:none;box-sizing:border-box;background:#f8f9fa;" />
        </div>
      </div>
      <!-- List -->
      <div style="flex:1;overflow-y:auto;padding:0 16px;">
        <!-- Solo option -->
        <div @click="select('独自训练')" style="padding:14px 0;cursor:pointer;border-bottom:1px solid #f5f5f5;display:flex;align-items:center;justify-content:space-between;">
          <div style="display:flex;align-items:center;gap:10px;">
            <div style="width:36px;height:36px;border-radius:10px;background:#e8f4ff;display:flex;align-items:center;justify-content:center;">
              <IconUser :size="18" :stroke-width="2" style="color:#1989fa;" />
            </div>
            <div>
              <div style="font-size:15px;font-weight:600;">独自训练</div>
              <div style="font-size:12px;color:#999;">无需陪练，自己练习</div>
            </div>
          </div>
          <div style="width:22px;height:22px;border-radius:50%;border:2px solid #ddd;display:flex;align-items:center;justify-content:center;" :style="model==='独自训练'?'background:#1989fa;border-color:#1989fa;':''">
            <span v-if="model==='独自训练'" style="color:#fff;font-size:12px;">✓</span>
          </div>
        </div>
        <!-- Player list -->
        <div v-if="list.length===0" style="text-align:center;padding:40px 0;color:#bbb;font-size:14px;">暂无匹配球员</div>
        <div v-for="p in list" :key="p.id" @click="select(p.name)" style="padding:14px 0;cursor:pointer;border-bottom:1px solid #f5f5f5;display:flex;align-items:center;justify-content:space-between;">
          <div>
            <div style="font-size:15px;font-weight:500;">{{ p.name }}</div>
          </div>
          <div style="width:22px;height:22px;border-radius:50%;border:2px solid #ddd;display:flex;align-items:center;justify-content:center;" :style="model===p.name?'background:#1989fa;border-color:#1989fa;':''">
            <span v-if="model===p.name" style="color:#fff;font-size:12px;">✓</span>
          </div>
        </div>
      </div>
      <div style="height:env(safe-area-inset-bottom);"></div>
    </div>
  </div>
</template>
