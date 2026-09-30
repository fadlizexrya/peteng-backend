package usecase

import (
	"context"
	"errors"

	"peteng-backend/internal/auth"
	"peteng-backend/internal/domain"
	"peteng-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type WargaUsecase struct {
	repo *repository.WargaRepository
	jwt  *auth.JWTManager
}

func NewWargaUsecase(
	repo *repository.WargaRepository,
	jwt *auth.JWTManager,
) *WargaUsecase {
	return &WargaUsecase{
		repo: repo,
		jwt:  jwt,
	}
}

func (u *WargaUsecase) Register(
	ctx context.Context,
	req *domain.RegisterReq,
) (*domain.Warga, error) {

	// Hash password sebelum disimpan ke database.
	hashedPwd, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

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

	// Jangan kirim hash password ke Flutter.
	warga.PasswordEmail = ""

	return warga, nil
}

type LoginResult struct {
	Token string        `json:"token"`
	User  *domain.Warga `json:"user"`
}

func (u *WargaUsecase) Login(
	ctx context.Context,
	req *domain.LoginReq,
) (*LoginResult, error) {

	warga, err := u.repo.GetByEmail(
		ctx,
		req.Email,
	)

	if err != nil {
		return nil, errors.New("email tidak ditemukan")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(warga.PasswordEmail),
		[]byte(req.Password),
	)

	if err != nil {
		return nil, errors.New("password salah")
	}

	// Buat JWT setelah email dan password benar.
	token, err := u.jwt.GenerateToken(
		warga.IDWarga,
		warga.Email,
	)

	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	// Jangan pernah kirim hash password.
	warga.PasswordEmail = ""

	return &LoginResult{
		Token: token,
		User:  warga,
	}, nil
}

func (u *WargaUsecase) BuatLaporan(
	ctx context.Context,
	req *domain.LaporanReq,
) error {
	return u.repo.CreateLaporan(ctx, req)
}