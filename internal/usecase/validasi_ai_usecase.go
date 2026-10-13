package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type ValidasiAIUsecase struct {
	repo     *repository.ValidasiAIRepository
	aiClient AIServiceClient
}

func NewValidasiAIUsecase(
	repo *repository.ValidasiAIRepository,
	aiClient AIServiceClient,
) *ValidasiAIUsecase {
	return &ValidasiAIUsecase{repo: repo, aiClient: aiClient}
}

func (u *ValidasiAIUsecase) ProsesValidasiAI(
	ctx context.Context,
	idLaporan string,
	laporanRepo *repository.LaporanRepository,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}
	if u.aiClient == nil {
		return errors.New("ai client tidak terkonfigurasi")
	}

	// Status transition already starts in MulaiValidasi. This function assumes diproses_ai.
	req := &domain.ValidasiAIRequest{
		IDLaporan:              idLaporan,
		FotoURL:                 "",
		Kategori:                "PJU_RUSAK",
		TingkatKegelapanManual:  1,
	}

	hasil, err := u.aiClient.ValidasiLaporan(ctx, req)
	if err != nil {
		return err
	}

	return u.repo.SimpanHasilAI(ctx, idLaporan, hasil)
}

func (u *ValidasiAIUsecase) MulaiValidasi(
	ctx context.Context,
	idLaporan string,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}

	return u.repo.MulaiValidasi(ctx, idLaporan)
}

// Simulasi ini hanya untuk pengembangan dan pengujian.
func (u *ValidasiAIUsecase) SimpanHasilMock(
	ctx context.Context,
	idLaporan string,
	req *domain.ValidasiAIMockRequest,
) error {
	if idLaporan == "" {
		return errors.New("id laporan wajib diisi")
	}

	if req.HasilValidasi != "valid" &&
		req.HasilValidasi != "tidak_valid" {
		return errors.New(
			"hasil_validasi harus valid atau tidak_valid",
		)
	}

	if req.TingkatKegelapanAI < 1 ||
		req.TingkatKegelapanAI > 3 {
		return errors.New(
			"tingkat_kegelapan_ai harus bernilai 1, 2, atau 3",
		)
	}

	if req.Catatan == "" {
		return errors.New("catatan wajib diisi")
	}

	return u.repo.SimpanHasilMock(ctx, idLaporan, req)
}

func (u *ValidasiAIUsecase) GetTerbaru(
	ctx context.Context,
	idLaporan string,
) (*domain.ValidasiAI, error) {
	if idLaporan == "" {
		return nil, errors.New("id laporan wajib diisi")
	}

	return u.repo.GetTerbaru(ctx, idLaporan)
}
