package handlers

import (
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
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	salonID := r.URL.Query().Get("salonId")
	if salonID == "" {
		writeErr(w, http.StatusBadRequest, "salonId required")
		return
	}

	services, err := h.store.ListBySalon(r.Context(), salonID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, services)
}
