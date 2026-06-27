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
        <h3 style="margin-bottom: 12px;">📐 K 因子（动态调整）</h3>
        <table style="width: 100%; border-collapse: collapse; font-size: 14px;">
          <tr style="background: #f8f9fa;"><th style="padding: 10px; text-align: left;">分差范围</th><th style="padding: 10px;">K 值</th><th style="padding: 10px; text-align: left;">说明</th></tr>
          <tr><td style="padding: 10px; border-bottom: 1px solid #f0f0f0;">0 – 200</td><td style="padding: 10px; text-align: center; font-weight: 700; color: #1989fa;">32</td><td style="padding: 10px;">势均力敌，充分波动</td></tr>
          <tr><td style="padding: 10px; border-bottom: 1px solid #f0f0f0;">201 – 400</td><td style="padding: 10px; text-align: center; font-weight: 700; color: #1989fa;">24</td><td style="padding: 10px;">有差距，缩小波动</td></tr>
          <tr><td style="padding: 10px; border-bottom: 1px solid #f0f0f0;">400+</td><td style="padding: 10px; text-align: center; font-weight: 700; color: #1989fa;">16</td><td style="padding: 10px;">差距悬殊，防刷分</td></tr>
          <tr><td style="padding: 10px;">🎯 冷门</td><td style="padding: 10px; text-align: center; font-weight: 700; color: #ee0a24;">×1.2</td><td style="padding: 10px;">低分赢高分（差>100）加权</td></tr>
        </table>
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
        <h3 style="margin-bottom: 12px;">⚔️ 实战举例</h3>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">同分段（1500 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">胜者 <b style="color: #07c160;">+16</b>，败者 <b style="color: #ee0a24;">−16</b>（五五开）</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 100 分（1600 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">1600 赢 → <b style="color: #07c160;">+12</b> / −12 <span style="color: #969799;">|</span> 1500 赢（冷门）→ <b style="color: #07c160;">+25</b> / −25</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 300 分（1800 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">1800 赢 → <b style="color: #07c160;">+4</b> / −4 <span style="color: #969799;">|</span> 1500 赢 → <b style="color: #07c160;">+24</b> / −24</div>
        </div>
        <div>
          <div style="font-weight: 600; margin-bottom: 4px;">差 500 分（2000 vs 1500）</div>
          <div style="font-size: 13px; color: #666;">2000 赢 → <b style="color: #07c160;">+1</b> / −1 <span style="color: #969799;">|</span> 1500 赢 → <b style="color: #07c160;">+18</b> / −18</div>
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
