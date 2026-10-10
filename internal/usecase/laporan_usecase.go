package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"

	"github.com/jackc/pgx/v5"
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
	if req.TingkatKegelapan < 1 || req.TingkatKegelapan > 3 {
		return nil, errors.New(
			"tingkat_kegelapan harus bernilai 1, 2, atau 3",
		)
	}

	switch req.Kategori {
	case "PJU_RUSAK", "PJU_TIDAK_ADA", "PJU_MATI", "LAINNYA":
	default:
		return nil, errors.New("kategori laporan tidak valid")
	}

	if req.Foto == "" {
		return nil, errors.New("foto wajib diisi")
	}
	if req.KeteranganLokasi == "" {
		return nil, errors.New("keterangan_lokasi wajib diisi")
	}
	if req.Deskripsi == "" {
		return nil, errors.New("deskripsi wajib diisi")
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, errors.New("latitude tidak valid")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, errors.New("longitude tidak valid")
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

	if err := u.laporanRepo.CreateWithProgress(ctx, laporan); err != nil {
		return nil, err
	}

	return laporan, nil
}

func (u *LaporanUsecase) DaftarLaporanWarga(
	ctx context.Context,
	idWarga int,
) ([]domain.Laporan, error) {
	return u.laporanRepo.GetByWargaID(ctx, idWarga)
}

func (u *LaporanUsecase) DetailLaporanWarga(
	ctx context.Context,
	idWarga int,
	idLaporan string,
) (*domain.LaporanDetail, error) {
	laporan, err := u.laporanRepo.GetByIDAndWargaID(
		ctx,
		idLaporan,
		idWarga,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	progres, err := u.progresRepo.GetByLaporanID(
		ctx,
		laporan.IDLaporan,
	)
	if err != nil {
		return nil, err
	}

	return &domain.LaporanDetail{
		Laporan: *laporan,
		Progres: progres,
	}, nil
}

func (u *LaporanUsecase) DaftarSemuaLaporan(
	ctx context.Context,
) ([]domain.Laporan, error) {
	return u.laporanRepo.GetAll(ctx)
}

func (u *LaporanUsecase) UpdateStatusLaporanDishub(
	ctx context.Context,
	idLaporan string,
	idDishub int,
	req *domain.UpdateStatusLaporanReq,
) error {
	switch req.Status {
	case "diterima", "ditolak", "diproses", "selesai":
	default:
		return errors.New("status laporan tidak valid")
	}

	if req.Judul == "" {
		return errors.New("judul progres wajib diisi")
	}

	if req.Deskripsi == "" {
		return errors.New("deskripsi progres wajib diisi")
	}

	return u.laporanRepo.UpdateStatusWithProgress(
		ctx,
		idLaporan,
		idDishub,
		req.Status,
		req.Judul,
		req.Deskripsi,
	)
}
