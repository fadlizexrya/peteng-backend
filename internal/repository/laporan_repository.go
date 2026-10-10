package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"peteng-backend/internal/domain"
)

type LaporanRepository struct {
	db *pgxpool.Pool
}

func NewLaporanRepository(db *pgxpool.Pool) *LaporanRepository {
	return &LaporanRepository{
		db: db,
	}
}

// Membuat laporan dan progres awal dalam satu transaksi.
func (r *LaporanRepository) CreateWithProgress(
	ctx context.Context,
	laporan *domain.Laporan,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queryLaporan := `
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'diajukan')
		RETURNING id_laporan, status, tanggal_laporan, updated_at
	`

	err = tx.QueryRow(
		ctx,
		queryLaporan,
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
		&laporan.Status,
		&laporan.TanggalLaporan,
		&laporan.UpdatedAt,
	)
	if err != nil {
		return err
	}

	queryProgres := `
		INSERT INTO progres_laporan (
			id_laporan,
			judul,
			deskripsi
		)
		VALUES ($1, $2, $3)
	`

	_, err = tx.Exec(
		ctx,
		queryProgres,
		laporan.IDLaporan,
		"Laporan berhasil dibuat",
		"Laporan warga berhasil diterima dan menunggu proses validasi AI.",
	)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// Mengambil seluruh laporan milik seorang warga.
func (r *LaporanRepository) GetByWargaID(
	ctx context.Context,
	idWarga int,
) ([]domain.Laporan, error) {
	query := `
		SELECT
			id_laporan,
			id_warga,
			foto,
			latitude,
			longitude,
			keterangan_lokasi,
			kategori,
			tingkat_kegelapan,
			deskripsi,
			status,
			tanggal_laporan,
			updated_at
		FROM laporan
		WHERE id_warga = $1
		ORDER BY tanggal_laporan DESC
	`

	rows, err := r.db.Query(ctx, query, idWarga)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	laporans := make([]domain.Laporan, 0)

	for rows.Next() {
		var laporan domain.Laporan

		err := rows.Scan(
			&laporan.IDLaporan,
			&laporan.IDWarga,
			&laporan.Foto,
			&laporan.Latitude,
			&laporan.Longitude,
			&laporan.KeteranganLokasi,
			&laporan.Kategori,
			&laporan.TingkatKegelapan,
			&laporan.Deskripsi,
			&laporan.Status,
			&laporan.TanggalLaporan,
			&laporan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		laporans = append(laporans, laporan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return laporans, nil
}

func (r *LaporanRepository) GetByIDAndWargaID(
	ctx context.Context,
	idLaporan string,
	idWarga int,
) (*domain.Laporan, error) {
	query := `
		SELECT
			id_laporan,
			id_warga,
			foto,
			latitude,
			longitude,
			keterangan_lokasi,
			kategori,
			tingkat_kegelapan,
			deskripsi,
			status,
			tanggal_laporan,
			updated_at
		FROM laporan
		WHERE id_laporan = $1
		  AND id_warga = $2
	`

	var laporan domain.Laporan

	err := r.db.QueryRow(
		ctx,
		query,
		idLaporan,
		idWarga,
	).Scan(
		&laporan.IDLaporan,
		&laporan.IDWarga,
		&laporan.Foto,
		&laporan.Latitude,
		&laporan.Longitude,
		&laporan.KeteranganLokasi,
		&laporan.Kategori,
		&laporan.TingkatKegelapan,
		&laporan.Deskripsi,
		&laporan.Status,
		&laporan.TanggalLaporan,
		&laporan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &laporan, nil
}

func (r *LaporanRepository) GetAll(ctx context.Context) ([]domain.Laporan, error) {
	query := `
		SELECT
			id_laporan,
			id_warga,
			foto,
			latitude,
			longitude,
			keterangan_lokasi,
			kategori,
			tingkat_kegelapan,
			deskripsi,
			status,
			tanggal_laporan,
			updated_at
		FROM laporan
		ORDER BY tanggal_laporan DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	laporans := make([]domain.Laporan, 0)

	for rows.Next() {
		var laporan domain.Laporan

		err := rows.Scan(
			&laporan.IDLaporan,
			&laporan.IDWarga,
			&laporan.Foto,
			&laporan.Latitude,
			&laporan.Longitude,
			&laporan.KeteranganLokasi,
			&laporan.Kategori,
			&laporan.TingkatKegelapan,
			&laporan.Deskripsi,
			&laporan.Status,
			&laporan.TanggalLaporan,
			&laporan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		laporans = append(laporans, laporan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return laporans, nil
}

func (r *LaporanRepository) UpdateStatusWithProgress(
	ctx context.Context,
	idLaporan string,
	idDishub int,
	status string,
	judul string,
	deskripsi string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Perbarui status laporan.
	queryUpdate := `
		UPDATE laporan
		SET status = $1,
		    updated_at = NOW()
		WHERE id_laporan = $2
		RETURNING id_laporan
	`

	var updatedID string
	err = tx.QueryRow(
		ctx,
		queryUpdate,
		status,
		idLaporan,
	).Scan(&updatedID)
	if err != nil {
		return err
	}

	// Catat perubahan sebagai riwayat progres Dishub.
	queryProgress := `
		INSERT INTO progres_laporan (
			id_laporan,
			id_dinas_perhubungan,
			judul,
			deskripsi
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err = tx.Exec(
		ctx,
		queryProgress,
		updatedID,
		idDishub,
		judul,
		deskripsi,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
