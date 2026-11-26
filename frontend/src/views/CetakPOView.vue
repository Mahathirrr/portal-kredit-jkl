<template>
  <div class="print-sheet" v-if="p">
    <div class="no-print" style="margin-bottom:16px;">
      <button class="btn" @click="printPage">Cetak</button>
      <router-link :to="`/pengajuan/${p.id}`" style="margin-left:8px;">Kembali</router-link>
    </div>
    <h2>Dokumen Purchase Order</h2>
    <p>Nomor pengajuan: {{ p.nomor }}</p>
    <p>Dealer: {{ p.kendaraan?.dealer }}</p>
    <p>Konsumen: {{ p.konsumen?.nama }}</p>
    <p>Unit: {{ p.kendaraan?.merk }} {{ p.kendaraan?.model }} {{ p.kendaraan?.warna }}</p>
    <p>Total harga: Rp {{ formatNumber(p.kendaraan?.harga) }}</p>
    <p style="margin-top:40px;">PO dicetak setelah pengajuan disetujui.</p>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api'

const route = useRoute()
const p = ref(null)

function formatNumber(n) {
  return new Intl.NumberFormat('id-ID').format(n || 0)
}

function printPage() {
  window.print()
}

onMounted(async () => {
  p.value = await api.getPengajuan(route.params.id)
})
</script>