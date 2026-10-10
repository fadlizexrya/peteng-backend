package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"peteng-backend/internal/repository"
)

const osrmBaseURL = "https://router.project-osrm.org"

// Bobot penilaian rute PETENG.
const (
	BobotPenerangan = 0.60
	BobotJarak      = 0.25
	BobotDurasi     = 0.15

	AmbangCakupanData = 0.60
)

// NavigasiUsecase menangani pencarian dan penilaian rute.
type NavigasiUsecase struct {
	client        *http.Client
	ruasJalanRepo *repository.RuasJalanRepository
}

func NewNavigasiUsecase(
	ruasJalanRepo *repository.RuasJalanRepository,
) *NavigasiUsecase {
	return &NavigasiUsecase{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
		ruasJalanRepo: ruasJalanRepo,
	}
}

// Struktur respons dari OSRM.
type RuteOSRM struct {
	Distance float64         `json:"distance"`
	Duration float64         `json:"duration"`
	Geometry json.RawMessage `json:"geometry"`
}

type ResponsOSRM struct {
	Code    string     `json:"code"`
	Message string     `json:"message,omitempty"`
	Routes  []RuteOSRM `json:"routes"`
}

// Hasil penilaian kondisi penerangan.
type PenilaianPenerangan struct {
	PanjangRuteDianalisisMeter float64 `json:"panjang_rute_dianalisis_meter"`
	PanjangTerangMeter         float64 `json:"panjang_terang_meter"`
	PanjangRedupMeter          float64 `json:"panjang_redup_meter"`
	PanjangGelapMeter          float64 `json:"panjang_gelap_meter"`
	PanjangTidakDiketahuiMeter float64 `json:"panjang_tidak_diketahui_meter"`

	PersentaseTerang         float64 `json:"persentase_terang"`
	PersentaseRedup          float64 `json:"persentase_redup"`
	PersentaseGelap          float64 `json:"persentase_gelap"`
	PersentaseDiketahui      float64 `json:"persentase_diketahui"`
	PersentaseTidakDiketahui float64 `json:"persentase_tidak_diketahui"`

	SkorPenerangan  *float64 `json:"skor_penerangan"`
	StatusData      string   `json:"status_data"`
	StatusPenilaian string   `json:"status_penilaian"`
}

// RuteNavigasi menyimpan informasi dan skor setiap alternatif.
type RuteNavigasi struct {
	NomorRute        int                 `json:"nomor_rute"`
	JarakMeter       float64             `json:"jarak_meter"`
	DurasiDetik      float64             `json:"durasi_detik"`
	Geometry         json.RawMessage     `json:"geometry"`
	Penerangan       PenilaianPenerangan `json:"penerangan"`
	SkorJarak        float64             `json:"skor_jarak"`
	SkorDurasi       float64             `json:"skor_durasi"`
	SkorAkhir        *float64            `json:"skor_akhir"`
	Peringkat        int                 `json:"peringkat,omitempty"`
	Direkomendasikan bool                `json:"direkomendasikan"`
}

// HasilNavigasi adalah respons utama pencarian rute.
type HasilNavigasi struct {
	Rute                 []RuteNavigasi `json:"rute"`
	NomorRuteRekomendasi *int           `json:"nomor_rute_rekomendasi"`
	StatusRekomendasi    string         `json:"status_rekomendasi"`
}

// Menghitung persentase panjang terhadap total panjang rute.
func persentasePanjang(panjang, total float64) float64 {
	if total <= 0 {
		return 0
	}

	return math.Round(panjang/total*10000) / 100
}

