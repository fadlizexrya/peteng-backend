package http

import (
	"encoding/json"
	"net/http"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/middleware"
	"peteng-backend/internal/usecase"
)

type LaporanHandler struct {
	usecase *usecase.LaporanUsecase
}

func NewLaporanHandler(
	u *usecase.LaporanUsecase,
) *LaporanHandler {
	return &LaporanHandler{
		usecase: u,
	}
}

func (h *LaporanHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Ambil ID warga dari JWT middleware.
	userIDValue := r.Context().Value(middleware.UserIDKey)

	userID, ok := userIDValue.(int)
	if !ok {
		http.Error(
			w,
			"ID warga tidak ditemukan dari token",
			http.StatusUnauthorized,
		)
		return
	}

	var req domain.LaporanReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"Format request salah",
			http.StatusBadRequest,
		)
		return
	}

	laporan, err := h.usecase.BuatLaporan(
		r.Context(),
		userID,
		&req,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"status":  "sukses",
			"message": "Laporan berhasil dikirim",
			"data":    laporan,
		},
	)
}
