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

func (h *LaporanHandler) UpdateStatusDishub(
	w http.ResponseWriter,
	r *http.Request,
) {
	// ID akun Dishub diambil dari JWT.
	userIDValue := r.Context().Value(middleware.UserIDKey)
	idDishub, ok := userIDValue.(int)
	if !ok {
		http.Error(
			w,
			`{"status":"gagal","message":"ID Dishub tidak ditemukan dari token"}`,
			http.StatusUnauthorized,
		)
		return
	}

	idLaporan := chi.URLParam(r, "id_laporan")
	if idLaporan == "" {
		http.Error(
			w,
			`{"status":"gagal","message":"ID laporan wajib diisi"}`,
			http.StatusBadRequest,
		)
		return
	}

	var req domain.UpdateStatusLaporanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			`{"status":"gagal","message":"Format request tidak valid"}`,
			http.StatusBadRequest,
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
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(
				w,
				`{"status":"gagal","message":"Laporan tidak ditemukan"}`,
				http.StatusNotFound,
			)
			return
		}

		// Validasi request yang tidak sesuai.
		if err.Error() == "status laporan tidak valid" ||
			err.Error() == "judul progres wajib diisi" ||
			err.Error() == "deskripsi progres wajib diisi" {
			http.Error(
				w,
				`{"status":"gagal","message":"`+err.Error()+`"}`,
				http.StatusBadRequest,
			)
			return
		}

		http.Error(
			w,
			`{"status":"gagal","message":"Gagal memperbarui status laporan"}`,
			http.StatusInternalServerError,
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
