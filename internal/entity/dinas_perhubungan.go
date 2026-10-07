package entity

import "time"

type DinasPerhubungan struct {
	IDDinasPerhubungan   int       `json:"id_dinas_perhubungan"`
	NamaDinasPerhubungan string    `json:"nama_dinas_perhubungan"`
	Email                string    `json:"email"`
	PasswordEmail        string    `json:"-"` // Disembunyikan saat return JSON
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Request DTO untuk Login & Register
type RegisterDinasRequest struct {
	NamaDinasPerhubungan string `json:"nama_dinas_perhubungan"`
	Email                string `json:"email"`
	PasswordEmail        string `json:"password_email"`
}

type LoginDinasRequest struct {
	Email         string `json:"email"`
	PasswordEmail string `json:"password_email"`
}
