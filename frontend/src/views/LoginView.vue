<template>
  <div class="page" style="max-width: 420px;">
    <div class="card">
      <h2>Login PT JKL</h2>
      <p class="muted">Gunakan akun demo: dealer1, marketing1, supervisor1, admin1. Password: password123</p>
      <form @submit.prevent="submit">
        <label>Username</label>
        <input v-model="username" required />
        <label style="margin-top: 10px;">Password</label>
        <input v-model="password" type="password" required />
        <p v-if="error" style="color:#c0392b;">{{ error }}</p>
        <button class="btn" style="margin-top: 14px; width: 100%;">Masuk</button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const username = ref('dealer1')
const password = ref('password123')
const error = ref('')
const router = useRouter()
const auth = useAuthStore()

async function submit() {
  error.value = ''
  try {
    await auth.login(username.value, password.value)
    router.push({ name: 'home' })
  } catch (e) {
    error.value = e.message
  }
}
</script>