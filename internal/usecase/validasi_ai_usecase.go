package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type ValidasiAIUsecase struct {
	repo *repository.ValidasiAIRepository
}

func NewValidasiAIUsecase(
	repo *repository.ValidasiAIRepository,
) *ValidasiAIUsecase {
	return &ValidasiAIUsecase{repo: repo}
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
