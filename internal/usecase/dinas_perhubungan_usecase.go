package usecase

import (
	"context"
	"errors"
	"peteng-backend/internal/entity"
	"peteng-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type DinasPerhubunganUsecase interface {
	Register(ctx context.Context, req entity.RegisterDinasRequest) (*entity.DinasPerhubungan, error)
	Login(ctx context.Context, req entity.LoginDinasRequest) (*entity.DinasPerhubungan, error)
}

type dinasUsecase struct {
	repo repository.DinasPerhubunganRepository
}

func NewDinasPerhubunganUsecase(repo repository.DinasPerhubunganRepository) DinasPerhubunganUsecase {
	return &dinasUsecase{repo: repo}
}

func (u *dinasUsecase) Register(ctx context.Context, req entity.RegisterDinasRequest) (*entity.DinasPerhubungan, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.PasswordEmail), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	dinas := &entity.DinasPerhubungan{
		NamaDinasPerhubungan: req.NamaDinasPerhubungan,
		Email:                req.Email,
		PasswordEmail:        string(hashedPassword),
	}

	err = u.repo.Create(ctx, dinas)
	if err != nil {
		return nil, err
	}

	return dinas, nil
}

func (u *dinasUsecase) Login(ctx context.Context, req entity.LoginDinasRequest) (*entity.DinasPerhubungan, error) {
	dinas, err := u.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dinas.PasswordEmail), []byte(req.PasswordEmail))
	if err != nil {
		return nil, errors.New("password salah")
	}

	return dinas, nil
}
