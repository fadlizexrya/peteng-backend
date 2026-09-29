package repository

import (
	"context"
	"errors"
	"log"
	"peteng-backend/internal/entity"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DinasPerhubunganRepository interface {
	Create(ctx context.Context, dinas *entity.DinasPerhubungan) error
	FindByEmail(ctx context.Context, email string) (*entity.DinasPerhubungan, error)
}

type dinasRepository struct {
	db *pgxpool.Pool
}

func NewDinasPerhubunganRepository(db *pgxpool.Pool) DinasPerhubunganRepository {
	return &dinasRepository{db: db}
}

func (r *dinasRepository) Create(ctx context.Context, dinas *entity.DinasPerhubungan) error {
	query := `
		INSERT INTO dinas_perhubungan (nama_dinas_perhubungan, email, password_email)
		VALUES ($1, TRIM(LOWER($2)), $3)
		RETURNING id_dinas_perhubungan, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, dinas.NamaDinasPerhubungan, dinas.Email, dinas.PasswordEmail).
		Scan(&dinas.IDDinasPerhubungan, &dinas.CreatedAt, &dinas.UpdatedAt)
}

func (r *dinasRepository) FindByEmail(ctx context.Context, email string) (*entity.DinasPerhubungan, error) {
	query := `
		SELECT id_dinas_perhubungan, nama_dinas_perhubungan, email, password_email, created_at, updated_at
		FROM dinas_perhubungan
		WHERE LOWER(TRIM(email)) = LOWER(TRIM($1))
	`
	dinas := &entity.DinasPerhubungan{}
	cleanEmail := strings.TrimSpace(email)

	err := r.db.QueryRow(ctx, query, cleanEmail).
		Scan(
			&dinas.IDDinasPerhubungan,
			&dinas.NamaDinasPerhubungan,
			&dinas.Email,
			&dinas.PasswordEmail,
			&dinas.CreatedAt,
			&dinas.UpdatedAt,
		)

	if err != nil {
		// Log error asli dari database ke terminal Go untuk mempermudah analisa
		log.Printf("[DATABASE ERROR] FindByEmail gagal untuk '%s': %v", cleanEmail, err)
		return nil, errors.New("petugas/admin dinas perhubungan tidak ditemukan")
	}

	return dinas, nil
}
