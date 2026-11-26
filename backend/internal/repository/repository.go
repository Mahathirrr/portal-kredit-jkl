package repository

import (
	"context"
	"database/sql"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindUserByID(ctx context.Context, id int64) (*domain.User, error)
	CreateDemoUsersIfEmpty(ctx context.Context, users []domain.User) error
}

type PengajuanRepository interface {
	Create(ctx context.Context, p *domain.Pengajuan) error
	UpdateStatus(ctx context.Context, id int64, status string, alasan string) error
	SetMarketing(ctx context.Context, id int64, marketingID int64) error
	FindByID(ctx context.Context, id int64) (*domain.Pengajuan, error)
	List(ctx context.Context, filter ListFilter) ([]domain.Pengajuan, error)
	UpsertKonsumen(ctx context.Context, data domain.Konsumen) error
	UpsertKendaraan(ctx context.Context, data domain.Kendaraan) error
	UpsertPinjaman(ctx context.Context, data domain.Pinjaman) error
	AddDokumen(ctx context.Context, doc domain.Dokumen) error
	ListDokumen(ctx context.Context, pengajuanID int64) ([]domain.Dokumen, error)
	FindDokumen(ctx context.Context, docID int64) (*domain.Dokumen, error)
	AddRiwayat(ctx context.Context, pengajuanID int64, lama, baru, catatan string, userID int64) error
	SaveApproval(ctx context.Context, approval domain.Approval) error
	SavePencairan(ctx context.Context, pencairan domain.Pencairan) error
	FindPencairan(ctx context.Context, pengajuanID int64) (*domain.Pencairan, error)
	NextNomor(ctx context.Context) (string, error)
}

type ListFilter struct {
	DealerUserID *int64
	StatusIn     []string
}

type Store struct {
	DB *sql.DB
}