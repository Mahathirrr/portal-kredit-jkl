<template>
  <div class="print-sheet" v-if="p">
    <div class="no-print" style="margin-bottom:16px;">
      <button class="btn" @click="printPage">Cetak</button>
      <router-link :to="`/pengajuan/${p.id}`" style="margin-left:8px;">Kembali</router-link>
    </div>
    <h2>Dokumen Kontrak Kredit</h2>
    <p>Nomor pengajuan: {{ p.nomor }}</p>
    <p>Nama konsumen: {{ p.konsumen?.nama }}</p>
    <p>NIK: {{ p.konsumen?.nik }}</p>
    <p>Kendaraan: {{ p.kendaraan?.merk }} {{ p.kendaraan?.model }} {{ p.kendaraan?.tipe }}</p>
    <p>Harga: Rp {{ formatNumber(p.kendaraan?.harga) }}</p>
    <p>Uang muka: Rp {{ formatNumber(p.pinjaman?.uang_muka) }}</p>
    <p>Tenor: {{ p.pinjaman?.tenor_bulan }} bulan</p>
    <p>Angsuran: Rp {{ formatNumber(p.pinjaman?.angsuran_bulanan) }}</p>
    <p style="margin-top:40px;">Kontrak dicetak, ditandatangani, lalu hasil pindaian diunggah oleh marketing.</p>
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