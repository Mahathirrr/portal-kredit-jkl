const apiBase = '/api'

async function request(path, options = {}) {
  const headers = { ...(options.headers || {}) }
  const token = localStorage.getItem('token')
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }
  if (options.body && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(`${apiBase}${path}`, { ...options, headers })
  const text = await res.text()
  let data = {}
  if (text) {
    data = JSON.parse(text)
  }
  if (!res.ok) {
    throw new Error(data.error || 'Permintaan gagal')
  }
  return data
}

export const api = {
  login(username, password) {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password })
    })
  },
  me() {
    return request('/auth/me')
  },
  listPengajuan() {
    return request('/pengajuan')
  },
  createPengajuan() {
    return request('/pengajuan', { method: 'POST' })
  },
  getPengajuan(id) {
    return request(`/pengajuan/${id}`)
  },
  updateKonsumen(id, body) {
    return request(`/pengajuan/${id}/konsumen`, { method: 'PUT', body: JSON.stringify(body) })
  },
  updateKendaraan(id, body) {
    return request(`/pengajuan/${id}/kendaraan`, { method: 'PUT', body: JSON.stringify(body) })
  },
  updatePinjaman(id, body) {
    return request(`/pengajuan/${id}/pinjaman`, { method: 'PUT', body: JSON.stringify(body) })
  },
  uploadDokumen(id, jenis, file) {
    const form = new FormData()
    form.append('jenis', jenis)
    form.append('file', file)
    return request(`/pengajuan/${id}/dokumen`, { method: 'POST', body: form })
  },
  submitDealer(id) {
    return request(`/pengajuan/${id}/submit-dealer`, { method: 'POST' })
  },
  submitMarketing(id) {
    return request(`/pengajuan/${id}/submit-marketing`, { method: 'POST' })
  },
  approve(id, catatan) {
    return request(`/pengajuan/${id}/approve`, { method: 'POST', body: JSON.stringify({ catatan }) })
  },
  reject(id, catatan) {
    return request(`/pengajuan/${id}/reject`, { method: 'POST', body: JSON.stringify({ catatan }) })
  },
  markSigned(id) {
    return request(`/pengajuan/${id}/mark-signed`, { method: 'POST' })
  },
  disburse(id, body) {
    return request(`/pengajuan/${id}/disburse`, { method: 'POST', body: JSON.stringify(body) })
  }
}