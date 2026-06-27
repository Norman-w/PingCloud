<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  playerAName: string
  playerBName: string
  initialScoreA?: number
  initialScoreB?: number
}>()

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
  (e: 'submit', scoreA: number, scoreB: number): void
}>()

const scoreA = ref('')
const scoreB = ref('')
const format = ref(3)
const submitting = ref(false)

function formatLabel(w: number) {
  if (w === 2) return '三局两胜'
  if (w === 4) return '七局四胜'
  return '五局三胜'
}

function quickScores(): [number, number][] {
  const w = format.value
  const scores: [number, number][] = []
  for (let l = 0; l < w; l++) scores.push([w, l])
  for (let l = 0; l < w; l++) scores.push([l, w])
  return scores
}

function onCancel() {
  scoreA.value = ''; scoreB.value = ''
  emit('update:show', false)
}

async function onSubmit() {
  const sa = parseInt(scoreA.value), sb = parseInt(scoreB.value)
  if (isNaN(sa) || isNaN(sb) || sa === sb) return
  submitting.value = true
  emit('submit', sa, sb)
  // parent should close dialog and set submitting = false via prop change
  setTimeout(() => { submitting.value = false }, 500)
}

watch(() => props.show, (v) => {
  if (v) {
    scoreA.value = props.initialScoreA != null ? String(props.initialScoreA) : ''
    scoreB.value = props.initialScoreB != null ? String(props.initialScoreB) : ''
  }
})
</script>

<template>
  <div v-if="show" style="position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 500; display: flex; align-items: center; justify-content: center;" @click.self="onCancel">
    <div style="background: #fff; border-radius: 16px; padding: 24px 20px; width: 90%; max-width: 360px;">
      <h3 style="text-align: center; margin-bottom: 8px; font-size: 18px;">录入比分</h3>

      <div style="display: flex; gap: 6px; margin-bottom: 16px;">
        <button v-for="w in [2,3,4]" :key="w" @click="format = w"
          style="flex: 1; padding: 8px 0; border-radius: 8px; font-size: 12px; font-weight: 600; border: 2px solid; cursor: pointer;"
          :style="format === w ? { background: '#1989fa', color: '#fff', borderColor: '#1989fa' } : { background: '#fff', color: '#666', borderColor: '#ebedf0' }">
          {{ formatLabel(w) }}
        </button>
      </div>

      <div style="display: flex; align-items: center; justify-content: center; gap: 12px;">
        <div style="text-align: center;">
          <div style="font-weight: 600;">{{ playerAName }}</div>
          <input v-model="scoreA" type="number" min="0" :max="format" placeholder="0"
            style="width: 60px; height: 48px; text-align: center; font-size: 28px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 12px; outline: none; margin-top: 8px;" />
        </div>
        <div style="font-size: 28px; font-weight: 800; color: #969799; padding-top: 20px;">:</div>
        <div style="text-align: center;">
          <div style="font-weight: 600;">{{ playerBName }}</div>
          <input v-model="scoreB" type="number" min="0" :max="format" placeholder="0"
            style="width: 60px; height: 48px; text-align: center; font-size: 28px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 12px; outline: none; margin-top: 8px;" />
        </div>
      </div>

      <div style="display: flex; gap: 6px; flex-wrap: wrap; justify-content: center; margin-top: 16px;">
        <button v-for="(s, i) in quickScores()" :key="i"
          @click="scoreA=String(s[0]);scoreB=String(s[1])"
          style="padding: 6px 14px; border-radius: 16px; border: none; font-size: 14px; font-weight: 500; cursor: pointer;"
          :style="s[0] > s[1] ? { background: '#e8f8ef', color: '#07c160' } : { background: '#fde8e8', color: '#ee0a24' }">
          {{ s[0] }}:{{ s[1] }}
        </button>
      </div>

      <div style="display: flex; gap: 12px; margin-top: 20px;">
        <button @click="onCancel" style="flex: 1; padding: 14px; background: #f5f5f5; border: none; border-radius: 24px; font-size: 15px; cursor: pointer;">取消</button>
        <button @click="onSubmit" :disabled="submitting" style="flex: 2; padding: 14px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 15px; font-weight: 600; cursor: pointer;">
          {{ submitting ? '提交中...' : '确认提交' }}
        </button>
      </div>
    </div>
  </div>
</template>
