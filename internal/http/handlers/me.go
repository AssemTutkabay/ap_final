package handlers

import (
	"net/http"

	"ap_final/internal/http/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler { return &MeHandler{} }

func (h *MeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
