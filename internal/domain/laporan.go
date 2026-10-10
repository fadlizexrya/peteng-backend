package domain

import "time"

type Laporan struct {
	IDLaporan        string    `json:"id_laporan"`
	IDWarga          int       `json:"id_warga"`
	Foto             string    `json:"foto"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	KeteranganLokasi string    `json:"keterangan_lokasi"`
	Kategori         string    `json:"kategori"`
	TingkatKegelapan int16     `json:"tingkat_kegelapan"`
	Deskripsi        string    `json:"deskripsi"`
	Status           string    `json:"status"`
	TanggalLaporan   time.Time `json:"tanggal_laporan"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type LaporanReq struct {
	Foto             string  `json:"foto"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	KeteranganLokasi string  `json:"keterangan_lokasi"`
	Kategori         string  `json:"kategori"`
	TingkatKegelapan int16   `json:"tingkat_kegelapan"`
	Deskripsi        string  `json:"deskripsi"`
}

type LaporanDetail struct {
	Laporan Laporan          `json:"laporan"`
	Progres []ProgresLaporan `json:"progres"`
}

type UpdateStatusLaporanReq struct {
	Status    string `json:"status"`
	Judul     string `json:"judul"`
	Deskripsi string `json:"deskripsi"`
}
