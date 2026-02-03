package handlers

import (
	"encoding/json"
	"net/http"

	"ap_final/internal/store"
)

type ServicesHandler struct {
	store store.ServiceStore
}

func NewServicesHandler(store store.ServiceStore) *ServicesHandler {
	return &ServicesHandler{store: store}
}

func (h *ServicesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	salonID := r.URL.Query().Get("salonId")

	services, err := h.store.List(r.Context(), salonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(services)
}

