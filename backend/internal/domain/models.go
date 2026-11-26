package domain

import "time"

const (
	RoleDealer     = "dealer"
	RoleMarketing  = "marketing"
	RoleSupervisor = "supervisor"
	RoleAdmin      = "admin"

	StatusDraftDealer       = "draft_dealer"
	StatusMenungguMarketing = "menunggu_marketing"
	StatusMenungguApproval  = "menunggu_approval"
	StatusDisetujui         = "disetujui"
	StatusDitolak           = "ditolak"
	StatusMenungguTTD       = "menunggu_ttd"
	StatusTTDLengkap        = "ttd_lengkap"
	StatusDanaCair          = "dana_cair"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Nama         string    `json:"nama"`
	CreatedAt    time.Time `json:"created_at"`
}

type Pengajuan struct {
	ID               int64     `json:"id"`
	Nomor            string    `json:"nomor"`
	DealerUserID     int64     `json:"dealer_user_id"`
	MarketingUserID  *int64    `json:"marketing_user_id,omitempty"`
	Status           string    `json:"status"`
	AlasanPenolakan  string    `json:"alasan_penolakan,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Konsumen         *Konsumen `json:"konsumen,omitempty"`
	Kendaraan        *Kendaraan `json:"kendaraan,omitempty"`
	Pinjaman         *Pinjaman `json:"pinjaman,omitempty"`
	Dokumen          []Dokumen `json:"dokumen,omitempty"`
}

type Konsumen struct {
	PengajuanID       int64      `json:"pengajuan_id"`
	Nama              string     `json:"nama"`
	NIK               string     `json:"nik"`
	TanggalLahir      *time.Time `json:"tanggal_lahir,omitempty"`
	StatusPerkawinan  string     `json:"status_perkawinan"`
	NamaPasangan      string     `json:"nama_pasangan"`
	NIKPasangan       string     `json:"nik_pasangan"`
}

type Kendaraan struct {
	PengajuanID int64   `json:"pengajuan_id"`
	Dealer      string  `json:"dealer"`
	Merk        string  `json:"merk"`
	Model       string  `json:"model"`
	Tipe        string  `json:"tipe"`
	Warna       string  `json:"warna"`
	Harga       float64 `json:"harga"`
}

type Pinjaman struct {
	PengajuanID       int64   `json:"pengajuan_id"`
	Asuransi          string  `json:"asuransi"`
	UangMuka          float64 `json:"uang_muka"`
	TenorBulan        int     `json:"tenor_bulan"`
	AngsuranBulanan   float64 `json:"angsuran_bulanan"`
}

type Dokumen struct {
	ID          int64     `json:"id"`
	PengajuanID int64     `json:"pengajuan_id"`
	Jenis       string    `json:"jenis"`
	NamaFile    string    `json:"nama_file"`
	PathFile    string    `json:"-"`
	UploadedBy  int64     `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Approval struct {
	PengajuanID int64     `json:"pengajuan_id"`
	Disetujui   bool      `json:"disetujui"`
	Catatan     string    `json:"catatan"`
	ApprovedBy  int64     `json:"approved_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Pencairan struct {
	PengajuanID   int64     `json:"pengajuan_id"`
	Nominal       float64   `json:"nominal"`
	TanggalCair   time.Time `json:"tanggal_cair"`
	Catatan       string    `json:"catatan"`
	AdminUserID   int64     `json:"admin_user_id"`
	CreatedAt     time.Time `json:"created_at"`
}