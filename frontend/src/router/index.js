import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '../views/HomeView.vue'
import PengajuanView from '../views/PengajuanView.vue'
import CetakKontrakView from '../views/CetakKontrakView.vue'
import CetakPOView from '../views/CetakPOView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView },
    { path: '/', name: 'home', component: HomeView },
    { path: '/pengajuan/:id', name: 'pengajuan', component: PengajuanView },
    { path: '/pengajuan/:id/kontrak', name: 'kontrak', component: CetakKontrakView },
    { path: '/pengajuan/:id/po', name: 'po', component: CetakPOView }
  ]
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'login' && !token) {
    return { name: 'login' }
  }
  if (to.name === 'login' && token) {
    return { name: 'home' }
  }
})

export default router