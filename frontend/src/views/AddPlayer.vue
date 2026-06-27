<script setup lang="ts">
import { ref } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import { api } from '../api'

const name = ref('')
const initialRating = ref('')
const submitting = ref(false)

async function onSubmit() {
  if (!name.value.trim()) { showToast('请输入姓名'); return }
  submitting.value = true
  try {
    const rating = initialRating.value ? parseInt(initialRating.value) : undefined
    const player = await api.createPlayer({ name: name.value.trim(), initial_rating: rating })
    showSuccessToast(`${player.name} 添加成功！积分: ${player.current_rating}`)
    name.value = ''
    initialRating.value = ''
  } catch (e: any) { showToast('添加失败: ' + e.message) }
  finally { submitting.value = false }
}
</script>

<template>
  <div style="padding: 80px 20px 120px; background: #f0f2f5; min-height: 100vh;">
    <h3 style="margin-bottom: 20px; font-size: 22px;">添加球员</h3>

    <div style="background: #fff; border-radius: 12px; padding: 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">
      <div style="margin-bottom: 12px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">姓名</label>
        <input v-model="name" placeholder="请输入球员姓名"
          style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 16px; outline: none; box-sizing: border-box;" />
      </div>
      <div>
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">初始积分</label>
        <input v-model="initialRating" type="number" placeholder="默认 1500"
          style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 16px; outline: none; box-sizing: border-box;" />
      </div>
    </div>

    <div style="margin-top: 20px;">
      <button :disabled="submitting" @click="onSubmit"
        style="width: 100%; padding: 14px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 16px; font-weight: 600; cursor: pointer;">
        {{ submitting ? '添加中...' : '确认添加' }}
      </button>
    </div>

    <p style="margin-top: 16px; font-size: 13px; color: #ad8b00; background: #fffbe6; padding: 12px; border-radius: 8px;">
      不填初始积分默认 1500 分。积分会在比赛后自动更新。
    </p>
  </div>
</template>
