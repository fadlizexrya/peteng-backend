package domain

import "time"

type Warga struct {
	IDWarga       int       `json:"id_warga"`
	Nama          string    `json:"nama"`
	Email         string    `json:"email"`
	PasswordEmail string    `json:"password_email,omitempty"`
	NomorTelepon  string    `json:"nomor_telepon"`
	CreatedAt     time.Time `json:"created_at"`
}

type RegisterReq struct {
	Nama         string `json:"nama"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	NomorTelepon string `json:"nomor_telepon"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LaporanReq struct {
	IDWarga   int     `json:"id_warga"`
	Deskripsi string  `json:"deskripsi"`
	FotoURL   string  `json:"foto_url"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
