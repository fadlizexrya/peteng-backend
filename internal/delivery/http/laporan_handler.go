package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

// GET /api/warga/laporan/{id_laporan}
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

// GET /api/dinas/laporan
func (h *LaporanHandler) GetAllForDishub(
	w http.ResponseWriter,
	r *http.Request,
) {
	laporans, err := h.usecase.DaftarSemuaLaporan(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "gagal",
			"message": "Gagal mengambil daftar laporan",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": "Daftar laporan berhasil diambil",
		"data":    laporans,
	})
}

// PATCH /api/dinas/laporan/{id_laporan}/status
func (h *LaporanHandler) UpdateStatusDishub(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Ambil ID akun Dishub dari JWT.
	userIDValue := r.Context().Value(middleware.UserIDKey)

	idDishub, ok := userIDValue.(int)
	if !ok {
		writeJSONError(
			w,
			http.StatusUnauthorized,
			"ID Dishub tidak ditemukan dari token",
		)
		return
	}

	idLaporan := chi.URLParam(r, "id_laporan")
	if idLaporan == "" {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"ID laporan wajib diisi",
		)
		return
	}

	var req domain.UpdateStatusLaporanReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"Format request tidak valid",
		)
		return
	}

	err := h.usecase.UpdateStatusLaporanDishub(
		r.Context(),
		idLaporan,
		idDishub,
		&req,
	)

	if err != nil {
		// Laporan tidak ditemukan.
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(
				w,
				http.StatusNotFound,
				"Laporan tidak ditemukan",
			)
			return
		}

		// Kesalahan validasi request.
		if err.Error() == "status laporan tidak valid" ||
			err.Error() == "judul progres wajib diisi" ||
			err.Error() == "deskripsi progres wajib diisi" {
			writeJSONError(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		// Transisi status tidak diperbolehkan.
		if strings.HasPrefix(
			err.Error(),
			"transisi status tidak diizinkan:",
		) {
			writeJSONError(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
			return
		}

		// Error database atau server lainnya.
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"Gagal memperbarui status laporan",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": "Status laporan dan progres berhasil diperbarui",
	})
}

// Helper untuk respons error JSON.
func writeJSONError(
	w http.ResponseWriter,
	statusCode int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "gagal",
		"message": message,
	})
}
