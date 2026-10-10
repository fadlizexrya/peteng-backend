package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"peteng-backend/internal/usecase"
)

type NavigasiHandler struct {
	navigasiUsecase *usecase.NavigasiUsecase
}

func NewNavigasiHandler(
	navigasiUsecase *usecase.NavigasiUsecase,
) *NavigasiHandler {
	return &NavigasiHandler{
		navigasiUsecase: navigasiUsecase,
	}
}

func (h *NavigasiHandler) CariRute(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	startLat, err := strconv.ParseFloat(query.Get("start_lat"), 64)
	if err != nil || query.Get("start_lat") == "" {
		writeNavigasiJSON(w, http.StatusBadRequest, map[string]any{
			"status":  "gagal",
			"message": "Parameter start_lat wajib berupa angka",
		})
		return
	}

	startLon, err := strconv.ParseFloat(query.Get("start_lon"), 64)
	if err != nil || query.Get("start_lon") == "" {
		writeNavigasiJSON(w, http.StatusBadRequest, map[string]any{
			"status":  "gagal",
			"message": "Parameter start_lon wajib berupa angka",
		})
		return
	}

	endLat, err := strconv.ParseFloat(query.Get("end_lat"), 64)
	if err != nil || query.Get("end_lat") == "" {
		writeNavigasiJSON(w, http.StatusBadRequest, map[string]any{
			"status":  "gagal",
			"message": "Parameter end_lat wajib berupa angka",
		})
		return
	}

	endLon, err := strconv.ParseFloat(query.Get("end_lon"), 64)
	if err != nil || query.Get("end_lon") == "" {
		writeNavigasiJSON(w, http.StatusBadRequest, map[string]any{
			"status":  "gagal",
			"message": "Parameter end_lon wajib berupa angka",
		})
		return
	}

	hasil, err := h.navigasiUsecase.CariRute(
		r.Context(),
		startLat,
		startLon,
		endLat,
		endLon,
	)
	if err != nil {
		writeNavigasiJSON(w, http.StatusBadGateway, map[string]any{
			"status":  "gagal",
			"message": "Tidak dapat memperoleh rute dari layanan routing",
		})
		return
	}

	writeNavigasiJSON(w, http.StatusOK, map[string]any{
		"status": "sukses",
		"data":   hasil,
	})
}

func writeNavigasiJSON(
	w http.ResponseWriter,
	statusCode int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
