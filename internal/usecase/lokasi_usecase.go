package usecase

import (
	"context"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"
)

type LokasiUsecase struct {
	repo *repository.LokasiRepository
}

func NewLokasiUsecase(
	repo *repository.LokasiRepository,
) *LokasiUsecase {
	return &LokasiUsecase{
		repo: repo,
	}
}

func (u *LokasiUsecase) GetAll(
	ctx context.Context,
) ([]domain.Lokasi, error) {

	return u.repo.GetAll(ctx)
}