// CariRute mencari alternatif rute melalui OSRM,
// kemudian menilai penerangan dan menghitung peringkatnya.
func (u *NavigasiUsecase) CariRute(
	ctx context.Context,
	startLat, startLon, endLat, endLon float64,
) (*HasilNavigasi, error) {

	if startLat < -90 || startLat > 90 ||
		endLat < -90 || endLat > 90 {
		return nil, fmt.Errorf(
			"latitude harus berada antara -90 dan 90",
		)
	}

	if startLon < -180 || startLon > 180 ||
		endLon < -180 || endLon > 180 {
		return nil, fmt.Errorf(
			"longitude harus berada antara -180 dan 180",
		)
	}

	if u.ruasJalanRepo == nil {
		return nil, fmt.Errorf(
			"repository ruas jalan belum dikonfigurasi",
		)
	}

	// OSRM menggunakan urutan longitude,latitude.
	koordinat := fmt.Sprintf(
		"%f,%f;%f,%f",
		startLon, startLat, endLon, endLat,
	)

	endpoint := fmt.Sprintf(
		"%s/route/v1/driving/%s",
		osrmBaseURL,
		url.PathEscape(koordinat),
	)

	params := url.Values{}
	params.Set("alternatives", "3")
	params.Set("overview", "full")
	params.Set("geometries", "geojson")
	params.Set("steps", "false")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membuat request OSRM: %w", err,
		)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal menghubungi layanan routing: %w", err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"layanan routing mengembalikan HTTP %d",
			resp.StatusCode,
		)
	}

	var hasilOSRM ResponsOSRM

	if err := json.NewDecoder(resp.Body).Decode(&hasilOSRM); err != nil {
		return nil, fmt.Errorf(
			"gagal membaca respons OSRM: %w", err,
		)
	}

	if hasilOSRM.Code != "Ok" {
		if hasilOSRM.Message != "" {
			return nil, fmt.Errorf("OSRM: %s", hasilOSRM.Message)
		}

		return nil, fmt.Errorf(
			"OSRM tidak menemukan rute: %s",
			hasilOSRM.Code,
		)
	}

	if len(hasilOSRM.Routes) == 0 {
		return nil, fmt.Errorf("OSRM tidak mengembalikan alternatif rute")
	}

	hasil := &HasilNavigasi{
		Rute:              make([]RuteNavigasi, 0, len(hasilOSRM.Routes)),
		StatusRekomendasi: "belum_ada_rekomendasi",
	}

	// Nilai setiap alternatif berdasarkan data ruas jalan.
	for i, rute := range hasilOSRM.Routes {
		penilaianDB, err := u.ruasJalanRepo.HitungPenerangan(
			ctx,
			rute.Geometry,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"gagal menilai penerangan rute %d: %w",
				i+1,
				err,
			)
		}

		panjangRute := penilaianDB.PanjangRuteMeter
		terang := penilaianDB.PanjangTerangMeter
		redup := penilaianDB.PanjangRedupMeter
		gelap := penilaianDB.PanjangGelapMeter
		tidakDiketahui := penilaianDB.PanjangTidakDiketahuiMeter

		diketahui := terang + redup + gelap

		penilaian := PenilaianPenerangan{
			PanjangRuteDianalisisMeter: math.Round(panjangRute*100) / 100,
			PanjangTerangMeter:         math.Round(terang*100) / 100,
			PanjangRedupMeter:          math.Round(redup*100) / 100,
			PanjangGelapMeter:          math.Round(gelap*100) / 100,
			PanjangTidakDiketahuiMeter: math.Round(tidakDiketahui*100) / 100,

			PersentaseTerang:         persentasePanjang(terang, panjangRute),
			PersentaseRedup:          persentasePanjang(redup, panjangRute),
			PersentaseGelap:          persentasePanjang(gelap, panjangRute),
			PersentaseDiketahui:      persentasePanjang(diketahui, panjangRute),
			PersentaseTidakDiketahui: persentasePanjang(tidakDiketahui, panjangRute),

			StatusData:      "belum_terverifikasi",
			StatusPenilaian: "data_belum_cukup",
		}

		if penilaianDB.AdaDataSimulasi {
			penilaian.StatusData = "simulasi"
		}

		// Skor penerangan menggunakan data yang memiliki kategori.
		// Terang = 100, redup = 60, gelap = 0.
		if diketahui > 0 {
			skor := math.Round(
				(terang*100+redup*60)/diketahui*100,
			) / 100

			penilaian.SkorPenerangan = &skor

			if panjangRute > 0 &&
				diketahui/panjangRute < AmbangCakupanData {
				penilaian.StatusPenilaian = "cakupan_data_rendah"
			} else {
				penilaian.StatusPenilaian = "cakupan_data_memadai"
			}
		} else if panjangRute > 0 {
			penilaian.StatusPenilaian = "belum_ada_data_penerangan"
		}

		hasil.Rute = append(hasil.Rute, RuteNavigasi{
			NomorRute:   i + 1,
			JarakMeter:  rute.Distance,
			DurasiDetik: rute.Duration,
			Geometry:    rute.Geometry,
			Penerangan:  penilaian,
		})
	}

	hitungPeringkatRute(hasil)

	return hasil, nil
}

// skorRelatif mengubah jarak atau durasi menjadi skor 0-100.
// Nilai minimum mendapat skor 100.
func skorRelatif(nilaiMinimum, nilaiRute float64) float64 {
	if nilaiRute <= 0 || nilaiMinimum <= 0 {
		return 0
	}

	skor := (nilaiMinimum / nilaiRute) * 100

	return math.Round(skor*100) / 100
}

