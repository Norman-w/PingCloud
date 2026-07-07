<script setup lang="ts">
import { ref } from 'vue'
import { showToast, showSuccessToast } from 'vant'
import { IconHandFinger, IconHandGrab } from '@tabler/icons-vue'
import { api } from '../api'

const name = ref('')
const gender = ref('')
const grip = ref('')
const phone = ref('')
const fhType = ref(''); const fhColor = ref(''); const fhModel = ref('')
const bhType = ref(''); const bhColor = ref(''); const bhModel = ref('')
const initialRating = ref('')
const referenceRating = ref('')
const submitting = ref(false)

const rubberColors = [
  {v:'红',c:'#e74c3c'},{v:'黑',c:'#1a1a1a'},{v:'其他',c:'#999'},
]
const rubberTypes = ['反胶','长胶','生胶','正胶','防弧']

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
      grip: grip.value,
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

      <!-- Grip -->
      <div style="margin-bottom: 14px;">
        <label style="font-size: 14px; color: #646566; display: block; margin-bottom: 6px;">握拍方式</label>
        <div style="display:flex;gap:8px;">
          <button type="button" @click="grip=grip===''?'':''"
            style="flex:1;padding:12px 8px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;display:flex;flex-direction:column;align-items:center;gap:4px;"
            :style="grip===''||!grip?{background:'#e8f4ff',color:'#1989fa',borderColor:'#1989fa'}:{background:'#fff',color:'#999',borderColor:'#eee'}">
            <span style="font-size:20px;">—</span><span>不限</span></button>
          <button type="button" @click="grip=grip==='penhold'?'':'penhold'"
            style="flex:1;padding:12px 8px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;display:flex;flex-direction:column;align-items:center;gap:4px;"
            :style="grip==='penhold'?{background:'#e8f4ff',color:'#1989fa',borderColor:'#1989fa'}:{background:'#fff',color:'#999',borderColor:'#eee'}">
            <IconHandFinger :size="24" :stroke-width="2" /><span>直板</span></button>
          <button type="button" @click="grip=grip==='shakehand'?'':'shakehand'"
            style="flex:1;padding:12px 8px;border-radius:10px;border:2px solid;font-size:14px;font-weight:600;cursor:pointer;display:flex;flex-direction:column;align-items:center;gap:4px;"
            :style="grip==='shakehand'?{background:'#e8f4ff',color:'#1989fa',borderColor:'#1989fa'}:{background:'#fff',color:'#999',borderColor:'#eee'}">
            <IconHandGrab :size="24" :stroke-width="2" /><span>横板</span></button>
        </div>
      </div>

      <!-- Forehand -->
      <div style="margin-bottom:10px;padding:12px;background:#f8f9fa;border-radius:12px;">
        <div style="font-size:13px;font-weight:700;color:#1989fa;margin-bottom:8px;display:flex;align-items:center;gap:6px;"><span style="display:inline-block;width:10px;height:10px;border-radius:2px;background:#1989fa;"></span> 正手胶皮</div>
        <div style="display:flex;flex-wrap:wrap;gap:5px;margin-bottom:6px;">
          <button v-for="t in rubberTypes" :key="'fh'+t" type="button" @click="fhType = fhType===t?'':t"
            style="padding:7px 12px;border-radius:18px;border:1.5px solid;font-size:12px;font-weight:600;cursor:pointer;"
            :style="fhType===t?{background:'#1989fa',color:'#fff',borderColor:'#1989fa'}:{background:'#fff',color:'#666',borderColor:'#ddd'}">
            {{ t }}</button>
        </div>
        <div style="display:flex;flex-wrap:wrap;gap:4px;align-items:center;margin-bottom:4px;">
          <button v-for="c in rubberColors" :key="'fhc'+c.v" type="button" @click="fhColor = fhColor===c.v?'':c.v"
            style="width:30px;height:30px;border-radius:50%;border:2px solid;cursor:pointer;"
            :style="c.v==='其他'
              ? {background:'#fff', borderColor: fhColor===c.v ? '#1989fa' : '#ccc', borderStyle: fhColor===c.v ? 'solid' : 'dashed', boxShadow: fhColor===c.v ? '0 0 0 2px #1989fa40' : 'none'}
              : {background:c.c, borderColor: fhColor===c.v ? '#1989fa' : c.c, boxShadow: fhColor===c.v ? '0 0 0 2px #1989fa40' : 'none'}">
          </button>
          <input v-model="fhModel" placeholder="品牌、型号（选填）" style="flex:1;min-width:80px;padding:6px 10px;border:1px solid #ddd;border-radius:8px;font-size:13px;outline:none;" />
        </div>
      </div>

      <!-- Backhand -->
      <div style="padding:12px;background:#f8f9fa;border-radius:12px;">
        <div style="font-size:13px;font-weight:700;color:#e74c3c;margin-bottom:8px;display:flex;align-items:center;gap:6px;"><span style="display:inline-block;width:10px;height:10px;border-radius:2px;background:#e74c3c;"></span> 反手胶皮</div>
        <div style="display:flex;flex-wrap:wrap;gap:5px;margin-bottom:6px;">
          <button v-for="t in rubberTypes" :key="'bh'+t" type="button" @click="bhType = bhType===t?'':t"
            style="padding:7px 12px;border-radius:18px;border:1.5px solid;font-size:12px;font-weight:600;cursor:pointer;"
            :style="bhType===t?{background:'#e74c3c',color:'#fff',borderColor:'#e74c3c'}:{background:'#fff',color:'#666',borderColor:'#ddd'}">
            {{ t }}</button>
        </div>
        <div style="display:flex;flex-wrap:wrap;gap:4px;align-items:center;margin-bottom:4px;">
          <button v-for="c in rubberColors" :key="'bhc'+c.v" type="button" @click="bhColor = bhColor===c.v?'':c.v"
            style="width:30px;height:30px;border-radius:50%;border:2px solid;cursor:pointer;"
            :style="c.v==='其他'
              ? {background:'#fff', borderColor: bhColor===c.v ? '#e74c3c' : '#ccc', borderStyle: bhColor===c.v ? 'solid' : 'dashed', boxShadow: bhColor===c.v ? '0 0 0 2px #e74c3c40' : 'none'}
              : {background:c.c, borderColor: bhColor===c.v ? '#e74c3c' : c.c, boxShadow: bhColor===c.v ? '0 0 0 2px #e74c3c40' : 'none'}">
          </button>
          <input v-model="bhModel" placeholder="品牌、型号（选填）" style="flex:1;min-width:80px;padding:6px 10px;border:1px solid #ddd;border-radius:8px;font-size:13px;outline:none;" />
        </div>
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
