package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
	"github.com/Mahathirrr/kredit-jkl/internal/repository"
)

var (
	errForbidden = errors.New("tidak memiliki akses")
	errStatus    = errors.New("status pengajuan tidak valid untuk aksi ini")
)

type PengajuanService struct {
	Repo repository.PengajuanRepository
}

func (s PengajuanService) ListForUser(ctx context.Context, user domain.User) ([]domain.Pengajuan, error) {
	filter := repository.ListFilter{}
	switch user.Role {
	case domain.RoleDealer:
		filter.DealerUserID = &user.ID
	case domain.RoleMarketing:
		// Marketing melihat berkas setelah dealer mengirim, termasuk yang ditolak.
		filter.StatusIn = []string{
			domain.StatusMenungguMarketing,
			domain.StatusMenungguApproval,
			domain.StatusDitolak,
			domain.StatusMenungguTTD,
			domain.StatusTTDLengkap,
			domain.StatusDanaCair,
		}
	case domain.RoleSupervisor:
		filter.StatusIn = []string{
			domain.StatusMenungguApproval,
			domain.StatusDitolak,
			domain.StatusMenungguTTD,
			domain.StatusTTDLengkap,
			domain.StatusDanaCair,
		}
	case domain.RoleAdmin:
		filter.StatusIn = []string{domain.StatusMenungguTTD, domain.StatusTTDLengkap, domain.StatusDanaCair}
	}
	return s.Repo.List(ctx, filter)
}

func (s PengajuanService) CreateDraft(ctx context.Context, dealer domain.User) (*domain.Pengajuan, error) {
	if dealer.Role != domain.RoleDealer {
		return nil, errForbidden
	}
	nomor, err := s.Repo.NextNomor(ctx)
	if err != nil {
		return nil, err
	}
	p := &domain.Pengajuan{
		Nomor:        nomor,
		DealerUserID: dealer.ID,
		Status:       domain.StatusDraftDealer,
	}
	if err := s.Repo.Create(ctx, p); err != nil {
		return nil, err
	}
	_ = s.Repo.AddRiwayat(ctx, p.ID, "", domain.StatusDraftDealer, "draft dibuat", dealer.ID)
	return s.Repo.FindByID(ctx, p.ID)
}

func (s PengajuanService) Get(ctx context.Context, user domain.User, id int64) (*domain.Pengajuan, error) {
	p, err := s.Repo.FindByID(ctx, id)
	if err != nil || p == nil {
		return p, err
	}
	if !canView(user, p) {
		return nil, errForbidden
	}
	return p, nil
}

