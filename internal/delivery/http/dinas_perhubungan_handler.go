package http

import (
	"encoding/json"
	"net/http"
	"peteng-backend/internal/entity"
	"peteng-backend/internal/usecase"
)

type DinasPerhubunganHandler struct {
	usecase usecase.DinasPerhubunganUsecase
}

func NewDinasPerhubunganHandler(
	u usecase.DinasPerhubunganUsecase,
) *DinasPerhubunganHandler {
	return &DinasPerhubunganHandler{
		usecase: u,
	}
}

func (h *DinasPerhubunganHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req entity.RegisterDinasRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Payload request tidak valid",
		})
		return
	}

	res, err := h.usecase.Register(r.Context(), req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Akun Dinas Perhubungan berhasil dibuat",
		"data":    res,
	})
}

func (h *DinasPerhubunganHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	var req entity.LoginDinasRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Payload request tidak valid",
		})
		return
	}

	// Login sekarang mengembalikan:
	// data Dinas Perhubungan + JWT token + error
	res, token, err := h.usecase.Login(
		r.Context(),
		req,
	)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login Dinas Perhubungan berhasil",
		"token":   token,
		"user":    res,
	})
}
