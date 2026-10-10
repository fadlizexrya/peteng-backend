package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"peteng-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ValidasiAIRepository struct {
	db *pgxpool.Pool
}

func NewValidasiAIRepository(db *pgxpool.Pool) *ValidasiAIRepository {
	return &ValidasiAIRepository{db: db}
}

// Memulai validasi dengan mengubah status diajukan menjadi diproses_ai.
func (r *ValidasiAIRepository) MulaiValidasi(
	ctx context.Context,
	idLaporan string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status
		FROM laporan
		WHERE id_laporan = $1
		FOR UPDATE
	`, idLaporan).Scan(&status)
	if err != nil {
		return err
	}

	if status != "diajukan" {
		return fmt.Errorf(
			"laporan harus berstatus diajukan, status sekarang: %s",
			status,
		)
	}

	_, err = tx.Exec(ctx, `
		UPDATE laporan
		SET status = 'diproses_ai',
		    updated_at = NOW()
		WHERE id_laporan = $1
	`, idLaporan)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO progres_laporan (
			id_laporan,
			judul,
			deskripsi
		)
		VALUES ($1, $2, $3)
	`, idLaporan,
		"Validasi AI dimulai",
		"Laporan sedang menunggu proses validasi kecerdasan buatan.",
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Menyimpan hasil mock AI dan mengubah status menjadi diterima/ditolak.
func (r *ValidasiAIRepository) SimpanHasilMock(
	ctx context.Context,
	idLaporan string,
	req *domain.ValidasiAIMockRequest,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status
		FROM laporan
		WHERE id_laporan = $1
		FOR UPDATE
	`, idLaporan).Scan(&status)
	if err != nil {
		return err
	}

	if status != "diproses_ai" {
		return fmt.Errorf(
			"laporan harus berstatus diproses_ai, status sekarang: %s",
			status,
		)
	}

	hasil := req.HasilValidasi
	statusAkhir := "ditolak"
	judulProgres := "Validasi AI ditolak"

	if hasil == "valid" {
		statusAkhir = "diterima"
		judulProgres = "Validasi AI diterima"
	}

	objekJSON, err := json.Marshal([]any{})
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO validasi_ai (
			id_laporan,
			hasil_validasi,
			tingkat_kegelapan_ai,
			objek_terdeteksi,
			catatan,
			model_version
		)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6)
	`,
		idLaporan,
		hasil,
		req.TingkatKegelapanAI,
		string(objekJSON),
		req.Catatan,
		"mock-v1",
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE laporan
		SET status = $1,
		    updated_at = NOW()
		WHERE id_laporan = $2
	`, statusAkhir, idLaporan)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO progres_laporan (
			id_laporan,
			judul,
			deskripsi
		)
		VALUES ($1, $2, $3)
	`,
		idLaporan,
		judulProgres,
		req.Catatan,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Mengambil hasil validasi AI terbaru untuk sebuah laporan.
func (r *ValidasiAIRepository) GetTerbaru(
	ctx context.Context,
	idLaporan string,
) (*domain.ValidasiAI, error) {
	var hasil domain.ValidasiAI
	var objekJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id_validasi,
			id_laporan,
			hasil_validasi,
			tingkat_kegelapan_ai,
			cnn_confidence,
			objek_terdeteksi,
			yolo_confidence,
			catatan,
			model_version,
			tanggal_validasi
		FROM validasi_ai
		WHERE id_laporan = $1
		ORDER BY tanggal_validasi DESC
		LIMIT 1
	`, idLaporan).Scan(
		&hasil.IDValidasi,
		&hasil.IDLaporan,
		&hasil.HasilValidasi,
		&hasil.TingkatKegelapanAI,
		&hasil.CNNConfidence,
		&objekJSON,
		&hasil.YOLOConfidence,
		&hasil.Catatan,
		&hasil.ModelVersion,
		&hasil.TanggalValidasi,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	hasil.ObjekTerdeteksi = json.RawMessage(objekJSON)
	return &hasil, nil
}
