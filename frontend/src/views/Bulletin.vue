<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { bulletins, markRead, getReadIds, type Bulletin } from '../bulletins'

const router = useRouter()
const list = ref<{ b: Bulletin; read: boolean }[]>([])

onMounted(() => {
  const ids = getReadIds()
  list.value = bulletins.map(b => ({ b, read: ids.includes(b.id) }))
  // Mark all as read on visit
  bulletins.forEach(b => markRead(b.id))
  list.value = bulletins.map(b => ({ b, read: true }))
})
</script>

<template>
  <div style="min-height:100vh;background:#f0f2f5;padding-bottom:80px;">
    <div class="hero">
      <div style="display:flex;align-items:center;gap:12px;">
        <span @click="router.back()" style="cursor:pointer;font-size:22px;">&#8592;</span>
        <div class="hero-title">公告</div>
      </div>
    </div>

    <div style="padding:16px;" v-if="list.length===0">
      <div class="card" style="text-align:center;padding:40px;color:#969799;">暂无公告</div>
    </div>

    <div v-for="{b, read} in list" :key="b.id" class="card" style="margin-bottom:12px;" :style="{opacity:read?0.85:1}">
      <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px;">
        <span v-if="!read" style="width:8px;height:8px;background:#ee0a24;border-radius:50%;flex-shrink:0;"></span>
        <span style="font-size:13px;color:#969799;">{{ b.date }}</span>
      </div>
      <div style="font-weight:700;font-size:16px;margin-bottom:8px;">{{ b.title }}</div>
      <div style="font-size:14px;color:#666;line-height:1.8;white-space:pre-wrap;">{{ b.content }}</div>
    </div>

    <div style="padding:16px;">
      <button @click="router.back()" style="width:100%;padding:16px;background:#1989fa;color:#fff;border:none;border-radius:24px;font-size:17px;font-weight:600;cursor:pointer;">返回首页</button>
    </div>
  </div>
</template>
