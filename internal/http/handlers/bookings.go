package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/service"
)

type BookingsHandler struct {
	svc *service.BookingService
}

func NewBookingsHandler(svc *service.BookingService) *BookingsHandler {
	return &BookingsHandler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func isConflictErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "conflict:")
}

// Handle handles /bookings (POST create, GET list)
func (h *BookingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req CreateBookingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}

		b, err := h.svc.Create(r.Context(), req.ClientID, req.MasterID, req.ServiceID, req.StartAt)
		if err != nil {
			if isConflictErr(err) {
				writeErr(w, http.StatusConflict, err.Error())
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, b)

	case http.MethodGet:
		clientID := r.URL.Query().Get("clientId")
		masterID := r.URL.Query().Get("masterId")

		list, err := h.svc.List(r.Context(), clientID, masterID)
		if err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, list)

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleByID handles /bookings/{id} (DELETE cancel)
func (h *BookingsHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/bookings/")
	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodDelete {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	b, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}