// hitungPeringkatRute menghitung skor gabungan dan memilih rekomendasi.
func hitungPeringkatRute(hasil *HasilNavigasi) {
	if hasil == nil {
		return
	}

	// Reset hasil rekomendasi agar tidak menyisakan nilai lama.
	hasil.NomorRuteRekomendasi = nil
	hasil.StatusRekomendasi = "belum_ada_rekomendasi"

	if len(hasil.Rute) == 0 {
		return
	}

	// Rute layak harus memiliki skor penerangan dan cakupan data memadai.
	layakDirekomendasikan := func(rute RuteNavigasi) bool {
		return rute.Penerangan.SkorPenerangan != nil &&
			rute.Penerangan.StatusPenilaian == "cakupan_data_memadai"
	}

	// Reset skor dan status rekomendasi setiap rute.
	for i := range hasil.Rute {
		hasil.Rute[i].SkorJarak = 0
		hasil.Rute[i].SkorDurasi = 0
		hasil.Rute[i].SkorAkhir = nil
		hasil.Rute[i].Peringkat = 0
		hasil.Rute[i].Direkomendasikan = false
	}

	// Cari jarak dan durasi minimum hanya dari rute yang layak.
	var jarakMinimum, durasiMinimum float64
	adaRuteLayak := false

	for _, rute := range hasil.Rute {
		if !layakDirekomendasikan(rute) {
			continue
		}

		if rute.JarakMeter <= 0 || rute.DurasiDetik <= 0 {
			continue
		}

		if !adaRuteLayak ||
			rute.JarakMeter < jarakMinimum {
			jarakMinimum = rute.JarakMeter
		}

		if !adaRuteLayak ||
			rute.DurasiDetik < durasiMinimum {
			durasiMinimum = rute.DurasiDetik
		}

		adaRuteLayak = true
	}

	if !adaRuteLayak {
		return
	}

	// Hitung skor hanya untuk rute yang layak.
	for i := range hasil.Rute {
		rute := &hasil.Rute[i]

		if !layakDirekomendasikan(*rute) ||
			rute.JarakMeter <= 0 ||
			rute.DurasiDetik <= 0 {
			continue
		}

		rute.SkorJarak = skorRelatif(
			jarakMinimum,
			rute.JarakMeter,
		)

		rute.SkorDurasi = skorRelatif(
			durasiMinimum,
			rute.DurasiDetik,
		)

		skorPenerangan := *rute.Penerangan.SkorPenerangan

		skorAkhir := BobotPenerangan*skorPenerangan +
			BobotJarak*rute.SkorJarak +
			BobotDurasi*rute.SkorDurasi

		skorAkhir = math.Round(skorAkhir*100) / 100
		rute.SkorAkhir = &skorAkhir
	}

	// Rute layak diurutkan terlebih dahulu berdasarkan skor akhir.
	// Jika skor sama, gunakan skor penerangan, jarak, lalu durasi.
	sort.SliceStable(hasil.Rute, func(i, j int) bool {
		a := hasil.Rute[i]
		b := hasil.Rute[j]

		aLayak := a.SkorAkhir != nil
		bLayak := b.SkorAkhir != nil

		if aLayak != bLayak {
			return aLayak
		}

		if !aLayak {
			return false
		}

		if *a.SkorAkhir != *b.SkorAkhir {
			return *a.SkorAkhir > *b.SkorAkhir
		}

		skorA := *a.Penerangan.SkorPenerangan
		skorB := *b.Penerangan.SkorPenerangan

		if skorA != skorB {
			return skorA > skorB
		}

		if a.JarakMeter != b.JarakMeter {
			return a.JarakMeter < b.JarakMeter
		}

		return a.DurasiDetik < b.DurasiDetik
	})

	// Beri peringkat hanya kepada rute yang mempunyai skor akhir.
	peringkat := 0

	for i := range hasil.Rute {
		rute := &hasil.Rute[i]

		if rute.SkorAkhir == nil {
			continue
		}

		peringkat++
		rute.Peringkat = peringkat

		if peringkat == 1 {
			rute.Direkomendasikan = true

			nomorRute := rute.NomorRute
			hasil.NomorRuteRekomendasi = &nomorRute

			if rute.Penerangan.StatusData == "simulasi" {
				hasil.StatusRekomendasi = "rekomendasi_simulasi"
			} else {
				hasil.StatusRekomendasi = "rekomendasi_berdasarkan_data"
			}
		}
	}
}
