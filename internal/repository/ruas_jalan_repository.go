package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HasilPenilaianPenerangan struct {
	PanjangRuteMeter           float64
	PanjangTerangMeter         float64
	PanjangRedupMeter          float64
	PanjangGelapMeter          float64
	PanjangTidakDiketahuiMeter float64
	JumlahSegmen               int
	AdaDataSimulasi            bool
}

type RuasJalanRepository struct {
	db *pgxpool.Pool
}

func NewRuasJalanRepository(db *pgxpool.Pool) *RuasJalanRepository {
	return &RuasJalanRepository{db: db}
}

func (r *RuasJalanRepository) HitungPenerangan(
	ctx context.Context,
	geometry json.RawMessage,
) (*HasilPenilaianPenerangan, error) {

	if !json.Valid(geometry) {
		return nil, fmt.Errorf("geometri rute bukan JSON yang valid")
	}

	const query = `
		WITH rute AS (
			SELECT ST_Transform(
				ST_SetSRID(
					ST_GeomFromGeoJSON($1),
					4326
				),
				32749
			) AS geom
		),
		ruas AS (
			SELECT
				id_ruas,
				tingkat_penerangan,
				sumber_data,
				ST_Transform(geom, 32749) AS geom
			FROM public.ruas_jalan
		),
		segmen_rute AS (
			SELECT
				(d).geom AS geom
			FROM rute r
			CROSS JOIN LATERAL ST_DumpSegments(
				ST_Segmentize(r.geom, 10.0)
			) AS d
		),
		klasifikasi AS (
			SELECT
				s.geom,
				ST_Length(s.geom) AS panjang_meter,
				j.tingkat_penerangan,
				j.sumber_data
			FROM segmen_rute s
			LEFT JOIN LATERAL (
				SELECT
					rj.tingkat_penerangan,
					rj.sumber_data
				FROM ruas rj
				WHERE ST_DWithin(
					rj.geom,
					ST_LineInterpolatePoint(s.geom, 0.5),
					15.0
				)
				ORDER BY ST_Distance(
					rj.geom,
					ST_LineInterpolatePoint(s.geom, 0.5)
				)
				LIMIT 1
			) j ON TRUE
		)
		SELECT
			COALESCE(SUM(panjang_meter), 0)::float8,
			COALESCE(SUM(panjang_meter) FILTER (
				WHERE tingkat_penerangan = 'terang'
			), 0)::float8,
			COALESCE(SUM(panjang_meter) FILTER (
				WHERE tingkat_penerangan = 'redup'
			), 0)::float8,
			COALESCE(SUM(panjang_meter) FILTER (
				WHERE tingkat_penerangan = 'gelap'
			), 0)::float8,
			COALESCE(SUM(panjang_meter) FILTER (
				WHERE tingkat_penerangan = 'belum_diketahui'
				   OR tingkat_penerangan IS NULL
			), 0)::float8,
			COUNT(*)::int,
			COALESCE(BOOL_OR(
				sumber_data = 'simulasi_pengujian'
			), FALSE)
		FROM klasifikasi
	`

	hasil := &HasilPenilaianPenerangan{}

	err := r.db.QueryRow(
		ctx,
		query,
		string(geometry),
	).Scan(
		&hasil.PanjangRuteMeter,
		&hasil.PanjangTerangMeter,
		&hasil.PanjangRedupMeter,
		&hasil.PanjangGelapMeter,
		&hasil.PanjangTidakDiketahuiMeter,
		&hasil.JumlahSegmen,
		&hasil.AdaDataSimulasi,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal menghitung panjang penerangan rute: %w",
			err,
		)
	}

	return hasil, nil
}
