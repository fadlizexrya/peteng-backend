package repository

import (
	"context"
	"errors"
	"time"

	"peteng-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound dipakai usecase untuk membedakan "tidak ada baris" dari error DB.
var ErrNotFound = errors.New("data tidak ditemukan")

type WargaRepository struct {
	db *pgxpool.Pool
}

func NewWargaRepository(db *pgxpool.Pool) *WargaRepository {
	return &WargaRepository{db: db}
}

// EnsureAuthSchema menambahkan kolom yang dibutuhkan fitur verifikasi email dan
// login Google. Aman dijalankan berulang kali (IF NOT EXISTS).
//
// Baris lama diberi email_verified = TRUE supaya akun yang sudah ada tidak
// tiba-tiba terkunci, lalu default kolom diubah menjadi FALSE untuk akun baru.
func (r *WargaRepository) EnsureAuthSchema(ctx context.Context) error {
	statements := []string{
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS email_verified boolean NOT NULL DEFAULT true`,
		`ALTER TABLE public.warga ALTER COLUMN email_verified SET DEFAULT false`,
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS google_id text`,
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS auth_provider text NOT NULL DEFAULT 'local'`,
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS verification_token text`,
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS verification_expires_at timestamptz`,
		`ALTER TABLE public.warga ADD COLUMN IF NOT EXISTS verification_attempts integer NOT NULL DEFAULT 0`,
	}

	for _, statement := range statements {
		if _, err := r.db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	// Index unik untuk google_id. Kalau gagal (mis. ada data dobel), jangan
	// hentikan aplikasi; cukup dilaporkan oleh pemanggil.
	_, err := r.db.Exec(
		ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS warga_google_id_unique
		 ON public.warga (google_id) WHERE google_id IS NOT NULL`,
	)

	return err
}

func (r *WargaRepository) CreateWarga(ctx context.Context, warga *domain.Warga) error {
	query := `
		INSERT INTO warga (
			nama,
			email,
			password_email,
			nomor_telepon,
			email_verified,
			google_id,
			auth_provider,
			verification_token,
			verification_expires_at,
			verification_attempts
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0)
		RETURNING id_warga, created_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		warga.Nama,
		warga.Email,
		warga.PasswordEmail,
		warga.NomorTelepon,
		warga.EmailVerified,
		warga.GoogleID,
		warga.AuthProvider,
		nullableString(warga.VerificationToken),
		warga.VerificationExpiresAt,
	).Scan(&warga.IDWarga, &warga.CreatedAt)
}

func (r *WargaRepository) GetByEmail(ctx context.Context, email string) (*domain.Warga, error) {
	query := `
		SELECT
			id_warga,
			nama,
			email,
			password_email,
			nomor_telepon,
			email_verified,
			google_id,
			auth_provider,
			COALESCE(verification_token, ''),
			verification_expires_at,
			COALESCE(verification_attempts, 0),
			created_at
		FROM warga
		WHERE lower(email) = lower($1)
		LIMIT 1
	`

	return scanWarga(r.db.QueryRow(ctx, query, email))
}

func (r *WargaRepository) GetByGoogleID(ctx context.Context, googleID string) (*domain.Warga, error) {
	query := `
		SELECT
			id_warga,
			nama,
			email,
			password_email,
			nomor_telepon,
			email_verified,
			google_id,
			auth_provider,
			COALESCE(verification_token, ''),
			verification_expires_at,
			COALESCE(verification_attempts, 0),
			created_at
		FROM warga
		WHERE google_id = $1
		LIMIT 1
	`

	return scanWarga(r.db.QueryRow(ctx, query, googleID))
}

func (r *WargaRepository) SetVerification(
	ctx context.Context,
	email string,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE warga
		 SET verification_token = $2,
		     verification_expires_at = $3,
		     verification_attempts = 0
		 WHERE lower(email) = lower($1)`,
		email,
		tokenHash,
		expiresAt,
	)

	return err
}

func (r *WargaRepository) IncrementVerificationAttempts(
	ctx context.Context,
	email string,
) (int, error) {
	var attempts int

	err := r.db.QueryRow(
		ctx,
		`UPDATE warga
		 SET verification_attempts = COALESCE(verification_attempts, 0) + 1
		 WHERE lower(email) = lower($1)
		 RETURNING verification_attempts`,
		email,
	).Scan(&attempts)

	return attempts, err
}

func (r *WargaRepository) ClearVerification(ctx context.Context, email string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE warga
		 SET verification_token = NULL,
		     verification_expires_at = NULL
		 WHERE lower(email) = lower($1)`,
		email,
	)

	return err
}

func (r *WargaRepository) MarkEmailVerified(ctx context.Context, email string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE warga
		 SET email_verified = true,
		     verification_token = NULL,
		     verification_expires_at = NULL,
		     verification_attempts = 0
		 WHERE lower(email) = lower($1)`,
		email,
	)

	return err
}

func (r *WargaRepository) UpdateGoogleID(ctx context.Context, idWarga int, googleID string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE warga SET google_id = $2 WHERE id_warga = $1`,
		idWarga,
		googleID,
	)

	return err
}

func scanWarga(row pgx.Row) (*domain.Warga, error) {
	var warga domain.Warga

	err := row.Scan(
		&warga.IDWarga,
		&warga.Nama,
		&warga.Email,
		&warga.PasswordEmail,
		&warga.NomorTelepon,
		&warga.EmailVerified,
		&warga.GoogleID,
		&warga.AuthProvider,
		&warga.VerificationToken,
		&warga.VerificationExpiresAt,
		&warga.VerificationAttempts,
		&warga.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &warga, nil
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}

	return value
}
