package handlers

import (
	"net/http"

	"ap_final/internal/store"
)

type SalonsHandler struct {
	store store.SalonStore
}

func NewSalonsHandler(store store.SalonStore) *SalonsHandler {
	return &SalonsHandler{store: store}
}

func (h *SalonsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	city := r.URL.Query().Get("city")
	list, err := h.store.ListPublic(r.Context(), city)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, list)
}
