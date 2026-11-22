package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/usecase"
)

type WargaHandler struct {
	usecase *usecase.WargaUsecase
}

func NewWargaHandler(u *usecase.WargaUsecase) *WargaHandler {
	return &WargaHandler{usecase: u}
}

func (h *WargaHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format_salah", "Format request salah")
		return
	}

	res, err := h.usecase.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "data_tidak_valid", "Nama, email, dan password minimal 8 karakter wajib diisi.")
		case errors.Is(err, usecase.ErrEmailRegistered):
			writeError(w, http.StatusConflict, "email_terdaftar", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "server_error", "Registrasi gagal. Coba lagi.")
		}
		return
	}

	payload := map[string]interface{}{
		"status":     "sukses",
		"message":    res.Message,
		"email_sent": res.EmailSent,
	}

	if res.DebugCode != "" {
		payload["debug_code"] = res.DebugCode
	}

	writeJSON(w, http.StatusCreated, payload)
}

func (h *WargaHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req domain.VerifyEmailReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format_salah", "Format request salah")
		return
	}

	warga, err := h.usecase.VerifyEmail(r.Context(), &req)
	if err != nil {
		writeVerifyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "sukses",
		"message": "Email berhasil diverifikasi. Silakan login.",
		"user":    warga,
	})
}

func (h *WargaHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req domain.ResendVerificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format_salah", "Format request salah")
		return
	}

	res, err := h.usecase.ResendVerification(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "data_tidak_valid", "Email wajib diisi.")
		case errors.Is(err, usecase.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "tidak_terdaftar", "Email belum terdaftar.")
		default:
			writeError(w, http.StatusInternalServerError, "server_error", "Gagal mengirim ulang kode. Coba lagi.")
		}
		return
	}

	payload := map[string]interface{}{
		"status":     "sukses",
		"message":    res.Message,
		"email_sent": res.EmailSent,
	}

	if res.DebugCode != "" {
		payload["debug_code"] = res.DebugCode
	}

	writeJSON(w, http.StatusOK, payload)
}

func (h *WargaHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format_salah", "Format request salah")
		return
	}

	res, err := h.usecase.Login(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			writeError(w, http.StatusUnauthorized, "email_tidak_ditemukan", "email tidak ditemukan")
		case errors.Is(err, usecase.ErrWrongPassword):
			writeError(w, http.StatusUnauthorized, "password_salah", "password salah")
		case errors.Is(err, usecase.ErrGoogleAccount):
			writeError(w, http.StatusForbidden, "akun_google", err.Error())
		case errors.Is(err, usecase.ErrEmailNotVerified):
			writeError(w, http.StatusForbidden, "email_belum_diverifikasi", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "server_error", "Login gagal. Coba lagi.")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "sukses",
		"token":  res.Token,
		"user":   res.User,
	})
}

func (h *WargaHandler) RegisterWithGoogle(w http.ResponseWriter, r *http.Request) {
	var req domain.GoogleAuthReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format_salah", "Format request salah")
		return
	}

	res, err := h.usecase.RegisterWithGoogle(
		r.Context(),
		req.IDToken,
		req.Password,
	)

	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "data_tidak_valid", "Password PETENG minimal 8 karakter.")
		case errors.Is(err, usecase.ErrGoogleNotConfigured):
			writeError(w, http.StatusServiceUnavailable, "google_belum_dikonfigurasi", err.Error())
		case errors.Is(err, usecase.ErrInvalidGoogleToken):
			writeError(w, http.StatusUnauthorized, "token_google_tidak_valid", "Token Google tidak valid. Coba pilih akun Google kembali.")
		case errors.Is(err, usecase.ErrGoogleEmailUnverified):
			writeError(w, http.StatusBadRequest, "email_google_belum_verifikasi", err.Error())
		case errors.Is(err, usecase.ErrEmailRegistered):
			writeError(w, http.StatusConflict, "email_terdaftar", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "server_error", "Registrasi Google gagal. Coba lagi.")
		}
		return
	}

	payload := map[string]interface{}{
		"status":     "sukses",
		"message":    res.Message,
		"email":      res.Email,
		"email_sent": res.EmailSent,
	}

	if res.DebugCode != "" {
		payload["debug_code"] = res.DebugCode
	}

	writeJSON(w, http.StatusCreated, payload)
}

// VerifyEmailLink menangani link verifikasi yang diklik dari email.
func (h *WargaHandler) VerifyEmailLink(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	code := r.URL.Query().Get("code")

	_, err := h.usecase.VerifyEmail(r.Context(), &domain.VerifyEmailReq{
		Email: email,
		Code:  code,
	})

	if err != nil {
		status := http.StatusBadRequest
		message := "Kode verifikasi tidak valid atau sudah kedaluwarsa."

		if errors.Is(err, usecase.ErrTooManyAttempts) {
			status = http.StatusTooManyRequests
			message = "Terlalu banyak percobaan. Buka aplikasi dan minta kode verifikasi baru."
		}

		writeVerifyPage(w, status, false, message)
		return
	}

	writeVerifyPage(
		w,
		http.StatusOK,
		true,
		"Email berhasil diverifikasi. Silakan kembali ke aplikasi PETENG dan login.",
	)
}

func writeVerifyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidCode):
		writeError(w, http.StatusBadRequest, "kode_salah", "Kode verifikasi salah. Periksa kembali email kamu.")
	case errors.Is(err, usecase.ErrCodeExpired):
		writeError(w, http.StatusBadRequest, "kode_kedaluwarsa", "Kode verifikasi sudah kedaluwarsa. Minta kode baru.")
	case errors.Is(err, usecase.ErrTooManyAttempts):
		writeError(w, http.StatusTooManyRequests, "terlalu_banyak_percobaan", "Terlalu banyak percobaan. Minta kode verifikasi baru.")
	case errors.Is(err, usecase.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "tidak_terdaftar", "Email belum terdaftar.")
	default:
		writeError(w, http.StatusInternalServerError, "server_error", "Verifikasi gagal. Coba lagi.")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]interface{}{
		"status":  "gagal",
		"code":    code,
		"message": message,
	})
}

func writeVerifyPage(w http.ResponseWriter, status int, success bool, message string) {
	title := "Verifikasi gagal"
	accent := "#B3261E"
	icon := "&#10007;"

	if success {
		title = "Email terverifikasi"
		accent = "#026343"
		icon = "&#10003;"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s - PETENG</title>
</head>
<body style="margin:0;background:#F1F1EF;font-family:Segoe UI,Arial,sans-serif;">
  <div style="min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px;box-sizing:border-box;">
    <div style="max-width:420px;width:100%%;background:#fff;border-radius:18px;padding:32px;text-align:center;box-shadow:0 10px 30px rgba(23,37,84,.12);">
      <div style="width:64px;height:64px;border-radius:50%%;background:%s;color:#fff;font-size:34px;line-height:64px;margin:0 auto 18px;">%s</div>
      <h1 style="margin:0 0 10px;font-size:20px;color:#172554;">%s</h1>
      <p style="margin:0;color:#5B6070;font-size:14px;line-height:1.6;">%s</p>
      <p style="margin:24px 0 0;color:#9AA0AE;font-size:11px;">PETENG - Petakan Gelap, Temukan Terang</p>
    </div>
  </div>
</body>
</html>`,
		html.EscapeString(title),
		accent,
		icon,
		html.EscapeString(title),
		html.EscapeString(message),
	)
}
