<script setup lang="ts">
import { useRouter } from 'vue-router'
import { IconArrowLeft } from '@tabler/icons-vue'
const router = useRouter()
</script>

<template>
  <div style="min-height: 100vh; background: #f0f2f5; padding-bottom: 80px;">
    <div class="hero">
      <div style="display: flex; align-items: center; gap: 12px;">
        <span @click="router.back()" style="cursor: pointer; font-size: 22px;">&#8592;</span>
        <div class="hero-title">📖 积分规则</div>
      </div>
      <div class="hero-sub">USATT Elo 变体 · 赛前冻结 · 胜者奖励</div>
    </div>

    <div style="padding: 16px;">

      <!-- 核心公式 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">🧮 核心公式</h3>
        <div style="background: #f8f9fa; padding: 16px; border-radius: 8px; font-size: 14px; line-height: 2;">
          <div><b>预期胜率</b> = 1 / ( 1 + 10<sup>(对手分 − 自己分) / 400</sup> )</div>
          <div><b>积分变化</b> = K × ( 实际结果 − 预期胜率 )</div>
          <div style="color: #969799; margin-top: 4px;">实际结果：赢 = 1.0，输 = 0.0</div>
        </div>
      </div>

      <!-- K 因子 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">📐 K 因子（固定值）</h3>
        <p style="font-size: 14px; color: #666; line-height: 1.8;">
          使用固定 <b style="color: #1989fa;">K = 32</b>，不再因分差而降低。
          差距越大，冷门奖励<b>自然越大</b>（Elo 公式自身特性），
          不需要人工降 K 来限制。
        </p>
        <div style="background: #f0f6ff; padding: 12px; border-radius: 8px; margin-top: 12px; font-size: 13px;">
          <b>为什么固定 K？</b> 动态 K（分差大→K 变小）会导致大冷门奖励反而少，
          违反直觉。固定 K 让数学更纯粹：你赢了一个更强的对手，就该得更多分。
        </div>
      </div>

      <!-- 预期胜率表 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">📊 预期胜率速查</h3>
        <table style="width: 100%; border-collapse: collapse; font-size: 13px;">
          <tr style="background: #f8f9fa;"><th style="padding: 8px;">分差</th><th style="padding: 8px;">高分胜率</th><th style="padding: 8px;">低分胜率</th></tr>
          <tr v-for="[d,h,l] in [[0,50,50],[50,57,43],[100,64,36],[150,70,30],[200,76,24],[250,81,19],[300,85,15],[400,91,9],[500,95,5]]" :key="d">
            <td style="padding: 6px; text-align: center; border-bottom: 1px solid #f0f0f0;">{{ d }}</td>
            <td style="padding: 6px; text-align: center; color: #07c160; font-weight: 600;">{{ h }}%</td>
            <td style="padding: 6px; text-align: center; color: #ee0a24;">{{ l }}%</td>
          </tr>
        </table>
      </div>

      <!-- 实战举例 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">⚔️ 实战举例（K=32，胜者奖励 +1）</h3>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">同分段（1500 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">胜者 <b style="color: #07c160;">+17</b>，败者 <b style="color: #ee0a24;">−16</b></div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 100 分（1600 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+13</b> / −12 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+22</b> / −21</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 300 分（1800 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+6</b> / −5 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+28</b> / −27</div>
        </div>
        <div>
          <div style="font-weight: 600; margin-bottom: 4px;">差 500 分（2000 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+3</b> / −2 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+31</b> / −30</div>
        </div>
        <div style="background: #fffbe6; padding: 12px; border-radius: 8px; margin-top: 16px; font-size: 13px; color: #ad8b00;">
          💡 低分赢高分的奖励：差100→+22，差300→+28，差500→+31。差距越大，冷门奖励越大，符合直觉。
        </div>
      </div>

      <!-- 赛前冻结 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">❄️ 赛前积分冻结</h3>
        <p style="font-size: 14px; color: #666; line-height: 1.8;">
          创建活动时，所有选手的<b>当前积分被快照</b>保存为「赛前积分」。<br>
          活动内 <b>所有比赛均使用赛前积分</b> 计算 Elo，<br>
          不会因为比赛顺序不同而产生不同结果。<br>
          活动结束时<b>一次性结算</b>到真实积分。
        </p>
        <div style="background: #fffbe6; padding: 12px; border-radius: 8px; margin-top: 12px; font-size: 13px; color: #ad8b00;">
          💡 单场快速记分（非活动模式）仍使用逐场即时更新。
        </div>
      </div>

      <!-- 胜者奖励 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">💉 胜者奖励分</h3>
        <p style="font-size: 14px; color: #666; line-height: 1.8;">
          封闭小圈子长期互打，<b>总分恒定</b>会导致水平提升无法体现在积分上。<br>
          因此引入胜者奖励：每场胜者额外 <b style="color: #07c160;">+1 分</b>（不扣败者）。
        </p>
        <div style="margin-top: 12px; padding: 12px; background: #f0f6ff; border-radius: 8px; font-size: 13px;">
          <div>• 每场向系统注入 1 分，打破零和困局</div>
          <div>• 持续赢球的人积分自然上涨</div>
          <div>• 通过环境变量 <code>RATING_BONUS=1</code> 控制（可调可关）</div>
        </div>
      </div>

      <!-- 弃权 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">🚩 弃权规则</h3>
        <p style="font-size: 14px; color: #666; line-height: 1.8;">
          弃权比赛<b>双方积分不变</b>（rating_change = 0）。<br>
          胜者计胜场，败者<b>不计负场</b>，单独统计为「弃权」。<br>
          胜率计算<b>排除弃权场次</b>。
        </p>
      </div>

    </div>
  </div>
</template>
