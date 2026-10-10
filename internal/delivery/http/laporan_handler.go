package http

import (
	"encoding/json"
	"net/http"

	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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

// POST /api/warga/laporan
func (h *LaporanHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": "Laporan berhasil dikirim",
		"data":    laporan,
	})
}

// GET /api/warga/laporan
func (h *LaporanHandler) GetMyReports(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	laporans, err := h.usecase.DaftarLaporanWarga(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"Gagal mengambil daftar laporan",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": "Daftar laporan berhasil diambil",
		"data":    laporans,
	})
}

func (h *LaporanHandler) GetDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	idLaporan := chi.URLParam(r, "id_laporan")
	if idLaporan == "" {
		http.Error(
			w,
			"ID laporan wajib diisi",
			http.StatusBadRequest,
		)
		return
	}

	detail, err := h.usecase.DetailLaporanWarga(
		r.Context(),
		userID,
		idLaporan,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(
				w,
				"Laporan tidak ditemukan",
				http.StatusNotFound,
			)
			return
		}

		http.Error(
			w,
			"Gagal mengambil detail laporan",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": "Detail laporan berhasil diambil",
		"data":    detail,
	})
}
