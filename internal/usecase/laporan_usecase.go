package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"

	"github.com/jackc/pgx/v5"
)

type LaporanUsecase struct {
	laporanRepo  *repository.LaporanRepository
	progresRepo  *repository.ProgresLaporanRepository
	validasiRepo *repository.ValidasiAIRepository
	validasiAI   *ValidasiAIUsecase
}

func NewLaporanUsecase(
	laporanRepo *repository.LaporanRepository,
	progresRepo *repository.ProgresLaporanRepository,
	validasiAI ...*ValidasiAIUsecase,
) *LaporanUsecase {
	u := &LaporanUsecase{
		laporanRepo: laporanRepo,
		progresRepo: progresRepo,
	}
	if len(validasiAI) > 0 {
		u.validasiAI = validasiAI[0]
		u.validasiRepo = validasiAI[0].repo
	}
	return u
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
	case "pju_mati_total",
		"pju_redup_kedip",
		"tiang_roboh_rusak",
		"belum_ada_jaringan",
		"lainnya":
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

	if u.validasiAI != nil {
		// Process after returning the successful report response. If AI is
		// unavailable, keep the report and expose a retryable failed status.
		go func(id string) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			if err := u.validasiAI.ProsesValidasiAI(ctx, id, u.laporanRepo); err != nil {
				log.Printf("validasi AI laporan %s gagal: %v", id, err)
			}
		}(laporan.IDLaporan)
	}

	return laporan, nil
}

func (u *LaporanUsecase) DaftarLaporanWarga(
	ctx context.Context,
	idWarga int,
) ([]domain.Laporan, error) {
	laporans, err := u.laporanRepo.GetByWargaID(ctx, idWarga)
	if err != nil {
		return nil, err
	}
	if err := u.attachAIResults(ctx, laporans); err != nil {
		return nil, err
	}
	return laporans, nil
}

func (u *LaporanUsecase) attachAIResults(ctx context.Context, laporans []domain.Laporan) error {
	if u.validasiRepo == nil {
		return nil
	}
	for i := range laporans {
		hasil, err := u.validasiRepo.GetTerbaru(ctx, laporans[i].IDLaporan)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		laporans[i].HasilValidasiAI = hasil
	}
	return nil
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

	if u.validasiRepo != nil {
		hasil, err := u.validasiRepo.GetTerbaru(ctx, laporan.IDLaporan)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		laporan.HasilValidasiAI = hasil
	}
	detail := &domain.LaporanDetail{
		Laporan:       *laporan,
		HasilValidasi: laporan.HasilValidasiAI,
		Progres:       progres,
	}
	return detail, nil
}

func (u *LaporanUsecase) DaftarSemuaLaporan(
	ctx context.Context,
) ([]domain.Laporan, error) {
	laporans, err := u.laporanRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if u.validasiRepo == nil {
		return laporans, nil
	}
	if err := u.attachAIResults(ctx, laporans); err != nil {
		return nil, err
	}
	return laporans, nil
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
