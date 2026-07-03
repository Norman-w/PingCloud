<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  show: boolean
  maleName: string
  femaleName: string
  handicapPoints: number
}>()

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
  (e: 'submit', g1m: number, g1f: number, g2m: number, g2f: number, g3m?: number, g3f?: number): void
}>()

const g1m = ref(''), g1f = ref('')
const g2m = ref(''), g2f = ref('')
const g3m = ref(''), g3f = ref('')

function quickBO3(): [string, string, string, string, string, string][] {
  return [
    ['11', '8', '11', '9', '', ''],
    ['9', '11', '11', '8', '11', '7'],
    ['11', '13', '11', '8', '', ''],
    ['8', '11', '12', '10', '11', '9'],
    ['11', '6', '12', '10', '', ''],
    ['7', '11', '11', '5', '11', '8'],
  ]
}

function onCancel() {
  g1m.value = ''; g1f.value = ''
  g2m.value = ''; g2f.value = ''
  g3m.value = ''; g3f.value = ''
  emit('update:show', false)
}

function onSubmit() {
  const gm1 = parseInt(g1m.value), gf1 = parseInt(g1f.value)
  const gm2 = parseInt(g2m.value), gf2 = parseInt(g2f.value)
  if (isNaN(gm1) || isNaN(gf1) || isNaN(gm2) || isNaN(gf2)) return
  if (gm1 === gf1 || gm2 === gf2) return

  const gm3 = parseInt(g3m.value), gf3 = parseInt(g3f.value)
  if (!isNaN(gm3) && !isNaN(gf3) && gm3 !== gf3) {
    emit('submit', gm1, gf1, gm2, gf2, gm3, gf3)
  } else {
    emit('submit', gm1, gf1, gm2, gf2)
  }
}

watch(() => props.show, (v) => {
  if (v) {
    g1m.value = ''; g1f.value = ''
    g2m.value = ''; g2f.value = ''
    g3m.value = ''; g3f.value = ''
  }
})
</script>

<template>
  <div v-if="show" style="position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 500; display: flex; align-items: center; justify-content: center;" @click.self="onCancel">
    <div style="background: #fff; border-radius: 16px; padding: 24px 20px; width: 90%; max-width: 360px;">
      <h3 style="text-align: center; margin-bottom: 4px; font-size: 18px;">录入比分</h3>
      <div v-if="handicapPoints > 0" style="text-align:center;font-size:12px;color:#ee0a24;margin-bottom:8px;">⚠ 让{{ handicapPoints }}分</div>

      <!-- Game 1 -->
      <div style="font-size:12px;font-weight:600;color:#969799;">第1局</div>
      <div style="display: flex; align-items: center; justify-content: center; gap: 12px;">
        <div style="text-align: center; flex: 1;">
          <div style="font-weight: 500; font-size: 13px;">{{ maleName }}</div>
          <input v-model="g1m" type="number" min="0" placeholder="0"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none; margin-top: 4px;" />
        </div>
        <div style="font-size: 24px; font-weight: 800; color: #969799; padding-top: 16px;">:</div>
        <div style="text-align: center; flex: 1;">
          <div style="font-weight: 500; font-size: 13px;">{{ femaleName }}</div>
          <input v-model="g1f" type="number" min="0" placeholder="0"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none; margin-top: 4px;" />
        </div>
      </div>

      <!-- Game 2 -->
      <div style="font-size:12px;font-weight:600;color:#969799;margin-top:10px;">第2局</div>
      <div style="display: flex; align-items: center; justify-content: center; gap: 12px;">
        <div style="text-align: center; flex: 1;">
          <input v-model="g2m" type="number" min="0" placeholder="0"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
        <div style="font-size: 24px; font-weight: 800; color: #969799;">:</div>
        <div style="text-align: center; flex: 1;">
          <input v-model="g2f" type="number" min="0" placeholder="0"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
      </div>

      <!-- Game 3 (always shown, optional) -->
      <div style="font-size:12px;font-weight:600;color:#969799;margin-top:10px;">第3局 <span style="font-weight:400;color:#c8c9cc;">(选填)</span></div>
      <div style="display: flex; align-items: center; justify-content: center; gap: 12px;">
        <div style="text-align: center; flex: 1;">
          <input v-model="g3m" type="number" min="0" placeholder="-"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
        <div style="font-size: 24px; font-weight: 800; color: #969799;">:</div>
        <div style="text-align: center; flex: 1;">
          <input v-model="g3f" type="number" min="0" placeholder="-"
            style="width: 50px; height: 40px; text-align: center; font-size: 24px; font-weight: 700; border: 2px solid #ebedf0; border-radius: 10px; outline: none;" />
        </div>
      </div>

      <!-- Quick scores -->
      <div style="display: flex; gap: 6px; flex-wrap: wrap; justify-content: center; margin-top: 14px;">
        <button v-for="(s, i) in quickBO3()" :key="i"
          @click="g1m=s[0];g1f=s[1];g2m=s[2];g2f=s[3];g3m=s[4]||'';g3f=s[5]||''"
          style="padding: 5px 10px; border-radius: 14px; border: none; font-size: 12px; font-weight: 500; cursor: pointer; background: #f0f2f5; color: #666;">
          {{ s[4] ? `${s[0]}:${s[1]} ${s[2]}:${s[3]} ${s[4]}:${s[5]}` : `${s[0]}:${s[1]} ${s[2]}:${s[3]}` }}
        </button>
      </div>

      <div style="display: flex; gap: 12px; margin-top: 16px;">
        <button @click="onCancel" style="flex: 1; padding: 14px; background: #f5f5f5; border: none; border-radius: 24px; font-size: 15px; cursor: pointer;">取消</button>
        <button @click="onSubmit" style="flex: 2; padding: 14px; background: #1989fa; color: #fff; border: none; border-radius: 24px; font-size: 15px; font-weight: 600; cursor: pointer;">确认提交</button>
      </div>
    </div>
  </div>
</template>
