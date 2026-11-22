package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"peteng-backend/internal/auth"
	"peteng-backend/internal/domain"
	"peteng-backend/internal/mailer"
	"peteng-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

const (
	verificationTTL      = 15 * time.Minute
	maxVerificationTries = 5
)

var (
	ErrInvalidInput          = errors.New("data pendaftaran tidak valid")
	ErrEmailRegistered       = errors.New("email sudah terdaftar")
	ErrEmailNotVerified      = errors.New("email belum diverifikasi. Cek kotak masuk email kamu")
	ErrUserNotFound          = errors.New("email tidak ditemukan")
	ErrWrongPassword         = errors.New("password salah")
	ErrGoogleAccount         = errors.New("akun ini terdaftar dengan Google. Silakan masuk dengan Google")
	ErrInvalidCode           = errors.New("kode verifikasi salah")
	ErrCodeExpired           = errors.New("kode verifikasi sudah kedaluwarsa. Minta kode baru")
	ErrTooManyAttempts       = errors.New("terlalu banyak percobaan. Minta kode verifikasi baru")
	ErrInvalidGoogleToken    = errors.New("token Google tidak valid")
	ErrGoogleNotConfigured   = errors.New("login Google belum dikonfigurasi di server")
	ErrGoogleEmailUnverified = errors.New("email Google belum terverifikasi")
)

type WargaUsecase struct {
	repo   *repository.WargaRepository
	jwt    *auth.JWTManager
	mailer *mailer.Mailer
	secret []byte
}

func NewWargaUsecase(
	repo *repository.WargaRepository,
	jwt *auth.JWTManager,
	mail *mailer.Mailer,
) *WargaUsecase {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "peteng-dev-secret"
	}

	return &WargaUsecase{
		repo:   repo,
		jwt:    jwt,
		mailer: mail,
		secret: []byte(secret),
	}
}

type RegisterResult struct {
	Message   string `json:"message"`
	Email     string `json:"email"`
	EmailSent bool   `json:"email_sent"`
	DebugCode string `json:"debug_code,omitempty"`
}

func (u *WargaUsecase) Register(
	ctx context.Context,
	req *domain.RegisterReq,
) (*RegisterResult, error) {
	nama := strings.TrimSpace(req.Nama)
	email := normalizeEmail(req.Email)
	telepon := strings.TrimSpace(req.NomorTelepon)

	if nama == "" || email == "" || len(req.Password) < 8 {
		return nil, ErrInvalidInput
	}

	existing, err := u.repo.GetByEmail(ctx, email)

	if err == nil {
		if existing.EmailVerified {
			return nil, ErrEmailRegistered
		}

		// Akun sudah ada tapi belum diverifikasi: kirim kode baru.
		return u.issueAndSend(ctx, existing, "Kode verifikasi baru sudah dikirim ke email kamu.")
	}

	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	warga := &domain.Warga{
		Nama:          nama,
		Email:         email,
		PasswordEmail: string(hashedPassword),
		NomorTelepon:  telepon,
		EmailVerified: false,
		AuthProvider:  "local",
	}

	if err := u.repo.CreateWarga(ctx, warga); err != nil {
		return nil, err
	}

	return u.issueAndSend(ctx, warga, "Akun berhasil dibuat. Kode verifikasi sudah dikirim ke email kamu.")
}

func (u *WargaUsecase) VerifyEmail(
	ctx context.Context,
	req *domain.VerifyEmailReq,
) (*domain.Warga, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)

	if email == "" || code == "" {
		return nil, ErrInvalidCode
	}

	warga, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	if warga.EmailVerified {
		warga.PasswordEmail = ""
		return warga, nil
	}

	if warga.VerificationToken == "" || warga.VerificationExpiresAt == nil {
		return nil, ErrInvalidCode
	}

	if time.Now().After(*warga.VerificationExpiresAt) {
		return nil, ErrCodeExpired
	}

	if warga.VerificationAttempts >= maxVerificationTries {
		_ = u.repo.ClearVerification(ctx, email)
		return nil, ErrTooManyAttempts
	}

	expected := u.hashCode(email, code)

	if !hmac.Equal([]byte(expected), []byte(warga.VerificationToken)) {
		attempts, err := u.repo.IncrementVerificationAttempts(ctx, email)
		if err == nil && attempts >= maxVerificationTries {
			_ = u.repo.ClearVerification(ctx, email)
			return nil, ErrTooManyAttempts
		}

		return nil, ErrInvalidCode
	}

	if err := u.repo.MarkEmailVerified(ctx, email); err != nil {
		return nil, err
	}

	warga.EmailVerified = true
	warga.VerificationToken = ""
	warga.VerificationExpiresAt = nil
	warga.VerificationAttempts = 0
	warga.PasswordEmail = ""

	return warga, nil
}

