package middleware

import (
	"context"
	"net/http"
	"strings"

	"peteng-backend/internal/auth"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
	RoleKey   contextKey = "role"
)

func JWTMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Ambil Authorization Header
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(
					w,
					"Authorization header diperlukan",
					http.StatusUnauthorized,
				)
				return
			}

			// Format yang diharapkan:
			// Authorization: Bearer <token>

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(
					w,
					"Format Authorization tidak valid",
					http.StatusUnauthorized,
				)
				return
			}

			tokenString := parts[1]

			// Validasi JWT
			claims, err := jwtManager.ValidateToken(tokenString)

			if err != nil {
				http.Error(
					w,
					"Token tidak valid atau sudah expired",
					http.StatusUnauthorized,
				)
				return
			}

			// Simpan informasi user ke context
			ctx := context.WithValue(
				r.Context(),
				UserIDKey,
				claims.UserID,
			)

			ctx = context.WithValue(
				ctx,
				EmailKey,
				claims.Email,
			)

			ctx = context.WithValue(
				ctx,
				RoleKey,
				claims.Role,
			)

			// Lanjutkan request
			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}
