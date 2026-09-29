package usecase

import (
	"context"
	"errors"
	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type WargaUsecase struct {
	repo *repository.WargaRepository
}

func NewWargaUsecase(repo *repository.WargaRepository) *WargaUsecase {
	return &WargaUsecase{repo: repo}
}

func (u *WargaUsecase) Register(ctx context.Context, req *domain.RegisterReq) (*domain.Warga, error) {
	// Hash Password demi keamanan
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	warga := &domain.Warga{
		Nama:          req.Nama,
		Email:         req.Email,
		PasswordEmail: string(hashedPwd),
		NomorTelepon:  req.NomorTelepon,
	}

	err = u.repo.CreateWarga(ctx, warga)
	if err != nil {
		return nil, err
	}
	warga.PasswordEmail = "" // Kosongkan hash di response
	return warga, nil
}

func (u *WargaUsecase) Login(ctx context.Context, req *domain.LoginReq) (*domain.Warga, error) {
	warga, err := u.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("email tidak ditemukan")
	}

	err = bcrypt.CompareHashAndPassword([]byte(warga.PasswordEmail), []byte(req.Password))
	if err != nil {
		return nil, errors.New("password salah")
	}

	warga.PasswordEmail = ""
	return warga, nil
}

func (u *WargaUsecase) BuatLaporan(ctx context.Context, req *domain.LaporanReq) error {
	return u.repo.CreateLaporan(ctx, req)
}