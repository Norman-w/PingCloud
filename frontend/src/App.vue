<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { IconHome, IconTournament, IconUserPlus } from '@tabler/icons-vue'

const router = useRouter()
const route = useRoute()

const tabs = [
  { name: 'Home', label: '首页', component: IconHome },
  { name: 'SessionView', label: '活动', component: IconTournament },
  { name: 'AddPlayer', label: '球员', component: IconUserPlus },
]

const active = computed(() => {
  const name = route.name as string
  return tabs.find(t => t.name === name) ? name : ''
})

function onTabChange(name: string) {
  router.push({ name })
}
</script>

<template>
  <div id="app-shell">
    <div class="app-body">
      <router-view />
    </div>
    <nav class="tabbar">
      <button
        v-for="tab in tabs"
        :key="tab.name"
        class="tabbar-item"
        :class="{ active: active === tab.name }"
        @click="onTabChange(tab.name)"
      >
        <component :is="tab.component" :size="24" :stroke-width="active === tab.name ? 2.5 : 1.5" />
        <span class="tab-label">{{ tab.label }}</span>
      </button>
    </nav>
  </div>
</template>

<style scoped>
#app-shell {
  min-height: 100vh;
  background: #f0f2f5;
}

.app-body {
  padding-bottom: 64px;
}

.tabbar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  padding: 6px 0 max(8px, env(safe-area-inset-bottom));
  z-index: 999;
}

.tabbar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 6px 4px;
  cursor: pointer;
  background: none;
  border: none;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  color: #969799;
  transition: color 0.2s;
}

.tabbar-item.active {
  color: #1989fa;
}

.tab-label {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.2px;
}

.tabbar-item.active .tab-label {
  font-weight: 700;
}
</style>
