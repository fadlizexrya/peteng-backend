package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"peteng-backend/internal/domain"
	"peteng-backend/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type ValidasiAIHandler struct {
	usecase *usecase.ValidasiAIUsecase
}

func NewValidasiAIHandler(
	u *usecase.ValidasiAIUsecase,
) *ValidasiAIHandler {
	return &ValidasiAIHandler{usecase: u}
}

// POST /api/dinas/laporan/{id_laporan}/validasi-ai/mulai
func (h *ValidasiAIHandler) Mulai(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id_laporan")

	err := h.usecase.MulaiValidasi(r.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if errors.Is(err, pgx.ErrNoRows) {
			statusCode = http.StatusNotFound
		} else if len(err.Error()) >= 29 &&
			err.Error()[:29] == "laporan harus berstatus diajukan" {
			statusCode = http.StatusBadRequest
		}

		writeJSONError(w, statusCode, err.Error())
		return
	}

	writeJSONSuccess(w, http.StatusOK, "Laporan masuk proses validasi AI")
}

// POST /api/dinas/laporan/{id_laporan}/validasi-ai/mock
func (h *ValidasiAIHandler) SimpanHasilMock(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id_laporan")

	var req domain.ValidasiAIMockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	err := h.usecase.SimpanHasilMock(r.Context(), id, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError

		if errors.Is(err, pgx.ErrNoRows) {
			statusCode = http.StatusNotFound
		} else if err.Error() == "hasil_validasi harus valid atau tidak_valid" ||
			err.Error() == "tingkat_kegelapan_ai harus bernilai 1, 2, atau 3" ||
			err.Error() == "catatan wajib diisi" ||
			len(err.Error()) >= 34 &&
				err.Error()[:34] == "laporan harus berstatus diproses_ai" {
			statusCode = http.StatusBadRequest
		}

		writeJSONError(w, statusCode, err.Error())
		return
	}

	writeJSONSuccess(w, http.StatusOK, "Hasil validasi mock berhasil disimpan")
}

// GET /api/dinas/laporan/{id_laporan}/validasi-ai
func (h *ValidasiAIHandler) GetTerbaru(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id_laporan")

	hasil, err := h.usecase.GetTerbaru(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "Hasil validasi belum tersedia")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, "Gagal mengambil hasil validasi")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "sukses",
		"data":   hasil,
	})
}

func writeJSONSuccess(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "sukses",
		"message": message,
	})
}
