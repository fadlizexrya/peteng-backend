package domain

import "time"

type ProgresLaporan struct {
	IDProgres          string    `json:"id_progres"`
	IDLaporan          string    `json:"id_laporan"`
	IDDinasPerhubungan *int      `json:"id_dinas_perhubungan"`
	Judul              string    `json:"judul"`
	Deskripsi          string    `json:"deskripsi"`
	FotoURL            *string   `json:"foto_url"`
	TanggalProgres     time.Time `json:"tanggal_progres"`
}
