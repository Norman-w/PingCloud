<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div style="min-height: 100vh; background: #f0f2f5; padding-bottom: 80px;">
    <div class="hero">
      <div style="display: flex; align-items: center; gap: 12px;">
        <span @click="router.back()" style="cursor: pointer; font-size: 22px;">&#8592;</span>
        <div class="hero-title">积分规则</div>
      </div>
      <div class="hero-sub">友谊第一，比赛第二</div>
    </div>

    <div style="padding: 16px;">

      <!-- 声明 -->
      <div class="card" style="margin: 0 0 12px; background: #fffbe6;">
        <h3 style="margin-bottom: 8px;">📢 重要声明</h3>
        <p style="font-size: 13px; color: #ad8b00; line-height: 1.8;">
          本系统积分<b>仅供内部娱乐参考</b>，不对标开球网，不反映个人真实竞技水平。
          积分规则也<b>并非与开球网完全统一</b>。
          打得越多，积分越稳定，越能反映出咱们圈子内大家的相对水平。
          <b>友谊第一，比赛第二！</b>
        </p>
      </div>

      <!-- 核心理念 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">💡 核心理念</h3>
        <div style="font-size: 14px; color: #666; line-height: 2;">
          <div>• 积分只是<b>圈子内的相对参考</b>，不是绝对水平</div>
          <div>• <b>打得越多越准</b>：场次少时波动大，打多了自然稳定</div>
          <div>• 赢了高分对手 <b>涨得快</b>，输了低分对手 <b>掉得多</b></div>
          <div>• 同水平互打，积分在<b>小范围内波动</b></div>
        </div>
      </div>

      <!-- 积分变化速查 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">📊 积分变化速查</h3>
        <p style="font-size: 13px; color: #969799; margin-bottom: 12px;">对手分数和你差距越大，冷门奖励越高</p>
        <table style="width: 100%; border-collapse: collapse; font-size: 13px;">
          <tr style="background: #f8f9fa;">
            <th style="padding: 8px;">你们的分差</th>
            <th style="padding: 8px;">你赢（正常）</th>
            <th style="padding: 8px; color: #ee0a24;">他赢（冷门）</th>
          </tr>
          <tr v-for="r in [
            ['0–12 分', '±8', '±8'],
            ['13–37 分', '±7', '±12'],
            ['38–62 分', '±6', '±16'],
            ['63–87 分', '±5', '±20'],
            ['88–112 分', '±4', '±25'],
            ['113–137 分', '±3', '±30'],
            ['138–162 分', '±2', '±35'],
            ['163–187 分', '±2', '±40'],
            ['188–212 分', '±1', '±45'],
            ['213 分以上', '±0–1', '±50'],
          ]" :key="r[0]">
            <td style="padding: 6px; text-align: center; border-bottom: 1px solid #f0f0f0;">{{ r[0] }}</td>
            <td style="padding: 6px; text-align: center; color: #1989fa; font-weight: 600;">{{ r[1] }}</td>
            <td style="padding: 6px; text-align: center; color: #ee0a24; font-weight: 600;">{{ r[2] }}</td>
          </tr>
        </table>
        <div style="background: #f0f6ff; padding: 10px; border-radius: 8px; margin-top: 12px; font-size: 12px; color: #666;">
          💡 新手前 30 场翻倍，老手按此表。高手（2400 分以上）减半。
        </div>
      </div>

      <!-- 实战举例 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">⚔️ 举个栗子</h3>
        <div style="font-size: 14px; color: #666; line-height: 2;">
          <div>你是 <b>1550</b>，对手是 <b>1680</b>（差 130 分）</div>
          <div style="margin-top: 8px;">你赢了 → <b style="color: #07c160; font-size: 18px;">+30</b>（冷门奖励！）</div>
          <div>你输了 → <b style="color: #ee0a24;">−3</b>（正常，毕竟分高）</div>
          <div style="margin-top: 8px; color: #969799; font-size: 13px;">
            反过来，他赢你只 +3，他输你 −30
          </div>
        </div>
      </div>

      <!-- 常见问题 -->
      <div class="card" style="margin: 0 0 12px;">
        <h3 style="margin-bottom: 12px;">🙋 常见问题</h3>

        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">为什么我积分一直不动？</div>
          <div style="font-size: 13px; color: #666;">你赢的都是应该赢的（积分变化小），输的也都是正常的。多打、多赢冷门才会涨。</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">弃权影响积分吗？</div>
          <div style="font-size: 13px; color: #666;">不影响。弃权双方积分不变，但胜者会计胜场。</div>
        </div>
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; margin-bottom: 4px;">为什么活动里积分没变？</div>
          <div style="font-size: 13px; color: #666;">活动内使用开赛前的积分计算，活动结束后一次性结算。</div>
        </div>
        <div>
          <div style="font-weight: 600; margin-bottom: 4px;">积分能反映真实水平吗？</div>
          <div style="font-size: 13px; color: #666;">不能。本系统仅供球友圈内部娱乐，不对标任何专业积分体系。打得越多，排名越能反映你在本圈的相对位置。</div>
        </div>
      </div>

    </div>
  </div>
</template>
