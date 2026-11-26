<template>
  <div v-if="route.name !== 'login'" class="topbar">
    <div>
      <strong>PT JKL Kredit Motor</strong>
      <span v-if="auth.user">({{ auth.user.role }} - {{ auth.user.nama }})</span>
    </div>
    <button v-if="auth.user" class="btn ghost" @click="logout">Keluar</button>
  </div>
  <router-view />
</template>

<script setup>
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>