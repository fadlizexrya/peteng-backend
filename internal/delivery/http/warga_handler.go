package http

import (
	"encoding/json"
	"net/http"
	"peteng-backend/internal/domain"
	"peteng-backend/internal/usecase"
)

type WargaHandler struct {
	usecase *usecase.WargaUsecase
}

func NewWargaHandler(u *usecase.WargaUsecase) *WargaHandler {
	return &WargaHandler{usecase: u}
}

func (h *WargaHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format request salah", http.StatusBadRequest)
		return
	}

	res, err := h.usecase.Register(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "sukses", "data": res})
}

func (h *WargaHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format request salah", http.StatusBadRequest)
		return
	}

	res, err := h.usecase.Login(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "sukses", "token": res.Token, "user": res.User})
}
