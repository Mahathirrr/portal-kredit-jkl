<template>
  <div class="page" v-if="p">
    <div class="card">
      <h2>Pengajuan {{ p.nomor }}</h2>
      <p class="muted">Status: {{ labelStatus(p.status) }}</p>
      <p v-if="p.alasan_penolakan" style="color:#c0392b;">Alasan tolak: {{ p.alasan_penolakan }}</p>
    </div>

    <div class="card">
      <h3>Isi berkas</h3>
      <p>Konsumen: {{ p.konsumen?.nama || '-' }} / NIK {{ p.konsumen?.nik || '-' }}</p>
      <p>Perkawinan: {{ p.konsumen?.status_perkawinan || '-' }} <span v-if="p.konsumen?.nama_pasangan">/ pasangan {{ p.konsumen.nama_pasangan }}</span></p>
      <p>Kendaraan: {{ p.kendaraan?.merk || '-' }} {{ p.kendaraan?.model }} {{ p.kendaraan?.tipe }}, {{ p.kendaraan?.warna }}, dealer {{ p.kendaraan?.dealer || '-' }}</p>
      <p>Harga: Rp {{ formatUang(p.kendaraan?.harga) }}</p>
      <p>Pinjaman: {{ p.pinjaman?.asuransi || '-' }}, DP Rp {{ formatUang(p.pinjaman?.uang_muka) }}, {{ p.pinjaman?.tenor_bulan || 0 }} bulan, angsuran Rp {{ formatUang(p.pinjaman?.angsuran_bulanan) }}</p>
    </div>

    <div v-if="canEditKonsumen" class="card">
      <h3>Data Konsumen</h3>
      <div class="grid-2">
        <div>
          <label>Nama</label>
          <input v-model="konsumen.nama" :readonly="auth.user.role === 'marketing'" />
        </div>
        <div>
          <label>NIK</label>
          <input v-model="konsumen.nik" maxlength="16" :readonly="auth.user.role === 'marketing'" />
        </div>
        <div v-if="auth.user.role === 'marketing'">
          <label>Tanggal lahir</label>
          <input v-model="konsumen.tanggal_lahir" type="date" />
        </div>
        <div v-if="auth.user.role === 'marketing'">
          <label>Status perkawinan</label>
          <select v-model="konsumen.status_perkawinan">
            <option value="">Pilih</option>
            <option value="belum_menikah">Belum menikah</option>
            <option value="menikah">Menikah</option>
          </select>
        </div>
        <div v-if="auth.user.role === 'marketing' && konsumen.status_perkawinan === 'menikah'">
          <label>Nama pasangan</label>
          <input v-model="konsumen.nama_pasangan" />
        </div>
        <div v-if="auth.user.role === 'marketing' && konsumen.status_perkawinan === 'menikah'">
          <label>NIK pasangan</label>
          <input v-model="konsumen.nik_pasangan" maxlength="16" />
        </div>
      </div>
      <button class="btn" style="margin-top:12px;" @click="simpanKonsumen">Simpan konsumen</button>
    </div>

    <div v-if="canEditKendaraan" class="card">
      <h3>Data Kendaraan</h3>
      <div class="grid-2">
        <div><label>Dealer</label><input v-model="kendaraan.dealer" /></div>
        <div><label>Merk</label><input v-model="kendaraan.merk" /></div>
        <div><label>Model</label><input v-model="kendaraan.model" /></div>
        <div><label>Tipe</label><input v-model="kendaraan.tipe" /></div>
        <div><label>Warna</label><input v-model="kendaraan.warna" /></div>
        <div><label>Harga</label><input v-model.number="kendaraan.harga" type="number" /></div>
      </div>
      <button class="btn" style="margin-top:12px;" @click="simpanKendaraan">Simpan kendaraan</button>
    </div>

    <div v-if="canEditPinjaman" class="card">
      <h3>Data Pinjaman</h3>
      <div class="grid-2">
        <div><label>Asuransi</label><input v-model="pinjaman.asuransi" /></div>
        <div><label>Uang muka</label><input v-model.number="pinjaman.uang_muka" type="number" /></div>
        <div><label>Tenor bulan</label><input v-model.number="pinjaman.tenor_bulan" type="number" /></div>
        <div><label>Angsuran bulanan</label><input v-model.number="pinjaman.angsuran_bulanan" type="number" /></div>
      </div>
      <button class="btn" style="margin-top:12px;" @click="simpanPinjaman">Simpan pinjaman</button>
    </div>

    <div class="card">
      <h3>Dokumen</h3>
      <ul>
        <li v-for="d in p.dokumen || []" :key="d.id">{{ d.jenis }}: {{ d.nama_file }}</li>
      </ul>
      <div v-if="uploadOptions.length" class="grid-2" style="margin-top:12px;">
        <div>
          <label>Jenis dokumen</label>
          <select v-model="uploadJenis">
            <option v-for="j in uploadOptions" :key="j" :value="j">{{ j }}</option>
          </select>
        </div>
        <div>
          <label>File</label>
          <input type="file" accept=".pdf,.jpg,.jpeg,.png" @change="onFile" />
        </div>
      </div>
      <button v-if="uploadOptions.length && selectedFile" class="btn" @click="unggah">Unggah dokumen</button>
    </div>

    <div class="card no-print">
      <h3>Aksi</h3>
      <p v-if="info" :style="error ? 'color:#c0392b' : ''" class="muted">{{ info }}</p>
      <button v-if="auth.user.role === 'dealer' && p.status === 'draft_dealer'" class="btn" @click="kirimDealer">Kirim ke marketing</button>
      <button v-if="auth.user.role === 'marketing' && p.status === 'menunggu_marketing'" class="btn" @click="kirimMarketing">Kirim ke approval</button>
      <button v-if="auth.user.role === 'supervisor' && p.status === 'menunggu_approval'" class="btn" @click="setujui">Setujui</button>
      <button v-if="auth.user.role === 'supervisor' && p.status === 'menunggu_approval'" class="btn warn" @click="tolak">Tolak</button>
      <button v-if="showCetak" class="btn ghost" @click="bukaKontrak">Cetak kontrak</button>
      <button v-if="showCetak" class="btn ghost" @click="bukaPO">Cetak PO</button>
      <button v-if="auth.user.role === 'marketing' && p.status === 'menunggu_ttd'" class="btn" @click="ttdLengkap">Tandai TTD lengkap</button>
      <div v-if="auth.user.role === 'admin' && p.status === 'ttd_lengkap'" class="grid-2" style="margin-top:12px;">
        <div><label>Nominal cair</label><input v-model.number="cair.nominal" type="number" /></div>
        <div><label>Tanggal cair</label><input v-model="cair.tanggal_cair" type="date" /></div>
        <div><label>Catatan</label><input v-model="cair.catatan" /></div>
      </div>
      <button v-if="auth.user.role === 'admin' && p.status === 'ttd_lengkap'" class="btn" @click="prosesCair">Catat pencairan</button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const p = ref(null)
