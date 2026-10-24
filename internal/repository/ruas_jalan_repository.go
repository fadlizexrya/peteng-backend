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
	Segmen                     []SegmenPenerangan
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
    segmen_rute AS (
        SELECT
            (d).path AS path,
            (d).geom AS geom
        FROM rute r
        CROSS JOIN LATERAL ST_DumpSegments(
            ST_Segmentize(r.geom, 10.0)
        ) AS d
    ),
    titik_segmen AS (
        SELECT
            path,
            geom,
            ST_Length(geom) AS panjang_meter,
            ST_LineInterpolatePoint(geom, 0.5) AS titik_tengah
        FROM segmen_rute
    ),
    klasifikasi AS (
        SELECT
            s.path,
            s.geom,
            s.panjang_meter,

            COALESCE(
                simulasi.tingkat_penerangan,
                NULL
            ) AS tingkat_penerangan,

            CASE
                WHEN simulasi.sumber_data IS NOT NULL
                    THEN simulasi.sumber_data
                WHEN osm.id IS NOT NULL
                    THEN 'OpenStreetMap'
                ELSE 'tidak_tersedia'
            END AS sumber_data

        FROM titik_segmen s

        -- Prioritas pertama: data penerangan PETENG
        LEFT JOIN LATERAL (
            SELECT
                rj.tingkat_penerangan::text AS tingkat_penerangan,
                rj.sumber_data::text AS sumber_data
            FROM public.ruas_jalan rj
            WHERE ST_DWithin(
                ST_Transform(rj.geom, 32749),
                s.titik_tengah,
                15.0
            )
            ORDER BY ST_Distance(
                ST_Transform(rj.geom, 32749),
                s.titik_tengah
            )
            LIMIT 1
        ) simulasi ON TRUE

        -- Prioritas kedua: geometri jalan OpenStreetMap
        LEFT JOIN LATERAL (
            SELECT
                osm_jalan.id
            FROM public.ruas_jalan_osm osm_jalan
            WHERE simulasi.sumber_data IS NULL
              AND osm_jalan.geom IS NOT NULL
              AND ST_DWithin(
                  ST_Transform(osm_jalan.geom, 32749),
                  s.titik_tengah,
                  15.0
              )
            ORDER BY ST_Distance(
                ST_Transform(osm_jalan.geom, 32749),
                s.titik_tengah
            )
            LIMIT 1
        ) osm ON TRUE
    )
    SELECT
        ST_X(
            ST_Transform(ST_StartPoint(geom), 4326)
        )::float8 AS awal_lon,

        ST_Y(
            ST_Transform(ST_StartPoint(geom), 4326)
        )::float8 AS awal_lat,

        ST_X(
            ST_Transform(ST_EndPoint(geom), 4326)
        )::float8 AS akhir_lon,

        ST_Y(
            ST_Transform(ST_EndPoint(geom), 4326)
        )::float8 AS akhir_lat,

        panjang_meter::float8,

        CASE
            WHEN tingkat_penerangan IN ('terang', 'redup', 'gelap')
                THEN tingkat_penerangan
            ELSE 'tidak_diketahui'
        END AS status,

        sumber_data,

        COALESCE(
            sumber_data = 'simulasi_pengujian',
            FALSE
        ) AS data_simulasi

    FROM klasifikasi
    ORDER BY path
`

	hasil := &HasilPenilaianPenerangan{
		Segmen: make([]SegmenPenerangan, 0),
	}

	rows, err := r.db.Query(ctx, query, string(geometry))
	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil segmen penerangan: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			awalLon, awalLat   float64
			akhirLon, akhirLat float64
			panjang            float64
			status             string
			sumberData         string
			dataSimulasi       bool
		)

		if err := rows.Scan(
			&awalLon,
			&awalLat,
			&akhirLon,
			&akhirLat,
			&panjang,
			&status,
			&sumberData,
			&dataSimulasi,
		); err != nil {
			return nil, fmt.Errorf(
				"gagal membaca segmen penerangan: %w",
				err,
			)
		}

		hasil.Segmen = append(
			hasil.Segmen,
			SegmenPenerangan{
				KoordinatAwal:  [2]float64{awalLon, awalLat},
				KoordinatAkhir: [2]float64{akhirLon, akhirLat},
				PanjangMeter:   panjang,
				Status:         status,
				SumberData:     sumberData,
			},
		)

		hasil.PanjangRuteMeter += panjang

		switch status {
		case "terang":
			hasil.PanjangTerangMeter += panjang
		case "redup":
			hasil.PanjangRedupMeter += panjang
		case "gelap":
			hasil.PanjangGelapMeter += panjang
		default:
			hasil.PanjangTidakDiketahuiMeter += panjang
		}

		if dataSimulasi {
			hasil.AdaDataSimulasi = true
		}

		hasil.JumlahSegmen++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"gagal membaca hasil segmen penerangan: %w",
			err,
		)
	}

	return hasil, nil
}

type SegmenPenerangan struct {
	KoordinatAwal  [2]float64 `json:"koordinat_awal"`
	KoordinatAkhir [2]float64 `json:"koordinat_akhir"`
	PanjangMeter   float64    `json:"panjang_meter"`
	Status         string     `json:"status"`
	SumberData     string     `json:"sumber_data"`
}

// TitikDalamWilayah memeriksa apakah titik berada di dalam
// atau tepat pada batas Kabupaten Indramayu.
func (r *RuasJalanRepository) TitikDalamWilayah(
	ctx context.Context,
	latitude, longitude float64,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM public.batas_wilayah
			WHERE tingkat_wilayah = 'kabupaten'
			  AND LOWER(TRIM(nama_wilayah)) = 'kabupaten indramayu'
			  AND ST_Covers(
				  geom,
				  ST_SetSRID(
					  ST_MakePoint($1, $2),
					  4326
				  )
			  )
		)
	`

	var dalamWilayah bool

	err := r.db.QueryRow(
		ctx,
		query,
		longitude,
		latitude,
	).Scan(&dalamWilayah)

	if err != nil {
		return false, fmt.Errorf(
			"gagal memeriksa titik terhadap batas wilayah: %w",
			err,
		)
	}

	return dalamWilayah, nil
}

// RuteDalamWilayah memeriksa apakah seluruh geometri rute
// berada di dalam atau pada batas Kabupaten Indramayu.
func (r *RuasJalanRepository) RuteDalamWilayah(
	ctx context.Context,
	geometry json.RawMessage,
) (bool, error) {
	if !json.Valid(geometry) {
		return false, fmt.Errorf(
			"geometri rute bukan JSON yang valid",
		)
	}

	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM public.batas_wilayah
			WHERE tingkat_wilayah = 'kabupaten'
			  AND LOWER(TRIM(nama_wilayah)) = 'kabupaten indramayu'
			  AND ST_Covers(
				  geom,
				  ST_SetSRID(
					  ST_GeomFromGeoJSON($1),
					  4326
				  )
			  )
		)
	`

	var dalamWilayah bool

	err := r.db.QueryRow(
		ctx,
		query,
		string(geometry),
	).Scan(&dalamWilayah)

	if err != nil {
		return false, fmt.Errorf(
			"gagal memeriksa geometri rute terhadap batas wilayah: %w",
			err,
		)
	}

	return dalamWilayah, nil
}
