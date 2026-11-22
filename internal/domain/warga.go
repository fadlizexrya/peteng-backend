package domain

import "time"

type Warga struct {
	IDWarga       int       `json:"id_warga"`
	Nama          string    `json:"nama"`
	Email         string    `json:"email"`
	PasswordEmail string    `json:"password_email,omitempty"`
	NomorTelepon  string    `json:"nomor_telepon"`
	EmailVerified bool      `json:"email_verified"`
	GoogleID      *string   `json:"google_id,omitempty"`
	AuthProvider  string    `json:"auth_provider"`
	CreatedAt     time.Time `json:"created_at"`

	// Data internal verifikasi email. Tidak pernah dikirim ke Flutter.
	VerificationToken     string     `json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`
	VerificationAttempts  int        `json:"-"`
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

type VerifyEmailReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResendVerificationReq struct {
	Email string `json:"email"`
}

type GoogleAuthReq struct {
	IDToken  string `json:"id_token"`
	Password string `json:"password"`
}
