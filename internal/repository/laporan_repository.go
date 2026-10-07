package repository

import (
	"context"

	"peteng-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LaporanRepository struct {
	db *pgxpool.Pool
}

func NewLaporanRepository(db *pgxpool.Pool) *LaporanRepository {
	return &LaporanRepository{
		db: db,
	}
}

func (r *LaporanRepository) Create(
	ctx context.Context,
	laporan *domain.Laporan,
) error {

	query := `
		INSERT INTO laporan (
			id_warga,
			foto,
			latitude,
			longitude,
			keterangan_lokasi,
			kategori,
			tingkat_kegelapan,
			deskripsi,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			'diajukan'
		)
		RETURNING
			id_laporan,
			tanggal_laporan,
			updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		laporan.IDWarga,
		laporan.Foto,
		laporan.Latitude,
		laporan.Longitude,
		laporan.KeteranganLokasi,
		laporan.Kategori,
		laporan.TingkatKegelapan,
		laporan.Deskripsi,
	).Scan(
		&laporan.IDLaporan,
		&laporan.TanggalLaporan,
		&laporan.UpdatedAt,
	)
}
