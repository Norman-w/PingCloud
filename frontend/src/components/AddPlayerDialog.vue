<script setup lang="ts">
import { ref } from 'vue'

defineProps<{ show: boolean; players: { id: number; name: string; current_rating: number }[]; excludeIds: number[]; sessionName: string }>()
const emit = defineEmits<{ (e: 'update:show', v: boolean): void; (e: 'add', playerId: number): void }>()
const selectedId = ref<number | null>(null)
</script>

<template>
  <div v-if="show" style="position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 500; display: flex; align-items: flex-end;" @click.self="emit('update:show', false)">
    <div style="background: #fff; border-radius: 16px 16px 0 0; padding: 24px 16px 80px; width: 100%; max-height: 70vh; overflow-y: auto;">
      <h3 style="text-align: center; margin-bottom: 16px; font-size: 18px;">拉人加入「{{ sessionName }}」</h3>
      <div style="font-size: 13px; color: #969799; margin-bottom: 12px; text-align: center;">已打过的场次保留，没打的重新编排</div>

      <div v-for="p in players.filter(x => !excludeIds.includes(x.id))" :key="p.id"
        @click="selectedId = p.id"
        style="display: flex; align-items: center; padding: 14px 16px; border-radius: 8px; margin-bottom: 4px; cursor: pointer;"
        :style="{ background: selectedId === p.id ? '#e8f4ff' : '#f8f9fa', border: selectedId === p.id ? '2px solid #1989fa' : '2px solid transparent' }">
        <div style="flex: 1;">
          <div style="font-weight: 500;">{{ p.name }}</div>
          <div style="font-size: 13px; color: #969799;">{{ p.current_rating }} 分</div>
        </div>
        <span v-if="selectedId === p.id" style="color: #1989fa; font-weight: 600;">✓</span>
      </div>

      <div v-if="players.filter(x => !excludeIds.includes(x.id)).length === 0"
        style="text-align: center; padding: 20px; color: #969799;">
        没有可加入的球员
      </div>

      <div style="display: flex; gap: 12px; margin-top: 20px;">
        <button @click="emit('update:show', false)" style="flex: 1; padding: 14px; background: #f5f5f5; border: none; border-radius: 24px; font-size: 15px; cursor: pointer;">取消</button>
        <button @click="selectedId && emit('add', selectedId); selectedId = null" :disabled="!selectedId"
          style="flex: 2; padding: 14px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 15px; font-weight: 600; cursor: pointer;"
          :style="{ opacity: selectedId ? 1 : 0.5 }">确认拉入</button>
      </div>
    </div>
  </div>
</template>
