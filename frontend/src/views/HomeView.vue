<template>
  <div class="page">
    <div class="card">
      <div style="display:flex; justify-content:space-between; align-items:center;">
        <h2>Daftar Pengajuan</h2>
        <button v-if="auth.user?.role === 'dealer'" class="btn" @click="buatDraft">Buat Draft</button>
      </div>
      <p v-if="message" class="muted">{{ message }}</p>
      <table v-if="items.length">
        <thead>
          <tr>
            <th>Nomor</th>
            <th>Status</th>
            <th>Diperbarui</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in items" :key="p.id">
            <td>{{ p.nomor }}</td>
            <td>{{ labelStatus(p.status) }}</td>
            <td>{{ formatDate(p.updated_at) }}</td>
            <td><router-link :to="`/pengajuan/${p.id}`">Buka</router-link></td>
          </tr>
        </tbody>
      </table>
      <p v-else class="muted">Belum ada pengajuan.</p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'

const items = ref([])
const message = ref('')
const auth = useAuthStore()

function labelStatus(status) {
  const map = {
    draft_dealer: 'Draft dealer',
    menunggu_marketing: 'Menunggu marketing',
    menunggu_approval: 'Menunggu approval',
    ditolak: 'Ditolak',
    menunggu_ttd: 'Menunggu TTD',
    ttd_lengkap: 'TTD lengkap',
    dana_cair: 'Dana cair'
  }
  return map[status] || status
}

function formatDate(val) {
  return new Date(val).toLocaleString('id-ID')
}

async function load() {
  items.value = await api.listPengajuan()
}

async function buatDraft() {
  const p = await api.createPengajuan()
  message.value = `Draft ${p.nomor} dibuat`
  await load()
}

onMounted(load)
</script>