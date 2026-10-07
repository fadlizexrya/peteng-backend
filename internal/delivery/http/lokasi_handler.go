package http

import (
	"encoding/json"
	"net/http"

	"peteng-backend/internal/usecase"
)

type LokasiHandler struct {
	usecase *usecase.LokasiUsecase
}

func NewLokasiHandler(
	u *usecase.LokasiUsecase,
) *LokasiHandler {
	return &LokasiHandler{
		usecase: u,
	}
}

func (h *LokasiHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {

	lokasi, err := h.usecase.GetAll(r.Context())

	if err != nil {
		http.Error(
			w,
			"Gagal mengambil data lokasi",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"status": "sukses",
			"data":   lokasi,
		},
	)
}