package handlers

import (
	"encoding/json"
	"net/http"

	"ap_final/internal/domain"
)

type SalonsHandler struct {
	salons []domain.Salon
}

func NewSalonsHandler() *SalonsHandler {
	return &SalonsHandler{
		salons: []domain.Salon{
			{ID: "s1", Name: "Glow Studio"},
			{ID: "s2", Name: "Nail Bar"},
		},
	}
}

func (h *SalonsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.salons)
}
