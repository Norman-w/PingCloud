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

      <!-- 核心算法 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">🧮 积分算法：USATT 查表法</h3>
        <p style="font-size: 14px; color: #666; line-height: 1.8; margin-bottom: 12px;">
          与开球网一致，使用 <b>USATT 原始查表法</b> 而非纯 Elo 公式。
          根据双方分差直接查表得出积分变化。
        </p>
        <table style="width: 100%; border-collapse: collapse; font-size: 13px;">
          <tr style="background: #f8f9fa;"><th style="padding: 8px;">分差范围</th><th style="padding: 8px;">高分赢（±）</th><th style="padding: 8px; color: #ee0a24;">低分赢（±）</th></tr>
          <tr v-for="r in [[0,12,8,8],[13,37,7,12],[38,62,6,16],[63,87,5,20],[88,112,4,25],[113,137,3,30],[138,162,2,35],[163,187,2,40],[188,212,1,45],[213,237,1,50],[238,9999,0,50]]" :key="r[0]">
            <td style="padding: 6px; text-align: center; border-bottom: 1px solid #f0f0f0;">{{ r[0] }}{{ r[1] >= 9999 ? '+' : '–'+r[1] }}</td>
            <td style="padding: 6px; text-align: center; color: #1989fa; font-weight: 600;">{{ r[2] }}</td>
            <td style="padding: 6px; text-align: center; color: #ee0a24; font-weight: 600;">{{ r[3] }}</td>
          </tr>
        </table>
        <div style="background: #f0f6ff; padding: 12px; border-radius: 8px; margin-top: 12px; font-size: 13px;">
          💡 以上为 <b>K=20</b> 基准值。新人 K=40 翻倍，高手 K=10 减半。
        </div>
      </div>

      <!-- K 因子 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">📐 K 因子（按选手经验）</h3>
        <table style="width: 100%; border-collapse: collapse; font-size: 14px;">
          <tr style="background: #f8f9fa;"><th style="padding: 10px; text-align: left;">K 值</th><th style="padding: 10px; text-align: left;">适用条件</th><th style="padding: 10px; text-align: left;">相当于</th></tr>
          <tr><td style="padding: 10px; font-weight: 700; color: #1989fa; text-align: center;">40</td><td style="padding: 10px;">新选手（比赛 &lt; 30 场）</td><td style="padding: 10px; font-size: 13px;">表值 ×2</td></tr>
          <tr><td style="padding: 10px; font-weight: 700; color: #1989fa; text-align: center;">20</td><td style="padding: 10px;">常规选手（比赛 ≥ 30 场）</td><td style="padding: 10px; font-size: 13px;">表值 ×1（基准）</td></tr>
          <tr><td style="padding: 10px; font-weight: 700; color: #1989fa; text-align: center;">10</td><td style="padding: 10px;">高手（积分 ≥ 2400）</td><td style="padding: 10px; font-size: 13px;">表值 ×0.5</td></tr>
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
        <h3 style="margin-bottom: 12px;">⚔️ 实战举例（K=20 基准）</h3>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">同分段（差 0-12）</div>
          <div style="font-size: 13px; color: #666;">胜者 <b style="color: #07c160;">+8</b>，败者 <b style="color: #ee0a24;">−8</b></div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 50 分（1600 vs 1550）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+6</b> / −6 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+16</b> / −16</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 100 分（1650 vs 1550）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+4</b> / −4 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+25</b> / −25</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">差 130 分（1680 vs 1550）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+3</b> / −3 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+30</b> / −30 🔥</div>
        </div>
        <div>
          <div style="font-weight: 600; margin-bottom: 4px;">差 200 分（1750 vs 1550）</div>
          <div style="font-size: 13px; color: #666;">高分赢 → <b style="color: #07c160;">+1</b> / −1 <span style="color: #969799;">|</span> 低分赢 → <b style="color: #07c160;">+45</b> / −45</div>
        </div>
        <div style="background: #fffbe6; padding: 12px; border-radius: 8px; margin-top: 16px; font-size: 13px; color: #ad8b00;">
          💡 差 130 低分赢 = <b>+30</b>（K=20）。新人 K=40 → +60。与你开球网实战一致。
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
