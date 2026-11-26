package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
)

func (s Store) Create(ctx context.Context, p *domain.Pengajuan) error {
	const q = `INSERT INTO pengajuan (nomor, dealer_user_id, status) VALUES (?, ?, ?)`
	res, err := s.DB.ExecContext(ctx, q, p.Nomor, p.DealerUserID, p.Status)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	_, _ = s.DB.ExecContext(ctx, `INSERT INTO data_konsumen (pengajuan_id) VALUES (?)`, id)
	_, _ = s.DB.ExecContext(ctx, `INSERT INTO data_kendaraan (pengajuan_id) VALUES (?)`, id)
	_, _ = s.DB.ExecContext(ctx, `INSERT INTO data_pinjaman (pengajuan_id) VALUES (?)`, id)
	return nil
}

func (s Store) UpdateStatus(ctx context.Context, id int64, status string, alasan string) error {
	// Alasan hanya ditulis saat ditolak, supaya catatan approval tidak tertukar.
	if status == "ditolak" {
		_, err := s.DB.ExecContext(ctx, `UPDATE pengajuan SET status = ?, alasan_penolakan = ? WHERE id = ?`, status, alasan, id)
		return err
	}
	_, err := s.DB.ExecContext(ctx, `UPDATE pengajuan SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s Store) SetMarketing(ctx context.Context, id int64, marketingID int64) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE pengajuan SET marketing_user_id = ? WHERE id = ?`, marketingID, id)
	return err
}

