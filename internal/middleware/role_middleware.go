package middleware

import (
	"net/http"
)

func RequireRole(allowedRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			role, ok := r.Context().Value(RoleKey).(string)

			if !ok || role == "" {
				http.Error(
					w,
					"Role tidak ditemukan",
					http.StatusForbidden,
				)
				return
			}

			if role != allowedRole {
				http.Error(
					w,
					"Akses ditolak",
					http.StatusForbidden,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