// canView menjaga tiap peran hanya membuka berkas pada tahap yang menjadi tugasnya.
func canView(user domain.User, p *domain.Pengajuan) bool {
	switch user.Role {
	case domain.RoleDealer:
		return p.DealerUserID == user.ID
	case domain.RoleMarketing:
		return p.Status != domain.StatusDraftDealer
	case domain.RoleSupervisor:
		switch p.Status {
		case domain.StatusMenungguApproval, domain.StatusDitolak, domain.StatusMenungguTTD, domain.StatusTTDLengkap, domain.StatusDanaCair:
			return true
		default:
			return false
		}
	case domain.RoleAdmin:
		switch p.Status {
		case domain.StatusMenungguTTD, domain.StatusTTDLengkap, domain.StatusDanaCair:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (s PengajuanService) UpdateKonsumen(ctx context.Context, user domain.User, data domain.Konsumen) error {
	p, err := s.mustAccess(ctx, user, data.PengajuanID)
	if err != nil {
		return err
	}
	if user.Role == domain.RoleDealer && p.Status != domain.StatusDraftDealer {
		return errStatus
	}
	if user.Role == domain.RoleMarketing && p.Status != domain.StatusMenungguMarketing {
		return errStatus
	}
	if user.Role != domain.RoleDealer && user.Role != domain.RoleMarketing {
		return errForbidden
	}
	if !validNIK(data.NIK) {
		return errors.New("NIK harus 16 digit angka")
	}
	// Marketing hanya menambah data yang belum ada. Nama dan NIK dari dealer tidak ditulis ulang.
	if user.Role == domain.RoleMarketing && p.Konsumen != nil {
		data.Nama = p.Konsumen.Nama
		data.NIK = p.Konsumen.NIK
	}
	return s.Repo.UpsertKonsumen(ctx, data)
}

func (s PengajuanService) UpdateKendaraan(ctx context.Context, user domain.User, data domain.Kendaraan) error {
	p, err := s.mustAccess(ctx, user, data.PengajuanID)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleDealer || p.Status != domain.StatusDraftDealer {
		return errStatus
	}
	if strings.TrimSpace(data.Dealer) == "" || strings.TrimSpace(data.Merk) == "" || data.Harga <= 0 {
		return errors.New("data kendaraan belum lengkap")
	}
	return s.Repo.UpsertKendaraan(ctx, data)
}

func (s PengajuanService) UpdatePinjaman(ctx context.Context, user domain.User, data domain.Pinjaman) error {
	p, err := s.mustAccess(ctx, user, data.PengajuanID)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleMarketing || p.Status != domain.StatusMenungguMarketing {
		return errStatus
	}
	return s.Repo.UpsertPinjaman(ctx, data)
}

func (s PengajuanService) SubmitDealer(ctx context.Context, user domain.User, id int64) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleDealer || p.Status != domain.StatusDraftDealer {
		return errStatus
	}
	if err := s.requireDocs(p, []string{"ktp", "spk", "bukti_tanda_jadi"}); err != nil {
		return err
	}
	if p.Konsumen == nil || strings.TrimSpace(p.Konsumen.Nama) == "" || !validNIK(p.Konsumen.NIK) {
		return errors.New("data konsumen dealer belum lengkap")
	}
	if p.Kendaraan == nil || strings.TrimSpace(p.Kendaraan.Merk) == "" || strings.TrimSpace(p.Kendaraan.Dealer) == "" || p.Kendaraan.Harga <= 0 {
		return errors.New("data kendaraan belum lengkap")
	}
	return s.changeStatus(ctx, p, domain.StatusMenungguMarketing, "", user.ID)
}

func (s PengajuanService) SubmitMarketing(ctx context.Context, user domain.User, id int64) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleMarketing || p.Status != domain.StatusMenungguMarketing {
		return errStatus
	}
	if p.Konsumen == nil || p.Konsumen.TanggalLahir == nil || strings.TrimSpace(p.Konsumen.StatusPerkawinan) == "" {
		return errors.New("data konsumen marketing belum lengkap")
	}
	if strings.EqualFold(p.Konsumen.StatusPerkawinan, "menikah") {
		if strings.TrimSpace(p.Konsumen.NamaPasangan) == "" || !validNIK(p.Konsumen.NIKPasangan) {
			return errors.New("data pasangan belum lengkap")
		}
	}
	if err := s.requireDocs(p, []string{"kartu_keluarga"}); err != nil {
		return err
	}
	if p.Pinjaman == nil || strings.TrimSpace(p.Pinjaman.Asuransi) == "" || p.Pinjaman.TenorBulan <= 0 || p.Pinjaman.AngsuranBulanan <= 0 || p.Pinjaman.UangMuka < 0 {
		return errors.New("data pinjaman belum lengkap")
	}
	_ = s.Repo.SetMarketing(ctx, id, user.ID)
	return s.changeStatus(ctx, p, domain.StatusMenungguApproval, "", user.ID)
}

func (s PengajuanService) Approve(ctx context.Context, user domain.User, id int64, catatan string) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleSupervisor || p.Status != domain.StatusMenungguApproval {
		return errStatus
	}
	_ = s.Repo.SaveApproval(ctx, domain.Approval{PengajuanID: id, Disetujui: true, Catatan: catatan, ApprovedBy: user.ID})
	return s.changeStatus(ctx, p, domain.StatusMenungguTTD, catatan, user.ID)
}