func (s Store) FindByID(ctx context.Context, id int64) (*domain.Pengajuan, error) {
	const q = `SELECT id, nomor, dealer_user_id, marketing_user_id, status, COALESCE(alasan_penolakan, ''), created_at, updated_at
		FROM pengajuan WHERE id = ?`
	var p domain.Pengajuan
	var marketing sql.NullInt64
	if err := s.DB.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.Nomor, &p.DealerUserID, &marketing, &p.Status, &p.AlasanPenolakan, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if marketing.Valid {
		p.MarketingUserID = &marketing.Int64
	}
	konsumen, err := s.getKonsumen(ctx, id)
	if err != nil {
		return nil, err
	}
	kendaraan, err := s.getKendaraan(ctx, id)
	if err != nil {
		return nil, err
	}
	pinjaman, err := s.getPinjaman(ctx, id)
	if err != nil {
		return nil, err
	}
	dokumen, err := s.ListDokumen(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Konsumen = konsumen
	p.Kendaraan = kendaraan
	p.Pinjaman = pinjaman
	p.Dokumen = dokumen
	return &p, nil
}

func (s Store) List(ctx context.Context, filter ListFilter) ([]domain.Pengajuan, error) {
	var conds []string
	var args []any
	if filter.DealerUserID != nil {
		conds = append(conds, "dealer_user_id = ?")
		args = append(args, *filter.DealerUserID)
	}
	if len(filter.StatusIn) > 0 {
		placeholders := strings.Repeat("?,", len(filter.StatusIn))
		placeholders = placeholders[:len(placeholders)-1]
		conds = append(conds, fmt.Sprintf("status IN (%s)", placeholders))
		for _, st := range filter.StatusIn {
			args = append(args, st)
		}
	}
	query := `SELECT id, nomor, dealer_user_id, marketing_user_id, status, COALESCE(alasan_penolakan, ''), created_at, updated_at FROM pengajuan`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id DESC"
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Pengajuan
	for rows.Next() {
		var p domain.Pengajuan
		var marketing sql.NullInt64
		if err := rows.Scan(&p.ID, &p.Nomor, &p.DealerUserID, &marketing, &p.Status, &p.AlasanPenolakan, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if marketing.Valid {
			p.MarketingUserID = &marketing.Int64
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s Store) UpsertKonsumen(ctx context.Context, data domain.Konsumen) error {
	const q = `UPDATE data_konsumen SET nama=?, nik=?, tanggal_lahir=?, status_perkawinan=?, nama_pasangan=?, nik_pasangan=? WHERE pengajuan_id=?`
	_, err := s.DB.ExecContext(ctx, q, data.Nama, data.NIK, data.TanggalLahir, data.StatusPerkawinan, data.NamaPasangan, data.NIKPasangan, data.PengajuanID)
	return err
}

func (s Store) UpsertKendaraan(ctx context.Context, data domain.Kendaraan) error {
	const q = `UPDATE data_kendaraan SET dealer=?, merk=?, model=?, tipe=?, warna=?, harga=? WHERE pengajuan_id=?`
	_, err := s.DB.ExecContext(ctx, q, data.Dealer, data.Merk, data.Model, data.Tipe, data.Warna, data.Harga, data.PengajuanID)
	return err
}

func (s Store) UpsertPinjaman(ctx context.Context, data domain.Pinjaman) error {
	const q = `UPDATE data_pinjaman SET asuransi=?, uang_muka=?, tenor_bulan=?, angsuran_bulanan=? WHERE pengajuan_id=?`
	_, err := s.DB.ExecContext(ctx, q, data.Asuransi, data.UangMuka, data.TenorBulan, data.AngsuranBulanan, data.PengajuanID)
	return err
}

func (s Store) AddDokumen(ctx context.Context, doc domain.Dokumen) error {
	// Jenis yang sama diganti, supaya satu berkas tidak punya dua KTP.
	_, _ = s.DB.ExecContext(ctx, `DELETE FROM dokumen WHERE pengajuan_id = ? AND jenis = ?`, doc.PengajuanID, doc.Jenis)
	const q = `INSERT INTO dokumen (pengajuan_id, jenis, nama_file, path_file, uploaded_by) VALUES (?, ?, ?, ?, ?)`
	res, err := s.DB.ExecContext(ctx, q, doc.PengajuanID, doc.Jenis, doc.NamaFile, doc.PathFile, doc.UploadedBy)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	doc.ID = id
	return nil
}

func (s Store) ListDokumen(ctx context.Context, pengajuanID int64) ([]domain.Dokumen, error) {
	const q = `SELECT id, pengajuan_id, jenis, nama_file, path_file, uploaded_by, created_at FROM dokumen WHERE pengajuan_id=? ORDER BY id`
	rows, err := s.DB.QueryContext(ctx, q, pengajuanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []domain.Dokumen
	for rows.Next() {
		var d domain.Dokumen
		if err := rows.Scan(&d.ID, &d.PengajuanID, &d.Jenis, &d.NamaFile, &d.PathFile, &d.UploadedBy, &d.CreatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s Store) FindDokumen(ctx context.Context, docID int64) (*domain.Dokumen, error) {
	const q = `SELECT id, pengajuan_id, jenis, nama_file, path_file, uploaded_by, created_at FROM dokumen WHERE id=?`
	var d domain.Dokumen
	if err := s.DB.QueryRowContext(ctx, q, docID).Scan(&d.ID, &d.PengajuanID, &d.Jenis, &d.NamaFile, &d.PathFile, &d.UploadedBy, &d.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s Store) AddRiwayat(ctx context.Context, pengajuanID int64, lama, baru, catatan string, userID int64) error {
	const q = `INSERT INTO riwayat_status (pengajuan_id, status_lama, status_baru, catatan, user_id) VALUES (?, ?, ?, ?, ?)`
	_, err := s.DB.ExecContext(ctx, q, pengajuanID, lama, baru, nullIfEmpty(catatan), userID)
	return err
}

func (s Store) SaveApproval(ctx context.Context, approval domain.Approval) error {
	const q = `INSERT INTO approval_pengajuan (pengajuan_id, disetujui, catatan, approved_by) VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE disetujui=VALUES(disetujui), catatan=VALUES(catatan), approved_by=VALUES(approved_by), created_at=CURRENT_TIMESTAMP`
	_, err := s.DB.ExecContext(ctx, q, approval.PengajuanID, approval.Disetujui, nullIfEmpty(approval.Catatan), approval.ApprovedBy)
	return err
}

func (s Store) SavePencairan(ctx context.Context, pencairan domain.Pencairan) error {
	const q = `INSERT INTO pencairan_dana (pengajuan_id, nominal, tanggal_cair, catatan, admin_user_id) VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE nominal=VALUES(nominal), tanggal_cair=VALUES(tanggal_cair), catatan=VALUES(catatan), admin_user_id=VALUES(admin_user_id)`
	_, err := s.DB.ExecContext(ctx, q, pencairan.PengajuanID, pencairan.Nominal, pencairan.TanggalCair.Format("2006-01-02"), nullIfEmpty(pencairan.Catatan), pencairan.AdminUserID)
	return err
}

func (s Store) FindPencairan(ctx context.Context, pengajuanID int64) (*domain.Pencairan, error) {
	const q = `SELECT pengajuan_id, nominal, tanggal_cair, COALESCE(catatan,''), admin_user_id, created_at FROM pencairan_dana WHERE pengajuan_id=?`
	var p domain.Pencairan
	var t time.Time
	if err := s.DB.QueryRowContext(ctx, q, pengajuanID).Scan(&p.PengajuanID, &p.Nominal, &t, &p.Catatan, &p.AdminUserID, &p.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.TanggalCair = t
	return &p, nil
}

func (s Store) NextNomor(ctx context.Context) (string, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) FROM pengajuan`).Scan(&n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PJ-%06d", n+1), nil
}

func (s Store) getKonsumen(ctx context.Context, pengajuanID int64) (*domain.Konsumen, error) {
	const q = `SELECT pengajuan_id, nama, nik, tanggal_lahir, status_perkawinan, nama_pasangan, nik_pasangan FROM data_konsumen WHERE pengajuan_id=?`
	var k domain.Konsumen
	var t sql.NullTime
	if err := s.DB.QueryRowContext(ctx, q, pengajuanID).Scan(&k.PengajuanID, &k.Nama, &k.NIK, &t, &k.StatusPerkawinan, &k.NamaPasangan, &k.NIKPasangan); err != nil {
		return nil, err
	}
	if t.Valid {
		k.TanggalLahir = &t.Time
	}
	return &k, nil
}

func (s Store) getKendaraan(ctx context.Context, pengajuanID int64) (*domain.Kendaraan, error) {
	const q = `SELECT pengajuan_id, dealer, merk, model, tipe, warna, harga FROM data_kendaraan WHERE pengajuan_id=?`
	var k domain.Kendaraan
	if err := s.DB.QueryRowContext(ctx, q, pengajuanID).Scan(&k.PengajuanID, &k.Dealer, &k.Merk, &k.Model, &k.Tipe, &k.Warna, &k.Harga); err != nil {
		return nil, err
	}
	return &k, nil
}

func (s Store) getPinjaman(ctx context.Context, pengajuanID int64) (*domain.Pinjaman, error) {
	const q = `SELECT pengajuan_id, asuransi, uang_muka, tenor_bulan, angsuran_bulanan FROM data_pinjaman WHERE pengajuan_id=?`
	var p domain.Pinjaman
	if err := s.DB.QueryRowContext(ctx, q, pengajuanID).Scan(&p.PengajuanID, &p.Asuransi, &p.UangMuka, &p.TenorBulan, &p.AngsuranBulanan); err != nil {
		return nil, err
	}
	return &p, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}