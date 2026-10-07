package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type LaporanUsecase struct {
	laporanRepo *repository.LaporanRepository
	progresRepo *repository.ProgresLaporanRepository
}

func NewLaporanUsecase(
	laporanRepo *repository.LaporanRepository,
	progresRepo *repository.ProgresLaporanRepository,
) *LaporanUsecase {
	return &LaporanUsecase{
		laporanRepo: laporanRepo,
		progresRepo: progresRepo,
	}
}

func (u *LaporanUsecase) BuatLaporan(
	ctx context.Context,
	idWarga int,
	req *domain.LaporanReq,
) (*domain.Laporan, error) {

	if req.TingkatKegelapan < 1 ||
		req.TingkatKegelapan > 3 {
		return nil, errors.New(
			"tingkat_kegelapan harus bernilai 1, 2, atau 3",
		)
	}

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

	err := u.laporanRepo.Create(
		ctx,
		laporan,
	)

	if err != nil {
		return nil, err
	}

	err = u.progresRepo.Create(
		ctx,
		laporan.IDLaporan,
		"Laporan berhasil dibuat",
		"Laporan warga berhasil diterima dan menunggu proses validasi AI.",
	)

	if err != nil {
		return nil, err
	}

	return laporan, nil
}
