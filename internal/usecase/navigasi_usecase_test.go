package usecase

import "testing"

func floatPtr(v float64) *float64 {
	return &v
}

func TestHitungPeringkatRute_KombinasiSkor(t *testing.T) {
	hasil := &HasilNavigasi{
		Rute: []RuteNavigasi{
			{
				NomorRute:   1,
				JarakMeter:  2000,
				DurasiDetik: 300,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(80),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
			{
				NomorRute:   2,
				JarakMeter:  2500,
				DurasiDetik: 400,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(80),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
			{
				NomorRute:   3,
				JarakMeter:  3500,
				DurasiDetik: 500,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(95),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
		},
	}

	hitungPeringkatRute(hasil)

	if len(hasil.Rute) != 3 {
		t.Fatalf("diharapkan 3 rute, didapat %d", len(hasil.Rute))
	}

	// Rute 1 seharusnya menang karena skor gabungannya tertinggi.
	ruteTeratas := hasil.Rute[0]

	if ruteTeratas.NomorRute != 1 {
		t.Errorf(
			"diharapkan rute 1 menjadi peringkat pertama, didapat rute %d",
			ruteTeratas.NomorRute,
		)
	}

	if ruteTeratas.SkorAkhir == nil {
		t.Fatal("skor akhir rute pertama tidak boleh nil")
	}

	if *ruteTeratas.SkorAkhir != 88 {
		t.Errorf(
			"diharapkan skor akhir 88, didapat %.2f",
			*ruteTeratas.SkorAkhir,
		)
	}

	if !ruteTeratas.Direkomendasikan {
		t.Error("rute pertama seharusnya direkomendasikan")
	}

	if hasil.NomorRuteRekomendasi == nil ||
		*hasil.NomorRuteRekomendasi != 1 {
		t.Error("nomor rute rekomendasi seharusnya 1")
	}

	if hasil.StatusRekomendasi != "rekomendasi_simulasi" {
		t.Errorf(
			"diharapkan status rekomendasi_simulasi, didapat %s",
			hasil.StatusRekomendasi,
		)
	}
}

func TestHitungPeringkatRute_CakupanRendahTidakDirekomendasikan(
	t *testing.T,
) {
	hasil := &HasilNavigasi{
		Rute: []RuteNavigasi{
			{
				NomorRute:   1,
				JarakMeter:  1000,
				DurasiDetik: 100,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(100),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_rendah",
				},
			},
		},
	}

	hitungPeringkatRute(hasil)

	if hasil.Rute[0].SkorAkhir != nil {
		t.Error("rute dengan cakupan data rendah tidak boleh mendapat skor akhir")
	}

	if hasil.Rute[0].Direkomendasikan {
		t.Error("rute dengan cakupan data rendah tidak boleh direkomendasikan")
	}

	if hasil.NomorRuteRekomendasi != nil {
		t.Error("seharusnya tidak ada nomor rute rekomendasi")
	}

	if hasil.StatusRekomendasi != "belum_ada_rekomendasi" {
		t.Errorf(
			"diharapkan belum_ada_rekomendasi, didapat %s",
			hasil.StatusRekomendasi,
		)
	}
}

func TestHitungPeringkatRute_TanpaDataPenerangan(
	t *testing.T,
) {
	hasil := &HasilNavigasi{
		Rute: []RuteNavigasi{
			{
				NomorRute:   1,
				JarakMeter:  1000,
				DurasiDetik: 100,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  nil,
					StatusData:      "belum_terverifikasi",
					StatusPenilaian: "belum_ada_data_penerangan",
				},
			},
		},
	}

	hitungPeringkatRute(hasil)

	if hasil.Rute[0].SkorAkhir != nil {
		t.Error("skor akhir harus nil jika tidak ada data penerangan")
	}

	if hasil.Rute[0].Direkomendasikan {
		t.Error("rute tanpa data penerangan tidak boleh direkomendasikan")
	}

	if hasil.NomorRuteRekomendasi != nil {
		t.Error("seharusnya tidak ada rekomendasi")
	}
}

func TestHitungPeringkatRute_MelewatiRuteDenganJarakTidakValid(t *testing.T) {
	hasil := &HasilNavigasi{
		Rute: []RuteNavigasi{
			{
				NomorRute:   1,
				JarakMeter:  0,
				DurasiDetik: 100,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(100),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
			{
				NomorRute:   2,
				JarakMeter:  1500,
				DurasiDetik: 120,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(80),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
		},
	}

	hitungPeringkatRute(hasil)

	if hasil.NomorRuteRekomendasi == nil {
		t.Fatal("seharusnya ada rekomendasi untuk rute yang valid")
	}

	if *hasil.NomorRuteRekomendasi != 2 {
		t.Errorf(
			"diharapkan rute 2 direkomendasikan, didapat rute %d",
			*hasil.NomorRuteRekomendasi,
		)
	}

	for _, rute := range hasil.Rute {
		if rute.NomorRute == 1 {
			if rute.SkorAkhir != nil {
				t.Error("rute dengan jarak nol tidak boleh memiliki skor akhir")
			}
			if rute.Direkomendasikan {
				t.Error("rute dengan jarak nol tidak boleh direkomendasikan")
			}
		}
	}
}

func TestHitungPeringkatRute_SemuaRuteTidakValid(t *testing.T) {
	hasil := &HasilNavigasi{
		Rute: []RuteNavigasi{
			{
				NomorRute:   1,
				JarakMeter:  0,
				DurasiDetik: 100,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(90),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
			{
				NomorRute:   2,
				JarakMeter:  1000,
				DurasiDetik: 0,
				Penerangan: PenilaianPenerangan{
					SkorPenerangan:  floatPtr(80),
					StatusData:      "simulasi",
					StatusPenilaian: "cakupan_data_memadai",
				},
			},
		},
	}

	hitungPeringkatRute(hasil)

	if hasil.NomorRuteRekomendasi != nil {
		t.Error("tidak boleh ada rekomendasi jika semua rute tidak valid")
	}

	if hasil.StatusRekomendasi != "belum_ada_rekomendasi" {
		t.Errorf(
			"diharapkan status belum_ada_rekomendasi, didapat %q",
			hasil.StatusRekomendasi,
		)
	}

	for _, rute := range hasil.Rute {
		if rute.SkorAkhir != nil {
			t.Errorf("rute %d tidak boleh memiliki skor akhir", rute.NomorRute)
		}
		if rute.Direkomendasikan {
			t.Errorf("rute %d tidak boleh direkomendasikan", rute.NomorRute)
		}
	}
}
