<script setup lang="ts">
import { ref } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import { api } from '../api'

const name = ref('')
const gender = ref('')
const phone = ref('')
const initialRating = ref('')
const referenceRating = ref('')
const submitting = ref(false)

const maleSvg = `<svg viewBox="0 0 48 48" width="28" height="28"><circle cx="24" cy="16" r="10" fill="none" stroke="currentColor" stroke-width="3"/><line x1="24" y1="26" x2="24" y2="44" stroke="currentColor" stroke-width="3"/><line x1="14" y1="34" x2="34" y2="34" stroke="currentColor" stroke-width="3"/></svg>`
const femaleSvg = `<svg viewBox="0 0 48 48" width="28" height="28"><circle cx="24" cy="14" r="10" fill="none" stroke="currentColor" stroke-width="3"/><line x1="24" y1="24" x2="24" y2="36" stroke="currentColor" stroke-width="3"/><line x1="16" y1="30" x2="32" y2="42" stroke="currentColor" stroke-width="3"/><line x1="32" y1="30" x2="16" y2="42" stroke="currentColor" stroke-width="3"/></svg>`

async function onSubmit() {
  if (!name.value.trim()) { showToast('请输入姓名'); return }
  submitting.value = true
  try {
    const ir = initialRating.value ? parseInt(initialRating.value) : undefined
    const rr = referenceRating.value ? parseInt(referenceRating.value) : 0
    const player = await api.createPlayer({
      name: name.value.trim(),
      gender: gender.value,
      phone: phone.value.trim(),
      initial_rating: ir,
      reference_rating: rr,
    } as any)
    showSuccessToast(`${player.name} 添加成功！积分: ${player.current_rating}`)
    name.value = ''
    gender.value = ''
    initialRating.value = ''
    referenceRating.value = ''
  } catch (e: any) { showToast('添加失败: ' + e.message) }
  finally { submitting.value = false }
}
</script>

<template>
  <div style="padding: 80px 20px 120px; background: #f0f2f5; min-height: 100vh;">
    <h3 style="margin-bottom: 20px; font-size: 22px;">添加球员</h3>

    <div style="background: #fff; border-radius: 12px; padding: 16px; box-shadow: 0 2px 12px rgba(0,0,0,0.06);">

      <!-- Name -->
      <div style="margin-bottom: 14px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">姓名</label>
        <input v-model="name" placeholder="请输入球员姓名"
          style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 16px; outline: none; box-sizing: border-box;" />
      </div>

      <!-- Phone -->
      <div style="margin-bottom: 14px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">手机号</label>
        <input v-model="phone" placeholder="用于短信验证登录（选填）"
          style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 16px; outline: none; box-sizing: border-box;" />
      </div>

      <!-- Gender -->
      <div style="margin-bottom: 14px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">性别</label>
        <div style="display: flex; gap: 12px;">
          <button type="button" @click="gender = 'male'"
            style="flex:1;padding:12px;border-radius:10px;border:2px solid;cursor:pointer;background:transparent;display:flex;align-items:center;justify-content:center;gap:8px;font-size:15px;font-weight:600;transition:all 0.2s;"
            :style="gender === 'male' ? {borderColor:'#1989fa',color:'#1989fa',background:'#e8f4ff'} : {borderColor:'#ddd',color:'#999'}">
            <span v-html="maleSvg" :style="{color: gender === 'male' ? '#1989fa' : '#ccc'}"></span>
            男
          </button>
          <button type="button" @click="gender = 'female'"
            style="flex:1;padding:12px;border-radius:10px;border:2px solid;cursor:pointer;background:transparent;display:flex;align-items:center;justify-content:center;gap:8px;font-size:15px;font-weight:600;transition:all 0.2s;"
            :style="gender === 'female' ? {borderColor:'#ee0a24',color:'#ee0a24',background:'#fde8ef'} : {borderColor:'#ddd',color:'#999'}">
            <span v-html="femaleSvg" :style="{color: gender === 'female' ? '#ee0a24' : '#ccc'}"></span>
            女
          </button>
        </div>
      </div>

      <!-- Initial rating -->
      <div style="margin-bottom: 14px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">小圈子排位初始积分</label>
        <input v-model="initialRating" type="number" placeholder="默认 1500"
          style="width: 100%; padding: 12px; border: 1px solid #ebedf0; border-radius: 8px; font-size: 16px; outline: none; box-sizing: border-box;" />
      </div>

      <!-- Reference rating -->
      <div>
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">开球网参考积分</label>
        <input v-model="referenceRating" type="number" placeholder="选填，仅作参考"
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
      不填初始积分默认 1500 分。开球网参考积分仅作参考，不影响排位计算。
    </p>
  </div>
</template>
