package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"peteng-backend/internal/domain"
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

func (r *ProgresLaporanRepository) GetByLaporanID(
	ctx context.Context,
	idLaporan string,
) ([]domain.ProgresLaporan, error) {
	query := `
		SELECT
			id_progres,
			id_laporan,
			id_dinas_perhubungan,
			judul,
			deskripsi,
			foto_url,
			tanggal_progres
		FROM progres_laporan
		WHERE id_laporan = $1
		ORDER BY tanggal_progres ASC
	`

	rows, err := r.db.Query(ctx, query, idLaporan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progresList := make([]domain.ProgresLaporan, 0)

	for rows.Next() {
		var progres domain.ProgresLaporan

		err := rows.Scan(
			&progres.IDProgres,
			&progres.IDLaporan,
			&progres.IDDinasPerhubungan,
			&progres.Judul,
			&progres.Deskripsi,
			&progres.FotoURL,
			&progres.TanggalProgres,
		)
		if err != nil {
			return nil, err
		}

		progresList = append(progresList, progres)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return progresList, nil
}
