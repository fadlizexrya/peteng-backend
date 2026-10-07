package repository

import (
	"context"

	"peteng-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LokasiRepository struct {
	db *pgxpool.Pool
}

func NewLokasiRepository(db *pgxpool.Pool) *LokasiRepository {
	return &LokasiRepository{
		db: db,
	}
}

func (r *LokasiRepository) GetAll(
	ctx context.Context,
) ([]domain.Lokasi, error) {

	query := `
		SELECT
			id_lokasi,
			alamat,
			latitude,
			longitude
		FROM lokasi
		ORDER BY id_lokasi ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var lokasiList []domain.Lokasi

	for rows.Next() {

		var lokasi domain.Lokasi

		err := rows.Scan(
			&lokasi.IDLokasi,
			&lokasi.Alamat,
			&lokasi.Latitude,
			&lokasi.Longitude,
		)

		if err != nil {
			return nil, err
		}

		lokasiList = append(lokasiList, lokasi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lokasiList, nil
}
