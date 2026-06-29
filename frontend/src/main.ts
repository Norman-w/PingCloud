import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'

import 'vant/lib/index.css'
import './style.css'

const routes = [
  { path: '/', name: 'Home', component: () => import('./views/Home.vue') },
  { path: '/session', name: 'SessionView', component: () => import('./views/SessionView.vue') },
  { path: '/add-player', name: 'AddPlayer', component: () => import('./views/AddPlayer.vue') },
  { path: '/record', name: 'RecordMatch', component: () => import('./views/RecordMatch.vue') },
  { path: '/player/:id', name: 'PlayerDetail', component: () => import('./views/PlayerDetail.vue') },
  { path: '/rules', name: 'Rules', component: () => import('./views/Rules.vue') },
  { path: '/bulletin', name: 'Bulletin', component: () => import('./views/Bulletin.vue') },
  { path: '/scoreboard', name: 'Scoreboard', component: () => import('./views/Scoreboard.vue') },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.mount('#app')
