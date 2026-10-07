package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgresLaporanRepository struct {
	db *pgxpool.Pool
}

func NewProgresLaporanRepository(
	db *pgxpool.Pool,
) *ProgresLaporanRepository {
	return &ProgresLaporanRepository{
		db: db,
	}
}

func (r *ProgresLaporanRepository) Create(
	ctx context.Context,
	idLaporan string,
	judul string,
	deskripsi string,
) error {

	query := `
		INSERT INTO progres_laporan (
			id_laporan,
			judul,
			deskripsi
		)
		VALUES (
			$1,
			$2,
			$3
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		idLaporan,
		judul,
		deskripsi,
	)

	return err
}