const info = ref('')
const error = ref(false)
const uploadJenis = ref('')
const selectedFile = ref(null)
const cair = reactive({ nominal: 0, tanggal_cair: '', catatan: '' })

const konsumen = reactive({ nama: '', nik: '', tanggal_lahir: '', status_perkawinan: '', nama_pasangan: '', nik_pasangan: '' })
const kendaraan = reactive({ dealer: '', merk: '', model: '', tipe: '', warna: '', harga: 0 })
const pinjaman = reactive({ asuransi: '', uang_muka: 0, tenor_bulan: 0, angsuran_bulanan: 0 })

const canEditKonsumen = computed(() => {
  if (auth.user.role === 'dealer') return p.value?.status === 'draft_dealer'
  if (auth.user.role === 'marketing') return p.value?.status === 'menunggu_marketing'
  return false
})
const canEditKendaraan = computed(() => auth.user.role === 'dealer' && p.value?.status === 'draft_dealer')
const canEditPinjaman = computed(() => auth.user.role === 'marketing' && p.value?.status === 'menunggu_marketing')
const showCetak = computed(() => ['menunggu_ttd', 'ttd_lengkap', 'dana_cair'].includes(p.value?.status))

const uploadOptions = computed(() => {
  if (auth.user.role === 'dealer' && p.value?.status === 'draft_dealer') return ['ktp', 'spk', 'bukti_tanda_jadi']
  if (auth.user.role === 'marketing' && p.value?.status === 'menunggu_marketing') return ['kartu_keluarga']
  if (auth.user.role === 'marketing' && p.value?.status === 'menunggu_ttd') return ['kontrak_ttd', 'po_ttd']
  return []
})

