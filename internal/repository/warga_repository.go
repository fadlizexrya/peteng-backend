package repository

import (
	"context"
	"peteng-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WargaRepository struct {
	db *pgxpool.Pool
}

func NewWargaRepository(db *pgxpool.Pool) *WargaRepository {
	return &WargaRepository{db: db}
}

func (r *WargaRepository) CreateWarga(ctx context.Context, warga *domain.Warga) error {
	query := `
		INSERT INTO warga (nama, email, password_email, nomor_telepon)
		VALUES ($1, $2, $3, $4)
		RETURNING id_warga, created_at
	`
	return r.db.QueryRow(ctx, query, warga.Nama, warga.Email, warga.PasswordEmail, warga.NomorTelepon).
		Scan(&warga.IDWarga, &warga.CreatedAt)
}

func (r *WargaRepository) GetByEmail(ctx context.Context, email string) (*domain.Warga, error) {
	query := `SELECT id_warga, nama, email, password_email, nomor_telepon FROM warga WHERE email = $1`
	var w domain.Warga
	err := r.db.QueryRow(ctx, query, email).
		Scan(&w.IDWarga, &w.Nama, &w.Email, &w.PasswordEmail, &w.NomorTelepon)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
