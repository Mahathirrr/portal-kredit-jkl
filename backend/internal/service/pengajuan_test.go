package service

import (
	"testing"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
)

func TestValidNIK(t *testing.T) {
	if !validNIK("3201012345678901") {
		t.Fatal("NIK 16 digit harus diterima")
	}
	if validNIK("320101234567890A") {
		t.Fatal("NIK berisi huruf harus ditolak")
	}
}

func TestRequireDocs(t *testing.T) {
	svc := PengajuanService{}
	p := &domain.Pengajuan{
		Dokumen: []domain.Dokumen{
			{Jenis: "ktp"},
			{Jenis: "spk"},
		},
	}
	err := svc.requireDocs(p, []string{"ktp", "spk", "bukti_tanda_jadi"})
	if err == nil {
		t.Fatal("seharusnya dokumen kurang terdeteksi")
	}
}