func (u *WargaUsecase) ResendVerification(
	ctx context.Context,
	req *domain.ResendVerificationReq,
) (*RegisterResult, error) {
	email := normalizeEmail(req.Email)

	if email == "" {
		return nil, ErrInvalidInput
	}

	warga, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	if warga.EmailVerified {
		return &RegisterResult{
			Message:   "Email kamu sudah terverifikasi. Silakan login.",
			EmailSent: false,
		}, nil
	}

	return u.issueAndSend(ctx, warga, "Kode verifikasi baru sudah dikirim ke email kamu.")
}

type LoginResult struct {
	Token string        `json:"token"`
	User  *domain.Warga `json:"user"`
}

func (u *WargaUsecase) Login(
	ctx context.Context,
	req *domain.LoginReq,
) (*LoginResult, error) {
	warga, err := u.repo.GetByEmail(ctx, normalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	if warga.AuthProvider == "google" {
		return nil, ErrGoogleAccount
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(warga.PasswordEmail),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrWrongPassword
	}

	if !warga.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	token, err := u.jwt.GenerateToken(warga.IDWarga, warga.Email, "warga")
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	warga.PasswordEmail = ""

	return &LoginResult{Token: token, User: warga}, nil
}

func (u *WargaUsecase) RegisterWithGoogle(
	ctx context.Context,
	idToken string,
	password string,
) (*RegisterResult, error) {
	// Validasi password PETENG.
	if len(password) < 8 {
		return nil, ErrInvalidInput
	}

	// Validasi konfigurasi Google.
	clientIDs := googleClientIDs()
	if len(clientIDs) == 0 {
		return nil, ErrGoogleNotConfigured
	}

	// Validasi ID Token dari Google.
	payload, err := validateGoogleToken(
		ctx,
		strings.TrimSpace(idToken),
		clientIDs,
	)
	if err != nil {
		log.Printf("validasi token Google gagal: %v", err)
		return nil, ErrInvalidGoogleToken
	}

	email := normalizeEmail(
		claimString(payload.Claims, "email"),
	)

	if email == "" || !claimBool(payload.Claims, "email_verified") {
		return nil, ErrGoogleEmailUnverified
	}

	name := strings.TrimSpace(
		claimString(payload.Claims, "name"),
	)

	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	googleID := strings.TrimSpace(payload.Subject)
	if googleID == "" {
		return nil, ErrInvalidGoogleToken
	}

	// Pastikan Google ID belum digunakan akun lain.
	existingGoogle, err := u.repo.GetByGoogleID(ctx, googleID)

	if err == nil && existingGoogle != nil {
		return nil, ErrEmailRegistered
	}

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	// Pastikan email belum digunakan.
	existingEmail, err := u.repo.GetByEmail(ctx, email)

	if err == nil && existingEmail != nil {
		return nil, ErrEmailRegistered
	}

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	// Hash password PETENG menggunakan bcrypt.
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	// Buat akun baru.
	warga := &domain.Warga{
		Nama:          name,
		Email:         email,
		PasswordEmail: string(hashedPassword),
		NomorTelepon:  "",
		EmailVerified: false,
		GoogleID:      &googleID,
		AuthProvider:  "local",
	}

	if err := u.repo.CreateWarga(ctx, warga); err != nil {
		return nil, err
	}

	// Kirim OTP ke email Google.
	return u.issueAndSend(
		ctx,
		warga,
		"Akun berhasil dibuat. Kode verifikasi sudah dikirim ke email kamu.",
	)
}

func (u *WargaUsecase) LoginWithGoogle(
	ctx context.Context,
	idToken string,
) (*LoginResult, error) {
	clientIDs := googleClientIDs()

	if len(clientIDs) == 0 {
		return nil, ErrGoogleNotConfigured
	}

	payload, err := validateGoogleToken(ctx, strings.TrimSpace(idToken), clientIDs)
	if err != nil {
		log.Printf("validasi token Google gagal: %v", err)
		return nil, ErrInvalidGoogleToken
	}

	email := strings.ToLower(strings.TrimSpace(claimString(payload.Claims, "email")))

	if email == "" || !claimBool(payload.Claims, "email_verified") {
		return nil, ErrGoogleEmailUnverified
	}

	name := strings.TrimSpace(claimString(payload.Claims, "name"))
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	googleID := payload.Subject

	warga, err := u.repo.GetByGoogleID(ctx, googleID)

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	if warga == nil {
		warga, err = u.repo.GetByEmail(ctx, email)

		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}

		if warga != nil {
			// Akun lokal dengan email sama: tautkan ke Google.
			if err := u.repo.UpdateGoogleID(ctx, warga.IDWarga, googleID); err != nil {
				return nil, err
			}

			if !warga.EmailVerified {
				if err := u.repo.MarkEmailVerified(ctx, email); err != nil {
					return nil, err
				}

				warga.EmailVerified = true
			}

			warga.GoogleID = &googleID
		} else {
			randomPassword, err := randomHex(32)
			if err != nil {
				return nil, err
			}

			hashedPassword, err := bcrypt.GenerateFromPassword(
				[]byte(randomPassword),
				bcrypt.DefaultCost,
			)
			if err != nil {
				return nil, err
			}

			warga = &domain.Warga{
				Nama:          name,
				Email:         email,
				PasswordEmail: string(hashedPassword),
				NomorTelepon:  "",
				EmailVerified: true,
				GoogleID:      &googleID,
				AuthProvider:  "google",
			}

			if err := u.repo.CreateWarga(ctx, warga); err != nil {
				return nil, err
			}
		}
	}

	token, err := u.jwt.GenerateToken(warga.IDWarga, warga.Email, "warga")
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	warga.PasswordEmail = ""

	return &LoginResult{Token: token, User: warga}, nil
}

