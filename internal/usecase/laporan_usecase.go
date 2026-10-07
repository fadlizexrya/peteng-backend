package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type LaporanUsecase struct {
	repo *repository.LaporanRepository
}

func NewLaporanUsecase(
	repo *repository.LaporanRepository,
) *LaporanUsecase {
	return &LaporanUsecase{
		repo: repo,
	}
}

func (u *LaporanUsecase) BuatLaporan(
	ctx context.Context,
	idWarga int,
	req *domain.LaporanReq,
) (*domain.Laporan, error) {

	// Validasi tingkat kegelapan dari warga.
	if req.TingkatKegelapan < 1 || req.TingkatKegelapan > 3 {
		return nil, errors.New(
			"tingkat_kegelapan harus bernilai 1, 2, atau 3",
		)
	}

	// Validasi kategori laporan.
	switch req.Kategori {
	case "PJU_RUSAK",
		"PJU_TIDAK_ADA",
		"PJU_MATI",
		"LAINNYA":
		// kategori valid
	default:
		return nil, errors.New(
			"kategori laporan tidak valid",
		)
	}

	// ID warga berasal dari JWT, bukan dari request Flutter.
	laporan := &domain.Laporan{
		IDWarga:          idWarga,
		Foto:             req.Foto,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		KeteranganLokasi: req.KeteranganLokasi,
		Kategori:         req.Kategori,
		TingkatKegelapan: req.TingkatKegelapan,
		Deskripsi:        req.Deskripsi,
	}

	err := u.repo.Create(ctx, laporan)
	if err != nil {
		return nil, err
	}

	return laporan, nil
}
