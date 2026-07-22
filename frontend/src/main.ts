import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'

import 'vant/lib/index.css'
import './style.css'

const routes = [
  { path: '/', name: 'Home', component: () => import('./views/Home.vue') },
  { path: '/session', name: 'SessionView', component: () => import('./views/SessionView.vue') },
  { path: '/add-player', name: 'AddPlayer', component: () => import('./views/AddPlayer.vue') },
  { path: '/me', name: 'Me', component: () => import('./views/Me.vue') },
  { path: '/skill-train/:id', name: 'SkillTrain', component: () => import('./views/SkillTrain.vue') },
  { path: '/record', name: 'RecordMatch', component: () => import('./views/RecordMatch.vue') },
  { path: '/player/:id', name: 'PlayerDetail', component: () => import('./views/PlayerDetail.vue') },
  { path: '/rules', name: 'Rules', component: () => import('./views/Rules.vue') },
  { path: '/bulletin', name: 'Bulletin', component: () => import('./views/Bulletin.vue') },
  { path: '/headtohead', name: 'HeadToHead', component: () => import('./views/HeadToHead.vue') },
  { path: '/scoreboard', name: 'Scoreboard', component: () => import('./views/Scoreboard.vue') },
  { path: '/fun-match', name: 'FunMatch', component: () => import('./views/FunMatch.vue') },
  { path: '/team-battle', name: 'TeamBattle', component: () => import('./views/TeamBattle.vue') },
  { path: '/tournament', name: 'Tournament', component: () => import('./views/Tournament.vue') },
  { path: '/admin/login', name: 'AdminLogin', component: () => import('./views/AdminLogin.vue') },
  { path: '/admin', name: 'Admin', component: () => import('./views/Admin.vue') },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.mount('#app')