func (s PengajuanService) Reject(ctx context.Context, user domain.User, id int64, alasan string) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleSupervisor || p.Status != domain.StatusMenungguApproval {
		return errStatus
	}
	if strings.TrimSpace(alasan) == "" {
		return errors.New("alasan penolakan wajib diisi")
	}
	_ = s.Repo.SaveApproval(ctx, domain.Approval{PengajuanID: id, Disetujui: false, Catatan: alasan, ApprovedBy: user.ID})
	return s.changeStatus(ctx, p, domain.StatusDitolak, alasan, user.ID)
}

func (s PengajuanService) MarkSignedUploaded(ctx context.Context, user domain.User, id int64) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleMarketing || p.Status != domain.StatusMenungguTTD {
		return errStatus
	}
	if err := s.requireDocs(p, []string{"kontrak_ttd", "po_ttd"}); err != nil {
		return err
	}
	return s.changeStatus(ctx, p, domain.StatusTTDLengkap, "", user.ID)
}

func (s PengajuanService) Disburse(ctx context.Context, user domain.User, id int64, nominal float64, tanggal time.Time, catatan string) error {
	p, err := s.mustAccess(ctx, user, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleAdmin || p.Status != domain.StatusTTDLengkap {
		return errStatus
	}
	if nominal <= 0 {
		return errors.New("nominal pencairan tidak valid")
	}
	_ = s.Repo.SavePencairan(ctx, domain.Pencairan{
		PengajuanID: id,
		Nominal:     nominal,
		TanggalCair: tanggal,
		Catatan:     catatan,
		AdminUserID: user.ID,
	})
	return s.changeStatus(ctx, p, domain.StatusDanaCair, catatan, user.ID)
}

func (s PengajuanService) AddDokumen(ctx context.Context, user domain.User, doc domain.Dokumen) error {
	p, err := s.mustAccess(ctx, user, doc.PengajuanID)
	if err != nil {
		return err
	}
	if err := s.canUpload(user, p, doc.Jenis); err != nil {
		return err
	}
	return s.Repo.AddDokumen(ctx, doc)
}

func (s PengajuanService) canUpload(user domain.User, p *domain.Pengajuan, jenis string) error {
	dealerDocs := map[string]bool{"ktp": true, "spk": true, "bukti_tanda_jadi": true}
	if user.Role == domain.RoleDealer {
		if p.Status != domain.StatusDraftDealer || !dealerDocs[jenis] {
			return errStatus
		}
		return nil
	}
	if user.Role == domain.RoleMarketing {
		if p.Status == domain.StatusMenungguMarketing && jenis == "kartu_keluarga" {
			return nil
		}
		if p.Status == domain.StatusMenungguTTD && (jenis == "kontrak_ttd" || jenis == "po_ttd") {
			return nil
		}
		return errStatus
	}
	return errForbidden
}

func (s PengajuanService) mustAccess(ctx context.Context, user domain.User, id int64) (*domain.Pengajuan, error) {
	return s.Get(ctx, user, id)
}

func (s PengajuanService) changeStatus(ctx context.Context, p *domain.Pengajuan, baru, catatan string, userID int64) error {
	lama := p.Status
	if err := s.Repo.UpdateStatus(ctx, p.ID, baru, catatan); err != nil {
		return err
	}
	return s.Repo.AddRiwayat(ctx, p.ID, lama, baru, catatan, userID)
}

func (s PengajuanService) requireDocs(p *domain.Pengajuan, jenis []string) error {
	have := map[string]bool{}
	for _, d := range p.Dokumen {
		have[d.Jenis] = true
	}
	var missing []string
	for _, j := range jenis {
		if !have[j] {
			missing = append(missing, j)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("dokumen belum lengkap: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validNIK(nik string) bool {
	nik = strings.TrimSpace(nik)
	if len(nik) != 16 {
		return false
	}
	for _, c := range nik {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}