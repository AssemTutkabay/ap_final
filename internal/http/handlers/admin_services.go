package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ap_final/internal/domain"
	"ap_final/internal/http/middleware"
	"ap_final/internal/store"
)

type AdminServicesHandler struct {
	services store.ServiceStore
	admins   store.SalonAdminStore
}

func NewAdminServicesHandler(s store.ServiceStore, a store.SalonAdminStore) *AdminServicesHandler {
	return &AdminServicesHandler{services: s, admins: a}
}

type AdminCreateServiceRequest struct {
	ID          string `json:"id,omitempty"`
	SalonID     string `json:"salonId"`
	Name        string `json:"name"`
	PriceKZT    int    `json:"priceKzt"`
	DurationMin int    `json:"durationMin"`
	IsActive    bool   `json:"isActive"`
}

type AdminUpdateServiceRequest struct {
	Name        string `json:"name"`
	PriceKZT    int    `json:"priceKzt"`
	DurationMin int    `json:"durationMin"`
	IsActive    bool   `json:"isActive"`
}

// POST /admin/services
func (h *AdminServicesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AdminCreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.SalonID = strings.TrimSpace(req.SalonID)
	req.Name = strings.TrimSpace(req.Name)

	if req.SalonID == "" || req.Name == "" {
		writeErr(w, http.StatusBadRequest, "salonId and name required")
		return
	}
	if req.PriceKZT < 0 || req.DurationMin <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid priceKzt/durationMin")
		return
	}
	if req.ID == "" {
		req.ID = newID("sv_")
	}

	// salon_admin: только свои салоны
	if u.Role == domain.RoleSalonAdmin {
		allowed, err := h.admins.IsAdminOfSalon(r.Context(), u.ID, req.SalonID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !allowed {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	sv := domain.Service{
		ID:          req.ID,
		SalonID:     req.SalonID,
		Name:        req.Name,
		PriceKZT:    req.PriceKZT,
		DurationMin: req.DurationMin,
		IsActive:    req.IsActive,
	}

	created, err := h.services.CreateForSalon(r.Context(), sv)
	if err != nil {
		if err == store.ErrConflict {
			writeErr(w, http.StatusConflict, "conflict")
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// PUT/DELETE /admin/services/{id}
func (h *AdminServicesHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFrom(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/services/")
	id = strings.TrimSuffix(id, "/")
	id = strings.TrimSpace(id)

	if id == "" || strings.Contains(id, "/") {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}

	// salon_admin ownership: узнаем salonId через GetByID
	if u.Role == domain.RoleSalonAdmin {
		sv, ok2, err := h.services.GetByID(r.Context(), id)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !ok2 {
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		allowed, err := h.admins.IsAdminOfSalon(r.Context(), u.ID, sv.SalonID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !allowed {
			writeErr(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	switch r.Method {
	case http.MethodPut:
		var req AdminUpdateServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad json")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			writeErr(w, http.StatusBadRequest, "name required")
			return
		}
		if req.PriceKZT < 0 || req.DurationMin <= 0 {
			writeErr(w, http.StatusBadRequest, "invalid priceKzt/durationMin")
			return
		}

		sv := domain.Service{
			ID:          id,
			Name:        req.Name,
			PriceKZT:    req.PriceKZT,
			DurationMin: req.DurationMin,
			IsActive:    req.IsActive,
		}

		updated, err := h.services.Update(r.Context(), sv)
		if err != nil {
			if err == store.ErrNotFound {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			if err == store.ErrConflict {
				writeErr(w, http.StatusConflict, "conflict")
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, updated)

	case http.MethodDelete:
		err := h.services.Disable(r.Context(), id)
		if err != nil {
			if err == store.ErrNotFound {
				writeErr(w, http.StatusNotFound, "not found")
				return
			}
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})

	default:
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