watch(uploadOptions, (opts) => {
  if (opts.length) uploadJenis.value = opts[0]
})

function formatUang(n) {
  return new Intl.NumberFormat('id-ID').format(n || 0)
}

async function jalankan(fn, sukses) {
  error.value = false
  info.value = ''
  try {
    await fn()
    info.value = sukses
    await load()
  } catch (e) {
    error.value = true
    info.value = e.message
  }
}

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

async function load() {
  p.value = await api.getPengajuan(route.params.id)
  Object.assign(konsumen, p.value.konsumen || {})
  if (konsumen.tanggal_lahir && String(konsumen.tanggal_lahir).includes('T')) {
    konsumen.tanggal_lahir = String(konsumen.tanggal_lahir).slice(0, 10)
  }
  Object.assign(kendaraan, p.value.kendaraan || {})
  Object.assign(pinjaman, p.value.pinjaman || {})
}

function onFile(e) {
  selectedFile.value = e.target.files?.[0] || null
}

function simpanKonsumen() {
  const body = { ...konsumen }
  if (body.tanggal_lahir) body.tanggal_lahir = new Date(body.tanggal_lahir).toISOString()
  else body.tanggal_lahir = null
  return jalankan(() => api.updateKonsumen(p.value.id, body), 'Data konsumen tersimpan')
}

function simpanKendaraan() {
  return jalankan(() => api.updateKendaraan(p.value.id, kendaraan), 'Data kendaraan tersimpan')
}

function simpanPinjaman() {
  return jalankan(() => api.updatePinjaman(p.value.id, pinjaman), 'Data pinjaman tersimpan')
}

function unggah() {
  return jalankan(async () => {
    await api.uploadDokumen(p.value.id, uploadJenis.value, selectedFile.value)
    selectedFile.value = null
  }, 'Dokumen diunggah')
}

function kirimDealer() {
  return jalankan(() => api.submitDealer(p.value.id), 'Berkas dikirim ke marketing')
}

function kirimMarketing() {
  return jalankan(() => api.submitMarketing(p.value.id), 'Berkas dikirim ke atasan')
}

function setujui() {
  return jalankan(() => api.approve(p.value.id, ''), 'Pengajuan disetujui')
}

function tolak() {
  const catatan = window.prompt('Alasan penolakan')
  if (!catatan) return
  return jalankan(() => api.reject(p.value.id, catatan), 'Pengajuan ditolak')
}

function ttdLengkap() {
  return jalankan(() => api.markSigned(p.value.id), 'TTD sudah lengkap')
}

function prosesCair() {
  return jalankan(() => api.disburse(p.value.id, cair), 'Pencairan dicatat')
}

function bukaKontrak() {
  router.push({ name: 'kontrak', params: { id: p.value.id } })
}

function bukaPO() {
  router.push({ name: 'po', params: { id: p.value.id } })
}

onMounted(load)
</script>