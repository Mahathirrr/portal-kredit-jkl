# Portal Kredit Kendaraan

Aplikasi web Go dan Vue 3 untuk jawaban tes coding PT. JKL. Satu berkas pengajuan dipakai dari dealer sampai pencairan, tanpa fotokopi dan ketik ulang.

Dibuat oleh **Muhammad Mahathir**.

## Struktur proyek

| Lokasi | Tanggung jawab |
| --- | --- |
| `backend/cmd/api` | Menjalankan server HTTP dan migrasi awal. |
| `backend/internal/handler` | Menerima HTTP dan mengembalikan JSON atau unduhan. |
| `backend/internal/middleware` | Memeriksa JWT dan identitas pengguna. |
| `backend/internal/service` | Aturan peran, kelengkapan data, dan perpindahan status. |
| `backend/internal/repository` | Query MariaDB. Interface repository dipakai oleh service. |
| `backend/internal/domain` | Bentuk data pengajuan, konsumen, kendaraan, pinjaman, dan dokumen. |
| `database/migrations` | Skema tabel pengajuan, dokumen, approval, pencairan, dan riwayat. |
| `frontend/src/views` | Halaman login, daftar, detail berkas, cetak kontrak, dan cetak PO. |
| `frontend/src/stores` | Sesi login di browser. |
| `scripts/test-flow.sh` | Uji alur lengkap lewat API. |

## Menjalankan aplikasi

Prasyarat: Go 1.22+, Node.js 22+, dan MariaDB atau MySQL yang sudah berjalan di mesin lokal. Jangan memakai Docker. Jalankan perintah dari root repo.

Siapkan database dan user:

```bash
sudo mariadb -e "
CREATE DATABASE IF NOT EXISTS kredit_jkl CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'kredit_jkl'@'localhost' IDENTIFIED BY 'kredit_jkl_dev';
GRANT ALL PRIVILEGES ON kredit_jkl.* TO 'kredit_jkl'@'localhost';
FLUSH PRIVILEGES;
"
```

Salin `.env.example` bila ingin mengubah port, koneksi, atau rahasia JWT. Nilai bawaan sudah mengarah ke database di atas.

```bash
cp .env.example .env
```

Isi `JWT_SECRET` dengan keluaran perintah `openssl rand -base64 48`. Gunakan kunci yang berbeda pada setiap lingkungan. Server tetap bisa jalan dengan nilai bawaan `ganti-di-production` untuk percobaan lokal.

Jalankan API:

```bash
cd backend
go run ./cmd/api
```

API berada di `http://127.0.0.1:8080`. Saat pertama kali jalan, tabel dibuat dan empat akun demo diisi jika tabel user masih kosong.

Pada terminal lain, jalankan tampilan:

```bash
cd frontend
npm install
npm run dev
```

Buka `http://127.0.0.1:5173/login`. Permintaan `/api` dari Vue diteruskan ke backend.

## Akun demo

Kata sandi semua akun adalah `password123`.

| Peran | Username | Akses |
| --- | --- | --- |
| Sales Dealer | `dealer1` | Membuat draft, mengisi data deal, mengunggah KTP, SPK, dan bukti tanda jadi, lalu mengirim ke marketing. Hanya melihat pengajuan miliknya. |
| Marketing | `marketing1` | Melengkapi tanggal lahir, status perkawinan, data pasangan, KK, dan pinjaman. Mengirim ke atasan. Setelah disetujui, mengunggah pindaian kontrak dan PO yang sudah ditandatangani. |
| Atasan Marketing | `supervisor1` | Melihat isi berkas, menyetujui, atau menolak dengan alasan. |
| Admin Backoffice | `admin1` | Mencetak kontrak dan PO dari data yang sudah disetujui, lalu mencatat pencairan setelah tanda tangan lengkap. |

Login memakai JWT HS256 berumur 8 jam. Token dikirim pada header `Authorization` dan disimpan di browser agar sesi tidak hilang saat halaman dimuat ulang. Password disimpan sebagai hash bcrypt.

Unggahan disimpan di folder `uploads` dan hanya bisa diunduh lewat endpoint yang memeriksa login serta kepemilikan berkas. Jenis yang diterima adalah pdf, jpg, dan png, dengan ukuran paling besar 5 MB. Dokumen sejenis yang diunggah ulang menggantikan file sebelumnya.

## Alur

1. Dealer membuat draft, mengisi nama, NIK, dealer, kendaraan, dan harga, lalu mengunggah KTP, SPK, dan bukti tanda jadi. Status menjadi `menunggu_marketing`.
2. Marketing tidak mengetik ulang nama, NIK, dan kendaraan. Marketing mengisi tanggal lahir, status perkawinan, data pasangan bila menikah, kartu keluarga, asuransi, uang muka, tenor, dan angsuran. Status menjadi `menunggu_approval`.
3. Atasan menyetujui atau menolak dengan alasan. Pengajuan yang ditolak berhenti di status `ditolak`.
4. Jika disetujui, status menjadi `menunggu_ttd`. Kontrak dan PO dicetak dari data berkas yang sama.
5. Marketing mengunggah pindaian kontrak dan PO yang sudah ditandatangani. Status menjadi `ttd_lengkap`.
6. Admin mencatat nominal, tanggal, dan catatan pencairan. Status menjadi `dana_cair`.

Setiap perubahan status mencatat pelaku dan waktu. Service menolak aksi yang tidak sesuai peran atau status, jadi tahap tidak bisa dilompati lewat URL langsung.

## Pemeriksaan

```bash
cd backend && go test ./...
cd frontend && npm run build
```

Saat API sudah berjalan, uji satu berkas dari dealer sampai cair:

```bash
./scripts/test-flow.sh
```

Skrip itu harus berakhir dengan status `dana_cair`.

## Batas implementasi

Halaman cetak adalah ringkasan data untuk demonstrasi, bukan kontrak hukum resmi. Tanda tangan basah tetap dilakukan di luar sistem. Aplikasi hanya menyimpan hasil pindaiannya. Verifikasi identitas, penilaian kredit, dan transfer dana bank tidak diminta soal ini. Pencairan di aplikasi adalah pencatatan setelah dokumen bertanda tangan diterima.