func (u *WargaUsecase) issueAndSend(
	ctx context.Context,
	warga *domain.Warga,
	message string,
) (*RegisterResult, error) {
	code, err := generateCode()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(verificationTTL)
	tokenHash := u.hashCode(warga.Email, code)

	if err := u.repo.SetVerification(
		ctx,
		warga.Email,
		tokenHash,
		expiresAt,
	); err != nil {
		return nil, err
	}

	warga.VerificationToken = tokenHash
	warga.VerificationExpiresAt = &expiresAt
	warga.VerificationAttempts = 0

	result := &RegisterResult{
		Message:   message,
		Email:     warga.Email,
		EmailSent: false,
	}

	if u.mailer != nil && u.mailer.Enabled() {
		if err := u.mailer.SendVerification(warga.Email, warga.Nama, code); err != nil {
			log.Printf("gagal mengirim email verifikasi ke %s: %v", warga.Email, err)
			result.Message = "Akun dibuat, tetapi email verifikasi gagal dikirim. Coba kirim ulang."
		} else {
			result.EmailSent = true
		}
	} else if debugCodeAllowed() {
		// Mode dev: SMTP belum diisi, kode dicetak di log dan response.
		log.Printf("[DEV] Kode verifikasi untuk %s: %s", warga.Email, code)
		result.DebugCode = code
	} else {
		log.Printf(
			"PERINGATAN: SMTP belum dikonfigurasi, email verifikasi untuk %s tidak terkirim",
			warga.Email,
		)
	}

	return result, nil
}

// debugCodeAllowed hanya mengizinkan kode verifikasi tampil di response
// ketika APP_ENV bukan production.
func debugCodeAllowed() bool {
	return strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) != "production"
}

func (u *WargaUsecase) hashCode(email string, code string) string {
	mac := hmac.New(sha256.New, u.secret)
	mac.Write([]byte(normalizeEmail(email)))
	mac.Write([]byte("|"))
	mac.Write([]byte(strings.TrimSpace(code)))

	return hex.EncodeToString(mac.Sum(nil))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateCode() (string, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", number.Int64()), nil
}

func randomHex(length int) (string, error) {
	buffer := make([]byte, length)

	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

func googleClientIDs() []string {
	raw := os.Getenv("GOOGLE_CLIENT_ID")

	if strings.TrimSpace(raw) == "" {
		raw = os.Getenv("GOOGLE_CLIENT_IDS")
	}

	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}

	return result
}

func validateGoogleToken(
	ctx context.Context,
	idToken string,
	clientIDs []string,
) (*idtoken.Payload, error) {
	if idToken == "" {
		return nil, errors.New("id token kosong")
	}

	var lastErr error

	for _, clientID := range clientIDs {
		payload, err := idtoken.Validate(ctx, idToken, clientID)
		if err == nil {
			return payload, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

func claimString(claims map[string]interface{}, key string) string {
	value, ok := claims[key]
	if !ok || value == nil {
		return ""
	}

	if text, ok := value.(string); ok {
		return text
	}

	return fmt.Sprint(value)
}

func claimBool(claims map[string]interface{}, key string) bool {
	value, ok := claims[key]
	if !ok || value == nil {
		return false
	}

	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(typed, "true")
	default:
		return false
	}
}
