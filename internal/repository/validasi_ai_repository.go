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

	if status != "diajukan" && status != "gagal_validasi" {
		return fmt.Errorf(
			"laporan harus berstatus diajukan atau gagal_validasi, status sekarang: %s",
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

// StatusLaporan mengembalikan status terkini untuk memvalidasi permintaan retry.
func (r *ValidasiAIRepository) StatusLaporan(ctx context.Context, idLaporan string) (string, error) {
	var status string
	err := r.db.QueryRow(ctx, `SELECT status FROM laporan WHERE id_laporan = $1`, idLaporan).Scan(&status)
	return status, err
}

// TandaiGagal menyimpan kondisi gagal yang bisa ditampilkan dan dicoba ulang.
func (r *ValidasiAIRepository) TandaiGagal(ctx context.Context, idLaporan, alasan string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE laporan
		SET status = 'gagal_validasi', updated_at = NOW()
		WHERE id_laporan = $1 AND status IN ('diajukan', 'diproses_ai', 'gagal_validasi')
	`, idLaporan)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO progres_laporan (id_laporan, judul, deskripsi)
		VALUES ($1, 'Validasi AI gagal', $2)
	`, idLaporan, alasan)
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

// Menyimpan hasil AI nyata dan mengubah status laporan ke diterima/ditolak.
func (r *ValidasiAIRepository) SimpanHasilAI(
	ctx context.Context,
	idLaporan string,
	hasil *domain.ValidasiAIResponse,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM laporan WHERE id_laporan = $1 FOR UPDATE`, idLaporan).Scan(&status); err != nil {
		return err
	}
	if status != "diproses_ai" {
		return fmt.Errorf("laporan harus berstatus diproses_ai, status sekarang: %s", status)
	}

	if hasil == nil {
		return errors.New("hasil validasi AI kosong")
	}
	statusAkhir := "ditolak"
	judulProgres := "Validasi AI ditolak"
	if hasil.Data.HasilValidasi == "valid" {
		statusAkhir = "diterima"
		judulProgres = "Validasi AI diterima"
	}
	if hasil.Data.HasilValidasi != "valid" && hasil.Data.HasilValidasi != "tidak_valid" {
		return errors.New("hasil validasi AI tidak valid")
	}
	if hasil.Data.HasilValidasi == "valid" && (hasil.Data.TingkatKegelapanAI == nil || *hasil.Data.TingkatKegelapanAI < 1 || *hasil.Data.TingkatKegelapanAI > 3) {
		return errors.New("hasil valid memerlukan tingkat kegelapan 1-3")
	}
	if hasil.Data.HasilValidasi == "tidak_valid" && hasil.Data.TingkatKegelapanAI != nil {
		return errors.New("hasil tidak_valid harus tanpa tingkat kegelapan")
	}

	var responseJSON any
	if len(hasil.Data.HasilJSON) > 0 {
		responseJSON = string(hasil.Data.HasilJSON)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO riwayat_hasil_ai (
			id_laporan, request_id, ai_status, photo_status,
			lighting_class, confidence, model_version, pesan, response_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
	`, idLaporan, hasil.Data.RequestID, hasil.Data.AIStatus,
		stringOrNil(hasil.Data.PhotoStatus), stringOrNil(hasil.Data.LightingClass),
		hasil.Data.Confidence, hasil.Data.ModelVersion, hasil.Data.Message, responseJSON)
	if err != nil {
		return err
	}

	objekJSON, err := json.Marshal(hasil.Data.ObjekTerdeteksi)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO validasi_ai (
			id_laporan,
			hasil_validasi,
			tingkat_kegelapan_ai,
			cnn_confidence,
			objek_terdeteksi,
			catatan,
			model_version
		)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
	`,
		idLaporan,
		hasil.Data.HasilValidasi,
		hasil.Data.TingkatKegelapanAI,
		hasil.Data.CNNConfidence,
		string(objekJSON),
		hasil.Data.Catatan,
		hasil.Data.ModelVersion,
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
		hasil.Data.Catatan,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Mengambil hasil validasi AI terbaru untuk sebuah laporan.
func stringOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func (r *ValidasiAIRepository) GetTerbaru(
	ctx context.Context,
	idLaporan string,
) (*domain.ValidasiAI, error) {
	var hasil domain.ValidasiAI
	var objekJSON []byte
	var analisis domain.RiwayatHasilAI
	var requestID, aiStatus, modelVersion *string

	err := r.db.QueryRow(ctx, `
		SELECT
			v.id_validasi,
			v.id_laporan,
			v.hasil_validasi,
			v.tingkat_kegelapan_ai,
			v.cnn_confidence,
			v.objek_terdeteksi,
			v.catatan,
			v.model_version,
			v.tanggal_validasi,
			a.request_id,
			a.ai_status,
			a.photo_status,
			a.lighting_class,
			a.confidence,
			a.model_version,
			a.pesan,
			a.response_json,
			a.tanggal_proses
		FROM validasi_ai v
		LEFT JOIN LATERAL (
			SELECT request_id, ai_status, photo_status, lighting_class, confidence,
			       model_version, pesan, response_json, tanggal_proses
			FROM riwayat_hasil_ai
			WHERE id_laporan = v.id_laporan
			ORDER BY tanggal_proses DESC
			LIMIT 1
		) a ON TRUE
		WHERE v.id_laporan = $1
		ORDER BY v.tanggal_validasi DESC
		LIMIT 1
	`, idLaporan).Scan(
		&hasil.IDValidasi,
		&hasil.IDLaporan,
		&hasil.HasilValidasi,
		&hasil.TingkatKegelapanAI,
		&hasil.CNNConfidence,
		&objekJSON,
		&hasil.Catatan,
		&hasil.ModelVersion,
		&hasil.TanggalValidasi,
		&requestID,
		&aiStatus,
		&analisis.PhotoStatus,
		&analisis.LightingClass,
		&analisis.Confidence,
		&modelVersion,
		&analisis.Pesan,
		&analisis.ResponseJSON,
		&analisis.TanggalProses,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	hasil.ObjekTerdeteksi = json.RawMessage(objekJSON)
	if requestID != nil && aiStatus != nil {
		analisis.RequestID = *requestID
		analisis.AIStatus = *aiStatus
		if modelVersion != nil {
			analisis.ModelVersion = modelVersion
		}
		hasil.AnalisisAI = &analisis
	}
	return &hasil, nil
}
